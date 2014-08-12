package service

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "mingle/types"
    "bitgirder/objpath"
//    "log"
)

type RequestDefinition struct {
    Operation *types.OperationDefinition
    AuthenticationType mg.TypeReference
}

func newRequestDefinition( 
    opDef *types.OperationDefinition ) *RequestDefinition {

    return &RequestDefinition{ Operation: opDef }
}

type OperationMap struct {
    instMap *InstanceMap
    defs *types.DefinitionMap
}

func NewOperationMap( defs *types.DefinitionMap ) *OperationMap {
    return &OperationMap{ defs: defs, instMap: NewInstanceMap() }
}

func ( m *OperationMap ) mustAddRequestDefinitions(
    opMaps *mg.IdentifierMap, sd *types.ServiceDefinition ) {

    var authTyp mg.TypeReference
    if secQn := sd.Security; secQn != nil { 
        authTyp = types.MustAuthTypeOf( secQn, m.defs )
    }
    for _, opDef := range sd.Operations {
        reqDef := newRequestDefinition( opDef )
        if authTyp != nil { reqDef.AuthenticationType = authTyp }
        opMaps.Put( opDef.Name, reqDef )
    }
}

func ( m *OperationMap ) MustAddServiceInstance(
    ns *mg.Namespace, svc *mg.Identifier, typ *mg.QualifiedTypeName ) {

    if _, miss := m.instMap.GetOk( ns, svc ); miss == nil {
        panic( libErrorf( "map already has instnace: %s", 
            FormatInstanceId( ns, svc ) ) )
    } else {
        reqDefs := mg.NewIdentifierMap()
        m.instMap.Put( ns, svc, reqDefs )
        def := types.MustGetDefinition( typ, m.defs )
        sd := def.( *types.ServiceDefinition )
        m.mustAddRequestDefinitions( reqDefs, sd )
    }
}

func ( m *OperationMap ) ExpectOperationForRequest(
    ctx *RequestContext, path objpath.PathNode ) ( *RequestDefinition, error ) {

    def, err := m.instMap.getRequestValue( ctx, path )
    if err == nil { return def.( *RequestDefinition ), nil }
    return nil, err
}

type typedReqIface struct {
    iface RequestReactorInterface
    m *OperationMap
    reqDef *RequestDefinition
}

func ( i *typedReqIface ) StartRequest( 
    ctx *RequestContext, path objpath.PathNode ) ( err error ) {

    if i.reqDef, err = i.m.ExpectOperationForRequest( ctx, path ); err != nil {
        return err
    }
    return i.iface.StartRequest( ctx, path )
}

func ( i *typedReqIface ) newCastReactor( 
    typ mg.TypeReference ) *types.CastReactor {
    
    res := types.NewCastReactor( typ, i.m.defs )
    res.SkipPathSetter = true
    return res
}

func ( i *typedReqIface ) StartAuthentication( 
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    authTyp := i.reqDef.AuthenticationType 
    if authTyp == nil { 
        return nil, NewRequestError( path, errMsgNoAuthExpected )
    }
    rct, err := i.iface.StartAuthentication( path )
    if err != nil { return nil, err }
    cr := i.newCastReactor( authTyp )
    return mgRct.InitReactorPipeline( cr, rct ), nil
}

type typedReqParamsFactory struct {
    dt *mgRct.DepthTracker
    iface *typedReqIface
}

func ( f typedReqParamsFactory ) GetFieldSet(
    path objpath.PathNode ) ( *types.FieldSet, error ) {

    if f.dt.Depth() == 1 { 
        return f.iface.reqDef.Operation.Signature.Fields, nil
    }
    return nil, nil
}

func ( i *typedReqIface ) StartParameters( 
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    rct, err := i.iface.StartParameters( path )
    if err != nil { return nil, err }
    fsf := typedReqParamsFactory{ dt: mgRct.NewDepthTracker(), iface: i }
    cr := i.newCastReactor( mg.TypeSymbolMap )
    cr.FieldSetFactory = fsf
    return mgRct.InitReactorPipeline( fsf.dt, cr, rct ), nil
}

func AsTypedRequestReactorInterface( 
    iface RequestReactorInterface, m *OperationMap ) RequestReactorInterface {

    return &typedReqIface{ iface: iface, m: m }
}
