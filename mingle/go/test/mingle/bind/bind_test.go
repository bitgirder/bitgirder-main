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
        func() mgRct.BuilderFactory {
            res := mgRct.NewFunctionsBuilderFactory()
            res.StructFunc = 
                func( _ *mgRct.StructStartEvent ) ( mgRct.FieldSetBuilder, 
                                                    error ) {

                    res := mgRct.NewFunctionsFieldSetBuilder()
                    res.Value = new( S1 )
                    res.RegisterField(
                        mkId( "f1" ),
                        func( path objpath.PathNode ) ( mgRct.BuilderFactory, 
                                                        error ) {
                            res, ok := reg.m.GetOk( mg.QnameInt32 )
                            if ok { return res.( mgRct.BuilderFactory ), nil }
                            return nil, nil
                        },
                        func( val interface{}, path objpath.PathNode ) error {
                            res.Value.( *S1 ).f1 = val.( int32 )
                            return nil
                        },
                    )
                    return res, nil
                }
            return res
        }(),
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

func getDefaultValBindTests( tests []*BindTest ) []*BindTest {
    p := mg.MakeTestIdPath
    addTest := func( t *BindTest ) {
        t.Domain = domainPackageBindTest
        tests = append( tests, t )
    }
    addOk := func( in mg.Value, expct interface{} ) {
        addTest( &BindTest{ Mingle: in, Bound: expct } )
    }
    addOk( mg.NullVal, nil )
    addOk( mg.Boolean( true ), true )
    addOk( mg.Buffer( []byte{ 0 } ), []byte{ 0 } )
    addOk( mg.String( "s" ), "s" )
    addOk( mg.Int32( 1 ), int32( 1 ) )
    addOk( mg.Int64( 1 ), int64( 1 ) )
    addOk( mg.Uint32( 1 ), uint32( 1 ) )
    addOk( mg.Uint64( 1 ), uint64( 1 ) )
    addOk( mg.Float32( 1.0 ), float32( 1.0 ) )
    addOk( mg.Float64( 1.0 ), float64( 1.0 ) )
    tm1 := mg.MustTimestamp( "2013-10-19T02:47:00-08:00" )
    addOk( tm1, time.Time( tm1 ) )
    s1V1 := parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) )
    e1V1 := parser.MustEnum( "ns1@v1/E1", "v1" )
    addOk( s1V1, &S1{ f1: 1 } )
    addOk( e1V1, E1V1 )
    addTest(
        &BindTest{
            Mingle: parser.MustStruct( "ns1@v1/CustomVisitable", 
                "f1", int32( 1 ),
            ),
            Bound: customVisitable( 1 ),
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
    addVisitErr := func( 
        bound interface{}, path objpath.PathNode, msg string ) {

        addTest(
            &BindTest{
                Bound: bound,
                Direction: BindTestDirectionOut,
                Error: NewVisitError( path, msg ),
            },
        )
    }
    addVisitErr( 
        unregisteredType( 1 ), 
        nil,
        "unknown type for visit: bind.unregisteredType",
    )
    addVisitErr( failOnVisitType( 1 ), nil, "test-failure" )
    return tests
}

func getBindTests() []*BindTest {
    return getDefaultValBindTests( []*BindTest{} )
}

type defaultBindTestCallInterface int

func ( c defaultBindTestCallInterface ) CreateReactors( 
    _ *BindTest ) []interface{} {

    return []interface{}{}
}

func TestBind( t *testing.T ) {
    ensureTestBuilderFactories()
    iface := defaultBindTestCallInterface( 1 )
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
