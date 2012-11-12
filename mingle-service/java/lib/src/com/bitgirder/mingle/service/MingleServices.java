package com.bitgirder.mingle.service;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Completion;
import com.bitgirder.lang.TypedString;

import com.bitgirder.lang.reflect.ReflectUtils;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPathFormatter;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.parser.SyntaxException;

import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.ProcessRpcClient;
import com.bitgirder.process.ProcessActivity;

import com.bitgirder.mingle.parser.MingleParsers;

import com.bitgirder.mingle.model.AbstractExceptionExchanger;
import com.bitgirder.mingle.model.AbstractStructExchanger;
import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.ImplFactory;
import com.bitgirder.mingle.model.MingleException;
import com.bitgirder.mingle.model.MingleExceptionBuilder;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleList;
import com.bitgirder.mingle.model.MingleListIterator;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleNamespace;
import com.bitgirder.mingle.model.MingleNull;
import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleServiceResponse;
import com.bitgirder.mingle.model.MingleString;
import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleStructBuilder;
import com.bitgirder.mingle.model.MingleSymbolMap;
import com.bitgirder.mingle.model.MingleSymbolMapAccessor;
import com.bitgirder.mingle.model.MingleTypeCastException;
import com.bitgirder.mingle.model.MingleTypeName;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.MingleValidation;
import com.bitgirder.mingle.model.MingleValidationException;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleValueExchanger;

import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;
import java.lang.annotation.ElementType;
import java.lang.annotation.Annotation;

import java.lang.reflect.Method;
import java.lang.reflect.AnnotatedElement;

import java.util.Collection;
import java.util.Map;
import java.util.List;
import java.util.Set;

