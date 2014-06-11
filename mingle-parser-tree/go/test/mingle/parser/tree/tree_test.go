package tree

import (
//    "log"
    "testing"
    "bytes"
    "sort"
    "bitgirder/assert"
    "bitgirder/objpath"
    mg "mingle"
    "mingle/parser"
)

var mgDeclNm = mg.MustDeclaredTypeName
var mgNs = mg.MustNamespace
var mgId = mg.MustIdentifier

func sxTyp( s string ) *parser.CompletableTypeReference {
    opts := &parser.Options{ Reader: bytes.NewBufferString( s ) }
    b := parser.NewBuilder( parser.New( opts ) )
    typ, _, err := b.ExpectTypeReference( nil )
    if err != nil { panic( err ) }
    return typ
}

func keyedElts( args ...interface{} ) *KeyedElements {
    ke := newKeyedElements()
    for i, e := 0, len( args ); i < e; i += 2 {
        id := mgId( args[ i ].( string ) )
        elt := args[ i + 1 ].( SyntaxElement )
        ke.Add( id, elt )
    }
    return ke
}

type treeCheck struct {
    *assert.PathAsserter
    path objpath.PathNode
    t *testing.T
}

func failTreeCheck( path objpath.PathNode, args []interface{}, t *testing.T ) {
    args2 := make( []interface{}, 1, len( args ) + 1 )
    args2[ 0 ] = objpath.Format( path, objpath.StringDotFormatter ) + ":"
    args2 = append( args2, args... )
    t.Fatal( args2... )
}

func newTreeCheck( path objpath.PathNode, t *testing.T ) *treeCheck {
    f := func( args ...interface{} ) { failTreeCheck( path, args, t ) }
    a := assert.AsAsserter( f )
    return &treeCheck{ assert.NewPathAsserter( a ), path, t }
}

func ( t *treeCheck ) descend( node string ) *treeCheck {
    return newTreeCheck( t.path.Descend( node ), t.t )
}

func ( t *treeCheck ) startList() *treeCheck {
    return newTreeCheck( t.path.StartList(), t.t )
}

func ( t *treeCheck ) next() *treeCheck {
    return newTreeCheck( t.path.( *objpath.ListNode ).Next(), t.t )
}

func ( t *treeCheck ) equalLen( l1, l2 int ) int {
    t.Equalf( l1, l2, "Slice lengths differ: %d != %d", l1, l2 ) 
    return l1
}

func ( t *treeCheck ) equalTypeListEntry( e1, e2 *TypeListEntry ) {
    t.descend( "Name" ).Equal( e1.Name, e2.Name )
    t.descend( "Loc" ).Equal( e1.Loc, e2.Loc )
}

func ( t *treeCheck ) equalTypeListEntries( l1, l2 []*TypeListEntry ) {
    lt := t.startList()
    for i, e := 0, t.equalLen( len( l1 ), len( l2 ) ); i < e; i++ {
        lt.equalTypeListEntry( l1[ i ], l2[ i ] )
        lt = lt.next()
    }
}

func ( t *treeCheck ) equalImport( i1, i2 *Import ) {
    t.descend( "Start" ).Equal( i1.Start, i2.Start )
    t.descend( "Namespace" ).Equal( i1.Namespace, i2.Namespace )
    t.descend( "NamespaceLoc" ).Equal( i1.NamespaceLoc, i2.NamespaceLoc )
    t.descend( "IsGlob" ).Equal( i1.IsGlob, i2.IsGlob )
    t.descend( "Includes" ).equalTypeListEntries( i1.Includes, i2.Includes )
    t.descend( "Excludes" ).equalTypeListEntries( i1.Excludes, i2.Excludes )
}

func ( t *treeCheck ) equalImports( arr1, arr2 []*Import ) {
    l := t.equalLen( len( arr1 ), len( arr2 ) )
    t = t.startList()
    for idx := 0; idx < l; idx++ {
        i1, i2 := arr1[ idx ], arr2[ idx ] 
        t.equalImport( i1, i2 )
        t = t.next()
    }
}

func ( t *treeCheck ) equalNsDecl( ns1, ns2 *NamespaceDecl ) {
    t.descend( "Namespace" ).Equal( ns1.Namespace, ns2.Namespace )
    t.descend( "Start" ).Equal( ns1.Start, ns2.Start )
}

func ( t *treeCheck ) equalTypeDeclInfo( i1, i2 *TypeDeclInfo ) {
    t.descend( "Name" ).Equal( i1.Name, i2.Name )
    t.descend( "NameLoc" ).Equal( i1.NameLoc, i2.NameLoc )
    t.descend( "SuperType" ).equalType( i1.SuperType, i2.SuperType )
    t.descend( "SuperTypeLoc" ).Equal( i1.SuperTypeLoc, i2.SuperTypeLoc )
}

// Checking right now is loose on restrictions, and we really only check that
// the two are of the same nullity. A full exhaustive check might
// convert the non-nil restrictions to some string form and disregard their
// locations, which in the expected versions are tied to an anonymous input
// string
func ( t *treeCheck ) equalType( t1, t2 *parser.CompletableTypeReference ) {
    parser.AssertCompletableTypeReference( t1, t2, t.PathAsserter )
}

func ( t *treeCheck ) equalPrimary( p1, p2 *PrimaryExpression ) {
    t.descend( "PrimLoc" ).Equal( p1.PrimLoc, p2.PrimLoc )
    t.descend( "Prim" ).Equal( p1.Prim, p2.Prim )
}

func ( t *treeCheck ) equalQualified( q1, q2 *QualifiedExpression ) {
    t.descend( "Lhs" ).equalExpression( q1.Lhs, q2.Lhs )
    t.descend( "Id" ).Equal( q1.Id, q2.Id )
    t.descend( "IdLoc" ).Equal( q1.IdLoc, q2.IdLoc )
}

