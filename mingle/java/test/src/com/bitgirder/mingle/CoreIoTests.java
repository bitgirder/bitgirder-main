package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.test.Test;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.LabeledTestCall;

import com.bitgirder.io.IoTestFactory;
import com.bitgirder.io.IoUtils;

import java.io.ByteArrayInputStream;

import java.util.List;
import java.util.Map;

@Test
final
class CoreIoTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private static void code( Object... args ) { CodeLoggers.code( args ); }
    
    private final static byte TC_END = 0;
    private final static byte TC_INVALID_DATA_TEST = 1; 
    private final static byte TC_ROUNDTRIP_TEST = 2;
    private final static byte TC_SEQUENCE_ROUNDTRIP_TEST = 3;

    private final static Map< String, Object > ROUNDTRIP_VALS;

    private 
    static 
    void 
    codef( String tmpl, 
           Object... args ) 
    { 
        CodeLoggers.codef( tmpl, args ); 
    }

    private
    abstract
    class CoreIoTest
    extends LabeledTestCall
    {
        final String key;

        byte[] buffer;

        private 
        CoreIoTest( CharSequence prefix,
                    CharSequence lbl ) 
        { 
            super( prefix + "/" + lbl ); 
            this.key = lbl.toString();
        }

        final
        MingleBinReader
        createReader()
        {
            return MingleBinReader.
                create( new ByteArrayInputStream( this.buffer ) );
        }

        final
        Object
        readValue( MingleBinReader rd,
                   Object rep )
            throws Exception
        {
            if ( rep instanceof MingleValue ) return rd.readValue();
            throw state.failf( "unhandled read type: %s", rep.getClass() );
        }

        final
        void
        expectValue( MingleBinReader rd,
                     Object expct )
            throws Exception
        {
            Object act = readValue( rd, expct );
            state.equal( expct.getClass(), act.getClass() );
            state.equal( expct, act );
        }

        public
        void
        call()
            throws Exception
        {
            state.fail( "unimplemented" );
        }
    }

    private
    final
    class RoundtripTest
    extends CoreIoTest
    {
        private RoundtripTest( CharSequence lbl ) { super( "roundtrip", lbl ); }

        public
        void
        call()
            throws Exception
        {
            code( "buffer (%d): %s", this.buffer.length,
                IoUtils.asHexString( this.buffer ) );

            Object expct =
                state.get( ROUNDTRIP_VALS, this.key, "ROUNDTRIP_VALS" );
 
            MingleBinReader mgRd = createReader();
            expectValue( mgRd, expct );
        }
    }

    private
    final
    class SequenceRoundtripTest
    extends CoreIoTest
    {
        private 
        SequenceRoundtripTest( CharSequence lbl ) 
        { 
            super( "sequence-roundtrip", lbl ); 
        }
    }

    private
    final
    class InvalidDataTest
    extends CoreIoTest
    {
        private CharSequence message;

        private
        InvalidDataTest( CharSequence lbl )
        {
            super( "invalid-data", lbl );
        }
    }

    private
    final
    class TestReader
    extends IoTestFactory.LeTestReader< CoreIoTest >
    {
        private TestReader() { super( "core-io-tests.bin" ); }

        protected
        void
        readHeader()
            throws Exception
        {
            expectInt32( 1 );
        }

        private
        < T extends CoreIoTest >
        T
        readBuffer( T t )
            throws Exception
        {
            t.buffer = leRd().readByteArray();
            return t;
        }

        private
        RoundtripTest
        readRoundtripTest()
            throws Exception
        {
            return readBuffer( new RoundtripTest( leRd().readUtf8() ) );
        }

        private
        SequenceRoundtripTest
        readSequenceRoundtripTest()
            throws Exception
        {
            return readBuffer( new SequenceRoundtripTest( leRd().readUtf8() ) );
        }

        private
        InvalidDataTest
        readInvalidDataTest()
            throws Exception
        {
            InvalidDataTest res = new InvalidDataTest( leRd().readUtf8() );
            res.message = leRd().readUtf8();
            return readBuffer( res );
        }

        protected
        CoreIoTest
        readNext()
            throws Exception
        {
            byte tc = leRd().readByte();
            
            switch ( tc ) {
            case TC_ROUNDTRIP_TEST: return readRoundtripTest();
            case TC_SEQUENCE_ROUNDTRIP_TEST: return readSequenceRoundtripTest();
            case TC_INVALID_DATA_TEST: return readInvalidDataTest();
            case TC_END: return end();
            }

            throw failf( "unhandled tc: 0x%02x", tc );
        }
    }

    @InvocationFactory
    private
    List< ? >
    getTests()
        throws Exception
    {
        return new TestReader().call();
    }

    private
    static
    Map< String, Object >
    createRoundtripVals()
    {
        Map< String, Object > res = Lang.newMap( String.class, Object.class,
//    b.setVal( "null-val", NullVal )
//    b.setVal( "string-empty", String( "" ) )
//    b.setVal( "string-val1", String( "hello" ) )
//    b.setVal( "bool-true", Boolean( true ) )
//    b.setVal( "bool-false", Boolean( false ) )
//    b.setVal( "buffer-empty", Buffer( []byte{} ) )
//    b.setVal( "buffer-nonempty", Buffer( []byte( "hello" ) ) )
//    b.setVal( "int32-min", Int32( math.MaxInt32 ) )
//    b.setVal( "int32-max", Int32( math.MinInt32 ) )
//    b.setVal( "int32-pos1", Int32( int32( 1 ) ) )
//    b.setVal( "int32-zero", Int32( int32( 0 ) ) )
//    b.setVal( "int32-neg1", Int32( int32( -1 ) ) )
//    b.setVal( "int64-min", Int64( math.MaxInt64 ) )
//    b.setVal( "int64-max", Int64( math.MinInt64 ) )
//    b.setVal( "int64-pos1", Int64( int64( 1 ) ) )
//    b.setVal( "int64-zero", Int64( int64( 0 ) ) )
//    b.setVal( "int64-neg1", Int64( int64( -1 ) ) )
//    b.setVal( "uint32-min", Uint32( math.MaxUint32 ) )
//    b.setVal( "uint32-max", Uint32( uint32( 0 ) ) )
//    b.setVal( "uint32-pos1", Uint32( uint32( 1 ) ) )
//    b.setVal( "uint32-zero", Uint32( uint32( 0 ) ) )
            "uint64-min", new MingleUint64( 0L ),
            "uint64-max", new MingleUint64( 0xFFFFFFFFFFFFFFFFL ),
            "uint64-pos1", new MingleUint64( 1L )
//    b.setVal( "uint64-min", Uint64( math.MaxUint64 ) )
//    b.setVal( "uint64-max", Uint64( uint64( 0 ) ) )
//    b.setVal( "uint64-pos1", Uint64( uint64( 1 ) ) )
//    b.setVal( "uint64-zero", Uint64( uint64( 0 ) ) )
//    b.setVal( "float32-val1", Float32( float32( 1 ) ) )
//    b.setVal( "float32-max", Float32( math.MaxFloat32 ) )
//    b.setVal( "float32-smallest-nonzero",
//        Float32( math.SmallestNonzeroFloat32 ) )
//    b.setVal( "float64-val1", Float64( float64( 1 ) ) )
//    b.setVal( "float64-max", Float64( math.MaxFloat64 ) )
//    b.setVal( "float64-smallest-nonzero",
//        Float64( math.SmallestNonzeroFloat64 ) )
//    b.setVal( "time-val1", MustTimestamp( "2013-10-19T02:47:00-08:00" ) )
//    b.setVal( "enum-val1", MustEnum( "ns1@v1/E1", "val1" ) )
//    b.setVal( "symmap-empty", MustSymbolMap() )
//
//    b.setVal( "symmap-flat", 
//        MustSymbolMap( "k1", int32( 1 ), "k2", int32( 2 ) ) )
//
//    b.setVal( "symmap-nested",
//        MustSymbolMap( "k1", MustSymbolMap( "kk1", int32( 1 ) ) ) )
//
//    b.setVal( "struct-empty", MustStruct( "ns1@v1/T1" ) )
//    b.setVal( "struct-flat", MustStruct( "ns1@v1/T1", "k1", int32( 1 ) ) )
//    b.setVal( "list-empty", MustList() )
//    b.setVal( "list-scalars", MustList( int32( 1 ), "hello" ) )
//
//    b.setVal( "list-nested",
//        MustList( int32( 1 ), MustList(), MustList( "hello" ), NullVal ) )
//
//    b.setVal( "id1", id( "id1" ) )
//    b.setVal( "id1-id2", id( "id1-id2" ) )
        );

        return res;
    }

    static
    {
        ROUNDTRIP_VALS = createRoundtripVals();
    }
}
