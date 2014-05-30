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
