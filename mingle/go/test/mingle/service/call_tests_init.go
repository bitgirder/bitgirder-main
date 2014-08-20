package service

import (
    mg "mingle"
)

func GetTckTestCalls() []*TckTestCall {
    return []*TckTestCall{
        &TckTestCall{
            Context: &RequestContext{
                Namespace: mkNs( "mingle:tck@v1" ),
                Service: mkId( "svc1" ),
                Operation: mkId( "getFixedInt" ),
            },
            Expect: &ResultExpectation{ Result: mg.Int32( 1 ) },
        },
    }
}
