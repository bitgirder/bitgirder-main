package compiler

import (
    "testing"
    "fmt"
    "bytes"
    "bitgirder/assert"
    mg "mingle"
    "mingle/parser/tree"
    "mingle/types"
    "mingle/parser"
)

var mkQn = parser.MustQualifiedTypeName
var mkTyp = parser.MustTypeReference

func fldDef( nm, typ string, defl interface{} ) *types.FieldDefinition {
    return types.MakeFieldDef( nm, typ, defl )
}

func makeStructDefWithConstructors( 
    nm string, 
    flds []*types.FieldDefinition, 
    cons []*types.ConstructorDefinition ) *types.StructDefinition {

    res := types.MakeStructDef( nm, flds )
    res.Constructors = cons
    return res
}

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

type testSource struct {
    name string
    source string
}

func makeErrorKey( name string, line, col int, msg string ) string {
    return fmt.Sprintf( "%s:%d:%d:%q", name, line, col, msg )
}

type errorExpect struct {
    name string
    line, col int
    message string
}

func ( ee errorExpect ) key() string {
    return makeErrorKey( ee.name, ee.line, ee.col, ee.message )
}

type compilerTest struct {
    *assert.PathAsserter
    t *testing.T
    name string
    sources []testSource
    libs []testSource
    errs []errorExpect
    expctDefs *types.DefinitionMap
}

func newCompilerTest( name string ) *compilerTest {
    return &compilerTest{ 
        name: name, 
        sources: []testSource{},
        libs: []testSource{},
        errs: []errorExpect{},
        expctDefs: types.NewDefinitionMap(),
    }
}

func ( et *compilerTest ) errorf( tmpl string, argv ...interface{} ) {
    et.t.Errorf( et.name + ": " + tmpl, argv... )
}

func ( et *compilerTest ) addLib( name, src string ) *compilerTest {
    et.libs = append( et.libs, testSource{ name, src } )
    return et
}

func ( et *compilerTest ) addSource( name, src string ) *compilerTest {
    ets := testSource{ name: name, source: src }
    et.sources = append( et.sources, ets )
    return et
}

func ( et *compilerTest ) setSource( src string ) *compilerTest {
    if len( et.sources ) == 0 { 
        et.addSource( "<>", src ) 
    } else { panic( "Attempt to call setSource with sources already present" ) }
    return et
}

func ( et *compilerTest ) expectSrcError( 
    name string, line, col int, msg string ) *compilerTest {
    err := errorExpect{ name, line, col, msg }
    et.errs = append( et.errs, err )
    return et
}

func ( et *compilerTest ) expectError( 
    line, col int, msg string ) *compilerTest {
    return et.expectSrcError( "<>", line, col, msg )
}

func ( et *compilerTest ) expectDef( def types.Definition ) *compilerTest {
    et.expctDefs.MustAdd( def )
    return et
}

func ( et *compilerTest ) compile( 
    srcs []testSource, extTypes *types.DefinitionMap ) *CompilationResult {
    comp := NewCompilation()
    comp.SetExternalTypes( extTypes )
    for _, src := range srcs {
        rd := bytes.NewBufferString( src.source )
        if unit, err := tree.ParseSource( src.name, rd ); err == nil {
            comp.AddSource( unit )
        } else { et.Fatal( err ) }
    }
    cr, err := comp.Execute()
    if err == nil { return cr }
    et.Fatal( err )
    panic( "Unreached" )
}

func ( et *compilerTest ) compileResult() *CompilationResult {
    extTypes := types.CoreTypesV1()
    if len( et.libs ) > 0 {
        cr := et.compile( et.libs, extTypes )
        extTypes.MustAddFrom( cr.BuiltTypes )
    }
    return et.compile( et.sources, extTypes )
}

func ( et *compilerTest ) assertDefs( cr *CompilationResult ) {
    a := et.PathAsserter.Descend( "(expctDefs)" )
    et.expctDefs.EachDefinition( func( def types.Definition ) {
        nm := def.GetName()
        a2 := a.Descend( nm )
        if builtDef := cr.BuiltTypes.Get( nm ); builtDef == nil {
            a2.Fatalf( "not built" )
        } else { types.NewDefAsserter( a ).AssertDef( def, builtDef ) }
    })
}

func ( et *compilerTest ) makeErrorMap() map[ string ]errorExpect {
    res := make( map[ string ]errorExpect, len( et.errs ) )
    for _, err := range et.errs { res[ err.key() ] = err }
    return res
}

func ( et *compilerTest ) checkError( 
    err *Error, errMap map[ string ]errorExpect ) int {
    lc := err.Location
    k := makeErrorKey( lc.Source, lc.Line, lc.Col, err.Message )
    if _, ok := errMap[ k ]; ok {
        delete( errMap, k )
        return 0
    }
    et.errorf( "Unexpected compiler error: %s", err )
    return 1
}

func ( et *compilerTest ) call() {
    et.Log( "calling" )
    cr := et.compileResult()
    errMap := et.makeErrorMap()
    errCount := 0
    for _, err := range cr.Errors { errCount += et.checkError( err, errMap ) }
    for _, err := range errMap {
        et.errorf( "Error was not encountered: %v", err )
    }
    errCount += len( errMap )
    if errCount > 0 { et.t.FailNow() }
    et.assertDefs( cr )
}
