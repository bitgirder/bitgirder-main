package mingle

import (
    "testing"
    "bitgirder/assert"
    "bitgirder/objpath"
    "bitgirder/stack"
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

func ( c *ReactorTestCall ) feedStructureEvents( 
    evs []ReactorEvent, tt ReactorTopType ) ( *StructuralReactor, error ) {
    rct := NewStructuralReactor( tt )
//    pip := InitReactorPipeline( NewDebugReactor( c ), rct )
    pip := InitReactorPipeline( rct )
    for _, ev := range evs { 
        if err := pip.ProcessEvent( ev ); err != nil { return nil, err }
    }
    return rct, nil
}

func ( c *ReactorTestCall ) callStructuralError(
    ss *StructuralReactorErrorTest ) {
    if _, err := c.feedStructureEvents( ss.Events, ss.TopType ); err == nil {
        c.Fatalf( "Expected error (%T): %s", ss.Error, ss.Error ) 
    } else { c.Equal( ss.Error, err ) }
}

func ( c *ReactorTestCall ) assertEventExpectations( 
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

func ( c *ReactorTestCall ) callEventPath( pt *EventPathTest ) {
    
    rct := NewPathSettingProcessor();
    if pt.StartPath != nil { rct.SetStartPath( pt.StartPath ) }

    chk := newEventPathCheckReactor( pt.Events, c.PathAsserter )

//    pip := InitReactorPipeline( NewDebugReactor( c ), rct, chk )
    pip := InitReactorPipeline( rct, chk )
    
    src := eventExpectSource( pt.Events )
    if err := FeedEventSource( src, pip ); err != nil { c.Fatal( err ) }

    chk.complete()
}

func ( c *ReactorTestCall ) callValueBuild( vb ValueBuildTest ) {
    rct := NewValueBuilder()
    pip := InitReactorPipeline( rct )
    if err := VisitValue( vb.Val, pip ); err == nil {
        c.Equal( vb.Val, rct.GetValue() )
    } else { c.Fatal( err ) }
}

// simple fixed impl of FieldOrderGetter
type fogImpl []FieldOrderReactorTestOrder

func ( fog fogImpl ) FieldOrderFor( qn *QualifiedTypeName ) FieldOrder {
    for _, ord := range fog {
        if ord.Type.Equals( qn ) { return ord.Order }
    }
    return nil
}

type orderCheckReactor struct {
    *assert.PathAsserter
    fo *FieldOrderReactorTest
    stack *stack.Stack
}

func ( ocr *orderCheckReactor ) Init( rpi *ReactorPipelineInit ) {}

func ( ocr *orderCheckReactor ) Key() ReactorKey {
    return ReactorKey( "mingle.orderCheckReactor" )
}

func ( ocr *orderCheckReactor ) push( val interface{} ) {
    ocr.stack.Push( val )
}

type orderTracker struct {
    ocr *orderCheckReactor
    ord FieldOrder
    idx int
}

func ( ot *orderTracker ) checkField( fld *Identifier ) {
    fldIdx := -1
    for i, spec := range ot.ord {
        if spec.Field.Equals( fld ) { 
            fldIdx = i
            break
        }
    }
    if fldIdx < 0 { return } // Okay -- not a constrained field
    if fldIdx >= ot.idx {
        ot.idx = fldIdx // if '>' then assume we skipped some optional fields
        return
    }
    ot.ocr.Fatalf( "Expected field %s but saw %s", ot.ord[ ot.idx ].Field, fld )
}

func ( ocr *orderCheckReactor ) startStruct( qn *QualifiedTypeName ) {
    for _, ord := range ocr.fo.Orders {
        if ord.Type.Equals( qn ) {
            ot := &orderTracker{ ocr: ocr, idx: 0, ord: ord.Order }
            ocr.push( ot )
            return
        }
    }
    ocr.push( "struct" )
}

func ( ocr *orderCheckReactor ) startField( fld *Identifier ) {
    if ot, ok := ocr.stack.Peek().( *orderTracker ); ok {
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
    case EndEvent: ocr.stack.Pop()
    }
    return rep.ProcessEvent( ev )
}

func ( c *ReactorTestCall ) callFieldOrderReactor( fo *FieldOrderReactorTest ) {
    vb := NewValueBuilder()
    chk := &orderCheckReactor{ 
        PathAsserter: c.PathAsserter,
        fo: fo,
        stack: stack.NewStack(),
    }
    ordRct := NewFieldOrderReactor( fogImpl( fo.Orders ) )
    pip := InitReactorPipeline( ordRct, chk, vb )
    AssertFeedEventSource( eventSliceSource( fo.Source ), pip, c )
    assert.Equal( fo.Expect, vb.GetValue() )
}

func ( c *ReactorTestCall ) callFieldOrderPathTest( fo *FieldOrderPathTest ) {
    ord := NewFieldOrderReactor( fogImpl( fo.Orders ) )
    chk := newEventPathCheckReactor( fo.Expect, c.PathAsserter )
    pip := InitReactorPipeline( ord, chk )
    src := eventSliceSource( fo.Source )
    AssertFeedEventSource( src, pip, c )
    chk.complete()
}

func ( c *ReactorTestCall ) assertMissingFieldsError(
    mfe *MissingFieldsError, err error ) {
    if mfe == nil { c.Fatal( err ) }
    if act, ok := err.( *MissingFieldsError ); ok {
        c.Descend( "Location" ).Equal( mfe.Location(), act.Location() )
        c.Descend( "Error" ).Equal( mfe.Error(), act.Error() )
    } else { c.Fatal( err ) }
}

func ( c *ReactorTestCall ) callFieldOrderMissingFields(
    mf *FieldOrderMissingFieldsTest ) {
    vb := NewValueBuilder()
    ord := NewFieldOrderReactor( fogImpl( mf.Orders ) )
    rct := InitReactorPipeline( ord, vb )
    for _, ev := range mf.Source {
        if err := rct.ProcessEvent( ev ); err != nil { 
            c.assertMissingFieldsError( mf.Error, err )
            return
        }
    }
    if e2 := mf.Error; e2 != nil { 
        c.Fatalf( "Expected error (%T): %s", e2, e2 ) 
    }
    c.Equalf( mf.Expect, vb.GetValue(), "expected %s but got %s", 
        QuoteValue( mf.Expect ), QuoteValue( vb.GetValue() ) )
}

type castInterfaceImpl struct { 
    typers *QnameMap
    c *ReactorTestCall
}

type castInterfaceTyper struct { *IdentifierMap }

func ( t castInterfaceTyper ) FieldTypeFor( 
    fld *Identifier, path objpath.PathNode ) ( TypeReference, error ) {
    if t.HasKey( fld ) { return t.Get( fld ).( TypeReference ), nil }
    return nil, NewValueCastErrorf( path, "unrecognized field: %s", fld )
}

func newCastInterfaceImpl( c *ReactorTestCall ) *castInterfaceImpl {
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
    qn *QualifiedTypeName, path objpath.PathNode ) ( FieldTyper, error ) {
    if ci.typers.HasKey( qn ) { return ci.typers.Get( qn ).( FieldTyper ), nil }
    if qn.ExternalForm() == "ns1@v1/FailType" {
        return nil, NewValueCastErrorf( path, "test-message-fail-type" )
    }
    return nil, nil
}

func ( ci *castInterfaceImpl ) InferStructFor( qn *QualifiedTypeName ) bool {
    return ci.typers.HasKey( qn )
}

func ( ci *castInterfaceImpl ) CastAtomic(
    v Value,
    at *AtomicTypeReference,
    path objpath.PathNode ) ( Value, error, bool ) {

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
            return nil, NewValueCastErrorf( path, "test-message-cast3" ), true
        }
        return nil, NewValueCastErrorf( path, "Unexpected val: %s", s ), true
    }
    return nil, NewTypeCastErrorValue( at, v, path ), true
}

