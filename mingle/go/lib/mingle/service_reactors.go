package mingle

import (
    "bitgirder/objpath"
    "bitgirder/pipeline"
    "bytes"
//    "log"
)

type ServiceRequestReactorInterface interface {

    Namespace( ns *Namespace, path objpath.PathNode ) error

    Service( svc *Identifier, path objpath.PathNode ) error

    Operation( op *Identifier, path objpath.PathNode ) error

    GetAuthenticationProcessor( 
        path objpath.PathNode ) ( ReactorEventProcessor, error )

    GetParametersProcessor( 
        path objpath.PathNode ) ( ReactorEventProcessor, error )
}

type requestFieldType int

// declared in the preferred arrival order
const (
    reqFieldNone = requestFieldType( iota )
    reqFieldNs
    reqFieldSvc
    reqFieldOp
    reqFieldAuth
    reqFieldParams
)

type ServiceRequestReactor struct {

    iface ServiceRequestReactorInterface

    evProc ReactorEventProcessor

    // 0: before StartStruct{ QnameServiceRequest } and after final *EndEvent
    //
    // 1: when reading or expecting a service request field (namespace, service,
    // etc)
    //
    // > 1: accumulating some nested value for 'parameters' or 'authentication' 
    depth int 

    fld requestFieldType

    hadParams bool // true if the input contained explicit params
//    paramsSynth bool // true when we are synthesizing empty params
}

func NewServiceRequestReactor( 
    iface ServiceRequestReactorInterface ) *ServiceRequestReactor {
    return &ServiceRequestReactor{ iface: iface }
}

func ( sr *ServiceRequestReactor ) updateEvProc( ev ReactorEvent ) {
    switch ev.( type ) {
    case *FieldStartEvent: return
    case *StructStartEvent, *ListStartEvent, *MapStartEvent: sr.depth++
    case *EndEvent: sr.depth--
    }
    if sr.depth == 1 { sr.evProc, sr.fld = nil, reqFieldNone } 
}

type svcReqCastIface int

func ( c svcReqCastIface ) InferStructFor( qn *QualifiedTypeName ) bool {
    return qn.Equals( QnameServiceRequest )
}

type svcReqFieldTyper int

func ( t svcReqFieldTyper ) FieldTypeFor( 
    fld *Identifier, path objpath.PathNode ) ( TypeReference, error ) {

    if fld.Equals( IdParameters ) { return TypeSymbolMap, nil }
    return TypeValue, nil
}

func ( c svcReqCastIface ) FieldTyperFor( 
    qn *QualifiedTypeName, path objpath.PathNode ) ( FieldTyper, error ) {
    
    if qn.Equals( QnameServiceRequest ) { return svcReqFieldTyper( 1 ), nil }
    return nil, nil
}

func ( c svcReqCastIface ) CastAtomic( 
    in Value, 
    at *AtomicTypeReference, 
    path objpath.PathNode ) ( Value, error, bool ) {

    return nil, nil, false
}

type svcReqFieldOrderGetter int

func ( g svcReqFieldOrderGetter ) FieldOrderFor( 
    qn *QualifiedTypeName ) FieldOrder {
    if qn.Equals( QnameServiceRequest ) { return svcReqFieldOrder }
    return nil
}

func ( sr *ServiceRequestReactor ) InitializePipeline( 
    pip *pipeline.Pipeline ) {
    
    EnsureStructuralReactor( pip ) 
    EnsurePathSettingProcessor( pip )
    pip.Add( NewCastReactor( TypeServiceRequest, svcReqCastIface( 1 ) ) )
    pip.Add( NewFieldOrderReactor( svcReqFieldOrderGetter( 1 ) ) )
}

func ( sr *ServiceRequestReactor ) invalidValueErr( 
    path objpath.PathNode, desc string ) error {

    return NewValueCastErrorf( path, "invalid value: %s", desc )
}

func ( sr *ServiceRequestReactor ) startStruct( ev *StructStartEvent ) error {
    if sr.fld == reqFieldNone { // we're at the top of the request
        if ev.Type.Equals( QnameServiceRequest ) { return nil }
        // panic because upstream cast should have checked already
        panic( libErrorf( "Unexpected service request type: %s", ev.Type ) )
    }
    return sr.invalidValueErr( ev.GetPath(), ev.Type.ExternalForm() )
}

func ( sr *ServiceRequestReactor ) checkStartField( fs *FieldStartEvent ) {
    if sr.fld == reqFieldNone { return }
    panic( libErrorf( "Saw field start '%s' while sr.fld is %d", 
        fs.Field, sr.fld ) )
}

