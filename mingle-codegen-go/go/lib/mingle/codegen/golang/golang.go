package golang

import (
    mg "mingle"
    "mingle/types"
    "mingle/codegen"
    "os"
    "log"
    "bytes"
    "fmt"
    "strings"
    "strconv"
//    "unicode"
    "go/ast"
    "go/printer"
    "go/token"
)

const (
    tmplGetter = "Get%s"
    tmplSetter = "Set%s"
    prefixGet = "Get"
    prefixSet = "Set"
)

func starExpr( expr ast.Expr ) *ast.StarExpr { return &ast.StarExpr{ X: expr } }

func selExpr( x ast.Expr, sel *ast.Ident ) *ast.SelectorExpr {
    return &ast.SelectorExpr{ x, sel }
}

func callExpr( fun ast.Expr, args... ast.Expr ) *ast.CallExpr {
    return &ast.CallExpr{ Fun: fun, Args: args }
}

func astField( nm *ast.Ident, typ ast.Expr ) *ast.Field {
    return &ast.Field{ Names: []*ast.Ident{ nm }, Type: typ }
}

func strLit( s string ) *ast.BasicLit {
    return &ast.BasicLit{ Kind: token.STRING, Value: strconv.Quote( s ) }
}

func fmtImplIdString( key string ) string {
    return fmt.Sprintf( "_mg%s", key )
}

func fmtImplIdStringf( tmpl string, argv ...interface{} ) string {
    return fmtImplIdString( fmt.Sprintf( tmpl, argv... ) )
}

func fmtImplIdSeq( key string, idx int ) string {
    return fmtImplIdString( fmt.Sprintf( "%s%d", key, idx ) )
}

var (
    goKwdNil = ast.NewIdent( "nil" )
    goLcError = ast.NewIdent( "error" )
    goLcInt = ast.NewIdent( "int" )
    goLcNew = ast.NewIdent( "new" )
    goLcReg = ast.NewIdent( "reg" )
    goLcRes = ast.NewIdent( "res" )
    goLcVal = ast.NewIdent( "val" )
    goLcVc = ast.NewIdent( "vc" )
    goPkgBind = "mingle/bind"
    goPkgParser = "mingle/parser"
    goUcCheckedStructFactory = ast.NewIdent( "CheckedStructFactory" )
    goUcMustAddValue = ast.NewIdent( "MustAddValue" )
    goUcMustQn = ast.NewIdent( "MustQualifiedTypeName" )
    goUcRegistry = ast.NewIdent( "Registry" )
    goUcVisitContext = ast.NewIdent( "VisitContext" )
    goUcVisitStruct = ast.NewIdent( "VisitStruct" )
    goUcVisitValue = ast.NewIdent( "VisitValue" )
    goUcMustRegistryForDomain = ast.NewIdent( "MustRegistryForDomain" )
    goUcDomainDefault = ast.NewIdent( "DomainDefault" )
    goLcInit = ast.NewIdent( "init" )
)

type importIdMap map[ string ] *ast.Ident

type builtinTypeExpr struct {
    goPaths []string // go paths that need to be imported for this builtin | nil
    typExpr func( m importIdMap ) ast.Expr // resolves this type using m
}

func newStaticBuiltinTypeExpr( e ast.Expr ) *builtinTypeExpr {
    return &builtinTypeExpr{
        typExpr: func( _ importIdMap ) ast.Expr { return e },
    }
}

