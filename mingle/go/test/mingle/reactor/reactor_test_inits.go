package reactor

import (
    mg "mingle"
    "mingle/parser"
    "bitgirder/objpath"
    "fmt"
)

func initValueBuildZeroRefTests( b *ReactorTestSetBuilder ) {
    qn := parser.MustQualifiedTypeName( "ns1@v1/S1" )
    listStart := func() *ListStartEvent {
        return NewListStartEvent( mg.TypeOpaqueList, mg.PointerIdNull )
    }
    b.AddTests(
        &ValueBuildTest{
            Val: mg.MustList(
                int32( 1 ),
                parser.MustSymbolMap( 
                    "f1", int32( 1 ),
                    "f2", mg.MustList( "hello" ),
                ),
                mg.NewHeapValue( mg.Int32( int32( 1 ) ) ),
                mg.NewHeapValue( parser.MustStruct( "ns1@v1/S1" ) ),
            ),
            Source: CopySource(
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
            parser.MustStruct( "ns1@v1/S1",
                "f1", mg.Int32( 1 ),
                "f2", mg.NewHeapValue( mg.Int32( 2 ) ),
                "f3", 
                    mg.NewHeapValue( 
                        mg.MustList( mg.NewHeapValue( mg.Int32( 1 ) ) ) ),
                "f4", mg.NewHeapValue( 
                    parser.MustSymbolMap( 
                        "g1", mg.NullVal, "g2", mg.Int32( 1 ) ) ),
            ),
        ),
    )
    valPtr1 := mg.NewHeapValue( mg.Int32( 1 ) )
    addTest( mg.MustList( valPtr1, valPtr1, valPtr1 ) )
    initValueBuildZeroRefTests( b )
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
    evValuePtr1 := NewValueAllocationEvent( mg.TypeInt64, 1 )
    listStart := func( i int ) *ListStartEvent {
        return NewListStartEvent( mg.TypeOpaqueList, ptrId( i ) )
    }
    mapStart := func( i int ) *MapStartEvent {
        return NewMapStartEvent( ptrId( i ) )
    }
    valAlloc := func( i int ) *ValueAllocationEvent {
        return NewValueAllocationEvent( mg.TypeValue, ptrId( i ) )
    }
    evListStart := listStart( 1 )
    evMapStart := mapStart( 2 )
    mk1 := func( 
        errMsg string, evs ...ReactorEvent ) *StructuralReactorErrorTest {
        return &StructuralReactorErrorTest{
            Events: CopySource( evs ),
            Error: rctError( nil, errMsg ),
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
        mk1( 
            "Expected field name or end of fields but got allocation of mingle:core@v1/Int64",
            evStartStruct1, evValuePtr1,
        ),
        mk1( "Expected field name or end of fields but got reference",
            evStartStruct1, ptrRef( 2 ),
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
        mk2( "Expected struct but got allocation of mingle:core@v1/Int64", 
            ReactorTopTypeStruct, evValuePtr1 ),
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
        mk1( "reference 1 is cyclic",
            listStart( 1 ), ptrRef( 1 ),
        ),
        mk1( "reference 1 is cyclic",
            listStart( 1 ),
            NewValueEvent( mg.Int32( int32( 1 ) ) ), ptrRef( 1 ),
        ),
        mk1( "reference 1 is cyclic",
            listStart( 1 ), listStart( 2 ), ptrRef( 1 ),
        ),
        mk1( "reference 1 is cyclic",
            listStart( 1 ), mapStart( 2 ), evStartField1, ptrRef( 1 ),
        ),
        mk1( "reference 1 is cyclic", 
            mapStart( 1 ), evStartField1, ptrRef( 1 ),
        ),
        mk1( "reference 1 is cyclic",
            mapStart( 1 ), evStartField1, listStart( 2 ), ptrRef( 1 ),
        ),
        mk1( "reference 1 is cyclic",
            mapStart( 1 ), 
            evStartField1, 
            mapStart( 2 ), 
            evStartField1,
            ptrRef( 1 ),
        ),
        mk1( "reference 1 is cyclic",
            valAlloc( 1 ), ptrRef( 1 ),
        ),
        mk1( "reference 1 is cyclic",
            listStart( 2 ), valAlloc( 1 ), listStart( 2 ), ptrRef( 1 ),
        ),
        mk1( "reference 1 is cyclic",
            valAlloc( 1 ),
            evStartStruct1, evStartField1, ptrRef( 1 ),
        ),
    )
    addAllocFail := func( ev ReactorEvent, errStr string ) {
        s := fmt.Sprintf( 
            "allocation of mingle:core@v1/Int32 followed by %s", errStr )
        b.AddTests( mk1( s, ptrAlloc( mg.TypeInt32, 1 ), ev ) )
    }
    addAllocFail( NewValueEvent( mg.Int64( int64( 1 ) ) ), 
        "mingle:core@v1/Int64" )
    addAllocFail( NewValueEvent( mg.NullVal ), "mingle:core@v1/Null" )
    addAllocFail( 
        NewListStartEvent( mg.TypeOpaqueList, ptrId( 1 ) ),
        "start of mingle:core@v1/Value?*",
    )
    addAllocFail( NewMapStartEvent( ptrId( 1 ) ), "mingle:core@v1/SymbolMap" )
    addAllocFail( NewValueEvent( parser.MustStruct( "ns1@v1/S1" ) ),
        "ns1@v1/S1" )
    addAllocFail( 
        ptrAlloc( mg.TypeInt32, 2 ), 
        "allocation of mingle:core@v1/Int32" )
    b.AddTests(
        mk1( "allocation of &(mingle:core@v1/Int32) followed by allocation of &(mingle:core@v1/Int64)",
            ptrAlloc( typeRef( "&Int32" ), 1 ),
            ptrAlloc( typeRef( "&Int64" ), 2 ),
        ),
        mk1( "allocation of &(mingle:core@v1/Int32) followed by allocation of mingle:core@v1/Int64",
            ptrAlloc( typeRef( "&Int32" ), 1 ),
            ptrAlloc( typeRef( "Int64" ), 2 ),
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
          saw: "allocation of mingle:core@v1/Int32",
          ev: ptrAlloc( mg.TypeInt32, 1 ),
        },
        { expctTyp: "Int32", 
          saw: "start of mingle:core@v1/Int32*",
          ev: NewListStartEvent( listTypeRef( "Int32*" ), ptrId( 2 ) ),
        },
        { expctTyp: "&Int32",
          saw: typeRef( "Int32" ).ExternalForm(),
          ev: NewValueEvent( mg.Int32( int32( 1 ) ) ),
        },
        { expctTyp: "&Int32",
          saw: "allocation of mingle:core@v1/Int64",
          ev: ptrAlloc( mg.TypeInt64, 2 ),
        },
        { expctTyp: "SymbolMap",
          saw: typeRef( "Int32" ).ExternalForm(),
          ev: NewValueEvent( mg.Int32( int32( 1 ) ) ),
        },
    } {
        lt := listTypeRef( s.expctTyp + "*" )
        lse := NewListStartEvent( lt, ptrId( 1 ) )
        addListFail( s.expctTyp, s.saw, lse, s.ev )
    }
    // check that we're correctly handling errors in a nested list
    b.AddTests(
        mk1(
            "expected list value of type mingle:core@v1/Int32 but saw mingle:core@v1/Int64",
            NewListStartEvent( listTypeRef( "Int32**" ), ptrId( 1 ) ),
            NewListStartEvent( listTypeRef( "Int32*" ), ptrId( 2 ) ),
            NewValueEvent( mg.Int32( 1 ) ),
            NewValueEvent( mg.Int64( 2 ) ),
        ),
    )
    b.AddTests(
        mk1(
            "allocation of mingle:core@v1/SymbolMap followed by mingle:core@v1/Int32",
            ptrAlloc( mg.TypeSymbolMap, 1 ),
            NewValueEvent( mg.Int32( int32( 1 ) ) ),
        ),
        mk1(
            "allocation of ns1@v1/S1 followed by ns1@v1/S2",
            ptrAlloc( typeRef( "ns1@v1/S1" ), 1 ),
            NewValueEvent( parser.MustStruct( "ns1@v1/S2" ) ),
        ),
    )
}

func initPointerReferenceCheckTests( b *ReactorTestSetBuilder ) {
    id, p := mg.MakeTestId, mg.MakeTestIdPath
    fld := func( i int ) *FieldStartEvent {
        return NewFieldStartEvent( id( i ) )
    }
    ival := NewValueEvent( mg.Int32( 1 ) )
    qn := parser.MustQualifiedTypeName( "ns1@v1/S1" )
    add := func( path objpath.PathNode, msg string, evs ...ReactorEvent ) {
        b.AddTests(
            &PointerEventCheckTest{
                Events: CopySource( evs ),
                Error: rctError( path, msg ),
            },
        )
    }
    add0 := func( path objpath.PathNode, evs ...ReactorEvent ) {
        add( path, "attempt to reference null pointer", evs... )
    }
    listStart := func( id mg.PointerId ) *ListStartEvent {
        return NewListStartEvent( mg.TypeOpaqueList, id )
    }
    add0( p( "0" ), listStart( mg.PointerIdNull ), ptrRef( 0 ) )
    add0( p( "1" ), listStart( mg.PointerIdNull ), ival, ptrRef( 0 ) )
    add0( p( 1, 2 ),
        NewMapStartEvent( 0 ), fld( 1 ), 
            NewMapStartEvent( 0 ), fld( 2 ), ptrRef( 0 ) )
    addReallocCheck := func( errEv ReactorEvent ) {
        msg := "attempt to redefine reference: 1"
        add( p( "1" ), msg,
            listStart( ptrId( 2 ) ), 
                ptrAlloc( mg.TypeInt32, 1 ), ival, errEv,
        )
        add( p( "1" ), msg, listStart( ptrId( 1 ) ), ival, errEv )
        add( p( 1 ), msg, NewMapStartEvent( ptrId( 1 ) ), fld( 1 ), errEv )
        add( p( 2 ), msg,
            NewStructStartEvent( qn ), 
            fld( 1 ), ptrAlloc( mg.TypeInt32, 1 ), ival, 
            fld( 2 ), errEv,
        )
    }
    addReallocCheck( ptrAlloc( mg.TypeInt32, 1 ) )
    addReallocCheck( listStart( ptrId( 1 ) ) )
    addReallocCheck( NewMapStartEvent( ptrId( 1 ) ) )
    add( p( "0" ), "unrecognized reference: 1",
        listStart( ptrId( 2 ) ), ptrRef( 1 ) )
    add( p( 2 ),
        "unrecognized reference: 2",
        NewMapStartEvent( ptrId( 3 ) ),
        fld( 1 ),
        ptrAlloc( mg.TypeInt32, 1 ), ival,
        fld( 2 ),
        ptrRef( 2 ),
    )
    b.AddTests(
        &PointerEventCheckTest{
            Events: CopySource( 
                []ReactorEvent{
                    NewStructStartEvent( qn ),
                    fld( 1 ), ptrAlloc( mg.TypeInt32, 0 ), ival,
                    fld( 2 ), listStart( mg.PointerIdNull ), NewEndEvent(),
                    fld( 3 ), 
                        NewMapStartEvent( mg.PointerIdNull ), NewEndEvent(),
                    NewEndEvent(),
                },
            ),
        },
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
    idFact := NewTestPointerIdFactory()
    evEnd := NewEndEvent()
    addTest := func( name string, evs ...EventExpectation ) {
        ptrStart := ee( idFact.NextValueAllocation( mg.TypeValue ), nil )
        evsWithPtr := append( []EventExpectation{ ptrStart }, evs... )
        b.AddTests(
            &EventPathTest{ Name: name, Events: evs },
            &EventPathTest{ Name: name + "-pointer", Events: evsWithPtr },
        )
    }
    addTest( "empty" )
    addTest( "top-value", ee( evValue( 1 ), nil ) )
    addTest( "empty-struct",
        ee( evStartStruct1, nil ),
        ee( evEnd, nil ),
    )
    addTest( "empty-map",
        ee( idFact.NextMapStart(), nil ),
        ee( evStartField( 1 ), p( 1 ) ),
            ee( evValue( 1 ), p( 1 ) ),
        ee( evEnd, nil ),
    )
    addTest( "flat-struct",
        ee( evStartStruct1, nil ),
        ee( evStartField( 1 ), p( 1 ) ),
            ee( evValue( 1 ), p( 1 ) ),
        ee( evStartField( 2 ), p( 2 ) ),
            ee( idFact.NextValueAllocation( mg.TypeInt64 ), p( 2 ) ),
                ee( evValue( 2 ), p( 2 ) ),
        ee( evEnd, nil ),
    )
    addTest( "empty-list",
        ee( idFact.NextValueListStart(), nil ),
        ee( NewEndEvent(), nil ),
    )
    addTest( "flat-list",
        ee( idFact.NextValueListStart(), nil ),
            ee( evValue( 1 ), p( "0" ) ),
            ee( evValue( 1 ), p( "1" ) ),
            ee( idFact.NextValueAllocation( mg.TypeInt64 ), p( "2" ) ),
                ee( evValue( 2 ), p( "2" ) ),
            ee( ptrRef( 2 ), p( "3" ) ),
            ee( evValue( 4 ), p( "4" ) ),
        ee( NewEndEvent(), nil ),
    )
    addTest( "nested-list1",
        ee( idFact.NextValueListStart(), nil ),
            ee( idFact.NextMapStart(), p( "0" ) ),
                ee( evStartField( 1 ), p( "0", 1 ) ),
                ee( evValue( 1 ), p( "0", 1 ) ),
                ee( NewEndEvent(), p( "0" ) ),
            ee( idFact.NextValueListStart(), p( "1" ) ),
                ee( evValue( 1 ), p( "1", "0" ) ),
                ee( NewEndEvent(), p( "1" ) ),
            ee( idFact.NextValueAllocation( mg.TypeSymbolMap ), p( "2" ) ),
                ee( idFact.NextMapStart(), p( "2" ) ),
                    ee( evStartField( 1 ), p( "2", 1 ) ),
                    ee( evValue( 1 ), p( "2", 1 ) ),
                    ee( NewEndEvent(), p( "2" ) ),
            ee( idFact.NextValueAllocation( typeRef( "Int64*" ) ), p( "3" ) ),
                ee( idFact.NextListStart( listTypeRef( "Int64*" ) ), p( "3" ) ),
                    ee( evValue( 1 ), p( "3", "0" ) ),
                    ee( NewEndEvent(), p( "3" ) ),
            ee( evValue( 4 ), p( "4" ) ),
        ee( NewEndEvent(), nil ),
    )
    addTest( "nested-list2",
        ee( idFact.NextValueListStart(), nil ),
            ee( idFact.NextMapStart(), p( "0" ) ),
            ee( NewEndEvent(), p( "0" ) ),
            ee( evValue( 1 ), p( "1" ) ),
        ee( NewEndEvent(), nil ),
    )
    addTest( "nested-list3",
        ee( idFact.NextValueListStart(), nil ),
            ee( evValue( 1 ), p( "0" ) ),
            ee( idFact.NextMapStart(), p( "1" ) ),
                ee( evStartField( 1 ), p( "1", 1 ) ),
                    ee( evValue( 1 ), p( "1", 1 ) ),
                ee( NewEndEvent(), p( "1" ) ),
            ee( idFact.NextValueAllocation( mg.TypeOpaqueList ), p( "2" ) ),
                ee( idFact.NextValueListStart(), p( "2" ) ),
                    ee( evValue( 1 ), p( "2", "0" ) ),
                    ee( idFact.NextValueAllocation( mg.TypeOpaqueList ), 
                            p( "2", "1" ) ),
                        ee( idFact.NextValueListStart(), p( "2", "1" ) ),
                            ee( evValue( 1 ), p( "2", "1", "0" ) ),
                            ee( evValue( 2 ), p( "2", "1", "1" ) ),
                        ee( NewEndEvent(), p( "2", "1" ) ),
                    ee( evValue( 3 ), p( "2", "2" ) ),
                ee( NewEndEvent(), p( "2" ) ),
        ee( NewEndEvent(), nil ),
    )
    addTest( "list-regress1",
        ee( idFact.NextValueListStart(), nil ),
            ee( idFact.NextValueListStart(), p( "0" ) ),
            ee( NewEndEvent(), p( "0" ) ),
            ee( evValue( 1 ), p( "1" ) ),
            ee( evValue( 1 ), p( "2" ) ),
        ee( NewEndEvent(), nil ),
    )
    addTest( "flat-map",
        ee( idFact.NextMapStart(), nil ),
        ee( evStartField( 1 ), p( 1 ) ),
            ee( evValue( 1 ), p( 1 ) ),
        ee( NewEndEvent(), nil ),
    )
    addTest( "struct-with-containers1",
        ee( evStartStruct1, nil ),
        ee( evStartField( 1 ), p( 1 ) ),
            ee( idFact.NextValueListStart(), p( 1 ) ),
                ee( evValue( 1 ), p( 1, "0" ) ),
                ee( evValue( 1 ), p( 1, "1" ) ),
            ee( NewEndEvent(), p( 1 ) ),
        ee( evStartField( 2 ), p( 2 ) ),
            ee( idFact.NextValueAllocation( mg.TypeInt64 ), p( 2 ) ),
                ee( evValue( 1 ), p( 2 ) ),
        ee( evStartField( 3 ), p( 3 ) ),
            ee( idFact.NextValueListStart(), p( 3 ) ),
                ee( idFact.NextValueAllocation( mg.TypeInt64 ), p( 3, "0" ) ),
                    ee( evValue( 0 ), p( 3, "0" ) ),
                ee( idFact.NextValueAllocation( mg.TypeInt64 ), p( 3, "1" ) ),
                    ee( evValue( 0 ), p( 3, "1" ) ),
            ee( NewEndEvent(), p( 3 ) ),
        ee( NewEndEvent(), nil ),
    )
    addTest( "struct-with-containers2",
        ee( evStartStruct1, nil ),
        ee( evStartField( 1 ), p( 1 ) ),
            ee( idFact.NextMapStart(), p( 1 ) ),
            ee( evStartField( 2 ), p( 1, 2 ) ),
                ee( idFact.NextValueListStart(), p( 1, 2 ) ),
                    ee( evValue( 1 ), p( 1, 2, "0" ) ),
                    ee( evValue( 1 ), p( 1, 2, "1" ) ),
                    ee( idFact.NextValueListStart(), p( 1, 2, "2" ) ),
                        ee( evValue( 1 ), p( 1, 2, "2", "0" ) ),
                        ee( idFact.NextMapStart(), p( 1, 2, "2", "1" ) ),
                        ee( evStartField( 1 ), p( 1, 2, "2", "1", 1 ) ),
                            ee( evValue( 1 ), p( 1, 2, "2", "1", 1 ) ),
                        ee( evStartField( 2 ), p( 1, 2, "2", "1", 2 ) ),
                            ee( idFact.NextValueAllocation( mg.TypeInt64 ), 
                                p( 1, 2, "2", "1", 2 ) ),
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
                { idFact.NextMapStart(), p( 2 ) },
                { evStartField( 1 ), p( 2, 1 ) },
                { evValue( 1 ), p( 2, 1 ) },
                { NewEndEvent(), p( 2 ) },
            },
            StartPath: p( 2 ),
        },
        &EventPathTest{
            Name: "non-empty-list-start-path",
            Events: []EventExpectation{ 
                { idFact.NextMapStart(), p( 2, "3" ) },
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
    idFact := NewTestPointerIdFactory()
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
            idFact.NextMapStart(), 
                flds[ 0 ], val1, 
                flds[ 1 ], ss2, flds[ 0 ], val1, NewEndEvent(),
            NewEndEvent(),
        },
        []ReactorEvent{ 
            idFact.NextValueListStart(), val1, val1, NewEndEvent() },
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
                                idFact.NextValueListStart(), 
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
    addTest2( mg.MustList(), idFact.NextValueListStart(), NewEndEvent() )
    addTest2( 
        mg.MustList( i1 ), idFact.NextValueListStart(), val1, NewEndEvent() )
    addTest2( parser.MustSymbolMap(), idFact.NextMapStart(), NewEndEvent() )
    addTest2( 
        parser.MustSymbolMap( ids[ 0 ], i1 ), 
        idFact.NextMapStart(), flds[ 0 ], val1, NewEndEvent(),
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
    mapStart := NewMapStartEvent( mg.PointerIdNull )
    valListStart := NewListStartEvent( mg.TypeOpaqueList, mg.PointerIdNull )
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
        return NewListStartEvent( lt, mg.PointerIdNull )
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

//func initRequestTests( b *ReactorTestSetBuilder ) {
//    id := parser.MustIdentifier
//    idFact := NewTestPointerIdFactory()
//    ns1 := parser.MustNamespace( "ns1@v1" )
//    svc1 := id( "service1" )
//    op1 := id( "op1" )
//    params1 := parser.MustSymbolMap( "f1", int32( 1 ) )
//    authQn := qname( "ns1@v1/Auth1" )
//    auth1 := parser.MustStruct( authQn, "f1", int32( 1 ) )
//    evFldNs := NewFieldStartEvent( mg.IdNamespace )
//    evFldSvc := NewFieldStartEvent( mg.IdService )
//    evFldOp := NewFieldStartEvent( mg.IdOperation )
//    evFldParams := NewFieldStartEvent( mg.IdParameters )
//    evFldAuth := NewFieldStartEvent( mg.IdAuthentication )
//    evFldF1 := NewFieldStartEvent( id( "f1" ) )
//    evReqTyp := NewStructStartEvent( mg.QnameRequest )
//    evNs1 := NewValueEvent( mg.String( ns1.ExternalForm() ) )
//    evSvc1 := NewValueEvent( mg.String( svc1.ExternalForm() ) )
//    evOp1 := NewValueEvent( mg.String( op1.ExternalForm() ) )
//    i32Val1 := NewValueEvent( mg.Int32( 1 ) )
//    evParams1 := []ReactorEvent{ 
//        idFact.NextMapStart(), evFldF1, i32Val1, NewEndEvent() }
//    evAuth1 := []ReactorEvent{ 
//        NewStructStartEvent( authQn ), evFldF1, i32Val1, NewEndEvent() }
//    addSucc1 := func( evs ...interface{} ) {
//        b.AddTests(
//            &RequestReactorTest{
//                Source: CopySource( flattenEvs( evs... ) ),
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
//    b.AddTests(
//        &RequestReactorTest{
//            Source: CopySource(
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
//            mg.IdNamespace, mg.NamespaceAsBytes( ns1 ),
//            mg.IdService, mg.IdentifierAsBytes( svc1 ),
//            mg.IdOperation, mg.IdentifierAsBytes( op1 ),
//        }
//        if params != nil { pairs = append( pairs, mg.IdParameters, params ) }
//        if auth != nil { pairs = append( pairs, mg.IdAuthentication, auth ) }
//        return parser.MustStruct( mg.QnameRequest, pairs... )
//    }
//    addSucc2 := func( src interface{}, authExpct mg.Value ) {
//        b.AddTests(
//            &RequestReactorTest{
//                Namespace: ns1,
//                Service: svc1,
//                Operation: op1,
//                Parameters: mg.EmptySymbolMap(),
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
//            mg.IdNamespace, ns1.ExternalForm(),
//            mg.IdService, svc1.ExternalForm(),
//            mg.IdOperation, op1.ExternalForm(),
//        }
//        if paramsIn != nil { 
//            pairs = append( pairs, mg.IdParameters, paramsIn ) 
//        }
//        if auth != nil { pairs = append( pairs, mg.IdAuthentication, auth ) }
//        t.Source = parser.MustStruct( mg.QnameRequest, pairs... )
//        b.AddTests( t )
//    }
//    pathParams := objpath.RootedAt( mg.IdParameters )
//    evsEmptyParams := []EventExpectation{ 
//        { idFact.NextMapStart(), pathParams }, { NewEndEvent(), pathParams } }
//    pathAuth := objpath.RootedAt( mg.IdAuthentication )
//    addPathSucc( nil, parser.MustSymbolMap(), evsEmptyParams, nil, nil )
//    addPathSucc( 
//        parser.MustSymbolMap(), 
//        parser.MustSymbolMap(), 
//        evsEmptyParams, 
//        nil, 
//        nil,
//    )
//    idF1 := id( "f1" )
//    addPathSucc(
//        parser.MustSymbolMap( idF1, mg.Int32( 1 ) ),
//        parser.MustSymbolMap( idF1, mg.Int32( 1 ) ),
//        []EventExpectation{
//            { idFact.NextMapStart(), pathParams },
//            { evFldF1, pathParams.Descend( idF1 ) },
//            { i32Val1, pathParams.Descend( idF1 ) },
//            { NewEndEvent(), pathParams },
//        },
//        nil, nil,
//    )
//    addPathSucc( 
//        nil, parser.MustSymbolMap(), evsEmptyParams,
//        mg.Int32( 1 ), []EventExpectation{ { i32Val1, pathAuth } },
//    )
//    addPathSucc(
//        nil, parser.MustSymbolMap(), evsEmptyParams,
//        auth1, []EventExpectation{
//            { NewStructStartEvent( authQn ), pathAuth },
//            { evFldF1, pathAuth.Descend( idF1 ) },
//            { i32Val1, pathAuth.Descend( idF1 ) },
//            { NewEndEvent(), pathAuth },
//        },
//    )
//    writeMgIo := func( f func( w *mg.BinWriter ) ) mg.Buffer {
//        bb := &bytes.Buffer{}
//        w := mg.NewWriter( bb )
//        f( w )
//        return mg.Buffer( bb.Bytes() )
//    }
//    nsBuf := func( ns *mg.Namespace ) mg.Buffer {
//        return writeMgIo( func( w *mg.BinWriter ) { w.WriteNamespace( ns ) } )
//    }
//    idBuf := func( id *mg.Identifier ) mg.Buffer {
//        return writeMgIo( func( w *mg.BinWriter ) { w.WriteIdentifier( id ) } )
//    }
//    b.AddTests(
//        &RequestReactorTest{
//            Namespace: ns1,
//            Service: svc1,
//            Operation: op1,
//            Parameters: mg.EmptySymbolMap(),
//            Source: parser.MustStruct( mg.QnameRequest,
//                mg.IdNamespace, nsBuf( ns1 ),
//                mg.IdService, idBuf( svc1 ),
//                mg.IdOperation, idBuf( op1 ),
//            ),
//        },
//    )
//    b.AddTests(
//        &RequestReactorTest{
//            Source: parser.MustStruct( "ns1@v1/S1" ),
//            Error: mg.NewTypeCastError(
//                mg.TypeRequest, typeRef( "ns1@v1/S1" ), nil ),
//        },
//    )
//    createReqVcErr := func( 
//        val interface{}, 
//        path objpath.PathNode, 
//        msg string ) *RequestReactorTest {
//
//        return &RequestReactorTest{
//            Source: mg.MustValue( val ),
//            Error: mg.NewValueCastError( path, msg ),
//        }
//    }
//    addReqVcErr := func( val interface{}, path objpath.PathNode, msg string ) {
//        b.AddTests( createReqVcErr( val, path, msg ) )
//    }
//    addReqVcErr(
//        parser.MustSymbolMap( mg.IdNamespace, true ), 
//        objpath.RootedAt( mg.IdNamespace ),
//        "invalid value: mingle:core@v1/Boolean",
//    )
//    addReqVcErr(
//        parser.MustSymbolMap( mg.IdNamespace, parser.MustSymbolMap() ),
//        objpath.RootedAt( mg.IdNamespace ),
//        "invalid value: mingle:core@v1/SymbolMap",
//    )
//    addReqVcErr(
//        parser.MustSymbolMap( 
//            mg.IdNamespace, parser.MustStruct( "ns1@v1/S1" ) ),
//        objpath.RootedAt( mg.IdNamespace ),
//        "invalid value: ns1@v1/S1",
//    )
//    addReqVcErr(
//        parser.MustSymbolMap( mg.IdNamespace, mg.MustList() ),
//        objpath.RootedAt( mg.IdNamespace ),
//        "invalid value: mingle:core@v1/Value?*",
//    )
//    func() {
//        test := createReqVcErr(
//            parser.MustSymbolMap( 
//                mg.IdNamespace, ns1.ExternalForm(), mg.IdService, true ),
//            objpath.RootedAt( mg.IdService ),
//            "invalid value: mingle:core@v1/Boolean",
//        )
//        test.Namespace = ns1
//        b.AddTests( test )
//    }()
//    func() {
//        test := createReqVcErr(
//            parser.MustSymbolMap( 
//                mg.IdNamespace, ns1.ExternalForm(),
//                mg.IdService, svc1.ExternalForm(),
//                mg.IdOperation, true,
//            ),
//            objpath.RootedAt( mg.IdOperation ),
//            "invalid value: mingle:core@v1/Boolean",
//        )
//        test.Namespace, test.Service = ns1, svc1
//        b.AddTests( test )
//    }()
//    b.AddTests(
//        &RequestReactorTest{
//            Source: parser.MustSymbolMap(
//                mg.IdNamespace, ns1.ExternalForm(),
//                mg.IdService, svc1.ExternalForm(),
//                mg.IdOperation, op1.ExternalForm(),
//                mg.IdParameters, true,
//            ),
//            Namespace: ns1,
//            Service: svc1,
//            Operation: op1,
//            Error: mg.NewTypeCastError(
//                mg.TypeSymbolMap,
//                mg.TypeBoolean,
//                objpath.RootedAt( mg.IdParameters ),
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
//            parser.MustSymbolMap( pairs... ), objpath.RootedAt( path ), msg )
//    }
//    addBinRdErr := func( 
//        path *mg.Identifier, msg string, pairs ...interface{} ) {
//
//        b.AddTests( createBinRdErr( path, msg, pairs... ) )
//    }
//    badBuf := []byte{ 0x0f }
//    addBinRdErr( 
//        mg.IdNamespace, 
//        "[offset 0]: Expected type code 0x02 but got 0x0f",
//        mg.IdNamespace, badBuf )
//    func() {
//        test := createBinRdErr(
//            mg.IdService, 
//            "[offset 0]: Expected type code 0x01 but got 0x0f",
//            mg.IdNamespace, ns1.ExternalForm(), 
//            mg.IdService, badBuf,
//        )
//        test.Namespace = ns1
//        b.AddTests( test )
//    }()
//    func() {
//        test := createBinRdErr(
//            mg.IdOperation, 
//            "[offset 0]: Expected type code 0x01 but got 0x0f",
//            mg.IdNamespace, ns1.ExternalForm(),
//            mg.IdService, svc1.ExternalForm(),
//            mg.IdOperation, badBuf,
//        )
//        test.Namespace, test.Service = ns1, svc1
//        b.AddTests( test )
//    }()
//    addReqVcErr(
//        parser.MustSymbolMap( mg.IdNamespace, "ns1::ns2" ),
//        objpath.RootedAt( mg.IdNamespace ),
//        "[<input>, line 1, col 5]: Illegal start of identifier part: \":\" " +
//        "(U+003A)",
//    )
//    func() {
//        test := createReqVcErr(
//            parser.MustSymbolMap( 
//                mg.IdNamespace, ns1.ExternalForm(), mg.IdService, "2bad" ),
//            objpath.RootedAt( mg.IdService ),
//            "[<input>, line 1, col 1]: Illegal start of identifier part: " +
//            "\"2\" (U+0032)",
//        )
//        test.Namespace = ns1
//        b.AddTests( test )
//    }()
//    func() {
//        test := createReqVcErr(
//            parser.MustSymbolMap(
//                mg.IdNamespace, ns1.ExternalForm(),
//                mg.IdService, svc1.ExternalForm(),
//                mg.IdOperation, "2bad",
//            ),
//            objpath.RootedAt( mg.IdOperation ),
//            "[<input>, line 1, col 1]: Illegal start of identifier part: " +
//            "\"2\" (U+0032)",
//        )
//        test.Namespace, test.Service = ns1, svc1
//        b.AddTests( test )
//    }()
//    t1Bad := qname( "foo@v1/Request" )
//    b.AddTests(
//        &RequestReactorTest{
//            Source: parser.MustStruct( t1Bad ),
//            Error: mg.NewTypeCastError(
//                mg.TypeRequest, t1Bad.AsAtomicType(), nil ),
//        },
//    )
//    // Not exhaustively re-testing all ways a field could be missing (assume for
//    // now that field order tests will handle that). Instead, we are just
//    // getting basic coverage that the field order supplied by the request
//    // reactor is in fact being set up correctly and that we have set up the
//    // right required fields.
//    b.AddTests(
//        &RequestReactorTest{
//            Source: parser.MustSymbolMap( 
//                mg.IdNamespace, ns1.ExternalForm(),
//                mg.IdOperation, op1.ExternalForm(),
//            ),
//            Namespace: ns1,
//            Error: mg.NewMissingFieldsError( 
//                nil, []*mg.Identifier{ mg.IdService } ),
//        },
//    )
//}
//
//func initResponseTests( b *ReactorTestSetBuilder ) {
//    id := mg.MakeTestId
//    idFact := NewTestPointerIdFactory()
//    addSucc := func( in, res, err mg.Value ) {
//        b.AddTests(
//            &ResponseReactorTest{ In: in, ResVal: res, ErrVal: err } )
//    }
//    i32Val1 := mg.Int32( 1 )
//    err1 := parser.MustStruct( "ns1@v1/Err1", "f1", int32( 1 ) )
//    addSucc( parser.MustStruct( mg.QnameResponse ), nil, nil )
//    addSucc( parser.MustSymbolMap(), nil, nil )
//    addSucc( parser.MustSymbolMap( mg.IdResult, mg.NullVal ), mg.NullVal, nil )
//    addSucc( parser.MustSymbolMap( mg.IdResult, i32Val1 ), i32Val1, nil )
//    addSucc( parser.MustSymbolMap( mg.IdError, mg.NullVal ), nil, mg.NullVal )
//    addSucc( parser.MustSymbolMap( mg.IdError, err1 ), nil, err1 )
//    addSucc( parser.MustSymbolMap( mg.IdError, int32( 1 ) ), nil, i32Val1 )
//    pathRes := objpath.RootedAt( mg.IdResult )
//    pathResF1 := pathRes.Descend( id( 1 ) )
//    pathErr := objpath.RootedAt( mg.IdError )
//    pathErrF1 := pathErr.Descend( id( 1 ) )
//    b.AddTests(
//        &ResponseReactorTest{
//            In: parser.MustStruct( mg.QnameResponse, "result", int32( 1 ) ),
//            ResVal: i32Val1,
//            ResEvents: []EventExpectation{ 
//                { NewValueEvent( i32Val1 ), pathRes },
//            },
//        },
//        &ResponseReactorTest{
//            In: parser.MustSymbolMap( 
//                "result", parser.MustSymbolMap( "f1", int32( 1 ) ) ),
//            ResVal: parser.MustSymbolMap( "f1", int32( 1 ) ),
//            ResEvents: []EventExpectation{
//                { idFact.NextMapStart(), pathRes },
//                { NewFieldStartEvent( id( 1 ) ), pathResF1 },
//                { NewValueEvent( i32Val1 ), pathResF1 },
//                { NewEndEvent(), pathRes },
//            },
//        },
//        &ResponseReactorTest{
//            In: parser.MustSymbolMap( "error", int32( 1 ) ),
//            ErrVal: i32Val1,
//            ErrEvents: []EventExpectation{ 
//                { NewValueEvent( i32Val1 ), pathErr },
//            },
//        },
//        &ResponseReactorTest{
//            In: parser.MustSymbolMap( "error", err1 ),
//            ErrVal: err1,
//            ErrEvents: []EventExpectation{
//                { NewStructStartEvent( err1.Type ), pathErr },
//                { NewFieldStartEvent( id( 1 ) ), pathErrF1 },
//                { NewValueEvent( i32Val1 ), pathErrF1 },
//                { NewEndEvent(), pathErr },
//            },
//        },
//    )
//    addFail := func( in mg.Value, err error ) {
//        b.AddTests( &ResponseReactorTest{ In: in, Error: err } )
//    }
//    addFail(
//        err1.Fields,
//        mg.NewUnrecognizedFieldError( nil, id( 1 ) ),
//    )
//    addFail(
//        parser.MustStruct( "ns1@v1/Response" ),
//        mg.NewTypeCastError( 
//            mg.TypeResponse, typeRef( "ns1@v1/Response" ), nil ),
//    )
//    addFail(
//        parser.MustSymbolMap( mg.IdResult, i32Val1, mg.IdError, err1 ),
//        mg.NewValueCastError( 
//            nil, "response has both a result and an error value" ),
//    )
//}
//
//func initServiceTests( b *ReactorTestSetBuilder ) {
//    initRequestTests( b )
//    initResponseTests( b )
//}

func initReactorTests( b *ReactorTestSetBuilder ) {
    initStructuralReactorTests( b )
    initValueBuildReactorTests( b )
    initPointerReferenceCheckTests( b )
    initEventPathTests( b )
    initFieldOrderReactorTests( b )
//    initServiceTests( b )
}

func init() { 
    reactorTestNs = parser.MustNamespace( "mingle:reactor@v1" )
    AddTestInitializer( reactorTestNs, initReactorTests ) 
}
