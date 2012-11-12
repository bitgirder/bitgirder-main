package tree 

import (
//    "log"
    mg "mingle"
    "mingle/parser/syntax"
    "mingle/parser/lexer"
    "mingle/parser/loc"
    "io"
    "fmt"
    "strings"
)

const (
    kwdAlias = lexer.KeywordAlias
    kwdDefault = lexer.KeywordDefault
    kwdEnum = lexer.KeywordEnum
    kwdImport = lexer.KeywordImport
    kwdNamespace = lexer.KeywordNamespace
    kwdPrototype = lexer.KeywordPrototype
    kwdService = lexer.KeywordService
    kwdStruct = lexer.KeywordStruct
    kwdThrows = lexer.KeywordThrows
)

var (
    idVersion = mg.MustIdentifier( "version" )
    IdConstructor = mg.MustIdentifier( "constructor" )
    IdSecurity = mg.MustIdentifier( "security" )

    typeDeclKwds = []lexer.Keyword{ 
        kwdStruct, 
        kwdEnum,
        kwdPrototype,
        kwdService,
        kwdAlias,
    }

    binaryOps = []lexer.SpecialToken{
        lexer.SpecialTokenPlus,
        lexer.SpecialTokenMinus,
        lexer.SpecialTokenAsterisk,
        lexer.SpecialTokenForwardSlash,
    }

    unaryOps = []lexer.SpecialToken{
        lexer.SpecialTokenMinus,
        lexer.SpecialTokenPlus,
    }

    tkOpenBracket = lexer.SpecialTokenOpenBracket
    tkCloseParen = lexer.SpecialTokenCloseParen
    tkCloseBrace = lexer.SpecialTokenCloseBrace
    tkCloseBracket = lexer.SpecialTokenCloseBracket
    tkComma = lexer.SpecialTokenComma
    tkColon = lexer.SpecialTokenColon
    tkAsperand = lexer.SpecialTokenAsperand
    tkAsterisk = lexer.SpecialTokenAsterisk
    tkSemicolon = lexer.SpecialTokenSemicolon
    tkSynthEnd = lexer.SpecialTokenSynthEnd
    tkPeriod = lexer.SpecialTokenPeriod
    tkForwardSlash = lexer.SpecialTokenForwardSlash
    tkMinus = lexer.SpecialTokenMinus
)

func panicf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "tree: " + tmpl, argv... )
}

// enclEnd is the token which ends the enclosing body of the fields (call sig or
// struct body at the moment); seps is all toks that can end a field, and
// will include enclEnd.
type fieldEnds struct {
    seps []lexer.SpecialToken
    enclEnd lexer.SpecialToken
}

var fldEndsStruct = &fieldEnds{
    seps: []lexer.SpecialToken{ tkSemicolon, tkSynthEnd, tkCloseBrace },
    enclEnd: tkCloseBrace,
}

var fldEndsCall = &fieldEnds{
    seps: []lexer.SpecialToken{ tkComma, tkCloseParen },
    enclEnd: tkCloseParen,
}

type Locatable interface { Locate() *loc.Location }

type TypeListEntry struct {
    Name *mg.DeclaredTypeName
    Loc *loc.Location
}

func ( e *TypeListEntry ) Locate() *loc.Location { return e.Loc }

type Import struct {
    Start *loc.Location
    Namespace *mg.Namespace
    NamespaceLoc *loc.Location
    IsGlob bool
    Includes []*TypeListEntry
    Excludes []*TypeListEntry
}

func ( i *Import ) Locate() *loc.Location { return i.Start }

func ( i *Import ) sanityCheck() {
    if len( i.Includes ) == 0 {
        if ! i.IsGlob {
            tmpl := "Import at %s is not a glob and has no includes"
            panic( panicf( tmpl, i.Start ) )
        }
    } else {
        if i.IsGlob {
            tmpl := "Created import at %s with glob and includes"
            panic( panicf( tmpl, i.Start ) )
        }
        if len( i.Excludes ) > 0 {
            tmpl := "Created import at %s with includes and excludes"
            panic( panicf( tmpl, i.Start ) )
        }
    }
} 

type NamespaceDecl struct {
    Start *loc.Location
    Namespace *mg.Namespace
}

func ( nsd *NamespaceDecl ) Locate() *loc.Location { return nsd.Start }

