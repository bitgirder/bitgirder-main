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
