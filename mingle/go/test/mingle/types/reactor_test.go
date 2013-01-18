package types

import ( 
    "testing"
    "bitgirder/assert"
)

func TestReactors( t *testing.T ) {
    a := assert.NewListPathAsserter( t )
    for _, rt := range StdReactorTests {
        a.Logf( "rt: %v", rt )
        a = a.Next()
    }
}
