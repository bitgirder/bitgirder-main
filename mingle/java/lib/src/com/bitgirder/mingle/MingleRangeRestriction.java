package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

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
    private final Class< ? extends MingleValue > typeTok;

    private final int hc;

    private 
    MingleRangeRestriction( boolean minClosed,
                            MingleValue min,
                            MingleValue max,
                            boolean maxClosed,
                            Class< ? extends MingleValue > typeTok )
    { 
        this.minClosed = minClosed;
        this.min = min;
        this.max = max;
        this.maxClosed = maxClosed;
        this.typeTok = typeTok;

        this.hc = 
            Arrays.hashCode( new Object[] { minClosed, min, max, maxClosed } );
    }

    private
    void
    appendVal( StringBuilder sb,
               MingleValue mv )
    {
        if ( typeTok.equals( MingleTimestamp.class ) )
        {
            MingleTimestamp t = state.cast( MingleTimestamp.class, mv );
            Lang.appendRfc4627String( sb, t.getRfc3339String() );
        }
        else Mingle.appendInspection( sb, mv );
    }

    void
    appendExternalForm( StringBuilder sb )
    {
        sb.append( minClosed ? "[" : "(" );
        if ( min != null ) appendVal( sb, min );
        sb.append( "," );
        if ( max != null ) appendVal( sb, max );
        sb.append( maxClosed ? "]" : ")" );
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
        if ( ! ( other instanceof MingleRangeRestriction ) ) return false;
 
        MingleRangeRestriction o = (MingleRangeRestriction) other;

        return 
            typeTok.equals( o.typeTok ) && 
            minClosed == o.minClosed &&
            maxClosed == o.maxClosed &&
            equalVal( min, o.min ) &&
            equalVal( max, o.max );
    }

    private
    static
    void
    checkCreateType( MingleValue mv,
                     Class< ? extends MingleValue > cls,
                     String bound )
    {
        if ( mv == null ) return;

        state.isTruef( cls.isInstance( mv ),
            "%s value is of type %s where %s is expected",
            bound, mv.getClass().getName(), cls.getName()
        );
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
            boolean maxClosed,
            Class< ? extends MingleValue > typeTok )
    {
        inputs.notNull( typeTok, "typeTok" );
        checkCreateType( min, typeTok, "min" );
        checkCreateType( max, typeTok, "max" );
        checkClosedHasVal( minClosed, min, "min" );
        checkClosedHasVal( maxClosed, max, "max" );

        return new MingleRangeRestriction( 
            minClosed, min, max, maxClosed, typeTok );
    }
}
