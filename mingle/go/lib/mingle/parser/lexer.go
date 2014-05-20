package parser

import (
    mg "mingle"
//    "log"
    "io"
    "strconv"
    "container/list"
    "fmt"
    "bufio"
    "errors"
    "bytes"
    "unicode"
    "unicode/utf16"
)

const replChar = rune( 0xfffd )

type Token interface{}

type idPart string

type DeclaredTypeName string

type Keyword string

const (
    KeywordAlias = Keyword( "alias" )
    KeywordDefault = Keyword( "default" )
    KeywordEnum = Keyword( "enum" )
    KeywordFalse = Keyword( "false" )
    KeywordFunc = Keyword( "func" )
    KeywordImport = Keyword( "import" )
    KeywordNamespace = Keyword( "namespace" )
    KeywordOp = Keyword( "op" )
    KeywordPrototype = Keyword( "prototype" )
    KeywordReturn = Keyword( "return" )
    KeywordService = Keyword( "service" )
    KeywordStruct = Keyword( "struct" )
    KeywordThrows = Keyword( "throws" )
    KeywordTrue = Keyword( "true" )
)

var kwdMap map[ string ]Keyword

func init() {
    kwdMap = make( map[ string ]Keyword )
    kwdMap[ "alias" ] = KeywordAlias
    kwdMap[ "default" ] = KeywordDefault
    kwdMap[ "enum" ] = KeywordEnum
    kwdMap[ "false" ] = KeywordFalse
    kwdMap[ "func" ] = KeywordFunc
    kwdMap[ "import" ] = KeywordImport
    kwdMap[ "namespace" ] = KeywordNamespace
    kwdMap[ "op" ] = KeywordOp
    kwdMap[ "prototype" ] = KeywordPrototype
    kwdMap[ "return" ] = KeywordReturn
    kwdMap[ "service" ] = KeywordService
    kwdMap[ "struct" ] = KeywordStruct
    kwdMap[ "throws" ] = KeywordThrows
    kwdMap[ "true" ] = KeywordTrue
}

const runeReplChar = rune( '\ufffd' )

func isLexErr( err error ) bool { return err != nil && err != io.EOF }

type stackElt struct {
    tok Token
    lc *Location
    eol bool
}

type unreadElt struct {
    elt stackElt
    synthLoc *Location
}

// May at some point add Lexer.SetSourceName( string ) to allow callers to set
// value of loc.Source
type Lexer struct {
    reader *bufio.Reader
    SourceName string
    line int
    col int
    newlineUnreadCol int
    isExternal bool
    strip bool
    sawEof bool
    synthLoc *Location
    unread *unreadElt
    stack [ 2 ]stackElt
    stackLen int
}

func ( lx *Lexer ) stackEmpty() bool { return lx.stackLen == 0 }

func ( lx *Lexer ) push( e stackElt ) {
    if lx.stackLen < 2 { 
        lx.stack[ lx.stackLen ] = e
        lx.stackLen++
    } else { 
        panic( fmt.Errorf( "lexer: Invalid stack len: %d", lx.stackLen ) ) 
    }
}

func ( lx *Lexer ) stackAccess() int {
    if lx.stackEmpty() { panic( "lexer: stack is empty" ) }
    return lx.stackLen - 1
}

func ( lx *Lexer ) peek() stackElt { return lx.stack[ lx.stackAccess() ] }

func ( lx *Lexer ) pop() stackElt {
    idx := lx.stackAccess()
    res := lx.stack[ idx ]
    // zero the previous element to simplify debugging (hard to tell a stale
    // element from a live one)
    lx.stack[ idx ] = stackElt{}
    lx.stackLen--
    return res
}

// Calling this method invalidates any call to UnreadToken() until the next read
// operation, and will mark the current location and the previous read as if it
// were the read of a token which could imply a statement end.
func ( lx *Lexer ) SetSynthEnd() { 
    lx.unread = nil
    if lx.stackEmpty() {
        lx.synthLoc = lx.makeLocation()
    } else {
        elt := lx.peek()
        if elt.tok != SpecialTokenSynthEnd {
            lx.synthLoc = lx.makeLocation()
            lx.updateSynthLoc( elt.tok ) // clears it if meaningless
        }
    }
}

var unreadNoValErr = errors.New( "lexer: No value to unread" )

