package main

import (
    mg "mingle"
    "mingle/testgen"
    "fmt"
    "log"
)

const (

    fileVersion = int32( 1 )

    ttEnd = int8( 0 )
    ttCastReactorTest = int8( 1 )
    ttValueBuildTest = int8( 2 )
    ttStructuralReactorErrorTest = int8( 3 )
    ttEventPathTest = int8( 4 )
    ttFieldOrderReactorTest = int8( 5 )
    ttFieldOrderMissingFieldsTest = int8( 6 )
    ttFieldOrderPathTest = int8( 7 )
    ttServiceRequestReactorTest = int8( 8 )
    ttServiceResponseReactorTest = int8( 9 )

    tcValueEvent = int8( 1 )
    tcStructStartEvent = int8( 2 )
    tcFieldStartEvent = int8( 3 )
    tcListStartEvent = int8( 4 )
    tcMapStartEvent = int8( 5 )
    tcEndEvent = int8( 6 )

    tcErrNil = int8( 0 )
    tcErrReactorError = int8( 1 )
    tcErrMissingFieldsError = int8( 2 )
)

func writeTestType( w *mg.BinWriter, tt int8 ) error {
    return w.WriteInt8( tt )
}

func writeNilError( w *mg.BinWriter ) error { return w.WriteInt8( tcErrNil ) }

func writeReactorError( w *mg.BinWriter, re *mg.ReactorError ) error {
    if err := w.WriteInt8( tcErrReactorError ); err != nil { return err }
    return w.WriteUtf8( re.Error() )
}

func writeMissingFieldsError( 
    w *mg.BinWriter, 
    me *mg.MissingFieldsError,
) error {
    if me == nil { return writeNilError( w ) }
    if err := w.WriteInt8( tcErrMissingFieldsError ); err != nil { return err }
    if err := w.WriteUtf8( me.Error() ); err != nil { return err }
    return w.WriteIdPath( me.Location() )
}

func writeError( w *mg.BinWriter, e error ) error {
    switch v := e.( type ) {
    case nil: return writeNilError( w )
    case *mg.ReactorError: return writeReactorError( w, v )
    case *mg.MissingFieldsError: return writeMissingFieldsError( w, v )
    }
    panic( fmt.Errorf( "unhandled error: %T", e ) )
}

func writeCastReactorTest( w *mg.BinWriter, t *mg.CastReactorTest ) error {
    if err := writeTestType( w, ttCastReactorTest ); err != nil { return err }
    if err := w.WriteValue( t.In ); err != nil { return err }
    if err := w.WriteValue( t.Expect ); err != nil { return err }
    if err := w.WriteIdPath( t.Path ); err != nil { return err }
    if err := w.WriteTypeReference( t.Type ); err != nil { return err }
    if err := w.WriteUtf8( t.Profile ); err != nil { return err }
    return nil
}

func writeValueBuildTest( w *mg.BinWriter, t mg.ValueBuildTest ) error {
    if err := writeTestType( w, ttValueBuildTest ); err != nil { return err }
    return w.WriteValue( t.Val )
}

func writeReactorEvent( w *mg.BinWriter, ev mg.ReactorEvent ) error {
    switch v := ev.( type ) {
    case mg.ValueEvent: 
        if err := w.WriteInt8( tcValueEvent ); err != nil { return err }
        return w.WriteValue( v.Val )
    case mg.StructStartEvent:
        if err := w.WriteInt8( tcStructStartEvent ); err != nil { return err }
        return w.WriteQualifiedTypeName( v.Type )
    case mg.FieldStartEvent:
        if err := w.WriteInt8( tcFieldStartEvent ); err != nil { return err }
        return w.WriteIdentifier( v.Field )
    case mg.ListStartEvent: return w.WriteInt8( tcListStartEvent )
    case mg.MapStartEvent: return w.WriteInt8( tcMapStartEvent )
    case mg.EndEvent: return w.WriteInt8( tcEndEvent )
    }
    panic( fmt.Errorf( "unhandled event type: %T", ev ) )
}

func writeReactorEvents( w *mg.BinWriter, evs []mg.ReactorEvent ) error {
    if err := w.WriteInt32( int32( len( evs ) ) ); err != nil { return err }
    for _, ev := range evs {
        if err := writeReactorEvent( w, ev ); err != nil { return err }
    }
    return nil
}

func writeStructuralReactorErrorTest( 
    w *mg.BinWriter, 
    t *mg.StructuralReactorErrorTest,
) error {
    if err := writeTestType( w, ttStructuralReactorErrorTest ); err != nil {
        return err
    }
    if err := writeReactorEvents( w, t.Events ); err != nil { return err }
    if err := writeError( w, t.Error ); err != nil { return err }
    return w.WriteInt8( int8( t.TopType ) )
}

