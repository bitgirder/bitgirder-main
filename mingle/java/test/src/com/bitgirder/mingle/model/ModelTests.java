package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.io.IoUtils;
import com.bitgirder.io.Base64Encoder;

import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Lang;
import com.bitgirder.lang.TypedString;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPaths;

import com.bitgirder.mingle.parser.MingleParsers;

import com.bitgirder.parser.SyntaxException;
import com.bitgirder.parser.SourceTextLocation;

import com.bitgirder.test.Test;
import com.bitgirder.test.LabeledTestCall;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.TestFailureExpector;

import java.util.Arrays;
import java.util.Iterator;
import java.util.List;
import java.util.Map;
import java.util.Date;
import java.util.GregorianCalendar;

import java.nio.ByteBuffer;

import java.sql.Timestamp;

import java.util.concurrent.Callable;

@Test
final
class ModelTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static ObjectPath< MingleIdentifier > OBJ_PATH1 =
        ObjectPath.getRoot( MingleIdentifier.create( "fld1" ) ).
                   descend( MingleIdentifier.create( "fld2" ) );
    
    private final static ObjectPath< MingleIdentifier > OBJ_PATH2 =
        OBJ_PATH1.startImmutableList().
            next().
            next().
            next().
            descend( MingleIdentifier.create( "fld3" ) );
    
    private final static ObjectPath< MingleIdentifier > OBJ_PATH3 =
        ObjectPath.getRoot( MingleIdentifier.create( "val" ) );

    private final static MingleIdentifier ID_F1 = 
        MingleIdentifier.create( "f1" );
    
    private final static EnumDefinition ENUM1_DEF =
        EnumDefinition.create(
            QualifiedTypeName.create( "mingle:model@v1/Enum1" ),
            Lang.asList(
                MingleIdentifier.create( "e1" ),
                MingleIdentifier.create( "e2" )
            )
        );
    
    private final static TypeDefinitionLookup LOOKUP1;

    private
    void
    assertIdentifier( CharSequence lcFormExpct,
                      MingleIdentifier ident )
    {
        state.equalString( lcFormExpct, ident.getExternalForm() );
    }

    private
    void
    assertIdentifiers( CharSequence extFormExpct,
                       CharSequence... extForms )
    {
        for ( CharSequence extForm : extForms )
        {
            MingleIdentifier ident = MingleParsers.createIdentifier( extForm );
            state.equalString( extFormExpct, ident.getExternalForm() );
        }
    }
 
    @Test
    private
    void
    testCreateIdentifier()
    {
        assertIdentifiers( "hello", "hello" );
        assertIdentifiers( "hello1", "hello1" );

        assertIdentifiers( "hello-there", 
            "hello-there", "hello_there", "helloThere" );

        assertIdentifiers( "hello-there2",
            "hello-there2", "hello_there2", "helloThere2" );
 
        assertIdentifiers( "hello3-there",
            "hello3-there", "hello3_there", "hello3There" );
    }

    private
    void
    assertFormat( MingleIdentifier id,
                  CharSequence expct,
                  MingleIdentifierFormat fmt )
    {
        state.equalString( expct, MingleModels.format( id, fmt ) );
    }

    @Test
    private
    void
    testIdentifierFormatters()
    {
        MingleIdentifier id = MingleParsers.createIdentifier( "test-ident" );

        assertFormat( id, "test-ident", MingleIdentifierFormat.LC_HYPHENATED );
        assertFormat( id, "test_ident", MingleIdentifierFormat.LC_UNDERSCORE );
        assertFormat( id, "testIdent", MingleIdentifierFormat.LC_CAMEL_CAPPED );
    }

    private static enum TestEnum { CONSTANT_ONE; }

    @Test
    private
    void
    testJavaEnumConversions()
    {
        String[] idStrs = 
            new String[] { "constant-one", "constant_one", "constantOne" };

        for ( String idStr : idStrs )
        {
            MingleIdentifier id = MingleParsers.createIdentifier( idStr );

            state.equal(
                TestEnum.CONSTANT_ONE,
                MingleModels.valueOf( TestEnum.class, id ) );
        }

        state.equal( 
            MingleIdentifier.create( "constant-one" ),
            MingleIdentifier.create( TestEnum.CONSTANT_ONE ) );
    }

    private
    void
    assertRfc3339String( MingleTimestamp t,
                         CharSequence expct,
                         CharSequence actual,
                         boolean doParse )
        throws Exception
    {
        state.equalString( expct, actual );
 
        if ( doParse )
        {
            state.equalInt(
                0, t.compareTo( MingleParsers.parseTimestamp( actual ) ) );
        }
    }

    @Test
    private
    void
    testTimestampRfc3339()
        throws Exception
    {
        MingleTimestamp t =
            new MingleTimestamp.Builder().
                setYear( 2010 ).
                setMonth( 4 ).
                setDate( 13 ).
                setHour( 15 ).
                setMinute( 30 ).
                setSeconds( 12 ).
                setNanos( 450100 ).
                build();
 
        String expctBase = "2010-04-13T15:30:12";

        assertRfc3339String( 
            t, expctBase + "Z", t.getRfc3339String( 0 ), false );

        assertRfc3339String( 
            t, expctBase + ".000450100Z", t.getRfc3339String(), true );

        assertRfc3339String(
            t, expctBase + ".00Z", t.getRfc3339String( 2 ), false );
        
        assertRfc3339String(
            t, expctBase + ".000450100Z", t.getRfc3339String( 9 ), true );
    }

    @Test
    private
    void
    testParseTimestamp0()
        throws Exception
    {
        MingleTimestamp ts1 = 
            MingleParsers.parseTimestamp( 
                ModelTestInstances.TEST_TIMESTAMP1_STRING );

        state.equalString( 
            ModelTestInstances.TEST_TIMESTAMP1_STRING, ts1.getRfc3339String() );
    }

    @Test( expected = SyntaxException.class,
           expectedPattern = "^Invalid timestamp:.*" )
    private
    void
    testParseTimestampFails()
        throws Exception
    {
        // missing seconds
        MingleParsers.parseTimestamp( "2009-01-02T22:10+09:00" );
    }

    @Test
    private
    void
    testTimestampConversions()
    {
        MingleTimestamp t = 
            MingleTimestamp.create( "2007-08-24T13:15:43.123456789-08:00" );
        
        // Just checking frac parts for the explicitly constructed timestamp
        Date d = t.asJavaDate();
        state.equalInt( 123, (int) ( d.getTime() % 1000 ) );
        
        GregorianCalendar c = t.asJavaCalendar();
        state.equalInt( 123, (int) ( c.getTimeInMillis() % 1000 ) );

        Timestamp ts = t.asSqlTimestamp();
        state.equalInt( 123456789, ts.getNanos() );
    }

    @Test
    private
    void
    testMingleModelsAccessors()
        throws Exception
    {
        MingleStruct ms = ModelTestInstances.TEST_STRUCT1_INST1;

        // Serves as a base code coverage of ModelTestInstances.assertEqual
        // itself. This is less useful as the library has matured, but early on
        // was a good way to fail fast for certain types of errors. 
        ModelTestInstances.assertEqual( 
            ModelTestInstances.TEST_STRUCT1_INST1, ms );

        MingleSymbolMapAccessor acc = 
            MingleSymbolMapAccessor.create( ms.getFields() );

        state.equalString( "hello", acc.expectString( "string1" ) );
        state.equalInt( 32234, acc.expectInt( "int1" ) );
        state.isTrue( acc.expectBoolean( "bool1" ) );

        state.equal(
            ModelTestInstances.TEST_BYTE_BUFFER1,
            acc.expectByteBuffer( "buffer1" ) );
        
        state.equalString(
            ModelTestInstances.TEST_TIMESTAMP1_STRING, 
            acc.expectMingleTimestamp( "timestamp1" ).getRfc3339String() );
        
        state.equalString(
            ModelTestInstances.TEST_TIMESTAMP2_STRING,
            acc.expectMingleTimestamp( "timestamp2" ).getRfc3339String() );
 
        ModelTestInstances.assertEqual(
            ModelTestInstances.TEST_LIST1, acc.expectMingleList( "list1" ) );

        ModelTestInstances.assertEqual(
            ModelTestInstances.TEST_SYM_MAP1,
            acc.expectMingleSymbolMap( "symbol-map1" ) );
 
        ModelTestInstances.assertEqual(
            ModelTestInstances.TEST_EXCEPTION1_INST1,
            acc.expectMingleException( "exception1" ) );

        ModelTestInstances.assertEqual(
            ModelTestInstances.TEST_ENUM1_CONSTANT1,
            acc.expectMingleEnum( "enum1" ) );
    }

    private
    final
    static
    class SymbolMapImpl
    implements MingleSymbolMap
    {
        private final Map< MingleIdentifier, MingleValue > m;

        private
        SymbolMapImpl( Map< MingleIdentifier, MingleValue > m )
        {
            this.m = Lang.unmodifiableMap( m );
        }

        public MingleValue get( MingleIdentifier id ) { return m.get( id ); }

        public 
        boolean 
        hasField( MingleIdentifier id ) 
        {
            return m.containsKey( id );
        }

        public
        Iterable< MingleIdentifier >
        getFields()
        {
            return m.keySet();
        }
    }

    @Test
    private
    void
    testSymbolMapAccessorConvertsMingleNullToJavaNull()
    {
        Map< MingleIdentifier, MingleValue > m = Lang.newMap();

        m.put( 
            MingleParsers.createIdentifier( "explicit-null" ), 
            MingleNull.getInstance() );

        MingleSymbolMapAccessor acc = 
            MingleSymbolMapAccessor.create( new SymbolMapImpl( m ) );

        state.isTrue( acc.getMingleValue( "explicit-null" ) == null );

        try 
        { 
            acc.expectMingleValue( "explicit-null" ); 
            state.fail( "expectMingleValue() returned normally" );
        }
        catch ( MingleValidationException ignore ) {}
    }

    private
    final
    class AccessTypeReferenceTest
    extends LabeledTestCall
    implements TestFailureExpector
    {
        private final CharSequence refStr;
        private final boolean useExpect;
        private final CharSequence errPat;

        private final MingleIdentifier fld = MingleIdentifier.create( "fld" );

        private
        AccessTypeReferenceTest( CharSequence refStr,
                                 boolean useExpect,
                                 CharSequence errPat )
        {
            super(
                Strings.crossJoin( "=", ",",
                    "refStr", refStr,
                    "useExpect", useExpect
                )
            );

            this.refStr = refStr;
            this.useExpect = useExpect;
            this.errPat = errPat;
        }

        public
        Class< ? extends Throwable >
        expectedFailureClass()
        {
            return errPat == null ? null : MingleValidationException.class;
        }

        public 
        CharSequence 
        expectedFailurePattern() 
        { 
            return fld.getExternalForm() + ": " + errPat; 
        }

        private
        MingleTypeReference
        accessTypeReference( MingleSymbolMapAccessor acc )
        {
            return useExpect 
                ? MingleModels.expectTypeReference( acc, fld )
                : MingleModels.getTypeReference( acc, fld );
        }

        public
        void
        call()
        {
            MingleSymbolMapBuilder b = MingleModels.symbolMapBuilder();

            if ( refStr != null ) b.setString( fld, refStr );

            MingleSymbolMapAccessor acc = 
                MingleSymbolMapAccessor.create( b.build() );
            
            MingleTypeReference ref = accessTypeReference( acc );
            
            if ( state.sameNullity( refStr, ref ) )
            {
                state.equal( MingleTypeReference.create( refStr ), ref );
            }
        }
    }

    @InvocationFactory
    private
    List< AccessTypeReferenceTest >
    testAccessTypeReference()
    {
        List< AccessTypeReferenceTest > l = Lang.newList();

        for ( int i = 0; i < 2; ++i )
        {
            boolean useExpect = i == 0;

            l.add( 
                new AccessTypeReferenceTest( 
                    null, 
                    useExpect, 
                    useExpect ? "value is null" : null
                )
            );

            l.add( 
                new AccessTypeReferenceTest( 
                    "", useExpect, "\\Q<> [1,0]: Unexpected end of input\\E" )
            );

            l.add( 
                new AccessTypeReferenceTest( 
                    "n-n", 
                    useExpect, 
                    "\\Q<> [1,2]: Expected '@' but got '-'\\E" ) );

            l.add( 
                new AccessTypeReferenceTest( "ns1@v1/Foo", useExpect, null ) );
        }

        return l;
    }

    private
    MingleStructBuilder
    createTestStruct1Builder()
    {
        MingleStructBuilder res = MingleModels.structBuilder();

        res.setType( 
            AtomicTypeReference.create(
                ModelTestInstances.TEST_STRUCT1_TYPE.resolveIn(
                    ModelTestInstances.TEST_NS
                )
            )
        );

        return res;
    }

    private
    CharSequence
    format( ObjectPath< MingleIdentifier > path )
    {
        return MingleModels.format( path );
    }

    private
    static
    MingleTypeReference
    typeRef( CharSequence str )
    {
        return MingleTypeReference.create( str );
    }

    private
    final
    static
    class CoercionTest
    extends LabeledTestCall
    implements TestFailureExpector
    {
        private final MingleValue mv;
        private final MingleValue expct;
        private final MingleTypeReference expctTyp;
        private final Class< ? extends Throwable > errCls;
        private final CharSequence errPat;
        private final CharSequence errLoc;

        private
        static
        CharSequence
        makeLabel( MingleValue mv,
                   MingleValue expct,
                   MingleTypeReference expctTyp,
                   Class< ? extends Throwable > errCls,
                   CharSequence errPat,
                   CharSequence errLoc )
        {
            return Strings.crossJoin( "=", ",",
                "from", MingleModels.inspect( mv ),
                "expctVal", expct,
                "expctTyp", expctTyp,
                "errCls", errCls,
                "errPat", errPat,
                "errLoc", errLoc
            );
        }

        private
        CoercionTest( MingleValue mv,
                      MingleValue expct,
                      MingleTypeReference expctTyp,
                      Class< ? extends Throwable > errCls,
                      CharSequence errPat,
                      CharSequence errLoc )
        {
            super( makeLabel( mv, expct, expctTyp, errCls, errPat, errLoc ) );

            this.mv = mv;
            this.expct = expct;
            this.expctTyp = expctTyp;
            this.errCls = errCls;
            this.errPat = errPat;
            this.errLoc = errLoc;
        }

        private
        CoercionTest( MingleValue mv,
                      MingleValue expct )
        {
            this( 
                mv, 
                expct, 
                MingleModels.typeReferenceOf( expct ), 
                null, 
                null, 
                null 
            );
        }

        private
        CoercionTest( MingleValue mv,
                      MingleValue expct,
                      MingleTypeReference expctTyp )
        {
            this( 
                mv, 
                expct, 
                expctTyp, 
                null,
                null,
                null 
            );
        }

        private
        CoercionTest( MingleValue mv,
                      Class< ? extends MingleValue > expctCls )
        {
            this( 
                mv, 
                null, 
                MingleModels.typeReferenceOf( expctCls ), 
                MingleTypeCastException.class,
                null,
                null
            );
        }

        private
        CoercionTest( MingleValue mv,
                      MingleTypeReference expctTyp,
                      CharSequence errLoc )
        {
            this( 
                mv, 
                null, 
                expctTyp, 
                MingleTypeCastException.class,
                null,
                errLoc 
            );
        }

        private
        CoercionTest( MingleValue mv,
                      MingleTypeReference expctTyp,
                      Class< ? extends Throwable > errCls,
                      CharSequence errPat,
                      CharSequence errLoc )
        {
            this( mv, null, expctTyp, errCls, errPat, errLoc );
        }

        public
        Class< ? extends Throwable >
        expectedFailureClass()
        {
            return errCls;
        }

        public
        CharSequence
        expectedFailurePattern()
        {
            if ( errCls == null ) return null;
            else
            {
                String locPref = errLoc == null ? "" : "\\Q" + errLoc + ": \\E";

                if ( errPat == null )
                {
                    return
                        locPref + 
                        "\\QExpected mingle value of type " +
                        expctTyp.getExternalForm() + " but found " +
                        MingleModels.typeReferenceOf( mv.getClass() ) + "\\E";
                }
                else return locPref + errPat;
            }
        }

        public
        void
        call()
            throws Exception
        {
            ObjectPath< MingleIdentifier > path =
                ObjectPath.< MingleIdentifier >getRoot();

            MingleValue val = 
                MingleModels.asMingleInstance( expctTyp, mv, path );

            ModelTestInstances.assertEqual( expct, val );
        }
    }

    private
    MingleTypeReference
    stringRestrictType( CharSequence pat )
    {
        return 
            MingleTypeReference.create(
                "mingle:core@v1/String~" + Lang.getRfc4627String( pat ) );
    }

    // Coercion tests involving number types are handled in a separate
    // InvocationFactory
    @InvocationFactory
    private
    List< CoercionTest >
    testAsMingleInstance()
    {
        // For some of the tests we only want to go as far as millis precision
        // (but we do want to make sure we go that far at least)
        MingleTimestamp ts = 
            MingleTimestamp.create( "2010-01-24T13:15:43.123000000-04:00" );

        return
            Lang.< CoercionTest >asList(

                new CoercionTest(
                    MingleNull.getInstance(),
                    MingleNull.getInstance() ),
                
                new CoercionTest(
                    MingleModels.asMingleString( "hello" ),
                    MingleNull.class
                ),

                new CoercionTest(
                    MingleModels.asMingleString( "hello" ),
                    MingleModels.asMingleString( "hello" ) ),

                new CoercionTest(
                    MingleModels.asMingleBoolean( true ), 
                    MingleModels.asMingleString( "true" ) ),
                
                new CoercionTest( MingleList.create(), MingleString.class ),
 
                new CoercionTest(
                    MingleModels.
                        asMingleBuffer( ModelTestInstances.TEST_BYTE_BUFFER1 ),
                    MingleModels.asMingleString(
                        new Base64Encoder().
                            encode( ModelTestInstances.TEST_BYTE_BUFFER1 ) )
                ),

                new CoercionTest(
                    ModelTestInstances.TEST_TIMESTAMP1,
                    MingleModels.asMingleString(
                        ModelTestInstances.TEST_TIMESTAMP1.getRfc3339String() )
                ),

                new CoercionTest(
                    ModelTestInstances.TEST_ENUM1_CONSTANT1, 
                    MingleModels.asMingleString( "constant1" ) ),
                
                new CoercionTest(
                    MingleModels.asMingleBoolean( true ),
                    MingleModels.asMingleBoolean( true ) ),
                
                new CoercionTest(
                    MingleModels.asMingleString( "true" ),
                    MingleModels.asMingleBoolean( true ) ),
                
                new CoercionTest(
                    MingleModels.asMingleBuffer(
                        ModelTestInstances.TEST_BYTE_BUFFER1 ),
                    MingleModels.asMingleBuffer(
                        ModelTestInstances.TEST_BYTE_BUFFER1 ) ),
                
                new CoercionTest(
                    MingleModels.asMingleString(
                        new Base64Encoder().encode(
                            ModelTestInstances.TEST_BYTE_BUFFER1 ) ),
                    MingleModels.asMingleBuffer(
                        ModelTestInstances.TEST_BYTE_BUFFER1 ) ),
 
                new CoercionTest( ts, ts ),
                
                new CoercionTest(
                    MingleModels.asMingleString( ts.getRfc3339String() ), ts ),
                
                new CoercionTest(
                    MingleList.create(
                        MingleModels.asMingleString( "s1" ),
                        MingleModels.asMingleString( "s2" )
                    ),
                    MingleList.create(
                        MingleModels.asMingleString( "s1" ),
                        MingleModels.asMingleString( "s2" )
                    ),
                    MingleTypeReference.create( "mingle:core@v1/String*" )
                ),

                new CoercionTest(
                    MingleList.create(
                        MingleList.create(
                            MingleModels.asMingleInt32( 1 ),
                            MingleModels.asMingleFloat( 1f )
                        ),
                        MingleList.create(
                            MingleModels.asMingleInt64( 2 ),
                            MingleModels.asMingleInt32( 2 )
                        )
                    ),
                    MingleList.create(
                        MingleList.create(
                            MingleModels.asMingleInt64( 1L ),
                            MingleModels.asMingleInt64( 1L )
                        ),
                        MingleList.create(
                            MingleModels.asMingleInt64( 2L ),
                            MingleModels.asMingleInt64( 2L )
                        )
                    ),
                    MingleTypeReference.create( "mingle:core@v1/Int64**" )
                ),

                new CoercionTest(
                    MingleList.create(
                        MingleModels.asMingleInt64( 1 ),
                        MingleNull.getInstance(),
                        MingleModels.asMingleString( "hi" )
                    ),
                    MingleList.create(
                        MingleModels.asMingleString( "1" ),
                        MingleNull.getInstance(),
                        MingleModels.asMingleString( "hi" )
                    ),
                    MingleTypeReference.create( "mingle:core@v1/String?*" )
                ),

                new CoercionTest(
                    MingleList.create(),
                    MingleList.create(),
                    MingleTypeReference.create( "mingle:core@v1/Int64*" )
                ),

                new CoercionTest(
                    MingleList.create(
                        MingleList.create(),
                        MingleList.create()
                    ),
                    MingleList.create(
                        MingleList.create(),
                        MingleList.create()
                    ),
                    MingleTypeReference.create( "mingle:core@v1/Int64**" )
                ),

                new CoercionTest(
                    MingleList.create(),
                    MingleList.create(),
                    MingleTypeReference.create( "mingle:core@v1/Int64***" )
                ),

                new CoercionTest(
                    MingleList.create(
                        MingleList.create(), // will fail cast
                        MingleModels.asMingleString( "stuff" )
                    ),
                    MingleTypeReference.create( "mingle:core@v1/String*" ),
                    MingleTypeCastException.class,
                    "Expected mingle value of type mingle:core@v1/String but " +
                        "found mingle:core@v1/Value\\*",
                    "[ 0 ]"
                ),

                new CoercionTest(
                    MingleList.create(),
                    MingleTypeReference.create( "mingle:core@v1/String+" ),
                    MingleValidationException.class,
                    "list is empty",
                    null
                ),

                new CoercionTest(
                    MingleList.create(
                        MingleList.create(
                            MingleModels.asMingleString( "s1" ) ),
                        MingleList.create()
                    ),
                    MingleTypeReference.create( "mingle:core@v1/String+*" ),
                    MingleValidationException.class,
                    "list is empty",
                    "[ 1 ]"
                ),

                
                new CoercionTest(
                    MingleModels.asMingleString( "21" ),
                    MingleModels.asMingleInt64( 21L ),
                    MingleTypeReference.create( "mingle:core@v1/Int64?" )
                ),

                new CoercionTest(
                    MingleNull.getInstance(),
                    MingleNull.getInstance(),
                    MingleTypeReference.create( "mingle:core@v1/Int64?" )
                ),

                new CoercionTest(
                    MingleList.create(),
                    MingleTypeReference.create( "mingle:core@v1/String?" ),
                    null
                ),
                
                new CoercionTest(
                    MingleModels.asMingleInt32( 12 ),
                    MingleTypeReference.create( "ns1@v1/SomeStruct" ),
                    null
                ),

                new CoercionTest(
                    MingleModels.structBuilder().setType( "ns@v1/S1" ).build(),
                    MingleModels.structBuilder().setType( "ns@v1/S1" ).build(),
                    MingleTypeReference.create( "ns@v1/S1" )
                ),

                new CoercionTest(
                    MingleModels.asMingleString( "abbbc" ),
                    MingleModels.asMingleString( "abbbc" ),
                    stringRestrictType( "^ab+c$" )
                ),
                
                new CoercionTest(
                    MingleModels.asMingleString( "ac" ),
                    stringRestrictType( "^ab+c$" ),
                    MingleValidationException.class,
                    "\\QValue does not match \"^ab+c$\": \"ac\"\\E",
                    null
                ),

                new CoercionTest(
                    MingleNull.getInstance(),
                    MingleNull.getInstance(),
                    NullableTypeReference.create( stringRestrictType( "a*" ) )
                ),

                new CoercionTest(
                    MingleModels.asMingleString( "aaaaaa" ),
                    MingleModels.asMingleString( "aaaaaa" ),
                    NullableTypeReference.create( stringRestrictType( "a*" ) )
                ),

                new CoercionTest(
                    MingleNull.getInstance(),
                    stringRestrictType( "a*" ),
                    null
                ),

                new CoercionTest(
                    MingleList.create(
                        MingleModels.asMingleString( "a" ),
                        MingleModels.asMingleString( "aaaaaa" )
                    ),
                    MingleList.create(
                        MingleModels.asMingleString( "a" ),
                        MingleModels.asMingleString( "aaaaaa" )
                    ),
                    ListTypeReference.
                        create( stringRestrictType( "a+" ), true )
                ),

                new CoercionTest(
                    MingleList.create(
                        MingleModels.asMingleString( "a" ),
                        MingleNull.getInstance(),
                        MingleModels.asMingleString( "aaaaaa" )
                    ),
                    MingleList.create(
                        MingleModels.asMingleString( "a" ),
                        MingleNull.getInstance(),
                        MingleModels.asMingleString( "aaaaaa" )
                    ),
                    ListTypeReference.create( 
                        NullableTypeReference.
                            create( stringRestrictType( "a+" ) ),
                        true 
                    )
                ),

                new CoercionTest(
                    MingleList.create(
                        MingleModels.asMingleString( "a" ),
                        MingleModels.asMingleString( "b" )
                    ),
                    ListTypeReference.
                        create( stringRestrictType( "a*" ), true ),
                    MingleValidationException.class,
                    "\\QValue does not match \"a*\": \"b\"\\E",
                    "[ 1 ]"
                ),

                new CoercionTest(
                    MingleList.create(
                        MingleModels.asMingleString( "123" ),
                        MingleModels.asMingleInt64( 129 )
                    ),
                    MingleList.create(
                        MingleModels.asMingleString( "123" ),
                        MingleModels.asMingleString( "129" )
                    ),
                    ListTypeReference.
                        create( stringRestrictType( "\\d+" ), true )
                ),

                new CoercionTest(
                    MingleModels.asMingleInt64( 1 ),
                    MingleModels.asMingleInt64( 1 ),
                    typeRef( "mingle:core@v1/Int64~[-1,1]" )
                ),
                
                new CoercionTest(
                    MingleModels.asMingleInt64( 1L ),
                    MingleModels.asMingleInt64( 1 ),
                    typeRef( "mingle:core@v1/Int64~(,2)" )
                ),
                
                new CoercionTest(
                    MingleModels.asMingleString( "1" ),
                    MingleModels.asMingleInt64( 1 ),
                    typeRef( "mingle:core@v1/Int64~[1,1]" )
                ),

                new CoercionTest(
                    MingleModels.asMingleInt64( 1 ),
                    MingleModels.asMingleInt64( 1 ),
                    typeRef( "mingle:core@v1/Int64~(,111)" )
                ),

                // Now just get basic touch coverage of other range types
                new CoercionTest(
                    MingleModels.asMingleInt64( -1 ),
                    MingleModels.asMingleInt64( -1 ),
                    typeRef( "mingle:core@v1/Int64~[-2,32)" )
                ),

                new CoercionTest(
                    MingleModels.asMingleInt32( -1 ),
                    MingleModels.asMingleInt32( -1 ),
                    typeRef( "mingle:core@v1/Int32~[-2,32)" )
                ),

                new CoercionTest(
                    MingleModels.asMingleDouble( -1.1d ),
                    MingleModels.asMingleDouble( -1.1d ),
                    typeRef( "mingle:core@v1/Double~[-2.0,32)" )
                ),

                new CoercionTest(
                    MingleModels.asMingleFloat( -1.1f ),
                    MingleModels.asMingleFloat( -1.1f ),
                    typeRef( "mingle:core@v1/Float~[-2.0,32)" )
                ),

                new CoercionTest(
                    MingleTimestamp.create( "2006-01-01T12:01:02.0-00:00" ),
                    MingleTimestamp.create( "2006-01-01T12:01:02.0-00:00" ),
                    typeRef(
                        "mingle:core@v1/Timestamp~[" +
                            "\"2005-01-01T12:01:02.0-00:00\"," +
                            "\"2007-01-01T12:01:02.0-00:00\"" +
                        "]"
                    )
                ),

                new CoercionTest(
                    MingleModels.asMingleInt32( 12 ),
                    typeRef( "mingle:core@v1/Int32~[0,10)" ),
                    MingleValidationException.class,
                    "\\QValue is not in range [0,10): 12\\E",
                    null
                ),

                new CoercionTest(
                    MingleModels.asMingleInt32( -12 ),
                    typeRef( "mingle:core@v1/Int32~[0,10)" ),
                    MingleValidationException.class,
                    "\\QValue is not in range [0,10): -12\\E",
                    null
                ),

                new CoercionTest(
                    MingleModels.asMingleInt32( 10 ),
                    typeRef( "mingle:core@v1/Int32~[0,10)" ),
                    MingleValidationException.class,
                    "\\QValue is not in range [0,10): 10\\E",
                    null
                ),

                new CoercionTest(
                    MingleModels.asMingleInt32( 0 ),
                    typeRef( "mingle:core@v1/Int32~(0,10]" ),
                    MingleValidationException.class,
                    "\\QValue is not in range (0,10]: 0\\E",
                    null
                ),

                new CoercionTest(
                    MingleTimestamp.create( "2009-01-01T12:01:02.0-00:00" ),
                    typeRef(
                        "mingle:core@v1/Timestamp~[" +
                            "\"2005-01-01T12:01:02.0-00:00\"," +
                            "\"2007-01-01T12:01:02.0-00:00\"" +
                        "]"
                    ),
                    MingleValidationException.class,
                    "\\QValue is not in range " +
                    "[\"2005-01-01T12:01:02.000000000+00:00\"," +
                    "\"2007-01-01T12:01:02.000000000+00:00\"]: " +
                    "2009-01-01T12:01:02.000000000+00:00\\E",
                    null
                )
            );
    }

    private
    void
    addCoercionProduct( MingleValue v1,
                        MingleValue v2,
                        List< CoercionTest > l )
    {
        l.add( new CoercionTest( v1, v2 ) );
        l.add( new CoercionTest( v2, v1 ) );
    }

    private
    void
    addCoercionProduct( MingleValue mv,
                        List< MingleValue > vals,
                        List< CoercionTest > l )
    {
        for ( MingleValue val : vals ) addCoercionProduct( mv, val, l );
    }

    @InvocationFactory
    private
    List< CoercionTest >
    testNumericCoercionPermutations()
    {
        List< CoercionTest > res = Lang.newList();

        int i = 100;

        List< MingleValue > numVals = Lang.< MingleValue >asList(
            MingleModels.asMingleInt64( i ),
            MingleModels.asMingleInt32( i ),
            MingleModels.asMingleDouble( (double) i ),
            MingleModels.asMingleFloat( (float) i )
        );

        for ( MingleValue mv : numVals ) 
        {
            addCoercionProduct( mv, numVals, res );

            String s = Integer.toString( i );

            if ( mv instanceof MingleDouble || mv instanceof MingleFloat )
            {
                s = s + ".0";
            }

            addCoercionProduct( mv, MingleModels.asMingleString( s ), res );
        }

        return res;
    }

    private
    final
    class AsMingleValueTest
    extends LabeledTestCall
    {
        private final Object jvObj;
        private final MingleValue mvExpct;

        private
        AsMingleValueTest( Object jvObj,
                           MingleValue mvExpct )
        {
            super(
                Strings.crossJoin( "=", ",",
                    "jvObj", jvObj,
                    "mvExpct", MingleModels.inspect( mvExpct )
                )
            );

            this.jvObj = jvObj;
            this.mvExpct = mvExpct;
        }

        public
        void
        call()
            throws Exception
        {
            ModelTestInstances.assertEqual(
                mvExpct, MingleModels.asMingleValue( jvObj ) );
        }
    }

    @InvocationFactory
    private
    List< AsMingleValueTest >
    testAsMingleValue()
    {
        Date d1 = new Date();
        GregorianCalendar c1 = new GregorianCalendar();

        return Lang.asList(
            
            new AsMingleValueTest( null, MingleNull.getInstance() ),

            // just test mv --> mv
            new AsMingleValueTest( 
                MingleModels.asMingleString( "hi" ),
                MingleModels.asMingleString( "hi" ) ),

            new AsMingleValueTest(
                Long.valueOf( 12 ),
                MingleModels.asMingleInt64( 12 ) ),
            
            new AsMingleValueTest(
                Integer.valueOf( 12 ),
                MingleModels.asMingleInt32( 12 ) ),
            
            new AsMingleValueTest(
                Short.valueOf( (short) 12 ),
                MingleModels.asMingleInt32( 12 ) ),
            
            new AsMingleValueTest(
                Byte.valueOf( (byte) 12 ),
                MingleModels.asMingleInt32( 12 ) ),
            
            new AsMingleValueTest( 
                Character.valueOf( 'A' ),
                MingleModels.asMingleInt32( (int) 'A' ) ),
            
            new AsMingleValueTest(
                Double.valueOf( 12.1d ),
                MingleModels.asMingleDouble( 12.1d ) ),
            
            new AsMingleValueTest(
                Float.valueOf( 12.1f ),
                MingleModels.asMingleFloat( 12.1f ) ),

            new AsMingleValueTest(
                new byte[] { 0x00b },
                MingleModels.asMingleBuffer( new byte[] { 0x00b } ) ),
            
            new AsMingleValueTest(
                ByteBuffer.wrap( new byte[] { 0x00b } ),
                MingleModels.asMingleBuffer( new byte[] { 0x00b } ) ),

            new AsMingleValueTest(
                "hello", MingleModels.asMingleString( "hello" ) ),
            
            new AsMingleValueTest(
                new StringBuilder( "stuff" ),
                MingleModels.asMingleString( "stuff" ) ),
            
            new AsMingleValueTest(
                Boolean.valueOf( true ), MingleBoolean.TRUE ),
            
            new AsMingleValueTest(
                Lang.asList( 
                    1L, 
                    1,
                    'a',
                    12.1d,
                    12.1f,
                    "stuff",
                    new byte[] { 0x00b },
                    Lang.< Object >asList(),
                    Lang.< Object >asList( "more-stuff" ),
                    Lang.newMap( String.class, Object.class,
                        "key1", 1L,
                        "key2", "val2"
                    )
                ),

                MingleList.create(
                    MingleModels.asMingleInt64( 1L ),
                    MingleModels.asMingleInt32( 1 ),
                    MingleModels.asMingleInt32( (int) 'a' ),
                    MingleModels.asMingleDouble( 12.1d ),
                    MingleModels.asMingleFloat( 12.1f ),
                    MingleModels.asMingleString( "stuff" ),
                    MingleModels.asMingleBuffer( new byte[] { 0x00b } ),
                    MingleList.create(),
                    MingleList.create(
                        MingleModels.asMingleString( "more-stuff" ) ),
                    MingleModels.symbolMapBuilder().
                        setInt64( "key1", 1L ).
                        setString( "key2", "val2" ).
                        build()
                )
            ),

            new AsMingleValueTest(
                new Object[] { "stuff", 12L },
                MingleList.create(
                    MingleModels.asMingleString( "stuff" ),
                    MingleModels.asMingleInt64( 12L )
                )
            ),

            new AsMingleValueTest(
                Lang.newMap( String.class, Object.class, 
                    "key1", "val1",
                    "key2", 12.1d
                ),
                MingleModels.symbolMapBuilder().
                    setString( "key1", "val1" ).
                    setDouble( "key2", 12.1d ).
                    build()
            ),

            new AsMingleValueTest( d1, MingleTimestamp.fromDate( d1 ) ),
            
            new AsMingleValueTest( 
                c1, 
                new MingleTimestamp.Builder().setFromCalendar( c1 ).build()
            )
        );
    }

    private
    final
    static
    class ValueException
    extends RuntimeException
    {
        private ValueException( String msg ) { super( msg ); }
    }

    private
    final
    static
    class ValueErrorFactoryImpl
    implements MingleModels.ValueErrorFactory
    {
        public
        RuntimeException
        createFail( ObjectPath< String > path,
                    String msg )
        {
            return 
                new ValueException( 
                    ObjectPaths.format( path, ObjectPaths.DOT_FORMATTER ) +
                    ": " + msg
                );
        }
    }

    private
    void
    assertMingleValueErrorFactory( Object obj,
                                   String root,
                                   CharSequence errPathStr )
    {
        try
        {
            MingleModels.asMingleValue(
                obj,
                ObjectPath.< String >getRoot( root ),
                new ValueErrorFactoryImpl()
            );

            state.fail();
        }
        catch ( ValueException ve )
        {
            state.equalString(
                errPathStr + ": Can't convert instance of class " +
                    "java.lang.Object to mingle value",
                ve.getMessage()
            );
        }
    }                    

    @Test
    private
    void
    testAsMingleValueErrorFactory()
    {
        assertMingleValueErrorFactory( new Object(), "rootObj", "rootObj" );

        assertMingleValueErrorFactory( 
            new Object[] { 12, new Object() }, "rootObj", "rootObj[ 1 ]" );
    }

    @Test
    private
    void
    testMingleTimestampFromJavaTimeFactories()
        throws Exception
    {
        GregorianCalendar cal =
            ModelTestInstances.TEST_TIMESTAMP2_GREGORIAN_CALENDAR;

        MingleTimestamp ts =
            MingleTimestamp.fromMillis( cal.getTimeInMillis() );
        
        ModelTestInstances.assertEqual( 
            ModelTestInstances.TEST_TIMESTAMP2, ts );

        ModelTestInstances.assertEqual(
            ModelTestInstances.TEST_TIMESTAMP2, 
            MingleTimestamp.fromDate( cal.getTime() ) );
    }

    @Test
    private
    void
    testNegativeNumericTimeFactories()
    {
        MingleTimestamp ts = 
            MingleTimestamp.create( "1960-12-11T10:08:32.002Z" );
 
        long millis = ts.getTimeInMillis();
        state.isTrue( millis < 0L ); // otherwise test isn't really testing much

        state.equal( ts, MingleTimestamp.fromMillis( millis ) );
    }

    // Regression test against a bug found in early MingleTimestamp development
    // in which first version of MingleTimestamp.getTimeInMillis() didn't return
    // the result of adjusting its internal GregorianCalendar instance's
    // getTimeInMillis() value with the separately stored nanos
    @Test
    private
    void
    testMingleTimestampUnixTimeRoundtrip()
        throws Exception
    {
        MingleTimestamp ts = MingleTimestamp.now();

        ModelTestInstances.assertEqual(
            ts, MingleTimestamp.fromMillis( ts.getTimeInMillis() ) );
    }

    private
    void
    assertNumberFormatException( MingleTypeReference typ )
    {
        String numStr = "not-a-number";

        try
        {
            MingleModels.asMingleInstance(
                typ,
                MingleModels.asMingleString( numStr ),
                ObjectPath.< MingleIdentifier >getRoot() );
            
            state.fail();
        }
        catch ( MingleValidationException mve )
        {
            state.equalString( 
                "Invalid number format: " + numStr,
                mve.getDescription() );
        }
    }
 
    @Test
    private
    void
    testMingleInt64NumberFormatException()
    {
        assertNumberFormatException( MingleModels.TYPE_REF_MINGLE_INT64 );
    }
 
    @Test
    private
    void
    testMingleDoubleNumberFormatException()
    {
        assertNumberFormatException( MingleModels.TYPE_REF_MINGLE_DOUBLE );
    }

    // We test the string map assembly exceptions and some other basic cases
    // here, but test successful assembly of nontrivial string mapped structures
    // in model.bind.MingleBindingTests, where it is easier to test and assemble
    // actual values

    @Test
    private
    void
    testStringMapToMingleValueEmptyPairs()
    {
        MingleSymbolMap msm = (MingleSymbolMap)
            MingleModels.stringMapToMingleValue( 
                Lang.< String, String >emptyMap() );
        
        state.isFalse( msm.getFields().iterator().hasNext() );
    }

    @Test( expected = IllegalArgumentException.class,
           expectedPattern = 
            "^Invalid identifier 'ident-234s' in path: bad.ident-234s$" )
    private
    void
    testStringMapToMingleValueBadIdentifierFails()
    {
        MingleModels.stringMapToMingleValue(
            Lang.newMap( String.class, String.class,
                "bad.ident-234s", "value" ) );
    }

    @Test( expected = IllegalArgumentException.class,
           expectedPattern =
            "^Attempt to mix list and field nodes: something$" )
    private
    void
    testStringMapMixedListAndFieldsFails()
    {
        MingleModels.stringMapToMingleValue(
            Lang.newMap( String.class, String.class,
                "something.1", "value1",
                "something.someField", "value2" ) );
    }

    @Test( expected = IllegalArgumentException.class,
           expectedPattern =
            "^Invalid type reference for 'field1:type': bad-type-ref$" )
    private
    void
    testStringMapInvalidTypeReferenceFails()
    {
        MingleModels.stringMapToMingleValue(
            Lang.newMap( String.class, String.class,
                "field1:type", "bad-type-ref" ) );
    }

    @Test( expected = MingleValidationException.class,
           expectedPattern = "fld1.fld2: value is <= 0" )
    private
    void
    testMingleValidationPositiveI()
    {
        MingleValidator v = MingleModels.createValidator( OBJ_PATH1 );

        state.equalInt( 1, MingleValidation.positiveI( v, 1 ) );
        state.equalInt( 1, MingleValidation.positiveI( v, 1, 2 ) );
        state.equalInt( 1, MingleValidation.positiveI( v, null, 1 ) );

        MingleValidation.positiveI( v, -1, 1 );
    }

    @Test( expected = MingleValidationException.class,
           expectedPattern = "fld1.fld2: value is <= 0" )
    private
    void
    testMingleValidationPositiveL()
    {
        MingleValidator v = MingleModels.createValidator( OBJ_PATH1 );

        state.equal( 1L, MingleValidation.positiveL( v, 1L ) );
        state.equal( 1L, MingleValidation.positiveL( v, 1L, 2L ) );
        state.equal( 1L, MingleValidation.positiveL( v, null, 1L ) );

        MingleValidation.positiveL( v, -1L, 1L );
    }

    @Test( expected = MingleValidationException.class,
           expectedPattern = "fld1.fld2: value is < 0" )
    private
    void
    testMingleValidationNonnegativeL()
    {
        MingleValidator v = MingleModels.createValidator( OBJ_PATH1 );

        state.equal( 0L, MingleValidation.nonnegativeL( v, 0L ) );
        state.equal( 3L, MingleValidation.nonnegativeL( v, 3L, 1L ) );
        state.equal( 3L, MingleValidation.nonnegativeL( v, null, 3L ) );

        MingleValidation.nonnegativeL( v, -1L, 1L );
    }

    @Test( expected = MingleValidationException.class,
           expectedPattern = "fld1.fld2: value is < 0" )
    private
    void
    testMingleValidationNonnegativeI()
    {
        MingleValidator v = MingleModels.createValidator( OBJ_PATH1 );

        state.equalInt( 0, MingleValidation.nonnegativeI( v, 0 ) );
        state.equalInt( 3, MingleValidation.nonnegativeI( v, 3, 1 ) );
        state.equalInt( 3, MingleValidation.nonnegativeI( v, null, 3 ) );

        MingleValidation.nonnegativeI( v, -1, 1 );
    }

    @Test( expected = MingleValidationException.class,
           expectedPattern = "fld1.fld2.field4: value is null" )
    private
    void
    testMingleInvocationValidatorExpect()
    {
        MingleInvocationValidator v = 
            MingleModels.createInvocationValidator( OBJ_PATH1 );

        v.expect( "field3", new Object() ); // make sure it passes on non-null
        v.expect( "field4", null );
    }

    // regression for a NPE bug in the sym map builder returned from
    // MingleModels.symbolMapBuilder()
    @Test
    private
    void
    testMingleModelsSymMapBuilderNpeRegression0()
    {
        MingleSymbolMap msm =
            MingleModels.symbolMapBuilder().
                setString( "s1", "hello" ).
                setInt64( "i1", 12 ).
                build();
        
        MingleSymbolMapAccessor acc = MingleSymbolMapAccessor.create( msm );

        state.equalString( "hello", acc.expectString( "s1" ) );
        state.equalInt( 12, acc.expectInt( "i1" ) );
    }

    @Test
    private
    void
    testRelativeTypeNameResolveIn()
    {
        MingleNamespace ns = MingleNamespace.create( "ns1:ns2@v1" );

        MingleTypeName typ = MingleTypeName.create( "Type1" );

        state.equalString(
            "ns1:ns2@v1/Type1",
            RelativeTypeName.create( typ ).resolveIn( ns ).getExternalForm() );
        
        state.equalString(
            "ns1:ns2@v1/Type1/Type1",
            RelativeTypeName.create( typ, typ ).
                resolveIn( ns ).
                getExternalForm()
        );
    }

    @Test
    private
    void
    testMingleTypeNameResolveIn()
    {
        state.equalString(
            "ns1:ns2@v1/Type1",
            MingleTypeName.create( "Type1" ).
                resolveIn( MingleNamespace.create( "ns1:ns2@v1" ) ).
                getExternalForm()
        );
    }

    private
    void
    assertPrimDefModelClass( CharSequence typeStr,
                             Class< ? extends MingleValue > expctCls )
    {
        state.equal(
            expctCls,
            PrimitiveDefinition.modelClassFor(
                QualifiedTypeName.create(
                    MingleNamespace.create( "mingle:core@v1" ),
                    MingleTypeName.create( typeStr )
                )
            )
        );
    }

    @Test
    private
    void
    testPrimitiveDefinitionModelClassFor()
    {
        assertPrimDefModelClass( "Value", MingleValue.class );
        assertPrimDefModelClass( "Null", MingleNull.class );
        assertPrimDefModelClass( "String", MingleString.class );
        assertPrimDefModelClass( "Int64", MingleInt64.class );
        assertPrimDefModelClass( "Int32", MingleInt32.class );
        assertPrimDefModelClass( "Double", MingleDouble.class );
        assertPrimDefModelClass( "Float", MingleFloat.class );
        assertPrimDefModelClass( "Boolean", MingleBoolean.class );
        assertPrimDefModelClass( "Timestamp", MingleTimestamp.class );
        assertPrimDefModelClass( "Buffer", MingleBuffer.class );
    }

    private
    void
    assertPrimDefForQname( CharSequence qnStr )
    {
        QualifiedTypeName qn = QualifiedTypeName.create( qnStr );

        state.equal(
            qn, state.notNull( PrimitiveDefinition.forName( qn ) ).getName() );
    }

    @Test
    private
    void
    testPrimDefForQname()
    {
        assertPrimDefForQname( "mingle:core@v1/Value" );
        assertPrimDefForQname( "mingle:core@v1/Null" );
        assertPrimDefForQname( "mingle:core@v1/String" );
        assertPrimDefForQname( "mingle:core@v1/Int64" );
        assertPrimDefForQname( "mingle:core@v1/Int32" );
        assertPrimDefForQname( "mingle:core@v1/Double" );
        assertPrimDefForQname( "mingle:core@v1/Float" );
        assertPrimDefForQname( "mingle:core@v1/Boolean" );
        assertPrimDefForQname( "mingle:core@v1/Timestamp" );
        assertPrimDefForQname( "mingle:core@v1/Buffer" );
    }

    @Test
    private
    void
    testRelativeNameEqualsAndHashCode()
    {
        MingleTypeName t1 = MingleTypeName.create( "T1" );
        MingleTypeName t2 = MingleTypeName.create( "T2" );
        MingleTypeName t3 = MingleTypeName.create( "T3" );

        RelativeTypeName nmX1 = RelativeTypeName.create( t1, t2 );
        RelativeTypeName nmX2 = RelativeTypeName.create( t1, t2 );
        RelativeTypeName nmY = RelativeTypeName.create( t1, t3 );

        state.equalInt( nmX1.hashCode(), nmX2.hashCode() );
        state.isFalse( nmX1.hashCode() == nmY.hashCode() );

        state.isTrue( nmX1.equals( nmX1 ) );
        state.equal( nmX1, nmX2 );
        state.isFalse( nmX1.equals( nmY ) );
    }

    @Test
    private
    void
    testNsEquality()
    {
        MingleNamespace ns1 = MingleNamespace.create( "ns1:ns2@v1" );
        MingleNamespace ns2 = MingleNamespace.create( "ns1:ns2@v1" );
        MingleNamespace ns3 = MingleNamespace.create( "ns1:ns3@v1" );
        MingleNamespace ns4 = MingleNamespace.create( "ns1:ns2@v2" );

        assertEquality( ns1, ns2, ns3 );
        assertEquality( ns1, ns2, ns4 );

        state.isFalse( ns3.equals( ns4 ) );
    }

    private
    void
    assertTypeNameIn( QualifiedTypeName qn,
                      MingleTypeReference ref )
    {
        state.equal( qn, MingleModels.typeNameIn( ref ) );
    }

    @Test
    private
    void
    testTypeNameIn()
    {
        QualifiedTypeName qn = 
            QualifiedTypeName.create(
                MingleNamespace.create( "ns1:ns2@v1" ),
                MingleTypeName.create( "Type1" ),
                MingleTypeName.create( "Type2" )
            );
        
        MingleTypeReference ref = AtomicTypeReference.create( qn );
        
        assertTypeNameIn( qn, ref );
        assertTypeNameIn( qn, ref = NullableTypeReference.create( ref ) );
        assertTypeNameIn( qn, ref = NullableTypeReference.create( ref ) );
        assertTypeNameIn( qn, ref = ListTypeReference.create( ref, true ) );
        assertTypeNameIn( qn, ref = ListTypeReference.create( ref, false ) );
        assertTypeNameIn( qn, ref = NullableTypeReference.create( ref ) );
        assertTypeNameIn( qn, ref = ListTypeReference.create( ref, false ) );
    }

    @Test
    private
    void
    testQnameEqualsAndHashCode()
    {
        QualifiedTypeName x1 = QualifiedTypeName.create( "foo:bar@v1/X" );
        QualifiedTypeName x2 = QualifiedTypeName.create( "foo:bar@v1/X" );
        QualifiedTypeName y = QualifiedTypeName.create( "foo:bar@v1/Y" );

        state.isFalse( x1 == x2 ); // otherwise we're not really testing much
        state.equal( x1, x2 );
        state.equalInt( x1.hashCode(), x2.hashCode() );

        state.isFalse( x1.equals( y ) );
    }

    private
    void
    assertTypeRefOf( Class< ? extends MingleValue > cls )
    {
        // Skip leading "Mingle"
        String smplNm = cls.getSimpleName().substring( 6 );
        String expct = "mingle:core@v1/" + smplNm;

        state.equalString(
            expct, MingleModels.typeReferenceOf( cls ).getExternalForm() );
    }

    @Test
    private
    void
    testTypeRefOfForCoreValueTypes()
    {
        assertTypeRefOf( MingleBoolean.class );
        assertTypeRefOf( MingleBuffer.class );
        assertTypeRefOf( MingleDouble.class );
        assertTypeRefOf( MingleEnum.class );
        assertTypeRefOf( MingleException.class );
        assertTypeRefOf( MingleFloat.class );
        assertTypeRefOf( MingleInt32.class );
        assertTypeRefOf( MingleInt64.class );
        assertTypeRefOf( MingleNull.class );
        assertTypeRefOf( MingleString.class );
        assertTypeRefOf( MingleStruct.class );
        assertTypeRefOf( MingleStructure.class );
        assertTypeRefOf( MingleSymbolMap.class );
        assertTypeRefOf( MingleTimestamp.class );
        assertTypeRefOf( MingleValue.class );
    }

    @Test
    private
    void
    testTypeRefOfValues()
    {
        // We don't exhaustively check all non-declared types for now, but just
        // make sure that fall through is working for at least one (String
        // below)
        state.equalString(
            "mingle:core@v1/String",
            MingleModels.typeReferenceOf( MingleModels.asMingleString( "" ) ).
                getExternalForm()
        );

        state.equalString(
            "ns1@v1/Blah",
            MingleModels.typeReferenceOf(
                MingleModels.structBuilder().setType( "ns1@v1/Blah" ).build()
            ).
            getExternalForm()
        );
        
        state.equalString(
            "ns1@v1/Blah",
            MingleModels.typeReferenceOf(
                MingleModels.exceptionBuilder().setType( "ns1@v1/Blah" ).build()
            ).
            getExternalForm()
        );

        state.equalString(
            "ns1@v1/Blah",
            MingleModels.typeReferenceOf(
                new MingleEnum.Builder().
                    setValue( "blah" ).
                    setType( "ns1@v1/Blah" ).
                    build()
            ).
            getExternalForm()
        );
    }

    @Test
    private
    void
    testGetFieldWithDefault()
        throws Exception
    {
        FieldDefinition fd = 
            new FieldDefinition.Builder().
                setType( MingleTypeReference.create( "mingle:core@v1/Int64" ) ).
                setName( MingleIdentifier.create( "f1" ) ).
                setDefault( MingleModels.asMingleInt64( 1L ) ).
                build();
        
        MingleSymbolMap m1 = MingleModels.getEmptySymbolMap();

        MingleSymbolMap m2 = 
            MingleModels.symbolMapBuilder().setInt64( "f1", 2L ).build();

        ModelTestInstances.assertEqual(
            MingleModels.asMingleInt64( 1L ), MingleModels.get( m1, fd ) );

        ModelTestInstances.assertEqual(
            MingleModels.asMingleInt64( 2L ), MingleModels.get( m2, fd ) );
    }

    // Seems silly but worth having some coverage anyway
    @Test
    private
    void
    testGetFieldWithoutDefault()
        throws Exception
    {
        FieldDefinition fd =
            new FieldDefinition.Builder().
                setType( MingleTypeReference.create( "mingle:core@v1/Int64" ) ).
                setName( MingleIdentifier.create( "f1" ) ).
                build();
        
        MingleSymbolMap m1 = MingleModels.getEmptySymbolMap();

        MingleSymbolMap m2 = 
            MingleModels.symbolMapBuilder().setInt64( "f1", 1L ).build();

        ModelTestInstances.assertEqual(
            MingleNull.getInstance(), MingleModels.get( m1, fd ) );
 
        ModelTestInstances.assertEqual(
            MingleModels.asMingleInt64( 1L ), MingleModels.get( m2, fd ) );
    }

    @Test
    private
    void
    testGetFieldReturnsImplicitEmptyList()
        throws Exception
    {
        for ( String quant : new String[] { "*", "**", "+*", "?+*" } )
        {
            String typStr = "ns1@v1/Struct1" + quant;

            ModelTestInstances.assertEqual(
                MingleModels.getEmptyList(),
                MingleModels.get(
                    MingleModels.getEmptySymbolMap(),
                    new FieldDefinition.Builder().
                        setType( MingleTypeReference.create( typStr ) ).
                        setName( MingleIdentifier.create( "f1" ) ).
                        build()
                )
            );
        }
    }

    @Test
    private
    void
    testGetFieldNullOnNonAbsentEmptyListTypeValue()
        throws Exception
    {
        ModelTestInstances.assertEqual(
            MingleNull.getInstance(),
            MingleModels.get(
                MingleModels.getEmptySymbolMap(),
                new FieldDefinition.Builder().
                    setType( 
                        MingleTypeReference.
                            create( "mingle:core@v1/String+" ) ).
                    setName( MingleIdentifier.create( "f1" ) ).
                    build()
            )
        );
    }

    // Since our initial version of mingle runtimes serialize type defs using
    // JSON, which can't differentiate between mingle literals of type int32,
    // int64, or integral (similarly decimal, double, or float), we have this
    // test in place to ensure that field defs rebuilt from json encoded type
    // defs correctly coerce the default vals into the particular runtime type.
    // We start by just testing string and an int type for now; others will
    // follow if there is a need to test them separately
    @Test
    private
    void
    testStringFieldDefinitionDefaultTypeCoercion()
    {
        // Somewhat contrived, since it's unlikely this would ever occur (int
        // literal instead of string) but good just to stress that String is
        // handled correctly
        FieldDefinition fd1 =
            new FieldDefinition.Builder().
                setType( 
                    MingleTypeReference.create( "mingle:core@v1/String" ) ).
                setName( MingleIdentifier.create( "f1" ) ).
                setDefault( MingleModels.asMingleInt64( 12121212 ) ).
                build();
        
        state.equalString( 
            "12121212", state.cast( MingleString.class, fd1.getDefault() ) );
    } 

    @Test
    private
    void
    testInt64FieldDefinitionDefaultTypeCoercion()
    {
        FieldDefinition fd1 =
            new FieldDefinition.Builder().
                setType( MingleTypeReference.create( "mingle:core@v1/Int64" ) ).
                setName( MingleIdentifier.create( "f1" ) ).
                setDefault( MingleModels.asMingleInt32( 1 ) ).
                build();
 
        state.equal(
            MingleModels.asMingleInt64( 1L ),
            state.cast( MingleInt64.class, fd1.getDefault() )
        );
    }

    // Regression put in during some refactoring that introduced a new null
    // value bug in MingleSymbolMapAccessor.get*() methods. Now fixed to return
    // null as expected but not to fail
    @Test
    private
    void
    testMingleSymbolMapAccessGetValueNullRegression()
    {
        state.equal(
            null,
            MingleSymbolMapAccessor.create( MingleModels.getEmptySymbolMap() ).
                getString( "whatever" )
        );
    }

    // basic equality test for three objects, where expectation is that x1 !=
    // x2, x1.equals( x2 ), and not x1.equals( y )
    private
    void
    assertEquality( Object x1,
                    Object x2,
                    Object y )
    {
        state.isFalse( x1 == x2 ); // otherwise the test is trivial        

        state.equal( x1, x1 );
        state.equal( x1, x2 );
        state.equal( x2, x1 );
        state.equalInt( x1.hashCode(), x2.hashCode() );
        state.isFalse( x1.equals( y ) );
        state.isFalse( y.equals( x1 ) );
        state.isFalse( x1.equals( null ) );
    }

    @Test
    private
    void
    testRegexRestrictionEquality()
    {
        MingleRegexRestriction x1 = MingleRegexRestriction.create( "ab" );
        MingleRegexRestriction x2 = MingleRegexRestriction.create( "ab" );
        MingleRegexRestriction y = MingleRegexRestriction.create( "zz" );

        assertEquality( x1, x2, y );
    }

    @Test
    private
    void
    testRangeRestrictionEquality()
    {
        assertEquality(
            MingleRangeRestriction.create( 
                true, 
                MingleModels.asMingleInt64( 1 ),
                MingleModels.asMingleInt64( 2 ),
                true,
                MingleInt64.class
            ),
            MingleRangeRestriction.create( 
                true, 
                MingleModels.asMingleInt64( 1 ),
                MingleModels.asMingleInt64( 2 ),
                true,
                MingleInt64.class
            ),
            MingleRangeRestriction.create( 
                true, 
                MingleModels.asMingleInt32( 1 ),
                MingleModels.asMingleInt32( 2 ),
                true,
                MingleInt32.class
            )
        );
    }

    @Test
    private
    void
    testAtomicTypeRefEqualityWithRestrictions()
    {
        MingleRegexRestriction x1 = MingleRegexRestriction.create( "ab" );
        MingleRegexRestriction x2 = MingleRegexRestriction.create( "ab" );
        MingleRegexRestriction y = MingleRegexRestriction.create( "zz" );

        QualifiedTypeName qnX1 = QualifiedTypeName.create( "ns@v1/S1" );
        QualifiedTypeName qnX2 = QualifiedTypeName.create( "ns@v1/S1" );
        QualifiedTypeName qnY = QualifiedTypeName.create( "ns@v1/S2" );

        // check identity
        AtomicTypeReference typ1 = AtomicTypeReference.create( qnX1, x1 );
        state.equal( typ1, typ1 );

        state.equal(
            AtomicTypeReference.create( qnX1, x1 ),
            AtomicTypeReference.create( qnX2, x1 ) );

        state.equal(
            AtomicTypeReference.create( qnX1, x1 ),
            AtomicTypeReference.create( qnX1, x2 ) );
        
        state.isFalse(
            AtomicTypeReference.create( qnX1, x1 ).equals(
                AtomicTypeReference.create( qnX1, y ) ) );
        
        state.isFalse(
            AtomicTypeReference.create( qnX1, x1 ).equals(
                AtomicTypeReference.create( qnY, x1 ) ) );
    }

    private
    void
    assertTypeRefEqualsImpl( CharSequence refStr )
    {
        MingleTypeReference t1 = typeRef( refStr );
        MingleTypeReference t2 = typeRef( refStr );

        state.isFalse( t1 == t2 ); // sanity check that this is a real test

        state.equalInt( t1.hashCode(), t2.hashCode() );
        state.isTrue( t1.equals( t2 ) );
    }

    private
    void
    assertTypesDiffer( CharSequence refStr1,
                       CharSequence refStr2 )
    {
        state.isFalse( typeRef( refStr1 ).equals( refStr2 ) );
    }

    // for list and nullable type ref equals/hashcode tests we don't re-test
    // various atomic type bases and rely on the fact that list/nullable type
    // equality derives its correctness, aside from the list/nullable specific
    // part, recursively on the correctness of equality as implemented by its
    // component type
    @Test
    private
    void
    testListTypeEquals()
    {
        assertTypeRefEqualsImpl( "A*" );
        assertTypeRefEqualsImpl( "A+" );
        assertTypeRefEqualsImpl( "A/B*" );
        assertTypeRefEqualsImpl( "ns1:ns2@v1/A/B**+*" );

        assertTypesDiffer( "A*", "B*" );
        assertTypesDiffer( "A+", "A*" );
        assertTypesDiffer( "A/A+", "A+" );
        assertTypesDiffer( "ns1@v1/A*", "A*" );
        assertTypesDiffer( "ns1@v1/A+", "ns1@v1/A*" );
        assertTypesDiffer( "A**+*", "A**+" );
    }

    @Test
    private
    void
    testNullableTypeEquals()
    {
        assertTypeRefEqualsImpl( "A?" );
        assertTypeRefEqualsImpl( "ns1@v1/A/B?" );
        assertTypeRefEqualsImpl( "ns1:ns2@v1/A?" );
        assertTypeRefEqualsImpl( "ns1@v1/A*+?" );

        assertTypesDiffer( "A", "A?" );
        assertTypesDiffer( "A?", "B?" );
        assertTypesDiffer( "ns1@v1/A", "ns1@v1/A?" );
        assertTypesDiffer( "ns1@v1/A?", "ns2@v1/A?" );
        assertTypesDiffer( "ns1@v1/A+?", "ns1@v1/A*?" );
    }

    @Test
    private
    void
    testEnumDefGetEnumValue()
        throws Exception
    {
        state.isTrue( 
            ENUM1_DEF.getEnumValue( MingleIdentifier.create( "x" ) ) == null );
        
        ModelTestInstances.assertEqual(
            MingleEnum.create(
                AtomicTypeReference.create( ENUM1_DEF.getName() ),
                MingleIdentifier.create( "e2" )
            ),
            ENUM1_DEF.getEnumValue( MingleIdentifier.create( "e2" ) )
        );
    }

    private
    final
    class ExchangerTest
    extends AbstractExchangerTest< ExchangerTest >
    {
        private ExchangerTest( CharSequence lbl ) { super( lbl ); }

        private
        void
        assertEquals( MingleValidationException mve1,
                      MingleValidationException mve2 )
        {
            state.equalString( mve1.getMessage(), mve2.getMessage() );
    
            state.equalString(
                format( mve1.getLocation() ), format( mve2.getLocation() ) );
        }
    
        private
        void
        assertEquals( MingleTypeCastException e1,
                      MingleTypeCastException e2 )
        {
            state.equal( e1.getExpectedType(), e2.getExpectedType() );
            state.equal( e1.getActualType(), e2.getActualType() );
    
            state.equalString(
                format( e1.getLocation() ), format( e2.getLocation() ) );
        }
    
        protected
        void
        assertExchange( Object jvObj1,
                        Object jvObj2 )
        {
            Class< ? > cls = jvObj1.getClass();

            if ( cls.equals( MingleValidationException.class ) )
            {
                assertEquals( 
                    (MingleValidationException) jvObj1,
                    (MingleValidationException) jvObj2 );
            }
            else if ( cls.equals( MingleTypeCastException.class ) )
            {
                assertEquals(
                    (MingleTypeCastException) jvObj1,
                    (MingleTypeCastException) jvObj2 );
            }
            else if ( MingleTypeReference.class.isAssignableFrom( cls ) )
            {
                state.equal( jvObj1, jvObj2 );
            }
            else if ( cls.equals( Enum1.class ) ||
                      cls.equals( ParsedString.class ) ||
                      cls.equals( MingleIdentifier.class ) ||
                      cls.equals( MingleNamespace.class ) ||
                      cls.equals( QualifiedTypeName.class ) ||
                      cls.equals( MingleIdentifiedName.class ) )
            {
                state.equal( jvObj1, jvObj2 );
            }
            else state.fail( "Unhandled type:", cls );
        }
    }

    private 
    static 
    enum Enum1 
    { 
        VAL1, VAL2; 

        final static AtomicTypeReference TYPE =
            (AtomicTypeReference) typeRef( "ns1@v1/Enum1" );
    }

    private
    final
    static
    class ParsedString
    extends TypedString< ParsedString >
    {
        private ParsedString( CharSequence cs ) { super( cs ); }
    }

    private
    final
    static
    class ParsedStringExchanger
    extends AbstractParsedStringExchanger< ParsedString >
    {
        private final static MingleString BAD_MG_VAL =
            MingleModels.asMingleString( "bad-mg-val" );

        private
        ParsedStringExchanger()
        {
            super( 
                (AtomicTypeReference) typeRef( "mingle:model@v1/ParsedString" ),
                ParsedString.class
            );
        }

        protected
        ParsedString
        parse( CharSequence cs )
            throws SyntaxException
        {
            if ( cs.toString().equals( BAD_MG_VAL.toString() ) )
            {
                throw 
                    new SyntaxException(
                        "bad-sentinel",
                        SourceTextLocation.create( "<>", 1, 12 ) );
            }
            else return new ParsedString( cs );
        }

        protected CharSequence asString( ParsedString s ) { return s; }
    }

    @InvocationFactory
    private
    List< ExchangerTest >
    testExchanger()
    {
        return Lang.asList(
            
            new ExchangerTest( "MingleValidationException" ).
                setExchanger( 
                    MingleModels.exchangerFor(
                        MingleModels.TYPE_REF_VALIDATION_EXCEPTION )
                ).
                setJvObj( new MingleValidationException( "test", OBJ_PATH2 ) ).
                setMgVal(
                    MingleModels.exceptionBuilder().
                        setType( "mingle:core@v1/ValidationException" ).
                        f().setString( "message", "test" ).
                        f().setString( "location", format( OBJ_PATH2 ) ).
                        build()
                ),
            
            new ExchangerTest( "MingleTypeCastException" ).
                setExchanger(
                    MingleModels.exchangerFor(
                        MingleModels.TYPE_REF_TYPE_CAST_EXCEPTION )
                ).
                setJvObj(
                    new MingleTypeCastException(
                        typeRef( "ns1@v1/S1" ), 
                        typeRef( "ns2@v1/S2" ), 
                        OBJ_PATH2 
                    ) 
                ).
                setMgVal(
                    MingleModels.exceptionBuilder().
                        setType( "mingle:core@v1/TypeCastException" ).
                        f().setString( "expected-type", "ns1@v1/S1" ).
                        f().setString( "actual-type", "ns2@v1/S2" ).
                        f().setString( "location", format( OBJ_PATH2 ) ).
                        build()
                ),
            
            new ExchangerTest( "expanded-enum" ).
                setExchanger( 
                    MingleModels.createExchanger( 
                        Enum1.TYPE, Enum1.class ) ).
                setJvObj( Enum1.VAL1 ).
                setMgVal(
                    new MingleEnum.Builder().
                        setType( "ns1@v1/Enum1" ).
                        setValue( "val1" ).
                        build()
                ),
            
            new ExchangerTest( "string-parse-exchanger-success" ).
                setExchanger( new ParsedStringExchanger() ).
                setJvObj( new ParsedString( "hello" ) ).
                setMgVal( MingleModels.asMingleString( "hello" ) ),
            
            (ExchangerTest) 
            new ExchangerTest( "string-parse-exchanger-failure" ).
                setExchanger( new ParsedStringExchanger() ).
                setMgVal( ParsedStringExchanger.BAD_MG_VAL ).
                expectFailure( 
                    MingleValidationException.class,
                    "\\Q[col 12]: bad-sentinel\\E"
                ),
            
            new ExchangerTest( "mingle-ident-ok" ).
                setExchanger( 
                    MingleModels.exchangerFor(
                        MingleModels.TYPE_REF_MINGLE_IDENTIFIER ) ).
                setMgVal( MingleModels.asMingleString( "hello-there" ) ).
                setJvObj( MingleIdentifier.create( "helloThere" ) ),
            
            (ExchangerTest)
            new ExchangerTest( "mingle-ident-fail-parse" ).
                setExchanger(
                    MingleModels.exchangerFor(
                        MingleModels.TYPE_REF_MINGLE_IDENTIFIER ) ).
                setMgVal( MingleModels.asMingleString( "hello-929" ) ).
                expectFailure(
                    MingleValidationException.class,
                    "\\Q[col 7]: Invalid part beginning: 9\\E"
                ),
            
            // for the rest of these not exhaustively testing error/success;
            // just testing success as a way to make sure that exchangers are
            // correctly installed and operational
            new ExchangerTest( "mingle-namespace-ok" ).
                setExchanger(
                    MingleModels.exchangerFor(
                        MingleModels.TYPE_REF_MINGLE_NAMESPACE ) ).
                setMgVal( MingleModels.asMingleString( "ns1@v1" ) ).
                setJvObj( MingleNamespace.create( "ns1@v1" ) ),
            
            new ExchangerTest( "qualified-type-name-ok" ).
                setExchanger(
                    MingleModels.exchangerFor(
                        MingleModels.TYPE_REF_QUALIFIED_TYPE_NAME ) ).
                setMgVal( MingleModels.asMingleString( "ns1@v1/N1/N2" ) ).
                setJvObj( QualifiedTypeName.create( "ns1@v1/N1/N2" ) ),
            
            new ExchangerTest( "type-ref-ok" ).
                setExchanger(
                    MingleModels.exchangerFor(
                        MingleModels.TYPE_REF_MINGLE_TYPE_REFERENCE ) ).
                setMgVal( MingleModels.asMingleString( "ns1@v1/T*+?" ) ).
                setJvObj( MingleTypeReference.create( "ns1@v1/T*+?" ) ),
            
            new ExchangerTest( "identified-name-ok" ).
                setExchanger(
                    MingleModels.exchangerFor(
                        MingleModels.TYPE_REF_MINGLE_IDENTIFIED_NAME ) ).
                setMgVal( MingleModels.asMingleString( "ns1:ns2@v1/a1/a2" ) ).
                setJvObj( MingleIdentifiedName.create( "ns1:ns2@v1/a1/a2" ) )
        );
    }

    @Test
    private
    void
    testEnumExchangerStringInput()
    {
        state.equal(
            Enum1.VAL1,
            MingleModels.
                createExchanger( Enum1.TYPE, Enum1.class ).
                asJavaValue( MingleModels.asMingleString( "val1" ), OBJ_PATH3 )
        );
    }

    @Test( expected = MingleValidationException.class,
           expectedPattern = "val: Invalid enum value: not-a-val" )
    private
    void
    testEnumExchangerExpandedFailBadConstant()
    {
        MingleModels.
            createExchanger( Enum1.TYPE, Enum1.class ).
            asJavaValue(
                new MingleEnum.Builder().
                    setType( "ns1@v1/Enum1" ).
                    setValue( "not-a-val" ).
                    build(),
                OBJ_PATH3
            );
    }

    @Test( expected = MingleValidationException.class,
           expectedPattern = "val: Invalid enum value: not-a-val" )
    private
    void
    testEnumExchangerStringifiedFailBadConstant()
    {
        MingleModels.createExchanger( Enum1.TYPE, Enum1.class ).
            asJavaValue(
                MingleModels.asMingleString( "not-a-val" ),
                OBJ_PATH3
            );
    }

    @Test( expected = MingleValidationException.class,
           expectedPattern = 
            "val: Expected mingle value of type ns1@v1/Enum1 " +
            "but found ns2@v1/Blah" )
    private
    void
    testEnumExchangerFailBadType()
    {
        MingleModels.
            createExchanger( Enum1.TYPE, Enum1.class ).
            asJavaValue(
                new MingleEnum.Builder().
                    setType( "ns2@v1/Blah" ).
                    setValue( "val1" ).
                    build(),
                OBJ_PATH3
            );
    }

    @Test
    private
    void
    testAsMingleListShallowSuccess()
    {
        MingleList l = 
            MingleModels.asMingleListInstance( 
                (ListTypeReference) typeRef( "ns1@v1/Something*" ), 
                MingleList.create( MingleModels.asMingleString( "s1" ) ),
                true, 
                OBJ_PATH2 
            );
 
        Iterator< MingleValue > it = l.iterator();
        state.equalString( "s1", (MingleString) it.next() );
        state.isFalse( it.hasNext() );
    }

    @Test
    private
    void
    testAsStructureFieldsExplicitType()
    {
        state.equal(
            MingleModels.asMingleString( "hello" ),
            MingleModels.asStructureFields(
                MingleModels.structBuilder().
                    setType( "ns1@v1/Type1" ).
                    f().setString( "f1", "hello" ).
                    build(),
                typeRef( "ns1@v1/Type1" ),
                OBJ_PATH3
            ).
            get( ID_F1 )
        );
    }

    @Test
    private
    void
    testAsStructureFieldsImplicitType()
    {
        state.equal(
            MingleModels.asMingleString( "hello" ),
            MingleModels.asStructureFields(
                MingleModels.symbolMapBuilder().
                    setString( "f1", "hello" ).
                    build(),
                typeRef( "ns1@v1/Type1" ),
                OBJ_PATH3
            ).
            get( ID_F1 )
        );
    }
        
    @Test( expected = MingleTypeCastException.class,
           expectedPattern =
            "\\Qval: Expected mingle value of type ns1@v1/Type1 " +
                "but found ns1@v1/Type2\\E" )
    private
    void
    testAsStructureFieldsBadExplicitType()
    {
        MingleModels.asStructureFields(
            MingleModels.structBuilder().
                setType( "ns1@v1/Type2" ).
                build(),
            typeRef( "ns1@v1/Type1" ),
            OBJ_PATH3
        );
    }

    @Test( expected = MingleTypeCastException.class,
           expectedPattern =
            "\\Qval: Expected mingle value of type ns1@v1/Type1 " +
                "but found mingle:core@v1/Int64\\E" )
    private
    void
    testAsStructureFieldsBadStructureType()
    {
        MingleModels.asStructureFields(
            MingleModels.asMingleInt64( 1L ),
            typeRef( "ns1@v1/Type1" ),
            OBJ_PATH3
        );
    }

    @Test( expected = MingleValidationException.class,
           expectedPattern = "val: Unexpected type: ns1@v1/S1" )
    private
    void
    testExpectStructErrorMessage()
    {
        MingleModels.expectStruct(
            MingleModels.structBuilder().
                setType( "ns1@v1/S1" ).
                build(),
            OBJ_PATH3,
            typeRef( "ns1@v1/S2" )
        );
    }

    private
    void
    assertIdentifiedNameFormat( CharSequence nmStr,
                                CharSequence pathExpct )
    {
        MingleIdentifiedName nm = MingleIdentifiedName.create( nmStr );

        state.equalString( pathExpct, MingleModels.filePathFor( nm ) );

        // also get explicit coverage of the sb append version used by the
        // default method internally
        state.equalString(
            pathExpct, 
            MingleModels.appendFilePathFor( nm, new StringBuilder() ) );
    }

    @Test
    private
    void
    testMingleIdentifiedNameFormats()
    {
        assertIdentifiedNameFormat( "ns1@v1/id1", "ns1.v1/id1" );
        
        assertIdentifiedNameFormat(
            "ns1:ns2@v1/id1/id2", "ns1/ns2.v1/id1/id2" );
    }

    @Test
    private
    void
    testTypeDefintitionLookupImplSuccess()
    {
        QualifiedTypeName qn = QualifiedTypeName.create( "ns1@v1/S1" );
        StructDefinition sd = (StructDefinition) LOOKUP1.expectType( qn );
        state.equal( qn, sd.getName() );
    }

    @Test( expected = NoSuchTypeDefinitionException.class,
           expectedPattern = "\\Qns1@v1/Blah\\E" )
    private
    void
    testTypeDefinitionLookupImplFailure()
    {
        LOOKUP1.expectType( QualifiedTypeName.create( "ns1@v1/Blah" ) );
    }

    @Test
    private
    void
    testAsMingleStructure()
    {
        AtomicTypeReference typ = (AtomicTypeReference) typeRef( "ns1@v1/T" );

        MingleSymbolMap m = MingleModels.symbolMapBuilder().build();

        MingleStruct ms = MingleModels.asMingleStruct( typ, m );
        state.isTrue( ms.getType() == typ );
        state.isTrue( ms.getFields() == m );

        MingleException me = MingleModels.asMingleException( typ, m );
        state.isTrue( me.getType() == typ );
        state.isTrue( me.getFields() == m );
    }

    @Test
    private
    void
    testMingleSymbolMapBuilderSetAll()
    {
        MingleSymbolMap m1 =
            MingleModels.symbolMapBuilder().
                setString( "f1", "v1" ).
                setString( "f2", "v1" ).
                setString( "f3", "v1" ).
                build();
        
        MingleSymbolMap m2 =
            MingleModels.symbolMapBuilder().
                setString( "f1", "v2" ).
                setString( "f4", "v2" ).
                setAll( m1 ).
                setString( "f2", "v2" ).
                build();
        
        MingleSymbolMapAccessor acc = MingleSymbolMapAccessor.create( m2 );
        state.equalString( "v1", acc.expectString( "f1" ) );
        state.equalString( "v2", acc.expectString( "f2" ) );
        state.equalString( "v1", acc.expectString( "f3" ) );
        state.equalString( "v2", acc.expectString( "f4" ) );
    }

    private
    final
    class IsAssignableTest
    extends LabeledTestCall
    {
        private final TypeDefinitionLookup lk;
        private final MingleTypeReference lhs;
        private final MingleTypeReference rhs;
        private final boolean expct;

        private
        IsAssignableTest( TypeDefinitionLookup lk,
                          CharSequence lhs,
                          CharSequence rhs,
                          boolean expct )
        {
            super( Strings.crossJoin( "=", ",", "lhs", lhs, "rhs", rhs ) );

            this.lk = lk;
            this.lhs = typeRef( lhs );
            this.rhs = typeRef( rhs );
            this.expct = expct;
        }

        public
        void
        call()
            throws Exception
        {
            state.equal( expct, MingleModels.isAssignable( lhs, rhs, lk ) );
        }
    }

    private
    void
    addIsAssignableBaseTests( List< IsAssignableTest > res )
    {
        for ( String[] pair : 
                new String[][] {
                    new String[] { "S1", "S1" },
                    new String[] { "S1", "S2" },
                    new String[] { "S1", "S3" },
                    new String[] { "S2", "S3" },
                    new String[] { "S2", "S4" },
                    new String[] { "S1", "T1" }
                } )
        {
            String lhs = pair[ 0 ];
            String rhs = pair[ 1 ];

            boolean expct = 
                ! ( rhs.equals( "T1" ) || 
                    ( rhs.equals( "S4" ) && ( ! lhs.equals( "S1" ) ) ) );
            
            lhs = "ns1@v1/" + lhs;
            rhs = "ns1@v1/" + rhs;

            res.addAll(
                Lang.asList(
                    new IsAssignableTest( LOOKUP1, lhs, rhs, expct ),
                    new IsAssignableTest( LOOKUP1, lhs + "?", rhs, expct ),

                    new IsAssignableTest( 
                        LOOKUP1, lhs + "?", rhs + "?", expct ),
                    
                    new IsAssignableTest( 
                        LOOKUP1, lhs + "*", rhs + "*", expct ),

                    new IsAssignableTest(
                        LOOKUP1, lhs + "*", rhs + "+", expct ),

                    new IsAssignableTest(
                        LOOKUP1, lhs + "+", rhs + "+", expct ),

                    new IsAssignableTest( LOOKUP1, lhs, rhs + "*", false ),
                    new IsAssignableTest( LOOKUP1, lhs, rhs + "+", false ),
                    new IsAssignableTest( LOOKUP1, lhs, rhs + "?", false ),

                    new IsAssignableTest(
                        LOOKUP1, lhs + "+", rhs + "*", false ),

                    new IsAssignableTest(
                        LOOKUP1, lhs + "*", rhs + "?*", false ),

                    new IsAssignableTest(
                        LOOKUP1, lhs + "+", rhs + "?*", false ),

                    new IsAssignableTest(
                        LOOKUP1, lhs + "+", rhs + "?+", false ),

                    new IsAssignableTest(
                        LOOKUP1, lhs + "*", rhs + "*?", false ),

                    new IsAssignableTest(
                        LOOKUP1, lhs + "*", rhs + "**", false ),

                    new IsAssignableTest(
                        LOOKUP1, lhs + "*", rhs + "*+", false )

                )
            );
        }
    }

    private
    void
    addIsAssignableMingleValueTests( List< IsAssignableTest > res )
    {
        for ( String s : new String[] { "", "*", "?", "+", "**", "*+", "+*?" } )
        {
            String rhs = "ns1@v1/S1" + s;

            res.add(
                new IsAssignableTest( 
                    LOOKUP1, "mingle:core@v1/Value" + s, rhs, true ) );
        }

        res.addAll( Lang.< IsAssignableTest >asList( 

            new IsAssignableTest(
                LOOKUP1, "mingle:core@v1/Value*", "ns1@v1/S1+", true ),
            
            new IsAssignableTest(
                LOOKUP1, "mingle:core@v1/Value+", "ns1@v1/S3+", true ),
            
            new IsAssignableTest(
                LOOKUP1, "mingle:core@v1/Value+", "ns1@v1/S3+", true ),

            new IsAssignableTest( LOOKUP1,
                "mingle:core@v1/Value*", "ns1@v1/S3+*", true ),
            
            new IsAssignableTest( LOOKUP1,
                "mingle:core@v1/Value", "ns1@v1/S4?", false ),
            
            new IsAssignableTest( LOOKUP1,
                "mingle:core@v1/Value*", "ns1@v1/S1", false ),
            
            new IsAssignableTest( LOOKUP1,
                "mingle:core@v1/Value+", "ns1@v1/S1", false ),
            
            new IsAssignableTest( LOOKUP1,
                "mingle:core@v1/Value+", "ns1@v1/S1*", false )
        
        ));
    }

    @InvocationFactory
    private
    List< IsAssignableTest >
    testIsAssignable()
    {
        List< IsAssignableTest > res = Lang.newList();

        addIsAssignableBaseTests( res );
        addIsAssignableMingleValueTests( res );

        res.addAll( Lang.< IsAssignableTest >asList(
            
            new IsAssignableTest( LOOKUP1, "ns1@v1/S1", "ns1@v1/S1+?", false ),
            new IsAssignableTest( LOOKUP1, "ns1@v1/S1", "ns1@v1/S1+**", false ),

            new IsAssignableTest( LOOKUP1, "ns1@v1/S1?", "ns1@v1/S1*", false ),
            new IsAssignableTest( LOOKUP1, "ns1@v1/S1?", "ns1@v1/S1+", false ),
            new IsAssignableTest( LOOKUP1, "ns1@v1/S1?", "ns1@v1/S3+", false ),
            new IsAssignableTest( LOOKUP1, "ns1@v1/S1?", "ns1@v1/S3*", false ),
            new IsAssignableTest( LOOKUP1, "ns1@v1/S1??", "ns1@v1/S1??", true ),
            new IsAssignableTest( LOOKUP1, "ns1@v1/S1??", "ns1@v1/S2??", true ),
            new IsAssignableTest( LOOKUP1, "ns1@v1/S1?", "ns1@v1/S2??", false ),
            new IsAssignableTest( LOOKUP1, "ns1@v1/S1??", "ns1@v1/S2?", true ),
            new IsAssignableTest( LOOKUP1, "ns1@v1/S1?", "ns1@v1/S1*?", false ),
            new IsAssignableTest( LOOKUP1, "ns1@v1/S1?", "ns1@v1/S1+?", false ),
            new IsAssignableTest( LOOKUP1, "ns1@v1/S1?", "ns1@v1/S2*?", false ),
            new IsAssignableTest( LOOKUP1, "ns1@v1/S1?", "ns1@v1/S2+?", false )
        ));

        return res;
    }

    private
    static
    void
    putLookup1StructType( TypeDefinitionLookup.Builder lkBld,
                          CharSequence typ,
                          CharSequence sprTyp )
    {
        StructDefinition.Builder b = new StructDefinition.Builder();
        b.setFields( FieldSet.getEmptyFieldSet() );

        QualifiedTypeName qn = QualifiedTypeName.create( typ );
        b.setName( qn );

        if ( sprTyp != null ) b.setSuperType( typeRef( sprTyp ) );

        lkBld.addType( b.build() );
    }

    static
    {
        TypeDefinitionLookup.Builder b = new TypeDefinitionLookup.Builder();

        putLookup1StructType( b, "ns1@v1/S1", null );
        putLookup1StructType( b, "ns1@v1/S2", "ns1@v1/S1" );
        putLookup1StructType( b, "ns1@v1/S3", "ns1@v1/S2" );
        putLookup1StructType( b, "ns1@v1/S4", "ns1@v1/S1" );
        putLookup1StructType( b, "ns1@v1/T1", null );

        LOOKUP1 = b.build();
    }
}
