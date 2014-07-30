package reactor

import (
    mg "mingle"
    "bitgirder/assert"
    "bitgirder/stack"
    "bitgirder/objpath"
)

func builderTestErrorProduceValue() ( interface{}, error ) {
    return mg.String( "placeholder-val" ), nil
}

type builderTestErrorFactory int

func ( ef builderTestErrorFactory ) BuildValue( 
    ve *ValueEvent ) ( interface{}, error ) {

    return ve.Val, testErrForValue( ve.Val, ve.GetPath() )
}

func ( ef builderTestErrorFactory ) StartMap( 
    mse *MapStartEvent ) ( FieldSetBuilder, error ) {

    return builderTestErrorFieldSetBuilder( 1 ), nil
}

func ( ef builderTestErrorFactory ) StartStruct( 
    sse *StructStartEvent ) ( FieldSetBuilder, error ) {

    if sse.Type.Equals( buildReactorErrorTestQn ) {
        return nil, testErrForEvent( sse )
    }
    return builderTestErrorFieldSetBuilder( 1 ), nil
}

func ( ef builderTestErrorFactory ) StartList( 
    lse *ListStartEvent ) ( ListBuilder, error ) {

    if mg.TypeNameIn( lse.Type ).Equals( buildReactorErrorTestQn ) {
        return nil, testErrForEvent( lse )
    }
    return builderTestErrorListBuilder( 1 ), nil
}

type builderTestErrorListBuilder int

func ( lb builderTestErrorListBuilder ) AddValue( 
    val interface{}, path objpath.PathNode ) error {

    return testErrForValue( val.( mg.Value ), path )
}

func ( lb builderTestErrorListBuilder ) NextBuilderFactory() BuilderFactory {
    return builderTestErrorFactory( 1 )
}

func ( lb builderTestErrorListBuilder ) ProduceValue(
    ee *EndEvent ) ( interface{}, error ) {

    return builderTestErrorProduceValue()
}

type builderTestErrorFieldSetBuilder int

func ( fs builderTestErrorFieldSetBuilder ) StartField( 
    fse *FieldStartEvent ) ( BuilderFactory, error ) {
    
    if fse.Field.Equals( buildReactorErrorTestField ) {
        return nil, testErrForPath( objpath.ParentOf( fse.GetPath() ) )
    }
    return builderTestErrorFactory( 1 ), nil
}

func ( fs builderTestErrorFieldSetBuilder ) SetValue( 
    fld *mg.Identifier, val interface{}, path objpath.PathNode ) error {

    return testErrForValue( val.( mg.Value ), path )
}

func ( fs builderTestErrorFieldSetBuilder ) ProduceValue( 
    ee *EndEvent ) ( interface{}, error ) {

    return builderTestErrorProduceValue()
}

type int32SliceBuilder struct {
    s []int32
}

func ( b *int32SliceBuilder ) AddValue( 
    val interface{}, path objpath.PathNode ) error {

    b.s = append( b.s, val.( int32 ) )
    return nil
}

func ( b *int32SliceBuilder ) NextBuilderFactory() BuilderFactory {
    return testBuilderFactory
}

func ( b *int32SliceBuilder ) ProduceValue(
    ee *EndEvent ) ( interface{}, error ) {

    return b.s, nil
}

var testBuilderFactory = NewFunctionsBuilderFactory()
var testBuilderFactoryFailOnly = NewFunctionsBuilderFactory()

