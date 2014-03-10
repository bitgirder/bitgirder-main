package bind

import (
    mg "mingle"
)

type BinderContext struct {}

func NewBinderContext() *BinderContext { return &BinderContext{} }

type UnbindReactor interface {
    
    mg.ReactorEventProcessor

    GoValue() ( interface{}, error )
}

type Binder interface {
    
    BindToReactor( goVal interface{}, 
                   rct mg.ReactorEventProcessor, 
                   ctx *BinderContext ) error

    NewUnbindReactor( ctx *BinderContext ) UnbindReactor
}
