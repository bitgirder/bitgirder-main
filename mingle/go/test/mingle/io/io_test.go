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
    AssertRoundtripReadValue( rt, NewReader( bb ), a )
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
    la = a.StartList()
    for _, val := range rt.Seq {
        assertReadValue( rd, val, la )
        la = la.Next()
    }
}

func assertInvalidData( t *mg.BinIoInvalidDataTest, a *assert.PathAsserter ) {
    rd := NewReader( bytes.NewBuffer( t.Input ) )
    if val, err := rd.ReadValue(); err == nil {
        a.Fatalf( "expected %s (%T) but got val: %s", err, err, 
            mg.QuoteValue( val ) )
    } else { 
        if ioe, ok := err.( *mg.BinIoError ); ok {
            a.Equal( t.ErrMsg, ioe.Error() )
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
