package service

import (
    "bitgirder/objpath"
    mgRct "mingle/reactor"
    mg "mingle"
    "mingle/parser"
//    "log"
)

func initBaseRequestTests( b *mgRct.ReactorTestSetBuilder ) {
    addTyped := func( t *ReactorBaseTest ) {
        t.Type = QnameRequest
        b.AddTests( t )
    }
    add := func( in mg.Value, expct interface{} ) {
        addTyped( &ReactorBaseTest{ In: in, Expect: expct } )
    }
    addErr := func( in mg.Value, err error ) {
        addTyped( &ReactorBaseTest{ In: in, Error: err } )
    }
    add(
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
    add(
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
    add(
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
    add(
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
    typReq := QnameRequest.AsAtomicType()
    addErr(
        mg.Int32( 1 ),
        mg.NewTypeCastError( typReq, mg.TypeInt32, nil ),
    )
    addErr(
        parser.MustStruct( "ns1@v1/Bad" ),
        mg.NewTypeCastError( typReq, asType( "ns1@v1/Bad" ), nil ),
    )
    addErr(
        parser.MustStruct( QnameRequest, "namespace", "Bad" ),
        mg.NewValueCastError( 
            objpath.RootedAt( mkId( "namespace" ) ), 
            "[<input>, line 1, col 1]: Illegal start of identifier part: \"B\" (U+0042)",
        ),
    )
    addErr(
        parser.MustStruct( QnameRequest, 
            "namespace", "ns1@v1",
            "service", "bad$id",
        ),
        mg.NewValueCastError( 
            objpath.RootedAt( mkId( "service" ) ), 
            "[<input>, line 1, col 4]: Invalid id rune: \"$\" (U+0024)",
        ),
    )
    addErr(
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
    addErr(
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
    addErr(
        parser.MustStruct( QnameRequest, "f1", int32( 1 ) ),
        mg.NewUnrecognizedFieldError( nil, mkId( "f1" ) ),
    )
    addErr(
        parser.MustStruct( QnameRequest,
            "namespace", "bad@v1",
            "service", "svc1",
            "operation", "op1",
        ),
        &testError{ msg: "test-error-bad-ns" },
    )
    addErr(
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
    addErr(
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

func initBaseResponseTests( b *mgRct.ReactorTestSetBuilder ) {
    addTyped := func( t *ReactorBaseTest ) {
        t.Type = QnameResponse
        b.AddTests( t )
    }
    add := func( in mg.Value, expct interface{} ) {
        addTyped( &ReactorBaseTest{ In: in, Expect: expct } )
    }
    addErr := func( in mg.Value, err error ) {
        addTyped( &ReactorBaseTest{ In: in, Error: err } )
    }
    add(
        parser.MustStruct( QnameResponse, "result", int32( 1 ) ),
        &responseExpect{ result: mg.Int32( 1 ) },
    )
    add(
        parser.MustStruct( QnameResponse, "error", int32( 1 ) ),
        &responseExpect{ err: mg.Int32( 1 ) },
    )
    add( 
        parser.MustSymbolMap( "result", int32( 1 ) ),
        &responseExpect{ result: mg.Int32( 1 ) },
    )
    add(
        parser.MustStruct( QnameResponse,
            "error", mg.NullVal,
            "result", mg.NullVal,
        ),
        &responseExpect{},
    )
    add( parser.MustStruct( QnameResponse ), &responseExpect{} )
    add( parser.MustSymbolMap(), &responseExpect{} )
    addErr(
        parser.MustStruct( QnameResponse, "f1", int32( 1 ) ),
        mg.NewUnrecognizedFieldError( nil, mkId( "f1" ) ),
    )
    addErr(
        parser.MustStruct( QnameResponse,
            "result", int32( 1 ),
            "error", int32( 1 ),
        ),
        NewResponseError( nil, "response contains both result and error" ),
    )
}

func initBaseReactorTests( b *mgRct.ReactorTestSetBuilder ) {
    initBaseRequestTests( b )
//    initBaseResponseTests( b )
}

func initReactorTests( b *mgRct.ReactorTestSetBuilder ) {
    initBaseReactorTests( b )
}

func init() {
    mgRct.AddTestInitializer( mkNs( "mingle:service@v1" ), initReactorTests )
}
