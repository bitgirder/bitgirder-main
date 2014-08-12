package service

import (
    mg "mingle"
    "mingle/parser"
    "mingle/types"
    "mingle/tck"
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

// adds to mingle:tck@v1 types those corresponding to:
//
// ---------------------------------------------------
//
//  @version v1
//
//  namespace mingle:service:fail
//
//  service Service1 { op1() Null }
//
// ---------------------------------------------------
//
//  @version v1
//
//  namespace mingle:service
//
//  struct Err1 { f1 Int32 }
//
//  service Service1 {
//
//      failStartAuthentication() Null
//
//      failStartParameters() Null
//
//      failResponse() Int32 throws Err1
//  }
//
// ---------------------------------------------------
//
// Note that this is a method and not a package-level 'var' since we want it to
// include the builtin types via types.MakeV1DefMap(), some of which are added
// in this package's init, making them unavailable to a package-level var
// initializer
func getTestTypeDefs() *types.DefinitionMap {
    res := types.MakeV1DefMap(
        types.MakeServiceDef( "mingle:service:fail@v1/Service1", "",
            types.MakeOpDef( "op1",
                types.MakeCallSig( 
                    []*types.FieldDefinition{}, "Null", []string{} ),
            ),
        ),
        types.MakeServiceDef( 
            "mingle:service@v1/Service1", 
            "mingle:service@v1/Auth1",
            types.MakeOpDef( "failStartAuthentication",
                types.MakeCallSig( 
                    []*types.FieldDefinition{}, "Null", []string{} ),
            ),
            types.MakeOpDef( "failStartParameters",
                types.MakeCallSig( 
                    []*types.FieldDefinition{}, "Null", []string{} ),
            ),
            types.MakeOpDef( "failResponse",
                types.MakeCallSig( 
                    []*types.FieldDefinition{}, 
                    "Null", 
                    []string{ "mingle:service@v1/Err1" },
                ),
            ),
        ),
        types.MakeStructDef( "mingle:service@v1/Err1",
            []*types.FieldDefinition{
                types.MakeFieldDef( "f1", "Int32", nil ),
            },
        ),
        &types.PrototypeDefinition{
            Name: parser.MustQualifiedTypeName( "mingle:service@v1/Auth1" ),
            Signature: types.MakeCallSig(
                []*types.FieldDefinition{
                    types.MakeFieldDef( "authentication", "Int32", nil ),
                },
                "String",
                []string{ "ns1@v1/AuthErr1" },
            ),
        },
    )
    res.MustAddFrom( tck.GetDefinitions() )
    return res
}