var (
    
    identBool = newStaticBuiltinTypeExpr( ast.NewIdent( "bool" ) )

    identBuffer = newStaticBuiltinTypeExpr( 
        &ast.ArrayType{ Elt: ast.NewIdent( "byte" ) } )

    identString = newStaticBuiltinTypeExpr( ast.NewIdent( "string" ) )

    identInt32 = newStaticBuiltinTypeExpr( ast.NewIdent( "int32" ) )
    
    identUint32 = newStaticBuiltinTypeExpr( ast.NewIdent( "uint32" ) )
    
    identFloat32 = newStaticBuiltinTypeExpr( ast.NewIdent( "float32" ) )

    identInt64 = newStaticBuiltinTypeExpr( ast.NewIdent( "int64" ) )
    
    identUint64 = newStaticBuiltinTypeExpr( ast.NewIdent( "uint64" ) )
    
    identFloat64 = newStaticBuiltinTypeExpr( ast.NewIdent( "float64" ) )
    
    identTimeTime = &builtinTypeExpr{
        goPaths: []string{ "time" },
        typExpr: func( ids importIdMap ) ast.Expr {
            return selExpr( ids[ "time" ], ast.NewIdent( "Time" ) )
        },
    }

    goTypInterface = &ast.CompositeLit{ Type: ast.NewIdent( "interface" ) }

    identValue = newStaticBuiltinTypeExpr( goTypInterface )

    identSymbolMap = newStaticBuiltinTypeExpr(
        &ast.MapType{ Key: ast.NewIdent( "string" ), Value: goTypInterface },
    )
)

type pkgGen struct {
    g *Generator
    ns *mg.Namespace
    defs []types.Definition
    file *ast.File
    pathIds []*mg.Identifier
    importIds importIdMap
    importSpecs []*ast.ImportSpec
    decls []ast.Decl
    qnVars *mg.QnameMap
    regInitIds []*ast.Ident
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
        qnVars: mg.NewQnameMap(),
        regInitIds: make( []*ast.Ident, 0, 16 ),
    }
}

func ( pg *pkgGen ) addDecl( decl ast.Decl ) {
    pg.decls = append( pg.decls, decl )
}

func ( pg *pkgGen ) addRegInitId( nm *ast.Ident ) {
    pg.regInitIds = append( pg.regInitIds, nm )
}

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
    pg.decls = make( []ast.Decl, 0, len( pg.defs ) )
    if err := pg.setPathIds(); err != nil { return err }
    pg.setFileName()
    return nil
}

type importCollector struct {
    m map[ string ] string
    pg *pkgGen
    nextPrivId int
}

func ( c *importCollector ) collectGoPath( gp string ) {
    c.m[ gp ] = gp
}

func ( c *importCollector ) collectPkgPrivGoPath( gp string ) {
    c.m[ gp ] = fmtImplIdSeq( "Pkg", c.nextPrivId )
    c.nextPrivId++
}

func ( c *importCollector ) collectCodegenImplImports() {
    if len( c.pg.defs ) == 0 { return }
    c.collectGoPath( goPkgBind )
    c.collectPkgPrivGoPath( goPkgParser )
}

func ( c *importCollector ) collectQnameImport( qn *mg.QualifiedTypeName ) {
    if ns := qn.Namespace; ! ns.Equals( c.pg.ns ) {
        c.collectGoPath( c.pg.g.pkgPathStringFor( ns ) )
    }
}

func ( c *importCollector ) collectTypeImport( typ mg.TypeReference ) {
    qn := mg.TypeNameIn( typ )
    bi := c.pg.g.builtinTypeExpressionFor( qn )
    if bi == nil { 
        c.collectQnameImport( qn ) 
        return
    } 
    for _, gp := range bi.goPaths { c.collectGoPath( gp ) }
}

func ( c *importCollector ) collectFieldImports( fs *types.FieldSet ) {
    fs.EachDefinition( func( fd *types.FieldDefinition ) {
        c.collectTypeImport( fd.Type )
    })
}

func ( c *importCollector ) collectDefImports( def types.Definition ) {
    switch v := def.( type ) {
    case *types.StructDefinition: c.collectFieldImports( v.Fields )
    case *types.SchemaDefinition: c.collectFieldImports( v.Fields )
    }
}

func ( c *importCollector ) buildImports() {
    c.pg.importIds = make( map[ string ] *ast.Ident, len( c.m ) )
    c.pg.importSpecs = make( []*ast.ImportSpec, 0, len( c.m ) )
    for gp, alias := range c.m {
        spec := &ast.ImportSpec{ Path: strLit( gp ) }
        var importId *ast.Ident
        if alias == gp { 
            strs := strings.Split( alias, "/" )
            importId = ast.NewIdent( strs[ len( strs ) - 1 ] )
        } else {
            spec.Name = ast.NewIdent( alias ) 
            importId = spec.Name
        }
        c.pg.importIds[ gp ] = importId
        c.pg.importSpecs = append( c.pg.importSpecs, spec )
    }
}

