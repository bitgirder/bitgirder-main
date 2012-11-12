package com.bitgirder.mingle.bind;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.process.ProcessActivity;
import com.bitgirder.process.ProcessFailureTarget;

import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleException;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleNamespace;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleSymbolMap;
import com.bitgirder.mingle.model.MingleSymbolMapBuilder;
import com.bitgirder.mingle.model.MingleNull;
import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleServiceResponse;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.FieldDefinition;

import com.bitgirder.mingle.service.MingleRpcClient;
import com.bitgirder.mingle.service.AbstractMingleRpcClientHandler;

public
abstract
class BoundServiceClient
extends ProcessActivity
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static ObjectPath< String > PARAM_PATH_ROOT =
        ObjectPath.getRoot();

    private final static ObjectPath< MingleIdentifier > RESP_PATH_ROOT =    
        ObjectPath.getRoot( MingleServiceResponse.IDENT_RESULT );
    
    private final static ObjectPath< MingleIdentifier > EXCEPTION_PATH_ROOT =
        ObjectPath.getRoot( MingleServiceResponse.IDENT_EXCEPTION );

    private final MingleNamespace ns;
    private final MingleIdentifier svcId;
    private final MingleTypeReference authInputType;
    private final Object authObj;
    private final MingleRpcClient rpcCli;
    private final MingleBinder mb;

    protected
    BoundServiceClient( Builder b )
    {
        super( inputs.notNull( b, "b" ) );

        this.ns = inputs.notNull( b.ns, "ns" );
        this.svcId = inputs.notNull( b.svcId, "svcId" );
        this.rpcCli = inputs.notNull( b.rpcCli, "rpcCli" );
        this.authInputType = b.authInputType;
        this.authObj = b.authObj;
        this.mb = inputs.notNull( b.mb, "mb" );
    }

    public MingleBinder getBinder() { return mb; }

    public
    abstract
    class AbstractCall< V, O extends AbstractCall< V, O > >
    {
        private final MingleTypeReference respTyp;
        private final boolean useOpaqueRetType;

        private ObjectReceiver< ? super V > recv;
        private MingleNamespace ns = BoundServiceClient.this.ns;
        private MingleIdentifier svcId = BoundServiceClient.this.svcId;
        private MingleIdentifier op;
        private Object authObj;
        private ProcessFailureTarget ft = BoundServiceClient.this;
        private MingleSymbolMap params;

        protected 
        AbstractCall( MingleTypeReference respTyp,
                      boolean useOpaqueRetType )
        {
            this.respTyp = respTyp;
            this.useOpaqueRetType = useOpaqueRetType;
        }

        private O castThis() { return Lang.< O >castUnchecked( this ); }

        public
        final
        O
        receiveWith( ObjectReceiver< ? super V > recv )
        {
            this.recv = inputs.notNull( recv, "recv" );
            return castThis();
        }

        public
        final
        O
        setNamespace( MingleNamespace ns )
        {
            this.ns = inputs.notNull( ns, "ns" );
            return castThis();
        }

        public
        final
        O
        setServiceId( MingleIdentifier svcId )
        {
            this.svcId = inputs.notNull( svcId, "svcId" );
            return castThis();
        }

        public
        final
        O
        setOperation( MingleIdentifier op )
        {
            this.op = inputs.notNull( op, "op" );
            return castThis();
        }

        protected
        final
        O
        implSetAuthentication( Object authObj,
                               String paramName )
        {
            this.authObj = inputs.notNull( authObj, paramName );
            return castThis();
        }

        // Currently only exposing for testing purposes; see handling in
        // asMingleAuth
        final
        O
        setMingleAuthentication( MingleValue authObj )
        {
            return implSetAuthentication( authObj, "authObj" );
        }

        public
        final
        O
        setFailureTarget( ProcessFailureTarget ft )
        {
            this.ft = inputs.notNull( ft, "ft" );
            return castThis();
        }

        public
        final
        O
        setParameters( MingleSymbolMap params )
        {
            this.params = inputs.notNull( params, "params" );
            return castThis();
        }

        protected
        final
        void
        setParameter( MingleSymbolMapBuilder b,
                      FieldDefinition fd,
                      String jvId,
                      Object jVal,
                      MingleBinder mb,
                      ObjectPath< String > path,
                      boolean useOpaque )
        {
            MingleBinders.setField( fd, jvId, jVal, b, mb, path, useOpaque );
        }

        protected
        abstract
        void
        implSetParameters( MingleSymbolMapBuilder b,
                           MingleBinder mb,
                           ObjectPath< String > path );

        private
        MingleServiceRequest.Builder
        reqBuilder()
        {
            return 
                new MingleServiceRequest.Builder().
                    setNamespace( ns ).
                    setService( svcId ).
                    setOperation( op );
        }

        private
        MingleValue
        asMingleAuth( Object authObj )
        {
            if ( authObj instanceof MingleValue ) return (MingleValue) authObj;
            else 
            {
                return 
                    MingleBinders.asMingleValue( mb, authInputType, authObj );
            }
        }

        private
        void
        setAuth( MingleServiceRequest.Builder b )
        {
            Object authObj = this.authObj;
            if ( authObj == null ) authObj = BoundServiceClient.this.authObj;

            if ( authObj != null )
            {
                state.isFalse( 
                    authInputType == null,
                    "service does not take authentication" );

                b.setAuthentication( asMingleAuth( authObj ) );
            }
        }

        private
        MingleServiceRequest
        buildRequest()
        {
            MingleServiceRequest.Builder b = reqBuilder();

            if ( params == null )
            {
                implSetParameters( b.params(), mb, PARAM_PATH_ROOT );
            }
            else
            {
                for ( MingleIdentifier fld : params.getFields() )
                {
                    b.p().set( fld, params.get( fld ) );
                }
            }

            setAuth( b );

            return b.build();
        }

        private
        final
        class ClientHandler
        extends AbstractMingleRpcClientHandler
        {
            private final ProcessFailureTarget ft;
            
            private ClientHandler( ProcessFailureTarget ft ) { this.ft = ft; }

            private void ftFail( Throwable th ) { ft.fail( th ); }

            @Override protected void rpcFailed( Throwable th ) { ftFail( th ); }

            // runtime typing is enforced by mingle so we're okay with unchecked
            // cast for non-null values
            private
            V
            asJavaRespVal( MingleValue respVal )
            {
                if ( ( respVal == null || respVal instanceof MingleNull ) &&
                     useOpaqueRetType )
                {
                    return null;
                }
                else
                {
                    return
                        Lang.< V >castUnchecked(
                            MingleBinders.
                                asJavaValue( 
                                    mb, 
                                    respTyp, 
                                    respVal, 
                                    RESP_PATH_ROOT, 
                                    useOpaqueRetType ) 
                        );
                }
            }

            private
            void
            receiveSuccess( MingleValue respVal )
            {
                try { recv.receive( asJavaRespVal( respVal ) ); }
                catch ( Throwable th ) { ftFail( th ); }
            }

            private
            void
            receiveException( MingleException me )
            {
                Throwable th = (Throwable) 
                    MingleBinders.asJavaValue( 
                        mb, 
                        me.getType(), 
                        me, 
                        EXCEPTION_PATH_ROOT 
                    );
                
                ftFail( th );
            }

            @Override
            protected 
            void 
            rpcSucceeded( MingleServiceResponse resp ) 
            {
                if ( resp.isOk() ) receiveSuccess( resp.getResult() );
                else receiveException( resp.getException() );
            }
        } 

        private
        void
        startCall( MingleServiceRequest req )
        {
            rpcCli.beginRpc( req, new ClientHandler( ft ) );
        }

        // Can be re-used (serially only of course, and always from within the
        // associated process thread) and will make the call with the most
        // recently-set parameters. Parameters may be changed and a call c2 may
        // be started after a previous call c1 even if c1 has not completed.
        public
        final
        void
        call()
        {
            MingleServiceRequest req = buildRequest();
//            code( "Built req:", MingleModels.inspect( req, false ) );

            startCall( req );
        }
    }
 
    public
    abstract
    static
    class Builder< C extends BoundServiceClient, B extends Builder< C, B > >
    extends ProcessActivity.Builder< B >
    {
        private MingleNamespace ns;
        private MingleIdentifier svcId;
        private MingleRpcClient rpcCli;
        private MingleTypeReference authInputType;
        private Object authObj;
        private MingleBinder mb;

        public
        final
        B
        setNamespace( MingleNamespace ns )
        {
            this.ns = inputs.notNull( ns, "ns" );
            return castThis();
        }

        public
        final
        B
        setServiceId( MingleIdentifier svcId )
        {
            this.svcId = inputs.notNull( svcId, "svcId" );
            return castThis();
        }

        public
        final
        B
        setRpcClient( MingleRpcClient rpcCli )
        {
            this.rpcCli = inputs.notNull( rpcCli, "rpcCli" );
            return castThis();
        }

        public
        final
        B
        setBinder( MingleBinder mb )
        {
            this.mb = inputs.notNull( mb, "mb" );
            return castThis();
        }

        protected
        final
        void
        implSetAuthInputType( MingleTypeReference authInputType )
        {
            this.authInputType = 
                state.notNull( authInputType, "authInputType" );
        }

        protected
        final
        B
        implSetAuthentication( Object authObj,
                               String paramName )
        {
            this.authObj = inputs.notNull( authObj, paramName );
            return castThis();
        }

        public
        abstract
        C
        build();
    }
}
