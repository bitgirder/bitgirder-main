package com.bitgirder.lang.reflect;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.lang.reflect.Constructor;

public
final
class ConstructorInvocation
extends AbstractMethodInvocation< Constructor< ? > >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private ConstructorInvocation( Builder b ) { super( b ); }

    Object
    doInvoke( Object[] args )
        throws Exception
    {
        Object inst = instance();
        if ( inst != null ) args[ 0 ] = inst;

        return ReflectUtils.invoke( getTarget(), args );
    }

    public
    final
    static
    class Builder
    extends AbstractMethodInvocation.Builder< Constructor< ? >, 
                                              ConstructorInvocation,
                                              Builder >
    {
        int
        getFirstParameterIndex( Constructor< ? > c )
        {
            Class< ? > declCls = c.getDeclaringClass();

            boolean isInner =
                declCls.getEnclosingClass() != null &&
                ( ! ReflectUtils.isStatic( declCls ) );

            return isInner ? 1 : 0;
        }

        Class< ? >[]
        getParameterTypes( Constructor< ? > c )
        {
            return c.getParameterTypes();
        }

        ConstructorInvocation
        completeBuild() 
        { 
            return new ConstructorInvocation( this );
        }
    }
}
