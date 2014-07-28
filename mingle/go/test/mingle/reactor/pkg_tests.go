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

type testError struct { 
    path objpath.PathNode
    msg string 
}

func ( e *testError ) Error() string { return mg.FormatError( e.path, e.msg ) }

func newTestError( path objpath.PathNode, msg string ) *testError {
    return &testError{ path: path, msg: msg }
}

const (
    bindTestProfileDefault = "default"
    bindTestProfileError = "error"
)

type BuildReactorTest struct { 
    Val mg.Value 
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
