package types

import (
    "fmt"
//    "log"
    "sort"
    mg "mingle"
)

func idUnsafe( parts ...string ) *mg.Identifier {
    return mg.NewIdentifierUnsafe( parts )
}

type DefinitionGetter interface {
    GetDefinition( qn *mg.QualifiedTypeName ) ( Definition, bool )
}

func MustGetDefinition( 
    qn *mg.QualifiedTypeName, dg DefinitionGetter ) Definition {

    if res, ok := dg.GetDefinition( qn ); ok { return res }
    panic( libErrorf( "no definition found for name: %s", qn ) )
}

type DefinitionMap struct {
    m *mg.QnameMap
    builtIn *mg.QnameMap
}

func NewDefinitionMap() *DefinitionMap {
    return &DefinitionMap{ m: mg.NewQnameMap(), builtIn: mg.NewQnameMap() }
}

func ( dm *DefinitionMap ) setBuiltIn( qn *mg.QualifiedTypeName ) {
    dm.builtIn.Put( qn, true )
}

func ( m *DefinitionMap ) Len() int { return m.m.Len() }

func ( m *DefinitionMap ) GetDefinition( 
    qn *mg.QualifiedTypeName ) ( Definition, bool ) {
    d, ok := m.m.GetOk( qn  )
    if ok { return d.( Definition ), true }
    return nil, false
}

func ( m *DefinitionMap ) Get( qn *mg.QualifiedTypeName ) Definition {
    if d, ok := m.GetDefinition( qn ); ok { return d }
    return nil
}

func ( m *DefinitionMap ) HasKey( qn *mg.QualifiedTypeName ) bool {
    return m.m.HasKey( qn )
}

func ( m *DefinitionMap ) HasBuiltInDefinition( 
    qn *mg.QualifiedTypeName ) bool {
    return m.builtIn.HasKey( qn )
}

func ( m *DefinitionMap ) Add( d Definition ) error {
    return m.m.PutSafe( d.GetName(), d )
}

// sole impl for now; we may add a version Add() that returns error on dupes, in
// which case this method will call into that one
func ( m *DefinitionMap ) MustAdd( d Definition ) {
    if err := m.Add( d ); err != nil { panic( err ) }
}

func ( m *DefinitionMap ) MustAddAll( defs ...Definition ) {
    for _, def := range defs { m.MustAdd( def ) }
}

func ( m *DefinitionMap ) MustAddFrom( m2 *DefinitionMap ) {
    m2.EachDefinition( func( d Definition ) { m.MustAdd( d ) } )
}

func ( m *DefinitionMap ) EachDefinition( f func( d Definition ) ) {
    m.m.EachPair(
        func( qn *mg.QualifiedTypeName, d interface{} ) {
            f( d.( Definition ) )
        },
    )
}

type Definition interface {
    GetName() *mg.QualifiedTypeName
}

type PrimitiveDefinition struct { Name *mg.QualifiedTypeName }

func ( pd *PrimitiveDefinition ) GetName() *mg.QualifiedTypeName {
    return pd.Name
}

type UnionTypeDefinitionInput interface {
    Len() int
    TypeAtIndex( idx int ) mg.TypeReference
}

type UnionTypeDefinitionError struct {
    ErrorGroups [][]int
}

func ( e *UnionTypeDefinitionError ) Error() string {
    return "union contains one or more ambiguous types"
}

// Should not be created directly -- use CreateUnionTypeDefinition()
type UnionTypeDefinition struct {
    Types []mg.TypeReference
}

func UnionTypeKeyForType( typ mg.TypeReference ) string {
    switch v := typ.( type ) {
    case *mg.AtomicTypeReference: return v.Name().ExternalForm()
    case *mg.NullableTypeReference: return UnionTypeKeyForType( v.Type )
    case *mg.PointerTypeReference: return UnionTypeKeyForType( v.Type )
    case *mg.ListTypeReference: 
        return UnionTypeKeyForType( v.ElementType ) + "[]"
    }
    panic( libErrorf( "unhandled type: %T", typ ) )
}

// sorts union type error groups by increasing order of each group's first
// element (groups themselves are assumed to already be sorted in increasing
// order)
type errorGroupSort [][]int

func ( s errorGroupSort ) Len() int { return len( s ) }

func ( s errorGroupSort ) Less( i, j int ) bool {
    return s[ i ][ 0 ] < s[ j ][ 0 ]
}

func ( s errorGroupSort ) Swap( i, j int ) { s[ i ], s[ j ] = s[ j ], s[ i ] }

