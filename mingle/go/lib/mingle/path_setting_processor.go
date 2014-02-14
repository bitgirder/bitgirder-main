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
    ev ReactorEvent, rep ReactorEventProcessor ) error {

    proc.prepareEvent( ev )
    if err := rep.ProcessEvent( ev ); err != nil { return err }
    proc.eventProcessed( ev )
    return nil
}
