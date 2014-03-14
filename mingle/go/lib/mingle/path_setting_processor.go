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

type psListContext struct {
    basePath objpath.PathNode
    path *objpath.ListNode
}

//type endType int
//
//const (
//    endTypeList = endType( iota )
//    endTypeMap
//    endTypeStruct
//    endTypeField
//)
//
//type psStackElt struct {
//    et endType
//    awaitingList0 bool
//    basePath objpath.PathNode
//    path objpath.PathNode
//}

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

//func ( proc *PathSettingProcessor ) inspect() string {
//    return fmt.Sprintf( "<%T path: %s>", proc, proc.inspectPath() )
//}

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

//func ( proc *PathSettingProcessor ) pathPop() {
//    if proc.path != nil { proc.path = proc.path.Parent() }
//}

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

func ( proc *PathSettingProcessor ) prepareListStart() {
    proc.prepareValue() // this list may be itself be a value in another list
    proc.stack.Push( &psListContext{ basePath: proc.topPath() } )
}

func ( proc *PathSettingProcessor ) prepareStructure() {
    proc.prepareValue()
    proc.stack.Push( psMapContext{ path: proc.topPath() } )
//    proc.stack.Push( &psFieldContext{ basePath: proc.topPath() } )
}

func ( proc *PathSettingProcessor ) prepareStartField( f *Identifier ) {
//    fldCtx := proc.stack.Peek().( *psFieldContext )
    fldPath := objpath.Descend( proc.topPath(), f )
    proc.stack.Push( psFieldContext{ path: fldPath } )
}

func ( proc *PathSettingProcessor ) prepareEnd() {
    if proc.stack.IsEmpty() { return }
    if _, ok := proc.stack.Peek().( *psListContext ); ok { proc.stack.Pop() }
}

func ( proc *PathSettingProcessor ) prepareEvent( ev ReactorEvent ) {
    switch v := ev.( type ) {
    case *ValueEvent: proc.prepareValue()
    case *ListStartEvent: proc.prepareListStart()
    case *MapStartEvent, *StructStartEvent: proc.prepareStructure()
    case *FieldStartEvent: proc.prepareStartField( v.Field )
    case *EndEvent: proc.prepareEnd()
    }
    if path := proc.topPath(); path != nil { ev.SetPath( path ) }
}

func ( proc *PathSettingProcessor ) processedValue() {
    if proc.stack.IsEmpty() { return }
    if _, ok := proc.stack.Peek().( psFieldContext ); ok { proc.stack.Pop() }
}

func ( proc *PathSettingProcessor ) processedEnd() {
    if proc.stack.IsEmpty() { return }
    if _, ok := proc.stack.Peek().( psMapContext ); ok { proc.stack.Pop() }
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
