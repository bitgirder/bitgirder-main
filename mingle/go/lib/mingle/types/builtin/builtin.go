package builtin

import (
    "mingle/types"
    mg "mingle"
)

var identifierParts *mg.Identifier
var identifierVersion *mg.Identifier
var identifierLocation *mg.Identifier
var identifierMessage *mg.Identifier
var identifierField *mg.Identifier
var identifierFields *mg.Identifier
var typeIdentifierPart *mg.AtomicTypeReference
var typeIdentifierPartsList *mg.ListTypeReference
var typeIdentifierPointer *mg.PointerTypeReference
var typeIdentifierPointerList *mg.ListTypeReference
var typeIdentifierPathPartsList *mg.ListTypeReference
var typeNonEmptyStringList *mg.ListTypeReference
var typeNonEmptyBufferList *mg.ListTypeReference

func idUnsafe( parts ...string ) *mg.Identifier {
    return mg.NewIdentifierUnsafe( parts )
}

func initNames() {
    identifierParts = idUnsafe( "parts" )
    identifierVersion = idUnsafe( "version" )
    identifierLocation = idUnsafe( "location" )
    identifierMessage = idUnsafe( "message" )
    identifierField = idUnsafe( "field" )
    identifierFields = idUnsafe( "fields" )
    typeIdentifierPart = &mg.AtomicTypeReference{
        Name: mg.QnameString,
        Restriction: mg.MustRegexRestriction( "^[a-z][a-z0-9]*$" ),
    }
    typeIdentifierPartsList = &mg.ListTypeReference{
        ElementType: typeIdentifierPart,
        AllowsEmpty: false,
    }
    typeIdentifierPointer = mg.NewPointerTypeReference( mg.TypeIdentifier )
    typeIdentifierPointerList = &mg.ListTypeReference{
        ElementType: typeIdentifierPointer,
        AllowsEmpty: false,
    }
    typeIdentifierPathPartsList = &mg.ListTypeReference{
        ElementType: mg.TypeValue,
        AllowsEmpty: false,
    }
    typeNonEmptyStringList = &mg.ListTypeReference{
        ElementType: mg.TypeString,
        AllowsEmpty: false,
    }
    typeNonEmptyBufferList = &mg.ListTypeReference{
        ElementType: mg.TypeBuffer,
        AllowsEmpty: false,
    }
}

func AddLocatableErrorFields( sd *types.StructDefinition ) {
    sd.Fields.MustAdd(
        &types.FieldDefinition{
            Name: identifierMessage,
            Type: mg.MustNullableTypeReference( mg.TypeString ),
        },
    )
    sd.Fields.MustAdd(
        &types.FieldDefinition{
            Name: identifierLocation,
            Type: mg.MustNullableTypeReference( 
                mg.NewPointerTypeReference( mg.TypeIdentifierPath ),
            ),
        },
    )
}

func NewLocatableErrorDefinition( 
    qn *mg.QualifiedTypeName ) *types.StructDefinition {

    res := types.NewStructDefinition()
    res.Name = qn
    AddLocatableErrorFields( res )
    return res
}

var builtinTypes *types.DefinitionMap

func BuiltinTypes() *types.DefinitionMap {
    res := types.NewDefinitionMap()
    res.MustAddFrom( builtinTypes )
    return res
}

func MustAddBuiltinType( def types.Definition ) { builtinTypes.MustAdd( def ) }

func initStandardError() {
    sd := types.NewSchemaDefinition()
    sd.Name = mg.QnameStandardError
    sd.Fields.MustAdd(
        &types.FieldDefinition{
            Name: identifierMessage,
            Type: mg.MustNullableTypeReference( mg.TypeString ),
        },
    )
    MustAddBuiltinType( sd )
}

func initCoreV1Types() {
    for _, primTyp := range mg.PrimitiveTypes {
        pd := &types.PrimitiveDefinition{}
        pd.Name = primTyp.Name
        MustAddBuiltinType( pd )
    }
    MustAddBuiltinType( &types.PrimitiveDefinition{ Name: mg.QnameValue } )
    initStandardError()
}

func initIdentifierType() {
    sd := types.NewStructDefinition()
    sd.Name = mg.QnameIdentifier
    sd.Fields.Add( 
        &types.FieldDefinition{
            Name: identifierParts,
            Type: typeIdentifierPartsList,
        },
    )
    sd.Constructors = append( sd.Constructors,
        &types.ConstructorDefinition{ mg.TypeString },
        &types.ConstructorDefinition{ mg.TypeBuffer },
    )
    MustAddBuiltinType( sd )
}

func initNamespaceType() {
    sd := types.NewStructDefinition()
    sd.Name = mg.QnameNamespace
    sd.Fields.Add(
        &types.FieldDefinition{ 
            Name: identifierVersion, 
            Type: typeIdentifierPointer,
        },
    )
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: identifierParts,
            Type: typeIdentifierPointerList,
        },
    )
    sd.Constructors = append( sd.Constructors,
        &types.ConstructorDefinition{ mg.TypeString },
        &types.ConstructorDefinition{ mg.TypeBuffer },
    )
    MustAddBuiltinType( sd )
}

func initIdentifierPathType() {
    sd := types.NewStructDefinition()
    sd.Name = mg.QnameIdentifierPath
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: identifierParts,
            Type: typeIdentifierPathPartsList,
        },
    )
    sd.Constructors = append( sd.Constructors,
        &types.ConstructorDefinition{ mg.TypeString },
        &types.ConstructorDefinition{ mg.TypeBuffer },
    )
    MustAddBuiltinType( sd )
}

func initUnrecognizedFieldError() {
    sd := types.NewStructDefinition()
    sd.Name = mg.QnameUnrecognizedFieldError
    AddLocatableErrorFields( sd )
    sd.Fields.MustAdd(
        &types.FieldDefinition{
            Name: identifierField,
            Type: mg.NewPointerTypeReference( mg.TypeIdentifier ),
        },
    )
    MustAddBuiltinType( sd )
}

func initMissingFieldsError() {
    sd := types.NewStructDefinition()
    sd.Name = mg.QnameMissingFieldsError
    AddLocatableErrorFields( sd )
    sd.Fields.MustAdd(
        &types.FieldDefinition{
            Name: identifierFields,
            Type: typeIdentifierPointerList,
        },
    )
    MustAddBuiltinType( sd )
}

func initLangV1Types() {
    initIdentifierType()
    initNamespaceType()
    initIdentifierPathType()
    MustAddBuiltinType( NewLocatableErrorDefinition( mg.QnameCastError ) )
    initUnrecognizedFieldError()
    initMissingFieldsError()
}

func initBuiltinTypes() {
    builtinTypes = types.NewDefinitionMap()
    initCoreV1Types()
    initLangV1Types()
}
