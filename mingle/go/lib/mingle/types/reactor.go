package types

import (
    mg "mingle"
    "bitgirder/objpath"
    "container/list"
//    "log"
)

// A synthetic type which we use along with a mingle.CastReactor to cast request
// parameter maps to conform to their respective OperationDefinitions
var typeTypedParameterMap = 
    mg.MustQualifiedTypeName( "mingle:types@v1/TypedParameterMap" )

type castIface struct { 
    dm *DefinitionMap 
}

func ( ci castIface ) InferStructFor( qn *mg.QualifiedTypeName ) bool {
    if def, ok := ci.dm.GetOk( qn ); ok {
        _, res := def.( *StructDefinition )
        return res
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
                tmpl := "Super type %s of %s is not a struct"
                panic( libErrorf( tmpl, spr, sd.GetName() ) )
            }
        } else {
            tmpl := "Can't find super type %s of %s"
            panic( libErrorf( tmpl, spr, sd.GetName() ) )
        }
    }
    return flds
}

func typeNameIn( fd *FieldDefinition ) *mg.QualifiedTypeName {
    nm := mg.TypeNameIn( fd.Type )
    if qn, ok := nm.( *mg.QualifiedTypeName ); ok { return qn }
    panic( libErrorf( 
        "Name in type %s is not a qname: %s (%T)", fd.Type, nm, nm ) )
}

func expectDef( dm *DefinitionMap, qn *mg.QualifiedTypeName ) Definition {
    if def, ok := dm.GetOk( qn ); ok { return def }
    panic( libErrorf( "map has no definition for type %s", qn ) )
}

func expectAuthTypeOf( 
    secQn *mg.QualifiedTypeName, dm *DefinitionMap ) mg.TypeReference {
    if def, ok := dm.GetOk( secQn ); ok {
        if protDef, ok := def.( *PrototypeDefinition ); ok {
            flds := protDef.Signature.GetFields()
            if fd := flds.Get( mg.IdAuthentication ); fd != nil {
                return fd.Type
            }
            panic( libErrorf( "No auth for security: %s", secQn ) )
        }
        panic( libErrorf( "Not a prototype: %s", secQn ) )
    }
    panic( libErrorf( "No such security def: %s", secQn ) )
}

func valDefOf( fd *FieldDefinition, dm *DefinitionMap ) Definition {
    qn := typeNameIn( fd )
    return expectDef( dm, qn )
}

type fieldTyper struct { 
    flds []*FieldSet 
    dm *DefinitionMap
}

func ( ft fieldTyper ) FieldTypeOf(
    fld *mg.Identifier, pg mg.PathGetter ) ( mg.TypeReference, error ) {
    for _, flds := range ft.flds {
        if fd := flds.Get( fld ); fd != nil { return fd.Type, nil }
    }
    // use parent path since we're positioned on the failed field itself
    par := objpath.ParentOf( pg.GetPath() )
    return nil, mg.NewUnrecognizedFieldError( par, fld )
}

func ( ci castIface ) FieldTyperFor(
    qn *mg.QualifiedTypeName, pg mg.PathGetter ) ( mg.FieldTyper, error ) {
    flds := make( []*FieldSet, 0, 2 )
    for nm := qn; nm != nil; {
        if def, ok := ci.dm.GetOk( nm ); ok {
            if sd, ok := def.( *StructDefinition ); ok {
                flds = append( flds, sd.Fields )
                nm = sd.GetSuperType()
                continue
            } else { return nil, notAStructError( nm, pg.GetPath() ) } 
        }
        nm = nil
    }
    if len( flds ) > 0 { return fieldTyper{ flds: flds, dm: ci.dm }, nil }
    tmpl := "No field type info for type %s"
    return nil, mg.NewValueCastErrorf( pg.GetPath(), tmpl, qn )
}

func completeCastEnum(
    id *mg.Identifier, 
    ed *EnumDefinition, 
    pg mg.PathGetter ) ( *mg.Enum, error ) {
    if res := ed.GetValue( id ); res != nil { return res, nil }
    tmpl := "illegal value for enum %s: %s"
    return nil, mg.NewValueCastErrorf( pg.GetPath(), tmpl, ed.GetName(), id )
}

