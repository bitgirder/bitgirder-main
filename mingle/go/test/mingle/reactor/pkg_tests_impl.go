package reactor

import (
    mg "mingle"
    "bitgirder/assert"
    "bitgirder/stack"
)

func ( t *ValueBuildTest ) Call( c *ReactorTestCall ) {
    rcts := []interface{}{}
    rcts = append( rcts, NewDebugReactor( c ) )
    vb := NewValueBuilder()
    rcts = append( rcts, vb )
    pip := InitReactorPipeline( rcts... )
    var err error
    if t.Source == nil {
        c.Logf( "visiting %s", mg.QuoteValue( t.Val ) )
        err = VisitValue( t.Val, pip )
    } else { err = FeedSource( t.Source, pip ) }
    if err == nil {
        mg.AssertEqualWireValues( t.Val, vb.GetValue(), c.PathAsserter )
    } else { c.Fatal( err ) }
}

func ( t *StructuralReactorErrorTest ) Call( c *ReactorTestCall ) {
    rct := NewStructuralReactor( t.TopType )
//    pip := InitReactorPipeline( rct )
    pip := InitReactorPipeline( NewDebugReactor( c ), rct )
    src := eventSliceSource( t.Events )
    c.Logf( "calling structural test, err: %s", t.Error )
    if err := FeedEventSource( src, pip ); err == nil {
        c.Fatalf( "Expected error (%T): %s", t.Error, t.Error ) 
    } else { c.EqualErrors( t.Error, err ) }
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
    mg.AssertEqualWireValues( t.Expect, vb.GetValue(), c.PathAsserter )
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
        mg.AssertEqualWireValues( expct, vb.GetValue(), a ) 
    }
}
