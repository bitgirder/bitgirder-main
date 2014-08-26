package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Lang;

import com.bitgirder.lang.PatternHelper;

import java.util.Arrays;
import java.util.List;

import java.util.regex.Pattern;

public
final
class MingleIdentifier
implements Comparable< MingleIdentifier >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final static Pattern PART_MATCHER = 
        PatternHelper.compile( "^[a-z][a-z0-9]*$" );

    private final String[] parts; // in canonical, lowercase form

    MingleIdentifier( String[] parts ) 
    { 
        this.parts = inputs.noneNull( parts, "parts" ); 
    }

    public int hashCode() { return Arrays.hashCode( parts ); }

    // meant only for package-level read-only access
    String[] getPartsArray() { return parts; }

    public 
    List< String > 
    getParts() 
    { 
        return Lang.unmodifiableList( Arrays.asList( parts ) ); 
    }

    public
    boolean
    equals( Object other )
    {
        return
            other == this ||
            ( other instanceof MingleIdentifier &&
              Arrays.equals( parts, ( (MingleIdentifier) other ).parts ) );
    }

    private
    CharSequence
    formatLcCamelCapped()
    {
        StringBuilder res = new StringBuilder();

        res.append( parts[ 0 ] );

        for ( int i = 1, e = parts.length; i < e; ++i )
        {
            res.append( Character.toUpperCase( parts[ i ].charAt( 0 ) ) );
            res.append( parts[ i ].substring( 1 ) );
        }

        return res;
    }

    public
    CharSequence
    format( MingleIdentifierFormat idFmt )
    {
        inputs.notNull( idFmt, "idFmt" );

        switch ( idFmt )
        {
            case LC_HYPHENATED: return Strings.join( "-", parts );
            case LC_UNDERSCORE: return Strings.join( "_", parts );
            case LC_CAMEL_CAPPED: return formatLcCamelCapped();

            default: throw state.createFail( "Unrecognized format:", idFmt );
        }
    }

    public CharSequence getExternalForm() { return Strings.join( "-", parts ); }

    @Override 
    public final String toString() { return getExternalForm().toString(); }

    // we'll optimize this down the road when needed (obvious choice would be to
    // proceed lexicographically through the parts array of each)
    public
    int
    compareTo( MingleIdentifier other )
    {
        if ( other == null ) throw new NullPointerException();
        else return toString().compareTo( other.toString() );
    }

    static
    boolean
    isValidPart( CharSequence part )
    {
        inputs.notNull( part, "part" );
        return PART_MATCHER.matcher( part ).matches();
    }

    public
    static
    MingleIdentifier
    create( CharSequence str )
    {
        inputs.notNull( str, "str" );
        return MingleParser.createIdentifier( str );
    }

    public
    static
    MingleIdentifier
    parse( CharSequence str )
        throws MingleSyntaxException
    {
        inputs.notNull( str, "str" );
        return MingleParser.parseIdentifier( str );
    }

    public
    static
    MingleIdentifier
    create( Enum< ? > e )
    {
        return create( inputs.notNull( e, "e" ).name().toLowerCase() );
    }

    public
    static
    MingleIdentifier
    parse( Enum< ? > e )
        throws MingleSyntaxException
    {
        return parse( inputs.notNull( e, "e" ).name().toLowerCase() );
    }
}
