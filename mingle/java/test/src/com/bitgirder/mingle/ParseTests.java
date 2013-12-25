package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.TypedString;

import com.bitgirder.test.Test;
import com.bitgirder.test.TestCall;
import com.bitgirder.test.LabeledTestObject;
import com.bitgirder.test.InvocationFactory;

import java.util.List;
import java.util.Set;
import java.util.Arrays;
import java.util.Map;

@Test
final
class ParseTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private static void code( Object... args ) { CodeLoggers.code( args ); }

    private 
    static 
    void 
    codef( String tmpl, 
           Object... args ) 
    { 
        CodeLoggers.code( args ); 
    }

    private final static Map< ErrorOverrideKey, Object > ERR_OVERRIDES =
        Lang.newMap( ErrorOverrideKey.class, Object.class,

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
            
            errMsgKey( TestType.DECLARED_TYPE_NAME, "Bad-Char" ),
                "Unexpected trailing data \"-\" (U+002D)",
            
            errMsgKey( TestType.NAMESPACE, "ns1:ns2@v1:ns3" ),
                "Unexpected trailing data \":\" (U+003A)",
            
            errMsgKey( TestType.NAMESPACE, "ns1:ns2@v1@v2" ),
                "Unexpected trailing data \"@\" (U+0040)",
            
            errMsgKey( TestType.NAMESPACE, "ns1:ns2@v1/Stuff" ),
                "Unexpected trailing data \"/\" (U+002F)",
            
            errMsgKey( TestType.NAMESPACE, "ns1.ns2@v1" ),
                "Expected ':' or '@' but found: '.'",
            
            errMsgKey( TestType.NAMESPACE, "ns1 : ns2:ns3@v1" ),
                "Unexpected identifier character: \" \" (U+0020)",
            
            errMsgKey( TestType.QUALIFIED_TYPE_NAME, "ns1/T1" ),
                "Expected ':' or '@' but found: '/'",
            
            errMsgKey( TestType.QUALIFIED_TYPE_NAME, "ns1@v1" ),
                "Expected '/' but found: END",
            
            errMsgKey( TestType.QUALIFIED_TYPE_NAME, "ns1@v1/T1/" ),
                "Unexpected trailing data \"/\" (U+002F)",
            
            errMsgKey( TestType.TYPE_REFERENCE, "ns1@v1" ),
                "Expected '/' but found: END",
            
            errMsgKey( TestType.TYPE_REFERENCE, "/T1" ),
                "Expected identifier or declared type name but found: '/'",
            
            errMsgKey( TestType.TYPE_REFERENCE, "ns1@v1/T1*?-+" ),
                "Unrecognized token start: \"-\" (U+002D)",

            errMsgKey( TestType.TYPE_REFERENCE, "ns1@v1/T1*? +" ),
                "Unrecognized token start: \" \" (U+0020)",
            
            errMsgKey( TestType.TYPE_REFERENCE, "mingle:core@v1~\"s*\"" ),
                "Expected '/' but found: '~'",
            
            errMsgKey( TestType.TYPE_REFERENCE, "S1~12.1" ),
                "Expected restriction but found: NUMBER",

            errMsgKey( 
                TestType.TYPE_REFERENCE, "mingle:core@v1/String~=\"sdf\"" ),
                "Unrecognized token start: \"=\" (U+003D)",
            
            errMsgKey( TestType.TYPE_REFERENCE, "mingle:core@v1/String~" ),
                "Expected type restriction but found END",
            
            errMsgKey( TestType.TYPE_REFERENCE, "String~\"ab[a-z\"" ),
                "(near pattern string char 5) Unclosed character class: " +
                "\"ab[a-z\"",
            
            errMsgKey( TestType.TYPE_REFERENCE, 
                "mingle:core@v1/Timestamp~[\"2012-01-01T12:00:00Z\"," +
                "\"2012-01-02T12:00:00Z\"]" ),
                new ExtFormOverride(
                    "mingle:core@v1/Timestamp~" +
                    "[\"2012-01-01T12:00:00.000000000Z\"," +
                    "\"2012-01-02T12:00:00.000000000Z\"]"
                ),
            
            errMsgKey( TestType.TYPE_REFERENCE, 
                "Timestamp~[\"2012-01-01T12:00:00Z\"," +
                "\"2012-01-02T12:00:00Z\"]" ),
                new ExtFormOverride(
                    "mingle:core@v1/Timestamp~" +
                    "[\"2012-01-01T12:00:00.000000000Z\"," +
                    "\"2012-01-02T12:00:00.000000000Z\"]"
                ),
            
            errMsgKey( TestType.TYPE_REFERENCE, 
                "mingle:core@v1/Float32~[1,2)" ),
                new ExtFormOverride( "mingle:core@v1/Float32~[1.0,2.0)" ),
            
            errMsgKey( TestType.TYPE_REFERENCE, "Float32~[1,2)" ),
                new ExtFormOverride( "mingle:core@v1/Float32~[1.0,2.0)" ),

            errMsgKey( TestType.TYPE_REFERENCE, "Int32~[--3,4)" ),
                "Number has invalid or empty integer part",
            
            errMsgKey( TestType.TYPE_REFERENCE, "Int32~[-\"abc\",2)" ),
                "Number has invalid or empty integer part",

            errMsgKey( TestType.TYPE_REFERENCE, "Int32~[abc,2)" ),
                "Expected range value but found: IDENTIFIER",
            
            errMsgKey( TestType.TYPE_REFERENCE, "Int32~(1:2)" ),
                "Expected ',' but found: ':'",
            
            errMsgKey( TestType.TYPE_REFERENCE, "Int32~[1,3}" ),
                "Unrecognized token start: \"}\" (U+007D)",
            
            errMsgKey( TestType.TYPE_REFERENCE, "Timestamp~[\"2001-0x-22\",)" ),
                "Invalid min value in range restriction: " +
                "(at or near char 0) Invalid timestamp: \"2001-0x-22\""
        );

    private
    final
    static
    class ErrorOverrideKey
    {
        private final Object[] arr;

        private
        ErrorOverrideKey( TestType tt,
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
            if ( ! ( o instanceof ErrorOverrideKey ) ) return false;

            return Arrays.equals( arr, ( (ErrorOverrideKey) o ).arr );
        }

        public
        String
        toString()
        {
            return Lang.asList( arr ).toString();
        }
    }

    private
    static
    ErrorOverrideKey
    errMsgKey( TestType tt,
               String msg )
    {
        return new ErrorOverrideKey( tt, msg );
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
    class ExtFormOverride
    extends TypedString< ExtFormOverride >
    {
        private ExtFormOverride( CharSequence s ) { super( s ); }
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

        private ErrorOverrideKey overrideKey() { return errMsgKey( tt, in ); }

        private
        Object
        override()
        {
            return ERR_OVERRIDES.get( overrideKey() );
        }

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
                case DECLARED_TYPE_NAME: return DeclaredTypeName.parse( in );
                case NAMESPACE: return MingleNamespace.parse( in );
                case IDENTIFIED_NAME: return MingleIdentifiedName.parse( in );
                case QUALIFIED_TYPE_NAME: return QualifiedTypeName.parse( in );
                case TYPE_REFERENCE: return MingleTypeReference.parse( in );

                default: 
                    throw state.createFailf( "Unhandled test type: %s", tt );
            }
        }

        private
        CharSequence
        expectedExtForm()
        {
            Object override = override();

            if ( override instanceof ExtFormOverride ) 
            {
                code( "Returning override ext form:", override );
                return (ExtFormOverride) override;
            }

            return extForm;
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

                case DECLARED_TYPE_NAME:
                    return ( (DeclaredTypeName) val ).getExternalForm();

                case NAMESPACE:
                    return ( (MingleNamespace) val ).getExternalForm();

                case IDENTIFIED_NAME:
                    return ( (MingleIdentifiedName) val ).getExternalForm();
                
                case QUALIFIED_TYPE_NAME:
                    return ( (QualifiedTypeName) val ).getExternalForm();
                
                case TYPE_REFERENCE:
                    return ( (MingleTypeReference) val ).getExternalForm();

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
                state.equalString( expectedExtForm(), extFormOf( val ) );
            }
        }

        private
        int
        expectErrCol( int defl )
        {
            Object override = override();

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
            Object override = override();

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

            if ( ! ( ex instanceof MingleSyntaxException ) ) throw ex;

            MingleSyntaxException mse = (MingleSyntaxException) ex;
            
            state.equalString( expectErrString( pee.msg ), mse.getError() );
            state.equalInt( expectErrCol( pee.col ), mse.getColumn() );
        }

        private
        void
        assertRestrictionError( Exception ex )
            throws Exception
        {
            RestrictionErrorExpectation ee = 
                (RestrictionErrorExpectation) errExpct;

            if ( ! ( ex instanceof MingleSyntaxException ) ) throw ex;

            MingleSyntaxException mse = (MingleSyntaxException) ex;

            state.equalString( 
                expectErrString( ee.toString() ), mse.getError() );
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
            else if ( errExpct instanceof RestrictionErrorExpectation )
            {
                assertRestrictionError( ex );
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
    void
    checkOverrides( List< CoreParseTest > l )
    {
        Set< ErrorOverrideKey > s = Lang.newSet( ERR_OVERRIDES.keySet() );

        for ( CoreParseTest t : l )
        {
            ErrorOverrideKey k = t.overrideKey();
            if ( s.contains( k ) ) s.remove( k );
        }

        state.isTrue( s.isEmpty(), "Unmatched overrides:", s );
    }

    private
    final
    static
    class ReaderImpl
    extends MingleTestGen.StructFileReader< CoreParseTest >
    {
        private ReaderImpl() { super( "parser-tests.bin" ); }

        private
        ParseErrorExpectation
        convertParseErrorExpect( MingleStructAccessor acc )
        {
            return new ParseErrorExpectation(
                acc.expectInt( "col" ),
                acc.expectString( "message" )
            );
        }

        private
        RestrictionErrorExpectation
        convertRestrictErrorExpect( MingleStructAccessor acc )
        {
            return new RestrictionErrorExpectation(
                acc.expectString( "message" ) );
        }

        private
        MingleLexer.Number
        convertNumericToken( MingleStructAccessor acc )
        {
            // we skip expChar

            return new MingleLexer.Number(
                acc.expectBoolean( "negative" ),
                acc.getString( "int" ),
                acc.getString( "frac" ),
                acc.getString( "exp" )
            );
        }

        private
        MingleIdentifier
        convertIdentifier( MingleStructAccessor acc )
        {
            MingleListAccessor la = acc.expectListAccessor( "parts" ); 

            List< String > parts = Lang.newList();

            for ( MingleListAccessor.Traversal t = la.traversal();
                    t.hasNext(); )
            {
                parts.add( t.nextString() );
            }

            return new MingleIdentifier( parts.toArray( new String[] {} ) );
        }

        private
        MingleIdentifier[]
        convertIdList( MingleListAccessor acc )
        {
            List< MingleIdentifier > parts = Lang.newList();

            MingleListAccessor.Traversal t = acc.traversal();

            while ( t.hasNext() ) {
                parts.add( convertIdentifier( t.nextStructAccessor() ) );
            }

            return parts.toArray( new MingleIdentifier[ parts.size() ] );
        }

        private
        MingleNamespace
        convertNamespace( MingleStructAccessor acc )
        {
            return new MingleNamespace(
                convertIdList( acc.expectListAccessor( "parts" ) ),
                convertIdentifier( acc.expectStructAccessor( "version" ) )
            );
        }

        private
        DeclaredTypeName
        convertDeclName( MingleStructAccessor acc )
        {
            return new DeclaredTypeName( acc.expectString( "name" ) );
        }

        private
        QualifiedTypeName
        convertQname( MingleStructAccessor acc )
        {
            return new QualifiedTypeName(
                convertNamespace( acc.expectStructAccessor( "namespace" ) ),
                convertDeclName( acc.expectStructAccessor( "name" ) )
            );
        }

        private
        MingleIdentifiedName
        convertIdentifiedName( MingleStructAccessor acc )
        {
            return new MingleIdentifiedName(
                convertNamespace( acc.expectStructAccessor( "namespace" ) ),
                convertIdList( acc.expectListAccessor( "names" ) )
            );
        }

        private
        MingleRegexRestriction
        convertRegexRestriction( MingleStructAccessor acc )
        {
            return MingleRegexRestriction.
                create( acc.expectString( "pattern" ) );
        }

        private
        MingleRangeRestriction
        convertRangeRestriction( MingleStructAccessor acc,
                                 QualifiedTypeName qn )
        {
            return MingleRangeRestriction.createChecked(
                acc.expectBoolean( "minClosed" ),
                acc.getMingleValue( "min" ),
                acc.getMingleValue( "max" ),
                acc.expectBoolean( "maxClosed" ),
                Mingle.valueClassFor( qn )
            );
        }

        private
        MingleValueRestriction
        convertRestriction( MingleStructAccessor acc,
                            QualifiedTypeName qn )
        {
            String typ = acc.getType().getName().getExternalForm().toString();
            
            if ( typ.equals( "RegexRestriction" ) ) {
                return convertRegexRestriction( acc );
            } else if ( typ.equals( "RangeRestriction" ) ) {
                return convertRangeRestriction( acc, qn );
            }

            throw state.failf( "unhandled restriction: %s", typ );
        }

        private
        AtomicTypeReference
        convertAtomicRef( MingleStructAccessor acc )
        {
            TypeName nm = 
                (TypeName) convertValue( acc.expectStructAccessor( "name" ) );

            MingleValueRestriction rst = null;
            
            MingleStructAccessor rstAcc = 
                acc.getStructAccessor( "restriction" );

            if ( rstAcc != null ) {
                rst = convertRestriction( rstAcc, (QualifiedTypeName) nm );
            }

            return new AtomicTypeReference( nm, rst );
        }

        private
        ListTypeReference
        convertListTypeRef( MingleStructAccessor acc )
        {
            return new ListTypeReference(
                (MingleTypeReference) 
                    convertValue( acc.expectStructAccessor( "elementType" ) ),
                acc.expectBoolean( "allowsEmpty" )
            );
        }

        private
        NullableTypeReference
        convertNullableTypeRef( MingleStructAccessor acc )
        {
            return new NullableTypeReference(
                (MingleTypeReference) 
                    convertValue( acc.expectStructAccessor( "type" ) ) );
        }

        private
        Object
        convertValue( MingleStructAccessor acc )
        {
            if ( acc == null ) return null;

            String typ = acc.getType().getName().getExternalForm().toString();

            if ( typ.equals( "ParseErrorExpect" ) ) {
                return convertParseErrorExpect( acc );
            } else if ( typ.equals( "RestrictionErrorExpect" ) ) {
                return convertRestrictErrorExpect( acc );
            } else if ( typ.equals( "StringToken" ) ) {
                return acc.expectMingleString( "string" );
            } else if ( typ.equals( "NumericToken" ) ) {
                return convertNumericToken( acc );
            } else if ( typ.equals( "Identifier" ) ) {
                return convertIdentifier( acc );
            } else if ( typ.equals( "Namespace" ) ) { 
                return convertNamespace( acc );
            } else if ( typ.equals( "DeclaredTypeName" ) ) {
                return convertDeclName( acc );
            } else if ( typ.equals( "QualifiedTypeName" ) ) {
                return convertQname( acc );
            } else if ( typ.equals( "IdentifiedName" ) ) {
                return convertIdentifiedName( acc );
            } else if ( typ.equals( "AtomicTypeReference" ) ) {
                return convertAtomicRef( acc );
            } else if ( typ.equals( "ListTypeReference" ) ) {
                return convertListTypeRef( acc );
            } else if ( typ.equals( "NullableTypeReference" ) ) {
                return convertNullableTypeRef( acc );
            }

            throw state.failf( "unhandled type: %s", typ );
        }

        public CoreParseTest convertStruct( MingleStruct ms ) 
        {
            CoreParseTest res = new CoreParseTest();
 
            MingleStructAccessor acc = MingleStructAccessor.forStruct( ms );

            res.in = acc.expectString( "in" );

            res.tt = Mingle.asJavaEnumValue( TestType.class, 
                MingleIdentifier.create( acc.expectString( "test-type" ) ) );

            res.extForm = acc.expectString( "externalForm" );

            res.expct = convertValue( acc.getStructAccessor( "expect" ) );

            res.errExpct = (ErrorExpectation) 
                convertValue( acc.getStructAccessor( "error" ) );

            return res;
        }
    }

    @InvocationFactory
    private
    List< CoreParseTest >
    testCoreParse()
        throws Exception
    {
        List< CoreParseTest > res = new ReaderImpl().read();

        checkOverrides( res );

        return res;
    }
}
