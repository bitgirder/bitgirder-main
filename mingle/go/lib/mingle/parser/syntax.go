package parser

import(
//    "log"
    mg "mingle"
    "bytes"
    "io"
    "fmt"
    "strings"
)

type TokenNode struct {
    Token Token
    Loc *Location
}

func ( tn *TokenNode ) Identifier() *mg.Identifier { 
    return tn.Token.( *mg.Identifier )
}

func ( tn *TokenNode ) DeclaredTypeName() *mg.DeclaredTypeName {
    return tn.Token.( *mg.DeclaredTypeName )
}

func ( tn *TokenNode ) Number() *NumericToken {
    return tn.Token.( *NumericToken )
}

func ( tn *TokenNode ) SpecialToken() SpecialToken {
    return tn.Token.( SpecialToken )
}

func ( tn *TokenNode ) IsSpecial( s SpecialToken ) bool {
    if spec, ok := tn.Token.( SpecialToken ); ok { return spec == s }
    return false
}

func ( tn *TokenNode ) IsKeyword( k Keyword ) bool {
    if kwd, ok := tn.Token.( Keyword ); ok { return kwd == k }
    return false
}

type Builder struct {
    lx *Lexer
}

func NewBuilder( lx *Lexer ) *Builder { return &Builder{ lx: lx } }

func newSyntaxBuilderExt( s string ) *Builder {
    opts := &LexerOptions{
        Reader: bytes.NewBufferString( s ),
        SourceName: ParseSourceInput,
        IsExternal: true,
    }
    return NewBuilder( NewLexer( opts ) )
}

type tokenExpectType int

const (
    tokenExpectAny = tokenExpectType( iota )
    tokenExpectIdentifier
    tokenExpectDeclTypeName
    tokenExpectNumber
)

func ( sb *Builder ) SetSynthEnd() { sb.lx.SetSynthEnd() }

