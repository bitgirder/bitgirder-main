package service

import (
    "fmt"
    mg "mingle"
    "mingle/types"
    "mingle/types/builtin"
    "bitgirder/objpath"
)

var (
    NsService *mg.Namespace

    QnameRequest *mg.QualifiedTypeName
    TypeRequest *mg.AtomicTypeReference

    QnameResponse *mg.QualifiedTypeName
    TypeResponse *mg.AtomicTypeReference

    QnameRequestError *mg.QualifiedTypeName
    TypeRequestError *mg.AtomicTypeReference

    QnameResponseError *mg.QualifiedTypeName
    TypeResponseError *mg.AtomicTypeReference

    // pkg-only for the moment, though could become public if needed; only used
    // as a synthetic struct type for request parameter casts at the moment
    qnameRequestParameters *mg.QualifiedTypeName
    typeRequestParameters *mg.AtomicTypeReference

    IdNamespace *mg.Identifier
    IdService *mg.Identifier
    IdOperation *mg.Identifier
    IdAuthentication *mg.Identifier
    IdParameters *mg.Identifier
    IdResult *mg.Identifier
    IdError *mg.Identifier

    externalErrorTypes = mg.NewQnameMap()
)

type RequestContext struct {
    Namespace *mg.Namespace
    Service *mg.Identifier
    Operation *mg.Identifier
}

// assumes that NsService has been initialized
func initNamePair ( 
    nm string ) ( *mg.QualifiedTypeName, *mg.AtomicTypeReference ) {

    qn := &mg.QualifiedTypeName{
        Namespace: NsService,
        Name: mg.NewDeclaredTypeNameUnsafe( nm ),
    }
    return qn, qn.AsAtomicType()
}

func initNames() {
    mkId := func( parts ...string ) *mg.Identifier {
        return mg.NewIdentifierUnsafe( parts )
    }
    NsService = &mg.Namespace{
        Parts: []*mg.Identifier{ mkId( "mingle" ), mkId( "service" ) },
        Version: mkId( "v1" ),
    }
    QnameRequest, TypeRequest = initNamePair( "Request" )
    QnameResponse, TypeResponse = initNamePair( "Response" )
    QnameRequestError, TypeRequestError = initNamePair( "RequestError" )
    QnameResponseError, TypeResponseError = initNamePair( "ResponseError" )
    qnameRequestParameters, typeRequestParameters = 
        initNamePair( "RequestParameters" )
    IdNamespace = mkId( "namespace" )
    IdService = mkId( "service" )
    IdOperation = mkId( "operation" )
    IdAuthentication = mkId( "authentication" )
    IdParameters = mkId( "parameters" )
    IdResult = mkId( "result" )
    IdError = mkId( "error" )
}

func initExternalErrorTypes() {
    externalErrorTypes.Put( QnameRequestError, true )
}

func initReqType() {
    sd := types.NewStructDefinition()
    sd.Name = QnameRequest
    addFld := func( nm *mg.Identifier, typ mg.TypeReference ) {
        sd.Fields.Add( &types.FieldDefinition{ Name: nm, Type: typ } )
    }
    idPtr := mg.NewPointerTypeReference( types.TypeIdentifier )
    addFld( IdNamespace, mg.NewPointerTypeReference( types.TypeNamespace ) )
    addFld( IdService, idPtr )
    addFld( IdOperation, idPtr )
    addFld( IdAuthentication, mg.TypeNullableValue )
    addFld( IdParameters, mg.MustNullableTypeReference( mg.TypeSymbolMap ) )
    types.MustAddBuiltinType( sd )
}

func initRespType() {
    sd := types.NewStructDefinition()
    sd.Name = QnameResponse
    addFld := func( nm *mg.Identifier ) {
        sd.Fields.Add(
            &types.FieldDefinition{ Name: nm, Type: mg.TypeNullableValue } )
    }
    addFld( IdResult )
    addFld( IdError )
    types.MustAddBuiltinType( sd )
}

func initErrType( qn *mg.QualifiedTypeName ) {
    types.MustAddBuiltinType( builtin.NewLocatableErrorDefinition( qn ) )
}