type SyntaxElement interface {}

type Expression interface { Locatable }

type PrimaryExpression struct {
    Prim interface{}
    PrimLoc *loc.Location
}

func ( pe *PrimaryExpression ) Locate() *loc.Location { return pe.PrimLoc }

type QualifiedExpression struct {
    Lhs Expression
    Id *mg.Identifier
    IdLoc *loc.Location
}

func ( qe *QualifiedExpression ) Locate() *loc.Location { 
    return qe.Lhs.Locate()
}

type UnaryExpression struct {
    Op lexer.SpecialToken
    OpLoc *loc.Location
    Exp Expression
}

func ( ue *UnaryExpression ) Locate() *loc.Location { return ue.OpLoc }

type BinaryExpression struct {
    Left, Right Expression
    Op lexer.SpecialToken
    OpLoc *loc.Location
}

func ( be *BinaryExpression ) Locate() *loc.Location { return be.Left.Locate() }

type ListExpression struct {
    Elements []Expression
    Start *loc.Location
}

func ( le *ListExpression ) Locate() *loc.Location { return le.Start }

type ConstructorDecl struct {
    Start *loc.Location
    ArgType *syntax.CompletableTypeReference
    ArgTypeLoc *loc.Location
}

func ( cd *ConstructorDecl ) Locate() *loc.Location { return cd.Start }

type FieldDecl struct {
    Name *mg.Identifier
    NameLoc *loc.Location
    Type *syntax.CompletableTypeReference
    TypeLoc *loc.Location
    Default Expression
}

func ( fd *FieldDecl ) Locate() *loc.Location { return fd.NameLoc }

type KeyedElements struct {
    elts *mg.IdentifierMap // vals are []SyntaxElement
}

func newKeyedElements() *KeyedElements {
    return &KeyedElements{ mg.NewIdentifierMap() }
}

func ( ke *KeyedElements ) Add( key *mg.Identifier, elt SyntaxElement ) {
    var elts []SyntaxElement
    if val := ke.elts.Get( key ); val != nil {
        elts = val.( []SyntaxElement ) 
    } else { elts = make( []SyntaxElement, 0, 4 ) }
    elts = append( elts, elt )
    ke.elts.Put( key, elts )
}

func ( ke *KeyedElements ) Len() int { return ke.elts.Len() }

func ( ke *KeyedElements ) Get( key *mg.Identifier ) []SyntaxElement {
    if val := ke.elts.Get( key ); val != nil { return val.( []SyntaxElement ) }
    return nil
}

func ( ke *KeyedElements ) EachPair( 
    f func( key *mg.Identifier, elts []SyntaxElement ) ) {
    ke.elts.EachPair( func( key *mg.Identifier, val interface{} ) {
        f( key, val.( []SyntaxElement ) )
    })
}

type TypeDecl interface {
    GetName() *mg.DeclaredTypeName
    Locatable
}

type Keyed interface { GetKeyedElements() *KeyedElements }

type FieldContainer interface { GetFields() []*FieldDecl }

type TypeDeclInfo struct {
    Name *mg.DeclaredTypeName
    NameLoc *loc.Location
    SuperType *syntax.CompletableTypeReference
    SuperTypeLoc *loc.Location
}

func ( i *TypeDeclInfo ) Locate() *loc.Location { return i.NameLoc }

type StructDecl struct {
    Start *loc.Location
    Info *TypeDeclInfo
    Fields []*FieldDecl
    *KeyedElements
}

func ( sd *StructDecl ) GetTypeInfo() *TypeDeclInfo { return sd.Info }
func ( sd *StructDecl ) GetName() *mg.DeclaredTypeName { return sd.Info.Name }
func ( sd *StructDecl ) Locate() *loc.Location { return sd.Start }
func ( sd *StructDecl ) GetFields() []*FieldDecl { return sd.Fields }

func ( sd *StructDecl ) GetKeyedElements() *KeyedElements {
    return sd.KeyedElements
}

type EnumValue struct {
    Value *mg.Identifier
    ValueLoc *loc.Location
}

func ( ev *EnumValue ) Locate() *loc.Location { return ev.ValueLoc }

type EnumDecl struct {
    Start *loc.Location
    Name *mg.DeclaredTypeName
    NameLoc *loc.Location
    Values []*EnumValue
}

