package compiler

import (
    "fmt"
    "log"
    "strings"
    "bytes"
    "sort"
    "container/list"
    "bitgirder/objpath"
    "mingle/parser/tree"
    "mingle/parser"
    "mingle/code"
    mgRct "mingle/reactor"
    interp "mingle/interpreter"
    "mingle/types"
    mg "mingle"
)

type implError struct { msg string }
func ( e *implError ) Error() string { return e.msg }

func implErrorf( tmpl string, argv ...interface{} ) *implError {
    return &implError{ fmt.Sprintf( "compiler: " + tmpl, argv... ) }
}

func formatKeyedDef( id *mg.Identifier ) string {
    return "@" + id.Format( mg.LcCamelCapped )
}

func qnameIn( typ mg.TypeReference ) *mg.QualifiedTypeName {
    return mg.TypeNameIn( typ )
}

func baseTypeIsNull( typ mg.TypeReference ) bool {
    return qnameIn( typ ).Equals( mg.QnameNull )
}

func baseTypeIsNum( typ mg.TypeReference ) bool {
    qn := qnameIn( typ )
    return qn.Equals( mg.QnameInt32 ) ||
           qn.Equals( mg.QnameInt64 ) ||
           qn.Equals( mg.QnameUint32 ) ||
           qn.Equals( mg.QnameUint64 ) ||
           qn.Equals( mg.QnameFloat32 ) ||
           qn.Equals( mg.QnameFloat64 )
}

func isAtomic( typ mg.TypeReference ) bool {
    _, res := typ.( *mg.AtomicTypeReference )
    return res
}

func asUnrestrictedType( typ mg.TypeReference ) *mg.AtomicTypeReference {
    switch v := typ.( type ) {
    case *mg.AtomicTypeReference:
        if v.Restriction == nil { return v }
        return &mg.AtomicTypeReference{ Name: v.Name }
    case *mg.ListTypeReference: return asUnrestrictedType( v.ElementType )
    case *mg.NullableTypeReference: return asUnrestrictedType( v.Type )
    }
    panic( implErrorf( "Unhandled type reference: %T", typ ) )
}

var (
    typeValList = &mg.ListTypeReference{ mg.TypeValue, true }

    idAuthentication = mg.NewIdentifierUnsafe( []string{ "authentication" } )

    objpathConstExp = 
        objpath.RootedAt( mg.NewIdentifierUnsafe( []string{ "const-val" } ) )
)

type CompilationResult struct {
    BuiltTypes *types.DefinitionMap
    Errors []*Error
}

type Compilation struct {
    sources []*tree.NsUnit
    typeDecls *mg.QnameMap
    extTypes *types.DefinitionMap
    scopesByNs *mg.NamespaceMap
    onDefaults []func()
    builtTypes *types.DefinitionMap
    errs []*Error

    // Can be temporarily set in rare circumstances where an operation
    // which may generate compile errors needs to be invoked for the purposes of
    // its result only, but without any errors being recorded
    ignoreErrors bool
}

func NewCompilation() *Compilation {
    return &Compilation{ 
        sources: []*tree.NsUnit{},
        typeDecls: mg.NewQnameMap(),
        scopesByNs: mg.NewNamespaceMap(),
        onDefaults: []func(){},
        builtTypes: types.NewDefinitionMap(),
        errs: make( []*Error, 0, 4 ),
    }
}

func ( c *Compilation ) logf( tmpl string, args ...interface{} ) {
    log.Printf( tmpl, args... )
}

func ( c *Compilation ) validate() error {
    if c.extTypes == nil { c.extTypes = types.NewDefinitionMap() }
    return nil
}

func ( c *Compilation ) AddSource( u *tree.NsUnit ) *Compilation {
    c.sources = append( c.sources, u )
    return c
}

func ( c *Compilation ) awaitDefaults( f func() ) {
    c.onDefaults = append( c.onDefaults, f )
}

func ( c *Compilation ) SetExternalTypes( 
    extTypes *types.DefinitionMap ) *Compilation {
    c.extTypes = extTypes
    return c
}

func ( c *Compilation ) buildScopeForNs( ns *mg.Namespace ) *buildScope {
    if bs := c.scopesByNs.Get( ns ); bs != nil { return bs.( *buildScope ) }
    panic( implErrorf( "No build scope for %s", ns ) )
}

func ( c *Compilation ) typeDeclsGet( 
    qn *mg.QualifiedTypeName ) ( tree.TypeDecl, bool ) {
    if td := c.typeDecls.Get( qn ); td != nil {
        return td.( tree.TypeDecl ), true
    }
    return nil, false
}

func ( c *Compilation ) typeDefForQn( 
    qn *mg.QualifiedTypeName ) types.Definition {

    if def := c.builtTypes.Get( qn ); def != nil { return def }
    if def := c.extTypes.Get( qn ); def != nil { return def }
    return nil
}

func ( c *Compilation ) typeDefForType(
    typ mg.TypeReference ) types.Definition {
    return c.typeDefForQn( qnameIn( typ ) )
}

type Error struct {
    Location *parser.Location
    Message string
}

func ( e *Error ) Error() string { 
    return fmt.Sprintf( "%s: %s", e.Location, e.Message )
}

func locationFor( locVal interface{} ) *parser.Location {
    if locVal == nil { return nil }
    switch v := locVal.( type ) {
    case *parser.Location: return v
    case tree.Locatable: return v.Locate()
    }
    panic( implErrorf( "Can't get location for value of type: %T", locVal ) )
}

func ( c *Compilation ) addError( locVal interface{}, msg string ) {
    err := &Error{ locationFor( locVal ), msg }
    if ! c.ignoreErrors { c.errs = append( c.errs, err ) }
}

func ( c *Compilation ) addErrorf( 
    locVal interface{}, tmpl string, argv ...interface{} ) {
    c.addError( locVal, fmt.Sprintf( tmpl, argv... ) )
}

func ( c *Compilation ) addAssignError(
    locVal interface{},
    expctType mg.TypeReference,
    actType mg.TypeReference ) {
    c.addErrorf( locVal, "Can't assign value of type %s to %s",
        expctType, actType )
}

func ( c *Compilation ) putBuiltType( d types.Definition ) {
    c.builtTypes.MustAdd( d )
}

func locStr( elt tree.Locatable ) string { return elt.Locate().String() }

func ( c *Compilation ) touchDecl( td tree.TypeDecl, ns *mg.Namespace ) bool {
    qn := td.GetName().ResolveIn( ns )
    if prev, ok := c.typeDeclsGet( qn ); ok {
        c.addErrorf( td, "Type %s is already declared in %s", 
            td.GetName(), locStr( prev ) )
        return false
    } 
    if c.extTypes.HasKey( qn ) {
        c.addErrorf( td, 
            "Type %s conflicts with an externally loaded type", td.GetName() )
        return false
    }
    c.typeDecls.Put( qn, td )
    return true
}

type importNsMap map[ string ]*mg.Namespace

type buildScope struct {
    c *Compilation
    nsUnit *tree.NsUnit
    importResolves importNsMap
}

func ( bs *buildScope ) namespace() *mg.Namespace {
    return bs.nsUnit.NsDecl.Namespace
}

type typeResolution struct {
    errLoc *parser.Location
    aliasChain []*mg.QualifiedTypeName
}

func newTypeResolution( errLoc *parser.Location ) *typeResolution {
    return &typeResolution{ errLoc, []*mg.QualifiedTypeName{} }
}

func ( bs *buildScope ) validateQname(
    qn *mg.QualifiedTypeName, nmLoc *parser.Location ) *mg.QualifiedTypeName {
    if bs.c.typeDecls.HasKey( qn ) { return qn }
    if bs.c.extTypes.HasKey( qn ) { return qn }
    bs.c.addErrorf( nmLoc, "Unresolved type: %s", qn )
    return nil
}

