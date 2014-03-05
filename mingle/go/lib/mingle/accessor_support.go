package mingle

import (
    "bitgirder/objpath"
)

func accSupportCast( 
    val Value, typ TypeReference, path objpath.PathNode ) ( Value, error ) {

    if at, ok := typ.( *AtomicTypeReference ); ok {
        return CastAtomic( val, at, path )
    }
    if TypeOpaqueList.Equals( typ ) {
        if _, ok := val.( *List ); ok { return val, nil }
    }
    return nil, NewTypeCastErrorValue( typ, val, path )
}
