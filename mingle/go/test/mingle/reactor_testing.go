package mingle

import (
    "bitgirder/objpath"
    "bitgirder/assert"
    "fmt"
    "encoding/base64"
    "bytes"
    "strconv"
//    "log"
)

var StdReactorTests []interface{}

func AddStdReactorTests( t ...interface{} ) {
    StdReactorTests = append( StdReactorTests, t... )
}

type NamedTest interface { TestName() string }

func makeTestId( i int ) *Identifier {
    return MustIdentifier( fmt.Sprintf( "f%d", i ) )
}

func mustInt( s string ) int {
    res, err := strconv.Atoi( s )
    if ( err != nil ) { panic( err ) }
    return res
}

func flattenEvs( vals ...interface{} ) []ReactorEvent {
    res := make( []ReactorEvent, 0, len( vals ) )
    for _, val := range vals {
        switch v := val.( type ) {
        case ReactorEvent: res = append( res, v )
        case []ReactorEvent: res = append( res, v... )
        case []interface{}: res = append( res, flattenEvs( v... )... )
        default: panic( libErrorf( "Uhandled ev type for flatten: %T", v ) )
        }
    }
    return res
}

// to simplify test creation, we reuse event instances when constructing input
// event sequences, and send them to this method only at the end to ensure that
// we get a distinct sequence of event values for each test
func copySource( evs []ReactorEvent ) []ReactorEvent {
    res := make( []ReactorEvent, len( evs ) )
    for i, ev := range evs { res[ i ] = CopyEvent( ev, false ) }
    return res
}

func startTestIdPath( elt interface{} ) objpath.PathNode {
    switch v := elt.( type ) {
    case int: return objpath.RootedAt( makeTestId( v ) )
    case string: return objpath.RootedAtList().SetIndex( mustInt( v ) )
    }
    panic( libErrorf( "unhandled elt: %T", elt ) )
}

func makeTestIdPath( elts ...interface{} ) objpath.PathNode { 
    if len( elts ) == 0 { return nil }
    res := startTestIdPath( elts[ 0 ] )
    for i, e := 1, len( elts ); i < e; i++ {
        switch v := elts[ i ].( type ) {
        case int: res = res.Descend( makeTestId( v ) ) 
        case string: res = res.StartList().SetIndex( mustInt( v ) )
        default: panic( libErrorf( "unhandled elt: %T", v ) )
        }
    }
    return res
}

type ValueBuildTest struct { Val Value }

func initValueBuildReactorTests() {
    s1 := MustStruct( "ns1@v1/S1",
        "val1", String( "hello" ),
        "list1", MustList(),
        "map1", MustSymbolMap(),
        "struct1", MustStruct( "ns1@v1/S2" ),
    )
    mk := func( v Value ) interface{} { return ValueBuildTest{ v } }
    StdReactorTests = append( StdReactorTests,
        mk( String( "hello" ) ),
        mk( MustList() ),
        mk( MustList( 1, 2, 3 ) ),
        mk( MustList( 1, MustList(), MustList( 1, 2 ) ) ),
        mk( MustSymbolMap() ),
        mk( MustSymbolMap( "f1", "v1", "f2", MustList(), "f3", s1 ) ),
        mk( s1 ),
        mk( MustStruct( "ns1@v1/S3" ) ),
    )
}

type StructuralReactorErrorTest struct {
    Events []ReactorEvent
    Error *ReactorError
    TopType ReactorTopType
}

type EventExpectation struct {
    Event ReactorEvent
    Path objpath.PathNode
}

type EventPathTest struct {
    Name string
    Events []EventExpectation
    StartPath objpath.PathNode
}

func ( ept EventPathTest ) TestName() string { return ept.Name }

func initStructuralReactorTests() {
    evStartStruct1 := NewStructStartEvent( qname( "ns1@v1/S1" ) )
    id := makeTestId
    evStartField1 := NewFieldStartEvent( id( 1 ) )
    evStartField2 := NewFieldStartEvent( id( 2 ) )
    evValue1 := NewValueEvent( Int64( int64( 1 ) ) )
    mk1 := func( 
        errMsg string, evs ...ReactorEvent ) *StructuralReactorErrorTest {
        return &StructuralReactorErrorTest{
            Events: evs,
            Error: rctError( errMsg ),
        }
    }
    mk2 := func( 
        errMsg string, 
        tt ReactorTopType, 
        evs ...ReactorEvent ) *StructuralReactorErrorTest {
        res := mk1( errMsg, evs... )
        res.TopType = tt
        return res
    }
    AddStdReactorTests(
        mk1( "Saw start of field 'f2' while expecting a value for field 'f1'",
            evStartStruct1, evStartField1, evStartField2,
        ),
        mk1( "Saw start of field 'f2' while expecting a value for field 'f1'",
            evStartStruct1, evStartField1, NewMapStartEvent(), evStartField1,
            evStartField2,
        ),
        mk1( "Saw start of field 'f1' after value was built",
            evStartStruct1, NewEndEvent(), evStartField1,
        ),
        mk1( "Expected field name or end of fields but got value",
            evStartStruct1, evValue1,
        ),
        mk1( "Expected field name or end of fields but got list start",
            evStartStruct1, NewListStartEvent(),
        ),
        mk1( "Expected field name or end of fields but got map start",
            evStartStruct1, NewMapStartEvent(),
        ),
        mk1( "Expected field name or end of fields but got start of struct " +
                evStartStruct1.Type.ExternalForm(),
            evStartStruct1, evStartStruct1,
        ),
        mk1( "Saw end while expecting a value for field 'f1'",
            evStartStruct1, evStartField1, NewEndEvent(),
        ),
        mk1( "Saw start of field 'f1' while expecting a list value",
            evStartStruct1, evStartField1, NewListStartEvent(), evStartField1,
        ),
        mk2( "Expected struct but got value", ReactorTopTypeStruct, evValue1 ),
        mk2( "Expected struct but got list start", ReactorTopTypeStruct,
            NewListStartEvent(),
        ),
        mk2( "Expected struct but got map start", ReactorTopTypeStruct,
            NewMapStartEvent(),
        ),
        mk2( "Expected struct but got start of field 'f1'", 
            ReactorTopTypeStruct, evStartField1,
        ),
        mk2( "Expected struct but got end", 
            ReactorTopTypeStruct, NewEndEvent() ),
        mk1( "Multiple entries for field: f1",
            evStartStruct1, 
            evStartField1, evValue1,
            evStartField2, evValue1,
            evStartField1,
        ),
    )
}

func initEventPathTests() {
    evStartStruct1 := NewStructStartEvent( qname( "ns1@v1/S1" ) )
    id := makeTestId
    evStartField1 := NewFieldStartEvent( id( 1 ) )
    evStartField2 := NewFieldStartEvent( id( 2 ) )
    evValue1 := NewValueEvent( Int64( int64( 1 ) ) )
    evEnd := NewEndEvent()
    mk := func( name string, evs ...EventExpectation ) *EventPathTest {
        return &EventPathTest{ Name: name, Events: evs }
    }
    p := makeTestIdPath
    ee := func( ev ReactorEvent, p objpath.PathNode ) EventExpectation {
        return EventExpectation{ Event: ev, Path: p }
    }
    AddStdReactorTests(
        mk( "empty" ),
        mk( "top-value", ee( evValue1, nil ) ),
        mk( "empty-struct", 
            ee( evStartStruct1, nil ), 
            ee( evStartField1, p( 1 ) ),
            ee( evValue1, p( 1 ) ),
            ee( evEnd, nil ),
        ),
        mk( "empty-map",
            ee( NewMapStartEvent(), nil ),
            ee( evStartField1, p( 1 ) ),
            ee( evValue1, p( 1 ) ),
            ee( evEnd, nil ),
        ),
        mk( "flat-struct",
            ee( evStartStruct1, nil ),
            ee( evStartField1, p( 1 ) ),
            ee( evValue1, p( 1 ) ),
            ee( evEnd, nil ),
        ),
        mk( "empty-list",
            ee( NewListStartEvent(), nil ),
            ee( NewEndEvent(), nil ),
        ),
        mk( "flat-list",
            ee( NewListStartEvent(), nil ),
            ee( evValue1, p( "0" ) ),
            ee( evValue1, p( "1" ) ),
            ee( NewEndEvent(), nil ),
        ),
        mk( "nested-list1",
            ee( NewListStartEvent(), nil ),
            ee( NewMapStartEvent(), p( "0" ) ),
            ee( NewEndEvent(), p( "0" ) ),
            ee( NewEndEvent(), nil ),
        ),
        mk( "nested-list2",
            ee( NewListStartEvent(), nil ),
            ee( NewMapStartEvent(), p( "0" ) ),
            ee( NewEndEvent(), p( "0" ) ),
            ee( evValue1, p( "1" ) ),
            ee( NewEndEvent(), nil ),
        ),
        mk( "nested-list3",
            ee( NewListStartEvent(), nil ),
            ee( evValue1, p( "0" ) ),
            ee( NewMapStartEvent(), p( "1" ) ),
            ee( evStartField1, p( "1", 1 ) ),
            ee( evValue1, p( "1", 1 ) ),
            ee( NewEndEvent(), p( "1" ) ),
            ee( NewEndEvent(), nil ),
        ),
        mk( "flat-map",
            ee( NewMapStartEvent(), nil ),
            ee( evStartField1, p( 1 ) ),
            ee( evValue1, p( 1 ) ),
            ee( NewEndEvent(), nil ),
        ),
        mk( "struct-with-containers1",
            ee( evStartStruct1, nil ),
            ee( evStartField1, p( 1 ) ),
            ee( NewListStartEvent(), p( 1 ) ),
            ee( evValue1, p( 1, "0" ) ),
            ee( evValue1, p( 1, "1" ) ),
            ee( NewEndEvent(), p( 1 ) ),
            ee( NewEndEvent(), nil ),
        ),
        mk( "struct-with-containers2",
            ee( evStartStruct1, nil ),
            ee( evStartField1, p( 1 ) ),
                ee( NewMapStartEvent(), p( 1 ) ),
                ee( evStartField2, p( 1, 2 ) ),
                    ee( NewListStartEvent(), p( 1, 2 ) ),
                        ee( evValue1, p( 1, 2, "0" ) ),
                        ee( evValue1, p( 1, 2, "1" ) ),
                        ee( NewListStartEvent(), p( 1, 2, "2" ) ),
                            ee( evValue1, p( 1, 2, "2", "0" ) ),
                            ee( NewMapStartEvent(), p( 1, 2, "2", "1" ) ),
                            ee( evStartField1, p( 1, 2, "2", "1", 1 ) ),
                                ee( evValue1, p( 1, 2, "2", "1", 1 ) ),
                            ee( NewEndEvent(), p( 1, 2, "2", "1" ) ),
                        ee( NewEndEvent(), p( 1, 2, "2" ) ),
                    ee( NewEndEvent(), p( 1, 2 ) ),
                ee( NewEndEvent(), p( 1 ) ),
            ee( NewEndEvent(), nil ),
        ),
    )
    AddStdReactorTests(
        &EventPathTest{
            Name: "non-empty-start-path",
            Events: []EventExpectation{ 
                { NewMapStartEvent(), p( 2, "4" ) },
                { evStartField1, p( 2, "4", 1 ) },
            },
            StartPath: p( 2, "3" ),
        },
    )
}

