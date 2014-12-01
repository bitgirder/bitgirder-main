package golang

import (
    "mingle/types"
    "bitgirder/stub"
)

type Generator struct {
    Definitions *types.DefinitionMap
}

func NewGenerator() *Generator { return &Generator{} }

func ( g *Generator ) Generate() error {
    return stub.Unimplemented()
}