func ( bs *buildScope ) resolveImport( 
    nm *mg.DeclaredTypeName ) *mg.QualifiedTypeName {
    if ns := bs.importResolves[ nm.ExternalForm() ]; ns != nil {
        return nm.ResolveIn( ns )
    }
    return nil
}

func ( bs *buildScope ) expandDeclaredTypeName( 
    nm *mg.DeclaredTypeName, nmLoc *parser.Location ) *mg.QualifiedTypeName {
    qn := nm.ResolveIn( bs.namespace() )
    // order of these is important since final test reassigns to qn
    if bs.c.typeDecls.HasKey( qn ) { return qn }
    if bs.c.extTypes.HasKey( qn ) { return qn }
    if qn = bs.resolveImport( nm ); qn != nil { return qn }
    if qn = nm.ResolveIn( mg.CoreNsV1 ); bs.c.extTypes.HasKey( qn ) { 
        return qn 
    }
    bs.c.addErrorf( nmLoc, "Unresolved type: %s", nm )
    return nil
}

func ( bs *buildScope ) qnameFor(
    nm mg.TypeName, nmLoc *parser.Location ) *mg.QualifiedTypeName {

    switch v := nm.( type ) {
    case *mg.QualifiedTypeName: return bs.validateQname( v, nmLoc )
    case *mg.DeclaredTypeName: return bs.expandDeclaredTypeName( v, nmLoc )
    }
    panic( implErrorf( "unhandled type name: %T", nm ) )
}

// returns nil if nm can't be succesfully resolved or validated, emitting errors
// as with bs.qnameFor(). If a name is resolved and is not the name of an alias
// definition, it is returned. If the resolved name is the name of an alias
// definition that resolves to an atomic type with no restriction, that atomic
// type's qname is returned. Otherwise, nil is returned but with no errors
// emitted. If there develops the need to distinguish between no-name and
// name-but-non-trivial-alias, we can change this call's return type to carry
// that information.
func ( bs *buildScope ) qnameForMixin(
    nm mg.TypeName, nmLoc *parser.Location ) *mg.QualifiedTypeName {

    qn := bs.qnameFor( nm, nmLoc )
    if qn == nil { return nil }
    if def := bs.c.typeDefForQn( qn ); def != nil {
        if ad, ok := def.( *types.AliasedTypeDefinition ); ok {
            if at, ok := ad.AliasedType.( *mg.AtomicTypeReference ); ok {
                if at.Restriction == nil { qn = at.Name }
            }
        }
    }
    return qn
}

func ( bs *buildScope ) addRestrictionTargetTypeError(
    qn *mg.QualifiedTypeName, 
    rx parser.RestrictionSyntax, 
    errLoc *parser.Location ) {

    rxNm := ""
    switch rx.( type ) {
    case *parser.RangeRestrictionSyntax: rxNm = "range"
    case *parser.RegexRestrictionSyntax: rxNm = "regex"
    default: panic( libErrorf( "unhandled restriction: %T", rx ) )
    }
    bs.c.addErrorf( errLoc, "Invalid target type for %s restriction: %s", 
        rxNm, qn )
}

func ( bs *buildScope ) resolveRegexRestriction( 
    qn *mg.QualifiedTypeName, 
    rx *parser.RegexRestrictionSyntax,
    errLoc *parser.Location ) mg.ValueRestriction {

    if qn.Equals( mg.QnameString ) { 
        if rr, err := mg.NewRegexRestriction( rx.Pat ); err == nil { 
            return rr 
        } else {
            bs.c.addError( rx.Loc, err.Error() )
            return nil
        }
    }
    bs.addRestrictionTargetTypeError( qn, rx, errLoc )
    return nil
}

func ( bs *buildScope ) parseTimestamp( 
    str string, errLoc *parser.Location ) ( mg.Timestamp, bool ) {
    
    tm, err := parser.ParseTimestamp( str )
    if err == nil { return tm, true }
    if pe, ok := err.( *parser.ParseError ); ok {
        bs.c.addError( errLoc, pe.Message )
        return tm, false
    }
    bs.c.addError( errLoc, err.Error() )
    return tm, false
}

var rangeValTypeNames []*mg.QualifiedTypeName

func init() {
    rangeValTypeNames = []*mg.QualifiedTypeName{
        mg.QnameString,
        mg.QnameInt32,
        mg.QnameInt64,
        mg.QnameUint32,
        mg.QnameUint64,
        mg.QnameFloat32,
        mg.QnameFloat64,
        mg.QnameTimestamp,
    }
}

func ( bs *buildScope ) setRangeNumValue(
    valPtr *mg.Value,
    rx *parser.NumRestrictionSyntax,
    qn *mg.QualifiedTypeName,
    errLoc *parser.Location,
    bound string ) int {

    if ! mg.IsNumericTypeName( qn ) {
        bs.c.addErrorf( rx.Loc, "Got number as %s value for range", bound )
        return 1
    }
    if mg.IsIntegerTypeName( qn ) && ( ! rx.Num.IsInt() ) {
        bs.c.addErrorf( rx.Loc, "Got decimal as %s value for range", bound )
        return 1
    }
    num, err := mg.ParseNumber( rx.LiteralString(), qn )
    if err == nil {
        *valPtr = num
        return 0
    }
    bs.c.addError( rx.Loc, err.Error() )
    return 1
}

func ( bs *buildScope ) setRangeTimestampValue(
    valPtr *mg.Value, str string, errLoc *parser.Location ) int {
        
    if tm, ok := bs.parseTimestamp( str, errLoc ); ok {
        *valPtr = tm
        return 0
    }
    return 1
}

func ( bs *buildScope ) setRangeStringValue(
    valPtr *mg.Value,
    rx *parser.StringRestrictionSyntax,
    qn *mg.QualifiedTypeName,
    errLoc *parser.Location,
    bound string ) int {

    switch {
    case qn.Equals( mg.QnameString ):
        *valPtr = mg.String( rx.Str )
        return 0
    case qn.Equals( mg.QnameTimestamp ):
        return bs.setRangeTimestampValue( valPtr, rx.Str, errLoc )
    }
    bs.c.addErrorf( rx.Loc, "Got string as %s value for range", bound )
    return 1
}

// bound is which bound to report in the error: "min" or "max"
func ( bs *buildScope ) setRangeValue(
    valPtr *mg.Value, 
    rx parser.RestrictionSyntax,
    qn *mg.QualifiedTypeName, 
    errLoc *parser.Location,
    bound string ) int {

    switch v := rx.( type ) {
    case *parser.NumRestrictionSyntax: 
        return bs.setRangeNumValue( valPtr, v, qn, errLoc, bound )
    case *parser.StringRestrictionSyntax: 
        return bs.setRangeStringValue( valPtr, v, qn, errLoc, bound )
    }
    panic( libErrorf( "unhandled restriction: %T", rx ) )
}

func areAdjacentInts( min, max mg.Value ) bool {
    switch minV := min.( type ) {
    case mg.Int32: 
        return int32( max.( mg.Int32 ) ) - int32( minV ) == int32( 1 )
    case mg.Uint32: 
        return uint32( max.( mg.Uint32 ) ) - uint32( minV ) == uint32( 1 )
    case mg.Int64: 
        return int64( max.( mg.Int64 ) ) - int64( minV ) == int64( 1 )
    case mg.Uint64: 
        return uint64( max.( mg.Uint64 ) ) - uint64( minV ) == uint64( 1 )
    }
    return false
}

func ( bs *buildScope ) checkRangeBounds( 
    rr *mg.RangeRestriction, errLoc *parser.Location ) int {

    failed := false
    switch i := rr.Min.( mg.Comparer ).Compare( rr.Max ); {
    case i == 0: failed = ! ( rr.MinClosed && rr.MaxClosed )
    case i > 0: failed = true
    case i < 0: 
        open := ! ( rr.MinClosed || rr.MaxClosed )
        failed = open && areAdjacentInts( rr.Min, rr.Max )
    }
    if failed { 
        bs.c.addError( errLoc, "Unsatisfiable range" )
        return 1
    }
    return 0
}

