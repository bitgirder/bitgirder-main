package interpreter

import (
    "fmt"
    mg "mingle"
    "mingle/code"
)

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

func evalListVal( lv *code.ListValue, ctx *context ) ( *mg.List, error ) {
    res := mg.MakeList( len( lv.Values ) )
    for _, eltVal := range lv.Values {
        if evRes, err := evaluate( eltVal, ctx ); err == nil {
            res.Add( evRes )
        } else { return nil, err }
    }
    return res, nil
}

func evalIdRef( 
    idRef *code.IdentifierReference, ctx *context ) ( mg.Value, error ) {
    return nil, ctx.failEvalWith( &UnboundIdentifierError{ idRef.Id } )
}

func negate( neg *code.Negation, ctx *context ) ( mg.Value, error ) {
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

func evaluate( exp code.Expression, ctx *context ) ( mg.Value, error ) {
    switch v := exp.( type ) {
    case code.Boolean: return mg.Boolean( bool( v ) ), nil
    case code.Int32: return mg.Int32( int32( v ) ), nil
    case code.Int64: return mg.Int64( int64( v ) ), nil
    case code.Uint32: return mg.Uint32( uint32( v ) ), nil
    case code.Uint64: return mg.Uint64( uint64( v ) ), nil
    case code.Float32: return mg.Float32( float32( v ) ), nil
    case code.Float64: return mg.Float64( float64( v ) ), nil
    case code.String: return mg.String( string( v ) ), nil
    case *code.EnumValue: return v.Value, nil
    case *code.Timestamp: return v.Value, nil
    case *code.ListValue: return evalListVal( v, ctx )
    case *code.IdentifierReference: return evalIdRef( v, ctx )
    case *code.Negation: return negate( v, ctx )
    }
    panic( fmt.Sprintf( "Unexpected expression: %T", exp ) )
}

func Evaluate( exp code.Expression ) ( mg.Value, error ) {
    ctx := &context{}
    return evaluate( exp, ctx )
}
