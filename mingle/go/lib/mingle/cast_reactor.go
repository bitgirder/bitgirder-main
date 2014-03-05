package mingle

import (
    "bitgirder/objpath"
    "bitgirder/pipeline"
    "bitgirder/stack"
//    "log"
)

type FieldTyper interface {

    // path will be positioned to the map/struct containing fld, but will not
    // itself include fld
    FieldTypeFor( 
        fld *Identifier, path objpath.PathNode ) ( TypeReference, error )
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
    InferStructFor( qn *QualifiedTypeName ) bool

    FieldTyperFor( 
        qn *QualifiedTypeName, path objpath.PathNode ) ( FieldTyper, error )

    CastAtomic( 
        in Value, 
        at *AtomicTypeReference, 
        path objpath.PathNode ) ( Value, error, bool )
    
    // will be called when act != expct to determine whether to continue
    // processing an input having type act
    AllowAssignment( expct, act *QualifiedTypeName ) bool
}

type valueFieldTyper int

func ( vt valueFieldTyper ) FieldTypeFor( 
    fld *Identifier, path objpath.PathNode ) ( TypeReference, error ) {
    return TypeNullableValue, nil
}

type castInterfaceDefault int

func ( i castInterfaceDefault ) FieldTyperFor( 
    qn *QualifiedTypeName, path objpath.PathNode ) ( FieldTyper, error ) {
    return valueFieldTyper( 1 ), nil
}

func ( i castInterfaceDefault ) InferStructFor( at *QualifiedTypeName ) bool {
    return false
}

func ( i castInterfaceDefault ) CastAtomic( 
    v Value, 
    at *AtomicTypeReference, 
    path objpath.PathNode ) ( Value, error, bool ) {

    return nil, nil, false
}

func ( i castInterfaceDefault ) AllowAssignment( 
    expct, act *QualifiedTypeName ) bool {

    return false
}

type CastReactor struct {
    iface CastInterface
    stack *stack.Stack
    SkipPathSetter bool
}

func NewCastReactor( expct TypeReference, iface CastInterface ) *CastReactor {
    res := &CastReactor{ stack: stack.NewStack(), iface: iface }
    res.stack.Push( expct )
    return res
}

func NewDefaultCastReactor( expct TypeReference ) *CastReactor {
    return NewCastReactor( expct, castInterfaceDefault( 1 ) )
}

func ( cr *CastReactor ) InitializePipeline( pip *pipeline.Pipeline ) {
    EnsureStructuralReactor( pip )
    if ! cr.SkipPathSetter { EnsurePathSettingProcessor( pip ) }
}

type listCast struct {
    sawValues bool
    lt *ListTypeReference
    startPath objpath.PathNode
}

func ( cr *CastReactor ) errStackUnrecognized() error {
    return libErrorf( "unrecognized stack element: %T", cr.stack.Peek() )
}

func ( cr *CastReactor ) processAtomicValue(
    ve *ValueEvent,
    at *AtomicTypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    mv, err, ok := cr.iface.CastAtomic( ve.Val, at, ve.GetPath() )

    if ! ok { 
        mv, err = castAtomicWithCallType( ve.Val, at, callTyp, ve.GetPath() ) 
    }

    if err != nil { return err }

    ve.Val = mv
    return next.ProcessEvent( ve )
}

func ( cr *CastReactor ) processNullableValue(
    ve *ValueEvent,
    nt *NullableTypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    if _, ok := ve.Val.( *Null ); ok { return next.ProcessEvent( ve ) }
    return cr.processValueWithType( ve, nt.Type, callTyp, next )
}

func ( cr *CastReactor ) processValueWithType(
    ve *ValueEvent,
    typ TypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    switch v := typ.( type ) {
    case *AtomicTypeReference: 
        return cr.processAtomicValue( ve, v, callTyp, next )
    case *NullableTypeReference:
        return cr.processNullableValue( ve, v, callTyp, next )
    case *ListTypeReference:
        return NewTypeCastErrorValue( callTyp, ve.Val, ve.GetPath() )
    }
    panic( libErrorf( "unhandled type: %T", typ ) )
}

