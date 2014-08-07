package service

import (
    "mingle/parser"
)

var (
    mkNs = parser.MustNamespace
    mkId = parser.MustIdentifier
    mkQn = parser.MustQualifiedTypeName
    asType = parser.AsTypeReference
)
