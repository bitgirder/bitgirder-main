package types

import (
    "reflect"
    "fmt"
    "sort"
    "mingle/parser"
//    "log"
    "bitgirder/assert"
    mg "mingle"
)

var (
    mkQn = parser.MustQualifiedTypeName
    mkId = parser.MustIdentifier
    mkNs = parser.MustNamespace
    mkTyp = parser.MustTypeReference
    asType = parser.AsTypeReference
)

func idSetFor( m *mg.IdentifierMap ) []*mg.Identifier {
    res := make( []*mg.Identifier, 0, m.Len() )
    m.EachPair( func( id *mg.Identifier, _ interface{} ) {
        res = append( res, id )
    })
    return res
}

func MakeFieldDef( 
    nm string, typ interface{}, defl interface{} ) *FieldDefinition {

    res := &FieldDefinition{ Name: mkId( nm ), Type: asType( typ ) }
    if defl != nil { 
        if val, err := mg.AsValue( defl ); err == nil {
            res.Default = val
        } else { panic( err ) }
    }
    return res
}

func buildFieldSet( fs *FieldSet, flds []*FieldDefinition ) {
    for _, fld := range flds { fs.MustAdd( fld ) }
}

func MakeFieldSet( flds ...*FieldDefinition ) *FieldSet {
    res := NewFieldSet()
    buildFieldSet( res, flds )
    return res
}

func MakeStructDef( qn string, flds []*FieldDefinition ) *StructDefinition {
    if flds == nil { flds = []*FieldDefinition{} }
    res := NewStructDefinition()
    res.Name = parser.MustQualifiedTypeName( qn )
    buildFieldSet( res.Fields, flds )
    return res
}

func MakeSchemaDef( qn string, flds []*FieldDefinition ) *SchemaDefinition {
    res := NewSchemaDefinition()
    res.Name = mkQn( qn )
    buildFieldSet( res.Fields, flds )
    return res
}

func MakeEnumDef( qn string, vals ...string ) *EnumDefinition {
    res := &EnumDefinition{
        Name: parser.MustQualifiedTypeName( qn ),
        Values: make( []*mg.Identifier, len( vals ) ),
    }
    for i, val := range vals { res.Values[ i ] = parser.MustIdentifier( val ) }
    return res
}

func MakeCallSig( 
    flds []*FieldDefinition, retType string, throws []string ) *CallSignature {

    res := NewCallSignature()
    for _, fld := range flds { res.Fields.MustAdd( fld ) }
    res.Return = mkTyp( retType )
    throwsTyps := make( []mg.TypeReference, len( throws ) )
    for i, typ := range throws { throwsTyps[ i ] = mkTyp( typ ) }
    if len( throwsTyps ) > 0 {
        res.Throws = MustUnionTypeDefinitionTypes( throwsTyps... )
    }
    return res
}

func MakeOpDef( nm string, sig *CallSignature ) *OperationDefinition {
    return &OperationDefinition{ Name: mkId( nm ), Signature: sig }
}

func MakeServiceDef(
    qn, secQn string, opDefs ...*OperationDefinition ) *ServiceDefinition {
    res := NewServiceDefinition()
    res.Name = parser.MustQualifiedTypeName( qn )
    res.Operations = append( res.Operations, opDefs... )
    if secQn != "" { res.Security = parser.MustQualifiedTypeName( secQn ) }
    return res
}

func mustAddDefs( dm *DefinitionMap, defs []Definition ) *DefinitionMap {
    for _, d := range defs { dm.MustAdd( d ) }
    return dm
}

