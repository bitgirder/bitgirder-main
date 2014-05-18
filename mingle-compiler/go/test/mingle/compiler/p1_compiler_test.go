package compiler

import (
    "testing"
    "fmt"
//    "log"
    "bytes"
    mg "mingle"
    "bitgirder/assert"
    "mingle/types"
    "mingle/parser/tree"
)

var p1Sources []string
var p1Defs []*types.DefinitionMap

func p1SourceNameFor( indx int ) string {
    return fmt.Sprintf( "test-source%d", indx )
}

func parseP1Sources() ( []*tree.NsUnit, error ) {
    res := make( []*tree.NsUnit, len( p1Sources ) )
    for i, src := range p1Sources {
        nm := p1SourceNameFor( i )
        buf := bytes.NewBufferString( src )
        if nsUnit, err := tree.ParseSource( nm, buf ); err == nil {
            res[ i ] = nsUnit
        } else { return nil, err }
    }
    return res, nil
}

type p1CompilerRun struct {
    *testing.T
    testUnits []*tree.NsUnit
    sourceCount int
    singleSrc bool
}

func ( r *p1CompilerRun ) buildCompilation( 
    extTypes *types.DefinitionMap, i int ) ( *Compilation, int ) {
    c := NewCompilation()
    c.SetExternalTypes( extTypes )
    if r.singleSrc {
        for ; i < r.sourceCount; i++ { c.AddSource( r.testUnits[ i ] ) }
    } else {
        c.AddSource( r.testUnits[ i ] )
        i++
    }
    return c, i
}

func ( r *p1CompilerRun ) getExpectedDefs( srcIdx int ) *mg.QnameMap {
    res := mg.NewQnameMap()
    i := srcIdx - 1
    if r.singleSrc { i = 0 }
    for ; i < srcIdx; i++ { 
        p1Defs[ i ].EachDefinition( func( d types.Definition ) {
            qn := d.GetName()
            if res.HasKey( qn ) {
                r.Fatalf( "Multiple expected defs: %s", qn )
            } else { res.Put( qn, d ) }
        })
    }
    return res
}

func ( r *p1CompilerRun ) assertExpectedDefsEmpty( m *mg.QnameMap ) {
    if m.Len() != 0 {
        m.EachPair(
            func( qn *mg.QualifiedTypeName, _ interface{} ) {
                r.Errorf( "No result built for %s", qn )
            },
        )
        r.FailNow()
    }
}

func ( r *p1CompilerRun ) assertBuiltTypes( 
    built *types.DefinitionMap, srcIdx int ) {
    m := r.getExpectedDefs( srcIdx )
//    da := newDefAsserter( r.T )
    a := assert.NewPathAsserter( r.T )
    built.EachDefinition(
        func( defAct types.Definition ) {
            qn := defAct.GetName()
            if expct := m.Get( qn ); expct != nil {
                defExpct := expct.( types.Definition )
//                da.descend( qn.ExternalForm() ).assertDef( defExpct, defAct )
                da := types.NewDefAsserter( a.Descend( qn.ExternalForm() ) )
                da.AssertDef( defExpct, defAct )
                m.Delete( qn )
            } else {
                r.Fatalf( "No expected result for %s", qn )
            }
        },
    )
    r.assertExpectedDefsEmpty( m )
}

func ( r *p1CompilerRun ) call() {
    for extTypes, i := types.CoreTypesV1(), 0; i < r.sourceCount; {
        var comp *Compilation
        comp, i = r.buildCompilation( extTypes, i )
        if cr, err := comp.Execute(); err == nil {
            if len( cr.Errors ) == 0 {
                built := roundtripCompilation( cr.BuiltTypes, r.T )
                r.assertBuiltTypes( built, i )
                extTypes.MustAddFrom( cr.BuiltTypes )
            } else {
                for _, err := range cr.Errors { r.Error( err ) }
                r.FailNow()
            }
        } else { r.Fatal( err ) }
    }
}