type FieldOrderReactorTestOrder struct {
    Order FieldOrder
    Type *QualifiedTypeName
}

func testOrderWithIds( 
    typ *QualifiedTypeName, ids ...*Identifier ) FieldOrderReactorTestOrder {
    ord := make( []FieldOrderSpecification, len( ids ) )
    for i, id := range ids { 
        ord[ i ] = FieldOrderSpecification{ Field: id, Required: false }
    }
    return FieldOrderReactorTestOrder{ Type: typ, Order: ord }
}

type FieldOrderReactorTest struct {
    Source []ReactorEvent
    Expect Value
    Orders []FieldOrderReactorTestOrder
}

func initFieldOrderValueTests() {
    flds := make( []ReactorEvent, 5 )
    ids := make( []*Identifier, len( flds ) )
    for i := 0; i < len( flds ); i++ {
        ids[ i ] = id( fmt.Sprintf( "f%d", i ) )
        flds[ i ] = NewFieldStartEvent( ids[ i ] )
    }
    i1 := Int32( int32( 1 ) )
    val1 := NewValueEvent( i1 )
    t1, t2 := qname( "ns1@v1/S1" ), qname( "ns1@v1/S2" )
    ss1, ss2 := NewStructStartEvent( t1 ), NewStructStartEvent( t2 )
    ss2Val1 := MustStruct( t2, ids[ 0 ], i1 )
    // expct sequences for instance of ns1@v1/S1 by field f0 ...
    fldVals := []Value{
        i1,
        MustSymbolMap( ids[ 0 ], i1, ids[ 1 ], ss2Val1 ),
        MustList( i1, i1 ),
        ss2Val1,
        i1,
    }
    mkExpct := func( ord ...int ) *Struct {
        pairs := []interface{}{}
        for _, fldNum := range ord {
            pairs = append( pairs, ids[ fldNum ], fldVals[ fldNum ] )
        }
        return MustStruct( t1, pairs... )
    }
    // val sequences for fields f0 ...
    fldEvs := [][]ReactorEvent {
        []ReactorEvent{ val1 },
        []ReactorEvent{
            NewMapStartEvent(), 
                flds[ 0 ], val1, 
                flds[ 1 ], ss2, flds[ 0 ], val1, NewEndEvent(),
            NewEndEvent(),
        },
        []ReactorEvent{ NewListStartEvent(), val1, val1, NewEndEvent() },
        []ReactorEvent{ ss2, flds[ 0 ], val1, NewEndEvent() },
        []ReactorEvent{ val1 },
    }
    mkSrc := func( ord ...int ) []ReactorEvent {
        res := []ReactorEvent{ ss1 }
        for _, fldNum := range ord {
            res = append( res, flds[ fldNum ] )
            res = append( res, fldEvs[ fldNum ]... )
        }
        return append( res, NewEndEvent() )
    }
    addTest1 := func( src []ReactorEvent, expct Value ) {
        AddStdReactorTests(
            &FieldOrderReactorTest{ 
                Source: copySource( src ), 
                Expect: expct, 
                Orders: []FieldOrderReactorTestOrder{
                    testOrderWithIds( t1,
                        ids[ 0 ], ids[ 1 ], ids[ 2 ], ids[ 3 ] ),
                },
            },
        )
    }
    for _, ord := range [][]int {
        []int{ 0, 1, 2, 3 }, // first one should be straight passthrough
        []int{ 3, 2, 1, 0 },
        []int{ 0, 3, 2, 1 }, 
        []int{ 0, 2, 3, 1 },
        []int{ 0, 1 },
        []int{ 0, 2 },
        []int{ 0, 3 },
        []int{ 2, 0 },
        []int{ 2, 1 },
        []int{ 4, 3, 0, 1, 2 },
        []int{ 4, 3, 2, 1, 0 },
        []int{ 1, 4, 3, 0, 2 },
        []int{ 1, 4, 3, 2, 0 },
        []int{ 0, 4, 3, 2, 1 },
        []int{ 0, 4, 3, 1, 2 },
        []int{ 0, 1, 3, 2, 4 },
    } {
        addTest1( mkSrc( ord... ), mkExpct( ord... ) )
    }
    // Test nested orderings
    AddStdReactorTests(
        &FieldOrderReactorTest{
            Source: copySource( 
                []ReactorEvent{
                    ss1, 
                        flds[ 0 ], val1,
                        flds[ 1 ], ss1,
                            flds[ 2 ], NewListStartEvent(), val1, NewEndEvent(),
                            flds[ 1 ], val1,
                        NewEndEvent(),
                    NewEndEvent(),
                },
            ),
            Orders: []FieldOrderReactorTestOrder{
                testOrderWithIds( t1, ids[ 1 ], ids[ 0 ], ids[ 2 ] ),
            },
            Expect: MustStruct( t1,
                ids[ 0 ], i1,
                ids[ 1 ], MustStruct( t1,
                    ids[ 2 ], MustList( i1 ),
                    ids[ 1 ], i1,
                ),
            ),
        },
    )
    // Test generic un-field-ordered values at the top-level as well
    for i := 0; i < 4; i++ { addTest1( fldEvs[ i ], fldVals[ i ] ) }
    // Test arbitrary values with no orders in play
    addTest2 := func( expct Value, src ...ReactorEvent ) {
        AddStdReactorTests(
            &FieldOrderReactorTest{
                Source: copySource( src ),
                Expect: expct,
                Orders: []FieldOrderReactorTestOrder{},
            },
        )
    }
    addTest2( i1, val1 )
    addTest2( MustList(), NewListStartEvent(), NewEndEvent() )
    addTest2( MustList( i1 ), NewListStartEvent(), val1, NewEndEvent() )
    addTest2( MustSymbolMap(), NewMapStartEvent(), NewEndEvent() )
    addTest2( 
        MustSymbolMap( ids[ 0 ], i1 ), 
        NewMapStartEvent(), flds[ 0 ], val1, NewEndEvent(),
    )
    addTest2( MustStruct( ss1.Type ), ss1, NewEndEvent() )
    addTest2( 
        MustStruct( ss1.Type, ids[ 0 ], i1 ),
        ss1, flds[ 0 ], val1, NewEndEvent(),
    )
}

type FieldOrderMissingFieldsTest struct {
    Orders []FieldOrderReactorTestOrder
    Source []ReactorEvent
    Expect Value
    Error *MissingFieldsError
}