func ( t *treeCheck ) equalExpression( e1, e2 Expression ) { 
    switch v := e1.( type ) {
    case *PrimaryExpression: t.equalPrimary( v, e2.( *PrimaryExpression ) )
    case *QualifiedExpression: 
        t.equalQualified( v, e2.( *QualifiedExpression ) )
    default: t.Equal( e1, e2 ) 
    }
}

func ( t *treeCheck ) equalField( f1, f2 *FieldDecl ) {
    t.descend( "Name" ).Equal( f1.Name, f2.Name )
    t.descend( "NameLoc" ).Equal( f1.NameLoc, f2.NameLoc )
    t.descend( "Type" ).equalType( f1.Type, f2.Type )
    t.descend( "TypeLoc" ).Equal( f1.TypeLoc, f2.TypeLoc )
    t.descend( "Default" ).equalExpression( f1.Default, f2.Default )
}

func ( t *treeCheck ) equalFields( arr1, arr2 []*FieldDecl ) {
    l := t.equalLen( len( arr1 ), len( arr2 ) )
    for t, i := t.startList(), 0; i < l; i++ {
        t.equalField( arr1[ i ], arr2[ i ] )
        t = t.next()
    }
}

func ( t *treeCheck ) equalConstructorDecl( cd1, cd2 *ConstructorDecl ) {
    t.descend( "Start" ).Equal( cd1.Start, cd2.Start )
    t.descend( "ArgType" ).equalType( cd1.ArgType, cd2.ArgType )
    t.descend( "ArgTypeLoc" ).Equal( cd1.ArgTypeLoc, cd2.ArgTypeLoc )
}

func ( t *treeCheck ) equalSecurityDecl( sd1, sd2 *SecurityDecl ) {
    t.descend( "Start" ).Equal( sd1.Start, sd2.Start )
    t.descend( "Name" ).Equal( sd1.Name, sd2.Name )
    t.descend( "NameLoc" ).Equal( sd1.NameLoc, sd2.NameLoc )
}

func ( t *treeCheck ) equalSyntaxElement( se1, se2 SyntaxElement ) {
    switch v := se1.( type ) {
    case *ConstructorDecl: t.equalConstructorDecl( v, se2.( *ConstructorDecl ) )
    case *SecurityDecl: t.equalSecurityDecl( v, se2.( *SecurityDecl ) )
    default: t.Fatalf( "Unhandled elt: %T", v )
    }
}

func ( t *treeCheck ) equalKeyedElements( ke1, ke2 *KeyedElements ) {
    if ke1 == nil || ke2 == nil {
        if ! ( ke1 == nil && ke2 == nil ) {
            t.Fatalf( "ke1 != ke2: %v != %v", ke1, ke2 )
        }
    }
    if l1, l2 := ke1.Len(), ke2.Len(); l1 == l2 {
        ke1.EachPair( func( key *mg.Identifier, elts1 []SyntaxElement ) {
            kt := t.descend( key.ExternalForm() )
            if elts2 := ke2.Get( key ); elts2 == nil {
                kt.Fatal( "ke2 has no value" )
            } else { 
                l := kt.equalLen( len( elts1 ), len( elts2 ) )
                for lt, idx := kt.startList(), 0; idx < l; idx++ {
                    lt.equalSyntaxElement( elts1[ idx ], elts2[ idx ] )
                    lt = lt.next()
                }
            }
        })
    } else { t.Fatalf( "Keyed elt sizes differ, ke1:%d, ke2:%d", l1, l2 ) }
}

func ( t *treeCheck ) equalStructDecl( sd1, sd2 *StructDecl ) {
    t.descend( "Start" ).Equal( sd1.Start, sd2.Start )
    t.descend( "Info" ).equalTypeDeclInfo( sd1.Info, sd2.Info )
    t.descend( "Fields" ).equalFields( sd1.Fields, sd2.Fields )
    t.descend( "KeyedElements" ).
        equalKeyedElements( sd1.KeyedElements, sd2.KeyedElements )
}

func ( t *treeCheck ) equalEnumDecl( ed1, ed2 *EnumDecl ) {
    t.descend( "Start" ).Equal( ed1.Start, ed2.Start )
    t.descend( "Name" ).Equal( ed1.Name, ed2.Name )
    t.descend( "NameLoc" ).Equal( ed1.NameLoc, ed2.NameLoc )
    t.descend( "Values" ).Equal( ed1.Values, ed2.Values )
}

func ( t *treeCheck ) equalAliasDecl( ad1, ad2 *AliasDecl ) {
    t.descend( "Start" ).Equal( ad1.Start, ad2.Start )
    t.descend( "Name" ).Equal( ad1.Name, ad2.Name )
    t.descend( "NameLoc" ).Equal( ad1.NameLoc, ad2.NameLoc )
    t.descend( "Target" ).equalType( ad1.Target, ad2.Target )
    t.descend( "TargetLoc" ).Equal( ad1.TargetLoc, ad2.TargetLoc )
}

func ( t *treeCheck ) equalThrown( arr1, arr2 []*ThrownType ) {
    l := t.equalLen( len( arr1 ), len( arr2 ) )
    for lt, idx := t.startList(), 0; idx < l; idx++ {
        tt1, tt2 := arr1[ idx ], arr2[ idx ]
        lt.descend( "Type" ).equalType( tt1.Type, tt2.Type )
        lt.descend( "TypeLoc" ).Equal( tt1.TypeLoc, tt2.TypeLoc )
        lt = lt.next()
    }
}

