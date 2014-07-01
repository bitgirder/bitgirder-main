package types

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "bitgirder/objpath"
    "bitgirder/pipeline"
    "bitgirder/stack"
    "log"
    "fmt"
    "bytes"
)

func asMapStartEvent( ev mgRct.ReactorEvent ) *mgRct.MapStartEvent {
    res := mgRct.NewMapStartEvent( mg.PointerIdNull ) 
    res.SetPath( ev.GetPath() )
    return res
}

func notAFieldSetTypeError( 
    p objpath.PathNode, qn *mg.QualifiedTypeName ) error {

    return mg.NewValueCastErrorf( p, "not a type with fields: %s", qn )
}

func newUnrecognizedTypeError(
    p objpath.PathNode, qn *mg.QualifiedTypeName ) error {

    return mg.NewValueCastErrorf( p, "unrecognized type: %s", qn )
}

func notAnEnumTypeError( typ mg.TypeReference, path objpath.PathNode ) error {
    return mg.NewValueCastErrorf( path, "not an enum type: %s", typ )
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

type FieldTyper interface {

    // path will be positioned to the map/struct containing fld, but will not
    // itself include fld
    FieldTypeFor( 
        fld *mg.Identifier, path objpath.PathNode ) ( mg.TypeReference, error )
}

type valueFieldTyper int

func ( vt valueFieldTyper ) FieldTypeFor( 
    fld *mg.Identifier, path objpath.PathNode ) ( mg.TypeReference, error ) {
    return mg.TypeNullableValue, nil
}

type CastReactor struct {

    dm *DefinitionMap

    stack *stack.Stack

    SkipPathSetter bool

    CastingList func( le *mgRct.ListStartEvent, lt *mg.ListTypeReference ) error

    AllocationSuppressed func( ve *mgRct.ValueAllocationEvent ) error

    ProcessValueReference func( 
        ev *mgRct.ValueReferenceEvent, 
        typ mg.TypeReference,
        next mgRct.ReactorEventProcessor ) error
}

func ( cr *CastReactor ) dumpStack( pref string ) {
    bb := &bytes.Buffer{}
    fmt.Fprintf( bb, "%s: [", pref )
    cr.stack.VisitTop( func( v interface{} ) {
        msg := fmt.Sprintf( "%T", v )
        switch v2 := v.( type ) {
        case mg.TypeReference: msg = v2.ExternalForm()
        }
        fmt.Fprintf( bb, msg )
        fmt.Fprintf( bb, ", " )
    })
    fmt.Fprintf( bb, " ]" )
    log.Print( bb.String() )
}

func NewCastReactor0( expct mg.TypeReference ) *CastReactor {
    res := &CastReactor{ stack: stack.NewStack() }
    res.stack.Push( expct )
    return res
}

func ( cr *CastReactor ) InitializePipeline( pip *pipeline.Pipeline ) {
    mgRct.EnsureStructuralReactor( pip )
    if ! cr.SkipPathSetter { mgRct.EnsurePathSettingProcessor( pip ) }
}

type fieldCtx struct {
    depth int
    await *mg.IdentifierMap
}

func newFieldCtx( flds *FieldSet ) *fieldCtx {
    res := &fieldCtx{ await: mg.NewIdentifierMap() }
    flds.EachDefinition( func( fd *FieldDefinition ) {
        res.await.Put( fd.Name, fd )
    })
    return res
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

type listCast struct {
    sawValues bool
    lt *mg.ListTypeReference
    startPath objpath.PathNode
}

type valAllocCast struct { 
    typ mg.TypeReference 
    id mg.PointerId
}

func ( cr *CastReactor ) errStackUnrecognized() error {
    return libErrorf( "unrecognized stack element: %T", cr.stack.Peek() )
}

func ( cr *CastReactor ) castAtomic(
    v mg.Value,
    at *mg.AtomicTypeReference,
    path objpath.PathNode ) ( mg.Value, error, bool ) {

    if def, ok := cr.dm.GetOk( at.Name ); ok {
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

func ( cr *CastReactor ) valueEventForAtomicCast( 
    ve *mgRct.ValueEvent, 
    at *mg.AtomicTypeReference, 
    callTyp mg.TypeReference ) ( error, *mgRct.ValueEvent ) {

    mv, err, ok := cr.castAtomic( ve.Val, at, ve.GetPath() )
    if ! ok { 
        mv, err = castAtomicWithCallType( ve.Val, at, callTyp, ve.GetPath() ) 
    }
    if err != nil { return err, nil }
    res := mgRct.CopyEvent( ve, true ).( *mgRct.ValueEvent )
    res.Val = mv
    return nil, res
}

func ( cr *CastReactor ) sendAllocEvent( 
    ev *mgRct.ValueAllocationEvent, next mgRct.ReactorEventProcessor ) error {
    
    cr.stack.Push( valAllocCast{ typ: ev.Type, id: ev.Id } )
    return next.ProcessEvent( ev )
}

func ( cr *CastReactor ) sendSynthAllocEvent( 
    typ mg.TypeReference, 
    ev mgRct.ReactorEvent, 
    next mgRct.ReactorEventProcessor ) error {

    alloc := mgRct.NewValueAllocationEvent( typ, mg.PointerIdNull )
    alloc.SetPath( ev.GetPath() )
    return cr.sendAllocEvent( alloc, next )
}

func ( cr *CastReactor ) completedValue() {
    if _, ok := cr.stack.Peek().( valAllocCast ); ok { cr.stack.Pop() }
}

func ( cr *CastReactor ) processAtomicValue(
    ve *mgRct.ValueEvent,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    err, ve2 := cr.valueEventForAtomicCast( ve, at, callTyp )
    if err != nil { return err }
    if err = next.ProcessEvent( ve2 ); err != nil { return err }
    cr.completedValue()
    return nil
}

func ( cr *CastReactor ) processPointerValue(
    ve *mgRct.ValueEvent,
    pt *mg.PointerTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    if err := cr.sendSynthAllocEvent( pt.Type, ve, next ); err != nil { 
        return err 
    }
    return cr.processValueWithType( ve, pt.Type, callTyp, next )
}

func nullValueEventForType( 
    ve *mgRct.ValueEvent, typ mg.TypeReference ) *mgRct.ValueEvent {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference: return ve
    case *mg.PointerTypeReference: return nullValueEventForType( ve, v.Type )
    case *mg.NullableTypeReference: return nullValueEventForType( ve, v.Type )
    case *mg.ListTypeReference: return ve
    }
    panic( libErrorf( "unhandled type reference: %T", typ ) )
}

func ( cr *CastReactor ) processNullableValue(
    ve *mgRct.ValueEvent,
    nt *mg.NullableTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    if _, ok := ve.Val.( *mg.Null ); ok { 
        return next.ProcessEvent( nullValueEventForType( ve, nt ) ) 
    }
    return cr.processValueWithType( ve, nt.Type, callTyp, next )
}

func ( cr *CastReactor ) processValueForListType(
    ve *mgRct.ValueEvent,
    typ *mg.ListTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    if _, ok := ve.Val.( *mg.Null ); ok {
        return newNullValueCastError( ve.GetPath() )
    }
    return mg.NewTypeCastErrorValue( callTyp, ve.Val, ve.GetPath() )
}

func ( cr *CastReactor ) processValueWithType(
    ve *mgRct.ValueEvent,
    typ mg.TypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference: 
        return cr.processAtomicValue( ve, v, callTyp, next )
    case *mg.PointerTypeReference:
        return cr.processPointerValue( ve, v, callTyp, next )
    case *mg.NullableTypeReference:
        return cr.processNullableValue( ve, v, callTyp, next )
    case *mg.ListTypeReference:
        return cr.processValueForListType( ve, v, callTyp, next )
    }
    panic( libErrorf( "unhandled type: %T", typ ) )
}

func ( cr *CastReactor ) processValue( 
    ve *mgRct.ValueEvent, next mgRct.ReactorEventProcessor ) error {

    switch v := cr.stack.Peek().( type ) {
    case valAllocCast: return cr.processValueWithType( ve, v.typ, v.typ, next )
    case mg.TypeReference: 
        cr.stack.Pop()
        return cr.processValueWithType( ve, v, v, next )
    case *listCast:
        v.sawValues = true
        typ := v.lt.ElementType
        return cr.processValueWithType( ve, typ, typ, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) processValueAllocationWithPointerType(
    pt *mg.PointerTypeReference, 
    ev *mgRct.ValueAllocationEvent,
    next mgRct.ReactorEventProcessor ) error {

    ev2 := mgRct.CopyEvent( ev, true ).( *mgRct.ValueAllocationEvent )
    ev2.Type = pt.Type
    if ev.Id != mg.PointerIdNull && cr.AllocationSuppressed != nil {
        if ! pt.Type.Equals( ev.Type ) {
            if err := cr.AllocationSuppressed( ev ); err != nil { return err }
        }
    }
    return cr.sendAllocEvent( ev2, next )
}

func ( cr *CastReactor ) shouldSuppressAllocation( typ mg.TypeReference ) bool {
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

func ( cr *CastReactor ) processValueAllocationWithoutPointerType(
    ev *mgRct.ValueAllocationEvent,
    typ mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    if ! cr.shouldSuppressAllocation( typ ) { return next.ProcessEvent( ev ) }

    if ev.Id != mg.PointerIdNull && cr.AllocationSuppressed != nil {
        if err := cr.AllocationSuppressed( ev ); err != nil { return err }
    }
    return nil
}

func ( cr *CastReactor ) processValueAllocation(
    ev *mgRct.ValueAllocationEvent, next mgRct.ReactorEventProcessor ) error {

    switch v := cr.stack.Peek().( type ) {
    case *mg.PointerTypeReference: 
        cr.stack.Pop()
        return cr.processValueAllocationWithPointerType( v, ev, next )
    case mg.TypeReference: 
        return cr.processValueAllocationWithoutPointerType( ev, v, next )
    case *listCast:
        typ := v.lt.ElementType
        return cr.processValueAllocationWithoutPointerType( ev, typ, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) processValueReference(
    ev *mgRct.ValueReferenceEvent, next mgRct.ReactorEventProcessor ) error {

    typ := cr.stack.Pop().( mg.TypeReference )
    cr.dumpStack( fmt.Sprintf( "before processing ref %s", typ ) )
    if cr.ProcessValueReference != nil { 
        if err := cr.ProcessValueReference( ev, typ, next ); err != nil { 
            return err 
        }
        return nil
    }
    return next.ProcessEvent( ev )
}

func ( cr *CastReactor ) implMapStart(
    ev mgRct.ReactorEvent, 
    ft FieldTyper, 
    next mgRct.ReactorEventProcessor ) error {

    cr.stack.Push( ft )
    return next.ProcessEvent( ev )
}

type fieldTyper struct { 
    flds *FieldSet 
    dm *DefinitionMap
    ignoreUnrecognized bool
}

func ( ft *fieldTyper ) FieldTypeFor(
    fld *mg.Identifier, path objpath.PathNode ) ( mg.TypeReference, error ) {

    if fd := ft.flds.Get( fld ); fd != nil { return fd.Type, nil }
    if ft.ignoreUnrecognized { return mg.TypeValue, nil }
    return nil, mg.NewUnrecognizedFieldError( path, fld )
}

func ( cr *CastReactor ) fieldTyperForStruct(
    def *StructDefinition, path objpath.PathNode ) ( *fieldTyper, error ) {

    return &fieldTyper{ flds: def.Fields, dm: cr.dm }, nil
}

func ( cr *CastReactor ) fieldTyperForSchema( 
    sd *SchemaDefinition ) *fieldTyper {

    return &fieldTyper{ flds: sd.Fields, dm: cr.dm, ignoreUnrecognized: true }
}

func ( cr *CastReactor ) fieldTyperFor(
    qn *mg.QualifiedTypeName, path objpath.PathNode ) ( *fieldTyper, error ) {

    if def, ok := cr.dm.GetOk( qn ); ok {
        switch v := def.( type ) {
        case *StructDefinition: return cr.fieldTyperForStruct( v, path )
        case *SchemaDefinition: return cr.fieldTyperForSchema( v ), nil
        default: return nil, notAFieldSetTypeError( path, qn )
        }
    }
    tmpl := "no field type info for type %s"
    return nil, mg.NewValueCastErrorf( path, tmpl, qn )
}

func ( cr *CastReactor ) completeStartStruct(
    ss *mgRct.StructStartEvent, next mgRct.ReactorEventProcessor ) error {

    ft, err := cr.fieldTyperFor( ss.Type, ss.GetPath() )
    if err != nil { return err }

    return cr.implMapStart( ss, ft, next )
}

func ( cr *CastReactor ) inferStructForQname( qn *mg.QualifiedTypeName ) bool {
    if def, ok := cr.dm.GetOk( qn ); ok {
        if _, ok = def.( *StructDefinition ); ok { return true }
        if _, ok = def.( *SchemaDefinition ); ok { return true }
    }
    return false
}

func ( cr *CastReactor ) inferStructForMap(
    me *mgRct.MapStartEvent,
    at *mg.AtomicTypeReference,
    next mgRct.ReactorEventProcessor ) ( error, bool ) {

    if ! cr.inferStructForQname( at.Name ) { return nil, false }

    ev := mgRct.NewStructStartEvent( at.Name )
    ev.SetPath( me.GetPath() )

    return cr.completeStartStruct( ev, next ), true
}

func ( cr *CastReactor ) processMapStartWithAtomicType(
    me *mgRct.MapStartEvent,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    if at.Equals( mg.TypeSymbolMap ) || at.Equals( mg.TypeValue ) {
        return cr.implMapStart( me, valueFieldTyper( 1 ), next )
    }

    if err, ok := cr.inferStructForMap( me, at, next ); ok { return err }

    return mg.NewTypeCastError( callTyp, mg.TypeSymbolMap, me.GetPath() )
}

func ( cr *CastReactor ) processMapStartWithPointerType(
    me *mgRct.MapStartEvent,
    pt *mg.PointerTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    if err := cr.sendSynthAllocEvent( mg.TypeSymbolMap, me, next ); err != nil {
        return err
    }
    return cr.processMapStartWithType( me, pt.Type, callTyp, next )
}

func ( cr *CastReactor ) processMapStartWithType(
    me *mgRct.MapStartEvent, 
    typ mg.TypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference:
        return cr.processMapStartWithAtomicType( me, v, callTyp, next )
    case *mg.PointerTypeReference:
        return cr.processMapStartWithPointerType( me, v, callTyp, next )
    case *mg.NullableTypeReference:
        return cr.processMapStartWithType( me, v.Type, callTyp, next )
    }
    return mg.NewTypeCastError( callTyp, typ, me.GetPath() )
}

func ( cr *CastReactor ) processMapStart(
    me *mgRct.MapStartEvent, next mgRct.ReactorEventProcessor ) error {
    
    switch v := cr.stack.Peek().( type ) {
    case mg.TypeReference: 
        cr.stack.Pop()
        return cr.processMapStartWithType( me, v, v, next )
    case valAllocCast:
        return cr.processMapStartWithType( me, v.typ, v.typ, next )
    case *listCast:
        v.sawValues = true
        typ := v.lt.ElementType
        return cr.processMapStartWithType( me, typ, typ, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) processFieldStart(
    fs *mgRct.FieldStartEvent, next mgRct.ReactorEventProcessor ) error {

    ft := cr.stack.Peek().( FieldTyper )
    
    typ, err := ft.FieldTypeFor( fs.Field, fs.GetPath().Parent() )
    if err != nil { return err }

    cr.stack.Push( typ )
    return next.ProcessEvent( fs )
}

func ( cr *CastReactor ) processEnd(
    ee *mgRct.EndEvent, next mgRct.ReactorEventProcessor ) error {

    switch v := cr.stack.Peek().( type ) {
    case *listCast:
        cr.stack.Pop()
        if ! ( v.sawValues || v.lt.AllowsEmpty ) {
            return mg.NewValueCastError( v.startPath, "empty list" )
        }
    case FieldTyper: cr.stack.Pop()
    }

    if err := next.ProcessEvent( ee ); err != nil { return err }
    cr.completedValue()
    return nil
}

func ( cr *CastReactor ) allowAssignment( 
    expct, act *mg.QualifiedTypeName ) bool {

    if _, ok := cr.dm.GetOk( act ); ! ok { return false }
    return canAssignType( expct, act, cr.dm )
}

func ( cr *CastReactor ) processStructStartWithAtomicType(
    ss *mgRct.StructStartEvent,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    if at.Equals( mg.TypeSymbolMap ) {
        me := asMapStartEvent( ss )
        return cr.processMapStartWithAtomicType( me, at, callTyp, next )
    }

    if at.Name.Equals( ss.Type ) || at.Equals( mg.TypeValue ) ||
       cr.allowAssignment( at.Name, ss.Type ) {
        return cr.completeStartStruct( ss, next )
    }

    failTyp := &mg.AtomicTypeReference{ Name: ss.Type }
    return mg.NewTypeCastError( callTyp, failTyp, ss.GetPath() )
}

func ( cr *CastReactor ) processStructStartWithPointerType(
    ss *mgRct.StructStartEvent,
    pt *mg.PointerTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    if err := cr.sendSynthAllocEvent( pt.Type, ss, next ); err != nil { 
        return err 
    }
    return cr.processStructStartWithType( ss, pt.Type, callTyp, next )
}

func ( cr *CastReactor ) processStructStartWithType(
    ss *mgRct.StructStartEvent,
    typ mg.TypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference:
        return cr.processStructStartWithAtomicType( ss, v, callTyp, next )
    case *mg.PointerTypeReference:
        return cr.processStructStartWithPointerType( ss, v, callTyp, next )
    case *mg.NullableTypeReference:
        return cr.processStructStartWithType( ss, v.Type, callTyp, next )
    }
    return mg.NewTypeCastError( typ, callTyp, ss.GetPath() )
}

func ( cr *CastReactor ) processStructStart(
    ss *mgRct.StructStartEvent, next mgRct.ReactorEventProcessor ) error {

    switch v := cr.stack.Peek().( type ) {
    case mg.TypeReference:
        cr.stack.Pop()
        return cr.processStructStartWithType( ss, v, v, next )
    case valAllocCast:
        return cr.processStructStartWithType( ss, v.typ, v.typ, next )
    case *listCast:
        v.sawValues = true
        typ := v.lt.ElementType
        return cr.processStructStartWithType( ss, typ, typ, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) processListStartWithAtomicType(
    le *mgRct.ListStartEvent,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    if at.Equals( mg.TypeValue ) {
        return cr.processListStartWithType( 
            le, mg.TypeOpaqueList, callTyp, next )
    }

    return mg.NewTypeCastError( callTyp, mg.TypeOpaqueList, le.GetPath() )
}

func ( cr *CastReactor ) processListStartWithPointerType(
    le *mgRct.ListStartEvent,
    pt *mg.PointerTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    if err := cr.sendSynthAllocEvent( le.Type, le, next ); err != nil {
        return err
    }
    return cr.processListStartWithType( le, pt.Type, callTyp, next )
}

func ( cr *CastReactor ) processListStartWithListType(
    le *mgRct.ListStartEvent,
    lt *mg.ListTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {
    
    if ! le.Type.Equals( lt ) {
        if le.Id != mg.PointerIdNull {
            if cr.CastingList != nil {
                if err := cr.CastingList( le, lt ); err != nil { return err }
            }
        }
    }
    sp := objpath.CopyOf( le.GetPath() )
    cr.stack.Push( &listCast{ lt: lt, startPath: sp } )
    return next.ProcessEvent( le )
}

func ( cr *CastReactor ) processListStartWithType(
    le *mgRct.ListStartEvent,
    typ mg.TypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference:
        return cr.processListStartWithAtomicType( le, v, callTyp, next )
    case *mg.PointerTypeReference:
        return cr.processListStartWithPointerType( le, v, callTyp, next )
    case *mg.ListTypeReference:
        return cr.processListStartWithListType( le, v, callTyp, next )
    case *mg.NullableTypeReference:
        return cr.processListStartWithType( le, v.Type, callTyp, next )
    }
    panic( libErrorf( "unhandled type: %T", typ ) )
}

func ( cr *CastReactor ) processListStart( 
    le *mgRct.ListStartEvent, next mgRct.ReactorEventProcessor ) error {

    switch v := cr.stack.Peek().( type ) {
    case mg.TypeReference:
        cr.stack.Pop()
        return cr.processListStartWithType( le, v, v, next )
    case valAllocCast: 
        return cr.processListStartWithType( le, v.typ, v.typ, next )
    case *listCast:
        v.sawValues = true
        return cr.processListStartWithType( le, v.lt.ElementType, v.lt, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) ProcessEvent(
    ev mgRct.ReactorEvent, next mgRct.ReactorEventProcessor ) ( err error ) {

//    cr.dumpStack( "entering ProcessEvent()" )
//    defer cr.dumpStack( "after ProcessEvent()" )
    switch v := ev.( type ) {
    case *mgRct.ValueEvent: return cr.processValue( v, next )
    case *mgRct.ValueAllocationEvent: 
        return cr.processValueAllocation( v, next )
    case *mgRct.ValueReferenceEvent: return cr.processValueReference( v, next )
    case *mgRct.MapStartEvent: return cr.processMapStart( v, next )
    case *mgRct.FieldStartEvent: return cr.processFieldStart( v, next )
    case *mgRct.StructStartEvent: return cr.processStructStart( v, next )
    case *mgRct.ListStartEvent: return cr.processListStart( v, next )
    case *mgRct.EndEvent: return cr.processEnd( v, next )
    }
    panic( libErrorf( "unhandled event: %T", ev ) )
}
