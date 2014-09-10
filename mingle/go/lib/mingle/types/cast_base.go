package types

import (
    mg "mingle"
    "bitgirder/objpath"
)

func newNullCastError( path objpath.PathNode ) *mg.CastError {
    return mg.NewCastError( path, "Value is null" )
}
