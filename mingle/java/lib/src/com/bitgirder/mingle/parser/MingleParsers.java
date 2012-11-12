package com.bitgirder.mingle.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.PatternHelper;
import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.io.Charsets;

import com.bitgirder.mingle.model.MingleNamespace;
import com.bitgirder.mingle.model.MingleTypeName;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleIdentifierFormat;
import com.bitgirder.mingle.model.MingleIdentifiedName;
import com.bitgirder.mingle.model.MingleEnum;
import com.bitgirder.mingle.model.MingleTimestamp;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.ImplFactory;
import com.bitgirder.mingle.model.QualifiedTypeName;
import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.MingleRegexRestriction;
import com.bitgirder.mingle.model.MingleRangeRestriction;
import com.bitgirder.mingle.model.MingleValueRestriction;
import com.bitgirder.mingle.model.MingleValidationException;
import com.bitgirder.mingle.model.PrimitiveDefinition;

import com.bitgirder.parser.SyntaxException;
import com.bitgirder.parser.SourceTextLocation;

import java.util.List;
import java.util.Queue;
import java.util.Iterator;

import java.util.regex.Pattern;
import java.util.regex.PatternSyntaxException;
import java.util.regex.Matcher;

import java.nio.ByteBuffer;

public
final
class MingleParsers
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static ObjectPath< MingleIdentifier > ROOT_PATH =
        ObjectPath.getRoot();

    final static String UNEXPECTED_END_MSG = "Unexpected end of input";

    private final static class FactoryAccessor {}

    final static ImplFactory IMPL_FACT =
        new ImplFactory( new FactoryAccessor() );

    private final static Pattern INTEGER_PAT =
        PatternHelper.compile( "^-?\\d+$" );

    // patterns and capture groups used for by parseObjectPath()
    private final static Pattern WS_MATCHER = PatternHelper.compile( "\\s+" );
    private final static Pattern DOT_MATCHER = PatternHelper.compile( "\\." );

    private final static Pattern SINGLE_SLASH_MATCHER = 
        PatternHelper.compile( "/" );
 
    // Captures strings expected to be an identifier optionally followed by a
    // list index. In fact this pattern will capture illegal strings too, for
    // isntance 9foo[ 3 ] ("9foo" is not an identifier), but these will be
    // caught by the parser when attempting to parse the identifier.
    private final static Pattern PATH_ELEMENT_MATCHER =
        PatternHelper.compile( "^([^\\[]+)(?:\\[(\\d+)\\])?$" );

    // capture groups from PATH_ELEMENT_MATCHER above
    private final static int PATH_ELEMENT_PARAM_GROUP = 1;
    private final static int PATH_ELEMENT_LIST_INDEX_GROUP = 2;

    // This regex is actually more permissive than the rfc since it does not
    // check things such as valid days of month, etc. We can tighten this up as
    // our needs dictate it, enforcing semantic checks either in the parser or
    // in the MingleTimestamp initialization code.
    private final static Pattern STRICT_RFC3339_TIMESTAMP_PATTERN =
        PatternHelper.compile(
            "(\\d{4})-(\\d{2})-(\\d{2})[Tt](\\d{2}):(\\d{2}):(\\d{2})" +
            "(?:\\.(\\d+))?(?:([zZ])|([+\\-]\\d{2}:\\d{2}))"
        );

    private final static int RFC3339_GROUP_YEAR = 1;
    private final static int RFC3339_GROUP_MONTH = 2;
    private final static int RFC3339_GROUP_DATE = 3;
    private final static int RFC3339_GROUP_HOUR = 4;
    private final static int RFC3339_GROUP_MINUTE = 5;
    private final static int RFC3339_GROUP_SECONDS = 6;
    private final static int RFC3339_GROUP_FRAC_PART = 7;
    private final static int RFC3339_GROUP_TIME_ZONE_ZULU = 8;
    private final static int RFC3339_GROUP_TIME_ZONE_UTC_OFFSET = 9;

    private MingleParsers() {}

    private
    static
    String
    makeMessage( Object... msg )
    {
        return Strings.join( " ", msg ).toString();
    }

    static
    SyntaxException
    syntaxException( Object... msg )
    {
        return new SyntaxException( makeMessage( msg ) );
    }

    static
    SyntaxException
    syntaxException( SourceTextLocation loc,
                     Object... msg )
    {
        state.notNull( loc, "loc" );
        return new SyntaxException( makeMessage( msg ), loc );
    }

    private
    static
    SyntaxException
    syntaxException( MingleToken tok,
                     Object... msg )
    {
        return syntaxException( tok.getLocation(), msg );
    }

    // indx is 0-based, anchorLoc may be null. If anchorLoc is null we
    // synthesize it and add 1 to indx; otherwise we assume that anchorLoc is
    // positioned appropriately such that we can simply add indx to its col val
    // to get the correct error column
    private
    static
    SyntaxException
    syntaxException( SourceTextLocation anchorLoc,
                     int indx,
                     Object... msg )
    {
        SourceTextLocation loc;

        if ( anchorLoc == null )
        {
            loc = SourceTextLocation.create( "<>", 1, indx + 1 );
        }
        else
        {
            loc =
                SourceTextLocation.create(
                    anchorLoc.getFileName(),
                    anchorLoc.getLine(),
                    anchorLoc.getColumn() + indx
                );
        }

        return syntaxException( loc, msg );
    }

    // takes 0-based indx, displays as 1-based
    private
    static
    SyntaxException
    syntaxException( int indx,
                     Object... msg )
    {
        return syntaxException( null, indx, msg );
    }

    static
    < V >
    V
    checkTrailingInput( V res,
                        MingleSyntaxScanner ss )
        throws SyntaxException
    {
        if ( ss.isEmpty() ) return res;
        else throw syntaxException( ss.remove(), "Trailing input" );
    }

    private
    static
    IllegalArgumentException
    createRethrow( CharSequence str,
                   SyntaxException se )
    {
        return 
            new IllegalArgumentException( 
                new StringBuilder().
                    append( "Invalid syntax in input '" ).
                    append( str ).
                    append( "': " ).
                    append( se.getMessage() ).
                    toString(), 
                se 
            );
    }

    private
    static
    MingleSyntaxScanner
    scannerFor( CharSequence str )
        throws SyntaxException
    {
        Queue< MingleToken > toks = MingleLexer.tokenizeString( str );

        MingleSyntaxScanner res = new MingleSyntaxScanner( toks );

        if ( res.isEmpty() ) throw syntaxException( -1, UNEXPECTED_END_MSG );
        else return res;
    }

    private
    static
    MingleNamespace
    doParseNamespace( CharSequence nsStr,
                      MingleIdentifier scopedVer )
        throws SyntaxException
    {
        inputs.notNull( nsStr, "nsStr" );

        MingleSyntaxScanner ss = scannerFor( nsStr );

        MingleNamespace res = scopedVer == null
            ? ss.expectNamespace() : ss.expectNamespace( scopedVer );

        return checkTrailingInput( res, ss );
    }

    public
    static
    MingleNamespace
    parseNamespace( CharSequence nsStr )
        throws SyntaxException
    {
        return doParseNamespace( nsStr, null );
    }

    public
    static
    MingleNamespace
    parseNamespace( CharSequence nsStr,
                    MingleIdentifier scopedVer )
        throws SyntaxException
    {
        inputs.notNull( scopedVer, "scopedVer" );
        return doParseNamespace( nsStr, scopedVer );
    }
 
    public
    static
    MingleNamespace
    createNamespace( CharSequence nsStr )
    {
        try { return doParseNamespace( nsStr, null ); }
        catch ( SyntaxException se ) { throw createRethrow( nsStr, se ); }
    }
 
    public
    static
    MingleNamespace
    createNamespace( CharSequence nsStr,
                     MingleIdentifier scopedVer )
    {
        inputs.notNull( scopedVer, "scopedVer" );

        try { return doParseNamespace( nsStr, scopedVer ); }
        catch ( SyntaxException se ) { throw createRethrow( nsStr, se ); }
    }

    private
    static
    MingleIdentifier[]
    expectIdentifiedNamePath( MingleSyntaxScanner ss )
        throws SyntaxException
    {
        List< MingleIdentifier > res = Lang.newList( 4 );

        ss.checkUnexpectedEnd();
        while ( ss.pollLiteral( SpecialLiteral.FORWARD_SLASH ) != null )
        {
            res.add( ss.expectIdentifier() );
        }

        // See note in expectQualifiedTypeName()
        if ( res.isEmpty() )
        {
            throw syntaxException( ss.peek(), "Missing name path" );
        }
        else return res.toArray( new MingleIdentifier[ res.size() ] );
    }

    public
    static
    MingleIdentifiedName
    parseIdentifiedName( CharSequence nmStr )
        throws SyntaxException
    {
        inputs.notNull( nmStr, "nmStr" );

        MingleSyntaxScanner ss = scannerFor( nmStr );

        MingleNamespace ns = ss.expectNamespace();
        MingleIdentifier[] ids = expectIdentifiedNamePath( ss );

        return 
            checkTrailingInput(
                IMPL_FACT.createIdentifiedNameUnsafe( ns, ids ),
                ss 
            );
    }

    public
    static
    MingleIdentifiedName
    createIdentifiedName( CharSequence nmStr )
    {
        try { return parseIdentifiedName( nmStr ); }
        catch ( SyntaxException se ) { throw createRethrow( nmStr, se ); }
    }

    // returns true if scan should continue; false if it should stop; throws
    // syntax exception if something amiss is discovered
    private
    static
    boolean
    scanTypeNameChar( CharSequence str,
                      SourceTextLocation anchorLoc,
                      int indx,
                      boolean isStart )
        throws SyntaxException
    {
        char ch = str.charAt( indx );

        Boolean res;

        if ( isUcAlpha( ch ) ) res = Boolean.valueOf( isStart );
        else
        {
            if ( isLcAlpha( ch ) || isDigit( ch ) )
            {
                if ( isStart ) res = null; else res = Boolean.TRUE;
            }
            else res = null;
        }

        if ( res == null )
        {
            throw 
                syntaxException( anchorLoc, indx, 
                    "Type name segments must start with an upper case char, "
                    + "got:", ch );
        }
        else return res.booleanValue();
    }

    private
    static
    int
    scanTypeNamePart( CharSequence str,
                      SourceTextLocation anchorLoc,
                      int indx,
                      List< String > parts )
        throws SyntaxException
    {
        int start = indx;
 
        boolean loop = true;
        for ( int e = str.length(); indx < e && loop; )
        {
            if ( scanTypeNameChar( str, anchorLoc, indx, indx == start ) ) 
            {
                ++indx; 
            }
            else loop = false;
        }

        state.isTrue( indx > start );
        parts.add( str.subSequence( start, indx ).toString() );

        return indx;
    }

    public
    static
    MingleTypeName
    parseTypeName( CharSequence nmStr,
                   SourceTextLocation anchorLoc )
        throws SyntaxException
    {
        if ( nmStr.length() > 0 )
        {
            List< String > parts = Lang.newList( 3 );

            int i = 0;
            for ( int e = nmStr.length(); i < e; )
            {
                i = scanTypeNamePart( nmStr, anchorLoc, i, parts );
            }

            String[] arr = parts.toArray( new String[ parts.size() ] );
            return IMPL_FACT.createMingleTypeName( arr );
        }
        else throw syntaxException( anchorLoc, -1, "Empty string" );
    }

    public
    static
    MingleTypeName
    parseTypeName( CharSequence nmStr )
        throws SyntaxException
    {
        inputs.notNull( nmStr, "nmStr" );
        return parseTypeName( nmStr, null );
    }

    public
    static
    MingleTypeName
    createTypeName( CharSequence nmStr )
    {
        try { return parseTypeName( nmStr ); }
        catch ( SyntaxException se ) { throw createRethrow( nmStr, se ); }
    }
    
    static boolean isUcAlpha( char ch ) { return ch >= 'A' && ch <= 'Z'; }
    static boolean isLcAlpha( char ch ) { return ch >= 'a' && ch <= 'z'; }
    static boolean isDigit( char ch ) { return ch >= '0' && ch <= '9'; }

    private
    static
    void
    checkMixedIdFormats( MingleIdentifierFormat idFmt,
                         MingleIdentifierFormat idFmtExpct,
                         int indx,
                         SourceTextLocation anchorLoc )
        throws SyntaxException
    {
        if ( idFmt != null && idFmtExpct != null && idFmt != idFmtExpct ) 
        {
            throw 
                syntaxException( 
                    anchorLoc,
                    indx, 
                    "Mixed identifier formats:", idFmt, "and", idFmtExpct
                );
        }
    }

    private
    static
    boolean
    isIdentifierPartEnd( CharSequence str,
                         int indx,
                         SourceTextLocation anchorLoc,
                         MingleIdentifierFormat idFmt )
        throws SyntaxException
    {
        char ch = str.charAt( indx );

        MingleIdentifierFormat idFmtExpct = null;

        if ( ch == '-' ) idFmtExpct = MingleIdentifierFormat.LC_HYPHENATED;
        else if ( ch == '_' ) idFmtExpct = MingleIdentifierFormat.LC_UNDERSCORE;
        else if ( isUcAlpha( ch ) )
        {
            idFmtExpct = MingleIdentifierFormat.LC_CAMEL_CAPPED;
        }
        else if ( ! ( isLcAlpha( ch ) || isDigit( ch ) ) )
        {
            throw
                syntaxException( 
                    anchorLoc, 
                    indx, 
                    "Expected lc-alpha or digit but got:", ch 
                );
        }

        checkMixedIdFormats( idFmt, idFmtExpct, indx, anchorLoc );

        return idFmtExpct != null;
    }

    private
    static
    String
    getIdentifierPart( CharSequence str,
                       int start,
                       int end )
    {
        int len = end - start;

        char[] part = new char[ len ];

        part[ 0 ] = Character.toLowerCase( str.charAt( start ) );

        for ( int i = 1; i < len; ++i )
        {
            part[ i ] = str.charAt( start + i );
        }

        return new String( part );
    }

    private
    static
    void
    checkIdentifierPartStart( CharSequence str,
                              int start,
                              SourceTextLocation anchorLoc,
                              List< String > parts,
                              MingleIdentifierFormat idFmt )
        throws SyntaxException
    {
        char ch = str.charAt( start );

        boolean ok;

        if ( isLcAlpha( ch ) ) ok = true;
        else
        {
            if ( isUcAlpha( ch ) )
            {
                ok = idFmt == MingleIdentifierFormat.LC_CAMEL_CAPPED &&
                     ( ! parts.isEmpty() );
            }
            else ok = false;
        }

        if ( ! ok )
        {
            throw syntaxException( 
                anchorLoc, start, "Invalid part beginning:", ch );
        }
    }

    private
    static
    int
    scanIdentifierPart( CharSequence str,
                        int indx,
                        SourceTextLocation anchorLoc,
                        List< String > parts,
                        MingleIdentifierFormat idFmt )
        throws SyntaxException
    {
        checkIdentifierPartStart( str, indx, anchorLoc, parts, idFmt );
        int start = indx++; // advance inline since start is now known good

        boolean loop = true;
        for ( int e = str.length(); indx < e && loop; )
        {
            if ( isIdentifierPartEnd( str, indx, anchorLoc, idFmt ) ) 
            {
                loop = false; 
            }
            else ++indx;
        }

        parts.add( getIdentifierPart( str, start, indx ) );

        return indx;
    }

    private
    static
    MingleIdentifierFormat
    getIdFormat( CharSequence str,
                 int indx,
                 SourceTextLocation anchorLoc )
        throws SyntaxException
    {
        char ch = str.charAt( indx );

        if ( ch == '-' ) return MingleIdentifierFormat.LC_HYPHENATED;
        else if ( ch == '_' ) return MingleIdentifierFormat.LC_UNDERSCORE;
        else if ( Character.isUpperCase( ch ) )
        {
            return MingleIdentifierFormat.LC_CAMEL_CAPPED;
        }
        else
        {
            throw 
                syntaxException( 
                    anchorLoc, indx, "Unrecognized identifier character:", ch );
        }
    }

    private
    static
    void
    checkNoTrailingChars( CharSequence str,
                          int i,
                          SourceTextLocation anchorLoc )
        throws SyntaxException
    {
        int len = str.length();
        state.isTrue( i == len ); // sanity check on method precondition
 
        char lastChar = str.charAt( i - 1 );

        if ( lastChar == '-' || lastChar == '_' )
        {
            throw syntaxException( 
                anchorLoc, i - 1, "Trailing separator:", lastChar );
        }
    }

    private
    static
    MingleIdentifier
    buildIdentifier( List< String > parts,
                     CharSequence str,
                     int finalPos,
                     SourceTextLocation anchorLoc )
        throws SyntaxException
    {
        String[] arr = parts.toArray( new String[ parts.size() ] );
        checkNoTrailingChars( str, finalPos, anchorLoc );

        return IMPL_FACT.createMingleIdentifier( arr );
    }

    private
    static
    int
    scanIdentifierRest( CharSequence str,
                        int i,
                        SourceTextLocation anchorLoc,
                        List< String > parts,
                        MingleIdentifierFormat idFmt )
        throws SyntaxException
    {
        if ( idFmt == null ) idFmt = getIdFormat( str, i, anchorLoc );
        
        for ( int e = str.length(); i < e; )
        {
            if ( idFmt != MingleIdentifierFormat.LC_CAMEL_CAPPED ) ++i;

            if ( i < e ) 
            {
                i = scanIdentifierPart( str, i, anchorLoc, parts, idFmt );
            }
        }

        return i;
    }

    static
    MingleIdentifier
    doParseIdentifier( CharSequence str,
                       MingleIdentifierFormat idFmt,
                       SourceTextLocation anchorLoc )
        throws SyntaxException
    {
        state.notNull( str, "str" );

        if ( str.length() == 0 ) 
        {
            throw syntaxException( anchorLoc, -1, "Empty string" );
        }

        List< String > parts = Lang.newList( 3 );

        int i = scanIdentifierPart( str, 0, anchorLoc, parts, null );

        if ( i < str.length() ) 
        {
            i = scanIdentifierRest( str, i, anchorLoc, parts, idFmt );
        }

        return buildIdentifier( parts, str, i, anchorLoc );
    }

    public
    static
    MingleIdentifier
    parseIdentifier( CharSequence str )
        throws SyntaxException
    {
        inputs.notNull( str, "str" );
        return doParseIdentifier( str, null, null );
    }

    public
    static
    MingleIdentifier
    createIdentifier( CharSequence str )
    {
        try { return parseIdentifier( str ); }
        catch ( SyntaxException se ) { throw createRethrow( str, se ); }
    }

    private
    static
    int
    parseInt( Matcher m,
              int group )
    {
        return Integer.parseInt( m.group( group ) );
    }

    private
    static
    void
    setTimeZone( MingleTimestamp.Builder b,
                 Matcher m )
    {
        if ( m.group( RFC3339_GROUP_TIME_ZONE_ZULU ) == null )
        {
            b.setTimeZone( 
                "GMT" + m.group( RFC3339_GROUP_TIME_ZONE_UTC_OFFSET ) );
        }
        else b.setTimeZone( "UTC" );
    }

    private
    static
    MingleTimestamp
    buildTimestamp( Matcher m )
    {
        MingleTimestamp.Builder b = new MingleTimestamp.Builder();

        b.setYear( parseInt( m, RFC3339_GROUP_YEAR ) );
        b.setMonth( parseInt( m, RFC3339_GROUP_MONTH ) );
        b.setDate( parseInt( m, RFC3339_GROUP_DATE ) );
        b.setHour( parseInt( m, RFC3339_GROUP_HOUR ) );
        b.setMinute( parseInt( m, RFC3339_GROUP_MINUTE ) );
        b.setSeconds( parseInt( m, RFC3339_GROUP_SECONDS ) );
 
        String fracPart = m.group( RFC3339_GROUP_FRAC_PART );
        if ( fracPart != null ) b.setFraction( fracPart );
        
        setTimeZone( b, m );

        return b.build();
    }

    private
    static
    MingleTimestamp
    doCreateTimestamp( CharSequence str )
    {
        inputs.notNull( str, "str" ); // do input checking for public methods

        Matcher m = STRICT_RFC3339_TIMESTAMP_PATTERN.matcher( str );
        return m.matches() ? buildTimestamp( m ) : null;
    }

    public
    static
    MingleTimestamp
    createTimestamp( CharSequence str )
    {
        try { return parseTimestamp( str ); }
        catch ( SyntaxException se ) { throw createRethrow( str, se ); }
    }

    public
    static
    MingleTimestamp
    parseTimestamp( CharSequence str )
        throws SyntaxException
    {
        MingleTimestamp res = doCreateTimestamp( str );

        if ( res == null ) 
        {
            throw new SyntaxException( "Invalid timestamp: " + str );
        }
        else return res;
    }

    public
    static
    MingleTypeReference
    doParseTypeReference( CharSequence str,
                          MingleIdentifier scopedVer )
        throws SyntaxException
    {
        inputs.notNull( str, "str" );

        MingleSyntaxScanner ss = scannerFor( str );

        MingleTypeReference res = scopedVer == null
            ? ss.expectTypeReference() : ss.expectTypeReference( scopedVer );

        return checkTrailingInput( res, ss );
    }

    public
    static
    MingleTypeReference
    parseTypeReference( CharSequence str )
        throws SyntaxException
    {
        return doParseTypeReference( str, null );
    }

    public
    static
    MingleTypeReference
    parseTypeReference( CharSequence str,
                        MingleIdentifier scopedVer )
        throws SyntaxException
    {
        inputs.notNull( scopedVer, "scopedVer" );
        return doParseTypeReference( str, scopedVer );
    }

    public
    static
    MingleTypeReference
    createTypeReference( CharSequence str )
    {
        try { return doParseTypeReference( str, null ); }
        catch ( SyntaxException se ) { throw createRethrow( str, se ); }
    }

    public
    static
    MingleTypeReference
    createTypeReference( CharSequence str,
                         MingleIdentifier scopedVer )
    {
        inputs.notNull( scopedVer, "scopedVer" );

        try { return doParseTypeReference( str, scopedVer ); }
        catch ( SyntaxException se ) { throw createRethrow( str, se ); }
    }

    public
    static
    QualifiedTypeName
    parseQualifiedTypeName( CharSequence str )
        throws SyntaxException
    {
        inputs.notNull( str, "str" );

        MingleSyntaxScanner ss = scannerFor( str );
        return checkTrailingInput( ss.expectQualifiedTypeName(), ss );
    }

    public
    static
    QualifiedTypeName
    createQualifiedTypeName( CharSequence str )
    {
        try { return parseQualifiedTypeName( str ); }
        catch ( SyntaxException se ) { throw createRethrow( str, se ); }
    }

    public
    static
    MingleEnum
    parseEnumLiteral( CharSequence str )
        throws SyntaxException
    {
        inputs.notNull( str, "str" );

        MingleSyntaxScanner ss = scannerFor( str );

        MingleTypeReference typeRef = ss.expectTypeReference();
        ss.expectLiteral( SpecialLiteral.PERIOD );
        MingleIdentifier id = ss.expectIdentifier();
        
        MingleEnum res =
            new MingleEnum.Builder().
                setType( (AtomicTypeReference) typeRef ).
                setValue( id ).
                build();

        return checkTrailingInput( res, ss );
    }

    public
    static
    MingleEnum
    createEnumLiteral( CharSequence str )
    {
        try { return parseEnumLiteral( str ); }
        catch ( SyntaxException se ) { throw createRethrow( str, se ); }
    }

    private
    static
    ObjectPath< MingleIdentifier >
    descend( ObjectPath< MingleIdentifier > path,
             String tok )
        throws SyntaxException
    {
        Matcher m = PATH_ELEMENT_MATCHER.matcher( tok );
        
        if ( m.matches() )
        {
            MingleIdentifier param =
                parseIdentifier( m.group( PATH_ELEMENT_PARAM_GROUP ) );
 
            path = path.descend( param );

            String indexStr = m.group( PATH_ELEMENT_LIST_INDEX_GROUP );

            if ( indexStr != null )
            {
                path = path.getListIndex( Integer.parseInt( indexStr ) );
            }

            return path;
        }
        else throw new SyntaxException( "Invalid path component: " + tok );
    }

    private
    static
    ObjectPath< MingleIdentifier >
    parseNonTrivialObjectPath( String wsCompressed )
        throws SyntaxException
    {
        String[] toks = DOT_MATCHER.split( wsCompressed );

        ObjectPath< MingleIdentifier > res = ObjectPath.getRoot();

        for ( int i = 0, e = toks.length; i < e; ++i )
        {
            try { res = descend( res, toks[ i ] ); }
            catch ( SyntaxException se )
            {
                throw new SyntaxException(
                    "Syntax exception in path element " + i + ": " +
                    se.getMessage(), se );
            }
        }

        return res;
    }

    // This parser may get more involved if we want to handle different types of
    // path encodings (such as elem1.elem2[ 3 ].elem3 as well as
    // elem1.elem2.3.elem3). For now we hand parse using regexes rather than run
    // this through the mingle grammar, but that choice is an implementation
    // detail
    public
    static
    ObjectPath< MingleIdentifier >
    parseObjectPath( CharSequence str )
        throws SyntaxException
    {
        inputs.notNull( str, "str" );

        // remove all whitespace (making our later parsing simpler) and split on
        // '.'
        String compressed = WS_MATCHER.matcher( str ).replaceAll( "" );

        if ( compressed.length() == 0 ) return ObjectPath.getRoot();
        else return parseNonTrivialObjectPath( compressed );
    }
 
    public
    static
    ObjectPath< MingleIdentifier >
    createObjectPath( CharSequence str )
    {
        try { return parseObjectPath( str ); }
        catch ( SyntaxException se ) { throw createRethrow( str, se ); }
    }

    private
    static
    void
    feedLexer( MingleLexer l,
               CharSequence txt )
        throws Exception
    {
        String[] lines = txt.toString().split( "\\r?\\n" );

        for ( String line : lines ) l.update( line, true );
    } 

    public
    static
    Queue< MingleToken >
    getTokens( CharSequence fileName,
               CharSequence txt )
        throws Exception
    {
        inputs.notNull( fileName, "fileName" );
        inputs.notNull( txt, "txt" );

        MingleLexer.TokenAccumulator acc = MingleLexer.createTokenAccumulator();

        MingleLexer l = 
            new MingleLexer.Builder().
                setFileName( fileName ).
                setTokenHandler( acc ).
                build();
 
        feedLexer( l, txt );
        
        return acc.getTokens();
    }

    public
    static
    Queue< MingleToken >
    getTokens( CharSequence fileName,
               ByteBuffer src )
        throws Exception
    {
        inputs.notNull( src, "src" );
        CharSequence txt = Charsets.UTF_8.asString( src );

        return getTokens( fileName, txt );
    }

    private
    static
    SyntaxException
    createFailRestriction( RestrictionSyntax sx,
                           Object... msg )
    {
        return syntaxException( sx.getLocation(), msg );
    }

    private
    static
    SyntaxException
    createFailRestrictionNotSupported( AtomicTypeReference.Name nm,
                                       RestrictionSyntax sx )
    {
        return 
            createFailRestriction( sx, "Restrictions not supported for", nm );
    }

    private
    static
    void
    checkNoIntFromDecimal( MingleTypeReference typ,
                           MingleValue mv,
                           RangeRestrictionSyntax sx )
        throws SyntaxException
    {
        if ( typ.equals( MingleModels.TYPE_REF_MINGLE_INT64 ) ||
             typ.equals( MingleModels.TYPE_REF_MINGLE_INT32 ) )
        {
            // We do this rather than an instanceof test against mv since mv
            // could be a MingleString (not really useful for number ranges, but
            // possible for timestamp ranges), and that string may or may not
            // represent an integral value
            if ( ! INTEGER_PAT.matcher( mv.toString() ).matches() )
            {
                throw createFailRestriction( sx, "Invalid integer literal" );
            }
        }
    }

    private
    static
    MingleValue
    getRangeValue( MingleTypeReference typ,
                   MingleValue mv,
                   RangeRestrictionSyntax sx )
        throws SyntaxException
    {
        if ( mv == null ) return null;
        else
        {
            checkNoIntFromDecimal( typ, mv, sx );

            try { return MingleModels.asMingleInstance( typ, mv, ROOT_PATH ); }
            catch ( MingleValidationException mve )
            {
                throw createFailRestriction( sx, mve.getMessage() );
            }
        }
    }

    private
    static
    MingleRangeRestriction< ? >
    buildRangeRestriction( PrimitiveDefinition def,
                           RangeRestrictionSyntax sx )
        throws SyntaxException
    {
        MingleTypeReference typ = AtomicTypeReference.create( def.getName() );

        try
        {
            return 
                IMPL_FACT.createRangeRestriction(
                    sx.includesLeft(),
                    getRangeValue( typ, sx.leftValue(), sx ),
                    getRangeValue( typ, sx.rightValue(), sx ),
                    sx.includesRight(),
                    def.getModelClass()
                );
        }
        catch ( SyntaxException se ) { throw se; }
        catch ( Exception ex )
        {
            throw createFailRestriction( sx, ex.getMessage() );
        }
    }

    private
    static
    MingleRangeRestriction< ? >
    buildRangeRestriction( QualifiedTypeName qn,
                           RangeRestrictionSyntax sx )
        throws SyntaxException
    {
        if ( qn.equals( PrimitiveDefinition.QNAME_INT64 ) ||
             qn.equals( PrimitiveDefinition.QNAME_INT32 ) ||
             qn.equals( PrimitiveDefinition.QNAME_DOUBLE ) ||
             qn.equals( PrimitiveDefinition.QNAME_FLOAT ) ||
             qn.equals( PrimitiveDefinition.QNAME_TIMESTAMP ) )
        {
            PrimitiveDefinition def = PrimitiveDefinition.forName( qn );
            return buildRangeRestriction( def, sx );
        }
        else 
        {
            throw createFailRestriction( sx,
                "Don't know how to apply range restriction to", qn );
        }
    }

    private
    static
    Pattern
    parseRegex( StringRestrictionSyntax sx )
        throws SyntaxException
    {
        try { return Pattern.compile( sx.getString().toString() ); }
        catch ( PatternSyntaxException pse )
        {
            StringBuilder msg = new StringBuilder( "Invalid regex: " );
            msg.append( PatternHelper.getSingleLineMessage( pse ) );
            throw createFailRestriction( sx, msg );
        }
    }

    private
    static
    MingleRegexRestriction
    buildStringRestriction( QualifiedTypeName qn,
                            StringRestrictionSyntax sx )
        throws SyntaxException
    {
        if ( qn.equals( PrimitiveDefinition.QNAME_STRING ) )
        {
            return MingleRegexRestriction.create( parseRegex( sx ) );
        }
        else 
        {
            throw createFailRestriction( 
                sx, "Don't know how to apply string restriction to", qn );
        }
    }

    private
    static
    MingleValueRestriction
    buildRestriction( QualifiedTypeName qn,
                      RestrictionSyntax sx )
        throws SyntaxException
    {
        if ( sx instanceof StringRestrictionSyntax )
        {
            return buildStringRestriction( qn, (StringRestrictionSyntax) sx );
        }
        else if ( sx instanceof RangeRestrictionSyntax )
        {
            return buildRangeRestriction( qn, (RangeRestrictionSyntax) sx );
        }
        else throw createFailRestrictionNotSupported( qn, sx );
    }

    public
    static
    AtomicTypeReference
    buildAtomicTypeReference( AtomicTypeReference.Name nm,
                              RestrictionSyntax sx )
        throws SyntaxException
    {
        inputs.notNull( nm, "nm" );
        inputs.notNull( sx, "sx" );

        if ( nm instanceof QualifiedTypeName )
        {
            QualifiedTypeName qn = (QualifiedTypeName) nm;

            MingleValueRestriction restriction = buildRestriction( qn, sx );

            return AtomicTypeReference.create( qn, restriction );
        }
        else throw createFailRestrictionNotSupported( nm, sx );
    }
}
