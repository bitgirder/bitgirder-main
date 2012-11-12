package com.bitgirder.application;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import java.io.PrintStream;

public
final
class Applications
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static String PROP_NAME_LOG_STREAM =
        Applications.class.getName() + ".logStream";

    private Applications() {}

    public
    static
    PrintStream
    getDefaultLogStream()
    {
        String strm = System.getProperty( PROP_NAME_LOG_STREAM );

        if ( strm == null ) return System.out;
        else
        {
            String check = strm.trim();

            if ( check.equalsIgnoreCase( "stdout" ) ) return System.out;
            else if ( check.equalsIgnoreCase( "stderr" ) ) return System.err;
            else 
            {
                throw 
                    state.createFail( 
                        "Illegal log stream (value for property '" + 
                        PROP_NAME_LOG_STREAM + "'):", strm 
                    );
            }
        }
    }

    public
    static
    void
    setDefaultLogSink( PrintStream ps )
    {
        inputs.notNull( ps, "ps" );
        CodeLoggers.replaceDefaultSink( CodeLoggers.createStreamLogger( ps ) );
    }

    public
    static
    void
    setDefaultLogSink()
    {
        setDefaultLogSink( getDefaultLogStream() );
    }
}
