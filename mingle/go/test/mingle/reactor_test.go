package mingle

import (
    "testing"
    "bitgirder/assert"
)

func callTestStructReactors( ms *Struct, t *testing.T ) {
    rct := NewStructBuilder()
    if err := VisitStruct( ms, rct ); err == nil {
        assert.Equal( ms, rct.GetStruct() )
    } else { t.Fatal( err ) }
}

func TestStructBuilderReactor( t *testing.T ) {
    callTestStructReactors( 
        MustStruct( "ns1@v1/S1",
            "val1", String( "hello" ),
            "list1", MustList(),
            "map1", MustSymbolMap(),
            "struct1", MustStruct( "ns1@v1/S2" ),
        ),
        t,
    )
}

func TestStructBuilderReactorErrors( t *testing.T ) {
    tests := []*ReactorErrorTest{}
    tests = append( tests, StdReactorErrorTests... )
    tests = append( tests,
        &ReactorErrorTest{
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
    fact := func() Reactor { return NewStructBuilder() }
    CallReactorErrorTests( fact, tests, t )
}
