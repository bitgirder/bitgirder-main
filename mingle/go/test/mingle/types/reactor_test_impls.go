package types

import ( 
    "bitgirder/assert"
    "bitgirder/objpath"
    mg "mingle"
    mgRct "mingle/reactor"
)

func ( t *CastReactorTest ) Call( c *mgRct.ReactorTestCall ) {
    rct := NewCastReactorDefinitionMap( t.Type, t.Map )
    vb := mgRct.NewValueBuilder()
    pip := mgRct.InitReactorPipeline( rct, vb )
    if err := mgRct.VisitValue( t.In, pip ); err == nil {
        mgRct.CheckNoError( t.Err, c )
        mg.EqualValues( t.Expect, vb.GetValue(), c )
    } else { 
        cae := mg.CastErrorAssert{ 
            ErrExpect: t.Err, ErrAct: err, PathAsserter: c.PathAsserter }
        cae.Call()
    }
}

func ( t *EventPathTest ) Call( c *mgRct.ReactorTestCall ) {
    chk := mgRct.NewEventPathCheckReactor( t.Expect, c.PathAsserter )
    rct := NewCastReactorDefinitionMap( t.Type, t.Map )
    pip := mgRct.InitReactorPipeline( rct, chk )
    mgRct.AssertFeedSource( t.Source, pip, c )
    chk.Complete()
}

type checker interface { check() }

func visitAndCheck(
    in mg.Value, 
    rep mgRct.ReactorEventProcessor, 
    chk checker, 
    errExpct error,
    c *mgRct.ReactorTestCall ) {

    if err := mgRct.VisitValue( in, rep ); err == nil {
        mgRct.CheckNoError( errExpct, c )
        chk.check()
    } else { c.EqualErrors( errExpct, err ) }
}

func initValueBuilder( 
    vbPtr **mgRct.ValueBuilder ) ( mgRct.ReactorEventProcessor, error ) {

    *vbPtr = mgRct.NewValueBuilder()
    return *vbPtr, nil
}

type requestCheck struct {
    *assert.PathAsserter
    st *ServiceRequestTest
    auth *mgRct.ValueBuilder
    params *mgRct.ValueBuilder
}

func ( chk *requestCheck ) GetAuthenticationReactor(
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    return initValueBuilder( &( chk.auth ) )
}

func ( chk *requestCheck ) GetParametersReactor(
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    return initValueBuilder( &( chk.params ) )
}

func ( chk *requestCheck ) check() {
    mgRct.CheckBuiltValue( 
        chk.st.Authentication, chk.auth, chk.Descend( "authentication" ) )
    mgRct.CheckBuiltValue( 
        chk.st.Parameters, chk.params, chk.Descend( "parameters" ) )
}

func ( t *ServiceRequestTest ) Call( c *mgRct.ReactorTestCall ) {
    chk := &requestCheck{ st: t, PathAsserter: c.PathAsserter }
    rct := NewRequestReactor( t.Maps.BuildOpMap(), chk )
    pip := mgRct.InitReactorPipeline( rct )
    visitAndCheck( t.In, pip, chk, t.Error, c )
}

type responseCheck struct {
    st *ServiceResponseTest
    *assert.PathAsserter
    resultProc, errorProc *mgRct.ValueBuilder
}

func ( chk *responseCheck ) GetResultReactor( 
    p objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    return initValueBuilder( &( chk.resultProc ) )
}

func ( chk *responseCheck ) GetErrorReactor( 
    p objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    return initValueBuilder( &( chk.errorProc ) )
}

func ( chk *responseCheck ) check() {
    mgRct.CheckBuiltValue( 
        chk.st.ResultValue, chk.resultProc, chk.Descend( "result" ) )
    mgRct.CheckBuiltValue( 
        chk.st.ErrorValue, chk.errorProc, chk.Descend( "error" ) )
}

func ( t *ServiceResponseTest ) Call( c *mgRct.ReactorTestCall ) {
    chk := &responseCheck{ st: t, PathAsserter: c.PathAsserter }
    svcDef := t.Definitions.MustGet( t.ServiceType ).( *ServiceDefinition )
    opDef := svcDef.mustFindOperation( t.Operation )
    rct := NewResponseReactor( t.Definitions, svcDef, opDef, chk )
    pip := mgRct.InitReactorPipeline( rct )
    visitAndCheck( t.In, pip, chk, t.Error, c )
}
