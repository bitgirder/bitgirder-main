package syntax

import(
//    "log"
    "io"
    "fmt"
    "strings"
    "bytes"
    "mingle/parser/lexer"
    "mingle/parser/loc"
)

type Identifier [][]byte

func ( id Identifier ) String() string {
    return string( bytes.Join( id, []byte{ '-' } ) )
}

type DeclaredTypeName []byte

type Namespace struct {
    Parts []Identifier   
    Version Identifier
}

type TypeName interface{}

type QualifiedTypeName struct {
    *Namespace
    Name DeclaredTypeName
}

type TokenNode struct {
    Token Token
    Loc *loc.Location
}

func ( tn *TokenNode ) Identifier() Identifier { 
    return tn.Token.( Identifier )
}

func ( tn *TokenNode ) DeclaredTypeName() DeclaredTypeName {
    return tn.Token.( DeclaredTypeName )
}

func ( tn *TokenNode ) Number() *lexer.NumericToken {
    return tn.Token.( *lexer.NumericToken )
}

func ( tn *TokenNode ) SpecialToken() lexer.SpecialToken {
    return tn.Token.( lexer.SpecialToken )
}

func ( tn *TokenNode ) IsSpecial( s lexer.SpecialToken ) bool {
    if spec, ok := tn.Token.( lexer.SpecialToken ); ok { return spec == s }
    return false
}

func ( tn *TokenNode ) IsKeyword( k lexer.Keyword ) bool {
    if kwd, ok := tn.Token.( lexer.Keyword ); ok { return kwd == k }
    return false
}

type Builder struct {
    lx *lexer.Lexer
}

func NewBuilder( lx *lexer.Lexer ) *Builder { return &Builder{ lx: lx } }

type tokenExpectType int

// Is any Token, but with lexer.Identifier and lexer.DeclaredTypeName
// replaced with the versions from this package
type Token interface{}

const (
    tokenExpectAny = tokenExpectType( iota )
    tokenExpectIdentifier
    tokenExpectDeclTypeName
    tokenExpectNumber
)

func ( sb *Builder ) SetSynthEnd() { sb.lx.SetSynthEnd() }

func ( sb *Builder ) convertToken( tok lexer.Token ) Token {
    if tok == nil { return nil }
    switch t := tok.( type ) {
    case lexer.Identifier:
        buf := [][]byte( t )
        parts := make( [][]byte, len( buf ) )
        for i, part := range buf { parts[ i ] = part }
        return Identifier( parts )
    case lexer.DeclaredTypeName: return DeclaredTypeName( []byte( t ) )
    }
    return Token( tok )
}

func ( sb * Builder ) readToken(
    et tokenExpectType ) ( tn *TokenNode, err error ) {
    var lxTok lexer.Token
    var lc *loc.Location
    switch et {
    case tokenExpectAny: lxTok, lc, err = sb.lx.ReadToken()
    case tokenExpectIdentifier: lxTok, lc, err = sb.lx.ReadIdentifier()
    case tokenExpectDeclTypeName: lxTok, lc, err = sb.lx.ReadDeclaredTypeName()
    case tokenExpectNumber: lxTok, lc, err = sb.lx.ReadNumber()
    default: panic( fmt.Errorf( "Unexpected token expect type: %v", et ) )
    }
    if lxTok != nil { tn = &TokenNode{ sb.convertToken( lxTok ), lc } }
    return
}

func ( sb *Builder ) peekToken(
    et tokenExpectType ) ( tn *TokenNode, err error ) {
    tn, err = sb.readToken( et )
//    if err != io.EOF { sb.lx.UnreadToken() }
    if err == nil { sb.lx.UnreadToken() }
    if err == io.EOF { err = nil }
    return
}

func ( sb *Builder ) implHasTokens( et tokenExpectType ) bool {
    tn, err := sb.peekToken( et )
    return ! ( tn == nil && err == nil )
}

func ( sb *Builder ) HasTokens() bool { 
    return sb.implHasTokens( tokenExpectAny )
}

func ( sb *Builder ) Location() *loc.Location { return sb.lx.Location() }

func ( sb *Builder ) ParseError( msg string, args ...interface{} ) error {
    return &loc.ParseError{ fmt.Sprintf( msg, args... ), sb.Location() }
}

const msgUnexpectedEndOfInput = "Unexpected end of input"

func ( sb *Builder ) errorUnexpectedEnd() error {
    return sb.ParseError( msgUnexpectedEndOfInput )
}