func castEnumFromString( 
    s string, ed *EnumDefinition, pg mg.PathGetter ) ( *mg.Enum, error ) {
    id, err := mg.ParseIdentifier( s )
    if err != nil {
        p := pg.GetPath()
        tmpl := "invalid enum value %q: %s"
        return nil, mg.NewValueCastErrorf( p, tmpl, s, err )
    }
    return completeCastEnum( id, ed, pg )
}

func castEnum( 
    val mg.Value, ed *EnumDefinition, pg mg.PathGetter ) ( *mg.Enum, error ) {
    switch v := val.( type ) {
    case mg.String: return castEnumFromString( string( v ), ed, pg )
    case *mg.Enum: 
        if v.Type.Equals( ed.GetName() ) {
            return completeCastEnum( v.Value, ed, pg )
        }
    }
    t := ed.GetName().AsAtomicType()
    return nil, mg.NewTypeCastErrorValue( t, val, pg.GetPath() )
}

func ( ci castIface ) CastAtomic(
    v mg.Value,
    at *mg.AtomicTypeReference,
    pg mg.PathGetter ) ( mg.Value, error, bool ) {
    if qn, ok := at.Name.( *mg.QualifiedTypeName ); ok {
        if def, ok := ci.dm.GetOk( qn ); ok {
            if ed, ok := def.( *EnumDefinition ); ok {
                res, err := castEnum( v, ed, pg )
                return res, err, true
            }
        }
    }
    return nil, nil, false
}

type fieldSetGetter interface {
    getFieldSets( 
        qn *mg.QualifiedTypeName, pg mg.PathGetter ) ( []*FieldSet, error )
}

type defMapFieldSetGetter struct { dm *DefinitionMap }

func ( fsg defMapFieldSetGetter ) getFieldSets(
    qn *mg.QualifiedTypeName, pg mg.PathGetter ) ( []*FieldSet, error ) {
    if def, ok := fsg.dm.GetOk( qn ); ok {
        if sd, ok := def.( *StructDefinition ); ok {
            return collectFieldSets( sd, fsg.dm ), nil
        } else { return nil, notAStructError( qn, pg.GetPath() ) }
    } else { return nil, newUnrecognizedTypeError( qn, pg.GetPath() ) }
    return nil, nil
}

type castReactor struct {
    castBase *mg.CastReactor
    dm *DefinitionMap
    fsg fieldSetGetter
    stack *list.List
    deflFeed *mg.EventPathReactor
}

func ( cr *castReactor ) Init( rpi *mg.ReactorPipelineInit ) {
    rpi.AddPipelineProcessor( cr.castBase )
}

func ( cr *castReactor ) GetPath() objpath.PathNode {
    res := cr.castBase.GetPath()
    if cr.deflFeed != nil { res = cr.deflFeed.AppendPath( res ) }
    return res
}

func ( cr *castReactor ) newValueCastErrorf( 
    tmpl string, args ...interface{} ) error {
    return mg.NewValueCastErrorf( cr.GetPath(), tmpl, args... )
}

func newUnrecognizedTypeError(
    qn *mg.QualifiedTypeName, p objpath.PathNode ) error {
    return mg.NewValueCastErrorf( p, "Unrecognized type: %s", qn )
}

func ( cr *castReactor ) newUnrecognizedTypeError( 
    qn *mg.QualifiedTypeName ) error {
    return newUnrecognizedTypeError( qn, cr.GetPath() )
}

func notAStructError( qn *mg.QualifiedTypeName, p objpath.PathNode ) error {
    return mg.NewValueCastErrorf( p, "Not a struct type: %s", qn )
}

func ( cr *castReactor ) notAStructError( qn *mg.QualifiedTypeName ) error {
    return notAStructError( qn, cr.GetPath() )
}

