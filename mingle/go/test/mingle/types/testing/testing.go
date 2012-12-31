package testing

import (
    "reflect"
    "fmt"
    "sort"
//    "log"
    "bitgirder/assert"
    mg "mingle"
    "mingle/types"
)

var (
    mkQn = mg.MustQualifiedTypeName
    mkId = mg.MustIdentifier
    mkTyp = mg.MustTypeReference
)

func idSetFor( m *mg.IdentifierMap ) []*mg.Identifier {
    res := make( []*mg.Identifier, 0, m.Len() )
    m.EachPair( func( id *mg.Identifier, _ interface{} ) {
        res = append( res, id )
    })
    return res
}

func MakeFieldDef( nm, typ string, defl interface{} ) *types.FieldDefinition {
    res := &types.FieldDefinition{
        Name: mg.MustIdentifier( nm ),
        Type: mg.MustTypeReference( typ ),
    }
    if defl != nil { 
        if val, err := mg.AsValue( defl ); err == nil {
            res.Default = val
        } else { panic( err ) }
    }
    return res
}

func MakeStructDef( 
    qn, sprTyp string, flds []*types.FieldDefinition ) *types.StructDefinition {
    if flds == nil { flds = []*types.FieldDefinition{} }
    res := types.NewStructDefinition()
    res.Name = mg.MustQualifiedTypeName( qn )
    if sprTyp != "" { res.SuperType = mg.MustTypeReference( sprTyp ) }
    for _, fld := range flds { res.Fields.MustAdd( fld ) }
    return res
}

func MakeStructDef2(
    qn, sprTyp string,
    flds []*types.FieldDefinition,
    cons []*types.ConstructorDefinition ) *types.StructDefinition {
    res := MakeStructDef( qn, sprTyp, flds )
    res.Constructors = append( res.Constructors, cons... )
    return res
}

func MakeEnumDef( qn string, vals ...string ) *types.EnumDefinition {
    res := &types.EnumDefinition{
        Name: mg.MustQualifiedTypeName( qn ),
        Values: make( []*mg.Identifier, len( vals ) ),
    }
    for i, val := range vals { res.Values[ i ] = mg.MustIdentifier( val ) }
    return res
}

func MakeCallSig( 
    flds []*types.FieldDefinition,
    retType string,
    throws []string ) *types.CallSignature {
    res := types.NewCallSignature()
    for _, fld := range flds { res.Fields.MustAdd( fld ) }
    res.Return = mg.MustTypeReference( retType )
    for _, typ := range throws { 
        res.Throws = append( res.Throws, mg.MustTypeReference( typ ) )
    }
    return res
}

func MakeServiceDef(
    qn, sprTyp string,
    opDefs []*types.OperationDefinition,
    secQn string ) *types.ServiceDefinition {
    res := types.NewServiceDefinition()
    res.Name = mg.MustQualifiedTypeName( qn )
    if sprTyp != "" { res.SuperType = mg.MustTypeReference( sprTyp ) }
    res.Operations = append( res.Operations, opDefs... )
    if secQn != "" { res.Security = mg.MustQualifiedTypeName( secQn ) }
    return res
}

func MakeDefMap( defs ...types.Definition ) *types.DefinitionMap {
    res := types.NewDefinitionMap()
    for _, d := range defs { res.MustAdd( d ) }
    return res
}

type DefAsserter struct {
    *assert.PathAsserter
}

func NewDefAsserter( f assert.Failer ) *DefAsserter {
    return &DefAsserter{ assert.NewPathAsserter( f ) }
}

func ( a *DefAsserter ) descend( node interface{} ) *DefAsserter {
    return &DefAsserter{ a.PathAsserter.Descend( node ) }
}

func ( a *DefAsserter ) startList() *DefAsserter {
    return &DefAsserter{ a.PathAsserter.StartList() }
}

func ( a *DefAsserter ) next() *DefAsserter {
    return &DefAsserter{ a.PathAsserter.Next() }
}

func ( a *DefAsserter ) equalType( v1, v2 interface{} ) interface{} {
    t1, t2 := reflect.TypeOf( v1 ), reflect.TypeOf( v2 )
    if t1 != t2 { a.Fatalf( "Expected %T but got %T", v1, v2 ) }
    return v2
}

func ( a *DefAsserter ) assertPrimDef( 
    p1 *types.PrimitiveDefinition, d2 types.Definition ) {
    _ = a.equalType( p1, d2 ).( *types.PrimitiveDefinition )
}

func ( a *DefAsserter ) assertAliasDef( 
    a1 *types.AliasedTypeDefinition, d2 types.Definition ) {
    a2 := a.equalType( a1, d2 ).( *types.AliasedTypeDefinition )
    a.Equal( a1, a2 )
}

func asCompStr( ids []*mg.Identifier ) string {
    strs := make( []string, len( ids ) )
    for i, id := range ids { strs[ i ] = id.ExternalForm() }
    sort.Strings( strs )
    return fmt.Sprintf( "%v", strs )
}

func ( a *DefAsserter ) assertIdSets( ids1, ids2 []*mg.Identifier ) {
    if cs1, cs2 := asCompStr( ids1 ), asCompStr( ids2 ); cs1 != cs2 {
        a.Fatalf( "Id sets differ: %s != %s", cs1, cs2 )
    }
}

func ( a *DefAsserter ) assertFieldDef( fd1, fd2 *types.FieldDefinition ) {
    a.descend( "(Name)" ).Equal( fd1.Name, fd2.Name )
    a.descend( "(Type)" ).True( fd1.Type.Equals( fd2.Type ) )
    a.descend( "(Default)" ).Equal( fd1.Default, fd2.Default )
}

