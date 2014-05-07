package reactor

import (
    "testing"
    "bitgirder/assert"
)

func TestReactors( t *testing.T ) {
    a := assert.NewPathAsserter( t )
    la := a.StartList();
    for _, rt := range getReactorTests() {
        ta := la
        if nt, ok := rt.( NamedReactorTest ); ok { 
            ta = a.Descend( nt.TestName() ) 
        }
        c := &ReactorTestCall{ PathAsserter: ta }
        c.Logf( "calling %T", rt )
        rt.Call( c )
        la = la.Next()
    }
}
