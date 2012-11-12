package syntax

import (
    "testing"
    "bitgirder/assert"
    "mingle/parser/lexer"
    "mingle/parser/loc"
    pt "mingle/parser/testing"
//    "log"
    "bytes"
    "strings"
)

func id( part string ) Identifier {
    return Identifier( [][]byte{ []byte( part ) } ) 
}

func declName( nm string ) DeclaredTypeName {
    return DeclaredTypeName( []byte( nm ) )
}

type syntaxBuildTester struct {
    *testing.T
    sb *Builder
}

func newSyntaxBuilder( in string, strip bool ) *Builder {
    opts := &lexer.Options{
        Reader: bytes.NewBufferString( in ),
        SourceName: loc.ParseSourceInput,
        Strip: strip,
    }
    return NewBuilder( lexer.New( opts ) )
}

func newSyntaxBuildTester( 
    in string, strip bool, t *testing.T ) *syntaxBuildTester {
    return &syntaxBuildTester{ T: t, sb: newSyntaxBuilder( in, strip ) }
}

func ( st *syntaxBuildTester ) expectToken( call string, expct Token ) {
    var tn *TokenNode
    var err error
    switch call {
    case "any": tn, err = st.sb.nextToken()
    case "expect-special": 
        tn, err = st.sb.ExpectSpecial( expct.( lexer.SpecialToken ) )
    case "poll-special": 
        tn, err = st.sb.PollSpecial( expct.( lexer.SpecialToken ) )
    case "numeric": tn, err = st.sb.ExpectNumericToken()
    case "identifier": tn, err = st.sb.ExpectIdentifier()
    case "decl-name": tn, err = st.sb.ExpectDeclaredTypeName()
    default: st.Fatalf( "Unhandled call: %s", call )
    }
    if err == nil { assert.Equal( expct, tn.Token ) } else { st.Fatal( err ) }
}

func ( st *syntaxBuildTester ) expectSynthEnd() {
    if _, err := st.sb.ExpectSpecial( lexer.SpecialTokenSynthEnd ); err != nil {
        st.Fatal( err )
    }
}

func ( st *syntaxBuildTester ) expectTypeReference() *CompletableTypeReference {
    ref, _, err := st.sb.ExpectTypeReference( nil )
    if err != nil { st.Fatal( err ) }
    return ref
}

func TestSkipWsOrComments( t *testing.T ) {
    for _, s := range []string {
        "a",
        "  a",
        "// comment\n\ta",
        "// comment\na",
    } {
        sb := newSyntaxBuilder( s, false )
        if err := sb.SkipWsOrComments(); err == nil {
            if tn, err := sb.ExpectIdentifier(); err == nil {
                id := tn.Identifier()
                assert.Equal( Identifier( [][]byte{ []byte( "a" ) } ), id )
            }
        } else { t.Fatal( err ) }
    }
}

func TestExpectSpecialMulti( t *testing.T ) {
    expct := []lexer.SpecialToken{ 
        lexer.SpecialTokenQuestionMark, lexer.SpecialTokenAsperand }
    for i := 0; i < len( expct ); i++ {
        specs := expct[ 0 : i + 1 ]
        st := newSyntaxBuildTester( "?", false, t )
        if _, err := st.sb.ExpectSpecial( specs... ); err != nil {
            t.Fatal( err )
        }
        for _, s := range []string { "+", "+?", "a?" } {
            st = newSyntaxBuildTester( s, false, t )
            if _, err := st.sb.ExpectSpecial( specs... ); err == nil {
                t.Fatalf( "Expected err, got '?'" )
            } else {
                var clause string
                if len( specs ) == 1 {
                    clause = string( specs[ 0 ] )
                } else { clause = `one of [ "?", "@" ]` }
                msg := `Expected ` + clause + ` but found: ` + string( s[ 0 ] )
                pt.AssertParseError( err, &pt.ParseErrorExpect{ 1, msg }, t )
            }
        }
    }
}

func TestTokenUnexpectedEndOfInput( t *testing.T ) {
    sb := newSyntaxBuilder( "a", false )
    sb.nextToken()
    if _, err := sb.nextToken(); err == nil {
        t.Fatal( "Expected error" )
    } else {
        if pe, ok := err.( *loc.ParseError ); ok {
            assert.Equal( 2, pe.Loc.Col )
            assert.Equal( "Unexpected end of input", pe.Message )
        } else { t.Fatal( err ) }
    }
}

func TestTokenListTrailingTokenError( t *testing.T ) {
    sb := newSyntaxBuilder( "a+", false )
    sb.nextToken()
    if err := sb.CheckTrailingToken(); err == nil {
        t.Fatalf( "Expected error" )
    } else {
        if pe, ok := err.( *loc.ParseError ); ok {
            assert.Equal( 2, pe.Loc.Col )
            assert.Equal( "Unexpected token: +", pe.Message )
        } else { t.Fatal( err ) }
    }
}

func assertScopedVersionIn(
    val interface{}, verExpct Identifier, t *testing.T ) {
    switch v := val.( type ) {
    case *Namespace: assert.Equal( verExpct, v.Version )
    case *QualifiedTypeName: assertScopedVersionIn( v.Namespace, verExpct, t )
    case *CompletableTypeReference: assertScopedVersionIn( v.Name, verExpct, t )
    default: t.Fatalf( "No version to check in %v", val )
    }
}

