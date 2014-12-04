package codegen

import (
    mg "mingle"
    "mingle/parser"
    "bitgirder/assert"
    "testing"
    "strings"
)

var (
    mkNs = parser.MustNamespace
)

func TestDefaultPathMapper( t *testing.T ) {
    la := assert.NewListPathAsserter( t )
    type input struct { nsStr string; expct string }
    chk := func( in input ) {
        ns := mkNs( in.nsStr )
        ids, err := DefaultPathMapper.MapPath( ns )
        if err != nil { la.Fatal( err ) }
        idStrs := make( []string, len( ids ) )
        for i, id := range ids { idStrs[ i ] = id.Format( mg.LcCamelCapped ) }
        act := strings.Join( idStrs, "/" )
        la.Equal( in.expct, act )
    }
    for _, in := range []input{
        { "ns1@v1", "ns1/v1" },
        { "ns1:ns2@v1", "ns1/v1/ns2" },
        { "ns1:ns2:ns3@v1", "ns1/v1/ns2/ns3" },
    } {
        chk( in )
        la = la.Next()
    }
}
