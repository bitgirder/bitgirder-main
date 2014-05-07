package reactor

import (
    mg "mingle"
    "bitgirder/assert"
    "bitgirder/objpath"
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

func ( b *ReactorTestSetBuilder ) AddTest( t ReactorTest ) {
    b.tests = append( b.tests, t )
}

type ReactorTestSetInitializer func( b *ReactorTestSetBuilder ) 

var testInits []ReactorTestSetInitializer

//func init() { testInits = make( []ReactorTestSetInitializer, 0, 4 ) }

func AddTestInitializer( ti ReactorTestSetInitializer ) {
    if testInits == nil { 
        testInits = make( []ReactorTestSetInitializer, 0, 1024 ) 
    }
    testInits = append( testInits, ti )
}

func getReactorTests() []ReactorTest {
    b := &ReactorTestSetBuilder{ tests: make( []ReactorTest, 0, 1024 ) }
    for _, ti := range testInits { ti( b ) }
    return b.tests
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

//func AssertFeedEventSource(
//    src reactorEventSource, proc ReactorEventProcessor, a assert.Failer ) {
//    
//    if err := FeedEventSource( src, proc ); err != nil { a.Fatal( err ) }
//}

type EventExpectation struct {
    Event ReactorEvent
    Path objpath.PathNode
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

//func AssertFeedSource( 
//    src interface{}, rct ReactorEventProcessor, a assert.Failer ) {
//
//    if err := FeedSource( src, rct ); err != nil { a.Fatal( err ) }
//}
