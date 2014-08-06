package service

import (
    "bitgirder/objpath"
    mgRct "mingle/reactor"
    mg "mingle"
    "mingle/parser"
//    "log"
)

type reactorTestBuilder struct {
    b *mgRct.ReactorTestSetBuilder
    typ *mg.QualifiedTypeName
    profile string
}

func ( b *reactorTestBuilder ) withProfile( 
    profile string ) *reactorTestBuilder {

    return &reactorTestBuilder{ b: b.b, typ: b.typ, profile: profile }
}

func ( b *reactorTestBuilder ) add( t *ReactorTest ) {
    t.Type = b.typ
    t.Profile = b.profile
    b.b.AddTests( t )
}

func ( b *reactorTestBuilder ) addOk( in mg.Value, expct interface{} ) {
    b.add( &ReactorTest{ In: in, Expect: expct } )
}

func ( b *reactorTestBuilder ) addErr( in mg.Value, err error ) {
    b.add( &ReactorTest{ In: in, Error: err } )
}

func initBaseRequestTests( tsb *mgRct.ReactorTestSetBuilder ) {
    b := &reactorTestBuilder{ b: tsb, typ: QnameRequest, profile: ProfileBase }
    b.addOk(
        parser.MustStruct( QnameRequest,
            "namespace", "ns1@v1",
            "service", "svc1",
            "operation", "op1",
            "parameters", mg.EmptySymbolMap(),
            "authentication", int32( 1 ),
        ),
        &requestExpect{
            ctx: &RequestContext{
                Namespace: mkNs( "ns1@v1" ),
                Service: mkId( "svc1" ),
                Operation: mkId( "op1" ),
            },
            params: mg.EmptySymbolMap(),
            auth: mg.Int32( 1 ),
        },
    )
    b.addOk(
        parser.MustSymbolMap(
            "namespace", "ns1@v1",
            "service", "svc1",
            "operation", "op1",
            "parameters", mg.EmptySymbolMap(),
            "authentication", int32( 1 ),
        ),
        &requestExpect{
            ctx: &RequestContext{
                Namespace: mkNs( "ns1@v1" ),
                Service: mkId( "svc1" ),
                Operation: mkId( "op1" ),
            },
            params: mg.EmptySymbolMap(),
            auth: mg.Int32( 1 ),
        },
    )
    b.addOk(
        parser.MustStruct( QnameRequest,
            "namespace", "ns1@v1",
            "service", "svc1",
            "operation", "op1",
        ),
        &requestExpect{
            ctx: &RequestContext{
                Namespace: mkNs( "ns1@v1" ),
                Service: mkId( "svc1" ),
                Operation: mkId( "op1" ),
            },
            params: mg.EmptySymbolMap(),
        },
    )
    b.addOk(
        parser.MustSymbolMap(
            "namespace", "ns1@v1",
            "service", "svc1",
            "operation", "op1",
            "parameters", parser.MustSymbolMap( "f1", int32( 1 ) ),
        ),
        &requestExpect{
            ctx: &RequestContext{
                Namespace: mkNs( "ns1@v1" ),
                Service: mkId( "svc1" ),
                Operation: mkId( "op1" ),
            },
            params: parser.MustSymbolMap( "f1", int32( 1 ) ),
        },
    )
    b.addOk(
        parser.MustSymbolMap(
            "namespace", "ns1@v1",
            "service", "svc1",
            "operation", "op1",
            "parameters", parser.MustSymbolMap( 
                "f1", parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
            ),
        ),
        &requestExpect{
            ctx: &RequestContext{
                Namespace: mkNs( "ns1@v1" ),
                Service: mkId( "svc1" ),
                Operation: mkId( "op1" ),
            },
            params: parser.MustSymbolMap( 
                "f1", parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
            ),
        },
    )
    typReq := QnameRequest.AsAtomicType()
    b.addErr(
        mg.Int32( 1 ),
        mg.NewTypeCastError( typReq, mg.TypeInt32, nil ),
    )
    b.addErr(
        parser.MustStruct( "ns1@v1/Bad" ),
        mg.NewTypeCastError( typReq, asType( "ns1@v1/Bad" ), nil ),
    )
    b.addErr(
        parser.MustStruct( QnameRequest, "namespace", "Bad" ),
        mg.NewValueCastError( 
            objpath.RootedAt( mkId( "namespace" ) ), 
            "[<input>, line 1, col 1]: Illegal start of identifier part: \"B\" (U+0042)",
        ),
    )
    b.addErr(
        parser.MustStruct( QnameRequest, 
            "namespace", "ns1@v1",
            "service", "bad$id",
        ),
        mg.NewValueCastError( 
            objpath.RootedAt( mkId( "service" ) ), 
            "[<input>, line 1, col 4]: Invalid id rune: \"$\" (U+0024)",
        ),
    )
    b.addErr(
        parser.MustStruct( QnameRequest, 
            "namespace", "ns1@v1",
            "service", "svc1",
            "operation", "bad$id",
        ),
        mg.NewValueCastError( 
            objpath.RootedAt( mkId( "operation" ) ),
            "[<input>, line 1, col 4]: Invalid id rune: \"$\" (U+0024)",
        ),
    )
    b.addErr(
        parser.MustStruct( QnameRequest ),
        mg.NewMissingFieldsError(
            nil,
            []*mg.Identifier{ 
                mkId( "namespace" ), 
                mkId( "service" ), 
                mkId( "operation" ),
            },
        ),
    )
    b.addErr(
        parser.MustStruct( QnameRequest, "f1", int32( 1 ) ),
        mg.NewUnrecognizedFieldError( nil, mkId( "f1" ) ),
    )
    b.addErr(
        parser.MustStruct( QnameRequest,
            "namespace", "bad@v1",
            "service", "svc1",
            "operation", "op1",
        ),
        &testError{ msg: "test-error-bad-ns" },
    )
    b.addErr(
        parser.MustStruct( QnameRequest,
            "namespace", "ns1@v1",
            "service", "svc1",
            "operation", "noAuthOp",
            "authentication", int32( 1 ),
        ),
        &testError{ 
            objpath.RootedAt( IdAuthentication ), 
            "test-error-no-auth-expected",
        },
    )
    b.addErr(
        parser.MustStruct( QnameRequest,
            "namespace", "ns1@v1",
            "service", "svc1",
            "operation", "badParams",
        ),
        &testError{ 
            objpath.RootedAt( IdParameters ),
            "test-error-bad-params",
        },
    )
}