type fieldCtx struct {
    endCount int
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

func ( cr *castReactor ) peek() *fieldCtx {
    if cr.stack.Len() == 0 { panic( libError( "fieldCtx stack empty" ) ) }
    return cr.stack.Front().Value.( *fieldCtx )
}

func ( cr *castReactor ) startStruct( typ *mg.QualifiedTypeName ) error {
    if flds, err := cr.fsg.getFieldSets( typ, cr ); err == nil {
        if flds != nil { cr.stack.PushFront( cr.newFieldCtx( flds ) ) }
    } else { return err }
    return nil
}

// We don't re-check here that fld is actually part of the defined field set or
// that this is the first time seeing it, since the upstream structural reactor
// and castIface will have validated that already
func ( cr *castReactor ) startField( fld *mg.Identifier ) {
    if cr.stack.Len() == 0 { return }
    fldCtx := cr.peek()
    fldCtx.await.Delete( fld )
}

func ( cr *castReactor ) feedDefault( 
    fld *mg.Identifier, defl mg.Value, rep mg.ReactorEventProcessor ) error {
    cr.deflFeed = mg.NewEventPathReactor( rep )
    defer func() { cr.deflFeed = nil }()
    fs := mg.FieldStartEvent{ fld }
    if err := cr.deflFeed.ProcessEvent( fs ); err != nil { return err }
    return mg.VisitValue( defl, cr.deflFeed )
}

func ( cr *castReactor ) processDefaults(
    fldCtx *fieldCtx, ee mg.EndEvent, rep mg.ReactorEventProcessor ) error {
    vis := func( fld *mg.Identifier, val interface{} ) error {
        fd := val.( *FieldDefinition )
        if defl := fd.GetDefault(); defl != nil { 
            if err := cr.feedDefault( fld, defl, rep ); err != nil { 
                return err 
            }
            fldCtx.await.Delete( fld )
        }
        return nil
    }
    return fldCtx.await.EachPairError( vis )
}

func ( cr *castReactor ) createMissingFieldsError( fldCtx *fieldCtx ) error {
    flds := make( []*mg.Identifier, 0, fldCtx.await.Len() )
    fldCtx.await.EachPair( func( fld *mg.Identifier, _ interface{} ) {
        flds = append( flds, fld )
    })
    return mg.NewMissingFieldsError( cr.GetPath(), flds )
}

func ( cr *castReactor ) end( 
    ev mg.EndEvent, rep mg.ReactorEventProcessor ) error {
    if cr.stack.Len() == 0 { return nil }
    fldCtx := cr.peek()
    if fldCtx.endCount == 0 {
        defer cr.stack.Remove( cr.stack.Front() )
        if err := cr.processDefaults( fldCtx, ev, rep ); err != nil {
            return err
        }
        fldCtx.removeOptFields()
        if fldCtx.await.Len() > 0 { 
            return cr.createMissingFieldsError( fldCtx )
        } else { return nil }
    }
    fldCtx.endCount--
    return nil
}

func ( cr *castReactor ) startContainer() {
    if cr.stack.Len() == 0 { return }
    cr.peek().endCount++
}

func ( cr *castReactor ) checkValue( v mg.Value ) error {
    if en, ok := v.( *mg.Enum ); ok {
        if def, ok := cr.dm.GetOk( en.Type ); ok {
            if _, ok := def.( *EnumDefinition ); ok {
                return nil // Later we'll check the value too
            } else {
                tmpl := "Not an enum type: %s"
                return cr.newValueCastErrorf( tmpl, en.Type )
            }
        } else { return cr.newUnrecognizedTypeError( en.Type ) }
    }
    return nil
}

func ( cr *castReactor ) ProcessEvent( 
    ev mg.ReactorEvent, rep mg.ReactorEventProcessor ) error {
    switch v := ev.( type ) {
    case mg.StructStartEvent: 
        if err := cr.startStruct( v.Type ); err != nil { return err }
    case mg.EndEvent: if err := cr.end( v, rep ); err != nil { return err }
    case mg.FieldStartEvent: cr.startField( v.Field )
    case mg.MapStartEvent, mg.ListStartEvent: cr.startContainer()
    case mg.ValueEvent: 
        if err := cr.checkValue( v.Val ); err != nil { return err }
    }
    return rep.ProcessEvent( ev )
}

func newCastReactor(
    typ mg.TypeReference, 
    iface mg.CastInterface,
    dm *DefinitionMap, 
    fsg fieldSetGetter,
    pg mg.PathGetter ) mg.PipelineProcessor {
    castBase := mg.NewCastReactor( typ, iface, pg )
    return &castReactor{ 
        castBase: castBase, 
        dm: dm,
        fsg: fsg,
        stack: &list.List{},
    }
}

func newCastReactor1(
    typ mg.TypeReference,
    dm *DefinitionMap,
    pg mg.PathGetter ) mg.PipelineProcessor {
    fsg := defMapFieldSetGetter{ dm }
    return newCastReactor( typ, castIface{ dm }, dm, fsg, pg )
}

func NewCastReactor( 
    typ mg.TypeReference, dm *DefinitionMap ) mg.PipelineProcessor {
    return newCastReactor1( typ, dm, nil )
}

type opMatcher struct {

    svcDefs *ServiceDefinitionMap

    ns *mg.Namespace
    sd *ServiceDefinition
    opDef *OperationDefinition
}

func ( om *opMatcher ) defMap() *DefinitionMap {
    return om.svcDefs.GetDefinitionMap()
}

func ( om *opMatcher ) Namespace( ns *mg.Namespace, pg mg.PathGetter ) error {
    if ! om.svcDefs.HasNamespace( ns ) {
        return mg.NewEndpointErrorNamespace( ns, pg.GetPath() )
    }
    om.ns = ns
    return nil
}

func ( om *opMatcher ) Service( svc *mg.Identifier, pg mg.PathGetter ) error {
    if sd, ok := om.svcDefs.GetOk( om.ns, svc ); ok {
        om.sd = sd
        return nil
    }
    return mg.NewEndpointErrorService( svc, pg.GetPath() )
}

func ( om *opMatcher ) Operation( op *mg.Identifier, pg mg.PathGetter ) error {
    if om.opDef = om.sd.findOperation( op ); om.opDef == nil {
        return mg.NewEndpointErrorOperation( op, pg.GetPath() )
    }
    return nil
}

type OpMatch interface {
}

type RequestReactorInterface interface {

    GetAuthenticationProcessor( 
        om OpMatch, pg mg.PathGetter ) ( mg.ReactorEventProcessor, error )

    GetParametersProcessor(
        om OpMatch, pg mg.PathGetter ) ( mg.ReactorEventProcessor, error )
}

type mgReqImpl struct {

    *opMatcher
    iface RequestReactorInterface

    sawAuth bool
}

func ( impl *mgReqImpl ) needsAuth() bool { return impl.sd.Security != nil }

func ( impl *mgReqImpl ) GetAuthenticationProcessor( 
    pg mg.PathGetter ) ( mg.ReactorEventProcessor, error ) {
    impl.sawAuth = true
    pg = mg.ImmediatePathGetter{ pg.GetPath() }
    if secQn := impl.sd.Security; secQn != nil { 
        authTyp := expectAuthTypeOf( secQn, impl.defMap() )
        dm := impl.defMap()
        fsg := defMapFieldSetGetter{ dm }
        cr := newCastReactor( authTyp, castIface{ dm }, dm, fsg, pg )
        authRct, err := 
            impl.iface.GetAuthenticationProcessor( impl.opMatcher, pg )
        if err != nil { return nil, err }
        return mg.InitReactorPipeline( cr, authRct ), nil
    }
    return mg.DiscardProcessor, nil
}

type parametersCastIface struct {
    ci castIface
    opDef *OperationDefinition
    defs *DefinitionMap
}

func ( pi parametersCastIface ) InferStructFor( 
    qn *mg.QualifiedTypeName ) bool {
    if qn.Equals( typeTypedParameterMap ) { return true }
    return pi.ci.InferStructFor( qn )
}

func ( pi parametersCastIface ) fieldSets() []*FieldSet {
    return []*FieldSet{ pi.opDef.Signature.Fields }
}

func ( pi parametersCastIface ) FieldTyperFor(
    qn *mg.QualifiedTypeName, pg mg.PathGetter ) ( mg.FieldTyper, error ) {
    if qn.Equals( typeTypedParameterMap ) { 
        return fieldTyper{ flds: pi.fieldSets(), dm: pi.defs }, nil
    }
    return pi.ci.FieldTyperFor( qn, pg )
}

func ( pi parametersCastIface ) CastAtomic(
    in mg.Value, 
    at *mg.AtomicTypeReference, 
    pg mg.PathGetter ) ( mg.Value, error, bool ) {
    return pi.ci.CastAtomic( in, at, pg )
}

func ( pi parametersCastIface ) getFieldSets(
    qn *mg.QualifiedTypeName, pg mg.PathGetter ) ( []*FieldSet, error ) {
    if qn.Equals( typeTypedParameterMap ) { return pi.fieldSets(), nil }
    return ( defMapFieldSetGetter{ pi.defs } ).getFieldSets( qn, pg )
}

type parametersReactor struct { rep mg.ReactorEventProcessor }

func ( pr parametersReactor ) ProcessEvent( ev mg.ReactorEvent ) error {
    if ss, ok := ev.( mg.StructStartEvent ); ok {
        if ss.Type.Equals( typeTypedParameterMap ) { ev = mg.EvMapStart }
    }
    return pr.rep.ProcessEvent( ev )
}

func ( impl *mgReqImpl ) checkGotAuth( pg mg.PathGetter ) error {
    if impl.needsAuth() {
        if ! impl.sawAuth {
            // take parent since pg itself will be at 'parameters'
            path := objpath.ParentOf( pg.GetPath() )
            flds := []*mg.Identifier{ mg.IdAuthentication }
            return mg.NewMissingFieldsError( path, flds )
        }
    }
    return nil
}

func ( impl *mgReqImpl ) GetParametersProcessor(
    pg mg.PathGetter ) ( mg.ReactorEventProcessor, error ) {
    if err := impl.checkGotAuth( pg ); err != nil { return nil, err }
    pg = mg.ImmediatePathGetter{ pg.GetPath() }
    dm := impl.defMap()
    pci := parametersCastIface{ 
        ci: castIface{ dm },
        defs: dm,
        opDef: impl.opDef,
    }
    typ := typeTypedParameterMap.AsAtomicType()
    cr := newCastReactor( typ, pci, dm, pci, pg )
    rep, err := impl.iface.GetParametersProcessor( impl.opMatcher, pg )
    if err != nil { return nil, err }
    paramsRct := parametersReactor{ rep }
    return mg.InitReactorPipeline( 
        cr, 
//        mg.NewDebugReactor( mg.DebugLoggerFunc( log.Printf ) ),
        paramsRct,
    ), nil
}

func NewRequestReactor( 
    svcDefs *ServiceDefinitionMap, 
    iface RequestReactorInterface ) *mg.ServiceRequestReactor {
    return mg.NewServiceRequestReactor(
        &mgReqImpl{ 
            opMatcher: &opMatcher{ svcDefs: svcDefs },
            iface: iface,
        },
    )
}

type errorProcReactor struct {
    defs *DefinitionMap
    throws []mg.TypeReference
    proc mg.ReactorEventProcessor
    pg mg.PathGetter
}

func errorForUnexpectedErrorType( 
    path objpath.PathNode, typ mg.TypeReference ) error {
    return mg.NewValueCastErrorf( path,
        "Error type is not a declared thrown type: %s", typ )
}

func ( epr *errorProcReactor ) errorForUnexpectedErrorType( 
    typ mg.TypeReference ) error {
    return errorForUnexpectedErrorType( epr.pg.GetPath(), typ )
}

func ( epr *errorProcReactor ) errorForUnexpectedErrorQname(
    qn *mg.QualifiedTypeName ) error {
    typ := &mg.AtomicTypeReference{ Name: qn }
    return epr.errorForUnexpectedErrorType( typ )
}

func ( epr *errorProcReactor ) errorForUnexpectedErrorValue( 
    ev mg.ReactorEvent ) error {
    var typ mg.TypeReference
    switch v := ev.( type ) {
    case mg.ValueEvent: typ = mg.TypeOf( v.Val )
    case mg.ListStartEvent: typ = mg.TypeOpaqueList
    case mg.MapStartEvent: typ = mg.TypeSymbolMap
    }
    if ( typ == nil ) { panic( libErrorf( "unhandled event type: %T", ev ) ) }
    return epr.errorForUnexpectedErrorType( typ )
}

func ( epr *errorProcReactor ) errorTypeForStruct( 
    qn *mg.QualifiedTypeName ) ( mg.TypeReference, bool ) {
    if epr.defs.HasBuiltInDefinition( qn ) {
        return &mg.AtomicTypeReference{ Name: qn }, true
    }
    for _, typ := range epr.throws {
        if mg.TypeNameIn( typ ).Equals( qn ) { return typ, true }
    }
    return nil, false
}

func ( epr *errorProcReactor ) ProcessEvent( ev mg.ReactorEvent ) error {
    if ( epr.proc == nil ) { 
        if ss, ok := ev.( mg.StructStartEvent ); ok {
            if typ, ok := epr.errorTypeForStruct( ss.Type ); ok {
                cr := newCastReactor1( typ, epr.defs, epr.pg )
                epr.proc = mg.InitReactorPipeline( cr )
            } else { return epr.errorForUnexpectedErrorQname( ss.Type ) }
        } else { return epr.errorForUnexpectedErrorValue( ev ) }
    }
    return epr.proc.ProcessEvent( ev )
}

type ResponseReactorInterface interface {
    GetResultProcessor( pg mg.PathGetter ) ( mg.ReactorEventProcessor, error )
    GetErrorProcessor( pg mg.PathGetter ) ( mg.ReactorEventProcessor, error )
}

type mgRespImpl struct {
    defs *DefinitionMap
    opDef *OperationDefinition
    iface ResponseReactorInterface
}

func ( impl *mgRespImpl ) GetResultProcessor( 
    pg mg.PathGetter ) ( mg.ReactorEventProcessor, error ) {
    resTyp := impl.opDef.Signature.Return
    ipg := mg.ImmediatePathGetter{ pg.GetPath() }
    cr := newCastReactor1( resTyp, impl.defs, ipg )
    rct, err := impl.iface.GetResultProcessor( pg )
    if err != nil { return nil, err }
    return mg.InitReactorPipeline( cr, rct ), nil
}

func ( impl *mgRespImpl ) GetErrorProcessor( 
    pg mg.PathGetter ) ( mg.ReactorEventProcessor, error ) {
    rct, err := impl.iface.GetErrorProcessor( pg )
    if err != nil { return nil, err }
    epr := &errorProcReactor{ 
        defs: impl.defs, 
        throws: impl.opDef.Signature.Throws,
        pg: mg.ImmediatePathGetter{ pg.GetPath() },
    }
    return mg.InitReactorPipeline( epr, rct ), nil
}

func NewResponseReactor(
    defs *DefinitionMap,
    opDef *OperationDefinition,
    iface ResponseReactorInterface ) *mg.ServiceResponseReactor {
    resTyp := opDef.Signature.Return
    qn := mg.TypeNameIn( resTyp ).( *mg.QualifiedTypeName )
    if ! defs.HasKey( qn ) {
        tmpl := "operation '%s' declares unrecognized return type: %s"
        panic( libErrorf( tmpl, opDef.Name, resTyp ) )
    }
    mgIface := &mgRespImpl{ iface: iface, defs: defs, opDef: opDef }
    return mg.NewServiceResponseReactor( mgIface )
}
