package parser

import (
    "testing"
    "bitgirder/assert"
//    "log"
    "bytes"
    "strings"
    "errors"
    mg "mingle"
)

type syntaxBuildTester struct {
    *testing.T
    sb *Builder
}

func newSyntaxBuilder( in string, strip bool ) *Builder {
    opts := &LexerOptions{
        Reader: bytes.NewBufferString( in ),
        SourceName: ParseSourceInput,
        Strip: strip,
    }
    return NewBuilder( NewLexer( opts ) )
}

func newSyntaxBuildTester( 
    in string, strip bool, t *testing.T ) *syntaxBuildTester {
    return &syntaxBuildTester{ T: t, sb: newSyntaxBuilder( in, strip ) }
}

func ( st *syntaxBuildTester ) expectToken( call string, expct Token ) {
    var tn *TokenNode
    var err error
    switch call {
    case "any": tn, err = st.sb.nextTokenNode()
    case "expect-special": 
        tn, err = st.sb.ExpectSpecial( expct.( SpecialToken ) )
    case "poll-special": 
        tn, err = st.sb.PollSpecial( expct.( SpecialToken ) )
    case "numeric": tn, err = st.sb.ExpectNumericToken()
    case "identifier": tn, err = st.sb.ExpectIdentifier()
    case "decl-name": tn, err = st.sb.ExpectDeclaredTypeName()
    default: st.Fatalf( "Unhandled call: %s", call )
    }
    if err == nil { assert.Equal( expct, tn.Token ) } else { st.Fatal( err ) }
}

func ( st *syntaxBuildTester ) expectSynthEnd() {
    if _, err := st.sb.ExpectSpecial( SpecialTokenSynthEnd ); err != nil {
        st.Fatal( err )
    }
}

func ( st *syntaxBuildTester ) expectTypeReference() *CompletableTypeReference {
    ref, err := st.sb.ExpectTypeReference( nil )
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
                idVal := tn.Identifier()
                assert.Equal( id( "a" ), idVal )
            }
        } else { t.Fatal( err ) }
    }
}

func TestSkipWs( t *testing.T ) {
    chk := func( in string, expct interface{} ) {
        a := assert.Asserter{ t }
        sb := newSyntaxBuilder( in, false )
        if err := sb.SkipWs(); err != nil { t.Fatal( err ) }
        tn, err := sb.PeekToken()
        if err != nil { t.Fatal( err ) }
        var act interface{}
        if tn != nil { act = tn.Token }
        a.Equal( expct, act )
    }
    chk( "", nil )
    chk( " \t\r\n", nil )
    chk( "   abc", id( "abc" ) )
    chk( "   # stuff", CommentToken( " stuff" ) )
}

func TestExpectSpecialMulti( t *testing.T ) {
    expct := []SpecialToken{ SpecialTokenQuestionMark, SpecialTokenAsperand }
    pa := assert.NewPathAsserter( t )
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
                AssertParseError( err, &ParseErrorExpect{ 1, msg }, pa )
            }
        }
    }
}

func TestTokenUnexpectedEndOfInput( t *testing.T ) {
    sb := newSyntaxBuilder( "a", false )
    sb.nextTokenNode()
    if _, err := sb.nextTokenNode(); err == nil {
        t.Fatal( "Expected error" )
    } else {
        if pe, ok := err.( *ParseError ); ok {
            assert.Equal( 2, pe.Loc.Col )
            assert.Equal( "Unexpected end of input", pe.Message )
        } else { t.Fatal( err ) }
    }
}

func TestTokenListTrailingTokenError( t *testing.T ) {
    sb := newSyntaxBuilder( "a+", false )
    sb.nextTokenNode()
    if err := sb.CheckTrailingToken(); err == nil {
        t.Fatalf( "Expected error" )
    } else {
        if pe, ok := err.( *ParseError ); ok {
            assert.Equal( 2, pe.Loc.Col )
            assert.Equal( "Unexpected token: +", pe.Message )
        } else { t.Fatal( err ) }
    }
}

func assertScopedVersionIn(
    val interface{}, verExpct *mg.Identifier, t *testing.T ) {
    switch v := val.( type ) {
    case *mg.Namespace: assert.Equal( verExpct, v.Version )
    case *mg.QualifiedTypeName: 
        assertScopedVersionIn( v.Namespace, verExpct, t )
    case *CompletableTypeReference: 
        nm := atomicExpressionIn( v.Expression ).Name
        assertScopedVersionIn( nm, verExpct, t )
    default: t.Fatalf( "No version to check in %v", val )
    }
}

func assertScopedVersion( 
    str string, scope, verExpct *mg.Identifier, call string, t *testing.T ) {
    sb := newSyntaxBuilder( str, false )
    var val interface{}
    var err error
    switch call {
    case "ns": val, _, err = sb.ExpectNamespace( scope )
    case "qn": val, _, err = sb.ExpectQualifiedTypeName( scope )
    case "type": val, err = sb.ExpectTypeReference( scope )
    default: t.Fatalf( "Unhandled call: %s", call )
    }
    if err != nil { 
        t.Fatalf( "Call %q failed for %q, scope: %v: %s", 
            call, str, scope, err )
    }
    assertScopedVersionIn( val, verExpct, t )
}

