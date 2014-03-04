package types

import (
    mg "mingle"
    "bitgirder/objpath"
//    "log"
    "bitgirder/stack"
    "bitgirder/pipeline"
)

func notAStructError( p objpath.PathNode, qn *mg.QualifiedTypeName ) error {
    return mg.NewValueCastErrorf( p, "not a struct type: %s", qn )
}

func newUnrecognizedTypeError(
    p objpath.PathNode, qn *mg.QualifiedTypeName ) error {

    return mg.NewValueCastErrorf( p, "unrecognized type: %s", qn )
}

//// A synthetic type which we use along with a mingle.CastReactor to cast request
//// parameter maps to conform to their respective OperationDefinitions
//var typeTypedParameterMap = 
//    mg.MustQualifiedTypeName( "mingle:types@v1/TypedParameterMap" )

type defMapCastIface struct { dm *DefinitionMap }

func ( ci defMapCastIface ) InferStructFor( qn *mg.QualifiedTypeName ) bool {
    if def, ok := ci.dm.GetOk( qn ); ok {
        if _, ok = def.( *StructDefinition ); ok { return true }
    }
    return false
}

func collectFieldSets( sd *StructDefinition, dm *DefinitionMap ) []*FieldSet {
    flds := make( []*FieldSet, 0, 2 )
    for {
        flds = append( flds, sd.Fields )
        spr := sd.GetSuperType()
        if spr == nil { break }
        if def, ok := dm.GetOk( spr ); ok {
            if sd, ok = def.( *StructDefinition ); ! ok {
                tmpl := "super type %s of %s is not a struct"
                panic( libErrorf( tmpl, spr, sd.GetName() ) )
            }
        } else {
            tmpl := "can't find super type %s of %s"
            panic( libErrorf( tmpl, spr, sd.GetName() ) )
        }
    }
    return flds
}

//func typeNameIn( fd *FieldDefinition ) *mg.QualifiedTypeName {
//    nm := mg.TypeNameIn( fd.Type )
//    if qn, ok := nm.( *mg.QualifiedTypeName ); ok { return qn }
//    panic( libErrorf( 
//        "Name in type %s is not a qname: %s (%T)", fd.Type, nm, nm ) )
//}
//
//func expectDef( dm *DefinitionMap, qn *mg.QualifiedTypeName ) Definition {
//    if def, ok := dm.GetOk( qn ); ok { return def }
//    panic( libErrorf( "map has no definition for type %s", qn ) )
//}
//
//func expectAuthTypeOf( 
//    secQn *mg.QualifiedTypeName, dm *DefinitionMap ) mg.TypeReference {
//    if def, ok := dm.GetOk( secQn ); ok {
//        if protDef, ok := def.( *PrototypeDefinition ); ok {
//            flds := protDef.Signature.GetFields()
//            if fd := flds.Get( mg.IdAuthentication ); fd != nil {
//                return fd.Type
//            }
//            panic( libErrorf( "No auth for security: %s", secQn ) )
//        }
//        panic( libErrorf( "Not a prototype: %s", secQn ) )
//    }
//    panic( libErrorf( "No such security def: %s", secQn ) )
//}
//
//func valDefOf( fd *FieldDefinition, dm *DefinitionMap ) Definition {
//    qn := typeNameIn( fd )
//    return expectDef( dm, qn )
//}

type fieldTyper struct { 
    flds []*FieldSet 
    dm *DefinitionMap
}

func ( ft fieldTyper ) FieldTypeFor(
    fld *mg.Identifier, path objpath.PathNode ) ( mg.TypeReference, error ) {

    for _, flds := range ft.flds {
        if fd := flds.Get( fld ); fd != nil { return fd.Type, nil }
    }
    // use parent path since we're positioned on the failed field itself
    par := objpath.ParentOf( path )
    return nil, mg.NewUnrecognizedFieldError( par, fld )
}

func ( ci defMapCastIface ) FieldTyperFor(
    qn *mg.QualifiedTypeName, path objpath.PathNode ) ( mg.FieldTyper, error ) {

    flds := make( []*FieldSet, 0, 2 )
    for nm := qn; nm != nil; {
        if def, ok := ci.dm.GetOk( nm ); ok {
            if sd, ok := def.( *StructDefinition ); ok {
                flds = append( flds, sd.Fields )
                nm = sd.GetSuperType()
                continue
            } else { return nil, notAStructError( path, nm ) } 
        }
        nm = nil
    }
    if len( flds ) > 0 { return fieldTyper{ flds: flds, dm: ci.dm }, nil }
    tmpl := "no field type info for type %s"
    return nil, mg.NewValueCastErrorf( path, tmpl, qn )
}

