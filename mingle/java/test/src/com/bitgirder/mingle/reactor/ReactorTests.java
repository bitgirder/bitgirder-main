package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.test.TestUtils;

import static com.bitgirder.mingle.MingleTestMethods.*;

import com.bitgirder.mingle.Mingle;
import com.bitgirder.mingle.MingleString;
import com.bitgirder.mingle.MingleInt32;
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

//    private final static Map< Object, Object > OBJECT_OVERRIDES;
    
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

//    private
//    final
//    static
//    class FieldTyperImpl
//    implements MingleValueCastReactor.FieldTyper
//    {
//        private final Map< MingleIdentifier, MingleTypeReference > types;
//
//        private
//        FieldTyperImpl( Object... pairs )
//        {
//            this.types = Lang.newMap( 
//                MingleIdentifier.class, MingleTypeReference.class, pairs );
//        }
//
//        public
//        MingleTypeReference
//        fieldTypeFor( MingleIdentifier fld,
//                      ObjectPath< MingleIdentifier > path )
//        {
//            MingleTypeReference res = types.get( fld );
//            if ( res != null ) return res;
//
//            throw new MingleValueCastException(
//                "unrecognized field: " + fld, path );
//        }
//    }
//
//    private
//    final
//    static
//    class CastDelegateImpl
//    implements MingleValueCastReactor.Delegate
//    {
//        private final static Map< QualifiedTypeName, FieldTyperImpl >
//            FIELD_TYPERS = Lang.newMap();
//
//        public
//        MingleValueCastReactor.FieldTyper
//        fieldTyperFor( QualifiedTypeName qn,
//                       ObjectPath< MingleIdentifier > path )
//            throws MingleValueCastException
//        {
//            if ( qn.equals( qname( "ns1@v1/FailType" ) ) ) {
//                throw new MingleValueCastException( 
//                    "test-message-fail-type", path );
//            }
//
//            return FIELD_TYPERS.get( qn );
//        }
//
//        public
//        boolean
//        inferStructFor( QualifiedTypeName qn )
//        {
//            return FIELD_TYPERS.containsKey( qn );
//        }
//
//        private
//        MingleValue
//        castS3String( String s,
//                      ObjectPath< MingleIdentifier > path )
//            throws MingleValueCastException
//        {
//            if ( s.equals( "cast1" ) ) {
//                return new MingleInt32( 1 );
//            } else if ( s.equals( "cast2" ) ) {
//                return new MingleInt32( -1 );
//            } else if ( s.equals( "cast3" ) ) {
//                String msg = "test-message-cast3";
//                throw new MingleValueCastException( msg, path );
//            }
//
//            throw new MingleValueCastException( "Unexpected val: " + s, path );
//        }
//
//        public
//        MingleValue
//        castAtomic( MingleValue mv,
//                    AtomicTypeReference at,
//                    ObjectPath< MingleIdentifier > path )
//            throws MingleValueCastException
//        {
//            if ( ! at.getName().equals( qname( "ns1@v1/S3" ) ) ) return null;
//            
//            if ( mv instanceof MingleString ) {
//                return castS3String( mv.toString(), path );
//            }
//
//            throw Mingle.failCastType( at, mv, path );
//        }
//
//        public
//        boolean
//        allowAssign( QualifiedTypeName targ,
//                     QualifiedTypeName act )
//        {
//            return targ.equals( qname( "ns1@v1/T1" ) ) &&
//                act.equals( qname( "ns1@v1/T1Sub1" ) );
//        }
//
//        static
//        {
//            FIELD_TYPERS.put( 
//                qname( "ns1@v1/T1" ),
//                new FieldTyperImpl( id( "f1" ), Mingle.TYPE_INT32 )
//            );
//
//            FIELD_TYPERS.put(
//                qname( "ns1@v1/T2" ),
//                new FieldTyperImpl( 
//                    id( "f1" ), Mingle.TYPE_INT32,
//                    id( "f2" ), atomic( qname( "ns1@v1/T1" ) )
//                )
//            );
//        }
//    }
//
//    private
//    final
//    static
//    class CastReactorTest
//    extends AbstractReactorTest
//    {
//        private MingleValue in;
//        private MingleValue expect;
//        private ObjectPath< MingleIdentifier > path;
//        private MingleTypeReference type;
//        private String profile;
//
//        private
//        MingleValueCastReactor
//        createCastReactor()
//        {
//            MingleValueCastReactor.Builder b =
//                new MingleValueCastReactor.Builder().
//                    setTargetType( type );
//
//            if ( profile != null ) 
//            {
//                state.isTruef( profile.equals( "interface-impl" ),
//                    "unhandled profile: %s", profile );
//                
//                b.setDelegate( new CastDelegateImpl() );
//            }
//
//            return b.build();
//        }
//
//        private
//        MingleReactorPipeline
//        createPipeline()
//        {
//            return new MingleReactorPipeline.Builder().
//                addProcessor( PathSettingProcessor.create( path ) ).
//                addProcessor( createCastReactor() ).
//                addReactor(
//                    MingleReactors.createDebugReactor( "[post-cast]" ) ).
//                addReactor( MingleValueBuilder.create() ).
//                build();
//        }
//
//        public
//        void
//        call()
//            throws Exception
//        {   
//            MingleReactorPipeline pip = createPipeline();
//
//            MingleReactors.visitValue( in, pip );
//
//            MingleValueBuilder vb = Pipelines.
//                lastElementOfType( pip.pipeline(), MingleValueBuilder.class );
//
//            codef( "expct: %s, act: %s", 
//                Mingle.inspect( expect ), Mingle.inspect( vb.value() ) );
//
//            MingleTests.assertEqual( expect, vb.value() );
//        }
//    }

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

