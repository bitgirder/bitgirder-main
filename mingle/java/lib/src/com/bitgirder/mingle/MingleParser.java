package com.bitgirder.mingle;

import static com.bitgirder.mingle.MingleLexer.SpecialLiteral;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.lang.path.ObjectPath;

import java.io.IOException;

import java.util.List;
import java.util.Map;

import java.util.regex.Pattern;
import java.util.regex.PatternSyntaxException;

final
class MingleParser
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static SpecialLiteral[] NS_SPEC_LITS =
        new SpecialLiteral[] {
            SpecialLiteral.COLON,
            SpecialLiteral.ASPERAND
        };
    
    private final static List< SpecialLiteral > TYPE_QUANTS =
        Lang.asList(
            SpecialLiteral.QUESTION_MARK,
            SpecialLiteral.ASTERISK,
            SpecialLiteral.PLUS
        );

    private final static SpecialLiteral[] RANGE_MIN_LITS =
        new SpecialLiteral[] {
            SpecialLiteral.OPEN_PAREN, 
            SpecialLiteral.OPEN_BRACKET
        };

    private final static SpecialLiteral[] RANGE_MAX_LITS =
        new SpecialLiteral[] {
            SpecialLiteral.CLOSE_PAREN,
            SpecialLiteral.CLOSE_BRACKET
        };
    
    private final static Map< TypeName, AtomicTypeReference > RANGE_TYPES =
        Lang.newMap( TypeName.class, AtomicTypeReference.class,
            Mingle.QNAME_INT32, Mingle.TYPE_INT32,
            Mingle.QNAME_INT64, Mingle.TYPE_INT64,
            Mingle.QNAME_UINT32, Mingle.TYPE_UINT32,
            Mingle.QNAME_UINT64, Mingle.TYPE_UINT64,
            Mingle.QNAME_FLOAT32, Mingle.TYPE_FLOAT32,
            Mingle.QNAME_FLOAT64, Mingle.TYPE_FLOAT64,
            Mingle.QNAME_STRING, Mingle.TYPE_STRING,
            Mingle.QNAME_TIMESTAMP, Mingle.TYPE_TIMESTAMP
        );

    private final static MingleIdentifier BOUND_RANGE_MIN =
        new MingleIdentifier( new String[] { "min" } );

    private final static MingleIdentifier BOUND_RANGE_MAX =
        new MingleIdentifier( new String[] { "max" } );

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
        else if ( tok instanceof SpecialLiteral ) 
        {
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

            for ( SpecialLiteral spec : specs )
            {
                if ( spec == act ) 
                {
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
    MingleIdentifiedName
    expectIdentifiedName()
        throws MingleSyntaxException,
               IOException
    {
        MingleNamespace ns = expectNamespace();
        
        List< MingleIdentifier > names = Lang.newList();

        while ( pollSpecial( SpecialLiteral.FORWARD_SLASH ) != null )
        {
            names.add( expectIdentifier() );
        }

        if ( names.isEmpty() ) throw fail( lxPos(), "Missing name" );

        return new MingleIdentifiedName( ns, toArray( names ) );
    }

    private
    TypeName
    expectTypeName( MingleNameResolver r )
        throws MingleSyntaxException,
               IOException
    {
        Object tok = peekToken();

        if ( tok instanceof MingleIdentifier ) return expectQname();
        else if ( tok instanceof DeclaredTypeName )
        {
            DeclaredTypeName nm = expectDeclaredTypeName();
            QualifiedTypeName qn = r.resolve( nm );

            return qn == null ? nm : qn;
        }

        String expctMsg = "identifier or declared type name";
        throw failUnexpectedToken( peekPos, tok, expctMsg );
    }

    private
    MingleSyntaxException
    failRestrictionTarget( TypeName targ,
                           int nmPos,
                           String errTyp )
    {
        String msg = String.format(
            "Invalid target type for %s restriction: %s", 
            errTyp, targ.getExternalForm()
        );

        return new MingleSyntaxException( msg, nmPos );
    }

    private
    MingleSyntaxException
    asSyntaxException( PatternSyntaxException pse,
                       MingleString patStr )
    {
        String msg = String.format(
            "(near pattern string char %d) %s: %s", 
            pse.getIndex(), 
            pse.getDescription(), 
            patStr.getExternalForm()
        );

        return new MingleSyntaxException( msg, curPos );
    }

    private
    MingleRegexRestriction
    expectRegexRestriction( TypeName nm,
                            int nmPos )
        throws MingleSyntaxException,
               IOException
    {
        MingleString patStr = (MingleString) nextToken();

        if ( nm.equals( Mingle.QNAME_STRING ) )
        {
            try
            {
                Pattern pat = Pattern.compile( patStr.toString() );
                return MingleRegexRestriction.create( pat );
            }
            catch ( PatternSyntaxException pse )
            {
                throw asSyntaxException( pse, patStr );
            }
        }
 
        throw failRestrictionTarget( nm, nmPos, "regex" );
    }

    private
    boolean
    getRangeBound( SpecialLiteral[] toks )
        throws MingleSyntaxException,
               IOException
    {
        SpecialLiteral spec = expectSpecial( toks );

        return spec == SpecialLiteral.OPEN_BRACKET || 
               spec == SpecialLiteral.CLOSE_BRACKET;
    }

    private
    MingleSyntaxException
    rangeValTypeFail( String valTypDesc,
                      MingleIdentifier bound,
                      int errPos )
    {
        return failf( errPos, 
            "Got %s as %s value for range", valTypDesc, bound );
    }

    private
    MingleValue
    castRangeValue( MingleValue v,
                    MingleTypeReference t,
                    MingleIdentifier bound,
                    int errPos )
        throws MingleSyntaxException
    {
        ObjectPath< MingleIdentifier > p = ObjectPath.getRoot( bound );

        try { return Mingle.castValue( v, t, p ); }
        catch ( MingleValidationException mve )
        {
            throw failf( errPos, "Invalid %s value in range restriction: %s", 
                bound, mve.getError() );
        }
    }

    private
    MingleValue
    asRangeValue( AtomicTypeReference typ,
                  MingleString s,
                  MingleIdentifier bound,
                  int errPos )
        throws MingleSyntaxException
    {
        if ( typ.equals( Mingle.TYPE_STRING ) || 
             typ.equals( Mingle.TYPE_TIMESTAMP ) )
        {
            return castRangeValue( s, typ, bound, errPos );
        }

        throw rangeValTypeFail( "string", bound, errPos );
    }

    // Can make this faster later and skip the pass through MingleString; for
    // now just let Mingle.castValue() do the real work
    private
    MingleValue
    asRangeValue( AtomicTypeReference typ,
                  MingleLexer.Number n,
                  MingleIdentifier bound,
                  int errPos )
        throws MingleSyntaxException
    {
        if ( Mingle.isNumberType( typ ) )
        {
            if ( Mingle.isIntegralType( typ ) &&
                 ! ( n.f == null && n.e == null ) )
            {
                throw rangeValTypeFail( "decimal", bound, errPos );
            }

            MingleString numStr = new MingleString( n.toString() );
            return castRangeValue( numStr, typ, bound, errPos );
        }
        
        throw rangeValTypeFail( "number", bound, errPos );
    }

    private
    boolean
    impliesNullRangeValue( Object tok )
    {
        return tok == SpecialLiteral.COMMA ||
               tok == SpecialLiteral.CLOSE_PAREN ||
               tok == SpecialLiteral.CLOSE_BRACKET;
    }

    private
    MingleValue
    getRangeValue( AtomicTypeReference rngTyp,
                   MingleIdentifier bound )
        throws MingleSyntaxException,
               IOException
    {
        Object tok = peekToken();

        if ( tok instanceof MingleString ) 
        {
            MingleString ms = (MingleString) nextToken();
            return asRangeValue( rngTyp, ms, bound, curPos );
        }
        else if ( tok instanceof MingleLexer.Number )
        {
            MingleLexer.Number n = (MingleLexer.Number) nextToken();
            return asRangeValue( rngTyp, n, bound, curPos );
        }
        else if ( impliesNullRangeValue( tok ) ) return null; // leave peekVal

        throw failUnexpectedToken( peekPos, tok, "range value" );
    }

    private
    final
    static
    class RangeBuilder
    {
        // minPos,maxPos are positions of min or max bound token

        private int minPos;
        private boolean minClosed;
        private MingleValue min;
        private MingleValue max;
        private boolean maxClosed;
        private int maxPos;

        private AtomicTypeReference rngTyp;

        private
        MingleRangeRestriction
        build()
        {
            Class< ? extends MingleValue > typeTok = 
                Mingle.expectValueClassFor( rngTyp );

            return MingleRangeRestriction.createChecked(
                minClosed, min, max, maxClosed, typeTok );
        }
    }

    // min,max have been cast successfully already, so the unchecked cast below
    // is okay
    private
    int
    getRangeCompSignum( RangeBuilder b )
    {
        Comparable< MingleValue > minCmp = Lang.castUnchecked( b.min );

        return minCmp.compareTo( b.max );
    }

    private
    boolean
    areAdjacentInts( RangeBuilder b )
    {
        if ( b.max instanceof MingleInt32 )
        {
            return ( (MingleInt32) b.max ).intValue() -
                   ( (MingleInt32) b.min ).intValue() == 1;
        }
        else if ( b.max instanceof MingleUint32 )
        {
            return ( (MingleUint32) b.max ).intValue() -
                   ( (MingleUint32) b.min ).intValue() == 1;
        }
        else if ( b.max instanceof MingleInt64 )
        {
            return ( (MingleInt64) b.max ).longValue() -
                   ( (MingleInt64) b.min ).longValue() == 1L;
        }
        else if ( b.max instanceof MingleUint64 )
        {
            return ( (MingleUint64) b.max ).longValue() -
                   ( (MingleUint64) b.min ).longValue() == 1L;
        }
        else return false;
    } 

    private
    void
    checkInifiniteRangeBounds( RangeBuilder b )
        throws MingleSyntaxException
    {
        boolean errLf = b.minClosed && b.min == null;
        boolean errRt = b.maxClosed && b.max == null;

        String msg = null;
        int errPos = 0;

        if ( errLf )
        {
            errPos = b.minPos;

            if ( errRt ) msg = "Infinite range must be open";
            else msg = "Infinite low range must be open";
        }
        else if ( errRt )
        {
            errPos = b.maxPos;
            msg = "Infinite high range must be open";
        }

        if ( msg != null ) throw fail( errPos, msg );
    }

    private
    void
    checkRangeSatisfiable( RangeBuilder b )
        throws MingleSyntaxException
    {
        checkInifiniteRangeBounds( b );

        if ( b.min == null || b.max == null ) return;
        int i = getRangeCompSignum( b );

        boolean sat = false;
        if ( i == 0 ) sat = b.minClosed && b.maxClosed;
        else if ( i > 0 ) sat = false;
        else
        {
            boolean open = ! ( b.minClosed || b.maxClosed );
            sat = ! ( open && areAdjacentInts( b ) );
        }

        if ( ! sat ) throw fail( b.minPos, "Unsatisfiable range" );
    }

    private
    MingleRangeRestriction
    expectRangeRestriction( TypeName nm,
                            int nmPos )
        throws MingleSyntaxException,
               IOException
    {
        RangeBuilder b = new RangeBuilder();

        b.rngTyp = RANGE_TYPES.get( nm );
        if ( b.rngTyp == null ) 
        {
            throw failRestrictionTarget( nm, nmPos, "range" );
        }

        b.minClosed = getRangeBound( RANGE_MIN_LITS );
        b.minPos = curPos;
        b.min = getRangeValue( b.rngTyp, BOUND_RANGE_MIN );
        expectSpecial( SpecialLiteral.COMMA );
        b.max = getRangeValue( b.rngTyp, BOUND_RANGE_MAX );
        b.maxClosed = getRangeBound( RANGE_MAX_LITS );
        b.maxPos = curPos;

        checkRangeSatisfiable( b );

        return b.build();
    }

    private
    MingleValueRestriction
    expectRestriction( TypeName nm,
                       int nmPos )
        throws MingleSyntaxException,
               IOException
    {
        Object tok = peekToken();

        if ( tok instanceof MingleString )
        {
            return expectRegexRestriction( nm, nmPos );
        }

        if ( tok == SpecialLiteral.OPEN_PAREN ||
             tok == SpecialLiteral.OPEN_BRACKET )
        {
            return expectRangeRestriction( nm, nmPos );
        }
 
        throw failUnexpectedToken( peekPos, peekTok, "restriction" );
    }

    private
    AtomicTypeReference
    expectAtomicTypeReference( MingleNameResolver r )
        throws MingleSyntaxException,
               IOException
    {
        checkUnexpectedEnd( "type reference" );

        int nmPos = nextPos();
        TypeName nm = expectTypeName( r );

        MingleValueRestriction vr = null;

        if ( pollSpecial( SpecialLiteral.TILDE ) != null )
        {
            checkUnexpectedEnd( "type restriction" );
            vr = expectRestriction( nm, nmPos );
        }

        return new AtomicTypeReference( nm, vr );
    }

    private
    MingleTypeReference
    quantifyType( AtomicTypeReference typ,
                  List< SpecialLiteral > quants )
    {
        MingleTypeReference res = typ;

        for ( SpecialLiteral quant : quants )
        {
            switch ( quant )
            {
                case QUESTION_MARK: 
                    res = new NullableTypeReference( res ); break;

                case PLUS: res = new ListTypeReference( res, false ); break;
                case ASTERISK: res = new ListTypeReference( res, true ); break;

                default: state.failf( "Unhandled quant: %s", quant );
            }
        }

        return res;
    }

    private
    SpecialLiteral
    readTypeQuant()
        throws MingleSyntaxException,
               IOException
    {
        if ( peekTok != null )
        {
            Object tok = nextToken();

            if ( tok instanceof SpecialLiteral )
            {
                if ( TYPE_QUANTS.contains( tok ) ) return (SpecialLiteral) tok;
            }

            throw failUnexpectedToken( curPos, tok, "type quantifier" );
        }

        return lx.readTypeQuant();
    }

    private
    List< SpecialLiteral >
    readTypeQuants()
        throws MingleSyntaxException,
               IOException
    {
        List< SpecialLiteral > res = Lang.newList();

        SpecialLiteral quant;
        while ( ( quant = readTypeQuant() ) != null ) res.add( quant );

        return res;
    }

    private
    MingleTypeReference
    expectTypeReference( MingleNameResolver r )
        throws MingleSyntaxException,
               IOException
    {
        AtomicTypeReference atr = expectAtomicTypeReference( r );

        List< SpecialLiteral > quants = readTypeQuants();
        return quantifyType( atr, quants );
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
        QUALIFIED_TYPE_NAME( "qualified type name" ),
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
            case DECLARED_TYPE_NAME: return p.expectDeclaredTypeName();
            case NAMESPACE: return p.expectNamespace();
            case QUALIFIED_TYPE_NAME: return p.expectQname();
            case IDENTIFIED_NAME: return p.expectIdentifiedName();
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
    MingleIdentifiedName
    parseIdentifiedName( CharSequence s )
        throws MingleSyntaxException
    {
        return (MingleIdentifiedName) doParse( s, ParseType.IDENTIFIED_NAME );
    }

    static
    MingleIdentifiedName
    createIdentifiedName( CharSequence s )
    {
        return (MingleIdentifiedName) doCreate( s, ParseType.IDENTIFIED_NAME );
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

    static
    MingleTypeReference
    parseTypeReference( CharSequence s,
                        MingleNameResolver r )
        throws MingleSyntaxException
    {
        inputs.notNull( r, "r" );

        MingleParser p = forString( s );

        try { return checkNoTrailing( p.expectTypeReference( r ), p ); }
        catch ( IOException ioe ) { throw failIOException( ioe ); }
    }

    static
    MingleTypeReference
    createTypeReference( CharSequence s,
                         MingleNameResolver r )
    {
        try { return parseTypeReference( s, r ); }
        catch ( MingleSyntaxException mse ) 
        { 
            throw failCreate( mse, "type reference" ); 
        }
    }
}
