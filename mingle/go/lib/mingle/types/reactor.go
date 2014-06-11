package types

import (
    mg "mingle"
    "mingle/parser"
    mgRct "mingle/reactor"
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

// A synthetic type which we use along with a mingle.CastReactor to cast request
// parameter maps to conform to their respective OperationDefinitions
var qnTypedParameterMap *mg.QualifiedTypeName
var typTypedParameterMap *mg.AtomicTypeReference

func init() {
    qnTypedParameterMap = &mg.QualifiedTypeName{
        Namespace: &mg.Namespace{
            Parts: []*mg.Identifier{
                mg.NewIdentifierUnsafe( []string{ "mingle" } ),
                mg.NewIdentifierUnsafe( []string{ "types" } ),
            },
            Version: mg.NewIdentifierUnsafe( []string{ "v1" } ),
        },
        Name: mg.NewDeclaredTypeNameUnsafe( "TypedParameterMap" ),
    }
    typTypedParameterMap = qnTypedParameterMap.AsAtomicType()
}

type defMapCastIface struct { dm *DefinitionMap }

func ( ci defMapCastIface ) InferStructFor( qn *mg.QualifiedTypeName ) bool {
    if def, ok := ci.dm.GetOk( qn ); ok {
        if _, ok = def.( *StructDefinition ); ok { return true }
    }
    return false
}

func ( ci defMapCastIface ) AllowAssignment(
    expct, act *mg.QualifiedTypeName ) bool {

    if _, ok := ci.dm.GetOk( act ); ! ok { return false }
    return canAssignType( expct, act, ci.dm )
}

type fieldTyper struct { 
    flds []*FieldSet 
    dm *DefinitionMap
}

func ( ft fieldTyper ) FieldTypeFor(
    fld *mg.Identifier, path objpath.PathNode ) ( mg.TypeReference, error ) {

    for _, flds := range ft.flds {
        if fd := flds.Get( fld ); fd != nil { return fd.Type, nil }
    }
    return nil, mg.NewUnrecognizedFieldError( path, fld )
}

func ( ci defMapCastIface ) FieldTyperFor(
    qn *mg.QualifiedTypeName, 
    path objpath.PathNode ) ( mgRct.FieldTyper, error ) {

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

    id, err := parser.ParseIdentifier( s )
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

    if def, ok := ci.dm.GetOk( at.Name ); ok {
        if ed, ok := def.( *EnumDefinition ); ok {
            res, err := castEnum( v, ed, path )
            return res, err, true
        }
    }
    return nil, nil, false
}

type fieldSetGetter interface {

    getFieldSets( 
        qn *mg.QualifiedTypeName, path objpath.PathNode ) ( []*FieldSet, error )
}

func fieldSetsForTypeInDefMap(
    qn *mg.QualifiedTypeName, 
    dm *DefinitionMap, 
    path objpath.PathNode ) ( []*FieldSet, error ) {

    if def, ok := dm.GetOk( qn ); ok {
        if sd, ok := def.( *StructDefinition ); ok {
            return collectFieldSets( sd, dm ), nil
        } 
        return nil, notAStructError( path, qn )
    } 
    return nil, newUnrecognizedTypeError( path, qn )
}

type defMapFieldSetGetter struct { dm *DefinitionMap }

func ( fsg defMapFieldSetGetter ) getFieldSets(
    qn *mg.QualifiedTypeName, path objpath.PathNode ) ( []*FieldSet, error ) {

    return fieldSetsForTypeInDefMap( qn, fsg.dm, path )
}

type castReactor struct {
    typ mg.TypeReference
    iface mgRct.CastInterface
    dm *DefinitionMap
    fsg fieldSetGetter
    stack *stack.Stack
    skipPathSetter bool
}

func ( cr *castReactor ) InitializePipeline( pip *pipeline.Pipeline ) {
    mgCastRct := mgRct.NewCastReactor( cr.typ, cr.iface )
    mgCastRct.SkipPathSetter = cr.skipPathSetter
    pip.Add( mgCastRct )
}

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

func ( cr *castReactor ) startStruct( ss *mgRct.StructStartEvent ) error {
    flds, err := cr.fsg.getFieldSets( ss.Type, ss.GetPath() )
    if err != nil { return err }
    if flds != nil { cr.stack.Push( cr.newFieldCtx( flds ) ) }
    return nil
}

// We don't re-check here that fld is actually part of the defined field set or
// since the upstream defMapCastIface will have validated that already
func ( cr *castReactor ) startField( fs *mgRct.FieldStartEvent ) error {
    if cr.stack.IsEmpty() { return nil }
    cr.stack.Peek().( *fieldCtx ).await.Delete( fs.Field )
    return nil
}

func feedDefault( 
    fld *mg.Identifier, 
    defl mg.Value, 
    p objpath.PathNode,
    next mgRct.ReactorEventProcessor ) error {

    fldPath := objpath.Descend( p, fld )
    fs := mgRct.NewFieldStartEvent( fld )
    fs.SetPath( fldPath )
    if err := next.ProcessEvent( fs ); err != nil { return err }
    ps := mgRct.NewPathSettingProcessorPath( fldPath )
    ps.SkipStructureCheck = true
    pip := mgRct.InitReactorPipeline( ps, next )
    return mgRct.VisitValue( defl, pip )
}

func processDefaults(
    fldCtx *fieldCtx, 
    p objpath.PathNode, 
    next mgRct.ReactorEventProcessor ) error {

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
    ev *mgRct.EndEvent, next mgRct.ReactorEventProcessor ) error {
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
func ( cr *castReactor ) valueEvent( ve *mgRct.ValueEvent ) error {
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
    ev mgRct.ReactorEvent, next mgRct.ReactorEventProcessor ) error {
    
    switch v := ev.( type ) {
    case *mgRct.StructStartEvent: return cr.startStruct( v )
    case *mgRct.FieldStartEvent: return cr.startField( v )
    case *mgRct.ValueEvent: return cr.valueEvent( v )
    case *mgRct.EndEvent: return cr.end( v, next )
    case *mgRct.ListStartEvent, *mgRct.MapStartEvent: return cr.startContainer()
    case *mgRct.ValueAllocationEvent, *mgRct.ValueReferenceEvent: return nil
    }
    panic( libErrorf( "unhandled event: %T", ev ) )
}

func ( cr *castReactor ) ProcessEvent( 
    ev mgRct.ReactorEvent, next mgRct.ReactorEventProcessor ) error {
    if err := cr.prepareProcessEvent( ev, next ); err != nil { return err }
    return next.ProcessEvent( ev )
}

func newCastReactorBase(
    typ mg.TypeReference, 
    iface mgRct.CastInterface,
    dm *DefinitionMap, 
    fsg fieldSetGetter ) *castReactor {

    return &castReactor{ 
        typ: typ,
        iface: iface,
        dm: dm,
        fsg: fsg,
        stack: stack.NewStack(),
    }
}

func newCastReactorDefinitionMap(
    typ mg.TypeReference, dm *DefinitionMap ) *castReactor {
 
    fsg := defMapFieldSetGetter{ dm }
    return newCastReactorBase( typ, defMapCastIface{ dm }, dm, fsg )
}

// the public version of newCastReactorDefinitionMap, typed to return something
// other than our internal *castReactor type; we could combine this with
// newCastReactorDefinitionMap if we end up making *castReactor public
func NewCastReactorDefinitionMap(
    typ mg.TypeReference, dm *DefinitionMap ) mgRct.PipelineProcessor {
    
    return newCastReactorDefinitionMap( typ, dm )
}

type RequestReactorInterface interface {

    GetAuthenticationReactor( 
        path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error )

    GetParametersReactor(
        path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error )
}

type mgReqImpl struct {

    svcDefs *ServiceDefinitionMap

    ns *mg.Namespace
    sd *ServiceDefinition
    opDef *OperationDefinition

    iface RequestReactorInterface

    sawAuth bool
}

func ( impl *mgReqImpl ) defMap() *DefinitionMap {
    return impl.svcDefs.GetDefinitionMap()
}

func ( impl *mgReqImpl ) Namespace( 
    ns *mg.Namespace, p objpath.PathNode ) error {

    if ! impl.svcDefs.HasNamespace( ns ) {
        return mg.NewEndpointErrorNamespace( ns, p )
    }
    impl.ns = ns
    return nil
}

func ( impl *mgReqImpl ) Service( 
    svc *mg.Identifier, p objpath.PathNode ) error {

    if sd, ok := impl.svcDefs.GetOk( impl.ns, svc ); ok {
        impl.sd = sd
        return nil
    }
    return mg.NewEndpointErrorService( svc, p )
}

func ( impl *mgReqImpl ) Operation( 
    op *mg.Identifier, p objpath.PathNode ) error {

    if impl.opDef = impl.sd.findOperation( op ); impl.opDef == nil {
        return mg.NewEndpointErrorOperation( op, p )
    }
    return nil
}

func ( impl *mgReqImpl ) needsAuth() bool { return impl.sd.Security != nil }

func ( impl *mgReqImpl ) GetAuthenticationReactor( 
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    impl.sawAuth = true
    if secQn := impl.sd.Security; secQn != nil { 
        authTyp := expectAuthTypeOf( secQn, impl.defMap() )
        cr := NewCastReactorDefinitionMap( authTyp, impl.defMap() )
        authRct, err := impl.iface.GetAuthenticationReactor( path )
        if err != nil { return nil, err }
        return mgRct.InitReactorPipeline( cr, authRct ), nil
    }
    return mgRct.DiscardProcessor, nil
}

type parametersCastIface struct {
    ci defMapCastIface
    opDef *OperationDefinition
    defs *DefinitionMap
}

func ( pi parametersCastIface ) InferStructFor( 
    qn *mg.QualifiedTypeName ) bool {

    if qn.Equals( qnTypedParameterMap ) { return true }
    return pi.ci.InferStructFor( qn )
}

func ( pi parametersCastIface ) fieldSets() []*FieldSet {
    return []*FieldSet{ pi.opDef.Signature.Fields }
}

func ( pi parametersCastIface ) FieldTyperFor(
    qn *mg.QualifiedTypeName, 
    path objpath.PathNode ) ( mgRct.FieldTyper, error ) {

    if qn.Equals( qnTypedParameterMap ) { 
        return fieldTyper{ flds: pi.fieldSets(), dm: pi.defs }, nil
    }
    return pi.ci.FieldTyperFor( qn, path )
}

func ( pi parametersCastIface ) CastAtomic(
    in mg.Value, 
    at *mg.AtomicTypeReference, 
    path objpath.PathNode ) ( mg.Value, error, bool ) {

    return pi.ci.CastAtomic( in, at, path )
}

func ( pi parametersCastIface ) AllowAssignment(
    expct, act *mg.QualifiedTypeName ) bool {

    return pi.ci.AllowAssignment( expct, act )
}

func ( pi parametersCastIface ) getFieldSets(
    qn *mg.QualifiedTypeName, path objpath.PathNode ) ( []*FieldSet, error ) {

    if qn.Equals( qnTypedParameterMap ) { return pi.fieldSets(), nil }
    return fieldSetsForTypeInDefMap( qn, pi.defs, path )
}

type parametersReactor struct { rep mgRct.ReactorEventProcessor }

func ( pr parametersReactor ) ProcessEvent( ev mgRct.ReactorEvent ) error {
    ev2 := ev
    if ss, ok := ev.( *mgRct.StructStartEvent ); ok {
        if ss.Type.Equals( qnTypedParameterMap ) { 
            ev2 = mgRct.NewMapStartEvent( mg.PointerIdNull ) 
            ev2.SetPath( ev.GetPath() )
        }
    }
    return pr.rep.ProcessEvent( ev2 )
}

// If we return an error we use the parent of path, since path will be
// positioned at the 'parameters' field, from which we call this func.
func ( impl *mgReqImpl ) checkGotAuth( path objpath.PathNode ) error {
    if impl.sawAuth || ( ! impl.needsAuth() ) { return nil }
    flds := []*mg.Identifier{ mg.IdAuthentication }
    par := objpath.ParentOf( path )
    return mg.NewMissingFieldsError( par, flds )
}

func ( impl *mgReqImpl ) GetParametersReactor(
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    if err := impl.checkGotAuth( path ); err != nil { return nil, err }
    typ := typTypedParameterMap
    dm := impl.defMap()
    ci := defMapCastIface{ impl.defMap() }
    pci := parametersCastIface{ ci: ci, defs: ci.dm, opDef: impl.opDef }
    cr := newCastReactorBase( typ, pci, dm, pci )
    cr.skipPathSetter = true
    proc, err := impl.iface.GetParametersReactor( path )
    if err != nil { return nil, err }
    paramsRct := parametersReactor{ proc }
    return mgRct.InitReactorPipeline( cr, paramsRct ), nil
}

func NewRequestReactor( 
    svcDefs *ServiceDefinitionMap, 
    iface RequestReactorInterface ) *mgRct.RequestReactor {

    reqImpl := &mgReqImpl{ svcDefs: svcDefs, iface: iface }
    return mgRct.NewRequestReactor( reqImpl )
}

func errorForUnexpectedErrorType( 
    path objpath.PathNode, typ mg.TypeReference ) error {

    return mg.NewValueCastErrorf( path,
        "error type is not a declared thrown type: %s", typ )
}

func errorForUnexpectedErrorValue( ev mgRct.ReactorEvent ) error {
    var typ mg.TypeReference
    switch v := ev.( type ) {
    case *mgRct.ValueEvent: typ = mg.TypeOf( v.Val )
    case *mgRct.ListStartEvent: typ = mg.TypeOpaqueList
    case *mgRct.MapStartEvent: typ = mg.TypeSymbolMap
    }
    if ( typ == nil ) { panic( libErrorf( "unhandled event type: %T", ev ) ) }
    return errorForUnexpectedErrorType( ev.GetPath(), typ )
}

type ResponseReactorInterface interface {

    GetResultReactor( 
        p objpath.PathNode ) ( mgRct.ReactorEventProcessor, error )

    GetErrorReactor( p objpath.PathNode ) ( mgRct.ReactorEventProcessor, error )
}

type mgRespImpl struct {
    defs *DefinitionMap
    svcDef *ServiceDefinition
    opDef *OperationDefinition
    iface ResponseReactorInterface
}

func ( impl *mgRespImpl ) GetResultReactor( 
    path objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    resTyp := impl.opDef.Signature.Return
    rct, err := impl.iface.GetResultReactor( path )
    if err != nil { return nil, err }
    cr := newCastReactorDefinitionMap( resTyp, impl.defs )
    cr.skipPathSetter = true
    return mgRct.InitReactorPipeline( cr, rct ), nil
}

type errorProcReactor struct {
    impl *mgRespImpl
    rct mgRct.ReactorEventProcessor
    proc mgRct.ReactorEventProcessor
}

func ( epr *errorProcReactor ) errorTypeForStruct( 
    qn *mg.QualifiedTypeName ) ( mg.TypeReference, bool ) {

    if epr.impl.defs.HasBuiltInDefinition( qn ) {
        return &mg.AtomicTypeReference{ Name: qn }, true
    }
    if _, ok := epr.impl.defs.GetOk( qn ); ! ok { return nil, false }
    opSig := epr.impl.opDef.Signature
    if typ, ok := canThrowErrorOfType( qn, opSig, epr.impl.defs ); ok {
        return typ, true
    }
    if secQn := epr.impl.svcDef.Security; secQn != nil {
        pd := expectProtoDef( secQn, epr.impl.defs )
        return canThrowErrorOfType( qn, pd.Signature, epr.impl.defs )
    }
    return nil, false
}

func ( epr *errorProcReactor ) initProc( ev mgRct.ReactorEvent ) error {
    if ss, ok := ev.( *mgRct.StructStartEvent ); ok {
        if typ, ok := epr.errorTypeForStruct( ss.Type ); ok {
            cr := newCastReactorDefinitionMap( typ, epr.impl.defs )
            cr.skipPathSetter = true
            epr.proc = mgRct.InitReactorPipeline( cr, epr.rct )
            return nil
        }
        typ := ss.Type.AsAtomicType()
        return errorForUnexpectedErrorType( ss.GetPath(), typ )
    } 
    return errorForUnexpectedErrorValue( ev )
}

func ( epr *errorProcReactor ) ProcessEvent( ev mgRct.ReactorEvent ) error {
    if ( epr.proc == nil ) { 
        if err := epr.initProc( ev ); err != nil { return err }
    }
    return epr.proc.ProcessEvent( ev )
}

func ( impl *mgRespImpl ) GetErrorReactor( 
    p objpath.PathNode ) ( mgRct.ReactorEventProcessor, error ) {

    rct, err := impl.iface.GetErrorReactor( p )
    if err != nil { return nil, err }
    return &errorProcReactor{ impl: impl, rct: rct }, nil
}

func NewResponseReactor(
    defs *DefinitionMap,
    svcDef *ServiceDefinition,
    opDef *OperationDefinition,
    iface ResponseReactorInterface ) *mgRct.ResponseReactor {

    resTyp := opDef.Signature.Return
    qn := mg.TypeNameIn( resTyp ).( *mg.QualifiedTypeName )
    if ! defs.HasKey( qn ) {
        tmpl := "operation '%s' declares unrecognized return type: %s"
        panic( libErrorf( tmpl, opDef.Name, resTyp ) )
    }
    mgIface := 
        &mgRespImpl{ iface: iface, defs: defs, svcDef: svcDef, opDef: opDef }
    return mgRct.NewResponseReactor( mgIface )
}
