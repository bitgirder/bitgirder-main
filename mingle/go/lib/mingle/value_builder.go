package mingle

import (
    "bitgirder/stack"
//    "log"
)

type accImpl interface {
    valueReady( val Value ) bool
    value() Value
}

type valPtrAcc struct {
    id PointerId
    val Value
}

func newValPtrAcc( id PointerId ) *valPtrAcc { return &valPtrAcc{ id: id } }

func ( vp *valPtrAcc ) valueReady( val Value ) bool {
    vp.val = val
    return true
}

func ( vp *valPtrAcc ) value() Value { return NewValuePointer( vp.val ) }

type mapAcc struct {
    arr []interface{} // alternating key, val to be passed to MustSymbolMap
}

func newMapAcc() *mapAcc {
    return &mapAcc{ arr: make( []interface{}, 0, 8 ) }
}

func ( ma *mapAcc ) value() Value { return MustSymbolMap( ma.arr... ) }

func ( ma *mapAcc ) startField( fld *Identifier ) {
    ma.arr = append( ma.arr, fld )
}

func ( ma *mapAcc ) valueReady( mv Value ) bool { 
    ma.arr = append( ma.arr, mv ) 
    return false
}

type structAcc struct {
    typ *QualifiedTypeName
    flds *mapAcc
}

func newStructAcc( typ *QualifiedTypeName ) *structAcc {
    return &structAcc{ typ: typ, flds: newMapAcc() }
}

func ( sa *structAcc ) value() Value {
    flds := sa.flds.value().( *SymbolMap )
    return &Struct{ Type: sa.typ, Fields: flds }
}

func ( sa *structAcc ) valueReady( mv Value ) bool { 
    return sa.flds.valueReady( mv ) 
}

type listAcc struct {
    vals []Value
}

func newListAcc() *listAcc {
    return &listAcc{ make( []Value, 0, 4 ) }
}

func ( la *listAcc ) value() Value { return NewList( la.vals ) }

func ( la *listAcc ) valueReady( mv Value ) bool {
    la.vals = append( la.vals, mv )
    return false
}

// Can make this public if needed
type valueAccumulator struct {
    val Value
    accs *stack.Stack
    refs map[ PointerId ] Value
}

func newValueAccumulator() *valueAccumulator {
    return &valueAccumulator{ 
        accs: stack.NewStack(), 
        refs: make( map[ PointerId ] Value ),
    }
}

func ( va *valueAccumulator ) pushAcc( acc accImpl ) {
    va.accs.Push( acc )
}

func ( va *valueAccumulator ) peekAcc() ( accImpl, bool ) {
    if va.accs.IsEmpty() { return nil, false }
    return va.accs.Peek().( accImpl ), true
}

func ( va *valueAccumulator ) popAcc() accImpl {
    return va.accs.Pop().( accImpl )
}

func ( va *valueAccumulator ) popAccValue() {
    acc := va.popAcc()
    val := acc.value()
    if ptrAcc, ok := acc.( *valPtrAcc ); ok { va.refs[ ptrAcc.id ] = val }
    va.valueReady( val )
}

func ( va *valueAccumulator ) valueReady( val Value ) {
    if acc, ok := va.peekAcc(); ok {
        if acc.valueReady( val ) { va.popAccValue() }
    } else { va.val = val }
}

// Panics if result of val is not ready
func ( va *valueAccumulator ) getValue() Value {
    if va.val == nil { panic( rctErrorf( "Value is not yet built" ) ) }
    return va.val
}

func ( va *valueAccumulator ) startField( fld *Identifier ) {
    acc, ok := va.peekAcc()
    if ok {
        var ma *mapAcc
        switch v := acc.( type ) {
        case *mapAcc: ma, ok = v, true
        case *structAcc: ma, ok = v.flds, true
        default: ok = false
        }
        if ok { ma.startField( fld ) }
    }
}

func ( va *valueAccumulator ) end() { va.popAccValue() }

func ( va *valueAccumulator ) pointerReferenced(
    vr *ValuePointerReferenceEvent ) error {

    if val, ok := va.refs[ vr.Id ]; ok {
        va.valueReady( val )
        return nil
    }
    return rctErrorf( "unhandled value pointer ref: %d", vr.Id )
}

func ( va *valueAccumulator ) ProcessEvent( ev ReactorEvent ) error {
    switch v := ev.( type ) {
    case *ValueEvent: va.valueReady( v.Val )
    case *ListStartEvent: va.pushAcc( newListAcc() )
    case *MapStartEvent: va.pushAcc( newMapAcc() )
    case *StructStartEvent: va.pushAcc( newStructAcc( v.Type ) )
    case *FieldStartEvent: va.startField( v.Field )
    case *EndEvent: va.end()
    case *ValuePointerAllocEvent: va.pushAcc( newValPtrAcc( v.Id ) )
    case *ValuePointerReferenceEvent: return va.pointerReferenced( v )
    default: panic( libErrorf( "Unhandled event: %T", ev ) )
    }
    return nil
}

type ValueBuilder struct {
    acc *valueAccumulator
}

func NewValueBuilder() *ValueBuilder {
    return &ValueBuilder{ acc: newValueAccumulator() }
}

func ( vb *ValueBuilder ) GetValue() Value { return vb.acc.getValue() }

func ( vb *ValueBuilder ) ProcessEvent( ev ReactorEvent ) error {
    if err := vb.acc.ProcessEvent( ev ); err != nil { return err }
    return nil
}
