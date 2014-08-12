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
//  namespace mingle:tck
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
//      op op1(): Int32
//
//      op op2( f1 S1 ): S1 throws Err1
//
//      op op3(): Null
//  }
//
//  service Service2 {
//  
//      @security Auth1
//
//      op op1(): Int32
//
//      op op2( f1 S1 ): S1 throws Err1
//  }
//
func addTckDefs( m *types.DefinitionMap ) {
    m.MustAdd(
        types.MakeServiceDef( "mingle:tck@v1/Service1", "",
            types.MakeOpDef( "op1",
                types.MakeCallSig( 
                    []*types.FieldDefinition{}, 
                    "Int32", 
                    []string{},
                ),
            ),
            types.MakeOpDef( "op2",
                types.MakeCallSig(
                    []*types.FieldDefinition{ 
                        types.MakeFieldDef( "f1", "mingle:tck@v1/S1", nil ),
                    },
                    "mingle:tck@v1/S1",
                    []string{ "mingle:tck@v1/Err1" },
                ),
            ),
            types.MakeOpDef( "op3",
                types.MakeCallSig(
                    []*types.FieldDefinition{},
                    "Null",
                    []string{},
                ),
            ),
        ),
    )
    m.MustAdd(
        types.MakeServiceDef( "mingle:tck@v1/Service2", "mingle:tck@v1/Auth1",
            types.MakeOpDef( "op1",
                types.MakeCallSig(
                    []*types.FieldDefinition{},
                    "Int32",
                    []string{},
                ),
            ),
            types.MakeOpDef( "op2",
                types.MakeCallSig(
                    []*types.FieldDefinition{
                        types.MakeFieldDef( "f1", "mingle:tck@v1/S1", nil ),
                    },
                    "mingle:tck@v1/S1",
                    []string{ "mingle:tck@v1/Err1" },
                ),
            ),
        ),
    )
    m.MustAdd(
        &types.PrototypeDefinition{
            Name: parser.MustQualifiedTypeName( "mingle:tck@v1/Auth1" ),
            Signature: types.MakeCallSig(
                []*types.FieldDefinition{
                    types.MakeFieldDef( "authentication", "Int32", nil ),
                },
                "String",
                []string{ "mingle:tck@v1/AuthErr1" },
            ),
        },
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck@v1/S1",
            []*types.FieldDefinition{
                types.MakeFieldDef( "f1", "Int32", nil ),
            },
        ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck@v1/Err1",
            []*types.FieldDefinition{
                types.MakeFieldDef( "f1", "Int32", nil ),
            },
        ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck@v1/Err2",
            []*types.FieldDefinition{
                types.MakeFieldDef( "f1", "Int32", nil ),
            },
        ),
    )
    m.MustAdd(
        types.MakeStructDef( "mingle:tck@v1/AuthErr1",
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
