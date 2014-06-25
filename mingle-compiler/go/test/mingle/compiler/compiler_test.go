package compiler

import (
    "testing"
    "bitgirder/assert"
    "mingle/types"
    "mingle/parser"
)

func TestCompiler( t *testing.T ) {
    tests := []*compilerTest{

        newCompilerTest( "base-field-types" ).
        addSource( "f1", `
            @version v1
            
            namespace ns1
            
            struct Struct1 {
                string1 String # required field with no default
                string2 String? # nullable String
                string3 String default "hello there"
                string4 String~"a*" default "aaaaa"
                string5 String~"^.*(a|b)$"?
                string6 String~[ "aaa", "aab" ]
                bool1 &Boolean?
                bool2 Boolean default true
                buf1 Buffer?
                timestamp1 &Timestamp?
                timestamp2 Timestamp default "2007-08-24T14:15:43.123450000-07:00"
                timestamp3 Timestamp~[
                    "2007-08-24T14:15:43.123450000-07:00",
                    "2008-08-24T14:15:43.123450000-07:00" ]
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
        ` ).
        expectDef(
            types.MakeStructDef(
                "ns1@v1/Struct1",
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
                        "string6", `mingle:core@v1/String~["aaa","aab"]`, nil ),
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
                        parser.MustTimestamp( 
                            "2007-08-24T14:15:43.123450000-07:00" ),
                    ),
                    types.MakeFieldDef(
                        "timestamp3",
                        `mingle:core@v1/Timestamp~["2007-08-24T14:15:43.123450000-07:00","2008-08-24T14:15:43.123450000-07:00"]`, nil ),
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
        ).
        expectDef(
            types.MakeStructDef(
                "ns1@v1/Struct2",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "inst1", "ns1@v1/Struct1", nil ),
                    types.MakeFieldDef( "inst2", "ns1@v1/Struct1*", nil ),
                },
            ),
        ),

        newCompilerTest( "enum-as-field-type" ).
        addSource( "f1", `
            
            @version v1

            namespace ns1

            enum Enum1 { red, green, lightGrey }
            
            struct Struct1 {}
            
            struct Struct2 {}

            struct Struct3 {
                string1 String?
                inst1 Struct2
                enum1 &Enum1?
                enum2 Enum1 default Enum1.green
            
                @constructor( Int64 )
                @constructor( ns1/Struct1 )
                @constructor( String~"^a+$" )
            }
        ` ).
        expectDef( types.MakeStructDef( "ns1@v1/Struct1", nil ) ).
        expectDef( types.MakeStructDef( "ns1@v1/Struct2", nil ) ).
        expectDef(
            makeStructDefWithConstructors(
                "ns1@v1/Struct3",
                []*types.FieldDefinition{
                    types.MakeFieldDef( 
                        "string1", "mingle:core@v1/String?", nil ),
                    types.MakeFieldDef( "inst1", "ns1@v1/Struct2", nil ),
                    types.MakeFieldDef( "enum1", "&ns1@v1/Enum1?", nil ),
                    types.MakeFieldDef( 
                        "enum2",
                        "ns1@v1/Enum1",
                        parser.MustEnum( "ns1@v1/Enum1", "green" ),
                    ),
                },
                []*types.ConstructorDefinition{
                    { mkTyp( "mingle:core@v1/Int64" ) },
                    { mkTyp( "ns1@v1/Struct1" ) },
                    { mkTyp( `mingle:core@v1/String~"^a+$"` ) },
                },
            ),
        ).
        expectDef(
            types.MakeEnumDef( "ns1@v1/Enum1", "red", "green", "lightGrey" ),
        ),

        newCompilerTest( "schema-variations" ).
        addLib( "lib1", 
            `@version v1; namespace ns2; schema Schema1 { g1 Int32 }` ).
        addSource( "f1", `
            @version v1
            namespace ns1

            # we forward declare and reference some schemas so we can check that
            # the compiler is actually reordering correctly.
            schema Schema4 { 
                f2 Int32
                f3 Int32
                @schema Schema3 
            }

            schema Schema1 { f1 Int32 }

            schema Schema2 { f2 Int32 }

            schema Schema3 { 
                f1 Int32
                f2 Int32 
            }

            schema Schema5 { 
                @schema Schema1
                @schema Schema2
                @schema Schema3
                @schema Schema4
            }

            alias Schema6 Schema1

            schema Schema7 {
                @schema Schema6
                f7 Int32
            }

            struct S1 {
                @schema Schema1 
            }

            struct S2 {
                @schema Schema5
                s2F1 Int32
            }

            struct S3 { 
                @schema Schema6 
            }

            schema Schema8 { 
                f1 Int32
                @schema ns2@v1/Schema1
            }

            struct S4 { 
                @schema ns2@v1/Schema1
                g2 Int32
            }

            schema Schema9 {}

            struct S5 { @schema Schema9 }

            struct S6 { 
                f1 Int32
                @schema Schema9
            }

            schema Schema10 { f1 Int32 default 1 }

            struct S7 {
                @schema Schema10
                f2 Int64
            }

            struct S8 {
                f1 Int32 default 1
                @schema Schema10
            }
        ` ).
        expectDef( 
            types.MakeStructDef( "ns1@v1/S1", 
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f1", "Int32", nil ),
                },
            ),
        ).
        expectDef(
            types.MakeStructDef( "ns1@v1/S2",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f1", "Int32", nil ),
                    types.MakeFieldDef( "f2", "Int32", nil ),
                    types.MakeFieldDef( "f3", "Int32", nil ),
                    types.MakeFieldDef( "s2F1", "Int32", nil ),
                },
            ),
        ).
        expectDef(
            types.MakeStructDef( "ns1@v1/S3",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f1", "Int32", nil ),
                },
            ),
        ).
        expectDef(
            types.MakeStructDef( "ns1@v1/S4",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "g1", "Int32", nil ),
                    types.MakeFieldDef( "g2", "Int32", nil ),
                },
            ),
        ).
        expectDef( types.MakeStructDef( "ns1@v1/S5", nil ) ).
        expectDef(
            types.MakeStructDef( "ns1@v1/S6",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f1", "Int32", nil ),
                },
            ),
        ).
        expectDef(
            types.MakeStructDef( "ns1@v1/S7",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f1", "Int32", int32( 1 ) ),
                    types.MakeFieldDef( "f2", "Int64", nil ),
                },
            ),
        ).
        expectDef(
            types.MakeStructDef( "ns1@v1/S8",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f1", "Int32", int32( 1 ) ),
                },
            ),
        ).
        expectDef(
            types.MakeSchemaDef( "ns1@v1/Schema1",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f1", "Int32", nil ),
                },
            ),
        ).
        expectDef(
            types.MakeSchemaDef( "ns1@v1/Schema2",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f2", "Int32", nil ),
                },
            ),
        ).
        expectDef(
            types.MakeSchemaDef( "ns1@v1/Schema3",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f1", "Int32", nil ),
                    types.MakeFieldDef( "f2", "Int32", nil ),
                },
            ),
        ).
        expectDef(
            types.MakeSchemaDef( "ns1@v1/Schema4",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f1", "Int32", nil ),
                    types.MakeFieldDef( "f2", "Int32", nil ),
                    types.MakeFieldDef( "f3", "Int32", nil ),
                },
            ),
        ).
        expectDef(
            types.MakeSchemaDef( "ns1@v1/Schema5",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f1", "Int32", nil ),
                    types.MakeFieldDef( "f2", "Int32", nil ),
                    types.MakeFieldDef( "f3", "Int32", nil ),
                },
            ),
        ).
        expectDef(
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v1/Schema6" ),
                AliasedType: mkTyp( "ns1@v1/Schema1" ),
            },
        ).
        expectDef(
            types.MakeSchemaDef( "ns1@v1/Schema7",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f1", "Int32", nil ),
                    types.MakeFieldDef( "f7", "Int32", nil ),
                },
            ),
        ).
        expectDef(
            types.MakeSchemaDef( "ns1@v1/Schema8",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f1", "Int32", nil ),
                    types.MakeFieldDef( "g1", "Int32", nil ),
                },
            ),
        ).
        expectDef( types.MakeSchemaDef( "ns1@v1/Schema9", nil ) ).
        expectDef(
            types.MakeSchemaDef( "ns1@v1/Schema10",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f1", "Int32", int32( 1 ) ),
                },
            ),
        ),

