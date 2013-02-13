package types

import (
    mg "mingle"
    "bitgirder/objpath"
    "container/list"
//    "log"
)

type castIface struct { 
    dm *DefinitionMap 
}

func ( ci castIface ) InferStructFor( qn *mg.QualifiedTypeName ) bool {
    if def, ok := ci.dm.GetOk( qn ); ok {
        _, res := def.( *StructDefinition )
        return res
    }
    return false
}

func collectFieldSets( sd *StructDefinition, dm *DefinitionMap ) []*FieldSet {
    flds := make( []*FieldSet, 0, 2 )
    for {
        flds = append( flds, sd.Fields )
        spr := sd.GetSuperType()
        if spr == nil { break }
        if def, ok := dm.GetOk( spr ); ok {
            if sd, ok = def.( *StructDefinition ); ! ok {
                tmpl := "Super type %s of %s is not a struct"
                panic( libErrorf( tmpl, spr, sd.GetName() ) )
            }
        } else {
            tmpl := "Can't find super type %s of %s"
            panic( libErrorf( tmpl, spr, sd.GetName() ) )
        }
    }
    return flds
}

func typeNameIn( fd *FieldDefinition ) *mg.QualifiedTypeName {
    nm := mg.TypeNameIn( fd.Type )
    if qn, ok := nm.( *mg.QualifiedTypeName ); ok { return qn }
    panic( libErrorf( 
        "Name in type %s is not a qname: %s (%T)", fd.Type, nm, nm ) )
}

func expectDef( dm *DefinitionMap, qn *mg.QualifiedTypeName ) Definition {
    if def, ok := dm.GetOk( qn ); ok { return def }
    panic( libErrorf( "map has no definition for type %s", qn ) )
}

func valDefOf( fd *FieldDefinition, dm *DefinitionMap ) Definition {
    qn := typeNameIn( fd )
    return expectDef( dm, qn )
}

type fieldTyper struct { 
    flds []*FieldSet 
    dm *DefinitionMap
}

//func asOpaqueType( t mg.TypeReference ) mg.TypeReference {
//    switch t2 := t.( type ) {
//    case *mg.AtomicTypeReference: return mg.TypeValue
//    case *mg.NullableTypeReference: 
//        return &mg.NullableTypeReference{ asOpaqueType( t2.Type ) }
//    case *mg.ListTypeReference:
//        return &mg.ListTypeReference{ 
//            ElementType: asOpaqueType( t2.ElementType ),
//            AllowsEmpty: t2.AllowsEmpty,
//        }
//    }
//    panic( libErrorf( "Unhandled type reference: %T", t ) )
//}
//
//func ( ft fieldTyper ) effectiveFieldTypeOf( 
//    fd *FieldDefinition ) mg.TypeReference {
//    valDef := valDefOf( fd, ft.dm )
//    switch valDef.( type ) {
//    case *EnumDefinition: return asOpaqueType( fd.Type )
//    }
//    return fd.Type
//}

func ( ft fieldTyper ) FieldTypeOf(
    fld *mg.Identifier, pg mg.PathGetter ) ( mg.TypeReference, error ) {
    for _, flds := range ft.flds {
        if fd := flds.Get( fld ); fd != nil { return fd.Type, nil }
    }
    // use parent path since we're positioned on the failed field itself
    par := objpath.ParentOf( pg.GetPath() )
    return nil, mg.NewUnrecognizedFieldError( par, fld )
}

func ( ci castIface ) FieldTyperFor(
    qn *mg.QualifiedTypeName, pg mg.PathGetter ) ( mg.FieldTyper, error ) {
    flds := make( []*FieldSet, 0, 2 )
    for nm := qn; nm != nil; {
        if def, ok := ci.dm.GetOk( nm ); ok {
            if sd, ok := def.( *StructDefinition ); ok {
                flds = append( flds, sd.Fields )
                nm = sd.GetSuperType()
                continue
            } else { return nil, notAStructError( nm, pg.GetPath() ) } 
        }
        nm = nil
    }
    if len( flds ) > 0 { return fieldTyper{ flds: flds, dm: ci.dm }, nil }
    tmpl := "No field type info for type %s"
    return nil, mg.NewValueCastErrorf( pg.GetPath(), tmpl, qn )
}