// strips pointers, nullability, and differentiation between lists that allow
// empty and those that do not, and organizes types into groups that are thereby
// equal, or groups those that are not together by index
func checkUnionType( 
    in UnionTypeDefinitionInput ) ( []mg.TypeReference, [][]int ) {
    
    m := make( map[ string ] []int )
    typs := make( []mg.TypeReference, in.Len() )
    for i, e := 0, in.Len(); i < e; i++ {
        typ := in.TypeAtIndex( i )
        typs[ i ] = typ
        key := UnionTypeKeyForType( typ )
        matched, ok := m[ key ]
        if ! ok { matched = make( []int, 0, 4 ) }
        matched = append( matched, i )
        m[ key ] = matched
    }
    errGrps := make( [][]int, 0, len( m ) )
    for _, v := range m { 
        if len( v ) > 1 { errGrps = append( errGrps, v ) }
    }
    sort.Sort( errorGroupSort( errGrps ) )
    return typs, errGrps
}

func CreateUnionTypeDefinition( 
    in UnionTypeDefinitionInput ) ( *UnionTypeDefinition, error ) {

    if in.Len() == 0 { panic( libErrorf( "attempt to create empty union" ) ) }
    typs, errGroups := checkUnionType( in )
    if len( errGroups ) == 0 { return &UnionTypeDefinition{ typs }, nil }
    return nil, &UnionTypeDefinitionError{ errGroups } 
}

type unionTypesInput []mg.TypeReference

func ( i unionTypesInput ) Len() int { return len( i ) }

func ( i unionTypesInput ) TypeAtIndex( idx int ) mg.TypeReference {
    return i[ idx ]
}

func MustUnionTypeDefinitionTypes( 
    typs ...mg.TypeReference ) *UnionTypeDefinition {

    res, err := CreateUnionTypeDefinition( unionTypesInput( typs ) )
    if err == nil { return res }
    panic( err )
}

type FieldDefinition struct {
    Name *mg.Identifier
    Type mg.TypeReference
    Default mg.Value
}

func ( fd *FieldDefinition ) GetDefault() mg.Value {
    if fd.Default == nil {
        if lt, ok := fd.Type.( *mg.ListTypeReference ); ok {
            if lt.AllowsEmpty { return mg.EmptyList() }
        }
        return nil
    }
    return fd.Default // may still be nil
}

func ( fd *FieldDefinition ) Equals( fd2 *FieldDefinition ) bool {
    if ! fd.Name.Equals( fd2.Name ) { return false }
    if ! fd.Type.Equals( fd2.Type ) { return false }
    if fd.Default == nil { return fd2.Default == nil }
    if fd2.Default == nil { return false }
    return mg.EqualValues( fd.Default, fd2.Default )
}

type FieldSet struct {
    flds *mg.IdentifierMap
}

func NewFieldSet() *FieldSet { return &FieldSet{ mg.NewIdentifierMap() } }

func ( fs *FieldSet ) Len() int { return fs.flds.Len() }

func ( fs *FieldSet ) Get( id *mg.Identifier ) *FieldDefinition {
    if fd := fs.flds.Get( id ); fd != nil { return fd.( *FieldDefinition ) }
    return nil
}

func ( fs *FieldSet ) GetFieldNames() []*mg.Identifier {
    res := make( []*mg.Identifier, 0, fs.flds.Len() )
    fs.flds.EachPair( func( id *mg.Identifier, _ interface{} ) {
        res = append( res, id )
    })
    return res
}

func ( fs *FieldSet ) Add( fd *FieldDefinition ) error {
    return fs.flds.PutSafe( fd.Name, fd )
}

func ( fs *FieldSet ) MustAdd( fd *FieldDefinition ) {
    if nm := fd.Name; fs.flds.HasKey( nm ) {
        panic( fmt.Errorf( "FieldSet already has field: %s", nm ) )
    } else { fs.flds.Put( nm, fd ) }
}

func ( fs *FieldSet ) MustAddAll( fs2 *FieldSet ) {
    fs2.EachDefinition( func( fd *FieldDefinition ) { fs.MustAdd( fd ) } )
}

func ( fs *FieldSet ) EachDefinition( f func( fd *FieldDefinition ) ) {
    fs.flds.EachPair( func( id *mg.Identifier, fd interface{} ) {
        f( fd.( *FieldDefinition ) )
    })
}

