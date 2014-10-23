package service

import (
    mg "mingle"
    "mingle/parser"
)

var (
    mkNs = parser.MustNamespace
    mkId = parser.MustIdentifier
    mkQn = parser.MustQualifiedTypeName
    asType = parser.AsTypeReference
)

type ResultExpectation struct {
    Result mg.Value
    Error mg.Value
}

type TckTestCall struct {
    Context *RequestContext
    Parameters *mg.SymbolMap
    Authentication mg.Value
    Expect *ResultExpectation
}
