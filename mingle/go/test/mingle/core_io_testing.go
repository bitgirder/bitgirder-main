package mingle

import (
    "bytes"
    "strings"
    "fmt"
    "math"
//    "log"
)

// we define tests in this package that are actually tested in subpackages in
// order to consolidate test definitions in as few places as possible.

type BinIoRoundtripTest struct {
    Name string
    Val interface{}
}

type binIoRoundtripTestBuilder struct {

    // tests is the definitive and deterministic ordering (the determinism is
    // helpful when a test fails); nmCheck is used internally to prevent test
    // construction duplicates
    tests []interface{}
    nmCheck map[ string ]interface{}
}

// convenience funcs with assertions on our own code against typos
func ( b *binIoRoundtripTestBuilder ) getVal( nm string ) interface{} {
    if val, ok := b.nmCheck[ nm ]; ok { return val }
    panic( libErrorf( "valMap[ %q ]: no value", nm ) )
}

// sets a test with given name and val, returning val unchanged
func ( b* binIoRoundtripTestBuilder ) setVal( 
    nm string, val interface{} ) interface{} {

    test := &BinIoRoundtripTest{ Name: nm, Val: val }
    if _, ok := b.nmCheck[ nm ]; ok {
        panic( libErrorf( "valMap[ %q ] already exists", nm ) )
    }
    b.nmCheck[ nm ] = test
    b.tests = append( b.tests, test )
    return val
}

func ( b *binIoRoundtripTestBuilder ) addValueTests() {
    ns1V1 := mkNs( mkId( "v1" ), mkId( "ns1" ) )
    ns1V1T1 := mkQn( ns1V1, mkDeclNm( "T1" ) )
    b.setVal( "null-val", NullVal )
    b.setVal( "string-empty", String( "" ) )
    b.setVal( "string-val1", String( "hello" ) )
    b.setVal( "bool-true", Boolean( true ) )
    b.setVal( "bool-false", Boolean( false ) )
    b.setVal( "buffer-empty", Buffer( []byte{} ) )
    b.setVal( "buffer-nonempty", Buffer( []byte{ 0x00, 0x01 } ) )
    b.setVal( "int32-min", Int32( math.MinInt32 ) )
    b.setVal( "int32-max", Int32( math.MaxInt32 ) )
    b.setVal( "int32-pos1", Int32( int32( 1 ) ) )
    b.setVal( "int32-zero", Int32( int32( 0 ) ) )
    b.setVal( "int32-neg1", Int32( int32( -1 ) ) )
    b.setVal( "int64-min", Int64( math.MinInt64 ) )
    b.setVal( "int64-max", Int64( math.MaxInt64 ) )
    b.setVal( "int64-pos1", Int64( int64( 1 ) ) )
    b.setVal( "int64-zero", Int64( int64( 0 ) ) )
    b.setVal( "int64-neg1", Int64( int64( -1 ) ) )
    b.setVal( "uint32-max", Uint32( math.MaxUint32 ) )
    b.setVal( "uint32-min", Uint32( uint32( 0 ) ) )
    b.setVal( "uint32-pos1", Uint32( uint32( 1 ) ) )
    b.setVal( "uint64-max", Uint64( math.MaxUint64 ) )
    b.setVal( "uint64-min", Uint64( uint64( 0 ) ) )
    b.setVal( "uint64-pos1", Uint64( uint64( 1 ) ) )
    b.setVal( "float32-val1", Float32( float32( 1 ) ) )
    b.setVal( "float32-max", Float32( math.MaxFloat32 ) )
    b.setVal( "float32-smallest-nonzero",
        Float32( math.SmallestNonzeroFloat32 ) )
    b.setVal( "float64-val1", Float64( float64( 1 ) ) )
    b.setVal( "float64-max", Float64( math.MaxFloat64 ) )
    b.setVal( "float64-smallest-nonzero",
        Float64( math.SmallestNonzeroFloat64 ) )
    b.setVal( "time-val1", MustTimestamp( "2013-10-19T02:47:00-08:00" ) )
    b.setVal( "enum-val1", &Enum{ 
        Type: mkQn( ns1V1, mkDeclNm( "E1" ) ),
        Value: mkId( "val1" ),
    } )
    b.setVal( "symmap-empty", MustSymbolMap() )
    b.setVal( "symmap-flat", 
        MustSymbolMap( 
            mkId( "k1" ), int32( 1 ), 
            mkId( "k2" ), int32( 2 ),
            mkId( "k3" ), int32( 1 ),
        ),
    )
    b.setVal( "symmap-nested",
        MustSymbolMap( mkId( "k1" ), 
            MustSymbolMap( mkId( "kk1" ), int32( 1 ) ) ) )

    b.setVal( "struct-empty", MustStruct( ns1V1T1 ) )
    b.setVal( "struct-flat", MustStruct( ns1V1T1, mkId( "k1" ), int32( 1 ) ) )
    b.setVal( "list-empty", MustList() )
    b.setVal( "list-scalars", MustList( int32( 1 ), "hello" ) )
    b.setVal( "list-nested",
        MustList( int32( 1 ), MustList(), MustList( "hello" ), NullVal ) )
    b.setVal( "list-typed",
        MustList( 
            &ListTypeReference{ TypeInt32, true }, 
            int32( 0 ), int32( 1 ),
        ),
    )
}

