package service

import (
    "bitgirder/objpath"
    mgRct "mingle/reactor"
    mg "mingle"
    "mingle/parser"
//    "log"
)

type reactorTestBuilder struct {
    b *mgRct.ReactorTestSliceBuilder
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

func ( b *reactorTestBuilder ) addOk( in, expct interface{} ) {
    b.add( &ReactorTest{ In: in, Expect: expct } )
}

func ( b *reactorTestBuilder ) addErr( in interface{}, err error ) {
    b.add( &ReactorTest{ In: in, Error: err } )
}

func ( b *reactorTestBuilder ) makeImplErrorBuilder() *reactorTestBuilder {
    res := b.copyBuilder()
    res.errProfile = ErrorProfileImpl
    return res
}

func ( b *reactorTestBuilder ) addRequestImplErrorTests() {
    b2 := b.makeImplErrorBuilder()
    b2.addErr(
        parser.MustStruct( QnameRequest,
            "namespace", "mingle:service:fail@v1",
            "service", "svc1",
            "operation", "op1",
        ),
        &testError{ msg: "start-request-impl-error" },
    )
    b2.addErr(
        parser.MustStruct( QnameRequest,
            "namespace", "mingle:service@v1",
            "service", "svc1",
            "operation", "failStartAuthentication",
            "authentication", int32( 1 ),
        ),
        &testError{ 
            objpath.RootedAt( IdAuthentication ), 
            "start-authentication-impl-error",
        },
    )
    b2.addErr(
        parser.MustStruct( QnameRequest,
            "namespace", "mingle:service@v1",
            "service", "svc1",
            "authentication", int32( 1 ),
            "operation", "failStartParameters",
        ),
        &testError{ 
            objpath.RootedAt( IdParameters ),
            "start-parameters-impl-error",
        },
    )
}

func ( b *reactorTestBuilder ) addResponseImplErrorTests() {
    b2 := b.makeImplErrorBuilder()
    reqCtx := &RequestContext{
        Namespace: mkNs( "mingle:service@v1" ),
        Service: mkId( "svc1" ),
        Operation: mkId( "failResponse" ),
    }
    b2.addErr(
        &responseInput{
            in: parser.MustStruct( QnameResponse, "result", int32( 1 ) ),
            reqCtx: reqCtx,
        },
        &testError{ objpath.RootedAt( IdResult ), "impl-error" },
    )
    b2.addErr(
        &responseInput{
            in: parser.MustStruct( QnameResponse, 
                "error", parser.MustStruct( "ns1@v1/Err1", "f1", int32( 1 ) ),
            ),
            reqCtx: reqCtx,
        },
        &testError{ objpath.RootedAt( IdError ), "impl-error" },
    )
}

func initBaseRequestTests( tsb *mgRct.ReactorTestSliceBuilder ) {
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
        mg.NewTypeInputError( TypeRequest, mg.TypeInt32, nil ),
    )
    b.addErr(
        parser.MustStruct( "ns1@v1/Bad" ),
        mg.NewTypeInputError( TypeRequest, asType( "ns1@v1/Bad" ), nil ),
    )
    b.addErr(
        parser.MustStruct( QnameRequest, "namespace", "Bad" ),
        mg.NewInputError( 
            objpath.RootedAt( mkId( "namespace" ) ), 
            "[<input>, line 1, col 1]: Illegal start of identifier part: \"B\" (U+0042)",
        ),
    )
    b.addErr(
        parser.MustStruct( QnameRequest, 
            "namespace", "ns1@v1",
            "service", "bad$id",
        ),
        mg.NewInputError( 
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
        mg.NewInputError( 
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
    b.addRequestImplErrorTests()
}

func initBaseResponseTests( tsb *mgRct.ReactorTestSliceBuilder ) {
    b := &reactorTestBuilder{ 
        b: tsb, 
        typ: QnameResponse, 
        rctProfile: ReactorProfileBase,
    }
    mkInput := func( in mg.Value ) interface{} {
        return &responseInput{
            in: in,
            reqCtx: &RequestContext{
                Namespace: mkNs( "mingle:service@v1" ),
                Service: mkId( "baseService" ),
                Operation: mkId( "testBaseResponseReactor" ),
            },
        }
    }
    b.addOk(
        mkInput( parser.MustStruct( QnameResponse, "result", int32( 1 ) ) ),
        &ResultExpectation{ Result: mg.Int32( 1 ) },
    )
    b.addOk(
        mkInput(
            parser.MustStruct( QnameResponse, 
                "result", parser.MustStruct( "ns1@v1/Err1", "f1", int32( 1 ) ),
            ),
        ),
        &ResultExpectation{
            Result: parser.MustStruct( "ns1@v1/Err1", "f1", int32( 1 ) ),
        },
    )
    b.addOk(
        mkInput( parser.MustStruct( QnameResponse, "error", int32( 1 ) ) ),
        &ResultExpectation{ Error: mg.Int32( 1 ) },
    )
    b.addOk( 
        mkInput( parser.MustSymbolMap( "result", int32( 1 ) ) ),
        &ResultExpectation{ Result: mg.Int32( 1 ) },
    )
    b.addOk(
        mkInput(
            parser.MustStruct( QnameResponse,
                "error", mg.NullVal,
                "result", mg.NullVal,
            ),
        ),
        &ResultExpectation{},
    )
    b.addOk( 
        mkInput( parser.MustStruct( QnameResponse ) ), 
        &ResultExpectation{},
    )
    b.addOk( mkInput( parser.MustSymbolMap() ), &ResultExpectation{} )
    b.addErr(
        mkInput( parser.MustStruct( QnameResponse, "f1", int32( 1 ) ) ),
        mg.NewUnrecognizedFieldError( nil, mkId( "f1" ) ),
    )
    b.addErr(
        mkInput(
            parser.MustStruct( QnameResponse,
                "result", int32( 1 ),
                "error", int32( 1 ),
            ),
        ),
        NewResponseError( nil, respErrMsgMultipleResponseFields ),
    )
    b.addResponseImplErrorTests()
}

func initBaseReactorTests( b *mgRct.ReactorTestSliceBuilder ) {
    initBaseRequestTests( b )
    initBaseResponseTests( b )
}

func initTypedRequestTests( tsb *mgRct.ReactorTestSliceBuilder ) {
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
    addOk( "mingle:tck@v1", "svc1", "getFixedInt",
        &requestExpect{ params: mg.EmptySymbolMap() },
    )
    addOk( "mingle:tck@v1", "svc1", "echoS1",
        &requestExpect{ 
            params: parser.MustSymbolMap( 
                "f1", parser.MustStruct( "mingle:tck@v1/S1", "f1", int32( 1 ) ),
            ),
        },
        "parameters", parser.MustSymbolMap( 
            "f1", parser.MustStruct( "mingle:tck@v1/S1", "f1", int32( 1 ) ),
        ),
    )
    addOk( "mingle:tck@v1", "svc1", "echoS1",
        &requestExpect{ 
            params: parser.MustSymbolMap( 
                "f1", parser.MustStruct( "mingle:tck@v1/S1", "f1", int32( 1 ) ),
            ),
        },
        "parameters", parser.MustSymbolMap( 
            "f1", parser.MustSymbolMap( "f1", int64( 1 ) ),
        ),
    )
    addOk( "mingle:tck@v1", "svc1", "echoS1",
        &requestExpect{ 
            params: parser.MustSymbolMap( 
                "f1", parser.MustStruct( "mingle:tck@v1/S1", "f1", int32( 1 ) ),
            ),
        },
        "parameters", parser.MustSymbolMap( 
            "f1", parser.MustStruct( "mingle:tck@v1/S1", "f1", int64( 1 ) ),
        ),
    )
    addOk( "mingle:tck@v1", "svc2", "getFixedInt",
        &requestExpect{ auth: mg.Int32( 1 ), params: mg.EmptySymbolMap() },
        "authentication", int32( 1 ),
    )
    addOk( "mingle:tck@v1", "svc2", "getFixedInt",
        &requestExpect{ auth: mg.Int32( 1 ), params: mg.EmptySymbolMap() },
        "authentication", int64( 1 ),
    )
    addOk( "mingle:tck@v1", "svc2", "echoS1",
        &requestExpect{
            params: parser.MustSymbolMap( 
                "f1", parser.MustStruct( "mingle:tck@v1/S1", "f1", int32( 1 ) ),
            ),
            auth: mg.Int32( 1 ),
        },
        "authentication", int64( 1 ),
        "parameters", parser.MustSymbolMap( 
            "f1", parser.MustStruct( "mingle:tck@v1/S1", "f1", int64( 1 ) ),
        ),
    )
    addErr( "no:such@v1", "svc1", "getFixedInt",
        NewRequestError( nil, "no services in namespace: no:such@v1" ),
    )
    addErr( "mingle:tck@v1", "noSuchService", "getFixedInt",
        NewRequestError( 
            nil, 
            "namespace mingle:tck@v1 has no service with id: no-such-service",
        ),
    )
    addErr( "mingle:tck@v1", "svc1", "noSuchOp",
        NewRequestError( 
            nil, 
            "service mingle:tck@v1.svc1 has no such operation: no-such-op",
        ),
    )
    addErr( "mingle:tck@v1", "svc1", "getFixedInt",
        NewRequestError(
            objpath.RootedAt( IdAuthentication ),
            "service does not accept authentication",
        ),
        "authentication", int32( 1 ),
    )
    addErr( "mingle:tck@v1", "svc2", "getFixedInt",
        mg.NewTypeInputError(
            mg.TypeInt32,
            mg.TypeBuffer,
            objpath.RootedAt( IdAuthentication ),
        ),
        "authentication", []byte{ 0 },
    )
    addErr( "mingle:tck@v1", "svc1", "echoS1",
        mg.NewMissingFieldsError(
            objpath.RootedAt( IdParameters ),
            []*mg.Identifier{ mkId( "f1" ) },
        ),
    )
    addErr( "mingle:tck@v1", "svc1", "echoS1",
        mg.NewUnrecognizedFieldError(
            objpath.RootedAt( IdParameters ),
            mkId( "badField" ),
        ),
        "parameters", parser.MustSymbolMap(
            "f1", parser.MustStruct( "mingle:tck@v1/S1", "f1", int32( 1 ) ),
            "badField", int32( 2 ),
        ),
    )
    addErr( "mingle:tck@v1", "svc1", "echoS1",
        mg.NewTypeInputError(
            asType( "mingle:tck@v1/S1" ),
            mg.TypeBuffer,
            objpath.RootedAt( IdParameters ).Descend( mkId( "f1" ) ),
        ),
        "parameters", parser.MustSymbolMap( "f1", []byte{ 0 } ),
    )
    b.addRequestImplErrorTests()
}

func initTypedResponseTests( tsb *mgRct.ReactorTestSliceBuilder ) {
    b := &reactorTestBuilder{
        b: tsb, 
        typ: QnameResponse, 
        rctProfile: ReactorProfileTyped,
    }
    mkInput := func( ns, svc, op string, in mg.Value ) *responseInput {
        return &responseInput{
            in: in,
            reqCtx: &RequestContext{
                Namespace: mkNs( ns ),
                Service: mkId( svc ),
                Operation: mkId( op ),
            },
        }
    }
    addOk := func( ns, svc, op string, in mg.Value, expct interface{} ) {
        b.addOk( mkInput( ns, svc, op, in ), expct )
    }
    addErr := func( ns, svc, op string, in mg.Value, err error ) {
        b.addErr( mkInput( ns, svc, op, in ), err )
    }
    mkResp := func( fld string, val interface{} ) mg.Value {
        return parser.MustStruct( QnameResponse, fld, val )
    }
    mkRes := func( val interface{} ) mg.Value { return mkResp( "result", val ) }
    mkErr := func( val interface{} ) mg.Value { return mkResp( "error", val ) }
    // not exhaustively testing all external error types, just checking one to
    // confirm that external error types are in fact picked up by the reactor
    addImplicitErrCoverage := func( ns, svc, op string ) {
        f := func( qn *mg.QualifiedTypeName ) {
            addOk( ns, svc, op,
                mkErr( parser.MustStruct( qn, "message", "test-message") ),
                &ResultExpectation{ 
                    Error: parser.MustStruct( qn, "message", "test-message" ),
                },
            )
        }
        f( QnameRequestError )
    }
    addImplicitErrCoverage( "mingle:tck@v1", "svc1", "getFixedInt" )
    addImplicitErrCoverage( "mingle:tck@v1", "svc1", "echoS1" )
    addImplicitErrCoverage( "mingle:tck@v1", "svc2", "getFixedInt" )
    addOk( "mingle:tck@v1", "svc1", "getFixedInt",
        mkRes( int32( 1 ) ),
        &ResultExpectation{ Result: mg.Int32( 1 ) },
    )
    addOk( "mingle:tck@v1", "svc1", "getFixedInt",
        mkRes( int64( 1 ) ),
        &ResultExpectation{ Result: mg.Int32( 1 ) },
    )
    addOk( "mingle:tck@v1", "svc1", "echoS1",
        mkRes( parser.MustStruct( "mingle:tck@v1/S1", "f1", int32( 1 ) ) ),
        &ResultExpectation{
            Result: parser.MustStruct( "mingle:tck@v1/S1", "f1", int32( 1 ) ),
        },
    )
    addOk( "mingle:tck@v1", "svc1", "echoS1",
        mkRes( parser.MustStruct( "mingle:tck@v1/S1", "f1", int64( 1 ) ) ),
        &ResultExpectation{
            Result: parser.MustStruct( "mingle:tck@v1/S1", "f1", int32( 1 ) ),
        },
    )
    addOk( "mingle:tck@v1", "svc1", "echoS1",
        mkErr( parser.MustStruct( "mingle:tck@v1/Err1", "f1", int64( 1 ) ) ),
        &ResultExpectation{
            Error: parser.MustStruct( "mingle:tck@v1/Err1", "f1", int32( 1 ) ),
        },
    )
    addOk( "mingle:tck@v1", "svc2", "getFixedInt",
        mkErr( 
            parser.MustStruct( "mingle:tck@v1/AuthErr1", "f1", int64( 1 ) ) ),
        &ResultExpectation{
            Error: parser.MustStruct( "mingle:tck@v1/AuthErr1", 
                "f1", int32( 1 ),
            ),
        },
    )
    addErr( "mingle:tck@v1", "svc1", "echoS1",
        mkErr( parser.MustStruct( "mingle:tck@v1/Err1", "f1", []byte{ 0 } ) ),
        mg.NewTypeInputError(
            mg.TypeInt32,
            mg.TypeBuffer,
            objpath.RootedAt( IdError ).Descend( mkId( "f1" ) ),
        ),
    )
    addErr( "mingle:tck@v1", "svc1", "getFixedInt",
        mkErr( parser.MustStruct( "mingle:tck@v1/Err2", "f1", int32( 1 ) ) ),
        NewResponseError( 
            objpath.RootedAt( IdError ),
            "unexpected error: mingle:tck@v1/Err2",
        ),
    )
    addErr( "mingle:tck@v1", "svc1", "echoS1",
        mkErr( parser.MustStruct( "mingle:tck@v1/Err2" ) ),
        NewResponseError( 
            objpath.RootedAt( IdError ),
            "unexpected error: mingle:tck@v1/Err2",
        ),
    )
    addErr( "mingle:tck@v1", "svc2", "getFixedInt",
        mkErr( parser.MustStruct( "mingle:tck@v1/Err2" ) ),
        NewResponseError( 
            objpath.RootedAt( IdError ),
            "unexpected error: mingle:tck@v1/Err2",
        ),
    )
    addErr( "mingle:tck@v1", "svc2", "getFixedInt",
        mkErr( 
            parser.MustStruct( "mingle:tck@v1/AuthErr1", "f1", []byte{ 0 } ),
        ),
        mg.NewTypeInputError(
            mg.TypeInt32,
            mg.TypeBuffer,
            objpath.RootedAt( IdError ).Descend( mkId( "f1" ) ),
        ),
    )
    b.addResponseImplErrorTests()
}

func initTypedReactorTests( b *mgRct.ReactorTestSliceBuilder ) {
    initTypedRequestTests( b )
    initTypedResponseTests( b )
}

func GetReactorTests() []mgRct.ReactorTest {
    b := mgRct.NewReactorTestSliceBuilder()
    initBaseReactorTests( b )
    initTypedReactorTests( b )
    return b.GetTests()
}
