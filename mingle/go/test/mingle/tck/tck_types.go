package tck

import (
    "mingle/types"
    "mingle/parser"
)

// manually adding typedefs that would correspond to:
//
// ---------------------------------------------------
//
//  @version v1
//
//  namespace mingle:tck:data
//
//  struct ScalarsBasic {
//      stringF1 String
//      bool1 Boolean1
//      buffer1 Buffer
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
//  struct EnumHolder { enum1 Enum1 }
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
//      enum1F1 Enum1 default const2
//      timeF1 Timestamp default "2014-10-19T00:00:00Z"
//  }
//
//  struct Struct1 {
//      f1 Int32
//      f2 String
//  }
//
//  schema Shema1 { f1 Int32 }
//
//  struct Struct2 {
//      @mixin Schema1
//      f2 SymbolMap
//  }
//
//  struct Nullables {
//      mapF1 SymbolMap?
//      valF1 Value?
//      stringF1 String?
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
