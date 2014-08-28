package com.bitgirder.mingle;

import static com.bitgirder.mingle.MingleLexer.SpecialLiteral;

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

    private final static SpecialLiteral[] NS_SPEC_LITS =
        new SpecialLiteral[] {
            SpecialLiteral.COLON,
            SpecialLiteral.ASPERAND
        };

    private final MingleLexer lx;

    // peekPos and curPos are stored as 1-indexed, unlike the 0-indexed pos of
    // MingleLexer from which they are obtained
    private int peekPos;
    private Object peekTok;
    private int curPos;

    private MingleParser( MingleLexer lx ) { this.lx = lx; }

    private int lxPos() { return ( (int) lx.position() ) + 1; }

    private int nextPos() { return peekTok == null ? lxPos() : peekPos; }

    private
    Object
    peekToken()
        throws MingleSyntaxException,
               IOException
    {
        if ( peekTok == null )
        {
            peekPos = lxPos();
            peekTok = lx.nextToken();
        }

        return peekTok;
    }

    private
    void
    checkUnexpectedEnd( String msg )
        throws MingleSyntaxException,
               IOException
    {
        if ( peekTok == null ) lx.checkUnexpectedEnd( msg );
    }

    private
    Object
    clearPeek()
    {
        state.isFalse( peekTok == null, "clearPeek() called without peek tok" );

        Object res = peekTok;

        peekPos = -1;
        peekTok = null;

        return res;
    }

    private
    Object
    nextToken()
        throws MingleSyntaxException,
               IOException
    {
        if ( peekTok == null )
        {
            curPos = lxPos();
            return lx.nextToken();
        }
        else
        {
            curPos = peekPos;
            return clearPeek();
        }
    }

    private
    MingleSyntaxException
    fail( int col,
          String msg )
    {
        return new MingleSyntaxException( msg, col );
    }

    private
    MingleSyntaxException
    failf( int col,
           String msg,
           Object... args )
    {
        return fail( col, String.format( msg, args ) );
    }

    private
    String
    errStringFor( Object tok )
    {
        if ( tok == null ) return "END";
        else if ( tok instanceof MingleString ) return "STRING";
        else if ( tok instanceof MingleLexer.Number ) return "NUMBER";
        else if ( tok instanceof MingleIdentifier ) return "IDENTIFIER";
        else if ( tok instanceof SpecialLiteral ) {
            return ( (SpecialLiteral) tok ).inspect();
        }
        
        throw state.createFailf( "Unexpected token: %s", tok );
    }

    private
    MingleSyntaxException
    failUnexpectedToken( int pos,
                         Object tok,
                         String expctMsg )
    {
        String msg = String.format(
            "Expected %s but found: %s", expctMsg, errStringFor( tok ) );
        
        return fail( pos, msg );
    }

    private
    MingleSyntaxException
    failUnexpectedToken( Object tok,
                         String expctMsg )  
    {
        return failUnexpectedToken( curPos, tok, expctMsg );
    }

    private
    String
    expectStringFor( SpecialLiteral[] specs )
    {
        String[] strs = new String[ specs.length ];

        for ( int i = 0, e = strs.length; i < e; ++i )
        {
            strs[ i ] = specs[ i ].inspect();
        }

        return Strings.join( " or ", (Object[]) strs ).toString();
    } 

    private
    SpecialLiteral
    pollSpecial( SpecialLiteral... specs )
        throws MingleSyntaxException,
               IOException
    {
        Object tok = peekToken();

        if ( tok instanceof SpecialLiteral )
        {
            SpecialLiteral act = (SpecialLiteral) tok;

            for ( SpecialLiteral spec : specs ) {
                if ( spec == act ) {
                    nextToken();
                    return act;
                }
            }
        }

        return null;
    }

    private
    SpecialLiteral
    expectSpecial( SpecialLiteral... specs )
        throws MingleSyntaxException,
               IOException
    {
        SpecialLiteral res = pollSpecial( specs );
        
        if ( res == null )
        {
            Object failTok = peekToken();
            int failPos = peekPos;
            String expctStr = expectStringFor( specs );
            
            throw failUnexpectedToken( failPos, failTok, expctStr );
        }
 
        return res;
    }

    private
    void
    checkNoTrailing()
        throws MingleSyntaxException,
               IOException
    {
        if ( peekTok == null ) lx.checkNoTrailing();
        else throw failUnexpectedToken( peekPos, peekTok, "END" );
    }

    private
    MingleIdentifier[]
    toArray( List< MingleIdentifier > l )
    {
        return l.toArray( new MingleIdentifier[ l.size() ] );
    }

    private
    < V >
    V
    peekTyped( Class< V > cls,
               String expctMsg )
        throws MingleSyntaxException
    {
        if ( peekTok == null ) return null;
        
        if ( cls.isInstance( peekTok ) ) return cls.cast( clearPeek() );

        throw failUnexpectedToken( peekPos, peekTok, expctMsg );
    }

    private
    MingleIdentifier
    expectIdentifier()
        throws MingleSyntaxException,
               IOException
    {
        MingleIdentifier res = 
            peekTyped( MingleIdentifier.class, "identifier" );

        if ( res == null ) res = lx.parseIdentifier( null );

        return res;
    }

    private
    DeclaredTypeName
    expectDeclaredTypeName()
        throws MingleSyntaxException,
               IOException
    {
        DeclaredTypeName res =
            peekTyped( DeclaredTypeName.class, "declared type name" );

        if ( res == null ) res = lx.parseDeclaredTypeName();

        return res;
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
            if ( expectSpecial( NS_SPEC_LITS ) == SpecialLiteral.COLON )
            {
                parts.add( expectIdentifier() );
            }
            else ver = expectIdentifier();
        }

        return new MingleNamespace( toArray( parts ), ver );
    }

    private
    QualifiedTypeName
    expectQname()
        throws MingleSyntaxException,
               IOException
    {
        MingleNamespace ns = expectNamespace();
        expectSpecial( SpecialLiteral.FORWARD_SLASH );
        DeclaredTypeName nm = expectDeclaredTypeName();

        return new QualifiedTypeName( ns, nm );
    }

    private
    QualifiedTypeName
    expectTypeName( MingleNameResolver r )
        throws MingleSyntaxException,
               IOException
    {
        Object tok = peekToken();

        if ( tok instanceof MingleIdentifier ) return expectQname();
        else if ( tok instanceof DeclaredTypeName )
        {
            int errPos = nextPos();
            DeclaredTypeName nm = expectDeclaredTypeName();
            
            QualifiedTypeName qn = r.resolve( nm );
            if ( qn != null ) return qn;

            throw new MingleSyntaxException( 
                "cannot resolve as a standard type: " + nm, errPos );
        }

        String expctMsg = "identifier or declared type name";
        throw failUnexpectedToken( peekPos, tok, expctMsg );
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
        DECLARED_TYPE_NAME( "declared type name" ),
        NAMESPACE( "namespace" ),
        QUALIFIED_TYPE_NAME( "qualified type name" );

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
            case DECLARED_TYPE_NAME: return p.expectDeclaredTypeName();
            case NAMESPACE: return p.expectNamespace();
            case QUALIFIED_TYPE_NAME: return p.expectQname();
            default: throw state.createFail( "Unhandled parse type:", typ );
        }
    }

    private
    static
    RuntimeException
    failIOException( IOException ioe )
    {
        return new RuntimeException( 
                "Got IOException from string source", ioe ); 
    }

    private
    static
    RuntimeException
    failCreate( MingleSyntaxException mse,
                String errNm )
    {
        return new RuntimeException( "Couldn't parse " + errNm, mse );
    }

    private
    static
    < V >
    V
    checkNoTrailing( V obj,
                     MingleParser p )
        throws MingleSyntaxException
    {
        try { p.checkNoTrailing(); }
        catch ( IOException ioe ) { throw failIOException( ioe ); }
        
        return obj;
    }

    private
    static
    Object
    doParse( CharSequence s,
             ParseType typ )
        throws MingleSyntaxException
    {
        MingleParser p = forString( s );

        try { return checkNoTrailing( callParse( p, typ ), p ); }
        catch ( IOException ioe ) { throw failIOException( ioe ); }
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
            throw failCreate( ex, typ.errNm ); 
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
    DeclaredTypeName
    parseDeclaredTypeName( CharSequence s )
        throws MingleSyntaxException
    {
        return (DeclaredTypeName) doParse( s, ParseType.DECLARED_TYPE_NAME );
    }

    static
    DeclaredTypeName
    createDeclaredTypeName( CharSequence s )
    {
        return (DeclaredTypeName) doCreate( s, ParseType.DECLARED_TYPE_NAME );
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

    static
    QualifiedTypeName
    parseQualifiedTypeName( CharSequence s )
        throws MingleSyntaxException
    {
        return (QualifiedTypeName) doParse( s, ParseType.QUALIFIED_TYPE_NAME );
    }

    static
    QualifiedTypeName
    createQualifiedTypeName( CharSequence s )
    {
        return (QualifiedTypeName) doCreate( s, ParseType.QUALIFIED_TYPE_NAME );
    }
}
