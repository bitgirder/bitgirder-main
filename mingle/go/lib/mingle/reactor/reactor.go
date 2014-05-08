package reactor

import (
    "fmt"
    "bitgirder/objpath"
    "bitgirder/pipeline"
    "strings"
    mg "mingle"
//    "log"
)

type ReactorError struct { 
    ve mg.ValueErrorImpl
    msg string 
}

func ( e *ReactorError ) Error() string { return e.ve.MakeError( e.msg ) }

func ( e *ReactorError ) Message() string { return e.msg }

func ( e *ReactorError) Location() objpath.PathNode { return e.ve.Location() }

func rctError( path objpath.PathNode, msg string ) *ReactorError { 
    res := &ReactorError{ msg: msg, ve: mg.ValueErrorImpl{} } 
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
    Val mg.Value 
}

func NewValueEvent( val mg.Value ) *ValueEvent { 
    return &ValueEvent{ Val: val, reactorEventImpl: &reactorEventImpl{} } 
}

type ValueAllocationEvent struct {
    *reactorEventImpl
    Type mg.TypeReference
    Id mg.PointerId
}

func NewValueAllocationEvent( 
    typ mg.TypeReference, id mg.PointerId ) *ValueAllocationEvent {

    return &ValueAllocationEvent{ 
        Id: id, 
        Type: typ,
        reactorEventImpl: &reactorEventImpl{},
    }
}

type ValueReferenceEvent struct {
    *reactorEventImpl
    Id mg.PointerId
}

func NewValueReferenceEvent( id mg.PointerId ) *ValueReferenceEvent {
    return &ValueReferenceEvent{
        Id: id,
        reactorEventImpl: &reactorEventImpl{},
    }
}

type StructStartEvent struct { 
    *reactorEventImpl
    Type *mg.QualifiedTypeName 
}

func NewStructStartEvent( typ *mg.QualifiedTypeName ) *StructStartEvent {
    return &StructStartEvent{ Type: typ, reactorEventImpl: &reactorEventImpl{} }
}

func isStructStart( ev ReactorEvent ) bool {
    _, ok := ev.( *StructStartEvent )
    return ok
}

type MapStartEvent struct {
    *reactorEventImpl
    Id mg.PointerId
}

func NewMapStartEvent( id mg.PointerId ) *MapStartEvent {
    return &MapStartEvent{ Id: id, reactorEventImpl: &reactorEventImpl{} }
}

type FieldStartEvent struct { 
    *reactorEventImpl
    Field *mg.Identifier 
}

func NewFieldStartEvent( fld *mg.Identifier ) *FieldStartEvent {
    return &FieldStartEvent{ Field: fld, reactorEventImpl: &reactorEventImpl{} }
}

type ListStartEvent struct {
    *reactorEventImpl
    Type *mg.ListTypeReference // the element type
    Id mg.PointerId
}

func NewListStartEvent( 
    typ *mg.ListTypeReference, id mg.PointerId ) *ListStartEvent {

    return &ListStartEvent{ 
        reactorEventImpl: &reactorEventImpl{}, 
        Type: typ,
        Id: id,
    }
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
        pairs = append( pairs, []string{ "value", mg.QuoteValue( v.Val ) } )
    case *StructStartEvent:
        pairs = append( pairs, []string{ "type", v.Type.ExternalForm() } )
    case *ListStartEvent:
        pairs = append( pairs, 
            []string{ "id", v.Id.String() },
            []string{ "type", v.Type.ExternalForm() },
        )
    case *MapStartEvent:
        pairs = append( pairs, []string{ "id", v.Id.String() } )
    case *FieldStartEvent:
        pairs = append( pairs, []string{ "field", v.Field.ExternalForm() } )
    case *ValueAllocationEvent:
        pairs = append( pairs, 
            []string{ "id", v.Id.String() }, 
            []string{ "type", v.Type.ExternalForm() },
        )
    case *ValueReferenceEvent:
        pairs = append( pairs, []string{ "id", v.Id.String() } )
    }
    if p := ev.GetPath(); p != nil {
        pairs = append( pairs, []string{ "path", mg.FormatIdPath( p ) } )
    }
    elts := make( []string, len( pairs ) )
    for i, pair := range pairs { elts[ i ] = strings.Join( pair, " = " ) }
    return fmt.Sprintf( "[ %s ]", strings.Join( elts, ", " ) )
}

