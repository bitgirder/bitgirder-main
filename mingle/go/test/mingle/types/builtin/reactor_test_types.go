package builtin

import (
    mg "mingle"
    "mingle/types"
)

type BuiltinTypeTest struct {
    In mg.Value
    Expect interface{}
    Type mg.TypeReference
    Map *types.DefinitionMap
    Err error
}
