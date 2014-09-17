package builtin

import (
    "mingle/parser"
    "mingle/types"
)

var (
    mkId = parser.MustIdentifier
    mkNs = parser.MustNamespace
    mkQn = parser.MustQualifiedTypeName
    asType = parser.AsTypeReference
    
    reactorTestNs = mkNs( "mingle:types:builtin@v1" )
)

func MakeDefMap( defs ...types.Definition ) *types.DefinitionMap {
    res := BuiltinTypes()
    res.MustAddAll( defs... )
    return res
}
