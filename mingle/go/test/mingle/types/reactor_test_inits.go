package types

import (
    mg "mingle"
    "mingle/parser"
    mgRct "mingle/reactor"
    "bitgirder/objpath"
    "encoding/base64"
    "fmt"
)

var reactorTestNs *mg.Namespace

var newVcErr = mg.NewValueCastError

var pathInVal = objpath.RootedAt( mkId( "inVal" ) )

func mustHeapVal( v interface{} ) *mg.HeapValue {
    return mg.NewHeapValue( mg.MustValue( v ) )
}

func newTcErr( expct, act interface{}, p objpath.PathNode ) *mg.ValueCastError {
    return mg.NewTypeCastError( asType( expct ), asType( act ), p )
}

func makeIdList( strs ...string ) []*mg.Identifier {
    res := make( []*mg.Identifier, len( strs ) )
    for i, str := range strs { res[ i ] = parser.MustIdentifier( str ) }
    return res
}

var testValBuf1 = mg.Buffer( []byte{ byte( 0 ), byte( 1 ), byte( 2 ) } )
var testValTm1 = parser.MustTimestamp( "2007-08-24T13:15:43.123450000-08:00" )

type rtInit struct { b *mgRct.ReactorTestSetBuilder }

func ( rti *rtInit ) addTests( tests ...mgRct.ReactorTest ) {
    rti.b.AddTests( tests... )
}

func ( rti *rtInit ) addSucc( 
    in, expct interface{}, typ interface{}, dm *DefinitionMap ) {

    rti.addTests(
        &CastReactorTest{ 
            Map: dm,
            In: mg.MustValue( in ), 
            Expect: mg.MustValue( expct ), 
            Type: asType( typ ),
        },
    )
}

func ( rti *rtInit ) addIdent( 
    in interface{}, typ interface{}, dm *DefinitionMap ) {

    v := mg.MustValue( in )
    rti.addSucc( v, v, typ, dm )
}

func ( rti *rtInit ) addVcError(
    in interface{}, typ interface{}, msg string, dm *DefinitionMap ) {

    rti.addTests(
        &CastReactorTest{
            Map: dm,
            In: mg.MustValue( in ),
            Type: asType( typ ),
            Err: newVcErr( nil, msg ),
        },
    )
}

func ( rti *rtInit ) addNullValueError( 
    val interface{}, typ interface{}, dm *DefinitionMap ) {

    rti.addVcError( val, typ, "Value is null", dm )
}

func ( rti *rtInit ) addTcError(
    in interface{}, expct, act interface{}, dm *DefinitionMap ) {

    rti.addTests(
        &CastReactorTest{
            Map: dm,
            In: mg.MustValue( in ),
            Type: asType( expct ),
            Err: newTcErr( expct, act, nil ),
        },
    )
}

func ( rti *rtInit ) addBaseTypeTests() {
    dm := NewDefinitionMap()
    rti.addIdent( mg.Boolean( true ), mg.TypeBoolean, dm )
    rti.addIdent( testValBuf1, mg.TypeBuffer, dm )
    rti.addIdent( "s", mg.TypeString, dm )
    rti.addIdent( mg.Int32( 1 ), mg.TypeInt32, dm )
    rti.addIdent( mg.Int64( 1 ), mg.TypeInt64, dm )
    rti.addIdent( mg.Uint32( 1 ), mg.TypeUint32, dm )
    rti.addIdent( mg.Uint64( 1 ), mg.TypeUint64, dm )
    rti.addIdent( mg.Float32( 1.0 ), mg.TypeFloat32, dm )
    rti.addIdent( mg.Float64( 1.0 ), mg.TypeFloat64, dm )
    rti.addIdent( testValTm1, mg.TypeTimestamp, dm )
    rti.addIdent( nil, mg.TypeNullableValue, dm )
    rti.addSucc( 
        mg.Int32( -1 ), mg.Uint32( uint32( 4294967295 ) ), mg.TypeUint32, dm )
    rti.addSucc( 
        mg.Int64( -1 ), mg.Uint32( uint32( 4294967295 ) ), mg.TypeUint32, dm )
    rti.addSucc( 
        mg.Int32( -1 ), 
        mg.Uint64( uint64( 18446744073709551615 ) ), 
        mg.TypeUint64,
        dm,
    )
    rti.addSucc( 
        mg.Int64( -1 ), 
        mg.Uint64( uint64( 18446744073709551615 ) ), 
        mg.TypeUint64,
        dm,
    )
    rti.addSucc( "true", true, mg.TypeBoolean, dm )
    rti.addSucc( "TRUE", true, mg.TypeBoolean, dm )
    rti.addSucc( "TruE", true, mg.TypeBoolean, dm )
    rti.addSucc( "false", false, mg.TypeBoolean, dm )
    rti.addSucc( true, "true", mg.TypeString, dm )
    rti.addSucc( false, "false", mg.TypeString, dm )
}

func ( rti *rtInit ) addMiscTcErrors() {
    dm := MakeV1DefMap( MakeStructDef( "ns1@v1/S1", nil ) )
    add := func( in interface{}, expct, act interface{} ) {
        rti.addTests(
            &CastReactorTest{
                Map: dm,
                In: mg.MustValue( in ),
                Type: asType( expct ),
                Err: newTcErr( expct, act, nil ),
            },
        )
    }
    add( "s", mg.TypeNull, mg.TypeString )
    add( int32( 1 ), "Buffer", "Int32" )
    add( int32( 1 ), "Buffer?", "Int32" )
    add( true, "Float32", "Boolean" )
    add( true, "&Float32", "Boolean" )
    add( true, "&Float32?", "Boolean" )
    add( true, "Int32", "Boolean" )
    add( true, "&Int32", "Boolean" )
    add( true, "&Int32?", "Boolean" )
    add( mg.MustList( 1, 2 ), mg.TypeString, mg.TypeOpaqueList )
    add( mg.MustList(), "String?", mg.TypeOpaqueList )
    add( "s", "String*", "String" )
    s1 := parser.MustStruct( "ns1@v1/S1" )
    rti.addTests(
        &CastReactorTest{
            Map: dm,
            In: mg.MustList( 1, s1 ),
            Type: asType( "Int32*" ),
            Err: newTcErr( 
                asType( "Int32" ),
                s1.Type.AsAtomicType(),
                objpath.RootedAtList().Next(),
            ),
        },
    )
    rti.addTcError( s1, "&Int32?", s1.Type.AsAtomicType(), dm )
    rti.addTcError( 12, s1.Type.AsAtomicType(), "Int64", dm )
    for _, prim := range mg.PrimitiveTypes {
        // not an err for prims mg.Value and mg.SymbolMap
        if prim != mg.TypeSymbolMap { 
            rti.addTcError( s1, prim, s1.Type.AsAtomicType(), dm )
        }
    }
}

func ( rti *rtInit ) addMiscVcErrors() {
    dm := NewDefinitionMap()
    addErr := func( in interface{}, typ interface{}, err error ) {
        rti.addTests(
            &CastReactorTest{
                Map: dm,
                Type: asType( typ ),
                In: mg.MustValue( in ),
                Err: err,
            },
        )
    }
    add := func( in interface{}, typ interface{}, msg string ) {
        addErr( in, typ, newVcErr( nil, msg ) )
    }
    add( "s", mg.TypeBoolean, `Invalid boolean value: "s"` )
    add( nil, mg.TypeString, "Value is null" )
    add( nil, `String~"a"`, "Value is null" )
    add( mg.MustList(), "String+", "empty list" )
    addErr( 
        mg.MustList( mg.MustList( int32( 1 ), int32( 2 ) ), mg.MustList() ), 
        "Int32+*", 
        newVcErr( objpath.RootedAtList().Next(), "empty list" ),
    )
}

func ( rti *rtInit ) addNonRootPathTestErrors() {
    rti.addTests(
        &CastReactorTest{
            Path: pathInVal,
            Map: NewDefinitionMap(),
            In: mg.MustValue( true ),
            Type: mg.TypeBuffer,
            Err: newTcErr( mg.TypeBuffer, mg.TypeBoolean, pathInVal ),
        },
        &CastReactorTest{
            Path: pathInVal,
            Map: NewDefinitionMap(),
            In: mg.MustList( testValBuf1, true ),
            Type: asType( "Buffer*" ),
            Err: newTcErr( 
                mg.TypeBuffer, mg.TypeBoolean, pathInVal.StartList().Next() ),
        },
    )
}

func ( rti *rtInit ) addStringTests() {
    dm := NewDefinitionMap()
    rti.addIdent( "s", "String?", dm )
    rti.addIdent( "abbbc", `String~"^ab+c$"`, dm )
    rti.addIdent( "abbbc", `String~"^ab+c$"?`, dm )
    rti.addIdent( nil, `String~"^ab+c$"?`, dm )
    rti.addIdent( "", `String~"^a*"?`, dm )
    rti.addIdent( "ab", `String~["aa","ab"]`, dm )
    rti.addIdent( "ab", `String~["aa","ac")`, dm )
    rti.addSucc( 
        mg.MustList( "123", 129 ), 
        mg.MustList( "123", "129" ),
        `String~"^\\d+$"*`,
        dm,
    )
    for _, quant := range []string { "*", "+", "?*", "*?" } {
        val := mg.MustList( "a", "aaaaaa" )
        rti.addSucc( val, val, `String~"^a+$"` + quant, dm )
    }
    rti.addVcError( 
        "ac", 
        `String~"^ab+c$"`,
        `Value "ac" does not satisfy restriction "^ab+c$"`,
        dm,
    )
    rti.addVcError(
        "ab",
        `String~"^a*$"?`,
        "Value \"ab\" does not satisfy restriction \"^a*$\"",
        dm,
    )
    rti.addVcError(
        "ac",
        `String~["aa","ab"]`,
        "Value \"ac\" does not satisfy restriction [\"aa\",\"ab\"]",
        dm,
    )
    rti.addVcError(
        "ac",
        `String~["aa","ac")`,
        "Value \"ac\" does not satisfy restriction [\"aa\",\"ac\")",
        dm,
    )
    rti.addTests(
        &CastReactorTest{
            Map: dm,
            In: mg.MustList( "a", "b" ),
            Type: asType( `String~"^a+$"*` ),
            Path: pathInVal,
            Err: newVcErr(
                pathInVal.StartList().Next(),
                "Value \"b\" does not satisfy restriction \"^a+$\"",
            ),
        },
    )
    rti.addTcError( mg.EmptySymbolMap(), mg.TypeString, mg.TypeSymbolMap, dm )
    rti.addTcError( mg.EmptyList(), mg.TypeString, mg.TypeOpaqueList, dm )
}

