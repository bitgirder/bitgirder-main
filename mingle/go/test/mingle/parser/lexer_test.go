package parser

import (
    "testing"
    "io"
    "bitgirder/assert"
    "bytes"
    mg "mingle"
//    "log"
)

// A mish-mash of tokens used to assert base coverage of the token set and
// location reporting; more exhaustive and specific tests by token type are
// found in TestTokenVariations() and TestCoreParseTests()
const lexerSrc1 =
`namespace bitgirder:mingle x
# Comment at start of line
#Comment with no ws after # sign (<-- '#' in comment test too)
    # Indented comment before field
	# <-- literal \t precedes comment
camelCappedString: Text;
""
"\n\r\t\f\b\"\\\u01fF"
0 0.0 -1 -1.3e-5
: { } ; ~ ( ) [ ] , ? - -> / . * + < > @ &
:}-->
`

type lexerAsserter struct {
    lx *Lexer
    *testing.T
}

func newLexerAsserter( src string, strip bool, t *testing.T ) *lexerAsserter {
    return &lexerAsserter{ newTestLexer( src, strip ), t }
}

func ( a *lexerAsserter ) expectToken( line, col int, expct interface{} ) {
//    log.Printf( "Expecting (%T) %q at %d:%d", expct, expct, line, col )
    if tok, loc, err := a.lx.ReadToken(); err == nil {
//        log.Printf( "Read (%T) %v at %s", tok, tok, loc )
        assert.Equal( expct, tok )
        assert.Equal( line, loc.Line )
        assert.Equal( col, loc.Col )
    } else { a.Fatal( err ) }
}

func ( a *lexerAsserter ) expectEof() {
    if tok, loc, err := a.lx.ReadToken(); err != io.EOF {
        a.Fatalf( "Expected eof, got tok (%T) %#s at %#v", tok, tok, loc )
    }
}