func ( lx *Lexer ) UnreadToken() {
    if ue := lx.unread; ue == nil {
        panic( unreadNoValErr )
    } else {
        lx.synthLoc = ue.synthLoc
        lx.push( ue.elt )
        lx.unread = nil
    }
}

func ( lx *Lexer ) makeLocation() *Location {
    return &Location{ 
        Source: lx.SourceName, 
        Line: lx.line, 
        Col: lx.col,
    }
}

func ( lx *Lexer ) Location() *Location { 
    if lx.stackEmpty() { return lx.makeLocation() }
    return lx.peek().lc
}

func ( lx *Lexer ) makeParseError( 
    adj int, fmtStr string, args ...interface{} ) *ParseError {
    l := lx.makeLocation()
    l.Col += adj
    return &ParseError{ fmt.Sprintf( fmtStr, args... ), l }
}

// returns an error pointed at the rune just read (1 behind current position)
func ( lx *Lexer ) prevError( 
    fmtStr string, args ...interface{} ) error {
    return lx.makeParseError( -1, fmtStr, args... )
}

// returns an error at the current read location (the location of the next rune
// to be returned by readRune())
func ( lx *Lexer ) parseError( fmtStr string, args ...interface{} ) error {
    return lx.makeParseError( 0, fmtStr, args... )
}

func ( lx *Lexer ) readRune() ( rune, error ) {
    r, _, err := lx.reader.ReadRune()
    if err == nil {
        if r == '\n' {
            lx.line++ 
            lx.newlineUnreadCol = lx.col
            lx.col = 1
        } else { 
            lx.newlineUnreadCol = -1
            lx.col++ 
        }
    }
    return r, err
}

func ( lx *Lexer ) unreadRune() {
    if err := lx.reader.UnreadRune(); err != nil { panic( err ) }
    if lx.newlineUnreadCol < 0 {
        lx.col--
    } else {
        lx.line--
        lx.col = lx.newlineUnreadCol
        lx.newlineUnreadCol = -1
    }
}

// does not advance lexer position
func ( lx *Lexer ) peekRune() ( rune, error ) {
    r, err := lx.readRune()
    if err == nil { lx.unreadRune() }
    return r, err
}

// useful after a call to unreadRune() when caller expects a rune to be present
func ( lx *Lexer ) mustRune( expct rune ) rune {
    r, err := lx.readRune()
    if err != nil { panic( err ) }
    if expct > 0 && r != expct {
        panic( lx.parseError( "Expected rune %U but got %U", expct, r ) )
    }
    return r
}

func isIdentLower( r rune ) bool { return r >= 'a' && r <= 'z' }
func isIdentCap( r rune ) bool { return r >= 'A' && r <= 'Z' }
func isIdentDigit( r rune ) bool { return r >= '0' && r <= '9' }

func isIdentTailChar( r rune ) bool { 
    return isIdentLower( r ) || isIdentDigit( r )
}

type idSeparation int

const (
    idSepExternal = idSeparation( iota )
    idSepHyphen
    idSepUnderscore
    idSepCap
)

func ( lx *Lexer ) initialidPartSep() idSeparation {
    if lx.isExternal { return idSepExternal }
    return idSepCap
}

func ( lx *Lexer ) errorUnrecognizedIdentRune( r rune ) error {
    return lx.prevError( "Invalid id rune: %q (%U)", string( r ), r )
}

func ( lx *Lexer ) readidPartFirstRune( 
    idSep idSeparation, partNum int ) ( r rune, err error ) {
    if partNum > 0 {
        switch idSep {
        case idSepHyphen: lx.mustRune( '-' )
        case idSepUnderscore: lx.mustRune( '_' )
        }
    }
    switch r, err = lx.readRune(); {
    case err != nil: 
        if err == io.EOF { 
            msg := "Empty identifier"
            if partNum > 0 { msg = "Empty identifier part" }
            err = lx.parseError( msg )
        }
    case isIdentLower( r ): return
    case isIdentCap( r ) && idSep == idSepCap: r = unicode.ToLower( r )
    default: 
        err = lx.prevError( 
                "Illegal start of identifier part: %q (%U)", string( r ), r )
    }
    return
}

