package mingle

import (
    "container/list"
    "fmt"
    "bitgirder/objpath"
    "bitgirder/pipeline"
    "bytes"
    "strings"
    "bitgirder/stack"
//    "log"
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

func NewValueEvent( val Value ) *ValueEvent { 
    return &ValueEvent{ Val: val, reactorEventImpl: &reactorEventImpl{} } 
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
    default: panic( libErrorf( "unhandled copy target: %T", ev ) )
    }
    if withPath { res.SetPath( ev.GetPath() ) }
    return res
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

// Could break this into separate methods later if needed: NewReactorPipeline()
// and ReactorPipeline.Init()
func InitReactorPipeline( elts ...interface{} ) ReactorEventProcessor {
    pip := pipeline.NewPipeline()
    for _, elt := range elts { pip.Add( elt ) }
    var res ReactorEventProcessor = DiscardProcessor
    pip.VisitReverse( func( elt interface{} ) {
        res = makePipelineReactor( elt, res ) 
    })
    return res
}

func visitSymbolMap( 
    m *SymbolMap, callStart bool, rep ReactorEventProcessor ) error {

    if callStart {
        if err := rep.ProcessEvent( NewMapStartEvent() ); err != nil { 
            return err 
        }
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

type StructuralReactor struct {
    stack *stack.Stack
    topTyp ReactorTopType
    done bool
}

func NewStructuralReactor( topTyp ReactorTopType ) *StructuralReactor {
    return &StructuralReactor{ stack: stack.NewStack(), topTyp: topTyp }
}

type listAccType int

type mapStructureCheck struct {
    seen *IdentifierMap
}

func newMapStructureCheck() *mapStructureCheck {
    return &mapStructureCheck{ seen: NewIdentifierMap() }
}

func ( mc *mapStructureCheck ) startField( fld *Identifier ) error {
    if mc.seen.HasKey( fld ) {
        return rctErrorf( "Multiple entries for field: %s", fld.ExternalForm() )
    }
    mc.seen.Put( fld, true )
    return nil
}

func ( sr *StructuralReactor ) descForEvent( ev ReactorEvent ) string {
    switch v := ev.( type ) {
    case *ListStartEvent: return "list start"
    case *MapStartEvent: return "map start"
    case *EndEvent: return "end"
    case *ValueEvent: return "value"
    case *FieldStartEvent: return sr.sawDescFor( v.Field )
    case *StructStartEvent: return sr.sawDescFor( v.Type )
    }
    panic( libErrorf( "unhandled event: %T", ev ) )
}

func ( sr *StructuralReactor ) expectDescFor( val interface{} ) string {
    if val == nil { return "BEGIN" }
    switch v := val.( type ) {
    case *Identifier: 
        return fmt.Sprintf( "a value for field '%s'", v.ExternalForm() )
    case listAccType: return "a list value"
    }
    panic( libErrorf( "unhandled desc value: %T", val ) )
}

func ( sr *StructuralReactor ) sawDescFor( val interface{} ) string {
    if val == nil { return "BEGIN" }
    switch v := val.( type ) {
    case *Identifier: 
        return fmt.Sprintf( "start of field '%s'", v.ExternalForm() )
    case *QualifiedTypeName:
        return fmt.Sprintf( "start of struct %s", v.ExternalForm() )
    case ReactorEvent: return sr.descForEvent( v )
    }
    panic( libErrorf( "unhandled val: %T", val ) )
}

func ( sr *StructuralReactor ) checkNotDone( ev ReactorEvent ) error {
    if ! sr.done { return nil }
    return rctErrorf( "Saw %s after value was built", sr.sawDescFor( ev ) );
}

func ( sr *StructuralReactor ) failTopType( ev ReactorEvent ) error {
    desc := sr.descForEvent( ev )
    return rctErrorf( "Expected %s but got %s", sr.topTyp, desc )
}

func ( sr *StructuralReactor ) couldStartWithEvent( ev ReactorEvent ) bool {
    topIsVal := sr.topTyp == ReactorTopTypeValue
    switch ev.( type ) {
    case *ValueEvent: return topIsVal
    case *ListStartEvent: return sr.topTyp == ReactorTopTypeList || topIsVal
    case *MapStartEvent: return sr.topTyp == ReactorTopTypeMap || topIsVal
    case *StructStartEvent: return sr.topTyp == ReactorTopTypeStruct || topIsVal
    }
    return false
}

func ( sr *StructuralReactor ) checkTopType( ev ReactorEvent ) error {
    if sr.couldStartWithEvent( ev ) { return nil }    
    return sr.failTopType( ev )
}

func ( sr *StructuralReactor ) failUnexpectedMapEnd( val interface{} ) error {
    desc := sr.sawDescFor( val )
    return rctErrorf( "Expected field name or end of fields but got %s", desc )
}

func ( sr *StructuralReactor ) execValueCheck( 
    ev ReactorEvent, pushIfOk interface{} ) error {

    top := sr.stack.Peek()

    switch v := top.( type ) {
    case nil, listAccType, *Identifier:
        if v == nil {
            if err := sr.checkTopType( ev ); err != nil { return err }
        }
        if pushIfOk != nil { sr.stack.Push( pushIfOk ) }
        return nil
    case *mapStructureCheck: return sr.failUnexpectedMapEnd( ev )
    }

    return rctErrorf( "Saw %s while expecting %s", 
        sr.sawDescFor( ev ), sr.expectDescFor( top ) );
}

func ( sr *StructuralReactor ) completeValue() {
    if _, ok := sr.stack.Peek().( *Identifier ); ok { sr.stack.Pop() }
    sr.done = sr.stack.IsEmpty()
}

func ( sr *StructuralReactor ) checkValue( ev ReactorEvent ) error {
    if err := sr.execValueCheck( ev, nil ); err != nil { return err }
    sr.completeValue()
    return nil
}

func ( sr *StructuralReactor ) checkStructureStart( ev ReactorEvent ) error {
    return sr.execValueCheck( ev, newMapStructureCheck() )
}

func ( sr *StructuralReactor ) checkFieldStart( fs *FieldStartEvent ) error {
    if sr.stack.IsEmpty() { return sr.failTopType( fs ) }
    top := sr.stack.Peek()
    if mc, ok := top.( *mapStructureCheck ); ok {
        if err := mc.startField( fs.Field ); err != nil { return err }
        sr.stack.Push( fs.Field )
        return nil
    }
    return rctErrorf( "Saw start of field '%s' while expecting %s",
        fs.Field.ExternalForm(), sr.expectDescFor( top ) )
}

func ( sr *StructuralReactor ) checkListStart( le *ListStartEvent ) error {
    return sr.execValueCheck( le, listAccType( 1 ) )
}

func ( sr *StructuralReactor ) checkEnd( ee *EndEvent ) error {
    if sr.stack.IsEmpty() { return sr.failTopType( ee ) }
    top := sr.stack.Peek()
    switch top.( type ) {
    case *mapStructureCheck, listAccType:
        sr.stack.Pop()
        sr.completeValue()
        return nil
    }
    return rctErrorf( "Saw end while expecting %s", sr.expectDescFor( top ) )
}

func ( sr *StructuralReactor ) ProcessEvent( ev ReactorEvent ) error {
    if err := sr.checkNotDone( ev ); err != nil { return err }
    switch v := ev.( type ) {
    case *ValueEvent: return sr.checkValue( v )
    case *StructStartEvent, *MapStartEvent: return sr.checkStructureStart( ev )
    case *FieldStartEvent: return sr.checkFieldStart( v )
    case *EndEvent: return sr.checkEnd( v )
    case *ListStartEvent: return sr.checkListStart( v )
    default: panic( libErrorf( "unhandled event: %T", ev ) )
    }
    return nil
}

func EnsureStructuralReactor( pip *pipeline.Pipeline ) {
    var sr *StructuralReactor
    pip.VisitReverse( func ( p interface{} ) {
        if sr != nil { return }
        sr, _ = p.( *StructuralReactor )
    })
    if sr == nil { pip.Add( NewStructuralReactor( ReactorTopTypeValue ) ) }
}

func EnsurePathSettingProcessor( pip *pipeline.Pipeline ) {
    var ps *PathSettingProcessor
    pip.VisitReverse( func( p interface{} ) {
        if ps != nil { return }
        ps, _ = p.( *PathSettingProcessor )
    })
    if ps == nil { pip.Add( NewPathSettingProcessor() ) }
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

func ( proc *PathSettingProcessor ) InitializePipeline( 
    pip *pipeline.Pipeline ) {

    if ! proc.skipStructureCheck { EnsureStructuralReactor( pip ) }
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
    case *ValueEvent: proc.prepareValue()
    case *ListStartEvent: proc.prepareListStart()
    case *MapStartEvent: proc.prepareStructure( endTypeMap )
    case *StructStartEvent: proc.prepareStructure( endTypeStruct )
    case *FieldStartEvent: proc.prepareStartField( v.Field )
    case *EndEvent: proc.prepareEnd()
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
    case *ValueEvent: proc.processedValue()
    case *EndEvent: proc.processedEnd()
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
    accs *stack.Stack
}

func newValueAccumulator() *valueAccumulator {
    return &valueAccumulator{ accs: stack.NewStack() }
}

func ( va *valueAccumulator ) pushAcc( acc accImpl ) {
    va.accs.Push( acc )
}

func ( va *valueAccumulator ) peekAcc() ( accImpl, bool ) {
    if va.accs.IsEmpty() { return nil, false }
    return va.accs.Peek().( accImpl ), true
}

func ( va *valueAccumulator ) popAcc() accImpl {
    res, ok := va.peekAcc()
    if ! ok { panic( libErrorf( "popAcc() called on empty stack" ) ) }
    va.accs.Pop()
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
    case *ValueEvent: va.valueReady( v.Val )
    case *ListStartEvent: va.pushAcc( newListAcc() )
    case *MapStartEvent: va.pushAcc( newMapAcc() )
    case *StructStartEvent: va.pushAcc( newStructAcc( v.Type ) )
    case *FieldStartEvent: va.startField( v.Field )
    case *EndEvent: if err := va.end(); err != nil { return err }
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

func ( cr *CastReactor ) InitializePipeline( pip *pipeline.Pipeline ) {
    EnsureStructuralReactor( pip )
    EnsurePathSettingProcessor( pip )
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
    ve *ValueEvent,
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
    ve *ValueEvent,
    nt *NullableTypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    if _, ok := ve.Val.( *Null ); ok { return next.ProcessEvent( ve ) }
    return cr.processValueWithType( ve, nt.Type, callTyp, next )
}

func ( cr *CastReactor ) processValueWithType(
    ve *ValueEvent,
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
    ve *ValueEvent, next ReactorEventProcessor ) error {

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
    return next.ProcessEvent( ev )
}

func ( cr *CastReactor ) completeStartStruct(
    ss *StructStartEvent, next ReactorEventProcessor ) error {

    ft, err := cr.iface.FieldTyperFor( ss.Type, ss.GetPath() )
    if err != nil { return err }

    if ft == nil { ft = valueFieldTyper( 1 ) }
    return cr.implStartMap( ss, ft, next )
}

func ( cr *CastReactor ) inferStructForMap(
    me *MapStartEvent,
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
    me *MapStartEvent,
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
    me *MapStartEvent, 
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
    me *MapStartEvent, next ReactorEventProcessor ) error {
    
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
    fs *FieldStartEvent, next ReactorEventProcessor ) error {

    ft := cr.stack.Peek().( FieldTyper )
    
    typ, err := ft.FieldTypeFor( fs.Field, fs.GetPath().Parent() )
    if err != nil { return err }

    cr.stack.Push( typ )
    return next.ProcessEvent( fs )
}

func ( cr *CastReactor ) processEnd(
    ee *EndEvent, next ReactorEventProcessor ) error {

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
    ss *StructStartEvent,
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
    ss *StructStartEvent,
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
    ss *StructStartEvent, next ReactorEventProcessor ) error {

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
    le *ListStartEvent,
    at *AtomicTypeReference,
    callTyp TypeReference,
    next ReactorEventProcessor ) error {

    if at.Equals( TypeValue ) {
        return cr.processListStartWithType( le, TypeOpaqueList, callTyp, next )
    }

    return NewTypeCastError( callTyp, TypeOpaqueList, le.GetPath() )
}

func ( cr *CastReactor ) processListStartWithType(
    le *ListStartEvent,
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
    le *ListStartEvent, next ReactorEventProcessor ) error {

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
    case *ValueEvent: return cr.processValue( v, next )
    case *MapStartEvent: return cr.processStartMap( v, next )
    case *FieldStartEvent: return cr.processFieldStart( v, next )
    case *StructStartEvent: return cr.processStructStart( v, next )
    case *ListStartEvent: return cr.processListStart( v, next )
    case *EndEvent: return cr.processEnd( v, next )
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

func ( fo *FieldOrderReactor ) InitializePipeline( pip *pipeline.Pipeline ) {
    EnsureStructuralReactor( pip )
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
    ev *FieldStartEvent ) error {

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
    case *StructStartEvent: return sp.processStructStart( v )
    case *FieldStartEvent: return sp.processFieldStart( v )
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

func ( sp *structOrderProcessor ) processValue( v *ValueEvent ) error {
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
    ev *StructStartEvent, next ReactorEventProcessor,
) error {
    ord := fo.fog.FieldOrderFor( ev.Type )
    if ord == nil { return fo.processContainerStart( ev, next ) }
    sp := &structOrderProcessor{ ord: ord }
    sp.next = fo.structOrderGetNextProc( next )
    return fo.pushProc( sp, ev )
}

func ( fo *FieldOrderReactor ) processValue( 
    v *ValueEvent, next ReactorEventProcessor ) error {

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
    case *ListStartEvent, *MapStartEvent: 
        return fo.processContainerStart( v, next )
    case *StructStartEvent: return fo.processStructStart( v, next )
    case *FieldStartEvent: return fo.peekProc().ProcessEvent( v )
    case *ValueEvent: return fo.processValue( v, next )
    case *EndEvent: return fo.processEnd( v, next )
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

    Namespace( ns *Namespace, path objpath.PathNode ) error

    Service( svc *Identifier, path objpath.PathNode ) error

    Operation( op *Identifier, path objpath.PathNode ) error

    GetAuthenticationProcessor( 
        path objpath.PathNode ) ( ReactorEventProcessor, error )

    GetParametersProcessor( 
        path objpath.PathNode ) ( ReactorEventProcessor, error )
}

type requestFieldType int

// declared in the preferred arrival order
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

    // 0: before StartStruct{ QnameServiceRequest } and after final *EndEvent
    //
    // 1: when reading or expecting a service request field (namespace, service,
    // etc)
    //
    // > 1: accumulating some nested value for 'parameters' or 'authentication' 
    depth int 

    fld requestFieldType

    hadParams bool // true if the input contained explicit params
//    paramsSynth bool // true when we are synthesizing empty params
}

func NewServiceRequestReactor( 
    iface ServiceRequestReactorInterface ) *ServiceRequestReactor {
    return &ServiceRequestReactor{ iface: iface }
}

func ( sr *ServiceRequestReactor ) updateEvProc( ev ReactorEvent ) {
    switch ev.( type ) {
    case *FieldStartEvent: return
    case *StructStartEvent, *ListStartEvent, *MapStartEvent: sr.depth++
    case *EndEvent: sr.depth--
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

func ( sr *ServiceRequestReactor ) InitializePipeline( 
    pip *pipeline.Pipeline ) {
    
    EnsureStructuralReactor( pip ) 
    EnsurePathSettingProcessor( pip )
    pip.Add( NewCastReactor( TypeServiceRequest, svcReqCastIface( 1 ) ) )
    pip.Add( NewFieldOrderReactor( svcReqFieldOrderGetter( 1 ) ) )
}

func ( sr *ServiceRequestReactor ) invalidValueErr( 
    path objpath.PathNode, desc string ) error {

    return NewValueCastErrorf( path, "invalid value: %s", desc )
}

func ( sr *ServiceRequestReactor ) startStruct( ev *StructStartEvent ) error {
    if sr.fld == reqFieldNone { // we're at the top of the request
        if ev.Type.Equals( QnameServiceRequest ) { return nil }
        // panic because upstream cast should have checked already
        panic( libErrorf( "Unexpected service request type: %s", ev.Type ) )
    }
    return sr.invalidValueErr( ev.GetPath(), ev.Type.ExternalForm() )
}

func ( sr *ServiceRequestReactor ) checkStartField( fs *FieldStartEvent ) {
    if sr.fld == reqFieldNone { return }
    panic( libErrorf( "Saw field start '%s' while sr.fld is %d", 
        fs.Field, sr.fld ) )
}

func ( sr *ServiceRequestReactor ) startField( 
    fs *FieldStartEvent ) ( err error ) {

    sr.checkStartField( fs )
    switch fld := fs.Field; {
    case fld.Equals( IdNamespace ): sr.fld = reqFieldNs
    case fld.Equals( IdService ): sr.fld = reqFieldSvc
    case fld.Equals( IdOperation ): sr.fld = reqFieldOp
    case fld.Equals( IdAuthentication ): 
        sr.fld = reqFieldAuth
        sr.evProc, err = sr.iface.GetAuthenticationProcessor( fs.GetPath() )
    case fld.Equals( IdParameters ): 
        sr.fld = reqFieldParams
        sr.evProc, err = sr.iface.GetParametersProcessor( fs.GetPath() )
        if err == nil { sr.hadParams = true }
    default: err = NewUnrecognizedFieldError( fs.GetPath(), fs.Field )
    }
    return
}

func ( sr *ServiceRequestReactor ) getFieldValueForString(
    s string, 
    path objpath.PathNode, 
    reqFld requestFieldType ) ( res interface{}, err error ) {

    switch reqFld {
    case reqFieldNs: res, err = ParseNamespace( s )
    case reqFieldSvc, reqFieldOp: res, err = ParseIdentifier( s )
    default:
        panic( libErrorf( "Unhandled req fld type for string: %d", reqFld ) )
    }
    if err != nil { err = NewValueCastError( path, err.Error() ) }
    return
}

func ( sr *ServiceRequestReactor ) getFieldValueForBuffer(
    buf []byte, 
    path objpath.PathNode,
    reqFld requestFieldType ) ( res interface{}, err error ) {

    bin := NewReader( bytes.NewReader( buf ) )
    switch reqFld {
    case reqFieldNs: res, err = bin.ReadNamespace()
    case reqFieldSvc, reqFieldOp: res, err = bin.ReadIdentifier()
    default:
        panic( libErrorf( "Unhandled req fld type for buffer: %d", reqFld ) )
    }
    if err != nil { err = NewValueCastError( path, err.Error() ) }
    return
}

func ( sr *ServiceRequestReactor ) getFieldValue( 
    ve *ValueEvent, reqFld requestFieldType ) ( interface{}, error ) {
    path := ve.GetPath()
    switch v := ve.Val.( type ) {
    case String: return sr.getFieldValueForString( string( v ), path, reqFld )
    case Buffer: return sr.getFieldValueForBuffer( []byte( v ), path, reqFld )
    }
    return nil, sr.invalidValueErr( path, TypeOf( ve.Val ).ExternalForm() )
}

func ( sr *ServiceRequestReactor ) namespace( ve *ValueEvent ) error {
    ns, err := sr.getFieldValue( ve, reqFieldNs )
    if err == nil { 
        return sr.iface.Namespace( ns.( *Namespace ), ve.GetPath() )
    }
    return err
}

func ( sr *ServiceRequestReactor ) readIdent( 
    ve *ValueEvent, reqFld requestFieldType ) error {

    v2, err := sr.getFieldValue( ve, reqFld )
    if err != nil { return err }
    id := v2.( *Identifier )
    path := ve.GetPath()
    switch reqFld {
    case reqFieldSvc: return sr.iface.Service( id, path )
    case reqFieldOp: return sr.iface.Operation( id, path )
    default: panic( libErrorf( "Unhandled req fld type: %d", reqFld ) )
    }
    return nil
}

func ( sr *ServiceRequestReactor ) value( ve *ValueEvent ) error {
    defer func() { sr.fld = reqFieldNone }()
    switch sr.fld {
    case reqFieldNs: return sr.namespace( ve )
    case reqFieldSvc, reqFieldOp: return sr.readIdent( ve, sr.fld )
    }
    panic( libErrorf( "Unhandled req field type: %d", sr.fld ) )
}

func ( sr *ServiceRequestReactor ) visitSyntheticParams(
    rct ReactorEventProcessor, startPath objpath.PathNode ) error {
    ps := NewPathSettingProcessor()
    ps.skipStructureCheck = true
    var path objpath.PathNode
    if startPath == nil { 
        path = objpath.RootedAt( IdParameters ) 
    } else { 
        path = startPath.Descend( IdParameters ) 
    }
    ps.SetStartPath( path )
    pip := InitReactorPipeline( ps, rct )
    return VisitValue( EmptySymbolMap(), pip )
}

func ( sr *ServiceRequestReactor ) end( ee *EndEvent ) error {
    if sr.hadParams { return nil }
    ep, err := sr.iface.GetParametersProcessor( ee.GetPath() );
    if err != nil { return err }
    return sr.visitSyntheticParams( ep, ee.GetPath() )
}

func ( sr *ServiceRequestReactor ) ProcessEvent( ev ReactorEvent ) error {
    defer sr.updateEvProc( ev )
    if sr.evProc != nil { return sr.evProc.ProcessEvent( ev ) }
    switch v := ev.( type ) {
    case *FieldStartEvent: return sr.startField( v )
    case *StructStartEvent: return sr.startStruct( v )
    case *ValueEvent: return sr.value( v )
    case *ListStartEvent: 
        return sr.invalidValueErr( v.GetPath(), TypeOpaqueList.ExternalForm() )
    case *MapStartEvent: 
        return sr.invalidValueErr( v.GetPath(), TypeSymbolMap.ExternalForm() )
    case *EndEvent: return sr.end( v )
    default: panic( libErrorf( "Unhandled event: %T", v ) )
    }
    return nil
}

type ServiceResponseReactorInterface interface {
    GetResultProcessor( path objpath.PathNode ) ( ReactorEventProcessor, error )
    GetErrorProcessor( path objpath.PathNode ) ( ReactorEventProcessor, error )
}

type ServiceResponseReactor struct {

    iface ServiceResponseReactorInterface

    evProc ReactorEventProcessor

    // depth is similar to in ServiceRequestReactor
    depth int
    
    hadProc bool
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

func ( sr *ServiceResponseReactor ) InitializePipeline( 
    pip *pipeline.Pipeline ) {

    EnsureStructuralReactor( pip )
    pip.Add( NewCastReactor( TypeServiceResponse, svcRespCastIface( 1 ) ) )
}

func ( sr *ServiceResponseReactor ) updateEvProc( ev ReactorEvent ) {
    switch ev.( type ) {
    case *FieldStartEvent: return
    case *StructStartEvent, *MapStartEvent, *ListStartEvent: sr.depth++
    case *EndEvent: sr.depth--
    }
    if sr.depth == 1 { 
        if sr.evProc != nil { sr.hadProc, sr.evProc = true, nil }
    }
}

// Note that the error path uses Parent() since we'll be positioned on the field
// (result/error) that is the second value, but the error, if we have one, is
// really at the response level itself
func ( sr *ServiceResponseReactor ) sendEvProcEvent( ev ReactorEvent ) error {
    shouldFail := sr.hadProc
    if shouldFail {
        if ve, ok := ev.( *ValueEvent ); ok {
            if _, isNull := ve.Val.( *Null ); isNull { shouldFail = false }
        }
    }
    if shouldFail {
        msg := "response has both a result and an error value"
        return NewValueCastError( ev.GetPath().Parent(), msg )
    }
    return sr.evProc.ProcessEvent( ev )
}

func ( sr *ServiceResponseReactor ) startStruct( t *QualifiedTypeName ) error {
    if t.Equals( QnameServiceResponse ) { return nil }
    panic( libErrorf( "Got unexpected (toplevel) struct type: %s", t ) )
}

func ( sr *ServiceResponseReactor ) startField( fs *FieldStartEvent ) error {
    var err error
    fld, path := fs.Field, fs.GetPath()
    switch {
    case fld.Equals( IdResult ): 
        sr.evProc, err = sr.iface.GetResultProcessor( path )
    case fld.Equals( IdError ): 
        sr.evProc, err = sr.iface.GetErrorProcessor( path )
    default: return NewUnrecognizedFieldError( path.Parent(), fld )
    }
    return err
}

func ( sr *ServiceResponseReactor ) ProcessEvent( ev ReactorEvent ) error {
    defer sr.updateEvProc( ev )
    if sr.evProc != nil { return sr.sendEvProcEvent( ev ) }
    switch v := ev.( type ) {
    case *StructStartEvent: return sr.startStruct( v.Type )
    case *FieldStartEvent: return sr.startField( v )
    case *EndEvent: return nil
    }
    evStr := EventToString( ev )
    panic( libErrorf( "Saw event %s while evProc == nil", evStr ) )
}

type DebugLogger interface {
    Log( msg string )
}

type DebugLoggerFunc func( string )

func ( f DebugLoggerFunc ) Log( msg string ) { f( msg ) }

type DebugReactor struct { 
    l DebugLogger 
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