func TestLexerBasic( t *testing.T ) {
    a := newLexerAsserter( lexerSrc1, false, t )
    a.expectToken( 1, 1, Keyword( "namespace" ) )
    a.expectToken( 1, 10, WhitespaceToken( " " ) )
    a.expectToken( 1, 11, id( "bitgirder" ) )
    a.expectToken( 1, 20, SpecialTokenColon )
    a.expectToken( 1, 21, id( "mingle" ) )
    a.expectToken( 1, 27, WhitespaceToken( " " ) )
    a.expectToken( 1, 28, id( "x" ) )
    a.expectToken( 1, 29, SpecialTokenSynthEnd )
    a.expectToken( 1, 29, WhitespaceToken( "\n" ) )
    a.expectToken( 2, 1, CommentToken( " Comment at start of line\n" ) )
    a.expectToken( 3, 1, CommentToken( 
        "Comment with no ws after # sign (<-- '#' in comment test too)\n" ) )
    a.expectToken( 4, 1, WhitespaceToken( "    " ) )
    a.expectToken( 4, 5, CommentToken( " Indented comment before field\n" ) )
    a.expectToken( 5, 1, WhitespaceToken( "\t" ) )
    a.expectToken( 5, 2, CommentToken( " <-- literal \\t precedes comment\n" ) )
    a.expectToken( 6, 1, id( "camel", "capped", "string" ) )
    a.expectToken( 6, 18, SpecialTokenColon ) 
    a.expectToken( 6, 19, WhitespaceToken( " " ) )
    a.expectToken( 6, 20, makeTypeName( "Text" ) )
    a.expectToken( 6, 24, SpecialTokenSemicolon )
    a.expectToken( 6, 25, WhitespaceToken( "\n" ) )
    a.expectToken( 7, 1, StringToken( "" ) )
    a.expectToken( 7, 3, SpecialTokenSynthEnd )
    a.expectToken( 7, 3, WhitespaceToken( "\n" ) )
    a.expectToken( 8, 1, StringToken( "\n\r\t\f\b\"\\\u01fF" ) )
    a.expectToken( 8, 23, SpecialTokenSynthEnd )
    a.expectToken( 8, 23, WhitespaceToken( "\n" ) )
    a.expectToken( 9, 1, &NumericToken{ "0", "", "", 0 } )
    a.expectToken( 9, 2, WhitespaceToken( " " ) )
    a.expectToken( 9, 3, &NumericToken{ "0", "0", "", 0 } )
    a.expectToken( 9, 6, WhitespaceToken( " " ) )
    a.expectToken( 9, 7, SpecialTokenMinus )
    a.expectToken( 9, 8, &NumericToken{ "1", "", "", 0 } )
    a.expectToken( 9, 9, WhitespaceToken( " " ) )
    a.expectToken( 9, 10, SpecialTokenMinus )
    a.expectToken( 9, 11, &NumericToken{ "1", "3", "-5", 'e' } )
    a.expectToken( 9, 17, SpecialTokenSynthEnd )
    a.expectToken( 9, 17, WhitespaceToken( "\n" ) )
    a.expectToken( 10, 1, SpecialTokenColon )
    a.expectToken( 10, 2, WhitespaceToken( " " ) )
    a.expectToken( 10, 3, SpecialTokenOpenBrace )
    a.expectToken( 10, 4, WhitespaceToken( " " ) )
    a.expectToken( 10, 5, SpecialTokenCloseBrace )
    a.expectToken( 10, 6, WhitespaceToken( " " ) )
    a.expectToken( 10, 7, SpecialTokenSemicolon )
    a.expectToken( 10, 8, WhitespaceToken( " " ) )
    a.expectToken( 10, 9, SpecialTokenTilde )
    a.expectToken( 10, 10, WhitespaceToken( " " ) )
    a.expectToken( 10, 11, SpecialTokenOpenParen )
    a.expectToken( 10, 12, WhitespaceToken( " " ) )
    a.expectToken( 10, 13, SpecialTokenCloseParen )
    a.expectToken( 10, 14, WhitespaceToken( " " ) )
    a.expectToken( 10, 15, SpecialTokenOpenBracket )
    a.expectToken( 10, 16, WhitespaceToken( " " ) )
    a.expectToken( 10, 17, SpecialTokenCloseBracket )
    a.expectToken( 10, 18, WhitespaceToken( " " ) )
    a.expectToken( 10, 19, SpecialTokenComma )
    a.expectToken( 10, 20, WhitespaceToken( " " ) )
    a.expectToken( 10, 21, SpecialTokenQuestionMark )
    a.expectToken( 10, 22, WhitespaceToken( " " ) )
    a.expectToken( 10, 23, SpecialTokenMinus )
    a.expectToken( 10, 24, WhitespaceToken( " " ) )
    a.expectToken( 10, 25, SpecialTokenReturns )
    a.expectToken( 10, 27, WhitespaceToken( " " ) )
    a.expectToken( 10, 28, SpecialTokenForwardSlash )
    a.expectToken( 10, 29, WhitespaceToken( " " ) )
    a.expectToken( 10, 30, SpecialTokenPeriod )
    a.expectToken( 10, 31, WhitespaceToken( " " ) )
    a.expectToken( 10, 32, SpecialTokenAsterisk )
    a.expectToken( 10, 33, WhitespaceToken( " " ) )
    a.expectToken( 10, 34, SpecialTokenPlus )
    a.expectToken( 10, 35, WhitespaceToken( " " ) )
    a.expectToken( 10, 36, SpecialTokenLessThan )
    a.expectToken( 10, 37, WhitespaceToken( " " ) )
    a.expectToken( 10, 38, SpecialTokenGreaterThan )
    a.expectToken( 10, 39, WhitespaceToken( " " ) )
    a.expectToken( 10, 40, SpecialTokenAsperand )
    a.expectToken( 10, 41, WhitespaceToken( " " ) )
    a.expectToken( 10, 42, SpecialTokenAmpersand )
    a.expectToken( 10, 43, WhitespaceToken( "\n" ) )
    a.expectToken( 11, 1, SpecialTokenColon )
    a.expectToken( 11, 2, SpecialTokenCloseBrace )
    a.expectToken( 11, 3, SpecialTokenMinus )
    a.expectToken( 11, 4, SpecialTokenReturns )
    a.expectToken( 11, 6, WhitespaceToken( "\n" ) )
    a.expectEof()
}

func expectEof( lx *Lexer, failer assert.Failer ) {
    // pull off the optional synthetic end and then expect eof (okay to get
    // eof twice)
    f := func() { failer.Fatal( "Expected eof" ) }
    if tok2, _, err2 := lx.ReadToken(); err2 != io.EOF {
        if err2 == nil && tok2 != SpecialTokenSynthEnd { f() }
    }
    if _, _, err2 := lx.ReadToken(); err2 != io.EOF { f() }
}

func parseToken( 
    in string, t *testing.T ) ( tok Token, loc *Location, err error ) {
    lx := newTestLexer( in, false )
    tok, loc, err = lx.ReadToken()
    if err == nil { expectEof( lx, t ) }
    return
}

