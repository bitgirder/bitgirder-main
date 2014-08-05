package reactor

import (
    mg "mingle"
    "mingle/parser"
    "bitgirder/assert"
    "bitgirder/objpath"
    "testing"
//    "log"
)

var reactorTestNs *mg.Namespace

var mkQn = parser.MustQualifiedTypeName
var mkId = parser.MustIdentifier
var asType = parser.AsTypeReference

func listTypeRef( val interface{} ) *mg.ListTypeReference {
    return asType( val ).( *mg.ListTypeReference )
}

type ReactorTestCall struct {
    *assert.PathAsserter
}

type ReactorTest interface {
    Call( call *ReactorTestCall )
}

type NamedReactorTest interface { TestName() string }

type ReactorTestSetBuilder struct {
    tests []ReactorTest
}

func ( b *ReactorTestSetBuilder ) AddTests( t ...ReactorTest ) {
    b.tests = append( b.tests, t... )
}

type ReactorTestSetInitializer func( b *ReactorTestSetBuilder ) 

type testInitContext struct {
    ns *mg.Namespace
    ti ReactorTestSetInitializer
}

var testInits []testInitContext

func AddTestInitializer( ns *mg.Namespace, ti ReactorTestSetInitializer ) {
    if testInits == nil { testInits = make( []testInitContext, 0, 4 ) }
    testInits = append( testInits, testInitContext{ ns: ns, ti: ti } )
}

func getReactorTestsInNamespace( ns *mg.Namespace ) []ReactorTest {
    b := &ReactorTestSetBuilder{ tests: make( []ReactorTest, 0, 1024 ) }
    for _, ctx := range testInits { 
        if ctx.ns.Equals( ns ) { ctx.ti( b ) }
    }
    return b.tests
}

func RunReactorTestsInNamespace( ns *mg.Namespace, t *testing.T ) {
    a := assert.NewPathAsserter( t )
    la := a.StartList();
    for _, rt := range getReactorTestsInNamespace( ns ) {
        ta := la
        if nt, ok := rt.( NamedReactorTest ); ok { 
            ta = a.Descend( nt.TestName() ) 
        }
        c := &ReactorTestCall{ PathAsserter: ta }
        c.Logf( "calling %T", rt )
        rt.Call( c )
        la = la.Next()
    }
}

func EqualEvents( expct, act ReactorEvent, a *assert.PathAsserter ) {

    expct = CopyEvent( expct, true )
    act = CopyEvent( act, true )
    a.Equalf( expct, act, "events are not equal: %s != %s",
        EventToString( expct ), EventToString( act ) )
}

func flattenEvs( vals ...interface{} ) []ReactorEvent {
    res := make( []ReactorEvent, 0, len( vals ) )
    for _, val := range vals {
        switch v := val.( type ) {
        case ReactorEvent: res = append( res, v )
        case []ReactorEvent: res = append( res, v... )
        case []interface{}: res = append( res, flattenEvs( v... )... )
        default: panic( libErrorf( "Uhandled ev type for flatten: %T", v ) )
        }
    }
    return res
}

// to simplify test creation, we reuse event instances when constructing input
// event sequences, and send them to this method only at the end to ensure that
// we get a distinct sequence of event values for each test
func CopySource( evs []ReactorEvent ) []ReactorEvent {
    res := make( []ReactorEvent, len( evs ) )
    for i, ev := range evs { res[ i ] = CopyEvent( ev, false ) }
    return res
}

type reactorEventSource interface {
    Len() int
    EventAt( int ) ReactorEvent
}

func FeedEventSource( 
    src reactorEventSource, proc ReactorEventProcessor ) error {

    for i, e := 0, src.Len(); i < e; i++ {
        if err := proc.ProcessEvent( src.EventAt( i ) ); err != nil { 
            return err
        }
    }
    return nil
}

func AssertFeedEventSource(
    src reactorEventSource, proc ReactorEventProcessor, a assert.Failer ) {
    
    if err := FeedEventSource( src, proc ); err != nil { a.Fatal( err ) }
}

