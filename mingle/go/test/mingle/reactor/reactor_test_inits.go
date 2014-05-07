package reactor

import (
    mg "mingle"
//    "bitgirder/objpath"
//    "bitgirder/assert"
//    "fmt"
//    "encoding/base64"
//    "bytes"
//    "strconv"
//    "log"
)

//var StdReactorTests []interface{}
//
//func init() { StdReactorTests = []interface{}{} }
//
//func AddStdReactorTests( t ...interface{} ) {
//    StdReactorTests = append( StdReactorTests, t... )
//}
//
//type NamedTest interface { TestName() string }
//
//func MakeTestId( i int ) *mg.Identifier {
//    return MustIdentifier( fmt.Sprintf( "f%d", i ) )
//}
//
//func mustInt( s string ) int {
//    res, err := strconv.Atoi( s )
//    if ( err != nil ) { panic( err ) }
//    return res
//}
//
//func flattenEvs( vals ...interface{} ) []ReactorEvent {
//    res := make( []ReactorEvent, 0, len( vals ) )
//    for _, val := range vals {
//        switch v := val.( type ) {
//        case ReactorEvent: res = append( res, v )
//        case []ReactorEvent: res = append( res, v... )
//        case []interface{}: res = append( res, flattenEvs( v... )... )
//        default: panic( libErrorf( "Uhandled ev type for flatten: %T", v ) )
//        }
//    }
//    return res
//}
//
//// to simplify test creation, we reuse event instances when constructing input
//// event sequences, and send them to this method only at the end to ensure that
//// we get a distinct sequence of event values for each test
//func CopySource( evs []ReactorEvent ) []ReactorEvent {
//    res := make( []ReactorEvent, len( evs ) )
//    for i, ev := range evs { res[ i ] = CopyEvent( ev, false ) }
//    return res
//}
//
//func startTestIdPath( elt interface{} ) objpath.PathNode {
//    switch v := elt.( type ) {
//    case int: return objpath.RootedAt( mg.MakeTestId( v ) )
//    case string: return objpath.RootedAtList().SetIndex( mustInt( v ) )
//    }
//    panic( libErrorf( "unhandled elt: %T", elt ) )
//}
//
//func MakeTestIdPath( elts ...interface{} ) objpath.PathNode { 
//    if len( elts ) == 0 { return nil }
//    res := startTestIdPath( elts[ 0 ] )
//    for i, e := 1, len( elts ); i < e; i++ {
//        switch v := elts[ i ].( type ) {
//        case int: res = res.Descend( mg.MakeTestId( v ) ) 
//        case string: res = res.StartList().SetIndex( mustInt( v ) )
//        default: panic( libErrorf( "unhandled elt: %T", v ) )
//        }
//    }
//    return res
//}
//
//func ptrId( i int ) mg.PointerId { return mg.PointerId( uint64( i ) ) }
//
//func ptrAlloc( typ TypeReferenceInitializer, i int ) *ValueAllocationEvent {
//    return NewValueAllocationEvent( asTypeReference( typ ), ptrId( i ) )
//}
//
//func ptrRef( i int ) *ValueReferenceEvent {
//    return NewValueReferenceEvent( ptrId( i ) )
//}
//
//type heapTestValue struct {
//    id mg.PointerId
//    val mg.Value
//}
//
//func ( ht heapTestValue ) valImpl() {}
//func ( ht heapTestValue ) Address() mg.PointerId { return ht.id }
//func ( ht heapTestValue ) Dereference() mg.Value { return ht.val }
//
//type TestPointerIdFactory struct { id mg.PointerId }
//
//func NewTestPointerIdFactory() *TestPointerIdFactory {
//    return &TestPointerIdFactory{ id: mg.PointerId( 1 ) }
//}
//
//func ( f *TestPointerIdFactory ) NextPointerId() mg.PointerId {
//    res := f.id
//    f.id++
//    return res
//}
//
//func ( f *TestPointerIdFactory ) NextListStart( 
//    lt *mg.ListTypeReference ) *ListStartEvent {
//
//    return NewListStartEvent( lt, f.NextPointerId() )
//}
//
//func ( f *TestPointerIdFactory ) NextValueListStart() *ListStartEvent {
//    return f.NextListStart( mg.TypeOpaqueList )
//}
//
//func ( f *TestPointerIdFactory ) NextMapStart() *MapStartEvent {
//    return NewMapStartEvent( f.NextPointerId() )
//}
//
//func ( f *TestPointerIdFactory ) NextValueAllocation( 
//    typ TypeReferenceInitializer ) *ValueAllocationEvent {
//
//    return NewValueAllocationEvent( asTypeReference( typ ), f.NextPointerId() )
//}
//
//func ( f *TestPointerIdFactory ) nextHeapTestValue( val mg.Value ) heapTestValue {
//    return heapTestValue{ val: val, id: f.NextPointerId() }
//}

type ValueBuildTest struct { 
    Val mg.Value 
    Source []ReactorEvent
}

func ( vb *ValueBuildTest ) Call( c *ReactorTestCall ) {
    rct := NewValueBuilder()
//    pip := InitReactorPipeline( NewDebugReactor( c ), rct )
    pip := InitReactorPipeline( rct )
    var err error
    if vb.Source == nil {
//        c.Logf( "visiting %s", QuoteValue( vb.Val ) )
        err = VisitValue( vb.Val, pip )
    } else { err = FeedSource( vb.Source, pip ) }
    if err == nil {
        mg.EqualWireValues( vb.Val, rct.GetValue(), c.PathAsserter )
    } else { c.Fatal( err ) }
}

func initValueBuildZeroRefTests( b *ReactorTestSetBuilder ) {
    qn := mg.MustQualifiedTypeName( "ns1@v1/S1" )
    listStart := func() *ListStartEvent {
        return NewListStartEvent( mg.TypeOpaqueList, mg.PointerIdNull )
    }
    b.AddTest(
        &ValueBuildTest{
            Val: mg.MustList(
                int32( 1 ),
                mg.MustSymbolMap( 
                    "f1", int32( 1 ),
                    "f2", mg.MustList( "hello" ),
                ),
                mg.NewHeapValue( mg.Int32( int32( 1 ) ) ),
                mg.NewHeapValue( mg.MustStruct( "ns1@v1/S1" ) ),
            ),
            Source: mg.CopySource(
                []ReactorEvent{
                    listStart(),
                        NewValueEvent( mg.Int32( int32( 1 ) ) ),
                        NewMapStartEvent( mg.PointerIdNull ),
                            NewFieldStartEvent( mg.MakeTestId( 1 ) ),
                                NewValueEvent( mg.Int32( int32( 1 ) ) ),
                            NewFieldStartEvent( mg.MakeTestId( 2 ) ),
                                listStart(),
                                    NewValueEvent( mg.String( "hello" ) ),
                                NewEndEvent(),
                        NewEndEvent(),
                        ptrAlloc( mg.TypeInt32, 0 ), 
                            NewValueEvent( mg.Int32( int32( 1 ) ) ),
                        ptrAlloc( qn.AsAtomicType(), 0 ), 
                            NewStructStartEvent( qn ), NewEndEvent(),
                    NewEndEvent(),
                },
            ),
        },
    )
}

//func initValueBuildReactorCycleTests() {
//    cyc := NewCyclicValues()
//    AddStdReactorTests(
//        ValueBuildTest{ Val: cyc.S1 },
//        ValueBuildTest{ Val: cyc.L1 },
//        ValueBuildTest{ Val: cyc.M1 },
//    )
//    qn1 := mg.MustQualifiedTypeName( "ns1@v1/S1" )
//    fld := mg.MakeTestId
//    // we create s1 pointing to s2 pointing to s1, but where s2 has non-cyclic
//    // fields as well, and then feed s2 field events in various orders in order
//    // to uncover errors related to clearing field state after encountering
//    // forward references
//    s1, s2 := mg.NewHeapValue( mg.NewStruct( qn1 ) ), mg.NewHeapValue( mg.NewStruct( qn1 ) )
//    s1Val, s2Val := s1.Dereference().( *mg.Struct ), s2.Dereference().( *mg.Struct )
//    s1ValTyp, s2ValTyp := s1Val.Type.AsAtomicType(), s2Val.Type.AsAtomicType()
//    s1Val.Fields.Put( fld( 1 ), s2 )
//    i1Val := mg.Int32( int32( 1 ) )
//    s2Val.Fields.Put( fld( 1 ), i1Val )
//    s2Val.Fields.Put( fld( 2 ), s1 )
//    s2Val.Fields.Put( fld( 3 ), s1 )
//    s2Val.Fields.Put( fld( 4 ), i1Val )
//    AddStdReactorTests(
//        ValueBuildTest{
//            Val: s1,
//            Source: []ReactorEvent{
//                NewValueAllocationEvent( s1ValTyp, s1.Address() ),
//                NewStructStartEvent( qn1 ),
//                    NewFieldStartEvent( fld( 1 ) ),
//                        NewValueAllocationEvent( s2ValTyp, s2.Address() ),
//                        NewStructStartEvent( qn1 ),
//                            NewFieldStartEvent( fld( 1 ) ),
//                                NewValueEvent( i1Val ),
//                            NewFieldStartEvent( fld( 2 ) ),
//                                NewValueReferenceEvent( s1.Address() ),
//                            NewFieldStartEvent( fld( 3 ) ),
//                                NewValueReferenceEvent( s1.Address() ),
//                            NewFieldStartEvent( fld( 4 ) ),
//                                NewValueEvent( i1Val ),
//                            NewEndEvent(),
//                    NewEndEvent(),
//            },
//        },
//    )
//}

func initValueBuildReactorTests( b *ReactorTestSetBuilder ) {
    s1 := mg.MustStruct( "ns1@v1/S1",
        "val1", mg.String( "hello" ),
        "list1", mg.MustList(),
        "map1", mg.MustSymbolMap(),
        "struct1", mg.MustStruct( "ns1@v1/S2" ),
    )
    addTest := func( v mg.Value ) { b.AddTest( &ValueBuildTest{ Val: v } ) }
    addTest( mg.String( "hello" ) )
    addTest( mg.MustList() )
    addTest( mg.MustList( 1, 2, 3 ) )
    addTest( mg.MustList( 1, mg.MustList(), mg.MustList( 1, 2 ) ) )
    addTest( mg.MustSymbolMap() )
    addTest( mg.MustSymbolMap( "f1", "v1", "f2", mg.MustList(), "f3", s1 ) )
    addTest( s1 )
    addTest( mg.MustStruct( "ns1@v1/S3" ) )
    addTest( mg.NewHeapValue( mg.String( "hello" ) ) )
    addTest( mg.NewHeapValue( mg.MustList() ) )
    addTest( 
        mg.NewHeapValue( 
            mg.MustList(
                mg.NewHeapValue( mg.Int32( 0 ) ),
                mg.Int32( 1 ),
                mg.NewHeapValue( mg.MustList( 0, 1 ) ),
                mg.String( "s1" ),
                mg.NewHeapValue( mg.String( "s2" ) ),
            ),
        ),
    )
    addTest( mg.NewHeapValue( s1 ) )
    addTest( 
        mg.NewHeapValue(
            mg.MustStruct( "ns1@v1/S1",
                "f1", mg.Int32( 1 ),
                "f2", mg.NewHeapValue( mg.Int32( 2 ) ),
                "f3", 
                    mg.NewHeapValue( 
                        mg.MustList( mg.NewHeapValue( mg.Int32( 1 ) ) ) ),
                "f4", mg.NewHeapValue( 
                    mg.MustSymbolMap( "g1", mg.NullVal, "g2", mg.Int32( 1 ) ) ),
            ),
        ),
    )
    valPtr1 := mg.NewHeapValue( mg.Int32( 1 ) )
    addTest( mg.MustList( valPtr1, valPtr1, valPtr1 ) )
//    initValueBuildZeroRefTests()
//    initValueBuildReactorCycleTests()
}

