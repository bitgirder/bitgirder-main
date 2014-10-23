package testing

import ( 
    mg "mingle"
    "mingle/types"
    "mingle/types/builtin"
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
    return nil, mg.NewCastError( path, "custom-field-set-test-error" )
}

func matchUnion2Type( in types.UnionMatchInput ) ( mg.TypeReference, bool ) {
    if _, ok := in.TypeIn.( *mg.ListTypeReference ); ok {
        return asType( "String*" ), true
    }
    switch {
    case in.TypeIn.Equals( mg.TypeInt64 ): return mg.TypeUint64, true
    }
    return nil, false
}

func formatTypeErrorCustom( 
    expct, act mg.TypeReference, path objpath.PathNode ) ( error, bool ) {

    if expct.Equals( mg.TypeBuffer ) && act.Equals( mg.TypeInt32 ) {
        return mg.NewCastError( path, "bad-int32-for-buffer" ), true
    }
    return nil, false
}

func addCustomErrorFormatting( cr *types.CastReactor ) {
    cr.FormatTypeError = formatTypeErrorCustom
}

func ( t *CastReactorTest ) addCastReactor( 
    rcts []interface{}, c *mgRct.ReactorTestCall ) []interface{} {

    cr := types.NewCastReactor( t.Type, t.Map )
    builtin.CastBuiltinTypes( cr )
    switch t.Profile {
    case ProfileCastDisable: 
        cr.AddPassthroughField( mkQn( "ns1@v1/S1" ), mkId( "f1" ) )
        cr.AddPassthroughField( mkQn( "ns1@v1/Schema1" ), mkId( "f1" ) )
    case ProfileCustomFieldSet:
        dt := mgRct.NewDepthTracker()
        rcts = append( rcts, dt )
        cr.FieldSetFactory = customFieldSetFactory{ dt, c }
    case ProfileUnionImpl:
        cr.SetUnionDefinitionMatcher(
            mkQn( "ns1@v1/Union2" ), matchUnion2Type )
    case ProfileCustomErrorFormatting: addCustomErrorFormatting( cr )
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
    } else { c.EqualErrors( t.Err, err ) }
}

func ( t *EventPathTest ) Call( c *mgRct.ReactorTestCall ) {
    chk := mgRct.NewEventPathCheckReactor( t.Expect, c.PathAsserter )
    rct := types.NewCastReactor( t.Type, t.Map )
    pip := mgRct.InitReactorPipeline( rct, chk )
    mgRct.AssertFeedSource( t.Source, pip, c )
    chk.Complete()
}
