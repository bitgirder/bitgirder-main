package com.bitgirder.mingle.service;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleNamespace;
import com.bitgirder.mingle.model.MingleList;
import com.bitgirder.mingle.model.MingleString;

import java.util.Set;

public
final
class TestServiceConstants
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public final static MingleNamespace TEST_NS =
        MingleNamespace.create( "mingle:service:test@v1" );
 
    public final static MingleIdentifier TEST_SVC =
        MingleIdentifier.create( "test-service" );

    public final static MingleIdentifier OP_GET_TEST_STRUCT1_INST1 =
        MingleIdentifier.create( "get-test-struct1-inst1" );

    public final static MingleIdentifier OP_REVERSE_STRING =
        MingleIdentifier.create( "reverse-string" );
    
    public final static MingleIdentifier OP_OVERLOADED_OPERATION =
        MingleIdentifier.create( "overloaded-operation" );
    
    public final static MingleIdentifier OP_FILTERED_OPERATION =
        MingleIdentifier.create( "filtered-operation" );

    public final static MingleIdentifier OP_GET_TYPED_STRING_LIST =
        MingleIdentifier.create( "get-typed-string-list" );

    public final static MingleIdentifier OP_DO_DELAYED_ECHO =
        MingleIdentifier.create( "do-delayed-echo" );
    
    public final static MingleIdentifier OP_BLOCKING_CALL =
        MingleIdentifier.create( "blocking-call" );

    public final static MingleIdentifier OP_DO_AUTHENTICATED_ACTION =
        MingleIdentifier.create( "do-authenticated-action" );
    
    public final static MingleIdentifier OP_DO_AUTHORIZED_ACTION =
        MingleIdentifier.create( "do-authorized-action" );

    public final static MingleIdentifier OP_TEST_FAILURES =
        MingleIdentifier.create( "test-failures" );
 
    public final static MingleIdentifier OP_TEST_ASYNC_FAILURES =
        MingleIdentifier.create( "test-async-failures" );

    public final static MingleIdentifier OP_TEST_CHILD_FAILURES =
        MingleIdentifier.create( "test-child-failures" );

    public final static MingleIdentifier OP_FAIL_VALIDATION_EXCEPTION_INTERNAL =
        MingleIdentifier.create( "fail-validation-exception-internal" );

    public final static MingleString STRING_RETVAL1 = 
        MingleModels.asMingleString( "string retval 1" );

    public final static String THE_UNACCEPTABLE = 
        "a string that should not be passed";

    public final static String UNACCEPTABLE_STRING_VALUE_MESSAGE =
        "Unacceptable string value";

    public final static MingleIdentifier ID_STR = 
        MingleIdentifier.create( "str" );
     
    public final static MingleIdentifier ID_REVERSE =
        MingleIdentifier.create( "reverse" );

    public final static MingleIdentifier ID_COPIES =
        MingleIdentifier.create( "copies" );

    public final static MingleIdentifier ID_AUTH_TOKEN =
        MingleIdentifier.create( "auth-token" );
    
    public final static MingleIdentifier ID_AUTH_EXPIRED =
        MingleIdentifier.create( "auth-expired" );

    public final static MingleIdentifier ID_FAILURE_TYPE =
        MingleIdentifier.create( "failure-type" );
    
    public final static MingleIdentifier ID_ECHO_VALUE =
        MingleIdentifier.create( "echo-value" );
    
    public final static MingleIdentifier ID_DELAY_MILLIS = 
        MingleIdentifier.create( "delay-millis" );

    public final static MingleIdentifier ID_OP_ID =
        MingleIdentifier.create( "op-id" );

    public final static String FAIL_TYPE_EXCEPTION = "exception";
    public final static String FAIL_TYPE_ERROR = "error";

    public final static String FAIL_TYPE_TEST_EXCEPTION1_INST1 =
        "test-exception1-inst1";

    public final static String VALID_AUTH_TOKEN1 = "golden ticket";
    public final static String VALID_AUTH_TOKEN2 = "snickety snicket";

    public final static Set< String > AUTHENTICATED_TOKENS =
        Lang.unmodifiableSet(
            Lang.< String >newSet( VALID_AUTH_TOKEN1, VALID_AUTH_TOKEN2 ) );

    public final static Set< String > AUTHORIZED_TOKENS =
        Lang.unmodifiableSet( Lang.< String >newSet( VALID_AUTH_TOKEN1 ) );

    public final static MingleList TYPED_STRING_LIST1 =
        MingleList.create(
            MingleModels.asMingleString( "string-1" ),
            MingleModels.asMingleString( "string-2" ),
            MingleModels.asMingleString( "string-3" ) );

    private TestServiceConstants() {}
}