func ( b *binIoRoundtripTestBuilder ) addDefinitionTests() {
    set := func( ef interface { ExternalForm() string } ) { 
        fqNm := fmt.Sprintf( "%T", ef )
        lastDot := strings.LastIndex( fqNm, "." )
        simplNm := fqNm[ lastDot + 1 : ]
        b.setVal( fmt.Sprintf( "%s/%s", simplNm, ef ), ef )
    }
    set( mkId( "id1" ) )
    set( mkId( "id1", "id2" ) )
    ns1V1 := mkNs( mkId( "v1" ), mkId( "ns1" ) )
    set( ns1V1 )
    ns1ns2V1 := mkNs( mkId( "v1" ), mkId( "ns1" ), mkId( "ns2" ) )
    set( ns1ns2V1 )
    set( mkDeclNm( "T1" ) )
    set( mkQn( ns1ns2V1, mkDeclNm( "T1" ) ) )
    mkV1Typ := func( nm string, rx ValueRestriction ) *AtomicTypeReference {
        ns := mkNs( mkId( "v1" ), mkId( "mingle" ), mkId( "core" ) )
        qn := mkQn( ns, mkDeclNm( nm ) )
        return NewAtomicTypeReference( qn, rx )
    }
    set( mkV1Typ( "String", MustRegexRestriction( "a" ) ) )
    set( 
        mkV1Typ( 
            "String", 
            &RangeRestriction{ true, String( "a" ), String( "b" ), true },
        ),
    )
    set( 
        mkV1Typ( 
            "Timestamp",
            &RangeRestriction{
                true,
                MustTimestamp( "2012-01-01T00:00:00Z" ),
                MustTimestamp( "2012-02-01T00:00:00Z" ),
                true,
            },
        ),
    )
    set( 
        mkV1Typ(
            "Int32",
            &RangeRestriction{ false, Int32( 0 ), Int32( 10 ), false },
        ),
    )
    set(
        mkV1Typ(
            "Int64",
            &RangeRestriction{ true, Int64( 0 ), Int64( 10 ), true },
        ),
    )
    set( 
        mkV1Typ(
            "Uint32",
            &RangeRestriction{ false, Uint32( 0 ), Uint32( 10 ), false },
        ),
    )
    set( 
        mkV1Typ( 
            "Uint64",
            &RangeRestriction{ true, Uint64( 0 ), Uint64( 0 ), true },
        ),
    )
    set(
        mkV1Typ(
            "Float32",
            &RangeRestriction{ false, Float32( 0.0 ), Float32( 1.0 ), true },
        ),
    )
    set( 
        mkV1Typ(
            "Float64",
            &RangeRestriction{ true, Float64( 0.0 ), Float64( 1.0 ), false },
        ),
    )
    set( 
        mkV1Typ(
            "Float64",
            &RangeRestriction{ MinClosed: false, MaxClosed: false },
        ),
    )
    typNs1V1T1 := mkQn( ns1V1, mkDeclNm( "T1" ) ).AsAtomicType()
    set( typNs1V1T1 )
    set( &ListTypeReference{ ElementType: typNs1V1T1, AllowsEmpty: false } )
    set( &ListTypeReference{ ElementType: typNs1V1T1, AllowsEmpty: true } )
    set( 
        &NullableTypeReference{
            &ListTypeReference{ ElementType: typNs1V1T1, AllowsEmpty: true },
        },
    )
    set( NewPointerTypeReference( typNs1V1T1 ) )
    set( &NullableTypeReference{ NewPointerTypeReference( typNs1V1T1 ) } )
    set( 
        &ListTypeReference{
            AllowsEmpty: true,
            ElementType: &NullableTypeReference{
                &ListTypeReference{
                    AllowsEmpty: false,
                    ElementType: &NullableTypeReference{
                        NewPointerTypeReference( typNs1V1T1 ),
                    },
                },
            },
        },
    )
}

