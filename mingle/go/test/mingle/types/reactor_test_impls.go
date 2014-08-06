package types

import ( 
    mg "mingle"
    "mingle/bind"
    mgRct "mingle/reactor"
)

func ( t *CastReactorTest ) newCastReactor() *CastReactor {
    res := NewCastReactor( t.Type, t.Map )
    switch t.Profile {
    case ProfileCastDisable: 
        res.AddPassthroughField( mkQn( "ns1@v1/S1" ), mkId( "f1" ) )
        res.AddPassthroughField( mkQn( "ns1@v1/Schema1" ), mkId( "f1" ) )
    }
    return res
}

func ( t *CastReactorTest ) Call( c *mgRct.ReactorTestCall ) {
    rcts := []interface{}{}
    if p := t.Path; p != nil {
        rcts = append( rcts, mgRct.NewPathSettingProcessorPath( p ) )
    }
//    rcts = append( rcts, mgRct.NewDebugReactor( c ) )
    rcts = append( rcts, t.newCastReactor() )
    vb := mgRct.NewBuildReactor( mgRct.ValueBuilderFactory )
    rcts = append( rcts, vb )
    pip := mgRct.InitReactorPipeline( rcts... )
    if inVal, ok := t.In.( mg.Value ); ok {
        c.Logf( "casting as %s: %s", t.Type, mg.QuoteValue( inVal ) )
    }
    if err := mgRct.FeedSource( t.In, pip ); err == nil {
        mgRct.CheckNoError( t.Err, c )
        act := vb.GetValue().( mg.Value )
        c.Logf( "got %s, expect %s", mg.QuoteValue( act ),
            mg.QuoteValue( t.Expect ) )
        mg.AssertEqualValues( t.Expect, act, c )
    } else { 
        cae := mg.CastErrorAssert{ 
            ErrExpect: t.Err, ErrAct: err, PathAsserter: c.PathAsserter }
        cae.Call()
    }
}

func ( t *EventPathTest ) Call( c *mgRct.ReactorTestCall ) {
    chk := mgRct.NewEventPathCheckReactor( t.Expect, c.PathAsserter )
    rct := NewCastReactor( t.Type, t.Map )
    pip := mgRct.InitReactorPipeline( rct, chk )
    mgRct.AssertFeedSource( t.Source, pip, c )
    chk.Complete()
}

func ( t *BuiltinTypeTest ) createBindReactor( 
    c *mgRct.ReactorTestCall ) *mgRct.BuildReactor {

    reg := bind.RegistryForDomain( bind.DomainDefault )
    if bf, ok := reg.BuilderFactoryForType( t.Type ); ok {
        return bind.NewBuildReactor( bf )
    }
    c.Fatalf( "no binder for type: %s", t.Type )
    panic( libError( "unreachable" ) )
}

func ( t *BuiltinTypeTest ) Call( c *mgRct.ReactorTestCall ) {
    c.Logf( "expcting %s as type: %s", mg.QuoteValue( t.In ), t.Type )
    br := t.createBindReactor( c )
    cr := NewCastReactor( t.Type, V1Types() )
    pip := mgRct.InitReactorPipeline( cr, mgRct.NewDebugReactor( c ), br )
    if err := mgRct.VisitValue( t.In, pip ); err == nil {
        c.Equal( t.Expect, br.GetValue() )
    } else { c.EqualErrors( t.Err, err ) }
}
