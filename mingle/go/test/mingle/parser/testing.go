package parser

import (
    "testing"
    "bitgirder/assert"
    mg "mingle"
//    "log"
)

func id( strs ...string ) *mg.Identifier {
    return mg.NewIdentifierUnsafe( strs )
}

func ws( str string ) WhitespaceToken { 
    return WhitespaceToken( []byte( str ) )
}

func makeTypeName( str string ) DeclaredTypeName {
    return DeclaredTypeName( []byte( str ) )
}

type ParseErrorExpect struct {
    Col int
    Message string
}

func AssertParseError(
    err error, errExpct *ParseErrorExpect, t *testing.T ) {
    if pErr, ok := err.( *ParseError); ok {
        if pErr.Message != errExpct.Message {
            t.Fatalf( "Got error message %#v but expected %#v",
                pErr.Message, errExpct.Message )
        }
        if pErr.Loc.Col != errExpct.Col {
            t.Fatalf( "Got col %d but expected %d", pErr.Loc.Col, errExpct.Col )
        }
        if pErr.Loc.Line != 1 { 
            t.Fatalf( "Unexpected err line %d", pErr.Loc.Line ) 
        }
        if pErr.Loc.Source != ParseSourceInput {
            t.Fatalf( "Unexpected error source: %#v", pErr.Loc.Source )
        }
    } else { t.Fatalf( "%s (%T)", err, err ) }
}

func AssertParsePanic( errExpct *ParseErrorExpect, t *testing.T, f func() ) {
    errHndlr := func( err interface{} ) {
        if parseErr, ok := err.( *ParseError ); ok {
            AssertParseError( parseErr, errExpct, t )
        } else { t.Fatal( err ) }
    }
    assert.AssertPanic( f, errHndlr )
}
