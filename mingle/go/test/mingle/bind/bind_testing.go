package bind

import (
    mg "mingle"
    "bitgirder/objpath"
)

type RoundtripTest struct {
    Value mg.Value
    Object interface{}
    Type mg.TypeReference
}

type UnbindErrorTest struct {
    Source []mg.ReactorEvent
    Error error
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
        &UnbindErrorTest{
            Source: []mg.ReactorEvent{ mg.NewValueEvent( mg.Int64( 1 ) ), },
            Type: mg.TypeInt32,
            Error: mg.NewTypeCastError( mg.TypeInt32, mg.TypeInt64, nil ),
        },
        &UnbindErrorTest{
            Source: []mg.ReactorEvent{ mg.NewValueEvent( mg.Int32( 1 ) ), },
            Type: mg.MustTypeReference( "Int32*" ),
            Error: NewUnbindError( nil, 
                `expected list start but got [ type = *mingle.ValueEvent, value = 1 ]`,
            ),
        },
        &UnbindErrorTest{
            Source: []mg.ReactorEvent{ mg.NewValueEvent( mg.NullVal ) },
            Type: mg.MustTypeReference( "Int32*" ),
            Error: NewUnbindError( nil, 
                `expected list start but got [ type = *mingle.ValueEvent, value = null ]`,
            ),
        },
        &UnbindErrorTest{
            Source: []mg.ReactorEvent{ mg.NewListStartEvent() },
            Type: mg.TypeInt32,
            Error: NewUnbindError( nil,
                `unexpected event [ type = *mingle.ListStartEvent ] for value unbind`,
            ),
        },
        &UnbindErrorTest{
            Source: []mg.ReactorEvent{ mg.NewValueEvent( mg.NullVal ) },
            Type: mg.TypeInt32,
            Error: mg.NewTypeCastError( mg.TypeInt32, mg.TypeNull, nil ),
        },
        &UnbindErrorTest{
            Source: []mg.ReactorEvent{
                mg.NewListStartEvent(),
                mg.NewListStartEvent(),
                mg.NewEndEvent(),
                mg.NewValueEvent( mg.Int32( 1 ) ),
            },
            Type: mg.MustTypeReference( "Int32**" ),
            Error: NewUnbindError(
                objpath.RootedAtList().SetIndex( 1 ),
                `expected list start but got [ type = *mingle.ValueEvent, value = 1, path = [ 1 ] ]`,
            ),
        },
    )
}

func StandardBindTests() []interface{} {
    res := []interface{}{}
    res = rtTestsAddPrimitives( res )
    return res
}