func TestTokenVariations( t *testing.T ) {
    f := func( s string, expct Token ) {
        if tok, l, err := parseToken( s, t ); err == nil {
            assert.Equal( &Location{ 1, 1, ParseSourceInput }, l )
            assert.Equal( expct, tok )
        } else { t.Fatal( err ) }
    }
    f( "x", id( "x" ) )
    f( "x12", id( "x12" ) )
    f( "someStuff", id( "some", "stuff" ) )
    f( "T", makeTypeName( "T" ) )
    f( "T1", makeTypeName( "T1" ) )
    f( "SomeType", makeTypeName( "SomeType" ) )
}

func TestTokenFailures( t *testing.T ) {
    f := func( in string, col int, expctMsg string ) {
        if tok, _, err := parseToken( in, t ); err == nil {
            t.Fatalf( "Expected error but got token %s", tok )
        } else {
            if pe, ok := err.( *ParseError ); ok {
                assert.Equal( expctMsg, pe.Message )
                assert.Equal( col, pe.Loc.Col )
            } else { t.Fatal( err ) }
        }
    }
    f( "ǿ", 1, `Unexpected char: "ǿ" (U+01FF)` ) // outside string literal
    f( "bad_underscore", 4, `Invalid id rune: "_" (U+005F)` )
    f( "Bad_Type", 4, "Illegal type name rune: \"_\" (U+005F)" )
}

func TestNumericTokenStringer( t *testing.T ) {
    f := func( tok *NumericToken, expct string ) {
        assert.Equal( expct, tok.String() )
    }
    f( &NumericToken{ "1", "", "", 0 }, "1" )
    f( &NumericToken{ "1", "1", "", 0 }, "1.1" )
    f( &NumericToken{ "1", "1", "1", 'e' }, "1.1e1" )
    f( &NumericToken{ "1", "1", "-1", 'E' }, "1.1E-1" )
}

func TestReadTypeNameFailure( t *testing.T ) {
    f := func( text, errMsg string, col int ) {
        lx := newTestLexer( text, false )
        if _, _, err := lx.ReadDeclaredTypeName(); err == nil {
            t.Fatalf( "Expected error for %q", text )
        } else {
            if pe, ok := err.( *ParseError ); ok {
                assert.Equal( errMsg, pe.Message )
                assert.Equal( col, pe.Loc.Col )
            } else { t.Fatal( err ) }
        }
    }
    f( "2Bad", "Illegal type name start: \"2\" (U+0032)", 1 )
    f( "", "Empty type name", 1 )
    f( "A\u01ffbadname", "Illegal type name rune: \"ǿ\" (U+01FF)", 2 )
}

type tokExpct struct {
    line, col int
    source string
    tok Token
}

type statementEndTest struct {
    in string
    toks []tokExpct
}

func assertStatementEnd( test statementEndTest, strip bool, t *testing.T ) {
    a := newLexerAsserter( test.in, strip, t )
    for _, te := range test.toks {
        check := true
        if strip {
            switch te.tok.( type ) {
            case CommentToken, WhitespaceToken: check = false
            }
        }
        if check { a.expectToken( te.line, te.col, te.tok ) }
    }
    a.expectEof()
}