func ( c *ReactorTestCall ) addCastReactors( 
    ct *CastReactorTest, rcts []interface{} ) []interface{} {

    ps := NewPathSettingProcessor()
    ps.SetStartPath( ct.Path )

    switch ct.Profile {
    case "": rcts = append( rcts, ps, NewDefaultCastReactor( ct.Type ) )
    case "interface-impl": 
        cr := NewCastReactor( ct.Type, newCastInterfaceImpl( c ) )
        rcts = append( rcts, ps, cr )
    default: panic( libErrorf( "Unhandled profile: %s", ct.Profile ) )
    }
    return rcts
}

func ( c *ReactorTestCall ) callCast( ct *CastReactorTest ) {
    rcts := []interface{}{}
    rcts = append( rcts, NewDebugReactor( c ) )
    rcts = c.addCastReactors( ct, rcts )
    vb := NewValueBuilder()
    rcts = append( rcts, vb )
    pip := InitReactorPipeline( rcts... )
//    c.Logf( "Casting %s as %s", QuoteValue( ct.In ), ct.Type )
    if err := VisitValue( ct.In, pip ); err == nil { 
        if errExpct := ct.Err; errExpct != nil {
            c.Fatalf( "Expected error (%T): %s", errExpct, errExpct )
        }
        c.Equal( ct.Expect, vb.GetValue() )
    } else { AssertCastError( ct.Err, err, c.PathAsserter ) }
}

