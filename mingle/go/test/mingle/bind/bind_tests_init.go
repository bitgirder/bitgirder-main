package bind

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "mingle/parser"
    "bitgirder/objpath"
)

var stdBindTests = []interface{}{}

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

// these are meant to exercise various of the bind reactor itself, particularly
// in terms of its error handling and paths
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

func initFunctionBinderTests() {
    add := func( in mg.Value, expct interface{}, bf BinderFactory ) {
        stdBindTests = append( stdBindTests,
            &BinderImplTest{
                In: in,
                Expect: expct,
                Factory: bf,
            },
        )
    }
    addErr := func( in mg.Value, err error, bf BinderFactory ) {
        stdBindTests = append( stdBindTests,
            &BinderImplTest{
                In: in,
                Error: err,
                Factory: bf,
            },
        )
    }
    add(
        mg.Int32( int32( 1 ) ),
        int32( 1 ),
        &FunctionBinderFactory{
            Value: func( ve *mgRct.ValueEvent ) ( interface{}, error ) {
                return DefaultBindingForValue( ve.Val ), nil
            },
        },
    )
    add(
        mg.Int32( int32( 1 ) ),
        int32( 1 ),
        &FunctionBinderFactory{ 
            Value: AsSequentialValueFunction( DefaultValueBinderFunction ),
        },
    )
    add(
        mg.Int32( int32( 1 ) ),
        int32( -1 ),
        &FunctionBinderFactory{
            Value: AsSequentialValueFunction(
                func( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
                    return - int32( ve.Val.( mg.Int32 ) ), nil, true
                },
                DefaultValueBinderFunction,
            ),
        },
    )
//    add(
//        parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
//        S1{ f1: 1 },
//        &FunctionBinderFactory{}.
//            SetStructBinder(
//                NewFunctionFieldSetBinder().
//                Create( func() interface{} { return S1{} } ).
//                Field( 
//                    mkId( "f1" ),
//                    DefaultBinderFactory,
//                    func( 
//                        val interface{}, 
//                        path objpath.PathNode,
//                        obj interface{} ) ( interface{}, error ) {
// 
//                        res := obj.( S1 )
//                        res.f1 = val.( int32 )
//                        return res, nil
//                    },
//                ).
//                Validate(),
//            ),
//        },
//    )
    addErr(
        mg.MustList( asTyp( "ns1@v1/S1*" ) ),
        NewBindError( nil, "unhandled value: ns1@v1/S1*" ),
        &FunctionBinderFactory{},
    )
    addErr(
        mg.EmptySymbolMap(),
        NewBindError( nil, "unhandled value: mingle:core@v1/SymbolMap" ),
        &FunctionBinderFactory{},
    )
    addErr(
        parser.MustStruct( "ns1@v1/S1" ),
        NewBindError( nil, "unhandled value: ns1@v1/S1" ),
        &FunctionBinderFactory{},
    )
    addErr(
        mg.Int32( int32( 1 ) ),
        NewBindError( nil, "unhandled value: mingle:core@v1/Int32" ),
        &FunctionBinderFactory{},
    )
    addErr(
        mg.Int32( int32( 1 ) ),
        NewBindError( nil, "unhandled value: mingle:core@v1/Int32" ),
        &FunctionBinderFactory{
            Value: AsSequentialValueFunction(
                func( ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {
                    return nil, nil, false
                },
            ),
        },
    )
}

func init() {
    initDefaultValBindTests()
    initCustomValBindTests()
    initFunctionBinderTests()
}
