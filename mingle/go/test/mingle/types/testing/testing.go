package testing

import (
    mg "mingle"
    "mingle/parser"
)

var (
    newVcErr = mg.NewInputError
    mkQn = parser.MustQualifiedTypeName
    mkId = parser.MustIdentifier
    mkNs = parser.MustNamespace
    mkTyp = parser.MustTypeReference
    asType = parser.AsTypeReference
)
