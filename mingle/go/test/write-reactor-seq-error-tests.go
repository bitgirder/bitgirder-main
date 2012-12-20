package main

import (
    mg "mingle"
    "mingle/testgen"
)

const (

    fileVersion = int32( 1 )

    fileEnd = int8( 0 )
    seqTest = int8( 1 )
)

func writeTest( w *mg.BinWriter, t *mg.ReactorSeqErrorTest ) error {
    if err := w.WriteInt8( seqTest ); err != nil { return err }
    if err := w.WriteInt32( int32( len( t.Seq ) ) ); err != nil { return err }
    for _, s := range t.Seq {
        if err := w.WriteUtf8( s ); err != nil { return err }
    }
    if err := w.WriteUtf8( t.ErrMsg ); err != nil { return err }
    if err := w.WriteUtf8( t.TopType.String() ); err != nil { return err }
    return nil
}

func main() {
    testgen.WriteOutFile( func( w *mg.BinWriter ) error {
        if err := w.WriteInt32( fileVersion ); err != nil { return err }
        for _, t := range mg.StdReactorSeqErrorTests {
            if err := writeTest( w, t ); err != nil { return err }
        }
        return w.WriteInt8( fileEnd )
    })
}
