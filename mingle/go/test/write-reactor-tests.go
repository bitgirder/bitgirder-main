package main

import (
    mg "mingle"
    "mingle/testgen"
    "fmt"
//    "log"
    "bitgirder/objpath"
    "reflect"
)

func mkStruct( nm string, pairs ...interface{} ) *mg.Struct {
    typStr := fmt.Sprintf( "mingle:core:testing@v1/%s", nm )
    return mg.MustStruct( typStr, pairs... )
}

func crtToValue( crt *mg.CastReactorTest ) *mg.Struct {
    return mkStruct( "CastReactorTest", 
        "in", crt.In,
        "expect", crt.Expect,
        "path", asValue( crt.Path ),
        "type", asValue( crt.Type ),
        "err", asValue( crt.Err ),
        "profile", asValue( crt.Profile ),
    )
}

func typeRefToValue( t mg.TypeReference ) mg.Value {
    return mg.Buffer( mg.TypeReferenceAsBytes( t ) )
}

func sliceAsValue( s interface{} ) *mg.List {
    v := reflect.ValueOf( s )
    if v.Kind() != reflect.Slice { panic( fmt.Errorf( "not a slice: %T", s ) ) }
    vals := make( []mg.Value, v.Len() )
    for i, e := 0, v.Len(); i < e; i++ {
        vals[ i ] = asValue( v.Index( i ).Interface() )
    }
    return mg.NewList( vals )
}

type valueError interface {
    error
    Message() string
    Location() objpath.PathNode
}

func valueErrorFieldPairs( ve valueError ) []interface{} {
    return []interface{}{
        "message", ve.Message(),
        "location", asValue( ve.Location() ),
    }
}

func vceAsValue( vce *mg.ValueCastError ) mg.Value {
    return mkStruct( "ValueCastError", valueErrorFieldPairs( vce )... )
}

func ufeAsValue( ufe *mg.UnrecognizedFieldError ) mg.Value {
    return mkStruct( "UnrecognizedFieldError", valueErrorFieldPairs( ufe )... )
}

func mfeAsValue( mfe *mg.MissingFieldsError ) mg.Value {
    pairs := valueErrorFieldPairs( mfe )
    pairs = append( pairs, "fields", asValue( mfe.Fields() ) )
    return mkStruct( "MissingFieldsError", pairs... )
}

func fomfTestAsValue( t *mg.FieldOrderMissingFieldsTest ) mg.Value {
    pairs := []interface{}{
        "orders", asValue( t.Orders ),
        "source", asValue( t.Source ),
        "expect", asValue( t.Expect ),
    }
    if t.Error != nil { pairs = append( pairs, "error", asValue( t.Error ) ) }
    return mkStruct( "FieldOrderMissingFieldsTest", pairs... )
}    

func fopTestAsValue( t *mg.FieldOrderPathTest ) mg.Value {
    return mkStruct( "FieldOrderPathTest",
        "source", asValue( t.Source ),
        "expect", asValue( t.Expect ),
        "orders", asValue( t.Orders ),
    )
}

func fieldOrderReactorTestAsValue( t *mg.FieldOrderReactorTest ) mg.Value {
    return mkStruct( "FieldOrderReactorTest",
        "source", asValue( t.Source ),
        "expect", t.Expect,
        "orders", asValue( t.Orders ),
    )
}

func structuralReactorErrorTestAsValue( 
    t *mg.StructuralReactorErrorTest ) mg.Value {
    return mkStruct( "StructuralReactorErrorTest", 
        "events", asValue( t.Events ),
        "top-type", asValue( t.TopType ),
        "error", asValue( t.Error ),
    )
}

func eventPathTestAsValue( t *mg.EventPathTest ) mg.Value {
    return mkStruct( "EventPathTest",
        "events", asValue( t.Events ),
        "start-path", asValue( t.StartPath ),
    )
}

func svcRespTestAsValue( t *mg.ServiceResponseReactorTest ) mg.Value {
    return mkStruct( "ServiceResponseReactorTest",
        "in", t.In,
        "res-val", t.ResVal,
        "res-events", asValue( t.ResEvents ),
        "err-val", t.ErrVal,
        "err-events", asValue( t.ErrEvents ),
        "error", asValue( t.Error ),
    )
}

func eeAsValue( ee mg.EventExpectation ) mg.Value {
    return mkStruct( "EventExpectation",
        "event", asValue( ee.Event ),
        "path", asValue( ee.Path ),
    )
}

func fieldOrderSpecAsValue( s mg.FieldOrderSpecification ) mg.Value {
    return mkStruct( "FieldOrderSpecification",
        "field", asValue( s.Field ),
        "required", s.Required,
    )
}

