package com.bitgirder.jetty7;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.io.IoUtils;

import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleServiceResponse;
import com.bitgirder.mingle.model.MingleStruct;

import com.bitgirder.mingle.service.MingleServices;
import com.bitgirder.mingle.service.MingleServiceCallContext;
import com.bitgirder.mingle.service.MingleServiceEndpoint;

import com.bitgirder.mingle.codec.MingleCodecException;
import com.bitgirder.mingle.codec.MingleCodecs;

import com.bitgirder.mingle.http.MingleHttpCodecContext;
import com.bitgirder.mingle.http.MingleHttpCodecFactory;

import com.bitgirder.http.HttpRequestMessage;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.ProcessRpcClient;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.parser.SyntaxException;

import java.nio.ByteBuffer;

import java.io.IOException;

import java.util.List;
import java.util.Map;
import java.util.Enumeration;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

import javax.servlet.ServletException;

import org.eclipse.jetty.server.Request;
import org.eclipse.jetty.server.Response;
import org.eclipse.jetty.server.Handler;
import org.eclipse.jetty.server.AsyncContinuation;

import org.eclipse.jetty.server.handler.AbstractHandler;

public
final
class Jetty7MingleConnector
extends AbstractVoidProcess 
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public final static String DEFAULT_SERVICE_PATH = "/service";
    public final static Duration DEFAULT_CONTINUATION_TIMEOUT =
        Duration.fromSeconds( 60 );

    private final static String KEY_CODEC_CONTEXT = "codec-context";
    private final static String KEY_MINGLE_RESPONSE = "mingle-response";
    private final static String KEY_ENDPOINT_CALL = "endpoint-call";

    private static final String HDR_ID_STYLE = "x-service-id-style";

    public final static MingleServices.ControlName CONTROL_SERVLET_REQUEST =
        new MingleServices.ControlName( "servlet-request" );

    private final MingleServiceEndpoint endpoint;
    private final MingleHttpCodecFactory codecFact;
    private final String svcPath;
    private final Duration contTimeout;

    private 
    Jetty7MingleConnector( Builder b )
    {
        super( ProcessRpcClient.create() );

        this.endpoint = inputs.notNull( b.endpoint, "endpoint" );
        this.codecFact = inputs.notNull( b.codecFact, "codecFact" );
        this.svcPath = b.svcPath;
        this.contTimeout = b.contTimeout;
    }

    public String getServicePath() { return svcPath; }

    private
    MingleServiceResponse
    getInternalServiceExceptionResponse()
    {
        return
            MingleServiceResponse.createFailure(
                MingleServices.getInternalServiceException() );
    }

    private
    Map< String, List< String > >
    getHeadersMap( AsyncContinuation cont )
    {
        Map< String, List< String > > m = Lang.newMap();

        for ( Enumeration< ? > en = cont.getBaseRequest().getHeaderNames();
                en.hasMoreElements(); )
        {
            String hdr = (String) en.nextElement();

            for ( Enumeration< ? > en2 =
                    cont.getBaseRequest().getHeaders( hdr ); 
                  en2.hasMoreElements(); )
            {
                Lang.putAppend( m, hdr, (String) en2.nextElement() );
            }
        }

        return m;
    }

    private
    CharSequence
    dumpCont( AsyncContinuation cont,
              boolean withHdrs )
    {
        StringBuilder res =
            new StringBuilder().
                append( "cont: " ).append( cont ).
                append( ", status: " ).append( cont.getStatusString() ).
                append( ", " ).append( KEY_MINGLE_RESPONSE ).append( ": " ).
                append( cont.getAttribute( KEY_MINGLE_RESPONSE ) ).
                append( ", isExpired: " ).append( cont.isExpired() );
        
        if ( withHdrs )
        {
            res.append( ", headers: " ).append( getHeadersMap( cont ) );
        }

        return res;
    }

    private
    final
    class EndpointCall
    extends AbstractTask
    {
        private final MingleServiceCallContext cc;
        private final AsyncContinuation cont;

        // It seems that Jetty reuses its AsyncContinuation instances across
        // subsequent requests, meaning that the state of the reference we hold
        // can't necessarily be used as an indicator of the state of this
        // particular call. This becomes especially important in the case in
        // which the continuation times out and the caller gets a response to
        // that effect (see respondContinuationTimeout) but this instances rpc
        // handler may well receive another response after that timeout. At that
        // time the continuation we hold here may be serving another request or
        // in the IDLE,initial state, but regardless, it is no longer relevant
        // to this instance. Also we make this value volatile since it is read
        // from the process thread but written to by one of the jetty worker
        // threads.
        private volatile boolean callDone;

        private
        EndpointCall( MingleServiceCallContext cc,
                      AsyncContinuation cont )
        {
            this.cc = cc;
            this.cont = cont;
        }

        private
        void
        sendResponse( MingleServiceResponse mgResp )
        {
            cont.setAttribute( KEY_MINGLE_RESPONSE, mgResp );
            if ( ! callDone ) cont.resume();
        }

        private
        final
        class RpcHandler
        extends ProcessRpcClient.AbstractResponseHandler
        {
            // for now we don't warn about remote exits because they typically
            // arrive here either because the owning application is shutting
            // down abruptly or because the remote process failed abnormally. In
            // the former case there's no need to litter the logs with
            // stacktraces and in the latter case there should already be a more
            // meaningful stacktrace or failure message somewhere else from the
            // actual remote process
            private 
            boolean 
            shouldWarn( Throwable th ) 
            { 
                return ! ( th instanceof ProcessRpcClient.RemoteExitException );
            }

            @Override
            protected
            void
            rpcFailed( Throwable th )
            {
                if ( shouldWarn( th ) )
                {
                    warn( 
                        th, 
                        "endpoint call failed; returning internal service " +
                        "exception to caller" );
                }

                sendResponse( getInternalServiceExceptionResponse() );
            }
            
            @Override
            protected
            void
            rpcSucceeded( Object resp )
            {
                sendResponse( (MingleServiceResponse) resp );
            }
        }

        protected 
        void 
        runImpl() 
        { 
            behavior( ProcessRpcClient.class ).
                beginRpc( endpoint, cc, new RpcHandler() );
        }
    }

    private
    final
    static
    class BadRequestException
    extends Exception
    {
        private BadRequestException( String msg ) { super( msg ); }
    }

    private
    final
    class HandlerImpl
    extends AbstractHandler
    {
        private
        void
        addHeaders( Request req,
                    HttpRequestMessage.Builder b )
        {
            for ( Enumeration en = req.getHeaderNames(); en.hasMoreElements(); )
            {   
                String nm = (String) en.nextElement();

                if ( nm != null )
                {
                    for ( Enumeration en2 = req.getHeaders( nm );
                            en2.hasMoreElements(); )
                    {
                        String val = (String) en2.nextElement();

                        if ( val != null ) b.h().addField( nm, val );
                    }
                }
            }
        }

        private
        HttpRequestMessage
        asHttpRequestMessage( Request req )
        {
            StringBuilder uri = new StringBuilder( req.getRequestURI() );
            String qs = req.getQueryString();
            if ( qs != null ) uri.append( '?' ).append( qs );

            HttpRequestMessage.Builder b = 
                new HttpRequestMessage.Builder().
                    setMethod( req.getMethod() ).
                    setRequestUri( uri );

            addHeaders( req, b );

            return b.build();
        }

        private
        MingleHttpCodecContext
        getCodecContext( AsyncContinuation cont )
            throws BadRequestException
        {
            HttpRequestMessage msg = 
                asHttpRequestMessage( cont.getBaseRequest() );

            try { return codecFact.codecContextFor( msg ); }
            catch ( MingleCodecException mce )
            {
                throw new BadRequestException( mce.getMessage() );
            }
        }

        private
        Exception
        getRequestDecodeException( Exception ex )
        {
            if ( ex instanceof SyntaxException )
            {
                return 
                    new BadRequestException( 
                        "Request syntax exception: " + ex.getMessage() );
            }
            else if ( ex instanceof MingleCodecException )
            {
                return
                    new BadRequestException(
                        "Request object invalid: " + ex.getMessage() );
            }
            else return ex;
        }
 
        private
        MingleServiceRequest
        getServiceRequest( AsyncContinuation cont )
            throws Exception
        {
            ByteBuffer bin = 
                IoUtils.toByteBuffer( 
                    cont.getRequest().getInputStream(), true );

            MingleHttpCodecContext codecCtx = getCodecContext( cont );
            cont.setAttribute( KEY_CODEC_CONTEXT, codecCtx );

            try 
            { 
                return 
                    MingleServices.asServiceRequest(
                        MingleCodecs.fromByteBuffer(
                            codecCtx.codec(), bin, MingleStruct.class ) );
            }
            catch ( Exception ex ) { throw getRequestDecodeException( ex ); }
        }

        private
        MingleServiceCallContext
        getCallContext( AsyncContinuation cont )
            throws Exception
        {
            MingleServiceRequest mgReq = getServiceRequest( cont );

            MingleServiceCallContext res = 
                MingleServiceCallContext.create( mgReq );
            
            res.attachments().put( CONTROL_SERVLET_REQUEST, cont.getRequest() );

            return res;
        }
 
        private
        void
        beginRequest( final AsyncContinuation cont )
            throws Exception
        {
            Request req = cont.getBaseRequest();

            String method = req.getMethod();

            if ( method.equals( "POST" ) )
            {
                MingleServiceCallContext cc = getCallContext( cont );
                EndpointCall call = new EndpointCall( cc, cont );
                cont.setAttribute( KEY_ENDPOINT_CALL, call );
                cont.setTimeout( contTimeout.asMillis() );
                cont.suspend();

                submit( call );
            }
            else throw new ServletException( "Unsupported method: " + method );
        }
 
        private
        void
        writeServiceResponse( AsyncContinuation cont,
                              MingleServiceResponse mgResp )
            throws Exception
        {
            MingleHttpCodecContext codecCtx =
                state.cast(
                    MingleHttpCodecContext.class,
                    cont.getAttribute( KEY_CODEC_CONTEXT ) );
    
            ByteBuffer respBytes = 
                MingleCodecs.toByteBuffer(
                    codecCtx.codec(), MingleServices.asMingleStruct( mgResp ) );
    
            Response resp = cont.getBaseRequest().getResponse();
            resp.setStatus( HttpServletResponse.SC_OK );
            resp.setContentType( codecCtx.contentType().toString() );
            resp.setContentLength( respBytes.remaining() );
            resp.getOutputStream().write( IoUtils.toByteArray( respBytes ) );
        }

        private
        void
        respondContinuationTimeout( AsyncContinuation cont )
            throws Exception
        {
            writeServiceResponse( cont, getInternalServiceExceptionResponse() );
        }

        private
        void
        beginResponse( AsyncContinuation cont )
            throws Exception
        {
            EndpointCall call =
                (EndpointCall) cont.getAttribute( KEY_ENDPOINT_CALL );

            call.callDone = true;

            if ( cont.isExpired() ) respondContinuationTimeout( cont );
            else
            {
                MingleServiceResponse mgResp =
                    (MingleServiceResponse) 
                        cont.getAttribute( KEY_MINGLE_RESPONSE );
    
                state.isFalse( 
                    mgResp == null, "Request resumed with no mingle response" );
    
                writeServiceResponse( cont, mgResp );
            }
        }
    
        private
        void
        handleServiceRequestImpl( AsyncContinuation cont )
            throws Exception
        {
            if ( cont.isInitial() ) beginRequest( cont );
            else beginResponse( cont );
        }
    
        private
        void
        respondBadRequest( Request req,
                           String msg )
            throws ServletException,
                   IOException
        {
            Response resp = req.getResponse();
    
            resp.setStatus( HttpServletResponse.SC_BAD_REQUEST );
            resp.setContentType( "text/plain;charset=utf-8" );
    
            resp.getOutputStream().write( ( msg + "\n" ).getBytes( "utf-8" ) );
    
            req.setHandled( true );
        }
    
        private
        void
        respondBadRequest( Request req,
                           BadRequestException ex )
            throws ServletException,
                   IOException
        {
            respondBadRequest( req, ex.getMessage() );
        }
    
        // Some implementations of ServletException appear not to respect the
        // Throwable passed to the constructor but do respect those set via
        // initCause(), so we set it using the latter approach. 
        //
        // Also we make sure to add the message of the cause th into the message
        // of the exception we're rethrowing as a baseline safety net against
        // our never seeing the cause trace at all because there is no logging
        // of stack traces in place for whatever reason, in which case we can at
        // least know its message
        private
        void
        rethrowInternalServiceError( Throwable th )
            throws ServletException
        {
            ServletException se = 
                new ServletException( 
                    "Internal service error: " + th.getMessage() );
    
            se.initCause( th );
            warn( se );
    
            throw se;
        }
    
        private
        void
        handleServiceRequest( Request req )
            throws IOException,
                   ServletException
        {
            try
            {
                handleServiceRequestImpl( req.getAsyncContinuation() );
                req.setHandled( true );
            }
            catch ( ServletException se ) { throw se; }
            catch ( BadRequestException bre ) { respondBadRequest( req, bre ); }
            catch ( Throwable th ) { rethrowInternalServiceError( th ); }
        }
    
        @Override
        public
        void
        handle( String target,
                Request request,
                HttpServletRequest servletReq,
                HttpServletResponse servletResp )
            throws IOException,
                   ServletException
        {
            if ( target.equals( svcPath ) ) handleServiceRequest( request );
            else respondBadRequest( request, "Unrecognized path: " + target );
        }
    }

    public Handler getHandler() { return new HandlerImpl(); }

    public 
    void 
    stop() 
    { 
        submit( new AbstractTask() { protected void runImpl() { exit(); } } );
    }

    protected void startImpl() {}

    public
    final
    static
    class Builder
    {
        private MingleServiceEndpoint endpoint;
        private String svcPath = DEFAULT_SERVICE_PATH;
        private Duration contTimeout = DEFAULT_CONTINUATION_TIMEOUT;
        private MingleHttpCodecFactory codecFact;

        public
        Builder
        setMingleEndpoint( MingleServiceEndpoint endpoint )
        {
            this.endpoint = inputs.notNull( endpoint, "endpoint" );
            return this;
        }

        public
        Builder
        setServicePath( String svcPath )
        {
            this.svcPath = inputs.notNull( svcPath, "svcPath" );
            return this;
        }

        public
        Builder
        setContinuationTimeout( Duration contTimeout )
        {
            this.contTimeout = inputs.notNull( contTimeout, "contTimeout" );
            return this;
        }

        public
        Builder
        setCodecFactory( MingleHttpCodecFactory codecFact )
        {
            this.codecFact = inputs.notNull( codecFact, "codecFact" );
            return this;
        }

        public
        Jetty7MingleConnector
        build()
        {
            return new Jetty7MingleConnector( this );
        }
    }
}
