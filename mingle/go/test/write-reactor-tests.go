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
    pairs := valueErrorFieldPairs( ufe )
    pairs = append( pairs, "field", asValue( ufe.Field ) )
    return mkStruct( "UnrecognizedFieldError", pairs... )
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

func respTestAsValue( t *mg.ResponseReactorTest ) mg.Value {
    pairs := []interface{}{
        "in", t.In,
        "res-val", t.ResVal,
        "err-val", t.ErrVal,
        "error", asValue( t.Error ),
    }
    setEvs := func( k string, evs []mg.EventExpectation ) {
        if evs != nil { pairs = append( pairs, k, asValue( evs ) ) }
    }
    setEvs( "res-events", t.ResEvents )
    setEvs( "err-events", t.ErrEvents )
    return mkStruct( "ResponseReactorTest", pairs... )
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

func asFeedSource( src interface{} ) mg.Value {
    switch v := src.( type ) {
    case mg.Value: return mkStruct( "ValueSource", "value", v )
    case []mg.ReactorEvent: 
        return mkStruct( "ReactorEventSource", "events", asValue( v ) )
    }
    panic( fmt.Errorf( "unhandled source: %T", src ) )
}

func reqTestAsValue( st *mg.RequestReactorTest ) mg.Value {
    pairs := []interface{}{
        "source", asFeedSource( st.Source ),
        "authentication", asValue( st.Authentication ),
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
    if m := st.Parameters; m != nil { pairs = append( pairs, "parameters", m ) }
    if evs := st.ParameterEvents; evs != nil {
        pairs = append( pairs, "parameter-events", asValue( evs ) )
    }
    if evs := st.AuthenticationEvents; evs != nil {
        pairs = append( pairs, "authentication-events", asValue( evs ) )
    }
    return mkStruct( "RequestReactorTest", pairs... )
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
    case *mg.ValueEvent: return mkStruct( "ValueEvent", "val", v.Val )
    case *mg.StructStartEvent:
        return mkStruct( "StructStartEvent", "type", asValue( v.Type ) )
    case *mg.MapStartEvent: return mkStruct( "MapStartEvent" )
    case *mg.FieldStartEvent:
        return mkStruct( "FieldStartEvent", "field", asValue( v.Field ) )
    case *mg.ListStartEvent: return mkStruct( "ListStartEvent" )
    case *mg.EndEvent: return mkStruct( "EndEvent" )
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
    case *mg.RequestReactorTest: return reqTestAsValue( v )
    case *mg.ResponseReactorTest: return respTestAsValue( v )
    }
    panic( fmt.Errorf( "unhandled: %T", val ) )
}

type testData []interface{}

func ( td testData ) Len() int { return len( td ) }

func ( td testData ) StructAt( i int ) *mg.Struct {
    return asValue( td[ i ] ).( *mg.Struct )
}

func main() { testgen.WriteStructFile( testData( mg.StdReactorTests ) ) }
