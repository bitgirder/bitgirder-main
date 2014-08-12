package reactor

import (
    "testing"
    mg "mingle"
    "mingle/parser"
    "bitgirder/assert"
    "bitgirder/objpath"
)

func TestReactors( t *testing.T ) {
    RunReactorTestsInNamespace( reactorTestNs, t )
}

func TestVisitPath( t *testing.T ) {
    la := assert.NewListPathAsserter( t )
    p := mg.MakeTestIdPath
    chk := func( 
        path objpath.PathNode, val mg.Value, evs ...EventExpectation ) {

        rct := NewEventPathCheckReactor( evs, la ) 
        if err := VisitValuePath( val, rct, path ); err != nil { 
            la.Fatal( err ) 
        }
        rct.Complete()
        la = la.Next()
    }
    chk( 
        nil,
        mg.Int32( 1 ), 
        EventExpectation{
            Event: NewValueEvent( mg.Int32( 1 ) ),
        },
    )
    chk(
        p( 1 ),
        mg.Int32( 1 ),
        EventExpectation{
            Event: NewValueEvent( mg.Int32( 1 ) ),
            Path: p( 1 ),
        },
    )
    chk(
        p( 100 ),
        parser.MustStruct( "ns1@v1/S1", "f1", int32( 1 ) ),
        EventExpectation{
            Event: NewStructStartEvent( mkQn( "ns1@v1/S1" ) ),
            Path: p( 100 ),
        },
        EventExpectation{
            Event: NewFieldStartEvent( mkId( "f1" ) ),
            Path: p( 100, 1 ),
        },
        EventExpectation{
            Event: NewValueEvent( mg.Int32( 1 ) ),
            Path: p( 100, 1 ),
        },
        EventExpectation{
            Event: NewEndEvent(),
            Path: p( 100 ),
        },
    )
    chk(
        p( 100 ),
        parser.MustSymbolMap( "f1", int32( 1 ) ),
        EventExpectation{
            Event: NewMapStartEvent(),
            Path: p( 100 ),
        },
        EventExpectation{
            Event: NewFieldStartEvent( mkId( "f1" ) ),
            Path: p( 100, 1 ),
        },
        EventExpectation{
            Event: NewValueEvent( mg.Int32( 1 ) ),
            Path: p( 100, 1 ),
        },
        EventExpectation{
            Event: NewEndEvent(),
            Path: p( 100 ),
        },
    )
    chk(
        p( 100 ),
        mg.MustList( int32( 1 ) ),
        EventExpectation{
            Event: NewListStartEvent( mg.TypeOpaqueList ),
            Path: p( 100 ),
        },
        EventExpectation{
            Event: NewValueEvent( mg.Int32( 1 ) ),
            Path: p( 100, "0" ),
        },
        EventExpectation{
            Event: NewEndEvent(),
            Path: p( 100 ),
        },
    )
}

func TestTypeOfEvent( t *testing.T ) {
    a := assert.Asserter{ t }
    chk := func( ev ReactorEvent, typ mg.TypeReference ) {
        a.Truef( typ.Equals( TypeOfEvent( ev ) ), "blah" )
    }
    chk( NewValueEvent( mg.Int32( 1 ) ), mg.TypeInt32 )
    lt1 := parser.MustTypeReference( "ns1@v1/S1*" ).( *mg.ListTypeReference )
    chk( NewListStartEvent( lt1 ), lt1 )
    chk( NewMapStartEvent(), mg.TypeSymbolMap )
    st1 := parser.MustQualifiedTypeName( "ns1@v1/S1" )
    chk( NewStructStartEvent( st1 ), st1.AsAtomicType() )
}
