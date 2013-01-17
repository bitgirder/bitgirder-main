package mingle

func typeRef( s string ) TypeReference { return MustTypeReference( s ) }

var qname = MustQualifiedTypeName

func atomicRef( s string ) *AtomicTypeReference {
    return typeRef( s ).( *AtomicTypeReference )
}

func id( s string ) *Identifier { return MustIdentifier( s ) }