func ( lx *Lexer ) readNonIdentTailChar( 
    r rune, 
    idSep idSeparation ) ( idSep2 idSeparation, 
                           partDone bool,
                           idDone bool, 
                           err error ) {
    switch {
    case r == '-' && ( idSep == idSepHyphen || idSep == idSepExternal ):
        idSep2, partDone = idSepHyphen, true
    case r == '_' && ( idSep == idSepUnderscore || idSep == idSepExternal ):
        idSep2, partDone = idSepUnderscore, true
    case isIdentCap( r ) && ( idSep == idSepCap || idSep == idSepExternal ):
        idSep2, partDone = idSepCap, true
    case isSpecialTokChar( r ) || isWhitespace( r ): 
        partDone, idDone = true, true
    default: err = lx.errorUnrecognizedIdentRune( r )
    }
    if err == nil { lx.unreadRune() }
    return
} 

func ( lx *Lexer ) accumulateidPart(
    idSep idSeparation, partNum int ) ( part idPart, 
                                        idSep2 idSeparation, 
                                        idDone bool, 
                                        err error ) {
    buf := bytes.Buffer{}
    var r rune
    if r, err = lx.readidPartFirstRune( idSep, partNum ); err == nil { 
        buf.WriteRune( r )
    } else { return }
    for part == "" && err == nil {
        var partDone bool
        switch r, err = lx.readRune(); {
        case isLexErr( err ): {}
        case err == io.EOF: partDone, idDone = true, true
        case isIdentTailChar( r ): buf.WriteRune( r )
        default: 
            idSep2, partDone, idDone, err = lx.readNonIdentTailChar( r, idSep )
        }
        if partDone { part = idPart( buf.String() ) }
    }
    return
}

func ( lx *Lexer ) parseIdentifier() ( id *mg.Identifier, err error ) {
    idSep := lx.initialidPartSep()
    parts := make( []string, 0, 3 )
    for id == nil && err == nil {
        var part idPart
        var idDone bool
        if part, idSep, idDone, err = 
            lx.accumulateidPart( idSep, len( parts ) ); ! isLexErr( err ) {
            parts = append( parts, string( part ) )
            if idDone { id = mg.NewIdentifierUnsafe( parts ) }
        }
    }
    return
}

func isKeyword( id *mg.Identifier ) ( Keyword, bool ) {
    kwd, ok := kwdMap[ id.ExternalForm() ]
    return kwd, ok
}

func ( lx *Lexer ) readIdentifier() ( id Token, err error ) {
    if id, err = lx.parseIdentifier(); isLexErr( err ) { return }
    if ( ! lx.isExternal ) { 
        if kwd, ok := isKeyword( id.( *mg.Identifier ) ); ok { id = kwd } 
    }
    return 
}

func isWhitespace( r rune ) bool {
    return r == ' ' || r == '\t' || r == '\r' || r == '\n'
}

type WhitespaceToken []byte

func ( ws WhitespaceToken ) hasNewline() bool {
    return bytes.IndexRune( []byte( ws ), '\n' ) >= 0
}

func ( lx *Lexer ) readWhitespace() ( tok WhitespaceToken, err error ) {
    buf := &bytes.Buffer{}
    for err == nil {
        var r rune
        r, err = lx.readRune()
        if ! isLexErr( err ) {
            if isWhitespace( r ) {
                buf.WriteRune( r )
            } else { 
                if err != io.EOF { lx.unreadRune() }
                return WhitespaceToken( buf.Bytes() ), err
            }
        }
    }
    return nil, err
}

var specialTokChars []byte
func init() { specialTokChars = []byte( ":;{}~()[],?<->/.*+@&" ); }

func isSpecialTokChar( r rune ) bool {
    return bytes.IndexRune( specialTokChars, r ) >= 0
}

type SpecialToken string

const SpecialTokenColon = SpecialToken( ":" )
const SpecialTokenOpenBrace = SpecialToken( "{" )
const SpecialTokenCloseBrace = SpecialToken( "}" )
const SpecialTokenSemicolon = SpecialToken( ";" )
const SpecialTokenSynthEnd = SpecialToken( "<;>" )
const SpecialTokenTilde = SpecialToken( "~" )
const SpecialTokenOpenParen = SpecialToken( "(" )
const SpecialTokenCloseParen = SpecialToken( ")" )
const SpecialTokenOpenBracket = SpecialToken( "[" )
const SpecialTokenCloseBracket = SpecialToken( "]" )
const SpecialTokenComma = SpecialToken( "," )
const SpecialTokenQuestionMark = SpecialToken( "?" )
const SpecialTokenReturns = SpecialToken( "->" )
const SpecialTokenMinus = SpecialToken( "-" )
const SpecialTokenForwardSlash = SpecialToken( "/" )
const SpecialTokenPeriod = SpecialToken( "." )
const SpecialTokenAsterisk = SpecialToken( "*" )
const SpecialTokenPlus = SpecialToken( "+" )
const SpecialTokenLessThan = SpecialToken( "<" )
const SpecialTokenGreaterThan = SpecialToken( ">" )
const SpecialTokenAsperand = SpecialToken( "@" )
const SpecialTokenAmpersand = SpecialToken( "&" )

