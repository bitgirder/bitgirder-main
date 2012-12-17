package mingle

import (
    "container/list"
    "fmt"
)

// Misc design note: Reactor is meant to be composable, and the design should
// lend itself to allow reactor chaining to accomplish specific things. For
// instance, a binary or json codec may not itself track and generate errors
// upon being fed a symbol map with duplicate keys, since doing complicates its
// implementation and incurs memory/cpu cost. However, when callers want to
// strictly check this, they might wrap the reactor in another one that tracks
// field names and generates an error on a duplicate, but otherwise passes calls
// down. Other chaining examples might be reactors which check the call sequence
// for validity with some known schema or data definition.

type Reactor interface {
    Value( v Value ) error
    StartStruct( typ TypeReference ) error
    StartMap() error
    StartList() error
    StartField( fld *Identifier ) error
    End() error
}

type ReactorError struct { msg string }

func ( e *ReactorError ) Error() string { return e.msg }

func rctErrorf( tmpl string, args ...interface{} ) *ReactorError {
    return &ReactorError{ fmt.Sprintf( tmpl, args... ) }
}

func getTopLevelStructStartExpectError( valName string ) error {
    return rctErrorf( "Expected top-level struct start but got %s", valName )
}

type reactorImplMap struct {
    pending *Identifier
}

func newReactorImplMap() *reactorImplMap { return &reactorImplMap{} }

func ( m *reactorImplMap ) startField( fld *Identifier ) error {
    if m.pending == nil {
        m.pending = fld
        return nil
    }
    tmpl := "Saw start of field '%s' while expecting a value for '%s'"
    return rctErrorf( tmpl, fld, m.pending )
}

// Clears m.pending on nil return val
func ( m *reactorImplMap ) checkValue( valName string ) error {
    if m.pending == nil {
        tmpl := "Expected field name or end of fields but got %s"
        return rctErrorf( tmpl, valName )
    }
    m.pending = nil
    return nil 
}

func ( m *reactorImplMap ) end() error {
    if m.pending == nil { return nil }
    return rctErrorf( 
        "Saw end while expecting value for field '%s'", m.pending )
}

type ReactorImpl struct {
    stack *list.List
    done bool
}

func NewReactorImpl() *ReactorImpl {
    return &ReactorImpl{ stack: &list.List{} }
}

func ( ri *ReactorImpl ) checkActive( call string ) error {
    if ri.done { return rctErrorf( "%s() called, but struct is built", call ) }
    return nil
}

// ri.stack known to be non-empty when this returns without error
func ( ri *ReactorImpl ) checkValue( valName string ) error {
    if ri.stack.Len() == 0 {
        return getTopLevelStructStartExpectError( valName )
    }
    elt := ri.stack.Front().Value
    if m, ok := elt.( *reactorImplMap ); ok { return m.checkValue( valName ) }
    return nil
}

func ( ri *ReactorImpl ) push( elt interface{} ) { ri.stack.PushFront( elt ) }

func ( ri *ReactorImpl ) StartStruct() error {
    if err := ri.checkActive( "StartStruct" ); err != nil { return err }
    // skip check if we're pushing the top level struct
    if ri.stack.Len() > 0 {
        if err := ri.checkValue( "struct start" ); err != nil { return err }
    }
    ri.push( newReactorImplMap() )
    return nil
}

func ( ri *ReactorImpl ) StartMap() error {
    if err := ri.checkActive( "StartMap" ); err != nil { return err }
    if err := ri.checkValue( "map start" ); err != nil { return err }
    ri.push( newReactorImplMap() )
    return nil
}

func ( ri *ReactorImpl ) StartList() error {
    if err := ri.checkActive( "StartList" ); err != nil { return err }
    if err := ri.checkValue( "list start" ); err != nil { return err }
    ri.push( "list" )
    return nil
}

func ( ri *ReactorImpl ) StartField( fld *Identifier ) error {
    if err := ri.checkActive( "StartField" ); err != nil { return err }
    if ok := ri.stack.Len() > 0; ok {
        elt := ri.stack.Front().Value
        switch v := elt.( type ) {
        case string: 
            tmpl := "Expected list value but got start of field '%s'"
            return rctErrorf( tmpl, fld )
        case *reactorImplMap: return v.startField( fld )
        default: panic( libErrorf( "Invalid stack element: %T", elt ) )
        }
    }
    errLoc := fmt.Sprintf( "field '%s'", fld )
    return getTopLevelStructStartExpectError( errLoc )
}

func ( ri *ReactorImpl ) Value() error {
    if err := ri.checkActive( "Value" ); err != nil { return err }
    if err := ri.checkValue( "value" ); err != nil { return err }
    return nil
}

