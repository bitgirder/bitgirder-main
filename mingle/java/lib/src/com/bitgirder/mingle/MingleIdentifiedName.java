package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;
import java.util.Arrays;

public
final
class MingleIdentifiedName
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static MingleIdentifierFormat ID_FMT =
        MingleIdentifierFormat.LC_HYPHENATED;

    private MingleNamespace ns;
    private MingleIdentifier[] names;

    MingleIdentifiedName( MingleNamespace ns,
                          MingleIdentifier[] names )
    {
        this.ns = inputs.notNull( ns, "ns" );
        this.names = inputs.noneNull( names, "names" );
    }

    public MingleNamespace getNamespace() { return ns; }

    public
    List< MingleIdentifier >
    getNames()
    {
        return Lang.unmodifiableList( Lang.asList( names ) );
    }

    public
    CharSequence
    getExternalForm()
    {
        StringBuilder res = 
            new StringBuilder().
                append( ns.format( ID_FMT ) );
        
        for ( MingleIdentifier id : names )
        {
            CharSequence idStr = id.format( ID_FMT );

            res.append( '/' ).
                append( idStr );
        }

        return res;
    }

    @Override public String toString() { return getExternalForm().toString(); }

    public int hashCode() { return ns.hashCode() | Arrays.hashCode( names ); }

    public
    boolean
    equals( Object o )
    {
        if ( o == this ) return true;
        else if ( o instanceof MingleIdentifiedName )
        {
            MingleIdentifiedName n = (MingleIdentifiedName) o;
            return ns.equals( n.ns ) && Arrays.equals( names, n.names );
        }
        else return false;
    }

    public
    static
    MingleIdentifiedName
    create( CharSequence str )
    {
        inputs.notNull( str, "str" );
        return MingleParser.createIdentifiedName( str );
    }

    public
    static
    MingleIdentifiedName
    parse( CharSequence str )
        throws MingleSyntaxException
    {
        inputs.notNull( str, "str" );
        return MingleParser.parseIdentifiedName( str );
    }
}
