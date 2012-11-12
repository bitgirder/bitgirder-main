package com.bitgirder.mingle.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.lang.reflect.ReflectUtils;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.io.IoUtils;
import com.bitgirder.io.Charsets;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.parser.DocumentParser;
import com.bitgirder.parser.SyntaxException;
import com.bitgirder.parser.SourceTextLocation;

import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleIdentifiedName;
import com.bitgirder.mingle.model.MingleNamespace;
import com.bitgirder.mingle.model.MingleTypeName;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.MingleEnum;
import com.bitgirder.mingle.model.MingleString;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleInt64;
import com.bitgirder.mingle.model.MingleDouble;
import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.QualifiedTypeName;

import com.bitgirder.test.Test;
import com.bitgirder.test.LabeledTestCall;
import com.bitgirder.test.TestFailureExpector;
import com.bitgirder.test.InvocationFactory;

import java.util.Iterator;
import java.util.List;
import java.util.Queue;

import java.util.regex.Pattern;

import java.nio.ByteBuffer;

@Test
public
final
class MingleParserTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();
    
    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private MingleParserTests() {}

    private
    ByteBuffer
    loadTestFile( CharSequence name )
        throws Exception
    {
        return 
            IoUtils.toByteBuffer(
                ReflectUtils.getResourceAsStream( getClass(), name ) );
    }

    private
    final
    class ParseTest
    extends LabeledTestCall
    implements TestFailureExpector
    {
        final CharSequence str;
        final Integer errPos;
        final CharSequence expct;

        Class< ? > expctCls;

        private
        ParseTest( CharSequence str,
                   Integer errPos,
                   CharSequence expct )
        {
            super( str );

            this.str = str;
            this.errPos = errPos;
            this.expct = expct;
        }

        private
        ParseTest( CharSequence str,
                   CharSequence expct )
        {   
            this( str, null, expct );
        }

        private ParseTest( CharSequence str ) { this( str, str ); }

        public
        Class< ? extends Throwable >
        expectedFailureClass()
        {
            return errPos == null ? null : SyntaxException.class;
        }

        public 
        CharSequence 
        expectedFailurePattern() 
        { 
            if ( errPos == null ) return null;
            else if ( errPos < 0 ) return expct;
            else
            {
                SourceTextLocation loc = 
                    SourceTextLocation.create( "<>", 1, errPos );

                return Pattern.quote(
                    new SyntaxException( expct.toString(), loc ).getMessage() );
            }
        }

        private
        final
        class TypeRefParseResultBuilder
        implements MingleSyntaxScanner.TypeReferenceBuilder< CharSequence >
        {
            private
            void
            appendRestriction( StringBuilder sb,
                               RestrictionSyntax sx )
            {
                if ( sx instanceof StringRestrictionSyntax )
                {
                    Lang.appendRfc4627String( 
                        sb, ( (StringRestrictionSyntax) sx ).getString() );
                }
                else if ( sx instanceof RangeRestrictionSyntax )
                {
                    RangeRestrictionSyntax rrs = (RangeRestrictionSyntax) sx;

                    MingleValue lv = rrs.leftValue();
                    MingleValue rv = rrs.rightValue();

                    sb.append( rrs.includesLeft() ? "[" : "(" ).
                       append( lv == null ? "" : MingleModels.inspect( lv ) ).
                       append( "," ).
                       append( rv == null ? "" : MingleModels.inspect( rv ) ).
                       append( rrs.includesRight() ? "]" : ")" );
                }
            }

            public
            CharSequence
            buildResult( AtomicTypeReference.Name nm,
                         RestrictionSyntax sx,
                         MingleSyntaxScanner.TypeCompleter tc )
            {
                StringBuilder res = new StringBuilder();
                res.append( nm );
                
                if ( sx != null )
                {
                    res.append( "~" );
                    appendRestriction( res, sx );
                }

                res.append( tc.getQuantifierString() );

                return res;
            }
        }

        private
        CharSequence
        getTypeReferenceParseResult()
            throws Exception
        {
            ScannerImpl si = createScanner( str );

            return 
                MingleParsers.checkTrailingInput( 
                    si.expectTypeReference( new TypeRefParseResultBuilder() ),
                    si
                );
        }

        CharSequence
        getParseResult()
            throws Exception
        {
            if ( expctCls.equals( MingleIdentifier.class ) )
            {
                return MingleParsers.parseIdentifier( str ).getExternalForm();
            }
            else if ( expctCls.equals( MingleTypeName.class ) )
            {
                return MingleParsers.parseTypeName( str ).getExternalForm();
            }
            else if ( expctCls.equals( MingleNamespace.class ) )
            {
                return MingleParsers.parseNamespace( str ).getExternalForm();
            }
            else if ( expctCls.equals( MingleTypeReference.class ) )
            {
                return getTypeReferenceParseResult();
            }
            else if ( expctCls.equals( StandardParseMarker.class ) )
            {
                // Just do the standard parse
                return MingleTypeReference.parse( str ).getExternalForm();
            }
            else if ( expctCls.equals( MingleIdentifiedName.class ) )
            {
                return
                    MingleParsers.parseIdentifiedName( str ).getExternalForm();
            }
            else throw state.createFail( "expctCls:", expctCls );
        }

        public
        void
        call()
            throws Exception
        {
            state.equalString( expct, getParseResult() );
        }
    }

    private
    List< ParseTest >
    asTests( Class< ? > expctCls,
             ParseTest... arr )
    {
        for ( ParseTest t : arr ) t.expctCls = expctCls;
        return Lang.asList( arr );
    }

    @InvocationFactory
    private
    List< ParseTest >
    testIdentifierParse()
    {
        return asTests( MingleIdentifier.class,

            new ParseTest( "test", "test" ),
            new ParseTest( "test1", "test1" ),
            new ParseTest( "test_stuff", "test-stuff" ),
            new ParseTest( "test-stuff", "test-stuff" ),
            new ParseTest( "testStuff", "test-stuff" ),
            new ParseTest( "test2-stuff2", "test2-stuff2" ),
            new ParseTest( "test2_stuff2", "test2-stuff2" ),
            new ParseTest( "test2Stuff2", "test2-stuff2" ),
            new ParseTest( "multiADJAcentCaps", "multi-a-d-j-acent-caps" ),

            new ParseTest( "2bad", 1, "Invalid part beginning: 2" ),
            new ParseTest( "2", 1, "Invalid part beginning: 2" ),
            new ParseTest( "bad-2", 5, "Invalid part beginning: 2" ),
            new ParseTest( "bad-2bad", 5, "Invalid part beginning: 2" ),

            new ParseTest( 
                "AcapCannotStart", 1, "Invalid part beginning: A" ),

            new ParseTest( 
                "giving-mixedMessages", 
                13, 
                "Mixed identifier formats: LC_HYPHENATED and LC_CAMEL_CAPPED"
            ),

            new ParseTest( 
                "too--many-seps", 5, "Invalid part beginning: -" ),

            new ParseTest( "trailing-dash-", 14, "Trailing separator: -" ),

            new ParseTest( 
                "trailing_underscore_", 20, "Trailing separator: _" ),

            new ParseTest( "", 0, "Empty string" )
        );
    }

    @InvocationFactory
    private
    List< ParseTest >
    testParseTypeName()
    {
        return asTests( MingleTypeName.class,
            
            new ParseTest( "A" ),
            new ParseTest( "FooBar" ),
            new ParseTest( "Foo2Bar" ),
            new ParseTest( "Foo2Bar2" ),
            new ParseTest( "AB" ),
            new ParseTest( "AB2" ),
            new ParseTest( "AFizzleDizzle" ),
            new ParseTest( "AFizzleDizzleP" ),

            new ParseTest( 
                "ABad$Char", 
                5, 
                "Type name segments must start with an upper case char, got: $"
            ),

            new ParseTest( 
                "Also-Bad", 
                5, 
                "Type name segments must start with an upper case char, got: -"
            ),

            new ParseTest( 
                "leadingLowerCaseIsBad", 
                1, 
                "Type name segments must start with an upper case char, got: l" 
            ),

            new ParseTest(
                "2BadAsWell",
                1,
                "Type name segments must start with an upper case char, got: 2" 
            ),

            new ParseTest( "", 0, "Empty string" )
        );
    }

    @InvocationFactory
    private
    List< ParseTest >
    testParseNamespace()
    {
        return asTests( MingleNamespace.class,
            
            new ParseTest( "ns@v1" ),
            new ParseTest( "ns1:ns2:ns3@v1" ),
            new ParseTest( "nsCap1:nsCap2:ns3@v1" ),

            new ParseTest( "2bad:ns@v1", 1, "Expected identifier but got: 2" ),

            new ParseTest( "ns1:ns2:", -1, "Unexpected end of input" ),
            
            new ParseTest( 
                "ns1:ns2:@v1", 9, "Expected identifier but got: '@'" ),

            new ParseTest(
                "ns1:non-camel-capped-ident:noGood@v1", 
                8, 
                "Expected '@' but got '-'" 
            ),

            new ParseTest( "ns1.ns2@v1", 4, "Expected '@' but got '.'" ),

            new ParseTest( 
                "ns1 : ns2:ns3@v1", 4, "Expected '@' but got: whitespace" ),
            
            new ParseTest( "ns1:ns2@v1/Stuff", 11, "Trailing input" ),

            new ParseTest( "@v1", 1, "Expected identifier but got: '@'" ),

            new ParseTest( "ns1@V2", 5, "Invalid part beginning: V" ),

            new ParseTest( "ns1@", -1, "Unexpected end of input" ),
            
            new ParseTest( 
                "ns1@ v1", 5, "Expected identifier but got: whitespace" )
        );
    }

    private
    final
    class ScopedParseTest
    extends LabeledTestCall
    {
        private final String testType;

        private final CharSequence strExpct;

        private final MingleIdentifier scopedVer = 
            MingleIdentifier.create( "v1" );

        private
        ScopedParseTest( String testType,
                         CharSequence str,
                         String strExpct )
        {
            super( str );

            this.testType = testType;
            this.strExpct = strExpct;
        }

        private
        ScopedParseTest( String testType,
                         CharSequence str,
                         Integer errPos,
                         CharSequence errPat )
        {
            this( testType, str, null );

            expectFailure( 
                SyntaxException.class,
                "\\Q<> [1," + errPos + "]: " + errPat + "\\E"
            );
        }

        private
        CharSequence
        getParseResult()
            throws Exception
        {
            if ( testType.equals( "ns" ) )
            {
                return 
                    MingleNamespace.parse( getLabel(), scopedVer ).
                    getExternalForm();
            }
            else if ( testType.equals( "typ" ) )
            {
                return 
                    MingleTypeReference.parse( getLabel(), scopedVer ).
                    getExternalForm();
            }
            else throw state.createFail( "Unrecognized testType:", testType );
        }

        public
        void
        call()
            throws Exception
        {
            CharSequence str = getParseResult();
            code( "str:", str );
            state.equalString( strExpct, str );
        }
    }

    @InvocationFactory
    private
    List< ScopedParseTest >
    testParseScoped()
    {
        return Lang.asList(
            
            new ScopedParseTest( "ns", "ns1", "ns1@v1" ),
            new ScopedParseTest( "ns", "ns1:ns2", "ns1:ns2@v1" ),
            new ScopedParseTest( "ns", "ns1:ns2@v1", "ns1:ns2@v1" ),
            new ScopedParseTest( "ns", "ns1:ns2@v2", "ns1:ns2@v2" ),
            new ScopedParseTest( "ns", "ns1@v2", "ns1@v2" ),

            new ScopedParseTest( 
                "ns", "ns1:ns2@BadVer", 9, "Invalid part beginning: B" ),
            
            new ScopedParseTest( "typ", "ns1/T1", "ns1@v1/T1" ),
            new ScopedParseTest( "typ", "ns1:ns2/T1/T2", "ns1:ns2@v1/T1/T2" ),
            new ScopedParseTest( "typ", "ns1@v2/T1", "ns1@v2/T1" ),

            new ScopedParseTest(
                "typ", 
                "ns1@v2/badType", 
                8, 
                "Type name segments must start with an upper case char, got: b"
            ),
            
            new ScopedParseTest( 
                "typ", 
                "ns1/badType", 
                5,
                "Type name segments must start with an upper case char, got: b"
            )
        );
    }

    // These tests only check that the parser correctly parses restriction
    // types and hands them off as RestrictionSyntax to an instance of
    // MingleSyntaxScanner.TypeReferenceBuilder, not that they actually make any
    // sense. For actual test coverage that
    // MingleSyntaxScanner.StandardTypeReferenceBuilder does reasonable things
    // with reasonable inputs, we use the tests in
    // com.bitgirder.mingle.model.ModelTests as well as
    // testParseStandardMingleTypeReference()
    @InvocationFactory
    private
    List< ParseTest >
    testParseMingleTypeReference()
    {
        return asTests( MingleTypeReference.class,
 
            new ParseTest( "mingle:test@v1/Struct1" ),
            new ParseTest( "mingle:test@v1/Struct1/Nested1" ),
            new ParseTest( "Struct1" ),
            new ParseTest( "Struct1/Struct2" ),
            new ParseTest( "mingle:test@v1/Struct1?" ),
            new ParseTest( "Struct1***+?" ),
            new ParseTest( "Struct1 * \n*   \t*+?", "Struct1***+?" ),
            new ParseTest( "Struct1/Struct2*" ),
            new ParseTest( "String1~\"^a+$\"" ),
            new ParseTest( "String1 ~\n\t\"a*\"", "String1~\"a*\"" ),
            new ParseTest( "ns1:ns2@v1/String~\"B*\"" ),
            new ParseTest( "ns1:ns2@v1/String~\"a|b*\"*+" ),
            new ParseTest( "mingle:core@v1/String~\"a$\"" ),
            new ParseTest( "Num~[1,-2]" ),
            new ParseTest( "Num~(,8]?*" ),
            new ParseTest( "Num~(8,)" ),
            new ParseTest( "Num~(-100,100)?" ),
            new ParseTest( "Num~(,)" ),
            new ParseTest( "Str~[\"a\",\"aaaa\")" ),
            new ParseTest( "Num~[-3,- 5]", "Num~[-3,-5]" ),

            new ParseTest( "mingle:test", -1, "Unexpected end of input" ),
            new ParseTest( "Struct1/", -1, "Unexpected end of input" ),
            new ParseTest( "Struct1/*", 9, "Expected type name but got: '*'" ),
            new ParseTest( "mingle:test@v1/Struct1*-+", 24, "Trailing input" ),

            new ParseTest( 
                "mingle:test@v1/Struct-1/Bad", 22, "Trailing input" ),

            new ParseTest( "String~", -1, "Unexpected end of input" ),

            new ParseTest( 
                "mingle:core@v1/~\"s*\"", 
                16, "Expected type name but got: '~'" ),

            new ParseTest( 
                "mingle:core@v1~\"s*\"", 15, "Invalid type name" ),

            new ParseTest( 
                "mingle:core@v1/String ~= \"sdf\"", 
                24, 
                "Unexpected token start: =" 
            ),

            new ParseTest( "Int~(1:2)", 7, "Expected ',' but got ':'" ),
            new ParseTest( "Int~[1,3}", 9, "Invalid range delimiter: }" ),
            new ParseTest( "Int~[abc,2)", 6, "Invalid range literal: abc" ),
            new ParseTest( "Int~[--3,4)", 7, "Can't negate value: -" ),
            new ParseTest( "Int~[,]", 4, "Infinite range must be open" ),
            new ParseTest( "Int~[8,]", 4, "Infinite high range must be open" ),
            new ParseTest( "Int~[,8]", 4, "Infinite low range must be open" ),
            new ParseTest( "S1~12.1", 3, "Unexpected restriction" ),
            new ParseTest( "", -1, "Unexpected end of input" )
        );
    }

    private final static class StandardParseMarker {}

    @InvocationFactory
    private
    List< ParseTest >
    testParseStandardMingleTypeReference()
    {
        return asTests( StandardParseMarker.class,

            new ParseTest( "Stuff" ),
            new ParseTest( "Stuff*" ),
            new ParseTest( "Stuff?" ),
            new ParseTest( "Stuff?*+**" ),
            new ParseTest( "mingle:core@v1/String" ),
            new ParseTest( "mingle:core@v1/String*" ),
            new ParseTest( "mingle:core@v1/String~\"a+\"" ),
            new ParseTest( "mingle:core@v1/String~\"a+\"*" ),
            new ParseTest( "mingle:core@v1/Int64~[-1,1]*" ),

            // Bizarre but allowed nonetheless
            new ParseTest( 
                "mingle:core@v1/Int64~[\"-1\",\"1\"]*",
                "mingle:core@v1/Int64~[-1,1]*" 
            ),

            new ParseTest( "mingle:core@v1/Int32~(,12]?" ),
            new ParseTest( "mingle:core@v1/Int32~(-100,100)?" ),

            new ParseTest( 
                "mingle:core@v1/Double~[-1.1e5,)",
                "mingle:core@v1/Double~[-110000.0,)" 
            ),

            new ParseTest( "mingle:core@v1/Float~(,)" ),

            new ParseTest( 
                "mingle:core@v1/Timestamp~[" +
                "\"2010-01-24T13:15:43.123000000-04:00\"," +
                "\"2010-01-24T13:15:43.123000000-05:00\"]" ),

            new ParseTest( 
                "Stuff~\"a\"", 
                6, 
                "Restrictions not supported for Stuff"
            ),

            new ParseTest( 
                "ns1:ns2@v1/Blah~\"a\"", 
                16, 
                "Don't know how to apply string restriction to ns1:ns2@v1/Blah"
            ),

            new ParseTest( 
                "mingle:core@v1/String~\"ab[a-z\"", 
                22, 
                "Invalid regex: Unclosed character class (near index 5)"
            ),

            new ParseTest(
                "ns1@v1/Blah~[1,2]",
                12,
                "Don't know how to apply range restriction to ns1@v1/Blah"
            ),

            new ParseTest(
                "mingle:core@v1/Int64~\"a*$\"",
                21,
                "Don't know how to apply string restriction to " +
                    "mingle:core@v1/Int64"
            ),

            new ParseTest(
                "mingle:core@v1/Int64~[,12)",
                21,
                "Infinite low range must be open"
            ),

            new ParseTest(
                "mingle:core@v1/Int64~[12,]",
                21,
                "Infinite high range must be open"
            ),

            new ParseTest(
                "mingle:core@v1/Int64~[12,10]", 21, "max < min ( 10 < 12 )" ),

            new ParseTest(
                "mingle:core@v1/Int64~(1.0,2.0)", 
                21, 
                "Invalid integer literal" 
            ),

            // Also check that we catch this when encoded as strings 
            new ParseTest(
                "mingle:core@v1/Int64~(\"1.0\",2.0)", 
                21, 
                "Invalid integer literal"
            ),

            new ParseTest(
                "mingle:core@v1/Buffer~[1,2]",
                22,
                "Don't know how to apply range restriction to " +
                    "mingle:core@v1/Buffer"
            )
        );
    }

    @InvocationFactory
    private
    List< ParseTest >
    testParseIdentifiedName()
    {
        return asTests( MingleIdentifiedName.class,
            
            new ParseTest( "some:ns@v1/someId1" ),
            new ParseTest( "some:ns@v1/someId1/someId2" ),
            new ParseTest( "singleNs@v1/singIdent" ),
        
            new ParseTest( "someNs@v1", -1, "Unexpected end of input" ),
            new ParseTest( "some:ns@v1", -1, "Unexpected end of input" ),
            new ParseTest( "some:ns@v1/", -1, "Unexpected end of input" ),

            new ParseTest( 
                "some:ns@v2/trailingSlash/", -1, "Unexpected end of input" ),

            new ParseTest( "some:ns@v1/some-id1", 16, "Trailing input" ),

            new ParseTest( 
                "some:ns@v1/some_id1", 16, "Unexpected token start: _" ),

            new ParseTest( 
                "some:ns@v1/SomeId", 12, "Invalid part beginning: S" ),

            new ParseTest( "", 0, "Unexpected end of input" ),

            new ParseTest( 
                "/some:ns@v1/noGood/leadingSlash", 
                1, 
                "Expected identifier but got: '/'"
            )
        );
    }

    @Test
    private
    void
    testParseEnumLiteral()
        throws SyntaxException
    {
        MingleEnum expct =
            new MingleEnum.Builder().
                setType( "test:ns1@v1/TestEnum" ).
                setValue( "enum-val1" ).
                build();
        
        MingleEnum actual = 
            MingleParsers.parseEnumLiteral( "test:ns1@v1/TestEnum.enumVal1" );
 
        state.equal( expct.getType(), actual.getType() );
        state.equal( expct.getValue(), actual.getValue() );
    }

    private
    void
    assertPathRoundtrip( ObjectPath< MingleIdentifier > expct )
        throws SyntaxException
    {
        CharSequence pathStr = MingleModels.asString( expct );

        ObjectPath< MingleIdentifier > actual =
            MingleParsers.parseObjectPath( pathStr );
        
        // this is a suitable enough way to assert for now, in the absence of a
        // more generalized pathwalk assertion
        state.equalString( pathStr, MingleModels.asString( actual ) );
    }

    @Test
    private
    void
    testParseObjectPath1()
        throws SyntaxException
    {
        assertPathRoundtrip( 
            ObjectPath.< MingleIdentifier >getRoot().
                descend( MingleParsers.createIdentifier( "elem1" ) ).
                descend( MingleParsers.createIdentifier( "elem2" ) ).
                getListIndex( 14 ).
                descend( MingleParsers.createIdentifier( "elem3" ) ) );
    }

    @Test
    private
    void
    testParseObjectPath2()
        throws SyntaxException
    {
        assertPathRoundtrip(
            ObjectPath.< MingleIdentifier >getRoot().
                descend( MingleParsers.createIdentifier( "elem1" ) ) );
    }

    @Test
    private
    void
    testParseObjectPath3()
        throws SyntaxException
    {
        assertPathRoundtrip(
            ObjectPath.< MingleIdentifier >getRoot().
                descend( MingleParsers.createIdentifier( "elem1" ) ).
                descend( MingleParsers.createIdentifier( "elem2" ) ).
                getListIndex( 32 ) );
    }

    @Test
    private
    void
    testParseObjectPath4()
        throws SyntaxException
    {
        assertPathRoundtrip( ObjectPath.< MingleIdentifier >getRoot() );
    }

    // Not doing exhaustive testing of qname parsing, since that is done
    // implicitly elsewhere, such as in the type ref parsing tests; we're really
    // just after coverage of the public frontend to
    // MingleParsers.(create|parse)QualifiedTypeName()
    @Test
    private
    void
    testParseQname()
        throws SyntaxException
    {
        QualifiedTypeName qn =
            QualifiedTypeName.create(
                MingleNamespace.create( "ns1:ns2@v1" ),
                MingleTypeName.create( "T1" ),
                MingleTypeName.create( "T2" )
            );
        
        state.equal( 
            qn, MingleParsers.createQualifiedTypeName( "ns1:ns2@v1/T1/T2" ) );
        
        state.equal( 
            qn, MingleParsers.parseQualifiedTypeName( "ns1:ns2@v1/T1/T2" ) );
    }

    public
    static
    Queue< MingleToken >
    getTokens( CharSequence fileName,
               CharSequence src )
        throws Exception
    {
        inputs.notNull( src, "src" );
        return MingleParsers.getTokens( fileName, src );
    }

    public
    static
    Queue< MingleToken >
    getTokens( CharSequence src )
        throws Exception
    {
        return getTokens( "<>", src );
    }

    public
    static
    Queue< MingleToken >
    getTokens( CharSequence fileName,
               ByteBuffer src )
        throws Exception
    {
        return getTokens( fileName, Charsets.UTF_8.asString( src ) );
    }

    public
    static
    Queue< MingleToken >
    getTokens( ByteBuffer src )
        throws Exception
    {
        return getTokens( "<>", src );
    }

    private
    final
    static
    class LexerAsserter
    {
        private final Queue< MingleToken > queue;
        private final CharSequence fileName;

        private 
        LexerAsserter( Queue< MingleToken > queue,
                       CharSequence fileName ) 
        { 
            this.queue = queue; 
            this.fileName = fileName;
        }

        < V >
        V
        expectToken( int line,
                     int col,
                     Class< V > expctCls )
        {
            MingleToken tok = queue.remove();
            
            SourceTextLocation loc = tok.getLocation();
            state.equalString( fileName, loc.getFileName() );
            state.equalInt( line, loc.getLine() );
            state.equalInt( col, loc.getColumn() );

            return state.cast( expctCls, tok.getObject() );
        }

        private
        void
        expectStringToken( int line,
                           int col,
                           CharSequence expct,
                           Class< ? extends CharSequence > cls )
        {
            state.equalString( expct, expectToken( line, col, cls ) );
        }

        void
        expectIdentifier( int line,
                          int col,
                          CharSequence expct )
        {
            expectStringToken( line, col, expct, IdentifiableText.class );
        }

        void
        expectWhitespace( int line,
                          int col,
                          CharSequence expct )
        {
            expectStringToken( line, col, expct, WhitespaceText.class );
        }

        void
        expectComment( int line,
                       int col,
                       CharSequence expct )
        {
            expectStringToken( line, col, expct, CommentText.class );
        }

        void
        expectString( int line,
                      int col,
                      CharSequence expct )
        {
            expectStringToken( line, col, expct, MingleString.class );
        }

        void
        expectSpecial( int line,
                       int col,
                       SpecialLiteral expct )
        {
            state.equal( 
                expct, expectToken( line, col, SpecialLiteral.class ) );
        }

        void
        expectTypeName( int line,
                        int col,
                        CharSequence expct )
        {
            expectStringToken( line, col, expct, IdentifiableText.class );
        }

        private
        void
        expectNum( int line,
                   int col,
                   CharSequence expct,
                   Class< ? > expctCls )
        {
            state.equalString(
                expct,
                expectToken( line, col, expctCls ).toString()
            );
        }

        void
        expectInt( int line,
                   int col,
                   CharSequence expct )
        {
            expectNum( line, col, expct, MingleInt64.class );
        }

        void
        expectDecimal( int line,
                       int col,
                       CharSequence expct )
        {
            expectNum( line, col, expct, MingleDouble.class );
        }

        void expectDone() { state.isTrue( queue.isEmpty() ); }
    }

    @Test
    private
    void
    testLexerBasic()
        throws Exception
    {
        ByteBuffer src = loadTestFile( "lex-test1" );
        LexerAsserter la = new LexerAsserter( getTokens( src ), "<>" );

        la.expectIdentifier( 1, 1, "namespace" );
        la.expectWhitespace( 1, 10, " " );
        la.expectIdentifier( 1, 11, "bitgirder" );
        la.expectSpecial( 1, 20, SpecialLiteral.COLON );
        la.expectIdentifier( 1, 21, "mingle" );
        la.expectSpecial( 2, 1, SpecialLiteral.OPEN_BRACE );
        la.expectComment( 4, 1, " Comment at start of line" );

        la.expectComment( 5, 1, 
            "Comment with no ws after # sign (<-- '#' in comment test too)" );

        la.expectWhitespace( 6, 1, "    " );
        la.expectComment( 6, 5, " Indented comment before field" );
        la.expectWhitespace( 7, 1, "\t" );
        la.expectComment( 7, 2, " <-- literal \\t precedes comment" );
        la.expectIdentifier( 8, 1, "camelCappedString" );
        la.expectSpecial( 8, 18, SpecialLiteral.COLON );
        la.expectWhitespace( 8, 19, " " );
        la.expectTypeName( 8, 20, "Text" );
        la.expectSpecial( 8, 24, SpecialLiteral.SEMICOLON );
        la.expectWhitespace( 8, 25, " " );
        la.expectComment( 8, 26, " A comment after stuff on a line" );
        la.expectString( 9, 1, "" );
        la.expectString( 10, 1, "\n\r\t\f\b\"\\/\u01fF" );
        la.expectString( 11, 1, "^[a-z]\\w*$" );
        la.expectInt( 12, 1, "0" );
        la.expectWhitespace( 12, 2, " " );
        la.expectDecimal( 12, 3, "0.0" );
        la.expectWhitespace( 12, 6, " " );
        la.expectSpecial( 12, 7, SpecialLiteral.MINUS );
        la.expectInt( 12, 8, "1" );
        la.expectWhitespace( 12, 9, " " );
        la.expectSpecial( 12, 10, SpecialLiteral.MINUS );
        la.expectDecimal( 12, 11, "1.3E-5" );
        la.expectSpecial( 13, 1, SpecialLiteral.COLON );
        la.expectSpecial( 14, 1, SpecialLiteral.OPEN_BRACE );
        la.expectSpecial( 15, 1, SpecialLiteral.CLOSE_BRACE );
        la.expectSpecial( 16, 1, SpecialLiteral.SEMICOLON );
        la.expectSpecial( 17, 1, SpecialLiteral.TILDE );
        la.expectSpecial( 18, 1, SpecialLiteral.OPEN_PAREN );
        la.expectSpecial( 19, 1, SpecialLiteral.CLOSE_PAREN );
        la.expectSpecial( 20, 1, SpecialLiteral.OPEN_BRACKET );
        la.expectSpecial( 21, 1, SpecialLiteral.CLOSE_BRACKET );
        la.expectSpecial( 22, 1, SpecialLiteral.COMMA );
        la.expectSpecial( 23, 1, SpecialLiteral.QUESTION_MARK );
        la.expectSpecial( 24, 1, SpecialLiteral.MINUS );
        la.expectSpecial( 25, 1, SpecialLiteral.RETURNS );
        la.expectSpecial( 26, 1, SpecialLiteral.FORWARD_SLASH );
        la.expectSpecial( 27, 1, SpecialLiteral.PERIOD );
        la.expectSpecial( 28, 1, SpecialLiteral.ASTERISK );
        la.expectSpecial( 29, 1, SpecialLiteral.PLUS );
        la.expectSpecial( 30, 1, SpecialLiteral.LESS_THAN );
        la.expectSpecial( 31, 1, SpecialLiteral.GREATER_THAN );
        la.expectSpecial( 32, 1, SpecialLiteral.ASPERAND );
        la.expectSpecial( 33, 1, SpecialLiteral.COLON );
        la.expectSpecial( 33, 2, SpecialLiteral.CLOSE_BRACE );
        la.expectSpecial( 33, 3, SpecialLiteral.MINUS );
        la.expectSpecial( 33, 4, SpecialLiteral.RETURNS );

        la.expectDone();
    }

    private
    final
    class ScannerImpl
    extends MingleSyntaxScanner
    {
        private ScannerImpl( Queue< MingleToken > toks ) { super( toks ); }

        private
        void
        assertIdent( CharSequence expct )
            throws SyntaxException
        {
            state.equalString( expct, expectIdentifier().getExternalForm() );
        }

        private void assertEmpty() { state.isTrue( isEmpty() ); }
    }

    private
    ScannerImpl
    createScanner( CharSequence str )
        throws Exception
    {
        ByteBuffer bb = Charsets.UTF_8.asByteBuffer( str );
        return new ScannerImpl( getTokens( bb ) );
    }

    // Though other tests have implicit coverage of this, we want to catch it
    // here independently of them
    @Test
    private
    void
    testLastTokenWithNoEol()
        throws Exception
    {
        ScannerImpl si = createScanner( "implicitEol" );
        si.assertIdent( "implicit-eol" );
        si.assertEmpty();
    }

    @Test
    private
    void
    testScannerPeekVariants()
        throws Exception
    {
        ScannerImpl si = createScanner( "test" );

        MingleToken tok = si.peekToken( IdentifiableText.class );
        state.equalString( "test", (IdentifiableText) tok.getObject() );
        state.equalString( "test", si.peekInstance( IdentifiableText.class ) );

        state.isTrue( si.peekToken( SpecialLiteral.class ) == null );
        state.isTrue( si.peekInstance( SpecialLiteral.class ) == null );

        // check that test is still there
        si.assertIdent( "test" );
        si.assertEmpty();
    }

    @Test
    private
    void
    testScannerSkipNonTrivialWs()
        throws Exception
    {
        ScannerImpl si = createScanner( "  \r\n  getThisOne" );
        si.skipWhitespace();
        si.assertIdent( "get-this-one" );
        si.assertEmpty();
    }

    @Test
    private
    void
    testScannerSkipTrivialWs()
        throws Exception
    {
        ScannerImpl si = createScanner( "noWhitespaceHere" );
        si.skipWhitespace();
        si.assertIdent( "no-whitespace-here" );
        si.assertEmpty();
    }

    // Coverage of expectIdentifier() and expectIdentifier( MingleIdentifier )
    @Test
    private
    void
    testScannerExpectIdentifier()
        throws Exception
    {
        ScannerImpl si = createScanner( "ident1 ident2" );

        si.assertIdent( "ident1" );
        si.skipWhitespace();

        MingleIdentifier id2 = MingleIdentifier.create( "ident2" );

        state.equalString(
            "ident2", 
            (IdentifiableText) si.expectIdentifier( id2 ).getObject() );

        si.assertEmpty();
    }

    @Test
    private
    void
    testScannerExpectWhitespaceMulti()
        throws Exception
    {
        ScannerImpl si = createScanner( "id \n  \tid" );

        si.assertIdent( "id" );

        List< WhitespaceText > l = si.expectWhitespaceMulti();
        state.equalString( " |  \t", Strings.join( "|", l ) );

        si.assertIdent( "id" );
        si.assertEmpty();
    }

    @Test( expected = SyntaxException.class,
           expectedPattern = "Unexpected end of input" )
    private
    void
    testScannerExpectWhitespaceMultiFailsOnEnd()
        throws Exception
    {
        createScanner( "" ).expectWhitespaceMulti();
    }

    @Test( expected = SyntaxException.class,
           expectedPattern = "\\Q<> [1,1]: Expected whitespace\\E" )
    private
    void
    testScannerExpectWhitespaceMultiFailsOnHead()
        throws Exception
    {
        createScanner( "id" ).expectWhitespaceMulti();
    }

    @Test
    private
    void
    testScannerPollIdentifiableText()
        throws Exception
    {
        ScannerImpl si = createScanner( "test+" );

        state.isTrue( si.pollIdentifiableText( "notHere" ) == null );

        state.equalString( 
            "test", 
            (IdentifiableText) si.pollIdentifiableText( "test" ).getObject()
        );

        state.equal( 
            SpecialLiteral.PLUS, 
            si.expectLiteral( SpecialLiteral.PLUS ).getObject() );

        si.assertEmpty();
    }

    @Test
    private
    void
    testScannerExpectText()
        throws Exception
    {
        ScannerImpl si = createScanner( "test" );

        MingleToken tok = si.expectText();
        state.equalString( "test", (IdentifiableText) tok.getObject() );

        si.assertEmpty();
    }

    @Test
    private
    void
    testScannerPollTokenObject()
        throws Exception
    {
        ScannerImpl si = createScanner( "@stuff" );

        MingleToken tok = si.pollTokenObject( SpecialLiteral.ASPERAND );
        state.equal( SpecialLiteral.ASPERAND, tok.getObject() );

        // make sure we removed it
        state.isTrue( si.pollTokenObject( SpecialLiteral.ASPERAND ) == null );

        state.equalString( 
            "stuff", (IdentifiableText) si.expectText( "stuff" ).getObject() );
    }

    @Test
    private
    void
    testScannerPollRestrictionString()
        throws Exception
    {
        ScannerImpl si = createScanner( "~\"^a+$\"" );

        StringRestrictionSyntax sx =
            (StringRestrictionSyntax) si.pollRestrictionSyntax();
 
        code( "sx:", sx );
        state.equalString( "^a+$", sx.getString() );
    }

    @Test
    private
    void
    testScannerPollRestrictionRange()
        throws Exception
    {
        ScannerImpl si = createScanner( "~ [1,3)" );

        RangeRestrictionSyntax sx =
            (RangeRestrictionSyntax) si.pollRestrictionSyntax();

        state.isTrue( sx.includesLeft() );
        state.equal( MingleModels.asMingleInt64( 1 ), sx.leftValue() );
        state.equal( MingleModels.asMingleInt64( 3 ), sx.rightValue() );
        state.isFalse( sx.includesRight() );
    }

    // Fix for early version of code which resulted in NPE instead of
    // SyntaxException
    @Test( expected = SyntaxException.class,
           expectedPattern = "\\Q<> [1,1]: Expected type reference start\\E" )
    private
    void
    testScannerExpectTypeReferenceOnEmptyTypeRefRegression()
        throws Exception
    {
        ScannerImpl si = createScanner( ")" );

        si.expectTypeReference();
    }
}
