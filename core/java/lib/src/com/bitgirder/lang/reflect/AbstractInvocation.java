package com.bitgirder.lang.reflect;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.Map;

public
abstract
class AbstractInvocation< T >
implements ReflectedInvocation
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public final static boolean DEFAULT_IGNORE_UNMATCHED_KEYS = false;

    private final T target;
    private final Object inst;
    private final boolean ignoreUnmatchedKeys;

    AbstractInvocation( Builder< T, ?, ? > b )
    {
        this.target = inputs.notNull( b.target, "target" );
        this.inst = b.inst;
        this.ignoreUnmatchedKeys = b.ignoreUnmatchedKeys;
    }

    public final T getTarget() { return target; }
    final Object instance() { return inst; }

    final
    void
    keyUnmatched( Object key )
    {
        if ( ! ignoreUnmatchedKeys ) 
        {
            throw new UnmatchedParameterKeyException( key );
        }
    }

    abstract
    Object
    invokeImpl( Map< Object, Object > params )
        throws Exception;

    public
    final
    Object
    invoke( Map< Object, Object > params )
        throws Exception
    {
        inputs.noneNull( 
            inputs.notNull( params, "params" ).keySet(), "keySet" );

        return invokeImpl( params );
    }

    public
    static
    abstract
    class Builder< T,
                   I extends AbstractInvocation< T >, 
                   B extends Builder< T, I, B > >
    {
        private T target;
        private Object inst;
        private boolean ignoreUnmatchedKeys = DEFAULT_IGNORE_UNMATCHED_KEYS;

        final
        T
        expectTarget()
        {
            inputs.notNull( target, "target" );
            return target;
        }

        final
        B
        castThis()
        {
            @SuppressWarnings( "unchecked" )
            B res = (B) this;

            return res;
        }

        public
        B
        setTarget( T target )
        {
            this.target = inputs.notNull( target, "target" );
            return castThis();
        }

        public
        B
        setInstance( Object inst )
        {
            this.inst = inputs.notNull( inst, "inst" );
            return castThis();
        }

        public
        B
        setIgnoreUnmatchedKeys( boolean ignoreUnmatchedKeys )
        {
            this.ignoreUnmatchedKeys = ignoreUnmatchedKeys;
            return castThis();
        }
    }
}
