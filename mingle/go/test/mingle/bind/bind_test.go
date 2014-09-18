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
}

func ( s *S1 ) VisitValue( vc VisitContext ) error {
    return visitF1Struct( "ns1@v1/S1", s.f1, vc )
}

type E1 string 

const (
    E1V1 = E1( "v1" )
    E1V2 = E1( "v2" )
)

func ( e E1 ) VisitValue( vc VisitContext ) error {
    me := parser.MustEnum( "ns1@v1/E1", string( e ) )
    ve := mgRct.NewValueEvent( me )
    return vc.Destination.ProcessEvent( ve )
}

type unregisteredType int

type failOnVisitType int

type customVisitable int32

func visitOkTestFunc( val interface{}, vc VisitContext ) ( error, bool ) {
    switch v := val.( type ) {
    case customVisitable: 
        return visitF1Struct( "ns1@v1/CustomVisitable", int32( v ), vc ), true
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
        mkQn( "ns1@v1/S1" ),
        CheckedStructFactory( 
            reg, 
            func() interface{} { return &S1{} },
            &CheckedFieldSetter{ 
                Field: mkId( "f1" ), 
                Type: mg.TypeInt32, 
                Assign: func( acc, val interface{} ) { 
                    acc.( *S1 ).f1 = val.( int32 ) 
                },
            },
        ),
    )
    reg.MustAddValue(
        mkQn( "ns1@v1/E1" ),
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
    res.Put( mkId( "s1-val1" ), &S1{ f1: 1 } )
    res.Put( mkId( "e1-val1" ), E1V1 )
    res.Put( mkId( "custom-visitable-val1" ), customVisitable( 1 ) )
    res.Put( mkId( "unregistered-type-val1" ), unregisteredType( 1 ) )
    res.Put( mkId( "fail-on-visit-type-val1" ), failOnVisitType( 1 ) )
    return res
}

func getDefaultValBindTests( tests []*BindTest ) []*BindTest {
    p := mg.MakeTestIdPath
    addTest := func( t *BindTest ) {
        t.Domain = domainPackageBindTest
        tests = append( tests, t )
    }
    addOk := func( in mg.Value, id string ) {
        addTest( &BindTest{ Mingle: in, BoundId: mkId( id ) } )
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
    s1V1 := parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) )
    e1V1 := parser.MustEnum( "ns1@v1/E1", "v1" )
    addOk( s1V1, "s1-val1" )
    addOk( e1V1, "e1-val1" )
    addTest(
        &BindTest{
            Mingle: parser.MustStruct( "ns1@v1/CustomVisitable", 
                "f1", int32( 1 ),
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
        parser.MustStruct( "ns1@v1/Bad" ),
        nil,
        "unhandled value: ns1@v1/Bad",
    )
    addInErr(
        parser.MustEnum( "ns1@v1/Bad", "e1" ),
        nil,
        "unhandled value: ns1@v1/Bad",
    )
    addInErr(
        parser.MustStruct( "ns1@v1/S1", "f1", int64( 1 ) ),
        p( 1 ),
        "unhandled value: mingle:core@v1/Int64",
    )
    addInErr(
        parser.MustStruct( "ns1@v1/S1",
            "f1", parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
        ),
        p( 1 ),
        "unhandled value: ns1@v1/S1",
    )
    addInErr(
        parser.MustStruct( "ns1@v1/S1",
            "f1", parser.MustEnum( "ns1@v1/E1", "v1" ),
        ),
        p( 1 ),
        "unhandled value: ns1@v1/E1",
    )
    addInErr(
        parser.MustStruct( "ns1@v1/S1", 
            "f1", mg.MustList( asType( "ns1@v1/S1*" ), "ns1@v1/S1*" ),
        ),
        p( 1 ),
        "unhandled value: ns1@v1/S1*",
    )
    addInErr(
        parser.MustStruct( "ns1@v1/S1", "f1", mg.EmptySymbolMap() ),
        p( 1 ),
        "unhandled value: mingle:core@v1/SymbolMap",
    )
    addInErr(
        mg.MustList( asType( "ns1@v1/Bad*" ) ),
        nil,
        "unhandled value: ns1@v1/Bad*",
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

func getBindTests() []*BindTest {
    return getDefaultValBindTests( []*BindTest{} )
}

type defaultBindTestCallInterface struct {
    boundVals *mg.IdentifierMap
}

func ( c defaultBindTestCallInterface ) BoundValues() *mg.IdentifierMap {
    return c.boundVals
}

func ( c defaultBindTestCallInterface ) CreateReactors( 
    _ *BindTest ) []interface{} {

    return []interface{}{}
}

func TestBind( t *testing.T ) {
    ensureTestBuilderFactories()
    iface := defaultBindTestCallInterface{ getDefaultValBindTestValues() }
    AssertBindTests( getBindTests(), iface, assert.NewPathAsserter( t ) )
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
