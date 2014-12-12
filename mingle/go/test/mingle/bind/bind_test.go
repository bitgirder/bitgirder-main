package bind

import (
    "testing"
    "bitgirder/assert"
    "bitgirder/objpath"
    mg "mingle"
    mgRct "mingle/reactor"
    "mingle/parser"
    "time"
    "fmt"
)

func visitF1Struct( nm string, f1 int32, vc VisitContext ) error {
    out := vc.Destination
    return mgRct.VisitValue( parser.MustStruct( nm, "f1", f1 ), out )
}

type S1 struct {
    f1 int32
    f2 []int32
}

func ( s *S1 ) VisitValue( vc VisitContext ) error {
    qn := mkQn( "mingle:bind@v1/S1" )
    f := func() error { 
        if err := VisitFieldValue( vc, mkId( "f1" ), s.f1 ); err != nil { 
            return err
        }
        return VisitFieldFunc( vc, mkId( "f2" ), func() error {
            lt := asType( "Int32*" ).( *mg.ListTypeReference )
            elt := func( i int ) interface{} { return s.f2[ i ] }
            return VisitListValue( vc, lt, len( s.f2 ), elt )
        })
    }
    return VisitStruct( vc, qn, f )
}

type E1 string 

const (
    E1V1 = E1( "v1" )
    E1V2 = E1( "v2" )
)

func ( e E1 ) VisitValue( vc VisitContext ) error {
    me := parser.MustEnum( "mingle:bind@v1/E1", string( e ) )
    ve := mgRct.NewValueEvent( me )
    return vc.Destination.ProcessEvent( ve )
}

type unregisteredType int

type failOnVisitType int

type customVisitable struct {
    f1 int32
    f2 []int32
}

func visitCustomVisitable( cv customVisitable, vc VisitContext ) error {
    qn := mkQn( "mingle:bind@v1/CustomVisitable" )
    return VisitStruct( vc, qn, func() error {
        err := VisitFieldFunc( vc, mkId( "f1" ), func() error {
            return vc.EventSender().Value( mg.Int32( int32( cv.f1 ) ) )
        })
        if err != nil { return err }
        lt := asType( "Int32*" ).( *mg.ListTypeReference )
        return VisitFieldFunc( vc, mkId( "f2" ), func() error {
            return VisitListFunc( vc, lt, len( cv.f2 ), func( i int ) error {
                return vc.EventSender().Value( mg.Int32( cv.f2[ i ] ) )
            })
        })
    })
}

func visitOkTestFunc( val interface{}, vc VisitContext ) ( error, bool ) {
    switch v := val.( type ) {
    case customVisitable: return visitCustomVisitable( v, vc ), true
    case failOnVisitType: return NewVisitError( vc.Path, "test-failure" ), true
    }
    return nil, false
}

// one-time guard for ensureTestBuilderFactories()
var didEnsureTestBuilderFactories = false

// we would otherwise do this in an init() block, except we don't want to deal
// with the possibility that this would run before the default domain itself is
// initialized (dependent packages won't have this concern)
func ensureTestBuilderFactories() {
    if didEnsureTestBuilderFactories { return }
    didEnsureTestBuilderFactories = true
    reg := NewRegistry()
    regsByDomain.Put( domainPackageBindTest, reg )
    addPrimBindings( reg )
    reg.MustAddValue(
        mkQn( "mingle:bind@v1/S1" ),
        CheckedStructFactory( 
            reg, 
            func() interface{} { return &S1{} },
            nil,
            &CheckedFieldSetter{ 
                Field: mkId( "f1" ), 
                Type: mg.TypeInt32, 
                Assign: func( acc, val interface{} ) { 
                    acc.( *S1 ).f1 = val.( int32 ) 
                },
            },
            &CheckedFieldSetter{
                Field: mkId( "f2" ),
                StartField: CheckedListFieldStarter(
                    func() interface{} { return make( []int32, 0, 4 ) },
                    ListElementFactoryFuncForType( mg.TypeInt32 ),
                    func( l, val interface{} ) interface{} {
                        return append( l.( []int32 ), val.( int32 ) )
                    },
                ),
                Assign: func( acc, val interface{} ) {
                    acc.( *S1 ).f2 = val.( []int32 )
                },
            },
        ),
    )
    reg.MustAddValue(
        mkQn( "mingle:bind@v1/E1" ),
        func() mgRct.BuilderFactory {
            res := mgRct.NewFunctionsBuilderFactory()
            res.ValueFunc = func( 
                ve *mgRct.ValueEvent ) ( interface{}, error, bool ) {

                if e, ok := ve.Val.( *mg.Enum ); ok {
                    return E1( e.Value.ExternalForm() ), nil, true
                }
                return nil, nil, false
            }
            return res
        }(),
    )
    reg.AddVisitValueOkFunc( visitOkTestFunc )
}

