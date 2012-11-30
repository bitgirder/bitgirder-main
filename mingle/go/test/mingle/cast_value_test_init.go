package mingle

import (
    "bitgirder/objpath"
    "encoding/base64"
    "fmt"
)

type CastValueTest struct {
    
    In Value
    Expect Value
    Type TypeReference
    Path objpath.PathNode
    Err error
}

var cvtPathDefault = objpath.RootedAt( MustIdentifier( "inVal" ) )

type cvtInit struct {
    
    arr []*CastValueTest

    buf1 Buffer
    tm1 Timestamp
    map1 *SymbolMap
    en1 *Enum
    struct1 *Struct
}

func ( t *cvtInit ) initStdVals() {
    t.buf1 = Buffer( []byte{ byte( 0 ), byte( 1 ), byte( 2 ) } )
    t.tm1 = Now()
    t.map1 = MustSymbolMap( "key1", 1, "key2", "val2" )
    t.en1 = MustEnum( "ns1@v1/En1", "en-val1" )
    t.struct1 = MustStruct( "ns1@v1/S1", "key1", "val1" )
}

func ( t *cvtInit ) addCvt( cvt *CastValueTest ) { 
    t.arr = append( t.arr, cvt ) 
}

func ( t *cvtInit ) addCvtDefault( cvt *CastValueTest ) {
    cvt.Path = cvtPathDefault
    t.addCvt( cvt )
}

func ( t *cvtInit ) addSucc( 
    in, expct interface{}, typ TypeReferenceInitializer ) {
    t.addCvtDefault( 
        &CastValueTest{ 
            In: MustValue( in ), 
            Expect: MustValue( expct ), 
            Type: asTypeReference( typ ),
        },
    )
}

func ( t *cvtInit ) addIdent( in interface{}, typ TypeReferenceInitializer ) {
    v := MustValue( in )
    t.addSucc( v, v, asTypeReference( typ ) )
}

func ( t *cvtInit ) addBaseTypeTests() {
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
    t.addIdent( t.struct1, TypeStruct )
    t.addIdent( t.struct1, t.struct1.Type )
    t.addIdent( 1, TypeValue )
    t.addIdent( nil, TypeNull )
    t.addSucc( Int32( -1 ), Uint32( uint32( 4294967295 ) ), TypeUint32 )
    t.addSucc( Int64( -1 ), Uint32( uint32( 4294967295 ) ), TypeUint32 )
    t.addSucc( 
        Int32( -1 ), Uint64( uint64( 18446744073709551615 ) ), TypeUint64 )
    t.addSucc( 
        Int64( -1 ), Uint64( uint64( 18446744073709551615 ) ), TypeUint64 )
    t.addSucc( t.en1, t.en1, TypeEnum )
    t.addSucc( "true", true, TypeBoolean )
    t.addSucc( "TRUE", true, TypeBoolean )
    t.addSucc( "TruE", true, TypeBoolean )
    t.addSucc( "false", false, TypeBoolean )
    t.addSucc( true, "true", TypeString )
    t.addSucc( false, "false", TypeString )
}

func ( t *cvtInit ) addTcError(
    in interface{}, typExpct, typAct TypeReferenceInitializer ) {
    err := newTypeCastError( 
        asTypeReference( typExpct ),
        asTypeReference( typAct ),
        cvtPathDefault,
    )
    t.addCvtDefault( 
        &CastValueTest{
            In: MustValue( in ),
            Type: asTypeReference( typExpct ),
            Err: err,
        },
    )
}

func ( t *cvtInit ) addMiscTcErrors() {
    t.addTcError( t.en1, "ns1@v1/Bad", t.en1.Type )
    t.addTcError( t.struct1, "ns1@v1/Bad", t.struct1.Type )
    t.addTcError( "s", TypeNull, TypeString )
    t.addTcError( MustList( 1, 2 ), TypeString, "Value*" )
    t.addTcError( MustList(), "String?", "Value*" )
    t.addTcError( "s", "String*", "String" )
    t.addCvtDefault(
        &CastValueTest{
            In: MustList( 1, t.struct1 ),
            Type: asTypeReference( "Int32*" ),
            Err: newTypeCastError(
                asTypeReference( "Int32" ),
                t.struct1.Type,
                cvtPathDefault.StartList().Next(),
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

func ( t *cvtInit ) addVcError0( 
    val interface{}, typ TypeReferenceInitializer, path idPath, msg string ) {
    typRef := asTypeReference( typ )
    t.addCvtDefault(
        &CastValueTest{
            In: MustValue( val ),
            Type: typRef,
            Err: newValueCastError( typRef, path, msg ),
        },
    )
}

func ( t *cvtInit ) addVcError( 
    val interface{}, typ TypeReferenceInitializer, msg string ) {
    t.addVcError0( val, typ, cvtPathDefault, msg )
}

func ( t *cvtInit ) addMiscVcErrors() {
    t.addVcError( "s", TypeBoolean, `Invalid boolean value: "s"` )
    t.addVcError( nil, TypeString, "Value is null" )
    t.addVcError( nil, `String~"a"`, "Value is null" )
    t.addVcError( MustList(), "String+", "List is empty" )
    t.addVcError0( 
        MustList( MustList( 1, 2 ), MustList() ), 
        "Int32+*", 
        cvtPathDefault.StartList().Next(),
        "List is empty",
    )
}

func ( t *cvtInit ) addStringTests() {
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
        cvtPathDefault.StartList().Next(),
        "Value \"b\" does not satisfy restriction \"^a+$\"",
    )
}

func ( t *cvtInit ) addIdentityNumTests() {
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

func ( t *cvtInit ) addTruncateNumTests() {
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

func ( t *cvtInit ) addNumTests() {
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

func ( t *cvtInit ) addBufferTests() {
    buf1B64 := base64.StdEncoding.EncodeToString( t.buf1 )
    t.addSucc( t.buf1, buf1B64, TypeString )
    t.addSucc( buf1B64, t.buf1, TypeBuffer  )
}

func ( t *cvtInit ) addTimeTests() {
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

func ( t *cvtInit ) addEnumTests() {
    t.addSucc( t.en1, "en-val1", TypeString  )
}

func ( t *cvtInit ) addNullableTests() {
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
func ( t *cvtInit ) addListTests() {
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

func ( t *cvtInit ) call() []*CastValueTest {
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
    return t.arr
}

// We generate these directly instead of during an init() func so that we can
// have access to things defined by the library (TypeBoolean, etc)
func GetCastValueTests() []*CastValueTest {
    return ( &cvtInit{ arr: []*CastValueTest{} } ).call()
}
