package com.bitgirder.mingle.service;

import static com.bitgirder.mingle.service.TestServiceConstants.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.process.AbstractVoidProcess;
import com.bitgirder.process.AbstractProcess;
import com.bitgirder.process.ProcessExit;
import com.bitgirder.process.ProcessRpcServer;
import com.bitgirder.process.ExecutorProcess;

import com.bitgirder.mingle.model.MingleServiceResponse;
import com.bitgirder.mingle.model.MingleServiceRequest;
import com.bitgirder.mingle.model.MingleSymbolMap;
import com.bitgirder.mingle.model.MingleSymbolMapAccessor;
import com.bitgirder.mingle.model.MingleBoolean;
import com.bitgirder.mingle.model.MingleString;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleValidation;
import com.bitgirder.mingle.model.MingleValidationException;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleException;
import com.bitgirder.mingle.model.ModelTestInstances;

import com.bitgirder.mingle.parser.MingleParsers;

import java.lang.annotation.Annotation;
import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

import java.util.concurrent.ConcurrentMap;

public
final
class NativeTestService
extends AbstractMingleService
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public NativeTestService() {}

    public void stop() { behavior( ProcessRpcServer.class ).stop(); }

    @MingleServices.Operation
    private
    MingleServiceResponse
    getTestStruct1Inst1( MingleServiceCallContext ctx )
        throws Exception
    {
        return 
            createSuccessResponse( ctx, ModelTestInstances.TEST_STRUCT1_INST1 );
    }

    private
    MingleString
    reverse( MingleString str )
    { 
        StringBuilder sb = new StringBuilder( str.length() );

        for ( int i = str.length() - 1; i >= 0; --i )
        {
            sb.append( str.charAt( i ) );
        }
        
        return MingleModels.asMingleString( sb );
    }

    private
    MingleString
    copiesOf( MingleString str,
              int copies )
    {
        StringBuilder res = new StringBuilder( str.length() * copies );
        for ( int i = 0; i < copies; ++i ) res.append( str );

        return MingleModels.asMingleString( res );
    }

    @MingleServices.Operation
    private
    MingleServiceResponse
    reverseString( MingleServiceCallContext ctx )
        throws Exception
    {
        MingleString str = ctx.getParameters().expectMingleString( ID_STR );
        
        MingleValidation.isFalse(
            str.toString().equals( THE_UNACCEPTABLE ),
            ctx.getParameterPath( ID_STR ), UNACCEPTABLE_STRING_VALUE_MESSAGE );

        return createSuccessResponse( ctx, reverse( str ) );
    }

    @MingleServices.Operation
    private
    final
    class DoDelayedEcho
    extends ProcessRpcServer.HandlerProcess< MingleServiceCallContext,
                                             MingleServiceResponse >
    {
        private DoDelayedEcho( MingleServiceCallContext ctx ) { super( ctx ); }

        private
        void
        sendEchoResult()
        {
            MingleValue res = 
                message().getParameters().expectMingleValue( ID_ECHO_VALUE );
            
            exit( createSuccessResponse( message(), res ) );
        }

        protected
        void
        startImpl()
        {
            Runnable task = 
                new AbstractTask() {
                    protected void runImpl() { sendEchoResult(); } };

            Duration delay =
                Duration.fromMillis(
                    message().getParameters().expectLong( ID_DELAY_MILLIS ) );
            
            submit( task, delay );
        }
    }
}
