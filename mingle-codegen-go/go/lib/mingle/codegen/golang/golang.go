package golang

import (
    mg "mingle"
    "mingle/types"
    "mingle/codegen"
//    "bitgirder/stub"
    "os"
    "log"
    "bytes"
    "fmt"
    "strings"
    "go/ast"
    "go/printer"
    "go/token"
)

type builtinTypeExpr struct {
    goPath string
    typExpr ast.Expr
}

var (
    
    identBool = &builtinTypeExpr{ typExpr: ast.NewIdent( "bool" ) }

    identBuffer = &builtinTypeExpr{ 
        typExpr: &ast.ArrayType{ Elt: ast.NewIdent( "byte" ) },
    }

    identString = &builtinTypeExpr{ typExpr: ast.NewIdent( "string" ) }

    identInt32 = &builtinTypeExpr{ typExpr: ast.NewIdent( "int32" ) }
    
    identUint32 = &builtinTypeExpr{ typExpr: ast.NewIdent( "uint32" ) }
    
    identFloat32 = &builtinTypeExpr{ typExpr: ast.NewIdent( "float32" ) }

    identInt64 = &builtinTypeExpr{ typExpr: ast.NewIdent( "int64" ) }
    
    identUint64 = &builtinTypeExpr{ typExpr: ast.NewIdent( "uint64" ) }
    
    identFloat64 = &builtinTypeExpr{ typExpr: ast.NewIdent( "float64" ) }
    
    identTimeTime = &builtinTypeExpr{
        goPath: "time",
        typExpr: ast.NewIdent( "Time" ),
    }

    identValue = &builtinTypeExpr{
        typExpr: &ast.InterfaceType{ Methods: &ast.FieldList{} },
    }

    identSymbolMap = &builtinTypeExpr{
        typExpr: &ast.MapType{ 
            Key: ast.NewIdent( "string" ), 
            Value: identValue.typExpr,
        },
    }
)

type pkgGen struct {
    g *Generator
    ns *mg.Namespace
    defs []types.Definition
    file *ast.File
    pathIds []*mg.Identifier
    imports map[ string ] *ast.ImportSpec
    idSeq int
}

// Instances are only good for a single call to Generate()
type Generator struct {
    Definitions *types.DefinitionMap
    DestDir string
    pkgGens *mg.NamespaceMap
}

func NewGenerator() *Generator { return &Generator{} }

func ( g *Generator ) pkgPathStringFor( ns *mg.Namespace ) string {
    if val, ok := g.pkgGens.GetOk( ns ); ok {
        return val.( *pkgGen ).goPackagePath()
    }
    panic( libErrorf( "unrecognized namespace: %s", ns ) )
}

func ( g *Generator ) builtinTypeExpressionFor( 
    qn *mg.QualifiedTypeName ) *builtinTypeExpr {

    switch {
    case qn.Equals( mg.QnameBoolean ): return identBool
    case qn.Equals( mg.QnameBuffer ): return identBuffer 
    case qn.Equals( mg.QnameString ): return identString
    case qn.Equals( mg.QnameInt32 ): return identInt32 
    case qn.Equals( mg.QnameUint32 ): return identUint32
    case qn.Equals( mg.QnameFloat32 ): return identFloat32
    case qn.Equals( mg.QnameInt64 ): return identInt64 
    case qn.Equals( mg.QnameUint64 ): return identUint64
    case qn.Equals( mg.QnameFloat64 ): return identFloat64
    case qn.Equals( mg.QnameTimestamp ): return identTimeTime
    case qn.Equals( mg.QnameValue ): return identValue
    case qn.Equals( mg.QnameSymbolMap ): return identSymbolMap
    }
    return nil
}

func ( g *Generator ) setPackageMap() {
    g.pkgGens = mg.NewNamespaceMap()
    g.Definitions.EachDefinition( func( def types.Definition ) {
        var pg *pkgGen
        ns := def.GetName().Namespace
        if v, ok := g.pkgGens.GetOk( ns ); ok { 
            pg = v.( *pkgGen )
        } else { 
            pg = g.newPkgGen( ns ) 
            g.pkgGens.Put( ns, pg )
        }
        pg.defs = append( pg.defs, def )
    })
}

func ( g *Generator ) eachPkgGen( f func( *pkgGen ) error ) error {
    caller := func( _ *mg.Namespace, val interface{} ) error {
        return f( val.( *pkgGen ) )
    }
    return g.pkgGens.EachPairError( caller )
}