func ( sb *Builder ) CheckUnexpectedEnd() error {
    if sb.HasTokens() { return nil }
    return sb.errorUnexpectedEnd()
}

func ( sb *Builder ) PeekToken() ( *TokenNode, error ) {
    return sb.peekToken( tokenExpectAny )
}

func ( sb *Builder ) implNextToken( 
    msg string, et tokenExpectType ) ( *TokenNode, error ) {
    tn, err := sb.readToken( et )
    if err == io.EOF {
        if msg == "" { msg = msgUnexpectedEndOfInput }
        err = sb.ParseError( msg )
    }
    return tn, err
}

func ( sb *Builder ) nextToken() ( *TokenNode, error ) {
    return sb.implNextToken( "", tokenExpectAny )
}

func ( sb *Builder ) MustNextToken() Token {
    tn, err := sb.nextToken()
    if err != nil { panic( err ) }
    return tn.Token
}

func ( sb *Builder ) peekUnexpectedTokenErrorNode() *TokenNode {
    tn, err := sb.PeekToken()
    if err != nil && err != io.EOF {
        errMsg := "syntax: Error in PeekToken() during error construction: %s"
        errLoc := sb.lx.Location()
        panic( &loc.ParseError{ fmt.Sprintf( errMsg, err ), errLoc } )
    }
    return tn
}

// tn nil means the error is at next token loc (or END)
func ( sb *Builder ) ErrorTokenUnexpected( 
    expctDesc string, tn *TokenNode ) error {
    if tn == nil { tn = sb.peekUnexpectedTokenErrorNode() }
    var tokStr string
    var lc *loc.Location
    if tn == nil {
        tokStr, lc = "END", sb.lx.Location()
    } else { tokStr, lc = sb.errorStringForToken( tn.Token ), tn.Loc }
    var msg string
    if expctDesc == "" {
        msg = fmt.Sprintf( "Unexpected token: %s", tokStr )
    } else { 
        msg = fmt.Sprintf( "Expected %s but found: %s", expctDesc, tokStr ) 
    }
    return &loc.ParseError{ msg, lc }
}

func IsSpecial( tok Token, spec lexer.SpecialToken ) bool {
    act, ok := tok.( lexer.SpecialToken )
    return ok && act == spec
}

func ( sb *Builder ) PollSpecial( 
    l ...lexer.SpecialToken ) ( *TokenNode, error ) {
    tn, err := sb.PeekToken()
    if err == nil && tn != nil {
        for _, t := range l {
            if IsSpecial( tn.Token, t ) {
                sb.MustNextToken() // consume it
                return tn, nil
            }
        }
    }
    return nil, err // err may also be nil, but poll failed to match regardless
}

func ( sb *Builder ) ExpectSpecial(
    l ...lexer.SpecialToken ) ( tn *TokenNode, err error ) {
    if tn, err = sb.PollSpecial( l... ); err == nil && tn == nil {
        var msg string
        if len( l ) == 1 {
            msg = string( l[ 0 ] ) 
        } else {
            arr := make( []string, len( l ) )
            for i, s := range l { arr[ i ] = fmt.Sprintf( "%#v", string( s ) ) }
            msg = fmt.Sprintf( "one of [ %s ]", strings.Join( arr, ", " ) )
        }
        err = sb.ErrorTokenUnexpected( msg, nil )
    }
    return
}

func ( sb *Builder ) errorStringForToken( tok Token ) string {
    switch v := tok.( type ) {
    case lexer.WhitespaceToken: return fmt.Sprintf( "%#v", string( v ) )
    case lexer.NumericToken: return v.String()
    }
    return fmt.Sprintf( "%s", tok )
}

func ( sb *Builder ) CheckTrailingToken() error {
    var tn *TokenNode
    var err error
    if tn, err = sb.PollSpecial( lexer.SpecialTokenSynthEnd ); err != nil { 
        return err
    }
    tn, err = sb.PeekToken()
    if err != nil { return err }
    if tn != nil { return sb.ErrorTokenUnexpected( "", tn ) }
    return nil
}

func ( sb *Builder ) SkipWsOrComments() ( err error ) {
    for loop := true; loop && err == nil; {
        var tn *TokenNode
        if tn, err = sb.PeekToken(); err == nil && tn != nil {
            switch tn.Token.( type ) {
            case lexer.CommentToken, lexer.WhitespaceToken: sb.MustNextToken()
            default: loop = false
            }
        } else { loop = false } // possibly because tn == nil
    }
    return err
}

