package types

import (
    mg "mingle"
    "bitgirder/objpath"
)

var StdReactorTests = []interface{}{}

func addStdReactorTests( t ...interface{} ) {
    StdReactorTests = append( StdReactorTests, t... )
}

var newVcErr = mg.NewValueCastError

type CastReactorTest struct {
    Map *DefinitionMap
    Type mg.TypeReference
    In mg.Value
    Expect mg.Value
    Err error
}

func addBaseFieldCastTests() {
    p := func( fld string ) objpath.PathNode {
        return objpath.RootedAt( mg.MustIdentifier( fld ) )
    }
    qn1Str := "ns1@v1/S1"
    qn1 := mg.MustQualifiedTypeName( qn1Str )
    s1F1 := func( val interface{} ) *mg.Struct {
        return mg.MustStruct( qn1, "f1", val )
    }
    s1DefMap := func( typ string ) *DefinitionMap {
        fld := MakeFieldDef( "f1", typ, nil )
        return MakeDefMap( 
            MakeStructDef( qn1Str, "", []*FieldDefinition{ fld } ) )
    }
    s1F1Add := func( in, expct interface{}, typ string, err error ) {
        t := &CastReactorTest{ 
            Type: qn1.AsAtomicType(),
            In: s1F1( in ), 
            Map: s1DefMap( typ ),
        }
        if expct != nil { t.Expect = s1F1( expct ) }
        if err != nil { t.Err = err }
        addStdReactorTests( t )
    }
    s1F1Succ := func( in, expct interface{}, typ string ) {
        s1F1Add( in, expct, typ, nil )
    }
    s1F1Fail := func( in interface{}, typ string, err error ) {
        s1F1Add( in, nil, typ, err )
    }
    s1F1Succ( int32( 1 ), int32( 1 ), "Int32" )
    s1F1Succ( "1", int32( 1 ), "Int32" )
    s1F1Succ( int32( 1 ), int32( 1 ), "Value" )
    i32L1 := mg.MustList( int32( 1 ), int32( 2 ), int32( 3 ) )
    s1F1Succ( i32L1, i32L1, "Int32+" )
    s1F1Succ( i32L1, i32L1, "Value" )
    s1F1Succ( i32L1, i32L1, "Value*" )
    s1F1Succ( mg.MustList( "1", int64( 2 ), int32( 3 ) ), i32L1, "Int32*" )
    sm1 := mg.MustSymbolMap( "f1", int32( 1 ) )
    s1F1Succ( sm1, sm1, "SymbolMap" )
    s1F1Succ( sm1, sm1, "Value" )
    s1F1Succ( int32( 1 ), int32( 1 ), "Int32?" )
    s1F1Succ( nil, nil, "Int32?" )
    s1F1Fail( []byte{}, "Int32", newVcErr( p( "f1" ), "STUB" ) )
    s1F1Fail( mg.MustList( 1, 2 ), "Int32", newVcErr( p( "f1" ), "STUB" ) )
    s1F1Fail( nil, "Int32", newVcErr( p( "f1" ), "STUB" ) )
    s1F1Fail( int32( 1 ), "Int32+", newVcErr( p( "f1" ), "STUB" ) )
    s1F1Fail( mg.MustList(), "Int32+", newVcErr( p( "f1" ), "STUB" ) )
    s1F1Fail( 
        mg.MustList( []byte{} ), 
        "Int32*", 
        newVcErr( p( "f1" ).StartList().SetIndex( 0 ), "STUB" ),
    )
    s1F1Fail( int32( 1 ), "SymbolMap", newVcErr( p( "f1" ), "STUB" ) )
    s1F1Fail( i32L1, "SymbolMap", newVcErr( p( "f1" ), "STUB" ) )
}

