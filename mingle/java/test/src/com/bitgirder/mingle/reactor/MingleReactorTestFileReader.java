package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import static com.bitgirder.mingle.MingleTestMethods.*;

import com.bitgirder.mingle.Mingle;
import com.bitgirder.mingle.MingleUnrecognizedFieldException;
import com.bitgirder.mingle.MingleNamespace;
import com.bitgirder.mingle.MingleStruct;
import com.bitgirder.mingle.MingleSymbolMap;
import com.bitgirder.mingle.MingleUint64;
import com.bitgirder.mingle.MingleBuffer;
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

import java.util.List;
import java.util.Map;
import java.util.Queue;

public
abstract
class MingleReactorTestFileReader< T >
extends MingleTestGen.StructFileReader< T >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final MingleNamespace testNs;
    
    private final Map< QualifiedTypeName, Integer > seqsByType = Lang.newMap();

    private final static ListTypeReference TYPE_EVENT_LIST = 
        listType( atomic( qname( "mingle:reactor@v1/Event" ) ), true );

    private final static ListTypeReference TYPE_EVENT_EXPECTATION_LIST = 
        listType( atomic( qname( "mingle:reactor@v1/Event" ) ), true );

    private final static QualifiedTypeName QNAME_TEST_ERROR =
        qname( "mingle:reactor@v1/TestError" );

    private final static QualifiedTypeName QNAME_REACTOR_ERROR =
        qname( "mingle:reactor@v1/ReactorError" );

    private final static QualifiedTypeName QNAME_UNRECOGNIZED_FIELD_ERROR =
        qname( "mingle:core@v1/UnrecognizedFieldError" );

    private final static QualifiedTypeName QNAME_MISSING_FIELDS_ERROR =
        qname( "mingle:core@v1/MissingFieldsError" );

    private final static QualifiedTypeName QNAME_CAST_ERROR =
        qname( "mingle:core@v1/CastError" );

    protected
    MingleReactorTestFileReader( MingleNamespace testNs )
    {
        super( "reactor-tests.bin" );

        this.testNs = inputs.notNull( testNs, "testNs" );
    }

    protected
    final
    CharSequence
    makeName( MingleStruct ms,
              Object name )
    {
        inputs.notNull( ms, "ms" );

        QualifiedTypeName qn = ms.getType();

        if ( name == null ) {
            Integer seq = seqsByType.get( qn );
            if ( seq == null ) seq = Integer.valueOf( 0 );
            name = seq.toString();
            seqsByType.put( qn, seq + 1 );
        }

        return qn.getName() + "/" + name;
    }

    protected
    final
    MingleIdentifier
    asIdentifier( byte[] arr )
        throws Exception
    {
        if ( arr == null ) return null;
        
        return MingleBinReader.create( arr ).readIdentifier();
    }

    protected
    final
    MingleIdentifier
    asIdentifier( MingleSymbolMap m,
                  String fld )
        throws Exception
    {
        return asIdentifier( mapGet( m, fld, byte[].class ) );
    }

    protected
    final
    MingleNamespace
    asNamespace( byte[] arr )
        throws Exception
    {
        if ( arr == null ) return null;

        return MingleBinReader.create( arr ).readNamespace();
    }

