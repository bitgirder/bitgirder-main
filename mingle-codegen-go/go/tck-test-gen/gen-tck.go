package main

import (
    "log"
    "flag"
    mgTck "mingle/tck"
    "mingle/codegen/golang"
)

var outDir string

func init() {

    flag.StringVar( &outDir, "out-dir", "", 
        "output directory for built source" )
}

func main() {
    flag.Parse()
    log.Printf( "generating to: %s", outDir )
    gen := golang.NewGenerator()
    gen.Definitions = mgTck.GetDefinitions()
    if err := gen.Generate(); err != nil { log.Fatal( err ) }
}
