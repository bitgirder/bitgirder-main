package service

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "mingle/types"
    "bitgirder/objpath"
    "bitgirder/stub"
)

type RequestDefinition struct {
    Operation *types.OperationDefinition
    AuthenticationType mg.TypeReference
}

type OperationMap struct {
    nsMap *mg.NamespaceMap
    defs *types.DefinitionMap
}

func NewOperationMap( defs *types.DefinitionMap ) *OperationMap {
    return &OperationMap{ defs: defs, nsMap: mg.NewNamespaceMap() }
}

func ( m *OperationMap ) mustAddRequestDefinitions(
    opMaps *mg.IdentifierMap, sd *types.ServiceDefinition ) {

    var authTyp mg.TypeReference
    if secQn := sd.Security; secQn != nil { 
        authTyp = types.MustAuthTypeOf( secQn, m.defs )
    }
    for _, opDef := range sd.Operations {
        reqDef := &RequestDefinition{ Operation: opDef }
        if authTyp != nil { reqDef.AuthenticationType = authTyp }
        opMaps.Put( opDef.Name, reqDef )
    }
}

func ( m *OperationMap ) MustAddServiceInstance(
    ns *mg.Namespace, svc *mg.Identifier, typ *mg.QualifiedTypeName ) {

    sd := m.defs.MustGet( typ ).( *types.ServiceDefinition )
    var svcMaps *mg.IdentifierMap
    if v, ok := m.nsMap.GetOk( ns ); ok {
        svcMaps = v.( *mg.IdentifierMap )
    } else {
        svcMaps = mg.NewIdentifierMap()
        m.nsMap.Put( ns, svcMaps )
    }
    if _, ok := svcMaps.GetOk( svc ); ok {
        panic( libErrorf( "%s.%s already has an instance", ns, svc ) )
    } else {
        opMaps := mg.NewIdentifierMap()
        m.mustAddRequestDefinitions( opMaps, sd )
        svcMaps.Put( svc, opMaps )
    }
}

func ( m *OperationMap ) errUnknownEndpoint(
    ctx *RequestContext, errFld *mg.Identifier, path objpath.PathNode ) error {

    var tmpl string
    args := make( []interface{}, 0, 2 )
    switch {
    case errFld.Equals( IdNamespace ):
        tmpl = "endpoint has no such namespace: %s"
        args = append( args, ctx.Namespace )
    case errFld.Equals( IdService ):
        tmpl = "namespace %s has no such service: %s"
        args = append( args, ctx.Namespace, ctx.Service )
    case errFld.Equals( IdOperation ):
        tmpl = "service %s.%s has no such operation: %s"
        args = append( args, ctx.Namespace, ctx.Service, ctx.Operation )
    default: panic( libErrorf( "unhandled errFld: %s", errFld ) )
    }
    return NewRequestErrorf( path, tmpl, args... )
}

func ( m *OperationMap ) ExpectOperationForRequest(
    ctx *RequestContext, 
    path objpath.PathNode ) ( *types.OperationDefinition, error ) {

    if svcMapVal, ok := m.nsMap.GetOk( ctx.Namespace ); ok {
        svcMap := svcMapVal.( *mg.IdentifierMap )
        if opMapVal, ok := svcMap.GetOk( ctx.Service ); ok {
            opMap := opMapVal.( *mg.IdentifierMap )
            if opDefVal, ok := opMap.GetOk( ctx.Operation ); ok {
                return opDefVal.( *types.OperationDefinition ), nil
            }
            return nil, m.errUnknownEndpoint( ctx, IdOperation, path )
        }
        return nil, m.errUnknownEndpoint( ctx, IdService, path )
    }
    return nil, m.errUnknownEndpoint( ctx, IdNamespace, path )
}

type typedReqIface struct {
    iface RequestReactorInterface
    m *OperationMap
    od *types.OperationDefinition
}

func ( i *typedReqIface ) StartRequest( 
    ctx *RequestContext, path objpath.PathNode ) ( err error ) {

    i.od, err = i.m.ExpectOperationForRequest( ctx, path )
    return
}

func ( i *typedReqIface ) StartAuthentication( 
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    return nil, stub.Unimplemented()
}

func ( i *typedReqIface ) StartParameters( 
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    return nil, stub.Unimplemented()
}

func AsTypedRequestReactorInterface( 
    iface RequestReactorInterface, m *OperationMap ) RequestReactorInterface {

    return &typedReqIface{ iface: iface, m: m }
}