func TestStatementEnds( t *testing.T ) {
    mk := func( line, col int, tok Token ) tokExpct {
        return tokExpct{ line: line, col: col, tok: tok }
    }
    mk1 := func( col int, tok Token ) tokExpct { return mk( 1, col, tok ) }
    mk2 := func( col int, tok Token ) tokExpct { return mk( 2, col, tok ) }
    idA := id( "a" )
    idB := id( "b" )
    nmA := makeTypeName( "A" )
    sp := ws( " " )
    sc := SpecialTokenSemicolon
    end := SpecialTokenSynthEnd
    nl := ws( "\n" )
    stuff := CommentToken( []byte( "stuff" ) )
    tests := []statementEndTest{
        { "a;\n", []tokExpct{ mk1( 1, idA ), mk1( 2, sc ), mk1( 3, nl ) } },
        { "a\n", []tokExpct{ mk1( 1, idA ), mk1( 2, end ), mk1( 2, nl ) } },
        { "a\nb", []tokExpct{ 
            mk1( 1, idA ), mk1( 2, end ), mk1( 2, nl ), mk2( 1, idB ) },
        },
        { "a\n  \t b", []tokExpct{
            mk1( 1, idA ), mk1( 2, end ), mk1( 2, ws( "\n  \t " ) ),
            mk2( 5, idB ) },
        },
        { "\na\nb", []tokExpct{
            mk1( 1, nl ), 
            mk2( 1, idA ), mk2( 2, end ), mk2( 2, nl ),
            mk( 3, 1, idB ) },
        },
        { "a,\n]", []tokExpct{
            mk1( 1, idA ), mk1( 2, SpecialTokenComma ), mk1( 3, nl ),
            mk2( 1, SpecialTokenCloseBracket ), mk2( 2, end ) },
        },
        { "a #stuff", 
          []tokExpct{ 
            mk1( 1, idA ), mk1( 2, sp ), mk1( 2, end ), mk1( 3, stuff ) },
        },
        // From here, just basic coverage of other toks that can end
        { "A\n", []tokExpct{ mk1( 1, nmA ), mk1( 2, end ), mk1( 2, nl ) } },
        { "1.2\n", []tokExpct{ 
            mk1( 1, &NumericToken{ "1", "2", "", 0 } ), 
            mk1( 4, end ), 
            mk1( 4, nl ) },
        },
        { "\"abc\"\n", []tokExpct{ 
            mk1( 1, StringToken( "abc" ) ), mk1( 6, end ), mk1( 6, nl ) },
        },
        { ")\n", []tokExpct{
            mk1( 1, SpecialTokenCloseParen ), mk1( 2, end ), mk1( 2, nl ) },
        },
        { "]\n", []tokExpct{
            mk1( 1, SpecialTokenCloseBracket ), mk1( 2, end ), mk1( 2, nl ) },
        },
        { "}\n", []tokExpct{
            mk1( 1, SpecialTokenCloseBrace ), mk1( 2, end ), mk1( 2, nl ) },
        },
    }
    for _, kwd := range []string{ "true", "false", "return" } {
        endCol := 1 + len( kwd )
        kwdTest := statementEndTest{
            kwd + "\n", 
            []tokExpct{ 
                mk1( 1, kwdMap[ kwd ] ), 
                mk1( endCol, end ), 
                mk1( endCol, nl ),
            },
        }
        tests = append( tests, kwdTest )
    }
    for _, test := range tests {
    for _, strip := range []bool { true, false } {
        assertStatementEnd( test, strip, t )
    }}
}

// Regression for incorrect handling of special tokens at end of stream when
// token is also a prefix of a longer token
func TestTokenReadAtEof( t *testing.T ) {
    a := newLexerAsserter( "-", false, t )
    a.expectToken( 1, 1, SpecialTokenMinus )
    a.expectEof()
}

func TestManualSetSynthEnd( t *testing.T ) {
    a := newLexerAsserter( "+\n-", false, t )
    a.expectToken( 1, 1, SpecialTokenPlus )
    a.lx.SetSynthEnd()
    a.expectToken( 1, 2, SpecialTokenSynthEnd )
    a.expectToken( 1, 2, ws( "\n" ) )
    a.expectToken( 2, 1, SpecialTokenMinus )
    a.expectEof()
}

func TestUnreadToken( t *testing.T ) {
    for _, strip := range []bool{ true, false } {
        a := newLexerAsserter( "+a\nb", strip, t )
        pnc := func() {
            assert.AssertPanic(
                func() { a.lx.UnreadToken() },
                func( err interface{} ) { 
                    if err != lxUnreadNoValErr { t.Fatal( err ) }
                },
            )
        }
        twice := func( col, line int, expct Token ) {
            f:= func() { a.expectToken( col, line, expct ) }
            f(); a.lx.UnreadToken(); pnc(); f()
        }
        pnc() // check fail before anything read
        twice( 1, 1, SpecialTokenPlus )
        twice( 1, 2, id( "a" ) )
        twice( 1, 3, SpecialTokenSynthEnd )
        if ! strip { twice( 1, 3, ws( "\n" ) ) }
        twice( 2, 1, id( "b" ) )
        a.expectEof()
        a.lx.UnreadToken()
        a.expectToken( 2, 1, id( "b" ) )
    }
}

// Just basic coverage of the typed readers
func TestTypedTokenReaders( t *testing.T ) {
    a := newTestLexer( "a B 1", true )
    f := func( mode string, expct Token ) {
        var tok Token
        var err error
        switch mode {
        case "id": tok, _, err = a.ReadIdentifier( false )
        case "type": tok, _, err = a.ReadDeclaredTypeName()
        case "num": tok, _, err = a.ReadNumber()
        default: t.Fatalf( "Bad mode: %s", mode )
        }
        if err != nil { t.Fatal( err ) }
        assert.Equal( expct, tok )
    }
    f( "id", id( "a" ) )
    f( "type", makeTypeName( "B" ) )
    f( "num", &NumericToken{ Int: "1" } )
}