// First check that both have same field sets, then check field by field
func ( a *DefAsserter ) assertFieldSets( fs1, fs2 *types.FieldSet ) {
    a.assertIdSets( fs1.GetFieldNames(), fs2.GetFieldNames() )
    fs1.EachDefinition( func( fd1 *types.FieldDefinition ) {
        fd2 := fs2.Get( fd1.Name )
        a.descend( fd1.Name ).assertFieldDef( fd1, fd2 )
    })
}

func ( a *DefAsserter ) assertConstructors( 
    defs1, defs2 []*types.ConstructorDefinition ) {
    a.descend( "(Len)" ).Equal( len( defs1 ), len( defs2 ) )
    la := a.startList()
    for i, e := 0, len( defs1 ); i < e; i++ {
        cons1, cons2 := defs1[ i ], defs2[ i ]
        la.descend( "(Type)" ).Equal( cons1.Type, cons2.Type )
        la = la.next()
    }
}

func ( a *DefAsserter ) assertStructDef(
    s1 *types.StructDefinition, d2 types.Definition ) {
    s2 := a.equalType( s1, d2 ).( *types.StructDefinition )
    a.descend( "(SuperType)" ).Equal( s1.SuperType, s2.SuperType )
    a.descend( "(Fields)" ).assertFieldSets( s1.Fields, s2.Fields )
    a.descend( "(Constructors)" ).
        assertConstructors( s1.Constructors, s2.Constructors )
}

func ( a *DefAsserter ) assertEnumDef( 
    e1 *types.EnumDefinition, v2 interface{} ) {
    e2 := a.equalType( e1, v2 ).( *types.EnumDefinition )
    a.descend( "(Values)" ).assertIdSets( e1.Values, e2.Values )
}

func ( a *DefAsserter ) assertCallSig( s1, s2 *types.CallSignature ) {
    a.descend( "(Fields)" ).assertFieldSets( s1.Fields, s2.Fields )
    a.descend( "(Return)" ).Equal( s1.Return, s2.Return )
    throws1, throws2 := s1.Throws, s2.Throws
    ta := a.descend( "(Throws)" )
    ta.descend( "(Len)" ).Equal( len( throws1 ), len( throws2 ) )
    for la, i, e := ta.startList(), 0, len( throws1 ); i < e; i++ {
        la.Equal( throws1[ i ], throws2[ i ] )
        la = la.next()
    }
}

func ( a *DefAsserter ) assertProtoDef(
    p1 *types.PrototypeDefinition, v2 interface{} ) {
    p2 := a.equalType( p1, v2 ).( *types.PrototypeDefinition )
    a.descend( "Signature" ).assertCallSig( p1.Signature, p2.Signature )
}

func ( a *DefAsserter ) assertOpDef( od1, od2 *types.OperationDefinition ) {
    a.descend( "(Name)" ).Equal( od1.Name, od2.Name )
    a.descend( "(Signature" ).assertCallSig( od1.Signature, od2.Signature )
}

func ( a *DefAsserter ) assertOpDefs( 
    defs1, defs2 []*types.OperationDefinition ) {
    m1, m2 := types.OpDefsByName( defs1 ), types.OpDefsByName( defs2 )
    a.descend( "(Len)" ).Equal( m1.Len(), m2.Len() )
    a.descend( "(OpNames)" ).assertIdSets( idSetFor( m1 ), idSetFor( m2 ) )
    m1.EachPair( func( id *mg.Identifier, val interface{} ) {
        opDef1 := val.( *types.OperationDefinition )
        opDef2, _ := m2.Get( id ).( *types.OperationDefinition )
        a.descend( id.ExternalForm() ).assertOpDef( opDef1, opDef2 )
    })
}

func ( a *DefAsserter ) assertServiceDef(
    s1 *types.ServiceDefinition, v2 interface{} ) {
    s2 := a.equalType( s1, v2 ).( *types.ServiceDefinition )
    a.descend( "(SuperType)" ).Equal( s1.SuperType, s2.SuperType )
    a.descend( "(Operations)" ).assertOpDefs( s1.Operations, s2.Operations )
    a.descend( "(Security)" ).Equal( s1.Security, s2.Security )
}

func ( a *DefAsserter ) AssertDef( d1, d2 types.Definition ) {
    a.descend( "(Name)" ).Equal( d1.GetName(), d2.GetName() )
    switch v := d1.( type ) {
    case *types.PrimitiveDefinition: a.assertPrimDef( v, d2 )
    case *types.AliasedTypeDefinition: a.assertAliasDef( v, d2 )
    case *types.StructDefinition: a.assertStructDef( v, d2 )
    case *types.EnumDefinition: a.assertEnumDef( v, d2 )
    case *types.PrototypeDefinition: a.assertProtoDef( v, d2 )
    case *types.ServiceDefinition: a.assertServiceDef( v, d2 )
    default: a.Fatalf( "Unhandled def: %T", d1 )
    }
}

func ( a *DefAsserter ) AssertDefMaps( m1, m2 *types.DefinitionMap ) {
    a.descend( "(Len)" ).Equal( m1.Len(), m2.Len() )
    m1.EachDefinition( func( d1 types.Definition ) {
        nm := d1.GetName()
        da := a.descend( nm.ExternalForm() )
        if d2 := m2.Get( nm ); d2 == nil {
            da.Fatal( "No corresponding entry in m2" )
        } else { da.AssertDef( d1, d2 ) }
    })
}