package service

import (
    "testing"
    mgRct "mingle/reactor"
)

func TestReactors( t *testing.T ) {
    mgRct.RunReactorTestsInNamespace( NsService, t )
}
