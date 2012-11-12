package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.PathWiseAsserter;
import com.bitgirder.lang.path.ImmutableListPath;

import com.bitgirder.io.IoTestFactory;
import com.bitgirder.io.Base64Encoder;

import java.nio.ByteBuffer;

import java.util.Iterator;
import java.util.Set;
import java.util.GregorianCalendar;
import java.util.TimeZone;

import java.sql.Timestamp;

public
final
class ModelTestInstances
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static Base64Encoder enc = new Base64Encoder();

    private final static ObjectPath< MingleIdentifier > MG_ROOT =
        ObjectPath.getRoot();

    private final static PathWiseAsserter< MingleIdentifier > BASE_ASSERTER =
        new PathWiseAsserter< MingleIdentifier >( 
            MingleModels.getIdentifierPathFormatter() );

    private final static MingleIdentifier IDENT_NS =
        MingleIdentifier.create( "namespace" );

    private final static MingleIdentifier IDENT_SVC =
        MingleIdentifier.create( "service" );

    private final static MingleIdentifier IDENT_OP =
        MingleIdentifier.create( "operation" );

    private final static MingleIdentifier IDENT_PARAMS =
        MingleIdentifier.create( "parameters" );

    private final static MingleIdentifier IDENT_RESULT =
        MingleIdentifier.create( "result" );

    private final static MingleIdentifier IDENT_EXCEPTION =
        MingleIdentifier.create( "exception" );
    
    private final static MingleIdentifier IDENT_AUTHENTICATION =
        MingleIdentifier.create( "authentication" );

    public final static MingleNamespace TEST_NS =
        MingleNamespace.create( "mingle:test@v1" );
    
    public final static MingleTypeName TEST_STRUCT1_TYPE =
        MingleTypeName.create( "TestStruct1" );
    
    public final static MingleTypeName TEST_STRUCT2_TYPE =
        MingleTypeName.create( "TestStruct2" );
    
    public final static MingleTypeName TEST_EXCEPTION1_TYPE =
        MingleTypeName.create( "TestException1" );

    public final static MingleTypeName TEST_ENUM1_TYPE =
        MingleTypeName.create( "TestEnum1" );

    public final static MingleIdentifier TEST_SERVICE =
        MingleIdentifier.create( "a-service" );

    public final static MingleIdentifier TEST_OP =
        MingleIdentifier.create( "an-operation" );

    public final static MingleEnum TEST_ENUM1_CONSTANT1 =
        new MingleEnum.Builder().
            setType( 
                AtomicTypeReference.create(
                    TEST_ENUM1_TYPE.resolveIn( TEST_NS ) ) ).
            setValue( MingleIdentifier.create( "constant1" ) ).
            build();

    public final static MingleStruct TEST_STRUCT1_INST1;
    public final static MingleException TEST_EXCEPTION1_INST1;

    public final static ByteBuffer TEST_BYTE_BUFFER1;
 
    public final static MingleTimestamp TEST_TIMESTAMP1;
 
    public final static MingleTimestamp TEST_TIMESTAMP2;

    public final static String TEST_TIMESTAMP1_STRING =
        "2007-08-24T13:15:43.123450000-08:00";

    public final static String TEST_TIMESTAMP2_STRING =
        "2007-08-24T13:15:43.000000000-08:00";
 
    public final static Timestamp TEST_TIMESTAMP1_SQL_TIMESTAMP;

    public final static GregorianCalendar TEST_TIMESTAMP2_GREGORIAN_CALENDAR;

    public final static MingleList TEST_LIST1;

    public final static MingleSymbolMap TEST_SYM_MAP1;

    public final static MingleServiceRequest TEST_SVC_REQ1;

    // like req1 but with no parameters
    public final static MingleServiceRequest TEST_SVC_REQ2;
    
    // like req1 but with a struct for authentication
    public final static MingleServiceRequest TEST_SVC_REQ3;

    public final static MingleServiceResponse TEST_SVC_RESP1;

    public final static MingleServiceResponse TEST_SVC_RESP2;

    public final static MingleString TEST_SVC_REQ_AUTH_STR =
        MingleModels.asMingleString( "test-authentication-token" );

    private ModelTestInstances() {}

    private
    static
    void
    assertEqual( MingleString expct,
                 MingleValue actual,
                 PathWiseAsserter< MingleIdentifier > a )
    {
        a.equalString( expct, a.cast( MingleString.class, actual ) );
    }

    private
    static
    void
    assertEqual( MingleBoolean expct,
                 MingleValue actual,
                 PathWiseAsserter< MingleIdentifier > a )
    {
        a.equal( expct, a.cast( MingleBoolean.class, actual ) );
    }

    private
    static
    void
    assertEqual( MingleTimestamp expct,
                 MingleValue actual,
                 PathWiseAsserter< MingleIdentifier > a )
        throws Exception
    {
        MingleTimestamp ts2Dbg = (MingleTimestamp) actual;

        a.equal(
            0, expct.compareTo( a.cast( MingleTimestamp.class, actual ) ) );
    }

    private
    static
    void
    assertEqual( MingleBuffer expct,
                 MingleValue actual,
                 PathWiseAsserter< MingleIdentifier > a )
        throws Exception
    {
        a.equal( 
            expct.getByteBuffer(),
            a.cast( MingleBuffer.class, actual ).getByteBuffer() );
    }

    private
    static
    void
    assertEqual( MingleNull expct,
                 MingleValue actual,
                 PathWiseAsserter< MingleIdentifier > a )
    {
        a.isTrue( actual == null || actual instanceof MingleNull );
    }

    // We allow for actual to be a MingleEnum, a fully qualified enum literal,
    // or just an identifier which is assumed to be the same as the value of
    // expct
    private
    static
    void
    assertEqual( MingleEnum expct,
                 MingleValue actual,
                 PathWiseAsserter< MingleIdentifier > a )
    {
        AtomicTypeReference actualRef = null; // could remain so
        MingleIdentifier id;

        if ( actual instanceof MingleEnum ) 
        {
            id = ( (MingleEnum) actual ).getValue();
        }
        else if ( actual instanceof MingleString )
        {
            String s = actual.toString();
            if ( s.indexOf( '.' ) < 0 ) id = MingleIdentifier.create( s );
            else 
            {
                MingleEnum en2 = MingleEnum.create( s );
                actualRef = en2.getType();
                id = en2.getValue();
            }
        }
        else throw a.createFail( "Invalid enum value:", actual );

        if ( actualRef != null ) a.equal( expct.getType(), actualRef );
        a.equal( expct.getValue(), id );
    }

    private
    static
    MingleValue
    expectTyped( Class< ? extends MingleValue > cls,
                 MingleListIterator it )
    {
        if ( MingleString.class.isAssignableFrom( cls ) )
        {
            return it.nextMingleString();
        }
        else if ( MingleInt64.class.isAssignableFrom( cls ) )
        {
            return it.nextMingleInt64();
        }
        else if ( MingleInt32.class.isAssignableFrom( cls ) )
        {
            return it.nextMingleInt32();
        }
        else if ( MingleDouble.class.isAssignableFrom( cls ) )
        {
            return it.nextMingleDouble();
        }
        else if ( MingleFloat.class.isAssignableFrom( cls ) )
        {
            return it.nextMingleFloat();
        }
        else if ( MingleBoolean.class.isAssignableFrom( cls ) )
        {
            return it.nextMingleBoolean();
        }
        else if ( MingleTimestamp.class.isAssignableFrom( cls ) )
        {
            return it.nextMingleTimestamp();
        }
        else if ( MingleBuffer.class.isAssignableFrom( cls ) )
        {
            return it.nextMingleBuffer();
        }
        else if ( MingleStruct.class.isAssignableFrom( cls ) )
        {
            return it.nextMingleStruct();
        }
        else if ( MingleSymbolMap.class.isAssignableFrom( cls ) )
        {
            return it.nextMingleSymbolMap();
        }
        else if ( MingleList.class.isAssignableFrom( cls ) )
        {
            return it.nextMingleList();
        }
        else if ( MingleEnum.class.isAssignableFrom( cls ) )
        {
            return it.nextMingleEnum();
        }
        else if ( MingleNull.class.isAssignableFrom( cls ) )
        {
            return it.nextMingleNull();
        }
        else throw state.createFail( "Unexpected target type:", cls );
    }

    private
    static
    void
    assertEqual( MingleList expct,
                 MingleValue actual,
                 PathWiseAsserter< MingleIdentifier > a )
        throws Exception
    {
        MingleList l2 = a.cast( MingleList.class, actual );

        Iterator< MingleValue > it1 = expct.iterator();
        MingleListIterator it2 = MingleListIterator.forList( l2, a.getPath() );

        ImmutableListPath< MingleIdentifier > p = 
            a.getPath().startImmutableList();

        int indx = 0;
        while ( it1.hasNext() && it2.hasNext() )
        {
            MingleValue mv1 = it1.next();
            MingleValue mv2 = expectTyped( mv1.getClass(), it2 );

            PathWiseAsserter< MingleIdentifier > a2 =
                new PathWiseAsserter< MingleIdentifier >( p, a.getFormatter() );

            assertEqual( mv1, mv2, a2 );
            p = p.next();
        }

        a.isFalse( it1.hasNext(), "Expected list is longer than actual" );
        a.isFalse( it2.hasNext(), "Actual list is longer than expected" );
    }

    private
    static
    void
    assertFieldSets( MingleSymbolMap expct,
                     MingleSymbolMap actual,
                     PathWiseAsserter< MingleIdentifier > a )
    {
        Set< MingleIdentifier > s1 = Lang.collectSet( expct.getFields() );
        Set< MingleIdentifier > s2 = Lang.collectSet( actual.getFields() );

        a.equal( s1, s2 );
    }

    private
    static
    MingleValue
    expectTyped( Class< ? extends MingleValue > cls,
                 MingleSymbolMapAccessor acc,
                 MingleIdentifier fld )
    {
        if ( MingleString.class.isAssignableFrom( cls ) )
        {
            return acc.expectMingleString( fld );
        }
        else if ( MingleInt64.class.isAssignableFrom( cls ) )
        {
            return acc.expectMingleInt64( fld );
        }
        else if ( MingleInt32.class.isAssignableFrom( cls ) )
        {
            return acc.expectMingleInt32( fld );
        }
        else if ( MingleDouble.class.isAssignableFrom( cls ) )
        {
            return acc.expectMingleDouble( fld );
        }
        else if ( MingleFloat.class.isAssignableFrom( cls ) )
        {
            return acc.expectMingleFloat( fld );
        }
        else if ( MingleBuffer.class.isAssignableFrom( cls ) )
        {
            return acc.expectMingleBuffer( fld );
        }
        else if ( MingleBoolean.class.isAssignableFrom( cls ) )
        {
            return acc.expectMingleBoolean( fld );
        }
        else if ( MingleStruct.class.isAssignableFrom( cls ) )
        {
            return acc.expectMingleStruct( fld );
        }
        else if ( MingleException.class.isAssignableFrom( cls ) )
        {
            return acc.expectMingleException( fld );
        }
        else if ( MingleTimestamp.class.isAssignableFrom( cls ) )
        {
            return acc.expectMingleTimestamp( fld );
        }
        else if ( MingleSymbolMap.class.isAssignableFrom( cls ) )
        {
            return acc.expectMingleSymbolMap( fld );
        }
        else if ( MingleList.class.isAssignableFrom( cls ) )
        {
            return acc.expectMingleList( fld );
        }
        else if ( MingleEnum.class.isAssignableFrom( cls ) )
        {
            return acc.expectMingleValue( fld );
        }
        else throw state.createFail( "Unexpected target type:", cls );
    }

    private
    static
    void
    assertSymbolMapContents( MingleSymbolMap expct,
                             MingleSymbolMap actual,
                             PathWiseAsserter< MingleIdentifier > a )
        throws Exception
    {
        assertFieldSets( expct, actual, a );

        MingleSymbolMapAccessor acc = 
            MingleSymbolMapAccessor.create( actual, a.getPath() );

        for ( MingleIdentifier fld : expct.getFields() )
        {
            MingleValue expctVal = expct.get( fld );

            MingleValue actualVal = 
                expectTyped( expctVal.getClass(), acc, fld );

            assertEqual( expctVal, actualVal, a.descend( fld ) );
        }
    }

    private
    static
    void
    assertEqual( MingleSymbolMap expct,
                 MingleValue actual,
                 PathWiseAsserter< MingleIdentifier > a )
        throws Exception
    {
        MingleSymbolMap m2 = (MingleSymbolMap)
            MingleModels.
                asMingleInstance( 
                    MingleModels.TYPE_REF_MINGLE_SYMBOL_MAP, actual, MG_ROOT );

        assertSymbolMapContents( expct, m2, a );
    }

    private
    static
    < S extends MingleStructure >
    void
    assertStructureEquals( S expct,
                           Class< S > expctCls, 
                           MingleValue actual,
                           PathWiseAsserter< MingleIdentifier > a )
        throws Exception
    {
        // This will fail if actual is totally uncastable to expctCls (for
        // instance, it is just a bare MingleInt64), but will allow us to
        // meaningfully handle situations, such as those which often arise in
        // roundtrip testing, in which we expect a MingleException and receive
        // it, but as a MingleStruct.
        S s2 = 
            expctCls.cast(
                MingleModels.asMingleInstance( 
                    MingleModels.typeReferenceOf( expct ), actual, a.getPath() 
                )
            );

        a.equal( 
            MingleModels.getType( expct ), 
            MingleModels.getType( actual ) );

        assertSymbolMapContents( expct.getFields(), s2.getFields(), a );
    }

    private
    static
    void
    assertEqual( MingleStruct expct,
                 MingleValue actual,
                 PathWiseAsserter< MingleIdentifier > a )
        throws Exception
    {
        assertStructureEquals( expct, MingleStruct.class, actual, a );
    }

    private
    static
    void
    assertEqual( MingleException expct,
                 MingleValue actual,
                 PathWiseAsserter< MingleIdentifier > a )
        throws Exception
    {
        assertStructureEquals( expct, MingleException.class, actual, a );
    }

    public
    static
    void
    assertEqual( MingleValue expct,
                 MingleValue actual,
                 PathWiseAsserter< MingleIdentifier > a )
        throws Exception
    {
        inputs.notNull( a, "a" );

        if ( expct instanceof MingleNull )
        {
            assertEqual( (MingleNull) expct, actual, a );
        }
        else if ( a.sameNullity( expct, actual ) )
        {
            if ( expct instanceof MingleString )
            {
                assertEqual( (MingleString) expct, actual, a );
            }
            else if ( expct instanceof MingleBoolean )
            {
                assertEqual( (MingleBoolean) expct, actual, a );
            }
            else if ( expct instanceof MingleInt64 )
            {
                a.equal( expct, a.cast( MingleInt64.class, actual ) );
            }
            else if ( expct instanceof MingleInt32 )
            {
                a.equal( expct, a.cast( MingleInt32.class, actual ) );
            }
            else if ( expct instanceof MingleDouble )
            {
                a.equal( expct, a.cast( MingleDouble.class, actual ) );
            }
            else if ( expct instanceof MingleFloat )
            {
                a.equal( expct, a.cast( MingleFloat.class, actual ) );
            }
            else if ( expct instanceof MingleBuffer )
            {
                assertEqual( (MingleBuffer) expct, actual, a );
            }
            else if ( expct instanceof MingleList )
            {
                assertEqual( (MingleList) expct, actual, a );
            }
            else if ( expct instanceof MingleStruct )
            {
                assertEqual( (MingleStruct) expct, actual, a );
            }
            else if ( expct instanceof MingleException )
            {
                assertEqual( (MingleException) expct, actual, a );
            }
            else if ( expct instanceof MingleSymbolMap )
            {
                assertEqual( (MingleSymbolMap) expct, actual, a );
            }
            else if ( expct instanceof MingleTimestamp )
            {
                assertEqual( (MingleTimestamp) expct, actual, a );
            }
            else if ( expct instanceof MingleEnum )
            {
                assertEqual( (MingleEnum) expct, actual, a );
            }
            else 
            {
                a.fail( 
                    "Don't know how to assert equality for instance of",
                    expct.getClass() );
            }
        }
    }

    public
    static
    void
    assertEqual( MingleValue expct,
                 MingleValue actual )
        throws Exception
    {
        assertEqual( expct, actual, BASE_ASSERTER );
    }

    public
    static
    void
    assertEqual( MingleServiceRequest expct,
                 MingleServiceRequest actual )
        throws Exception
    {
        PathWiseAsserter< MingleIdentifier > a = BASE_ASSERTER;

        assertEqual(
            expct.getAuthentication(), actual.getAuthentication(),
            a.descend( IDENT_AUTHENTICATION ) );

        a.descend( IDENT_NS ).
           equal( expct.getNamespace(), actual.getNamespace() );
        
        a.descend( IDENT_SVC ).
           equal( expct.getService(), actual.getService() );
        
        a.descend( IDENT_OP ).
           equal( expct.getOperation(), actual.getOperation() );
        
        assertEqual(
            expct.getParameters(), actual.getParameters(),
            a.descend( IDENT_PARAMS ) );
    }

    public
    static
    void
    assertEqual( MingleServiceResponse expct,
                 MingleServiceResponse actual )
        throws Exception
    {
        PathWiseAsserter< MingleIdentifier > a = BASE_ASSERTER;
        
        if ( a.sameNullity( expct, actual ) )
        {
            a.equal( expct.isOk(), actual.isOk() );

            if ( expct.isOk() )
            {
                assertEqual( 
                    expct.getResult(), actual.getResult(), 
                    a.descend( IDENT_RESULT ) );
            }
            else
            {
                assertEqual(
                    expct.getException(), actual.getException(),
                    a.descend( IDENT_EXCEPTION ) );
            }
        }
    }

    static
    {
        ByteBuffer bb = ByteBuffer.allocate( 150 );

        for ( int i = 0, e = bb.capacity(); i < e; ++i ) bb.put( i, (byte) i );
        TEST_BYTE_BUFFER1 = bb.asReadOnlyBuffer();
    }

    static
    {
        // base builder for timestamps 1 and 2;
        MingleTimestamp.Builder b =
            new MingleTimestamp.Builder().
                setYear( 2007 ).
                setMonth( 8 ).
                setDate( 24 ).
                setHour( 13 ).
                setMinute( 15 ).
                setSeconds( 43 ).
                setTimeZone( "GMT-08:00" );
        
        // build ts2 with no frac part
        TEST_TIMESTAMP2 = b.build();

        // build ts1 (which came first in our code) with the frac part
        b.setFraction( "12345" );
        TEST_TIMESTAMP1 = b.build();

        // Similarly to above, use ts2 rep to build a base for ts1 rep. Note
        // that we're using zero-indexed months here
        TEST_TIMESTAMP2_GREGORIAN_CALENDAR = 
            new GregorianCalendar( TimeZone.getTimeZone( "GMT-08:00" ) );
        TEST_TIMESTAMP2_GREGORIAN_CALENDAR.set( GregorianCalendar.YEAR, 2007 );
        TEST_TIMESTAMP2_GREGORIAN_CALENDAR.set( GregorianCalendar.MONTH, 7 );
        TEST_TIMESTAMP2_GREGORIAN_CALENDAR.set( GregorianCalendar.DATE, 24 );

        TEST_TIMESTAMP2_GREGORIAN_CALENDAR.set( 
            GregorianCalendar.HOUR_OF_DAY, 13 );

        TEST_TIMESTAMP2_GREGORIAN_CALENDAR.set( GregorianCalendar.MINUTE, 15 );
        TEST_TIMESTAMP2_GREGORIAN_CALENDAR.set( GregorianCalendar.SECOND, 43 );

        TEST_TIMESTAMP2_GREGORIAN_CALENDAR.set( 
            GregorianCalendar.MILLISECOND, 0 );

        TEST_TIMESTAMP2_GREGORIAN_CALENDAR.setTimeZone( 
            TimeZone.getTimeZone( "GMT-08:00" ) );

        TEST_TIMESTAMP1_SQL_TIMESTAMP =
            new Timestamp( 
                TEST_TIMESTAMP2_GREGORIAN_CALENDAR.getTimeInMillis() );

        TEST_TIMESTAMP1_SQL_TIMESTAMP.setNanos( 123450000 );
    }

    static
    {
        MingleList.Builder b = new MingleList.Builder();

        for ( int i = 0; i < 5; ++i )
        {
            b.add( MingleModels.asMingleString( "string" + i ) );
        }

        TEST_LIST1 = b.build();
    }

    static
    {
        MingleSymbolMapBuilder b = MingleModels.symbolMapBuilder();

        b.setString( "string-sym1", "something to do here" );
        b.setInt64( "int-sym1", 1234 );
        b.setDouble( "decimal-sym1", 3.14 );
        b.setBoolean( "bool-sym1", false );
        b.set( "list-sym1", TEST_LIST1 );

        TEST_SYM_MAP1 = b.build();
    }

    static
    {
        MingleExceptionBuilder b = MingleModels.exceptionBuilder();
        b.setType( 
            AtomicTypeReference.create(
                TEST_EXCEPTION1_TYPE.resolveIn( TEST_NS ) ) );
        
        b.setMessage( "This is a test message" );

        b.fields().
          setInt64( "failure-count", 2 ).f().
          set( "exception-time", TEST_TIMESTAMP1 );
        
        TEST_EXCEPTION1_INST1 = b.build();
    }

    private
    static
    MingleStruct
    testStruct2( int i1 )
    {
        MingleStructBuilder b = MingleModels.structBuilder();

        b.setType(
            AtomicTypeReference.create(
                TEST_STRUCT2_TYPE.resolveIn( TEST_NS ) ) );
        
        return b.f().setInt32( "i1", i1 ).build();
    }

    static
    {
        MingleStructBuilder msb = MingleModels.structBuilder();
        msb.setType( 
            AtomicTypeReference.create(
                TEST_STRUCT1_TYPE.resolveIn( TEST_NS ) ) );

        msb.fields().
            setString( "string1", "hello" ).f().
            setBoolean( "bool1", true ).f().
            setInt64( "int1", 32234 ).f().
            setInt64( "int2", Long.MAX_VALUE ).f().
            setInt32( "int3", Integer.MAX_VALUE ).f().
            setDouble( "double1", 1.1d ).f().
            setFloat( "float1", 1.1f ).f().
            setBuffer( "buffer1", TEST_BYTE_BUFFER1 ).f().
            set( "enum1", TEST_ENUM1_CONSTANT1 ).f().
            set( "timestamp1", TEST_TIMESTAMP1 ).f().
            set( "timestamp2", TEST_TIMESTAMP2 ).f().
            set( "list1", TEST_LIST1 ).f().
            set( "symbol-map1", TEST_SYM_MAP1 ).f().
            set( "exception1", TEST_EXCEPTION1_INST1 ).f().
            set( "struct1", testStruct2( 111 ) );

        TEST_STRUCT1_INST1 = msb.build();
    }

    static
    {
        MingleServiceRequest.Builder b = 
            new MingleServiceRequest.Builder().
                setNamespace( TEST_NS ).
                setService( TEST_SERVICE ).
                setOperation( TEST_OP ).
                setAuthentication( TEST_SVC_REQ_AUTH_STR );
        
        TEST_SVC_REQ2 = b.build(); // no params
 
        b.params().
          setInt64( "an-int", 10101 ).p().
          setString( "a-string", "this is awesome" ).p().
          setBoolean( "a-boolean", false ).p().
          setDouble( "a-decimal", 34838.333 ).p().
          setBuffer( "a-buffer", TEST_BYTE_BUFFER1 ).p().
          set( "a-timestamp", TEST_TIMESTAMP1 ).p().
          set( "a-struct", TEST_STRUCT1_INST1 ).p().
          set( "a-symbol-map", TEST_SYM_MAP1 );
 
        TEST_SVC_REQ1 = b.build();

        b.setAuthentication( TEST_STRUCT1_INST1 );
        TEST_SVC_REQ3 = b.build();
    }

    static
    {
        TEST_SVC_RESP1 = 
            MingleServiceResponse.createSuccess( TEST_STRUCT1_INST1 );
        
        TEST_SVC_RESP2 =
            MingleServiceResponse.createFailure( TEST_EXCEPTION1_INST1 );
    }
}