func ( sr *ServiceRequestReactor ) startField( 
    fs *FieldStartEvent ) ( err error ) {

    sr.checkStartField( fs )
    switch fld := fs.Field; {
    case fld.Equals( IdNamespace ): sr.fld = reqFieldNs
    case fld.Equals( IdService ): sr.fld = reqFieldSvc
    case fld.Equals( IdOperation ): sr.fld = reqFieldOp
    case fld.Equals( IdAuthentication ): 
        sr.fld = reqFieldAuth
        sr.evProc, err = sr.iface.GetAuthenticationProcessor( fs.GetPath() )
    case fld.Equals( IdParameters ): 
        sr.fld = reqFieldParams
        sr.evProc, err = sr.iface.GetParametersProcessor( fs.GetPath() )
        if err == nil { sr.hadParams = true }
    default: err = NewUnrecognizedFieldError( fs.GetPath(), fs.Field )
    }
    return
}

func ( sr *ServiceRequestReactor ) getFieldValueForString(
    s string, 
    path objpath.PathNode, 
    reqFld requestFieldType ) ( res interface{}, err error ) {

    switch reqFld {
    case reqFieldNs: res, err = ParseNamespace( s )
    case reqFieldSvc, reqFieldOp: res, err = ParseIdentifier( s )
    default:
        panic( libErrorf( "Unhandled req fld type for string: %d", reqFld ) )
    }
    if err != nil { err = NewValueCastError( path, err.Error() ) }
    return
}

func ( sr *ServiceRequestReactor ) getFieldValueForBuffer(
    buf []byte, 
    path objpath.PathNode,
    reqFld requestFieldType ) ( res interface{}, err error ) {

    bin := NewReader( bytes.NewReader( buf ) )
    switch reqFld {
    case reqFieldNs: res, err = bin.ReadNamespace()
    case reqFieldSvc, reqFieldOp: res, err = bin.ReadIdentifier()
    default:
        panic( libErrorf( "Unhandled req fld type for buffer: %d", reqFld ) )
    }
    if err != nil { err = NewValueCastError( path, err.Error() ) }
    return
}

func ( sr *ServiceRequestReactor ) getFieldValue( 
    ve *ValueEvent, reqFld requestFieldType ) ( interface{}, error ) {
    path := ve.GetPath()
    switch v := ve.Val.( type ) {
    case String: return sr.getFieldValueForString( string( v ), path, reqFld )
    case Buffer: return sr.getFieldValueForBuffer( []byte( v ), path, reqFld )
    }
    return nil, sr.invalidValueErr( path, TypeOf( ve.Val ).ExternalForm() )
}

func ( sr *ServiceRequestReactor ) namespace( ve *ValueEvent ) error {
    ns, err := sr.getFieldValue( ve, reqFieldNs )
    if err == nil { 
        return sr.iface.Namespace( ns.( *Namespace ), ve.GetPath() )
    }
    return err
}

func ( sr *ServiceRequestReactor ) readIdent( 
    ve *ValueEvent, reqFld requestFieldType ) error {

    v2, err := sr.getFieldValue( ve, reqFld )
    if err != nil { return err }
    id := v2.( *Identifier )
    path := ve.GetPath()
    switch reqFld {
    case reqFieldSvc: return sr.iface.Service( id, path )
    case reqFieldOp: return sr.iface.Operation( id, path )
    default: panic( libErrorf( "Unhandled req fld type: %d", reqFld ) )
    }
    return nil
}

func ( sr *ServiceRequestReactor ) value( ve *ValueEvent ) error {
    defer func() { sr.fld = reqFieldNone }()
    switch sr.fld {
    case reqFieldNs: return sr.namespace( ve )
    case reqFieldSvc, reqFieldOp: return sr.readIdent( ve, sr.fld )
    }
    panic( libErrorf( "Unhandled req field type: %d", sr.fld ) )
}

func ( sr *ServiceRequestReactor ) visitSyntheticParams(
    rct ReactorEventProcessor, startPath objpath.PathNode ) error {
    ps := NewPathSettingProcessor()
    ps.skipStructureCheck = true
    var path objpath.PathNode
    if startPath == nil { 
        path = objpath.RootedAt( IdParameters ) 
    } else { 
        path = startPath.Descend( IdParameters ) 
    }
    ps.SetStartPath( path )
    pip := InitReactorPipeline( ps, rct )
    return VisitValue( EmptySymbolMap(), pip )
}

func ( sr *ServiceRequestReactor ) end( ee *EndEvent ) error {
    if sr.hadParams { return nil }
    ep, err := sr.iface.GetParametersProcessor( ee.GetPath() );
    if err != nil { return err }
    return sr.visitSyntheticParams( ep, ee.GetPath() )
}

