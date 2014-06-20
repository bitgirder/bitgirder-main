package types

import (
    mg "mingle"
    mgRct "mingle/reactor"
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

func TestReactors( t *testing.T ) {
    mgRct.RunReactorTestsInNamespace( reactorTestNs, t )
}
