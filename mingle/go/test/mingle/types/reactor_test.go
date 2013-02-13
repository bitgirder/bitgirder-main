package types

import ( 
    "testing"
    "bitgirder/assert"
    mg "mingle"
)

type testCall struct {
    *assert.PathAsserter
    rt interface{}
}

func ( tc testCall ) assertCastError( ct *CastReactorTest, err error ) {
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

func ( tc testCall ) callCast( ct *CastReactorTest ) {
    rct := NewCastReactor( ct.Type, ct.Map )
    vb := mg.NewValueBuilder()
    pip := mg.InitReactorPipeline( rct, vb )
    if err := mg.VisitValue( ct.In, pip ); err == nil {
        if errExpct := ct.Err; errExpct == nil {
            act := vb.GetValue()
            tc.Equal( ct.Expect, act )
        } else { tc.Fatalf( "Expected error (%T): %s", errExpct, errExpct ) }
    } else { tc.assertCastError( ct, err ) }
}

func ( tc testCall ) callEventPath( t *EventPathTest ) {
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

func ( tc testCall ) call() {
//    tc.Logf( "Calling test of type %T", tc.rt )
    switch v := tc.rt.( type ) {
    case *CastReactorTest: tc.callCast( v )
    case *EventPathTest: tc.callEventPath( v )
    default: panic( libErrorf( "Unhandled test type: %T", tc.rt ) )
    }
}

func TestReactors( t *testing.T ) {
    a := assert.NewListPathAsserter( t )
    for _, rt := range GetStdReactorTests() {
        ( testCall{ PathAsserter: a, rt: rt } ).call()
        a = a.Next()
    }
}
