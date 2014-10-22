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

func asMapStartEvent( ev mgRct.Event ) *mgRct.MapStartEvent {
    res := mgRct.NewMapStartEvent() 
    res.SetPath( ev.GetPath() )
    return res
}

func notAFieldSetTypeError( 
    p objpath.PathNode, qn *mg.QualifiedTypeName ) error {

    return mg.NewCastErrorf( p, "not a type with fields: %s", qn )
}

func fieldSetForTypeInDefMap(
    qn *mg.QualifiedTypeName, 
    dm DefinitionGetter, 
    path objpath.PathNode ) ( *FieldSet, error ) {

    if def, ok := dm.GetDefinition( qn ); ok {
        switch v := def.( type ) {
        case *StructDefinition: return v.Fields, nil
        case *SchemaDefinition: return v.Fields, nil
        default: return nil, notAFieldSetTypeError( path, qn )
        } 
    } 
    return nil, mg.NewCastErrorf( path, "unrecognized type: %s", qn )
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

type SymbolMapFieldSetGetter interface {
    GetFieldSet( path objpath.PathNode ) ( *FieldSet, error )
}

type CastReactor struct {

    dm DefinitionGetter

    stack *stack.Stack

    passFieldsByQn *mg.QnameMap

    passthroughTracker *mgRct.DepthTracker

    unionMatchFuncs *mg.QnameMap

    FieldSetFactory SymbolMapFieldSetGetter

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

func ( cr *CastReactor ) pushType( typ mg.TypeReference ) {
    cr.stack.Push( typ )
}

func NewCastReactor( 
    expct mg.TypeReference, dm DefinitionGetter ) *CastReactor {

    res := &CastReactor{ 
        stack: stack.NewStack(), 
        dm: dm,
        passFieldsByQn: mg.NewQnameMap(),
        unionMatchFuncs: mg.NewQnameMap(),
    }
    res.pushType( expct )
    return res
}

func ( cr *CastReactor ) passFieldsForQn( 
    qn *mg.QualifiedTypeName ) *mg.IdentifierMap {

    if v, ok := cr.passFieldsByQn.GetOk( qn ); ok {
        return v.( *mg.IdentifierMap )
    }
    return nil
}

func ( cr *CastReactor ) AddPassthroughField( 
    qn *mg.QualifiedTypeName, fld *mg.Identifier ) {

    pf := cr.passFieldsForQn( qn )
    if pf == nil {
        pf = mg.NewIdentifierMap()
        cr.passFieldsByQn.Put( qn, pf )
    }
    pf.Put( fld, true )
}

func ( cr *CastReactor ) SetUnionDefinitionMatcher(
    qn *mg.QualifiedTypeName, mf UnionMatchFunction ) {

    cr.unionMatchFuncs.Put( qn, mf )
}

func ( cr *CastReactor ) InitializePipeline( pip *pipeline.Pipeline ) {
    mgRct.EnsureStructuralReactor( pip )
    if ! cr.SkipPathSetter { mgRct.EnsurePathSettingProcessor( pip ) }
}

func ( cr *CastReactor ) processPassthrough(
    ev mgRct.Event, next mgRct.EventProcessor ) error {

    if err := next.ProcessEvent( ev ); err != nil { return err }
    if err := cr.passthroughTracker.ProcessEvent( ev ); err != nil { 
        return err 
    }
    if cr.passthroughTracker.Depth() == 0 { cr.passthroughTracker = nil }
    return nil
}

type fieldCast struct {
    ft fieldTyper
    await *mg.IdentifierMap
    passFields *mg.IdentifierMap
}

func ( fc *fieldCast ) isPassthroughField( fld *mg.Identifier ) bool {
    if m := fc.passFields; m != nil {
        if _, pass := m.GetOk( fld ); pass { return true }
    }
    return false
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
    next mgRct.EventProcessor ) error {

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
    next mgRct.EventProcessor ) error {

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

func ( cr *CastReactor ) unionTypeDefForAtomicType( 
    at *mg.AtomicTypeReference ) *UnionTypeDefinition {

    if td, ok := cr.dm.GetDefinition( at.Name() ); ok {
        if utd, ok := td.( *UnionDefinition ); ok { return utd.Union }
    }
    return nil
}

func ( cr *CastReactor ) getStructDef( 
    nm *mg.QualifiedTypeName ) *StructDefinition {

    if def, ok := cr.dm.GetDefinition( nm ); ok {
        if sd, ok := def.( *StructDefinition ); ok { return sd }
    }
    return nil
}

// Only handles positive mismatches in which a declared type carries a type that
// corresponds to a known definition, but for which that definition makes no
// sense. We let the case when no such definition exists at all be handled
// elsewhere.
func ( cr *CastReactor ) checkWellFormed(
    ev mgRct.Event,
    typ *mg.QualifiedTypeName,
    errDesc string,
    defCheck func( def Definition ) bool ) error {

    if def, ok := cr.dm.GetDefinition( typ ); ok {
        if defCheck( def ) { return nil }
        tmpl := "not %s type: %s"
        return mg.NewCastErrorf( ev.GetPath(), tmpl, errDesc, typ )
    }
    return nil
}

func ( cr *CastReactor ) checkValueWellFormed( ve *mgRct.ValueEvent ) error {
    en, ok := ve.Val.( *mg.Enum )
    if ! ok { return nil }
    chk := func( def Definition ) bool {
        _, ok := def.( *EnumDefinition ); 
        return ok
    }
    return cr.checkWellFormed( ve, en.Type, "an enum", chk )
}

func ( cr *CastReactor ) constructorTypeForType(
    typ mg.TypeReference, sd *StructDefinition ) mg.TypeReference {

    if sd.Constructors == nil { return nil }
    res, ok := sd.Constructors.MatchType( typ )
    if ok { return res }
    return nil
}

func ( cr *CastReactor ) castStructConstructor(
    v mg.Value,
    sd *StructDefinition,
    path objpath.PathNode ) ( mg.Value, error, bool ) {

    if cr.constructorTypeForType( mg.TypeOf( v ), sd ) != nil { 
        return v, nil, true 
    }
    return nil, nil, false
}

func ( cr *CastReactor ) castAtomic(
    v mg.Value,
    at *mg.AtomicTypeReference,
    path objpath.PathNode ) ( mg.Value, error, bool ) {

    if def, ok := cr.dm.GetDefinition( at.Name() ); ok {
        switch td := def.( type ) {
        case *EnumDefinition:
            res, err := castEnum( v, td, path )
            return res, err, true
        case *StructDefinition: return cr.castStructConstructor( v, td, path )
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
    next mgRct.EventProcessor ) error {

    err, ve2 := cr.valueEventForAtomicCast( ve, at, callTyp )
    if err != nil { return err }
    if err = next.ProcessEvent( ve2 ); err != nil { return err }
    return nil
}

func ( cr *CastReactor ) processNullableValue(
    ve *mgRct.ValueEvent,
    nt *mg.NullableTypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {

    if _, ok := ve.Val.( *mg.Null ); ok { return next.ProcessEvent( ve ) }
    return cr.processValueWithType( ve, nt.Type, callTyp, next )
}

func ( cr *CastReactor ) processValueForListType(
    ve *mgRct.ValueEvent,
    typ *mg.ListTypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {

    if _, ok := ve.Val.( *mg.Null ); ok {
        return newNullCastError( ve.GetPath() )
    }
    return mg.NewTypeCastErrorValue( callTyp, ve.Val, ve.GetPath() )
}

func ( cr *CastReactor ) matchUnionDefType(
    ev mgRct.Event, ud *UnionDefinition ) ( mg.TypeReference, bool ) {

    typ := mgRct.TypeOfEvent( ev )
    ut := ud.Union
    if mf, ok := cr.unionMatchFuncs.GetOk( ud.Name ); ok {
        return mf.( UnionMatchFunction )( UnionMatchInput{ typ, ut, cr.dm } )
    }
    return ut.MatchType( typ )
}

func ( cr *CastReactor ) getUnionApplication(
    ev mgRct.Event,
    at *mg.AtomicTypeReference,
    next mgRct.EventProcessor ) func() error {

    if def, ok := cr.dm.GetDefinition( at.Name() ); ok {
        if ud, ok := def.( *UnionDefinition ); ok {
            if mtch, ok := cr.matchUnionDefType( ev, ud ); ok {
                return func() error { 
                    cr.pushType( mtch )
                    return cr.ProcessEvent( ev, next ) 
                }
            }
        }
    }
    return nil
}

func ( cr *CastReactor ) processValueWithType(
    ve *mgRct.ValueEvent,
    typ mg.TypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference: 
        if f := cr.getUnionApplication( ve, v, next ); f != nil { return f() }
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
    ve *mgRct.ValueEvent, next mgRct.EventProcessor ) error {

    if err := cr.checkValueWellFormed( ve ); err != nil { return err }
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
    ev mgRct.Event, 
    ft fieldTyper, 
    fs *FieldSet,
    passFields *mg.IdentifierMap,
    next mgRct.EventProcessor ) error {

    fc := &fieldCast{ ft: ft, passFields: passFields }
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
    dm DefinitionGetter
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

    if def, ok := cr.dm.GetDefinition( qn ); ok {
        switch v := def.( type ) {
        case *StructDefinition: return cr.fieldSetTyperForStruct( v, path )
        case *SchemaDefinition: return cr.fieldSetTyperForSchema( v ), nil
        default: return nil, notAFieldSetTypeError( path, qn )
        }
    }
    tmpl := "no field type info for type %s"
    return nil, mg.NewCastErrorf( path, tmpl, qn )
}

func ( cr *CastReactor ) completeStartStruct(
    ss *mgRct.StructStartEvent, next mgRct.EventProcessor ) error {

    ft, err := cr.fieldSetTyperFor( ss.Type, ss.GetPath() )
    if err != nil { return err }
    var ev mgRct.Event = ss
    fs, err := fieldSetForTypeInDefMap( ss.Type, cr.dm, ss.GetPath() )
    if err != nil { return err }
    if def, ok := cr.dm.GetDefinition( ss.Type ); ok {
        if _, ok := def.( *SchemaDefinition ); ok { ev = asMapStartEvent( ss ) }
    } 
    pf := cr.passFieldsForQn( ss.Type )
    return cr.implMapStart( ev, ft, fs, pf, next )
}

func ( cr *CastReactor ) inferStructForQname( qn *mg.QualifiedTypeName ) bool {
    if def, ok := cr.dm.GetDefinition( qn ); ok {
        if _, ok = def.( *StructDefinition ); ok { return true }
        if _, ok = def.( *SchemaDefinition ); ok { return true }
    }
    return false
}

func ( cr *CastReactor ) inferStructForMap(
    me *mgRct.MapStartEvent,
    at *mg.AtomicTypeReference,
    next mgRct.EventProcessor ) ( error, bool ) {

    if ! cr.inferStructForQname( at.Name() ) { return nil, false }

    ev := mgRct.NewStructStartEvent( at.Name() )
    ev.SetPath( me.GetPath() )

    return cr.completeStartStruct( ev, next ), true
}

func ( cr *CastReactor ) processMapStartWithAtomicType(
    me *mgRct.MapStartEvent,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {

    if at.Equals( mg.TypeSymbolMap ) || at.Equals( mg.TypeValue ) {
        var ft fieldTyper = valueFieldTyper( 1 )
        var fs *FieldSet
        if cr.FieldSetFactory != nil {
            var err error
            fs, err = cr.FieldSetFactory.GetFieldSet( me.GetPath() )
            if err != nil { return err }
            if fs != nil { ft = &fieldSetTyper{ flds: fs, dm: cr.dm } }
        }
        return cr.implMapStart( me, ft, fs, nil, next )
    }

    if err, ok := cr.inferStructForMap( me, at, next ); ok { return err }

    return mg.NewTypeCastError( callTyp, mg.TypeSymbolMap, me.GetPath() )
}

func ( cr *CastReactor ) processMapStartWithType(
    me *mgRct.MapStartEvent, 
    typ mg.TypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference:
        if f := cr.getUnionApplication( me, v, next ); f != nil { return f() }
        return cr.processMapStartWithAtomicType( me, v, callTyp, next )
    case *mg.PointerTypeReference:
        return cr.processMapStartWithType( me, v.Type, callTyp, next )
    case *mg.NullableTypeReference:
        return cr.processMapStartWithType( me, v.Type, callTyp, next )
    }
    return mg.NewTypeCastError( callTyp, typ, me.GetPath() )
}

func ( cr *CastReactor ) processMapStart(
    me *mgRct.MapStartEvent, next mgRct.EventProcessor ) error {
    
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
    fs *mgRct.FieldStartEvent, next mgRct.EventProcessor ) error {

    fc := cr.stack.Peek().( *fieldCast )
    if fc.await != nil { fc.await.Delete( fs.Field ) }
    
    if fc.isPassthroughField( fs.Field ) {
        cr.passthroughTracker = mgRct.NewDepthTracker()
    } else {
        typ, err := fc.ft.fieldTypeFor( fs.Field, fs.GetPath().Parent() )
        if err != nil { return err }
        cr.pushType( typ )
    }

    return next.ProcessEvent( fs )
}

func ( cr *CastReactor ) processListEnd() error {
    lc := cr.stack.Pop().( *listCast )
    if ! ( lc.sawValues || lc.lt.AllowsEmpty ) {
        return mg.NewCastError( lc.startPath, "empty list" )
    }
    return nil
}

func ( cr *CastReactor ) processFieldsEnd( 
    ee *mgRct.EndEvent, next mgRct.EventProcessor ) error {

    fc := cr.stack.Pop().( *fieldCast )
    if fc.await == nil { return nil }
    p := ee.GetPath()
    if err := processDefaults( fc, p, next ); err != nil { return err }
    fc.removeOptFields()
    if fc.await.Len() > 0 { return createMissingFieldsError( p, fc ) }
    return nil
}

func ( cr *CastReactor ) processEnd(
    ee *mgRct.EndEvent, next mgRct.EventProcessor ) error {

    switch cr.stack.Peek().( type ) {
    case *listCast: if err := cr.processListEnd(); err != nil { return err }
    case *fieldCast: 
        if err := cr.processFieldsEnd( ee, next ); err != nil { return err }
    }

    if err := next.ProcessEvent( ee ); err != nil { return err }
    return nil
}

func ( cr *CastReactor ) checkStructWellFormed(
    ss *mgRct.StructStartEvent ) error {

    return cr.checkWellFormed( ss, ss.Type, "a struct",
        func ( def Definition ) bool {
            _, ok := def.( *StructDefinition )
            return ok
        },
    )
}

func ( cr *CastReactor ) allowStructStartForType( 
    ss *mgRct.StructStartEvent, expct *mg.QualifiedTypeName ) bool {

    if _, ok := cr.dm.GetDefinition( ss.Type ); ! ok { return false }
    if sd := cr.getStructDef( expct ); sd != nil {
        if cr.constructorTypeForType( ss.Type.AsAtomicType(), sd ) != nil {
            return true
        }
    }
    return canAssignType( expct, ss.Type, cr.dm )
}

func ( cr *CastReactor ) processStructStartWithAtomicType(
    ss *mgRct.StructStartEvent,
    at *mg.AtomicTypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {

    if at.Equals( mg.TypeSymbolMap ) {
        me := asMapStartEvent( ss )
        return cr.processMapStartWithAtomicType( me, at, callTyp, next )
    }

    if at.Name().Equals( ss.Type ) || at.Equals( mg.TypeValue ) ||
       cr.allowStructStartForType( ss, at.Name() ) {
        return cr.completeStartStruct( ss, next )
    }

    failTyp := mg.NewAtomicTypeReference( ss.Type, nil )
    return mg.NewTypeCastError( callTyp, failTyp, ss.GetPath() )
}

func ( cr *CastReactor ) processStructStartWithType(
    ss *mgRct.StructStartEvent,
    typ mg.TypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference:
        if f := cr.getUnionApplication( ss, v, next ); f != nil { return f() }
        return cr.processStructStartWithAtomicType( ss, v, callTyp, next )
    case *mg.PointerTypeReference:
        return cr.processStructStartWithType( ss, v.Type, callTyp, next )
    case *mg.NullableTypeReference:
        return cr.processStructStartWithType( ss, v.Type, callTyp, next )
    }
    return mg.NewTypeCastError( typ, callTyp, ss.GetPath() )
}

func ( cr *CastReactor ) processStructStart(
    ss *mgRct.StructStartEvent, next mgRct.EventProcessor ) error {

    if err := cr.checkStructWellFormed( ss ); err != nil { return err }
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
    next mgRct.EventProcessor ) error {

    if at.Equals( mg.TypeValue ) {
        return cr.processListStartWithType( 
            le, mg.TypeOpaqueList, callTyp, next )
    }
    if f := cr.getUnionApplication( le, at, next ); f != nil { return f() }
    if sd := cr.getStructDef( at.Name() ); sd != nil {
        if typ := cr.constructorTypeForType( le.Type, sd ); typ != nil {
            lt := typ.( *mg.ListTypeReference )
            return cr.processListStartWithListType( le, lt, callTyp, next )
        }
    }
    return mg.NewTypeCastError( callTyp, le.Type, le.GetPath() )
}

func ( cr *CastReactor ) processListStartWithListType(
    le *mgRct.ListStartEvent,
    lt *mg.ListTypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {
    
    sp := objpath.CopyOf( le.GetPath() )
    cr.stack.Push( &listCast{ lt: lt, startPath: sp } )
    return next.ProcessEvent( le )
}

func ( cr *CastReactor ) processListStartWithType(
    le *mgRct.ListStartEvent,
    typ mg.TypeReference,
    callTyp mg.TypeReference,
    next mgRct.EventProcessor ) error {

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
    le *mgRct.ListStartEvent, next mgRct.EventProcessor ) error {

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
    ev mgRct.Event, next mgRct.EventProcessor ) ( err error ) {

    if cr.passthroughTracker != nil { return cr.processPassthrough( ev, next ) }
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
