package stream

import (
    "testing"
    "log"
    "errors"
    "io"
    "bytes"
//    "time"
    mg "mingle"
    mgio "mingle/io"
    "bitgirder/hashes"
    "bitgirder/assert"
    "crypto/rand"
)

var idCommand = mg.MustIdentifier( "command" )

func nextBody( sz int ) ( []byte, []byte ) {
    res := make( []byte, sz )
    if _, err := rand.Read( res ); err != nil { panic( err ) }
    return res, hashes.Md5OfBytes( res )
}

type server struct {
    *testing.T
    conn Connection
}

func ( s *server ) respond( hdrs *mgio.Headers, body []byte ) {
    if body == nil { body = []byte{} }
    if err := WriteMessageBytes( s.conn, hdrs, body ); err != nil {
        s.Fatal( err )
    }
}

func ( s *server ) handleHashBody( req *mgio.Headers, body []byte ) {
    h := hashes.Md5OfBytes( body )
    hdrs := mgio.MustHeadersPairs()
    s.respond( hdrs, h )
}

func ( s *server ) handleEchoBody( req *mgio.Headers, body []byte ) {
    s.respond( mgio.MustHeadersPairs(), body )
}

func ( s *server ) handle( req *mgio.Headers, body []byte ) {
    acc := mg.NewSymbolMapAccessor( req.Fields(), nil )
    switch cmd := acc.MustGoStringById( idCommand ); cmd {
    case "hash-body": s.handleHashBody( req, body )
    case "echo-body": s.handleEchoBody( req, body )
    default: panic( "Unknown command: " + cmd )
    }
}

func ( s *server ) serve() {
    for {
        if req, body, err := ReadMessageBytes( s.conn ); err == nil {
            s.handle( req, body )
        } else { s.Fatal( err ) }
    }
}

func TestBasicExchanges( t *testing.T ) {
    cli, srv := OpenPipe()
    go (&server{ T: t, conn: srv }).serve()
    for i := 0; i < 10; i++ {
        body, h := nextBody( 100 )
        hdrs := mgio.MustHeadersPairs( "command", "hash-body" )
        if err := WriteMessageBytes( cli, hdrs, body ); err != nil { 
            t.Fatal( err ) 
        }
        if _, respBody, err := ReadMessageBytes( cli ); err == nil {
            assert.Equal( h, respBody )
        } else { log.Print( err ) }
    }
}

func assertEcho( cli Connection, buf []byte, t *testing.T ) {
    hdrs := mgio.MustHeadersPairs( "command", "echo-body" )
    var err error
    if buf == nil {
        err = WriteMessageNoBody( cli, hdrs )
    } else { err = WriteMessageBytes( cli, hdrs, buf ) }
    if err != nil { t.Fatal( err ) }
    if _, respBody, err := ReadMessageBytes( cli ); err == nil {
        assert.Equal( buf, respBody )
    } else { t.Fatal( err ) }
}

func TestEmptyMessage( t *testing.T ) {
    cli, srv := OpenPipe()
    go (&server{ T: t, conn: srv }).serve()
    assertEcho( cli, []byte{}, t )
}

func failMessageRead( hdrs *mgio.Headers, sz int64, r io.Reader ) error {
    return errors.New( "message-read-failer" )
}

func TestInvalidMessageVersionError( t *testing.T ) {
    buf := &bytes.Buffer{}
    if err := mgio.WriteBinary( int32( 30 ), buf ); err != nil { 
        t.Fatal( err ) 
    }
    if err := readMessage( buf, failMessageRead ); err == nil {
        t.Fatal( "Expected error" )
    } else if mve, ok := err.( *mgio.InvalidVersionError ); ok {
        msg := `Invalid message version: 0x0000001e (expected 0x00000001)`
        assert.Equal( msg, mve.Error() )
    } else { t.Fatal( err ) }
} 

func TestInvalidTypeCodeError( t *testing.T ) {
    buf := &bytes.Buffer{}
    if err := mgio.WriteBinary( messageVersion1, buf ); err != nil { 
        t.Fatal( err ) 
    }
    if err := mgio.WriteBinary( int32( 12 ), buf ); err != nil { 
        t.Fatal( err ) 
    }
    if err := readMessage( buf, failMessageRead ); err == nil {
        t.Fatal( "Expected error" )
    } else if tce, ok := err.( *mgio.InvalidTypeCodeError ); ok {
        msg := `Invalid type code: 0x0000000c (expected 0x00000001)`
        assert.Equal( msg, tce.Error() )
    } else { t.Fatal( err ) }
}