func TestUnreadOfNonSynthAfterExplicitSetSynthCancelsSynth( t *testing.T ) {
    a := newLexerAsserter( ".{\nb", true, t )
    a.expectToken( 1, 1, SpecialTokenPeriod )
    // next two simulate a peek()
    a.expectToken( 1, 2, SpecialTokenOpenBrace )
    a.lx.UnreadToken()
    a.lx.SetSynthEnd() // set synth end, expect brace but not synth end after
    a.expectToken( 1, 2, SpecialTokenOpenBrace )
    a.expectToken( 2, 1, id( "b" ) )
}

func TestExternalIdsNotAsKeywords( t *testing.T ) {
    // first check our baseline: that namespace is normally a keyword
    lx := NewLexer( &LexerOptions{ Reader: bytes.NewBufferString( "namespace" ) } )
    if tok, _, err := lx.ReadToken(); err == nil {
        assert.Equal( KeywordNamespace, tok )
    } else { t.Fatal( err ) }
    opts :=
        &LexerOptions{ 
            IsExternal: true, Reader: bytes.NewBufferString( "namespace" ) }
    lx = NewLexer( opts )
    if tok, _, err := lx.ReadToken(); err == nil {
        assert.Equal( id( "namespace" ), tok.( *mg.Identifier ) )
    } else { t.Fatal( err ) }
}

func TestNumberAccessorsSuccess( t *testing.T ) {
    succ := func( i, f, e string, eChar rune, num interface{} ) {
        tok := &NumericToken{ i, f, e, eChar }
        var val interface{}
        var err error
        switch num.( type ) {
        case int64: val, err = tok.Int64()
        case uint64: val, err = tok.Uint64()
        case float64: val, err = tok.Float64()
        default: t.Fatalf( "Unhandled expect val: %T", num )
        }
        if err == nil { assert.Equal( val, num ) } else { t.Fatal( err ) }
    }
    succ( "0", "", "", 0, int64( 0 ) )
    succ( "0", "", "", 0, uint64( 0 ) )
    succ( "1", "", "", 0, int64( 1 ) )
    succ( "1", "", "", 0, uint64( 1 ) )
    succ( "0", "0", "", 0, float64( 0 ) )
    succ( "0", "0", "0", 'e', float64( 0 ) )
    succ( "0", "0", "10", 'E', float64( 0 ) )
    succ( "0", "0", "-10", 'E', float64( 0 ) )
    succ( "1", "1", "", 0, float64( 1.1 ) )
    succ( "1", "1", "10", 'e', float64( 1.1e10 ) )
    succ( "1", "1", "-10", 'E', float64( 1.1e-10 ) )
    succ( "1", "", "10", 'e', float64( 1e10 ) )
}

func TestNumberAccessorsFail( t *testing.T ) {
    f := func( n *NumericToken, doInt bool, msg string ) {
        var err error
        if doInt { _, err = n.Int64() } else { _, err = n.Float64() }
        if err == nil { t.Fatalf( "Did not get err %q from %v", msg, n ) }
        assert.Equal( msg, err.Error() )
    }
    // just get basic coverage that strconv errors are passed back as expected,
    // and that a non-int token fails in Int64()
    f( &NumericToken{ Int: "a" }, true, 
        "strconv.ParseInt: parsing \"a\": invalid syntax" )
    f( &NumericToken{ Int: "1", Frac: "1" }, true, 
        "strconv.ParseInt: parsing \"1.1\": invalid syntax" )
    f( &NumericToken{ Int: "1", Frac: "x" }, false,
        "strconv.ParseFloat: parsing \"1.x\": invalid syntax" )
}

func TestNumberIsInt( t *testing.T ) {
    f := func( n *NumericToken, expct bool ) {
        assert.Equal( expct, n.IsInt() )
    }
    f( &NumericToken{ Int: "3" }, true )
    f( &NumericToken{ Int: "3", Frac: "1" }, false )
    f( &NumericToken{ Int: "3", Exp: "1" }, false )
}

func TestLexerWithoutRecognizeComments( t *testing.T ) {
    a := newLexerAsserter( "a 1 # blah", false, t )
    a.lx.RejectComments = true
    a.expectToken( 1, 1, id( "a" ) )
    a.expectToken( 1, 2, WhitespaceToken( " " ) )
    a.expectToken( 1, 3, &NumericToken{ Int: "1" } )
    a.expectToken( 1, 4, WhitespaceToken( " " ) )
    tok, _, err := a.lx.ReadToken()
    if err == nil { t.Fatalf( "expected error for '#', got: %s", tok ) }
    expct := &ParseErrorExpect{ 5, "Unexpected comment start" }
    AssertParseError( err, expct, assert.NewPathAsserter( t ) )
}
