package mingle

import (
    "container/list"
    "fmt"
    "bitgirder/objpath"
//    "log"
)

type ReactorError struct { msg string }

func ( e *ReactorError ) Error() string { return e.msg }

func rctError( msg string ) *ReactorError { return &ReactorError{ msg } }

func rctErrorf( tmpl string, args ...interface{} ) *ReactorError {
    return rctError( fmt.Sprintf( tmpl, args... ) )
}

type ReactorEvent interface {}

type ValueEvent struct { Val Value }
type StructStartEvent struct { Type TypeReference }

type MapStartEvent int
const EvMapStart = MapStartEvent( 0 )

type FieldStartEvent struct { Field *Identifier }

type ListStartEvent int
const EvListStart = ListStartEvent( 0 )

type EndEvent int
const EvEnd = EndEvent( 0 )

type ReactorEventProcessor interface { ProcessEvent( ReactorEvent ) error }

type ReactorKey string

type ReactorFactory interface { CreateReactor() Reactor }

type reactorFuncFactory func() Reactor

func ( rff reactorFuncFactory ) CreateReactor() Reactor { return rff() }

func asReactorFactory( f func() Reactor ) ReactorFactory {
    return reactorFuncFactory( f )
}

type Reactor interface {
    Key() ReactorKey
    Init( rpi *ReactorPipelineInit )
    ProcessEvent( ev ReactorEvent, rep ReactorEventProcessor ) error
}

type ReactorPipeline struct {
    reactors []Reactor
}

type ReactorPipelineInit struct { 
    rp *ReactorPipeline 
    reactors []Reactor
}

func findReactor( reactors []Reactor, key ReactorKey ) ( Reactor, bool ) {
    for _, rct := range reactors { if rct.Key() == key { return rct, true } }
    return nil, false
}

func ( rpi *ReactorPipelineInit ) EnsurePredecessor( 
    key ReactorKey, rf ReactorFactory ) Reactor {
    if rct, ok := findReactor( rpi.reactors, key ); ok { return rct }
    rct := rf.CreateReactor()
    rpi.reactors = append( rpi.reactors, rct )
    return rct
}

// Might make this Init() if needed later
func ( rp *ReactorPipeline ) init() {
    rpInit := &ReactorPipelineInit{ 
        rp: rp, 
        reactors: make( []Reactor, 0, len( rp.reactors ) ),
    }
    for _, rct := range rp.reactors { 
        rct.Init( rpInit )
        rpInit.reactors = append( rpInit.reactors, rct )
    }
    rp.reactors = rpInit.reactors
}

// Could break this into separate methods later if needed: NewReactorPipeline()
// and ReactorPipeline.Init()
func InitReactorPipeline( reactors ...Reactor ) *ReactorPipeline {
    res := &ReactorPipeline{ reactors: reactors }
    res.init()
    return res
}

type pipelineCall struct {
    rp *ReactorPipeline
    idx int
}

func ( pc pipelineCall ) ProcessEvent( re ReactorEvent ) error {
    if pc.idx == len( pc.rp.reactors ) { return nil }
    rct := pc.rp.reactors[ pc.idx ]
    nextPc := pipelineCall{ pc.rp, pc.idx + 1 }
    return rct.ProcessEvent( re, nextPc )
}

