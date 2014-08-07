package service

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "bitgirder/objpath"
)

type reactorTestChecker interface {
    checkValue( t *reactorTestCall )
}

func setBuilder(
    addr **mgRct.BuildReactor ) ( mgRct.ReactorEventProcessor, error ) {

    *addr = mgRct.NewBuildReactor( mgRct.ValueBuilderFactory )
    return *addr, nil
}

type baseReqBuilder struct {
    ctx *RequestContext
    authBldr *mgRct.BuildReactor
    paramsBldr *mgRct.BuildReactor
}

func ( b *baseReqBuilder ) StartRequest( 
    ctx *RequestContext, path objpath.PathNode ) error {

    b.ctx = ctx
    if b.ctx.Namespace.Equals( mkNs( "bad@v1" ) ) {
        return &testError{ path, "test-error-bad-ns" }
    }
    return nil
}

func ( b *baseReqBuilder ) StartAuthentication( 
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    if b.ctx.Operation.Equals( mkId( "noAuthOp" ) ) {
        return nil, &testError{ path, "test-error-no-auth-expected" }
    }
    return setBuilder( &b.authBldr )
}

func ( b *baseReqBuilder ) StartParameters(
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    if b.ctx.Operation.Equals( mkId( "badParams" ) ) {
        return nil, &testError{ path, "test-error-bad-params" }
    }
    return setBuilder( &b.paramsBldr )
}

func ( b *baseReqBuilder ) checkValue( t *reactorTestCall ) {
    expct := t.t.Expect.( *requestExpect )
    mgRct.CheckBuiltValue( expct.auth, b.authBldr, t.Descend( "auth" ) )
    mgRct.CheckBuiltValue( expct.params, b.paramsBldr, t.Descend( "params" ) )
}

type respValueBuilder struct {
    resBldr *mgRct.BuildReactor
    errBldr *mgRct.BuildReactor
    t *ReactorTest
}

func ( b *respValueBuilder ) implError( path objpath.PathNode ) error {
    if b.t.ErrorProfile == ErrorProfileImpl { 
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

func ( b *respValueBuilder ) checkValue( t *reactorTestCall ) {
    expct := t.t.Expect.( *responseExpect )
    mgRct.CheckBuiltValue( expct.result, b.resBldr, t.Descend( "result" ) )
    mgRct.CheckBuiltValue( expct.err, b.errBldr, t.Descend( "error" ) )
}

type reactorTestCall struct {

    *mgRct.ReactorTestCall
    t *ReactorTest

    chkObj reactorTestChecker
}

func ( t *reactorTestCall ) initBaseRequestTest() RequestReactorInterface {
    res := &baseReqBuilder{}
    t.chkObj = res 
    return res
}

func ( t *reactorTestCall ) initTypedRequestTest() RequestReactorInterface {
    res := &baseReqBuilder{}
    m := NewOperationMap( testTypeDefs )
    m.MustAddServiceInstance( 
        mkNs( "ns1@v1" ), mkId( "svc1" ), mkQn( "ns1@v1/Service1" ) )
    m.MustAddServiceInstance( 
        mkNs( "ns1@v1" ), mkId( "svc2" ), mkQn( "ns1@v1/Service2" ) )
    t.chkObj = res
    return AsTypedRequestReactorInterface( res, m )
}

func ( t *reactorTestCall ) initRequestTest() mgRct.ReactorEventProcessor {
    var reqIface RequestReactorInterface
    switch t.t.ReactorProfile {
    case ReactorProfileBase: reqIface = t.initBaseRequestTest()
    case ReactorProfileTyped: reqIface = t.initTypedRequestTest()
    default: t.Fatalf( "unhandled profile: %s", t.t.ReactorProfile )
    }
    rct := NewRequestReactor( reqIface )
    return mgRct.InitReactorPipeline( rct )
}

func ( t *reactorTestCall ) initResponseTest() mgRct.ReactorEventProcessor {
    respBld := &respValueBuilder{ t: t.t }
    rct := NewResponseReactor( respBld )
    t.chkObj = respBld
    return mgRct.InitReactorPipeline( rct )
}

func ( t *reactorTestCall ) initTest() mgRct.ReactorEventProcessor {
    switch typ := t.t.Type; {
    case typ.Equals( QnameRequest ): return t.initRequestTest()
    case typ.Equals( QnameResponse ): return t.initResponseTest()
    }
    panic( libErrorf( "unhandled expect type: %s", t.t.Type ) )
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

func ( t *reactorTestCall ) call() {
    rct := t.initTest()
    if t.feedSource( rct ) { t.chkObj.checkValue( t ) }
}

func ( t *ReactorTest ) Call( c *mgRct.ReactorTestCall ) {
    ( &reactorTestCall{ ReactorTestCall: c, t: t } ).call()
}
