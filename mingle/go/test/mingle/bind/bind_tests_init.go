package bind

import (
    mg "mingle"
    "mingle/parser"
    "bitgirder/objpath"
    "time"
)

var stdBindTests = []interface{}{}

var tm1 = mg.MustTimestamp( "2013-10-19T02:47:00-08:00" )

func initDefaultValBindTests() {
    p := mg.MakeTestIdPath
    add := func( in mg.Value, expct interface{} ) {
        stdBindTests = append( stdBindTests, 
            &BindTest{ 
                In: in, 
                Expect: expct,
                Domain: DomainDefault,
            },
        )
    }
    add( mg.NullVal, nil )
    add( mg.Boolean( true ), true )
    add( mg.Buffer( []byte{ 0 } ), []byte{ 0 } )
    add( mg.String( "s" ), "s" )
    add( mg.Int32( 1 ), int32( 1 ) )
    add( mg.Int64( 1 ), int64( 1 ) )
    add( mg.Uint32( 1 ), uint32( 1 ) )
    add( mg.Uint64( 1 ), uint64( 1 ) )
    add( mg.Float32( 1.0 ), float32( 1.0 ) )
    add( mg.Float64( 1.0 ), float64( 1.0 ) )
    add( tm1, time.Time( tm1 ) )
    s1V1 := parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) )
    e1V1 := parser.MustEnum( "ns1@v1/E1", "v1" )
    add( s1V1, S1{ f1: 1 } )
    add( e1V1, E1V1 )
    addErr := func( in mg.Value, path objpath.PathNode, msg string ) {
        stdBindTests = append( stdBindTests,
            &BindTest{
                In: in,
                Domain: DomainDefault,
                Error: NewBindError( path, msg ),
            },
        )
    }
    addErr(
        parser.MustStruct( "ns1@v1/Bad" ),
        nil,
        "unhandled value: ns1@v1/Bad",
    )
    addErr(
        parser.MustEnum( "ns1@v1/Bad", "e1" ),
        nil,
        "unhandled value: ns1@v1/Bad",
    )
    addErr(
        parser.MustStruct( "ns1@v1/S1", "f1", int64( 1 ) ),
        p( 1 ),
        "unhandled value: mingle:core@v1/Int64",
    )
    addErr(
        parser.MustStruct( "ns1@v1/S1",
            "f1", parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
        ),
        p( 1 ),
        "unhandled value: ns1@v1/S1",
    )
    addErr(
        parser.MustStruct( "ns1@v1/S1",
            "f1", parser.MustEnum( "ns1@v1/E1", "v1" ),
        ),
        p( 1 ),
        "unhandled value: ns1@v1/E1",
    )
    addErr(
        parser.MustStruct( "ns1@v1/S1", 
            "f1", mg.MustList( asType( "ns1@v1/S1*" ), "ns1@v1/S1*" ),
        ),
        p( 1 ),
        "unhandled value: ns1@v1/S1*",
    )
    addErr(
        parser.MustStruct( "ns1@v1/S1", "f1", mg.EmptySymbolMap() ),
        p( 1 ),
        "unhandled value: mingle:core@v1/SymbolMap",
    )
    addErr(
        mg.MustList( asType( "ns1@v1/Bad*" ) ),
        nil,
        "unhandled value: ns1@v1/Bad*",
    )
}

func init() {
    initDefaultValBindTests()
}
