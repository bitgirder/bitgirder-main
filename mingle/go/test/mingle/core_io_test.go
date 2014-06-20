package mingle

import (
    "testing"
    "bitgirder/assert"
    "bitgirder/objpath"
    "bytes"
)

func writeBinIoTestValue( val interface{}, w *BinWriter ) error {
    switch v := val.( type ) {
    case Value: return w.WriteScalarValue( v )
    case *Identifier: return w.WriteIdentifier( v )
    case PointerId: return w.WritePointerId( v )
    case objpath.PathNode: return w.WriteIdPath( v )
    case *Namespace: return w.WriteNamespace( v )
    case TypeName: return w.WriteTypeName( v )
    case TypeReference: return w.WriteTypeReference( v )
    }
    panic( libErrorf( "Unhandled expct obj: %T", val ) )
}

func assertReadScalar( expct Value, rd *BinReader, a *assert.PathAsserter ) {
    if tc, err := rd.ReadTypeCode(); err == nil {
        if act, err := rd.ReadScalarValue( tc ); err == nil {
            AssertEqualValues( expct, act, a )
        } else {
            a.Fatalf( "couldn't read act: %s", err )
        }
    } else {
        a.Fatalf( "couldn't get type code: %s", err )
    }
}

func assertBinIoRoundtripRead(
    expct interface{}, rd *BinReader, a *assert.PathAsserter ) {

    switch v := expct.( type ) {
    case Value: assertReadScalar( v, rd, a )
    case *Identifier:
        if id, err := rd.ReadIdentifier(); err == nil { 
            a.True( v.Equals( id ) )
        } else { a.Fatal( err ) }
    case PointerId:
        if id, err := rd.ReadPointerId(); err == nil {
            a.Equal( v, id )
        } else { a.Fatal( err ) }
    case objpath.PathNode:
        if n, err := rd.ReadIdPath(); err == nil {
            a.Equal( v, n ) 
        } else { a.Fatal( err ) }
    case *Namespace:
        if ns, err := rd.ReadNamespace(); err == nil {
            a.True( v.Equals( ns ) )
        } else { a.Fatal( err ) }
    case TypeName:
        if nm, err := rd.ReadTypeName(); err == nil {
            a.True( v.Equals( nm ) )
        } else { a.Fatal( err ) }
    case TypeReference:
        if typ, err := rd.ReadTypeReference(); err == nil {
            a.Truef( v.Equals( typ ), "expct (%v) != act (%v)", v, typ )
        } else { a.Fatal( err ) }
    default: a.Fatalf( "Unhandled expct expct: %T", expct )
    }
}

func assertBinIoRoundtrip( val interface{}, a *assert.PathAsserter ) {
    bb := &bytes.Buffer{}
    if err := writeBinIoTestValue( val, NewWriter( bb ) ); err != nil { 
        a.Fatal( err ) 
    }
    assertBinIoRoundtripRead( val, NewReader( bb ), a )
}

func TestCoreIo( t *testing.T ) {
    a := assert.NewPathAsserter( t )
    for _, test := range CreateCoreIoTests() { 
        if rt, ok := test.( *BinIoRoundtripTest ); ok {
            switch v := rt.Val.( type ) {
            case *Null, Boolean, Buffer, String, *Enum, Int32, Uint32, Int64,
                 Uint64, Float32, Float64, Timestamp, *Identifier, *Namespace,
                 *QualifiedTypeName, TypeReference, *DeclaredTypeName, 
                 PointerId, objpath.PathNode:
                assertBinIoRoundtrip( v, a.Descend( rt.Name ) )
            case *SymbolMap, *HeapValue, *List, *Struct: ; // okay but skip
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
        ElementType: &AtomicTypeReference{ Name: mkQn( ns, mkDeclNm( "L" ) ) },
    }
    typ2, err := TypeReferenceFromBytes( TypeReferenceAsBytes( typ ) ) 
    if err == nil { a.True( typ.Equals( typ2 ) ) } else { a.Fatal( err ) }
    p := idPathRootVal.Descend( mkId( "id1" ) )
    p2, err := IdPathFromBytes( IdPathAsBytes( p ) )
    if err == nil { 
        a.Equal( FormatIdPath( p ), FormatIdPath( p2 ) ) 
    } else { a.Fatal( err ) }
    id := mkId( "id1" )
    id2, err := IdentifierFromBytes( IdentifierAsBytes( id ) )
    if err == nil { a.True( id.Equals( id2 ) ) } else { a.Fatal( err ) }
    ns2, err := NamespaceFromBytes( NamespaceAsBytes( ns ) )
    if err == nil { a.True( ns.Equals( ns2 ) ) } else { a.Fatal( err ) }
}
