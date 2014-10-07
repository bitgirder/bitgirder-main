package parser

import (
    "bitgirder/assert"
    "bytes"
    mg "mingle"
//    "log"
)

func id( strs ...string ) *mg.Identifier {
    return mg.NewIdentifierUnsafe( strs )
}

func ws( str string ) WhitespaceToken { return WhitespaceToken( str ) }

var makeTypeName = mg.NewDeclaredTypeNameUnsafe

type ParseErrorExpect struct {
    Col int
    Message string
}

func AssertParseError(
    err error, errExpct *ParseErrorExpect, a *assert.PathAsserter ) {

    pErr, ok := err.( *ParseError)
    if ! ok { a.Fatal( err ) }
    a.Descend( "Message" ).Equal( errExpct.Message, pErr.Message )
    aLoc := a.Descend( "Loc" )
    aLoc.Descend( "Col" ).Equal( errExpct.Col, pErr.Loc.Col )
    aLoc.Descend( "Line" ).Equal( 1, pErr.Loc.Line )
    aLoc.Descend( "Source" ).Equal( ParseSourceInput, pErr.Loc.Source )
}

// mutates opts
func newTestLexerOptions( opts *LexerOptions, in string, strip bool ) *Lexer {
    opts.Reader = bytes.NewBufferString( in )
    opts.SourceName = ParseSourceInput
    opts.Strip = strip
    return NewLexer( opts )
}

func newTestLexer( in string, strip bool ) *Lexer {
    return newTestLexerOptions( &LexerOptions{}, in, strip )
}

func assertRegexRestriction( 
    expct *RegexRestrictionSyntax, 
    rs RestrictionSyntax,
    a *assert.PathAsserter ) {

    act, ok := rs.( *RegexRestrictionSyntax )
    a.Truef( ok, "not a regex restriction: %T", rs )
    a.Descend( "Loc" ).Equal( expct.Loc, act.Loc )
}

func assertRangeRestriction(
    expct *RangeRestrictionSyntax,
    rs RestrictionSyntax,
    a *assert.PathAsserter ) {

    act, ok := rs.( *RangeRestrictionSyntax )
    a.Truef( ok, "not a range restriction: %T", rs )
    a.Descend( "Loc" ).Equal( expct.Loc, act.Loc )
    a.Descend( "LeftClosed" ).Equal( expct.LeftClosed, act.LeftClosed )
    assertRestriction( expct.Left, act.Left, a.Descend( "Left" ) )
    assertRestriction( expct.Right, act.Right, a.Descend( "Right" ) )
    a.Descend( "RightClosed" ).Equal( expct.RightClosed, act.RightClosed )
}

func assertNumRestriction( 
    expct *NumRestrictionSyntax,
    rs RestrictionSyntax,
    a *assert.PathAsserter ) {

    act, ok := rs.( *NumRestrictionSyntax )
    a.Truef( ok, "not a num restriction: %T", rs )
    a.Descend( "IsNeg" ).Equal( expct.IsNeg, act.IsNeg )
    a.Descend( "Num" ).Equal( expct.Num, act.Num )
    a.Descend( "Loc" ).Equal( expct.Loc, act.Loc )
}

func assertRestriction( expct, act RestrictionSyntax, a *assert.PathAsserter ) {
    if expct == nil {
        a.Truef( act == nil, "got non-nil restriction: %s", act )
        return
    }
    switch v := expct.( type ) {
    case *RegexRestrictionSyntax: assertRegexRestriction( v, act, a )
    case *RangeRestrictionSyntax: assertRangeRestriction( v, act, a )
    case *NumRestrictionSyntax: assertNumRestriction( v, act, a )
    default: a.Fatalf( "unhandled restriction: %T", expct )
    }
}

func assertAtomicTypeExpression(
    expct *AtomicTypeExpression, v interface{}, a *assert.PathAsserter ) {

    a = a.Descend( "Restriction" )
    act, ok := v.( *AtomicTypeExpression )
    a.Truef( ok, "not atomic: %T", v )
    a.Descend( "Name" ).Equal( expct.Name, act.Name )
    a.Descend( "NameLoc" ).Equal( expct.NameLoc, act.NameLoc )
    assertRestriction( expct.Restriction, act.Restriction, a )
}

