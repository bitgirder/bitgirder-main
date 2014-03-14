package mingle

import (
    "bitgirder/assert"
    "bitgirder/objpath"
)

func EqualValues( expct, act Value, f assert.Failer ) {
    a := &assert.Asserter{ f } 
    if tm, tmOk := expct.( Timestamp ); tmOk {
        a.Truef( tm.Compare( act ) == 0, "input time was %s, got: %s", tm, act )
    } else { 
        a.Equalf( expct, act, 
            "expected %s but got %s", QuoteValue( expct ), QuoteValue( act ) )
    }
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

var id = MustIdentifier
