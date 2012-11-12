package io

import (
//    "log"
    "testing"
    "bytes"
    "bitgirder/assert"
    "io"
)

func TestBinReaderWriter( t *testing.T ) {
    buf := &bytes.Buffer{}
    var rd *BinReader
    var wr *BinWriter
    for _, le := range []bool { true, false } {
        if le {
            rd, wr = NewLeReader( buf ), NewLeWriter( buf )
        } else { rd, wr = NewBeReader( buf ), NewBeWriter( buf ) }
        buf1 := []byte{ 0, 1, 2 }
        str1 := "hello"
        if err := wr.WriteBin( int32( 0 ) ); err != nil { t.Fatal( err ) }
        if err := wr.WriteInt8( int8( 1 ) ); err != nil { t.Fatal( err ) }
        if err := wr.WriteUint8( uint8( 1 ) ); err != nil { t.Fatal( err ) }
        if err := wr.WriteInt32( int32( 1 ) ); err != nil { t.Fatal( err ) }
        if err := wr.WriteUint32( uint32( 1 ) ); err != nil { t.Fatal( err ) }
        if err := wr.WriteInt64( int64( 1 ) ); err != nil { t.Fatal( err ) }
        if err := wr.WriteUint64( uint64( 1 ) ); err != nil { t.Fatal( err ) }
        if err := wr.WriteFloat32( float32( 1.2 ) ); err != nil { 
            t.Fatal( err ) 
        }
        if err := wr.WriteFloat64( float64( 1.2 ) ); err != nil { 
            t.Fatal( err ) 
        }
        if err := wr.WriteBool( true ); err != nil { t.Fatal( err ) }
        if err := wr.WriteUint8( 2 ); err != nil { t.Fatal( err ) } // 'true'
        if err := wr.WriteBool( false ); err != nil { t.Fatal( err ) }
        if err := wr.WriteBuffer32( []byte{} ); err != nil { t.Fatal( err ) }
        if err := wr.WriteBuffer32( buf1 ); err != nil { t.Fatal( err ) }
        if err := wr.WriteUtf8( "" ); err != nil { t.Fatal( err ) }
        if err := wr.WriteUtf8( str1 ); err != nil { t.Fatal( err ) }
        i0 := new( int32 )
        if err := rd.ReadBin( i0 ); err == nil {
            assert.Equal( int32( 0 ), *i0 )
        } else { t.Fatal( err ) }
        if i, err := rd.ReadInt8(); err == nil {
            assert.Equal( int8( 1 ), i )
        } else { t.Fatal( err ) }
        if i, err := rd.ReadUint8(); err == nil {
            assert.Equal( uint8( 1 ), i )
        } else { t.Fatal( err ) }
        if i, err := rd.ReadInt32(); err == nil { 
            assert.Equal( int32( 1 ), i )
        } else { t.Fatal( err ) }
        if i, err := rd.ReadUint32(); err == nil {
            assert.Equal( uint32( 1 ), i )
        } else { t.Fatal( err ) }
        if i, err := rd.ReadInt64(); err == nil {
            assert.Equal( int64( 1 ), i )
        } else { t.Fatal( err ) }
        if i, err := rd.ReadUint64(); err == nil {
            assert.Equal( uint64( 1 ), i )
        } else { t.Fatal( err ) }
        if f, err := rd.ReadFloat32(); err == nil {
            assert.Equal( float32( 1.2 ), f )
        } else { t.Fatal( err ) }
        if f, err := rd.ReadFloat64(); err == nil {
            assert.Equal( float64( 1.2 ), f )
        } else { t.Fatal( err ) }
        chkBool := func( expct bool ) {
            if b, err := rd.ReadBool(); err == nil {
                assert.Equal( expct, b )
            } else { t.Fatal( err ) }
        }
        chkBool( true )
        chkBool( true )
        chkBool( false )
        if buf, err := rd.ReadBuffer32(); err == nil {
            assert.Equal( []byte{}, buf )
        } else { t.Fatal( err ) }
        if buf, err := rd.ReadBuffer32(); err == nil {
            assert.Equal( buf1, buf )
        } else { t.Fatal( err ) }
        if str, err := rd.ReadUtf8(); err == nil {
            assert.Equal( "", str )
        } else { t.Fatal( err ) }
        if str, err := rd.ReadUtf8(); err == nil {
            assert.Equal( str1, str )
        } else { t.Fatal( err ) }
        assert.Equal( 0, rd.rd.( *bytes.Buffer ).Len() )
    }
}

func TestReadErrorOnNegBufLen( t *testing.T ) {
    buf := &bytes.Buffer{}
    rd, wr := NewLeReader( buf ), NewLeWriter( buf )
    if err := wr.WriteInt32( -12 ); err != nil { t.Fatal( err ) }
    if _, err := rd.ReadBuffer32(); err == nil {
        t.Fatalf( "Expected neg-buf-len read error" )
    } else {
        assert.Equal( 
            "io.BinReader: read negative buffer size: -12", err.Error() )
    }
}

type NoOpReaderWriter struct {}

func ( n *NoOpReaderWriter ) Read( buf []byte ) ( int, error ) { return 0, nil }

func ( n *NoOpReaderWriter ) Write( buf []byte ) ( int, error ) { 
    return 0, nil 
}

type NoOpCloser struct { 
    *NoOpReaderWriter
    closeCalls int
}

func ( n *NoOpCloser ) Close() error { 
    n.closeCalls++
    return nil 
}

func TestCloseInvocations( t *testing.T ) {
    rw := &NoOpReaderWriter{}
    f := func( c io.Closer ) { 
        if err := c.Close(); err != nil { t.Fatal( err ) }
    }
    f( NewLeReader( rw ) )
    f( NewLeWriter( rw ) )
    c := &NoOpCloser{ rw, 0 }
    f( NewLeReader( c ) )
    assert.Equal( 1, c.closeCalls )
    f( NewLeWriter( c ) )
    assert.Equal( 2, c.closeCalls )
}