const expctDescNumericToken = "numeric token"

func ( sb *Builder ) implExpectTypedToken( 
    msg string, et tokenExpectType ) ( *TokenNode, error ) {
    tn, err := sb.implNextToken( msg, et )
    if err != nil { return nil, err }
    var ok bool
    switch tk := tn.Token; et {
    case tokenExpectIdentifier: _, ok = tk.( Identifier )
    case tokenExpectDeclTypeName: _, ok = tk.( DeclaredTypeName )
    case tokenExpectNumber: _, ok = tk.( *lexer.NumericToken )
    }
    if ok { return tn, nil }
    return nil, sb.ErrorTokenUnexpected( msg, tn )
}

func ( sb *Builder ) ExpectNumericToken() ( *TokenNode, error ) {
    return sb.implExpectTypedToken( expctDescNumericToken, tokenExpectNumber )
}

func ( sb *Builder ) ExpectIdentifier() ( *TokenNode, error ) {
    return sb.implExpectTypedToken( "identifier", tokenExpectIdentifier )
}

func ( sb *Builder ) ExpectDeclaredTypeName() ( *TokenNode, error ) {
    msg := "Expected declared type name"
    return sb.implExpectTypedToken( msg, tokenExpectDeclTypeName )
}

func ( sb *Builder ) accumulateNsParts( 
    ns *Namespace ) ( sawAsp bool, l *loc.Location, err error ) {
    for { 
        var tn *TokenNode
        if tn, err = sb.ExpectIdentifier(); err == nil {
            if l == nil { l = tn.Loc }
            ns.Parts = append( ns.Parts, tn.Identifier() ) 
        } else { return }
        if tn, err = sb.PeekToken(); err == nil && tn != nil {
            switch tn.Token {
            case lexer.SpecialTokenColon: sb.MustNextToken()
            case lexer.SpecialTokenAsperand: 
                sawAsp = true
                sb.MustNextToken()
                return
            default: return
            }
        } else { return }
    }
    panic( "unreachable" )
}

func ( sb *Builder ) ExpectNamespace( 
    verDefl Identifier ) ( ns *Namespace, l *loc.Location, err error ) {
    ns = &Namespace{ Parts: make( []Identifier, 0, 4 ) }
    var sawAsp bool
    if sawAsp, l, err = sb.accumulateNsParts( ns ); err != nil { return }
    switch {
    case sawAsp: 
        var tn *TokenNode
        if tn, err = sb.ExpectIdentifier(); err == nil {
            ns.Version = tn.Identifier()
        }
    default:
        if ns.Version = verDefl; ns.Version == nil {
            err = sb.ErrorTokenUnexpected( "':' or '@'", nil )
        }
    }
    return
}

func ( sb *Builder ) ExpectQualifiedTypeName( verDefl Identifier ) ( 
    qn *QualifiedTypeName, l *loc.Location, err error ) {
    qn = new( QualifiedTypeName )
    if qn.Namespace, l, err = sb.ExpectNamespace( verDefl ); err != nil { 
        return 
    }
    var tn *TokenNode
    if tn, err = sb.PollSpecial( lexer.SpecialTokenForwardSlash ); err == nil {
        if tn == nil { err = sb.ErrorTokenUnexpected( "type path", nil ) }
    } 
    if err != nil { return }
    if tn, err = sb.ExpectDeclaredTypeName(); err != nil { return }
    qn.Name = tn.DeclaredTypeName()
    return
}

func ( sb *Builder ) ExpectTypeName( 
    verDefl Identifier ) ( nm TypeName, l *loc.Location, err error ) {
    var tn *TokenNode
    if tn, err = sb.PeekToken(); err == nil && tn == nil {
        err = sb.errorUnexpectedEnd()
    } 
    if err != nil { return }
    switch tn.Token.( type ) {
    case Identifier: nm, l, err = sb.ExpectQualifiedTypeName( verDefl )
    case DeclaredTypeName: 
        if tn, err = sb.ExpectDeclaredTypeName(); err != nil { return }
        nm, l = tn.DeclaredTypeName(), tn.Loc
    default: 
        err = sb.ErrorTokenUnexpected( "identifier or declared type name", tn )
    }
    return
}

type TypeQuantifier int

const (
    TypeQuantifierNullable = TypeQuantifier( iota )
    TypeQuantifierList
    TypeQuantifierNonEmptyList
)

type quantList []lexer.SpecialToken

var quantToks []lexer.SpecialToken

