package hashes

import (
    "testing"
    "bitgirder/assert"
    "bytes"
    "io"
    "fmt"
    "crypto/md5"
)

const testString = "Golang"

// Hashes below were generated using openssl commands such as:
//
//  $ printf "Golang" | openssl dgst -md5 -hex
//
// This is not to say that we don't trust the Go implementation, only that we
// don't want to mask errors in our own potentially improper calling of the Go
// hash libraries, and so assert against an external reference implementation

const testMd5Str = "eda14e187a768b38eda999457c9cca1e"
var testMd5Buf []byte

const testSha1Str = "92f4645da18d229c1e99797e4ee0203092e4c3ed"
var testSha1Buf []byte

const testSha256Str =
    "50e56e797c4ac89f7994a37480fce29a8a0f0f123a695e2dc32d5632197e2318"
var testSha256Buf []byte

func hexToBuf( str string ) []byte {
    res := make( []byte, len( str ) / 2 )
    fmtStr := fmt.Sprintf( "%%%dx", len( str ) )
    if _, err := fmt.Sscanf( str, fmtStr, &res ); err != nil { panic( err ) }
    return res
}

func init() {
    testMd5Buf = hexToBuf( testMd5Str )
    testSha1Buf = hexToBuf( testSha1Str )
    testSha256Buf = hexToBuf( testSha256Str )
}

func readerFor( s string ) io.Reader { return bytes.NewBuffer( []byte( s ) ) }

type hashCall func() ( []byte )

//func assertHashes( expct []byte, calls ...func() ( []byte, error ) ) {
func assertHashes( expct []byte, calls ...hashCall ) {
    for _, call := range calls { assert.Equal( expct, call() ) }
}

func TestHashMethods( t *testing.T ) {
    h := md5.New()
    assertHashes( testMd5Buf,
        func() []byte { return HashOfBytes( h, []byte( testString ) ) },
        func() []byte { return HashOfString( h, testString ) },
        func() []byte { 
            return MustHash( HashOfReader( h, readerFor( testString ) ) )
        },
        func() []byte { return MustHashOfReader( h, readerFor( testString ) ) },
    )
}

func TestMd5( t *testing.T ) {
    assertHashes( testMd5Buf,
        func() []byte { return Md5OfBytes( []byte( testString ) ) },
        func() []byte { return Md5OfString( testString ) },
        func() []byte { 
            return MustHash( Md5OfReader( readerFor( testString ) ) )
        },
        func() []byte { return MustMd5OfReader( readerFor( testString ) ) },
    )
}

func TestSha1( t *testing.T ) {
    assertHashes( testSha1Buf,
        func() []byte { return Sha1OfBytes( []byte( testString ) ) },
        func() []byte { return Sha1OfString( testString ) },
        func() []byte { 
            return MustHash( Sha1OfReader( readerFor( testString ) ) )
        },
        func() []byte { return MustSha1OfReader( readerFor( testString ) ) },
    )
}

func TestSha256( t *testing.T ) {
    assertHashes( testSha256Buf,
        func() []byte { return Sha256OfBytes( []byte( testString ) ) },
        func() []byte { return Sha256OfString( testString ) },
        func() []byte { 
            return MustHash( Sha256OfReader( readerFor( testString ) ) )
        },
        func() []byte { return MustSha256OfReader( readerFor( testString ) ) },
    )
}
