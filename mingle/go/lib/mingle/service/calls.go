package service

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "mingle/bind"
    "mingle/types/builtin"
    "bitgirder/uuid"
//    "log"
)

type EndpointCallId string

func RandomEndpointCallId() EndpointCallId {
    return EndpointCallId( uuid.MustType4() )
}

type EndpointCallContext interface {

    CallId() EndpointCallId

    SendResult( f ReactorUserFunc ) error
}

type EndpointCallHandler interface {

    RequestReactorInterface( ctx EndpointCallContext ) RequestReactorInterface

    Respond( ctx EndpointCallContext ) error
}

type Endpoint interface {
    
    CreateHandler( ctx EndpointCallContext ) ( EndpointCallHandler, error )
}

type ClientCallInterface interface {

    SendRequest( out mgRct.EventProcessor ) error

    ResponseReactorInterface() ResponseReactorInterface
}

type Client interface {

    ExecuteCall( cci ClientCallInterface ) error
}

type RequestSend struct {
    Context *RequestContext
    Parameters *mg.SymbolMap
    Authentication mg.Value
    Destination mgRct.EventProcessor
}

func ( rs *RequestSend ) startRequestSend( 
    vc bind.VisitContext ) ( err error ) {
    
    es := vc.EventSender()
    if err = es.StartStruct( QnameRequest ); err != nil { return }
    if err = es.StartField( IdNamespace ); err != nil { return }
    rc := rs.Context
    if err = builtin.VisitNamespace( rc.Namespace, vc ); err != nil { return }
    if err = es.StartField( IdService ); err != nil { return }
    if err = builtin.VisitIdentifier( rc.Service, vc ); err != nil { return }
    if err = es.StartField( IdOperation ); err != nil { return }
    if err = builtin.VisitIdentifier( rc.Operation, vc ); err != nil { return }
    return
}

func ( rs *RequestSend ) Send() ( err error ) {
    reg := bind.MustRegistryForDomain( bind.DomainDefault )
    vc := bind.VisitContext{
        BindContext: bind.NewBindContext( reg ),
        Destination: rs.Destination,
    }
    if err = rs.startRequestSend( vc ); err != nil { return }
    es := vc.EventSender()
    if a := rs.Authentication; a != nil { 
        if err = es.StartField( IdAuthentication ); err != nil { return }
        if err = mgRct.VisitValue( a, vc.Destination ); err != nil { return }
    }
    if p := rs.Parameters; p != nil { 
        if err = es.StartField( IdParameters ); err != nil { return }
        if err = mgRct.VisitValue( p, vc.Destination ); err != nil { return }
    }
    return es.End()
}
