package io

import (
    "testing"
    "bytes"
    mg "mingle"
    mgt "mingle/testing"
    "bitgirder/assert"
)

func TestMustHeadersPairsFail( t *testing.T ) {
    assert.AssertPanic(
        func() { MustHeadersPairs( "f1", []interface{}{} ) },
        func( err interface{} ) {
            msg := `f1: Expected value of type mingle:core@v1/String but ` + 
                `found mingle:core@v1/Value*`
            assert.Equal( msg, err.( *mg.TypeCastError ).Error() )
        },
    )
}

func TestMustHeadersFail( t *testing.T ) {
    assert.AssertPanic(
        func() { MustHeaders( mg.MustSymbolMap( "f1", []interface{}{} ) ) },
        func( err interface{} ) {
            msg := `Non-String value for header 'f1': *mingle.List`
            assert.Equal( err.( error ).Error(), msg )
        },
    )
}

func TestHeaderRoundTrip( t *testing.T ) {
    f := func( h *Headers ) {
        buf := &bytes.Buffer{}
        if err := WriteHeaders( h, buf ); err != nil { t.Fatal( err ) }
        h2, err := ReadHeaders( buf )
        if err != nil { t.Fatal( err ) }
        mgt.LossyEqual( h.Fields(), h2.Fields(), t )
    }
    f( MustHeadersPairs() )
    f( MustHeadersPairs( "f1", "val1", "f2", 2 ) )
}

func TestInvalidHeadersVersionError( t *testing.T ) {
    buf := &bytes.Buffer{}
    badVer := int32( 30 )
    if err := WriteBinary( badVer, buf ); err != nil { t.Fatal( err ) }
    if _, err := ReadHeaders( buf ); err == nil {
        t.Fatal( "Expected error" )
    } else if ve, ok := err.( *InvalidVersionError ); ok {
        msg := `Invalid headers version: 0x0000001e (expected 0x00000001)`
        assert.Equal( msg, ve.Error() )
    } else { t.Fatal( err ) }
}

func TestInvalidTypeCodeError( t *testing.T ) {
    buf := &bytes.Buffer{}
    if err := WriteBinary( HeadersVersion1, buf ); err != nil { t.Fatal( err ) }
    badCode := int32( 12 )
    if err := WriteBinary( badCode, buf ); err != nil { t.Fatal( err ) }
    if _, err := ReadHeaders( buf ); err == nil {
        t.Fatal( "Expected error" )
    } else if tce, ok := err.( *InvalidTypeCodeError ); ok {
        assert.Equal( "Invalid type code: 0x0000000c", tce.Error() )
    } else { t.Fatal( err ) }
}