func ( pg *pkgGen ) collectImports() {
    c := importCollector{ pg: pg, m: make( map[ string ] string, 8 ) }
    c.collectCodegenImplImports()
    for _, def := range pg.defs { c.collectDefImports( def ) }
    c.buildImports()
}

func ( pg *pkgGen ) pkgQnVar( qn *mg.QualifiedTypeName ) *ast.Ident {
    if res, ok := pg.qnVars.GetOk( qn ); ok { return res.( *ast.Ident ) }
    res := ast.NewIdent( fmtImplIdSeq( "Qn", pg.qnVars.Len() ) )
    pg.qnVars.Put( qn, res )
    return res
}

// if we later have a way for callers to customize name mappings, this is where
// that logic would be applied
func ( pg *pkgGen ) identForTypeName( qn *mg.QualifiedTypeName ) *ast.Ident {
    return ast.NewIdent( qn.Name.ExternalForm() )
}

func ( pg *pkgGen ) atomicTypeExpressionFor( 
    at *mg.AtomicTypeReference ) ast.Expr {

    if bi := pg.g.builtinTypeExpressionFor( at.Name() ); bi != nil { 
        return pg.typeExpressionFor( bi )
    }
    typNm := pg.identForTypeName( at.Name() )
    ns := at.Name().Namespace
    if ns.Equals( pg.ns ) { return typNm }
    gp := pg.g.pkgPathStringFor( ns )
    return selExpr( pg.importIds[ gp ], typNm )
}

func ( pg *pkgGen ) listTypeExpressionFor( lt *mg.ListTypeReference ) ast.Expr {
    return &ast.ArrayType{ Elt: pg.typeExpressionFor( lt.ElementType ) }
}

func ( pg *pkgGen ) pointerTypeExpressionFor( 
    pt *mg.PointerTypeReference ) ast.Expr {

    return starExpr( pg.typeExpressionFor( pt.Type ) )
}

func ( pg *pkgGen ) typeExpressionFor( typ interface{} ) ast.Expr {
    switch v := typ.( type ) {
    case *mg.AtomicTypeReference: return pg.atomicTypeExpressionFor( v )
    case *mg.ListTypeReference: return pg.listTypeExpressionFor( v )
    case *mg.PointerTypeReference: return pg.pointerTypeExpressionFor( v )
    case *mg.NullableTypeReference: return pg.typeExpressionFor( v.Type )
    case *builtinTypeExpr: return v.typExpr( pg.importIds )
    }
    panic( libErrorf( "unhandled type reference: %T", typ ) )
}

func ( pg *pkgGen ) createTypeSpecForDef( def types.Definition ) *ast.TypeSpec {
    res := &ast.TypeSpec{}
    res.Name = pg.identForTypeName( def.GetName() )
    return res
}

// must be called after collectImports()
func ( pg *pkgGen ) pkgSel( pkgPath string, sel *ast.Ident ) *ast.SelectorExpr {
    if pkgId, ok := pg.importIds[ pkgPath ]; ok { return selExpr( pkgId, sel ) }
    panic( libErrorf( "pkgGen for %s has no id for go package %s",
        pg.ns, pkgPath ) )
}

func ( pg *pkgGen ) pkgSelStar( pkgPath string, sel *ast.Ident ) *ast.StarExpr {
    return starExpr( pg.pkgSel( pkgPath, sel ) )
}

type bodyBuilder struct {
    block *ast.BlockStmt
}

func ( bb *bodyBuilder ) addStmt( stmt ast.Stmt ) {
    bb.block.List = append( bb.block.List, stmt )
}

func ( bb *bodyBuilder ) addExprStmt( expr ast.Expr ) {
    bb.addStmt( &ast.ExprStmt{ X: expr } )
}

func ( bb *bodyBuilder ) addAssignment1( lhs, rhs ast.Expr, tok token.Token ) {
    bb.addStmt(
        &ast.AssignStmt{
            Lhs: []ast.Expr{ lhs },
            Rhs: []ast.Expr{ rhs },
            Tok: tok,
        },
    )
}

func ( bb *bodyBuilder ) addAssign1( lhs, rhs ast.Expr ) {
    bb.addAssignment1( lhs, rhs, token.ASSIGN )
}

