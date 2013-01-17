package types

import (
    "fmt"
//    "log"
    mg "mingle"
)

type DefinitionMap struct {
    m *mg.QnameMap
}

func NewDefinitionMap() *DefinitionMap {
    return &DefinitionMap{ mg.NewQnameMap() }
}

func ( m *DefinitionMap ) Len() int { return m.m.Len() }

func ( m *DefinitionMap ) Get( qn *mg.QualifiedTypeName ) Definition {
    if d := m.m.Get( qn ); d != nil { return d.( Definition ) }
    return nil
}

func ( m *DefinitionMap ) HasKey( qn *mg.QualifiedTypeName ) bool {
    return m.m.HasKey( qn )
}

func ( m *DefinitionMap ) Add( d Definition ) error {
    return m.m.PutSafe( d.GetName(), d )
}

// sole impl for now; we may add a version Add() that returns error on dupes, in
// which case this method will call into that one
func ( m *DefinitionMap ) MustAdd( d Definition ) {
    if err := m.Add( d ); err != nil { panic( err ) }
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

type Descendant interface { 
    // Can return nil to indicate no super type
    GetSuperType() mg.TypeReference 
}

type Definition interface {
    GetName() *mg.QualifiedTypeName
}

type PrimitiveDefinition struct { Name *mg.QualifiedTypeName }

func ( pd *PrimitiveDefinition ) GetName() *mg.QualifiedTypeName {
    return pd.Name
}

type FieldDefinition struct {
    Name *mg.Identifier
    Type mg.TypeReference
    Default mg.Value
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

func ( fs *FieldSet ) EachDefinition( f func( fd *FieldDefinition ) ) {
    fs.flds.EachPair( func( id *mg.Identifier, fd interface{} ) {
        f( fd.( *FieldDefinition ) )
    })
}

type FieldContainer interface { GetFields() *FieldSet }

type CallSignature struct {
    Fields *FieldSet
    Return mg.TypeReference
    Throws []mg.TypeReference
}

func NewCallSignature() *CallSignature {
    return &CallSignature{ 
        Fields: NewFieldSet(),
        Throws: []mg.TypeReference{},
    }
}

func ( cs *CallSignature ) GetFields() *FieldSet { return cs.Fields }

type PrototypeDefinition struct {
    Name *mg.QualifiedTypeName
    Signature *CallSignature
}

func ( pd *PrototypeDefinition ) GetName() *mg.QualifiedTypeName {
    return pd.Name
}

type ConstructorDefinition struct { Type mg.TypeReference }

type StructDefinition struct {
    Name *mg.QualifiedTypeName
    SuperType mg.TypeReference
    Fields *FieldSet
    Constructors []*ConstructorDefinition
}

func NewStructDefinition() *StructDefinition {
    return &StructDefinition{ 
        Fields: NewFieldSet(),
        Constructors: []*ConstructorDefinition{},
    }
}

func ( sd *StructDefinition ) GetName() *mg.QualifiedTypeName {
    return sd.Name
}

func ( sd *StructDefinition ) GetSuperType() mg.TypeReference {
    return sd.SuperType
}

func ( sd *StructDefinition ) GetFields() *FieldSet { return sd.Fields }

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
    SuperType mg.TypeReference
    Operations []*OperationDefinition
    Security *mg.QualifiedTypeName
}

func NewServiceDefinition() *ServiceDefinition {
    return &ServiceDefinition{ Operations: []*OperationDefinition{} }
}

func ( sd *ServiceDefinition ) GetName() *mg.QualifiedTypeName {
    return sd.Name
}

var coreTypesV1 *DefinitionMap

func CoreTypesV1() *DefinitionMap {
    res := NewDefinitionMap()
    res.MustAddFrom( coreTypesV1 )
    return res
}

func asCoreV1Qn( nm string ) *mg.QualifiedTypeName {
    return mg.MustDeclaredTypeName( nm ).ResolveIn( mg.CoreNsV1 )
}

func initCoreV1Prims() {
    for _, primTyp := range mg.PrimitiveTypes {
        pd := &PrimitiveDefinition{}
        pd.Name = primTyp.Name.( *mg.QualifiedTypeName )
        coreTypesV1.MustAdd( pd )
    }
}

func initCoreV1Exceptions() {
    var ed *StructDefinition
    ed = NewStructDefinition()
    ed.Name = asCoreV1Qn( "StandardException" )
    coreTypesV1.MustAdd( ed )
}

func init() {
    coreTypesV1 = NewDefinitionMap()
    initCoreV1Prims()
    initCoreV1Exceptions()
}
