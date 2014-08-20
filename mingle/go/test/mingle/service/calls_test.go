package service

import (
    "bitgirder/stub"
    "bitgirder/assert"
    mgRct "mingle/reactor"
    "testing"
)

type tckTest struct {
    ct *TckTestCall
    *assert.PathAsserter
    cli Client
}

func ( t *tckTest ) SendRequest( out mgRct.ReactorEventProcessor ) error {
    rs := &RequestSend{
        Context: t.ct.Context,
        Parameters: t.ct.Parameters,
        Authentication: t.ct.Authentication,
        Destination: out,
    }
    return rs.Send()
}

func ( t *tckTest ) StartResponse() ( ResponseReactorInterface, error ) {
    return nil, stub.Unimplemented()
}

func ( t *tckTest ) call() {
    if err := t.cli.ExecuteCall( t ); err != nil { t.Fatal( err ) }
}

func TestTckCallsBaseImpl( t *testing.T ) {
    la := assert.NewListPathAsserter( t )
    for _, ct := range GetTckTestCalls() {
        cte := &tckTest{ ct: ct, PathAsserter: la }
        ep := NewTckEndpoint()
        cte.cli = NewDirectCallClient( ep )
        cte.call()
    }
}
