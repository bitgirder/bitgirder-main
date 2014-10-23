package mingle

import (
    "bitgirder/objpath"
    "fmt"
    "regexp"
//    "log"
    "time"
    "sort"
    "bytes"
    "unicode"
    "strings"
    "strconv"
)

// values declared and accepted by this package are always > 0; 0 may be used
// privately as the null/unknown format
type IdentifierFormat uint

const (
    _ IdentifierFormat = iota // leave zero val internally for 'unknown'
    LcUnderscore 
    LcHyphenated
    LcCamelCapped
)

func ( idFmt IdentifierFormat ) String() string {
    switch idFmt {
    case LcUnderscore: return "lc-underscore"
    case LcHyphenated: return "lc-hyphenated"
    case LcCamelCapped: return "lc-camel-capped"
    }
    panic( fmt.Errorf( "Unrecognized id format: %d", idFmt ) )
}

func IdentifierFormatString( nm string ) ( idFmt IdentifierFormat, err error ) {
    switch nm {
    case "lc-underscore": idFmt = LcUnderscore
    case "lc-hyphenated": idFmt = LcHyphenated
    case "lc-camel-capped": idFmt = LcCamelCapped
    default: err = fmt.Errorf( "Unrecognized id format: %s", nm )
    }
    return
}

func MustIdentifierFormatString( nm string ) IdentifierFormat {
    res, err := IdentifierFormatString( nm )
    if err != nil { panic( err ) }
    return res
}

var IdentifierFormats = 
    []IdentifierFormat{ LcUnderscore, LcHyphenated, LcCamelCapped }

var IdentifierPartRegexp = regexp.MustCompile( "^[a-z][a-z0-9]*$" )

type IdentifierPartFormatError struct { Part string }

func ( e *IdentifierPartFormatError ) Error() string {
    return fmt.Sprintf( "invalid identifier part: %s", e.Part )
}

func getIdentifierPartError( s string ) error {
    if res := IdentifierPartRegexp.FindStringIndex( s ); res == nil {
        return &IdentifierPartFormatError{ s }
    }
    return nil
}
 
type Identifier struct {
    parts []string
}

// Meant for other mingle impl packages only; external callers should not count
// on the behavior of this remaining stable, or even of it continuing to exist
func NewIdentifierUnsafe( parts []string ) *Identifier {
    return &Identifier{ parts }
}

// returns the live parts array
func ( id *Identifier ) GetPartsUnsafe() []string { return id.parts }

func idSeparatorFor( idFmt IdentifierFormat ) (sep byte) {
    switch idFmt {
        case LcHyphenated: sep = byte( '-' )
        case LcUnderscore: sep = byte( '_' )
        default: panic( fmt.Sprintf( "Unhandled id format: %v", idFmt ) )
    }
    return
}

func ( id *Identifier ) formatToBuf( 
    buf []byte, idFmt IdentifierFormat ) []byte {
    for i, part := range id.parts {
        var cpStart int
        if idFmt == LcCamelCapped {
            if i > 0 {
                ch0 := rune( part[ 0 ] )
                buf = append( buf, byte( unicode.ToUpper( ch0 ) ) )
                cpStart = 1
            }
        } else {
            if i > 0 { buf = append( buf, idSeparatorFor( idFmt ) ) }
        }
        buf = append( buf, part[ cpStart : ]... )
    }
    return buf
}

func ( id *Identifier ) extFormToBuf( buf []byte ) []byte {
    return id.formatToBuf( buf, LcHyphenated )
}

func ( id *Identifier ) Format( idFmt IdentifierFormat ) string {
    return string( id.formatToBuf( make( []byte, 0, 32 ), idFmt ) )
}

func ( id *Identifier ) ExternalForm() string {
    return id.Format( LcHyphenated )
}

func ( id *Identifier ) String() string { return id.ExternalForm() }

func compareIdParts( p1, p2 string ) int {
    if p1 == p2 { return 0 }
    if p1 < p2 { return -1 }
    return 1
}

func ( id *Identifier ) Compare( id2 *Identifier ) int {
    l, r, swap := id, id2, 1
    if len( id.parts ) > len( id2.parts ) { l, r, swap = id2, id, -1 }
    for i, part := range l.parts {
        if cmp := compareIdParts( part, r.parts[ i ] ); cmp != 0 { 
            return swap * cmp 
        }
    }
    return swap * ( len( l.parts ) - len( r.parts ) )
}

func ( id *Identifier ) Equals( id2 *Identifier ) bool {
    return id2 != nil && id.Compare( id2 ) == 0
}

func ( id *Identifier ) dup() *Identifier {
    res := &Identifier{ parts: make( []string, len( id.parts ) ) }
    for i, part := range id.parts { res.parts[ i ] = part }
    return res
}

type idSort []*Identifier
func ( s idSort ) Len() int { return len( s ) }
func ( s idSort ) Less( i, j int ) bool { return s[ i ].Compare( s[ j ] ) < 0 }
func ( s idSort ) Swap( i, j int ) { s[ j ], s[ i ] = s[ i ], s[ j ] }

func SortIds( ids []*Identifier ) []*Identifier { 
    sort.Sort( idSort( ids ) ) 
    return ids
}

type Namespace struct {
    Parts []*Identifier
    Version *Identifier
}

func ( ns *Namespace ) formatToBuf( buf []byte, styl IdentifierFormat ) []byte {
    for i, id := range ns.Parts {
        if i > 0 { buf = append( buf, byte( ':' ) ) }
        buf = id.formatToBuf( buf, styl )
    }
    buf = append( buf, byte( '@' ) )
    return ns.Version.formatToBuf( buf, styl )
}

func ( ns *Namespace ) extFormToBuf( buf []byte ) []byte {
    return ns.formatToBuf( buf, LcCamelCapped )
}

func ( ns *Namespace ) ExternalForm() string {
    return string( ns.extFormToBuf( make( []byte, 0, 32 ) ) )
}

func ( ns *Namespace ) String() string { return ns.ExternalForm() }

func ( ns *Namespace ) Equals( ns2 *Namespace ) bool {
    if ns2 == nil { return false }
    if ns.Version.Equals( ns2.Version ) {
        f := func( i int ) bool { 
            return ns.Parts[ i ].Equals( ns2.Parts[ i ] ) 
        }
        return equalSlices( len( ns.Parts ), len( ns2.Parts ), f )
    }
    return false
}

// Interface for (DeclaredTypeName|QualifiedTypeName) 
type TypeName interface{
    ExternalForm() string
    Equals( n TypeName ) bool
    typeNameImpl()
}

type DeclaredTypeNameError struct { msg string }

func newDeclaredTypeNameError( msg string ) *DeclaredTypeNameError {
    return &DeclaredTypeNameError{ msg }
}

func ( e *DeclaredTypeNameError ) Error() string { return e.msg }

type DeclaredTypeName struct { nm string }

func NewDeclaredTypeNameUnsafe( nm string ) *DeclaredTypeName {
    return &DeclaredTypeName{ nm }
}

func ( dtn *DeclaredTypeName ) SetNameUnsafe( nm string ) { dtn.nm = nm }

var DeclaredTypeNameRegexp = regexp.MustCompile( "^([A-Z][a-z0-9]*)+$" )

