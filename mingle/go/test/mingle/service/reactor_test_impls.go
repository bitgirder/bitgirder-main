package service

import (
    mgRct "mingle/reactor"
    "bitgirder/objpath"
)

type reqValueBuilder struct {
    ctx *RequestContext
    authBldr *mgRct.BuildReactor
    paramsBldr *mgRct.BuildReactor
}

func ( b *reqValueBuilder ) StartRequest( 
    ctx *RequestContext, path objpath.PathNode ) error {

    b.ctx = ctx
    if b.ctx.Namespace.Equals( mkNs( "bad@v1" ) ) {
        return &testError{ path, "test-error-bad-ns" }
    }
    return nil
}

func ( b *reqValueBuilder ) setBuilder(
    addr **mgRct.BuildReactor ) ( mgRct.ReactorEventProcessor, error ) {

    *addr = mgRct.NewBuildReactor( mgRct.ValueBuilderFactory )
    return *addr, nil
}

func ( b *reqValueBuilder ) StartAuthentication( 
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    if b.ctx.Operation.Equals( mkId( "noAuthOp" ) ) {
        return nil, &testError{ path, "test-error-no-auth-expected" }
    }
    return b.setBuilder( &b.authBldr )
}

func ( b *reqValueBuilder ) StartParameters(
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    if b.ctx.Operation.Equals( mkId( "badParams" ) ) {
        return nil, &testError{ path, "test-error-bad-params" }
    }
    return b.setBuilder( &b.paramsBldr )
}

func ( t *ReactorBaseTest ) callRequest( c *mgRct.ReactorTestCall ) {
    reqBld := &reqValueBuilder{}
    rct := NewRequestReactor( reqBld )
    pip := mgRct.InitReactorPipeline( rct )
    if err := mgRct.FeedSource( t.In, pip ); err == nil {
        c.Falsef( t.Expect == nil, "did not expect a value" )
        expct := t.Expect.( *requestExpect )
        mgRct.CheckBuiltValue( 
            expct.auth, reqBld.authBldr, c.Descend( "auth" ) )
        mgRct.CheckBuiltValue( 
            expct.params, reqBld.paramsBldr, c.Descend( "params" ) )
    } else { c.EqualErrors( t.Error, err ) }
}

func ( t *ReactorBaseTest ) Call( c *mgRct.ReactorTestCall ) {
    switch {
    case t.Type.Equals( QnameRequest ): t.callRequest( c )
    default: c.Fatalf( "unhandled expect type: %s", t.Type )
    }
}
