package tree 

import (
//    "log"
    mg "mingle"
    "mingle/parser"
    "io"
    "fmt"
    "strings"
)

const (
    kwdAlias = parser.KeywordAlias
    kwdDefault = parser.KeywordDefault
    kwdEnum = parser.KeywordEnum
    kwdImport = parser.KeywordImport
    kwdNamespace = parser.KeywordNamespace
    kwdPrototype = parser.KeywordPrototype
    kwdSchema = parser.KeywordSchema
    kwdService = parser.KeywordService
    kwdStruct = parser.KeywordStruct
    kwdThrows = parser.KeywordThrows
    kwdUnion = parser.KeywordUnion
)

var (
    idVersion = mg.NewIdentifierUnsafe( []string{ "version" } )
    IdConstructor = mg.NewIdentifierUnsafe( []string{ "constructor" } )
    IdSecurity = mg.NewIdentifierUnsafe( []string{ "security" } )
    IdSchema = mg.NewIdentifierUnsafe( []string{ "schema" } )

    structureElementKeys = []*mg.Identifier{ IdConstructor, IdSchema }
    serviceElementKeys = []*mg.Identifier{ IdSecurity }

    typeDeclKwds = []parser.Keyword{ 
        kwdStruct, 
        kwdEnum,
        kwdPrototype,
        kwdService,
        kwdAlias,
        kwdSchema,
        kwdUnion,
    }

    binaryOps = []parser.SpecialToken{
        parser.SpecialTokenPlus,
        parser.SpecialTokenMinus,
        parser.SpecialTokenAsterisk,
        parser.SpecialTokenForwardSlash,
    }

    unaryOps = []parser.SpecialToken{
        parser.SpecialTokenMinus,
        parser.SpecialTokenPlus,
    }

    tkOpenBracket = parser.SpecialTokenOpenBracket
    tkCloseParen = parser.SpecialTokenCloseParen
    tkCloseBrace = parser.SpecialTokenCloseBrace
    tkCloseBracket = parser.SpecialTokenCloseBracket
    tkComma = parser.SpecialTokenComma
    tkColon = parser.SpecialTokenColon
    tkAsperand = parser.SpecialTokenAsperand
    tkAsterisk = parser.SpecialTokenAsterisk
    tkSemicolon = parser.SpecialTokenSemicolon
    tkSynthEnd = parser.SpecialTokenSynthEnd
    tkPeriod = parser.SpecialTokenPeriod
    tkForwardSlash = parser.SpecialTokenForwardSlash
    tkMinus = parser.SpecialTokenMinus
)

// enclEnd is the token which ends the enclosing body of the fields (call sig or
// struct body at the moment); seps is all toks that can end a field, and
// will include enclEnd.
type fieldEnds struct {
    seps []parser.SpecialToken
    enclEnd parser.SpecialToken
}

var fldEndsStruct = &fieldEnds{
    seps: []parser.SpecialToken{ tkSemicolon, tkSynthEnd, tkCloseBrace },
    enclEnd: tkCloseBrace,
}

var fldEndsCall = &fieldEnds{
    seps: []parser.SpecialToken{ tkComma, tkCloseParen },
    enclEnd: tkCloseParen,
}

type Locatable interface { Locate() *parser.Location }

type TypeListEntry struct {
    Name *mg.DeclaredTypeName
    Loc *parser.Location
}

func ( e *TypeListEntry ) Locate() *parser.Location { return e.Loc }

type Import struct {
    Start *parser.Location
    Namespace *mg.Namespace
    NamespaceLoc *parser.Location
    IsGlob bool
    Includes []*TypeListEntry
    Excludes []*TypeListEntry
}

func ( i *Import ) Locate() *parser.Location { return i.Start }

func ( i *Import ) sanityCheck() {
    if len( i.Includes ) == 0 {
        if ! i.IsGlob {
            tmpl := "Import at %s is not a glob and has no includes"
            panic( libErrorf( tmpl, i.Start ) )
        }
    } else {
        if i.IsGlob {
            tmpl := "Created import at %s with glob and includes"
            panic( libErrorf( tmpl, i.Start ) )
        }
        if len( i.Excludes ) > 0 {
            tmpl := "Created import at %s with includes and excludes"
            panic( libErrorf( tmpl, i.Start ) )
        }
    }
} 

type NamespaceDecl struct {
    Start *parser.Location
    Namespace *mg.Namespace
}

func ( nsd *NamespaceDecl ) Locate() *parser.Location { return nsd.Start }

type SyntaxElement interface {}

type Expression interface { Locatable }

type PrimaryExpression struct {
    Prim interface{}
    PrimLoc *parser.Location
}

func ( pe *PrimaryExpression ) Locate() *parser.Location { return pe.PrimLoc }

type QualifiedExpression struct {
    Lhs Expression
    Id *mg.Identifier
    IdLoc *parser.Location
}

func ( qe *QualifiedExpression ) Locate() *parser.Location { 
    return qe.Lhs.Locate()
}

type UnaryExpression struct {
    Op parser.SpecialToken
    OpLoc *parser.Location
    Exp Expression
}

func ( ue *UnaryExpression ) Locate() *parser.Location { return ue.OpLoc }

type BinaryExpression struct {
    Left, Right Expression
    Op parser.SpecialToken
    OpLoc *parser.Location
}

