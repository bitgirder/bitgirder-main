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
    if _, err := c.feedStructureEvents( ss.Events, ss.TopType ); err == nil {
        c.Fatalf( "Expected error (%T): %s", ss.Error, ss.Error ) 
    } else { c.Equal( ss.Error, err ) }
}

func ( c *reactorTestCall ) callStructuralPath(
    pt *StructuralReactorPathTest ) {
    sr, err := c.feedStructureEvents( pt.Events, ReactorTopTypeValue )
    if err != nil { c.Fatal( err ) }
    c.Equal( pt.Path, sr.GetPath() )
}

func ( c *reactorTestCall ) callValueBuild( vb ValueBuildTest ) {
    rct := NewValueBuilder()
    pip := InitReactorPipeline( rct )
    if err := VisitValue( vb.Val, pip ); err == nil {
        c.Equal( vb.Val, rct.GetValue() )
    } else { c.Fatal( err ) }
}

func ( c *reactorTestCall ) callCast( ct *CastReactorTest ) {
    rct := NewValueBuilder()
    pip := InitReactorPipeline( NewCastReactor( ct.Type, ct.Path ), rct )
    if err := VisitValue( ct.In, pip ); err != nil { c.Fatal( err ) }
    c.Equal( ct.Expect, rct.GetValue() )
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
