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

func appendNode( elt interface{}, f Formatter, apnd AppendFunc, isRoot bool ) {
    switch v := elt.( type ) {
    case *dictNode: {
        if ! isRoot { f.AppendSeparator( apnd ) }
        f.AppendDictKey( v.elt, apnd )
    }
    case *ListNode: {
        apnd( "[ " )
        apnd( strconv.Itoa( int( v.indx ) ) )
        apnd( " ]" )
    }
    default: panic( &PathElementError{ elt } )
    }
}

func Format( p PathNode, f Formatter ) string {
    res := &bytes.Buffer{}
    apnd := func( s string ) { res.WriteString( s ) }
    elts := ascentOrderFor( p )
    for i := len( elts ); i > 0; i-- {
        appendNode( elts[ i - 1 ], f, apnd, i == len( elts ) )
    }
    return res.String()
}

func RootedAt( root interface{} ) PathNode { return descend( nil, root ) }

func RootedAtList() *ListNode { return startList( nil ) }
