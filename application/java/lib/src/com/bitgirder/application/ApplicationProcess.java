package com.bitgirder.application;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.lang.annotation.Target;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.ElementType;

public
abstract
class ApplicationProcess
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();
    
    protected ApplicationProcess( Configurator c ) {}

    public
    abstract
    int
    execute()
        throws Exception;

    public
    static
    abstract
    class Configurator
    {
        protected Configurator() {}

        @Retention( value = RetentionPolicy.RUNTIME )
        @Target( { ElementType.METHOD } )
        public
        static
        @interface Argument
        {}
    }
}