func ( be *BinaryExpression ) Locate() *parser.Location { 
    return be.Left.Locate() 
}

type ListExpression struct {
    Elements []Expression
    Start *parser.Location
}

func ( le *ListExpression ) Locate() *parser.Location { return le.Start }

type ConstructorDecl struct {
    Start *parser.Location
    ArgType *parser.CompletableTypeReference
}

func ( cd *ConstructorDecl ) Locate() *parser.Location { return cd.Start }

type FieldDecl struct {
    Name *mg.Identifier
    NameLoc *parser.Location
    Type *parser.CompletableTypeReference
    Default Expression
}

func ( fd *FieldDecl ) Locate() *parser.Location { return fd.NameLoc }

type SchemaMixinDecl struct {
    Start *parser.Location
    Name mg.TypeName
    NameLoc *parser.Location
}

func ( sd *SchemaMixinDecl ) Locate() *parser.Location { return sd.Start }

type TypeDecl interface {
    GetName() *mg.DeclaredTypeName
    Locatable
}

type FieldContainer interface { GetFields() []*FieldDecl }

type TypeDeclInfo struct {
    Name *mg.DeclaredTypeName
    NameLoc *parser.Location
}

func ( i *TypeDeclInfo ) Locate() *parser.Location { return i.NameLoc }

type StructDecl struct {
    Start *parser.Location
    Info *TypeDeclInfo
    Fields []*FieldDecl
    Constructors []*ConstructorDecl
    Schemas []*SchemaMixinDecl
}

func ( sd *StructDecl ) GetTypeInfo() *TypeDeclInfo { return sd.Info }
func ( sd *StructDecl ) GetName() *mg.DeclaredTypeName { return sd.Info.Name }
func ( sd *StructDecl ) Locate() *parser.Location { return sd.Start }
func ( sd *StructDecl ) GetFields() []*FieldDecl { return sd.Fields }

func ( sd *StructDecl ) createKeyedEltsAcc() *mg.IdentifierMap {
    res := mg.NewIdentifierMap()
    res.Put( IdConstructor, make( []*ConstructorDecl, 0, 2 ) )
    res.Put( IdSchema, make( []*SchemaMixinDecl, 0, 2 ) )
    return res
}

func ( sd *StructDecl ) initKeyedElts( ke *mg.IdentifierMap ) {
    sd.Schemas = ke.Get( IdSchema ).( []*SchemaMixinDecl )
    sd.Constructors = ke.Get( IdConstructor ).( []*ConstructorDecl )
}

func ( sd *StructDecl ) setFields( flds []*FieldDecl ) { sd.Fields = flds }

func ( sd *StructDecl ) setInfo( inf *TypeDeclInfo ) { sd.Info = inf }

type SchemaDecl struct {
    Start *parser.Location
    Info *TypeDeclInfo
    Fields []*FieldDecl
    Schemas []*SchemaMixinDecl
}

func ( sd *SchemaDecl ) Locate() *parser.Location { return sd.Start }

func ( sd *SchemaDecl ) GetFields() []*FieldDecl { return sd.Fields }

func ( sd *SchemaDecl ) GetName() *mg.DeclaredTypeName { return sd.Info.Name }

func ( sd *SchemaDecl ) createKeyedEltsAcc() *mg.IdentifierMap {
    res := mg.NewIdentifierMap()
    res.Put( IdSchema, make( []*SchemaMixinDecl, 0, 2 ) )
    return res
}

func ( sd *SchemaDecl ) setInfo( inf *TypeDeclInfo ) { sd.Info = inf }

func ( sd *SchemaDecl ) initKeyedElts( ke *mg.IdentifierMap ) {
    sd.Schemas = ke.Get( IdSchema ).( []*SchemaMixinDecl )
}

func ( sd *SchemaDecl ) setFields( flds []*FieldDecl ) { sd.Fields = flds }

type structureDecl interface {
    createKeyedEltsAcc() *mg.IdentifierMap
    initKeyedElts( *mg.IdentifierMap )
    setFields( []*FieldDecl )
    setInfo( *TypeDeclInfo )
}

type EnumValue struct {
    Value *mg.Identifier
    ValueLoc *parser.Location
}

func ( ev *EnumValue ) Locate() *parser.Location { return ev.ValueLoc }

type EnumDecl struct {
    Start *parser.Location
    Name *mg.DeclaredTypeName
    NameLoc *parser.Location
    Values []*EnumValue
}

func ( ed *EnumDecl ) GetName() *mg.DeclaredTypeName { return ed.Name }
func ( ed *EnumDecl ) Locate() *parser.Location { return ed.Start }

type AliasDecl struct {
    Start *parser.Location
    Name *mg.DeclaredTypeName
    NameLoc *parser.Location
    Target *parser.CompletableTypeReference
}

func ( ad *AliasDecl ) GetName() *mg.DeclaredTypeName { return ad.Name }
func ( ad *AliasDecl ) Locate() *parser.Location { return ad.Start }

type ThrownType struct {
    Type *parser.CompletableTypeReference
}

func ( tt *ThrownType ) Locate() *parser.Location { return tt.Type.Location() }

type CallSignature struct {
    Start *parser.Location
    Fields []*FieldDecl
    Return *parser.CompletableTypeReference
    Throws []*ThrownType
}
func ( cs *CallSignature ) Locate() *parser.Location { return cs.Start }

