package bind

import (
    mg "mingle"
    "testing"
    "bitgirder/assert"
    "bitgirder/objpath"
)

type valueBindingImpl struct { val mg.Value }

func ( vb valueBindingImpl ) BoundValue() ( mg.Value, error ) { 
    return vb.val, nil
}

func bindingForValue( val interface{} ) interface{} {
    switch v := val.( type ) {
    case int32: return valueBindingImpl{ mg.MustValue( v ) }
    }
    panic( libErrorf( "unhandled test value: %T", val ) )
}

type valUnbinderImpl struct { 
    typ mg.TypeReference 
    val interface{}
}

func ( ub *valUnbinderImpl ) UnboundValue( 
    path objpath.PathNode ) ( interface{}, error ) { 

    if ub.val == nil { return nil, NewUnbindError( path, "no value built" ) }
    return ub.val, nil
}

func ( ub *valUnbinderImpl ) UnbindValue(
    val mg.Value, path objpath.PathNode ) error {

    switch v := val.( type ) {
    case mg.Int32: ub.val = int32( v ); return nil
    }
    return NewUnbindErrorf( path, "unhandled value: %T", val )
}

func unbinderForType( typ mg.TypeReference ) Unbinder {
    switch {
    case typ.Equals( mg.TypeInt32 ): return &valUnbinderImpl{ typ: typ }
    }
    panic( libErrorf( "unhandled unbinder type: %s", typ ) )
}

type roundtripTestCall struct {
    *assert.PathAsserter
    rt *RoundtripTest
    binder *Binder
}

func ( c *roundtripTestCall ) bindToValue() mg.Value {
    vb := mg.NewValueBuilder()
    binding := bindingForValue( c.rt.Object )
    if err := c.binder.Bind( binding, vb ); err != nil { c.Fatal( err ) }
    return vb.GetValue()
}

func ( c *roundtripTestCall ) unbindFromValue( val mg.Value ) interface{} {
    ub := unbinderForType( c.rt.Type )
    rct := c.binder.NewUnbindReactor( ub )
    if err := mg.VisitValue( val, rct ); err != nil { c.Fatal( err ) }
    c.True( rct.HasValue() )
    return rct.UnboundValue()
}

func ( c *roundtripTestCall ) call() {
    c.binder = NewBinder()
    val := c.bindToValue()
    goVal2 := c.unbindFromValue( val )
    c.Equal( c.rt.Object, goVal2 )
}

func callRoundtripTest( rt *RoundtripTest, a *assert.PathAsserter ) {
    ( &roundtripTestCall{ rt: rt, PathAsserter: a } ).call()
}

func TestBinderStandardTests( t *testing.T ) {
    pa := assert.NewPathAsserter( t )
    la := pa.StartList()
    for _, bt := range StandardBindTests() {
        switch v := bt.( type ) {
        case *RoundtripTest: callRoundtripTest( v, la )
        default: panic( libErrorf( "unhandled test: %T", bt ) )
        }
        la = la.Next()
    }
}
