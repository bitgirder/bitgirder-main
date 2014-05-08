package reactor

import (
    mg "mingle"
    "bitgirder/assert"
    "bitgirder/stack"
    "bitgirder/objpath"
    "bytes"
    "fmt"
)

func ( t *ValueBuildTest ) Call( c *ReactorTestCall ) {
    rcts := []interface{}{}
//    rcts = append( rcts, NewDebugReactor( c ) )
    vb := NewValueBuilder()
    rcts = append( rcts, vb )
    pip := InitReactorPipeline( rcts... )
    var err error
    if t.Source == nil {
//        c.Logf( "visiting %s", QuoteValue( t.Val ) )
        err = VisitValue( t.Val, pip )
    } else { err = FeedSource( t.Source, pip ) }
    if err == nil {
        mg.EqualWireValues( t.Val, vb.GetValue(), c.PathAsserter )
    } else { c.Fatal( err ) }
}

func ( t *StructuralReactorErrorTest ) Call( c *ReactorTestCall ) {
    rct := NewStructuralReactor( t.TopType )
    pip := InitReactorPipeline( rct )
    src := eventSliceSource( t.Events )
    if err := FeedEventSource( src, pip ); err == nil {
        c.Fatalf( "Expected error (%T): %s", t.Error, t.Error ) 
    } else { c.Equal( t.Error, err ) }
}

func ( t *PointerEventCheckTest ) Call( c *ReactorTestCall ) {
    rct := InitReactorPipeline( 
        NewPathSettingProcessor(), NewPointerCheckReactor() )
    if err := FeedSource( t.Events, rct ); err == nil {
        c.Truef( t.Error == nil, 
            "expected error (%T): %s", t.Error, t.Error )
    } else {
        c.EqualErrors( t.Error, err )
    }
}

func ( t *EventPathTest ) Call( c *ReactorTestCall ) {
    rct := NewPathSettingProcessor();
    if t.StartPath != nil { rct.SetStartPath( t.StartPath ) }
    chk := NewEventPathCheckReactor( t.Events, c.PathAsserter )
    pip := InitReactorPipeline( rct, chk )
    src := eventExpectSource( t.Events )
    if err := FeedEventSource( src, pip ); err != nil { c.Fatal( err ) }
    chk.Complete()
}

// simple fixed impl of FieldOrderGetter
type fogImpl []FieldOrderReactorTestOrder