func ( t *treeCheck ) equalSig( s1, s2 *CallSignature ) {
    t.descend( "Start" ).Equal( s1.Start, s2.Start )
    t.descend( "Fields" ).equalFields( s1.Fields, s2.Fields )
    t.descend( "Return" ).equalType( s1.Return, s2.Return )
    t.descend( "ReturnLoc" ).Equal( s1.ReturnLoc, s2.ReturnLoc )
    t.descend( "Throws" ).equalThrown( s1.Throws, s2.Throws )
}

func ( t *treeCheck ) equalPrototypeDecl( pd1, pd2 *PrototypeDecl ) {
    t.descend( "Start" ).Equal( pd1.Start, pd2.Start )
    t.descend( "Name" ).Equal( pd1.Name, pd2.Name )
    t.descend( "NameLoc" ).Equal( pd1.NameLoc, pd2.NameLoc )
    t.descend( "Sig" ).equalSig( pd1.Sig, pd2.Sig )
}

func ( t *treeCheck ) equalOperations( arr1, arr2 []*OperationDecl ) {
    l := t.equalLen( len( arr1 ), len( arr2 ) )
    for lt, idx := t.startList(), 0; idx < l; idx++ {
        od1, od2 := arr1[ idx ], arr2[ idx ]
        lt.descend( "Name" ).Equal( od1.Name, od2.Name )
        lt.descend( "NameLoc" ).Equal( od1.NameLoc, od2.NameLoc )
        lt.descend( "Call" ).equalSig( od1.Call, od2.Call )
        lt = lt.next()
    }
}

func ( t *treeCheck ) equalServiceDecl( sd1, sd2 *ServiceDecl ) {
    t.descend( "Start" ).Equal( sd1.Start, sd2.Start )
    t.descend( "Info" ).equalTypeDeclInfo( sd1.Info, sd2.Info )
    t.descend( "Operations" ).equalOperations( sd1.Operations, sd2.Operations )
    t.descend( "KeyedElements" ).
        equalKeyedElements( sd1.KeyedElements, sd2.KeyedElements )
}

func ( t *treeCheck ) equalTypeDecl( td1, td2 TypeDecl ) {
    switch v := td1.( type ) {
    case *StructDecl: t.equalStructDecl( v, td2.( *StructDecl ) )
    case *EnumDecl: t.equalEnumDecl( v, td2.( *EnumDecl ) )
    case *AliasDecl: t.equalAliasDecl( v, td2.( *AliasDecl ) )
    case *PrototypeDecl: t.equalPrototypeDecl( v, td2.( *PrototypeDecl ) )
    case *ServiceDecl: t.equalServiceDecl( v, td2.( *ServiceDecl ) )
    default: t.Fatalf( "Unhandled type decl type: %T", td1 )
    }
}    

func ( t *treeCheck ) equalTypeDecls( arr1, arr2 []TypeDecl ) {
    l := t.equalLen( len( arr1 ), len( arr2 ) )
    for t, idx := t.startList(), 0; idx < l; idx++ {
        td1, td2 := arr1[ idx ], arr2[ idx ]
        t.equalTypeDecl( td1, td2 )
        t = t.next()
    }
}

func ( t *treeCheck ) equalNsUnit( u1, u2 *NsUnit ) {
    t.descend( "SourceName" ).Equal( u1.SourceName, u2.SourceName )
    t.descend( "Imports" ).equalImports( u1.Imports, u2.Imports )
    t.descend( "NsDecl" ).equalNsDecl( u1.NsDecl, u2.NsDecl )
    t.descend( "TypeDecls" ).equalTypeDecls( u1.TypeDecls, u2.TypeDecls )
}

func assertParse( nm string, u *NsUnit, t *testing.T ) {
    if expct := testParseResults[ nm ]; expct == nil {
        t.Fatalf( "No result present for %s", nm )
    } else { newTreeCheck( objpath.RootedAt( nm ), t ).equalNsUnit( expct, u ) }
}

func parseSource( nm, src string ) ( *NsUnit, error ) {
    return ParseSource( nm, bytes.NewBufferString( src ) )
}

func TestParseSource( t *testing.T ) {
    nms := make( []string, 0, len( testSources ) )
    for nm, _ := range testSources { nms = append( nms, nm ) }
    sort.Strings( nms )
    for _, nm := range nms {
        txt := testSources[ nm ]
        if nsUnit, err := parseSource( nm, txt ); err == nil {
            assertParse( nm, nsUnit, t )
        } else { t.Fatal( err ) }
    }
}

func TestParseErrors( t *testing.T ) {
    for _, tt := range []struct { errMsg string; line, col int; src string } {
        { `Expected @ but found: namespace`, 1, 1, `namespace ns1` },
        { `Expected operation or keyed def but found: throws`, 5, 9,
`@version v1
namespace ns1
service S1 {
    op op1(): String
        throws T
}`,
        },
        { "Expected one of [ \";\", \"<;>\" ] but found: -", 2, 22,
`@version v1
import ns1@v1/[ S1 ] - [ S2 ]`,
        },
        { "Source version is 'v2' but namespace declared 'v1'", 1, 30, 
            "@version v1; namespace ns1@v2;" },
    } {
        if i, err := parseSource( "test-source", tt.src ); err == nil {
            t.Fatalf( "%d: Expected error %q in %q", i, tt.errMsg, tt.src )
        } else if pe, ok := err.( *parser.ParseError ); ok {
            assert.Equal( tt.errMsg, pe.Message )
            assert.Equal( tt.line, pe.Loc.Line )
            assert.Equal( tt.col, pe.Loc.Col )
        } else { t.Fatalf( "%d: %s", i, err ) }
    }
}

