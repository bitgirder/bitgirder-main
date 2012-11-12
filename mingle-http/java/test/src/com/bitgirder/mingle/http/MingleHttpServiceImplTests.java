package com.bitgirder.mingle.http;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.ProcessRpcClient;

import com.bitgirder.mingle.model.MingleIdentifierFormat;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleModels;

import com.bitgirder.net.SelectableChannelManager;
import com.bitgirder.net.NetTests;
import com.bitgirder.net.NetProtocolTransportFactory;

import com.bitgirder.net.ssl.NetSslTests;

import com.bitgirder.http.HttpOperations;
import com.bitgirder.http.HttpRequestMessage;
import com.bitgirder.http.HttpResponseMessage;
import com.bitgirder.http.HttpMessages;
import com.bitgirder.http.HttpClient;
import com.bitgirder.http.HttpClientConnection;

import com.bitgirder.io.Charsets;
import com.bitgirder.io.IoUtils;
import com.bitgirder.io.ProtocolProcessor;
import com.bitgirder.io.ByteBufferAccumulator;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestRuntime;
import com.bitgirder.test.LabeledTestObject;
import com.bitgirder.test.InvocationFactory;

import com.bitgirder.test.AbstractLabeledTestObject;

import java.util.List;

import java.nio.ByteBuffer;