func ( fog fogImpl ) FieldOrderFor( qn *mg.QualifiedTypeName ) FieldOrder {
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

func ( ot *orderTracker ) checkField( fld *mg.Identifier ) {
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

func ( ocr *orderCheckReactor ) startStruct( qn *mg.QualifiedTypeName ) {
    for _, ord := range ocr.fo.Orders {
        if ord.Type.Equals( qn ) {
            ot := &orderTracker{ ocr: ocr, idx: 0, ord: ord.Order }
            ocr.push( ot )
            return
        }
    }
    ocr.push( "struct" )
}

func ( ocr *orderCheckReactor ) startField( fld *mg.Identifier ) {
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

func ( t *FieldOrderReactorTest ) Call( c *ReactorTestCall ) {
    vb := NewValueBuilder()
    chk := &orderCheckReactor{ 
        PathAsserter: c.PathAsserter,
        fo: t,
        stack: stack.NewStack(),
    }
    ordRct := NewFieldOrderReactor( fogImpl( t.Orders ) )
//    pip := InitReactorPipeline( ordRct, NewDebugReactor( c ), chk, vb )
    pip := InitReactorPipeline( ordRct, chk, vb )
    AssertFeedEventSource( eventSliceSource( t.Source ), pip, c )
    mg.EqualWireValues( t.Expect, vb.GetValue(), c.PathAsserter )
}

func ( t *FieldOrderMissingFieldsTest ) assertMissingFieldsError(
    mfe *mg.MissingFieldsError, 
    err error,
    c *ReactorTestCall ) {

    if mfe == nil { c.Fatal( err ) }
    if act, ok := err.( *mg.MissingFieldsError ); ok {
        c.Descend( "Location" ).Equal( mfe.Location(), act.Location() )
        c.Descend( "Error" ).Equal( mfe.Error(), act.Error() )
    } else { c.Fatal( err ) }
}

func ( t *FieldOrderMissingFieldsTest ) Call( c *ReactorTestCall ) {
    vb := NewValueBuilder()
    ord := NewFieldOrderReactor( fogImpl( t.Orders ) )
    rct := InitReactorPipeline( ord, vb )
    for _, ev := range t.Source {
        if err := rct.ProcessEvent( ev ); err != nil { 
            t.assertMissingFieldsError( t.Error, err, c )
            return
        }
    }
    if e2 := t.Error; e2 != nil { 
        c.Fatalf( "Expected error (%T): %s", e2, e2 ) 
    }
    c.Equalf( t.Expect, vb.GetValue(), "expected %s but got %s", 
        mg.QuoteValue( t.Expect ), mg.QuoteValue( vb.GetValue() ) )
}

func ( t *FieldOrderPathTest ) Call( c *ReactorTestCall ) {
    ps := NewPathSettingProcessor()
    ord := NewFieldOrderReactor( fogImpl( t.Orders ) )
    chk := NewEventPathCheckReactor( t.Expect, c.PathAsserter )
    pip := InitReactorPipeline( ps, ord, chk )
    src := eventSliceSource( t.Source )
    AssertFeedEventSource( src, pip, c )
    chk.Complete()
}

type castInterfaceImpl struct { 
    typers *mg.QnameMap
    c *ReactorTestCall
}

type castInterfaceTyper struct { *mg.IdentifierMap }

func ( t castInterfaceTyper ) FieldTypeFor( 
    fld *mg.Identifier, path objpath.PathNode ) ( mg.TypeReference, error ) {
    if t.HasKey( fld ) { return t.Get( fld ).( mg.TypeReference ), nil }
    return nil, mg.NewValueCastErrorf( path, "unrecognized field: %s", fld )
}

func newCastInterfaceImpl( c *ReactorTestCall ) *castInterfaceImpl {
    res := &castInterfaceImpl{ typers: mg.NewQnameMap(), c: c }
    m1 := castInterfaceTyper{ mg.NewIdentifierMap() }
    m1.Put( mg.MustIdentifier( "f1" ), mg.TypeInt32 )
    qn := mg.MustQualifiedTypeName
    res.typers.Put( qn( "ns1@v1/T1" ), m1 )
    m2 := castInterfaceTyper{ mg.NewIdentifierMap() }
    m2.Put( mg.MustIdentifier( "f1" ), mg.TypeInt32 )
    m2.Put( mg.MustIdentifier( "f2" ), mg.MustTypeReference( "ns1@v1/T1" ) )
    res.typers.Put( qn( "ns1@v1/T2" ), m2 )
    return res
}

func ( ci *castInterfaceImpl ) FieldTyperFor( 
    qn *mg.QualifiedTypeName, path objpath.PathNode ) ( FieldTyper, error ) {
    if ci.typers.HasKey( qn ) { return ci.typers.Get( qn ).( FieldTyper ), nil }
    if qn.ExternalForm() == "ns1@v1/FailType" {
        return nil, mg.NewValueCastErrorf( path, "test-message-fail-type" )
    }
    return nil, nil
}

func ( ci *castInterfaceImpl ) InferStructFor( qn *mg.QualifiedTypeName ) bool {
    return ci.typers.HasKey( qn )
}

func ( ci *castInterfaceImpl ) CastAtomic(
    v mg.Value,
    at *mg.AtomicTypeReference,
    path objpath.PathNode ) ( mg.Value, error, bool ) {

    if _, ok := v.( *mg.Null ); ok {
        return nil, fmt.Errorf( "Unexpected null val in cast impl" ), true
    }
    if ! at.Equals( mg.MustTypeReference( "ns1@v1/S3" ) ) {
        return nil, nil, false
    }
    if s, ok := v.( mg.String ); ok {
        switch s {
        case "cast1": return mg.Int32( 1 ), nil, true
        case "cast2": return mg.Int32( -1 ), nil, true
        case "cast3":
            return nil, mg.NewValueCastErrorf( path, "test-message-cast3" ), true
        }
        return nil, mg.NewValueCastErrorf( path, "Unexpected val: %s", s ), true
    }
    return nil, mg.NewTypeCastErrorValue( at, v, path ), true
}

func ( ci *castInterfaceImpl ) AllowAssignment(
    expct, act *mg.QualifiedTypeName ) bool {

    return expct.Equals( qname( "ns1@v1/T1" ) ) &&
       act.Equals( qname( "ns1@v1/T1Sub1" ) )
}

type eventAccContext struct {
    event ReactorEvent
    evs []ReactorEvent
    id mg.PointerId
}

func newEventAccContext( ev ReactorEvent ) *eventAccContext {
    return &eventAccContext{ 
        event: ev, 
        evs: make( []ReactorEvent, 0, 4 ),
        id: mg.PointerIdNull,
    }
}

func ( ctx *eventAccContext ) saveEvent( ev ReactorEvent ) {
    ctx.evs = append( ctx.evs, CopyEvent( ev, false ) )
}

type pointerEventAccumulator struct {
    *assert.PathAsserter
    evs map[ mg.PointerId ] []ReactorEvent
    accs *stack.Stack
    autoSave bool
}

func newPointerEventAccumulator( 
    a *assert.PathAsserter ) *pointerEventAccumulator {

    return &pointerEventAccumulator{
        PathAsserter: a,
        evs: make( map[ mg.PointerId ] []ReactorEvent ),
        accs: stack.NewStack(),
    }
}

func ( pe *pointerEventAccumulator ) saveEvent( ev ReactorEvent ) {
    pe.accs.VisitTop( func( v interface{} ) {
        ctx := v.( *eventAccContext )
        if ctx.id != mg.PointerIdNull { ctx.saveEvent( ev ) }
    })
}

func ( pe *pointerEventAccumulator ) startSave( 
    ev ReactorEvent, id mg.PointerId ) {

    ctx := pe.accs.Peek().( *eventAccContext )
    pe.Equal( ev, ctx.event )
    ctx.id = id
    ctx.saveEvent( ev )
}

func ( pe *pointerEventAccumulator ) popAcc() {
    ctx := pe.accs.Pop().( *eventAccContext )
    if ctx.id != mg.PointerIdNull { pe.evs[ ctx.id ] = ctx.evs }
    pe.processValue()
}

func ( pe *pointerEventAccumulator ) processValue() {
    if pe.accs.IsEmpty() { return }
    ctx := pe.accs.Peek().( *eventAccContext )
    if _, ok := ctx.event.( *ValueAllocationEvent ); ok { pe.popAcc() }
}

func ( pe *pointerEventAccumulator ) processEnd() { pe.popAcc() }

func ( pe *pointerEventAccumulator ) optAutoSave( ev ReactorEvent ) {
    switch v := ev.( type ) {
    case *ValueAllocationEvent: pe.startSave( v, v.Id )
    case *ListStartEvent: pe.startSave( v, v.Id )
    case *MapStartEvent: pe.startSave( v, v.Id )
    }
}

func ( pe *pointerEventAccumulator ) ProcessEvent( ev ReactorEvent ) error {
    pe.saveEvent( ev )
    switch ev.( type ) {
    case *ListStartEvent, *MapStartEvent, *ValueAllocationEvent,
         *StructStartEvent:
        pe.accs.Push( newEventAccContext( ev ) )
    case *EndEvent: pe.processEnd()
    case *ValueEvent, *ValueReferenceEvent: pe.processValue()
    }
    if pe.autoSave { pe.optAutoSave( ev ) }
    return nil
}

type castTestPointerHandling struct {

    // only source events signaled by the cast reactor as involving a suppressed
    // or changed ref
    saves *pointerEventAccumulator 

    refs *pointerEventAccumulator // a copy of all source events with an id
}

func ( c *castTestPointerHandling ) AllowAssignment( 
    expct, act *mg.QualifiedTypeName ) bool {

    return false
}

func ( c *castTestPointerHandling ) CastAtomic(
    in mg.Value,
    at *mg.AtomicTypeReference,
    path objpath.PathNode ) ( mg.Value, error, bool ) {

    return nil, nil, false
}

func ( c *castTestPointerHandling ) FieldTypeFor(
    fld *mg.Identifier, path objpath.PathNode ) ( mg.TypeReference, error ) {

    t := mg.MustTypeReference
    switch s := fld.ExternalForm(); s {
    case "f0": return t( "Int32" ), nil
    case "f1": return t( "&Int32" ), nil
    case "f2": return t( "&Int32" ), nil
    case "f3": return t( "&Int64" ), nil
    case "f4": return t( "Int64" ), nil
    case "f5": return t( "Int64*" ), nil
    case "f6": return t( "Int32*" ), nil
    case "f7": return t( "String*" ), nil
    case "f8": return t( "&Int32*" ), nil
    case "f9": return t( "ns1@v1/S2" ), nil
    case "f10": return t( "ns1@v1/S2" ), nil
    case "f11": return t( "&ns1@v1/S2" ), nil
    case "f12": return t( "&Int64" ), nil
    case "f13": return t( "&Int64" ), nil
    case "f14": return t( "Int64" ), nil
    }
    return nil, rctErrorf( path, "unhandled field: %s", fld )
}

func ( c *castTestPointerHandling ) FieldTyperFor( 
    qn *mg.QualifiedTypeName,
    path objpath.PathNode ) ( FieldTyper, error ) { 

    return c, nil
}

func ( c *castTestPointerHandling ) InferStructFor(
    qn *mg.QualifiedTypeName ) bool {

    return false
}

func ( c *castTestPointerHandling ) sendEvents(
    evs []ReactorEvent, 
    typ mg.TypeReference,
    cr *CastReactor, 
    next ReactorEventProcessor ) error {
 
    cr.stack.Push( typ )
    for _, ev := range evs {
        ev = CopyEvent( ev, false )
        switch v := ev.( type ) {
        case *ListStartEvent: v.Id = mg.PointerIdNull
        case *MapStartEvent: v.Id = mg.PointerIdNull
        case *ValueAllocationEvent: v.Id = mg.PointerIdNull
        }
        if err := cr.ProcessEvent( ev, next ); err != nil { return err }
    }
    return nil
}

func ( c *castTestPointerHandling ) processValueReference(
    cr *CastReactor,
    ve *ValueReferenceEvent,
    typ mg.TypeReference,
    next ReactorEventProcessor ) error {
 
    if evs, ok := c.saves.evs[ ve.Id ]; ok {
        return c.sendEvents( evs, typ, cr, next )
    }
    if evs, ok := c.refs.evs[ ve.Id ]; ok {
        return c.sendEvents( evs, typ, cr, next )
    }
    return next.ProcessEvent( ve )
}

func ( c *castTestPointerHandling ) addDelegatesFor( cr *CastReactor ) {
    cr.AllocationSuppressed = func( ve *ValueAllocationEvent ) error {
        c.saves.startSave( ve, ve.Id )
        return nil
    }
    cr.CastingList = func( le *ListStartEvent, lt *mg.ListTypeReference ) error {
        c.saves.startSave( le, le.Id )
        return nil
    }
    cr.ProcessValueReference = func( 
        ve *ValueReferenceEvent, 
        typ mg.TypeReference, 
        next ReactorEventProcessor ) error {

        return c.processValueReference( cr, ve, typ, next )
    }
}

func ( t *CastReactorTest ) addCastReactors( 
    rcts []interface{}, c *ReactorTestCall ) []interface{} {

    ps := NewPathSettingProcessor()
    ps.SetStartPath( t.Path )
    c.Logf( "profile: %s", t.Profile )
    switch t.Profile {
    case "": rcts = append( rcts, ps, NewDefaultCastReactor( t.Type ) )
    case "interface-impl-basic": 
        cr := NewCastReactor( t.Type, newCastInterfaceImpl( c ) )
        rcts = append( rcts, ps, cr )
    case "interface-pointer-handling":
        cph := &castTestPointerHandling{}
        cph.saves = newPointerEventAccumulator( c.PathAsserter )
        cph.refs = newPointerEventAccumulator( c.PathAsserter )
        cph.refs.autoSave = true
        cr := NewCastReactor( t.Type, cph )
        cph.addDelegatesFor( cr )
        rcts = append( rcts, ps, cph.saves, cph.refs, cr )
    default: panic( libErrorf( "Unhandled profile: %s", t.Profile ) )
    }
    return rcts
}

func ( t *CastReactorTest ) Call( c *ReactorTestCall ) {
    rcts := []interface{}{}
    rcts = append( rcts, NewDebugReactor( c ) )
    rcts = t.addCastReactors( rcts, c )
    vb := NewValueBuilder()
    rcts = append( rcts, vb )
    pip := InitReactorPipeline( rcts... )
    logMsg := &bytes.Buffer{}
    fmt.Fprintf( logMsg, "casting to %s", t.Type )
    if valIn, ok := t.In.( mg.Value ); ok {
        fmt.Fprintf( logMsg, ", in val: %s", mg.QuoteValue( valIn ) )
    }
    if t.Expect != nil {
        fmt.Fprintf( logMsg, ", expect: %s", mg.QuoteValue( t.Expect ) )
    }
    c.Log( logMsg.String() )
    if err := FeedSource( t.In, pip ); err == nil { 
        if errExpct := t.Err; errExpct != nil {
            c.Fatalf( "Expected error (%T): %s", errExpct, errExpct )
        }
        c.Logf( "got val: %s", mg.QuoteValue( vb.GetValue() ) )
//        mg.EqualWireValues( ct.Expect, vb.GetValue(), c.PathAsserter )
        mg.EqualValues( t.Expect, vb.GetValue(), c.PathAsserter )
    } else { AssertCastError( t.Err, err, c.PathAsserter ) }
}

func CheckBuiltValue( 
    expct mg.Value, vb *ValueBuilder, a *assert.PathAsserter ) {

    if expct == nil {
        if vb != nil {
            a.Fatalf( "unexpected value: %s", mg.QuoteValue( vb.GetValue() ) )
        }
    } else { 
        a.Falsef( vb == nil, 
            "expecting value %s but value builder is nil", 
            mg.QuoteValue( expct ) )
        mg.EqualWireValues( expct, vb.GetValue(), a ) 
    }
}

func valueCheckReactor( 
    base ReactorEventProcessor, 
    evs []EventExpectation,
    a *assert.PathAsserter ) ReactorEventProcessor {

    if evs == nil { return base }
    evChk := NewEventPathCheckReactor( evs, a )
    evChk.ignorePointerIds = true
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
    ns *mg.Namespace, path objpath.PathNode ) error {

    return chk.checkVal( chk.st.Namespace, ns, reqFieldNs, "namespace" )
}

func ( chk *requestCheck ) Service( 
    svc *mg.Identifier, path objpath.PathNode ) error {

    return chk.checkVal( chk.st.Service, svc, reqFieldSvc, "service" )
}

func ( chk *requestCheck ) Operation( 
    op *mg.Identifier, path objpath.PathNode ) error {

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

func ( t *RequestReactorTest ) Call( c *ReactorTestCall ) {
    reqChk := &requestCheck{ 
        PathAsserter: c.PathAsserter, 
        st: t,
        reqFldMin: reqFieldNs,
    }
    rct := InitReactorPipeline( NewRequestReactor( reqChk ) )
    if err := FeedSource( t.Source, rct ); err == nil {
        checkNoError( t.Error, c )
        reqChk.checkRequest()
    } else { c.EqualErrors( t.Error, err ) }
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

func ( t *ResponseReactorTest ) Call( c *ReactorTestCall ) {
    chk := &responseCheck{ PathAsserter: c.PathAsserter, st: t }
    rct := InitReactorPipeline( NewResponseReactor( chk ) )
    if err := VisitValue( t.In, rct ); err == nil {
        checkNoError( t.Error, c )
        chk.check()
    } else { c.EqualErrors( t.Error, err ) }
}
