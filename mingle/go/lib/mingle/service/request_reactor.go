package service

import (
    mgRct "mingle/reactor"
    mg "mingle"
    "mingle/types"
    "mingle/bind"
    "bitgirder/pipeline"
    "bitgirder/objpath"
//    "log"
)

type reqOrderSlice mgRct.FieldOrder

var reqOrder reqOrderSlice

func initReqFieldOrder() {
    reqOrder = reqOrderSlice(
        mgRct.FieldOrder(
            []mgRct.FieldOrderSpecification{
                { Field: IdNamespace, Required: true },
                { Field: IdService, Required: true },
                { Field: IdOperation, Required: true },
                { Field: IdAuthentication, Required: false },
                { Field: IdParameters, Required: false },
            },
        ),
    )
}

func ( o reqOrderSlice ) FieldOrderFor( 
    qn *mg.QualifiedTypeName ) mgRct.FieldOrder {

    if ! qn.Equals( QnameRequest ) { return nil }
    return mgRct.FieldOrder( o )
}

type RequestReactorInterface interface {

    StartRequest( ctx *RequestContext, path objpath.PathNode ) error

    StartAuthentication( 
        path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) 

    StartParameters( 
        path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) 
}

type RequestReactor struct {

    iface RequestReactorInterface

    reg *bind.Registry

    reqCtx *RequestContext

    reqStartPath objpath.PathNode

    // present when we read a reqCtx field
    bldr *mgRct.BuildReactor
    bldrSetter func( val interface{} ) error

    // present when we send events to a downstream processor (auth, params)
    proc *proxyProc

    sawParams bool
}

func NewRequestReactor( iface RequestReactorInterface ) *RequestReactor {
    return &RequestReactor{ 
        iface: iface,
        reqCtx: &RequestContext{},
        reg: bind.RegistryForDomain( bind.DomainDefault ),
    }
}

func ( r *RequestReactor ) processBuilderEvent( ev mgRct.ReactorEvent ) error {
    if err := r.bldr.ProcessEvent( ev ); err != nil { return err }
    if r.bldr.HasValue() {
        if err := r.bldrSetter( r.bldr.GetValue() ); err != nil { return err }
        r.bldr, r.bldrSetter = nil, nil
    }
    return nil
}

func ( r *RequestReactor ) InitializePipeline( pip *pipeline.Pipeline ) {
    cr := types.NewCastReactor( TypeRequest, types.V1Types() )
    cr.AddPassthroughField( QnameRequest, IdParameters )
    cr.AddPassthroughField( QnameRequest, IdAuthentication )
    pip.Add( cr )
    pip.Add( mgRct.NewFieldOrderReactor( reqOrder ) )
}

func ( r *RequestReactor ) setBuilder(
    fld *mg.Identifier, 
    typ mg.TypeReference, 
    setter func( val interface{} ) error ) {

    bf := r.reg.MustBuilderFactoryForType( typ )
    r.bldr = mgRct.NewBuildReactor( bf )
    r.bldrSetter = setter
}

func ( r *RequestReactor ) startNamespace() {
    r.setBuilder(
        IdNamespace,
        types.TypeNamespace,
        func( val interface{} ) error { 
            r.reqCtx.Namespace = val.( *mg.Namespace )
            return nil
        },
    )
}

func ( r *RequestReactor ) startService() {
    r.setBuilder(
        IdService,
        types.TypeIdentifier,
        func( val interface{} ) error { 
            r.reqCtx.Service = val.( *mg.Identifier )
            return nil
        },
    )
}

func ( r *RequestReactor ) startOperation() {
    r.setBuilder(
        IdOperation,
        types.TypeIdentifier,
        func( val interface{} ) error { 
            r.reqCtx.Operation = val.( *mg.Identifier )
            return r.iface.StartRequest( r.reqCtx, r.reqStartPath )
        },
    )
}

func ( r *RequestReactor ) startProc( 
    rct mgRct.ReactorEventProcessor, startErr error ) error {

    if startErr != nil { return startErr }
    r.proc = newProxyProc( rct )
    return nil
}

func ( r *RequestReactor ) processProcEvent( ev mgRct.ReactorEvent ) error {
    if err := r.proc.ProcessEvent( ev ); err != nil { return err }
    if r.proc.isDone() { r.proc = nil }
    return nil
}

func ( r *RequestReactor ) startField( fse *mgRct.FieldStartEvent ) error {
    switch fld := fse.Field; {
    case fld.Equals( IdNamespace ): r.startNamespace()
    case fld.Equals( IdService ): r.startService()
    case fld.Equals( IdOperation ): r.startOperation()
    case fld.Equals( IdAuthentication ): 
        return r.startProc( r.iface.StartAuthentication( fse.GetPath() ) )
    case fld.Equals( IdParameters ): 
        r.sawParams = true
        return r.startProc( r.iface.StartParameters( fse.GetPath() ) )
    default: panic( libErrorf( "unexpected field: %s", fld ) )
    }
    return nil
}

func ( r *RequestReactor ) endRequest( ee *mgRct.EndEvent ) error {
    if r.sawParams { return nil }
    path := objpath.Descend( ee.GetPath(), IdParameters )
    rct, err := r.iface.StartParameters( path )
    if err != nil { return err }
    return mgRct.VisitValuePath( mg.EmptySymbolMap(), rct, path )
}

func ( r *RequestReactor ) ProcessEvent( ev mgRct.ReactorEvent ) error {
    if r.bldr != nil { return r.processBuilderEvent( ev ) }
    if r.proc != nil { return r.processProcEvent( ev ) }
    switch v := ev.( type ) {
    case *mgRct.StructStartEvent: 
        r.reqStartPath = objpath.CopyOf( ev.GetPath() )
        return nil;
    case *mgRct.FieldStartEvent: return r.startField( v )
    case *mgRct.EndEvent: return r.endRequest( v )
    }
    panic( libErrorf( "unhandled event: %T", ev ) )
}