type PrototypeDecl struct {
    Start *parser.Location
    Name *mg.DeclaredTypeName
    NameLoc *parser.Location
    Sig *CallSignature
}

func ( pd *PrototypeDecl ) GetName() *mg.DeclaredTypeName { return pd.Name }
func ( pd *PrototypeDecl ) Locate() *parser.Location { return pd.Start }

type OperationDecl struct {
    Name *mg.Identifier
    NameLoc *parser.Location
    Call *CallSignature
}

func ( od *OperationDecl ) Locate() *parser.Location { return od.NameLoc }

type SecurityDecl struct {
    Start *parser.Location
    Name mg.TypeName
    NameLoc *parser.Location
}

func ( sd *SecurityDecl ) Locate() *parser.Location { return sd.Start }

type ServiceDecl struct {
    Start *parser.Location
    Info *TypeDeclInfo
    Operations []*OperationDecl
    SecurityDecls []*SecurityDecl
}

func ( sd *ServiceDecl ) GetTypeInfo() *TypeDeclInfo { return sd.Info }
func ( sd *ServiceDecl ) GetName() *mg.DeclaredTypeName { return sd.Info.Name }
func ( sd *ServiceDecl ) Locate() *parser.Location { return sd.Start }

func ( sd *ServiceDecl ) createKeyedEltsAcc() *mg.IdentifierMap {
    res := mg.NewIdentifierMap()
    res.Put( IdSecurity, make( []*SecurityDecl, 0, 2 ) )
    return res
}

func ( sd *ServiceDecl ) initKeyedElts( ke *mg.IdentifierMap ) {
    sd.SecurityDecls = ke.Get( IdSecurity ).( []*SecurityDecl )
}

type UnionDecl struct {
    Start *parser.Location
    Info *TypeDeclInfo
    Types []*parser.CompletableTypeReference
}

func ( ud *UnionDecl ) GetName() *mg.DeclaredTypeName { return ud.Info.Name }

func ( ud *UnionDecl ) Locate() *parser.Location { return ud.Start }

type NsUnit struct {
    SourceName string
    Imports []*Import
    NsDecl *NamespaceDecl
    TypeDecls []TypeDecl
}

type parse struct {

    *parser.Builder

    // set before parsing anything else
    verDefl *mg.Identifier
}

func ( p *parse ) pollKeywordLoc( 
    kwds ...parser.Keyword ) ( parser.Keyword, *parser.Location, error ) {
    tn, err := p.PeekToken()
    if tn == nil { return "", nil, err }
    if kwdAct, ok := tn.Token.( parser.Keyword ); ok {
        for _, kwd := range kwds {
            if kwd == kwdAct {
                p.MustNextToken()
                return kwd, tn.Loc, nil
            }
        }
    }
    return "", nil, nil
}

func ( p *parse ) pollKeyword( 
    kwds ...parser.Keyword ) ( parser.Keyword, error ) {
    kwd, _, err := p.pollKeywordLoc( kwds... )
    return kwd, err
}

func ( p *parse ) expectKeyword( kwd parser.Keyword ) error {
    if act, err := p.pollKeyword( kwd ); err == nil {
        if act != kwd {
            msg := fmt.Sprintf( "keyword %q", kwd )
            return p.ErrorTokenUnexpected( msg, nil )
        }
    } else { return err }
    return nil
}

func ( p *parse ) passSpecial( 
    tok ...parser.SpecialToken ) ( *parser.Location, error ) {
    tn, err := p.ExpectSpecial( tok... )
    if err != nil { return nil, err }
    return tn.Loc, nil
}

func ( p *parse ) passSemicolon() ( *parser.Location, error ) {
    return p.passSpecial( parser.SpecialTokenSemicolon )
}

func ( p *parse ) passColon() ( *parser.Location, error ) {
    return p.passSpecial( parser.SpecialTokenColon )
}

func ( p *parse ) passForwardSlash() ( *parser.Location, error ) {
    return p.passSpecial( parser.SpecialTokenForwardSlash )
}

func ( p *parse ) passOpening( t parser.SpecialToken ) ( 
    lc *parser.Location, err error ) {
    if lc, err = p.passSpecial( t ); err != nil { return }
    return
}

func ( p *parse ) passClosing( 
    t parser.SpecialToken ) ( *parser.Location, error ) {

    return p.passSpecial( t )
}

func ( p *parse ) passOpenBrace() ( *parser.Location, error ) {
    return p.passOpening( parser.SpecialTokenOpenBrace )
}

func ( p *parse ) passCloseBrace() ( *parser.Location, error ) {
    return p.passSpecial( parser.SpecialTokenCloseBrace )
}

func ( p *parse ) passOpenParen() ( *parser.Location, error ) {
    return p.passOpening( parser.SpecialTokenOpenParen )
}

func ( p *parse ) passCloseParen() ( *parser.Location, error ) {
    return p.passClosing( parser.SpecialTokenCloseParen )
}

func ( p *parse ) passOpenBracket() ( *parser.Location, error ) {
    return p.passOpening( parser.SpecialTokenOpenBracket )
}

func ( p *parse ) passCloseBracket() ( *parser.Location, error ) {
    return p.passClosing( parser.SpecialTokenCloseBracket )
}

func ( p *parse ) passStatementEnd() ( *parser.Location, error ) {
    return p.passSpecial( tkSemicolon, tkSynthEnd )
}