func addBinIoRoundtripTests( tests []interface{} ) []interface{} {
    b := &binIoRoundtripTestBuilder{}
    b.nmCheck = map[ string ]interface{}{}
    b.tests = []interface{}{}
    b.addValueTests()
    b.addDefinitionTests()
    return b.tests
}

type BinIoSequenceRoundtripTest struct {
    Name string
    Seq []Value
}

func addBinIoSequenceRoundtripTests( tests []interface{} ) []interface{} {
    ns1V1S1 := mkQn( mkNs( mkId( "v1" ), mkId( "ns1" ) ), mkDeclNm( "S1" ) )
    return append( tests,
        &BinIoSequenceRoundtripTest{
            Name: "struct-sequence",
            Seq: []Value{
                MustStruct( ns1V1S1 ),
                MustStruct( ns1V1S1, mkId( "f1" ), int32( 1 ) ),
            },
        },
    )
}

// We'd need to change this in the event we actually create 0x64 valid types in
// BinIo libs
const binIoInvalidTypeCode = int32( 0x64 )

type BinIoInvalidDataTest struct {
    Name string
    ErrMsg string
    Input []byte
}

type utf8Input string

func appendInput( data interface{}, w *BinWriter ) {
    switch v := data.( type ) {
    case string: 
        if err := w.WriteScalarValue( String( v ) ); err != nil { panic( err ) }
    case uint8: if err := w.WriteUint8( v ); err != nil { panic( err ) }
    case IoTypeCode: if err := w.WriteTypeCode( v ); err != nil { panic( err ) }
    case int32: if err := w.WriteInt32( v ); err != nil { panic( err ) }
    case uint64: if err := w.WriteUint64( v ); err != nil { panic( err ) }
    case utf8Input: 
        if err := w.WriteUtf8( string( v ) ); err != nil { panic( err ) }
    case *Identifier: 
        if err := w.WriteIdentifier( v ); err != nil { panic( err ) }
    case *Namespace:
        if err := w.WriteNamespace( v ); err != nil { panic( err ) }
    case *QualifiedTypeName: 
        if err := w.WriteQualifiedTypeName( v ); err != nil { panic( err ) }
    case TypeReference:
        if err := w.WriteTypeReference( v ); err != nil { panic( err ) }
    default: panic( libErrorf( "unrecognized input elt: %T", v ) )
    }
}

func makeBinIoInvalidDataTest( data ...interface{} ) []byte {
    buf := &bytes.Buffer{}
    w := NewWriter( buf )
    for _, val := range data { appendInput( val, w ) }
    return buf.Bytes()
}