func ( fs *FieldSet ) ContainsFields( fs2 *FieldSet ) bool {
    res := true
    fs2.EachDefinition( func( fd *FieldDefinition ) {
        if ! res { return }
        if fd2 := fs.Get( fd.Name ); fd2 == nil {
            res = false
        } else { res = fd.Equals( fd2 ) }
    })
    return res
}

type FieldContainer interface { GetFields() *FieldSet }

type CallSignature struct {
    Fields *FieldSet
    Return mg.TypeReference
    Throws *UnionTypeDefinition
}

func NewCallSignature() *CallSignature {
    return &CallSignature{ Fields: NewFieldSet() }
}

func ( cs *CallSignature ) GetFields() *FieldSet { return cs.Fields }

type PrototypeDefinition struct {
    Name *mg.QualifiedTypeName
    Signature *CallSignature
}

func ( pd *PrototypeDefinition ) GetName() *mg.QualifiedTypeName {
    return pd.Name
}

type StructDefinition struct {
    Name *mg.QualifiedTypeName
    Fields *FieldSet
    Constructors *UnionTypeDefinition
}

func NewStructDefinition() *StructDefinition {
    return &StructDefinition{ Fields: NewFieldSet() }
}

func ( sd *StructDefinition ) GetName() *mg.QualifiedTypeName {
    return sd.Name
}

func ( sd *StructDefinition ) GetFields() *FieldSet { return sd.Fields }

func ( sd *StructDefinition ) MustMixinSchema( schema *SchemaDefinition ) {
    sd.Fields.MustAddAll( schema.Fields )
}

func ( sd *StructDefinition ) SatisfiesSchema( sc *SchemaDefinition ) bool {
    return sd.Fields.ContainsFields( sc.Fields )
}

type SchemaDefinition struct {
    Name *mg.QualifiedTypeName
    Fields *FieldSet
}

func NewSchemaDefinition() *SchemaDefinition {
    return &SchemaDefinition{ Fields: NewFieldSet() }
}

func ( sd *SchemaDefinition ) GetName() *mg.QualifiedTypeName { return sd.Name }

func ( sd *SchemaDefinition ) GetFields() *FieldSet { return sd.Fields }

type AliasedTypeDefinition struct {
    Name *mg.QualifiedTypeName
    AliasedType mg.TypeReference
}

func ( ad *AliasedTypeDefinition ) GetName() *mg.QualifiedTypeName {
    return ad.Name
}

type EnumValueMap struct { m *mg.IdentifierMap }

func ( m *EnumValueMap ) Get( id *mg.Identifier ) *mg.Enum {
    if res := m.m.Get( id ); res != nil { return res.( *mg.Enum ) }
    return nil
}

type EnumDefinition struct {
    Name *mg.QualifiedTypeName
    Values []*mg.Identifier
}

func ( ed *EnumDefinition ) GetName() *mg.QualifiedTypeName { return ed.Name }

func ( ed *EnumDefinition ) GetValueMap() *EnumValueMap {
    res := &EnumValueMap{ mg.NewIdentifierMap() }
    for _, val := range ed.Values {
        res.m.Put( val, &mg.Enum{ Type: ed.GetName(), Value: val } )
    }
    return res
}

func ( ed *EnumDefinition ) GetValue( id *mg.Identifier ) *mg.Enum {
    for _, val := range ed.Values {
        if val.Equals( id ) { 
            return &mg.Enum{ Type: ed.GetName(), Value: val } 
        }
    }
    return nil
}

type OperationDefinition struct {
    Name *mg.Identifier
    Signature *CallSignature
}

func OpDefsByName( defs []*OperationDefinition ) *mg.IdentifierMap {
    res := mg.NewIdentifierMap()
    for _, def := range defs { res.Put( def.Name, def ) }
    return res
}

type ServiceDefinition struct {
    Name *mg.QualifiedTypeName
    Operations []*OperationDefinition
    Security *mg.QualifiedTypeName
}

func NewServiceDefinition() *ServiceDefinition {
    return &ServiceDefinition{ Operations: []*OperationDefinition{} }
}

func ( sd *ServiceDefinition ) GetName() *mg.QualifiedTypeName {
    return sd.Name
}

func ( sd *ServiceDefinition ) findOperation( 
    op *mg.Identifier ) *OperationDefinition {
    for _, od := range sd.Operations { if od.Name.Equals( op ) { return od } }
    return nil
}

func ( sd *ServiceDefinition ) mustFindOperation(
    op *mg.Identifier ) *OperationDefinition {
    res := sd.findOperation( op )
    if ( res != nil ) { return res }
    panic( libErrorf( "service %s has no operation %s", sd.Name, op ) )
}