func ( p *parse ) peekSpecial( s parser.SpecialToken ) ( bool, error ) {
    if tn, err := p.PeekToken(); err == nil {
        return tn.IsSpecial( s ), nil
    } else { return false, err }
    return false, nil
}

func ( p *parse ) expectIdentifier() ( *mg.Identifier, 
                                       *parser.Location, 
                                       error ) {

    tn, err := p.ExpectIdentifier()
    if err != nil { return nil, nil, err }
    return tn.Identifier(), tn.Loc, nil
}

func ( p *parse ) expectNamespace() ( *mg.Namespace, *parser.Location, error ) {
    return p.ExpectNamespace( p.verDefl )
}

func ( p *parse ) expectDeclaredTypeName() (
    *mg.DeclaredTypeName, *parser.Location, error ) {
    tn, err := p.ExpectDeclaredTypeName()
    if err != nil { return nil, nil, err }
    return tn.DeclaredTypeName(), tn.Loc, nil
}

func ( p *parse ) expectTypeName() ( mg.TypeName, *parser.Location, error ) {
    return p.ExpectTypeName( p.verDefl )
}

func ( p *parse ) expectTypeReference() (
    *parser.CompletableTypeReference, error ) {

    return p.ExpectTypeReference( p.verDefl )
}

func ( p *parse ) expectCommaOrEnd( 
    end parser.SpecialToken ) ( sawEnd bool, err error ) {
    var sawComma bool
    var tn *parser.TokenNode
    tn, err = p.PollSpecial( parser.SpecialTokenComma, end )
    if err != nil { return }
    if tn != nil {
        if sawComma = tn.SpecialToken() == parser.SpecialTokenComma; sawComma {
            if tn, err = p.PollSpecial( end ); err != nil { return }
            sawEnd = tn != nil
        } else { sawEnd = true }
    } 
    if ! ( sawComma || sawEnd ) {
        err = p.ErrorTokenUnexpected( ", or " + string( end ), nil )
    }
    return
}

func ( p *parse ) setNsVersion() error {
    if _, err := p.passSpecial( parser.SpecialTokenAsperand ); err != nil {
        return err
    }
    if id, _, err := p.expectIdentifier(); err == nil {
        if id.Equals( idVersion ) {
            var tn *parser.TokenNode
            if tn, err = p.ExpectIdentifier(); err == nil { 
                p.verDefl = tn.Identifier()
            } else { return err }
        } else { return p.ParseError( "Expected @version" ) }
    }
    if _, err := p.passStatementEnd(); err != nil { return err }
    return nil
}

func ( p *parse ) pollImportNs() ( 
    ns *mg.Namespace, lc *parser.Location, err error ) {
    var tn *parser.TokenNode
    if tn, err = p.PeekToken(); err == nil {
        if _, ok := tn.Token.( *mg.Identifier ); ok {
            if ns, lc, err = p.expectNamespace(); err != nil { return }
            _, err = p.passForwardSlash()
            return
        }
    }
    return
}

func ( p *parse ) fillTypeListEntries( 
    names []*TypeListEntry ) ( res []*TypeListEntry, err error ) {
    res = names
    if _, err = p.ExpectSpecial( tkOpenBracket ); err != nil { return }
    for loop := true; loop; {
        var tn *parser.TokenNode
        if tn, err = p.PollSpecial( tkCloseBracket ); err == nil { 
            loop = tn == nil
        } else { return }
        if loop {
            e := new( TypeListEntry )
            if e.Name, e.Loc, err = p.expectDeclaredTypeName(); err == nil {
                res = append( res, e )
            } else { return }
            if tn, err = p.PollSpecial( tkCloseBracket ); err == nil {
                if loop = tn == nil; loop {
                    if _, err = p.ExpectSpecial( tkComma ); err != nil {
                        return
                    }
                }
            } else { return }
        }
    }
    return
}

func ( p *parse ) readTypeListEntries( names *[]*TypeListEntry ) ( err error ) {
    res := *names
    if res, err = p.fillTypeListEntries( res ); err != nil { return }
    if len( res ) == 0 { 
        err = p.ParseError( "Type list is empty" ) 
    } else { *names = res }
    return 
}

func ( p *parse ) completeImport( imprt *Import ) ( err error ) {
    var tn *parser.TokenNode
    if tn, err = p.PeekToken(); err != nil { return }
    if parser.IsSpecial( tn.Token, tkAsterisk ) {
        imprt.IsGlob = true
        p.MustNextToken()
        p.SetSynthEnd()
        if tn, err := p.PollSpecial( tkMinus ); err == nil {
            if tn != nil { err = p.readTypeListEntries( &imprt.Excludes ) }
        }
    } else if parser.IsSpecial( tn.Token, tkOpenBracket ) {
        if err = p.readTypeListEntries( &imprt.Includes ); err != nil { return }
    } else if _, ok := tn.Token.( *mg.DeclaredTypeName ); ok {
        e := &TypeListEntry{}
        if e.Name, e.Loc, err = p.expectDeclaredTypeName(); err == nil { 
            imprt.Includes = append( imprt.Includes, e )
        } else { return }
    } else { err = p.ErrorTokenUnexpected( "* or type name", tn ) }
    _, err = p.passStatementEnd()
    return
}