func init() {
    testBuilderFactory.ErrorFunc = 
        func( path objpath.PathNode, msg string ) error {
            return newTestError( path, msg )
        }
    testBuilderFactory.ValueFunc =
        NewBuildValueOkFunctionSequence(
            func( ve *ValueEvent ) ( interface{}, error, bool ) {
                if v, ok := ve.Val.( mg.Int32 ); ok {
                    i := int32( v )
                    if i < 0 { return nil, testErrForEvent( ve ), true }
                    return i, nil, true
                }
                return nil, nil, false
            },
            func( ve *ValueEvent ) ( interface{}, error, bool ) {
                if s, ok := ve.Val.( mg.String ); ok {
                    return string( s ), nil, true
                }
                return nil, nil, false
            },
        )
    testBuilderFactory.ListFunc = 
        func( lse *ListStartEvent ) ( ListBuilder, error ) {
            switch lt := lse.Type; {
            case lt.Equals( asType( "Int32*" ) ): 
                sb := NewFunctionsListBuilder()
                sb.Value = make( []int32, 0, 4 )
                sb.NextFunc = func() BuilderFactory { 
                    return testBuilderFactory 
                }
                sb.AddFunc = 
                    func( val, res interface{}, 
                          path objpath.PathNode ) ( interface{}, error ) {

                        arr := res.( []int32 )
                        if cap( arr ) == len( arr ) {
                            return nil, testErrForPath( path )
                        }
                        return append( arr, val.( int32 ) ), nil
                    }
                return sb, nil
            }
            return nil, nil
        }
    testBuilderFactory.StructFunc = 
        func( sse *StructStartEvent ) ( FieldSetBuilder, error ) {
    
            if ! sse.Type.Equals( mkQn( "ns1@v1/S1" ) ) { return nil, nil }
            res := NewFunctionsFieldSetBuilder() 
            res.Value = new( s1 )
            res.RegisterField(
                mkId( "f1" ),
                func( path objpath.PathNode ) ( BuilderFactory, error ) {
                    return testBuilderFactory, nil
                },
                func( val interface{}, path objpath.PathNode ) error {
                    res.Value.( *s1 ).f1 = val.( int32 )
                    return nil
                },
            )
            res.RegisterField(
                mkId( "f2" ),
                func( path objpath.PathNode ) ( BuilderFactory, error ) {
                    return testBuilderFactory, nil
                },
                func( val interface{}, path objpath.PathNode ) error {
                    res.Value.( *s1 ).f2 = val.( []int32 )
                    return nil
                },
            )
            res.RegisterField(
                mkId( "f3" ),
                func( path objpath.PathNode ) ( BuilderFactory, error ) {
                    return testBuilderFactory, nil
                },
                func( val interface{}, path objpath.PathNode ) error {
                    res.Value.( *s1 ).f3 = val.( *s1 )
                    return nil
                },
            )
            return res, nil
        }
    testBuilderFactory.MapFunc = 
        func( mse *MapStartEvent ) ( FieldSetBuilder, error ) {
            res := NewFunctionsFieldSetBuilder()
            res.Value = make( map[ string ]interface{} )
            res.RegisterCatchall(
                func( fse *FieldStartEvent ) ( BuilderFactory, error ) {
                    switch fse.Field.ExternalForm() {
                    case "f1", "f2", "f3": return testBuilderFactory, nil
                    }
                    return nil, nil
                },
                func( 
                    fld *mg.Identifier, 
                    val interface{}, 
                    path objpath.PathNode ) error {

                    m := res.Value.( map[ string ]interface{} )
                    m[ fld.ExternalForm() ] = val
                    return nil
                },
            )
            return res, nil
        }
    testBuilderFactoryFailOnly.ErrorFunc = 
        func( loc objpath.PathNode, msg string ) error {
            return newTestError( loc, msg )
        }
    testBuilderFactoryFailOnly.ValueFunc = 
        func( ve *ValueEvent ) ( interface{}, error, bool ) {
            return nil, testErrForEvent( ve ), true
        }
    testBuilderFactoryFailOnly.ListFunc =
        func( lse *ListStartEvent ) ( ListBuilder, error ) {
            return nil, testErrForEvent( lse )
        }
    testBuilderFactoryFailOnly.MapFunc =
        func( mse *MapStartEvent ) ( FieldSetBuilder, error ) {
            return nil, testErrForEvent( mse )
        }
    testBuilderFactoryFailOnly.StructFunc =
        func( sse *StructStartEvent ) ( FieldSetBuilder, error ) {
            return nil, testErrForEvent( sse )
        }
}

func ( t *BuildReactorTest ) getBuilderFactory() BuilderFactory {
    switch t.Profile {
    case builderTestProfileDefault: return ValueBuilderFactory
    case builderTestProfileError: return builderTestErrorFactory( 1 )
    case builderTestProfileImpl: return testBuilderFactory
    case builderTestProfileImplFailOnly: return testBuilderFactoryFailOnly
    }
    panic( libErrorf( "unhandled profile: %s", t.Profile ) )
}

func ( t *BuildReactorTest ) Call( c *ReactorTestCall ) {
    br := NewBuildReactor( t.getBuilderFactory() )
//    pip := InitReactorPipeline( NewDebugReactor( c ), br )
    pip := InitReactorPipeline( br )
    src := t.Source
    if src == nil { src = t.Val }
//    if mv, ok := src.( mg.Value ); ok {
//        c.Logf( "feeding %s", mg.QuoteValue( mv ) )
//    }
    if err := FeedSource( src, pip ); err == nil {
        act := br.GetValue()
        switch v := t.Val.( type ) {
        case mg.Value: 
            mg.AssertEqualValues( v, act.( mg.Value ), c.PathAsserter )
        default: c.Equal( v, act )
        }
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
    br := NewBuildReactor( ValueBuilderFactory )
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
    br := NewBuildReactor( ValueBuilderFactory )
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
    expct mg.Value, br *BuildReactor, a *assert.PathAsserter ) {

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
