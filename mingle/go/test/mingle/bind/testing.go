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

type BindTestCallInterface interface {
    CreateReactors( t *BindTest ) []interface{}
    BoundValues() *mg.IdentifierMap
}

type BindTestCallControl struct {
    Interface BindTestCallInterface
    MingleValueForAssert func( v mg.Value ) mg.Value
}

type bindTestCall struct {
    t *BindTest
    cc *BindTestCallControl
    *assert.PathAsserter
    reg *Registry
}

func ( t *bindTestCall ) iface() BindTestCallInterface { return t.cc.Interface }

func ( t *bindTestCall ) boundVal() interface{} {
    if val, ok := t.iface().BoundValues().GetOk( t.t.BoundId ); ok { 
        return val 
    }
    t.Logf( "no bound val for id: %s", t.t.BoundId )
    return nil
}

func ( t *bindTestCall ) debug() {
    qtVal := "(no Mingle value)"
    if mv := t.t.Mingle; mv != nil { qtVal = mg.QuoteValue( mv ) }
    t.Logf( "test.Mingle: %s, test.BoundId: %s, test.Error: %s, test.Direction: %d, test.Domain: %s",
        qtVal, t.t.BoundId, t.t.Error, t.t.Direction, t.t.Domain,
    )
}

func ( t *bindTestCall ) getBuilderFactory() mgRct.BuilderFactory {
    typ := t.t.Type
    if typ == nil { typ = mg.TypeOf( t.t.Mingle ) }
    if bf, ok := t.reg.BuilderFactoryForType( typ ); ok { return bf }
    if t.t.StrictTypeMatching {
        t.Fatalf( "no builder factory for type: %s", typ )
    }
    return NewBuilderFactory( t.reg )
}

func ( t *bindTestCall ) bindBindTest() {
    bf := t.getBuilderFactory()
    br := NewBuildReactor( bf )
    rcts := []interface{}{}
    if p := t.t.StartPath; p != nil {
        rcts = append( rcts, mgRct.NewPathSettingProcessorPath( p ) )
    }
    rcts = append( rcts, t.iface().CreateReactors( t.t )... )
    rcts = append( rcts, br )
    pip := mgRct.InitReactorPipeline( rcts... )
    if err := mgRct.VisitValue( t.t.Mingle, pip ); err == nil {
        t.EqualErrors( t.t.Error, err ) // fine if both nil
        t.Equal( t.boundVal(), br.GetValue() )
    } else {
        t.EqualErrors( t.t.Error, err )
    }
}

func ( t *bindTestCall ) visitBindTest() bool {
    vb := mgRct.NewBuildReactor( mgRct.ValueBuilderFactory )
    pip := mgRct.InitReactorPipeline( vb )
    bc := NewBindContext( t.reg )
    if o := t.t.SerialOptions; o != nil { bc.SerialOptions = o }
    vc := VisitContext{ BindContext: bc, Destination: pip }
    bv := t.boundVal()
    if err := VisitValue( bv, vc ); err != nil {
        t.EqualErrors( t.t.Error, err )
        return false
    }
    act := vb.GetValue().( mg.Value )
    if f := t.cc.MingleValueForAssert; f != nil { act = f( act ) }
    t.Logf( "asserting unbind, act: %s", mg.QuoteValue( act ) )
    mg.AssertEqualValues( t.t.Mingle, act, t.PathAsserter )
    return true
}

func ( t *bindTestCall ) call() {
    t.debug()
    t.reg = MustRegistryForDomain( t.t.Domain )
    if t.t.BoundId != nil && t.t.Direction.Includes( BindTestDirectionOut ) {
        if ok := t.visitBindTest(); ! ok { return }
    }
    if t.t.Direction.Includes( BindTestDirectionIn ) { t.bindBindTest() }
}

func AssertBindTests( 
    tests []*BindTest, cc *BindTestCallControl, a *assert.PathAsserter ) {

    la := a.StartList()
    for _, test := range tests {
        btc := &bindTestCall{ t: test, cc: cc, PathAsserter: la } 
        btc.call()
        la = la.Next()
    }
}