func ( rti *rtInit ) addIdentityNumTests() {
    dm := MakeV1DefMap( 
        MakeStructDef( "ns1@v1/S1", nil ),
        MakeEnumDef( "ns1@v1/E1", "e" ),
    )
    rti.addIdent( int64( 1 ), "Int64~[-1,1]", dm )
    rti.addIdent( int64( 1 ), "Int64~(,2)", dm )
    rti.addIdent( int64( 1 ), "Int64~[1,1]", dm )
    rti.addIdent( int64( 1 ), "Int64~[-2, 32)", dm )
    rti.addIdent( int32( 1 ), "Int32~[-2, 32)", dm )
    rti.addIdent( uint32( 3 ), "Uint32~[2,32)", dm )
    rti.addIdent( uint64( 3 ), "Uint64~[2,32)", dm )
    rti.addIdent( mg.Float32( -1.1 ), "Float32~[-2.0,32)", dm )
    rti.addIdent( mg.Float64( -1.1 ), "Float64~[-2.0,32)", dm )
    numTests := []struct{ val mg.Value; str string; typ mg.TypeReference } {
        { val: mg.Int32( 1 ), str: "1", typ: mg.TypeInt32 },
        { val: mg.Int64( 1 ), str: "1", typ: mg.TypeInt64 },
        { val: mg.Uint32( 1 ), str: "1", typ: mg.TypeUint32 },
        { val: mg.Uint64( 1 ), str: "1", typ: mg.TypeUint64 },
        { val: mg.Float32( 1.0 ), str: "1", typ: mg.TypeFloat32 },
        { val: mg.Float64( 1.0 ), str: "1", typ: mg.TypeFloat64 },
    }
    s1 := parser.MustStruct( "ns1@v1/S1" )
    e1 := parser.MustEnum( "ns1@v1/E1", "e" )
    for _, numCtx := range numTests {
        rti.addSucc( numCtx.val, numCtx.str, mg.TypeString, dm )
        rti.addSucc( numCtx.str, numCtx.val, numCtx.typ, dm )
        ptrVal := mg.NewHeapValue( numCtx.val )
        ptrTyp := mg.NewPointerTypeReference( numCtx.typ )
        rti.addSucc( numCtx.val, ptrVal, ptrTyp, dm )
        rti.addSucc( numCtx.str, ptrVal, ptrTyp, dm )
        rti.addSucc( ptrVal, numCtx.str, mg.TypeString, dm )
        rti.addSucc( ptrVal, numCtx.val, numCtx.typ, dm )
        rti.addTcError( mg.EmptySymbolMap(), numCtx.typ, mg.TypeSymbolMap, dm )
        rti.addTcError( mg.EmptySymbolMap(), ptrTyp, mg.TypeSymbolMap, dm )
        rti.addVcError( nil, numCtx.typ, "Value is null", dm )
        rti.addTcError( mg.EmptyList(), numCtx.typ, mg.TypeOpaqueList, dm )
        rti.addTcError( testValBuf1, numCtx.typ, mg.TypeBuffer, dm )
        rti.addTcError( s1, numCtx.typ, s1.Type.AsAtomicType(), dm )
        rti.addTcError( ptrVal, s1.Type.AsAtomicType(), numCtx.typ, dm )
        rti.addTcError( s1, ptrTyp, s1.Type.AsAtomicType(), dm )
        rti.addTcError( e1, numCtx.typ, e1.Type.AsAtomicType(), dm )
        rti.addTcError( ptrVal, e1.Type.AsAtomicType(), numCtx.typ, dm )
        rti.addTcError( e1, ptrTyp, e1.Type.AsAtomicType(), dm )
        for _, valCtx := range numTests {
            rti.addSucc( valCtx.val, numCtx.val, numCtx.typ, dm )
            rti.addSucc( 
                mg.NewHeapValue( valCtx.val ), numCtx.val, numCtx.typ, dm )
            rti.addSucc( valCtx.val, ptrVal, ptrTyp, dm )
            rti.addSucc( mg.NewHeapValue( valCtx.val ), ptrVal, ptrTyp, dm )
        }
    }
}

func ( rti *rtInit ) addTruncateNumTests() {
    dm := NewDefinitionMap()
    posVals := 
        []mg.Value{ mg.Float32( 1.1 ), mg.Float64( 1.1 ), mg.String( "1.1" ) }
    for _, val := range posVals {
        rti.addSucc( val, mg.Int32( 1 ), mg.TypeInt32, dm )
        rti.addSucc( val, mg.Int64( 1 ), mg.TypeInt64, dm )
        rti.addSucc( val, mg.Uint32( 1 ), mg.TypeUint32, dm )
        rti.addSucc( val, mg.Uint64( 1 ), mg.TypeUint64, dm )
    }
    negVals := []mg.Value{ 
        mg.Float32( -1.1 ), mg.Float64( -1.1 ), mg.String( "-1.1" ) }
    for _, val := range negVals {
        rti.addSucc( val, mg.Int32( -1 ), mg.TypeInt32, dm )
        rti.addSucc( val, mg.Int64( -1 ), mg.TypeInt64, dm )
    }
    rti.addSucc( int64( 1 << 31 ), int32( -2147483648 ), mg.TypeInt32, dm )
    rti.addSucc( int64( 1 << 33 ), int32( 0 ), mg.TypeInt32, dm )
    rti.addSucc( int64( 1 << 31 ), uint32( 1 << 31 ), mg.TypeUint32, dm )
}

func ( rti *rtInit ) addNumTests() {
    dm := NewDefinitionMap()
    for _, qn := range mg.NumericTypeNames {
        rti.addVcError( "not-a-num", qn.AsAtomicType(), 
            `invalid number: not-a-num`, dm )
    }
    rti.addIdentityNumTests()
    rti.addTruncateNumTests()
    rti.addSucc( "1", int64( 1 ), "Int64~[-1,1]", dm ) 
    rngErr := func( val string, typ mg.TypeReference ) {
        rti.addVcError( 
            val, typ, fmt.Sprintf( "value out of range: %s", val ), dm )
    }
    rngErr( "2147483648", mg.TypeInt32 )
    rngErr( "-2147483649", mg.TypeInt32 )
    rngErr( "9223372036854775808", mg.TypeInt64 )
    rngErr( "-9223372036854775809", mg.TypeInt64 )
    rngErr( "4294967296", mg.TypeUint32 )
    rti.addVcError( "-1", mg.TypeUint32, "value out of range: -1", dm )
    rti.addVcError( "-1", mg.NewPointerTypeReference( mg.TypeUint32 ), 
        "value out of range: -1", dm )
    rngErr( "18446744073709551616", mg.TypeUint64 )
    rti.addVcError( "-1", mg.TypeUint64, "value out of range: -1", dm )
    for _, tmpl := range []string{ "%s", "&%s", "&%s?" } {
        rti.addVcError(
            12, fmt.Sprintf( tmpl, "Int32~[0,10)" ), 
            "Value 12 does not satisfy restriction [0,10)", dm )
    }
}

func ( rti *rtInit ) addBufferTests() {
    dm := NewDefinitionMap()
    buf1B64 := mg.String( base64.StdEncoding.EncodeToString( testValBuf1 ) )
    rti.addSucc( testValBuf1, buf1B64, mg.TypeString, dm )
    rti.addSucc( mg.NewHeapValue( testValBuf1 ), buf1B64, mg.TypeString, dm )
    rti.addSucc( mg.NewHeapValue( testValBuf1 ), mg.NewHeapValue( buf1B64 ),
        mg.NewPointerTypeReference( mg.TypeString ), dm )
    rti.addSucc( buf1B64, testValBuf1, mg.TypeBuffer , dm )
    rti.addSucc( mg.NewHeapValue( buf1B64 ), testValBuf1, mg.TypeBuffer, dm )
    rti.addSucc( mg.NewHeapValue( buf1B64 ), mg.NewHeapValue( testValBuf1 ),
        mg.NewPointerTypeReference( mg.TypeBuffer ), dm )
    rti.addVcError( "abc$/@", mg.TypeBuffer, 
        "Invalid base64 string: illegal base64 data at input byte 3", dm )
}

func ( rti *rtInit ) addTimeTests() {
    dm := NewDefinitionMap()
    rti.addIdent( mg.Now(),
        `Timestamp~["1970-01-01T00:00:00Z","2200-01-01T00:00:00Z"]`, dm )
    rti.addSucc( testValTm1, testValTm1.Rfc3339Nano(), mg.TypeString, dm )
    rti.addSucc( testValTm1.Rfc3339Nano(), testValTm1, mg.TypeTimestamp, dm )
    rti.addVcError(
        parser.MustTimestamp( "2012-01-01T00:00:00Z" ),
        `mingle:core@v1/Timestamp~` +
            `["2000-01-01T00:00:00Z","2001-01-01T00:00:00Z"]`,
        "Value 2012-01-01T00:00:00Z does not satisfy restriction " +
            "[\"2000-01-01T00:00:00Z\",\"2001-01-01T00:00:00Z\"]",
        dm,
    )
}

func ( rti *rtInit ) addNullableTests() {
    dm := MakeV1DefMap( MakeStructDef( "ns1@v1/S1", nil ) )
    typs := []mg.TypeReference{}
    addNullSucc := func( expct interface{}, typ mg.TypeReference ) {
        rti.addSucc( nil, expct, typ, dm )
    }
    for _, prim := range mg.PrimitiveTypes {
        if mg.IsNullableType( prim ) {
            typs = append( typs, mg.MustNullableTypeReference( prim ) )
        } else {
            rti.addNullValueError( nil, prim, dm )
        }
    }
    typs = append( typs,
        asType( "&Null?" ),
        asType( "String?" ),
        asType( "String*?" ),
        asType( "&Int32?*?" ),
        asType( "String+?" ),
        asType( "&ns1@v1/T?" ),
        asType( "ns1@v1/T*?" ),
    )
    for _, typ := range typs { addNullSucc( nil, typ ) }
}

