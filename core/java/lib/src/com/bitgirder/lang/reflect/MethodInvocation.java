package com.bitgirder.lang.reflect;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.lang.reflect.Method;

public
final
class MethodInvocation
extends AbstractMethodInvocation< Method >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private MethodInvocation( Builder b ) { super( b ); }

    Object
    doInvoke( Object[] args )
        throws Exception
    {
        return ReflectUtils.invoke( getTarget(), instance(), args );
    }

    public
    static
    final
    class Builder
    extends AbstractMethodInvocation.Builder< Method, 
                                              MethodInvocation, 
                                              Builder >
    {
        int
        getFirstParameterIndex( Method m )
        { 
            return 0; 
        }

        Class< ? >[] 
        getParameterTypes( Method m ) 
        { 
            return m.getParameterTypes();
        }

        MethodInvocation
        completeBuild() 
        { 
            return new MethodInvocation( this ); 
        }
    }
}
