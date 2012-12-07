package testgen

import (
    "flag"
    "os"
    "fmt"
    "path/filepath"
    "log"
    bgio "bitgirder/io"
)

type OutFile struct {
    fname string
}

func NewOutFile() *OutFile {
    return &OutFile{}
}

func ( f *OutFile ) Name() string {
    if f.fname == "" { panic( libError( "No outfile set" ) ) }
    return f.fname
}

func ( f *OutFile ) SetParseArg() {
    flag.StringVar( &f.fname, "out-file", "", "Dest file for test data" )
}

func ( f *OutFile ) openBinWriter() ( wr *bgio.BinWriter, err error ) {
    if dir := filepath.Dir( f.fname ); dir != "." {
        if err = os.MkdirAll( dir, os.FileMode( 0755 ) ); err != nil {
            err = fmt.Errorf( "Couldn't create parent dir %s: %s", dir, err )
            return
        }
    }
    var fil *os.File
    if fil, err = os.Create( f.fname ); err != nil {
        err = fmt.Errorf( "Couldn't create %s: %s", f.fname, err )
        return
    } else { wr = bgio.NewLeWriter( fil ) }
    return 
}

func ( f *OutFile ) WithBinWriter( call func( *bgio.BinWriter ) error ) error {
    wr, err := f.openBinWriter()
    if err != nil { return err }
    defer wr.Close()
    log.Printf( "Writing %s", f.Name() )
    return call( wr )
}
