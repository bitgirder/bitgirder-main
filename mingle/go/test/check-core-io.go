package main

import (
    "log"
    "os"
    "fmt"
    "errors"
    bgio "bitgirder/io"
    mg "mingle"
    "bytes"
    "bitgirder/assert"
)

var rd *bgio.BinReader
var wr *bgio.BinWriter

var tests map[ string ]interface{}

type writeValueAsserter interface {
    AssertWriteValue( rd *mg.BinReader, a *assert.PathAsserter )
}

type responseCode int8

const (
    rcPassed = responseCode( int8( 0 ) )
    rcFailed = responseCode( int8( 1 ) )
)

type checkInstance struct {

    name string
    buffer []byte

    err error
}

func ( ci *checkInstance ) Fatal( args ...interface{} ) {
    ci.err = errors.New( fmt.Sprint( args... ) )
    panic( ci.err )
}

func ( ci *checkInstance ) assertWriteValue( wva writeValueAsserter ) {
    defer func() {
        if err := recover(); err != nil && err != ci.err { panic( err ) }
    }()
    a := assert.NewPathAsserter( ci )
    wva.AssertWriteValue( mg.NewReader( bytes.NewBuffer( ci.buffer ) ), a )
}

func ( ci *checkInstance ) getResponse() error {
    if test, ok := tests[ ci.name ]; ok {
        if wva, ok := test.( writeValueAsserter ); ok {
            ci.assertWriteValue( wva )
            return ci.err
        }
        return fmt.Errorf( 
            "don't know how to check values for test: %s", ci.name )
    }    
    return fmt.Errorf( "unrecognized test: %s", ci.name )
}

func initIo() func() {
    rd = bgio.NewLeReader( os.Stdin )
    wr = bgio.NewLeWriter( os.Stdout )
    return func() {
        log.Printf( "closing streams" )
        defer rd.Close()
        defer wr.Close()
    }
}

func readNext() ( res *checkInstance, err error ) {
    res = &checkInstance{}
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
    return writeResponse( test.getResponse() ), true
}

func runCheckLoop() {
    ok := true
    var err error
    for err == nil && ok { err, ok = checkNext() }
    if err != nil { log.Fatal( err ) }
}

func main() {
    defer ( initIo() )()
    tests = mg.CoreIoTestsByName()
    runCheckLoop()
    log.Printf( "checker exiting" )
}
