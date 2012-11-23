package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import java.io.IOException;

import java.util.List;

final
class MingleParser
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static MingleLexer.SpecialLiteral[] NS_SPEC_LITS =
        new MingleLexer.SpecialLiteral[] {
            MingleLexer.SpecialLiteral.COLON,
            MingleLexer.SpecialLiteral.ASPERAND
        };

    private final MingleLexer lx;

    private int peekPos;
    private Object peekTok;
    private int curPos;

    private MingleParser( MingleLexer lx ) { this.lx = lx; }

    private
    Object
    peekToken()
        throws MingleSyntaxException,
               IOException
    {
        if ( peekTok == null )
        {
            peekPos = (int) lx.position();
            peekTok = lx.nextToken();
        }

        return peekTok;
    }

    private
    Object
    nextToken()
        throws MingleSyntaxException,
               IOException
    {
        Object res;

        if ( peekTok == null )
        {
            curPos = (int) lx.position();
            res = lx.nextToken();
        }
        else
        {
            curPos = peekPos;
            res = peekTok;
            peekPos = -1;
            peekTok = null;
        }

        return res;
    }

    private
    MingleSyntaxException
    fail( int col,
          String msg )
    {
        return new MingleSyntaxException( msg, col );
    }

    private
    String
    errStringFor( Object tok )
    {
        if ( tok == null ) return "END";
        else if ( tok instanceof MingleString ) return "STRING";
        else if ( tok instanceof MingleLexer.Number ) return "NUMBER";
        else if ( tok instanceof MingleIdentifier ) return "IDENTIFIER";
        else if ( tok instanceof MingleLexer.SpecialLiteral ) 
        {
            return ( (MingleLexer.SpecialLiteral) tok ).inspect();
        }
        
        throw state.createFailf( "Unexpected token: %s", tok );
    }

    private
    MingleSyntaxException
    failUnexpectedToken( Object tok,
                         String expctMsg )  
    {
        String msg = String.format(
            "Expected %s but found: %s", expctMsg, errStringFor( tok ) );
        
        return fail( curPos + 1, msg );
    }

    private
    String
    expectStringFor( MingleLexer.SpecialLiteral[] specs )
    {
        String[] strs = new String[ specs.length ];

        for ( int i = 0, e = strs.length; i < e; ++i )
        {
            strs[ i ] = specs[ i ].inspect();
        }

        return Strings.join( " or ", (Object[]) strs ).toString();
    } 

    private
    MingleLexer.SpecialLiteral
    expectSpecial( MingleLexer.SpecialLiteral... specs )
        throws MingleSyntaxException,
               IOException
    {
        Object tok = nextToken();

        if ( tok instanceof MingleLexer.SpecialLiteral )
        {
            MingleLexer.SpecialLiteral act = (MingleLexer.SpecialLiteral) tok;

            for ( MingleLexer.SpecialLiteral spec : specs )
            {
                if ( spec == act ) return act;
            }
        }

        throw failUnexpectedToken( tok, expectStringFor( specs ) );
    }

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

    private
    MingleNamespace
    expectNamespace()
        throws MingleSyntaxException,
               IOException
    {
        List< MingleIdentifier > parts = Lang.newList();
        MingleIdentifier ver = null;

        parts.add( expectIdentifier() );

        while ( ver == null )
        {
            if ( expectSpecial( NS_SPEC_LITS ) ==
                 MingleLexer.SpecialLiteral.COLON )
            {
                parts.add( expectIdentifier() );
            }
            else ver = expectIdentifier();
        }

        return new MingleNamespace(
            parts.toArray( new MingleIdentifier[ parts.size() ] ), ver );
    }

    static
    MingleParser
    forString( CharSequence s )
    {
        inputs.notNull( s, "s" );
        return new MingleParser( MingleLexer.forString( s ) );
    }

    private
    static
    enum ParseType
    {
        IDENTIFIER( "identifier" ),
        NAMESPACE( "namespace" ),
        QUALIFIED_TYPE_NAME( "qualified type name" ),
        DECLARED_TYPE_NAME( "declared type name" ),
        TYPE_REFERENCE( "type reference" ),
        IDENTIFIED_NAME( "identified name" );

        private final String errNm;

        private ParseType( String errNm ) { this.errNm = errNm; }
    }

    private
    static
    Object
    callParse( MingleParser p,
               ParseType typ )
        throws MingleSyntaxException,
               IOException
    {
        switch ( typ )
        {
            case IDENTIFIER: return p.expectIdentifier();
            case NAMESPACE: return p.expectNamespace();
            default: throw state.createFail( "Unhandled parse type:", typ );
        }
    }

    private
    static
    Object
    doParse( CharSequence s,
             ParseType typ )
        throws MingleSyntaxException
    {
        MingleParser p = forString( s );

        try
        {
            Object res = callParse( p, typ );
            p.checkNoTrailing(); 

            return res;
        }
        catch ( IOException ioe ) 
        { 
            throw new RuntimeException( 
                "Got IOException from string source", ioe ); 
        }
    }

    private
    static
    Object
    doCreate( CharSequence s,
              ParseType typ )
    {
        try { return doParse( s, typ ); }
        catch ( MingleSyntaxException ex )
        {
            throw new RuntimeException( "Couldn't parse " + typ.errNm, ex );
        }
    }

    static
    MingleIdentifier
    parseIdentifier( CharSequence s )
        throws MingleSyntaxException
    {
        return (MingleIdentifier) doParse( s, ParseType.IDENTIFIER );
    }

    static
    MingleIdentifier
    createIdentifier( CharSequence s )
    {
        return (MingleIdentifier) doCreate( s, ParseType.IDENTIFIER );
    }

    static
    MingleNamespace
    parseNamespace( CharSequence s )
        throws MingleSyntaxException
    {
        return (MingleNamespace) doParse( s, ParseType.NAMESPACE );
    }

    static
    MingleNamespace
    createNamespace( CharSequence s )
    {
        return (MingleNamespace) doCreate( s, ParseType.NAMESPACE );
    }
}
