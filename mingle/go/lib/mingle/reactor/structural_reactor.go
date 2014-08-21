package reactor

import (
    mg "mingle"
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

type listStructureCheck struct { 
    typ mg.TypeReference 
}

type mapStructureCheck struct { 
    seen *mg.IdentifierMap 
}

func newMapStructureCheck() *mapStructureCheck {
    return &mapStructureCheck{ seen: mg.NewIdentifierMap() }
}

func ( mc *mapStructureCheck ) startField( fld *mg.Identifier ) error {
    if mc.seen.HasKey( fld ) {
        return NewReactorErrorf( nil, "Multiple entries for field: %s", 
            fld.ExternalForm() )
    }
    mc.seen.Put( fld, true )
    return nil
}

func ( sr *StructuralReactor ) descForEvent( ev ReactorEvent ) string {
    switch v := ev.( type ) {
    case *ListStartEvent: return sr.sawDescFor( v.Type )
    case *MapStartEvent: return mg.TypeSymbolMap.ExternalForm()
    case *EndEvent: return "end"
    case *ValueEvent: return mg.TypeOf( v.Val ).ExternalForm()
    case *FieldStartEvent: return sr.sawDescFor( v.Field )
    case *StructStartEvent: return sr.sawDescFor( v.Type )
    }
    panic( libErrorf( "unhandled event: %T", ev ) )
}

func ( sr *StructuralReactor ) expectDescFor( val interface{} ) string {
    if val == nil { return "BEGIN" }
    switch v := val.( type ) {
    case *mg.Identifier: 
        return fmt.Sprintf( "a value for field '%s'", v.ExternalForm() )
    case listStructureCheck: return "a list value"
    }
    panic( libErrorf( "unhandled desc value: %T", val ) )
}

func ( sr *StructuralReactor ) sawDescFor( val interface{} ) string {
    if val == nil { return "BEGIN" }
    switch v := val.( type ) {
    case *mg.Identifier: 
        return fmt.Sprintf( "start of field '%s'", v.ExternalForm() )
    case *mg.QualifiedTypeName: 
        return fmt.Sprintf( "start of struct %s", v.ExternalForm() )
    case *mg.ListTypeReference:
        return fmt.Sprintf( "start of %s", v.ExternalForm() )
    case ReactorEvent: return sr.descForEvent( v )
    }
    panic( libErrorf( "unhandled val: %T", val ) )
}

func ( sr *StructuralReactor ) checkNotDone( ev ReactorEvent ) error {
    if ! sr.done { return nil }
    return NewReactorErrorf( ev.GetPath(), "Saw %s after value was built", 
        sr.sawDescFor( ev ) );
}

func ( sr *StructuralReactor ) failTopType( ev ReactorEvent ) error {
    desc := sr.descForEvent( ev )
    return NewReactorErrorf( 
        ev.GetPath(), "Expected %s but got %s", sr.topTyp, desc )
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

func ( sr *StructuralReactor ) push( val interface{} ) { sr.stack.Push( val ) }

func ( sr *StructuralReactor ) failUnexpectedMapEnd( val interface{} ) error {
    desc := sr.sawDescFor( val )
    return NewReactorErrorf( nil, 
        "Expected field name or end of fields but got %s", desc )
}

func ( sr *StructuralReactor ) listValueTypeError( 
    expct mg.TypeReference, ev ReactorEvent ) error {

    return NewReactorErrorf( nil, "expected list value of type %s but saw %s",
        expct, sr.sawDescFor( ev ) )
}

// drops pointer and restriction expectations from effectiveType, checking only
// that the structure of the assignment would make sense (downstream casts can
// check the actual validity of the assignment)
func ( sr *StructuralReactor ) recursiveCheckValueTypeForList(
    calledType, effectiveType, valType mg.TypeReference, 
    ev ReactorEvent ) error {
    
    switch typ := effectiveType.( type ) {
    case *mg.PointerTypeReference:
        return sr.recursiveCheckValueTypeForList( 
            calledType, typ.Type, valType, ev )
    case *mg.AtomicTypeReference:
        if typ.Restriction != nil {
            return sr.recursiveCheckValueTypeForList(
                calledType, typ.Name.AsAtomicType(), valType, ev )
        }
    }
    if mg.CanAssignType( valType, effectiveType ) { return nil }
    return sr.listValueTypeError( calledType, ev )
}

func ( sr *StructuralReactor ) checkValueTypeForList(
    lc listStructureCheck, typ mg.TypeReference, ev ReactorEvent ) error {

    return sr.recursiveCheckValueTypeForList( lc.typ, lc.typ, typ, ev )
}

func ( sr *StructuralReactor ) checkEventForList(
    lc listStructureCheck, ev ReactorEvent ) error {

    switch v := ev.( type ) {
    case *ValueEvent: 
        return sr.checkValueTypeForList( lc, mg.TypeOf( v.Val ), v )
    case *ListStartEvent: 
        return sr.checkValueTypeForList( lc, v.Type, v )
    case *MapStartEvent: 
        return sr.checkValueTypeForList( lc, mg.TypeSymbolMap, ev )
    case *StructStartEvent: 
        return sr.checkValueTypeForList( lc, v.Type.AsAtomicType(), ev )
    }
    return nil
}

func ( sr *StructuralReactor ) execValueCheck( 
    ev ReactorEvent, pushIfOk interface{} ) ( err error ) {

    if sr.stack.IsEmpty() {
        err = sr.checkTopType( ev )
    } else {
        switch v := sr.stack.Peek().( type ) {
        case listStructureCheck: err = sr.checkEventForList( v, ev )
        case *mg.Identifier: break;
        case *mapStructureCheck: return sr.failUnexpectedMapEnd( ev )
        default: 
            err = NewReactorErrorf( ev.GetPath(), "Saw %s while expecting %s", 
                sr.sawDescFor( ev ), sr.expectDescFor( v ) );
        }
    }
    if err != nil { return }
    if pushIfOk != nil { sr.push( pushIfOk ) }
    return 
}

func ( sr *StructuralReactor ) completeValue() {
    if _, ok := sr.stack.Peek().( *mg.Identifier ); ok { sr.stack.Pop() }
    sr.done = sr.stack.IsEmpty()
}

func ( sr *StructuralReactor ) checkValue( ev ReactorEvent ) error {
    if err := sr.execValueCheck( ev, nil ); err != nil { return err }
    sr.completeValue()
    return nil
}

func ( sr *StructuralReactor ) checkStructureStart( ev ReactorEvent ) error {
    return sr.execValueCheck( ev,  newMapStructureCheck() )
}

func ( sr *StructuralReactor ) checkFieldStart( fs *FieldStartEvent ) error {
    if sr.stack.IsEmpty() { return sr.failTopType( fs ) }
    top := sr.stack.Peek()
    if mc, ok := top.( *mapStructureCheck ); ok {
        if err := mc.startField( fs.Field ); err != nil { return err }
        sr.push( fs.Field )
        return nil
    }
    return NewReactorErrorf( fs.GetPath(), 
        "Saw start of field '%s' while expecting %s",
        fs.Field.ExternalForm(), sr.expectDescFor( top ) )
}

func ( sr *StructuralReactor ) checkListStart( le *ListStartEvent ) error {
    lsc := listStructureCheck{ typ: le.Type.ElementType }
    return sr.execValueCheck( le, lsc )
}

func ( sr *StructuralReactor ) checkEnd( ee *EndEvent ) error {
    if sr.stack.IsEmpty() { return sr.failTopType( ee ) }
    top := sr.stack.Peek()
    switch top.( type ) {
    case *mapStructureCheck, listStructureCheck:
        sr.stack.Pop()
        sr.completeValue()
        return nil
    }
    return NewReactorErrorf( ee.GetPath(), 
        "Saw end while expecting %s", sr.expectDescFor( top ) )
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
