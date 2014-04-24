package mingle

import (
    "bytes"
    "bitgirder/objpath"
    "strings"
    "fmt"
    "math"
    "bitgirder/assert"
//    "log"
)

func assertWriteValueValue(
    rd *BinReader, expct interface{}, a *assert.PathAsserter ) {
    
    if val, err := rd.ReadValue(); err == nil {
        if expct == nil { val = NullVal }
        EqualWireValues( expct.( Value ), val, a )
    } else { a.Fatal( err ) }
}

func assertWriteValue(
    rd *BinReader, expct interface{}, a *assert.PathAsserter ) {

    switch v := expct.( type ) {
    case Value: assertWriteValueValue( rd, expct, a )
    case *Identifier:
        if id, err := rd.ReadIdentifier(); err == nil { 
            a.True( v.Equals( id ) )
        } else { a.Fatal( err ) }
    case objpath.PathNode:
        if n, err := rd.ReadIdPath(); err == nil {
            a.Equal( v, n ) 
        } else { a.Fatal( err ) }
    case *Namespace:
        if ns, err := rd.ReadNamespace(); err == nil {
            a.True( v.Equals( ns ) )
        } else { a.Fatal( err ) }
    case TypeName:
        if nm, err := rd.ReadTypeName(); err == nil {
            a.True( v.Equals( nm ) )
        } else { a.Fatal( err ) }
    case TypeReference:
        if typ, err := rd.ReadTypeReference(); err == nil {
            a.Truef( v.Equals( typ ), "expct (%v) != act (%v)", v, typ )
        } else { a.Fatal( err ) }
    default: a.Fatalf( "Unhandled expct expct: %T", expct )
    }
}

type BinIoRoundtripTest struct {
    Name string
    Val interface{}
}

func ( t *BinIoRoundtripTest ) AssertWriteValue( 
    rd *BinReader, a *assert.PathAsserter ) {

    assertWriteValue( rd, t.Val, a )
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
    b.setVal( "enum-val1", MustEnum( "ns1@v1/E1", "val1" ) )
    b.setVal( "symmap-empty", MustSymbolMap() )
    b.setVal( "val-ptr", NewHeapValue( Int32( 1 ) ) )

    val1Ptr := NewHeapValue( Int32( 1 ) )
    b.setVal( "val-ptr-with-refs", MustList( val1Ptr, val1Ptr, val1Ptr ) )

    b.setVal( "symmap-flat", 
        MustSymbolMap( 
            "k1", int32( 1 ), 
            "k2", int32( 2 ),
            "k3", NewHeapValue( Int32( int32( 1 ) ) ),
        ),
    )

    b.setVal( "symmap-with-refs",
        MustSymbolMap(
            "k1", val1Ptr, 
            "k2", val1Ptr,
            "k3", MustList( val1Ptr, val1Ptr ),
        ),
    )

    b.setVal( "symmap-nested",
        MustSymbolMap( "k1", MustSymbolMap( "kk1", int32( 1 ) ) ) )

    b.setVal( "struct-empty", MustStruct( "ns1@v1/T1" ) )
    b.setVal( "struct-flat", MustStruct( "ns1@v1/T1", "k1", int32( 1 ) ) )
    b.setVal( "list-empty", MustList() )
    b.setVal( "list-scalars", MustList( int32( 1 ), "hello" ) )

    b.setVal( "list-nested",
        MustList( int32( 1 ), MustList(), MustList( "hello" ), NullVal ) )

    b.setVal( "list-pointers",
        MustList(
            int32( 1 ),
            NewHeapValue( Int32( int32( 1 ) ) ),
            NewHeapValue(
                MustList( NewHeapValue( Int32( int32( 1 ) ) ) ) ),
            MustList( NewHeapValue( Int32( int32( 1 ) ) ) ),
        ),
    )
}

func ( b *binIoRoundtripTestBuilder ) addPathTests() {
    setPath := func( nm string, p objpath.PathNode ) objpath.PathNode {
        return b.setVal( nm, p ).( objpath.PathNode )
    }
    p1 := setPath( "p1", objpath.RootedAt( id( "id1" ) ) )
    p2 := setPath( "p2", p1.Descend( id( "id2" ) ) )
    p3 := setPath( "p3", p2.StartList().Next().Next() )
    setPath( "p4", p3.Descend( id( "id3" ) ) )
    setPath( "p5", objpath.RootedAtList().Descend( id( "id1" ) ) )
}

