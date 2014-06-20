package types

import (
    "fmt"
//    "log"
    mg "mingle"
)

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

func ( m *DefinitionMap ) GetOk( 
    qn *mg.QualifiedTypeName ) ( Definition, bool ) {
    d, ok := m.m.GetOk( qn  )
    if ok { return d.( Definition ), true }
    return nil, false
}

func ( m *DefinitionMap ) Get( qn *mg.QualifiedTypeName ) Definition {
    if d, ok := m.GetOk( qn ); ok { return d }
    return nil
}

func ( m *DefinitionMap ) MustGet( qn *mg.QualifiedTypeName ) Definition {
    if res, ok := m.GetOk( qn ); ok { return res }
    panic( libErrorf( "no definition for type: %s", qn ) )
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
    GetSuperType() *mg.QualifiedTypeName
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
    SuperType *mg.QualifiedTypeName
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

func ( sd *StructDefinition ) GetSuperType() *mg.QualifiedTypeName {
    return sd.SuperType
}

func ( sd *StructDefinition ) GetFields() *FieldSet { return sd.Fields }

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
    SuperType *mg.QualifiedTypeName
    Operations []*OperationDefinition
    Security *mg.QualifiedTypeName
}

func NewServiceDefinition() *ServiceDefinition {
    return &ServiceDefinition{ Operations: []*OperationDefinition{} }
}

func ( sd *ServiceDefinition ) GetName() *mg.QualifiedTypeName {
    return sd.Name
}

func ( sd *ServiceDefinition ) GetSuperType() *mg.QualifiedTypeName {
    return sd.SuperType
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

type ServiceDefinitionMap struct {
    defs *DefinitionMap
    nsMap *mg.NamespaceMap
}

func NewServiceDefinitionMap( defs *DefinitionMap ) *ServiceDefinitionMap {
    return &ServiceDefinitionMap{ defs: defs, nsMap: mg.NewNamespaceMap() }
}

func ( m *ServiceDefinitionMap ) GetDefinitionMap() *DefinitionMap {
    return m.defs
}

func ( m *ServiceDefinitionMap ) Put( 
    ns *mg.Namespace, svc *mg.Identifier, qn *mg.QualifiedTypeName ) error {
    if def, ok := m.defs.GetOk( qn ); ok {
        if sd, ok := def.( *ServiceDefinition ); ok {
            svcMap, ok := m.nsMap.GetOk( ns )
            if ! ok {
                svcMap = mg.NewIdentifierMap()
                m.nsMap.Put( ns, svcMap )
            }
            svcMap.( *mg.IdentifierMap ).Put( svc, sd )
            return nil
        }
        return libErrorf( "(%T).Put(): %s is not a service", m, qn )
    }
    return libErrorf( "(%T).Put(): no definition for name %s", qn )
}

func ( m *ServiceDefinitionMap ) MustPut(
    ns *mg.Namespace, svc *mg.Identifier, qn *mg.QualifiedTypeName ) {
    if err := m.Put( ns, svc, qn ); err != nil { panic( err ) }
}

func ( m *ServiceDefinitionMap ) HasNamespace( ns *mg.Namespace ) bool {
    return m.nsMap.HasKey( ns )
}

func ( m *ServiceDefinitionMap ) GetOk( 
    ns *mg.Namespace, svc *mg.Identifier ) ( *ServiceDefinition, bool ) {
    if svcMap, ok := m.nsMap.GetOk( ns ); ok {
        if sd, ok := svcMap.( *mg.IdentifierMap ).GetOk( svc ); ok {
            return sd.( *ServiceDefinition ), true
        }
    }
    return nil, false
}

func ( m *ServiceDefinitionMap ) MustGet(
    ns *mg.Namespace, svc *mg.Identifier ) *ServiceDefinition {
    if res, ok := m.GetOk( ns, svc ); ok { return res }
    panic( 
        libErrorf( "no service matches namespace '%s' and id '%s'", ns, svc ) )
}
