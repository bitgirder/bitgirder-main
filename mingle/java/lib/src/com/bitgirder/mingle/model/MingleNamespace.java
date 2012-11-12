package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.parser.MingleParsers;

import com.bitgirder.parser.SyntaxException;

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
    fmtId( MingleIdentifier id )
    {
        return
            MingleModels.format( id, MingleIdentifierFormat.LC_CAMEL_CAPPED );
    }

    public
    CharSequence
    getExternalForm()
    {
        StringBuilder sb = new StringBuilder();

        for ( int i = 0, e = parts.length; i < e; ++i )
        {
            sb.append( fmtId( parts[ i ] ) );
            
            if ( i < e - 1 ) sb.append( ':' );
        }

        sb.append( '@' );
        sb.append( fmtId( ver ) );

        return sb;
    }

    @Override 
    public final String toString() { return getExternalForm().toString(); }

    public
    static
    MingleNamespace
    create( CharSequence str )
    {
        return MingleParsers.createNamespace( inputs.notNull( str, "str" ) );
    }

    public
    static
    MingleNamespace
    parse( CharSequence str )
        throws SyntaxException
    {
        return MingleParsers.parseNamespace( inputs.notNull( str, "str" ) );
    }

    public
    static
    MingleNamespace
    create( CharSequence str,
            MingleIdentifier scopedVer )
    {
        inputs.notNull( str, "str" );
        inputs.notNull( scopedVer, "scopedVer" );

        return MingleParsers.createNamespace( str, scopedVer );
    }

    public
    static
    MingleNamespace
    parse( CharSequence str,
           MingleIdentifier scopedVer )
        throws SyntaxException
    {
        inputs.notNull( str, "str" );
        inputs.notNull( scopedVer, "scopedVer" );

        return MingleParsers.parseNamespace( str, scopedVer );
    }
}
