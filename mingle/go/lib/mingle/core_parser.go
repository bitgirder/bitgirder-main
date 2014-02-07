package mingle

import(
//    "log"
//    "io"
    "time"
    "fmt"
    "bytes"
    "mingle/parser/lexer"
    "mingle/parser/loc"
    "mingle/parser/syntax"
)

func doParseAndCreate( f func() ( interface{}, error ) ) interface{} {
    res, err := f()
    if err != nil { panic( err ) }
    return res
}

func implSxBldr( s string, isExt bool ) *syntax.Builder {
    opts := &lexer.Options{
        Reader: bytes.NewBufferString( s ),
        SourceName: loc.ParseSourceInput,
        IsExternal: isExt,
    }
    return syntax.NewBuilder( lexer.New( opts ) )
}

func sxBldr( s string ) *syntax.Builder {
    return implSxBldr( s, false )
}

func sxBldrExt( s string ) *syntax.Builder {
    return implSxBldr( s, true )
}

func ConvertSyntaxId( sxId syntax.Identifier ) *Identifier {
    return &Identifier{ sxId }
}

func ConvertSyntaxNamespace( sxNs *syntax.Namespace ) *Namespace {
    ns := new( Namespace )
    ns.Parts = make( []*Identifier, len( sxNs.Parts ) )
    for i, nm := range sxNs.Parts { ns.Parts[ i ] = ConvertSyntaxId( nm ) }
    ns.Version = ConvertSyntaxId( sxNs.Version )
    return ns
}

func ConvertSyntaxDeclaredTypeName( 
    nm syntax.DeclaredTypeName ) *DeclaredTypeName {
    return &DeclaredTypeName{ string( nm ) }
}

func ConvertSyntaxQname( sxQn *syntax.QualifiedTypeName ) *QualifiedTypeName {
    return &QualifiedTypeName{
        Namespace: ConvertSyntaxNamespace( sxQn.Namespace ), 
        Name: ConvertSyntaxDeclaredTypeName( sxQn.Name ),
    }
}

func ConvertSyntaxTypeName( nm syntax.TypeName ) TypeName {
    switch v := nm.( type ) {
    case *syntax.QualifiedTypeName: return ConvertSyntaxQname( v )
    case syntax.DeclaredTypeName: return ConvertSyntaxDeclaredTypeName( v )
    }
    panic( fmt.Errorf( "Unhandled syntax.TypeName: %T", nm ) )
}

func ParseIdentifier( input string ) ( id *Identifier, err error ) {
    sb := sxBldrExt( input )
    var sxId *syntax.TokenNode
    if sxId, err = sb.ExpectIdentifier(); err == nil {
        err = sb.CheckTrailingToken()
        id = ConvertSyntaxId( sxId.Identifier() )
    }
    return
}

func MustIdentifier( input string ) *Identifier {
    return doParseAndCreate( func() ( interface{}, error ) { 
        return ParseIdentifier( input ) 
    }).( *Identifier )
}

func ParseNamespace( input string ) ( ns *Namespace, err error ) {
    sb := sxBldrExt( input )
    var sxNs *syntax.Namespace
    sxNs, _, err = sb.ExpectNamespace( nil )
    if err == nil { err = sb.CheckTrailingToken() }
    if err == nil { ns = ConvertSyntaxNamespace( sxNs ) }
    return
}

func MustNamespace( input string ) *Namespace {
    return doParseAndCreate( func() ( interface{}, error ) {
        return ParseNamespace( input )
    }).( *Namespace )
}

func ParseDeclaredTypeName( input string ) ( nm *DeclaredTypeName, err error ) {
    sb := sxBldr( input )
    var sxNm *syntax.TokenNode
    sxNm, err = sb.ExpectDeclaredTypeName()
    if err == nil { err = sb.CheckTrailingToken() }
    if err == nil { 
        nm = ConvertSyntaxDeclaredTypeName( sxNm.DeclaredTypeName() ) 
    }
    return 
}

func MustDeclaredTypeName( input string ) *DeclaredTypeName {
    return doParseAndCreate( func() ( interface{}, error ) {
        return ParseDeclaredTypeName( input )
    }).( *DeclaredTypeName )
}

func ParseQualifiedTypeName(
    input string ) ( qn *QualifiedTypeName, err error ) {
    sb := sxBldr( input )
    var sxQn *syntax.QualifiedTypeName
    sxQn, _, err = sb.ExpectQualifiedTypeName( nil )
    if err == nil { err = sb.CheckTrailingToken() }
    if err == nil { qn = ConvertSyntaxQname( sxQn ) }
    return
}

func MustQualifiedTypeName( input string ) *QualifiedTypeName {
    return doParseAndCreate( func() ( interface{}, error ) {
        return ParseQualifiedTypeName( input )
    }).( *QualifiedTypeName )
}

func resolveStandardTypeName( nm TypeName ) TypeName {
    if dn, ok := nm.( *DeclaredTypeName ); ok {
        if qn, ok := ResolveInCore( dn ); ok { return qn }
    }
    return nm
}

