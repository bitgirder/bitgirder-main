package mingle

import (
    "bitgirder/assert"
    "bitgirder/objpath"
)

func EqualValues( expct, act Value, a assert.Failer ) {
    ( &assert.Asserter{ a } ).Equalf( expct, act,
        "expected %s but got %s", QuoteValue( expct ), QuoteValue( act ) )
}

func EqualPaths( expct, act objpath.PathNode, a assert.Failer ) {
    ( &assert.Asserter{ a } ).Equalf( 
        expct, 
        act,
        "expected path %q but got %q", FormatIdPath( expct ),
            FormatIdPath( act ),
    )
}

func typeRef( s string ) TypeReference { return MustTypeReference( s ) }

var qname = MustQualifiedTypeName

func atomicRef( s string ) *AtomicTypeReference {
    return typeRef( s ).( *AtomicTypeReference )
}

func id( s string ) *Identifier { return MustIdentifier( s ) }
