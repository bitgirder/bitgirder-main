package stub

import (
    "testing"
)

func TestStubUnimplemented( t *testing.T ) {
    err := Unimplemented()
    expct := "unimplemented: bitgirder/stub.TestStubUnimplemented"
    if msg := err.Error(); msg != expct {
        t.Fatalf( "unexpected err: %q", msg )
    }
}
