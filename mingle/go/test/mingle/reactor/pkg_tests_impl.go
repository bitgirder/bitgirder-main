package reactor

import (
    mg "mingle"
    "bitgirder/assert"
    "bitgirder/stack"
    "bitgirder/objpath"
)

func bindErrForPath( p objpath.PathNode ) error {
    return NewBindError( p, testMsgErrorBadValue )
}

func bindErrForEvent( ev ReactorEvent ) error {
    return bindErrForPath( ev.GetPath() )
}

func bindErrForValue( v mg.Value, p objpath.PathNode ) error {
    if v == bindReactorErrorTestVal { return bindErrForPath( p ) }
    return nil
}

func bindTestErrorProduceValue() ( interface{}, error ) {
    return mg.String( "placeholder-val" ), nil
}

type bindTestErrorFactory int

func ( ef bindTestErrorFactory ) BindValue( 
    ve *ValueEvent ) ( interface{}, error ) {

    return ve.Val, bindErrForValue( ve.Val, ve.GetPath() )
}

func ( ef bindTestErrorFactory ) StartMap( 
    mse *MapStartEvent ) ( FieldSetBinder, error ) {

    return bindTestErrorFieldSetBinder( 1 ), nil
}

func ( ef bindTestErrorFactory ) StartStruct( 
    sse *StructStartEvent ) ( FieldSetBinder, error ) {

    if sse.Type.Equals( bindReactorErrorTestQn ) {
        return nil, bindErrForEvent( sse )
    }
    return bindTestErrorFieldSetBinder( 1 ), nil
}

func ( ef bindTestErrorFactory ) StartList( 
    lse *ListStartEvent ) ( ListBinder, error ) {

    if mg.TypeNameIn( lse.Type ).Equals( bindReactorErrorTestQn ) {
        return nil, bindErrForEvent( lse )
    }
    return bindTestErrorListBinder( 1 ), nil
}

type bindTestErrorListBinder int

func ( lb bindTestErrorListBinder ) AddValue( 
    val interface{}, path objpath.PathNode ) error {

    return bindErrForValue( val.( mg.Value ), path )
}

func ( lb bindTestErrorListBinder ) NextBinderFactory() BinderFactory {
    return bindTestErrorFactory( 1 )
}

func ( lb bindTestErrorListBinder ) ProduceValue(
    ee *EndEvent ) ( interface{}, error ) {

    return bindTestErrorProduceValue()
}

type bindTestErrorFieldSetBinder int

func ( fs bindTestErrorFieldSetBinder ) StartField( 
    fse *FieldStartEvent ) ( BinderFactory, error ) {
    
    if fse.Field.Equals( bindReactorErrorTestField ) {
        return nil, bindErrForPath( objpath.ParentOf( fse.GetPath() ) )
    }
    return bindTestErrorFactory( 1 ), nil
}

func ( fs bindTestErrorFieldSetBinder ) SetValue( 
    fld *mg.Identifier, val interface{}, path objpath.PathNode ) error {

    return bindErrForValue( val.( mg.Value ), path )
}

func ( fs bindTestErrorFieldSetBinder ) ProduceValue( 
    ee *EndEvent ) ( interface{}, error ) {

    return bindTestErrorProduceValue()
}

func ( t *BindReactorTest ) getBinderFactory() BinderFactory {
    switch t.Profile {
    case bindTestProfileDefault: return ValueBinderFactory
    case bindTestProfileError: return bindTestErrorFactory( 1 )
    }
    panic( libErrorf( "unhandled profile: %s", t.Profile ) )
}

func ( t *BindReactorTest ) Call( c *ReactorTestCall ) {
    br := NewBindReactor( t.getBinderFactory() )
    pip := InitReactorPipeline( NewDebugReactor( c ), br )
    src := t.Source
    if src == nil { src = t.Val }
    if mv, ok := src.( mg.Value ); ok {
        c.Logf( "feeding %s", mg.QuoteValue( mv ) )
    }
    if err := FeedSource( src, pip ); err == nil {
        act := br.GetValue().( mg.Value )
        mg.AssertEqualValues( t.Val, act, c.PathAsserter )
    } else { c.EqualErrors( t.Error, err ) }
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
    br := NewBindReactor( ValueBinderFactory )
    chk := &orderCheckReactor{ 
        PathAsserter: c.PathAsserter,
        fo: t,
        stack: stack.NewStack(),
    }
    ordRct := NewFieldOrderReactor( fogImpl( t.Orders ) )
//    pip := InitReactorPipeline( ordRct, NewDebugReactor( c ), chk, vb )
    pip := InitReactorPipeline( ordRct, chk, br )
    AssertFeedEventSource( eventSliceSource( t.Source ), pip, c )
    act := br.GetValue().( mg.Value )
    mg.AssertEqualValues( t.Expect, act, c.PathAsserter )
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
    br := NewBindReactor( ValueBinderFactory )
    ord := NewFieldOrderReactor( fogImpl( t.Orders ) )
    rct := InitReactorPipeline( ord, br )
    for _, ev := range t.Source {
        if err := rct.ProcessEvent( ev ); err != nil { 
            t.assertMissingFieldsError( t.Error, err, c )
            return
        }
    }
    if e2 := t.Error; e2 != nil { 
        c.Fatalf( "Expected error (%T): %s", e2, e2 ) 
    }
    act := br.GetValue().( mg.Value )
    c.Equalf( t.Expect, act, "expected %s but got %s", 
        mg.QuoteValue( t.Expect ), mg.QuoteValue( act ) )
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
}

func newEventAccContext( ev ReactorEvent ) *eventAccContext {
    return &eventAccContext{ event: ev, evs: make( []ReactorEvent, 0, 4 ) }
}

func ( ctx *eventAccContext ) saveEvent( ev ReactorEvent ) {
    ctx.evs = append( ctx.evs, CopyEvent( ev, false ) )
}

func CheckBuiltValue( 
    expct mg.Value, br *BindReactor, a *assert.PathAsserter ) {

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
