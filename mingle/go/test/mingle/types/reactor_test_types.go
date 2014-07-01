package types

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "bitgirder/objpath"
)

type CastReactorTest struct {
    Path objpath.PathNode
    Map *DefinitionMap
    Type mg.TypeReference
    In mg.Value
    Expect mg.Value
    Err error
}

type EventPathTest struct {
    Source []mgRct.ReactorEvent
    Expect []mgRct.EventExpectation
    Type mg.TypeReference
    Map *DefinitionMap
}
