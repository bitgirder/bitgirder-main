package bincodec

import (
    mg "mingle"
    "mingle/codec"
//    "log"
    "io"
    "encoding/binary"
)

var CodecId = mg.MustIdentifier( "binary" )

type TypeCode uint8

var (
    TypeCodeEnd = TypeCode( 0x00 )
    TypeCodeBoolean = TypeCode( 0x01 )
    TypeCodeFloat64 = TypeCode( 0x03 )
    TypeCodeEnum = TypeCode( 0x04 )
    TypeCodeFloat32 = TypeCode( 0x05 )
    TypeCodeInt32 = TypeCode( 0x06 )
    TypeCodeUint32 = TypeCode( 0x07 )
    TypeCodeInt64 = TypeCode( 0x08 )
    TypeCodeUint64 = TypeCode( 0x09 )
    TypeCodeString = TypeCode( 0x0a )
    TypeCodeTimestamp = TypeCode( 0x0b )
    TypeCodeBuffer = TypeCode( 0x0d )
    TypeCodeUtf8String = TypeCode( 0x0e )
    TypeCodeList = TypeCode( 0x0f )
    TypeCodeStruct = TypeCode( 0x10 )
    TypeCodeSymbolMap = TypeCode( 0x11 )
    TypeCodeNull = TypeCode( 0x12 )
    TypeCodeField = TypeCode( 0x13 )
)

var byteOrder = binary.LittleEndian

type BinCodec struct {
}

func New() *BinCodec { return &BinCodec{} }

type encoder struct {
    w io.Writer
    bc *BinCodec
    impl *mg.ReactorImpl
    mgWr *mg.BinWriter
}

func ( e *encoder ) writeBin( data interface{} ) error {
    return binary.Write( e.w, byteOrder, data )
}

func ( e *encoder ) writeInt32( i int32 ) error { return e.writeBin( i ) }

func ( e *encoder ) writeTypeCode( tc TypeCode ) error {
    return e.writeBin( uint8( tc ) )
}

func ( e *encoder ) writeSizedBuffer( buf []byte ) ( err error ) {
    if err = e.writeInt32( int32( len( buf ) ) ); err != nil { return }
    _, err = e.w.Write( buf )
    return
}

func ( e *encoder ) writeUtf8( str string ) ( err error ) {
    if err = e.writeTypeCode( TypeCodeUtf8String ); err != nil { return }
    return e.writeSizedBuffer( []byte( str ) )
}

func ( e *encoder ) writeBoolean( b mg.Boolean ) error {
    val := int8( 0 )
    if bool( b ) { val = int8( 1 ) }
    if err := e.writeTypeCode( TypeCodeBoolean ); err != nil { return err }
    return e.writeBin( val )
}

func ( e *encoder ) writeNumber( tc TypeCode, data interface{} ) error {
    if err := e.writeTypeCode( tc ); err != nil { return err }
    return e.writeBin( data )
}

func ( e *encoder ) writeMgString( s mg.String ) error {
    if err := e.writeTypeCode( TypeCodeString ); err != nil { return err }
    return e.writeUtf8( string( s ) )
}

func ( e *encoder ) writeMgBuffer( b mg.Buffer ) error {
    if err := e.writeTypeCode( TypeCodeBuffer ); err != nil { return err }
    return e.writeSizedBuffer( []byte( b ) )
}

func ( e *encoder ) writeTimestamp( t mg.Timestamp ) error {
    if err := e.writeTypeCode( TypeCodeTimestamp ); err != nil { return err }
    return e.mgWr.WriteValue( t )
}

func ( e *encoder ) writeType( typ mg.TypeReference ) error {
    return e.mgWr.WriteTypeReference( typ )
}

func ( e *encoder ) writeIdentifier( id *mg.Identifier ) error {
    return e.mgWr.WriteIdentifier( id )
}

