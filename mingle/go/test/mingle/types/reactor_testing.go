package types

import (
    mg "mingle"
    mgSvc "mingle/service"
    "bitgirder/objpath"
)

var newVcErr = mg.NewValueCastError

func asType( val interface{} ) mg.TypeReference {
    switch v := val.( type ) {
    case mg.TypeReference: return v
    case string: return mg.MustTypeReference( v )
    }
    panic( libErrorf( "Unhandled type reference: %T", val ) )
}

func newTcErr( expct, act interface{}, p objpath.PathNode ) *mg.ValueCastError {
    return mg.NewTypeCastError( asType( expct ), asType( act ), p )
}

func makeIdList( strs ...string ) []*mg.Identifier {
    res := make( []*mg.Identifier, len( strs ) )
    for i, str := range strs { res[ i ] = mg.MustIdentifier( str ) }
    return res
}

type CastReactorTest struct {
    Map *DefinitionMap
    Type mg.TypeReference
    In mg.Value
    Expect mg.Value
    Err error
}

type rtInit struct {
    tests []interface{}
}

func ( rti *rtInit ) addTests( t ...interface{} ) {
    rti.tests = append( rti.tests, t... )
}

func ( rti *rtInit ) addBaseFieldCastTests() {
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
        return MakeV1DefMap( 
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
        rti.addTests( t )
    }
    s1F1Succ := func( in, expct interface{}, typ string ) {
        s1F1Add( in, expct, typ, nil )
    }
    s1F1Fail := func( in interface{}, typ string, err error ) {
        s1F1Add( in, nil, typ, err )
    }
    tcErr1 := func( expct, act interface{} ) *mg.ValueCastError {
        return newTcErr( expct, act, p( "f1" ) )
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
    s1F1Succ( nil, mg.NullVal, "Int32?" )
    s1F1Succ(
        mg.MustList( "1", nil, int64( 1 ) ),
        mg.MustList( int32( 1 ), nil, int32( 1 ) ),
        "Int32?*",
    )
    s1F1Fail( []byte{}, "Int32", tcErr1( mg.TypeInt32, mg.TypeBuffer ) )
    s1F1Fail( 
        mg.MustList( 1, 2 ), 
        "Int32", 
        tcErr1( mg.TypeInt32, mg.TypeOpaqueList ),
    )
    s1F1Fail( nil, "Int32", newVcErr( p( "f1" ), "Value is null" ) )
    s1F1Fail( int32( 1 ), "Int32+", tcErr1( "Int32+", "Int32" ) )
    s1F1Fail( mg.MustList(), "Int32+", newVcErr( p( "f1" ), "List is empty" ) )
    s1F1Fail( 
        mg.MustList( []byte{} ), 
        "Int32*", 
        newTcErr( "Int32", "Buffer", p( "f1" ).StartList().SetIndex( 0 ) ),
    )
    s1F1Fail( int32( 1 ), "SymbolMap", tcErr1( "SymbolMap", "Int32" ) )
    s1F1Fail( i32L1, "SymbolMap", tcErr1( "SymbolMap", mg.TypeOpaqueList ) )
}

func ( rti *rtInit ) addFieldSetCastTests() {
    dm := MakeV1DefMap(
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
        MakeStructDef(
            "ns1@v1/S3",
            "",
            []*FieldDefinition{ MakeFieldDef( "f1", "Int32?", nil ) },
        ),
    )
    addTest := func( in, expct *mg.Struct, err error ) {
        t := &CastReactorTest{ Map: dm, In: in, Type: in.Type.AsAtomicType() }
        if expct != nil { t.Expect = expct }
        if err != nil { t.Err = err }
        rti.addTests( t )
    }
    addSucc := func( in *mg.Struct ) { addTest( in, in, nil ) }
    addFail := func( in *mg.Struct, err error ) { addTest( in, nil, err ) }
    addSucc( mg.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ) )
    addSucc( mg.MustStruct( "ns1@v1/S2", "f1", int32( 1 ), "f2", int32( 2 ) ) )
    addSucc( mg.MustStruct( "ns1@v1/S3" ) )
    addSucc( mg.MustStruct( "ns1@v1/S3", "f1", int32( 1 ) ) )
    addFail(
        mg.MustStruct( "ns1@v1/S1" ),
        mg.NewMissingFieldsError( nil, makeIdList( "f1" ) ),
    )
    addFail(
        mg.MustStruct( "ns1@v1/S2", "f1", int32( 1 ) ),
        mg.NewMissingFieldsError( nil, makeIdList( "f2" ) ),
    )
    addFail(
        mg.MustStruct( "ns1@v1/S2" ),
        mg.NewMissingFieldsError( nil, makeIdList( "f1", "f2" ) ),
    )
    addFail(
        mg.MustStruct( "ns1@v1/S1", "f1", int32( 1 ), "f2", int32( 2 ) ),
        mg.NewUnrecognizedFieldError( nil, mg.MustIdentifier( "f2" ) ),
    )
    for _, i := range []string{ "1", "2" } {
        addFail(
            mg.MustStruct( "ns1@v1/S" + i, "f3", int32( 3 ) ),
            mg.NewUnrecognizedFieldError( nil, mg.MustIdentifier( "f3" ) ),
        )
    }
}

