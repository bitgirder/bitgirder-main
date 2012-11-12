package filerecv

import (
    "bitgirder/iotests"
    "bitgirder/hashes"
    "bitgirder/assert"
    "bytes"
    "io"
    "testing"
    "io/ioutil"
    "crypto/rand"
    "os"
)

func nextFile( t *testing.T ) *os.File {
    res, err := ioutil.TempFile( "", "filerecv-test-" )
    if err != nil { t.Fatal( err ) }
    return res
}

func nextSourceSha1( szBytes int, t *testing.T ) ( body, dig []byte ) {
    buf := make( []byte, szBytes )
    rand.Read( buf )
    return buf, hashes.Sha1OfBytes( buf )
}

func assertRecvDigest( dig []byte, f string, t *testing.T ) {
    if f, err := os.Open( f ); err == nil {
        defer f.Close()
        act := hashes.MustSha1OfReader( f )
        assert.Equal( dig, act )
    } else { t.Fatal( err ) }
}

func assertNotExists( nm string, t *testing.T ) {
    if _, err := os.Stat( nm ); err == nil {
        t.Fatalf( "File exists: %s", nm )
    } else {
        if ! os.IsNotExist( err ) { t.Fatal( err ) }
    }
}

const (
    tmpModeNone = iota
    tmpModeTmp
    tmpModeDest
)

const (
    failNot = iota
    failPreserve
    failClean
)

func setTempFile( 
    tmpMode int, recv *FileReceive, dest *os.File, t *testing.T ) *os.File {
    switch tmpMode {
    case tmpModeTmp: {
        tmp := nextFile( t )
        recv.TempFile = tmp.Name()
        return tmp
    }
    case tmpModeDest: recv.TempFile = dest.Name()
    case tmpModeNone:;
    default: t.Fatalf( "Unexpected tmpMode: %d", tmpMode )
    }
    return nil
}

type markerError struct {}

func ( e *markerError ) Error() string { return "Marker error" }

type failAfterSource struct { buf *bytes.Buffer }

func ( src *failAfterSource ) Read( p []byte ) ( int, error ) {
    if src.buf.Len() == 0 { return 0, &markerError{} }
    return src.buf.Read( p )
}

func makeFileSource( src []byte, failMode int ) io.Reader {
    buf := bytes.NewBuffer( src )
    if failMode == failNot { return buf }
    return &failAfterSource{ buf }
}

func assertFileReceiveErrorState(
    tmpMode, failMode int, recv *FileReceive, dig []byte, t *testing.T ) {
    var digFile string
    if tmpMode == tmpModeTmp {
        if failMode != failPreserve { assertNotExists( recv.DestFile, t ) }
        digFile = recv.TempFile
    } else { digFile = recv.DestFile }
    if failMode == failPreserve { 
        assertRecvDigest( dig, digFile, t )
    } else { assertNotExists( digFile, t ) }
}

func assertFileReceive( tmpMode, failMode int, t *testing.T ) {
    buf, dig := nextSourceSha1( 1000, t )
    var dest, tmp *os.File
    defer iotests.WipeEntries( dest, tmp )
    dest = nextFile( t )
    recv := &FileReceive{ DestFile: dest.Name() }
    recv.PreserveOnFail = failMode == failPreserve
    tmp = setTempFile( tmpMode, recv, dest, t )
    src := makeFileSource( buf, failMode )
    if sz, err := ReceiveFile( recv, src ); err == nil {
        assert.Equal( failNot, failMode )
        assert.Equal( int64( len( buf ) ), sz )
        assertRecvDigest( dig, dest.Name(), t )
        if tmpMode == tmpModeTmp { assertNotExists( tmp.Name(), t ) }
    } else if me, ok := err.( *markerError ); ok { 
        if failMode == failNot { t.Fatal( me ) }
        assertFileReceiveErrorState( tmpMode, failMode, recv, dig, t )
    } else { t.Fatal( err ) }
}

func TestFileReceiveSuccess( t *testing.T ) {
    for _, tmpMode := range []int { tmpModeNone, tmpModeDest, tmpModeTmp } {
    for _, failMode := range []int { failNot, failPreserve, failClean } {
        assertFileReceive( tmpMode, failMode, t )
    }}
}

func TestFileReceiveFailsMissingDestFile( t *testing. T ) {
    if _, err := ReceiveFile( &FileReceive{}, nil ); err != nil {
        assert.Equal( "FileReceive has no destination", err.Error() )
    } else { t.Fatal( "Expected error" ) }
}

func TestFileReceiveOverwritesPreviousTempFile( t *testing.T ) {
    var tmp, dest *os.File
    defer iotests.WipeEntries( tmp, dest )
    tmp, dest = nextFile( t ), nextFile( t )
    dest.Close()
    if _, err := tmp.Write( []byte( "first-body" ) ); err != nil { 
        t.Fatal( err )
    }
    recv := &FileReceive{ TempFile: tmp.Name(), DestFile: dest.Name() }
    src := bytes.NewBufferString( "new-body" )
    if _, err := ReceiveFile( recv, src ); err == nil {
        f, err := os.Open( dest.Name() )
        if err != nil { t.Fatal( err ) }
        expct := &bytes.Buffer{}
        if _, err := io.Copy( expct, f ); err != nil { t.Fatal( err ) }
        assert.Equal( "new-body", expct.String() )
    } else { t.Fatal( err ) }
}

func TestDottedTempName( t *testing.T ) {
    pairs := []string{
        "a", ".a",
        "/a", "/.a",
        "a/b/c", "a/b/.c",
        ".a", "..a",
        "a/.b", "a/..b",
        "/a/b/", "/a/.b",
    }
    for i, e := 0, len( pairs ); i < e; i += 2 {
        s, expct := pairs[ i ], pairs[ i + 1 ]
        assert.Equal( expct, ExpectDottedTempOf( s ) )
    }
    for _, s := range []string { "", "/", "//" } {
        if val, err := DottedTempOf( s ); err == nil {
            t.Fatalf( "Expected err on %#v but got %s", s, val )
        } else { assert.Equal( emptyPathErr, err ) }
    }
}
