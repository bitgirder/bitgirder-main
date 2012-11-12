package com.bitgirder.lang.reflect;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.Map;

import java.lang.reflect.Member;

public
abstract
class AbstractMethodInvocation< M extends Member >
extends AbstractInvocation< M >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    // For reflection parameters that are of a primitive type we need to supply
    // the primitive default instead of the Object default (null), which will
    // lead to IllegalArgumentException.
    private final static Map< Class, Object > PRIMITIVE_DEFAULTS;

    private final Map< Object, Integer > indices;
    private final Object[] defaults;
    private final int firstParameterIndex;

    AbstractMethodInvocation( Builder< M, ?, ? > b )
    {
        super( b );
        
        this.indices = Lang.unmodifiableCopy( b.indices );

        this.defaults = b.defaultsArr; 
        this.firstParameterIndex = b.firstParameterIndex;
    }

    public final Iterable< Object > getKeys() { return indices.keySet(); }

    public 
    final
    boolean 
    hasKey( Object key ) 
    { 
        return indices.containsKey( inputs.notNull( key, "key" ) );
    }

    public
    final
    int
    indexOf( Object key )
    {
        Integer res = indices.get( inputs.notNull( key, "key" ) );

        if ( res == null ) throw state.createFail( "No mapping for key:", key );
        else return res.intValue();
    }

    private
    Object[]
    getInvokeArgs( Map< Object, Object > params )
    {
        Object[] res = new Object[ defaults.length ];
        System.arraycopy( defaults, 0, res, 0, res.length );

        for ( Map.Entry< Object, Object > e : params.entrySet() )
        {
            Integer indx = indices.get( e.getKey() );

            if ( indx == null ) keyUnmatched( e.getKey() );
            else res[ indx + firstParameterIndex ] = e.getValue();
        }

        return res;
    }

    abstract
    Object
    doInvoke( Object[] args )
        throws Exception;

    final
    Object
    invokeImpl( Map< Object, Object > params )
        throws Exception
    {
        Object[] args = getInvokeArgs( params );

        return doInvoke( args );
    }

    public
    static
    abstract
    class Builder< M extends Member, 
                   I extends AbstractInvocation< M >, 
                   B extends Builder< M, I, B > >
    extends AbstractInvocation.Builder< M, I, B >
    {
        private final Map< Object, Integer > indices = Lang.newMap();
        private final Map< Integer, Object > defaults = Lang.newMap();

        private Object[] defaultsArr; // see initDefaultsArr()
        private int firstParameterIndex;

        private
        B
        doSetKey( Object key,
                  int indx,
                  Object defl )
        {
            inputs.nonnegativeI( indx, "indx" );
            inputs.notNull( key, "key" );

            inputs.isFalse( 
                indices.values().contains( indx ),
                "Attempt to set multiple mappings at index", indx );

            Lang.putUnique( indices, key, indx );

            if ( defl != null ) defaults.put( indx, defl );

            return castThis();
        }

        public
        B
        setKey( Object key,
                int indx )
        {
            return doSetKey( key, indx, null );
        }

        public
        B
        setKey( Object key,
                int indx,
                Object defl )
        {
            return doSetKey( key, indx, inputs.notNull( defl, "defl" ) );
        }

        private
        void
        initDefaultsArr( Class< ? >[] paramTypes )
        {
            defaultsArr = new Object[ paramTypes.length ];

            for ( int i = 0, e = paramTypes.length; i < e; ++i )
            {
                Object defl = defaults.get( i );

                if ( defl == null ) 
                {
                    defl = PRIMITIVE_DEFAULTS.get( paramTypes[ i ] );
                }

                defaultsArr[ i ] = defl; // could still be null
            }
        }

        abstract
        int
        getFirstParameterIndex( M member );

        abstract
        Class< ? >[]
        getParameterTypes( M member );

        abstract
        I
        completeBuild();

        public 
        final
        I
        build() 
        { 
            M member = expectTarget();

            firstParameterIndex = getFirstParameterIndex( member );
            initDefaultsArr( getParameterTypes( member ) );
            return completeBuild();
        }
    }

    static
    {
        PRIMITIVE_DEFAULTS =
            Lang.unmodifiableMap(
                Lang.newMap( Class.class, Object.class,
                    Byte.TYPE, Byte.valueOf( (byte) 0 ),
                    Short.TYPE, Short.valueOf( (short) 0 ),
                    Character.TYPE, Character.valueOf( (char) 0 ),
                    Integer.TYPE, Integer.valueOf( 0 ),
                    Long.TYPE, Long.valueOf( 0 ),
                    Float.TYPE, Float.valueOf( 0.0f ),
                    Double.TYPE, Double.valueOf( 0.0d ),
                    Boolean.TYPE, Boolean.FALSE ) );
    }
}
