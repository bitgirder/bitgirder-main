package com.bitgirder.mingle.http;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.ProcessActivity;
import com.bitgirder.process.ProcessRpcClient;

import com.bitgirder.io.ProtocolProcessor;
import com.bitgirder.io.ProtocolProcessors;
import com.bitgirder.io.DataSize;
import com.bitgirder.io.InsulateProcessor;
import com.bitgirder.io.Charsets;

import com.bitgirder.parser.SyntaxException;

import com.bitgirder.http.AbstractHttpResponder;
import com.bitgirder.http.HttpResponder;
import com.bitgirder.http.HttpResponderFactory;
import com.bitgirder.http.HttpRequestMessage;
import com.bitgirder.http.HttpResponseMessage;
import com.bitgirder.http.HttpConstants;
import com.bitgirder.http.HttpStatusCode;
import com.bitgirder.http.HttpServiceConnection;
import com.bitgirder.http.HttpProtocolProcessors;

import com.bitgirder.mingle.codec.MingleCodec;
import com.bitgirder.mingle.codec.MingleCodecs;
import com.bitgirder.mingle.codec.MingleDecoder;
import com.bitgirder.mingle.codec.MingleCodecException;

import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleServiceResponse;
import com.bitgirder.mingle.model.MingleStruct;

import com.bitgirder.mingle.service.MingleServiceCallContext;
import com.bitgirder.mingle.service.MingleServices;

import java.nio.ByteBuffer;

