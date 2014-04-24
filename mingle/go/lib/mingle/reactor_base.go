package mingle

import (
    "fmt"
    "bitgirder/objpath"
    "bitgirder/pipeline"
    "strings"
//    "log"
)

type ReactorError struct { 
    ve ValueErrorImpl
    msg string 
}

func ( e *ReactorError ) Error() string { return e.ve.MakeError( e.msg ) }

func ( e *ReactorError ) Message() string { return e.msg }

func ( e *ReactorError) Location() objpath.PathNode { return e.ve.Location() }

func rctError( path objpath.PathNode, msg string ) *ReactorError { 
    res := &ReactorError{ msg: msg, ve: ValueErrorImpl{} } 
    if path != nil { res.ve.Path = path }
    return res
}

func rctErrorf( 
    path objpath.PathNode, tmpl string, args ...interface{} ) *ReactorError {

    return rctError( path, fmt.Sprintf( tmpl, args... ) )
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

func NewValueEvent( val Value ) *ValueEvent { 
    return &ValueEvent{ Val: val, reactorEventImpl: &reactorEventImpl{} } 
}

type ValueAllocationEvent struct {
    *reactorEventImpl
    Id PointerId
}

func NewValueAllocationEvent( id PointerId ) *ValueAllocationEvent {
    return &ValueAllocationEvent{ 
        Id: id, 
        reactorEventImpl: &reactorEventImpl{},
    }
}

type ValueReferenceEvent struct {
    *reactorEventImpl
    Id PointerId
}

func NewValueReferenceEvent( id PointerId ) *ValueReferenceEvent {
    return &ValueReferenceEvent{
        Id: id,
        reactorEventImpl: &reactorEventImpl{},
    }
}

type StructStartEvent struct { 
    *reactorEventImpl
    Type *QualifiedTypeName 
}

func NewStructStartEvent( typ *QualifiedTypeName ) *StructStartEvent {
    return &StructStartEvent{ Type: typ, reactorEventImpl: &reactorEventImpl{} }
}

func isStructStart( ev ReactorEvent ) bool {
    _, ok := ev.( *StructStartEvent )
    return ok
}

type MapStartEvent struct {
    *reactorEventImpl
}

func NewMapStartEvent() *MapStartEvent {
    return &MapStartEvent{ reactorEventImpl: &reactorEventImpl{} }
}

type FieldStartEvent struct { 
    *reactorEventImpl
    Field *Identifier 
}

func NewFieldStartEvent( fld *Identifier ) *FieldStartEvent {
    return &FieldStartEvent{ Field: fld, reactorEventImpl: &reactorEventImpl{} }
}

type ListStartEvent struct {
    *reactorEventImpl
}

func NewListStartEvent() *ListStartEvent {
    return &ListStartEvent{ reactorEventImpl: &reactorEventImpl{} }
}

type EndEvent struct {
    *reactorEventImpl
}

func NewEndEvent() *EndEvent {
    return &EndEvent{ reactorEventImpl: &reactorEventImpl{} }
}

func EventToString( ev ReactorEvent ) string {
    pairs := [][]string{ { "type", fmt.Sprintf( "%T", ev ) } }
    switch v := ev.( type ) {
    case *ValueEvent: 
        pairs = append( pairs, []string{ "value", QuoteValue( v.Val ) } )
    case *StructStartEvent:
        pairs = append( pairs, []string{ "type", v.Type.ExternalForm() } )
    case *FieldStartEvent:
        pairs = append( pairs, []string{ "field", v.Field.ExternalForm() } )
    case *ValueAllocationEvent:
        pairs = append( pairs, []string{ "id", v.Id.String() } )
    case *ValueReferenceEvent:
        pairs = append( pairs, []string{ "id", v.Id.String() } )
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
    case *ValueEvent: res = NewValueEvent( v.Val )
    case *ListStartEvent: res = NewListStartEvent()
    case *MapStartEvent: res = NewMapStartEvent()
    case *StructStartEvent: res = NewStructStartEvent( v.Type )
    case *FieldStartEvent: res = NewFieldStartEvent( v.Field )
    case *EndEvent: res = NewEndEvent()
    case *ValueAllocationEvent: res = NewValueAllocationEvent( v.Id )
    case *ValueReferenceEvent: res = NewValueReferenceEvent( v.Id )
    default: panic( libErrorf( "unhandled copy target: %T", ev ) )
    }
    if withPath { res.SetPath( ev.GetPath() ) }
    return res
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

type ReactorEventProcessor interface { ProcessEvent( ReactorEvent ) error }

type ReactorEventProcessorFunc func( ev ReactorEvent ) error

func ( f ReactorEventProcessorFunc ) ProcessEvent( ev ReactorEvent ) error {
    return f( ev )
}

var DiscardProcessor = ReactorEventProcessorFunc( 
    func( ev ReactorEvent ) error { return nil } )

type PipelineProcessor interface {
    ProcessEvent( ev ReactorEvent, rep ReactorEventProcessor ) error
}

func makePipelineReactor( 
    elt interface{}, next ReactorEventProcessor ) ReactorEventProcessor {

    var f ReactorEventProcessorFunc

    switch v := elt.( type ) {
    case PipelineProcessor:
        f = func( ev ReactorEvent ) error { return v.ProcessEvent( ev, next ) }
    case ReactorEventProcessor:
        f = func( ev ReactorEvent ) error { 
            if err := v.ProcessEvent( ev ); err != nil { return err }
            return next.ProcessEvent( ev )
        }
    default: panic( libErrorf( "unhandled pipeline element: %T", elt ) )
    }

    return f
}

func InitReactorPipeline( elts ...interface{} ) ReactorEventProcessor {
    pip := pipeline.NewPipeline()
    for _, elt := range elts { pip.Add( elt ) }
    var res ReactorEventProcessor = DiscardProcessor
    pip.VisitReverse( func( elt interface{} ) {
        res = makePipelineReactor( elt, res ) 
    })
    return res
}

type valueVisit struct {
    rep ReactorEventProcessor
    visitMap map[ PointerId ] bool
}

func ( vv valueVisit ) visitAddressed( a Addressed ) ( error, bool ) {
    addr := a.Address()
    if _, ok := vv.visitMap[ addr ]; ok {
        ev := NewValueReferenceEvent( addr )
        return vv.rep.ProcessEvent( ev ), true
    }
    return nil, false
}

func ( vv valueVisit ) visitSymbolMap( m *SymbolMap, callStart bool ) error {
    if callStart {
        if err := vv.rep.ProcessEvent( NewMapStartEvent() ); err != nil { 
            return err 
        }
    }
    err := m.EachPairError( func( fld *Identifier, val Value ) error {
        ev := NewFieldStartEvent( fld )
        if err := vv.rep.ProcessEvent( ev ); err != nil { return err }
        return vv.visitValue( val )
    })
    if err != nil { return err }
    return vv.rep.ProcessEvent( NewEndEvent() )
}

func ( vv valueVisit ) visitStruct( ms *Struct ) error {
    ev := NewStructStartEvent( ms.Type )
    if err := vv.rep.ProcessEvent( ev ); err != nil { return err }
    return vv.visitSymbolMap( ms.Fields, false )
}

func ( vv valueVisit ) visitList( ml *List ) error {
    if err, ok := vv.visitAddressed( ml ); ok { return err }
    vv.visitMap[ ml.Address() ] = true
    if err := vv.rep.ProcessEvent( NewListStartEvent() ); err != nil { 
        return err 
    }
    for _, val := range ml.Values() {
        if err := vv.visitValue( val ); err != nil { return err }
    }
    return vv.rep.ProcessEvent( NewEndEvent() )
}

func ( vv valueVisit ) visitValuePointer( vp ValuePointer ) error {
    addr := vp.ValueAddress()
    if _, ok := vv.visitMap[ addr ]; ok {
        ev := NewValueReferenceEvent( addr )
        return vv.rep.ProcessEvent( ev )
    }
    ev := NewValueAllocationEvent( addr ) 
    if err := vv.rep.ProcessEvent( ev ); err != nil { return err }
    vv.visitMap[ addr ] = true
    return vv.visitValue( vp.Dereference() )
}

func ( vv valueVisit ) visitValue( mv Value ) error {
    switch v := mv.( type ) {
    case *Struct: return vv.visitStruct( v )
    case *SymbolMap: return vv.visitSymbolMap( v, true )
    case *List: return vv.visitList( v )
    case ValuePointer: return vv.visitValuePointer( v )
    }
    return vv.rep.ProcessEvent( NewValueEvent( mv ) )
}

func VisitValue( mv Value, rep ReactorEventProcessor ) error {
    vv := valueVisit{ rep: rep, visitMap: make( map[ PointerId ] bool ) }
    return vv.visitValue( mv )
}
