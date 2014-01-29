package com.bitgirder.mingle;

import static com.bitgirder.mingle.MingleTestMethods.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;

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
    final
    static
    class CastReactorTest
    extends TestImpl
    {
        private MingleValue in;
        private MingleValue expect;
        private ObjectPath< MingleIdentifier > path;
        private MingleTypeReference type;
        private String profile;

        private CastReactorTest( CharSequence name ) { super( name ); }

        private
        MingleValueCastReactor
        createCastReactor()
        {
            return MingleValueCastReactor.create();
        }

        public
        void
        call()
            throws Exception
        {   
            MingleValueReactorPipeline pip =
                new MingleValueReactorPipeline.Builder().
                    addProcessor( createCastReactor() ).
                    addReactor( MingleValueBuilder.create() ).
                    build();

            MingleValueReactors.visitValue( in, pip );

            MingleValueBuilder vb = Pipelines.
                lastElementOfType( pip.pipeline(), MingleValueBuilder.class );

            MingleTests.assertEqual( expect, vb.value() );
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
        makeName( MingleStruct ms,
                  Object name )
        {
            DeclaredTypeName declNm = ms.getType().getName();

            if ( name == null ) {
                Integer seq = seqsByType.get( declNm );
                if ( seq == null ) seq = Integer.valueOf( 0 );
                name = seq.toString();
                seqsByType.put( declNm, seq + 1 );
            }

            return declNm + "/" + name;
        }
        
        private
        QualifiedTypeName
        asQname( byte[] arr )
            throws Exception
        {
            if ( arr == null ) return null;

            return MingleBinReader.create( arr ).readQualifiedTypeName();
        }

        private
        MingleTypeReference
        asTypeReference( byte[] arr )
            throws Exception
        {
            if ( arr == null ) return null;

            return MingleBinReader.create( arr ).readTypeReference();
        }

        private
        ObjectPath< MingleIdentifier >
        asIdentifierPath( byte[] arr )
            throws Exception
        {
            if ( arr == null ) return null;

            return MingleBinReader.create( arr ).readIdPath();
        }

        private
        void
        setEventStartStruct( MingleValueReactorEvent ev,
                             MingleSymbolMap map )
            throws Exception
        {
            byte[] arr = mapExpect( map, "type", byte[].class );
            ev.setStartStruct( asQname( arr ) );
        }

        private
        void
        setEventStartField( MingleValueReactorEvent ev,
                            MingleSymbolMap map )
            throws Exception
        {
            byte[] arr = mapExpect( map, "field", byte[].class );
            MingleBinReader rd = MingleBinReader.create( arr );

            ev.setStartField( rd.readIdentifier() );
        }

        private
        MingleValueReactorEvent
        asReactorEvent( MingleStruct ms )
            throws Exception
        {
            MingleValueReactorEvent res = new MingleValueReactorEvent();

            String evName = ms.getType().getName().toString();
            MingleSymbolMap map = ms.getFields();

            if ( evName.equals( "StructStartEvent" ) ) {
                setEventStartStruct( res, map );
            } else if ( evName.equals( "FieldStartEvent" ) ) {
                setEventStartField( res, map );
            } else if ( evName.equals( "MapStartEvent" ) ) {
                res.setStartMap();
            } else if ( evName.equals( "ListStartEvent" ) ) {
                res.setStartList();
            } else if ( evName.equals( "EndEvent" ) ) {
                res.setEnd();
            } else if ( evName.equals( "ValueEvent" ) ) {
                res.setValue( mapExpect( map, "val", MingleValue.class ) );
            } else {
                state.failf( "unhandled event: %s", evName );
            }

            return res;
        }

        private
        List< MingleValueReactorEvent >
        asReactorEvents( MingleList ml )
            throws Exception
        {
            List< MingleValueReactorEvent > res = Lang.newList();
    
            for ( MingleValue mv : ml ) {
                res.add( asReactorEvent( (MingleStruct) mv ) );
            }

            return res;
        }

        private
        ValueBuildTest
        asValueBuildTest( MingleStruct ms )
        {
            MingleSymbolMap map = ms.getFields();

            MingleValue val = mapExpect( map, "val", MingleValue.class );

            String nm = String.format( "%s (%s)", 
                Mingle.inspect( val ), val.getClass().getName() );

            ValueBuildTest res = new ValueBuildTest( makeName( ms, nm ) );
            res.val = val;

            return res;
        }

        private
        MingleValueReactorException
        asReactorError( MingleSymbolMap map )
        {
            return new MingleValueReactorException(
                overridableString(
                    MingleValueReactorException.class,
                    mapExpect( map, "message", String.class )
                )
            );
        }

        private
        MingleValueCastException
        asValueCastException( MingleSymbolMap map )
            throws Exception
        {
            if ( map == null ) return null;
            
            return new MingleValueCastException(
                mapExpect( map, "message", String.class ),
                asIdentifierPath( mapExpect( map, "location", byte[].class ) )
            );
        }

        private
        Exception
        asError( MingleStruct ms )
            throws Exception
        {
            if ( ms == null ) return null;

            String nm = ms.getType().getName().toString();

            if ( nm.equals( "ValueCastError" ) ) {
                return asValueCastException( ms.getFields() );
            }

            throw state.failf( "unhandled error: %s", nm );
        }

        private
        StructuralErrorTest
        asStructuralErrorTest( MingleStruct ms )
            throws Exception
        {
            MingleSymbolMap map = ms.getFields();

            StructuralErrorTest res = 
                new StructuralErrorTest( makeName( ms, null ) );

            res.events = 
                asReactorEvents( mapExpect( map, "events", MingleList.class ) );
 
            MingleStruct rctErr = mapExpect( map, "error", MingleStruct.class );
            res.expectFailure( asReactorError( rctErr.getFields() ) );

            String ttStr =
                mapExpect( map, "topType", String.class ).toUpperCase();
            
            res.topType = MingleValueReactorTopType.valueOf( ttStr );

            return res;
        }

        private
        CastReactorTest
        asCastReactorTest( MingleStruct ms )
            throws Exception
        {
            CastReactorTest res = new CastReactorTest( makeName( ms, null ) );

            MingleSymbolMap map = ms.getFields();

            res.in = mapGet( map, "in", MingleValue.class );
            res.expect = mapGet( map, "expect", MingleValue.class );

            res.path = 
                asIdentifierPath( mapExpect( map, "path", byte[].class ) );

            res.type = 
                asTypeReference( mapExpect( map, "type", byte[].class ) );

            res.profile = mapGet( map, "profile", String.class );
            
            MingleStruct errStruct = mapGet( map, "err", MingleStruct.class );
            Exception err = asError( errStruct );
            if ( err != null ) res.expectFailure( err );

            return res;
        }

        protected
        TestImpl
        convertStruct( MingleStruct ms )
            throws Exception
        {
            String nm = ms.getType().getName().toString();
            MingleSymbolMap map = ms.getFields();

            if ( nm.equals( "ValueBuildTest" ) ) {
                return asValueBuildTest( ms );
            } else if ( nm.equals( "StructuralReactorErrorTest" ) ) {
                return asStructuralErrorTest( ms );
            } else if ( nm.equals( "CastReactorTest" ) ) {
                return asCastReactorTest( ms );
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