// Note: does not contain SpecialTokenSynthEnd
var allSpecialToks list.List

func init() {
    allSpecialToks.PushFront( SpecialTokenColon )
    allSpecialToks.PushFront( SpecialTokenOpenBrace )
    allSpecialToks.PushFront( SpecialTokenCloseBrace )
    allSpecialToks.PushFront( SpecialTokenSemicolon )
    allSpecialToks.PushFront( SpecialTokenTilde )
    allSpecialToks.PushFront( SpecialTokenOpenParen )
    allSpecialToks.PushFront( SpecialTokenCloseParen )
    allSpecialToks.PushFront( SpecialTokenOpenBracket )
    allSpecialToks.PushFront( SpecialTokenCloseBracket )
    allSpecialToks.PushFront( SpecialTokenComma )
    allSpecialToks.PushFront( SpecialTokenQuestionMark )
    allSpecialToks.PushFront( SpecialTokenReturns )
    allSpecialToks.PushFront( SpecialTokenMinus )
    allSpecialToks.PushFront( SpecialTokenForwardSlash )
    allSpecialToks.PushFront( SpecialTokenPeriod )
    allSpecialToks.PushFront( SpecialTokenAsterisk )
    allSpecialToks.PushFront( SpecialTokenPlus )
    allSpecialToks.PushFront( SpecialTokenLessThan )
    allSpecialToks.PushFront( SpecialTokenGreaterThan )
    allSpecialToks.PushFront( SpecialTokenAsperand )
    allSpecialToks.PushFront( SpecialTokenAmpersand )
}

// returns l1 if err is non-nil (including io.EOF); otherwise returns a new list
// containing only elements of l1 that might still be a match after reading the
// next rune
func ( lx *Lexer ) matchSpecialTokNext( 
    l1 *list.List, i int ) ( *list.List, error ) {
    r, err := lx.readRune()
    if err == io.EOF { return &list.List{}, err }
    if err != nil { return nil, err }
    l2 := &list.List{}
    for e := l1.Front(); e != nil; e = e.Next() {
        st := e.Value.( SpecialToken )
        if len( st ) > i && rune( st[ i ] ) == r { l2.PushFront( st ) }
    }
    return l2, nil
}

func ( lx *Lexer ) matchSpecialToks() ( *list.List, error ) {
    l1 := &allSpecialToks
    var err error
    for i := 0; err == nil && l1.Len() > 0; i++ {
        var l2 *list.List
        l2, err = lx.matchSpecialTokNext( l1, i )
        if ! isLexErr( err ) {
            if l2.Len() == 0 { 
                if err != io.EOF { lx.unreadRune() }
                // remove any potential matches longer than what we matched
                for e := l1.Front(); e != nil; e = e.Next() {
                    if len( e.Value.( SpecialToken ) ) > i { l1.Remove( e ) }
                }
                return l1, nil 
            } else { l1 = l2 }
        }
    }
    return l1, err
}

func ( lx *Lexer ) readSpecialTok() ( Token, error ) {
    m, err := lx.matchSpecialToks()
    if ! isLexErr( err ) {
        switch m.Len() {
        case 0: return nil, lx.parseError( "Unrecognized op or delimiter" )
        case 1: return m.Front().Value, nil
        default: panic( lx.parseError( "Ambiguous op or delimiter" ) )
        }
    }
    return nil, err
}

func isCommentStart( r rune ) bool { return r == '#' }

type CommentToken []byte

func ( lx *Lexer ) readComment() ( tok Token, err error ) {
    buf := bytes.Buffer{}
    lx.mustRune( '#' )
    for loop := true; loop && err == nil; {
        var r rune
        if r, err = lx.readRune(); err == nil {
            buf.WriteRune( r )
            loop = r != '\n'
        }
    }
    if ! isLexErr( err ) { return CommentToken( buf.Bytes() ), err }
    return nil, err
}

