package parser

import (
    "testing"
    "bitgirder/assert"
    "fmt"
    mg "mingle"
)

func TestTimestampParse( t *testing.T ) {
    f := func( src, expct string ) {
        tm, err := ParseTimestamp( src )
        if err != nil { t.Fatal( err ) }
        if expct == "" { expct = src }
        assert.Equal( expct, tm.Rfc3339Nano() )
    }
    f( mg.Now().Rfc3339Nano(), "" )
    f( "2012-01-01T12:00:00Z", "" )
    f( "2012-01-01T12:00:00+07:00", "" )
    f( "2012-01-01T12:00:00-07:00", "" )
    f( "2012-01-01T12:00:00+00:00", "2012-01-01T12:00:00Z" )
}

func TestTimestampParseFail( t *testing.T ) {
    str := "2012-01-10X12:00:00"
    if _, err := ParseTimestamp( str ); err == nil {
        t.Fatalf( "Was able to parse: %q", str )
    } else {
        pe := err.( *ParseError )
        assert.Equal( 
            fmt.Sprintf( "Invalid RFC3339 time: %q", str ), pe.Message )
        assert.Equal( 1, pe.Loc.Line )
        assert.Equal( 1, pe.Loc.Col )
    }
}