func ( b *binIoRoundtripTestBuilder ) addDefinitionTests() {
    set := func( ef extFormer ) { 
        fqNm := fmt.Sprintf( "%T", ef )
        lastDot := strings.LastIndex( fqNm, "." )
        simplNm := fqNm[ lastDot + 1 : ]
        b.setVal( fmt.Sprintf( "%s/%s", simplNm, ef ), ef )
    }
    set( id( "id1" ) )
    set( id( "id1-id2" ) )
    set( MustNamespace( "ns1@v1" ) )
    set( MustNamespace( "ns1:ns2@v1" ) )
    set( MustDeclaredTypeName( "T1" ) )
    set( MustQualifiedTypeName( "ns1:ns2@v1/T1" ) )
    set( MustTypeReference( `String~"a"` ) )
    set( MustTypeReference( `String~["a","b"]` ) )
    set( MustTypeReference( 
         `Timestamp~["2012-01-01T00:00:00Z","2012-02-01T00:00:00Z"]` ) )
    set( MustTypeReference( "Int32~(0,10)" ) )
    set( MustTypeReference( "Int64~[0,10]" ) )
    set( MustTypeReference( "Uint32~(0,10)" ) )
    set( MustTypeReference( "Uint64~[0,10]" ) )
    set( MustTypeReference( "Float32~(0.0,1.0]" ) )
    set( MustTypeReference( "Float64~[0.0,1.0)" ) )
    set( MustTypeReference( "Float64~(,)" ) )
    set( MustTypeReference( "ns1@v1/T1" ) )
    set( MustTypeReference( "ns1@v1/T1+" ) )
    set( MustTypeReference( "ns1@v1/T1*" ) )
    set( MustTypeReference( "ns1@v1/T1*?" ) )
    set( MustTypeReference( "&ns1@v1/T1?" ) )
    set( MustTypeReference( "&ns1@v1/T1?+?*" ) )
}

func addBinIoRoundtripTests( tests []interface{} ) []interface{} {
    b := &binIoRoundtripTestBuilder{}
    b.nmCheck = map[ string ]interface{}{}
    b.tests = []interface{}{}
    b.addValueTests()
    b.addPathTests()
    b.addDefinitionTests()
    return b.tests
}

type BinIoSequenceRoundtripTest struct {
    Name string
    Seq []interface{}
}

func ( t *BinIoSequenceRoundtripTest ) AssertWriteValue(
    rd *BinReader, a *assert.PathAsserter ) {

    la := a.StartList()
    for _, val := range t.Seq {
        assertWriteValue( rd, val, la )
        la = la.Next()
    }
}

func addBinIoSequenceRoundtripTests( tests []interface{} ) []interface{} {
    return append( tests,
        &BinIoSequenceRoundtripTest{
            Name: "struct-sequence",
            Seq: []interface{}{
                MustStruct( "ns1@v1/S1" ),
                MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
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

func appendInput( data interface{}, w *BinWriter ) {
    switch v := data.( type ) {
    case *Identifier: 
        if err := w.WriteIdentifier( v ); err != nil { panic( err ) }
    case string: 
        if err := w.WriteValue( String( v ) ); err != nil { panic( err ) }
    case uint8: if err := w.WriteUint8( v ); err != nil { panic( err ) }
    case int32: if err := w.WriteInt32( v ); err != nil { panic( err ) }
    case uint64: if err := w.WriteUint64( v ); err != nil { panic( err ) }
    case *QualifiedTypeName: 
        if err := w.WriteQualifiedTypeName( v ); err != nil { panic( err ) }
    default: panic( libErrorf( "Unrecognized input elt: %T", v ) )
    }
}

func makeBinIoInvalidDataTest( data ...interface{} ) []byte {
    buf := &bytes.Buffer{}
    w := NewWriter( buf )
    for _, val := range data { appendInput( val, w ) }
    return buf.Bytes()
}

func addBinIoInvalidDataTests( tests []interface{} ) []interface{} {
    return append( tests, 
        &BinIoInvalidDataTest{
            Name: "unexpected-top-level-type-code",
            ErrMsg: "[offset 0]: Unrecognized value code: 0x64",
            Input: makeBinIoInvalidDataTest( binIoInvalidTypeCode ),
        },
        &BinIoInvalidDataTest{
            Name: "unexpected-symmap-val-type-code",
            ErrMsg: `[offset 39]: Unrecognized value code: 0x64`,
            Input: makeBinIoInvalidDataTest(
                tcStruct, int32( -1 ), MustQualifiedTypeName( "ns@v1/S" ),
                tcField, MustIdentifier( "f1" ), binIoInvalidTypeCode,
            ),
        },
        &BinIoInvalidDataTest{
            Name: "unexpected-list-val-type-code",
            ErrMsg: `[offset 57]: Unrecognized value code: 0x64`,
            Input: makeBinIoInvalidDataTest(
                tcStruct, int32( -1 ), MustQualifiedTypeName( "ns@v1/S" ),
                tcField, MustIdentifier( "f1" ),
                tcList, int32( -1 ), uint64( 1 ),
                tcInt32, int32( 10 ), // an okay list val
                binIoInvalidTypeCode,
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

func WriteBinIoTestValue( obj interface{}, w *BinWriter ) error {
    switch v := obj.( type ) {
    case Value: return w.WriteValue( v )
    case *Identifier: return w.WriteIdentifier( v )
    case objpath.PathNode: return w.WriteIdPath( v )
    case *Namespace: return w.WriteNamespace( v )
    case TypeName: return w.WriteTypeName( v )
    case TypeReference: return w.WriteTypeReference( v )
    }
    panic( libErrorf( "Unhandled expct obj: %T", obj ) )
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
