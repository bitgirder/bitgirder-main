package reactor

import (
    mg "mingle"
    "bitgirder/stack"
    "bitgirder/pipeline"
//    "log"
)

type ValueBuilder struct {
    val mg.Value
    accs *stack.Stack
}

func NewValueBuilder() *ValueBuilder {
    return &ValueBuilder{ accs: stack.NewStack() }
}

func ( vb *ValueBuilder ) InitializePipeline( p *pipeline.Pipeline ) {
    EnsureStructuralReactor( p )
}

// Panics if result of val is not ready
func ( vb *ValueBuilder ) GetValue() mg.Value {
    if vb.val == nil { 
        panic( NewReactorErrorf( nil, "Value is not yet built" ) ) 
    }
    return vb.val
}

type mapAcc struct { 
    m *mg.SymbolMap
    curFld *mg.Identifier
}

func newMapAcc() *mapAcc { return &mapAcc{ m: mg.NewSymbolMap() } }

// asserts that ma.curFld is non-nil, clears it, and returns it
func ( ma *mapAcc ) clearField() *mg.Identifier {
    if ma.curFld == nil { panic( libError( "no current field" ) ) }
    res := ma.curFld
    ma.curFld = nil
    return res
}

func ( ma *mapAcc ) setValue( val mg.Value ) { 
    ma.m.Put( ma.clearField(), val ) 
}

func ( ma *mapAcc ) startField( fld *mg.Identifier ) {
    if cur := ma.curFld; cur != nil {
        panic( libErrorf( "saw start of field %s while cur is %s", fld, cur ) )
    }
    ma.curFld = fld
}

type structAcc struct {
    flds *mapAcc
    s *mg.Struct
}

func newStructAcc( typ *mg.QualifiedTypeName ) *structAcc {
    res := &structAcc{ flds: newMapAcc() }
    res.s = &mg.Struct{ Type: typ, Fields: res.flds.m }
    return res
}    

type listAcc struct { 
    l *mg.List 
}

func newListAcc() *listAcc { 
    return &listAcc{ l: mg.NewList( mg.TypeOpaqueList ) } 
}

func ( vb *ValueBuilder ) pushAcc( acc interface{} ) { vb.accs.Push( acc ) }

func ( vb *ValueBuilder ) peekAcc() ( interface{}, bool ) {
    return vb.accs.Peek(), ! vb.accs.IsEmpty()
}

func ( vb *ValueBuilder ) mustPeekAcc() interface{} {
    if res, ok := vb.peekAcc(); ok { return res }
    panic( libError( "acc stack is empty" ) )
}

func ( vb *ValueBuilder ) acceptValue( 
    acc interface{}, val mg.Value ) bool {

    switch v := acc.( type ) {
    case *mapAcc: v.setValue( val ); return false
    case *structAcc: v.flds.setValue( val ); return false
    case *listAcc: v.l.AddUnsafe( val ); return false
    }
    panic( libErrorf( "unhandled acc: %T", acc ) )
}

func ( vb *ValueBuilder ) valueForAcc( acc interface{} ) mg.Value {
    switch v := acc.( type ) {
    case *mapAcc: return v.m
    case *structAcc: return v.s
    case *listAcc: return v.l
    }
    panic( libErrorf( "unhandled acc: %T", acc ) )
}

func ( vb *ValueBuilder ) popAccValue() {
    vb.valueReady( vb.valueForAcc( vb.accs.Pop() ) )
}

func ( vb *ValueBuilder ) valueReady( val mg.Value ) {
    if acc, ok := vb.peekAcc(); ok {
        if vb.acceptValue( acc, val ) { vb.popAccValue() }
    } else {
        vb.val = val
    }
}

func ( vb *ValueBuilder ) startField( fld *mg.Identifier ) {
    acc, ok := vb.peekAcc()
    if ! ok { panic( libErrorf( "got field start %s with empty stack", fld ) ) }
    switch v := acc.( type ) {
    case *mapAcc: v.startField( fld )
    case *structAcc: v.flds.startField( fld )
    default:
        panic( libErrorf( "unexpected acc for start of field %s: %T", fld, v ) )
    }
}

func ( vb *ValueBuilder ) end() { vb.popAccValue() }

func ( vb *ValueBuilder ) ProcessEvent( ev ReactorEvent ) error {
    switch v := ev.( type ) {
    case *ValueEvent: vb.valueReady( v.Val )
    case *ListStartEvent: vb.pushAcc( newListAcc() )
    case *MapStartEvent: vb.pushAcc( newMapAcc() )
    case *StructStartEvent: vb.pushAcc( newStructAcc( v.Type ) )
    case *FieldStartEvent: vb.startField( v.Field )
    case *EndEvent: vb.end()
    default: panic( libErrorf( "Unhandled event: %T", ev ) )
    }
    return nil
}