func ( bb *bodyBuilder ) addDefine1( lhs, rhs ast.Expr ) {
    bb.addAssignment1( lhs, rhs, token.DEFINE )
}

func ( bb *bodyBuilder ) addReturn( exprs... ast.Expr ) {
    bb.addStmt( &ast.ReturnStmt{ Results: exprs } )
}

func ( pg *pkgGen ) newBodyBuilder() *bodyBuilder {
    return &bodyBuilder{ 
        block: &ast.BlockStmt{ List: make( []ast.Stmt, 0, 4 ) },
    }
}

type fieldListBuilder struct {
    fl *ast.FieldList
}

func ( pg *pkgGen ) newFieldListBuilder() *fieldListBuilder {
    return &fieldListBuilder{ fl: &ast.FieldList{} }
}

func ( flb *fieldListBuilder ) addField( fld *ast.Field ) {
    if flb.fl.List == nil { flb.fl.List = make( []*ast.Field, 0, 4 ) }
    flb.fl.List = append( flb.fl.List, fld )
}

func ( flb *fieldListBuilder ) addNamed( nm *ast.Ident, typ ast.Expr ) {
    flb.addField( &ast.Field{ Names: []*ast.Ident{ nm }, Type: typ } )
}

func ( flb *fieldListBuilder ) addAnon( typ ast.Expr ) {
    flb.addField( &ast.Field{ Type: typ } )
}

type funcTypeBuilder struct {
    funcType *ast.FuncType
    res *fieldListBuilder
    params *fieldListBuilder
}

func ( pg *pkgGen ) newFuncTypeBuilder() *funcTypeBuilder {
    res := &funcTypeBuilder{
        funcType: &ast.FuncType{
            Params: &ast.FieldList{},
            Results: &ast.FieldList{},
        },
    }
    res.res = &fieldListBuilder{ res.funcType.Results }
    res.params = &fieldListBuilder{ res.funcType.Params }
    return res
}

type funcDeclBuilder struct {
    pg *pkgGen
    funcDecl *ast.FuncDecl
    recvIdent *ast.Ident
    ftb *funcTypeBuilder
    bbInst *bodyBuilder
}

func ( fdb *funcDeclBuilder ) setReceiver( varNm string, typ ast.Expr ) {
    fdb.recvIdent = ast.NewIdent( varNm )
    flb := fdb.pg.newFieldListBuilder()
    flb.addNamed( fdb.recvIdent, typ )
    fdb.funcDecl.Recv = flb.fl
}

func ( fdb *funcDeclBuilder ) setPtrReceiver( varNm string, typ ast.Expr ) {
    fdb.setReceiver( varNm, starExpr( typ ) )
}

func ( fdb *funcDeclBuilder ) bb() *bodyBuilder {
    if fdb.bbInst == nil { fdb.bbInst = fdb.pg.newBodyBuilder() }
    return fdb.bbInst
}

func ( fdb *funcDeclBuilder ) build() *ast.FuncDecl {
    fdb.funcDecl.Type = fdb.ftb.funcType
    if fdb.bbInst != nil { fdb.funcDecl.Body = fdb.bbInst.block }
    return fdb.funcDecl
}

func ( fdb *funcDeclBuilder ) buildLiteral() *ast.FuncLit {
    return &ast.FuncLit{ Type: fdb.ftb.funcType, Body: fdb.bbInst.block }
}

func ( pg *pkgGen ) newFuncDeclBuilder() *funcDeclBuilder {
    return &funcDeclBuilder{
        funcDecl: &ast.FuncDecl{},
        ftb: pg.newFuncTypeBuilder(),
        pg: pg,
    }
}
type fieldGen struct {
    fd *types.FieldDefinition
    goId *ast.Ident
    mgId *mg.Identifier
    typeExpr ast.Expr
}

func ( pg *pkgGen ) generateField( 
    fd *types.FieldDefinition, encl types.Definition ) *fieldGen {

    return &fieldGen{
        fd: fd,
        mgId: fd.Name,
        goId: ast.NewIdent( fd.Name.Format( mg.LcCamelCapped ) ),
        typeExpr: pg.typeExpressionFor( fd.Type ),
    }
}

