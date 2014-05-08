package reactor

import (
    mg "mingle"
    "bitgirder/objpath"
    "bitgirder/pipeline"
    "bytes"
//    "log"
)

var ServiceRequestFieldOrder FieldOrder 

func init() {
    ServiceRequestFieldOrder = FieldOrder(
        []FieldOrderSpecification{
            { mg.IdNamespace, true },
            { mg.IdService, true },
            { mg.IdOperation, true },
            { mg.IdAuthentication, false },
            { mg.IdParameters, false },
        },
    )
}

type RequestReactorInterface interface {

    Namespace( ns *mg.Namespace, path objpath.PathNode ) error

    Service( svc *mg.Identifier, path objpath.PathNode ) error

    Operation( op *mg.Identifier, path objpath.PathNode ) error

    GetAuthenticationReactor( 
        path objpath.PathNode ) ( ReactorEventProcessor, error )

    GetParametersReactor( 
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

type RequestReactor struct {

    iface RequestReactorInterface

    evProc ReactorEventProcessor

    // 0: before StartStruct{ mg.QnameRequest } and after final *EndEvent
    //
    // 1: when reading or expecting a toplevel request field (namespace,
    // service, etc)
    //
    // > 1: accumulating some nested value for 'parameters' or 'authentication' 
    depth int 

    fld requestFieldType

    hadParams bool // true if the input contained explicit params
}

func NewRequestReactor( 
    iface RequestReactorInterface ) *RequestReactor {

    return &RequestReactor{ iface: iface }
}

func ( sr *RequestReactor ) updateDepth( ev ReactorEvent ) {
    switch ev.( type ) {
    case *FieldStartEvent: return
    case *StructStartEvent, *ListStartEvent, *MapStartEvent: sr.depth++
    case *EndEvent: sr.depth--
    }
    if sr.depth == 1 { sr.evProc, sr.fld = nil, reqFieldNone } 
}

type svcReqCastIface int

func ( c svcReqCastIface ) InferStructFor( qn *mg.QualifiedTypeName ) bool {
    return qn.Equals( mg.QnameRequest )
}

func ( c svcReqCastIface ) AllowAssignment( 
    expct, act *mg.QualifiedTypeName ) bool {

    return false
}

type svcReqFieldTyper int

func ( t svcReqFieldTyper ) FieldTypeFor( 
    fld *mg.Identifier, path objpath.PathNode ) ( mg.TypeReference, error ) {

    if fld.Equals( mg.IdParameters ) { return mg.TypeSymbolMap, nil }
    return mg.TypeNullableValue, nil
}

func ( c svcReqCastIface ) FieldTyperFor( 
    qn *mg.QualifiedTypeName, path objpath.PathNode ) ( FieldTyper, error ) {
    
    if qn.Equals( mg.QnameRequest ) { return svcReqFieldTyper( 1 ), nil }
    return nil, nil
}

func ( c svcReqCastIface ) CastAtomic( 
    in mg.Value, 
    at *mg.AtomicTypeReference, 
    path objpath.PathNode ) ( mg.Value, error, bool ) {

    return nil, nil, false
}

type svcReqFieldOrderGetter int

func ( g svcReqFieldOrderGetter ) FieldOrderFor( 
    qn *mg.QualifiedTypeName ) FieldOrder {

    if qn.Equals( mg.QnameRequest ) { return ServiceRequestFieldOrder }
    return nil
}

// note that cast reactor needs to be added ahead of field order reactor since
// the cast reactor may supply the top-level type (TypeRequest) for an input
// which is just a symbol map, and this top-level type is needed to access the
// field order.
func ( sr *RequestReactor ) InitializePipeline( 
    pip *pipeline.Pipeline ) {
    
    EnsureStructuralReactor( pip ) 
    EnsurePathSettingProcessor( pip )
    pip.Add( NewCastReactor( mg.TypeRequest, svcReqCastIface( 1 ) ) )
    pip.Add( NewFieldOrderReactor( svcReqFieldOrderGetter( 1 ) ) )
}

func ( sr *RequestReactor ) invalidValueErr( 
    path objpath.PathNode, desc string ) error {

    return mg.NewValueCastErrorf( path, "invalid value: %s", desc )
}

// top level type (reqFieldNone) should have been checked by upstream cast
// reactor, so we just check that here and panic. otherwise we don't expect a
// struct for any fields that we process so return that as an error
func ( sr *RequestReactor ) startStruct( ev *StructStartEvent ) error {
    if sr.fld == reqFieldNone { // we're at the top of the request
        if ev.Type.Equals( mg.QnameRequest ) { return nil }
        // panic because upstream cast should have checked already
        panic( libErrorf( "Unexpected service request type: %s", ev.Type ) )
    }
    return sr.invalidValueErr( ev.GetPath(), ev.Type.ExternalForm() )
}

func ( sr *RequestReactor ) checkStartField( fs *FieldStartEvent ) {
    if sr.fld == reqFieldNone { return }
    panic( libErrorf( "Saw field start '%s' while sr.fld is %d", 
        fs.Field, sr.fld ) )
}

func ( sr *RequestReactor ) startField( 
    fs *FieldStartEvent ) ( err error ) {

    sr.checkStartField( fs )
    switch fld := fs.Field; {
    case fld.Equals( mg.IdNamespace ): sr.fld = reqFieldNs
    case fld.Equals( mg.IdService ): sr.fld = reqFieldSvc
    case fld.Equals( mg.IdOperation ): sr.fld = reqFieldOp
    case fld.Equals( mg.IdAuthentication ): 
        sr.fld = reqFieldAuth
        sr.evProc, err = sr.iface.GetAuthenticationReactor( fs.GetPath() )
    case fld.Equals( mg.IdParameters ): 
        sr.fld = reqFieldParams
        sr.evProc, err = sr.iface.GetParametersReactor( fs.GetPath() )
        if err == nil { sr.hadParams = true }
    default: err = mg.NewUnrecognizedFieldError( fs.GetPath(), fs.Field )
    }
    return
}

func ( sr *RequestReactor ) getFieldValueForString(
    s string, 
    path objpath.PathNode, 
    reqFld requestFieldType ) ( res interface{}, err error ) {

    switch reqFld {
    case reqFieldNs: res, err = mg.ParseNamespace( s )
    case reqFieldSvc, reqFieldOp: res, err = mg.ParseIdentifier( s )
    default:
        panic( libErrorf( "Unhandled req fld type for string: %d", reqFld ) )
    }
    if err != nil { err = mg.NewValueCastError( path, err.Error() ) }
    return
}

func ( sr *RequestReactor ) getFieldValueForBuffer(
    buf []byte, 
    path objpath.PathNode,
    reqFld requestFieldType ) ( res interface{}, err error ) {

    bin := mg.NewReader( bytes.NewReader( buf ) )
    switch reqFld {
    case reqFieldNs: res, err = bin.ReadNamespace()
    case reqFieldSvc, reqFieldOp: res, err = bin.ReadIdentifier()
    default:
        panic( libErrorf( "Unhandled req fld type for buffer: %d", reqFld ) )
    }
    if err != nil { err = mg.NewValueCastError( path, err.Error() ) }
    return
}

func ( sr *RequestReactor ) getFieldValue( 
    ve *ValueEvent, reqFld requestFieldType ) ( interface{}, error ) {
    path := ve.GetPath()
    switch v := ve.Val.( type ) {
    case mg.String: return sr.getFieldValueForString( string( v ), path, reqFld )
    case mg.Buffer: return sr.getFieldValueForBuffer( []byte( v ), path, reqFld )
    }
    return nil, sr.invalidValueErr( path, mg.TypeOf( ve.Val ).ExternalForm() )
}

func ( sr *RequestReactor ) namespace( ve *ValueEvent ) error {
    ns, err := sr.getFieldValue( ve, reqFieldNs )
    if err == nil { 
        return sr.iface.Namespace( ns.( *mg.Namespace ), ve.GetPath() )
    }
    return err
}

func ( sr *RequestReactor ) readIdent( 
    ve *ValueEvent, reqFld requestFieldType ) error {

    v2, err := sr.getFieldValue( ve, reqFld )
    if err != nil { return err }
    id := v2.( *mg.Identifier )
    path := ve.GetPath()
    switch reqFld {
    case reqFieldSvc: return sr.iface.Service( id, path )
    case reqFieldOp: return sr.iface.Operation( id, path )
    default: panic( libErrorf( "Unhandled req fld type: %d", reqFld ) )
    }
    return nil
}

func ( sr *RequestReactor ) value( ve *ValueEvent ) error {
    defer func() { sr.fld = reqFieldNone }()
    switch sr.fld {
    case reqFieldNs: return sr.namespace( ve )
    case reqFieldSvc, reqFieldOp: return sr.readIdent( ve, sr.fld )
    }
    panic( libErrorf( "Unhandled req field type: %d", sr.fld ) )
}

func ( sr *RequestReactor ) visitSyntheticParams(
    rct ReactorEventProcessor, startPath objpath.PathNode ) error {
    ps := NewPathSettingProcessor()
    ps.SkipStructureCheck = true
    var path objpath.PathNode
    if startPath == nil { 
        path = objpath.RootedAt( mg.IdParameters ) 
    } else { 
        path = startPath.Descend( mg.IdParameters ) 
    }
    ps.SetStartPath( path )
    pip := InitReactorPipeline( ps, rct )
    return VisitValue( mg.EmptySymbolMap(), pip )
}

func ( sr *RequestReactor ) end( ee *EndEvent ) error {
    if sr.hadParams { return nil }
    ep, err := sr.iface.GetParametersReactor( ee.GetPath() );
    if err != nil { return err }
    return sr.visitSyntheticParams( ep, ee.GetPath() )
}

func ( sr *RequestReactor ) ProcessEvent( ev ReactorEvent ) error {
    defer sr.updateDepth( ev )
    if sr.evProc != nil { return sr.evProc.ProcessEvent( ev ) }
    switch v := ev.( type ) {
    case *FieldStartEvent: return sr.startField( v )
    case *StructStartEvent: return sr.startStruct( v )
    case *ValueEvent: return sr.value( v )
    case *ListStartEvent: 
        return sr.invalidValueErr( 
            v.GetPath(), mg.TypeOpaqueList.ExternalForm() )
    case *MapStartEvent: 
        return sr.invalidValueErr( 
            v.GetPath(), mg.TypeSymbolMap.ExternalForm() )
    case *EndEvent: return sr.end( v )
    default: panic( libErrorf( "Unhandled event: %T", v ) )
    }
    return nil
}

type ResponseReactorInterface interface {
    GetResultReactor( path objpath.PathNode ) ( ReactorEventProcessor, error )
    GetErrorReactor( path objpath.PathNode ) ( ReactorEventProcessor, error )
}

type ResponseReactor struct {

    iface ResponseReactorInterface

    evProc ReactorEventProcessor

    // depth is similar to in RequestReactor
    depth int
    
    hadProc bool
}

func NewResponseReactor( 
    iface ResponseReactorInterface ) *ResponseReactor {
    return &ResponseReactor{ iface: iface }
}

type svcRespCastIface int

func ( i svcRespCastIface ) InferStructFor( qn *mg.QualifiedTypeName ) bool {
    return qn.Equals( mg.QnameResponse )
}

func ( i svcRespCastIface ) FieldTyperFor( 
    qn *mg.QualifiedTypeName, path objpath.PathNode ) ( FieldTyper, error ) {

    return valueFieldTyper( 1 ), nil
}

func ( i svcRespCastIface ) AllowAssignment(
    expct, act *mg.QualifiedTypeName ) bool {

    return false
}

func ( i svcRespCastIface ) CastAtomic(
    in mg.Value, 
    at *mg.AtomicTypeReference, 
    path objpath.PathNode ) ( mg.Value, error, bool ) {

    return nil, nil, false
}

func ( sr *ResponseReactor ) InitializePipeline( 
    pip *pipeline.Pipeline ) {

    EnsureStructuralReactor( pip )
    pip.Add( NewCastReactor( mg.TypeResponse, svcRespCastIface( 1 ) ) )
}

func ( sr *ResponseReactor ) updateDepth( ev ReactorEvent ) {
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
func ( sr *ResponseReactor ) sendEvProcEvent( ev ReactorEvent ) error {
    shouldFail := sr.hadProc
    if shouldFail {
        if ve, ok := ev.( *ValueEvent ); ok {
            if _, isNull := ve.Val.( *mg.Null ); isNull { shouldFail = false }
        }
    }
    if shouldFail {
        msg := "response has both a result and an error value"
        return mg.NewValueCastError( ev.GetPath().Parent(), msg )
    }
    return sr.evProc.ProcessEvent( ev )
}

func ( sr *ResponseReactor ) startStruct( t *mg.QualifiedTypeName ) error {
    if t.Equals( mg.QnameResponse ) { return nil }
    panic( libErrorf( "Got unexpected (toplevel) struct type: %s", t ) )
}

func ( sr *ResponseReactor ) startField( fs *FieldStartEvent ) error {
    var err error
    fld, path := fs.Field, fs.GetPath()
    switch {
    case fld.Equals( mg.IdResult ): 
        sr.evProc, err = sr.iface.GetResultReactor( path )
    case fld.Equals( mg.IdError ): 
        sr.evProc, err = sr.iface.GetErrorReactor( path )
    default: return mg.NewUnrecognizedFieldError( path.Parent(), fld )
    }
    return err
}

func ( sr *ResponseReactor ) ProcessEvent( ev ReactorEvent ) error {
    defer sr.updateDepth( ev )
    if sr.evProc != nil { return sr.sendEvProcEvent( ev ) }
    switch v := ev.( type ) {
    case *StructStartEvent: return sr.startStruct( v.Type )
    case *FieldStartEvent: return sr.startField( v )
    case *EndEvent: return nil
    }
    evStr := EventToString( ev )
    panic( libErrorf( "Saw event %s while evProc == nil", evStr ) )
}