func completeCastEnum(
    id *mg.Identifier, 
    ed *EnumDefinition, 
    pg mg.PathGetter ) ( *mg.Enum, error ) {
    if res := ed.GetValue( id ); res != nil { return res, nil }
    tmpl := "illegal value for enum %s: %s"
    return nil, mg.NewValueCastErrorf( pg.GetPath(), tmpl, ed.GetName(), id )
}

func castEnumFromString( 
    s string, ed *EnumDefinition, pg mg.PathGetter ) ( *mg.Enum, error ) {
    id, err := mg.ParseIdentifier( s )
    if err != nil {
        p := pg.GetPath()
        tmpl := "invalid enum value %q: %s"
        return nil, mg.NewValueCastErrorf( p, tmpl, s, err )
    }
    return completeCastEnum( id, ed, pg )
}

func castEnum( 
    val mg.Value, ed *EnumDefinition, pg mg.PathGetter ) ( *mg.Enum, error ) {
    switch v := val.( type ) {
    case mg.String: return castEnumFromString( string( v ), ed, pg )
    case *mg.Enum: 
        if v.Type.Equals( ed.GetName() ) {
            return completeCastEnum( v.Value, ed, pg )
        }
    }
    t := ed.GetName().AsAtomicType()
    return nil, mg.NewTypeCastErrorValue( t, val, pg.GetPath() )
}

func ( ci castIface ) CastAtomic(
    v mg.Value,
    at *mg.AtomicTypeReference,
    pg mg.PathGetter ) ( mg.Value, error, bool ) {
    if qn, ok := at.Name.( *mg.QualifiedTypeName ); ok {
        if def, ok := ci.dm.GetOk( qn ); ok {
            if ed, ok := def.( *EnumDefinition ); ok {
                res, err := castEnum( v, ed, pg )
                return res, err, true
            }
        }
    }
    return nil, nil, false
}

type castReactor struct {
    castBase *mg.CastReactor
    dm *DefinitionMap
    stack *list.List
    deflFeed *mg.EventPathReactor
}

func ( cr *castReactor ) Init( rpi *mg.ReactorPipelineInit ) {
    rpi.AddPipelineProcessor( cr.castBase )
}

func ( cr *castReactor ) GetPath() objpath.PathNode {
    res := cr.castBase.GetPath()
    if cr.deflFeed != nil { res = cr.deflFeed.AppendPath( res ) }
    return res
}

func ( cr *castReactor ) newValueCastErrorf( 
    tmpl string, args ...interface{} ) error {
    return mg.NewValueCastErrorf( cr.GetPath(), tmpl, args... )
}

func ( cr *castReactor ) newUnrecognizedTypeError( 
    qn *mg.QualifiedTypeName ) error {
    return cr.newValueCastErrorf( "Unrecognized type: %s", qn )
}

func notAStructError( qn *mg.QualifiedTypeName, p objpath.PathNode ) error {
    return mg.NewValueCastErrorf( p, "Not a struct type: %s", qn )
}

func ( cr *castReactor ) notAStructError( qn *mg.QualifiedTypeName ) error {
    return notAStructError( qn, cr.GetPath() )
}

type fieldCtx struct {
    endCount int
    await *mg.IdentifierMap
}

func ( fc *fieldCtx ) removeOptFields() {
    done := make( []*mg.Identifier, 0, fc.await.Len() )
    fc.await.EachPair( func( _ *mg.Identifier, val interface{} ) {
        fd := val.( *FieldDefinition )
        if _, ok := fd.Type.( *mg.NullableTypeReference ); ok {
            done = append( done, fd.Name )
        }
    })
    for _, fld := range done { fc.await.Delete( fld ) }
}

func ( cr *castReactor ) newFieldCtx( sd *StructDefinition ) *fieldCtx {
    res := &fieldCtx{ await: mg.NewIdentifierMap() }
    for _, fs := range collectFieldSets( sd, cr.dm ) {
        fs.EachDefinition( func( fd *FieldDefinition ) {
            res.await.Put( fd.Name, fd )
        })
    }
    return res
}

func ( cr *castReactor ) peek() *fieldCtx {
    if cr.stack.Len() == 0 { panic( libError( "fieldCtx stack empty" ) ) }
    return cr.stack.Front().Value.( *fieldCtx )
}

func ( cr *castReactor ) startStruct( typ *mg.QualifiedTypeName ) error {
    if def, ok := cr.dm.GetOk( typ ); ok {
        if sd, ok := def.( *StructDefinition ); ok {
            cr.stack.PushFront( cr.newFieldCtx( sd ) )
        } else { return cr.notAStructError( typ ) }
    } else { return cr.newUnrecognizedTypeError( typ ) }
    return nil
}

