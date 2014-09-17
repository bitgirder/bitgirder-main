package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.test.TestUtils;

import static com.bitgirder.mingle.MingleTestMethods.*;

import com.bitgirder.mingle.Mingle;
import com.bitgirder.mingle.MingleString;
import com.bitgirder.mingle.MingleInt32;
import com.bitgirder.mingle.MingleInt64;
import com.bitgirder.mingle.MingleNamespace;
import com.bitgirder.mingle.MingleStruct;
import com.bitgirder.mingle.MingleSymbolMap;
import com.bitgirder.mingle.QualifiedTypeName;
import com.bitgirder.mingle.MingleIdentifier;
import com.bitgirder.mingle.MingleTypeReference;
import com.bitgirder.mingle.ListTypeReference;
import com.bitgirder.mingle.MingleTestGen;
import com.bitgirder.mingle.MingleList;
import com.bitgirder.mingle.MingleValue;
import com.bitgirder.mingle.MingleBinReader;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPaths;

import com.bitgirder.test.Test;
import com.bitgirder.test.InvocationFactory;
import com.bitgirder.test.LabeledTestCall;

import java.util.List;
import java.util.Map;
import java.util.Queue;
import java.util.Deque;

@Test
final
class MingleReactorTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static MingleNamespace TEST_NS =
        MingleNamespace.create( "mingle:reactor@v1" );

    private
    final
    static
    class StructuralErrorTest
    extends AbstractReactorTest
    {
        private List< MingleReactorEvent > events;
        private MingleReactorTopType topType;

        private StructuralErrorTest( CharSequence nm ) { super( nm ); }
        
        public
        void
        call()
            throws Exception
        {
            StructuralCheck chk = StructuralCheck.forTopType( topType );

            MingleReactorPipeline pip = new MingleReactorPipeline.Builder().
                addReactor( MingleReactors.createDebugReactor() ).
                addReactor( chk ).
                build();

            feedReactorEvents( events, pip );
        }
    }

    private
    final
    static
    class EventPathCheckReactor
    implements MingleReactor
    {
        private final Queue< EventExpectation > events;

        private
        EventPathCheckReactor( Queue< EventExpectation > events )
        {
            this.events = events;
        }

        public
        void
        processEvent( MingleReactorEvent ev )
        {
            state.isFalse( events.isEmpty(), "no more events expected" );

            ObjectPath< MingleIdentifier > expct = events.remove().path;
            assertIdPathsEqual( expct, ev.path() );
        }

        void checkComplete() { state.isTrue( events.isEmpty() ); }
    }

    private
    final
    static
    class EventPathTest
    extends AbstractReactorTest
    {
        private ObjectPath< MingleIdentifier > startPath;
        private Queue< EventExpectation > events;

        private EventPathTest( CharSequence name ) { super( name ); }

        private
        void
        feedEvents( MingleReactor rct )
            throws Exception
        {
            while ( ! events.isEmpty() ) {
                EventExpectation ee = events.peek();
                rct.processEvent( ee.event );
            }
        }

        public
        void
        call()
            throws Exception
        {
            codef( "event expectations: %s", events );

            PathSettingProcessor ps = startPath == null ?
                PathSettingProcessor.create() :
                PathSettingProcessor.create( startPath );

            EventPathCheckReactor chk = new EventPathCheckReactor( events );

            MingleReactorPipeline pip =
                new MingleReactorPipeline.Builder().
                    addReactor( MingleReactors.createDebugReactor() ).
                    addProcessor( ps ).
                    addReactor( chk ).
                    build();
 
            feedEvents( pip );
            chk.checkComplete();
        }
    }

    private
    static
    abstract
    class FieldOrderTest
    extends AbstractReactorTest
    implements FieldOrderProcessor.OrderGetter
    {
        List< MingleReactorEvent > source;
        Map< QualifiedTypeName, FieldOrder > orders;

        public
        final
        FieldOrder
        fieldOrderFor( QualifiedTypeName type )
        {
            return orders.get( type );
        }

        final
        FieldOrderProcessor
        createFieldOrderProcessor()
        {
            return FieldOrderProcessor.create( this );
        }

        final
        void
        feedSource( MingleReactor rct )
            throws Exception
        {
            for ( MingleReactorEvent ev : source ) rct.processEvent( ev );
        }
    }

    private
    final
    static
    class FieldOrderReactorTest
    extends FieldOrderTest
    implements MingleReactor
    {
        private MingleValue expect;

        private final Deque< Object > stack = Lang.newDeque();

        private
        final
        static
        class FieldTracker
        {
            private final FieldOrder ord;

            private int expctIdx;

            private 
            FieldTracker( FieldOrder ord )
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
            FieldOrder ord = fieldOrderFor( type );

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
        processEvent( MingleReactorEvent ev )
        {
            switch ( ev.type() ) {
            case LIST_START: stack.push( "list" ); break;
            case MAP_START: stack.push( "map" ); break;
            case STRUCT_START: pushStruct( ev.structType() ); break;
            case FIELD_START: startField( ev.field() ); break;
            case END: stack.pop(); break;
            }
        }

        public
        void
        call()
            throws Exception
        {
            BuildReactor br = new BuildReactor.Builder().
                setFactory( new ValueBuildFactory() ).
                build();

            MingleReactorPipeline pip =
                new MingleReactorPipeline.Builder().
                    addProcessor( createFieldOrderProcessor() ).
                    addReactor( this ).
                    addReactor( br ).
                    build();

            feedSource( pip );

            assertEqual( expect, (MingleValue) br.value() );
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
            BuildReactor br = new BuildReactor.Builder().
                setFactory( new ValueBuildFactory() ).
                build();

            MingleReactorPipeline pip = 
                new MingleReactorPipeline.Builder().
                    addProcessor( createFieldOrderProcessor() ).
                    addReactor( br ).
                    build();

            feedSource( pip );        

            if ( expect != null ) {
                assertEqual( expect, (MingleValue) br.value() );
            }
        }
    }

    private
    final
    static
    class FieldOrderPathTest
    extends FieldOrderTest
    implements MingleReactor,
               FieldOrderProcessor.OrderGetter
    {
        private Queue< EventExpectation > expect;

        public
        void
        processEvent( MingleReactorEvent ev )
        {
            EventExpectation ee = expect.remove();
            ee.event.setPath( ee.path );

            ReactorTestMethods.
                assertEventsEqual( ee.event, "ee.event", ev, "ev" );
        }

        public
        void
        call()
            throws Exception
        {
            MingleReactorPipeline pip =
                new MingleReactorPipeline.Builder().
                    addProcessor( PathSettingProcessor.create() ).
                    addProcessor( createFieldOrderProcessor() ).
                    addReactor( this ).
                    build();

            feedSource( pip );
            state.isTrue( expect.isEmpty() );
        }
    }

    private
    final
    static
    class DepthTrackerTest
    extends AbstractReactorTest
    implements MingleReactor
    {
        private List< MingleReactorEvent > source;
        private List< Integer > expect = Lang.newList();

        private DepthTracker dt = DepthTracker.create();
        private Queue< Integer > depths = Lang.newQueue();
        
        private DepthTrackerTest( CharSequence nm ) { super( nm ); }

        public
        void
        processEvent( MingleReactorEvent ev )
        {
            state.isFalsef( depths.isEmpty(), 
                "unexpected event: %s", ev.inspect() );
            
            state.equalInt( depths.remove(), dt.depth() );
        }

        public
        void
        call()
            throws Exception
        {
            MingleReactorPipeline pip = new MingleReactorPipeline.Builder().
                addReactor( dt ).
                addReactor( this ).
                build();
            
            depths.addAll( expect );
            feedSource( source, pip );

            state.isTruef( depths.isEmpty(), "depths remain: %s", depths );
        }
    }

    private
    static
    class TestImplReader
    extends MingleReactorTestFileReader< AbstractReactorTest >
    {
        private TestImplReader() { super( TEST_NS ); }
        
        private
        Object
        asBuildReactorExpectValue( MingleValue mv,
                                   String profile )
            throws Exception
        {
            if ( profile.equals( "impl" ) ) return asJavaTestObject( mv );
            return mv;
        }

        private
        BuildReactorTest
        asBuildReactorTest( MingleStruct ms )
            throws Exception
        {
            MingleSymbolMap map = ms.getFields();

            BuildReactorTest res = new BuildReactorTest();
            res.source = asFeedSource( map, "source" );
            res.profile = mapExpectString( map, "profile" );

            MingleValue mv = mapGetValue( map, "val" );
            res.val = asBuildReactorExpectValue( mv, res.profile );

            setOptError( res, map, "error" );

            Object valLbl = res.val instanceof MingleValue
                ? optInspect( (MingleValue) res.val ) : res.val;

            res.setLabel(
                "val", valLbl,
                "profile", res.profile,
                "source", res.sourceToString( res.getSource() ),
                "error", TestUtils.failureExpectationFor( res )
            );

            return res;
        }

        private
        StructuralErrorTest
        asStructuralErrorTest( MingleStruct ms )
            throws Exception
        {
            MingleSymbolMap map = ms.getFields();

            StructuralErrorTest res = 
                new StructuralErrorTest( makeName( ms, null ) );

            res.events = asReactorEvents( map, "events" );
 
            MingleStruct rctErr = mapExpect( map, "error", MingleStruct.class );
            res.expectFailure( asReactorException( rctErr.getFields() ) );

            String ttStr =
                mapExpect( map, "topType", String.class ).toUpperCase();
 
            res.topType = MingleReactorTopType.valueOf( ttStr );

            return res;
        }

        private
        EventPathTest
        asEventPathTest( MingleStruct ms )
            throws Exception
        {
            EventPathTest res = new EventPathTest( makeName( ms, null ) );

            MingleSymbolMap map = ms.getFields();

            res.startPath = asIdentifierPath( map, "startPath" );
            res.events = asEventExpectationQueue( map, "events" );

            return res;
        }

        private
        FieldOrder.FieldSpecification
        asFieldOrderSpecification( MingleSymbolMap map )
            throws Exception
        {
            return new FieldOrder.FieldSpecification(
                asIdentifier( map, "field" ),
                mapExpect( map, "required", Boolean.class )
            );
        }

        private
        FieldOrder
        asFieldOrder( MingleList ml )
            throws Exception
        {
            List< FieldOrder.FieldSpecification > fields = Lang.newList();

            for ( MingleValue mv : ml ) {
                MingleSymbolMap map = ( (MingleStruct) mv ).getFields();
                fields.add( asFieldOrderSpecification( map ) );
            }

            return new FieldOrder( fields );
        }

        private
        Map< QualifiedTypeName, FieldOrder >
        asFieldOrderMapByType( MingleList ml )
            throws Exception
        {
            Map< QualifiedTypeName, FieldOrder > res =
                Lang.newMap();

            for ( MingleValue mv : ml ) 
            {
                MingleSymbolMap map = ( (MingleStruct) mv ).getFields();

                QualifiedTypeName type = asQname(
                    mapExpect( map, "type", byte[].class ) );

                FieldOrder ord = asFieldOrder(
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

            t.source = asReactorEvents( res, "source" );

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
        FieldOrderMissingFieldsTest
        asFieldOrderMissingFieldsTest( MingleStruct ms )
            throws Exception
        {
            FieldOrderMissingFieldsTest res = new FieldOrderMissingFieldsTest();

            MingleSymbolMap map = initFieldOrderTest( res, ms );

            res.expect = mapGetValue( map, "expect" );

            setOptError( res, map, "error" );

            return res;
        }

        private
        FieldOrderPathTest
        asFieldOrderPathTest( MingleStruct ms )
            throws Exception
        {
            FieldOrderPathTest res = new FieldOrderPathTest();

            MingleSymbolMap map = initFieldOrderTest( res, ms );

            res.expect = asEventExpectationQueue( map, "expect" );

            return res;
        }

        private
        DepthTrackerTest
        asDepthTrackerTest( MingleStruct ms )
            throws Exception
        {
            DepthTrackerTest res = new DepthTrackerTest( makeName( ms, null ) );

            MingleSymbolMap m = ms.getFields();

            res.source = asReactorEvents( m, "source" );

            MingleList ml = mapExpect( m, "expect", MingleList.class );

            for ( MingleValue mv : ml ) {
                res.expect.add( ( (MingleInt64) mv ).intValue() );
            }

            return res;
        }

        private
        AbstractReactorTest
        convertTest( MingleStruct ms )
            throws Exception
        {
            String nm = ms.getType().getName().toString();
            MingleSymbolMap map = ms.getFields();

            if ( nm.equals( "StructuralReactorErrorTest" ) ) {
                return asStructuralErrorTest( ms );
            } else if ( nm.equals( "BuildReactorTest" ) ) {
                return asBuildReactorTest( ms );
            } else if ( nm.equals( "EventPathTest" ) ) {
                return asEventPathTest( ms );
            } else if ( nm.equals( "FieldOrderReactorTest" ) ) {
                return asFieldOrderReactorTest( ms );
            } else if ( nm.equals( "FieldOrderMissingFieldsTest" ) ) {
                return asFieldOrderMissingFieldsTest( ms );
            } else if ( nm.equals( "FieldOrderPathTest" ) ) {
                return asFieldOrderPathTest( ms );
            } else if ( nm.equals( "DepthTrackerTest" ) ) {
                return asDepthTrackerTest( ms );
            } else {
                throw state.failf( "unhandled test: %s", nm );
            }
        }

        protected
        AbstractReactorTest
        convertReactorTest( MingleStruct ms )
            throws Exception
        {
            AbstractReactorTest res = convertTest( ms );
            if ( res == null ) return null;

            return res;
        }
    }

    @InvocationFactory
    private
    List< AbstractReactorTest >
    testReactor()
        throws Exception
    {
        return new TestImplReader().read();
    }
}
