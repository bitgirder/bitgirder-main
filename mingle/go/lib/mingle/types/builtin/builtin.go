package builtin

import (
    "mingle/types"
    mg "mingle"
)

func idUnsafe( parts ...string ) *mg.Identifier {
    return mg.NewIdentifierUnsafe( parts )
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
            Name: idUnsafe( "message" ),
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

func initIdentifierPartType() *types.AliasedTypeDefinition {
    nm := mg.NewDeclaredTypeNameUnsafe( "IdentifierPart" )
    qn := nm.ResolveIn( mg.CoreNsV1 )
    res := &types.AliasedTypeDefinition{
        Name: qn,
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
    sd.Name = mg.QnameIdentifier
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
    sd.Name = mg.QnameNamespace
    idPtr := mg.NewPointerTypeReference( mg.TypeIdentifier )
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
    sd.Name = mg.QnameIdentifierPath
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
    sd.Name = mg.QnameUnrecognizedFieldError
    AddLocatableErrorFields( sd )
    sd.Fields.MustAdd(
        &types.FieldDefinition{
            Name: idUnsafe( "field" ),
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
            Name: idUnsafe( "fields" ),
            Type: &mg.ListTypeReference{
                ElementType: mg.NewPointerTypeReference( mg.TypeIdentifier ),
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
    MustAddBuiltinType( NewLocatableErrorDefinition( mg.QnameCastError ) )
    initUnrecognizedFieldError()
    initMissingFieldsError()
}

func initBuiltinTypes() {
    builtinTypes = types.NewDefinitionMap()
    initCoreV1Types()
    initLangV1Types()
}