var (
    tm1 = mg.MustTimestamp( "2013-10-19T02:47:00-08:00" )
)

func getDefaultValBindTestValues() *mg.IdentifierMap {
    res := mg.NewIdentifierMap()
    res.Put( mkId( "null-val" ), nil )
    res.Put( mkId( "true-val" ), true )
    res.Put( mkId( "buffer-val1" ), []byte{ 0 } )
    res.Put( mkId( "string-val1" ), "s" )
    res.Put( mkId( "int32-val1" ), int32( 1 ) )
    res.Put( mkId( "int64-val1" ), int64( 1 ) )
    res.Put( mkId( "uint32-val1" ), uint32( 1 ) )
    res.Put( mkId( "uint64-val1" ), uint64( 1 ) )
    res.Put( mkId( "float32-val1" ), float32( 1.0 ) )
    res.Put( mkId( "float64-val1" ), float64( 1.0 ) )
    res.Put( mkId( "time-val1" ), time.Time( tm1 ) )
    s1V1Inst1 := &S1{ f1: 1, f2: []int32{ 0, 1, 2 } }
    res.Put( mkId( "s1-val1" ), s1V1Inst1 )
    res.Put( mkId( "e1-val1" ), E1V1 )
    res.Put( 
        mkId( "custom-visitable-val1" ), 
        customVisitable{ f1: 1, f2: []int32{ 0, 1, 2 } },
    )
    res.Put( mkId( "unregistered-type-val1" ), unregisteredType( 1 ) )
    res.Put( mkId( "fail-on-visit-type-val1" ), failOnVisitType( 1 ) )
    res.Put(
        mkId( "map-inst1" ),
        map[ string ] interface{} { "an-id1": int32( 1 ) },
    )
    res.Put(
        mkId( "opaque-out-ptr-map-inst1" ),
        func() interface{} {
            ip1 := new( *int32 )
            *ip1 = new( int32 )
            **ip1 = int32( 1 )
            m := &( map[ string ] interface{}{ "anId1": ip1 } )
            return m
        }(),
    )
    res.Put( mkId( "int32-list1" ), []int32{ 0, 1 } )
    res.Put(
        mkId( "val-list-inst1" ),
        []interface{}{
            int32( 1 ),
            s1V1Inst1,
            E1V1,
            []interface{}{
                []interface{}{
                    int32( 1 ),
                    map[ string ] interface{}{ "k1": true },
                },
            },
            map[ string ] interface{}{
                "k1": []interface{}{},
                "k2": "hello",
                "k3": map[ string ] interface{}{},
            },
        },
    )
    res.Put(
        mkId( "opaque-out-ptr-val-list-inst1" ),
        func() interface{} {
            res := new( []interface{} )
            iPtr := new( int32 )
            *iPtr = int32( 1 )
            e1Ptr := new( E1 )
            *e1Ptr = E1V1
            l2Ptr := new( *[]interface{} )
            *l2Ptr = new( []interface{} )
            **l2Ptr = []interface{}{
                []interface{}{
                    int32( 1 ),
                    &( map[ string ] interface{}{ "k1": true } ),
                },
            }
            mp2Ptr := new( map[ string ] interface{} )
            *mp2Ptr = map[ string ] interface{}{
                "k1": []interface{}{},
                "k2": "hello",
                "k3": map[ string ] interface{}{},
            }
            *res = []interface{}{ iPtr, s1V1Inst1, e1Ptr, l2Ptr, mp2Ptr }
            return res
        }(),
    )
    res.Put( mkId( "bool-val-map" ), map[ bool ] interface{}{ true: true } )
    res.Put( 
        mkId( "string-int32-map" ), 
        map[ string ] int32 { "k1": int32( 1 ) },
    )
    res.Put( 
        mkId( "bool-val-map-at-nested-err-loc" ),
        map[ string ] interface{} {
            "p1": []interface{}{ 
                int32( 0 ), 
                int32( 1 ),
                map[ string ] int32 { "k1": int32( 1 ) },
            },
        },
    )
    res.Put( 
        mkId( "bad-id-key-map" ),
        map[ string ] interface{}{ "$bad-id": true }, 
    )
    return res
}

var mgS1V1 = parser.MustStruct( "mingle:bind@v1/S1", 
    "f1", int32( 1 ),
    "f2", mg.MustList( asType( "Int32*" ), 
        int32( 0 ), int32( 1 ), int32( 2 ),
    ),
)

var mgE1V1 = parser.MustEnum( "mingle:bind@v1/E1", "v1" )

