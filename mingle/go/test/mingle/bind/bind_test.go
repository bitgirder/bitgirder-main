package bind

import (
    "reflect"
    mg "mingle"
    mgRct "mingle/reactor"
    "testing"
    "bitgirder/assert"
    "bitgirder/stub"
    "bitgirder/objpath"
)

type customString string

type customBinderFactory int

func ( b customBinderFactory ) BindValue( 
    ve *mgRct.ValueEvent ) ( interface{}, error ) {

    if s, ok := ve.Val.( mg.String ); ok {
        return customString( string( s ) ), nil
    }
    return DefaultBindingForValue( ve.Val ), nil
}

func ( b customBinderFactory ) StartMap(
    mse *mgRct.MapStartEvent ) ( FieldSetBinder, error ) {

    return nil, stub.Unimplemented()
}

var qnS1 = mkQn( "ns1@v1/S1" )

type s1Binder struct {
    s1 S1
}

func ( b *s1Binder ) ProduceValue( 
    ee *mgRct.EndEvent ) ( interface{}, error ) {

    if b.s1.f1 == s1F1ValFailOnProduce {
        return nil, NewBindError( ee.GetPath(), testMsgErrorBadValue )
    }
    return b.s1, nil
}

func ( b *s1Binder ) SetValue(
    fld *mg.Identifier, val interface{}, path objpath.PathNode ) error {

    switch s := fld.ExternalForm(); s {
    case "f1":
        if i, ok := val.( int32 ); ok && i >= 0 {
            b.s1.f1 = i
            return nil
        } 
        return NewBindError( path, testMsgErrorBadValue )
    }
    return NewBindErrorf( path, "unhandled field: %s", fld )
}

func ( b *s1Binder ) StartField(
    fse *mgRct.FieldStartEvent ) ( BinderFactory, error ) {

    return customBinderFactory( 1 ), nil
}

func ( b customBinderFactory ) StartStruct(
    sse *mgRct.StructStartEvent ) ( FieldSetBinder, error ) {
 
    switch {
    case sse.Type.Equals( qnS1 ): return &s1Binder{}, nil
    }
    return nil, NewBindErrorf( sse.GetPath(), "unhandled type: %s", sse.Type )
}

type sliceBinder struct {
    slice interface{}
}

func ( sb *sliceBinder ) ProduceValue( 
    ee *mgRct.EndEvent ) ( interface{}, error ) {

    return sb.slice, nil
}

func ( sb *sliceBinder ) AddValue( 
    val interface{}, path objpath.PathNode ) error {
    
    slice := reflect.ValueOf( sb.slice )
    apnd := reflect.ValueOf( val )
    sb.slice = reflect.Append( slice, apnd ).Interface()
    return nil
}

func ( sb *sliceBinder ) NextBinderFactory() BinderFactory {
    return customBinderFactory( 1 )
}

func ( b customBinderFactory ) StartList(
    lse *mgRct.ListStartEvent ) ( ListBinder, error ) {

    switch s := lse.Type.ExternalForm(); s {
    case "ns1@v1/S1*": return &sliceBinder{ make( []S1, 0, 4 ) }, nil
    }
    return nil, NewBindErrorf( lse.GetPath(), "unhandled type: %s", lse.Type )
}

func binderFactoryForTest( bt *BindTest ) BinderFactory {
    switch bt.Profile {
    case TestProfileDefaultValue: return DefaultBinderFactory
    case TestProfileCustomValue: return customBinderFactory( 1 )
    }
    panic( libErrorf( "unhandled profile: %s", bt.Profile ) )
}

func callBindTest( bt *BindTest, a *assert.PathAsserter ) {
//    a.Logf( "binding %s (%T) as %s with profile %s", 
//        mg.QuoteValue( bt.In ), bt.In, bt.Type, bt.Profile )
    bf := binderFactoryForTest( bt )
    br := NewBindReactor( bf )
//    pip := mgRct.InitReactorPipeline( mgRct.NewDebugReactor( a ), br )
    pip := mgRct.InitReactorPipeline( br )
    if err := mgRct.VisitValue( bt.In, pip ); err == nil {
        a.Equal( bt.Expect, br.GetValue() )
    } else {
        a.EqualErrors( bt.Error, err )
    }
}

func TestBinding( t *testing.T ) {
    la := assert.NewListPathAsserter( t )
    for _, bt := range stdBindTests {
        callBindTest( bt, la )
        la = la.Next()
    }
}
