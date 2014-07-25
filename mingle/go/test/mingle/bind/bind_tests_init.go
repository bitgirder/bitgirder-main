package bind

import (
    mg "mingle"
    "mingle/parser"
    "bitgirder/objpath"
)

var stdBindTests = []*BindTest{}

var tm1 = mg.MustTimestamp( "2013-10-19T02:47:00-08:00" )

func initDefaultValBindTests() {
    add := func( in mg.Value, expct interface{} ) {
        stdBindTests = append( stdBindTests,
            &BindTest{ 
                In: in, 
                Expect: expct,
                Type: mg.TypeValue,
                Profile: TestProfileDefaultValue,
            },
        )
    }
    add( mg.Boolean( true ), true )
    add( mg.Buffer( []byte{ 0 } ), []byte{ 0 } )
    add( mg.String( "s" ), "s" )
    add( mg.Int32( 1 ), int32( 1 ) )
    add( mg.Int64( 1 ), int64( 1 ) )
    add( mg.Uint32( 1 ), uint32( 1 ) )
    add( mg.Uint64( 1 ), uint64( 1 ) )
    add( mg.Float32( 1.0 ), float32( 1.0 ) )
    add( mg.Float64( 1.0 ), float64( 1.0 ) )
    add( tm1, tm1 )
    add( 
        parser.MustSymbolMap( "f1", int32( 1 ) ),
        map[ string ]interface{}{ "f1": int32( 1 ) },
    )
    add( 
        parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
        map[ string ]interface{}{
            "$type": mkQn( "ns1@v1/S1" ),
            "f1": int32( 1 ),
        },
    )
    add(
        parser.MustEnum( "ns1@v1/E1", "v1" ),
        parser.MustEnum( "ns1@v1/E1", "v1" ),
    )
    add( mg.NullVal, mg.NullVal )
    add(
        mg.MustList( 
            int32( 1 ), 
            "a", 
            parser.MustEnum( "ns1@v1/E1", "v1" ),
            mg.MustList( int32( 2 ) ),
            parser.MustSymbolMap( "f1", int32( 1 ) ),
            parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
            mg.NullVal,
        ),
        []interface{}{
            int32( 1 ), 
            "a", 
            parser.MustEnum( "ns1@v1/E1", "v1" ),
            []interface{}{ int32( 2 ) },
            map[ string ]interface{}{ "f1": int32( 1 ) },
            map[ string ]interface{}{
                "$type": mkQn( "ns1@v1/S1" ),
                "f1": int32( 1 ),
            },
            mg.NullVal,
        },
    )
}

func initCustomValBindTests() {
    add := func( in mg.Value, expct interface{}, typ interface{} ) {
        stdBindTests = append( stdBindTests,
            &BindTest{
                In: in,
                Expect: expct,
                Type: asTyp( typ ),
                Profile: TestProfileCustomValue,
            },
        )
    }
    s1Val1 := parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) )
    goS1Val1 := S1{ f1: 1 }
    add( mg.Int32( int32( 1 ) ), int32( 1 ), "Int32" )
    add( parser.MustStruct( "ns1@v1/S1" ), S1{}, "ns1@v1/S1" )
    add( s1Val1, goS1Val1, "ns1@v1/S1" )
    add(
        mg.MustList( 
            asTyp( "ns1@v1/S1*" ), 
            parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
            parser.MustStruct( "ns1@v1/S1", "f1", int32( 2 ) ),
        ),
        []S1{ S1{ f1: 1 }, S1{ f1: 2 } },
        "ns1@v1/S1*",
    )
    addErr := func( in mg.Value, typ interface{}, err error ) {
        stdBindTests = append( stdBindTests,
            &BindTest{
                In: in,
                Type: asTyp( typ ),
                Profile: TestProfileCustomValue,
                Error: err,
            },
        )
    }
    addErr( 
        parser.MustStruct( "ns1@v1/S1", "f1", "bad-val" ),
        "ns1@v1/S1",
        NewBindError( objpath.RootedAt( mkId( "f1" ) ), testMsgErrorBadValue ),
    )
    addErr(
        mg.MustList(
            asTyp( "ns1@v1/S1*" ),
            parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
            parser.MustStruct( "ns1@v1/S1", "f1", "bad-val" ),
        ),
        "ns1@v1/S1*",
        NewBindError(
            objpath.RootedAtList().SetIndex( 1 ).Descend( mkId( "f1" ) ),
            testMsgErrorBadValue,
        ),
    )
    addErr(
        parser.MustStruct( "ns1@v1/S1", "f1", s1F1ValFailOnProduce ),
        "ns1@v1/S1",
        NewBindError( nil, testMsgErrorBadValue ),
    )
    addErr(
        mg.MustList( 
            asTyp( "ns1@v1/S1*" ), 
            parser.MustStruct( "ns1@v1/S1", "f1", s1F1ValFailOnProduce ),
        ),
        "ns1@v1/S1*",
        NewBindError( objpath.RootedAtList(), testMsgErrorBadValue ),
    ) 
    addErr(
        parser.MustStruct( "ns1@v1/S1", "f1", int32( -1 ) ),
        "ns1@v1/S1",
        NewBindError( objpath.RootedAt( mkId( "f1" ) ), testMsgErrorBadValue ),
    )
}

func init() {
    initDefaultValBindTests()
    initCustomValBindTests()
}
