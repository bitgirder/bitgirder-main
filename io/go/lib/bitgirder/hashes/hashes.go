package hashes

import (
    "crypto/md5"
    "crypto/sha1"
    "crypto/sha256"
    "hash"
    "io"
)

// In HashOf*, h will be Reset() upon entry but not upon exit

// Util function to expect and return the non-error result of a multi-value
// return
func MustHash( dig []byte, err error ) []byte {
    if err != nil { panic( err ) }
    return dig
}

// Returns any error encountered while reading from r
func HashOfReader( h hash.Hash, r io.Reader ) ( []byte, error ) {
    h.Reset()
    if _, err := io.Copy( h, r ); err != nil { return nil, err }
    return h.Sum( nil ), nil
}

func MustHashOfReader( h hash.Hash, r io.Reader ) []byte {
    return MustHash( HashOfReader( h, r ) )
}

// unlike HashOfReader, in which there could legitimately be an error reading
// from the input unrelated to the hash itself, this function panics on err
// since the caller is, by calling these methods, asserting that the input is
// valid
func HashOfBytes( h hash.Hash, b []byte ) []byte {
    h.Reset()
    if _, err := h.Write( b ); err != nil { panic( err ) }
    return h.Sum( nil )
}

func HashOfString( h hash.Hash, s string ) []byte {
    return HashOfBytes( h, []byte( s ) )
}

func Md5OfReader( r io.Reader ) ( []byte, error ) {
    return HashOfReader( md5.New(), r )
}

func MustMd5OfReader( r io.Reader ) []byte {
    return MustHashOfReader( md5.New(), r )
}

func Md5OfBytes( b []byte ) []byte { return HashOfBytes( md5.New(), b ) }
func Md5OfString( s string ) []byte { return HashOfString( md5.New(), s ) }

func Sha1OfReader( r io.Reader ) ( []byte, error ) {
    return HashOfReader( sha1.New(), r )
}

func MustSha1OfReader( r io.Reader ) []byte {
    return MustHashOfReader( sha1.New(), r )
}

func Sha1OfBytes( b []byte ) []byte { return HashOfBytes( sha1.New(), b ) }
func Sha1OfString( s string ) []byte { return HashOfString( sha1.New(), s ) }

func Sha256OfReader( r io.Reader ) ( []byte, error ) {
    return HashOfReader( sha256.New(), r )
}

func MustSha256OfReader( r io.Reader ) []byte {
    return MustHashOfReader( sha256.New(), r )
}

func Sha256OfBytes( b []byte ) []byte { return HashOfBytes( sha256.New(), b ) }

func Sha256OfString( s string ) []byte {
    return HashOfString( sha256.New(), s )
}
