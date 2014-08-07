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
    rctProfile string
    errProfile string
}

func ( b *reactorTestBuilder ) copyBuilder() *reactorTestBuilder {
    return &reactorTestBuilder{ 
        b: b.b, 
        typ: b.typ, 
        rctProfile: b.rctProfile,
        errProfile: b.errProfile,
    }
}

func ( b *reactorTestBuilder ) add( t *ReactorTest ) {
    t.Type = b.typ
    t.ReactorProfile = b.rctProfile
    t.ErrorProfile = b.errProfile
    b.b.AddTests( t )
}

func ( b *reactorTestBuilder ) addOk( in mg.Value, expct interface{} ) {
    b.add( &ReactorTest{ In: in, Expect: expct } )
}

func ( b *reactorTestBuilder ) addErr( in mg.Value, err error ) {
    b.add( &ReactorTest{ In: in, Error: err } )
}

func initBaseRequestTests( tsb *mgRct.ReactorTestSetBuilder ) {
    b := &reactorTestBuilder{ 
        b: tsb, 
        typ: QnameRequest, 
        rctProfile: ReactorProfileBase,
    }
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
    b.addErr(
        mg.Int32( 1 ),
        mg.NewTypeCastError( TypeRequest, mg.TypeInt32, nil ),
    )
    b.addErr(
        parser.MustStruct( "ns1@v1/Bad" ),
        mg.NewTypeCastError( TypeRequest, asType( "ns1@v1/Bad" ), nil ),
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
    b := &reactorTestBuilder{ 
        b: tsb, 
        typ: QnameResponse, 
        rctProfile: ReactorProfileBase,
    }
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
    errBldr := b.copyBuilder()
    errBldr.errProfile = ErrorProfileImpl
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

func initTypedRequestTests( tsb *mgRct.ReactorTestSetBuilder ) {
    b := &reactorTestBuilder{ 
        b: tsb, 
        typ: QnameRequest, 
        rctProfile: ReactorProfileTyped,
    }
    makeReq := func( ns, svc, op string, tail ...interface{} ) mg.Value {
        args := []interface{}{
            "namespace", ns,
            "service", svc,
            "operation", op,
        }
        args = append( args, tail... )
        return parser.MustStruct( QnameRequest, args... )
    }
    addOk := func( 
        ns, svc, op string, expct interface{}, tail ...interface{} ) {

        b.addOk( makeReq( ns, svc, op, tail... ), expct )
    }
    addErr := func( ns, svc, op string, err error, tail ...interface{} ) {
        b.addErr( makeReq( ns, svc, op, tail... ), err )
    }
    addOk( "ns1@v1", "svc1", "noOp",
        &requestExpect{ params: mg.EmptySymbolMap() },
    )
    addOk( "ns1@v1", "svc1", "op1",
        &requestExpect{ params: parser.MustSymbolMap( "f1", int32( 1 ) ) },
        "parameters", parser.MustSymbolMap( "f1", int32( 1 ) ),
    )
    addOk( "ns1@v1", "svc1", "op1",
        &requestExpect{ params: parser.MustSymbolMap( "f1", int32( 1 ) ) },
        "parameters", parser.MustSymbolMap( "f1", int64( 1 ) ),
    )
    addOk( "ns1@v1", "svc1", "op1",
        &requestExpect{ 
            params: parser.MustSymbolMap( 
                "f1", parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
            ),
        },
        "parameters", parser.MustSymbolMap( 
            "f1", parser.MustStruct( "ns1@v1/S1", "f1", int64( 1 ) ),
        ),
    )
    addOk( "ns1@v1", "svc2", "op1",
        &requestExpect{ auth: mg.Int32( 1 ) },
        "authentication", int32( 1 ),
    )
    addOk( "ns1@v1", "svc2", "op1",
        &requestExpect{ auth: mg.Int32( 1 ) },
        "authentication", int64( 1 ),
    )
    addOk( "ns1@v1", "svc2", "op2",
        &requestExpect{
            params: parser.MustSymbolMap( 
                "f1", parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
            ),
            auth: mg.Int32( 1 ),
        },
        "authentication", int64( 1 ),
        "parameters", parser.MustSymbolMap( 
            "f1", parser.MustStruct( "ns1@v1/S1", "f1", int64( 1 ) ),
        ),
    )
    addErr( "bad@v1", "svc1", "op1",
        NewRequestError( 
            objpath.RootedAt( IdNamespace ),
            "unrecognized value: bad@v1",
        ),
    )
    addErr( "ns1@v1", "bad", "op1",
        NewRequestError( 
            objpath.RootedAt( IdService ),
            "unrecognized value: bad",
        ),
    )
    addErr( "ns1@v1", "svc1", "bad",
        NewRequestError( 
            objpath.RootedAt( IdOperation ),
            "unrecognized value: bad",
        ),
    )
    addErr( "ns1@v1", "svc1", "op1",
        NewRequestError(
            objpath.RootedAt( IdAuthentication ),
            "this service does not accept authentication",
        ),
        "authentication", int32( 1 ),
    )
    addErr( "ns1@v1", "svc2", "op1",
        mg.NewTypeCastError(
            mg.TypeInt32,
            mg.TypeBuffer,
            objpath.RootedAt( IdAuthentication ),
        ),
        "authentication", []byte{ 0 },
    )
    addErr( "ns1@v1", "svc1", "op2",
        mg.NewMissingFieldsError(
            objpath.RootedAt( IdParameters ),
            []*mg.Identifier{ mkId( "f1" ) },
        ),
    )
    addErr( "ns1@v1", "svc1", "op2",
        mg.NewUnrecognizedFieldError(
            objpath.RootedAt( IdParameters ),
            mkId( "badField" ),
        ),
        "parameters", parser.MustSymbolMap(
            "f1", int32( 1 ),
            "f2", int32( 2 ),
        ),
    )
    addErr( "ns1@v1", "svc1", "op2",
        mg.NewTypeCastError(
            mg.TypeInt32,
            mg.TypeBuffer,
            objpath.RootedAt( IdParameters ).Descend( mkId( "f1" ) ),
        ),
        "parameters", parser.MustSymbolMap( "f1", []byte{ 0 } ),
    )
}

func initTypedReactorTests( b *mgRct.ReactorTestSetBuilder ) {
    initTypedRequestTests( b )
//    initTypedResponseTests( b )
}

func initReactorTests( b *mgRct.ReactorTestSetBuilder ) {
    initBaseReactorTests( b )
    initTypedReactorTests( b )
}

func init() {
    mgRct.AddTestInitializer( mkNs( "mingle:service@v1" ), initReactorTests )
}
