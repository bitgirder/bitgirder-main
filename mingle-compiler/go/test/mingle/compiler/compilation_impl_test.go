package compiler

import (
    "testing"
    "bitgirder/assert"
    "mingle/parser"
)

func TestDuplicateErrorsCondensed( t *testing.T ) {
    a := assert.NewPathAsserter( t )
    c := NewCompilation()
    c.addError( nil, "nil-err" )
    c.addError( nil, "nil-err" )
    lc := &parser.Location{ 1, 2, "s1" }
    c.addError( lc, "src-err" )
    c.addError( lc, "src-err" )
    cr := c.buildResult()    
    a.Equal( 2, len( cr.Errors ) )
    chk := func( msg string ) {
        for _, err := range cr.Errors { if err.Error() == msg { return } }
        a.Fatalf( "did not see error: %q", msg )
    }
    chk( "<nil>: nil-err" )
    chk( "[s1, line 1, col 2]: src-err" )
}
