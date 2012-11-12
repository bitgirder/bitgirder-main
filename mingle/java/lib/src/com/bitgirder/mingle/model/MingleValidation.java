package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Range;

import com.bitgirder.lang.path.ObjectPath;

public
final
class MingleValidation
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private MingleValidation() {}

    // location is known not-null by the time this is called
    private
    static
    MingleValidationException
    doCreateFail( ObjectPath< MingleIdentifier > location,
                  Object... msg )
    {
        String msgStr = 
            msg == null ? null : Strings.join( " ", msg ).toString();

        return new MingleValidationException( msgStr, location );
    }

    public
    static
    MingleValidationException
    createFail( ObjectPath< MingleIdentifier > location,
                Object... msg )
    {
        return doCreateFail( inputs.notNull( location, "location" ), msg );
    }

    public
    static
    void
    isFalse( boolean val,
             ObjectPath< MingleIdentifier > location,
             Object... msg )
    {
        if ( val )
        {
            inputs.notNull( location, "location" );
            throw doCreateFail( location, msg );
        }
    }

    public
    static
    void
    isTrue( boolean val,
            ObjectPath< MingleIdentifier > location,
            Object... msg )
    {
        isFalse( ! val, location, msg );
    }

    public
    static
    long
    nonnegativeL( MingleValidator v,
                  long val )
    {
        inputs.notNull( v, "v" );

        v.isTrue( val >= 0, "value is < 0" );
        return val;
    }

    public
    static
    long
    nonnegativeL( MingleValidator v,
                  Long val,
                  long defaultVal )
    {
        if ( val == null ) 
        {
            return inputs.nonnegativeL( defaultVal, "defaultVal" );
        }
        else return nonnegativeL( v, val.longValue() );
    }

    public
    static
    int
    nonnegativeI( MingleValidator v,
                  int val )
    {
        return (int) nonnegativeL( v, (long) val );
    }

    public
    static
    int
    nonnegativeI( MingleValidator v,
                  Integer val,
                  int defaultVal )
    {
        return (int) 
            nonnegativeL( 
                v,
                val == null ? null : Long.valueOf( val.longValue() ), 
                (long) defaultVal );
    }

    public
    static
    long
    positiveL( MingleValidator v,
               long val )
    {
        inputs.notNull( v, "v" );

        v.isTrue( val > 0, "value is <= 0" );
        return val;
    }

    public
    static
    long
    positiveL( MingleValidator v,
               Long val,
               long defaultVal )
    {
        if ( val == null ) return inputs.positiveL( defaultVal, "defaultVal" );
        else return positiveL( v, val.longValue() );
    }

    public
    static
    int
    positiveI( MingleValidator v,
               int val )
    {
        inputs.notNull( v, "v" );

        v.isTrue( val > 0, "value is <= 0" );
        return val;
    }

    public
    static
    int
    positiveI( MingleValidator v,
               Integer val,
               int defaultVal )
    {
        inputs.positiveI( defaultVal, "defaultVal" );

        return val == null ? defaultVal : positiveI( v, val.intValue() );
    }

    public
    static
    < V >
    V
    inRange( MingleValidator v,
             V val,
             Range< ? super V > range )
    {
        inputs.notNull( v, "v" );
        inputs.notNull( val, "val" );
        inputs.notNull( range, "range" );

        if ( range.includes( val ) ) return val;
        else 
        {
            throw createFail( 
                v.getPath(), "value is not in range", range + ":", val );
        }
    }
}
