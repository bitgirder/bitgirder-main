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
    "sort"
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

func assignmentStmt1( lhs, rhs ast.Expr, tok token.Token ) *ast.AssignStmt {
    return &ast.AssignStmt{
        Lhs: []ast.Expr{ lhs },
        Rhs: []ast.Expr{ rhs },
        Tok: tok,
    }
}

func assignStmt1( lhs, rhs ast.Expr ) *ast.AssignStmt {
    return assignmentStmt1( lhs, rhs, token.ASSIGN )
}

func defineStmt1( lhs, rhs ast.Expr ) *ast.AssignStmt {
    return assignmentStmt1( lhs, rhs, token.DEFINE )
}

func keyValExpr( k, v ast.Expr ) *ast.KeyValueExpr {
    return &ast.KeyValueExpr{ Key: k, Value: v }
}

func typeAssertExpr( x, typ ast.Expr ) *ast.TypeAssertExpr {
    return &ast.TypeAssertExpr{ X: x, Type: typ }
}

func derefExpr( x ast.Expr ) ast.Expr { 
    return starExpr( &ast.ParenExpr{ X: x } )
}

func comment1( text string ) *ast.CommentGroup {
    return &ast.CommentGroup{
        List: []*ast.Comment{
            &ast.Comment{ Text: fmt.Sprintf( "// %s\n", text ) },
        },
    }
}

