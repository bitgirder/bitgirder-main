package mingle

import (
    "testing"
    "bitgirder/assert"
)

type reactorTestCall struct {
    *assert.PathAsserter
    rt *ReactorTest
}

func ( c *reactorTestCall ) processStructural( 
    ss *StructuralReactorTestSource ) error {
    rct := NewStructuralReactor( ss.TopType )
    pip := InitReactorPipeline( rct )
    for _, ev := range ss.Events { 
        if err := pip.ProcessEvent( ev ); err != nil { return err }
    }
    return nil
}

func ( c *reactorTestCall ) processValueBuild( vb ValueBuildSource ) error {
    rct := NewValueBuilder()
    pip := InitReactorPipeline( rct )
    if err := VisitValue( vb.Val, pip ); err == nil {
        c.Equal( vb.Val, rct.GetValue() )
    } else { c.Fatal( err ) }
    return nil
}

func ( c *reactorTestCall ) processSource() error {
    switch s := c.rt.Source.( type ) {
    case *StructuralReactorTestSource: return c.processStructural( s )
    case ValueBuildSource: return c.processValueBuild( s )
    }
    panic( libErrorf( "Unhandled test source: %T", c.rt.Source ) )
}

func ( c *reactorTestCall ) assertErrors( expct, act error ) {
    c.Equal( expct, act )
}

func ( c *reactorTestCall ) call() {
    rtErr := c.rt.Error
    if err := c.processSource(); err == nil {
        if rtErr != nil { c.Fatalf( "Expected error (%T): %s", rtErr, rtErr ) }
    } else { 
        if rtErr == nil { c.Fatal( err ) }
        c.assertErrors( rtErr, err )
    }
}

func TestReactors( t *testing.T ) {
    a := assert.NewListPathAsserter( t )
    for _, rt := range StdReactorTests {
        ( &reactorTestCall{ PathAsserter: a, rt: rt } ).call()
        a = a.Next()
    }
}
