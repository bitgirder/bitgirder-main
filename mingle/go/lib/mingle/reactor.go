package mingle

import (
    "container/list"
    "fmt"
    "bitgirder/objpath"
    "bytes"
    "strings"
    "log"
    "bitgirder/stack"
)

type ReactorError struct { msg string }

func ( e *ReactorError ) Error() string { return e.msg }

func rctError( msg string ) *ReactorError { return &ReactorError{ msg } }

func rctErrorf( tmpl string, args ...interface{} ) *ReactorError {
    return rctError( fmt.Sprintf( tmpl, args... ) )
}

type ReactorEvent interface {

    // may be the empty path
    GetPath() objpath.PathNode

    SetPath( path objpath.PathNode )
}

type reactorEventImpl struct {
    path objpath.PathNode
}

func ( ri *reactorEventImpl ) GetPath() objpath.PathNode { return ri.path }

func ( ri *reactorEventImpl ) SetPath( path objpath.PathNode ) { 
    ri.path = path 
}

type ValueEvent struct { 
    *reactorEventImpl
    Val Value 
}

func NewValueEvent( val Value ) ValueEvent { 
    return ValueEvent{ Val: val, reactorEventImpl: &reactorEventImpl{} } 
}

type StructStartEvent struct { 
    *reactorEventImpl
    Type *QualifiedTypeName 
}

func NewStructStartEvent( typ *QualifiedTypeName ) StructStartEvent {
    return StructStartEvent{ Type: typ, reactorEventImpl: &reactorEventImpl{} }
}

func isStructStart( ev ReactorEvent ) bool {
    _, ok := ev.( StructStartEvent )
    return ok
}

type MapStartEvent struct {
    *reactorEventImpl
}

func NewMapStartEvent() MapStartEvent {
    return MapStartEvent{ reactorEventImpl: &reactorEventImpl{} }
}

type FieldStartEvent struct { 
    *reactorEventImpl
    Field *Identifier 
}

func NewFieldStartEvent( fld *Identifier ) FieldStartEvent {
    return FieldStartEvent{ Field: fld, reactorEventImpl: &reactorEventImpl{} }
}

type ListStartEvent struct {
    *reactorEventImpl
}

func NewListStartEvent() ListStartEvent {
    return ListStartEvent{ reactorEventImpl: &reactorEventImpl{} }
}

type EndEvent struct {
    *reactorEventImpl
}

func NewEndEvent() EndEvent {
    return EndEvent{ reactorEventImpl: &reactorEventImpl{} }
}

func EventToString( ev ReactorEvent ) string {
    pairs := [][]string{ { "type", fmt.Sprintf( "%T", ev ) } }
    switch v := ev.( type ) {
    case ValueEvent: 
        pairs = append( pairs, []string{ "value", QuoteValue( v.Val ) } )
    case StructStartEvent:
        pairs = append( pairs, []string{ "type", v.Type.ExternalForm() } )
    case FieldStartEvent:
        pairs = append( pairs, []string{ "field", v.Field.ExternalForm() } )
    }
    if p := ev.GetPath(); p != nil {
        pairs = append( pairs, []string{ "path", FormatIdPath( p ) } )
    }
    elts := make( []string, len( pairs ) )
    for i, pair := range pairs { elts[ i ] = strings.Join( pair, " = " ) }
    return fmt.Sprintf( "[ %s ]", strings.Join( elts, ", " ) )
}

func CopyEvent( ev ReactorEvent, withPath bool ) ReactorEvent {
    var res ReactorEvent
    switch v := ev.( type ) {
    case ValueEvent: res = NewValueEvent( v.Val )
    case ListStartEvent: res = NewListStartEvent()
    case MapStartEvent: res = NewMapStartEvent()
    case StructStartEvent: res = NewStructStartEvent( v.Type )
    case FieldStartEvent: res = NewFieldStartEvent( v.Field )
    case EndEvent: res = NewEndEvent()
    default: panic( libErrorf( "unhandled copy target: %T", ev ) )
    }
    if withPath { res.SetPath( ev.GetPath() ) }
    return res
}

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
        if err := rep.ProcessEvent( NewMapStartEvent() ); err != nil { return err }
    }
    err := m.EachPairError( func( fld *Identifier, val Value ) error {
        ev := NewFieldStartEvent( fld )
        if err := rep.ProcessEvent( ev ); err != nil { return err }
        return VisitValue( val, rep )
    })
    if err != nil { return err }
    return rep.ProcessEvent( NewEndEvent() )
}

