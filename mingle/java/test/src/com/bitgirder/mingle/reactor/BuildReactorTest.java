package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import static com.bitgirder.mingle.MingleTestMethods.*;

import com.bitgirder.mingle.Mingle;
import com.bitgirder.mingle.MingleString;
import com.bitgirder.mingle.MingleInt32;
import com.bitgirder.mingle.MingleStruct;
import com.bitgirder.mingle.MingleSymbolMap;
import com.bitgirder.mingle.QualifiedTypeName;
import com.bitgirder.mingle.MingleIdentifier;
import com.bitgirder.mingle.MingleTypeReference;
import com.bitgirder.mingle.ListTypeReference;
import com.bitgirder.mingle.MingleValue;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPaths;

import java.util.List;
import java.util.Map;

// Used in ReactorTests, but pulled out here into its own source file given how
// large the test and support code is
final
class BuildReactorTest
extends AbstractReactorTest
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final static MingleIdentifier ERR_FIELD = id( "bad-field" );

    private final static MingleValue PLACEHOLDER_VAL =
        new MingleString( "placeholder-val" );

    private final static String MSG_ERR_BAD_VAL =
        "test-message-error-bad-value";
    
    private final static QualifiedTypeName ERR_QNAME =
        qname( "ns1@v1/BadType" );

    private final static MingleTypeReference ERR_TYP_NEXT_FACTORY =
        listType( atomic( "ns1@v1/NextBuilderNilTest" ), true );

    private final static ListTypeReference INT32_LIST_TYPE =
        listType( Mingle.TYPE_INT32, true );

    private final static QualifiedTypeName QNAME_TEST_STRUCT1 =
        qname( "mingle:reactor@v1/TestStruct1" );

    private final static QualifiedTypeName QNAME_TEST_STRUCT2 =
        qname( "mingle:reactor@v1/TestStruct2" );

    private final static MingleInt32 ERR_TEST_VAL = new MingleInt32( 100 );

    Object val;
    Object source;
    String profile;

    BuildReactorTest() {}

    private
    static
    TestException
    testExceptionForPath( ObjectPath< MingleIdentifier > path )
    {
        return new TestException( path, MSG_ERR_BAD_VAL );
    }

    private
    static
    MingleValue
    errorResultForValue( MingleValue mv,
                         ObjectPath< MingleIdentifier > path )
        throws Exception
    {
        if ( ! mv.equals( ERR_TEST_VAL ) ) return mv;
        throw testExceptionForPath( path );
    }

    private
    final
    static
    class TestExceptionFactory
    implements ExceptionFactory
    {
        public
        Exception
        createException( ObjectPath< MingleIdentifier > path,
                         String msg )
        {
            return new TestException( path, msg );
        }
    }

    private
    final
    static
    class ErrorProfileFieldSetBuilder
    extends AbstractFieldSetBuilder
    {
        @Override
        public
        BuildReactor.Factory
        startField( MingleIdentifier fld,
                    ObjectPath< MingleIdentifier > path )
        {
            if ( fld.equals( ERR_FIELD ) ) {
                throw testExceptionForPath( ObjectPaths.parentOf( path ) );
            }

            return new ErrorProfileFactory();
        }

        @Override
        protected Object produceValue() { return PLACEHOLDER_VAL; }
    }

    private
    final
    static
    class ErrorProfileListBuilder
    extends AbstractListBuilder
    {
        private final ListTypeReference lt;

        private 
        ErrorProfileListBuilder( ListTypeReference lt ) 
        { 
            this.lt = lt; 
        }

        public
        BuildReactor.Factory
        nextFactory()
        {
            if ( lt.equals( ERR_TYP_NEXT_FACTORY ) ) return null;
            return new ErrorProfileFactory();
        }

        @Override
        public
        void
        addValue( Object val,
                  ObjectPath< MingleIdentifier > path )
            throws Exception
        {
            errorResultForValue( (MingleValue) val, path );
        }

        @Override
        protected Object produceValue() { return PLACEHOLDER_VAL; }
    }

    private
    final
    static
    class ErrorProfileFactory
    extends AbstractBuildFactory
    {
        @Override
        public
        Object
        buildValue( MingleValue mv,
                    ObjectPath< MingleIdentifier > path )
            throws Exception
        {
            return errorResultForValue( mv, path );
        }

        @Override
        protected
        BuildReactor.FieldSetBuilder
        startMap()
        {
            return new ErrorProfileFieldSetBuilder();
        }

        @Override
        public
        BuildReactor.FieldSetBuilder
        startStruct( QualifiedTypeName qn,
                     ObjectPath< MingleIdentifier > path )
        {
            if ( qn.equals( ERR_QNAME ) ) {
                throw testExceptionForPath( path );
            }

            return new ErrorProfileFieldSetBuilder();
        }

        @Override
        public
        BuildReactor.ListBuilder
        startList( ListTypeReference lt,
                   ObjectPath< MingleIdentifier > path )
        {
            if ( Mingle.typeNameIn( lt ).equals( ERR_QNAME ) ) {
                throw testExceptionForPath( path );
            }

            return new ErrorProfileListBuilder( lt );
        }
    }

    private
    static
    BuildReactor.Factory
    implStartField( MingleIdentifier fld )
    {
        List< MingleIdentifier > acpt =
            Lang.asList( id( "f1" ), id( "f2" ), id( "f3" ) );

        return acpt.contains( fld ) ? new ImplFactory() : null; 
    }

    private
    final
    static
    class ImplMapBuilder
    extends AbstractFieldSetBuilder
    {
        private final Map< String, Object > res = Lang.newMap();

        @Override protected Object produceValue() { return res; }

        @Override
        protected
        BuildReactor.Factory
        startField( MingleIdentifier fld )
        {
            return implStartField( fld );
        }

        @Override
        protected
        void
        setValue( MingleIdentifier fld,
                  Object val )
        {
            res.put( fld.getExternalForm(), val );
        }
    }

    private
    static
    class ImplListBuilder
    extends AbstractListBuilder
    {
        private final List< Object > l = Lang.newList();

        Integer maxLen = Integer.MAX_VALUE;

        protected Object convert( List< Object > l ) { return l; }

        @Override protected Object produceValue() { return convert( l ); }

        public BuildReactor.Factory nextFactory() { return new ImplFactory(); }

        @Override 
        public
        void 
        addValue( Object val,
                  ObjectPath< MingleIdentifier > path ) 
        { 
            if ( l.size() == maxLen ) throw testExceptionForPath( path );
            
            l.add( val );
        }
    }

    private
    static
    class Int32ListBuilder
    extends ImplListBuilder
    {
        @Override
        protected
        Object
        convert( List< Object > l )
        {
            int[] res = new int[ l.size() ];

            int i = 0;
            for ( Object o : l ) res[ i++ ] = (Integer) o;

            return new Int32List( res );
        }
    }

    private
    static
    abstract
    class TestStructBuilder< T >
    extends AbstractFieldSetBuilder
    {
        final T res; 

        TestStructBuilder( T res ) { this.res = res; }

        @Override protected Object produceValue() { return res; }

        @Override
        protected
        BuildReactor.Factory
        startField( MingleIdentifier fld )
        {
            return implStartField( fld );
        }
    }

    private
    final
    static
    class TestStruct1Builder
    extends TestStructBuilder< TestStruct1 >
    {
        private TestStruct1Builder() { super( new TestStruct1() ); }

        @Override
        public
        void
        setValue( MingleIdentifier fld,
                  Object val,
                  ObjectPath< MingleIdentifier > path )
        {
            if ( fld.equals( id( "f1" ) ) ) {
                res.f1 = (Integer) val;
            } else if ( fld.equals( id( "f2" ) ) ) {
                res.f2 = (Int32List) val;
            } else if ( fld.equals( id( "f3" ) ) ) {
                res.f3 = (TestStruct1) val;
            } else {
                state.failf( "unhandled field: %s", fld );
            }
        }
    }

    private
    final
    static
    class TestStruct2Builder
    extends TestStructBuilder< TestStruct2 >
    {
        private TestStruct2Builder() { super( new TestStruct2() ); }

        @Override
        protected
        void
        setValue( MingleIdentifier fld,
                  Object val )
        {}
    }

    private
    final
    static
    class ImplFactory
    extends AbstractBuildFactory
    {
        @Override
        protected
        BuildReactor.FieldSetBuilder
        startMap()
        {
            return new ImplMapBuilder();
        }

        @Override
        public
        BuildReactor.FieldSetBuilder
        startStruct( QualifiedTypeName qn,
                     ObjectPath< MingleIdentifier > path )
        {
            if ( qn.equals( QNAME_TEST_STRUCT1 ) ) {
                return new TestStruct1Builder();
            } else if ( qn.equals( QNAME_TEST_STRUCT2 ) ) {
                return new TestStruct2Builder();
            } else {
                return null;
            }
        }

        @Override
        protected
        BuildReactor.ListBuilder
        startList( ListTypeReference lt )
        {
            if ( lt.equals( INT32_LIST_TYPE ) ) {
                Int32ListBuilder res = new Int32ListBuilder();
                res.maxLen = 4;
                return res;
            }

            return null;
        }

        @Override
        public
        Object
        buildValue( MingleValue mv,
                    ObjectPath< MingleIdentifier > path )
        {
            if ( mv instanceof MingleInt32 ) {
                int i = ( (MingleInt32) mv ).intValue();
                if ( i < 0 ) throw testExceptionForPath( path );
                return i;
            } else if ( mv instanceof MingleString ) {
                return mv.toString();
            } else {
                return BuildReactor.UNHANDLED_VALUE_MARKER;
            }
        }
    }

    private
    final
    static
    class ImplFailOnlyFactory
    extends AbstractBuildFactory
    {
        @Override
        public
        BuildReactor.FieldSetBuilder
        startMap( ObjectPath< MingleIdentifier > path )
        {
            throw testExceptionForPath( path );
        }
    }

    Object getSource() { return source == null ? (MingleValue) val : source; }

    private
    BuildReactor.Factory
    buildFactory()
    {
        if ( profile.equals( "default" ) ) {
            return new ValueBuildFactory();
        } else if ( profile.equals( "error" ) ) {
            return new ErrorProfileFactory();
        } else if ( profile.equals( "impl" ) ) {
            return new ImplFactory();
        } else if ( profile.equals( "impl-fail-only" ) ) {
            return new ImplFailOnlyFactory();
        } else {
            throw state.failf( "unhandled profile: %s", profile );
        }
    }

    private
    BuildReactor
    buildReactor()
    {
        return new BuildReactor.Builder().
            setFactory( buildFactory() ).
            setExceptionFactory( new TestExceptionFactory() ).
            build();
    }

    private
    void
    assertBuild( BuildReactor br )
    {
        if ( profile.equals( "impl" ) ) {
            state.equal( val, br.value() );
        } else {
            assertEqual( (MingleValue) val, (MingleValue) br.value() );
        }
    }

    public
    void
    call()
        throws Exception
    {
        BuildReactor br = buildReactor();

        MingleReactorPipeline pip = new MingleReactorPipeline.Builder().
            addReactor( MingleReactors.createDebugReactor() ).
            addReactor( br ).
            build();

        feedSource( getSource(), pip );
 
        if ( expectedFailureClass() == null ) assertBuild( br );
    }
}