func ( ri *ReactorImpl ) End() error {
    if err := ri.checkActive( "End" ); err != nil { return err }
    if ri.stack.Len() == 0 { return getTopLevelStructStartExpectError( "end" ) }
    elt := ri.stack.Remove( ri.stack.Front() )
    switch v := elt.( type ) {
    case *reactorImplMap: if err := v.end(); err != nil { return err }
    case string: {} // list -- end() is always okay
    default: panic( libErrorf( "Unexpected stack element: %T", elt ) )
    }
    ri.done = ri.stack.Len() == 0
    return nil
}

type valAccumulator interface {
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
    typ TypeReference
    flds *mapAcc
}

func newStructAcc( typ TypeReference ) *structAcc {
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

type StructBuilder struct {
    s *Struct
    accs *list.List
    impl *ReactorImpl
}

func NewStructBuilder() *StructBuilder {
    return &StructBuilder{ accs: &list.List{}, impl: NewReactorImpl() }
}

func ( sb *StructBuilder ) Ready() bool { return sb.s != nil }

func ( sb *StructBuilder ) pushAcc( acc valAccumulator ) {
    sb.accs.PushFront( acc )
}

func ( sb *StructBuilder ) peekAcc() ( valAccumulator, bool ) {
    if sb.accs.Len() == 0 { return nil, false }
    return sb.accs.Front().Value.( valAccumulator ), true
}

func ( sb *StructBuilder ) popAcc() valAccumulator {
    res, ok := sb.peekAcc()
    if ! ok { panic( libErrorf( "popAcc() called on empty stack" ) ) }
    sb.accs.Remove( sb.accs.Front() )
    return res
}

func ( sb *StructBuilder ) valueReady( val Value ) {
    if acc, ok := sb.peekAcc(); ok {
        acc.valueReady( val )
    } else { sb.s = val.( *Struct ) }
}

// Panics if result of Ready() is false
func ( sb *StructBuilder ) GetStruct() *Struct {
    if sb.Ready() { return sb.s }
    panic( rctErrorf( "Attempt to access incomplete struct" ) )
}

func ( sb *StructBuilder ) StartStruct( typ TypeReference ) error {
    if err := sb.impl.StartStruct(); err != nil { return err }
    sb.pushAcc( newStructAcc( typ ) )
    return nil
}

func ( sb *StructBuilder ) StartMap() error {
    if err := sb.impl.StartMap(); err != nil { return err }
    sb.pushAcc( newMapAcc() )
    return nil
}

func ( sb *StructBuilder ) StartList() error {
    if err := sb.impl.StartList(); err != nil { return err }
    sb.pushAcc( newListAcc() )
    return nil
}

func ( sb *StructBuilder ) StartField( fld *Identifier ) error {
    if err := sb.impl.StartField( fld ); err != nil { return err }
    acc, ok := sb.peekAcc()
    if ok {
        var ma *mapAcc
        switch v := acc.( type ) {
        case *mapAcc: ma, ok = v, true
        case *structAcc: ma, ok = v.flds, true
        default: ok = false
        }
        if ok { ma.startField( fld ) }
    }
    return nil
}

func ( sb *StructBuilder ) End() error {
    if err := sb.impl.End(); err != nil { return err }
    acc := sb.popAcc()
    if val, err := acc.end(); err == nil {
        sb.valueReady( val )
    } else { return err }
    return nil
}

func ( sb *StructBuilder ) Value( mv Value ) error {
    if err := sb.impl.Value(); err != nil { return err }
    sb.valueReady( mv ) 
    return nil
}

func visitSymbolMap( 
    m *SymbolMap, callStart bool, rct Reactor ) error {
    if callStart {
        if err := rct.StartMap(); err != nil { return err }
    }
    err := m.EachPairError( func( fld *Identifier, val Value ) error {
        if err := rct.StartField( fld ); err != nil { return err }
        return visitValue( val, rct )
    })
    if err != nil { return err }
    return rct.End()
}

func visitStruct( ms *Struct, rct Reactor ) ( err error ) {
    if err = rct.StartStruct( ms.Type ); err != nil { return }
    return visitSymbolMap( ms.Fields, false, rct )
}

func visitList( ml *List, rct Reactor ) ( err error ) {
    if err = rct.StartList(); err != nil { return }
    for _, val := range ml.Values() {
        if err = visitValue( val, rct ); err != nil { return }
    }
    return rct.End()
}

func visitValue( mv Value, rct Reactor ) error {
    switch v := mv.( type ) {
    case *Struct: return visitStruct( v, rct )
    case *SymbolMap: return visitSymbolMap( v, true, rct )
    case *List: return visitList( v, rct )
    }
    return rct.Value( mv )
}

func VisitStruct( ms *Struct, rct Reactor ) error {
    return visitValue( ms, rct )
}
