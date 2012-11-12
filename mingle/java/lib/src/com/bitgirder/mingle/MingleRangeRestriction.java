package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.path.ObjectPath;

import java.util.Arrays;

public
final
class MingleRangeRestriction
extends MingleValueRestriction
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final boolean minClosed;
    private final MingleValue min;
    private final MingleValue max;
    private final boolean maxClosed;

    private final int hc;

    private 
    MingleRangeRestriction( boolean minClosed,
                            MingleValue min,
                            MingleValue max,
                            boolean maxClosed )
    { 
        this.minClosed = minClosed;
        this.min = min;
        this.max = max;
        this.maxClosed = maxClosed;

        this.hc = 
            Arrays.hashCode( new Object[] { minClosed, min, max, maxClosed } );
    }

    private
    void
    appendVal( StringBuilder sb,
               MingleValue mv )
    {
        throw new UnsupportedOperationException( "Unimplemented" );
//        if ( cls.equals( MingleTimestamp.class ) )
//        {
//            mv = 
//                MingleModels.asMingleString( 
//                    ( (MingleTimestamp) mv ).getRfc3339String() );
//        }
//
//        MingleModels.appendInspection( sb, mv );
    }

    void
    appendExternalForm( StringBuilder sb )
    {
        throw new UnsupportedOperationException( "Unimplemented" );
//        sb.append( range.includesMin() ? "[" : "(" );
//        if ( range.min() != null ) appendVal( sb, range.min() );
//        sb.append( "," );
//        if ( range.max() != null ) appendVal( sb, range.max() );
//        sb.append( range.includesMax() ? "]" : ")" );
    }

    public
    void
    validate( MingleValue mv,
              ObjectPath< MingleIdentifier > path )
    {
        throw new UnsupportedOperationException( "Unimplemented" );
//        if ( ! range.includes( cls.cast( mv ) ) )
//        {
//            throw 
//                new MingleValidationException( 
//                    "Value is not in range " + getExternalForm() + ": " +
//                    MingleModels.inspect( mv ), 
//                    path 
//                );
//        }
    }

    // We may choose later to handcode all or some of this; for now we just
    // convert obj to its default mingle value, ensure that it is testable
    // against this range instance, and return the comparison result
    public
    boolean
    acceptsJavaValue( Object obj )
    {
        throw new UnsupportedOperationException( "Unimplemented" );
//        inputs.notNull( obj, "obj" );
//
//        return range.includes(
//            cls.cast( 
//                MingleModels.asMingleInstance(
//                    MingleModels.typeReferenceOf( cls ),
//                    MingleModels.asMingleValue( obj ),
//                    ObjectPath.< MingleIdentifier >getRoot()
//                )
//            )
//        );
    }

    public int hashCode() { return hc; }

    private
    boolean
    equalVal( MingleValue mv1,
              MingleValue mv2 )
    {
        return mv1 == null ? mv2 == null : mv1.equals( mv2 );
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
                minClosed == o.minClosed &&
                maxClosed == o.maxClosed &&
                equalVal( min, o.min ) &&
                equalVal( max, o.max );
        }
        else return false;
    }

    private
    static
    Class< ? >
    checkEffectiveClass( MingleValue min,
                         MingleValue max )
    {
        if ( min == null ) return max == null ? null : max.getClass();
        
        if ( max != null && ( ! min.getClass().equals( max.getClass() ) ) )
        {
            state.failf( "min is of type %s but max is of type %s",
                min.getClass(), max.getClass() );
        }

        return min.getClass();
    }

    private
    static
    void
    checkClosedHasVal( boolean closed,
                       MingleValue mv,
                       String name )
    {
        state.isFalsef( closed && mv == null, 
            "%s is closed but value is null", name );
    }

    static
    MingleRangeRestriction
    create( boolean minClosed,
            MingleValue min,
            MingleValue max,
            boolean maxClosed )
    {
        Class< ? > cls = checkEffectiveClass( min, max );
        checkClosedHasVal( minClosed, min, "min" );
        checkClosedHasVal( maxClosed, max, "max" );

        return new MingleRangeRestriction( minClosed, min, max, maxClosed );
    }
}
