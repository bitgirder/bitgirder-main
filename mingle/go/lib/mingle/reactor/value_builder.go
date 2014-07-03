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
}

func ( vb *ValueBuilder ) GetValue() mg.Value { return vb.acc.getValue() }

type mapAcc struct { 
    m *mg.SymbolMap
    curFld *mg.Identifier
}

func newMapAcc() *mapAcc { return &mapAcc{ m: mg.NewSymbolMap() } }

// asserts that ma.curFld is non-nil, clears it, and returns it
func ( ma *mapAcc ) clearField() *mg.Identifier {
    if ma.curFld == nil { panic( libError( "no current field" ) ) }
    res := ma.curFld
    ma.curFld = nil
    return res
}

func ( ma *mapAcc ) setValue( val mg.Value ) { 
    ma.m.Put( ma.clearField(), val ) 
}

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
    l *mg.List 
}

func newListAcc() *listAcc { 
    return &listAcc{ l: mg.NewList( mg.TypeOpaqueList ) } 
}

// Can make this public if needed
type valueAccumulator struct {
    val mg.Value
    accs *stack.Stack
}

func newValueAccumulator() *valueAccumulator {
    return &valueAccumulator{ accs: stack.NewStack() }
}

func ( va *valueAccumulator ) pushAcc( acc interface{} ) { va.accs.Push( acc ) }

func ( va *valueAccumulator ) peekAcc() ( interface{}, bool ) {
    return va.accs.Peek(), ! va.accs.IsEmpty()
}

func ( va *valueAccumulator ) mustPeekAcc() interface{} {
    if res, ok := va.peekAcc(); ok { return res }
    panic( libError( "acc stack is empty" ) )
}

func ( va *valueAccumulator ) acceptValue( 
    acc interface{}, val mg.Value ) bool {

    switch v := acc.( type ) {
    case *mapAcc: v.setValue( val ); return false
    case *structAcc: v.flds.setValue( val ); return false
    case *listAcc: v.l.AddUnsafe( val ); return false
    }
    panic( libErrorf( "unhandled acc: %T", acc ) )
}

func ( va *valueAccumulator ) valueForAcc( acc interface{} ) mg.Value {
    switch v := acc.( type ) {
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

func ( va *valueAccumulator ) ProcessEvent( ev ReactorEvent ) error {
    switch v := ev.( type ) {
    case *ValueEvent: va.valueReady( v.Val )
    case *ListStartEvent: va.pushAcc( newListAcc() )
    case *MapStartEvent: va.pushAcc( newMapAcc() )
    case *StructStartEvent: va.pushAcc( newStructAcc( v.Type ) )
    case *FieldStartEvent: va.startField( v.Field )
    case *EndEvent: va.end()
    default: panic( libErrorf( "Unhandled event: %T", ev ) )
    }
    return nil
}

func ( vb *ValueBuilder ) ProcessEvent( ev ReactorEvent ) error {
    return vb.acc.ProcessEvent( ev )
}
