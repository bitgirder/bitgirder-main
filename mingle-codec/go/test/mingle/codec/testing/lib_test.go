package testing

import (
    "mingle/codec"
    gotest "testing"
)

var structBuilderFactory = 
    func() codec.Reactor { return codec.NewStructBuilder() }

// This would ideally just be in mingle/codec tests, but that would create a
// circular dep between mingle/codec and mingle/codec/testing, so we put these
// tests here.
func TestBaseReactorImpl( t *gotest.T ) {
    TestReactorErrorSequences( structBuilderFactory, t )
}

// These may later prove useful for reactors other than the struct builder, but
// for now that is the only place we track and handle duplicate keys.
func TestStructBuilderMultipleKeyErrors( t *gotest.T ) {
    TestReactorErrorSequence( 
        structBuilderFactory,
        []string{ 
            "start-struct", 
            "start-field1", "value",
            "start-field2", "value",
            "start-field1", "value",
            "end",
        },
        "Invalid fields: Multiple entries for key: f1",
        t,
    )
}
