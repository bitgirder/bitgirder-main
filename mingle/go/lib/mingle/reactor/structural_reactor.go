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
    id mg.PointerId
    typ mg.TypeReference 
}

type mapStructureCheck struct { 
    id mg.PointerId
    seen *mg.IdentifierMap 
}

func newMapStructureCheck() *mapStructureCheck {
    return &mapStructureCheck{ seen: mg.NewIdentifierMap() }
}

func ( mc *mapStructureCheck ) startField( fld *mg.Identifier ) error {
    if mc.seen.HasKey( fld ) {
        return rctErrorf( nil, "Multiple entries for field: %s", 
            fld.ExternalForm() )
    }
    mc.seen.Put( fld, true )
    return nil
}

type valAllocCheck struct { 
    id mg.PointerId
    typ mg.TypeReference 
}

func ( sr *StructuralReactor ) descForEvent( ev ReactorEvent ) string {
    switch v := ev.( type ) {
    case *ListStartEvent: return sr.sawDescFor( v.Type )
    case *MapStartEvent: return mg.TypeSymbolMap.ExternalForm()
    case *EndEvent: return "end"
    case *ValueEvent: return mg.TypeOf( v.Val ).ExternalForm()
    case *ValueAllocationEvent: return "allocation of &" + v.Type.ExternalForm()
    case *ValueReferenceEvent: return "reference"
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
    case valAllocCheck: return "allocation of &" + v.typ.ExternalForm()
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
    return rctErrorf( ev.GetPath(), "Saw %s after value was built", 
        sr.sawDescFor( ev ) );
}

func ( sr *StructuralReactor ) failTopType( ev ReactorEvent ) error {
    desc := sr.descForEvent( ev )
    return rctErrorf( ev.GetPath(), "Expected %s but got %s", sr.topTyp, desc )
}

func ( sr *StructuralReactor ) couldStartWithEvent( ev ReactorEvent ) bool {
    topIsVal := sr.topTyp == ReactorTopTypeValue
    switch ev.( type ) {
    case *ValueEvent, *ValueAllocationEvent: return topIsVal
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

func idForStructureCheck( val interface{} ) mg.PointerId { 
    switch v := val.( type ) {
    case listStructureCheck: return v.id
    case *mapStructureCheck: return v.id
    case valAllocCheck: return v.id
    }
    return mg.PointerIdNull
}

func ( sr *StructuralReactor ) checkNoCycle( id mg.PointerId ) ( err error ) {
    if id == mg.PointerIdNull { return nil }
    sr.stack.VisitTop( func ( elt interface{} ) {
        if err != nil { return }
        if eltId := idForStructureCheck( elt ); eltId == id {
            err = rctErrorf( nil, "reference %s is cyclic", id )
        }
    })
    return
}

func ( sr *StructuralReactor ) push( val interface{} ) { sr.stack.Push( val ) }

func ( sr *StructuralReactor ) failUnexpectedMapEnd( val interface{} ) error {
    desc := sr.sawDescFor( val )
    return rctErrorf( nil, 
        "Expected field name or end of fields but got %s", desc )
}

func ( sr *StructuralReactor ) listValueTypeError( 
    expct mg.TypeReference, ev ReactorEvent ) error {

    return rctErrorf( nil, "expected list value of type %s but saw %s",
        expct, sr.sawDescFor( ev ) )
}

func ( sr *StructuralReactor ) checkValueTypeForList(
    lc listStructureCheck, typ mg.TypeReference, ev ReactorEvent ) error {

    if mg.CanAssignType( typ, lc.typ ) { return nil }
    return sr.listValueTypeError( lc.typ, ev )
}

func ( sr *StructuralReactor ) checkValueEventForList(
    lc listStructureCheck, ve *ValueEvent ) error {

    if mg.CanAssign( ve.Val, lc.typ, false ) { return nil }
    return sr.listValueTypeError( lc.typ, ve )
}

func ( sr *StructuralReactor ) checkValueAllocForList(
    lc listStructureCheck, va *ValueAllocationEvent ) error {

    if isAssignableValueType( lc.typ ) { return nil }
    if lc.typ.Equals( mg.NewPointerTypeReference( va.Type ) ) { return nil }
    return sr.listValueTypeError( lc.typ, va )
}

func ( sr *StructuralReactor ) checkListStartEventForList(
    lc listStructureCheck, lse *ListStartEvent ) error {

    if isAssignableValueType( lc.typ ) || lc.typ.Equals( lse.Type ) { 
        return nil 
    }
    return sr.listValueTypeError( lc.typ, lse )
}

func ( sr *StructuralReactor ) checkEventForList(
    lc listStructureCheck, ev ReactorEvent ) error {

    switch v := ev.( type ) {
    case *ValueEvent: return sr.checkValueEventForList( lc, v )
    case *ValueAllocationEvent: return sr.checkValueAllocForList( lc, v )
    case *ListStartEvent: return sr.checkListStartEventForList( lc, v )
    case *MapStartEvent: 
        return sr.checkValueTypeForList( lc, mg.TypeSymbolMap, ev )
    case *StructStartEvent: 
        return sr.checkValueTypeForList( lc, v.Type.AsAtomicType(), ev )
    }
    return nil
}

func ( sr *StructuralReactor ) allocError(
    expct mg.TypeReference, ev ReactorEvent ) error {
    
    return rctErrorf( nil, "allocation of &%s followed by %s",
        expct, sr.descForEvent( ev ) )
}

// we don't check restrictions here, and leave that for a downstream reactor. we
// only check that the allocation is of an atomic type that matches that of the
// value
func ( sr *StructuralReactor ) checkValueEventForAlloc(
    expct mg.TypeReference, ve *ValueEvent ) error {

    if mg.CanAssign( ve.Val, expct, false ) { return nil }
    return sr.allocError( expct, ve )
}

func ( sr *StructuralReactor ) checkValueAllocEventForAlloc(
    expct mg.TypeReference, va *ValueAllocationEvent ) error {

    if pt, ok := expct.( *mg.PointerTypeReference ); ok {
        if pt.Equals( va.Type ) { return nil }
    }
    return sr.allocError( expct, va )
}

func ( sr *StructuralReactor ) checkValueAllocForStructStart(
    expct mg.TypeReference, sse *StructStartEvent ) error {
    
    if isAssignableValueType( expct ) { return nil }
    if at, ok := expct.( *mg.AtomicTypeReference ); ok {
        if at.Name.Equals( sse.Type ) { return nil }
    }
    return sr.allocError( expct, sse )
}

func ( sr *StructuralReactor ) checkEventForAlloc( 
    expct mg.TypeReference, ev ReactorEvent ) error {

    switch v := ev.( type ) {
    case *ValueEvent: return sr.checkValueEventForAlloc( expct, v )
    case *ValueAllocationEvent:
        return sr.checkValueAllocEventForAlloc( expct, v )
    case *ListStartEvent: if mg.CanAssignType( v.Type, expct ) { return nil }
    case *StructStartEvent: return sr.checkValueAllocForStructStart( expct, v )
    case *MapStartEvent: 
        if mg.CanAssignType( mg.TypeSymbolMap, expct ) { return nil }
    }
    return sr.allocError( expct, ev )
}

func ( sr *StructuralReactor ) execValueCheck( 
    ev ReactorEvent, pushIfOk interface{} ) ( err error ) {

    if sr.stack.IsEmpty() {
        err = sr.checkTopType( ev )
    } else {
        switch v := sr.stack.Peek().( type ) {
        case listStructureCheck: err = sr.checkEventForList( v, ev )
        case *mg.Identifier: break;
        case valAllocCheck: err = sr.checkEventForAlloc( v.typ, ev )
        case *mapStructureCheck: return sr.failUnexpectedMapEnd( ev )
        default: err = rctErrorf( ev.GetPath(), "Saw %s while expecting %s", 
            sr.sawDescFor( ev ), sr.expectDescFor( v ) );
        }
    }
    if err != nil { return }
    if pushIfOk != nil { sr.push( pushIfOk ) }
    return 
}

func ( sr *StructuralReactor ) completeValue() {
    for loop := ! sr.stack.IsEmpty(); loop; {
        if _, loop = sr.stack.Peek().( valAllocCheck ); loop { sr.stack.Pop() }
    }
    if _, ok := sr.stack.Peek().( *mg.Identifier ); ok { sr.stack.Pop() }
    sr.done = sr.stack.IsEmpty()
}

func ( sr *StructuralReactor ) checkValue( ev ReactorEvent ) error {
    if err := sr.execValueCheck( ev, nil ); err != nil { return err }
    sr.completeValue()
    return nil
}

func ( sr *StructuralReactor ) checkValueReference( 
    ev *ValueReferenceEvent ) error {

    if err := sr.checkNoCycle( ev.Id ); err != nil { return err }
    return sr.checkValue( ev )
}

func ( sr *StructuralReactor ) checkValueAlloc( 
    va *ValueAllocationEvent ) error {

    return sr.execValueCheck( va, valAllocCheck{ typ: va.Type, id: va.Id } )
}

func ( sr *StructuralReactor ) checkStructureStart( ev ReactorEvent ) error {
    chk := newMapStructureCheck()
    if ms, ok := ev.( *MapStartEvent ); ok { chk.id = ms.Id }
    return sr.execValueCheck( ev, chk )
}

func ( sr *StructuralReactor ) checkFieldStart( fs *FieldStartEvent ) error {
    if sr.stack.IsEmpty() { return sr.failTopType( fs ) }
    top := sr.stack.Peek()
    if mc, ok := top.( *mapStructureCheck ); ok {
        if err := mc.startField( fs.Field ); err != nil { return err }
        sr.push( fs.Field )
        return nil
    }
    return rctErrorf( fs.GetPath(), 
        "Saw start of field '%s' while expecting %s",
        fs.Field.ExternalForm(), sr.expectDescFor( top ) )
}

func ( sr *StructuralReactor ) checkListStart( le *ListStartEvent ) error {
    lsc := listStructureCheck{ typ: le.Type.ElementType, id: le.Id }
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
    return rctErrorf( ee.GetPath(), 
        "Saw end while expecting %s", sr.expectDescFor( top ) )
}

func ( sr *StructuralReactor ) ProcessEvent( ev ReactorEvent ) error {
    if err := sr.checkNotDone( ev ); err != nil { return err }
    switch v := ev.( type ) {
    case *ValueEvent: return sr.checkValue( v )
    case *ValueReferenceEvent: return sr.checkValueReference( v )
    case *ValueAllocationEvent: return sr.checkValueAlloc( v )
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
