package types

import (
    mg "mingle"
)

var (
    LangNsV1 *mg.Namespace
    QnameIdentifier *mg.QualifiedTypeName
    TypeIdentifier *mg.AtomicTypeReference
    QnameNamespace *mg.QualifiedTypeName
    TypeNamespace *mg.AtomicTypeReference
    QnameIdentifierPath *mg.QualifiedTypeName
    TypeIdentifierPath *mg.AtomicTypeReference
)

func mkLangPair( 
    nm string ) ( *mg.QualifiedTypeName, *mg.AtomicTypeReference ) {

    qn := &mg.QualifiedTypeName{
        Namespace: LangNsV1,
        Name: mg.NewDeclaredTypeNameUnsafe( nm ),
    }
    return qn, qn.AsAtomicType()
}

func initNames() {
    LangNsV1 = &mg.Namespace{
        Parts: []*mg.Identifier{ idUnsafe( "mingle" ), idUnsafe( "lang" ) },
        Version: idUnsafe( "v1" ),
    }
    QnameIdentifier, TypeIdentifier = mkLangPair( "Identifier" )
    QnameNamespace, TypeNamespace = mkLangPair( "Namespace" )
    QnameIdentifierPath, TypeIdentifierPath = mkLangPair( "IdentifierPath" )
}

var v1Types *DefinitionMap

func V1Types() *DefinitionMap {
    res := NewDefinitionMap()
    res.MustAddFrom( v1Types )
    return res
}

func MustAddBuiltinType( def Definition ) { v1Types.MustAdd( def ) }

func asCoreV1Qn( nm string ) *mg.QualifiedTypeName {
    return mg.NewDeclaredTypeNameUnsafe( nm ).ResolveIn( mg.CoreNsV1 )
}

func initCoreV1Prims() {
    for _, primTyp := range mg.PrimitiveTypes {
        pd := &PrimitiveDefinition{}
        pd.Name = primTyp.Name
        MustAddBuiltinType( pd )
    }
}

func initCoreV1ValueTypes() {
    MustAddBuiltinType( &PrimitiveDefinition{ Name: mg.QnameValue } )
}

func initCoreV1StandardError() *SchemaDefinition {
    ed := NewSchemaDefinition()
    ed.Name = asCoreV1Qn( "StandardError" )
    fd := &FieldDefinition{
        Name: idUnsafe( "message" ),
        Type: mg.MustNullableTypeReference( mg.TypeString ),
    }
    ed.Fields.MustAdd( fd )
    MustAddBuiltinType( ed )
    return ed
}

func newV1StandardError( 
    nm string, ns *mg.Namespace, stdErr *SchemaDefinition ) *StructDefinition {

    res := NewStructDefinition()
    res.Name = mg.NewDeclaredTypeNameUnsafe( nm ).ResolveIn( ns )
    res.mustMixinSchema( stdErr )
    return res
}

func newCoreV1StandardError( 
    nm string, stdErr *SchemaDefinition ) *StructDefinition {

    return newV1StandardError( nm, mg.CoreNsV1, stdErr )
}

func initCoreV1MissingFieldsError( stdErr *SchemaDefinition ) {
    ed := newCoreV1StandardError( "MissingFieldsError", stdErr )
    MustAddBuiltinType( ed )
}

func initCoreV1UnrecognizedFieldError( stdErr *SchemaDefinition ) {
    ed := newCoreV1StandardError( "UnrecognizedFieldError", stdErr )
    MustAddBuiltinType( ed )
}

func initCoreV1ValueCastError( stdErr *SchemaDefinition ) {
    ed := newCoreV1StandardError( "ValueCastError", stdErr )
    MustAddBuiltinType( ed )
}

func initServiceV1EndpointError( stdErr *SchemaDefinition ) {
    ed := newCoreV1StandardError( "EndpointError", stdErr )
    MustAddBuiltinType( ed )
}

func initCoreV1Exceptions() {
    stdErr := initCoreV1StandardError()
    initCoreV1MissingFieldsError( stdErr )
    initCoreV1UnrecognizedFieldError( stdErr )
    initCoreV1ValueCastError( stdErr )
    initServiceV1EndpointError( stdErr )
}

func langV1Qname( nm string ) *mg.QualifiedTypeName {
    return mg.NewDeclaredTypeNameUnsafe( nm ).ResolveIn( LangNsV1 )
}

func initIdentifierPartType() *AliasedTypeDefinition {
    res := &AliasedTypeDefinition{
        Name: langV1Qname( "IdentifierPart" ),
        AliasedType: &mg.AtomicTypeReference{
            Name: mg.QnameString,
            Restriction: mg.MustRegexRestriction( "^[a-z][a-z0-9]*$" ),
        },
    }
    MustAddBuiltinType( res )
    return res
}

func initIdentifierType( idPartTyp *AliasedTypeDefinition ) {
    sd := NewStructDefinition()
    sd.Name = QnameIdentifier
    sd.Fields.Add( 
        &FieldDefinition{
            Name: idUnsafe( "parts" ),
            Type: &mg.ListTypeReference{
                ElementType: idPartTyp.AliasedType,
                AllowsEmpty: false,
            },
        },
    )
    sd.Constructors = append( sd.Constructors,
        &ConstructorDefinition{ mg.TypeString },
        &ConstructorDefinition{ mg.TypeBuffer },
    )
    MustAddBuiltinType( sd )
}

func initNamespaceType() {
    sd := NewStructDefinition()
    sd.Name = QnameNamespace
    idPtr := mg.NewPointerTypeReference( TypeIdentifier )
    sd.Fields.Add(
        &FieldDefinition{
            Name: idUnsafe( "version" ),
            Type: idPtr,
        },
    )
    sd.Fields.Add(
        &FieldDefinition{
            Name: idUnsafe( "parts" ),
            Type: &mg.ListTypeReference{
                ElementType: idPtr,
                AllowsEmpty: false,
            },
        },
    )
    sd.Constructors = append( sd.Constructors,
        &ConstructorDefinition{ mg.TypeString },
        &ConstructorDefinition{ mg.TypeBuffer },
    )
    MustAddBuiltinType( sd )
}

func initIdentifierPathType() {
    sd := NewStructDefinition()
    sd.Name = QnameIdentifierPath
    sd.Fields.Add(
        &FieldDefinition{
            Name: idUnsafe( "parts" ),
            Type: &mg.ListTypeReference{
                ElementType: mg.TypeValue,
                AllowsEmpty: false,
            },
        },
    )
    sd.Constructors = append( sd.Constructors,
        &ConstructorDefinition{ mg.TypeString },
        &ConstructorDefinition{ mg.TypeBuffer },
    )
    MustAddBuiltinType( sd )
}

func initLangV1Types() {
    idPartType := initIdentifierPartType()
    initIdentifierType( idPartType )
    initNamespaceType()
    initIdentifierPathType()
}

func initV1Types() {
    v1Types = NewDefinitionMap()
    initCoreV1Prims()
    initCoreV1ValueTypes()
    initCoreV1Exceptions()
    initLangV1Types()
}

// package note: not safe to call before completion of package init
func NewV1DefinitionMap() *DefinitionMap {
    res := NewDefinitionMap()
    res.MustAddFrom( v1Types )
    v1Types.EachDefinition( func( def Definition ) {
        res.setBuiltIn( def.GetName() )
    })
    return res
}
