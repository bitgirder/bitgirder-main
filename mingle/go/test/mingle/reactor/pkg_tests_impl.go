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
    bb := &bytes.Buffer{}
    fmt.Fprintf( bb, "feeding " )
    if val, ok := t.Source.( mg.Value ); ok {
        fmt.Fprint( bb, mg.QuoteValue( val ) )
    } else { 
        fmt.Fprintf( bb, "%T", t.Source )
    }
    c.Log( bb.String() )
    if err := FeedSource( t.Source, rct ); err == nil {
        checkNoError( t.Error, c )
        reqChk.checkRequest()
    } else { c.EqualErrors( t.Error, err ) }
}