func appendDefaultTestsForProfile(
    tests []*BindTest, prof string ) []*BindTest {
    
    p := mg.MakeTestIdPath
    addTest := func( t *BindTest ) {
        t.Domain = domainPackageBindTest
        t.Profile = prof
        tests = append( tests, t )
    }
    addOk := func( in mg.Value, id string ) {
        boundId := mkId( id )
        addTest( &BindTest{ Mingle: in, BoundId: boundId } )
        addTest( 
            &BindTest{ 
                Mingle: in, 
                BoundId: boundId, 
                Type: mg.TypeValue,
                StrictTypeMatching: true,
                Direction: BindTestDirectionIn,
            },
        )
    }
    addOk( mg.NullVal, "null-val" )
    addOk( mg.Boolean( true ), "true-val" )
    addOk( mg.Buffer( []byte{ 0 } ), "buffer-val1" )
    addOk( mg.String( "s" ), "string-val1" )
    addOk( mg.Int32( 1 ), "int32-val1" )
    addOk( mg.Int64( 1 ), "int64-val1" )
    addOk( mg.Uint32( 1 ), "uint32-val1" )
    addOk( mg.Uint64( 1 ), "uint64-val1" )
    addOk( mg.Float32( 1.0 ), "float32-val1" )
    addOk( mg.Float64( 1.0 ), "float64-val1" )
    addOk( tm1, "time-val1" )
    addOk( mgS1V1, "s1-val1" )
    addOk( mgE1V1, "e1-val1" )
    addTest(
        &BindTest{
            Mingle: parser.MustStruct( "mingle:bind@v1/CustomVisitable", 
                "f1", int32( 1 ),
                "f2", mg.MustList( asType( "Int32*" ), 
                    int32( 0 ), int32( 1 ), int32( 2 ),
                ),
            ),
            BoundId: mkId( "custom-visitable-val1" ),
            Direction: BindTestDirectionOut,
        },
    )
    addInErr := func( in mg.Value, path objpath.PathNode, msg string ) {
        addTest(
            &BindTest{
                Mingle: in,
                Error: NewBindError( path, msg ),
                Direction: BindTestDirectionIn,
            },
        )
    }
    addInErr(
        parser.MustStruct( "mingle:bind@v1/Bad" ),
        nil,
        "unhandled value: mingle:bind@v1/Bad",
    )
    addInErr(
        parser.MustEnum( "mingle:bind@v1/Bad", "e1" ),
        nil,
        "unhandled value: mingle:bind@v1/Bad",
    )
    addInErr(
        parser.MustStruct( "mingle:bind@v1/S1", "f1", int64( 1 ) ),
        p( 1 ),
        "unhandled value: mingle:core@v1/Int64",
    )
    addInErr(
        parser.MustStruct( "mingle:bind@v1/S1",
            "f1", parser.MustStruct( "mingle:bind@v1/S1", "f1", int32( 1 ) ),
        ),
        p( 1 ),
        "unhandled value: mingle:bind@v1/S1",
    )
    addInErr(
        parser.MustStruct( "mingle:bind@v1/S1",
            "f1", parser.MustEnum( "mingle:bind@v1/E1", "v1" ),
        ),
        p( 1 ),
        "unhandled value: mingle:bind@v1/E1",
    )
    addInErr(
        parser.MustStruct( "mingle:bind@v1/S1", 
            "f1", mg.MustList( asType( "mingle:bind@v1/S1*" ), 
                parser.MustStruct( "mingle:bind@v1/S1" ),
            ),
        ),
        p( 1 ),
        "unhandled value: mingle:bind@v1/S1*",
    )
    addInErr(
        parser.MustStruct( "mingle:bind@v1/S1", "f1", mg.EmptySymbolMap() ),
        p( 1 ),
        "unhandled value: mingle:core@v1/SymbolMap",
    )
    addVisitErr := func( boundId string, path objpath.PathNode, msg string ) {
        addTest(
            &BindTest{
                BoundId: mkId( boundId ),
                Direction: BindTestDirectionOut,
                Error: NewVisitError( path, msg ),
            },
        )
    }
    addVisitErr( 
        "unregistered-type-val1",
        nil,
        "unknown type for visit: bind.unregisteredType",
    )
    addVisitErr( "fail-on-visit-type-val1", nil, "test-failure" )
    return tests
}

func appendDefaultTests( tests []*BindTest ) []*BindTest {
    tests = appendDefaultTestsForProfile( tests, BindTestProfileDefault )
    tests = append( tests,
        &BindTest{ 
            Mingle: mg.MustList( asType( "mingle:bind@v1/Bad*" ) ),
            Error: NewBindError( nil, "unhandled value: mingle:bind@v1/Bad*" ),
            Direction: BindTestDirectionIn,
            Profile: BindTestProfileDefault,
            Domain: DomainDefault,
        },
    )
    return tests
}

