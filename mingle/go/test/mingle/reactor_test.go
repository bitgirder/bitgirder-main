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

type castInterfaceImpl struct { typers *QnameMap }

type castInterfaceTyper struct { *IdentifierMap }

func ( t castInterfaceTyper ) FieldTypeOf( 
    fld *Identifier, pg PathGetter ) ( TypeReference, error ) {
    if t.HasKey( fld ) { return t.Get( fld ).( TypeReference ), nil }
    errPath := pg.GetPath().Parent()
    return nil, newValueCastErrorf( errPath, "unrecognized field: %s", fld )
}

func newCastInterfaceImpl() *castInterfaceImpl {
    res := &castInterfaceImpl{ typers: NewQnameMap() }
    m1 := castInterfaceTyper{ NewIdentifierMap() }
    m1.Put( MustIdentifier( "f1" ), TypeInt32 )
    qn := MustQualifiedTypeName
    res.typers.Put( qn( "ns1@v1/T1" ), m1 )
    m2 := castInterfaceTyper{ NewIdentifierMap() }
    m2.Put( MustIdentifier( "f1" ), TypeInt32 )
    m2.Put( MustIdentifier( "f2" ), MustTypeReference( "ns1@v1/T1" ) )
    res.typers.Put( qn( "ns1@v1/T2" ), m2 )
    return res
}

func ( ci *castInterfaceImpl ) FieldTyperFor( 
    qn *QualifiedTypeName, pg PathGetter ) ( FieldTyper, error ) {
    if ci.typers.HasKey( qn ) { return ci.typers.Get( qn ).( FieldTyper ), nil }
    if qn.ExternalForm() == "ns1@v1/FailType" {
        return nil, newValueCastErrorf( pg.GetPath(), "test-message-fail-type" )
    }
    return nil, nil
}

func ( ci *castInterfaceImpl ) InferStructFor( at *AtomicTypeReference ) bool {
    return ci.typers.HasKey( at.Name.( *QualifiedTypeName ) )
}

func ( c *reactorTestCall ) createCastReactor( 
    ct *CastReactorTest ) *CastReactor {
    switch ct.Profile {
    case "": return NewDefaultCastReactor( ct.Type, ct.Path )
    case "interface-impl": 
        return NewCastReactor( ct.Type, newCastInterfaceImpl(), ct.Path )
    }
    panic( libErrorf( "Unhandled profile: %s", ct.Profile ) )
}

func ( c *reactorTestCall ) callCast( ct *CastReactorTest ) {
    rct := NewValueBuilder()
    pip := InitReactorPipeline( c.createCastReactor( ct ), rct )
    if err := VisitValue( ct.In, pip ); err == nil { 
        c.Equal( ct.Expect, rct.GetValue() )
    } else { ( castErrorAssert{ ct, err, c.PathAsserter } ).call() }
}

func ( c *reactorTestCall ) call() {
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