func ( rti *rtInit ) addListTests() {
    dm := MakeV1DefMap(
        MakeStructDef( "ns1@v1/S1", 
            []*FieldDefinition{ MakeFieldDef( "f1", "&Int32?", nil ) } ),
    )
    for _, quant := range []string{ "*", "**", "***" } {
        rti.addSucc( []interface{}{}, mg.MustList(), "Int64" + quant, dm )
    }
    for _, quant := range []string{ "**", "*+" } {
        v := mg.MustList( mg.MustList(), mg.MustList() )
        rti.addIdent( v, "Int64" + quant, dm )
    }
    // test conversions in a deeply nested list
    rti.addSucc(
        []interface{}{
            []interface{}{ "1", int32( 1 ), int64( 1 ) },
            []interface{}{ float32( 1.0 ), float64( 1.0 ) },
            []interface{}{},
        },
        mg.MustList(
            mg.MustList( mg.Int64( 1 ), mg.Int64( 1 ), mg.Int64( 1 ) ),
            mg.MustList( mg.Int64( 1 ), mg.Int64( 1 ) ),
            mg.MustList(),
        ),
        "Int64**",
        dm,
    )
    rti.addSucc(
        []interface{}{ int64( 1 ), nil, "hi" },
        mg.MustList( "1", nil, "hi" ),
        "String?*",
        dm,
    )
    s1 := parser.MustStruct( "ns1@v1/S1" )
    rti.addSucc(
        []interface{}{ s1, mg.NewHeapValue( s1 ), nil },
        mg.MustList( mg.NewHeapValue( s1 ), mg.NewHeapValue( s1 ), mg.NullVal ),
        "&ns1@v1/S1?*",
        dm,
    )
    rti.addTests(
        &CastReactorTest{
            Map: dm,
            In: mg.MustValue( []interface{}{ mg.NewHeapValue( s1 ), nil } ),
            Type: asType( "&ns1@v1/S1*" ),
            Err: newVcErr( 
                objpath.RootedAtList().SetIndex( 1 ), "Value is null" ),
        },
        &CastReactorTest{
            Map: dm,
            In: mg.MustValue( []interface{}{ s1, nil } ),
            Type: asType( "ns1@v1/S1*" ),
            Err: newVcErr( 
                objpath.RootedAtList().SetIndex( 1 ), "Value is null" ),
        },
    )
    rti.addSucc(
        []interface{}{ 
            int32( 1 ), 
            []interface{}{}, 
            []interface{}{ int32( 1 ), int32( 2 ), int32( 3 ) },
            "s1", 
            s1, 
            nil,
        },
        mg.MustList(
            mg.NewHeapValue( mg.Int32( 1 ) ),
            mg.NewHeapValue( mg.MustList() ),
            mg.NewHeapValue( 
                mg.MustList( int32( 1 ), int32( 2 ), int32( 3 ) ) ),
            mg.NewHeapValue( mg.String( "s1" ) ),
            mg.NewHeapValue( s1 ),
            mg.NullVal,
        ),
        "&Value?*",
        dm,
    )
    rti.addSucc( mg.MustList(), mg.MustList(), mg.TypeValue, dm )
    intList1 := mg.MustList( int32( 1 ), int32( 2 ), int32( 3 ) )
    rti.addSucc( intList1, intList1, mg.TypeValue, dm )
    rti.addSucc( intList1, intList1, mg.TypeOpaqueList, dm )
    rti.addSucc( intList1, intList1, "Int32*?", dm )
    rti.addSucc( 
        mg.MustList(), 
        mg.NewHeapValue( mg.MustList() ), 
        mg.NewPointerTypeReference( asType( "&Int32*" ) ),
        dm,
    )
    rti.addSucc( 
        mg.NewHeapValue( mg.MustList() ), 
        mg.NewHeapValue( mg.MustList() ),
        mg.NewPointerTypeReference( asType( "&Int32*" ) ),
        dm,
    )
    rti.addSucc( 
        mg.NewHeapValue( mg.MustList() ), mg.MustList(), "&Int32*", dm )
    rti.addSucc( nil, mg.NullVal, "Int32*?", dm )
    rti.addNullValueError( nil, "Int32*", dm )
    rti.addNullValueError( nil, "Int32+", dm )
    rti.addNullValueError( mg.NewHeapValue( mg.NullVal ), "Int32+", dm )
    rti.addVcError( 
        mg.NewHeapValue( mg.MustList() ), "&Int32+", "empty list", dm )
    rti.addSucc( 
        nil, 
        mg.NullVal,
        mg.MustNullableTypeReference( asType( "&Int32*" ) ),
        dm,
    )
    rti.addSucc( 
        mg.NewHeapValue( mg.NullVal ), 
        mg.NullVal,
        mg.MustNullableTypeReference( asType( "&Int32*" ) ),
        dm,
    )
}

func ( rti *rtInit ) addMapTests() {
    dm := MakeV1DefMap(
        MakeStructDef( "ns1@v1/S1", 
            []*FieldDefinition{ MakeFieldDef( "f1", "Int32", nil ) },
        ),
    )
    m1 := mg.MustSymbolMap
    m2 := func() *mg.SymbolMap { 
        return parser.MustSymbolMap( "f1", int32( 1 ) )
    }
    rti.addSucc( m1(), m1(), mg.TypeSymbolMap, dm )
    rti.addSucc( m1(), m1(), mg.TypeValue, dm )
    rti.addSucc( m2(), m2(), mg.TypeSymbolMap, dm )
    rti.addSucc( m2(), m2(), "SymbolMap?", dm )
    s2 := &mg.Struct{ Type: mkQn( "ns2@v1/S1" ), Fields: m2() }
    rti.addSucc( s2, m2(), mg.TypeSymbolMap, dm )
    l1 := mg.MustList()
    l2 := mg.MustList( m1(), m2() )
    lt1 := asType( "SymbolMap*" )
    lt2 := asType( "SymbolMap+" )
    rti.addSucc( l1, l1, lt1, dm )
    rti.addSucc( l2, l2, lt2, dm )
    rti.addSucc(
        parser.MustSymbolMap( "f1", mg.NullVal ), 
        parser.MustSymbolMap( "f1", mg.NullVal ), 
        mg.TypeValue,
        dm,
    )
    rti.addSucc( mg.MustList( s2, s2 ), mg.MustList( m2(), m2() ), lt2, dm )
    rti.addTcError( int32( 1 ), mg.TypeSymbolMap, mg.TypeInt32, dm )
    rti.addTcError(
        mg.MustList( m1(), int32( 1 ) ),
        mg.TypeSymbolMap,
        mg.TypeOpaqueList,
        dm,
    )
    nester := 
        parser.MustSymbolMap( "f1", parser.MustSymbolMap( "f2", int32( 1 ) ) )
    rti.addSucc( nester, nester, mg.TypeSymbolMap, dm )
    rti.addSucc( m1(), mg.NewHeapValue( m1() ), "&SymbolMap", dm )
    rti.addSucc( 
        mg.NewHeapValue( m1() ), mg.NewHeapValue( m1() ), "&SymbolMap", dm )
    rti.addSucc( mg.NewHeapValue( m1() ), m1(), "SymbolMap", dm )
    rti.addSucc( nil, mg.NullVal, "SymbolMap?", dm )
    rti.addSucc( nil, mg.NullVal, "&SymbolMap?", dm )
    rti.addSucc( 
        mg.NewHeapValue( mg.NullVal ), 
        mg.NewHeapValue( mg.NullVal ),
        mg.NewPointerTypeReference( asType( "SymbolMap?" ) ),
        dm,
    )
    rti.addNullValueError( nil, "SymbolMap", dm )
    rti.addNullValueError( nil, "&SymbolMap", dm )
    rti.addNullValueError( mg.NewHeapValue( mg.NullVal ), "&SymbolMap", dm )
}

func ( rti *rtInit ) addBaseFieldCastTests() {
    p := mg.MakeTestIdPath
    qn1Str := "ns1@v1/S1"
    qn1 := parser.MustQualifiedTypeName( qn1Str )
    s1F1 := func( val interface{} ) *mg.Struct {
        return parser.MustStruct( qn1, "f1", val )
    }
    s1DefMap := func( typ string ) *DefinitionMap {
        fld := MakeFieldDef( "f1", typ, nil )
        return MakeV1DefMap( 
            MakeStructDef( qn1Str, []*FieldDefinition{ fld } ) )
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
        return newTcErr( expct, act, p( 1 ) )
    }
    s1F1Succ( int32( 1 ), int32( 1 ), "Int32" )
    s1F1Succ( "1", int32( 1 ), "Int32" )
    s1F1Succ( int32( 1 ), int32( 1 ), "Value" )
    i32L1 := mg.MustList( int32( 1 ), int32( 2 ), int32( 3 ) )
    s1F1Succ( i32L1, i32L1, "Int32+" )
    s1F1Succ( i32L1, i32L1, "Value" )
    s1F1Succ( i32L1, i32L1, "Value*" )
    s1F1Succ( mg.MustList( "1", int64( 2 ), int32( 3 ) ), i32L1, "Int32*" )
    sm1 := parser.MustSymbolMap( "f1", int32( 1 ) )
    s1F1Succ( sm1, sm1, "SymbolMap" )
    s1F1Succ( sm1, sm1, "Value" )
    s1F1Succ( int32( 1 ), mustHeapVal( int32( 1 ) ), "&Int32?" )
    s1F1Succ( mg.NullVal, mg.NullVal, "&Int32?" )
    s1F1Succ(
        mg.MustList( "1", nil, int64( 1 ) ),
        mg.MustList( 
            mustHeapVal( int32( 1 ) ), mg.NullVal, mustHeapVal( int32( 1 ) ) ),
        "&Int32?*",
    )
    s1F1Fail( []byte{}, "Int32", tcErr1( mg.TypeInt32, mg.TypeBuffer ) )
    s1F1Fail( 
        mg.MustList( 1, 2 ), 
        "Int32", 
        tcErr1( mg.TypeInt32, mg.TypeOpaqueList ),
    )
    s1F1Fail( nil, "Int32", newVcErr( p( 1 ), "Value is null" ) )
    s1F1Fail( int32( 1 ), "Int32+", tcErr1( "Int32+", "Int32" ) )
    s1F1Fail( mg.MustList(), "Int32+", newVcErr( p( 1 ), "empty list" ) )
    s1F1Fail( 
        mg.MustList( []byte{} ), 
        "Int32*", 
        newTcErr( "Int32", "Buffer", p( 1, "0" ) ),
    )
    s1F1Fail( int32( 1 ), "SymbolMap", tcErr1( "SymbolMap", "Int32" ) )
    s1F1Fail( i32L1, "SymbolMap", tcErr1( "SymbolMap", mg.TypeOpaqueList ) )
}

