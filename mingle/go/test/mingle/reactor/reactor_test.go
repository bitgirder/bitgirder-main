package reactor

import (
    "testing"
)

func TestReactors( t *testing.T ) {
    RunReactorTestsInNamespace( reactorTestNs, t )
}
