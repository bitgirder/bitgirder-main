package testing

import (
    "testing"
    mgRct "mingle/reactor"
    "bitgirder/assert"
)

func TestReactors( t *testing.T ) {
    mgRct.RunReactorTests( GetReactorTests(), assert.NewPathAsserter( t ) )
}