func isDeclaredTypeNameStart( r rune ) bool { return r >= 'A' && r <= 'Z' }

func isDeclaredTypeNameTail( r rune ) bool {
    return isDeclaredTypeNameStart( r ) ||
           ( r >= 'a' && r <= 'z' ) ||
           ( r >= '0' && r <= '9' )
}

func ( lx *Lexer ) readDeclaredTypeName() ( tok Token, err error ) {
    buf, r := bytes.Buffer{}, rune( 0 )
    switch r, err = lx.readRune(); {
    case isLexErr( err ): return
    case err == io.EOF: err = lx.parseError( "Empty type name" )
    case isDeclaredTypeNameStart( r ): buf.WriteRune( r )
    default: 
        msg := "Illegal type name start: %q (%U)"
        err = lx.prevError( msg, string( r ), r )
    }
    for tok == nil && err == nil {
        switch r, err = lx.readRune(); {
        case isLexErr( err ): break
        case isDeclaredTypeNameTail( r ): buf.WriteRune( r )
        case err == io.EOF || isWhitespace( r ) || isSpecialTokChar( r ):
            if err == nil { lx.unreadRune() } // not on io.EOF
            tok = DeclaredTypeName( buf.String() )
        default:
            msg := "Illegal type name rune: %q (%U)"
            err = lx.prevError( msg, string( r ), r )
        }
    }
    return 
}

func isStringLiteralStart( r rune ) bool { return r == '"' }

type StringToken string

func ( lx *Lexer ) readEscapedRuneHex() ( rune, error ) {
    buf := make( []rune, 0, 4 )
    for len( buf ) < cap( buf ) {
        if r, err := lx.readRune(); err == nil {
            if unicode.Is( unicode.Hex_Digit, r ) {
                buf = append( buf, r )
            } else {
                msg := "Invalid hex char in escape: %q (%U)"
                return 0, lx.prevError( msg, string( r ), r )
            }
        } else { return 0, err }
    }
    var res rune
    if i, err := strconv.ParseUint( string( buf ), 16, 32 ); err != nil {
        panic( lx.parseError( err.Error() ) )
    } else { res = rune( i ) } 
    return res, nil
}

func ( lx *Lexer ) completeSurrogatePair( high rune ) ( r rune, err error ) {
    var cur rune
    errStr, errAdj := "", 0
    if cur, err = lx.readRune(); err != nil { return }
    if cur == '\\' { 
        if cur, err = lx.readRune(); err != nil { return }
        if cur != 'u' { errStr, errAdj = "\\" + string( cur ), -2 }
    } else { errStr, errAdj = fmt.Sprintf( "%q (%U)", string( cur ), cur ), -1 }
    if errStr == "" {
        var low rune
        if low, err = lx.readEscapedRuneHex(); err != nil { return }
        if r = utf16.DecodeRune( high, low ); r == replChar {
            tmpl := "Invalid surrogate pair \\u%04X\\u%04X"
            err = lx.makeParseError( -12, tmpl, high, low )
        }
    } else { 
        tmpl := "Expected trailing surrogate, found: %s"
        err = lx.makeParseError( errAdj, tmpl, errStr )
    }
    return
}

func ( lx *Lexer ) readEscapedRune() ( r rune, err error ) {
    if r, err = lx.readEscapedRuneHex(); err != nil { return }
    if utf16.IsSurrogate( r ) { r, err = lx.completeSurrogatePair( r ) }
    return
}

func ( lx *Lexer ) readStringEscape() ( rune, error ) {
    r, err := lx.readRune()
    if err != nil { return 0, err }
    switch r {
    case 'n': return '\n', nil
    case 'r': return '\r', nil
    case 't': return '\t', nil
    case 'f': return '\f', nil
    case 'b': return '\b', nil
    case '"': return '"', nil
    case '\\': return '\\', nil
    case 'u': return lx.readEscapedRune()
    }
    return 0, lx.prevError( "Unrecognized escape: \\%s (%U)", string( r ), r )
}

func ( lx *Lexer ) appendStringRune( r rune, buf *bytes.Buffer ) ( err error ) {
    switch {
    case r == '\\':
        if r, err = lx.readStringEscape(); err != nil { return }
    case r >= rune( 0 ) && r < rune( 32 ):
        lx.unreadRune()
        msg := "Invalid control character in string literal: %q (%U)"
        return lx.parseError( msg, string( r ), r )
    }
    buf.WriteRune( r )
    return nil
}

