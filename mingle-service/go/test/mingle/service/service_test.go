package service

import (
    "testing"
    mg "mingle"
    mgRct "mingle/reactor"
    "bitgirder/assert"
)

func TestReactors( t *testing.T ) {
    mgRct.RunReactorTests( GetReactorTests(), assert.NewPathAsserter( t ) )
}

func TestFormatInstanceId( t *testing.T ) {
    a := assert.Asserter{ t }
    a.Equal( 
        "ns1@v1.svc1", FormatInstanceId( mkNs( "ns1@v1" ), mkId( "svc1" ) ) )
}

func TestInstanceMap( t *testing.T ) {
    a := assert.NewPathAsserter( t )
    ns1, ns2 := mkNs( "ns1@v1" ), mkNs( "ns2@v1" )
    svc1, svc2 := mkId( "svc1" ), mkId( "svc2" )
    m := NewInstanceMap()
    chk := func( 
        ns *mg.Namespace, 
        svc *mg.Identifier, 
        expct interface{}, 
        miss *mg.Identifier ) {

        act, missAct := m.GetOk( ns, svc )
        ta := a.Descend( FormatInstanceId( ns, svc ) )
        ta.Equal( miss, missAct )
        if miss == nil { ta.Equal( expct.( int ), act.( int ) ) }
    }
    chk( ns1, svc1, nil, IdNamespace )
    m.Put( ns1, svc1, 1 )
    chk( ns1, svc1, 1, nil )
    chk( ns1, svc2, nil, IdService )
    chk( ns2, svc1, nil, IdNamespace )
}
