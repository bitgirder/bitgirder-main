package testing

import (
    mg "mingle"
    "mingle/types"
    mgRct "mingle/reactor"
    "mingle/parser"
    "bitgirder/objpath"
)

var reactorTestNs = parser.MustNamespace( "mingle:types:testing@v1" )

const (
    ProfileCastDisable = "cast-disable"
    ProfileCustomFieldSet = "custom-field-set"
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
    Source []mgRct.ReactorEvent
    Expect []mgRct.EventExpectation
    Type mg.TypeReference
    Map *types.DefinitionMap
}

type BuiltinTypeTest struct {
    In mg.Value
    Expect interface{}
    Type mg.TypeReference
    Map *types.DefinitionMap
    Err error
}