func ( p *parse ) expectImportDecl( 
    lc *parser.Location ) ( imprt *Import, err error ) {
    imprt = &Import{ 
        Start: lc,
        Includes: []*TypeListEntry{},
        Excludes: []*TypeListEntry{},
    }
    if err = p.CheckUnexpectedEnd(); err != nil { return }
    if imprt.Namespace, imprt.NamespaceLoc, err = p.pollImportNs(); err != nil {
        return 
    }
    if err = p.completeImport( imprt ); err != nil { return }
    imprt.sanityCheck()
    return
}

func ( p *parse ) pollImport() ( imprt *Import, err error ) {
    var kwd parser.Keyword
    var lc *parser.Location
    if kwd, lc, err = p.pollKeywordLoc( kwdImport ); err == nil {
        if kwd != "" { imprt, err = p.expectImportDecl( lc ) }
    } 
    return
}

func ( p *parse ) pollImports() ( []*Import, error ) {
    res := make( []*Import, 0, 4 )
    for {
        if imprt, err := p.pollImport(); err == nil {
            if imprt == nil { return res, nil }
            res = append( res, imprt )
        } else { return nil, err }
    }
    panic( libErrorf( "unreachable" ) )
}

func ( p *parse ) expectNsUnitNs() ( decl *NamespaceDecl, err error ) {
    if err = p.expectKeyword( kwdNamespace ); err != nil { return }
    decl = new( NamespaceDecl )
    decl.Namespace, decl.Start, err = p.expectNamespace()
    if declVer := decl.Namespace.Version; ! declVer.Equals( p.verDefl ) {
        tmpl := "Source version is '%s' but namespace declared '%s'"
        err = p.ParseError( tmpl, declVer, p.verDefl )
        return
    }
    _, err = p.passStatementEnd()
    return
}

func ( p *parse ) expectTypeDeclInfo() ( info *TypeDeclInfo, err error ) {
    info = new( TypeDeclInfo )
    info.Name, info.NameLoc, err = p.expectDeclaredTypeName()
    return
}

func unexpectedKeyedElementMsg( key *mg.Identifier ) string {
    keyStr := key.Format( mg.LcCamelCapped )
    return fmt.Sprintf( "Unexpected keyed definition @%s", keyStr )
}

func ( p *parse ) errorUnexpectedKeyedElement( key *mg.Identifier ) error {
    return p.ParseError( unexpectedKeyedElementMsg( key ) )
}

func ( p *parse ) addConstructorDecl(
    elts *mg.IdentifierMap, lc *parser.Location ) ( err error ) {

    cd := &ConstructorDecl{ Start: lc }
    if _, err = p.passOpenParen(); err != nil { return }
    if cd.ArgType, err = p.expectTypeReference(); err != nil { return }
    if _, err = p.passCloseParen(); err != nil { return }
    s := append( elts.Get( IdConstructor ).( []*ConstructorDecl ), cd )
    elts.Put( IdConstructor, s )
    return
}

func ( p *parse ) addSecurityDecl( 
    elts *mg.IdentifierMap, lc *parser.Location ) ( err error ) {

    sd := &SecurityDecl{ Start: lc }
    if sd.Name, sd.NameLoc, err = p.expectTypeName(); err != nil { return }
    s := append( elts.Get( IdSecurity ).( []*SecurityDecl ), sd )
    elts.Put( IdSecurity, s )
    return
}

func ( p *parse ) addSchemaDecl( 
    elts *mg.IdentifierMap, lc *parser.Location ) ( err error ) {

    sd := &SchemaMixinDecl{ Start: lc }
    if sd.Name, sd.NameLoc, err = p.expectTypeName(); err != nil { return }
    s := append( elts.Get( IdSchema ).( []*SchemaMixinDecl ), sd )
    elts.Put( IdSchema, s )
    return
}

func isAllowedKeyedElement( key *mg.Identifier, allow []*mg.Identifier ) bool {
    for _, id := range allow { if id.Equals( key ) { return true } }
    return false
}

func ( p *parse ) expectKeyedElement( 
    elts *mg.IdentifierMap, allow []*mg.Identifier ) ( err error ) {

    var lc *parser.Location
    if lc, err = p.passSpecial( tkAsperand ); err != nil { return }
    var key *mg.Identifier
    if key, _, err = p.expectIdentifier(); err != nil { return }
    if ! isAllowedKeyedElement( key, allow ) {
        return &parser.ParseError{ unexpectedKeyedElementMsg( key ), lc }
    }
    switch {
    case key.Equals( IdConstructor ): err = p.addConstructorDecl( elts, lc )
    case key.Equals( IdSecurity ): err = p.addSecurityDecl( elts, lc )
    case key.Equals( IdSchema ): err = p.addSchemaDecl( elts, lc )
    default: err = p.errorUnexpectedKeyedElement( key )
    }
    if err != nil { return }
    var sawBrace bool
    if sawBrace, err = p.peekSpecial( tkCloseBrace ); err == nil {
        if ! sawBrace { _, err = p.passStatementEnd() }
    } 
    return
}

func isUnaryOp( t parser.SpecialToken ) bool {
    for _, op := range unaryOps { if t == op { return true } }
    return false
}

