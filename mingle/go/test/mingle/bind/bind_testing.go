package bind

import (
    mg "mingle"
)

type RoundtripTest struct {
    Value mg.Value
    Object interface{}
    Type mg.TypeReference
}

func rtTestsAddPrimitives( res []interface{} ) []interface{} {
    goIntList := func( sz int ) []int32 {
        res := make( []int32, sz )
        for i := 0; i < sz; i++ { res[ i ] = int32( i ) }
        return res
    }
    mgIntList := func( sz int ) *mg.List {
        res := make( []mg.Value, sz )
        for i := 0; i < sz; i++ { res[ i ] = mg.Int32( int32( i ) ) }
        return mg.NewList( res )
    }
    return append( res,
        &RoundtripTest{
            Value: mg.Int32( 1 ),
            Object: int32( 1 ),
            Type: mg.TypeInt32,
        },
        &RoundtripTest{
            Value: mgIntList( 3 ),
            Object: goIntList( 3 ),
            Type: mg.MustTypeReference( "Int32*" ),
        },
        &RoundtripTest{
            Value: mgIntList( 0 ),
            Object: goIntList( 0 ),
            Type: mg.MustTypeReference( "Int32*" ),
        },
        &RoundtripTest{
            Value: mg.MustList(
                mgIntList( 0 ),
                mgIntList( 1 ),
                mgIntList( 2 ),
            ),
            Object: [][]int32{ goIntList( 0 ), goIntList( 1 ), goIntList( 2 ) },
            Type: mg.MustTypeReference( "Int32*+" ),
        },
        &RoundtripTest{
            Value: mg.MustList( mgIntList( 0 ), mg.NullVal, mgIntList( 2 ) ),
            Object: [][]int32{ goIntList( 0 ), nil, goIntList( 2 ) },
            Type: mg.MustTypeReference( "Int32*?*" ),
        },
    )
}

func StandardBindTests() []interface{} {
    res := []interface{}{}
    res = rtTestsAddPrimitives( res )
    return res
}
