package compiler

import (
    "testing"
    "bitgirder/assert"
    "mingle/types"
)

func TestCompiler( t *testing.T ) {
    tests := []*compilerTest{

        newCompilerTest( "import-include-exclude-success" ).
        addLib( "lib1", "@version v1; namespace lib1; struct S4 {}" ).
        addSource( "f1", `
            @version v1
            namespace ns1
            struct S1 {}
            struct S2 {}
            struct S3 {}
        ` ).
        expectDef( types.MakeStructDef( "ns1@v1/S1", "", nil ) ).
        expectDef( types.MakeStructDef( "ns1@v1/S2", "", nil ) ).
        expectDef( types.MakeStructDef( "ns1@v1/S3", "", nil ) ).
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
            types.MakeStructDef( "ns2@v1/T1", "", 
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f", "ns1@v1/S1", nil ),
                },
            ),
        ).
        expectDef( types.MakeStructDef( "ns2@v1/S2", "", nil ) ).
        addSource( "f3", `
            @version v1
            import ns1@v1/* - [ S2 ]
            namespace ns3
            struct S2 {}
            struct T1 { f1 S1; f2 S3 }
        ` ).
        expectDef( types.MakeStructDef( "ns3@v1/S2", "", nil ) ).
        expectDef(
            types.MakeStructDef( "ns3@v1/T1", "",
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
            types.MakeStructDef( "ns1@v1/S1", "",
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
                f5 Uint64
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
            types.MakeStructDef( "ns1@v1/S3", "",
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
            types.MakeStructDef( "ns1@v1/S1", "",
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
            }
        ` ).
        expectError( 6, 1, "Invalid target type for range restriction" ).
        expectError( 7, 1, "Invalid target type for regex restriction" ).
        expectError( 8, 1, "Got number as min value for range" ).
        expectError( 9, 1, "Got number as max value for range" ).
        expectError( 10, 1, "Got number as max value for range" ).
        expectError( 11, 1, "Got string as min value for range" ).
        expectError( 12, 1, "Got string as max value for range" ).
        expectError( 13, 1, "Invalid target type for regex restriction" ).
        expectError( 14, 1, "Invalid target type for range restriction" ).
        expectError( 15, 1,"Unsatisfiable range" ).
        expectError( 16, 1,"Invalid min value in range restriction: val: Invalid timestamp: [<input>, line 1, col 1]: Invalid RFC3339 time: \"2001-0x-22\"" ).
        expectError( 17, 1,`error parsing regexp: missing closing ]: "[a-z"` ).
        expectError( 18, 1, "Unsatisfiable range" ).
        expectError( 19, 1, "Unsatisfiable range" ).
        expectError( 20, 1, "Unsatisfiable range" ).
        expectError( 21, 1, "Unsatisfiable range" ).
        expectError( 22, 1, "Unsatisfiable range" ).
        expectError( 23, 1, "Unsatisfiable range" ).
        expectError( 24, 1, "Unsatisfiable range" ).
        expectError( 25, 1, "Got decimal as min value for range" ).
        expectError( 26, 1, "Got decimal as max value for range" ).
        expectError( 27, 1, "Unsatisfiable range" ).
        expectError( 28, 1, "Unsatisfiable range" ).
        expectError( 29, 1, "Got string as min value for range" ).
        expectError( 30, 1, "Got string as max value for range" ),

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
        addLib( "p1Src1", p1Sources[ 0 ] ).
        addSource( "f1", `
            @version v1
            namespace ns2
            struct S1 { f1 String }
            struct S2 < S1 { f1 Int64 }
            struct S3 < S1 { f2 Int64 } # This is okay
            struct S4 < S3 { f1 Int64 } # But this is not
            struct S5 < ns1/Struct1 { string1 String }
            struct S6 { f1 Int64; f1 String }
        ` ).
        addSource( "f2",
            "@version v1; namespace ns3; struct S1 < ns2/S1 { f1 String }" ).
        expectSrcError(
            "f1", 9, 35, "Field f1 already defined at [f1, line 9, col 25]" ).
        expectSrcError( "f1", 5, 30, "Field f1 already defined in ns2@v1/S1" ).
        expectSrcError( "f1", 7, 30, "Field f1 already defined in ns2@v1/S1" ).
        expectSrcError( 
            "f1", 8, 39, "Field string1 already defined in ns1@v1/Struct1" ).
        expectSrcError( "f2", 1, 50, "Field f1 already defined in ns2@v1/S1" ),
            
        newCompilerTest( "op-field-redefinition" ).
        setSource( `
            @version v1
            namespace ns1
            service S1 { op op1( f String, f String ): String; }
        ` ).
        expectError( 4, 44, "Field f already defined at [<>, line 4, col 34]" ),
    
        newCompilerTest( "unresolved-field-type" ).
        setSource( "@version v1; namespace ns1; struct S1 { f1 Blah }" ).
        expectError( 1, 44, "Unresolved type: Blah" ),

        newCompilerTest( "unresolved-super-type" ).
        setSource( "@version v1; namespace ns1; struct S1 < nsX/Y {}" ).
        expectError( 1, 41, "Unresolved type: nsX@v1/Y" ),
 
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
        expectError( 12, 36, "Expected &mingle:core@v1/Int32? but got string" ).
        expectError( 13, 47, "Expected mingle:core@v1/Int32 but got string" ).
        expectError( 13, 52, "Expected mingle:core@v1/Int32 but got boolean" ).
        expectError( 13, 58, "Expected mingle:core@v1/Int32 but got float" ).
        expectError( 14, 36,
            "Expected mingle:core@v1/Float32 but got boolean" ).
        expectError( 15, 39,
            "Expected &mingle:core@v1/Float32? but got string" ).
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
 
        newCompilerTest( "invalid-supertypes" ).
        setSource( `
            @version v1 
            namespace ns1 
            struct S1 < String {}
            struct S2 {}
            struct S3 < S2+ {}
        ` ).
        expectError( 
            4, 13, "S1 cannot descend from type mingle:core@v1/String" ).
        expectError( 6, 25, "Non-atomic supertype for S3: ns1@v1/S2+" ),
    
        newCompilerTest( "type-self-descent" ).
        setSource( "@version v1; namespace ns1; struct S1 < S1 {}" ).
        expectError( 1, 29,
            "Type ns1@v1/S1 is involved in one or more circular " +
            "dependencies" ),
 
        newCompilerTest( "ancestral-self-descent" ).
        setSource( `
            @version v1 
            namespace ns1
            struct S1 < S3 {} 
            struct S2 < S1 {} 
            struct S3 < S3 {}
        ` ).
        expectError( 4, 13,
            "Type ns1@v1/S1 is involved in one or more circular dependencies" ).
        expectError( 5, 13,
            "Type ns1@v1/S2 is involved in one or more circular dependencies" ).
        expectError( 6, 13,
            "Type ns1@v1/S3 is involved in one or more circular dependencies" ),
 
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
    
        newCompilerTest( "restriction-syntax-errors" ).
        setSource( `
            @version v1
            namespace ns
            struct S { 
                f1 String~"a[bc"
                f2 String~[1,2] 
                f3 Int32~"a*"
                f4 Int32~[12,11)
            }
        ` ).
        expectError( 5, 20, "error parsing regexp: missing closing ]: `[bc`" ).
        expectError( 6, 20, "Got number as min value for range" ).
        expectError( 7, 20,
            "Invalid target type for regex restriction: mingle:core@v1/Int32" ).
        expectError( 8, 20, "Unsatisfiable range" ),
 
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
        expectError( 4, 26, "Multiple definitions of @security" ).
        expectError( 4, 42, "Multiple definitions of @security" ),
    
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

        newCompilerTest( "mislocated-keyed-elts" ).
        setSource( `
            @version v1
            namespace ns
            service Service1 { @constructor( String ); }
            struct Struct1 { @security Sec1; }
        ` ).
        expectError( 4, 32, "Unexpected declaration: @constructor" ).
        expectError( 5, 30, "Unexpected declaration: @security" ),
    
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
