package mingle

import (
    "testing"
    "bitgirder/objpath"
    "bitgirder/assert"
    "bytes"
)

func TestCoreIo( t *testing.T ) {
    la := assert.NewPathAsserter( t ).StartList()
    id := MustIdentifier
    p1 := objpath.RootedAt( id( "id1" ) )
    p2 := p1.Descend( id( "id2" ) )
    p3 := p2.StartList().Next().Next()
    p4 := p3.Descend( id( "id3" ) )
    p5 := objpath.RootedAtList().Descend( id( "id1" ) )
    for _, obj := range []interface{} {
        NullVal,
        String( "hello" ),
        Boolean( true ),
        Boolean( false ),
        Buffer( []byte{} ),
        Buffer( []byte( "hello" ) ),
        Int32( int32( 1 ) ),
        Int64( int64( 1 ) ),
        Uint32( uint32( 1 ) ),
        Uint64( uint64( 1 ) ),
        Float32( float32( 1 ) ),
        Float64( float64( 1 ) ),
        Now(),
        MustEnum( "ns1@v1/E1", "val1" ),
        MustSymbolMap(),
        MustSymbolMap( "k1", int32( 1 ) ),
        MustStruct( "ns1@v1/T1" ),
        MustStruct( "ns1@v1/T1", "k1", int32( 1 ) ),
        MustList(),
        MustList( int32( 1 ), "hello" ),
        id( "id1" ),
        id( "id1-id2" ),
        p1,
        p2,
        p3,
        p4,
        p5,
        MustNamespace( "ns1@v1" ),
        MustNamespace( "ns1:ns2@v1" ),
        MustDeclaredTypeName( "T1" ),
        MustQualifiedTypeName( "ns1:ns2@v1/T1" ),
        MustTypeReference( "T1" ),
        MustTypeReference( `String~"a"` ),
        MustTypeReference( `String~["a","b"]` ),
        MustTypeReference( 
            `Timestamp~["2012-01-01T00:00:00Z","2012-02-01T00:00:00Z"]` ),
        MustTypeReference( "Int32~(0,10)" ),
        MustTypeReference( "Int64~[0,10]" ),
        MustTypeReference( "Uint32~(0,10)" ),
        MustTypeReference( "Uint64~[0,10]" ),
        MustTypeReference( "Float32~(0.0,1.0]" ),
        MustTypeReference( "Float64~[0.0,1.0)" ),
        MustTypeReference( "Float64~(,)" ),
        MustTypeReference( "T1*" ),
        MustTypeReference( "T1+" ),
        MustTypeReference( "T1*?" ),
        MustTypeReference( "ns1@v1/T1" ),
        MustTypeReference( "ns1@v1/T1*" ),
        MustTypeReference( "ns1@v1/T1?" ),
    } {
        bb := &bytes.Buffer{}
        rd := NewReader( bb )
        wr := NewWriter( bb )
        var err error
        switch v := obj.( type ) {
        case Value: err = wr.WriteValue( v )
        case *Identifier: err = wr.WriteIdentifier( v )
        case objpath.PathNode: err = wr.WriteIdPath( v )
        case *Namespace: err = wr.WriteNamespace( v )
        case TypeName: err = wr.WriteTypeName( v )
        case TypeReference: err = wr.WriteTypeReference( v )
        default: t.Fatalf( "Unhandled expct obj: %T", obj )
        }
        if err != nil { la.Fatal( err ) }
        trailExpct := "this-should-be-left-after-read"
        bb.WriteString( trailExpct )
        switch v := obj.( type ) {
        case Value:
            if val, err := rd.ReadValue(); err == nil {
                obj2 := obj
                if obj == nil { obj2 = NullVal }
                la.Equal( obj2, val )
            } else { la.Fatal( err ) }
        case *Identifier:
            if id, err := rd.ReadIdentifier(); err == nil { 
                la.True( v.Equals( id ) )
            } else { la.Fatal( err ) }
        case objpath.PathNode:
            if n, err := rd.ReadIdPath(); err == nil {
                la.Equal( v, n ) 
            } else { la.Fatal( err ) }
        case *Namespace:
            if ns, err := rd.ReadNamespace(); err == nil {
                la.True( v.Equals( ns ) )
            } else { la.Fatal( err ) }
        case TypeName:
            if nm, err := rd.ReadTypeName(); err == nil {
                la.True( v.Equals( nm ) )
            } else { la.Fatal( err ) }
        case TypeReference:
            if typ, err := rd.ReadTypeReference(); err == nil {
                la.Truef( v.Equals( typ ), "expct (%v) != act (%v)", v, typ )
            } else { la.Fatal( err ) }
        default: t.Fatalf( "Unhandled expct obj: %T", obj )
        }
        la.Equal( trailExpct, bb.String() )
        la = la.Next()
    }
}

func TestBinReaderBadInputs( t *testing.T ) {
    la := assert.NewListPathAsserter( t )
    for _, fi := range BinWriterFailureInputs {
        rd := NewReader( bytes.NewBuffer( fi.Input ) )
        if val, err := rd.ReadValue(); err == nil {
            la.Fatalf( "Got val: %s", QuoteValue( val ) )
        } else {
            if ioe, ok := err.( *BinIoError ); ok {
                la.Equal( fi.ErrMsg, ioe.Error() )
            } else { la.Fatal( err ) }
        }
        la = la.Next()
    }
}
