package bind

import (
    mg "mingle"
    "mingle/parser"
    mgRct "mingle/reactor"
    "bitgirder/assert"
)

var mkId = parser.MustIdentifier
var mkQn = parser.MustQualifiedTypeName
var asType = parser.AsTypeReference

func bindBindTest( t *BindTest, reg *Registry, a *assert.PathAsserter ) {
    br := NewBuildReactor( NewBuilderFactory( reg ) )
    pip := mgRct.InitReactorPipeline( br )
    if err := mgRct.VisitValue( t.Mingle, pip ); err == nil {
        a.Equal( t.Bound, br.GetValue() )
    } else {
        a.EqualErrors( t.Error, err )
    }
}

func visitBindTest( t *BindTest, reg *Registry, a *assert.PathAsserter ) bool {
    vb := mgRct.NewBuildReactor( mgRct.ValueBuilderFactory )
    pip := mgRct.InitReactorPipeline( vb )
    bc := NewBindContext( reg )
    if err := VisitValue( t.Bound, pip, bc, nil ); err != nil {
        a.EqualErrors( t.Error, err )
        return false
    }
    mg.AssertEqualValues( t.Mingle, vb.GetValue().( mg.Value ), a )
    return true
}

func callBindTest( t *BindTest, a *assert.PathAsserter ) {
    reg := RegistryForDomain( t.Domain )
    if t.Bound != nil && t.Direction.Includes( BindTestDirectionOut ) {
        if ok := visitBindTest( t, reg, a ); ! ok { return }
    }
    if t.Direction.Includes( BindTestDirectionIn ) { bindBindTest( t, reg, a ) }
}

func AssertBindTests( tests []*BindTest, a *assert.PathAsserter ) {
    la := a.StartList()
    for _, test := range tests {
        callBindTest( test, la )
        la = la.Next()
    }
}
