package main

import (
    mg "mingle"
    "mingle/testgen"
    "bitgirder/objpath"
    "log"
    "fmt"
    "flag"
    bgio "bitgirder/io"
)

const (
    
    fileVersion = uint32( 0x01 )

    tcNil = uint8( 0x00 )
    tcTest = uint8( 0x01 )
    tcIn = uint8( 0x02 )
    tcExpect = uint8( 0x03 )
    tcType = uint8( 0x04 )
    tcPath = uint8( 0x05 )
    tcErr = uint8( 0x06 )
    tcVcErr = uint8( 0x07 )
    tcMsg = uint8( 0x08 )
    tcTcErr = uint8( 0x09 )
    tcExpected = uint8( 0x0a )
    tcActual = uint8( 0x0b )
)

type writer struct {
    w *bgio.BinWriter
    mgw *mg.BinWriter
}

func ( w writer ) writeTc( tc uint8 ) error { return w.w.WriteUint8( tc ) }

func ( w writer ) writeEnd() error { return w.writeTc( tcNil ) } 

func ( w writer ) writeValue( val mg.Value ) error {
    return w.mgw.WriteValue( val )
}

func ( w writer ) writeType( t mg.TypeReference ) error {
    return w.mgw.WriteTypeReference( t )
}

func ( w writer ) writePath( p objpath.PathNode ) ( err error ) {
    return w.mgw.WriteIdPath( p )
}

func ( w writer ) startValueError( ve mg.ValueError, tc uint8 ) ( err error ) {
    if err = w.writeTc( tc ); err != nil { return }
    if err = w.writeTc( tcMsg ); err != nil { return }
    if err = w.w.WriteUtf8( ve.Message() ); err != nil { return }
    if err = w.writeTc( tcPath ); err != nil { return }
    if err = w.writePath( ve.Location() ); err != nil { return }
    return
}

func ( w writer ) writeValueCastError( vce *mg.ValueCastError ) ( err error ) {
    if err = w.startValueError( vce, tcVcErr ); err != nil { return }
    return w.writeEnd()
}

func ( w writer ) writeTypeCastError( tcErr *mg.TypeCastError ) ( err error ) {
    if err = w.startValueError( tcErr, tcTcErr ); err != nil { return }
    if err = w.writeTc( tcExpected ); err != nil { return }
    if err = w.writeType( tcErr.Expected ); err != nil { return }
    if err = w.writeTc( tcActual ); err != nil { return }
    if err = w.writeType( tcErr.Actual ); err != nil { return }
    return w.writeEnd()
}

func ( w writer ) writeErr( cvtErr interface{} ) ( err error ) {
    switch v := cvtErr.( type ) {
    case *mg.ValueCastError: return w.writeValueCastError( v )
    case *mg.TypeCastError: return w.writeTypeCastError( v )
    }
    panic( fmt.Sprintf( "Unhandled err type: %T", cvtErr ) )
}

func ( w writer ) writeTest( cvt *mg.CastValueTest ) ( err error ) {
    if err = w.writeTc( tcTest ); err != nil { return }
    if err = w.writeTc( tcIn ); err != nil { return }
    if err = w.writeValue( cvt.In ); err != nil { return }
    if v := cvt.Expect; v != nil { 
        if err = w.writeTc( tcExpect ); err != nil { return }
        if err = w.writeValue( v ); err != nil { return }
    }
    if err = w.writeTc( tcType ); err != nil { return }
    if err = w.writeType( cvt.Type ); err != nil { return }
    if err = w.writeTc( tcPath ); err != nil { return }
    if err = w.writePath( cvt.Path ); err != nil { return }
    if cvt.Err != nil {
        if err = w.writeTc( tcErr ); err != nil { return }
        if err = w.writeErr( cvt.Err ); err != nil { return }
    }
    return w.writeEnd()
}

func ( w writer ) writeTests() ( err error ) {
    if err = w.w.WriteUint32( fileVersion ); err != nil { return }
    for _, cvt := range mg.GetCastValueTests() {
        if err = w.writeTest( cvt ); err != nil { return }
    }
    return w.writeEnd()
}

func main() {
    tgf := testgen.NewOutFile()
    tgf.SetParseArg()
    flag.Parse()
    err := tgf.WithBinWriter( func( w *bgio.BinWriter ) error {
        return ( writer{ w: w, mgw: mg.AsWriter( w ) } ).writeTests()
    })
    if err != nil { log.Fatal( err ) }
}