func completeCastEnum(
    id *mg.Identifier, 
    ed *EnumDefinition, 
    path objpath.PathNode ) ( *mg.Enum, error ) {

    if res := ed.GetValue( id ); res != nil { return res, nil }
    tmpl := "illegal value for enum %s: %s"
    return nil, mg.NewValueCastErrorf( path, tmpl, ed.GetName(), id )
}

func castEnumFromString( 
    s string, ed *EnumDefinition, path objpath.PathNode ) ( *mg.Enum, error ) {

    id, err := mg.ParseIdentifier( s )
    if err != nil {
        tmpl := "invalid enum value %q: %s"
        return nil, mg.NewValueCastErrorf( path, tmpl, s, err )
    }
    return completeCastEnum( id, ed, path )
}

func castEnum( 
    val mg.Value, 
    ed *EnumDefinition, 
    path objpath.PathNode ) ( *mg.Enum, error ) {

    switch v := val.( type ) {
    case mg.String: return castEnumFromString( string( v ), ed, path )
    case *mg.Enum: 
        if v.Type.Equals( ed.GetName() ) {
            return completeCastEnum( v.Value, ed, path )
        }
    }
    t := ed.GetName().AsAtomicType()
    return nil, mg.NewTypeCastErrorValue( t, val, path )
}

func ( ci defMapCastIface ) CastAtomic(
    v mg.Value,
    at *mg.AtomicTypeReference,
    path objpath.PathNode ) ( mg.Value, error, bool ) {

    if qn, ok := at.Name.( *mg.QualifiedTypeName ); ok {
        if def, ok := ci.dm.GetOk( qn ); ok {
            if ed, ok := def.( *EnumDefinition ); ok {
                res, err := castEnum( v, ed, path )
                return res, err, true
            }
        }
    }
    return nil, nil, false
}

type fieldSetGetter interface {

    getFieldSets( 
        qn *mg.QualifiedTypeName, path objpath.PathNode ) ( []*FieldSet, error )
}

type defMapFieldSetGetter struct { dm *DefinitionMap }

func ( fsg defMapFieldSetGetter ) getFieldSets(
    qn *mg.QualifiedTypeName, path objpath.PathNode ) ( []*FieldSet, error ) {

    if def, ok := fsg.dm.GetOk( qn ); ok {
        if sd, ok := def.( *StructDefinition ); ok {
            return collectFieldSets( sd, fsg.dm ), nil
        } 
        return nil, notAStructError( path, qn )
    } 
    return nil, newUnrecognizedTypeError( path, qn )
}

type castReactor struct {
    typ mg.TypeReference
    iface mg.CastInterface
    dm *DefinitionMap
    fsg fieldSetGetter
    stack *stack.Stack
//    deflFeed *mg.EventPathReactor
}

func ( cr *castReactor ) InitializePipeline( pip *pipeline.Pipeline ) {
    pip.Add( mg.NewCastReactor( cr.typ, cr.iface ) )
}

//
//func ( cr *castReactor ) Init( rpi *mg.ReactorPipelineInit ) {
//    rpi.AddPipelineProcessor( cr.castBase )
//}
//
//func ( cr *castReactor ) GetPath() objpath.PathNode {
//    res := cr.castBase.GetPath()
//    if cr.deflFeed != nil { res = cr.deflFeed.AppendPath( res ) }
//    return res
//}
//
//func ( cr *castReactor ) newValueCastErrorf( 
//    tmpl string, args ...interface{} ) error {
//    return mg.NewValueCastErrorf( cr.GetPath(), tmpl, args... )
//}
//
//func ( cr *castReactor ) newUnrecognizedTypeError( 
//    qn *mg.QualifiedTypeName ) error {
//    return newUnrecognizedTypeError( qn, cr.GetPath() )
//}
//
//func ( cr *castReactor ) notAStructError( qn *mg.QualifiedTypeName ) error {
//    return notAStructError( qn, cr.GetPath() )
//}

type fieldCtx struct {
    depth int
    await *mg.IdentifierMap
}