func ( lx *Lexer ) readStringLiteral() ( tok Token, err error ) {
    lx.mustRune( '"' ) 
    buf := &bytes.Buffer{}
    loop := true
    for err == nil && loop {
        var r rune
        r, err = lx.readRune()
        if err == nil {
            if loop = r != '"'; loop { err = lx.appendStringRune( r, buf ) }
        }
    }
    if ! isLexErr( err ) {
        if loop { return nil, lx.prevError( "Unterminated string literal" ) }
        return StringToken( buf.String() ), err
    }
    return nil, err
}

type NumericToken struct {
    Int, Frac, Exp string
    ExpRune rune
}

func ( n *NumericToken ) IsInt() bool { return n.Frac == "" && n.Exp == "" }

func ( n *NumericToken ) String() string {
    buf := bytes.Buffer{}
    buf.WriteString( n.Int )
    if n.Frac != "" {
        buf.WriteRune( '.' )
        buf.WriteString( n.Frac )
    }
    if n.ExpRune > 0 {
        buf.WriteRune( n.ExpRune )
        buf.WriteString( n.Exp )
    }
    return buf.String()
}

func ( n *NumericToken ) Int64() ( int64, error ) {
    return strconv.ParseInt( n.String(), 10, 64 )
}

func ( n *NumericToken ) Uint64() ( uint64, error ) {
    return strconv.ParseUint( n.String(), 10, 64 )
}

func ( n *NumericToken ) Float64() ( float64, error ) {
    return strconv.ParseFloat( n.String(), 64 )
}

func isDigit( r rune ) bool { return r >= '0' && r <= '9' }
func isNumStart( r rune ) bool { return isDigit( r ) }

const (
    numErrDescInt = "integer part"
    numErrDescFrac = "fractional part"
    numErrDescExp = "exponent"

    errMsgEmptyExp = "Number has empty or invalid exponent"
    errMsgEmptyFrac = "Number has empty or invalid fractional part"
    errMsgNumberRune = "Unexpected char in %s: %q (%U)"
)

func ( lx *Lexer ) errNumberRune( r rune, numErrDesc string ) error {
    return lx.prevError( errMsgNumberRune, numErrDesc, string( r ), r )
}

func canEndDigStr( r rune ) bool {
    return isSpecialTokChar( r ) || isWhitespace( r ) || r == 'e' || r == 'E'
}

func ( lx *Lexer ) readDigitString( numErrDesc string ) ( string, error ) {
    buf := bytes.Buffer{}
    var err error
    for loop := true; err == nil && loop; {
        var r rune
        if r, err = lx.readRune(); err == nil {
            if loop = isDigit( r ); loop { 
                buf.WriteRune( r ) 
            } else { 
                if canEndDigStr( r ) {
                    lx.unreadRune() 
                } else { err = lx.errNumberRune( r, numErrDesc ) }
            }
        }
    }
    return buf.String(), err
}

func ( lx *Lexer ) readNumIntPart( n *NumericToken ) ( err error ) {
    n.Int, err = lx.readDigitString( numErrDescInt )
    return
}

func ( lx *Lexer ) readNumFracPart( n *NumericToken ) ( err error ) {
    var r rune
    r, err = lx.readRune()
    if isLexErr( err ) { return err }
    if err == io.EOF { return nil }
    if r != '.' {
        lx.unreadRune()
        return 
    }
    n.Frac, err = lx.readDigitString( numErrDescFrac )
    if n.Frac == "" && ( ! isLexErr( err ) ) {
        return lx.prevError( errMsgEmptyFrac )
    }
    return
}

// We special case some easily detectable errors here, such as 1.2.3, 1f, 1T,
// etc, since these could never be part of a valid higher-level expression and
// are in all likelihood the result of an improperly formed numeric literal.
// Other characters are assumed to properly begin a new valid token. That is,
// "1+" is not valid as a number token, but is a valid beginning of the token
// sequence "1", "+", "2", and we don't attempt to distinguish that here.
func ( lx *Lexer ) hasNumExp() ( rune, error ) {
    r, err := lx.readRune()
    if isLexErr( err ) { return 0, err }
    if err == io.EOF { return 0, nil }
    if r == 'e' || r == 'E' { return r, nil }
    if isIdentLower( r ) || isDeclaredTypeNameStart( r ) || r == '.' {
        tmpl := "Expected exponent start or num end, found: %q (%U)"
        return 0, lx.prevError( tmpl, string( r ), r )
    }
    lx.unreadRune()
    return 0, nil
}

