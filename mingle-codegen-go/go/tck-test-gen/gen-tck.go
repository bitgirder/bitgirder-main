package main

import (
    "log"
    "flag"
    mg "mingle"
    mgTck "mingle/tck"
    _ "mingle/codegen/golang"
    "go/token"
    "go/ast"
    "go/printer"
    "bytes"
)

var outDir string

func init() {

    flag.StringVar( &outDir, "out-dir", "", 
        "output directory for built source" )
}

func main() {
    flag.Parse()
    log.Printf( "generating to: %s", outDir )
    log.Printf( "true: %v", mg.Boolean( true ) )
    log.Printf( "tck defs: %v", mgTck.GetDefinitions() )
    fs := token.NewFileSet()
    f1 := &ast.File{}
    f1.Name = ast.NewIdent( "test1" )
    bb := &bytes.Buffer{}
    cfg := &printer.Config{ Indent: 1, Tabwidth: 4, Mode: printer.UseSpaces }
    if err := cfg.Fprint( bb, fs, f1 ); err != nil {
        log.Fatal( err )
    }
    log.Printf( "source:\n%s", bb.String() )
}
