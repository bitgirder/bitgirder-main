package mingle

import (
    "testing"
    "bitgirder/assert"
    "bitgirder/objpath"
    "container/list"
    "fmt"
)

type noOpProcessor struct {
    initCalled bool
}

func ( p *noOpProcessor ) ProcessEvent( ev ReactorEvent ) error { return nil }

func ( p *noOpProcessor ) Init( rpi *ReactorPipelineInit ) {
    p.initCalled = true
}

type keyedNoOpProcessor struct {
    *noOpProcessor
    k ReactorKey
}

func ( kp *keyedNoOpProcessor ) Key() ReactorKey { return kp.k }

type initProcessor struct {
    find ReactorKey
    add interface{} 
    elt interface{}
}

func ( ip *initProcessor ) ProcessEvent( 
    ev ReactorEvent, rep ReactorEventProcessor ) error {
    return rep.ProcessEvent( ev )
}

func ( ip *initProcessor ) Init( rpi *ReactorPipelineInit ) {
    switch v := ip.add.( type ) {
    case ReactorEventProcessor: rpi.AddEventProcessor( v )
    case PipelineProcessor: rpi.AddPipelineProcessor( v )
    default: panic( libErrorf( "Bad add: %T", ip.add ) )
    }
    if elt, ok := rpi.FindByKey( ip.find ); ok { ip.elt = elt }
}

func TestReactorPipelineImpl( t *testing.T ) {
    a := assert.NewPathAsserter( t )
    p1 := &noOpProcessor{}
    p2 := &keyedNoOpProcessor{ 
        noOpProcessor: &noOpProcessor{}, 
        k: ReactorKey( "p2" ),
    }
    p3 := &noOpProcessor{}
    p4 := &initProcessor{ find: ReactorKey( "p2" ), add: p3 }
    p5 := &initProcessor{ find: ReactorKey( "p2" ), add: p4 }
    pip := InitReactorPipeline( p1, p2, p5 )
    a.Descend( "p1" ).True( p1.initCalled )
    a.Descend( "p2" ).True( p2.initCalled )
    a.Equal( p2, pip.MustFindByKey( ReactorKey( "p2" ) ) )
    a.Equal( p3, pip.elts[ 2 ] )
    a.Descend( "p3" ).True( p3.initCalled )
    a.Equal( p4, pip.elts[ 3 ] )
    a.Equal( p2, p4.elt )
    a.Equal( p2, p5.elt )
}

type reactorTestCall struct {
    *assert.PathAsserter
    rt ReactorTest
}

