package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

// N: head of derivations in syntactic grammar
// L: head of derivations in lexical grammar
// T: terminals in syntactic grammar, aka tokens built from L-goal derivations
//
public
final
class DocumentParserFactory< N, L, T >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final RecursiveDescentParserFactory< L, Character > lexicalPf;
    private final L lexicalGoal;
    private final Lexer.TokenBuilder< L, T > tokenBuilder;
    private final RecursiveDescentParserFactory< N, T > syntacticPf;
    private final N syntacticGoal;

    private
    DocumentParserFactory( Builder< N, L, T > b )
    {
        this.lexicalPf = inputs.notNull( b.lexicalPf, "lexicalPf" );
        this.lexicalGoal = inputs.notNull( b.lexicalGoal, "lexicalGoal" );
        this.tokenBuilder = inputs.notNull( b.tokenBuilder, "tokenBuilder" );
        this.syntacticPf = inputs.notNull( b.syntacticPf, "syntacticPf" );
        this.syntacticGoal = inputs.notNull( b.syntacticGoal, "syntacticGoal" );
    }

    RecursiveDescentParserFactory< L, Character >
    getLexicalParserFactory()
    {
        return lexicalPf;
    }

    L getLexicalGoal() { return lexicalGoal; }

    Lexer.TokenBuilder< L, T > getTokenBuilder() { return tokenBuilder; }

    RecursiveDescentParserFactory< N, T >
    getSyntacticParserFactory()
    {
        return syntacticPf;
    }

    N getSyntacticGoal() { return syntacticGoal; }

    public
    < D >
    DocumentParser.Builder< N, L, T, D >
    createParserBuilder()
    {
        return new DocumentParser.Builder< N, L, T, D >( this );
    }

    public
    static
    final
    class Builder< N, L, T >
    {
        private RecursiveDescentParserFactory< L, Character > lexicalPf;
        private L lexicalGoal;
        private Lexer.TokenBuilder< L, T > tokenBuilder;
        private RecursiveDescentParserFactory< N, T > syntacticPf;
        private N syntacticGoal;

        public
        Builder< N, L, T >
        setLexicalParserFactory( 
            RecursiveDescentParserFactory< L, Character > lexicalPf )
        {
            this.lexicalPf = inputs.notNull( lexicalPf, "lexicalPf" );
            return this;
        }

        public
        Builder< N, L, T >
        setLexicalGrammar( Grammar< L, Character > g )
        {
            inputs.notNull( g, "g" );

            return setLexicalParserFactory(
                RecursiveDescentParserFactory.< L, Character >forGrammar( g ) );
        }

        public
        Builder< N, L, T >
        setLexicalGoal( L lexicalGoal )
        {
            this.lexicalGoal = inputs.notNull( lexicalGoal, "lexicalGoal" );
            return this;
        }

        public
        Builder< N, L, T >
        setTokenBuilder( Lexer.TokenBuilder< L, T > tokenBuilder )
        {
            this.tokenBuilder = inputs.notNull( tokenBuilder, "tokenBuilder" );
            return this;
        }

        public
        Builder< N, L, T >
        setSyntacticParserFactory( 
            RecursiveDescentParserFactory< N, T > syntacticPf )
        {
            this.syntacticPf = inputs.notNull( syntacticPf, "syntacticPf" );
            return this;
        }

        public
        Builder< N, L, T >
        setSyntacticGrammar( Grammar< N, T > g )
        {
            inputs.notNull( g, "g" );

            return setSyntacticParserFactory(
                RecursiveDescentParserFactory.< N, T >forGrammar( g ) );
        }

        public
        Builder< N, L, T >
        setSyntacticGoal( N syntacticGoal )
        {
            this.syntacticGoal = 
                inputs.notNull( syntacticGoal, "syntacticGoal" );

            return this;
        }

        public
        DocumentParserFactory< N, L, T >
        build()
        {
            return new DocumentParserFactory< N, L, T >( this );
        }
    }
}
