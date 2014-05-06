package stack

import (
    "testing"
    "fmt"
)

func assert( pass bool, msg string, t *testing.T ) {
    if pass { return }
    t.Fatal( msg )
}

func assertPanicOnPopEmpty( s *Stack, t *testing.T ) {
    defer func() {
        if err := recover(); err == nil { t.Fatal( "no error from pop" ) }
    }()
    s.Pop()
}

func TestStackBasic( t *testing.T ) {
    s := NewStack()
    assert( s.IsEmpty(), "stack not empty", t )
    assert( s.Peek() == nil, "empty stack peek not nil", t )
    s.Push( 1 )
    assert( ! s.IsEmpty(), "stack reports empty", t )
    assert( s.Peek().( int ) == 1, "stack top is not 1", t )
    assert( s.Pop().( int ) == 1, "result of pop is not 1", t )
    assert( s.IsEmpty(), "stack not empty", t )
    s.Push( 1 )
    s.Push( 2 )
    assert( ! s.IsEmpty(), "stack reports empty", t )
    assert( s.Peek().( int ) == 2, "stack top is not 2", t )
    assert( s.Pop().( int ) == 2, "result of pop is not 2", t )
    assert( ! s.IsEmpty(), "stack reports empty", t )
    assert( s.Pop().( int ) == 1, "result of pop is not 1", t )
    assert( s.IsEmpty(), "stack not empty", t )
    assertPanicOnPopEmpty( s, t )
}

func TestStackVisit( t *testing.T ) {
    s := NewStack()
    s.Push( 0 )
    s.Push( 1 )
    s.Push( 2 )
    type checkCtx struct { acc, incr int }
    makeCheck := func( ctx *checkCtx ) func( interface{} ) {
        return func( v interface{} ) {
            i := v.( int ) 
            msg := fmt.Sprintf( "%d != %d", ctx.acc, i )
            assert( ctx.acc == i, msg, t )
            ctx.acc += ctx.incr
        }
    }
    topChk := &checkCtx{ acc: 2, incr: -1 }
    s.VisitTop( makeCheck( topChk ) )
    topMsg := fmt.Sprintf( "top visit acc is %d", topChk.acc )
    assert( topChk.acc == -1, topMsg, t )
}
