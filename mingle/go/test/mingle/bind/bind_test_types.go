package bind

import (
    mg "mingle"
)

type BindTest struct {
    In mg.Value
    Expect interface{}
    Type mg.TypeReference
    Domain *mg.Identifier
    Error error
}

type S1 struct {
    f1 int32
}

type E1 string 

const (
    E1V1 = E1( "v1" )
    E1V2 = E1( "v2" )
)
