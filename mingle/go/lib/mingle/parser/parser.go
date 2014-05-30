package parser

import (
    mg "mingle"
//    "log"
)

type parseStringFunc func( b *Builder ) ( *TokenNode, error )

func ParseIdentifier( s string ) ( *mg.Identifier, error ) {
    sb := newSyntaxBuilderExt( s )
    tn, err := sb.ExpectIdentifier()
    if err == nil { err = sb.CheckTrailingToken() }
    if err == nil { return tn.Identifier(), nil }
    return nil, err
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

