package service

import (
    mg "mingle"
    mgRct "mingle/reactor"
)

type directCallClient struct {
    Endpoint Endpoint
}

type directCallEndpointContext struct {
    callId EndpointCallId
    cci ClientCallInterface
}

func ( ctx *directCallEndpointContext ) CallId() EndpointCallId {
    return ctx.callId
}

func ( ctx *directCallEndpointContext ) sendResult( 
    id *mg.Identifier, f ReactorUserFunc ) error {

    rct := InitResponseReactorPipeline( ctx.cci.ResponseReactorInterface() )
    ss := mgRct.NewStructStartEvent( QnameResponse )
    if err := rct.ProcessEvent( ss ); err != nil { return err }
    fs := mgRct.NewFieldStartEvent( id )
    if err := rct.ProcessEvent( fs ); err != nil { return err }
    if err := f( rct ); err != nil { return err }
    return rct.ProcessEvent( mgRct.NewEndEvent() )
}

func ( ctx *directCallEndpointContext ) SendResult( f ReactorUserFunc ) error {
    return ctx.sendResult( IdResult, f )
}

func ( c *directCallClient ) ExecuteCall( cci ClientCallInterface ) error {
    ctx := &directCallEndpointContext{ 
        callId: RandomEndpointCallId(), 
        cci: cci,
    }
    ch, err := c.Endpoint.CreateHandler( ctx )
    if err != nil { return err }
    reqIface := ch.RequestReactorInterface( ctx )
    rct := InitRequestReactorPipeline( reqIface )
    if err := cci.SendRequest( rct ); err != nil { return err }
    return ch.Respond( ctx )
}

func NewDirectCallClient( ep Endpoint ) *directCallClient {
    return &directCallClient{ Endpoint: ep }
}
