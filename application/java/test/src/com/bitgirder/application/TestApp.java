package com.bitgirder.application;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

final
class TestApp
extends ApplicationProcess
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static String TEST_CODE_MESSAGE = "test-code-message";
    private final static String TEST_WARN_MESSAGE = "test-warn-message";

    private
    TestApp( Configurator c )
    {
        super( c );
    }
    
    public
    int
    execute()
    {
        CodeLoggers.code( TEST_CODE_MESSAGE );
        CodeLoggers.warn( TEST_WARN_MESSAGE );
        
        return 0;
    }

    private
    final
    static
    class Configurator
    extends ApplicationProcess.Configurator
    {}
}
