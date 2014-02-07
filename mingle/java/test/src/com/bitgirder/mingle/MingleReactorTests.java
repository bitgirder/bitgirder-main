package com.bitgirder.mingle;

import static com.bitgirder.mingle.MingleTestMethods.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

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
import java.util.Deque;

import java.util.regex.Pattern;

@Test
final
class MingleReactorTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static Map< Object, Map< String, String > > STRING_OVERRIDES;

    private final static Map< Object, Object > OBJECT_OVERRIDES;

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
//                    addReactor( MingleValueReactors.createDebugReactor() ).
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
                    addReactor( MingleValueReactors.createDebugReactor() ).
                    addProcessor( ps ).
                    addReactor( this ).
                    build();
 
            feedEvents( pip );
            state.isTrue( events.isEmpty() );
        }
    }

    private
    final
    static
    class FieldTyperImpl
    implements MingleValueCastReactor.FieldTyper
    {
        private final Map< MingleIdentifier, MingleTypeReference > types;

        private
        FieldTyperImpl( Object... pairs )
        {
            this.types = Lang.newMap( 
                MingleIdentifier.class, MingleTypeReference.class, pairs );
        }

        public
        MingleTypeReference
        fieldTypeFor( MingleIdentifier fld,
                      ObjectPath< MingleIdentifier > path )
        {
            MingleTypeReference res = types.get( fld );
            if ( res != null ) return res;

            throw new MingleValueCastException(
                "unrecognized field: " + fld, path );
        }
    }

    private
    final
    static
    class CastDelegateImpl
    implements MingleValueCastReactor.Delegate
    {
        private final static Map< QualifiedTypeName, FieldTyperImpl >
            FIELD_TYPERS = Lang.newMap();

        public
        MingleValueCastReactor.FieldTyper
        fieldTyperFor( QualifiedTypeName qn,
                       ObjectPath< MingleIdentifier > path )
            throws MingleValueCastException
        {
            if ( qn.equals( qname( "ns1@v1/FailType" ) ) ) {
                throw new MingleValueCastException( 
                    "test-message-fail-type", path );
            }

            return FIELD_TYPERS.get( qn );
        }

        public
        boolean
        inferStructFor( QualifiedTypeName qn )
        {
            return FIELD_TYPERS.containsKey( qn );
        }

        private
        MingleValue
        castS3String( String s,
                      ObjectPath< MingleIdentifier > path )
            throws MingleValueCastException
        {
            if ( s.equals( "cast1" ) ) {
                return new MingleInt32( 1 );
            } else if ( s.equals( "cast2" ) ) {
                return new MingleInt32( -1 );
            } else if ( s.equals( "cast3" ) ) {
                String msg = "test-message-cast3";
                throw new MingleValueCastException( msg, path );
            }

            throw new MingleValueCastException( "Unexpected val: " + s, path );
        }

        public
        MingleValue
        castAtomic( MingleValue mv,
                    AtomicTypeReference at,
                    ObjectPath< MingleIdentifier > path )
            throws MingleValueCastException
        {
            if ( ! at.getName().equals( qname( "ns1@v1/S3" ) ) ) return null;
            
            if ( mv instanceof MingleString ) {
                return castS3String( mv.toString(), path );
            }

            throw Mingle.failCastType( at, mv, path );
        }

        static
        {
            FIELD_TYPERS.put( 
                qname( "ns1@v1/T1" ),
                new FieldTyperImpl( id( "f1" ), Mingle.TYPE_INT32 )
            );

            FIELD_TYPERS.put(
                qname( "ns1@v1/T2" ),
                new FieldTyperImpl( 
                    id( "f1" ), Mingle.TYPE_INT32,
                    id( "f2" ), atomic( qname( "ns1@v1/T1" ) )
                )
            );
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
            MingleValueCastReactor.Builder b =
                new MingleValueCastReactor.Builder().
                    setTargetType( type );

            if ( profile != null ) 
            {
                state.isTruef( profile.equals( "interface-impl" ),
                    "unhandled profile: %s", profile );
                
                b.setDelegate( new CastDelegateImpl() );
            }

            return b.build();
        }

        private
        MingleValueReactorPipeline
        createPipeline()
        {
            return new MingleValueReactorPipeline.Builder().
                addProcessor( MinglePathSettingProcessor.create( path ) ).
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
    abstract
    class FieldOrderTest
    extends TestImpl
    implements MingleFieldOrderProcessor.OrderGetter
    {
        List< MingleValueReactorEvent > source;
        Map< QualifiedTypeName, MingleValueReactorFieldOrder > orders;

        public
        final
        MingleValueReactorFieldOrder
        fieldOrderFor( QualifiedTypeName type )
        {
            return orders.get( type );
        }

        final
        MingleFieldOrderProcessor
        createFieldOrderProcessor()
        {
            return MingleFieldOrderProcessor.create( this );
        }

        final
        void
        feedSource( MingleValueReactor rct )
            throws Exception
        {
            for ( MingleValueReactorEvent ev : source ) rct.processEvent( ev );
        }
    }

    private
    final
    static
    class FieldOrderReactorTest
    extends FieldOrderTest
    implements MingleValueReactor
    {
        private MingleValue expect;

        private final Deque< Object > stack = Lang.newDeque();

        private
        final
        static
        class FieldTracker
        {
            private final MingleValueReactorFieldOrder ord;

            private int expctIdx;

            private 
            FieldTracker( MingleValueReactorFieldOrder ord )
            {
                this.ord = ord;
            }

            private
            int
            orderIndexOfField( MingleIdentifier fld )
            {
                for ( int i = 0, e = ord.fields().size(); i < e; ++i ) {
                    if ( ord.fields().get( i ).field().equals( fld ) ) {
                        return i;
                    }
                }

                return -1;
            }
 
            private
            void
            checkField( MingleIdentifier fld )
            {
                int idx = orderIndexOfField( fld );

                if ( idx < 0 ) return;

                if ( idx >= expctIdx ) {
                    expctIdx = idx;
                    return;
                }

                state.failf( 
                    "Expected field %s (ord[ %d ]) but saw %s (ord[ %d ])",
                    ord.fields().get( expctIdx ).field(), expctIdx, fld, idx );
            }
        }

        private
        void
        pushStruct( QualifiedTypeName type )
        {
            MingleValueReactorFieldOrder ord = fieldOrderFor( type );

            if ( ord == null ) {
                stack.push( type );
                return;
            }

            stack.push( new FieldTracker( ord ) );
        }

        private
        void
        startField( MingleIdentifier fld )
        {
            if ( stack.peek() instanceof FieldTracker ) {
                ( (FieldTracker) stack.peek() ).checkField( fld );
            }
        }

        public
        void
        processEvent( MingleValueReactorEvent ev )
        {
            switch ( ev.type() ) {
            case START_LIST: stack.push( "list" ); break;
            case START_MAP: stack.push( "map" ); break;
            case START_STRUCT: pushStruct( ev.structType() ); break;
            case START_FIELD: startField( ev.field() ); break;
            case END: stack.pop(); break;
            }
        }

        public
        void
        call()
            throws Exception
        {
            MingleValueBuilder vb = MingleValueBuilder.create();

            MingleValueReactorPipeline pip =
                new MingleValueReactorPipeline.Builder().
                    addProcessor( createFieldOrderProcessor() ).
                    addReactor( this ).
                    addReactor( vb ).
                    build();

            feedSource( pip );

            MingleTests.assertEqual( expect, vb.value() );
        }
    }

    private
    final
    static
    class FieldOrderMissingFieldsTest
    extends FieldOrderTest
    {
        private MingleValue expect;

        public
        void
        call()
            throws Exception
        {
            MingleValueBuilder vb = MingleValueBuilder.create();

            MingleValueReactorPipeline pip = 
                new MingleValueReactorPipeline.Builder().
                    addProcessor( createFieldOrderProcessor() ).
                    addReactor( vb ).
                    build();

            feedSource( pip );        

            if ( expect != null ) MingleTests.assertEqual( expect, vb.value() );
        }
    }

    private
    final
    static
    class FieldOrderPathTest
    extends FieldOrderTest
    implements MingleValueReactor,
               MingleFieldOrderProcessor.OrderGetter
    {
        private Queue< EventExpectation > expect;

        public
        void
        processEvent( MingleValueReactorEvent ev )
        {
            EventExpectation ee = expect.remove();
            ee.event.setPath( ee.path );

            assertEventsEqual( ee.event, "ee.event", ev, "ev" );
        }

        public
        void
        call()
            throws Exception
        {
            MingleValueReactorPipeline pip =
                new MingleValueReactorPipeline.Builder().
                    addProcessor( MinglePathSettingProcessor.create() ).
                    addProcessor( createFieldOrderProcessor() ).
                    addReactor( this ).
                    build();

            feedSource( pip );
            state.isTrue( expect.isEmpty() );
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
        MingleIdentifier
        asIdentifier( byte[] arr )
            throws Exception
        {
            if ( arr == null ) return null;
            
            return MingleBinReader.create( arr ).readIdentifier();
        }

        private
        List< MingleIdentifier >
        asIdentifierList( MingleList ml )
            throws Exception
        {
            List< MingleIdentifier > res = Lang.newList();

            for ( MingleValue mv : ml ) {
                res.add( asIdentifier( ( (MingleBuffer) mv ).array() ) );
            }

            return res;
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
            ev.setStartField( asIdentifier( arr ) );
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
        void
        setCastReactorOverrides( CastReactorTest t )
        {
            Object ov = OBJECT_OVERRIDES.get( t.getLabel() );
            
            if ( ov instanceof MingleValue ) {
                t.expect = (MingleValue) ov;
            } else if ( ov instanceof MingleValueCastException ) {
                t.resetExpectFailure( (MingleValueCastException) ov );
            }
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
            setCastReactorOverrides( res );

            return res;
        }

        private
        MingleValueReactorFieldSpecification
        asFieldOrderSpecification( MingleSymbolMap map )
            throws Exception
        {
            return new MingleValueReactorFieldSpecification(
                asIdentifier( mapExpect( map, "field", byte[].class ) ),
                mapExpect( map, "required", Boolean.class )
            );
        }

        private
        MingleValueReactorFieldOrder
        asFieldOrder( MingleList ml )
            throws Exception
        {
            List< MingleValueReactorFieldSpecification > fields =
                Lang.newList();

            for ( MingleValue mv : ml )
            {
                MingleSymbolMap map = ( (MingleStruct) mv ).getFields();
                fields.add( asFieldOrderSpecification( map ) );
            }

            return new MingleValueReactorFieldOrder( fields );
        }

        private
        Map< QualifiedTypeName, MingleValueReactorFieldOrder >
        asFieldOrderMapByType( MingleList ml )
            throws Exception
        {
            Map< QualifiedTypeName, MingleValueReactorFieldOrder > res =
                Lang.newMap();

            for ( MingleValue mv : ml ) 
            {
                MingleSymbolMap map = ( (MingleStruct) mv ).getFields();

                QualifiedTypeName type = asQname(
                    mapExpect( map, "type", byte[].class ) );

                MingleValueReactorFieldOrder ord = asFieldOrder(
                    mapExpect( map, "order", MingleList.class ) );
                
                Lang.putUnique( res, type, ord );
            }

            return res;
        }

        private
        MingleSymbolMap
        initFieldOrderTest( FieldOrderTest t,
                            MingleStruct ms )
            throws Exception
        {
            t.setLabel( makeName( ms, null ) );

            MingleSymbolMap res = ms.getFields();

            t.source = asReactorEvents(
                mapExpect( res, "source", MingleList.class ) );

            t.orders = asFieldOrderMapByType(
                mapExpect( res, "orders", MingleList.class ) );
            
            return res;
        }

        private
        FieldOrderReactorTest
        asFieldOrderReactorTest( MingleStruct ms )
            throws Exception
        {
            FieldOrderReactorTest res = new FieldOrderReactorTest();

            MingleSymbolMap map = initFieldOrderTest( res, ms );

            res.expect = mapExpect( map, "expect", MingleValue.class );

            return res;
        }

        private
        MingleMissingFieldsException
        asMissingFieldsError( MingleSymbolMap map )
            throws Exception
        {
            return new MingleMissingFieldsException(
                asIdentifierList( 
                    mapExpect( map, "fields", MingleList.class ) ),
                asIdentifierPath( mapGet( map, "location", byte[].class ) )
            );
        }

        private
        FieldOrderMissingFieldsTest
        asFieldOrderMissingFieldsTest( MingleStruct ms )
            throws Exception
        {
            FieldOrderMissingFieldsTest res = new FieldOrderMissingFieldsTest();

            MingleSymbolMap map = initFieldOrderTest( res, ms );

            res.expect = mapGet( map, "expect", MingleValue.class );

            MingleStruct err = mapGet( map, "error", MingleStruct.class );

            if ( err != null ) {
                res.expectFailure( asMissingFieldsError( err.getFields() ) );
            }

            return res;
        }

        private
        FieldOrderPathTest
        asFieldOrderPathTest( MingleStruct ms )
            throws Exception
        {
            FieldOrderPathTest res = new FieldOrderPathTest();

            MingleSymbolMap map = initFieldOrderTest( res, ms );

            res.expect = asEventExpectationQueue(
                mapExpect( map, "expect", MingleList.class ) );

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
            } else if ( nm.equals( "FieldOrderReactorTest" ) ) {
                return asFieldOrderReactorTest( ms );
            } else if ( nm.equals( "FieldOrderMissingFieldsTest" ) ) {
                return asFieldOrderMissingFieldsTest( ms );
            } else if ( nm.equals( "FieldOrderPathTest" ) ) {
                return asFieldOrderPathTest( ms );
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

        ObjectPath< MingleIdentifier > inValRoot =
            ObjectPath.< MingleIdentifier >
                getRoot( MingleIdentifier.create( "in-val" ) );

        OBJECT_OVERRIDES.put(
            "CastReactorTest:in=\"abc$/@\" (mingle:core@v1/String),type=mingle:core@v1/Buffer,expect=null,profile=null",
            new MingleValueCastException(
                "Length of input 'abc$/@' (6) is not a multiple of 4", 
                inValRoot
            )
        );

        OBJECT_OVERRIDES.put(
            "CastReactorTest:in=\"s\" (mingle:core@v1/String),type=mingle:core@v1/Boolean,expect=null,profile=null",
            new MingleValueCastException(
                "(at or near char 1) Invalid boolean string: s", inValRoot )
        );

        OBJECT_OVERRIDES.put(
            "CastReactorTest:in=2012-01-01T00:00:00.000000000Z (mingle:core@v1/Timestamp),type=mingle:core@v1/Timestamp~[\"2000-01-01T00:00:00.000000000Z\",\"2001-01-01T00:00:00.000000000Z\"],expect=null,profile=null",
            new MingleValueCastException(
                "Value 2012-01-01T00:00:00.000000000Z does not satisfy restriction [\"2000-01-01T00:00:00.000000000Z\",\"2001-01-01T00:00:00.000000000Z\"]",
                inValRoot
            )
        );
    }
}
