package com.bitgirder.mingle;

import static com.bitgirder.mingle.MingleTestMethods.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.TypedString;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPaths;

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

    private
    static
    enum TestType
    {
        IDENTIFIER,
        NAMESPACE,
        DECLARED_TYPE_NAME,
        QUALIFIED_TYPE_NAME,
        IDENTIFIER_PATH;
    }

    private final static Map< ErrorOverrideKey, Object > ERR_OVERRIDES =
        Lang.newMap( ErrorOverrideKey.class, Object.class,
            
            errMsgKey( TestType.IDENTIFIER, "trailing-input/x" ),
                "Unexpected trailing data \"/\" (U+002F)",
            
            errMsgKey( TestType.IDENTIFIER, "giving-mixedMessages" ),
                "Unexpected identifier character: \"M\" (U+004D)",
            
            errMsgKey( TestType.IDENTIFIER, "a-bad-ch@r" ),
                "Unexpected trailing data \"@\" (U+0040)",

            errMsgKey( TestType.IDENTIFIER_PATH, "i1[ xx ]" ),
                "Expected path index but found: xx",

            errMsgKey( TestType.IDENTIFIER_PATH, "#stuff" ),
                "Unrecognized token start: \"#\" (U+0023)" ,

            errMsgKey( TestType.IDENTIFIER_PATH, "bad$Id" ),
                "Unexpected identifier character: \"$\" (U+0024)" ,

            errMsgKey( TestType.IDENTIFIER_PATH, "[]" ),
                "Expected path index but found: ]",

            errMsgKey( TestType.IDENTIFIER_PATH, "i1[ ]" ),
                "Expected path index but found: ]",

            errMsgKey( TestType.IDENTIFIER_PATH, "i1[ 1.0 ]" ),
                "invalid decimal index",

            errMsgKey( TestType.IDENTIFIER_PATH, "i1[ 1.1e1 ]" ),
                "invalid decimal index",

            errMsgKey( TestType.IDENTIFIER_PATH, "i1[ 1e1 ]" ),
                "invalid decimal index",

            errMsgKey( TestType.IDENTIFIER_PATH, "[ 1 ].bad$Id" ),
                "Unexpected identifier character: \"$\" (U+0024)",
            
            errMsgKey( TestType.DECLARED_TYPE_NAME, "Bad-Char" ),
                "Unexpected trailing data \"-\" (U+002D)",
            
            errMsgKey( TestType.NAMESPACE, "ns1:ns2@v1:ns3" ),
                "Unexpected trailing data \":\" (U+003A)",
            
            errMsgKey( TestType.NAMESPACE, "ns1:ns2@v1@v2" ),
                "Unexpected trailing data \"@\" (U+0040)",
            
            errMsgKey( TestType.NAMESPACE, "ns1:ns2@v1/Stuff" ),
                "Unexpected trailing data \"/\" (U+002F)",
            
            errMsgKey( TestType.NAMESPACE, "ns1.ns2@v1" ),
                "Expected : or @ but found: .",
            
            errMsgKey( TestType.NAMESPACE, "ns1:ns2" ),
                "Expected : or @ but found: END",
            
            errMsgKey( TestType.NAMESPACE, "ns1 : ns2:ns3@v1" ),
                "Unrecognized token start: \" \" (U+0020)",
            
            errMsgKey( TestType.QUALIFIED_TYPE_NAME, "ns1/T1" ),
                "Expected : or @ but found: /",
            
            errMsgKey( TestType.QUALIFIED_TYPE_NAME, "ns1@v1" ),
                "Expected / but found: END",
            
            errMsgKey( TestType.QUALIFIED_TYPE_NAME, "ns1@v1/T1/" ),
                "Unexpected trailing data \"/\" (U+002F)"
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
    final
    static
    class ParseErrorExpectation
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
        private ParseErrorExpectation errExpct;

        private
        void
        validate()
        {
            state.notNull( in, "in" );
            state.notNull( tt, "tt" );
        }

        public
        String
        getLabel()
        {
            return Strings.crossJoin( "=", ",",
                "in", Lang.getRfc4627String( in ),
                "tt", tt
            ).
            toString();
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
        Object
        doParse()
            throws Exception
        {
            switch ( tt ) {
            case IDENTIFIER: return MingleIdentifier.parse( in );
            case DECLARED_TYPE_NAME: return DeclaredTypeName.parse( in );
            case NAMESPACE: return MingleNamespace.parse( in );
            case QUALIFIED_TYPE_NAME: return QualifiedTypeName.parse( in );
            case IDENTIFIER_PATH: return MingleParser.parseIdentifierPath( in );
            }
            throw state.failf( "Unhandled test type: %s", tt );
        }

        private
        void
        assertEqualPaths( Object val )
        {
            ObjectPath< MingleIdentifier > pathExpct = 
                Lang.castUnchecked( expct );

            ObjectPath< MingleIdentifier > act = Lang.castUnchecked( val );

            state.isTruef( ObjectPaths.areEqual( pathExpct, act ),
                "expct != act: %s != %s",
                Mingle.formatIdPath( pathExpct ),
                Mingle.formatIdPath( act )
            );
        }

        private
        void
        assertExpectVal( Object val )
        {
            if ( errExpct == null ) {
                if ( tt == TestType.IDENTIFIER_PATH ) {
                    assertEqualPaths( val );
                } else {
                    state.equal( expct, val );
                }
            }
            else state.failf( "Got %s but expected error %s", val, errExpct );
        }

        private
        int
        expectErrCol( int defl )
        {
            Object override = override();

            if ( override == null || override instanceof String ) return defl;

            if ( override instanceof Integer ) return (Integer) override;

            if ( override instanceof ParseErrorExpectation ) {
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
            
            if ( override instanceof ParseErrorExpectation ) {
                return ( (ParseErrorExpectation) override ).msg;
            }

            throw state.createFail( "Unhandled override:", override );
        }

        private
        void
        assertFailure( Exception ex )
            throws Exception
        {
            if ( errExpct == null ) throw ex;
            if ( ! ( ex instanceof MingleSyntaxException ) ) throw ex;

            MingleSyntaxException mse = (MingleSyntaxException) ex;
            
            state.equalString( 
                expectErrString( errExpct.msg ), mse.getError() );

            state.equalInt( expectErrCol( errExpct.col ), mse.getColumn() );
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
        CoreParseTest
        createTest( MingleSymbolMap map )
        {
            String ttStr = mapExpect( map, "testType", String.class );
            ttStr = ttStr.replace( '-', '_' ).toUpperCase();
            
            // we prefer this to 'try { TestType.valueOf() } catch { ... } since
            // the 'catch' can only catch IllegalArgumentException, which could
            // be some error besides just an unmatched name, in which case we
            // can't tell whether to throw or suppress it.
            for ( TestType tt : TestType.class.getEnumConstants() ) {
                if ( tt.name().equals( ttStr ) ) {
                    CoreParseTest res = new CoreParseTest();
                    res.tt = TestType.valueOf( ttStr );
                    return res;
                }
            }

            return null;
        }

        private
        ParseErrorExpectation
        convertParseErrorExpect( MingleSymbolMap map )
        {
            return new ParseErrorExpectation(
                mapExpect( map, "col", Integer.class ),
                mapExpect( map, "message", String.class )
            );
        }

        private
        Object
        convertFromBuffer( byte[] buf,
                           String typ )
            throws Exception
        {
            MingleBinReader rd = MingleBinReader.create( buf );

            if ( typ.equals( "Identifier" ) ) {
                return rd.readIdentifier();
            } else if ( typ.equals( "Namespace" ) ) {
                return rd.readNamespace();
            } else if ( typ.equals( "DeclaredTypeName" ) ) {
                return rd.readDeclaredTypeName();
            } else if ( typ.equals( "QualifiedTypeName" ) ) {
                return rd.readQualifiedTypeName();
            }

            throw state.failf( "unhandled tt for buffer: %s", typ );
        }

        private
        Object
        convertFromBuffer( MingleSymbolMap map,
                           String typ )
            throws Exception
        {
            byte[] buf = mapExpect( map, "buffer", byte[].class );
            return convertFromBuffer( buf, typ );
        }

        private
        ObjectPath< MingleIdentifier >
        extendPath( ObjectPath< MingleIdentifier > path,
                    MingleValue mv )
            throws Exception
        {
            if ( mv instanceof MingleUint64 ) {
                int idx = ( (MingleUint64) mv ).intValue();
                return path.startImmutableList( idx );
            } 
            else if ( mv instanceof MingleBuffer ) 
            {
                byte[] buf = ( (MingleBuffer) mv ).array();

                MingleIdentifier id = (MingleIdentifier) 
                    convertFromBuffer( buf, "Identifier" );

                return path.descend( id );
            }
            else throw state.failf( "unhandled path elt: %s", mv.getClass() );
        }

        private
        ObjectPath< MingleIdentifier >
        convertIdPath( MingleSymbolMap map )
            throws Exception
        {
            MingleList ml = mapExpect( map, "path", MingleList.class );

            ObjectPath< MingleIdentifier > res = ObjectPath.getRoot();

            for ( MingleValue mv : ml ) res = extendPath( res, mv );

            return res;
        }

        private
        Object
        convertValue( MingleStruct ms )
            throws Exception
        {
            if ( ms == null ) return null;

            String typ = ms.getType().getName().getExternalForm().toString();
            MingleSymbolMap map = ms.getFields();

            if ( typ.equals( "ParseErrorExpect" ) ) {
                return convertParseErrorExpect( map );
            } 
            else if ( typ.equals( "Identifier" ) || typ.equals( "Namespace" ) ||
                      typ.equals( "DeclaredTypeName" ) ||
                      typ.equals( "QualifiedTypeName" ) ) 
            {
                return convertFromBuffer( map, typ );
            }
            else if ( typ.equals( "IdentifierPath" ) ) {
                return convertIdPath( map );
            }

            throw state.failf( "unhandled type: %s", typ );
        }

        public 
        CoreParseTest 
        convertStruct( MingleStruct ms ) 
            throws Exception
        {
            MingleSymbolMap map = ms.getFields();

            CoreParseTest res = createTest( map );
            if ( res == null ) return null;

            res.in = mapExpect( map, "in", String.class );

            res.extForm = mapExpect( map, "externalForm", String.class );

            res.expct = 
                convertValue( mapGet( map, "expect", MingleStruct.class ) );

            res.errExpct = (ParseErrorExpectation) 
                convertValue( mapGet( map, "error", MingleStruct.class ) );

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
