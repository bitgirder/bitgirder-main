package uuid

import (
    "crypto/rand"
    "fmt"
    "io"
)

// Largely verbatim from Russ Cox in golang group:
// https://groups.google.com/d/msg/golang-nuts/owCogizIuZs/CQzU4nLdu14J. Taking
// definitive def to be this one: http://tools.ietf.org/rfc/rfc4122.txt, to
// which the algorithm below is faithful
func Type4() ( string, error ) {
    b := make( []byte, 16 )
    _, err := io.ReadFull( rand.Reader, b )
    if err != nil { return "", err }
    b[ 6 ] = ( b[ 6 ] & 0x0f ) | 0x40
    b[ 8 ] = ( b[ 8 ] &^ 0x40 ) | 0x80
    res := fmt.Sprintf( "%x-%x-%x-%x-%x", 
        b[ : 4 ], b[ 4 : 6 ], b[ 6 : 8 ], b[ 8 : 10 ], b[ 10 : ] )
    return res, nil
}

func MustType4() string {
    res, err := Type4()
    if err != nil { panic( err ) }
    return res
}