func assertListTypeExpression(
    expct *ListTypeExpression, v interface{}, a *assert.PathAsserter ) {

    a = a.Descend( "listTyp" )
    act, ok := v.( *ListTypeExpression )
    a.Truef( ok, "not list type: %T", v )
    a.Descend( "Loc" ).Equal( expct.Loc, act.Loc )
    a.Descend( "AllowsEmpty" ).Equal( expct.AllowsEmpty, act.AllowsEmpty )
    assertEqualExpression( expct.Expression, act.Expression, a )
}

func assertNullableTypeExpression(
    expct *NullableTypeExpression, v interface{}, a *assert.PathAsserter ) {

    a = a.Descend( "nullTyp" )
    act, ok := v.( *NullableTypeExpression )
    a.Truef( ok, "not nullable type: %T", v )
    a.Descend( "Loc" ).Equal( expct.Loc, act.Loc )
    assertEqualExpression( expct.Expression, act.Expression, a )
}

func assertPointerTypeExpression(
    expct *PointerTypeExpression, v interface{}, a *assert.PathAsserter ) {

    a = a.Descend( "ptrTyp" )
    act, ok := v.( *PointerTypeExpression )
    a.Truef( ok, "not a pointer type: %T", v )
    a.Descend( "Loc" ).Equal( expct.Loc, act.Loc )
    assertEqualExpression( expct.Expression, act.Expression, a )
}

func assertEqualExpression( expct, act interface{}, a *assert.PathAsserter ) {
    switch v := expct.( type ) {
    case *AtomicTypeExpression: assertAtomicTypeExpression( v, act, a )
    case *ListTypeExpression: assertListTypeExpression( v, act, a )
    case *NullableTypeExpression: assertNullableTypeExpression( v, act, a )
    case *PointerTypeExpression: assertPointerTypeExpression( v, act, a )
    default: a.Fatalf( "unhandled exp: %T", expct )
    }
}

func AssertCompletableTypeReference(
    expct, act *CompletableTypeReference, a *assert.PathAsserter ) {

    if expct == nil {
        a.Truef( act == nil, "expected nil, got %s", act )
        return
    }
    assertEqualExpression( expct.Expression, act.Expression, a )
}

func MustIdentifier( s string ) *mg.Identifier {
    id, err := ParseIdentifier( s )
    if err == nil { return id }
    panic( err )
}

func MustNamespace( s string ) *mg.Namespace {
    ns, err := ParseNamespace( s )
    if err == nil { return ns }
    panic( err )
}

func MustDeclaredTypeName( s string ) *mg.DeclaredTypeName {
    nm, err := ParseDeclaredTypeName( s )
    if err == nil { return nm }
    panic( err )
}

func MustQualifiedTypeName( s string ) *mg.QualifiedTypeName {
    qn, err := ParseQualifiedTypeName( s )
    if err == nil { return qn }
    panic( err )
}

type unsafeTypeCompleter struct {}

func ( tc *unsafeTypeCompleter ) resolveName( 
    nm mg.TypeName ) *mg.QualifiedTypeName {

    if qn, ok := nm.( *mg.QualifiedTypeName ); ok { return qn }
    return nm.( *mg.DeclaredTypeName ).ResolveIn( mg.CoreNsV1 )
}

func ( tc *unsafeTypeCompleter ) getStringRestriction(
    rx *RegexRestrictionSyntax ) mg.ValueRestriction {

    return mg.MustRegexRestriction( rx.Pat )
}

