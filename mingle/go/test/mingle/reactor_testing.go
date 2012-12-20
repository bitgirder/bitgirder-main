package mingle

import ( 
    "testing"
    "bitgirder/assert"
)

type ReactorSeqErrorTest struct {
    Seq []string
    ErrMsg string
    TopType ReactorTopType
}

var StdReactorSeqErrorTests = []*ReactorSeqErrorTest{
    { Seq: []string{ "start-struct", "start-field1", "start-field2" }, 
      ErrMsg: "Saw start of field 'f2' while expecting a value for 'f1'",
    },
    { Seq: []string{ 
        "start-struct", "start-field1", "start-map", "start-field1",
        "start-field2" },
      ErrMsg: "Saw start of field 'f2' while expecting a value for 'f1'",
    },
    { Seq: []string{ "start-struct", "end", "start-field1" },
      ErrMsg: "StartField() called, but struct is built",
    },
    { Seq: []string{ "start-struct", "value" },
      ErrMsg: "Expected field name or end of fields but got value",
    },
    { Seq: []string{ "start-struct", "start-list" },
      ErrMsg: "Expected field name or end of fields but got list start",
    },
    { Seq: []string{ "start-struct", "start-map" },
      ErrMsg: "Expected field name or end of fields but got map start",
    },
    { Seq: []string{ "start-struct", "start-struct" },
      ErrMsg: "Expected field name or end of fields but got struct start",
    },
    { Seq: []string{ "start-struct", "start-field1", "end" },
      ErrMsg: "Saw end while expecting value for field 'f1'",
    },
    { Seq: []string{ 
        "start-struct", "start-field1", "start-list", "start-field1" },
      ErrMsg: "Expected list value but got start of field 'f1'",
    },
    { Seq: []string{ "value" },
      ErrMsg: "Expected struct but got value",
      TopType: ReactorTopTypeStruct,
    },
    { Seq: []string{ "start-list" },
      ErrMsg: "Expected struct but got list start",
      TopType: ReactorTopTypeStruct,
    },
    { Seq: []string{ "start-map" },
      ErrMsg: "Expected struct but got map start",
      TopType: ReactorTopTypeStruct,
    },
    { Seq: []string{ "start-field1" },
      ErrMsg: "Expected struct but got field 'f1'",
      TopType: ReactorTopTypeStruct,
    },
    { Seq: []string{ "end" },
      ErrMsg: "Expected struct but got end",
      TopType: ReactorTopTypeStruct,
    },
}

type reactorErrorTestCall struct {
    *assert.PathAsserter
    test *ReactorSeqErrorTest
}

func ( t *reactorErrorTestCall ) feedCall( call string, rct Reactor ) error {
    switch call {
    case "start-struct": 
        return rct.StartStruct( MustTypeReference( "ns1@v1/S1" ) )
    case "start-map": return rct.StartMap()
    case "start-list": return rct.StartList()
    case "start-field1": return rct.StartField( MustIdentifier( "f1" ) )
    case "start-field2": return rct.StartField( MustIdentifier( "f2" ) )
    case "value": return rct.Value( Int64( int64( 1 ) ) )
    case "end": return rct.End()
    }
    panic( libErrorf( "Unexpected test call: %s", call ) )
}

func ( t *reactorErrorTestCall ) feedSequence( rct Reactor ) error {
    for _, call := range t.test.Seq {
        if err := t.feedCall( call, rct ); err != nil { return err }
    }
    return nil
}

func ( t *reactorErrorTestCall ) call() {
    rct := NewValueBuilder()
    rct.SetTopType( t.test.TopType )
    if err := t.feedSequence( rct ); err == nil {
        t.Fatalf( "Got no err for %#v", t.test.Seq )
    } else {
        if _, ok := err.( *ReactorError ); ok {
            t.Equal( t.test.ErrMsg, err.Error() )
        } else { t.Fatal( err ) }
    }
}

func CallReactorSeqErrorTests( tests []*ReactorSeqErrorTest, t *testing.T ) {
    a := assert.NewPathAsserter( t ).StartList()
    for _, test := range tests {
        ( &reactorErrorTestCall{ PathAsserter: a, test: test } ).call()
        a = a.Next()
    }
}
