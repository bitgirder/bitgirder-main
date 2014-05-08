package types

import (
    mg "mingle"
    mgRct "mingle/reactor"
)

type CastReactorTest struct {
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

type ServiceMaps struct {
    Definitions *DefinitionMap
    ServiceIds *mg.IdentifierMap
}

type ServiceRequestTest struct {
    In mg.Value
    Maps *ServiceMaps 
    Parameters *mg.SymbolMap
    Authentication mg.Value
    Error error
}

type ServiceResponseTest struct {
    Definitions *DefinitionMap
    ServiceType *mg.QualifiedTypeName
    Operation *mg.Identifier
    In mg.Value
    ResultValue mg.Value
    ErrorValue mg.Value
    Error error
}
