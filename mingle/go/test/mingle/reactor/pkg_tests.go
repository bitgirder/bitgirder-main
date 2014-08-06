package reactor

import (
    mg "mingle"
    "mingle/parser"
    "bitgirder/objpath"
)

const testMsgErrorBadValue = "test-message-error-bad-value"
const buildReactorErrorTestVal = mg.Int32( int32( 100 ) )
var buildReactorErrorTestQn = parser.MustQualifiedTypeName( "ns1@v1/BadType" )
var buildReactorErrorTestField = parser.MustIdentifier( "bad-field" )

const (
    builderTestProfileDefault = "default"
    builderTestProfileError = "error"
    builderTestProfileImpl = "impl"
    builderTestProfileImplFailOnly = "impl-fail-only"
)

type BuildReactorTest struct { 
    Val interface{} 
    Source interface{}
    Profile string
    Error error
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

type DepthTrackerTest struct {
    Source []ReactorEvent
    Expect []int
}
