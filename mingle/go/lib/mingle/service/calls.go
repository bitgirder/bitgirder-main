package service

import (
    mg "mingle"
    mgRct "mingle/reactor"
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

    SendRequest( out mgRct.ReactorEventProcessor ) error

    ResponseReactorInterface() ResponseReactorInterface
}

type Client interface {

    ExecuteCall( cci ClientCallInterface ) error
}

type RequestSend struct {
    Context *RequestContext
    Parameters *mg.SymbolMap
    Authentication mg.Value
    Destination mgRct.ReactorEventProcessor
}

func ( rs *RequestSend ) Send() error {
    pairs := append( make( []interface{}, 0, 8 ),
        IdNamespace, rs.Context.Namespace.ExternalForm(),
        IdService, rs.Context.Service.ExternalForm(),
        IdOperation, rs.Context.Operation.ExternalForm(),
    )
    if p := rs.Parameters; p != nil { pairs = append( pairs, IdParameters, p ) }
    if a := rs.Authentication; a != nil { 
        pairs = append( pairs, IdAuthentication, a )
    }
    req := mg.MustStruct( QnameRequest, pairs... )
    return mgRct.VisitValue( req, rs.Destination )
}
