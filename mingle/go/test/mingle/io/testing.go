package io

import (
    mg "mingle"
    "io"
    "bitgirder/assert"
)

func WriteBinIoTestValue( obj interface{}, w io.Writer ) error {
    mw := NewWriter( w )
    switch v := obj.( type ) {
    case *mg.Identifier: return mw.WriteIdentifier( v )
    case *mg.Namespace: return mw.WriteNamespace( v )
    case mg.TypeName: return mw.WriteTypeName( v )
    case mg.TypeReference: return mw.WriteTypeReference( v )
    case mg.Value: return mw.WriteValue( v )
    }
    panic( libErrorf( "unhandled value: %T", obj ) )
}

func assertReadValue( 
    rd *BinReader, expct interface{}, a *assert.PathAsserter ) {

    fail := func( err error ) { a.Fatalf( "read val failed: %s", err ) }
    switch v := expct.( type ) {
    case mg.Value: 
        if act, err := rd.ReadValue(); err == nil {
            mg.AssertEqualValues( v, act, a )
        } else { fail( err ) }
    default: mg.AssertBinIoRoundtripRead( rd.BinReader, expct, a )
    }
}

func AssertRoundtripRead(
    rt *mg.BinIoRoundtripTest, rd *BinReader, a *assert.PathAsserter ) {
    
    assertReadValue( rd, rt.Val, a )
}

func AssertSequenceRoundtripRead(
    rt *mg.BinIoSequenceRoundtripTest, rd *BinReader, a *assert.PathAsserter ) {

    la := a.StartList()
    for _, val := range rt.Seq {
        assertReadValue( rd, val, la )
        la = la.Next()
    }
}
