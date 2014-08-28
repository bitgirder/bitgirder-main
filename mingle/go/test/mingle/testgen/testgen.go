package testgen

import (
    "flag"
    "os"
    "fmt"
    "path/filepath"
    "log"
    "io"
    mg "mingle"
    mgIo "mingle/io"
    mgRct "mingle/reactor"
    parser "mingle/parser"
    bgio "bitgirder/io"
)

const typFileEnd = "mingle:testgen@v1/TestFileEnd"

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

func ( f *OutFile ) openIoWriter() ( wr io.Writer, err error ) {
    if dir := filepath.Dir( f.fname ); dir != "." {
        if err = os.MkdirAll( dir, os.FileMode( 0755 ) ); err != nil {
            err = fmt.Errorf( "Couldn't create parent dir %s: %s", dir, err )
            return
        }
    }
    if wr, err = os.Create( f.fname ); err != nil {
        err = fmt.Errorf( "Couldn't create %s: %s", f.fname, err )
    } 
    return
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

func ( f *OutFile ) openMgWriter() ( *mgIo.BinWriter, error ) {
    w, err := f.openIoWriter()
    if err != nil { return nil, err }
    return mgIo.NewWriter( w ), nil
}

func ( f *OutFile ) WithBinWriter( call func( *bgio.BinWriter ) error ) error {
    wr, err := f.openBinWriter()
    if err != nil { return err }
    defer wr.Close()
    log.Printf( "Writing %s", f.Name() )
    return call( wr )
}

// Meant to be run as the workhorse of some program's main()
func WriteOutFile( call func( w *mgIo.BinWriter ) error ) {
    tgf := NewOutFile()
    tgf.SetParseArg()
    flag.Parse()
    w, err := tgf.openMgWriter()
    if err != nil { log.Fatal( err ) }
    defer w.Close()
    log.Printf( "Writing %s", tgf.Name() )
    if err = call( w ); err != nil { log.Fatal( err ) }
}

type StructDataSource interface {
    Len() int
    StructAt( i int ) *mg.Struct
}

func WriteStructFile( data StructDataSource ) {
    WriteOutFile( func( w *mgIo.BinWriter ) error {
        rct := w.AsReactor()
        for i, e := 0, data.Len(); i < e ; i++ {
            s := data.StructAt( i )
            if err := mgRct.VisitValue( s, rct ); err != nil { return err }
        }
        return mgRct.VisitValue( parser.MustStruct( typFileEnd ), rct )
    })
}