func ( bs *buildScope ) setRangeValues(
    rr *mg.RangeRestriction, 
    rx *parser.RangeRestrictionSyntax, 
    qn *mg.QualifiedTypeName,
    errLoc *parser.Location ) bool {

    fails := 0
    if rx.Left != nil {
        fails += bs.setRangeValue( &( rr.Min ), rx.Left, qn, errLoc, "min" )
    }
    if rx.Right != nil {
        fails += bs.setRangeValue( &( rr.Max ), rx.Right, qn, errLoc, "max" ) 
    }
    if ! ( rr.Min == nil || rr.Max == nil ) { 
        fails += bs.checkRangeBounds( rr, rx.Loc ) 
    }
    return fails == 0
}

func ( bs *buildScope ) resolveRangeRestriction(
    qn *mg.QualifiedTypeName,
    rx *parser.RangeRestrictionSyntax,
    errLoc *parser.Location ) mg.ValueRestriction {

    rr := &mg.RangeRestriction{ 
        MinClosed: rx.LeftClosed, 
        MaxClosed: rx.RightClosed,
    }
    for _, rvTypNm := range rangeValTypeNames {
        if qn.Equals( rvTypNm ) {
            if bs.setRangeValues( rr, rx, rvTypNm, errLoc ) { return rr }
            return nil
        }
    }
    bs.addRestrictionTargetTypeError( qn, rx, errLoc )
    return nil
}

func ( bs *buildScope ) resolveRestriction(
    qn *mg.QualifiedTypeName, 
    rx parser.RestrictionSyntax,
    errLoc *parser.Location ) mg.ValueRestriction {

    switch v := rx.( type ) {
    case *parser.RegexRestrictionSyntax: 
        return bs.resolveRegexRestriction( qn, v, errLoc )
    case *parser.RangeRestrictionSyntax: 
        return bs.resolveRangeRestriction( qn, v, errLoc )
    }
    panic( libErrorf( "unhandled restriction: %T", rx ) )
}

func ( bs *buildScope ) getAtomicTypeReference( 
    qn *mg.QualifiedTypeName, 
    rx parser.RestrictionSyntax,
    tr *typeResolution ) *mg.AtomicTypeReference {

    res := &mg.AtomicTypeReference{ Name: qn }
    if rx == nil { return res }
    res.Restriction = bs.resolveRestriction( qn, rx, tr.errLoc )
    if res.Restriction == nil { return nil }
    return res
}

func ( bs *buildScope ) unalias( 
    aliasVal interface{}, 
    aliasQn *mg.QualifiedTypeName, 
    tr *typeResolution ) mg.TypeReference {

    switch v := aliasVal.( type ) {
    case *tree.AliasDecl:
        return bs.c.buildScopeForNs( aliasQn.Namespace ).resolve( v.Target, tr )
    case *types.AliasedTypeDefinition: return v.AliasedType
    }
    panic( implErrorf( "Unhandled alias: %T", aliasVal ) )
}

func ( bs *buildScope ) addCircularAliasError(
    tr *typeResolution, qn *mg.QualifiedTypeName ) {
    buf := bytes.Buffer{}
    buf.WriteString( "Alias cycle: " )
    for _, elt := range tr.aliasChain {
        buf.WriteString( elt.ExternalForm() )
        buf.WriteString( " --> " )
    }
    buf.WriteString( qn.ExternalForm() )
    bs.c.addError( tr.errLoc, buf.String() )
}

func ( bs *buildScope ) addAlias( 
    qn *mg.QualifiedTypeName, tr *typeResolution ) bool {
    for _, elt := range tr.aliasChain {
        if elt.Equals( qn ) {
            bs.addCircularAliasError( tr, qn )
            return false
        }
    }
    tr.aliasChain = append( tr.aliasChain, qn )
    return true
}

func ( bs *buildScope ) aliasValFor(
    qn *mg.QualifiedTypeName, tr *typeResolution ) ( interface{}, bool ) {
    var alias interface{}
    if decl, ok := bs.c.typeDeclsGet( qn ); ok {
        if _, ok := decl.( *tree.AliasDecl ); ok { alias = decl }
    } else if td := bs.c.extTypes.Get( qn ); td != nil {
        if _, ok := td.( *types.AliasedTypeDefinition ); ok { alias = td }
    }
    if alias == nil { return nil, true }
    return alias, bs.addAlias( qn, tr )
}

type typeCompletion struct {
    bs *buildScope
    tr *typeResolution
}

func ( tc typeCompletion ) CompleteBaseType(
    nm mg.TypeName,
    rx parser.RestrictionSyntax,
    l *parser.Location ) ( mg.TypeReference, bool, error ) {

    qn := tc.bs.qnameFor( nm, tc.tr.errLoc )
    if qn == nil { return nil, false, nil }
    var res mg.TypeReference
    aliasVal, aliasOk := tc.bs.aliasValFor( qn, tc.tr )
    if aliasOk {
        if aliasVal == nil { 
            if at := tc.bs.getAtomicTypeReference( qn, rx, tc.tr ); at != nil {
                res = at
            }
        } else { res = tc.bs.unalias( aliasVal, qn, tc.tr ) }
    }
    if res == nil { return nil, false, nil }
    return res, true, nil
}

func ( bs *buildScope ) completeType( 
    typ *parser.CompletableTypeReference,
    tr *typeResolution ) mg.TypeReference {

    res, err := typ.CompleteType( typeCompletion{ tr: tr, bs: bs } )
    if err == nil { return res }
    bs.c.addError( typ.Location(), err.Error() )
    return nil
}

// This method may return non-nil even if some errors were encountered, to allow
// further processing to continue
func ( bs *buildScope ) resolve( 
    typ *parser.CompletableTypeReference,
    tr *typeResolution ) mg.TypeReference {
    res := bs.completeType( typ, tr )
    if res != nil && baseTypeIsNull( res ) && ( ! isAtomic( res ) ) {
        bs.c.addError( tr.errLoc, "Non-atomic use of Null type" )
    }
    return res 
}

func ( bs *buildScope ) resolveType( 
    typ *parser.CompletableTypeReference,
    errLoc *parser.Location ) mg.TypeReference {

    tr := newTypeResolution( errLoc )
    return bs.resolve( typ, tr )
} 

type buildContext struct {
    td tree.TypeDecl
    scope *buildScope
}

func ( bc buildContext ) qname() *mg.QualifiedTypeName {
    return bc.td.GetName().ResolveIn( bc.scope.namespace() )
}

type typeInformed interface { GetTypeInfo() *tree.TypeDeclInfo }

func ( bc buildContext ) typeInfo() *tree.TypeDeclInfo {
    if ti, ok := bc.td.( typeInformed ); ok { return ti.GetTypeInfo() }
    return nil
} 

func ( bc buildContext ) mustTypeInfo() *tree.TypeDeclInfo {
    if ti := bc.typeInfo(); ti != nil { return ti }
    panic( implErrorf( "no type info present for %s", bc.td.GetName() ) )
}

func ( c *Compilation ) isValidImport( qn *mg.QualifiedTypeName ) bool {
    return c.extTypes.HasKey( qn ) || c.typeDecls.HasKey( qn )
}

// Returns the initial set of non-qualified DeclaredTypeNames in play when
// processing imports for an nsUnit having srcNs
func ( c *Compilation ) initImportWorkingSet( 
    srcNs *mg.Namespace ) map[ string ]interface{} {
    res := map[ string ]interface{}{}
    c.typeDecls.EachPair( func( qn *mg.QualifiedTypeName, td interface{} ) {
        if qn.Namespace.Equals( srcNs ) { res[ qn.Name.ExternalForm() ] = td }
    })
    return res
}

