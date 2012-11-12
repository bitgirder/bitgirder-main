package com.bitgirder.mingle.service;

import static com.bitgirder.mingle.service.TestServiceConstants.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.StandardThread;
import com.bitgirder.lang.Completion;
import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.test.Test;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.TestRuntime;

import com.bitgirder.testing.Testing;

import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessActivity;
import com.bitgirder.process.ProcessRpcServer;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.ExecutorProcess;
import com.bitgirder.process.Processes;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.mingle.model.AbstractExchangerTest;
import com.bitgirder.mingle.model.MingleException;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleList;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleNamespace;
import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleServiceResponse;
import com.bitgirder.mingle.model.MingleString;
import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleStructBuilder;
import com.bitgirder.mingle.model.MingleSymbolMap;
import com.bitgirder.mingle.model.MingleTypeCastException;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.MingleValidationException;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.ModelTestInstances;

import com.bitgirder.mingle.parser.MingleParsers;

import java.util.List;

@Test
public
final
class MingleServiceTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static ObjectPath< MingleIdentifier > PATH1 =
        ObjectPath.getRoot( MingleIdentifier.create( "id1" ) ).
            descend( MingleIdentifier.create( "id2" ) ).
            startImmutableList().
            next().
            descend( MingleIdentifier.create( "id3" ) );
    
    private
    final
    class ExchangerTest
    extends AbstractExchangerTest< ExchangerTest >
    {
        private ExchangerTest( CharSequence lbl ) { super( lbl ); }

        private
        void
        assertEqual( Throwable e1,
                     Throwable e2 )
        {
            state.equal( e1.getMessage(), e2.getMessage() );
        }

        private
        void
        assertEqual( NoSuchNamespaceException e1,
                     NoSuchNamespaceException e2 )
        {
            assertEqual( (Throwable) e1, (Throwable) e2 );

            state.equalString(
                e1.getNamespace().getExternalForm(),
                e2.getNamespace().getExternalForm()
            );
        }

        private
        void
        assertEqual( NoSuchServiceException e1,
                     NoSuchServiceException e2 )
        {
            assertEqual( (Throwable) e1, (Throwable) e2 );

            state.equalString(
                e1.getService().getExternalForm(),
                e2.getService().getExternalForm()
            );
        }

        private
        void
        assertEqual( NoSuchOperationException e1,
                     NoSuchOperationException e2 )
        {
            assertEqual( (Throwable) e1, (Throwable) e2 );

            state.equalString(
                e1.getOperation().getExternalForm(),
                e2.getOperation().getExternalForm()
            );
        }

        private
        void
        assertEqual( MingleServiceRequest req1,
                     MingleServiceRequest req2 )
            throws Exception
        {
            state.equal( req1.getNamespace(), req2.getNamespace() );
            state.equal( req1.getService(), req2.getService() );
            state.equal( req1.getOperation(), req2.getOperation() );

            MingleSymbolMap p1 = req1.getParameters();
            MingleSymbolMap p2 = req2.getParameters();

            if ( state.sameNullity( p1, p2 ) )
            {
                ModelTestInstances.assertEqual( p1, p2 );
            }

            MingleValue auth1 = req1.getAuthentication();
            MingleValue auth2 = req2.getAuthentication();

            if ( state.sameNullity( auth1, auth2 ) )
            {
                ModelTestInstances.assertEqual( auth1, auth2 );
            }
        }

        private
        void
        assertEqual( MingleServiceResponse resp1,
                     MingleServiceResponse resp2 )
            throws Exception
        {
            if ( resp1.isOk() )
            {
                MingleValue res1 = resp1.getResult();
                MingleValue res2 = resp2.getResult();

                if ( state.sameNullity( res1, res2 ) )
                {
                    ModelTestInstances.assertEqual( res1, res2 );
                }
            }
            else
            {
                ModelTestInstances.
                    assertEqual( resp1.getException(), resp2.getException() );
            }
        }

        protected
        void
        assertExchange( Object jvObj1,
                        Object jvObj2 )
            throws Exception
        {
            Class< ? > cls = jvObj1.getClass();

            if ( cls.equals( InternalServiceException.class ) )
            {
                state.cast( InternalServiceException.class, jvObj2 );
            }
            else if ( cls.equals( NoSuchNamespaceException.class ) )
            {
                assertEqual( 
                    (NoSuchNamespaceException) jvObj1,
                    (NoSuchNamespaceException) jvObj2 );
            }
            else if ( cls.equals( NoSuchServiceException.class ) )
            {
                assertEqual( 
                    (NoSuchServiceException) jvObj1,
                    (NoSuchServiceException) jvObj2 );
            }
            else if ( cls.equals( NoSuchOperationException.class ) )
            {
                assertEqual( 
                    (NoSuchOperationException) jvObj1,
                    (NoSuchOperationException) jvObj2 );
            }
            else if ( cls.equals( AuthenticationMissingException.class ) )
            {
                assertEqual(
                    (AuthenticationMissingException) jvObj1,
                    (AuthenticationMissingException) jvObj2
                );
            }
            else if ( cls.equals( MingleServiceRequest.class ) )
            {
                assertEqual(
                    (MingleServiceRequest) jvObj1,
                    (MingleServiceRequest) jvObj2
                );
            }
            else if ( cls.equals( MingleServiceResponse.class ) )
            {
                assertEqual(
                    (MingleServiceResponse) jvObj1,
                    (MingleServiceResponse) jvObj2 
                );
            }
            else state.fail( "Unhandled:", jvObj1 );
        }
    }

    // By design this is just a reproduction of the logic inside the exchanger
    private
    static
    MingleStruct
    asMgReq( MingleServiceRequest req )
    {
        MingleStructBuilder b =
            MingleModels.structBuilder().
                setType( MingleServices.TYPE_REF_SERVICE_REQUEST );
        
        b.f().setString( "namespace", req.getNamespace().getExternalForm() );
        b.f().setString( "service", req.getService().getExternalForm() );
        b.f().setString( "operation", req.getOperation().getExternalForm() );

        if ( req.getAuthentication() != null )
        {
            b.f().set( "authentication", req.getAuthentication() );
        }

        b.f().set( "parameters", req.getParameters() );

        return b.build();
    }

    private
    ExchangerTest
    svcReqExchangerTest( MingleServiceRequest req,
                         int reqId )
    {   
        return
            new ExchangerTest( "test-svc-req" + reqId ).
                setJvObj( req ).
                setMgVal( asMgReq( req ) ).
                setExchanger(
                    MingleServices.exchangerFor(
                        MingleServices.TYPE_REF_SERVICE_REQUEST ) 
                );
    }

    private
    ExchangerTest
    svcRespExchangerTest( MingleServiceResponse resp,
                          int respId )
    {
        MingleStructBuilder b =
            MingleModels.structBuilder().
                setType( MingleServices.TYPE_REF_SERVICE_RESPONSE );
 
        if ( resp.isOk() ) b.f().set( "result", resp.getResult() );
        else b.f().set( "exception", resp.getException() );

        return
            new ExchangerTest( "test-svc-resp" + respId ).
                setJvObj( resp ).
                setMgVal( b.build() ).
                setExchanger(
                    MingleServices.exchangerFor(
                        MingleServices.TYPE_REF_SERVICE_RESPONSE )
                );
    }

    @InvocationFactory
    private
    List< ExchangerTest >
    testExchange()
    {
        return Lang.asList(
            
            new ExchangerTest( "internal-service-exception" ).
                setJvObj( new InternalServiceException() ).
                setExchanger( 
                    MingleServices.exchangerFor(
                        MingleServices.TYPE_REF_INTERNAL_SERVICE_EXCEPTION )
                ).
                setMgVal(
                    MingleModels.exceptionBuilder().
                        setType( "service@v1/InternalServiceException" ).
                        build()
                ),
            
            new ExchangerTest( "no-such-ns-exception" ).
                setJvObj( 
                    new NoSuchNamespaceException( 
                        MingleNamespace.create( "ns1:ns2@v1" ) )
                ).
                setExchanger(
                    MingleServices.exchangerFor(
                        MingleServices.TYPE_REF_NO_SUCH_NAMESPACE_EXCEPTION )
                ).
                setMgVal(
                    MingleModels.exceptionBuilder().
                        setType( "service@v1/NoSuchNamespaceException" ).
                        f().setString( "namespace", "ns1:ns2@v1" ).
                        build()
                ),
            
            new ExchangerTest( "no-such-svc-exception" ).
                setJvObj( 
                    new NoSuchServiceException( 
                        MingleIdentifier.create( "id1" ) )
                ).
                setExchanger(
                    MingleServices.exchangerFor(
                        MingleServices.TYPE_REF_NO_SUCH_SERVICE_EXCEPTION )
                ).
                setMgVal(
                    MingleModels.exceptionBuilder().
                        setType( "service@v1/NoSuchServiceException" ).
                        f().setString( "service", "id1" ).
                        build()
                ),
            
            new ExchangerTest( "no-such-op-exception" ).
                setJvObj( 
                    new NoSuchOperationException( 
                        MingleIdentifier.create( "id1" ) )
                ).
                setExchanger(
                    MingleServices.exchangerFor(
                        MingleServices.TYPE_REF_NO_SUCH_OPERATION_EXCEPTION )
                ).
                setMgVal(
                    MingleModels.exceptionBuilder().
                        setType( "service@v1/NoSuchOperationException" ).
                        f().setString( "operation", "id1" ).
                        build()
                ),
            
            new ExchangerTest( "authentication-missing-exception" ).
                setJvObj( new AuthenticationMissingException() ).
                setExchanger(
                    MingleServices.authExceptionExchangerFor( 
                        MingleServices.TYPE_REF_AUTHENTICATION_MISSING_EXCEPTION
                    )
                ).
                setMgVal(
                    MingleModels.exceptionBuilder().
                        setType( "service@v1/AuthenticationMissingException" ).
                        build()
                ),
            
            new ExchangerTest( "authentication-missing-exception-with-msg" ).
                setJvObj( new AuthenticationMissingException( "msg1" ) ).
                setExchanger(
                    MingleServices.authExceptionExchangerFor( 
                        MingleServices.TYPE_REF_AUTHENTICATION_MISSING_EXCEPTION
                    )
                ).
                setMgVal(
                    MingleModels.exceptionBuilder().
                        setType( "service@v1/AuthenticationMissingException" ).
                        f().setString( "message", "msg1" ).
                        build()
                ),
 
            svcReqExchangerTest( ModelTestInstances.TEST_SVC_REQ1, 1 ),
            svcReqExchangerTest( ModelTestInstances.TEST_SVC_REQ2, 2 ),
            svcReqExchangerTest( ModelTestInstances.TEST_SVC_REQ3, 3 ),

            svcRespExchangerTest( ModelTestInstances.TEST_SVC_RESP1, 1 ),
            svcRespExchangerTest( ModelTestInstances.TEST_SVC_RESP2, 2 )
        );
    }

    @Test( expected = MingleValidationException.class,
           expectedPattern =
            "\\QGot non-null values for both result and exception\\E" )
    private
    void
    testAmbiguousServiceResponseDetection()
    {
        MingleServices.exchangerFor( MingleServices.TYPE_REF_SERVICE_RESPONSE ).
            asJavaValue(
                MingleModels.structBuilder().
                    setType( MingleServices.TYPE_REF_SERVICE_RESPONSE ).f().
                    setString( "result", "blah" ).f().
                    set( "exception",
                        MingleModels.exceptionBuilder().
                            setType( "ns@v1/E1" ).
                            build()
                    ).
                    build(),
                ObjectPath.< MingleIdentifier >getRoot()
            );
    }
}
