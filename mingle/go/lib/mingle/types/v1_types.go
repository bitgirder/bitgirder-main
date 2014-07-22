package types

import (
    mg "mingle"
)

var v1Types *DefinitionMap

func V1Types() *DefinitionMap {
    res := NewDefinitionMap()
    res.MustAddFrom( v1Types )
    return res
}

func asCoreV1Qn( nm string ) *mg.QualifiedTypeName {
    return mg.NewDeclaredTypeNameUnsafe( nm ).ResolveIn( mg.CoreNsV1 )
}

func initCoreV1Prims() {
    for _, primTyp := range mg.PrimitiveTypes {
        pd := &PrimitiveDefinition{}
        pd.Name = primTyp.Name
        v1Types.MustAdd( pd )
    }
}

func initCoreV1ValueTypes() {
    v1Types.MustAdd( &PrimitiveDefinition{ Name: mg.QnameValue } )
}

func initCoreV1StandardError() *SchemaDefinition {
    ed := NewSchemaDefinition()
    ed.Name = asCoreV1Qn( "StandardError" )
    fd := &FieldDefinition{
        Name: mg.NewIdentifierUnsafe( []string{ "message" } ),
        Type: mg.MustNullableTypeReference( mg.TypeString ),
    }
    ed.Fields.MustAdd( fd )
    v1Types.MustAdd( ed )
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
    v1Types.MustAdd( ed )
}

func initCoreV1UnrecognizedFieldError( stdErr *SchemaDefinition ) {
    ed := newCoreV1StandardError( "UnrecognizedFieldError", stdErr )
    v1Types.MustAdd( ed )
}

func initCoreV1ValueCastError( stdErr *SchemaDefinition ) {
    ed := newCoreV1StandardError( "ValueCastError", stdErr )
    v1Types.MustAdd( ed )
}

func initServiceV1EndpointError( stdErr *SchemaDefinition ) {
    ed := newCoreV1StandardError( "EndpointError", stdErr )
    v1Types.MustAdd( ed )
}

func initCoreV1Exceptions() {
    stdErr := initCoreV1StandardError()
    initCoreV1MissingFieldsError( stdErr )
    initCoreV1UnrecognizedFieldError( stdErr )
    initCoreV1ValueCastError( stdErr )
    initServiceV1EndpointError( stdErr )
}

func langV1Qname( nm string ) *mg.QualifiedTypeName {
    return mg.NewDeclaredTypeNameUnsafe( nm ).ResolveIn( mg.LangNsV1 )
}

func initIdentifierPartType() *AliasedTypeDefinition {
    res := &AliasedTypeDefinition{
        Name: langV1Qname( "IdentifierPart" ),
        AliasedType: &mg.AtomicTypeReference{
            Name: mg.QnameString,
            Restriction: mg.MustRegexRestriction( "^[a-z][a-z0-9]*$" ),
        },
    }
    v1Types.MustAdd( res )
    return res
}

func initIdentifierType( idPartTyp *AliasedTypeDefinition ) {
    sd := NewStructDefinition()
    sd.Name = mg.QnameIdentifier
    sd.Fields.Add( 
        &FieldDefinition{
            Name: mg.NewIdentifierUnsafe( []string{ "parts" } ),
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
    v1Types.MustAdd( sd )
}

func initLangV1Types() {
    idPartType := initIdentifierPartType()
    initIdentifierType( idPartType )
}

func init() {
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