func initFieldOrderMissingFieldTests() {
    fldId := func( i int ) *Identifier { return id( fmt.Sprintf( "f%d", i ) ) }
    ord := FieldOrder( 
        []FieldOrderSpecification{
            { fldId( 0 ), true },
            { fldId( 1 ), true },
            { fldId( 2 ), false },
            { fldId( 3 ), false },
            { fldId( 4 ), true },
        },
    )
    t1 := qname( "ns1@v1/S1" )
    ords := []FieldOrderReactorTestOrder{ { Order: ord, Type: t1 } }
    mkSrc := func( flds []int ) []ReactorEvent {
        evs := []interface{}{ NewStructStartEvent( t1 ) }
        for _, fld := range flds {
            evs = append( evs, NewFieldStartEvent( fldId( fld ) ) )
            evs = append( evs, NewValueEvent( Int32( fld ) ) )
        }
        return flattenEvs( append( evs, NewEndEvent() ) )
    }
    mkVal := func( flds []int ) *Struct {
        pairs := make( []interface{}, 0, 2 * len( flds ) )
        for _, fld := range flds {
            pairs = append( pairs, fldId( fld ), Int32( fld ) )
        }
        return MustStruct( t1, pairs... )
    }
    addSucc := func( flds ...int ) {
        AddStdReactorTests(
            &FieldOrderMissingFieldsTest{
                Orders: ords,
                Source: copySource( mkSrc( flds ) ),
                Expect: mkVal( flds ),
            },
        )
    }
    addSucc( 0, 1, 4 )
    addSucc( 4, 0, 1 )
    addSucc( 0, 1, 3, 4 )
    addSucc( 0, 3, 1, 4 )
    addSucc( 0, 1, 4, 3, 2 )
    addErr := func( missIds []int, flds ...int ) {
        miss := make( []*Identifier, len( missIds ) )
        for i, missId := range missIds { miss[ i ] = fldId( missId ) }
        AddStdReactorTests(
            &FieldOrderMissingFieldsTest{
                Orders: ords,
                Source: copySource( mkSrc( flds ) ),
                Error: NewMissingFieldsError( nil, miss ),
            },
        )
    }
    addErr( []int{ 0 }, 1, 2, 3, 4 )
    addErr( []int{ 1 }, 0, 4, 3, 2 )
    addErr( []int{ 4 }, 3, 2, 1, 0 )
    addErr( []int{ 0, 1 }, 4 )
    addErr( []int{ 1, 4 }, 0 )
    addErr( []int{ 0, 4 }, 1 )
    addErr( []int{ 4 }, 1, 0 )
    addErr( []int{ 1 }, 4, 3, 0, 2 )
}

type FieldOrderPathTest struct {
    Source []ReactorEvent
    Expect []EventExpectation
    Orders []FieldOrderReactorTestOrder
}

func initFieldOrderPathTests() {
    i1 := Int32( int32( 1 ) )
    val1 := NewValueEvent( i1 )
    id := makeTestId
    typ := func( i int ) *QualifiedTypeName {
        return qname( fmt.Sprintf( "ns1@v1/S%d", i ) )
    }
    ss := func( i int ) *StructStartEvent { 
        return NewStructStartEvent( typ( i ) ) 
    }
    fld := func( i int ) *FieldStartEvent { 
        return NewFieldStartEvent( id( i ) ) 
    }
    p := makeTestIdPath
    expct1 := []EventExpectation{
        { ss( 1 ), nil },
            { fld( 0 ), p( 0 ) },
            { val1, p( 0 ) },
            { fld( 1 ), p( 1 ) },
            { NewMapStartEvent(), p( 1 ) },
                { fld( 1 ), p( 1, 1 ) },
                { val1, p( 1, 1 ) },
                { fld( 0 ), p( 1, 0 ) },
                { val1, p( 1, 0 ) },
            { NewEndEvent(), p( 1 ) },
            { fld( 2 ), p( 2 ) },
            { NewListStartEvent(), p( 2 ) },
                { val1, p( 2, "0" ) },
                { val1, p( 2, "1" ) },
            { NewEndEvent(), p( 2 ) },
            { fld( 3 ), p( 3 ) },
            { ss( 2 ), p( 3 ) },
                { fld( 0 ), p( 3, 0 ) },
                { val1, p( 3, 0 ) },
                { fld( 1 ), p( 3, 1 ) },
                { NewListStartEvent(), p( 3, 1 ) },
                    { val1, p( 3, 1, "0" ) },
                    { val1, p( 3, 1, "1" ) },
                { NewEndEvent(), p( 3, 1 ) },
            { NewEndEvent(), p( 3 ) },
            { fld( 4 ), p( 4 ) },
            { ss( 1 ), p( 4 ) },
                { fld( 0 ), p( 4, 0 ) },
                { val1, p( 4, 0 ) },
                { fld( 1 ), p( 4, 1 ) },
                { ss( 3 ), p( 4, 1 ) },
                    { fld( 0 ), p( 4, 1, 0 ) },
                    { val1, p( 4, 1, 0 ) },
                    { fld( 1 ), p( 4, 1, 1 ) },
                    { val1, p( 4, 1, 1 ) },
                { NewEndEvent(), p( 4, 1 ) },
                { fld( 2 ), p( 4, 2 ) },
                { ss( 3 ), p( 4, 2 ) },
                    { fld( 0 ), p( 4, 2, 0 ) },
                    { val1, p( 4, 2, 0 ) },
                    { fld( 1 ), p( 4, 2, 1 ) },
                    { val1, p( 4, 2, 1 ) },
                { NewEndEvent(), p( 4, 2 ) },
                { fld( 3 ), p( 4, 3 ) },
                { NewMapStartEvent(), p( 4, 3 ) },
                    { fld( 0 ), p( 4, 3, 0 ) },
                    { ss( 3 ), p( 4, 3, 0 ) },
                        { fld( 0 ), p( 4, 3, 0, 0 ) },
                        { val1, p( 4, 3, 0, 0 ) },
                        { fld( 1 ), p( 4, 3, 0, 1 ) },
                        { val1, p( 4, 3, 0, 1 ) },
                    { NewEndEvent(), p( 4, 3, 0 ) },
                    { fld( 1 ), p( 4, 3, 1 ) },
                    { ss( 3 ), p( 4, 3, 1 ) },
                        { fld( 0 ), p( 4, 3, 1, 0 ) },
                        { val1, p( 4, 3, 1, 0 ) },
                        { fld( 1 ), p( 4, 3, 1, 1 ) },
                        { val1, p( 4, 3, 1, 1 ) },
                    { NewEndEvent(), p( 4, 3, 1 ) },
                { NewEndEvent(), p( 4, 3 ) },
                { fld( 4 ), p( 4, 4 ) },
                { NewListStartEvent(), p( 4, 4 ) },
                    { ss( 3 ), p( 4, 4, "0" ) },
                        { fld( 0 ), p( 4, 4, "0", 0 ) },
                        { val1, p( 4, 4, "0", 0 ) },
                        { fld( 1 ), p( 4, 4, "0", 1 ) },
                        { val1, p( 4, 4, "0", 1 ) },
                    { NewEndEvent(), p( 4, 4, "0" ) },
                    { ss( 3 ), p( 4, 4, "1" ) },
                        { fld( 0 ), p( 4, 4, "1", 0 ) },
                        { val1, p( 4, 4, "1", 0 ) },
                        { fld( 1 ), p( 4, 4, "1", 1 ) },
                        { val1, p( 4, 4, "1", 1 ) },
                    { NewEndEvent(), p( 4, 4, "1" ) },
                { NewEndEvent(), p( 4, 4 ) },
            { NewEndEvent(), p( 4 ) },
        { NewEndEvent(), nil },
    }
    ords1 := []FieldOrderReactorTestOrder{
        testOrderWithIds( ss( 1 ).Type,
            id( 0 ), id( 1 ), id( 2 ), id( 3 ), id( 4 ) ),
        testOrderWithIds( ss( 2 ).Type, id( 0 ), id( 1 ) ),
        testOrderWithIds( ss( 3 ).Type, id( 0 ), id( 1 ) ),
    }
    evs := [][]ReactorEvent{
        []ReactorEvent{ val1 },
        []ReactorEvent{ 
            NewMapStartEvent(), fld( 1 ), val1, fld( 0 ), val1, NewEndEvent() },
        []ReactorEvent{ NewListStartEvent(), val1, val1, NewEndEvent() },
        []ReactorEvent{ 
            ss( 2 ), 
                fld( 0 ), val1, 
                fld( 1 ), NewListStartEvent(), val1, val1, NewEndEvent(),
            NewEndEvent(),
        },
        // val for f4 is nested and has nested ss2 instances that are in varying
        // need of reordering
        []ReactorEvent{ 
            ss( 1 ),
                fld( 0 ), val1,
                fld( 4 ), NewListStartEvent(),
                    ss( 3 ),
                        fld( 0 ), val1,
                        fld( 1 ), val1,
                    NewEndEvent(),
                    ss( 3 ),
                        fld( 1 ), val1,
                        fld( 0 ), val1,
                    NewEndEvent(),
                NewEndEvent(),
                fld( 2 ), ss( 3 ),
                    fld( 1 ), val1,
                    fld( 0 ), val1,
                NewEndEvent(),
                fld( 3 ), NewMapStartEvent(),
                    fld( 0 ), ss( 3 ),
                        fld( 1 ), val1,
                        fld( 0 ), val1,
                    NewEndEvent(),
                    fld( 1 ), ss( 3 ),
                        fld( 0 ), val1,
                        fld( 1 ), val1,
                    NewEndEvent(),
                NewEndEvent(),
                fld( 1 ), ss( 3 ),
                    fld( 0 ), val1,
                    fld( 1 ), val1,
                NewEndEvent(),
            NewEndEvent(),
        },
    }
    mkSrc := func( ord ...int ) []ReactorEvent {
        res := []ReactorEvent{ ss( 1 ) }
        for _, i := range ord {
            res = append( res, fld( i ) )
            res = append( res, evs[ i ]... )
        }
        return append( res, NewEndEvent() )
    }
    for _, ord := range [][]int{
        []int{ 0, 1, 2, 3, 4 },
        []int{ 4, 3, 2, 1, 0 },
        []int{ 2, 4, 0, 3, 1 },
    } {
        AddStdReactorTests(
            &FieldOrderPathTest{
                Source: copySource( mkSrc( ord... ) ),
                Expect: expct1,
                Orders: ords1,
            },
        )
    }
    AddStdReactorTests(
        &FieldOrderPathTest{
            Source: copySource(
                []ReactorEvent{
                    ss( 1 ),
                        fld( 0 ), val1,
                        fld( 7 ), val1,
                        fld( 2 ), val1,
                        fld( 1 ), val1,
                    NewEndEvent(),
                },
            ),
            Expect: []EventExpectation{
                { ss( 1 ), nil },
                { fld( 0 ), p( 0 ) },
                { val1, p( 0 ) },
                { fld( 7 ), p( 7 ) },
                { val1, p( 7 ) },
                { fld( 1 ), p( 1 ) },
                { val1, p( 1 ) },
                { fld( 2 ), p( 2 ) },
                { val1, p( 2 ) },
                { NewEndEvent(), nil },
            },
            Orders: []FieldOrderReactorTestOrder{
                testOrderWithIds( ss( 1 ).Type, id( 0 ), id( 1 ), id( 2 ) ),
            },
        },
    )
    // Regression for bug fixed in previous commit (f7fa84122047)
    AddStdReactorTests(
        &FieldOrderPathTest{
            Source: copySource(
                []ReactorEvent{ ss( 1 ), fld( 1 ), val1, NewEndEvent() } ),
            Expect: []EventExpectation{
                { ss( 1 ), nil },
                { fld( 1 ), p( 1 ) },
                { val1, p( 1 ) },
                { NewEndEvent(), nil },
            },
            Orders: []FieldOrderReactorTestOrder{
                testOrderWithIds( ss( 1 ).Type, id( 0 ), id( 1 ), id( 2 ) ),
            },
        },
    )
}