// We don't re-check here that fld is actually part of the defined field set or
// that this is the first time seeing it, since the upstream structural reactor
// and castIface will have validated that already
func ( cr *castReactor ) startField( fld *mg.Identifier ) {
    if cr.stack.Len() == 0 { return }
    fldCtx := cr.peek()
    fldCtx.await.Delete( fld )
}

func ( cr *castReactor ) feedDefault( 
    fld *mg.Identifier, defl mg.Value, rep mg.ReactorEventProcessor ) error {
    cr.deflFeed = mg.NewEventPathReactor( rep )
    defer func() { cr.deflFeed = nil }()
    fs := mg.FieldStartEvent{ fld }
    if err := cr.deflFeed.ProcessEvent( fs ); err != nil { return err }
    return mg.VisitValue( defl, cr.deflFeed )
}

func ( cr *castReactor ) processDefaults(
    fldCtx *fieldCtx, ee mg.EndEvent, rep mg.ReactorEventProcessor ) error {
    vis := func( fld *mg.Identifier, val interface{} ) error {
        fd := val.( *FieldDefinition )
        if defl := fd.GetDefault(); defl != nil { 
            if err := cr.feedDefault( fld, defl, rep ); err != nil { 
                return err 
            }
            fldCtx.await.Delete( fld )
        }
        return nil
    }
    return fldCtx.await.EachPairError( vis )
}

func ( cr *castReactor ) createMissingFieldsError( fldCtx *fieldCtx ) error {
    flds := make( []*mg.Identifier, 0, fldCtx.await.Len() )
    fldCtx.await.EachPair( func( fld *mg.Identifier, _ interface{} ) {
        flds = append( flds, fld )
    })
    return mg.NewMissingFieldsError( cr.GetPath(), flds )
//    mg.SortIds( flds )
//    strs := make( []string, len( flds ) )
//    for i, fld := range flds { strs[ i ] = fld.ExternalForm() }
//    fldsStr := strings.Join( strs, ", " )
//    return cr.newValueCastErrorf( "missing field(s): %s", fldsStr )
}

func ( cr *castReactor ) end( 
    ev mg.EndEvent, rep mg.ReactorEventProcessor ) error {
    if cr.stack.Len() == 0 { return nil }
    fldCtx := cr.peek()
    if fldCtx.endCount == 0 {
        if err := cr.processDefaults( fldCtx, ev, rep ); err != nil {
            return err
        }
        fldCtx.removeOptFields()
        if fldCtx.await.Len() > 0 { 
            return cr.createMissingFieldsError( fldCtx )
        } else { return nil }
    }
    fldCtx.endCount--
    return nil
}

func ( cr *castReactor ) startContainer() {
    if cr.stack.Len() == 0 { return }
    cr.peek().endCount++
}

func ( cr *castReactor ) checkValue( v mg.Value ) error {
    if en, ok := v.( *mg.Enum ); ok {
        if def, ok := cr.dm.GetOk( en.Type ); ok {
            if _, ok := def.( *EnumDefinition ); ok {
                return nil // Later we'll check the value too
            } else {
                tmpl := "Not an enum type: %s"
                return cr.newValueCastErrorf( tmpl, en.Type )
            }
        } else { return cr.newUnrecognizedTypeError( en.Type ) }
    }
    return nil
}

func ( cr *castReactor ) ProcessEvent( 
    ev mg.ReactorEvent, rep mg.ReactorEventProcessor ) error {
    switch v := ev.( type ) {
    case mg.StructStartEvent: 
        if err := cr.startStruct( v.Type ); err != nil { return err }
    case mg.EndEvent: if err := cr.end( v, rep ); err != nil { return err }
    case mg.FieldStartEvent: cr.startField( v.Field )
    case mg.MapStartEvent, mg.ListStartEvent: cr.startContainer()
    case mg.ValueEvent: 
        if err := cr.checkValue( v.Val ); err != nil { return err }
    }
    return rep.ProcessEvent( ev )
}

func NewCastReactor( 
    typ mg.TypeReference, dm *DefinitionMap ) mg.PipelineProcessor {
    castBase := mg.NewCastReactor( typ, castIface{ dm }, nil )
    return &castReactor{ 
        castBase: castBase, 
        dm: dm,
        stack: &list.List{},
    }
}
