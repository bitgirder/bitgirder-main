package builtin

import (
    "mingle/types"
    mg "mingle"
)

func mkId( parts ...string ) *mg.Identifier {
    return mg.NewIdentifierUnsafe( parts )
}

func NewLocatableErrorDefinition( 
    qn *mg.QualifiedTypeName ) *types.StructDefinition {

    res := types.NewStructDefinition()
    res.Fields.MustAdd(
        &types.FieldDefinition{
            Name: mkId( "message" ),
            Type: mg.MustNullableTypeReference( mg.TypeString ),
        },
    )
    res.Fields.MustAdd(
        &types.FieldDefinition{
            Name: mkId( "location" ),
            Type: mg.MustNullableTypeReference( 
                mg.NewPointerTypeReference( types.TypeIdentifierPath ),
            ),
        },
    )
    res.Name = qn
    return res
}
