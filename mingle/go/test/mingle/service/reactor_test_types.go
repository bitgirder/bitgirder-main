package service

import (
    mg "mingle"
    "bitgirder/objpath"
)

const (
    ProfileBase = "base"
    ProfileImplError = "impl-error"
)

type ReactorTest struct {
    Type *mg.QualifiedTypeName
    Expect interface{}
    Error error
    In mg.Value
    Profile string
}

type requestExpect struct {
    ctx *RequestContext
    params mg.Value
    auth mg.Value
}

type responseExpect struct {
    result mg.Value
    err mg.Value
}

type testError struct {
    path objpath.PathNode
    msg string
}

func ( t *testError ) Error() string { return mg.FormatError( t.path, t.msg ) }
