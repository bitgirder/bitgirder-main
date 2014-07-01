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

func notAFieldSetTypeError( 
    p objpath.PathNode, qn *mg.QualifiedTypeName ) error {

    return mg.NewValueCastErrorf( p, "not a type with fields: %s", qn )
}

func newUnrecognizedTypeError(
    p objpath.PathNode, qn *mg.QualifiedTypeName ) error {

    return mg.NewValueCastErrorf( p, "unrecognized type: %s", qn )
}

func asMapStartEvent( ev mgRct.ReactorEvent ) *mgRct.MapStartEvent {
    res := mgRct.NewMapStartEvent( mg.PointerIdNull ) 
    res.SetPath( ev.GetPath() )
    return res
}

type defMapCastIface struct { dm *DefinitionMap }

func ( ci defMapCastIface ) InferStructFor( qn *mg.QualifiedTypeName ) bool {
    if def, ok := ci.dm.GetOk( qn ); ok {
        if _, ok = def.( *StructDefinition ); ok { return true }
        if _, ok = def.( *SchemaDefinition ); ok { return true }
    }
    return false
}

func ( ci defMapCastIface ) AllowAssignment(
    expct, act *mg.QualifiedTypeName ) bool {

    if _, ok := ci.dm.GetOk( act ); ! ok { return false }
    return canAssignType( expct, act, ci.dm )
}

type fieldTyper struct { 
    flds *FieldSet 
    dm *DefinitionMap
    ignoreUnrecognized bool
}

func ( ft fieldTyper ) FieldTypeFor(
    fld *mg.Identifier, path objpath.PathNode ) ( mg.TypeReference, error ) {

    if fd := ft.flds.Get( fld ); fd != nil { return fd.Type, nil }
    if ft.ignoreUnrecognized { return mg.TypeValue, nil }
    return nil, mg.NewUnrecognizedFieldError( path, fld )
}

func ( ci defMapCastIface ) fieldTyperForStruct(
    def *StructDefinition, path objpath.PathNode ) ( FieldTyper, error ) {

    return fieldTyper{ flds: def.Fields, dm: ci.dm }, nil
}

func ( ci defMapCastIface ) fieldTyperForSchema( 
    sd *SchemaDefinition ) fieldTyper {

    return fieldTyper{ flds: sd.Fields, dm: ci.dm, ignoreUnrecognized: true }
}

func ( ci defMapCastIface ) FieldTyperFor(
    qn *mg.QualifiedTypeName, path objpath.PathNode ) ( FieldTyper, error ) {

    if def, ok := ci.dm.GetOk( qn ); ok {
        switch v := def.( type ) {
        case *StructDefinition: return ci.fieldTyperForStruct( v, path )
        case *SchemaDefinition: return ci.fieldTyperForSchema( v ), nil
        default: return nil, notAFieldSetTypeError( path, qn )
        }
    }
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
        if ev, ok := v.( *mg.Enum ); ok {
            if ! ev.Type.Equals( at.Name ) { return nil, nil, false }
            if _, ok := def.( *StructDefinition ); ok {
                return nil, notAnEnumTypeError( at, path ), true
            }
        }
    }
    return nil, nil, false
}

type fieldSetGetter interface {

    getFieldSet( 
        qn *mg.QualifiedTypeName, path objpath.PathNode ) ( *FieldSet, error )
}

func fieldSetForTypeInDefMap(
    qn *mg.QualifiedTypeName, 
    dm *DefinitionMap, 
    path objpath.PathNode ) ( *FieldSet, error ) {

    if def, ok := dm.GetOk( qn ); ok {
        switch v := def.( type ) {
        case *StructDefinition: return v.Fields, nil
        case *SchemaDefinition: return v.Fields, nil
        default: return nil, notAFieldSetTypeError( path, qn )
        } 
    } 
    return nil, newUnrecognizedTypeError( path, qn )
}

type defMapFieldSetGetter struct { dm *DefinitionMap }

func ( fsg defMapFieldSetGetter ) getFieldSet(
    qn *mg.QualifiedTypeName, path objpath.PathNode ) ( *FieldSet, error ) {

    return fieldSetForTypeInDefMap( qn, fsg.dm, path )
}

type castReactor struct {
    typ mg.TypeReference
    iface CastInterface
    dm *DefinitionMap
    fsg fieldSetGetter
    stack *stack.Stack
    skipPathSetter bool
}

func ( cr *castReactor ) shouldSuppressAllocation( typ mg.TypeReference ) bool {
    switch v := typ.( type ) {
    case *mg.AtomicTypeReference:
        if sd, hasDef := cr.dm.GetOk( v.Name ); hasDef {
            _, isSchema := sd.( *SchemaDefinition )
            return ! isSchema
        }
    case *mg.NullableTypeReference: return cr.shouldSuppressAllocation( v.Type )
    }
    return true
}

