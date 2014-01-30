package com.bitgirder.mingle;

import static com.bitgirder.mingle.MingleTestMethods.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ListPath;
import com.bitgirder.lang.path.DictionaryPath;

import com.bitgirder.pipeline.Pipelines;

import com.bitgirder.test.Test;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.LabeledTestCall;

import java.util.List;
import java.util.Map;
import java.util.Queue;

import java.util.regex.Pattern;

@Test
final
class MingleReactorTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static Map< Object, Map< String, String > > STRING_OVERRIDES;

    private final static Map< Object, Object > OBJECT_OVERRIDES;

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
        private TestImpl() { super(); }

        void
        setLabel( Object... pairs )
        {
            setLabel( getClass().getSimpleName() + ":" + 
                Strings.crossJoin( "=", ",", pairs ) );
        }
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
    class EventExpectation
    {
        private MingleValueReactorEvent event;
        private ObjectPath< MingleIdentifier > path;
    }

    private
    final
    static
    class EventPathTest
    extends TestImpl
    implements MingleValueReactor
    {
        private ObjectPath< MingleIdentifier > startPath;
        private ObjectPath< MingleIdentifier > finalPath;
        private Queue< EventExpectation > events;

        // the most recent path seen
        private ObjectPath< MingleIdentifier > lastPath;

        private EventPathTest( CharSequence name ) { super( name ); }

        private
        ObjectPath< MingleIdentifier >
        append( ObjectPath< MingleIdentifier > head,
                ObjectPath< MingleIdentifier > tail )
        {
            if ( tail == null ) return head;

            for ( ObjectPath< MingleIdentifier > elt : tail.collectDescent() )
            {
                if ( elt instanceof DictionaryPath ) 
                {
                    DictionaryPath< MingleIdentifier > dp = 
                        Lang.castUnchecked( elt );
                    
                    head = head.descend( dp.getKey() );
                }
                else if ( elt instanceof ListPath ) 
                {
                    ListPath< ? > lp = (ListPath< ? >) elt;
                    head = head.startImmutableList( lp.getIndex() );
                }
                else state.failf( "unhandled elt: %s", elt );
            }

            return head;
        }

        public
        void
        processEvent( MingleValueReactorEvent ev )
        {
            ObjectPath< MingleIdentifier > expct = events.peek().path;
            if ( startPath != null ) expct = append( startPath, expct );
            assertIdPathsEqual( expct, ev.path() );

            lastPath = ev.path();
            events.remove();
        }

        private
        void
        feedEvents( MingleValueReactor rct )
            throws Exception
        {
            while ( ! events.isEmpty() )
            {
                EventExpectation ee = events.peek();
                rct.processEvent( ee.event );
            }
        }

        public
        void
        call()
            throws Exception
        {
            MinglePathSettingProcessor ps = startPath == null ?
                MinglePathSettingProcessor.create() :
                MinglePathSettingProcessor.create( startPath );

            MingleValueReactorPipeline pip =
                new MingleValueReactorPipeline.Builder().
                    addProcessor( ps ).
                    addReactor( MingleValueReactors.createDebugReactor() ).
                    addReactor( this ).
                    build();
 
            feedEvents( pip );
            state.isTrue( events.isEmpty() );
            assertIdPathsEqual( finalPath, "finalPath", ps.path(), "ps.path" );
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

        private
        MingleValueCastReactor
        createCastReactor()
        {
            return MingleValueCastReactor.create( type );
        }

        private
        MingleValueReactorPipeline
        createPipeline()
        {
            ObjectPath< MingleIdentifier > rtPath = 
                ObjectPath.getRoot( MingleIdentifier.create( "in-val" ) );
            
            return new MingleValueReactorPipeline.Builder().
                addReactor( MingleValueReactors.createDebugReactor() ).
                addProcessor( MinglePathSettingProcessor.create( rtPath ) ).
                addProcessor( createCastReactor() ).
                addReactor( MingleValueBuilder.create() ).
                build();
        }

        public
        void
        call()
            throws Exception
        {   
            MingleValueReactorPipeline pip = createPipeline();

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
        MingleValue
        valOrNull( MingleValue mv )
        {
            if ( mv == null ) return MingleNull.getInstance();
            return mv;
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
        Queue< EventExpectation >
        asEventExpectationQueue( MingleList ml )
            throws Exception
        {
            Queue< EventExpectation > res = Lang.newQueue();

            for ( MingleValue mv : ml ) 
            {
                EventExpectation ee = new EventExpectation();

                MingleSymbolMap map = ( (MingleStruct) mv ).getFields();

                ee.event = asReactorEvent( 
                    mapExpect( map, "event", MingleStruct.class ) );

                ee.path =
                    asIdentifierPath( mapGet( map, "path", byte[].class ) );

                res.add( ee );
            }

            return res;
        }

        private
        EventPathTest
        asEventPathTest( MingleStruct ms )
            throws Exception
        {
            EventPathTest res = new EventPathTest( makeName( ms, null ) );

            MingleSymbolMap map = ms.getFields();

            res.startPath = 
                asIdentifierPath( mapGet( map, "startPath", byte[].class ) );

            res.finalPath =
                asIdentifierPath( mapGet( map, "finalPath", byte[].class ) );

            res.events = asEventExpectationQueue( 
                mapGet( map, "events", MingleList.class ) );

            return res;
        }

        private
        void
        setCastReactorTestValues( CastReactorTest t,
                                  MingleSymbolMap map )
            throws Exception
        {
            t.in = valOrNull( mapGet( map, "in", MingleValue.class ) );

            t.expect = valOrNull( mapGet( map, "expect", MingleValue.class ) );

            t.path = asIdentifierPath( mapExpect( map, "path", byte[].class ) );

            t.type = asTypeReference( mapExpect( map, "type", byte[].class ) );

            t.profile = mapGet( map, "profile", String.class );
            
            MingleStruct errStruct = mapGet( map, "err", MingleStruct.class );
            Exception err = asError( errStruct );
            if ( err != null ) t.expectFailure( err );
        }

        private
        void
        setCastReactorLabel( CastReactorTest t )
        {
            String inVal = null;

            if ( t.in != null ) 
            {
                inVal = String.format( "%s (%s)",
                    Mingle.inspect( t.in ), Mingle.inferredTypeOf( t.in ) );
            }

            t.setLabel(
                "in", inVal,
                "type", t.type,
                "expect", t.expect == null ? null : Mingle.inspect( t.expect ),
                "profile", t.profile
            );
        }

        private
        CastReactorTest
        asCastReactorTest( MingleStruct ms )
            throws Exception
        {
            CastReactorTest res = new CastReactorTest();

            MingleSymbolMap map = ms.getFields();
            setCastReactorTestValues( res, map );

            setCastReactorLabel( res );

            Object ov = OBJECT_OVERRIDES.get( res.getLabel() );
            if ( ov != null ) res.expect = (MingleValue) ov;

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
            } else if ( nm.equals( "EventPathTest" ) ) {
                return asEventPathTest( ms );
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

        OBJECT_OVERRIDES = Lang.newMap();

        OBJECT_OVERRIDES.put(
            "CastReactorTest:in=2007-08-24T21:15:43.123450000Z (mingle:core@v1/Timestamp),type=mingle:core@v1/String,expect=\"2007-08-24T13:15:43.12345-08:00\",profile=null",
            new MingleString( "2007-08-24T21:15:43.123450000Z" )
        );

        OBJECT_OVERRIDES.put(
            "CastReactorTest:in=1.0 (mingle:core@v1/Float64),type=mingle:core@v1/String,expect=\"1\",profile=null",
            new MingleString( "1.0" )
        );

        OBJECT_OVERRIDES.put(
            "CastReactorTest:in=1.0 (mingle:core@v1/Float32),type=mingle:core@v1/String,expect=\"1\",profile=null",
            new MingleString( "1.0" )
        );
    }
}
