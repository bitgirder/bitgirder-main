package bind

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "bitgirder/objpath"
)

const defaultMapKeyType = "$type"

type defaultBinderFactory int

var DefaultBinderFactory BinderFactory = defaultBinderFactory( 1 )

func DefaultBindingForValue( val mg.Value ) interface{} {
    switch v := val.( type ) {
    case mg.Boolean: return bool( v )
    case mg.Buffer: return []byte( v )
    case mg.String: return string( v )
    case mg.Int32: return int32( v )
    case mg.Uint32: return uint32( v )
    case mg.Float32: return float32( v )
    case mg.Int64: return int64( v )
    case mg.Uint64: return uint64( v )
    case mg.Float64: return float64( v )
    }
    return val
}

func ( b defaultBinderFactory ) BindValue( 
    ve *mgRct.ValueEvent ) ( interface{}, error ) {

    return DefaultBindingForValue( ve.Val ), nil
}

type defaultListBinder struct { vals []interface{} }

func ( lb *defaultListBinder ) AddValue(
    val interface{}, path objpath.PathNode ) error {

    lb.vals = append( lb.vals, val )
    return nil
}

func ( lb *defaultListBinder ) ProduceValue(
    ee *mgRct.EndEvent ) ( interface{}, error ) {

    return lb.vals, nil
}

func ( lb *defaultListBinder ) NextBinderFactory() BinderFactory {
    return DefaultBinderFactory
}

func ( b defaultBinderFactory ) StartList(
    lse *mgRct.ListStartEvent ) ( ListBinder, error ) {

    return &defaultListBinder{ vals: make( []interface{}, 0, 4 ) }, nil
}

type defaultFieldSetBinder map[ string ]interface{}

func ( b defaultFieldSetBinder ) StartField(
    fse *mgRct.FieldStartEvent ) ( BinderFactory, error ) {

    return DefaultBinderFactory, nil
}

func ( b defaultFieldSetBinder ) SetValue(
    fld *mg.Identifier, val interface{}, path objpath.PathNode ) error {

    b[ fld.ExternalForm() ] = val
    return nil
}

func ( b defaultFieldSetBinder ) ProduceValue( 
    ee *mgRct.EndEvent ) ( interface{}, error ) {

    return map[ string ]interface{}( b ), nil
}

func newDefaultFieldSetBinder() defaultFieldSetBinder {
    return defaultFieldSetBinder( make( map[ string ]interface{}, 4 ) )
}

func ( b defaultBinderFactory ) StartMap( 
    mse *mgRct.MapStartEvent ) ( FieldSetBinder, error ) {

    return newDefaultFieldSetBinder(), nil
}

func ( b defaultBinderFactory ) StartStruct(
    sse *mgRct.StructStartEvent ) ( FieldSetBinder, error ) {

    res := newDefaultFieldSetBinder()
    res[ defaultMapKeyType ] = sse.Type
    return res, nil
}
