package testing

import (
//    "log"
    "fmt"
    "bitgirder/assert"
    "bitgirder/objpath"
    mg "mingle"
)

type Failer interface {
    Fatal( args ...interface{} )
}

var idAct *mg.Identifier
func init() { idAct = mg.MustIdentifier( "act" ) }

type idPath objpath.PathNode

type asserter struct {

    f Failer
    path idPath

    *assert.Asserter
}

func newAsserter( f Failer, path idPath ) *asserter {
    res := &asserter{ 
        f: f, 
        path: path,
        Asserter: assert.AsAsserter( 
            func( args ...interface{} ) {
                args2 := make( []interface{}, 1, 1 + len( args ) )
                args2[ 0 ] = fmt.Sprintf( "[%s] ", mg.FormatIdPath( path ) )
                f.Fatal( append( args2, args... )... )
            },
        ),
    }

    return res
}

func ( a *asserter ) descend( fld *mg.Identifier ) *asserter {
    return newAsserter( a.f, a.path.Descend( fld ) )
}

func ( a *asserter ) startList() *asserter {
    return newAsserter( a.f, a.path.StartList() )
}

func ( a *asserter ) next() *asserter {
    return newAsserter( a.f, a.path.( *objpath.ListNode ).Next() )
}

func ( a *asserter ) equalTypes( t1, t2 mg.TypeReference ) {
    a.Truef( t1.Equals( t2 ), "%s != %s", t1, t2 )
}

func ( a *asserter ) failBadType( expctNm string, act mg.Value ) {
    a.Fatalf( "Expected %s but got %s (%T)", 
        expctNm, mg.QuoteValue( act ), act )
}

func ( a *asserter ) equalDefault( expct, act mg.Value ) {
    var err error
    expctTyp := mg.TypeOf( expct )
    if act, err = mg.CastValue( act, expctTyp, a.path ); err == nil {
        if comp, ok := expct.( mg.Comparer ); ok {
            a.Truef( comp.Compare( act ) == 0, "got %v, want %v", act, expct )
        } else { a.Equal( expct, act ) }
    } else { a.Fatal( err ) }
}

func ( a *asserter ) equalSymbolMaps( m1 *mg.SymbolMap, act mg.Value ) {
    if m2, ok := act.( *mg.SymbolMap ); ok {
        a.Equal( m1.Len(), m2.Len() )
        m1.EachPair( func( fld *mg.Identifier, val1 mg.Value ) {
            if val2 := m2.Get( fld ); val2 == nil {
                a.Fatalf( "No value for field %q", fld )
            } else { a.descend( fld ).equal( val1, val2 ) }
        })
    } else { a.failBadType( "symbol map", act ) }
}

func ( a *asserter ) equalStructs( s1 *mg.Struct, act mg.Value ) {
    if s2, ok := act.( *mg.Struct ); ok {
        a.Equal( s1.Type, s2.Type )
        a.equalSymbolMaps( s1.Fields, s2.Fields )
    } else { a.failBadType( "struct", act ) }
}

func ( a *asserter ) equalEnums( e1 *mg.Enum, act mg.Value ) {
    switch v := act.( type ) {
    case *mg.Enum:
        a.Equal( e1.Type, v.Type )
        a.Equal( e1.Value, v.Value )
    case mg.String: a.Equal( e1.Value, mg.MustIdentifier( string( v ) ) )
    default: a.failBadType( "enum", act )
    }
}

func ( a *asserter ) equalLists( l1 *mg.List, act mg.Value ) {
    if l2, ok := act.( *mg.List ); ok {
        a.Equal( l1.Len(), l2.Len() )
        la := a.startList()
        l2Vals := l2.Values()
        for i, l1Val := range l1.Values() {
            la.equal( l1Val, l2Vals[ i ] )
            la = la.next()
        }
    } else { a.failBadType( "list", act ) }
}

func ( a *asserter ) equal( expct, act mg.Value ) {
    switch v := expct.( type ) {
    case *mg.Struct: a.equalStructs( v, act )
    case *mg.Enum: a.equalEnums( v, act )
    case *mg.SymbolMap: a.equalSymbolMaps( v, act )
    case *mg.List: a.equalLists( v, act )
    default: a.equalDefault( expct, act )
    }
}

func LossyEqual( expct, act mg.Value, f Failer ) {
    path := objpath.RootedAt( idAct )
    newAsserter( f, path ).equal( expct, act )
}
