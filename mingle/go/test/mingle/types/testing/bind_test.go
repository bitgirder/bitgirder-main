package testing

import ( 
    "testing"
    "bitgirder/assert"
    "mingle/bind"
)

func getBindTests() []*bind.BindTest {
    res := []*bind.BindTest{}
    return res
}

func TestBind( t *testing.T ) {
    bind.AssertBindTests( getBindTests(), assert.NewPathAsserter( t ) )
}