func ( c *Compilation ) addImportByName(
    srcNs *mg.Namespace,
    toAdd *mg.DeclaredTypeName, 
    inNs *mg.Namespace,
    work map[ string ]interface{}, 
    res importNsMap,
    errLoc *parser.Location ) {
    k := toAdd.ExternalForm()
    var prev interface{}
    var ok bool
    prev, ok = work[ k ]
    if ! ok { prev, ok = res[ k ] }
    if ok {
        prefix, suffix := "Importing %s from %s would conflict with ", ""
        switch prev.( type ) {
        case *mg.Namespace: suffix = "previous import from %s"
        case tree.TypeDecl: suffix, prev = "declared type in %s", srcNs
        default: panic( implErrorf( "Unhandled prev val: %T", prev ) )
        }
        c.addErrorf( errLoc, prefix + suffix, toAdd, inNs, prev )
    } else { work[ k ] = inNs }
}

func importExcludes( imprt *tree.Import, qn *mg.QualifiedTypeName ) bool {
    for _, e := range imprt.Excludes {
        if e.Name.Equals( qn.Name ) { return true } 
    }
    return false
}

func ( c *Compilation ) addInitialGlobNames(
    srcNs *mg.Namespace,
    work map[ string ]interface{}, 
    res importNsMap, 
    imprt *tree.Import ) {
    errLoc := imprt.NamespaceLoc
    c.extTypes.EachDefinition(
        func ( td types.Definition ) {
            if qn := td.GetName(); qn.Namespace.Equals( imprt.Namespace ) {
                if ! importExcludes( imprt, qn ) {
                    c.addImportByName( 
                        srcNs, qn.Name, qn.Namespace, work, res, errLoc )
                }
            }
        },
    )
    c.typeDecls.EachPair(
        func( qn *mg.QualifiedTypeName, _ interface{} ) {
            if ns := qn.Namespace; ns.Equals( imprt.Namespace ) {
                if ! importExcludes( imprt, qn ) {
                    c.addImportByName( srcNs, qn.Name, ns, work, res, errLoc )
                }
            }
        },
    )
}

func ( c *Compilation ) checkImportTypes( 
    ns *mg.Namespace, typs []*tree.TypeListEntry ) []*mg.DeclaredTypeName {
    res := make( []*mg.DeclaredTypeName, 0, len( typs ) )
    for _, e := range typs {
        qn := e.Name.ResolveIn( ns ) 
        if c.isValidImport( qn ) {
            res = append( res, e.Name )
        } else { 
            c.addErrorf( e.Loc, "No such import in %s: %s", ns, e.Name )
        }
    }
    return res
}

func ( c *Compilation ) addImportsFrom(
    srcNs *mg.Namespace, imprt *tree.Import, m map[ string ]*mg.Namespace ) {
    ns := imprt.Namespace
    work := c.initImportWorkingSet( srcNs )
    if imprt.IsGlob { 
        c.addInitialGlobNames( srcNs, work, m, imprt ) 
    } else {
        for _, nm := range c.checkImportTypes( ns, imprt.Includes ) {
            c.addImportByName( srcNs, nm, ns, work, m, imprt.Locate() )
        }
    }
    for _, nm := range c.checkImportTypes( ns, imprt.Excludes ) {
        delete( work, nm.ExternalForm() )
    }
    for k, v := range work { 
        if ns, ok := v.( *mg.Namespace ); ok { m[ k ] = ns }
    }
}

func ( c *Compilation ) getImportResolves( u *tree.NsUnit ) importNsMap {
    res := make( map[ string ]*mg.Namespace )
    srcNs := u.NsDecl.Namespace
    for _, imprt := range u.Imports { c.addImportsFrom( srcNs, imprt, res ) }
    return res
}

func ( c *Compilation ) addBuildableContexts(
    work *list.List, seen *mg.QnameMap, ctxs []buildContext ) []buildContext {
    for e := work.Front(); e != nil; {
        bc := e.Value.( buildContext )
        qn := bc.qname()
        ctxs = append( ctxs, bc )
        seen.Put( qn, bc )
        // Due to the way List.Remove() works, we first advance e, and then
        // remove the element we just processed
        toRemove := e
        e = e.Next()
        work.Remove( toRemove )
    }
    return ctxs
}

func ( c *Compilation ) addCircularDepError( circ *list.List ) {
    for e := circ.Front(); e != nil; e = e.Next() {
        bc := e.Value.( buildContext )
        c.addErrorf( bc.td, 
            "Type %s is involved in one or more circular dependencies", 
            bc.qname() )
    }
}

func ( c *Compilation ) sortByBuildOrder( ctxs []buildContext ) []buildContext {
    res := make( []buildContext, 0, len( ctxs ) )
    work := list.New()
    for _, ctx := range ctxs { work.PushBack( ctx ) }
    seen := mg.NewQnameMap()
    for work.Len() != 0 {
        lenPre := work.Len()
        res = c.addBuildableContexts( work, seen, res )
        if lenPre == work.Len() { break }
    }
    if work.Len() != 0 { c.addCircularDepError( work ) }
    return res
}

// Init process of inner loop (ordering is important):
//
//  1: touch all declared decls, accumulating all that should be processed by
//  the compile (bcOk). This has the side effect that c.typeDecls will contain
//  all valid entries for the declaring source unit
//
//  2: build the build scope for the source unit. This requires the side effect
//  from step 1 for correct import processing, both to find valid imports and to
//  correctly fail for unresolved import targets
//
//  3: finally create the actual buildContexts for each entry in bcOk once
//  buildScope is ready
//
func ( c *Compilation ) initBuildContexts() []buildContext {
    res := make( []buildContext, 0, 16 )
    for _, src := range c.sources {
        ns := src.NsDecl.Namespace
        bcOk := make( []tree.TypeDecl, 0, len( src.TypeDecls ) )
        for _, td := range src.TypeDecls {
            if c.touchDecl( td, ns ) { bcOk = append( bcOk, td ) }
        }
        resolvs := c.getImportResolves( src )
        bs := &buildScope{ c: c, nsUnit: src, importResolves: resolvs }
        c.scopesByNs.Put( ns, bs )
        for _, td := range bcOk {
            res = append( res, buildContext{ scope: bs, td: td } )
        }
    }
    return c.sortByBuildOrder( res )
}

func ( c *Compilation ) printBuildOrder( ctxs []buildContext ) {
    strs := make( []string, len( ctxs ) )
    for i, ctx := range ctxs { strs[ i ] = ctx.td.GetName().ExternalForm() }
}

type qnameSort []buildContext

func ( s qnameSort ) Len() int { return len( s ) }

func ( s qnameSort ) Less( i, j int ) bool {
    return s[ i ].qname().ExternalForm() < s[ j ].qname().ExternalForm()
}

func ( s qnameSort ) Swap( i, j int ) { s[ i ], s[ j ] = s[ j ], s[ i ] }

func ( c *Compilation ) getAliasBuildOrder(
    ctxs []buildContext ) []buildContext {

    res := make( []buildContext, 0, 4 )
    for _, bc := range ctxs {
        if _, ok := bc.td.( *tree.AliasDecl ); ok { res = append( res, bc ) }
    }
    sort.Sort( qnameSort( res ) )
    return res
}

// We manually complete and seed the TypeCompletionContext with the type being
// aliased to ensure that error messages on circular alias chains begin with the
// type we're processing
func ( c *Compilation ) buildAliasedType( bc buildContext ) {
    ad, bs := bc.td.( *tree.AliasDecl ), bc.scope
    qn := bc.qname()
    tr := newTypeResolution( ad.Target.Location() )
    if ! bs.addAlias( qn, tr ) {
        panic( implErrorf( "Failed to add initial alias to chain: %s", qn ) )
    }
    if typ := bs.resolve( ad.Target, tr ); typ != nil {
        def := &types.AliasedTypeDefinition{}
        def.AliasedType = typ
        def.Name = qn
        c.putBuiltType( def )
    }
}

func ( c *Compilation ) buildAliasedTypes( ctxs []buildContext ) {
    for _, bc := range c.getAliasBuildOrder( ctxs ) { c.buildAliasedType( bc ) }
}

