package service

import (
    "bitgirder/objpath"
    mgRct "mingle/reactor"
    mg "mingle"
    "mingle/parser"
    "log"
)

func initBaseReactorTests( b *mgRct.ReactorTestSetBuilder ) {
    add := func( in mg.Value, expct *requestImpl ) {
        b.AddTests( &ServiceReactorBaseTest{ In: in, Expect: expct } )
    }
    addErr := func( in mg.Value, err error ) {
        b.AddTests( &ServiceReactorBaseTest{ In: in, Error: err } )
    }
    add(
        parser.MustStruct( QnameRequest,
            "namespace", "ns1@v1",
            "service", "svc1",
            "operation", "op1",
            "parameters", mg.EmptySymbolMap(),
            "authentication", int32( 1 ),
        ),
        &requestImpl{
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
        &requestImpl{
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
        &requestImpl{
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
        &requestImpl{
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
        mg.NewValueCastError( objpath.RootedAt( mkId( "namespace" ) ), "STUB" ),
    )
    addErr(
        parser.MustStruct( QnameRequest, 
            "namespace", "ns1@v1",
            "service", "not good",
        ),
        mg.NewValueCastError( objpath.RootedAt( mkId( "service" ) ), "STUB" ),
    )
    addErr(
        parser.MustStruct( QnameRequest, 
            "namespace", "ns1@v1",
            "service", "svc1",
            "operation", "not good",
        ),
        mg.NewValueCastError( objpath.RootedAt( mkId( "operation" ) ), "STUB" ),
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
}

func initReactorTests( b *mgRct.ReactorTestSetBuilder ) {
    log.Printf( "added tests" )
    initBaseReactorTests( b )
}

func init() {
    mgRct.AddTestInitializer( mkNs( "mingle:service@v1" ), initReactorTests )
}
