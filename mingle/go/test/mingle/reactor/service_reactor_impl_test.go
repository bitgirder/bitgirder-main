package reactor

//import (
//    mg "mingle"
//    "mingle/parser"
//    "testing"
//    "bitgirder/objpath"
//    "bitgirder/assert"
//)
//
//type serviceReactorImplTest struct {
//    *assert.PathAsserter
//    failOn *mg.Identifier
//    in mg.Value
//}
//
//func ( t serviceReactorImplTest ) Error() string { 
//    return "serviceReactorImplTest" 
//}
//
//func ( t serviceReactorImplTest ) Namespace( 
//    ns *mg.Namespace, path objpath.PathNode ) error {
//
//    return nil
//}
//
//func ( t serviceReactorImplTest ) Service( 
//    svc *mg.Identifier, path objpath.PathNode ) error {
//
//    return nil
//}
//
//func ( t serviceReactorImplTest ) Operation( 
//    op *mg.Identifier, path objpath.PathNode ) error {
//
//    return nil
//}
//
//func ( t serviceReactorImplTest ) makeErr( path objpath.PathNode ) error {
//    return mg.NewValueCastError( path, "test-error" )
//}
//
//func ( t serviceReactorImplTest ) getProcessor(
//    path objpath.PathNode, 
//    id *mg.Identifier ) ( ReactorEventProcessor, error ) {
//
//    if t.failOn.Equals( id ) { return nil, t.makeErr( path ) }
//    return DiscardProcessor, nil
//}
//
//func ( t serviceReactorImplTest ) GetAuthenticationReactor( 
//    path objpath.PathNode ) ( ReactorEventProcessor, error ) {
//
//    return t.getProcessor( path, mg.IdAuthentication )
//}
//
//func ( t serviceReactorImplTest ) GetParametersReactor( 
//    path objpath.PathNode ) ( ReactorEventProcessor, error ) {
//
//    return t.getProcessor( path, mg.IdParameters )
//}
//
//func ( t serviceReactorImplTest ) GetErrorReactor(
//    path objpath.PathNode ) ( ReactorEventProcessor, error ) {
//
//    return t.getProcessor( path, mg.IdError )
//}
//
//func ( t serviceReactorImplTest ) GetResultReactor(
//    path objpath.PathNode ) ( ReactorEventProcessor, error ) {
//
//    return t.getProcessor( path, mg.IdResult )
//}
//
//func ( t serviceReactorImplTest ) callWith( rct ReactorEventProcessor ) {
//    pip := InitReactorPipeline( rct )
//    err := VisitValue( t.in, pip )
//    errExpct := 
//        mg.NewValueCastError( objpath.RootedAt( t.failOn ), "test-error" )
//    t.EqualErrors( errExpct, err )
//}
//
//func TestRequestReactorImplErrors( t *testing.T ) {
//    a := assert.NewPathAsserter( t )
//    in := parser.MustStruct( mg.QnameRequest,
//        "namespace", "ns1@v1",
//        "service", "svc1",
//        "operation", "op1",
//        "parameters", parser.MustSymbolMap( "p1", 1 ),
//        "authentication", 1,
//    )
//    for _, failOn := range []*mg.Identifier{ 
//        mg.IdAuthentication, 
//        mg.IdParameters,
//    } {
//        test := serviceReactorImplTest{ 
//            PathAsserter: a.Descend( failOn ), 
//            failOn: failOn,
//            in: in,
//        }
//        test.callWith( NewRequestReactor( test ) )
//    }
//}
//
//func TestResponseReactorImplErrors( t *testing.T ) {
//    chk := func( failOn *mg.Identifier, in mg.Value ) {
//        test := serviceReactorImplTest{
//            PathAsserter: assert.NewPathAsserter( t ).Descend( failOn ),
//            failOn: failOn,
//            in: in,
//        }
//        test.callWith( NewResponseReactor( test ) )
//    }
//    chk( mg.IdResult, parser.MustSymbolMap( "result", 1 ) )
//    chk( mg.IdError, parser.MustSymbolMap( "error", 1 ) )
//}