func addFieldSetCastTests() {
    dm := MakeDefMap(
        MakeStructDef(
            "ns1@v1/S1",
            "",
            []*FieldDefinition{ MakeFieldDef( "f1", "Int32", nil ) },
        ),
        MakeStructDef(
            "ns1@v1/S2",
            "ns1@v1/S1",
            []*FieldDefinition{ MakeFieldDef( "f2", "Int32", nil ) },
        ),
    )
    addTest := func( in, expct *mg.Struct, err error ) {
        t := &CastReactorTest{ Map: dm, In: in, Type: in.Type.AsAtomicType() }
        if expct != nil { t.Expect = expct }
        if err != nil { t.Err = err }
        addStdReactorTests( t )
    }
    addSucc := func( in *mg.Struct ) { addTest( in, in, nil ) }
    addFail := func( in *mg.Struct, err error ) { addTest( in, nil, err ) }
    addSucc( mg.MustStruct( "ns1@v1/S1", "f1", 1 ) )
    addSucc( mg.MustStruct( "ns1@v1/S2", "f1", 1, "f2", 2 ) )
    addFail(
        mg.MustStruct( "ns1@v1/S1" ),
        newVcErr( nil, "Missing value for field: f1" ),
    )
    addFail(
        mg.MustStruct( "ns1@v1/S2", "f1", 1 ),
        newVcErr( nil, "Missing value for field: f2" ),
    )
    addFail( 
        mg.MustStruct( "ns1@v1/S1", "f1", 1, "f2", 2 ),
        newVcErr( nil, "Unrecognized field: f2" ),
    )
    for _, i := range []string{ "1", "2" } {
        addFail(
            mg.MustStruct( "ns1@v1/S" + i, "f3", 3 ),
            newVcErr( nil, "Unrecognized field: f3" ),
        )
    }
}

func addInferredStructCastTests() {
    // bare symmap at top
    // bare symmap as field
    dm := MakeDefMap(
        MakeStructDef(
            "ns1@v1/S1",
            "",
            []*FieldDefinition{
                MakeFieldDef( "f1", "Int32?", nil ),
                MakeFieldDef( "f2", "ns1@v1/S2?", nil ),
            },
        ),
        MakeStructDef(
            "ns1@v1/S2",
            "",
            []*FieldDefinition{ MakeFieldDef( "f1", "Int32", nil ) },
        ),
    )
    addSucc := func( in, expct mg.Value ) {
        addStdReactorTests(
            &CastReactorTest{
                Map: dm,
                Type: mg.MustTypeReference( "ns1@v1/S1" ),
                In: in,
                Expect: expct,
            },
        )
    }
    addSucc( 
        mg.MustSymbolMap( "f1", int32( 1 ) ),
        mg.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
    )
    addSucc( 
        mg.MustSymbolMap( "f2", mg.MustSymbolMap( "f1", int32( 1 ) ) ),
        mg.MustStruct( "ns1@v1/S1",
            "f2", mg.MustStruct( "ns1@v1/S2", "f1", int32( 1 ) ) ),
    )
}

func addStructValCastTests() {
    dm := MakeDefMap(
        MakeStructDef( "ns1@v1/S1", "",
            []*FieldDefinition{ MakeFieldDef( "f1", "Int32", nil ) } ),
        MakeEnumDef( "ns1@v1/E1", "e" ),
    )
    addFail := func( val mg.Value, err error ) {
        addStdReactorTests(
            &CastReactorTest{
                Map: dm,
                In: val,
                Type: mg.MustTypeReference( "ns1@v1/S1" ),
                Err: err,
            },
        )
    }
    addFail( mg.MustStruct( "ns1@v1/S2" ), newVcErr( nil, "STUB" ) )
    addFail( int32( 1 ), newVcErr( nil, "STUB" ) )
    addFail( 
        mg.MustSymbolMap( "f1", int32( 1 ), "f2", int32( 2 ) ),
        newVcErr( nil, "STUB" ),
    )
    addFail( mg.MustEnum( "ns1@v1/E1", "e" ), newVcErr( nil, "STUB" ) )
    addFail( mg.MustList(), newVcErr( nil, "STUB" ) )
}

