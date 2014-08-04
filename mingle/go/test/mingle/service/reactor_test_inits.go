package service

import (
    mgRct "mingle/reactor"
    mg "mingle"
    "mingle/parser"
    "log"
)

func initBaseReactorTests( b *mgRct.ReactorTestSetBuilder ) {
    b.AddTests(
        &ServiceReactorBaseTest{
            In: parser.MustStruct( QnameRequest,
                "namespace", "ns1@v1",
                "service", "svc1",
                "operation", "op1",
                "parameters", mg.EmptySymbolMap(),
                "authentication", int32( 1 ),
            ),
            Expect: requestImpl{
                ctx: &RequestContext{
                    Namespace: mkNs( "ns1@v1" ),
                    Service: mkId( "svc1" ),
                    Operation: mkId( "op1" ),
                },
                params: mg.EmptySymbolMap(),
                auth: mg.Int32( 1 ),
            },
        },
    )
}

func initReactorTests( b *mgRct.ReactorTestSetBuilder ) {
    log.Printf( "added tests" )
    initBaseReactorTests( b )
}

func init() {
    mgRct.AddTestInitializer( mkNs( "mingle:service@v1" ), initReactorTests )
}
