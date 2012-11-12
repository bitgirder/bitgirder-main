package objpath

import (
    "testing"
)

type eltType string

type fmtTest struct {
    path PathNode
    expct string
}

func ( ft *fmtTest ) call( t *testing.T ) {
    fmtr := DotFormatter( func( elt interface{}, apnd AppendFunc ) {
        apnd( string( elt.( eltType ) ) )
    })
    str := Format( ft.path, fmtr )
    if str != ft.expct { t.Fatalf( "expected %#v but got %#v", ft.expct, str ) }
}

func TestPathFormatter( t *testing.T ) {
    fmtTests := []*fmtTest {

        &fmtTest{ path: RootedAt( eltType( "p1" ) ), expct: "p1" },

        &fmtTest{ path: RootedAtList(), expct: "[ 0 ]" },
        
        &fmtTest{ 
            path: RootedAt( eltType( "p1" ) ).
                  Descend( eltType( "p2" ) ).
                  Descend( eltType( "p3" ) ),
            expct: "p1.p2.p3",
        },

        &fmtTest{
            path: RootedAt( eltType( "p1" ) ).
                  StartList().
                  Next().
                  Next().
                  Descend( eltType( "p2" ) ).
                  Descend( eltType( "p3" ) ).
                  StartList(),
            expct: "p1[ 2 ].p2.p3[ 0 ]",
        },

        &fmtTest{
            path: RootedAt( eltType( "p1" ) ).
                  StartList().
                  Next().
                  StartList().
                  StartList().
                  Next(),
            expct: "p1[ 1 ][ 0 ][ 1 ]",
        },

        &fmtTest{
            path: RootedAtList().Next().Descend( eltType( "p1" ) ),
            expct: "[ 1 ].p1",
        },
    }
    for _, ft := range fmtTests { ft.call( t ) }
}

type testString string

func ( ts testString ) String() string { return string( ts ) }

func TestStringDotFormatter( t *testing.T ) {
    f := func( p PathNode ) {
        if s := Format( p, StringDotFormatter ); s != "p1.p2[ 1 ].p3" {
            t.Fatalf( "Unexpected path: %q", s )
        }
    }
    f( RootedAt( "p1" ).Descend( "p2" ).StartList().Next().Descend( "p3" ) )
    f( RootedAt( testString( "p1" ) ).
       Descend( testString( "p2" ) ).
       StartList().
       Next().
       Descend( testString( "p3" ) ) )
    defer func() {
        if err := recover(); err != nil {
            if err.( error ).Error() != "Can't convert to string: int" {
                t.Fatal( err )
            }
        }
    }()
    Format( RootedAt( 1 ), StringDotFormatter )
}