func fieldOrderReactorTestOrderAsValue( 
    ord mg.FieldOrderReactorTestOrder ) mg.Value {
    return mkStruct( "FieldOrderReactorTestOrder",
        "type", asValue( ord.Type ),
        "order", asValue( ord.Order ),
    )
}

func svcReqTestAsValue( st *mg.ServiceRequestReactorTest ) mg.Value {
    pairs := []interface{}{
        "source", asValue( st.Source ),
        "parameter-events", asValue( st.ParameterEvents ),
        "authentication", asValue( st.Authentication ),
        "authentication-events", asValue( st.AuthenticationEvents ),
        "error", asValue( st.Error ),
    }
    if st.Namespace != nil {
        pairs = append( pairs, "namespace", asValue( st.Namespace ) )
    }
    addId := func( k string, id *mg.Identifier ) {
        if id != nil { pairs = append( pairs, k, asValue( id ) ) }
    }
    addId( "service", st.Service )
    addId( "operation", st.Operation )
    if st.Parameters != nil { 
        pairs = append( pairs, "parameters", st.Parameters )
    }
    return mkStruct( "ServiceRequestReactorTest", pairs... )
}

func asValue( val interface{} ) mg.Value {
    if val == nil { return mg.NullVal }
    switch v := val.( type ) {
    case mg.Value: return v
    case *mg.CastReactorTest: return crtToValue( v )
    case string: if v == "" { return mg.NullVal } else { return mg.String( v ) }
    case mg.TypeReference: return mg.Buffer( mg.TypeReferenceAsBytes( v ) )
    case *mg.Identifier: return mg.Buffer( mg.IdentifierAsBytes( v ) )
    case []*mg.Identifier: return sliceAsValue( v )
    case *mg.QualifiedTypeName: 
        return mg.Buffer( mg.QualifiedTypeNameAsBytes( v ) )
    case *mg.Namespace: return mg.Buffer( mg.NamespaceAsBytes( v ) )
    case objpath.PathNode: return mg.Buffer( mg.IdPathAsBytes( v ) )
    case *mg.ValueCastError: return vceAsValue( v )
    case mg.ValueBuildTest: return mkStruct( "ValueBuildTest", "val", v.Val )
    case mg.EventExpectation: return eeAsValue( v )
    case []mg.EventExpectation: return sliceAsValue( v )
    case mg.ValueEvent: return mkStruct( "ValueEvent", "val", v.Val )
    case mg.StructStartEvent:
        return mkStruct( "StructStartEvent", "type", asValue( v.Type ) )
    case mg.MapStartEvent: return mkStruct( "MapStartEvent" )
    case mg.FieldStartEvent:
        return mkStruct( "FieldStartEvent", "field", asValue( v.Field ) )
    case mg.ListStartEvent: return mkStruct( "ListStartEvent" )
    case mg.EndEvent: return mkStruct( "EndEvent" )
    case []mg.ReactorEvent: return sliceAsValue( v )
    case mg.ReactorTopType: return mg.String( v.String() )
    case *mg.ReactorError:
        return mkStruct( "ReactorError", "message", v.Error() )
    case *mg.UnrecognizedFieldError: return ufeAsValue( v )
    case *mg.MissingFieldsError: return mfeAsValue( v )
    case []mg.FieldOrderReactorTestOrder: return sliceAsValue( v )
    case mg.FieldOrderReactorTestOrder: 
        return fieldOrderReactorTestOrderAsValue( v )
    case *mg.FieldOrderPathTest: return fopTestAsValue( v )
    case *mg.FieldOrderMissingFieldsTest: return fomfTestAsValue( v )
    case mg.FieldOrder: return sliceAsValue( []mg.FieldOrderSpecification( v ) )
    case mg.FieldOrderSpecification: return fieldOrderSpecAsValue( v )
    case *mg.FieldOrderReactorTest: return fieldOrderReactorTestAsValue( v )
    case *mg.EventPathTest: return eventPathTestAsValue( v )
    case *mg.StructuralReactorErrorTest: 
        return structuralReactorErrorTestAsValue( v )
    case *mg.ServiceRequestReactorTest: return svcReqTestAsValue( v )
    case *mg.ServiceResponseReactorTest: return svcRespTestAsValue( v )
    }
    panic( fmt.Errorf( "unhandled: %T", val ) )
}

type testData []interface{}

func ( td testData ) Len() int { return len( td ) }

func ( td testData ) StructAt( i int ) *mg.Struct {
    return asValue( td[ i ] ).( *mg.Struct )
}

func main() { testgen.WriteStructFile( testData( mg.StdReactorTests ) ) }
