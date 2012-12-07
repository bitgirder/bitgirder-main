package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.TypedString;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestCall;
import com.bitgirder.test.LabeledTestObject;
import com.bitgirder.test.InvocationFactory;

import com.bitgirder.testing.TestData;

import com.bitgirder.io.BinReader;

import java.util.List;
import java.util.Iterator;

import java.io.InputStream;

@Test
final
class CastValueTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static int FILE_VERSION = 0x01;

    private final static byte TC_NIL = (byte) 0x00;
    private final static byte TC_TEST = (byte) 0x01;
    private final static byte TC_IN = (byte) 0x02;
    private final static byte TC_EXPECT = (byte) 0x03;
    private final static byte TC_TYPE = (byte) 0x04;
    private final static byte TC_PATH = (byte) 0x05;
    private final static byte TC_ERR = (byte) 0x06;
    private final static byte TC_VC_ERR = (byte) 0x07;
    private final static byte TC_MSG = (byte) 0x08;
    private final static byte TC_TC_ERR = (byte) 0x09;
    private final static byte TC_EXPECTED = (byte) 0x0a;
    private final static byte TC_ACTUAL = (byte) 0x0b;

    private final static MingleTypeReference[] UINT_TYPES =
        new MingleTypeReference[] { Mingle.TYPE_UINT32, Mingle.TYPE_UINT64 };

    private final List< Expectation > expectations = Lang.newList();

    private
    static
    boolean
    equalMaps( MingleSymbolMap m1,
               MingleValue mv2 )
    {
        if ( ! ( mv2 instanceof MingleSymbolMap ) ) return false;
            
        MingleSymbolMap m2 = (MingleSymbolMap) mv2;

        if ( ! m1.getKeySet().equals( m2.getKeySet() ) ) return false;
        
        for ( MingleIdentifier fld : m1.getFields() )
        {
            if ( ! equalVals( m1.get( fld ), m2.get( fld ) ) )
            {
                return false;
            }
        }

        return true;
    }

    private
    static
    boolean
    equalStructs( MingleStruct s1,
                  MingleValue mv2 )
    {
        if ( ! ( mv2 instanceof MingleStruct ) ) return false;

        MingleStruct s2 = (MingleStruct) mv2;

        return s1.getType().equals( s2.getType() ) &&
               equalMaps( s1.getFields(), s2.getFields() );
    }

    private
    static
    boolean
    equalLists( MingleList l1,
                MingleValue mv2 )
    {
        if ( ! ( mv2 instanceof MingleList ) ) return false;

        MingleList l2 = (MingleList) mv2;

        Iterator< MingleValue > i1 = l1.iterator();
        Iterator< MingleValue > i2 = l2.iterator();

        while ( i1.hasNext() )
        {
            if ( ! i2.hasNext() ) return false;
            if ( ! equalVals( i1.next(), i2.next() ) ) return false;
        }

        return ! i2.hasNext();
    }

    private
    static
    boolean
    equalVals( MingleValue mv1,
               MingleValue mv2 )
    {
        state.notNull( mv1, "mv1" );
        state.notNull( mv2, "mv2" );

        if ( mv1 instanceof MingleStruct )
        {
            return equalStructs( (MingleStruct) mv1, mv2 );
        }
        else if ( mv1 instanceof MingleSymbolMap )
        {
            return equalMaps( (MingleSymbolMap) mv1, mv2 );
        }
        else if ( mv1 instanceof MingleList )
        {
            return equalLists( (MingleList) mv1, mv2 );
        }
        else return mv1.equals( mv2 );
    }

    private
    final
    static
    class Expectation
    {
        private final MingleValue in;
        private final MingleTypeReference type;
        private final Object obj;

        private
        Expectation( MingleValue in,
                     MingleTypeReference type,
                     Object obj )
        {
            this.in = in;
            this.type = type;
            this.obj = obj;
        }
    }

    private
    final
    static
    class ErrMsg
    extends TypedString< ErrMsg >
    {
        private ErrMsg( CharSequence s ) { super( s ); }
    }

    private
    void
    addExpectation( MingleValue in,
                    MingleTypeReference typ,
                    Object obj )
    {
        expectations.add( new Expectation( in, typ, obj ) );
    }

    private
    void
    addExpectations()
        throws Exception
    {
        MingleString onePointO = new MingleString( "1.0" );

        addExpectation( 
            new MingleFloat32( 1.0f ), Mingle.TYPE_STRING, onePointO );

        addExpectation(
            new MingleFloat64( 1.0d ), Mingle.TYPE_STRING, onePointO );

        addExpectation(
            new MingleString( "18446744073709551616" ),
            Mingle.TYPE_UINT64,
            new ErrMsg( "Number is too large for uint64: 18446744073709551616" )
        );

        addExpectation(
            new MingleString( "4294967296" ),
            Mingle.TYPE_UINT32,
            new ErrMsg( "Number is too large for uint32: 4294967296" )
        );

        for ( MingleTypeReference typ : UINT_TYPES )
        {
            addExpectation(
                new MingleString( "not-a-num" ),
                typ,
                new ErrMsg( "Illegal embedded minus sign" )
            );

            addExpectation(
                new MingleString( "-1" ),
                typ,
                new ErrMsg( "Number is negative: -1" )
            );
        }

        addExpectation(
            new MingleString( "abc$/@" ),
            Mingle.TYPE_BUFFER,
            new ErrMsg( "Length of input 'abc$/@' (6) is not a multiple of 4" )
        );

        addExpectation(
            MingleTimestamp.parse( "2007-08-24T21:15:43.123450000Z" ),
            Mingle.TYPE_STRING,
            new MingleString( "2007-08-24T21:15:43.123450000Z" )
        );

        addExpectation(
            MingleTimestamp.parse( "2012-01-01T00:00:00.000000000Z" ),
            new AtomicTypeReference(
                Mingle.QNAME_TIMESTAMP,
                MingleRangeRestriction.createChecked(
                    true,
                    MingleTimestamp.parse( "2000-01-01T00:00:00.000000000Z" ),
                    MingleTimestamp.parse( "2001-01-01T00:00:00.000000000Z" ),
                    true,
                    MingleTimestamp.class
                )
            ),
            new ErrMsg(
                "Value 2012-01-01T00:00:00.000000000Z does not satisfy " +
                "restriction [\"2000-01-01T00:00:00.000000000Z\"," +
                "\"2001-01-01T00:00:00.000000000Z\"]"
            )
        );        

        addExpectation(
            new MingleString( "s" ),
            Mingle.TYPE_BOOLEAN,
            new ErrMsg( "(at or near char 1) Invalid boolean string: s" )
        );
    }

    private
    final
    static
    class Error
    {
        private String msg;
        private ObjectPath< MingleIdentifier > loc;
        private MingleValueException mve;

        private
        Error
        complete( MingleValueException mve )
        {
            state.notNull( msg, "msg" );
            state.notNull( loc, "loc" );
            this.mve = state.notNull( mve, "mve" );

            return this;
        }
    }

    private
    final
    class TestImpl
    implements LabeledTestObject,
               TestCall
    {
        private MingleValue in;
        private MingleValue expct;
        private MingleTypeReference type;
        private ObjectPath< MingleIdentifier > path;
        private Error err;

        public Object getInvocationTarget() { return this; }

        public
        CharSequence
        getLabel()
        {
            return Strings.crossJoin( "=", ",",
                "in", Mingle.inspect( in ),
                "inCls", in.getClass().getSimpleName(),
                "type", type.getExternalForm()
            );
        }

        private
        < V >
        V
        expectation( Class< V > cls )
        {
            for ( Expectation e : expectations )
            {
                if ( e.type.equals( type ) && equalVals( in, e.in ) )
                {
                    if ( cls.isInstance( e.obj ) ) return cls.cast( e.obj );
                }
            }

            return null;
        }

        private
        void
        assertErrorBase( MingleValueException mve )
        {
            ErrMsg e = expectation( ErrMsg.class );
            CharSequence errExpct = e == null ? err.mve.error() : e;
            state.equalString( errExpct, mve.error() );

            state.equalString(
                Mingle.formatIdPath( err.mve.location() ),
                Mingle.formatIdPath( mve.location() )
            );
        }

        private
        void
        assertTcError( MingleTypeCastException tcEx )
        {
            if ( err.mve instanceof MingleTypeCastException )
            {
                MingleTypeCastException expct =
                    (MingleTypeCastException) err.mve;
                
                state.equal( expct.expected(), tcEx.expected() );
                state.equal( expct.actual(), tcEx.actual() );
            }
            else throw tcEx;
        }

        private
        void
        assertError( MingleValueException mve )
        {
            assertErrorBase( mve );

            if ( mve instanceof MingleValueCastException );
            else if ( mve instanceof MingleTypeCastException )
            {
                assertTcError( (MingleTypeCastException) mve );
            }
            else throw mve;
        }

        private
        MingleValue
        getExpectVal()
        {
            MingleValue e = expectation( MingleValue.class );
            return e == null ? expct : e;
        }

        public
        void
        call()
            throws Exception
        {
            try
            {
                MingleValue act = Mingle.castValue( in, type, path );

                state.isTrue( err == null, 
                    "Expected error, got", Mingle.inspect( act ) );
 
                MingleTests.assertEqual( getExpectVal(), act );
            }
            catch ( MingleValueException mve )
            {
                if ( err == null ) throw mve;
                assertError( mve );
            }
        }
    }

    private
    final
    class TestReader
    {
        private final BinReader br;
        private final MingleBinReader mgr;

        private
        TestReader( InputStream is )
        {
            this.br = BinReader.asReaderLe( is );
            this.mgr = MingleBinReader.create( br );
        }

        private byte readTypeCode() throws Exception { return br.readByte(); }

        private
        Exception
        failTypeCode( String desc,
                      byte tc )
        {
            return state.createFailf( 
                "Unrecognized %s type code: 0x%02x", desc, tc );
        }

        private
        void
        expectUint32( int expct )
            throws Exception
        {
            int act = br.readInt();

            state.equalf( expct, act,
                "Expected 0x%08x but saw 0x%08x", expct, act );
        }

        private
        void
        expectFileVersion()
            throws Exception
        {
            expectUint32( FILE_VERSION );
        }

        private
        abstract
        class ErrorReader
        {
            final Error err = new Error();

            abstract
            MingleValueException
            buildResult()
                throws Exception;
            
            void
            readField( byte tc )
                throws Exception
            {
                throw failTypeCode( "error field", tc );
            }

            final
            Error
            read()
                throws Exception
            {
                while ( true )
                {
                    byte tc = readTypeCode();

                    switch ( tc )
                    {
                        case TC_MSG: err.msg = br.readUtf8(); break;
                        case TC_PATH: err.loc = mgr.readIdPath(); break;
                        case TC_NIL: return err.complete( buildResult() );

                        default: readField( tc );
                    }
                }
            }
        }

        private
        final
        class TcErrorReader
        extends ErrorReader
        {
            MingleTypeReference expected;
            MingleTypeReference actual;

            MingleValueException
            buildResult()
            {
                return 
                    new MingleTypeCastException( expected, actual, err.loc );
            }

            @Override
            void
            readField( byte tc )
                throws Exception
            {
                switch ( tc )
                {
                    case TC_EXPECTED: expected = mgr.readTypeReference(); break;
                    case TC_ACTUAL: actual = mgr.readTypeReference(); break;
                    default: super.readField( tc );
                }
            }
        }

        private
        final
        class VcErrorReader
        extends ErrorReader
        {
            MingleValueException
            buildResult()
            {
                return new MingleValueCastException( err.msg, err.loc );
            }
        }

        private
        Error
        readError()
            throws Exception
        {
            byte tc = readTypeCode();

            switch ( tc )
            {
                case TC_TC_ERR: return new TcErrorReader().read();
                case TC_VC_ERR: return new VcErrorReader().read();
                default: throw failTypeCode( "error type", tc );
            }
        }

        private
        TestImpl
        readTest()
            throws Exception
        {
            TestImpl res = new TestImpl();

            while ( true )
            {
                byte tc = readTypeCode();

                switch ( tc )
                {
                    case TC_NIL: return res;
                    case TC_IN: res.in = mgr.readValue(); break;
                    case TC_EXPECT: res.expct = mgr.readValue(); break;
                    case TC_TYPE: res.type = mgr.readTypeReference(); break;
                    case TC_PATH: res.path = mgr.readIdPath(); break;
                    case TC_ERR: res.err = readError(); break;
                    default: throw failTypeCode( "test field", tc );
                }
            }
        }

        private
        TestImpl
        readNext()
            throws Exception
        {
            byte tc = readTypeCode();

            switch ( tc )
            {
                case TC_NIL: return null;
                case TC_TEST: return readTest();

                default: 
                    throw state.createFailf( 
                        "Unhandled type code: 0x%02x", tc );
            }
        }

        private
        List< TestImpl >
        readTests()
            throws Exception
        {
            List< TestImpl > res = Lang.newList();

            expectFileVersion();

            TestImpl ti;
            while ( ( ti = readNext() ) != null ) res.add( ti );

            return res;
        }
    }

    @InvocationFactory
    private
    List< TestImpl >
    testCastValue()
        throws Exception
    {
        addExpectations();

        InputStream is = TestData.openFile( "cast-value-tests.bin" );
        try { return new TestReader( is ).readTests(); } finally { is.close(); }
    }

    // To test:
    //
    //  - In addition to message/loc for TypeCastError, check actual
    //  getExpected() and getActual() values on the error object
}