public
final
class MingleServices
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    final static ImplFactory IMPL_FACT = 
        new ImplFactory( new FactoryAccessor() );

    public final static MingleNamespace SERVICE_NS =
        MingleParsers.createNamespace( "service@v1" );

    private
    static
    AtomicTypeReference
    atomicType( CharSequence typNm )
    {
        return
            AtomicTypeReference.create(
                MingleTypeName.create( typNm ).resolveIn( SERVICE_NS ) );
    } 

    public final static AtomicTypeReference 
        TYPE_REF_INTERNAL_SERVICE_EXCEPTION =
            atomicType( "InternalServiceException" );
    
    public final static AtomicTypeReference
        TYPE_REF_NO_SUCH_NAMESPACE_EXCEPTION =
            atomicType( "NoSuchNamespaceException" );

    public final static AtomicTypeReference
        TYPE_REF_NO_SUCH_SERVICE_EXCEPTION =
            atomicType( "NoSuchServiceException" );

    public final static AtomicTypeReference
        TYPE_REF_NO_SUCH_OPERATION_EXCEPTION =
            atomicType( "NoSuchOperationException" );
 
    public final static AtomicTypeReference
        TYPE_REF_AUTHENTICATION_MISSING_EXCEPTION =
            atomicType( "AuthenticationMissingException" );
    
    public final static AtomicTypeReference
        TYPE_REF_SERVICE_REQUEST = atomicType( "ServiceRequest" );

    public final static AtomicTypeReference
        TYPE_REF_SERVICE_RESPONSE = atomicType( "ServiceResponse" );

    private final static Map< MingleTypeReference, MingleValueExchanger >
        EXCHANGERS;

    private final static Map< MingleTypeReference, MingleValueExchanger >
        AUTH_EXCEPTION_EXCHANGERS;

    private final static ObjectPath< MingleIdentifier > MG_ROOT =
        ObjectPath.getRoot();

    private final static ObjectPath< String > JV_ROOT = ObjectPath.getRoot();

    private final static Map< Class< ? >, MingleValueExchanger > JAVA_TYPES;

    public
    final
    static
    class OverloadedOperationException
    extends RuntimeException
    {
        OverloadedOperationException( MingleIdentifier op )
        {
            super( state.notNull( op, "op" ).getExternalForm().toString() );
        }
    }

    // Returns a MingleException if th can be mapped to an exception in the
    // service namespace; returns null otherwise.
    public
    static
    MingleException
    asServiceException( Throwable th )
    {
        inputs.notNull( th, "th" );

        MingleValueExchanger e = JAVA_TYPES.get( th.getClass() );

        if ( e == null ) return null;
        else 
        {
            ObjectPath< String > path = ObjectPath.getRoot();
            return (MingleException) e.asMingleValue( th, path );
        }
    }

    public
    static
    MingleException
    getInternalServiceException()
    {
        return asServiceException( new InternalServiceException() );
    }

    @Retention( RetentionPolicy.RUNTIME )
    @Target( { ElementType.METHOD, ElementType.TYPE } )
    public
    static
    @interface Operation
    {}

    private
    static
    MingleIdentifier
    getOperationName( CharSequence name,
                      String failTarget )
    {
        try { return MingleParsers.parseIdentifier( name ); }
        catch ( SyntaxException se )
        {
            throw new RuntimeException(
                "Couldn't parse name as a mingle identifier in " + failTarget, 
                se );
        }
    }

    static
    MingleIdentifier
    getOperationName( Method m )
    {
        state.notNull( m, "m" );

        return getOperationName( m.getName(), m.toString() );
    }

    static
    MingleIdentifier
    getOperationName( Class< ? > cls )
    {
        state.notNull( cls, "cls" );

        String clsNm = cls.getSimpleName();
        StringBuilder nm = new StringBuilder( clsNm.length() );

        nm.append( Character.toLowerCase( clsNm.charAt( 0 ) ) );
        if ( clsNm.length() > 1 ) nm.append( clsNm.substring( 1 ) );

        return getOperationName( nm, cls.toString() );
    }

    static
    MingleIdentifier
    getOperationName( AnnotatedElement elt )
    {
        state.notNull( elt, "elt" );

        if ( elt instanceof Class ) return getOperationName( (Class< ? >) elt );
        else if ( elt instanceof Method ) 
        {
            return getOperationName( (Method) elt );
        }
        else throw state.createFail( "Unsupported annotated element:", elt );
    }

    public
    static
    MingleRpcClient
    createRpcClient( AbstractProcess< ? > svc,
                     ProcessRpcClient transport )
    {
        inputs.notNull( svc, "svc" );
        inputs.notNull( transport, "transport" );

        return new DirectMingleRpcClient( svc, transport );
    }

    public
    static
    MingleRpcClient
    createRpcClient( AbstractProcess< ? > svc,
                     AbstractProcess< ? > caller )
    {
        inputs.notNull( caller, "caller" );

        return
            createRpcClient( svc, caller.behavior( ProcessRpcClient.class ) );
    }

    public
    static
    MingleRpcClient
    createRpcClient( AbstractProcess< ? > svc,
                     ProcessActivity.Context ctx )
    {
        inputs.notNull( ctx, "ctx" );

        return createRpcClient( svc, ctx.behavior( ProcessRpcClient.class ) );
    }

    public
    final
    static
    class ControlName
    extends TypedString< ControlName >
    {
        public ControlName( CharSequence cs ) { super( cs, "cs" ); }
    }

    private static final class FactoryAccessor { private FactoryAccessor() {} }

    static
    MingleValueExchanger
    exchangerFor( MingleTypeReference ref )
    {
        inputs.notNull( ref, "ref" );
        return inputs.get( EXCHANGERS, ref, "EXCHANGERS" );
    }

    private
    static
    MingleStruct
    asMingleStruct( Object obj,
                    MingleTypeReference ref )
    {
        return (MingleStruct) exchangerFor( ref ).asMingleValue( obj, JV_ROOT );
    }

    public
    static
    MingleStruct
    asMingleStruct( MingleServiceRequest req )
    {
        inputs.notNull( req, "req" );
        return asMingleStruct( req, TYPE_REF_SERVICE_REQUEST );
    }

    private
    static
    < V >
    V
    fromMingleStruct( Class< V > cls,
                      MingleStruct ms,
                      MingleTypeReference typeRef )
    {
        Object obj = exchangerFor( typeRef ).asJavaValue( ms, MG_ROOT );
        return cls.cast( obj );
    }

    public
    static
    < V >
    V
    fromMingleStruct( Class< V > cls,
                      MingleStruct ms )
    {
        inputs.notNull( cls, "cls" );
        inputs.notNull( ms, "ms" );

        if ( cls.equals( MingleServiceRequest.class ) )
        {
            return fromMingleStruct( cls, ms, TYPE_REF_SERVICE_REQUEST );
        }
        else if ( cls.equals( MingleServiceResponse.class ) )
        {
            return fromMingleStruct( cls, ms, TYPE_REF_SERVICE_RESPONSE );
        }
        else throw inputs.createFail( "Unhandled type:", cls );
    }

    public
    static
    MingleServiceResponse
    asServiceResponse( MingleStruct ms )
    {
        return fromMingleStruct( MingleServiceResponse.class, ms );
    }

    public
    static
    MingleServiceRequest
    asServiceRequest( MingleStruct ms )
    {
        return fromMingleStruct( MingleServiceRequest.class, ms );
    }

    public
    static
    MingleStruct
    asMingleStruct( MingleServiceResponse resp )
    {
        inputs.notNull( resp, "resp" );
        return asMingleStruct( resp, TYPE_REF_SERVICE_RESPONSE );
    }

    public
    static
    Iterable< MingleValueExchanger >
    getExchangers()
    {
        return EXCHANGERS.values();
    }

    public
    static
    Iterable< MingleValueExchanger >
    getAuthExceptionExchangers()
    {
        return AUTH_EXCEPTION_EXCHANGERS.values();
    }

    static
    MingleValueExchanger
    authExceptionExchangerFor( MingleTypeReference ref )
    {
        inputs.notNull( ref, "ref" );

        return 
            state.get( 
                AUTH_EXCEPTION_EXCHANGERS, ref, "AUTH_EXCEPTION_EXCHANGERS" );
    }

    private
    final
    static
    class InternalServiceExceptionExchanger
    extends AbstractExceptionExchanger< InternalServiceException >
    {
        private
        InternalServiceExceptionExchanger()
        {
            super(
                TYPE_REF_INTERNAL_SERVICE_EXCEPTION,
                InternalServiceException.class
            );
        }
        
        protected
        InternalServiceException
        buildException( MingleSymbolMapAccessor acc )
        {
            return new InternalServiceException();
        }

        protected
        MingleValue
        implAsMingleValue( InternalServiceException ex,
                           ObjectPath< String > path )
        {
            return exceptionBuilder().build();
        }
    }

    private
    final
    static
    class NoSuchNamespaceExceptionExchanger
    extends AbstractExceptionExchanger< NoSuchNamespaceException >
    {
        private final static MingleIdentifier ID_NS =
            MingleIdentifier.create( "namespace" );

        private
        NoSuchNamespaceExceptionExchanger()
        {
            super(
                TYPE_REF_NO_SUCH_NAMESPACE_EXCEPTION,
                NoSuchNamespaceException.class
            );
        }

        protected
        NoSuchNamespaceException
        buildException( MingleSymbolMapAccessor acc )
        {
            return
                new NoSuchNamespaceException( expectNamespace( acc, ID_NS ) );
        }

        protected
        MingleValue
        implAsMingleValue( NoSuchNamespaceException ex,
                           ObjectPath< String > path )
        {
            return
                exceptionBuilder().
                    f().setString( ID_NS, ex.getNamespace().getExternalForm() ).
                    build();
        }
    }

    private
    final
    static
    class NoSuchServiceExceptionExchanger
    extends AbstractExceptionExchanger< NoSuchServiceException >
    {
        private final static MingleIdentifier ID_SVC =
            MingleIdentifier.create( "service" );

        private
        NoSuchServiceExceptionExchanger()
        {
            super(
                TYPE_REF_NO_SUCH_SERVICE_EXCEPTION,
                NoSuchServiceException.class
            );
        }

        protected
        NoSuchServiceException
        buildException( MingleSymbolMapAccessor acc )
        {
            return
                new NoSuchServiceException( expectIdentifier( acc, ID_SVC ) );
        }

        protected
        MingleValue
        implAsMingleValue( NoSuchServiceException ex,
                           ObjectPath< String > path )
        {
            return
                exceptionBuilder().
                    f().setString( ID_SVC, ex.getService().getExternalForm() ).
                    build();
        }
    }

    private
    final
    static
    class NoSuchOperationExceptionExchanger
    extends AbstractExceptionExchanger< NoSuchOperationException >
    {
        private final static MingleIdentifier ID_OP =
            MingleIdentifier.create( "operation" );

        private
        NoSuchOperationExceptionExchanger()
        {
            super(
                TYPE_REF_NO_SUCH_OPERATION_EXCEPTION,
                NoSuchOperationException.class
            );
        }

        protected
        NoSuchOperationException
        buildException( MingleSymbolMapAccessor acc )
        {
            return
                new NoSuchOperationException( expectIdentifier( acc, ID_OP ) );
        }

        protected
        MingleValue
        implAsMingleValue( NoSuchOperationException ex,
                           ObjectPath< String > path )
        {
            return
                exceptionBuilder().
                    f().setString( ID_OP, ex.getOperation().getExternalForm() ).
                    build();
        }
    }

    private
    final
    static
    class AuthenticationMissingExceptionExchanger
    extends AbstractExceptionExchanger< AuthenticationMissingException >
    {
        private
        AuthenticationMissingExceptionExchanger()
        {
            super(
                TYPE_REF_AUTHENTICATION_MISSING_EXCEPTION,
                AuthenticationMissingException.class
            );
        }

        protected
        AuthenticationMissingException
        buildException( MingleSymbolMapAccessor acc )
        {
            String msg = getMessage( acc );

            return msg == null
                ? new AuthenticationMissingException()
                : new AuthenticationMissingException( msg );
        }

        protected
        MingleException
        implAsMingleValue( AuthenticationMissingException ex,
                           ObjectPath< String > path )
        {
            MingleExceptionBuilder b = exceptionBuilder();

            String msg = ex.getMessage();
            if ( msg != null ) b.f().setString( ID_MESSAGE, msg );

            return b.build();
        }
    }

    private
    final
    static
    class ServiceRequestExchanger
    extends AbstractStructExchanger< MingleServiceRequest >
    {
        private final static MingleIdentifier ID_NAMESPACE =
            MingleIdentifier.create( "namespace" );

        private final static MingleIdentifier ID_SERVICE =
            MingleIdentifier.create( "service" );

        private final static MingleIdentifier ID_OPERATION =
            MingleIdentifier.create( "operation" );

        private final static MingleIdentifier ID_PARAMETERS =
            MingleIdentifier.create( "parameters" );

        private final static MingleIdentifier ID_AUTHENTICATION =
            MingleIdentifier.create( "authentication" );

        private
        ServiceRequestExchanger()
        {
            super( TYPE_REF_SERVICE_REQUEST, MingleServiceRequest.class );
        }

        protected
        MingleServiceRequest
        buildStruct( MingleSymbolMapAccessor acc )
        {
            MingleServiceRequest.Builder b = new MingleServiceRequest.Builder();

            b.setNamespace( expectNamespace( acc, ID_NAMESPACE ) );
            b.setService( expectIdentifier( acc, ID_SERVICE ) );
            b.setOperation( expectIdentifier( acc, ID_OPERATION ) );

            MingleSymbolMap params = acc.getMingleSymbolMap( ID_PARAMETERS );
            if ( params != null ) b.params().setAll( params );

            MingleValue auth = acc.getMingleValue( ID_AUTHENTICATION );
            if ( auth != null ) b.setAuthentication( auth );

            return b.build();
        }

        protected
        MingleValue
        implAsMingleValue( MingleServiceRequest req,
                           ObjectPath< String > path )
        {
            MingleStructBuilder b = structBuilder();

            b.f().setString( 
                ID_NAMESPACE, req.getNamespace().getExternalForm() );
            
            b.f().setString( ID_SERVICE, req.getService().getExternalForm() );

            b.f().setString( 
                ID_OPERATION, req.getOperation().getExternalForm() );
            
            b.f().set( ID_PARAMETERS, req.getParameters() );

            MingleValue mv = req.getAuthentication();
            if ( mv != null ) b.f().set( ID_AUTHENTICATION, mv );

            return b.build();
        }
    }

    private
    final
    static
    class ServiceResponseExchanger
    extends AbstractStructExchanger< MingleServiceResponse >
    {
        private final static MingleIdentifier ID_RESULT =
            MingleIdentifier.create( "result" );

        private final static MingleIdentifier ID_EXCEPTION =
            MingleIdentifier.create( "exception" );

        private
        ServiceResponseExchanger()
        {
            super( TYPE_REF_SERVICE_RESPONSE, MingleServiceResponse.class );
        }

        protected
        MingleServiceResponse
        buildStruct( MingleSymbolMapAccessor acc )
        {
            MingleValue mgRes = acc.getMingleValue( ID_RESULT );
            MingleException mgEx = acc.getMingleException( ID_EXCEPTION );

            MingleValidation.isTrue(
                mgRes == null || mgEx == null,
                acc.getPath(),
                "Got non-null values for both result and exception"
            );

            if ( mgEx == null ) 
            {
                if ( mgRes == null || mgRes instanceof MingleNull )
                {
                    return MingleServiceResponse.createSuccess();
                }
                else return MingleServiceResponse.createSuccess( mgRes );
            }
            else return MingleServiceResponse.createFailure( mgEx );
        }

        protected
        MingleValue
        implAsMingleValue( MingleServiceResponse mgResp,
                           ObjectPath< String > path )
        {
            MingleStructBuilder b = structBuilder();

            if ( mgResp.isOk() ) b.f().set( ID_RESULT, mgResp.getResult() );
            else b.f().set( ID_EXCEPTION, mgResp.getException() );

            return b.build();
        }
    }

    private
    static
    void
    putExchanger( Map< MingleTypeReference, MingleValueExchanger > m,
                  MingleValueExchanger e )
    {
        m.put( e.getMingleType(), e );
    }

    static
    {
        Map< MingleTypeReference, MingleValueExchanger > m = Lang.newMap();

        putExchanger( m, new InternalServiceExceptionExchanger() );
        putExchanger( m, new NoSuchNamespaceExceptionExchanger() );
        putExchanger( m, new NoSuchServiceExceptionExchanger() );
        putExchanger( m, new NoSuchOperationExceptionExchanger() );
        putExchanger( m, new ServiceRequestExchanger() );
        putExchanger( m, new ServiceResponseExchanger() );

        EXCHANGERS = Lang.unmodifiableMap( m );

        Map< Class< ? >, MingleValueExchanger > m2 = Lang.newMap();

        for ( MingleValueExchanger e : EXCHANGERS.values() )
        {
            Lang.putUnique( m2, e.getJavaClass(), e );
        }

        JAVA_TYPES = Lang.unmodifiableMap( m2 );
    }

    static
    {
        Map< MingleTypeReference, MingleValueExchanger > m = Lang.newMap();

        putExchanger( m, new AuthenticationMissingExceptionExchanger() );

        AUTH_EXCEPTION_EXCHANGERS = Lang.unmodifiableMap( m );
    }
}