func ( fg *fieldGen ) astField() *ast.Field {
    return astField( fg.goId, fg.typeExpr )
}

func ( fg *fieldGen ) accessorIdent( tmpl string ) *ast.Ident { 
    lcId := fg.mgId.Format( mg.LcCamelCapped )
    ucId := &bytes.Buffer{}
    ucId.WriteString( strings.ToUpper( lcId[ 0 : 1 ] ) )
    ucId.WriteString( lcId[ 1 : ] )
    return ast.NewIdent( fmt.Sprintf( tmpl, ucId.String() ) )
}

func ( pg *pkgGen ) generateFields( 
    fs *types.FieldSet, encl types.Definition ) []*fieldGen {

    res := make( []*fieldGen, 0, fs.Len() )
    fs.EachDefinition( func( fd *types.FieldDefinition ) {
        res = append( res, pg.generateField( fd, encl ) )
    })
    return res
}

func newAstFieldList( flds []*fieldGen ) *ast.FieldList {
    res := &ast.FieldList{ List: make( []*ast.Field, len( flds ) ) }
    for i, fld := range flds { res.List[ i ] = fld.astField() }
    return res
}

type structGen struct {
    pg *pkgGen
    sd *types.StructDefinition
    typ *ast.TypeSpec
    flds []*fieldGen
    constructorName *ast.Ident
}

func ( sg *structGen ) setFields() {
    sg.flds = sg.pg.generateFields( sg.sd.Fields, sg.sd )
}

func ( sg *structGen ) addDecl() {
    sg.typ.Type = &ast.StructType{ Fields: newAstFieldList( sg.flds ) }
    decl := &ast.GenDecl{ Tok: token.TYPE, Specs: []ast.Spec{ sg.typ } }
    sg.pg.addDecl( decl )
}

func ( sg *structGen ) addGetter( fg *fieldGen ) {
    fdb := sg.pg.newFuncDeclBuilder()
    fdb.setPtrReceiver( "s", sg.typ.Name )
    fdb.funcDecl.Name = fg.accessorIdent( tmplGetter )
    fdb.ftb.res.addAnon( fg.typeExpr )
    fdb.bb().addReturn( selExpr( fdb.recvIdent, fg.goId ) )
    sg.pg.addDecl( fdb.build() )
}

func ( sg *structGen ) addSetter( fg *fieldGen ) {
    fdb := sg.pg.newFuncDeclBuilder()
    fdb.setPtrReceiver( "s", sg.typ.Name )
    fdb.funcDecl.Name = fg.accessorIdent( tmplSetter )
    fdb.ftb.params.addNamed( goLcVal, fg.typeExpr )
    fdb.bb().addAssign1( selExpr( fdb.recvIdent, fg.goId ), goLcVal )
    sg.pg.addDecl( fdb.build() )
}

func ( sg *structGen ) addAccessors() {
    for _, fg := range sg.flds {
        sg.addGetter( fg )
        sg.addSetter( fg )
    }
}

// side effect: sets sg.constructorName
func ( sg *structGen ) addFactories() {
    fdb := sg.pg.newFuncDeclBuilder()
    sg.constructorName = ast.NewIdent( "New" + sg.typ.Name.String() )
    fdb.funcDecl.Name = sg.constructorName
    resTyp := starExpr( sg.typ.Name )
    fdb.ftb.res.addAnon( resTyp )
    fdb.bb().addDefine1( goLcRes, callExpr( goLcNew, sg.typ.Name ) )
    fdb.bb().addReturn( goLcRes )
    sg.pg.addDecl( fdb.build() )
}

func ( sg *structGen ) addFieldVisitStatements( bb *bodyBuilder ) {
}

func ( sg *structGen ) createVisitReturnStmt() ast.Expr {
    visitLit := sg.pg.newFuncDeclBuilder()
    visitLit.ftb.res.addAnon( goLcError )
    sg.addFieldVisitStatements( visitLit.bb() )
    visitLit.bb().addReturn( goKwdNil )
    return callExpr(
        sg.pg.pkgSel( goPkgBind, goUcVisitStruct ),
        goLcVc,
        sg.pg.pkgQnVar( sg.sd.GetName() ),
        visitLit.buildLiteral(),
    )
}

