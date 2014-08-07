package service

import (
    mg "mingle"
    "mingle/parser"
    "mingle/types"
    "bitgirder/objpath"
)

const (
    ReactorProfileBase = "base"
    ReactorProfileTyped = "typed"
    ErrorProfileImpl = "impl-error"
)

type ReactorTest struct {
    Type *mg.QualifiedTypeName
    Expect interface{}
    Error error
    In interface{}
    ReactorProfile string
    ErrorProfile string
}

type responseInput struct {
    in mg.Value
    reqCtx *RequestContext
}

type requestExpect struct {
    ctx *RequestContext
    params mg.Value
    auth mg.Value
}

type responseExpect struct {
    result mg.Value
    err mg.Value
}

type testError struct {
    path objpath.PathNode
    msg string
}

func ( t *testError ) Error() string { return mg.FormatError( t.path, t.msg ) }

// manually creating typedefs that would correspond to:
//
//  @version v1
//
//  namespace ns1
//
//  struct S1 {
//      f1 Int32
//  }
//
//  struct Err1 {
//      f1 Int32
//  }
//
//  struct AuthErr1 {
//      f1 Int32
//  }
//
//  prototype Auth1( authentication Int32 ): String,
//      throws AuthErr1
//
//  service Service1 {
//
//      op op1(): Int32
//
//      op op2( f1 S1 ): S1 throws Err1
//  }
//
//  service Service2 {
//  
//      @security Auth1
//
//      op op1(): Null
//
//      op op2( f1 S1 ): S1 throws Err1
//  }
//
var testTypeDefs = types.MakeV1DefMap(
    types.MakeServiceDef( "ns1@v1/Service1", "",
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
                    types.MakeFieldDef( "f1", "ns1@v1/S1", nil ),
                },
                "ns1@v1/S1",
                []string{ "ns1@v1/Err1" },
            ),
        ),
    ),
    types.MakeServiceDef( "ns1@v1/Service2", "ns1@v1/Auth1",
        types.MakeOpDef( "op1",
            types.MakeCallSig(
                []*types.FieldDefinition{},
                "Null",
                []string{},
            ),
        ),
        types.MakeOpDef( "op2",
            types.MakeCallSig(
                []*types.FieldDefinition{
                    types.MakeFieldDef( "f1", "ns1@v1/S1", nil ),
                },
                "ns1@v1/S1",
                []string{ "ns1@v1/Err1" },
            ),
        ),
    ),
    &types.PrototypeDefinition{
        Name: parser.MustQualifiedTypeName( "ns1@v1/Auth1" ),
        Signature: types.MakeCallSig(
            []*types.FieldDefinition{
                types.MakeFieldDef( "authentication", "Int32", nil ),
            },
            "String",
            []string{ "ns1@v1/AuthErr1" },
        ),
    },
    types.MakeStructDef( "ns1@v1/S1",
        []*types.FieldDefinition{
            types.MakeFieldDef( "f1", "Int32", nil ),
        },
    ),
    types.MakeStructDef( "ns1@v1/Err1",
        []*types.FieldDefinition{
            types.MakeFieldDef( "f1", "Int32", nil ),
        },
    ),
    types.MakeStructDef( "ns1@v1/AuthErr1",
        []*types.FieldDefinition{
            types.MakeFieldDef( "f1", "Int32", nil ),
        },
    ),
)