func initFieldOrderReactorTests() {
    initFieldOrderValueTests()
    initFieldOrderMissingFieldTests()
    initFieldOrderPathTests()
}

type ServiceRequestReactorTest struct {
    Source interface{}
    Namespace *Namespace
    Service *Identifier
    Operation *Identifier
    Parameters *SymbolMap
    ParameterEvents []EventExpectation
    Authentication Value
    AuthenticationEvents []EventExpectation
    Error error
}

func initServiceRequestTests() {
    ns1 := MustNamespace( "ns1@v1" )
    svc1 := id( "service1" )
    op1 := id( "op1" )
    params1 := MustSymbolMap( "f1", int32( 1 ) )
    authQn := qname( "ns1@v1/Auth1" )
    auth1 := MustStruct( authQn, "f1", int32( 1 ) )
    evFldNs := NewFieldStartEvent( IdNamespace )
    evFldSvc := NewFieldStartEvent( IdService )
    evFldOp := NewFieldStartEvent( IdOperation )
    evFldParams := NewFieldStartEvent( IdParameters )
    evFldAuth := NewFieldStartEvent( IdAuthentication )
    evFldF1 := NewFieldStartEvent( id( "f1" ) )
    evReqTyp := NewStructStartEvent( QnameServiceRequest )
    evNs1 := NewValueEvent( String( ns1.ExternalForm() ) )
    evSvc1 := NewValueEvent( String( svc1.ExternalForm() ) )
    evOp1 := NewValueEvent( String( op1.ExternalForm() ) )
    i32Val1 := NewValueEvent( Int32( 1 ) )
    evParams1 := []ReactorEvent{ 
        NewMapStartEvent(), evFldF1, i32Val1, NewEndEvent() }
    evAuth1 := []ReactorEvent{ 
        NewStructStartEvent( authQn ), evFldF1, i32Val1, NewEndEvent() }
    addSucc1 := func( evs ...interface{} ) {
        AddStdReactorTests(
            &ServiceRequestReactorTest{
                Source: copySource( flattenEvs( evs... ) ),
                Namespace: ns1,
                Service: svc1,
                Operation: op1,
                Parameters: params1,
                Authentication: auth1,
            },
        )
    }
    fullOrderedReq1Flds := []interface{}{
        evFldNs, evNs1,
        evFldSvc, evSvc1,
        evFldOp, evOp1,
        evFldParams, evParams1,
        evFldAuth, evAuth1,
    }
    addSucc1( evReqTyp, fullOrderedReq1Flds, NewEndEvent() )
    addSucc1( NewMapStartEvent(), fullOrderedReq1Flds, NewEndEvent() )
    addSucc1( evReqTyp,
        evFldAuth, evAuth1,
        evFldOp, evOp1,
        evFldParams, evParams1,
        evFldNs, evNs1,
        evFldSvc, evSvc1,
        NewEndEvent(),
    )
    AddStdReactorTests(
        &ServiceRequestReactorTest{
            Source: copySource(
                flattenEvs( evReqTyp,
                    evFldNs, evNs1,
                    evFldSvc, evSvc1,
                    evFldOp, evOp1,
                    evFldAuth, i32Val1,
                    evFldParams, evParams1,
                    NewEndEvent(),
                ),
            ),
            Namespace: ns1,
            Service: svc1,
            Operation: op1,
            Authentication: Int32( 1 ),
            Parameters: params1,
        },
    )
    mkReq1 := func( params, auth Value ) *Struct {
        pairs := []interface{}{ 
            IdNamespace, ns1.ExternalForm(),
            IdService, svc1.ExternalForm(),
            IdOperation, op1.ExternalForm(),
        }
        if params != nil { pairs = append( pairs, IdParameters, params ) }
        if auth != nil { pairs = append( pairs, IdAuthentication, auth ) }
        return MustStruct( QnameServiceRequest, pairs... )
    }
    addSucc2 := func( src interface{}, authExpct Value ) {
        AddStdReactorTests(
            &ServiceRequestReactorTest{
                Namespace: ns1,
                Service: svc1,
                Operation: op1,
                Parameters: EmptySymbolMap(),
                Authentication: authExpct,
                Source: src,
            },
        )
    } 
    // check implicit params with(out) auth and using undetermined event
    // ordering
    addSucc2( mkReq1( nil, nil ), nil )
    addSucc2( mkReq1( nil, auth1 ), auth1 )
    // check implicit params with and without auth and with need for reordering
    addSucc2(
        flattenEvs( evReqTyp, 
            evFldSvc, evSvc1, 
            evFldOp, evOp1, 
            evFldNs, evNs1,
            NewEndEvent(),
        ),
        nil,
    )
    addSucc2(
        flattenEvs( evReqTyp,
            evFldSvc, evSvc1,
            evFldAuth, evAuth1,
            evFldOp, evOp1,
            evFldNs, evNs1,
            NewEndEvent(),
        ),
        auth1,
    )
    addPathSucc := func( 
        paramsIn, paramsExpct *SymbolMap, paramEvs []EventExpectation,
        auth Value, authEvs []EventExpectation ) {
        t := &ServiceRequestReactorTest{
            Namespace: ns1,
            Service: svc1,
            Operation: op1,
            Parameters: paramsExpct,
            ParameterEvents: paramEvs,
            Authentication: auth,
            AuthenticationEvents: authEvs,
        }
        pairs := []interface{}{
            IdNamespace, ns1.ExternalForm(),
            IdService, svc1.ExternalForm(),
            IdOperation, op1.ExternalForm(),
        }
        if paramsIn != nil { pairs = append( pairs, IdParameters, paramsIn ) }
        if auth != nil { pairs = append( pairs, IdAuthentication, auth ) }
        t.Source = MustStruct( QnameServiceRequest, pairs... )
        AddStdReactorTests( t )
    }
    pathParams := objpath.RootedAt( IdParameters )
    evsEmptyParams := []EventExpectation{ 
        { NewMapStartEvent(), pathParams }, { NewEndEvent(), pathParams } }
    pathAuth := objpath.RootedAt( IdAuthentication )
    addPathSucc( nil, MustSymbolMap(), evsEmptyParams, nil, nil )
    addPathSucc( MustSymbolMap(), MustSymbolMap(), evsEmptyParams, nil, nil )
    idF1 := id( "f1" )
    addPathSucc(
        MustSymbolMap( idF1, Int32( 1 ) ),
        MustSymbolMap( idF1, Int32( 1 ) ),
        []EventExpectation{
            { NewMapStartEvent(), pathParams },
            { evFldF1, pathParams.Descend( idF1 ) },
            { i32Val1, pathParams.Descend( idF1 ) },
            { NewEndEvent(), pathParams },
        },
        nil, nil,
    )
    addPathSucc( 
        nil, MustSymbolMap(), evsEmptyParams,
        Int32( 1 ), []EventExpectation{ { i32Val1, pathAuth } },
    )
    addPathSucc(
        nil, MustSymbolMap(), evsEmptyParams,
        auth1, []EventExpectation{
            { NewStructStartEvent( authQn ), pathAuth },
            { evFldF1, pathAuth.Descend( idF1 ) },
            { i32Val1, pathAuth.Descend( idF1 ) },
            { NewEndEvent(), pathAuth },
        },
    )
    writeMgIo := func( f func( w *BinWriter ) ) Buffer {
        bb := &bytes.Buffer{}
        w := NewWriter( bb )
        f( w )
        return Buffer( bb.Bytes() )
    }
    nsBuf := func( ns *Namespace ) Buffer {
        return writeMgIo( func( w *BinWriter ) { w.WriteNamespace( ns ) } )
    }
    idBuf := func( id *Identifier ) Buffer {
        return writeMgIo( func( w *BinWriter ) { w.WriteIdentifier( id ) } )
    }
    AddStdReactorTests(
        &ServiceRequestReactorTest{
            Namespace: ns1,
            Service: svc1,
            Operation: op1,
            Parameters: EmptySymbolMap(),
            Source: MustStruct( QnameServiceRequest,
                IdNamespace, nsBuf( ns1 ),
                IdService, idBuf( svc1 ),
                IdOperation, idBuf( op1 ),
            ),
        },
    )
    addReqVcErr := func( val interface{}, path idPath, msg string ) {
        AddStdReactorTests(
            &ServiceRequestReactorTest{
                Source: MustValue( val ),
                Error: NewValueCastError( path, msg ),
            },
        )
    }
    addReqVcErr(
        MustSymbolMap( IdNamespace, true ), 
        objpath.RootedAt( IdNamespace ),
        "invalid value: mingle:core@v1/Boolean",
    )
    addReqVcErr(
        MustSymbolMap( IdNamespace, MustSymbolMap() ),
        objpath.RootedAt( IdNamespace ),
        "invalid value: mingle:core@v1/SymbolMap",
    )
    addReqVcErr(
        MustSymbolMap( IdNamespace, MustStruct( "ns1@v1/S1" ) ),
        objpath.RootedAt( IdNamespace ),
        "invalid value: ns1@v1/S1",
    )
    addReqVcErr(
        MustSymbolMap( IdNamespace, MustList() ),
        objpath.RootedAt( IdNamespace ),
        "invalid value: mingle:core@v1/Value*",
    )
    addReqVcErr(
        MustSymbolMap( IdNamespace, ns1.ExternalForm(), IdService, true ),
        objpath.RootedAt( IdService ),
        "invalid value: mingle:core@v1/Boolean",
    )
    addReqVcErr(
        MustSymbolMap( 
            IdNamespace, ns1.ExternalForm(),
            IdService, svc1.ExternalForm(),
            IdOperation, true,
        ),
        objpath.RootedAt( IdOperation ),
        "invalid value: mingle:core@v1/Boolean",
    )
    AddStdReactorTests(
        &ServiceRequestReactorTest{
            Source: MustSymbolMap(
                IdNamespace, ns1.ExternalForm(),
                IdService, svc1.ExternalForm(),
                IdOperation, op1.ExternalForm(),
                IdParameters, true,
            ),
            Error: NewTypeCastError(
                TypeSymbolMap,
                TypeBoolean,
                objpath.RootedAt( IdParameters ),
            ),
        },
    )
    // Check that errors are bubbled up from
    // *BinWriter.Read(Identfier|Namespace) when parsing invalid
    // namespace/service/operation Buffers
    addBinRdErr := func( path *Identifier, msg string, pairs ...interface{} ) {
        addReqVcErr(
            MustSymbolMap( pairs... ),
            objpath.RootedAt( path ),
            msg,
        )
    }
    badBuf := []byte{ 0x0f }
    addBinRdErr( 
        IdNamespace, 
        "Expected type code 0x02 but got 0x0f",
        IdNamespace, badBuf )
    addBinRdErr( 
        IdService, 
        "Expected type code 0x01 but got 0x0f",
        IdNamespace, ns1.ExternalForm(), 
        IdService, badBuf,
    )
    addBinRdErr( 
        IdOperation, 
        "Expected type code 0x01 but got 0x0f",
        IdNamespace, ns1.ExternalForm(),
        IdService, svc1.ExternalForm(),
        IdOperation, badBuf,
    )
    addReqVcErr(
        MustSymbolMap( IdNamespace, "ns1::ns2" ),
        objpath.RootedAt( IdNamespace ),
        "[<input>, line 1, col 5]: Illegal start of identifier part: \":\" " +
        "(U+003A)",
    )
    addReqVcErr(
        MustSymbolMap( IdNamespace, ns1.ExternalForm(), IdService, "2bad" ),
        objpath.RootedAt( IdService ),
        "[<input>, line 1, col 1]: Illegal start of identifier part: \"2\" " +
        "(U+0032)",
    )
    addReqVcErr(
        MustSymbolMap(
            IdNamespace, ns1.ExternalForm(),
            IdService, svc1.ExternalForm(),
            IdOperation, "2bad",
        ),
        objpath.RootedAt( IdOperation ),
        "[<input>, line 1, col 1]: Illegal start of identifier part: \"2\" " +
        "(U+0032)",
    )
    t1Bad := qname( "foo@v1/Request" )
    AddStdReactorTests(
        &ServiceRequestReactorTest{
            Source: MustStruct( t1Bad ),
            Error: NewTypeCastError(
                TypeServiceRequest, t1Bad.AsAtomicType(), nil ),
        },
    )
    // Not exhaustively re-testing all ways a field could be missing (assume for
    // now that field order tests will handle that). Instead, we are just
    // getting basic coverage that the field order supplied by the request
    // reactor is in fact being set up correctly and that we have set up the
    // right required fields.
    AddStdReactorTests(
        &ServiceRequestReactorTest{
            Source: MustSymbolMap( 
                IdNamespace, ns1.ExternalForm(),
                IdOperation, op1.ExternalForm(),
            ),
            Error: NewMissingFieldsError( nil, []*Identifier{ IdService } ),
        },
    )
}

