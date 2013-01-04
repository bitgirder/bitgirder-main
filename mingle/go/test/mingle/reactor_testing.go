package mingle

type ReactorTest struct {
    Source interface{}
    Error error
}

var StdReactorTests []*ReactorTest

type ValueBuildSource struct { Val Value }

func initValueBuildReactorTests() {
    s1 := MustStruct( "ns1@v1/S1",
        "val1", String( "hello" ),
        "list1", MustList(),
        "map1", MustSymbolMap(),
        "struct1", MustStruct( "ns1@v1/S2" ),
    )
    mk := func( v Value ) *ReactorTest {
        return &ReactorTest{ Source: ValueBuildSource{ v } }
    }
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

type StructuralReactorTestSource struct {
    Events []ReactorEvent
    TopType ReactorTopType
}

func initStructuralReactorTests() {
    evStartStruct1 := StructStartEvent{ MustTypeReference( "ns1@v1/S1" ) }
    evStartField1 := FieldStartEvent{ MustIdentifier( "f1" ) }
    evStartField2 := FieldStartEvent{ MustIdentifier( "f2" ) }
    evValue1 := ValueEvent{ Int64( int64( 1 ) ) }
    mk1 := func( errMsg string, evs ...ReactorEvent ) *ReactorTest {
        return &ReactorTest{
            Source: &StructuralReactorTestSource{ Events: evs },
            Error: rctError( errMsg ),
        }
    }
    mk2 := func( 
        errMsg string, tt ReactorTopType, evs ...ReactorEvent ) *ReactorTest {
        res := mk1( errMsg, evs... )
        res.Source.( *StructuralReactorTestSource ).TopType = tt
        return res
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
    )
}

func init() {
    StdReactorTests = []*ReactorTest{}
    initValueBuildReactorTests()
    initStructuralReactorTests()
}
