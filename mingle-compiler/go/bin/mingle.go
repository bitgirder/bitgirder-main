package main

import (
    "log"
    "os"
    "errors"
    "mingle/parser/tree"
    "mingle/types"
    "mingle/compiler"
)

var mkErr = errors.New

func fail( err error ) { log.Fatal( err ) }

func checkOrFail( err error ) { if err != nil { fail( err ) } }

func createCompilation() *compiler.Compilation {
    res := compiler.NewCompilation()
    res.SetExternalTypes( types.CoreTypesV1() )
    return res
}

func getSourceFiles() ( []string, error ) {
    if len( os.Args ) == 1 { return nil, mkErr( "no input" ) }
    return os.Args[ 1 : ], nil
}

func addSourceFiles( c *compiler.Compilation, files []string ) error {
    for _, file := range files {
        r, err := os.Open( file )
        if err != nil { return err }
        nsUnit, err := tree.ParseSource( file, r )
        if err != nil { return err }
        c.AddSource( nsUnit )
    }
    return nil
}

func main() {
    c := createCompilation()
    srcFiles, err := getSourceFiles()
    checkOrFail( err )
    checkOrFail( addSourceFiles( c, srcFiles ) );  
    cr, err := c.Execute()
    checkOrFail( err )
    for _, ce := range cr.Errors { log.Print( ce ) }
}
