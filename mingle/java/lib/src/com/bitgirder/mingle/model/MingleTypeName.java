package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;

import com.bitgirder.mingle.parser.MingleParsers;

import com.bitgirder.parser.SyntaxException;

import java.util.Arrays;

public
final
class MingleTypeName
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    // each part already begins with a capital
    private final String[] parts;

    MingleTypeName( String[] parts )
    {
        this.parts = inputs.noneNull( parts, "parts" );
    }

    public int hashCode() { return Arrays.hashCode( parts ); }

    public
    boolean
    equals( Object other )
    {
        return
            other == this ||
            ( other instanceof MingleTypeName &&
              Arrays.equals( parts, ( (MingleTypeName) other ).parts ) );
    }

    public CharSequence getExternalForm() { return Strings.join( "", parts ); }

    @Override 
    public final String toString() { return getExternalForm().toString(); }

    public
    QualifiedTypeName
    resolveIn( MingleNamespace ns )
    {
        inputs.notNull( ns, "ns" );

        return 
            QualifiedTypeName.createUnsafe( ns, new MingleTypeName[] { this } );
    }

    public
    static
    MingleTypeName
    parse( CharSequence str )
        throws SyntaxException
    {
        return MingleParsers.parseTypeName( inputs.notNull( str, "str" ) );
    }

    public
    static
    MingleTypeName
    create( CharSequence str )
    {
        return MingleParsers.createTypeName( inputs.notNull( str, "str" ) );
    }
}
