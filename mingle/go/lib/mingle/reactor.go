package mingle

import (
    "container/list"
    "fmt"
//    "log"
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

type ReactorTopType int

const (
    ReactorTopTypeValue = ReactorTopType( iota )
    ReactorTopTypeList
    ReactorTopTypeMap 
    ReactorTopTypeStruct 
)

func ( t ReactorTopType ) String() string {
    switch t {
    case ReactorTopTypeValue: return "value"
    case ReactorTopTypeList: return "list"
    case ReactorTopTypeMap: return "map"
    case ReactorTopTypeStruct: return "struct"
    }
    panic( libErrorf( "Unhandled reactor top type: %d", t ) )
}

func getReactorTopTypeError( valName string, tt ReactorTopType ) error {
    return rctErrorf( "Expected %s but got %s", tt, valName )
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
    topTyp ReactorTopType
    done bool
}

func NewReactorImpl() *ReactorImpl {
    return &ReactorImpl{ stack: &list.List{}, topTyp: ReactorTopTypeStruct }
}

func ( ri *ReactorImpl ) getReactorTopTypeError( valName string ) error {
    return getReactorTopTypeError( valName, ri.topTyp )
}

func ( ri *ReactorImpl ) checkActive( call string ) error {
    if ri.done { return rctErrorf( "%s() called, but struct is built", call ) }
    return nil
}

// ri.stack known to be non-empty when this returns without error, unless top
// type is value.
func ( ri *ReactorImpl ) checkValue( valName string ) error {
    if ri.stack.Len() == 0 {
        if ri.topTyp == ReactorTopTypeValue { return nil }
        return ri.getReactorTopTypeError( valName )
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
    return getReactorTopTypeError( errLoc, ReactorTopTypeStruct )
}

func ( ri *ReactorImpl ) Value() error {
    if err := ri.checkActive( "Value" ); err != nil { return err }
    if err := ri.checkValue( "value" ); err != nil { return err }
    return nil
}

func ( ri *ReactorImpl ) End() error {
    if err := ri.checkActive( "End" ); err != nil { return err }
    if ri.stack.Len() == 0 { return ri.getReactorTopTypeError( "end" ) }
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

type ValueBuilder struct {
    val Value
    accs *list.List
    impl *ReactorImpl
}

func NewValueBuilder() *ValueBuilder {
    return &ValueBuilder{ accs: &list.List{}, impl: NewReactorImpl() }
}

// The result of setting this once reactions have begun is undefined.
func ( vb *ValueBuilder ) SetTopType( topTyp ReactorTopType ) {
    vb.impl.topTyp = topTyp
}

func ( vb *ValueBuilder ) Ready() bool { return vb.val != nil }

func ( vb *ValueBuilder ) pushAcc( acc valAccumulator ) {
    vb.accs.PushFront( acc )
}

func ( vb *ValueBuilder ) peekAcc() ( valAccumulator, bool ) {
    if vb.accs.Len() == 0 { return nil, false }
    return vb.accs.Front().Value.( valAccumulator ), true
}

func ( vb *ValueBuilder ) popAcc() valAccumulator {
    res, ok := vb.peekAcc()
    if ! ok { panic( libErrorf( "popAcc() called on empty stack" ) ) }
    vb.accs.Remove( vb.accs.Front() )
    return res
}

func ( vb *ValueBuilder ) valueReady( val Value ) {
    if acc, ok := vb.peekAcc(); ok {
        acc.valueReady( val )
    } else { vb.val = val }
}

// Panics if result of Ready() is false
func ( vb *ValueBuilder ) GetValue() Value {
    if vb.Ready() { return vb.val }
    panic( rctErrorf( "Attempt to access incomplete value" ) )
}

func ( vb *ValueBuilder ) StartStruct( typ TypeReference ) error {
    if err := vb.impl.StartStruct(); err != nil { return err }
    vb.pushAcc( newStructAcc( typ ) )
    return nil
}

func ( vb *ValueBuilder ) StartMap() error {
    if err := vb.impl.StartMap(); err != nil { return err }
    vb.pushAcc( newMapAcc() )
    return nil
}

func ( vb *ValueBuilder ) StartList() error {
    if err := vb.impl.StartList(); err != nil { return err }
    vb.pushAcc( newListAcc() )
    return nil
}

func ( vb *ValueBuilder ) StartField( fld *Identifier ) error {
    if err := vb.impl.StartField( fld ); err != nil { return err }
    acc, ok := vb.peekAcc()
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

func ( vb *ValueBuilder ) End() error {
    if err := vb.impl.End(); err != nil { return err }
    acc := vb.popAcc()
    if val, err := acc.end(); err == nil {
        vb.valueReady( val )
    } else { return err }
    return nil
}

func ( vb *ValueBuilder ) Value( mv Value ) error {
    if err := vb.impl.Value(); err != nil { return err }
    vb.valueReady( mv ) 
    return nil
}

func visitSymbolMap( 
    m *SymbolMap, callStart bool, rct Reactor ) error {
    if callStart {
        if err := rct.StartMap(); err != nil { return err }
    }
    err := m.EachPairError( func( fld *Identifier, val Value ) error {
        if err := rct.StartField( fld ); err != nil { return err }
        return VisitValue( val, rct )
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
        if err = VisitValue( val, rct ); err != nil { return }
    }
    return rct.End()
}

func VisitValue( mv Value, rct Reactor ) error {
    switch v := mv.( type ) {
    case *Struct: return visitStruct( v, rct )
    case *SymbolMap: return visitSymbolMap( v, true, rct )
    case *List: return visitList( v, rct )
    }
    return rct.Value( mv )
}
