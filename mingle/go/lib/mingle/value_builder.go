package mingle

import (
    "bitgirder/stack"
//    "log"
)

type accImpl interface {
    valueReady( val Value ) 
    end() ( Value, error )
}

type mapAcc struct {
    arr []interface{} // alternating key, val to be passed to MustSymbolMap
}

func newMapAcc() *mapAcc {
    return &mapAcc{ arr: make( []interface{}, 0, 8 ) }
}

func ( ma *mapAcc ) end() ( Value, error ) { 
    res, err := CreateSymbolMap( ma.arr... )
    if err == nil { return res, nil } 
    return nil, rctErrorf( "Invalid fields: %s", err.Error() )
}

func ( ma *mapAcc ) startField( fld *Identifier ) {
    ma.arr = append( ma.arr, fld )
}

func ( ma *mapAcc ) valueReady( mv Value ) { ma.arr = append( ma.arr, mv ) }

type structAcc struct {
    typ *QualifiedTypeName
    flds *mapAcc
}

func newStructAcc( typ *QualifiedTypeName ) *structAcc {
    return &structAcc{ typ: typ, flds: newMapAcc() }
}

func ( sa *structAcc ) end() ( Value, error ) {
    flds, err := sa.flds.end()
    if err != nil { return nil, err }
    return &Struct{ Type: sa.typ, Fields: flds.( *SymbolMap ) }, nil
}

func ( sa *structAcc ) valueReady( mv Value ) { sa.flds.valueReady( mv ) }

type listAcc struct {
    vals []Value
}

func newListAcc() *listAcc {
    return &listAcc{ make( []Value, 0, 4 ) }
}

func ( la *listAcc ) end() ( Value, error ) { 
    return NewList( la.vals ), nil
}

func ( la *listAcc ) valueReady( mv Value ) {
    la.vals = append( la.vals, mv )
}

// Can make this public if needed
type valueAccumulator struct {
    val Value
    accs *stack.Stack
}

func newValueAccumulator() *valueAccumulator {
    return &valueAccumulator{ accs: stack.NewStack() }
}

func ( va *valueAccumulator ) pushAcc( acc accImpl ) {
    va.accs.Push( acc )
}

func ( va *valueAccumulator ) peekAcc() ( accImpl, bool ) {
    if va.accs.IsEmpty() { return nil, false }
    return va.accs.Peek().( accImpl ), true
}

func ( va *valueAccumulator ) popAcc() accImpl {
    res, ok := va.peekAcc()
    if ! ok { panic( libErrorf( "popAcc() called on empty stack" ) ) }
    va.accs.Pop()
    return res
}

func ( va *valueAccumulator ) valueReady( val Value ) {
    if acc, ok := va.peekAcc(); ok {
        acc.valueReady( val )
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

func ( va *valueAccumulator ) end() error {
    acc := va.popAcc()
    if val, err := acc.end(); err == nil {
        va.valueReady( val )
    } else { return err }
    return nil
}

func ( va *valueAccumulator ) ProcessEvent( ev ReactorEvent ) error {
    switch v := ev.( type ) {
    case *ValueEvent: va.valueReady( v.Val )
    case *ListStartEvent: va.pushAcc( newListAcc() )
    case *MapStartEvent: va.pushAcc( newMapAcc() )
    case *StructStartEvent: va.pushAcc( newStructAcc( v.Type ) )
    case *FieldStartEvent: va.startField( v.Field )
    case *EndEvent: if err := va.end(); err != nil { return err }
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
