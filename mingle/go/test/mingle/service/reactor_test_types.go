package service

import (
    mg "mingle"
)

type ServiceReactorBaseTest struct {
    Expect interface{}
    Error error
    In mg.Value
}

type requestImpl struct {
    ctx *RequestContext
    params mg.Value
    auth mg.Value
}
