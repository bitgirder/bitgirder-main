package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;
import java.util.Arrays;

public
final
class RelativeTypeName
implements AtomicTypeReference.Name
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleTypeName[] name;

    private RelativeTypeName( MingleTypeName[] name ) { this.name = name; }

    public
    QualifiedTypeName
    resolveIn( MingleNamespace ns )
    {
        inputs.notNull( ns, "ns" );
        return QualifiedTypeName.createUnsafe( ns, name );
    }

    public
    List< MingleTypeName >
    getNames()
    {
        return Lang.unmodifiableList( Lang.asList( name ) );
    }

    @Override public int hashCode() { return Arrays.hashCode( name ); }

    @Override
    public
    boolean
    equals( Object other )
    {
        if ( other == this ) return true;
        else if ( other instanceof RelativeTypeName )
        {
            return Arrays.equals( name, ( (RelativeTypeName) other ).name );
        }
        else return false;
    }

    public
    CharSequence
    getExternalForm()
    {
        StringBuilder res = new StringBuilder();

        for ( int i = 0, e = name.length; i < e; )
        {
            res.append( name[ i ].getExternalForm() );
            if ( ++i < e ) res.append( '/' );
        }

        return res;
    }

    @Override public String toString() { return getExternalForm().toString(); }

    public
    static
    RelativeTypeName
    create( List< MingleTypeName > name )
    {
        return
            new RelativeTypeName(
                inputs.noneNull( name, "name" ).toArray(
                    new MingleTypeName[ name.size() ] ) );
    }

    public
    static
    RelativeTypeName
    create( MingleTypeName... name )
    {
        return create( Lang.asList( inputs.notNull( name, "name" ) ) );
    }
}