func ( e *encoder ) writeEnum( en *mg.Enum ) error {
    if err := e.writeTypeCode( TypeCodeEnum ); err != nil { return err }
    if err := e.writeType( en.Type ); err != nil { return err }
    return e.writeIdentifier( en.Value )
}

func ( e *encoder ) Value( val mg.Value ) error {
    if err := e.impl.Value(); err != nil { return err }
    switch v := val.( type ) {
    case mg.Boolean: return e.writeBoolean( v )
    case mg.Int32: return e.writeNumber( TypeCodeInt32, int32( v ) )
    case mg.Int64: return e.writeNumber( TypeCodeInt64, int64( v ) )
    case mg.Uint32: return e.writeNumber( TypeCodeUint32, uint32( v ) )
    case mg.Uint64: return e.writeNumber( TypeCodeUint64, uint64( v ) )
    case mg.Float32: return e.writeNumber( TypeCodeFloat32, float32( v ) )
    case mg.Float64: return e.writeNumber( TypeCodeFloat64, float64( v ) )
    case mg.String: return e.writeMgString( v )
    case mg.Timestamp: return e.writeTimestamp( v )
    case mg.Buffer: return e.writeMgBuffer( v )
    case *mg.Enum: return e.writeEnum( v )
    case *mg.Null: return e.writeTypeCode( TypeCodeNull )
    }
    panic( libErrorf( "Unhandled encode Value(): %T", val ) )
}

func ( e *encoder ) StartStruct( typ mg.TypeReference ) error {
    if err := e.impl.StartStruct(); err != nil { return err }
    if err := e.writeTypeCode( TypeCodeStruct ); err != nil { return err }
    if err := e.writeInt32( -1 ); err != nil { return err } // skip size for now
    return e.writeType( typ )
}

func ( e *encoder ) StartMap() error {
    if err := e.impl.StartMap(); err != nil { return err }
    return e.writeTypeCode( TypeCodeSymbolMap )
}

func ( e *encoder ) StartList() error {
    if err := e.impl.StartList(); err != nil { return err }
    if err := e.writeTypeCode( TypeCodeList ); err != nil { return err }
    return e.writeInt32( -1 ) // write -1 as size
} 

func ( e *encoder ) StartField( fld *mg.Identifier ) error {
    if err := e.impl.StartField( fld ); err != nil { return err }
    if err := e.writeTypeCode( TypeCodeField ); err != nil { return err }
    return e.writeIdentifier( fld )
}

func ( e *encoder ) End() error {
    if err := e.impl.End(); err != nil { return err }
    return e.writeTypeCode( TypeCodeEnd )
}

func ( bc *BinCodec ) EncoderTo( w io.Writer ) mg.Reactor {
    return &encoder{ 
        w: w, 
        bc: bc, 
        impl: mg.NewReactorImpl(),
        mgWr: mg.NewWriter( w ),
    }
}

type offsetTracker struct {
    r io.Reader
    off int64
}

func ( t *offsetTracker ) Read( p []byte ) ( n int, err error ) {
    n, err = t.r.Read( p )
    t.off += int64( n )
    return
}

type decode struct {
    bc *BinCodec
    r *offsetTracker
    rct mg.Reactor
    mgRd *mg.BinReader
}

func codecErrorfWithOffset( 
    tmpl string, off int64, args ...interface{} ) error {
    tmpl = "[offset %d]: " + tmpl
    args2 := make( []interface{}, 1, 1 + len( args ) )
    args2[ 0 ] = off
    args2 = append( args2, args... )
    return codec.Errorf( tmpl, args2... )
}

func ( dec *decode ) codecErrorf( tmpl string, args ...interface{} ) error {
    return codecErrorfWithOffset( tmpl, dec.r.off, args... )
}

func ( dec *decode ) unexpectedTypeCode( tc TypeCode ) error {
    off := dec.r.off - int64( 1 ) // reset to before the offending tc val
    return codecErrorfWithOffset( "Unexpected type code: 0x%02x", off, tc )
}