func CreateDeclaredTypeName( nm string ) ( *DeclaredTypeName, error ) {
    if res := DeclaredTypeNameRegexp.FindStringIndex( nm ); res == nil {
        msg := fmt.Sprintf( "invalid type name: %q", nm )
        return nil, newDeclaredTypeNameError( msg )
    }
    return NewDeclaredTypeNameUnsafe( nm ), nil
} 

func ( n *DeclaredTypeName ) typeNameImpl() {}

func ( n *DeclaredTypeName ) String() string { return n.nm }

func ( n *DeclaredTypeName ) ExternalForm() string { return n.String() }

func ( n *DeclaredTypeName ) Equals( other TypeName ) bool {
    if n2, ok := other.( *DeclaredTypeName ); ok { return n.nm == n2.nm }
    return false
}

func ( n *DeclaredTypeName ) ResolveIn( ns *Namespace ) *QualifiedTypeName {
    return &QualifiedTypeName{ Namespace: ns, Name: n }
}

func equalSlices( lenL, lenR int, comp func( i int ) bool ) bool {
    if lenL == lenR {
        for i := 0; i < lenL; i++ { if ! comp( i ) { return false } }
        return true
    } 
    return false
}

type QualifiedTypeName struct {
    Namespace *Namespace
    Name *DeclaredTypeName
}

func ( qn *QualifiedTypeName ) typeNameImpl() {}

func ( qn *QualifiedTypeName ) ExternalForm() string {
    res := make( []byte, 0, 32 )
    res = append( res, []byte( qn.Namespace.ExternalForm() )... )
    res = append( res, byte( '/' ) )
    res = append( res, []byte( ( qn.Name.nm ) )... )
    return string( res )
}

func ( qn *QualifiedTypeName ) String() string { return qn.ExternalForm() }

func ( qn *QualifiedTypeName ) Equals( n2 TypeName ) bool {
    if n2 == nil { return false }
    if qn2, ok := n2.( *QualifiedTypeName ); ok {
        return qn.Namespace.Equals( qn2.Namespace ) &&
               qn.Name.Equals( qn2.Name )
    }
    return false
}

func ( qn *QualifiedTypeName ) AsAtomicType() *AtomicTypeReference {
    return NewAtomicTypeReference( qn, nil )
}

// (atomic|list|nullable)
type TypeReference interface {
    ExternalForm() string
    Equals( t TypeReference ) bool
    String() string
    typeRefImpl()
}

type ValueRestriction interface {

    ExternalForm() string

    // Will panic if val is nil or is not the correct type for this instance
    AcceptsValue( val Value ) bool

    // vr may be nil
    equalsRestriction( vr ValueRestriction ) bool
}

const (

    errMsgEmptyRange = "empty range"

    errMsgUnsatisfiableRange = "unsatisfiable range"

    rxErrNameRegex = "regex"

    rxErrNameRange = "range"
)

type RestrictionError struct { msg string }

func ( e *RestrictionError ) Error() string { return e.msg }

func newRestrictionErrorInapplicable( 
    rxName string, attempted *QualifiedTypeName ) error {

    msg := fmt.Sprintf( "%s restriction cannot be applied to %s", 
        rxName, attempted )
    return &RestrictionError{ msg }
}

func equalComparers( c1, c2 interface{} ) bool {
    if c1 == c2 { return true }
    if c1 == nil { return c2 == nil }
    if c2 == nil { return false }
    return c1.( Comparer ).Compare( c2 ) == 0
}

type RegexRestriction struct {
    src string
    exp *regexp.Regexp
}

func CreateRegexRestriction( src string ) ( *RegexRestriction, error ) {
    exp, err := regexp.Compile( src )
    if err == nil { return &RegexRestriction{ src, exp }, nil }
    return nil, &RestrictionError{ err.Error() }
}

func MustRegexRestriction( src string ) *RegexRestriction {
    res, err := CreateRegexRestriction( src )
    if err == nil { return res }
    panic( err )
}

func ( r *RegexRestriction ) ExternalForm() string {
    return fmt.Sprintf( "%q", string( r.src ) )
}

func ( r *RegexRestriction ) Source() string { return r.src }

func ( r *RegexRestriction ) equalsRestriction( vr ValueRestriction ) bool {
    if vr == nil { return false }
    if r == vr { return true }
    if r2, ok := vr.( *RegexRestriction ); ok { return r.src == r2.src }
    return false
}

var errNilVal error
func init() { errNilVal = fmt.Errorf( "Value is nil" ) }

func ( r *RegexRestriction ) AcceptsValue( val Value ) bool {
    if val == nil { panic( errNilVal ) }
    return r.exp.MatchString( string( val.( String ) ) )
}

func formatRange( 
    minClosed bool, min, max Value, maxClosed bool, must bool ) string {

    buf := bytes.Buffer{}
    if minClosed { buf.WriteRune( '[' ) } else { buf.WriteRune( '(' ) }
    if min != nil { buf.WriteString( quoteRangeValue( min, must ) ) }
    buf.WriteRune( ',' )
    if max != nil { buf.WriteString( quoteRangeValue( max, must ) ) }
    if maxClosed { buf.WriteRune( ']' ) } else { buf.WriteRune( ')' ) }
    return buf.String()
}

type RangeRestriction struct {
    minClosed bool
    min Value 
    max Value
    maxClosed bool
}

// because *RangeRestriction instances cannot be empty and must have min/max of
// the same type this always returns a valid non-nil type
func ( rr *RangeRestriction ) inferredType() *QualifiedTypeName {
    var val = rr.min
    if val == nil { val = rr.max }
    return TypeNameIn( TypeOf( val ) )
}

func ( rr *RangeRestriction ) MinClosed() bool { return rr.minClosed }
func ( rr *RangeRestriction ) Min() Value { return rr.min }
func ( rr *RangeRestriction ) Max() Value { return rr.max }
func ( rr *RangeRestriction ) MaxClosed() bool { return rr.maxClosed }

func quoteRangeValue( val Value, must bool ) string {
    switch v := val.( type ) {
    case Int32, Int64, Uint32, Uint64, Float32, Float64: 
        return val.( fmt.Stringer ).String()
    case String: return fmt.Sprintf( "%q", string( v ) )
    case Timestamp: return fmt.Sprintf( "%q", v.Rfc3339Nano() )
    }
    if must { panic( libErrorf( "unhandled range val for quote: %T", val ) ) }
    return fmt.Sprintf( "<range val: %T>", val )
}

func ( r *RangeRestriction ) ExternalForm() string {
    return formatRange( r.minClosed, r.min, r.max, r.maxClosed, true )
}

func ( r *RangeRestriction ) equalsRestriction( vr ValueRestriction ) bool {
    if vr == nil { return false }
    if r == vr { return true }
    if r2, ok := vr.( *RangeRestriction ); ok {
        // do cheap tests first
        return r.MinClosed() == r2.MinClosed() &&
               r.MaxClosed() == r2.MaxClosed() &&
               equalComparers( r.Min(), r2.Min() ) &&
               equalComparers( r.Max(), r2.Max() )
    }
    return false
}

