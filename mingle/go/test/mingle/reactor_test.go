package mingle

import (
    "testing"
    "bitgirder/assert"
    "bitgirder/objpath"
    "bitgirder/stack"
    "fmt"
)

func ( c *ReactorTestCall ) callStructuralError(
    ss *StructuralReactorErrorTest ) {

    rct := NewStructuralReactor( ss.TopType )
    pip := InitReactorPipeline( rct )
    src := eventSliceSource( ss.Events )
    if err := FeedEventSource( src, pip ); err == nil {
        c.Fatalf( "Expected error (%T): %s", ss.Error, ss.Error ) 
    } else { c.Equal( ss.Error, err ) }
}

func ( c *ReactorTestCall ) callEventPath( pt *EventPathTest ) {
    
    rct := NewPathSettingProcessor();
    if pt.StartPath != nil { rct.SetStartPath( pt.StartPath ) }

    chk := NewEventPathCheckReactor( pt.Events, c.PathAsserter )
    pip := InitReactorPipeline( rct, chk )
 
    src := eventExpectSource( pt.Events )
    if err := FeedEventSource( src, pip ); err != nil { c.Fatal( err ) }

    chk.Complete()
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
    case *StructStartEvent: ocr.startStruct( v.Type )
    case *ListStartEvent: ocr.push( "list" )
    case *MapStartEvent: ocr.push( "map" )
    case *FieldStartEvent: ocr.startField( v.Field )
    case *EndEvent: ocr.stack.Pop()
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
    ps := NewPathSettingProcessor()
    ord := NewFieldOrderReactor( fogImpl( fo.Orders ) )
    chk := NewEventPathCheckReactor( fo.Expect, c.PathAsserter )
    pip := InitReactorPipeline( ps, ord, chk )
    src := eventSliceSource( fo.Source )
    AssertFeedEventSource( src, pip, c )
    chk.Complete()
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

func ( ci *castInterfaceImpl ) AllowAssignment(
    expct, act *QualifiedTypeName ) bool {

    return expct.Equals( qname( "ns1@v1/T1" ) ) &&
       act.Equals( qname( "ns1@v1/T1Sub1" ) )
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
//    rcts = append( rcts, NewDebugReactor( c ) )
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

func valueCheckReactor( 
    base ReactorEventProcessor, 
    evs []EventExpectation,
    a *assert.PathAsserter ) ReactorEventProcessor {

    if evs == nil { return base }
    evChk := NewEventPathCheckReactor( evs, a )
    return InitReactorPipeline( evChk, base )
}

type requestCheck struct {
    *assert.PathAsserter
    st *RequestReactorTest
    reqFldMin requestFieldType
    auth *ValueBuilder
    params *ValueBuilder
}

func ( chk *requestCheck ) checkOrder( f requestFieldType ) {
    if min := chk.reqFldMin; f < min {
        chk.Fatalf( "got req field %d, less than min: %d", f, min )
    }
    chk.reqFldMin = f
}

func ( chk *requestCheck ) checkVal(
    expct, act interface{}, f requestFieldType, desc string ) error {

    chk.checkOrder( f )
    chk.Descend( desc ).Equal( expct, act )
    return nil
}

func ( chk *requestCheck ) Namespace( 
    ns *Namespace, path objpath.PathNode ) error {

    return chk.checkVal( chk.st.Namespace, ns, reqFieldNs, "namespace" )
}

func ( chk *requestCheck ) Service( 
    svc *Identifier, path objpath.PathNode ) error {

    return chk.checkVal( chk.st.Service, svc, reqFieldSvc, "service" )
}

func ( chk *requestCheck ) Operation( 
    op *Identifier, path objpath.PathNode ) error {

    return chk.checkVal( chk.st.Operation, op, reqFieldOp, "operation" )
}

func ( chk *requestCheck ) getProcessor(
    f requestFieldType,
    vbPtr **ValueBuilder,
    evs []EventExpectation ) ( ReactorEventProcessor, error ) {   
    
    chk.checkOrder( f )
    *vbPtr = NewValueBuilder()
    return valueCheckReactor( *vbPtr, evs, chk.PathAsserter ), nil
}

func ( chk *requestCheck ) GetAuthenticationReactor(
    path objpath.PathNode ) ( ReactorEventProcessor, error ) {

    evs := chk.st.AuthenticationEvents
    return chk.getProcessor( reqFieldAuth, &( chk.auth ), evs )
}

func ( chk *requestCheck ) GetParametersReactor(
    path objpath.PathNode ) ( ReactorEventProcessor, error ) {

    evs := chk.st.ParameterEvents
    return chk.getProcessor( reqFieldParams, &( chk.params ), evs )
}

func ( chk *requestCheck ) checkRequest() {
    CheckBuiltValue( 
        chk.st.Authentication, chk.auth, chk.Descend( "authentication" ) )
    CheckBuiltValue( 
        chk.st.Parameters, chk.params, chk.Descend( "parameters" ) )
}

func ( c *ReactorTestCall ) callRequest(
    st *RequestReactorTest ) {

    reqChk := &requestCheck{ 
        PathAsserter: c.PathAsserter, 
        st: st,
        reqFldMin: reqFieldNs,
    }
    rct := InitReactorPipeline( NewRequestReactor( reqChk ) )
    if err := FeedSource( st.Source, rct ); err == nil {
        c.CheckNoError( st.Error )
        reqChk.checkRequest()
    } else { c.EqualErrors( st.Error, err ) }
}

type responseCheck struct {
    *assert.PathAsserter
    st *ResponseReactorTest
    err *ValueBuilder
    res *ValueBuilder
}

func ( rc *responseCheck ) GetResultReactor( 
    path objpath.PathNode ) ( ReactorEventProcessor, error ) {

    rc.res = NewValueBuilder()
    res := valueCheckReactor( rc.res, rc.st.ResEvents, rc.Descend( "result" ) )
    return res, nil
}

func ( rc *responseCheck ) GetErrorReactor(
    path objpath.PathNode ) ( ReactorEventProcessor, error ) {

    rc.err = NewValueBuilder()
    res := valueCheckReactor( rc.err, rc.st.ErrEvents, rc.Descend( "error" ) )
    return res, nil
}

func ( rc *responseCheck ) check() {
    CheckBuiltValue( rc.st.ResVal, rc.res, rc.Descend( "Result" ) )
    CheckBuiltValue( rc.st.ErrVal, rc.err, rc.Descend( "Error" ) )
}

func ( c *ReactorTestCall ) callResponse( 
    st *ResponseReactorTest ) {

    chk := &responseCheck{ PathAsserter: c.PathAsserter, st: st }
    rct := InitReactorPipeline( NewResponseReactor( chk ) )
    if err := VisitValue( st.In, rct ); err == nil {
        c.CheckNoError( st.Error )
        chk.check()
    } else { c.EqualErrors( st.Error, err ) }
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
    case *RequestReactorTest: c.callRequest( s )
    case *ResponseReactorTest: c.callResponse( s )
    default: panic( libErrorf( "Unhandled test source: %T", c.Test ) )
    }
}

func TestReactors( t *testing.T ) {
//    a := assert.NewPathAsserter( t )
//    la := a.StartList();
//    for _, rt := range StdReactorTests {
//        ta := la
//        if nt, ok := rt.( NamedTest ); ok { ta = a.Descend( nt.TestName() ) }
//        ( &ReactorTestCall{ PathAsserter: ta, Test: rt } ).call()
//        la = la.Next()
//    }
}
