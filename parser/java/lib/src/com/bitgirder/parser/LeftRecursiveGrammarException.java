package com.bitgirder.parser;

final
class LeftRecursiveGrammarException
extends RuntimeException
{
    LeftRecursiveGrammarException( Object head )
    {
        super( "Grammar contains left recursive production: " + head );
    }
}