var testSources = map[ string ]string{

    "testSource1":
`
# The source here is not intended to represent something which would compile,
# but only something which is syntactically correct and exercises all of the
# syntax of the language

@version v1

import ns1/Struct1 
import ns1@v2/Error3
import ns1@v2/* # Stuff in a comment
# left blank to preserve line nums below (previous test text removed)

namespace ns1

struct Struct1 {

    string1 String # this sure is a comment
    string2 &ns1:ns2@v2/String?
    string3 ns1@v1/String default "hello there"
    string4 ns1:ns2/String~"a*" default "aaaaa"
    string5 ns1/String~"^.*(a|b)$"?
    bool2 Boolean default true
# left blank to preserve line nums below (previous test text removed)
# left blank to preserve line nums below (previous test text removed)
    int2 Int64 default 1234 + 567
    int5 Int32~[0,) default 1111
    ints2 Int32+ default [ 1, 
        -2, 3, -4,
    ]
    ints3 Int32+ default [ 1, 2, ] # Trailing comma in list ok
    double1 Float64 default 3.1
    double2 Float64~(-1e-10,3]?
    float1 Float32 default 3.2
    float2 Float32; float3 Float32 default 1.2e1; float4 Float32
# left blank to preserve line nums below (previous test text removed)
# left blank to preserve line nums below (previous test text removed)
# left blank to preserve line nums below (previous test text removed)
# left blank to preserve line nums below (previous test text removed)
    float5 Float32
}

struct StructWithFinalComma{ f1 String }
struct EmptyStruct {}

struct Struct3 < Struct1 {

    string6 String?

    @constructor( Int64 )
    @constructor( ns1/Struct1 )
    @constructor( String~"^a+$" )
}

struct Struct4 < Struct1 {}

struct Error1 < StandardError {}
struct Error2 { failTime Int64 }

struct Error3 < Error1 { 
    @constructor( F1 )
    string2 String*
}

enum Enum1 { red, green, lightGrey }

alias Alias1 String?

prototype Proto1(): String~"abc"
prototype Proto2( f1 String ): String throws Error1
prototype Proto3( f1 Struct1, f2 ns1@v1/String default "hi" ): Struct1?

service Service1 {

    op op1(): String*

    op op2( param1 String,
            param2 ns1@v1/Struct1?,
            param3 ns1:ns2/Int64 default 12,
            param4 Alias1*,
            param5 Alias2 ): ns1/Struct2,
                throws Error1, Error3
    
    op op3(): Int64? throws Error2

    @security Sec1
}

service Service2 {}
service Service3 { @security Sec1 }
struct S { 
    f1 E1 default E1.green 
    f2 String* default []
}
`,
    "testSource2":
`@version v1
import ns1@v1/*
import ns1@v1/[ T1, T2 ]
import ns1@v1/[ T1, T2, ] # trailing comma in list ok
import ns1@v1/*-[T1] # No whitespace
import ns1@v1/* - [ T1, T2, ] # with whitespace and trailing comma
namespace ns2
`,
}

var testParseResults = make( map[ string ]*NsUnit )

type typeRefModFunc func( *parser.CompletableTypeReference ) 

func modTyp( 
    t *parser.CompletableTypeReference,
    f typeRefModFunc ) *parser.CompletableTypeReference {

    f( t )
    return t
}

func rxSetLoc( rx parser.RestrictionSyntax, l *parser.Location ) {
    switch v := rx.( type ) {
    case *parser.RegexRestrictionSyntax: v.Loc = l
    case *parser.StringRestrictionSyntax: v.Loc = l
    case *parser.NumRestrictionSyntax: v.Loc = l
    default: panic( libErrorf( "unhandled restriction: %T", rx ) )
    }
}

func typeWithRegexLoc( 
    t *parser.CompletableTypeReference, 
    l *parser.Location ) *parser.CompletableTypeReference {

    rxSetLoc( t.Restriction, l )
    return t
}

func typeWithRangeLoc(
    t *parser.CompletableTypeReference,
    locLeft, locRight *parser.Location ) *parser.CompletableTypeReference {

    rx := t.Restriction.( *parser.RangeRestrictionSyntax )
    if locLeft != nil { rxSetLoc( rx.Left, locLeft ) }
    if locRight != nil { rxSetLoc( rx.Right, locRight ) }
    return t
}

