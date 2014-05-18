package compiler

import (
    "testing"
    "bytes"
    "bitgirder/assert"
    mg "mingle"
    "mingle/parser/tree"
    "mingle/types"
)

func idSetFor( m *mg.IdentifierMap ) []*mg.Identifier {
    res := make( []*mg.Identifier, 0, m.Len() )
    m.EachPair( func( id *mg.Identifier, _ interface{} ) {
        res = append( res, id )
    })
    return res
}

func compileSingle( src string, f assert.Failer ) *CompilationResult {
    bb := bytes.NewBufferString( src )
    nsUnit, err := tree.ParseSource( "<input>", bb )
    if err != nil { f.Fatal( err ) }
    comp := NewCompilation().
            AddSource( nsUnit ).
            SetExternalTypes( types.CoreTypesV1() )
    compRes, err := comp.Execute()
    if err != nil { f.Fatal( err ) }
    return compRes
}

func failCompilerTest( cr *CompilationResult, t *testing.T ) {
    for _, err := range cr.Errors { t.Error( err ) }
    t.FailNow()
}

func roundtripCompilation( 
    m *types.DefinitionMap, f assert.Failer ) *types.DefinitionMap {

//    bb := &bytes.Buffer{}
//    wr, rd := types.NewBinWriter( bb ), types.NewBinReader( bb )
//    if err := wr.WriteDefinitionMap( m ); err != nil { f.Fatal( err ) }
//    m2, err := rd.ReadDefinitionMap()
//    if err != nil { f.Fatal( err ) }
//    return m2
    return m
}
