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
    Location objpath.PathNode
    Message string
}

func ( e *ReactorError ) Error() string { 
    return mg.FormatError( e.Location, e.Message )
}

func NewReactorError( path objpath.PathNode, msg string ) *ReactorError { 
    return &ReactorError{ Location: path, Message: msg }
}

func NewReactorErrorf( 
    path objpath.PathNode, tmpl string, args ...interface{} ) *ReactorError {

    return NewReactorError( path, fmt.Sprintf( tmpl, args... ) )
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
}

func NewMapStartEvent() *MapStartEvent {
    return &MapStartEvent{ reactorEventImpl: &reactorEventImpl{} }
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
}

func NewListStartEvent( typ *mg.ListTypeReference ) *ListStartEvent {
    return &ListStartEvent{ 
        reactorEventImpl: &reactorEventImpl{}, 
        Type: typ,
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
        pairs = append( pairs, []string{ "type", v.Type.ExternalForm() } )
    case *FieldStartEvent:
        pairs = append( pairs, []string{ "field", v.Field.ExternalForm() } )
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
    case *ListStartEvent: res = NewListStartEvent( v.Type )
    case *MapStartEvent: res = NewMapStartEvent()
    case *StructStartEvent: res = NewStructStartEvent( v.Type )
    case *FieldStartEvent: res = NewFieldStartEvent( v.Field )
    case *EndEvent: res = NewEndEvent()
    default: panic( libErrorf( "unhandled copy target: %T", ev ) )
    }
    if withPath { res.SetPath( ev.GetPath() ) }
    return res
}

func TypeOfEvent( ev ReactorEvent ) mg.TypeReference {
    switch v := ev.( type ) {
    case *ValueEvent: return mg.TypeOf( v.Val )
    case *ListStartEvent: return v.Type
    case *MapStartEvent: return mg.TypeSymbolMap
    case *StructStartEvent: return v.Type.AsAtomicType()
    }
    panic( libErrorf( "can't get type for: %T", ev ) )
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
    ev := NewListStartEvent( ml.Type )
    if err := vv.rep.ProcessEvent( ev ); err != nil { return err }
    for _, val := range ml.Values() {
        if err := vv.visitValue( val ); err != nil { return err }
    }
    return vv.rep.ProcessEvent( NewEndEvent() )
}

func ( vv valueVisit ) visitSymbolMap( sm *mg.SymbolMap ) error {
    ev := NewMapStartEvent()
    if err := vv.rep.ProcessEvent( ev ); err != nil { return err }
    return vv.visitSymbolMapFields( sm )
}

func ( vv valueVisit ) visitValue( mv mg.Value ) error {
    switch v := mv.( type ) {
    case *mg.Struct: return vv.visitStruct( v )
    case *mg.SymbolMap: return vv.visitSymbolMap( v )
    case *mg.List: return vv.visitList( v )
    }
    return vv.rep.ProcessEvent( NewValueEvent( mv ) )
}

type pathSetterCaller struct {
    ps *PathSettingProcessor
    rep ReactorEventProcessor
}

func ( c pathSetterCaller ) ProcessEvent( ev ReactorEvent ) error {
    return c.ps.ProcessEvent( ev, c.rep )
}

func VisitValuePath( 
    mv mg.Value, rep ReactorEventProcessor, path objpath.PathNode ) error {

    ps := NewPathSettingProcessor()
    if path != nil { ps.SetStartPath( path ) }
    vv := valueVisit{ rep: pathSetterCaller{ ps, rep } }
    return vv.visitValue( mv )
}

func VisitValue( mv mg.Value, rep ReactorEventProcessor ) error {
    return ( valueVisit{ rep: rep } ).visitValue( mv )
}

func isAssignableValueType( typ mg.TypeReference ) bool {
    return mg.CanAssignType( mg.TypeValue, typ )
}
