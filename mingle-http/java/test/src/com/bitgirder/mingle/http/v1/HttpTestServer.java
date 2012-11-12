package com.bitgirder.mingle.http.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.ObjectReceiver;
import com.bitgirder.lang.Lang;

import com.bitgirder.testing.TestRuntimeContext;
import com.bitgirder.testing.TestRuntimeContexts;

import com.bitgirder.process.ProcessRpcClient;

import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleIdentifiedName;
import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleModels;

import com.bitgirder.mingle.bind.MingleBinders;

import com.bitgirder.mingle.codec.MingleCodec;
import com.bitgirder.mingle.codec.MingleCodecs;
import com.bitgirder.mingle.codec.MingleCodecFactory;
import com.bitgirder.mingle.codec.MingleCodecFactories;

import com.bitgirder.mingle.io.MinglePipeManager;
import com.bitgirder.mingle.io.MingleMessagePipe;

import com.bitgirder.mingle.http.MingleHttpTesting;

import com.bitgirder.mingle.application.MingleApplicationProcess;

import java.util.Map;
import java.util.List;

import java.nio.ByteBuffer;

final
class HttpTestServer
extends MingleApplicationProcess< HttpTestServerConfig >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static MingleIdentifier ID_COMMAND = 
        MingleIdentifier.create( "command" );

    private final static MingleIdentifier ID_CODEC =
        MingleIdentifier.create( "codec" );

    private final static MingleIdentifier CMD_GET_HTTP_SERVERS =
        MingleIdentifier.create( "get-http-servers" );

    private final static MingleIdentifier CMD_CLOSE =
        MingleIdentifier.create( "close" );

    private MingleCodecFactory codecFact;

    private TestRuntimeContext rtCtx;

    private
    HttpTestServer()
    {
        super(
            new Builder< HttpTestServerConfig >().
                setConfigType( 
                    "bitgirder:mingle:http@v1/HttpTestServerConfig" ).
                setConfigClass( HttpTestServerConfig.class ).
                mixin( ProcessRpcClient.create() ).
                mixin( MinglePipeManager.create() )
        );
    }

    private
    MingleIdentifier
    expectId( MingleMessagePipe.Message msg,
              MingleIdentifier key )
    {
        return
            MingleIdentifier.
                create( msg.modifiers().expectMingleString( key ) );
    }

    private
    MingleCodec
    expectCodec( MingleMessagePipe.Message msg )
        throws Exception
    {
        return codecFact.expectCodec( expectId( msg, ID_CODEC ) );
    }

    private
    ByteBuffer
    encode( Object obj,
            MingleMessagePipe.Message msg )
        throws Exception
    {
        MingleCodec codec = expectCodec( msg );

        MingleStruct ms = 
            (MingleStruct) MingleBinders.asMingleValue( binder(), obj );

        return MingleCodecs.toByteBuffer( codec, ms );
    }

    private
    HttpTestServerInfo
    asServerInfo( MingleIdentifiedName name,
                  MingleHttpTesting.ServerLocation loc )
    {
        return
            new HttpTestServerInfo.Builder().
                setName( name.getExternalForm().toString() ).
                setHost( loc.host() ).
                setPort( loc.port() ).
                setUri( loc.uri() ).
                setIsSsl( loc.isSsl() ).
                build();
    }

    private
    List< HttpTestServerInfo >
    getServerInfos()
    {
        Map< MingleIdentifiedName, MingleHttpTesting.ServerLocation > locs =
            MingleHttpTesting.expectServerLocations( rtCtx.runtime() );

        List< HttpTestServerInfo > res = Lang.newList();
        
        for ( Map.Entry< MingleIdentifiedName, 
                         MingleHttpTesting.ServerLocation > e : 
                locs.entrySet() )
        {
            res.add( asServerInfo( e.getKey(), e.getValue() ) );
        }

        return res;
    }

    private
    void
    sendHttpServers( MingleMessagePipe.Message msg,
                     final MingleMessagePipe pipe )
        throws Exception
    {
        HttpTestServersInfo info =
            new HttpTestServersInfo.Builder().
                setServers( getServerInfos() ).
                build();
        
        pipe.send(
            MingleModels.getEmptySymbolMap(),
            encode( info, msg ),
            new AbstractTask() {
                protected void runImpl() { receiveNext( pipe ); }
            }
        );
    }

    private
    void
    close( MingleMessagePipe pipe )
    {
        pipe.send(
            MingleModels.getEmptySymbolMap(),
            new AbstractTask() {
                protected void runImpl() {
                    rtCtx.stop( new AbstractTask() {
                        protected void runImpl() { exit( 0 ); }
                    });
                }
            }
        );
    }

    private
    void
    processCommand( MingleIdentifier cmd,
                    MingleMessagePipe.Message msg,
                    MingleMessagePipe pipe )
        throws Exception
    {
        if ( cmd.equals( CMD_GET_HTTP_SERVERS ) ) sendHttpServers( msg, pipe );
        else if ( cmd.equals( CMD_CLOSE ) ) close( pipe );
    }

    private
    void
    receiveNext( final MingleMessagePipe pipe )
    {
        pipe.receive( new ObjectReceiver< MingleMessagePipe.Message >() {
            public void receive( MingleMessagePipe.Message msg )
                throws Exception
            {
                MingleIdentifier cmd = expectId( msg, ID_COMMAND );
                processCommand( cmd, msg, pipe );
            }
        });
    }

    protected
    void
    startApp()
        throws Exception
    {
        codecFact = MingleCodecFactories.loadDefault();

        TestRuntimeContexts.loadDefault(
            Runtime.getRuntime().availableProcessors(),
            getActivityContext(),
            new ObjectReceiver< TestRuntimeContext >() {
                public void receive( TestRuntimeContext rtCtx ) 
                {
                    HttpTestServer.this.rtCtx = rtCtx;

                    receiveNext(
                        behavior( MinglePipeManager.class ).manageStandardIo() 
                    );
                }
            }
        );
    }
}