// tn is the token -- which will be returned by NextToken() -- that tells us
// some sort of composite unary expression is expected
func ( p *parse ) expectCompositeUnaryExpression( 
    tn *parser.TokenNode ) ( e Expression, err error ) {
    switch spec := tn.SpecialToken(); {
    case spec == tkOpenBracket: e, err = p.expectListExpression()
    case isUnaryOp( spec ):
        p.MustNextToken()
        ue := &UnaryExpression{ Op: spec, OpLoc: tn.Loc }
        if ue.Exp, err = p.expectUnaryExpression(); err != nil { return }
        e = ue
    default: err = p.ErrorTokenUnexpected( "unary expression", tn )
    }
    return
}

func ( p *parse ) expectQualifiedAccessExpression( 
    typ *parser.CompletableTypeReference ) ( *QualifiedExpression, error ) {

    var err error
    pe := &PrimaryExpression{ Prim: typ, PrimLoc: typ.Location() }
    res := &QualifiedExpression{ Lhs: pe }
    if _, err = p.ExpectSpecial( tkPeriod ); err != nil { return nil, err }
    if res.Id, res.IdLoc, err = p.expectIdentifier(); err != nil {
        return nil, err
    }
    return res, nil
}

func ( p *parse ) expectIdentifiedExpression() ( Expression, error ) {
    tn, err := p.PeekToken()
    if err != nil { return nil, err }
    switch tn.Token.( type ) {
    case *mg.Identifier:
        pe := new( PrimaryExpression )
        pe.Prim, pe.PrimLoc, err = p.expectIdentifier()
        return pe, nil
    case *mg.DeclaredTypeName: 
        if typ, err := p.expectTypeReference(); err == nil {
            return p.expectQualifiedAccessExpression( typ )
        } else { return nil, err }
    }
    return nil, nil
}

func ( p *parse ) expectUnaryExpression() ( e Expression, err error ) {
    var tn *parser.TokenNode
    if err = p.CheckUnexpectedEnd(); err != nil { return }
    if tn, err = p.PeekToken(); err != nil { return }
    switch v := tn.Token.( type ) {
    case parser.StringToken, *parser.NumericToken: 
        p.MustNextToken()
        e = &PrimaryExpression{ Prim: v, PrimLoc: tn.Loc }
    case *mg.Identifier, *mg.DeclaredTypeName:
        e, err = p.expectIdentifiedExpression()
    case parser.Keyword:
        if v == parser.KeywordTrue || v == parser.KeywordFalse { 
            p.MustNextToken()
            e = &PrimaryExpression{ Prim: v, PrimLoc: tn.Loc } 
        }
    case parser.SpecialToken: e, err = p.expectCompositeUnaryExpression( tn )
    }
    if err == nil && e == nil { 
        err = p.ErrorTokenUnexpected( "unary expression", tn ) 
    }
    return
}

func ( p *parse ) expectListExpression() ( e *ListExpression, err error ) {
    e = &ListExpression{ Elements: make( []Expression, 0, 4 ) }
    if e.Start, err = p.passOpenBracket(); err != nil { return }
    for {
        var tn *parser.TokenNode 
        if tn, err = p.PeekToken(); err != nil { return }
        if parser.IsSpecial( tn.Token, tkCloseBracket ) { 
            p.MustNextToken() // consume ']'
            return
        }
        var elt Expression
        if elt, err = p.expectExpression(); err != nil { return }
        e.Elements = append( e.Elements, elt )
        var sawEnd bool
        sawEnd, err = p.expectCommaOrEnd( tkCloseBracket )
        if err != nil || sawEnd { return }
    }
    panic( libErrorf( "unreachable" ) )
}

func ( p *parse ) expectExpression() ( e Expression, err error ) {
    if e, err = p.expectUnaryExpression(); err != nil { return }
    for loop := true; loop; {
        var tn *parser.TokenNode
        if tn, err = p.PollSpecial( binaryOps... ); err != nil { return }
        if loop = tn != nil; loop {
            var right Expression
            if right, err = p.expectUnaryExpression(); err != nil { return }
            e = &BinaryExpression{ 
                Left: e, 
                Op: tn.SpecialToken(), 
                OpLoc: tn.Loc, 
                Right: right,
            }
        }
    }
    return
}

func ( p *parse ) expectFieldEnd( ends *fieldEnds ) ( sawEnd bool, err error ) {
    var tn *parser.TokenNode
    if tn, err = p.ExpectSpecial( ends.seps... ); err != nil { return }
    if tn == nil {
        err = p.ErrorTokenUnexpected( "field end", nil )
    } else { sawEnd = tn.SpecialToken() == ends.enclEnd }
    return
}

func ( p *parse ) expectFieldDecl(
    ends *fieldEnds ) ( fd *FieldDecl, sawEnd bool, err error ) {
    fd = new( FieldDecl )
    if fd.Name, fd.NameLoc, err = p.expectIdentifier(); err != nil { return }
    if fd.Type, err = p.expectTypeReference(); err != nil { return }
    var kwd parser.Keyword
    if kwd, err = p.pollKeyword( kwdDefault ); err != nil { return }
    if kwd != "" {
        if fd.Default, err = p.expectExpression(); err != nil { return }
    }
    sawEnd, err = p.expectFieldEnd( ends )
    return
}

