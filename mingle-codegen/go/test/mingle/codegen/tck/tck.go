package tck

import (
    mg "mingle"
    "mingle/parser"
    mgTck "mingle/tck"
)

type ValueTest struct {
    Mingle mg.Value
    Name string
}

type ValidationErrorTest struct {
    Message string
    Name string
}

var (
    mkQn = parser.MustQualifiedTypeName
)

func dataQn( nm string ) *mg.QualifiedTypeName {
    return mkQn( "mingle:tck:data@v1/" + nm )
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

func ( b *testsBuilder ) addScalarsBasic() {
    b.addTests( 
        &ValueTest{
            Mingle: dataStruct( "ScalarsBasic",
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
            Name: "scalars-basic-inst1",
        },
    )
}

func ( b *testsBuilder ) addErrorTests() {
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

func GetTckTests() []interface{} {
    b := &testsBuilder{ tests: make( []interface{}, 0, 256 ) }
    b.addScalarsBasic()
    b.addErrorTests()
    return b.tests
}