type requestCheck struct {
    *assert.PathAsserter
    st *ServiceRequestReactorTest
    reqFld requestFieldType
    auth *ValueBuilder
    params *ValueBuilder
}

func ( chk *requestCheck ) checkVal( 
    expct, act interface{}, 
    reqFldPrev, reqFldNext requestFieldType, 
    desc string ) error {
    chk.Descend( "reqFld" ).Equal( reqFldPrev, chk.reqFld )
    if chk.st.Error == nil { chk.Descend( desc ).Equal( expct, act ) }
    chk.reqFld = reqFldNext
    return nil
}

func ( chk *requestCheck ) Namespace( ns *Namespace, pg PathGetter ) error {
    return chk.checkVal( 
        chk.st.Namespace, ns, reqFieldNone, reqFieldNs, "namespace" )
}

func ( chk *requestCheck ) Service( svc *Identifier, pg PathGetter ) error {
    return chk.checkVal( 
        chk.st.Service, svc, reqFieldNs, reqFieldSvc, "service" )
}

func ( chk *requestCheck ) Operation( op *Identifier, pg PathGetter ) error {
    return chk.checkVal(
        chk.st.Operation, op, reqFieldSvc, reqFieldOp, "operation" )
}

type eventCheckReactor struct {
    *assert.PathAsserter
    ep ReactorEventProcessor
    evs []EventExpectation
    idx int
    pg PathGetter
}

func ( r *eventCheckReactor ) ProcessEvent( ev ReactorEvent ) error {
    defer func() {
        r.idx++
        r.PathAsserter = r.Next()
    }()
    if l := len( r.evs ); r.idx >= l { 
        r.Fatalf( "Expected only %d events", l ) 
    }
    ee := r.evs[ r.idx ]
    r.Descend( "event" ).Equal( ee.Event, ev )
    EqualPaths( ee.Path, r.pg.GetPath(), r.Descend( "path" ) )
    return r.ep.ProcessEvent( ev )
}

func optAsEventChecker( 
    ep ReactorEventProcessor, 
    pg PathGetter,
    evs []EventExpectation, 
    a *assert.PathAsserter ) ReactorEventProcessor {
    if evs == nil { return ep }
    return &eventCheckReactor{ 
        ep: ep, 
        evs: evs, 
        pg: pg,
        PathAsserter: a.StartList(),
    }
}

func ( chk *requestCheck ) GetAuthenticationProcessor(
    pg PathGetter ) ( ReactorEventProcessor, error ) {
    chk.Descend( "reqFld" ).Equal( reqFieldOp, chk.reqFld )
    chk.reqFld = reqFieldAuth
    chk.auth = NewValueBuilder()
    res := optAsEventChecker(
        chk.auth,
        pg,
        chk.st.AuthenticationEvents,
        chk.Descend( "authenticationEvents" ),
    )
    return res, nil
}