func ( fc *fieldCtx ) removeOptFields() {
    done := make( []*mg.Identifier, 0, fc.await.Len() )
    fc.await.EachPair( func( _ *mg.Identifier, val interface{} ) {
        fd := val.( *FieldDefinition )
        if _, ok := fd.Type.( *mg.NullableTypeReference ); ok {
            done = append( done, fd.Name )
        }
    })
    for _, fld := range done { fc.await.Delete( fld ) }
}

func ( cr *castReactor ) newFieldCtx( flds []*FieldSet ) *fieldCtx {
    res := &fieldCtx{ await: mg.NewIdentifierMap() }
    for _, fs := range flds {
        fs.EachDefinition( func( fd *FieldDefinition ) {
            res.await.Put( fd.Name, fd )
        })
    }
    return res
}

func ( cr *castReactor ) startStruct( ss *mg.StructStartEvent ) error {
    flds, err := cr.fsg.getFieldSets( ss.Type, ss.GetPath() )
    if err != nil { return err }
    if flds != nil { cr.stack.Push( cr.newFieldCtx( flds ) ) }
    return nil
}

// We don't re-check here that fld is actually part of the defined field set or
// since the upstream defMapCastIface will have validated that already
func ( cr *castReactor ) startField( fs *mg.FieldStartEvent ) error {
    if cr.stack.IsEmpty() { return nil }
    cr.stack.Peek().( *fieldCtx ).await.Delete( fs.Field )
    return nil
}

func feedDefault( 
    fld *mg.Identifier, 
    defl mg.Value, 
    p objpath.PathNode,
    next mg.ReactorEventProcessor ) error {

    ps := mg.NewPathSettingProcessorPath( p.Descend( fld ) )
    ps.SkipStructureCheck = true
    pip := mg.InitReactorPipeline( ps, next )
    return mg.VisitValue( defl, pip )
}

func processDefaults(
    fldCtx *fieldCtx, 
    p objpath.PathNode, 
    next mg.ReactorEventProcessor ) error {

    vis := func( fld *mg.Identifier, val interface{} ) error {
        fd := val.( *FieldDefinition )
        if defl := fd.GetDefault(); defl != nil { 
            if err := feedDefault( fld, defl, p, next ); err != nil { 
                return err 
            }
            fldCtx.await.Delete( fld )
        }
        return nil
    }
    return fldCtx.await.EachPairError( vis )
}

func createMissingFieldsError( p objpath.PathNode, fldCtx *fieldCtx ) error {
    flds := make( []*mg.Identifier, 0, fldCtx.await.Len() )
    fldCtx.await.EachPair( func( fld *mg.Identifier, _ interface{} ) {
        flds = append( flds, fld )
    })
    return mg.NewMissingFieldsError( p, flds )
}

func ( cr *castReactor ) end( 
    ev *mg.EndEvent, next mg.ReactorEventProcessor ) error {
    if cr.stack.IsEmpty() { return nil }
    fldCtx := cr.stack.Peek().( *fieldCtx )
    if fldCtx.depth > 0 {
        fldCtx.depth--
        return nil
    }
    cr.stack.Pop()
    p := ev.GetPath()
    if err := processDefaults( fldCtx, p, next ); err != nil { return err }
    fldCtx.removeOptFields()
    if fldCtx.await.Len() > 0 { return createMissingFieldsError( p, fldCtx ) }
    return nil
}

func ( cr *castReactor ) startContainer() error {
    if ! cr.stack.IsEmpty() { cr.stack.Peek().( *fieldCtx ).depth++ }
    return nil
}

// we only do value checks here that are specific to this cast, namely having to
// do with enums. If the value is an enum, we check that we recogzize the type
// and that the type is actually an enum. We don't actually check the enum value
// here though, and leave that for CastAtomic. Any other values aren't checked
// here and are left to CastAtomic or to the upstream processor.
func ( cr *castReactor ) valueEvent( ve *mg.ValueEvent ) error {
    if en, ok := ve.Val.( *mg.Enum ); ok {
        if def, ok := cr.dm.GetOk( en.Type ); ok {
            if _, ok := def.( *EnumDefinition ); ok { return nil }
            tmpl := "not an enum type: %s"
            return mg.NewValueCastErrorf( ve.GetPath(), tmpl, en.Type )
        } 
        return newUnrecognizedTypeError( ve.GetPath(), en.Type )
    }
    return nil
}