func ( g *Generator ) initPackages() {
    g.eachPkgGen( func( pg *pkgGen ) error { return pg.initPackage() } )
}

func ( g *Generator ) newPkgGen( ns *mg.Namespace ) *pkgGen {
    return &pkgGen{
        g: g,
        ns: ns,
        defs: make( []types.Definition, 0, 16 ),
    }
}

func ( pg *pkgGen ) nextId() *ast.Ident {
    res := ast.NewIdent( fmt.Sprintf( "mgCodegenId%d", pg.idSeq ) )
    pg.idSeq++
    return res
}

func ( pg *pkgGen ) nextPkgImportId() *ast.Ident { return pg.nextId() }

// eventually we'll look in here for a path mapper specific to pg.ns, and then
// would look for a path mapper in pg.g, and then finally use the default.
func ( pg *pkgGen ) setPathIds() ( err error ) {
    pg.pathIds, err = codegen.DefaultPathMapper.MapPath( pg.ns )
    return
}

func ( pg *pkgGen ) goPackagePath() string {
    res := make( []string, len( pg.pathIds ) )
    for i, id := range pg.pathIds { res[ i ] = id.Format( mg.LcCamelCapped ) }
    return strings.Join( res, "/" )
}

func ( pg *pkgGen ) setFileName() {
    idx := len( pg.pathIds ) - 1
    pg.file.Name = ast.NewIdent( pg.pathIds[ idx ].Format( mg.LcCamelCapped ) )
}

func ( pg *pkgGen ) initPackage() error {
    pg.file = &ast.File{}
    pg.file.Decls = make( []ast.Decl, 0, len( pg.defs ) )
    pg.imports = make( map[ string ] *ast.ImportSpec, 8 )
    if err := pg.setPathIds(); err != nil { return err }
    pg.setFileName()
    return nil
}

func ( pg *pkgGen ) importForGoPath( path string ) *ast.ImportSpec {
    res, ok := pg.imports[ path ]
    if ! ok {
        res = &ast.ImportSpec{
            Path: &ast.BasicLit{ Kind: token.STRING, Value: path },
            Name: pg.nextPkgImportId(),
        }
        gd := &ast.GenDecl{ Tok: token.IMPORT, Specs: []ast.Spec{ res } }
        pg.file.Decls = append( pg.file.Decls, gd )
        pg.imports[ path ] = res
    }
    return res
}

func ( pg *pkgGen ) pkgSelectorForGoPath( path string ) *ast.Ident {
    return pg.importForGoPath( path ).Name
}

func ( pg *pkgGen ) qualifiedTypeExpressionForGoPath(
    gp string, typNm *ast.Ident ) *ast.SelectorExpr {

    return &ast.SelectorExpr{ X: pg.pkgSelectorForGoPath( gp ), Sel: typNm }
}

func ( pg *pkgGen ) qualifiedTypeExpressionForNs(
    ns *mg.Namespace, typNm *ast.Ident ) *ast.SelectorExpr {
    
    gp := pg.g.pkgPathStringFor( ns )
    return pg.qualifiedTypeExpressionForGoPath( gp, typNm )
}

// if we later have a way for callers to customize name mappings, this is where
// that logic would be applied
func ( pg *pkgGen ) identForTypeName( qn *mg.QualifiedTypeName ) *ast.Ident {
    return ast.NewIdent( qn.Name.ExternalForm() )
}

// if we later allow custom per-field renaming, this is where it will happen
func ( pg *pkgGen ) identForField( 
    nm *mg.Identifier, encl types.Definition ) *ast.Ident {

    return ast.NewIdent( nm.Format( mg.LcCamelCapped ) )
}

func ( pg *pkgGen ) atomicTypeExpressionFor( 
    at *mg.AtomicTypeReference ) ast.Expr {

//    if expr, ok := pg.g.builtinTypeExpressionFor( at.Name() ); ok { 
    if bi := pg.g.builtinTypeExpressionFor( at.Name() ); bi != nil { 
        if bi.goPath == "" { return bi.typExpr }
        typNm := bi.typExpr.( *ast.Ident )
        return pg.qualifiedTypeExpressionForGoPath( bi.goPath, typNm )
    }
    typNm := pg.identForTypeName( at.Name() )
    ns := at.Name().Namespace
    if ns.Equals( pg.ns ) { return typNm }
    return pg.qualifiedTypeExpressionForNs( ns, typNm )
}