func ( ed *EnumDecl ) GetName() *mg.DeclaredTypeName { return ed.Name }
func ( ed *EnumDecl ) Locate() *loc.Location { return ed.Start }

type AliasDecl struct {
    Start *loc.Location
    Name *mg.DeclaredTypeName
    NameLoc *loc.Location
    Target *syntax.CompletableTypeReference
    TargetLoc *loc.Location
}

func ( ad *AliasDecl ) GetName() *mg.DeclaredTypeName { return ad.Name }
func ( ad *AliasDecl ) Locate() *loc.Location { return ad.Start }

type ThrownType struct {
    Type *syntax.CompletableTypeReference
    TypeLoc *loc.Location
}

func ( tt *ThrownType ) Locate() *loc.Location { return tt.TypeLoc }

type CallSignature struct {
    Start *loc.Location
    Fields []*FieldDecl
    Return *syntax.CompletableTypeReference
    ReturnLoc *loc.Location
    Throws []*ThrownType
}
func ( cs *CallSignature ) Locate() *loc.Location { return cs.Start }

type PrototypeDecl struct {
    Start *loc.Location
    Name *mg.DeclaredTypeName
    NameLoc *loc.Location
    Sig *CallSignature
}

func ( pd *PrototypeDecl ) GetName() *mg.DeclaredTypeName { return pd.Name }
func ( pd *PrototypeDecl ) Locate() *loc.Location { return pd.Start }

type OperationDecl struct {
    Name *mg.Identifier
    NameLoc *loc.Location
    Call *CallSignature
}

func ( od *OperationDecl ) Locate() *loc.Location { return od.NameLoc }

type SecurityDecl struct {
    Start *loc.Location
    Name mg.TypeName
    NameLoc *loc.Location
}

func ( sd *SecurityDecl ) Locate() *loc.Location { return sd.Start }

type ServiceDecl struct {
    Start *loc.Location
    Info *TypeDeclInfo
    Operations []*OperationDecl
    *KeyedElements
}

func ( sd *ServiceDecl ) GetTypeInfo() *TypeDeclInfo { return sd.Info }
func ( sd *ServiceDecl ) GetName() *mg.DeclaredTypeName { return sd.Info.Name }
func ( sd *ServiceDecl ) Locate() *loc.Location { return sd.Start }

func ( sd *ServiceDecl ) GetKeyedElements() *KeyedElements {
    return sd.KeyedElements
}

type NsUnit struct {
    SourceName string
    Imports []*Import
    NsDecl *NamespaceDecl
    TypeDecls []TypeDecl
}

type parse struct {

    *syntax.Builder

    // set before parsing anything else
    verDefl syntax.Identifier
}

