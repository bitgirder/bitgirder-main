package com.bitgirder.lang.reflect;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.Map;

import java.lang.reflect.Field;
import java.lang.reflect.Constructor;

// Unlike AbstractMethodInvocation, this class doesn't deal with defaults since
// the expectation is that bean classes can just declare their defaults inline
// with their field defs or otherwise handle defaults via a PostProcessor. We
// can always add defaults at a later point if necessary.
public
final
class BeanInvocation
extends AbstractInvocation< Class< ? > >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Constructor< ? > cons;
    private final Map< Object, Setter > setters;
    private final PostProcessor postProcessor;

    private
    BeanInvocation( Builder b )
    {
        super( b );

        this.cons = inputs.notNull( b.cons, "cons" );
        this.setters = Lang.unmodifiableCopy( b.setters );
        this.postProcessor = b.postProcessor;
    }

    public
    boolean 
    hasKey( Object key )
    {
        return setters.containsKey( inputs.notNull( key, "key" ) );
    }

    public Iterable< Object > getKeys() { return setters.keySet(); }

    private
    Object
    newInstance()
        throws Exception
    {
        Object inst = instance();

        Object[] args = inst == null 
            ? ReflectUtils.EMPTY_ARG_ARRAY : new Object[] { inst };

        return ReflectUtils.invoke( cons, args );
    }

    Object
    invokeImpl( Map< Object, Object > vals )
        throws Exception
    {
        Object res = newInstance();

        for ( Map.Entry< Object, Object > e : vals.entrySet() )
        {
            Object key = e.getKey();
            
            Setter s = setters.get( key );

            if ( s == null ) keyUnmatched( key );
            else s.set( res, e.getValue() );
        }

        return postProcessor == null ? res : postProcessor.process( res );
    }

    public
    static
    interface PostProcessor
    {
        public
        Object
        process( Object val )
            throws Exception;
    }

    private
    static
    interface Setter
    {
        public
        void
        set( Object inst,
             Object val )
            throws Exception;
    }

    private
    final
    static
    class FieldSetter
    implements Setter
    {
        private final Field f;

        private FieldSetter( Field f ) { this.f = f; }

        public
        void
        set( Object inst,
             Object val )
            throws Exception
        {
            f.set( inst, val );
        }
    }

    public
    final
    static
    class Builder
    extends AbstractInvocation.Builder< Class< ? >, BeanInvocation, Builder >
    {
        private Constructor< ? > cons;
        private final Map< Object, Setter > setters = Lang.newMap();
        private PostProcessor postProcessor;

        public
        Builder
        setConstructor( Constructor< ? > cons )
        {
            this.cons = inputs.notNull( cons, "cons" );
            return this;
        }

        public
        Builder
        setTargetAndConstructor( Class< ? > cls )
            throws Exception
        {
            setTarget( inputs.notNull( cls, "cls" ) );

            boolean isInner = 
                cls.getEnclosingClass() != null &&
                ! ReflectUtils.isStatic( cls );

            int expctArgc = isInner ? 1 : 0;

            for ( Constructor< ? > cons : cls.getDeclaredConstructors() )
            {
                if ( cons.getParameterTypes().length == expctArgc )
                {
                    cons.setAccessible( true );
                    return setConstructor( cons );
                }
            }

            throw inputs.createFail( 
                "Class has no default no-arg constructor:", cls );
        }

        private
        Builder
        doSetKey( Object key,
                  Setter s )
        {
            inputs.notNull( key, "key" );
            Lang.putUnique( setters, key, s );

            return this;
        }

        private
        Builder
        doSetKey( Object key,
                  Field f )
        {
            return doSetKey( key, new FieldSetter( inputs.notNull( f, "f" ) ) );
        }

        public
        Builder
        setKey( Object key,
                Field f )
        {
            return doSetKey( key, f );
        }

        public
        Builder
        setPostProcessor( PostProcessor postProcessor )
        {
            this.postProcessor = 
                inputs.notNull( postProcessor, "postProcessor" );

            return this;
        }

        public BeanInvocation build() { return new BeanInvocation( this ); }
    }
}