func ( lx *Lexer ) readNumExpPart( n *NumericToken ) ( err error ) {
    if n.ExpRune, err = lx.hasNumExp(); err != nil || n.ExpRune == 0 { return }
    var r rune
    if r, err = lx.readRune(); isLexErr( err ) { return }
    if err == io.EOF { return lx.prevError( errMsgEmptyExp ) }
    buf := bytes.Buffer{}
    emptyLen := 0
    if r == '-' || r == '+' {
        if r == '-' {
            buf.WriteRune( r )
            emptyLen++
        }
    } else { lx.unreadRune() }
    var digStr string
    if digStr, err = lx.readDigitString( numErrDescExp ); isLexErr( err ) { 
        return 
    }
    buf.WriteString( digStr )
    if buf.Len() == emptyLen { return lx.parseError( errMsgEmptyExp ) }
    n.Exp = buf.String()
    return
}

func ( lx *Lexer ) readNumber() ( num *NumericToken, err error ) {
    num = new( NumericToken )
    if err = lx.readNumIntPart( num ); err != nil { return }
    if err = lx.readNumFracPart( num ); err != nil { return }
    if err = lx.readNumExpPart( num ); err != nil { return }
    return
}

// Return a stack result:
//
//      - if synthLoc is set and elt triggers a synth end then leave stack as it
//      and return a synth end, clearing synthLoc
//
//      - else, pop elt and clear synthLoc (if not already clear) and return elt
//
// In either case we also save an appropriate unread result
//
func ( lx *Lexer ) readStack() ( tok Token, lc *Location, err error ) {
    elt := lx.peek()
    if elt.eol && lx.synthLoc != nil {
        tok, lc, err = SpecialTokenSynthEnd, lx.synthLoc, nil
        lx.synthLoc = nil
        lx.unread = &unreadElt{ stackElt{ tok, lc, false }, lx.synthLoc }
    } else {
        lx.pop() // remove elt
        lx.unread = &unreadElt{ elt, lx.synthLoc }
        tok, lc, err = elt.tok, elt.lc, nil
        if lx.stackEmpty() { lx.updateSynthLoc( tok ) }
    }
    return
}

func ( lx *Lexer ) passWsAndComments() ( sawEol bool, err error ) {
    for {
        var r rune
        if r, err = lx.peekRune(); err != nil { return }
        switch {
        case isWhitespace( r ): 
            var ws WhitespaceToken
            if ws, err = lx.readWhitespace(); err == nil { 
                sawEol = sawEol || ws.hasNewline()
            }
        case r == '#': 
            _, err = lx.readComment()
            sawEol = true
        default: return
        }
    }
    panic( "unreachable" )
}

func ( lx *Lexer ) recordEof(
    tok Token, lc *Location, err error ) ( Token, *Location, error ) {
    lx.sawEof = true
    if tok == nil {
        if lx.synthLoc != nil {
            tok, lc, err, lx.synthLoc =
                SpecialTokenSynthEnd, lx.synthLoc, nil, nil
        }
    } else { err = nil }
    return tok, lc, err
}

// If this call returns and err == nil, eol may have been set to false, even if
// a it was set to true by passWsAndComments(), if there is no synthLoc set.
func ( lx *Lexer ) doStrip() ( eol bool, tok Token, lc *Location, err error ) {
    if eol, err = lx.passWsAndComments(); err != nil { 
        if err == io.EOF {
            _, lc, err = lx.recordEof( SpecialTokenSynthEnd, lc, err )
        }
        return
    }
    if eol && lx.synthLoc == nil { eol = false }
    lc = lx.makeLocation()
    return
}

type tokCall func() ( Token, error )

func ( lx *Lexer ) callTokenRead( 
    f tokCall, lc *Location ) ( Token, *Location, error ) {

    tok, err := f()
    if err == io.EOF { tok, lc, err = lx.recordEof( tok, lc, err ) }
    return tok, lc, err
}

func hasEol( tok Token ) bool {
    switch v := tok.( type ) {
    case WhitespaceToken: return v.hasNewline()
    case CommentToken: return true
    }
    return false
}

