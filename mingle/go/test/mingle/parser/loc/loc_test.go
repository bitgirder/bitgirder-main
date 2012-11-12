package loc

import (
    "bitgirder/assert"
    "testing"
)

func TestParseErrorDup( t *testing.T ) {
    loc1 := &Location{ Line: 1, Col: 1, Source: "s1" }
    loc2 := loc1.Dup()
    assert.Equal( loc1.Line, loc2.Line )
    assert.Equal( loc1.Col, loc2.Col )
    assert.Equal( loc1.Source, loc2.Source )
    loc1.Line = 2
    assert.Equal( 1, loc2.Line ) // make sure loc2 is a copy, not ptr to *loc1
}

func TestLocationFormatting( t *testing.T ) {
    expct := "[test-source, line 1, col 7]"
    loc := &Location{ Line: 1, Col: 7, Source: "test-source" }
    assert.Equal( expct, loc.String() )
    assert.Equal( expct, (*loc).String() )
}

func TestParseErrorFormatting( t *testing.T ) {
    err := ParseError{
        Message: "test-message",
        Loc: &Location{ Line: 2, Col: 3, Source: "test-source" },
    }
    assert.Equal( "[test-source, line 2, col 3]: test-message", err.Error() )
}