func ( chk *requestCheck ) GetParametersProcessor(
    pg PathGetter ) ( ReactorEventProcessor, error ) {
    chk.Descend( "reqFld" ).True( 
        chk.reqFld == reqFieldOp || chk.reqFld == reqFieldAuth )
    chk.reqFld = reqFieldParams
    chk.params = NewValueBuilder()
    res := optAsEventChecker( 
        chk.params, 
        pg, 
        chk.st.ParameterEvents, 
        chk.Descend( "parameterEvents" ),
    )
    return res, nil
}

func ( chk *requestCheck ) checkRequest() {
    CheckBuiltValue( 
        chk.st.Authentication, chk.auth, chk.Descend( "authentication" ) )
    CheckBuiltValue( 
        chk.st.Parameters, chk.params, chk.Descend( "parameters" ) )
}

func ( c *ReactorTestCall ) feedServiceRequest(
    src interface{}, ep ReactorEventProcessor ) error {
    switch v := src.( type ) {
    case []ReactorEvent:
        for _, ev := range v { 
            if err := ep.ProcessEvent( ev ); err != nil { return err }
        }
        return nil
    case Value: return VisitValue( v, ep )
    }
    panic( libErrorf( "Uhandled source: %T", src ) )
}

func ( c *ReactorTestCall ) callServiceRequest(
    st *ServiceRequestReactorTest ) {
    c.Logf( "Skipping %T", st )
//    reqChk := &requestCheck{ PathAsserter: c.PathAsserter, st: st }
//    rct := InitReactorPipeline( NewServiceRequestReactor( reqChk ) )
//    if err := c.feedServiceRequest( st.Source, rct ); err == nil {
//        c.CheckNoError( st.Error )
//        reqChk.checkRequest()
//    } else { c.EqualErrors( st.Error, err ) }
}

type responseCheck struct {
    *assert.PathAsserter
    st *ServiceResponseReactorTest
    err *ValueBuilder
    res *ValueBuilder
}

func ( rc *responseCheck ) GetResultProcessor( 
    pg PathGetter ) ( ReactorEventProcessor, error ) {
    rc.res = NewValueBuilder()
    res := optAsEventChecker(
        rc.res, pg, rc.st.ResEvents, rc.Descend( "result" ) )
    return res, nil
}

func ( rc *responseCheck ) GetErrorProcessor(
    pg PathGetter ) ( ReactorEventProcessor, error ) {
    rc.err = NewValueBuilder()
    res := optAsEventChecker(
        rc.err, pg, rc.st.ErrEvents, rc.Descend( "error" ) )
    return res, nil
}

func ( rc *responseCheck ) check() {
    CheckBuiltValue( rc.st.ResVal, rc.res, rc.Descend( "Result" ) )
    CheckBuiltValue( rc.st.ErrVal, rc.err, rc.Descend( "Error" ) )
}

func ( c *ReactorTestCall ) callServiceResponse( 
    st *ServiceResponseReactorTest ) {
    c.Logf( "Skipping %T", st )
//    chk := &responseCheck{ PathAsserter: c.PathAsserter, st: st }
//    rct := InitReactorPipeline( NewServiceResponseReactor( chk ) )
//    if err := VisitValue( st.In, rct ); err == nil {
//        c.CheckNoError( st.Error )
//        chk.check()
//    } else { c.EqualErrors( st.Error, err ) }
}