func ( c *Compilation ) buildFieldDefinition(
    fldDecl *tree.FieldDecl, bs *buildScope ) *types.FieldDefinition {

    res := &types.FieldDefinition{ Name: fldDecl.Name }
    res.Type = bs.resolveType( fldDecl.Type, fldDecl.Type.Location() )
    if res.Type == nil { return nil }
    if baseTypeIsNull( res.Type ) {
        c.addError( fldDecl.Type.Location(), "Null type not allowed here" )
        return nil
    }
    return res
}

type builtField struct {
    def *types.FieldDefinition
    src interface{}
}

func ( bd builtField ) name() *mg.Identifier { return bd.def.Name }

type fieldSetBuilder struct {
    c *Compilation
    bc buildContext
    flds []*tree.FieldDecl
    schemas []*tree.SchemaMixinDecl
    fs *types.FieldSet
    work *mg.IdentifierMap
}

func ( fsb *fieldSetBuilder ) addDefinition( 
    fd *types.FieldDefinition, src interface{} ) {

    var flds []builtField
    key := fd.Name
    if fldsVal, ok := fsb.work.GetOk( key ); ok {
        flds = fldsVal.( []builtField )
    } else { flds = make( []builtField, 0, 4 ) }
    bd := builtField{ def: fd, src: src }
    flds = append( flds, bd )
    fsb.work.Put( key, flds )
}

func ( fsb *fieldSetBuilder ) addFieldsFromSchema( 
    sd *tree.SchemaMixinDecl ) int {
    
    qn := fsb.bc.scope.qnameForMixin( sd.Name, sd.NameLoc )
    if qn == nil { return 1 }
    switch v := fsb.c.typeDefForQn( qn ).( type ) {
    case *types.SchemaDefinition:
        v.Fields.EachDefinition( func( fd *types.FieldDefinition ) {
            fsb.addDefinition( fd, v ) 
        })
    default:
        fsb.c.addErrorf( sd, "not a mixin: %s", qn )
        return 1
    }
    return 0
}

func ( fsb *fieldSetBuilder ) addSchemaMixins() int { 
    if fsb.schemas == nil { return 0 }
    errs := 0
    for _, schema := range fsb.schemas {
        errs += fsb.addFieldsFromSchema( schema )
    }
    return errs
}

func ( fsb *fieldSetBuilder ) addDirectFieldDecls() int {
    errs := 0
    for _, fldDecl := range fsb.flds {
        fd := fsb.c.buildFieldDefinition( fldDecl, fsb.bc.scope )
        if fd == nil { errs++ } else { fsb.addDefinition( fd, fldDecl ) }
    }
    return errs
}

func ( fsb *fieldSetBuilder ) addBuiltFields() bool {
    errs := 0
    fsb.work.EachPair( func( fld *mg.Identifier, val interface{} ) {
        flds := val.( []builtField )
        if len( flds ) == 1 {
            fsb.fs.MustAdd( flds[ 0 ].def )
        } else {
            for _, bf := range flds {
                errs++
                fsb.c.addErrorf( fsb.bc.td, "field '%s' redefined", bf.name() )
            }
        }
    })
    return errs == 0
} 

func ( fsb *fieldSetBuilder ) build() bool {
    
    errs := fsb.addDirectFieldDecls()
    errs += fsb.addSchemaMixins()
    if errs == 0 { return fsb.addBuiltFields() }
    return false
}

func ( c *Compilation ) buildFieldSet( 
    bc buildContext, 
    flds []*tree.FieldDecl,
    schemas []*tree.SchemaMixinDecl,
    fs *types.FieldSet ) bool {

    fsb := &fieldSetBuilder{
        c: c,
        bc: bc,
        flds: flds,
        schemas: schemas,
        fs: fs,
        work: mg.NewIdentifierMap(),
    }
    return fsb.build()
}

func ( c *Compilation ) processConstructor(
    consDecl *tree.ConstructorDecl,
    seen map[ string ]bool,
    bs *buildScope ) *types.ConstructorDefinition {
    typ := bs.resolveType( consDecl.ArgType, consDecl.ArgType.Location() )
    if typ == nil { return nil }
    keyStr := typ.ExternalForm()
    if _, hadPrev := seen[ keyStr ]; hadPrev {
        c.addErrorf( consDecl, 
            "Duplicate constructor signature for type %s", typ )
        return nil
    } 
    seen[ keyStr ] = true
    return &types.ConstructorDefinition{ typ }
}

func ( c *Compilation ) processConstructors(
    consDecls []*tree.ConstructorDecl,
    sd *types.StructDefinition,
    bs *buildScope ) bool {
    ok := true
    seen := make( map[ string ]bool )
    for _, consDecl := range consDecls {
        if consDef := c.processConstructor( consDecl, seen, bs ); 
           consDef == nil {
            ok = false
        } else { sd.Constructors = append( sd.Constructors, consDef ) }
    }
    return ok
}

func ( c *Compilation ) buildConstructors(
    bc buildContext, sd *types.StructDefinition ) bool {

    arr := bc.td.( *tree.StructDecl ).Constructors
    if len( arr ) == 0 { return true }
    return c.processConstructors( arr, sd, bc.scope )
}

func ( c *Compilation ) buildStructType( bc buildContext ) {
    sd := types.NewStructDefinition()
    sd.Name = bc.qname() 
    decl := bc.td.( *tree.StructDecl )
    // always evaluate lhs even if ok is already false, so we generate possibly
    // more compiler errors in each run
    ok := true
    ok = c.buildFieldSet( bc, decl.Fields, decl.Schemas, sd.Fields ) && ok
    ok = c.buildConstructors( bc, sd ) && ok
    if ok { c.putBuiltType( sd ) }
}

type schemaBuildOrder struct {
    c *Compilation
    ord []buildContext
    nextIdx int
}

func ( c *Compilation ) newSchemaBuildOrder( 
    ctxs []buildContext ) *schemaBuildOrder {

    res := &schemaBuildOrder{ c: c, ord: make( []buildContext, 0, 16 ) }
    for _, bc := range ctxs { 
        if _, ok := bc.td.( *tree.SchemaDecl ); ok { 
            res.ord = append( res.ord, bc ) 
        }
    }
    return res
}

func ( bo *schemaBuildOrder ) initHoldMap() *mg.QnameMap {
    res := mg.NewQnameMap()
    for _, bc := range bo.ord { res.Put( bc.qname(), bc ) }
    return res
}

// for each mixin target we check whether the target is upstream of bc. We
// silently ignore situations in which the mixin does not resolve to a known
// name or when it resolves to a name built outside of this compilation unit,
// since either of these cases will be handled elsewhere and shouldn't block our
// ability to get a build order.
func ( bo *schemaBuildOrder ) hasDeps( 
    bc buildContext, ready *mg.QnameMap ) bool {

    sd := bc.td.( *tree.SchemaDecl )
    deps := 0
    for _, mixDecl := range sd.Schemas {
        qn := bc.scope.qnameForMixin( mixDecl.Name, mixDecl.NameLoc )
        if qn == nil { continue }
        if ! bo.c.typeDecls.HasKey( qn ) { continue }
        if ready.HasKey( qn ) { continue }
        deps++
    }
    return deps == 0
}

// as we move from hold --> ready, we keep the newly added qnames in added,
// since we can't do concurrent deletes from inside the EachPair block
func ( bo *schemaBuildOrder ) makeReady( ready, hold *mg.QnameMap ) bool {
    added := make( []*mg.QualifiedTypeName, 0, 4 )
    hold.EachPair( func( qn *mg.QualifiedTypeName, v interface{} ) {
        if bc := v.( buildContext ); bo.hasDeps( bc, ready ) {
            ready.Put( qn, bc )
            bo.ord[ bo.nextIdx ] = bc
            bo.nextIdx++
            added = append( added, qn )
        }
    })
    for _, qn := range added { hold.Delete( qn ) }
    return len( added ) > 0
}