type eventSliceSource []ReactorEvent
func ( src eventSliceSource ) Len() int { return len( src ) }
func ( src eventSliceSource ) EventAt( i int ) ReactorEvent { return src[ i ] }

type eventExpectSource []EventExpectation

func ( src eventExpectSource ) Len() int { return len( src ) }

func ( src eventExpectSource ) EventAt( i int ) ReactorEvent {
    return CopyEvent( src[ i ].Event, true )
}

func FeedSource( src interface{}, rct ReactorEventProcessor ) error {
    switch v := src.( type ) {
    case reactorEventSource: return FeedEventSource( v, rct )
    case []ReactorEvent: return FeedSource( eventSliceSource( v ), rct )
    case mg.Value: return VisitValue( v, rct )
    }
    panic( libErrorf( "unhandled source: %T", src ) )
}

func AssertFeedSource( 
    src interface{}, rct ReactorEventProcessor, a assert.Failer ) {

    if err := FeedSource( src, rct ); err != nil { a.Fatal( err ) }
}

type eventPathCheckReactor struct {
    a *assert.PathAsserter
    eeAssert *assert.PathAsserter
    expct []EventExpectation
    idx int
}

func ( r *eventPathCheckReactor ) ProcessEvent( ev ReactorEvent ) error {
    r.a.Truef( r.idx < len( r.expct ), "unexpected event: %v", ev )
    ee := r.expct[ r.idx ]
    r.idx++
    ee.Event.SetPath( ee.Path )
    EqualEvents( ee.Event, ev, r.eeAssert )
    r.eeAssert = r.eeAssert.Next()
    return nil
}

func ( r *eventPathCheckReactor ) Complete() {
    r.a.Equalf( r.idx, len( r.expct ), "not all events were seen" )
}

func NewEventPathCheckReactor( 
    expct []EventExpectation, a *assert.PathAsserter ) *eventPathCheckReactor {

    return &eventPathCheckReactor{ 
        expct: expct, 
        a: a,
        eeAssert: a.Descend( "expct" ).StartList(),
    }
}

func nextListStart( lt *mg.ListTypeReference ) *ListStartEvent {
    return NewListStartEvent( lt )
}

func nextValueListStart() *ListStartEvent {
    return nextListStart( mg.TypeOpaqueList )
}

func nextMapStart() *MapStartEvent { return NewMapStartEvent() }

func CheckNoError( err error, c *ReactorTestCall ) {
    if err != nil { c.Fatalf( "Got no error but expected %T: %s", err, err ) }
}

func AssertCastError( expct, act error, pa *assert.PathAsserter ) {
    ca := mg.CastErrorAssert{ ErrExpect: expct, ErrAct: act, PathAsserter: pa }
    ca.Call()
}

type testError struct { 
    path objpath.PathNode
    msg string 
}

func ( e *testError ) Error() string { return mg.FormatError( e.path, e.msg ) }

func newTestError( path objpath.PathNode, msg string ) *testError {
    return &testError{ path: path, msg: msg }
}

func testErrForPath( p objpath.PathNode ) error {
    return newTestError( p, testMsgErrorBadValue )
}

func testErrForEvent( ev ReactorEvent ) error {
    return testErrForPath( ev.GetPath() )
}

func testErrForValue( v mg.Value, p objpath.PathNode ) error {
    if v == buildReactorErrorTestVal { return testErrForPath( p ) }
    return nil
}

type S1 struct {
    f1 int32
    f2 []int32
    f3 *S1
}

type S2 struct {}

func CheckBuiltValue( 
    expct mg.Value, br *BuildReactor, a *assert.PathAsserter ) {

    if expct == nil {
        if br != nil {
            act := br.GetValue().( mg.Value )
            a.Fatalf( "unexpected value: %s", mg.QuoteValue( act ) )
        }
    } else { 
        a.Falsef( br == nil, "expecting value %s but value builder is nil", 
            mg.QuoteValue( expct ) )
        mg.AssertEqualValues( expct, br.GetValue().( mg.Value ), a ) 
    }
}
