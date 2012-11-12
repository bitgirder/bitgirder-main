package compiler

import (
    "testing"
    "fmt"
//    "log"
    "bytes"
    mg "mingle"
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
    da := newDefAsserter( r.T )
    built.EachDefinition(
        func( defAct types.Definition ) {
            qn := defAct.GetName()
            if expct := m.Get( qn ); expct != nil {
                defExpct := expct.( types.Definition )
                da.descend( qn.ExternalForm() ).assertDef( defExpct, defAct )
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
//    for i, e := 1, 1; i <= e; i++ {
        for j := 0; j < 2; j++ {
//        for j := 0; j < 1; j++ {
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
    bool1 Boolean?
    bool2 Boolean default true
    buf1 Buffer?
    timestamp1 Timestamp?
    timestamp2 Timestamp default "2007-08-24T13:15:43.123450000-08:00"
    int1 Int64
    int2 Int64 default 1234
    int3 Int64?
    int4 Int32 default 12
    int5 Int32~[0,) default 1111
    int6 Int64~(,)
    int7 Uint32
    int8 Uint64~[0,100)
    ints1 Int64*
    ints2 Int32+ default [ 1, -2, 3, -4 ]
    float1 Float64 default 3.1
    float2 Float64~(-1e-10,3]?
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
    enum1 Enum1?
    enum2 Enum1 default Enum1.green

    @constructor( Int64 )
    @constructor( ns1/Struct1 )
    @constructor( String~"^a+$" )
}

struct Exception1 < StandardException {}

struct Exception2 { failTime Int64 }

struct Exception3 < Exception1 { string2 String* }

alias Alias1 String?
alias Alias2 Struct1
alias Alias3 Alias1*
alias Alias4 String~"^a+$"
alias Alias5 Int64~[0,)

struct Struct5 {
    f1 Alias1
    f2 Alias1? default "hello"
    f3 Alias1*
    f4 Alias1+ default [ "a", "b" ]
    f5 Alias1?*+
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
         param2 Struct1?,
         param3 Int64 default 12,
         param4 Alias1*,
         param5 Alias2 ): ns1/Struct2,
            throws Exception1, Exception3
    
    op op3(): Int64? throws Exception2

    op op4(): Null
}

prototype Proto1(): String
prototype Proto2(): String throws Exception1
prototype Proto3( f1 Struct1, f2 String default "hi" ): Struct1?

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
    f15 Timestamp default "2007-08-24T13:15:43.123450000-08:00"
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
    inst1 ns1@v1/Struct1?
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

struct Struct2 < ns1@v1/Struct1 { f1 Struct1? }

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
    mkId := mg.MustIdentifier
    mkQn := mg.MustQualifiedTypeName
    mkTyp := mg.MustTypeReference
    p1Defs = []*types.DefinitionMap{
        makeDefMap(
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
            makeStructDef(
                "ns1@v1/Struct1",
                "",
                []*types.FieldDefinition{
                    makeFieldDef( "string1", "mingle:core@v1/String", nil ),
                    makeFieldDef( "string2", "mingle:core@v1/String?", nil ),
                    makeFieldDef( 
                        "string3", "mingle:core@v1/String", "hello there" ),
                    makeFieldDef(
                        "string4", `mingle:core@v1/String~"a*"`, "aaaaa" ),
                    makeFieldDef(
                        "string5", `mingle:core@v1/String~"^.*(a|b)$"?`, nil ),
                    makeFieldDef( "bool1", "mingle:core@v1/Boolean?", nil ),
                    makeFieldDef( "bool2", "mingle:core@v1/Boolean", true ),
                    makeFieldDef( "buf1", "mingle:core@v1/Buffer?", nil ),
                    makeFieldDef( 
                        "timestamp1", "mingle:core@v1/Timestamp?", nil ),
                    makeFieldDef(
                        "timestamp2", 
                        "mingle:core@v1/Timestamp",
                        mg.MustTimestamp( 
                            "2007-08-24T13:15:43.123450000-08:00" ),
                    ),
                    makeFieldDef( "int1", "mingle:core@v1/Int64", nil ),
                    makeFieldDef( 
                        "int2", "mingle:core@v1/Int64", int64( 1234 ) ),
                    makeFieldDef( "int3", "mingle:core@v1/Int64?", nil ),
                    makeFieldDef( "int4", "mingle:core@v1/Int32", int32( 12 ) ),
                    makeFieldDef( 
                        "int5", "mingle:core@v1/Int32~[0,)", int32( 1111 ) ),
                    makeFieldDef( "int6", "mingle:core@v1/Int64~(,)", nil ),
                    makeFieldDef( "int7", "mingle:core@v1/Uint32", nil ),
                    makeFieldDef( 
                        "int8", "mingle:core@v1/Uint64~[0,100)", nil ),
                    makeFieldDef( "ints1", "mingle:core@v1/Int64*", nil ),
                    makeFieldDef(
                        "ints2",
                        "mingle:core@v1/Int32+",
                        []interface{}{ 
                            int32( 1 ), int32( -2 ), int32( 3 ), int32( -4 ) },
                    ),
                    makeFieldDef(
                        "float1", "mingle:core@v1/Float64", float64( 3.1 ) ),
                    makeFieldDef(
                        "float2", "mingle:core@v1/Float64~(-1e-10,3]?", nil ),
                    makeFieldDef(
                        "float3", "mingle:core@v1/Float32", float32( 3.2 ) ),
                    makeFieldDef( "floats1", "mingle:core@v1/Float32*", nil ),
                    makeFieldDef( "val1", "mingle:core@v1/Value", nil ),
                    makeFieldDef( "val2", "mingle:core@v1/Value", int64( 12 ) ),
                    makeFieldDef( "list1", "mingle:core@v1/String*", nil ),
                    makeFieldDef( "list2", "mingle:core@v1/String**", nil ),
                    makeFieldDef( 
                        "list3", `mingle:core@v1/String~"abc$"*`, nil ),
                },
            ),
            makeStructDef(
                "ns1@v1/Struct2",
                "",
                []*types.FieldDefinition{
                    makeFieldDef( "inst1", "ns1@v1/Struct1", nil ),
                    makeFieldDef( "inst2", "ns1@v1/Struct1*", nil ),
                },
            ),
            makeStructDef2(
                "ns1@v1/Struct3",
                "ns1@v1/Struct1",
                []*types.FieldDefinition{
                    makeFieldDef( "string6", "mingle:core@v1/String?", nil ),
                    makeFieldDef( "inst1", "ns1@v1/Struct2", nil ),
                    makeFieldDef( "enum1", "ns1@v1/Enum1?", nil ),
                    makeFieldDef( 
                        "enum2",
                        "ns1@v1/Enum1",
                        &mg.Enum{
                            Type: mg.MustTypeReference( "ns1@v1/Enum1" ),
                            Value: mg.MustIdentifier( "green" ),
                        },
                    ),
                },
                []*types.ConstructorDefinition{
                    { mg.MustTypeReference( "mingle:core@v1/Int64" ) },
                    { mg.MustTypeReference( "ns1@v1/Struct1" ) },
                    { mg.MustTypeReference( `mingle:core@v1/String~"^a+$"` ) },
                },
            ),
            makeStructDef(
                "ns1@v1/Struct5",
                "",
                []*types.FieldDefinition{
                    makeFieldDef(
                        "f1", "mingle:core@v1/String?", nil ),
                    makeFieldDef(
                        "f2", "mingle:core@v1/String??", "hello" ),
                    makeFieldDef(
                        "f3", "mingle:core@v1/String?*", nil ),
                    makeFieldDef(
                        "f4", 
                        "mingle:core@v1/String?+", 
                        []interface{}{ "a", "b" },
                    ),
                    makeFieldDef(
                        "f5", "mingle:core@v1/String??*+", nil ),
                    makeFieldDef(
                        "f6", "ns1@v1/Struct1", nil ),
                    makeFieldDef(
                        "f7", "ns1@v1/Struct1*", nil ),
                    makeFieldDef(
                        "f8", 
                        "mingle:core@v1/String?*", 
                        []interface{}{ "hello" },
                    ),
                    makeFieldDef(
                        "f9", "mingle:core@v1/String?*+", nil ),
                    makeFieldDef(
                        "f10", `mingle:core@v1/String~"^a+$"`, "aaa" ),
                    makeFieldDef(
                        "f11", "mingle:core@v1/Int64~[0,)", int64( 12 ) ),
                },
            ),
            makeStructDef(
                "ns1@v1/Exception1",
                "mingle:core@v1/StandardException",
                []*types.FieldDefinition{},
            ),
            makeStructDef(
                "ns1@v1/Exception2",
                "",
                []*types.FieldDefinition{
                    makeFieldDef( "failTime", "mingle:core@v1/Int64", nil ),
                },
            ),
            makeStructDef(
                "ns1@v1/Exception3",
                "ns1@v1/Exception1",
                []*types.FieldDefinition{
                    makeFieldDef( "string2", "mingle:core@v1/String*", nil ),
                },
            ),
            makeEnumDef( "ns1@v1/Enum1", "red", "green", "lightGrey" ),
            &types.PrototypeDefinition{
                Name: mkQn( "ns1@v1/Proto1" ),
                Signature: makeCallSig(
                    []*types.FieldDefinition{},
                    "mingle:core@v1/String",
                    []string{},
                ),
            },
            &types.PrototypeDefinition{
                Name: mkQn( "ns1@v1/Proto2" ),
                Signature: makeCallSig(
                    []*types.FieldDefinition{},
                    "mingle:core@v1/String",
                    []string{ "ns1@v1/Exception1" },
                ),
            },
            &types.PrototypeDefinition{
                Name: mkQn( "ns1@v1/Proto3" ),
                Signature: makeCallSig(
                    []*types.FieldDefinition{
                        makeFieldDef( "f1", "ns1@v1/Struct1", nil ),
                        makeFieldDef( "f2", "mingle:core@v1/String", "hi" ),
                    },
                    "ns1@v1/Struct1?",
                    []string{},
                ),
            },
            &types.PrototypeDefinition{
                Name: mkQn( "ns1@v1/Sec1" ),
                Signature: makeCallSig(
                    []*types.FieldDefinition{
                        makeFieldDef( "authentication", "ns1@v1/Struct1", nil ),
                    },
                    "mingle:core@v1/Null",
                    []string{},
                ),
            },
            &types.PrototypeDefinition{
                Name: mkQn( "ns1@v1/Sec2" ),
                Signature: makeCallSig(
                    []*types.FieldDefinition{
                        makeFieldDef( "authentication", "ns1@v1/Struct1", nil ),
                    },
                    "mingle:core@v1/Int64~[9,10]",
                    []string{ "ns1@v1/Exception1", "ns1@v1/Exception2" },
                ),
            },
            makeServiceDef(
                "ns1@v1/Service1",
                "",
                []*types.OperationDefinition{
                    {   Name: mkId( "op1" ),
                        Signature: makeCallSig(
                            []*types.FieldDefinition{},
                            "mingle:core@v1/String*",
                            []string{},
                        ),
                    },
                    {   Name: mkId( "op2" ),
                        Signature: makeCallSig(
                            []*types.FieldDefinition{
                                makeFieldDef( 
                                    "param1", "mingle:core@v1/String", nil ),
                                makeFieldDef(
                                    "param2", "ns1@v1/Struct1?", nil ),
                                makeFieldDef(
                                    "param3",
                                    "mingle:core@v1/Int64", 
                                    int64( 12 ),
                                ),
                                makeFieldDef(
                                    "param4", "mingle:core@v1/String?*", nil ),
                                makeFieldDef( "param5", "ns1@v1/Struct1", nil ),
                            },
                            "ns1@v1/Struct2",
                            []string{ 
                                "ns1@v1/Exception1", "ns1@v1/Exception3" },
                        ),
                    },
                    {   Name: mkId( "op3" ),
                        Signature: makeCallSig(
                            []*types.FieldDefinition{},
                            "mingle:core@v1/Int64?",
                            []string{ "ns1@v1/Exception2" },
                        ),
                    },
                    {   Name: mkId( "op4" ),
                        Signature: makeCallSig(
                            []*types.FieldDefinition{},
                            "mingle:core@v1/Null",
                            []string{},
                        ),
                    },
                },
                "",
            ),
            makeServiceDef(
                "ns1@v1/Service2",
                "",
                []*types.OperationDefinition{
                    {   Name: mkId( "op1" ),
                        Signature: makeCallSig(
                            []*types.FieldDefinition{},
                            "mingle:core@v1/Int64",
                            []string{},
                        ),
                    },
                    {   Name: mkId( "op2" ),
                        Signature: makeCallSig(
                            []*types.FieldDefinition{},
                            "mingle:core@v1/Boolean",
                            []string{},
                        ),
                    },
                },
                "ns1@v1/Sec1",
            ),
            makeServiceDef(
                "ns1@v1/Service3",
                "",
                []*types.OperationDefinition{},
                "ns1@v1/Sec2",
            ),
            makeStructDef(
                "ns1@v1/FieldConstantTester",
                "",
                []*types.FieldDefinition{
                    makeFieldDef( "f1", "mingle:core@v1/Boolean", true ),
                    makeFieldDef( "f2", "mingle:core@v1/Boolean", false ),
                    makeFieldDef( "f3", "mingle:core@v1/Int32", int32( 1 ) ),
                    makeFieldDef( "f4", "mingle:core@v1/Int32", int32( -1 ) ),
                    makeFieldDef( "f5", "mingle:core@v1/Int64", int64( 1 ) ),
                    makeFieldDef( "f6", "mingle:core@v1/Int64", int64( -1 ) ),
                    makeFieldDef( 
                        "f7", "mingle:core@v1/Float32", float32( 1.0 ) ),
                    makeFieldDef(
                        "f8", "mingle:core@v1/Float32", float32( -1.0 ) ),
                    makeFieldDef(
                        "f9", "mingle:core@v1/Float64", float64( 1.0 ) ),
                    makeFieldDef(
                        "f10", "mingle:core@v1/Float64", float64( -1.0 ) ),
                    makeFieldDef(
                        "f11", "mingle:core@v1/Int32~[0,10)", int32( 8 ) ),
                    makeFieldDef( "f12", "mingle:core@v1/String", "a" ),
                    makeFieldDef( "f13", `mingle:core@v1/String~"a"`, "a" ),
                    makeFieldDef( "f14", "ns1@v1/Enum1",
                        &mg.Enum{
                            Type: mkTyp( "ns1@v1/Enum1" ), 
                            Value: mkId( "green" ),
                        },
                    ),
                    makeFieldDef( "f15", "mingle:core@v1/Timestamp",
                        mg.MustTimestamp( 
                            "2007-08-24T13:15:43.123450000-08:00" ) ),
                    makeFieldDef( 
                        "f16", "mingle:core@v1/String+",
                        []interface{}{ "a", "b", "c" } ),
                    makeFieldDef(
                        "f17", "mingle:core@v1/Int32*",
                        []interface{}{ int32( 1 ), int32( 2 ), int32( 3 ) } ),
                    makeFieldDef(
                        "f18", "mingle:core@v1/Float64*", []interface{}{} ),
                    makeFieldDef(
                        "f19", "mingle:core@v1/String*+",
                        []interface{}{
                            []interface{}{},
                            []interface{}{ "a", "b" },
                            []interface{}{ "c", "d", "e" },
                        },
                    ),
                    makeFieldDef( "f20", "mingle:core@v1/Uint32", uint32( 1 ) ),
                    makeFieldDef( 
                        "f21", "mingle:core@v1/Uint32", uint32( 4294967295 ) ),
                    makeFieldDef( "f22", "mingle:core@v1/Uint64", uint64( 0 ) ),
                    makeFieldDef(
                        "f23", 
                        "mingle:core@v1/Uint64", 
                        uint64( 18446744073709551615 ),
                    ),
                },
            ),
        ),
        makeDefMap(
            makeStructDef(
                "ns1@v1/Struct4",
                "",
                []*types.FieldDefinition{
                    makeFieldDef( "inst1", "ns1@v1/Struct1", nil ),
                    makeFieldDef( "inst2", "ns1@v1/Struct2", nil ),
                    makeFieldDef( "str1", "mingle:core@v1/String?", nil ),
                    makeFieldDef( "str2", "mingle:core@v1/String?*", nil ),
                },
            ),
            makeStructDef(
                "ns1@v1/Exception4",
                "ns1@v1/Exception3",
                []*types.FieldDefinition{
                    makeFieldDef( "int1", "mingle:core@v1/Int64", int64( 33 ) ),
                },
            ),
        ),
        makeDefMap(
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns2@v1/Alias1" ),
                AliasedType: mkTyp( "ns1@v1/Struct3" ),
            },
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns2@v1/Alias2" ),
                AliasedType: mkTyp( "mingle:core@v1/String?*" ),
            },
            makeStructDef(
                "ns2@v1/Struct1",
                "",
                []*types.FieldDefinition{
                    makeFieldDef( "inst1", "ns1@v1/Struct1?", nil ),
                    makeFieldDef( "inst2", "ns2@v1/Struct2", nil ),
                    makeFieldDef( "inst3", "ns1@v1/Struct4", nil ),
                    makeFieldDef( "inst4", "ns1@v1/Struct3", nil ),
                    makeFieldDef( "inst5", "ns1@v1/Struct1", nil ),
                    makeFieldDef( "inst6", "mingle:core@v1/String?*", nil ),
                },
            ),
            makeStructDef(
                "ns2@v1/Struct2",
                "",
                []*types.FieldDefinition{
                    makeFieldDef( "inst1", "ns2@v1/Struct1", nil ),
                    makeFieldDef( "inst2", "ns1@v1/Struct1", nil ),
                },
            ),
            makeStructDef(
                "ns2@v1/Exception1",
                "ns1@v1/Exception1",
                []*types.FieldDefinition{},
            ),
            makeStructDef(
                "ns2@v1/Exception2",
                "ns1@v1/Exception3",
                []*types.FieldDefinition{
                    makeFieldDef( "str1", "mingle:core@v1/String*", nil ),
                },
            ),
            makeServiceDef(
                "ns2@v1/Service1",
                "",
                []*types.OperationDefinition{
                    {   Name: mkId( "op1" ),
                        Signature: makeCallSig(
                            []*types.FieldDefinition{},
                            "ns2@v1/Struct2",
                            []string{ 
                                "ns2@v1/Exception1", "ns1@v1/Exception4" },
                        ),
                    },
                    {   Name: mkId( "op2" ),
                        Signature: makeCallSig(
                            []*types.FieldDefinition{
                                makeFieldDef( 
                                    "param1", "mingle:core@v1/String*+", nil ),
                                makeFieldDef(
                                    "param2", "ns1@v1/Struct4*", nil ),
                                makeFieldDef( "param3", "ns1@v1/Struct1", nil ),
                                makeFieldDef( "param4", "ns2@v1/Struct2", nil ),
                            },
                            "mingle:core@v1/String",
                            []string{},
                        ),
                    },
                },
                "",
            ),
        ),
        makeDefMap(
            makeStructDef(
                "ns1:globTestNs@v1/Struct1",
                "ns1@v1/Struct1",
                []*types.FieldDefinition{
                    makeFieldDef( "inst1", "ns1@v1/Struct2", nil ),
                    makeFieldDef( "inst2", "ns1@v1/Struct4", nil ),
                    makeFieldDef( "inst3", "ns1:globTestNs@v1/Struct1", nil ),
                },
            ),
            makeStructDef(
                "ns1:globTestNs@v1/Exception2",
                "",
                []*types.FieldDefinition{},
            ),
            makeServiceDef(
                "ns1:globTestNs@v1/Service1",
                "",
                []*types.OperationDefinition{
                    {   Name: mkId( "op1" ),
                        Signature: makeCallSig(
                            []*types.FieldDefinition{
                                makeFieldDef( 
                                    "param1", 
                                    "ns1:globTestNs@v1/Struct1*", 
                                    nil,
                                ),
                                makeFieldDef( "param2", "ns1@v1/Struct1", nil ),
                                makeFieldDef( "param3", "ns1@v1/Struct3", nil ),
                            },
                            "mingle:core@v1/String",
                            []string{
                                "ns2@v1/Exception1",
                                "ns1:globTestNs@v1/Exception2",
                                "ns1@v1/Exception3",
                            },
                        ),
                    },
                },
                "",
            ),
        ),
        makeDefMap(
//`@version v2
//
//import ns1@v1/Struct1 # import that should be shadowed
//
//namespace ns1
            makeStructDef(
                "ns1@v2/Struct1",
                "",
                []*types.FieldDefinition{
                    makeFieldDef( "f1", "mingle:core@v1/String", "hello" ),
                },
            ),
            makeStructDef(
                "ns1@v2/Struct2",
                "ns1@v1/Struct1",
                []*types.FieldDefinition{
                    makeFieldDef( "f1", "ns1@v2/Struct1?", nil ),
                },
            ),
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v2/Struct1V1" ),
                AliasedType: mkTyp( "ns1@v1/Struct1" ),
            },
            makeStructDef(
                "ns1@v2/Struct3",
                "ns1@v2/Struct1",
                []*types.FieldDefinition{
                    makeFieldDef( "f2", "ns1@v2/Struct2", nil ),
                    makeFieldDef( "f3", "mingle:core@v1/String?", nil ),
                    makeFieldDef( "f4", "ns1@v1/Struct1*", nil ),
                },
            ),
            makeServiceDef(
                "ns1@v2/Service1",
                "",
                []*types.OperationDefinition{
                    {   Name: mkId( "op1" ),
                        Signature: makeCallSig(
                            []*types.FieldDefinition{
                                makeFieldDef( "f1", "ns1@v2/Struct1", nil ),
                                makeFieldDef(
                                    "f2",
                                    "mingle:core@v1/Int64~[0,12)",
                                    nil,
                                ),
                                makeFieldDef( "f3", "ns1@v1/Struct1*+", nil ),
                            },
                            "mingle:core@v1/Null",
                            []string{},
                        ),
                    },
                },
                "",
            ),
        ),
    }
}
