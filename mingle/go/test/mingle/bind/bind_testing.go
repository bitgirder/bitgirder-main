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
    return append( res,
        &RoundtripTest{
            Value: mg.Int32( 1 ),
            Object: int32( 1 ),
            Type: mg.TypeInt32,
        },
    )
}

func StandardBindTests() []interface{} {
    res := []interface{}{}
    res = rtTestsAddPrimitives( res )
    return res
}
