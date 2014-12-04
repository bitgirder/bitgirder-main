package golang

import (
    mg "mingle"
    "mingle/types"
    "mingle/codegen"
//    "bitgirder/stub"
    "log"
    "bytes"
    "fmt"
    "strings"
    "go/ast"
    "go/printer"
    "go/token"
)

var (
    identBool = ast.NewIdent( "bool" )
    identBuffer = &ast.ArrayType{ Elt: ast.NewIdent( "byte" ) }
    identString = ast.NewIdent( "string" )
    identInt32 = ast.NewIdent( "int32" )
    identUint32 = ast.NewIdent( "uint32" )
    identFloat32 = ast.NewIdent( "float32" )
    identInt64 = ast.NewIdent( "int64" )
    identUint64 = ast.NewIdent( "uint64" )
    identFloat64 = ast.NewIdent( "float64" )

    identTimeTime = &ast.SelectorExpr{
        X: ast.NewIdent( "time" ),
        Sel: ast.NewIdent( "Time" ),
    }

    identValue = &ast.InterfaceType{ Methods: &ast.FieldList{} }

    identSymbolMap = &ast.MapType{ Key: identString, Value: identValue }
)

type pkgGen struct {
    g *Generator
    ns *mg.Namespace
    defs []types.Definition
    file *ast.File
    pathIds []*mg.Identifier
    imports *mg.NamespaceMap // vals are *ast.ImportSpec
    idSeq int
}

// Instances are only good for a single call to Generate()
type Generator struct {
    Definitions *types.DefinitionMap
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
    qn *mg.QualifiedTypeName ) ( ast.Expr, bool ) {

    switch {
    case qn.Equals( mg.QnameBoolean ): return identBool, true
    case qn.Equals( mg.QnameBuffer ): return identBuffer, true 
    case qn.Equals( mg.QnameString ): return identString, true
    case qn.Equals( mg.QnameInt32 ): return identInt32, true 
    case qn.Equals( mg.QnameUint32 ): return identUint32, true
    case qn.Equals( mg.QnameFloat32 ): return identFloat32, true
    case qn.Equals( mg.QnameInt64 ): return identInt64, true 
    case qn.Equals( mg.QnameUint64 ): return identUint64, true
    case qn.Equals( mg.QnameFloat64 ): return identFloat64, true
    case qn.Equals( mg.QnameTimestamp ): return identTimeTime, true
    case qn.Equals( mg.QnameValue ): return identValue, true
    case qn.Equals( mg.QnameSymbolMap ): return identSymbolMap, true
    }
    if qn.Namespace.Equals( mg.CoreNsV1 ) {
        return ast.NewIdent( "stub" ), true
    }
    return nil, false
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
    pg.imports = mg.NewNamespaceMap()
    if err := pg.setPathIds(); err != nil { return err }
    pg.setFileName()
    return nil
}

func ( pg *pkgGen ) importForNs( ns *mg.Namespace ) *ast.ImportSpec {
    val, ok := pg.imports.GetOk( ns )
    if ! ok {
        spec := &ast.ImportSpec{
            Path: &ast.BasicLit{ 
                Kind: token.STRING,
                Value: pg.g.pkgPathStringFor( ns ),
            },
            Name: pg.nextPkgImportId(),
        }
        gd := &ast.GenDecl{ Tok: token.IMPORT, Specs: []ast.Spec{ spec } }
        pg.file.Decls = append( pg.file.Decls, gd )
        pg.imports.Put( ns, spec )
        val = spec
    }
    return val.( *ast.ImportSpec )
}

func ( pg *pkgGen ) pkgSelectorFor( ns *mg.Namespace ) *ast.Ident {
    return pg.importForNs( ns ).Name
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

    if expr, ok := pg.g.builtinTypeExpressionFor( at.Name() ); ok { 
        return expr 
    }
    typNm := pg.identForTypeName( at.Name() )
    ns := at.Name().Namespace
    if ns.Equals( pg.ns ) { return typNm }
    return &ast.SelectorExpr{ X: pg.pkgSelectorFor( ns ), Sel: typNm }
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
    bb := &bytes.Buffer{}
    fs := token.NewFileSet()
    if err := cfg.Fprint( bb, fs, pg.file ); err != nil { return err }
    log.Printf( "for %s:\n%s", pg.ns, bb )
    return nil
}

func ( g *Generator ) writeOutputs() error {
    return g.eachPkgGen( func( pg *pkgGen ) error { 
        return g.writeOutput( pg )
    })
}

func ( g *Generator ) Generate() error {
    g.setPackageMap()
    g.initPackages()
    if err := g.generatePackages(); err != nil { return err }
    if err := g.writeOutputs(); err != nil { return err }
    return nil
}
