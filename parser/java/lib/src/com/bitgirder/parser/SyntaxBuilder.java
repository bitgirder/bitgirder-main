package com.bitgirder.parser;

public
interface SyntaxBuilder< N, T, S >
{
    public
    S
    buildSyntax( DerivationMatch< N, T > match );
}
