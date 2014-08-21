package service

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "bitgirder/objpath"
//    "bitgirder/stub"
//    "log"
)

type tckEndpointHandlerFuncInput struct {
    auth mg.Value
    params *mg.SymbolMap
    ctx EndpointCallContext
}

type tckEndpointHandlerFunc func( tckEndpointHandlerFuncInput ) error

func getFixedInt( in tckEndpointHandlerFuncInput ) error {
    resp := func( rct mgRct.ReactorEventProcessor ) error {
        return rct.ProcessEvent( mgRct.NewValueEvent( mg.Int32( 1 ) ) )
    }
    return in.ctx.SendResult( resp )
}

func tckSvc1Handlers() *mg.IdentifierMap {
    res := mg.NewIdentifierMap()
    res.Put( mkId( "get-fixed-int" ), tckEndpointHandlerFunc( getFixedInt ) )
    return res
}

type tckEndpoint struct {
    instMap *InstanceMap
}

func ( e *tckEndpoint ) CreateCallId() EndpointCallId {
    return RandomEndpointCallId()
}

type tckEndpointCallHandler struct {
    e *tckEndpoint
    hf tckEndpointHandlerFunc
    paramsBld *mgRct.BuildReactor
    authBld *mgRct.BuildReactor
}

func ( h *tckEndpointCallHandler ) StartRequest(
    ctx *RequestContext, path objpath.PathNode ) error {

    if hf, err := h.e.instMap.getRequestValue( ctx, path ); err == nil {
        h.hf = hf.( tckEndpointHandlerFunc )
    } else { return err }
    return nil
}

func ( h *tckEndpointCallHandler ) setBuilder( 
    addr **mgRct.BuildReactor ) ( mgRct.ReactorEventProcessor, error ) {

    *addr = mgRct.NewBuildReactor( mgRct.ValueBuilderFactory )
    return *addr, nil
}

func ( h *tckEndpointCallHandler ) StartAuthentication(
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    return h.setBuilder( &( h.authBld ) )
}

func ( h *tckEndpointCallHandler ) StartParameters(
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    return h.setBuilder( &( h.paramsBld ) )
}

func ( h *tckEndpointCallHandler ) RequestReactorInterface(
    ctx EndpointCallContext ) RequestReactorInterface { 
    
    return h 
}

func ( h *tckEndpointCallHandler ) optVal( b *mgRct.BuildReactor ) mg.Value {
    if b == nil { return nil }
    return b.GetValue().( mg.Value )
}

func ( h *tckEndpointCallHandler ) Respond( ctx EndpointCallContext ) error {
    in := tckEndpointHandlerFuncInput{ auth: h.optVal( h.authBld ), ctx: ctx }
    if v := h.optVal( h.paramsBld ); v != nil { 
        in.params = v.( *mg.SymbolMap )
    }
    return h.hf( in )
}

func ( e *tckEndpoint ) CreateHandler( 
    ctx EndpointCallContext ) ( EndpointCallHandler, error ) {

    return &tckEndpointCallHandler{ e: e }, nil
}

func NewTckEndpoint() Endpoint {
    res := &tckEndpoint{ instMap: NewInstanceMap() }
    res.instMap.Put( 
        mkNs( "mingle:tck@v1" ), mkId( "svc1" ), tckSvc1Handlers() )
    return res
}
