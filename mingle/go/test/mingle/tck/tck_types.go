package tck

import (
    mg "mingle"
    "mingle/types"
    "mingle/parser"
)

var (
    mkQn = parser.MustQualifiedTypeName
    asType = parser.AsTypeReference
)

// manually adding typedefs that would correspond to:
//
// ---------------------------------------------------
//
//  @version v1
//
//  namespace mingle:tck:data
//
//  struct EmptyStruct {}
//
//  struct ScalarsBasic {
//      stringF1 String
//      boolF1 Boolean1
//      bufferF1 Buffer
//      int32F1 Int32
//      int64F1 Int64
//      uint32F1 Uint32
//      uint64F1 Uint64
//      float32F1 Float32
//      float64F1 Float64
//      timeF1 Timestamp
//  }
//      
//  struct ScalarsRestrict {
//      stringF1 String~"^a+$"
//      stringF2 String~[ "aaa", "abb" ]
//      int32F1 Int32~( 0, 10 )
//      uint32F1 Uint32~( 0, 10 ]
//      int64F1 Int64~[ 0, 10 ]
//      uint64F1 UInt64~[ 0, 10 )
//      float32F1 Float32~( 0.0, 1.0 ]
//      float64F1 Float64~[ 0.0, 1.0 )
//      timeF1 Timestamp~[ "2013-10-19T00:00:00Z", "2014-10-19T00:00:00Z" )
//  }
//
//  enum Enum1 { const1, const2, const3 }
//
//  struct EnumHolder { enumF1 Enum1 }
//
//  struct MapHolder {
//      mapF1 SymbolMap
//      mapF2 SymbolMap?
//  }
//
//  union Union1 { Int32, ScalarsRestrict, &EnumHolder, &SymbolMap }
//
//  struct UnionHolder { union1F1 Union1 }
//
//  struct ValueHolder { valF1 Value }
//
//  struct ScalarFieldDefaults {
//      boolF1 Boolean default True
//      stringF1 String default "abc"
//      int32F1 Int32 default 1
//      uint32F1 Uint32 default 2
//      int64F1 Int64 default 3
//      uint64F1 Uint64 default 4
//      float32F1 Float32 default 5.0
//      float64F1 Float64 default 6.0
//      enum1F1 Enum1 default const2
//      timeF1 Timestamp default "2014-10-19T00:00:00Z"
//  }
//
//  struct Struct1 {
//      f1 Int32
//      f2 String
//  }
//
//  struct PointerStruct1 {
//      struct1F1 &&&Struct1
//      int32F1 &&&&int32
//  }
//
//  schema Schema1 { f1 Int32 }
//
//  struct Struct2 {
//      @mixin Schema1
//      f2 SymbolMap
//  }
//
//  struct Nullables {
//      boolF1 Boolean1?
//      bufferF1 Buffer?
//      int32F1 Int32?
//      int64F1 Int64?
//      uint32F1 Uint32?
//      uint64F1 Uint64?
//      float32F1 Float32?
//      float64F1 Float64?
//      timeF1 Timestamp?
//      stringF1 String?
//      mapF1 SymbolMap?
//      valF1 Value?
//      enum1PtrF1 &Enum1?
//      union1PtrF1 &Union1?
//      struct1F1 &Struct1?
//      schemaF1 Schema1?
//      int32PtrF1 &Int32?
//      int32ListF1 Int32*?
//  }
// 
//  struct Lists1 {
//      int32ListF1 Int32*
//      mapListF1 SymbolMap*?
//      union1ListF1 &Union1?*
//      schema1ListF1 Schema1*
//      struct1List1F1 Struct1*
//      enum1ListF1 Enum1+
//      int64PtrListF1 &Int64*
//      valueListF1 Value*
//      nullValueListF1 Value?*
//      valPtrListF1 &Value*
//      int32ListPtrF1 &( Int32* )
//      stringListListF1 String**
//  }
//
//  struct ListDefaults {
//
//      int32ListF1 Int32* [ -1, -2, -3 ]
//      int64ListF1 Int64* [ -6, -5, -4 ]
//      uint32ListF1 Uint32* [ 0, 10, 4294967295 ]
//      uint64ListF1 Uint64* [ 20, 30, 18446744073709551615 ]
//      float32ListF1 Float32* [ 0.0, -1.0 ]
//      float64ListF1 Float64* [ -2.0, 3.0 ]
//      stringListF1 String* [ "a", "b", "c" ]
//
//      timeListF1 Timestamp* [
//            "2014-10-19T00:00:00Z",
//            "2014-10-20T00:00:00Z",
//            "2014-10-21T00:00:00Z"
//      ]
//
//      enum1ListF1 Enum1* [ const1, const2, const1 ]
//  }
//
//  -------------------------------------------
//
//  @version v1
//
//  namespace mingle:tck:data2
//
//  struct Struct1 { f1 Int32 }
//
//  struct Struct2 {
//      f1 mingle:tck:data@v1/Struct1 
//      f2 mingle:tck:data2@v1/Struct1 
//  }
//
//  -------------------------------------------
//  
//  @version v1
//
//  namespace mingle:tck:service
//
//  struct S1 { f1 Int32 }
//
//  struct Err1 { f1 Int32 }
//
//  struct AuthErr1 { f1 Int32 }
//
//  prototype Auth1( authentication Int32 ): String,
//      throws AuthErr1
//
//  service Service1 {
//
//      op getFixedInt(): Int32
//
//      op echoS1( f1 S1 ): S1 throws Err1
//
//      op voidOp(): Null
//  }
//
//  service Service2 {
//  
//      @security Auth1
//
//      op getFixedInt(): Int32
//
//      op echoS1( f1 S1 ): S1 throws Err1
//  }
//
func addTckDefs( m *types.DefinitionMap ) {
    m.MustAdd( types.MakeStructDef( "mingle:tck:data@v1/EmptyStruct", nil ) )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:data@v1/ScalarsBasic",
            []*types.FieldDefinition{
                types.MakeFieldDef( "stringF1", "String", nil ),
                types.MakeFieldDef( "boolF1", "Boolean", nil ),
                types.MakeFieldDef( "bufferF1", "Buffer", nil ),
                types.MakeFieldDef( "int32F1", "Int32", nil ),
                types.MakeFieldDef( "int64F1", "Int64", nil ),
                types.MakeFieldDef( "uint32F1", "Uint32", nil ),
                types.MakeFieldDef( "uint64F1", "Uint64", nil ),
                types.MakeFieldDef( "float32F1", "Float32", nil ),
                types.MakeFieldDef( "float64F1", "Float64", nil ),
                types.MakeFieldDef( "timeF1", "Timestamp", nil ),
            },
        ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:data@v1/ScalarsRestrict",
            []*types.FieldDefinition{
                types.MakeFieldDef( "stringF1", `String~"^a+$"`, nil ),
                types.MakeFieldDef(
                    "stringF2", `String~[ "aaa", "abb" ]`, nil ),
                types.MakeFieldDef( "int32F1", "Int32~( 0, 10 )", nil ),
                types.MakeFieldDef( "uint32F1", "Uint32~( 0, 10 ]", nil ),
                types.MakeFieldDef( "int64F1", "Int64~[ 0, 10 ]", nil ),
                types.MakeFieldDef( "uint64F1", "Uint64~[ 0, 10 )", nil ),
                types.MakeFieldDef( "float32F1", "Float32~( 0.0, 1.0 ]", nil ),
                types.MakeFieldDef( "float64F1", "Float64~[ 0.0, 1.0 )", nil ),
                types.MakeFieldDef( 
                    "timeF1",
                    `Timestamp~[ "2013-10-19T00:00:00Z", ` +
                        `"2014-10-19T00:00:00Z" )`, 
                    nil,
                ),
            },
        ),
    )
    m.MustAdd(
        types.MakeEnumDef( "mingle:tck:data@v1/Enum1", 
            "const1", "const2", "const3" ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:data@v1/EnumHolder",
            []*types.FieldDefinition{
                types.MakeFieldDef( "enumF1", "mingle:tck:data@v1/Enum1", nil ),
            },
        ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:data@v1/MapHolder",
            []*types.FieldDefinition{
                types.MakeFieldDef( "mapF1", "SymbolMap", nil ),
                types.MakeFieldDef( "mapF2", "SymbolMap?", nil ),
            },
        ),
    )
    m.MustAdd(
        &types.UnionDefinition{
            Name: mkQn( "mingle:tck:data@v1/Union1" ),
            Union: types.MustUnionTypeDefinitionTypes(
                asType( "Int32" ),
                asType( "mingle:tck:data@v1/ScalarsRestrict" ),
                asType( "&mingle:tck:data@v1/EnumHolder" ),
                asType( "&SymbolMap" ),
            ),
        },
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:data@v1/UnionHolder",
            []*types.FieldDefinition{
                types.MakeFieldDef( 
                    "union1F1", "mingle:tck:data@v1/Union1", nil ),
            },
        ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:data@v1/ValueHolder",
            []*types.FieldDefinition{
                types.MakeFieldDef( "valF1", "Value", nil ),
            },
        ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:data@v1/ScalarFieldDefaults",
            []*types.FieldDefinition{
                types.MakeFieldDef( "boolF1", "Boolean", mg.Boolean( true ) ),
                types.MakeFieldDef( "stringF1", "String", mg.String( "abc" ) ),
                types.MakeFieldDef( "int32F1", "Int32", mg.Int32( 1 ) ),
                types.MakeFieldDef( "uint32F1", "Uint32", mg.Uint32( 2 ) ),
                types.MakeFieldDef( "int64F1", "Int64", mg.Int64( 3 ) ),
                types.MakeFieldDef( "uint64F1", "Uint64", mg.Uint64( 4 ) ),
                types.MakeFieldDef( "float32F1", "Float32", mg.Float32( 5.0 ) ),
                types.MakeFieldDef( "float64F1", "Float64", mg.Float64( 6.0 ) ),
                types.MakeFieldDef( 
                    "enum1F1",
                    "mingle:tck:data@v1/Enum1",
                    parser.MustEnum( "mingle:tck:data@v1/Enum1", "const2" ),
                ),
                types.MakeFieldDef(
                    "timeF1",
                    "Timestamp",
                    mg.MustTimestamp( "2014-10-19T00:00:00Z" ),
                ),
            },
        ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:data@v1/Struct1",
            []*types.FieldDefinition{
                types.MakeFieldDef( "f1", "Int32", nil ),
                types.MakeFieldDef( "f2", "String", nil ),
            },
        ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:data@v1/PointerStruct1",
            []*types.FieldDefinition{
                types.MakeFieldDef( 
                    "struct1F1", "&&&mingle:tck:data@v1/Struct1", nil ),
                types.MakeFieldDef( "int32F1", "&&&&Int32", nil ),
            },
        ),
    )
    m.MustAdd(
        types.MakeSchemaDef( "mingle:tck:data@v1/Schema1",
            []*types.FieldDefinition{
                types.MakeFieldDef( "f1", "Int32", nil ),
            },
        ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:data@v1/Struct2",
            []*types.FieldDefinition{
                types.MakeFieldDef( "f1", "Int32", nil ),
                types.MakeFieldDef( "f2", "String", nil ),
            },
        ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:data@v1/Nullables",
            []*types.FieldDefinition{
                types.MakeFieldDef( "boolF1", "Boolean1?", nil ),
                types.MakeFieldDef( "bufferF1", "Buffer?", nil ),
                types.MakeFieldDef( "int32F1", "Int32?", nil ),
                types.MakeFieldDef( "int64F1", "Int64?", nil ),
                types.MakeFieldDef( "uint32F1", "Uint32?", nil ),
                types.MakeFieldDef( "uint64F1", "Uint64?", nil ),
                types.MakeFieldDef( "float32F1", "Float32?", nil ),
                types.MakeFieldDef( "float64F1", "Float64?", nil ),
                types.MakeFieldDef( "timeF1", "Timestamp?", nil ),
                types.MakeFieldDef( "stringF1", "String?", nil ),
                types.MakeFieldDef( "mapF1", "SymbolMap?", nil ),
                types.MakeFieldDef( "valF1", "Value?", nil ),
                types.MakeFieldDef( 
                    "enum1PtrF1", "&mingle:tck:data@v1/Enum1?", nil ),
                types.MakeFieldDef(
                    "union1PtrF1", "&mingle:tck:data@v1/Union1?", nil ),
                types.MakeFieldDef(
                    "struct1F1", "&mingle:tck:data@v1/Struct1?", nil ),
                types.MakeFieldDef( 
                    "schemaF1", "&mingle:tck:data@v1/Schema1?", nil ),
                types.MakeFieldDef( "int32PtrF1", "&Int32?", nil ),
                types.MakeFieldDef( "int32ListF1", "Int32*?", nil ),
            },
        ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:data@v1/Lists1",
            []*types.FieldDefinition{
                types.MakeFieldDef( "int32ListF1", "Int32*", nil ),
                types.MakeFieldDef( "mapListF1", "SymbolMap*?", nil ),
                types.MakeFieldDef( 
                    "union1ListF1", "&mingle:tck:data@v1/Union1?*", nil ),
                types.MakeFieldDef(
                    "schema1ListF1", "mingle:tck:data@v1/Schema1*", nil ),
                types.MakeFieldDef(
                    "struct1List1F1", "mingle:tck:data@v1/Struct1*", nil ),
                types.MakeFieldDef(
                    "enum1ListF1", "mingle:tck:data@v1/Enum1+", nil ),
                types.MakeFieldDef( "int64PtrListF1", "&Int64*", nil ),
                types.MakeFieldDef( "valueListF1", "Value*", nil ),
                types.MakeFieldDef( "nullValueListF1", "Value?*", nil ),
                types.MakeFieldDef( "valPtrListF1", "&Value*", nil ),
                types.MakeFieldDef( "int32ListPtrF1", "&( Int32* )", nil ),
                types.MakeFieldDef( "stringListListF1", "String**", nil ),
            },
        ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:data@v1/ListDefaults",
            []*types.FieldDefinition{
                types.MakeFieldDef( 
                    "int32ListF1", 
                    "Int32*", 
                    mg.MustList(
                        asType( "Int32*" ), 
                        int32( -1 ),
                        int32( -2 ),
                        int32( -3 ),
                    ),
                ),
                types.MakeFieldDef(
                    "int64ListF1",
                    "Int64*",
                    mg.MustList(
                        asType( "Int64*" ),
                        int64( -6 ),
                        int64( -5 ),
                        int64( -4 ),
                    ),
                ),
                types.MakeFieldDef(
                    "unit32ListF1",
                    "Uint32*",
                    mg.MustList(
                        asType( "Uint32*" ),
                        uint32( 0 ),
                        uint32( 10 ),
                        uint32( 4294967295 ),
                    ),
                ),
                types.MakeFieldDef(
                    "uint64ListF1",
                    "Uint64*",
                    mg.MustList(
                        asType( "Uint64*" ),
                        uint64( 20 ),
                        uint64( 30 ), 
                        uint64( 18446744073709551615 ),
                    ),
                ),
                types.MakeFieldDef(
                    "float32ListF1",
                    "Float32*",
                    mg.MustList(
                        asType( "Float32*" ),
                        float32( 0.0 ),
                        float32( -1.0 ),
                    ),
                ),
                types.MakeFieldDef(
                    "float64ListF1",
                    "Float64*",
                    mg.MustList(
                        asType( "Float64*" ),
                        float64( -2.0 ),
                        float64( 3.0 ),
                    ),
                ),
                types.MakeFieldDef(
                    "stringListF1",
                    "String*",
                    mg.MustList( asType( "String*" ), "a", "b", "c" ),
                ),
                types.MakeFieldDef(
                    "timeListF1",
                    asType( "Timestamp*" ),
                    mg.MustList( 
                        asType( "Timestamp*" ),
                        mg.MustTimestamp( "2014-10-19T00:00:00Z" ),
                        mg.MustTimestamp( "2014-10-20T00:00:00Z" ),
                        mg.MustTimestamp( "2014-10-21T00:00:00Z" ),
                    ),
                ),
                types.MakeFieldDef(
                    "enum1ListF1",
                    "mingle:tck:data@v1/Enum1*",
                    mg.MustList(
                        asType( "mingle:tck:data@v1/Enum1*" ),
                        parser.MustEnum( "mingle:tck:data@v1/Enum1", "const1" ),
                        parser.MustEnum( "mingle:tck:data@v1/Enum1", "const2" ),
                        parser.MustEnum( "mingle:tck:data@v1/Enum1", "const1" ),
                    ),
                ),
            },
        ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:data2@v1/Struct1",
            []*types.FieldDefinition{
                types.MakeFieldDef( "f1", "Int32", nil ),
            },
        ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:data2@v1/Struct2",
            []*types.FieldDefinition{
                types.MakeFieldDef( "f1", "mingle:tck:data@v1/Struct1", nil ),
                types.MakeFieldDef( "f2", "mingle:tck:data2@v1/Struct1", nil ),
            },
        ),
    )
    m.MustAdd(
        types.MakeServiceDef( "mingle:tck:service@v1/Service1", "",
            types.MakeOpDef( "getFixedInt",
                types.MakeCallSig( 
                    []*types.FieldDefinition{}, 
                    "Int32", 
                    []string{},
                ),
            ),
            types.MakeOpDef( "echoS1",
                types.MakeCallSig(
                    []*types.FieldDefinition{ 
                        types.MakeFieldDef( 
                            "f1", "mingle:tck:service@v1/S1", nil ),
                    },
                    "mingle:tck:service@v1/S1",
                    []string{ "mingle:tck:service@v1/Err1" },
                ),
            ),
            types.MakeOpDef( "voidOp",
                types.MakeCallSig(
                    []*types.FieldDefinition{},
                    "Null",
                    []string{},
                ),
            ),
        ),
    )
    m.MustAdd(
        types.MakeServiceDef( 
            "mingle:tck:service@v1/Service2", 
            "mingle:tck:service@v1/Auth1",
            types.MakeOpDef( "getFixedInt",
                types.MakeCallSig(
                    []*types.FieldDefinition{},
                    "Int32",
                    []string{},
                ),
            ),
            types.MakeOpDef( "echoS1",
                types.MakeCallSig(
                    []*types.FieldDefinition{
                        types.MakeFieldDef( 
                            "f1", "mingle:tck:service@v1/S1", nil ),
                    },
                    "mingle:tck:service@v1/S1",
                    []string{ "mingle:tck:service@v1/Err1" },
                ),
            ),
        ),
    )
    m.MustAdd(
        &types.PrototypeDefinition{
            Name: parser.MustQualifiedTypeName( "mingle:tck:service@v1/Auth1" ),
            Signature: types.MakeCallSig(
                []*types.FieldDefinition{
                    types.MakeFieldDef( "authentication", "Int32", nil ),
                },
                "String",
                []string{ "mingle:tck:service@v1/AuthErr1" },
            ),
        },
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:service@v1/S1",
            []*types.FieldDefinition{
                types.MakeFieldDef( "f1", "Int32", nil ),
            },
        ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:service@v1/Err1",
            []*types.FieldDefinition{
                types.MakeFieldDef( "f1", "Int32", nil ),
            },
        ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:service@v1/Err2",
            []*types.FieldDefinition{
                types.MakeFieldDef( "f1", "Int32", nil ),
            },
        ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck:service@v1/AuthErr1",
            []*types.FieldDefinition{
                types.MakeFieldDef( "f1", "Int32", nil ),
            },
        ),
    )
}

func GetDefinitions() *types.DefinitionMap {
    res := types.NewDefinitionMap()
    addTckDefs( res )
    return res
}
