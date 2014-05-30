package parser

import ( 
    "testing"
    "bitgirder/assert"
    "errors"
)

var skipErrPlaceholder = errors.New( "placeholder" )

func parseCoreParseTest( 
    cpt *CoreParseTest, t *testing.T ) ( res interface{}, err error ) {

    t.Logf( "input for test type %s: %q", cpt.TestType, cpt.In )
    lx := newTestLexer( cpt.In, false )
    switch cpt.TestType {
    case TestTypeString, TestTypeNumber: 
        res, _, err = lx.ReadToken()
        if err == nil { expectEof( lx, t ) }
    case TestTypeIdentifier: res, err = ParseIdentifier( cpt.In )
    case TestTypeDeclaredTypeName: res, err = ParseDeclaredTypeName( cpt.In )
    case TestTypeNamespace: res, err = ParseNamespace( cpt.In )
    case TestTypeQualifiedTypeName: res, err = ParseQualifiedTypeName( cpt.In )
    default: 
        t.Logf( "skipping test type: %s", cpt.TestType )
        err = skipErrPlaceholder
        return 
//    default: t.Fatalf( "Unknown: %T", cpt.Expect )
    }
    return
}

func assertCoreParseError( cpt *CoreParseTest, err error, t *testing.T ) {
    if err == skipErrPlaceholder { return }
    if cpt.Err == nil { t.Fatal( err ) }
    if pe, ok := err.( *ParseError ); ok {
        ee := cpt.Err.( *ParseErrorExpect )
        assert.Equal( ee.Message, pe.Message )
        assert.Equal( ee.Col, pe.Loc.Col )
        assert.Equal( 1, pe.Loc.Line )
    } else { t.Fatal( err ) }
}

func assertCoreParse( cpt *CoreParseTest, t *testing.T ) {
    if act, err := parseCoreParseTest( cpt, t ); err == nil {
        if cpt.Err == nil { 
            assert.Equal( cpt.Expect, act ) 
        } else { t.Fatalf( "Got %s, expected error %#v", act, cpt.Err ) }
    } else { assertCoreParseError( cpt, err, t ) }
}

func TestCoreParseTests( t *testing.T ) {
    for _, cpt := range CoreParseTests { assertCoreParse( cpt, t ) }
}
