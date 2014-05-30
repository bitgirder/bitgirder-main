package parser

import ( 
    "testing"
    "bitgirder/assert"
)

func parseCoreParseTest( 
    cpt *CoreParseTest, 
    a *assert.PathAsserter ) ( res interface{}, err error ) {

    lx := newTestLexer( cpt.In, false )
    switch cpt.TestType {
    case TestTypeString, TestTypeNumber: 
        res, _, err = lx.ReadToken()
        if err == nil { expectEof( lx, a ) }
    case TestTypeIdentifier: res, err = ParseIdentifier( cpt.In )
    case TestTypeDeclaredTypeName: res, err = ParseDeclaredTypeName( cpt.In )
    case TestTypeNamespace: res, err = ParseNamespace( cpt.In )
    case TestTypeQualifiedTypeName: res, err = ParseQualifiedTypeName( cpt.In )
    case TestTypeTypeReference: res, err = ParseTypeReference( cpt.In )
    default: a.Fatalf( "Unknown: %T", cpt.Expect )
    }
    return
}

func assertCoreParseError( 
    cpt *CoreParseTest, err error, a *assert.PathAsserter ) {

    if cpt.Err == nil { a.Fatal( err ) }
    if pe, ok := err.( *ParseError ); ok {
        ee := cpt.Err.( *ParseErrorExpect )
        a.Equal( ee.Message, pe.Message )
        a.Equal( ee.Col, pe.Loc.Col )
        a.Equal( 1, pe.Loc.Line )
    } else { a.Fatal( err ) }
}

func assertCompletableTypeReference(
    expct, act *CompletableTypeReference, a *assert.PathAsserter ) {

    a.Descend( "ErrLoc" ).Equalf( expct.ErrLoc, act.ErrLoc,
        "%s != %s", expct.ErrLoc, act.ErrLoc )
    a.Descend( "Name" ).Equal( expct.Name, act.Name )
    a.Descend( "Restriction" ).Equal( expct.Restriction, act.Restriction )
    a.Descend( "ptrDepth" ).Equal( expct.ptrDepth, act.ptrDepth )
    a.Descend( "quants" ).Equal( expct.quants, act.quants )
}

func assertCoreParse( cpt *CoreParseTest, a *assert.PathAsserter ) {
    if act, err := parseCoreParseTest( cpt, a ); err == nil {
        if cpt.Err == nil { 
            switch expct := cpt.Expect.( type ) {
            case *CompletableTypeReference: 
                actRef := act.( *CompletableTypeReference )
                assertCompletableTypeReference( expct, actRef, a )
            default: a.Equal( cpt.Expect, act ) 
            }
        } else { a.Fatalf( "Got %s, expected error %#v", act, cpt.Err ) }
    } else { assertCoreParseError( cpt, err, a ) }
}

func TestCoreParseTests( t *testing.T ) {
    la := assert.NewListPathAsserter( t )
    for _, cpt := range CoreParseTests { 
        assertCoreParse( cpt, la ) 
        la = la.Next()
    }
}
