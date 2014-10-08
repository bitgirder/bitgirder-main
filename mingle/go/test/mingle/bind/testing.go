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

type bindTestCall struct {
    t *BindTest
    iface BindTestCallInterface
    *assert.PathAsserter
    reg *Registry
}

func ( t *bindTestCall ) boundVal() interface{} {
    if val, ok := t.iface.BoundValues().GetOk( t.t.BoundId ); ok { return val }
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
    rcts := append( t.iface.CreateReactors( t.t ), br )
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
//    pip := mgRct.InitReactorPipeline( vb )
    pip := mgRct.InitReactorPipeline( mgRct.NewDebugReactor( t ), vb )
    bc := NewBindContext( t.reg )
    if o := t.t.SerialOptions; o != nil { bc.SerialOptions = o }
    vc := VisitContext{ BindContext: bc, Destination: pip }
    if err := VisitValue( t.boundVal(), vc ); err != nil {
        t.EqualErrors( t.t.Error, err )
        return false
    }
    act := vb.GetValue().( mg.Value )
    mg.AssertEqualValues( t.t.Mingle, act, t.PathAsserter )
    return true
}

func ( t *bindTestCall ) call() {
//    t.debug()
    t.reg = MustRegistryForDomain( t.t.Domain )
    if t.t.BoundId != nil && t.t.Direction.Includes( BindTestDirectionOut ) {
        if ok := t.visitBindTest(); ! ok { return }
    }
    if t.t.Direction.Includes( BindTestDirectionIn ) { t.bindBindTest() }
}

func AssertBindTests( 
    tests []*BindTest, iface BindTestCallInterface, a *assert.PathAsserter ) {

    la := a.StartList()
    for _, test := range tests {
        ( &bindTestCall{ t: test, iface: iface, PathAsserter: la } ).call()
        la = la.Next()
    }
}
