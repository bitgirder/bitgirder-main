package com.bitgirder.mingle.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleModels;

import com.bitgirder.io.FileWrapper;
import com.bitgirder.io.ProtocolProcessor;
import com.bitgirder.io.AbstractProtocolProcessor;
import com.bitgirder.io.GunzipProcessor;
import com.bitgirder.io.IoProcessor;
import com.bitgirder.io.FileFeed;
import com.bitgirder.io.DataSize;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.AbstractProcess;

import com.bitgirder.application.ApplicationProcess;

import java.util.List;

import java.nio.ByteBuffer;

final
class JsonMingleStructLineLint
extends ApplicationProcess
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final FileWrapper inputFile;
    private final DataSize ioBufSize;
    private final boolean logStructs;

    private IoProcessor ioProc;

    private long started;

    private
    JsonMingleStructLineLint( Configurator c )
    {
        super( c );

        this.inputFile = inputs.notNull( c.inputFile, "inputFile" );
        this.ioBufSize = c.ioBufSize;
        this.logStructs = c.logStructs;
    }

    @Override
    protected
    void
    childExited( AbstractProcess< ? > proc,
                 ProcessExit< ? > exit )
    {
        if ( ! exit.isOk() ) fail( exit.getThrowable() );
        if ( ! hasChildren() ) exit();
    }

    private
    final
    class InputProcessor
    extends AbstractProtocolProcessor< List< MingleStruct > >
    {
        protected
        void
        processImpl( ProcessContext< List< MingleStruct > > ctx )
        {
            if ( logStructs ) 
            {
                for ( MingleStruct ms : ctx.object() )
                {
                    code( MingleModels.inspect( ms ) );
                }
            }

            doneOrData( ctx );
        }
    }

    private
    final
    class Reactor
    extends JsonMingleStructLineProcessor.AbstractReactor
    {
        @Override
        public
        void
        parseFailed( Throwable th,
                     long recNo )
            throws Exception
        {
            warn( "Failing at recNo:", recNo );

            if ( th instanceof Exception ) throw (Exception) th;
            else throw (Error) th;
        }
    }

    private boolean isGzip() { return inputFile.toString().endsWith( ".gz" ); }

    private
    ProtocolProcessor< ByteBuffer >
    getFileProcessor()
    {
        ProtocolProcessor< ByteBuffer > res =
            new JsonMingleStructLineProcessor.Builder().
                setInputProcessor( new InputProcessor() ).
                setReactor( new Reactor() ).
                build();
        
        if ( isGzip() ) 
        {
            res = GunzipProcessor.create( res, getActivityContext() );
        }

        return res;
    }

    private
    final
    class EventHandler
    extends FileFeed.AbstractEventHandler
    {
        private EventHandler() { super( self() ); }

        protected
        void
        fileCopyCompleteImpl()
        {
            code( "Completed feed of", inputFile );

            Duration elapsed = 
                Duration.fromMillis( System.currentTimeMillis() - started );

            code( "Ran in", elapsed.toStringSeconds( 3 ) + "s" );

            ioProc.stop();
        }
    }

    private
    FileFeed
    createFileFeed( ProtocolProcessor< ByteBuffer > proc )
    {
        spawn( ioProc = IoProcessor.create( 1 ) ); 

        return
            new FileFeed.Builder().
                setActivityContext( getActivityContext() ).
                setIoProcessor( ioProc ).
                setFile( inputFile ).
                setIoBufferSize( ioBufSize ).
                setProcessor( proc ).
                setEventHandler( new EventHandler() ).
                build();
    }

    protected
    void
    startImpl()
        throws Exception
    {
        inputFile.assertExists();

        ProtocolProcessor< ByteBuffer > proc = getFileProcessor();
        FileFeed ff = createFileFeed( proc );

        code( "Starting lint" );
        
        started = System.currentTimeMillis();
        ff.start();
    }

    private
    final
    static
    class Configurator
    extends ApplicationProcess.Configurator
    {
        private FileWrapper inputFile;
        private DataSize ioBufSize = DataSize.ofKilobytes( 20 );
        private boolean logStructs;

        @Argument
        private
        void
        setInputFile( String f )
        {
            inputFile = new FileWrapper( f ).assertExists();
        }

        @Argument
        private
        void
        setIoBufSize( String sz )
        {
            ioBufSize = DataSize.fromString( sz );
        }

        @Argument private void setLogStructs( boolean b ) { logStructs = b; }
    }
}
