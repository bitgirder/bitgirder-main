package bind

import (
    mg "mingle"
    mgRct "mingle/reactor"
    "mingle/parser"
    "bitgirder/objpath"
)

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
    Bound interface{}
    Direction BindTestDirection
    Type mg.TypeReference
    Domain *mg.Identifier
    Error error
}

type S1 struct {
    f1 int32
}

func ( s *S1 ) VisitValue( 
    out mgRct.ReactorEventProcessor, 
    bc *BindContext, 
    path objpath.PathNode ) error {

    qn := mkQn( "ns1@v1/S1" )
    ss := mgRct.NewStructStartEvent( qn )
    if err := out.ProcessEvent( ss ); err != nil { return err }
    fs := mgRct.NewFieldStartEvent( mkId( "f1" ) )
    if err := out.ProcessEvent( fs ); err != nil { return err }
    ve := mgRct.NewValueEvent( mg.Int32( s.f1 ) )
    if err := out.ProcessEvent( ve ); err != nil { return err }
    ee := mgRct.NewEndEvent()
    return out.ProcessEvent( ee )
}

type E1 string 

const (
    E1V1 = E1( "v1" )
    E1V2 = E1( "v2" )
)

func ( e E1 ) VisitValue(
    out mgRct.ReactorEventProcessor,
    bc *BindContext,
    path objpath.PathNode ) error {

    me := parser.MustEnum( "ns1@v1/E1", string( e ) )
    ve := mgRct.NewValueEvent( me )
    return out.ProcessEvent( ve )
}