func ( sg *structGen ) addVisitor() {
    fdb := sg.pg.newFuncDeclBuilder()
    fdb.funcDecl.Name = goUcVisitValue
    fdb.setPtrReceiver( "s", sg.typ.Name )
    vcType := sg.pg.pkgSel( goPkgBind, goUcVisitContext )
    fdb.ftb.params.addNamed( goLcVc, vcType )
    fdb.ftb.res.addAnon( goLcError )
    fdb.bb().addReturn( sg.createVisitReturnStmt() )
    sg.pg.addDecl( fdb.build() )
}

func ( sg *structGen ) bindingInstFactLit() ast.Expr {
    fdb := sg.pg.newFuncDeclBuilder()
    fdb.ftb.res.addAnon( goTypInterface )
    fdb.bb().addReturn( callExpr( sg.constructorName ) )
    return fdb.buildLiteral()
}

func ( sg *structGen ) bindingMustAddValueStmt() ast.Stmt {
    return &ast.ExprStmt{
        callExpr( 
            selExpr( goLcReg, goUcMustAddValue ),
            sg.pg.pkgQnVar( sg.sd.GetName() ), 
            callExpr( 
                sg.pg.pkgSel( goPkgBind, goUcCheckedStructFactory ),
                goLcReg,
                sg.bindingInstFactLit(),
                goKwdNil,
            ),
        ),
    }
}

func ( sg *structGen ) addBinding() {
    fdb := sg.pg.newFuncDeclBuilder()
    nm := fmtImplIdStringf( "AddBindingFor%s", sg.sd.GetName().Name )
    fdb.funcDecl.Name = ast.NewIdent( nm )
    fdb.ftb.params.addNamed( 
        goLcReg, sg.pg.pkgSelStar( goPkgBind, goUcRegistry ) )
    fdb.bb().addStmt( sg.bindingMustAddValueStmt() )
    sg.pg.addDecl( fdb.build() )
    sg.pg.addRegInitId( fdb.funcDecl.Name )
}

func ( pg *pkgGen ) generateStruct( sd *types.StructDefinition ) {
    sg := &structGen{ pg: pg, sd: sd, typ: pg.createTypeSpecForDef( sd ) }
    sg.setFields()
    sg.addDecl()
    sg.addFactories()
    sg.addAccessors()
    sg.addVisitor()
    sg.addBinding()
}

func ( pg *pkgGen ) generateEnum( ed *types.EnumDefinition ) {
    typ := &ast.TypeSpec{}
    typ.Name = pg.identForTypeName( ed.GetName() )
    typ.Type = goLcInt
    decl := &ast.GenDecl{ Tok: token.TYPE, Specs: []ast.Spec{ typ } }
    pg.addDecl( decl )
}

func ( pg *pkgGen ) generateUnion( ud *types.UnionDefinition ) {
    typ := &ast.TypeSpec{}
    typ.Name = pg.identForTypeName( ud.GetName() )
    typ.Type = &ast.InterfaceType{ Methods: &ast.FieldList{} }
    decl := &ast.GenDecl{ Tok: token.TYPE, Specs: []ast.Spec{ typ } }
    pg.addDecl( decl )
}

func ( pg *pkgGen ) getSchemaMethods( 
    sd *types.SchemaDefinition ) *ast.FieldList {
 
    flb := pg.newFieldListBuilder()
    for _, fld := range pg.generateFields( sd.Fields, sd ) {
        ftb := pg.newFuncTypeBuilder()
        ftb.res.addAnon( fld.typeExpr )
        flb.addNamed( fld.accessorIdent( tmplGetter ), ftb.funcType )
    }
    return flb.fl
}

func ( pg *pkgGen ) generateSchema( sd *types.SchemaDefinition ) {
    typ := pg.createTypeSpecForDef( sd )
    methods := pg.getSchemaMethods( sd )
    typ.Type = &ast.InterfaceType{ Methods: methods }
    decl := &ast.GenDecl{ Tok: token.TYPE, Specs: []ast.Spec{ typ } }
    pg.addDecl( decl )
}