func appendOpaqueTests( tests []*BindTest ) []*BindTest {
    tests = appendDefaultTestsForProfile( tests, BindTestProfileOpaque )
    addTestBase := func( t *BindTest ) {
        t.Domain = domainPackageBindTest
        t.Type = mg.TypeValue
        tests = append( tests, t )
    }
    addTest := func( in mg.Value, boundId string, dir BindTestDirection ) {
        addTestBase(
            &BindTest{ 
                Mingle: in, 
                BoundId: mkId( boundId ),
                Direction: dir,
                Profile: BindTestProfileOpaque,
            },
        )
    }
    addRt := func( in mg.Value, boundId string ) {
        addTest( in, boundId, BindTestDirectionRoundtrip )
    }
    addOut := func( in mg.Value, boundId string ) {
        addTest( in, boundId, BindTestDirectionOut )
    }
    addOk := func( in mg.Value, boundId string ) {
        addRt( in, boundId )
        addOut( in, "opaque-out-ptr-" + boundId )
    }
    addOutErr := func( boundId string, err error ) {
        addTestBase(
            &BindTest{
                BoundId: mkId( boundId ),
                Direction: BindTestDirectionOut,
                Profile: BindTestProfileOpaque,
                Error: err,
            },
        )
    }
    mgMapInst1 := parser.MustSymbolMap( "anId1", int32( 1 ) )
    addOk( mgMapInst1, "map-inst1" )
    addOk( 
        mg.MustList( 
            int32( 1 ), 
            mgS1V1, 
            mgE1V1,
            mg.MustList(
                mg.MustList( int32( 1 ), parser.MustSymbolMap( "k1", true ) ),
            ),
            parser.MustSymbolMap(
                "k1", mg.MustList(),
                "k2", "hello",
                "k3", mg.EmptySymbolMap(),
            ),
        ), 
        "val-list-inst1",
    )
    addOut( mg.MustList( int32( 0 ), int32( 1 ) ), "int32-list1" )
    addOutErr( 
        "bool-val-map", 
        NewVisitError( nil, "unknown type for visit: map[bool]interface {}" ),
    )
    addOutErr( 
        "string-int32-map",
        NewVisitError( nil, "unknown type for visit: map[string]int32" ),
    )
    addOutErr( 
        "bad-id-key-map",
        NewBindError( 
            nil, 
            "[<input>, line 1, col 1]: Illegal start of identifier part: \"$\" (U+0024)",
        ),
    )
    addOutErr(
        "bool-val-map-at-nested-err-loc",
        NewVisitError(
            objpath.RootedAt( mkId( "p1" ) ).StartList().SetIndex( 2 ),
            "unknown type for visit: map[string]int32",
        ),
    )
    return tests
}

func getDefaultValBindTests( tests []*BindTest ) []*BindTest {
    res := make( []*BindTest, 0, 32 )
    res = appendDefaultTests( res )
    res = appendOpaqueTests( res )
    return res
}

func getBindTests() []*BindTest {
    return getDefaultValBindTests( []*BindTest{} )
}

type defaultBindTestCallInterface struct {
    boundVals *mg.IdentifierMap
}

func ( c defaultBindTestCallInterface ) BoundValues() *mg.IdentifierMap {
    return c.boundVals
}

func TestBind( t *testing.T ) {
    ensureTestBuilderFactories()
    iface := defaultBindTestCallInterface{ getDefaultValBindTestValues() }
    cc := &BindTestCallControl{ Interface: iface }
    AssertBindTests( getBindTests(), cc, assert.NewPathAsserter( t ) )
}

func TestRegistryAccessors( t *testing.T ) {
    a := assert.NewPathAsserter( t )
    reg := NewRegistry()
    chkGetType := func( 
        key interface{}, expctOk bool, bf mgRct.BuilderFactory ) {

        ta := a.Descend( fmt.Sprint( a ) )
        var act mgRct.BuilderFactory
        var ok bool
        switch v := key.( type ) {
        case mg.TypeReference: act, ok = reg.BuilderFactoryForType( v )
        case *mg.QualifiedTypeName: act, ok = reg.BuilderFactoryForName( v )
        default: a.Fatalf( "unhandled key: %T", key )
        }
        ta.Descend( "ok" ).Equal( expctOk, ok )
        if ok { ta.Descend( "bf" ).Equal( bf, act ) }
    }
    strBld := mgRct.NewFunctionsBuilderFactory()
    reg.MustAddValue( mg.QnameString, strBld )
    chkGetType( mg.TypeString, true, strBld )
    chkGetType( mg.QnameString, true, strBld )
    chkGetType( asType( `String~"a"` ), true, strBld )
    chkGetType( mg.TypeInt32, false, nil )
    chkGetType( mg.QnameInt32, false, nil )
}
