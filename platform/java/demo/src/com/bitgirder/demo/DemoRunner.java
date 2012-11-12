package com.bitgirder.demo;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.reflect.ReflectUtils;

import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.Processes;
import com.bitgirder.process.ProcessWaiter;
import com.bitgirder.process.ProcessContext;
import com.bitgirder.process.ProcessExecutor;

public
final
class DemoRunner
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private DemoRunner() {}

    private
    static
    < V >
    void
    runDemo( AbstractProcess< V > demo )
        throws Exception
    {
        ProcessExecutor exec = Processes.createExecutor();
        ProcessWaiter< V > w = ProcessWaiter.create();

        ProcessContext< V > ctx = Processes.createRootContext( w, exec );
        Processes.start( demo, ctx );

        w.awaitExit();
        w.get();
    }

    public
    static
    void
    run( Class< ? > demoCls )
        throws Exception
    {
        inputs.notNull( demoCls, "demoCls" );
        
        Object demo = ReflectUtils.newInstance( demoCls );

        if ( demo instanceof AbstractProcess ) 
        {
            runDemo( (AbstractProcess< ? >) demo );
        }
        else if ( demo instanceof SimpleDemo ) ( (SimpleDemo) demo ).runDemo();
        else state.fail( "Unrecognized demo object:", demo );
    }

    private
    static
    void
    runAndExit( Class< ? > cls )
    {
        try 
        {
            run( cls );
            System.exit( 0 );
        }
        catch ( Throwable th )
        {
            th.printStackTrace( System.err );
            System.exit( -1 );
        }
    }

    public
    static
    void
    main( String[] args )
        throws Exception
    {
        inputs.isTrue( 
            args.length == 1, 
            DemoRunner.class, "takes only a single argument (the class name " +
                "of the demo to run)"
        );

        Class< ? > cls = Class.forName( args[ 0 ] );

        inputs.isTrue( 
            cls.isAnnotationPresent( Demo.class ),
            "Class", cls, 
            "does not carry annotation @" + Demo.class.getName() );

        runAndExit( cls );
    }

    public
    static
    interface SimpleDemo
    {
        public
        void
        runDemo()
            throws Exception;
    }

}
