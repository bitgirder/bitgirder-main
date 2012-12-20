package mingle

import (
    "testing"
    "bitgirder/assert"
)

func assertValueReactorRoundtrip( v Value, a *assert.PathAsserter ) {
    rct := NewValueBuilder()
    rct.SetTopType( ReactorTopTypeValue )
    if err := VisitValue( v, rct ); err == nil {
        a.Equal( v, rct.GetValue() )
    } else { a.Fatal( err ) }
}

func TestValueBuilderReactor( t *testing.T ) {
    a := assert.NewPathAsserter( t ).StartList()
    s1 := MustStruct( "ns1@v1/S1",
        "val1", String( "hello" ),
        "list1", MustList(),
        "map1", MustSymbolMap(),
        "struct1", MustStruct( "ns1@v1/S2" ),
    )
    for _, val := range []Value{
        String( "hello" ),
        MustList(),
        MustList( 1, 2, 3 ),
        MustList( 1, MustList(), MustList( 1, 2 ) ),
        MustSymbolMap(),
        MustSymbolMap( "f1", "v1", "f2", MustList(), "f3", s1 ),
        s1,
        MustStruct( "ns1@v1/S3" ),
    } {
        assertValueReactorRoundtrip( val, a )
        a = a.Next()
    }
}

func TestValueBuilderReactorErrors( t *testing.T ) {
    tests := []*ReactorSeqErrorTest{}
    tests = append( tests, StdReactorSeqErrorTests... )
    tests = append( tests,
        &ReactorSeqErrorTest{
            Seq: []string{ 
                "start-struct", 
                "start-field1", "value",
                "start-field2", "value",
                "start-field1", "value",
                "end",
            },
            ErrMsg: "Invalid fields: Multiple entries for key: f1",
        },
    )
    CallReactorSeqErrorTests( tests, t )
}