func initBaseResponseTests( tsb *mgRct.ReactorTestSetBuilder ) {
    b := &reactorTestBuilder{ b: tsb, typ: QnameResponse, profile: ProfileBase }
    b.addOk(
        parser.MustStruct( QnameResponse, "result", int32( 1 ) ),
        &responseExpect{ result: mg.Int32( 1 ) },
    )
    b.addOk(
        parser.MustStruct( QnameResponse, 
            "result", parser.MustStruct( "ns1@v1/Err1", "f1", int32( 1 ) ),
        ),
        &responseExpect{
            result: parser.MustStruct( "ns1@v1/Err1", "f1", int32( 1 ) ),
        },
    )
    b.addOk(
        parser.MustStruct( QnameResponse, "error", int32( 1 ) ),
        &responseExpect{ err: mg.Int32( 1 ) },
    )
    b.addOk( 
        parser.MustSymbolMap( "result", int32( 1 ) ),
        &responseExpect{ result: mg.Int32( 1 ) },
    )
    b.addOk(
        parser.MustStruct( QnameResponse,
            "error", mg.NullVal,
            "result", mg.NullVal,
        ),
        &responseExpect{},
    )
    b.addOk( parser.MustStruct( QnameResponse ), &responseExpect{} )
    b.addOk( parser.MustSymbolMap(), &responseExpect{} )
    b.addErr(
        parser.MustStruct( QnameResponse, "f1", int32( 1 ) ),
        mg.NewUnrecognizedFieldError( nil, mkId( "f1" ) ),
    )
    b.addErr(
        parser.MustStruct( QnameResponse,
            "result", int32( 1 ),
            "error", int32( 1 ),
        ),
        NewResponseError( nil, respErrMsgMultipleResponseFields ),
    )
    errBldr := b.withProfile( ProfileImplError )
    errBldr.addErr(
        parser.MustStruct( QnameResponse, "result", int32( 1 ) ),
        &testError{ objpath.RootedAt( IdResult ), "impl-error" },
    )
    errBldr.addErr(
        parser.MustStruct( QnameResponse, "error", int32( 1 ) ),
        &testError{ objpath.RootedAt( IdError ), "impl-error" },
    )
}

func initBaseReactorTests( b *mgRct.ReactorTestSetBuilder ) {
    initBaseRequestTests( b )
    initBaseResponseTests( b )
}

//func initTypedReactorTests( b ) {
//    initTypedRequestTests( b )
//    initTypedResponseTests( b )
//}

func initReactorTests( b *mgRct.ReactorTestSetBuilder ) {
    initBaseReactorTests( b )
//    initTypedReactorTests( b )
}

func init() {
    mgRct.AddTestInitializer( mkNs( "mingle:service@v1" ), initReactorTests )
}
