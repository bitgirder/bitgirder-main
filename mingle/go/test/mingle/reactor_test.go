package mingle

import (
    "testing"
    "bitgirder/assert"
)

type reactorTestCall struct {
    *assert.PathAsserter
    rt ReactorTest
}

func ( c *reactorTestCall ) feedStructureEvents( 
    evs []ReactorEvent, tt ReactorTopType ) ( *StructuralReactor, error ) {
    rct := NewStructuralReactor( tt )
    pip := InitReactorPipeline( rct )
    for _, ev := range evs { 
        if err := pip.ProcessEvent( ev ); err != nil { return nil, err }
    }
    return rct, nil
}

func ( c *reactorTestCall ) callStructuralError(
    ss *StructuralReactorErrorTest ) {
    c.Logf( "Expecting error: %s", ss.Error )
    if _, err := c.feedStructureEvents( ss.Events, ss.TopType ); err == nil {
        c.Fatalf( "Expected error (%T): %s", ss.Error, ss.Error ) 
    } else { c.Equal( ss.Error, err ) }
}

func ( c *reactorTestCall ) callStructuralPath(
    pt *StructuralReactorPathTest ) {
    sr, err := c.feedStructureEvents( pt.Events, ReactorTopTypeValue )
    if err != nil { c.Fatal( err ) }
    var act idPath
    if pt.StartPath == nil { 
        act = sr.GetPath() 
    } else { act = sr.AppendPath( pt.StartPath ) }
    c.Equal( pt.Path, act )
}

func ( c *reactorTestCall ) callValueBuild( vb ValueBuildTest ) {
    rct := NewValueBuilder()
    pip := InitReactorPipeline( rct )
    if err := VisitValue( vb.Val, pip ); err == nil {
        c.Equal( vb.Val, rct.GetValue() )
    } else { c.Fatal( err ) }
}

type castErrorAssert struct {
    ct *CastReactorTest
    err error
    *assert.PathAsserter
}

// Returns a path asserter that can be used further
func ( cea castErrorAssert ) assertValueError( 
    expct, act ValueError ) *assert.PathAsserter {
    a := cea.Descend( "Err" )
    a.Descend( "Error()" ).Equal( expct.Error(), act.Error() )
    a.Descend( "Message()" ).Equal( expct.Message(), act.Message() )
    a.Descend( "Location()" ).Equal( expct.Location(), act.Location() )
    return a
}

func ( cea castErrorAssert ) assertTcError() {
    if act, ok := cea.err.( *TypeCastError ); ok {
        expct := cea.ct.Err.( *TypeCastError )
        a := cea.assertValueError( expct, act )
        a.Descend( "expcted" ).Equal( expct.Expected, act.Expected )
        a.Descend( "actual" ).Equal( expct.Actual, act.Actual )
    } else { cea.Fatal( cea.err ) }
}

func ( cea castErrorAssert ) assertVcError() {
    if act, ok := cea.err.( *ValueCastError ); ok {
        cea.assertValueError( cea.ct.Err.( *ValueCastError ), act )
    } else { cea.Fatal( cea.err ) }
}

func ( cea castErrorAssert ) call() {
    switch cea.ct.Err.( type ) {
    case nil: cea.Fatal( cea.err )
    case *TypeCastError: cea.assertTcError()
    case *ValueCastError: cea.assertVcError()
    default: cea.Fatalf( "Unhandled Err type: %T", cea.ct.Err )
    }
}

func ( c *reactorTestCall ) callCast( ct *CastReactorTest ) {
    c.Logf( "Casting %s to %s", QuoteValue( ct.In ), ct.Type )
    rct := NewValueBuilder()
    pip := InitReactorPipeline( NewCastReactor( ct.Type, ct.Path ), rct )
    if err := VisitValue( ct.In, pip ); err == nil { 
        c.Equal( ct.Expect, rct.GetValue() )
    } else { ( castErrorAssert{ ct, err, c.PathAsserter } ).call() }
}

func ( c *reactorTestCall ) call() {
    c.Logf( "Calling %T", c.rt )
    switch s := c.rt.( type ) {
    case *StructuralReactorErrorTest: c.callStructuralError( s )
    case *StructuralReactorPathTest: c.callStructuralPath( s )
    case ValueBuildTest: c.callValueBuild( s )
    case *CastReactorTest: c.callCast( s )
    default: panic( libErrorf( "Unhandled test source: %T", c.rt ) )
    }
}

func TestReactors( t *testing.T ) {
    a := assert.NewListPathAsserter( t )
    for _, rt := range StdReactorTests {
        ( &reactorTestCall{ PathAsserter: a, rt: rt } ).call()
        a = a.Next()
    }
}
