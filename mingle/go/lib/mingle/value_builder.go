package mingle

import (
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

func ( vb *ValueBuilder ) InitPipeline( p *pipeline.Pipeline ) {
    EnsureStructuralReactor( p )
}

func ( vb *ValueBuilder ) GetValue() Value { return vb.acc.getValue() }

type valPtrAcc struct {
    id PointerId
    val *HeapValue
}

func ( vp *valPtrAcc ) mustValue() *HeapValue {
    if vp.val == nil {
        panic( libErrorf( "no val for ptr acc with id %s", vp.id ) )
    }
    return vp.val
}

type mapAcc struct { 
    m *SymbolMap
    curFld *Identifier
}

func newMapAcc() *mapAcc { return &mapAcc{ m: NewSymbolMap() } }

// asserts that ma.curFld is non-nil, clears it, and returns it
func ( ma *mapAcc ) clearField() *Identifier {
    if ma.curFld == nil { panic( libError( "no current field" ) ) }
    res := ma.curFld
    ma.curFld = nil
    return res
}

func ( ma *mapAcc ) setValue( val Value ) { ma.m.Put( ma.clearField(), val ) }

func ( ma *mapAcc ) startField( fld *Identifier ) {
    if cur := ma.curFld; cur != nil {
        panic( libErrorf( "saw start of field %s while cur is %s", fld, cur ) )
    }
    ma.curFld = fld
}

type structAcc struct {
    flds *mapAcc
    s *Struct
}

func newStructAcc( typ *QualifiedTypeName ) *structAcc {
    res := &structAcc{ flds: newMapAcc() }
    res.s = &Struct{ Type: typ, Fields: res.flds.m }
    return res
}    

type listAcc struct { 
    id PointerId
    l *List 
}

func newListAcc( id PointerId ) *listAcc { 
    return &listAcc{ id: id, l: NewList() } 
}

type valueBuildResolution struct {
    id PointerId
    f func( Value )
}

// Can make this public if needed
type valueAccumulator struct {
    val Value
    accs *stack.Stack
    refs map[ PointerId ] interface{}
    resolutions []valueBuildResolution
}

func newValueAccumulator() *valueAccumulator {
    return &valueAccumulator{ 
        accs: stack.NewStack(), 
        refs: make( map[ PointerId ] interface{} ),
        resolutions: make( []valueBuildResolution, 0, 4 ),
    }
}

func ( va *valueAccumulator ) addResolver( id PointerId, f func( Value ) ) {
    res := valueBuildResolution{ id: id, f: f }
    va.resolutions = append( va.resolutions, res )
}

func ( va *valueAccumulator ) pushAcc( acc interface{} ) { va.accs.Push( acc ) }

func ( va *valueAccumulator ) idForAcc( acc interface{} ) PointerId {
    switch v := acc.( type ) {
    case *valPtrAcc: return v.id
    case *listAcc: return v.id
    }
    panic( libErrorf( "not a addressable acc: %T", acc ) )
}

func ( va *valueAccumulator ) pushReferenceAcc( acc interface{} ) {
    va.refs[ va.idForAcc( acc ) ] = acc
    va.pushAcc( acc )
}

func ( va *valueAccumulator ) peekAcc() ( interface{}, bool ) {
    return va.accs.Peek(), ! va.accs.IsEmpty()
}

func ( va *valueAccumulator ) mustPeekAcc() interface{} {
    if res, ok := va.peekAcc(); ok { return res }
    panic( libError( "acc stack is empty" ) )
}

func ( va *valueAccumulator ) setForwardFieldValue( ma *mapAcc, id PointerId ) {
    fld := ma.clearField()
    va.addResolver( id, func( val Value ) { ma.m.Put( fld, val ) } )
}

func ( va *valueAccumulator ) forwardValueReferenced( id PointerId ) {
    switch v := va.mustPeekAcc().( type ) {
    case *structAcc: va.setForwardFieldValue( v.flds, id )
    case *mapAcc: va.setForwardFieldValue( v, id )
    default: panic( libErrorf( "unhandled acc: %T", v ) )
    }
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

func ( va *valueAccumulator ) resolve() {
    for _, rs := range va.resolutions {
        acc := va.refs[ rs.id ]
        rs.f( va.valueForAcc( acc ) )
    }
}

func ( va *valueAccumulator ) valueReady( val Value ) {
    if acc, ok := va.peekAcc(); ok {
        if va.acceptValue( acc, val ) { va.popAccValue() }
    } else {
        va.resolve()
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

func ( va *valueAccumulator ) isForwardRef( acc interface{} ) bool {
    if vp, ok := acc.( *valPtrAcc ); ok { return vp.val == nil }
    return false
}

func ( va *valueAccumulator ) valueReferenced( vr *ValueReferenceEvent ) error {
    if acc, ok := va.refs[ vr.Id ]; ok {
        if va.isForwardRef( acc ) {
            va.forwardValueReferenced( vr.Id )
        } else { 
            va.valueReady( va.valueForAcc( acc ) ) 
        }
        return nil
    }
    return rctErrorf( vr.GetPath(), "unhandled value pointer ref: %s", vr.Id )
}

func ( va *valueAccumulator ) ProcessEvent( ev ReactorEvent ) error {
    switch v := ev.( type ) {
    case *ValueEvent: va.valueReady( v.Val )
    case *ListStartEvent: va.pushReferenceAcc( newListAcc( v.Id ) )
    case *MapStartEvent: va.pushAcc( newMapAcc() )
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
