package bind

import (
    mgRct "mingle/reactor"
)

type FunctionBinderValueFunction func( 
    ve *mgRct.ValueEvent ) ( interface{}, error )

type FunctionBinderValueAttempter func(
    ve *mgRct.ValueEvent ) ( interface{}, error, bool )

func DefaultValueBinderFunction( 
    ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {

    return DefaultBindingForValue( ve.Val ), nil, true
}

func AsSequentialValueFunction( 
    funcs ...FunctionBinderValueAttempter ) FunctionBinderValueFunction {

    if funcs == nil { funcs = []FunctionBinderValueAttempter{} }
    return func( ve *mgRct.ValueEvent ) ( interface{}, error ) {
        for _, f := range funcs {
            if res, err, ok := f( ve ); ok { return res, err }
        }
        return nil, failBinderType( ve )
    }
}
        
type FunctionBinderFactory struct {
    Value FunctionBinderValueFunction
}

func ( f *FunctionBinderFactory ) BindValue( 
    ve *mgRct.ValueEvent ) ( interface{}, error ) {
    
    if f.Value == nil { return nil, failBinderType( ve ) }
    return f.Value( ve )
}

func ( f *FunctionBinderFactory ) StartList(
    lse *mgRct.ListStartEvent ) ( ListBinder, error ) {

    return nil, failBinderType( lse )
}

func ( f *FunctionBinderFactory ) StartMap(
    mse *mgRct.MapStartEvent ) ( FieldSetBinder, error ) {

    return nil, failBinderType( mse )
}

func ( f *FunctionBinderFactory ) StartStruct(
    sse *mgRct.StructStartEvent ) ( FieldSetBinder, error ) {

    return nil, failBinderType( sse )
}
