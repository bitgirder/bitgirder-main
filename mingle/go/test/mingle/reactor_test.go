package mingle

import (
    "testing"
    "bitgirder/assert"
    "container/list"
)

type reactorTestCall struct {
    *assert.PathAsserter
    rt ReactorTest
}

func ( c *reactorTestCall ) feedStructureEvents( 
    evs []ReactorEvent, tt ReactorTopType ) ( *StructuralReactor, error ) {
    rct := NewStructuralReactor( tt )
    pip := InitReactorPipeline( rct )
    for _, ev := range evs { 
        if err := pip.ProcessEvent( ev ); err != nil { return nil, err }
    }
    return rct, nil
}

func ( c *reactorTestCall ) callStructuralError(
    ss *StructuralReactorErrorTest ) {
    if _, err := c.feedStructureEvents( ss.Events, ss.TopType ); err == nil {
        c.Fatalf( "Expected error (%T): %s", ss.Error, ss.Error ) 
    } else { c.Equal( ss.Error, err ) }
}

type eventExpectCheck struct {
    idx int
    expect []EventExpectation
    *assert.PathAsserter
    pg PathGetter
}

func ( foc *eventExpectCheck ) Key() ReactorKey {
    return ReactorKey( "mingle.eventExpectCheck" )
}

func ( foc *eventExpectCheck ) Init( rpi *ReactorPipelineInit ) {
    rpi.VisitPredecessors( func( rct Reactor ) {
        if pg, ok := rct.( PathGetter ); ok { foc.pg = pg }
    })
    foc.Falsef( foc.pg == nil, "No path getter predecessor found" )
}

func ( foc *eventExpectCheck ) ProcessEvent(
    ev ReactorEvent, rep ReactorEventProcessor ) error {
    defer func() { foc.idx++ }()
    expct := foc.expect[ foc.idx ]
    foc.Logf( "Receiving event (%T) at %s: %v (expecting %T: %v)", 
        ev, FormatIdPath( foc.pg.GetPath() ), ev, expct.Event, expct.Event )
    foc.Equal( ev, expct.Event )
    foc.Equal( FormatIdPath( expct.Path ), FormatIdPath( foc.pg.GetPath() ) )
    return nil
}

type reactorEventSource interface {
    Len() int
    EventAt( int ) ReactorEvent
}

type eventSliceSource []ReactorEvent
func ( src eventSliceSource ) Len() int { return len( src ) }
func ( src eventSliceSource ) EventAt( i int ) ReactorEvent { return src[ i ] }

type eventExpectSource []EventExpectation

func ( src eventExpectSource ) Len() int { return len( src ) }

func ( src eventExpectSource ) EventAt( i int ) ReactorEvent {
    return src[ i ].Event
}

func ( c *reactorTestCall ) assertEventExpectations( 
    src reactorEventSource, 
    expct []EventExpectation,
    rcts []Reactor ) *ReactorPipeline {
    rcts2 := []Reactor{ NewStructuralReactor( ReactorTopTypeValue ) }
    rcts2 = append( rcts2, rcts... )
    chk := &eventExpectCheck{ expect: expct, PathAsserter: c.PathAsserter }
    rcts2 = append( rcts2, chk )
    pip := InitReactorPipeline( rcts2... )
    for i, e := 0, src.Len(); i < e; i++ {
        ev := src.EventAt( i )
        if err := pip.ProcessEvent( ev ); err != nil { c.Fatal( err ) }
    }
    c.Equal( len( expct ), chk.idx )
    return pip
}

func ( c *reactorTestCall ) callStructuralPath(
    pt *StructuralReactorPathTest ) {
    src := eventExpectSource( pt.Events )
    pip := c.assertEventExpectations( src, pt.Events, []Reactor{} )
    sr := pip.MustReactorForKey( ReactorKeyStructuralReactor ).
              ( *StructuralReactor )
    var act idPath
    if pt.StartPath == nil { 
        act = sr.GetPath() 
    } else { act = sr.AppendPath( pt.StartPath ) }
    c.Equal( pt.FinalPath, act )
}

func ( c *reactorTestCall ) callValueBuild( vb ValueBuildTest ) {
    rct := NewValueBuilder()
    pip := InitReactorPipeline( rct )
    if err := VisitValue( vb.Val, pip ); err == nil {
        c.Equal( vb.Val, rct.GetValue() )
    } else { c.Fatal( err ) }
}

// simple fixed impl of FieldOrderGetter
type fogImpl []*Identifier