func addBinIoInvalidDataTests( tests []interface{} ) []interface{} {
    nsV1 := mkNs( mkId( "v1" ), mkId( "ns" ) )
    qnNsV1S := mkQn( nsV1, mkDeclNm( "S" ) )
    idF1 := mkId( "f1" )
    return append( tests, 
        &BinIoInvalidDataTest{
            Name: "unexpected-top-level-type-code",
            ErrMsg: "[offset 0]: unrecognized value code: 0x64",
            Input: makeBinIoInvalidDataTest( binIoInvalidTypeCode ),
        },
        &BinIoInvalidDataTest{
            Name: "unexpected-symmap-val-type-code",
            ErrMsg: `[offset 35]: unrecognized value code: 0x64`,
            Input: makeBinIoInvalidDataTest(
                IoTypeCodeStruct, qnNsV1S,
                IoTypeCodeField, idF1, binIoInvalidTypeCode,
            ),
        },
        &BinIoInvalidDataTest{
            Name: "unexpected-list-val-type-code",
            ErrMsg: `[offset 88]: unrecognized value code: 0x64`,
            Input: makeBinIoInvalidDataTest(
                IoTypeCodeStruct, qnNsV1S,
                IoTypeCodeField, idF1,
                IoTypeCodeList, 
                    &ListTypeReference{
                        AllowsEmpty: true,
                        ElementType: TypeInt32,
                    }, // type
                    IoTypeCodeInt32, int32( 10 ), // an okay list val
                    binIoInvalidTypeCode,
            ),
        },
        &BinIoInvalidDataTest{
            Name: "invalid-list-type",
            ErrMsg: `[offset 1]: Expected type code 0x06 but got 0x05`,
            Input: makeBinIoInvalidDataTest( IoTypeCodeList, TypeInt32 ),
        },
        &BinIoInvalidDataTest{
            Name: "invalid-identifier-part",
            ErrMsg: `[offset 36]: invalid identifier part: Part`,
            Input: makeBinIoInvalidDataTest(
                IoTypeCodeStruct, qnNsV1S,
                IoTypeCodeField, 
                    IoTypeCodeId, 
                        uint8( 2 ), utf8Input( "bad" ), utf8Input( "Part" ),
            ),
        },
        &BinIoInvalidDataTest{
            Name: "invalid-declared-type-name",
            ErrMsg: `[offset 21]: invalid type name: "A$BadName"`,
            Input: makeBinIoInvalidDataTest(
                IoTypeCodeStruct,
                    IoTypeCodeQn, 
                        nsV1, 
                        IoTypeCodeDeclNm, utf8Input( "A$BadName" ),
            ),
        },
    )
}

// We can't create this result in an init() func since we make use of other
// library functions, such as MustTypeReference, which assume proper library
// initialization.
func CreateCoreIoTests() []interface{} {
    res := []interface{}{}
    res = addBinIoRoundtripTests( res )
    res = addBinIoSequenceRoundtripTests( res )
    res = addBinIoInvalidDataTests( res )
    return res
}

func CoreIoTestNameFor( test interface{} ) string {
    mk := func( pref, nm string ) string {
        return fmt.Sprintf( "%s/%s", pref, nm )
    }
    switch v := test.( type ) {
    case *BinIoRoundtripTest: return mk( "roundtrip", v.Name )
    case *BinIoSequenceRoundtripTest: return mk( "sequence-roundtrip", v.Name )
    case *BinIoInvalidDataTest: return mk( "invalid-data", v.Name )
    }
    panic( libErrorf( "unhandled test: %T", test ) )
}

func CoreIoTestsByName() map[ string ]interface{} {
    tests := CreateCoreIoTests()
    res := make( map[ string ]interface{}, len( tests ) )
    for _, t := range tests { res[ CoreIoTestNameFor( t ) ] = t }
    return res
}
