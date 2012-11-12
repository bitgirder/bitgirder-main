package com.bitgirder.demo.mingle;

import com.bitgirder.mingle.service.AbstractMingleService;
import com.bitgirder.mingle.service.MingleServiceCallContext;
import com.bitgirder.mingle.service.MingleServices;

import com.bitgirder.mingle.model.MingleServiceResponse;
import com.bitgirder.mingle.model.MingleInvocationValidator;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleValidation;
import com.bitgirder.mingle.model.MingleValue;

import com.bitgirder.mingle.model.MingleIdentifier;

import com.bitgirder.process.ProcessRpcServer;

import com.bitgirder.concurrent.Duration;

import java.util.List;

// An implementation of the demo service using "native" objects, ie. just those
// from com.bitgirder.mingle.model without any binding to java classes.
final
class NativeDemoService
extends AbstractMingleService
{
    // constants for parameter names we'll be using

    private final static MingleIdentifier ID_STRING = 
        MingleIdentifier.create( "string" );

    private final static MingleIdentifier ID_COPIES =
        MingleIdentifier.create( "copies" );

    private final static MingleIdentifier ID_REVERSE =
        MingleIdentifier.create( "reverse" );

    private final static MingleIdentifier ID_FIB1 =
        MingleIdentifier.create( "fib1" );

    private final static MingleIdentifier ID_FIB2 =
        MingleIdentifier.create( "fib2" );

    private final static MingleIdentifier ID_SEQ_LEN =
        MingleIdentifier.create( "seqLen" );

    NativeDemoService() {}

    // Utility method to extract an int that we expect to be present in the
    // request as the given parameter ('fld'), validate that it is positive, and
    // return it as a Java int.
    private
    int
    positiveI( MingleServiceCallContext ctx,
               MingleIdentifier fld )
    {
        // Creates a validator with error location information positioned at the
        // top-level parameters.
        MingleInvocationValidator v =
            MingleModels.createInvocationValidator( ctx.getParametersPath() );

        // Call positiveI passing in as the first parmeter the validator
        // positioned at the error location for the given parameter
        return
            MingleValidation.positiveI(
                v.field( fld ), ctx.getParameters().expectInt( fld ) );
    }

    // Implementation od do-op1. See BoundDemoService.doOp1 for the op
    // description.
    @MingleServices.Operation
    private
    MingleServiceResponse
    doOp1( MingleServiceCallContext ctx )
    {
        // get and validate params and convert to java objects
        String str = ctx.getParameters().expectString( ID_STRING );
        int copies = positiveI( ctx, ID_COPIES );
        boolean reverse = ctx.getParameters().getBoolean( ID_REVERSE );

        // get the result as a mingle value
        String res = DemoUtils.copyAndReverse( str, copies, reverse );
        MingleValue mv = MingleModels.asMingleString( res );

        // return it as a mingle service response
        return createSuccessResponse( ctx, mv );
    }

    // util method to extract and validate the sequence length
    private
    int
    getFibSeqLen( MingleServiceCallContext mgCtx,
                  MingleIdentifier fldId )
    {
        int res = positiveI( mgCtx, fldId );

        MingleValidation.isTrue(
            res >= 2, mgCtx.getParametersPath().descend( fldId ),
            "sequence length must be at least 2" );

        return res;
    }

    // Async version of do-op2, which computes a fibonacci sequence of a given
    // length. We just submit the completion with an arbitrary delay to simulate
    // a long-running computation
    @MingleServices.Operation
    private
    void
    doOp2( 
        final MingleServiceCallContext mgCtx,
        final ProcessRpcServer.ResponderContext< MingleServiceResponse > opCtx )
    {
        final int fib1 = positiveI( mgCtx, ID_FIB1 );
        final int fib2 = positiveI( mgCtx, ID_FIB2 );

        MingleValidation.isTrue( 
            fib1 <= fib2, mgCtx.getParametersPath(), ID_FIB1 + ">" + ID_FIB2 );

        final int seqLen = getFibSeqLen( mgCtx, ID_SEQ_LEN );

        submit(
            new AbstractTask() {
                protected void runImpl() 
                {
                    List< Long > fibRes = 
                        DemoUtils.getFibRes( fib1, fib2, seqLen );

                    MingleValue mv = MingleModels.asMingleValue( fibRes );

                    opCtx.respond( createSuccessResponse( mgCtx, mv ) );
                }
            },
            Duration.fromMillis( 500 )
        );
    }
}
