package types

import ( 
    "testing"
    "bitgirder/assert"
    mg "mingle"
)

type ReactorTestCall struct {
    *mg.ReactorTestCall
}

func ( tc *ReactorTestCall ) assertCastError( ct *CastReactorTest, err error ) {
    cae := mg.CastErrorAssert{ 
        ErrExpect: ct.Err, ErrAct: err, PathAsserter: tc.PathAsserter }
    switch ct.Err.( type ) {
    case *mg.UnrecognizedFieldError:
        if ufe, ok := err.( *mg.UnrecognizedFieldError ); ok {
            tc.Equal( ct.Err, ufe )
        } else { cae.FailActErrType() }
    case *mg.MissingFieldsError:
        if mfe, ok := err.( *mg.MissingFieldsError ); ok {
            tc.Equal( ct.Err, mfe )
        } else { cae.FailActErrType() }
    default: cae.Call()
    }
}

func ( tc *ReactorTestCall ) callCast( ct *CastReactorTest ) {
    rct := NewCastReactor( ct.Type, ct.Map )
    vb := mg.NewValueBuilder()
    pip := mg.InitReactorPipeline( rct, vb )
    if err := mg.VisitValue( ct.In, pip ); err == nil {
        tc.CheckNoError( ct.Err )
        mg.EqualValues( ct.Expect, vb.GetValue(), tc )
    } else { tc.assertCastError( ct, err ) }
}

func ( tc *ReactorTestCall ) callEventPath( t *EventPathTest ) {
    mg.AssertEventPaths( 
        t.Source,
        t.Expect, 
        []interface{}{ 
            NewCastReactor( t.Type, t.Map ),
//            mg.NewDebugReactor( tc ),
        },
        tc.PathAsserter,
    )
}

type checker interface { check() }

func ( tc *ReactorTestCall ) visitAndCheck(
    in mg.Value, rep mg.ReactorEventProcessor, chk checker, errExpct error ) {
    if err := mg.VisitValue( in, rep ); err == nil {
        tc.CheckNoError( errExpct )
        chk.check()
    } else { tc.EqualErrors( errExpct, err ) }
}

type requestCheck struct {
    *assert.PathAsserter
    st *ServiceRequestTest
    auth *mg.ValueBuilder
    params *mg.ValueBuilder
}

func ( chk *requestCheck ) GetAuthenticationProcessor(
    om OpMatch, pg mg.PathGetter ) ( mg.ReactorEventProcessor, error ) {
    chk.auth = mg.NewValueBuilder()
    return chk.auth, nil
}

func ( chk *requestCheck ) GetParametersProcessor(
    om OpMatch, pg mg.PathGetter ) ( mg.ReactorEventProcessor, error ) {
    chk.params = mg.NewValueBuilder()
    return chk.params, nil
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
//    tc.Logf( "Feeding request: %s", mg.QuoteValue( st.In ) )
    tc.visitAndCheck( st.In, pip, chk, st.Error )
}

func ( tc *ReactorTestCall ) call() {
//    tc.Logf( "Calling test of type %T", tc.Test )
    switch v := tc.Test.( type ) {
    case *CastReactorTest: tc.callCast( v )
    case *EventPathTest: tc.callEventPath( v )
    case *ServiceRequestTest: tc.callServiceRequest( v )
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
