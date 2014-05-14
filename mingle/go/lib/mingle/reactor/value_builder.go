package reactor

import (
    mg "mingle"
    "bitgirder/stack"
    "bitgirder/pipeline"
//    "log"
)

type ValueBuilder struct {
    acc *valueAccumulator
}

func NewValueBuilder() *ValueBuilder {
    return &ValueBuilder{ acc: newValueAccumulator() }
}

func ( vb *ValueBuilder ) InitializePipeline( p *pipeline.Pipeline ) {
    EnsureStructuralReactor( p )
    EnsurePointerCheckReactor( p )
}

func ( vb *ValueBuilder ) GetValue() mg.Value { return vb.acc.getValue() }

type valPtrAcc struct {
    id mg.PointerId
    val *mg.HeapValue
}

func ( vp *valPtrAcc ) mustValue() *mg.HeapValue {
    if vp.val == nil {
        panic( libErrorf( "no val for ptr acc with id %s", vp.id ) )
    }
    return vp.val
}

type mapAcc struct { 
    id mg.PointerId
    m *mg.SymbolMap
    curFld *mg.Identifier
}

func newMapAcc() *mapAcc { return &mapAcc{ m: mg.NewSymbolMap() } }

func newMapAccWithId( id mg.PointerId ) *mapAcc {
    res := newMapAcc()
    res.id = id
    return res
}

// asserts that ma.curFld is non-nil, clears it, and returns it
func ( ma *mapAcc ) clearField() *mg.Identifier {
    if ma.curFld == nil { panic( libError( "no current field" ) ) }
    res := ma.curFld
    ma.curFld = nil
    return res
}

func ( ma *mapAcc ) setValue( val mg.Value ) { ma.m.Put( ma.clearField(), val ) }

func ( ma *mapAcc ) startField( fld *mg.Identifier ) {
    if cur := ma.curFld; cur != nil {
        panic( libErrorf( "saw start of field %s while cur is %s", fld, cur ) )
    }
    ma.curFld = fld
}

type structAcc struct {
    flds *mapAcc
    s *mg.Struct
}

func newStructAcc( typ *mg.QualifiedTypeName ) *structAcc {
    res := &structAcc{ flds: newMapAcc() }
    res.s = &mg.Struct{ Type: typ, Fields: res.flds.m }
    return res
}    

type listAcc struct { 
    id mg.PointerId
    l *mg.List 
}

func newListAcc( id mg.PointerId ) *listAcc { 
    return &listAcc{ id: id, l: mg.NewList( mg.TypeOpaqueList ) } 
}

// Can make this public if needed
type valueAccumulator struct {
    val mg.Value
    accs *stack.Stack
    refs map[ mg.PointerId ] interface{}
}

func newValueAccumulator() *valueAccumulator {
    return &valueAccumulator{ 
        accs: stack.NewStack(), 
        refs: make( map[ mg.PointerId ] interface{} ),
    }
}

func ( va *valueAccumulator ) pushAcc( acc interface{} ) { va.accs.Push( acc ) }

func ( va *valueAccumulator ) idForAcc( acc interface{} ) mg.PointerId {
    switch v := acc.( type ) {
    case *valPtrAcc: return v.id
    case *listAcc: return v.id
    case *mapAcc: return v.id
    }
    panic( libErrorf( "not a addressable acc: %T", acc ) )
}

func ( va *valueAccumulator ) pushReferenceAcc( acc interface{} ) {
    if id := va.idForAcc( acc ); id != mg.PointerIdNull { va.refs[ id ] = acc }
    va.pushAcc( acc )
}

func ( va *valueAccumulator ) peekAcc() ( interface{}, bool ) {
    return va.accs.Peek(), ! va.accs.IsEmpty()
}

func ( va *valueAccumulator ) mustPeekAcc() interface{} {
    if res, ok := va.peekAcc(); ok { return res }
    panic( libError( "acc stack is empty" ) )
}

func ( va *valueAccumulator ) acceptValue( acc interface{}, val mg.Value ) bool {
    switch v := acc.( type ) {
    case *valPtrAcc: v.val = mg.NewHeapValue( val ); return true
    case *mapAcc: v.setValue( val ); return false
    case *structAcc: v.flds.setValue( val ); return false
    case *listAcc: v.l.Add( val ); return false
    }
    panic( libErrorf( "unhandled acc: %T", acc ) )
}

func ( va *valueAccumulator ) valueForAcc( acc interface{} ) mg.Value {
    switch v := acc.( type ) {
    case *valPtrAcc: return v.mustValue()
    case *mapAcc: return v.m
    case *structAcc: return v.s
    case *listAcc: return v.l
    }
    panic( libErrorf( "unhandled acc: %T", acc ) )
}

func ( va *valueAccumulator ) popAccValue() {
    va.valueReady( va.valueForAcc( va.accs.Pop() ) )
}

func ( va *valueAccumulator ) valueReady( val mg.Value ) {
    if acc, ok := va.peekAcc(); ok {
        if va.acceptValue( acc, val ) { va.popAccValue() }
    } else {
        va.val = val
    }
}

// Panics if result of val is not ready
func ( va *valueAccumulator ) getValue() mg.Value {
    if va.val == nil { panic( rctErrorf( nil, "Value is not yet built" ) ) }
    return va.val
}

func ( va *valueAccumulator ) startField( fld *mg.Identifier ) {
    acc, ok := va.peekAcc()
    if ! ok { panic( libErrorf( "got field start %s with empty stack", fld ) ) }
    switch v := acc.( type ) {
    case *mapAcc: v.startField( fld )
    case *structAcc: v.flds.startField( fld )
    default:
        panic( libErrorf( "unexpected acc for start of field %s: %T", fld, v ) )
    }
}

func ( va *valueAccumulator ) end() { va.popAccValue() }

func ( va *valueAccumulator ) valueReferenced( vr *ValueReferenceEvent ) error {
    if acc, ok := va.refs[ vr.Id ]; ok {
        va.valueReady( va.valueForAcc( acc ) ) 
        return nil
    }
    return rctErrorf( vr.GetPath(), "unhandled value pointer ref: %s", vr.Id )
}

func ( va *valueAccumulator ) ProcessEvent( ev ReactorEvent ) error {
    switch v := ev.( type ) {
    case *ValueEvent: va.valueReady( v.Val )
    case *ListStartEvent: va.pushReferenceAcc( newListAcc( v.Id ) )
    case *MapStartEvent: va.pushReferenceAcc( newMapAccWithId( v.Id ) )
    case *StructStartEvent: va.pushAcc( newStructAcc( v.Type ) )
    case *FieldStartEvent: va.startField( v.Field )
    case *EndEvent: va.end()
    case *ValueAllocationEvent: va.pushReferenceAcc( &valPtrAcc{ id: v.Id } )
    case *ValueReferenceEvent: return va.valueReferenced( v )
    default: panic( libErrorf( "Unhandled event: %T", ev ) )
    }
    return nil
}

func ( vb *ValueBuilder ) ProcessEvent( ev ReactorEvent ) error {
    return vb.acc.ProcessEvent( ev )
}
