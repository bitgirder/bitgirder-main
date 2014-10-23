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
    identifierTypes = idUnsafe( "types" )
    identifierUnion = idUnsafe( "union" )
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
    
    typeIdentifierPointerList = 
        &mg.ListTypeReference{ ElementType: typeIdentifierPointer }

    typeIdentifierPathPartsList = 
        &mg.ListTypeReference{ ElementType: mg.TypeIdentifierPathPart }

    typeNonEmptyStringList = &mg.ListTypeReference{ ElementType: mg.TypeString }

    typeNonEmptyBufferList = &mg.ListTypeReference{ ElementType: mg.TypeBuffer }

    typesNs = &mg.Namespace{
        Parts: []*mg.Identifier{ idUnsafe( "mingle" ), idUnsafe( "types" ) },
        Version: idUnsafe( "v1" ),
    }

    QnamePrimitiveDefinition, TypePrimitiveDefinition = 
        mkTypesQnTypPair( "PrimitiveDefinition" )

    QnameFieldDefinition, TypeFieldDefinition = 
        mkTypesQnTypPair( "FieldDefinition" )

    typeFieldDefList = &mg.ListTypeReference{ 
        ptrTyp( TypeFieldDefinition ),
        true,
    }

    QnameFieldSet, TypeFieldSet = mkTypesQnTypPair( "FieldSet" )

    QnameCallSignature, TypeCallSignature = mkTypesQnTypPair( "CallSignature" )

    QnamePrototypeDefinition, TypePrototypeDefinition = 
        mkTypesQnTypPair( "PrototypeDefinition" )

    QnameUnionTypeDefinition, TypeUnionTypeDefinition = 
        mkTypesQnTypPair( "UnionTypeDefinition" )

    typeUnionTypeTypesList = 
        &mg.ListTypeReference{ ElementType: mg.TypeTypeReference }

    QnameUnionDefinition, TypeUnionDefinition = 
        mkTypesQnTypPair( "UnionDefinition" )

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

    typeOpDefList = &mg.ListTypeReference{
        ElementType: ptrTyp( TypeOperationDefinition ),
        AllowsEmpty: true,
    }

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
var nilTyp = mg.MustNullableTypeReference

func nilPtrTyp( typ mg.TypeReference ) *mg.NullableTypeReference {
    return nilTyp( ptrTyp( typ ) )
}

func mustAddBuiltinStruct( 
    qn *mg.QualifiedTypeName, 
    flds ...*types.FieldDefinition ) *types.StructDefinition {

    sd := types.NewStructDefinition()
    sd.Name = qn
    for _, fld := range flds { sd.Fields.Add( fld ) }
    MustAddBuiltinType( sd )
    return sd
}

func mkField0( 
    nm *mg.Identifier, typ mg.TypeReference ) *types.FieldDefinition {

    return &types.FieldDefinition{ Name: nm, Type: typ }
}

func initPrimitiveV1Types() {
    for _, primTyp := range mg.PrimitiveTypes {
        pd := &types.PrimitiveDefinition{}
        pd.Name = primTyp.Name()
        MustAddBuiltinType( pd )
    }
    MustAddBuiltinType( &types.PrimitiveDefinition{ Name: mg.QnameValue } )
}

