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

func assertReadValue( rd *BinReader, expct mg.Value, a *assert.PathAsserter ) {
    if act, err := rd.ReadValue(); err == nil {
        mg.AssertEqualValues( expct, act, a )
    } else { a.Fatalf( "read val failed: %s", err ) }
}

func AssertRoundtripReadValue(
    rt *mg.BinIoRoundtripTest, rd *BinReader, a *assert.PathAsserter ) {
    
    assertReadValue( rd, rt.Val.( mg.Value ), a )
}
