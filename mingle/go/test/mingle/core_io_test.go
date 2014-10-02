package mingle

import (
    "testing"
    "bitgirder/assert"
    "bytes"
)

func writeBinIoTestValue( val interface{}, w *BinWriter ) error {
    switch v := val.( type ) {
    case Value: return w.WriteScalarValue( v )
    case *Identifier: return w.WriteIdentifier( v )
    case *Namespace: return w.WriteNamespace( v )
    case TypeName: return w.WriteTypeName( v )
    case TypeReference: return w.WriteTypeReference( v )
    }
    panic( libErrorf( "Unhandled expct obj: %T", val ) )
}

func assertBinIoRoundtrip( val interface{}, a *assert.PathAsserter ) {
    bb := &bytes.Buffer{}
    if err := writeBinIoTestValue( val, NewWriter( bb ) ); err != nil { 
        a.Fatal( err ) 
    }
    AssertBinIoRoundtripRead( NewReader( bb ), val, a )
}

func TestCoreIo( t *testing.T ) {
    a := assert.NewPathAsserter( t )
    for _, test := range CreateCoreIoTests() { 
        if rt, ok := test.( *BinIoRoundtripTest ); ok {
            switch v := rt.Val.( type ) {
            case *Null, Boolean, Buffer, String, *Enum, Int32, Uint32, Int64,
                 Uint64, Float32, Float64, Timestamp, *Identifier, *Namespace,
                 *QualifiedTypeName, TypeReference, *DeclaredTypeName:
                assertBinIoRoundtrip( v, a.Descend( rt.Name ) )
            case *SymbolMap, *List, *Struct: ; // okay but skip
            default: a.Fatalf( "unhandled rt val: %T", rt.Val )
            }
        }
    }
}

func TestAsAndFromBytes( t *testing.T ) {
    a := assert.NewPathAsserter( t )
    ns := mkNs( mkId( "v1" ), mkId( "ns1" ) )
    qn := mkQn( ns, mkDeclNm( "T1" ) )
    qn2, err := QualifiedTypeNameFromBytes( QualifiedTypeNameAsBytes( qn ) )
    if err == nil { a.True( qn.Equals( qn2 ) ) } else { a.Fatal( err ) }
    typ := &ListTypeReference{
        AllowsEmpty: true,
        ElementType: NewAtomicTypeReference( mkQn( ns, mkDeclNm( "L" ) ), nil ),
    }
    typ2, err := TypeReferenceFromBytes( TypeReferenceAsBytes( typ ) ) 
    if err == nil { a.True( typ.Equals( typ2 ) ) } else { a.Fatal( err ) }
    id := mkId( "id1" )
    id2, err := IdentifierFromBytes( IdentifierAsBytes( id ) )
    if err == nil { a.True( id.Equals( id2 ) ) } else { a.Fatal( err ) }
    ns2, err := NamespaceFromBytes( NamespaceAsBytes( ns ) )
    if err == nil { a.True( ns.Equals( ns2 ) ) } else { a.Fatal( err ) }
}
