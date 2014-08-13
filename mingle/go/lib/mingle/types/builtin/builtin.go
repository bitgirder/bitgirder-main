package builtin

import (
    "mingle/types"
    mg "mingle"
)

var (
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
    nsLangV1 := &mg.Namespace{
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
                mg.NewPointerTypeReference( types.TypeIdentifierPath ),
            ),
        },
    )
    res.Name = qn
    return res
}