func ( cr *castReactor ) prepareProcessEvent(
    ev mg.ReactorEvent, next mg.ReactorEventProcessor ) error {
    
    switch v := ev.( type ) {
    case *mg.StructStartEvent: return cr.startStruct( v )
    case *mg.FieldStartEvent: return cr.startField( v )
    case *mg.ValueEvent: return cr.valueEvent( v )
    case *mg.EndEvent: return cr.end( v, next )
    case *mg.ListStartEvent, *mg.MapStartEvent: return cr.startContainer()
    }
    panic( libErrorf( "unhandled event: %T", ev ) )
}

func ( cr *castReactor ) ProcessEvent( 
    ev mg.ReactorEvent, next mg.ReactorEventProcessor ) error {
    if err := cr.prepareProcessEvent( ev, next ); err != nil { return err }
    return next.ProcessEvent( ev )
}

func newCastReactorBase(
    typ mg.TypeReference, 
    iface mg.CastInterface,
    dm *DefinitionMap, 
    fsg fieldSetGetter ) mg.PipelineProcessor {

    return &castReactor{ 
        typ: typ,
        iface: iface,
        dm: dm,
        fsg: fsg,
        stack: stack.NewStack(),
    }
}

func NewCastReactorDefinitionMap(
    typ mg.TypeReference, dm *DefinitionMap ) mg.PipelineProcessor {
 
    fsg := defMapFieldSetGetter{ dm }
    return newCastReactorBase( typ, defMapCastIface{ dm }, dm, fsg )
}

