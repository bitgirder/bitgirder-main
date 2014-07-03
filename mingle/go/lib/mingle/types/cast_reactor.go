package types

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "bitgirder/objpath"
    "bitgirder/pipeline"
    "bitgirder/stack"
    "log"
    "fmt"
    "bytes"
)

func asMapStartEvent( ev mgRct.ReactorEvent ) *mgRct.MapStartEvent {
    res := mgRct.NewMapStartEvent() 
    res.SetPath( ev.GetPath() )
    return res
}

func notAFieldSetTypeError( 
    p objpath.PathNode, qn *mg.QualifiedTypeName ) error {

    return mg.NewValueCastErrorf( p, "not a type with fields: %s", qn )
}

func newUnrecognizedTypeError(
    p objpath.PathNode, qn *mg.QualifiedTypeName ) error {

    return mg.NewValueCastErrorf( p, "unrecognized type: %s", qn )
}

func notAnEnumTypeError( typ mg.TypeReference, path objpath.PathNode ) error {
    return mg.NewValueCastErrorf( path, "not an enum type: %s", typ )
}

func fieldSetForTypeInDefMap(
    qn *mg.QualifiedTypeName, 
    dm *DefinitionMap, 
    path objpath.PathNode ) ( *FieldSet, error ) {

    if def, ok := dm.GetOk( qn ); ok {
        switch v := def.( type ) {
        case *StructDefinition: return v.Fields, nil
        case *SchemaDefinition: return v.Fields, nil
        default: return nil, notAFieldSetTypeError( path, qn )
        } 
    } 
    return nil, newUnrecognizedTypeError( path, qn )
}

type fieldTyper interface {

    // path will be positioned to the map/struct containing fld, but will not
    // itself include fld
    fieldTypeFor( 
        fld *mg.Identifier, path objpath.PathNode ) ( mg.TypeReference, error )
}

type valueFieldTyper int

func ( vt valueFieldTyper ) fieldTypeFor( 
    fld *mg.Identifier, path objpath.PathNode ) ( mg.TypeReference, error ) {
    return mg.TypeNullableValue, nil
}

type CastReactor struct {

    dm *DefinitionMap

    stack *stack.Stack

    SkipPathSetter bool
}

func ( cr *CastReactor ) dumpStack( pref string ) {
    bb := &bytes.Buffer{}
    fmt.Fprintf( bb, "%s: [", pref )
    cr.stack.VisitTop( func( v interface{} ) {
        msg := fmt.Sprintf( "%T", v )
        switch v2 := v.( type ) {
        case mg.TypeReference: msg = v2.ExternalForm()
        }
        fmt.Fprintf( bb, msg )
        fmt.Fprintf( bb, ", " )
    })
    fmt.Fprintf( bb, " ]" )
    log.Print( bb.String() )
}

func NewCastReactor( expct mg.TypeReference, dm *DefinitionMap ) *CastReactor {
    res := &CastReactor{ stack: stack.NewStack(), dm: dm }
    res.stack.Push( expct )
    return res
}

func ( cr *CastReactor ) InitializePipeline( pip *pipeline.Pipeline ) {
    mgRct.EnsureStructuralReactor( pip )
    if ! cr.SkipPathSetter { mgRct.EnsurePathSettingProcessor( pip ) }
}

type fieldCast struct {
    ft fieldTyper
    await *mg.IdentifierMap
}

func ( fc *fieldCast ) removeOptFields() {
    done := make( []*mg.Identifier, 0, fc.await.Len() )
    fc.await.EachPair( func( _ *mg.Identifier, val interface{} ) {
        fd := val.( *FieldDefinition )
        if _, ok := fd.Type.( *mg.NullableTypeReference ); ok {
            done = append( done, fd.Name )
        }
    })
    for _, fld := range done { fc.await.Delete( fld ) }
}

func feedDefault( 
    fld *mg.Identifier, 
    defl mg.Value, 
    p objpath.PathNode,
    next mgRct.ReactorEventProcessor ) error {

    fldPath := objpath.Descend( p, fld )
    fs := mgRct.NewFieldStartEvent( fld )
    fs.SetPath( fldPath )
    if err := next.ProcessEvent( fs ); err != nil { return err }
    ps := mgRct.NewPathSettingProcessorPath( fldPath )
    ps.SkipStructureCheck = true
    pip := mgRct.InitReactorPipeline( ps, next )
    return mgRct.VisitValue( defl, pip )
}

