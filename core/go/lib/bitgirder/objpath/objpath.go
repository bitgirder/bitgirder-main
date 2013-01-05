package objpath

import (
    "bytes"
    "fmt"
//    "log"
    "strconv"
)

type PathElementError struct { elt interface{} }

func ( e *PathElementError ) Error() string {
    return fmt.Sprintf( "Unrecognized path element (%T): %v", e.elt, e.elt )
}

type PathNode interface {
    Descend( elt interface{} ) PathNode
    StartList() *ListNode
    parentOf() PathNode // nil when root
}

type dictNode struct {
    parent PathNode
    elt interface{}
}

type ListNode struct {
    parent PathNode
    indx int
}

func descend( parent PathNode, elt interface{} ) PathNode {
    return &dictNode{ parent, elt }
}

func startList( parent PathNode ) *ListNode { return &ListNode{ parent, 0 } }

func ( n *dictNode ) parentOf() PathNode { return n.parent }

func ( n *dictNode ) Descend( elt interface{} ) PathNode {
    return descend( n, elt ) 
}

func ( n *dictNode ) StartList() *ListNode { return startList( n ) }

func ( l *ListNode ) parentOf() PathNode { return l.parent }

func ( l *ListNode ) Descend( elt interface{} ) PathNode {
    return descend( l, elt )
}

func ( l *ListNode ) StartList() *ListNode { return startList( l ) }

func ( l *ListNode ) Next() *ListNode { 
    return &ListNode{ l.parent, l.indx + 1 } 
}

func ( l *ListNode ) SetIndex( indx int ) *ListNode {
    l.indx = indx
    return l
}

type AppendFunc func( s string )

type Formatter interface {
    AppendSeparator( apnd AppendFunc )
    AppendDictKey( elt interface{}, apnd AppendFunc )
}

type dotFormatter func( elt interface{}, apnd AppendFunc )

func ( f dotFormatter ) AppendSeparator( apnd AppendFunc ) { apnd( "." ) }

func ( f dotFormatter ) AppendDictKey( elt interface{}, apnd AppendFunc ) {
    f( elt, apnd )
}

func DotFormatter( f func( elt interface{}, apnd AppendFunc ) ) Formatter {
    return dotFormatter( f )
}

var StringDotFormatter Formatter

func init() {
    StringDotFormatter = 
        DotFormatter( func( elt interface{}, apnd AppendFunc ) {
            switch v := elt.( type ) {
            case string: apnd( v )
            case fmt.Stringer: apnd( v.String() )
            default: panic( fmt.Errorf( "Can't convert to string: %T", elt ) )
            }
        })
}

func ascentOrderFor( n PathNode ) []interface{} {
    res := make( []interface{}, 0, 5 )
    for elt := n; elt != nil; elt = elt.parentOf() { res = append( res, elt ) }
    return res
}

type Visitor interface {
    Descend( elt interface{} ) error
    List( idx int ) error
}

func Visit( p PathNode, v Visitor ) error {
    elts := ascentOrderFor( p )
    for i := len( elts ); i > 0; i-- {
        switch n := elts[ i - 1 ].( type ) {
        case *dictNode:
            if err := v.Descend( n.elt ); err != nil { return err }
        case *ListNode:
            if err := v.List( n.indx ); err != nil { return err }
        default: panic( libErrorf( "Unhandled node type: %T", n ) )
        }
    }
    return nil
}

type formatVisitor struct {
    sawRoot bool
    f Formatter
    apnd AppendFunc
}

func ( v *formatVisitor ) Descend( elt interface{} ) error {
    if v.sawRoot { v.f.AppendSeparator( v.apnd ) }
    v.sawRoot = true // mayb was already
    v.f.AppendDictKey( elt, v.apnd )
    return nil
}

func ( v *formatVisitor ) List( idx int ) error {
    v.sawRoot = true
    v.apnd( "[ " )
    v.apnd( strconv.Itoa( idx ) )
    v.apnd( " ]" )
    return nil
}

func Format( p PathNode, f Formatter ) string {
    res := &bytes.Buffer{}
    v := &formatVisitor{
        f: f,
        apnd: func( s string ) { res.WriteString( s ) },
    }
    Visit( p, v )
    return res.String()
}

func RootedAt( root interface{} ) PathNode { return descend( nil, root ) }

func RootedAtList() *ListNode { return startList( nil ) }

func Descend( p PathNode, elt interface{} ) PathNode {
    if p == nil { return RootedAt( elt ) }
    return p.Descend( elt )
}

func StartList( p PathNode ) *ListNode {
    if p == nil { return RootedAtList() }
    return p.StartList()
}