func ( p *parse ) expectStructBody( sd structureDecl ) error {
    flds := make( []*FieldDecl, 0, 4 )
    ke := sd.createKeyedEltsAcc()
    for loop := true; loop; {
        tn, err := p.PeekToken()
        if err != nil { return err }
        switch {
        case parser.IsSpecial( tn.Token, parser.SpecialTokenAsperand ):
            if err = p.expectKeyedElement( ke, structureElementKeys ); 
               err != nil {
                return err
            }
        case parser.IsSpecial( tn.Token, parser.SpecialTokenCloseBrace ):
            loop, _ = false, p.MustNextToken()
        default:
            var fld *FieldDecl
            var sawEnd bool
            fld, sawEnd, err = p.expectFieldDecl( fldEndsStruct )
            if err != nil { return err }
            flds = append( flds, fld )
            loop = ! sawEnd
        }
    }
    sd.setFields( flds )
    sd.initKeyedElts( ke )
    if _, err := p.passStatementEnd(); err != nil { return err }
    return nil
}

func ( p *parse ) expectStructureDecl( sd structureDecl ) error {
    info, err := p.expectTypeDeclInfo()
    if err != nil { return err }
    sd.setInfo( info )
    if _, err = p.passOpenBrace(); err != nil { return err }
    return p.expectStructBody( sd )
}

func ( p *parse ) expectStructDecl(
    start *parser.Location ) ( sd *StructDecl, err error ) {

    sd = &StructDecl{ Start: start }
    return sd, p.expectStructureDecl( sd )
}

func ( p *parse ) expectSchemaDecl(
    start *parser.Location ) ( *SchemaDecl, error ) {

    sd := &SchemaDecl{ Start: start }
    return sd, p.expectStructureDecl( sd )
}

func ( p *parse ) completeEnumDecl( ed *EnumDecl ) ( err error ) {
    for {
        var val *mg.Identifier
        var lc *parser.Location
        if val, lc, err = p.expectIdentifier(); err == nil {
            ed.Values = append( ed.Values, &EnumValue{ val, lc } )
        } else { return }
        var sawEnd bool
        sawEnd, err = p.expectCommaOrEnd( parser.SpecialTokenCloseBrace )
        if err != nil { return }
        if sawEnd { 
            _, err = p.passStatementEnd()
            return 
        }
    }
    panic( libErrorf( "unreachable" ) )
}

func ( p *parse ) expectEnumDecl( 
    start *parser.Location ) ( ed *EnumDecl, err error ) {
    ed = &EnumDecl{ Start: start }
    if ed.Name, ed.NameLoc, err = p.expectDeclaredTypeName(); err != nil {
        return 
    }
    if _, err = p.passOpenBrace(); err != nil { return }
    ed.Values = make( []*EnumValue, 0, 8 )
    err = p.completeEnumDecl( ed )
    return
}

func ( p *parse ) expectAliasDecl(
    start *parser.Location ) ( ad *AliasDecl, err error ) {
    ad = &AliasDecl{ Start: start }
    if ad.Name, ad.NameLoc, err = p.expectDeclaredTypeName(); err != nil { 
        return
    }
    ad.Target, err = p.expectTypeReference()
    _, err = p.passStatementEnd()
    return
}

func ( p *parse ) completeCallFields() error {
    _, err := p.ExpectSpecial( tkColon )
    return err
}

func ( p *parse ) collectCallFields( cs *CallSignature ) ( err error ) {
    for {
        var tn *parser.TokenNode
        if tn, err = p.PollSpecial( tkCloseParen ); tn != nil || err != nil { 
            return p.completeCallFields()
        }
        var fld *FieldDecl
        var sawEnd bool
        fld, sawEnd, err = p.expectFieldDecl( fldEndsCall )
        if err != nil { return }
        cs.Fields = append( cs.Fields, fld )
        if sawEnd { return p.completeCallFields() }
    }
    panic( libErrorf( "unreachable" ) )
}

func ( p *parse ) expectThrownType() ( tt *ThrownType, err error ) {
    tt = new( ThrownType )
    tt.Type, err = p.expectTypeReference()
    return
}

func ( p *parse ) collectCallThrownTypes( cs *CallSignature ) ( err error ) {
    var kwd parser.Keyword
    if kwd, err = p.pollKeyword( kwdThrows ); kwd != "" {
        for loop := true; loop; {
            var tt *ThrownType
            if tt, err = p.expectThrownType(); err != nil { return }
            cs.Throws = append( cs.Throws, tt )
            var tn *parser.TokenNode
            if tn, err = p.PollSpecial( tkComma ); err != nil { return }
            loop = tn != nil
        }
    }
    return
}

func ( p *parse ) expectCallSignature() ( cs *CallSignature, err error ) {
    cs = &CallSignature{ 
        Fields: make( []*FieldDecl, 0, 4 ),
        Throws: make( []*ThrownType, 0, 4 ),
    }
    if cs.Start, err = p.passOpenParen(); err != nil { return }
    if err = p.collectCallFields( cs ); err != nil { return }
    if cs.Return, err = p.expectTypeReference(); err != nil { return }
    if _, err = p.PollSpecial( tkComma ); err != nil { return }
    err = p.collectCallThrownTypes( cs )
    _, err = p.passStatementEnd()
    return
}

func ( p *parse ) expectPrototypeDecl(
    start *parser.Location ) ( pd *PrototypeDecl, err error ) {
    pd = &PrototypeDecl{ Start: start }
    if pd.Name, pd.NameLoc, err = p.expectDeclaredTypeName(); err != nil {
        return
    }
    pd.Sig, err = p.expectCallSignature()
    return
}

