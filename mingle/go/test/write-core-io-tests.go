package main

import (
    "mingle/testgen"
    "bytes"
    "fmt"
    mg "mingle"
)

type typeCode int8
const (
    tcEnd = typeCode( iota )
    tcInvalidDataTest
    tcRoundtripTest
    tcSequenceRoundtripTest
)

func writeTypeCode( tc typeCode, w *mg.BinWriter ) error {
    return w.WriteInt8( int8( tc ) )
}

func writeValue( val interface{}, bb *bytes.Buffer ) error {
    w := mg.NewWriter( bb )
    return mg.WriteBinIoTestValue( val, w )
}

func writeInvalidDataTest( t *mg.BinIoInvalidDataTest, w *mg.BinWriter ) error {
    if err := writeTypeCode( tcInvalidDataTest, w ); err != nil { return err }
    if err := w.WriteUtf8( t.Name ); err != nil { return err }
    if err := w.WriteUtf8( t.ErrMsg ); err != nil { return err }
    if err := w.WriteBuffer32( t.Input ); err != nil { return err }
    return nil
}

func writeRoundtripTest( t *mg.BinIoRoundtripTest, w *mg.BinWriter ) error {
    if err := writeTypeCode( tcRoundtripTest, w ); err != nil { return err }
    if err := w.WriteUtf8( t.Name ); err != nil { return err }
    bb := &bytes.Buffer{}
    if err := writeValue( t.Val, bb ); err != nil { return err }
    if err := w.WriteBuffer32( bb.Bytes() ); err != nil { return err }
    return nil
}

func writeSequenceTest(
    t *mg.BinIoSequenceRoundtripTest,
    w *mg.BinWriter,
) error {
    if err := writeTypeCode( tcSequenceRoundtripTest, w ); err != nil { 
        return err 
    }
    if err := w.WriteUtf8( t.Name ); err != nil { return err }
    bb := &bytes.Buffer{}
    for _, val := range t.Seq {
        if err := writeValue( val, bb ); err != nil { return err }
    }
    if err := w.WriteBuffer32( bb.Bytes() ); err != nil { return err }
    return nil
}

func writeTest( test interface{}, w *mg.BinWriter ) error { 
    switch v := test.( type ) {
    case *mg.BinIoInvalidDataTest: return writeInvalidDataTest( v, w )
    case *mg.BinIoRoundtripTest: return writeRoundtripTest( v, w )
    case *mg.BinIoSequenceRoundtripTest: return writeSequenceTest( v, w )
    }
    return fmt.Errorf( "unhandled test type: %T", test )
}

func writeTests( w *mg.BinWriter ) error {
    for _, test := range mg.CreateCoreIoTests() {
        if err := writeTest( test, w ); err != nil { return err }
    }
    return writeTypeCode( tcEnd, w )
}

func main() { testgen.WriteOutFile( writeTests ) }