func init() {
    quantToks = []lexer.SpecialToken{
        lexer.SpecialTokenPlus,
        lexer.SpecialTokenAsterisk,
        lexer.SpecialTokenQuestionMark,
    }
}

func ( sb *Builder ) expectTypeRefQuantCompleter() ( 
    quantList, error ) {
    quants := make( []lexer.SpecialToken, 0, 2 )
    for loop := true; sb.HasTokens() && loop; {
        var tn *TokenNode
        var err error
        if tn, err = sb.PollSpecial( quantToks... ); err == nil {
            if loop = tn != nil; loop {
                quants = append( quants, tn.SpecialToken() ) 
                sb.SetSynthEnd()
            }
        } else { return nil, err }
    }
    return quantList( quants ), nil
}

var rangeClosers []lexer.SpecialToken

func init() {
    rangeClosers = []lexer.SpecialToken{
        lexer.SpecialTokenCloseParen, 
        lexer.SpecialTokenCloseBracket,
    }
}

// removes and returns next token if it is '+' or '-', returning true if it was
// '-', false if it was anything other than '-' (not necessarily '+'); if '+' or
// '-' is seen
func ( sb *Builder ) PollPlusMinus() ( *TokenNode, bool, error ) {
    tn, err := sb.PeekToken()
    if err == nil && tn != nil {
        if spec, ok := tn.Token.( lexer.SpecialToken ); ok {
            if spec == lexer.SpecialTokenMinus || 
               spec == lexer.SpecialTokenPlus {
                sb.MustNextToken() // consume it
                return tn, spec == lexer.SpecialTokenMinus, nil
            }
        }
    }
    return nil, false, err
}

func ( sb *Builder ) expectRestrictionSyntax() (
    RestrictionSyntax, error ) {
    pmTok, neg, err := sb.PollPlusMinus()
    if err != nil { return nil, err }
    var tn *TokenNode
    if tn, err = sb.nextToken(); err != nil { return nil, err }
    if tn == nil { return nil, sb.errorUnexpectedEnd() }
    if s, ok := tn.Token.( lexer.StringToken ); ok { 
        // string ok unless we want num
        if pmTok != nil { return nil, sb.ErrorTokenUnexpected( "number", tn ) }
        return &StringRestrictionSyntax{ string( s ), tn.Loc }, nil
    }
    if num, ok := tn.Token.( *lexer.NumericToken ); ok {
        numLoc := tn.Loc
        if pmTok != nil { numLoc = pmTok.Loc }
        return &NumRestrictionSyntax{ neg, num, numLoc }, nil
    }
    return nil, sb.ErrorTokenUnexpected( "range value", tn )
}

func ( sb *Builder ) completeRangeLeftSide(
    rg *RangeRestrictionSyntax ) ( err error ) {
    if err = sb.SkipWsOrComments(); err != nil { return }
    var tn *TokenNode
    if tn, err = sb.PollSpecial( lexer.SpecialTokenComma ); err != nil { 
        return 
    }
    if tn == nil {
        if rg.Left, err = sb.expectRestrictionSyntax(); err != nil { return }
        if err = sb.SkipWsOrComments(); err != nil { return }
        if _, err = sb.ExpectSpecial( lexer.SpecialTokenComma ); err != nil { 
            return 
        }
    }
    return
}

func ( sb *Builder ) checkRange( 
    rg *RangeRestrictionSyntax, leftLoc, rightLoc *loc.Location ) error {
    errLf := rg.Left == nil && rg.LeftClosed
    errRt := rg.Right == nil && rg.RightClosed
    msg := ""
    var errLoc *loc.Location
    if errLf {
        if errRt { 
            msg, errLoc = "Infinite range must be open", leftLoc
        } else { msg, errLoc = "Infinite low range must be open", leftLoc }
    } else if errRt { 
        msg, errLoc = "Infinite high range must be open", rightLoc 
    }
    if msg != "" { return &loc.ParseError{ msg, errLoc } }
    return nil
}

func ( sb *Builder ) completeRangeRightSide( 
    rg *RangeRestrictionSyntax ) ( rightLoc *loc.Location, err error ) {
    var tn *TokenNode
    if tn, err = sb.PollSpecial( rangeClosers... ); err != nil { return }
    if tn == nil { 
        if rg.Right, err = sb.expectRestrictionSyntax(); err != nil { return }
        if err = sb.SkipWsOrComments(); err != nil { return }
        if tn, err = sb.ExpectSpecial( rangeClosers... ); err != nil { return }
    }
    rightLoc = tn.Loc
    closeTok := tn.SpecialToken() // either set before or inside branch above
    rg.RightClosed = closeTok == lexer.SpecialTokenCloseBracket
    return
}

