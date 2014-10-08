package bind

import (
    mg "mingle"
    "bitgirder/objpath"
)

var domainPackageBindTest = mkId( "package-bind-test" )

type BindTestDirection int

const (
    BindTestDirectionRoundtrip = iota
    BindTestDirectionIn
    BindTestDirectionOut
)

func ( d BindTestDirection ) Includes( d2 BindTestDirection ) bool {
    return d == d2 || d == BindTestDirectionRoundtrip
}

type BindTest struct {
    Mingle mg.Value
    BoundId *mg.Identifier
    Direction BindTestDirection
    StartPath objpath.PathNode
    Type mg.TypeReference
    StrictTypeMatching bool
    Domain *mg.Identifier
    SerialOptions *SerialOptions
    Error error
}
