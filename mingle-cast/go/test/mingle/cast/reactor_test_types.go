package cast

import (
    mg "mingle"
    "mingle/types"
    mgRct "mingle/reactor"
    "bitgirder/objpath"
)

const (
    ProfileCastDisable = "cast-disable"
    ProfileCustomFieldSet = "custom-field-set"
    ProfileUnionImpl = "union-impl"
    ProfileCustomErrorFormatting = "custom-error-formatting"
)

type CastReactorTest struct {
    Path objpath.PathNode
    Map *types.DefinitionMap
    Type mg.TypeReference
    In interface{}
    Expect mg.Value
    Err error
    Profile string
}

type EventPathTest struct {
    Source []mgRct.Event
    Expect []mgRct.EventExpectation
    Type mg.TypeReference
    Map *types.DefinitionMap
}
