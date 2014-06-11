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

func newTestLexer( in string, strip bool ) *Lexer {
    return New(
        &Options{
            Reader: bytes.NewBufferString( in ),
            SourceName: ParseSourceInput,
            Strip: strip,
        },
    )
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

func AssertCompletableTypeReference(
    expct, act *CompletableTypeReference, a *assert.PathAsserter ) {

    if expct == nil {
        a.Truef( act == nil, "expected nil, got %s", act )
        return
    }
    a.Descend( "ErrLoc" ).Equalf( expct.ErrLoc, act.ErrLoc,
        "%s != %s", expct.ErrLoc, act.ErrLoc )
    a.Descend( "Name" ).Equal( expct.Name, act.Name )
    assertRestriction( 
        expct.Restriction, act.Restriction, a.Descend( "Restriction" ) )
    a.Descend( "ptrDepth" ).Equal( expct.ptrDepth, act.ptrDepth )
    a.Descend( "quants" ).Equal( expct.quants, act.quants )
}

func MustQname( s string ) *mg.QualifiedTypeName {
    qn, err := ParseQualifiedTypeName( s )
    if err == nil { return qn }
    panic( err )
}
