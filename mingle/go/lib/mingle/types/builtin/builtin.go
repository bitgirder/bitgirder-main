package builtin

import (
    "mingle/types"
    mg "mingle"
)

func idUnsafe( parts ...string ) *mg.Identifier {
    return mg.NewIdentifierUnsafe( parts )
}

func mkTypesQnTypPair( 
    nm string ) ( *mg.QualifiedTypeName, *mg.AtomicTypeReference ) {

    if typesNs == nil { panic( libError( "typesNs not initialized" ) ) }
    qn := &mg.QualifiedTypeName{
        Namespace: typesNs,
        Name: mg.NewDeclaredTypeNameUnsafe( nm ),
    }
    return qn, qn.AsAtomicType()
}

var (
    identifierParts = idUnsafe( "parts" )
    identifierVersion = idUnsafe( "version" )
    identifierLocation = idUnsafe( "location" )
    identifierMessage = idUnsafe( "message" )
    identifierField = idUnsafe( "field" )
    identifierFields = idUnsafe( "fields" )
    identifierName = idUnsafe( "name" )
    identifierNamespace = idUnsafe( "namespace" )
    identifierRestriction = idUnsafe( "restriction" )
    identifierPattern = idUnsafe( "pattern" )
    identifierMinClosed = idUnsafe( "min", "closed" )
    identifierMin = idUnsafe( "min" )
    identifierMax = idUnsafe( "max" )
    identifierMaxClosed = idUnsafe( "max", "closed" )
    identifierElementType = idUnsafe( "element", "type" )
    identifierAllowsEmpty = idUnsafe( "allows", "empty" )
    identifierType = idUnsafe( "type" )

    typeIdentifierPartsList = &mg.ListTypeReference{
        ElementType: mg.NewAtomicTypeReference(
            mg.QnameString,
            mg.MustRegexRestriction( mg.IdentifierPartRegexp.String() ),
        ),
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

    typesNs = &mg.Namespace{
        Parts: []*mg.Identifier{ idUnsafe( "mingle" ), idUnsafe( "types" ) },
        Version: idUnsafe( "v1" ),
    }

    QnamePrimitiveDefinition, TypePrimitiveDefinition = 
        mkTypesQnTypPair( "PrimitiveDefinition" )

    QnameFieldDefinition, TypeFieldDefinition = 
        mkTypesQnTypPair( "FieldDefinition" )

    QnameFieldSet, TypeFieldSet = mkTypesQnTypPair( "FieldSet" )

    QnameCallSignature, TypeCallSignature = mkTypesQnTypPair( "CallSignature" )

    QnamePrototypeDefinition, TypePrototypeDefinition = 
        mkTypesQnTypPair( "PrototypeDefinition" )

    QnameUnionTypeDefinition, TypeUnionTypeDefinition = 
        mkTypesQnTypPair( "UnionTypeDefinition" )

    QnameStructDefinition, TypeStructDefinition = 
        mkTypesQnTypPair( "StructDefinition" )

    QnameSchemaDefinition, TypeSchemaDefinition = 
        mkTypesQnTypPair( "SchemaDefinition" )

    QnameAliasedTypeDefinition, TypeAliasedTypeDefinition = 
        mkTypesQnTypPair( "AliasedTypeDefinition" )

    QnameEnumDefinition, TypeEnumDefinition = 
        mkTypesQnTypPair( "EnumDefinition" )

    QnameOperationDefinition, TypeOperationDefinition = 
        mkTypesQnTypPair( "OperationDefinition" )

    QnameServiceDefinition, TypeServiceDefinition = 
        mkTypesQnTypPair( "ServiceDefinition" )
)

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
        pd.Name = primTyp.Name()
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
    sd.Constructors = 
        types.MustUnionTypeDefinitionTypes( mg.TypeString, mg.TypeBuffer )
    MustAddBuiltinType( sd )
}

func initDeclaredTypeNameType() {
    sd := types.NewStructDefinition()
    sd.Name = mg.QnameDeclaredTypeName
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: identifierName,
            Type: mg.NewAtomicTypeReference(
                mg.QnameString,
                mg.MustRegexRestriction( mg.DeclaredTypeNameRegexp.String() ),
            ),
        },
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
    sd.Constructors = 
        types.MustUnionTypeDefinitionTypes( mg.TypeString, mg.TypeBuffer )
    MustAddBuiltinType( sd )
}

func initQnameType() {
    sd := types.NewStructDefinition()
    sd.Name = mg.QnameQualifiedTypeName
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: identifierNamespace,
            Type: mg.NewPointerTypeReference( mg.TypeNamespace ),
        },
    )
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: identifierName,
            Type: mg.NewPointerTypeReference( mg.TypeDeclaredTypeName ),
        },
    )
    MustAddBuiltinType( sd )
}

