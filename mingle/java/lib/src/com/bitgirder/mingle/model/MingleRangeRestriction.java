package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Range;

import com.bitgirder.lang.path.ObjectPath;

import java.util.Arrays;

public
final
class MingleRangeRestriction< V extends MingleValue & Comparable< V > >
extends MingleValueRestriction
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final Range< V > range;
    private final Class< V > cls;

    private 
    MingleRangeRestriction( Range< V > range,
                            Class< V > cls ) 
    { 
        this.range = range; 
        this.cls = cls;
    }

    private
    void
    appendVal( StringBuilder sb,
               MingleValue mv )
    {
        if ( cls.equals( MingleTimestamp.class ) )
        {
            mv = 
                MingleModels.asMingleString( 
                    ( (MingleTimestamp) mv ).getRfc3339String() );
        }

        MingleModels.appendInspection( sb, mv );
    }

    void
    appendExternalForm( StringBuilder sb )
    {
        sb.append( range.includesMin() ? "[" : "(" );
        if ( range.min() != null ) appendVal( sb, range.min() );
        sb.append( "," );
        if ( range.max() != null ) appendVal( sb, range.max() );
        sb.append( range.includesMax() ? "]" : ")" );
    }

    public
    void
    validate( MingleValue mv,
              ObjectPath< MingleIdentifier > path )
    {
        if ( ! range.includes( cls.cast( mv ) ) )
        {
            throw 
                new MingleValidationException( 
                    "Value is not in range " + getExternalForm() + ": " +
                    MingleModels.inspect( mv ), 
                    path 
                );
        }
    }

    // We may choose later to handcode all or some of this; for now we just
    // convert obj to its default mingle value, ensure that it is testable
    // against this range instance, and return the comparison result
    public
    boolean
    acceptsJavaValue( Object obj )
    {
        inputs.notNull( obj, "obj" );

        return range.includes(
            cls.cast( 
                MingleModels.asMingleInstance(
                    MingleModels.typeReferenceOf( cls ),
                    MingleModels.asMingleValue( obj ),
                    ObjectPath.< MingleIdentifier >getRoot()
                )
            )
        );
    }

    public
    int
    hashCode()
    {
        return Arrays.hashCode( new Object[] { range.min(), range.max() } );
    }

    public
    boolean
    equals( Object other )
    {
        if ( other == this ) return true;
        else if ( other instanceof MingleRangeRestriction )
        {
            MingleRangeRestriction o = (MingleRangeRestriction) other;

            return 
                o.range.includesMin() == range.includesMin() &&
                o.range.includesMax() == range.includesMax() &&
                Arrays.equals(
                    new Object[] { o.range.min(), o.range.max() },
                    new Object[] { range.min(), range.max() }
                );
        }
        else return false;
    }

    private
    static
    < V extends MingleValue & Comparable< V > >
    MingleRangeRestriction< V >
    createImpl( boolean closedMin,
                MingleValue min,
                MingleValue max,
                boolean closedMax,
                Class< V > cls )
    {
        V vMin = cls.cast( min );
        V vMax = cls.cast( max );

        Range< V > range;
        
        if ( closedMin )
        {
            if ( closedMax ) range = Range.closed( vMin, vMax );
            else range = Range.openMax( vMin, vMax );
        }
        else
        {
            if ( closedMax ) range = Range.openMin( vMin, vMax );
            else range = Range.open( vMin, vMax );
        }

        return new MingleRangeRestriction< V >( range, cls );
    }

    static
    MingleRangeRestriction< ? >
    create( boolean closedMin,
            MingleValue min,
            MingleValue max,
            boolean closedMax,
            Class< ? > cls )
    {
        inputs.notNull( cls, "cls" );

        if ( cls.equals( MingleTimestamp.class ) )
        {
            return 
                createImpl( 
                    closedMin, min, max, closedMax, MingleTimestamp.class );
        }
        if ( cls.equals( MingleInt64.class ) )
        {
            return 
                createImpl( closedMin, min, max, closedMax, MingleInt64.class );
        }
        if ( cls.equals( MingleInt32.class ) )
        {
            return 
                createImpl( closedMin, min, max, closedMax, MingleInt32.class );
        }
        if ( cls.equals( MingleDouble.class ) )
        {
            return 
                createImpl( 
                    closedMin, min, max, closedMax, MingleDouble.class );
        }
        if ( cls.equals( MingleFloat.class ) )
        {
            return 
                createImpl( closedMin, min, max, closedMax, MingleFloat.class );
        }
        else throw inputs.createFail( "Range not supported for", cls );
    }
}