func ( dec *decode ) readBin( data interface{} ) error {
    return binary.Read( dec.r, byteOrder, data )
}

func ( dec *decode ) readUint8() ( num uint8, err error ) {
    err = dec.readBin( &num )
    return
}

func ( dec *decode ) readInt8() ( num int8, err error ) {
    err = dec.readBin( &num )
    return
}

func ( dec *decode ) readTypeCode() ( tc TypeCode, err error ) {
    var num uint8
    if num, err = dec.readUint8(); err == nil { tc = TypeCode( num ) }
    return
}

func ( dec *decode ) expectTypeCode( expct TypeCode ) ( TypeCode, error ) {
    off := dec.r.off
    tc, err := dec.readTypeCode()
    if err == nil && tc != expct {
        tmpl := "Saw type code 0x%02x but expected 0x%02x"
        err = codecErrorfWithOffset( tmpl, off, tc, expct )
    }
    return tc, err
}

func ( dec *decode ) readInt32() ( num int32, err error ) {
    err = dec.readBin( &num )
    return
}

func ( dec *decode ) readSizedBuffer() ( []byte, int64, error ) {
    sz, err := dec.readInt32()
    if err != nil { return nil, -1, err }
    buf := make( []byte, sz )
    off := dec.r.off
    _, err = io.ReadFull( dec.r, buf )
    return buf, off, err
}

func ( dec *decode ) readUtf8() ( string, int64, error ) {
    buf, off, err := dec.readSizedBuffer()
    if err != nil { return "", -1, err }
    return string( buf ), off, nil
}

func ( dec *decode ) readString() ( string, int64, error ) {
    tc, err := dec.readTypeCode()
    if err != nil { return "", -1, err }
    switch tc {
    case TypeCodeUtf8String: return dec.readUtf8()
    }
    return "", -1, dec.unexpectedTypeCode( tc )
}

func ( dec *decode ) readMgString() error {
    var s string
    s, _, err := dec.readString()
    if err != nil { return err }
    return dec.rct.Value( mg.String( s ) )
}

func ( dec *decode ) readMgBuffer() error {
    buf, _, err := dec.readSizedBuffer()
    if err != nil { return err }
    return dec.rct.Value( mg.Buffer( buf ) )
}

func ( dec *decode ) readBoolean() error {
    val, err := dec.readInt8()
    if err != nil { return err }
    return dec.rct.Value( mg.Boolean( val > 0 ) )
}

func ( dec *decode ) readMgFloat64() error {
    var num float64
    if err := dec.readBin( &num ); err != nil { return err }
    return dec.rct.Value( mg.Float64( num ) )
}

func ( dec *decode ) readMgFloat32() error {
    var num float32
    if err := dec.readBin( &num ); err != nil { return err }
    return dec.rct.Value( mg.Float32( num ) )
}

func ( dec *decode ) readMgInt32() error {
    var num int32
    if err := dec.readBin( &num ); err != nil { return err }
    return dec.rct.Value( mg.Int32( num ) )
}

func ( dec *decode ) readMgInt64() error {
    var num int64
    if err := dec.readBin( &num ); err != nil { return err }
    return dec.rct.Value( mg.Int64( num ) )
}

func ( dec *decode ) readMgUint32() error {
    var num uint32
    if err := dec.readBin( &num ); err != nil { return err }
    return dec.rct.Value( mg.Uint32( num ) )
}

func ( dec *decode ) readMgUint64() error {
    var num uint64
    if err := dec.readBin( &num ); err != nil { return err }
    return dec.rct.Value( mg.Uint64( num ) )
}

func ( dec *decode ) readTimestamp() error {
    off := dec.r.off
    val, err := dec.mgRd.ReadValue()
    if err == nil { 
        if tm, ok := val.( mg.Timestamp ); ok {
            return dec.rct.Value( tm )
        } else { return codecErrorfWithOffset( "Expected timestamp", off ) }
    } 
    return err
}