func ( rp *ReactorPipeline ) ProcessEvent( re ReactorEvent ) error {
//    var err error
//    for _, rct := range rp.reactors { 
//        if re, err = rct.ProcessEvent( re ); err != nil { return err }
//    }
//    return nil
    return ( pipelineCall{ rp, 0 } ).ProcessEvent( re )
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

func ( m *structuralMap ) startField( fld *Identifier ) error {
    if m.pending == nil {
        if m.keys.HasKey( fld ) {
            return rctErrorf( "Multiple entries for field: %s", fld )
        }
        m.keys.Put( fld, true )
        m.pending = fld
        return nil
    }
    tmpl := "Saw start of field '%s' while expecting a value for '%s'"
    return rctErrorf( tmpl, fld, m.pending )
}

// Clears m.pending on nil return val when valReady
func ( m *structuralMap ) checkValue( valName string, valReady bool ) error {
    if m.pending == nil {
        tmpl := "Expected field name or end of fields but got %s"
        return rctErrorf( tmpl, valName )
    }
    if valReady { m.pending = nil }
    return nil 
}

func ( m *structuralMap ) end() error {
    if m.pending == nil { return nil }
    return rctErrorf( 
        "Saw end while expecting value for field '%s'", m.pending )
}

type structuralListIndex int

type StructuralReactor struct {
    stack *list.List
    topTyp ReactorTopType
    done bool
}

func NewStructuralReactor( topTyp ReactorTopType ) *StructuralReactor {
    return &StructuralReactor{ 
        stack: &list.List{}, 
        topTyp: topTyp,
    }
}

func ( sr *StructuralReactor ) Init( rpi *ReactorPipelineInit ) {}

const ReactorKeyStructuralReactor = ReactorKey( "mingle.StructuralReactor" )

var StructuralReactorFactory = asReactorFactory(
    func() Reactor { return NewStructuralReactor( ReactorTopTypeValue ) },
)

func ( sr *StructuralReactor ) Key() ReactorKey { 
    return ReactorKeyStructuralReactor
}

func ( sr *StructuralReactor ) buildPath( e *list.Element, p idPath ) idPath {
    if e == nil { return p }
    switch v := e.Value.( type ) {
    case *structuralMap: 
        if fld := v.pending; fld != nil { p = objpath.Descend( p, fld ) }
    case structuralListIndex: p = objpath.StartList( p ).SetIndex( int( v ) )
    default: panic( libErrorf( "Unhandled stack element: %T", e.Value ) )
    }
    return sr.buildPath( e.Prev(), p )
}

func ( sr *StructuralReactor ) appendPath( p idPath ) idPath {
    return sr.buildPath( sr.stack.Back(), p )
}

func ( sr *StructuralReactor ) GetPath() objpath.PathNode {
    return sr.appendPath( nil )
}

func ( sr *StructuralReactor ) AppendPath( 
    p objpath.PathNode ) objpath.PathNode {
    return sr.appendPath( p )
}

func ( sr *StructuralReactor ) Error( msg string ) error {
    if p := sr.GetPath(); p != nil {
        msg = fmt.Sprintf( "%s: %s", FormatIdPath( p ), msg )
    }
    return rctError( msg )
}

func ( sr *StructuralReactor ) Errorf( 
    tmpl string, args ...interface{} ) error {
    return sr.Error( fmt.Sprintf( tmpl, args... ) )
}

func ( sr *StructuralReactor ) getReactorTopTypeError( valName string ) error {
    return getReactorTopTypeError( valName, sr.topTyp )
}

func ( sr *StructuralReactor ) checkActive( call string ) error {
    if sr.done { return rctErrorf( "%s() called, but struct is built", call ) }
    return nil
}

// sr.stack known to be non-empty when this returns without error, unless top
// type is value.
func ( sr *StructuralReactor ) checkValue( 
    valName string, valReady bool ) error {
    if sr.stack.Len() == 0 {
        if sr.topTyp == ReactorTopTypeValue { return nil }
        return sr.getReactorTopTypeError( valName )
    }
    elt := sr.stack.Front().Value
    if m, ok := elt.( *structuralMap ); ok { 
        return m.checkValue( valName, valReady ) 
    }
    return nil
}

func ( sr *StructuralReactor ) push( elt interface{} ) { 
    sr.stack.PushFront( elt ) 
}

func ( sr *StructuralReactor ) startStruct() error {
    if err := sr.checkActive( "StartStruct" ); err != nil { return err }
    // skip check if we're pushing the top level struct
    if sr.stack.Len() > 0 {
        if err := sr.checkValue( "struct start", false ); err != nil { 
            return err 
        }
    }
    sr.push( newStructuralMap() )
    return nil
}

func ( sr *StructuralReactor ) startMap() error {
    if err := sr.checkActive( "StartMap" ); err != nil { return err }
    if err := sr.checkValue( "map start", false ); err != nil { return err }
    sr.push( newStructuralMap() )
    return nil
}

func ( sr *StructuralReactor ) startList() error {
    if err := sr.checkActive( "StartList" ); err != nil { return err }
    if err := sr.checkValue( "list start", false ); err != nil { return err }
    sr.push( structuralListIndex( 0 ) )
    return nil
}

func ( sr *StructuralReactor ) startField( fld *Identifier ) error {
    if err := sr.checkActive( "StartField" ); err != nil { return err }
    if ok := sr.stack.Len() > 0; ok {
        elt := sr.stack.Front().Value
        switch v := elt.( type ) {
        case structuralListIndex: 
            tmpl := "Expected list value but got start of field '%s'"
            return rctErrorf( tmpl, fld )
        case *structuralMap: return v.startField( fld )
        default: panic( libErrorf( "Invalid stack element: %T", elt ) )
        }
    }
    errLoc := fmt.Sprintf( "field '%s'", fld )
    return getReactorTopTypeError( errLoc, ReactorTopTypeStruct )
}

func ( sr *StructuralReactor ) value() error {
    if err := sr.checkActive( "Value" ); err != nil { return err }
    if err := sr.checkValue( "value", true ); err != nil { return err }
//    if sr.stack.Len() > 0 {
//        front := sr.stack.Front()
//        if idx, ok := front.Value.( structuralListIndex ); ok {
//            front.Value = structuralListIndex( idx + 1 )
//        }
//    }
    return nil
}

func ( sr *StructuralReactor ) end() error {
    if err := sr.checkActive( "End" ); err != nil { return err }
    if sr.stack.Len() == 0 { return sr.getReactorTopTypeError( "end" ) }
    elt := sr.stack.Remove( sr.stack.Front() )
    switch v := elt.( type ) {
    case *structuralMap: if err := v.end(); err != nil { return err }
    case structuralListIndex: {} // list -- end() is always okay
    default: panic( libErrorf( "Unexpected stack element: %T", elt ) )
    }
    // if we're not done then we just completed an intermediate value which
    // needs to be processed
    if sr.done = sr.stack.Len() == 0; ! sr.done { return sr.value() }
    return nil
}

func ( sr *StructuralReactor ) incrementIfList() {
    if sr.stack.Len() == 0 { return }
    front := sr.stack.Front()
    if idx, ok := front.Value.( structuralListIndex ); ok {
        front.Value = structuralListIndex( idx + 1 )
    }
}

func ( sr *StructuralReactor ) ProcessEvent( 
    ev ReactorEvent, rep ReactorEventProcessor ) error {
    var err error
    isValue := false
    switch v := ev.( type ) {
    case StructStartEvent: err = sr.startStruct()
    case MapStartEvent: err = sr.startMap()
    case ListStartEvent: err = sr.startList()
    case FieldStartEvent: err = sr.startField( v.Field )
    case ValueEvent: err, isValue = sr.value(), true
    case EndEvent: err, isValue = sr.end(), true
    default: panic( libErrorf( "Unhandled event: %T", ev ) )
    }
    if err == nil { err = rep.ProcessEvent( ev ) }
    if err == nil && isValue && ( ! sr.done ) { sr.incrementIfList() }
    return err
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
}

const ReactorKeyValueBuilder = ReactorKey( "mingle.ValueBuilder" )

func NewValueBuilder() *ValueBuilder {
    return &ValueBuilder{ accs: &list.List{} }
}

func ( vb *ValueBuilder ) Init( rpi *ReactorPipelineInit ) {}

func ( vb *ValueBuilder ) Key() ReactorKey { return ReactorKeyValueBuilder }

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

// Panics if result of val is not ready
func ( vb *ValueBuilder ) GetValue() Value {
    if vb.val == nil { panic( rctErrorf( "Value is not yet built" ) ) }
    return vb.val
}

func ( vb *ValueBuilder ) startField( fld *Identifier ) {
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
}

func ( vb *ValueBuilder ) end() error {
    acc := vb.popAcc()
    if val, err := acc.end(); err == nil {
        vb.valueReady( val )
    } else { return err }
    return nil
}

func ( vb *ValueBuilder ) ProcessEvent( 
    ev ReactorEvent, rep ReactorEventProcessor ) error {
    switch v := ev.( type ) {
    case ValueEvent: vb.valueReady( v.Val )
    case ListStartEvent: vb.pushAcc( newListAcc() )
    case MapStartEvent: vb.pushAcc( newMapAcc() )
    case StructStartEvent: vb.pushAcc( newStructAcc( v.Type ) )
    case FieldStartEvent: vb.startField( v.Field )
    case EndEvent: if err := vb.end(); err != nil { return err }
    default: panic( libErrorf( "Unhandled event: %T", ev ) )
    }
    return rep.ProcessEvent( ev )
}

type castContext struct {
    elt interface{}
    expct TypeReference
}

type mapCast struct {}

func ( mc mapCast ) startField( 
    fse FieldStartEvent, rep ReactorEventProcessor ) error {
    return rep.ProcessEvent( fse )
}

func ( mc mapCast ) value( ve ValueEvent, rep ReactorEventProcessor ) error {
    return rep.ProcessEvent( ve )
}

func ( mc mapCast ) end( ee EndEvent, rep ReactorEventProcessor ) error {
    return rep.ProcessEvent( ee )
}

type listCast struct {
    lt *ListTypeReference
    sawVals bool
}

type CastReactor struct {
    path idPath
    sr *StructuralReactor
    stack *list.List // stack of castContext
}

func NewCastReactor( expct TypeReference, path objpath.PathNode ) *CastReactor {
    res := &CastReactor{
        path: path,
        stack: &list.List{},
    }
    res.stack.PushFront( castContext{ elt: expct, expct: expct } )
    return res
}

const ReactorKeyCastReactor = ReactorKey( "mingle.CastReactor" )

func ( cr *CastReactor ) Init( rpi *ReactorPipelineInit ) {
    cr.sr = rpi.EnsurePredecessor( 
        ReactorKeyStructuralReactor, 
        StructuralReactorFactory,
    ).( *StructuralReactor )
}

func ( cr *CastReactor ) Key() ReactorKey { return ReactorKeyCastReactor }

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

func ( cr *CastReactor ) getPath() idPath { return cr.sr.AppendPath( cr.path ) }

func ( cr *CastReactor ) newTypeCastErrorPath( 
    act TypeReference, p idPath ) *TypeCastError {
    return newTypeCastError( cr.peek().expct, act, p )
}

func ( cr *CastReactor ) newTypeCastError( act TypeReference ) *TypeCastError {
    return cr.newTypeCastErrorPath( act, cr.getPath() )
}

func ( cr *CastReactor ) castContextPanic( 
    cc castContext, desc string ) error {
    return libErrorf( "Unhandled stack element for %s: %T", desc, cc.elt )
}

func ( cr *CastReactor ) stackTypePanic( desc string ) error {
    return cr.castContextPanic( cr.peek(), desc )
}

func ( cr *CastReactor ) completeValue( 
    ve ValueEvent, t TypeReference, rep ReactorEventProcessor ) error {
    switch typVal := t.( type ) {
    case *AtomicTypeReference: 
        v2, err := castAtomic( ve.Val, typVal, cr.getPath() )
        if err == nil { return rep.ProcessEvent( ValueEvent{ v2 } ) }
        return err
    case *NullableTypeReference: 
        if _, ok := ve.Val.( *Null ); ok { return rep.ProcessEvent( ve ) }
        return cr.completeValue( ve, typVal.Type, rep )
    case *ListTypeReference: return cr.newTypeCastError( TypeOf( ve.Val ) )
    }
    panic( libErrorf( "Uhandled type: %T", t ) )
}

func ( cr *CastReactor ) value( 
    ve ValueEvent, rep ReactorEventProcessor ) error {
    switch elt := cr.peek().elt.( type ) {
    case *AtomicTypeReference, *NullableTypeReference, *ListTypeReference:
        return cr.completeValue( ve, elt.( TypeReference ), rep )
    case *listCast: 
        elt.sawVals = true
        return cr.completeValue( ve, elt.lt.ElementType, rep )
    case mapCast: return elt.value( ve, rep )
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
        p := cr.getPath().Parent()
        return cr.newTypeCastErrorPath( typeOpaqueList, p )
    }
    panic( libErrorf( "Uhandled type: %T", typ ) )
}

func ( cr *CastReactor ) startList( 
    le ListStartEvent, rep ReactorEventProcessor ) error {
    switch elt := cr.peek().elt.( type ) {
    case *ListTypeReference, *AtomicTypeReference, *NullableTypeReference: 
        return cr.completeStartList( elt.( TypeReference ), le, rep )
    case *listCast: return cr.completeStartList( elt.lt.ElementType, le, rep )
    }
    panic( cr.stackTypePanic( "list start" ) )
}

func ( cr *CastReactor ) startMap( 
    sm MapStartEvent, rep ReactorEventProcessor ) error {
    switch elt := cr.peek().elt.( type ) {
    case *AtomicTypeReference:
        if elt.Equals( TypeSymbolMap ) {
            cr.push( castContext{ elt: mapCast{}, expct: TypeSymbolMap } )
            return rep.ProcessEvent( sm )
        }
        return cr.newTypeCastError( TypeSymbolMap )
    }
    panic( cr.stackTypePanic( "start map" ) )
}

func ( cr *CastReactor ) completeStartStruct( 
    ss StructStartEvent, t TypeReference, rep ReactorEventProcessor ) error {
    if t.Equals( ss.Type ) {
        cr.push( castContext{ elt: mapCast{}, expct: ss.Type } )
        return rep.ProcessEvent( ss )
    }
    return cr.newTypeCastError( ss.Type )
}

func ( cr *CastReactor ) startStruct( 
    ss StructStartEvent, rep ReactorEventProcessor ) error {
    switch elt := cr.peek().elt.( type ) {
    case *AtomicTypeReference: return cr.completeStartStruct( ss, elt, rep )
    case *ListTypeReference, *NullableTypeReference:
        return cr.newTypeCastError( ss.Type )
    case *listCast: return cr.completeStartStruct( ss, elt.lt.ElementType, rep )
    }
    panic( cr.stackTypePanic( "start struct" ) )
}

func ( cr *CastReactor ) startField( 
    fse FieldStartEvent, rep ReactorEventProcessor ) error {
    mc := cr.peek().elt.( mapCast ) // okay since structure check precedes
    return mc.startField( fse, rep )
}

func ( cr *CastReactor ) noteEndAsValIfList() {
    if cr.stack.Len() == 0 { return }
    if lc, ok := cr.peek().elt.( *listCast ); ok { lc.sawVals = true }
}

func ( cr *CastReactor ) end( ee EndEvent, rep ReactorEventProcessor ) error {
    cc := cr.pop()
    cr.noteEndAsValIfList()
    switch v := cc.elt.( type ) {
    case mapCast: return v.end( ee, rep )
    case *listCast:
        if ! ( v.sawVals || v.lt.AllowsEmpty ) {
            return newValueCastErrorf( cr.getPath(), "List is empty" )
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

func CastValue( 
    mgVal Value, typ TypeReference, path objpath.PathNode ) ( Value, error ) {
    vb := NewValueBuilder()
    pip := InitReactorPipeline( NewCastReactor( typ, path ), vb )
    if err := VisitValue( mgVal, pip ); err != nil { return nil, err }
    return vb.GetValue(), nil
//    if path == nil { return nil, errors.New( "path arg is nil" ) }
//    if mgVal == nil { return nil, errors.New( "mgVal is nil" ) }
//    return castValue( mgVal, typ, path )
}
