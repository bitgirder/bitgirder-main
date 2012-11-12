package com.bitgirder.jetty7;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.Stoppable;

import com.bitgirder.mingle.service.MingleServiceEndpoint;

import org.eclipse.jetty.server.Server;
import org.eclipse.jetty.server.Connector;

import org.eclipse.jetty.server.nio.SelectChannelConnector;

import org.eclipse.jetty.util.component.LifeCycle;

public
final
class Jetty7HttpMingleServer
extends AbstractVoidProcess
implements Stoppable
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Jetty7MingleConnector mgConn;
    private final int httpPort;
    private final EventHandler eh;

    private Jetty7LifeCycleManager< Server > srvMgr;

    private boolean begunStop;

    private
    Jetty7HttpMingleServer( Builder b )
    {
        this.mgConn = inputs.notNull( b.mgConn, "mgConn" );
        this.httpPort = b.httpPort;
        this.eh = b.eh;
    }

    @Override
    protected
    void
    childExited( AbstractProcess< ? > child,
                 ProcessExit< ? > exit )
    {
        warnIfFailed( child, exit );

        if ( ! hasChildren() ) exit();
    }

    private
    final
    class ConnectorListener
    extends AbstractLifeCycleListener
    {
        @Override
        public
        void
        lifeCycleStarted( LifeCycle event )
        {
            SelectChannelConnector conn = (SelectChannelConnector) event;

            if ( eh != null ) 
            {
                try { eh.httpStarted( conn.getLocalPort() ); }
                catch ( Throwable th )
                {
                    warn( th, "Event handler failed in httpStarted()" );
                }
            }
        }
    }

    protected
    void
    startImpl()
        throws Exception
    {
        Server srv = new Server();
        srv.setStopAtShutdown( false );

        SelectChannelConnector conn = new SelectChannelConnector();
        conn.addLifeCycleListener( new ConnectorListener() );
        conn.setPort( httpPort );
        srv.setConnectors( new Connector[] { conn } );
 
        spawn( mgConn );
        srv.setHandler( mgConn.getHandler() ); 
 
        srvMgr = new Jetty7LifeCycleManager< Server >( srv );
        spawn( srvMgr );
    }

    private
    void
    doStop()
    {
        if ( ! begunStop )
        {
            begunStop = true;
            
            code( "Stopping jetty server" );
            srvMgr.stop();

            code( "Stopping mingle connector" );
            mgConn.stop();
        }
    }

    public
    void
    stop()
    {
        submit( new AbstractTask() { protected void runImpl() { doStop(); } } );
    }

    public
    static
    interface EventHandler
    {
        public
        void
        httpStarted( int port );
    }

    public
    final
    static
    class Builder
    {
        private Jetty7MingleConnector mgConn;
        private int httpPort = 0;
        private EventHandler eh;

        public
        Builder
        setMingleConnector( Jetty7MingleConnector mgConn )
        {
            this.mgConn = inputs.notNull( mgConn, "mgConn" );
            return this;
        }

        public
        Builder
        setMingleEndpoint( MingleServiceEndpoint mgEndpoint )
        {
            inputs.notNull( mgEndpoint, "mgEndpoint" );

            setMingleConnector(
                new Jetty7MingleConnector.Builder().
                    setMingleEndpoint( mgEndpoint ).
                    build() );
            
            return this;
        }

        public
        Builder
        setHttpPort( int httpPort )
        {
            this.httpPort = inputs.nonnegativeI( httpPort, "httpPort" );
            return this;
        }

        public
        Builder
        setEventHandler( EventHandler eh )
        {
            this.eh = inputs.notNull( eh, "eh" );
            return this;
        }

        public
        Jetty7HttpMingleServer
        build()
        {
            return new Jetty7HttpMingleServer( this );
        }
    }
}
