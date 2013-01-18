package mingle

import (
    "bitgirder/objpath"
    "fmt"
    "encoding/base64"
)

type ReactorTest interface {}

var StdReactorTests []ReactorTest

func AddStdReactorTests( t ...ReactorTest ) {
    StdReactorTests = append( StdReactorTests, t... )
}

type ValueBuildTest struct { Val Value }

func initValueBuildReactorTests() {
    s1 := MustStruct( "ns1@v1/S1",
        "val1", String( "hello" ),
        "list1", MustList(),
        "map1", MustSymbolMap(),
        "struct1", MustStruct( "ns1@v1/S2" ),
    )
    mk := func( v Value ) ReactorTest { return ValueBuildTest{ v } }
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

type StructuralReactorPathTest struct {
    Events []EventExpectation
    StartPath objpath.PathNode
    FinalPath objpath.PathNode
}

func initStructuralReactorTests() {
    evStartStruct1 := StructStartEvent{ qname( "ns1@v1/S1" ) }
    idF1 := MustIdentifier( "f1" )
    evStartField1 := FieldStartEvent{ idF1 }
    idF2 := MustIdentifier( "f2" )
    evStartField2 := FieldStartEvent{ idF2 }
    evValue1 := ValueEvent{ Int64( int64( 1 ) ) }
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
    mk3 := func( finalPath idPath, 
                 evs ...EventExpectation ) *StructuralReactorPathTest {
        return &StructuralReactorPathTest{ Events: evs, FinalPath: finalPath }
    }
    idPath1 := objpath.RootedAt( idF1 )
    lpRoot := func() *objpath.ListNode { return objpath.RootedAtList() }
    StdReactorTests = append( StdReactorTests,
        mk1( "Saw start of field 'f2' while expecting a value for 'f1'",
            evStartStruct1, evStartField1, evStartField2,
        ),
        mk1( "Saw start of field 'f2' while expecting a value for 'f1'",
            evStartStruct1, evStartField1, EvMapStart, evStartField1,
            evStartField2,
        ),
        mk1( "StartField() called, but struct is built",
            evStartStruct1, EvEnd, evStartField1,
        ),
        mk1( "Expected field name or end of fields but got value",
            evStartStruct1, evValue1,
        ),
        mk1( "Expected field name or end of fields but got list start",
            evStartStruct1, EvListStart,
        ),
        mk1( "Expected field name or end of fields but got map start",
            evStartStruct1, EvMapStart,
        ),
        mk1( "Expected field name or end of fields but got struct start",
            evStartStruct1, evStartStruct1,
        ),
        mk1( "Saw end while expecting value for field 'f1'",
            evStartStruct1, evStartField1, EvEnd,
        ),
        mk1( "Expected list value but got start of field 'f1'",
            evStartStruct1, evStartField1, EvListStart, evStartField1,
        ),
        mk2( "Expected struct but got value", ReactorTopTypeStruct, evValue1 ),
        mk2( "Expected struct but got list start", ReactorTopTypeStruct,
            EvListStart,
        ),
        mk2( "Expected struct but got map start", ReactorTopTypeStruct,
            EvMapStart,
        ),
        mk2( "Expected struct but got field 'f1'", ReactorTopTypeStruct,
            evStartField1,
        ),
        mk2( "Expected struct but got end", ReactorTopTypeStruct, EvEnd ),
        mk1( "Multiple entries for field: f1",
            evStartStruct1, 
            evStartField1, evValue1,
            evStartField2, evValue1,
            evStartField1,
        ),
        mk3( nil ),
        mk3( nil, EventExpectation{ evValue1, nil } ),
        mk3( idPath1, 
             EventExpectation{ evStartStruct1, nil }, 
             EventExpectation{ evStartField1, idPath1 },
        ),
        mk3( idPath1,
             EventExpectation{ EvMapStart, nil },
             EventExpectation{ evStartField1, idPath1 },
        ),
        mk3( nil, 
             EventExpectation{ evStartStruct1, nil },
             EventExpectation{ evStartField1, idPath1 },
             EventExpectation{ evValue1, idPath1 },
        ),
        mk3( lpRoot(), EventExpectation{ EvListStart, nil } ),
        mk3( lpRoot(),
             EventExpectation{ EvListStart, nil },
             EventExpectation{ EvMapStart, lpRoot() },
        ),
        mk3( lpRoot().SetIndex( 1 ),
             EventExpectation{ EvListStart, nil },
             EventExpectation{ evValue1, lpRoot() },
        ),
        mk3( lpRoot().SetIndex( 1 ),
             EventExpectation{ EvListStart, nil },
             EventExpectation{ evValue1, lpRoot() },
             EventExpectation{ EvMapStart, lpRoot().SetIndex( 1 ) },
        ),
        mk3( idPath1,
             EventExpectation{ EvMapStart, nil },
             EventExpectation{ evStartField1, idPath1 },
             EventExpectation{ EvMapStart, idPath1 },
        ),
        mk3( idPath1.Descend( idF2 ).
                     StartList().
                     Next().
                     Next().
                     StartList().
                     Next().
                     Descend( idF1 ),
             EventExpectation{ evStartStruct1, nil },
             EventExpectation{ evStartField1, idPath1 },
             EventExpectation{ EvMapStart, idPath1 },
             EventExpectation{ evStartField2, idPath1.Descend( idF2 ) },
             EventExpectation{ EvListStart, idPath1.Descend( idF2 ) },
             EventExpectation{ evValue1,
                idPath1.Descend( idF2 ).StartList().SetIndex( 0 ),
             },
             EventExpectation{ evValue1, 
                idPath1.Descend( idF2 ).StartList().SetIndex( 1 ),
             },
             EventExpectation{ EvListStart, 
                idPath1.Descend( idF2 ).StartList().SetIndex( 2 ),
             },
             EventExpectation{ evValue1, 
                idPath1.Descend( idF2 ).StartList().SetIndex( 2 ).
                        StartList().SetIndex( 0 ),
             },
             EventExpectation{ EvMapStart, 
                idPath1.Descend( idF2 ).StartList().SetIndex( 2 ).
                        StartList().SetIndex( 1 ),
             },
             EventExpectation{ evStartField1, 
                idPath1.Descend( idF2 ).StartList().SetIndex( 2 ).
                        StartList().SetIndex( 1 ).Descend( idF1 ),
             },
        ),
        &StructuralReactorPathTest{
            Events: []EventExpectation{ 
                { EvMapStart, nil },
                { evStartField1, idPath1 },
            },
            StartPath: objpath.RootedAt( idF2 ).StartList().SetIndex( 3 ),
            FinalPath: 
                objpath.RootedAt( idF2 ).
                    StartList().
                    SetIndex( 3 ).
                    Descend( idF1 ),
        },
    )
}

type FieldOrderReactorTest struct {
    Source []ReactorEvent
    Expect Value
    Order []*Identifier
}

func initFieldOrderValueTests() {
    flds := make( []ReactorEvent, 5 )
    ids := make( []*Identifier, len( flds ) )
    for i := 0; i < len( flds ); i++ {
        ids[ i ] = id( fmt.Sprintf( "f%d", i ) )
        flds[ i ] = FieldStartEvent{ ids[ i ] }
    }
    i1 := Int32( int32( 1 ) )
    val1 := ValueEvent{ i1 }
    t1, t2 := qname( "ns1@v1/S1" ), qname( "ns1@v1/S2" )
    ss1, ss2 := StructStartEvent{ t1 }, StructStartEvent{ t2 }
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
            EvMapStart, 
                flds[ 0 ], val1, 
                flds[ 1 ], ss2, flds[ 0 ], val1, EvEnd,
            EvEnd,
        },
        []ReactorEvent{ EvListStart, val1, val1, EvEnd },
        []ReactorEvent{ ss2, flds[ 0 ], val1, EvEnd },
        []ReactorEvent{ val1 },
    }
    mkSrc := func( ord ...int ) []ReactorEvent {
        res := []ReactorEvent{ ss1 }
        for _, fldNum := range ord {
            res = append( res, flds[ fldNum ] )
            res = append( res, fldEvs[ fldNum ]... )
        }
        return append( res, EvEnd )
    }
    addTest1 := func( src []ReactorEvent, expct Value ) {
        AddStdReactorTests(
            &FieldOrderReactorTest{ 
                Source: src, 
                Expect: expct, 
                Order: []*Identifier{ ids[ 0 ], ids[ 3 ], ids[ 2 ], ids[ 1 ] },
            },
        )
    }
    for _, ord := range [][]int {
        []int{ 0, 3, 2, 1 }, // first one should be straight passthrough
        []int{ 0, 1, 2, 3 },
        []int{ 3, 2, 1, 0 },
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
            Source: []ReactorEvent{
                ss1, 
                    flds[ 0 ], val1,
                    flds[ 1 ], ss1,
                        flds[ 2 ], EvListStart, val1, EvEnd,
                        flds[ 1 ], val1,
                    EvEnd,
                EvEnd,
            },
            Order: []*Identifier{ ids[ 1 ], ids[ 0 ], ids[ 2 ] },
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
}

