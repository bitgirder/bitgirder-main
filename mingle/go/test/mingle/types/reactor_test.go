package types

import ( 
    "testing"
    "bitgirder/assert"
    "bitgirder/objpath"
    mg "mingle"
)

type ReactorTestCall struct { *mg.ReactorTestCall }

func ( tc *ReactorTestCall ) callCast( ct *CastReactorTest ) {
    rct := NewCastReactorDefinitionMap( ct.Type, ct.Map )
    vb := mg.NewValueBuilder()
    pip := mg.InitReactorPipeline( rct, vb )
    if err := mg.VisitValue( ct.In, pip ); err == nil {
        tc.CheckNoError( ct.Err )
        mg.EqualValues( ct.Expect, vb.GetValue(), tc )
    } else { 
        cae := mg.CastErrorAssert{ 
            ErrExpect: ct.Err, ErrAct: err, PathAsserter: tc.PathAsserter }
        cae.Call()
    }
}

func ( tc *ReactorTestCall ) callEventPath( t *EventPathTest ) {
    chk := mg.NewEventPathCheckReactor( t.Expect, tc.PathAsserter )
    rct := NewCastReactorDefinitionMap( t.Type, t.Map )
    pip := mg.InitReactorPipeline( rct, chk )
    mg.AssertFeedSource( t.Source, pip, tc )
    chk.Complete()
}

type checker interface { check() }

func ( tc *ReactorTestCall ) visitAndCheck(
    in mg.Value, rep mg.ReactorEventProcessor, chk checker, errExpct error ) {

    if err := mg.VisitValue( in, rep ); err == nil {
        tc.CheckNoError( errExpct )
        chk.check()
    } else { tc.EqualErrors( errExpct, err ) }
}

func initValueBuilder( 
    vbPtr **mg.ValueBuilder ) ( mg.ReactorEventProcessor, error ) {

    *vbPtr = mg.NewValueBuilder()
    return *vbPtr, nil
}

type requestCheck struct {
    *assert.PathAsserter
    st *ServiceRequestTest
    auth *mg.ValueBuilder
    params *mg.ValueBuilder
}

func ( chk *requestCheck ) GetAuthenticationReactor(
    path objpath.PathNode ) ( mg.ReactorEventProcessor, error ) {

    return initValueBuilder( &( chk.auth ) )
}

func ( chk *requestCheck ) GetParametersReactor(
    path objpath.PathNode ) ( mg.ReactorEventProcessor, error ) {

    return initValueBuilder( &( chk.params ) )
}

func ( chk *requestCheck ) check() {
    mg.CheckBuiltValue( 
        chk.st.Authentication, chk.auth, chk.Descend( "authentication" ) )
    mg.CheckBuiltValue( 
        chk.st.Parameters, chk.params, chk.Descend( "parameters" ) )
}

func ( tc *ReactorTestCall ) callServiceRequest( st *ServiceRequestTest ) {
    chk := &requestCheck{ st: st, PathAsserter: tc.PathAsserter }
    rct := NewRequestReactor( st.Maps.BuildOpMap(), chk )
    pip := mg.InitReactorPipeline( rct )
    tc.visitAndCheck( st.In, pip, chk, st.Error )
}

type responseCheck struct {
    st *ServiceResponseTest
    *assert.PathAsserter
    resultProc, errorProc *mg.ValueBuilder
}

func ( chk *responseCheck ) GetResultReactor( 
    p objpath.PathNode ) ( mg.ReactorEventProcessor, error ) {

    return initValueBuilder( &( chk.resultProc ) )
}

func ( chk *responseCheck ) GetErrorReactor( 
    p objpath.PathNode ) ( mg.ReactorEventProcessor, error ) {

    return initValueBuilder( &( chk.errorProc ) )
}

func ( chk *responseCheck ) check() {
    mg.CheckBuiltValue( 
        chk.st.ResultValue, chk.resultProc, chk.Descend( "result" ) )
    mg.CheckBuiltValue( 
        chk.st.ErrorValue, chk.errorProc, chk.Descend( "error" ) )
}

func ( tc *ReactorTestCall ) callServiceResponse( st *ServiceResponseTest ) {
    chk := &responseCheck{ st: st, PathAsserter: tc.PathAsserter }
    svcDef := st.Definitions.MustGet( st.ServiceType ).( *ServiceDefinition )
    opDef := svcDef.mustFindOperation( st.Operation )
    rct := NewResponseReactor( st.Definitions, svcDef, opDef, chk )
    pip := mg.InitReactorPipeline( rct )
    tc.visitAndCheck( st.In, pip, chk, st.Error )
}

func ( tc *ReactorTestCall ) call() {
//    tc.Logf( "Calling test of type %T", tc.Test )
    switch v := tc.Test.( type ) {
    case *CastReactorTest: tc.callCast( v )
    case *EventPathTest: tc.callEventPath( v )
    case *ServiceRequestTest: tc.callServiceRequest( v )
    case *ServiceResponseTest: tc.callServiceResponse( v )
    default: panic( libErrorf( "Unhandled test type: %T", tc.Test ) )
    }
}

func TestReactors( t *testing.T ) {
    a := assert.NewListPathAsserter( t )
    for _, rt := range GetStdReactorTests() {
        tc := &ReactorTestCall{ 
            ReactorTestCall: &mg.ReactorTestCall{
                PathAsserter: a, 
                Test: rt,
            },
        }
        tc.call()
        a = a.Next()
    }
}