func addEnumValCastTests() {
    dm := MakeDefMap(
        MakeStructDef( 
            "ns1@v1/S1", "", 
            []*FieldDefinition{ MakeFieldDef( "f1", "ns1@v1/E1", nil ) },
        ),
        MakeStructDef( "ns1@v1/S2", "", []*FieldDefinition{} ),
        MakeEnumDef( "ns1@v1/E1", "c1", "c2" ),
        MakeEnumDef( "ns1@v1/E2", "c1", "c2" ),
    )
    addTest := func( in, expct mg.Value, err error ) {
        t := &CastReactorTest{
            Map: dm,
            In: mg.MustStruct( "ns1@v1/S1", "f1", in ), 
            Type: mg.MustTypeReference( "ns1@v1/S1" ),
        }
        if expct != nil { t.Expect = mg.MustStruct( "ns1@v1/S1", "f1", expct ) }
        if err != nil { t.Err = err }
        addStdReactorTests( t )
    }
    addSucc := func( in, expct mg.Value ) { addTest( in, expct, nil ) }
    addFail := func( in mg.Value, err error ) { addTest( in, nil, err ) }
    e1 := mg.MustEnum( "ns1@v1/E1", "c1" )
    addSucc( e1, e1 )
    addSucc( "c1", e1 )
    errStub := newVcErr( nil, "STUB" )
    addFail( mg.MustEnum( "ns1@v1/E1", "c3" ), errStub )
    addFail( mg.MustEnum( "ns1@v1/E2", "c1" ), errStub )
    addFail( int32( 1 ), errStub )
    addFail( mg.MustList(), errStub )
    addFail( mg.MustSymbolMap(), errStub )
    addFail( mg.MustStruct( "ns1@v1/S2" ), errStub )
}

func addCastTests() {
    addBaseFieldCastTests()
    addFieldSetCastTests()
    addStructValCastTests() 
    addEnumValCastTests()
    addInferredStructCastTests()
}

func addDefaultTests() {
    qn1 := mg.MustQualifiedTypeName( "ns1@v1/S1" )
    e1 := mg.MustEnum( "ns1@v1/E1", "c1" )
    deflPairs := []interface{}{
        "f1", int32( 0 ),
        "f2", "str-defl",
        "f3", e1,
        "f4", mg.MustList( int32( 0 ), int32( 1 ) ),
        "f5", true,
    }
    s1FldTyps := []string{ "Int32", "String", "ns1@v1/E1", "Int32+", "Boolean" }
    if len( s1FldTyps ) != len( deflPairs ) / 2 { panic( "Mismatched len" ) }
    s1Flds := make( []*FieldDefinition, len( s1FldTyps ) )
    for i, typ := range s1FldTyps {
        s1Flds[ i ] = MakeFieldDef(
            deflPairs[ i * 2 ].( string ),
            typ,
            deflPairs[ ( i * 2 ) + 1 ].( mg.Value ),
        )
    }
    dm := MakeDefMap(
        MakeStructDef( "ns1@v1/S1", "", s1Flds ),
        MakeEnumDef( "ns1@v1/E1", "c1", "c2" ),
    )
    addSucc := func( in, expct mg.Value ) {
        addStdReactorTests(
            &CastReactorTest{
                Map: dm,
                In: in,
                Expect: expct,
                Type: qn1.AsAtomicType(),
            },
        )
    }
    noDefls := mg.MustStruct( qn1,
        "f1", int32( 1 ), 
        "f2", "s1", 
        "f3", mg.MustEnum( "ns1@v1/E1", "c2" ), 
        "f4", mg.MustList( int32( 1 ) ),
        "f5", false,
    )
    allDefls := mg.MustStruct( qn1, deflPairs... )
    addSucc( noDefls, noDefls )
    addSucc( mg.MustStruct( qn1 ), allDefls )
    addSucc( mg.MustSymbolMap(), allDefls )
    f1OnlyPairs := 
        append( []interface{}{ "f1", int32( 1 ) }, deflPairs[ 2 : ]... )
    addSucc( 
        mg.MustSymbolMap( "f1", int32( 1 ) ),
        mg.MustStruct( qn1, f1OnlyPairs... ),
    )
    // nothing special, just base coverage that bad input trumps good defaults
    addStdReactorTests(
        &CastReactorTest{
            Map: dm,
            In: mg.MustSymbolMap( "f1", []byte{} ),
            Type: qn1.AsAtomicType(),
            Err: newVcErr( nil, "STUB" ),
        },
    )
}

func init() {
    addCastTests()
    addDefaultTests()
}
