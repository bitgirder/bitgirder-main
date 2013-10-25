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

// Don't retest all the semantics of casts -- just check that the CastValue()
// implementation on top of CastReactor correctly returns a success and error
// vals
func TestCastValueCalls( t *testing.T ) {
    p := objpath.RootedAt( id( "f1" ) )
    if v, err := CastValue( String( "1" ), TypeInt32, p ); err == nil {
        assert.Equal( Int32( 1 ), v )
    } else { t.Fatal( err ) }
    if _, err := CastValue( Int32( 1 ), TypeBuffer, p ); err == nil {
        t.Fatalf( "No err" )
    } else {
        if tc, ok := err.( *ValueCastError ); ok {
            assert.Equal( FormatIdPath( p ), FormatIdPath( tc.Location() ) )
            assert.Equal( 
                "Expected value of type mingle:core@v1/Buffer but found " +
                "mingle:core@v1/Int32",
                tc.Message() )
        } else { t.Fatal( err ) }
    }
}
