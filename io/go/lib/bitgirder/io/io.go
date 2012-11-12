package io

import (
    "encoding/binary"
    "fmt"
    goIo "io"
)

func implClose( v interface{} ) error {
    if c, ok := v.( goIo.Closer ); ok { return c.Close() }
    return nil
}

type BinReader struct {
    rd goIo.Reader
    ord binary.ByteOrder
}

func NewReader( rd goIo.Reader, ord binary.ByteOrder ) *BinReader {
    return &BinReader{ rd, ord }
}

func NewLeReader( rd goIo.Reader ) *BinReader {
    return NewReader( rd, binary.LittleEndian )
}

func NewBeReader( rd goIo.Reader ) *BinReader {
    return NewReader( rd, binary.BigEndian )
}

func ( r *BinReader ) Close() error { return implClose( r.rd ) }

func ( r *BinReader ) ReadBin( data interface{} ) error {
    return binary.Read( r.rd, r.ord, data )
}

func ( r *BinReader ) ReadInt8() ( i int8, err error ) {
    err = r.ReadBin( &i )
    return
}

func ( r *BinReader ) ReadUint8() ( i uint8, err error ) {
    err = r.ReadBin( &i )
    return
}

func ( r *BinReader ) ReadInt32() ( i int32, err error ) {
    err = r.ReadBin( &i )
    return
}

func ( r *BinReader ) ReadUint32() ( i uint32, err error ) {
    err = r.ReadBin( &i )
    return
}

func ( r *BinReader ) ReadInt64() ( i int64, err error ) {
    err = r.ReadBin( &i )
    return
}

func ( r *BinReader ) ReadUint64() ( i uint64, err error ) {
    err = r.ReadBin( &i )
    return
}

func ( r *BinReader ) ReadFloat32() ( f float32, err error ) {
    err = r.ReadBin( &f )
    return
}

func ( r *BinReader ) ReadFloat64() ( f float64, err error ) {
    err = r.ReadBin( &f )
    return
}

func ( r *BinReader ) ReadBool() ( b bool, err error ) {
    i, err := r.ReadUint8()
    return i > 0, err
}

func ( r *BinReader ) ReadBuffer32() ( buf []byte, err error ) {
    var sz int32
    if sz, err = r.ReadInt32(); err != nil { return }
    if sz < 0 { 
        err = fmt.Errorf( "io.BinReader: read negative buffer size: %d", sz )
        return
    }
    buf = make( []byte, sz )
    _, err = goIo.ReadFull( r.rd, buf )
    return
}

func ( r *BinReader ) ReadUtf8() ( str string, err error ) {
    var buf []byte
    if buf, err = r.ReadBuffer32(); err != nil { return }
    str = string( buf )
    return
}

type BinWriter struct {
    wr goIo.Writer
    ord binary.ByteOrder
}

func NewWriter( wr goIo.Writer, ord binary.ByteOrder ) *BinWriter {
    return &BinWriter{ wr, ord }
}

func NewLeWriter( wr goIo.Writer ) *BinWriter {
    return NewWriter( wr, binary.LittleEndian )
}

func NewBeWriter( wr goIo.Writer ) *BinWriter {
    return NewWriter( wr, binary.BigEndian )
}

func ( w *BinWriter ) Close() error { return implClose( w.wr ) }

func ( w *BinWriter ) WriteBin( data interface{} ) error {
    return binary.Write( w.wr, w.ord, data )
}

func ( w *BinWriter ) WriteInt8( i int8 ) error { return w.WriteBin( i ) }
func ( w *BinWriter ) WriteUint8( i uint8 ) error { return w.WriteBin( i ) }
func ( w *BinWriter ) WriteInt32( i int32 ) error { return w.WriteBin( i ) }
func ( w *BinWriter ) WriteUint32( i uint32 ) error { return w.WriteBin( i ) }
func ( w *BinWriter ) WriteInt64( i int64 ) error { return w.WriteBin( i ) }
func ( w *BinWriter ) WriteUint64( i uint64 ) error { return w.WriteBin( i ) }
func ( w *BinWriter ) WriteFloat32( f float32 ) error { return w.WriteBin( f ) }
func ( w *BinWriter ) WriteFloat64( f float64 ) error { return w.WriteBin( f ) }

func ( w *BinWriter ) WriteBool( b bool ) error {
    val := uint8( 0 )
    if b { val = uint8( 1 ) }
    return w.WriteBin( val )
}

func ( w *BinWriter ) WriteBuffer32( buf []byte ) error {
    if err := w.WriteInt32( int32( len( buf ) ) ); err != nil { return err }
    _, err := w.wr.Write( buf )
    return err
}

// Can make this more efficient at some point
func ( w *BinWriter ) WriteUtf8( str string ) error {
    return w.WriteBuffer32( []byte( str ) )
}