//type StructuralReactorErrorTest struct {
//    Events []ReactorEvent
//    Error *ReactorError
//    TopType ReactorTopType
//}
//
//type EventExpectation struct {
//    Event ReactorEvent
//    Path objpath.PathNode
//}
//
//type EventPathTest struct {
//    Name string
//    Events []EventExpectation
//    StartPath objpath.PathNode
//}
//
//func ( ept EventPathTest ) TestName() string { return ept.Name }
//
//// we only add here error tests; we assume that a value build reactor sits
//// behind a structural reactor and so let ValueBuildTest successes imply correct
//// behavior of the structural check reactor for valid inputs
//func initStructuralReactorTests() {
//    evStartStruct1 := NewStructStartEvent( qname( "ns1@v1/S1" ) )
//    id := mg.MakeTestId
//    evStartField1 := NewFieldStartEvent( id( 1 ) )
//    evStartField2 := NewFieldStartEvent( id( 2 ) )
//    evValue1 := NewValueEvent( mg.Int64( int64( 1 ) ) )
//    evValuePtr1 := NewValueAllocationEvent( mg.TypeInt64, 1 )
//    evListStart := NewListStartEvent( mg.TypeOpaqueList, ptrId( 1 ) )
//    evMapStart := NewMapStartEvent( ptrId( 2 ) )
//    mk1 := func( 
//        errMsg string, evs ...ReactorEvent ) *StructuralReactorErrorTest {
//        return &StructuralReactorErrorTest{
//            Events: mg.CopySource( evs ),
//            Error: rctError( nil, errMsg ),
//        }
//    }
//    mk2 := func( 
//        errMsg string, 
//        tt ReactorTopType, 
//        evs ...ReactorEvent ) *StructuralReactorErrorTest {
//        res := mk1( errMsg, evs... )
//        res.TopType = tt
//        return res
//    }
//    AddStdReactorTests(
//        mk1( "Saw start of field 'f2' while expecting a value for field 'f1'",
//            evStartStruct1, evStartField1, evStartField2,
//        ),
//        mk1( "Saw start of field 'f2' while expecting a value for field 'f1'",
//            evStartStruct1, evStartField1, evMapStart, evStartField1,
//            evStartField2,
//        ),
//        mk1( "Saw start of field 'f1' after value was built",
//            evStartStruct1, NewEndEvent(), evStartField1,
//        ),
//        mk1( "Expected field name or end of fields but got value",
//            evStartStruct1, evValue1,
//        ),
//        mk1( "Expected field name or end of fields but got &value",
//            evStartStruct1, evValuePtr1,
//        ),
//        mk1( "Expected field name or end of fields but got &reference",
//            evStartStruct1, ptrRef( 2 ),
//        ),
//        mk1( "Expected field name or end of fields but got list start",
//            evStartStruct1, evListStart,
//        ),
//        mk1( "Expected field name or end of fields but got map start",
//            evStartStruct1, evMapStart,
//        ),
//        mk1( "Expected field name or end of fields but got start of struct " +
//                evStartStruct1.Type.ExternalForm(),
//            evStartStruct1, evStartStruct1,
//        ),
//        mk1( "Saw end while expecting a value for field 'f1'",
//            evStartStruct1, evStartField1, NewEndEvent(),
//        ),
//        mk1( "Saw start of field 'f1' while expecting a list value",
//            evStartStruct1, evStartField1, evListStart, evStartField1,
//        ),
//        mk2( "Expected struct but got value", ReactorTopTypeStruct, evValue1 ),
//        mk2( "Expected struct but got &value", ReactorTopTypeStruct, 
//            evValuePtr1 ),
//        mk2( "Expected struct but got list start", ReactorTopTypeStruct,
//            evListStart,
//        ),
//        mk2( "Expected struct but got map start", ReactorTopTypeStruct,
//            evMapStart,
//        ),
//        mk2( "Expected struct but got start of field 'f1'", 
//            ReactorTopTypeStruct, evStartField1,
//        ),
//        mk2( "Expected struct but got end", 
//            ReactorTopTypeStruct, NewEndEvent() ),
//        mk1( "Multiple entries for field: f1",
//            evStartStruct1, 
//            evStartField1, evValue1,
//            evStartField2, evValue1,
//            evStartField1,
//        ),
//    )
//}
//
//type PointerEventCheckTest struct {
//    Events []ReactorEvent
//    Error error // if nil then Events should be fed through without error
//}
//
//func initPointerReferenceCheckTests() {
//    id, p := mg.MakeTestId, MakeTestIdPath
//    fld := func( i int ) *FieldStartEvent {
//        return NewFieldStartEvent( id( i ) )
//    }
//    ival := NewValueEvent( mg.Int32( 1 ) )
//    qn := mg.MustQualifiedTypeName( "ns1@v1/S1" )
//    add := func( path objpath.PathNode, msg string, evs ...ReactorEvent ) {
//        AddStdReactorTests(
//            &PointerEventCheckTest{
//                Events: mg.CopySource( evs ),
//                Error: rctError( path, msg ),
//            },
//        )
//    }
//    add0 := func( path objpath.PathNode, evs ...ReactorEvent ) {
//        add( path, "attempt to reference null pointer", evs... )
//    }
//    listStart := func( id mg.PointerId ) *ListStartEvent {
//        return NewListStartEvent( mg.TypeOpaqueList, id )
//    }
//    add0( p( "0" ), listStart( mg.PointerIdNull ), ptrRef( 0 ) )
//    add0( p( "1" ), listStart( mg.PointerIdNull ), ival, ptrRef( 0 ) )
//    add0( p( 1, 2 ),
//        NewMapStartEvent( 0 ), fld( 1 ), 
//            NewMapStartEvent( 0 ), fld( 2 ), ptrRef( 0 ) )
//    addReallocCheck := func( errEv ReactorEvent ) {
//        msg := "attempt to redefine reference: 1"
//        add( p( "1" ), msg,
//            listStart( ptrId( 2 ) ), 
//                ptrAlloc( mg.TypeInt32, 1 ), ival, errEv,
//        )
//        add( p( "1" ), msg, listStart( ptrId( 1 ) ), ival, errEv )
//        add( p( 1 ), msg, NewMapStartEvent( ptrId( 1 ) ), fld( 1 ), errEv )
//        add( p( 1 ), msg,
//            ptrAlloc( mg.TypeInt32, 1 ), 
//            NewStructStartEvent( qn ), fld( 1 ), errEv )
//    }
//    addReallocCheck( ptrAlloc( mg.TypeInt32, 1 ) )
//    addReallocCheck( listStart( ptrId( 1 ) ) )
//    addReallocCheck( NewMapStartEvent( ptrId( 1 ) ) )
//    add( p( "0" ), "unrecognized reference: 1",
//        listStart( ptrId( 2 ) ), ptrRef( 1 ) )
//    add( p( 2 ),
//        "unrecognized reference: 2",
//        NewMapStartEvent( ptrId( 3 ) ),
//        fld( 1 ),
//        ptrAlloc( mg.TypeInt32, 1 ), ival,
//        fld( 2 ),
//        ptrRef( 2 ),
//    )
//    AddStdReactorTests(
//        &PointerEventCheckTest{
//            Events: mg.CopySource( 
//                []ReactorEvent{
//                    NewStructStartEvent( qn ),
//                    fld( 1 ), ptrAlloc( mg.TypeInt32, 0 ), ival,
//                    fld( 2 ), listStart( mg.PointerIdNull ), NewEndEvent(),
//                    fld( 3 ), NewMapStartEvent( mg.PointerIdNull ), NewEndEvent(),
//                    NewEndEvent(),
//                },
//            ),
//        },
//    )
//}
//
//func initEventPathTests() {
//    p := MakeTestIdPath
//    ee := func( ev ReactorEvent, p objpath.PathNode ) EventExpectation {
//        return EventExpectation{ Event: ev, Path: p }
//    }
//    evStartStruct1 := NewStructStartEvent( qname( "ns1@v1/S1" ) )
//    id := mg.MakeTestId
//    evStartField := func( i int ) *FieldStartEvent {
//        return NewFieldStartEvent( id( i ) )
//    }
//    evValue := func( i int64 ) *ValueEvent {
//        return NewValueEvent( mg.Int64( i ) )
//    }
//    idFact := NewTestPointerIdFactory()
//    evEnd := NewEndEvent()
//    addTest := func( name string, evs ...EventExpectation ) {
//        ptrStart := ee( idFact.NextValueAllocation( mg.TypeValue ), nil )
//        evsWithPtr := append( []EventExpectation{ ptrStart }, evs... )
//        AddStdReactorTests(
//            &EventPathTest{ Name: name, Events: evs },
//            &EventPathTest{ Name: name + "-pointer", Events: evsWithPtr },
//        )
//    }
//    addTest( "empty" )
//    addTest( "top-value", ee( evValue( 1 ), nil ) )
//    addTest( "empty-struct",
//        ee( evStartStruct1, nil ),
//        ee( evEnd, nil ),
//    )
//    addTest( "empty-map",
//        ee( idFact.NextMapStart(), nil ),
//        ee( evStartField( 1 ), p( 1 ) ),
//            ee( evValue( 1 ), p( 1 ) ),
//        ee( evEnd, nil ),
//    )
//    addTest( "flat-struct",
//        ee( evStartStruct1, nil ),
//        ee( evStartField( 1 ), p( 1 ) ),
//            ee( evValue( 1 ), p( 1 ) ),
//        ee( evStartField( 2 ), p( 2 ) ),
//            ee( idFact.NextValueAllocation( mg.TypeInt64 ), p( 2 ) ),
//                ee( evValue( 2 ), p( 2 ) ),
//        ee( evEnd, nil ),
//    )
//    addTest( "empty-list",
//        ee( idFact.NextValueListStart(), nil ),
//        ee( NewEndEvent(), nil ),
//    )
//    addTest( "flat-list",
//        ee( idFact.NextValueListStart(), nil ),
//            ee( evValue( 1 ), p( "0" ) ),
//            ee( evValue( 1 ), p( "1" ) ),
//            ee( idFact.NextValueAllocation( mg.TypeInt64 ), p( "2" ) ),
//                ee( evValue( 2 ), p( "2" ) ),
//            ee( ptrRef( 2 ), p( "3" ) ),
//            ee( evValue( 4 ), p( "4" ) ),
//        ee( NewEndEvent(), nil ),
//    )
//    addTest( "nested-list1",
//        ee( idFact.NextValueListStart(), nil ),
//            ee( idFact.NextMapStart(), p( "0" ) ),
//                ee( evStartField( 1 ), p( "0", 1 ) ),
//                ee( evValue( 1 ), p( "0", 1 ) ),
//                ee( NewEndEvent(), p( "0" ) ),
//            ee( idFact.NextValueListStart(), p( "1" ) ),
//                ee( evValue( 1 ), p( "1", "0" ) ),
//                ee( NewEndEvent(), p( "1" ) ),
//            ee( idFact.NextValueAllocation( mg.TypeSymbolMap ), p( "2" ) ),
//                ee( idFact.NextMapStart(), p( "2" ) ),
//                    ee( evStartField( 1 ), p( "2", 1 ) ),
//                    ee( evValue( 1 ), p( "2", 1 ) ),
//                    ee( NewEndEvent(), p( "2" ) ),
//            ee( idFact.NextValueAllocation( "Int64*" ), p( "3" ) ),
//                ee( idFact.NextValueListStart(), p( "3" ) ),
//                    ee( evValue( 1 ), p( "3", "0" ) ),
//                    ee( NewEndEvent(), p( "3" ) ),
//            ee( evValue( 4 ), p( "4" ) ),
//        ee( NewEndEvent(), nil ),
//    )
//    addTest( "nested-list2",
//        ee( idFact.NextValueListStart(), nil ),
//            ee( idFact.NextMapStart(), p( "0" ) ),
//            ee( NewEndEvent(), p( "0" ) ),
//            ee( evValue( 1 ), p( "1" ) ),
//        ee( NewEndEvent(), nil ),
//    )
//    addTest( "nested-list3",
//        ee( idFact.NextValueListStart(), nil ),
//            ee( evValue( 1 ), p( "0" ) ),
//            ee( idFact.NextMapStart(), p( "1" ) ),
//                ee( evStartField( 1 ), p( "1", 1 ) ),
//                    ee( evValue( 1 ), p( "1", 1 ) ),
//                ee( NewEndEvent(), p( "1" ) ),
//            ee( idFact.NextValueAllocation( mg.TypeOpaqueList ), p( "2" ) ),
//                ee( idFact.NextValueListStart(), p( "2" ) ),
//                    ee( evValue( 1 ), p( "2", "0" ) ),
//                    ee( idFact.NextValueAllocation( mg.TypeOpaqueList ), 
//                            p( "2", "1" ) ),
//                        ee( idFact.NextValueListStart(), p( "2", "1" ) ),
//                            ee( evValue( 1 ), p( "2", "1", "0" ) ),
//                            ee( evValue( 2 ), p( "2", "1", "1" ) ),
//                        ee( NewEndEvent(), p( "2", "1" ) ),
//                    ee( evValue( 3 ), p( "2", "2" ) ),
//                ee( NewEndEvent(), p( "2" ) ),
//        ee( NewEndEvent(), nil ),
//    )
//    addTest( "list-regress1",
//        ee( idFact.NextValueListStart(), nil ),
//            ee( idFact.NextValueListStart(), p( "0" ) ),
//            ee( NewEndEvent(), p( "0" ) ),
//            ee( evValue( 1 ), p( "1" ) ),
//            ee( evValue( 1 ), p( "2" ) ),
//        ee( NewEndEvent(), nil ),
//    )
//    addTest( "flat-map",
//        ee( idFact.NextMapStart(), nil ),
//        ee( evStartField( 1 ), p( 1 ) ),
//            ee( evValue( 1 ), p( 1 ) ),
//        ee( NewEndEvent(), nil ),
//    )
//    addTest( "struct-with-containers1",
//        ee( evStartStruct1, nil ),
//        ee( evStartField( 1 ), p( 1 ) ),
//            ee( idFact.NextValueListStart(), p( 1 ) ),
//                ee( evValue( 1 ), p( 1, "0" ) ),
//                ee( evValue( 1 ), p( 1, "1" ) ),
//            ee( NewEndEvent(), p( 1 ) ),
//        ee( evStartField( 2 ), p( 2 ) ),
//            ee( idFact.NextValueAllocation( mg.TypeInt64 ), p( 2 ) ),
//                ee( evValue( 1 ), p( 2 ) ),
//        ee( evStartField( 3 ), p( 3 ) ),
//            ee( idFact.NextValueListStart(), p( 3 ) ),
//                ee( idFact.NextValueAllocation( mg.TypeInt64 ), p( 3, "0" ) ),
//                    ee( evValue( 0 ), p( 3, "0" ) ),
//                ee( idFact.NextValueAllocation( mg.TypeInt64 ), p( 3, "1" ) ),
//                    ee( evValue( 0 ), p( 3, "1" ) ),
//            ee( NewEndEvent(), p( 3 ) ),
//        ee( NewEndEvent(), nil ),
//    )
//    addTest( "struct-with-containers2",
//        ee( evStartStruct1, nil ),
//        ee( evStartField( 1 ), p( 1 ) ),
//            ee( idFact.NextMapStart(), p( 1 ) ),
//            ee( evStartField( 2 ), p( 1, 2 ) ),
//                ee( idFact.NextValueListStart(), p( 1, 2 ) ),
//                    ee( evValue( 1 ), p( 1, 2, "0" ) ),
//                    ee( evValue( 1 ), p( 1, 2, "1" ) ),
//                    ee( idFact.NextValueListStart(), p( 1, 2, "2" ) ),
//                        ee( evValue( 1 ), p( 1, 2, "2", "0" ) ),
//                        ee( idFact.NextMapStart(), p( 1, 2, "2", "1" ) ),
//                        ee( evStartField( 1 ), p( 1, 2, "2", "1", 1 ) ),
//                            ee( evValue( 1 ), p( 1, 2, "2", "1", 1 ) ),
//                        ee( evStartField( 2 ), p( 1, 2, "2", "1", 2 ) ),
//                            ee( idFact.NextValueAllocation( mg.TypeInt64 ), 
//                                p( 1, 2, "2", "1", 2 ) ),
//                            ee( evValue( 2 ), p( 1, 2, "2", "1", 2 ) ),
//                        ee( NewEndEvent(), p( 1, 2, "2", "1" ) ),
//                    ee( NewEndEvent(), p( 1, 2, "2" ) ),
//                ee( NewEndEvent(), p( 1, 2 ) ),
//            ee( NewEndEvent(), p( 1 ) ),
//        ee( NewEndEvent(), nil ),
//    )
//    AddStdReactorTests(
//        &EventPathTest{
//            Name: "non-empty-dict-start-path",
//            Events: []EventExpectation{
//                { idFact.NextMapStart(), p( 2 ) },
//                { evStartField( 1 ), p( 2, 1 ) },
//                { evValue( 1 ), p( 2, 1 ) },
//                { NewEndEvent(), p( 2 ) },
//            },
//            StartPath: p( 2 ),
//        },
//        &EventPathTest{
//            Name: "non-empty-list-start-path",
//            Events: []EventExpectation{ 
//                { idFact.NextMapStart(), p( 2, "3" ) },
//                { evStartField( 1 ), p( 2, "3", 1 ) },
//                { evValue( 1 ), p( 2, "3", 1 ) },
//                { NewEndEvent(), p( 2, "3" ) },
//            },
//            StartPath: p( 2, "3" ),
//        },
//    )
//}
//
//type FieldOrderReactorTestOrder struct {
//    Order FieldOrder
//    Type *mg.QualifiedTypeName
//}
//
//func testOrderWithIds( 
//    typ *mg.QualifiedTypeName, ids ...*Identifier ) FieldOrderReactorTestOrder {
//    ord := make( []FieldOrderSpecification, len( ids ) )
//    for i, id := range ids { 
//        ord[ i ] = FieldOrderSpecification{ Field: id, Required: false }
//    }
//    return FieldOrderReactorTestOrder{ Type: typ, Order: ord }
//}
//
//type FieldOrderReactorTest struct {
//    Source []ReactorEvent
//    Expect mg.Value
//    Orders []FieldOrderReactorTestOrder
//}
//
//func initFieldOrderValueTests() {
//    idFact := NewTestPointerIdFactory()
//    flds := make( []ReactorEvent, 5 )
//    ids := make( []*mg.Identifier, len( flds ) )
//    for i := 0; i < len( flds ); i++ {
//        ids[ i ] = id( fmt.Sprintf( "f%d", i ) )
//        flds[ i ] = NewFieldStartEvent( ids[ i ] )
//    }
//    i1 := mg.Int32( int32( 1 ) )
//    val1 := NewValueEvent( i1 )
//    t1, t2 := qname( "ns1@v1/S1" ), qname( "ns1@v1/S2" )
//    ss1, ss2 := NewStructStartEvent( t1 ), NewStructStartEvent( t2 )
//    ss2Val1 := mg.MustStruct( t2, ids[ 0 ], i1 )
//    // expct sequences for instance of ns1@v1/S1 by field f0 ...
//    fldVals := []mg.Value{
//        i1,
//        mg.MustSymbolMap( ids[ 0 ], i1, ids[ 1 ], ss2Val1 ),
//        mg.MustList( i1, i1 ),
//        ss2Val1,
//        i1,
//    }
//    mkExpct := func( ord ...int ) *mg.Struct {
//        pairs := []interface{}{}
//        for _, fldNum := range ord {
//            pairs = append( pairs, ids[ fldNum ], fldVals[ fldNum ] )
//        }
//        return mg.MustStruct( t1, pairs... )
//    }
//    // val sequences for fields f0 ...
//    fldEvs := [][]ReactorEvent {
//        []ReactorEvent{ val1 },
//        []ReactorEvent{
//            idFact.NextMapStart(), 
//                flds[ 0 ], val1, 
//                flds[ 1 ], ss2, flds[ 0 ], val1, NewEndEvent(),
//            NewEndEvent(),
//        },
//        []ReactorEvent{ 
//            idFact.NextValueListStart(), val1, val1, NewEndEvent() },
//        []ReactorEvent{ ss2, flds[ 0 ], val1, NewEndEvent() },
//        []ReactorEvent{ val1 },
//    }
//    mkSrc := func( ord ...int ) []ReactorEvent {
//        res := []ReactorEvent{ ss1 }
//        for _, fldNum := range ord {
//            res = append( res, flds[ fldNum ] )
//            res = append( res, fldEvs[ fldNum ]... )
//        }
//        return append( res, NewEndEvent() )
//    }
//    addTest1 := func( src []ReactorEvent, expct mg.Value ) {
//        AddStdReactorTests(
//            &FieldOrderReactorTest{ 
//                Source: mg.CopySource( src ), 
//                Expect: expct, 
//                Orders: []FieldOrderReactorTestOrder{
//                    testOrderWithIds( t1,
//                        ids[ 0 ], ids[ 1 ], ids[ 2 ], ids[ 3 ] ),
//                },
//            },
//        )
//    }
//    for _, ord := range [][]int {
//        []int{ 0, 1, 2, 3 }, // first one should be straight passthrough
//        []int{ 3, 2, 1, 0 },
//        []int{ 0, 3, 2, 1 }, 
//        []int{ 0, 2, 3, 1 },
//        []int{ 0, 1 },
//        []int{ 0, 2 },
//        []int{ 0, 3 },
//        []int{ 2, 0 },
//        []int{ 2, 1 },
//        []int{ 4, 3, 0, 1, 2 },
//        []int{ 4, 3, 2, 1, 0 },
//        []int{ 1, 4, 3, 0, 2 },
//        []int{ 1, 4, 3, 2, 0 },
//        []int{ 0, 4, 3, 2, 1 },
//        []int{ 0, 4, 3, 1, 2 },
//        []int{ 0, 1, 3, 2, 4 },
//    } {
//        addTest1( mkSrc( ord... ), mkExpct( ord... ) )
//    }
//    // Test nested orderings
//    AddStdReactorTests(
//        &FieldOrderReactorTest{
//            Source: mg.CopySource( 
//                []ReactorEvent{
//                    ss1, 
//                        flds[ 0 ], val1,
//                        flds[ 1 ], ss1,
//                            flds[ 2 ], 
//                                idFact.NextValueListStart(), 
//                                    val1, NewEndEvent(),
//                            flds[ 1 ], val1,
//                        NewEndEvent(),
//                    NewEndEvent(),
//                },
//            ),
//            Orders: []FieldOrderReactorTestOrder{
//                testOrderWithIds( t1, ids[ 1 ], ids[ 0 ], ids[ 2 ] ),
//            },
//            Expect: mg.MustStruct( t1,
//                ids[ 0 ], i1,
//                ids[ 1 ], mg.MustStruct( t1,
//                    ids[ 2 ], mg.MustList( i1 ),
//                    ids[ 1 ], i1,
//                ),
//            ),
//        },
//    )
//    // Test generic un-field-ordered values at the top-level as well
//    for i := 0; i < 4; i++ { addTest1( fldEvs[ i ], fldVals[ i ] ) }
//    // Test arbitrary values with no orders in play
//    addTest2 := func( expct mg.Value, src ...ReactorEvent ) {
//        AddStdReactorTests(
//            &FieldOrderReactorTest{
//                Source: mg.CopySource( src ),
//                Expect: expct,
//                Orders: []FieldOrderReactorTestOrder{},
//            },
//        )
//    }
//    addTest2( i1, val1 )
//    addTest2( mg.MustList(), idFact.NextValueListStart(), NewEndEvent() )
//    addTest2( mg.MustList( i1 ), idFact.NextValueListStart(), val1, NewEndEvent() )
//    addTest2( mg.MustSymbolMap(), idFact.NextMapStart(), NewEndEvent() )
//    addTest2( 
//        mg.MustSymbolMap( ids[ 0 ], i1 ), 
//        idFact.NextMapStart(), flds[ 0 ], val1, NewEndEvent(),
//    )
//    addTest2( mg.MustStruct( ss1.Type ), ss1, NewEndEvent() )
//    addTest2( 
//        mg.MustStruct( ss1.Type, ids[ 0 ], i1 ),
//        ss1, flds[ 0 ], val1, NewEndEvent(),
//    )
//}
//
//type FieldOrderMissingFieldsTest struct {
//    Orders []FieldOrderReactorTestOrder
//    Source []ReactorEvent
//    Expect mg.Value
//    Error *MissingFieldsError
//}
//
//func initFieldOrderMissingFieldTests() {
//    fldId := func( i int ) *mg.Identifier { return id( fmt.Sprintf( "f%d", i ) ) }
//    ord := FieldOrder( 
//        []FieldOrderSpecification{
//            { fldId( 0 ), true },
//            { fldId( 1 ), true },
//            { fldId( 2 ), false },
//            { fldId( 3 ), false },
//            { fldId( 4 ), true },
//        },
//    )
//    t1 := qname( "ns1@v1/S1" )
//    ords := []FieldOrderReactorTestOrder{ { Order: ord, Type: t1 } }
//    mkSrc := func( flds []int ) []ReactorEvent {
//        evs := []interface{}{ NewStructStartEvent( t1 ) }
//        for _, fld := range flds {
//            evs = append( evs, NewFieldStartEvent( fldId( fld ) ) )
//            evs = append( evs, NewValueEvent( mg.Int32( fld ) ) )
//        }
//        return flattenEvs( append( evs, NewEndEvent() ) )
//    }
//    mkVal := func( flds []int ) *mg.Struct {
//        pairs := make( []interface{}, 0, 2 * len( flds ) )
//        for _, fld := range flds {
//            pairs = append( pairs, fldId( fld ), mg.Int32( fld ) )
//        }
//        return mg.MustStruct( t1, pairs... )
//    }
//    addSucc := func( flds ...int ) {
//        AddStdReactorTests(
//            &FieldOrderMissingFieldsTest{
//                Orders: ords,
//                Source: mg.CopySource( mkSrc( flds ) ),
//                Expect: mkVal( flds ),
//            },
//        )
//    }
//    addSucc( 0, 1, 4 )
//    addSucc( 4, 0, 1 )
//    addSucc( 0, 1, 3, 4 )
//    addSucc( 0, 3, 1, 4 )
//    addSucc( 0, 1, 4, 3, 2 )
//    addErr := func( missIds []int, flds ...int ) {
//        miss := make( []*mg.Identifier, len( missIds ) )
//        for i, missId := range missIds { miss[ i ] = fldId( missId ) }
//        AddStdReactorTests(
//            &FieldOrderMissingFieldsTest{
//                Orders: ords,
//                Source: mg.CopySource( mkSrc( flds ) ),
//                Error: mg.NewMissingFieldsError( nil, miss ),
//            },
//        )
//    }
//    addErr( []int{ 0 }, 1, 2, 3, 4 )
//    addErr( []int{ 1 }, 0, 4, 3, 2 )
//    addErr( []int{ 4 }, 3, 2, 1, 0 )
//    addErr( []int{ 0, 1 }, 4 )
//    addErr( []int{ 1, 4 }, 0 )
//    addErr( []int{ 0, 4 }, 1 )
//    addErr( []int{ 4 }, 1, 0 )
//    addErr( []int{ 1 }, 4, 3, 0, 2 )
//}
//
//type FieldOrderPathTest struct {
//    Source []ReactorEvent
//    Expect []EventExpectation
//    Orders []FieldOrderReactorTestOrder
//}
//
//func initFieldOrderPathTests() {
//    mapStart := NewMapStartEvent( mg.PointerIdNull )
//    listStart := NewListStartEvent( mg.TypeOpaqueList, mg.PointerIdNull )
//    i1 := mg.Int32( int32( 1 ) )
//    val1 := NewValueEvent( i1 )
//    id := mg.MakeTestId
//    typ := func( i int ) *mg.QualifiedTypeName {
//        return qname( fmt.Sprintf( "ns1@v1/S%d", i ) )
//    }
//    ss := func( i int ) *StructStartEvent { 
//        return NewStructStartEvent( typ( i ) ) 
//    }
//    fld := func( i int ) *FieldStartEvent { 
//        return NewFieldStartEvent( id( i ) ) 
//    }
//    p := MakeTestIdPath
//    expct1 := []EventExpectation{
//        { ss( 1 ), nil },
//            { fld( 0 ), p( 0 ) },
//            { val1, p( 0 ) },
//            { fld( 1 ), p( 1 ) },
//            { mapStart, p( 1 ) },
//                { fld( 1 ), p( 1, 1 ) },
//                { val1, p( 1, 1 ) },
//                { fld( 0 ), p( 1, 0 ) },
//                { val1, p( 1, 0 ) },
//            { NewEndEvent(), p( 1 ) },
//            { fld( 2 ), p( 2 ) },
//            { listStart, p( 2 ) },
//                { val1, p( 2, "0" ) },
//                { val1, p( 2, "1" ) },
//            { NewEndEvent(), p( 2 ) },
//            { fld( 3 ), p( 3 ) },
//            { ss( 2 ), p( 3 ) },
//                { fld( 0 ), p( 3, 0 ) },
//                { val1, p( 3, 0 ) },
//                { fld( 1 ), p( 3, 1 ) },
//                { listStart, p( 3, 1 ) },
//                    { val1, p( 3, 1, "0" ) },
//                    { val1, p( 3, 1, "1" ) },
//                { NewEndEvent(), p( 3, 1 ) },
//            { NewEndEvent(), p( 3 ) },
//            { fld( 4 ), p( 4 ) },
//            { ss( 1 ), p( 4 ) },
//                { fld( 0 ), p( 4, 0 ) },
//                { val1, p( 4, 0 ) },
//                { fld( 1 ), p( 4, 1 ) },
//                { ss( 3 ), p( 4, 1 ) },
//                    { fld( 0 ), p( 4, 1, 0 ) },
//                    { val1, p( 4, 1, 0 ) },
//                    { fld( 1 ), p( 4, 1, 1 ) },
//                    { val1, p( 4, 1, 1 ) },
//                { NewEndEvent(), p( 4, 1 ) },
//                { fld( 2 ), p( 4, 2 ) },
//                { ss( 3 ), p( 4, 2 ) },
//                    { fld( 0 ), p( 4, 2, 0 ) },
//                    { val1, p( 4, 2, 0 ) },
//                    { fld( 1 ), p( 4, 2, 1 ) },
//                    { val1, p( 4, 2, 1 ) },
//                { NewEndEvent(), p( 4, 2 ) },
//                { fld( 3 ), p( 4, 3 ) },
//                { mapStart, p( 4, 3 ) },
//                    { fld( 0 ), p( 4, 3, 0 ) },
//                    { ss( 3 ), p( 4, 3, 0 ) },
//                        { fld( 0 ), p( 4, 3, 0, 0 ) },
//                        { val1, p( 4, 3, 0, 0 ) },
//                        { fld( 1 ), p( 4, 3, 0, 1 ) },
//                        { val1, p( 4, 3, 0, 1 ) },
//                    { NewEndEvent(), p( 4, 3, 0 ) },
//                    { fld( 1 ), p( 4, 3, 1 ) },
//                    { ss( 3 ), p( 4, 3, 1 ) },
//                        { fld( 0 ), p( 4, 3, 1, 0 ) },
//                        { val1, p( 4, 3, 1, 0 ) },
//                        { fld( 1 ), p( 4, 3, 1, 1 ) },
//                        { val1, p( 4, 3, 1, 1 ) },
//                    { NewEndEvent(), p( 4, 3, 1 ) },
//                { NewEndEvent(), p( 4, 3 ) },
//                { fld( 4 ), p( 4, 4 ) },
//                { listStart, p( 4, 4 ) },
//                    { ss( 3 ), p( 4, 4, "0" ) },
//                        { fld( 0 ), p( 4, 4, "0", 0 ) },
//                        { val1, p( 4, 4, "0", 0 ) },
//                        { fld( 1 ), p( 4, 4, "0", 1 ) },
//                        { val1, p( 4, 4, "0", 1 ) },
//                    { NewEndEvent(), p( 4, 4, "0" ) },
//                    { ss( 3 ), p( 4, 4, "1" ) },
//                        { fld( 0 ), p( 4, 4, "1", 0 ) },
//                        { val1, p( 4, 4, "1", 0 ) },
//                        { fld( 1 ), p( 4, 4, "1", 1 ) },
//                        { val1, p( 4, 4, "1", 1 ) },
//                    { NewEndEvent(), p( 4, 4, "1" ) },
//                { NewEndEvent(), p( 4, 4 ) },
//            { NewEndEvent(), p( 4 ) },
//        { NewEndEvent(), nil },
//    }
//    ords1 := []FieldOrderReactorTestOrder{
//        testOrderWithIds( ss( 1 ).Type,
//            id( 0 ), id( 1 ), id( 2 ), id( 3 ), id( 4 ) ),
//        testOrderWithIds( ss( 2 ).Type, id( 0 ), id( 1 ) ),
//        testOrderWithIds( ss( 3 ).Type, id( 0 ), id( 1 ) ),
//    }
//    evs := [][]ReactorEvent{
//        []ReactorEvent{ val1 },
//        []ReactorEvent{ 
//            mapStart, 
//                fld( 1 ), val1, fld( 0 ), val1, NewEndEvent() },
//        []ReactorEvent{ listStart, val1, val1, NewEndEvent() },
//        []ReactorEvent{ 
//            ss( 2 ), 
//                fld( 0 ), val1, 
//                fld( 1 ), listStart, val1, val1, NewEndEvent(),
//            NewEndEvent(),
//        },
//        // val for f4 is nested and has nested ss2 instances that are in varying
//        // need of reordering
//        []ReactorEvent{ 
//            ss( 1 ),
//                fld( 0 ), val1,
//                fld( 4 ), listStart,
//                    ss( 3 ),
//                        fld( 0 ), val1,
//                        fld( 1 ), val1,
//                    NewEndEvent(),
//                    ss( 3 ),
//                        fld( 1 ), val1,
//                        fld( 0 ), val1,
//                    NewEndEvent(),
//                NewEndEvent(),
//                fld( 2 ), ss( 3 ),
//                    fld( 1 ), val1,
//                    fld( 0 ), val1,
//                NewEndEvent(),
//                fld( 3 ), mapStart,
//                    fld( 0 ), ss( 3 ),
//                        fld( 1 ), val1,
//                        fld( 0 ), val1,
//                    NewEndEvent(),
//                    fld( 1 ), ss( 3 ),
//                        fld( 0 ), val1,
//                        fld( 1 ), val1,
//                    NewEndEvent(),
//                NewEndEvent(),
//                fld( 1 ), ss( 3 ),
//                    fld( 0 ), val1,
//                    fld( 1 ), val1,
//                NewEndEvent(),
//            NewEndEvent(),
//        },
//    }
//    mkSrc := func( ord ...int ) []ReactorEvent {
//        res := []ReactorEvent{ ss( 1 ) }
//        for _, i := range ord {
//            res = append( res, fld( i ) )
//            res = append( res, evs[ i ]... )
//        }
//        return append( res, NewEndEvent() )
//    }
//    for _, ord := range [][]int{
//        []int{ 0, 1, 2, 3, 4 },
//        []int{ 4, 3, 2, 1, 0 },
//        []int{ 2, 4, 0, 3, 1 },
//    } {
//        AddStdReactorTests(
//            &FieldOrderPathTest{
//                Source: mg.CopySource( mkSrc( ord... ) ),
//                Expect: expct1,
//                Orders: ords1,
//            },
//        )
//    }
//    AddStdReactorTests(
//        &FieldOrderPathTest{
//            Source: mg.CopySource(
//                []ReactorEvent{
//                    ss( 1 ),
//                        fld( 0 ), val1,
//                        fld( 7 ), val1,
//                        fld( 2 ), val1,
//                        fld( 1 ), val1,
//                    NewEndEvent(),
//                },
//            ),
//            Expect: []EventExpectation{
//                { ss( 1 ), nil },
//                { fld( 0 ), p( 0 ) },
//                { val1, p( 0 ) },
//                { fld( 7 ), p( 7 ) },
//                { val1, p( 7 ) },
//                { fld( 1 ), p( 1 ) },
//                { val1, p( 1 ) },
//                { fld( 2 ), p( 2 ) },
//                { val1, p( 2 ) },
//                { NewEndEvent(), nil },
//            },
//            Orders: []FieldOrderReactorTestOrder{
//                testOrderWithIds( ss( 1 ).Type, id( 0 ), id( 1 ), id( 2 ) ),
//            },
//        },
//    )
//    // Regression for bug fixed in previous commit (f7fa84122047)
//    AddStdReactorTests(
//        &FieldOrderPathTest{
//            Source: mg.CopySource(
//                []ReactorEvent{ ss( 1 ), fld( 1 ), val1, NewEndEvent() } ),
//            Expect: []EventExpectation{
//                { ss( 1 ), nil },
//                { fld( 1 ), p( 1 ) },
//                { val1, p( 1 ) },
//                { NewEndEvent(), nil },
//            },
//            Orders: []FieldOrderReactorTestOrder{
//                testOrderWithIds( ss( 1 ).Type, id( 0 ), id( 1 ), id( 2 ) ),
//            },
//        },
//    )
//}
//
//func initFieldOrderReactorTests() {
//    initFieldOrderValueTests()
//    initFieldOrderMissingFieldTests()
//    initFieldOrderPathTests()
//}
//
//type RequestReactorTest struct {
//    Source interface{}
//    Namespace *Namespace
//    Service *mg.Identifier
//    Operation *mg.Identifier
//    Parameters *mg.SymbolMap
//    ParameterEvents []EventExpectation
//    Authentication mg.Value
//    AuthenticationEvents []EventExpectation
//    Error error
//}
//
//func initRequestTests() {
//    idFact := NewTestPointerIdFactory()
//    ns1 := MustNamespace( "ns1@v1" )
//    svc1 := id( "service1" )
//    op1 := id( "op1" )
//    params1 := mg.MustSymbolMap( "f1", int32( 1 ) )
//    authQn := qname( "ns1@v1/Auth1" )
//    auth1 := mg.MustStruct( authQn, "f1", int32( 1 ) )
//    evFldNs := NewFieldStartEvent( IdNamespace )
//    evFldSvc := NewFieldStartEvent( IdService )
//    evFldOp := NewFieldStartEvent( IdOperation )
//    evFldParams := NewFieldStartEvent( IdParameters )
//    evFldAuth := NewFieldStartEvent( IdAuthentication )
//    evFldF1 := NewFieldStartEvent( id( "f1" ) )
//    evReqTyp := NewStructStartEvent( QnameRequest )
//    evNs1 := NewValueEvent( mg.String( ns1.ExternalForm() ) )
//    evSvc1 := NewValueEvent( mg.String( svc1.ExternalForm() ) )
//    evOp1 := NewValueEvent( mg.String( op1.ExternalForm() ) )
//    i32Val1 := NewValueEvent( mg.Int32( 1 ) )
//    evParams1 := []ReactorEvent{ 
//        idFact.NextMapStart(), evFldF1, i32Val1, NewEndEvent() }
//    evAuth1 := []ReactorEvent{ 
//        NewStructStartEvent( authQn ), evFldF1, i32Val1, NewEndEvent() }
//    addSucc1 := func( evs ...interface{} ) {
//        AddStdReactorTests(
//            &RequestReactorTest{
//                Source: mg.CopySource( flattenEvs( evs... ) ),
//                Namespace: ns1,
//                Service: svc1,
//                Operation: op1,
//                Parameters: params1,
//                Authentication: auth1,
//            },
//        )
//    }
//    fullOrderedReq1Flds := []interface{}{
//        evFldNs, evNs1,
//        evFldSvc, evSvc1,
//        evFldOp, evOp1,
//        evFldAuth, evAuth1,
//        evFldParams, evParams1,
//    }
//    addSucc1( evReqTyp, fullOrderedReq1Flds, NewEndEvent() )
//    addSucc1( idFact.NextMapStart(), fullOrderedReq1Flds, NewEndEvent() )
//    addSucc1( evReqTyp,
//        evFldAuth, evAuth1,
//        evFldOp, evOp1,
//        evFldParams, evParams1,
//        evFldNs, evNs1,
//        evFldSvc, evSvc1,
//        NewEndEvent(),
//    )
//    AddStdReactorTests(
//        &RequestReactorTest{
//            Source: mg.CopySource(
//                flattenEvs( evReqTyp,
//                    evFldNs, evNs1,
//                    evFldSvc, evSvc1,
//                    evFldOp, evOp1,
//                    evFldAuth, i32Val1,
//                    evFldParams, evParams1,
//                    NewEndEvent(),
//                ),
//            ),
//            Namespace: ns1,
//            Service: svc1,
//            Operation: op1,
//            Authentication: mg.Int32( 1 ),
//            Parameters: params1,
//        },
//    )
//    mkReq1 := func( params, auth mg.Value ) *mg.Struct {
//        pairs := []interface{}{ 
//            IdNamespace, NamespaceAsBytes( ns1 ),
//            IdService, IdentifierAsBytes( svc1 ),
//            IdOperation, IdentifierAsBytes( op1 ),
//        }
//        if params != nil { pairs = append( pairs, IdParameters, params ) }
//        if auth != nil { pairs = append( pairs, IdAuthentication, auth ) }
//        return mg.MustStruct( QnameRequest, pairs... )
//    }
//    addSucc2 := func( src interface{}, authExpct mg.Value ) {
//        AddStdReactorTests(
//            &RequestReactorTest{
//                Namespace: ns1,
//                Service: svc1,
//                Operation: op1,
//                Parameters: EmptySymbolMap(),
//                Authentication: authExpct,
//                Source: src,
//            },
//        )
//    } 
//    // check implicit params with(out) auth and using undetermined event
//    // ordering
//    addSucc2( mkReq1( nil, nil ), nil )
//    addSucc2( mkReq1( nil, auth1 ), auth1 )
//    // check implicit params with and without auth and with need for reordering
//    addSucc2(
//        flattenEvs( evReqTyp, 
//            evFldSvc, evSvc1, 
//            evFldOp, evOp1, 
//            evFldNs, evNs1,
//            NewEndEvent(),
//        ),
//        nil,
//    )
//    addSucc2(
//        flattenEvs( evReqTyp,
//            evFldSvc, evSvc1,
//            evFldAuth, evAuth1,
//            evFldOp, evOp1,
//            evFldNs, evNs1,
//            NewEndEvent(),
//        ),
//        auth1,
//    )
//    addPathSucc := func( 
//        paramsIn, paramsExpct *mg.SymbolMap, paramEvs []EventExpectation,
//        auth mg.Value, authEvs []EventExpectation ) {
//        t := &RequestReactorTest{
//            Namespace: ns1,
//            Service: svc1,
//            Operation: op1,
//            Parameters: paramsExpct,
//            ParameterEvents: paramEvs,
//            Authentication: auth,
//            AuthenticationEvents: authEvs,
//        }
//        pairs := []interface{}{
//            IdNamespace, ns1.ExternalForm(),
//            IdService, svc1.ExternalForm(),
//            IdOperation, op1.ExternalForm(),
//        }
//        if paramsIn != nil { pairs = append( pairs, IdParameters, paramsIn ) }
//        if auth != nil { pairs = append( pairs, IdAuthentication, auth ) }
//        t.Source = mg.MustStruct( QnameRequest, pairs... )
//        AddStdReactorTests( t )
//    }
//    pathParams := objpath.RootedAt( IdParameters )
//    evsEmptyParams := []EventExpectation{ 
//        { idFact.NextMapStart(), pathParams }, { NewEndEvent(), pathParams } }
//    pathAuth := objpath.RootedAt( IdAuthentication )
//    addPathSucc( nil, mg.MustSymbolMap(), evsEmptyParams, nil, nil )
//    addPathSucc( mg.MustSymbolMap(), mg.MustSymbolMap(), evsEmptyParams, nil, nil )
//    idF1 := id( "f1" )
//    addPathSucc(
//        mg.MustSymbolMap( idF1, mg.Int32( 1 ) ),
//        mg.MustSymbolMap( idF1, mg.Int32( 1 ) ),
//        []EventExpectation{
//            { idFact.NextMapStart(), pathParams },
//            { evFldF1, pathParams.Descend( idF1 ) },
//            { i32Val1, pathParams.Descend( idF1 ) },
//            { NewEndEvent(), pathParams },
//        },
//        nil, nil,
//    )
//    addPathSucc( 
//        nil, mg.MustSymbolMap(), evsEmptyParams,
//        mg.Int32( 1 ), []EventExpectation{ { i32Val1, pathAuth } },
//    )
//    addPathSucc(
//        nil, mg.MustSymbolMap(), evsEmptyParams,
//        auth1, []EventExpectation{
//            { NewStructStartEvent( authQn ), pathAuth },
//            { evFldF1, pathAuth.Descend( idF1 ) },
//            { i32Val1, pathAuth.Descend( idF1 ) },
//            { NewEndEvent(), pathAuth },
//        },
//    )
//    writeMgIo := func( f func( w *BinWriter ) ) Buffer {
//        bb := &bytes.Buffer{}
//        w := NewWriter( bb )
//        f( w )
//        return Buffer( bb.Bytes() )
//    }
//    nsBuf := func( ns *Namespace ) Buffer {
//        return writeMgIo( func( w *BinWriter ) { w.WriteNamespace( ns ) } )
//    }
//    idBuf := func( id *mg.Identifier ) Buffer {
//        return writeMgIo( func( w *BinWriter ) { w.WriteIdentifier( id ) } )
//    }
//    AddStdReactorTests(
//        &RequestReactorTest{
//            Namespace: ns1,
//            Service: svc1,
//            Operation: op1,
//            Parameters: EmptySymbolMap(),
//            Source: mg.MustStruct( QnameRequest,
//                IdNamespace, nsBuf( ns1 ),
//                IdService, idBuf( svc1 ),
//                IdOperation, idBuf( op1 ),
//            ),
//        },
//    )
//    AddStdReactorTests(
//        &RequestReactorTest{
//            Source: mg.MustStruct( "ns1@v1/S1" ),
//            Error: mg.NewTypeCastError(
//                TypeRequest, MustTypeReference( "ns1@v1/S1" ), nil ),
//        },
//    )
//    createReqVcErr := func( 
//        val interface{}, path idPath, msg string ) *RequestReactorTest {
//
//        return &RequestReactorTest{
//            Source: MustValue( val ),
//            Error: mg.NewValueCastError( path, msg ),
//        }
//    }
//    addReqVcErr := func( val interface{}, path idPath, msg string ) {
//        AddStdReactorTests( createReqVcErr( val, path, msg ) )
//    }
//    addReqVcErr(
//        mg.MustSymbolMap( IdNamespace, true ), 
//        objpath.RootedAt( IdNamespace ),
//        "invalid value: mingle:core@v1/Boolean",
//    )
//    addReqVcErr(
//        mg.MustSymbolMap( IdNamespace, mg.MustSymbolMap() ),
//        objpath.RootedAt( IdNamespace ),
//        "invalid value: mingle:core@v1/SymbolMap",
//    )
//    addReqVcErr(
//        mg.MustSymbolMap( IdNamespace, mg.MustStruct( "ns1@v1/S1" ) ),
//        objpath.RootedAt( IdNamespace ),
//        "invalid value: ns1@v1/S1",
//    )
//    addReqVcErr(
//        mg.MustSymbolMap( IdNamespace, mg.MustList() ),
//        objpath.RootedAt( IdNamespace ),
//        "invalid value: mingle:core@v1/Value?*",
//    )
//    func() {
//        test := createReqVcErr(
//            mg.MustSymbolMap( IdNamespace, ns1.ExternalForm(), IdService, true ),
//            objpath.RootedAt( IdService ),
//            "invalid value: mingle:core@v1/Boolean",
//        )
//        test.Namespace = ns1
//        AddStdReactorTests( test )
//    }()
//    func() {
//        test := createReqVcErr(
//            mg.MustSymbolMap( 
//                IdNamespace, ns1.ExternalForm(),
//                IdService, svc1.ExternalForm(),
//                IdOperation, true,
//            ),
//            objpath.RootedAt( IdOperation ),
//            "invalid value: mingle:core@v1/Boolean",
//        )
//        test.Namespace, test.Service = ns1, svc1
//        AddStdReactorTests( test )
//    }()
//    AddStdReactorTests(
//        &RequestReactorTest{
//            Source: mg.MustSymbolMap(
//                IdNamespace, ns1.ExternalForm(),
//                IdService, svc1.ExternalForm(),
//                IdOperation, op1.ExternalForm(),
//                IdParameters, true,
//            ),
//            Namespace: ns1,
//            Service: svc1,
//            Operation: op1,
//            Error: mg.NewTypeCastError(
//                mg.TypeSymbolMap,
//                TypeBoolean,
//                objpath.RootedAt( IdParameters ),
//            ),
//        },
//    )
//    // Check that errors are bubbled up from
//    // *BinWriter.Read(Identfier|Namespace) when parsing invalid
//    // namespace/service/operation Buffers
//    createBinRdErr := func( path *mg.Identifier, msg string, 
//        pairs ...interface{} ) *RequestReactorTest {
//
//        return createReqVcErr(
//            mg.MustSymbolMap( pairs... ), objpath.RootedAt( path ), msg )
//    }
//    addBinRdErr := func( path *mg.Identifier, msg string, pairs ...interface{} ) {
//        AddStdReactorTests( createBinRdErr( path, msg, pairs... ) )
//    }
//    badBuf := []byte{ 0x0f }
//    addBinRdErr( 
//        IdNamespace, 
//        "Expected type code 0x02 but got 0x0f",
//        IdNamespace, badBuf )
//    func() {
//        test := createBinRdErr(
//            IdService, 
//            "Expected type code 0x01 but got 0x0f",
//            IdNamespace, ns1.ExternalForm(), 
//            IdService, badBuf,
//        )
//        test.Namespace = ns1
//        AddStdReactorTests( test )
//    }()
//    func() {
//        test := createBinRdErr(
//            IdOperation, 
//            "Expected type code 0x01 but got 0x0f",
//            IdNamespace, ns1.ExternalForm(),
//            IdService, svc1.ExternalForm(),
//            IdOperation, badBuf,
//        )
//        test.Namespace, test.Service = ns1, svc1
//        AddStdReactorTests( test )
//    }()
//    addReqVcErr(
//        mg.MustSymbolMap( IdNamespace, "ns1::ns2" ),
//        objpath.RootedAt( IdNamespace ),
//        "[<input>, line 1, col 5]: Illegal start of identifier part: \":\" " +
//        "(U+003A)",
//    )
//    func() {
//        test := createReqVcErr(
//            mg.MustSymbolMap( IdNamespace, ns1.ExternalForm(), IdService, "2bad" ),
//            objpath.RootedAt( IdService ),
//            "[<input>, line 1, col 1]: Illegal start of identifier part: " +
//            "\"2\" (U+0032)",
//        )
//        test.Namespace = ns1
//        AddStdReactorTests( test )
//    }()
//    func() {
//        test := createReqVcErr(
//            mg.MustSymbolMap(
//                IdNamespace, ns1.ExternalForm(),
//                IdService, svc1.ExternalForm(),
//                IdOperation, "2bad",
//            ),
//            objpath.RootedAt( IdOperation ),
//            "[<input>, line 1, col 1]: Illegal start of identifier part: " +
//            "\"2\" (U+0032)",
//        )
//        test.Namespace, test.Service = ns1, svc1
//        AddStdReactorTests( test )
//    }()
//    t1Bad := qname( "foo@v1/Request" )
//    AddStdReactorTests(
//        &RequestReactorTest{
//            Source: mg.MustStruct( t1Bad ),
//            Error: mg.NewTypeCastError(
//                TypeRequest, t1Bad.AsAtomicType(), nil ),
//        },
//    )
//    // Not exhaustively re-testing all ways a field could be missing (assume for
//    // now that field order tests will handle that). Instead, we are just
//    // getting basic coverage that the field order supplied by the request
//    // reactor is in fact being set up correctly and that we have set up the
//    // right required fields.
//    AddStdReactorTests(
//        &RequestReactorTest{
//            Source: mg.MustSymbolMap( 
//                IdNamespace, ns1.ExternalForm(),
//                IdOperation, op1.ExternalForm(),
//            ),
//            Namespace: ns1,
//            Error: mg.NewMissingFieldsError( nil, []*mg.Identifier{ IdService } ),
//        },
//    )
//}
//
//type ResponseReactorTest struct {
//    In mg.Value
//    ResVal mg.Value
//    ResEvents []EventExpectation
//    ErrVal mg.Value
//    ErrEvents []EventExpectation
//    Error error
//}
//
//func initResponseTests() {
//    idFact := NewTestPointerIdFactory()
//    addSucc := func( in, res, err mg.Value ) {
//        AddStdReactorTests(
//            &ResponseReactorTest{ In: in, ResVal: res, ErrVal: err } )
//    }
//    i32Val1 := mg.Int32( 1 )
//    err1 := mg.MustStruct( "ns1@v1/Err1", "f1", int32( 1 ) )
//    addSucc( mg.MustStruct( QnameResponse ), nil, nil )
//    addSucc( mg.MustSymbolMap(), nil, nil )
//    addSucc( mg.MustSymbolMap( IdResult, mg.NullVal ), mg.NullVal, nil )
//    addSucc( mg.MustSymbolMap( IdResult, i32Val1 ), i32Val1, nil )
//    addSucc( mg.MustSymbolMap( IdError, mg.NullVal ), nil, mg.NullVal )
//    addSucc( mg.MustSymbolMap( IdError, err1 ), nil, err1 )
//    addSucc( mg.MustSymbolMap( IdError, int32( 1 ) ), nil, i32Val1 )
//    pathRes := objpath.RootedAt( IdResult )
//    pathResF1 := pathRes.Descend( id( "f1" ) )
//    pathErr := objpath.RootedAt( IdError )
//    pathErrF1 := pathErr.Descend( id( "f1" ) )
//    AddStdReactorTests(
//        &ResponseReactorTest{
//            In: mg.MustStruct( QnameResponse, "result", int32( 1 ) ),
//            ResVal: i32Val1,
//            ResEvents: []EventExpectation{ 
//                { NewValueEvent( i32Val1 ), pathRes },
//            },
//        },
//        &ResponseReactorTest{
//            In: mg.MustSymbolMap( "result", mg.MustSymbolMap( "f1", int32( 1 ) ) ),
//            ResVal: mg.MustSymbolMap( "f1", int32( 1 ) ),
//            ResEvents: []EventExpectation{
//                { idFact.NextMapStart(), pathRes },
//                { NewFieldStartEvent( id( "f1" ) ), pathResF1 },
//                { NewValueEvent( i32Val1 ), pathResF1 },
//                { NewEndEvent(), pathRes },
//            },
//        },
//        &ResponseReactorTest{
//            In: mg.MustSymbolMap( "error", int32( 1 ) ),
//            ErrVal: i32Val1,
//            ErrEvents: []EventExpectation{ 
//                { NewValueEvent( i32Val1 ), pathErr },
//            },
//        },
//        &ResponseReactorTest{
//            In: mg.MustSymbolMap( "error", err1 ),
//            ErrVal: err1,
//            ErrEvents: []EventExpectation{
//                { NewStructStartEvent( err1.Type ), pathErr },
//                { NewFieldStartEvent( id( "f1" ) ), pathErrF1 },
//                { NewValueEvent( i32Val1 ), pathErrF1 },
//                { NewEndEvent(), pathErr },
//            },
//        },
//    )
//    addFail := func( in mg.Value, err error ) {
//        AddStdReactorTests( &ResponseReactorTest{ In: in, Error: err } )
//    }
//    addFail(
//        err1.Fields,
//        NewUnrecognizedFieldError( nil, id( "f1" ) ),
//    )
//    addFail(
//        mg.MustStruct( "ns1@v1/Response" ),
//        mg.NewTypeCastError( 
//            TypeResponse, MustTypeReference( "ns1@v1/Response" ), nil ),
//    )
//    addFail(
//        mg.MustSymbolMap( IdResult, i32Val1, IdError, err1 ),
//        mg.NewValueCastError( 
//            nil, "response has both a result and an error value" ),
//    )
//}
//
//func initServiceTests() {
//    initRequestTests()
//    initResponseTests()
//}
//
//type CastReactorTest struct {
//    In interface{}
//    Expect mg.Value
//    Type mg.TypeReference
//    Path objpath.PathNode
//    Err error
//    Profile string
//}
//
//var crtPathDefault = objpath.RootedAt( MustIdentifier( "inVal" ) )
//
//type crtInit struct {
//    buf1 Buffer
//    tm1 mg.Timestamp
//    map1 *mg.SymbolMap
//    en1 *Enum
//    struct1 *mg.Struct
//}
//
//func ( t *crtInit ) initStdVals() {
//    t.buf1 = Buffer( []byte{ byte( 0 ), byte( 1 ), byte( 2 ) } )
//    t.tm1 = MustTimestamp( "2007-08-24T13:15:43.123450000-08:00" )
//    t.map1 = mg.MustSymbolMap( "key1", 1, "key2", "val2" )
//    t.en1 = MustEnum( "ns1@v1/E1", "en-val1" )
//    t.struct1 = mg.MustStruct( "ns1@v1/S1", "key1", "val1" )
//}
//
//func ( t *crtInit ) addCrt( crt *CastReactorTest ) { 
//    StdReactorTests = append( StdReactorTests, crt ) 
//}
//
//func ( t *crtInit ) addCrtDefault( crt *CastReactorTest ) {
//    crt.Path = crtPathDefault
//    t.addCrt( crt )
//}
//
//func ( t *crtInit ) createSucc(
//    in, expct interface{}, typ TypeReferenceInitializer ) *CastReactorTest {
//    return &CastReactorTest{ 
//        In: MustValue( in ), 
//        Expect: MustValue( expct ), 
//        Type: asTypeReference( typ ),
//    }
//}
//
//func ( t *crtInit ) addSucc( 
//    in, expct interface{}, typ TypeReferenceInitializer ) {
//    t.addCrtDefault( t.createSucc( in, expct, typ ) )
//}
//
//func ( t *crtInit ) addIdent( in interface{}, typ TypeReferenceInitializer ) {
//    v := MustValue( in )
//    t.addSucc( v, v, asTypeReference( typ ) )
//}
//
//func ( t *crtInit ) addBaseTypeTests() {
//    t.addIdent( Boolean( true ), TypeBoolean )
//    t.addIdent( t.buf1, mg.TypeBuffer )
//    t.addIdent( "s", mg.TypeString )
//    t.addIdent( mg.Int32( 1 ), mg.TypeInt32 )
//    t.addIdent( mg.Int64( 1 ), mg.TypeInt64 )
//    t.addIdent( Uint32( 1 ), TypeUint32 )
//    t.addIdent( Uint64( 1 ), TypeUint64 )
//    t.addIdent( mg.Float32( 1.0 ), mg.TypeFloat32 )
//    t.addIdent( mg.Float64( 1.0 ), mg.TypeFloat64 )
//    t.addIdent( t.tm1, TypeTimestamp )
//    t.addIdent( t.en1, t.en1.Type )
//    t.addIdent( t.map1, mg.TypeSymbolMap )
//    t.addIdent( t.struct1, t.struct1.Type )
//    t.addIdent( nil, mg.TypeNullableValue )
//    t.addSucc( mg.Int32( -1 ), Uint32( uint32( 4294967295 ) ), TypeUint32 )
//    t.addSucc( mg.Int64( -1 ), Uint32( uint32( 4294967295 ) ), TypeUint32 )
//    t.addSucc( 
//        mg.Int32( -1 ), Uint64( uint64( 18446744073709551615 ) ), TypeUint64 )
//    t.addSucc( 
//        mg.Int64( -1 ), Uint64( uint64( 18446744073709551615 ) ), TypeUint64 )
//    t.addSucc( "true", true, TypeBoolean )
//    t.addSucc( "TRUE", true, TypeBoolean )
//    t.addSucc( "TruE", true, TypeBoolean )
//    t.addSucc( "false", false, TypeBoolean )
//    t.addSucc( true, "true", mg.TypeString )
//    t.addSucc( false, "false", mg.TypeString )
//}
//
//func ( t *crtInit ) createTcError0(
//    in interface{}, 
//    typExpct, typAct, callTyp TypeReferenceInitializer, 
//    p idPath ) *CastReactorTest {
//
//    return &CastReactorTest{
//        In: MustValue( in ),
//        Type: asTypeReference( callTyp ),
//        Err: mg.NewTypeCastError( 
//            asTypeReference( typExpct ),
//            asTypeReference( typAct ),
//            p,
//        ),
//    }
//}
//
//func ( t *crtInit ) addTcError0(
//    in interface{}, 
//    typExpct, typAct, callTyp TypeReferenceInitializer, 
//    p idPath ) {
//
//    t.addCrtDefault( t.createTcError0( in, typExpct, typAct, callTyp, p ) )
//}
//
//func ( t *crtInit ) createTcError(
//    in interface{}, 
//    typExpct, typAct TypeReferenceInitializer ) *CastReactorTest {
//    return t.createTcError0( in, typExpct, typAct, typExpct, crtPathDefault )
//}
//
//func ( t *crtInit ) addTcError(
//    in interface{}, typExpct, typAct TypeReferenceInitializer ) {
//    t.addTcError0( in, typExpct, typAct, typExpct, crtPathDefault )
//}
//
//func ( t *crtInit ) addMiscTcErrors() {
//    t.addTcError( t.en1, "ns1@v1/Bad", t.en1.Type )
//    t.addTcError( t.struct1, "ns1@v1/Bad", t.struct1.Type )
//    t.addTcError( "s", mg.TypeNull, mg.TypeString )
//    t.addTcError( int32( 1 ), "Buffer", "Int32" )
//    t.addTcError( int32( 1 ), "Buffer?", "Int32" )
//    t.addTcError( true, "Float32", "Boolean" )
//    t.addTcError( true, "&Float32", "Boolean" )
//    t.addTcError( true, "&Float32?", "Boolean" )
//    t.addTcError( true, "Int32", "Boolean" )
//    t.addTcError( true, "&Int32", "Boolean" )
//    t.addTcError( true, "&Int32?", "Boolean" )
//    t.addTcError( mg.MustList( 1, 2 ), mg.TypeString, mg.TypeOpaqueList )
//    t.addTcError( mg.MustList(), "String?", mg.TypeOpaqueList )
//    t.addTcError( "s", "String*", "String" )
//    t.addCrtDefault(
//        &CastReactorTest{
//            In: mg.MustList( 1, t.struct1 ),
//            Type: asTypeReference( "Int32*" ),
//            Err: mg.NewTypeCastError(
//                asTypeReference( "Int32" ),
//                &mg.AtomicTypeReference{ Name: t.struct1.Type },
//                crtPathDefault.StartList().Next(),
//            ),
//        },
//    )
//    t.addTcError( t.struct1, "&Int32?", t.struct1.Type )
//    t.addTcError( 12, t.struct1.Type, "Int64" )
//    for _, prim := range PrimitiveTypes {
//        // not an err for prims mg.Value and mg.SymbolMap
//        if prim != mg.TypeSymbolMap { 
//            t.addTcError( t.struct1, prim, t.struct1.Type )
//        }
//    }
//}
//
//func ( t *crtInit ) createVcError0(
//    val interface{}, 
//    typ TypeReferenceInitializer, 
//    path idPath, 
//    msg string ) *CastReactorTest {
//    return &CastReactorTest{
//        In: MustValue( val ),
//        Type: asTypeReference( typ ),
//        Err: mg.NewValueCastError( path, msg ),
//    }
//}
//    
//
//func ( t *crtInit ) addVcError0( 
//    val interface{}, typ TypeReferenceInitializer, path idPath, msg string ) {
//    t.addCrtDefault( t.createVcError0( val, typ, path, msg ) )
//}
//
//func ( t *crtInit ) createVcError(
//    val interface{}, 
//    typ TypeReferenceInitializer, 
//    msg string ) *CastReactorTest {
//    return t.createVcError0( val, typ, crtPathDefault, msg )
//}
//
//func ( t *crtInit ) addVcError( 
//    val interface{}, typ TypeReferenceInitializer, msg string ) {
//    t.addVcError0( val, typ, crtPathDefault, msg )
//}
//
//func ( t *crtInit ) addNullValueError(
//    val interface{}, typ TypeReferenceInitializer ) {
//
//    t.addVcError( val, typ, "Value is null" )
//}
//
//func ( t *crtInit ) addMiscVcErrors() {
//    t.addVcError( "s", TypeBoolean, `Invalid boolean value: "s"` )
//    t.addVcError( nil, mg.TypeString, "Value is null" )
//    t.addVcError( nil, `String~"a"`, "Value is null" )
//    t.addVcError( mg.MustList(), "String+", "empty list" )
//    t.addVcError0( 
//        mg.MustList( mg.MustList( int32( 1 ), int32( 2 ) ), mg.MustList() ), 
//        "Int32+*", 
//        crtPathDefault.StartList().Next(),
//        "empty list",
//    )
//}
//
//func ( t *crtInit ) addPtrRefSucc(
//    typExpct TypeReferenceInitializer, valExpct mg.Value, evs ...ReactorEvent ) {
//
//    AddStdReactorTests(
//        &CastReactorTest{
//            Expect: valExpct,
//            In: mg.CopySource( evs ),
//            Type: asTypeReference( typExpct ),
//        },
//    )
//}
//
//func ( t *crtInit ) addPtrRefFail0(
//    typExpct, typAct TypeReferenceInitializer, 
//    path objpath.PathNode,
//    evs ...ReactorEvent ) {
//    
//    expct, act := asTypeReference( typExpct ), asTypeReference( typAct )
//    AddStdReactorTests(
//        &CastReactorTest{
//            In: mg.CopySource( evs ),
//            Type: expct,
//            Path: path,
//            Err: NewValueCastErrorf( crtPathDefault,
//                "expected %s but got a reference to %s", expct, act ),
//        },
//    )
//}
//
//func ( t *crtInit ) addPtrRefFail( 
//    typExpct, typAct TypeReferenceInitializer, evs ...ReactorEvent ) {
//
//    t.addPtrRefFail0( typExpct, typAct, crtPathDefault, evs... )
//}
//
//func ( t *crtInit ) addStringTests() {
//    t.addIdent( "s", "String?" )
//    t.addIdent( "abbbc", `String~"^ab+c$"` )
//    t.addIdent( "abbbc", `String~"^ab+c$"?` )
//    t.addIdent( nil, `String~"^ab+c$"?` )
//    t.addIdent( "", `String~"^a*"?` )
//    t.addSucc( 
//        mg.MustList( "123", 129 ), 
//        mg.MustList( "123", "129" ),
//        `String~"^\\d+$"*`,
//    )
//    for _, quant := range []string { "*", "+", "?*", "*?" } {
//        val := mg.MustList( "a", "aaaaaa" )
//        t.addSucc( val, val, `String~"^a+$"` + quant )
//    }
//    t.addVcError( 
//        "ac", 
//        `String~"^ab+c$"`,
//        `Value "ac" does not satisfy restriction "^ab+c$"`,
//    )
//    t.addVcError(
//        "ab",
//        `String~"^a*$"?`,
//        "Value \"ab\" does not satisfy restriction \"^a*$\"",
//    )
//    t.addVcError0(
//        mg.MustList( "a", "b" ),
//        `String~"^a+$"*`,
//        crtPathDefault.StartList().Next(),
//        "Value \"b\" does not satisfy restriction \"^a+$\"",
//    )
//    t.addTcError( EmptySymbolMap(), mg.TypeString, mg.TypeSymbolMap )
//    t.addTcError( EmptyList(), mg.TypeString, mg.TypeOpaqueList )
//}
//
//func ( t *crtInit ) addIdentityNumTests() {
//    t.addIdent( int64( 1 ), "Int64~[-1,1]" )
//    t.addIdent( int64( 1 ), "Int64~(,2)" )
//    t.addIdent( int64( 1 ), "Int64~[1,1]" )
//    t.addIdent( int64( 1 ), "Int64~[-2, 32)" )
//    t.addIdent( int32( 1 ), "Int32~[-2, 32)" )
//    t.addIdent( uint32( 3 ), "Uint32~[2,32)" )
//    t.addIdent( uint64( 3 ), "Uint64~[2,32)" )
//    t.addIdent( mg.Float32( -1.1 ), "Float32~[-2.0,32)" )
//    t.addIdent( mg.Float64( -1.1 ), "Float64~[-2.0,32)" )
//    numTests := []struct{ val Value; str string; typ mg.TypeReference } {
//        { val: mg.Int32( 1 ), str: "1", typ: mg.TypeInt32 },
//        { val: mg.Int64( 1 ), str: "1", typ: mg.TypeInt64 },
//        { val: Uint32( 1 ), str: "1", typ: TypeUint32 },
//        { val: Uint64( 1 ), str: "1", typ: TypeUint64 },
//        { val: mg.Float32( 1.0 ), str: "1", typ: mg.TypeFloat32 },
//        { val: mg.Float64( 1.0 ), str: "1", typ: mg.TypeFloat64 },
//    }
//    s1 := mg.MustStruct( "ns1@v1/S1" )
//    for _, numCtx := range numTests {
//        t.addSucc( numCtx.val, numCtx.str, mg.TypeString )
//        t.addSucc( numCtx.str, numCtx.val, numCtx.typ )
//        ptrVal := mg.NewHeapValue( numCtx.val )
//        ptrTyp := NewPointerTypeReference( numCtx.typ )
//        t.addSucc( numCtx.val, ptrVal, ptrTyp )
//        t.addSucc( numCtx.str, ptrVal, ptrTyp )
//        t.addSucc( ptrVal, numCtx.str, mg.TypeString )
//        t.addSucc( ptrVal, numCtx.val, numCtx.typ )
//        t.addTcError( EmptySymbolMap(), numCtx.typ, mg.TypeSymbolMap )
//        t.addTcError( EmptySymbolMap(), ptrTyp, mg.TypeSymbolMap )
//        t.addVcError( nil, numCtx.typ, "Value is null" )
//        t.addTcError( EmptyList(), numCtx.typ, mg.TypeOpaqueList )
//        t.addTcError( t.buf1, numCtx.typ, mg.TypeBuffer )
//        t.addCrtDefault(
//            t.createTcError0(
//                mg.NewHeapValue( t.buf1 ),
//                ptrTyp.Type,
//                mg.TypeBuffer,
//                ptrTyp,
//                crtPathDefault,
//            ),
//        )
//        t.addTcError( s1, numCtx.typ, s1.Type )
//        t.addTcError( ptrVal, s1.Type, numCtx.typ )
//        t.addTcError( s1, ptrTyp, s1.Type )
//        for _, valCtx := range numTests {
//            t.addSucc( valCtx.val, numCtx.val, numCtx.typ )
//            t.addSucc( mg.NewHeapValue( valCtx.val ), numCtx.val, numCtx.typ )
//            t.addSucc( valCtx.val, ptrVal, ptrTyp )
//            t.addSucc( mg.NewHeapValue( valCtx.val ), ptrVal, ptrTyp )
//        }
//    }
//}
//
//func ( t *crtInit ) addTruncateNumTests() {
//    posVals := []mg.Value{ mg.Float32( 1.1 ), mg.Float64( 1.1 ), mg.String( "1.1" ) }
//    for _, val := range posVals {
//        t.addSucc( val, mg.Int32( 1 ), mg.TypeInt32 )
//        t.addSucc( val, mg.Int64( 1 ), mg.TypeInt64 )
//        t.addSucc( val, Uint32( 1 ), TypeUint32 )
//        t.addSucc( val, Uint64( 1 ), TypeUint64 )
//    }
//    negVals := []mg.Value{ mg.Float32( -1.1 ), mg.Float64( -1.1 ), mg.String( "-1.1" ) }
//    for _, val := range negVals {
//        t.addSucc( val, mg.Int32( -1 ), mg.TypeInt32 )
//        t.addSucc( val, mg.Int64( -1 ), mg.TypeInt64 )
//    }
//    t.addSucc( int64( 1 << 31 ), int32( -2147483648 ), mg.TypeInt32 )
//    t.addSucc( int64( 1 << 33 ), int32( 0 ), mg.TypeInt32 )
//    t.addSucc( int64( 1 << 31 ), uint32( 1 << 31 ), TypeUint32 )
//}
//
//func ( t *crtInit ) addNumTests() {
//    for _, typ := range NumericTypes {
//        t.addVcError( "not-a-num", typ, `invalid syntax: not-a-num` )
//    }
//    t.addIdentityNumTests()
//    t.addTruncateNumTests()
//    t.addSucc( "1", int64( 1 ), "Int64~[-1,1]" ) // just cover mg.String with range
//    rngErr := func( val string, typ mg.TypeReference ) {
//        t.addVcError( val, typ, fmt.Sprintf( "value out of range: %s", val ) )
//    }
//    rngErr( "2147483648", mg.TypeInt32 )
//    rngErr( "-2147483649", mg.TypeInt32 )
//    rngErr( "9223372036854775808", mg.TypeInt64 )
//    rngErr( "-9223372036854775809", mg.TypeInt64 )
//    rngErr( "4294967296", TypeUint32 )
//    t.addVcError( "-1", TypeUint32, "value out of range: -1" )
//    t.addVcError( "-1", NewPointerTypeReference( TypeUint32 ), 
//        "value out of range: -1" )
//    rngErr( "18446744073709551616", TypeUint64 )
//    t.addVcError( "-1", TypeUint64, "value out of range: -1" )
//    for _, tmpl := range []string{ "%s", "&%s", "%s?" } {
//        t.addVcError(
//            12, fmt.Sprintf( tmpl, "Int32~[0,10)" ), 
//            "Value 12 does not satisfy restriction [0,10)" )
//    }
//}
//
//func ( t *crtInit ) addBufferTests() {
//    buf1B64 := mg.String( base64.StdEncoding.EncodeToString( t.buf1 ) )
//    t.addSucc( t.buf1, buf1B64, mg.TypeString )
//    t.addSucc( mg.NewHeapValue( t.buf1 ), buf1B64, mg.TypeString )
//    t.addSucc( mg.NewHeapValue( t.buf1 ), mg.NewHeapValue( buf1B64 ),
//        NewPointerTypeReference( mg.TypeString ) )
//    t.addSucc( buf1B64, t.buf1, mg.TypeBuffer  )
//    t.addSucc( mg.NewHeapValue( buf1B64 ), t.buf1, mg.TypeBuffer )
//    t.addSucc( mg.NewHeapValue( buf1B64 ), mg.NewHeapValue( t.buf1 ),
//        NewPointerTypeReference( mg.TypeBuffer ) )
//    t.addVcError( "abc$/@", mg.TypeBuffer, 
//        "Invalid base64 string: illegal base64 data at input byte 3" )
//}
//
//func ( t *crtInit ) addTimeTests() {
//    t.addIdent(
//        Now(), `Timestamp~["1970-01-01T00:00:00Z","2200-01-01T00:00:00Z"]` )
//    t.addSucc( t.tm1, t.tm1.Rfc3339Nano(), mg.TypeString )
//    t.addSucc( t.tm1.Rfc3339Nano(), t.tm1, TypeTimestamp )
//    t.addVcError(
//        MustTimestamp( "2012-01-01T00:00:00Z" ),
//        `mingle:core@v1/Timestamp~` +
//            `["2000-01-01T00:00:00Z","2001-01-01T00:00:00Z"]`,
//        "Value 2012-01-01T00:00:00Z does not satisfy restriction " +
//            "[\"2000-01-01T00:00:00Z\",\"2001-01-01T00:00:00Z\"]",
//    )
//}
//
//func ( t *crtInit ) addEnumTests() {
//    ptrTyp :=
//        NewPointerTypeReference( &mg.AtomicTypeReference{ Name: t.en1.Type } )
//    t.addSucc( mg.NewHeapValue( t.en1 ), mg.NewHeapValue( t.en1 ), ptrTyp )
//    t.addSucc( t.en1, mg.NewHeapValue( t.en1 ), ptrTyp )
//    t.addSucc( t.en1, "en-val1", mg.TypeString  )
//    t.addSucc( t.en1, mg.NewHeapValue( MustValue( "en-val1" ) ), 
//        NewPointerTypeReference( mg.TypeString ) )
//    t.addTcError( EmptySymbolMap(), t.en1.Type, mg.TypeSymbolMap )
//    t.addNullValueError( nil, t.en1.Type )
//    t.addTcError( t.en1, "ns1@v1/E2", t.en1.Type )
//    t.addTcError( mg.NewHeapValue( t.en1 ), "ns1@v1/E2", t.en1.Type )
//    t.addTcError( t.en1, "&ns1@v1/E2", t.en1.Type )
//    t.addCrtDefault(
//        t.createTcError0(
//            mg.NewHeapValue( t.en1 ),
//            "ns1@v1/E2",
//            "ns1@v1/E1",
//            "&ns1@v1/E2",
//            crtPathDefault,
//        ),
//    )
//}
//
//func ( t *crtInit ) addNullableTests() {
//    typs := []mg.TypeReference{}
//    addNullSucc := func( expct interface{}, typ mg.TypeReference ) {
//        t.addSucc( nil, expct, typ )
//    }
//    for _, prim := range PrimitiveTypes {
//        if isNullableType( prim ) {
//            typs = append( typs, MustNullableTypeReference( prim ) )
//        } else {
//            t.addNullValueError( nil, prim )
//        }
//    }
//    typs = append( typs,
//        MustTypeReference( "&Null?" ),
//        MustTypeReference( "String?" ),
//        MustTypeReference( "String*?" ),
//        MustTypeReference( "&Int32?*?" ),
//        MustTypeReference( "String+?" ),
//        MustTypeReference( "&ns1@v1/T?" ),
//        MustTypeReference( "ns1@v1/T*?" ),
//    )
//    for _, typ := range typs { addNullSucc( nil, typ ) }
//}
//
//func ( t *crtInit ) addListTests() {
//    for _, quant := range []string{ "*", "**", "***" } {
//        t.addSucc( []interface{}{}, mg.MustList(), "Int64" + quant )
//    }
//    for _, quant := range []string{ "**", "*+" } {
//        v := mg.MustList( mg.MustList(), mg.MustList() )
//        t.addIdent( v, "Int64" + quant )
//    }
//    // test conversions in a deeply nested list
//    t.addSucc(
//        []interface{}{
//            []interface{}{ "1", int32( 1 ), int64( 1 ) },
//            []interface{}{ float32( 1.0 ), float64( 1.0 ) },
//            []interface{}{},
//        },
//        mg.MustList(
//            mg.MustList( mg.Int64( 1 ), mg.Int64( 1 ), mg.Int64( 1 ) ),
//            mg.MustList( mg.Int64( 1 ), mg.Int64( 1 ) ),
//            mg.MustList(),
//        ),
//        "Int64**",
//    )
//    t.addSucc(
//        []interface{}{ int64( 1 ), nil, "hi" },
//        mg.MustList( "1", nil, "hi" ),
//        "String?*",
//    )
//    s1 := mg.MustStruct( "ns1@v1/S1" )
//    t.addSucc(
//        []interface{}{ s1, mg.NewHeapValue( s1 ), nil },
//        mg.MustList( mg.NewHeapValue( s1 ), mg.NewHeapValue( s1 ), mg.NullVal ),
//        "&ns1@v1/S1?*",
//    )
//    t.addVcError0(
//        []interface{}{ mg.NewHeapValue( s1 ), nil },
//        "&ns1@v1/S1*",
//        crtPathDefault.StartList().SetIndex( 1 ),
//        "Value is null",
//    )
//    t.addVcError0(
//        []interface{}{ s1, nil },
//        "ns1@v1/S1*",
//        crtPathDefault.StartList().SetIndex( 1 ),
//        "Value is null",
//    )
//    t.addSucc(
//        []interface{}{ 
//            int32( 1 ), 
//            []interface{}{}, 
//            []interface{}{ int32( 1 ), int32( 2 ), int32( 3 ) },
//            "s1", 
//            s1, 
//            nil,
//        },
//        mg.MustList(
//            mg.NewHeapValue( mg.Int32( 1 ) ),
//            mg.NewHeapValue( mg.MustList() ),
//            mg.NewHeapValue( mg.MustList( int32( 1 ), int32( 2 ), int32( 3 ) ) ),
//            mg.NewHeapValue( mg.String( "s1" ) ),
//            mg.NewHeapValue( s1 ),
//            mg.NullVal,
//        ),
//        "&Value?*",
//    )
//    t.addSucc( mg.MustList(), mg.MustList(), mg.TypeValue )
//    intList1 := mg.MustList( int32( 1 ), int32( 2 ), int32( 3 ) )
//    t.addSucc( intList1, intList1, mg.TypeValue )
//    t.addSucc( intList1, intList1, mg.TypeOpaqueList )
//    t.addSucc( intList1, intList1, "Int32*?" )
//    t.addSucc( 
//        mg.MustList(), 
//        mg.NewHeapValue( mg.MustList() ), 
//        NewPointerTypeReference( MustTypeReference( "&Int32*" ) ),
//    )
//    t.addSucc( 
//        mg.NewHeapValue( mg.MustList() ), 
//        mg.NewHeapValue( mg.MustList() ),
//        NewPointerTypeReference( MustTypeReference( "&Int32*" ) ),
//    )
//    t.addSucc( mg.NewHeapValue( mg.MustList() ), mg.MustList(), "&Int32*" )
//    t.addSucc( nil, mg.NullVal, "Int32*?" )
//    t.addNullValueError( nil, "Int32*" )
//    t.addNullValueError( nil, "Int32+" )
//    t.addNullValueError( mg.NewHeapValue( mg.NullVal ), "Int32+" )
//    t.addVcError( mg.NewHeapValue( mg.MustList() ), "&Int32+", "empty list" )
//    t.addSucc( 
//        nil, 
//        mg.NullVal,
//        MustNullableTypeReference( MustTypeReference( "&Int32*" ) ),
//    )
//    t.addSucc( 
//        mg.NewHeapValue( mg.NullVal ), 
//        mg.NullVal,
//        MustNullableTypeReference( MustTypeReference( "&Int32*" ) ),
//    )
//}
//
//func ( t *crtInit ) addMapTests() {
//    m1 := mg.MustSymbolMap
//    m2 := func() *mg.SymbolMap { return mg.MustSymbolMap( "f1", int32( 1 ) ) }
//    t.addSucc( m1(), m1(), mg.TypeSymbolMap )
//    t.addSucc( m1(), m1(), mg.TypeValue )
//    t.addSucc( m2(), m2(), mg.TypeSymbolMap )
//    t.addSucc( m2(), m2(), "SymbolMap?" )
//    s2 := &mg.Struct{ Type: qname( "ns2@v1/S1" ), Fields: m2() }
//    t.addSucc( s2, m2(), mg.TypeSymbolMap )
//    l1 := mg.MustList()
//    l2 := mg.MustList( m1(), m2() )
//    lt1 := MustTypeReference( "SymbolMap*" )
//    lt2 := MustTypeReference( "SymbolMap+" )
//    t.addSucc( l1, l1, lt1 )
//    t.addSucc( l2, l2, lt2 )
//    t.addSucc(
//        mg.MustSymbolMap( "f1", mg.NullVal ), 
//        mg.MustSymbolMap( "f1", mg.NullVal ), 
//        mg.TypeValue,
//    )
//    t.addSucc( mg.MustList( s2, s2 ), mg.MustList( m2(), m2() ), lt2 )
//    t.addTcError( int32( 1 ), mg.TypeSymbolMap, mg.TypeInt32 )
//    t.addTcError0(
//        mg.MustList( m1(), int32( 1 ) ),
//        mg.TypeSymbolMap,
//        mg.TypeInt32,
//        lt2,
//        crtPathDefault.StartList().SetIndex( 1 ),
//    )
//    nester := mg.MustSymbolMap( "f1", mg.MustSymbolMap( "f2", int32( 1 ) ) )
//    t.addSucc( nester, nester, mg.TypeSymbolMap )
//    t.addSucc( m1(), mg.NewHeapValue( m1() ), "&SymbolMap" )
//    t.addSucc( mg.NewHeapValue( m1() ), mg.NewHeapValue( m1() ), "&SymbolMap" )
//    t.addSucc( mg.NewHeapValue( m1() ), m1(), "SymbolMap" )
//    t.addSucc( nil, mg.NullVal, "SymbolMap?" )
//    t.addSucc( nil, mg.NullVal, "&SymbolMap?" )
//    t.addSucc( 
//        mg.NewHeapValue( mg.NullVal ), 
//        mg.NewHeapValue( mg.NullVal ),
//        NewPointerTypeReference( MustTypeReference( "SymbolMap?" ) ),
//    )
//    t.addNullValueError( nil, "SymbolMap" )
//    t.addNullValueError( nil, "&SymbolMap" )
//    t.addNullValueError( mg.NewHeapValue( mg.NullVal ), "&SymbolMap" )
//}
//
//func ( t *crtInit ) addStructTests() {
//    qn1 := qname( "ns1@v1/T1" )
//    t1 := qn1.AsAtomicType()
//    s1 := mg.MustStruct( qn1 )
//    s2 := mg.MustStruct( qn1, "f1", int32( 1 ) )
//    qn2 := qname( "ns1@v1/T2" )
//    t2 := qn2.AsAtomicType()
//    s3 := mg.MustStruct( qn2,
//        "f1", int32( 1 ),
//        "f2", s1,
//        "f3", s2,
//        "f4", mg.MustList( s1, s2 ),
//    )
//    t.addSucc( s1, s1, mg.TypeValue )
//    t.addSucc( s1, s1, t1 )
//    t.addSucc( s2, s2, t1 )
//    t.addSucc( s1, mg.NewHeapValue( s1 ), "&ns1@v1/T1?" )
//    t.addSucc( s3, s3, t2 )
//    l1 := mg.MustList( s1, s2 )
//    t.addSucc( l1, l1, &mg.ListTypeReference{ t1, false } )
//    t.addSucc( l1, l1, &mg.ListTypeReference{ t1, true } )
//    s4 := mg.MustStruct( "ns1@v1/T4", "f1", mg.NullVal )
//    t.addSucc( s4, s4, s4.Type )
//    f1 := func( in interface{}, inTyp TypeReferenceInitializer ) {
//        t.addTcError0(
//            mg.MustList( s1, in ),
//            t1,
//            inTyp,
//            &mg.ListTypeReference{ t1, false },
//            crtPathDefault.StartList().SetIndex( 1 ),
//        )
//    }
//    f1( s3, t2 )
//    f1( int32( 1 ), "Int32" )
//    t.addSucc( mg.NewHeapValue( s1 ), mg.NewHeapValue( s1 ), "&ns1@v1/T1" )
//    t.addSucc( s1, mg.NewHeapValue( s1 ), "&ns1@v1/T1" )
//    t.addSucc( mg.NewHeapValue( s1 ), s1, "ns1@v1/T1" )
//    t.addSucc( nil, mg.NullVal, "&ns1@v1/T1?" )
//    t.addNullValueError( nil, "&ns1@v1/T1" )
//    t.addNullValueError( mg.NewHeapValue( mg.NullVal ), "&ns1@v1/T1" )
//    t.addTcError( s1, "ns1@v1/T2", "ns1@v1/T1" )
//    t.addTcError( mg.NewHeapValue( s1 ), "ns1@v1/T2", "ns1@v1/T1" )
//    t.addTcError( s1, "ns1@v1/T2", "ns1@v1/T1" )
//    t.addTcError( mg.NewHeapValue( s1 ), "ns1@v1/T2", "ns1@v1/T1" )
//}
//
//func ( t *crtInit ) addInterfaceImplBasicTests() {
//    add := func( crt *CastReactorTest ) {
//        crt.Profile = "interface-impl-basic"
//        t.addCrtDefault( crt )
//    }
//    addSucc := func( in, expct interface{}, typ TypeReferenceInitializer ) {
//        add( t.createSucc( in, expct, typ ) )
//    }
//    t1 := qname( "ns1@v1/T1" )
//    t2 := qname( "ns1@v1/T2" )
//    s1 := mg.MustStruct( t1, "f1", int32( 1 ) )
//    addSucc( mg.MustStruct( t1, "f1", "1" ), s1, t1 )
//    addSucc( mg.MustSymbolMap( "f1", "1" ), s1, t1 )
//    addSucc( "cast1", int32( 1 ), "ns1@v1/S3" )
//    addSucc( "cast2", int32( -1 ), "ns1@v1/S3" )
//    s1Sub1 := mg.MustStruct( "ns1@v1/T1Sub1" )
//    addSucc( s1Sub1, s1Sub1, "ns1@v1/T1" )
//    addSucc( 
//        mg.MustList( "cast1", "cast2" ), 
//        mg.MustList( int32( 1 ), int32( -1 ) ),
//        "ns1@v1/S3*",
//    )
//    addSucc( nil, nil, "&ns1@v1/S3?" )
//    arb := mg.MustStruct( "ns1@v1/Arbitrary", "f1", int32( 1 ) )
//    addSucc( arb, arb, arb.Type )
//    add( t.createTcError( int32( 1 ), "ns1@v1/S3", mg.TypeInt32 ) )
//    add( t.createTcError( arb, "ns1@v1/S1", arb.Type ) )
//    add( 
//        t.createTcError0( 
//            int32( 1 ), 
//            "ns1@v1/S3", 
//            mg.TypeInt32, 
//            "&ns1@v1/S3?", 
//            crtPathDefault,
//        ),
//    )
//    add( 
//        t.createTcError0( 
//            mg.MustList( int32( 1 ) ),
//            "ns1@v1/S3",
//            mg.TypeInt32,
//            "ns1@v1/S3*",
//            crtPathDefault.StartList(),
//        ),
//    )
//    add( t.createVcError( "cast3", "ns1@v1/S3", "test-message-cast3" ) )
//    add( t.createVcError( "cast3", "&ns1@v1/S3?", "test-message-cast3" ) )
//    add(
//        t.createVcError0( 
//            mg.MustList( "cast2", "cast3" ),
//            "ns1@v1/S3+",
//            crtPathDefault.StartList().SetIndex( 1 ),
//            "test-message-cast3",
//        ),
//    )
//    s2InFlds := mg.MustSymbolMap( "f1", "1", "f2", mg.MustSymbolMap( "f1", "1" ) )
//    s2 := mg.MustStruct( t2, "f1", int32( 1 ), "f2", s1 )
//    addSucc( &mg.Struct{ Type: t2, Fields: s2InFlds }, s2, t2 )
//    addSucc( s2InFlds, s2, t2 )
//    add( t.createTcError( mg.MustStruct( t2, "f1", int32( 1 ) ), t1, t2 ) )
//    add( 
//        t.createTcError0(
//            mg.MustStruct( t1, "f1", mg.MustList( 1, 2 ) ),
//            mg.TypeInt32,
//            mg.TypeOpaqueList,
//            t1,
//            crtPathDefault.Descend( id( "f1" ) ),
//        ),
//    )
//    extraFlds1 := mg.MustSymbolMap( "f1", int32( 1 ), "x1", int32( 0 ) )
//    failExtra1 := func( val interface{} ) {
//        msg := "unrecognized field: x1"
//        add( t.createVcError0( val, t1, crtPathDefault, msg ) )
//    }
//    failExtra1( &mg.Struct{ Type: t1, Fields: extraFlds1 } )
//    failExtra1( extraFlds1 )
//    failTyp := qname( "ns1@v1/FailType" )
//    add(
//        t.createVcError0(
//            mg.MustStruct( failTyp ), 
//            failTyp, 
//            crtPathDefault,
//            "test-message-fail-type",
//        ),
//    )
//}
//
//// target definition would be:
////
////  struct S1 {
////      f0 mg.Int32
////      f1 &mg.Int32
////      f2 &mg.Int32
////      f3 &mg.Int64
////      f4 mg.Int64
////      f5 Int64*
////      f6 Int32*
////      f7 String*
////      f8 &Int32*
////      f9 S2
////      f10 S2
////      f11 &S2
////      f12 &mg.Int64
////      f13 &mg.Int64
////      f14 mg.Int64
////  }
////
//func ( t *crtInit ) addInterfacePointerHandlingTests() {
//    qn := func( i int ) *mg.QualifiedTypeName {
//        return mg.MustQualifiedTypeName( fmt.Sprintf( "ns1@v1/S%d", i ) )
//    }
//    typ := func( i int ) *mg.AtomicTypeReference {
//        return &mg.AtomicTypeReference{ Name: qn( i ) }
//    }
//    fldEv := func( i int ) *FieldStartEvent {
//        return NewFieldStartEvent( mg.MakeTestId( i ) )
//    }
//    list := func( typ mg.TypeReference, vals ...interface{} ) *mg.List {
//        res := mg.MustList( vals... )
//        res.Type = &mg.ListTypeReference{ ElementType: typ }
//        return res
//    }
//    hv12 := mg.NewHeapValue( mg.Int64( int64( 12 ) ) )
//    AddStdReactorTests(
//        &CastReactorTest{
//            In: mg.CopySource(
//                []ReactorEvent{
//                    NewStructStartEvent( qn( 1 ) ),
//                        fldEv( 0 ),
//                            ptrAlloc( mg.TypeInt64, 1 ), 
//                                NewValueEvent( mg.Int64( int64( 1 ) ) ),
//                        fldEv( 1 ), ptrRef( 1 ),
//                        fldEv( 2 ),
//                            ptrAlloc( mg.TypeInt64, 2 ), 
//                                NewValueEvent( mg.Int64( int64( 2 ) ) ),
//                        fldEv( 3 ), ptrRef( 2 ),
//                        fldEv( 4 ), ptrRef( 2 ),
//                        fldEv( 5 ), 
//                            NewListStartEvent( 
//                                &mg.ListTypeReference{ ElementType: mg.TypeInt32 },
//                                ptrId( 3 ),
//                            ),
//                                NewValueEvent( mg.Int32( int32( 0 ) ) ),
//                                NewValueEvent( mg.Int32( int32( 1 ) ) ),
//                                NewEndEvent(),
//                        fldEv( 6 ), ptrRef( 3 ),
//                        fldEv( 7 ), ptrRef( 3 ),
//                        fldEv( 8 ), ptrRef( 3 ),
//                        fldEv( 9 ), 
//                            ptrAlloc( typ( 2 ), 4 ),
//                                NewStructStartEvent( qn( 2 ) ), NewEndEvent(),
//                        fldEv( 10 ), ptrRef( 4 ),
//                        fldEv( 11 ), ptrRef( 4 ),
//                        fldEv( 12 ), 
//                            ptrAlloc( mg.TypeInt64, 5 ),
//                                NewValueEvent( mg.Int64( int64( 12 ) ) ),
//                        fldEv( 13 ), ptrRef( 5 ),
//                        fldEv( 14 ), ptrRef( 5 ),
//                    NewEndEvent(),
//                },
//            ),
//            Expect: mg.MustStruct( qn( 1 ),
//                "f0", mg.Int32( int32( 1 ) ),
//                "f1", mg.NewHeapValue( mg.Int32( int32( 1 ) ) ),
//                "f2", mg.NewHeapValue( mg.Int32( int32( 2 ) ) ),
//                "f3", mg.NewHeapValue( mg.Int64( int64( 2 ) ) ),
//                "f4", mg.Int64( int64( 2 ) ),
//                "f5", list( mg.TypeInt64, int64( 0 ), int64( 1 ) ),
//                "f6", list( mg.TypeInt32, int32( 0 ), int32( 1 ) ),
//                "f7", list( mg.TypeString, "0", "1" ),
//                "f8", list( NewPointerTypeReference( mg.TypeInt32 ),
//                    mg.NewHeapValue( mg.Int32( int32( 0 ) ) ),
//                    mg.NewHeapValue( mg.Int32( int32( 1 ) ) ),
//                ),
//                "f9", mg.MustStruct( qn( 2 ) ),
//                "f10", mg.MustStruct( qn( 2 ) ),
//                "f11", mg.NewHeapValue( mg.MustStruct( qn( 2 ) ) ),
//                "f12", hv12,
//                "f13", hv12,
//                "f14", mg.Int64( int64( 12 ) ),
//            ),
//            Type: typ( 1 ),
//            Profile: "interface-pointer-handling",
//        },
//    )
//}
//
//func ( t *crtInit ) addInterfaceImplTests() {
//    t.addInterfaceImplBasicTests()
//    t.addInterfacePointerHandlingTests()
//}
//
//func ( t *crtInit ) call() {
//    t.initStdVals()
//    t.addBaseTypeTests()
//    t.addMiscTcErrors()
//    t.addMiscVcErrors()
//    t.addNumTests()
//    t.addStringTests()
//    t.addBufferTests()
//    t.addTimeTests()
//    t.addEnumTests()
//    t.addNullableTests()
//    t.addListTests()
//    t.addMapTests()
//    t.addStructTests()
//    t.addInterfaceImplTests()
//}
//
//func initCastReactorTests() { ( &crtInit{} ).call() }