func ( sb *Builder ) completeRangeRestriction( 
    tn *TokenNode ) ( rg *RangeRestrictionSyntax, err error ) {
    leftClosed := tn.SpecialToken() == lexer.SpecialTokenOpenBracket
    leftLoc := tn.Loc
    rg = &RangeRestrictionSyntax{ LeftClosed: leftClosed }
    if err = sb.completeRangeLeftSide( rg ); err != nil { return }
    if err = sb.SkipWsOrComments(); err != nil { return }
    var rightLoc *loc.Location
    if rightLoc, err = sb.completeRangeRightSide( rg ); err != nil { return }
    return rg, sb.checkRange( rg, leftLoc, rightLoc )
}

func ( sb *Builder ) pollTypeRefRestriction() (
    sx RestrictionSyntax, err error ) {
    if err = sb.SkipWsOrComments(); err != nil { return }
    var tn *TokenNode
    if tn, err = sb.PollSpecial( lexer.SpecialTokenTilde ); 
        err == nil && tn != nil {
        if err = sb.SkipWsOrComments(); err != nil { return }
        if tn, err = sb.nextToken(); err != nil { return }
        if IsSpecial( tn.Token, lexer.SpecialTokenOpenParen ) || 
           IsSpecial( tn.Token, lexer.SpecialTokenOpenBracket ) {
            sx, err = sb.completeRangeRestriction( tn )
            sb.SetSynthEnd()
        } else if s, ok := tn.Token.( lexer.StringToken ); ok {
            sx = &RegexRestrictionSyntax{ string( s ), tn.Loc }
            sb.SetSynthEnd()
        } else {
            err = sb.ErrorTokenUnexpected( "type restriction", tn )
        }
    } else { sb.SetSynthEnd() }
    return
}

type CompletableTypeReference struct {
    Name TypeName
    Restriction RestrictionSyntax
    quants quantList
}

type TypeCompleter func( typ interface{}, quant TypeQuantifier ) interface{}

func ( ctr *CompletableTypeReference ) CompleteType( 
    typ interface{}, tc TypeCompleter ) interface{} {
    for _, quant := range ctr.quants {
        switch quant {
        case lexer.SpecialTokenPlus: typ = tc( typ, TypeQuantifierNonEmptyList )
        case lexer.SpecialTokenAsterisk: typ = tc( typ, TypeQuantifierList )
        case lexer.SpecialTokenQuestionMark: 
            typ = tc( typ, TypeQuantifierNullable )
        default: panic( fmt.Errorf( "Unexpected quant: %s", quant ) )
        }
    }
    return typ
}

// In addition to returning a type ref, this method will, via its helper
// methods, ensure that sb.SetSynthEnd() is called after the token that
// completes the type reference, effectively making type references capable of
// ending statements with a synthetic end token, just as lexer does with
// numbers, strings, etc.
func ( sb *Builder ) ExpectTypeReference(
    verDefl Identifier ) ( ref *CompletableTypeReference, 
                           l *loc.Location,
                           err error ) {
    tmp := new( CompletableTypeReference )
    if tmp.Name, l, err = sb.ExpectTypeName( verDefl ); err != nil { return }
    if tmp.Restriction, err = sb.pollTypeRefRestriction(); err != nil { return }
    if tmp.quants, err = sb.expectTypeRefQuantCompleter(); err != nil { return }
    ref = tmp
    return
}

// meant for string/num restriction only right now
type RestrictionSyntax interface{}

type LiteralStringer interface {
    LiteralString() string
}

type StringRestrictionSyntax struct {
    Str string
    Loc *loc.Location
}

func ( s *StringRestrictionSyntax ) LiteralString() string { return s.Str }

type NumRestrictionSyntax struct {
    IsNeg bool
    Num *lexer.NumericToken

    // beginning of numeric expression, not neccesarily Num (could be +/-)
    Loc *loc.Location 
}

func ( nx *NumRestrictionSyntax ) LiteralString() string {
    numStr := nx.Num.String()
    if nx.IsNeg { numStr = "-" + numStr }
    return numStr
}

type RangeRestrictionSyntax struct {
    LeftClosed bool
    Left RestrictionSyntax
    Right RestrictionSyntax
    RightClosed bool
}

type RegexRestrictionSyntax struct {
    Pat string
    Loc *loc.Location
}
