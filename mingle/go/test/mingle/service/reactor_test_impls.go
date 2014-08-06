package service

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "bitgirder/objpath"
)

func setBuilder(
    addr **mgRct.BuildReactor ) ( mgRct.ReactorEventProcessor, error ) {

    *addr = mgRct.NewBuildReactor( mgRct.ValueBuilderFactory )
    return *addr, nil
}

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

func ( t *ReactorTest ) feedSource( 
    rct mgRct.ReactorEventProcessor, c *mgRct.ReactorTestCall ) bool {

    if mv, ok := t.In.( mg.Value ); ok {
        c.Logf( "feeding %s", mg.QuoteValue( mv ) )
    }
    err := mgRct.FeedSource( t.In, rct )
    if err == nil { 
        c.Falsef( t.Expect == nil, "did not expect a value" )
        return true
    }
    c.EqualErrors( t.Error, err )
    return false
}

type respValueBuilder struct {
    resBldr *mgRct.BuildReactor
    errBldr *mgRct.BuildReactor
    t *ReactorTest
}

func ( b *respValueBuilder ) implError( path objpath.PathNode ) error {
    if b.t.Profile == ProfileImplError { 
        return &testError{ path, "impl-error" } 
    }
    return nil
}

func ( b *respValueBuilder ) StartResult( 
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    if err := b.implError( path ); err != nil { return nil, err }
    return setBuilder( &b.resBldr )
}

func ( b *respValueBuilder ) StartError( 
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    if err := b.implError( path ); err != nil { return nil, err }
    return setBuilder( &b.errBldr )
}

type reactorTestCall struct {

    *mgRct.ReactorTestCall
    t *ReactorTest

    reqBld *reqValueBuilder
    respBld *respValueBuilder
}

func ( t *reactorTestCall ) createRequestReactor() mgRct.ReactorEventProcessor {
    t.reqBld = &reqValueBuilder{}
    rct := NewRequestReactor( t.reqBld )
    return mgRct.InitReactorPipeline( rct )
}

func ( t *reactorTestCall ) createResponseReactor(
    ) mgRct.ReactorEventProcessor {

    t.respBld = &respValueBuilder{ t: t.t }
    rct := NewResponseReactor( t.respBld )
    return mgRct.InitReactorPipeline( rct )
}

func ( t *reactorTestCall ) feedSource( rct mgRct.ReactorEventProcessor ) bool {
    if mv, ok := t.t.In.( mg.Value ); ok {
        t.Logf( "feeding %s", mg.QuoteValue( mv ) )
    }
    err := mgRct.FeedSource( t.t.In, rct )
    if err == nil { 
        t.Falsef( t.t.Expect == nil, "did not expect a value" )
        return true
    }
    t.EqualErrors( t.t.Error, err )
    return false
}

func ( t *reactorTestCall ) createReactor() mgRct.ReactorEventProcessor {
    switch typ := t.t.Type; {
    case typ.Equals( QnameRequest ): return t.createRequestReactor()
    case typ.Equals( QnameResponse ): return t.createResponseReactor()
    }
    panic( libErrorf( "unhandled expect type: %s", t.t.Type ) )
}

func ( t *reactorTestCall ) completeRequest() {
    expct := t.t.Expect.( *requestExpect )
    mgRct.CheckBuiltValue( expct.auth, t.reqBld.authBldr, t.Descend( "auth" ) )
    mgRct.CheckBuiltValue( 
        expct.params, t.reqBld.paramsBldr, t.Descend( "params" ) )
}

func ( t *reactorTestCall ) completeResponse() {
    expct := t.t.Expect.( *responseExpect )
    mgRct.CheckBuiltValue( 
        expct.result, t.respBld.resBldr, t.Descend( "result" ) )
    mgRct.CheckBuiltValue( expct.err, t.respBld.errBldr, t.Descend( "error" ) )
}

func ( t *reactorTestCall ) complete() { 
    switch typ := t.t.Type; {
    case typ.Equals( QnameRequest ): t.completeRequest()
    case typ.Equals( QnameResponse ): t.completeResponse()
    default: t.Fatalf( "unhandled type: %s", typ )
    }
}

func ( t *reactorTestCall ) call() {
    rct := t.createReactor()
    if t.t.feedSource( rct, t.ReactorTestCall ) { t.complete() }
}

func ( t *ReactorTest ) Call( c *mgRct.ReactorTestCall ) {
    ( &reactorTestCall{ ReactorTestCall: c, t: t } ).call()
//    switch {
//    case t.Type.Equals( QnameRequest ): t.callRequest( c )
//    case t.Type.Equals( QnameResponse ): t.callResponse( c )
//    default: c.Fatalf( "unhandled expect type: %s", t.Type )
//    }
}
