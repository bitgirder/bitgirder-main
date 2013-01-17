package testval

import (
    mg "mingle"
    "fmt"
)

var TestStruct1Inst1 *mg.Struct
var Timestamp1 mg.Timestamp
var Timestamp2 mg.Timestamp
var TestEnum1Inst1 *mg.Enum
var List1 *mg.List
var SymbolMap1 *mg.SymbolMap
var TypeCovStruct1 *mg.Struct

var Buffer1 mg.Buffer

func init() {
    Buffer1 = mg.Buffer( make( []byte, 150 ) )
    for i := 0; i < len( Buffer1 ); i++ { Buffer1[ i ] = byte( i ) }
}

func init() {
    TestEnum1Inst1 = mg.MustEnum( "mingle:test@v1/TestEnum1", "constant1" )
    Timestamp1 = mg.MustTimestamp( "2007-08-24T13:15:43.123450000-08:00" )
    Timestamp2 = mg.MustTimestamp( "2007-08-24T13:15:43-08:00" )
    list1Vals := make( []interface{}, 5 )
    for i, e := 0, len( list1Vals ); i < e; i++ {
        list1Vals[ i ] = mg.String( fmt.Sprintf( "string%d", i ) )
    }
    List1 = mg.MustList( list1Vals... )
    SymbolMap1 = mg.MustSymbolMap(
        "string-sym1", "something to do here",
        "int-sym1", int64( 1234 ),
        "decimal-sym1", float64( 3.14 ),
        "bool-sym1", false,
        "list-sym1", List1,
    )
    TestStruct1Inst1 = mg.MustStruct( "mingle:test@v1/TestStruct1",
        "string1", "hello",
        "bool1", true,
        "int1", int64( 32234 ),
        "int2", int64( 9223372036854775807 ),
        "int3", int32( 2147483647 ),
        "double1", float64( 1.1 ),
        "float1", float32( 1.1 ),
        "buffer1", Buffer1,
        "enum1", TestEnum1Inst1,
        "timestamp1", Timestamp1,
        "timestamp2", Timestamp2,
        "list1", List1,
        "symbol-map1", SymbolMap1,
        "struct1", 
            mg.MustStruct( "mingle:test@v1/TestStruct2", "i1", int32( 111 ) ),
    )
    TypeCovStruct1 = mg.MustStruct( "mingle:test@v1/TypeCov",
        "f1", "hello",
        "f2", true,
        "f3", int32( 1 ),
        "f4", int64( 1 ),
        "f5", uint32( 1 ),
        "f6", uint64( 1 ),
        "f7", float32( 1 ),
        "f8", float64( 1 ),
        "f9", Buffer1,
        "f10", TestEnum1Inst1,
        "f11", Timestamp1,
        "f12", mg.MustList(),
        "f13", mg.MustList( int32( 1 ), "hello", mg.MustList( true ) ),
        "f14", mg.MustSymbolMap(),
        "f15", mg.MustSymbolMap(
            "k1", "hello",
            "k2", mg.MustList( "a", "b", "c" ),
            "k3", mg.MustSymbolMap( "kk1", true ),
        ),
        "f16", TestStruct1Inst1,
    )
}