func ( rti *rtInit ) addStructValCastTests() {
    dm := MakeV1DefMap(
        MakeStructDef( "ns1@v1/S1", "",
            []*FieldDefinition{ MakeFieldDef( "f1", "Int32", nil ) } ),
        MakeEnumDef( "ns1@v1/E1", "e" ),
    )
    t1 := mg.MustTypeReference( "ns1@v1/S1" )
    addFail := func( val interface{}, err error ) {
        rti.addTests(
            &CastReactorTest{ 
                Map: dm, 
                In: mg.MustValue( val ), 
                Type: t1, 
                Err: err,
            },
        )
    }
    tcErr1 := func( act interface{}, p objpath.PathNode ) error {
        return newTcErr( "ns1@v1/S1", act, p )
    }
    addFail( mg.MustStruct( "ns1@v1/S2" ), tcErr1( "ns1@v1/S2", nil ) )
    addFail( int32( 1 ), tcErr1( "Int32", nil ) )
    addFail( mg.MustEnum( "ns1@v1/E1", "e" ), tcErr1( "ns1@v1/E1", nil ) )
    addFail( 
        mg.MustEnum( "ns1@v1/S1", "e" ), 
        newVcErr( nil, "Not an enum type: ns1@v1/S1" ),
    )
    addFail( mg.MustList(), tcErr1( mg.TypeOpaqueList, nil ) )
}

