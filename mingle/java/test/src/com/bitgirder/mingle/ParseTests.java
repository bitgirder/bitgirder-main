package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.TypedString;

import com.bitgirder.io.BinReader;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestCall;
import com.bitgirder.test.LabeledTestObject;
import com.bitgirder.test.InvocationFactory;

import com.bitgirder.testing.TestData;

import java.io.InputStream;
import java.io.BufferedInputStream;

import java.util.List;
import java.util.Arrays;
import java.util.Map;
import java.util.Queue;

@Test
final
class ParseTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static String FILE_NAME = "parser-tests.bin";
    
    private final static int FILE_VERSION = 1;
    
    private final static byte FLD_TEST_TYPE = (byte) 0x01;
    private final static byte FLD_INPUT = (byte) 0x02;
    private final static byte FLD_EXPECT = (byte) 0x03;
    private final static byte FLD_ERROR = (byte) 0x04;
    private final static byte FLD_END = (byte) 0x05;
    private final static byte FLD_EXTERNAL_FORM = (byte) 0x06;

    private final static byte TYPE_IDENTIFIER = (byte) 0x01;
    private final static byte TYPE_NAMESPACE = (byte) 0x02;
    private final static byte TYPE_DECLARED_TYPE_NAME = (byte) 0x03;
    private final static byte TYPE_QUALIFIED_TYPE_NAME = (byte) 0x05;
    private final static byte TYPE_IDENTIFIED_NAME = (byte) 0x06;
    private final static byte TYPE_REGEX_RESTRICTION = (byte) 0x07;
    private final static byte TYPE_RANGE_RESTRICTION = (byte) 0x08;
    private final static byte TYPE_ATOMIC_TYPE_REFERENCE = (byte) 0x09;
    private final static byte TYPE_LIST_TYPE_REFERENCE = (byte) 0x0a;
    private final static byte TYPE_NULLABLE_TYPE_REFERENCE = (byte) 0x0b;
    private final static byte TYPE_NIL = (byte) 0x0c;
    private final static byte TYPE_INT32 = (byte) 0x0d;
    private final static byte TYPE_INT64 = (byte) 0x0e;
    private final static byte TYPE_FLOAT32 = (byte) 0x0f;
    private final static byte TYPE_FLOAT64 = (byte) 0x10;
    private final static byte TYPE_STRING = (byte) 0x11;
    private final static byte TYPE_TIMESTAMP = (byte) 0x12;
    private final static byte TYPE_BOOLEAN = (byte) 0x13;
    private final static byte TYPE_PARSE_ERROR = (byte) 0x14;
    private final static byte TYPE_RESTRICTION_ERROR = (byte) 0x15;
    private final static byte TYPE_STRING_TOKEN = (byte) 0x16;
    private final static byte TYPE_NUMERIC_TOKEN = (byte) 0x17;
    private final static byte TYPE_UINT32 = (byte) 0x18;
    private final static byte TYPE_UINT64 = (byte) 0x19;

    private final static byte ELT_TYPE_FILE_END = (byte) 0x00;
    private final static byte ELT_TYPE_PARSE_TEST = (byte) 0x01;

    private final static Map< ErrorMessageKey, Object > ERR_MSG_OVERRIDES =
        Lang.newMap( ErrorMessageKey.class, Object.class,

            errMsgKey( TestType.STRING, "\"\\u012k\"" ),
                "Invalid hex char in escape: \"k\" (U+006B)",

            errMsgKey( TestType.STRING, "\"\\u01k2\"" ),
                "Invalid hex char in escape: \"k\" (U+006B)",

            errMsgKey( TestType.STRING, "\"\\u012\"" ),
                "Invalid hex char in escape: \"\"\" (U+0022)",

            errMsgKey( TestType.STRING, "\"\\u01\"" ),
                "Invalid hex char in escape: \"\"\" (U+0022)",

            errMsgKey( TestType.STRING, "\"\\u0\"" ),
                "Invalid hex char in escape: \"\"\" (U+0022)",

            errMsgKey( TestType.STRING, "\"\\u\"" ),
                "Invalid hex char in escape: \"\"\" (U+0022)",

            errMsgKey( TestType.STRING, "\"\\k\"" ),
                new ParseErrorExpectation(
                    2, "Unrecognized escape: \\k (U+006B)" ),
        
            errMsgKey( TestType.STRING, "\"a" ), 3,
            errMsgKey( TestType.STRING, "\"\\\"" ), 4,
            errMsgKey( TestType.STRING, "\"\\U001f\"" ), 2,

            errMsgKey( TestType.STRING, "\"abc\u0001def\"" ),
                "Invalid control character U+0001 in string literal",
            
            errMsgKey( TestType.STRING, "\"abc\ndef\"" ),
                "Invalid control character U+000A in string literal",
            
            errMsgKey( TestType.STRING, "\"abc\fdef\"" ),
                "Invalid control character U+000C in string literal",
            
            errMsgKey( TestType.STRING, "\"abc\bdef\"" ),
                "Invalid control character U+0008 in string literal",
            
            errMsgKey( TestType.STRING, "\"abc\rdef\"" ),
                "Invalid control character U+000D in string literal",
            
            errMsgKey( TestType.STRING, "\"abc\tdef\"" ),
                "Invalid control character U+0009 in string literal",
            
            errMsgKey( TestType.STRING, "\"a\\ud834\\u0061\"" ),
                new ParseErrorExpectation(
                    9, "Expected trailing surrogate, found: \"a\" (U+0061)" ),
            
            errMsgKey( TestType.STRING, "\"a\\udd1e\\ud834\"" ),
                "Trailing surrogate U+DD1E is not preceded by a leading " +
                "surrogate",
            
            errMsgKey( TestType.NUMBER, "1.2.3" ),
                "Unexpected char in fractional part: \".\" (U+002E)",
            
            errMsgKey( TestType.NUMBER, "0.x3" ),
                "Number has empty or invalid fractional part",
            
            errMsgKey( TestType.IDENTIFIER, "trailing-input/x" ),
                "Unexpected trailing data \"/\" (U+002F)",
            
            errMsgKey( TestType.IDENTIFIER, "giving-mixedMessages" ),
                "Unexpected identifier character: \"M\" (U+004D)",
            
            errMsgKey( TestType.IDENTIFIER, "a-bad-ch@r" ),
                "Unexpected trailing data \"@\" (U+0040)",
            
            errMsgKey( TestType.NAMESPACE, "ns1:ns2@v1:ns3" ),
                "Unexpected trailing data \":\" (U+003A)",
            
            errMsgKey( TestType.NAMESPACE, "ns1:ns2@v1@v2" ),
                "Unexpected trailing data \"@\" (U+0040)",
            
            errMsgKey( TestType.NAMESPACE, "ns1:ns2@v1/Stuff" ),
                "Unexpected trailing data \"/\" (U+002F)",
            
            errMsgKey( TestType.NAMESPACE, "ns1.ns2@v1" ),
                "Expected ':' or '@' but found: '.'",
            
            errMsgKey( TestType.NAMESPACE, "ns1 : ns2:ns3@v1" ),
                "Unexpected identifier character: \" \" (U+0020)"
        );

    private
    final
    static
    class ErrorMessageKey
    {
        private final Object[] arr;

        private
        ErrorMessageKey( TestType tt,
                         String msg )
        {
            arr = new Object[] { tt, msg };
        }

        public int hashCode() { return Arrays.hashCode( arr ); }

        public
        boolean
        equals( Object o )
        {
            if ( o == this ) return true;
            if ( ! ( o instanceof ErrorMessageKey ) ) return false;

            return Arrays.equals( arr, ( (ErrorMessageKey) o ).arr );
        }
    }

    private
    static
    ErrorMessageKey
    errMsgKey( TestType tt,
               String msg )
    {
        return new ErrorMessageKey( tt, msg );
    }

    private
    static
    enum TestType
    {
        STRING,
        NUMBER,
        IDENTIFIER,
        NAMESPACE,
        DECLARED_TYPE_NAME,
        QUALIFIED_TYPE_NAME,
        IDENTIFIED_NAME,
        TYPE_REFERENCE;
    }

    private
    static
    interface ErrorExpectation
    {}

    private
    final
    static
    class ParseErrorExpectation
    implements ErrorExpectation
    {
        private final int col;
        private final String msg;

        private
        ParseErrorExpectation( int col,
                               String msg )
        {
            this.col = col;
            this.msg = msg;
        }
    }

    // Not used until/if we actually parse restrictions, but leaving in as a
    // placeholder if nothing else
    private
    final
    static
    class RestrictionErrorExpectation
    extends TypedString< RestrictionErrorExpectation >
    implements ErrorExpectation
    {
        private RestrictionErrorExpectation( CharSequence s ) { super( s ); }
    }

    private
    final
    static
    class CoreParseTest
    implements LabeledTestObject,
               TestCall
    {
        private String in;
        private TestType tt;
        private Object expct;
        private String extForm;
        private ErrorExpectation errExpct;

        private
        void
        validate()
        {
            state.notNull( in, "in" );
            state.notNull( tt, "tt" );
        }

        public
        CharSequence
        getLabel()
        {
            return Strings.crossJoin( "=", ",",
                "in", Lang.getRfc4627String( in ),
                "tt", tt
            );
        }

        public Object getInvocationTarget() { return this; }

        private
        < V >
        V
        expectOneTok( Class< V > cls )
            throws Exception
        {
            MingleLexer lex = MingleLexer.forString( in );

            V res = cls.cast( lex.nextToken() );

            state.isTrue( lex.nextToken() == null, "Trailing input" );

            return res;
        }

        private
        Object
        doParse()
            throws Exception
        {
            switch ( tt )
            {
                case STRING: return expectOneTok( MingleString.class );
                case NUMBER: return expectOneTok( MingleLexer.Number.class );
                case IDENTIFIER: return MingleIdentifier.parse( in );
                case NAMESPACE: return MingleNamespace.parse( in );

                default: 
                    throw state.createFailf( "Unhandled test type: %s", tt );
            }
        }

        private
        CharSequence
        extFormOf( Object val )
        {
            switch ( tt )
            {
                case STRING: return ( (MingleString) val ).getExternalForm();

                case IDENTIFIER: 
                    return ( (MingleIdentifier) val ).getExternalForm();

                case NAMESPACE:
                    return ( (MingleNamespace) val ).getExternalForm();

                default: 
                    throw state.createFailf( 
                        "ext form not known for test type: %s", tt );
            }
        }

        private
        void
        assertExpectVal( Object val )
        {
            if ( errExpct == null ) state.equal( expct, val );
            else state.failf( "Got %s but expected error %s", val, errExpct );

            if ( tt != TestType.NUMBER )
            {
                state.equalString( extForm, extFormOf( val ) );
            }
        }

        private
        int
        expectErrCol( int defl )
        {
            Object override = ERR_MSG_OVERRIDES.get( errMsgKey( tt, in ) );

            if ( override == null || override instanceof String ) return defl;

            if ( override instanceof Integer ) return (Integer) override;

            if ( override instanceof ParseErrorExpectation )
            {
                return ( (ParseErrorExpectation) override ).col;
            }

            throw state.createFail( "Unhandled override:", override );
        }

        private
        String
        expectErrString( String defl )
        {
            Object override = ERR_MSG_OVERRIDES.get( errMsgKey( tt, in ) );

            if ( override == null || override instanceof Integer ) return defl;
            if ( override instanceof String ) return (String) override;
            
            if ( override instanceof ParseErrorExpectation )
            {
                return ( (ParseErrorExpectation) override ).msg;
            }

            throw state.createFail( "Unhandled override:", override );
        }

        private
        void
        assertParseError( Exception ex )
            throws Exception
        {
            ParseErrorExpectation pee = (ParseErrorExpectation) errExpct;

            if ( ! (ex instanceof MingleSyntaxException ) ) throw ex;

            MingleSyntaxException mse = (MingleSyntaxException) ex;
            
            state.equalString( expectErrString( pee.msg ), mse.getError() );
            state.equalInt( expectErrCol( pee.col ), mse.getColumn() );
        }

        private
        void
        assertFailure( Exception ex )
            throws Exception
        {
            if ( errExpct == null ) throw ex;
            else if ( errExpct instanceof ParseErrorExpectation )
            {
                assertParseError( ex );
            }
            else state.failf( "Unhandled error expectation: %s", errExpct );
        }

        public
        void
        call()
            throws Exception
        {
            try { assertExpectVal( doParse() ); }
            catch ( Exception ex ) { assertFailure( ex ); }
        }
    }

    private
    final
    static
    class InputException
    extends Exception
    {
        private InputException( String msg ) { super( msg ); }
    }

    private
    final
    static
    class CoreParseTestReader
    {
        private final BinReader rd;

        private boolean readHeader;

        private
        CoreParseTestReader( InputStream is )
        {
            rd = BinReader.asReaderLe( new BufferedInputStream( is ) );
        }

        private
        Exception
        failf( String fmt,
               Object... args )
        {
            return new InputException( String.format( fmt, args ) );
        }

        private
        void
        readHeader()
            throws Exception
        {
            int i = rd.readInt();

            if ( i == FILE_VERSION ) readHeader = true;
            else throw failf( "Unhandled file version: 0x%04x", i );
        }

        private
        int
        readLen( int minVal,
                 String failTmpl )
            throws Exception
        {
            int res = rd.readInt();

            if ( res < minVal ) throw failf( failTmpl, res );
            return res;
        }

        private
        byte
        expectTypeCode( byte tc )
            throws Exception
        {
            byte act = rd.readByte();

            if ( act == tc ) return tc;
            
            throw failf( "Expected type code 0x%02x but got 0x%02x", tc, act );
        }

        private
        TestType
        readTestType()
            throws Exception
        {
            String ttStr = rd.readUtf8();
            String enValStr = ttStr.replace( '-', '_' ).toUpperCase();

            return TestType.valueOf( enValStr );
        }

        private
        Object
        readJvPrimVal()
            throws Exception
        {
            byte tc = rd.readByte();

            switch ( tc )
            {
                case TYPE_BOOLEAN: return rd.readBoolean();
                default: throw failf( "Unrecognized prim type: 0x%02x", tc );
            }
        }

        // Returns null if an empty string is read
        private
        String
        readOptUtf8()
            throws Exception
        {
            String res = rd.readUtf8();
            return res.length() == 0 ? null : res;
        }

        // Read but ignore the actual num token data for now
        private
        MingleLexer.Number
        readNumToken()
            throws Exception
        {
            MingleLexer.Number res = new MingleLexer.Number(
                Boolean.class.cast( readJvPrimVal() ),
                readOptUtf8(),
                readOptUtf8(),
                readOptUtf8()
            );

            rd.readUtf8(); // skip expChar

            return res;
        }

        private
        MingleIdentifier
        readIdentifier( boolean expctTc )
            throws Exception
        {
            if ( expctTc ) expectTypeCode( TYPE_IDENTIFIER );

            int len = readLen( 1, "Invalid id parts len: %d" );

            String[] parts = new String[ len ];
            for ( int i = 0; i < len; ++i ) parts[ i ] = rd.readUtf8();

            return new MingleIdentifier( parts );
        }

        private
        MingleIdentifier[]
        readIdentifiers()
            throws Exception
        {
            int len = readLen( 1, "Invalid namespace parts len: %d" );

            MingleIdentifier[] res = new MingleIdentifier[ len ];
            for ( int i = 0; i < len; ++i ) res[ i ] = readIdentifier( true );

            return res;
        }

        private
        MingleNamespace
        readNamespace( boolean expectTc )
            throws Exception
        {
            if ( expectTc ) expectTypeCode( TYPE_NAMESPACE );

            return new MingleNamespace( 
                readIdentifiers(), 
                readIdentifier( true ) 
            );
        }

        private
        MingleDeclaredTypeName
        readDeclName( boolean expctType )
            throws Exception
        {
            if ( expctType ) expectTypeCode( TYPE_DECLARED_TYPE_NAME );

            return new MingleDeclaredTypeName( rd.readUtf8() );
        }

        private
        QualifiedTypeName
        readQname()
            throws Exception
        {
            return new QualifiedTypeName(
                readNamespace( true ), 
                readDeclName( true ) 
            );
        }

        private
        MingleIdentifiedName
        readIdentifiedName()
            throws Exception
        {
            return new MingleIdentifiedName(
                readNamespace( true ), 
                readIdentifiers() 
            );
        }

        private
        AtomicTypeReference
        readAtomicType()
            throws Exception
        {
            return new AtomicTypeReference(
                (AtomicTypeReference.Name) readVal(),
                (MingleValueRestriction) readVal()
            );
        }

        private
        ListTypeReference
        readListType()
            throws Exception
        {
            return new ListTypeReference(
                (MingleTypeReference) readVal(),
                (Boolean) readJvPrimVal()
            );
        }

        private
        NullableTypeReference
        readNullableType()
            throws Exception
        {
            return new NullableTypeReference( (MingleTypeReference) readVal() );
        }

        private
        MingleRegexRestriction
        readRegexRestriction()
            throws Exception
        {
            return MingleRegexRestriction.create( rd.readUtf8() );
        }

        private
        MingleRangeRestriction
        readRangeRestriction()
            throws Exception
        {
            return MingleRangeRestriction.create(
                (Boolean) readJvPrimVal(),
                (MingleValue) readVal(),
                (MingleValue) readVal(),
                (Boolean) readJvPrimVal()
            );
        }

        private
        Object
        readVal()
            throws Exception
        {
            byte tc = rd.readByte();

            switch ( tc )
            {
                case TYPE_NIL: return null;
                case TYPE_STRING: return new MingleString( rd.readUtf8() );
                case TYPE_BOOLEAN: 
                    return MingleBoolean.valueOf( rd.readBoolean() );
                case TYPE_INT32: return new MingleInt32( rd.readInt() );
                case TYPE_UINT32: return new MingleUint32( rd.readInt() );
                case TYPE_INT64: return new MingleInt64( rd.readLong() );
                case TYPE_UINT64: return new MingleUint64( rd.readLong() );
                case TYPE_FLOAT32: return new MingleFloat32( rd.readFloat() );
                case TYPE_FLOAT64: return new MingleFloat64( rd.readDouble() );
                case TYPE_TIMESTAMP:
                    return MingleTimestamp.create( rd.readUtf8() );
                case TYPE_STRING_TOKEN: 
                    return new MingleString( rd.readUtf8() );
                case TYPE_NUMERIC_TOKEN: return readNumToken();
                case TYPE_IDENTIFIER: return readIdentifier( false );
                case TYPE_NAMESPACE: return readNamespace( false );
                case TYPE_DECLARED_TYPE_NAME: return readDeclName( false );
                case TYPE_QUALIFIED_TYPE_NAME: return readQname();
                case TYPE_IDENTIFIED_NAME: return readIdentifiedName();
                case TYPE_ATOMIC_TYPE_REFERENCE: return readAtomicType();
                case TYPE_LIST_TYPE_REFERENCE: return readListType();
                case TYPE_NULLABLE_TYPE_REFERENCE: return readNullableType();
                case TYPE_REGEX_RESTRICTION: return readRegexRestriction();
                case TYPE_RANGE_RESTRICTION: return readRangeRestriction();

                default: throw failf( "Unrecognized value type: 0x%02x", tc );
            }
        }

        private
        ErrorExpectation
        readErrorExpect()
            throws Exception
        {
            byte tc = rd.readByte();

            switch ( tc )
            {
                case TYPE_PARSE_ERROR: 
                    return new ParseErrorExpectation(
                        rd.readInt(), 
                        rd.readUtf8() 
                    );

                case TYPE_RESTRICTION_ERROR:
                    return new RestrictionErrorExpectation( rd.readUtf8() );

                default: 
                    throw failf( "Unhandled error expect type: 0x%02x", tc );
            }
        }

        private
        boolean
        readField( CoreParseTest cpt )
            throws Exception
        {
            byte fld = rd.readByte();

            switch ( fld )
            {
                case FLD_TEST_TYPE: cpt.tt = readTestType(); return false;
                case FLD_INPUT: cpt.in = rd.readUtf8(); return false;
                case FLD_EXPECT: cpt.expct = readVal(); return false;
                case FLD_ERROR: cpt.errExpct = readErrorExpect(); return false;
                case FLD_EXTERNAL_FORM: 
                    cpt.extForm = rd.readUtf8(); return false;
                case FLD_END: return true;

                default: throw failf( "Unrecognized field type: 0x%02x", fld );
            }
        }

        private
        CoreParseTest
        readParseTest()
            throws Exception
        {
            CoreParseTest res = new CoreParseTest();

            while ( ! readField( res ) );

            res.validate();

            return res;
        }

        private
        CoreParseTest
        next()
            throws Exception
        {
            if ( ! readHeader ) readHeader();

            byte b = rd.readByte();

            switch ( b )
            {
                case ELT_TYPE_FILE_END: return null;
                case ELT_TYPE_PARSE_TEST: return readParseTest();
                default: throw failf( "Unrecognized elt type: 0x%02x", b );
            }
        }
    }


    @InvocationFactory
    private
    List< CoreParseTest >
    testCoreParse()
        throws Exception
    {
        List< CoreParseTest > res = Lang.newList();

        InputStream is = TestData.openFile( FILE_NAME );
        CoreParseTestReader rd = new CoreParseTestReader( is );

        CoreParseTest t = null;
        while ( ( t = rd.next() ) != null ) res.add( t );

        return res;
    }
}