//        newCompilerTest( "standard-error-schema" ).
//        addSource( "f1", `
//            @version v1
//            namespace ns1
//            struct Error1 { @schema StandardError }
//        ` ),
    
        newCompilerTest( "alias-tests" ).
        addSource( "f1", `

            @version v1
            namespace ns1

            struct Struct1 {}
            
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
        ` ).
        expectDef(
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v1/Alias1" ),
                AliasedType: mkTyp( "mingle:core@v1/String?" ),
            },
        ).
        expectDef(
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v1/Alias2" ),
                AliasedType: mkTyp( "ns1@v1/Struct1" ),
            },
        ).
        expectDef(
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v1/Alias3" ),
                AliasedType: mkTyp( "mingle:core@v1/String?*" ),
            },
        ).
        expectDef(
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v1/Alias4" ),
                AliasedType: mkTyp( `mingle:core@v1/String~"^a+$"` ),
            },
        ).
        expectDef(
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v1/Alias5" ),
                AliasedType: mkTyp( `mingle:core@v1/Int64~[0,)` ),
            },
        ).
        expectDef( types.MakeStructDef( "ns1@v1/Struct1", nil ) ).
        expectDef(
            types.MakeStructDef(
                "ns1@v1/Struct5",
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
        ),

        newCompilerTest( "service-tests" ).
        addSource( "f1", `

            @version v1
            namespace ns1

            struct Struct1 {}

            struct Struct2 {}

            alias Alias1 String?

            alias Alias2 Struct1

            struct Exception1 {}

            struct Exception2 {}

            struct Exception3 {}
            
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
        ` ).
        expectDef(
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v1/Alias1" ),
                AliasedType: mkTyp( `mingle:core@v1/String?` ),
            },
        ).
        expectDef(
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v1/Alias2" ),
                AliasedType: mkTyp( `ns1@v1/Struct1` ),
            },
        ).
        expectDef( types.MakeStructDef( "ns1@v1/Struct1", nil ) ).
        expectDef( types.MakeStructDef( "ns1@v1/Struct2", nil ) ).
        expectDef( types.MakeStructDef( "ns1@v1/Exception1", nil ) ).
        expectDef( types.MakeStructDef( "ns1@v1/Exception2", nil ) ).
        expectDef( types.MakeStructDef( "ns1@v1/Exception3", nil ) ).
        expectDef(
            &types.PrototypeDefinition{
                Name: mkQn( "ns1@v1/Proto1" ),
                Signature: types.MakeCallSig(
                    []*types.FieldDefinition{},
                    "mingle:core@v1/String",
                    []string{},
                ),
            },
        ).
        expectDef(
            &types.PrototypeDefinition{
                Name: mkQn( "ns1@v1/Proto2" ),
                Signature: types.MakeCallSig(
                    []*types.FieldDefinition{},
                    "mingle:core@v1/String",
                    []string{ "ns1@v1/Exception1" },
                ),
            },
        ).
        expectDef(
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
        ).
        expectDef(
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
        ).
        expectDef(
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
        ).
        expectDef(
            types.MakeServiceDef(
                "ns1@v1/Service1",
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
        ).
        expectDef(
            types.MakeServiceDef(
                "ns1@v1/Service2",
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
        ).
        expectDef( types.MakeServiceDef( "ns1@v1/Service3", "ns1@v1/Sec2" ) ),

        newCompilerTest( "field-constant-tests" ).
        addSource( "f1", `
            @version v1
            namespace ns1

            enum Enum1 { red, green }
            
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
        ` ).
        expectDef(
            types.MakeStructDef(
                "ns1@v1/FieldConstantTester",
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
                        parser.MustEnum( "ns1@v1/Enum1", "green" ),
                    ),
                    types.MakeFieldDef( "f15", "mingle:core@v1/Timestamp",
                        parser.MustTimestamp( 
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
        ).
        expectDef( types.MakeEnumDef( "ns1@v1/Enum1", "red", "green" ) ),

        newCompilerTest( "import-tests" ).
        addSource( "f1", `
            @version v1
            namespace ns1

#            schema Schema1 { schemaField Int32 }

            struct Struct1 {}
            struct Struct2 {}
            struct Struct3 {}

            struct Exception1 {}
            struct Exception2 {}
            struct Exception3 {}

            alias Alias1 String?
            alias Alias2 Struct1
            alias Alias3 String?*

            service Service1 {}
        ` ).
        addSource( "f2", `
            @version v1
            namespace ns1
            
            struct Struct4 {
#                @schema Schema1
                inst1 Struct1
                inst2 Struct2
                str1 Alias1 # Implicitly brought in from compiler-src1.mg
                str2 Alias3
            }

            struct Exception4 {}

        ` ).
        addSource( "f3", `
            @version v1
            
            import ns1@v1/Struct4 # redundant but legal explicit version
            import ns1/Exception3
            
            namespace ns2@v1 # version is redundant but legal
            
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
            
            struct Exception1 {}

            struct Exception2 {}
#            struct Exception2 { @schema ns1@v1/Schema1 }
            
            service Service1 {
            
                op op1(): Struct2 throws Exception1, ns1/Exception4
            
                op op2( param1 String*+,
                        param2 Struct4*,
                        param3 ns1/Struct1,
                        param4 Struct2 ): String
            }
        ` ).
        addSource( "f4", `
            @version v1
            
            import ns1/* - [ 
#                Schema1, Exception1, Exception2, Struct1, Service1,
                Exception1, Exception2, Struct1, Service1,
            ]

            import ns2/Exception1 
            
            namespace ns1:globTestNs
            
            struct Struct1 {
#                @schema ns1/Schema1
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
        ` ).
        addSource( "f5", `
            @version v2
            
            namespace ns1

#            schema Schema1 { v2F1 String }
            
            struct Struct1 { f1 String default "hello" }
            
            struct Struct2 { f1 &Struct1? }

#            struct Struct2 {
#                @schema ns1@v1/Struct1
#                f1 &Struct1? 
#            }
            
            alias Struct1V1 ns1@v1/Struct1
            
            struct Struct3 { 
#                @schema Schema1
                f2 Struct2 
                f3 ns1@v1/Alias1
                f4 Struct1V1*
            }
 
            service Service1 {
                op op1( f1 Struct1, f2 Int64~[0,12), f3 ns1@v1/Struct1*+ ): Null
            }
        ` ).
//        expectDef(
//            types.MakeSchemaDef( 
//                "ns1@v1/Schema1",
//                []*types.FieldDefinition{
//                    types.MakeFieldDef( "schemaField", "Int32", nil ),
//                },
//            ),
//        ).
        expectDef( types.MakeStructDef( "ns1@v1/Struct1", nil ) ).
        expectDef( types.MakeStructDef( "ns1@v1/Struct2", nil ) ).
        expectDef( types.MakeStructDef( "ns1@v1/Struct3", nil ) ).
        expectDef( types.MakeStructDef( "ns1@v1/Exception1", nil ) ).
        expectDef( types.MakeStructDef( "ns1@v1/Exception2", nil ) ).
        expectDef( types.MakeStructDef( "ns1@v1/Exception3", nil ) ).
        expectDef( types.MakeStructDef( "ns1@v1/Exception4", nil ) ).
        expectDef(
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v1/Alias1" ),
                AliasedType: mkTyp( "mingle:core@v1/String?" ),
            },
        ).
        expectDef(
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v1/Alias2" ),
                AliasedType: mkTyp( "ns1@v1/Struct1" ),
            },
        ).
        expectDef(
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v1/Alias3" ),
                AliasedType: mkTyp( "mingle:core@v1/String?*" ),
            },
        ).
        expectDef( types.MakeServiceDef( "ns1@v1/Service1", "" ) ).
        expectDef(
            types.MakeStructDef(
                "ns1@v1/Struct4",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "inst1", "ns1@v1/Struct1", nil ),
                    types.MakeFieldDef( "inst2", "ns1@v1/Struct2", nil ),
                    types.MakeFieldDef( "str1", "mingle:core@v1/String?", nil ),
                    types.MakeFieldDef( 
                        "str2", "mingle:core@v1/String?*", nil ),
                },
            ),
        ).
        expectDef(
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns2@v1/Alias1" ),
                AliasedType: mkTyp( "ns1@v1/Struct3" ),
            },
        ).
        expectDef(
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns2@v1/Alias2" ),
                AliasedType: mkTyp( "mingle:core@v1/String?*" ),
            },
        ).
        expectDef(
            types.MakeStructDef(
                "ns2@v1/Struct1",
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
        ).
        expectDef(
            types.MakeStructDef(
                "ns2@v1/Struct2",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "inst1", "ns2@v1/Struct1", nil ),
                    types.MakeFieldDef( "inst2", "ns1@v1/Struct1", nil ),
                },
            ),
        ).
        expectDef( types.MakeStructDef( "ns2@v1/Exception1", nil ) ).
        expectDef(
            types.MakeStructDef(
                "ns2@v1/Exception2",
                []*types.FieldDefinition{
//                    types.MakeFieldDef( "str1", "mingle:core@v1/String*", nil ),
                },
            ),
        ).
        expectDef(
            types.MakeServiceDef(
                "ns2@v1/Service1",
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
        ).
        expectDef(
            types.MakeStructDef(
                "ns1:globTestNs@v1/Struct1",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "inst1", "ns1@v1/Struct2", nil ),
                    types.MakeFieldDef( "inst2", "ns1@v1/Struct4", nil ),
                    types.MakeFieldDef( 
                        "inst3", "ns1:globTestNs@v1/Struct1", nil ),
                },
            ),
        ).
        expectDef(
            types.MakeStructDef(
                "ns1:globTestNs@v1/Exception2",
                []*types.FieldDefinition{},
            ),
        ).
        expectDef(
            types.MakeServiceDef(
                "ns1:globTestNs@v1/Service1",
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
        ).
        expectDef(
            types.MakeStructDef(
                "ns1@v2/Struct1",
                []*types.FieldDefinition{
                    types.MakeFieldDef( 
                        "f1", "mingle:core@v1/String", "hello" ),
                },
            ),
        ).
        expectDef(
            types.MakeStructDef(
                "ns1@v2/Struct2",
//                "ns1@v1/Struct1",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f1", "&ns1@v2/Struct1?", nil ),
                },
            ),
        ).
        expectDef(
            &types.AliasedTypeDefinition{
                Name: mkQn( "ns1@v2/Struct1V1" ),
                AliasedType: mkTyp( "ns1@v1/Struct1" ),
            },
        ).
        expectDef(
            types.MakeStructDef(
                "ns1@v2/Struct3",
//                "ns1@v2/Struct1",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f2", "ns1@v2/Struct2", nil ),
                    types.MakeFieldDef( "f3", "mingle:core@v1/String?", nil ),
                    types.MakeFieldDef( "f4", "ns1@v1/Struct1*", nil ),
                },
            ),
        ).
        expectDef(
            types.MakeServiceDef(
                "ns1@v2/Service1",
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

        newCompilerTest( "import-include-exclude-success" ).
        addLib( "lib1", "@version v1; namespace lib1; struct S4 {}" ).
        addSource( "f1", `
            @version v1
            namespace ns1
            struct S1 {}
            struct S2 {}
            struct S3 {}
        ` ).
        expectDef( types.MakeStructDef( "ns1@v1/S1", nil ) ).
        expectDef( types.MakeStructDef( "ns1@v1/S2", nil ) ).
        expectDef( types.MakeStructDef( "ns1@v1/S3", nil ) ).
        addSource( "f2", `
            @version v1
            import ns1@v1/[ S1, S3 ]
            import lib1@v1/* - [ S4 ]
            namespace ns2
            struct T1 { f S1 }
            struct S2 {} # Okay since we don't import ns1@v1/S2
            struct S4 {} # Okay (no lib1@v1/S4)
        ` ).
        expectDef( 
            types.MakeStructDef( "ns2@v1/T1", 
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f", "ns1@v1/S1", nil ),
                },
            ),
        ).
        expectDef( types.MakeStructDef( "ns2@v1/S2", nil ) ).
        expectDef( types.MakeStructDef( "ns2@v1/S4", nil ) ).
        addSource( "f3", `
            @version v1
            import ns1@v1/* - [ S2 ]
            namespace ns3
            struct S2 {}
            struct T1 { f1 S1; f2 S3 }
        ` ).
        expectDef( types.MakeStructDef( "ns3@v1/S2", nil ) ).
        expectDef(
            types.MakeStructDef( "ns3@v1/T1",
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f1", "ns1@v1/S1", nil ),
                    types.MakeFieldDef( "f2", "ns1@v1/S3", nil ),
                },
            ),
        ),

        newCompilerTest( "import-include-exclude-fail" ).
        addSource( "f1", `
            @version v1
            namespace ns1
            struct S1 {}
            struct S2 {}
            struct S3 {}
        ` ).
        addLib( "lib1", "@version v1; namespace ns1C; struct S2 {}" ).
        addSource( "f2", `
            @version v1
            namespace ns1B
            struct S2 {}
        ` ).
        addSource( "f3", `
            @version v1
            import ns1/[ S4 ]
            import ns1/S2
            import ns1B/*
            import ns1C/*
            namespace ns2
            struct S1 {}
        ` ).
        expectSrcError( "f3", 3, 26, "No such import in ns1@v1: S4" ).
        expectSrcError( "f3", 5, 20, 
            "Importing S2 from ns1B@v1 would conflict with previous import " +
            "from ns1@v1" ).
        expectSrcError( "f3", 6, 20, 
            "Importing S2 from ns1C@v1 would conflict with previous import " +
            "from ns1@v1" ).
        addSource( "f4", `
            @version v1
            import ns1/* - [ S2 ]
            namespace ns3
            struct T1 { f1 S1; f2 S2 }
        ` ).
        expectSrcError( "f4", 5, 35, "Unresolved type: S2" ).
        addSource( "f5", 
            "@version v1; import ns1/* - [ S4 ]; namespace ns4; struct T1 {}" ).
        expectSrcError( "f5", 1, 31, "No such import in ns1@v1: S4" ).
        addSource( "f6", 
            "@version v1; import ns1/*; namespace ns4; struct S1 {}" ).
        expectSrcError( "f6", 1, 21,
            "Importing S1 from ns1@v1 would conflict with declared type in " +
            "ns4@v1" ),

        newCompilerTest( "core-type-implicit-resolution" ).
        setSource( `
            @version v1
            namespace ns1
            struct S1 {
                f1 Boolean
                f2 Buffer
                f3 String
                f4 Int32
                f5 Uint32
                f6 Int64
                f7 Uint64
                f8 Float32
                f9 Float64
                f10 Timestamp
                f11 SymbolMap
                f12 Int32*+
                f13 &Int32
            }
        `).
        expectDef(
            types.MakeStructDef( "ns1@v1/S1",
                []*types.FieldDefinition{
                    fldDef( "f1", "mingle:core@v1/Boolean", nil ),
                    fldDef( "f2", "mingle:core@v1/Buffer", nil ),
                    fldDef( "f3", "mingle:core@v1/String", nil ),
                    fldDef( "f4", "mingle:core@v1/Int32", nil ),
                    fldDef( "f5", "mingle:core@v1/Uint32", nil ),
                    fldDef( "f6", "mingle:core@v1/Int64", nil ),
                    fldDef( "f7", "mingle:core@v1/Uint64", nil ),
                    fldDef( "f8", "mingle:core@v1/Float32", nil ),
                    fldDef( "f9", "mingle:core@v1/Float64", nil ),
                    fldDef( "f10", "mingle:core@v1/Timestamp", nil ),
                    fldDef( "f11", "mingle:core@v1/SymbolMap", nil ),
                    fldDef( "f12", "mingle:core@v1/Int32*+", nil ),
                    fldDef( "f13", "&mingle:core@v1/Int32", nil ),
                },
            ),
        ),

        newCompilerTest( "nullable-type-handling" ).
        setSource( `
            @version v1
            namespace ns1
            struct S1 {}
            enum E1 { c1 }
            struct S2 {
                f1 Boolean?
                f2 Int32?
                f3 Uint32?
                f4 Int64?
                f5 Uint64?
                f6 Float32?
                f7 Float64?
                f8 Timestamp?
                f9 E1?
                f10 S1?
            }
            struct S3 {
                f1 String?
                f2 Buffer?
                f3 SymbolMap?
                f4 Int32*?
            }
        ` ).
        expectError( 7, 20, "not a nullable type" ).
        expectError( 8, 20, "not a nullable type" ).
        expectError( 9, 20, "not a nullable type" ).
        expectError( 10, 20, "not a nullable type" ).
        expectError( 11, 20, "not a nullable type" ).
        expectError( 12, 20, "not a nullable type" ).
        expectError( 13, 20, "not a nullable type" ).
        expectError( 14, 20, "not a nullable type" ).
        expectError( 15, 20, "not a nullable type" ).
        expectError( 16, 21, "not a nullable type" ).
        expectDef(
            types.MakeStructDef( "ns1@v1/S3",
                []*types.FieldDefinition{
                    fldDef( "f1", "mingle:core@v1/String?", nil ),
                    fldDef( "f2", "mingle:core@v1/Buffer?", nil ),
                    fldDef( "f3", "mingle:core@v1/SymbolMap?", nil ),
                    fldDef( "f4", "mingle:core@v1/Int32*?", nil ),
                },
            ),
        ),

        newCompilerTest( "valid-restrictions" ).
        setSource( `
            @version v1
            namespace ns1
            struct S1 {
                f1 String~"a"
                f2 String~["aaa", "bbb"]

                # We simultaneously permute primitive num types and interval
                # combinations with the next 4
                f3 Int32~( 0, 2 ]
                f4 Uint32~[ 0, 1 ]
                f5 Int64~[ 0, 2 )
                f6 Uint64~( 0, 2 )

                f7 Float32~[ 1, 2 )
                f8 Float64~[ 0.1, 2.1 )

                f9 Timestamp~[ "2012-01-01T12:00:00Z", "2012-01-02T12:00:00Z" ] 
            }
        ` ).
        expectDef(
            types.MakeStructDef( "ns1@v1/S1",
                []*types.FieldDefinition{
                    fldDef( "f1", `mingle:core@v1/String~"a"`, nil ),
                    fldDef( "f2", `mingle:core@v1/String~["aaa", "bbb"]`, nil ),
                    fldDef( "f3", `mingle:core@v1/Int32~(0,2]`, nil ),
                    fldDef( "f4", `mingle:core@v1/Uint32~[0,1]`, nil ),
                    fldDef( "f5", `mingle:core@v1/Int64~[0,2)`, nil ),
                    fldDef( "f6", `mingle:core@v1/Uint64~(0,2)`, nil ),
                    fldDef( "f7", `mingle:core@v1/Float32~[1,2)`, nil ),
                    fldDef( "f8", `mingle:core@v1/Float64~[0.1,2.1)`, nil ),
                    fldDef( 
                        "f9",
                        `mingle:core@v1/Timestamp~["2012-01-01T12:00:00Z","2012-01-02T12:00:00Z"]`,
                        nil,
                    ),
                
                },
            ),
        ),

        newCompilerTest( "invalid-restrictions" ).
        setSource( `
            @version v1
            namespace ns1
            struct S1 {}
            struct S2 {
                f1 S1~(,)
                f2 S1~"a"
                f3 String~[0, "1")
                f4 String~["0", 1)
                f5 Timestamp~(,1)
                f6 Int32~["a", 2)
                f7 Int32~(1, "20" )
                f8 Int32~"a"
                f9 Buffer~[0,1]
                f10 Timestamp~[ "2012-01-02T12:00:00Z", "2012-01-01T12:00:00Z" ]
                f11 Timestamp~["2001-0x-22",)
                f12 String~"ab[a-z"
                f13 Int32~[0,-1]
                f14 Uint32~(0,0)
                f15 Int64~[0,0)
                f16 Uint64~(0,0]
                f17 Int32~(0,1)
                f18 String~("a","a")
                f19 Timestamp~( "2012-01-01T12:00:00Z", "2012-01-01T12:00:00Z" )
                f20 Int32~[1.0,2]
                f21 Int32~[1,2.0]
                f22 Float32~(1.0,1.0)
                f23 Float64~(0.0,-1.0)
                f24 Int32~("1",3]
                f25 Int32~[0,"2")
                f26 String~["aab", "aaa"]
            }
        ` ).
        expectError( 6, 20, 
            "Invalid target type for range restriction: ns1@v1/S1" ).
        expectError( 7, 20, 
            "Invalid target type for regex restriction: ns1@v1/S1" ).
        expectError( 8, 28, "Got number as min value for range" ).
        expectError( 9, 33, "Got number as max value for range" ).
        expectError( 10, 32, "Got number as max value for range" ).
        expectError( 11, 27, "Got string as min value for range" ).
        expectError( 12, 30, "Got string as max value for range" ).
        expectError( 13, 20, 
            "Invalid target type for regex restriction: mingle:core@v1/Int32" ).
        expectError( 14, 20, 
            "Invalid target type for range restriction: mingle:core@v1/Buffer",
        ).
        expectError( 15, 31,"Unsatisfiable range" ).
        expectError( 16, 21, `Invalid RFC3339 time: "2001-0x-22"` ).
        expectError( 17, 28,
            "error parsing regexp: missing closing ]: `[a-z`" ).
        expectError( 18, 27, "Unsatisfiable range" ).
        expectError( 19, 28, "Unsatisfiable range" ).
        expectError( 20, 27, "Unsatisfiable range" ).
        expectError( 21, 28, "Unsatisfiable range" ).
        expectError( 22, 27, "Unsatisfiable range" ).
        expectError( 23, 28, "Unsatisfiable range" ).
        expectError( 24, 31, "Unsatisfiable range" ).
        expectError( 25, 28, "Got decimal as min value for range" ).
        expectError( 26, 30, "Got decimal as max value for range" ).
        expectError( 27, 29, "Unsatisfiable range" ).
        expectError( 28, 29, "Unsatisfiable range" ).
        expectError( 29, 28, "Got string as min value for range" ).
        expectError( 30, 30, "Got string as max value for range" ).
        expectError( 31, 28, "Unsatisfiable range" ),

        newCompilerTest( "schema-cycle-errors" ).
        setSource( `
            @version v1
            namespace ns1

            schema Schema1 { 
                f1 Int32 
                @schema Schema2
            }

            schema Schema2 { 
                f2 Int32
                @schema Schema1 
            }

            schema Schema3 {
                f3 Int32
                @schema Schema5
            }

            schema Schema4 {
                f4 Int32
                @schema Schema3
            }

            schema Schema5 {
                f5 Int32
                @schema Schema4
            }
        ` ).
        expectGlobalError(
            "Schemas are involved in one or more mixin cycles: ns1@v1/Schema1, ns1@v1/Schema2, ns1@v1/Schema3, ns1@v1/Schema4, ns1@v1/Schema5",
        ),

        newCompilerTest( "schema-errors" ).
        addLib( "lib1",
            `@version v1; namespace ns2; schema Schema1 { f1 Int32 }` ).
        setSource( `
            @version v1
            namespace ns1
            schema Schema1 {}
            schema Schema2 { @schema NoSuch }
            struct Struct1 { @schema NoSuch }
            schema Schema3 { f1 Int32 }
            schema Schema4 { @schema Schema3; f1 Int64 }
            schema Schema5 { @schema Schema3; f1 Int32 default 1 }
            schema Schema6 { @schema ns2@v1/Schema1; f1 Int64 }
            struct Struct2 { @schema ns2@v1/Schema1; f1 Int64 }
            schema Schema7 { f1 Int32 }
            struct Struct3 { f1 Int64; @schema Schema3; @schema Schema7 }
            struct Struct4 { @schema Struct1 }
            alias BadAlias1 Struct4
            struct Struct5 { @schema BadAlias1 }
            alias BadAlias2 String*
            struct Struct6 { @schema ns1@v1/BadAlias2 }
            schema Schema8 { f1 Int32 default 1 }
            struct Struct7 { f1 Int32; @schema Schema8 }
        ` ).
        expectError( 5, 38, "Unresolved type: NoSuch" ).
        expectError( 6, 38, "Unresolved type: NoSuch" ).
        expectError( 8, 13,
            "f1 declared at [<>, line 8, col 47] conflicts with other definitions" ).
        expectError( 8, 13, 
            "f1 mixed in from ns1@v1/Schema3 conflicts with other definitions" ).
        expectError( 9, 13,
            "f1 mixed in from ns1@v1/Schema3 conflicts with other definitions" ).
        expectError( 9, 13,
            "f1 declared at [<>, line 9, col 47] conflicts with other definitions" ).
        expectError( 10, 13,
            "f1 declared at [<>, line 10, col 54] conflicts with other definitions" ).
        expectError( 10, 13,
            "f1 mixed in from ns2@v1/Schema1 conflicts with other definitions" ).
        expectError( 11, 13,
            "f1 declared at [<>, line 11, col 54] conflicts with other definitions" ).
        expectError( 11, 13,
            "f1 mixed in from ns2@v1/Schema1 conflicts with other definitions" ).
        expectError( 13, 13,
            "f1 declared at [<>, line 13, col 30] conflicts with other definitions" ).
        expectError( 13, 13,
            "f1 mixed in from ns1@v1/Schema3 conflicts with other definitions" ).
        expectError( 13, 13,
            "f1 mixed in from ns1@v1/Schema7 conflicts with other definitions" ).
        expectError( 14, 38, "not a schema: Struct1" ).
        expectError( 16, 38, "not a schema: BadAlias1" ).
        expectError( 18, 38, "not a schema: ns1@v1/BadAlias2" ).
        expectError( 20, 13,
            "f1 declared at [<>, line 20, col 30] conflicts with other definitions" ).
        expectError( 20, 13,
            "f1 mixed in from ns1@v1/Schema8 conflicts with other definitions" ),

        newCompilerTest( "dup-decls-in-same-source" ).
        setSource( `
            @version v1
            namespace ns1
            struct S1 {}
            struct S1 {}
        ` ).
        expectError( 5, 13,
            "Type S1 is already declared in [<>, line 4, col 13]" ),

        newCompilerTest( "dup-decls-in-different-source" ).
        addSource( "f1", "@version v1; namespace ns1; struct S1 {}" ).
        addSource( "f2", "@version v1; namespace ns1; struct S1 {}" ).
        expectSrcError(
            "f2", 1, 29,
            "Type S1 is already declared in [f1, line 1, col 29]" ),
 
        newCompilerTest( "dup-decl-in-runtime-lib-and-src" ).
        addLib( "lib1", "@version v1; namespace ns1; struct S1 {}" ).
        addSource( "f1", "@version v1; namespace ns1; struct S1 {}" ).
        expectSrcError( 
            "f1", 1, 29, "Type S1 conflicts with an externally loaded type" ),
            
        newCompilerTest( "field-redefinition" ).
        setSource( `
            @version v1
            namespace ns1
            schema Schema1 { f String; f String }
            struct Struct { f String; f String }
            service S1 { op op1( f String, f String ): String; }
        ` ).
        expectError( 4, 30, "field 'f' is multiply-declared" ).
        expectError( 4, 40, "field 'f' is multiply-declared" ).
        expectError( 5, 29, "field 'f' is multiply-declared" ).
        expectError( 5, 39, "field 'f' is multiply-declared" ).
        expectError( 6, 34, "field 'f' is multiply-declared" ).
        expectError( 6, 44, "field 'f' is multiply-declared" ),
    
        newCompilerTest( "unresolved-field-type" ).
        setSource( "@version v1; namespace ns1; struct S1 { f1 Blah }" ).
        expectError( 1, 44, "Unresolved type: Blah" ),
 
        newCompilerTest( "unresolved-op-param-type" ).
        setSource( `
            @version v1
            namespace ns1
            service S1 { op op1( f B ): String; }
        ` ).
        expectError( 4, 36, "Unresolved type: B" ),
 
        newCompilerTest( "unresolved-op-ret-type" ).
        setSource( 
            "@version v1; namespace ns1; service S1 { op op1(): Blah; }" ).
        expectError( 1, 52, "Unresolved type: Blah" ),
    
        newCompilerTest( "unresolved-type-from-different-version" ).
        addSource( "f1", "@version v1; namespace ns1; struct T1 {}" ).
        addSource( "f2", `
            @version v2
            import ns1@v1/S1
            namespace ns1
            struct S1 { f ns1@v1/S1 }
            service S2 { op op1( f ns1@v1/S1 ): ns1@v1/S1; }
            prototype P1( f ns1@v1/S1 ): ns1@v1/S1
        ` ).
        expectSrcError( "f2", 3, 27, "No such import in ns1@v1: S1" ).
        expectSrcError( "f2", 5, 27, "Unresolved type: ns1@v1/S1" ).
        expectSrcError( "f2", 6, 36, "Unresolved type: ns1@v1/S1" ).
        expectSrcError( "f2", 6, 49, "Unresolved type: ns1@v1/S1" ).
        expectSrcError( "f2", 7, 29, "Unresolved type: ns1@v1/S1" ).
        expectSrcError( "f2", 7, 42, "Unresolved type: ns1@v1/S1" ),
 
        newCompilerTest( "bad-default-vals" ).
        setSource( `
            @version v1
            namespace ns1
            enum E1 { val1, val2 }
            struct S1 { 
                f1 String default -"a" 
                f2 String default 12
                f3 String? default true
                f4 String* default [ "a", 1, "c", false, 1.2 ]
                f5 Int32 default 1.1
                f6 Int32 default false
                f7 &Int32? default "2"
                f8 Int32+ default [ 1, 2, -3, "a", true, 1.2 ]
                f9 Float32 default true
                f10 &Float32? default "1.0"
                f11 Float32* default [ 1.1, true, "a", [] ]
                f12 Boolean default "true"
                f13 Boolean default 1
                f14 E1 default 1
                f15 E1 default "val1"
                f16 S1 default 12
                f17 S1 default true
                f18 S1 default "hi"
                f19 Timestamp default 1
                f20 Timestamp default true
                f21 String default []
                f22 E1 default val1 # bare enum val not ok
            }
        ` ).
        expectError( 6, 36,
            "Cannot negate values of type mingle:core@v1/String" ).
        expectError( 7, 35, "Expected mingle:core@v1/String but got number" ).
        expectError( 8, 36, "Expected mingle:core@v1/String? but got boolean" ).
        expectError( 9, 43, "Expected mingle:core@v1/String but got number" ).
        expectError( 9, 51, "Expected mingle:core@v1/String but got boolean" ).
        expectError( 9, 58, "Expected mingle:core@v1/String but got number" ).
        expectError( 10, 34, "Expected mingle:core@v1/Int32 but got float" ).
        expectError( 11, 34, "Expected mingle:core@v1/Int32 but got boolean" ).
        expectError( 12, 36, 
            "Expected &(mingle:core@v1/Int32)? but got string" ).
        expectError( 13, 47, "Expected mingle:core@v1/Int32 but got string" ).
        expectError( 13, 52, "Expected mingle:core@v1/Int32 but got boolean" ).
        expectError( 13, 58, "Expected mingle:core@v1/Int32 but got float" ).
        expectError( 14, 36,
            "Expected mingle:core@v1/Float32 but got boolean" ).
        expectError( 15, 39,
            "Expected &(mingle:core@v1/Float32)? but got string" ).
        expectError( 16, 45, 
            "Expected mingle:core@v1/Float32 but got boolean" ).
        expectError( 16, 51, "Expected mingle:core@v1/Float32 but got string" ).
        expectError( 16, 56, "List value not expected" ).
        expectError( 17, 37, "Expected mingle:core@v1/Boolean but got string" ).
        expectError( 18, 37, "Expected mingle:core@v1/Boolean but got number" ).
        expectError( 19, 32, "Expected ns1@v1/E1 but got number" ).
        expectError( 20, 32, "Expected ns1@v1/E1 but got string" ).
        expectError( 21, 32, "Expected ns1@v1/S1 but got number" ).
        expectError( 22, 32, "Expected ns1@v1/S1 but got boolean" ).
        expectError( 23, 32, "Expected ns1@v1/S1 but got string" ).
        expectError( 24, 39,
            "Expected mingle:core@v1/Timestamp but got number" ).
        expectError( 25, 39, 
            "Expected mingle:core@v1/Timestamp but got boolean" ).
        expectError( 26, 36, "List value not expected" ).
        expectError( 27, 32, "Found identifier in constant expression: val1" ),
    
        newCompilerTest( "invalid-timestamp-strings" ).
        setSource( `
            @version v1
            namespace ns1
            struct S1 { f1 Timestamp default "" }
            struct S2 { f1 Timestamp default "2001-01-02.12" }
        ` ).
        expectError( 4, 46, `Invalid RFC3339 time: ""` ).
        expectError( 5, 46, `Invalid RFC3339 time: "2001-01-02.12"` ),
 
        newCompilerTest( "redefined-op-name" ).
        setSource( `
            @version v1
            namespace ns1
            service S1 { op op1(): String; op op1( f1 String ): String; }
        ` ).
        expectError( 4, 47, "Operation already defined: op1" ),
    
        newCompilerTest( "null-field-type" ).
        setSource( `
            @version v1
            namespace ns1
            struct S1 { f1 Null; f2 Null?; }
            service Srv1 { op op1( f1 Null, f2 Null*, f3 Null? ): Null?; }
        ` ).
        expectError( 4, 28, "Null type not allowed here" ).
        expectError( 4, 37, "Non-atomic use of Null type" ).
        expectError( 4, 37, "Null type not allowed here" ).
        expectError( 5, 39, "Null type not allowed here" ).
        expectError( 5, 48, "Non-atomic use of Null type" ).
        expectError( 5, 48, "Null type not allowed here" ).
        expectError( 5, 58, "Non-atomic use of Null type" ).
        expectError( 5, 58, "Null type not allowed here" ).
        expectError( 5, 67, "Non-atomic use of Null type" ),
 
        newCompilerTest( "restriction-value-errors" ).
        setSource( `
            @version v1
            namespace ns
            struct S { 
                f1 String~"^a+$" default "bbb"
                f2 Int32~[8,9] default 12
            }
        ` ).
        expectError( 5, 42, 
            `Value "bbb" does not satisfy restriction "^a+$"` ).
        expectError( 6, 40,
            "Value 12 does not satisfy restriction [8,9]" ),

        newCompilerTest( "duplicate-enum-constants" ).
        setSource( 
            "@version v1; namespace ns; enum E1 { c1, c2, c2, c3, c1 }" ).
        expectError( 1, 46, "Duplicate definition of enum value: c2" ).
        expectError( 1, 54, "Duplicate definition of enum value: c1" ),
    
        newCompilerTest( "default-for-unbound-enum-type" ).
        setSource(
            "@version v1; namespace ns; struct S1 { f1 E2 default nope }" ).
        expectError( 1, 43, "Unresolved type: E2" ),
    
        newCompilerTest( "multi-sec-defs" ).
        setSource( `
            @version v1
            namespace ns
            service S1 { @security Sec1; @security Sec2; }
        ` ).
        expectError( 4, 13, 
            "Multiple security declarations are not supported" ),
    
        newCompilerTest( "invalid-sec-type" ).
        setSource( `
            @version v1
            namespace ns
            service S1 { @security String; }
            service S2 { @security NotDefined; }
        ` ).
        expectError( 
            4, 36, "Illegal @security type: mingle:core@v1/String" ).
        expectError( 5, 36, "Unresolved type: NotDefined" ),
    
        newCompilerTest( "sec-missing-authentication" ).
        setSource( `
            @version v1
            namespace ns
            prototype Sec(): String;
            service S1 { @security Sec; }
        ` ).
        expectError( 5, 36, "ns@v1/Sec has no authentication field" ),
 
        newCompilerTest( "sec-with-default-authentication" ).
        setSource( `
            @version v1
            namespace ns
            prototype Sec( authentication Int64 default 12 ): Null;
            service S1 { @security Sec; }
        ` ).
        expectError( 
            5, 36, "ns@v1/Sec supplies a default authentication value" ),

        newCompilerTest( "constructor-errors" ).
        setSource( `
            @version v1
            namespace ns
            struct S1 { @constructor( Blah ); }
            struct S2 { @constructor( String ); @constructor( String ); }
        ` ).
        expectError( 4, 39, "Unresolved type: Blah" ).
        expectError( 5, 49,
            "Duplicate constructor signature for type mingle:core@v1/String" ),
    
        newCompilerTest( "alias-errors" ).
        setSource( `
            @version v1
            namespace ns
            alias T1 String
            alias T2 Blah
            alias T1 String
            struct S1 {}
            alias T3 S1~"^.$"
            alias T4 String
            struct S2 { f T4 default [] }
            alias T5 String~"^a+$"
            struct S3 { f T5 default "b" }
            alias T6 Int64~[0,3]
            struct S4 { f T6 default -1 }
        ` ).
        expectError( 5, 22, "Unresolved type: Blah" ).
        expectError( 
            6, 13, "Type T1 is already declared in [<>, line 4, col 13]" ).
        expectError( 
            8, 22, "Invalid target type for regex restriction: ns@v1/S1" ).
        expectError( 10, 38, "List value not expected" ).
        expectError( 12, 38, `Value "b" does not satisfy restriction "^a+$"` ).
        expectError( 14, 38, `Value -1 does not satisfy restriction [0,3]` ),
 
        newCompilerTest( "circular-aliasing-within-ns" ).
        setSource( "@version v1; namespace ns1; alias A1 A2; alias A2 A1*;" ).
        expectError( 
            1, 38, "Alias cycle: ns1@v1/A1 --> ns1@v1/A2 --> ns1@v1/A1" ).
        expectError(
            1, 51, "Alias cycle: ns1@v1/A2 --> ns1@v1/A1 --> ns1@v1/A2" ),
    
        newCompilerTest( "multi-link-circular-aliasing" ).
        setSource( `
            @version v1 
            namespace ns1 
            alias A1 A2
            alias A2 A3*
            alias A3 A1+
        ` ).
        expectError( 4, 22,
            "Alias cycle: ns1@v1/A1 --> ns1@v1/A2 --> ns1@v1/A3 --> ns1@v1/A1",
        ).
        expectError( 5, 22,
            "Alias cycle: ns1@v1/A2 --> ns1@v1/A3 --> ns1@v1/A1 --> ns1@v1/A2",
        ).
        expectError( 6, 22,
            "Alias cycle: ns1@v1/A3 --> ns1@v1/A1 --> ns1@v1/A2 --> ns1@v1/A3",
        ),
    
        newCompilerTest( "circular-aliasing-across-namespaces" ).
        addSource( "f1", "@version v1; namespace ns1; alias A ns2/A;" ).
        addSource( "f2", "@version v1; namespace ns2; alias A ns3/A;" ).
        addSource( "f3", "@version v1; namespace ns3; alias A ns1/A;" ).
        expectSrcError( "f1", 1, 37,
            "Alias cycle: ns1@v1/A --> ns2@v1/A --> ns3@v1/A --> ns1@v1/A" ).
        expectSrcError( "f2", 1, 37,
            "Alias cycle: ns2@v1/A --> ns3@v1/A --> ns1@v1/A --> ns2@v1/A" ).
        expectSrcError( "f3", 1, 37,
            "Alias cycle: ns3@v1/A --> ns1@v1/A --> ns2@v1/A --> ns3@v1/A" ),
    }
    a := assert.NewPathAsserter( t )
    for _, test := range tests {
        test.PathAsserter = a.Descend( test.name )
        test.t = t
        test.call()
    }
}