func ( p *parse ) pollKeywordLoc( 
    kwds ...lexer.Keyword ) ( lexer.Keyword, *loc.Location, error ) {
    tn, err := p.PeekToken()
    if tn == nil { return "", nil, err }
    if kwdAct, ok := tn.Token.( lexer.Keyword ); ok {
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
    kwds ...lexer.Keyword ) ( lexer.Keyword, error ) {
    kwd, _, err := p.pollKeywordLoc( kwds... )
    return kwd, err
}

func ( p *parse ) expectKeyword( kwd lexer.Keyword ) error {
    if act, err := p.pollKeyword( kwd ); err == nil {
        if act != kwd {
            msg := fmt.Sprintf( "keyword %q", kwd )
            return p.ErrorTokenUnexpected( msg, nil )
        }
    } else { return err }
    return nil
}

func ( p *parse ) passSpecial( 
    tok ...lexer.SpecialToken ) ( *loc.Location, error ) {
    tn, err := p.ExpectSpecial( tok... )
    if err != nil { return nil, err }
    return tn.Loc, nil
}

func ( p *parse ) passSemicolon() ( *loc.Location, error ) {
    return p.passSpecial( lexer.SpecialTokenSemicolon )
}

func ( p *parse ) passColon() ( *loc.Location, error ) {
    return p.passSpecial( lexer.SpecialTokenColon )
}

func ( p *parse ) passForwardSlash() ( *loc.Location, error ) {
    return p.passSpecial( lexer.SpecialTokenForwardSlash )
}

func ( p *parse ) passOpening( t lexer.SpecialToken ) ( 
    lc *loc.Location, err error ) {
    if lc, err = p.passSpecial( t ); err != nil { return }
    return
}

func ( p *parse ) passClosing( t lexer.SpecialToken ) ( *loc.Location, error ) {
    return p.passSpecial( t )
}

func ( p *parse ) passOpenBrace() ( *loc.Location, error ) {
    return p.passOpening( lexer.SpecialTokenOpenBrace )
}

func ( p *parse ) passCloseBrace() ( *loc.Location, error ) {
    return p.passSpecial( lexer.SpecialTokenCloseBrace )
}

func ( p *parse ) passOpenParen() ( *loc.Location, error ) {
    return p.passOpening( lexer.SpecialTokenOpenParen )
}

func ( p *parse ) passCloseParen() ( *loc.Location, error ) {
    return p.passClosing( lexer.SpecialTokenCloseParen )
}

func ( p *parse ) passOpenBracket() ( *loc.Location, error ) {
    return p.passOpening( lexer.SpecialTokenOpenBracket )
}

func ( p *parse ) passCloseBracket() ( *loc.Location, error ) {
    return p.passClosing( lexer.SpecialTokenCloseBracket )
}

func ( p *parse ) passStatementEnd() ( *loc.Location, error ) {
    return p.passSpecial( tkSemicolon, tkSynthEnd )
}

func ( p *parse ) peekSpecial( s lexer.SpecialToken ) ( bool, error ) {
    if tn, err := p.PeekToken(); err == nil {
        return tn.IsSpecial( s ), nil
    } else { return false, err }
    return false, nil
}

func ( p *parse ) expectIdentifier() ( *mg.Identifier, *loc.Location, error ) {
    tn, err := p.ExpectIdentifier()
    if err != nil { return nil, nil, err }
    return mg.ConvertSyntaxId( tn.Identifier() ), tn.Loc, nil
}

func ( p *parse ) expectNamespace() ( *mg.Namespace, *loc.Location, error ) {
    ns, loc, err := p.ExpectNamespace( p.verDefl )
    if err != nil { return nil, nil, err }
    return mg.ConvertSyntaxNamespace( ns ), loc, nil
}

func ( p *parse ) expectDeclaredTypeName() (
    *mg.DeclaredTypeName, *loc.Location, error ) {
    tn, err := p.ExpectDeclaredTypeName()
    if err != nil { return nil, nil, err }
    nm := mg.ConvertSyntaxDeclaredTypeName( tn.DeclaredTypeName() )
    return nm, tn.Loc, nil
}

func ( p *parse ) expectTypeName() ( mg.TypeName, *loc.Location, error ) {
    sxNm, lc, err := p.ExpectTypeName( p.verDefl )
    if err != nil { return nil, nil, err }
    return mg.ConvertSyntaxTypeName( sxNm ), lc, nil
}

func ( p *parse ) expectTypeReference() (
    *syntax.CompletableTypeReference, *loc.Location, error ) {
    return p.ExpectTypeReference( p.verDefl )
}

func ( p *parse ) expectCommaOrEnd( 
    end lexer.SpecialToken ) ( sawEnd bool, err error ) {
    var sawComma bool
    var tn *syntax.TokenNode
    tn, err = p.PollSpecial( lexer.SpecialTokenComma, end )
    if err != nil { return }
    if tn != nil {
        if sawComma = tn.SpecialToken() == lexer.SpecialTokenComma; sawComma {
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
    if _, err := p.passSpecial( lexer.SpecialTokenAsperand ); err != nil {
        return err
    }
    if id, _, err := p.expectIdentifier(); err == nil {
        if id.Equals( idVersion ) {
            var tn *syntax.TokenNode
            if tn, err = p.ExpectIdentifier(); err == nil { 
                p.verDefl = tn.Identifier()
            } else { return err }
        } else { return p.ParseError( "Expected @version" ) }
    }
    if _, err := p.passStatementEnd(); err != nil { return err }
    return nil
}

func ( p *parse ) pollImportNs() ( 
    ns *mg.Namespace, lc *loc.Location, err error ) {
    var tn *syntax.TokenNode
    if tn, err = p.PeekToken(); err == nil {
        if _, ok := tn.Token.( syntax.Identifier ); ok {
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
        var tn *syntax.TokenNode
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
    var tn *syntax.TokenNode
    if tn, err = p.PeekToken(); err != nil { return }
    if syntax.IsSpecial( tn.Token, tkAsterisk ) {
        imprt.IsGlob = true
        p.MustNextToken()
        p.SetSynthEnd()
        if tn, err := p.PollSpecial( tkMinus ); err == nil {
            if tn != nil { err = p.readTypeListEntries( &imprt.Excludes ) }
        }
    } else if syntax.IsSpecial( tn.Token, tkOpenBracket ) {
        if err = p.readTypeListEntries( &imprt.Includes ); err != nil { return }
    } else if _, ok := tn.Token.( syntax.DeclaredTypeName ); ok {
        e := &TypeListEntry{}
        if e.Name, e.Loc, err = p.expectDeclaredTypeName(); err == nil { 
            imprt.Includes = append( imprt.Includes, e )
        } else { return }
    } else { err = p.ErrorTokenUnexpected( "* or type name", tn ) }
    _, err = p.passStatementEnd()
    return
}

func ( p *parse ) expectImportDecl( 
    lc *loc.Location ) ( imprt *Import, err error ) {
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
    var kwd lexer.Keyword
    var lc *loc.Location
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
    panic( panicf( "unreachable" ) )
}

func ( p *parse ) expectNsUnitNs() ( decl *NamespaceDecl, err error ) {
    if err = p.expectKeyword( kwdNamespace ); err != nil { return }
    decl = new( NamespaceDecl )
    decl.Namespace, decl.Start, err = p.expectNamespace()
    srcVer := mg.ConvertSyntaxId( p.verDefl )
    if declVer := decl.Namespace.Version; ! declVer.Equals( srcVer ) {
        tmpl := "Source version is '%s' but namespace declared '%s'"
        err = p.ParseError( tmpl, declVer, srcVer )
        return
    }
    _, err = p.passStatementEnd()
    return
}

func ( p *parse ) expectTypeDeclInfo() ( info *TypeDeclInfo, err error ) {
    info = new( TypeDeclInfo )
    if info.Name, info.NameLoc, err = p.expectDeclaredTypeName(); err != nil { 
        return 
    }
    var tn *syntax.TokenNode
    if tn, err = p.PollSpecial( lexer.SpecialTokenLessThan ); err == nil {
        if tn != nil {
            info.SuperType, info.SuperTypeLoc, err = p.expectTypeReference()
        }
    }
    return
}

func ( p *parse ) errorUnexpectedKeyedElement( key *mg.Identifier ) error {
    keyStr := key.Format( mg.LcCamelCapped )
    return p.ParseError( "Unexpected keyed definition @%s", keyStr )
}

func ( p *parse ) expectConstructorDecl(
    lc *loc.Location ) ( cd *ConstructorDecl, err error ) {
    cd = &ConstructorDecl{ Start: lc }
    if _, err = p.passOpenParen(); err != nil { return }
    if cd.ArgType, cd.ArgTypeLoc, err = p.expectTypeReference(); err != nil { 
        return 
    }
    if _, err = p.passCloseParen(); err != nil { return }
    return
}

func ( p *parse ) expectSecurityDecl(
    lc *loc.Location ) ( sd *SecurityDecl, err error ) {
    sd = &SecurityDecl{ Start: lc }
    sd.Name, sd.NameLoc, err = p.expectTypeName()
    return
}

func ( p *parse ) expectKeyedElement( keyed *KeyedElements ) ( err error ) {
    var lc *loc.Location
    if lc, err = p.passSpecial( tkAsperand ); err != nil { return }
    var key *mg.Identifier
    var elt SyntaxElement
    if key, _, err = p.expectIdentifier(); err != nil { return }
    switch {
    case key.Equals( IdConstructor ): elt, err = p.expectConstructorDecl( lc )
    case key.Equals( IdSecurity ): elt, err = p.expectSecurityDecl( lc )
    default: err = p.errorUnexpectedKeyedElement( key )
    }
    if err != nil { return }
    var sawBrace bool
    if sawBrace, err = p.peekSpecial( tkCloseBrace ); err == nil {
        if ! sawBrace { _, err = p.passStatementEnd() }
    } 
    if err != nil { return }
    if elt != nil { keyed.Add( key, elt ) }
    return
}

func isUnaryOp( t lexer.SpecialToken ) bool {
    for _, op := range unaryOps { if t == op { return true } }
    return false
}

// tn is the token -- which will be returned by NextToken() -- that tells us
// some sort of composite unary expression is expected
func ( p *parse ) expectCompositeUnaryExpression( 
    tn *syntax.TokenNode ) ( e Expression, err error ) {
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
    typ *syntax.CompletableTypeReference,
    lc *loc.Location ) ( *QualifiedExpression, error ) {
    var err error
    pe := &PrimaryExpression{ Prim: typ, PrimLoc: lc }
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
    case syntax.Identifier:
        pe := new( PrimaryExpression )
        pe.Prim, pe.PrimLoc, err = p.expectIdentifier()
        return pe, nil
    case syntax.DeclaredTypeName: 
        if typ, lc, err := p.expectTypeReference(); err == nil {
            return p.expectQualifiedAccessExpression( typ, lc )
        } else { return nil, err }
    }
    return nil, nil
}

func ( p *parse ) expectUnaryExpression() ( e Expression, err error ) {
    var tn *syntax.TokenNode
    if err = p.CheckUnexpectedEnd(); err != nil { return }
    if tn, err = p.PeekToken(); err != nil { return }
    switch v := tn.Token.( type ) {
    case lexer.StringToken, *lexer.NumericToken: 
        p.MustNextToken()
        e = &PrimaryExpression{ Prim: v, PrimLoc: tn.Loc }
    case syntax.Identifier, syntax.DeclaredTypeName:
        e, err = p.expectIdentifiedExpression()
    case lexer.Keyword:
        if v == lexer.KeywordTrue || v == lexer.KeywordFalse { 
            p.MustNextToken()
            e = &PrimaryExpression{ Prim: v, PrimLoc: tn.Loc } 
        }
    case lexer.SpecialToken: e, err = p.expectCompositeUnaryExpression( tn )
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
        var tn *syntax.TokenNode 
        if tn, err = p.PeekToken(); err != nil { return }
        if syntax.IsSpecial( tn.Token, tkCloseBracket ) { 
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
    panic( panicf( "unreachable" ) )
}

func ( p *parse ) expectExpression() ( e Expression, err error ) {
    if e, err = p.expectUnaryExpression(); err != nil { return }
    for loop := true; loop; {
        var tn *syntax.TokenNode
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
    var tn *syntax.TokenNode
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
    if fd.Type, fd.TypeLoc, err = p.expectTypeReference(); err != nil { return }
    var kwd lexer.Keyword
    if kwd, err = p.pollKeyword( kwdDefault ); err != nil { return }
    if kwd != "" {
        if fd.Default, err = p.expectExpression(); err != nil { return }
    }
    sawEnd, err = p.expectFieldEnd( ends )
    return
}

func ( p *parse ) expectStructBody( sd *StructDecl ) ( err error ) {
    sd.Fields = make( []*FieldDecl, 0, 4 )
    sd.KeyedElements = newKeyedElements()
    for loop := true; loop && err == nil; {
        var tn *syntax.TokenNode
        if tn, err = p.PeekToken(); err != nil { return }
        if syntax.IsSpecial( tn.Token, lexer.SpecialTokenAsperand ) {
            err = p.expectKeyedElement( sd.KeyedElements )
        } else if syntax.IsSpecial( tn.Token, lexer.SpecialTokenCloseBrace ) { 
            loop, _ = false, p.MustNextToken()
        } else {
            var fld *FieldDecl
            var sawEnd bool
            fld, sawEnd, err = p.expectFieldDecl( fldEndsStruct )
            if err != nil { return }
            sd.Fields = append( sd.Fields, fld )
            loop = ! sawEnd
        }
    }
    if err == nil { _, err = p.passStatementEnd() }
    return
}

func ( p *parse ) expectStructDecl(
    start *loc.Location ) ( sd *StructDecl, err error ) {
    sd = &StructDecl{ Start: start }
    if sd.Info, err = p.expectTypeDeclInfo(); err != nil { return }
    if _, err = p.passOpenBrace(); err != nil { return }
    err = p.expectStructBody( sd )
    return
}

func ( p *parse ) completeEnumDecl( ed *EnumDecl ) ( err error ) {
    for {
        var val *mg.Identifier
        var lc *loc.Location
        if val, lc, err = p.expectIdentifier(); err == nil {
            ed.Values = append( ed.Values, &EnumValue{ val, lc } )
        } else { return }
        var sawEnd bool
        sawEnd, err = p.expectCommaOrEnd( lexer.SpecialTokenCloseBrace )
        if err != nil { return }
        if sawEnd { 
            _, err = p.passStatementEnd()
            return 
        }
    }
    panic( panicf( "unreachable" ) )
}

func ( p *parse ) expectEnumDecl( 
    start *loc.Location ) ( ed *EnumDecl, err error ) {
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
    start *loc.Location ) ( ad *AliasDecl, err error ) {
    ad = &AliasDecl{ Start: start }
    if ad.Name, ad.NameLoc, err = p.expectDeclaredTypeName(); err != nil { 
        return
    }
    ad.Target, ad.TargetLoc, err = p.expectTypeReference()
    _, err = p.passStatementEnd()
    return
}

func ( p *parse ) completeCallFields() error {
    _, err := p.ExpectSpecial( tkColon )
    return err
}

func ( p *parse ) collectCallFields( cs *CallSignature ) ( err error ) {
    for {
        var tn *syntax.TokenNode
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
    panic( panicf( "unreachable" ) )
}

func ( p *parse ) expectThrownType() ( tt *ThrownType, err error ) {
    tt = new( ThrownType )
    tt.Type, tt.TypeLoc, err = p.expectTypeReference()
    return
}

func ( p *parse ) collectCallThrownTypes( cs *CallSignature ) ( err error ) {
    var kwd lexer.Keyword
    if kwd, err = p.pollKeyword( kwdThrows ); kwd != "" {
        for loop := true; loop; {
            var tt *ThrownType
            if tt, err = p.expectThrownType(); err != nil { return }
            cs.Throws = append( cs.Throws, tt )
            var tn *syntax.TokenNode
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
    if cs.Return, cs.ReturnLoc, err = p.expectTypeReference(); err != nil {
        return
    }
    if _, err = p.PollSpecial( tkComma ); err != nil { return }
    err = p.collectCallThrownTypes( cs )
    _, err = p.passStatementEnd()
    return
}

func ( p *parse ) expectPrototypeDecl(
    start *loc.Location ) ( pd *PrototypeDecl, err error ) {
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
    start *loc.Location ) ( sd *ServiceDecl, err error ) {
    sd = &ServiceDecl{ Start: start, KeyedElements: newKeyedElements() }
    sd.Operations = make( []*OperationDecl, 0, 4 )
    if sd.Info, err = p.expectTypeDeclInfo(); err != nil { return }
    if _, err = p.passOpenBrace(); err != nil { return }
    for err == nil {
        var tn *syntax.TokenNode
        if tn, err = p.PeekToken(); err != nil { return }
        if syntax.IsSpecial( tn.Token, tkCloseBrace ) {
            p.MustNextToken()
            _, err = p.passStatementEnd()
            return
        } else if syntax.IsSpecial( tn.Token, tkAsperand ) {
            err = p.expectKeyedElement( sd.KeyedElements )
        } else if tn.IsKeyword( lexer.KeywordOp ) {
            p.MustNextToken()
            err = p.collectCallSignature( sd )
        } else { err = p.ErrorTokenUnexpected( "operation or keyed def", tn ) }
    }
    return
}

func ( p *parse ) expectTypeDecl(
    kwd lexer.Keyword, start *loc.Location ) ( TypeDecl, error ) {
    switch kwd {
    case kwdStruct: return p.expectStructDecl( start )
    case kwdEnum: return p.expectEnumDecl( start )
    case kwdAlias: return p.expectAliasDecl( start )
    case kwdPrototype: return p.expectPrototypeDecl( start )
    case kwdService: return p.expectServiceDecl( start )
    }
    panic( panicf( "Unimplemented: %s", kwd ) )
}

func ( p *parse ) pollTypeDecl() ( td TypeDecl, err error ) {
    var kwd lexer.Keyword
    var lc *loc.Location
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
    panic( panicf( "unreachable" ) )
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
    opts := &lexer.Options{ Reader: r, SourceName: srcNm, Strip: true }
    lx := lexer.New( opts )
    return (&parse{ Builder: syntax.NewBuilder( lx ) }).expectNsUnit( srcNm )
}
