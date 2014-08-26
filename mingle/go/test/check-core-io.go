package main

import (
    "log"
    "os"
    "fmt"
    "errors"
    bgIo "bitgirder/io"
    mg "mingle"
    mgIo "mingle/io"
    "bytes"
    "sort"
    "strings"
    "bitgirder/assert"
)

var rd *bgIo.BinReader
var wr *bgIo.BinWriter

var tests map[ string ]interface{}

type writeValueAsserter interface {
    AssertWriteValue( rd *mg.BinReader, a *assert.PathAsserter )
}

type responseCode int8

const (
    rcPassed = responseCode( int8( 0 ) )
    rcFailed = responseCode( int8( 1 ) )
)

func dumpTestNames() {
    nms := make( []string, 0, len( tests ) )
    for nm, _ := range tests { nms = append( nms, nm ) }
    sort.Strings( nms )
    log.Printf( "known test names: %s", strings.Join( nms, ", " ) )
}

type checkInstance struct {

    name string
    buffer []byte

    err error
}

func ( ci *checkInstance ) Fatal( args ...interface{} ) {
    ci.err = errors.New( fmt.Sprint( args... ) )
    panic( ci.err )
}

func ( ci *checkInstance ) callWriteValueAssert( test interface{} ) {
    defer func() {
        if err := recover(); err != nil && err != ci.err { panic( err ) }
    }()
    a := assert.NewPathAsserter( ci )
    rd := mgIo.NewReader( bytes.NewBuffer( ci.buffer ) )
    switch v := test.( type ) {
    case *mg.BinIoRoundtripTest: mgIo.AssertRoundtripRead( v, rd, a )
    case *mg.BinIoSequenceRoundtripTest: 
        mgIo.AssertSequenceRoundtripRead( v, rd, a )
    default: ci.err = fmt.Errorf( "unhandled test type: %T", test )
    }
}

func ( ci *checkInstance ) getResponse() error {
    if test, ok := tests[ ci.name ]; ok {
        ci.callWriteValueAssert( test )
        return ci.err
    }    
    return fmt.Errorf( "unrecognized test: %s", ci.name )
}

func initIo() func() {
    rd = bgIo.NewLeReader( os.Stdin )
    wr = bgIo.NewLeWriter( os.Stdout )
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
    log.Printf( "core io checker starting" )
    defer ( initIo() )()
    tests = mg.CoreIoTestsByName()
//    dumpTestNames()
    runCheckLoop()
    log.Printf( "checker exiting" )
}
