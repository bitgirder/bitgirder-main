package mingle

import (
    "testing"
    "bitgirder/objpath"
    "bitgirder/assert"
)

type serviceReactorImplTest struct {
    *assert.PathAsserter
    failOn *Identifier
    in Value
}

func ( t serviceReactorImplTest ) Error() string { 
    return "serviceReactorImplTest" 
}

func ( t serviceReactorImplTest ) Namespace( 
    ns *Namespace, path objpath.PathNode ) error {

    return nil
}

func ( t serviceReactorImplTest ) Service( 
    svc *Identifier, path objpath.PathNode ) error {

    return nil
}

func ( t serviceReactorImplTest ) Operation( 
    op *Identifier, path objpath.PathNode ) error {

    return nil
}

func ( t serviceReactorImplTest ) makeErr( path objpath.PathNode ) error {
    return NewValueCastError( path, "test-error" )
}

func ( t serviceReactorImplTest ) getProcessor(
    path objpath.PathNode, id *Identifier ) ( ReactorEventProcessor, error ) {

    if t.failOn.Equals( id ) { return nil, t.makeErr( path ) }
    return DiscardProcessor, nil
}

func ( t serviceReactorImplTest ) GetAuthenticationProcessor( 
    path objpath.PathNode ) ( ReactorEventProcessor, error ) {

    return t.getProcessor( path, IdAuthentication )
}

func ( t serviceReactorImplTest ) GetParametersProcessor( 
    path objpath.PathNode ) ( ReactorEventProcessor, error ) {

    return t.getProcessor( path, IdParameters )
}

func ( t serviceReactorImplTest ) GetErrorProcessor(
    path objpath.PathNode ) ( ReactorEventProcessor, error ) {

    return t.getProcessor( path, IdError )
}

func ( t serviceReactorImplTest ) GetResultProcessor(
    path objpath.PathNode ) ( ReactorEventProcessor, error ) {

    return t.getProcessor( path, IdResult )
}

func ( t serviceReactorImplTest ) callWith( rct ReactorEventProcessor ) {
    pip := InitReactorPipeline( rct )
    err := VisitValue( t.in, pip )
    errExpct := NewValueCastError( objpath.RootedAt( t.failOn ), "test-error" )
    t.EqualErrors( errExpct, err )
}

func TestRequestReactorImplErrors( t *testing.T ) {
    a := assert.NewPathAsserter( t )
    in := MustStruct( QnameServiceRequest,
        "namespace", "ns1@v1",
        "service", "svc1",
        "operation", "op1",
        "parameters", MustSymbolMap( "p1", 1 ),
        "authentication", 1,
    )
    for _, failOn := range []*Identifier{ IdAuthentication, IdParameters } {
        test := serviceReactorImplTest{ 
            PathAsserter: a.Descend( failOn ), 
            failOn: failOn,
            in: in,
        }
        test.callWith( NewServiceRequestReactor( test ) )
    }
}

func TestResponseReactorImplErrors( t *testing.T ) {
    chk := func( failOn *Identifier, in Value ) {
        test := serviceReactorImplTest{
            PathAsserter: assert.NewPathAsserter( t ).Descend( failOn ),
            failOn: failOn,
            in: in,
        }
        test.callWith( NewServiceResponseReactor( test ) )
    }
    chk( IdResult, MustSymbolMap( "result", 1 ) )
    chk( IdError, MustSymbolMap( "error", 1 ) )
}