func initCoreV1Types() {
    initPrimitiveV1Types()
    stdErr := types.NewSchemaDefinition()
    stdErr.Name = mg.QnameStandardError
    stdErr.Fields.Add( mkField0( identifierMessage, nilTyp( mg.TypeString ) ) )
    MustAddBuiltinType( stdErr )
    idDef := mustAddBuiltinStruct( mg.QnameIdentifier,
        mkField0( identifierParts, typeIdentifierPartsList ),
    )
    idDef.Constructors = 
        types.MustUnionTypeDefinitionTypes( mg.TypeString, mg.TypeBuffer )
    declNmNameType := mg.NewAtomicTypeReference(
        mg.QnameString,
        mg.MustRegexRestriction( mg.DeclaredTypeNameRegexp.String() ),
    )
    mustAddBuiltinStruct( mg.QnameDeclaredTypeName,
        mkField0( identifierName, declNmNameType ),
    )
    nsDef := mustAddBuiltinStruct( mg.QnameNamespace,
        mkField0( identifierVersion, typeIdentifierPointer ),
        mkField0( identifierParts, typeIdentifierPointerList ),
    )
    nsDef.Constructors = 
        types.MustUnionTypeDefinitionTypes( mg.TypeString, mg.TypeBuffer )
    mustAddBuiltinStruct( mg.QnameQualifiedTypeName,
        mkField0( identifierNamespace, ptrTyp( mg.TypeNamespace ) ),
        mkField0( identifierName, ptrTyp( mg.TypeDeclaredTypeName ) ),
    )
    mustAddBuiltinStruct( mg.QnameRangeRestriction,
        &types.FieldDefinition{
            Name: identifierMinClosed,
            Type: mg.TypeBoolean,
            Default: mg.Boolean( false ),
        },
        mkField0( identifierMin, mg.TypeNullableValue ),
        mkField0( identifierMax, mg.TypeNullableValue ),
        &types.FieldDefinition{
            Name: identifierMaxClosed,
            Type: mg.TypeBoolean,
            Default: mg.Boolean( false ),
        },
    )
    mustAddBuiltinStruct( mg.QnameRegexRestriction,
        mkField0( identifierPattern, mg.TypeString ),
    )
    MustAddBuiltinType(
        &types.UnionDefinition{
            Name: mg.QnameValueRestriction,
            Union: types.MustUnionTypeDefinitionTypes(
                mg.TypeRangeRestriction,
                mg.TypeRegexRestriction,
            ),
        },
    )
    mustAddBuiltinStruct( mg.QnameAtomicTypeReference,
        mkField0( identifierName, ptrTyp( mg.TypeQualifiedTypeName ) ),
        mkField0( identifierRestriction, nilPtrTyp( mg.TypeValueRestriction ) ),
    )
    mustAddBuiltinStruct( mg.QnameListTypeReference,
        mkField0( identifierElementType, mg.TypeTypeReference ),
        mkField0( identifierAllowsEmpty, mg.TypeBoolean ),
    )
    mustAddBuiltinStruct( mg.QnameNullableTypeReference,
        mkField0( identifierType, mg.TypeTypeReference ),
    )
    mustAddBuiltinStruct( mg.QnamePointerTypeReference,
        mkField0( identifierType, mg.TypeTypeReference ),
    )
    MustAddBuiltinType(
        &types.UnionDefinition{
            Name: mg.QnameTypeReference,
            Union: types.MustUnionTypeDefinitionTypes(
                mg.TypeAtomicTypeReference,
                mg.TypeListTypeReference,
                mg.TypeNullableTypeReference,
                mg.TypePointerTypeReference,
            ),
        },
    )
    MustAddBuiltinType(
        &types.UnionDefinition{
            Name: mg.QnameIdentifierPathPart,
            Union: types.MustUnionTypeDefinitionTypes(
                mg.TypeString,
                mg.TypeBuffer,
                mg.TypeIdentifier,
                mg.TypeUint64,
            ),
        },
    )
    idPathDef := mustAddBuiltinStruct( mg.QnameIdentifierPath,
        mkField0( identifierParts, typeIdentifierPathPartsList ),
    )
    idPathDef.Constructors = 
        types.MustUnionTypeDefinitionTypes( mg.TypeString, mg.TypeBuffer )
    MustAddBuiltinType( NewLocatableErrorDefinition( mg.QnameInputError ) )
    ufeDef := mustAddBuiltinStruct( mg.QnameUnrecognizedFieldError,
        mkField0( identifierField, ptrTyp( mg.TypeIdentifier ) ),
    )
    AddLocatableErrorFields( ufeDef )
    mfeDef := mustAddBuiltinStruct( mg.QnameMissingFieldsError,
        mkField0( identifierFields, typeIdentifierPointerList ),
    )
    AddLocatableErrorFields( mfeDef )
}

func initTypesTypes() {
    mustAddBuiltinStruct( QnamePrimitiveDefinition,
        mkField0( identifierName, ptrTyp( mg.TypeQualifiedTypeName ) ),
    )
    mustAddBuiltinStruct( QnameFieldDefinition,
        mkField0( identifierName, typeIdentifierPointer ),
        mkField0( identifierType, mg.TypeTypeReference ),
        mkField0( identifierDefault, mg.TypeNullableValue ),
    )
    mustAddBuiltinStruct( QnameFieldSet,
        mkField0( identifierFields, typeFieldDefList ),
    )
    mustAddBuiltinStruct( QnameUnionTypeDefinition,
        mkField0( identifierTypes, typeUnionTypeTypesList ),
    )
    mustAddBuiltinStruct( QnameUnionDefinition,
        mkField0( identifierName, ptrTyp( mg.TypeQualifiedTypeName ) ),
        mkField0( identifierUnion, ptrTyp( TypeUnionTypeDefinition ) ),
    )
    mustAddBuiltinStruct( QnameCallSignature,
        mkField0( identifierFields, ptrTyp( TypeFieldSet ) ),
        mkField0( identifierReturn, mg.TypeTypeReference ),
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
        mkField0( identifierAliasedType, mg.TypeTypeReference ),
    )
    mustAddBuiltinStruct( QnameEnumDefinition,
        mkField0( identifierName, ptrTyp( mg.TypeQualifiedTypeName ) ),
        mkField0( identifierValues, typeIdentifierPointerList ),
    )
    mustAddBuiltinStruct( QnameOperationDefinition,
        mkField0( identifierName, ptrTyp( mg.TypeIdentifier ) ),
        mkField0( identifierSignature, ptrTyp( TypeCallSignature ) ),
    )
    mustAddBuiltinStruct( QnameServiceDefinition,
        mkField0( identifierName, ptrTyp( mg.TypeQualifiedTypeName ) ),
        mkField0( identifierOperations, typeOpDefList ),
        mkField0( identifierSecurity, nilPtrTyp( mg.TypeQualifiedTypeName ) ),
    ) 
}

func initBuiltinTypes() {
    builtinTypes = types.NewDefinitionMap()
    initCoreV1Types()
    initTypesTypes()
}
