package mingle

import (
    "bitgirder/objpath"
    "bitgirder/pipeline"
    "bitgirder/stack"
//    "log"
)

func EnsurePathSettingProcessor( pip *pipeline.Pipeline ) {
    var ps *PathSettingProcessor
    pip.VisitReverse( func( p interface{} ) {
        if ps != nil { return }
        ps, _ = p.( *PathSettingProcessor )
    })
    if ps == nil { pip.Add( NewPathSettingProcessor() ) }
}

type psFieldContext struct { path objpath.PathNode }

type psMapContext struct { path objpath.PathNode }

type psValueAllocContext struct { path objpath.PathNode }

type psListContext struct {
    basePath objpath.PathNode
    path *objpath.ListNode
}

type PathSettingProcessor struct {
    stack *stack.Stack
    startPath objpath.PathNode
    SkipStructureCheck bool
}

func NewPathSettingProcessor() *PathSettingProcessor {
    return &PathSettingProcessor{ stack: stack.NewStack() }
}

func inspectPath( path objpath.PathNode ) string {
    if path == nil { return "<nil>" }
    return FormatIdPath( path )
}

func ( proc *PathSettingProcessor ) SetStartPath( p objpath.PathNode ) {
    if p == nil { return }
    proc.startPath = objpath.CopyOf( p )
}

func NewPathSettingProcessorPath( p objpath.PathNode ) *PathSettingProcessor {
    res := NewPathSettingProcessor()
    res.SetStartPath( p )
    return res
}

func ( proc *PathSettingProcessor ) InitializePipeline( 
    pip *pipeline.Pipeline ) {

    if ! proc.SkipStructureCheck { EnsureStructuralReactor( pip ) }
}

func ( proc *PathSettingProcessor ) topPath() objpath.PathNode {
    if proc.stack.IsEmpty() { return proc.startPath }
    switch v := proc.stack.Peek().( type ) {
    case psValueAllocContext:
        if v.path == nil { return nil }
        return v.path
    case psFieldContext: 
        if v.path == nil { return nil }
        return v.path
    case psMapContext:
        if v.path == nil { return nil }
        return v.path
    case *psListContext: 
        if v.path != nil { return v.path }
        if v.basePath != nil { return v.basePath }
        return nil
    }
    panic( libErrorf( "unhandled stack element: %T", proc.stack.Peek() ) )
}

func ( proc *PathSettingProcessor ) updateList() {
    if proc.stack.IsEmpty() { return }
    if lc, ok := proc.stack.Peek().( *psListContext ); ok {
        if lc.path == nil {
            if lc.basePath == nil {
                lc.path = objpath.RootedAtList()
            } else {
                lc.path = lc.basePath.StartList()
            }
        } else {
            lc.path.Increment()
        }
    }
}

func ( proc *PathSettingProcessor ) prepareValue() { proc.updateList() }

func ( proc *PathSettingProcessor ) prepareValueAllocation() {
    proc.updateList()
    proc.stack.Push( psValueAllocContext{ path: proc.topPath() } )
}

func ( proc *PathSettingProcessor ) prepareListStart() {
    proc.prepareValue() // this list may be itself be a value in another list
    proc.stack.Push( &psListContext{ basePath: proc.topPath() } )
}

func ( proc *PathSettingProcessor ) prepareStructure() {
    proc.prepareValue()
    proc.stack.Push( psMapContext{ path: proc.topPath() } )
}

func ( proc *PathSettingProcessor ) prepareStartField( f *Identifier ) {
    fldPath := objpath.Descend( proc.topPath(), f )
    proc.stack.Push( psFieldContext{ path: fldPath } )
}

func ( proc *PathSettingProcessor ) prepareEnd() {
    if proc.stack.IsEmpty() { return }
    if _, ok := proc.stack.Peek().( *psListContext ); ok { proc.stack.Pop() }
}

func ( proc *PathSettingProcessor ) prepareEvent( ev ReactorEvent ) {
    switch v := ev.( type ) {
    case *ValueEvent, *ValueReferenceEvent: proc.prepareValue()
    case *ValueAllocationEvent: proc.prepareValueAllocation()
    case *ListStartEvent: proc.prepareListStart()
    case *MapStartEvent, *StructStartEvent: proc.prepareStructure()
    case *FieldStartEvent: proc.prepareStartField( v.Field )
    case *EndEvent: proc.prepareEnd()
    }
    if path := proc.topPath(); path != nil { ev.SetPath( path ) }
}

func ( proc *PathSettingProcessor ) optCompleteAllocation() {
    if _, ok := proc.stack.Peek().( psValueAllocContext ); ok { 
        proc.stack.Pop() 
    }
}

func ( proc *PathSettingProcessor ) processedValue() {
    if proc.stack.IsEmpty() { return }
    proc.optCompleteAllocation()
    if _, ok := proc.stack.Peek().( psFieldContext ); ok { proc.stack.Pop() }
}

func ( proc *PathSettingProcessor ) processedEnd() {
    if proc.stack.IsEmpty() { return }
    if _, ok := proc.stack.Peek().( psMapContext ); ok { proc.stack.Pop() }
    proc.optCompleteAllocation()
    proc.processedValue()
}

func ( proc *PathSettingProcessor ) eventProcessed( ev ReactorEvent ) {
    switch ev.( type ) {
    case *ValueEvent: proc.processedValue()
    case *EndEvent: proc.processedEnd()
    }
}

func ( proc *PathSettingProcessor ) ProcessEvent(
    ev ReactorEvent, rep ReactorEventProcessor ) error {

    proc.prepareEvent( ev )
    if err := rep.ProcessEvent( ev ); err != nil { return err }
    proc.eventProcessed( ev )
    return nil
}
