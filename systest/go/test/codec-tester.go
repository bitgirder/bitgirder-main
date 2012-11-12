package main

import (
    "log"
    "os"
    "fmt"
    "strings"
    "bitgirder/objpath"
    "bitgirder/assert"
    mg "mingle"
    mgio "mingle/io"
    mgTesting "mingle/testing"
    "mingle/io/stream"
    "mingle/codec/testing"
    "mingle/codec"
    _ "mingle/codec/json"
    _ "mingle/codec/bincodec"
)

var (
    idCommand = mg.MustIdentifier( "command" )
    idSpecKey = mg.MustIdentifier( "spec-key" )
    idCodecId = mg.MustIdentifier( "codec-id" )
    idException = mg.MustIdentifier( "exception" )

    cmdGetSpecKeys = mg.MustIdentifier( "get-spec-keys" )
    cmdGetSpec = mg.MustIdentifier( "get-spec" )
    cmdCheckEncode = mg.MustIdentifier( "check-encode" )
    cmdClose = mg.MustIdentifier( "close" )

    pathHdrs = objpath.RootedAt( mg.MustIdentifier( "headers" ) )

    eng = testing.GetDefaultTestEngine()
)

type request struct {
    *mg.SymbolMapAccessor
}

func headersAccessorFor( hdrs *mgio.Headers ) *mg.SymbolMapAccessor {
    return mg.NewSymbolMapAccessor( hdrs.Fields(), pathHdrs )
}

type server struct {
    conn stream.Connection
}

func ( s *server ) respond( hdrs *mgio.Headers, body []byte ) {
    if err := stream.WriteMessageBytes( s.conn, hdrs, body ); err != nil {
        log.Fatal( err )
    }
}

func ( s *server ) respondErrorf( msg string, args ...interface{} ) {
    errStr := fmt.Sprintf( msg, args... )
    hdrs := mgio.MustHeadersPairs( idException, errStr )
    s.respond( hdrs, []byte{} )
}

func ( s *server ) respondEmpty() {
    s.respond( mgio.MustHeadersPairs(), []byte{} )
}

func ( s *server ) getValAsId( 
    fld *mg.Identifier, req *request ) *mg.Identifier {
    str, err := req.GetGoStringById( fld )
    if err != nil {
        s.respondErrorf( "%s", err )
        return nil
    }
    id, err := mg.ParseIdentifier( str )
    if err != nil {
        s.respondErrorf( "Invalid value for header %q: %s", fld, err )
        return nil
    }
    return id
}

func ( s *server ) getCodecId( req *request ) *mg.Identifier {
    return s.getValAsId( idCodecId, req )
}

func ( s *server ) getCodecFactory( 
    spec *testing.TestSpec, req *request ) testing.CodecFactory {
    cdcId := s.getCodecId( req )
    if cdcId == nil { return nil }
    cdcFact := eng.GetCodecFactory( cdcId )
    if cdcFact == nil { s.respondErrorf( "No such codec: %s", cdcId ) }
    return cdcFact
}

func ( s *server ) getCodec( 
    spec *testing.TestSpec, req *request ) codec.Codec {
    cdcFact := s.getCodecFactory( spec, req )
    if cdcFact == nil { return nil }
    return cdcFact( spec.Headers() )
}

func ( s *server ) getSpecKey( req *request ) string {
    key, err := req.GetGoStringById( idSpecKey )
    if err == nil { return key }
    s.respondErrorf( "For key %s: %s", idSpecKey, err )
    return ""
}

func ( s *server ) getSpec( req *request ) *testing.TestSpec {
    specKey := s.getSpecKey( req )
    if specKey == "" { return nil }
    spec := eng.GetTestSpec( specKey )
    if spec == nil { s.respondErrorf( "Unknown spec: %s", specKey ) }
    return spec
}

func ( s *server ) handleGetSpecKeys( req *request ) {
    cdcId := s.getCodecId( req )
    if cdcId == nil { return }
    keys := eng.SpecKeysFor( cdcId )
    str := strings.Join( keys, "," )
    s.respond( mgio.MustHeadersPairs(), []byte( str ) )
}

func ( s *server ) handleGetSpec( req *request ) {
    spec := s.getSpec( req )
    if spec == nil { return }
    cdcFact := s.getCodecFactory( spec, req )
    if cdcFact == nil { return }
    buf, err := spec.GetBytes( cdcFact )
    if err != nil { log.Fatal( err ) }
    s.respond( mgio.MustHeadersPairs(), buf )
}

type serverFailer struct { 
    srv *server 
    failed bool
}

func ( sf *serverFailer ) Fatal( args ...interface{} ) {
    sf.failed = true
    sf.srv.respondErrorf( "Check failed: %s", fmt.Sprint( args... ) )
}

type encodeCheck struct {
    cdc codec.Codec
    buf []byte
    srv *server
    spec *testing.TestSpec
}

func ( chk *encodeCheck ) run( sf *serverFailer, f func() ) {
    f()
    if ! sf.failed { chk.srv.respondEmpty() }
}

func ( chk *encodeCheck ) checkRoundTrip() {
    if act, err := codec.DecodeBytes( chk.cdc, chk.buf ); err == nil {
        sf := &serverFailer{ srv: chk.srv }
        expct := chk.spec.Action.( *testing.RoundTrip ).Object
        chk.run( sf, func() { mgTesting.LossyEqual( expct, act, sf ) } )
    } else { chk.srv.respondErrorf( "Bad decode: %s", err.Error() ) }
}

func ( chk *encodeCheck ) checkEncodeValue() {
    sf := &serverFailer{ srv: chk.srv }
    ev := chk.spec.Action.( *testing.EncodeValue )
    a := &assert.Asserter{ sf }
    ce := &testing.CheckableEncode{ chk.buf, a }
    chk.run( sf, func() { ev.Check( ce ) } )
}

func ( s *server ) handleCheckEncode( req *request, buf []byte ) {
    spec := s.getSpec( req )
    if spec == nil { return }
    cdc := s.getCodec( spec, req )
    if cdc == nil { return }
    chk := &encodeCheck{ cdc: cdc, buf: buf, srv: s, spec: spec }
    switch spec.Action.( type ) {
    case *testing.RoundTrip: chk.checkRoundTrip()
    case *testing.EncodeValue: chk.checkEncodeValue()
    default: 
        tmpl := "Don't know how to check encode for spec %s"
        s.respondErrorf( tmpl, spec.KeyString() )
    }
}

func ( s *server ) commandFor( req *request ) *mg.Identifier {
    return s.getValAsId( idCommand, req )
}

func ( s *server ) handleMessage() bool {
    if hdrs, body, err := stream.ReadMessageBytes( s.conn ); err == nil {
        req := &request{ headersAccessorFor( hdrs ) }
        switch cmd := s.commandFor( req ); cmd != nil {
        case cmd.Equals( cmdGetSpecKeys ): s.handleGetSpecKeys( req )
        case cmd.Equals( cmdGetSpec ): s.handleGetSpec( req )
        case cmd.Equals( cmdCheckEncode ): s.handleCheckEncode( req, body )
        case cmd.Equals( cmdClose ): 
            s.respondEmpty()
            return false
        default: s.respondErrorf( "Unrecognized command: %s", cmd )
        }
    } else { log.Fatal( err ) }
    return true
}

func ( s *server ) serve() { 
    log.Printf( "Starting server" )
    for run := true; run; run = s.handleMessage() {}
}

func main() {
    conn := stream.NewConnection( os.Stdin, os.Stdout )
    (&server{ conn }).serve()
    log.Printf( "server done; exiting" )
}