func ( fog fogImpl ) FieldOrderFor( at *AtomicTypeReference ) []*Identifier {
    if at.Equals( atomicRef( "ns1@v1/S1" ) ) { return fog }
    return nil
}

type logReactor struct {
    key string
    a *assert.PathAsserter
}

func ( r logReactor ) Init( rpi *ReactorPipelineInit ) {}
func ( r logReactor ) Key() ReactorKey { return ReactorKey( r.key ) }

func ( r logReactor ) ProcessEvent( 
    ev ReactorEvent, rep ReactorEventProcessor ) error {
    r.a.Logf( "Receiving event (%T) %v", ev, ev ) 
    return rep.ProcessEvent( ev )
}

type orderCheckReactor struct {
    *assert.PathAsserter
    fo *FieldOrderReactorTest
    stack *list.List
}

func ( ocr *orderCheckReactor ) Init( rpi *ReactorPipelineInit ) {}

func ( ocr *orderCheckReactor ) Key() ReactorKey {
    return ReactorKey( "mingle.orderCheckReactor" )
}

func ( ocr *orderCheckReactor ) push( val interface{} ) {
    ocr.stack.PushFront( val )
}

type orderTracker struct {
    ocr *orderCheckReactor
    idx int
}

func ( ot *orderTracker ) checkField( fld *Identifier ) {
    fldIdx := -1
    for i, id := range ot.ocr.fo.Order {
        if id.Equals( fld ) { 
            fldIdx = i
            break
        }
    }
    if fldIdx < 0 { return } // Okay -- not a constrained field
    switch {
    case fldIdx == ot.idx: ot.idx++
    case fldIdx > ot.idx: ot.idx = fldIdx // assume skipping optional fields
    default:
        ot.ocr.Fatalf( "Expected field %s but saw %s",
            ot.ocr.fo.Order[ ot.idx ], fld )
    }
}

func ( ocr *orderCheckReactor ) startStruct( at *AtomicTypeReference ) {
    if at.Equals( atomicRef( "ns1@v1/S1" ) ) {
        ocr.push( &orderTracker{ ocr: ocr, idx: 0 } )
    } else { ocr.push( "struct" ) }
}

func ( ocr *orderCheckReactor ) startField( fld *Identifier ) {
    if ot, ok := ocr.stack.Front().Value.( *orderTracker ); ok {
        ot.checkField( fld )
    }
}

func ( ocr *orderCheckReactor ) ProcessEvent(
    ev ReactorEvent, rep ReactorEventProcessor ) error {
    switch v := ev.( type ) {
    case StructStartEvent: ocr.startStruct( v.Type )
    case ListStartEvent: ocr.push( "list" )
    case MapStartEvent: ocr.push( "map" )
    case FieldStartEvent: ocr.startField( v.Field )
    case EndEvent: ocr.stack.Remove( ocr.stack.Front() )
    }
    return rep.ProcessEvent( ev )
}

func ( c *reactorTestCall ) callFieldOrderReactor( fo *FieldOrderReactorTest ) {
    vb := NewValueBuilder()
    chk := &orderCheckReactor{ 
        PathAsserter: c.PathAsserter,
        fo: fo,
        stack: &list.List{},
    }
    pip := InitReactorPipeline(
        NewFieldOrderReactor( fogImpl( fo.Order ) ),
        chk,
        vb,
    )
    for _, ev := range fo.Source {
        if err := pip.ProcessEvent( ev ); err != nil { c.Fatal( err ) }
    }
    assert.Equal( fo.Expect, vb.GetValue() )
}

func ( c *reactorTestCall ) callFieldOrderPath( fo *FieldOrderPathTest ) {
    c.assertEventExpectations(
        eventSliceSource( fo.Source ),
        fo.Expect,
        []Reactor{ NewFieldOrderReactor( fogImpl( fo.Order ) ) },
    )
}

type castErrorAssert struct {
    ct *CastReactorTest
    err error
    *assert.PathAsserter
}

// Returns a path asserter that can be used further
func ( cea castErrorAssert ) assertValueError( 
    expct, act ValueError ) *assert.PathAsserter {
    a := cea.Descend( "Err" )
    a.Descend( "Error()" ).Equal( expct.Error(), act.Error() )
    a.Descend( "Message()" ).Equal( expct.Message(), act.Message() )
    a.Descend( "Location()" ).Equal( expct.Location(), act.Location() )
    return a
}

