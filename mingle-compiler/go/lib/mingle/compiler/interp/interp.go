package interp

import(
    "fmt"
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

type UnboundIdentifierError struct { Id *mg.Identifier }

func ( e *UnboundIdentifierError ) Error() string {
    return fmt.Sprintf( "Unbound identifier: %s", e.Id )
}

type EvaluationError struct { Err error }

func ( e *EvaluationError ) Error() string { return e.Err.Error() }

type context struct {}

func ( ctx *context ) failEvalWith( err error ) error {
    return &EvaluationError{ err }
}

func ( ctx *context ) failEvalf( tmpl string, argv ...interface{} ) error {
    return ctx.failEvalWith( &EvaluationError{ fmt.Errorf( tmpl, argv... ) } )
}

func evalListVal( lv *ListValue, ctx *context ) ( *mg.List, error ) {
    res := mg.NewList( mg.TypeOpaqueList )
    for _, eltVal := range lv.Values {
        if evRes, err := evaluate( eltVal, ctx ); err == nil {
            res.AddUnsafe( evRes )
        } else { return nil, err }
    }
    return res, nil
}

func evalIdRef( 
    idRef *IdentifierReference, ctx *context ) ( mg.Value, error ) {
    return nil, ctx.failEvalWith( &UnboundIdentifierError{ idRef.Id } )
}

func negate( neg *Negation, ctx *context ) ( mg.Value, error ) {
    val, err := evaluate( neg.Exp, ctx )
    if err != nil { return nil, err }
    switch v := val.( type ) {
    case mg.Int32: return mg.Int32( -int32( v ) ), nil
    case mg.Int64: return mg.Int64( -int64( v ) ), nil
    case mg.Float32: return mg.Float32( -float32( v ) ), nil
    case mg.Float64: return mg.Float64( -float64( v ) ), nil
    }
    return nil, ctx.failEvalf( "Attempt to negate value of type %T", val )
}

func evaluate( exp Expression, ctx *context ) ( mg.Value, error ) {
    switch v := exp.( type ) {
    case Boolean: return mg.Boolean( bool( v ) ), nil
    case Int32: return mg.Int32( int32( v ) ), nil
    case Int64: return mg.Int64( int64( v ) ), nil
    case Uint32: return mg.Uint32( uint32( v ) ), nil
    case Uint64: return mg.Uint64( uint64( v ) ), nil
    case Float32: return mg.Float32( float32( v ) ), nil
    case Float64: return mg.Float64( float64( v ) ), nil
    case String: return mg.String( string( v ) ), nil
    case *EnumValue: return v.Value, nil
    case *Timestamp: return v.Value, nil
    case *ListValue: return evalListVal( v, ctx )
    case *IdentifierReference: return evalIdRef( v, ctx )
    case *Negation: return negate( v, ctx )
    }
    panic( fmt.Sprintf( "Unexpected expression: %T", exp ) )
}

func Evaluate( exp Expression ) ( mg.Value, error ) {
    ctx := &context{}
    return evaluate( exp, ctx )
}