func ( bo *schemaBuildOrder ) sort() bool {
    ready, hold := mg.NewQnameMap(), bo.initHoldMap()
    for loop := true; loop && hold.Len() > 0; {
        loop = bo.makeReady( ready, hold )
    }
    if hold.Len() == 0 { return true }
    strs := make( []string, 0, hold.Len() )
    hold.EachPair( func( qn *mg.QualifiedTypeName, _ interface{} ) {
        strs = append( strs, qn.ExternalForm() )
    })
    names := strings.Join( strs, ", " )
    tmpl := "Schemas are involved in one or more mixin cycles: %s"
    bo.c.addErrorf( nil, tmpl, names )
    return false
}

func ( bo *schemaBuildOrder ) getOrder() []buildContext {
    if ! bo.sort() { return nil }
    return bo.ord
}

func ( c *Compilation ) buildSchemaType( bc buildContext ) {
    decl := bc.td.( *tree.SchemaDecl )
    sd := types.NewSchemaDefinition()
    sd.Name = bc.qname()
    if c.buildFieldSet( bc, decl.Fields, decl.Schemas, sd.Fields ) {
        c.putBuiltType( sd )
    }
}

func ( c *Compilation ) buildSchemaTypes( ctxs []buildContext ) {
    ord := c.newSchemaBuildOrder( ctxs ).getOrder()
    if ord == nil { return }
    for _, bc := range ord { c.buildSchemaType( bc ) }
}

func ( c *Compilation ) buildEnumType( bc buildContext ) {
    decl := bc.td.( *tree.EnumDecl )
    ed := &types.EnumDefinition{ Name: bc.qname(), Values: []*mg.Identifier{} }
    ok, seen := true, mg.NewIdentifierMap()
    for _, valDecl := range decl.Values {
        if val := valDecl.Value; seen.HasKey( val ) {
            ok = false
            c.addErrorf( valDecl.ValueLoc, 
                "Duplicate definition of enum value: %s", val )
        } else {
            ed.Values = append( ed.Values, val )
            seen.Put( val, true )
        }
    }
    if ok { c.putBuiltType( ed ) }
}

func ( c *Compilation ) setCallSignatureFields(
    decl *tree.CallSignature, sig *types.CallSignature, bs *buildScope ) bool {

    fldDecls := decl.Fields
    fldDefs, errs := make( []*types.FieldDefinition, 0, len( fldDecls ) ), 0
    for _, fldDecl := range fldDecls {
        if fldDef := c.buildFieldDefinition( fldDecl, bs ); fldDef != nil {
            fldDefs = append( fldDefs, fldDef ) 
        } else { errs++ }
    }
    if errs == 0 { 
        for _, fldDef := range fldDefs { sig.Fields.MustAdd( fldDef ) } 
    }
    return errs == 0
}

func ( c *Compilation ) buildCallSignature(
    decl *tree.CallSignature, bs *buildScope ) *types.CallSignature {
    res, ok := types.NewCallSignature(), true
    ok = c.setCallSignatureFields( decl, res, bs ) && ok
    if retTyp := bs.resolveType( decl.Return, decl.Return.Location() ); 
       retTyp == nil {
        ok = false
    } else { res.Return = retTyp }
    for _, thrown := range decl.Throws {
        if thrownTyp := bs.resolveType( thrown.Type, thrown.Type.Location() );
           thrownTyp == nil {
            ok = false
        } else { res.Throws = append( res.Throws, thrownTyp ) }
    }
    if ok { return res }
    return nil
}

func ( c *Compilation ) buildPrototypeType( bc buildContext ) {
    decl := bc.td.( *tree.PrototypeDecl )
    if sig := c.buildCallSignature( decl.Sig, bc.scope ); sig != nil {
        proto := &types.PrototypeDefinition{ Name: bc.qname(), Signature: sig }
        c.putBuiltType( proto )
    }
}

func ( c *Compilation ) buildOpDef(
    opDecl *tree.OperationDecl, bs *buildScope ) *types.OperationDefinition {
    res := &types.OperationDefinition{ Name: opDecl.Name }
    if res.Signature = c.buildCallSignature( opDecl.Call, bs );
       res.Signature == nil {
        return nil
    }
    return res
}

func ( c *Compilation ) buildOpDefs( 
    opDecls []*tree.OperationDecl,
    opDefs []*types.OperationDefinition,
    bs *buildScope ) ( []*types.OperationDefinition, bool ) {
    seen, ok := mg.NewIdentifierMap(), true
    for _, opDecl := range opDecls {
        if opDef := c.buildOpDef( opDecl, bs ); opDef != nil {
            if nm := opDef.Name; seen.HasKey( nm ) {
                c.addErrorf( opDecl.NameLoc, 
                    "Operation already defined: %s", nm )
                ok = false
            } else {
                opDefs = append( opDefs, opDef )
                seen.Put( nm, true )
            }
        }
    }
    return opDefs, ok
}

func ( c *Compilation ) validateAsSecurityDef(
    proto *types.PrototypeDefinition, errLoc *parser.Location ) bool {
    nm, flds := proto.Name, proto.Signature.Fields
    authFld := flds.Get( idAuthentication )
    if authFld == nil {
        c.addErrorf( errLoc, "%s has no authentication field", nm )
        return false
    } 
    c.awaitDefaults( func() {
        if authFld.Default != nil {
            c.addErrorf( errLoc, 
                "%s supplies a default authentication value", nm )
        }
    })
    if flds.Len() > 1 {
        c.addErrorf( errLoc, "%s has one or more unrecognized fields", nm )
        return false
    }
    return true
}

func ( c *Compilation ) processSecurityDecl(
    decl *tree.SecurityDecl, bs *buildScope ) *mg.QualifiedTypeName {
    res := bs.qnameFor( decl.Name, decl.NameLoc )
    if res == nil { return nil }
    switch def := c.typeDefForQn( res ).( type ) {
    case nil: c.addErrorf( decl.NameLoc, "Unknown @security: %s", res )
    case *types.PrototypeDefinition:
        if ! c.validateAsSecurityDef( def, decl.NameLoc ) { res = nil }
    default: c.addErrorf( decl.NameLoc, "Illegal @security type: %s", res )
    }
    return res
}

func ( c *Compilation ) buildOptSecurityDecl(
    decl *tree.ServiceDecl, sd *types.ServiceDefinition, bs *buildScope ) bool {

    switch len( decl.SecurityDecls ) {
    case 0: return true
    case 1:
        qn := c.processSecurityDecl( decl.SecurityDecls[ 0 ], bs )
        if qn != nil { 
            sd.Security = qn 
            return true
        }
    default: 
        c.addError( decl, "Multiple security declarations are not supported" )
    }
    return false
}

func ( c *Compilation ) buildServiceType( bc buildContext ) {
    decl := bc.td.( *tree.ServiceDecl )
    sd, ok := types.NewServiceDefinition(), true
    sd.Name = bc.qname()
    var opDefsOk bool
    sd.Operations, opDefsOk = 
        c.buildOpDefs( decl.Operations, sd.Operations, bc.scope ) 
    ok = opDefsOk && ok
    ok = c.buildOptSecurityDecl( decl, sd, bc.scope ) && ok
    if ok { c.putBuiltType( sd ) }
}

func ( c *Compilation ) buildTypesInitial( ctxs []buildContext ) {
    for _, bc := range ctxs {
        switch bc.td.( type ) {
        case *tree.StructDecl: c.buildStructType( bc )
        case *tree.EnumDecl: c.buildEnumType( bc )
        case *tree.PrototypeDecl: c.buildPrototypeType( bc )
        case *tree.ServiceDecl: c.buildServiceType( bc )
        }
    }
}

type compiledExpression struct {
    exp code.Expression
    typ mg.TypeReference
}

type prefixNode interface {
    compile( expctType mg.TypeReference, bs *buildScope ) *compiledExpression
}

type prefixLeaf struct { exp tree.Expression }