type ServiceResponseReactorTest struct {
    In Value
    ResVal Value
    ResEvents []EventExpectation
    ErrVal Value
    ErrEvents []EventExpectation
    Error error
}

func initServiceResponseTests() {
    addSucc := func( in, res, err Value ) {
        AddStdReactorTests(
            &ServiceResponseReactorTest{ In: in, ResVal: res, ErrVal: err } )
    }
    i32Val1 := Int32( 1 )
    err1 := MustStruct( "ns1@v1/Err1", "f1", int32( 1 ) )
    addSucc( MustStruct( QnameServiceResponse ), nil, nil )
    addSucc( MustSymbolMap(), nil, nil )
    addSucc( MustSymbolMap( IdResult, NullVal ), NullVal, nil )
    addSucc( MustSymbolMap( IdResult, i32Val1 ), i32Val1, nil )
    addSucc( MustSymbolMap( IdError, NullVal ), nil, NullVal )
    addSucc( MustSymbolMap( IdError, err1 ), nil, err1 )
    addSucc( MustSymbolMap( IdError, int32( 1 ) ), nil, i32Val1 )
    pathRes := objpath.RootedAt( IdResult )
    pathResF1 := pathRes.Descend( id( "f1" ) )
    pathErr := objpath.RootedAt( IdError )
    pathErrF1 := pathErr.Descend( id( "f1" ) )
    AddStdReactorTests(
        &ServiceResponseReactorTest{
            In: MustStruct( QnameServiceResponse, "result", int32( 1 ) ),
            ResVal: i32Val1,
            ResEvents: []EventExpectation{ 
                { NewValueEvent( i32Val1 ), pathRes },
            },
        },
        &ServiceResponseReactorTest{
            In: MustSymbolMap( "result", MustSymbolMap( "f1", int32( 1 ) ) ),
            ResVal: MustSymbolMap( "f1", int32( 1 ) ),
            ResEvents: []EventExpectation{
                { NewMapStartEvent(), pathRes },
                { NewFieldStartEvent( id( "f1" ) ), pathResF1 },
                { NewValueEvent( i32Val1 ), pathResF1 },
                { NewEndEvent(), pathRes },
            },
        },
        &ServiceResponseReactorTest{
            In: MustSymbolMap( "error", int32( 1 ) ),
            ErrVal: i32Val1,
            ErrEvents: []EventExpectation{ 
                { NewValueEvent( i32Val1 ), pathErr },
            },
        },
        &ServiceResponseReactorTest{
            In: MustSymbolMap( "error", err1 ),
            ErrVal: err1,
            ErrEvents: []EventExpectation{
                { NewStructStartEvent( err1.Type ), pathErr },
                { NewFieldStartEvent( id( "f1" ) ), pathErrF1 },
                { NewValueEvent( i32Val1 ), pathErrF1 },
                { NewEndEvent(), pathErr },
            },
        },
    )
    addFail := func( in Value, err error ) {
        AddStdReactorTests( &ServiceResponseReactorTest{ In: in, Error: err } )
    }
    addFail(
        err1.Fields,
        NewUnrecognizedFieldError( nil, id( "f1" ) ),
    )
    addFail(
        MustStruct( "ns1@v1/ServiceResponse" ),
        NewTypeCastError( 
            TypeServiceResponse, 
            MustTypeReference( "ns1@v1/ServiceResponse" ),
            nil, 
        ),
    )
    addFail(
        MustSymbolMap( IdResult, i32Val1, IdError, err1 ),
        NewValueCastError( 
            nil, "response has both a result and an error value" ),
    )
}

func initServiceTests() {
    initServiceRequestTests()
    initServiceResponseTests()
}

