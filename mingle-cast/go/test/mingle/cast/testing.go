package cast

import (
    "mingle/parser"
)

var (
    mkQn = parser.MustQualifiedTypeName
    mkId = parser.MustIdentifier
    mkNs = parser.MustNamespace
    mkTyp = parser.MustTypeReference
    asType = parser.AsTypeReference
)
