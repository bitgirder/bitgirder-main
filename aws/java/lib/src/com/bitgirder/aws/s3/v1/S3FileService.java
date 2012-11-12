package com.bitgirder.aws.s3.v1;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.ObjectReceiver;
import com.bitgirder.lang.Lang;

import com.bitgirder.io.IoProcessor;
import com.bitgirder.io.IoProcessors;
import com.bitgirder.io.IoExceptionFactory;
import com.bitgirder.io.FileWrapper;
import com.bitgirder.io.FileReceive;
import com.bitgirder.io.FileSend;
import com.bitgirder.io.DigestProcessor;
import com.bitgirder.io.ProtocolProcessor;

import com.bitgirder.io.v1.V1Io;

import com.bitgirder.process.ProcessActivity;

import com.bitgirder.crypto.CryptoUtils;
import com.bitgirder.crypto.MessageDigester;

import com.bitgirder.http.HttpHeaderName;
import com.bitgirder.http.HttpResponseMessage;

import java.util.List;
import java.util.Map;

import java.nio.ByteBuffer;

import java.nio.channels.FileChannel;

final
class S3FileService
extends AbstractS3FileService
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static IoExceptionFactory excptFact =
        V1Io.getIoExceptionFactory();

    private final S3Client s3Cli;
    private final IoProcessor ioProc;

    private 
    S3FileService( Builder b ) 
    { 
        super( b );
        
        this.s3Cli = inputs.notNull( b.s3Cli, "s3Cli" );
        this.ioProc = inputs.notNull( b.ioProc, "ioProc" );
    }

    private
    IoProcessor.Client
    createIoClient( ProcessActivity.Context ctx )
    {
        return ioProc.createClient( ctx, excptFact );
    }

    private
    abstract
    class AbstractOp< O extends Operation >
    extends ProcessActivity
    {
        final O op;

        private IoProcessor.Client ioCli;

        private 
        AbstractOp( O op ) 
        { 
            super( activityContextFor( op ) );
            this.op = op; 
        }

        final
        IoProcessor.Client
        ioCli()
        {
            if ( ioCli == null ) ioCli = createIoClient( getActivityContext() );
            return ioCli;
        }
    }

    private
    final
    class UploadFileOp
    extends AbstractOp< UploadFile >
    {
        private UploadFileOp( UploadFile op ) { super( op ); }

        private
        final
        class UploadBodyProducer
        implements S3ObjectPutRequest.BodyProducer
        {
            private final FileChannel fc;

            private FileSend body;
            private long clen;

            private UploadBodyProducer( FileChannel fc ) { this.fc = fc; }

            public
            void
            init( ProcessActivity.Context ctx )
                throws Exception
            {
                body =
                    new FileSend.Builder().
                        setChannel( fc ).
                        setLength( clen = fc.size() ).
                        setCloseOnComplete( true ).
                        setClient( createIoClient( ctx ) ).
                        build();
            }

            public ProtocolProcessor< ByteBuffer > getBody() { return body; }
            public long getContentLength() { return clen; }
        }

        private
        void
        addMetaHeaders( S3ObjectPutRequest.Builder b )
        {
            for ( S3ObjectMetaData md : op.meta )
            {
                b.addMetaField( 
                    S3Constants.X_AMZ_META_STRING + md.key(), md.value() );
            }
        }

        private
        S3ObjectPutRequest
        createPutRequest( FileChannel fc )
        {
            S3ObjectPutRequest.Builder b = new S3ObjectPutRequest.Builder();
            
            b.setBody( new UploadBodyProducer( fc ) );

            b.setLocation( op.location );
            if ( op.md5 != null ) b.setContentMd5( op.md5 );

            String ctyp = op.contentType;
            if ( ctyp != null ) b.setContentType( ctyp );

            addMetaHeaders( b );

            return b.build();
        }

        private
        FileUploadInfo
        createResponse( S3ObjectPutResponse resp )
        {
            return 
                new FileUploadInfo.Builder().
                    setS3Response( resp.info() ).
                    setLocation( op.location ).
                    setFile( op.path ).
                    build();
        }

        private
        void
        fileOpened( FileChannel fc )
        {
            beginRpc(
                s3Cli,
                createPutRequest( fc ),
                new OpRpcHandler( op ) {
                    @Override protected void rpcSucceeded( Object resp ) 
                    {
                        op.respond( 
                            createResponse( (S3ObjectPutResponse) resp ) );
                    }
                }
            );
        }
 
        private
        void
        start()
            throws Exception
        {
            ioCli().openFile(
                new FileWrapper( op.path ),
                IoProcessors.FileOpenMode.READ,
                new ObjectReceiver< FileChannel >() {
                    public void receive( FileChannel fc ) {
                        fileOpened( fc );
                    }
                }
            );
        }
    }

    protected
    void
    start( UploadFile op )
        throws Exception
    {
        new UploadFileOp( op ).start();
    }

    private
    final
    class DownloadFileOp
    extends AbstractOp< DownloadFile >
    {
        private DownloadFileOp( DownloadFile op ) { super( op ); }

        private
        final
        class BodyHandlerImpl
        implements S3Request.BodyHandler< ByteBuffer >
        {
            private final FileChannel fc;

            private FileReceive recv;
            private DigestProcessor< ByteBuffer > digProc;

            private BodyHandlerImpl( FileChannel fc ) { this.fc = fc; }

            public
            void
            init( ProcessActivity.Context ctx )
            {
                recv =
                    new FileReceive.Builder().
                        setClient( createIoClient( ctx ) ).
                        setChannel( fc ).
                        setCloseOnComplete( true ).
                        build();
            }

            public
            ProtocolProcessor< ByteBuffer >
            getBodyReceiver( HttpResponseMessage msg )
            {
                MessageDigester md5 = CryptoUtils.createDigester( "md5" );
                return digProc = DigestProcessor.create( recv, md5 );
            }
            
            public
            ByteBuffer
            getCompletionObject()
                throws Exception
            {
                return digProc.digest();
            }
        }

        private
        S3ObjectGetRequest
        createGetRequest( FileChannel fc )
        {
            S3ObjectGetRequest.Builder b = new S3ObjectGetRequest.Builder();

            b.setLocation( op.location );
            b.setBodyHandler( new BodyHandlerImpl( fc ) );

            return b.build();
        }

        private
        FileDownloadInfo
        createResponse( S3ObjectGetResponse resp )
        {
            return
                new FileDownloadInfo.Builder().
                    setFile( op.path ).
                    setMd5( (ByteBuffer) resp.getBodyObject() ).
                    setS3Response( resp.info() ).
                    setLocation( op.location ).
                    build();
        }

        private
        void
        openedFile( FileChannel fc )
        {
            beginRpc(
                s3Cli,
                createGetRequest( fc ),
                new OpRpcHandler( op ) {
                    @Override protected void rpcSucceeded( Object resp ) 
                    {
                        op.respond(
                            createResponse( (S3ObjectGetResponse) resp ) );
                    }
                }
            );
        }

        private
        void
        start()
        {
            ioCli().openFile(
                new FileWrapper( op.path ),
                IoProcessors.FileOpenMode.TRUNCATE,
                new ObjectReceiver< FileChannel >() {
                    public void receive( FileChannel fc ) {
                        openedFile( fc );
                    }
                }
            );
        }
    }

    protected
    void
    start( DownloadFile op )
    {
        new DownloadFileOp( op ).start();
    }

    final
    static
    class Builder
    extends AbstractS3FileService.Builder< Builder >
    {
        private S3Client s3Cli;
        private IoProcessor ioProc;

        public
        Builder
        setS3Client( S3Client s3Cli )
        {
            this.s3Cli = inputs.notNull( s3Cli, "s3Cli" );
            return this;
        }

        public
        Builder
        setIoProcessor( IoProcessor ioProc )
        {
            this.ioProc = inputs.notNull( ioProc, "ioProc" );
            return this;
        }

        public
        S3FileService
        build()
        {
            return new S3FileService( this );
        }
    }
}
