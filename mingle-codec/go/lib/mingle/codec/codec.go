package codec

import (
    "io"
    "fmt"
    "bytes"
    mg "mingle"
//    "log"
)

type Codec interface {
    EncoderTo( w io.Writer ) mg.Reactor
    DecodeFrom( rd io.Reader, rct mg.Reactor ) error
}

func Encode( ms *mg.Struct, cdc Codec, w io.Writer ) error {
    rct := cdc.EncoderTo( w )
    return mg.VisitValue( ms, rct )
}

func EncodeBytes( ms *mg.Struct, cdc Codec ) ( []byte, error ) {
    buf := &bytes.Buffer{}
    if err := Encode( ms, cdc, buf ); err != nil { return nil, err }
    return buf.Bytes(), nil
}

func Decode( cdc Codec, rd io.Reader ) ( *mg.Struct, error ) {
    rct := mg.NewValueBuilder()
    if err := cdc.DecodeFrom( rd, rct ); err != nil { return nil, err }
    return rct.GetValue().( *mg.Struct ), nil
}

func DecodeBytes( cdc Codec, buf []byte ) ( *mg.Struct, error ) {
    return Decode( cdc, bytes.NewBuffer( buf ) )
}

type CodecError struct {
    msg string
}

func ( e *CodecError ) Error() string { return e.msg }

func Error( msg string ) *CodecError { return &CodecError{ msg } }

func Errorf( msg string, args ...interface{} ) *CodecError {
    return Error( fmt.Sprintf( msg, args... ) )
}

func regKeyFor( id *mg.Identifier ) string { return id.ExternalForm() }

type CodecRegistrationError struct { msg string }
func ( e *CodecRegistrationError ) Error() string { return e.msg }

type CodecRegistration struct {
    Codec Codec
    Id *mg.Identifier
    Source string
}

func ( r *CodecRegistration ) key() string { return regKeyFor( r.Id ) }

var registry = make( map[ string ]*CodecRegistration )

func GetCodecById( id *mg.Identifier ) Codec {
    if reg := registry[ regKeyFor( id ) ]; reg != nil { return reg.Codec }
    return nil
}

func MustCodecById( id *mg.Identifier ) Codec {
    if cdc := GetCodecById( id ); cdc != nil { return cdc }
    panic( fmt.Errorf( "No such codec: %s", id ) )
}

func RegisterCodec( reg *CodecRegistration ) error {
    key := reg.key()
    if prev, ok := registry[ key ]; ok {
        msg := "Codec %q already registered by %q"
        return &CodecRegistrationError{ fmt.Sprintf( msg, key, prev.Source ) }
    }
    registry[ key ] = reg
    return nil
}
