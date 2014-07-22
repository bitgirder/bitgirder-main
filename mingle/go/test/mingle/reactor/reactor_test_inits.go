package reactor

import (
    mg "mingle"
    "mingle/parser"
    "bitgirder/objpath"
    "fmt"
)

func initValueBuildReactorTests( b *ReactorTestSetBuilder ) {
    s1 := parser.MustStruct( "ns1@v1/S1",
        "val1", mg.String( "hello" ),
        "list1", mg.MustList(),
        "map1", parser.MustSymbolMap(),
        "struct1", parser.MustStruct( "ns1@v1/S2" ),
    )
    addTest := func( v mg.Value ) { b.AddTests( &ValueBuildTest{ Val: v } ) }
    addTest( mg.String( "hello" ) )
    addTest( mg.MustList() )
    addTest( mg.MustList( 1, 2, 3 ) )
    addTest( mg.MustList( 1, mg.MustList(), mg.MustList( 1, 2 ) ) )
    addTest( parser.MustSymbolMap() )
    addTest( parser.MustSymbolMap( "f1", "v1", "f2", mg.MustList(), "f3", s1 ) )
    addTest( s1 )
    addTest( parser.MustStruct( "ns1@v1/S3" ) )
    addTest( parser.MustStruct( "ns1@v1/S3", "f1", int32( 1 ) ) )
    e1 := parser.MustEnum( "ns1@v1/E1", "e" )
    addTest( e1 )
    addTest( 
        parser.MustStruct( "ns1@v1/S3",
            "f1", int32( 1 ),
            "f2", e1,
            "f3", mg.MustList( int32( 1 ), e1 ),
            "f4", parser.MustSymbolMap( "f1", e1 ),
        ),
    )
}

