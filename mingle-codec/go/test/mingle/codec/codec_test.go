package codec

import (
    "testing"
    "io"
    mg "mingle"
    "bitgirder/assert"
    "errors"
//    "log"
)

var noOpCodecErr = errors.New( "no-op codec; nothing to see here" )

type NoOpCodec struct {}

func ( c *NoOpCodec ) EncoderTo( w io.Writer ) mg.Reactor {
    panic( noOpCodecErr )
}

func ( c *NoOpCodec ) DecodeFrom( rd io.Reader, r mg.Reactor ) error {
    return noOpCodecErr
}

var noOpReg = 
    &CodecRegistration{
        Codec: &NoOpCodec{},
        Id: mg.MustIdentifier( "no-op" ),
        Source: "codec/no-op",
    }

func TestCodecRegistryAddAndAccess( t *testing.T ) {
    if err := RegisterCodec( noOpReg ); err != nil { t.Fatal( err ) }
    reg := &CodecRegistration{
        Codec: &NoOpCodec{},
        Id: mg.MustIdentifier( "no-op" ),
        Source: "irrelevant",
    }
    if err := RegisterCodec( reg ); err == nil {
        t.Fatal( "Expected error" )
    } else if re, ok := err.( *CodecRegistrationError ); ok {
        expct := `Codec "no-op" already registered by "codec/no-op"`
        assert.Equal( expct, re.Error() )
    } else { t.Fatal( err ) }
    assert.Equal( noOpReg.Codec, GetCodecById( noOpReg.Id ) )
    assert.Equal( nil, GetCodecById( mg.MustIdentifier( "blah" ) ) )
}

var fixedCodecStruct = mg.MustStruct( "ns1@v1/S1" )
var fixedCodecBuf = []byte{ 0 }

type fixedValueCodec struct {}

type fixedValueWriteReactor struct {
    w io.Writer
    didWrite bool
}

func ( f *fixedValueWriteReactor ) StartStruct( typ mg.TypeReference ) error {
    if ! f.didWrite { 
        if _, err := f.w.Write( fixedCodecBuf ); err != nil { return err }
        f.didWrite = true
    }
    return nil
}

func ( f *fixedValueWriteReactor ) StartList() error { return nil }
func ( f *fixedValueWriteReactor ) StartMap() error { return nil }

func ( f *fixedValueWriteReactor ) StartField( fld *mg.Identifier ) error {
    return nil
}

func ( f *fixedValueWriteReactor ) Value( v mg.Value ) error { return nil }
func ( f *fixedValueWriteReactor ) End() error { return nil }

func ( f *fixedValueCodec ) EncoderTo( w io.Writer ) mg.Reactor {
    return &fixedValueWriteReactor{ w: w }
}

type discardWriter int

func ( w discardWriter ) Write( p []byte ) ( int, error ) {
    return len( p ), nil
}

func ( f *fixedValueCodec ) DecodeFrom( rd io.Reader, rct mg.Reactor ) error {
    if _, err := io.Copy( discardWriter( 0 ), rd ); err != nil { return err }
    return mg.VisitValue( fixedCodecStruct, rct )
}

func TestCodecBufferUtilMethods( t *testing.T ) {
    c := &fixedValueCodec{}
    if buf, err := EncodeBytes( fixedCodecStruct, c ); err == nil {
        assert.Equal( fixedCodecBuf, buf )
    } else { t.Fatal( err ) }
    if ms, err := DecodeBytes( c, fixedCodecBuf ); err == nil {
        assert.Equal( fixedCodecStruct, ms )
    } else { t.Fatal( err ) }
}
