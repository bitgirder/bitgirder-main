package stack

import (
    "container/list"
)

type Stack struct { l *list.List }

func NewStack() *Stack { return &Stack{ l: &list.List{} } }

func ( s *Stack ) IsEmpty() bool { return s.l.Len() == 0 }

func ( s *Stack ) Peek() interface{} {
    if s.IsEmpty() { return nil }
    return s.l.Front().Value
}

func ( s *Stack ) Pop() interface {} {
    if s.IsEmpty() { panic( libErrorf( "call to Pop() of empty stack" ) ) }
    return s.l.Remove( s.l.Front() )
}

func ( s *Stack ) Push( val interface{} ) { s.l.PushFront( val ) }