func ( pg *pkgGen ) generateDef( def types.Definition ) {
    switch v := def.( type ) {
    case *types.StructDefinition: pg.generateStruct( v )
    case *types.EnumDefinition: pg.generateEnum( v )
    case *types.UnionDefinition: pg.generateUnion( v )
    case *types.SchemaDefinition: pg.generateSchema( v )
    }
}

func ( pg *pkgGen ) addGenDecl( tok token.Token, spec ast.Spec ) {
    gd := &ast.GenDecl{ Tok: tok, Specs: []ast.Spec{ spec } }
    pg.file.Decls = append( pg.file.Decls, gd )
}

func ( pg *pkgGen ) assembleImportSpecs() {
    for _, spec := range pg.importSpecs { pg.addGenDecl( token.IMPORT, spec ) }
}

func ( pg *pkgGen ) valExprForVarObj( varObj interface{} ) ast.Expr {
    switch v := varObj.( type ) {
    case *mg.QualifiedTypeName:
        sel := pg.pkgSel( goPkgParser, goUcMustQn )
        return callExpr( sel, strLit( v.ExternalForm() ) )
    }
    panic( libErrorf( "unhandled varObj: %T", varObj ) )
}

func ( pg *pkgGen ) addVarDecl( varNm *ast.Ident, varObj interface{} ) {
    valSpec := &ast.ValueSpec{ 
        Names: []*ast.Ident{ varNm },
        Values: []ast.Expr{ pg.valExprForVarObj( varObj ) },
    }
    pg.addGenDecl( token.VAR, valSpec )
}

func ( pg *pkgGen ) assembleVarDecls() {
    pg.qnVars.EachPair( func( qn *mg.QualifiedTypeName, val interface{} ) {
        pg.addVarDecl( val.( *ast.Ident ), qn )
    })
}

func ( pg *pkgGen ) addInitFunc() {
    fdb := pg.newFuncDeclBuilder()
    fdb.funcDecl.Name = goLcInit
    fdb.bb().addDefine1( 
        goLcReg, 
        callExpr( 
            pg.pkgSel( goPkgBind, goUcMustRegistryForDomain ),
            pg.pkgSel( goPkgBind, goUcDomainDefault ),
        ),
    )
    for _, regInitId := range pg.regInitIds {
        fdb.bb().addExprStmt( callExpr( regInitId, goLcReg ) )
    }
    pg.decls = append( pg.decls, fdb.build() )
}

func ( pg *pkgGen ) assembleDecls() {
    pg.file.Decls = make( []ast.Decl, 0, 16 )
    pg.assembleImportSpecs()
    pg.assembleVarDecls()
    pg.addInitFunc()
    pg.file.Decls = append( pg.file.Decls, pg.decls... )
}

func ( pg *pkgGen ) generatePackage() {
    pg.collectImports()
    for _, def := range pg.defs { pg.generateDef( def ) }
    pg.assembleDecls()
}

func ( g *Generator ) generatePackages() error {
    return g.eachPkgGen( func( pg *pkgGen ) error { 
        pg.generatePackage()
        return nil
    })
}

func ( g *Generator ) dumpGeneratedSource( pg *pkgGen ) error {
    cfg := &printer.Config{ Tabwidth: 4, Mode: printer.UseSpaces, Indent: 1 }
    fs := token.NewFileSet()
    bb := &bytes.Buffer{}
    if err := cfg.Fprint( bb, fs, pg.file ); err != nil { return err }
    log.Printf( "for %s:\n%s", pg.ns, bb )
    return nil
}

func ( g *Generator ) writeOutput( pg *pkgGen ) error {
    if err := g.dumpGeneratedSource( pg ); err != nil { return err }
    cfg := &printer.Config{ Tabwidth: 4, Mode: printer.UseSpaces }
    dir := fmt.Sprintf( "%s/%s", g.DestDir, pg.goPackagePath() )
    if err := os.MkdirAll( dir, os.ModeDir | 0777 ); err != nil { return err }
    fname := fmt.Sprintf( "%s/%s.go", dir, "mingle_generated" )
    file, err := os.Create( fname )
    if err != nil { return err }
    defer file.Close()
    log.Printf( "writing to %s", fname )
    fs := token.NewFileSet()
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