func initRangeRestrictionType() {
    sd := types.NewStructDefinition()
    sd.Name = mg.QnameRangeRestriction
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: identifierMinClosed,
            Type: mg.TypeBoolean,
            Default: mg.Boolean( false ),
        },
    )
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: identifierMin,
            Type: mg.TypeNullableValue,
        },
    )
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: identifierMax,
            Type: mg.TypeNullableValue,
        },
    )
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: identifierMaxClosed,
            Type: mg.TypeBoolean,
            Default: mg.Boolean( false ),
        },
    )
    MustAddBuiltinType( sd )
}

func initRegexRestrictionType() {
    sd := types.NewStructDefinition()
    sd.Name = mg.QnameRegexRestriction
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: identifierPattern,
            Type: mg.TypeString,
        },
    )
    MustAddBuiltinType( sd )
}

func initRestrictionTypes() {
    initRangeRestrictionType()
    initRegexRestrictionType()
}

func initAtomicTypeReferenceType() {
    initRestrictionTypes()
    sd := types.NewStructDefinition()
    sd.Name = mg.QnameAtomicTypeReference
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: identifierName,
            Type: mg.NewPointerTypeReference( mg.TypeQualifiedTypeName ),
        },
    )
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: identifierRestriction,
            Type: mg.TypeNullableValue, 
        },
    )
    MustAddBuiltinType( sd )
}

func initListTypeReferenceType() {
    sd := types.NewStructDefinition()
    sd.Name = mg.QnameListTypeReference
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: identifierElementType,
            Type: mg.TypeValue, 
        },
    )
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: identifierAllowsEmpty,
            Type: mg.TypeBoolean,
        },
    )
    MustAddBuiltinType( sd )
}

func initNullableTypeReferenceType() {
    sd := types.NewStructDefinition()
    sd.Name = mg.QnameNullableTypeReference
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: identifierType,
            Type: mg.TypeValue,
        },
    )
    MustAddBuiltinType( sd )
}

func initPointerTypeReferenceType() {
    sd := types.NewStructDefinition()
    sd.Name = mg.QnamePointerTypeReference
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: identifierType,
            Type: mg.TypeValue,
        },
    )
    MustAddBuiltinType( sd )
}

func initTypeReferenceTypes() {
    initAtomicTypeReferenceType()
    initListTypeReferenceType()
    initPointerTypeReferenceType()
    initNullableTypeReferenceType()
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
    sd.Constructors = 
        types.MustUnionTypeDefinitionTypes( mg.TypeString, mg.TypeBuffer )
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
    initDeclaredTypeNameType()
    initNamespaceType()
    initQnameType()
    initTypeReferenceTypes()
    initIdentifierPathType()
    MustAddBuiltinType( NewLocatableErrorDefinition( mg.QnameCastError ) )
    initUnrecognizedFieldError()
    initMissingFieldsError()
}

func initTypesTypes() {
//type PrimitiveDefinition struct { Name *mg.QualifiedTypeName }
//type FieldDefinition struct {
//    Name *mg.Identifier
//    Type mg.TypeReference
//    Default mg.Value
//}
//type FieldSet struct {
//    flds *mg.IdentifierMap
//}
//type CallSignature struct {
//    Fields *FieldSet
//    Return mg.TypeReference
//    Throws []mg.TypeReference
//}
//type PrototypeDefinition struct {
//    Name *mg.QualifiedTypeName
//    Signature *CallSignature
//}
//type ConstructorDefinition struct { Type mg.TypeReference }
//type StructDefinition struct {
//    Name *mg.QualifiedTypeName
//    Fields *FieldSet
//    Constructors []*ConstructorDefinition
//}
//type SchemaDefinition struct {
//    Name *mg.QualifiedTypeName
//    Fields *FieldSet
//}
//type AliasedTypeDefinition struct {
//    Name *mg.QualifiedTypeName
//    AliasedType mg.TypeReference
//}
//type EnumDefinition struct {
//    Name *mg.QualifiedTypeName
//    Values []*mg.Identifier
//}
//type OperationDefinition struct {
//    Name *mg.Identifier
//    Signature *CallSignature
//}
//type ServiceDefinition struct {
//    Name *mg.QualifiedTypeName
//    Operations []*OperationDefinition
//    Security *mg.QualifiedTypeName
//}
}

func initBuiltinTypes() {
    builtinTypes = types.NewDefinitionMap()
    initCoreV1Types()
    initLangV1Types()
    initTypesTypes()
}
