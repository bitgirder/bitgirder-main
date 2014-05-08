package reactor

import (
    mg "mingle"
    "bitgirder/assert"
    "testing"
//    "log"
)

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

//func init() { testInits = make( []ReactorTestSetInitializer, 0, 4 ) }

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
//        c.Logf( "calling %T", rt )
        rt.Call( c )
        la = la.Next()
    }
}

func eventForEqualityCheck( 
    ev ReactorEvent, ignorePointerIds bool ) ReactorEvent {

    ev = CopyEvent( ev, true )
    switch v := ev.( type ) {
    case *ValueAllocationEvent: v.Id = mg.PointerIdNull
    case *ValueReferenceEvent: v.Id = mg.PointerIdNull
    case *MapStartEvent: v.Id = mg.PointerIdNull
    case *ListStartEvent: v.Id = mg.PointerIdNull
    }
    return ev
}

func EqualEvents( 
    expct, act ReactorEvent, ignorePointerIds bool, a *assert.PathAsserter ) {

    expct = eventForEqualityCheck( expct, ignorePointerIds )
    act = eventForEqualityCheck( act, ignorePointerIds )
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
    ignorePointerIds bool
}

func ( r *eventPathCheckReactor ) ProcessEvent( ev ReactorEvent ) error {
    r.a.Truef( r.idx < len( r.expct ), "unexpected event: %v", ev )
    ee := r.expct[ r.idx ]
    r.idx++
    ee.Event.SetPath( ee.Path )
    EqualEvents( ee.Event, ev, r.ignorePointerIds, r.eeAssert )
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

type heapTestValue struct {
    id mg.PointerId
    val mg.Value
}

func ( ht heapTestValue ) valImpl() {}
func ( ht heapTestValue ) Address() mg.PointerId { return ht.id }
func ( ht heapTestValue ) Dereference() mg.Value { return ht.val }

type TestPointerIdFactory struct { id mg.PointerId }

func NewTestPointerIdFactory() *TestPointerIdFactory {
    return &TestPointerIdFactory{ id: mg.PointerId( 1 ) }
}

func ( f *TestPointerIdFactory ) NextPointerId() mg.PointerId {
    res := f.id
    f.id++
    return res
}

func ( f *TestPointerIdFactory ) NextListStart( 
    lt *mg.ListTypeReference ) *ListStartEvent {

    return NewListStartEvent( lt, f.NextPointerId() )
}

func ( f *TestPointerIdFactory ) NextValueListStart() *ListStartEvent {
    return f.NextListStart( mg.TypeOpaqueList )
}

func ( f *TestPointerIdFactory ) NextMapStart() *MapStartEvent {
    return NewMapStartEvent( f.NextPointerId() )
}

func ( f *TestPointerIdFactory ) NextValueAllocation( 
    typ mg.TypeReference ) *ValueAllocationEvent {

    return NewValueAllocationEvent( typ, f.NextPointerId() )
}

func ( f *TestPointerIdFactory ) nextHeapTestValue( 
    val mg.Value ) heapTestValue {

    return heapTestValue{ val: val, id: f.NextPointerId() }
}

func CheckNoError( err error, c *ReactorTestCall ) {
    if err != nil { c.Fatalf( "Got no error but expected %T: %s", err, err ) }
}

func AssertCastError( expct, act error, pa *assert.PathAsserter ) {
    ca := mg.CastErrorAssert{ ErrExpect: expct, ErrAct: act, PathAsserter: pa }
    ca.Call()
}
