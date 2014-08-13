package types

import ( 
    mg "mingle"
    "mingle/bind"
    mgRct "mingle/reactor"
    "bitgirder/objpath"
)

type customFieldSetFactory struct {
    dt *mgRct.DepthTracker
    c *mgRct.ReactorTestCall
}

func ( f customFieldSetFactory ) GetFieldSet( 
    path objpath.PathNode ) ( *FieldSet, error ) {

    f.c.Logf( "depth is %d at path %s", f.dt.Depth(), mg.FormatIdPath( path ) )
    switch d := f.dt.Depth(); {
    case d == 3: return nil, nil
    case d < 4:
        fs := MakeFieldSet(
            MakeFieldDef( "f1", "String?", nil ),
            MakeFieldDef( "f2", "SymbolMap?", nil ),
        )
        return fs, nil
    }
    return nil, mg.NewValueCastError( path, "custom-field-set-test-error" )
}

func ( t *CastReactorTest ) addCastReactor( 
    rcts []interface{}, c *mgRct.ReactorTestCall ) []interface{} {

    cr := NewCastReactor( t.Type, t.Map )
    switch t.Profile {
    case ProfileCastDisable: 
        cr.AddPassthroughField( mkQn( "ns1@v1/S1" ), mkId( "f1" ) )
        cr.AddPassthroughField( mkQn( "ns1@v1/Schema1" ), mkId( "f1" ) )
    case ProfileCustomFieldSet:
        dt := mgRct.NewDepthTracker()
        rcts = append( rcts, dt )
        cr.FieldSetFactory = customFieldSetFactory{ dt, c }
    }
    return append( rcts, cr )
}

func ( t *CastReactorTest ) Call( c *mgRct.ReactorTestCall ) {
    rcts := []interface{}{}
    if p := t.Path; p != nil {
        rcts = append( rcts, mgRct.NewPathSettingProcessorPath( p ) )
    }
//    rcts = append( rcts, mgRct.NewDebugReactor( c ) )
    rcts = t.addCastReactor( rcts, c )
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
    cr := NewCastReactor( t.Type, BuiltinTypes() )
    pip := mgRct.InitReactorPipeline( cr, mgRct.NewDebugReactor( c ), br )
    if err := mgRct.VisitValue( t.In, pip ); err == nil {
        c.Equal( t.Expect, br.GetValue() )
    } else { c.EqualErrors( t.Err, err ) }
}
