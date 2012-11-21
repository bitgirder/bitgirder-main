package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.io.IOException;

final
class MingleParser
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleLexer lx;

    private MingleParser( MingleLexer lx ) { this.lx = lx; }

    private
    void
    checkNoTrailing()
        throws MingleSyntaxException,
               IOException
    {
        lx.checkNoTrailing();
    }

    private
    MingleIdentifier
    expectIdentifier()
        throws MingleSyntaxException,
               IOException
    {
        return lx.parseIdentifier( null );
    }

    static
    MingleParser
    forString( CharSequence s )
    {
        inputs.notNull( s, "s" );
        return new MingleParser( MingleLexer.forString( s ) );
    }

    static
    MingleIdentifier
    parseIdentifier( CharSequence s )
        throws MingleSyntaxException
    {
        MingleParser p = forString( s );

        try
        {
            MingleIdentifier res = p.expectIdentifier();
            p.checkNoTrailing(); 

            return res;
        }
        catch ( IOException ioe ) { throw new RuntimeException( ioe ); }
    }
}
