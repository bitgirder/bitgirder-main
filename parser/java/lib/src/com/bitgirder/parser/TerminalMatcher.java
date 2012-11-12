package com.bitgirder.parser;

interface TerminalMatcher< T >
extends ProductionMatcher
{
    boolean
    isMatch( T terminal );
}
