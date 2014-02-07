package mingle

import (
    "testing"
    "bitgirder/objpath"
    "bitgirder/assert"
)

func TestCastValueErrorFormatting( t *testing.T ) {
    path := objpath.RootedAt( id( "f1" ) )
    err := NewValueCastErrorf( path, "Blah %s", "X" )
    assert.Equal( "f1: Blah X", err.Error() )
}