func ( c *reactorTestCall ) feedStructureEvents( 
    evs []ReactorEvent, tt ReactorTopType ) ( *StructuralReactor, error ) {
    rct := NewStructuralReactor( tt )
//    pip := InitReactorPipeline( NewDebugReactor( c ), rct )
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

func ( c *reactorTestCall ) assertEventExpectations( 
    src reactorEventSource, 
    expct []EventExpectation,
    rcts []interface{} ) *ReactorPipeline {
    return assertEventExpectations( src, expct, rcts, c.PathAsserter )
}

type pathCheckReactor struct {
    expct []EventExpectation
    pg PathGetter
    *assert.PathAsserter
    idx int
}

func ( r *pathCheckReactor ) ProcessEvent( ev ReactorEvent ) error {
    ee := r.expct[ r.idx ]
    r.Equal( ee.Event, ev )
    r.Equal( ee.Path, r.pg.GetPath() )
    r.idx++
    return nil
}

// Used as to verify that an EventPathReactor would, when used as a PathGetter
// for its wrapped processor, present the expected event paths. We use this
// separate method both to check that EventPathReactor behaves consistently with
// other path getters on the same input stream and also to have explicit
// coverage of EventPathReactor (testing FieldOrderReactor and others gives
// implicit coverage)
func assertEventPathReactorOn( 
    src reactorEventSource, expct []EventExpectation, a *assert.PathAsserter ) {
    a.Equal( src.Len(), len( expct ) )
    pcr := &pathCheckReactor{ expct: expct, PathAsserter: a }
    epr := NewEventPathReactor( pcr )
    pcr.pg = epr
    for i, e := 0, src.Len(); i < e; i++ {
        ev := src.EventAt( i )
        if err := epr.ProcessEvent( ev ); err != nil { a.Fatal( err ) }
    }
}

func ( c *reactorTestCall ) callEventPath(
    pt *EventPathTest ) {
    src := eventExpectSource( pt.Events )
    assertEventPathReactorOn( src, pt.Events, c.Descend( "epRctChk" ) )
    pip := c.assertEventExpectations( src, pt.Events, []interface{}{} )
    sr := pip.MustFindByKey( ReactorKeyStructuralReactor ).
              ( *StructuralReactor )
//    c.Logf( "Checking final paths, start path: %v", pt.StartPath )
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

func ( fog fogImpl ) FieldOrderFor( qn *QualifiedTypeName ) []*Identifier {
    if qn.Equals( MustQualifiedTypeName( "ns1@v1/S1" ) ) { return fog }
    return nil
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

func ( ocr *orderCheckReactor ) startStruct( qn *QualifiedTypeName ) {
    if qn.Equals( MustQualifiedTypeName( "ns1@v1/S1" ) ) {
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
        []interface{}{ NewFieldOrderReactor( fogImpl( fo.Order ) ) },
    )
}

type castInterfaceImpl struct { 
    typers *QnameMap
    c *reactorTestCall
}

type castInterfaceTyper struct { *IdentifierMap }

func ( t castInterfaceTyper ) FieldTypeOf( 
    fld *Identifier, pg PathGetter ) ( TypeReference, error ) {
    if t.HasKey( fld ) { return t.Get( fld ).( TypeReference ), nil }
    errPath := pg.GetPath().Parent()
    return nil, NewValueCastErrorf( errPath, "unrecognized field: %s", fld )
}

func newCastInterfaceImpl( c *reactorTestCall ) *castInterfaceImpl {
    res := &castInterfaceImpl{ typers: NewQnameMap(), c: c }
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
        return nil, NewValueCastErrorf( pg.GetPath(), "test-message-fail-type" )
    }
    return nil, nil
}

func ( ci *castInterfaceImpl ) InferStructFor( qn *QualifiedTypeName ) bool {
    return ci.typers.HasKey( qn )
}

func ( ci *castInterfaceImpl ) CastAtomic(
    v Value,
    at *AtomicTypeReference,
    pg PathGetter ) ( Value, error, bool ) {
    if _, ok := v.( *Null ); ok {
        return nil, fmt.Errorf( "Unexpected null val in cast impl" ), true
    }
    if ! at.Equals( MustTypeReference( "ns1@v1/S3" ) ) {
        return nil, nil, false
    }
    if s, ok := v.( String ); ok {
        switch s {
        case "cast1": return Int32( 1 ), nil, true
        case "cast2": return Int32( -1 ), nil, true
        case "cast3":
            p := pg.GetPath()
            return nil, NewValueCastErrorf( p, "test-message-cast3" ), true
        }
        p := pg.GetPath()
        return nil, NewValueCastErrorf( p, "Unexpected val: %s", s ), true
    }
    return nil, NewTypeCastErrorValue( at, v, pg.GetPath() ), true
}

func ( c *reactorTestCall ) createCastReactor( 
    ct *CastReactorTest ) *CastReactor {
    switch ct.Profile {
    case "": return NewDefaultCastReactor( ct.Type, ct.Path )
    case "interface-impl": 
        return NewCastReactor( ct.Type, newCastInterfaceImpl( c ), ct.Path )
    }
    panic( libErrorf( "Unhandled profile: %s", ct.Profile ) )
}

func ( c *reactorTestCall ) callCast( ct *CastReactorTest ) {
    rct := NewValueBuilder()
    pip := InitReactorPipeline( 
//        NewDebugReactor( c ),
        c.createCastReactor( ct ), 
        rct,
    )
//    c.Logf( "Casting %s as %s", QuoteValue( ct.In ), ct.Type )
    if err := VisitValue( ct.In, pip ); err == nil { 
        if errExpct := ct.Err; errExpct != nil {
            c.Fatalf( "Expected error (%T): %s", errExpct, errExpct )
        }
        c.Equal( ct.Expect, rct.GetValue() )
    } else { AssertCastError( ct.Err, err, c.PathAsserter ) }
}

func ( c *reactorTestCall ) call() {
//    c.Logf( "Calling reactor test of type %T", c.rt )
    switch s := c.rt.( type ) {
    case *StructuralReactorErrorTest: c.callStructuralError( s )
    case *EventPathTest: c.callEventPath( s )
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

func TestEventStackGetAndAppendPath( t *testing.T ) {
    get := func( s *eventStack, expct idPath ) {
        assert.Equal( expct, s.GetPath() )
    }
    apnd := func( s *eventStack, start, expct idPath ) {
        assert.Equal( expct, s.AppendPath( start ) )
    }
    s := newEventStack()
    p1 := objpath.RootedAt( id( "f1" ) )
    lp1 := objpath.RootedAtList()
    chkRoot := func() {
        get( s, nil )
        apnd( s, nil, nil )
        apnd( s, p1, p1 )
    }
    chkRoot()
    s.pushMap( "" )
    chkRoot()
    s.pushList( listIndex( 1 ) )
    get( s, lp1.Next() )
    apnd( s, p1, p1.StartList().Next() )
}