func initResultTestSource1() {
    lc1 := func( line, col int ) *parser.Location {
        return &parser.Location{ Source: "testSource1", Line: line, Col: col }
    }
    sxTyp1 := func( s string, line, col int ) *parser.CompletableTypeReference {
        res := sxTyp( s )
        res.ErrLoc = lc1( line, col )
        return res
    }
    testParseResults[ "testSource1" ] = &NsUnit{
        SourceName: "testSource1",
        NsDecl: &NamespaceDecl{ 
            Namespace: mgNs( "ns1@v1" ), 
            Start: lc1( 13, 11 ),
        },
        Imports: []*Import{
            {
                Start: lc1( 8, 1 ),
                Namespace: mgNs( "ns1@v1" ),
                NamespaceLoc: lc1( 8, 8 ),
                IsGlob: false,
                Includes: []*TypeListEntry{
                    { Name: mgDeclNm( "Struct1" ), Loc: lc1( 8, 12 ) },
                },
            },
            {
                Start: lc1( 9, 1 ),
                Namespace: mgNs( "ns1@v2" ),
                NamespaceLoc: lc1( 9, 8 ),
                IsGlob: false,
                Includes: []*TypeListEntry{
                    { Name: mgDeclNm( "Error3" ), Loc: lc1( 9, 15 ) },
                },
            },
            {
                Start: lc1( 10, 1 ),
                Namespace: mgNs( "ns1@v2" ),
                NamespaceLoc: lc1( 10, 8 ),
                IsGlob: true,
            },
        },
        TypeDecls: []TypeDecl{
            &StructDecl{
                Start: lc1( 15, 1 ),
                Info: &TypeDeclInfo{ 
                    Name: mgDeclNm( "Struct1" ), 
                    NameLoc: lc1( 15, 8 ),
                },
                Fields: []*FieldDecl{
                    { Name: mgId( "string1" ), 
                      Type: sxTyp1( "String", 17, 13 ),
                      NameLoc: lc1( 17, 5 ),
                      TypeLoc: lc1( 17, 13 ) },
                    { Name: mgId( "string2" ), 
                      Type: sxTyp1( "&ns1:ns2@v2/String?", 18, 13 ),
                      NameLoc: lc1( 18, 5 ),
                      TypeLoc: lc1( 18, 13 ) },
                    { Name: mgId( "string3" ),
                      Type: sxTyp1( "ns1@v1/String", 19, 13 ),
                      NameLoc: lc1( 19, 5 ),
                      TypeLoc: lc1( 19, 13 ),
                      Default: &PrimaryExpression{
                        Prim: parser.StringToken( "hello there" ),
                        PrimLoc: lc1( 19, 35 ),
                      } },
                    { Name: mgId( "string4" ),
                      Type: typeWithRegexLoc(
                        sxTyp1( `ns1:ns2@v1/String~"a*"`, 20, 13 ),
                        lc1( 20, 28 ),
                      ),
                      NameLoc: lc1( 20, 5 ),
                      TypeLoc: lc1( 20, 13 ),
                      Default: &PrimaryExpression{
                        Prim: parser.StringToken( "aaaaa" ),
                        PrimLoc: lc1( 20, 41 ),
                      } },
                    { Name: mgId( "string5" ),
                      Type: typeWithRegexLoc(
                        sxTyp1( `ns1@v1/String~"^.*(a|b)$"?`, 21, 13 ),
                        lc1( 21, 24 ),
                      ),
                      NameLoc: lc1( 21, 5 ),
                      TypeLoc: lc1( 21, 13 ) },
                    { Name: mgId( "bool2" ),
                      Type: sxTyp1( "Boolean", 22, 11 ),
                      NameLoc: lc1( 22, 5 ),
                      TypeLoc: lc1( 22, 11 ),
                      Default: &PrimaryExpression{
                        Prim: parser.KeywordTrue,
                        PrimLoc: lc1( 22, 27 ),
                      } },
                    { Name: mgId( "int2" ),
                      Type: sxTyp1( "Int64", 25, 10 ),
                      NameLoc: lc1( 25, 5 ),
                      TypeLoc: lc1( 25, 10 ),
                      Default: &BinaryExpression{
                        Left: &PrimaryExpression{
                            Prim: &parser.NumericToken{ "1234", "", "", 0 },
                            PrimLoc: lc1( 25, 24 ),
                        },
                        Op: parser.SpecialTokenPlus,
                        OpLoc: lc1( 25, 29 ),
                        Right: &PrimaryExpression{
                            Prim: &parser.NumericToken{ "567", "", "", 0 },
                            PrimLoc: lc1( 25, 31 ),
                        },
                      } },
                    { Name: mgId( "int5" ),
                      Type: typeWithRangeLoc(
                        sxTyp1( "Int32~[0,)", 26, 10 ),
                        lc1( 26, 17 ),
                        nil,
                      ),
                      NameLoc: lc1( 26, 5 ),
                      TypeLoc: lc1( 26, 10 ),
                      Default: &PrimaryExpression{
                        Prim: &parser.NumericToken{ "1111", "", "", 0 },
                        PrimLoc: lc1( 26, 29 ),
                      } },
                    { Name: mgId( "ints2" ),
                      Type: sxTyp1( "Int32+", 27, 11 ),
                      NameLoc: lc1( 27, 5 ),
                      TypeLoc: lc1( 27, 11 ),
                      Default: &ListExpression{
                        Start: lc1( 27, 26 ),
                        Elements: []Expression{
                            &PrimaryExpression{
                                Prim: &parser.NumericToken{ "1", "", "", 0 },
                                PrimLoc: lc1( 27, 28 ),
                            },
                            &UnaryExpression{
                                Op: parser.SpecialTokenMinus,
                                OpLoc: lc1( 28, 9 ),
                                Exp: &PrimaryExpression{
                                    Prim: &parser.NumericToken{ Int: "2" },
                                    PrimLoc: lc1( 28, 10 ),
                                },
                            },
                            &PrimaryExpression{
                                Prim: &parser.NumericToken{ "3", "", "", 0 },
                                PrimLoc: lc1( 28, 13 ),
                            },
                            &UnaryExpression{
                                Op: parser.SpecialTokenMinus,
                                OpLoc: lc1( 28, 16 ),
                                Exp: &PrimaryExpression{
                                    Prim: &parser.NumericToken{ Int: "4" },
                                    PrimLoc: lc1( 28, 17 ),
                                },
                            },
                        },
                      },
                    },
                    { Name: mgId( "ints3" ),
                      Type: sxTyp1( "Int32+", 30, 11 ),
                      NameLoc: lc1( 30, 5 ),
                      TypeLoc: lc1( 30, 11 ),
                      Default: &ListExpression{
                        Start: lc1( 30, 26 ),
                        Elements: []Expression{
                            &PrimaryExpression{
                                Prim: &parser.NumericToken{ "1", "", "", 0 },
                                PrimLoc: lc1( 30, 28 ),
                            },
                            &PrimaryExpression{
                                Prim: &parser.NumericToken{ "2", "", "", 0 },
                                PrimLoc: lc1( 30, 31 ),
                            },
                        },
                      },
                    },
                    { Name: mgId( "double1" ),
                      Type: sxTyp1( "Float64", 31, 13 ),
                      NameLoc: lc1( 31, 5 ),
                      TypeLoc: lc1( 31, 13 ),
                      Default: &PrimaryExpression{
                        Prim: &parser.NumericToken{ "3", "1", "", 0 },
                        PrimLoc: lc1( 31, 29 ),
                      } },
                    { Name: mgId( "double2" ),
                      Type: typeWithRangeLoc(
                        sxTyp1( "Float64~(-1e-10,3]?", 32, 13 ),
                        lc1( 32, 22 ),
                        lc1( 32, 29 ),
                      ),
                      NameLoc: lc1( 32, 5 ),
                      TypeLoc: lc1( 32, 13 ) },
                    { Name: mgId( "float1" ),
                      Type: sxTyp1( "Float32", 33, 12 ),
                      NameLoc: lc1( 33, 5 ),
                      TypeLoc: lc1( 33, 12 ),
                      Default: &PrimaryExpression{
                        Prim: &parser.NumericToken{ "3", "2", "", 0 },
                        PrimLoc: lc1( 33, 28 ),
                      } },
                    { Name: mgId( "float2" ), 
                      Type: sxTyp1( "Float32", 34, 12 ),
                      NameLoc: lc1( 34, 5 ),
                      TypeLoc: lc1( 34, 12 ) },
                    { Name: mgId( "float3" ),
                      Type: sxTyp1( "Float32", 34, 28 ),
                      NameLoc: lc1( 34, 21 ),
                      TypeLoc: lc1( 34, 28 ),
                      Default: &PrimaryExpression{
                        Prim: &parser.NumericToken{ "1", "2", "1", 'e' },
                        PrimLoc: lc1( 34, 44 ),
                      } },
                    { Name: mgId( "float4" ), 
                      NameLoc: lc1( 34, 51 ),
                      Type: sxTyp1( "Float32", 34, 58 ),
                      TypeLoc: lc1( 34, 58 ) },
                    { Name: mgId( "float5" ), 
                      NameLoc: lc1( 39, 5 ),
                      Type: sxTyp1( "Float32", 39, 12 ),
                      TypeLoc: lc1( 39, 12 ) },
                },
                KeyedElements: keyedElts(),
            },
            &StructDecl{
                Start: lc1( 42, 1 ),
                Info: &TypeDeclInfo{ 
                    Name: mgDeclNm( "StructWithFinalComma" ),
                    NameLoc: lc1( 42, 8 ),
                },
                Fields: []*FieldDecl{
                    { Name: mgId( "f1" ), 
                      NameLoc: lc1( 42, 30 ),
                      Type: sxTyp1( "String", 42, 33 ),
                      TypeLoc: lc1( 42, 33 ) },
                },
                KeyedElements: keyedElts(),
            },
            &StructDecl{
                Start: lc1( 43, 1 ),
                Info: &TypeDeclInfo{ 
                    Name: mgDeclNm( "EmptyStruct" ),
                    NameLoc: lc1( 43, 8 ),
                },
                KeyedElements: keyedElts(),
            },
            &StructDecl{
                Start: lc1( 45, 1 ),
                Info: &TypeDeclInfo{
                    Name: mgDeclNm( "Struct3" ),
                    NameLoc: lc1( 45, 8 ),
                    SuperType: sxTyp1( "Struct1", 45, 18 ),
                    SuperTypeLoc: lc1( 45, 18 ),
                },
                Fields: []*FieldDecl{
                    { Name: mgId( "string6" ), 
                      NameLoc: lc1( 47, 5 ),
                      Type: sxTyp1( "String?", 47, 13 ),
                      TypeLoc: lc1( 47, 13 ) },
                },
                KeyedElements: keyedElts(
                    "constructor", &ConstructorDecl{ 
                        Start: lc1( 49, 5 ),
                        ArgType: sxTyp1( "Int64", 49, 19 ),
                        ArgTypeLoc: lc1( 49, 19 ),
                    },
                    "constructor", &ConstructorDecl{ 
                        Start: lc1( 50, 5 ),
                        ArgType: sxTyp1( "ns1@v1/Struct1", 50, 19 ),
                        ArgTypeLoc: lc1( 50, 19 ),
                    },
                    "constructor", &ConstructorDecl{ 
                        Start: lc1( 51, 5 ),
                        ArgType: typeWithRegexLoc(
                            sxTyp1( `String~"^a+$"`, 51, 19 ),
                            lc1( 51, 26 ),
                        ),
                        ArgTypeLoc: lc1( 51, 19 ),
                    },
                ),
            },
            &StructDecl{
                Start: lc1( 54, 1 ),
                Info: &TypeDeclInfo{
                    Name: mgDeclNm( "Struct4" ),
                    NameLoc: lc1( 54, 8 ),
                    SuperType: sxTyp1( "Struct1", 54, 18 ),
                    SuperTypeLoc: lc1( 54, 18 ),
                },
                Fields: []*FieldDecl{},
                KeyedElements: keyedElts(),
            },
            &StructDecl{
                Start: lc1( 56, 1 ),
                Info: &TypeDeclInfo{
                    Name: mgDeclNm( "Error1" ),
                    NameLoc: lc1( 56, 8 ),
                    SuperType: sxTyp1( "StandardError", 56, 17 ),
                    SuperTypeLoc: lc1( 56, 17 ),
                },
                Fields: []*FieldDecl{},
                KeyedElements: keyedElts(),
            },
            &StructDecl{
                Start: lc1( 57, 1 ),
                Info: &TypeDeclInfo{ 
                    Name: mgDeclNm( "Error2" ),
                    NameLoc: lc1( 57, 8 ),
                },
                Fields: []*FieldDecl{
                    { Name: mgId( "failTime" ), 
                      NameLoc: lc1( 57, 17 ),
                      Type: sxTyp1( "Int64", 57, 26 ),
                      TypeLoc: lc1( 57, 26 ) },
                },
                KeyedElements: keyedElts(),
            },
            &StructDecl{
                Start: lc1( 59, 1 ),
                Info: &TypeDeclInfo{
                    Name: mgDeclNm( "Error3" ),
                    NameLoc: lc1( 59, 8 ),
                    SuperType: sxTyp1( "Error1", 59, 17 ),
                    SuperTypeLoc: lc1( 59, 17 ),
                },
                Fields: []*FieldDecl{
                    { Name: mgId( "string2" ), 
                      NameLoc: lc1( 61, 5 ),
                      Type: sxTyp1( "String*", 61, 13 ),
                      TypeLoc: lc1( 61, 13 ) },
                },
                KeyedElements: keyedElts(
                    "constructor", &ConstructorDecl{ 
                        Start: lc1( 60, 5 ),
                        ArgType: sxTyp1( "F1", 60, 19 ),
                        ArgTypeLoc: lc1( 60, 19 ),
                    },
                ),
            },
            &EnumDecl{
                Start: lc1( 64, 1 ),
                Name: mgDeclNm( "Enum1" ),
                NameLoc: lc1( 64, 6 ),
                Values: []*EnumValue{ 
                    { mgId( "red" ), lc1( 64, 14 ) },
                    { mgId( "green" ), lc1( 64, 19 ) },
                    { mgId( "lightGrey" ), lc1( 64, 26 ) },
                },
            },
            &AliasDecl{
                Start: lc1( 66, 1 ),
                Name: mgDeclNm( "Alias1" ),
                NameLoc: lc1( 66, 7 ),
                Target: sxTyp1( "String?", 66, 14 ),
                TargetLoc: lc1( 66, 14 ),
            },
            &PrototypeDecl{
                Start: lc1( 68, 1 ),
                Name: mgDeclNm( "Proto1" ),
                NameLoc: lc1( 68, 11 ),
                Sig: &CallSignature{
                    Start: lc1( 68, 17 ),
                    Fields: []*FieldDecl{},
                    Return: typeWithRegexLoc(
                        sxTyp1( `String~"abc"`, 68, 21 ),
                        lc1( 68, 28 ),
                    ),
                    ReturnLoc: lc1( 68, 21 ),
                    Throws: []*ThrownType{},
                },
            },
            &PrototypeDecl{
                Start: lc1( 69, 1 ),
                Name: mgDeclNm( "Proto2" ),
                NameLoc: lc1( 69, 11 ),
                Sig: &CallSignature{
                    Start: lc1( 69, 17 ),
                    Fields: []*FieldDecl{
                        { Name: mgId( "f1" ),
                          NameLoc: lc1( 69, 19 ),
                          Type: sxTyp1( "String", 69, 22 ),
                          TypeLoc: lc1( 69, 22 ) },
                    },
                    Return: sxTyp1( "String", 69, 32 ),
                    ReturnLoc: lc1( 69, 32 ),
                    Throws: []*ThrownType{
                        { Type: sxTyp1( "Error1", 69, 46 ), 
                          TypeLoc: lc1( 69, 46 ),
                        },
                    },
                },
            },
            &PrototypeDecl{
                Start: lc1( 70, 1 ),
                Name: mgDeclNm( "Proto3" ),
                NameLoc: lc1( 70, 11 ),
                Sig: &CallSignature{
                    Start: lc1( 70, 17 ),
                    Fields: []*FieldDecl{
                        { Name: mgId( "f1" ),
                          NameLoc: lc1( 70, 19 ),
                          Type: sxTyp1( "Struct1", 70, 22 ),
                          TypeLoc: lc1( 70, 22 ) },
                        { Name: mgId( "f2" ),
                          NameLoc: lc1( 70, 31 ),
                          Type: sxTyp1( "ns1@v1/String", 70, 34 ),
                          TypeLoc: lc1( 70, 34 ),
                          Default: &PrimaryExpression{
                            Prim: parser.StringToken( "hi" ),
                            PrimLoc: lc1( 70, 56 ),
                          } },
                    },
                    Return: sxTyp1( "Struct1?", 70, 64 ),
                    ReturnLoc: lc1( 70, 64 ),
                },
            },
            &ServiceDecl{
                Start: lc1( 72, 1 ),
                Info: &TypeDeclInfo{
                    Name: mgDeclNm( "Service1" ),
                    NameLoc: lc1( 72, 9 ),
                },
                Operations: []*OperationDecl{
                    { Name: mgId( "op1" ),
                      NameLoc: lc1( 74, 8 ),
                      Call: &CallSignature{
                        Start: lc1( 74, 11 ),
                        Fields: []*FieldDecl{},
                        Return: sxTyp1( "String*", 74, 15 ),
                        ReturnLoc: lc1( 74, 15 ),
                        Throws: []*ThrownType{},
                      },
                    },
                    { Name: mgId( "op2" ),
                      NameLoc: lc1( 76, 8 ),
                      Call: &CallSignature{
                        Start: lc1( 76, 11 ),
                        Fields: []*FieldDecl{
                            { Name: mgId( "param1" ),
                              NameLoc: lc1( 76, 13 ),
                              Type: sxTyp1( "String", 76, 20 ),
                              TypeLoc: lc1( 76, 20 ) },
                            { Name: mgId( "param2" ),
                              NameLoc: lc1( 77, 13 ),
                              Type: sxTyp1( "ns1@v1/Struct1?", 77, 20 ),
                              TypeLoc: lc1( 77, 20 ) },
                            { Name: mgId( "param3" ),
                              NameLoc: lc1( 78, 13 ),
                              Type: sxTyp1( "ns1:ns2@v1/Int64", 78, 20 ),
                              TypeLoc: lc1( 78, 20 ),
                              Default: &PrimaryExpression{
                                Prim: &parser.NumericToken{ Int: "12" },
                                PrimLoc: lc1( 78, 42 ),
                              } },
                            { Name: mgId( "param4" ),
                              NameLoc: lc1( 79, 13 ),
                              Type: sxTyp1( "Alias1*", 79, 20 ),
                              TypeLoc: lc1( 79, 20 ) },
                            { Name: mgId( "param5" ),
                              NameLoc: lc1( 80, 13 ),
                              Type: sxTyp1( "Alias2", 80, 20 ),
                              TypeLoc: lc1( 80, 20 ) },
                        },
                        Return: sxTyp1( "ns1@v1/Struct2", 80, 30 ),
                        ReturnLoc: lc1( 80, 30 ),
                        Throws: []*ThrownType{
                            { Type: sxTyp1( "Error1", 81, 24 ),
                              TypeLoc: lc1( 81, 24 ) },
                            { Type: sxTyp1( "Error3", 81, 32 ),
                              TypeLoc: lc1( 81, 32 ) },
                        },
                      },
                    },
                    { Name: mgId( "op3" ),
                      NameLoc: lc1( 83, 8 ),
                      Call: &CallSignature{
                        Start: lc1( 83, 11 ),
                        Fields: []*FieldDecl{},
                        Return: sxTyp1( "Int64?", 83, 15 ),
                        ReturnLoc: lc1( 83, 15 ),
                        Throws: []*ThrownType{
                            { Type: sxTyp1( "Error2", 83, 29 ),
                              TypeLoc: lc1( 83, 29 ) },
                        },
                      },
                    },
                },
                KeyedElements: keyedElts(
                    "security", &SecurityDecl{
                        Start: lc1( 85, 5 ),
                        Name: mg.MustDeclaredTypeName( "Sec1" ),
                        NameLoc: lc1( 85, 15 ),
                    },
                ),
            },
            &ServiceDecl{
                Start: lc1( 88, 1 ),
                Info: &TypeDeclInfo{
                    Name: mgDeclNm( "Service2" ),
                    NameLoc: lc1( 88, 9 ),
                },
                KeyedElements: keyedElts(),
            },
            &ServiceDecl{
                Start: lc1( 89, 1 ),
                Info: &TypeDeclInfo{
                    Name: mgDeclNm( "Service3" ),
                    NameLoc: lc1( 89, 9 ),
                },
                KeyedElements: keyedElts(
                    "security", &SecurityDecl{
                        Start: lc1( 89, 20 ),
                        Name: mg.MustDeclaredTypeName( "Sec1" ),
                        NameLoc: lc1( 89, 30 ),
                    },
                ),
            },
            &StructDecl{
                Start: lc1( 90, 1 ),
                Info: &TypeDeclInfo{
                    Name: mgDeclNm( "S" ),
                    NameLoc: lc1( 90, 8 ),
                },
                Fields: []*FieldDecl{
                    { Name: mgId( "f1" ), 
                      Type: sxTyp1( "E1", 91, 8 ),
                      NameLoc: lc1( 91, 5 ),
                      TypeLoc: lc1( 91, 8 ),
                      Default: &QualifiedExpression{
                        Lhs: &PrimaryExpression{
                            Prim: sxTyp1( "E1", 91, 19 ), 
                            PrimLoc: lc1( 91, 19 ),
                        },
                        Id: mgId( "green" ),
                        IdLoc: lc1( 91, 22 ),
                      },
                    },
                    { Name: mgId( "f2" ),
                      Type: sxTyp1( "String*", 92, 8 ),
                      NameLoc: lc1( 92, 5 ),
                      TypeLoc: lc1( 92, 8 ),
                      Default: &ListExpression{
                        Elements: []Expression{},
                        Start: lc1( 92, 24 ),
                      },
                    },
                },
                KeyedElements: keyedElts(),
            },
        },
    }
    lc2 := func( line, col int ) *parser.Location {
        return &parser.Location{ Source: "testSource2", Line: line, Col: col }
    }
    testParseResults[ "testSource2" ] = &NsUnit{
        SourceName: "testSource2",
        NsDecl: &NamespaceDecl{ 
            Namespace: mgNs( "ns2@v1" ), 
            Start: lc2( 7, 11 ),
        },
        Imports: []*Import{
            {
                Start: lc2( 2, 1 ),
                Namespace: mgNs( "ns1@v1" ),
                NamespaceLoc: lc2( 2, 8 ),
                IsGlob: true,
            },
            {
                Start: lc2( 3, 1 ),
                Namespace: mgNs( "ns1@v1" ),
                NamespaceLoc: lc2( 3, 8 ),
                IsGlob: false,
                Includes: []*TypeListEntry{
                    { Name: mgDeclNm( "T1" ), Loc: lc2( 3, 17 ) },
                    { Name: mgDeclNm( "T2" ), Loc: lc2( 3, 21 ) },
                },
            },
            {
                Start: lc2( 4, 1 ),
                Namespace: mgNs( "ns1@v1" ),
                NamespaceLoc: lc2( 4, 8 ),
                IsGlob: false,
                Includes: []*TypeListEntry{
                    { Name: mgDeclNm( "T1" ), Loc: lc2( 4, 17 ) },
                    { Name: mgDeclNm( "T2" ), Loc: lc2( 4, 21 ) },
                },
            },
            {
                Start: lc2( 5, 1 ),
                Namespace: mgNs( "ns1@v1" ),
                NamespaceLoc: lc2( 5, 8 ),
                IsGlob: true,
                Excludes: []*TypeListEntry{
                    { Name: mgDeclNm( "T1" ), Loc: lc2( 5, 18 ) },
                },
            },
            {
                Start: lc2( 6, 1 ),
                Namespace: mgNs( "ns1@v1" ),
                NamespaceLoc: lc2( 6, 8 ),
                IsGlob: true,
                Excludes: []*TypeListEntry{
                    { Name: mgDeclNm( "T1" ), Loc: lc2( 6, 21 ) },
                    { Name: mgDeclNm( "T2" ), Loc: lc2( 6, 25 ) },
                },
            },
        },
    }
}

func init() {
    initResultTestSource1()
}