func initReactorTests( b *ReactorTestSetBuilder ) {
//    StdReactorTests = []interface{}{}
    initValueBuildReactorTests( b )
//    initStructuralReactorTests()
//    initPointerReferenceCheckTests()
//    initEventPathTests()
//    initFieldOrderReactorTests()
//    initServiceTests()
//    initCastReactorTests()
}

func init() { AddTestInitializer( initReactorTests ) }

//
//type CastErrorAssert struct {
//    ErrExpect, ErrAct error
//    *assert.PathAsserter
//}
//
//func ( cea CastErrorAssert ) FailActErrType() {
//    cea.Fatalf(
//        "Expected error of type %T but got %T: %s",
//        cea.ErrExpect, cea.ErrAct, cea.ErrAct )
//}
//
//// Returns a path asserter that can be used further
//func ( cea CastErrorAssert ) assertValueError( 
//    expct, act ValueError ) *assert.PathAsserter {
//    a := cea.Descend( "Err" )
//    a.Descend( "Error()" ).Equal( expct.Error(), act.Error() )
//    a.Descend( "Message()" ).Equal( expct.Message(), act.Message() )
//    a.Descend( "Location()" ).Equal( expct.Location(), act.Location() )
//    return a
//}
//
//func ( cea CastErrorAssert ) assertVcError() {
//    if act, ok := cea.ErrAct.( *ValueCastError ); ok {
//        cea.assertValueError( cea.ErrExpect.( *ValueCastError ), act )
//    } else { cea.FailActErrType() }
//}
//
//func ( cea CastErrorAssert ) assertMissingFieldsError() {
//    if act, ok := cea.ErrAct.( *MissingFieldsError ); ok {
//        cea.assertValueError( cea.ErrExpect.( ValueError ), act )
//    } else { cea.FailActErrType() }
//}
//
//func ( cea CastErrorAssert ) assertUnrecognizedFieldError() {
//    if act, ok := cea.ErrAct.( *UnrecognizedFieldError ); ok {
//        cea.assertValueError( cea.ErrExpect.( ValueError ), act )
//    } else { cea.FailActErrType() }
//}
//
//func ( cea CastErrorAssert ) Call() {
//    switch cea.ErrExpect.( type ) {
//    case nil: cea.Fatal( cea.ErrAct )
//    case *ValueCastError: cea.assertVcError()
//    case *MissingFieldsError: cea.assertMissingFieldsError()
//    case *UnrecognizedFieldError: cea.assertUnrecognizedFieldError()
//    default: cea.Fatalf( "Unhandled Err type: %T", cea.ErrExpect )
//    }
//}
//
//func AssertCastError( expct, act error, pa *assert.PathAsserter ) {
//    ca := CastErrorAssert{ ErrExpect: expct, ErrAct: act, PathAsserter: pa }
//    ca.Call()
//}
//
//func eventForEqualityCheck( 
//    ev ReactorEvent, ignorePointerIds bool ) ReactorEvent {
//
//    ev = CopyEvent( ev, true )
//    switch v := ev.( type ) {
//    case *ValueAllocationEvent: v.Id = mg.PointerIdNull
//    case *ValueReferenceEvent: v.Id = mg.PointerIdNull
//    case *MapStartEvent: v.Id = mg.PointerIdNull
//    case *ListStartEvent: v.Id = mg.PointerIdNull
//    }
//    return ev
//}
//
//func EqualEvents( 
//    expct, act ReactorEvent, ignorePointerIds bool, a *assert.PathAsserter ) {
//
//    expct = eventForEqualityCheck( expct, ignorePointerIds )
//    act = eventForEqualityCheck( act, ignorePointerIds )
//    a.Equalf( expct, act, "events are not equal: %s != %s",
//        EventToString( expct ), EventToString( act ) )
//}
//
//type reactorEventSource interface {
//    Len() int
//    EventAt( int ) ReactorEvent
//}
//
//func FeedEventSource( 
//    src reactorEventSource, proc ReactorEventProcessor ) error {
//
//    for i, e := 0, src.Len(); i < e; i++ {
//        if err := proc.ProcessEvent( src.EventAt( i ) ); err != nil { 
//            return err
//        }
//    }
//    return nil
//}
//
//func AssertFeedEventSource(
//    src reactorEventSource, proc ReactorEventProcessor, a assert.Failer ) {
//    
//    if err := FeedEventSource( src, proc ); err != nil { a.Fatal( err ) }
//}
//
//type eventSliceSource []ReactorEvent
//func ( src eventSliceSource ) Len() int { return len( src ) }
//func ( src eventSliceSource ) EventAt( i int ) ReactorEvent { return src[ i ] }
//
//type eventExpectSource []EventExpectation
//
//func ( src eventExpectSource ) Len() int { return len( src ) }
//
//func ( src eventExpectSource ) EventAt( i int ) ReactorEvent {
//    return CopyEvent( src[ i ].Event, true )
//}
//
//func FeedSource( src interface{}, rct ReactorEventProcessor ) error {
//    switch v := src.( type ) {
//    case reactorEventSource: return FeedEventSource( v, rct )
//    case []ReactorEvent: return FeedSource( eventSliceSource( v ), rct )
//    case mg.Value: return VisitValue( v, rct )
//    }
//    panic( libErrorf( "unhandled source: %T", src ) )
//}
//
//func AssertFeedSource( 
//    src interface{}, rct ReactorEventProcessor, a assert.Failer ) {
//
//    if err := FeedSource( src, rct ); err != nil { a.Fatal( err ) }
//}
//
//type eventPathCheckReactor struct {
//    a *assert.PathAsserter
//    eeAssert *assert.PathAsserter
//    expct []EventExpectation
//    idx int
//    ignorePointerIds bool
//}
//
//func ( r *eventPathCheckReactor ) ProcessEvent( ev ReactorEvent ) error {
//    r.a.Truef( r.idx < len( r.expct ), "unexpected event: %v", ev )
//    ee := r.expct[ r.idx ]
//    r.idx++
//    ee.Event.SetPath( ee.Path )
//    EqualEvents( ee.Event, ev, r.ignorePointerIds, r.eeAssert )
//    r.eeAssert = r.eeAssert.Next()
//    return nil
//}
//
//func ( r *eventPathCheckReactor ) Complete() {
//    r.a.Equalf( r.idx, len( r.expct ), "not all events were seen" )
//}
//
//func NewEventPathCheckReactor( 
//    expct []EventExpectation, a *assert.PathAsserter ) *eventPathCheckReactor {
//
//    return &eventPathCheckReactor{ 
//        expct: expct, 
//        a: a,
//        eeAssert: a.Descend( "expct" ).StartList(),
//    }
//}
//
//type ReactorTestCall struct {
//    *assert.PathAsserter
//    Test interface{}
//}
//
//func ( c *ReactorTestCall ) CheckNoError( err error ) {
//    if err != nil { c.Fatalf( "Got no error but expected %T: %s", err, err ) }
//}
//
//func ( c *ReactorTestCall ) EqualErrors( expct, act error ) {
//    if expct == nil { c.Fatal( act ) }
//    c.Equalf( expct, act, "expected %q (%T) but got %q (%T)",
//        expct, expct, act, act )
//}
//
//func CheckBuiltValue( expct mg.Value, vb *ValueBuilder, a *assert.PathAsserter ) {
//    if expct == nil {
//        if vb != nil {
//            a.Fatalf( "unexpected value: %s", QuoteValue( vb.GetValue() ) )
//        }
//    } else { 
//        a.Falsef( vb == nil, 
//            "expecting value %s but value builder is nil", QuoteValue( expct ) )
//        EqualWireValues( expct, vb.GetValue(), a ) 
//    }
//}
