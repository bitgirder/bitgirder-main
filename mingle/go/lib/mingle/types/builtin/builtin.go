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
}

func NewLocatableErrorDefinition( 
    qn *mg.QualifiedTypeName ) *types.StructDefinition {

    res := types.NewStructDefinition()
    res.Fields.MustAdd(
        &types.FieldDefinition{
            Name: idUnsafe( "message" ),
            Type: mg.MustNullableTypeReference( mg.TypeString ),
        },
    )
    res.Fields.MustAdd(
        &types.FieldDefinition{
            Name: idUnsafe( "location" ),
            Type: mg.MustNullableTypeReference( 
                mg.NewPointerTypeReference( TypeIdentifierPath ),
            ),
        },
    )
    res.Name = qn
    return res
}

var builtinTypes *types.DefinitionMap

func BuiltinTypes() *types.DefinitionMap {
    res := types.NewDefinitionMap()
    res.MustAddFrom( builtinTypes )
    return res
}

func MustAddBuiltinType( def types.Definition ) { builtinTypes.MustAdd( def ) }

func asCoreV1Qn( nm string ) *mg.QualifiedTypeName {
    return mg.NewDeclaredTypeNameUnsafe( nm ).ResolveIn( mg.CoreNsV1 )
}

func initCoreV1Prims() {
    for _, primTyp := range mg.PrimitiveTypes {
        pd := &types.PrimitiveDefinition{}
        pd.Name = primTyp.Name
        MustAddBuiltinType( pd )
    }
}

func initCoreV1ValueTypes() {
    MustAddBuiltinType( &types.PrimitiveDefinition{ Name: mg.QnameValue } )
}

func initCoreV1StandardError() *types.SchemaDefinition {
    ed := types.NewSchemaDefinition()
    ed.Name = asCoreV1Qn( "StandardError" )
    fd := &types.FieldDefinition{
        Name: idUnsafe( "message" ),
        Type: mg.MustNullableTypeReference( mg.TypeString ),
    }
    ed.Fields.MustAdd( fd )
    MustAddBuiltinType( ed )
    return ed
}

func newV1StandardError( 
    nm string, 
    ns *mg.Namespace, 
    stdErr *types.SchemaDefinition ) *types.StructDefinition {

    res := types.NewStructDefinition()
    res.Name = mg.NewDeclaredTypeNameUnsafe( nm ).ResolveIn( ns )
    res.MustMixinSchema( stdErr )
    return res
}

func newCoreV1StandardError( 
    nm string, stdErr *types.SchemaDefinition ) *types.StructDefinition {

    return newV1StandardError( nm, mg.CoreNsV1, stdErr )
}

func initCoreV1MissingFieldsError( stdErr *types.SchemaDefinition ) {
    ed := newCoreV1StandardError( "MissingFieldsError", stdErr )
    MustAddBuiltinType( ed )
}

func initCoreV1UnrecognizedFieldError( stdErr *types.SchemaDefinition ) {
    ed := newCoreV1StandardError( "UnrecognizedFieldError", stdErr )
    MustAddBuiltinType( ed )
}

func initCoreV1ValueCastError( stdErr *types.SchemaDefinition ) {
    ed := newCoreV1StandardError( "ValueCastError", stdErr )
    MustAddBuiltinType( ed )
}

func initServiceV1EndpointError( stdErr *types.SchemaDefinition ) {
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

func initLangV1Types() {
    idPartType := initIdentifierPartType()
    initIdentifierType( idPartType )
    initNamespaceType()
    initIdentifierPathType()
}

func initBuiltinTypes() {
    builtinTypes = types.NewDefinitionMap()
    initCoreV1Prims()
    initCoreV1ValueTypes()
    initCoreV1Exceptions()
    initLangV1Types()
}
