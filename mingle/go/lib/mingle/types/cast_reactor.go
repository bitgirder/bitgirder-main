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

type FieldTyper interface {

    // path will be positioned to the map/struct containing fld, but will not
    // itself include fld
    FieldTypeFor( 
        fld *mg.Identifier, path objpath.PathNode ) ( mg.TypeReference, error )
}

type CastInterface interface {

    // Called when start of a symbol map arrives when an atomic type having name
    // qn (or a nullable or list type containing such an atomic type) is the
    // cast reactor's expected type. Returning true from this function will
    // cause the cast reactor to treat the symbol map start as if it were a
    // struct start with atomic type qn. 
    //
    // One motivating use for this is for cast reactors that react to inputs
    // conforming to a known schema and receive unadorned maps for structured
    // field values and wish to cause further processing to behave as if the
    // struct were explicitly signalled in the input
    InferStructFor( qn *mg.QualifiedTypeName ) bool

    FieldTyperFor( 
        qn *mg.QualifiedTypeName, path objpath.PathNode ) ( FieldTyper, error )

    CastAtomic( 
        in mg.Value, 
        at *mg.AtomicTypeReference, 
        path objpath.PathNode ) ( mg.Value, error, bool )
    
    // will be called when act != expct to determine whether to continue
    // processing an input having type act
    AllowAssignment( expct, act *mg.QualifiedTypeName ) bool
}

type valueFieldTyper int

func ( vt valueFieldTyper ) FieldTypeFor( 
    fld *mg.Identifier, path objpath.PathNode ) ( mg.TypeReference, error ) {
    return mg.TypeNullableValue, nil
}

type castInterfaceDefault int

func ( i castInterfaceDefault ) FieldTyperFor( 
    qn *mg.QualifiedTypeName, path objpath.PathNode ) ( FieldTyper, error ) {
    return valueFieldTyper( 1 ), nil
}

func ( i castInterfaceDefault ) InferStructFor( 
    at *mg.QualifiedTypeName ) bool {

    return false
}

func ( i castInterfaceDefault ) CastAtomic( 
    v mg.Value, 
    at *mg.AtomicTypeReference, 
    path objpath.PathNode ) ( mg.Value, error, bool ) {

    return nil, nil, false
}

func ( i castInterfaceDefault ) AllowAssignment( 
    expct, act *mg.QualifiedTypeName ) bool {

    return false
}

type CastReactor struct {

    iface CastInterface

    stack *stack.Stack

    SkipPathSetter bool

    CastingList func( le *mgRct.ListStartEvent, lt *mg.ListTypeReference ) error

    ShouldSuppressAllocation func( typ mg.TypeReference ) bool

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

func NewCastReactor( 
    expct mg.TypeReference, iface CastInterface ) *CastReactor {

    res := &CastReactor{ stack: stack.NewStack(), iface: iface }
    res.stack.Push( expct )
    return res
}

func NewDefaultCastReactor( expct mg.TypeReference ) *CastReactor {
    return NewCastReactor( expct, castInterfaceDefault( 1 ) )
}

func ( cr *CastReactor ) InitializePipeline( pip *pipeline.Pipeline ) {
    mgRct.EnsureStructuralReactor( pip )
    if ! cr.SkipPathSetter { mgRct.EnsurePathSettingProcessor( pip ) }
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

func ( cr *CastReactor ) valueEventForAtomicCast( 
    ve *mgRct.ValueEvent, 
    at *mg.AtomicTypeReference, 
    callTyp mg.TypeReference ) ( error, *mgRct.ValueEvent ) {

    mv, err, ok := cr.iface.CastAtomic( ve.Val, at, ve.GetPath() )
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
    if cr.ShouldSuppressAllocation == nil { return true }
    return cr.ShouldSuppressAllocation( typ )
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

func ( cr *CastReactor ) completeStartStruct(
    ss *mgRct.StructStartEvent, next mgRct.ReactorEventProcessor ) error {

    ft, err := cr.iface.FieldTyperFor( ss.Type, ss.GetPath() )
    if err != nil { return err }

    if ft == nil { ft = valueFieldTyper( 1 ) }
    return cr.implMapStart( ss, ft, next )
}

func ( cr *CastReactor ) inferStructForMap(
    me *mgRct.MapStartEvent,
    at *mg.AtomicTypeReference,
    next mgRct.ReactorEventProcessor ) ( error, bool ) {

    if ! cr.iface.InferStructFor( at.Name ) { return nil, false }

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

func ( cr *CastReactor ) processStructStartWithAtomicType(
    ss *mgRct.StructStartEvent,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    if at.Equals( mg.TypeSymbolMap ) {
        me := mgRct.NewMapStartEvent( mg.PointerIdNull )
        me.SetPath( ss.GetPath() )
        return cr.processMapStartWithAtomicType( me, at, callTyp, next )
    }

    if at.Name.Equals( ss.Type ) || at.Equals( mg.TypeValue ) ||
       cr.iface.AllowAssignment( at.Name, ss.Type ) {
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
    ev mgRct.ReactorEvent, next mgRct.ReactorEventProcessor ) error {

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
