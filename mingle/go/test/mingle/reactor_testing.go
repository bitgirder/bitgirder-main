package mingle

import ( 
    "testing"
    "bitgirder/assert"
//    "log"
)

type ReactorErrorTest struct {
    Seq []string
    ErrMsg string
}

type ReactorErrorTestFactory func() Reactor

var StdReactorErrorTests = []*ReactorErrorTest{
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
      ErrMsg: "Expected top-level struct start but got value",
    },
    { Seq: []string{ "start-list" },
      ErrMsg: "Expected top-level struct start but got list start",
    },
    { Seq: []string{ "start-map" },
      ErrMsg: "Expected top-level struct start but got map start",
    },
    { Seq: []string{ "start-field1" },
      ErrMsg: "Expected top-level struct start but got field 'f1'",
    },
    { Seq: []string{ "end" },
      ErrMsg: "Expected top-level struct start but got end",
    },
    { Seq: []string{ 
        "start-struct", 
        "start-field1", "value",
        "start-field2", "value",
        "start-field1", "value",
        "end",
      },
      ErrMsg: "Invalid fields: Multiple entries for key: f1",
    },
}

type reactorErrorTestCall struct {
    *assert.PathAsserter
    test *ReactorErrorTest
    fact ReactorErrorTestFactory
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
    rct := t.fact()
    if err := t.feedSequence( rct ); err == nil {
        t.Fatalf( "Got no err for %#v", t.test.Seq )
    } else {
        if _, ok := err.( *ReactorError ); ok {
            assert.Equal( t.test.ErrMsg, err.Error() )
        } else { t.Fatal( err ) }
    }
}

func CallReactorErrorTests( 
    f ReactorErrorTestFactory, tests []*ReactorErrorTest, t *testing.T ) {
    a := assert.NewPathAsserter( t ).StartList()
    for _, test := range tests {
        ( &reactorErrorTestCall{ PathAsserter: a, test: test, fact: f } ).call()
        a = a.Next()
    }
}