// Loops through test sources in batches of increasing size, compiling each set
// in two ways: all as one input but also incrementally, with the compiled
// output from each run being the input for compiling the next file. 
//
// We loop through test sources 1..N, testing the program made up of source [1],
// [1,2], [1,2,3]... to help simplify debugging. In theory there is no error
// that will be uncovered in compiling sources [1..k] that wouldn't also be in
// [1..N] for k<N, but when working through issues in source k it can be helpful
// not to have the extra compiler noise from sources >k.
func TestCompilerRunP1( t *testing.T ) {
    if srcLen, defLen := len( p1Sources ), len( p1Defs ); srcLen != defLen {
        t.Fatalf( "#sources (%d) != #defs (%d)", srcLen, defLen )
    }
    units, err := parseP1Sources()
    if err != nil { t.Fatal( err ) }
    for i, e := 0, len( p1Sources ); i < e; i++ {
        for j := 0; j < 2; j++ {
            ( &p1CompilerRun{ 
                sourceCount: i + 1, 
                singleSrc: j == 0, 
                testUnits: units,
                T: t,
            } ).call()
        }
    }
}

func init() {
    p1Sources = []string{
`@version v1

namespace ns1

struct Struct1 {
    string1 String # required field with no default
    string2 String? # nullable String
    string3 String default "hello there"
    string4 String~"a*" default "aaaaa"
    string5 String~"^.*(a|b)$"?
    bool1 &Boolean?
    bool2 Boolean default true
    buf1 Buffer?
    timestamp1 &Timestamp?
    timestamp2 Timestamp default "2007-08-24T14:15:43.123450000-07:00"
    int1 Int64
    int2 Int64 default 1234
    int3 &Int64?
    int4 Int32 default 12
    int5 Int32~[0,) default 1111
    int6 Int64~(,)
    int7 Uint32
    int8 Uint64~[0,100)
    ints1 Int64*
    ints2 Int32+ default [ 1, -2, 3, -4 ]
    float1 Float64 default 3.1
    float2 &Float64~(-1e-10,3]?
    float3 Float32 default 3.2
    floats1 Float32*
    val1 Value
    val2 Value default 12
    list1 String* 
    list2 String**
    list3 String~"abc$"*
}

struct Struct2 {
    inst1 ns1/Struct1 # Unnecessary but legal fqname reference
    inst2 Struct1*
}

enum Enum1 { red, green, lightGrey }

struct Struct3 < Struct1 {
    string6 String?
    inst1 Struct2
    enum1 &Enum1?
    enum2 Enum1 default Enum1.green

    @constructor( Int64 )
    @constructor( ns1/Struct1 )
    @constructor( String~"^a+$" )
}

struct Exception1 < StandardError {}

struct Exception2 { failTime Int64 }

struct Exception3 < Exception1 { string2 String* }

alias Alias1 String?
alias Alias2 Struct1
alias Alias3 Alias1*
alias Alias4 String~"^a+$"
alias Alias5 Int64~[0,)

struct Struct5 {
    f1 Alias1
    f2 Alias1 default "hello"
    f3 Alias1*
    f4 Alias1+ default [ "a", "b" ]
    f5 Alias1*+
    f6 Alias2
    f7 Alias2*
    f8 Alias3 default [ "hello" ]
    f9 Alias3+
    f10 Alias4 default "aaa"
    f11 Alias5 default 12
}

service Service1 {

    op op1(): String*

    op op2( param1 String,
         param2 &Struct1?,
         param3 Int64 default 12,
         param4 Alias1*,
         param5 Alias2 ): ns1/Struct2,
            throws Exception1, Exception3
    
    op op3(): &Int64? throws Exception2

    op op4(): Null
}

prototype Proto1(): String
prototype Proto2(): String throws Exception1
prototype Proto3( f1 Struct1, f2 String default "hi" ): &Struct1?

prototype Sec1( authentication Struct1 ): Null

service Service2 {

    op op1(): Int64

    # Auth that throws no exceptions
    @security Sec1

    op op2(): Boolean
}

prototype Sec2( authentication Struct1 ): Int64~[9,10],
    throws Exception1, 
           Exception2

# Test of @security with throws attr
service Service3 { @security Sec2 }

struct FieldConstantTester {
    f1 Boolean default true
    f2 Boolean default false
    f3 Int32 default 1
    f4 Int32 default -1
    f5 Int64 default 1
    f6 Int64 default -1
    f7 Float32 default 1.0
    f8 Float32 default -1.0
    f9 Float64 default 1
    f10 Float64 default -1
    f11 Int32~[0,10) default 8
    f12 String default "a"
    f13 String~"a" default "a"
    f14 Enum1 default Enum1.green
    f15 Timestamp default "2007-08-24T14:15:43.123450000-07:00"
    f16 String+ default [ "a", "b", "c" ]
    f17 Int32* default [ 1, 2, 3 ]
    f18 Float64* default []
    f19 String*+ default [ [], [ "a", "b" ], [ "c", "d", "e" ] ]
    f20 Uint32 default 1
    f21 Uint32 default 4294967295
    f22 Uint64 default 0
    f23 Uint64 default 18446744073709551615
}
`,

`@version v1

namespace ns1

struct Struct4 {
    inst1 Struct1
    inst2 Struct2
    str1 Alias1 # Implicitly brought in from compiler-src1.mg
    str2 Alias3
}

struct Exception4 < Exception3 { int1 Int64 default 33 }
`,

`@version v1

import ns1@v1/Struct4 # redundant but legal explicit version
import ns1/Exception3

namespace ns2@v1 # also redundant but legal

# alias in this namespace that points to type in another namespace
alias Alias1 ns1/Struct3

# alias in this namespace that points to alias in another namespace
alias Alias2 ns1/Alias3

struct Struct1 {
    inst1 &ns1@v1/Struct1?
    inst2 Struct2
    inst3 Struct4
    inst4 Alias1
    inst5 ns1/Alias2
    inst6 Alias2
}

struct Struct2 {
    inst1 Struct1
    inst2 ns1/Struct1
}

struct Exception1 < ns1/Exception1 {}

struct Exception2 < Exception3 { str1 String* }

service Service1 {

    op op1(): Struct2 throws Exception1, ns1/Exception4

    op op2( param1 String*+,
            param2 Struct4*,
            param3 ns1/Struct1,
            param4 Struct2 ): String
}
`,

`@version v1

import ns1/* - [ Exception1, Exception2, Struct1, Service1 ]
import ns2/Exception1 

namespace ns1:globTestNs

struct Struct1 < ns1/Struct1 {
    inst1 Struct2
    inst2 Struct4
    inst3 Struct1 
}

struct Exception2 {}

service Service1 {

    op op1( param1 Struct1*,
            param2 ns1/Struct1,
            param3 Struct3 ): String,
                throws Exception1, Exception2, Exception3
}
`,

`@version v2

namespace ns1

struct Struct1 { f1 String default "hello" }

struct Struct2 < ns1@v1/Struct1 { f1 &Struct1? }

alias Struct1V1 ns1@v1/Struct1

struct Struct3 < Struct1 { 
    f2 Struct2 
    f3 ns1@v1/Alias1
    f4 Struct1V1*
}

service Service1 {
    op op1( f1 Struct1, f2 Int64~[0,12), f3 ns1@v1/Struct1*+ ): Null
}
`,
    }
    mkQn := mg.MustQualifiedTypeName
    mkTyp := mg.MustTypeReference
    p1Defs = []*types.DefinitionMap{
        types.MakeDefMap(
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v1/Alias1" ),
                AliasedType: mkTyp( "mingle:core@v1/String?" ),
            },
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v1/Alias2" ),
                AliasedType: mkTyp( "ns1@v1/Struct1" ),
            },
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v1/Alias3" ),
                AliasedType: mkTyp( "mingle:core@v1/String?*" ),
            },
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v1/Alias4" ),
                AliasedType: mkTyp( `mingle:core@v1/String~"^a+$"` ),
            },
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v1/Alias5" ),
                AliasedType: mkTyp( `mingle:core@v1/Int64~[0,)` ),
            },
            types.MakeStructDef(
                "ns1@v1/Struct1",
                "",
                []*types.FieldDefinition{
                    types.MakeFieldDef( 
                        "string1", "mingle:core@v1/String", nil ),
                    types.MakeFieldDef( 
                        "string2", "mingle:core@v1/String?", nil ),
                    types.MakeFieldDef( 
                        "string3", "mingle:core@v1/String", "hello there" ),
                    types.MakeFieldDef(
                        "string4", `mingle:core@v1/String~"a*"`, "aaaaa" ),
                    types.MakeFieldDef(
                        "string5", `mingle:core@v1/String~"^.*(a|b)$"?`, nil ),
                    types.MakeFieldDef( 
                        "bool1", "&mingle:core@v1/Boolean?", nil ),
                    types.MakeFieldDef( 
                        "bool2", "mingle:core@v1/Boolean", true ),
                    types.MakeFieldDef( "buf1", "mingle:core@v1/Buffer?", nil ),
                    types.MakeFieldDef( 
                        "timestamp1", "&mingle:core@v1/Timestamp?", nil ),
                    types.MakeFieldDef(
                        "timestamp2", 
                        "mingle:core@v1/Timestamp",
                        mg.MustTimestamp( 
                            "2007-08-24T14:15:43.123450000-07:00" ),
                    ),
                    types.MakeFieldDef( "int1", "mingle:core@v1/Int64", nil ),
                    types.MakeFieldDef( 
                        "int2", "mingle:core@v1/Int64", int64( 1234 ) ),
                    types.MakeFieldDef( "int3", "&mingle:core@v1/Int64?", nil ),
                    types.MakeFieldDef( 
                        "int4", "mingle:core@v1/Int32", int32( 12 ) ),
                    types.MakeFieldDef( 
                        "int5", "mingle:core@v1/Int32~[0,)", int32( 1111 ) ),
                    types.MakeFieldDef( 
                        "int6", "mingle:core@v1/Int64~(,)", nil ),
                    types.MakeFieldDef( "int7", "mingle:core@v1/Uint32", nil ),
                    types.MakeFieldDef( 
                        "int8", "mingle:core@v1/Uint64~[0,100)", nil ),
                    types.MakeFieldDef( "ints1", "mingle:core@v1/Int64*", nil ),
                    types.MakeFieldDef(
                        "ints2",
                        "mingle:core@v1/Int32+",
                        []interface{}{ 
                            int32( 1 ), int32( -2 ), int32( 3 ), int32( -4 ) },
                    ),
                    types.MakeFieldDef(
                        "float1", "mingle:core@v1/Float64", float64( 3.1 ) ),
                    types.MakeFieldDef(
                        "float2", "&mingle:core@v1/Float64~(-1e-10,3]?", nil ),
                    types.MakeFieldDef(
                        "float3", "mingle:core@v1/Float32", float32( 3.2 ) ),
                    types.MakeFieldDef( 
                        "floats1", "mingle:core@v1/Float32*", nil ),
                    types.MakeFieldDef( "val1", "mingle:core@v1/Value", nil ),
                    types.MakeFieldDef( 
                        "val2", "mingle:core@v1/Value", int64( 12 ) ),
                    types.MakeFieldDef( 
                        "list1", "mingle:core@v1/String*", nil ),
                    types.MakeFieldDef( 
                        "list2", "mingle:core@v1/String**", nil ),
                    types.MakeFieldDef( 
                        "list3", `mingle:core@v1/String~"abc$"*`, nil ),
                },
            ),
            types.MakeStructDef(
                "ns1@v1/Struct2",
                "",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "inst1", "ns1@v1/Struct1", nil ),
                    types.MakeFieldDef( "inst2", "ns1@v1/Struct1*", nil ),
                },
            ),
            types.MakeStructDef2(
                "ns1@v1/Struct3",
                "ns1@v1/Struct1",
                []*types.FieldDefinition{
                    types.MakeFieldDef( 
                        "string6", "mingle:core@v1/String?", nil ),
                    types.MakeFieldDef( "inst1", "ns1@v1/Struct2", nil ),
                    types.MakeFieldDef( "enum1", "&ns1@v1/Enum1?", nil ),
                    types.MakeFieldDef( 
                        "enum2",
                        "ns1@v1/Enum1",
                        mg.MustEnum( "ns1@v1/Enum1", "green" ),
                    ),
                },
                []*types.ConstructorDefinition{
                    { mg.MustTypeReference( "mingle:core@v1/Int64" ) },
                    { mg.MustTypeReference( "ns1@v1/Struct1" ) },
                    { mg.MustTypeReference( `mingle:core@v1/String~"^a+$"` ) },
                },
            ),
            types.MakeStructDef(
                "ns1@v1/Struct5",
                "",
                []*types.FieldDefinition{
                    types.MakeFieldDef(
                        "f1", "mingle:core@v1/String?", nil ),
                    types.MakeFieldDef(
                        "f2", "mingle:core@v1/String?", "hello" ),
                    types.MakeFieldDef(
                        "f3", "mingle:core@v1/String?*", nil ),
                    types.MakeFieldDef(
                        "f4", 
                        "mingle:core@v1/String?+", 
                        []interface{}{ "a", "b" },
                    ),
                    types.MakeFieldDef(
                        "f5", "mingle:core@v1/String?*+", nil ),
                    types.MakeFieldDef(
                        "f6", "ns1@v1/Struct1", nil ),
                    types.MakeFieldDef(
                        "f7", "ns1@v1/Struct1*", nil ),
                    types.MakeFieldDef(
                        "f8", 
                        "mingle:core@v1/String?*", 
                        []interface{}{ "hello" },
                    ),
                    types.MakeFieldDef(
                        "f9", "mingle:core@v1/String?*+", nil ),
                    types.MakeFieldDef(
                        "f10", `mingle:core@v1/String~"^a+$"`, "aaa" ),
                    types.MakeFieldDef(
                        "f11", "mingle:core@v1/Int64~[0,)", int64( 12 ) ),
                },
            ),
            types.MakeStructDef(
                "ns1@v1/Exception1",
                "mingle:core@v1/StandardError",
                []*types.FieldDefinition{},
            ),
            types.MakeStructDef(
                "ns1@v1/Exception2",
                "",
                []*types.FieldDefinition{
                    types.MakeFieldDef( 
                        "failTime", "mingle:core@v1/Int64", nil ),
                },
            ),
            types.MakeStructDef(
                "ns1@v1/Exception3",
                "ns1@v1/Exception1",
                []*types.FieldDefinition{
                    types.MakeFieldDef( 
                        "string2", "mingle:core@v1/String*", nil ),
                },
            ),
            types.MakeEnumDef( "ns1@v1/Enum1", "red", "green", "lightGrey" ),
            &types.PrototypeDefinition{
                Name: mkQn( "ns1@v1/Proto1" ),
                Signature: types.MakeCallSig(
                    []*types.FieldDefinition{},
                    "mingle:core@v1/String",
                    []string{},
                ),
            },
            &types.PrototypeDefinition{
                Name: mkQn( "ns1@v1/Proto2" ),
                Signature: types.MakeCallSig(
                    []*types.FieldDefinition{},
                    "mingle:core@v1/String",
                    []string{ "ns1@v1/Exception1" },
                ),
            },
            &types.PrototypeDefinition{
                Name: mkQn( "ns1@v1/Proto3" ),
                Signature: types.MakeCallSig(
                    []*types.FieldDefinition{
                        types.MakeFieldDef( "f1", "ns1@v1/Struct1", nil ),
                        types.MakeFieldDef( 
                            "f2", "mingle:core@v1/String", "hi" ),
                    },
                    "&ns1@v1/Struct1?",
                    []string{},
                ),
            },
            &types.PrototypeDefinition{
                Name: mkQn( "ns1@v1/Sec1" ),
                Signature: types.MakeCallSig(
                    []*types.FieldDefinition{
                        types.MakeFieldDef( 
                            "authentication", "ns1@v1/Struct1", nil ),
                    },
                    "mingle:core@v1/Null",
                    []string{},
                ),
            },
            &types.PrototypeDefinition{
                Name: mkQn( "ns1@v1/Sec2" ),
                Signature: types.MakeCallSig(
                    []*types.FieldDefinition{
                        types.MakeFieldDef( 
                            "authentication", "ns1@v1/Struct1", nil ),
                    },
                    "mingle:core@v1/Int64~[9,10]",
                    []string{ "ns1@v1/Exception1", "ns1@v1/Exception2" },
                ),
            },
            types.MakeServiceDef(
                "ns1@v1/Service1",
                "",
                "",
                types.MakeOpDef( "op1",
                    types.MakeCallSig(
                        []*types.FieldDefinition{},
                        "mingle:core@v1/String*",
                        []string{},
                    ),
                ),
                types.MakeOpDef( "op2",
                    types.MakeCallSig(
                        []*types.FieldDefinition{
                            types.MakeFieldDef( 
                                "param1", "mingle:core@v1/String", nil ),
                            types.MakeFieldDef(
                                "param2", "&ns1@v1/Struct1?", nil ),
                            types.MakeFieldDef(
                                "param3", "mingle:core@v1/Int64", int64( 12 ) ),
                            types.MakeFieldDef(
                                "param4", "mingle:core@v1/String?*", nil ),
                            types.MakeFieldDef( 
                                "param5", "ns1@v1/Struct1", nil ),
                        },
                        "ns1@v1/Struct2",
                        []string{ "ns1@v1/Exception1", "ns1@v1/Exception3" },
                    ),
                ),
                types.MakeOpDef( "op3",
                    types.MakeCallSig(
                        []*types.FieldDefinition{},
                        "&mingle:core@v1/Int64?",
                        []string{ "ns1@v1/Exception2" },
                    ),
                ),
                types.MakeOpDef( "op4",
                    types.MakeCallSig(
                        []*types.FieldDefinition{},
                        "mingle:core@v1/Null",
                        []string{},
                    ),
                ),
            ),
            types.MakeServiceDef(
                "ns1@v1/Service2",
                "",
                "ns1@v1/Sec1",
                types.MakeOpDef( "op1",
                    types.MakeCallSig(
                        []*types.FieldDefinition{},
                        "mingle:core@v1/Int64",
                        []string{},
                    ),
                ),
                types.MakeOpDef( "op2",
                    types.MakeCallSig(
                        []*types.FieldDefinition{},
                        "mingle:core@v1/Boolean",
                        []string{},
                    ),
                ),
            ),
            types.MakeServiceDef( "ns1@v1/Service3", "", "ns1@v1/Sec2" ),
            types.MakeStructDef(
                "ns1@v1/FieldConstantTester",
                "",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f1", "mingle:core@v1/Boolean", true ),
                    types.MakeFieldDef( "f2", "mingle:core@v1/Boolean", false ),
                    types.MakeFieldDef( 
                        "f3", "mingle:core@v1/Int32", int32( 1 ) ),
                    types.MakeFieldDef( 
                        "f4", "mingle:core@v1/Int32", int32( -1 ) ),
                    types.MakeFieldDef( 
                        "f5", "mingle:core@v1/Int64", int64( 1 ) ),
                    types.MakeFieldDef( 
                        "f6", "mingle:core@v1/Int64", int64( -1 ) ),
                    types.MakeFieldDef( 
                        "f7", "mingle:core@v1/Float32", float32( 1.0 ) ),
                    types.MakeFieldDef(
                        "f8", "mingle:core@v1/Float32", float32( -1.0 ) ),
                    types.MakeFieldDef(
                        "f9", "mingle:core@v1/Float64", float64( 1.0 ) ),
                    types.MakeFieldDef(
                        "f10", "mingle:core@v1/Float64", float64( -1.0 ) ),
                    types.MakeFieldDef(
                        "f11", "mingle:core@v1/Int32~[0,10)", int32( 8 ) ),
                    types.MakeFieldDef( "f12", "mingle:core@v1/String", "a" ),
                    types.MakeFieldDef( 
                        "f13", `mingle:core@v1/String~"a"`, "a" ),
                    types.MakeFieldDef( "f14", "ns1@v1/Enum1",
                        mg.MustEnum( "ns1@v1/Enum1", "green" ),
                    ),
                    types.MakeFieldDef( "f15", "mingle:core@v1/Timestamp",
                        mg.MustTimestamp( 
                            "2007-08-24T14:15:43.123450000-07:00" ) ),
                    types.MakeFieldDef( 
                        "f16", "mingle:core@v1/String+",
                        []interface{}{ "a", "b", "c" } ),
                    types.MakeFieldDef(
                        "f17", "mingle:core@v1/Int32*",
                        []interface{}{ int32( 1 ), int32( 2 ), int32( 3 ) } ),
                    types.MakeFieldDef(
                        "f18", "mingle:core@v1/Float64*", []interface{}{} ),
                    types.MakeFieldDef(
                        "f19", "mingle:core@v1/String*+",
                        []interface{}{
                            []interface{}{},
                            []interface{}{ "a", "b" },
                            []interface{}{ "c", "d", "e" },
                        },
                    ),
                    types.MakeFieldDef( 
                        "f20", "mingle:core@v1/Uint32", uint32( 1 ) ),
                    types.MakeFieldDef( 
                        "f21", "mingle:core@v1/Uint32", uint32( 4294967295 ) ),
                    types.MakeFieldDef( 
                        "f22", "mingle:core@v1/Uint64", uint64( 0 ) ),
                    types.MakeFieldDef(
                        "f23", 
                        "mingle:core@v1/Uint64", 
                        uint64( 18446744073709551615 ),
                    ),
                },
            ),
        ),
        types.MakeDefMap(
            types.MakeStructDef(
                "ns1@v1/Struct4",
                "",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "inst1", "ns1@v1/Struct1", nil ),
                    types.MakeFieldDef( "inst2", "ns1@v1/Struct2", nil ),
                    types.MakeFieldDef( "str1", "mingle:core@v1/String?", nil ),
                    types.MakeFieldDef( 
                        "str2", "mingle:core@v1/String?*", nil ),
                },
            ),
            types.MakeStructDef(
                "ns1@v1/Exception4",
                "ns1@v1/Exception3",
                []*types.FieldDefinition{
                    types.MakeFieldDef( 
                        "int1", "mingle:core@v1/Int64", int64( 33 ) ),
                },
            ),
        ),
        types.MakeDefMap(
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns2@v1/Alias1" ),
                AliasedType: mkTyp( "ns1@v1/Struct3" ),
            },
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns2@v1/Alias2" ),
                AliasedType: mkTyp( "mingle:core@v1/String?*" ),
            },
            types.MakeStructDef(
                "ns2@v1/Struct1",
                "",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "inst1", "&ns1@v1/Struct1?", nil ),
                    types.MakeFieldDef( "inst2", "ns2@v1/Struct2", nil ),
                    types.MakeFieldDef( "inst3", "ns1@v1/Struct4", nil ),
                    types.MakeFieldDef( "inst4", "ns1@v1/Struct3", nil ),
                    types.MakeFieldDef( "inst5", "ns1@v1/Struct1", nil ),
                    types.MakeFieldDef( 
                        "inst6", "mingle:core@v1/String?*", nil ),
                },
            ),
            types.MakeStructDef(
                "ns2@v1/Struct2",
                "",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "inst1", "ns2@v1/Struct1", nil ),
                    types.MakeFieldDef( "inst2", "ns1@v1/Struct1", nil ),
                },
            ),
            types.MakeStructDef(
                "ns2@v1/Exception1",
                "ns1@v1/Exception1",
                []*types.FieldDefinition{},
            ),
            types.MakeStructDef(
                "ns2@v1/Exception2",
                "ns1@v1/Exception3",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "str1", "mingle:core@v1/String*", nil ),
                },
            ),
            types.MakeServiceDef(
                "ns2@v1/Service1",
                "",
                "",
                types.MakeOpDef( "op1",
                    types.MakeCallSig(
                        []*types.FieldDefinition{},
                        "ns2@v1/Struct2",
                        []string{ "ns2@v1/Exception1", "ns1@v1/Exception4" },
                    ),
                ),
                types.MakeOpDef( "op2",
                    types.MakeCallSig(
                        []*types.FieldDefinition{
                            types.MakeFieldDef( 
                                "param1", "mingle:core@v1/String*+", nil ),
                            types.MakeFieldDef(
                                "param2", "ns1@v1/Struct4*", nil ),
                            types.MakeFieldDef( 
                                "param3", "ns1@v1/Struct1", nil ),
                            types.MakeFieldDef( 
                                "param4", "ns2@v1/Struct2", nil ),
                        },
                        "mingle:core@v1/String",
                        []string{},
                    ),
                ),
            ),
        ),
        types.MakeDefMap(
            types.MakeStructDef(
                "ns1:globTestNs@v1/Struct1",
                "ns1@v1/Struct1",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "inst1", "ns1@v1/Struct2", nil ),
                    types.MakeFieldDef( "inst2", "ns1@v1/Struct4", nil ),
                    types.MakeFieldDef( 
                        "inst3", "ns1:globTestNs@v1/Struct1", nil ),
                },
            ),
            types.MakeStructDef(
                "ns1:globTestNs@v1/Exception2",
                "",
                []*types.FieldDefinition{},
            ),
            types.MakeServiceDef(
                "ns1:globTestNs@v1/Service1",
                "",
                "",
                types.MakeOpDef( "op1",
                    types.MakeCallSig(
                        []*types.FieldDefinition{
                            types.MakeFieldDef( 
                                "param1", "ns1:globTestNs@v1/Struct1*", nil ),
                            types.MakeFieldDef( 
                                "param2", "ns1@v1/Struct1", nil ),
                            types.MakeFieldDef( 
                                "param3", "ns1@v1/Struct3", nil ),
                        },
                        "mingle:core@v1/String",
                        []string{
                            "ns2@v1/Exception1",
                            "ns1:globTestNs@v1/Exception2",
                            "ns1@v1/Exception3",
                        },
                    ),
                ),
            ),
        ),
        types.MakeDefMap(
            types.MakeStructDef(
                "ns1@v2/Struct1",
                "",
                []*types.FieldDefinition{
                    types.MakeFieldDef( 
                        "f1", "mingle:core@v1/String", "hello" ),
                },
            ),
            types.MakeStructDef(
                "ns1@v2/Struct2",
                "ns1@v1/Struct1",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f1", "&ns1@v2/Struct1?", nil ),
                },
            ),
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v2/Struct1V1" ),
                AliasedType: mkTyp( "ns1@v1/Struct1" ),
            },
            types.MakeStructDef(
                "ns1@v2/Struct3",
                "ns1@v2/Struct1",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f2", "ns1@v2/Struct2", nil ),
                    types.MakeFieldDef( "f3", "mingle:core@v1/String?", nil ),
                    types.MakeFieldDef( "f4", "ns1@v1/Struct1*", nil ),
                },
            ),
            types.MakeServiceDef(
                "ns1@v2/Service1",
                "",
                "",
                types.MakeOpDef( "op1",
                    types.MakeCallSig(
                        []*types.FieldDefinition{
                            types.MakeFieldDef( "f1", "ns1@v2/Struct1", nil ),
                            types.MakeFieldDef(
                                "f2", "mingle:core@v1/Int64~[0,12)", nil ),
                            types.MakeFieldDef( "f3", "ns1@v1/Struct1*+", nil ),
                        },
                        "mingle:core@v1/Null",
                        []string{},
                    ),
                ),
            ),
        ),
    }
}