func TestScopedVersion( t *testing.T ) {
    v1 := id( "v1" )
    v2 := id( "v2" )
    for _, nsStr := range []string { "ns1:ns2@v1", "ns1:ns2@v2", "ns1:ns2" } {
    for _, scope := range []*mg.Identifier { v1, nil } {
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
        opts := &LexerOptions{
            Reader: bytes.NewBufferString( in ), 
            Strip: strip,
            SourceName: ParseSourceInput,
        }
        lx := NewLexer( opts )
        st := &syntaxBuildTester{ T: t, sb: NewBuilder( lx ) }
        skip := func() {
            if ! strip {
                if err := st.sb.SkipWsOrComments(); err != nil {
                    t.Fatal( err )
                }
            }
        }
        skip()
        st.expectToken( "expect-special", SpecialTokenPlus )
        skip()
        st.expectToken( "poll-special", SpecialTokenPlus )
        skip()
        st.expectToken( "numeric", &NumericToken{ Int: "3" } )
        st.expectSynthEnd()
        skip()
        st.expectToken( "identifier", id( "id" ) )
        skip()
        st.expectSynthEnd()
        if ! strip { skip() } // since synth end will have stopped prev skip
        st.expectToken( "decl-name", makeTypeName( "Name1" ) )
    }
}

func TestTypeReferenceSetsSynth( t *testing.T ) {
    a := assert.NewPathAsserter( t )
    nmA := makeTypeName( "A" )
    lc := func( col int ) *Location {
        return &Location{ Line: 1, Source: ParseSourceInput, Col: col }
    }
    regex := &RegexRestrictionSyntax{ Pat: "a", Loc: lc( 3 ) }
    rng := &RangeRestrictionSyntax{
        Loc: lc( 3 ),
        LeftClosed: true, 
        Left: &NumRestrictionSyntax{
            Num: &NumericToken{ Int: "0" }, Loc: lc( 4 ) },
        Right: &NumRestrictionSyntax{
            Num: &NumericToken{ Int: "1" }, Loc: lc( 6 ) },
        RightClosed: true,
    }
    for _, s := range []struct { in string; ref *CompletableTypeReference }{
        { 
            "A", 
            &CompletableTypeReference{ 
                Expression: &AtomicTypeExpression{
                    Name: nmA,
                    NameLoc: lc( 1 ),
                },
            },
        },
        { 
            "A+", 
            &CompletableTypeReference{ 
                Expression: &ListTypeExpression{
                    Loc: lc( 2 ),
                    Expression: &AtomicTypeExpression{
                        Name: nmA,
                        NameLoc: lc( 1 ),
                    },
                    AllowsEmpty: false,
                },
            },
        },
        { 
            `A~"a"+`, 
            &CompletableTypeReference{ 
                Expression: &ListTypeExpression{
                    Loc: lc( 6 ),
                    Expression: &AtomicTypeExpression{
                        Name: nmA,
                        NameLoc: lc( 1 ),
                        Restriction: regex,
                    },
                    AllowsEmpty: false,
                },
            },
        },
        { 
            `A~"a"`, 
            &CompletableTypeReference{ 
                Expression: &AtomicTypeExpression{
                    Name: nmA,
                    NameLoc: lc( 1 ),
                    Restriction: regex, 
                },
            },
        },
        { 
            `A~[0,1]+`, 
            &CompletableTypeReference{ 
                Expression: &ListTypeExpression{
                    Loc: lc( 8 ),
                    Expression: &AtomicTypeExpression{
                        Name: nmA,
                        NameLoc: lc( 1 ),
                        Restriction: rng,
                    },
                    AllowsEmpty: false,
                },
            },
        },
        { 
            `A~[0,1]`, 
            &CompletableTypeReference{ 
                Expression: &AtomicTypeExpression{
                    NameLoc: lc( 1 ),
                    Name: nmA, 
                    Restriction: rng, 
                },
            },
        },
        {
            `&A*`,
            &CompletableTypeReference{
                Expression: &ListTypeExpression{
                    Loc: lc( 3 ),
                    AllowsEmpty: true,
                    Expression: &PointerTypeExpression{
                        Loc: lc( 1 ),
                        Expression: &AtomicTypeExpression{
                            Name: nmA,
                            NameLoc: lc( 2 ),
                        },
                    },
                },
            },
        },
        {
            `(A)`,
            &CompletableTypeReference{
                Expression: &AtomicTypeExpression{
                    Name: nmA,
                    NameLoc: lc( 2 ),
                },
            },
        },
        {
            `&(A)`,
            &CompletableTypeReference{
                Expression: &PointerTypeExpression{
                    Loc: lc( 1 ),
                    Expression: &AtomicTypeExpression{
                        Name: nmA,
                        NameLoc: lc( 3 ),
                    },
                },
            },
        },
    } {
        st := newSyntaxBuildTester( s.in + "\n", false, t )
        if ref, err := st.sb.ExpectTypeReference( nil ); err == nil {
            AssertCompletableTypeReference( s.ref, ref, a )
        } else { t.Fatal( err ) }
        st.expectSynthEnd()
        st.expectToken( "any", WhitespaceToken( []byte( "\n" ) ) )
    }
}

type typeCompleterImpl struct {
    in string
    err error
    notOk bool
    expct mg.TypeReference
}

func ( tc *typeCompleterImpl ) qnameForName( 
    nm mg.TypeName ) *mg.QualifiedTypeName {
    
    switch {
    case nm.Equals( mg.NewDeclaredTypeNameUnsafe( "String" ) ):
        return mg.QnameString
    case nm.Equals( mg.NewDeclaredTypeNameUnsafe( "Int32" ) ):
        return mg.QnameInt32
    }
    return nm.( *mg.QualifiedTypeName )
}

func ( tc *typeCompleterImpl ) addRestriction(
    at *mg.AtomicTypeReference, rx RestrictionSyntax ) error {

    if rx == nil { return nil }
    mkInt := func( rx RestrictionSyntax ) mg.Int32 {
        n := rx.( *NumRestrictionSyntax ).Num 
        res, err := mg.ParseNumber( n.String(), mg.QnameInt32 )
        if err == nil { return res.( mg.Int32 ) }
        panic( err )
    }
    switch v := rx.( type ) {
    case *RegexRestrictionSyntax:
        var err error
        at.Restriction, err = mg.NewRegexRestriction( v.Pat )
        if err != nil { return err }
    case *RangeRestrictionSyntax:
        rr := &mg.RangeRestriction{
            MinClosed: v.LeftClosed, 
            MaxClosed: v.RightClosed,
        }
        if v.Left != nil { rr.Min = mkInt( v.Left ) }
        if v.Right != nil { rr.Max = mkInt( v.Right ) }
        at.Restriction = rr
    default: panic( libErrorf( "unhandled restriction: %T", rx ) )
    }
    return nil
}

func ( tc *typeCompleterImpl ) CompleteBaseType( 
    nm mg.TypeName,
    rx RestrictionSyntax,
    l *Location ) ( mg.TypeReference, bool, error ) {

    if tc.notOk { return nil, false, tc.err }
    if tc.err != nil { return nil, true, tc.err }
    at := &mg.AtomicTypeReference{ Name: tc.qnameForName( nm ) }
    if err := tc.addRestriction( at, rx ); err != nil { return nil, false, err }
    return at, true, nil
}

func TestCompleteType( t *testing.T ) {
    la := assert.NewListPathAsserter( t )
    for _, tt := range []*typeCompleterImpl{
        { in: "mingle:core@v1/String", expct: mg.TypeString },
        {
            in: `mingle:core@v1/String~"a"`,
            expct: &mg.AtomicTypeReference{
                Name: mg.QnameString,
                Restriction: mg.MustRegexRestriction( "a" ),
            },
        },
        { 
            in: `mingle:core@v1/Int32~[0,2)`,
            expct: &mg.AtomicTypeReference{
                Name: mg.QnameInt32,
                Restriction: &mg.RangeRestriction{
                    MinClosed: true,
                    Min: mg.Int32( int32( 0 ) ),
                    Max: mg.Int32( int32( 2 ) ),
                    MaxClosed: false,
                },
            },
        },
        { in: "Int32", expct: mg.TypeInt32 },
        {
            in: "&&String",
            expct: mg.NewPointerTypeReference(
                mg.NewPointerTypeReference( mg.TypeString ) ),
        },
        {
            in: "&String?",
            expct: mg.MustNullableTypeReference(
                mg.NewPointerTypeReference( mg.TypeString ) ),
        },
        {
            in: "String*+*",
            expct: &mg.ListTypeReference{
                ElementType: &mg.ListTypeReference{
                    ElementType: &mg.ListTypeReference{
                        ElementType: mg.TypeString,
                        AllowsEmpty: true,
                    },
                    AllowsEmpty: false,
                },
                AllowsEmpty: true,
            },
        },
        { in: "Int32?", err: mg.NewNullableTypeError( mg.TypeInt32 ) },
        { in: "Stuff", err: errors.New( "test-error" ) },
        { in: "Stuff", notOk: true, err: errors.New( "test-error" ) },
        { in: "Stuff", notOk: true },
    } {
        ct, err := ParseTypeReference( tt.in )
        if err != nil { la.Fatal( err ) }
        typ, err := ct.CompleteType( tt )
        if err == nil {
            if tt.notOk {
                la.Truef( typ == nil, "got a type: %s", typ )
            } else { la.Equal( tt.expct, typ ) }
        } else { la.EqualErrors( tt.err, err ) }
        la = la.Next()
    }
}