func MakeDefMap( defs ...Definition ) *DefinitionMap {
    return mustAddDefs( NewDefinitionMap(), defs )
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

func ( a *DefAsserter ) equalTypeRef( t1, t2 mg.TypeReference ) {
    a.Truef( t1.Equals( t2 ), "%s != %s", t1, t2 )
}

func ( a *DefAsserter ) assertPrimDef( 
    p1 *PrimitiveDefinition, d2 Definition ) {
    _ = a.equalType( p1, d2 ).( *PrimitiveDefinition )
}

func ( a *DefAsserter ) assertAliasDef( 
    a1 *AliasedTypeDefinition, d2 Definition ) {

    a2 := a.equalType( a1, d2 ).( *AliasedTypeDefinition )
    a.Descend( "Name" ).Equal( a1.Name, a2.Name )
    a.Descend( "AliasedType" ).Equal( a1.AliasedType, a2.AliasedType )
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

func ( a *DefAsserter ) assertFieldDef( fd1, fd2 *FieldDefinition ) {
    a.descend( "(Name)" ).Equal( fd1.Name, fd2.Name )
    a.descend( "(Type)" ).equalTypeRef( fd1.Type, fd2.Type )
    mg.AssertEqualValues( fd1.Default, fd2.Default, a.descend( "(Default)" ) )
}

// First check that both have same field sets, then check field by field
func ( a *DefAsserter ) assertFieldSets( fs1, fs2 *FieldSet ) {
    a.assertIdSets( fs1.GetFieldNames(), fs2.GetFieldNames() )
    fs1.EachDefinition( func( fd1 *FieldDefinition ) {
        fd2 := fs2.Get( fd1.Name )
        a.descend( fd1.Name ).assertFieldDef( fd1, fd2 )
    })
}

func ( a *DefAsserter ) assertUnionType( ut1, ut2 *UnionTypeDefinition ) {
    if ut1 == nil && ut2 == nil { return }
    if ut1 == nil { a.Truef( ut2 == nil, "ut2 not nil" ) }
    if ut2 == nil { a.Truef( ut1 == nil, "ut1 not nil" ) }
    typs1, typs2 := ut1.Types, ut2.Types
    a.descend( "(Len)" ).Equal( len( typs1 ), len( typs2 ) )
    la := a.startList()
    for i, e := 0, len( typs1 ); i < e; i++ {
        la.descend( "(Type)" ).Equal( typs1[ i ], typs2[ i ] )
        la = la.next()
    }
}

func ( a *DefAsserter ) assertStructDef(
    s1 *StructDefinition, d2 Definition ) {
    s2 := a.equalType( s1, d2 ).( *StructDefinition )
    a.descend( "(Fields)" ).assertFieldSets( s1.Fields, s2.Fields )
    a.descend( "(Constructors)" ).
        assertUnionType( s1.Constructors, s2.Constructors )
}

func ( a *DefAsserter ) assertSchemaDef( s1 *SchemaDefinition, d2 Definition ) {
    s2 := a.equalType( s1, d2 ).( *SchemaDefinition )
    a.descend( "(Fields)" ).assertFieldSets( s1.Fields, s2.Fields )
}

func ( a *DefAsserter ) assertEnumDef( 
    e1 *EnumDefinition, v2 interface{} ) {
    e2 := a.equalType( e1, v2 ).( *EnumDefinition )
    a.descend( "(Values)" ).assertIdSets( e1.Values, e2.Values )
}

func ( a *DefAsserter ) assertCallSig( s1, s2 *CallSignature ) {
    a.descend( "(Fields)" ).assertFieldSets( s1.Fields, s2.Fields )
    a.descend( "(Return)" ).Equal( s1.Return, s2.Return )
    a.descend( "(Throws)" ).assertUnionType( s1.Throws, s2.Throws )
}

func ( a *DefAsserter ) assertProtoDef(
    p1 *PrototypeDefinition, v2 interface{} ) {
    p2 := a.equalType( p1, v2 ).( *PrototypeDefinition )
    a.descend( "Signature" ).assertCallSig( p1.Signature, p2.Signature )
}

func ( a *DefAsserter ) assertOpDef( od1, od2 *OperationDefinition ) {
    a.descend( "(Name)" ).Equal( od1.Name, od2.Name )
    a.descend( "(Signature" ).assertCallSig( od1.Signature, od2.Signature )
}

func ( a *DefAsserter ) assertOpDefs( 
    defs1, defs2 []*OperationDefinition ) {
    m1, m2 := OpDefsByName( defs1 ), OpDefsByName( defs2 )
    a.descend( "(Len)" ).Equal( m1.Len(), m2.Len() )
    a.descend( "(OpNames)" ).assertIdSets( idSetFor( m1 ), idSetFor( m2 ) )
    m1.EachPair( func( id *mg.Identifier, val interface{} ) {
        opDef1 := val.( *OperationDefinition )
        opDef2, _ := m2.Get( id ).( *OperationDefinition )
        a.descend( id.ExternalForm() ).assertOpDef( opDef1, opDef2 )
    })
}

func ( a *DefAsserter ) assertServiceDef(
    s1 *ServiceDefinition, v2 interface{} ) {

    s2 := a.equalType( s1, v2 ).( *ServiceDefinition )
    a.descend( "(Operations)" ).assertOpDefs( s1.Operations, s2.Operations )
    a.descend( "(Security)" ).Equal( s1.Security, s2.Security )
}

func ( a *DefAsserter ) AssertDef( d1, d2 Definition ) {
    a.descend( "(Name)" ).Equal( d1.GetName(), d2.GetName() )
    a = a.descend( d1.GetName() )
    switch v := d1.( type ) {
    case *PrimitiveDefinition: a.assertPrimDef( v, d2 )
    case *AliasedTypeDefinition: a.assertAliasDef( v, d2 )
    case *StructDefinition: a.assertStructDef( v, d2 )
    case *SchemaDefinition: a.assertSchemaDef( v, d2 )
    case *EnumDefinition: a.assertEnumDef( v, d2 )
    case *PrototypeDefinition: a.assertProtoDef( v, d2 )
    case *ServiceDefinition: a.assertServiceDef( v, d2 )
    default: a.Fatalf( "Unhandled def: %T", d1 )
    }
}

func ( a *DefAsserter ) AssertDefMaps( m1, m2 *DefinitionMap ) {
    a.descend( "(Len)" ).Equal( m1.Len(), m2.Len() )
    m1.EachDefinition( func( d1 Definition ) {
        nm := d1.GetName()
        da := a.descend( nm.ExternalForm() )
        if d2 := m2.Get( nm ); d2 == nil {
            da.Fatal( "No corresponding entry in m2" )
        } else { da.AssertDef( d1, d2 ) }
    })
}
