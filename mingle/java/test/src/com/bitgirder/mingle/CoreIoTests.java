package com.bitgirder.mingle;

import static com.bitgirder.mingle.MingleTestMethods.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.test.Test;
import com.bitgirder.test.Before;
import com.bitgirder.test.After;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.LabeledTestCall;

import com.bitgirder.testing.TestData;

import com.bitgirder.io.IoTestFactory;
import com.bitgirder.io.IoUtils;
import com.bitgirder.io.IoTests;
import com.bitgirder.io.PipedProcess;
import com.bitgirder.io.BinReader;
import com.bitgirder.io.BinWriter;

import java.io.InputStream;
import java.io.OutputStream;
import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.File;

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

    private final static byte RESP_CODE_PASSED = 0;
    private final static byte RESP_CODE_FAILED = 1;

    private final static Map< String, Object > TEST_VALS = Lang.newMap();

    private PipedProcess checker;
    private final Object checkerSync = new Object();

    // must be called while holding lock for checkerSync
    private
    void
    startChecker()
        throws Exception
    {
        state.isTrue( checker == null, "checker already started" );

        String cmd = 
            TestData.expectCommand( "check-core-io" ).getAbsolutePath();

        checker = PipedProcess.start( cmd );
    }

    @After 
    private 
    void 
    stopChecker() 
        throws Exception 
    { 
        if ( checker != null ) checker.kill(); 
    }

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
        byte[] buffer;

        // lazily instantiated via getWriter()
        private ByteArrayOutputStream bos;
        private MingleBinWriter mgWr;

        private CoreIoTest( CharSequence lbl ) { super( lbl ); }

        final
        Object
        valueExpected()
        {
            return state.get( TEST_VALS, getLabel().toString(), "TEST_VALS" );
        }

        final
        MingleBinReader
        createReader()
        {
            return MingleBinReader.
                create( new ByteArrayInputStream( this.buffer ) );
        }

        final
        MingleValue
        readMingleValue( MingleBinReader rd )
            throws Exception
        {
            return MingleTestMethods.readValue( rd );
        }

        final
        Object
        readValue( MingleBinReader rd,
                   Object rep )
            throws Exception
        {
            if ( rep instanceof MingleValue ) return readMingleValue( rd );

            if ( rep instanceof QualifiedTypeName ) {
                return rd.readQualifiedTypeName();
            }

            if ( rep instanceof MingleNamespace ) return rd.readNamespace();
            if ( rep instanceof MingleIdentifier ) return rd.readIdentifier();

            if ( rep instanceof DeclaredTypeName ) {
                return rd.readDeclaredTypeName();
            }

            if ( rep instanceof MingleTypeReference ) {
                return rd.readTypeReference();
            }

            throw state.failf( "unhandled read type: %s", rep.getClass() );
        }

        final
        Object
        expectValue( MingleBinReader rd,
                     Object expct )
            throws Exception
        {
            Object act = readValue( rd, expct );
            state.equal( expct.getClass(), act.getClass() );

            if ( ! expct.equals( act ) ) 
            {
                Object expctDesc = expct;
                Object actDesc = act;

                if ( expct instanceof MingleValue ) {
                    expctDesc = Mingle.inspect( (MingleValue) expct );
                    actDesc = Mingle.inspect( (MingleValue) act );
                }
            
                state.failf( "expected %s, got %s", expctDesc, actDesc );
            }

            return act;
        }

        final
        MingleBinWriter
        getWriter()
        {
            if ( mgWr == null )
            {
                bos = new ByteArrayOutputStream();
                mgWr = MingleBinWriter.create( bos );
            }

            return mgWr;
        }

        final
        void
        writeTestValue( Object val )
            throws Exception
        {
            MingleBinWriter mgWr = getWriter();

            if ( val instanceof MingleValue ) 
            {
                MingleValueReactors.visitValue( 
                    (MingleValue) val, mgWr.asReactor() );
            } 
            else if ( val instanceof MingleIdentifier ) {
                mgWr.writeIdentifier( (MingleIdentifier) val );
            } else if ( val instanceof MingleNamespace ) {
                mgWr.writeNamespace( (MingleNamespace) val );
            } else if ( val instanceof DeclaredTypeName ) {
                mgWr.writeDeclaredTypeName( (DeclaredTypeName) val );
            } else if ( val instanceof QualifiedTypeName ) {
                mgWr.writeQualifiedTypeName( (QualifiedTypeName) val );
            } else if ( val instanceof MingleTypeReference ) {
                mgWr.writeTypeReference( (MingleTypeReference) val );
            } else {
                state.failf( "unhandled write val: %s", val.getClass() );
            }
        }

        final
        byte[]
        writeBuffer()
        {
            state.notNull( mgWr, "no writer set" );
            return bos.toByteArray();
        }

        private
        void
        readCheckRes( BinReader br )
            throws Exception
        {
            int rc = br.readByte();

            switch ( rc ) {
            case RESP_CODE_PASSED: break;
            case RESP_CODE_FAILED:
                state.failf( 
                    "check failed with remote message: %s", br.readUtf8() );
            default: state.failf( "unhandled response code: 0x%02x", rc );
            }
        }

        private
        void
        sendCheckValue( BinWriter bw )
            throws Exception
        {
            bw.writeUtf8( getLabel() );
            bw.writeByteArray( writeBuffer() );
        }

        final
        void
        checkWriteValue()
            throws Exception
        {
            synchronized ( checkerSync )
            {
                if ( checker == null ) startChecker();

                checker.usePipe(
                    new ObjectReceiver< InputStream >() {
                        public void receive( InputStream is ) throws Exception {
                            readCheckRes( BinReader.asReaderLe( is ) );
                        }
                    },
                    new ObjectReceiver< OutputStream >() {
                        public void receive( OutputStream os ) 
                            throws Exception 
                        {
                            sendCheckValue( BinWriter.asWriterLe( os ) );
                        }
                    }
                );
            }
        }
    }
    
    private
    final
    class RoundtripTest
    extends CoreIoTest
    {
        private RoundtripTest( CharSequence lbl ) { super( lbl ); }

        public
        void
        call()
            throws Exception
        {
            Object expct = valueExpected();
 
            MingleBinReader mgRd = createReader();
            Object act = expectValue( mgRd, expct );

            writeTestValue( act );
            checkWriteValue();
        }
    }

    private
    final
    class SequenceRoundtripTest
    extends CoreIoTest
    {
        private SequenceRoundtripTest( CharSequence lbl ) { super( lbl ); }

        public
        void
        call()
            throws Exception
        {
            MingleList seq = (MingleList) valueExpected();

            MingleBinReader rd = createReader();

            for ( MingleValue expct : seq ) {
                MingleValue act = (MingleValue) expectValue( rd, expct );
                writeTestValue( act );
            }

            checkWriteValue();
        }
    }

    private
    final
    class InvalidDataTest
    extends CoreIoTest
    {
        private CharSequence message;

        private InvalidDataTest( CharSequence lbl ) { super( lbl ); }

        private
        void
        assertBinaryException( MingleBinaryException mbe )
        {
            CharSequence expct = (String) TEST_VALS.get( getLabel() );
            if ( expct == null ) expct = message;

            state.equalString( expct, mbe.getMessage() );
        }

        public
        void
        call()
            throws Exception
        {
            MingleBinReader rd = createReader();

            try {
                MingleValue val = readMingleValue( rd );
                state.failf( "was able to read value: %s", val );
            } catch ( MingleBinaryException mbe ) { 
                assertBinaryException( mbe ); 
            }
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
    testCoreIo()
        throws Exception
    {
        return new TestReader().call();
    }

    private
    static
    < V >
    V
    putTestValue( String key,
                  V val )
    {
        Lang.putUnique( TEST_VALS, key, val );
        return val;
    }

    private
    static
    void
    putTestValues( String prefix,
                   Map< String, Object > vals )
    {
        for ( Map.Entry< String, Object > e : vals.entrySet() )
        {
            putTestValue( prefix + "/" + e.getKey(), e.getValue() );
        }
    }

    private
    static
    < V >
    V
    putRoundtripValue( String key,
                       V val )
    {
        return putTestValue( "roundtrip/" + key, val );
    }

    private
    static
    void
    putValueRoundtrips()
    {
        Map< String, Object > vals = Lang.newMap( String.class, Object.class,
            "null-val", MingleNull.getInstance(),
            "string-empty", new MingleString( "" ),
            "string-val1", new MingleString( "hello" ),
            "bool-true", MingleBoolean.TRUE,
            "bool-false", MingleBoolean.FALSE,
            "buffer-empty", new MingleBuffer( new byte[] {} ),
            "buffer-nonempty", new MingleBuffer( new byte[] { 0x00, 0x01 } ),
            "int32-min", new MingleInt32( Integer.MIN_VALUE ),
            "int32-max", new MingleInt32( Integer.MAX_VALUE ),
            "int32-pos1", new MingleInt32( 1 ),
            "int32-zero", new MingleInt32( 0 ),
            "int32-neg1", new MingleInt32( -1 ),
            "int64-min", new MingleInt64( Long.MIN_VALUE ),
            "int64-max", new MingleInt64( Long.MAX_VALUE ),
            "int64-zero", new MingleInt64( 0L ),
            "int64-pos1", new MingleInt64( 1L ),
            "int64-neg1", new MingleInt64( -1L ),
            "uint32-min", new MingleUint32( 0 ),
            "uint32-max", new MingleUint32( 0xFFFFFFFF ),
            "uint32-pos1", new MingleUint32( 1 ),
            "uint64-min", new MingleUint64( 0L ),
            "uint64-max", new MingleUint64( 0xFFFFFFFFFFFFFFFFL ),
            "uint64-pos1", new MingleUint64( 1L ),
            "float32-val1", new MingleFloat32( 1.0f ),
            "float32-max", new MingleFloat32( Float.MAX_VALUE ),
            "float32-smallest-nonzero", new MingleFloat32( Float.MIN_VALUE ),
            "float64-val1", new MingleFloat64( 1.0d ),
            "float64-max", new MingleFloat64( Double.MAX_VALUE ),
            "float64-smallest-nonzero", new MingleFloat64( Double.MIN_VALUE ),
            "time-val1", MingleTimestamp.create( "2013-10-19T02:47:00-08:00" ),

            "enum-val1", 
                MingleEnum.create( qname( "ns1@v1/E1" ), id( "val1" ) ),

            "symmap-empty", MingleSymbolMap.empty(),

            "symmap-flat",
                new MingleSymbolMap.Builder().
                    setInt32( "k1", 1 ).
                    setInt32( "k2", 2 ).
                    setInt32( "k3", 1 ).
                    build(),
            
            "symmap-nested",
                new MingleSymbolMap.Builder().
                    set( "k1",
                        new MingleSymbolMap.Builder().
                            setInt32( "kk1", 1 ).
                            build()
                    ).
                    build(),

            "struct-empty",
                new MingleStruct.Builder().setType( "ns1@v1/T1" ).build(),
                
            "struct-flat",
                new MingleStruct.Builder().
                    setType( "ns1@v1/T1" ).
                    setInt32( "k1", 1 ).
                    build(),

            "list-empty", emptyList(),

            "list-scalars", 
                MingleList.asList(
                    Mingle.TYPE_OPAQUE_LIST,
                    new MingleInt32( 1 ), new MingleString( "hello" ) ),

            "list-nested",
                MingleList.asList(
                    Mingle.TYPE_OPAQUE_LIST,
                    new MingleInt32( 1 ),
                    emptyList(),
                    MingleList.asList( 
                        Mingle.TYPE_OPAQUE_LIST, new MingleString( "hello" ) ),
                    MingleNull.getInstance()
                ),

            "list-typed",
                MingleList.asList(
                    listType( Mingle.TYPE_INT32, true ),
                    new MingleInt32( 0 ), new MingleInt32( 1 )
                )
        );

        for ( Map.Entry< String, Object > e : vals.entrySet() ) {
            putRoundtripValue( e.getKey(), e.getValue() );
        }
    }

    private
    static
    String
    keyForType( Object val )
    {
        String res = val.getClass().getSimpleName();

        if ( res.startsWith( "Mingle" ) ) return res.substring( 6 );

        return res;
    }

    private
    static
    void
    putDefinitionRoundtripVals( Object... vals )
    {
        for ( Object val : vals )
        {
            String key = keyForType( val );
            putRoundtripValue( key + "/" + val.toString(), val );
        }
    }

    private
    static
    void
    putDefinitionRoundtrips()
    {
        putDefinitionRoundtripVals(

            MingleIdentifier.create( "id1" ),
            MingleIdentifier.create( "id1-id2" ),

            MingleNamespace.create( "ns1@v1" ),
            MingleNamespace.create( "ns1:ns2@v1" ),

            DeclaredTypeName.create( "T1" ),

            QualifiedTypeName.create( "ns1:ns2@v1/T1" ),

            atomic( qname( "mingle:core@v1/String" ),
                MingleRegexRestriction.create( "a" ) ),

            atomic( qname( "mingle:core@v1/String" ),
                MingleRangeRestriction.create(
                    true,
                    new MingleString( "a" ),
                    new MingleString( "b" ),
                    true,
                    MingleString.class
                )
            ),
            
            atomic( qname( "mingle:core@v1/Int32" ),
                MingleRangeRestriction.createChecked(
                    false,
                    new MingleInt32( 0 ),
                    new MingleInt32( 10 ),
                    false,
                    MingleInt32.class
                )
            ),
            
            atomic( qname( "mingle:core@v1/Int64" ),
                MingleRangeRestriction.createChecked(
                    true,
                    new MingleInt64( 0L ),
                    new MingleInt64( 10L ),
                    true,
                    MingleInt64.class
                )
            ),
            
            atomic( qname( "mingle:core@v1/Uint32" ),
                MingleRangeRestriction.createChecked(
                    false,
                    new MingleUint32( 0 ),
                    new MingleUint32( 10 ),
                    false,
                    MingleUint32.class
                )
            ),
            
            atomic( qname( "mingle:core@v1/Uint64" ),
                MingleRangeRestriction.createChecked(
                    true,
                    new MingleUint64( 0L ),
                    new MingleUint64( 10L ),
                    true,
                    MingleUint64.class
                )
            ),
            
            atomic( qname( "mingle:core@v1/Float64" ),
                MingleRangeRestriction.createChecked(
                    false, null, null, false, MingleFloat64.class )
            ),
 
            atomic( qname( "ns1@v1/T1" ) ),

            listType( atomic( qname( "ns1@v1/T1" ) ), true ),
            
            listType( atomic( qname( "ns1@v1/T1" ) ), false ),

            listType( 
                nullableType(
                    nullableType(
                        listType(
                            nullableType( atomic( qname( "ns1@v1/T1" ) ) ),
                            false
                        )
                    )
                ),
                true
            ),

            nullableType( listType( atomic( qname( "ns1@v1/T1" ) ), true ) ),
 
            nullableType( atomic( qname( "ns1@v1/T1" ) ) ),

            ptrType( atomic( qname( "ns1@v1/T1" ) ) ),

            nullableType( ptrType( atomic( qname( "ns1@v1/T1" ) ) ) ),

            listType(
                nullableType(
                    listType(
                        nullableType(
                            ptrType( atomic( qname( "ns1@v1/T1" ) ) )
                        ),
                        false
                    )
                ),
                true
            )
        );

        // due to differences in how we serialize some values ("0" vs "0.0"), we
        // handcode the key for a few test values
 
        putRoundtripValue( 
            "AtomicTypeReference/mingle:core@v1/Timestamp~" +
                "[\"2012-01-01T00:00:00Z\",\"2012-02-01T00:00:00Z\"]",
            atomic( qname( "mingle:core@v1/Timestamp" ),
                MingleRangeRestriction.create(
                    true,
                    MingleTimestamp.create( "2012-01-01T00:00:00Z" ),
                    MingleTimestamp.create( "2012-02-01T00:00:00Z" ),
                    true,
                    MingleTimestamp.class
                )
            )
        );

        putRoundtripValue(
            "AtomicTypeReference/mingle:core@v1/Float64~[0,1)",
            atomic( qname( "mingle:core@v1/Float64" ),
                MingleRangeRestriction.create(
                    true,
                    new MingleFloat64( 0.0d ),
                    new MingleFloat64( 1.0d ),
                    false,
                    MingleFloat64.class
                )
            )
        );

        putRoundtripValue(
            "AtomicTypeReference/mingle:core@v1/Float32~(0,1]",
            atomic( qname( "mingle:core@v1/Float32" ),
                MingleRangeRestriction.create(
                    false,
                    new MingleFloat32( 0.0f ),
                    new MingleFloat32( 1.0f ),
                    true,
                    MingleFloat32.class
                )
            )
        );

        putRoundtripValue(
            "AtomicTypeReference/mingle:core@v1/Uint64~[0,0]",
            atomic( qname( "mingle:core@v1/Uint64" ),
                MingleRangeRestriction.create(
                    true,
                    new MingleUint64( 0 ),
                    new MingleUint64( 0 ),
                    true,
                    MingleUint64.class
                )
            )
        );
    }

    private
    static
    void
    putRoundtripVals()
    {
        putValueRoundtrips();
        putDefinitionRoundtrips();
    }

    private
    static
    void
    putSequenceRoundtripValue( String key,
                               Object val )
    {
        putTestValue( "sequence-roundtrip/" + key, val );
    }

    private
    static
    void
    putSequenceRoundtripVals()
    {
        putTestValues( "sequence-roundtrip", 
            Lang.newMap( String.class, Object.class,
                "struct-sequence",
                    MingleList.asList(
                        Mingle.TYPE_OPAQUE_LIST,
                        new MingleStruct.Builder().
                            setType( "ns1@v1/S1" ).
                            build(),
                        new MingleStruct.Builder().
                            setType( "ns1@v1/S1" ).
                            setInt32( "f1", 1 ).
                            build()
                    )
            )
        );
    }

    private
    static
    void
    putInvalidDataMessages()
    {
        putTestValues( "invalid-data", Lang.newMap( String.class, Object.class,
            "unexpected-top-level-type-code",
                "[offset 0]: Expected mingle value but saw type code 0x64",
            "unexpected-list-val-type-code",
                "[offset 96]: Expected mingle value but saw type code 0x64",
            "unexpected-symmap-val-type-code",
                "[offset 39]: Expected mingle value but saw type code 0x64",
            "invalid-list-type",
                "[offset 1]: Expected list type reference but saw type code 0x05",
            "invalid-declared-type-name",
                "[offset 25]: (at or near char 2) Illegal type name rune: \"$\" (U+0024)"
        ));
    }

    static
    {
        putRoundtripVals();
        putSequenceRoundtripVals();
        putInvalidDataMessages();
    }
}
