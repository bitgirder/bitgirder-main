package parser

import ( 
    "testing"
    "bitgirder/assert"
    "errors"
)

var skipErrPlaceholder = errors.New( "placeholder" )

func parseCoreParseTest( 
    cpt *CoreParseTest, t *testing.T ) ( tok Token, err error ) {

    a := newLexerAsserter( cpt.In, false, t )
    switch cpt.TestType {
    case TestTypeString: tok, _, err = a.lx.ReadToken()
    case TestTypeNumber: 
        if tok, _, err = a.lx.ReadToken(); err == nil {
            if tok == SpecialTokenMinus && cpt.Expect != nil {
                assert.True( cpt.Expect.( *NumericTokenTest ).Negative )
                tok, _, err = a.lx.ReadToken() // now get the number
            }
        }
    default: 
        t.Logf( "skipping test type: %s", cpt.TestType )
        err = skipErrPlaceholder
        return 
//    default: t.Fatalf( "Unknown: %T", cpt.Expect )
    }
    if err == nil { expectEof( a.lx, t ) }
    return
}

func convCptVal( val interface{} ) interface{} {
    switch v := val.( type ) {
    case *NumericTokenTest: return v.Token
    }
    return val
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
    if tok, err := parseCoreParseTest( cpt, t ); err == nil {
        if cpt.Err == nil { 
            expct := convCptVal( cpt.Expect )
            assert.Equal( expct, tok ) 
        } else { t.Fatalf( "Got %s, expected error %#v", tok, cpt.Err ) }
    } else { assertCoreParseError( cpt, err, t ) }
}

func TestCoreParseTests( t *testing.T ) {
    for _, cpt := range CoreParseTests { assertCoreParse( cpt, t ) }
}
