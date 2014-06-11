package bind

import (
    mg "mingle"
    "testing"
    "reflect"
    "bitgirder/assert"
    "bitgirder/objpath"
)

var mkTyp = parser.MustTypeReference

type valueBindingImpl struct { val mg.Value }

func ( vb valueBindingImpl ) BoundValue( 
    path objpath.PathNode ) ( mg.Value, error ) { 

    return vb.val, nil
}

type listBindingImpl struct { 
    list reflect.Value
    idx int
}

func newListBindingImpl( list interface{} ) *listBindingImpl {
    return &listBindingImpl{ list: reflect.ValueOf( list ) }
}

func ( lb *listBindingImpl ) HasNextBinding() bool {
    return lb.idx < lb.list.Len()
}

func ( lb *listBindingImpl ) NextBinding( 
    path objpath.PathNode ) ( interface{}, error ) {

    idx := lb.idx
    lb.idx++
    var val interface{}
    rv := lb.list.Index( idx )
    switch rv.Kind() {
    case reflect.Slice: if ! rv.IsNil() { val = rv.Interface() }
    default: val = rv.Interface()
    }
    return bindingForValue( path, val )
}

func bindingForValue( 
    path objpath.PathNode, val interface{} ) ( interface{}, error ) {

    if val == nil { return valueBindingImpl{ mg.NullVal }, nil }
    if reflect.TypeOf( val ).Kind() == reflect.Slice {
        return newListBindingImpl( val ), nil
    }
    return valueBindingImpl{ mg.MustValue( val ) }, nil
}

type listUnbinderImpl struct {
    list interface{}
    start func() interface{}
    apnd func( list, val interface{} ) interface{}
    next func() Unbinder
}

func ( lu *listUnbinderImpl ) StartList( 
    path objpath.PathNode ) ( ListUnbinder, error ) {

    lu.list = lu.start()
    return lu, nil
}

func ( lu *listUnbinderImpl ) Append( 
    val interface{}, path objpath.PathNode ) ( ListUnbinder, error ) {

    lu.list = lu.apnd( lu.list, val )
    return lu, nil
}

func ( lu *listUnbinderImpl ) NextUnbinder(
    path objpath.PathNode ) ( Unbinder, error ) {

    return lu.next(), nil
}

func ( lu *listUnbinderImpl ) EndList( 
    path objpath.PathNode ) ( interface{}, error ) {

    return lu.list, nil
}

func newListUnbinder( typ mg.TypeReference ) ListUnbinder {

    switch {
    case typ.Equals( mkTyp( "Int32*" ) ) ||
         typ.Equals( mkTyp( "Int32*?" ) ): 
        return &listUnbinderImpl{
            start: func() interface{} { return []int32{} },
            apnd: func( list, val interface{} ) interface{} {
                return append( list.( []int32 ), val.( int32 ) )
            },
            next: func() Unbinder { return Int32ValueUnbinder },
        }
    case typ.Equals( mkTyp( "Int32*+" ) ) ||
         typ.Equals( mkTyp( "Int32*?*" ) ) ||
         typ.Equals( mkTyp( "Int32**" ) ):
        return &listUnbinderImpl{
            start: func() interface{} { return [][]int32{} },
            apnd: func( list, val interface{} ) interface{} {
                if val == nil { return append( list.( [][]int32 ), nil ) }
                return append( list.( [][]int32 ), val.( []int32 ) )
            },
            next: func() Unbinder {
                eltTyp := typ.( *mg.ListTypeReference ).ElementType
                return newListUnbinder( eltTyp )
            },
        }
    }
    panic( libErrorf( "unhandled list type: %s", typ ) )
}

func valueUnbinderForType( 
    path objpath.PathNode, typ *mg.AtomicTypeReference ) ( Unbinder, error ) {

    switch {
    case typ.Equals( mg.TypeInt32 ): return Int32ValueUnbinder, nil
    }
    return nil, NewUnbindErrorf( path, "unhandled unbinder type: %s", typ )
}

