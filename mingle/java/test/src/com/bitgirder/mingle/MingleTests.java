package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.test.Test;

@Test
final
class MingleTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }
 
    private final static MingleTimestamp TEST_TIMESTAMP1;
 
    private final static MingleTimestamp TEST_TIMESTAMP2;

    private final static String TEST_TIMESTAMP1_STRING =
        "2007-08-24T13:15:43.123450000-08:00";

    private final static String TEST_TIMESTAMP2_STRING =
        "2007-08-24T13:15:43.000000000-08:00";

    private
    void
    assertValueClassFor( MingleTypeReference typ,
                         Class< ? extends MingleValue > expct )
    {
        state.equal( Mingle.valueClassFor( typ ), expct );
    }

    @Test
    private
    void
    testJavaClassFor()
    {
        assertValueClassFor( Mingle.TYPE_BOOLEAN, MingleBoolean.class );
        assertValueClassFor( Mingle.TYPE_INT32, MingleInt32.class );
        assertValueClassFor( Mingle.TYPE_INT64, MingleInt64.class );
        assertValueClassFor( Mingle.TYPE_UINT32, MingleUint32.class );
        assertValueClassFor( Mingle.TYPE_UINT64, MingleUint64.class );
        assertValueClassFor( Mingle.TYPE_FLOAT32, MingleFloat32.class );
        assertValueClassFor( Mingle.TYPE_FLOAT64, MingleFloat64.class );
        assertValueClassFor( Mingle.TYPE_STRING, MingleString.class );
        assertValueClassFor( Mingle.TYPE_BUFFER, MingleBuffer.class );
        assertValueClassFor( Mingle.TYPE_TIMESTAMP, MingleTimestamp.class );
        assertValueClassFor( Mingle.TYPE_VALUE, MingleValue.class );
        assertValueClassFor( Mingle.TYPE_NULL, MingleNull.class );

        assertValueClassFor(
            new AtomicTypeReference( new DeclaredTypeName( "Blah" ), null ),
            null
        );
    }

    private
    void
    assertInspection( MingleValue mv,
                      CharSequence expct )
    {
        state.equalString( expct, Mingle.inspect( mv ) );
    }

    @Test
    private
    void
    testInspection()
    {
        assertInspection( MingleBoolean.TRUE, "true" );
        assertInspection( MingleBoolean.FALSE, "false" );
        assertInspection( new MingleInt32( 1 ), "1" );
        assertInspection( new MingleUint32( 2 ), "2" );
        assertInspection( new MingleInt64( -1 ), "-1" );
        assertInspection( new MingleUint64( 1 ), "1" );
        assertInspection( new MingleFloat32( 1.1f ), "1.1" );
        assertInspection( new MingleFloat64( 1.1 ), "1.1" );
        assertInspection( new MingleString( "" ), "\"\"" );
        assertInspection( new MingleString( "abc\t\rd" ), "\"abc\\t\\rd\"" );
        assertInspection( TEST_TIMESTAMP1, TEST_TIMESTAMP1_STRING );

        assertInspection( new MingleBuffer( new byte[] {} ), "buffer:[]" );
        assertInspection( 
            new MingleBuffer( new byte[] { (byte) 0, (byte) 1 } ), 
            "buffer:[0001]" 
        );

        assertInspection( MingleNull.getInstance(), "null" );
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
    }
}
