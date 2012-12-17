package testing

import (
    gotest "testing"
    mg "mingle"
    "mingle/codec"
    mgtest "mingle/testing"
    mgio "mingle/io"
    "mingle/testing/testval"
    "bitgirder/assert"
    "bytes"
    "encoding/binary"
    "sort"
//    "encoding/hex"
    "fmt"
//    "log"
)

type TestSpecError struct { msg string }
func ( e *TestSpecError ) Error() string { return e.msg }

func keyStringFor( specId, codecId *mg.Identifier ) string {
    if codecId == nil { return specId.ExternalForm() }
    return specId.ExternalForm() + "/" + codecId.ExternalForm()
}

type Action interface {}

type TestSpec struct {
    CodecId *mg.Identifier
    CodecOpts *mgio.Headers
    Id *mg.Identifier
    Action
    OutboundCodec codec.Codec
}

func ( s *TestSpec ) KeyString() string {
    return keyStringFor( s.Id, s.CodecId )
}

var emptyHeaders = mgio.MustHeadersPairs()

func ( s *TestSpec ) Headers() *mgio.Headers {
    if s.CodecOpts == nil { return emptyHeaders }
    return s.CodecOpts
}

var orderLe = binary.LittleEndian

const (
    actTypeRoundTrip = int32( iota + 1 )
    actTypeFailDecode
    actTypeDecodeInput
    actTypeEncodeValue
)

type serializer struct {
    spec *TestSpec
    buf *bytes.Buffer
    cf CodecFactory
    cdc codec.Codec // lazily initialized
}

func ( s *serializer ) codec() codec.Codec {
    if s.cdc == nil { s.cdc = s.spec.OutboundCodec }
    if s.cdc == nil { s.cdc = s.cf( s.spec.Headers() ) }
    return s.cdc
}

func ( s *serializer ) writeBin( data interface{} ) error {
    return binary.Write( s.buf, orderLe, data )
}

func ( s *serializer ) writeUtf8( str string ) error {
    strBuf := []byte( str )
    sz := int32( len( strBuf ) )
    if err := s.writeBin( sz ); err != nil { return err }
    _, err := s.buf.Write( strBuf )
    return err
}

func ( s *serializer ) writeId( id *mg.Identifier ) error {
    str := ""
    if id != nil { str = id.ExternalForm() }
    return s.writeUtf8( str )
}

func ( s *serializer ) writeActCode( ac int32 ) error {
    return s.writeBin( ac )
}

func ( s *serializer ) writeBuf( buf []byte ) error {
    if err := s.writeBin( int32( len( buf ) ) ); err != nil { return err }
    _, err := s.buf.Write( buf )
    return err
}

func ( s *serializer ) writeStruct( ms *mg.Struct ) error {
    buf, err := codec.EncodeBytes( ms, s.codec() )
    if err != nil { return err }
    return s.writeBuf( buf )
}

func ( s *serializer ) writeRoundTrip( rt *RoundTrip ) error {
    if err := s.writeActCode( actTypeRoundTrip ); err != nil { return err }
    return s.writeStruct( rt.Object )
}

func ( s *serializer ) writeFailDecode( fd *FailDecode ) error {
    if err := s.writeActCode( actTypeFailDecode ); err != nil { return err }
    if err := s.writeUtf8( fd.ErrorMessage ); err != nil { return err }
    return s.writeBuf( fd.Input )
}

func ( s *serializer ) writeDecodeInput( di *DecodeInput ) error {
    if err := s.writeActCode( actTypeDecodeInput ); err != nil { return err }
    if err := s.writeBuf( di.Input ); err != nil { return err }
    return s.writeStruct( di.Expect )
}

func ( s *serializer ) writeEncodeValue( ev *EncodeValue ) error {
    if err := s.writeActCode( actTypeEncodeValue ); err != nil { return err }
    return s.writeStruct( ev.Value )
}

func ( s *serializer ) writeAction( a Action ) error {
    switch v := a.( type ) {
    case *RoundTrip: return s.writeRoundTrip( v )
    case *FailDecode: return s.writeFailDecode( v )
    case *DecodeInput: return s.writeDecodeInput( v )
    case *EncodeValue: return s.writeEncodeValue( v )
    }
    panic( fmt.Errorf( "Unhandled action type: %T", a ) )
}

