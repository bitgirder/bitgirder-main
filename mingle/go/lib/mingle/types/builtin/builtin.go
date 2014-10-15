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
    identifierAliasedType = idUnsafe( "aliased", "type" )
    identifierAllowsEmpty = idUnsafe( "allows", "empty" )
    identifierConstructors = idUnsafe( "constructors" )
    identifierDefault = idUnsafe( "default" )
    identifierElementType = idUnsafe( "element", "type" )
    identifierField = idUnsafe( "field" )
    identifierFields = idUnsafe( "fields" )
    identifierLocation = idUnsafe( "location" )
    identifierMax = idUnsafe( "max" )
    identifierMaxClosed = idUnsafe( "max", "closed" )
    identifierMessage = idUnsafe( "message" )
    identifierMin = idUnsafe( "min" )
    identifierMinClosed = idUnsafe( "min", "closed" )
    identifierName = idUnsafe( "name" )
    identifierNamespace = idUnsafe( "namespace" )
    identifierOperation = idUnsafe( "operation" )
    identifierOperations = idUnsafe( "operations" )
    identifierParts = idUnsafe( "parts" )
    identifierPattern = idUnsafe( "pattern" )
    identifierRestriction = idUnsafe( "restriction" )
    identifierReturn = idUnsafe( "return" )
    identifierSecurity = idUnsafe( "security" )
    identifierSignature = idUnsafe( "signature" )
    identifierThrows = idUnsafe( "throws" )
    identifierType = idUnsafe( "type" )
    identifierValues = idUnsafe( "values" )
    identifierVersion = idUnsafe( "version" )

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

    QnameFieldSetEntry, TypeFieldSetEntry = mkTypesQnTypPair( "FieldSetEntry" )

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

var ptrTyp = mg.NewPointerTypeReference

func nilPtrTyp( typ mg.TypeReference ) *mg.NullableTypeReference {
    return mg.MustNullableTypeReference( ptrTyp( typ ) )
}

func mustAddBuiltinStruct( 
    qn *mg.QualifiedTypeName, flds ...*types.FieldDefinition ) {

    sd := types.NewStructDefinition()
    sd.Name = qn
    for _, fld := range flds { sd.Fields.Add( fld ) }
    MustAddBuiltinType( sd )
}

func mkField0( 
    nm *mg.Identifier, typ mg.TypeReference ) *types.FieldDefinition {

    return &types.FieldDefinition{ Name: nm, Type: typ }
}

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
    mustAddBuiltinStruct( QnamePrimitiveDefinition,
        mkField0( identifierName, ptrTyp( mg.TypeQualifiedTypeName ) ),
    )
    mustAddBuiltinStruct( QnameFieldDefinition,
        mkField0( identifierName, typeIdentifierPointer ),
        mkField0( identifierType, mg.TypeValue ),
        mkField0( identifierDefault, mg.TypeNullableValue ),
    )
    mustAddBuiltinStruct( QnameFieldSetEntry,
        mkField0( identifierName, typeIdentifierPointer ),
        mkField0( identifierField, TypeFieldDefinition ),
    )
    mustAddBuiltinStruct( QnameFieldSet,
        mkField0( 
            identifierFields, 
            &mg.ListTypeReference{
                ElementType: TypeFieldSetEntry,
                AllowsEmpty: true,
            },
        ),
    )
    mustAddBuiltinStruct( QnameCallSignature,
        mkField0( identifierFields, ptrTyp( TypeFieldSet ) ),
        mkField0( identifierReturn, mg.TypeValue ),
        mkField0( identifierThrows, nilPtrTyp( TypeUnionTypeDefinition ) ),
    )
    mustAddBuiltinStruct( QnamePrototypeDefinition,
        mkField0( identifierName, ptrTyp( mg.TypeQualifiedTypeName ) ),
        mkField0( identifierSignature, ptrTyp( TypeCallSignature ) ),
    )
    mustAddBuiltinStruct( QnameStructDefinition,
        mkField0( identifierName, ptrTyp( mg.TypeQualifiedTypeName ) ),
        mkField0( identifierFields, ptrTyp( TypeFieldSet ) ),
        mkField0( 
            identifierConstructors, nilPtrTyp( TypeUnionTypeDefinition ) ),
    )
    mustAddBuiltinStruct( QnameSchemaDefinition,
        mkField0( identifierName, ptrTyp( mg.TypeQualifiedTypeName ) ),
        mkField0( identifierFields, ptrTyp( TypeFieldSet ) ),
    )
    mustAddBuiltinStruct( QnameAliasedTypeDefinition,
        mkField0( identifierName, ptrTyp( mg.TypeQualifiedTypeName ) ),
        mkField0( identifierAliasedType, mg.TypeValue ),
    )
    mustAddBuiltinStruct( QnameEnumDefinition,
        mkField0( identifierName, ptrTyp( mg.TypeQualifiedTypeName ) ),
        mkField0( 
            identifierValues,
            &mg.ListTypeReference{
                ElementType: typeIdentifierPointer,
                AllowsEmpty: false,
            },
        ),
    )
    mustAddBuiltinStruct( QnameOperationDefinition,
        mkField0( identifierName, ptrTyp( mg.TypeQualifiedTypeName ) ),
        mkField0( identifierSignature, ptrTyp( TypeCallSignature ) ),
    )
    mustAddBuiltinStruct( QnameServiceDefinition,
        mkField0( identifierName, ptrTyp( mg.TypeQualifiedTypeName ) ),
        mkField0(
            identifierOperations,
            &mg.ListTypeReference{
                ElementType: ptrTyp( TypeOperationDefinition ),
                AllowsEmpty: true,
            },
        ),
        mkField0( identifierSecurity, nilPtrTyp( mg.TypeQualifiedTypeName ) ),
    ) 
}

func initBuiltinTypes() {
    builtinTypes = types.NewDefinitionMap()
    initCoreV1Types()
    initLangV1Types()
    initTypesTypes()
}