type CastReactorTest struct {
    In Value
    Expect Value
    Type TypeReference
    Path objpath.PathNode
    Err error
    Profile string
}

var crtPathDefault = objpath.RootedAt( MustIdentifier( "inVal" ) )

type crtInit struct {
    buf1 Buffer
    tm1 Timestamp
    map1 *SymbolMap
    en1 *Enum
    struct1 *Struct
}

func ( t *crtInit ) initStdVals() {
    t.buf1 = Buffer( []byte{ byte( 0 ), byte( 1 ), byte( 2 ) } )
    t.tm1 = MustTimestamp( "2007-08-24T13:15:43.123450000-08:00" )
    t.map1 = MustSymbolMap( "key1", 1, "key2", "val2" )
    t.en1 = MustEnum( "ns1@v1/En1", "en-val1" )
    t.struct1 = MustStruct( "ns1@v1/S1", "key1", "val1" )
}

func ( t *crtInit ) addCrt( crt *CastReactorTest ) { 
    StdReactorTests = append( StdReactorTests, crt ) 
}

func ( t *crtInit ) addCrtDefault( crt *CastReactorTest ) {
    crt.Path = crtPathDefault
    t.addCrt( crt )
}

func ( t *crtInit ) createSucc(
    in, expct interface{}, typ TypeReferenceInitializer ) *CastReactorTest {
    return &CastReactorTest{ 
        In: MustValue( in ), 
        Expect: MustValue( expct ), 
        Type: asTypeReference( typ ),
    }
}

func ( t *crtInit ) addSucc( 
    in, expct interface{}, typ TypeReferenceInitializer ) {
    t.addCrtDefault( t.createSucc( in, expct, typ ) )
}

func ( t *crtInit ) addIdent( in interface{}, typ TypeReferenceInitializer ) {
    v := MustValue( in )
    t.addSucc( v, v, asTypeReference( typ ) )
}

func ( t *crtInit ) addBaseTypeTests() {
    t.addIdent( Boolean( true ), TypeBoolean )
    t.addIdent( t.buf1, TypeBuffer )
    t.addIdent( "s", TypeString )
    t.addIdent( Int32( 1 ), TypeInt32 )
    t.addIdent( Int64( 1 ), TypeInt64 )
    t.addIdent( Uint32( 1 ), TypeUint32 )
    t.addIdent( Uint64( 1 ), TypeUint64 )
    t.addIdent( Float32( 1.0 ), TypeFloat32 )
    t.addIdent( Float64( 1.0 ), TypeFloat64 )
    t.addIdent( t.tm1, TypeTimestamp )
    t.addIdent( t.en1, t.en1.Type )
    t.addIdent( t.map1, TypeSymbolMap )
    t.addIdent( t.struct1, t.struct1.Type )
    t.addIdent( 1, TypeValue )
    t.addIdent( nil, TypeNull )
    t.addSucc( Int32( -1 ), Uint32( uint32( 4294967295 ) ), TypeUint32 )
    t.addSucc( Int64( -1 ), Uint32( uint32( 4294967295 ) ), TypeUint32 )
    t.addSucc( 
        Int32( -1 ), Uint64( uint64( 18446744073709551615 ) ), TypeUint64 )
    t.addSucc( 
        Int64( -1 ), Uint64( uint64( 18446744073709551615 ) ), TypeUint64 )
    t.addSucc( "true", true, TypeBoolean )
    t.addSucc( "TRUE", true, TypeBoolean )
    t.addSucc( "TruE", true, TypeBoolean )
    t.addSucc( "false", false, TypeBoolean )
    t.addSucc( true, "true", TypeString )
    t.addSucc( false, "false", TypeString )
}

func ( t *crtInit ) createTcError0(
    in interface{}, 
    typExpct, typAct, callTyp TypeReferenceInitializer, 
    p idPath ) *CastReactorTest {
    err := NewTypeCastError( 
        asTypeReference( typExpct ),
        asTypeReference( typAct ),
        p,
    )
    return &CastReactorTest{
        In: MustValue( in ),
        Type: asTypeReference( callTyp ),
        Err: err,
    }
}

func ( t *crtInit ) addTcError0(
    in interface{}, 
    typExpct, typAct, callTyp TypeReferenceInitializer, 
    p idPath ) {
    t.addCrtDefault( t.createTcError0( in, typExpct, typAct, callTyp, p ) )
}

func ( t *crtInit ) createTcError(
    in interface{}, 
    typExpct, typAct TypeReferenceInitializer ) *CastReactorTest {
    return t.createTcError0( in, typExpct, typAct, typExpct, crtPathDefault )
}

func ( t *crtInit ) addTcError(
    in interface{}, typExpct, typAct TypeReferenceInitializer ) {
    t.addTcError0( in, typExpct, typAct, typExpct, crtPathDefault )
}

func ( t *crtInit ) addMiscTcErrors() {
    t.addTcError( t.en1, "ns1@v1/Bad", t.en1.Type )
    t.addTcError( t.struct1, "ns1@v1/Bad", t.struct1.Type )
    t.addTcError( "s", TypeNull, TypeString )
    t.addTcError( int32( 1 ), "Buffer", "Int32" )
    t.addTcError( int32( 1 ), "Buffer?", "Int32" )
    t.addTcError( true, "Float32", "Boolean" )
    t.addTcError( true, "Float32?", "Boolean" )
    t.addTcError( true, "Int32", "Boolean" )
    t.addTcError( true, "Int32?", "Boolean" )
    t.addTcError( MustList( 1, 2 ), TypeString, "Value*" )
    t.addTcError( MustList(), "String?", "Value*" )
    t.addTcError( "s", "String*", "String" )
    t.addCrtDefault(
        &CastReactorTest{
            In: MustList( 1, t.struct1 ),
            Type: asTypeReference( "Int32*" ),
            Err: NewTypeCastError(
                asTypeReference( "Int32" ),
                &AtomicTypeReference{ Name: t.struct1.Type },
                crtPathDefault.StartList().Next(),
            ),
        },
    )
    t.addTcError( t.struct1, "Int32?", t.struct1.Type )
    t.addTcError( 12, t.struct1.Type, "Int64" )
    for _, prim := range PrimitiveTypes {
        // not an err for prims Value and SymbolMap
        if ! ( prim == TypeValue || prim == TypeSymbolMap ) { 
            t.addTcError( t.struct1, prim, t.struct1.Type )
        }
    }
}

func ( t *crtInit ) createVcError0(
    val interface{}, 
    typ TypeReferenceInitializer, 
    path idPath, 
    msg string ) *CastReactorTest {
    return &CastReactorTest{
        In: MustValue( val ),
        Type: asTypeReference( typ ),
        Err: NewValueCastError( path, msg ),
    }
}
    

func ( t *crtInit ) addVcError0( 
    val interface{}, typ TypeReferenceInitializer, path idPath, msg string ) {
    t.addCrtDefault( t.createVcError0( val, typ, path, msg ) )
}

func ( t *crtInit ) createVcError(
    val interface{}, 
    typ TypeReferenceInitializer, 
    msg string ) *CastReactorTest {
    return t.createVcError0( val, typ, crtPathDefault, msg )
}

func ( t *crtInit ) addVcError( 
    val interface{}, typ TypeReferenceInitializer, msg string ) {
    t.addVcError0( val, typ, crtPathDefault, msg )
}

func ( t *crtInit ) addMiscVcErrors() {
    t.addVcError( "s", TypeBoolean, `Invalid boolean value: "s"` )
    t.addVcError( nil, TypeString, "Value is null" )
    t.addVcError( nil, `String~"a"`, "Value is null" )
    t.addVcError( MustList(), "String+", "List is empty" )
    t.addVcError0( 
        MustList( MustList( int32( 1 ), int32( 2 ) ), MustList() ), 
        "Int32+*", 
        crtPathDefault.StartList().Next(),
        "List is empty",
    )
}

func ( t *crtInit ) addStringTests() {
    t.addIdent( "s", "String?" )
    t.addIdent( "abbbc", `String~"^ab+c$"` )
    t.addIdent( "abbbc", `String~"^ab+c$"?` )
    t.addIdent( nil, `String~"^ab+c$"?` )
    t.addIdent( "", `String~"^a*"?` )
    t.addSucc( 
        MustList( "123", 129 ), 
        MustList( "123", "129" ),
        `String~"^\\d+$"*`,
    )
    for _, quant := range []string { "*", "+", "?*", "*?" } {
        val := MustList( "a", "aaaaaa" )
        t.addSucc( val, val, `String~"^a+$"` + quant )
    }
    t.addVcError( 
        "ac", 
        `String~"^ab+c$"`,
        `Value "ac" does not satisfy restriction "^ab+c$"`,
    )
    t.addVcError(
        "ab",
        `String~"^a*$"?`,
        "Value \"ab\" does not satisfy restriction \"^a*$\"",
    )
    t.addVcError0(
        MustList( "a", "b" ),
        `String~"^a+$"*`,
        crtPathDefault.StartList().Next(),
        "Value \"b\" does not satisfy restriction \"^a+$\"",
    )
}

