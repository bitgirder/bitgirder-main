package main

import (
    "log"
    "flag"
    mgTck "mingle/tck"
    "mingle/codegen/golang"
)

var destDir string

func init() {

    flag.StringVar( &destDir, "dest-dir", "", 
        "destput directory for built source" )
}

func main() {
    flag.Parse()
    log.Printf( "generating to: %s", destDir )
    gen := golang.NewGenerator()
    gen.Definitions = mgTck.GetDefinitions()
    gen.DestDir = destDir
    if err := gen.Generate(); err != nil { log.Fatal( err ) }
}
