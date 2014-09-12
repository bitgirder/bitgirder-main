package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;

import java.util.Arrays;
import java.util.List;

public
final
class MingleNamespace
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final MingleIdentifier[] parts;
    private final MingleIdentifier ver;

    MingleNamespace( MingleIdentifier[] parts,
                     MingleIdentifier ver ) 
    { 
        this.parts = inputs.noneNull( parts, "parts" );
        this.ver = inputs.notNull( ver, "ver" );
    }

    public
    List< MingleIdentifier >
    getParts()
    {
        return Lang.unmodifiableList( Lang.asList( parts ) );
    }

    public MingleIdentifier getVersion() { return ver; }

    public int hashCode() { return Arrays.hashCode( parts ) | ver.hashCode(); }

    public
    boolean
    equals( Object other )
    {
        if ( other == this ) return true;
        else if ( other instanceof MingleNamespace )
        {
            MingleNamespace ns2 = (MingleNamespace) other;

            return Arrays.equals( parts, ns2.parts ) && ver.equals( ns2.ver );
        }
        else return false;
    }

    private
    CharSequence
    fmtId( MingleIdentifier id,
           MingleIdentifierFormat fmt )
    {
        return id.format( fmt );
    }

    public
    String
    format( MingleIdentifierFormat fmt )
    {
        inputs.notNull( fmt, "fmt" );
        
        StringBuilder sb = new StringBuilder();

        for ( int i = 0, e = parts.length; i < e; ++i )
        {
            sb.append( fmtId( parts[ i ], fmt ) );
            
            if ( i < e - 1 ) sb.append( ':' );
        }

        sb.append( '@' );
        sb.append( fmtId( ver, fmt ) );

        return sb.toString();
    }

    public
    String
    getExternalForm()
    {
        return format( MingleIdentifierFormat.LC_CAMEL_CAPPED );
    }

    @Override 
    public final String toString() { return getExternalForm(); }

    public
    static
    MingleNamespace
    create( CharSequence str )
    {
        inputs.notNull( str, "str" );
        return MingleParser.createNamespace( str );
    }

    public
    static
    MingleNamespace
    parse( CharSequence str )
        throws MingleSyntaxException
    {
        inputs.notNull( str, "str" );
        return MingleParser.parseNamespace( str );
    }
}
