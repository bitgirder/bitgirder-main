package mingle

import (
    "testing"
    "bitgirder/objpath"
    "bitgirder/assert"
)

func TestCastValueErrorFormatting( t *testing.T ) {
    path := objpath.RootedAt( id( "f1" ) )
    err := newValueCastErrorf( path, "Blah %s", "X" )
    assert.Equal( "f1: Blah X", err.Error() )
}

type cvtRunner struct {
    cvt *CastValueTest
    *assert.PathAsserter
}

// Returns a path asserter that can be used further
func ( r *cvtRunner ) assertValueError( 
    expct, act ValueError ) *assert.PathAsserter {
    a := r.Descend( "Err" )
    a.Descend( "Error()" ).Equal( expct.Error(), act.Error() )
    a.Descend( "Message()" ).Equal( expct.Message(), act.Message() )
    a.Descend( "Location()" ).Equal( expct.Location(), act.Location() )
    return a
}

func ( r *cvtRunner ) assertTcError( err error ) {
    if act, ok := err.( *TypeCastError ); ok {
        expct := r.cvt.Err.( *TypeCastError )
        a := r.assertValueError( expct, act )
        a.Descend( "expcted" ).Equal( expct.Expected, act.Expected )
        a.Descend( "actual" ).Equal( expct.Actual, act.Actual )
    } else { r.Fatal( err ) }
}

func ( r *cvtRunner ) assertVcError( err error ) {
    if act, ok := err.( *ValueCastError ); ok {
        r.assertValueError( r.cvt.Err.( *ValueCastError ), act )
    } else { r.Fatal( err ) }
}

func ( r *cvtRunner ) assertError( err error ) {
    switch r.cvt.Err.( type ) {
    case nil: r.Fatal( err )
    case *TypeCastError: r.assertTcError( err )
    case *ValueCastError: r.assertVcError( err )
    default: r.Fatalf( "Unhandled Err type: %T", r.cvt.Err )
    }
}

func ( r *cvtRunner ) call() {
    if act, err := CastValue( r.cvt.In, r.cvt.Type, r.cvt.Path ); err == nil {
        if r.cvt.Err != nil { r.Fatal( "Expected error" ) }
        r.Equal( r.cvt.Expect, act )
    } else { r.assertError( err ) }
}

func TestCastValue( t *testing.T ) {
    a := assert.NewPathAsserter( t ).Descend( "cvTests" ).StartList()
    for _, cvt := range GetCastValueTests() {
        ( &cvtRunner{ cvt: cvt, PathAsserter: a } ).call()
        a = a.Next()
    }
}
