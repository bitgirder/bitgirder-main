package tck

import ( 
    "bitgirder/assert"
    "testing"
    cgTck "mingle/codegen/tck"
)

func TestTck( t *testing.T ) {
    la := assert.NewListPathAsserter( t )
    tests := cgTck.GetTckTests()
    for _, test := range tests {
        la.Logf( "would test: %T", test )
        la = la.Next()
    }
}