func unbinderForType( 
    path objpath.PathNode, typ mg.TypeReference ) ( Unbinder, error ) {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference: return valueUnbinderForType( path, v )
    case *mg.NullableTypeReference: return unbinderForType( path, v.Type )
    case *mg.ListTypeReference: return newListUnbinder( v ), nil
    }
    return nil, NewUnbindErrorf( path, "unhandled unbinder type: %s", typ )
}

func mustUnbinderForType( typ mg.TypeReference ) Unbinder {
    ub, err := unbinderForType( nil, typ )
    if err == nil { return ub }
    panic( err )
}

type roundtripTestCall struct {
    *assert.PathAsserter
    rt *RoundtripTest
    binder *Binder
}

func ( c *roundtripTestCall ) bindToValue() mg.Value {
    vb := mg.NewValueBuilder()
    binding, err := bindingForValue( nil, c.rt.Object )
    if err != nil { c.Fatal( err ) }
    if err = c.binder.Bind( binding, vb ); err != nil { c.Fatal( err ) }
    return vb.GetValue()
}

func ( c *roundtripTestCall ) unbindFromValue( val mg.Value ) interface{} {
    ub, err := unbinderForType( nil, c.rt.Type )
    if err != nil { c.Fatal( err ) }
    ps := mg.NewPathSettingProcessor()
    dbg := mg.NewDebugReactor( c )
    rct := c.binder.NewUnbindReactor( ub )
    pip := mg.InitReactorPipeline( ps, dbg, rct )
    if err = mg.VisitValue( val, pip ); err != nil { c.Fatal( err ) }
    c.True( rct.HasValue() )
    return rct.UnboundValue()
}

func ( c *roundtripTestCall ) call() {
    c.binder = NewBinder()
    val := c.bindToValue()
    mg.EqualValues( c.rt.Value, val, c )
    goVal2 := c.unbindFromValue( val )
    c.Equal( c.rt.Object, goVal2 )
}

func callRoundtripTest( rt *RoundtripTest, a *assert.PathAsserter ) {
    ( &roundtripTestCall{ rt: rt, PathAsserter: a } ).call()
}

func assertUnbindError( 
    expct *UnbindError, act error, a *assert.PathAsserter ) {

    if ue, ok := act.( *UnbindError ); ok {
        a.Equalf( expct, ue, "expected %s (%T) but got %s (%T)",
            expct, expct, act, act )
    } else {
        a.Fatal( act )
    }
}

func assertError( expct, act error, a *assert.PathAsserter ) {
    switch v := expct.( type ) {
    case *UnbindError: assertUnbindError( v, act, a )
    default: mg.AssertCastError( expct, act, a )
    }
}

func callUnbindErrorTest( ue *UnbindErrorTest, a *assert.PathAsserter ) {
    b := NewBinder()
    ps := mg.NewPathSettingProcessor()
//    if _, ok := ue.Type.( *mg.ListTypeReference ); ok {
//        ps.SetStartPath( objpath.RootedAtList() )
//    }
    dbg := mg.NewDebugReactor( a )
    rct := b.NewUnbindReactor( mustUnbinderForType( ue.Type ) )
    pip := mg.InitReactorPipeline( ps, dbg, rct )
    if err := mg.FeedSource( ue.Source, pip ); err == nil {
        a.Fatalf( "expected error (%T): %s", ue.Error, ue.Error )
    } else {
        if ue.Error == nil { a.Fatal( err ) }
        assertError( ue.Error, err, a )
    }
}

func TestBinderStandardTests( t *testing.T ) {
    pa := assert.NewPathAsserter( t )
    la := pa.StartList()
    for _, bt := range StandardBindTests() {
        switch v := bt.( type ) {
        case *RoundtripTest: callRoundtripTest( v, la )
        case *UnbindErrorTest: callUnbindErrorTest( v, la )
        default: panic( libErrorf( "unhandled test: %T", bt ) )
        }
        la = la.Next()
    }
}
