package main

import (
    "log"
    "os"
//    "fmt"
    bgio "bitgirder/io"
)

var rd *bgio.BinReader
var wr *bgio.BinWriter

type responseCode int8

const (
    rcPassed = responseCode( int8( 0 ) )
    rcFailed = responseCode( int8( 1 ) )
)

type testSpec struct {
    name string
    buffer []byte
}

func readNext() ( res *testSpec, err error ) {
    res = &testSpec{}
    if res.name, err = rd.ReadUtf8(); err != nil { return }
    if res.buffer, err = rd.ReadBuffer32(); err != nil { return }
    return
}

func writeResponse( err error ) error {
    if err == nil { return wr.WriteInt8( int8( rcPassed ) ) }
    if ioErr := wr.WriteInt8( int8( rcFailed ) ); ioErr != nil { return ioErr }
    return wr.WriteUtf8( err.Error() )
}

func checkNext() ( error, bool ) {
    test, err := readNext()
    if err != nil { return err, false }
    if test == nil { return nil, false }
    return writeResponse( nil ), true
}

func main() {
    var err error
    rd = bgio.NewLeReader( os.Stdin )
    defer rd.Close()
    wr = bgio.NewLeWriter( os.Stdout )
    defer wr.Close()
    ok := true
    for err == nil && ok { err, ok = checkNext() }
    if err != nil { log.Fatal( err ) }
}
