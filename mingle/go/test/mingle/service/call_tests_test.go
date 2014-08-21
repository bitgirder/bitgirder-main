package service

import (
    "bitgirder/assert"
    "bitgirder/objpath"
    mg "mingle"
    mgRct "mingle/reactor"
    "testing"
)

type tckTest struct {
    ct *TckTestCall
    *assert.PathAsserter
    cli Client
    errBld *mgRct.BuildReactor
    resBld *mgRct.BuildReactor
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

func ( t *tckTest ) setBuilder( 
    addr **mgRct.BuildReactor ) ( mgRct.ReactorEventProcessor, error ) {

    *addr = mgRct.NewBuildReactor( mgRct.ValueBuilderFactory )
    return *addr, nil
}

func ( t *tckTest ) StartError( 
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    return t.setBuilder( &( t.errBld ) )
}

func ( t *tckTest ) StartResult(
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    return t.setBuilder( &( t.resBld ) )
}

func ( t *tckTest ) ResponseReactorInterface() ResponseReactorInterface {
    return t
}

func ( t *tckTest ) equalBuiltValue(
    expct mg.Value, actBld *mgRct.BuildReactor, a *assert.PathAsserter ) {

    var act mg.Value
    if actBld != nil { act = actBld.GetValue().( mg.Value ) }
    mg.AssertEqualValues( expct, act, a )
}

func ( t *tckTest ) checkResult() {
    t.equalBuiltValue( t.ct.Expect.Result, t.resBld, t.Descend( "Result" ) )
    t.equalBuiltValue( t.ct.Expect.Error, t.errBld, t.Descend( "Error" ) )
}

func ( t *tckTest ) call() {
    if err := t.cli.ExecuteCall( t ); err != nil { t.Fatal( err ) }
    t.checkResult()
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
