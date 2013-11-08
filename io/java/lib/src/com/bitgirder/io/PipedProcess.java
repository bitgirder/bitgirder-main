package com.bitgirder.io;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.ObjectReceiver;

import java.util.concurrent.Executors;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.ArrayBlockingQueue;

import java.util.concurrent.locks.ReentrantLock;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;

// Provides some basic management of communicating with another process over a
// pipe. Instances are not inherently thread safe, so callers must take care to
// prevent concurrent access.
public
final
class PipedProcess
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Process proc;
    private final ExecutorService es;

    private final ReentrantLock lock = new ReentrantLock();

    private
    PipedProcess( Process proc,
                  ExecutorService es )
    {
        this.proc = proc;
        this.es = es;
    }

    private void shutdownExecutor() { es.shutdownNow(); }

    public
    void
    kill()
    {
        shutdownExecutor();
        proc.destroy();
    }

    private
    final
    class ProcOutUser
    implements Runnable
    {
        private final ObjectReceiver< InputStream > procOut;

        private final ArrayBlockingQueue< Boolean > sem =
            new ArrayBlockingQueue< Boolean >( 1 );

        private Throwable th;

        private
        ProcOutUser( ObjectReceiver< InputStream > procOut )
        {
            this.procOut = procOut;
        }

        public
        void
        run()
        {
            try { procOut.receive( proc.getInputStream() ); }
            catch ( Throwable th ) { this.th = th; }

            state.isTrue( sem.offer( Boolean.TRUE ) );
        }

        void
        join()
            throws Exception
        {
            sem.take();

            if ( th instanceof Error ) throw (Error) th;
            if ( th instanceof Exception ) throw (Exception) th;
        }
    }

    // blocks until both arguments have executed. The threads on which they
    // execute are not specified (but may include the caller's thread), but they
    // will be allowed to execute concurrently with respect to each other.
    public
    void
    usePipe( ObjectReceiver< InputStream > procOut,
             ObjectReceiver< OutputStream > procIn )
        throws Exception
    {
        state.isTrue( lock.tryLock(), "concurrent access not supported" );
        try
        {
            ProcOutUser u = new ProcOutUser( procOut );
            es.submit( u );

            procIn.receive( proc.getOutputStream() );
            u.join();
        }
        finally { lock.unlock(); }
    }

    // This method mutates pb regarding its streams, but does not change
    // anything related to the command or environment
    public
    static
    PipedProcess
    start( ProcessBuilder pb )
        throws IOException
    {
        inputs.notNull( pb, "pb" );

        pb.redirectError( ProcessBuilder.Redirect.INHERIT );

        Process proc = pb.start();
        ExecutorService es = Executors.newFixedThreadPool( 1 );

        return new PipedProcess( proc, es );
    }
}
