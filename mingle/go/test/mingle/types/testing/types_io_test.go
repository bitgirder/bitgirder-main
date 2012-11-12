package testing

import (
    gotest "testing"
    mg "mingle"
    "mingle/types"
    "bitgirder/assert"
    "bytes"
)

func TestTypesIo( t *gotest.T ) {
    m1 := types.NewDefinitionMap()
    for _, def := range []types.Definition{
        &types.PrimitiveDefinition{ mkQn( "ns1@v1/Prim1" ) },
        &types.AliasedTypeDefinition{
            Name: mkQn( "ns1@v1/A1" ),
            AliasedType: mkTyp( "ns1@v2/T1" ),
        },
        &types.PrototypeDefinition{
            Name: mkQn( "ns1@v1/Proto1" ),
            Signature: MakeCallSig(
                []*types.FieldDefinition{},
                "ns1@v1/T1",
                []string{},
            ),
        },
        &types.PrototypeDefinition{
            Name: mkQn( "ns1@v1/Proto2" ),
            Signature: MakeCallSig(
                []*types.FieldDefinition{
                    MakeFieldDef( "f1", "ns1@v1/T", nil ),
                    MakeFieldDef( "f2", "ns1@v1/T", int32( 1 ) ),
                },
                "ns1@v1/T1",
                []string{ "ns1@v1/T1", "ns1@v1/T2" },
            ),
        },
        MakeStructDef( "ns1@v1/Struct1", "", []*types.FieldDefinition{} ),
        MakeStructDef2( 
            "ns1@v1/Struct2",
            "ns1@v1/Struct1",
            []*types.FieldDefinition{
                MakeFieldDef( "f1", "ns1@v1/T", int32( 1 ) ),
            },
            []*types.ConstructorDefinition{ { Type: mkTyp( "ns1@v1/T" ) } },
        ),
        &types.EnumDefinition{
            Name: mkQn( "ns1@v1/En1" ),
            Values: []*mg.Identifier{ mkId( "e1" ), mkId( "e2" ) },
        },
        MakeServiceDef( "ns1@v1/Svc1", "", []*types.OperationDefinition{}, "" ),
        MakeServiceDef(
            "ns1@v1/Svc2",
            "ns1@v1/Svc1",
            []*types.OperationDefinition{
                {   Name: mkId( "op1" ),
                    Signature: MakeCallSig(
                        []*types.FieldDefinition{
                            MakeFieldDef( "f1", "ns1@v1/T", nil ),
                            MakeFieldDef( "f2", "ns1@v1/T", int32( 1 ) ),
                        },
                        "ns1@v1/T",
                        []string{ "ns1@v1/Ex1", "ns1@v1/Ex2" },
                    ),
                },
                {   Name: mkId( "op2" ),
                    Signature: MakeCallSig(
                        []*types.FieldDefinition{},
                        "ns1@v1/T",
                        []string{},
                    ),
                },
            },
            "ns1@v1/Security1",
        ),
    } {
        m1.MustAdd( def )
    }
    bb := &bytes.Buffer{}
    rd, wr := types.NewBinReader( bb ), types.NewBinWriter( bb )
    tailExpct := "trailing-data"
    if err := wr.WriteDefinitionMap( m1 ); err == nil {
        bb.WriteString( tailExpct )
    } else { t.Fatal( err ) }
    if m2, err := rd.ReadDefinitionMap(); err == nil {
        NewDefAsserter( t ).AssertDefMaps( m1, m2 )
        assert.Equal( tailExpct, bb.String() )
    } else { t.Fatal( err ) }
}

// To test:
//  - ReadDefinitionMap() detects duplicate map entries
