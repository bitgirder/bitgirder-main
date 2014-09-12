package main

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "mingle/parser"
    "mingle/testgen"
    "fmt"
    "bitgirder/objpath"
    "reflect"
)

var (

    asType = parser.AsTypeReference

    nsMg = parser.MustNamespace( "mingle:core@v1" )
    nsMgRct = parser.MustNamespace( "mingle:reactor@v1" )
)

func mkStruct( ns *mg.Namespace, nm string, pairs ...interface{} ) *mg.Struct {
    qn := &mg.QualifiedTypeName{ 
        Namespace: ns, 
        Name: parser.MustDeclaredTypeName( nm ),
    }
    return parser.MustStruct( qn, pairs... )
}

func mgStruct( nm string, pairs ...interface{} ) *mg.Struct {
    return mkStruct( nsMg, nm, pairs... )
}

func mgRctStruct( nm string, pairs ...interface{} ) *mg.Struct {
    return mkStruct( nsMgRct, nm, pairs... )
}

func sliceAsValue( listType, s interface{} ) *mg.List {
    res := mg.NewList( asType( listType ).( *mg.ListTypeReference ) )
    v := reflect.ValueOf( s )
    if v.Kind() != reflect.Slice { panic( fmt.Errorf( "not a slice: %T", s ) ) }
    for i, e := 0, v.Len(); i < e; i++ {
        res.AddUnsafe( asValue( v.Index( i ).Interface() ) )
    }
    return res
}

func locatableErrorPairs( p objpath.PathNode, msg string ) []interface{} {
    return []interface{}{ "message", msg, "location", asValue( p ) }
}

func ufeAsValue( ufe *mg.UnrecognizedFieldError ) mg.Value {
    pairs := locatableErrorPairs( ufe.Location, ufe.Message )
    pairs = append( pairs, "field", asValue( ufe.Field ) )
    return mgStruct( "UnrecognizedFieldError", pairs... )
}

func mfeAsValue( mfe *mg.MissingFieldsError ) mg.Value {
    pairs := locatableErrorPairs( mfe.Location, mfe.Message )
    pairs = append( pairs, "fields", asValue( mfe.Fields() ) )
    return mgStruct( "MissingFieldsError", pairs... )
}

func fomfTestAsValue( t *mgRct.FieldOrderMissingFieldsTest ) mg.Value {
    pairs := []interface{}{
        "orders", asValue( t.Orders ),
        "source", asValue( t.Source ),
        "expect", asValue( t.Expect ),
    }
    if t.Error != nil { pairs = append( pairs, "error", asValue( t.Error ) ) }
    return mgRctStruct( "FieldOrderMissingFieldsTest", pairs... )
}    

func fopTestAsValue( t *mgRct.FieldOrderPathTest ) mg.Value {
    return mgRctStruct( "FieldOrderPathTest",
        "source", asValue( t.Source ),
        "expect", asValue( t.Expect ),
        "orders", asValue( t.Orders ),
    )
}

func fieldOrderReactorTestAsValue( t *mgRct.FieldOrderReactorTest ) mg.Value {
    return mgRctStruct( "FieldOrderReactorTest",
        "source", asValue( t.Source ),
        "expect", t.Expect,
        "orders", asValue( t.Orders ),
    )
}

func structuralReactorErrorTestAsValue( 
    t *mgRct.StructuralReactorErrorTest ) mg.Value {

    return mkStruct( nsMgRct, "StructuralReactorErrorTest", 
        "events", asValue( t.Events ),
        "top-type", asValue( t.TopType ),
        "error", asValue( t.Error ),
    )
}

func eventPathTestAsValue( t *mgRct.EventPathTest ) mg.Value {
    return mgRctStruct( "EventPathTest",
        "name", t.Name,
        "events", asValue( t.Events ),
        "start-path", asValue( t.StartPath ),
    )
}

func eeAsValue( ee mgRct.EventExpectation ) mg.Value {
    return mgRctStruct( "EventExpectation",
        "event", asValue( ee.Event ),
        "path", asValue( ee.Path ),
    )
}

func eventAsValue( ev mgRct.Event ) mg.Value {
    switch v := ev.( type ) {
    case *mgRct.ValueEvent: return mgRctStruct( "ValueEvent", "val", v.Val )
    case *mgRct.StructStartEvent:
        return mgRctStruct( "StructStartEvent", "type", asValue( v.Type ) )
    case *mgRct.MapStartEvent: return mgRctStruct( "MapStartEvent" )
    case *mgRct.FieldStartEvent:
        return mgRctStruct( "FieldStartEvent", "field", asValue( v.Field ) )
    case *mgRct.ListStartEvent: 
        return mgRctStruct( "ListStartEvent", "type", asValue( v.Type ) )
    case *mgRct.EndEvent: return mgRctStruct( "EndEvent" )
    }
    panic( fmt.Errorf( "unhandled event: %T", ev ) )
}

func reactorErrorAsValue( e *mgRct.ReactorError ) mg.Value {
    return mgRctStruct( "ReactorError", 
        "message", e.Error(),
        "location", asValue( e.Location ),
    )
}

func fieldOrderSpecAsValue( s mgRct.FieldOrderSpecification ) mg.Value {
    return mgRctStruct( "FieldOrderSpecification",
        "field", asValue( s.Field ),
        "required", s.Required,
    )
}

func fieldOrderAsValue( fo mgRct.FieldOrder ) mg.Value {
    return sliceAsValue( 
        "mingle:reactor@v1/FieldOrderSpecification*",
        []mgRct.FieldOrderSpecification( fo ),
    )
}