func ( p *parse ) collectCallSignature( sd *ServiceDecl ) ( err error ) {
    od := &OperationDecl{}
    if od.Name, od.NameLoc, err = p.expectIdentifier(); err != nil { return }
    od.Call, err = p.expectCallSignature()
    sd.Operations = append( sd.Operations, od )
    return
}

func ( p *parse ) expectServiceDecl(
    start *parser.Location ) ( sd *ServiceDecl, err error ) {

    sd = &ServiceDecl{ Start: start }
    sd.Operations = make( []*OperationDecl, 0, 4 )
    ke := sd.createKeyedEltsAcc()
    if sd.Info, err = p.expectTypeDeclInfo(); err != nil { return }
    if _, err = p.passOpenBrace(); err != nil { return }
    for err == nil {
        var tn *parser.TokenNode
        if tn, err = p.PeekToken(); err != nil { return }
        if parser.IsSpecial( tn.Token, tkCloseBrace ) {
            p.MustNextToken()
            if _, err = p.passStatementEnd(); err != nil { return }
            sd.initKeyedElts( ke )
            return
        } else if parser.IsSpecial( tn.Token, tkAsperand ) {
            err = p.expectKeyedElement( ke, serviceElementKeys )
        } else if tn.IsKeyword( parser.KeywordOp ) {
            p.MustNextToken()
            err = p.collectCallSignature( sd )
        } else { err = p.ErrorTokenUnexpected( "operation or keyed def", tn ) }
    }
    return
}

func ( p *parse ) passUnionTypeElement() ( bool, error ) {
    tn, err := p.PollSpecial( tkComma )
    if err != nil { return false, err }
    tn, err = p.PollSpecial( tkCloseBrace )
    if err != nil { return false, err }
    sawEnd := tn != nil
    if sawEnd {
        if _, err = p.passStatementEnd(); err != nil { return false, err }
    }
    return sawEnd, nil
}

func ( p *parse ) expectUnionDecl( 
    start *parser.Location ) ( ud *UnionDecl, err error ) {

    ud = &UnionDecl{ Start: start }
    ud.Types = make( []*parser.CompletableTypeReference, 0, 4 )
    if ud.Info, err = p.expectTypeDeclInfo(); err != nil { return }
    if _, err = p.passOpenBrace(); err != nil { return }
    for sawClose := false; ! sawClose; {
        var typ *parser.CompletableTypeReference
        if typ, err = p.expectTypeReference(); err != nil { 
            return 
        } else { ud.Types = append( ud.Types, typ ) }
        if sawClose, err = p.passUnionTypeElement(); err != nil { return }
    }
    return
}

func ( p *parse ) expectTypeDecl(
    kwd parser.Keyword, start *parser.Location ) ( TypeDecl, error ) {
    switch kwd {
    case kwdStruct: return p.expectStructDecl( start )
    case kwdEnum: return p.expectEnumDecl( start )
    case kwdAlias: return p.expectAliasDecl( start )
    case kwdPrototype: return p.expectPrototypeDecl( start )
    case kwdService: return p.expectServiceDecl( start )
    case kwdSchema: return p.expectSchemaDecl( start )
    case kwdUnion: return p.expectUnionDecl( start )
    }
    panic( libErrorf( "Unimplemented: %s", kwd ) )
}

func ( p *parse ) pollTypeDecl() ( td TypeDecl, err error ) {
    var kwd parser.Keyword
    var lc *parser.Location
    if kwd, lc, err = p.pollKeywordLoc( typeDeclKwds... ); err == nil {
        if kwd == "" {
            if p.HasTokens() {
                strs := make( []string, len( typeDeclKwds ) )
                for i, k := range( typeDeclKwds ) { strs[ i ] = string( k ) }
                kwdExpctStr := strings.Join( strs, "|" )
                err = p.ErrorTokenUnexpected( kwdExpctStr, nil )
            }
        } else { td, err = p.expectTypeDecl( kwd, lc ) }
    }
    return 
}

func ( p *parse ) pollTypeDecls() ( decls []TypeDecl, err error ) {
    decls = make( []TypeDecl, 0, 10 )
    for {
        var decl TypeDecl
        if decl, err = p.pollTypeDecl(); err == nil {
            if decl == nil { 
                if p.HasTokens() {
                    err = p.ErrorTokenUnexpected( "end of source", nil )
                }
                return 
            }
            decls = append( decls, decl )
        } else { return }
    }
    panic( libErrorf( "unreachable" ) )
}

func ( p *parse ) expectNsUnit( srcNm string ) ( u *NsUnit, err error ) {
    if err = p.setNsVersion(); err != nil { return }
    u = &NsUnit{ SourceName: srcNm }
    if u.Imports, err = p.pollImports(); err != nil { return }
    if u.NsDecl, err = p.expectNsUnitNs(); err != nil { return }
    if u.TypeDecls, err = p.pollTypeDecls(); err != nil { return }
    return
}

func ParseSource( srcNm string, r io.Reader ) ( *NsUnit, error ) {
    opts := &parser.LexerOptions{ Reader: r, SourceName: srcNm, Strip: true }
    lx := parser.NewLexer( opts )
    return (&parse{ Builder: parser.NewBuilder( lx ) }).expectNsUnit( srcNm )
}
