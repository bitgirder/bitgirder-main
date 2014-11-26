package main

import (
    "log"
    "flag"
    mg "mingle"
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
    log.Printf( "true: %v", mg.Boolean( true ) )
    log.Printf( "tck defs: %v", mgTck.GetDefinitions() )
    log.Printf( "stuff: %s", golang.Test1 )
}
