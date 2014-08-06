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

func setBuilder(
    addr **mgRct.BuildReactor ) ( mgRct.ReactorEventProcessor, error ) {

    *addr = mgRct.NewBuildReactor( mgRct.ValueBuilderFactory )
    return *addr, nil
}

func ( b *reqValueBuilder ) StartAuthentication( 
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    if b.ctx.Operation.Equals( mkId( "noAuthOp" ) ) {
        return nil, &testError{ path, "test-error-no-auth-expected" }
    }
    return setBuilder( &b.authBldr )
}

func ( b *reqValueBuilder ) StartParameters(
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    if b.ctx.Operation.Equals( mkId( "badParams" ) ) {
        return nil, &testError{ path, "test-error-bad-params" }
    }
    return setBuilder( &b.paramsBldr )
}

func ( t *ReactorBaseTest ) feedSource( 
    rct mgRct.ReactorEventProcessor, c *mgRct.ReactorTestCall ) bool {

    err := mgRct.FeedSource( t.In, rct )
    if err == nil { 
        c.Falsef( t.Expect == nil, "did not expect a value" )
        return true
    }
    c.EqualErrors( t.Error, err )
    return false
}

func ( t *ReactorBaseTest ) callRequest( c *mgRct.ReactorTestCall ) {
    reqBld := &reqValueBuilder{}
    rct := NewRequestReactor( reqBld )
    pip := mgRct.InitReactorPipeline( rct )
    if ! t.feedSource( pip, c ) { return }
    expct := t.Expect.( *requestExpect )
    mgRct.CheckBuiltValue( expct.auth, reqBld.authBldr, c.Descend( "auth" ) )
    mgRct.CheckBuiltValue( 
        expct.params, reqBld.paramsBldr, c.Descend( "params" ) )
}

type respValueBuilder struct {
    resBldr *mgRct.BuildReactor
    errBldr *mgRct.BuildReactor
}

func ( b *respValueBuilder ) StartResult( 
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    return setBuilder( &b.resBldr )
}

func ( b *respValueBuilder ) StartError( 
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    return setBuilder( &b.errBldr )
}

func ( t *ReactorBaseTest ) callResponse( c *mgRct.ReactorTestCall ) {
    respBld := &respValueBuilder{}
    rct := NewResponseReactor( respBld )
    pip := mgRct.InitReactorPipeline( rct )
    if ! t.feedSource( pip, c ) { return }
    expct := t.Expect.( *responseExpect )
    mgRct.CheckBuiltValue( 
        expct.result, respBld.resBldr, c.Descend( "result" ) )
    mgRct.CheckBuiltValue( expct.err, respBld.errBldr, c.Descend( "error" ) )
}

func ( t *ReactorBaseTest ) Call( c *mgRct.ReactorTestCall ) {
    switch {
    case t.Type.Equals( QnameRequest ): t.callRequest( c )
    case t.Type.Equals( QnameResponse ): t.callResponse( c )
    default: c.Fatalf( "unhandled expect type: %s", t.Type )
    }
}