//        private
//        List< MingleIdentifier >
//        asIdentifierList( MingleList ml )
//            throws Exception
//        {
//            if ( ml == null ) return null;
//
//            List< MingleIdentifier > res = Lang.newList();
//
//            for ( MingleValue mv : ml ) {
//                res.add( asIdentifier( ( (MingleBuffer) mv ).array() ) );
//            }
//
//            return res;
//        }

    protected
    final
    QualifiedTypeName
    asQname( byte[] arr )
        throws Exception
    {
        if ( arr == null ) return null;

        return MingleBinReader.create( arr ).readQualifiedTypeName();
    }

    protected
    final
    MingleTypeReference
    asTypeReference( byte[] arr )
        throws Exception
    {
        if ( arr == null ) return null;

        return MingleBinReader.create( arr ).readTypeReference();
    }

    protected
    final
    ObjectPath< MingleIdentifier >
    asIdentifierPath( MingleList ml )
        throws Exception
    {
        if ( ml == null ) return null;

        ObjectPath< MingleIdentifier > res = ObjectPath.getRoot();

        for ( MingleValue mv : ml ) {
            if ( mv instanceof MingleUint64 ) {
                MingleUint64 i = (MingleUint64) mv;
                res = res.startImmutableList( i.intValue() );
            } else if ( mv instanceof MingleBuffer ) {
                MingleBuffer b = (MingleBuffer) mv;
                res = res.descend( asIdentifier( b.array() ) );
            } else {
                state.failf( "unhandled path element: %s", 
                    Mingle.inspect( mv ) );
            }
        }

        return res;
    }

    protected
    final
    ObjectPath< MingleIdentifier >
    asIdentifierPath( MingleSymbolMap m,
                      String fld )
        throws Exception
    {
        return asIdentifierPath( mapGet( m, fld, MingleList.class ) );
    }

    private
    void
    setEventStartStruct( MingleReactorEvent ev,
                         MingleSymbolMap map )
        throws Exception
    {
        byte[] arr = mapExpect( map, "type", byte[].class );
        ev.setStartStruct( asQname( arr ) );
    }

    private
    void
    setEventStartList( MingleReactorEvent ev,
                       MingleSymbolMap map )
        throws Exception
    {
        byte[] arr = mapExpect( map, "type", byte[].class );
        ev.setStartList( (ListTypeReference) asTypeReference( arr ) );
    }

    private
    void
    setEventStartField( MingleReactorEvent ev,
                        MingleSymbolMap map )
        throws Exception
    {
        byte[] arr = mapExpect( map, "field", byte[].class );
        ev.setStartField( asIdentifier( arr ) );
    }

    protected
    final
    MingleReactorEvent
    asReactorEvent( MingleStruct ms )
        throws Exception
    {
        inputs.notNull( ms, "ms" );

        MingleReactorEvent res = new MingleReactorEvent();

        String evName = ms.getType().getName().toString();
        MingleSymbolMap map = ms.getFields();

        if ( evName.equals( "StructStartEvent" ) ) {
            setEventStartStruct( res, map );
        } else if ( evName.equals( "FieldStartEvent" ) ) {
            setEventStartField( res, map );
        } else if ( evName.equals( "MapStartEvent" ) ) {
            res.setStartMap();
        } else if ( evName.equals( "ListStartEvent" ) ) {
            setEventStartList( res, map );
        } else if ( evName.equals( "EndEvent" ) ) {
            res.setEnd();
        } else if ( evName.equals( "ValueEvent" ) ) {
            res.setValue( mapExpect( map, "val", MingleValue.class ) );
        } else {
            state.failf( "unhandled event: %s", evName );
        }

        return res;
    }

    protected
    final
    List< MingleReactorEvent >
    asReactorEvents( MingleList ml )
        throws Exception
    {
        inputs.notNull( ml, "ml" );

        List< MingleReactorEvent > res = Lang.newList();

        for ( MingleValue mv : ml ) {
            res.add( asReactorEvent( (MingleStruct) mv ) );
        }

        return res;
    }

    protected
    final
    List< MingleReactorEvent >
    asReactorEvents( MingleSymbolMap m,
                     String fld )
        throws Exception
    {
        inputs.notNull( m, "m" );
        inputs.notNull( fld, "fld" );

        return asReactorEvents( mapExpect( m, fld, MingleList.class ) );
    }

    protected
    final
    MingleReactorException
    asReactorException( MingleSymbolMap map )
    {
        inputs.notNull( map, "map" );
        return new MingleReactorException( mapExpectString( map, "message" ) );
    }

    private
    boolean
    isList( MingleValue mv,
            ListTypeReference lt )
    {
        if ( ! ( mv instanceof MingleList ) ) return false;

        MingleList ml = (MingleList) mv;
        return ml.type().equals( lt );
    }

    private
    EventExpectation
    asEventExpectation( MingleStruct ms )
        throws Exception
    {
        MingleSymbolMap map = ms.getFields();

        MingleReactorEvent event = asReactorEvent(
            mapExpect( map, "event", MingleStruct.class ) );

        ObjectPath< MingleIdentifier > path = asIdentifierPath( map, "path" );

        return new EventExpectation( event, path );
    }

    private
    Queue< EventExpectation >
    asEventExpectationQueue( MingleList ml )
        throws Exception
    {
        if ( ml == null ) return null;

        Queue< EventExpectation > res = Lang.newQueue();

        for ( MingleValue mv : ml ) {
            res.add( asEventExpectation( (MingleStruct) mv ) );
        }

        return res;
    }

    protected
    final
    Object
    asFeedSource( MingleSymbolMap map,
                  String fld )
        throws Exception
    {
        MingleValue mv = mapGetValue( map, fld );

        if ( mv == null ) {
            return null;
        } else if ( isList( mv, TYPE_EVENT_LIST ) ) {
            return asReactorEvents( (MingleList) mv );
        } else if ( isList( mv, TYPE_EVENT_EXPECTATION_LIST ) ) {
            return asEventExpectationQueue( (MingleList) mv );
        } else {
            return mv;
        }
    }

    private
    TestException
    asTestException( MingleSymbolMap m )
        throws Exception
    {
        return new TestException(
            mapGetString( m, "message" ),
            asIdentifierPath( m, "location" )
        );
    }

    private
    MingleUnrecognizedFieldException
    asUnrecognizedFieldException( MingleSymbolMap m )
        throws Exception
    {
        return new MingleUnrecognizedFieldException(
            asIdentifier( m, "field" ),
            asIdentifierPath( m, "location" )
        );
    }

    // subclasses can override, calling super.asError() as their default return
    // val 
    protected
    Throwable
    asError( MingleStruct ms )
        throws Exception
    {
        if ( ms.getType().equals( QNAME_TEST_ERROR ) ) {
            return asTestException( ms.getFields() );
        } else if ( ms.getType().equals( QNAME_REACTOR_ERROR ) ) {
            return asReactorException( ms.getFields() );
        } else if ( ms.getType().equals( QNAME_UNRECOGNIZED_FIELD_ERROR ) ) {
            return asUnrecognizedFieldException( ms.getFields() );
        } else throw state.failf( "unhandled error: %s", ms.getType() );
    }

    protected
    final
    void
    setOptError( AbstractReactorTest t,
                 MingleValue err )
        throws Exception
    {
        if ( err != null ) t.expectFailure( asError( (MingleStruct) err ) );
    }

    protected
    abstract
    T
    convertReactorTest( MingleStruct ms )
        throws Exception;

    protected
    final
    T
    convertStruct( MingleStruct ms )
        throws Exception
    {
        if ( ! ms.getType().getNamespace().equals( testNs ) ) return null;
        return convertReactorTest( ms );
    }
}
