package mingle

import (
    "container/list"
    "fmt"
    "bitgirder/objpath"
    "bytes"
//    "log"
)

type ReactorError struct { msg string }

func ( e *ReactorError ) Error() string { return e.msg }

func rctError( msg string ) *ReactorError { return &ReactorError{ msg } }

func rctErrorf( tmpl string, args ...interface{} ) *ReactorError {
    return rctError( fmt.Sprintf( tmpl, args... ) )
}

type ReactorEvent interface { reactorEventImpl() }

type ValueEvent struct { Val Value }
func ( ve ValueEvent ) reactorEventImpl() {}

type StructStartEvent struct { Type *QualifiedTypeName }
func ( sse StructStartEvent ) reactorEventImpl() {}

type MapStartEvent int
func ( mse MapStartEvent ) reactorEventImpl() {}

const EvMapStart = MapStartEvent( 0 )

type FieldStartEvent struct { Field *Identifier }
func ( fse FieldStartEvent ) reactorEventImpl() {}

type ListStartEvent int
func( lse ListStartEvent ) reactorEventImpl() {}

const EvListStart = ListStartEvent( 0 )

type EndEvent int
func ( ee EndEvent ) reactorEventImpl() {}

const EvEnd = EndEvent( 0 )

type ReactorEventProcessor interface { ProcessEvent( ReactorEvent ) error }

type discardEventProcessor int

func ( d discardEventProcessor ) ProcessEvent( ev ReactorEvent ) error {
    return nil
}

var DiscardProcessor = discardEventProcessor( 1 )

type ReactorKey string

type PipelineInitializer interface { Init( rpi *ReactorPipelineInit ) }

type PipelineProcessor interface {
    ProcessEvent( ev ReactorEvent, rep ReactorEventProcessor ) error
}

type KeyedProcessor interface { Key() ReactorKey }

type ReactorPipeline struct {
    elts []interface{}
}

type ReactorPipelineInit struct { 
    rp *ReactorPipeline 
    elts []interface{}
}

func findReactor( 
    elts []interface{}, key ReactorKey ) ( KeyedProcessor, bool ) {
    for _, rct := range elts { 
        if kr, ok := rct.( KeyedProcessor ); ok {
            if kr.Key() == key { return kr, true }
        }
    }
    return nil, false
}

func ( rpi *ReactorPipelineInit ) FindByKey( 
    k ReactorKey ) ( KeyedProcessor, bool ) {
    return findReactor( rpi.elts, k )
}

// public frontends enforce that elt is of a valid type for a pipeline
// (AddEventProcessor(), etc)
func ( rpi *ReactorPipelineInit ) implAdd( elt interface{} ) {
    if ri, ok := elt.( PipelineInitializer ); ok { ri.Init( rpi ) }
    rpi.elts = append( rpi.elts, elt )
}

func ( rpi *ReactorPipelineInit ) AddEventProcessor( 
    rep ReactorEventProcessor ) {
    rpi.implAdd( rep )
}

func ( rpi *ReactorPipelineInit ) AddPipelineProcessor( pp PipelineProcessor ) {
    rpi.implAdd( pp )
}

func ( rpi *ReactorPipelineInit ) VisitPredecessors( f func( interface{} ) ) {
    for _, rct := range rpi.elts { f( rct ) }
}

// Might make this Init() if needed later
func ( rp *ReactorPipeline ) init() {
    rpInit := &ReactorPipelineInit{ 
        rp: rp,
        elts: make( []interface{}, 0, len( rp.elts ) ),
    } 
    for _, elt := range rp.elts { rpInit.implAdd( elt ) }
    rp.elts = rpInit.elts
}

func LastPathGetter( rpi *ReactorPipelineInit ) PathGetter {
    var res PathGetter
    rpi.VisitPredecessors( func( rct interface{} ) {
        if pg, ok := rct.( PathGetter ); ok { res = pg }
    })
    return res
}

// Could break this into separate methods later if needed: NewReactorPipeline()
// and ReactorPipeline.Init()
func InitReactorPipeline( elts ...interface{} ) *ReactorPipeline {
    res := &ReactorPipeline{ elts: elts }
    res.init()
    return res
}

type pipelineCall struct {
    rp *ReactorPipeline
    idx int
}

func ( pc pipelineCall ) ProcessEvent( re ReactorEvent ) error {
    if pc.idx == len( pc.rp.elts ) { return nil }
    nextPc := pipelineCall{ pc.rp, pc.idx + 1 }
    elt := pc.rp.elts[ pc.idx ]
    switch v := elt.( type ) {
    case PipelineProcessor: return v.ProcessEvent( re, nextPc )
    case ReactorEventProcessor:
        if err := v.ProcessEvent( re ); err != nil { return err }
        return nextPc.ProcessEvent( re )
    }
    panic( libErrorf( "Unhandled pipeline element: %T", elt ) )
}

func ( rp *ReactorPipeline ) ProcessEvent( re ReactorEvent ) error {
    return ( pipelineCall{ rp, 0 } ).ProcessEvent( re )
}

func ( rp *ReactorPipeline ) FindByKey( 
    k ReactorKey ) ( KeyedProcessor, bool ) {
    return findReactor( rp.elts, k )
}

func ( rp *ReactorPipeline ) MustFindByKey( k ReactorKey ) KeyedProcessor {
    if rct, ok := rp.FindByKey( k ); ok { return rct }
    panic( libErrorf( "No reactor for key %q", k ) )
}

func visitSymbolMap( 
    m *SymbolMap, callStart bool, rep ReactorEventProcessor ) error {
    if callStart {
        if err := rep.ProcessEvent( EvMapStart ); err != nil { return err }
    }
    err := m.EachPairError( func( fld *Identifier, val Value ) error {
        ev := FieldStartEvent{ fld }
        if err := rep.ProcessEvent( ev ); err != nil { return err }
        return VisitValue( val, rep )
    })
    if err != nil { return err }
    return rep.ProcessEvent( EvEnd )
}

func visitStruct( ms *Struct, rep ReactorEventProcessor ) error {
    ev := StructStartEvent{ ms.Type }
    if err := rep.ProcessEvent( ev ); err != nil { return err }
    return visitSymbolMap( ms.Fields, false, rep )
}

func visitList( ml *List, rep ReactorEventProcessor ) error {
    if err := rep.ProcessEvent( EvListStart ); err != nil { return err }
    for _, val := range ml.Values() {
        if err := VisitValue( val, rep ); err != nil { return err }
    }
    return rep.ProcessEvent( EvEnd )
}

