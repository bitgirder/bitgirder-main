package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.pipeline.Pipelines;

import com.bitgirder.test.Test;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.LabeledTestCall;

import java.util.List;
import java.util.Map;

import java.util.regex.Pattern;

@Test
final
class MingleReactorTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static DeclaredTypeName TYP_VALUE_BUILD_TEST =
        DeclaredTypeName.create( "ValueBuildTest" );

    private final static DeclaredTypeName TYP_STRUCTURAL_REACTOR_ERROR_TEST =
        DeclaredTypeName.create( "StructuralReactorErrorTest" );

    private final static Map< Object, Map< String, String > > STRING_OVERRIDES;

    private static void code( Object... args ) { CodeLoggers.code( args ); }

    private 
    static 
    void 
    codef( String tmpl, 
           Object... args ) 
    { 
        CodeLoggers.codef( tmpl, args ); 
    }

    private
    static
    String
    overridableString( Object overrideKey,
                       String strKey )
    {
        Map< String, String > m = STRING_OVERRIDES.get( overrideKey );

        if ( m == null || ( ! m.containsKey( strKey ) ) ) return strKey;
        return m.get( strKey );
    }

    private
    static
    abstract
    class TestImpl
    extends LabeledTestCall
    {
        private TestImpl( CharSequence nm ) { super( nm ); }
    }

    private
    static
    final
    class ValueBuildTest
    extends TestImpl
    {
        private MingleValue val;

        private ValueBuildTest( CharSequence nm ) { super( nm ); }

        public
        void
        call()
            throws Exception
        {
            MingleValueReactorPipeline pip = 
                MingleValueReactors.createValueBuilderPipeline();

            MingleValueReactors.visitValue( val, pip );

            MingleValueBuilder bld = 
                Pipelines.lastElementOfType( 
                    pip.pipeline(), MingleValueBuilder.class );
            
            MingleTests.assertEqual( val, bld.value() );
        }
    }

    private
    final
    static
    class StructuralErrorTest
    extends TestImpl
    {
        private List< MingleValueReactorEvent > events;
        private MingleValueReactorTopType topType;

        private StructuralErrorTest( CharSequence nm ) { super( nm ); }
        
        public
        void
        call()
            throws Exception
        {
            code( "expect err: " + expectedFailurePattern() );

            MingleValueStructuralCheck chk =
                MingleValueStructuralCheck.createWithTopType( topType );

            MingleValueReactorPipeline pip =
                new MingleValueReactorPipeline.Builder().
                    addReactor( MingleValueReactors.createDebugReactor() ).
                    addReactor( chk ).
                    build();

            for ( MingleValueReactorEvent ev : events ) {
                pip.processEvent( ev );
            }
        }
    }

    private
    static
    class TestImplReader
    extends MingleTestGen.StructFileReader< TestImpl >
    {
        private final Map< DeclaredTypeName, Integer > seqsByType =
            Lang.newMap();

        private
        TestImplReader()
        {
            super( "reactor-tests.bin" );
        }

        private
        CharSequence
        makeName( MingleStructAccessor testObj,
                  Object name )
        {
            DeclaredTypeName declNm = testObj.getType().getName();

            if ( name == null ) {
                Integer seq = seqsByType.get( declNm );
                if ( seq == null ) seq = Integer.valueOf( 0 );
                name = seq.toString();
                seqsByType.put( declNm, seq + 1 );
            }

            return declNm + "/" + name;
        }

        private
        void
        setEventStartStruct( MingleValueReactorEvent ev,
                             MingleStructAccessor acc )
            throws Exception
        {
            MingleBinReader rd = 
                MingleBinReader.create( acc.expectByteArray( "type" ) );

            ev.setStartStruct( rd.readQualifiedTypeName() );
        }

        private
        void
        setEventStartField( MingleValueReactorEvent ev,
                            MingleStructAccessor acc )
            throws Exception
        {
            MingleBinReader rd =
                MingleBinReader.create( acc.expectByteArray( "field" ) );

            ev.setStartField( rd.readIdentifier() );
        }

        private
        MingleValueReactorEvent
        convertReactorEvent( MingleStructAccessor acc )
            throws Exception
        {
            MingleValueReactorEvent res = new MingleValueReactorEvent();

            String evName = acc.getType().getName().toString();

            if ( evName.equals( "StructStartEvent" ) ) {
                setEventStartStruct( res, acc );
            } else if ( evName.equals( "FieldStartEvent" ) ) {
                setEventStartField( res, acc );
            } else if ( evName.equals( "MapStartEvent" ) ) {
                res.setStartMap();
            } else if ( evName.equals( "ListStartEvent" ) ) {
                res.setStartList();
            } else if ( evName.equals( "EndEvent" ) ) {
                res.setEnd();
            } else if ( evName.equals( "ValueEvent" ) ) {
                res.setValue( acc.expectMingleValue( "val" ) );
            } else {
                state.failf( "unhandled event: %s", evName );
            }

            return res;
        }

        private
        List< MingleValueReactorEvent >
        convertReactorEvents( MingleListAccessor acc )
            throws Exception
        {
            List< MingleValueReactorEvent > res = Lang.newList();
            
            MingleListAccessor.Traversal t = acc.traversal();
    
            while ( t.hasNext() ) {
                res.add( convertReactorEvent( t.nextStructAccessor() ) );
            }

            return res;
        }

        private
        ValueBuildTest
        convertValueBuildTest( MingleStructAccessor acc )
        {
            MingleValue val = acc.expectMingleValue( "val" );

            String nm = String.format( "%s (%s)", 
                Mingle.inspect( val ), val.getClass().getName() );

            ValueBuildTest res = new ValueBuildTest( makeName( acc, nm ) );
            res.val = val;

            return res;
        }

        private
        MingleValueReactorException
        convertReactorError( MingleStructAccessor acc )
        {
            return new MingleValueReactorException(
                overridableString(
                    MingleValueReactorException.class,
                    acc.expectString( "message" )
                )
            );
        }

        private
        StructuralErrorTest
        convertStructuralErrorTest( MingleStructAccessor acc )
            throws Exception
        {
            StructuralErrorTest res = 
                new StructuralErrorTest( makeName( acc, null ) );

            res.events =
                convertReactorEvents( acc.expectListAccessor( "events" ) );
            
            res.expectFailure( 
                convertReactorError( acc.expectStructAccessor( "error" ) ) );

            res.topType = MingleValueReactorTopType.
                valueOf( acc.expectString( "topType" ).toUpperCase() );

            return res;
        }

        protected
        TestImpl
        convertStruct( MingleStruct ms )
            throws Exception
        {
            DeclaredTypeName nm = ms.getType().getName();

            MingleStructAccessor acc = MingleStructAccessor.forStruct( ms );

            if ( nm.equals( TYP_VALUE_BUILD_TEST ) ) {
                return convertValueBuildTest( acc );
            } else if ( nm.equals( TYP_STRUCTURAL_REACTOR_ERROR_TEST ) ) {
                return convertStructuralErrorTest( acc );
            }
            
//            codef( "skipping test: %s", nm );
            return null;
        }
    }

    @InvocationFactory
    private
    List< TestImpl >
    testReactor()
        throws Exception
    {
        return new TestImplReader().read();
    }

    private
    static
    void
    addStringOverride( Object overrideKey,
                       String strKey,
                       String override )
    {
        Map< String, String > m = STRING_OVERRIDES.get( overrideKey );

        if ( m == null ) {
            m = Lang.newMap();
            STRING_OVERRIDES.put( overrideKey, m );
        }

        Lang.putUnique( m, strKey, override );
    }

    static
    {
        STRING_OVERRIDES = Lang.newMap();

        addStringOverride( MingleValueReactorException.class,
            "StartField() called, but struct is built",
            "Saw start of field 'f1' after value was built"
        );
    }
}
