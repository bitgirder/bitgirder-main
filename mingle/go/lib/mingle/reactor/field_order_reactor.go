package reactor

import (
    mg "mingle"
    "bitgirder/objpath"
    "bitgirder/pipeline"
    "bitgirder/stack"
    "container/list"
//    "log"
)

type FieldOrderSpecification struct {
    Field *mg.Identifier
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
    FieldOrderFor( qn *mg.QualifiedTypeName ) FieldOrder
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
    acc []Event
}

func ( s *structOrderFieldState ) ProcessEvent( ev Event ) error {
    s.acc = append( s.acc, CopyEvent( ev, false ) )
    return nil
}

type structOrderProcessor struct {
    ord FieldOrder
    next EventProcessor
    startPath objpath.PathNode
    fieldQueue *list.List
    states *mg.IdentifierMap
    cur *structOrderFieldState
}

func ( sp *structOrderProcessor ) fieldReactor() EventProcessor {
    if sp.cur == nil || sp.cur.acc == nil { return sp.next }
    return sp.cur
}

func ( sp *structOrderProcessor ) processStructStart( ev Event ) error {
    if p := ev.GetPath(); p != nil { sp.startPath = objpath.CopyOf( p ) }
    sp.states = mg.NewIdentifierMap()
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
        sp.cur.acc = make( []Event, 0, 64 )
        return nil
    }
    return sp.next.ProcessEvent( ev )
}

func ( sp *structOrderProcessor ) ProcessEvent( ev Event ) error {
    switch v := ev.( type ) {
    case *StructStartEvent: return sp.processStructStart( v )
    case *FieldStartEvent: return sp.processFieldStart( v )
    }
    panic( libErrorf( "unexpected event: %T", ev ) )
}

func ( sp *structOrderProcessor ) getFieldSender() EventProcessor {

    ps := NewPathSettingProcessor()
    if p := sp.startPath; p != nil { ps.SetStartPath( objpath.CopyOf( p ) ) }
    ps.SkipStructureCheck = true
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

func ( sp *structOrderProcessor ) checkRequiredFields( ev Event ) error {
    missing := make( []*mg.Identifier, 0, 4 )
    sp.states.EachPair( func( k *mg.Identifier, v interface{} ) {
        state := v.( *structOrderFieldState )
        if state.spec.Required && ( ! state.seen ) {
            missing = append( missing, state.spec.Field )
        }
    })
    if len( missing ) == 0 { return nil }
    return mg.NewMissingFieldsError( ev.GetPath(), missing )
}

func ( sp *structOrderProcessor ) endStruct( ev Event ) error {
    if err := sp.sendReadyFields( true ); err != nil { return err }
    if err := sp.checkRequiredFields( ev ); err != nil { return err }
    return sp.next.ProcessEvent( ev )
}

func ( fo *FieldOrderReactor ) peekProc() EventProcessor {
    if fo.stack.IsEmpty() { return nil }
    return fo.stack.Peek().( EventProcessor )
}

func ( fo *FieldOrderReactor ) peekStructProc() *structOrderProcessor {
    rep := fo.peekProc()
    if rep == nil { return nil }
    if res, ok := rep.( *structOrderProcessor ); ok { return res }
    return nil
}

func ( fo *FieldOrderReactor ) pushProc( 
    next EventProcessor, ev Event ) error {

    fo.stack.Push( next )
    return next.ProcessEvent( ev )
}

func ( fo *FieldOrderReactor ) processContainerStart(
    ev Event, next EventProcessor ) error {

    if sp := fo.peekStructProc(); sp != nil {
        return fo.pushProc( sp.fieldReactor(), ev )
    }
    if ! fo.stack.IsEmpty() { next = fo.peekProc() }
    return fo.pushProc( next, ev )
}

func ( fo *FieldOrderReactor ) structOrderGetNextProc( 
    next EventProcessor ) EventProcessor {
    if fo.stack.IsEmpty() { return next }
    rep := fo.peekProc()
    if sp, ok := rep.( *structOrderProcessor ); ok { return sp.fieldReactor() }
    return rep
}

func ( fo *FieldOrderReactor ) processStructStart(
    ev *StructStartEvent, next EventProcessor,
) error {
    ord := fo.fog.FieldOrderFor( ev.Type )
    if ord == nil { return fo.processContainerStart( ev, next ) }
    sp := &structOrderProcessor{ ord: ord }
    sp.next = fo.structOrderGetNextProc( next )
    return fo.pushProc( sp, ev )
}

func ( fo *FieldOrderReactor ) processValue( 
    v *ValueEvent, next EventProcessor ) error {

    if sp := fo.peekStructProc(); sp != nil { return sp.processValue( v ) }
    return fo.peekProc().ProcessEvent( v )
}

func ( fo *FieldOrderReactor ) processEnd(
    ev Event, next EventProcessor ) error {
    rep := fo.stack.Pop().( EventProcessor )
    if sp, ok := rep.( *structOrderProcessor ); ok {
        if err := sp.endStruct( ev ); err != nil { return err }
    } else {
        if err := rep.ProcessEvent( ev ); err != nil { return err }
    }
    if sp := fo.peekStructProc(); sp != nil { return sp.valueEnded() }
    return nil
}

func ( fo *FieldOrderReactor ) processEventWithStack( 
    ev Event, next EventProcessor ) error {

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
    ev Event, next EventProcessor ) error {

    if fo.stack.IsEmpty() && ( ! isStructStart( ev ) ) {
        return next.ProcessEvent( ev )
    }
    return fo.processEventWithStack( ev, next )
}