// sets synthLoc appropriately:
//
//  - if tok is non-EOL whitespace then leave synthLoc untouched
//
//  - otherwise set synthLoc to current location if tok could imply EOL, clear
//  synthLoc if not
//
func ( lx *Lexer ) updateSynthLoc( tok Token ) {
    setLoc := false
    switch v := tok.( type ) {
    case *mg.Identifier, DeclaredTypeName, StringToken, *NumericToken: 
        setLoc = true
    case WhitespaceToken: if v.hasNewline() { lx.synthLoc = nil }
    case SpecialToken:
        switch v {
        case SpecialTokenCloseBrace, 
             SpecialTokenCloseBracket,
             SpecialTokenCloseParen:
            setLoc = true
        default: lx.synthLoc = nil
        }
    case Keyword:
        if v == KeywordTrue || v == KeywordFalse || v == KeywordReturn {
            setLoc = true
        } else { lx.synthLoc = nil }
    default: lx.synthLoc = nil
    }
    if setLoc { lx.synthLoc = lx.makeLocation() }
}

// Read algorithm is:
//
//  - if there is a saved result return it via readStack()
//
//  - if we have seen EOF, return the EOF result
//
//  - if lx.strip is set, strip tokens and indicate whether an EOL is in play
//
//  - read the next non-whitespace token. if lx.strip is not set then also check
//  whether it indicates EOL (if it is set then the token just read will never
//  indicate EOL and the strip process will have detected it previously)
//
//  - if EOL is in play and synthLoc is not nil, push the current result onto
//  the stack, making the return val retVal a synth end, clearing synthLoc
//
//  - if EOL is not in play update synthLoc
//
//  - set the unread value according to retVal and EOL status and return 
//
func ( lx *Lexer ) implReadToken( 
    f tokCall ) ( tok Token, lc *Location, err error ) {

    defer func() { if err != nil && err != io.EOF { lx.unread = nil } }()
    if ! lx.stackEmpty() { return lx.readStack() }
    if lx.sawEof { return nil, nil, io.EOF }
    lc = lx.makeLocation()
    eol := false
    if lx.strip {
        eol, tok, lc, err = lx.doStrip()
        if ! ( tok == nil && err == nil ) { return }
    }
    if tok, lc, err = lx.callTokenRead( f, lc ); err != nil { return }
    if ! lx.strip { eol = hasEol( tok ) }
    if ! eol { lx.updateSynthLoc( tok ) }
    if eol && lx.synthLoc != nil {
        lx.push( stackElt{ tok, lc, eol } )
        tok, lc, lx.synthLoc = SpecialTokenSynthEnd, lx.synthLoc, nil
    }
    lx.unread = &unreadElt{ stackElt{ tok, lc, eol }, lx.synthLoc }
    return
}

func ( lx *Lexer ) ReadToken() ( Token, *Location, error ) {
    return lx.implReadToken( func() ( Token, error ) {
        r, err := lx.peekRune()
        if err != nil { return nil, err }
        switch {
        case isIdentLower( r ): return lx.readIdentifier()
        case isWhitespace( r ): return lx.readWhitespace()
        case isSpecialTokChar( r ): return lx.readSpecialTok()
        case isCommentStart( r ): return lx.readComment()
        case isDeclaredTypeNameStart( r ): return lx.readDeclaredTypeName()
        case isStringLiteralStart( r ): return lx.readStringLiteral()
        case isNumStart( r ): return lx.readNumber()
        }
        return nil, lx.parseError( "Unexpected char: %q (%U)", string( r ), r )
    })
}

func ( lx *Lexer ) ReadIdentifier() ( Token, *Location, error ) {
    return lx.implReadToken( func() ( Token, error ) {
        return lx.readIdentifier()
    })
}

func ( lx *Lexer ) ReadDeclaredTypeName() ( Token, *Location, error ) {
    return lx.implReadToken( func() ( Token, error ) {
        return lx.readDeclaredTypeName()
    })
}

func ( lx *Lexer ) ReadNumber() ( Token, *Location, error ) {
    return lx.implReadToken( func() ( Token, error ) {
        return lx.readNumber()
    })
}

type Options struct {
    SourceName string
    Reader io.Reader
    IsExternal bool
    Strip bool
}

func New( opts *Options ) *Lexer {
    return &Lexer{ 
        reader: bufio.NewReader( opts.Reader ), 
        line: 1,
        col: 1,
        newlineUnreadCol: -1,
        isExternal: opts.IsExternal,
        SourceName: opts.SourceName,
        strip: opts.Strip,
    }
}