// 2nd return val indicates whether the first return val is an int type
func numExprResTypeOf( 
    expctType mg.TypeReference, 
    n *parser.NumericToken,
    errLoc *parser.Location,
    bs *buildScope ) ( mg.TypeReference, bool ) {
    if expctType == nil || qnameIn( expctType ).Equals( mg.QnameValue ) {
        if n.IsInt() { return mg.TypeInt64, true }
        return mg.TypeFloat64, false
    }
    switch qn := qnameIn( expctType ); true {
    case qn.Equals( mg.QnameInt32 ): return mg.TypeInt32, true
    case qn.Equals( mg.QnameInt64 ): return mg.TypeInt64, true
    case qn.Equals( mg.QnameUint32 ): return mg.TypeUint32, true
    case qn.Equals( mg.QnameUint64 ): return mg.TypeUint64, true
    case qn.Equals( mg.QnameFloat32 ): return mg.TypeFloat32, false
    case qn.Equals( mg.QnameFloat64 ): return mg.TypeFloat64, false
    }
    bs.c.addErrorf( errLoc, "Expected %s but got number", expctType )
    return nil, false
}

func asIntExpression(
    n *parser.NumericToken,
    resType mg.TypeReference,
    errLoc *parser.Location,
    bs *buildScope ) *compiledExpression {
    res := &compiledExpression{ typ: resType }
    if n.IsInt() {
        var sInt int64
        var uInt uint64
        var err error
        if resType.Equals( mg.TypeUint32 ) || resType.Equals( mg.TypeUint64 ) {
            uInt, err = n.Uint64()
        } else { sInt, err = n.Int64() }
        if err == nil {
            switch {
            case resType.Equals( mg.TypeInt32 ): res.exp = code.Int32( sInt )
            case resType.Equals( mg.TypeInt64 ): res.exp = code.Int64( sInt )
            case resType.Equals( mg.TypeUint32 ): res.exp = code.Uint32( uInt )
            case resType.Equals( mg.TypeUint64 ): res.exp = code.Uint64( uInt )
            default: panic( libErrorf( "Unhandled int type: %s", resType ) )
            }
            return res
        } else { panic( implErrorf( "Couldn't process int: %s", err ) ) }
    }
    bs.c.addErrorf( errLoc, "Expected %s but got float", resType )
    return nil
}

func asFloatExpression(
    n *parser.NumericToken,
    resType mg.TypeReference,
    errLoc *parser.Location,
    bs *buildScope ) *compiledExpression {
    res := &compiledExpression{ typ: resType }
    f, err := n.Float64(); 
    if err == nil { 
        if resType.Equals( mg.TypeFloat32 ) {
            res.exp = code.Float32( f )
        } else { res.exp = code.Float64( f ) }
        return res
    }
    panic( implErrorf( "Couldn't process float: %s", err ) )
}

func asNumberExpression(
    n *parser.NumericToken, 
    expctType mg.TypeReference,
    errLoc *parser.Location,
    bs *buildScope ) *compiledExpression {
    resType, takesInt := numExprResTypeOf( expctType, n, errLoc, bs )
    if resType == nil { return nil }
    if takesInt { return asIntExpression( n, resType, errLoc, bs ) }
    return asFloatExpression( n, resType, errLoc, bs )
}

func asStringExpression(
    str string,
    strLoc *parser.Location,
    expctType mg.TypeReference,
    bs *buildScope ) *compiledExpression {
    resType := expctType
    if expctType == nil || qnameIn( expctType ).Equals( mg.QnameString ) {
        if expctType == nil { resType = mg.TypeString }
        return &compiledExpression{ code.String( str ), resType }
    } else if qnameIn( expctType ).Equals( mg.QnameTimestamp ) {
        if expctType == nil { resType = mg.TypeTimestamp }
        if tm, ok := bs.parseTimestamp( str, strLoc ); ok {
            return &compiledExpression{ 
                &code.Timestamp{ tm }, mg.TypeTimestamp }
        }
        return nil
    }
    bs.c.addErrorf( strLoc, "Expected %s but got string", expctType )
    return nil
}

func asBooleanExpression( 
    kwd parser.Keyword, 
    errLoc *parser.Location,
    expctType mg.TypeReference,
    bs *buildScope ) *compiledExpression {
    res := &compiledExpression{}
    if expctType == nil {
        res.typ = mg.TypeBoolean
    } else {
        switch qn := qnameIn( expctType ); true {
        case qn.Equals( mg.QnameValue ): res.typ = mg.TypeBoolean
        case qn.Equals( mg.QnameBoolean ): res.typ = expctType
        default:
            bs.c.addErrorf( errLoc, "Expected %s but got boolean", expctType )
            return nil
        }
    }
    switch kwd {
    case parser.KeywordTrue: res.exp = code.Boolean( true )
    case parser.KeywordFalse: res.exp = code.Boolean( false )
    default:
        panic( implErrorf( "Invalid keyword as primary expression: %s", kwd ) )
    }
    return res
}

func asIdReferenceExpression(
    id *mg.Identifier,
    errLoc *parser.Location,
    typ mg.TypeReference,
    bs *buildScope ) *compiledExpression {
    return &compiledExpression{ &code.IdentifierReference{ id }, typ }
}

func ( l *prefixLeaf ) compilePrimary(
    pe *tree.PrimaryExpression, 
    expctType mg.TypeReference, 
    bs *buildScope ) *compiledExpression {
    switch v := pe.Prim.( type ) {
    case *parser.NumericToken: 
        return asNumberExpression( v, expctType, pe.PrimLoc, bs )
    case parser.StringToken: 
        return asStringExpression( string( v ), pe.PrimLoc, expctType, bs )
    case parser.Keyword: 
        return asBooleanExpression( v, pe.PrimLoc, expctType, bs )
    case *mg.Identifier:
        return asIdReferenceExpression( v, pe.PrimLoc, expctType, bs )
    }
    bs.c.addErrorf( pe, "Unhandled prim expression: %T", pe.Prim )
    return nil
}

func ( l *prefixLeaf ) compileNegation( 
    exp *compiledExpression, 
    errLoc *parser.Location, 
    bs *buildScope ) *compiledExpression {
    if baseTypeIsNum( exp.typ ) {
        return &compiledExpression{ &code.Negation{ exp.exp }, exp.typ }
    }
    bs.c.addErrorf( errLoc, "Cannot negate values of type %s", exp.typ )
    return nil
}

func ( l *prefixLeaf ) compileUnary( 
    exp *tree.UnaryExpression, 
    expctType mg.TypeReference, 
    bs *buildScope ) *compiledExpression {
    prim := l.implCompile( exp.Exp, expctType, bs )
    if prim == nil { return nil }
    errLoc := exp.Exp.Locate()
    switch exp.Op {
    case parser.SpecialTokenMinus: return l.compileNegation( prim, errLoc, bs )
    }
    bs.c.addErrorf( exp.OpLoc, "Illegal unary op: %s", exp.Op )
    return nil
}

func ( l *prefixLeaf ) compileEnumAccess(
    enDef *types.EnumDefinition,
    expType mg.TypeReference,
    id *mg.Identifier,
    idLoc *parser.Location,
    bs *buildScope ) *compiledExpression {
    if enVal := enDef.GetValue( id ); enVal != nil {
        return &compiledExpression{ &code.EnumValue{ enVal }, expType }
    }
    bs.c.addErrorf( idLoc, "Invalid value for enum %s: %s", 
        enDef.GetName(), id )
    return nil
}

func ( l *prefixLeaf ) compileQualified(
    exp *tree.QualifiedExpression,
    expctType mg.TypeReference,
    bs *buildScope ) *compiledExpression {
    if prim, ok := exp.Lhs.( *tree.PrimaryExpression ); ok {
        if t, ok := prim.Prim.( *parser.CompletableTypeReference ); ok {
            if typ := bs.resolveType( t, prim.Locate() ); typ != nil {
                if expctType == nil || typ.Equals( expctType ) {
                    if def := bs.c.typeDefForType( typ ); def != nil {
                        if enDef, ok := def.( *types.EnumDefinition ); ok {
                            return l.compileEnumAccess( 
                                enDef, typ, exp.Id, exp.IdLoc, bs )
                        }
                    }
                } else {
                    bs.c.addAssignError( prim, expctType, typ )
                    return nil
                }
            }
        }
    }
    panic( implErrorf( "Unhandled lhs (%T): %s", exp.Lhs, exp.Lhs ) )
}

