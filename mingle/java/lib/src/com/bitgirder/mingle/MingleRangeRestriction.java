package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

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

    final boolean minClosed;
    final V min;
    final V max;
    final boolean maxClosed;
    private final Class< V > typeTok;

    private final int hc;

    private 
    MingleRangeRestriction( boolean minClosed,
                            MingleValue min,
                            MingleValue max,
                            boolean maxClosed,
                            Class< V > typeTok )
    { 
        this.minClosed = minClosed;
        this.min = min == null ? null : typeTok.cast( min );
        this.max = max == null ? null : typeTok.cast( max );
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

    boolean
    implValidate( MingleValue mv )
    {
        V val = typeTok.cast( mv );

        int minCmp = min == null ? 1 : val.compareTo( min );
        int maxCmp = max == null ? -1 : val.compareTo( max );

        return ( ( minCmp > 0 ) || ( minCmp == 0 && minClosed ) ) &&
               ( ( maxCmp < 0 ) || ( maxCmp == 0 && maxClosed ) );
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
    < V extends MingleValue & Comparable< V > >
    MingleRangeRestriction< V >
    create( boolean minClosed,
            MingleValue min,
            MingleValue max,
            boolean maxClosed,
            Class< V > typeTok )
    {
        inputs.notNull( typeTok, "typeTok" );
        checkCreateType( min, typeTok, "min" );
        checkCreateType( max, typeTok, "max" );
        checkClosedHasVal( minClosed, min, "min" );
        checkClosedHasVal( maxClosed, max, "max" );

        return new MingleRangeRestriction< V >( 
            minClosed, min, max, maxClosed, typeTok );
    }

    static
    MingleRangeRestriction< ? >
    createChecked( boolean minClosed,
                   MingleValue min,
                   MingleValue max,
                   boolean maxClosed,
                   Class< ? extends MingleValue > typeTok )
    {
        state.notNull( typeTok, "typeTok" );

        if ( typeTok.equals( MingleInt32.class ) )
        {
            return create( minClosed, min, max, maxClosed, MingleInt32.class );
        }
        else if ( typeTok.equals( MingleInt64.class ) )
        {
            return create( minClosed, min, max, maxClosed, MingleInt64.class );
        }
        else if ( typeTok.equals( MingleUint32.class ) )
        {
            return create( minClosed, min, max, maxClosed, MingleUint32.class );
        }
        else if ( typeTok.equals( MingleUint64.class ) )
        {
            return create( minClosed, min, max, maxClosed, MingleUint64.class );
        }
        else if ( typeTok.equals( MingleFloat32.class ) )
        {
            return 
                create( minClosed, min, max, maxClosed, MingleFloat32.class );
        }
        else if ( typeTok.equals( MingleFloat64.class ) )
        {
            return create
                ( minClosed, min, max, maxClosed, MingleFloat64.class );
        }
        else if ( typeTok.equals( MingleString.class ) )
        {
            return create( minClosed, min, max, maxClosed, MingleString.class );
        }
        else if ( typeTok.equals( MingleTimestamp.class ) )
        {
            return 
                create( minClosed, min, max, maxClosed, MingleTimestamp.class );
        }
        else throw state.createFail( "Not a comparable value class:", typeTok );
    }
}
