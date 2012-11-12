package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.parser.MingleParsers;

import com.bitgirder.parser.SyntaxException;

import java.util.List;
import java.util.Arrays;

public
final
class MingleIdentifiedName
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private MingleNamespace ns;
    private MingleIdentifier[] names;

    private
    MingleIdentifiedName( MingleNamespace ns,
                          MingleIdentifier[] names )
    {
        this.ns = ns;
        this.names = names;
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
                append( ns.getExternalForm() );
        
        for ( MingleIdentifier id : names )
        {
            CharSequence idStr =
                MingleModels.
                    format( id, MingleIdentifierFormat.LC_CAMEL_CAPPED );

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

    private
    static
    MingleIdentifiedName
    doCreate( MingleNamespace ns,
              MingleIdentifier[] names,
              boolean doCopy )
    {
        inputs.notNull( ns, "ns" );
        inputs.noneNull( names, "names" );

        if ( doCopy )
        {
            MingleIdentifier[] arr = new MingleIdentifier[ names.length ];
            System.arraycopy( names, 0, arr, 0, names.length );

            names = arr;
        }
        
        return new MingleIdentifiedName( ns, names );
    }

    public
    static
    MingleIdentifiedName
    create( MingleNamespace ns,
            MingleIdentifier... names )
    {
        return doCreate( ns, names, true );
    }

    static
    MingleIdentifiedName
    createUnsafe( MingleNamespace ns,
                  MingleIdentifier[] names )
    {
        return doCreate( ns, names, false );
    }

    public
    static
    MingleIdentifiedName
    create( CharSequence str )
    {
        return 
            MingleParsers.createIdentifiedName( inputs.notNull( str, "str" ) );
    }

    public
    static
    MingleIdentifiedName
    parse( CharSequence str )
        throws SyntaxException
    {
        return 
            MingleParsers.parseIdentifiedName( inputs.notNull( str, "str" ) );
    }
}
