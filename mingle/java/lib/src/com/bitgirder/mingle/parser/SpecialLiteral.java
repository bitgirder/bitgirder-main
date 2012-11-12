package com.bitgirder.mingle.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
enum SpecialLiteral
{
    COLON( ":" ),
    OPEN_BRACE( "{" ),
    CLOSE_BRACE( "}" ),
    SEMICOLON( ";" ),
    TILDE( "~" ),
    OPEN_PAREN( "(" ),
    CLOSE_PAREN( ")" ),
    OPEN_BRACKET( "[" ),
    CLOSE_BRACKET( "]" ),
    COMMA( "," ),
    QUESTION_MARK( "?" ),
    RETURNS( "->" ),
    MINUS( "-" ),
    FORWARD_SLASH( "/" ),
    PERIOD( "." ),
    ASTERISK( "*" ),
    PLUS( "+" ),
    LESS_THAN( "<" ),
    GREATER_THAN( ">" ),
    ASPERAND( "@" );

    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final static String ALPHABET = ":;{}~()[],?<->/.*+@";

    private final String lit;

    private SpecialLiteral( String lit ) { this.lit = lit; }

    public int length() { return lit.length(); }
    public char charAt( int indx ) { return lit.charAt( indx ); }
    public String getLiteral() { return lit; }

    public
    CharSequence
    getQuoted()
    {
        return 
            new StringBuilder( length() + 2 ).
                append( '\'' ).
                append( getLiteral() ).
                append( '\'' );
    }
}