func fieldOrderReactorTestOrderAsValue( 
    ord mgRct.FieldOrderReactorTestOrder ) mg.Value {

    return mgRctStruct( "FieldOrderReactorTestOrder",
        "type", asValue( ord.Type ),
        "order", asValue( ord.Order ),
    )
}

func buildReactorTestAsValue( t *mgRct.BuildReactorTest ) mg.Value {
    pairs := []interface{}{
        "val", asValue( t.Val ),
        "source", asValue( t.Source ),
        "error", asValue( t.Error ),
    }
    if s := t.Profile; s != "" { pairs = append( pairs, "profile", s ) }
    return mgRctStruct( "BuildReactorTest", pairs... )
}

func mgRctTestErrorAsValue( e *mgRct.TestError ) mg.Value {
    return mgRctStruct( "TestError",
        "message", e.Message,
        "location", asValue( e.Location ),
    )
}

type idPathAcc struct { l *mg.List }

func ( acc idPathAcc ) Descend( elt interface{} ) error {
    acc.l.AddUnsafe( asValue( elt ) )
    return nil
}

func ( acc idPathAcc ) List( idx uint64 ) error {
    acc.l.AddUnsafe( mg.Uint64( idx ) )
    return nil
}

func idPathAsValue( p objpath.PathNode ) mg.Value {
    acc := idPathAcc{ mg.NewList( mg.TypeOpaqueList ) }
    if err := objpath.Visit( p, acc ); err != nil { panic( err ) }
    return acc.l
}

func mgRctS1AsValue( s *mgRct.TestStruct1 ) mg.Value {
    pairs := []interface{}{
        "f1", s.F1,
        "f2", asValue( s.F2 ),
    }
    if s.F3 != nil { pairs = append( pairs, "f3", asValue( s.F3 ) ) }
    return parser.MustStruct( "mingle:reactor@v1/TestStruct1", pairs... )
}

func strMapAsValue( m map[ string ]interface{} ) mg.Value {
    pairs := make( []interface{}, 0, 2 * len( m ) )
    for k, v := range m { pairs = append( pairs, k, asValue( v ) ) }
    return parser.MustSymbolMap( pairs... )
}

func depthTrackerTestAsValue( t *mgRct.DepthTrackerTest ) mg.Value {
    return mgRctStruct( "DepthTrackerTest",
        "source", asValue( t.Source ),
        "expect", asValue( t.Expect ),
    )
}

func asValue( val interface{} ) mg.Value {
    if val == nil { return mg.NullVal }
    switch v := val.( type ) {
    case mg.Value: return v
    case string: if v == "" { return mg.NullVal } else { return mg.String( v ) }
    case int32: return mg.Int32( v )
    case []int32: return sliceAsValue( "Int32*", v )
    case int: return mg.Int64( int64( v ) )
    case []int: return sliceAsValue( "Int64*", v )
    case map[ string ]interface{}: return strMapAsValue( v )
    case mg.TypeReference: return mg.Buffer( mg.TypeReferenceAsBytes( v ) )
    case *mg.Identifier: return mg.Buffer( mg.IdentifierAsBytes( v ) )
    case []*mg.Identifier: 
        return sliceAsValue( "mingle:core@v1/Identifier*", v )
    case *mg.QualifiedTypeName: 
        return mg.Buffer( mg.QualifiedTypeNameAsBytes( v ) )
    case *mgRct.TestStruct1: return mgRctS1AsValue( v )
    case mgRct.TestStruct2: return mgRctStruct( "S2" )
    case objpath.PathNode: return idPathAsValue( v )
    case *mgRct.TestError: return mgRctTestErrorAsValue( v ) 
    case *mgRct.BuildReactorTest: return buildReactorTestAsValue( v )
    case mgRct.EventExpectation: return eeAsValue( v )
    case []mgRct.EventExpectation: 
        return sliceAsValue( "mingle:reactor@v1/EventExpectation*", v )
    case mgRct.Event: return eventAsValue( v )
    case []mgRct.Event: return sliceAsValue( "mingle:reactor@v1/Event*", v )
    case mgRct.ReactorTopType: return mg.String( v.String() )
    case *mgRct.ReactorError: return reactorErrorAsValue( v )
    case *mg.UnrecognizedFieldError: return ufeAsValue( v )
    case *mg.MissingFieldsError: return mfeAsValue( v )
    case []mgRct.FieldOrderReactorTestOrder: 
        return sliceAsValue( 
            "mingle:reactor@v1/FieldOrderReactorTestOrder*", v )
    case mgRct.FieldOrderReactorTestOrder: 
        return fieldOrderReactorTestOrderAsValue( v )
    case *mgRct.FieldOrderPathTest: return fopTestAsValue( v )
    case *mgRct.FieldOrderMissingFieldsTest: return fomfTestAsValue( v )
    case mgRct.FieldOrder: return fieldOrderAsValue( v )
    case mgRct.FieldOrderSpecification: return fieldOrderSpecAsValue( v )
    case *mgRct.FieldOrderReactorTest: return fieldOrderReactorTestAsValue( v )
    case *mgRct.EventPathTest: return eventPathTestAsValue( v )
    case *mgRct.StructuralReactorErrorTest: 
        return structuralReactorErrorTestAsValue( v )
    case *mgRct.DepthTrackerTest: return depthTrackerTestAsValue( v )
    }
    panic( fmt.Errorf( "unhandled: %T", val ) )
}

type testData []mgRct.ReactorTest

func ( td testData ) Len() int { return len( td ) }

func ( td testData ) StructAt( i int ) *mg.Struct {
    return asValue( td[ i ] ).( *mg.Struct )
}

func getReactorTests() []mgRct.ReactorTest {
    res := mgRct.GetReactorTests()
    return res
}

func main() { testgen.WriteStructFile( testData( getReactorTests() ) ) }
