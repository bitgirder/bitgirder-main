package com.bitgirder.mingle.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.process.ProcessBehavior;

import java.io.InputStream;
import java.io.BufferedInputStream;
import java.io.OutputStream;
import java.io.PrintStream;
import java.io.UnsupportedEncodingException;

import java.util.List;

public
final
class MinglePipeManager
extends ProcessBehavior
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final List< MingleMessagePipe > pipes = Lang.newList();
    
    private MinglePipeManager() {}

    protected void startImpl() {}

    @Override
    protected
    void
    beginShutdown()
    {
        for ( MingleMessagePipe pipe : pipes ) pipe.stop();
        shutdownComplete();
    }

    private
    PrintStream
    createPrintStream( OutputStream out )
    {
        try { return new PrintStream( out, true, "utf-8" ); }
        catch ( UnsupportedEncodingException uee )
        {
            throw new RuntimeException( uee ); 
        }
    }

    public
    MingleMessagePipe
    managePipe( InputStream in,
                OutputStream out )
    {
        inputs.notNull( in, "in" );
        inputs.notNull( out, "out" );

        PrintStream ps = createPrintStream( out );

        MingleMessagePipe res = 
            new MingleMessagePipe( in, ps, getActivityContext() );
        
        pipes.add( res );
        res.start();

        return res;
    }

    public
    MingleMessagePipe
    manageStandardIo()
    {
        return managePipe( new BufferedInputStream( System.in ), System.out );
    }

    public static MinglePipeManager create() { return new MinglePipeManager(); }
}