func ( r *RangeRestriction ) AcceptsValue( val Value ) bool {
    if val == nil { panic( errNilVal ) }
    if r.Min() != nil {
        switch i := r.Min().( Comparer ).Compare( val ); {
        case i == 0: if ! r.MinClosed() { return false }
        case i > 0: return false
        }
    }
    if r.Max() != nil {
        switch i := r.Max().( Comparer ).Compare( val ); {
        case i == 0: if ! r.MaxClosed() { return false }
        case i < 0: return false
        }
    }
    return true
}

type RangeRestrictionBuilder struct {
    Type *QualifiedTypeName
    MinClosed bool
    Min Value 
    Max Value
    MaxClosed bool
}

func ( b *RangeRestrictionBuilder ) RangeToString() string {
    return formatRange( b.MinClosed, b.Min, b.Max, b.MaxClosed, false )
}

func ( b *RangeRestrictionBuilder ) checkApplicable() error {
    for _, rrt := range restrictableRangeTypeNames {
        if b.Type.Equals( rrt ) { return nil }
    }
    return newRestrictionErrorInapplicable( rxErrNameRange, b.Type )
}

func ( b *RangeRestrictionBuilder ) checkEmpty() error {
    if b.Min == nil && b.Max == nil {
        return &RestrictionError{ errMsgEmptyRange } 
    }
    return nil
}