func ( tc *unsafeTypeCompleter ) setRangeValue(
    valPtr *mg.Value,
    qn *mg.QualifiedTypeName,
    rx RestrictionSyntax ) {

    sx, _ := rx.( *StringRestrictionSyntax )
    nx, _ := rx.( *NumRestrictionSyntax )
    switch {
    case mg.IsNumericTypeName( qn ):
        if num, err := mg.ParseNumber( nx.LiteralString(), qn ); err == nil {
            *valPtr = num
        } else { panic( err ) }
    case qn.Equals( mg.QnameTimestamp ):
        if tm, err := ParseTimestamp( sx.Str ); err == nil {
            *valPtr = tm
        } else { panic( err ) }
    case qn.Equals( mg.QnameString ): *valPtr = mg.String( sx.Str )
    default: panic( libErrorf( "unhandled range type: %s", qn ) )
    }
}

func ( tc *unsafeTypeCompleter ) getRangeRestriction(
    qn *mg.QualifiedTypeName, rx *RangeRestrictionSyntax ) mg.ValueRestriction {

    var min, max mg.Value
    if l := rx.Left; l != nil { tc.setRangeValue( &( min ), qn, l ) }
    if r := rx.Right; r != nil { tc.setRangeValue( &( max ), qn, r ) }
    return mg.MustRangeRestriction( 
        qn, rx.LeftClosed, min, max, rx.RightClosed )
}

func ( tc *unsafeTypeCompleter ) getRestriction(
    qn *mg.QualifiedTypeName, rx RestrictionSyntax ) mg.ValueRestriction {

    if qn.Equals( mg.QnameString ) {
        if regx, ok := rx.( *RegexRestrictionSyntax ); ok {
            return tc.getStringRestriction( regx )
        }
    }
    return tc.getRangeRestriction( qn, rx.( *RangeRestrictionSyntax ) )
}

func ( tc *unsafeTypeCompleter ) CompleteBaseType(
    nm mg.TypeName, 
    rx RestrictionSyntax, 
    errLoc *Location ) ( mg.TypeReference, bool, error ) {

    qn := tc.resolveName( nm )
    var vr mg.ValueRestriction
    if rx != nil { vr = tc.getRestriction( qn, rx ) }
    return mg.NewAtomicTypeReference( qn, vr ), true, nil
//    at := mg.NewAtomicTypeReference( tc.resolveName( nm ), nil )
//    if rx != nil { tc.setRestriction( at, rx ) }
//    return at, true, nil
}

func MustTypeReference( s string ) mg.TypeReference {
    ct, err := ParseTypeReference( s )
    if err != nil { panic( err ) }
    res, err := ct.CompleteType( &unsafeTypeCompleter{} )
    if err != nil { panic( err ) }
    return res
}

func AsTypeReference( val interface{} ) mg.TypeReference {
    switch v := val.( type ) {
    case string: return MustTypeReference( v )
    case mg.TypeReference: return v
    case *mg.QualifiedTypeName: return v.AsAtomicType()
    }
    panic( libErrorf( "unhandled type reference value: %T", val ) )
}

func MustTimestamp( s string ) mg.Timestamp {
    tm, err := ParseTimestamp( s )
    if err == nil { return tm }
    panic( err )
}

func mustAsQname( val interface{} ) *mg.QualifiedTypeName {
    if qn, ok := val.( *mg.QualifiedTypeName ); ok { return qn }
    return MustQualifiedTypeName( val.( string ) )
}

func mustAsId( val interface{} ) *mg.Identifier {
    if id, ok := val.( *mg.Identifier ); ok { return id }
    return MustIdentifier( val.( string ) )
}

func mustMapPairs( pairs []interface{} ) []interface{} {
    res := make( []interface{}, len( pairs ) )
    for i, e := 0, len( res ); i < e; i += 2 {
        res[ i ] = mustAsId( pairs[ i ] )
        res[ i + 1 ] = pairs[ i + 1 ]
    }
    return res
}

func MustSymbolMap( pairs ...interface{} ) *mg.SymbolMap {
    return mg.MustSymbolMap( mustMapPairs( pairs )... )
}

func MustStruct( typ interface{}, pairs ...interface{} ) *mg.Struct {
    return mg.MustStruct( mustAsQname( typ ), mustMapPairs( pairs )... )
}

func MustEnum( typ interface{}, val interface{} ) *mg.Enum {
    return &mg.Enum{ Type: mustAsQname( typ ), Value: mustAsId( val ) }
}
