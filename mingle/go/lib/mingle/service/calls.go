package service

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "bitgirder/uuid"
)

type EndpointCallId string

func RandomEndpointCallId() EndpointCallId {
    return EndpointCallId( uuid.MustType4() )
}

type Endpoint interface {
    
    CreateCallId() EndpointCallId

    StartRequest( id EndpointCallId ) ( RequestReactorInterface, error )
}

type ClientCallInterface interface {

    SendRequest( out mgRct.ReactorEventProcessor ) error

    StartResponse() ( ResponseReactorInterface, error )
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
        IdNamespace, rs.Context.Namespace,
        IdService, rs.Context.Service,
        IdOperation, rs.Context.Operation,
    )
    if p := rs.Parameters; p != nil { pairs = append( pairs, IdParameters, p ) }
    if a := rs.Authentication; a != nil { 
        pairs = append( pairs, IdAuthentication, a )
    }
    req := mg.MustStruct( QnameRequest, pairs... )
    return mgRct.VisitValue( req, rs.Destination )
}