func assertScopedVersion( 
    str string, scope, verExpct Identifier, call string, t *testing.T ) {
    sb := newSyntaxBuilder( str, false )
    var val interface{}
    var err error
    switch call {
    case "ns": val, _, err = sb.ExpectNamespace( scope )
    case "qn": val, _, err = sb.ExpectQualifiedTypeName( scope )
    case "type": val, _, err = sb.ExpectTypeReference( scope )
    default: t.Fatalf( "Unhandled call: %s", call )
    }
    if err != nil { 
        t.Fatalf( "Call %q failed for %q, scope: %v: %s", 
            call, str, scope, err )
    }
    assertScopedVersionIn( val, verExpct, t )
}

func TestScopedVersion( t *testing.T ) {
    v1 := Identifier( [][]byte{ []byte( "v1" ) } )
    v2 := Identifier( [][]byte{ []byte( "v2" ) } )
    for _, nsStr := range []string { "ns1:ns2@v1", "ns1:ns2@v2", "ns1:ns2" } {
    for _, scope := range []Identifier { v1, Identifier( nil ) } {
        if ! ( scope == nil && nsStr == "ns1:ns2" ) {
            verExpct := v1
            if strings.Contains( nsStr, "@v2" ) { verExpct = v2 }
            assertScopedVersion( nsStr, scope, verExpct, "ns", t )
            assertScopedVersion( nsStr + "/T1", scope, verExpct, "qn", t )
            assertScopedVersion( nsStr + "/T1*", scope, verExpct, "type", t )
        }
    }}
}

// Regression for bug that caused panic when reading ws at end of a document
func TestSkipWsOrCommentsAtEnd( t *testing.T ) {
    sb := newSyntaxBuilder( " ", false )
    if err := sb.SkipWsOrComments(); err != nil { t.Fatal( err ) }
}

func TestTokenExpectsWithStripVariants( t *testing.T ) {
    in := " + + # stuff \n3\n \t id # more stuff\n #leading stuff\n\tName1"
    for _, strip := range []bool{ true, false } {
        opts := &lexer.Options{
            Reader: bytes.NewBufferString( in ), 
            Strip: strip,
            SourceName: loc.ParseSourceInput,
        }
        lx := lexer.New( opts )
        st := &syntaxBuildTester{ T: t, sb: NewBuilder( lx ) }
        skip := func() {
            if ! strip {
                if err := st.sb.SkipWsOrComments(); err != nil {
                    t.Fatal( err )
                }
            }
        }
        skip()
        st.expectToken( "expect-special", lexer.SpecialTokenPlus )
        skip()
        st.expectToken( "poll-special", lexer.SpecialTokenPlus )
        skip()
        st.expectToken( "numeric", &lexer.NumericToken{ Int: "3" } )
        st.expectSynthEnd()
        skip()
        st.expectToken( "identifier", Identifier( [][]byte{ []byte( "id" ) } ) )
        skip()
        st.expectSynthEnd()
        if ! strip { skip() } // since synth end will have stopped prev skip
        st.expectToken( "decl-name", DeclaredTypeName( []byte( "Name1" ) ) )
    }
}

func TestTypeReferenceSetsSynth( t *testing.T ) {
    nmA := DeclaredTypeName( []byte( "A" ) )
    quants := quantList( []lexer.SpecialToken{ lexer.SpecialTokenPlus } )
    emptyQuants := quantList( []lexer.SpecialToken{} )
    lc := func( col int ) *loc.Location {
        return &loc.Location{ Line: 1, Source: loc.ParseSourceInput, Col: col }
    }
    regex := &RegexRestrictionSyntax{ Pat: "a", Loc: lc( 3 ) }
    rng := &RangeRestrictionSyntax{
        LeftClosed: true, 
        Left: &NumRestrictionSyntax{
            Num: &lexer.NumericToken{ Int: "0" }, Loc: lc( 4 ) },
        Right: &NumRestrictionSyntax{
            Num: &lexer.NumericToken{ Int: "1" }, Loc: lc( 6 ) },
        RightClosed: true,
    }
    for typ, expct := range map[ string ]*CompletableTypeReference {
        "A": &CompletableTypeReference{ nmA, nil, emptyQuants },
        "A+": &CompletableTypeReference{ nmA, nil, quants },
        `A~"a"+`: &CompletableTypeReference{ nmA, regex, quants },
        `A~"a"`: &CompletableTypeReference{ nmA, regex, emptyQuants },
        `A~[0,1]+`: &CompletableTypeReference{ nmA, rng, quants },
        `A~[0,1]`: &CompletableTypeReference{ nmA, rng, emptyQuants },
    } {
        st := newSyntaxBuildTester( typ + "\n", false, t )
        if ref, _, err := st.sb.ExpectTypeReference( nil ); err == nil {
            assert.Equal( expct, ref )
        } else { t.Fatal( err ) }
        st.expectSynthEnd()
        st.expectToken( "any", lexer.WhitespaceToken( []byte( "\n" ) ) )
    }
}
