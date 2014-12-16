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
    putVal( "core-simple-pointers-inst1", func() interface{} {
        res := v1Data.NewCoreSimplePointers()
        bl := new( bool )
        *bl = true
        res.SetBoolF1( bl )
        buf := new( []byte )
        *buf = []byte{ 0, 1 }
        res.SetBufferF1( buf )
        i32 := new( int32 )
        *i32 = 1
        res.SetInt32F1( i32 )
        i64 := new( int64 )
        *i64 = 2
        res.SetInt64F1( i64 )
        ui32 := new( uint32 )
        *ui32 = 3
        res.SetUint32F1( ui32 )
        ui64 := new( uint64 )
        *ui64 = 4
        res.SetUint64F1( ui64 )
        fl32 := new( float32 )
        *fl32 = 5.0
        res.SetFloat32F1( fl32 )
        fl64 := new( float64 )
        *fl64 = 6.0
        res.SetFloat64F1( fl64 )
        s := new( string )
        *s = "hello"
        res.SetStringF1( s )
        t := new( time.Time )
        *t = time.Time( mgTck.Timestamp1 )
        res.SetTimeF1( t )
        return res
    })
    putVal( "value-holder-int32", func() interface{} {
        res := v1Data.NewValueHolder()
        res.SetValF1( int32( 1 ) )
        return res
    })
    putVal( "value-holder-map1", func() interface{} {
        res := v1Data.NewValueHolder()
        res.SetValF1( map[ string ]interface{}{ "k1": int32( 1 ) } )
        return res
    })
    putVal( "value-holder-struct1-inst1", func() interface{} {
        res := v1Data.NewValueHolder()
        s1 := v1Data.NewStruct1()
        s1.SetF1( 1 )
        s1.SetF2( "hello" )
        res.SetValF1( s1 )
        return res
    })
    putVal( "map-holder-inst1", func() interface{} {
        res := v1Data.NewMapHolder()
        res.SetMapF1( map[ string ] interface{} { "f1": int32( 1 ) } )
        return res
    })
    putVal( "map-holder-inst2", func() interface{} {
        res := v1Data.NewMapHolder()
        res.SetMapF1( map[ string ] interface{} { "f1": int32( 1 ) } )
        res.SetMapF2( map[ string ] interface{} { "f2": "string-val" } )
        res.SetMapF3( &( map[ string ] interface{} { "f2": true } ) )
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
