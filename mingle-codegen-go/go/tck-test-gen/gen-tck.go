package main

import (
    "log"
    "flag"
    mgTck "mingle/tck"
    "strings"
    "mingle/types"
    "mingle/codegen/golang"
)

var destDir string

func init() {

    flag.StringVar( &destDir, "dest-dir", "", 
        "destput directory for built source" )
}

func getGenInput( defs *types.DefinitionMap ) []types.Definition {
    res := make( []types.Definition, 0, 16 )
    defs.EachDefinition( func( def types.Definition ) {
        nsStr := def.GetName().Namespace.ExternalForm()
        if strings.HasPrefix( nsStr, "mingle:tck" ) { res = append( res, def ) }
    })
    return res
}

func main() {
    flag.Parse()
    log.Printf( "generating to: %s", destDir )
    gen := golang.NewGenerator()
    gen.Types = mgTck.GetDefinitions()
    gen.Input = getGenInput( gen.Types )
    gen.DestDir = destDir
    if err := gen.Generate(); err != nil { log.Fatal( err ) }
}
