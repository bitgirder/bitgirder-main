package tck

import (
    mg "mingle"
    "mingle/bind"
    "mingle/parser"
    mgTck "mingle/tck"
    "bitgirder/objpath"
)

var (
    mkId = parser.MustIdentifier
    mkNs = parser.MustNamespace
    mkQn = parser.MustQualifiedTypeName
    asType = parser.AsTypeReference
)

func valueTestWithType( 
    mv mg.Value, idNm string, typ mg.TypeReference ) *bind.BindTest {

    return &bind.BindTest{
        Mingle: mv,
        BoundId: mkId( idNm ),
        Direction: bind.BindTestDirectionRoundtrip,
        Type: typ,
        Domain: bind.DomainDefault,
    }
}

func valueTest( mv mg.Value, idNm string ) *bind.BindTest {
    return valueTestWithType( mv, idNm, mg.TypeOf( mv ) )
}

type ValidationErrorTest struct {
    Message string
    Path objpath.PathNode
    Name string
}

var dataNs = parser.MustNamespace( "mingle:tck:data@v1" )

func dataQn( nm string ) *mg.QualifiedTypeName {
    return mkQn( dataNs.ExternalForm() + "/" + nm )
}

func dataStruct( nm string, pairs... interface{} ) *mg.Struct {
    return parser.MustStruct( dataQn( nm ), pairs... )
}

type testsBuilder struct {
    tests []interface{}
}

func ( b *testsBuilder ) addTests( tests... interface{} ) { 
    b.tests = append( b.tests, tests... )
}

func ( b *testsBuilder ) addEmptyStruct() {
    b.addTests( valueTest( dataStruct( "EmptyStruct" ), "empty-struct-inst1" ) )
}

func ( b *testsBuilder ) addScalarsBasic() {
    b.addTests( 
        valueTest(
            dataStruct( "ScalarsBasic",
                "stringF1", "hello",
                "bool1", true,
                "buffer1", []byte{ 0, 1, 2 },
                "int32F1", int32( 1 ),
                "int64F1", int64( 2 ),
                "uint32F1", uint32( 3 ),
                "uint64F1", uint64( 4 ),
                "float32F1", float32( 5.0 ),
                "float64F1", float64( 6.0 ),
                "timeF1", mgTck.Timestamp1,
            ),
            "scalars-basic-inst1",
        ),
    )
}

func scalarsRestrictInst1() *mg.Struct {
    return dataStruct( "ScalarsRestrict",
        "stringF1", "aaaa",
        "stringF2", "aaaaaaab",
        "int32F1", int32( 1 ),
        "uint32F1", int32( 2 ),
        "int64F1", int64( 3 ),
        "uint64F1", uint64( 4 ),
        "float32F1", float32( 0.5 ),
        "float64F1", float64( 0.6 ),
        "timeF1", mg.MustTimestamp( "2013-10-20T00:00:00Z" ),
    )
}

func ( b *testsBuilder ) addScalarsRestrict() {
    b.addTests(
        valueTest( scalarsRestrictInst1(), "scalars-restrict-inst1" ),
        &ValidationErrorTest{
            Message: "STUB",
            Name: "scalars-restrict-stringF1-rx-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "scalars-restrict-stringF2-rx-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "scalars-restrict-int32-rx-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "scalars-restrict-uint32-rx-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "scalars-restrict-int64-rx-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "scalars-restrict-uint64-rx-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "scalars-restrict-float32-rx-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "scalars-restrict-float64-rx-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "scalars-restrict-timeF1-rx-err",
        },
    )
}

func dataEnum1( val string ) *mg.Enum {
    return parser.MustEnum( dataQn( "Enum1" ), val )
}

func ( b *testsBuilder ) addEnum1Tests() {
    b.addTests(
        valueTest( dataEnum1( "const1" ), "enum1-const1" ),
        &ValidationErrorTest{
            Message: "no such constant: bad-val",
            Name: "enum1-no-such-constant",
        },
    )
}

func ( b *testsBuilder ) addEnumHolderTests() {
    b.addTests(
        valueTest(
            dataStruct( "EnumHolder", "enum1", dataEnum1( "const1" ) ),
            "enum-holder-inst1",
        ),
    )
}

