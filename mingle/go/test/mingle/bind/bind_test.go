package bind

import (
    "testing"
    "bitgirder/assert"
    "bitgirder/objpath"
    mg "mingle"
    mgRct "mingle/reactor"
)

// one-time guard for ensureTestBuilderFactories()
var didEnsureTestBuilderFactories = false

// we would otherwise do this in an init() block, except we don't want to deal
// with the possibility that this would run before the default domain itself is
// initialized (dependent packages won't have this concern)
func ensureTestBuilderFactories() {
    if didEnsureTestBuilderFactories { return }
    didEnsureTestBuilderFactories = true
    reg := BindRegistryForDomain( DomainDefault )
    reg.MustAddValue(
        mkQn( "ns1@v1/S1" ),
        func() mgRct.BuilderFactory {
            res := mgRct.NewFunctionsBuilderFactory()
            res.StructFunc = 
                func( _ *mgRct.StructStartEvent ) ( mgRct.FieldSetBuilder, 
                                                    error ) {

                    res := mgRct.NewFunctionsFieldSetBuilder()
                    res.Value = new( S1 )
                    res.FinalValue = func() interface{} {
                        return *( res.Value.( *S1 ) )
                    }
                    res.RegisterField(
                        mkId( "f1" ),
                        func( path objpath.PathNode ) ( mgRct.BuilderFactory, 
                                                        error ) {
                            res, ok := reg.m.GetOk( mg.QnameInt32 )
                            if ok { return res.( mgRct.BuilderFactory ), nil }
                            return nil, nil
                        },
                        func( val interface{}, path objpath.PathNode ) error {
                            res.Value.( *S1 ).f1 = val.( int32 )
                            return nil
                        },
                    )
                    return res, nil
                }
            return res
        }(),
    )
    reg.MustAddValue(
        mkQn( "ns1@v1/E1" ),
        func() mgRct.BuilderFactory {
            res := mgRct.NewFunctionsBuilderFactory()
            res.ValueFunc = func( 
                ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {

                if e, ok := ve.Val.( *mg.Enum ); ok {
                    return E1( e.Value.ExternalForm() ), nil, true
                }
                return nil, nil, false
            }
            return res
        }(),
    )
}

func callBindTest( t *BindTest, a *assert.PathAsserter ) {
    a.Logf( "visiting %s", mg.QuoteValue( t.In ) )
    ensureTestBuilderFactories()
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
