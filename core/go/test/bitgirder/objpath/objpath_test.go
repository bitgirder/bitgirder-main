package objpath

import (
    "testing"
    "reflect"
    "errors"
)

var visErr = errors.New( "visitor error" )

func assertDeepEq( expct, act interface{}, t *testing.T ) {
    if ! reflect.DeepEqual( expct, act ) {
        t.Fatalf( "Expected %#v but got %#v", expct, act )
    }
}

type pathBuildVisitor struct {
    p PathNode
    calls int
    t *testing.T
}

func ( v *pathBuildVisitor ) takeCall() error {
    if v.calls < 0 { return nil }
    if v.calls == 0 { v.t.Fatal( "No calls remain" ) }
    v.calls--
    if v.calls == 0 { return visErr }
    return nil
}

func ( v *pathBuildVisitor ) Descend( elt interface{} ) error {
    if err := v.takeCall(); err != nil { return err }
    if v.p == nil { v.p = RootedAt( elt ) } else { v.p = v.p.Descend( elt ) }
    return nil
}

func ( v *pathBuildVisitor ) List( idx int ) error {
    if err := v.takeCall(); err != nil { return err }
    var ln *ListNode
    if v.p == nil { ln = RootedAtList() } else { ln = v.p.StartList() }
    for i := 0; i < idx; i++ { ln = ln.Next() }
    v.p = ln
    return nil
}

func assertVisit( p PathNode, t *testing.T ) PathNode {
    b := &pathBuildVisitor{ t: t, calls: -1 }
    Visit( p, b )
    assertDeepEq( p, b.p, t )
    return p
}

func TestVisit( t *testing.T ) {
    p1 := assertVisit( RootedAt( "a" ), t )
    p1 = assertVisit( p1.Descend( "b" ), t )
    p1 = assertVisit( p1.StartList(), t )
    p1 = assertVisit( p1.( *ListNode ).Next(), t )
    p1 = assertVisit( p1.( *ListNode ).Next(), t )
    p1 = assertVisit( p1.StartList(), t )
    p1 = assertVisit( p1.( *ListNode ).Next(), t )
    p1 = assertVisit( p1.Descend( "c" ), t )
}

func TestVisitAbortOnError( t *testing.T ) {
    f := func( p PathNode, calls int ) {
        b := &pathBuildVisitor{ t: t, calls: calls }
        if err := Visit( p, b ); err == nil {
            t.Fatal( "Expected err" )
        } else { 
            assertDeepEq( visErr, err, t ) 
            assertDeepEq( 0, b.calls, t )
        }
    }
    f( RootedAt( 0 ).Descend( 1 ).Descend( 2 ).Descend( 3 ), 3 )
    f( RootedAt( 0 ).StartList().Next().Next().Next().Descend( 1 ), 3 )
}

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
                  SetIndex( 8 ).
                  Descend( eltType( "p2" ) ),
            expct: "p1[ 8 ].p2",
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

// We assume for the purposes of this test that StringDotFormatter, as tested
// elsewhere, is correct.
func TestDescendAndStartList( t *testing.T ) {
    np1 := RootedAt( "p1" )
    chk := func( p PathNode, expct string ) {
        if act := Format( p, StringDotFormatter ); act != expct {
            t.Fatalf( "expected %q but got %q", expct, act )
        }
    }
    chk( Descend( nil, "p1" ), "p1" )
    chk( Descend( np1, "p2" ), "p1.p2" )
    chk( StartList( nil ), "[ 0 ]" )
    chk( StartList( np1 ).Next(), "p1[ 1 ]" )
}

func TestParentOfUtilMethod( t *testing.T ) {
    if p := ParentOf( nil ); p != nil { t.Fatalf( "ParentOf( nil ) is %v", p ) }
    p1 := RootedAt( "p1" )
    chk := func( p PathNode ) {
        if par := ParentOf( p ); par != p1 {
            t.Fatalf( "Parent of %v is not %v: %v", p, p1, par )
        }
    }
    chk( p1.Descend( "p2" ) )
    chk( p1.StartList() )
    chk( p1.StartList().SetIndex( 3 ) )
    chk( p1.StartList().Next() )
}
