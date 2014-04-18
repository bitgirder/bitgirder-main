package mingle

import (
    "bitgirder/stack"
    "bitgirder/pipeline"
//    "log"
)

type valPtrAcc struct {
    id PointerId
    val *HeapValue
}

type mapAccPair struct { 
    fld *Identifier
    val Value
}

type mapAcc struct { pairs []mapAccPair }

func newMapAcc() *mapAcc { return &mapAcc{ []mapAccPair{} } }

func ( ma *mapAcc ) startField( fld *Identifier ) {
    ma.pairs = append( ma.pairs, mapAccPair{ fld: fld } )
}

func ( ma *mapAcc ) setValue( val Value ) {
    pairPtr := &( ma.pairs[ len( ma.pairs ) - 1 ] )
    pairPtr.val = val
}

func ( ma *mapAcc ) makeMap() *SymbolMap {
    res := NewSymbolMap()
    for _, pair := range ma.pairs { res.Put( pair.fld, pair.val ) }
    return res
}

type structAcc struct {
    typ *QualifiedTypeName
    flds *mapAcc
}

type listAcc struct { l *List }

func newListAcc() *listAcc { return &listAcc{ NewList() } }

// Can make this public if needed
type valueAccumulator struct {
    val Value
    accs *stack.Stack
    refs map[ PointerId ] *valPtrAcc
}

func newValueAccumulator() *valueAccumulator {
    return &valueAccumulator{ 
        accs: stack.NewStack(), 
        refs: make( map[ PointerId ] *valPtrAcc ),
    }
}

func ( va *valueAccumulator ) pushAcc( acc interface{} ) { va.accs.Push( acc ) }

func ( va *valueAccumulator ) peekAcc() ( interface{}, bool ) {
    return va.accs.Peek(), ! va.accs.IsEmpty()
}

func ( va *valueAccumulator ) acceptValue( acc interface{}, val Value ) bool {
    switch v := acc.( type ) {
    case *valPtrAcc: v.val = NewHeapValue( val ); return true
    case *mapAcc: v.setValue( val ); return false
    case *structAcc: v.flds.setValue( val ); return false
    case *listAcc: v.l.Add( val ); return false
    }
    panic( libErrorf( "unhandled acc: %T", acc ) )
}

func ( va *valueAccumulator ) valueForAcc( acc interface{} ) Value {
    switch v := acc.( type ) {
    case *valPtrAcc: return v.val
    case *mapAcc: return v.makeMap()
    case *structAcc: return &Struct{ Type: v.typ, Fields: v.flds.makeMap() }
    case *listAcc: return v.l
    }
    panic( libErrorf( "unhandled acc: %T", acc ) )
}

func ( va *valueAccumulator ) popAccValue() {
    va.valueReady( va.valueForAcc( va.accs.Pop() ) )
}

func ( va *valueAccumulator ) valueReady( val Value ) {
    if acc, ok := va.peekAcc(); ok {
        if va.acceptValue( acc, val ) { va.popAccValue() }
    } else {
        va.val = val
    }
}

// Panics if result of val is not ready
func ( va *valueAccumulator ) getValue() Value {
    if va.val == nil { panic( rctErrorf( nil, "Value is not yet built" ) ) }
    return va.val
}

func ( va *valueAccumulator ) startField( fld *Identifier ) {
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

func ( va *valueAccumulator ) allocValue( id PointerId ) {
    acc := &valPtrAcc{ id: id }
    va.refs[ id ] = acc
    va.pushAcc( acc )
}

func ( va *valueAccumulator ) valueReferenced( vr *ValueReferenceEvent ) error {
    if valPtrAcc, ok := va.refs[ vr.Id ]; ok {
        va.valueReady( valPtrAcc.val )
        return nil
    }
    return rctErrorf( vr.GetPath(), "unhandled value pointer ref: %s", vr.Id )
}

func ( va *valueAccumulator ) ProcessEvent( ev ReactorEvent ) error {
    switch v := ev.( type ) {
    case *ValueEvent: va.valueReady( v.Val )
    case *ListStartEvent: va.pushAcc( newListAcc() )
    case *MapStartEvent: va.pushAcc( newMapAcc() )
    case *StructStartEvent: 
        va.pushAcc( &structAcc{ typ: v.Type, flds: newMapAcc() } )
    case *FieldStartEvent: va.startField( v.Field )
    case *EndEvent: va.end()
    case *ValueAllocationEvent: va.allocValue( v.Id )
    case *ValueReferenceEvent: va.valueReferenced( v )
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

func ( vb *ValueBuilder ) InitPipeline( p *pipeline.Pipeline ) {
    EnsureStructuralReactor( p )
}

func ( vb *ValueBuilder ) GetValue() Value { return vb.acc.getValue() }

func ( vb *ValueBuilder ) ProcessEvent( ev ReactorEvent ) error {
    if err := vb.acc.ProcessEvent( ev ); err != nil { return err }
    return nil
}
