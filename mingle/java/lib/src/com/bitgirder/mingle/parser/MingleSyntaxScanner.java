package com.bitgirder.mingle.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.parser.SyntaxException;
import com.bitgirder.parser.SourceTextLocation;

import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleTypeName;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.MingleIdentifierFormat;
import com.bitgirder.mingle.model.MingleNamespace;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleString;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleInt64;
import com.bitgirder.mingle.model.MingleDouble;
import com.bitgirder.mingle.model.QualifiedTypeName;
import com.bitgirder.mingle.model.RelativeTypeName;
import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.NullableTypeReference;
import com.bitgirder.mingle.model.ListTypeReference;

import java.util.Queue;
import java.util.List;

// Meant to be subclassed by classes adding further scanner methods or used as
// an adjunct to parsing directly
public
class MingleSyntaxScanner
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static MingleIdentifier VER_MIGRATE =
        MingleIdentifier.create( "ver-migrate" );

    private final Queue< ? extends MingleToken > toks;

    public
    MingleSyntaxScanner( Queue< ? extends MingleToken > toks )
    {
        this.toks = inputs.notNull( toks, "toks" );
    }

    // Returns the live queue
    public final Queue< ? extends MingleToken > getTokens() { return toks; }

    public final MingleToken peek() { return toks.peek(); }
    public final MingleToken remove() { return toks.remove(); }
    public final boolean isEmpty() { return toks.isEmpty(); }

    public
    final
    SyntaxException
    syntaxException( Object... msg )
    {
        return MingleParsers.syntaxException( msg );
    }

    public
    final
    SyntaxException
    syntaxException( SourceTextLocation loc,
                     Object... msg )
    {
        inputs.notNull( loc, "loc" );
        return MingleParsers.syntaxException( loc, msg );
    }

    public
    final
    SyntaxException
    syntaxException( MingleToken tok,
                     Object... msg )
    {
        inputs.notNull( tok, "tok" );

        return syntaxException( tok.getLocation(), msg );
    }

    private
    CharSequence 
    quoteIdentifier( MingleIdentifier id )
    {
        MingleIdentifierFormat fmt = MingleIdentifierFormat.LC_CAMEL_CAPPED;

        return 
            new StringBuilder().
                append( '\'' ).
                append( MingleModels.format( id, fmt ) ).
                append( '\'' );
    }

    private
    CharSequence
    dispNameForInstance( Object obj )
    {
        if ( obj instanceof MingleToken )
        {
            return dispNameForInstance( ( (MingleToken) obj ).getObject() );
        }
        else if ( obj instanceof WhitespaceText ) return "whitespace";
        else if ( obj instanceof SpecialLiteral )
        {
            return ( (SpecialLiteral) obj ).getQuoted();
        }
        else if ( obj instanceof MingleIdentifier ) 
        {
            return quoteIdentifier( (MingleIdentifier) obj );
        }
        else return String.valueOf( obj );
    }

    private
    CharSequence
    dispNameForType( Class< ? > cls,
                     CharSequence errName )
    {
        return errName == null ? cls.getSimpleName() : errName;
    }

    public
    boolean
    isLcAlpha( char ch ) 
    { 
        return MingleParsers.isLcAlpha( ch ); 
    }

    public
    boolean
    isUcAlpha( char ch ) 
    { 
        return MingleParsers.isUcAlpha( ch ); 
    }

    public
    boolean
    isDigit( char ch ) 
    { 
        return MingleParsers.isDigit( ch ); 
    }

    private
    SyntaxException
    createUnexpectedEndException()
    {
        return new SyntaxException( MingleParsers.UNEXPECTED_END_MSG );
    }

    public
    final
    MingleToken
    expectToken()
        throws SyntaxException
    {
        MingleToken tok = toks.poll();

        if ( tok == null ) throw createUnexpectedEndException();
        else return tok;
    }

    public
    final
    MingleToken
    expectToken( MingleToken tok,
                 Class< ? > cls,
                 CharSequence errName )
        throws SyntaxException
    {
        inputs.notNull( tok, "tok" );
        inputs.notNull( cls, "cls" );

        if ( cls.isInstance( tok.getObject() ) ) return tok;
        else 
        {
            throw 
                syntaxException( 
                    tok,
                    "Expected", dispNameForType( cls, errName ), 
                    "but got:", dispNameForInstance( tok )
                );
        }
    }

    public
    final
    void
    checkUnexpectedEnd()
        throws SyntaxException
    {
        if ( toks.isEmpty() ) throw createUnexpectedEndException();
    }

    public
    final
    MingleToken
    expectToken( Class< ? > cls,
                 CharSequence errName )
        throws SyntaxException
    {
        inputs.notNull( cls, "cls" );
        return expectToken( expectToken(), cls, errName );
    }
    
    public
    final
    MingleToken
    expectToken( Class< ? > cls )
        throws SyntaxException
    {
        return expectToken( cls, null );
    }

    public
    final
    MingleToken
    pollToken( Class< ? > cls )
    {
        MingleToken tok = toks.peek();

        if ( tok != null && cls.isInstance( tok.getObject() ) ) 
        {
            toks.remove();
            return tok;
        }
        else return null;
    }

    public
    final
    MingleToken
    pollTokenObject( Object obj )
    {
        MingleToken tok = toks.peek();

        if ( tok != null && tok.getObject().equals( obj ) )
        {
            toks.remove();
            return tok;
        }
        else return null;
    }

    public
    final
    < V >
    V
    pollInstance( Class< V > cls )
    {
        MingleToken tok = pollToken( cls );
        return tok == null ? null : cls.cast( tok.getObject() );
    }

    public
    final
    MingleToken
    peekToken( Class< ? > cls )
    {
        MingleToken tok = peek();

        if ( tok != null && cls.isInstance( tok.getObject() ) ) return tok;
        else return null;
    }

    public
    final
    < V >
    V
    peekInstance( Class< V > cls )
    {
        MingleToken tok = peekToken( cls );
        return tok == null ? null : cls.cast( tok.getObject() );
    }

    public
    final
    < V >
    V
    expectInstance( Class< V > cls,
                    String dispName )
        throws SyntaxException
    {
        inputs.notNull( cls, "cls" );

        return cls.cast( expectToken( cls, dispName ).getObject() );
    }

    public
    final
    < V >
    V
    expectInstance( Class< V > cls )
        throws SyntaxException
    {
        return expectInstance( cls, null );
    }

    public
    final
    < V >
    V
    expectInstance( MingleToken tok,
                    Class< V > cls,
                    CharSequence errName )
        throws SyntaxException
    {
        inputs.notNull( tok, "tok" );
        inputs.notNull( cls, "cls" );

        return cls.cast( expectToken( tok, cls, errName ).getObject() );
    }

    public
    final
    < V >
    V
    expectInstance( MingleToken tok,
                    Class< V > cls )
        throws SyntaxException
    {
        return expectInstance( tok, cls, null );
    }

    public
    final
    SpecialLiteral
    pollLiteral( SpecialLiteral expct )
    {
        inputs.notNull( expct, "expct" );

        MingleToken tok = toks.peek();

        if ( tok != null && expct.equals( tok.getObject() ) ) 
        {
            toks.remove();
            return expct;
        }
        else return null;
    }

    public
    final
    MingleToken
    expectLiteral( SpecialLiteral expct )
        throws SyntaxException
    {
        inputs.notNull( expct, "expct" );

        MingleToken tok = 
            expectToken( SpecialLiteral.class, expct.getQuoted() );

        if ( expct.equals( tok.getObject() ) ) return tok;
        else
        {
            throw
                syntaxException(
                    tok,
                    "Expected", dispNameForInstance( expct ), "but got", 
                    dispNameForInstance( tok )
                );
        }
    }

    public
    final
    MingleToken
    pollIdentifiableText( CharSequence expct )
    {
        MingleToken tok = peekToken( IdentifiableText.class );

        if ( tok == null ) return null;
        else
        {
            IdentifiableText txt = (IdentifiableText) tok.getObject();
    
            if ( txt != null && txt.equalsString( expct ) )
            {
                remove();
                return tok;
            }
            else return null;
        }
    } 

    public
    final
    void
    skipWhitespace()
    {
        while ( pollToken( WhitespaceText.class ) != null );
    }

    public
    final
    List< WhitespaceText >
    expectWhitespaceMulti()
        throws SyntaxException
    {
        checkUnexpectedEnd();

        List< WhitespaceText > res = Lang.newList();

        for ( WhitespaceText t = pollInstance( WhitespaceText.class );
              t != null;
              t = pollInstance( WhitespaceText.class ) )
        {
            res.add( t );
        }

        if ( res.isEmpty() )
        {
            throw syntaxException( peek(), "Expected whitespace" );
        }
        else return res;
    }

    private
    MingleToken
    doExpectText( CharSequence errNm )
        throws SyntaxException
    {
        return expectToken( IdentifiableText.class, errNm );
    }

    public
    final
    MingleToken
    expectText( CharSequence errNm )
        throws SyntaxException
    {
        return doExpectText( inputs.notNull( errNm, "errNm" ) );
    }

    public
    final
    MingleToken
    expectText()
        throws SyntaxException
    {
        return doExpectText( null );
    }

    public
    final
    MingleTypeName
    parseTypeName( MingleToken t )
        throws SyntaxException
    {
        inputs.notNull( t, "t" );

        return
            MingleParsers.parseTypeName( 
                (IdentifiableText) t.getObject(), 
                t.getLocation() 
            );
    } 

    public
    final
    MingleTypeName
    expectTypeName()
        throws SyntaxException
    {
        MingleToken t = expectText( "type name" );

        return parseTypeName( t );
    }

    public
    final
    MingleIdentifier
    parseIdentifier( MingleToken t )
        throws SyntaxException
    {
        inputs.notNull( t, "t" );

        return
            MingleParsers.doParseIdentifier(
                (IdentifiableText) t.getObject(),
                MingleIdentifierFormat.LC_CAMEL_CAPPED,
                t.getLocation()
            );
    }

    public
    final
    MingleIdentifier
    expectIdentifier()
        throws SyntaxException
    {
        MingleToken t = expectText( "identifier" );
        
        return parseIdentifier( t );
    }

    public
    final
    MingleToken
    expectIdentifier( MingleIdentifier expct )
        throws SyntaxException
    {
        inputs.notNull( expct, "expct" );

        MingleToken t = expectToken( IdentifiableText.class, "identifier" );
 
        MingleIdentifier actual = parseIdentifier( t );
        
        if ( expct.equals( actual ) ) return t;
        else
        {
            throw syntaxException( t,
                "Expected identifier", dispNameForInstance( expct ), 
                "but got", dispNameForInstance( actual )
            );
        }
    }

    private
    List< MingleToken >
    expectNsParts()
        throws SyntaxException
    {
        List< MingleToken > parts = Lang.newList( 4 );

        String errNm = "identifier";

        parts.add( expectToken( IdentifiableText.class, errNm ) );

        for ( boolean loop = true; loop && ( ! toks.isEmpty() ); )
        {
            if ( pollLiteral( SpecialLiteral.COLON ) != null )
            {
                parts.add( expectToken( IdentifiableText.class, errNm ) );
            }
            else loop = false;
        }
        
        return parts;
    }

    private
    MingleNamespace
    buildNamespace( List< MingleToken > parts,
                    MingleIdentifier ver )
        throws SyntaxException
    {
        MingleIdentifier[] idents = new MingleIdentifier[ parts.size() ];

        int i = 0;
        for ( MingleToken part : parts )
        {
            idents[ i++ ] = 
                MingleParsers.doParseIdentifier( 
                    (IdentifiableText) part.getObject(),
                    MingleIdentifierFormat.LC_CAMEL_CAPPED, 
                    part.getLocation()
                );
        }

        return MingleParsers.IMPL_FACT.createMingleNamespace( idents, ver );
    }

    private
    MingleNamespace
    doExpectNamespace( MingleIdentifier scopedVer )
        throws SyntaxException
    {
        List< MingleToken > parts = expectNsParts();

        Object asp = scopedVer == null
            ? expectLiteral( SpecialLiteral.ASPERAND ) 
            : pollLiteral( SpecialLiteral.ASPERAND );

        MingleIdentifier ver = asp == null ? scopedVer : expectIdentifier();
        state.notNull( ver );

        return buildNamespace( parts, ver );
    }

    public
    final
    MingleNamespace
    expectNamespace()
        throws SyntaxException
    {
        return doExpectNamespace( null );
    }

    public
    final
    MingleNamespace
    expectNamespace( MingleIdentifier scopedVer )
        throws SyntaxException
    {
        return doExpectNamespace( inputs.notNull( scopedVer, "scopedVer" ) );
    }

    private
    QualifiedTypeName
    doExpectQualifiedTypeName( MingleIdentifier scopedVer )
        throws SyntaxException
    {
        MingleNamespace ns = doExpectNamespace( scopedVer );

        List< MingleTypeName > names = Lang.newList();

        checkUnexpectedEnd(); // Don't even bother if we're empty

        while ( pollLiteral( SpecialLiteral.FORWARD_SLASH ) != null )
        {
            names.add( expectTypeName() );
        }

        // Because of initial checkUnexpectedEnd() before loop, we know here
        // that either we read a name or never entered loop due to unexpected
        // input
        if ( names.isEmpty() )
        {
            throw syntaxException( toks.peek(), "Invalid type name" );
        }
        else return QualifiedTypeName.create( ns, names );
    }

    public
    final
    QualifiedTypeName
    expectQualifiedTypeName()
        throws SyntaxException
    {
        return doExpectQualifiedTypeName( null );
    }

    private
    RelativeTypeName
    expectRelativeTypeName()
        throws SyntaxException
    {
        List< MingleTypeName > names = Lang.newList( 2 );

        do { names.add( expectTypeName() ); }
        while ( pollLiteral( SpecialLiteral.FORWARD_SLASH ) != null );

        return RelativeTypeName.create( names );
    }
    
    private
    AtomicTypeReference.Name
    doExpectAtomicTypeReferenceName( MingleIdentifier scopedVer )
        throws SyntaxException
    {
        checkUnexpectedEnd();
//        IdentifiableText txt = peekInstance( IdentifiableText.class );
        MingleToken tok = peek();

        if ( tok.getObject() instanceof IdentifiableText )
        {
            IdentifiableText txt = (IdentifiableText) tok.getObject();
            
            if ( isLcAlpha( txt.charAt( 0 ) ) ) 
            {
                return doExpectQualifiedTypeName( scopedVer );
            }
            else return expectRelativeTypeName();
        }
        else throw syntaxException( tok, "Expected type reference start" );
    }
    
    public
    final
    AtomicTypeReference.Name
    expectAtomicTypeReferenceName()
        throws SyntaxException
    {
        return doExpectAtomicTypeReferenceName( null );
    }

    private
    RestrictionSyntax
    expectStringRestriction( SourceTextLocation startLoc )
        throws SyntaxException
    {
        return
            new StringRestrictionSyntax( 
                expectInstance( MingleString.class ),
                startLoc
            );
    }

    // returns true if closedLit was seen; false if openLit was seen; fails
    // otherwise
    private
    boolean
    expectRangeDelim( SpecialLiteral closedLit,
                      SpecialLiteral openLit )
        throws SyntaxException
    {
        skipWhitespace();

        MingleToken tok = expectToken( SpecialLiteral.class );
        SpecialLiteral lit = (SpecialLiteral) tok.getObject();

        if ( lit == closedLit ) return true;
        else if ( lit == openLit ) return false;
        else 
        {
            throw syntaxException( 
                tok, "Invalid range delimiter:", lit.getLiteral() );
        }
    }

    private
    MingleValue
    getNegatedRangeValue()
        throws SyntaxException
    {
        skipWhitespace();
        MingleToken tok = expectToken();
        Object obj = tok.getObject();

        if ( obj instanceof MingleInt64 )
        {
            return 
                MingleModels.asMingleInt64(
                    -( (MingleInt64) obj ).longValue() );
        }
        else if ( obj instanceof MingleDouble )
        {
            return
                MingleModels.asMingleDouble(
                    -( (MingleDouble) obj ).doubleValue() );
        }
        else 
        {
            Object errObj = obj instanceof SpecialLiteral
                ? ( (SpecialLiteral) obj ).getLiteral() : obj;

            throw syntaxException( tok, "Can't negate value:", errObj );
        }
    }

    private
    MingleValue
    getRangeValue()
        throws SyntaxException
    {
        skipWhitespace();
        MingleToken tok = peek();
        Object obj = tok.getObject();

        if ( obj instanceof MingleString ||
             obj instanceof MingleInt64 ||
             obj instanceof MingleDouble )
        {
            remove();
            return (MingleValue) obj;
        }
        else if ( obj == SpecialLiteral.MINUS ) 
        {
            remove();
            return getNegatedRangeValue();
        }
        else if ( obj == SpecialLiteral.COMMA ||
                  obj == SpecialLiteral.CLOSE_PAREN ||
                  obj == SpecialLiteral.CLOSE_BRACKET )
        {
            return null;
        }
        else throw syntaxException( tok, "Invalid range literal:", obj );
    }

    private
    RangeRestrictionSyntax
    createRangeRestriction( boolean includesLeft,
                            MingleValue leftVal,
                            MingleValue rightVal,
                            boolean includesRight,
                            SourceTextLocation startLoc )
        throws SyntaxException
    {
        boolean infLowClosed = includesLeft && leftVal == null;
        boolean infHighClosed = includesRight && rightVal == null;

        if ( infLowClosed || infHighClosed )
        {
            StringBuilder sb = new StringBuilder( "Infinite " );

            if ( infLowClosed )
            {
                if ( ! infHighClosed ) sb.append( "low " );
            }
            else sb.append( "high " );

            sb.append( "range must be open" );

            throw syntaxException( startLoc, sb );
        }

        return new RangeRestrictionSyntax(
                includesLeft, leftVal, rightVal, includesRight, startLoc );
    }

    private
    RestrictionSyntax
    expectRangeRestriction( SourceTextLocation startLoc )
        throws SyntaxException
    {
        boolean includesLeft = 
            expectRangeDelim( 
                SpecialLiteral.OPEN_BRACKET, SpecialLiteral.OPEN_PAREN );

        MingleValue leftVal = getRangeValue();
        
        skipWhitespace();
        expectLiteral( SpecialLiteral.COMMA );

        MingleValue rightVal = getRangeValue();
        
        boolean includesRight = 
            expectRangeDelim(
                SpecialLiteral.CLOSE_BRACKET, SpecialLiteral.CLOSE_PAREN );
        
        return 
            createRangeRestriction(
                includesLeft, leftVal, rightVal, includesRight, startLoc );
    }

    private
    RestrictionSyntax
    buildRestrictionSyntax( MingleToken tildeTok )
        throws SyntaxException
    {
        skipWhitespace();
        checkUnexpectedEnd();
        Object obj = peek().getObject();

        if ( obj instanceof MingleString )
        {
            return expectStringRestriction( tildeTok.getLocation() );
        }
        else if ( obj == SpecialLiteral.OPEN_PAREN ||
                  obj == SpecialLiteral.OPEN_BRACKET )
        {
            return expectRangeRestriction( tildeTok.getLocation() );
        }
        else throw syntaxException( tildeTok, "Unexpected restriction" );
    }

    public
    final
    RestrictionSyntax
    pollRestrictionSyntax()
        throws SyntaxException
    {
        skipWhitespace();

        MingleToken tok = peek();

        if ( tok != null && tok.getObject().equals( SpecialLiteral.TILDE ) )
        {
            remove(); // remove the TILDE
            return buildRestrictionSyntax( tok );
        }
        else return null;
    }
    
    public
    final
    static
    class TypeCompleter
    {
        private final List< SpecialLiteral > quants;

        private
        TypeCompleter( List< SpecialLiteral > quants )
        {
            this.quants = quants;
        }

        // Exposing now only for test code in this package
        CharSequence
        getQuantifierString()
        {
            StringBuilder sb = new StringBuilder();

            for ( SpecialLiteral quant : quants ) 
            {
                sb.append( quant.getLiteral() );
            }

            return sb;
        }

        private
        MingleTypeReference
        quantify( MingleTypeReference typ,
                  SpecialLiteral quant )
        {
            switch ( quant )
            {
                case PLUS: return ListTypeReference.create( typ, false );
                case ASTERISK: return ListTypeReference.create( typ, true );
                case QUESTION_MARK: return NullableTypeReference.create( typ );
                default: throw state.createFail( "bad quant:", quant );
            }
        }

        public
        MingleTypeReference
        completeType( MingleTypeReference typ )
        {
            inputs.notNull( typ, "typ" );

            MingleTypeReference res = typ;
            for ( SpecialLiteral quant : quants ) res = quantify( res, quant );

            return res;
        }
    }

    private
    TypeCompleter
    getTypeCompleter()
    {
        List< SpecialLiteral > quants = Lang.newList();
        
        while ( true )
        {
            skipWhitespace();
            
            SpecialLiteral lit = peekInstance( SpecialLiteral.class );

            if ( lit == SpecialLiteral.ASTERISK ||
                 lit == SpecialLiteral.QUESTION_MARK ||
                 lit == SpecialLiteral.PLUS )
            {
                remove();
                quants.add( lit );
            }
            else break;
        }

        return new TypeCompleter( quants );
    }

    public
    static
    interface TypeReferenceBuilder< V >
    {
        public
        V
        buildResult( AtomicTypeReference.Name nm,
                     RestrictionSyntax sx,
                     TypeCompleter tc )
            throws SyntaxException;
    }

    private
    < V >
    V
    doExpectTypeReference( TypeReferenceBuilder< V > bldr,
                           MingleIdentifier scopedVer )
        throws SyntaxException
    {
        checkUnexpectedEnd();

        AtomicTypeReference.Name nm = 
            doExpectAtomicTypeReferenceName( scopedVer );

        RestrictionSyntax sx = pollRestrictionSyntax();
        TypeCompleter tc = getTypeCompleter();

        return bldr.buildResult( nm, sx, tc );
    }

    public
    final
    < V >
    V
    expectTypeReference( TypeReferenceBuilder< V > bldr )
        throws SyntaxException
    {
        inputs.notNull( bldr, "bldr" );
        return doExpectTypeReference( bldr, null );
    }

    public
    final
    < V >
    V
    expectTypeReference( TypeReferenceBuilder< V > bldr,
                         MingleIdentifier scopedVer )
        throws SyntaxException
    {
        inputs.notNull( bldr, "bldr" );
        inputs.notNull( scopedVer, "scopedVer" );

        return doExpectTypeReference( bldr, scopedVer );
    }

    private
    final
    static
    class StandardTypeReferenceBuilder
    implements TypeReferenceBuilder< MingleTypeReference >
    {
        public
        MingleTypeReference
        buildResult( AtomicTypeReference.Name nm,
                     RestrictionSyntax sx,
                     TypeCompleter tc )
            throws SyntaxException
        {
            AtomicTypeReference atr;

            if ( sx == null ) atr = AtomicTypeReference.create( nm );
            else atr = MingleParsers.buildAtomicTypeReference( nm, sx );
 
            return tc.completeType( atr );
        }
    }

    public
    final
    MingleTypeReference
    expectTypeReference()
        throws SyntaxException
    {
        return 
            doExpectTypeReference( new StandardTypeReferenceBuilder(), null );
    }

    public
    final
    MingleTypeReference
    expectTypeReference( MingleIdentifier scopedVer )
        throws SyntaxException
    {
        inputs.notNull( scopedVer, "scopedVer" );

        return 
            expectTypeReference( 
                new StandardTypeReferenceBuilder(), scopedVer );
    }
}
