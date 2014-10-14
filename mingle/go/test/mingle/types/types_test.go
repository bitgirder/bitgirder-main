package types

import (
    mg "mingle"
    "testing"
    "bitgirder/assert"
)

func TestFieldGetDefault( t *testing.T ) {
    chk := func( typ string, defl, expct interface{} ) {
        fd := MakeFieldDef( "f1", typ, defl )
        assert.Equal( expct, fd.GetDefault() )
    }
    chk( "Int32", int32( 1 ), mg.Int32( 1 ) )
    chk( "Int32", nil, nil )
    for _, quant := range []string{ "*", "+" } {
        l := mg.MustList( int32( 1 ) )
        chk( "Int32" + quant, l, l )
    }
    chk( "Int32*", nil, mg.EmptyList() )
}

func TestFieldDefinitionsEqual( t *testing.T ) {
    la := assert.NewListPathAsserter( t )
    for _, t := range []struct{ f1, f2 *FieldDefinition; eq bool }{
        { 
            MakeFieldDef( "f", "Int32", nil ),
            MakeFieldDef( "f", "Int32", nil ),
            true,
        },
        { 
            MakeFieldDef( "f", "Int32", nil ),
            MakeFieldDef( "f2", "Int32", nil ),
            false,
        },
        { 
            MakeFieldDef( "f", "Int32", nil ),
            MakeFieldDef( "f", "Int64", nil ),
            false,
        },
        { 
            MakeFieldDef( "f", "Int32", mg.Int32( int32( 1 ) ) ),
            MakeFieldDef( "f", "Int32", nil ),
            false,
        },
        { 
            MakeFieldDef( "f", "Int32", mg.Int32( int32( 1 ) ) ),
            MakeFieldDef( "f", "Int32", mg.Int32( int32( 2 ) ) ),
            false,
        },
        { 
            MakeFieldDef( "f", "Int32", mg.Int32( int32( 1 ) ) ),
            MakeFieldDef( "f", "Int32", mg.Int32( int32( 1 ) ) ),
            true,
        },
    } {
        la.Equal( t.eq, t.f1.Equals( t.f2 ) )
        la = la.Next()
    }
}

func TestFieldSetContainsFields( t *testing.T ) {
    la := assert.NewListPathAsserter( t )
    for _, t := range []struct{ s1, s2 *FieldSet; eq bool }{
        {
            MakeFieldSet(),
            MakeFieldSet(),
            true,
        },
        {
            MakeFieldSet(),
            MakeFieldSet( 
                MakeFieldDef( "f1", "Int32", nil ),
            ),
            false,
        },
        {
            MakeFieldSet( 
                MakeFieldDef( "f1", "Int32", nil ),
            ),
            MakeFieldSet(),
            true,
        },
        {
            MakeFieldSet(
                MakeFieldDef( "f1", "Int32", nil ),
            ),
            MakeFieldSet(
                MakeFieldDef( "f1", "Int32", nil ),
            ),
            true,
        },
        {
            MakeFieldSet(
                MakeFieldDef( "f1", "Int32", nil ),
                MakeFieldDef( "f2", "Int32", nil ),
            ),
            MakeFieldSet(
                MakeFieldDef( "f1", "Int32", nil ),
            ),
            true,
        },
        {
            MakeFieldSet(
                MakeFieldDef( "f1", "Int32", nil ),
            ),
            MakeFieldSet(
                MakeFieldDef( "f1", "Int32", nil ),
                MakeFieldDef( "f2", "Int32", nil ),
            ),
            false,
        },
        {
            MakeFieldSet(
                MakeFieldDef( "f1", "Int32", nil ),
                MakeFieldDef( "f2", "Int32", nil ),
            ),
            MakeFieldSet(
                MakeFieldDef( "f1", "Int32", nil ),
                MakeFieldDef( "f2", "Int32", nil ),
            ),
            true,
        },
    } {
        la.Equal( t.eq, t.s1.ContainsFields( t.s2 ) )
        la = la.Next()
    }
}

type unionTypeDefinitionTest struct {
    input []mg.TypeReference
    errGroups [][]int
}

func ( t *unionTypeDefinitionTest ) Len() int { return len( t.input ) }

func ( t *unionTypeDefinitionTest ) TypeAtIndex( idx int ) mg.TypeReference {
    return t.input[ idx ]
}

func ( t *unionTypeDefinitionTest ) call( a *assert.PathAsserter ) {
    ud, err := CreateUnionTypeDefinition( t )
    if t.errGroups == nil {
        a.EqualErrors( nil, err )
        a.Equal( ud.Types, t.input )
        return
    }
    if err == nil { a.Fatalf( "expected err groups: %#v", t.errGroups ) }
    udErr, ok := err.( *UnionTypeDefinitionError )
    if ! ok { a.Fatal( err ) }
    a.Equal( t.errGroups, udErr.ErrorGroups )
}

func getUnionDefinitionTests() []*unionTypeDefinitionTest {
    res := make( []*unionTypeDefinitionTest, 0, 4 )
    res = append( res,
        &unionTypeDefinitionTest{
            input: []mg.TypeReference{
                mg.TypeInt32,
                mg.TypeUint32,
                mg.NewPointerTypeReference( mg.TypeString ),
                &mg.ListTypeReference{ mg.TypeString, false },
                mg.MustNullableTypeReference( mg.TypeSymbolMap ),
            },
        },
        &unionTypeDefinitionTest{
            input: []mg.TypeReference{ mg.TypeString, mg.TypeString },
            errGroups: [][]int{ []int{ 0, 1 } },
        },
        &unionTypeDefinitionTest{
            input: []mg.TypeReference{
                mg.TypeInt32,
                mg.NewPointerTypeReference( mg.TypeInt32 ),
                mg.NewPointerTypeReference( mg.TypeInt64 ),
                mg.MustNullableTypeReference(
                    mg.NewPointerTypeReference( mg.TypeInt64 ) ),
                &mg.ListTypeReference{ mg.TypeInt32, true },
                &mg.ListTypeReference{ mg.TypeInt32, false },
                mg.NewPointerTypeReference(
                    &mg.ListTypeReference{ mg.TypeInt32, true } ),
                &mg.ListTypeReference{
                    mg.NewPointerTypeReference( mg.TypeInt32 ), 
                    true,
                },
            },
            errGroups: [][]int{
                []int{ 0, 1 },
                []int{ 2, 3 },
                []int{ 4, 5, 6, 7 },
            },
        },
    )
    return res
}

func TestUnionDefinition( t *testing.T ) {
    la := assert.NewListPathAsserter( t )
    for _, test := range getUnionDefinitionTests() {
        test.call( la )
        la = la.Next()
    }
}
