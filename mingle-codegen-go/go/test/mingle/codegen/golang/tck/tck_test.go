package tck

import ( 
    "bitgirder/assert"
    "testing"
    cgTck "mingle/codegen/tck"
    _ "mingle/v1/tck/data"
    _ "mingle/v1/tck/data2"
)

func TestTck( t *testing.T ) {
    la := assert.NewListPathAsserter( t )
    tests := cgTck.GetTckTests()
    for _, test := range tests {
        la.Logf( "would test: %T", test )
        la = la.Next()
    }
}
