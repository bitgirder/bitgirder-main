package builtin

import (
    "mingle/types"
    mg "mingle"
)

var (
    nsLangV1 *mg.Namespace
    QnameIdentifier *mg.QualifiedTypeName
    TypeIdentifier *mg.AtomicTypeReference
    QnameNamespace *mg.QualifiedTypeName
    TypeNamespace *mg.AtomicTypeReference
    QnameIdentifierPath *mg.QualifiedTypeName
    TypeIdentifierPath *mg.AtomicTypeReference
    QnameCastError *mg.QualifiedTypeName
    TypeCastError *mg.AtomicTypeReference
    QnameUnrecognizedFieldError *mg.QualifiedTypeName
    TypeUnrecognizedFieldError *mg.AtomicTypeReference
    QnameMissingFieldsError *mg.QualifiedTypeName
    TypeMissingFieldsError *mg.AtomicTypeReference
)

func idUnsafe( parts ...string ) *mg.Identifier {
    return mg.NewIdentifierUnsafe( parts )
}

func mkPair( 
    ns *mg.Namespace, 
    nm string ) ( *mg.QualifiedTypeName, *mg.AtomicTypeReference ) { 

    declNm := mg.NewDeclaredTypeNameUnsafe( nm )
    qn := &mg.QualifiedTypeName{ Namespace: ns, Name: declNm }
    return qn, qn.AsAtomicType()
}

func initNames() {
    nsLangV1 = &mg.Namespace{
        Version: idUnsafe( "v1" ),
        Parts: []*mg.Identifier{ idUnsafe( "mingle" ), idUnsafe( "lang" ) },
    }
    QnameIdentifier, TypeIdentifier = mkPair( nsLangV1, "Identifier" )
    QnameNamespace, TypeNamespace = mkPair( nsLangV1, "Namespace" )
    QnameIdentifierPath, TypeIdentifierPath = 
        mkPair( nsLangV1, "IdentifierPath" )
    QnameCastError, TypeCastError = mkPair( mg.CoreNsV1, "CastError" )
    QnameUnrecognizedFieldError, TypeUnrecognizedFieldError = 
        mkPair( mg.CoreNsV1, "UnrecognizedFieldError" )
    QnameMissingFieldsError, TypeMissingFieldsError = 
        mkPair( mg.CoreNsV1, "MissingFieldsError" )
}

func AddLocatableErrorFields( sd *types.StructDefinition ) {
    sd.Fields.MustAdd(
        &types.FieldDefinition{
            Name: idUnsafe( "message" ),
            Type: mg.MustNullableTypeReference( mg.TypeString ),
        },
    )
    sd.Fields.MustAdd(
        &types.FieldDefinition{
            Name: idUnsafe( "location" ),
            Type: mg.MustNullableTypeReference( 
                mg.NewPointerTypeReference( TypeIdentifierPath ),
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

func initCoreV1Types() {
    for _, primTyp := range mg.PrimitiveTypes {
        pd := &types.PrimitiveDefinition{}
        pd.Name = primTyp.Name
        MustAddBuiltinType( pd )
    }
    MustAddBuiltinType( &types.PrimitiveDefinition{ Name: mg.QnameValue } )
}

func langV1Qname( nm string ) *mg.QualifiedTypeName {
    return mg.NewDeclaredTypeNameUnsafe( nm ).ResolveIn( nsLangV1 )
}

func initIdentifierPartType() *types.AliasedTypeDefinition {
    res := &types.AliasedTypeDefinition{
        Name: langV1Qname( "IdentifierPart" ),
        AliasedType: &mg.AtomicTypeReference{
            Name: mg.QnameString,
            Restriction: mg.MustRegexRestriction( "^[a-z][a-z0-9]*$" ),
        },
    }
    MustAddBuiltinType( res )
    return res
}

func initIdentifierType( idPartTyp *types.AliasedTypeDefinition ) {
    sd := types.NewStructDefinition()
    sd.Name = QnameIdentifier
    sd.Fields.Add( 
        &types.FieldDefinition{
            Name: idUnsafe( "parts" ),
            Type: &mg.ListTypeReference{
                ElementType: idPartTyp.AliasedType,
                AllowsEmpty: false,
            },
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
    sd.Name = QnameNamespace
    idPtr := mg.NewPointerTypeReference( TypeIdentifier )
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: idUnsafe( "version" ),
            Type: idPtr,
        },
    )
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: idUnsafe( "parts" ),
            Type: &mg.ListTypeReference{
                ElementType: idPtr,
                AllowsEmpty: false,
            },
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
    sd.Name = QnameIdentifierPath
    sd.Fields.Add(
        &types.FieldDefinition{
            Name: idUnsafe( "parts" ),
            Type: &mg.ListTypeReference{
                ElementType: mg.TypeValue,
                AllowsEmpty: false,
            },
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
    sd.Name = QnameUnrecognizedFieldError
    AddLocatableErrorFields( sd )
    sd.Fields.MustAdd(
        &types.FieldDefinition{
            Name: idUnsafe( "field" ),
            Type: mg.NewPointerTypeReference( TypeIdentifier ),
        },
    )
    MustAddBuiltinType( sd )
}

func initMissingFieldsError() {
    sd := types.NewStructDefinition()
    sd.Name = QnameMissingFieldsError
    AddLocatableErrorFields( sd )
    sd.Fields.MustAdd(
        &types.FieldDefinition{
            Name: idUnsafe( "fields" ),
            Type: &mg.ListTypeReference{
                ElementType: mg.NewPointerTypeReference( TypeIdentifier ),
                AllowsEmpty: false,
            },
        },
    )
    MustAddBuiltinType( sd )
}

func initLangV1Types() {
    idPartType := initIdentifierPartType()
    initIdentifierType( idPartType )
    initNamespaceType()
    initIdentifierPathType()
    MustAddBuiltinType( NewLocatableErrorDefinition( QnameCastError ) )
    initUnrecognizedFieldError()
    initMissingFieldsError()
}

func initBuiltinTypes() {
    builtinTypes = types.NewDefinitionMap()
    initCoreV1Types()
    initLangV1Types()
}
