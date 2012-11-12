package code

import(
    mg "mingle"
)

type Expression interface {}

type Boolean bool
type Int32 int32
type Int64 int64
type Uint32 uint32
type Uint64 uint64
type Float32 float32
type Float64 float64
type String string
type EnumValue struct { Value *mg.Enum }
type Timestamp struct { Value mg.Timestamp }

type ListValue struct { Values []Expression }

func NewListValue() *ListValue { return &ListValue{ []Expression{} } }

type IdentifierReference struct { Id *mg.Identifier }

type Negation struct { Exp Expression }