type RestrictionTypeError struct { msg string }

func NewRestrictionTypeError( msg string ) *RestrictionTypeError {
    return &RestrictionTypeError{ msg }
}

func errorRestrictionTargetType( 
    nm TypeName, sx syntax.RestrictionSyntax ) *RestrictionTypeError {
    sxNm := ""
    switch sx.( type ) {
    case *syntax.RangeRestrictionSyntax: sxNm = "range"
    case *syntax.RegexRestrictionSyntax: sxNm = "regex"
    default: panic( fmt.Errorf( "Unhandled restriction type: %T", sx ) )
    }
    msg := fmt.Sprintf( "Invalid target type for %s restriction: %s", sxNm, nm )
    return NewRestrictionTypeError( msg )
}

func ( e *RestrictionTypeError ) Error() string { return e.msg }

func resolveRegexRestriction( 
    qn *QualifiedTypeName, 
    rx *syntax.RegexRestrictionSyntax ) ( ValueRestriction, error ) {
    if qn.Equals( QnameString ) { 
        if rr, err := NewRegexRestriction( rx.Pat ); err == nil {
            return rr, nil
        } else { return nil, &RestrictionTypeError{ err.Error() } }
    }
    return nil, errorRestrictionTargetType( qn, rx )
}

var rangeValTypes []*AtomicTypeReference
func init() {
    rangeValTypes = []*AtomicTypeReference{
        TypeString,
        TypeInt32,
        TypeInt64,
        TypeUint32,
        TypeUint64,
        TypeFloat32,
        TypeFloat64,
        TypeTimestamp,
    }
}

// helper for castRangeValue()
func checkRangeValueCast(
    sx interface{}, typ TypeReference, bound string ) error {
    typStr := ""
    switch v := sx.( type ) {
    case *syntax.StringRestrictionSyntax: 
        if IsNumericType( typ ) { typStr = "string" }
    case *syntax.NumRestrictionSyntax: 
        if ! IsNumericType( typ ) { typStr = "number" }
        if IsIntegerType( typ ) && 
           ( ! ( v.Num.Frac == "" && v.Num.Exp == "" ) ) {
            typStr = "decimal"
        }
    default: panic( fmt.Errorf( "Unhandled type: %T", sx ) )
    }
    if typStr != "" {
        msg := fmt.Sprintf( "Got %s as %s value for range", typStr, bound )
        return &RestrictionTypeError{ msg }
    }
    return nil
}

// bound is which bound to report in the error: "min" or "max"
func castRangeValue( 
    sx interface{},
    at *AtomicTypeReference, 
    bound string ) ( val Value, err error ) {
    if err = checkRangeValueCast( sx, at, bound ); err != nil { return }
    ms := String( sx.( syntax.LiteralStringer ).LiteralString() )
    if val, err = castAtomic( ms, at, idPathRootVal ); err != nil {
        msg := "Invalid %s value in range restriction: %s" 
        err = &RestrictionTypeError{ fmt.Sprintf( msg, bound, err.Error() ) }
    }
    return
}

func areAdjacentInts( min, max Value ) bool {
    switch minV := min.( type ) {
    case Int32: return int32( max.( Int32 ) ) - int32( minV ) == int32( 1 )
    case Uint32: return uint32( max.( Uint32 ) ) - uint32( minV ) == uint32( 1 )
    case Int64: return int64( max.( Int64 ) ) - int64( minV ) == int64( 1 )
    case Uint64: return uint64( max.( Uint64 ) ) - uint64( minV ) == uint64( 1 )
    }
    return false
}

func checkRangeBounds( rr *RangeRestriction ) error {
    failed := false
    switch i := rr.Min.( Comparer ).Compare( rr.Max ); {
    case i == 0: failed = ! ( rr.MinClosed && rr.MaxClosed )
    case i > 0: failed = true
    case i < 0: 
        open := ! ( rr.MinClosed || rr.MaxClosed )
        failed = open && areAdjacentInts( rr.Min, rr.Max )
    }
    if failed { return &RestrictionTypeError{ "Unsatisfiable range" } }
    return nil
}

func setRangeValues(
    rr *RangeRestriction, 
    rx *syntax.RangeRestrictionSyntax, 
    at *AtomicTypeReference ) ( err error ) {
    if rx.Left != nil {
        if rr.Min, err = castRangeValue( rx.Left, at, "min" ); err != nil { 
            return 
        }
    }
    if rx.Right != nil {
        if rr.Max, err = castRangeValue( rx.Right, at, "max" ); err != nil {
            return
        }
    }
    if ! ( rr.Min == nil || rr.Max == nil ) { err = checkRangeBounds( rr ) }
    return
}

