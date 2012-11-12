package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

public
final
class QualifiedTypeName
implements AtomicTypeReference.Name
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleNamespace ns;
    private final MingleDeclaredTypeName name;

    QualifiedTypeName( MingleNamespace ns,
                       MingleDeclaredTypeName name )
    {
        this.ns = inputs.notNull( ns, "ns" );
        this.name = inputs.notNull( name, "name" );
    }

    public MingleNamespace getNamespace() { return ns; }
    public MingleDeclaredTypeName getName() { return name; }

    public int hashCode() { return ns.hashCode() ^ name.hashCode(); }

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

                return ns.equals( n2.ns ) && name.equals( n2.name );
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
                append( name.getExternalForm() );
    }

    @Override public String toString() { return getExternalForm().toString(); }

    public
    static
    QualifiedTypeName
    create( MingleNamespace ns,
            MingleDeclaredTypeName name )
    {
        inputs.notNull( ns, "ns" );
        inputs.notNull( name, "name" );

        return new QualifiedTypeName( ns, name );
    }

    public
    static
    QualifiedTypeName
    create( CharSequence str )
    {
        throw new UnsupportedOperationException( "Unimplemented" );
    }

    public
    static
    QualifiedTypeName
    parse( CharSequence str )
        throws MingleSyntaxException
    {
        throw new UnsupportedOperationException( "Unimplemented" );
    }
}
