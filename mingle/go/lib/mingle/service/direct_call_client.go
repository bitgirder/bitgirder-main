package service

import (
)

type directCallClient struct {
    Endpoint Endpoint
}

func ( c *directCallClient ) ExecuteCall( cci ClientCallInterface ) error {
    id := c.Endpoint.CreateCallId()
    reqIface, err := c.Endpoint.StartRequest( id )
    if err != nil { return err }
    rct := NewRequestReactor( reqIface )
    return cci.SendRequest( rct )
}

func NewDirectCallClient( ep Endpoint ) *directCallClient {
    return &directCallClient{ Endpoint: ep }
}