func resolveRangeRestriction(
    qn *QualifiedTypeName,
    rx *syntax.RangeRestrictionSyntax ) ( ValueRestriction, error ) {
    rr := 
        &RangeRestriction{ MinClosed: rx.LeftClosed, MaxClosed: rx.RightClosed }
    for _, rvTyp := range rangeValTypes {
        if qn.Equals( rvTyp.Name ) {
            if err := setRangeValues( rr, rx, rvTyp ); err != nil {
                return nil, err
            }
            return rr, nil
        }
    }
    return nil, errorRestrictionTargetType( qn, rx )
}

func ResolveStandardRestriction(
    qn *QualifiedTypeName, 
    sx syntax.RestrictionSyntax ) ( ValueRestriction, error ) {
    switch v := sx.( type ) {
    case *syntax.RegexRestrictionSyntax: return resolveRegexRestriction( qn, v )
    case *syntax.RangeRestrictionSyntax: return resolveRangeRestriction( qn, v )
    }
    panic( fmt.Errorf( "Unhandled restriction: %v", sx ) )
}

// could be inlined in parseStandardTypeReference(), but we use this in tests as
// well, so pull it out separately
func parseCompletableTypeReference( 
    s string ) ( *syntax.CompletableTypeReference, error ) {
    sb := sxBldr( s )
    ctr, _, err := sb.ExpectTypeReference( nil )
    if err == nil { err = sb.CheckTrailingToken() }
    return ctr, err
}

func typeCompleter( val interface{}, tq syntax.TypeQuantifier ) interface{} {
    typ := val.( TypeReference )
    switch tq {
    case syntax.TypeQuantifierNullable: return &NullableTypeReference{ typ }
    case syntax.TypeQuantifierList: return &ListTypeReference{ typ, true }
    case syntax.TypeQuantifierNonEmptyList:
        return &ListTypeReference{ typ, false }
    }
    panic( fmt.Errorf( "Unhandled type quantifier: %v", tq ) )
}

func CompleteType( 
    typ TypeReference, ctr *syntax.CompletableTypeReference ) TypeReference {
    return ctr.CompleteType( typ, typeCompleter ).( TypeReference )
}

func parseTypeReference( s string ) ( TypeReference, error ) {
    ctr, err := parseCompletableTypeReference( s )
    if err != nil { return nil, err }
    nm := resolveStandardTypeName( ConvertSyntaxTypeName( ctr.Name ) )
    var vr ValueRestriction
    if sx := ctr.Restriction; sx != nil {
        if qn, ok := nm.( *QualifiedTypeName ); ok {
            if vr, err = ResolveStandardRestriction( qn, sx ); err != nil {
                return nil, err
            }
        } else { return nil, errorRestrictionTargetType( nm, sx ) }
    }
    atr := &AtomicTypeReference{ Name: nm, Restriction: vr }
    return CompleteType( atr, ctr ), nil
}

func ParseTypeReferenceBytes( input []byte ) ( TypeReference, error ) {
    return parseTypeReference( string( input ) )
}

func ParseTypeReference( input string ) ( TypeReference, error ) {
    return parseTypeReference( input )
}

func MustTypeReference( input string ) TypeReference {
    return doParseAndCreate( func() ( interface{}, error ) {
        return ParseTypeReference( input )
    }).( TypeReference )
}

func ParseIdentifiedName( input string ) ( nm *IdentifiedName, err error ) {
    sb := sxBldrExt( input )
    nm = &IdentifiedName{ Names: make( []*Identifier, 0, 4 ) }
    var sxNs *syntax.Namespace
    if sxNs, _, err = sb.ExpectNamespace( nil ); err == nil { 
        nm.Namespace = ConvertSyntaxNamespace( sxNs )
    } else { return }
    for sb.HasTokens() && err == nil {
        if _, err = sb.ExpectSpecial( lexer.SpecialTokenForwardSlash ); 
            err == nil {
            var sxId *syntax.TokenNode
            if sxId, err = sb.ExpectIdentifier(); err == nil {
                id := ConvertSyntaxId( sxId.Identifier() )
                nm.Names = append( nm.Names, id )
            }
        }
    }
    if err == nil && len( nm.Names ) == 0 {
        err = &loc.ParseError{ "Missing name", sb.Location() }
    }
    return
}

func MustIdentifiedName( input string ) *IdentifiedName {
    return doParseAndCreate( func() ( interface{}, error ) {
        return ParseIdentifiedName( input )
    }).( *IdentifiedName )
}

func ParseTimestamp( str string ) ( Timestamp, error ) {
    t, err := time.Parse( time.RFC3339Nano, str )
    if err != nil {
        parseErr := &loc.ParseError{
            Message: fmt.Sprintf( "Invalid RFC3339 time: %q", str ),
            Loc: &loc.Location{ 1, 1, loc.ParseSourceInput },
        }
        return Timestamp( t ), parseErr
    }
    return Timestamp( t ), nil
}

func MustTimestamp( str string ) Timestamp {
    t, err := ParseTimestamp( str )
    if err != nil { panic( err ) }
    return t
}