func processDefaults(
    fc *fieldCast,
    p objpath.PathNode, 
    next mgRct.ReactorEventProcessor ) error {

    vis := func( fld *mg.Identifier, val interface{} ) error {
        fd := val.( *FieldDefinition )
        if defl := fd.GetDefault(); defl != nil { 
            if err := feedDefault( fld, defl, p, next ); err != nil { 
                return err 
            }
            fc.await.Delete( fld )
        }
        return nil
    }
    return fc.await.EachPairError( vis )
}

func createMissingFieldsError( p objpath.PathNode, fc *fieldCast ) error {
    flds := make( []*mg.Identifier, 0, fc.await.Len() )
    fc.await.EachPair( func( fld *mg.Identifier, _ interface{} ) {
        flds = append( flds, fld )
    })
    return mg.NewMissingFieldsError( p, flds )
}

type listCast struct {
    sawValues bool
    lt *mg.ListTypeReference
    startPath objpath.PathNode
}

func ( cr *CastReactor ) errStackUnrecognized() error {
    return libErrorf( "unrecognized stack element: %T", cr.stack.Peek() )
}

func ( cr *CastReactor ) castAtomic(
    v mg.Value,
    at *mg.AtomicTypeReference,
    path objpath.PathNode ) ( mg.Value, error, bool ) {

    if def, ok := cr.dm.GetOk( at.Name ); ok {
        if ed, ok := def.( *EnumDefinition ); ok {
            res, err := castEnum( v, ed, path )
            return res, err, true
        } 
        if ev, ok := v.( *mg.Enum ); ok {
            if ! ev.Type.Equals( at.Name ) { return nil, nil, false }
            if _, ok := def.( *StructDefinition ); ok {
                return nil, notAnEnumTypeError( at, path ), true
            }
        }
    }
    return nil, nil, false
}

func ( cr *CastReactor ) valueEventForAtomicCast( 
    ve *mgRct.ValueEvent, 
    at *mg.AtomicTypeReference, 
    callTyp mg.TypeReference ) ( error, *mgRct.ValueEvent ) {

    mv, err, ok := cr.castAtomic( ve.Val, at, ve.GetPath() )
    if ! ok { 
        mv, err = castAtomicWithCallType( ve.Val, at, callTyp, ve.GetPath() ) 
    }
    if err != nil { return err, nil }
    res := mgRct.CopyEvent( ve, true ).( *mgRct.ValueEvent )
    res.Val = mv
    return nil, res
}

func ( cr *CastReactor ) processAtomicValue(
    ve *mgRct.ValueEvent,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    err, ve2 := cr.valueEventForAtomicCast( ve, at, callTyp )
    if err != nil { return err }
    if err = next.ProcessEvent( ve2 ); err != nil { return err }
    return nil
}

func nullValueEventForType( 
    ve *mgRct.ValueEvent, typ mg.TypeReference ) *mgRct.ValueEvent {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference: return ve
    case *mg.PointerTypeReference: return nullValueEventForType( ve, v.Type )
    case *mg.NullableTypeReference: return nullValueEventForType( ve, v.Type )
    case *mg.ListTypeReference: return ve
    }
    panic( libErrorf( "unhandled type reference: %T", typ ) )
}

func ( cr *CastReactor ) processNullableValue(
    ve *mgRct.ValueEvent,
    nt *mg.NullableTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    if _, ok := ve.Val.( *mg.Null ); ok { 
        return next.ProcessEvent( nullValueEventForType( ve, nt ) ) 
    }
    return cr.processValueWithType( ve, nt.Type, callTyp, next )
}

func ( cr *CastReactor ) processValueForListType(
    ve *mgRct.ValueEvent,
    typ *mg.ListTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    if _, ok := ve.Val.( *mg.Null ); ok {
        return newNullValueCastError( ve.GetPath() )
    }
    return mg.NewTypeCastErrorValue( callTyp, ve.Val, ve.GetPath() )
}

