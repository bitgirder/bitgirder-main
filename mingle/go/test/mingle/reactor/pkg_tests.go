package reactor

import (
    mg "mingle"
    "bitgirder/objpath"
)

type ValueBuildTest struct { 
    Val mg.Value 
    Source []ReactorEvent
}

type EventExpectation struct {
    Event ReactorEvent
    Path objpath.PathNode
}

type EventPathTest struct {
    Name string
    Events []EventExpectation
    StartPath objpath.PathNode
}

func ( ept EventPathTest ) TestName() string { return ept.Name }

type StructuralReactorErrorTest struct {
    Events []ReactorEvent
    Error *ReactorError
    TopType ReactorTopType
}

type PointerEventCheckTest struct {
    Events []ReactorEvent
    Error error // if nil then Events should be fed through without error
}

type FieldOrderReactorTestOrder struct {
    Order FieldOrder
    Type *mg.QualifiedTypeName
}

type FieldOrderReactorTest struct {
    Source []ReactorEvent
    Expect mg.Value
    Orders []FieldOrderReactorTestOrder
}

type FieldOrderMissingFieldsTest struct {
    Orders []FieldOrderReactorTestOrder
    Source []ReactorEvent
    Expect mg.Value
    Error *mg.MissingFieldsError
}

type FieldOrderPathTest struct {
    Source []ReactorEvent
    Expect []EventExpectation
    Orders []FieldOrderReactorTestOrder
}

type RequestReactorTest struct {
    Source interface{}
    Namespace *mg.Namespace
    Service *mg.Identifier
    Operation *mg.Identifier
    Parameters *mg.SymbolMap
    ParameterEvents []EventExpectation
    Authentication mg.Value
    AuthenticationEvents []EventExpectation
    Error error
}