func ( b *testsBuilder ) addMapHolderTests() {
    b.addTests(
        valueTest(
            dataStruct( "MapHolder",
                "mapF1", parser.MustSymbolMap( "f1", int32( 1 ) ),
            ),
            "map-holder-inst1",
        ),
        valueTest(
            dataStruct( "MapHolder",
                "mapF1", parser.MustSymbolMap( "f1", int32( 1 ) ),
                "mapF2", parser.MustSymbolMap( "f2", "string-val" ),
            ),
            "map-holder-inst2",
        ),
    )
}

func ( b *testsBuilder ) addUnionHolderTests() {
    b.addTests(
        valueTest(
            dataStruct( "UnionHolder", "union1F1", int32( 1 ) ),
            "union-holder-int32-inst",
        ),
        valueTest(
            dataStruct( "UnionHolder", "union1F1", scalarsRestrictInst1() ),
            "union-holder-scalars-restrict-inst",
        ),
        valueTest(
            dataStruct( "UnionHolder", "union1F1", dataEnum1( "const1" ) ),
            "union-holder-enum1-inst",
        ),
        valueTest(
            dataStruct( "UnionHolder", 
                "union1F1", parser.MustSymbolMap( "f1", int32( 1 ) ),
            ),
            "union-holder-map-inst",
        ),
        &ValidationErrorTest{
            Message: "bad type for union: Int64",
            Name: "union-holder-bad-union-type",
        },
    )
}

func ( b *testsBuilder ) addValueHolderTests() {
    b.addTests(
        valueTest(
            dataStruct( "ValueHolder", "valF1", int32( 1 ) ),
            "value-holder-int32",
        ),
    )
}

func ( b *testsBuilder ) addMissingFieldTests() {
    b.addTests(
        &ValidationErrorTest{
            Message: "missing field: stringF1",
            Name: "scalars-basic-missing-fields-string-f1",
        },
        &ValidationErrorTest{
            Message: "missing fields: int32F1, stringF1",
            Name: "scalars-basic-missing-fields-int32-f1-string-f1",
        },
    )
}

func ( b *testsBuilder ) addScalarFieldDefaults() {
    b.addTests(
        valueTest(
            dataStruct( "ScalarFieldDefaults",
                "boolF1", true,
                "stringF1", "abc",
                "int32F1", int32( 1 ),
                "uint32F1", uint32( 2 ),
                "int64F1", int64( 3 ),
                "uint64F1", uint64( 4 ),
                "float32F1", float32( 5.0 ),
                "float64F1", float64( 6.0 ),
                "enum1F1", dataEnum1( "const2" ),
                "timeF1", mg.MustTimestamp( "2014-10-19T00:00:00Z" ),
            ),
            "scalar-field-defaults-with-defaults",
        ),
        valueTest(
            dataStruct( "ScalarFieldDefaults",
                "boolF1", false,
                "stringF1", "dddd",
                "int32F1", int32( 11 ),
                "uint32F1", uint32( 12 ),
                "int64F1", int64( 13 ),
                "uint64F1", uint64( 14 ),
                "float32F1", float32( 15.0 ),
                "float64F1", float64( 16.0 ),
                "enum1F1", dataEnum1( "const3" ),
                "timeF1", mg.MustTimestamp( "2014-10-20T00:00:00Z" ),
            ),
            "scalar-field-defaults-inst1",
        ),
    )
}

func ( b *testsBuilder ) addSchema1Tests() {
    b.addTests(
        valueTest(
            dataStruct( "Struct2", "f1", int32( 1 ), "f2", "abc" ),
            "struct2-inst1",
        ),
    )
}

func ( b *testsBuilder ) addPointerStruct1Tests() {
    b.addTests(
        valueTest(
            dataStruct( "PointerStruct1",
                "struct1F1", dataStruct( "Struct1",
                    "f1", int32( 1 ),
                    "f2", "abc",
                ),
                "int32F1", int32( 2 ),
            ),
            "pointer-struct1-inst1",
        ),
    )
}