func writeEventExpectation( w *mg.BinWriter, ee mg.EventExpectation ) error {
    if err := writeReactorEvent( w, ee.Event ); err != nil { return err }
    return w.WriteIdPath( ee.Path )
}

func writeEventExpectations(
    w *mg.BinWriter,
    evs []mg.EventExpectation,
) error {
    if err := w.WriteInt32( int32( len( evs ) ) ); err != nil { return err }
    for _, ev := range evs {
        if err := writeEventExpectation( w, ev ); err != nil { return err }
    }
    return nil
}

func writeEventPathTest( w *mg.BinWriter, t *mg.EventPathTest ) error {
    if err := writeTestType( w, ttEventPathTest ); err != nil { return err }
    if err := writeEventExpectations( w, t.Events ); err != nil { return err }
    if err := w.WriteIdPath( t.StartPath ); err != nil { return err }
    return w.WriteIdPath( t.FinalPath )
}

func writeFieldOrderReactorTest(
    w *mg.BinWriter,
    t *mg.FieldOrderReactorTest,
) error {
    if err := writeTestType( w, ttFieldOrderReactorTest ); err != nil { 
        return err
    }
    if err := writeReactorEvents( w, t.Source ); err != nil { return err }
    if err := w.WriteValue( t.Expect ); err != nil { return err }
    return w.WriteIdentifiers( t.Order ); 
}

func writeFieldOrderPathTest( 
    w *mg.BinWriter, 
    t *mg.FieldOrderPathTest,
) error {
    if err := writeTestType( w, ttFieldOrderPathTest ); err != nil { 
        return err
    }
    if err := writeReactorEvents( w, t.Source ); err != nil { return err }
    if err := writeEventExpectations( w, t.Expect ); err != nil { return err }
    return w.WriteIdentifiers( t.Order )
}

func writeFieldOrderSpecification(
    w *mg.BinWriter,
    spec mg.FieldOrderSpecification,
) error {
    if err := w.WriteIdentifier( spec.Field ); err != nil { return err }
    return w.WriteBool( spec.Required )
}

func writeFieldOrder( w *mg.BinWriter, ord mg.FieldOrder ) error {
    if err := w.WriteInt32( int32( len( ord ) ) ); err != nil { return err }
    for _, spec := range ord {
        if err := writeFieldOrderSpecification( w, spec ); err != nil {
            return err
        }
    }
    return nil
}

func writeFieldOrderMissingFieldsTest(
    w *mg.BinWriter,
    t *mg.FieldOrderMissingFieldsTest,
) error {
    if err := writeTestType( w, ttFieldOrderMissingFieldsTest ); err != nil {
        return err
    }
    if err := writeFieldOrder( w, t.Order ); err != nil { return err }
    if err := writeReactorEvents( w, t.Source ); err != nil { return err }
    if err := w.WriteValue( t.Expect ); err != nil { return err }
    return writeError( w, t.Error )
}

func writeServiceRequestReactorTest(
    w *mg.BinWriter,
    t *mg.ServiceRequestReactorTest,
) error {
    v := mg.MustStruct( "mingle:test:core@v1/ServiceRequestReactorTest",
    )
    return w.WriteValue( v ) 
}

func writeTest( w *mg.BinWriter, t interface{} ) error {
    switch v := t.( type ) {
    case *mg.CastReactorTest: return writeCastReactorTest( w, v )
    case mg.ValueBuildTest: return writeValueBuildTest( w, v )
    case *mg.StructuralReactorErrorTest: 
        return writeStructuralReactorErrorTest( w, v )
    case *mg.EventPathTest: return writeEventPathTest( w, v )
    case *mg.FieldOrderReactorTest: return writeFieldOrderReactorTest( w, v )
    case *mg.FieldOrderPathTest: return writeFieldOrderPathTest( w, v )
    case *mg.FieldOrderMissingFieldsTest:
        return writeFieldOrderMissingFieldsTest( w, v )
    case *mg.ServiceRequestReactorTest:
        return writeServiceRequestReactorTest( w, v )
    }
    log.Printf( "skipping unhandled test: %T", t )
    return nil
}

func main() {
    testgen.WriteOutFile( func( w *mg.BinWriter ) error {
        if err := w.WriteInt32( fileVersion ); err != nil { return err }
        for _, t := range mg.StdReactorTests {
            if err := writeTest( w, t ); err != nil { return err }
        }
        return writeTestType( w, ttEnd )
    })
}
