package service

import (
    mgRct "mingle/reactor"
    "bitgirder/objpath"
    "bitgirder/stub"
)

type tckEndpoint struct {
    instMap *InstanceMap
}

func ( e *tckEndpoint ) CreateCallId() EndpointCallId {
    return RandomEndpointCallId()
}

type tckEndpointCallHandler struct {
    e *tckEndpoint
    callId EndpointCallId
}

func ( h *tckEndpointCallHandler ) StartRequest(
    ctx *RequestContext, path objpath.PathNode ) error {

    if _, err := h.e.instMap.getRequestValue( ctx, path ); err != nil {
        return err
    }
    return stub.Unimplemented()
}

func ( h *tckEndpointCallHandler ) StartAuthentication(
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    return nil, stub.Unimplemented()
}

func ( h *tckEndpointCallHandler ) StartParameters(
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    return nil, stub.Unimplemented()
}

func ( e *tckEndpoint ) StartRequest( 
    id EndpointCallId ) ( RequestReactorInterface, error ) {

    return &tckEndpointCallHandler{ e: e, callId: id }, nil
}

func NewTckEndpoint() Endpoint {
    res := &tckEndpoint{ instMap: NewInstanceMap() }
    return res
}
