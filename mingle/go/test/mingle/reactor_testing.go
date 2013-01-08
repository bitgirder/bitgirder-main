package mingle

import (
    "bitgirder/objpath"
    "fmt"
    "encoding/base64"
)

type ReactorTest interface {}

var StdReactorTests []ReactorTest

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

type StructuralReactorPathTest struct {
    Events []ReactorEvent
    StartPath objpath.PathNode
    Path objpath.PathNode
}

func initStructuralReactorTests() {
    evStartStruct1 := StructStartEvent{ MustTypeReference( "ns1@v1/S1" ) }
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
    mk3 := func( path idPath, evs ...ReactorEvent ) *StructuralReactorPathTest {
        return &StructuralReactorPathTest{ Events: evs, Path: path }
    }
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
        mk3( nil, evValue1 ),
        mk3( objpath.RootedAt( idF1 ), evStartStruct1, evStartField1 ),
        mk3( objpath.RootedAt( idF1 ), EvMapStart, evStartField1 ),
        mk3( nil, evStartStruct1, evStartField1, evValue1 ),
        mk3( objpath.RootedAtList(), EvListStart ),
        mk3( objpath.RootedAtList(), EvListStart, EvMapStart ),
        mk3( objpath.RootedAtList().Next(), EvListStart, evValue1 ),
        mk3( objpath.RootedAtList().SetIndex( 1 ),
            EvListStart, evValue1, EvMapStart,
        ),
        mk3( objpath.RootedAt( idF1 ), EvMapStart, evStartField1, EvMapStart ),
        mk3( 
            objpath.RootedAt( idF1 ).
                Descend( idF2 ).
                StartList().
                Next().
                Next().
                StartList().
                Next().
                Descend( idF1 ),
            evStartStruct1, evStartField1,
            EvMapStart, evStartField2,
            EvListStart,
            evValue1,
            evValue1,
            EvListStart,
            evValue1,
            EvMapStart, evStartField1,
        ),
        &StructuralReactorPathTest{
            Events: []ReactorEvent{ EvMapStart, evStartField1 },
            StartPath: objpath.RootedAt( idF2 ).StartList().SetIndex( 3 ),
            Path: 
                objpath.RootedAt( idF2 ).
                    StartList().
                    SetIndex( 3 ).
                    Descend( idF1 ),
        },
    )
}

type CastReactorTest struct {
    In Value
    Expect Value
    Path objpath.PathNode
    Type TypeReference
    Err error
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

func ( t *crtInit ) addCvt( crt *CastReactorTest ) { 
    StdReactorTests = append( StdReactorTests, crt ) 
}

func ( t *crtInit ) addCvtDefault( crt *CastReactorTest ) {
    crt.Path = crtPathDefault
    t.addCvt( crt )
}

func ( t *crtInit ) addSucc( 
    in, expct interface{}, typ TypeReferenceInitializer ) {
    t.addCvtDefault( 
        &CastReactorTest{ 
            In: MustValue( in ), 
            Expect: MustValue( expct ), 
            Type: asTypeReference( typ ),
        },
    )
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

func ( t *crtInit ) addTcError(
    in interface{}, typExpct, typAct TypeReferenceInitializer ) {
    err := newTypeCastError( 
        asTypeReference( typExpct ),
        asTypeReference( typAct ),
        crtPathDefault,
    )
    t.addCvtDefault( 
        &CastReactorTest{
            In: MustValue( in ),
            Type: asTypeReference( typExpct ),
            Err: err,
        },
    )
}

func ( t *crtInit ) addMiscTcErrors() {
    t.addTcError( t.en1, "ns1@v1/Bad", t.en1.Type )
    t.addTcError( t.struct1, "ns1@v1/Bad", t.struct1.Type )
    t.addTcError( "s", TypeNull, TypeString )
    t.addTcError( MustList( 1, 2 ), TypeString, "Value*" )
    t.addTcError( MustList(), "String?", "Value*" )
    t.addTcError( "s", "String*", "String" )
    t.addCvtDefault(
        &CastReactorTest{
            In: MustList( 1, t.struct1 ),
            Type: asTypeReference( "Int32*" ),
            Err: newTypeCastError(
                asTypeReference( "Int32" ),
                t.struct1.Type,
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

func ( t *crtInit ) addVcError0( 
    val interface{}, typ TypeReferenceInitializer, path idPath, msg string ) {
    typRef := asTypeReference( typ )
    t.addCvtDefault(
        &CastReactorTest{
            In: MustValue( val ),
            Type: typRef,
            Err: newValueCastError( path, msg ),
        },
    )
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
}

func initCastReactorTests() { ( &crtInit{} ).call() }

func init() {
    StdReactorTests = []ReactorTest{}
    initValueBuildReactorTests()
    initStructuralReactorTests()
    initCastReactorTests()
}