func ( c *ReactorTestCall ) call() {
//    c.Logf( "Calling reactor test of type %T", c.Test )
    switch s := c.Test.( type ) {
    case *StructuralReactorErrorTest: c.callStructuralError( s )
    case ValueBuildTest: c.callValueBuild( s )
    case *EventPathTest: c.callEventPath( s )
    case *FieldOrderReactorTest: c.callFieldOrderReactor( s )
    case *FieldOrderPathTest: c.callFieldOrderPathTest( s )
    case *FieldOrderMissingFieldsTest: c.callFieldOrderMissingFields( s )
    case *CastReactorTest: c.callCast( s )
    case *ServiceRequestReactorTest: c.callServiceRequest( s )
    case *ServiceResponseReactorTest: c.callServiceResponse( s )
    default: panic( libErrorf( "Unhandled test source: %T", c.Test ) )
    }
}

func TestReactors( t *testing.T ) {
    a := assert.NewPathAsserter( t )
    la := a.StartList();
    for _, rt := range StdReactorTests {
        ta := la
        if nt, ok := rt.( NamedTest ); ok { ta = a.Descend( nt.TestName() ) }
        a.Logf( "calling instance of of %T", rt )
        ( &ReactorTestCall{ PathAsserter: ta, Test: rt } ).call()
        la = la.Next()
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

type reactorImplTest struct {
    *assert.PathAsserter
    failOn *Identifier
    in Value
}

func ( t reactorImplTest ) Error() string { return "reactorImplTest" }

func ( t reactorImplTest ) Namespace( ns *Namespace, pg PathGetter ) error {
    return nil
}

func ( t reactorImplTest ) Service( svc *Identifier, pg PathGetter ) error {
    return nil
}

func ( t reactorImplTest ) Operation( op *Identifier, pg PathGetter ) error {
    return nil
}

func ( t reactorImplTest ) makeErr( pg PathGetter ) error {
    return NewValueCastError( pg.GetPath(), "test-error" )
}

func ( t reactorImplTest ) getProcessor(
    pg PathGetter,
    id *Identifier ) ( ReactorEventProcessor, error ) {
    if t.failOn.Equals( id ) { return nil, t.makeErr( pg ) }
    return DiscardProcessor, nil
}

func ( t reactorImplTest ) GetAuthenticationProcessor( 
    pg PathGetter ) ( ReactorEventProcessor, error ) {
    return t.getProcessor( pg, IdAuthentication )
}

func ( t reactorImplTest ) GetParametersProcessor( 
    pg PathGetter ) ( ReactorEventProcessor, error ) {
    return t.getProcessor( pg, IdParameters )
}

func ( t reactorImplTest ) GetErrorProcessor(
    pg PathGetter ) ( ReactorEventProcessor, error ) {
    return t.getProcessor( pg, IdError )
}

func ( t reactorImplTest ) GetResultProcessor(
    pg PathGetter ) ( ReactorEventProcessor, error ) {
    return t.getProcessor( pg, IdResult )
}

func ( t reactorImplTest ) callWith( rct ReactorEventProcessor ) {
    pip := InitReactorPipeline( rct )
    err := VisitValue( t.in, pip )
    errExpct := NewValueCastError( objpath.RootedAt( t.failOn ), "test-error" )
    t.EqualErrors( errExpct, err )
}

func TestRequestReactorImplErrors( t *testing.T ) {
    a := assert.NewPathAsserter( t )
    in := MustStruct( QnameServiceRequest,
        "namespace", "ns1@v1",
        "service", "svc1",
        "operation", "op1",
        "parameters", MustSymbolMap( "p1", 1 ),
        "authentication", 1,
    )
    for _, failOn := range []*Identifier{ IdAuthentication, IdParameters } {
        t := reactorImplTest{ 
            PathAsserter: a.Descend( failOn ), 
            failOn: failOn,
            in: in,
        }
        t.callWith( NewServiceRequestReactor( t ) )
    }
}

func TestResponseReactorImplErrors( t *testing.T ) {
    chk := func( failOn *Identifier, in Value ) {
        test := reactorImplTest{
            PathAsserter: assert.NewPathAsserter( t ).Descend( failOn ),
            failOn: failOn,
            in: in,
        }
        test.callWith( NewServiceResponseReactor( test ) )
    }
    chk( IdResult, MustSymbolMap( "result", 1 ) )
    chk( IdError, MustSymbolMap( "error", 1 ) )
}