func CopyEvent( ev ReactorEvent, withPath bool ) ReactorEvent {
    var res ReactorEvent
    switch v := ev.( type ) {
    case *ValueEvent: res = NewValueEvent( v.Val )
    case *ListStartEvent: res = NewListStartEvent( v.Type, v.Id )
    case *MapStartEvent: res = NewMapStartEvent( v.Id )
    case *StructStartEvent: res = NewStructStartEvent( v.Type )
    case *FieldStartEvent: res = NewFieldStartEvent( v.Field )
    case *EndEvent: res = NewEndEvent()
    case *ValueAllocationEvent: res = NewValueAllocationEvent( v.Type, v.Id )
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
    visitMap map[ mg.PointerId ] bool
}

// If a has been visited this method sends a reference event to vv.rep and
// returns ( err, false ) where err is the value returned by vv.rep. If a has
// not been visited before returns ( nil, true ) and updates visitMap for a.
func ( vv valueVisit ) visitReference( a mg.Addressed ) ( error, bool ) {
    addr := a.Address()
    if _, ok := vv.visitMap[ addr ]; ok {
        ev := NewValueReferenceEvent( addr )
        return vv.rep.ProcessEvent( ev ), true
    }
    vv.visitMap[ addr ] = true
    return nil, false
}

func ( vv valueVisit ) visitSymbolMapFields( m *mg.SymbolMap ) error {
    err := m.EachPairError( func( fld *mg.Identifier, val mg.Value ) error {
        ev := NewFieldStartEvent( fld )
        if err := vv.rep.ProcessEvent( ev ); err != nil { return err }
        return vv.visitValue( val )
    })
    if err != nil { return err }
    return vv.rep.ProcessEvent( NewEndEvent() )
}

func ( vv valueVisit ) visitStruct( ms *mg.Struct ) error {
    ev := NewStructStartEvent( ms.Type )
    if err := vv.rep.ProcessEvent( ev ); err != nil { return err }
    return vv.visitSymbolMapFields( ms.Fields )
}

func ( vv valueVisit ) visitList( ml *mg.List ) error {
    if err, ok := vv.visitReference( ml ); ok { return err }
    ev := NewListStartEvent( ml.Type, ml.Address() )
    if err := vv.rep.ProcessEvent( ev ); err != nil { return err }
    for _, val := range ml.Values() {
        if err := vv.visitValue( val ); err != nil { return err }
    }
    return vv.rep.ProcessEvent( NewEndEvent() )
}

func ( vv valueVisit ) visitSymbolMap( sm *mg.SymbolMap ) error {
    if err, ok := vv.visitReference( sm ); ok { return err }
    ev := NewMapStartEvent( sm.Address() )
    if err := vv.rep.ProcessEvent( ev ); err != nil { return err }
    return vv.visitSymbolMapFields( sm )
}

func ( vv valueVisit ) visitValuePointer( vp mg.ValuePointer ) error {
    if err, ok := vv.visitReference( vp ); ok { return err }
    typ := mg.TypeOf( vp.Dereference() )
    ev := NewValueAllocationEvent( typ, vp.Address() ) 
    if err := vv.rep.ProcessEvent( ev ); err != nil { return err }
    return vv.visitValue( vp.Dereference() )
}

func ( vv valueVisit ) visitValue( mv mg.Value ) error {
    switch v := mv.( type ) {
    case *mg.Struct: return vv.visitStruct( v )
    case *mg.SymbolMap: return vv.visitSymbolMap( v )
    case *mg.List: return vv.visitList( v )
    case mg.ValuePointer: return vv.visitValuePointer( v )
    }
    return vv.rep.ProcessEvent( NewValueEvent( mv ) )
}

func VisitValue( mv mg.Value, rep ReactorEventProcessor ) error {
    vv := valueVisit{ rep: rep, visitMap: make( map[ mg.PointerId ] bool ) }
    return vv.visitValue( mv )
}
