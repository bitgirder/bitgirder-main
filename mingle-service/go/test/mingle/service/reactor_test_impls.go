package service

import (
    mg "mingle"
    "mingle/types"
    mgRct "mingle/reactor"
    "bitgirder/objpath"
)

func newTestOperationMap() *OperationMap {
    res := NewOperationMap( getTestTypeDefs() )
    add := func( svc, qnStr string ) {
        qn := mkQn( qnStr )
        res.MustAddServiceInstance( qn.Namespace, mkId( svc ), qn )
    }
    add( "svc1", "mingle:tck:service@v1/Service1" )
    add( "svc2", "mingle:tck:service@v1/Service2" )
    add( "svc1", "mingle:service:fail@v1/Service1" )
    add( "svc1", "mingle:service@v1/Service1" )
    return res
}

type reactorTestChecker interface {
    checkValue( t *reactorTestCall )
}

func setBuilder(
    addr **mgRct.BuildReactor ) ( mgRct.EventProcessor, error ) {

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
    if b.ctx.Namespace.Equals( mkNs( "mingle:service:fail@v1" ) ) {
        return &testError{ path, "start-request-impl-error" }
    }
    return nil
}

func ( b *baseReqBuilder ) StartAuthentication( 
    path objpath.PathNode ) ( mgRct.EventProcessor, error ) {

    if b.ctx.Operation.Equals( mkId( "failStartAuthentication" ) ) {
        return nil, &testError{ path, "start-authentication-impl-error" }
    }
    return setBuilder( &b.authBldr )
}

func ( b *baseReqBuilder ) StartParameters(
    path objpath.PathNode ) ( mgRct.EventProcessor, error ) {

    if b.ctx.Operation.Equals( mkId( "failStartParameters" ) ) {
        return nil, &testError{ path, "start-parameters-impl-error" }
    }
    return setBuilder( &b.paramsBldr )
}

func ( b *baseReqBuilder ) checkValue( t *reactorTestCall ) {
    expct := t.t.Expect.( *requestExpect )
    qt := func( val interface{} ) string {
        switch v := val.( type ) {
        case nil: return "nil"
        case *mgRct.BuildReactor: 
            if v == nil { return "<nil>" }
            return mg.QuoteValue( v.GetValue().( mg.Value ) )
        case mg.Value: return mg.QuoteValue( v )
        }
        panic( libErrorf( "unhandled val: %T", val ) )
    }
    t.Logf( "got auth=%s, params=%s, expecting auth=%s, params=%s",
        qt( b.authBldr ), qt( b.paramsBldr ), qt( expct.auth ),
        qt( expct.params ) )
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
    path objpath.PathNode ) ( mgRct.EventProcessor, error ) {

    if err := b.implError( path ); err != nil { return nil, err }
    return setBuilder( &b.resBldr )
}

func ( b *respValueBuilder ) StartError( 
    path objpath.PathNode ) ( mgRct.EventProcessor, error ) {

    if err := b.implError( path ); err != nil { return nil, err }
    return setBuilder( &b.errBldr )
}

func ( b *respValueBuilder ) checkValue( t *reactorTestCall ) {
    expct := t.t.Expect.( *ResultExpectation )
    mgRct.CheckBuiltValue( expct.Result, b.resBldr, t.Descend( "result" ) )
    mgRct.CheckBuiltValue( expct.Error, b.errBldr, t.Descend( "error" ) )
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
    t.chkObj = res
    return AsTypedRequestReactorInterface( res, newTestOperationMap() )
}

func ( t *reactorTestCall ) initRequestTest() mgRct.EventProcessor {
    var reqIface RequestReactorInterface
    switch t.t.ReactorProfile {
    case ReactorProfileBase: reqIface = t.initBaseRequestTest()
    case ReactorProfileTyped: reqIface = t.initTypedRequestTest()
    default: t.Fatalf( "unhandled profile: %s", t.t.ReactorProfile )
    }
    rct := NewRequestReactor( reqIface )
    return mgRct.InitReactorPipeline( rct )
}

func ( t *reactorTestCall ) newRespValueBuilder() *respValueBuilder {
    return &respValueBuilder{ t: t.t }
}

func ( t *reactorTestCall ) initBaseResponseTest() ResponseReactorInterface {
    res := t.newRespValueBuilder()
    t.chkObj = res
    return res
}

func ( t *reactorTestCall ) initTypedResponseTest() ResponseReactorInterface {
    m := newTestOperationMap()
    ri := t.t.In.( *responseInput )
    reqDef, err := m.ExpectOperationForRequest( ri.reqCtx, nil )
    if err != nil { panic( err ) }
    bldr := t.newRespValueBuilder()
    t.chkObj = bldr
    opSig := reqDef.Operation.Signature
    retTyp := opSig.Return
    errTyps := make( []*types.UnionTypeDefinition, 0, 2 )
    addErrTyp := func( sig *types.CallSignature ) {
        if ut := sig.Throws; ut != nil { errTyps = append( errTyps, ut ) }
    }
    addErrTyp( opSig )
    if secQn := reqDef.Service.Security; secQn != nil {
        addErrTyp( types.MustPrototypeDefinition( secQn, m.defs ).Signature )
    }
    return AsTypedResponseReactorInterface( bldr, retTyp, errTyps, m.defs )
}

func ( t *reactorTestCall ) initResponseTest() mgRct.EventProcessor {
    var respIface ResponseReactorInterface
    switch t.t.ReactorProfile {
    case ReactorProfileBase: respIface = t.initBaseResponseTest()
    case ReactorProfileTyped: respIface = t.initTypedResponseTest()
    default: t.Fatalf( "unhandled profile: %s", t.t.ReactorProfile )
    }
    rct := NewResponseReactor( respIface )
    return mgRct.InitReactorPipeline( rct )
}

func ( t *reactorTestCall ) initTest() mgRct.EventProcessor {
    switch typ := t.t.Type; {
    case typ.Equals( QnameRequest ): return t.initRequestTest()
    case typ.Equals( QnameResponse ): return t.initResponseTest()
    }
    panic( libErrorf( "unhandled expect type: %s", t.t.Type ) )
}

func ( t *reactorTestCall ) getFeedSource() interface{} {
    switch v := t.t.In.( type ) {
    case *responseInput: return v.in
    }
    return t.t.In
}

func ( t *reactorTestCall ) feedSource( rct mgRct.EventProcessor ) bool {
    src := t.getFeedSource()
    if mv, ok := src.( mg.Value ); ok {
        t.Logf( "feeding %s", mg.QuoteValue( mv ) )
    }
    err := mgRct.FeedSource( src, rct )
    if err == nil && t.t.Expect != nil { return true }
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
