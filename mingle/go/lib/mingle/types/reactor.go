package types

import (
    mg "mingle"
    mgRct "mingle/reactor"
//    "log"
    "bitgirder/stack"
    "bitgirder/pipeline"
)

type castReactor struct {
    typ mg.TypeReference
    dm *DefinitionMap
    stack *stack.Stack
    skipPathSetter bool
}

func ( cr *castReactor ) InitializePipeline( pip *pipeline.Pipeline ) {
    mgCastRct := NewCastReactor0( cr.typ )
    mgCastRct.dm = cr.dm
    mgCastRct.SkipPathSetter = cr.skipPathSetter
    pip.Add( mgCastRct )
}

func ( cr *castReactor ) startStruct( 
    ss *mgRct.StructStartEvent ) ( mgRct.ReactorEvent, error ) {

    flds, err := fieldSetForTypeInDefMap( ss.Type, cr.dm, ss.GetPath() )
    if err != nil { return nil, err }
    if flds != nil { cr.stack.Push( newFieldCtx( flds ) ) }
    if def, ok := cr.dm.GetOk( ss.Type ); ok {
        if _, ok := def.( *SchemaDefinition ); ok {
            return asMapStartEvent( ss ), nil
        }
    } 
    return ss, nil
}

// We don't re-check here that fld is actually part of the defined field set or
// since the upstream processing will have validated that already
func ( cr *castReactor ) startField( 
    fs *mgRct.FieldStartEvent ) ( mgRct.ReactorEvent, error ) {

    if cr.stack.IsEmpty() { return fs, nil }
    cr.stack.Peek().( *fieldCtx ).await.Delete( fs.Field )
    return fs, nil
}

func ( cr *castReactor ) end( 
    ev *mgRct.EndEvent, 
    next mgRct.ReactorEventProcessor ) ( mgRct.ReactorEvent, error ) {

    if cr.stack.IsEmpty() { return ev, nil }
    fldCtx := cr.stack.Peek().( *fieldCtx )
    if fldCtx.depth > 0 {
        fldCtx.depth--
        return ev, nil
    }
    cr.stack.Pop()
    p := ev.GetPath()
    if err := processDefaults( fldCtx, p, next ); err != nil { return nil, err }
    fldCtx.removeOptFields()
    if fldCtx.await.Len() > 0 { 
        return nil, createMissingFieldsError( p, fldCtx ) 
    }
    return ev, nil
}

func ( cr *castReactor ) startContainer() error {
    if ! cr.stack.IsEmpty() { cr.stack.Peek().( *fieldCtx ).depth++ }
    return nil
}

// we only do value checks here that are specific to this cast, namely having to
// do with enums. If the value is an enum, we check that we recogzize the type
// and that the type is actually an enum. We don't actually check the enum value
// here though, and leave that for CastAtomic. Any other values aren't checked
// here and are left to CastAtomic or to the upstream processor.
func ( cr *castReactor ) valueEvent( 
    ve *mgRct.ValueEvent ) ( mgRct.ReactorEvent, error ) {

    if en, ok := ve.Val.( *mg.Enum ); ok {
        if def, ok := cr.dm.GetOk( en.Type ); ok {
            if _, ok := def.( *EnumDefinition ); ok { return ve, nil }
            enTyp := en.Type.AsAtomicType()
            return nil, notAnEnumTypeError( enTyp, ve.GetPath() )
        } 
        return nil, newUnrecognizedTypeError( ve.GetPath(), en.Type )
    }
    return ve, nil
}

func ( cr *castReactor ) prepareProcessEvent(
    ev mgRct.ReactorEvent, 
    next mgRct.ReactorEventProcessor ) ( mgRct.ReactorEvent, error ) {
    
    switch v := ev.( type ) {
    case *mgRct.StructStartEvent: return cr.startStruct( v )
    case *mgRct.FieldStartEvent: return cr.startField( v )
    case *mgRct.ValueEvent: return cr.valueEvent( v )
    case *mgRct.EndEvent: return cr.end( v, next )
    case *mgRct.ListStartEvent, *mgRct.MapStartEvent: 
        return ev, cr.startContainer()
    case *mgRct.ValueAllocationEvent, *mgRct.ValueReferenceEvent: return ev, nil
    }
    panic( libErrorf( "unhandled event: %T", ev ) )
}

func ( cr *castReactor ) ProcessEvent( 
    ev mgRct.ReactorEvent, next mgRct.ReactorEventProcessor ) error {

    ev, err := cr.prepareProcessEvent( ev, next )
    if err != nil { return err }
    return next.ProcessEvent( ev )
}

// the public version of newCastReactorDefinitionMap, typed to return something
// other than our internal *castReactor type; we could combine this with
// newCastReactorDefinitionMap if we end up making *castReactor public
func NewCastReactorDefinitionMap(
    typ mg.TypeReference, dm *DefinitionMap ) mgRct.PipelineProcessor {
 
    return &castReactor{ typ: typ, dm: dm, stack: stack.NewStack() }
}