func ( b *RangeRestrictionBuilder ) checkValueTypes() error {
    f := func( bound string, val Value ) error {
        if val == nil { return nil }
        qn := TypeNameIn( TypeOf( val ) )
        if qn.Equals( b.Type ) { return nil }
        msg := fmt.Sprintf( "illegal %s value of type %s in range of type %s",
            bound, qn, b.Type )
        return &RestrictionError{ msg }
    }
    if err := f( "min", b.Min ); err != nil { return err }
    return f( "max", b.Max )
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

func ( b *RangeRestrictionBuilder ) checkSatisfiable() error {
    // first check: an open or half-open non-empty range is always satisfiable
    if b.Min == nil || b.Max == nil { return nil }
    switch i := b.Min.( Comparer ).Compare( b.Max ); {
    case i == 0: if b.MinClosed && b.MaxClosed { return nil }
    case i > 0: ;
    case i < 0: 
        open := ! ( b.MinClosed || b.MaxClosed )
        adjacent := areAdjacentInts( b.Min, b.Max ) 
        if ! ( open && adjacent ) { return nil }
    }
    return &RestrictionError{ errMsgUnsatisfiableRange }
}

// order of checks here is important -- ex: satisfiability check assumes that
// range is known non-empty
func ( b *RangeRestrictionBuilder ) Build() ( *RangeRestriction, error ) {
    if err := b.checkEmpty(); err != nil { return nil, err }
    if err := b.checkApplicable(); err != nil { return nil, err }
    if err := b.checkValueTypes(); err != nil { return nil, err }
    if err := b.checkSatisfiable(); err != nil { return nil, err }
    return &RangeRestriction{ b.MinClosed, b.Min, b.Max, b.MaxClosed }, nil
}

type AtomicTypeReference struct {
    name *QualifiedTypeName
    restriction ValueRestriction
}

func NewAtomicTypeReference( 
    nm *QualifiedTypeName, vr ValueRestriction ) *AtomicTypeReference {

    return &AtomicTypeReference{ name: nm, restriction: vr }
}

func atomicCheckRestrictionApplicable( 
    nm *QualifiedTypeName, vr ValueRestriction ) error {

    switch v := vr.( type ) {
    case *RegexRestriction: 
        if nm.Equals( QnameString ) { return nil }
        return newRestrictionErrorInapplicable( rxErrNameRegex, nm )
    case *RangeRestriction: 
        rngTyp := v.inferredType()
        if nm.Equals( rngTyp ) { return nil }
        msg := fmt.Sprintf( "cannot apply %s range to base type %s", 
            rngTyp, nm )
        return &RestrictionError{ msg }
    }
    panic( libErrorf( "unhandled restriction: %T", vr ) )
}

func CreateAtomicTypeReference(
    nm *QualifiedTypeName, 
    vr ValueRestriction ) ( *AtomicTypeReference, error ) {

    if vr != nil {
        if err := atomicCheckRestrictionApplicable( nm, vr ); err != nil {
            return nil, err
        }
    }
    return &AtomicTypeReference{ name: nm, restriction: vr }, nil
}

func ( t *AtomicTypeReference ) Name() *QualifiedTypeName { return t.name }

func ( t *AtomicTypeReference ) Restriction() ValueRestriction {
    return t.restriction
}

func ( t *AtomicTypeReference ) typeRefImpl() {}

func ( t *AtomicTypeReference ) ExternalForm() string {
    nm := t.name.ExternalForm()
    if t.restriction == nil { return nm }
    return fmt.Sprintf( "%s~%s", nm, t.restriction.ExternalForm() )
}

func ( t *AtomicTypeReference ) String() string { return t.ExternalForm() }

func ( t *AtomicTypeReference ) Equals( t2 TypeReference ) bool {
    if t2 == nil { return false }
    if t == t2 { return true }
    if at2, ok := t2.( *AtomicTypeReference ); ok {
        if t.name.Equals( at2.name ) {
            if t.restriction == nil { return at2.restriction == nil }
            return t.restriction.equalsRestriction( at2.restriction )
        }
    }
    return false
}

type ListTypeReference struct {
    ElementType TypeReference
    AllowsEmpty bool
}

func ( t *ListTypeReference ) typeRefImpl() {}

func ( t *ListTypeReference ) ExternalForm() string {
    var quant string
    if t.AllowsEmpty { quant = "*" } else { quant = "+" }
    return t.ElementType.ExternalForm() + quant
}

func ( t *ListTypeReference ) String() string { return t.ExternalForm() }

func ( t *ListTypeReference ) Equals( ref TypeReference ) bool {
    if ref == nil { return false }
    if t2, ok := ref.( *ListTypeReference ); ok {
        return t.AllowsEmpty == t2.AllowsEmpty &&
               t.ElementType.Equals( t2.ElementType )
    }
    return false
}

type NullableTypeReference struct {
    Type TypeReference
}

// later versions may hold the offending type as a field
type NullableTypeError struct {}

func ( nte *NullableTypeError ) Error() string { return "not a nullable type" }

func NewNullableTypeError( typ TypeReference ) *NullableTypeError {
    return &NullableTypeError{}
}

func IsNullableType( typ TypeReference ) bool {
    switch v := typ.( type ) {
    case *ListTypeReference: return true;
    case *NullableTypeReference: return false;
    case *PointerTypeReference: return true;
    case *AtomicTypeReference:
        if ! v.Name().Namespace.Equals( CoreNsV1 ) { return false }
        return ! ( v.Name().Equals( QnameBoolean ) || 
                   v.Name().Equals( QnameTimestamp ) || 
                   IsNumericTypeName( v.Name() ) )
    }
    panic( libErrorf( "unhandled type: %T", typ ) )
}

func MustNullableTypeReference( typ TypeReference ) *NullableTypeReference {
    if ! IsNullableType( typ ) {
        panic( NewNullableTypeError( typ ) )
    }
    return &NullableTypeReference{ Type: typ }
}

func ( t *NullableTypeReference ) typeRefImpl() {}

func ( t *NullableTypeReference ) ExternalForm() string {
    return t.Type.ExternalForm() + "?"
}

func ( t *NullableTypeReference ) String() string { return t.ExternalForm() }

func ( typ *NullableTypeReference ) Equals( ref TypeReference ) bool {
    if ref == nil { return false }
    if typ2, ok := ref.( *NullableTypeReference ); ok {
        return typ.Type.Equals( typ2.Type )
    }
    return false
}

type PointerTypeReference struct { Type TypeReference }

func NewPointerTypeReference( typ TypeReference ) *PointerTypeReference {
    return &PointerTypeReference{ Type: typ }
}

func ( pt *PointerTypeReference ) Equals( ref TypeReference ) bool {
    if ref == nil { return false }
    if pt2, ok := ref.( *PointerTypeReference ); ok {
        return pt.Type.Equals( pt2.Type )
    }
    return false
}

func ( pt *PointerTypeReference ) ExternalForm() string {
    return "&(" + pt.Type.ExternalForm() + ")"
}

func ( pt *PointerTypeReference ) String() string { return pt.ExternalForm() }

func ( pt *PointerTypeReference ) typeRefImpl() {}

func AtomicTypeIn( ref TypeReference ) *AtomicTypeReference {
    switch v := ref.( type ) {
    case *AtomicTypeReference: return v
    case *ListTypeReference: return AtomicTypeIn( v.ElementType )
    case *NullableTypeReference: return AtomicTypeIn( v.Type )
    case *PointerTypeReference: return AtomicTypeIn( v.Type )
    }
    panic( fmt.Errorf( "No atomic type in %s (%T)", ref, ref ) )
}

func TypeNameIn( typ TypeReference ) *QualifiedTypeName {
    return AtomicTypeIn( typ ).Name()
}

type Value interface{ valImpl() }

type goValPath objpath.PathNode // keys are string

var inValPathRoot objpath.PathNode
func init() { inValPathRoot = objpath.RootedAt( "inVal" ) }

type ValueTypeError struct { 
    loc goValPath
    msg string 
}

var goValPathFormatter objpath.Formatter

func init() {
    goValPathFormatter = 
        objpath.DotFormatter(
            func( elt interface{}, apnd objpath.AppendFunc ) {
                apnd( elt.( string ) )
            },
        )
}

func ( e *ValueTypeError ) Error() string { 
    if e.loc == nil { return e.msg }
    locStr := objpath.Format( e.loc, goValPathFormatter )
    return locStr + ": " + e.msg 
}

type ValueError interface {
    Location() objpath.PathNode
    Message() string 
    error
}

// path is an idPath
func FormatError( path objpath.PathNode, msg string ) string {
    if path == nil { return msg }
    return FormatIdPath( path ) + ": " + msg 
}

type InputError struct {
    Location objpath.PathNode
    Message string
}

func ( e *InputError ) Error() string { 
    return FormatError( e.Location, e.Message )
}

func NewInputError( path idPath, msg string ) *InputError {
    return &InputError{ Message: msg, Location: path }
}

func NewInputErrorf( path idPath, tmpl string, args ...interface{} ) *InputError {
    return NewInputError( path, fmt.Sprintf( tmpl, args... ) )
}

func NewTypeInputError( 
    expct, act TypeReference, path objpath.PathNode ) *InputError {

    return NewInputErrorf( 
        path,
        "Expected value of type %s but found %s",
        expct.ExternalForm(), act.ExternalForm(),
    )
}

func NewTypeInputErrorValue( 
    t TypeReference, val Value, path objpath.PathNode ) *InputError {

    return NewTypeInputError( t, TypeOf( val ), path )
}

type Comparer interface {

    // Should panic (such as with type assertion panic) if val is not same type
    // as the instance or is nil
    Compare( val interface{} ) int
}

type String string
func ( s String ) valImpl() {}
func ( s String ) String() string { return string( s ) }

func ( s String ) Compare( val interface{} ) int {
    switch s2 := val.( String ); {
    case s < s2: return -1
    case s == s2: return 0
    }
    return 1
}

type Boolean bool
func ( b Boolean ) valImpl() {}
func ( b Boolean ) String() string { return fmt.Sprint( bool( b ) ) }

type Int64 int64
func ( i Int64 ) valImpl() {}
func ( i Int64 ) String() string { return fmt.Sprint( int64( i ) ) }

func ( i Int64 ) Compare( val interface{} ) int {
    switch v := val.( Int64 ); {
    case i < v: return -1
    case i > v: return 1
    }
    return 0
}

type Int32 int32

func ( i Int32 ) valImpl() {}
func ( i Int32 ) String() string { return fmt.Sprint( int32( i ) ) }

func ( i Int32 ) Compare( val interface{} ) int {
    return Int64( i ).Compare( Int64( val.( Int32 ) ) )
}

type Uint64 uint64
func ( i Uint64 ) valImpl() {}
func ( i Uint64 ) String() string { return fmt.Sprint( uint64( i ) ) }

func ( i Uint64 ) Compare( val interface{} ) int {
    switch v := val.( Uint64 ); {
    case i < v: return -1
    case i > v: return 1
    }
    return 0
}

type Uint32 uint32
func ( i Uint32 ) valImpl() {}
func ( i Uint32 ) String() string { return fmt.Sprint( uint32( i ) ) }

func ( i Uint32 ) Compare( val interface{} ) int {
    return Uint64( i ).Compare( Uint64( val.( Uint32 ) ) )
}

type Float64 float64
func ( d Float64 ) valImpl() {}
func ( d Float64 ) String() string { return fmt.Sprint( float64( d ) ) }

func ( d Float64 ) Compare ( val interface{} ) int {
    switch d2 := val.( Float64 ); {
    case d < d2: return -1
    case d == d2: return 0
    }
    return 1
}

type Float32 float32
func ( f Float32 ) valImpl() {}
func ( f Float32 ) String() string { return fmt.Sprint( float32( f ) ) }

func ( f Float32 ) Compare( val interface{} ) int {
    return Float64( f ).Compare( Float64( val.( Float32 ) ) )
}

type Buffer []byte
func ( b Buffer ) valImpl() {}

type Null struct {}
func ( n Null ) valImpl() {}
var NullVal *Null
func init() { NullVal = &Null{} }

func IsNull( val Value ) bool {
    _, isNull := val.( *Null )
    return isNull
}

type Timestamp time.Time

func ( t Timestamp ) valImpl() {}

func Now() Timestamp { return Timestamp( time.Now() ) }

func ( t Timestamp ) Rfc3339Nano() string {
    return time.Time( t ).Format( time.RFC3339Nano )
}

func ( t Timestamp ) String() string { return t.Rfc3339Nano() }

func ( t Timestamp ) Compare( val interface{} ) int {
    tm1 := time.Time( t )
    switch tm2 := time.Time( val.( Timestamp ) ); {
    case tm1.Before( tm2 ): return -1
    case tm1.After( tm2 ): return 1
    }
    return 0
}

func equalMaps( m1, m2 *SymbolMap ) bool {
    if m1.Len() != m2.Len() { return false }
    res := true
    m1.EachPair( func( fld *Identifier, val Value ) {
        if ! res { return }
        if val2, ok := m2.GetOk( fld ); ok {
            res = EqualValues( val, val2 )
        } else { res = false }
    })
    return res
}

func equalStructs( s1, s2 *Struct ) bool {
    return s1.Type.Equals( s2.Type ) && equalMaps( s1.Fields, s2.Fields )
}

func equalLists( l1, l2 *List ) bool {
    if ! l1.Type.Equals( l2.Type ) { return false }
    if len( l1.vals ) != len( l2.vals ) { return false }
    for i, v1 := range l1.vals {
        if ! EqualValues( v1, l2.vals[ i ] ) { return false }
    }
    return true
}

func EqualValues( v1, v2 Value ) bool {
    switch v := v1.( type ) {
    case *Null: if _, ok := v2.( *Null ); ok { return true }
    case Boolean: if b, ok := v2.( Boolean ); ok { return v == b }
    case Int32: if i, ok := v2.( Int32 ); ok { return v == i }
    case Uint32: if i, ok := v2.( Uint32 ); ok { return v == i }
    case Int64: if i, ok := v2.( Int64 ); ok { return v == i }
    case Uint64: if i, ok := v2.( Uint64 ); ok { return v == i }
    case Float32: if i, ok := v2.( Float32 ); ok { return v == i }
    case Float64: if i, ok := v2.( Float64 ); ok { return v == i }
    case Buffer: if b, ok := v2.( Buffer ); ok { return bytes.Equal( v, b ) }
    case String: if s, ok := v2.( String ); ok { return v == s }
    case Timestamp:
        if t, ok := v2.( Timestamp ); ok { return v.Compare( t ) == 0 }
    case *Enum: 
        if e, ok := v2.( *Enum ); ok { 
            return v.Type.Equals( e.Type ) && v.Value.Equals( e.Value )
        }
    case *SymbolMap: 
        if m, ok := v2.( *SymbolMap ); ok { return equalMaps( v, m ) }
    case *Struct: if s, ok := v2.( *Struct ); ok { return equalStructs( v, s ) }
    case *List: if l, ok := v2.( *List ); ok { return equalLists( v, l ) }
    }
    return false
}

func asListValue( inVals []interface{}, path goValPath ) ( *List, error ) {
    vals := make( []Value, len( inVals ) )
    lp := path.StartList()
    for i, inVal := range inVals { 
        var err error
        if vals[ i ], err = implAsValue( inVal, lp ); err != nil { 
            return nil, err 
        }
        lp = lp.Next()
    }
    return &List{ vals: vals, Type: TypeOpaqueList }, nil
}

func asAtomicValue( 
    inVal interface{}, path goValPath ) ( val Value, err error ) {
    if inVal == nil { return NullVal, nil }
    switch v := inVal.( type ) {
    case string: val = String( v )
    case String: val = v
    case bool: val = Boolean( v )
    case Boolean: val = v
    case []byte: val = Buffer( v )
    case Buffer: val = v
    case int8: val = Int32( int32( v ) )
    case int16: val = Int32( int32( v ) )
    case int32: val = Int32( v )
    case Int32: val = v
    case int: val = Int64( int64( v ) )
    case int64: val = Int64( v )
    case Int64: val = v
    case uint64: val = Uint64( v )
    case Uint64: val = v
    case uint32: val = Uint32( v )
    case Uint32: val = v
    case float32: val = Float32( v )
    case Float32: val = v
    case float64: val = Float64( v )
    case Float64: val = v
    case Timestamp: val = v
    case time.Time: val = Timestamp( v )
    case *List: val = v
    case *SymbolMap: val = v
    case *Enum: val = v
    case *Struct: val = v
    case *Null: val = v
    default:
        msg := "Unhandled mingle value %v (%T)"
        err = &ValueTypeError{ path, fmt.Sprintf( msg, inVal, inVal ) }
    }
    return
}

func implAsValue( inVal interface{}, path goValPath ) ( Value, error ) {
    switch v := inVal.( type ) {
    case []interface{}: return asListValue( v, path )
    }
    return asAtomicValue( inVal, path )
}

func AsValue( inVal interface{} ) ( Value, error ) {
    return implAsValue( inVal, inValPathRoot )
}

func MustValue( inVal interface{} ) Value {
    val, err := AsValue( inVal )
    if err != nil { panic( err ) }
    return val
}

type List struct { 
    Type *ListTypeReference
    vals []Value 
}

func NewList( typ *ListTypeReference ) *List { 
    return &List{ vals: []Value{}, Type: typ } 
}

func NewListValues( vals []Value ) *List { 
    return &List{ vals: vals, Type: TypeOpaqueList }
}

// if we allow immutable lists later we can have this return a fixed immutable
// empty instance
func EmptyList() *List { return NewList( TypeOpaqueList ) }

func ( l *List ) AddUnsafe( val Value ) { l.vals = append( l.vals, val ) }

func ( l *List ) valImpl() {}

// returned slice is live at least as long as the next call to l.Add()
func ( l *List ) Values() []Value { return l.vals }

func ( l *List ) Get( idx int ) Value { return l.vals[ idx ] }

func ( l *List ) Set( v Value, idx int ) { l.vals[ idx ] = v }

func ( l *List ) Len() int { return len( l.vals ) }

func CreateList( vals ...interface{} ) ( *List, error ) {
    res := &List{ Type: TypeOpaqueList }
    if len( vals ) > 0 {
        if typ, ok := vals[ 0 ].( TypeReference ); ok {
            if lt, ok := typ.( *ListTypeReference ); ok {
                res.Type = lt
                vals = vals[ 1 : ]
            } else {
                return nil, fmt.Errorf( "first arg not a list type: %s", typ )
            }
        }
    }
    res.vals = make( []Value, len( vals ) )
    for i, val := range vals {
        var err error
        if res.vals[ i ], err = AsValue( val ); err != nil { return nil, err }
    }
    return res, nil
}

func MustList( vals ...interface{} ) *List {
    res, err := CreateList( vals... )
    if err != nil { panic( err ) }
    return res
}

type SymbolMap struct {
    m *IdentifierMap
}

func NewSymbolMap() *SymbolMap { return &SymbolMap{ NewIdentifierMap() } }

// Later if we decide to make read-only variants of *SymbolMap we could return a
// single instance here
func EmptySymbolMap() *SymbolMap { return NewSymbolMap() }

func ( m *SymbolMap ) valImpl() {}

func ( m *SymbolMap ) Len() int { return m.m.Len() }

func ( m *SymbolMap ) EachPairError ( 
    f func( *Identifier, Value ) error ) error {

    return m.m.EachPairError( func( fld *Identifier, val interface{} ) error {
        return f( fld, val.( Value ) )
    })
}

func ( m *SymbolMap ) EachPair( f func( *Identifier, Value ) ) {
    m.m.EachPair( func( fld *Identifier, v interface{} ) {
        f( fld, v.( Value ) )
    })
}

func ( m *SymbolMap ) GetKeys() []*Identifier { return m.m.GetKeys() }

func ( m *SymbolMap ) GetOk( fld *Identifier ) ( Value, bool ) {
    if val, ok := m.m.GetOk( fld ); ok { return val.( Value ), true }
    return nil, false
}

func ( m *SymbolMap ) Get( fld *Identifier ) Value {
    if val, ok := m.GetOk( fld ); ok { return val }
    return nil
}

func ( m *SymbolMap ) HasKey( fld *Identifier ) bool {
    return m.m.HasKey( fld )
}

func ( m *SymbolMap ) Put( fld *Identifier, val Value ) { m.m.Put( fld, val ) }

type MapLiteralError struct { msg string }

func ( e *MapLiteralError ) Error() string { return e.msg } 

func mapLiteralErrorf( fmtStr string, args ...interface{} ) *MapLiteralError {
    return &MapLiteralError{ fmt.Sprintf( fmtStr, args... ) }
}

func makePairError( err error, indx int ) error {
    return mapLiteralErrorf( "error in map literal pairs at index %d: %s",
        indx, err )
}

func createSymbolMapEntry( 
    pairs []interface{}, idx int ) ( fld *Identifier, val Value, err error ) {

    fldIdx, valIdx := idx, idx + 1
    fldIdVal, ok := pairs[ fldIdx ], false
    if fld, ok = fldIdVal.( *Identifier ); ! ok {
        err = makePairError(
            fmt.Errorf( "invalid key type: %T", fldIdVal ), idx )
        return
    }
    if val, err = AsValue( pairs[ valIdx ] ); err != nil { 
        err = makePairError( err, valIdx )
        return
    }
    return
}

func CreateSymbolMap( pairs ...interface{} ) ( m *SymbolMap, err error ) {
    if pLen := len( pairs ); pLen % 2 == 1 { 
        err = mapLiteralErrorf( "invalid pairs len: %d", pLen )
    } else { 
        m = NewSymbolMap()
        var fld *Identifier
        var val Value
        for i := 0; i < pLen; i += 2 {
            fld, val, err = createSymbolMapEntry( pairs, i )
            if err != nil { return }
            if m.HasKey( fld ) {
                tmpl := "duplicate entry for '%s' starting at index %d"
                err = mapLiteralErrorf( tmpl, fld, i )
                return
            }
            m.Put( fld, val )
        }
    }
    return
}

func MustSymbolMap( pairs ...interface{} ) *SymbolMap {
    res, err := CreateSymbolMap( pairs... )
    if err != nil { panic( err ) }
    return res
}

type Enum struct {
    Type *QualifiedTypeName
    Value *Identifier
}

func ( e *Enum ) valImpl() {}

// In the go code we deal with *Struct, but even though these are go pointers,
// they correspond to struct values in the mingle runtime. 
type Struct struct {
    Type *QualifiedTypeName
    Fields *SymbolMap
}

func NewStruct( typ *QualifiedTypeName ) *Struct {
    return &Struct{ Type: typ, Fields: NewSymbolMap() }
}

func ( s *Struct ) valImpl() {}

func CreateStruct(
    typ *QualifiedTypeName, pairs ...interface{} ) ( *Struct, error ) {

    res := &Struct{ Type: typ }
    if flds, err := CreateSymbolMap( pairs... ); err == nil {
        res.Fields = flds
    } else { return nil, err }
    return res, nil
}

func MustStruct( typ *QualifiedTypeName, pairs ...interface{} ) *Struct {
    res, err := CreateStruct( typ, pairs... )
    if err != nil { panic( err ) }
    return res
}

type idPath objpath.PathNode // elts are *Identifier

var idPathRootVal idPath
func init() { 
    idPathRootVal = objpath.RootedAt( NewIdentifierUnsafe( []string{ "val" } ) )
}

var idPathFormatter objpath.Formatter

func init() {
    f := func( elt interface{}, apnd objpath.AppendFunc ) {
        apnd( elt.( *Identifier ).ExternalForm() )
    }
    idPathFormatter = objpath.DotFormatter( f )
}

func FormatIdPath( p objpath.PathNode ) string {
    return objpath.Format( p, idPathFormatter )
}

var (
    QnameBoolean *QualifiedTypeName
    TypeBoolean *AtomicTypeReference
    QnameBuffer *QualifiedTypeName
    TypeBuffer *AtomicTypeReference
    QnameString *QualifiedTypeName
    TypeString *AtomicTypeReference
    QnameInt32 *QualifiedTypeName
    TypeInt32 *AtomicTypeReference
    QnameInt64 *QualifiedTypeName
    TypeInt64 *AtomicTypeReference
    QnameUint32 *QualifiedTypeName
    TypeUint32 *AtomicTypeReference
    QnameUint64 *QualifiedTypeName
    TypeUint64 *AtomicTypeReference
    QnameFloat32 *QualifiedTypeName
    TypeFloat32 *AtomicTypeReference
    QnameFloat64 *QualifiedTypeName
    TypeFloat64 *AtomicTypeReference
    QnameTimestamp *QualifiedTypeName
    TypeTimestamp *AtomicTypeReference
    QnameSymbolMap *QualifiedTypeName
    TypeSymbolMap *AtomicTypeReference
    QnameNull *QualifiedTypeName
    TypeNull *AtomicTypeReference
    QnameValue *QualifiedTypeName
    TypeValue *AtomicTypeReference
    TypeNullableValue *NullableTypeReference
    TypeOpaqueList *ListTypeReference
    QnameIdentifier *QualifiedTypeName
    TypeIdentifier *AtomicTypeReference
    QnameIdentifierPart *QualifiedTypeName
    TypeIdentifierPart *AtomicTypeReference
    QnameDeclaredTypeName *QualifiedTypeName
    TypeDeclaredTypeName *AtomicTypeReference
    QnameNamespace *QualifiedTypeName
    TypeNamespace *AtomicTypeReference
    QnameQualifiedTypeName *QualifiedTypeName
    TypeQualifiedTypeName *AtomicTypeReference
    QnameAtomicTypeReference *QualifiedTypeName
    TypeAtomicTypeReference *AtomicTypeReference
    QnameRangeRestriction *QualifiedTypeName
    TypeRangeRestriction *AtomicTypeReference
    QnameRegexRestriction *QualifiedTypeName
    TypeRegexRestriction *AtomicTypeReference
    QnameValueRestriction *QualifiedTypeName
    TypeValueRestriction *AtomicTypeReference
    QnameListTypeReference *QualifiedTypeName
    TypeListTypeReference *AtomicTypeReference
    QnameNullableTypeReference *QualifiedTypeName
    TypeNullableTypeReference *AtomicTypeReference
    QnamePointerTypeReference *QualifiedTypeName
    TypePointerTypeReference *AtomicTypeReference
    QnameTypeReference *QualifiedTypeName
    TypeTypeReference *AtomicTypeReference
    QnameIdentifierPath *QualifiedTypeName
    TypeIdentifierPath *AtomicTypeReference
    QnameIdentifierPathPart *QualifiedTypeName
    TypeIdentifierPathPart *AtomicTypeReference
    QnameStandardError *QualifiedTypeName
    TypeStandardError *AtomicTypeReference
    QnameInputError *QualifiedTypeName
    TypeInputError *AtomicTypeReference
    QnameUnrecognizedFieldError *QualifiedTypeName
    TypeUnrecognizedFieldError *AtomicTypeReference
    QnameMissingFieldsError *QualifiedTypeName
    TypeMissingFieldsError *AtomicTypeReference
)

var coreQnameResolver map[ string ]*QualifiedTypeName
var PrimitiveTypes []*AtomicTypeReference
var NumericTypeNames []*QualifiedTypeName
var restrictableRangeTypeNames []*QualifiedTypeName

var CoreNsV1 *Namespace

func init() {
    id := func( s string ) *Identifier {
        return NewIdentifierUnsafe( []string{ s } )
    }
    CoreNsV1 = &Namespace{
        Parts: []*Identifier{ id( "mingle" ), id( "core" ) },
        Version: id( "v1" ),
    }
    coreQnameResolver = make( map[ string ]*QualifiedTypeName )
    f1 := func( s string ) ( *QualifiedTypeName, *AtomicTypeReference ) {
        qn := &QualifiedTypeName{ CoreNsV1, &DeclaredTypeName{ s } }
        coreQnameResolver[ qn.Name.ExternalForm() ] = qn
        return qn, qn.AsAtomicType()
    }
    QnameBoolean, TypeBoolean = f1( "Boolean" )
    QnameBuffer, TypeBuffer = f1( "Buffer" )
    QnameString, TypeString = f1( "String" )
    QnameInt32, TypeInt32 = f1( "Int32" )
    QnameInt64, TypeInt64 = f1( "Int64" )
    QnameUint32, TypeUint32 = f1( "Uint32" )
    QnameUint64, TypeUint64 = f1( "Uint64" )
    QnameFloat32, TypeFloat32 = f1( "Float32" )
    QnameFloat64, TypeFloat64 = f1( "Float64" )
    QnameTimestamp, TypeTimestamp = f1( "Timestamp" )
    QnameValue, TypeValue = f1( "Value" )
    QnameSymbolMap, TypeSymbolMap = f1( "SymbolMap" )
    QnameNull, TypeNull = f1( "Null" )
    PrimitiveTypes = []*AtomicTypeReference{
        TypeNull,
        TypeString,
        TypeFloat64,
        TypeFloat32,
        TypeInt64,
        TypeInt32,
        TypeUint32,
        TypeUint64,
        TypeBoolean,
        TypeTimestamp,
        TypeBuffer,
        TypeSymbolMap,
    }
    TypeNullableValue = &NullableTypeReference{ TypeValue }
    TypeOpaqueList = &ListTypeReference{ TypeNullableValue, true }
    NumericTypeNames = []*QualifiedTypeName{
        QnameInt32,
        QnameInt64,
        QnameUint32,
        QnameUint64,
        QnameFloat32,
        QnameFloat64,
    }
    restrictableRangeTypeNames = []*QualifiedTypeName{
        QnameString,
        QnameTimestamp,
        QnameInt32,
        QnameUint32,
        QnameInt64,
        QnameUint64,
        QnameFloat32,
        QnameFloat64,
    }
    QnameIdentifier, TypeIdentifier = f1( "Identifier" )
    QnameIdentifierPart, TypeIdentifierPart = f1( "IdentifierPart" )
    QnameDeclaredTypeName, TypeDeclaredTypeName = f1( "DeclaredTypeName" )
    QnameNamespace, TypeNamespace = f1( "Namespace" )
    QnameQualifiedTypeName, TypeQualifiedTypeName = f1( "QualifiedTypeName" )
    QnameAtomicTypeReference, TypeAtomicTypeReference = 
        f1( "AtomicTypeReference" )
    QnameRangeRestriction, TypeRangeRestriction = f1( "RangeRestriction" )
    QnameRegexRestriction, TypeRegexRestriction = f1( "RegexRestriction" )
    QnameValueRestriction, TypeValueRestriction = f1( "ValueRestriction" )
    QnameListTypeReference, TypeListTypeReference = f1( "ListTypeReference" )
    QnameNullableTypeReference, TypeNullableTypeReference = 
        f1( "NullableTypeReference" )
    QnamePointerTypeReference, TypePointerTypeReference = 
        f1( "PointerTypeReference" )
    QnameTypeReference, TypeTypeReference = f1( "TypeReference" )
    QnameIdentifierPath, TypeIdentifierPath = f1( "IdentifierPath" )
    QnameIdentifierPathPart, TypeIdentifierPathPart = f1( "IdentifierPathPart" )
    QnameStandardError, TypeStandardError = f1( "StandardError" )
    QnameInputError, TypeInputError = f1( "InputError" )
    QnameUnrecognizedFieldError, TypeUnrecognizedFieldError = 
        f1( "UnrecognizedFieldError" )
    QnameMissingFieldsError, TypeMissingFieldsError = f1( "MissingFieldsError" )
}

func ResolveInCore( nm *DeclaredTypeName ) ( *QualifiedTypeName, bool ) {
    qn, ok := coreQnameResolver[ nm.ExternalForm() ]
    return qn, ok
}

func IsIntegerTypeName( qn *QualifiedTypeName ) bool {
    return qn.Equals( QnameInt32 ) || 
           qn.Equals( QnameInt64 ) ||
           qn.Equals( QnameUint32 ) ||
           qn.Equals( QnameUint64 )
}

func IsNumericTypeName( qn *QualifiedTypeName ) bool {
    for _, nm := range NumericTypeNames { if nm.Equals( qn ) { return true } }
    return false
}

func TypeOf( mgVal Value ) TypeReference {
    switch v := mgVal.( type ) {
    case Boolean: return TypeBoolean
    case Buffer: return TypeBuffer
    case String: return TypeString
    case Int32: return TypeInt32
    case Int64: return TypeInt64
    case Uint32: return TypeUint32
    case Uint64: return TypeUint64
    case Float32: return TypeFloat32
    case Float64: return TypeFloat64
    case Timestamp: return TypeTimestamp
    case *Enum: return v.Type.AsAtomicType()
    case *SymbolMap: return TypeSymbolMap
    case *Struct: return v.Type.AsAtomicType()
    case *List: return v.Type
    case *Null: return TypeNull
    }
    panic( libErrorf( "unhandled arg to typeOf (%T): %v", mgVal, mgVal ) )
}

func canAssignAtomic( 
    val Value, at *AtomicTypeReference, useRestriction bool ) bool {

    if at.Name().Equals( QnameNull ) { return true }
    if _, ok := val.( *Null ); ok { return false }
    if at.Name().Equals( QnameValue ) { return true }
    switch vt := TypeOf( val ).( type ) {
    case *AtomicTypeReference:
        if ! vt.Name().Equals( at.Name() ) { return false }
        if at.Restriction() == nil || ( ! useRestriction ) { return true }
        return at.Restriction().AcceptsValue( val )
    }
    return false
}

func CanAssign( val Value, typ TypeReference, useRestriction bool ) bool {
    switch t := typ.( type ) {
    case *AtomicTypeReference: return canAssignAtomic( val, t, useRestriction )
    case *PointerTypeReference: return false
    case *NullableTypeReference:
        if _, ok := val.( *Null ); ok { return true }
        return CanAssign( val, t.Type, useRestriction )
    case *ListTypeReference:
        if l, ok := val.( *List ); ok { return t.Equals( l.Type ) }
    default: panic( libErrorf( "unhandled type for assign: %T", typ ) )
    }
    return false
}

func canAssignAtomicType( 
    from TypeReference, to *AtomicTypeReference, relaxRestrictions bool ) bool {
    if to.Name().Equals( QnameValue ) { return true }
    f, ok := from.( *AtomicTypeReference );
    if ! ok { return false }
    if ! f.Name().Equals( to.Name() ) { return false }
    if relaxRestrictions {
        if to.Restriction() == nil { return true }
        // f.Restriction() could still be nil, so we make it the operand
        return to.Restriction().equalsRestriction( f.Restriction() )
    }
    if f.Restriction() == nil { return to.Restriction() == nil }
    return f.Restriction().equalsRestriction( to.Restriction() )
}

func canAssignNullableType( 
    from TypeReference, 
    to *NullableTypeReference, 
    relaxRestrictions bool ) bool {

    if f, ok := from.( *NullableTypeReference ); ok { from = f.Type }
    return canAssignType( from, to.Type, relaxRestrictions )
}

func canAssignPointerType( from TypeReference, to *PointerTypeReference ) bool {
    return from.Equals( to )
}

// A simple rigid check. Because lists are mutable, both element types and
// emptiability must match, other changes would be allowed in one side of the
// assignment that could not be allowed by the other
func canAssignListType( from TypeReference, to *ListTypeReference ) bool {
    return from.Equals( to )
}

func canAssignType( from, to TypeReference, relaxRestrictions bool ) bool {
    switch t := to.( type ) {
    case *AtomicTypeReference: 
        return canAssignAtomicType( from, t, relaxRestrictions )
    case *NullableTypeReference: 
        return canAssignNullableType( from, t, relaxRestrictions )
    case *PointerTypeReference: return canAssignPointerType( from, t )
    case *ListTypeReference: return canAssignListType( from, t )
    default: panic( libErrorf( "unhandled type: %T", to ) )
    }
    return false
}

func CanAssignType( from, to TypeReference ) bool {
    return canAssignType( from, to, true )
}

type NumberFormatError struct { msg string }

func ( e *NumberFormatError ) Error() string { return e.msg }

func newNumberRangeError( in string ) *NumberFormatError {
    msg := fmt.Sprintf( "value out of range: %s", in )
    return &NumberFormatError{ msg: msg }
}

func newNumberSyntaxError( 
    qn *QualifiedTypeName, in string ) *NumberFormatError {

    return &NumberFormatError{ msg: fmt.Sprintf( "invalid %s: %s", qn, in ) }
}

func parseIntNumberInitial(
    s string,
    bitSize int,
    numType *QualifiedTypeName ) ( sInt int64, uInt uint64, err error ) {

    if numType.Equals( QnameUint32 ) || numType.Equals( QnameUint64 ) {
        if len( s ) > 0 && s[ 0 ] == '-' {
            err = newNumberRangeError( s )
        } else {
            uInt, err = strconv.ParseUint( s, 10, bitSize )
            sInt = int64( uInt ) // do this even if err != nil
        }
    } else {
        sInt, err = strconv.ParseInt( s, 10, bitSize )
        uInt = uint64( sInt )
    }
    return
}

func parseIntNumber( 
    s string, bitSize int, numTyp *QualifiedTypeName ) ( Value, error ) {

    sInt, uInt, parseErr := parseIntNumberInitial( s, bitSize, numTyp )
    if parseErr != nil { return nil, parseErr }
    switch {
    case numTyp.Equals( QnameInt32 ): return Int32( sInt ), nil
    case numTyp.Equals( QnameUint32 ): return Uint32( uInt ), nil
    case numTyp.Equals( QnameInt64 ): return Int64( sInt ), nil
    case numTyp.Equals( QnameUint64 ): return Uint64( uInt ), nil
    }
    panic( libErrorf( "unhandled num type: %s", numTyp ) )
}

func parseFloatNumber(
    s string, bitSize int, numTyp *QualifiedTypeName ) ( Value, error ) {

    f, err := strconv.ParseFloat( string( s ), bitSize )
    if err != nil { return nil, err }
    switch {
    case numTyp.Equals( QnameFloat32 ): return Float32( f ), nil
    case numTyp.Equals( QnameFloat64 ): return Float64( f ), nil
    }
    panic( libErrorf( "unhandled num type: %s", numTyp ) )
}

func asParseNumberError( qn *QualifiedTypeName, s string, err error ) error {
    ne, ok := err.( *strconv.NumError )
    if ! ok { return err }
    switch {
    case ne.Err == strconv.ErrRange: err = newNumberRangeError( s )
    case ne.Err == strconv.ErrSyntax: err = newNumberSyntaxError( qn, s )
    }
    return err
}

func ParseNumber( s string, qn *QualifiedTypeName ) ( val Value, err error ) {
    switch {
    case qn.Equals( QnameInt32 ): val, err = parseIntNumber( s, 32, qn )
    case qn.Equals( QnameUint32 ): val, err = parseIntNumber( s, 32, qn )
    case qn.Equals( QnameInt64 ): val, err = parseIntNumber( s, 64, qn )
    case qn.Equals( QnameUint64 ): val, err = parseIntNumber( s, 64, qn )
    case qn.Equals( QnameFloat32 ): val, err = parseFloatNumber( s, 32, qn )
    case qn.Equals( QnameFloat64 ): val, err = parseFloatNumber( s, 64, qn )
    default: panic( libErrorf( "unhandled number type: %s", qn ) )
    }
    if err != nil { err = asParseNumberError( qn, s, err ) }
    return
}

type MissingFieldsError struct {
    Message string
    Location objpath.PathNode
    flds []*Identifier // stored sorted
}

func ( e *MissingFieldsError ) SetFields( flds []*Identifier ) {

    e.flds = make( []*Identifier, len( flds ) )
    copy( e.flds, flds )
    SortIds( e.flds )
}

func NewMissingFieldsError( 
    path objpath.PathNode, flds []*Identifier ) *MissingFieldsError {

    res := &MissingFieldsError{ Location: path }
    res.SetFields( flds )
    return res
}

func idSliceToString( flds []*Identifier ) string {
    strs := make( []string, len( flds ) )
    for i, fld := range flds { strs[ i ] = fld.ExternalForm() }
    return strings.Join( strs, ", " )
}

func ( e *MissingFieldsError ) Error() string {
    msg := e.Message
    if msg == "" { 
        msg = fmt.Sprintf( "missing field(s): %s", idSliceToString( e.flds ) )
    }
    return FormatError( e.Location, msg )
}

func ( e *MissingFieldsError ) Fields() []*Identifier { return e.flds }

type UnrecognizedFieldError struct {
    Message string
    Location objpath.PathNode
    Field *Identifier
}

func NewUnrecognizedFieldError( 
    p objpath.PathNode, fld *Identifier ) *UnrecognizedFieldError {

    return &UnrecognizedFieldError{ Location: p, Field: fld }
}

func ( e *UnrecognizedFieldError ) Error() string {
    msg := e.Message
    if msg == "" { msg = fmt.Sprintf( "unrecognized field: %s", e.Field ) }
    return FormatError( e.Location, msg )
}
