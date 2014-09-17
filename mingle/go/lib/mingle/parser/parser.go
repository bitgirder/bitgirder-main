package parser

import (
    mg "mingle"
    "time"
    "fmt"
    "bitgirder/objpath"
//    "log"
    "strconv"
)

type parseStringFunc func( b *Builder ) ( *TokenNode, error )

func ParseIdentifier( s string ) ( *mg.Identifier, error ) {
    sb := newSyntaxBuilderExt( s )
    tn, err := sb.ExpectIdentifier()
    if err == nil { err = sb.CheckTrailingToken() }
    if err == nil { return tn.Identifier(), nil }
    return nil, err
}

func idPathParseErrorUnexpectedToken( tn *TokenNode, b *Builder ) error {
    return b.ErrorTokenUnexpected( "identifier or list index", tn )
}

func idPathParseParseIndexNum( tn *TokenNode ) ( uint64, error ) {
    num := tn.Number()
    if ! num.IsInt() {
        return 0, NewParseErrorf( tn.Loc, "invalid decimal index: %s", num )
    }
    res, err := num.Uint64()
    if err == nil { return res, nil }
    if ne, ok := err.( *strconv.NumError ); ok { 
        if ne.Err == strconv.ErrRange {
            err = &ParseError{ Loc: tn.Loc, Message: "list index out of range" }
        }
    }
    return 0, err
}

// pre: will be positioned at opening '['
func idPathParseExpectIndex( b *Builder ) ( res uint64, err error ) {
    b.mustSpecial( SpecialTokenOpenBracket )
    if err = b.SkipWs(); err != nil { return }
    var tn *TokenNode
    if tn, err = b.PollSpecial( SpecialTokenMinus ); tn != nil {
        return 0, NewParseErrorf( tn.Loc, "negative list index" )
    }
    if tn, err = b.ExpectNumericToken(); err != nil { return }
    if res, err = idPathParseParseIndexNum( tn ); err != nil { return }
    if err = b.SkipWs(); err != nil { return }
    _, err = b.ExpectSpecial( SpecialTokenCloseBracket )
    return
}

func idPathParseBegin( b *Builder ) ( objpath.PathNode, error ) {
    if err := b.SkipWs(); err != nil { return nil, err }
    if err := b.CheckUnexpectedEnd(); err != nil { return nil, err }    
    tn, err := b.PeekToken()
    if err != nil { return nil, err }
    switch {
    case tn.IsSpecial( SpecialTokenOpenBracket ): 
        idx, err := idPathParseExpectIndex( b )
        if err == nil { return objpath.RootedAtList().SetIndex( idx ), nil }
        return nil, err
    case tn.IsIdentifier(): 
        b.MustNextToken()
        return objpath.RootedAt( tn.Identifier() ), nil
    }
    return nil, idPathParseErrorUnexpectedToken( tn, b )
}

// next token will be '.'
func idPathParseDescend( 
    p objpath.PathNode, b *Builder ) ( objpath.PathNode, error ) {

    b.mustSpecial( SpecialTokenPeriod )
    if err := b.SkipWs(); err != nil { return nil, err }
    tn, err := b.ExpectIdentifier()
    if err != nil { return nil, err }
    return p.Descend( tn.Identifier() ), nil
}

func idPathParseStartList(
    p objpath.PathNode, b *Builder ) ( objpath.PathNode, error ) {

    idx, err := idPathParseExpectIndex( b )
    if err != nil { return nil, err }
    return p.StartList().SetIndex( idx ), nil
}

// precondition: ws has been skipped and b.HasMoreTokens()
func idPathParseBuildNext( 
    p objpath.PathNode, b *Builder ) ( objpath.PathNode, error ) {

    tn, err := b.PeekToken()
    if err != nil { return nil, err }
    switch {
    case tn.IsSpecial( SpecialTokenPeriod ): return idPathParseDescend( p, b )
    case tn.IsSpecial( SpecialTokenOpenBracket ): 
        return idPathParseStartList( p, b )
    }
    return nil, idPathParseErrorUnexpectedToken( tn, b )
}

func ParseIdentifierPath( s string ) ( res objpath.PathNode, err error ) {
    b := newSyntaxBuilderExt( s )
    if res, err = idPathParseBegin( b ); err != nil { return }
    for {
        if err = b.SkipWs(); err != nil { return }
        if ! b.HasTokens() { break }
        if res, err = idPathParseBuildNext( res, b ); err != nil { return }
    }
    return
}

func ParseNamespace( s string ) ( *mg.Namespace, error ) {
    sb := newSyntaxBuilderExt( s )
    ns, _, err := sb.ExpectNamespace( nil )
    if err == nil { err = sb.CheckTrailingToken() }
    if err == nil { return ns, nil }
    return nil, err
}

func ParseDeclaredTypeName( s string ) ( *mg.DeclaredTypeName, error ) {
    sb := newSyntaxBuilderExt( s )
    tn, err := sb.ExpectDeclaredTypeName()
    if err == nil { err = sb.CheckTrailingToken() }
    if err == nil { return tn.DeclaredTypeName(), nil }
    return nil, err
}

func ParseQualifiedTypeName( s string ) ( *mg.QualifiedTypeName, error ) {
    sb := newSyntaxBuilderExt( s )
    qn, _, err := sb.ExpectQualifiedTypeName( nil )
    if err == nil { err = sb.CheckTrailingToken() }
    if err == nil { return qn, nil }
    return nil, err
}

func ParseTypeReference( s string ) ( *CompletableTypeReference, error ) {
    sb := newSyntaxBuilderExt( s )
    ctr, err := sb.ExpectTypeReference( nil )
    if err == nil { err = sb.CheckTrailingToken() }
    if err == nil { return ctr, nil }
    return nil, err
}

func ParseTimestamp( str string ) ( mg.Timestamp, error ) {
    t, err := time.Parse( time.RFC3339Nano, str )
    if err != nil {
        parseErr := &ParseError{
            Message: fmt.Sprintf( "Invalid RFC3339 time: %q", str ),
            Loc: &Location{ 1, 1, ParseSourceInput },
        }
        return mg.Timestamp( t ), parseErr
    }
    return mg.Timestamp( t ), nil
}
