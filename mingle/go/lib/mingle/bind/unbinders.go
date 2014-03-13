package bind

import (
    mg "mingle"
    "bitgirder/objpath"
)

type valUnbinderFunc func( 
    val mg.Value, path objpath.PathNode ) ( interface{}, error )

func ( f valUnbinderFunc ) UnbindValue(
    val mg.Value, path objpath.PathNode ) ( interface{}, error ) {

    return f( val, path )
}

var Int32ValueUnbinder ValueUnbinder

func init() {

    Int32ValueUnbinder = valUnbinderFunc(
        func( val mg.Value, path objpath.PathNode ) ( interface{}, error ) {
            if i32, ok := val.( mg.Int32 ); ok { return int32( i32 ), nil }
            return nil, mg.NewTypeCastErrorValue( mg.TypeInt32, val, path )
        },
    )
}
