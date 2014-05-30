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
    opts := &Options{
        Reader: bytes.NewBufferString( s ),
        SourceName: ParseSourceInput,
        IsExternal: true,
    }
    return NewBuilder( New( opts ) )
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

type TypeQuantifier int

const (
    TypeQuantifierNullable = TypeQuantifier( iota )
    TypeQuantifierList
    TypeQuantifierNonEmptyList
)

type quantList []SpecialToken

var quantToks []SpecialToken

func init() {
    quantToks = []SpecialToken{
        SpecialTokenPlus,
        SpecialTokenAsterisk,
        SpecialTokenQuestionMark,
    }
}

func ( sb *Builder ) appendQuantToken( 
    quants []SpecialToken, 
    tn *TokenNode ) ( []SpecialToken, error ) {

    st := tn.SpecialToken()
    if l := len( quants ); l > 0 {
        prev := quants[ l - 1 ]
        if st == SpecialTokenQuestionMark &&
           prev == SpecialTokenQuestionMark {
            msg := "a nullable type cannot itself be made nullable"
            return nil, &ParseError{ msg, tn.Loc }
        }
    }
    quants = append( quants, tn.SpecialToken() ) 
    sb.SetSynthEnd()
    return quants, nil
}

func ( sb *Builder ) expectTypeRefQuantCompleter() ( 
    quantList, error ) {
    quants := make( []SpecialToken, 0, 2 )
    for loop := true; sb.HasTokens() && loop; {
        var tn *TokenNode
        var err error
        if tn, err = sb.PollSpecial( quantToks... ); err == nil {
            if loop = tn != nil; loop {
                if quants, err = sb.appendQuantToken( quants, tn ); err != nil {
                    return nil, err
                }
            }
        } else { return nil, err }
    }
    return quantList( quants ), nil
}

var rangeClosers []SpecialToken

func init() {
    rangeClosers = []SpecialToken{
        SpecialTokenCloseParen, 
        SpecialTokenCloseBracket,
    }
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

func ( sb *Builder ) pollPointerDepth() ( int, error ) {
    res := 0
    for {
        tok, err := sb.PollSpecial( SpecialTokenAmpersand )
        if err != nil { return 0, err }
        if tok == nil { break }
        res++
    }
    return res, nil
}

func ( sb *Builder ) expectRestrictionSyntax() (
    RestrictionSyntax, error ) {
    pmTok, neg, err := sb.PollPlusMinus()
    if err != nil { return nil, err }
    var tn *TokenNode
    if tn, err = sb.nextToken(); err != nil { return nil, err }
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
    leftLoc := tn.Loc
    rg = &RangeRestrictionSyntax{ LeftClosed: leftClosed }
    if err = sb.completeRangeLeftSide( rg ); err != nil { return }
    if err = sb.SkipWsOrComments(); err != nil { return }
    var rightLoc *Location
    if rightLoc, err = sb.completeRangeRightSide( rg ); err != nil { return }
    return rg, sb.checkRange( rg, leftLoc, rightLoc )
}

func ( sb *Builder ) pollTypeRefRestriction() (
    sx RestrictionSyntax, err error ) {
    if err = sb.SkipWsOrComments(); err != nil { return }
    var tn *TokenNode
    if tn, err = sb.PollSpecial( SpecialTokenTilde ); 
        err == nil && tn != nil {
        if err = sb.SkipWsOrComments(); err != nil { return }
        if tn, err = sb.nextToken(); err != nil { return }
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

type CompletableTypeReference struct {
    ErrLoc *Location
    Name mg.TypeName
    Restriction RestrictionSyntax
    ptrDepth int
    quants quantList
}

func ( ctr *CompletableTypeReference ) completionError( err error ) error {
    return &ParseError{ err.Error(), ctr.ErrLoc }
}

type TypeCompleter interface {
    AsListType( typ interface{}, allowsEmpty bool ) ( interface{}, error )
    AsNullableType( typ interface{} ) ( interface{}, error )
    AsPointerType( typ interface{} ) ( interface{}, error )
}

func ( ctr *CompletableTypeReference ) CompleteType( 
    typ interface{}, tc TypeCompleter ) ( interface{}, error ) {

    var err error
    for i := 0; i < ctr.ptrDepth; i++ { 
        if typ, err = tc.AsPointerType( typ ); err != nil { 
            return nil, ctr.completionError( err )
        }
    }
    for _, quant := range ctr.quants {
        switch quant {
        case SpecialTokenPlus: typ, err = tc.AsListType( typ, false )
        case SpecialTokenAsterisk: typ, err = tc.AsListType( typ, true )
        case SpecialTokenQuestionMark: typ, err = tc.AsNullableType( typ )
        default: panic( fmt.Errorf( "Unexpected quant: %s", quant ) )
        }
        if err != nil { return nil, ctr.completionError( err ) }
    }
    return typ, nil
}

// In addition to returning a type ref, this method will, via its helper
// methods, ensure that sb.SetSynthEnd() is called after the token that
// completes the type reference, effectively making type references capable of
// ending statements with a synthetic end token, just as lexer does with
// numbers, strings, etc.
func ( sb *Builder ) ExpectTypeReference(
    verDefl *mg.Identifier ) ( ref *CompletableTypeReference, 
                               l *Location,
                               err error ) {

    tmp := &CompletableTypeReference{ ErrLoc: sb.Location() }
    if tmp.ptrDepth, err = sb.pollPointerDepth(); err != nil { return }
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
    LeftClosed bool
    Left RestrictionSyntax
    Right RestrictionSyntax
    RightClosed bool
}

type RegexRestrictionSyntax struct {
    Pat string
    Loc *Location
}
