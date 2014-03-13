package mingle

import (
    "fmt"
    "bitgirder/pipeline"
    "bitgirder/stack"
//    "log"
)

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
