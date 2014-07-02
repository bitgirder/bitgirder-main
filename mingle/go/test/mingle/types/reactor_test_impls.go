package types

import ( 
    mg "mingle"
    mgRct "mingle/reactor"
)

func ( t *CastReactorTest ) Call( c *mgRct.ReactorTestCall ) {
    rcts := []interface{}{}
    if p := t.Path; p != nil {
        rcts = append( rcts, mgRct.NewPathSettingProcessorPath( p ) )
    }
//    rcts = append( rcts, mgRct.NewDebugReactor( c ) )
    rcts = append( rcts, NewCastReactor( t.Type, t.Map ) )
    vb := mgRct.NewValueBuilder()
    rcts = append( rcts, vb )
    pip := mgRct.InitReactorPipeline( rcts... )
//    c.Logf( "casting as %s: %s", t.Type, mg.QuoteValue( t.In ) )
    if err := mgRct.VisitValue( t.In, pip ); err == nil {
        mgRct.CheckNoError( t.Err, c )
//        c.Logf( "got %s", mg.QuoteValue( vb.GetValue() ) )
        mg.AssertEqualValues( t.Expect, vb.GetValue(), c )
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
