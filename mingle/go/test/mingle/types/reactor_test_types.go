package types

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "bitgirder/objpath"
)

const (
    ProfileCastDisabled = "cast-disabled"
)

type CastReactorTest struct {
    Path objpath.PathNode
    Map *DefinitionMap
    Type mg.TypeReference
    In mg.Value
    Expect mg.Value
    Err error
    Profile string
}

type EventPathTest struct {
    Source []mgRct.ReactorEvent
    Expect []mgRct.EventExpectation
    Type mg.TypeReference
    Map *DefinitionMap
}

type BuiltinTypeTest struct {
    In mg.Value
    Expect interface{}
    Type mg.TypeReference
    Map *DefinitionMap
    Err error
}
