package mingle

import (
    "bitgirder/stack"
    "log"
)

type valPtrAcc struct {
    id PointerId
    val interface{}
    valPtr *ValuePointer
}

type mapAccPair struct { 
    fld *Identifier
    val interface{}
}

type mapAcc struct { pairs []mapAccPair }

func newMapAcc() *mapAcc { return &mapAcc{ []mapAccPair{} } }

func ( ma *mapAcc ) startField( fld *Identifier ) {
    ma.pairs = append( ma.pairs, mapAccPair{ fld: fld } )
}

func ( ma *mapAcc ) setValue( val interface{} ) {
    pair := &( ma.pairs[ len( ma.pairs ) - 1 ] )
    pair.val = val
}

type structAcc struct {
    typ *QualifiedTypeName
    flds *mapAcc
}

type listAcc struct { vals []interface{} }

func newListAcc() *listAcc { return &listAcc{ []interface{}{} } }

type valueAccFwdRefResolution struct {
    f func( Value )
    ref *valPtrAcc 
}

// Can make this public if needed
type valueAccumulator struct {
    val Value
    accs *stack.Stack
    resolvers []valueAccFwdRefResolution
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

func ( va *valueAccumulator ) acceptValue( acc, val interface{} ) bool {
    log.Printf( "%T accepting %T", acc, val )
    switch v := acc.( type ) {
    case *valPtrAcc: v.val = val; return true
    case *mapAcc: v.setValue( val ); return false
    case *structAcc: v.flds.setValue( val ); return false
    case *listAcc: v.vals = append( v.vals, val ); return false
    }
    panic( libErrorf( "unhandled acc: %T", acc ) )
}

func ( va *valueAccumulator ) resolveRef( 
    ref *valPtrAcc, f func( val Value ) ) {

    res := valueAccFwdRefResolution{ f: f, ref: ref }
    va.resolvers = append( va.resolvers, res )
}

func ( va *valueAccumulator ) valueOrFwdRef( 
    val interface{} ) ( Value, *valPtrAcc ) {

    switch v := val.( type ) {
    case Value: return v, nil
    case *valPtrAcc: return nil, v
    }
    panic( libErrorf( "not a value or a forward ref: %T", val ) )
}

func ( va *valueAccumulator ) valueForValuePtrAcc( pa *valPtrAcc ) Value {
    val, fwdRef := va.valueOrFwdRef( pa.val )
    pa.valPtr = NewValuePointer( val )
    if fwdRef != nil {
        va.resolveRef( fwdRef, func( v Value ) { pa.valPtr.Val = v } )
    }
    return pa.valPtr
}

func ( va *valueAccumulator ) valueForMapAcc( ma *mapAcc ) *SymbolMap {
    res := NewSymbolMap()
    for _, pair := range ma.pairs {
        val, ref := va.valueOrFwdRef( pair.val )
        if ref == nil {
            res.Put( pair.fld, val )
        } else {
            va.resolveRef( ref, func( v Value ) { res.Put( pair.fld, v ) } )
        }
    }
    return res
}

func ( va *valueAccumulator ) valueForListAcc( la *listAcc ) Value {
    res := MakeList( len( la.vals ) )
    for i, elt := range la.vals {
        eltVal, eltRef := va.valueOrFwdRef( elt )
        if eltRef == nil {
            res.Set( eltVal, i )
        } else {
            var idxCapture = i
            va.resolveRef( eltRef, func( v Value ) { 
                log.Printf( "setting %v for res[ %d ]", v, idxCapture )
                res.Set( v, idxCapture ) 
            })
        }
    }
    return res
}    

func ( va *valueAccumulator ) valueForAcc( acc interface{} ) Value {
    switch v := acc.( type ) {
    case *valPtrAcc: return va.valueForValuePtrAcc( v )
    case *mapAcc: return va.valueForMapAcc( v )
    case *structAcc:
        return &Struct{ Type: v.typ, Fields: va.valueForMapAcc( v.flds ) }
    case *listAcc: return va.valueForListAcc( v )
    }
    panic( libErrorf( "unhandled acc: %T", acc ) )
}

func ( va *valueAccumulator ) popAccValue() {
    va.valueReady( va.valueForAcc( va.accs.Pop() ) )
}

func ( va *valueAccumulator ) resolveFwdRefs() {
    for _, rslv := range va.resolvers { 
        log.Printf( "apply resolver func to ref: %#v", rslv.ref )
        rslv.f( rslv.ref.valPtr ) 
    }
}

func ( va *valueAccumulator ) valueReady( val interface{} ) {
    if acc, ok := va.peekAcc(); ok {
        if va.acceptValue( acc, val ) { va.popAccValue() }
    } else {
        va.val = val.( Value )
        va.resolveFwdRefs()
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

func ( va *valueAccumulator ) pointerAlloced( id PointerId ) {
    acc := &valPtrAcc{ id: id }
    va.refs[ id ] = acc
    va.pushAcc( acc )
}

func ( va *valueAccumulator ) pointerReferenced(
    vr *ValuePointerReferenceEvent ) error {

    if valPtrAcc, ok := va.refs[ vr.Id ]; ok {
        va.valueReady( valPtrAcc )
        return nil
    }
    return rctErrorf( vr.GetPath(), "unhandled value pointer ref: %d", vr.Id )
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
    case *ValuePointerAllocEvent: va.pointerAlloced( v.Id )
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
