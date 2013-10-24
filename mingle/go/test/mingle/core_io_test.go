package mingle

import (
    "testing"
    "bitgirder/objpath"
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
    wr *BinWriter,
    obj interface{}, 
    a *assert.PathAsserter,
) {
    var err error
    switch v := obj.( type ) {
    case Value: err = wr.WriteValue( v )
    case *Identifier: err = wr.WriteIdentifier( v )
    case objpath.PathNode: err = wr.WriteIdPath( v )
    case *Namespace: err = wr.WriteNamespace( v )
    case TypeName: err = wr.WriteTypeName( v )
    case TypeReference: err = wr.WriteTypeReference( v )
    default: a.Fatalf( "Unhandled expct obj: %T", obj )
    }
    if err != nil { a.Fatal( err ) }
}

func assertBinIoRoundtripReadValue(
    rd *BinReader,
    expct interface{},
    a *assert.PathAsserter,
) {
    if val, err := rd.ReadValue(); err == nil {
        if expct == nil { val = NullVal }
        if tm, tmOk := expct.( Timestamp ); tmOk {
            a.Truef( tm.Compare( val ) == 0, "input time was %s, got: %s",
                tm, val )
        } else { a.Equal( expct, val ) }
    } else { a.Fatal( err ) }
}

func assertBinIoRoundtripRead(
    rd *BinReader,
    expct interface{},
    a *assert.PathAsserter,
) {
    switch v := expct.( type ) {
    case Value: assertBinIoRoundtripReadValue( rd, expct, a )
    case *Identifier:
        if id, err := rd.ReadIdentifier(); err == nil { 
            a.True( v.Equals( id ) )
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

func assertBinIoRoundtrip( rt *BinIoRoundtripTest, a *assert.PathAsserter ) {
    a = a.Descend( rt.Name )
    bb := &bytes.Buffer{}
    assertBinIoRoundtripWrite( NewWriter( bb ), rt.Val, a )
    assertBinIoRoundtripRead( NewReader( bb ), rt.Val, a )
}

func assertBinIoSequenceRoundtrip( 
    rt *BinIoSequenceRoundtripTest,
    a *assert.PathAsserter,
) {
    a = a.Descend( rt.Name )
    bb := &bytes.Buffer{}
    wr := NewWriter( bb )
    la := a.StartList()
    for _, val := range rt.Seq {
        assertBinIoRoundtripWrite( wr, val, la )
        la = la.Next()
    }
    la = a.StartList()
    rd := NewReader( bb )
    for _, val := range rt.Seq {
        assertBinIoRoundtripRead( rd, val, la )
        la = la.Next()
    }
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
