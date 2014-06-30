package types

import (
    mg "mingle"
    "bitgirder/objpath"
)

func newNullValueCastError( path objpath.PathNode ) *mg.ValueCastError {
    return mg.NewValueCastError( path, "Value is null" )
}