func VisitValue( mv Value, rep ReactorEventProcessor ) error {
    switch v := mv.( type ) {
    case *Struct: return visitStruct( v, rep )
    case *SymbolMap: return visitSymbolMap( v, true, rep )
    case *List: return visitList( v, rep )
    }
    return rep.ProcessEvent( ValueEvent{ mv } )
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

type structuralMap struct {
    pending *Identifier
    keys *IdentifierMap
}

func newStructuralMap() *structuralMap { 
    return &structuralMap{ keys: NewIdentifierMap() } 
}

type listIndex int

type eventStack struct {
    *list.List
}

func newEventStack() *eventStack {
    return &eventStack{ List: &list.List{} }
}

func ( s *eventStack ) buildPath( e *list.Element, p idPath ) idPath {
    if e == nil { return p }
    switch v := e.Value.( type ) {
    case *Identifier: p = objpath.Descend( p, v )
    case listIndex: 
        if v >= 0 { p = objpath.StartList( p ).SetIndex( int( v ) ) }
    }
    return s.buildPath( e.Prev(), p )
}

func ( s *eventStack ) AppendPath( p idPath ) idPath {
    return s.buildPath( s.Back(), p )
}

func ( s *eventStack ) GetPath() objpath.PathNode { return s.AppendPath( nil ) }

func ( s *eventStack ) isEmpty() bool { return s.Len() == 0 }

func ( s *eventStack ) peekElt() *list.Element { return s.Front() }

func ( s *eventStack ) peek() interface{} {
    if elt := s.peekElt(); elt != nil { return elt.Value }
    return nil
}

func ( s *eventStack ) pop() interface{} { return s.Remove( s.Front() ) }

func ( s *eventStack ) pushMap( val interface{} ) { s.PushFront( val ) }

func ( s *eventStack ) pushField( fld *Identifier ) { 
    s.PushFront( fld ) 
}

func ( s *eventStack ) pushList( val interface{} ) { s.PushFront( val ) }

// Increments the list index if the top of stack is a list
func ( s *eventStack ) prepareListVal() {
    if elt := s.peekElt(); elt != nil {
        if idx, ok := elt.Value.( listIndex ); ok {
            elt.Value = listIndex( idx + 1 )
        }
    }
}

// ReactorEventProcessor that keeps track of the path inherent in an event
// stream. This processor does not check that the stream of events itself
// represents a valid object.
type EventPathReactor struct {
    stack *eventStack
    rep ReactorEventProcessor
}

func NewEventPathReactor( rep ReactorEventProcessor ) *EventPathReactor {
    return &EventPathReactor{ rep: rep, stack: newEventStack() }
}

func ( epr *EventPathReactor ) GetPath() objpath.PathNode {
    return epr.stack.GetPath()
}

func ( epr *EventPathReactor ) AppendPath( 
    src objpath.PathNode ) objpath.PathNode {
    return epr.stack.AppendPath( src )
}

func ( epr *EventPathReactor ) preProcessValue() {
    epr.stack.prepareListVal()
}

func ( epr *EventPathReactor ) preProcess( ev ReactorEvent ) {
    switch v := ev.( type ) {
    case FieldStartEvent: epr.stack.pushField( v.Field )
    case MapStartEvent, StructStartEvent: 
        epr.preProcessValue()
        epr.stack.pushMap( "map" )
    case ListStartEvent: 
        epr.preProcessValue()
        epr.stack.pushList( listIndex( -1 ) )
    case EndEvent: epr.stack.pop()
    case ValueEvent: epr.preProcessValue()
    default: panic( libErrorf( "Unhandled event: %T", ev ) )
    }
}

func ( epr *EventPathReactor ) completeValue() {
    if epr.stack.isEmpty() { return }
    switch v := epr.stack.peek().( type ) {
    case *Identifier, string: epr.stack.pop()
    case listIndex: {}
    default: panic( libErrorf( "Unhandled feed stack element: %T", v ) )
    }
}

func ( epr *EventPathReactor ) postProcess( ev ReactorEvent ) error {
    switch ev.( type ) {
    case ValueEvent, EndEvent: epr.completeValue()
    }
    return nil
}

func ( epr *EventPathReactor ) ProcessEvent( ev ReactorEvent ) error {
    epr.preProcess( ev )
    if err := epr.rep.ProcessEvent( ev ); err != nil { return err }
    epr.postProcess( ev )
    return nil
}

type StructuralReactor struct {
    stack *eventStack
    topTyp ReactorTopType
    done bool
}

func NewStructuralReactor( topTyp ReactorTopType ) *StructuralReactor {
    return &StructuralReactor{ 
        stack: newEventStack(),
        topTyp: topTyp,
    }
}

const ReactorKeyStructuralReactor = ReactorKey( "mingle.StructuralReactor" )

func ( sr *StructuralReactor ) Key() ReactorKey { 
    return ReactorKeyStructuralReactor
}

func ( sr *StructuralReactor ) GetPath() objpath.PathNode {
    return sr.stack.GetPath()
}

func ( sr *StructuralReactor ) AppendPath( 
    p objpath.PathNode ) objpath.PathNode {
    return sr.stack.AppendPath( p )
}

func ( sr *StructuralReactor ) getReactorTopTypeError( valName string ) error {
    return getReactorTopTypeError( valName, sr.topTyp )
}

func ( sr *StructuralReactor ) checkActive( call string ) error {
    if sr.done { return rctErrorf( "%s() called, but struct is built", call ) }
    return nil
}

func ( sr *StructuralReactor ) mapIsTop() *structuralMap {
    if elt := sr.stack.peek(); elt != nil {
        if m, ok := elt.( *structuralMap ); ok { return m }
    }
    return nil
}

func ( sr *StructuralReactor ) checkMapValue( 
    valName string, m *structuralMap ) error {
    if m.pending != nil { return nil }
    tmpl := "Expected field name or end of fields but got %s"
    return rctErrorf( tmpl, valName )
}

// sr.stack known to be non-empty when this returns without error, unless top
// type is value.
func ( sr *StructuralReactor ) checkValueWithNameFunc( 
    nmFunc func() string ) error {
    if sr.stack.isEmpty() {
        if sr.topTyp == ReactorTopTypeValue { return nil }
        return sr.getReactorTopTypeError( nmFunc() )
    }
    switch v := sr.stack.peek().( type ) {
    case *Identifier: {}
    case listIndex: {}
    case *structuralMap:
        if err := sr.checkMapValue( nmFunc(), v ); err != nil { return err }
    default: panic( libErrorf( "Unexpected stack elt for value: %T", v ) )
    }
    return nil
}

func ( sr *StructuralReactor ) checkValue( valName string ) error {
    return sr.checkValueWithNameFunc( func() string { return valName } )
}

func ( sr *StructuralReactor ) prepareListVal() {
    sr.stack.prepareListVal()
}

func ( sr *StructuralReactor ) implStartMap() error {
    sr.prepareListVal()
    sr.stack.pushMap( newStructuralMap() )
    return nil
}

func ( sr *StructuralReactor ) startStruct( typ *QualifiedTypeName ) error {
    if err := sr.checkActive( "StartStruct" ); err != nil { return err }
    // skip check if we're pushing the top level struct
    if ! sr.stack.isEmpty() {
        nmFunc := func() string { 
            return fmt.Sprintf( "start of struct %s", typ.ExternalForm() )
        }
        if err := sr.checkValueWithNameFunc( nmFunc ); err != nil { return err }
    }
    return sr.implStartMap()
}

func ( sr *StructuralReactor ) startMap() error {
    if err := sr.checkActive( "StartMap" ); err != nil { return err }
    if err := sr.checkValue( "map start" ); err != nil { return err }
    return sr.implStartMap()
}

// Note about the prepareListVal() call below: it has nothing to do with the
// list we're starting; it pertains to the (possible) list to which we're adding
// the current list as a value.
func ( sr *StructuralReactor ) startList() error {
    if err := sr.checkActive( "StartList" ); err != nil { return err }
    if err := sr.checkValue( "list start" ); err != nil { return err }
    sr.prepareListVal() 
    sr.stack.pushList( listIndex( -1 ) )
    return nil
}

func ( sr *StructuralReactor ) startMapField( 
    fld *Identifier, m *structuralMap ) error {
    if m.pending != nil {
        panic( libErrorf( "startMapField while pending: %s", m.pending ) )
    }
    if m.keys.HasKey( fld ) {
        return rctErrorf( "Multiple entries for field: %s", fld )
    }
    m.keys.Put( fld, true )
    m.pending = fld
    return nil
}

func ( sr *StructuralReactor ) startField( fld *Identifier ) error {
    if err := sr.checkActive( "StartField" ); err != nil { return err }
    if elt := sr.stack.peek(); elt != nil {
        switch v := elt.( type ) {
        case listIndex: 
            tmpl := "Saw start of field '%s' while expecting a list value"
            return rctErrorf( tmpl, fld )
        case *structuralMap: 
            if err := sr.startMapField( fld, v ); err != nil { return err }
            sr.stack.pushField( fld )
            return nil
        case *Identifier:
            tmpl := 
                "Saw start of field '%s' while expecting a value for field '%s'"
            return rctErrorf( tmpl, fld, v )
        default: panic( libErrorf( "Invalid stack element: %v (%T)", v, v ) )
        }
    }
    errLoc := fmt.Sprintf( "start of field '%s'", fld )
    return getReactorTopTypeError( errLoc, ReactorTopTypeStruct )
}

func ( sr *StructuralReactor ) value( isAtomic bool ) error {
    if err := sr.checkActive( "Value" ); err != nil { return err }
    if err := sr.checkValue( "value" ); err != nil { return err }
    if isAtomic { sr.prepareListVal() }
    return nil
}

func ( sr *StructuralReactor ) end() error {
    if err := sr.checkActive( "End" ); err != nil { return err }
    if sr.stack.isEmpty() { return sr.getReactorTopTypeError( "end" ) }
    switch v := sr.stack.pop().( type ) {
    case *Identifier:
        return rctErrorf( "Saw end while expecting a value for field '%s'", v )
    case *structuralMap, listIndex: {} // end() is always okay
    default: panic( libErrorf( "Unexpected stack element: %T", v ) )
    }
    // if we're not done then we just completed an intermediate value which
    // needs to be processed
    if sr.done = sr.stack.isEmpty(); ! sr.done { return sr.value( false ) }
    return nil
}

func ( sr *StructuralReactor ) update( ev ReactorEvent ) ( bool, error ) {
    switch v := ev.( type ) {
    case StructStartEvent: return false, sr.startStruct( v.Type )
    case MapStartEvent: return false, sr.startMap()
    case ListStartEvent: return false, sr.startList()
    case FieldStartEvent: return false, sr.startField( v.Field )
    case ValueEvent: return true, sr.value( true )
    case EndEvent: return true, sr.end()
    }
    panic( libErrorf( "Unhandled event: %T", ev ) )
}

func ( sr *StructuralReactor ) mapValue( m *structuralMap ) { m.pending = nil }

func ( sr *StructuralReactor ) downstreamDone( ev ReactorEvent, isValue bool ) {
    if isValue {
        if _, ok := sr.stack.peek().( *Identifier ); ok { sr.stack.pop() }
        if m := sr.mapIsTop(); m != nil { sr.mapValue( m ) }
    }
}

func ( sr *StructuralReactor ) ProcessEvent(
    ev ReactorEvent, rep ReactorEventProcessor ) error {
    if isValue, err := sr.update( ev ); err == nil {
        if err = rep.ProcessEvent( ev ); err == nil {
            sr.downstreamDone( ev, isValue )
        } else { return err }
    } else { return err }
    return nil
}

func EnsureStructuralReactor( rpi *ReactorPipelineInit ) *StructuralReactor {
    k := ReactorKeyStructuralReactor
    if elt, ok := rpi.FindByKey( k ); ok {
        if sr, ok := elt.( *StructuralReactor ); ok { return sr }
        tmpl := "Element keyed at %s is not a *StructuralReactor: %T"
        panic( libErrorf( tmpl, k, elt ) )
    }
    sr := NewStructuralReactor( ReactorTopTypeValue )
    rpi.AddPipelineProcessor( sr )
    return sr
}

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
    accs *list.List
}

