package tck

import ( 
    mg "mingle"
    "mingle/parser"
    "bitgirder/assert"
    "testing"
    cgTck "mingle/codegen/tck"
    "mingle/bind"
    v1Data "mingle/v1/tck/data"
    _ "mingle/v1/tck/data2"
)

var (
    mkId = parser.MustIdentifier
)

var boundValuesById = mg.NewIdentifierMap()

func init() {
    boundValuesById.Put( mkId( "empty-struct-inst1" ), v1Data.NewEmptyStruct() )
}

type bindTestCall struct {
    *assert.PathAsserter
    bt *bind.BindTest
}

func ( c *bindTestCall ) CreateReactors( t *bind.BindTest ) []interface{} {
    return []interface{}{}
}

func ( c *bindTestCall ) BoundValues() *mg.IdentifierMap {
    return boundValuesById
}

func ( c *bindTestCall ) call() { 
    c.Logf( "bt.BoundId: %s", c.bt.BoundId )
    btcc := &bind.BindTestCallControl{ Interface: c }
    bind.AssertBindTest( c.bt, btcc, c.PathAsserter ) 
}

func TestTck( t *testing.T ) {
    la := assert.NewListPathAsserter( t )
    tests := cgTck.GetTckTests()
    for _, test := range tests {
        switch v := test.( type ) {
        case *bind.BindTest: ( &bindTestCall{ bt: v, PathAsserter: la } ).call()
        default: la.Logf( "skipping: %T", test )
        }
        la = la.Next()
    }
}
