package codec

import (
    "testing"
    "io"
    mg "mingle"
    "mingle/testing/testval"
    "bitgirder/assert"
    "errors"
//    "log"
)

var noOpCodecErr = errors.New( "no-op codec; nothing to see here" )

type NoOpCodec struct {}

func ( c *NoOpCodec ) EncoderTo( w io.Writer ) Reactor {
    panic( noOpCodecErr )
}

func ( c *NoOpCodec ) DecodeFrom( rd io.Reader, r Reactor ) error {
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

func ( f *fixedValueCodec ) EncoderTo( w io.Writer ) Reactor {
    return &fixedValueWriteReactor{ w: w }
}

type discardWriter int

func ( w discardWriter ) Write( p []byte ) ( int, error ) {
    return len( p ), nil
}

func ( f *fixedValueCodec ) DecodeFrom( rd io.Reader, rct Reactor ) error {
    if _, err := io.Copy( discardWriter( 0 ), rd ); err != nil { return err }
    return VisitStruct( fixedCodecStruct, rct )
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

func callTestStructReactors( ms *mg.Struct, t *testing.T ) {
    rct := NewStructBuilder()
    if err := VisitStruct( ms, rct ); err == nil {
        assert.Equal( ms, rct.GetStruct() )
    } else { t.Fatal( err ) }
}

func TestStructReactors( t *testing.T ) {
    callTestStructReactors( testval.TypeCovStruct1, t )
    callTestStructReactors( 
        mg.MustStruct( "ns1@v1/S1",
            "list1", mg.MustList(),
            "map1", mg.MustSymbolMap(),
            "struct1", mg.MustStruct( "ns1@v1/S2" ),
        ),
        t,
    )
}

type reactorErrorTest struct {
    *testing.T
    seq []string
    errMsg string
}

func ( t *reactorErrorTest ) feedCall( call string, rct Reactor ) error {
    switch call {
    case "start-struct": 
        return rct.StartStruct( mg.MustTypeReference( "ns1@v1/S1" ) )
    case "start-map": return rct.StartMap()
    case "start-list": return rct.StartList()
    case "start-field1": return rct.StartField( mg.MustIdentifier( "f1" ) )
    case "start-field2": return rct.StartField( mg.MustIdentifier( "f2" ) )
    case "value": return rct.Value( mg.Int64( int64( 1 ) ) )
    case "end": return rct.End()
    }
    panic( errorf( "Unexpected test call: %s", call ) )
}

func ( t *reactorErrorTest ) feedSequence( rct Reactor ) error {
    for _, call := range t.seq {
        if err := t.feedCall( call, rct ); err != nil { return err }
    }
    return nil
}

func ( t *reactorErrorTest ) call() {
    rct := NewStructBuilder()
    if err := t.feedSequence( rct ); err == nil {
        t.Fatalf( "Got no err for %#v", t.seq )
    } else {
        if _, ok := err.( *CodecError ); ok {
            assert.Equal( t.errMsg, err.Error() )
        } else { t.Fatal( err ) }
    }
}

func TestReactorErrors( t *testing.T ) {
    tests := []*reactorErrorTest{
        { seq: []string{ "start-struct", "start-field1", "start-field2" }, 
          errMsg: "Saw start of field 'f2' while expecting a value for 'f1'",
        },
        { seq: []string{ 
            "start-struct", "start-field1", "start-map", "start-field1",
            "start-field2" },
          errMsg: "Saw start of field 'f2' while expecting a value for 'f1'",
        },
        { seq: []string{ "start-struct", "end", "start-field1" },
          errMsg: "StartField() called, but struct is built",
        },
        { seq: []string{ "start-struct", "value" },
          errMsg: "Expected field name or end of fields but got value",
        },
        { seq: []string{ "start-struct", "start-list" },
          errMsg: "Expected field name or end of fields but got list start",
        },
        { seq: []string{ "start-struct", "start-map" },
          errMsg: "Expected field name or end of fields but got map start",
        },
        { seq: []string{ "start-struct", "start-struct" },
          errMsg: "Expected field name or end of fields but got struct start",
        },
        { seq: []string{ "start-struct", "start-field1", "end" },
          errMsg: "Saw end while expecting value for field 'f1'",
        },
        { seq: []string{ 
            "start-struct", "start-field1", "start-list", "start-field1" },
          errMsg: "Expected list value but got start of field 'f1'",
        },
        { seq: []string{ "value" },
          errMsg: "Expected top-level struct start but got value",
        },
        { seq: []string{ "start-list" },
          errMsg: "Expected top-level struct start but got list start",
        },
        { seq: []string{ "start-map" },
          errMsg: "Expected top-level struct start but got map start",
        },
        { seq: []string{ "start-field1" },
          errMsg: "Expected top-level struct start but got field 'f1'",
        },
        { seq: []string{ "end" },
          errMsg: "Expected top-level struct start but got end",
        },
        { seq: []string{ 
            "start-struct", 
            "start-field1", "value",
            "start-field2", "value",
            "start-field1", "value",
            "end",
          },
          errMsg: "Invalid fields: Multiple entries for key: f1",
        },
    }
    for _, test := range tests {
        test.T = t
        test.call()
    }
}