//    private
//    final
//    static
//    class FieldOrderPathTest
//    extends FieldOrderTest
//    implements MingleReactor,
//               FieldOrderProcessor.OrderGetter
//    {
//        private Queue< EventExpectation > expect;
//
//        public
//        void
//        processEvent( MingleReactorEvent ev )
//        {
//            EventExpectation ee = expect.remove();
//            ee.event.setPath( ee.path );
//
//            assertEventsEqual( ee.event, "ee.event", ev, "ev" );
//        }
//
//        public
//        void
//        call()
//            throws Exception
//        {
//            MingleReactorPipeline pip =
//                new MingleReactorPipeline.Builder().
//                    addProcessor( PathSettingProcessor.create() ).
//                    addProcessor( createFieldOrderProcessor() ).
//                    addReactor( this ).
//                    build();
//
//            feedSource( pip );
//            state.isTrue( expect.isEmpty() );
//        }
//    }
//
//    private
//    final
//    static
//    class ReactorCheck
//    {
//        private MingleValue value;
//        private Queue< EventExpectation > events;
//
//        private final MingleValueBuilder vb = MingleValueBuilder.create();
//        private EventPathCheckReactor evChk;
//
//        private
//        MingleReactor
//        reactor()
//        {
//            if ( events == null ) return vb;
//
//            evChk = new EventPathCheckReactor( events );
//
//            return new MingleReactorPipeline.Builder().
//                addReactor( vb ).
//                addReactor( evChk ).
//                build();
//        }
//
//        private
//        void
//        checkComplete()
//        {
//            if ( value != null ) MingleTests.assertEqual( value, vb.value() );
//            if ( evChk != null ) evChk.checkComplete();
//        }
//    }
//
//    private
//    final
//    static
//    class RequestReactorTest
//    extends AbstractReactorTest
//    implements MingleRequestReactor.Delegate
//    {
//        private Object source;
//
//        private MingleNamespace namespace;
//        private MingleIdentifier service;
//        private MingleIdentifier operation;
//        private final ReactorCheck paramsChk = new ReactorCheck();
//        private final ReactorCheck authChk = new ReactorCheck();
//
//        private TopFieldType reqFldMin = TopFieldType.NONE;
//
//        private
//        void
//        checkOrder( TopFieldType ft )
//        {
//            state.isFalsef( 
//                ft.ordinal() < reqFldMin.ordinal(),
//                "saw top field %s but min is %s", ft, reqFldMin );
//
//            reqFldMin = ft;
//        }
//
//        private
//        void
//        checkValue( Object expct,
//                    Object act,
//                    ObjectPath< MingleIdentifier > p,
//                    TopFieldType ft )
//        {
//            checkOrder( ft );            
//
//            state.equalf( expct, act,
//                "%s: expected %s but got %s", 
//                Mingle.formatIdPath( p ), expct, act );
//        }
//
//        public
//        void
//        namespace( MingleNamespace ns,
//                   ObjectPath< MingleIdentifier > p )
//        {
//            checkValue( namespace, ns, p, TopFieldType.NAMESPACE );
//        }
//
//        public
//        void
//        service( MingleIdentifier svc,
//                 ObjectPath< MingleIdentifier > p )
//        {
//            checkValue( service, svc, p, TopFieldType.SERVICE );
//        }
//
//        public
//        void
//        operation( MingleIdentifier op,
//                   ObjectPath< MingleIdentifier > p )
//        {
//            checkValue( operation, op, p, TopFieldType.OPERATION );
//        }
//
//        public
//        MingleReactor
//        getAuthenticationReactor( ObjectPath< MingleIdentifier > p )
//        {
//            checkOrder( TopFieldType.AUTHENTICATION );
//            return authChk.reactor();
//        }
//
//        public
//        MingleReactor
//        getParametersReactor( ObjectPath< MingleIdentifier > p )
//        {
//            checkOrder( TopFieldType.PARAMETERS );
//            return paramsChk.reactor();
//        }
//
//        public
//        void
//        call()
//            throws Exception
//        {
//            MingleRequestReactor rct = 
//                MingleRequestReactor.create( this );
//
//            MingleReactorPipeline pip =
//                new MingleReactorPipeline.Builder().
//                    addReactor( MingleReactors.createDebugReactor() ).
//                    addReactor( rct ).
//                    build();
//
//            feedSource( source, pip );
//
//            authChk.checkComplete();
//            paramsChk.checkComplete();
//        }
//    }
//
//    private
//    final
//    static
//    class ResponseReactorTest
//    extends AbstractReactorTest
//    implements MingleResponseReactor.Delegate
//    {
//        private MingleValue in;
//
//        private final ReactorCheck resChk = new ReactorCheck();
//        private final ReactorCheck errChk = new ReactorCheck();
//
//        public
//        MingleReactor
//        getResultReactor( ObjectPath< MingleIdentifier > p )
//        {
////            return resChk.reactor();
//            return new MingleReactorPipeline.Builder().
//                addReactor(
//                    MingleReactors.createDebugReactor( "[res]" ) ).
//                addReactor( resChk.reactor() ).
//                build();
//        }
//
//        public
//        MingleReactor
//        getErrorReactor( ObjectPath< MingleIdentifier > p )
//        {
//            return errChk.reactor();
//        }
//
//        public
//        void
//        call()
//            throws Exception
//        {
//            MingleResponseReactor rct = MingleResponseReactor.create( this );
//
//            MingleReactorPipeline pip =
//                new MingleReactorPipeline.Builder().
//                    addReactor( 
//                        MingleReactors.createDebugReactor( "[test]" ) ).
//                    addReactor( rct ).
//                    build();
//
//            feedSource( in, pip );
//
//            resChk.checkComplete();
//            errChk.checkComplete();
//        }
//    }

    private
    static
    class TestImplReader
    extends MingleReactorTestFileReader< AbstractReactorTest >
    {
        private TestImplReader() { super( TEST_NS ); }

//        private
//        void
//        setErrorOverride( TestImpl t )
//        {
//            Object ov = OBJECT_OVERRIDES.get( t.getLabel() );
//
//            if ( ov instanceof Exception ) {
//                t.resetExpectFailure( (Exception) ov );
//            }
//        }

//        private
//        MingleValue
//        valOrNull( MingleValue mv )
//        {
//            if ( mv == null ) return MingleNull.getInstance();
//            return mv;
//        }
        
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

//        private
//        < E extends MingleValueException >
//        E
//        asValueException( Class< E > cls,
//                          MingleStruct ms )
//            throws Exception
//        {
//            MingleSymbolMap map = ms.getFields();
//
//            String msg = mapExpect( map, "message", String.class );
//
//            ObjectPath< MingleIdentifier > loc = 
//                asIdentifierPath( mapGet( map, "location", byte[].class ) );
//
//            return ReflectUtils.newInstance(
//                cls,
//                new Class< ? >[]{ String.class, ObjectPath.class },
//                new Object[]{ msg, loc }
//            );
//        }
//
//        private
//        MingleMissingFieldsException
//        asMissingFieldsException( MingleStruct ms )
//            throws Exception
//        {
//            MingleSymbolMap map = ms.getFields();
//
//            List< MingleIdentifier > flds =
//                asIdentifierList( 
//                    mapExpect( map, "fields", MingleList.class ) );
//
//            ObjectPath< MingleIdentifier > loc =
//                asIdentifierPath( mapGet( map, "location", byte[].class ) );
//
//            return new MingleMissingFieldsException( flds, loc );
//        }
//
//        private
//        MingleUnrecognizedFieldException
//        asUnrecognizedFieldException( MingleStruct ms )
//            throws Exception
//        {
//            MingleSymbolMap map = ms.getFields();
//
//            return new MingleUnrecognizedFieldException(
//                asIdentifier( mapExpect( map, "field", byte[].class ) ),
//                asIdentifierPath( mapGet( map, "location", byte[].class ) )
//            );
//        }
//
//        private
//        Exception
//        asError( MingleStruct ms )
//            throws Exception
//        {
//            if ( ms == null ) return null;
//
//            String nm = ms.getType().getName().toString();
//
//            if ( nm.equals( "ValueCastError" ) ) {
//                return asValueException( MingleValueCastException.class, ms );
//            } else if ( nm.equals( "MissingFieldsError" ) ) {
//                return asMissingFieldsException( ms );
//            } else if ( nm.equals( "UnrecognizedFieldError" ) ) {
//                return asUnrecognizedFieldException( ms );
//            }
//
//            throw state.failf( "unhandled error: %s", nm );
//        }

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

//        private
//        void
//        setCastReactorTestValues( CastReactorTest t,
//                                  MingleSymbolMap map )
//            throws Exception
//        {
//            t.in = valOrNull( mapGet( map, "in", MingleValue.class ) );
//
//            t.expect = valOrNull( mapGet( map, "expect", MingleValue.class ) );
//
//            t.path = asIdentifierPath( mapExpect( map, "path", byte[].class ) );
//
//            t.type = asTypeReference( mapExpect( map, "type", byte[].class ) );
//
//            t.profile = mapGet( map, "profile", String.class );
//            
//            MingleStruct errStruct = mapGet( map, "err", MingleStruct.class );
//            Exception err = asError( errStruct );
//            if ( err != null ) t.expectFailure( err );
//        }
//
//        private
//        void
//        setCastReactorLabel( CastReactorTest t )
//        {
//            String inVal = null;
//
//            if ( t.in != null ) 
//            {
//                inVal = String.format( "%s (%s)",
//                    Mingle.inspect( t.in ), Mingle.inferredTypeOf( t.in ) );
//            }
//
//            t.setLabel(
//                "in", inVal,
//                "type", t.type,
//                "expect", t.expect == null ? null : Mingle.inspect( t.expect ),
//                "profile", t.profile
//            );
//        }
//
//        private
//        void
//        setCastReactorOverrides( CastReactorTest t )
//        {
//            Object ov = OBJECT_OVERRIDES.get( t.getLabel() );
//            
//            if ( ov instanceof MingleValue ) t.expect = (MingleValue) ov;
//        }
//
//        private
//        CastReactorTest
//        asCastReactorTest( MingleStruct ms )
//            throws Exception
//        {
//            CastReactorTest res = new CastReactorTest();
//
//            MingleSymbolMap map = ms.getFields();
//            setCastReactorTestValues( res, map );
//
//            setCastReactorLabel( res );
//            setCastReactorOverrides( res ); // needs label set first
//
//            return res;
//        }

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

//        private
//        MingleMissingFieldsException
//        asMissingFieldsError( MingleSymbolMap map )
//            throws Exception
//        {
//            return new MingleMissingFieldsException(
//                asIdentifierList( 
//                    mapExpect( map, "fields", MingleList.class ) ),
//                asIdentifierPath( mapGet( map, "location", byte[].class ) )
//            );
//        }

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

//        private
//        FieldOrderPathTest
//        asFieldOrderPathTest( MingleStruct ms )
//            throws Exception
//        {
//            FieldOrderPathTest res = new FieldOrderPathTest();
//
//            MingleSymbolMap map = initFieldOrderTest( res, ms );
//
//            res.expect = asEventExpectationQueue(
//                mapExpect( map, "expect", MingleList.class ) );
//
//            return res;
//        }
//
//        private
//        void
//        setOptError( TestImpl ti,
//                     MingleSymbolMap map )
//            throws Exception
//        {
//            MingleStruct errStruct = mapGet( map, "error", MingleStruct.class );
//            if ( errStruct != null ) ti.expectFailure( asError( errStruct ) );
//        }
//
//        private
//        RequestReactorTest
//        asRequestReactorTest( MingleStruct ms )
//            throws Exception
//        {
//            RequestReactorTest res = new RequestReactorTest();
//            res.setLabel( makeName( ms, null ) );
//
//            MingleSymbolMap map = ms.getFields();
//
//            res.source = 
//                asObject( mapExpect( map, "source", MingleValue.class ) );
//
//            res.namespace = 
//                asNamespace( mapGet( map, "namespace", byte[].class ) );
//
//            res.service =
//                asIdentifier( mapGet( map, "service", byte[].class ) );
//
//            res.operation = 
//                asIdentifier( mapGet( map, "operation", byte[].class ) );
//
//            res.paramsChk.value = 
//                mapGet( map, "parameters", MingleSymbolMap.class );
//
//            res.paramsChk.events = 
//                asEventExpectationQueue(
//                    mapGet( map, "parameterEvents", MingleList.class ) );
//
//            res.authChk.value =
//                mapGet( map, "authentication", MingleValue.class );
//
//            res.authChk.events =
//                asEventExpectationQueue(
//                    mapGet( map, "authenticationEvents", MingleList.class ) );
//            
//            setOptError( res, map );
//
//            return res;
//        }
//
//        private
//        ResponseReactorTest
//        asResponseReactorTest( MingleStruct ms )
//            throws Exception
//        {
//            ResponseReactorTest res = new ResponseReactorTest();
//            res.setLabel( makeName( ms, null ) );
//
//            MingleSymbolMap map = ms.getFields();
//
//            res.in = mapExpect( map, "in", MingleValue.class );
//
//            res.resChk.value = mapGet( map, "resVal", MingleValue.class );
//
//            res.resChk.events = asEventExpectationQueue( 
//                mapGet( map, "resEvents", MingleList.class ) );
//            
//            res.errChk.value = mapGet( map, "errVal", MingleValue.class );
//
//            res.errChk.events = asEventExpectationQueue(
//                mapGet( map, "errEvents", MingleList.class ) );
//
//            setOptError( res, map );
//
//            return res;
//        }

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
//            } else if ( nm.equals( "CastReactorTest" ) ) {
//                return asCastReactorTest( ms );
            } else if ( nm.equals( "FieldOrderReactorTest" ) ) {
                return asFieldOrderReactorTest( ms );
            } else if ( nm.equals( "FieldOrderMissingFieldsTest" ) ) {
                return asFieldOrderMissingFieldsTest( ms );
//            } else if ( nm.equals( "FieldOrderPathTest" ) ) {
//                return asFieldOrderPathTest( ms );
//            } else if ( nm.equals( "RequestReactorTest" ) ) {
//                return asRequestReactorTest( ms );
//            } else if ( nm.equals( "ResponseReactorTest" ) ) {
//                return asResponseReactorTest( ms );
            } else {
//                throw state.failf( "unhandled test: %s", nm );
                codef( "skipping test: %s", nm );
                return null;
            }
        }

        protected
        AbstractReactorTest
        convertReactorTest( MingleStruct ms )
            throws Exception
        {
            AbstractReactorTest res = convertTest( ms );
            if ( res == null ) return null;

//            setErrorOverride( res );

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

//    static
//    {
//        OBJECT_OVERRIDES = Lang.newMap();
//
//        OBJECT_OVERRIDES.put(
//            "CastReactorTest:in=2007-08-24T21:15:43.123450000Z (mingle:core@v1/Timestamp),type=mingle:core@v1/String,expect=\"2007-08-24T13:15:43.12345-08:00\",profile=null",
//            new MingleString( "2007-08-24T21:15:43.123450000Z" )
//        );
//
//        OBJECT_OVERRIDES.put(
//            "CastReactorTest:in=1.0 (mingle:core@v1/Float64),type=mingle:core@v1/String,expect=\"1\",profile=null",
//            new MingleString( "1.0" )
//        );
//
//        OBJECT_OVERRIDES.put(
//            "CastReactorTest:in=1.0 (mingle:core@v1/Float32),type=mingle:core@v1/String,expect=\"1\",profile=null",
//            new MingleString( "1.0" )
//        );
//
//        ObjectPath< MingleIdentifier > inValRoot =
//            ObjectPath.< MingleIdentifier >
//                getRoot( MingleIdentifier.create( "in-val" ) );
//
//        OBJECT_OVERRIDES.put(
//            "CastReactorTest:in=\"abc$/@\" (mingle:core@v1/String),type=mingle:core@v1/Buffer,expect=null,profile=null",
//            new MingleValueCastException(
//                "Length of input 'abc$/@' (6) is not a multiple of 4", 
//                inValRoot
//            )
//        );
//
//        OBJECT_OVERRIDES.put(
//            "CastReactorTest:in=\"s\" (mingle:core@v1/String),type=mingle:core@v1/Boolean,expect=null,profile=null",
//            new MingleValueCastException(
//                "(at or near char 1) Invalid boolean string: s", inValRoot )
//        );
//
//        OBJECT_OVERRIDES.put(
//            "CastReactorTest:in=2012-01-01T00:00:00.000000000Z (mingle:core@v1/Timestamp),type=mingle:core@v1/Timestamp~[\"2000-01-01T00:00:00.000000000Z\",\"2001-01-01T00:00:00.000000000Z\"],expect=null,profile=null",
//            new MingleValueCastException(
//                "Value 2012-01-01T00:00:00.000000000Z does not satisfy restriction [\"2000-01-01T00:00:00.000000000Z\",\"2001-01-01T00:00:00.000000000Z\"]",
//                inValRoot
//            )
//        );
//
//        OBJECT_OVERRIDES.put(
//            "RequestReactorTest/15",
//            new MingleValueCastException(
//                "can't convert to namespace from mingle:core@v1/Boolean",
//                ObjectPath.getRoot( Mingle.ID_NAMESPACE )
//            )
//        );
//
//        OBJECT_OVERRIDES.put(
//            "RequestReactorTest/19",
//            new MingleValueCastException(
//                "can't convert to identifier from mingle:core@v1/Boolean",
//                ObjectPath.getRoot( Mingle.ID_SERVICE )
//            )
//        );
//
//        OBJECT_OVERRIDES.put(
//            "RequestReactorTest/20",
//            new MingleValueCastException(
//                "can't convert to identifier from mingle:core@v1/Boolean",
//                ObjectPath.getRoot( Mingle.ID_OPERATION )
//            )
//        );
//
//        OBJECT_OVERRIDES.put(
//            "RequestReactorTest/22",
//            new MingleValueCastException(
//                "could not read namespace: [offset 0]: Expected namespace but saw type code 0x0f",
//                ObjectPath.getRoot( Mingle.ID_NAMESPACE )
//            )
//        );
//
//        OBJECT_OVERRIDES.put(
//            "RequestReactorTest/23",
//            new MingleValueCastException(
//                "could not read identifier: [offset 0]: Expected identifier but saw type code 0x0f",
//                ObjectPath.getRoot( Mingle.ID_SERVICE )
//            )
//        );
//
//        OBJECT_OVERRIDES.put(
//            "RequestReactorTest/24",
//            new MingleValueCastException(
//                "could not read identifier: [offset 0]: Expected identifier but saw type code 0x0f",
//                ObjectPath.getRoot( Mingle.ID_OPERATION )
//            )
//        );
//
//        OBJECT_OVERRIDES.put(
//            "RequestReactorTest/25",
//            new MingleValueCastException(
//                "could not parse namespace: (at or near char 5) Illegal start of identifier part: \":\" (U+003A)",
//                ObjectPath.getRoot( Mingle.ID_NAMESPACE )
//            )
//        );
//
//        OBJECT_OVERRIDES.put(
//            "RequestReactorTest/26",
//            new MingleValueCastException(
//                "could not parse identifier: (at or near char 1) Illegal start of identifier part: \"2\" (U+0032)",
//                ObjectPath.getRoot( Mingle.ID_SERVICE )
//            )
//        );
//
//        OBJECT_OVERRIDES.put(
//            "RequestReactorTest/27",
//            new MingleValueCastException(
//                "could not parse identifier: (at or near char 1) Illegal start of identifier part: \"2\" (U+0032)",
//                ObjectPath.getRoot( Mingle.ID_OPERATION )
//            )
//        );
//    }
}
