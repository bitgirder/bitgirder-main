package builtin

import (
    "testing"
    mgRct "mingle/reactor"
)

func TestReactors( t *testing.T ) {
    mgRct.RunReactorTestsInNamespace( reactorTestNs, t )
}