func ( rti *rtInit ) addInferredStructCastTests() {
    // bare symmap at top
    // bare symmap as field
    dm := MakeV1DefMap(
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
        rti.addTests(
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

func ( rti *rtInit ) addEnumValCastTests() {
    dm := MakeV1DefMap(
        MakeStructDef( "ns1@v1/S1", "", []*FieldDefinition{} ),
        MakeEnumDef( "ns1@v1/E1", "c1", "c2" ),
        MakeEnumDef( "ns1@v1/E2", "c1", "c2" ),
    )
    addTest := func( in, expct interface{}, typ interface{}, err error ) {
        t := &CastReactorTest{
            Map: dm,
            In: mg.MustValue( in ),
            Type: asType( typ ),
        }
        if expct != nil { t.Expect = mg.MustValue( expct ) }
        if err != nil { t.Err = err }
        rti.addTests( t )
    }
    addSucc := func( in, expct interface{}, typ interface{} ) { 
        addTest( in, expct, typ, nil ) 
    }
    addFail := func( in interface{}, typ interface{}, err error ) { 
        addTest( in, nil, typ, err ) 
    }
    e1 := mg.MustEnum( "ns1@v1/E1", "c1" )
    for _, quant := range []string{ "", "?" } {
        typStr := "ns1@v1/E1" + quant
        addSucc( e1, e1, typStr )
        addSucc( "c1", e1, typStr )
        addFail( 
            int32( 1 ), 
            typStr, 
            newTcErr( "ns1@v1/E1", mg.TypeInt32, nil ),
        )
    }
    addSucc( mg.MustList( e1, "c1" ), mg.MustList( e1, e1 ), "ns1@v1/E1*" )
    addSucc( mg.NullVal, mg.NullVal, "ns1@v1/E1?" )
    vcErr := func( msg string ) error { return newVcErr( nil, msg ) }
    for _, in := range []interface{} { 
        "c3", mg.MustEnum( "ns1@v1/E1", "c3" ),
    } {
        addFail( 
            in, 
            "ns1@v1/E1",
            vcErr( "illegal value for enum ns1@v1/E1: c3" ),
        )
    }
    addFail( 
        "2bad", 
        "ns1@v1/E1", 
        vcErr( "invalid enum value \"2bad\": [<input>, line 1, col 1]: " +
               "Illegal start of identifier part: \"2\" (U+0032)" ),
    )
    addFail( 
        mg.MustEnum( "ns1@v1/E2", "c1" ), 
        "ns1@v1/E1", 
        newTcErr( "ns1@v1/E1", "ns1@v1/E2", nil ),
    )
    addFail( 
        mg.MustStruct( "ns1@v1/E1" ), 
        "ns1@v1/E1", 
        vcErr( "Not a struct type: ns1@v1/E1" ),
    )
    addFail( 
        int32( 1 ), "ns1@v1/E1+", newTcErr( "ns1@v1/E1+", mg.TypeInt32, nil ) )
    addFail( 
        mg.MustList( int32( 1 ) ), 
        "ns1@v1/E1*", 
        newTcErr( "ns1@v1/E1", mg.TypeInt32, objpath.RootedAtList() ),
    )
    addFail( 
        mg.MustList(), 
        "ns1@v1/E1", 
        newTcErr( "ns1@v1/E1", mg.TypeOpaqueList, nil ),
    )
    addFail( 
        mg.MustSymbolMap(), 
        "ns1@v1/E1", 
        newTcErr( "ns1@v1/E1", mg.TypeSymbolMap, nil ),
    )
    // even though S2 not in type map, we still expect an upstream type cast
    // error
    addFail( 
        mg.MustStruct( "ns1@v1/S2" ), 
        "ns1@v1/E1", 
        newTcErr( "ns1@v1/E1", "ns1@v1/S2", nil ),
    )
}

// Just coverage that structs and defined types don't cause things to go nutso
// when nested in various ways
func ( rti *rtInit ) addDeepCatchallTests() {
    dm := MakeV1DefMap(
        MakeStructDef( "ns1@v1/S1", "",
            []*FieldDefinition{
                MakeFieldDef( "f1", "Int32", int32( 1 ) ),
            },
        ),
        MakeStructDef( "ns1@v1/S2", "",
            []*FieldDefinition{
                MakeFieldDef( "f1", "ns1@v1/S1", nil ),
                MakeFieldDef( "f2", "ns1@v1/E1", nil ),
                MakeFieldDef( "f3", "ns1@v1/S1*", nil ),
                MakeFieldDef( "f4", "ns1@v1/E1+", nil ),
                MakeFieldDef( "f5", "SymbolMap", nil ),
                MakeFieldDef( "f6", "Value", nil ),
                MakeFieldDef( "f7", "Value", nil ),
                MakeFieldDef( "f8", "Value*", nil ),
            },
        ),
        MakeEnumDef( "ns1@v1/E1", "e1", "e2" ),
    )        
    in := mg.MustStruct( "ns1@v1/S2",
        "f1", mg.MustStruct( "ns1@v1/S1", "f1", int32( 2 ) ),
        "f2", "e1",
        "f3", mg.MustList( 
            mg.MustSymbolMap(),
            mg.MustSymbolMap( "f1", int32( 2 ) ),
        ),
        "f4", mg.MustList( mg.MustEnum( "ns1@v1/E1", "e1" ), "e2" ),
        "f5", mg.MustSymbolMap(
            "f1", int32( 1 ),
            "f2", mg.MustEnum( "ns1@v1/E1", "e1" ),
            "f3", mg.MustStruct( "ns1@v1/S1", "f1", int32( 3 ) ),
        ),
        "f6", mg.MustStruct( "ns1@v1/S1" ),
        "f7", mg.MustEnum( "ns1@v1/E1", "e1" ),
        "f8", mg.MustList(
            mg.MustStruct( "ns1@v1/S1", "f1", int32( 2 ) ),
            mg.MustEnum( "ns1@v1/E1", "e2" ),
        ),
    )
    expct := mg.MustStruct( "ns1@v1/S2",
        "f1", mg.MustStruct( "ns1@v1/S1", "f1", int32( 2 ) ),
        "f2", mg.MustEnum( "ns1@v1/E1", "e1" ),
        "f3", mg.MustList(
            mg.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
            mg.MustStruct( "ns1@v1/S1", "f1", int32( 2 ) ),
        ),
        "f4", mg.MustList(
            mg.MustEnum( "ns1@v1/E1", "e1" ),
            mg.MustEnum( "ns1@v1/E1", "e2" ),
        ),
        "f5", mg.MustSymbolMap(
            "f1", int32( 1 ),
            "f2", mg.MustEnum( "ns1@v1/E1", "e1" ),
            "f3", mg.MustStruct( "ns1@v1/S1", "f1", int32( 3 ) ),
        ),
        "f6", mg.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
        "f7", mg.MustEnum( "ns1@v1/E1", "e1" ),
        "f8", mg.MustList(
            mg.MustStruct( "ns1@v1/S1", "f1", int32( 2 ) ),
            mg.MustEnum( "ns1@v1/E1", "e2" ),
        ),
    )
    rti.addTests(
        &CastReactorTest{
            In: in,
            Expect: expct,
            Type: expct.Type.AsAtomicType(),
            Map: dm,
        },
    )
}

func ( rti *rtInit ) addDefaultCastTests() {
    p := func( fldStr string ) objpath.PathNode {
        return objpath.RootedAt( mg.MustIdentifier( fldStr ) )
    }
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
            mg.MustValue( deflPairs[ ( i * 2 ) + 1 ] ),
        )
    }
    dm := MakeV1DefMap(
        MakeStructDef( "ns1@v1/S1", "", s1Flds ),
        MakeStructDef( "ns1@v1/S2", "",
            []*FieldDefinition{
                MakeFieldDef( "f1", "Int32", nil ),
                MakeFieldDef( "f2", "Int32", mg.Int32( int32( 1 ) ) ),
            },
        ),
        MakeStructDef( "ns1@v1/S3", "",
            []*FieldDefinition{ MakeFieldDef( "f1", "Int32*", nil ) } ),
        MakeEnumDef( "ns1@v1/E1", "c1", "c2" ),
    )
    addSucc := func( in, expct mg.Value, typ interface{} ) {
        rti.addTests(
            &CastReactorTest{
                Map: dm,
                In: in,
                Expect: expct,
                Type: asType( typ ),
            },
        )
    }
    addSucc1 := func( in, expct mg.Value ) { addSucc( in, expct, "ns1@v1/S1" ) }
    noDefls := mg.MustStruct( qn1,
        "f1", int32( 1 ), 
        "f2", "s1", 
        "f3", mg.MustEnum( "ns1@v1/E1", "c2" ), 
        "f4", mg.MustList( int32( 1 ) ),
        "f5", false,
    )
    allDefls := mg.MustStruct( qn1, deflPairs... )
    addSucc1( noDefls, noDefls )
    addSucc1( mg.MustStruct( qn1 ), allDefls )
    addSucc1( mg.MustSymbolMap(), allDefls )
    f1OnlyPairs := 
        append( []interface{}{ "f1", int32( 1 ) }, deflPairs[ 2 : ]... )
    addSucc1( 
        mg.MustSymbolMap( "f1", int32( 1 ) ),
        mg.MustStruct( qn1, f1OnlyPairs... ),
    )
    addSucc(
        mg.MustStruct( "ns1@v1/S2", "f1", int32( 1 ) ),
        mg.MustStruct( "ns1@v1/S2", "f1", int32( 1 ), "f2", int32( 1 ) ),
        "ns1@v1/S2",
    )
    addSucc(
        mg.MustStruct( "ns1@v1/S2", "f1", int32( 1 ), "f2", int32( -1 ) ),
        mg.MustStruct( "ns1@v1/S2", "f1", int32( 1 ), "f2", int32( -1 ) ),
        "ns1@v1/S2",
    )
    s3Inst1 :=
        mg.MustStruct( "ns1@v1/S3", 
            "f1", mg.MustList( int32( 1 ), int32( 2 ) ) )
    addSucc( s3Inst1, s3Inst1, "ns1@v1/S3" )
    addSucc( 
        mg.MustStruct( "ns1@v1/S3" ),
        mg.MustStruct( "ns1@v1/S3", "f1", mg.MustList() ),
        "ns1@v1/S3",
    )
    addFail := func( in interface{}, typ interface{}, err error ) {
        rti.addTests(
            &CastReactorTest{
                Map: dm,
                In: mg.MustValue( in ),
                Type: asType( typ ),
                Err: err,
            },
        )
    }
    // nothing special, just base coverage that bad input trumps good defaults
    addFail(
        mg.MustSymbolMap( "f1", []byte{} ),
        "ns1@v1/S1",
        newTcErr( mg.TypeInt32, mg.TypeBuffer, p( "f1" ) ),
    )
}

type EventPathTest struct {
    Source []mg.ReactorEvent
    Expect []mg.EventExpectation
    Type mg.TypeReference
    Map *DefinitionMap
}

func ( rti *rtInit ) addDefaultPathTests() {
    dm := MakeV1DefMap(
        MakeStructDef( "ns1@v1/S1", "",
            []*FieldDefinition{
                MakeFieldDef( "f1", "Int32", nil ),
                MakeFieldDef( "f2", "Int32*", nil ),
                MakeFieldDef( "f3", "SymbolMap", nil ),
                MakeFieldDef( "f4", "Int32", int32( 1 ) ),
            },
        ),
    )
    id := mg.MustIdentifier
    p := func( fld string ) objpath.PathNode {
        return objpath.RootedAt( id( fld ) )
    }
    qn1 := mg.MustQualifiedTypeName( "ns1@v1/S1" )
    t1 := qn1.AsAtomicType()
    ss1 := mg.StructStartEvent{ qn1 }
    fse := func( fld string ) mg.FieldStartEvent {
        return mg.FieldStartEvent{ id( fld ) }
    }
    iv1 := mg.ValueEvent{ mg.Int32( 1 ) }
    src, expct := []mg.ReactorEvent{}, []mg.EventExpectation{}
    apnd := func( ev mg.ReactorEvent, p objpath.PathNode, synth bool ) {
        if ! synth { src = append( src, ev ) }
        expct = append( expct, mg.EventExpectation{ Event: ev, Path: p } )
    }
    apnd( ss1, nil, false )
    apnd( fse( "f1" ), p( "f1" ), false )
    apnd( iv1, p( "f1" ), false )
    apnd( fse( "f2" ), p( "f2" ), false )
    apnd( mg.EvListStart, p( "f2" ), false )
    apnd( iv1, p( "f2" ).StartList().SetIndex( 0 ), false )
    apnd( iv1, p( "f2" ).StartList().SetIndex( 1 ), false )
    apnd( mg.EvEnd, p( "f2" ), false )
    apnd( fse( "f3" ), p( "f3" ), false )
    apnd( mg.EvMapStart, p( "f3" ), false )
    apnd( fse( "f1" ), p( "f3" ).Descend( id( "f1" ) ), false )
    apnd( iv1, p( "f3" ).Descend( id( "f1" ) ), false )
    apnd( mg.EvEnd, p( "f3" ), false )
    apnd( fse( "f4" ), p( "f4" ), true )
    apnd( iv1, p( "f4" ), true )
    apnd( mg.EvEnd, nil, false )
    rti.addTests(
        &EventPathTest{ Source: src, Expect: expct, Type: t1, Map: dm } )
}

type ServiceRequestTest struct {
    In mg.Value
    Parameters *mg.SymbolMap
    Authentication mg.Value
    Error error
}

func ( rti *rtInit ) addServiceRequestTests() {
    addSucc := func( in mg.Value, params *mg.SymbolMap, auth mg.Value ) { 
        rti.addTests(
            &ServiceRequestTest{
                In: in, 
                Parameters: params, 
                Authentication: auth,
            },
        )
    }
    addErr := func( in mg.Value, err error ) {
        rti.addTests( &ServiceRequestTest{ In: in, Error: err } )
    }
    id := mg.MustIdentifier
    pathParams := objpath.RootedAt( mg.IdParameters )
    pathAuth := objpath.RootedAt( mg.IdAuthentication )
    mkReq := func( 
        ns, svc, op string, params, auth interface{} ) *mg.SymbolMap {
        pairs := []interface{}{
            mg.IdNamespace, ns,
            mg.IdService, svc,
            mg.IdOperation, op,
        }
        if params != nil { 
            pairs = append( pairs, mg.IdParameters, mg.MustValue( params ) )
        }
        if auth != nil {
            pairs = append( pairs, mg.IdAuthentication, mg.MustValue( auth ) )
        }
        return mg.MustSymbolMap( pairs... )
    }
    addSucc( 
        mkReq( "ns1@v1", "svc1", "op1", mg.MustSymbolMap( "p1", "1" ), "1" ),
        mg.MustSymbolMap( "p1", int32( 1 ) ),
        mg.Int32( 1 ),
    )
    addSucc(
        mkReq( "ns1@v1", "svc1", "op1", nil, int32( 1 ) ),
        mg.MustSymbolMap(),
        mg.Int32( 1 ),
    )
    op2Params1 := mg.MustSymbolMap(
        "p1", int32( 1 ),
        "p2", mg.MustEnum( "ns1@v1/E1", "e1" ),
        "p3", mg.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
        "p4", mg.MustList( int32( 1 ), int32( 2 ), int32( 3 ) ),
    )
    op2ParamsDefl := func( extra ...interface{} ) *mg.SymbolMap {
        pairs := []interface{}{
            "p1", int32( 42 ),
            "p2", mg.MustEnum( "ns1@v1/E1", "e2" ),
            "p4", mg.MustList( int32( -3 ), int32( -2 ), int32( -1 ) ),
        }
        pairs = append( pairs, extra... )
        return mg.MustSymbolMap( pairs... )
    }
    auth1Val1 := mg.MustStruct( "ns1@v1/Auth1", "p1", int32( 1 ) )
    addSucc(
        mkReq( 
            "ns1@v1", "svc1", "op2",
             mg.MustSymbolMap(
                "p1", "1",
                "p2", "e1",
                "p3", mg.MustSymbolMap( "f1", "1" ),
                "p4", mg.MustList( int64( 1 ), "2", int32( 3 ) ),
            ),
            mg.MustSymbolMap( "p1", "1" ),
        ),
        op2Params1,
        auth1Val1,
    )
    addSucc(
        mkReq( "ns1@v1", "svc1", "op2", nil, auth1Val1 ),
        op2ParamsDefl(),
        auth1Val1,
    )
    addSucc(
        mkReq( "ns1@v1", "svc1", "op2", 
            mg.MustSymbolMap( "p3", mg.MustStruct( "ns1@v1/S1", "f1", "1" ) ),
            auth1Val1,
        ),
        op2ParamsDefl( "p3", mg.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ) ),
        auth1Val1,
    )
    addSucc(
        mkReq( "ns1@v1", "svc1", "op3", mg.MustSymbolMap(), nil ),
        mg.MustSymbolMap(), 
        nil,
    )
    addSucc(
        mkReq( "ns1@v1", "svc1", "op3", nil, nil ), 
        mg.MustSymbolMap(),
        nil,
    )
    addErr(
        mkReq( 
            "ns1@v1", "svc1", "op1", 
            mg.MustSymbolMap( "p1", int32( 1 ) ),
            nil,
        ),
        mgSvc.ErrAuthenticationMissing,
    )
    addErr(
        mkReq( 
            "ns1@v1", "svc1", "op2", mg.MustSymbolMap( "p1", []byte{} ), nil ),
        newTcErr( 
            mg.TypeInt32, mg.TypeBuffer, pathParams.Descend( id( "p1" ) ) ),
    )
    addErr(
        mkReq( "ns1@v1", "svc1", "op2", mg.MustSymbolMap( "p2", "bad" ), nil ),
        newVcErr( pathParams.Descend( id( "p2" ) ), "STUB" ),
    )
    addErr(
        mkReq( 
            "ns1@v1", "svc1", "op2", 
            op2Params1, 
            mg.MustStruct( "ns1@v1/Auth2" ),
        ),
        newTcErr( "ns1@v1/Auth1", "ns1@v1/Auth2", pathAuth ),
    )
    addErr(
        mkReq( "ns1@v1", "svc1", "op1", 
            mg.MustSymbolMap( "not-a-field", false ),
            int32( 1 ),
        ),
        mg.NewUnrecognizedFieldError( pathParams, id( "not-a-field" ) ),
    ) 
    addErr(
        mkReq( "ns1@v1", "svc1", "op4", nil, nil ),
        mg.NewMissingFieldsError( pathParams, makeIdList( "p1", "p2" ) ),
    )
    addErr(
        mkReq( "ns1@v2", "svc1", "op1", nil, nil ),
        &mgSvc.NoSuchNamespaceError{ mg.MustNamespace( "ns1@v2" ) },
    )
    addErr(
        mkReq( "ns1@v1", "svc2", "op1", nil, nil ),
        &mgSvc.NoSuchServiceError{ mg.MustIdentifier( "svc2" ) },
    )
    addErr(
        mkReq( "ns1@v1", "svc1", "badOp", nil, nil ),
        &mgSvc.NoSuchOperationError{ mg.MustIdentifier( "badOp" ) },
    )
}