type FieldOrderPathTest struct {
    Source []ReactorEvent
    Expect []EventExpectation
    Order []*Identifier
}

func initFieldOrderPathTests() {
    i1 := Int32( int32( 1 ) )
    val1 := ValueEvent{ i1 }
    id := func( i int ) *Identifier {
        return MustIdentifier( fmt.Sprintf( "f%d", i ) )
    }
    typ := func( i int ) *QualifiedTypeName {
        return qname( fmt.Sprintf( "ns1@v1/S%d", i ) )
    }
    ss := func( i int ) StructStartEvent { return StructStartEvent{ typ( i ) } }
    fld := func( i int ) FieldStartEvent { return FieldStartEvent{ id( i ) } }
    p := func( i int ) objpath.PathNode { return objpath.RootedAt( id( i ) ) }
    expct1 := []EventExpectation{
        { ss( 1 ), nil },
            { fld( 2 ), p( 2 ) },
            { EvListStart, p( 2 ) },
                { val1, p( 2 ).StartList().SetIndex( 0 ) },
                { val1, p( 2 ).StartList().SetIndex( 1 ) },
            { EvEnd, p( 2 ) },
            { fld( 0 ), p( 0 ) },
            { val1, p( 0 ) },
            { fld( 3 ), p( 3 ) },
            { ss( 2 ), p( 3 ) },
                { fld( 1 ), p( 3 ).Descend( id( 1 ) ) },
                { val1, p( 3 ).Descend( id( 1 ) ) },
                { fld( 2 ), p( 3 ).Descend( id( 2 ) ) },
                { EvListStart, p( 3 ).Descend( id( 2 ) ) },
                    { val1, 
                      p( 3 ).Descend( id( 2 ) ).StartList().SetIndex( 0 ) },
                    { val1, 
                      p( 3 ).Descend( id( 2 ) ).StartList().SetIndex( 1 ) },
                { EvEnd, p( 3 ).Descend( id( 2 ) ) },
            { EvEnd, p( 3 ) },
            { fld( 1 ), p( 1 ) },
            { EvMapStart, p( 1 ) },
                { fld( 1 ), p( 1 ).Descend( id( 1 ) ) },
                { val1, p( 1 ).Descend( id( 1 ) ) },
                { fld( 0 ), p( 1 ).Descend( id( 0 ) ) },
                { val1, p( 1 ).Descend( id( 0 ) ) },
            { EvEnd, p( 1 ) },
        { EvEnd, nil },
    }
    ord1 := []*Identifier{ id( 2 ), id( 0 ), id( 3 ), id( 1 ) }
    evs := [][]ReactorEvent{
        []ReactorEvent{ val1 },
        []ReactorEvent{ EvMapStart, fld( 1 ), val1, fld( 0 ), val1, EvEnd },
        []ReactorEvent{ EvListStart, val1, val1, EvEnd },
        []ReactorEvent{ 
            ss( 2 ), 
                fld( 1 ), val1, 
                fld( 2 ), EvListStart, val1, val1, EvEnd,
            EvEnd,
        },
    }
    mkSrc := func( ord ...int ) []ReactorEvent {
        res := []ReactorEvent{ ss( 1 ) }
        for _, i := range ord {
            res = append( res, fld( i ) )
            res = append( res, evs[ i ]... )
        }
        return append( res, EvEnd )
    } 
    for _, ord := range [][]int{
        []int{ 0, 1, 2, 3 },
        []int{ 3, 2, 1, 0 },
        []int{ 2, 0, 3, 1 },
    } {
        AddStdReactorTests(
            &FieldOrderPathTest{
                Source: mkSrc( ord... ),
                Expect: expct1,
                Order: ord1,
            },
        )
    }
}