func ( sb * Builder ) readToken(
    et tokenExpectType ) ( tn *TokenNode, err error ) {
    var lxTok Token
    var lc *Location
    switch et {
    case tokenExpectAny: lxTok, lc, err = sb.lx.ReadToken()
    case tokenExpectIdentifier: lxTok, lc, err = sb.lx.ReadIdentifier()
    case tokenExpectDeclTypeName: lxTok, lc, err = sb.lx.ReadDeclaredTypeName()
    case tokenExpectNumber: lxTok, lc, err = sb.lx.ReadNumber()
    default: panic( fmt.Errorf( "Unexpected token expect type: %v", et ) )
    }
    if lxTok != nil { tn = &TokenNode{ lxTok, lc } }
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

func ( sb *Builder ) Location() *Location { return sb.lx.Location() }

func ( sb *Builder ) ParseError( msg string, args ...interface{} ) error {
    return &ParseError{ fmt.Sprintf( msg, args... ), sb.Location() }
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

func ( sb *Builder ) nextTokenNode() ( *TokenNode, error ) {
    return sb.implNextToken( "", tokenExpectAny )
}

func ( sb *Builder ) mustNextTokenNode() *TokenNode {
    tn, err := sb.nextTokenNode()
    if err == nil { return tn }
    panic( err )
}

func ( sb *Builder ) MustNextToken() Token {
    return sb.mustNextTokenNode().Token
}

func ( sb *Builder ) peekUnexpectedTokenErrorNode() *TokenNode {
    tn, err := sb.PeekToken()
    if err != nil && err != io.EOF {
        errMsg := "syntax: Error in PeekToken() during error construction: %s"
        errLoc := sb.lx.Location()
        panic( &ParseError{ fmt.Sprintf( errMsg, err ), errLoc } )
    }
    return tn
}

// tn nil means the error is at next token loc (or END)
func ( sb *Builder ) ErrorTokenUnexpected( 
    expctDesc string, tn *TokenNode ) error {
    if tn == nil { tn = sb.peekUnexpectedTokenErrorNode() }
    var tokStr string
    var lc *Location
    if tn == nil {
        tokStr, lc = "END", sb.lx.Location()
    } else { tokStr, lc = sb.errorStringForToken( tn.Token ), tn.Loc }
    var msg string
    if expctDesc == "" {
        msg = fmt.Sprintf( "Unexpected token: %s", tokStr )
    } else { 
        msg = fmt.Sprintf( "Expected %s but found: %s", expctDesc, tokStr ) 
    }
    return &ParseError{ msg, lc }
}

func IsSpecial( tok Token, spec SpecialToken ) bool {
    act, ok := tok.( SpecialToken )
    return ok && act == spec
}

func ( sb *Builder ) PollSpecial( 
    l ...SpecialToken ) ( *TokenNode, error ) {
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
    l ...SpecialToken ) ( tn *TokenNode, err error ) {
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
    case WhitespaceToken: return fmt.Sprintf( "%#v", string( v ) )
    case *NumericToken: return v.String()
    }
    return fmt.Sprintf( "%s", tok )
}

func ( sb *Builder ) CheckTrailingToken() error {
    var tn *TokenNode
    var err error
    if tn, err = sb.PollSpecial( SpecialTokenSynthEnd ); err != nil { 
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
            case CommentToken, WhitespaceToken: sb.MustNextToken()
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
    case tokenExpectIdentifier: _, ok = tk.( *mg.Identifier )
    case tokenExpectDeclTypeName: _, ok = tk.( *mg.DeclaredTypeName )
    case tokenExpectNumber: _, ok = tk.( *NumericToken )
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
    ns *mg.Namespace ) ( sawAsp bool, l *Location, err error ) {
    for { 
        var tn *TokenNode
        if tn, err = sb.ExpectIdentifier(); err == nil {
            if l == nil { l = tn.Loc }
            ns.Parts = append( ns.Parts, tn.Identifier() ) 
        } else { return }
        if tn, err = sb.PeekToken(); err == nil && tn != nil {
            switch tn.Token {
            case SpecialTokenColon: sb.MustNextToken()
            case SpecialTokenAsperand: 
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
    verDefl *mg.Identifier ) ( ns *mg.Namespace, l *Location, err error ) {

    ns = &mg.Namespace{ Parts: make( []*mg.Identifier, 0, 4 ) }
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

func ( sb *Builder ) ExpectQualifiedTypeName( 
    verDefl *mg.Identifier ) ( qn *mg.QualifiedTypeName, 
                               l *Location, 
                               err error ) {

    qn = new( mg.QualifiedTypeName )
    if qn.Namespace, l, err = sb.ExpectNamespace( verDefl ); err != nil { 
        return 
    }
    var tn *TokenNode
    if tn, err = sb.PollSpecial( SpecialTokenForwardSlash ); err == nil {
        if tn == nil { err = sb.ErrorTokenUnexpected( "type path", nil ) }
    } 
    if err != nil { return }
    if tn, err = sb.ExpectDeclaredTypeName(); err != nil { return }
    qn.Name = tn.DeclaredTypeName()
    return
}

func ( sb *Builder ) ExpectTypeName( 
    verDefl *mg.Identifier ) ( nm mg.TypeName, l *Location, err error ) {
    var tn *TokenNode
    if tn, err = sb.PeekToken(); err == nil && tn == nil {
        err = sb.errorUnexpectedEnd()
    } 
    if err != nil { return }
    switch tn.Token.( type ) {
    case *mg.Identifier: nm, l, err = sb.ExpectQualifiedTypeName( verDefl )
    case *mg.DeclaredTypeName: 
        if tn, err = sb.ExpectDeclaredTypeName(); err != nil { return }
        nm, l = tn.DeclaredTypeName(), tn.Loc
    default: 
        err = sb.ErrorTokenUnexpected( "identifier or declared type name", tn )
    }
    return
}

type AtomicTypeExpression struct {
    Name mg.TypeName
    NameLoc *Location
    Restriction RestrictionSyntax
}

type PointerTypeExpression struct {
    Loc *Location
    Expression interface{}
}

type ListTypeExpression struct {
    Loc *Location
    Expression interface{}
    AllowsEmpty bool
}

type NullableTypeExpression struct {
    Loc *Location
    Expression interface{}
}

func atomicExpressionIn( e interface{} ) *AtomicTypeExpression {
    switch v := e.( type ) {
    case *AtomicTypeExpression: return v
    case *ListTypeExpression: return atomicExpressionIn( v.Expression )
    case *NullableTypeExpression: return atomicExpressionIn( v.Expression )
    case *PointerTypeExpression: return atomicExpressionIn( v.Expression )
    }
    panic( libErrorf( "unhandled expression: %T", e ) )
}

func headLocationOfExpression( e interface{} ) *Location {
    switch v := e.( type ) {
    case *AtomicTypeExpression: return v.NameLoc
    case *ListTypeExpression: return headLocationOfExpression( v.Expression )
    case *NullableTypeExpression: 
        return headLocationOfExpression( v.Expression )
    case *PointerTypeExpression: return v.Loc
    }
    panic( libErrorf( "unhandled type exp: %T", e ) )
}

type CompletableTypeReference struct {
    Expression interface{}
}

func ( t *CompletableTypeReference ) Location() *Location {
    return headLocationOfExpression( t.Expression )
}

type TypeCompleter interface {

    CompleteBaseType ( 
        mg.TypeName, 
        RestrictionSyntax, 
        *Location ) ( mg.TypeReference, bool, error )
}

var quantToks []SpecialToken

func init() {
    quantToks = []SpecialToken{
        SpecialTokenPlus,
        SpecialTokenAsterisk,
        SpecialTokenQuestionMark,
    }
}

var rangeClosers []SpecialToken

func init() {
    rangeClosers = []SpecialToken{
        SpecialTokenCloseParen, 
        SpecialTokenCloseBracket,
    }
}

// next token is known to be SpecialTokenAmpersand
func ( sb *Builder ) expectPointerTypeExpression(
    verDefl *mg.Identifier ) ( *PointerTypeExpression, error ) {

    tn := sb.mustNextTokenNode()
    if err := sb.CheckUnexpectedEnd(); err != nil { return nil, err }
    e, err := sb.expectTypeExpressionBase( verDefl )
    if err != nil { return nil, err }
    return &PointerTypeExpression{ tn.Loc, e }, nil
}

// removes and returns next token if it is '+' or '-', returning true if it was
// '-', false if it was anything other than '-' (not necessarily '+'); if '+' or
// '-' is seen
func ( sb *Builder ) PollPlusMinus() ( *TokenNode, bool, error ) {
    tn, err := sb.PeekToken()
    if err == nil && tn != nil {
        if spec, ok := tn.Token.( SpecialToken ); ok {
            if spec == SpecialTokenMinus || 
               spec == SpecialTokenPlus {
                sb.MustNextToken() // consume it
                return tn, spec == SpecialTokenMinus, nil
            }
        }
    }
    return nil, false, err
}

func ( sb *Builder ) pollPointerDepth() ( int, *Location, error ) {
    depth := 0
    var l *Location
    for {
        tok, err := sb.PollSpecial( SpecialTokenAmpersand )
        if err != nil { return 0, nil, err }
        if tok == nil { break }
        if l == nil { l = tok.Loc }
        depth++
    }
    return depth, l, nil
}

func ( sb *Builder ) expectRestrictionSyntax() (
    RestrictionSyntax, error ) {
    pmTok, neg, err := sb.PollPlusMinus()
    if err != nil { return nil, err }
    var tn *TokenNode
    if tn, err = sb.nextTokenNode(); err != nil { return nil, err }
    if tn == nil { return nil, sb.errorUnexpectedEnd() }
    if s, ok := tn.Token.( StringToken ); ok { 
        // string ok unless we want num
        if pmTok != nil { return nil, sb.ErrorTokenUnexpected( "number", tn ) }
        return &StringRestrictionSyntax{ string( s ), tn.Loc }, nil
    }
    if num, ok := tn.Token.( *NumericToken ); ok {
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
    if tn, err = sb.PollSpecial( SpecialTokenComma ); err != nil { 
        return 
    }
    if tn == nil {
        if rg.Left, err = sb.expectRestrictionSyntax(); err != nil { return }
        if err = sb.SkipWsOrComments(); err != nil { return }
        if _, err = sb.ExpectSpecial( SpecialTokenComma ); err != nil { 
            return 
        }
    }
    return
}

func ( sb *Builder ) checkRange( 
    rg *RangeRestrictionSyntax, leftLoc, rightLoc *Location ) error {
    errLf := rg.Left == nil && rg.LeftClosed
    errRt := rg.Right == nil && rg.RightClosed
    msg := ""
    var errLoc *Location
    if errLf {
        if errRt { 
            msg, errLoc = "Infinite range must be open", leftLoc
        } else { msg, errLoc = "Infinite low range must be open", leftLoc }
    } else if errRt { 
        msg, errLoc = "Infinite high range must be open", rightLoc 
    }
    if msg != "" { return &ParseError{ msg, errLoc } }
    return nil
}

func ( sb *Builder ) completeRangeRightSide( 
    rg *RangeRestrictionSyntax ) ( rightLoc *Location, err error ) {
    var tn *TokenNode
    if tn, err = sb.PollSpecial( rangeClosers... ); err != nil { return }
    if tn == nil { 
        if rg.Right, err = sb.expectRestrictionSyntax(); err != nil { return }
        if err = sb.SkipWsOrComments(); err != nil { return }
        if tn, err = sb.ExpectSpecial( rangeClosers... ); err != nil { return }
    }
    rightLoc = tn.Loc
    closeTok := tn.SpecialToken() // either set before or inside branch above
    rg.RightClosed = closeTok == SpecialTokenCloseBracket
    return
}

func ( sb *Builder ) completeRangeRestriction( 
    tn *TokenNode ) ( rg *RangeRestrictionSyntax, err error ) {

    leftClosed := tn.SpecialToken() == SpecialTokenOpenBracket
    rg = &RangeRestrictionSyntax{ Loc: tn.Loc, LeftClosed: leftClosed }
    if err = sb.completeRangeLeftSide( rg ); err != nil { return }
    if err = sb.SkipWsOrComments(); err != nil { return }
    var rightLoc *Location
    if rightLoc, err = sb.completeRangeRightSide( rg ); err != nil { return }
    return rg, sb.checkRange( rg, rg.Loc, rightLoc )
}

func ( sb *Builder ) pollTypeRefRestriction() (
    sx RestrictionSyntax, err error ) {

    if err = sb.SkipWsOrComments(); err != nil { return }
    var tn *TokenNode
    if tn, err = sb.PollSpecial( SpecialTokenTilde ); 
        err == nil && tn != nil {
        if err = sb.SkipWsOrComments(); err != nil { return }
        if tn, err = sb.nextTokenNode(); err != nil { return }
        if IsSpecial( tn.Token, SpecialTokenOpenParen ) || 
           IsSpecial( tn.Token, SpecialTokenOpenBracket ) {
            sx, err = sb.completeRangeRestriction( tn )
            sb.SetSynthEnd()
        } else if s, ok := tn.Token.( StringToken ); ok {
            sx = &RegexRestrictionSyntax{ string( s ), tn.Loc }
            sb.SetSynthEnd()
        } else {
            err = sb.ErrorTokenUnexpected( "type restriction", tn )
        }
    } else { sb.SetSynthEnd() }
    return
}

func canStartAtomicType( tn *TokenNode ) bool {
    switch tn.Token.( type ) {
    case *mg.Identifier, *mg.DeclaredTypeName: return true
    }
    return false
}

func ( sb *Builder ) expectAtomicTypeExpression(
    verDefl *mg.Identifier ) ( *AtomicTypeExpression, error ) {

    nm, nmLoc, err := sb.ExpectTypeName( verDefl )
    if err != nil { return nil, err }
    res := &AtomicTypeExpression{ Name: nm, NameLoc: nmLoc }
    if res.Restriction, err = sb.pollTypeRefRestriction(); err != nil { 
        return nil, err
    }
    return res, nil
}

func ( sb *Builder ) pollCloseParen() ( bool, error ) {
    if err := sb.SkipWsOrComments(); err != nil { return false, err }
    if ! sb.HasTokens() { return false, nil }
    tn, err := sb.PeekToken()
    if err != nil { return false, err }
    res := tn.IsSpecial( SpecialTokenCloseParen )
    if res { sb.mustNextTokenNode() }
    return res, nil
}

func ( sb *Builder ) expectGroupedTypeExpression(
    verDefl *mg.Identifier ) ( interface{}, error ) {

    openNode := sb.mustNextTokenNode() // '('
    if err := sb.SkipWsOrComments(); err != nil { return nil, err }
    res, err := sb.pollTypeExpression( verDefl )
    if err != nil { return nil, err }
    ok, err := sb.pollCloseParen()
    if err != nil { return nil, err }
    if ok { return res, nil }
    return nil, &ParseError{ `Unmatched "("`, openNode.Loc }
}

func ( sb *Builder ) expectTypeExpressionBase(
    verDefl *mg.Identifier ) ( interface{}, error ) {

    tn, err := sb.PeekToken()
    if err != nil { return nil, err }
    if tn.IsSpecial( SpecialTokenAmpersand ) {
        return sb.expectPointerTypeExpression( verDefl )
    }
    if canStartAtomicType( tn ) { 
        return sb.expectAtomicTypeExpression( verDefl ) 
    }
    if tn.IsSpecial( SpecialTokenOpenParen ) {
        return sb.expectGroupedTypeExpression( verDefl )
    }
    return nil, sb.ErrorTokenUnexpected( "type reference", tn )
}

func ( sb *Builder ) applyTypeQuantifier( 
    e interface{}, tn *TokenNode ) ( interface{}, error ) {

    var err error
    switch tn.SpecialToken() {
    case SpecialTokenAsterisk: e = &ListTypeExpression{ tn.Loc, e, true }
    case SpecialTokenPlus: e = &ListTypeExpression{ tn.Loc, e, false }
    case SpecialTokenQuestionMark:
        if _, ok := e.( *NullableTypeExpression ); ok {
            msg := "a nullable type cannot itself be made nullable"
            err = &ParseError{ msg, tn.Loc }
        } else { e = &NullableTypeExpression{ tn.Loc, e } }
    default: err = sb.ErrorTokenUnexpected( "type quantifier", tn )
    }
    if err != nil { return nil, err }
    sb.SetSynthEnd()
    return e, nil
}

func ( sb *Builder ) applyTypeQuantifiers( 
    e interface{} ) ( interface{}, error ) {

    for loop := true; sb.HasTokens() && loop; {
        tn, err := sb.PollSpecial( quantToks... )
        if err != nil { return nil, err }
        if loop = tn != nil; ! loop { break }
        e, err = sb.applyTypeQuantifier( e, tn )
        if err != nil { return nil, err }
    }
    return e, nil
}

func ( sb *Builder ) pollTypeExpression( 
    verDefl *mg.Identifier ) ( interface{}, error ) {

    if err := sb.SkipWsOrComments(); err != nil { return nil, err }
    base, err := sb.expectTypeExpressionBase( verDefl );
    if err != nil { return nil, err }
    return sb.applyTypeQuantifiers( base )
}

// In addition to returning a type ref, this method will, via its helper
// methods, ensure that sb.SetSynthEnd() is called after the token that
// completes the type reference, effectively making type references capable of
// ending statements with a synthetic end token, just as lexer does with
// numbers, strings, etc.
func ( sb *Builder ) ExpectTypeReference(
    verDefl *mg.Identifier ) ( ref *CompletableTypeReference, err error ) {

    ref = &CompletableTypeReference{}
    ref.Expression, err = sb.pollTypeExpression( verDefl )
    return 
}

func applyListTypeCompletion(
    e *ListTypeExpression, 
    atRepl mg.TypeReference ) ( mg.TypeReference, error ) {
        
    et, err := applyTypeCompletion( e.Expression, atRepl )
    if err != nil { return nil, err }
    lt := &mg.ListTypeReference{ ElementType: et, AllowsEmpty: e.AllowsEmpty }
    return lt, nil
}

func applyNullableTypeCompletion( 
    e *NullableTypeExpression, 
    atRepl mg.TypeReference ) ( mg.TypeReference, error ) {

    nt, err := applyTypeCompletion( e.Expression, atRepl )
    if err != nil { return nil, err }
    if mg.IsNullableType( nt ) {
        return mg.MustNullableTypeReference( nt ), nil
    } 
    return nil, mg.NewNullableTypeError( nt )
}

func applyPointerTypeCompletion( 
    e *PointerTypeExpression, 
    atRepl mg.TypeReference ) ( mg.TypeReference, error ) {

    pt, err := applyTypeCompletion( e.Expression, atRepl )
    if err != nil { return nil, err }
    return mg.NewPointerTypeReference( pt ), nil
}

func applyTypeCompletion( 
    e interface{}, atRepl mg.TypeReference ) ( mg.TypeReference, error ) {

    switch v := e.( type ) {
    case *AtomicTypeExpression: return atRepl, nil
    case *ListTypeExpression: return applyListTypeCompletion( v, atRepl )
    case *NullableTypeExpression: 
        return applyNullableTypeCompletion( v, atRepl )
    case *PointerTypeExpression: return applyPointerTypeCompletion( v, atRepl )
    }
    panic( libErrorf( "unhandled type exp: %T", e ) )
}

func ( t *CompletableTypeReference ) CompleteType( 
    tc TypeCompleter ) ( mg.TypeReference, error ) {

    at := atomicExpressionIn( t.Expression )
    res, ok, err := tc.CompleteBaseType( at.Name, at.Restriction, at.NameLoc )
    if ! ( ok && err == nil ) { return nil, err }
    return applyTypeCompletion( t.Expression, res )
}

// meant for string/num restriction only right now
type RestrictionSyntax interface{}

type LiteralStringer interface {
    LiteralString() string
}

type StringRestrictionSyntax struct {
    Str string
    Loc *Location
}

func ( s *StringRestrictionSyntax ) LiteralString() string { return s.Str }

type NumRestrictionSyntax struct {
    IsNeg bool
    Num *NumericToken

    // beginning of numeric expression, not neccesarily Num (could be +/-)
    Loc *Location 
}

func ( nx *NumRestrictionSyntax ) LiteralString() string {
    numStr := nx.Num.String()
    if nx.IsNeg { numStr = "-" + numStr }
    return numStr
}

type RangeRestrictionSyntax struct {
    Loc *Location
    LeftClosed bool
    Left RestrictionSyntax
    Right RestrictionSyntax
    RightClosed bool
}

type RegexRestrictionSyntax struct {
    Pat string
    Loc *Location
}