public
final
class MingleHttpResponder
implements HttpResponderFactory
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final AbstractProcess< ? > endpoint;
    private final MingleHttpCodecFactory codecFact;

    // hardcode for now
    private final DataSize chunkSize = DataSize.ofKilobytes( 4 ); 

    private
    MingleHttpResponder( Builder b )
    {
        this.endpoint = inputs.notNull( b.endpoint, "endpoint" );
        this.codecFact = inputs.notNull( b.codecFact, "codecFact" );
    }

    private
    final
    class RpcResponder
    extends AbstractHttpResponder
    {
        private MingleHttpCodecContext cdCtx;
        private MingleDecoder< MingleStruct > dec;
        private InsulateProcessor ins;

        private 
        RpcResponder( HttpResponder.Context ctx ) 
        { 
            super( ctx ); 

            // Just access to check it's there
            behavior( ProcessRpcClient.class ); 
        }

        private
        MingleHttpCodecContext
        getCodecContext()
            throws Exception
        {
            try { return codecFact.codecContextFor( request() ); }
            catch ( MingleCodecException mce )
            {
                respondBadRequest( mce.getMessage() + "\n" );
                return null;
            }
            catch ( Exception ex ) { throw ex; }
        }

        @Override
        protected
        ProtocolProcessor< ByteBuffer >
        getRequestBodyReceiver()
            throws Exception
        {
            cdCtx = getCodecContext();

            if ( cdCtx == null ) 
            {
                return 
                    asBodyReceiver( ProtocolProcessors.getDiscardProcessor() );
            }
            else
            {
                dec = cdCtx.codec().createDecoder( MingleStruct.class );
 
                ins =
                    InsulateProcessor.create(
                        MingleCodecs.createReceiveProcessor( dec ) );
                
                return asBodyReceiver( ins );
            }
        }

        private
        void
        respondBadRequest( String msg )
        {
            ByteBuffer body = Charsets.UTF_8.asByteBufferUnchecked( msg );

            respond(
                sendBuilder().
                setMessage(
                    responseBuilder().
                    setStatus( HttpStatusCode.BAD_REQUEST ).
                    h().setContentLength( body.remaining() ).
                    h().setContentType( "text/plain; charset=utf-8" ).
                    build()
                ).
                setBody( body ).
                build()
            );
        }

        private
        void
        respondBadRequest( Throwable th )
            throws Exception
        {
            respondBadRequest( th.getMessage() );
        }

        private
        MingleServiceRequest
        getRequest()
            throws Exception
        {
            try { return MingleServices.asServiceRequest( dec.getResult() ); }
            catch ( MingleCodecException ex )
            {
                respondBadRequest( 
                    "Request object invalid: " + ex.getMessage() + "\n" );

                return null;
            }
        }
 
        private
        void
        respondInternalHttpFailure( Throwable th )
        {
            warn( 
                th, 
                "Failing with external HTTP failure (see actual attached)" );
 
            respond( HttpStatusCode.INTERNAL_SERVER_ERROR );
        }

        private
        ProtocolProcessor< ByteBuffer >
        getCodecSend( MingleServiceResponse resp )
        {
            try 
            {
                MingleStruct ms = MingleServices.asMingleStruct( resp );
                return MingleCodecs.createSendProcessor( cdCtx.codec(), ms );
            }
            catch ( Throwable th ) 
            { 
                respondInternalHttpFailure( th ); 
                return null;
            }
        }

        private
        ProtocolProcessor< ByteBuffer >
        getResponseBody( MingleServiceResponse resp,
                         HttpResponseMessage.Builder b )
        {
            ProtocolProcessor< ByteBuffer > proc = getCodecSend( resp );

            if ( proc == null ) return null;
            else
            {
                boolean gzip = request().h().hasAcceptableEncoding( "gzip" );

                b.h().setTransferEncoding( "chunked" );

                if ( gzip ) 
                {
                    proc = asGzippedSend( proc );
                    b.h().setContentEncoding( "gzip" );
                }

                return createChunkedSend( proc, chunkSize );
            }
        }

        private
        void
        sendResponse( MingleServiceResponse resp )
        {
            HttpResponseMessage.Builder b = 
                responseBuilder().
                setStatus( HttpStatusCode.OK ).
                h().setContentType( cdCtx.contentType() );

            ProtocolProcessor< ByteBuffer > body = getResponseBody( resp, b );

            if ( body != null )
            {
                respond(
                    sendBuilder().
                    setMessage( b.build() ).
                    setBody( body ).
                    build()
                );
            }
            // else we've already responded with the error
        }

        private
        final
        class EndpointRpcHandler
        extends ProcessRpcClient.AbstractResponseHandler
        {
            public
            void
            rpcSucceeded( Object resp )
            {
                sendResponse( (MingleServiceResponse) resp );
            }

            public 
            void 
            rpcFailed( Throwable th ) 
            { 
//                failInternalMingle( th ); 
                throw new UnsupportedOperationException( "Unimplemented" );
            }
        }

        private
        void
        beginRpc()
            throws Exception
        {
            MingleServiceRequest req = getRequest();

            if ( req != null )
            {
                behavior( ProcessRpcClient.class ).beginRpc(
                    endpoint,
                    MingleServiceCallContext.create( req ),
                    new EndpointRpcHandler()
                );
            }
        }

        private
        void
        respondRequestReadFailure( Throwable th )
            throws Exception
        {
            if ( th instanceof SyntaxException )
            {
                respondBadRequest( 
                    "Request syntax exception: " + th.getMessage() + "\n" );
            }
            else if ( th instanceof MingleCodecException )
            {
                respondBadRequest( th );
            }
            else respondInternalHttpFailure( th );
        }

        protected
        void
        implBeginResponse()
            throws Exception
        {
            if ( ins.getThrowable() == null ) beginRpc();
            else respondRequestReadFailure( ins.getThrowable() );
        }
    }

    public
    HttpResponder
    responderFor( HttpResponder.Context ctx )
    {
        return new RpcResponder( ctx );
    }

    public
    final
    static
    class Builder
    {
        private AbstractProcess< ? > endpoint;
        private MingleHttpCodecFactory codecFact;

        public
        Builder
        setEndpoint( AbstractProcess< ? > endpoint )
        {
            this.endpoint = inputs.notNull( endpoint, "endpoint" );
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
        MingleHttpResponder
        build() 
        { 
            return new MingleHttpResponder( this ); 
        }
    }
}
