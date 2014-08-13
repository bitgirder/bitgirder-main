package builtin

import (
    "mingle/parser"
)

var (
    mkId = parser.MustIdentifier
    mkNs = parser.MustNamespace
    mkQn = parser.MustQualifiedTypeName
    asType = parser.AsTypeReference
    
    reactorTestNs = mkNs( "mingle:types:builtin@v1" )
)