public
final
class MingleHttpServiceImplTests
extends AbstractLabeledTestObject
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleHttpTesting.ServerLocation testLoc;
    private final TestRuntime rt;

    private
    MingleHttpServiceImplTests( Builder b )
    {
        super( inputs.notNull( inputs.notNull( b, "b" ).label, "label" ) );

        this.testLoc = inputs.notNull( b.testLoc, "testLoc" );
        this.rt = inputs.notNull( b.rt, "rt" );
    }

    private
    abstract
    class BufferRequestTest
    extends AbstractVoidProcess
    {
        private int exitWait;

        private 
        BufferRequestTest()
        { 
            super( SelectableChannelManager.create() );
        }

        public final Object getInvocationTarget() { return this; }

        private void exitConditional() { if ( --exitWait == 0 ) exit(); }

        private
        HttpClient
        httpCli()
        {
            NetProtocolTransportFactory sslTransFact =
                NetSslTests.createTransportFactory( true, rt );

            return
                new HttpClient.Builder().
                    setNetworking( NetTests.expectSelectorManager( rt ) ).
                    setActivityContext( getActivityContext() ).
                    setSslTransportFactory( sslTransFact ).
                    build();
        }

        abstract
        ByteBuffer
        buildRequest( HttpRequestMessage.Builder b )
            throws Exception;

        final
        CharSequence
        asString( Iterable< ByteBuffer > bufs )
            throws Exception
        {
            return IoUtils.asString( bufs, Charsets.UTF_8.newDecoder() );
        }

        abstract
        void
        checkResponse( HttpResponseMessage msg,
                       Iterable< ByteBuffer > body )
            throws Exception;

        private
        final
        class RequestImpl
        extends HttpOperations.AbstractRequest
        {
            private final ByteBufferAccumulator acc = 
                ByteBufferAccumulator.create( 256 );

            private RequestImpl() { super( getActivityContext() ); }

            protected
            ProtocolProcessor< ByteBuffer >
            createSend()
                throws Exception
            {
                HttpRequestMessage.Builder b = 
                    requestBuilder().
                    setMethod( "POST" ).
                    setRequestUri( testLoc.uri() ).
                    h().setHost( testLoc.host(), testLoc.port() ).
                    h().setConnection( "Close" );

                ByteBuffer body = buildRequest( b );

                b.h().setContentLength( body.remaining() );

                return
                    sendBuilder().
                    setMessage( b.build() ).
                    setBody( body ).
                    build();
            }

            @Override 
            protected 
            ProtocolProcessor< ByteBuffer >
            getResponseBodyReceiver() 
            { 
                return asBodyReceiver( acc ); 
            }

            @Override
            protected
            void
            requestComplete( HttpClientConnection conn )
                throws Exception
            {
                checkResponse( response(), acc.getBuffers() );

                conn.close();
                exitConditional();
            }
        }

        private
        final
        class ConnectHandler
        extends HttpClient.AbstractConnectHandler
        {
            private ConnectHandler() { super( self() ); }

            @Override 
            protected 
            void 
            connectSucceededImpl( HttpClientConnection conn )
            {
                ++exitWait;
                conn.begin( new RequestImpl() );
            }

            @Override protected void transportClosed() { exitConditional(); }
        }

        protected
        final
        void
        startImpl()
            throws Exception
        {
            httpCli().connect(
                testLoc.host(),
                testLoc.port(),
                testLoc.isSsl(),
                new ConnectHandler()
            );
        }
    }

    @Test
    private
    final
    class BadRequestOnMissingCtypeTest
    extends BufferRequestTest
    {
        ByteBuffer
        buildRequest( HttpRequestMessage.Builder b )
        {
            return ByteBuffer.allocate( 1 );
        }

        void
        checkResponse( HttpResponseMessage msg,
                       Iterable< ByteBuffer > body )
            throws Exception
        {
            state.equalString( "Missing content type\n", asString( body ) );
        }
    }

    private
    final
    class IdStyleTest
    extends BufferRequestTest
    implements LabeledTestObject
    {
        private final MingleIdentifierFormat fmt;

        private IdStyleTest( MingleIdentifierFormat fmt ) { this.fmt = fmt; }
        
        public 
        final 
        CharSequence 
        getLabel() 
        { 
            return fmt == null ? "default" : fmt.name();
        }

        private
        ByteBuffer
        createJsonReq()
            throws Exception
        {
            return 
                Charsets.UTF_8.asByteBuffer(
                    "{" +
                        "\"$type\":\"service@v1/ServiceRequest\", " +
                        "\"namespace\":\"mingle:tck@v1\", " +
                        "\"service\":\"service1\", " +
                        "\"operation\":\"echoOpaqueValue\", " +
                        "\"parameters\":{\"value\":{\"check-this\":1}}" +
                    "}"
                );
        }

        ByteBuffer
        buildRequest( HttpRequestMessage.Builder b )
            throws Exception
        {
            b.h().setContentType( "application/json" );

            if ( fmt != null )
            {
                b.h().setField(
                    MingleHttpConstants.HEADER_ID_STYLE,
                    MingleHttpProcessing.asIdStyleHeaderValue( fmt )
                );
            }

            return createJsonReq();
        }

        void
        checkResponse( HttpResponseMessage msg,
                       Iterable< ByteBuffer > body )
            throws Exception
        {
            CharSequence json = asString( body );

            CharSequence expct = 
                MingleModels.format(
                    MingleIdentifier.create( "check-this" ),
                    fmt == null ? MingleHttpConstants.DEFAULT_ID_STYLE : fmt
                );

            state.isTrue( json.toString().indexOf( expct.toString() ) >= 0 );
            exit();
        }
    }

    @InvocationFactory
    private
    List< IdStyleTest >
    testIdStyle()
    {
        List< IdStyleTest > res = Lang.newList();
        res.add( new IdStyleTest( null ) );

        for ( MingleIdentifierFormat fmt : 
                MingleIdentifierFormat.class.getEnumConstants() )
        {
            res.add( new IdStyleTest( fmt ) );
        }

        return res;
    }

    private
    final
    class JsonParseFailureTest
    extends BufferRequestTest
    implements LabeledTestObject
    {
        private final CharSequence bodyText;
        private final CharSequence errExpct;
        private final CharSequence lbl;

        private
        JsonParseFailureTest( CharSequence bodyText,
                              CharSequence errExpct,
                              CharSequence lbl )
        {
            this.bodyText = bodyText;
            this.errExpct = errExpct;
            this.lbl = lbl;
        }

        public CharSequence getLabel() { return lbl; }

        ByteBuffer
        buildRequest( HttpRequestMessage.Builder b )
            throws Exception
        {
            b.h().setContentType( "application/json" );

            return Charsets.UTF_8.asByteBuffer( bodyText );
        }

        void
        checkResponse( HttpResponseMessage msg,
                       Iterable< ByteBuffer > body )
            throws Exception
        {
            state.equalInt( 400, msg.statusCodeValue() );

            state.equalString( errExpct, asString( body ) );
        }
    }

    @InvocationFactory
    private
    List< JsonParseFailureTest >
    testJsonParseFailure()
    {
        return Lang.asList(
            
            new JsonParseFailureTest(
                "{ \"bad: 12 }", 
                "Request syntax exception: [ <>; line 1, col 3 ]\n",
                "unterminated-json-key"
            ),
            
            // some of these are redundant with the similar tests in JsonTests
            // or in tests of the JsonMingleCodec, but we repeat them here to
            // ensure coverage and exception transmission at the Jetty layer
            new JsonParseFailureTest(
                "{", 
                "Request syntax exception: Insufficient or malformed data in " +
                    "buffer\n",
                "open-brace-only"
            ),
            
            // 2 chars is enough for JSON charset detection but is still an
            // illegal docuemnt
            new JsonParseFailureTest(
                "{ ", 
                "Request syntax exception: Unmatched document\n",
                "open-brace-and-space"
            ),
            
            // A valid JSON text but not a valid MingleServiceRequest
            new JsonParseFailureTest(
                "{}", 
                "Request object invalid: $type: Missing value\n",
                "empty-json-doc"
            ),

            // Totally empty doc
            new JsonParseFailureTest(
                "",
                "Request syntax exception: Insufficient or malformed data in " +
                    "buffer\n",
                "zero-length-req-body"
            ),

            new JsonParseFailureTest(
                "  \r\n\t   ",
                "Request syntax exception: Unmatched document\n",
                "whitespace-only-body"
            )
        );
    }

    public
    final
    static
    class Builder
    {
        private MingleHttpTesting.ServerLocation testLoc;
        private CharSequence label;
        private TestRuntime rt;

        public
        Builder
        setLocation( MingleHttpTesting.ServerLocation testLoc )
        {
            this.testLoc = inputs.notNull( testLoc, "testLoc" );
            return this;
        }

        public
        Builder
        setLabel( CharSequence label )
        {
            this.label = inputs.notNull( label, "label" );
            return this;
        }

        public
        Builder
        setRuntime( TestRuntime rt )
        {
            this.rt = inputs.notNull( rt, "rt" );
            return this;
        }

        public 
        MingleHttpServiceImplTests 
        build() 
        { 
            return new MingleHttpServiceImplTests( this );
        }
    }
}