func ( s *TestSpec ) GetBytes( cf CodecFactory ) ( buf []byte, err error ) {
    ser := &serializer{ spec: s, buf: &bytes.Buffer{}, cf: cf }
    if err = ser.writeId( s.Id ); err != nil { return }
    if err = ser.writeId( s.CodecId ); err != nil { return }
    if err = mgio.WriteHeaders( s.Headers(), ser.buf ); err != nil { return }
    if err = ser.writeAction( s.Action ); err != nil { return }
    buf = ser.buf.Bytes()
    return
}

type RoundTrip struct {
    Object *mg.Struct
}

type FailDecode struct {
    ErrorMessage string
    Input []byte
}

type DecodeInput struct {
    Input []byte
    Expect *mg.Struct
}

type CheckableEncode struct {
    Buffer []byte
    *assert.Asserter
}

type EncodeCheck func( ce *CheckableEncode )

type EncodeValue struct {
    Value *mg.Struct
    Check EncodeCheck
}

type TestEngine struct {
    specs map[ string ]*TestSpec // keystring comes from *TestSpec.KeyString()
    codecFactories map[ string ]CodecFactory
}

func newTestEngine() *TestEngine {
    return &TestEngine{ 
        specs: make( map[ string ]*TestSpec ),
        codecFactories: make( map[ string ]CodecFactory ),
    }
}

func ( e *TestEngine ) PutCodecFactory( id *mg.Identifier, cf CodecFactory ) {
    e.codecFactories[ id.ExternalForm() ] = cf
}

func ( e *TestEngine ) GetCodecFactory( id *mg.Identifier ) CodecFactory {
    return e.codecFactories[ id.ExternalForm() ]
}

func ( e *TestEngine ) MustPutSpecs( specs ...*TestSpec ) {
    for _, spec := range specs {
        key := spec.KeyString()
        if prev := e.specs[ key ]; prev != nil {
            panic( fmt.Errorf( "Spec %s is already set", key ) )
        } else { e.specs[ key ] = spec }
    }
}

func ( e *TestEngine ) SpecsFor( codecId *mg.Identifier ) []*TestSpec {
    res := make( []*TestSpec, 0, 10 )
    for _, spec := range e.specs {
        if id := spec.CodecId; id == nil || id.Equals( codecId ) {
            res = append( res, spec )
        }
    }
    return res
}

func ( e *TestEngine ) SpecKeysFor( codecId *mg.Identifier ) []string {
    specs := e.SpecsFor( codecId )
    res := make( []string, len( specs ) )
    for i, spec := range specs { res[ i ] = spec.KeyString() }
    return res
}

func ( e *TestEngine ) GetTestSpec( key string ) *TestSpec {
    return e.specs[ key ]
}

var stdEngine *TestEngine

func init() {
    stdEngine = newTestEngine()
    stdEngine.MustPutSpecs( 
        &TestSpec{ 
            Id: mg.MustIdentifier( "test-struct1-inst1" ),
            Action: &RoundTrip{ testval.TestStruct1Inst1 },
        },
        &TestSpec{
            Id: mg.MustIdentifier( "type-cov-struct1" ),
            Action: &RoundTrip{ testval.TypeCovStruct1 },
        },
        &TestSpec{
            Id: mg.MustIdentifier( "empty-struct" ),
            Action: &RoundTrip{ mg.MustStruct( "ns1@v1/S1" ) },
        },
        &TestSpec{
            Id: mg.MustIdentifier( "empty-val-struct" ),
            Action: &RoundTrip{
                mg.MustStruct( "ns1@v1/S1",
                    "buf1", mg.Buffer( []byte{} ),
                    "str1", "",
                    "list1", mg.MustList(),
                    "map1", mg.MustSymbolMap(),
                ),
            },
        },
        &TestSpec{
            Id: mg.MustIdentifier( "nulls-in-list" ),
            Action: &RoundTrip{
                mg.MustStruct( "ns1@v1/S1",
                    "list1", mg.MustList( "s1", nil, nil, "s4" ) ) },
        },
        &TestSpec{
            Id: mg.MustIdentifier( "unicode-handler" ),
            Action: &RoundTrip{
                mg.MustStruct( "ns1@v1/S1",
                    "s0", "hello",
                    "s1", "\u01FE", // Ǿ
                    "s2", "\U0001D11E", // g-clef (utf-16 \uD83F\uDD1E)
                ),
            },
        },
    )
}