func visitStruct( ms *Struct, rep ReactorEventProcessor ) error {
    ev := NewStructStartEvent( ms.Type )
    if err := rep.ProcessEvent( ev ); err != nil { return err }
    return visitSymbolMap( ms.Fields, false, rep )
}

func visitList( ml *List, rep ReactorEventProcessor ) error {
    if err := rep.ProcessEvent( NewListStartEvent() ); err != nil { return err }
    for _, val := range ml.Values() {
        if err := VisitValue( val, rep ); err != nil { return err }
    }
    return rep.ProcessEvent( NewEndEvent() )
}

func VisitValue( mv Value, rep ReactorEventProcessor ) error {
    switch v := mv.( type ) {
    case *Struct: return visitStruct( v, rep )
    case *SymbolMap: return visitSymbolMap( v, true, rep )
    case *List: return visitList( v, rep )
    }
    return rep.ProcessEvent( NewValueEvent( mv ) )
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

func EnsurePathSettingProcessor( rpi *ReactorPipelineInit ) {
    ok := false
    rpi.VisitPredecessors( func( p interface{} ) {
        if ok { return }
        _, ok = p.( *PathSettingProcessor )
    })
    if ! ok { rpi.AddPipelineProcessor( NewPathSettingProcessor() ) }
}

type endType int

const (
    endTypeList = endType( iota )
    endTypeMap
    endTypeStruct
    endTypeField
)


type PathSettingProcessor struct {
    endTypes *stack.Stack
    awaitingList0 bool
    path objpath.PathNode
    skipStructureCheck bool
}

func NewPathSettingProcessor() *PathSettingProcessor {
    return &PathSettingProcessor{ endTypes: stack.NewStack() }
}

func ( proc *PathSettingProcessor ) SetStartPath( p objpath.PathNode ) {
    if p == nil { return }
    proc.path = objpath.CopyOf( p )
}

func ( proc *PathSettingProcessor ) Init( rpi *ReactorPipelineInit ) {
    if ! proc.skipStructureCheck { EnsureStructuralReactor( rpi ) }
}

func ( proc *PathSettingProcessor ) pathPop() {
    if proc.path != nil { proc.path = proc.path.Parent() }
}

func ( proc *PathSettingProcessor ) updateList() {
    if proc.awaitingList0 {
        if proc.path == nil { 
            proc.path = objpath.RootedAtList() 
        } else {
            proc.path = proc.path.StartList()
        }
        proc.awaitingList0 = false
    } else {
        if lp, ok := proc.path.( *objpath.ListNode ); ok { lp.Increment() }
    }
}

func ( proc *PathSettingProcessor ) prepareValue() { proc.updateList() }

func ( proc *PathSettingProcessor ) prepareListStart() {
    proc.prepareValue() // this list may be a new value in a nested list
    proc.endTypes.Push( endTypeList )
    proc.awaitingList0 = true
}

func ( proc *PathSettingProcessor ) prepareStructure( et endType ) {
    proc.prepareValue()
    proc.endTypes.Push( et )
}

func ( proc *PathSettingProcessor ) prepareStartField( f *Identifier ) {
    proc.endTypes.Push( endTypeField )
    if proc.path == nil {
        proc.path = objpath.RootedAt( f )
    } else {
        proc.path = proc.path.Descend( f )
    }
}

func ( proc *PathSettingProcessor ) prepareEnd() {
    if top := proc.endTypes.Peek(); top != nil {
        if top.( endType ) == endTypeList { proc.pathPop() }
    }
}

func ( proc *PathSettingProcessor ) prepareEvent( ev ReactorEvent ) {
    switch v := ev.( type ) {
    case ValueEvent: proc.prepareValue()
    case ListStartEvent: proc.prepareListStart()
    case MapStartEvent: proc.prepareStructure( endTypeMap )
    case StructStartEvent: proc.prepareStructure( endTypeStruct )
    case FieldStartEvent: proc.prepareStartField( v.Field )
    case EndEvent: proc.prepareEnd()
    }
    if proc.path != nil { ev.SetPath( proc.path ) }
}

func ( proc *PathSettingProcessor ) processedValue() {
    if top := proc.endTypes.Peek(); top != nil {
        if top.( endType ) == endTypeField { 
            proc.endTypes.Pop()
            proc.pathPop() 
        }
    }
}

func ( proc *PathSettingProcessor ) processedEnd() {
    et := proc.endTypes.Pop().( endType )
    switch et {
    case endTypeList, endTypeStruct, endTypeMap: proc.processedValue()
    default: panic( libErrorf( "unexpected end type for END: %d", et ) )
    }
}

func ( proc *PathSettingProcessor ) eventProcessed( ev ReactorEvent ) {
    switch ev.( type ) {
    case ValueEvent: proc.processedValue()
    case EndEvent: proc.processedEnd()
    }
}

func ( proc *PathSettingProcessor ) ProcessEvent(
    ev ReactorEvent,
    rep ReactorEventProcessor,
) error {
    proc.prepareEvent( ev )
    if err := rep.ProcessEvent( ev ); err != nil { return err }
    proc.eventProcessed( ev )
    return nil
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

type FieldTyper interface {

    // path will be positioned to the map/struct containing fld, but will not
    // itself include fld
    FieldTypeFor( 
        fld *Identifier, path objpath.PathNode ) ( TypeReference, error )
}

type CastInterface interface {

    // Called when start of a symbol map arrives when an atomic type having name
    // qn (or a nullable or list type containing such an atomic type) is the
    // cast reactor's expected type. Returning true from this function will
    // cause the cast reactor to treat the symbol map start as if it were a
    // struct start with atomic type qn. 
    //
    // One motivating use for this is for cast reactors that react to inputs
    // conforming to a known schema and receive unadorned maps for structured
    // field values and wish to cause further processing to behave as if the
    // struct were explicitly signalled in the input
    InferStructFor( qn *QualifiedTypeName ) bool

    FieldTyperFor( 
        qn *QualifiedTypeName, path objpath.PathNode ) ( FieldTyper, error )

    CastAtomic( 
        in Value, 
        at *AtomicTypeReference, 
        path objpath.PathNode ) ( Value, error, bool )
}

type valueFieldTyper int

func ( vt valueFieldTyper ) FieldTypeFor( 
    fld *Identifier, path objpath.PathNode ) ( TypeReference, error ) {
    return TypeNullableValue, nil
}

type castInterfaceDefault int

func ( i castInterfaceDefault ) FieldTyperFor( 
    qn *QualifiedTypeName, path objpath.PathNode ) ( FieldTyper, error ) {
    return valueFieldTyper( 1 ), nil
}

func ( i castInterfaceDefault ) InferStructFor( at *QualifiedTypeName ) bool {
    return false
}

func ( i castInterfaceDefault ) CastAtomic( 
    v Value, 
    at *AtomicTypeReference, 
    path objpath.PathNode ) ( Value, error, bool ) {

    return nil, nil, false
}

type CastReactor struct {
    iface CastInterface
    stack *stack.Stack
}

func NewCastReactor( expct TypeReference, iface CastInterface ) *CastReactor {
    res := &CastReactor{ stack: stack.NewStack(), iface: iface }
    res.stack.Push( expct )
    return res
}

func NewDefaultCastReactor( expct TypeReference ) *CastReactor {
    return NewCastReactor( expct, castInterfaceDefault( 1 ) )
}

func ( cr *CastReactor ) Init( rpi *ReactorPipelineInit ) {
    EnsureStructuralReactor( rpi )
    EnsurePathSettingProcessor( rpi )
}

type listCast struct {
    sawValues bool
    lt *ListTypeReference
    startPath objpath.PathNode
}

func ( cr *CastReactor ) errStackUnrecognized() error {
    return libErrorf( "unrecognized stack element: %T", cr.stack.Peek() )
}

func ( cr *CastReactor ) processAtomicValue(
    ve ValueEvent,
    at *AtomicTypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    mv, err, ok := cr.iface.CastAtomic( ve.Val, at, ve.GetPath() )

    if ! ok { 
        mv, err = castAtomicWithCallType( ve.Val, at, callTyp, ve.GetPath() ) 
    }

    if err != nil { return err }

    ve.Val = mv
    return next.ProcessEvent( ve )
}

func ( cr *CastReactor ) processNullableValue(
    ve ValueEvent,
    nt *NullableTypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    if _, ok := ve.Val.( *Null ); ok { return next.ProcessEvent( ve ) }
    return cr.processValueWithType( ve, nt.Type, callTyp, next )
}

func ( cr *CastReactor ) processValueWithType(
    ve ValueEvent,
    typ TypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    switch v := typ.( type ) {
    case *AtomicTypeReference: 
        return cr.processAtomicValue( ve, v, callTyp, next )
    case *NullableTypeReference:
        return cr.processNullableValue( ve, v, callTyp, next )
    case *ListTypeReference:
        return NewTypeCastErrorValue( callTyp, ve.Val, ve.GetPath() )
    }
    panic( libErrorf( "unhandled type: %T", typ ) )
}

func ( cr *CastReactor ) processValue( 
    ve ValueEvent, next ReactorEventProcessor ) error {

    switch v := cr.stack.Peek().( type ) {
    case TypeReference: 
        cr.stack.Pop()
        return cr.processValueWithType( ve, v, v, next )
    case *listCast:
        v.sawValues = true
        typ := v.lt.ElementType
        return cr.processValueWithType( ve, typ, typ, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) implStartMap(
    ev ReactorEvent, ft FieldTyper, next ReactorEventProcessor ) error {

    cr.stack.Push( ft )
    log.Printf( "pushed field typer: %v", ft )
    return next.ProcessEvent( ev )
}

func ( cr *CastReactor ) completeStartStruct(
    ss StructStartEvent, next ReactorEventProcessor ) error {

    ft, err := cr.iface.FieldTyperFor( ss.Type, ss.GetPath() )
    if err != nil { return err }

    if ft == nil { ft = valueFieldTyper( 1 ) }
    return cr.implStartMap( ss, ft, next )
}

func ( cr *CastReactor ) inferStructForMap(
    me MapStartEvent,
    at *AtomicTypeReference,
    next ReactorEventProcessor ) ( error, bool ) {

    qn, ok := at.Name.( *QualifiedTypeName )
    if ! ok { return nil, false }

    if ! cr.iface.InferStructFor( qn ) { return nil, false }

    ev := NewStructStartEvent( qn )
    ev.SetPath( me.GetPath() )

    return cr.completeStartStruct( ev, next ), true
}

func ( cr *CastReactor ) processStartMapWithAtomicType(
    me MapStartEvent,
    at *AtomicTypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    if at.Equals( TypeSymbolMap ) || at.Equals( TypeValue ) {
        return cr.implStartMap( me, valueFieldTyper( 1 ), next )
    }

    if err, ok := cr.inferStructForMap( me, at, next ); ok { return err }

    return NewTypeCastError( callTyp, at, me.GetPath() )
}

func ( cr *CastReactor ) processStartMapWithType(
    me MapStartEvent, 
    typ TypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    switch v := typ.( type ) {
    case *AtomicTypeReference:
        return cr.processStartMapWithAtomicType( me, v, callTyp, next )
    case *NullableTypeReference:
        return cr.processStartMapWithType( me, v.Type, callTyp, next )
    }
    return NewTypeCastError( callTyp, typ, me.GetPath() )
}

func ( cr *CastReactor ) processStartMap(
    me MapStartEvent, next ReactorEventProcessor ) error {
    
    switch v := cr.stack.Peek().( type ) {
    case TypeReference: 
        cr.stack.Pop()
        return cr.processStartMapWithType( me, v, v, next )
    case *listCast:
        v.sawValues = true
        typ := v.lt.ElementType
        return cr.processStartMapWithType( me, typ, typ, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) processFieldStart(
    fs FieldStartEvent, next ReactorEventProcessor ) error {

    ft := cr.stack.Peek().( FieldTyper )
    
    typ, err := ft.FieldTypeFor( fs.Field, fs.GetPath().Parent() )
    if err != nil { return err }

    cr.stack.Push( typ )
    return next.ProcessEvent( fs )
}

func ( cr *CastReactor ) processEnd(
    ee EndEvent, next ReactorEventProcessor ) error {

    switch v := cr.stack.Peek().( type ) {
    case *listCast:
        cr.stack.Pop()
        if ! ( v.sawValues || v.lt.AllowsEmpty ) {
            return NewValueCastError( v.startPath, "List is empty" )
        }
    case FieldTyper: cr.stack.Pop()
    }

    return next.ProcessEvent( ee )
}

func ( cr *CastReactor ) processStructStartWithAtomicType(
    ss StructStartEvent,
    at *AtomicTypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    if at.Equals( TypeSymbolMap ) {
        me := NewMapStartEvent()
        me.SetPath( ss.GetPath() )
        return cr.processStartMapWithAtomicType( me, at, callTyp, next )
    }

    if at.Name.Equals( ss.Type ) || at.Equals( TypeValue ) {
        return cr.completeStartStruct( ss, next )
    }

    failTyp := &AtomicTypeReference{ Name: ss.Type }
    return NewTypeCastError( callTyp, failTyp, ss.GetPath() )
}

func ( cr *CastReactor ) processStructStartWithType(
    ss StructStartEvent,
    typ TypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    switch v := typ.( type ) {
    case *AtomicTypeReference:
        return cr.processStructStartWithAtomicType( ss, v, callTyp, next )
    case *NullableTypeReference:
        return cr.processStructStartWithType( ss, v.Type, callTyp, next )
    }
    return NewTypeCastError( typ, callTyp, ss.GetPath() )
}

func ( cr *CastReactor ) processStructStart(
    ss StructStartEvent, next ReactorEventProcessor ) error {

    switch v := cr.stack.Peek().( type ) {
    case TypeReference:
        cr.stack.Pop()
        return cr.processStructStartWithType( ss, v, v, next )
    case *listCast:
        v.sawValues = true
        typ := v.lt.ElementType
        return cr.processStructStartWithType( ss, typ, typ, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) processListStartWithAtomicType(
    le ListStartEvent,
    at *AtomicTypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    if at.Equals( TypeValue ) {
        return cr.processListStartWithType( le, TypeOpaqueList, callTyp, next )
    }

    return NewTypeCastError( callTyp, TypeOpaqueList, le.GetPath() )
}

func ( cr *CastReactor ) processListStartWithType(
    le ListStartEvent,
    typ TypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    switch v := typ.( type ) {
    case *AtomicTypeReference:
        return cr.processListStartWithAtomicType( le, v, callTyp, next )
    case *ListTypeReference:
        sp := objpath.CopyOf( le.GetPath() )
        cr.stack.Push( &listCast{ lt: v, startPath: sp } )
        return next.ProcessEvent( le )
    case *NullableTypeReference:
        return cr.processListStartWithType( le, v.Type, callTyp, next )
    }
    panic( libErrorf( "unhandled type: %T", typ ) )
}

func ( cr *CastReactor ) processListStart( 
    le ListStartEvent, next ReactorEventProcessor ) error {

    switch v := cr.stack.Peek().( type ) {
    case TypeReference:
        cr.stack.Pop()
        return cr.processListStartWithType( le, v, v, next )
    case *listCast:
        v.sawValues = true
        return cr.processListStartWithType( le, v.lt.ElementType, v.lt, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) ProcessEvent(
    ev ReactorEvent, next ReactorEventProcessor ) error {

    switch v := ev.( type ) {
    case ValueEvent: return cr.processValue( v, next )
    case MapStartEvent: return cr.processStartMap( v, next )
    case FieldStartEvent: return cr.processFieldStart( v, next )
    case StructStartEvent: return cr.processStructStart( v, next )
    case ListStartEvent: return cr.processListStart( v, next )
    case EndEvent: return cr.processEnd( v, next )
    }
    panic( libErrorf( "unhandled event: %T", ev ) )
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
//
// returning nil is allowed and is equivalent to returning the empty ordering
type FieldOrderGetter interface {
    FieldOrderFor( qn *QualifiedTypeName ) FieldOrder
}

// Reorders events for selected struct types according to an order determined by
// a FieldOrderGetter.
type FieldOrderReactor struct {
    fog FieldOrderGetter
    stack *stack.Stack
}

func NewFieldOrderReactor( fog FieldOrderGetter ) *FieldOrderReactor {
    return &FieldOrderReactor{ fog: fog, stack: stack.NewStack() }
}

func ( fo *FieldOrderReactor ) Init( rpi *ReactorPipelineInit ) {
    EnsureStructuralReactor( rpi )
    EnsurePathSettingProcessor( rpi )
}

type structOrderFieldState struct {
    spec FieldOrderSpecification
    seen bool
    acc []ReactorEvent
}

func ( s *structOrderFieldState ) ProcessEvent( ev ReactorEvent ) error {
    s.acc = append( s.acc, CopyEvent( ev, false ) )
    return nil
}

type structOrderProcessor struct {
    ord FieldOrder
    next ReactorEventProcessor
    startPath objpath.PathNode
    fieldQueue *list.List
    states *IdentifierMap
    cur *structOrderFieldState
}

func ( sp *structOrderProcessor ) fieldReactor() ReactorEventProcessor {
    if sp.cur == nil || sp.cur.acc == nil { return sp.next }
    return sp.cur
}

func ( sp *structOrderProcessor ) processStructStart( ev ReactorEvent ) error {
    if p := ev.GetPath(); p != nil { sp.startPath = objpath.CopyOf( p ) }
    sp.states = NewIdentifierMap()
    sp.fieldQueue = &list.List{}
    for _, spec := range sp.ord {
        state := &structOrderFieldState{ spec: spec }
        sp.states.Put( spec.Field, state )
        sp.fieldQueue.PushBack( state )
    }
    return sp.next.ProcessEvent( ev )
}

func ( sp *structOrderProcessor ) shouldAccumulate( 
    s *structOrderFieldState ) bool {

    if s == nil { return false }
    if sp.fieldQueue.Len() == 0 { return false }
    frnt := sp.fieldQueue.Front()
    state := frnt.Value.( *structOrderFieldState )
    if state.spec.Field.Equals( s.spec.Field ) {
        sp.fieldQueue.Remove( frnt )
        return false
    }
    return true
}

func ( sp *structOrderProcessor ) processFieldStart( 
    ev FieldStartEvent ) error {

    if sp.cur != nil { 
        panic( libErrorf( "saw field '%s' while processing '%s'", 
            ev.Field, sp.cur.spec.Field ) )
    }
    if v, ok := sp.states.GetOk( ev.Field ); ok {
        sp.cur = v.( *structOrderFieldState )
    } else { sp.cur = nil }
    if sp.cur != nil { sp.cur.seen = true }
    if sp.shouldAccumulate( sp.cur ) {
        sp.cur.acc = make( []ReactorEvent, 0, 64 )
        return nil
    }
    return sp.next.ProcessEvent( ev )
}

func ( sp *structOrderProcessor ) ProcessEvent( ev ReactorEvent ) error {
    switch v := ev.( type ) {
    case StructStartEvent: return sp.processStructStart( v )
    case FieldStartEvent: return sp.processFieldStart( v )
    }
    panic( libErrorf( "unexpected event: %T", ev ) )
}

func ( sp *structOrderProcessor ) getFieldSender() ReactorEventProcessor {

    ps := NewPathSettingProcessor()
    if p := sp.startPath; p != nil { ps.SetStartPath( objpath.CopyOf( p ) ) }
    ps.skipStructureCheck = true
    return InitReactorPipeline( ps, sp.next )
}

func ( sp *structOrderProcessor ) sendReadyField( 
    state *structOrderFieldState ) error {

    rep := sp.getFieldSender()

    fsEv := NewFieldStartEvent( state.spec.Field )
    if err := rep.ProcessEvent( fsEv ); err != nil { return err }

    for _, ev := range state.acc {
        if err := rep.ProcessEvent( ev ); err != nil { return err }
    }

    return nil
}

func ( sp *structOrderProcessor ) sendReadyFields( isFinal bool ) error {
    for sp.fieldQueue.Len() > 0 {
        frnt := sp.fieldQueue.Front()
        state := frnt.Value.( *structOrderFieldState )
        if state.acc == nil {
            if isFinal && ( ! state.spec.Required ) {
                sp.fieldQueue.Remove( frnt )
                continue
            }
            return nil
        }
        sp.fieldQueue.Remove( frnt )
        if err := sp.sendReadyField( state ); err != nil { return err }
    }
    return nil
}

func ( sp *structOrderProcessor ) completeField() error {
    sp.cur = nil
    return sp.sendReadyFields( false )
}

func ( sp *structOrderProcessor ) processValue( v ValueEvent ) error {
    if sp.cur == nil || sp.cur.acc == nil {
        if err := sp.next.ProcessEvent( v ); err != nil { return err }
    } else {
        sp.cur.acc = append( sp.cur.acc, CopyEvent( v, false ) )
    }
    return sp.completeField()
}

func ( sp *structOrderProcessor ) valueEnded() error {
    return sp.completeField()
}

func ( sp *structOrderProcessor ) checkRequiredFields( ev ReactorEvent ) error {
    missing := make( []*Identifier, 0, 4 )
    sp.states.EachPair( func( k *Identifier, v interface{} ) {
        state := v.( *structOrderFieldState )
        if state.spec.Required && ( ! state.seen ) {
            missing = append( missing, state.spec.Field )
        }
    })
    if len( missing ) == 0 { return nil }
    return NewMissingFieldsError( ev.GetPath(), missing )
}

func ( sp *structOrderProcessor ) endStruct( ev ReactorEvent ) error {
    if err := sp.sendReadyFields( true ); err != nil { return err }
    if err := sp.checkRequiredFields( ev ); err != nil { return err }
    return sp.next.ProcessEvent( ev )
}

func ( fo *FieldOrderReactor ) peekProc() ReactorEventProcessor {
    if fo.stack.IsEmpty() { return nil }
    return fo.stack.Peek().( ReactorEventProcessor )
}

func ( fo *FieldOrderReactor ) peekStructProc() *structOrderProcessor {
    rep := fo.peekProc()
    if rep == nil { return nil }
    if res, ok := rep.( *structOrderProcessor ); ok { return res }
    return nil
}

func ( fo *FieldOrderReactor ) pushProc( 
    next ReactorEventProcessor, ev ReactorEvent ) error {

    fo.stack.Push( next )
    return next.ProcessEvent( ev )
}

func ( fo *FieldOrderReactor ) processContainerStart(
    ev ReactorEvent, next ReactorEventProcessor ) error {

    if sp := fo.peekStructProc(); sp != nil {
        return fo.pushProc( sp.fieldReactor(), ev )
    }
    if ! fo.stack.IsEmpty() { next = fo.peekProc() }
    return fo.pushProc( next, ev )
}

func ( fo *FieldOrderReactor ) structOrderGetNextProc( 
    next ReactorEventProcessor ) ReactorEventProcessor {
    if fo.stack.IsEmpty() { return next }
    rep := fo.peekProc()
    if sp, ok := rep.( *structOrderProcessor ); ok { return sp.fieldReactor() }
    return rep
}

func ( fo *FieldOrderReactor ) processStructStart(
    ev StructStartEvent, next ReactorEventProcessor,
) error {
    ord := fo.fog.FieldOrderFor( ev.Type )
    if ord == nil { return fo.processContainerStart( ev, next ) }
    sp := &structOrderProcessor{ ord: ord }
    sp.next = fo.structOrderGetNextProc( next )
    return fo.pushProc( sp, ev )
}

func ( fo *FieldOrderReactor ) processValue( 
    v ValueEvent, next ReactorEventProcessor ) error {

    if sp := fo.peekStructProc(); sp != nil { return sp.processValue( v ) }
    return fo.peekProc().ProcessEvent( v )
}

func ( fo *FieldOrderReactor ) processEnd(
    ev ReactorEvent, next ReactorEventProcessor ) error {
    rep := fo.stack.Pop().( ReactorEventProcessor )
    if sp, ok := rep.( *structOrderProcessor ); ok {
        if err := sp.endStruct( ev ); err != nil { return err }
    } else {
        if err := rep.ProcessEvent( ev ); err != nil { return err }
    }
    if sp := fo.peekStructProc(); sp != nil { return sp.valueEnded() }
    return nil
}

func ( fo *FieldOrderReactor ) processEventWithStack( 
    ev ReactorEvent, next ReactorEventProcessor ) error {

    switch v := ev.( type ) {
    case ListStartEvent, MapStartEvent: 
        return fo.processContainerStart( v, next )
    case StructStartEvent: return fo.processStructStart( v, next )
    case FieldStartEvent: return fo.peekProc().ProcessEvent( v )
    case ValueEvent: return fo.processValue( v, next )
    case EndEvent: return fo.processEnd( v, next )
    }
    panic( libErrorf( "unhandled event: %T", ev ) )
}

func ( fo *FieldOrderReactor ) ProcessEvent( 
    ev ReactorEvent, next ReactorEventProcessor ) error {

    if fo.stack.IsEmpty() && ( ! isStructStart( ev ) ) {
        return next.ProcessEvent( ev )
    }
    return fo.processEventWithStack( ev, next )
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

func ( t svcReqFieldTyper ) FieldTypeFor( 
    fld *Identifier, path objpath.PathNode ) ( TypeReference, error ) {

    if fld.Equals( IdParameters ) { return TypeSymbolMap, nil }
    return TypeValue, nil
}

func ( c svcReqCastIface ) FieldTyperFor( 
    qn *QualifiedTypeName, path objpath.PathNode ) ( FieldTyper, error ) {
    
    if qn.Equals( QnameServiceRequest ) { return svcReqFieldTyper( 1 ), nil }
    return nil, nil
}

func ( c svcReqCastIface ) CastAtomic( 
    in Value, 
    at *AtomicTypeReference, 
    path objpath.PathNode ) ( Value, error, bool ) {

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
    cr := NewCastReactor( TypeServiceRequest, svcReqCastIface( 1 ) )
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
        if err := ep.ProcessEvent( NewMapStartEvent() ); err != nil { return err }
        if err := ep.ProcessEvent( NewEndEvent() ); err != nil { return err }
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
    qn *QualifiedTypeName, path objpath.PathNode ) ( FieldTyper, error ) {

    return valueFieldTyper( 1 ), nil
}

func ( i svcRespCastIface ) CastAtomic(
    in Value, 
    at *AtomicTypeReference, 
    path objpath.PathNode ) ( Value, error, bool ) {

    return nil, nil, false
}

func ( sr *ServiceResponseReactor ) Init( rpi *ReactorPipelineInit ) {
    EnsureStructuralReactor( rpi )
    cr := NewCastReactor( TypeServiceResponse, svcRespCastIface( 1 ) )
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
    Log( msg string )
}

type DebugLoggerFunc func( string )

func ( f DebugLoggerFunc ) Log( msg string ) { f( msg ) }

type DebugReactor struct { 
    l DebugLogger 
    key ReactorKey
    Label string
}

func ( dr *DebugReactor ) ProcessEvent( ev ReactorEvent ) error {
    msg := EventToString( ev )
    if dr.Label != "" { msg = fmt.Sprintf( "[%s] %s", dr.Label, msg ) }
    dr.l.Log( msg )
    return nil
}

func NewDebugReactor( l DebugLogger ) *DebugReactor { 
    return &DebugReactor{ l: l }
}