func ( cr *castReactor ) InitializePipeline( pip *pipeline.Pipeline ) {
    mgCastRct := NewCastReactor( cr.typ, cr.iface )
    mgCastRct.SkipPathSetter = cr.skipPathSetter
    mgCastRct.ShouldSuppressAllocation = func( typ mg.TypeReference ) bool {
        return cr.shouldSuppressAllocation( typ )
    }
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

func ( cr *castReactor ) newFieldCtx( flds *FieldSet ) *fieldCtx {
    res := &fieldCtx{ await: mg.NewIdentifierMap() }
    flds.EachDefinition( func( fd *FieldDefinition ) {
        res.await.Put( fd.Name, fd )
    })
    return res
}

func ( cr *castReactor ) startStruct( 
    ss *mgRct.StructStartEvent ) ( mgRct.ReactorEvent, error ) {

    flds, err := cr.fsg.getFieldSet( ss.Type, ss.GetPath() )
    if err != nil { return nil, err }
    if flds != nil { cr.stack.Push( cr.newFieldCtx( flds ) ) }
    if def, ok := cr.dm.GetOk( ss.Type ); ok {
        if _, ok := def.( *SchemaDefinition ); ok {
            return asMapStartEvent( ss ), nil
        }
    } 
    return ss, nil
}

// We don't re-check here that fld is actually part of the defined field set or
// since the upstream defMapCastIface will have validated that already
func ( cr *castReactor ) startField( 
    fs *mgRct.FieldStartEvent ) ( mgRct.ReactorEvent, error ) {

    if cr.stack.IsEmpty() { return fs, nil }
    cr.stack.Peek().( *fieldCtx ).await.Delete( fs.Field )
    return fs, nil
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
    ev *mgRct.EndEvent, 
    next mgRct.ReactorEventProcessor ) ( mgRct.ReactorEvent, error ) {

    if cr.stack.IsEmpty() { return ev, nil }
    fldCtx := cr.stack.Peek().( *fieldCtx )
    if fldCtx.depth > 0 {
        fldCtx.depth--
        return ev, nil
    }
    cr.stack.Pop()
    p := ev.GetPath()
    if err := processDefaults( fldCtx, p, next ); err != nil { return nil, err }
    fldCtx.removeOptFields()
    if fldCtx.await.Len() > 0 { 
        return nil, createMissingFieldsError( p, fldCtx ) 
    }
    return ev, nil
}

func ( cr *castReactor ) startContainer() error {
    if ! cr.stack.IsEmpty() { cr.stack.Peek().( *fieldCtx ).depth++ }
    return nil
}

func notAnEnumTypeError( typ mg.TypeReference, path objpath.PathNode ) error {
    return mg.NewValueCastErrorf( path, "not an enum type: %s", typ )
}

// we only do value checks here that are specific to this cast, namely having to
// do with enums. If the value is an enum, we check that we recogzize the type
// and that the type is actually an enum. We don't actually check the enum value
// here though, and leave that for CastAtomic. Any other values aren't checked
// here and are left to CastAtomic or to the upstream processor.
func ( cr *castReactor ) valueEvent( 
    ve *mgRct.ValueEvent ) ( mgRct.ReactorEvent, error ) {

    if en, ok := ve.Val.( *mg.Enum ); ok {
        if def, ok := cr.dm.GetOk( en.Type ); ok {
            if _, ok := def.( *EnumDefinition ); ok { return ve, nil }
            enTyp := en.Type.AsAtomicType()
            return nil, notAnEnumTypeError( enTyp, ve.GetPath() )
        } 
        return nil, newUnrecognizedTypeError( ve.GetPath(), en.Type )
    }
    return ve, nil
}

func ( cr *castReactor ) prepareProcessEvent(
    ev mgRct.ReactorEvent, 
    next mgRct.ReactorEventProcessor ) ( mgRct.ReactorEvent, error ) {
    
    switch v := ev.( type ) {
    case *mgRct.StructStartEvent: return cr.startStruct( v )
    case *mgRct.FieldStartEvent: return cr.startField( v )
    case *mgRct.ValueEvent: return cr.valueEvent( v )
    case *mgRct.EndEvent: return cr.end( v, next )
    case *mgRct.ListStartEvent, *mgRct.MapStartEvent: 
        return ev, cr.startContainer()
    case *mgRct.ValueAllocationEvent, *mgRct.ValueReferenceEvent: return ev, nil
    }
    panic( libErrorf( "unhandled event: %T", ev ) )
}

func ( cr *castReactor ) ProcessEvent( 
    ev mgRct.ReactorEvent, next mgRct.ReactorEventProcessor ) error {

    ev, err := cr.prepareProcessEvent( ev, next )
    if err != nil { return err }
    return next.ProcessEvent( ev )
}

func newCastReactorBase(
    typ mg.TypeReference, 
    iface CastInterface,
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