func newValueAccumulator() *valueAccumulator {
    return &valueAccumulator{ accs: &list.List{} }
}

func ( va *valueAccumulator ) pushAcc( acc accImpl ) {
    va.accs.PushFront( acc )
}

func ( va *valueAccumulator ) peekAcc() ( accImpl, bool ) {
    if va.accs.Len() == 0 { return nil, false }
    return va.accs.Front().Value.( accImpl ), true
}

func ( va *valueAccumulator ) popAcc() accImpl {
    res, ok := va.peekAcc()
    if ! ok { panic( libErrorf( "popAcc() called on empty stack" ) ) }
    va.accs.Remove( va.accs.Front() )
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
    case ValueEvent: va.valueReady( v.Val )
    case ListStartEvent: va.pushAcc( newListAcc() )
    case MapStartEvent: va.pushAcc( newMapAcc() )
    case StructStartEvent: va.pushAcc( newStructAcc( v.Type ) )
    case FieldStartEvent: va.startField( v.Field )
    case EndEvent: if err := va.end(); err != nil { return err }
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

type castContext struct {
    elt interface{}
    expct TypeReference
}

type FieldTyper interface {
    FieldTypeOf( fld *Identifier, pg PathGetter ) ( TypeReference, error )
}

type mapCast struct {
    fldType TypeReference
    typer FieldTyper
}

func newMapCast( typer FieldTyper ) *mapCast { return &mapCast{ typer: typer } }

type listCast struct {
    lt *ListTypeReference
    sawVals bool
}

type CastInterface interface {

    InferStructFor( qn *QualifiedTypeName ) bool

    FieldTyperFor( qn *QualifiedTypeName, pg PathGetter ) ( FieldTyper, error )

    CastAtomic( 
        in Value, 
        at *AtomicTypeReference, 
        pg PathGetter ) ( Value, error, bool )
}

type castInterfaceDefault struct {}

type valueFieldTyper int

func ( vt valueFieldTyper ) FieldTypeOf( 
    fld *Identifier, pg PathGetter ) ( TypeReference, error ) {
    return TypeNullableValue, nil
}

func ( i castInterfaceDefault ) FieldTyperFor( 
    qn *QualifiedTypeName, pg PathGetter ) ( FieldTyper, error ) {
    return valueFieldTyper( 1 ), nil
}

func ( i castInterfaceDefault ) InferStructFor( at *QualifiedTypeName ) bool {
    return false
}

func ( i castInterfaceDefault ) CastAtomic( 
    v Value, at *AtomicTypeReference, pg PathGetter ) ( Value, error, bool ) {
    return nil, nil, false
}

type CastReactor struct {
    pg PathGetter
    iface CastInterface
    stack *list.List // stack of castContext
    sr *StructuralReactor
}

func NewCastReactor( 
    expct TypeReference, iface CastInterface, pg PathGetter ) *CastReactor {
    res := &CastReactor{ pg: pg, stack: &list.List{}, iface: iface }
    res.stack.PushFront( castContext{ elt: expct, expct: expct } )
    return res
}

func NewDefaultCastReactor( expct TypeReference, pg PathGetter ) *CastReactor {
    return NewCastReactor( expct, castInterfaceDefault{}, pg )
}

func ( cr *CastReactor ) Init( rpi *ReactorPipelineInit ) {
    cr.sr = EnsureStructuralReactor( rpi )
}

func ( cr *CastReactor ) checkStackNonEmpty() {
    if cr.stack.Len() == 0 { panic( libErrorf( "Empty cast reactor stack" ) ) }
}

func ( cr *CastReactor ) peek() castContext {
    cr.checkStackNonEmpty()
    return cr.stack.Front().Value.( castContext )
}

func ( cr *CastReactor ) pop() castContext {
    cc := cr.peek()
    cr.stack.Remove( cr.stack.Front() )
    return cc
}

func ( cr *CastReactor ) push( cc castContext ) { cr.stack.PushFront( cc ) }

func ( cr *CastReactor ) GetPath() objpath.PathNode { 
    var p idPath
    if cr.pg != nil { p = cr.pg.GetPath() }
    return cr.sr.AppendPath( p )
}

func ( cr *CastReactor ) expectedType() TypeReference {
    cc := cr.peek()
    if mc, ok := cc.elt.( *mapCast ); ok { return mc.fldType }
    return cc.expct
}

func ( cr *CastReactor ) newTypeCastErrorPath( 
    act TypeReference, p idPath ) *ValueCastError {
    return NewTypeCastError( cr.expectedType(), act, p )
}

func ( cr *CastReactor ) newTypeCastError( act TypeReference ) *ValueCastError {
    return cr.newTypeCastErrorPath( act, cr.GetPath() )
}

func ( cr *CastReactor ) castContextPanic( 
    cc castContext, desc string ) error {
    return libErrorf( "Unhandled stack element for %s: %T", desc, cc.elt )
}

func ( cr *CastReactor ) stackTypePanic( desc string ) error {
    return cr.castContextPanic( cr.peek(), desc )
}

func ( cr *CastReactor ) castAtomic(
    v Value, at *AtomicTypeReference ) ( Value, error ) {
    if val, err, done := cr.iface.CastAtomic( v, at, cr ); done {
        return val, err
    }
    return castAtomic( v, at, cr.GetPath() )
}

func ( cr *CastReactor ) completeValue( 
    ve ValueEvent, t TypeReference, rep ReactorEventProcessor ) error {
    switch typVal := t.( type ) {
    case *AtomicTypeReference: 
        v2, err := cr.castAtomic( ve.Val, typVal )
        if err == nil { return rep.ProcessEvent( ValueEvent{ v2 } ) }
        return err
    case *NullableTypeReference: 
        if _, ok := ve.Val.( *Null ); ok { return rep.ProcessEvent( ve ) }
        return cr.completeValue( ve, typVal.Type, rep )
    case *ListTypeReference: return cr.newTypeCastError( TypeOf( ve.Val ) )
    }
    panic( libErrorf( "Unhandled type: %T", t ) )
}

func ( cr *CastReactor ) value(
    ve ValueEvent, rep ReactorEventProcessor ) error {
    switch elt := cr.peek().elt.( type ) {
    case TypeReference: return cr.completeValue( ve, elt, rep )
    case *listCast: 
        elt.sawVals = true
        return cr.completeValue( ve, elt.lt.ElementType, rep )
    case *mapCast: return cr.completeValue( ve, elt.fldType, rep )
    }
    panic( cr.stackTypePanic( "value" ) )
}

func ( cr *CastReactor ) completeStartList( 
    typ TypeReference, le ListStartEvent, rep ReactorEventProcessor ) error {
    switch t := typ.( type ) {
    case *ListTypeReference:
        cc := castContext{ expct: t.ElementType, elt: &listCast{ lt: t } }
        cr.push( cc )
        return rep.ProcessEvent( le )
    case *NullableTypeReference: return cr.completeStartList( t.Type, le, rep )
    case *AtomicTypeReference:
        if t.Equals( TypeValue ) { 
            return cr.completeStartList( TypeOpaqueList, le, rep )
        }
        return cr.newTypeCastErrorPath( TypeOpaqueList, cr.GetPath() )
    }
    panic( libErrorf( "Unhandled type: %T", typ ) )
}

func ( cr *CastReactor ) startList( 
    le ListStartEvent, rep ReactorEventProcessor ) error {
    switch elt := cr.peek().elt.( type ) {
    case TypeReference: return cr.completeStartList( elt, le, rep )
    case *listCast: return cr.completeStartList( elt.lt.ElementType, le, rep )
    case *mapCast: return cr.completeStartList( elt.fldType, le, rep )
    }
    panic( cr.stackTypePanic( "list start" ) )
}

func ( cr *CastReactor ) inferredStructTypeOf( 
    typ TypeReference ) *QualifiedTypeName {
    switch t := typ.( type ) {
    case *AtomicTypeReference: 
        qn := t.Name.( *QualifiedTypeName )
        if cr.iface.InferStructFor( qn ) { return qn }
    case *NullableTypeReference: return cr.inferredStructTypeOf( t.Type )
    }
    return nil
}

func ( cr *CastReactor ) completeStartMap(
    typ TypeReference, sm MapStartEvent, rep ReactorEventProcessor ) error {
    switch t2 := typ.( type ) {
    case *ListTypeReference: 
        return cr.completeStartMap( t2.ElementType, sm, rep )
    case *NullableTypeReference: return cr.completeStartMap( t2.Type, sm, rep )
    case *AtomicTypeReference: {} // rest of method below
    default: panic( libErrorf( "Unhandled type reference: %T", typ ) )
    }
    if typ.Equals( TypeSymbolMap ) || typ.Equals( TypeValue ) {
        mc := newMapCast( valueFieldTyper( 1 ) )
        cr.push( castContext{ elt: mc, expct: TypeSymbolMap } )
        return rep.ProcessEvent( sm )
    }
    if qn := cr.inferredStructTypeOf( typ ); qn != nil {
        at := &AtomicTypeReference{ Name: qn }
        return cr.completeStartStruct( StructStartEvent{ qn }, at, rep )
    }
    return cr.newTypeCastError( TypeSymbolMap )
}

func ( cr *CastReactor ) startMap( 
    sm MapStartEvent, rep ReactorEventProcessor ) error {
    switch elt := cr.peek().elt.( type ) {
    case *AtomicTypeReference, *NullableTypeReference: 
        return cr.completeStartMap( elt.( TypeReference ), sm, rep )
    case *listCast: return cr.completeStartMap( elt.lt.ElementType, sm, rep )
    case *mapCast: return cr.completeStartMap( elt.fldType, sm, rep )
    }
    panic( cr.stackTypePanic( "start map" ) )
}

func ( cr *CastReactor ) completeStartStruct( 
    ss StructStartEvent, t TypeReference, rep ReactorEventProcessor ) error {
    if nt, ok := t.( *NullableTypeReference ); ok {
        return cr.completeStartStruct( ss, nt.Type, rep )
    }
    var expctTyp TypeReference
    var ev ReactorEvent
    at := &AtomicTypeReference{ Name: ss.Type }
    switch {
    case t.Equals( at ) || t.Equals( TypeValue ): expctTyp, ev = at, ss
    case t.Equals( TypeSymbolMap ): expctTyp, ev = TypeSymbolMap, EvMapStart
    default: return cr.newTypeCastError( at )
    }
    ft, err := cr.iface.FieldTyperFor( ss.Type, cr )
    if err != nil { return err }
    if ft == nil { ft = valueFieldTyper( 1 ) }
    cr.push( castContext{ elt: newMapCast( ft ), expct: expctTyp } )
    return rep.ProcessEvent( ev )
}

func ( cr *CastReactor ) startStruct( 
    ss StructStartEvent, rep ReactorEventProcessor ) error {
    switch elt := cr.peek().elt.( type ) {
    case *AtomicTypeReference, *NullableTypeReference: 
        return cr.completeStartStruct( ss, elt.( TypeReference ), rep )
    case *ListTypeReference: 
        return cr.newTypeCastError( &AtomicTypeReference{ Name: ss.Type } )
    case *listCast: return cr.completeStartStruct( ss, elt.lt.ElementType, rep )
    case *mapCast: return cr.completeStartStruct( ss, elt.fldType, rep )
    }
    panic( cr.stackTypePanic( "start struct" ) )
}

func ( cr *CastReactor ) startField( 
    fse FieldStartEvent, rep ReactorEventProcessor ) error {
    mc := cr.peek().elt.( *mapCast ) // okay since structure check precedes
    var err error
    mc.fldType, err = mc.typer.FieldTypeOf( fse.Field, cr )
    if err != nil { return err }
    return rep.ProcessEvent( fse )
}

func ( cr *CastReactor ) noteEndAsValIfList() {
    if cr.stack.Len() == 0 { return }
    if lc, ok := cr.peek().elt.( *listCast ); ok { lc.sawVals = true }
}

func ( cr *CastReactor ) end( ee EndEvent, rep ReactorEventProcessor ) error {
    cc := cr.pop()
    cr.noteEndAsValIfList()
    switch v := cc.elt.( type ) {
    case *mapCast: return rep.ProcessEvent( ee )
    case *listCast:
        if ! ( v.sawVals || v.lt.AllowsEmpty ) {
            return NewValueCastErrorf( cr.GetPath(), "List is empty" )
        }
        return rep.ProcessEvent( ee )
    }
    panic( cr.castContextPanic( cc, "end" ) )
}

func ( cr *CastReactor ) ProcessEvent( 
    ev ReactorEvent, rep ReactorEventProcessor ) error {
    switch v := ev.( type ) {
    case ValueEvent: return cr.value( v, rep )
    case ListStartEvent: return cr.startList( v, rep )
    case MapStartEvent: return cr.startMap( v, rep )
    case StructStartEvent: return cr.startStruct( v, rep )
    case FieldStartEvent: return cr.startField( v, rep )
    case EndEvent: return cr.end( v, rep )
    }
    panic( libErrorf( "Unhandled event: %T", ev ) )
}

func castValue( mgVal Value, cr *CastReactor ) ( Value, error ) {
    vb := NewValueBuilder()
    pip := InitReactorPipeline( cr, vb )
    if err := VisitValue( mgVal, pip ); err != nil { return nil, err }
    return vb.GetValue(), nil
}

func CastValue( 
    mgVal Value, typ TypeReference, path objpath.PathNode ) ( Value, error ) {
//    vb := NewValueBuilder()
    pg := ImmediatePathGetter{ path }
//    pip := InitReactorPipeline( NewDefaultCastReactor( typ, pg ), vb )
//    if err := VisitValue( mgVal, pip ); err != nil { return nil, err }
//    return vb.GetValue(), nil
    return castValue( mgVal, NewDefaultCastReactor( typ, pg ) )
}

type FieldOrderSpecification struct {
    Field *Identifier
    Required bool
}

type FieldOrder []FieldOrderSpecification

// Returns a field ordering for use by a FieldOrderReactor. The ordering is such
// that for any fields f1, f2 such that f1 appears before f2 in the ordering, f1
// will be sent to the associated FieldOrderReactor's downstream processors
// ahead of f2. For fields not appearing in an ordering, there are no guarantees
// as to when they will appear relative to ordered fields. 
type FieldOrderGetter interface {
    FieldOrderFor( qn *QualifiedTypeName ) FieldOrder
}

// Reorders events for selected struct types according to an order determined by
// a FieldOrderGetter.
//
// The implementation is based on a stack of *fieldOrderCtx instances, each of
// which tracks field orderings for some struct type. In cases where a struct
// has no specified order, the *fieldOrderCtx tracks the trivial empty ordering.
type FieldOrderReactor struct {
    fog FieldOrderGetter
    stack *list.List
    pg PathGetter
}

func NewFieldOrderReactor( fog FieldOrderGetter ) *FieldOrderReactor {
    return &FieldOrderReactor{
        fog: fog,
        stack: &list.List{},
    }
}

func ( fo *FieldOrderReactor ) Init( rpi *ReactorPipelineInit ) {
    EnsureStructuralReactor( rpi )
    fo.pg = LastPathGetter( rpi )
}

type fieldOrderCtx struct {
    parent *fieldOrderCtx
    ord FieldOrder
    valStates *IdentifierMap
    idx int
    acc []ReactorEvent
    accFld *Identifier
    valDepth int
    epRct *EventPathReactor
    encl *FieldOrderReactor // The enclosing reactor
}

func ( foc *fieldOrderCtx ) failRepeated( fld *Identifier ) error {
    return libErrorf( "repeated field: %s", fld )
}

func ( foc *fieldOrderCtx ) nextFieldSpec() FieldOrderSpecification {
    if foc.idx < len( foc.ord ) { return foc.ord[ foc.idx ] }
    panic( libErrorf( "no next field in order" ) )
}

func ( foc *fieldOrderCtx ) nextFieldId() *Identifier {
    return foc.nextFieldSpec().Field
}

func ( foc *fieldOrderCtx ) checkAccNil( errLoc string ) {
    if foc.acc != nil { panic( libErrorf( "acc is non-nil %s", errLoc ) ) }
}

func ( foc *fieldOrderCtx ) sendEvent( 
    ev ReactorEvent, rep ReactorEventProcessor ) error {
    if foc.acc == nil {
        if foc.parent == nil { return rep.ProcessEvent( ev ) }
        return foc.parent.sendEvent( ev, rep )
    }
    foc.acc = append( foc.acc, ev )
    return nil
}

//func ( foc *fieldOrderCtx ) startField( fld *Identifier ) error {
//    foc.checkAccNil( fmt.Sprintf( "at start of field %s", fld ) )
//    foc.accFld = fld
//    if ! foc.valStates.HasKey( fld ) { return nil }
//    vs := foc.valStates.Get( fld )
//    switch v := vs.( type ) {
//    case bool:
//        if v { panic( foc.failRepeated( fld ) ) }
//        nxt := foc.nextFieldSpec()
//        if nxt.Field.Equals( fld ) { return nil }
//        if nxt.Required {
//            return NewMissingFieldsError(   
//                foc.encl.GetPath(), []*Identifier{ nxt.Field } )
//        }
//        foc.acc = []ReactorEvent{}
//        return nil
//    case []ReactorEvent: panic( foc.failRepeated( fld ) )
//    }
//    panic( libErrorf( "Unhandled val state: %T", vs ) )
//}

func ( foc *fieldOrderCtx ) startField( fld *Identifier ) {
    foc.checkAccNil( fmt.Sprintf( "at start of field %s", fld ) )
    foc.accFld = fld
    if ! foc.valStates.HasKey( fld ) { return }
    switch vs := foc.valStates.Get( fld ).( type ) {
    case bool:
        if vs { panic( foc.failRepeated( fld ) ) }
        if nxt := foc.nextFieldId(); nxt.Equals( fld ) { return }
        foc.acc = []ReactorEvent{}
    case []ReactorEvent: panic( foc.failRepeated( fld ) )
    default: panic( libErrorf( "Unhandled val state: %T", vs ) )
    }
}

func ( foc *fieldOrderCtx ) updateValDepth( ev ReactorEvent ) {
    switch ev.( type ) {
    case MapStartEvent, ListStartEvent: foc.valDepth++
    case EndEvent: foc.valDepth--
    }
}

func ( foc *fieldOrderCtx ) completeEvent( 
    ev ReactorEvent, rep ReactorEventProcessor ) error {
    foc.updateValDepth( ev )
    if foc.valDepth == 0 && foc.isFieldCompleter( ev ) { 
        return foc.fieldCompleted( rep ) 
    }
    return nil
}

// Initial sanity check of p being a DictNode: we should only ever be feeding
// fields of a map/struct
func ( foc *fieldOrderCtx ) appendFeedPath( 
    p objpath.PathNode ) objpath.PathNode {
    switch p.( type ) {
    case nil: {}
    case *objpath.DictNode: 
        p = p.Parent() // It's a field, but we're feeding a sibling field
    default:
        panic( libErrorf( "Not a dict node (%T): %s", p, FormatIdPath( p ) ) )
    }
    return foc.epRct.stack.AppendPath( p )
}

func ( foc *fieldOrderCtx ) feedValue(
    fld *Identifier, acc []ReactorEvent, rep ReactorEventProcessor ) error {
    foc.checkAccNil( "at feedValue()" )
    foc.epRct = NewEventPathReactor( rep )
    defer func() { foc.epRct = nil }()
    for _, ev := range acc {
        if err := foc.sendEvent( ev, foc.epRct ); err != nil { return err }
    }
    return nil
}

func ( foc *fieldOrderCtx ) sendReadyValues( rep ReactorEventProcessor ) error {
    for foc.idx < len( foc.ord ) {
        fld := foc.nextFieldId()
        vs := foc.valStates.Get( fld )
        if acc, ok := vs.( []ReactorEvent ); ok {
            if err := foc.feedValue( fld, acc, rep ); err != nil { return err }
            foc.valStates.Put( fld, true )
            foc.idx++
        } else { break }
    }
    return nil
}

func ( foc *fieldOrderCtx ) isFieldCompleter( ev ReactorEvent ) bool {
    switch ev.( type ) {
    case EndEvent, ValueEvent: return true
    }
    return false
}

// Upon entry, we are in one of two states, as indicated by whether foc.accFld
// == foc.nextFieldId()
//
//  - 'live': we just sent all field value events directly to the downstream
//  processors; there is no accumulator foc.acc
//
//  - 'accumulating': we saw the current field foc.accFld out of order and
//  accumulated events into foc.acc
//
// If we were live then we bump foc.idx and set the valState to true and, before
// returning, send any previously accumulated and now-order-appropriate values
// via sendReadyValues(). 
//
// If we were accumulating an out-of-order field, we set the accumulated events
// for later and return, and do not attempt to send any previously accumulated
// fields
func ( foc *fieldOrderCtx ) fieldCompleted( rep ReactorEventProcessor ) error {
    if foc.idx < len( foc.ord ) && foc.accFld.Equals( foc.nextFieldId() ) {
        foc.idx++
        foc.valStates.Put( foc.accFld, true )
    }
    fld := foc.accFld // Save before we nil it in case we enter branch below
    foc.accFld = nil
    if foc.acc != nil {
        foc.valStates.Put( fld, foc.acc )
        foc.acc = nil 
        return nil
    }
    return foc.sendReadyValues( rep )
}

func ( foc *fieldOrderCtx ) processEvent(
    ev ReactorEvent, rep ReactorEventProcessor ) error {
    if fs, ok := ev.( FieldStartEvent ); ok && foc.accFld == nil {
//        if err := foc.startField( fs.Field ); err != nil { return err }
        foc.startField( fs.Field )
    }
    if err := foc.sendEvent( ev, rep ); err != nil { return err }
    return foc.completeEvent( ev, rep )
}

func ( foc *fieldOrderCtx ) checkRequiredFields() error {
    miss := []*Identifier{}
    for _, spec := range foc.ord {
        switch v := foc.valStates.Get( spec.Field ).( type ) {
        case bool: 
            if spec.Required && ( ! v ) { miss = append( miss, spec.Field ) }
        case []ReactorEvent: {}
        default: panic( libErrorf( "Unhandled val state: %T", v ) )
        }
    }
    if len( miss ) > 0 {
        return NewMissingFieldsError( foc.encl.GetPath(), miss )
    }
    return nil
}

func ( foc *fieldOrderCtx ) endStruct( 
    ee EndEvent, rep ReactorEventProcessor ) error {
    foc.checkAccNil( "at end of struct" )
    if err := foc.checkRequiredFields(); err != nil { return err }
    for i, e := foc.idx, len( foc.ord ); i < e; i++ {
        fld := foc.ord[ i ].Field
        vs := foc.valStates.Get( fld )
        if acc, ok := vs.( []ReactorEvent ); ok {
            if err := foc.feedValue( fld, acc, rep ); err != nil { return err }
        }
    }
    return foc.sendEvent( ee, rep )
}

func ( fo *FieldOrderReactor ) peek() *fieldOrderCtx {
    if fo.stack.Len() == 0 { 
        panic( libErrorf( "field order stack is empty" ) ) 
    }
    return fo.stack.Front().Value.( *fieldOrderCtx )
}

func ( fo *FieldOrderReactor ) pop() *fieldOrderCtx {
    res := fo.peek()
    fo.stack.Remove( fo.stack.Front() )
    return res
}

var emptyFieldOrder = FieldOrder( []FieldOrderSpecification{} )

func ( fo *FieldOrderReactor ) startStruct( qn *QualifiedTypeName ) {
    ord := fo.fog.FieldOrderFor( qn )
    if ord == nil { ord = emptyFieldOrder }
    valStates := NewIdentifierMap()
    for _, spec := range ord { valStates.Put( spec.Field, false ) }
    foc := &fieldOrderCtx{ encl: fo, ord: ord, valStates: valStates }
    if fo.stack.Len() > 0 { foc.parent = fo.peek() }
    // since parent won't see this event directly, we account for it here:
    if foc.parent != nil { foc.parent.valDepth++ }
    fo.stack.PushFront( foc )
}

func ( fo *FieldOrderReactor ) ProcessEvent(
    ev ReactorEvent, rep ReactorEventProcessor ) error {
    if ss, ok := ev.( StructStartEvent ); ok { fo.startStruct( ss.Type ) }
    if fo.stack.Len() == 0 { return rep.ProcessEvent( ev ) }
    foc := fo.peek()
    if ee, ok := ev.( EndEvent ); ok && foc.valDepth == 0 {
        defer fo.pop()
        if err := foc.endStruct( ee, rep ); err != nil { return err }
        if foc.parent == nil { return nil }
        return foc.parent.completeEvent( ev, rep )
    }
    return foc.processEvent( ev, rep )
}

func ( fo *FieldOrderReactor ) GetPath() objpath.PathNode {
    res := fo.pg.GetPath()
    if fo.stack.Len() == 0 { return res }
    for e := fo.stack.Front(); e != nil; e = e.Next() {
        foc := e.Value.( *fieldOrderCtx )
        if foc.epRct != nil { return foc.appendFeedPath( res ) }
    }
    return res
}

type ServiceRequestReactorInterface interface {
    Namespace( ns *Namespace, pg PathGetter ) error
    Service( svc *Identifier, pg PathGetter ) error
    Operation( op *Identifier, pg PathGetter ) error
    GetAuthenticationProcessor( pg PathGetter ) ( ReactorEventProcessor, error )
    GetParametersProcessor( pg PathGetter ) ( ReactorEventProcessor, error )
}

type requestFieldType int

const (
    reqFieldNone = requestFieldType( iota )
    reqFieldNs
    reqFieldSvc
    reqFieldOp
    reqFieldAuth
    reqFieldParams
)

type ServiceRequestReactor struct {

    iface ServiceRequestReactorInterface

    evProc ReactorEventProcessor

    // 0: before StartStruct{ QnameServiceRequest } and after final EndEvent
    // 1: when reading a service request field (namespace, service, etc)
    // > 1: accumulating some nested value for 'parameters' or 'authentication' 
    depth int 

    fld requestFieldType

    pg PathGetter

    hadParams bool // true if the input contained explicit params
    paramsSynth bool // true when we are synthesizing empty params
}

func NewServiceRequestReactor( 
    iface ServiceRequestReactorInterface ) *ServiceRequestReactor {
    return &ServiceRequestReactor{ iface: iface }
}

func ( sr *ServiceRequestReactor ) GetPath() objpath.PathNode {
    res := sr.pg.GetPath()
    if sr.paramsSynth { res = objpath.Descend( res, IdParameters ) }
    return res
}

func ( sr *ServiceRequestReactor ) updateEvProc( ev ReactorEvent ) {
    if _, ok := ev.( FieldStartEvent ); ok { return }
    switch ev.( type ) {
    case StructStartEvent, ListStartEvent, MapStartEvent: sr.depth++
    case EndEvent: sr.depth--
    }
    if sr.depth == 1 { sr.evProc, sr.fld = nil, reqFieldNone } 
}

type svcReqCastIface int

func ( c svcReqCastIface ) InferStructFor( qn *QualifiedTypeName ) bool {
    return qn.Equals( QnameServiceRequest )
}

type svcReqFieldTyper int

func ( t svcReqFieldTyper ) FieldTypeOf( 
    fld *Identifier, pg PathGetter ) ( TypeReference, error ) {
    if fld.Equals( IdParameters ) { return TypeSymbolMap, nil }
    return TypeValue, nil
}

func ( c svcReqCastIface ) FieldTyperFor( 
    qn *QualifiedTypeName, pg PathGetter ) ( FieldTyper, error ) {
    if qn.Equals( QnameServiceRequest ) { return svcReqFieldTyper( 1 ), nil }
    return nil, nil
}

func ( c svcReqCastIface ) CastAtomic( 
    in Value, at *AtomicTypeReference, pg PathGetter ) ( Value, error, bool ) {
    return nil, nil, false
}

type svcReqFieldOrderGetter int

func ( g svcReqFieldOrderGetter ) FieldOrderFor( 
    qn *QualifiedTypeName ) FieldOrder {
    if qn.Equals( QnameServiceRequest ) { return svcReqFieldOrder }
    return nil
}

func ( sr *ServiceRequestReactor ) Init( rpi *ReactorPipelineInit ) {
    EnsureStructuralReactor( rpi ) 
    cr := NewCastReactor( TypeServiceRequest, svcReqCastIface( 1 ), nil )
    rpi.AddPipelineProcessor( cr )
    fo := NewFieldOrderReactor( svcReqFieldOrderGetter( 1 ) )
    rpi.AddPipelineProcessor( fo )
    sr.pg = LastPathGetter( rpi )
}

func ( sr *ServiceRequestReactor ) invalidValueErr( desc string ) error {
    return NewValueCastErrorf( sr.GetPath(), "invalid value: %s", desc )
}

func ( sr *ServiceRequestReactor ) startStruct( ev StructStartEvent ) error {
    if sr.fld == reqFieldNone { // we're at the top of the request
        if ev.Type.Equals( QnameServiceRequest ) { return nil }
        // panic because upstream cast should have checked already
        panic( libErrorf( "Unexpected service request type: %s", ev.Type ) )
    }
    return sr.invalidValueErr( ev.Type.ExternalForm() )
}

func ( sr *ServiceRequestReactor ) startField( 
    fs FieldStartEvent ) ( err error ) {
    if sr.fld != reqFieldNone {
        panic( libErrorf( 
            "Saw field start '%s' while sr.fld is %d", fs.Field, sr.fld ) )
    }
    switch fld := fs.Field; {
    case fld.Equals( IdNamespace ): sr.fld = reqFieldNs
    case fld.Equals( IdService ): sr.fld = reqFieldSvc
    case fld.Equals( IdOperation ): sr.fld = reqFieldOp
    case fld.Equals( IdAuthentication ): 
        sr.fld = reqFieldAuth
        sr.evProc, err = sr.iface.GetAuthenticationProcessor( sr )
        if err != nil { return }
    case fld.Equals( IdParameters ): 
        sr.fld = reqFieldParams
        sr.evProc, err = sr.iface.GetParametersProcessor( sr )
        if err != nil { return }
        sr.hadParams = true
    }
    if sr.fld == reqFieldNone {
        return NewUnrecognizedFieldError( sr.GetPath(), fs.Field )
    }
    return nil
}

func ( sr *ServiceRequestReactor ) getFieldValueForString(
    s string, reqFld requestFieldType ) ( res interface{}, err error ) {
    switch reqFld {
    case reqFieldNs: res, err = ParseNamespace( s )
    case reqFieldSvc, reqFieldOp: res, err = ParseIdentifier( s )
    default:
        panic( libErrorf( "Unhandled req fld type for string: %d", reqFld ) )
    }
    if err != nil { err = NewValueCastError( sr.GetPath(), err.Error() ) }
    return
}

func ( sr *ServiceRequestReactor ) getFieldValueForBuffer(
    buf []byte, reqFld requestFieldType ) ( res interface{}, err error ) {
    bin := NewReader( bytes.NewReader( buf ) )
    switch reqFld {
    case reqFieldNs: res, err = bin.ReadNamespace()
    case reqFieldSvc, reqFieldOp: res, err = bin.ReadIdentifier()
    default:
        panic( libErrorf( "Unhandled req fld type for buffer: %d", reqFld ) )
    }
    if err != nil { err = NewValueCastError( sr.GetPath(), err.Error() ) }
    return
}

func ( sr *ServiceRequestReactor ) getFieldValue( 
    val Value, reqFld requestFieldType ) ( interface{}, error ) {
    switch v := val.( type ) {
    case String: return sr.getFieldValueForString( string( v ), reqFld )
    case Buffer: return sr.getFieldValueForBuffer( []byte( v ), reqFld )
    }
    return nil, sr.invalidValueErr( TypeOf( val ).ExternalForm() )
}

func ( sr *ServiceRequestReactor ) namespace( val Value ) error {
    ns, err := sr.getFieldValue( val, reqFieldNs )
    if err == nil { return sr.iface.Namespace( ns.( *Namespace ), sr ) }
    return err
}

func ( sr *ServiceRequestReactor ) readIdent( 
    val Value, reqFld requestFieldType ) error {
    v2, err := sr.getFieldValue( val, reqFld )
    if err == nil {
        id := v2.( *Identifier )
        switch reqFld {
        case reqFieldSvc: return sr.iface.Service( id, sr )
        case reqFieldOp: return sr.iface.Operation( id, sr )
        default: panic( libErrorf( "Unhandled req fld type: %d", reqFld ) )
        }
    }
    return err
}

func ( sr *ServiceRequestReactor ) value( val Value ) error {
    defer func() { sr.fld = reqFieldNone }()
    switch sr.fld {
    case reqFieldNs: return sr.namespace( val )
    case reqFieldSvc, reqFieldOp: return sr.readIdent( val, sr.fld )
    }
    panic( libErrorf( "Unhandled req field type: %d", sr.fld ) )
}

func ( sr *ServiceRequestReactor ) end() error {
    if ! sr.hadParams {
        defer func() { sr.paramsSynth = false }()
        sr.paramsSynth = true
        ep, err := sr.iface.GetParametersProcessor( sr );
        if err != nil { return err }
        if err := ep.ProcessEvent( EvMapStart ); err != nil { return err }
        if err := ep.ProcessEvent( EvEnd ); err != nil { return err }
    }
    return nil
}

func ( sr *ServiceRequestReactor ) ProcessEvent( ev ReactorEvent ) error {
    defer sr.updateEvProc( ev )
    if sr.evProc != nil { return sr.evProc.ProcessEvent( ev ) }
    switch v := ev.( type ) {
    case FieldStartEvent: return sr.startField( v )
    case StructStartEvent: return sr.startStruct( v )
    case ValueEvent: return sr.value( v.Val )
    case ListStartEvent: 
        return sr.invalidValueErr( TypeOpaqueList.ExternalForm() )
    case MapStartEvent: 
        return sr.invalidValueErr( TypeSymbolMap.ExternalForm() )
    case EndEvent: return sr.end()
    default: panic( libErrorf( "Unhandled event: %T", v ) )
    }
    return nil
}

type ServiceResponseReactorInterface interface {
    GetResultProcessor( pg PathGetter ) ( ReactorEventProcessor, error )
    GetErrorProcessor( pg PathGetter ) ( ReactorEventProcessor, error )
}

type ServiceResponseReactor struct {

    iface ServiceResponseReactorInterface

    pg PathGetter

    evProc ReactorEventProcessor
    depth int
    gotEvProcVal bool
}

func ( sr *ServiceResponseReactor ) getPath() objpath.PathNode {
    return sr.pg.GetPath()
}

func NewServiceResponseReactor( 
    iface ServiceResponseReactorInterface ) *ServiceResponseReactor {
    return &ServiceResponseReactor{ iface: iface }
}

type svcRespCastIface int

func ( i svcRespCastIface ) InferStructFor( qn *QualifiedTypeName ) bool {
    return qn.Equals( QnameServiceResponse )
}

func ( i svcRespCastIface ) FieldTyperFor( 
    qn *QualifiedTypeName, pg PathGetter ) ( FieldTyper, error ) {
    return valueFieldTyper( 1 ), nil
}

func ( i svcRespCastIface ) CastAtomic(
    in Value, at *AtomicTypeReference, pg PathGetter ) ( Value, error, bool ) {
    return nil, nil, false
}

func ( sr *ServiceResponseReactor ) Init( rpi *ReactorPipelineInit ) {
    EnsureStructuralReactor( rpi )
    cr := NewCastReactor( TypeServiceResponse, svcRespCastIface( 1 ), nil )
    rpi.AddPipelineProcessor( cr )
    sr.pg = LastPathGetter( rpi )
}

func ( sr *ServiceResponseReactor ) GetPath() objpath.PathNode {
    return sr.pg.GetPath()
}

func ( sr *ServiceResponseReactor ) updateEvProc( ev ReactorEvent ) {
    if _, ok := ev.( FieldStartEvent ); ok { return }
    switch ev.( type ) {
    case StructStartEvent, MapStartEvent, ListStartEvent: sr.depth++
    case EndEvent: sr.depth--
    }
    if sr.depth == 1 { 
        if sr.evProc != nil { sr.gotEvProcVal, sr.evProc = true, nil }
    }
}

// Note that the error path uses Parent() since we'll be positioned on the field
// (result/error) that is the second value, but the error is really at the
// response level itself
func ( sr *ServiceResponseReactor ) sendEvProcEvent( ev ReactorEvent ) error {
    isErr := sr.gotEvProcVal
    if isErr {
        if ve, ok := ev.( ValueEvent ); ok {
            if _, isNull := ve.Val.( *Null ); isNull { isErr = false }
        }
    }
    if isErr {
        msg := "response has both a result and an error value"
        return NewValueCastError( sr.getPath().Parent(), msg )
    }
    return sr.evProc.ProcessEvent( ev )
}

func ( sr *ServiceResponseReactor ) startStruct( t *QualifiedTypeName ) error {
    if t.Equals( QnameServiceResponse ) { return nil }
    panic( libErrorf( "Got unexpected (toplevel) struct type: %s", t ) )
}

func ( sr *ServiceResponseReactor ) startField( fld *Identifier ) error {
    var err error
    switch {
    case fld.Equals( IdResult ): 
        sr.evProc, err = sr.iface.GetResultProcessor( sr )
    case fld.Equals( IdError ): 
        sr.evProc, err = sr.iface.GetErrorProcessor( sr )
    default: return NewUnrecognizedFieldError( sr.getPath().Parent(), fld )
    }
    return err
}

func ( sr *ServiceResponseReactor ) ProcessEvent( ev ReactorEvent ) error {
    defer sr.updateEvProc( ev )
    if sr.evProc != nil { return sr.sendEvProcEvent( ev ) }
    switch v := ev.( type ) {
    case StructStartEvent: return sr.startStruct( v.Type )
    case FieldStartEvent: return sr.startField( v.Field )
    case EndEvent: return nil
    }
    panic( libErrorf( "Saw event %v (%T) while evProc == nil", ev, ev ) )
}

type DebugLogger interface {
    Logf( tmpl string, args ...interface{} )
}

type DebugLoggerFunc func( string, ...interface{} )

func ( f DebugLoggerFunc ) Logf( tmpl string, args ...interface{} ) {
    f( tmpl, args... )
}

type debugReactor struct { 
    l DebugLogger 
    key ReactorKey
    pg PathGetter
}

func ( dr *debugReactor ) Init( rpi *ReactorPipelineInit ) {
    dr.pg = LastPathGetter( rpi )
}

func ( dr *debugReactor ) ProcessEvent( ev ReactorEvent ) error {
    var path string
    if dr.pg == nil { 
        path = "< path unknown >" 
    } else { path = FormatIdPath( dr.pg.GetPath() ) }
    dr.l.Logf( "%s: %v (%T)", path, ev, ev )
    return nil
}

func NewDebugReactor( l DebugLogger ) ReactorEventProcessor { 
    return &debugReactor{ l: l }
}