func initTypes() {
    initReqType()
    initRespType()
    initErrType( QnameRequestError )
    initErrType( QnameResponseError )
}

type RequestError struct {
    Path objpath.PathNode
    Message string
}

func ( e *RequestError ) Error() string {
    return mg.FormatError( e.Path, e.Message )
}

func NewRequestError( path objpath.PathNode, msg string ) *RequestError {
    return &RequestError{ Path: path, Message: msg }
}

func NewRequestErrorf( 
    path objpath.PathNode, tmpl string, args ...interface{} ) *RequestError {

    return &RequestError{ Path: path, Message: fmt.Sprintf( tmpl, args... ) }
}

type ResponseError struct { 
    Path objpath.PathNode
    Message string
}

func ( e *ResponseError ) Error() string {
    return mg.FormatError( e.Path, e.Message )
}

func NewResponseError( path objpath.PathNode, msg string ) *ResponseError {
    return &ResponseError{ Path: path, Message: msg }
}

func NewResponseErrorf( 
    path objpath.PathNode, tmpl string, argv ...interface{} ) *ResponseError {

    return &ResponseError{ Path: path, Message: fmt.Sprintf( tmpl, argv... ) }
}

const respErrMsgMultipleResponseFields =
    "response contains both a result and an error"

func FormatInstanceId( ns *mg.Namespace, svc *mg.Identifier ) string {
    return fmt.Sprintf( "%s.%s", ns.ExternalForm(), svc.ExternalForm() )
}

type InstanceMap struct {
    nsMap *mg.NamespaceMap
}

func NewInstanceMap() *InstanceMap {
    return &InstanceMap{ nsMap: mg.NewNamespaceMap() }
}

func ( m *InstanceMap ) getSvcMap( 
    ns *mg.Namespace, create bool ) *mg.IdentifierMap {

    if res, ok := m.nsMap.GetOk( ns ); ok { return res.( *mg.IdentifierMap ) }
    if ! create { return nil }
    res := mg.NewIdentifierMap()
    m.nsMap.Put( ns, res )
    return res
}

func ( m *InstanceMap ) GetOk( 
    ns *mg.Namespace, svc *mg.Identifier ) ( interface{}, *mg.Identifier ) {

    if svcMap := m.getSvcMap( ns, false ); svcMap != nil {
        if res, ok := svcMap.GetOk( svc ); ok { return res, nil }
        return nil, IdService
    }
    return nil, IdNamespace
}

func ( m *InstanceMap ) Put(
    ns *mg.Namespace, svc *mg.Identifier, val interface{} ) {

    m.getSvcMap( ns, true ).Put( svc, val )
}

func newUnknownEndpointError(
    ctx *RequestContext, errFld *mg.Identifier, path objpath.PathNode ) error {

    var tmpl string
    args := make( []interface{}, 0, 2 )
    switch {
    case errFld.Equals( IdNamespace ):
        tmpl = "no services in namespace: %s"
        args = append( args, ctx.Namespace )
    case errFld.Equals( IdService ):
        tmpl = "namespace %s has no service with id: %s"
        args = append( args, ctx.Namespace, ctx.Service )
    case errFld.Equals( IdOperation ):
        tmpl = "service %s has no such operation: %s"
        instId := FormatInstanceId( ctx.Namespace, ctx.Service )
        args = append( args, instId, ctx.Operation )
    default: panic( libErrorf( "unhandled errFld: %s", errFld ) )
    }
    return NewRequestErrorf( path, tmpl, args... )
}

// could make this public if needed at some point
func ( m *InstanceMap ) getRequestValue( 
    ctx *RequestContext, path objpath.PathNode ) ( interface{}, error ) {

    v, miss := m.GetOk( ctx.Namespace, ctx.Service )
    if miss == nil {
        if res, ok := v.( *mg.IdentifierMap ).GetOk( ctx.Operation ); ok {
            return res, nil
        }
        return nil, newUnknownEndpointError( ctx, IdOperation, path )
    }
    return nil, newUnknownEndpointError( ctx, miss, path )
}

const errMsgNoAuthExpected = "service does not accept authentication"

func isExternalErrorType( qn *mg.QualifiedTypeName ) bool {
    return externalErrorTypes.HasKey( qn )
}
