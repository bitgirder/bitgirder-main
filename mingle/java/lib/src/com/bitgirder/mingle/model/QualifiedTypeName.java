package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.parser.SyntaxException;

import com.bitgirder.mingle.parser.MingleParsers;

import java.util.List;
import java.util.Arrays;

public
final
class QualifiedTypeName
implements AtomicTypeReference.Name
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleNamespace ns;
    private final MingleTypeName[] name;

    private
    QualifiedTypeName( MingleNamespace ns,
                       MingleTypeName[] name )
    {
        this.ns = ns;
        this.name = name;
    }

    public MingleNamespace getNamespace() { return ns; }

    public 
    List< MingleTypeName > 
    getName() 
    { 
        return Lang.unmodifiableList( Arrays.asList( name ) ); 
    }

    public int hashCode() { return ns.hashCode() ^ Arrays.hashCode( name ); }

    public
    boolean
    equals( Object other )
    {
        if ( other == this ) return true;
        else
        {
            if ( other instanceof QualifiedTypeName )
            {
                QualifiedTypeName n2 = (QualifiedTypeName) other;

                return ns.equals( n2.ns ) && Arrays.equals( name, n2.name );
            }
            else return false;
        }
    }

    public
    CharSequence
    getExternalForm()
    {
        return 
            new StringBuilder( ns.getExternalForm() ).
                append( "/" ).
                append( Strings.join( "/", (Object[]) name ) );
    }

    @Override public String toString() { return getExternalForm().toString(); }

    public
    static
    QualifiedTypeName
    create( MingleNamespace ns,
            List< MingleTypeName > name )
    {
        inputs.notNull( ns, "ns" );
        inputs.isFalse( name.isEmpty(), "No name elements in qname" );
 
        MingleTypeName[] nameArr =
            inputs.noneNull( name, "name" ).toArray(
                new MingleTypeName[ name.size() ] );
        
        return new QualifiedTypeName( ns, nameArr );
    }

    public
    static
    QualifiedTypeName
    create( MingleNamespace ns,
            MingleTypeName... name )
    {
        return create( ns, Lang.asList( inputs.notNull( name, "name" ) ) );
    }

    // Used by other classes in this package to create qname views of other data
    // that is known to be non-null and immutable
    static
    QualifiedTypeName
    createUnsafe( MingleNamespace ns,
                  MingleTypeName[] name ) 
    {
        return new QualifiedTypeName( ns, name );
    }

    public
    static
    QualifiedTypeName
    create( CharSequence str )
    {
        return 
            MingleParsers.
                createQualifiedTypeName( inputs.notNull( str, "str" ) );
    }

    public
    static
    QualifiedTypeName
    parse( CharSequence str )
        throws SyntaxException
    {
        return
            MingleParsers.
                parseQualifiedTypeName( inputs.notNull( str, "str" ) );
    }
}
