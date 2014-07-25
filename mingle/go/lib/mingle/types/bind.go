package types

import (
    mgRct "mingle/reactor"
    "mingle/bind"
    "mingle/parser"
    mg "mingle"
//    "bitgirder/stub"
)

// if err is something that should be sent to caller as a value error, a value
// error is returned; otherwise err is returned unchanged
func asValueError( ve mgRct.ReactorEvent, err error ) error {
    switch v := err.( type ) {
    case *parser.ParseError:
        err = mg.NewValueCastError( ve.GetPath(), v.Error() )
    case *mg.BinIoError: err = mg.NewValueCastError( ve.GetPath(), v.Error() )
    }
    return err
}

var IdentifierBinderFactory = &bind.FunctionBinderFactory{
}
