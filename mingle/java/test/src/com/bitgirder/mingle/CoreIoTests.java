package com.bitgirder.mingle;

import static com.bitgirder.mingle.MingleTestMethods.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPaths;

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

    private final static Map< String, Object > ROUNDTRIP_VALS;

    private PipedProcess checker;

    @Before
    private
    void
    startChecker()
        throws Exception
    {
        File cmd = TestData.expectFile( "check-core-io" );
        ProcessBuilder pb = new ProcessBuilder( cmd.getAbsolutePath() );
        checker = PipedProcess.start( pb );
    }

    @After private void stopChecker() throws Exception { checker.kill(); }

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

        // lazily instantiated via getWriter()
        private ByteArrayOutputStream bos;
        private MingleBinWriter mgWr;

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
            if ( rep instanceof ObjectPath ) return rd.readIdPath();

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

        private
        void
        assertIdPaths( Object expct,
                       Object act )
        {
            ObjectPath< MingleIdentifier > p1 = Lang.castUnchecked( expct );
            ObjectPath< MingleIdentifier > p2 = Lang.castUnchecked( act );

            state.isTrue( ObjectPaths.areEqual( p1, p2 ) );
        }

        final
        Object
        expectValue( MingleBinReader rd,
                     Object expct )
            throws Exception
        {
            Object act = readValue( rd, expct );
            state.equal( expct.getClass(), act.getClass() );

            if ( expct instanceof ObjectPath ) {
                assertIdPaths( expct, act );
            } else { 
                state.equal( expct, act ); 
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

            if ( val instanceof MingleValue ) {
                mgWr.writeValue( (MingleValue) val );
            } else if ( val instanceof MingleIdentifier ) {
                mgWr.writeIdentifier( (MingleIdentifier) val );
            } else if ( val instanceof MingleNamespace ) {
                mgWr.writeNamespace( (MingleNamespace) val );
            } else if ( val instanceof DeclaredTypeName ) {
                mgWr.writeDeclaredTypeName( (DeclaredTypeName) val );
            } else if ( val instanceof QualifiedTypeName ) {
                mgWr.writeQualifiedTypeName( (QualifiedTypeName) val );
            } else if ( val instanceof ObjectPath ) {
                mgWr.writeIdentifierPath( Mingle.castIdPath( val ) );
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
            Object expct =
                state.get( ROUNDTRIP_VALS, this.key, "ROUNDTRIP_VALS" );
 
            MingleBinReader mgRd = createReader();
            Object act = expectValue( mgRd, expct );

            writeTestValue( act );

            codef( "buffer: %s, wrote: %s",
                IoUtils.asHexString( buffer ),
                IoUtils.asHexString( writeBuffer() ) );

//            IoTests.assertEqual( buffer, writeBuffer() );
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
    void
    putValueRoundtrips( Map< String, Object > m )
    {
        m.putAll( Lang.newMap( String.class, Object.class,
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

            "list-empty", MingleList.empty(),

            "list-scalars", 
                MingleList.asList(
                    new MingleInt32( 1 ), new MingleString( "hello" ) ),

            "list-nested",
                MingleList.asList(
                    new MingleInt32( 1 ),
                    MingleList.empty(),
                    MingleList.asList( new MingleString( "hello" ) ),
                    MingleNull.getInstance()
                )
        ));
    }

    private
    static
    void
    putPathRoundtrips( Map< String, Object > m )
    {
        ObjectPath< MingleIdentifier > p1 = 
            Lang.putUnique( m, "p1", idPathRoot( "id1" ) );
        
        ObjectPath< MingleIdentifier > p2 = 
            Lang.putUnique( m, "p2", p1.descend( id( "id2" ) ) );

        ObjectPath< MingleIdentifier > p3 =
            Lang.putUnique( m, "p3", p2.startImmutableList().next().next() );
        
        ObjectPath< MingleIdentifier > p4 =
            Lang.putUnique( m, "p4", p3.descend( id( "id3" ) ) );

        Lang.putUnique( m, "p5", 
            ObjectPath.getRoot().startImmutableList().descend( id( "id1" ) ) );
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
    putDefinitionRoundtripVals( Map< String, Object > m,
                                Object... vals )
    {
        for ( Object val : vals )
        {
            String key = keyForType( val );
            Lang.putUnique( m, key + "/" + val.toString(), val );
        }
    }

    private
    static
    void
    putDefinitionRoundtrips( Map< String, Object > m )
    {
        putDefinitionRoundtripVals( m,

            MingleIdentifier.create( "id1" ),
            MingleIdentifier.create( "id1-id2" ),

            MingleNamespace.create( "ns1@v1" ),
            MingleNamespace.create( "ns1:ns2@v1" ),

            DeclaredTypeName.create( "T1" ),

            QualifiedTypeName.create( "ns1:ns2@v1/T1" ),

            MingleTypeReference.create( "T1" ),
            MingleTypeReference.create( "String~\"a\"" ),
            MingleTypeReference.create( "String~[\"a\",\"b\"]" ),
            MingleTypeReference.create( "Int32~(0,10)" ),
            MingleTypeReference.create( "Int64~[0,10]" ),
            MingleTypeReference.create( "Uint32~(0,10)" ),
            MingleTypeReference.create( "Uint64~[0,10]" ),
            MingleTypeReference.create( "Float64~(,)" ),
            MingleTypeReference.create( "T1*" ),
            MingleTypeReference.create( "T1+" ),
            MingleTypeReference.create( "T1*?" ),
            MingleTypeReference.create( "ns1@v1/T1" ),
            MingleTypeReference.create( "ns1@v1/T1*" ),
            MingleTypeReference.create( "ns1@v1/T1?" )
        );

        // due to differences in how we serialize some values ("0" vs "0.0"), we
        // handcode the key for a few test values
        
        Lang.putUnique( m,
            "AtomicTypeReference/mingle:core@v1/Timestamp~" +
                "[\"2012-01-01T00:00:00Z\",\"2012-02-01T00:00:00Z\"]",
            MingleTypeReference.create(
                "Timestamp~[\"2012-01-01T00:00:00Z\",\"2012-02-01T00:00:00Z\"]"
            )
        );

        Lang.putUnique( m,
            "AtomicTypeReference/mingle:core@v1/Float64~[0,1)",
            MingleTypeReference.create( "Float64~[0.0,1.0)" )
        );

        Lang.putUnique( m,
            "AtomicTypeReference/mingle:core@v1/Float32~(0,1]",
            MingleTypeReference.create( "Float32~(0.0,1.0]" )
        );
    }

    private
    static
    Map< String, Object >
    createRoundtripVals()
    {
        Map< String, Object > res = Lang.newMap();

        putValueRoundtrips( res );
        putDefinitionRoundtrips( res );
        putPathRoundtrips( res );

        return res;
    }

    static
    {
        ROUNDTRIP_VALS = createRoundtripVals();
    }
}