func ( l *prefixLeaf ) compileListExpression(
    le *tree.ListExpression,
    expctType mg.TypeReference,
    bs *buildScope ) *compiledExpression {
    if expctType == nil { expctType = typeValList }
    if lt, ok := expctType.( *mg.ListTypeReference ); ok {
        valExp := code.NewListValue()
        for _, elt := range le.Elements {
            eltComp := bs.c.buildExpression( elt, lt.ElementType, bs )
            if eltComp == nil {
                ok = false
            } else { valExp.Values = append( valExp.Values, eltComp.exp ) }
        }
        if ok { return &compiledExpression{ valExp, expctType } }
        return nil
    }
    bs.c.addErrorf( le, "List value not expected" )
    return nil
}

func ( l *prefixLeaf ) implCompile(
    exp tree.Expression, 
    expctType mg.TypeReference, 
    bs *buildScope ) *compiledExpression {
    switch v := exp.( type ) {
    case *tree.PrimaryExpression: return l.compilePrimary( v, expctType, bs )
    case *tree.UnaryExpression: return l.compileUnary( v, expctType, bs )
    case *tree.QualifiedExpression: 
        return l.compileQualified( v, expctType, bs )
    case *tree.ListExpression:
        return l.compileListExpression( v, expctType, bs )
    }
    panic( implErrorf( "Unhandled leaf expression: %T", exp ) )
}

func ( l *prefixLeaf ) compile( 
    expctType mg.TypeReference, bs *buildScope ) *compiledExpression {
    return l.implCompile( l.exp, expctType, bs )
}

func ( c *Compilation ) infixToPrefix( exp tree.Expression ) prefixNode {
    switch v := exp.( type ) {
    case *tree.PrimaryExpression, 
         *tree.UnaryExpression, 
         *tree.ListExpression,
         *tree.QualifiedExpression:
        return &prefixLeaf{ v }
    case *tree.BinaryExpression:
        c.addErrorf( v, "Bin expressions not yet supported" )
    }
    panic( implErrorf( "Unhandled expression: %T", exp ) )
}

func ( c *Compilation ) buildExpression(
    exp tree.Expression, 
    expctType mg.TypeReference,
    bs *buildScope ) *compiledExpression {
    expTree := c.infixToPrefix( exp )
    if expTree == nil { return nil }
    return expTree.compile( expctType, bs )
}

func castConstVal( val mg.Value, typ mg.TypeReference ) ( mg.Value, error ) {
    rct := mgRct.NewDefaultCastReactor( typ )
    vb := mgRct.NewValueBuilder()
    pip := mgRct.InitReactorPipeline( rct, vb )
    if err := mgRct.VisitValue( val, pip ); err != nil { return nil, err }
    return vb.GetValue(), nil
}

func ( c *Compilation ) validateConstVal(
    val mg.Value, typ mg.TypeReference, errLoc *parser.Location ) bool {
    if _, err := castConstVal( val, typ ); err != nil {
        if ve, ok := err.( *mg.ValueCastError ); ok {
            c.addError( errLoc, ve.Message() )
        } else { c.addError( errLoc, err.Error() ) }
        return false
    }
    return true
}

func ( c *Compilation ) evaluateConstant(
    exp *compiledExpression,
    expctType mg.TypeReference,
    errLoc *parser.Location, 
    bs *buildScope ) mg.Value {
    val, err := interp.Evaluate( exp.exp )
    if err == nil {
        if ! c.validateConstVal( val, expctType, errLoc ) { return nil }
    } else {
        if evErr, ok := err.( *interp.EvaluationError ); ok {
            if ubErr, ok := evErr.Err.( *interp.UnboundIdentifierError ); ok {
                c.addErrorf( errLoc, 
                    "Found identifier in constant expression: %s", ubErr.Id )
                err = nil
            }
        }
        if err != nil { c.addError( errLoc, err.Error() ) }
        return nil
    }
    return val
}

func ( c *Compilation ) setFieldDefaults( 
    fldDecls []*tree.FieldDecl, fs *types.FieldSet, bs *buildScope ) {
    for _, fldDecl := range fldDecls {
        if deflExp := fldDecl.Default; deflExp != nil {
            fldDef := fs.Get( fldDecl.Name )
            fldType := fldDef.Type
            if exp := c.buildExpression( deflExp, fldType, bs ); exp != nil {
                errLoc := deflExp.Locate()
                fldDef.Default = c.evaluateConstant( exp, fldType, errLoc, bs )
            }
        } 
    }
}

func ( c *Compilation ) setStructFieldDefaults(
    bc buildContext, def types.Definition ) {
    c.setFieldDefaults( 
        bc.td.( tree.FieldContainer ).GetFields(),
        def.( types.FieldContainer ).GetFields(),
        bc.scope,
    )
}

func ( c *Compilation ) setServiceOpFieldDefaults(
    bc buildContext, sd *types.ServiceDefinition ) {
    decl := bc.td.( *tree.ServiceDecl )
    opDefs := types.OpDefsByName( sd.Operations )
    for _, opDecl := range decl.Operations {
        nm := opDecl.Name
        if opDef := opDefs.Get( nm ); opDef != nil {
            c.setFieldDefaults(
                opDecl.Call.Fields, 
                opDef.( *types.OperationDefinition ).Signature.GetFields(), 
                bc.scope )
        }
    }
}

func ( c *Compilation ) setDefFieldDefaults( ctxs []buildContext ) {
    for _, bc := range ctxs {
        switch def := c.typeDefForQn( bc.qname() ).( type ) {
        case *types.StructDefinition: c.setStructFieldDefaults( bc, def )
        case *types.PrototypeDefinition:
            fldDecls := bc.td.( *tree.PrototypeDecl ).Sig.Fields
            c.setFieldDefaults( fldDecls, def.Signature.GetFields(), bc.scope )
        case *types.ServiceDefinition: c.setServiceOpFieldDefaults( bc, def )
        }
    }
    for _, f := range c.onDefaults { f() }
}

func ( c *Compilation ) buildResult() *CompilationResult {
    return &CompilationResult{
        Errors: c.errs,
        BuiltTypes: c.builtTypes,
    }
}

// - Touch all type decls in all NsUnits. After this step it will be the case
// that no NsUnit purports to declare a type declared in any other NsUnit or in
// any imported libs. After this phase all types available to all NsUnits will
// be known, though not necessarily defined.
//
// - Build all aliased type defs. As of this writing, this step could actually
// be combined with the one following, since all type aliases are (re-)resolved
// dynamically. Later though there may be a pre-resolution phase, and in that
// case this step will need to precede other compilation steps. As for the
// moment, the reason to isolate this step is to ensure a specific processing
// order, which is only important to aid in our assertion of compiler error
// handling (for circular alias chains we'd like to use the order of error
// emission as part of our assertions).
//
// - Define all schema types
//
// - Define all declared instantiable types using the types named in step 1, but
// ignoring field defaults, since correct evaluation of default expressions may
// depend on knowledge of some as-yet-undefined type.
//
// - For any instantiable type defined above involving a field default
// expression, redefine that type, this time computing the field defaults.
//
// - Define services.
//
func ( c *Compilation ) Execute() ( cr *CompilationResult, err error ) {
    if err = c.validate(); err != nil { return }
    ctxs := c.initBuildContexts()
    c.printBuildOrder( ctxs )
    c.buildAliasedTypes( ctxs )
    c.buildSchemaTypes( ctxs )
    c.buildTypesInitial( ctxs )
    c.setDefFieldDefaults( ctxs )
    return c.buildResult(), nil
}