func ( cr *CastReactor ) processValue( 
    ve *ValueEvent, next ReactorEventProcessor ) error {

    switch v := cr.stack.Peek().( type ) {
    case TypeReference: 
        cr.stack.Pop()
        return cr.processValueWithType( ve, v, v, next )
    case *listCast:
        v.sawValues = true
        typ := v.lt.ElementType
        return cr.processValueWithType( ve, typ, typ, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) implMapStart(
    ev ReactorEvent, ft FieldTyper, next ReactorEventProcessor ) error {

    cr.stack.Push( ft )
    return next.ProcessEvent( ev )
}

func ( cr *CastReactor ) completeStartStruct(
    ss *StructStartEvent, next ReactorEventProcessor ) error {

    ft, err := cr.iface.FieldTyperFor( ss.Type, ss.GetPath() )
    if err != nil { return err }

    if ft == nil { ft = valueFieldTyper( 1 ) }
    return cr.implMapStart( ss, ft, next )
}

func ( cr *CastReactor ) inferStructForMap(
    me *MapStartEvent,
    at *AtomicTypeReference,
    next ReactorEventProcessor ) ( error, bool ) {

    qn, ok := at.Name.( *QualifiedTypeName )
    if ! ok { return nil, false }

    if ! cr.iface.InferStructFor( qn ) { return nil, false }

    ev := NewStructStartEvent( qn )
    ev.SetPath( me.GetPath() )

    return cr.completeStartStruct( ev, next ), true
}

func ( cr *CastReactor ) processMapStartWithAtomicType(
    me *MapStartEvent,
    at *AtomicTypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    if at.Equals( TypeSymbolMap ) || at.Equals( TypeValue ) {
        return cr.implMapStart( me, valueFieldTyper( 1 ), next )
    }

    if err, ok := cr.inferStructForMap( me, at, next ); ok { return err }

    return NewTypeCastError( callTyp, TypeSymbolMap, me.GetPath() )
}

func ( cr *CastReactor ) processMapStartWithType(
    me *MapStartEvent, 
    typ TypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    switch v := typ.( type ) {
    case *AtomicTypeReference:
        return cr.processMapStartWithAtomicType( me, v, callTyp, next )
    case *NullableTypeReference:
        return cr.processMapStartWithType( me, v.Type, callTyp, next )
    }
    return NewTypeCastError( callTyp, typ, me.GetPath() )
}

func ( cr *CastReactor ) processMapStart(
    me *MapStartEvent, next ReactorEventProcessor ) error {
    
    switch v := cr.stack.Peek().( type ) {
    case TypeReference: 
        cr.stack.Pop()
        return cr.processMapStartWithType( me, v, v, next )
    case *listCast:
        v.sawValues = true
        typ := v.lt.ElementType
        return cr.processMapStartWithType( me, typ, typ, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) processFieldStart(
    fs *FieldStartEvent, next ReactorEventProcessor ) error {

    ft := cr.stack.Peek().( FieldTyper )
    
    typ, err := ft.FieldTypeFor( fs.Field, fs.GetPath().Parent() )
    if err != nil { return err }

    cr.stack.Push( typ )
    return next.ProcessEvent( fs )
}

func ( cr *CastReactor ) processEnd(
    ee *EndEvent, next ReactorEventProcessor ) error {

    switch v := cr.stack.Peek().( type ) {
    case *listCast:
        cr.stack.Pop()
        if ! ( v.sawValues || v.lt.AllowsEmpty ) {
            return NewValueCastError( v.startPath, "List is empty" )
        }
    case FieldTyper: cr.stack.Pop()
    }

    return next.ProcessEvent( ee )
}

func ( cr *CastReactor ) processStructStartWithAtomicType(
    ss *StructStartEvent,
    at *AtomicTypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    if at.Equals( TypeSymbolMap ) {
        me := NewMapStartEvent()
        me.SetPath( ss.GetPath() )
        return cr.processMapStartWithAtomicType( me, at, callTyp, next )
    }

    if at.Name.Equals( ss.Type ) || at.Equals( TypeValue ) ||
       cr.iface.AllowAssignment( at.Name.( *QualifiedTypeName ), ss.Type ) {
        return cr.completeStartStruct( ss, next )
    }

    failTyp := &AtomicTypeReference{ Name: ss.Type }
    return NewTypeCastError( callTyp, failTyp, ss.GetPath() )
}

func ( cr *CastReactor ) processStructStartWithType(
    ss *StructStartEvent,
    typ TypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    switch v := typ.( type ) {
    case *AtomicTypeReference:
        return cr.processStructStartWithAtomicType( ss, v, callTyp, next )
    case *NullableTypeReference:
        return cr.processStructStartWithType( ss, v.Type, callTyp, next )
    }
    return NewTypeCastError( typ, callTyp, ss.GetPath() )
}

func ( cr *CastReactor ) processStructStart(
    ss *StructStartEvent, next ReactorEventProcessor ) error {

    switch v := cr.stack.Peek().( type ) {
    case TypeReference:
        cr.stack.Pop()
        return cr.processStructStartWithType( ss, v, v, next )
    case *listCast:
        v.sawValues = true
        typ := v.lt.ElementType
        return cr.processStructStartWithType( ss, typ, typ, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) processListStartWithAtomicType(
    le *ListStartEvent,
    at *AtomicTypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    if at.Equals( TypeValue ) {
        return cr.processListStartWithType( le, TypeOpaqueList, callTyp, next )
    }

    return NewTypeCastError( callTyp, TypeOpaqueList, le.GetPath() )
}

func ( cr *CastReactor ) processListStartWithType(
    le *ListStartEvent,
    typ TypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    switch v := typ.( type ) {
    case *AtomicTypeReference:
        return cr.processListStartWithAtomicType( le, v, callTyp, next )
    case *ListTypeReference:
        sp := objpath.CopyOf( le.GetPath() )
        cr.stack.Push( &listCast{ lt: v, startPath: sp } )
        return next.ProcessEvent( le )
    case *NullableTypeReference:
        return cr.processListStartWithType( le, v.Type, callTyp, next )
    }
    panic( libErrorf( "unhandled type: %T", typ ) )
}

func ( cr *CastReactor ) processListStart( 
    le *ListStartEvent, next ReactorEventProcessor ) error {

    switch v := cr.stack.Peek().( type ) {
    case TypeReference:
        cr.stack.Pop()
        return cr.processListStartWithType( le, v, v, next )
    case *listCast:
        v.sawValues = true
        return cr.processListStartWithType( le, v.lt.ElementType, v.lt, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) ProcessEvent(
    ev ReactorEvent, next ReactorEventProcessor ) error {

    switch v := ev.( type ) {
    case *ValueEvent: return cr.processValue( v, next )
    case *MapStartEvent: return cr.processMapStart( v, next )
    case *FieldStartEvent: return cr.processFieldStart( v, next )
    case *StructStartEvent: return cr.processStructStart( v, next )
    case *ListStartEvent: return cr.processListStart( v, next )
    case *EndEvent: return cr.processEnd( v, next )
    }
    panic( libErrorf( "unhandled event: %T", ev ) )
}
