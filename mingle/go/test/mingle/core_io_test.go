package mingle

import (
    "testing"
    "bitgirder/assert"
    "bytes"
    "fmt"
)

func assertBinIoInvalidData( 
    idt *BinIoInvalidDataTest, 
    a *assert.PathAsserter,
) {
    a = a.Descend( idt.Name )
    rd := NewReader( bytes.NewBuffer( idt.Input ) )
    if val, err := rd.ReadValue(); err == nil {
        a.Fatalf( "Got val: %s", QuoteValue( val ) )
    } else {
        if ioe, ok := err.( *BinIoError ); ok {
            a.Equal( idt.ErrMsg, ioe.Error() )
        } else { a.Fatal( err ) }
    }
}

func assertBinIoRoundtripWrite(
    wr *BinWriter, obj interface{}, a *assert.PathAsserter ) {

    if err := WriteBinIoTestValue( obj, wr ); err != nil { a.Fatal( err ) }
}

func assertBinIoRoundtrip( rt *BinIoRoundtripTest, a *assert.PathAsserter ) {
    a = a.Descend( rt.Name )
    bb := &bytes.Buffer{}
    assertBinIoRoundtripWrite( NewWriter( bb ), rt.Val, a )
    rt.AssertWriteValue( NewReader( bb ), a )
}

func assertBinIoSequenceRoundtrip( 
    rt *BinIoSequenceRoundtripTest, a *assert.PathAsserter ) {

    a = a.Descend( rt.Name )
    bb := &bytes.Buffer{}
    wr := NewWriter( bb )
    la := a.StartList()
    for _, val := range rt.Seq {
        assertBinIoRoundtripWrite( wr, val, la )
        la = la.Next()
    }
    rt.AssertWriteValue( NewReader( bb ), a )
}

func TestCoreIo( t *testing.T ) {
    a := assert.NewPathAsserter( t )
    for _, test := range CreateCoreIoTests() { 
        ta := a.Descend( fmt.Sprintf( "%T", test ) )
        switch v := test.( type ) {
        case *BinIoInvalidDataTest: assertBinIoInvalidData( v, ta )
        case *BinIoRoundtripTest: assertBinIoRoundtrip( v, ta )
        case *BinIoSequenceRoundtripTest: assertBinIoSequenceRoundtrip( v, ta )
        default: a.Fatalf( "unhandled test type: %T", test )
        }
    }
}

func TestAsAndFromBytes( t *testing.T ) {
    a := assert.NewPathAsserter( t )
    qn := MustQualifiedTypeName( "ns1@v1/T1" )
    qn2, err := QualifiedTypeNameFromBytes( QualifiedTypeNameAsBytes( qn ) )
    if err == nil { a.True( qn.Equals( qn2 ) ) } else { a.Fatal( err ) }
    typ := MustTypeReference( "ns1@v1/L*" )
    typ2, err := TypeReferenceFromBytes( TypeReferenceAsBytes( typ ) ) 
    if err == nil { a.True( typ.Equals( typ2 ) ) } else { a.Fatal( err ) }
    p := idPathRootVal.Descend( id( "id1" ) )
    p2, err := IdPathFromBytes( IdPathAsBytes( p ) )
    if err == nil { 
        a.Equal( FormatIdPath( p ), FormatIdPath( p2 ) ) 
    } else { a.Fatal( err ) }
    id := MustIdentifier( "id1" )
    id2, err := IdentifierFromBytes( IdentifierAsBytes( id ) )
    if err == nil { a.True( id.Equals( id2 ) ) } else { a.Fatal( err ) }
    ns := MustNamespace( "ns1@v1" )
    ns2, err := NamespaceFromBytes( NamespaceAsBytes( ns ) )
    if err == nil { a.True( ns.Equals( ns2 ) ) } else { a.Fatal( err ) }
}
