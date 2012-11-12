package com.bitgirder.test;

import java.lang.annotation.Target;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Retention;
import java.lang.annotation.ElementType;

@Retention( RetentionPolicy.RUNTIME )
@Target( { ElementType.METHOD, ElementType.TYPE } )
public
@interface Test
{
    // Used as default for expected(), since we can't use 'null'
    public
    final
    static
    class NoneExpectedException
    extends Exception
    {
        private NoneExpectedException() {}
    }

    public
    Class< ? extends Throwable >
    expected()
    default NoneExpectedException.class;

    public
    String
    expectedPattern()
    default ".*";

//    String
//    timeout()
//    default "30s";

    @Retention( RetentionPolicy.RUNTIME )
    @Target( { ElementType.METHOD } )
    public
    static
    @interface Constructor
    {}
}