func ( b *testsBuilder ) addNullablesTests() {
    b.addTests(
        valueTest( dataStruct( "Nullables" ), "nullables-inst1" ),
        valueTest(
            dataStruct( "Nullables",
                "boolF1", true,
                "bufferF1", []byte{ 0 },
                "int32F1", int32( 1 ),
                "int64F1", int64( 2 ),
                "uint32F1", uint32( 3 ),
                "uint64F1", uint64( 4 ),
                "float32F1", float32( 5 ),
                "float64F1", float64( 6 ),
                "timeF1", mg.MustTimestamp( "2013-10-20T00:00:00Z" ),
                "stringF1", "abc",
                "mapF1", parser.MustSymbolMap( "f1", int32( 1 ) ),
                "valF1", int32( 1 ),
                "enum1PtrF1", dataEnum1( "const1" ),
                "union1PtrF1", int32( 2 ),
                "struct1F1", dataStruct( "Struct1",
                    "f1", int32( 3 ),
                    "f2", "def",
                ),
                "schemaF1", dataStruct( "Struct2",
                    "f1", int32( 4 ),
                    "f2", "ghi",
                ),
                "int32PtrF1", int32( 5 ),
                "int32ListF1", mg.MustList( asType( "Int32*" ), int32( 6 ) ), 
            ),
            "nullables-inst2",
        ),
    )
}
 
func ( b *testsBuilder ) addLists1Tests() {
    b.addTests(
        valueTest(
            dataStruct( "Lists1",
                "int32ListF1", mg.MustList(
                    asType( "Int32*" ), int32( 1 ), int32( 2 ) ),
                "mapListF1", mg.MustList(
                    asType( "SymbolMap*" ), 
                    parser.MustSymbolMap( "f1", int32( 1 ) ),
                    parser.MustSymbolMap( "f2", int32( 2 ) ),
                ),
                "union1ListF1", mg.MustList(
                    asType( "&mingle:tck:data@v1/Union1?*" ),
                    int32( 1 ),
                    scalarsRestrictInst1(),
                    mg.NullVal,
                ),
                "schema1ListF1", mg.MustList(
                    asType( "mingle:tck:data@v1/Schema1*" ),
                    dataStruct( "Struct1", "f1", int32( 1 ), "f2", "abc" ),
                    dataStruct( "Struct2", "f1", int32( 2 ), "f2", "def" ),
                ),
                "struct1List1F1", mg.MustList(
                    asType( "mingle:tck:data@v1/Struct1*" ),
                    dataStruct( "Struct1", "f1", int32( 1 ), "f2", "abc" ),
                    dataStruct( "Struct1", "f1", int32( 2 ), "f2", "def" ),
                ),
                "enum1ListF1", mg.MustList(
                    asType( "mingle:tck:data@v1/Enum1+" ),
                    dataEnum1( "const1" ),
                    dataEnum1( "const2" ),
                ),
                "int64PtrListF1", mg.MustList(
                    asType( "&Int64*" ), int64( 1 ), int64( 2 ) ),
                "valueListF1", mg.MustList(
                    asType( "Value*" ), int32( 1 ), "abc" ),
                "nullValueListF1", mg.MustList(
                    asType( "Value?*" ), int32( 1 ), mg.NullVal ),
                "valPtrListF1", mg.MustList(
                    asType( "&Value*" ), int32( 1 ), "abc" ),
                "int32ListPtrF1", mg.MustList(
                    asType( "Int32*" ), int32( 1 ), int32( 2 ) ),
                "stringListListF1", mg.MustList(
                    asType( "String**" ), 
                    mg.MustList( asType( "String*" ), "abc", "def" ),
                    mg.MustList( asType( "String*" ), "ghi", "jkl" ),
                ),
            ),
            "lists1-inst1",
        ),
        &ValidationErrorTest{
            Message: "STUB",
            Name: "list1-int32-list-f1-elt-type-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "list1-int32-list-f1-null-elt-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "list1-map-list-f1-elt-type-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "list1-union-list-f1-elt-type-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "list1-schema1-list-f1-elt-type-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "list1-struct1-list1-f1-elt-type-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "list1-enum1-list1-f1-elt-type-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "list1-int64-ptr-list-f1-elt-type-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "list1-value-list-f1-null-elt-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "list1-val-ptr-list-f1-null-elt-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "list1-int32-list-ptr-f1-elt-type-err",
        },
        &ValidationErrorTest{
            Message: "STUB",
            Name: "list1-string-list-list-f1-elt-type-err",
        },
    )
}