func ( cr *CastReactor ) processValueWithType(
    ve *mgRct.ValueEvent,
    typ mg.TypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference: 
        return cr.processAtomicValue( ve, v, callTyp, next )
    case *mg.PointerTypeReference:
        return cr.processValueWithType( ve, v.Type, callTyp, next )
    case *mg.NullableTypeReference:
        return cr.processNullableValue( ve, v, callTyp, next )
    case *mg.ListTypeReference:
        return cr.processValueForListType( ve, v, callTyp, next )
    }
    panic( libErrorf( "unhandled type: %T", typ ) )
}

func ( cr *CastReactor ) processValue( 
    ve *mgRct.ValueEvent, next mgRct.ReactorEventProcessor ) error {

    switch v := cr.stack.Peek().( type ) {
    case mg.TypeReference: 
        cr.stack.Pop()
        return cr.processValueWithType( ve, v, v, next )
    case *listCast:
        v.sawValues = true
        typ := v.lt.ElementType
        return cr.processValueWithType( ve, typ, typ, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) implMapStart(
    ev mgRct.ReactorEvent, 
    ft fieldTyper, 
    fs *FieldSet,
    next mgRct.ReactorEventProcessor ) error {

    fc := &fieldCast{ ft: ft }
    if fs != nil {
        fc.await = mg.NewIdentifierMap()
        fs.EachDefinition( func( fd *FieldDefinition ) {
            fc.await.Put( fd.Name, fd )
        })
    }
    cr.stack.Push( fc )
    return next.ProcessEvent( ev )
}

type fieldSetTyper struct { 
    flds *FieldSet 
    dm *DefinitionMap
    ignoreUnrecognized bool
}

func ( ft *fieldSetTyper ) fieldTypeFor(
    fld *mg.Identifier, path objpath.PathNode ) ( mg.TypeReference, error ) {

    if fd := ft.flds.Get( fld ); fd != nil { return fd.Type, nil }
    if ft.ignoreUnrecognized { return mg.TypeValue, nil }
    return nil, mg.NewUnrecognizedFieldError( path, fld )
}

func ( cr *CastReactor ) fieldSetTyperForStruct(
    def *StructDefinition, path objpath.PathNode ) ( *fieldSetTyper, error ) {

    return &fieldSetTyper{ flds: def.Fields, dm: cr.dm }, nil
}

func ( cr *CastReactor ) fieldSetTyperForSchema( 
    sd *SchemaDefinition ) *fieldSetTyper {

    return &fieldSetTyper{ 
        flds: sd.Fields, 
        dm: cr.dm, 
        ignoreUnrecognized: true,
    }
}

func ( cr *CastReactor ) fieldSetTyperFor(
    qn *mg.QualifiedTypeName, 
    path objpath.PathNode ) ( *fieldSetTyper, error ) {

    if def, ok := cr.dm.GetOk( qn ); ok {
        switch v := def.( type ) {
        case *StructDefinition: return cr.fieldSetTyperForStruct( v, path )
        case *SchemaDefinition: return cr.fieldSetTyperForSchema( v ), nil
        default: return nil, notAFieldSetTypeError( path, qn )
        }
    }
    tmpl := "no field type info for type %s"
    return nil, mg.NewValueCastErrorf( path, tmpl, qn )
}

func ( cr *CastReactor ) completeStartStruct(
    ss *mgRct.StructStartEvent, next mgRct.ReactorEventProcessor ) error {

    ft, err := cr.fieldSetTyperFor( ss.Type, ss.GetPath() )
    if err != nil { return err }
    var ev mgRct.ReactorEvent = ss
    fs, err := fieldSetForTypeInDefMap( ss.Type, cr.dm, ss.GetPath() )
    if err != nil { return err }
    if def, ok := cr.dm.GetOk( ss.Type ); ok {
        if _, ok := def.( *SchemaDefinition ); ok { ev = asMapStartEvent( ss ) }
    } 
    return cr.implMapStart( ev, ft, fs, next )
}

func ( cr *CastReactor ) inferStructForQname( qn *mg.QualifiedTypeName ) bool {
    if def, ok := cr.dm.GetOk( qn ); ok {
        if _, ok = def.( *StructDefinition ); ok { return true }
        if _, ok = def.( *SchemaDefinition ); ok { return true }
    }
    return false
}

func ( cr *CastReactor ) inferStructForMap(
    me *mgRct.MapStartEvent,
    at *mg.AtomicTypeReference,
    next mgRct.ReactorEventProcessor ) ( error, bool ) {

    if ! cr.inferStructForQname( at.Name ) { return nil, false }

    ev := mgRct.NewStructStartEvent( at.Name )
    ev.SetPath( me.GetPath() )

    return cr.completeStartStruct( ev, next ), true
}

func ( cr *CastReactor ) processMapStartWithAtomicType(
    me *mgRct.MapStartEvent,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    if at.Equals( mg.TypeSymbolMap ) || at.Equals( mg.TypeValue ) {
        return cr.implMapStart( me, valueFieldTyper( 1 ), nil, next )
    }

    if err, ok := cr.inferStructForMap( me, at, next ); ok { return err }

    return mg.NewTypeCastError( callTyp, mg.TypeSymbolMap, me.GetPath() )
}

func ( cr *CastReactor ) processMapStartWithType(
    me *mgRct.MapStartEvent, 
    typ mg.TypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference:
        return cr.processMapStartWithAtomicType( me, v, callTyp, next )
    case *mg.PointerTypeReference:
        return cr.processMapStartWithType( me, v.Type, callTyp, next )
    case *mg.NullableTypeReference:
        return cr.processMapStartWithType( me, v.Type, callTyp, next )
    }
    return mg.NewTypeCastError( callTyp, typ, me.GetPath() )
}

func ( cr *CastReactor ) processMapStart(
    me *mgRct.MapStartEvent, next mgRct.ReactorEventProcessor ) error {
    
    switch v := cr.stack.Peek().( type ) {
    case mg.TypeReference: 
        cr.stack.Pop()
        return cr.processMapStartWithType( me, v, v, next )
    case *listCast:
        v.sawValues = true
        typ := v.lt.ElementType
        return cr.processMapStartWithType( me, typ, typ, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) processFieldStart(
    fs *mgRct.FieldStartEvent, next mgRct.ReactorEventProcessor ) error {

    fc := cr.stack.Peek().( *fieldCast )
    if fc.await != nil { fc.await.Delete( fs.Field ) }
    
    typ, err := fc.ft.fieldTypeFor( fs.Field, fs.GetPath().Parent() )
    if err != nil { return err }

    cr.stack.Push( typ )
    return next.ProcessEvent( fs )
}

func ( cr *CastReactor ) processListEnd() error {
    lc := cr.stack.Pop().( *listCast )
    if ! ( lc.sawValues || lc.lt.AllowsEmpty ) {
        return mg.NewValueCastError( lc.startPath, "empty list" )
    }
    return nil
}

func ( cr *CastReactor ) processFieldsEnd( 
    ee *mgRct.EndEvent, next mgRct.ReactorEventProcessor ) error {

    fc := cr.stack.Pop().( *fieldCast )
    if fc.await == nil { return nil }
    p := ee.GetPath()
    if err := processDefaults( fc, p, next ); err != nil { return err }
    fc.removeOptFields()
    if fc.await.Len() > 0 { return createMissingFieldsError( p, fc ) }
    return nil
}

func ( cr *CastReactor ) processEnd(
    ee *mgRct.EndEvent, next mgRct.ReactorEventProcessor ) error {

    switch cr.stack.Peek().( type ) {
    case *listCast: if err := cr.processListEnd(); err != nil { return err }
    case *fieldCast: 
        if err := cr.processFieldsEnd( ee, next ); err != nil { return err }
    }

    if err := next.ProcessEvent( ee ); err != nil { return err }
    return nil
}

func ( cr *CastReactor ) allowAssignment( 
    expct, act *mg.QualifiedTypeName ) bool {

    if _, ok := cr.dm.GetOk( act ); ! ok { return false }
    return canAssignType( expct, act, cr.dm )
}

func ( cr *CastReactor ) processStructStartWithAtomicType(
    ss *mgRct.StructStartEvent,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    if at.Equals( mg.TypeSymbolMap ) {
        me := asMapStartEvent( ss )
        return cr.processMapStartWithAtomicType( me, at, callTyp, next )
    }

    if at.Name.Equals( ss.Type ) || at.Equals( mg.TypeValue ) ||
       cr.allowAssignment( at.Name, ss.Type ) {
        return cr.completeStartStruct( ss, next )
    }

    failTyp := &mg.AtomicTypeReference{ Name: ss.Type }
    return mg.NewTypeCastError( callTyp, failTyp, ss.GetPath() )
}

func ( cr *CastReactor ) processStructStartWithType(
    ss *mgRct.StructStartEvent,
    typ mg.TypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference:
        return cr.processStructStartWithAtomicType( ss, v, callTyp, next )
    case *mg.PointerTypeReference:
        return cr.processStructStartWithType( ss, v.Type, callTyp, next )
    case *mg.NullableTypeReference:
        return cr.processStructStartWithType( ss, v.Type, callTyp, next )
    }
    return mg.NewTypeCastError( typ, callTyp, ss.GetPath() )
}

func ( cr *CastReactor ) processStructStart(
    ss *mgRct.StructStartEvent, next mgRct.ReactorEventProcessor ) error {

    switch v := cr.stack.Peek().( type ) {
    case mg.TypeReference:
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
    le *mgRct.ListStartEvent,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    if at.Equals( mg.TypeValue ) {
        return cr.processListStartWithType( 
            le, mg.TypeOpaqueList, callTyp, next )
    }

    return mg.NewTypeCastError( callTyp, mg.TypeOpaqueList, le.GetPath() )
}

func ( cr *CastReactor ) processListStartWithListType(
    le *mgRct.ListStartEvent,
    lt *mg.ListTypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {
    
    sp := objpath.CopyOf( le.GetPath() )
    cr.stack.Push( &listCast{ lt: lt, startPath: sp } )
    return next.ProcessEvent( le )
}

func ( cr *CastReactor ) processListStartWithType(
    le *mgRct.ListStartEvent,
    typ mg.TypeReference,
    callTyp mg.TypeReference,
    next mgRct.ReactorEventProcessor ) error {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference:
        return cr.processListStartWithAtomicType( le, v, callTyp, next )
    case *mg.PointerTypeReference:
        return cr.processListStartWithType( le, v.Type, callTyp, next )
    case *mg.ListTypeReference:
        return cr.processListStartWithListType( le, v, callTyp, next )
    case *mg.NullableTypeReference:
        return cr.processListStartWithType( le, v.Type, callTyp, next )
    }
    panic( libErrorf( "unhandled type: %T", typ ) )
}

func ( cr *CastReactor ) processListStart( 
    le *mgRct.ListStartEvent, next mgRct.ReactorEventProcessor ) error {

    switch v := cr.stack.Peek().( type ) {
    case mg.TypeReference:
        cr.stack.Pop()
        return cr.processListStartWithType( le, v, v, next )
    case *listCast:
        v.sawValues = true
        return cr.processListStartWithType( le, v.lt.ElementType, v.lt, next )
    }
    panic( cr.errStackUnrecognized() )
}

func ( cr *CastReactor ) ProcessEvent(
    ev mgRct.ReactorEvent, next mgRct.ReactorEventProcessor ) ( err error ) {

//    cr.dumpStack( "entering ProcessEvent()" )
//    defer cr.dumpStack( "after ProcessEvent()" )
    switch v := ev.( type ) {
    case *mgRct.ValueEvent: return cr.processValue( v, next )
    case *mgRct.MapStartEvent: return cr.processMapStart( v, next )
    case *mgRct.FieldStartEvent: return cr.processFieldStart( v, next )
    case *mgRct.StructStartEvent: return cr.processStructStart( v, next )
    case *mgRct.ListStartEvent: return cr.processListStart( v, next )
    case *mgRct.EndEvent: return cr.processEnd( v, next )
    }
    panic( libErrorf( "unhandled event: %T", ev ) )
}