func ( rti *rtInit ) addFieldSetCastTests() {
    id := parser.MustIdentifier
    mkId := mg.MakeTestId
    p := mg.MakeTestIdPath
    dm := MakeV1DefMap(
        MakeStructDef(
            "ns1@v1/S1",
            []*FieldDefinition{ MakeFieldDef( "f1", "Int32", nil ) },
        ),
        MakeStructDef(
            "ns1@v1/S2",
            []*FieldDefinition{ 
                MakeFieldDef( "f1", "Int32", nil ),
                MakeFieldDef( "f2", "Int32", nil ),
            },
        ),
        MakeStructDef(
            "ns1@v1/S3",
            []*FieldDefinition{ MakeFieldDef( "f1", "&Int32?", nil ) },
        ),
        MakeStructDef(
            "ns1@v1/S4",
            []*FieldDefinition{ MakeFieldDef( "f1", "&ns1@v1/S1?", nil ) },
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
    addSucc( parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ) )
    addSucc( 
        parser.MustStruct( "ns1@v1/S2", "f1", int32( 1 ), "f2", int32( 2 ) ) )
    addSucc( parser.MustStruct( "ns1@v1/S3" ) )
    addSucc( parser.MustStruct( "ns1@v1/S3", "f1", mustHeapVal( int32( 1 ) ) ) )
    s1Inst1 := mustHeapVal( parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ) )
    addSucc( parser.MustStruct( "ns1@v1/S4", "f1", s1Inst1 ) )
    addFail(
        parser.MustStruct( "ns1@v1/S1" ),
        mg.NewMissingFieldsError( nil, makeIdList( "f1" ) ),
    )
    addFail(
        parser.MustStruct( "ns1@v1/S2", "f1", int32( 1 ) ),
        mg.NewMissingFieldsError( nil, makeIdList( "f2" ) ),
    )
    addFail(
        parser.MustStruct( "ns1@v1/S2" ),
        mg.NewMissingFieldsError( nil, makeIdList( "f1", "f2" ) ),
    )
    addFail(
        parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ), "f2", int32( 2 ) ),
        mg.NewUnrecognizedFieldError( nil, mkId( 2 ) ),
    )
    addFail(
        parser.MustStruct( "ns1@v1/S4",
            "f1", parser.MustStruct( "ns1@v1/S1", "not-a-field", int32( 1 ) ) ),
        mg.NewUnrecognizedFieldError( p( 1 ), id( "not-a-field" ) ),
    )
    for _, i := range []string{ "1", "2" } {
        addFail(
            parser.MustStruct( "ns1@v1/S" + i, "f3", int32( 3 ) ),
            mg.NewUnrecognizedFieldError( nil, mkId( 3 ) ),
        )
    }
}

func ( rti *rtInit ) addStructValCastTests() {
    dm := MakeV1DefMap(
        MakeStructDef( "ns1@v1/S1",
            []*FieldDefinition{ MakeFieldDef( "f1", "Int32", nil ) } ),
        MakeStructDef( "ns1@v1/S2", nil ),
        MakeEnumDef( "ns1@v1/E1", "e" ),
    )
    t1 := asType( "ns1@v1/S1" )
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
    tcErr1 := func( act interface{} ) error {
        return newTcErr( "ns1@v1/S1", act, nil )
    }
    addFail( parser.MustStruct( "ns1@v1/S2" ), tcErr1( "ns1@v1/S2" ) )
    addFail( int32( 1 ), tcErr1( "Int32" ) )
    addFail( parser.MustEnum( "ns1@v1/E1", "e" ), tcErr1( "ns1@v1/E1" ) )
    addFail( 
        parser.MustEnum( "ns1@v1/S1", "e" ), 
        newVcErr( nil, "not an enum type: ns1@v1/S1" ),
    )
    addFail( mg.MustList(), tcErr1( mg.TypeOpaqueList ) )
    s1 := parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) )
    s2 := parser.MustStruct( "ns1@v1/S2" )
    rti.addSucc( s1, s1, mg.TypeValue, dm )
    rti.addSucc( s1, s1, "ns1@v1/S1", dm )
    rti.addSucc( s2, s2, "ns1@v1/S2", dm )
    rti.addSucc( s1, mg.NewHeapValue( s1 ), "&ns1@v1/S1?", dm )
    l1 := mg.MustList( s1, s1 )
    rti.addSucc( l1, l1, &mg.ListTypeReference{ t1, false }, dm )
    rti.addSucc( l1, l1, &mg.ListTypeReference{ t1, true }, dm )
    rti.addTcError( int32( 1 ), s1.Type.AsAtomicType(), mg.TypeInt32, dm )
    rti.addSucc( 
        mg.NewHeapValue( s1 ), mg.NewHeapValue( s1 ), "&ns1@v1/S1", dm )
    rti.addSucc( s1, mg.NewHeapValue( s1 ), "&ns1@v1/S1", dm )
    rti.addSucc( mg.NewHeapValue( s1 ), s1, "ns1@v1/S1", dm )
    rti.addSucc( nil, mg.NullVal, "&ns1@v1/S1?", dm )
    rti.addNullValueError( nil, "&ns1@v1/S1", dm )
    rti.addNullValueError( mg.NewHeapValue( mg.NullVal ), "&ns1@v1/S1", dm )
    rti.addTcError( s1, "ns1@v1/S2", "ns1@v1/S1", dm )
    rti.addTcError( mg.NewHeapValue( s1 ), "ns1@v1/S2", "ns1@v1/S1", dm )
    rti.addTcError( s1, "ns1@v1/S2", "ns1@v1/S1", dm )
    rti.addTcError( mg.NewHeapValue( s1 ), "ns1@v1/S2", "ns1@v1/S1", dm )
}

func ( rti *rtInit ) addInferredStructCastTests() {
    dm := MakeV1DefMap(
        MakeStructDef(
            "ns1@v1/S1",
            []*FieldDefinition{
                MakeFieldDef( "f1", "&Int32?", nil ),
                MakeFieldDef( "f2", "&ns1@v1/S2?", nil ),
            },
        ),
        MakeStructDef(
            "ns1@v1/S2",
            []*FieldDefinition{ MakeFieldDef( "f1", "Int32", nil ) },
        ),
    )
    addSucc := func( in, expct mg.Value ) {
        rti.addTests(
            &CastReactorTest{
                Map: dm,
                Type: asType( "ns1@v1/S1" ),
                In: in,
                Expect: expct,
            },
        )
    }
    i1HeapVal := mustHeapVal( int32( 1 ) )
    addSucc( 
        parser.MustSymbolMap( "f1", int32( 1 ) ),
        parser.MustStruct( "ns1@v1/S1", "f1", i1HeapVal ),
    )
    s2HeapVal := 
        mustHeapVal( parser.MustStruct( "ns1@v1/S2", "f1", int32( 1 ) ) )
    addSucc( 
        parser.MustSymbolMap( "f2", parser.MustSymbolMap( "f1", i1HeapVal ) ),
        parser.MustStruct( "ns1@v1/S1", "f2", s2HeapVal ),
    )
}

func ( rti *rtInit ) addStructTests() {
    rti.addStructValCastTests() 
    rti.addInferredStructCastTests()
}