func ( dec *decode ) readType() ( typ mg.TypeReference, err error ) {
    return dec.mgRd.ReadTypeReference()
}

func ( dec *decode ) readIdentifier() ( id *mg.Identifier, err error ) {
    return dec.mgRd.ReadIdentifier()
}

func ( dec *decode ) readField() error {
    if id, err := dec.readIdentifier(); err == nil {
        if err = dec.rct.StartField( id ); err != nil { return err }
    } else { return err }
    tc, err := dec.readTypeCode()
    if err != nil { return err }
    return dec.readValue( tc )
}

func ( dec *decode ) readFields() error {
    for {
        tc, err := dec.readTypeCode()
        if err != nil { return err }
        switch tc {
        case TypeCodeEnd: return dec.rct.End()
        case TypeCodeField: if err = dec.readField(); err != nil { return err }
        default: return dec.unexpectedTypeCode( tc )
        }
    }
    panic( libErrorf( "unreachable" ) )
}

func ( dec *decode ) readSymbolMap() error { 
    if err := dec.rct.StartMap(); err != nil { return err }
    return dec.readFields() 
}

func ( dec *decode ) readStruct() error {
    if _, err := dec.readInt32(); err != nil { return err } // size ignored
    if typ, err := dec.readType(); err == nil { 
        if err = dec.rct.StartStruct( typ ); err != nil { return err }
    } else { return err }
    return dec.readFields()
}

func ( dec *decode ) readEnum() error {
    typ, err := dec.readType()
    if err != nil { return err }
    val, err := dec.readIdentifier()
    if err != nil { return err }
    return dec.rct.Value( &mg.Enum{ typ, val } )
}

func ( dec *decode ) readList() error {
    // first off: read and ignore (for now, at least) list size
    if _, err := dec.readInt32(); err != nil { return err } 
    if err := dec.rct.StartList(); err != nil { return err }
    for {
        tc, err := dec.readTypeCode()
        if err != nil { return err }
        if tc == TypeCodeEnd { return dec.rct.End() }
        if err = dec.readValue( tc ); err != nil { return err }
    }
    panic( libErrorf( "unreachable" ) )
}

func ( dec *decode ) readValue( tc TypeCode ) error {
    switch tc {
    case TypeCodeBoolean: return dec.readBoolean()
    case TypeCodeFloat64: return dec.readMgFloat64()
    case TypeCodeFloat32: return dec.readMgFloat32()
    case TypeCodeInt32: return dec.readMgInt32()
    case TypeCodeInt64: return dec.readMgInt64()
    case TypeCodeUint32: return dec.readMgUint32()
    case TypeCodeUint64: return dec.readMgUint64()
    case TypeCodeTimestamp: return dec.readTimestamp()
    case TypeCodeBuffer: return dec.readMgBuffer()
    case TypeCodeString: return dec.readMgString()
    case TypeCodeList: return dec.readList()
    case TypeCodeStruct: return dec.readStruct()
    case TypeCodeSymbolMap: return dec.readSymbolMap()
    case TypeCodeEnum: return dec.readEnum()
    case TypeCodeNull: return dec.rct.Value( mg.NullVal )
    }
    return dec.unexpectedTypeCode( tc )
}

func ( bc *BinCodec ) DecodeFrom( r io.Reader, rct mg.Reactor ) error {
    ot := &offsetTracker{ r: r }
    dec := &decode{ bc: bc, r: ot, rct: rct, mgRd: mg.NewReader( ot ) }
    if _, err := dec.expectTypeCode( TypeCodeStruct ); err != nil { return err }
    return dec.readStruct()
}

func init() {
    codec.RegisterCodec(
        &codec.CodecRegistration{
            Codec: New(),
            Id: CodecId,
            Source: "mingle/codec/bincodec",
        },
    )
}
