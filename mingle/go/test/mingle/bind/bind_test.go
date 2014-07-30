package bind

import (
    "testing"
    "bitgirder/assert"
    mgRct "mingle/reactor"
)

func callBindTest( t *BindTest, a *assert.PathAsserter ) {
    reg := BindRegistryForDomain( t.Domain )
    br := mgRct.NewBuildReactor( NewBindBuilderFactory( reg ) )
    pip := mgRct.InitReactorPipeline( br )
    if err := mgRct.VisitValue( t.In, pip ); err == nil {
        a.Equal( t.Expect, br.GetValue() )
    } else {
        a.EqualErrors( t.Error, err )
    }
}

func TestBinding( t *testing.T ) {
    la := assert.NewListPathAsserter( t )
    for _, sbt := range stdBindTests {
        callBindTest( sbt.( *BindTest ), la )
        la = la.Next()
    }
}
