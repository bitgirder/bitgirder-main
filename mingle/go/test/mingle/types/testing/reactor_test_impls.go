package testing

import ( 
    mg "mingle"
    "mingle/types"
    mgRct "mingle/reactor"
    "bitgirder/objpath"
)

type customFieldSetFactory struct {
    dt *mgRct.DepthTracker
    c *mgRct.ReactorTestCall
}

func ( f customFieldSetFactory ) GetFieldSet( 
    path objpath.PathNode ) ( *types.FieldSet, error ) {

    switch d := f.dt.Depth(); {
    case d == 3: return nil, nil
    case d < 4:
        fs := types.MakeFieldSet(
            types.MakeFieldDef( "f1", "String?", nil ),
            types.MakeFieldDef( "f2", "SymbolMap?", nil ),
        )
        return fs, nil
    }
    return nil, mg.NewValueCastError( path, "custom-field-set-test-error" )
}

func ( t *CastReactorTest ) addCastReactor( 
    rcts []interface{}, c *mgRct.ReactorTestCall ) []interface{} {

    cr := types.NewCastReactor( t.Type, t.Map )
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
    rct := types.NewCastReactor( t.Type, t.Map )
    pip := mgRct.InitReactorPipeline( rct, chk )
    mgRct.AssertFeedSource( t.Source, pip, c )
    chk.Complete()
}