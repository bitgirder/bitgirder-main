package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.Set;
import java.util.Iterator;

public
final
class RecursiveDescentParserFactory< N, T >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final Grammar< N, T > grammar;

    private
    RecursiveDescentParserFactory( Grammar< N, T > grammar )
    {
        this.grammar = grammar;
    }

    public
    RecursiveDescentParser.Builder< N, T >
    createParserBuilder()
    {
        return 
            new RecursiveDescentParser.Builder< N, T >().
                setGrammar( grammar );
    }

    public
    RecursiveDescentParser< N, T >
    createParser( N goal )
    {
        inputs.notNull( goal, "goal" );

        return createParserBuilder().
               setGoal( goal ).
               build();
    }

    public
    static
    < N, T >
    RecursiveDescentParserFactory< N, T >
    forGrammar( Grammar< N, T > grammar )
        throws LeftRecursiveGrammarException
    {
        inputs.notNull( grammar, "grammar" );

        Grammars.assertNoLeftRecursion( grammar );
        return new RecursiveDescentParserFactory< N, T >( grammar );
    }
}
