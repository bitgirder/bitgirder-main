package bind

import (
    mg "mingle"
)

const testMsgErrorBadValue = "test-message-error-bad-value"

type TestProfile string

const (
    TestProfileDefaultValue = TestProfile( "default-value" )
    TestProfileCustomValue = TestProfile( "custom-value" )
)

type BindTest struct {
    In mg.Value
    Expect interface{}
    Type mg.TypeReference
    Profile TestProfile
    Error error
}

type boundMap map[ string ]interface{}

type boundList []interface{}

const s1F1ValFailOnProduce = int32( 100 )

type S1 struct {
    f1 int32
}