func ( t *crtInit ) addIdentityNumTests() {
    t.addIdent( int64( 1 ), "Int64~[-1,1]" )
    t.addIdent( int64( 1 ), "Int64~(,2)" )
    t.addIdent( int64( 1 ), "Int64~[1,1]" )
    t.addIdent( int64( 1 ), "Int64~[-2, 32)" )
    t.addIdent( int32( 1 ), "Int32~[-2, 32)" )
    t.addIdent( uint32( 3 ), "Uint32~[2,32)" )
    t.addIdent( uint64( 3 ), "Uint64~[2,32)" )
    t.addIdent( Float32( -1.1 ), "Float32~[-2.0,32)" )
    t.addIdent( Float64( -1.1 ), "Float64~[-2.0,32)" )
    numTests := []struct{ val Value; str string; typ TypeReference } {
        { val: Int32( 1 ), str: "1", typ: TypeInt32 },
        { val: Int64( 1 ), str: "1", typ: TypeInt64 },
        { val: Uint32( 1 ), str: "1", typ: TypeUint32 },
        { val: Uint64( 1 ), str: "1", typ: TypeUint64 },
        { val: Float32( 1.0 ), str: "1", typ: TypeFloat32 },
        { val: Float64( 1.0 ), str: "1", typ: TypeFloat64 },
    }
    for _, numCtx := range numTests {
        t.addSucc( numCtx.val, numCtx.str, TypeString )
        t.addSucc( numCtx.str, numCtx.val, numCtx.typ )
        for _, valCtx := range numTests {
            t.addSucc( valCtx.val, numCtx.val, numCtx.typ )
        }
    }
}

func ( t *crtInit ) addTruncateNumTests() {
    posVals := []Value{ Float32( 1.1 ), Float64( 1.1 ), String( "1.1" ) }
    for _, val := range posVals {
        t.addSucc( val, Int32( 1 ), TypeInt32 )
        t.addSucc( val, Int64( 1 ), TypeInt64 )
        t.addSucc( val, Uint32( 1 ), TypeUint32 )
        t.addSucc( val, Uint64( 1 ), TypeUint64 )
    }
    negVals := []Value{ Float32( -1.1 ), Float64( -1.1 ), String( "-1.1" ) }
    for _, val := range negVals {
        t.addSucc( val, Int32( -1 ), TypeInt32 )
        t.addSucc( val, Int64( -1 ), TypeInt64 )
    }
    t.addSucc( int64( 1 << 31 ), int32( -2147483648 ), TypeInt32 )
    t.addSucc( int64( 1 << 33 ), int32( 0 ), TypeInt32 )
    t.addSucc( int64( 1 << 31 ), uint32( 1 << 31 ), TypeUint32 )
}

func ( t *crtInit ) addNumTests() {
    for _, typ := range NumericTypes {
        t.addVcError( "not-a-num", typ, `invalid syntax: not-a-num` )
    }
    t.addIdentityNumTests()
    t.addTruncateNumTests()
    t.addSucc( "1", int64( 1 ), "Int64~[-1,1]" ) // just cover String with range
    rngErr := func( val string, typ TypeReference ) {
        t.addVcError( val, typ, fmt.Sprintf( "value out of range: %s", val ) )
    }
    rngErr( "2147483648", TypeInt32 )
    rngErr( "-2147483649", TypeInt32 )
    rngErr( "9223372036854775808", TypeInt64 )
    rngErr( "-9223372036854775809", TypeInt64 )
    rngErr( "4294967296", TypeUint32 )
    t.addVcError( "-1", TypeUint32, "value out of range: -1" )
    rngErr( "18446744073709551616", TypeUint64 )
    t.addVcError( "-1", TypeUint64, "value out of range: -1" )
    t.addVcError(
        12, "Int32~[0,10)", "Value 12 does not satisfy restriction [0,10)" )
}

func ( t *crtInit ) addBufferTests() {
    buf1B64 := base64.StdEncoding.EncodeToString( t.buf1 )
    t.addSucc( t.buf1, buf1B64, TypeString )
    t.addSucc( buf1B64, t.buf1, TypeBuffer  )
    t.addVcError( "abc$/@", TypeBuffer, 
        "Invalid base64 string: illegal base64 data at input byte 3" )
}

func ( t *crtInit ) addTimeTests() {
    t.addIdent(
        Now(), `Timestamp~["1970-01-01T00:00:00Z","2200-01-01T00:00:00Z"]` )
    t.addSucc( t.tm1, t.tm1.Rfc3339Nano(), TypeString )
    t.addSucc( t.tm1.Rfc3339Nano(), t.tm1, TypeTimestamp )
    t.addVcError(
        MustTimestamp( "2012-01-01T00:00:00Z" ),
        `mingle:core@v1/Timestamp~` +
            `["2000-01-01T00:00:00Z","2001-01-01T00:00:00Z"]`,
        "Value 2012-01-01T00:00:00Z does not satisfy restriction " +
            "[\"2000-01-01T00:00:00Z\",\"2001-01-01T00:00:00Z\"]",
    )
}

func ( t *crtInit ) addEnumTests() {
    t.addSucc( t.en1, "en-val1", TypeString  )
}

func ( t *crtInit ) addNullableTests() {
    typs := []TypeReference{}
    for _, prim := range PrimitiveTypes {
        typs = append( typs, &NullableTypeReference{ prim } )
    }
    typs = append( typs,
        MustTypeReference( "Null" ),
        MustTypeReference( "String??" ),
        MustTypeReference( "String*?" ),
        MustTypeReference( "String+?" ),
        MustTypeReference( "ns1@v1/T?" ),
        MustTypeReference( "ns1@v1/T*?" ),
    )
    for _, typ := range typs { t.addSucc( nil, nil, typ ) }
}

func ( t *crtInit ) addListTests() {
    for _, quant := range []string{ "*", "**", "***" } {
        t.addSucc( []interface{}{}, MustList(), "Int64" + quant )
    }
    for _, quant := range []string{ "**", "*+" } {
        v := MustList( MustList(), MustList() )
        t.addIdent( v, "Int64" + quant )
    }
    // test conversions in a deeply nested list
    t.addSucc(
        []interface{}{
            []interface{}{ "1", int32( 1 ), int64( 1 ) },
            []interface{}{ float32( 1.0 ), float64( 1.0 ) },
            []interface{}{},
        },
        MustList(
            MustList( Int64( 1 ), Int64( 1 ), Int64( 1 ) ),
            MustList( Int64( 1 ), Int64( 1 ) ),
            MustList(),
        ),
        "Int64**",
    )
    t.addSucc(
        []interface{}{ int64( 1 ), nil, "hi" },
        MustList( "1", nil, "hi" ),
        "String?*",
    )
    t.addSucc( MustList(), MustList(), TypeValue )
    intList1 := MustList( int32( 1 ), int32( 2 ), int32( 3 ) )
    t.addSucc( intList1, intList1, TypeValue )
    t.addSucc( intList1, intList1, TypeOpaqueList )
    t.addSucc( intList1, intList1, "Int32*?" )
}

func ( t *crtInit ) addMapTests() {
    m1 := MustSymbolMap()
    m2 := MustSymbolMap( "f1", int32( 1 ) )
    t.addSucc( m1, m1, TypeSymbolMap )
    t.addSucc( m1, m1, TypeValue )
    t.addSucc( m2, m2, TypeSymbolMap )
    t.addSucc( m2, m2, &NullableTypeReference{ TypeSymbolMap } )
    s2 := &Struct{ Type: qname( "ns2@v1/S1" ), Fields: m2 }
    t.addSucc( s2, m2, TypeSymbolMap )
    l1 := MustList()
    l2 := MustList( m1, m2 )
    lt1 := &ListTypeReference{ TypeSymbolMap, true }
    lt2 := &ListTypeReference{ TypeSymbolMap, false }
    t.addSucc( l1, l1, lt1 )
    t.addSucc( l2, l2, lt2 )
    t.addSucc( MustList( s2, s2 ), MustList( m2, m2 ), lt2 )
    t.addTcError( int32( 1 ), TypeSymbolMap, TypeInt32 )
    t.addTcError0(
        MustList( m1, int32( 1 ) ),
        TypeSymbolMap,
        TypeInt32,
        lt2,
        crtPathDefault.StartList().SetIndex( 1 ),
    )
    nester := MustSymbolMap( "f1", MustSymbolMap( "f2", int32( 1 ) ) )
    t.addSucc( nester, nester, TypeSymbolMap )
}

func ( t *crtInit ) addStructTests() {
    qn1 := qname( "ns1@v1/T1" )
    t1 := qn1.AsAtomicType()
    s1 := MustStruct( qn1 )
    s2 := MustStruct( qn1, "f1", int32( 1 ) )
    qn2 := qname( "ns1@v1/T2" )
    t2 := qn2.AsAtomicType()
    s3 := MustStruct( qn2,
        "f1", int32( 1 ),
        "f2", s1,
        "f3", s2,
        "f4", MustList( s1, s2 ),
    )
    t.addSucc( s1, s1, TypeValue )
    t.addSucc( s1, s1, t1 )
    t.addSucc( s2, s2, t1 )
    t.addSucc( s1, s1, &NullableTypeReference{ t1 } )
    t.addSucc( s3, s3, t2 )
    l1 := MustList( s1, s2 )
    t.addSucc( l1, l1, &ListTypeReference{ t1, false } )
    t.addSucc( l1, l1, &ListTypeReference{ t1, true } )
    f1 := func( in interface{}, inTyp TypeReferenceInitializer ) {
        t.addTcError0(
            MustList( s1, in ),
            t1,
            inTyp,
            &ListTypeReference{ t1, false },
            crtPathDefault.StartList().SetIndex( 1 ),
        )
    }
    f1( s3, t2 )
    f1( int32( 1 ), "Int32" )
}