func comment1f( tmpl string, argv ...interface{} ) *ast.CommentGroup {
    return comment1( fmt.Sprintf( tmpl, argv... ) )
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

// drops nullables, converts all lists to allow empty, and drops atomic
// restrictions
func typeRepForType( typ mg.TypeReference ) mg.TypeReference {
    switch v := typ.( type ) {
    case *mg.AtomicTypeReference: return v.Name().AsAtomicType()
    case *mg.PointerTypeReference:
        return mg.NewPointerTypeReference( typeRepForType( v.Type ) )
    case *mg.NullableTypeReference: return typeRepForType( v.Type )
    case *mg.ListTypeReference:
        return &mg.ListTypeReference{ typeRepForType( v.ElementType ), false }
    }
    panic( libErrorf( "unhandled type: %T", typ ) )
}

func typeKeyForType( typ mg.TypeReference ) string {
    return typeRepForType( typ ).ExternalForm()
}

var (
    goKwdNil = ast.NewIdent( "nil" )
    goLcErr = ast.NewIdent( "err" )
    goLcError = ast.NewIdent( "error" )
    goLcFieldValue = ast.NewIdent( "fieldValue" )
    goLcInit = ast.NewIdent( "init" )
    goLcInt = ast.NewIdent( "int" )
    goLcNew = ast.NewIdent( "new" )
    goLcObj = ast.NewIdent( "obj" )
    goLcReg = ast.NewIdent( "reg" )
    goLcRes = ast.NewIdent( "res" )
    goLcVal = ast.NewIdent( "val" )
    goLcVc = ast.NewIdent( "vc" )
    goPkgBind = "mingle/bind"
    goPkgParser = "mingle/parser"
    goPkgReactor = "mingle/reactor"
    goPkgCodegen = "mingle/codegen"
    goUcMustBuilderFactoryForQname = 
        ast.NewIdent( "MustBuilderFactoryForQname" )
    goUcBuilderFactory = ast.NewIdent( "BuilderFactory" )
    goUcNewConverterFactory = ast.NewIdent( "NewConverterFactory" )
    goUcAsAtomicType = ast.NewIdent( "AsAtomicType" )
    goUcAssign = ast.NewIdent( "Assign" )
    goUcCheckedFieldSetter = ast.NewIdent( "CheckedFieldSetter" )
    goUcCheckedStructFactory = ast.NewIdent( "CheckedStructFactory" )
    goUcDomainDefault = ast.NewIdent( "DomainDefault" )
    goUcField = ast.NewIdent( "Field" )
    goUcMustAddValue = ast.NewIdent( "MustAddValue" )
    goUcMustIdentifier = ast.NewIdent( "MustIdentifier" )
    goUcMustQualifiedTypeName = ast.NewIdent( "MustQualifiedTypeName" )
    goUcMustRegistryForDomain = ast.NewIdent( "MustRegistryForDomain" )
    goUcRegistry = ast.NewIdent( "Registry" )
    goUcType = ast.NewIdent( "Type" )
    goUcVisitContext = ast.NewIdent( "VisitContext" )
    goUcVisitFieldValue = ast.NewIdent( "VisitFieldValue" )
    goUcVisitStruct = ast.NewIdent( "VisitStruct" )
    goUcVisitValue = ast.NewIdent( "VisitValue" )
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

    goTypeInterface = &ast.CompositeLit{ Type: ast.NewIdent( "interface" ) }

    identValue = newStaticBuiltinTypeExpr( goTypeInterface )

    identSymbolMap = newStaticBuiltinTypeExpr(
        &ast.MapType{ Key: ast.NewIdent( "string" ), Value: goTypeInterface },
    )
)

type sequencedId struct {
    key string
    i int
}

func ( id sequencedId ) goId() *ast.Ident {
    return ast.NewIdent( fmtImplIdSeq( id.key, id.i ) )
}

type idSequence struct {
    key string
    idx int
}

func ( s *idSequence ) next() sequencedId {
    res := sequencedId{ key: s.key, i: s.idx }
    s.idx++
    return res
}

type typeKeyMap map[ string ] mg.TypeReference

func ( m typeKeyMap ) putType( typ mg.TypeReference ) {
    m[ typ.ExternalForm() ] = typ
}

type pkgGen struct {
    g *Generator
    ns *mg.Namespace
    defs []types.Definition
    file *ast.File
    pathIds []*mg.Identifier
    importIds importIdMap
    importSpecs []*ast.ImportSpec
    decls []ast.Decl
    idVars *mg.IdentifierMap
    idSeq idSequence
    qnVars *mg.QnameMap
    qnSeq idSequence
    atomicTypeVars *mg.QnameMap
    atomicTypeSeq idSequence
    defaultFactoryIdVars *mg.QnameMap
    defaultFactoryIdSeq idSequence
    regInitIds []*ast.Ident
    typeKeys typeKeyMap
    builderFactoryIds map[ string ] *ast.Ident
}

// Instances are only good for a single call to Generate()
type Generator struct {
    Types *types.DefinitionMap
    Input []types.Definition
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
    for _, def := range g.Input {
        var pg *pkgGen
        ns := def.GetName().Namespace
        if v, ok := g.pkgGens.GetOk( ns ); ok { 
            pg = v.( *pkgGen )
        } else { 
            pg = g.newPkgGen( ns ) 
            g.pkgGens.Put( ns, pg )
        }
        pg.defs = append( pg.defs, def )
    }
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
        idVars: mg.NewIdentifierMap(),
        idSeq: idSequence{ key: "Id" },
        qnVars: mg.NewQnameMap(),
        qnSeq: idSequence{ key: "Qn" },
        atomicTypeVars: mg.NewQnameMap(),
        atomicTypeSeq: idSequence{ key: "AtomicType" },
        defaultFactoryIdVars: mg.NewQnameMap(),
        defaultFactoryIdSeq: idSequence{ key: "DefaultBuilderFactory" },
        regInitIds: make( []*ast.Ident, 0, 16 ),
        typeKeys: typeKeyMap( make( map[ string ] mg.TypeReference ) ),
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

// the case in which no def is found likely indicates an error elsewhere, but we
// don't attempt to handle it here
func ( g *Generator ) isPrimitiveQn( qn *mg.QualifiedTypeName ) bool {
    if def, ok := g.Types.GetDefinition( qn ); ok {
        _, isPrim := def.( *types.PrimitiveDefinition )
        return isPrim
    }
    return false
}

// see error handling note at g.isPrimitiveQn()
func ( g *Generator ) isStructQn( qn *mg.QualifiedTypeName ) bool {
    if def, ok := g.Types.GetDefinition( qn ); ok {
        _, isStruct := def.( *types.StructDefinition )
        return isStruct
    }
    return false
}

// see error handling note at g.isPrimitiveQn()
func ( g *Generator ) isEnumQn( qn *mg.QualifiedTypeName ) bool {
    if def, ok := g.Types.GetDefinition( qn ); ok {
        _, isEnum := def.( *types.EnumDefinition )
        return isEnum
    }
    return false
}

func ( pg *pkgGen ) atomicTypeUsesDefaultFactory( 
    at *mg.AtomicTypeReference ) bool {

    qn := at.Name() 
    return pg.g.isPrimitiveQn( qn ) || pg.g.isEnumQn( qn )
}

func ( pg *pkgGen ) pointerTypeUsesDefaultFactory( 
    baseTyp mg.TypeReference ) bool {

    if at, ok := baseTyp.( *mg.AtomicTypeReference ); ok {
        return pg.g.isStructQn( at.Name() )
    }
    return false
}

func ( pg *pkgGen ) typeUsesDefaultFactory( typ mg.TypeReference ) bool {
    switch v := typ.( type ) {
    case *mg.AtomicTypeReference: return pg.atomicTypeUsesDefaultFactory( v )
    case *mg.PointerTypeReference: 
        return pg.pointerTypeUsesDefaultFactory( v.Type )
    case *mg.NullableTypeReference: return pg.typeUsesDefaultFactory( v.Type )
    }
    return false
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
    c.collectPkgPrivGoPath( goPkgReactor )
    c.collectPkgPrivGoPath( goPkgCodegen )
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

func ( pg *pkgGen ) putTypeKey( typ mg.TypeReference ) {
    pg.typeKeys[ typ.ExternalForm() ] = typ
}

func ( pg *pkgGen ) collectTypeKeysForTypeRep( typ mg.TypeReference ) {
    if pg.typeUsesDefaultFactory( typ ) { return }
    switch v := typ.( type ) {
    case *mg.AtomicTypeReference: pg.putTypeKey( v )
    case *mg.PointerTypeReference: 
        pg.putTypeKey( v )
        pg.collectTypeKeysForTypeRep( v.Type )
    case *mg.NullableTypeReference: pg.collectTypeKeysForTypeRep( v.Type )
    case *mg.ListTypeReference:
        pg.putTypeKey( v )
        pg.collectTypeKeysForTypeRep( v.ElementType )
    }
}

func ( pg *pkgGen ) collectTypeKeyForType( typ mg.TypeReference ) {
    pg.collectTypeKeysForTypeRep( typeRepForType( typ ) )
}

func ( pg *pkgGen ) collectTypeKeysForFieldSet( fs *types.FieldSet ) {
    fs.EachDefinition( func( fd *types.FieldDefinition ) {
        pg.collectTypeKeyForType( fd.Type )
    })
}

func ( pg *pkgGen ) collectTypeKeysForDef( def types.Definition ) {
    switch v := def.( type ) {
    case *types.StructDefinition: pg.collectTypeKeysForFieldSet( v.Fields )
    case *types.SchemaDefinition: pg.collectTypeKeysForFieldSet( v.Fields )
    }
}

func ( pg *pkgGen ) collectTypeKeys() {
    for _, def := range pg.defs { pg.collectTypeKeysForDef( def ) }
}

func ( pg *pkgGen ) pkgIdVar( id *mg.Identifier ) *ast.Ident {
    res, ok := pg.idVars.GetOk( id )
    if ! ok {
        res = pg.idSeq.next()
        pg.idVars.Put( id, res )
    }
    return res.( sequencedId ).goId()
}

func ( pg *pkgGen ) pkgQnVar( qn *mg.QualifiedTypeName ) *ast.Ident {
    res, ok := pg.qnVars.GetOk( qn )
    if ! ok {
        res = pg.qnSeq.next()
        pg.qnVars.Put( qn, res )
    }
    return res.( sequencedId ).goId()
}

func ( pg *pkgGen ) ensureQnVar( qn *mg.QualifiedTypeName ) {
    pg.pkgQnVar( qn )
}

func ( pg *pkgGen ) pkgDefaultFactoryIdVar( 
    qn *mg.QualifiedTypeName ) *ast.Ident {

    res, ok := pg.defaultFactoryIdVars.GetOk( qn )
    if ! ok {
        pg.ensureQnVar( qn )
        res = pg.defaultFactoryIdSeq.next()
        pg.defaultFactoryIdVars.Put( qn, res )
    }
    return res.( sequencedId ).goId()
}

// returns id for a var of type AtomicTypeReference corresponding to at,
// ignoring at.Restriction
func ( pg *pkgGen ) pkgAtomicTypeVar( at *mg.AtomicTypeReference ) ast.Expr {
    res, ok := pg.atomicTypeVars.GetOk( at.Name() )
    if ! ok {
        pg.ensureQnVar( at.Name() )
        res = pg.atomicTypeSeq.next()
        pg.atomicTypeVars.Put( at.Name(), res )
    }
    return res.( sequencedId ).goId()
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
    bb.addStmt( assignmentStmt1( lhs, rhs, tok ) )
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

func ( pg *pkgGen ) mustBuilderFactoryIdForType( 
    typ mg.TypeReference ) *ast.Ident {

    typeKey := typeKeyForType( typ )
    res, ok := pg.builderFactoryIds[ typeKey ]
    if ! ok { 
        tmpl := "no builder fact func with key %s for type %s"
        panic( libErrorf( tmpl, typeKey, typ ) )
    }
    return res 
}

// at the moment, the only builtin types we use have builder factories are
// atomic primitives or are pointers to struct types, both of which are keyed by
// qname, so we have a simple codepath here for the moment.
func ( pg *pkgGen ) defaultBuilderFactoryExprForType( 
    typ mg.TypeReference ) ast.Expr {

    return pg.pkgDefaultFactoryIdVar( mg.TypeNameIn( typ ) )
}

func ( pg *pkgGen ) builderFactoryExprForType( typ mg.TypeReference ) ast.Expr {
    if pg.typeUsesDefaultFactory( typ ) {
        return pg.defaultBuilderFactoryExprForType( typ )
    }
    return callExpr( pg.mustBuilderFactoryIdForType( typ ), goLcReg )
}

func ( pg *pkgGen ) addAtomicBuilderFactoryFuncBody(
    bb *bodyBuilder, at *mg.AtomicTypeReference ) error {

    bb.addReturn( goKwdNil )
    return nil
}

func ( pg *pkgGen ) addPointerBuilderFactoryFuncBody(
    bb *bodyBuilder, baseTyp mg.TypeReference ) error {

    retExp := callExpr(
        pg.pkgSel( goPkgReactor, goUcNewConverterFactory ),
        pg.builderFactoryExprForType( baseTyp ),
        goKwdNil,
    )
//    bf BuilderFactory, cf ValueConverterFunc ) BuilderFactory {
    bb.addReturn( retExp )
    return nil
}

func ( pg *pkgGen ) addListBuilderFactoryFuncBody(
    bb *bodyBuilder, eltTyp mg.TypeReference ) error {

    bb.addReturn( goKwdNil )
    return nil
}

func ( pg *pkgGen ) addBuilderFactoryFuncBody(
    bb *bodyBuilder, typeKey string ) error {

    switch v := pg.typeKeys[ typeKey ].( type ) {
    case *mg.AtomicTypeReference:
        return pg.addAtomicBuilderFactoryFuncBody( bb, v )
    case *mg.PointerTypeReference:
        return pg.addPointerBuilderFactoryFuncBody( bb, v.Type )
    case *mg.ListTypeReference:
        return pg.addListBuilderFactoryFuncBody( bb, v.ElementType )
    }
    panic( libErrorf( "unexpected typeKey for builder fact: %s", typeKey ) )
}

func ( pg *pkgGen ) addBuilderFactoryFunc( typeKey string ) error {
    fdb := pg.newFuncDeclBuilder()
    fdb.funcDecl.Name = pg.builderFactoryIds[ typeKey ]
    fdb.funcDecl.Doc = comment1f( "builds %s", typeKey )
    fdb.ftb.res.addAnon( pg.pkgSel( goPkgReactor, goUcBuilderFactory ) )
    regType := pg.pkgSelStar( goPkgBind, goUcRegistry )
    fdb.ftb.params.addNamed( goLcReg, regType )
    if err := pg.addBuilderFactoryFuncBody( fdb.bb(), typeKey ); err != nil {
        return err
    }
    pg.addDecl( fdb.build() )
    return nil
}

// first we declare all initialize all factory func names, then build the actual
// defs (some of which may refer to others)
func ( pg *pkgGen ) addBuilderFactoryFuncs() error {
    sz := len( pg.typeKeys )
    pg.builderFactoryIds = make( map[ string ] *ast.Ident, sz )
    seq := &idSequence{ key: "BuilderFactoryFunc" }
    buildOrder := make( []string, 0, sz )
    for typeKey, _ := range pg.typeKeys { 
        pg.builderFactoryIds[ typeKey ] = seq.next().goId()
        buildOrder = append( buildOrder, typeKey )
    }
    for _, typeKey := range buildOrder {
        if err := pg.addBuilderFactoryFunc( typeKey ); err != nil { return err }
    }
    return nil
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

func ( sg *structGen ) ptrType() ast.Expr { return starExpr( sg.typ.Name ) }

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
    resTyp := sg.ptrType()
    fdb.ftb.res.addAnon( resTyp )
    fdb.bb().addDefine1( goLcRes, callExpr( goLcNew, sg.typ.Name ) )
    fdb.bb().addReturn( goLcRes )
    sg.pg.addDecl( fdb.build() )
}

func ( sg *structGen ) fieldVisitValueExprForField(
    fg *fieldGen, val ast.Expr ) ast.Expr {
 
    return callExpr(
        sg.pg.pkgSel( goPkgBind, goUcVisitFieldValue ),
        goLcVc,
        sg.pg.pkgIdVar( fg.mgId ),
        val,
    )
}

func ( sg *structGen ) fieldVisitExprForPointerType( 
    fg *fieldGen, 
    baseTyp mg.TypeReference, 
    val ast.Expr ) ( ast.Expr, error ) {

    if at, ok := baseTyp.( *mg.AtomicTypeReference ); ok {
        def, ok := sg.pg.g.Types.GetDefinition( at.Name() )
        if ! ok { return nil, libErrorf( "no definition for %s", at.Name() ) }
        if _, ok := def.( *types.StructDefinition ); ! ok {
            return sg.fieldVisitExprForType( fg, baseTyp, derefExpr( val ) )
        }
    }
    return sg.fieldVisitValueExprForField( fg, val ), nil
}

func ( sg *structGen ) fieldVisitExprForType( 
    fg *fieldGen, typ mg.TypeReference, valExpr ast.Expr ) ( ast.Expr, error ) {

    switch v := typ.( type ) {
    case *mg.PointerTypeReference: 
        return sg.fieldVisitExprForPointerType( fg, v.Type, valExpr )
    }
    return sg.fieldVisitValueExprForField( fg, valExpr ), nil
}

func ( sg *structGen ) fieldVisitExprForField( 
    fg *fieldGen, valId *ast.Ident ) ( ast.Expr, error ) {

    return sg.fieldVisitExprForType( fg, fg.fd.Type, selExpr( valId, fg.goId ) )
}

func ( sg *structGen ) addFieldVisitStatements( 
    valId *ast.Ident, bb *bodyBuilder ) error {

    for _, fg := range sg.flds {
        visExpr, err := sg.fieldVisitExprForField( fg, valId )
        if err != nil { return err }
        errRetBody := sg.pg.newBodyBuilder()
        errRetBody.addReturn( goLcErr )
        bb.addStmt(
            &ast.IfStmt{
                Init: defineStmt1( goLcErr, visExpr ),
                Cond: &ast.BinaryExpr{ X: goLcErr, Op: token.NEQ, Y: goKwdNil },
                Body: errRetBody.block,
            },
        )
    }
    return nil
}

func ( sg *structGen ) createVisitReturnStmt( 
    valId *ast.Ident ) ( ast.Expr, error ) {

    visitLit := sg.pg.newFuncDeclBuilder()
    visitLit.ftb.res.addAnon( goLcError )
    if err := sg.addFieldVisitStatements( valId, visitLit.bb() ); err != nil {
        return nil, err
    }
    visitLit.bb().addReturn( goKwdNil )
    res := callExpr(
        sg.pg.pkgSel( goPkgBind, goUcVisitStruct ),
        goLcVc,
        sg.pg.pkgQnVar( sg.sd.GetName() ),
        visitLit.buildLiteral(),
    )
    return res, nil
}

func ( sg *structGen ) addVisitor() error {
    fdb := sg.pg.newFuncDeclBuilder()
    fdb.funcDecl.Name = goUcVisitValue
    fdb.setPtrReceiver( "s", sg.typ.Name )
    vcType := sg.pg.pkgSel( goPkgBind, goUcVisitContext )
    fdb.ftb.params.addNamed( goLcVc, vcType )
    fdb.ftb.res.addAnon( goLcError )
    if retExp, err := sg.createVisitReturnStmt( fdb.recvIdent ); err == nil {
        fdb.bb().addReturn( retExp )
    } else { return err }
    sg.pg.addDecl( fdb.build() )
    return nil
}

func ( sg *structGen ) bindingInstFactLit() ast.Expr {
    fdb := sg.pg.newFuncDeclBuilder()
    fdb.ftb.res.addAnon( goTypeInterface )
    fdb.bb().addReturn( callExpr( sg.constructorName ) )
    return fdb.buildLiteral()
}

func ( sg *structGen ) getBindingFieldActionForAtomicType(
    at *mg.AtomicTypeReference, fg *fieldGen ) *ast.KeyValueExpr {

    return keyValExpr( goUcType, sg.pg.pkgAtomicTypeVar( at ) )
}

func ( sg *structGen ) getBindingFieldActionForType( 
    typ mg.TypeReference, fg *fieldGen ) *ast.KeyValueExpr {

    switch v := typ.( type ) {
    case *mg.AtomicTypeReference:
        return sg.getBindingFieldActionForAtomicType( v, fg )
    case *mg.NullableTypeReference:
        return sg.getBindingFieldActionForType( v.Type, fg )
    }
    // stub for development
    return keyValExpr( goUcType, sg.pg.pkgAtomicTypeVar( mg.TypeValue ) )
}

func ( sg *structGen ) getBindingFieldAssignExpr( fg *fieldGen ) ast.Expr {
    fdb := sg.pg.newFuncDeclBuilder()
    fdb.ftb.params.addNamed( goLcObj, goTypeInterface )
    fdb.ftb.params.addNamed( goLcFieldValue, goTypeInterface )
    fdb.bb().addAssign1( 
        selExpr( typeAssertExpr( goLcObj, sg.ptrType() ), fg.goId ),
        typeAssertExpr( goLcFieldValue, fg.typeExpr ),
    )
    return fdb.buildLiteral()
}

func ( sg *structGen ) bindingAppendFieldSetters( argv []ast.Expr ) []ast.Expr {
    for _, fg := range sg.flds {
        elts := make( []ast.Expr, 0, 3 )
        fldField := keyValExpr( goUcField, sg.pg.pkgIdVar( fg.mgId ) )
        elts = append( elts, fldField )
        elts = append( elts, sg.getBindingFieldActionForType( fg.fd.Type, fg ) )
        assignExpr := sg.getBindingFieldAssignExpr( fg )
        elts = append( elts, keyValExpr( goUcAssign, assignExpr ) )
        argv = append( argv, &ast.UnaryExpr{
            Op: token.AND,
            X: &ast.CompositeLit{
                Type: sg.pg.pkgSel( goPkgBind, goUcCheckedFieldSetter ),
                Elts: elts,
            },
        })
    }
    return argv
}

func ( sg *structGen ) bindingMustAddValueStmt() ast.Stmt {
    callArgs := []ast.Expr{ goLcReg, sg.bindingInstFactLit(), goKwdNil }
    callArgs = sg.bindingAppendFieldSetters( callArgs )
    return &ast.ExprStmt{
        callExpr( 
            selExpr( goLcReg, goUcMustAddValue ),
            sg.pg.pkgQnVar( sg.sd.GetName() ), 
            callExpr( 
                sg.pg.pkgSel( goPkgBind, goUcCheckedStructFactory ),
                callArgs...,
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

func ( pg *pkgGen ) generateStruct( sd *types.StructDefinition ) error {
    sg := &structGen{ pg: pg, sd: sd, typ: pg.createTypeSpecForDef( sd ) }
    sg.setFields()
    sg.addDecl()
    sg.addFactories()
    sg.addAccessors()
    if err := sg.addVisitor(); err != nil { return err }
    sg.addBinding()
    return nil
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

func ( pg *pkgGen ) generateDef( def types.Definition ) error {
    switch v := def.( type ) {
    case *types.StructDefinition: return pg.generateStruct( v ) 
    case *types.EnumDefinition: pg.generateEnum( v )
    case *types.UnionDefinition: pg.generateUnion( v )
    case *types.SchemaDefinition: pg.generateSchema( v )
    }
    return nil
}

func ( pg *pkgGen ) addGenDecl( tok token.Token, spec ast.Spec ) {
    gd := &ast.GenDecl{ Tok: tok, Specs: []ast.Spec{ spec } }
    pg.file.Decls = append( pg.file.Decls, gd )
}

func ( pg *pkgGen ) assembleImportSpecs() {
    for _, spec := range pg.importSpecs { pg.addGenDecl( token.IMPORT, spec ) }
}

func ( pg *pkgGen ) mustParseExpr( nm *ast.Ident, lit string ) ast.Expr {
    return callExpr( pg.pkgSel( goPkgParser, nm ), strLit( lit ) )
}

type atomicVarDecl struct {
    qn *mg.QualifiedTypeName
}

type varDeclType struct { 
    typ ast.Expr
}

func ( pg *pkgGen ) valExprForVarObj( varObj interface{} ) ast.Expr {
    switch v := varObj.( type ) {
    case *mg.Identifier:
        return pg.mustParseExpr( goUcMustIdentifier, v.ExternalForm() )
    case *mg.QualifiedTypeName: 
        return pg.mustParseExpr( goUcMustQualifiedTypeName, v.ExternalForm() )
    case atomicVarDecl: 
        return callExpr( selExpr( pg.pkgQnVar( v.qn ), goUcAsAtomicType ) )
    }
    return nil
}

func ( pg *pkgGen ) valTypeForVarObj( varObj interface{} ) ast.Expr {
    switch v := varObj.( type ) {
    case varDeclType: return v.typ
    }
    return nil
}

type sequencedVarDecl struct {
    id sequencedId
    varObj interface{}
}

type sequencedVarDeclSort []sequencedVarDecl

func ( s sequencedVarDeclSort ) Len() int { return len( s ) }

func ( s sequencedVarDeclSort ) Less( i, j int ) bool {
    return s[ i ].id.i < s[ j ].id.i
}

func ( s sequencedVarDeclSort ) Swap( i, j int ) {
    s[ i ], s[ j ] = s[ j ], s[ i ]
}

func ( pg *pkgGen ) addVarDecls( decls []sequencedVarDecl ) {
    sort.Sort( sequencedVarDeclSort( decls ) )
    for _, decl := range decls { 
        valSpec := &ast.ValueSpec{ Names: []*ast.Ident{ decl.id.goId() } }
        if exp := pg.valExprForVarObj( decl.varObj ); exp != nil {
            valSpec.Values = []ast.Expr{ exp }
        }
        if typ := pg.valTypeForVarObj( decl.varObj ); typ != nil {
            valSpec.Type = typ
        }
        pg.addGenDecl( token.VAR, valSpec )
    }
}

func ( pg *pkgGen ) defaultFactoryIdVarDecls() []sequencedVarDecl {
    res := make( []sequencedVarDecl, 0, pg.defaultFactoryIdVars.Len() )
    pg.defaultFactoryIdVars.EachPair( 
        func( qn *mg.QualifiedTypeName, val interface{} ) {
            res = append( res, sequencedVarDecl{ val.( sequencedId ), qn } )
        },
    )
    return res
}

func ( pg *pkgGen ) assembleDefaultFactoryIdVarDecls() {
    decls := pg.defaultFactoryIdVarDecls()
    for i, decl := range decls {
        typ := varDeclType{ pg.pkgSel( goPkgReactor, goUcBuilderFactory ) }
        decls[ i ] = sequencedVarDecl{ decl.id, typ }
    }
    pg.addVarDecls( decls )
}

func ( pg *pkgGen ) assembleVarDecls() {
    idVars := make( []sequencedVarDecl, 0, pg.idVars.Len() )
    pg.idVars.EachPair( func( id *mg.Identifier, val interface{} ) {
        idVars = append( idVars, sequencedVarDecl{ val.( sequencedId ), id } )
    })
    pg.addVarDecls( idVars )
    qnVars := make( []sequencedVarDecl, 0, pg.qnVars.Len() )
    pg.qnVars.EachPair( func( qn *mg.QualifiedTypeName, val interface{} ) {
        qnVars = append( qnVars, sequencedVarDecl{ val.( sequencedId ), qn } )
    })
    pg.addVarDecls( qnVars )
    atVars := make( []sequencedVarDecl, 0, pg.atomicTypeVars.Len() )
    pg.atomicTypeVars.EachPair( 
        func( qn *mg.QualifiedTypeName, val interface{} ) {
            seqId, atDecl := val.( sequencedId ), atomicVarDecl{ qn }
            atVars = append( atVars, sequencedVarDecl{ seqId, atDecl } )
        },
    )
    pg.addVarDecls( atVars )
    pg.assembleDefaultFactoryIdVarDecls()
}

func ( pg *pkgGen ) addDefaultBuilderFactoryAssignments( bb *bodyBuilder ) {
    decls := pg.defaultFactoryIdVarDecls()
    sort.Sort( sequencedVarDeclSort( decls ) )
    for _, decl := range decls {
        qn := decl.varObj.( *mg.QualifiedTypeName )
        qnVar := pg.pkgQnVar( qn )
        bb.addAssign1( 
            decl.id.goId(), 
            callExpr(
                pg.pkgSel( goPkgCodegen, goUcMustBuilderFactoryForQname ),
                qnVar,
                goLcReg,
            ),
        )
    }
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
    pg.addDefaultBuilderFactoryAssignments( fdb.bb() )
    pg.decls = append( pg.decls, fdb.build() )
}

func ( pg *pkgGen ) assembleDecls() {
    pg.file.Decls = make( []ast.Decl, 0, 16 )
    pg.assembleImportSpecs()
    pg.assembleVarDecls()
    pg.addInitFunc()
    pg.file.Decls = append( pg.file.Decls, pg.decls... )
}

func ( pg *pkgGen ) generatePackage() error {
    pg.collectImports()
    pg.collectTypeKeys()
    if err := pg.addBuilderFactoryFuncs(); err != nil { return err }
    for _, def := range pg.defs { 
        if err := pg.generateDef( def ); err != nil { return err }
    }
    pg.assembleDecls()
    return nil
}

func ( g *Generator ) generatePackages() error {
    return g.eachPkgGen( func( pg *pkgGen ) error { 
        return pg.generatePackage()
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
//    if err := g.dumpGeneratedSource( pg ); err != nil { return err }
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
    if g.DestDir == "" { return libError( "no dest dir for generator" ) }
    g.setPackageMap()
    g.initPackages()
    if err := g.generatePackages(); err != nil { return err }
    if err := g.writeOutputs(); err != nil { return err }
    return nil
}