func ( rti *rtInit ) addSchemaCastTests() {
    schema1Nil := &mg.NullableTypeReference{ mkTyp( "ns1@v1/Schema1" ) }
    mgId := parser.MustIdentifier
    dm := MakeV1DefMap(
        MakeSchemaDef( 
            "ns1@v1/Schema1", 
            []*FieldDefinition{
                MakeFieldDef( "f1", "Int32", nil ),
            },
        ),
        MakeSchemaDef(
            "ns1@v1/Schema2",
            []*FieldDefinition{
                MakeFieldDef( "f1", "ns1@v1/Schema1", nil ),
                MakeFieldDef( "f2", schema1Nil, nil ),
            },
        ),
        MakeStructDef(
            "ns1@v1/S1",
            []*FieldDefinition{
                MakeFieldDef( "f1", "Int32", nil ),
                MakeFieldDef( "f2", "Int32", nil ),
            },
        ),
        MakeStructDef(
            "ns1@v1/S2",
            []*FieldDefinition{
                MakeFieldDef( "f1", "ns1@v1/Schema1", nil ),
                MakeFieldDef( "f2", schema1Nil, nil ),
            },
        ),
        MakeStructDef(
            "ns1@v1/S3",
            []*FieldDefinition{
                MakeFieldDef( "f1", "Int32", nil ),
                MakeFieldDef( "f2", "Int64", nil ),
            },
        ),
    )
    addSucc := func( in, expct mg.Value, typ interface{} ) {
        rti.addTests(
            &CastReactorTest{
                Map: dm,
                Type: asType( typ ),
                In: in,
                Expect: expct,
            },
        )
    }
    addSucc( 
        parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ), "f2", int32( 1 ) ),
        parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ), "f2", int32( 1 ) ),
        "ns1@v1/Schema1",
    )
    addSucc( 
        parser.MustStruct( "ns1@v1/S1", "f1", int64( 1 ), "f2", int64( 1 ) ),
        parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ), "f2", int32( 1 ) ),
        "ns1@v1/Schema1",
    )
    addSucc( 
        parser.MustSymbolMap( "f1", int32( 1 ) ),
        parser.MustSymbolMap( "f1", int32( 1 ) ),
        "ns1@v1/Schema1",
    )
    addSucc( 
        parser.MustSymbolMap( "f1", int32( 1 ), "f2", int32( 1 ) ),
        parser.MustSymbolMap( "f1", int32( 1 ), "f2", int32( 1 ) ),
        "ns1@v1/Schema1",
    )
    addSucc( 
        parser.MustSymbolMap( "f1", int64( 1 ), "f2", int64( 1 ) ),
        parser.MustSymbolMap( "f1", int32( 1 ), "f2", int64( 1 ) ),
        "ns1@v1/Schema1",
    )
    addSucc( 
        mg.NewHeapValue( parser.MustSymbolMap( "f1", int64( 1 ) ) ),
        mg.NewHeapValue( parser.MustSymbolMap( "f1", int32( 1 ) ) ),
        "ns1@v1/Schema1",
    )
    addSucc( 
        mg.NewHeapValue(
            parser.MustStruct( "ns1@v1/S1", 
                "f1", int32( 1 ), 
                "f2", int32( 1 ),
            ),
        ),
        mg.NewHeapValue(
            parser.MustStruct( "ns1@v1/S1", 
                "f1", int32( 1 ), 
                "f2", int32( 1 ),
            ),
        ),
        schema1Nil,
    )
    addSucc( 
        parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ), "f2", int32( 1 ) ),
        parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ), "f2", int32( 1 ) ),
        schema1Nil,
    )
    addSucc( mg.NullVal, mg.NullVal, schema1Nil )
    addSucc(
        parser.MustStruct( "ns1@v1/S2",
            "f1", parser.MustStruct( "ns1@v1/S1", 
                "f1", int32( 1 ), 
                "f2", int32( 1 ),
            ),
            "f2", mg.NewHeapValue(
                parser.MustStruct( "ns1@v1/S1", 
                    "f1", int32( 1 ), 
                    "f2", int32( 1 ),
                ),
            ),
        ),
        parser.MustStruct( "ns1@v1/S2",
            "f1", parser.MustStruct( "ns1@v1/S1", 
                "f1", int32( 1 ), 
                "f2", int32( 1 ),
            ),
            "f2", mg.NewHeapValue(
                parser.MustStruct( "ns1@v1/S1", 
                    "f1", int32( 1 ), 
                    "f2", int32( 1 ),
                ),
            ),
        ),
        "ns1@v1/Schema2",
    )
    addSucc(
        parser.MustStruct( "ns1@v1/S2",
            "f1", parser.MustStruct( "ns1@v1/S1", 
                "f1", int32( 1 ), 
                "f2", int32( 1 ),
            ),
            "f2", parser.MustStruct( "ns1@v1/S1", 
                "f1", int32( 1 ), 
                "f2", int32( 1 ),
            ),
        ),
        parser.MustStruct( "ns1@v1/S2",
            "f1", parser.MustStruct( "ns1@v1/S1", 
                "f1", int32( 1 ), 
                "f2", int32( 1 ),
            ),
            "f2", parser.MustStruct( "ns1@v1/S1", 
                "f1", int32( 1 ), 
                "f2", int32( 1 ),
            ),
        ),
        "ns1@v1/Schema2",
    )
    addSucc(
        parser.MustSymbolMap(
            "f1", parser.MustSymbolMap( "f1", int32( 1 ), "f2", int32( 1 ) ),
            "f2", parser.MustSymbolMap( "f1", int32( 1 ), "f2", int32( 1 ) ),
        ),
        parser.MustSymbolMap(
            "f1", parser.MustSymbolMap( "f1", int32( 1 ), "f2", int32( 1 ) ),
            "f2", parser.MustSymbolMap( "f1", int32( 1 ), "f2", int32( 1 ) ),
        ),
        "ns1@v1/Schema2",
    )
    addSucc(
        parser.MustSymbolMap(
            "f1", parser.MustSymbolMap( "f1", int64( 1 ), "f2", int64( 1 ) ),
            "f2", parser.MustSymbolMap( "f1", int64( 1 ), "f2", int64( 1 ) ),
        ),
        parser.MustSymbolMap(
            "f1", parser.MustSymbolMap( "f1", int32( 1 ), "f2", int64( 1 ) ),
            "f2", parser.MustSymbolMap( "f1", int32( 1 ), "f2", int64( 1 ) ),
        ),
        "ns1@v1/Schema2",
    )
    addSucc(
        mg.MustList( 
            parser.MustSymbolMap( "f1", int32( 1 ) ),
            parser.MustSymbolMap( "f1", int64( 1 ) ),
        ),
        mg.MustList( 
            parser.MustSymbolMap( "f1", int32( 1 ) ),
            parser.MustSymbolMap( "f1", int32( 1 ) ),
        ),
        "ns1@v1/Schema1*",
    )
    addSucc(
        mg.MustList(
            parser.MustSymbolMap( "f1", int32( 1 ) ),
            parser.MustStruct( "ns1@v1/S1", 
                "f1", int32( 1 ), "f2", int32( 2 ) ),
            parser.MustStruct( "ns1@v1/S3", 
                "f1", int32( 1 ), "f2", int64( 2 ) ),
        ),
        mg.MustList(
            parser.MustSymbolMap( "f1", int32( 1 ) ),
            parser.MustStruct( "ns1@v1/S1", 
                "f1", int32( 1 ), "f2", int32( 2 ) ),
            parser.MustStruct( "ns1@v1/S3", 
                "f1", int32( 1 ), "f2", int64( 2 ) ),
        ),
        "ns1@v1/Schema1*",
    )
    addFail := func( in mg.Value, typ interface{}, err error ) {
        rti.addTests(
            &CastReactorTest{ Map: dm, In: in, Type: asType( typ ), Err: err },
        )
    }
    addFail(
        parser.MustSymbolMap(),
        "ns1@v1/Schema1",
        mg.NewMissingFieldsError( nil, []*mg.Identifier{ mgId( "f1" ) } ),
    )
    addFail(
        parser.MustSymbolMap(
            "f1", parser.MustSymbolMap( "f1", int32( 1 ), "f2", int32( 1 ) ),
            "f2", parser.MustSymbolMap( "f2", int32( 1 ) ),
        ),
        "ns1@v1/Schema2",
        mg.NewMissingFieldsError(
            mg.MakeTestIdPath( 2 ),
            []*mg.Identifier{ mgId( "f1" ) },
        ),
    )
    addFail( 
        mg.Int32( int32( 1 ) ),
        "ns1@v1/Schema1",
        newTcErr( "ns1@v1/Schema1", mg.TypeInt32, nil ),
    )
    addFail( 
        parser.MustStruct( "ns1@v1/S3", "f1", int32( 1 ), "f2", int64( 1 ) ),
        "ns1@v1/Schema2",
        newTcErr( "ns1@v1/Schema2", "ns1@v1/S3", nil ),
    )
    addFail(
        mg.MustList(
            parser.MustStruct( "ns1@v1/S2",
                "f1", parser.MustStruct( "ns1@v1/S1", 
                    "f1", int32( 1 ), 
                    "f2", int32( 1 ),
                ),
            ),
            parser.MustStruct( "ns1@v1/S3", 
                "f1", int32( 1 ), "f2", int64( 1 ),
            ),
        ),
        "ns1@v1/Schema2*",
        newTcErr( "ns1@v1/Schema2", "ns1@v1/S3", mg.MakeTestIdPath( "1" ) ),
    )
    addFail(
        parser.MustSymbolMap(
            "f1", parser.MustSymbolMap( "f1", int32( 1 ), "f2", int32( 1 ) ),
            "f2", int32( 1 ),
        ),
        "ns1@v1/Schema2",
        newTcErr( "ns1@v1/Schema1", mg.TypeInt32, mg.MakeTestIdPath( 2 ) ),
    )
}