// we only add here error tests; we assume that a value build reactor sits
// behind a structural reactor and so let ValueBuildTest successes imply correct
// behavior of the structural check reactor for valid inputs
func initStructuralReactorTests( b *ReactorTestSetBuilder ) {
    evStartStruct1 := NewStructStartEvent( qname( "ns1@v1/S1" ) )
    id := mg.MakeTestId
    evStartField1 := NewFieldStartEvent( id( 1 ) )
    evStartField2 := NewFieldStartEvent( id( 2 ) )
    evValue1 := NewValueEvent( mg.Int64( int64( 1 ) ) )
    evListStart := nextValueListStart()
    evMapStart := nextMapStart()
    mk1 := func( 
        errMsg string, evs ...ReactorEvent ) *StructuralReactorErrorTest {
        return &StructuralReactorErrorTest{
            Events: CopySource( evs ),
            Error: NewReactorError( nil, errMsg ),
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
    b.AddTests(
        mk1( "Saw start of field 'f2' while expecting a value for field 'f1'",
            evStartStruct1, evStartField1, evStartField2,
        ),
        mk1( "Saw start of field 'f2' while expecting a value for field 'f1'",
            evStartStruct1, evStartField1, evMapStart, evStartField1,
            evStartField2,
        ),
        mk1( "Saw start of field 'f1' after value was built",
            evStartStruct1, NewEndEvent(), evStartField1,
        ),
        mk1( 
            "Expected field name or end of fields but got mingle:core@v1/Int64",
            evStartStruct1, evValue1,
        ),
        mk1( "Expected field name or end of fields but got start of mingle:core@v1/Value?*",
            evStartStruct1, evListStart,
        ),
        mk1( "Expected field name or end of fields but got mingle:core@v1/SymbolMap",
            evStartStruct1, evMapStart,
        ),
        mk1( "Expected field name or end of fields but got start of struct " +
                evStartStruct1.Type.ExternalForm(),
            evStartStruct1, evStartStruct1,
        ),
        mk1( "Saw end while expecting a value for field 'f1'",
            evStartStruct1, evStartField1, NewEndEvent(),
        ),
        mk1( "Saw start of field 'f1' while expecting a list value",
            evStartStruct1, evStartField1, evListStart, evStartField1,
        ),
        mk2( "Expected struct but got mingle:core@v1/Int64", 
            ReactorTopTypeStruct, evValue1 ),
        mk2( "Expected struct but got start of mingle:core@v1/Value?*", 
            ReactorTopTypeStruct, evListStart,
        ),
        mk2( "Expected struct but got mingle:core@v1/SymbolMap", 
            ReactorTopTypeStruct, evMapStart,
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
    addListFail := func( expct, saw string, evs ...ReactorEvent ) {
        msg := fmt.Sprintf( "expected list value of type %s but saw %s", 
            typeRef( expct ), saw )
        b.AddTests( mk1( msg, evs... ) )
    }
    for _, s := range []struct { expctTyp, saw string; ev ReactorEvent } {
        { expctTyp: "Int32", 
          saw: typeRef( "Int64" ).ExternalForm(),
          ev: NewValueEvent( mg.Int64( int64( 1 ) ) ),
        },
        { expctTyp: "Int32", 
          saw: "start of mingle:core@v1/Int32*",
          ev: nextListStart( listTypeRef( "Int32*" ) ),
        },
        { expctTyp: "&Int32",
          saw: typeRef( "Int32" ).ExternalForm(),
          ev: NewValueEvent( mg.Int32( int32( 1 ) ) ),
        },
        { expctTyp: "SymbolMap",
          saw: typeRef( "Int32" ).ExternalForm(),
          ev: NewValueEvent( mg.Int32( int32( 1 ) ) ),
        },
    } {
        lt := listTypeRef( s.expctTyp + "*" )
        lse := nextListStart( lt )
        addListFail( s.expctTyp, s.saw, lse, s.ev )
    }
    // check that we're correctly handling errors in a nested list
    b.AddTests(
        mk1(
            "expected list value of type mingle:core@v1/Int32 but saw mingle:core@v1/Int64",
            nextListStart( listTypeRef( "Int32**" ) ),
            nextListStart( listTypeRef( "Int32*" ) ),
            NewValueEvent( mg.Int32( 1 ) ),
            NewValueEvent( mg.Int64( 2 ) ),
        ),
    )
}

func initEventPathTests( b *ReactorTestSetBuilder ) {
    p := mg.MakeTestIdPath
    ee := func( ev ReactorEvent, p objpath.PathNode ) EventExpectation {
        return EventExpectation{ Event: ev, Path: p }
    }
    evStartStruct1 := NewStructStartEvent( qname( "ns1@v1/S1" ) )
    id := mg.MakeTestId
    evStartField := func( i int ) *FieldStartEvent {
        return NewFieldStartEvent( id( i ) )
    }
    evValue := func( i int64 ) *ValueEvent {
        return NewValueEvent( mg.Int64( i ) )
    }
    evEnd := NewEndEvent()
    addTest := func( name string, evs ...EventExpectation ) {
        b.AddTests( &EventPathTest{ Name: name, Events: evs } )
    }
    addTest( "empty" )
    addTest( "top-value", ee( evValue( 1 ), nil ) )
    addTest( "empty-struct",
        ee( evStartStruct1, nil ),
        ee( evEnd, nil ),
    )
    addTest( "empty-map",
        ee( nextMapStart(), nil ),
        ee( evStartField( 1 ), p( 1 ) ),
            ee( evValue( 1 ), p( 1 ) ),
        ee( evEnd, nil ),
    )
    addTest( "flat-struct",
        ee( evStartStruct1, nil ),
        ee( evStartField( 1 ), p( 1 ) ),
            ee( evValue( 1 ), p( 1 ) ),
        ee( evStartField( 2 ), p( 2 ) ),
            ee( evValue( 2 ), p( 2 ) ),
        ee( evEnd, nil ),
    )
    addTest( "empty-list",
        ee( nextValueListStart(), nil ),
        ee( NewEndEvent(), nil ),
    )
    addTest( "flat-list",
        ee( nextValueListStart(), nil ),
            ee( evValue( 1 ), p( "0" ) ),
            ee( evValue( 1 ), p( "1" ) ),
            ee( evValue( 2 ), p( "2" ) ),
            ee( evValue( 3 ), p( "3" ) ),
        ee( NewEndEvent(), nil ),
    )
    addTest( "nested-list1",
        ee( nextValueListStart(), nil ),
            ee( nextMapStart(), p( "0" ) ),
                ee( evStartField( 1 ), p( "0", 1 ) ),
                ee( evValue( 1 ), p( "0", 1 ) ),
                ee( NewEndEvent(), p( "0" ) ),
            ee( nextValueListStart(), p( "1" ) ),
                ee( evValue( 1 ), p( "1", "0" ) ),
                ee( NewEndEvent(), p( "1" ) ),
            ee( nextMapStart(), p( "2" ) ),
                ee( evStartField( 1 ), p( "2", 1 ) ),
                ee( evValue( 1 ), p( "2", 1 ) ),
                ee( NewEndEvent(), p( "2" ) ),
            ee( nextListStart( listTypeRef( "Int64*" ) ), p( "3" ) ),
                ee( evValue( 1 ), p( "3", "0" ) ),
                ee( NewEndEvent(), p( "3" ) ),
            ee( evValue( 4 ), p( "4" ) ),
        ee( NewEndEvent(), nil ),
    )
    addTest( "nested-list2",
        ee( nextValueListStart(), nil ),
            ee( nextMapStart(), p( "0" ) ),
            ee( NewEndEvent(), p( "0" ) ),
            ee( evValue( 1 ), p( "1" ) ),
        ee( NewEndEvent(), nil ),
    )
    addTest( "nested-list3",
        ee( nextValueListStart(), nil ),
            ee( evValue( 1 ), p( "0" ) ),
            ee( nextMapStart(), p( "1" ) ),
                ee( evStartField( 1 ), p( "1", 1 ) ),
                    ee( evValue( 1 ), p( "1", 1 ) ),
                ee( NewEndEvent(), p( "1" ) ),
            ee( nextValueListStart(), p( "2" ) ),
                ee( evValue( 1 ), p( "2", "0" ) ),
                ee( nextValueListStart(), p( "2", "1" ) ),
                    ee( evValue( 1 ), p( "2", "1", "0" ) ),
                    ee( evValue( 2 ), p( "2", "1", "1" ) ),
                ee( NewEndEvent(), p( "2", "1" ) ),
                ee( evValue( 3 ), p( "2", "2" ) ),
            ee( NewEndEvent(), p( "2" ) ),
        ee( NewEndEvent(), nil ),
    )
    addTest( "list-regress1",
        ee( nextValueListStart(), nil ),
            ee( nextValueListStart(), p( "0" ) ),
            ee( NewEndEvent(), p( "0" ) ),
            ee( evValue( 1 ), p( "1" ) ),
            ee( evValue( 1 ), p( "2" ) ),
        ee( NewEndEvent(), nil ),
    )
    addTest( "flat-map",
        ee( nextMapStart(), nil ),
        ee( evStartField( 1 ), p( 1 ) ),
            ee( evValue( 1 ), p( 1 ) ),
        ee( NewEndEvent(), nil ),
    )
    addTest( "struct-with-containers1",
        ee( evStartStruct1, nil ),
        ee( evStartField( 1 ), p( 1 ) ),
            ee( nextValueListStart(), p( 1 ) ),
                ee( evValue( 1 ), p( 1, "0" ) ),
                ee( evValue( 1 ), p( 1, "1" ) ),
            ee( NewEndEvent(), p( 1 ) ),
        ee( evStartField( 2 ), p( 2 ) ),
            ee( evValue( 1 ), p( 2 ) ),
        ee( evStartField( 3 ), p( 3 ) ),
            ee( nextValueListStart(), p( 3 ) ),
                ee( evValue( 0 ), p( 3, "0" ) ),
                ee( evValue( 0 ), p( 3, "1" ) ),
            ee( NewEndEvent(), p( 3 ) ),
        ee( NewEndEvent(), nil ),
    )
    addTest( "struct-with-containers2",
        ee( evStartStruct1, nil ),
        ee( evStartField( 1 ), p( 1 ) ),
            ee( nextMapStart(), p( 1 ) ),
            ee( evStartField( 2 ), p( 1, 2 ) ),
                ee( nextValueListStart(), p( 1, 2 ) ),
                    ee( evValue( 1 ), p( 1, 2, "0" ) ),
                    ee( evValue( 1 ), p( 1, 2, "1" ) ),
                    ee( nextValueListStart(), p( 1, 2, "2" ) ),
                        ee( evValue( 1 ), p( 1, 2, "2", "0" ) ),
                        ee( nextMapStart(), p( 1, 2, "2", "1" ) ),
                        ee( evStartField( 1 ), p( 1, 2, "2", "1", 1 ) ),
                            ee( evValue( 1 ), p( 1, 2, "2", "1", 1 ) ),
                        ee( evStartField( 2 ), p( 1, 2, "2", "1", 2 ) ),
                            ee( evValue( 2 ), p( 1, 2, "2", "1", 2 ) ),
                        ee( NewEndEvent(), p( 1, 2, "2", "1" ) ),
                    ee( NewEndEvent(), p( 1, 2, "2" ) ),
                ee( NewEndEvent(), p( 1, 2 ) ),
            ee( NewEndEvent(), p( 1 ) ),
        ee( NewEndEvent(), nil ),
    )
    b.AddTests(
        &EventPathTest{
            Name: "non-empty-dict-start-path",
            Events: []EventExpectation{
                { nextMapStart(), p( 2 ) },
                { evStartField( 1 ), p( 2, 1 ) },
                { evValue( 1 ), p( 2, 1 ) },
                { NewEndEvent(), p( 2 ) },
            },
            StartPath: p( 2 ),
        },
        &EventPathTest{
            Name: "non-empty-list-start-path",
            Events: []EventExpectation{ 
                { nextMapStart(), p( 2, "3" ) },
                { evStartField( 1 ), p( 2, "3", 1 ) },
                { evValue( 1 ), p( 2, "3", 1 ) },
                { NewEndEvent(), p( 2, "3" ) },
            },
            StartPath: p( 2, "3" ),
        },
    )
}

func testOrderWithIds( 
    typ *mg.QualifiedTypeName, 
    ids ...*mg.Identifier ) FieldOrderReactorTestOrder {

    ord := make( []FieldOrderSpecification, len( ids ) )
    for i, id := range ids { 
        ord[ i ] = FieldOrderSpecification{ Field: id, Required: false }
    }
    return FieldOrderReactorTestOrder{ Type: typ, Order: ord }
}

func initFieldOrderValueTests( b *ReactorTestSetBuilder ) {
    flds := make( []ReactorEvent, 5 )
    ids := make( []*mg.Identifier, len( flds ) )
    for i := 0; i < len( flds ); i++ {
        ids[ i ] = mg.MakeTestId( i )
        flds[ i ] = NewFieldStartEvent( ids[ i ] )
    }
    i1 := mg.Int32( int32( 1 ) )
    val1 := NewValueEvent( i1 )
    t1, t2 := qname( "ns1@v1/S1" ), qname( "ns1@v1/S2" )
    ss1, ss2 := NewStructStartEvent( t1 ), NewStructStartEvent( t2 )
    ss2Val1 := parser.MustStruct( t2, ids[ 0 ], i1 )
    // expct sequences for instance of ns1@v1/S1 by field f0 ...
    fldVals := []mg.Value{
        i1,
        parser.MustSymbolMap( ids[ 0 ], i1, ids[ 1 ], ss2Val1 ),
        mg.MustList( i1, i1 ),
        ss2Val1,
        i1,
    }
    mkExpct := func( ord ...int ) *mg.Struct {
        pairs := []interface{}{}
        for _, fldNum := range ord {
            pairs = append( pairs, ids[ fldNum ], fldVals[ fldNum ] )
        }
        return parser.MustStruct( t1, pairs... )
    }
    // val sequences for fields f0 ...
    fldEvs := [][]ReactorEvent {
        []ReactorEvent{ val1 },
        []ReactorEvent{
            nextMapStart(), 
                flds[ 0 ], val1, 
                flds[ 1 ], ss2, flds[ 0 ], val1, NewEndEvent(),
            NewEndEvent(),
        },
        []ReactorEvent{ nextValueListStart(), val1, val1, NewEndEvent() },
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
    addTest1 := func( src []ReactorEvent, expct mg.Value ) {
        b.AddTests(
            &FieldOrderReactorTest{ 
                Source: CopySource( src ), 
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
    b.AddTests(
        &FieldOrderReactorTest{
            Source: CopySource( 
                []ReactorEvent{
                    ss1, 
                        flds[ 0 ], val1,
                        flds[ 1 ], ss1,
                            flds[ 2 ], 
                                nextValueListStart(), 
                                    val1, NewEndEvent(),
                            flds[ 1 ], val1,
                        NewEndEvent(),
                    NewEndEvent(),
                },
            ),
            Orders: []FieldOrderReactorTestOrder{
                testOrderWithIds( t1, ids[ 1 ], ids[ 0 ], ids[ 2 ] ),
            },
            Expect: parser.MustStruct( t1,
                ids[ 0 ], i1,
                ids[ 1 ], parser.MustStruct( t1,
                    ids[ 2 ], mg.MustList( i1 ),
                    ids[ 1 ], i1,
                ),
            ),
        },
    )
    // Test generic un-field-ordered values at the top-level as well
    for i := 0; i < 4; i++ { addTest1( fldEvs[ i ], fldVals[ i ] ) }
    // Test arbitrary values with no orders in play
    addTest2 := func( expct mg.Value, src ...ReactorEvent ) {
        b.AddTests(
            &FieldOrderReactorTest{
                Source: CopySource( src ),
                Expect: expct,
                Orders: []FieldOrderReactorTestOrder{},
            },
        )
    }
    addTest2( i1, val1 )
    addTest2( mg.MustList(), nextValueListStart(), NewEndEvent() )
    addTest2( mg.MustList( i1 ), nextValueListStart(), val1, NewEndEvent() )
    addTest2( parser.MustSymbolMap(), nextMapStart(), NewEndEvent() )
    addTest2( 
        parser.MustSymbolMap( ids[ 0 ], i1 ), 
        nextMapStart(), flds[ 0 ], val1, NewEndEvent(),
    )
    addTest2( parser.MustStruct( ss1.Type ), ss1, NewEndEvent() )
    addTest2( 
        parser.MustStruct( ss1.Type, ids[ 0 ], i1 ),
        ss1, flds[ 0 ], val1, NewEndEvent(),
    )
}

func initFieldOrderMissingFieldTests( b *ReactorTestSetBuilder ) {
    fldId := mg.MakeTestId
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
            evs = append( evs, NewValueEvent( mg.Int32( fld ) ) )
        }
        return flattenEvs( append( evs, NewEndEvent() ) )
    }
    mkVal := func( flds []int ) *mg.Struct {
        pairs := make( []interface{}, 0, 2 * len( flds ) )
        for _, fld := range flds {
            pairs = append( pairs, fldId( fld ), mg.Int32( fld ) )
        }
        return parser.MustStruct( t1, pairs... )
    }
    addSucc := func( flds ...int ) {
        b.AddTests(
            &FieldOrderMissingFieldsTest{
                Orders: ords,
                Source: CopySource( mkSrc( flds ) ),
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
        miss := make( []*mg.Identifier, len( missIds ) )
        for i, missId := range missIds { miss[ i ] = fldId( missId ) }
        b.AddTests(
            &FieldOrderMissingFieldsTest{
                Orders: ords,
                Source: CopySource( mkSrc( flds ) ),
                Error: mg.NewMissingFieldsError( nil, miss ),
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

func initFieldOrderPathTests( b *ReactorTestSetBuilder ) {
    mapStart := NewMapStartEvent()
    valListStart := NewListStartEvent( mg.TypeOpaqueList )
    i1 := mg.Int32( int32( 1 ) )
    val1 := NewValueEvent( i1 )
    id := mg.MakeTestId
    typ := func( i int ) *mg.QualifiedTypeName {
        return qname( fmt.Sprintf( "ns1@v1/S%d", i ) )
    }
    ss := func( i int ) *StructStartEvent { 
        return NewStructStartEvent( typ( i ) ) 
    }
    ssListStart := func( i int ) *ListStartEvent {
        lt := &mg.ListTypeReference{ 
            ElementType: typ( i ).AsAtomicType(), 
            AllowsEmpty: true,
        }
        return NewListStartEvent( lt )
    }
    fld := func( i int ) *FieldStartEvent { 
        return NewFieldStartEvent( id( i ) ) 
    }
    p := mg.MakeTestIdPath
    expct1 := []EventExpectation{
        { ss( 1 ), nil },
            { fld( 0 ), p( 0 ) },
            { val1, p( 0 ) },
            { fld( 1 ), p( 1 ) },
            { mapStart, p( 1 ) },
                { fld( 1 ), p( 1, 1 ) },
                { val1, p( 1, 1 ) },
                { fld( 0 ), p( 1, 0 ) },
                { val1, p( 1, 0 ) },
            { NewEndEvent(), p( 1 ) },
            { fld( 2 ), p( 2 ) },
            { valListStart, p( 2 ) },
                { val1, p( 2, "0" ) },
                { val1, p( 2, "1" ) },
            { NewEndEvent(), p( 2 ) },
            { fld( 3 ), p( 3 ) },
            { ss( 2 ), p( 3 ) },
                { fld( 0 ), p( 3, 0 ) },
                { val1, p( 3, 0 ) },
                { fld( 1 ), p( 3, 1 ) },
                { valListStart, p( 3, 1 ) },
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
                { mapStart, p( 4, 3 ) },
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
                { ssListStart( 3 ), p( 4, 4 ) },
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
            mapStart, 
                fld( 1 ), val1, fld( 0 ), val1, NewEndEvent() },
        []ReactorEvent{ valListStart, val1, val1, NewEndEvent() },
        []ReactorEvent{ 
            ss( 2 ), 
                fld( 0 ), val1, 
                fld( 1 ), valListStart, val1, val1, NewEndEvent(),
            NewEndEvent(),
        },
        // val for f4 is nested and has nested ss2 instances that are in varying
        // need of reordering
        []ReactorEvent{ 
            ss( 1 ),
                fld( 0 ), val1,
                fld( 4 ), ssListStart( 3 ),
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
                fld( 3 ), mapStart,
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
        b.AddTests(
            &FieldOrderPathTest{
                Source: CopySource( mkSrc( ord... ) ),
                Expect: expct1,
                Orders: ords1,
            },
        )
    }
    b.AddTests(
        &FieldOrderPathTest{
            Source: CopySource(
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
    b.AddTests(
        &FieldOrderPathTest{
            Source: CopySource(
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

func initFieldOrderReactorTests( b *ReactorTestSetBuilder ) {
    initFieldOrderValueTests( b )
    initFieldOrderMissingFieldTests( b )
    initFieldOrderPathTests( b )
}

func initReactorTests( b *ReactorTestSetBuilder ) {
    initStructuralReactorTests( b )
    initValueBuildReactorTests( b )
    initEventPathTests( b )
    initFieldOrderReactorTests( b )
}

func init() { 
    reactorTestNs = parser.MustNamespace( "mingle:reactor@v1" )
    AddTestInitializer( reactorTestNs, initReactorTests ) 
}
