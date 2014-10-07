package io

import (
    "testing"
    "bitgirder/assert"
    "bytes"
    mg "mingle"
)

func assertWriteValue( wr *BinWriter, val mg.Value, a *assert.PathAsserter ) {
    if err := wr.WriteValue( val ); err != nil {
        a.Fatalf( "couldn't write val: %s", err )
    }
}

func assertRoundtrip( rt *mg.BinIoRoundtripTest, a *assert.PathAsserter ) {
    val, ok := rt.Val.( mg.Value )
    if ! ok { return }
    bb := &bytes.Buffer{}
    assertWriteValue( NewWriter( bb ), val, a )
    AssertRoundtripRead( rt, NewReader( bb ), a )
}

func assertSequenceRoundtrip( 
    rt *mg.BinIoSequenceRoundtripTest, a *assert.PathAsserter ) {

    bb := &bytes.Buffer{}
    wr, rd := NewWriter( bb ), NewReader( bb )
    la := a.StartList()
    for _, val := range rt.Seq {
        assertWriteValue( wr, val, la )
        la = la.Next()
    }
    AssertSequenceRoundtripRead( rt, rd, a )
}

func assertInvalidData( t *mg.BinIoInvalidDataTest, a *assert.PathAsserter ) {
    rd := NewReader( bytes.NewBuffer( t.Input ) )
    var err error
    switch t.ReadType {
    case mg.BinIoInvalidDataTestReadTypeValue: _, err = rd.ReadValue()
    case mg.BinIoInvalidDataTestReadTypeAtomicType: 
        _, err = rd.ReadAtomicTypeReference()
    default: a.Fatalf( "unhandled read type: %s", t.ReadType )
    }
    if ioe, ok := err.( *mg.BinIoError ); ok {
        a.Equal( t.ErrMsg, ioe.Error() )
    } else { 
        if err == nil {
            a.Fatalf( "expected error: %s", t.ErrMsg )
        } else { a.Fatal( err ) }
    }
}

func TestIo( t *testing.T ) {
    a := assert.NewPathAsserter( t )
    for _, test := range mg.CreateCoreIoTests() {
        ta := a.Descend( mg.CoreIoTestNameFor( test ) )
        switch v := test.( type ) {
        case *mg.BinIoRoundtripTest: assertRoundtrip( v, ta )
        case *mg.BinIoSequenceRoundtripTest: assertSequenceRoundtrip( v, ta )
        case *mg.BinIoInvalidDataTest: assertInvalidData( v, ta )
        default: ta.Fatalf( "unimplemented: %T", test )
        }
    }
}
