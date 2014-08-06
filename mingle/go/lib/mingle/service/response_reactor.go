package service

import (
    mgRct "mingle/reactor"
    mg "mingle"
    "mingle/types"
    "bitgirder/pipeline"
    "bitgirder/objpath"
)

type ResponseReactorInterface interface {
    
    StartResult( path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error )
    
    StartError( path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error )
}

type ResponseReactor struct {

    iface ResponseReactorInterface

    proc *proxyProc

    nextFld *mg.Identifier
}

func NewResponseReactor( iface ResponseReactorInterface ) *ResponseReactor {
    return &ResponseReactor{ iface: iface }
}

func ( r *ResponseReactor ) InitializePipeline( pip *pipeline.Pipeline ) {
    pip.Add( types.NewCastReactor( TypeResponse, types.V1Types() ) )
}

func ( r *ResponseReactor ) processProcEvent( ev mgRct.ReactorEvent ) error {
    if err := r.proc.ProcessEvent( ev ); err != nil { return err }
    if r.proc.isDone() { r.proc = nil }
    return nil
}

func ( r *ResponseReactor ) processNextFieldEvent(
    ev mgRct.ReactorEvent ) error {

    defer func() { r.nextFld = nil }()
    if ve, ok := ev.( *mgRct.ValueEvent ); ok {
        if _, isNull := ve.Val.( *mg.Null ); isNull { return nil }
    }
    var rct mgRct.ReactorEventProcessor
    var err error
    switch p := ev.GetPath(); {
    case r.nextFld.Equals( IdResult ): rct, err = r.iface.StartResult( p )
    case r.nextFld.Equals( IdError ): rct, err = r.iface.StartError( p )
    default: panic( libErrorf( "unhandled field: %s", r.nextFld ) )
    }
    if err != nil { return err }
    r.proc = &proxyProc{ proc: rct }
    return r.ProcessEvent( ev )
}

func ( r *ResponseReactor ) startField( fse *mgRct.FieldStartEvent ) error {
    r.nextFld = fse.Field
    return nil
}

func ( r *ResponseReactor ) ProcessEvent( ev mgRct.ReactorEvent ) error {
    if r.proc != nil { return r.processProcEvent( ev ) }
    if r.nextFld != nil { return r.processNextFieldEvent( ev ) }
    switch v := ev.( type ) {
    case *mgRct.StructStartEvent: return nil
    case *mgRct.FieldStartEvent: return r.startField( v )
    case *mgRct.EndEvent: return nil
    }
    panic( libErrorf( "unhandled event: %T", ev ) )
}