func initFieldOrderReactorTests() {
    initFieldOrderValueTests()
    initFieldOrderPathTests()
}

type CastReactorTest struct {
    In Value
    Expect Value
    Path objpath.PathNode
    Type TypeReference
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
    err := newTypeCastError( 
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
    t.addTcError( MustList( 1, 2 ), TypeString, "Value*" )
    t.addTcError( MustList(), "String?", "Value*" )
    t.addTcError( "s", "String*", "String" )
    t.addCrtDefault(
        &CastReactorTest{
            In: MustList( 1, t.struct1 ),
            Type: asTypeReference( "Int32*" ),
            Err: newTypeCastError(
                asTypeReference( "Int32" ),
                &AtomicTypeReference{ Name: t.struct1.Type },
                crtPathDefault.StartList().Next(),
            ),
        },
    )
    t.addTcError( t.struct1, "Int32?", t.struct1.Type )
    t.addTcError( 12, t.struct1.Type, "Int64" )
    for _, prim := range PrimitiveTypes {
        if prim != TypeValue { // Value would actually be valid cast
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
    for _, valCtx := range numTests {
        t.addSucc( numCtx.val, numCtx.str, TypeString )
        t.addSucc( numCtx.str, numCtx.val, numCtx.typ )
        t.addSucc( valCtx.val, numCtx.val, numCtx.typ )
    }}
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
    t.addVcError( "-1", TypeUint32, "invalid syntax: -1" )
    rngErr( "18446744073709551616", TypeUint64 )
    t.addVcError( "-1", TypeUint64, "invalid syntax: -1" )
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
    t.addSucc( intList1, intList1, typeOpaqueList )
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
    addTcErr := func( in interface{}, expct, act TypeReferenceInitializer ) {
        add( t.createTcError( in, expct, act ) )
    }
    t1 := qname( "ns1@v1/T1" )
    t2 := qname( "ns1@v1/T2" )
    s1 := MustStruct( t1, "f1", int32( 1 ) )
    addSucc( MustStruct( t1, "f1", "1" ), s1, t1 )
    addSucc( MustSymbolMap( "f1", "1" ), s1, t1 )
    arb := MustStruct( "ns1@v1/Arbitrary", "f1", int32( 1 ) )
    addSucc( arb, arb, arb.Type )
    s2InFlds := MustSymbolMap( "f1", "1", "f2", MustSymbolMap( "f1", "1" ) )
    s2 := MustStruct( t2, "f1", int32( 1 ), "f2", s1 )
    addSucc( &Struct{ Type: t2, Fields: s2InFlds }, s2, t2 )
    addSucc( s2InFlds, s2, t2 )
    addTcErr( MustStruct( t2, "f1", int32( 1 ) ), t1, t2 )
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
    StdReactorTests = []ReactorTest{}
    initValueBuildReactorTests()
    initStructuralReactorTests()
    initFieldOrderReactorTests()
    initCastReactorTests()
}