func ( rti *rtInit ) addEnumValCastTests() {
    dm := MakeV1DefMap(
        MakeStructDef( "ns1@v1/S1", []*FieldDefinition{} ),
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
    e1 := parser.MustEnum( "ns1@v1/E1", "c1" )
    for _, t := range []struct{ typStr string; expct mg.Value }{
        { typStr: "ns1@v1/E1", expct: e1 },
        { typStr: "&ns1@v1/E1?", expct: mustHeapVal( e1 ) },
    } {
        addSucc( e1, t.expct, t.typStr )
        addSucc( "c1", t.expct, t.typStr )
        addFail( 
            int32( 1 ), 
            t.typStr, 
            newTcErr( "ns1@v1/E1", mg.TypeInt32, nil ),
        )
    }
    addSucc( mg.MustList( e1, "c1" ), mg.MustList( e1, e1 ), "ns1@v1/E1*" )
    addSucc( mg.NullVal, mg.NullVal, "&ns1@v1/E1?" )
    vcErr := func( msg string ) error { return newVcErr( nil, msg ) }
    for _, in := range []interface{} { 
        "c3", parser.MustEnum( "ns1@v1/E1", "c3" ),
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
        parser.MustEnum( "ns1@v1/E2", "c1" ), 
        "ns1@v1/E1", 
        newTcErr( "ns1@v1/E1", "ns1@v1/E2", nil ),
    )
    addFail( 
        parser.MustStruct( "ns1@v1/E1" ), 
        "ns1@v1/E1", 
        vcErr( "not a type with fields: ns1@v1/E1" ),
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
        parser.MustSymbolMap(), 
        "ns1@v1/E1", 
        newTcErr( "ns1@v1/E1", mg.TypeSymbolMap, nil ),
    )
    // even though S2 not in type map, we still expect an upstream type cast
    // error
    addFail( 
        parser.MustStruct( "ns1@v1/S2" ), 
        "ns1@v1/E1", 
        newTcErr( "ns1@v1/E1", "ns1@v1/S2", nil ),
    )
}

// Just coverage that structs and defined types don't cause things to go nutso
// when nested in various ways
func ( rti *rtInit ) addDeepCatchallTests() {
    dm := MakeV1DefMap(
        MakeStructDef( "ns1@v1/S1",
            []*FieldDefinition{
                MakeFieldDef( "f1", "Int32", int32( 1 ) ),
            },
        ),
        MakeStructDef( "ns1@v1/S2",
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
    in := parser.MustStruct( "ns1@v1/S2",
        "f1", parser.MustStruct( "ns1@v1/S1", "f1", int32( 2 ) ),
        "f2", "e1",
        "f3", mg.MustList( 
            parser.MustSymbolMap(),
            parser.MustSymbolMap( "f1", int32( 2 ) ),
        ),
        "f4", mg.MustList( parser.MustEnum( "ns1@v1/E1", "e1" ), "e2" ),
        "f5", parser.MustSymbolMap(
            "f1", int32( 1 ),
            "f2", parser.MustEnum( "ns1@v1/E1", "e1" ),
            "f3", parser.MustStruct( "ns1@v1/S1", "f1", int32( 3 ) ),
        ),
        "f6", parser.MustStruct( "ns1@v1/S1" ),
        "f7", parser.MustEnum( "ns1@v1/E1", "e1" ),
        "f8", mg.MustList(
            parser.MustStruct( "ns1@v1/S1", "f1", int32( 2 ) ),
            parser.MustEnum( "ns1@v1/E1", "e2" ),
        ),
    )
    expct := parser.MustStruct( "ns1@v1/S2",
        "f1", parser.MustStruct( "ns1@v1/S1", "f1", int32( 2 ) ),
        "f2", parser.MustEnum( "ns1@v1/E1", "e1" ),
        "f3", mg.MustList(
            parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
            parser.MustStruct( "ns1@v1/S1", "f1", int32( 2 ) ),
        ),
        "f4", mg.MustList(
            parser.MustEnum( "ns1@v1/E1", "e1" ),
            parser.MustEnum( "ns1@v1/E1", "e2" ),
        ),
        "f5", parser.MustSymbolMap(
            "f1", int32( 1 ),
            "f2", parser.MustEnum( "ns1@v1/E1", "e1" ),
            "f3", parser.MustStruct( "ns1@v1/S1", "f1", int32( 3 ) ),
        ),
        "f6", parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
        "f7", parser.MustEnum( "ns1@v1/E1", "e1" ),
        "f8", mg.MustList(
            parser.MustStruct( "ns1@v1/S1", "f1", int32( 2 ) ),
            parser.MustEnum( "ns1@v1/E1", "e2" ),
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
    p := mg.MakeTestIdPath
    qn1 := parser.MustQualifiedTypeName( "ns1@v1/S1" )
    e1 := parser.MustEnum( "ns1@v1/E1", "c1" )
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
        MakeStructDef( "ns1@v1/S1", s1Flds ),
        MakeStructDef( "ns1@v1/S2",
            []*FieldDefinition{
                MakeFieldDef( "f1", "Int32", nil ),
                MakeFieldDef( "f2", "Int32", mg.Int32( int32( 1 ) ) ),
            },
        ),
        MakeStructDef( "ns1@v1/S3",
            []*FieldDefinition{ MakeFieldDef( "f1", "Int32*", nil ) } ),
        MakeStructDef( "ns1@v1/S4",
            []*FieldDefinition{ 
                MakeFieldDef( "f1", "ns1@v1/S1", nil ),
                MakeFieldDef( "f2", "Int32", int32( 1 ) ),
            },
        ),
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
    noDefls := parser.MustStruct( qn1,
        "f1", int32( 1 ), 
        "f2", "s1", 
        "f3", parser.MustEnum( "ns1@v1/E1", "c2" ), 
        "f4", mg.MustList( int32( 1 ) ),
        "f5", false,
    )
    allDefls := parser.MustStruct( qn1, deflPairs... )
    addSucc1( noDefls, noDefls )
    addSucc1( parser.MustStruct( qn1 ), allDefls )
    addSucc1( parser.MustSymbolMap(), allDefls )
    // Added as a regression for a bug that prevented proper handling of
    // defaults in a struct which itself had struct values
    addSucc(
        parser.MustSymbolMap( "f1", parser.MustStruct( qn1 ) ),
        parser.MustStruct( "ns1@v1/S4", "f1", allDefls, "f2", int32( 1 ) ),
        "ns1@v1/S4",
    )
    f1OnlyPairs := 
        append( []interface{}{ "f1", int32( 1 ) }, deflPairs[ 2 : ]... )
    addSucc1( 
        parser.MustSymbolMap( "f1", int32( 1 ) ),
        parser.MustStruct( qn1, f1OnlyPairs... ),
    )
    addSucc(
        parser.MustStruct( "ns1@v1/S2", "f1", int32( 1 ) ),
        parser.MustStruct( "ns1@v1/S2", "f1", int32( 1 ), "f2", int32( 1 ) ),
        "ns1@v1/S2",
    )
    addSucc(
        parser.MustStruct( "ns1@v1/S2", "f1", int32( 1 ), "f2", int32( -1 ) ),
        parser.MustStruct( "ns1@v1/S2", "f1", int32( 1 ), "f2", int32( -1 ) ),
        "ns1@v1/S2",
    )
    s3Inst1 :=
        parser.MustStruct( "ns1@v1/S3", 
            "f1", mg.MustList( int32( 1 ), int32( 2 ) ) )
    addSucc( s3Inst1, s3Inst1, "ns1@v1/S3" )
    addSucc( 
        parser.MustStruct( "ns1@v1/S3" ),
        parser.MustStruct( "ns1@v1/S3", "f1", mg.MustList() ),
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
        parser.MustSymbolMap( "f1", []byte{} ),
        "ns1@v1/S1",
        newTcErr( mg.TypeInt32, mg.TypeBuffer, p( 1 ) ),
    )
}

func ( rti *rtInit ) addDefaultPathTests() {
    dm := MakeV1DefMap(
        MakeStructDef( "ns1@v1/S1",
            []*FieldDefinition{
                MakeFieldDef( "f1", "Int32", nil ),
                MakeFieldDef( "f2", "Int32*", nil ),
                MakeFieldDef( "f3", "SymbolMap", nil ),
                MakeFieldDef( "f4", "Int32", int32( 1 ) ),
            },
        ),
    )
    p := mg.MakeTestIdPath
    idFact := mgRct.NewTestPointerIdFactory()
    mkTestId := mg.MakeTestId
    qn1 := parser.MustQualifiedTypeName( "ns1@v1/S1" )
    t1 := qn1.AsAtomicType()
    ss1 := mgRct.NewStructStartEvent( qn1 )
    fse := func( i int ) *mgRct.FieldStartEvent {
        return mgRct.NewFieldStartEvent( mkTestId( i ) )
    }
    iv1 := mgRct.NewValueEvent( mg.Int32( 1 ) )
    src, expct := []mgRct.ReactorEvent{}, []mgRct.EventExpectation{}
    apnd := func( ev mgRct.ReactorEvent, p objpath.PathNode, synth bool ) {
        if ! synth { src = append( src, ev ) }
        expct = append( expct, mgRct.EventExpectation{ Event: ev, Path: p } )
    }
    apnd( ss1, nil, false )
    apnd( fse( 1 ), p( 1 ), false )
    apnd( iv1, p( 1 ), false )
    apnd( fse( 2 ), p( 2 ), false )
    fld2Typ := asType( "Int32*" ).( *mg.ListTypeReference )
    apnd( idFact.NextListStart( fld2Typ ), p( 2 ), false )
    apnd( iv1, p( 2, "0" ), false )
    apnd( iv1, p( 2, "1" ), false )
    apnd( mgRct.NewEndEvent(), p( 2 ), false )
    apnd( fse( 3 ), p( 3 ), false )
    apnd( idFact.NextMapStart(), p( 3 ), false )
    apnd( fse( 1 ), p( 3, 1 ), false )
    apnd( iv1, p( 3, 1 ), false )
    apnd( mgRct.NewEndEvent(), p( 3 ), false )
    apnd( fse( 4 ), p( 4 ), true )
    apnd( iv1, p( 4 ), true )
    apnd( mgRct.NewEndEvent(), nil, false )
    rti.addTests(
        &EventPathTest{ 
            Source: mgRct.CopySource( src ), 
            Expect: expct, 
            Type: t1, 
            Map: dm,
        },
    )
}

func ( sm *ServiceMaps ) BuildOpMap() *ServiceDefinitionMap {
    res := NewServiceDefinitionMap( sm.Definitions )
    sm.ServiceIds.EachPair( func( svcId *mg.Identifier, val interface{} ) {
        qn := val.( *mg.QualifiedTypeName )
        res.MustPut( qn.Namespace, svcId, qn )
    })
    return res
}

func ( rti *rtInit ) addServiceRequestTests() {
    id := parser.MustIdentifier
    dm := MakeV1DefMap(
        MakeStructDef( "ns1@v1/S1",
            []*FieldDefinition{ MakeFieldDef( "f1", "Int32", nil ) } ),
        MakeSchemaDef( "ns1@v1/Schema1",
            []*FieldDefinition{ MakeFieldDef( "f1", "Int32", nil ) } ),
        MakeStructDef( "ns1@v1/Auth1",
            []*FieldDefinition{ MakeFieldDef( "f1", "Int32", nil ) } ),
        MakeEnumDef( "ns1@v1/E1", "e1", "e2" ),
        &PrototypeDefinition{
            Name: parser.MustQualifiedTypeName( "ns1@v1/Sec1" ),
            Signature: MakeCallSig(
                []*FieldDefinition{
                    MakeFieldDef(
                        "authentication",
                        "ns1@v1/Auth1",
                        nil,
                    ),
                },
                "Null",
                []string{},
            ),
        },
        &PrototypeDefinition{
            Name: parser.MustQualifiedTypeName( "ns1@v1/SecSchema1" ),
            Signature: MakeCallSig(
                []*FieldDefinition{
                    MakeFieldDef(
                        "authentication",
                        "ns1@v1/Schema1",
                        nil,
                    ),
                },
                "Null",
                []string{},
            ),
        },
        MakeServiceDef(
            "ns1@v1/Service1", "",
            MakeOpDef( "op1",
                MakeCallSig(
                    []*FieldDefinition{ MakeFieldDef( "p1", "&Int32?", nil ) },
                    "Null",
                    []string{},
                ),
            ),
            MakeOpDef( "op2",
                MakeCallSig( []*FieldDefinition{}, "Null", []string{} ) ),
            MakeOpDef( "op3",
                MakeCallSig(
                    []*FieldDefinition{
                        MakeFieldDef( "p1", "Int32", nil ),
                        MakeFieldDef( "p2", "Int32", nil ),
                    },
                    "Null",
                    []string{},
                ),
            ),
            MakeOpDef( "op4",
                MakeCallSig(
                    []*FieldDefinition{
                        MakeFieldDef( "p1", "ns1@v1/S1", nil ),
                    },
                    "Null",
                    []string{},
                ),
            ),
            MakeOpDef( "op5",
                MakeCallSig(
                    []*FieldDefinition{
                        MakeFieldDef( "p1", "ns1@v1/Schema1", nil ),
                    },
                    "Null",
                    []string{},
                ),
            ),
        ),
        MakeServiceDef(
            "ns1@v1/Service2", "ns1@v1/Sec1",
            MakeOpDef( "op1",
                MakeCallSig(
                    []*FieldDefinition{
                        MakeFieldDef( "p1", "Int32", int32( 42 ) ),
                        MakeFieldDef( 
                            "p2", 
                            "ns1@v1/E1", 
                            parser.MustEnum( "ns1@v1/E1", "e2" ),
                        ),
                        MakeFieldDef( "p3", "&ns1@v1/S1?", nil ),
                        MakeFieldDef(
                            "p4",
                            "Int32+",
                            mg.MustList( 
                                int32( -3 ), int32( -2 ), int32( -1 ) ),
                        ),
                    },
                    "Null",
                    []string{},
                ),
            ),
        ),
        MakeServiceDef(
            "ns1@v1/Service3", "ns1@v1/SecSchema1",
            MakeOpDef( "op1",
                MakeCallSig( []*FieldDefinition{}, "Null", []string{} ),
            ),
        ),
    )
    svcIds := mg.NewIdentifierMap();
    svcIds.Put( id( "svc1" ), 
        parser.MustQualifiedTypeName( "ns1@v1/Service1" ) )
    svcIds.Put( id( "svc2" ), 
        parser.MustQualifiedTypeName( "ns1@v1/Service2" ) )
    svcIds.Put( id( "svc3" ),
        parser.MustQualifiedTypeName( "ns1@v1/Service3" ) )
    maps := &ServiceMaps{ Definitions: dm, ServiceIds: svcIds }
    s1Inst1 := parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) )
    addSucc := func( in mg.Value, params *mg.SymbolMap, auth mg.Value ) {
        if params == nil { params = mg.EmptySymbolMap() }
        rti.addTests(
            &ServiceRequestTest{
                Maps: maps,
                In: in,
                Parameters: params,
                Authentication: auth,
            },
        )
    }
    addErr := func( in mg.Value, err error ) {
        rti.addTests( &ServiceRequestTest{ In: in, Maps: maps, Error: err } )
    }
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
        return parser.MustSymbolMap( pairs... )
    }
    addSucc( 
        mkReq( "ns1@v1", "svc1", "op1", 
            parser.MustSymbolMap( "p1", "1" ), "1" ),
        parser.MustSymbolMap( "p1", mustHeapVal( int32( 1 ) ) ),
        nil,
    )
    addSucc(
        mkReq( "ns1@v1", "svc1", "op1", nil, nil ),
        parser.MustSymbolMap(),
        nil,
    )
    svc2Op1Params1 := parser.MustSymbolMap(
        "p1", int32( 1 ),
        "p2", parser.MustEnum( "ns1@v1/E1", "e1" ),
        "p3", mustHeapVal( parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ) ),
        "p4", mg.MustList( int32( 1 ), int32( 2 ), int32( 3 ) ),
    )
    svc2Op1ParamsDefl := func( extra ...interface{} ) *mg.SymbolMap {
        pairs := []interface{}{
            "p1", int32( 42 ),
            "p2", parser.MustEnum( "ns1@v1/E1", "e2" ),
            "p4", mg.MustList( int32( -3 ), int32( -2 ), int32( -1 ) ),
        }
        pairs = append( pairs, extra... )
        return parser.MustSymbolMap( pairs... )
    }
    auth1Val1 := parser.MustStruct( "ns1@v1/Auth1", "f1", int32( 1 ) )
    addSucc(
        mkReq( 
            "ns1@v1", "svc2", "op1",
             parser.MustSymbolMap(
                "p1", "1",
                "p2", "e1",
                "p3", parser.MustSymbolMap( "f1", "1" ),
                "p4", mg.MustList( int64( 1 ), "2", int32( 3 ) ),
            ),
            parser.MustSymbolMap( "f1", "1" ),
        ),
        svc2Op1Params1,
        auth1Val1,
    )
    addSucc(
        mkReq( "ns1@v1", "svc2", "op1", nil, auth1Val1 ),
        svc2Op1ParamsDefl(),
        auth1Val1,
    )
    addSucc(
        mkReq( "ns1@v1", "svc2", "op1", 
            parser.MustSymbolMap( "p3", 
                parser.MustStruct( "ns1@v1/S1", "f1", "1" ) ),
            auth1Val1,
        ),
        svc2Op1ParamsDefl( 
            "p3", 
            mustHeapVal( parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ) ),
        ),
        auth1Val1,
    )
    addSucc(
        mkReq( "ns1@v1", "svc1", "op2", parser.MustSymbolMap(), nil ),
        parser.MustSymbolMap(), 
        nil,
    )
    addSucc(
        mkReq( "ns1@v1", "svc1", "op2", nil, nil ), 
        parser.MustSymbolMap(),
        nil,
    )
    addSucc(
        mkReq( "ns1@v1", "svc1", "op2", nil, int32( 1 ) ),
        parser.MustSymbolMap(),
        nil,
    )
    mkOp4Succ := func( p1Val interface{} ) {
        addSucc(
            mkReq( "ns1@v1", "svc1", "op4", 
                parser.MustSymbolMap( "p1", p1Val ), 
                nil,
            ),
            parser.MustSymbolMap( "p1", p1Val ),
            nil,
        )
    }
    mkOp4Succ( s1Inst1 )
    addSucc(
        mkReq( 
            "ns1@v1", "svc1", "op5",
            parser.MustSymbolMap( 
                "p1", parser.MustSymbolMap( "f1", int32( 1 ) ),
            ),
            nil,
        ),
        parser.MustSymbolMap( "p1", parser.MustSymbolMap( "f1", int32( 1 ) ) ),
        nil,
    )
    addSucc(
        mkReq(
            "ns1@v1", "svc1", "op5",
            parser.MustSymbolMap(
                "p1", parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
            ),
            nil,
        ),
        parser.MustSymbolMap(
            "p1", parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ) ),
        nil,
    )
    addErr(
        mkReq(
            "ns1@v1", "svc1", "op5",
            parser.MustSymbolMap( "p1", mg.EmptySymbolMap() ),
            nil,
        ),
        mg.NewMissingFieldsError(
            pathParams.Descend( parser.MustIdentifier( "p1" ) ),
            []*mg.Identifier{ parser.MustIdentifier( "f1" ) },
        ),
    )
    addSucc(
        mkReq( 
            "ns1@v1", "svc3", "op1",
            nil,
            parser.MustSymbolMap( "f1", int32( 1 ) ),
        ),
        nil,
        parser.MustSymbolMap( "f1", int32( 1 ) ),
    )
    addSucc(
        mkReq( 
            "ns1@v1", "svc3", "op1",
            nil,
            parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
        ),
        nil,
        parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
    )
    addErr(
        mkReq( 
            "ns1@v1", "svc3", "op1",
            nil,
            parser.MustSymbolMap(),
        ),
        mg.NewMissingFieldsError(
            pathAuth, []*mg.Identifier{ parser.MustIdentifier( "f1" ) } ),
    )
    addErr(
        mkReq( "ns1@v1", "svc2", "op1", parser.MustSymbolMap(), nil ),
        mg.NewMissingFieldsError( 
            nil, []*mg.Identifier{ mg.IdAuthentication } ),
    )
    addErr(
        mkReq( 
            "ns1@v1", "svc2", "op1", 
            parser.MustSymbolMap( "p1", []byte{} ), 
            auth1Val1,
        ),
        newTcErr( 
            mg.TypeInt32, mg.TypeBuffer, pathParams.Descend( id( "p1" ) ) ),
    )
    addErr(
        mkReq( 
            "ns1@v1", "svc2", "op1", 
            parser.MustSymbolMap( "p2", "bad" ), 
            auth1Val1,
        ),
        newVcErr( 
            pathParams.Descend( id( "p2" ) ), 
            "illegal value for enum ns1@v1/E1: bad",
        ),
    )
    addErr(
        mkReq( 
            "ns1@v1", "svc2", "op1", 
            svc2Op1Params1, 
            parser.MustStruct( "ns1@v1/Auth2" ),
        ),
        newTcErr( "ns1@v1/Auth1", "ns1@v1/Auth2", pathAuth ),
    )
    addErr(
        mkReq( "ns1@v1", "svc1", "op1", 
            parser.MustSymbolMap( "not-a-field", false ),
            int32( 1 ),
        ),
        mg.NewUnrecognizedFieldError( pathParams, id( "not-a-field" ) ),
    ) 
    addErr(
        mkReq( "ns1@v1", "svc1", "op3", nil, nil ),
        mg.NewMissingFieldsError( pathParams, makeIdList( "p1", "p2" ) ),
    )
    addErr(
        mkReq( "ns1@v2", "svc1", "op1", nil, nil ),
        mg.NewEndpointErrorNamespace( 
            parser.MustNamespace( "ns1@v2" ), 
            objpath.RootedAt( mg.IdNamespace ) ),
    )
    addErr(
        mkReq( "ns1@v1", "svc4", "op1", nil, nil ),
        mg.NewEndpointErrorService( 
            parser.MustIdentifier( "svc4" ), objpath.RootedAt( mg.IdService ) ),
    )
    addErr(
        mkReq( "ns1@v1", "svc1", "badOp", nil, nil ),
        mg.NewEndpointErrorOperation( 
            parser.MustIdentifier( "badOp" ), 
            objpath.RootedAt( mg.IdOperation ) ),
    )
}