func ( sr *ServiceRequestReactor ) ProcessEvent( ev ReactorEvent ) error {
    defer sr.updateEvProc( ev )
    if sr.evProc != nil { return sr.evProc.ProcessEvent( ev ) }
    switch v := ev.( type ) {
    case *FieldStartEvent: return sr.startField( v )
    case *StructStartEvent: return sr.startStruct( v )
    case *ValueEvent: return sr.value( v )
    case *ListStartEvent: 
        return sr.invalidValueErr( v.GetPath(), TypeOpaqueList.ExternalForm() )
    case *MapStartEvent: 
        return sr.invalidValueErr( v.GetPath(), TypeSymbolMap.ExternalForm() )
    case *EndEvent: return sr.end( v )
    default: panic( libErrorf( "Unhandled event: %T", v ) )
    }
    return nil
}

type ServiceResponseReactorInterface interface {
    GetResultProcessor( path objpath.PathNode ) ( ReactorEventProcessor, error )
    GetErrorProcessor( path objpath.PathNode ) ( ReactorEventProcessor, error )
}

type ServiceResponseReactor struct {

    iface ServiceResponseReactorInterface

    evProc ReactorEventProcessor

    // depth is similar to in ServiceRequestReactor
    depth int
    
    hadProc bool
}

func NewServiceResponseReactor( 
    iface ServiceResponseReactorInterface ) *ServiceResponseReactor {
    return &ServiceResponseReactor{ iface: iface }
}

type svcRespCastIface int

func ( i svcRespCastIface ) InferStructFor( qn *QualifiedTypeName ) bool {
    return qn.Equals( QnameServiceResponse )
}

func ( i svcRespCastIface ) FieldTyperFor( 
    qn *QualifiedTypeName, path objpath.PathNode ) ( FieldTyper, error ) {

    return valueFieldTyper( 1 ), nil
}

func ( i svcRespCastIface ) CastAtomic(
    in Value, 
    at *AtomicTypeReference, 
    path objpath.PathNode ) ( Value, error, bool ) {

    return nil, nil, false
}

func ( sr *ServiceResponseReactor ) InitializePipeline( 
    pip *pipeline.Pipeline ) {

    EnsureStructuralReactor( pip )
    pip.Add( NewCastReactor( TypeServiceResponse, svcRespCastIface( 1 ) ) )
}

func ( sr *ServiceResponseReactor ) updateEvProc( ev ReactorEvent ) {
    switch ev.( type ) {
    case *FieldStartEvent: return
    case *StructStartEvent, *MapStartEvent, *ListStartEvent: sr.depth++
    case *EndEvent: sr.depth--
    }
    if sr.depth == 1 { 
        if sr.evProc != nil { sr.hadProc, sr.evProc = true, nil }
    }
}

// Note that the error path uses Parent() since we'll be positioned on the field
// (result/error) that is the second value, but the error, if we have one, is
// really at the response level itself
func ( sr *ServiceResponseReactor ) sendEvProcEvent( ev ReactorEvent ) error {
    shouldFail := sr.hadProc
    if shouldFail {
        if ve, ok := ev.( *ValueEvent ); ok {
            if _, isNull := ve.Val.( *Null ); isNull { shouldFail = false }
        }
    }
    if shouldFail {
        msg := "response has both a result and an error value"
        return NewValueCastError( ev.GetPath().Parent(), msg )
    }
    return sr.evProc.ProcessEvent( ev )
}

func ( sr *ServiceResponseReactor ) startStruct( t *QualifiedTypeName ) error {
    if t.Equals( QnameServiceResponse ) { return nil }
    panic( libErrorf( "Got unexpected (toplevel) struct type: %s", t ) )
}

func ( sr *ServiceResponseReactor ) startField( fs *FieldStartEvent ) error {
    var err error
    fld, path := fs.Field, fs.GetPath()
    switch {
    case fld.Equals( IdResult ): 
        sr.evProc, err = sr.iface.GetResultProcessor( path )
    case fld.Equals( IdError ): 
        sr.evProc, err = sr.iface.GetErrorProcessor( path )
    default: return NewUnrecognizedFieldError( path.Parent(), fld )
    }
    return err
}

func ( sr *ServiceResponseReactor ) ProcessEvent( ev ReactorEvent ) error {
    defer sr.updateEvProc( ev )
    if sr.evProc != nil { return sr.sendEvProcEvent( ev ) }
    switch v := ev.( type ) {
    case *StructStartEvent: return sr.startStruct( v.Type )
    case *FieldStartEvent: return sr.startField( v )
    case *EndEvent: return nil
    }
    evStr := EventToString( ev )
    panic( libErrorf( "Saw event %s while evProc == nil", evStr ) )
}