func ( cea castErrorAssert ) assertTcError() {
    if act, ok := cea.err.( *TypeCastError ); ok {
        expct := cea.ct.Err.( *TypeCastError )
        a := cea.assertValueError( expct, act )
        a.Descend( "expcted" ).Equal( expct.Expected, act.Expected )
        a.Descend( "actual" ).Equal( expct.Actual, act.Actual )
    } else { cea.Fatal( cea.err ) }
}

func ( cea castErrorAssert ) assertVcError() {
    if act, ok := cea.err.( *ValueCastError ); ok {
        cea.assertValueError( cea.ct.Err.( *ValueCastError ), act )
    } else { cea.Fatal( cea.err ) }
}

func ( cea castErrorAssert ) call() {
    switch cea.ct.Err.( type ) {
    case nil: cea.Fatal( cea.err )
    case *TypeCastError: cea.assertTcError()
    case *ValueCastError: cea.assertVcError()
    default: cea.Fatalf( "Unhandled Err type: %T", cea.ct.Err )
    }
}

type castInterfaceImpl struct { typers *QnameMap }

type castInterfaceTyper struct { *IdentifierMap }

func ( t castInterfaceTyper ) FieldTypeOf( 
    fld *Identifier, pg PathGetter ) ( TypeReference, error ) {
    if t.HasKey( fld ) { return t.Get( fld ).( TypeReference ), nil }
    errPath := pg.GetPath().Parent()
    return nil, newValueCastErrorf( errPath, "unrecognized field: %s", fld )
}

func newCastInterfaceImpl() *castInterfaceImpl {
    res := &castInterfaceImpl{ typers: NewQnameMap() }
    m1 := castInterfaceTyper{ NewIdentifierMap() }
    m1.Put( MustIdentifier( "f1" ), TypeInt32 )
    qn := MustQualifiedTypeName
    res.typers.Put( qn( "ns1@v1/T1" ), m1 )
    m2 := castInterfaceTyper{ NewIdentifierMap() }
    m2.Put( MustIdentifier( "f1" ), TypeInt32 )
    m2.Put( MustIdentifier( "f2" ), MustTypeReference( "ns1@v1/T1" ) )
    res.typers.Put( qn( "ns1@v1/T2" ), m2 )
    return res
}

func ( ci *castInterfaceImpl ) FieldTyperFor( 
    qn *QualifiedTypeName, pg PathGetter ) ( FieldTyper, error ) {
    if ci.typers.HasKey( qn ) { return ci.typers.Get( qn ).( FieldTyper ), nil }
    if qn.ExternalForm() == "ns1@v1/FailType" {
        return nil, newValueCastErrorf( pg.GetPath(), "test-message-fail-type" )
    }
    return nil, nil
}

func ( ci *castInterfaceImpl ) InferStructFor( at *AtomicTypeReference ) bool {
    return ci.typers.HasKey( at.Name.( *QualifiedTypeName ) )
}

func ( c *reactorTestCall ) createCastReactor( 
    ct *CastReactorTest ) *CastReactor {
    switch ct.Profile {
    case "": return NewDefaultCastReactor( ct.Type, ct.Path )
    case "interface-impl": 
        return NewCastReactor( ct.Type, newCastInterfaceImpl(), ct.Path )
    }
    panic( libErrorf( "Unhandled profile: %s", ct.Profile ) )
}

func ( c *reactorTestCall ) callCast( ct *CastReactorTest ) {
    rct := NewValueBuilder()
    pip := InitReactorPipeline( c.createCastReactor( ct ), rct )
    if err := VisitValue( ct.In, pip ); err == nil { 
        c.Equal( ct.Expect, rct.GetValue() )
    } else { ( castErrorAssert{ ct, err, c.PathAsserter } ).call() }
}

func ( c *reactorTestCall ) call() {
    c.Logf( "Calling reactor test of type %T", c.rt )
    switch s := c.rt.( type ) {
    case *StructuralReactorErrorTest: c.callStructuralError( s )
    case *StructuralReactorPathTest: c.callStructuralPath( s )
    case ValueBuildTest: c.callValueBuild( s )
    case *FieldOrderReactorTest: c.callFieldOrderReactor( s )
    case *FieldOrderPathTest: c.callFieldOrderPath( s )
    case *CastReactorTest: c.callCast( s )
    default: panic( libErrorf( "Unhandled test source: %T", c.rt ) )
    }
}

func TestReactors( t *testing.T ) {
    a := assert.NewListPathAsserter( t )
    for _, rt := range StdReactorTests {
        ( &reactorTestCall{ PathAsserter: a, rt: rt } ).call()
        a = a.Next()
    }
}