type ServiceResponseTest struct {
    In mg.Value
    Expect mg.Value
    ResultType mg.TypeReference
    Error error
}

func ( rti *rtInit ) addServiceResponseTests() {
    okResp := func( in interface{} ) mg.Value {
        return mg.MustSymbolMap( "result", in )
    }
    errResp := func( in interface{} ) mg.Value {
        return mg.MustSymbolMap( "error", in )
    }
    id := mg.MustIdentifier
    pathRes := objpath.RootedAt( mg.IdResult )
    pathErr := objpath.RootedAt( mg.IdError )
    addSucc := func( in, expct mg.Value, resTyp interface{} ) {
        rti.addTests(
            &ServiceResponseTest{
                In: in,
                Expect: expct,
                ResultType: asType( resTyp ),
            },
        )
    }
    addResSucc := func( in, expct interface{}, resTyp interface{} ) {
        addSucc( okResp( in ), okResp( expct ), resTyp )
    }
    addErrSucc := func( in, expct interface{} ) {
        addSucc( errResp( in ), errResp( expct ), mg.TypeNullableValue )
    }
    addResSucc( mg.NullVal, mg.NullVal, mg.TypeNull )
    addResSucc( nil, nil, mg.TypeNull )
    addResSucc( int32( 1 ), int32( 1 ), mg.TypeInt32 )
    addResSucc( "1", int32( 1 ), mg.TypeInt32 )
    addResSucc( mg.NullVal, nil, "Int32?" )
    addResSucc( nil, nil, "Int32?" )
    en1 := mg.MustEnum( "ns1@v1/E1", "e1" )
    addResSucc( en1, en1, "ns1@v1/E1" )
    addResSucc( "e1", en1, "ns1@v1/E1" )
    s1 := mg.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) )
    addResSucc( s1, s1, "ns1@v1/S1" )
    addResSucc( mg.MustStruct( "ns1@v1/S1" ), s1, "ns1@v1/S1" )
    err1 := mg.MustStruct( "ns1@v1/Err1", "f1", int32( 1 ) )
    addErrSucc( err1, err1 )
    addErrSucc( mg.MustStruct( "ns1@v1/Err1" ), err1 )
    // We're not really checking here that the error values are correct as
    // mingle struct values (most should have at least a 'message' field), only
    // that the types are allowed by the response cast even when not explicitly
    // declared in an operation definition (which will be the common case)
    for _, errTyp := range []string{
        "mingle:core@v1/MissingFieldsError",
        "mingle:core@v1/UnrecognizedFieldError",
        "mingle:core@v1/TypeCastError",
        "mingle:core@v1/ValueCastError",
        "mingle:core@v1/UnrecognizedEndpointError",
    } {
        err := mg.MustStruct( errTyp )
        addErrSucc( err, err )
    }
    addFail := func( in interface{}, resTyp interface{}, err error ) {
        rti.addTests(
            &ServiceResponseTest{
                In: mg.MustValue( in ),
                ResultType: asType( resTyp ),
                Error: err,
            },
        )
    }
    addResFail := func( in interface{}, resTyp interface{}, err error ) {
        addFail( okResp( in ), resTyp, err )
    }
    addErrFail := func( in interface{}, err error ) {
        addFail( errResp( in ), mg.TypeNullableValue, err )
    }
    addResFail( 
        []byte{},
        mg.TypeInt32, 
        newTcErr( mg.TypeInt32, mg.TypeBuffer, pathRes ),
    )
    addResFail(
        int32( 1 ),
        mg.TypeNull,
        newTcErr( mg.TypeNull, mg.TypeInt32, pathRes ),
    )
    addResFail( 
        mg.MustStruct( "ns1@v1/S2" ),
        "ns1@v1/S1", 
        newTcErr( "ns1@v1/S1", "ns1@v1/S2", pathRes ),
    )
    addResFail(
        mg.MustEnum( "ns1@v1/E1", "bad" ),
        "ns1@v1/E1",
        newVcErr( pathRes, "STUB" ),
    )
    addErrFail( mg.MustStruct( "ns1@v1/BadErr" ), newVcErr( pathErr, "STUB" ) )
    addErrFail(
        mg.MustStruct( "ns1@v1/Err1", "not-a-field", int32( 1 ) ),
        mg.NewUnrecognizedFieldError( pathErr, id( "not-a-field" ) ),
    )
    addErrFail(
        mg.MustStruct( "ns1@v1/UndeclaredErr" ),
        newVcErr( pathErr, "STUB" ),
    )
}

func ( rti *rtInit ) init() {
    rti.addBaseFieldCastTests()
    rti.addFieldSetCastTests()
    rti.addStructValCastTests() 
    rti.addInferredStructCastTests()
    rti.addEnumValCastTests()
    rti.addDeepCatchallTests()
    rti.addDefaultCastTests()
    rti.addDefaultPathTests()
//    rti.addServiceRequestTests()
//    rti.addServiceResponseTests()
}

// The tests returned might normally be created during an init() block, but
// creating them benefits from using methods, like NewV1DefinitionMap(), which
// themselves are not safe to call until after package init.
func GetStdReactorTests() []interface{} {
    rti := &rtInit{ tests: []interface{}{} }
    rti.init()
    return rti.tests
}
