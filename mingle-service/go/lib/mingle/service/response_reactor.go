package service

import (
    mgRct "mingle/reactor"
    mg "mingle"
    "mingle/cast"
    "mingle/types/builtin"
    "bitgirder/pipeline"
    "bitgirder/objpath"
//    "log"
)

type ResponseReactorInterface interface {
    
    StartResult( path objpath.PathNode ) ( mgRct.EventProcessor, error )
    
    StartError( path objpath.PathNode ) ( mgRct.EventProcessor, error )
}

type ResponseReactor struct {

    iface ResponseReactorInterface

    proc *proxyProc

    nextFld *mg.Identifier

    sawNonNullField bool
}

func NewResponseReactor( iface ResponseReactorInterface ) *ResponseReactor {
    return &ResponseReactor{ iface: iface }
}

func ( r *ResponseReactor ) InitializePipeline( pip *pipeline.Pipeline ) {
    cr := cast.NewReactor( TypeResponse, builtin.BuiltinTypes() )
    cr.AddPassthroughField( QnameResponse, IdResult )
    cr.AddPassthroughField( QnameResponse, IdError )
    pip.Add( cr )
}

func ( r *ResponseReactor ) processProcEvent( ev mgRct.Event ) error {
    if err := r.proc.ProcessEvent( ev ); err != nil { return err }
    if r.proc.isDone() { r.proc = nil }
    return nil
}

// returns ( skip, err ) pair, where non-nil err should be returned to the even
// processor, and a skip value of true indicates that the caller should ignore
// ev entirely for the field
func ( r *ResponseReactor ) checkNextFieldEventOk( 
    ev mgRct.Event ) ( bool, error ) {
    
    if ve, ok := ev.( *mgRct.ValueEvent ); ok {
        if _, skip := ve.Val.( *mg.Null ); skip { return true, nil }
    } 
    if r.sawNonNullField {
        err := NewResponseError( objpath.ParentOf( ev.GetPath() ),
            respErrMsgMultipleResponseFields )
        return false, err
    }
    return false, nil
}

func ( r *ResponseReactor ) processNextFieldEvent(
    ev mgRct.Event ) error {

    defer func() { r.nextFld = nil }()
    skip, err := r.checkNextFieldEventOk( ev )
    if err != nil { return err }
    if skip { return nil } else { r.sawNonNullField = true }
    var rct mgRct.EventProcessor
    switch p := ev.GetPath(); {
    case r.nextFld.Equals( IdResult ): rct, err = r.iface.StartResult( p )
    case r.nextFld.Equals( IdError ): rct, err = r.iface.StartError( p )
    default: panic( libErrorf( "unhandled field: %s", r.nextFld ) )
    }
    if err != nil { return err }
    r.proc = newProxyProc( rct )
    return r.ProcessEvent( ev )
}

func ( r *ResponseReactor ) startField( fse *mgRct.FieldStartEvent ) error {
    r.nextFld = fse.Field
    return nil
}

func ( r *ResponseReactor ) ProcessEvent( ev mgRct.Event ) error {
    if r.proc != nil { return r.processProcEvent( ev ) }
    if r.nextFld != nil { return r.processNextFieldEvent( ev ) }
    switch v := ev.( type ) {
    case *mgRct.StructStartEvent: return nil
    case *mgRct.FieldStartEvent: return r.startField( v )
    case *mgRct.EndEvent: return nil
    }
    panic( libErrorf( "unhandled event: %T", ev ) )
}

func InitResponseReactorPipeline( 
    iface ResponseReactorInterface ) mgRct.EventProcessor {

    return mgRct.InitReactorPipeline( NewResponseReactor( iface ) )
}