//type opMatcher struct {
//
//    svcDefs *ServiceDefinitionMap
//
//    ns *mg.Namespace
//    sd *ServiceDefinition
//    opDef *OperationDefinition
//}
//
//func ( om *opMatcher ) defMap() *DefinitionMap {
//    return om.svcDefs.GetDefinitionMap()
//}
//
//func ( om *opMatcher ) Namespace( ns *mg.Namespace, pg mg.PathGetter ) error {
//    if ! om.svcDefs.HasNamespace( ns ) {
//        return mg.NewEndpointErrorNamespace( ns, pg.GetPath() )
//    }
//    om.ns = ns
//    return nil
//}
//
//func ( om *opMatcher ) Service( svc *mg.Identifier, pg mg.PathGetter ) error {
//    if sd, ok := om.svcDefs.GetOk( om.ns, svc ); ok {
//        om.sd = sd
//        return nil
//    }
//    return mg.NewEndpointErrorService( svc, pg.GetPath() )
//}
//
//func ( om *opMatcher ) Operation( op *mg.Identifier, pg mg.PathGetter ) error {
//    if om.opDef = om.sd.findOperation( op ); om.opDef == nil {
//        return mg.NewEndpointErrorOperation( op, pg.GetPath() )
//    }
//    return nil
//}
//
//type OpMatch interface {
//}
//
//type RequestReactorInterface interface {
//
//    GetAuthenticationProcessor( 
//        om OpMatch, pg mg.PathGetter ) ( mg.ReactorEventProcessor, error )
//
//    GetParametersProcessor(
//        om OpMatch, pg mg.PathGetter ) ( mg.ReactorEventProcessor, error )
//}
//
//type mgReqImpl struct {
//
//    *opMatcher
//    iface RequestReactorInterface
//
//    sawAuth bool
//}
//
//func ( impl *mgReqImpl ) needsAuth() bool { return impl.sd.Security != nil }
//
//func ( impl *mgReqImpl ) GetAuthenticationProcessor( 
//    pg mg.PathGetter ) ( mg.ReactorEventProcessor, error ) {
//    impl.sawAuth = true
//    pg = mg.ImmediatePathGetter{ pg.GetPath() }
//    if secQn := impl.sd.Security; secQn != nil { 
//        authTyp := expectAuthTypeOf( secQn, impl.defMap() )
//        dm := impl.defMap()
//        fsg := defMapFieldSetGetter{ dm }
//        cr := newCastReactor( authTyp, defMapCastIface{ dm }, dm, fsg, pg )
//        authRct, err := 
//            impl.iface.GetAuthenticationProcessor( impl.opMatcher, pg )
//        if err != nil { return nil, err }
//        return mg.InitReactorPipeline( cr, authRct ), nil
//    }
//    return mg.DiscardProcessor, nil
//}
//
//type parametersCastIface struct {
//    ci defMapCastIface
//    opDef *OperationDefinition
//    defs *DefinitionMap
//}
//
//func ( pi parametersCastIface ) InferStructFor( 
//    qn *mg.QualifiedTypeName ) bool {
//    if qn.Equals( typeTypedParameterMap ) { return true }
//    return pi.ci.InferStructFor( qn )
//}
//
//func ( pi parametersCastIface ) fieldSets() []*FieldSet {
//    return []*FieldSet{ pi.opDef.Signature.Fields }
//}
//
//func ( pi parametersCastIface ) FieldTyperFor(
//    qn *mg.QualifiedTypeName, pg mg.PathGetter ) ( mg.FieldTyper, error ) {
//    if qn.Equals( typeTypedParameterMap ) { 
//        return fieldTyper{ flds: pi.fieldSets(), dm: pi.defs }, nil
//    }
//    return pi.ci.FieldTyperFor( qn, pg )
//}
//
//func ( pi parametersCastIface ) CastAtomic(
//    in mg.Value, 
//    at *mg.AtomicTypeReference, 
//    pg mg.PathGetter ) ( mg.Value, error, bool ) {
//    return pi.ci.CastAtomic( in, at, pg )
//}
//
//func ( pi parametersCastIface ) getFieldSets(
//    qn *mg.QualifiedTypeName, pg mg.PathGetter ) ( []*FieldSet, error ) {
//    if qn.Equals( typeTypedParameterMap ) { return pi.fieldSets(), nil }
//    return ( defMapFieldSetGetter{ pi.defs } ).getFieldSets( qn, pg )
//}
//
//type parametersReactor struct { rep mg.ReactorEventProcessor }
//
//func ( pr parametersReactor ) ProcessEvent( ev mg.ReactorEvent ) error {
//    if ss, ok := ev.( mg.StructStartEvent ); ok {
//        if ss.Type.Equals( typeTypedParameterMap ) { ev = mg.NewMapStartEvent() }
//    }
//    return pr.rep.ProcessEvent( ev )
//}
//
//func ( impl *mgReqImpl ) checkGotAuth( pg mg.PathGetter ) error {
//    if impl.needsAuth() {
//        if ! impl.sawAuth {
//            // take parent since pg itself will be at 'parameters'
//            path := objpath.ParentOf( pg.GetPath() )
//            flds := []*mg.Identifier{ mg.IdAuthentication }
//            return mg.NewMissingFieldsError( path, flds )
//        }
//    }
//    return nil
//}
//
//func ( impl *mgReqImpl ) GetParametersProcessor(
//    pg mg.PathGetter ) ( mg.ReactorEventProcessor, error ) {
//    if err := impl.checkGotAuth( pg ); err != nil { return nil, err }
//    pg = mg.ImmediatePathGetter{ pg.GetPath() }
//    dm := impl.defMap()
//    pci := parametersCastIface{ 
//        ci: defMapCastIface{ dm },
//        defs: dm,
//        opDef: impl.opDef,
//    }
//    typ := typeTypedParameterMap.AsAtomicType()
//    cr := newCastReactor( typ, pci, dm, pci, pg )
//    rep, err := impl.iface.GetParametersProcessor( impl.opMatcher, pg )
//    if err != nil { return nil, err }
//    paramsRct := parametersReactor{ rep }
//    return mg.InitReactorPipeline( 
//        cr, 
////        mg.NewDebugReactor( mg.DebugLoggerFunc( log.Printf ) ),
//        paramsRct,
//    ), nil
//}
//
//func NewRequestReactor( 
//    svcDefs *ServiceDefinitionMap, 
//    iface RequestReactorInterface ) *mg.ServiceRequestReactor {
//    return mg.NewServiceRequestReactor(
//        &mgReqImpl{ 
//            opMatcher: &opMatcher{ svcDefs: svcDefs },
//            iface: iface,
//        },
//    )
//}
//
//type errorProcReactor struct {
//    defs *DefinitionMap
//    throws []mg.TypeReference
//    proc mg.ReactorEventProcessor
//    pg mg.PathGetter
//}
//
//func errorForUnexpectedErrorType( 
//    path objpath.PathNode, typ mg.TypeReference ) error {
//    return mg.NewValueCastErrorf( path,
//        "Error type is not a declared thrown type: %s", typ )
//}
//
//func ( epr *errorProcReactor ) errorForUnexpectedErrorType( 
//    typ mg.TypeReference ) error {
//    return errorForUnexpectedErrorType( epr.pg.GetPath(), typ )
//}
//
//func ( epr *errorProcReactor ) errorForUnexpectedErrorQname(
//    qn *mg.QualifiedTypeName ) error {
//    typ := &mg.AtomicTypeReference{ Name: qn }
//    return epr.errorForUnexpectedErrorType( typ )
//}
//
//func ( epr *errorProcReactor ) errorForUnexpectedErrorValue( 
//    ev mg.ReactorEvent ) error {
//    var typ mg.TypeReference
//    switch v := ev.( type ) {
//    case mg.ValueEvent: typ = mg.TypeOf( v.Val )
//    case mg.ListStartEvent: typ = mg.TypeOpaqueList
//    case mg.MapStartEvent: typ = mg.TypeSymbolMap
//    }
//    if ( typ == nil ) { panic( libErrorf( "unhandled event type: %T", ev ) ) }
//    return epr.errorForUnexpectedErrorType( typ )
//}
//
//func ( epr *errorProcReactor ) errorTypeForStruct( 
//    qn *mg.QualifiedTypeName ) ( mg.TypeReference, bool ) {
//    if epr.defs.HasBuiltInDefinition( qn ) {
//        return &mg.AtomicTypeReference{ Name: qn }, true
//    }
//    for _, typ := range epr.throws {
//        if mg.TypeNameIn( typ ).Equals( qn ) { return typ, true }
//    }
//    return nil, false
//}
//
//func ( epr *errorProcReactor ) ProcessEvent( ev mg.ReactorEvent ) error {
//    if ( epr.proc == nil ) { 
//        if ss, ok := ev.( mg.StructStartEvent ); ok {
//            if typ, ok := epr.errorTypeForStruct( ss.Type ); ok {
//                cr := newCastReactor( typ, epr.defs, epr.pg )
//                epr.proc = mg.InitReactorPipeline( cr )
//            } else { return epr.errorForUnexpectedErrorQname( ss.Type ) }
//        } else { return epr.errorForUnexpectedErrorValue( ev ) }
//    }
//    return epr.proc.ProcessEvent( ev )
//}
//
//type ResponseReactorInterface interface {
//    GetResultProcessor( pg mg.PathGetter ) ( mg.ReactorEventProcessor, error )
//    GetErrorProcessor( pg mg.PathGetter ) ( mg.ReactorEventProcessor, error )
//}
//
//type mgRespImpl struct {
//    defs *DefinitionMap
//    opDef *OperationDefinition
//    iface ResponseReactorInterface
//}
//
//func ( impl *mgRespImpl ) GetResultProcessor( 
//    pg mg.PathGetter ) ( mg.ReactorEventProcessor, error ) {
//    resTyp := impl.opDef.Signature.Return
//    ipg := mg.ImmediatePathGetter{ pg.GetPath() }
//    cr := newCastReactor( resTyp, impl.defs, ipg )
//    rct, err := impl.iface.GetResultProcessor( pg )
//    if err != nil { return nil, err }
//    return mg.InitReactorPipeline( cr, rct ), nil
//}
//
//func ( impl *mgRespImpl ) GetErrorProcessor( 
//    pg mg.PathGetter ) ( mg.ReactorEventProcessor, error ) {
//    rct, err := impl.iface.GetErrorProcessor( pg )
//    if err != nil { return nil, err }
//    epr := &errorProcReactor{ 
//        defs: impl.defs, 
//        throws: impl.opDef.Signature.Throws,
//        pg: mg.ImmediatePathGetter{ pg.GetPath() },
//    }
//    return mg.InitReactorPipeline( epr, rct ), nil
//}
//
//func NewResponseReactor(
//    defs *DefinitionMap,
//    opDef *OperationDefinition,
//    iface ResponseReactorInterface ) *mg.ServiceResponseReactor {
//    resTyp := opDef.Signature.Return
//    qn := mg.TypeNameIn( resTyp ).( *mg.QualifiedTypeName )
//    if ! defs.HasKey( qn ) {
//        tmpl := "operation '%s' declares unrecognized return type: %s"
//        panic( libErrorf( tmpl, opDef.Name, resTyp ) )
//    }
//    mgIface := &mgRespImpl{ iface: iface, defs: defs, opDef: opDef }
//    return mg.NewServiceResponseReactor( mgIface )
//}