func ( t *crtInit ) addInterfaceImplTests() {
    add := func( crt *CastReactorTest ) {
        crt.Profile = "interface-impl"
        t.addCrtDefault( crt )
    }
    addSucc := func( in, expct interface{}, typ TypeReferenceInitializer ) {
        add( t.createSucc( in, expct, typ ) )
    }
    t1 := qname( "ns1@v1/T1" )
    t2 := qname( "ns1@v1/T2" )
    s1 := MustStruct( t1, "f1", int32( 1 ) )
    addSucc( MustStruct( t1, "f1", "1" ), s1, t1 )
    addSucc( MustSymbolMap( "f1", "1" ), s1, t1 )
    addSucc( "cast1", int32( 1 ), "ns1@v1/S3" )
    addSucc( "cast2", int32( -1 ), "ns1@v1/S3" )
    addSucc( 
        MustList( "cast1", "cast2" ), 
        MustList( int32( 1 ), int32( -1 ) ),
        "ns1@v1/S3*",
    )
    addSucc( nil, nil, "ns1@v1/S3?" )
    arb := MustStruct( "ns1@v1/Arbitrary", "f1", int32( 1 ) )
    addSucc( arb, arb, arb.Type )
    add( t.createTcError( int32( 1 ), "ns1@v1/S3", TypeInt32 ) )
    add( 
        t.createTcError0( 
            int32( 1 ), 
            "ns1@v1/S3", 
            TypeInt32, 
            "ns1@v1/S3?", 
            crtPathDefault,
        ),
    )
    add( 
        t.createTcError0( 
            MustList( int32( 1 ) ),
            "ns1@v1/S3",
            TypeInt32,
            "ns1@v1/S3*",
            crtPathDefault.StartList(),
        ),
    )
    add( t.createVcError( "cast3", "ns1@v1/S3", "test-message-cast3" ) )
    add( t.createVcError( "cast3", "ns1@v1/S3?", "test-message-cast3" ) )
    add(
        t.createVcError0( 
            MustList( "cast2", "cast3" ),
            "ns1@v1/S3+",
            crtPathDefault.StartList().SetIndex( 1 ),
            "test-message-cast3",
        ),
    )
    s2InFlds := MustSymbolMap( "f1", "1", "f2", MustSymbolMap( "f1", "1" ) )
    s2 := MustStruct( t2, "f1", int32( 1 ), "f2", s1 )
    addSucc( &Struct{ Type: t2, Fields: s2InFlds }, s2, t2 )
    addSucc( s2InFlds, s2, t2 )
    add( t.createTcError( MustStruct( t2, "f1", int32( 1 ) ), t1, t2 ) )
    add( 
        t.createTcError0(
            MustStruct( t1, "f1", MustList( 1, 2 ) ),
            TypeInt32,
            TypeOpaqueList,
            t1,
            crtPathDefault.Descend( id( "f1" ) ),
        ),
    )
    extraFlds1 := MustSymbolMap( "f1", int32( 1 ), "x1", int32( 0 ) )
    failExtra1 := func( val interface{} ) {
        msg := "unrecognized field: x1"
        add( t.createVcError0( val, t1, crtPathDefault, msg ) )
    }
    failExtra1( &Struct{ Type: t1, Fields: extraFlds1 } )
    failExtra1( extraFlds1 )
    failTyp := qname( "ns1@v1/FailType" )
    add(
        t.createVcError0(
            MustStruct( failTyp ), 
            failTyp, 
            crtPathDefault,
            "test-message-fail-type",
        ),
    )
}

func ( t *crtInit ) call() {
    t.initStdVals()
    t.addBaseTypeTests()
    t.addMiscTcErrors()
    t.addMiscVcErrors()
    t.addNumTests()
    t.addStringTests()
    t.addBufferTests()
    t.addTimeTests()
    t.addEnumTests()
    t.addNullableTests()
    t.addListTests()
    t.addMapTests()
    t.addStructTests()
    t.addInterfaceImplTests()
}

func initCastReactorTests() { ( &crtInit{} ).call() }

func init() {
    StdReactorTests = []interface{}{}
    initValueBuildReactorTests()
    initStructuralReactorTests()
    initEventPathTests()
    initFieldOrderReactorTests()
    initServiceTests()
    initCastReactorTests()
}

type CastErrorAssert struct {
    ErrExpect, ErrAct error
    *assert.PathAsserter
}

func ( cea CastErrorAssert ) FailActErrType() {
    cea.Fatalf(
        "Expected error of type %T but got %T: %s",
        cea.ErrExpect, cea.ErrAct, cea.ErrAct )
}

// Returns a path asserter that can be used further
func ( cea CastErrorAssert ) assertValueError( 
    expct, act ValueError ) *assert.PathAsserter {
    a := cea.Descend( "Err" )
    a.Descend( "Error()" ).Equal( expct.Error(), act.Error() )
    a.Descend( "Message()" ).Equal( expct.Message(), act.Message() )
    a.Descend( "Location()" ).Equal( expct.Location(), act.Location() )
    return a
}

func ( cea CastErrorAssert ) assertVcError() {
    if act, ok := cea.ErrAct.( *ValueCastError ); ok {
        cea.assertValueError( cea.ErrExpect.( *ValueCastError ), act )
    } else { cea.FailActErrType() }
}

func ( cea CastErrorAssert ) Call() {
    switch cea.ErrExpect.( type ) {
    case nil: cea.Fatal( cea.ErrAct )
    case *ValueCastError: cea.assertVcError()
    default: cea.Fatalf( "Unhandled Err type: %T", cea.ErrExpect )
    }
}

func AssertCastError( expct, act error, pa *assert.PathAsserter ) {
    ca := CastErrorAssert{ ErrExpect: expct, ErrAct: act, PathAsserter: pa }
    ca.Call()
}

func EqualEvents( expct, act ReactorEvent, a *assert.PathAsserter ) {
    a.Equalf( expct, act, "events are not equal: %s != %s",
        EventToString( expct ), EventToString( act ) )
}

type reactorEventSource interface {
    Len() int
    EventAt( int ) ReactorEvent
}

func FeedEventSource( 
    src reactorEventSource, 
    proc ReactorEventProcessor,
) error {
    for i, e := 0, src.Len(); i < e; i++ {
        if err := proc.ProcessEvent( src.EventAt( i ) ); err != nil { 
            return err
        }
    }
    return nil
}

func AssertFeedEventSource(
    src reactorEventSource, proc ReactorEventProcessor, a assert.Failer ) {
    
    if err := FeedEventSource( src, proc ); err != nil { a.Fatal( err ) }
}

type eventSliceSource []ReactorEvent
func ( src eventSliceSource ) Len() int { return len( src ) }
func ( src eventSliceSource ) EventAt( i int ) ReactorEvent { return src[ i ] }

type eventExpectSource []EventExpectation

func ( src eventExpectSource ) Len() int { return len( src ) }

func ( src eventExpectSource ) EventAt( i int ) ReactorEvent {
    return src[ i ].Event
}

type eventPathCheckReactor struct {
    a *assert.PathAsserter
    eeAssert *assert.PathAsserter
    expct []EventExpectation
    idx int
}

func ( r *eventPathCheckReactor ) ProcessEvent( ev ReactorEvent ) error {
    r.a.Truef( r.idx < len( r.expct ), "unexpected event: %v", ev )
    ee := r.expct[ r.idx ]
    r.idx++
    ee.Event.SetPath( ee.Path )
    EqualEvents( ee.Event, ev, r.eeAssert )
    r.eeAssert = r.eeAssert.Next()
    return nil
}

func ( r *eventPathCheckReactor ) complete() {
    r.a.Equalf( r.idx, len( r.expct ), "not all events were seen" )
}

func newEventPathCheckReactor( 
    expct []EventExpectation, a *assert.PathAsserter ) *eventPathCheckReactor {

    return &eventPathCheckReactor{ 
        expct: expct, 
        a: a,
        eeAssert: a.Descend( "expct" ).StartList(),
    }
}

type ReactorTestCall struct {
    *assert.PathAsserter
    Test interface{}
}

func ( c *ReactorTestCall ) CheckNoError( err error ) {
    if err != nil { c.Fatalf( "Got no error but expected %T: %s", err, err ) }
}

func ( c *ReactorTestCall ) EqualErrors( expct, act error ) {
    if expct == nil { c.Fatal( act ) }
    c.Equalf( expct, act, "expected %q (%T) but got %q (%T)",
        expct, expct, act, act )
}

func CheckBuiltValue( expct Value, vb *ValueBuilder, a *assert.PathAsserter ) {
    if expct == nil {
        if vb != nil {
            a.Fatalf( "unexpected value: %s", QuoteValue( vb.GetValue() ) )
        }
    } else { 
        a.Falsef( vb == nil, 
            "expecting value %s but value builder is nil", QuoteValue( expct ) )
        EqualValues( expct, vb.GetValue(), a ) 
    }
}