func ( pg *pkgGen ) listTypeExpressionFor( lt *mg.ListTypeReference ) ast.Expr {
    return &ast.ArrayType{ Elt: pg.typeExpressionFor( lt.ElementType ) }
}

func ( pg *pkgGen ) pointerTypeExpressionFor( 
    pt *mg.PointerTypeReference ) ast.Expr {

    return &ast.StarExpr{ X: pg.typeExpressionFor( pt.Type ) }
}

func ( pg *pkgGen ) typeExpressionFor( typ mg.TypeReference ) ast.Expr {
    switch v := typ.( type ) {
    case *mg.AtomicTypeReference: return pg.atomicTypeExpressionFor( v )
    case *mg.ListTypeReference: return pg.listTypeExpressionFor( v )
    case *mg.PointerTypeReference: return pg.pointerTypeExpressionFor( v )
    case *mg.NullableTypeReference: return pg.typeExpressionFor( v.Type )
    }
    panic( libErrorf( "unhandled type reference: %T", typ ) )
}

func ( pg *pkgGen ) addField( 
    flds *ast.FieldList, 
    fldDef *types.FieldDefinition,
    encl types.Definition ) {

    fld := &ast.Field{}
    fld.Names = []*ast.Ident{ pg.identForField( fldDef.Name, encl ) }
    fld.Type = pg.typeExpressionFor( fldDef.Type )
    flds.List = append( flds.List, fld )
}

func ( pg *pkgGen ) addFields( 
    flds *ast.FieldList, fs *types.FieldSet, encl types.Definition ) {

    fs.EachDefinition( func( fldDef *types.FieldDefinition ) {
        pg.addField( flds, fldDef, encl )
    })
}

func ( pg *pkgGen ) newStructDecl( sd *types.StructDefinition ) ast.Decl {
    typ := &ast.TypeSpec{}
    typ.Name = pg.identForTypeName( sd.GetName() )
    flds := &ast.FieldList{ List: make( []*ast.Field, 0, sd.Fields.Len() ) }
    pg.addFields( flds, sd.Fields, sd )
    typ.Type = &ast.StructType{ Fields: flds }
    return &ast.GenDecl{ Tok: token.TYPE, Specs: []ast.Spec{ typ } }
}

func ( pg *pkgGen ) generateStruct( sd *types.StructDefinition ) error {
    decl := pg.newStructDecl( sd )
    pg.file.Decls = append( pg.file.Decls, decl )
    return nil
}

func ( pg *pkgGen ) generateDef( def types.Definition ) error {
    switch v := def.( type ) {
    case *types.StructDefinition: return pg.generateStruct( v )
    }
    return nil
}

func ( pg *pkgGen ) generatePackage() error {
    for _, def := range pg.defs {
        if err := pg.generateDef( def ); err != nil { return err }
    }
    return nil
}

func ( g *Generator ) generatePackages() error {
    return g.eachPkgGen( func( pg *pkgGen ) error { 
        return pg.generatePackage()
    })
}

func ( g *Generator ) writeOutput( pg *pkgGen ) error {
    cfg := &printer.Config{ Tabwidth: 4, Mode: printer.UseSpaces, Indent: 1 }
    fs := token.NewFileSet()
    bb := &bytes.Buffer{}
    if err := cfg.Fprint( bb, fs, pg.file ); err != nil { return err }
    log.Printf( "for %s:\n%s", pg.ns, bb )
    dir := fmt.Sprintf( "%s/%s", g.DestDir, pg.goPackagePath() )
    if err := os.MkdirAll( dir, os.ModeDir | 0777 ); err != nil { return err }
    fname := fmt.Sprintf( "%s/%s.go", dir, "mingle_generated" )
    file, err := os.Create( fname )
    if err != nil { return err }
    defer file.Close()
    log.Printf( "writing to %s", fname )
    return cfg.Fprint( file, fs, pg.file )
}

func ( g *Generator ) writeOutputs() error {
    return g.eachPkgGen( func( pg *pkgGen ) error { 
        return g.writeOutput( pg )
    })
}

func ( g *Generator ) Generate() error {
    if g.DestDir == "" { panic( libError( "no dest dir for generator" ) ) }
    g.setPackageMap()
    g.initPackages()
    if err := g.generatePackages(); err != nil { return err }
    if err := g.writeOutputs(); err != nil { return err }
    return nil
}
