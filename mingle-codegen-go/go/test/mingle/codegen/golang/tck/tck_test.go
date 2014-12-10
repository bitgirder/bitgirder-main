package tck

import ( 
    mg "mingle"
    "mingle/parser"
    "bitgirder/assert"
    "testing"
    cgTck "mingle/codegen/tck"
    mgTck "mingle/tck"
    "mingle/bind"
    v1Data "mingle/v1/tck/data"
    "time"
)

var (
    mkId = parser.MustIdentifier
)

var boundValuesById = mg.NewIdentifierMap()

func init() {
    putVal := func( nm string, f func() interface{} ) {
        boundValuesById.Put( mkId( nm ), f() )
    }
    putVal( "empty-struct-inst1", func() interface{} { 
        return v1Data.NewEmptyStruct()
    })
    putVal( "scalars-basic-inst1", func() interface{} {
        res := v1Data.NewScalarsBasic()
        res.SetStringF1( "hello" )
        res.SetBoolF1( true )
        res.SetBufferF1( []byte{ 0, 1, 2 } )
        res.SetInt32F1( 1 )
        res.SetInt64F1( 2 )
        res.SetUint32F1( 3 )
        res.SetUint64F1( 4 )
        res.SetFloat32F1( 5.0 )
        res.SetFloat64F1( 6.0 )
        res.SetTimeF1( time.Time( mgTck.Timestamp1 ) )
        return res
    })
}

type bindTestCall struct {
    *assert.PathAsserter
    bt *bind.BindTest
}

func ( c *bindTestCall ) BoundValues() *mg.IdentifierMap {
    return boundValuesById
}

func ( c *bindTestCall ) call() { 
    if ! boundValuesById.HasKey( c.bt.BoundId ) { return }
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