func ( rti *rtInit ) addServiceResponseTests() {
    opIdStr := "op1"
    id := parser.MustIdentifier
    qn := parser.MustQualifiedTypeName
    svcType := qn( "ns1@v1/Service1" )
    newRespTest := func( od *OperationDefinition ) *ServiceResponseTest {
        dm := MakeV1DefMap(
            MakeStructDef( 
                "ns1@v1/S1", 
                []*FieldDefinition{ 
                    MakeFieldDef( "f1", "&Int32?", nil ),
                    MakeFieldDef( "f2", "&ns1@v1/S1?", nil ),
                },
            ),
            MakeSchemaDef( "ns1@v1/Schema1",
                []*FieldDefinition{ MakeFieldDef( "f1", "&Int32?", nil ) },
            ),
            MakeStructDef(
                "ns1@v1/Err1",
                []*FieldDefinition{
                    MakeFieldDef( "f1", "&Int32?", nil ),
                    MakeFieldDef( "f2", "&ns1@v1/S1?", nil ),
                },
            ),
            MakeStructDef(
                "ns1@v1/SecErr1",
                []*FieldDefinition{ MakeFieldDef( "f1", "&Int32?", nil ) },
            ),
            MakeEnumDef( "ns1@v1/E1", "e1" ),
            &PrototypeDefinition{
                Name: qn( "ns1@v1/Sec1" ),
                Signature: MakeCallSig(
                    []*FieldDefinition{
                        MakeFieldDef( "authentication", "Int32", nil ),
                    },
                    "Int32",
                    []string{ "ns1@v1/SecErr1" },
                ),
            },
            MakeServiceDef( svcType.ExternalForm(), "ns1@v1/Sec1", od ),
        )
        return &ServiceResponseTest{ 
            Definitions: dm,
            ServiceType: svcType,
            Operation: id( opIdStr ),
        }
    }
    opDef := func( resTyp string, throws []string ) *OperationDefinition {
        callSig := MakeCallSig( []*FieldDefinition{}, resTyp, throws )
        return MakeOpDef( opIdStr, callSig )
    }
    inVal := func( in interface{}, key string ) mg.Value {
        return parser.MustSymbolMap( key, in )
    }
    okResp := func( in interface{} ) mg.Value { return inVal( in, "result" ) }
    errResp := func( in interface{} ) mg.Value { return inVal( in, "error" ) }
    pathRes := objpath.RootedAt( mg.IdResult )
    pathErr := objpath.RootedAt( mg.IdError )
    addSucc := func( 
        in, expctRes, expctErr interface{}, resTyp string, throws []string ) {
        st := newRespTest( opDef( resTyp, throws ) )
        st.In = mg.MustValue( in )
        if ( expctRes != nil ) { st.ResultValue = mg.MustValue( expctRes ) }
        if ( expctErr != nil ) { st.ErrorValue = mg.MustValue( expctErr ) }
        rti.addTests( st )
    }
    addResSucc := func( in, val interface{}, resTyp string ) {
        addSucc( okResp( in ), mg.MustValue( val ), nil, resTyp, []string{} )
    }
    addErrSucc := func( in, err interface{}, throws ...string ) {
        addSucc( errResp( in ), nil, err, "Null", throws )
    }
    addResSucc( nil, nil, "Null" )
    addResSucc( int32( 1 ), int32( 1 ), "Int32" )
    addResSucc( "1", int32( 1 ), "Int32" )
    addResSucc( nil, nil, "&Int32?" )
    en1 := parser.MustEnum( "ns1@v1/E1", "e1" )
    addResSucc( en1, en1, "ns1@v1/E1" )
    addResSucc( "e1", en1, "ns1@v1/E1" )
    addResSucc( 
        parser.MustSymbolMap( "f1", mustHeapVal( int32( 1 ) ) ),
        parser.MustSymbolMap( "f1", mustHeapVal( int32( 1 ) ) ),
        "ns1@v1/Schema1",
    )
    addResSucc(
        parser.MustStruct( "ns1@v1/S1", "f1", mustHeapVal( int32( 1 ) ) ),
        parser.MustStruct( "ns1@v1/S1", "f1", mustHeapVal( int32( 1 ) ) ),
        "ns1@v1/Schema1",
    )
    i1HeapVal := func() *mg.HeapValue { return mustHeapVal( int32( 1 ) ) }
    s1 := parser.MustStruct( "ns1@v1/S1", "f1", i1HeapVal() )
    addResSucc( s1, s1, "ns1@v1/S1" )
    err1 := parser.MustStruct( "ns1@v1/Err1", "f1", i1HeapVal() )
    addErrSucc( err1, err1, "ns1@v1/Err1" )
    addErrSucc( err1, err1, "ns1@v1/Err1", "ns1@v1/SomeOtherErr" )
    secErr1 := parser.MustStruct( "ns1@v1/SecErr1", "f1", i1HeapVal() )
    addErrSucc( secErr1, secErr1 )
    addErrSucc( secErr1, secErr1, "ns1@v1/Err1" )
    // We're not really checking here that the error values are correct as
    // mingle struct values (most should have at least a 'message' field), only
    // that the types are allowed by the response cast even when not explicitly
    // declared in an operation definition (which will be the common case)
    for _, errTyp := range []string{
        "mingle:core@v1/MissingFieldsError",
        "mingle:core@v1/UnrecognizedFieldError",
        "mingle:core@v1/ValueCastError",
        "mingle:core@v1/EndpointError",
    } {
        err := parser.MustStruct( errTyp )
        addErrSucc( err, err )
    }
    addFail := 
        func( in interface{}, resTyp string, throws []string, err error ) {
        st := newRespTest( opDef( resTyp, throws ) )
        st.In = mg.MustValue( in )
        st.Error = err
        rti.addTests( st )
    }
    addResFail := func( in interface{}, resTyp string, err error ) {
        addFail( okResp( in ), resTyp, []string{}, err )
    }
    addErrFail := func( in interface{}, throws []string, err error ) {
        addFail( errResp( in ), "Null", throws, err )
    }
    addResFail( 
        []byte{},
        "Int32",
        newTcErr( mg.TypeInt32, mg.TypeBuffer, pathRes ),
    )
    addResFail(
        int32( 1 ),
        "Null",
        newTcErr( mg.TypeNull, mg.TypeInt32, pathRes ),
    )
    addResFail( 
        parser.MustStruct( "ns1@v1/S3" ),
        "ns1@v1/S1", 
        newTcErr( "ns1@v1/S1", "ns1@v1/S3", pathRes ),
    )
    addResFail(
        parser.MustEnum( "ns1@v1/E1", "bad" ),
        "ns1@v1/E1",
        newVcErr( pathRes, "illegal value for enum ns1@v1/E1: bad" ),
    )
    addResFail(
        parser.MustStruct( "ns1@v1/S1", "not-a-field", int32( 1 ) ),
        "ns1@v1/S1",
        mg.NewUnrecognizedFieldError( pathRes, id( "not-a-field" ) ),
    )
    addResFail(
        parser.MustStruct( "ns1@v1/S1", 
            "f2", parser.MustStruct( "ns1@v1/S1", "f1", []byte{} ) ),
        "ns1@v1/S1",
        newTcErr( 
            "&Int32?", 
            "Buffer", 
            pathRes.Descend( id( "f2" ) ).Descend( id( "f1" ) ),
        ),
    )
    addErrFail(
        parser.MustStruct( "ns1@v1/Err1", "not-a-field", int32( 1 ) ),
        []string{ "ns1@v1/Err1" },
        mg.NewUnrecognizedFieldError( pathErr, id( "not-a-field" ) ),
    )
    addErrFail(
        parser.MustStruct( "ns1@v1/Err1",
            "f2", parser.MustStruct( "ns1@v1/S1", "not-a-field", 1 ) ),
        []string{ "ns1@v1/Err1" },
        mg.NewUnrecognizedFieldError( 
            pathErr.Descend( id( "f2" ) ), id( "not-a-field" ) ),
    )
    addErrFail(
        parser.MustStruct( "ns1@v1/UndeclaredErr" ),
        []string{},
        errorForUnexpectedErrorType( pathErr, mkTyp( "ns1@v1/UndeclaredErr" ) ),
    )
    addErrFail(
        parser.MustStruct( "ns1@v1/UndeclaredErr" ),
        []string{ "ns1@v1/Err1" },
        errorForUnexpectedErrorType( pathErr, mkTyp( "ns1@v1/UndeclaredErr" ) ),
    )
    addErrFail(
        parser.MustStruct( "ns1@v1/SecErr1", "not-a-field", int32( 1 ) ),
        []string{},
        mg.NewUnrecognizedFieldError( pathErr, id( "not-a-field" ) ),
    )
    addErrFail(
        int32( 1 ),
        []string{},
        errorForUnexpectedErrorType( pathErr, mg.TypeInt32 ),
    )
    addErrFail(
        int32( 1 ),
        []string{ "ns1@v1/Err1" },
        errorForUnexpectedErrorType( pathErr, mg.TypeInt32 ),
    )
    addErrFail(
        mg.MustList(),
        []string{ "ns1@v1/Err1" },
        errorForUnexpectedErrorType( pathErr, mg.TypeOpaqueList ),
    )
    addErrFail(
        parser.MustSymbolMap(),
        []string{ "ns1@v1/Err1" },
        errorForUnexpectedErrorType( pathErr, mg.TypeSymbolMap ),
    )
}

func ( rti *rtInit ) call() {
    rti.addBaseTypeTests()    
    rti.addMiscTcErrors()
    rti.addMiscVcErrors()
    rti.addNonRootPathTestErrors()
    rti.addStringTests()
    rti.addNumTests()
    rti.addBufferTests()
    rti.addTimeTests()
    rti.addNullableTests()
    rti.addListTests()
    rti.addMapTests()
    rti.addBaseFieldCastTests()
    rti.addFieldSetCastTests()
    rti.addStructTests()
    rti.addSchemaCastTests()
    rti.addEnumValCastTests()
    rti.addDeepCatchallTests()
    rti.addDefaultCastTests()
    rti.addDefaultPathTests()
    rti.addServiceRequestTests()
    rti.addServiceResponseTests()
}

func init() {
    reactorTestNs = parser.MustNamespace( "mingle:types@v1" )
    f := func( b *mgRct.ReactorTestSetBuilder ) { ( &rtInit{ b: b } ).call() }
    mgRct.AddTestInitializer( reactorTestNs, f )
}