func ( b *testsBuilder ) addListDefaultsTests() {
    b.addTests(
        valueTest(
            dataStruct( "ListDefaults",
                "int32ListF1", mg.MustList(
                    asType( "Int32*" ), int32( -1 ), int32( -2 ), int32( -3 ) ),
                "int64ListF1", mg.MustList(
                    asType( "Int64*" ), int64( -6 ), int64( -5 ), int64( -4 ) ),
                "uint32ListF1", mg.MustList(
                    asType( "Uint32*" ),
                    uint32( 0 ),
                    uint32( 10 ),
                    uint32( 4294967295 ),
                ),
                "uint64ListF1", mg.MustList(
                    asType( "Uint64*" ),
                    uint64( 20 ),
                    uint64( 30 ),
                    uint64( 18446744073709551615 ),
                ),
                "float32ListF1", mg.MustList(
                    asType( "Float32*" ), float32( 0.0 ), float32( -1.0 ) ),
                "float64ListF1", mg.MustList(
                    asType( "Float64*" ), float64( -2.0 ), float64( 3.0 ) ),
                "stringListF1", mg.MustList(
                    asType( "String*" ), "a", "b", "c" ),
                "timeListF1", mg.MustList(
                    asType( "Timestamp*" ),
                    mg.MustTimestamp( "2014-10-19T00:00:00Z" ),
                    mg.MustTimestamp( "2014-10-20T00:00:00Z" ),
                    mg.MustTimestamp( "2014-10-21T00:00:00Z" ),
                ),
                "enum1ListF1", mg.MustList(
                    asType( "mingle:tck:data@v1/Enum1*" ),
                    dataEnum1( "const1" ),
                    dataEnum1( "const2" ),
                    dataEnum1( "const1" ),
                ),
            ),
            "list-defaults-inst1",
        ),
        valueTest(
            dataStruct( "ListDefaults",
                "int32ListF1", mg.MustList(
                    asType( "Int32*" ), int32( 1 ), int32( 2 ), int32( 3 ) ),
                "int64ListF1", mg.MustList(
                    asType( "Int64*" ), int64( 6 ), int64( 5 ), int64( 4 ) ),
                "uint32ListF1", mg.MustList(
                    asType( "Uint32*" ),
                    uint32( 11 ),
                ),
                "uint64ListF1", mg.MustList(
                    asType( "Uint64*" ),
                    uint64( 21 ),
                ),
                "float32ListF1", mg.MustList(
                    asType( "Float32*" ), float32( 10.0 ), float32( 11.0 ) ),
                "float64ListF1", mg.MustList(
                    asType( "Float64*" ), float64( 12.0 ), float64( 13.0 ) ),
                "stringListF1", mg.MustList( asType( "String*" ), "d", "e" ),
                "timeListF1", mg.MustList(
                    asType( "Timestamp*" ),
                    mg.MustTimestamp( "2015-10-19T00:00:00Z" ),
                    mg.MustTimestamp( "2015-10-20T00:00:00Z" ),
                ),
                "enum1ListF1", mg.MustList(
                    asType( "mingle:tck:data@v1/Enum1*" ),
                    dataEnum1( "const2" ),
                    dataEnum1( "const3" ),
                ),
            ),
            "list-defaults-inst2",
        ),
    )
}

func ( b *testsBuilder ) addData2Tests() {
    b.addTests(
        valueTest(
            parser.MustStruct( "mingle:tck:data2@v1/Struct2",
                "f1", parser.MustStruct( "mingle:tck:data@v1/Struct1",
                    "f1", int32( 1 ),
                    "f2", "abc",
                ),
                "f2", parser.MustStruct( "mingle:tck:data2@v1/Struct1",
                    "f1", int32( 2 ),
                ),
            ),
            "data2-struct2-inst1",
        ),
    )
}

func GetTckTests() []interface{} {
    b := &testsBuilder{ tests: make( []interface{}, 0, 256 ) }
    b.addEmptyStruct()
    b.addScalarsBasic()
    b.addScalarsRestrict()
    b.addEnum1Tests()
    b.addEnumHolderTests()
    b.addMapHolderTests()
    b.addUnionHolderTests()
    b.addValueHolderTests()
    b.addScalarFieldDefaults()
    b.addSchema1Tests()
    b.addPointerStruct1Tests()
    b.addNullablesTests()
    b.addLists1Tests()
    b.addListDefaultsTests()
    b.addData2Tests()
    b.addMissingFieldTests()
    return b.tests
}