func GetDefaultTestEngine() *TestEngine { return stdEngine }

type CodecFactory func( codecOpts *mgio.Headers ) codec.Codec

type stdTest struct {
    *gotest.T
    fact CodecFactory
    spec *TestSpec
}

func ( t *stdTest ) codec() codec.Codec {
    return t.fact( t.spec.Headers() )
}

func ( t *stdTest ) callRoundTrip() {
    rt := t.spec.Action.( *RoundTrip )
    buf, err := codec.EncodeBytes( rt.Object, t.codec() )
//    log.Printf( "Encoded:\n%s", hex.Dump( buf ) )
    if err != nil { t.Fatal( err ) }
    var act mg.Value
//    log.Printf( "Starting decode" )
    if act, err = codec.DecodeBytes( t.codec(), buf ); err != nil { 
        t.Fatal( err ) 
    }
//    log.Printf( "Decoded: %s", mg.QuoteValue( act ) )
    mgtest.LossyEqual( rt.Object, act, t )
}

func ( t *stdTest ) callDecodeError() {
    fd := t.spec.Action.( *FailDecode )
    if _, err := codec.DecodeBytes( t.codec(), fd.Input ); err == nil {
        t.Fatalf( "Expected error %q", fd.ErrorMessage )
    } else if ce, ok := err.( *codec.CodecError ); ok {
        assert.Equal( fd.ErrorMessage, ce.Error() )
    } else { t.Fatal( err ) }
}

func ( t *stdTest ) callDecodeInput() {
    di := t.spec.Action.( *DecodeInput )
    if ms, err := codec.DecodeBytes( t.codec(), di.Input ); err == nil {
        mgtest.LossyEqual( di.Expect, ms, t )
    } else { t.Fatal( err ) }
}

func ( t *stdTest ) callEncodeValue() {
    ev := t.spec.Action.( *EncodeValue )
    if buf, err := codec.EncodeBytes( ev.Value, t.codec() ); err == nil {
        a := &assert.Asserter{ t.T }
        ev.Check( &CheckableEncode{ buf, a } )
    } else { t.Fatal( err ) }
}

func ( t *stdTest ) call() {
//    t.Logf( "Calling test on spec %s", t.spec.KeyString() )
    switch v := t.spec.Action.( type ) {
    case *RoundTrip: t.callRoundTrip()
    case *FailDecode: t.callDecodeError()
    case *DecodeInput: t.callDecodeInput()
    case *EncodeValue: t.callEncodeValue()
    default: t.Fatalf( "Unexpected action: %T", v )
    }
}

type byNameSorter []*TestSpec

func ( s byNameSorter ) Len() int { return len( s ) }

func ( s byNameSorter ) Less( i, j int ) bool {
    return s[ i ].KeyString() < s[ j ].KeyString()
}

func ( s byNameSorter ) Swap( i, j int ) { s[ i ], s[ j ] = s[ j ], s[ i ] }

func TestCodecSpecs( codecId *mg.Identifier, t *gotest.T ) {
    fact := stdEngine.GetCodecFactory( codecId )
    if fact == nil { t.Fatalf( "No codec factory for %s", codecId ) }
    specsOrig := stdEngine.SpecsFor( codecId )
    sorted := make( []*TestSpec, len( specsOrig ) )
    copy( sorted, specsOrig ) 
    sort.Sort( byNameSorter( sorted ) )
    for _, spec := range sorted {
        (&stdTest{ fact: fact, T: t, spec: spec }).call() 
    }
}

func TestCodecSpec( codecId *mg.Identifier, key string, t *gotest.T ) {
    fact := stdEngine.GetCodecFactory( codecId )
    if spec := stdEngine.GetTestSpec( key ); spec == nil {
        t.Fatalf( "No such test spec: %s", key )
    } else {
        (&stdTest{ fact: fact, T: t, spec: spec }).call()
    }
}

func TestCodecRegistration( 
    id *mg.Identifier, t *gotest.T, f func( cdc codec.Codec ) ) {
    if cdc := codec.GetCodecById( id ); cdc == nil {
        t.Fatalf( "no codec with id %s", id )
    } else { f( cdc ) }
}
