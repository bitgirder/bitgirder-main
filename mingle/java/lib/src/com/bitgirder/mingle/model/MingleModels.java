package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.PatternHelper;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ImmutableListPath;
import com.bitgirder.lang.path.ObjectPaths;
import com.bitgirder.lang.path.ObjectPathFormatter;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.mingle.parser.MingleParsers;

import com.bitgirder.parser.SyntaxException;

import java.nio.ByteBuffer;

import java.io.IOException;

import java.util.Map;
import java.util.Set;
import java.util.HashSet;
import java.util.Arrays;
import java.util.List;
import java.util.Date;
import java.util.GregorianCalendar;
import java.util.TimeZone;
import java.util.Iterator;
import java.util.Collections;
import java.util.Comparator;

import java.util.regex.Pattern;

import java.sql.Timestamp;

public
final
class MingleModels
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static ObjectPathFormatter< MingleIdentifier > 
        PATH_FORMATTER = new PathFormatterImpl();

    public final static MingleNamespace NS_MINGLE_CORE =
        MingleNamespace.create( "mingle:core@v1" );

    public final static AtomicTypeReference TYPE_REF_MINGLE_STRING =
        AtomicTypeReference.create(
            MingleTypeName.create( "String" ).resolveIn( NS_MINGLE_CORE ) );

    public final static AtomicTypeReference TYPE_REF_MINGLE_ENUM =
        AtomicTypeReference.create(
            MingleTypeName.create( "Enum" ).resolveIn( NS_MINGLE_CORE ) );

    public final static AtomicTypeReference TYPE_REF_MINGLE_INT64 =
        AtomicTypeReference.create(
            MingleTypeName.create( "Int64" ).resolveIn( NS_MINGLE_CORE ) );

    public final static AtomicTypeReference TYPE_REF_MINGLE_INT32 =
        AtomicTypeReference.create(
            MingleTypeName.create( "Int32" ).resolveIn( NS_MINGLE_CORE ) );

    public final static AtomicTypeReference TYPE_REF_MINGLE_DOUBLE =
        AtomicTypeReference.create(
            MingleTypeName.create( "Double" ).resolveIn( NS_MINGLE_CORE ) );

    public final static AtomicTypeReference TYPE_REF_MINGLE_FLOAT =
        AtomicTypeReference.create(
            MingleTypeName.create( "Float" ).resolveIn( NS_MINGLE_CORE ) );

    public final static AtomicTypeReference TYPE_REF_MINGLE_BOOLEAN =
        AtomicTypeReference.create(
            MingleTypeName.create( "Boolean" ).resolveIn( NS_MINGLE_CORE ) );

    public final static AtomicTypeReference TYPE_REF_MINGLE_BUFFER =
        AtomicTypeReference.create(
            MingleTypeName.create( "Buffer" ).resolveIn( NS_MINGLE_CORE ) );

    public final static AtomicTypeReference TYPE_REF_MINGLE_TIMESTAMP =
        AtomicTypeReference.create(
            MingleTypeName.create( "Timestamp" ).resolveIn( NS_MINGLE_CORE ) );

    public final static AtomicTypeReference TYPE_REF_MINGLE_SYMBOL_MAP =
        AtomicTypeReference.create(
            MingleTypeName.create( "SymbolMap" ).resolveIn( NS_MINGLE_CORE ) );

    public final static AtomicTypeReference TYPE_REF_MINGLE_STRUCT =
        AtomicTypeReference.create(
            MingleTypeName.create( "Struct" ).resolveIn( NS_MINGLE_CORE ) );

    public final static AtomicTypeReference TYPE_REF_MINGLE_STRUCTURE =
        AtomicTypeReference.create(
            MingleTypeName.create( "Structure" ).resolveIn( NS_MINGLE_CORE ) );
    
    public final static AtomicTypeReference TYPE_REF_MINGLE_EXCEPTION =
        AtomicTypeReference.create(
            MingleTypeName.create( "Exception" ).resolveIn( NS_MINGLE_CORE ) );

    public final static AtomicTypeReference TYPE_REF_MINGLE_NULL =
        AtomicTypeReference.create(
            MingleTypeName.create( "Null" ).resolveIn( NS_MINGLE_CORE ) );

    public final static AtomicTypeReference TYPE_REF_MINGLE_VALUE =
        AtomicTypeReference.create(
            MingleTypeName.create( "Value" ).resolveIn( NS_MINGLE_CORE ) );
    
    public final static AtomicTypeReference TYPE_REF_MINGLE_IDENTIFIER =
        (AtomicTypeReference)
            MingleTypeReference.create( "mingle:model@v1/MingleIdentifier" );
    
    public final static AtomicTypeReference TYPE_REF_MINGLE_NAMESPACE =
        (AtomicTypeReference)
            MingleTypeReference.create( "mingle:model@v1/MingleNamespace" );
    
    public final static AtomicTypeReference TYPE_REF_QUALIFIED_TYPE_NAME =
        (AtomicTypeReference)
            MingleTypeReference.create( "mingle:model@v1/QualifiedTypeName" );
    
    public final static AtomicTypeReference TYPE_REF_MINGLE_TYPE_REFERENCE =
        (AtomicTypeReference)
            MingleTypeReference.create( "mingle:model@v1/MingleTypeReference" );
    
    public final static AtomicTypeReference TYPE_REF_MINGLE_IDENTIFIED_NAME =
        (AtomicTypeReference)
            MingleTypeReference.
                create( "mingle:model@v1/MingleIdentifiedName" );

    final static ListTypeReference TYPE_REF_MINGLE_VALUE_LIST =
        ListTypeReference.create( TYPE_REF_MINGLE_VALUE, true );

    private final static Map< MingleTypeName, Class< ? extends MingleValue > >
        JAVA_TYPE_NAMES;
    
    public final static AtomicTypeReference TYPE_REF_VALIDATION_EXCEPTION =
        AtomicTypeReference.create(
            MingleTypeName.
                create( "ValidationException" ).resolveIn( NS_MINGLE_CORE ) );
    
    public final static AtomicTypeReference TYPE_REF_TYPE_CAST_EXCEPTION =
        AtomicTypeReference.create(
            MingleTypeName.
                create( "TypeCastException" ).resolveIn( NS_MINGLE_CORE ) );

    private final static Map< MingleTypeReference, MingleValueExchanger >
        EXCHANGERS;

    private final static Set< Class< ? > > INTEGRAL_JAVA_TYPES;
    private final static Set< Class< ? > > DECIMAL_JAVA_TYPES;

    private final static MingleList EMPTY_LIST =
        new MingleList.Builder().build();

    private final static MingleSymbolMap EMPTY_SYMBOL_MAP =
        symbolMapBuilder().build();

    private final static Pattern DESCENT_PATH_DOT_SPLITTER =
        PatternHelper.compile( "\\." );

    private final static Pattern DESCENT_PATH_LIST_INDEX_MATCHER =
        PatternHelper.compile( "\\d+" );

    private final static ValueErrorFactory DEFAULT_ERR_FACT =
        new ValueErrorFactory()
        {
            public
            RuntimeException
            createFail( ObjectPath< String > path,
                        String msg )
            {
                StringBuilder sb = new StringBuilder();
        
                if ( path != null )
                {
                    ObjectPaths.
                        appendFormat( path, ObjectPaths.DOT_FORMATTER, sb );

                    sb.append( ": " );
                }
                
                sb.append( msg );
        
                return new IllegalArgumentException( sb.toString() );
            }
        };

    private MingleModels() {}

    public
    static
    MingleStructBuilder
    structBuilder()
    {
        return new MingleStructBuilder();
    }

    public
    static
    MingleExceptionBuilder
    exceptionBuilder()
    {
        return new MingleExceptionBuilder();
    }

    public
    static
    StandaloneMingleSymbolMapBuilder
    symbolMapBuilder()
    {
        return StandaloneMingleSymbolMapBuilder.create();
    }

    public
    static
    MingleStruct
    createMingleStruct( final AtomicTypeReference typeRef,
                        final MingleSymbolMap msm )
    {
        inputs.notNull( typeRef, "typeRef" );
        inputs.notNull( msm, "msm" );

        return new MingleStruct()
        {
            public AtomicTypeReference getType() { return typeRef; }
            public MingleSymbolMap getFields() { return msm; }
        };
    }

    public
    static
    Set< MingleIdentifier >
    getKeySet( MingleSymbolMap m )
    {
        inputs.notNull( m, "m" );

        if ( m instanceof DefaultMingleSymbolMap ) 
        {
            return ( (DefaultMingleSymbolMap) m ).getKeySet();
        }
        else
        {
            Set< MingleIdentifier > res = Lang.newSet();
            for ( MingleIdentifier id : m.getFields() ) res.add( id );

            return Lang.unmodifiableSet( res );
        }
    }

    public
    static
    MingleString
    asMingleString( CharSequence str )
    {
        inputs.notNull( str, "str" );

        return ( str instanceof MingleString )
            ? (MingleString) str : new MingleString( str );
    }

    public
    static
    MingleBoolean
    asMingleBoolean( boolean b )
    {
        return MingleBoolean.valueOf( b );
    }

    public
    static
    MingleInt64
    asMingleInt64( long l )
    {
        return new MingleInt64( l );
    }

    public
    static
    MingleInt32
    asMingleInt32( int i )
    {
        return new MingleInt32( i ); 
    }

    public
    static
    MingleDouble
    asMingleDouble( double d )
    {
        return new MingleDouble( d );
    }

    public
    static
    MingleFloat
    asMingleFloat( float f )
    {
        return new MingleFloat( f );
    }

    public
    static
    MingleStruct
    asMingleStruct( final AtomicTypeReference typ,
                    final MingleSymbolMap flds )
    {
        inputs.notNull( typ, "typ" );
        inputs.notNull( flds, "flds" );

        return new MingleStruct() 
        {
            public AtomicTypeReference getType() { return typ; }
            public MingleSymbolMap getFields() { return flds; }
        };
    }

    public
    static
    MingleException
    asMingleException( final AtomicTypeReference typ,
                       final MingleSymbolMap flds )
    {
        inputs.notNull( typ, "typ" );
        inputs.notNull( flds, "flds" );

        return new MingleException()
        {
            public AtomicTypeReference getType() { return typ; }
            public MingleSymbolMap getFields() { return flds; }
        };
    }

    public
    static
    MingleBuffer
    asMingleBuffer( ByteBuffer data )
    {
        return new MingleBuffer( inputs.notNull( data, "data" ) );
    }

    public
    static
    MingleBuffer
    asMingleBuffer( byte[] data )
    {
        return 
            new MingleBuffer( 
                ByteBuffer.wrap( inputs.notNull( data, "data" ) ) );
    }

    static
    boolean
    isIntegralType( Class< ? > cls )
    {
        state.notNull( cls, "cls" );
        return INTEGRAL_JAVA_TYPES.contains( cls );
    }

    // Can make this faster when needed -- caller must ensure that the
    // toString() representation of num is indeed a valid integer (of any size).
    // That is, calling 'new BigInteger( num.toString() )' would always be
    // valid.
    //
    // Any object for which isIntegralType() returns true will be valid to pass
    // in to this method
    private
    static
    MingleValue
    asMingleIntegral( Object num )
    {
        state.notNull( num, "num" );
        
        if ( num instanceof Long )
        {
            return asMingleInt64( ( (Long) num ).longValue() );
        }
        else
        {
            int i;

            if ( num instanceof Character )
            {
                i = (int) ( (Character) num ).charValue();
            }
            else i = ( (Number) num ).intValue();

            return asMingleInt32( i );
        }
    }

    static
    boolean
    isDecimalType( Class< ? > cls )
    {
        return DECIMAL_JAVA_TYPES.contains( cls );
    }

    private
    static
    MingleValue
    asMingleDecimal( Object num )
    {
        state.notNull( num, "num" );

        if ( num instanceof Double ) return asMingleDouble( (Double) num );
        else if ( num instanceof Float ) return asMingleFloat( (Float) num );
        else 
        {
            throw state.createFail( 
                "Invalid number type:", num.getClass().getName() );
        }
    }

    private
    static
    < T >
    ObjectPath< T >
    descend( ObjectPath< T > path,
             T key )
    {
        return path == null ? null : path.descend( key );
    }

    private
    static
    < T >
    ImmutableListPath< T >
    startList( ObjectPath< T > path )
    {
        return path == null ? null : path.startImmutableList();
    }

    public
    static
    interface ValueErrorFactory
    {
        public
        RuntimeException
        createFail( ObjectPath< String > path,
                    String msg );
    }

    private
    static
    < T >
    RuntimeException
    createFail( ValueErrorFactory ef,
                ObjectPath< String > path,
                Object... arr )
    {
        return ef.createFail( path, Strings.join( " ", arr ).toString() );
    }

    public
    static
    MingleList
    asMingleList( List< ? > l,
                  ImmutableListPath< String > path,
                  ValueErrorFactory ef )
    {
        inputs.notNull( l, "l" );

        MingleList.Builder b = new MingleList.Builder();

        for ( Object obj : l ) 
        {
            b.add( asMingleValueImpl( obj, path, ef ) );
            if ( path != null ) path = path.next();
        }

        return b.build();
    }

    private
    static
    MingleSymbolMap
    asMingleSymbolMap( Map< ?, ? > m,
                       ObjectPath< String > path,
                       ValueErrorFactory ef )
    {
        MingleSymbolMapBuilder b = symbolMapBuilder();

        for ( Map.Entry< ?, ? > e : m.entrySet() )
        {
            Object key = e.getKey();
            if ( key == null ) throw createFail( ef, path, "map has null key" );

            MingleValue val = 
                asMingleValueImpl( 
                    e.getValue(), descend( path, key.toString() ), ef );

            if ( key instanceof CharSequence ) b.set( (CharSequence) key, val );
            else if ( key instanceof MingleIdentifier )
            {
                b.set( (MingleIdentifier) key, val );
            }
            else 
            {
                throw createFail( ef, path,
                    "Encountered a non-CharSequence map key:", key, 
                    key == null ? "" : "(" + key.getClass().getName() + ")" );
            }
        }

        return b.build();
    }

    private
    static
    MingleTimestamp.Builder
    initTimestampBuilder( GregorianCalendar c )
    {
        return
            new MingleTimestamp.Builder().
                setFromCalendar( c );
    }
    
    private
    static
    MingleTimestamp
    asMingleTimestamp( Date d )
    {
        GregorianCalendar c = 
            new GregorianCalendar( TimeZone.getTimeZone( "GMT+00:00" ) );
        c.setTimeInMillis( d.getTime() );

        MingleTimestamp.Builder b = initTimestampBuilder( c );

        if ( d instanceof Timestamp )
        {
            b.setNanos( ( (Timestamp) d ).getNanos() );
        }

        return b.build();
    }

    private
    static
    MingleTimestamp
    asMingleTimestamp( GregorianCalendar c )
    {
        return initTimestampBuilder( c ).build();
    }

    // obj or path could be null, but not errFact
    private
    static
    MingleValue
    asMingleValueImpl( Object obj,
                       ObjectPath< String > path,
                       ValueErrorFactory errFact )
    {
        if ( obj == null ) return MingleNull.getInstance();
        else if ( obj instanceof MingleValue ) return (MingleValue) obj;
        else
        {
            Class< ? > cls = obj.getClass();

            if ( isIntegralType( cls ) ) return asMingleIntegral( obj );
            else if ( isDecimalType( cls ) ) return asMingleDecimal( obj );
            else if ( obj instanceof byte[] )
            {
                return asMingleBuffer( (byte[]) obj );
            }
            else if ( obj instanceof ByteBuffer )
            {
                return asMingleBuffer( (ByteBuffer) obj );
            }
            else if ( obj instanceof CharSequence )
            {
                return asMingleString( (CharSequence) obj );
            }
            else if ( obj instanceof Boolean )
            {
                return asMingleBoolean( (Boolean) obj );
            }
            else if ( obj instanceof List< ? > )
            {
                return 
                    asMingleList( (List< ? >) obj, startList( path ), errFact );
            }
            else if ( obj instanceof Object[] )
            {
                List< ? > l = Arrays.asList( (Object[]) obj );
                return asMingleList( l, startList( path ), errFact );
            }
            else if ( obj instanceof Map< ?, ? > )
            {
                return asMingleSymbolMap( (Map< ?, ? >) obj, path, errFact );
            }
            else if ( obj instanceof Date )
            {
                return asMingleTimestamp( (Date) obj );
            }
            else if ( obj instanceof GregorianCalendar )
            {
                return asMingleTimestamp( (GregorianCalendar) obj );
            }
            else
            {
                throw createFail( 
                    errFact,
                    path,
                    "Can't convert instance of", cls, "to mingle value" );
            }
        }
    }

    public
    static
    MingleValue
    asMingleValue( Object obj,
                   ObjectPath< String > path,
                   ValueErrorFactory errFact )
    {
        inputs.notNull( path, "path" );
        inputs.notNull( errFact, "errFact" );

        return asMingleValueImpl( obj, path, errFact );
    }

    public
    static
    MingleValue
    asMingleValue( Object obj,
                   ObjectPath< String > path )
    {
        inputs.notNull( path, "path" );
        return asMingleValueImpl( obj, path, DEFAULT_ERR_FACT );
    }

    public
    static
    MingleValue
    asMingleValue( Object obj )
    {
        return asMingleValueImpl( obj, null, DEFAULT_ERR_FACT );
    }

    public
    static
    MingleValidator
    createValidator( ObjectPath< MingleIdentifier > path )
    {
        return new DefaultMingleValidator( inputs.notNull( path, "path" ) );
    }

    public
    static
    MingleValidator
    createValidator()
    {
        return createValidator( ObjectPath.< MingleIdentifier > getRoot() );
    }

    public
    static
    MingleInvocationValidator
    createInvocationValidator( ObjectPath< MingleIdentifier > path )
    {
        return new MingleInvocationValidator( inputs.notNull( path, "path" ) );
    }

    public
    static
    MingleInvocationValidator
    createInvocationValidator()
    {
        return 
            createInvocationValidator( 
                ObjectPath.< MingleIdentifier >getRoot() );
    }

    private
    static
    StringBuilder
    doAppendInspection( StringBuilder sb,
                        MingleString ms )
    {
        Lang.appendRfc4627String( sb, ms );
        return sb;
    }

    private
    static
    StringBuilder
    doAppendInspection( StringBuilder sb,
                        MingleEnum me )
    {
        MingleIdentifier constant = me.getValue();
        MingleIdentifierFormat fmt = MingleIdentifierFormat.LC_UNDERSCORE;

        return
            sb.append( me.getType().getExternalForm() ).
               append( '.' ).
               append( MingleModels.format( constant, fmt ) );
    }

    private
    static
    StringBuilder
    doAppendInspection( StringBuilder sb,
                        MingleList ml )
    {
        Iterator< MingleValue > it = ml.iterator();

        sb.append( '[' );

        if ( it.hasNext() )
        {
            sb.append( ' ' );

            while ( it.hasNext() )
            {
                doAppendInspection( sb, it.next() );
                if ( it.hasNext() ) sb.append( ',' );
                sb.append( ' ' );
            }
        }

        return sb.append( ']' );
    }

    private
    static
    StringBuilder
    doAppendInspection( StringBuilder sb,
                        MingleStructure ms )
    {
        sb.append( getType( ms ).getExternalForm() );
        sb.append( ":" );

        return doAppendInspection( sb, ms.getFields() );
    }

    private
    static
    StringBuilder
    doAppendInspection( StringBuilder sb,
                        MingleSymbolMap msm )
    {
        sb.append( '{' );

        Iterator< MingleIdentifier > it = msm.getFields().iterator();

        if ( it.hasNext() )
        {
            sb.append( ' ' );

            while ( it.hasNext() )
            {
                MingleIdentifier fld = it.next();

                sb.append( fld.getExternalForm() ).
                   append( ": " );
                
                doAppendInspection( sb, msm.get( fld ) );

                if ( it.hasNext() ) sb.append( ',' );
                sb.append( ' ' );
            }
        }

        return sb.append( '}' );
    }

    private
    static
    StringBuilder
    doAppendInspection( StringBuilder sb,
                        MingleValue mv )
    {
        if ( mv instanceof MingleString )
        {
            return doAppendInspection( sb, (MingleString) mv );
        }
        else if ( mv instanceof MingleEnum )
        {
            return doAppendInspection( sb, (MingleEnum) mv );
        }
        else if ( mv instanceof MingleList )
        {
            return doAppendInspection( sb, (MingleList) mv );
        }
        else if ( mv instanceof MingleStructure )
        {
            return doAppendInspection( sb, (MingleStructure) mv );
        }
        else if ( mv instanceof MingleSymbolMap )
        {
            return doAppendInspection( sb, (MingleSymbolMap) mv );
        }
        else if ( mv instanceof MingleNull) 
        {
            sb.append( "null" );
            return sb;
        }
        else 
        {
            sb.append( mv );
            return sb;
        }
    }

    public
    static
    StringBuilder
    appendInspection( StringBuilder sb,
                      MingleValue mv )
    {
        return
            doAppendInspection(
                inputs.notNull( sb, "sb" ), inputs.notNull( mv, "mv" ) );
    }

    public
    static
    CharSequence
    inspect( MingleValue mv )
    {
        inputs.notNull( mv, "mv" );

        return doAppendInspection( new StringBuilder(), mv );
    }

    // does null checking of resp for public frontends
    private
    static
    StringBuilder
    doAppendInspection( StringBuilder sb,
                        MingleServiceResponse resp )
    {
        inputs.notNull( resp, "resp" );
        
        sb.append( "{ " );

        if ( resp.isOk() ) 
        {
            sb.append( "result: " );
            appendInspection( sb, resp.getResult() );
        }
        else
        {
            sb.append( "exception: " );
            appendInspection( sb, resp.getException() );
        }

        return sb.append( " }" );
    }

    public
    static
    StringBuilder
    appendInspection( StringBuilder sb,
                      MingleServiceResponse resp )
    {
        return doAppendInspection( inputs.notNull( sb, "sb" ), resp );
    }

    public
    static
    CharSequence
    inspect( MingleServiceResponse resp )
    {
        return doAppendInspection( new StringBuilder(), resp );
    }

    private
    static
    StringBuilder
    doAppendInspection( StringBuilder sb,
                        MingleServiceRequest req,
                        boolean showAuthentication )
    {
        inputs.notNull( req, "req" );

        sb.append( "{ " ).
           append( "namespace: " ).
           append( req.getNamespace().getExternalForm() ).
           append( ", service: " ).
           append( req.getService().getExternalForm() ).
           append( ", operation: " ).
           append( req.getOperation().getExternalForm() );

        if ( showAuthentication )
        {
            sb.append( ", authentication: " );
            doAppendInspection( sb, req.getAuthentication() );
        }

        sb.append( ", parameters: " );
        doAppendInspection( sb, req.getParameters() );

        return sb.append( " }" );
    }

    public
    static
    StringBuilder
    appendInspection( StringBuilder sb,
                      MingleServiceRequest req,
                      boolean showAuthentication )
    {
        return
            doAppendInspection( 
                inputs.notNull( sb, "sb" ), req, showAuthentication );
    }

    public
    static
    CharSequence
    inspect( MingleServiceRequest req,
             boolean showAuthentication )
    {
        return 
            doAppendInspection( new StringBuilder(), req, showAuthentication );
    }

    private
    static
    MingleTypeCastException
    castException( MingleTypeReference typExpct,
                   MingleValue valActual,
                   ObjectPath< MingleIdentifier > path )
    {
        return
            new MingleTypeCastException( 
                typExpct, 
                typeReferenceOf( valActual ), 
                path 
            );
    }

    private
    static
    class DescentNode
    {
        private Map< Object, DescentNode > children;

        private boolean sawFields;
        private boolean sawList;

        private String terminal;

        private AtomicTypeReference typeRef;

        private
        Map< Object, DescentNode >
        children()
        {
            if ( children == null ) children = Lang.newMap();
            return children;
        }

        // we allow sparse indices so have to account for the holes. In practice
        // holes are unlikely, but if we end up dealing with very sparse lists
        // we always have the option of building a small list facade on top of
        // the actual indices we see in children
        private
        Iterator< Map.Entry< Object, DescentNode > >
        sortListEntries()
        {
            List< Map.Entry< Object, DescentNode > > l = 
                Lang.newList( children.entrySet() );
            
            Collections.sort( l,
                new Comparator< Map.Entry< Object, DescentNode > >()
                {
                    public
                    int
                    compare( Map.Entry< Object, DescentNode > e1,
                             Map.Entry< Object, DescentNode > e2 )
                    {
                        Integer i1 = (Integer) e1.getKey();
                        Integer i2 = (Integer) e2.getKey();

                        return i1.compareTo( i2 );
                    }
                }
            );

            return l.iterator();
        }

        private
        void
        addListItems( Iterator< Map.Entry< Object, DescentNode > > it,
                      MingleList.Builder b )
        {
            // we increment ++i as part of the loop here to ensure that it is
            // always positioned at the next element (which may be a hole); at
            // the end of the loop body i is always equal to the index most
            // recently filled with a non-null element
            for ( int i = 0; it.hasNext(); ++i )
            {
                Map.Entry< Object, DescentNode > e = it.next();

                int next = (Integer) e.getKey(); 

                while ( i < next )
                {
                    b.add( MingleNull.getInstance() );
                    ++i;
                }

                b.add( e.getValue().buildMingleValue() );
            }
        }

        private
        MingleList
        buildList()
        {
            state.isTrue( 
                typeRef == null, "Lists do not support the 'type' element" );

            MingleList.Builder b = new MingleList.Builder();

            Iterator< Map.Entry< Object, DescentNode > > it = sortListEntries();

            addListItems( it, b );

            return b.build();
        }

        private
        MingleSymbolMap
        buildSymbolMap()
        {
            MingleSymbolMapBuilder b = symbolMapBuilder();

            for ( Map.Entry< Object, DescentNode > e : children().entrySet() )
            {
                b.set( 
                    (MingleIdentifier) e.getKey(), 
                    e.getValue().buildMingleValue() );
            }

            return b.build();
        }

        private
        MingleStruct
        buildStruct()
        {
            return createMingleStruct( typeRef, buildSymbolMap() );
        }

        private
        MingleValue
        buildMingleValue()
        {
            if ( terminal == null )
            {
                return sawList
                    ? buildList()
                    : typeRef == null ? buildSymbolMap() : buildStruct();
            }
            else return asMingleString( terminal );
        }

        @Override
        public
        String
        toString()
        {
            return 
                "[ sawFields: " + sawFields + 
                ", sawList: " + sawList + 
                ", terminal: " + terminal +
                ", typeRef: " + typeRef +
                ", children: " + children + " ]";
        }
    }

    private
    static
    void
    assertPathHomogeneity( boolean fail,
                           String prefix )
    {
        inputs.isFalse( fail, "Attempt to mix list and field nodes:", prefix );
    }

    // Returns either a MingleIdentifier of the field or an Integer of the list
    // index depending on the type of node. As a side effect this method fails
    // if it is detected that the node has mixed fields (some list indices and
    // some fields)
    private
    static
    Object
    getNodeElement( DescentNode n,
                    String s,
                    String path,
                    String prefix,
                    int pathEnd )
    {
        if ( DESCENT_PATH_LIST_INDEX_MATCHER.matcher( s ).matches() )
        {
            assertPathHomogeneity( n.sawFields, prefix );
            n.sawList = true;

            return Integer.valueOf( Integer.parseInt( s ) );
        }
        else
        {
            assertPathHomogeneity( n.sawList, prefix );
            n.sawFields = true;

            try { return MingleIdentifier.parse( s ); }
            catch ( SyntaxException se )
            {
                throw inputs.createFail(
                    "Invalid identifier '" + s + "' in path:", 
                    path.substring( 0, pathEnd ) );
            }
        }
    }

    private
    static
    DescentNode
    ensureInternalNode( DescentNode n,
                        String path,
                        int beg,
                        int end )
    {
        String s = path.substring( beg, end );

        // Everything up to the separator before s, if any
        String prefix = path.substring( 0, Math.max( 0, beg - 1 ) );

        Object elt = getNodeElement( n, s, path, prefix, end );

        DescentNode res = n.children().get( elt );

        if ( res == null ) 
        {
            res = new DescentNode();
            n.children().put( elt, res );
        }

        return res;
    }

    // Note that this is currently billed as a state check, not an input check,
    // since the only publicly allowable input is a string Map which, by its own
    // mappiness should ensure that there are no duplicates. So, if we detect
    // one then the assumption is that it is an internal programming error.
    private
    static
    void
    addTerminal( DescentNode n,
                 String terminal,
                 String errPath )
    {
        if ( n.terminal == null ) n.terminal = terminal;
        else
        {
            throw state.createFail(
                "Duplicate values for path '" + errPath + "', saw '" +
                n.terminal + "' and '" + terminal + "'" );
        }
    }

    private
    static
    AtomicTypeReference
    parseTypeRef( String path,
                  String refStr )
    {
        Throwable th = null;
        
        try 
        { 
            return (AtomicTypeReference) MingleTypeReference.parse( refStr );
        }
        catch ( SyntaxException se ) { th = se; }
        catch ( ClassCastException cce ) { th = cce; }

        throw
            new IllegalArgumentException(
                "Invalid type reference for '" + path + "': " + refStr,
                state.notNull( th ) );
    }

    private
    static
    void
    addType( DescentNode n,
             String path,
             int indx,
             String terminal )
    {
        String ctl = path.substring( indx );

        inputs.isTrue( ctl.equals( "type" ), "Invalid control argument:", ctl );

        if ( n.typeRef == null ) n.typeRef = parseTypeRef( path, terminal );
        else inputs.fail( "Multiple type references for '" + path + "'" );
    }

    // returns either the first ':' or '.' found, or path.length() if none
    // remain
    private
    static
    int
    getSegmentEnd( String path,
                   int indx )
    {
        int dotIndx = path.indexOf( '.', indx );
        int colIndx = path.indexOf( ':', indx );
 
        if ( dotIndx < 0 && colIndx < 0 ) return path.length();
        else
        {
            if ( dotIndx < 0 ) dotIndx = Integer.MAX_VALUE;
            if ( colIndx < 0 ) colIndx = Integer.MAX_VALUE;

            return Math.min( dotIndx, colIndx );
        }
    }

    private
    static
    void
    addNode( DescentNode n,
             String path,
             int indx,
             String terminal )
    {
        int end = getSegmentEnd( path, indx );

        n = ensureInternalNode( n, path, indx, end );

        if ( end == path.length() ) addTerminal( n, terminal, path );
        else
        {
            switch ( path.charAt( end ) )
            {
                case '.': addNode( n, path, end + 1, terminal ); break;
                case ':': addType( n, path, end + 1, terminal ); break;
                default: state.fail();
            }
        }
    }

    public
    static
    MingleValue
    stringMapToMingleValue( Map< String, String > strMap )
    {
        inputs.noneNull( strMap, "strMap" );

        DescentNode root = new DescentNode();

        for ( Map.Entry< String, String > e : strMap.entrySet() )
        {
            addNode( root, e.getKey(), 0, e.getValue() );
        }

        return root.buildMingleValue();
    }

    private
    static
    MingleString
    coerceMingleString( AtomicTypeReference typ,
                        MingleValue val,
                        MingleTypeReference outerTyp,
                        MingleValue outerVal,
                        ObjectPath< MingleIdentifier > path )
    {
        CharSequence res;

        if ( val instanceof MingleInt64 ||
             val instanceof MingleInt32 ||
             val instanceof MingleDouble ||
             val instanceof MingleFloat ||
             val instanceof MingleBoolean ) 
        {
            res = val.toString();
        }
        else if ( val instanceof MingleTimestamp )
        {
            res = ( (MingleTimestamp) val ).getRfc3339String();
        }
        else if ( val instanceof MingleBuffer )
        {
            res = ( (MingleBuffer) val ).asBase64String();
        }
        else if ( val instanceof MingleEnum )
        {
            res = ( (MingleEnum) val ).getValue().getExternalForm();
        }
        else throw castException( outerTyp, outerVal, path );

        return MingleModels.asMingleString( res );
    }

    private
    static
    MingleValidationException
    asMingleNumberFormatException( MingleString str,
                                   ObjectPath< MingleIdentifier > path )
    {
        return 
            MingleValidation.createFail( path, "Invalid number format:", str );
    }

    // There's nothing particularly inspiring about this long branchey method,
    // other than that it handles all types of numeric conversions in a single
    // place (and is meant to be covered by the CoercionTests in 
    // ModelTests.testNumericCoercionPermutations(). Ultimately we may find that
    // we'll want to short circuit some of the logic here to shorten common code
    // paths, but for now this should be perfectly sufficient and correct, which
    // is nice.
    private
    static
    MingleValue
    coerceMingleNum( Class< ? extends MingleValue > targCls,
                     MingleValue val,
                     MingleTypeReference outerTyp,
                     MingleValue outerVal,
                     ObjectPath< MingleIdentifier > path )
    {
        if ( val instanceof MingleString )
        {
            try
            {
                String str = val.toString();

                if ( targCls.equals( MingleInt64.class ) )
                {
                    return asMingleInt64( Long.parseLong( str ) );
                }
                else if ( targCls.equals( MingleInt32.class ) )
                {
                    return asMingleInt32( Integer.parseInt( str ) );
                }
                else if ( targCls.equals( MingleDouble.class ) )
                {
                    return asMingleDouble( Double.parseDouble( str ) );
                }
                else if ( targCls.equals( MingleFloat.class ) )
                {
                    return asMingleFloat( Float.parseFloat( str ) );
                }
                else throw state.createFail( targCls );
            }
            catch ( NumberFormatException nfe )
            {
                throw asMingleNumberFormatException( (MingleString) val, path );
            }
        }
        else if ( val instanceof MingleInt64 )
        {
            long l = ( (MingleInt64) val ).longValue();

            if ( targCls.equals( MingleInt32.class ) )
            {
                return asMingleInt32( (int) l );
            }
            else if ( targCls.equals( MingleDouble.class ) )
            {
                return asMingleDouble( (double) l );
            }
            else if ( targCls.equals( MingleFloat.class ) )
            {
                return asMingleFloat( (float) l );
            }
            else if ( targCls.equals( MingleString.class ) )
            { 
                return asMingleString( Long.toString( l ) );
            }
            else throw state.createFail( targCls );
        }
        else if ( val instanceof MingleInt32 )
        {
            int i = ( (MingleInt32) val ).intValue();

            if ( targCls.equals( MingleInt64.class ) )
            {
                return asMingleInt64( (long) i );
            }
            else if ( targCls.equals( MingleDouble.class ) )
            {
                return asMingleDouble( (double) i );
            }
            else if ( targCls.equals( MingleFloat.class ) )
            {
                return asMingleFloat( (float) i );
            }
            else if ( targCls.equals( MingleString.class ) )
            { 
                return asMingleString( Integer.toString( i ) );
            }
            else throw state.createFail( targCls );
        }
        else if ( val instanceof MingleDouble )
        {
            double d = ( (MingleDouble) val ).doubleValue();

            if ( targCls.equals( MingleInt64.class ) )
            {
                return asMingleInt64( (long) d );
            }
            else if ( targCls.equals( MingleInt32.class ) )
            {
                return asMingleInt32( (int) d );
            }
            else if ( targCls.equals( MingleFloat.class ) )
            {
                return asMingleFloat( (float) d );
            }
            else if ( targCls.equals( MingleString.class ) )
            { 
                return asMingleString( Double.toString( d ) );
            }
            else throw state.createFail( targCls );
        }
        else if ( val instanceof MingleFloat )
        {
            float f = ( (MingleFloat) val ).floatValue();

            if ( targCls.equals( MingleInt64.class ) )
            {
                return asMingleInt64( (long) f );
            }
            else if ( targCls.equals( MingleInt32.class ) )
            {
                return asMingleInt32( (int) f );
            }
            else if ( targCls.equals( MingleDouble.class ) )
            {
                return asMingleDouble( (double) f );
            }
            else if ( targCls.equals( MingleString.class ) )
            { 
                return asMingleString( Float.toString( f ) );
            }
            else throw state.createFail( targCls );
        }
        else throw castException( outerTyp, outerVal, path );
    }

    // We may decide to also add coercion from integrals here at some point to
    // handle c-style booleans
    private
    static
    MingleBoolean
    coerceMingleBoolean( MingleValue val,
                         MingleTypeReference outerTyp,
                         MingleValue outerVal,
                         ObjectPath< MingleIdentifier > path )
    {
        if ( val instanceof MingleString )
        {
            try { return MingleBoolean.parse( (MingleString) val ); }
            catch ( SyntaxException se )
            {
                throw new MingleValidationException(
                    "Invalid boolean string: " + val, path );
            }
        }
        else throw castException( outerTyp, outerVal, path );
    }

    private
    static
    MingleTimestamp
    coerceMingleTimestamp( MingleValue val,
                           MingleTypeReference outerTyp,
                           MingleValue outerVal,
                           ObjectPath< MingleIdentifier > path )
    {
        if ( val instanceof MingleString )
        {
            try { return MingleParsers.parseTimestamp( (MingleString) val ); }
            catch ( SyntaxException se )
            {
                throw new MingleValidationException(
                    "Invalid timestamp string: " + val, path );
            }
        }
        else throw castException( outerTyp, outerVal, path );
    }

    private
    static
    MingleBuffer
    coerceMingleBuffer( MingleValue val,
                        MingleTypeReference outerTyp,
                        MingleValue outerVal,
                        ObjectPath< MingleIdentifier > path )
    {
        if ( val instanceof MingleString )
        {
            try { return MingleBuffer.fromBase64String( (MingleString) val ); }
            catch ( IOException ioe )
            {
                throw new MingleValidationException(
                    "Invalid base64 buffer (" + ioe.getMessage() + ")", path );
            }
        }
        else throw castException( outerTyp, outerVal, path );
    }

    private
    static
    MingleSymbolMap
    coerceMingleSymbolMap( MingleValue val,
                           MingleTypeReference outerTyp,
                           MingleValue outerVal,
                           ObjectPath< MingleIdentifier > path )
    {
        if ( val instanceof MingleStructure )
        {
            return ( (MingleStructure) val ).getFields();
        }
        else throw castException( outerTyp, outerVal, path );
    }

    private
    static
    MingleStruct
    coerceMingleStruct( MingleValue val,
                        MingleTypeReference outerTyp,
                        MingleValue outerVal,
                        ObjectPath< MingleIdentifier > path )
    {
        if ( val instanceof MingleStruct ) return (MingleStruct) val;
        else throw castException( outerTyp, outerVal, path );
    }

    private
    final
    static
    class CoercedMingleException
    implements MingleException
    {
        private final MingleStruct struct;

        private 
        CoercedMingleException( MingleStruct struct )
        {
            this.struct = struct; 
        }

        public AtomicTypeReference getType() { return struct.getType(); }
    
        public MingleSymbolMap getFields() { return struct.getFields(); }
    }

    private
    static
    MingleException
    coerceMingleException( MingleValue val,
                           MingleTypeReference outerTyp,
                           MingleValue outerVal,
                           ObjectPath< MingleIdentifier > path )
    {
        if ( val instanceof MingleException ) return (MingleException) val;
        else if ( val instanceof MingleStruct )
        {
            return new CoercedMingleException( (MingleStruct) val );
        }
        else throw castException( outerTyp, outerVal, path );
    } 

    private
    static
    MingleEnum
    coerceMingleEnum( MingleValue val,
                      MingleTypeReference outerTyp,
                      MingleValue outerVal,
                      ObjectPath< MingleIdentifier > path )
    {
        if ( val instanceof MingleEnum ) return (MingleEnum) val;
        if ( val instanceof MingleString )
        {
            try { return MingleEnum.parse( (MingleString) val ); }
            catch ( SyntaxException se )
            {
                throw new MingleValidationException(
                    "Invalid enum constant: " + val, path );
            }
        }
        else throw castException( outerTyp, outerVal, path );
    }

    private
    static
    boolean
    isBaseTypeOf( AtomicTypeReference typ,
                  AtomicTypeReference test )
    {
        return typ.getName().equals( test.getName() );
    }

    private
    static
    boolean
    isBaseInstance( AtomicTypeReference typ,
                    MingleValue mv )
    {
        if ( typ.equals( TYPE_REF_MINGLE_VALUE ) ) return true;
        else
        {
            MingleTypeReference valTyp = typeReferenceOf( mv );

            if ( valTyp instanceof AtomicTypeReference )
            {
                AtomicTypeReference atr = (AtomicTypeReference) valTyp;
                return atr.getName().equals( typ.getName() );
            }
            else return false;
        }
    }

    private
    static
    MingleValue
    getAtomicValueBase( AtomicTypeReference typ,
                        MingleValue mv,
                        MingleTypeReference outerTyp,
                        MingleValue outerVal,
                        ObjectPath< MingleIdentifier > path )
    {
        if ( isBaseInstance( typ, mv ) ) return mv;
        else if ( isBaseTypeOf( typ, TYPE_REF_MINGLE_STRING ) )
        {
            return coerceMingleString( typ, mv, outerTyp, outerVal, path );
        }
        else if ( isBaseTypeOf( typ, TYPE_REF_MINGLE_INT64 ) )
        {
            return 
                coerceMingleNum( 
                    MingleInt64.class, mv, outerTyp, outerVal, path );
        }
        else if ( isBaseTypeOf( typ, TYPE_REF_MINGLE_INT32 ) )
        {
            return 
                coerceMingleNum( 
                    MingleInt32.class, mv, outerTyp, outerVal, path );
        }
        else if ( isBaseTypeOf( typ, TYPE_REF_MINGLE_DOUBLE ) )
        {
            return 
                coerceMingleNum( 
                    MingleDouble.class, mv, outerTyp, outerVal, path );
        }
        else if ( isBaseTypeOf( typ, TYPE_REF_MINGLE_FLOAT ) )
        {
            return 
                coerceMingleNum( 
                    MingleFloat.class, mv, outerTyp, outerVal, path );
        }
        else if ( isBaseTypeOf( typ, TYPE_REF_MINGLE_BOOLEAN ) )
        {
            return coerceMingleBoolean( mv, outerTyp, outerVal, path );
        }
        else if ( isBaseTypeOf( typ, TYPE_REF_MINGLE_TIMESTAMP ) )
        {
            return coerceMingleTimestamp( mv, outerTyp, outerVal, path );
        }
        else if ( isBaseTypeOf( typ, TYPE_REF_MINGLE_BUFFER ) )
        {
            return coerceMingleBuffer( mv, outerTyp, outerVal, path );
        }
        else if ( isBaseTypeOf( typ, TYPE_REF_MINGLE_SYMBOL_MAP ) )
        {
            return coerceMingleSymbolMap( mv, outerTyp, outerVal, path );
        }
        else if ( isBaseTypeOf( typ, TYPE_REF_MINGLE_STRUCT ) )
        {
            return coerceMingleStruct( mv, outerTyp, outerVal, path );
        }
        else if ( isBaseTypeOf( typ, TYPE_REF_MINGLE_EXCEPTION ) )
        {
            return coerceMingleException( mv, outerTyp, outerVal, path );
        }
        else if ( isBaseTypeOf( typ, TYPE_REF_MINGLE_ENUM ) )
        {
            return coerceMingleEnum( mv, outerTyp, outerVal, path );
        }
        else throw castException( outerTyp, outerVal, path );
    }

    private
    static
    MingleValue
    asAtomicMingleInstance( AtomicTypeReference typ,
                            MingleValue mv,
                            MingleTypeReference outerTyp,
                            MingleValue outerVal,
                            ObjectPath< MingleIdentifier > path )
    {
        MingleValue res = 
            getAtomicValueBase( typ, mv, outerTyp, outerVal, path );
        
        MingleValueRestriction restriction = typ.getRestriction();
        
        if ( restriction == null ) return res;
        else
        {
            restriction.validate( res, path );
            return res;
        }
    }

    private
    static
    MingleList
    buildListInstance( ListTypeReference typ,
                       Iterator< MingleValue > it,
                       ObjectPath< MingleIdentifier > path )
    {
        MingleTypeReference eltTyp = typ.getElementTypeReference();
        ImmutableListPath< MingleIdentifier > lp = path.startImmutableList();

        MingleList.Builder b = new MingleList.Builder();

        while ( it.hasNext() )
        {
            b.add( asMingleInstance( eltTyp, it.next(), lp ) );
            lp = lp.next();
        }

        return b.build();
    }

    private
    static
    MingleList
    asMingleListInstance( ListTypeReference typ,
                          MingleValue mv,
                          boolean shallow,
                          MingleTypeReference outerTyp,
                          MingleValue outerVal,
                          ObjectPath< MingleIdentifier > path )
    {
        if ( mv instanceof MingleList )
        {
            MingleList ml = (MingleList) mv;
            Iterator< MingleValue > it = ml.iterator();
            
            if ( it.hasNext() ) 
            {
                return shallow ? ml : buildListInstance( typ, it, path );
            }
            else
            {
                if ( typ.allowsEmpty() ) return ml;
                else
                {
                    throw 
                        new MingleValidationException( "list is empty", path );
                }
            }
        }
        else throw castException( outerTyp, outerVal, path );
    }

    public
    static
    MingleList
    asMingleListInstance( ListTypeReference lt,
                          MingleValue mv,
                          boolean shallow,
                          ObjectPath< MingleIdentifier > path )
    {
        inputs.notNull( lt, "lt" );
        inputs.notNull( mv, "mv" );
        inputs.notNull( path, "path" );

        return asMingleListInstance( lt, mv, shallow, lt, mv, path );
    }

    private
    static
    MingleValue
    asNullableMingleInstance( NullableTypeReference typ,
                              MingleValue mv,
                              MingleTypeReference outerTyp,
                              MingleValue outerVal,
                              ObjectPath< MingleIdentifier > path )
    {
        if ( mv instanceof MingleNull ) return mv;
        else 
        {
            return 
                asMingleInstance( 
                    typ.getTypeReference(), 
                    mv, 
                    outerTyp,
                    outerVal, 
                    path 
                );
        }
    }

    private
    static
    MingleValue
    asMingleInstance( MingleTypeReference typ,
                      MingleValue mv,
                      MingleTypeReference outerTyp,
                      MingleValue outerVal,
                      ObjectPath< MingleIdentifier > path )
    {
        if ( typ instanceof AtomicTypeReference )
        {
            AtomicTypeReference at = (AtomicTypeReference) typ;
            return asAtomicMingleInstance( at, mv, outerTyp, outerVal, path );
        }
        else if ( typ instanceof ListTypeReference )
        {
            ListTypeReference lt = (ListTypeReference) typ;

            return 
                asMingleListInstance( lt, mv, false, outerTyp, outerVal, path );
        }
        else if ( typ instanceof NullableTypeReference )
        {
            NullableTypeReference nt = (NullableTypeReference) typ;
            return asNullableMingleInstance( nt, mv, outerTyp, outerVal, path );
        }
        else throw state.createFail( "Unhandled type:", typ );
    }

    public
    static
    MingleValue
    asMingleInstance( MingleTypeReference typ,
                      MingleValue mv,
                      ObjectPath< MingleIdentifier > path )
    {
        inputs.notNull( typ, "typ" );
        inputs.notNull( mv, "mv" );
        inputs.notNull( path, "path" );

        return asMingleInstance( typ, mv, typ, mv, path );
    }

    private
    static
    QualifiedTypeName
    expectQname( AtomicTypeReference atr )
    {
        AtomicTypeReference.Name nm = atr.getName();

        if ( nm instanceof QualifiedTypeName ) return (QualifiedTypeName) nm;
        else throw state.createFail( "Expected qname but got:", nm );
    }

    private
    static
    NullableTypeReference
    nullableSuperTypeOf( NullableTypeReference ntr,
                         TypeDefinitionLookup types )
    {
        MingleTypeReference spr = superTypeOf( ntr.getTypeReference(), types );
        return spr == null ? null : NullableTypeReference.create( spr );
    }

    private
    static
    ListTypeReference
    listSuperTypeOf( ListTypeReference ltr,
                     TypeDefinitionLookup types )
    {
        MingleTypeReference spr =
            superTypeOf( ltr.getElementTypeReference(), types );
        
        if ( spr == null ) return null;
        else return ListTypeReference.create( spr, ltr.allowsEmpty() );
    }

    private
    static
    AtomicTypeReference
    atomicSuperTypeOf( AtomicTypeReference atr,
                       TypeDefinitionLookup types )
    {
        QualifiedTypeName qn = expectQname( atr );

        TypeDefinition td = types.expectType( qn );

        MingleTypeReference res = td.getSuperType();

        if ( res == null ) return null;
        else
        {
            state.isTrue( 
                res instanceof AtomicTypeReference,
                "Super type of atomic type", atr, "is not atomic:", res );
            
            return (AtomicTypeReference) res;
        }
    }

    private
    static
    MingleTypeReference
    superTypeOf( MingleTypeReference typ,
                 TypeDefinitionLookup types )
    {
        if ( typ instanceof NullableTypeReference )
        {
            return nullableSuperTypeOf( (NullableTypeReference) typ, types );
        }
        else if ( typ instanceof ListTypeReference )
        {
            return listSuperTypeOf( (ListTypeReference) typ, types );
        }
        else if ( typ instanceof AtomicTypeReference )
        {
            return atomicSuperTypeOf( (AtomicTypeReference) typ, types );
        }
        else throw state.createFail( "Unhandled type:", typ );
    }
    
    private
    static
    boolean
    isAssignableNullable( NullableTypeReference lhs,
                          MingleTypeReference rhs,
                          TypeDefinitionLookup types )
    {
        if ( rhs instanceof NullableTypeReference )
        {
            rhs = ( (NullableTypeReference) rhs ).getTypeReference();
        }

        return isAssignable( lhs.getTypeReference(), rhs, types );
    }

    private
    static
    boolean
    isAssignableList( ListTypeReference lhs,
                      MingleTypeReference rhs,
                      TypeDefinitionLookup types )
    {
        if ( rhs instanceof ListTypeReference )
        {
            ListTypeReference rhLtr = (ListTypeReference) rhs;

            if ( lhs.allowsEmpty() || ( ! rhLtr.allowsEmpty() ) )
            {
                return 
                    isAssignable(
                        lhs.getElementTypeReference(),
                        rhLtr.getElementTypeReference(),
                        types
                    );
            }
            else return false;
        }
        else return false;
    }
 
    private
    static
    boolean
    isAssignableAtomic( AtomicTypeReference lhs,
                        MingleTypeReference rhs,
                        TypeDefinitionLookup types )
    {
        if ( rhs instanceof AtomicTypeReference )
        {
            if ( lhs.equals( rhs ) || lhs.equals( TYPE_REF_MINGLE_VALUE ) ) 
            {
                return true;
            }
            else
            {
                MingleTypeReference spr = superTypeOf( rhs, types );
                
                if ( spr == null ) return false;
                else return isAssignable( lhs, spr, types );
            }
        }
        else if ( rhs instanceof ListTypeReference )
        {
            return lhs.equals( TYPE_REF_MINGLE_VALUE );
        }
        else return false;
    }

    public
    static
    boolean
    isAssignable( MingleTypeReference lhs,
                  MingleTypeReference rhs,
                  TypeDefinitionLookup types )
    {
        inputs.notNull( lhs, "lhs" );
        inputs.notNull( rhs, "rhs" );
        inputs.notNull( types, "types" );

        if ( lhs instanceof NullableTypeReference )
        {
            return 
                isAssignableNullable( (NullableTypeReference) lhs, rhs, types );
        }
        else if ( lhs instanceof ListTypeReference )
        {
            return isAssignableList( (ListTypeReference) lhs, rhs, types );
        }
        else if ( lhs instanceof AtomicTypeReference )
        {
            return isAssignableAtomic( (AtomicTypeReference) lhs, rhs, types );
        }
        else return false;
    }

    public
    static
    CharSequence
    asString( ObjectPath< MingleIdentifier > path )
    {
        inputs.notNull( path, "path" );

        return ObjectPaths.format( path, PATH_FORMATTER );
    }

    public
    static
    ObjectPathFormatter< MingleIdentifier >
    getIdentifierPathFormatter()
    {
        return PATH_FORMATTER;
    }

    public
    static
    CharSequence
    format( ObjectPath< MingleIdentifier > path )
    {
        inputs.notNull( path, "path" );
        return ObjectPaths.format( path, PATH_FORMATTER );
    }

    private
    static
    CharSequence
    formatLcCamelCapped( MingleIdentifier id )
    {
        String[] parts = id.getPartsArray();

        StringBuilder res = new StringBuilder();

        res.append( parts[ 0 ] );

        for ( int i = 1, e = parts.length; i < e; ++i )
        {
            res.append( Character.toUpperCase( parts[ i ].charAt( 0 ) ) );
            res.append( parts[ i ].substring( 1 ) );
        }

        return res;
    }

    public
    static
    CharSequence
    format( MingleIdentifier id,
            MingleIdentifierFormat fmt )
    {
        inputs.notNull( id, "id" );
        inputs.notNull( fmt, "fmt" );

        switch ( fmt )
        {
            case LC_HYPHENATED: return Strings.join( "-", id.getPartsArray() );
            case LC_UNDERSCORE: return Strings.join( "_", id.getPartsArray() );
            case LC_CAMEL_CAPPED: return formatLcCamelCapped( id );

            default: throw state.createFail( "Unrecognized format:", fmt );
        }
    }

    public static MingleList getEmptyList() { return EMPTY_LIST; }

    public 
    static 
    MingleSymbolMap 
    getEmptySymbolMap() 
    { 
        return EMPTY_SYMBOL_MAP;
    }

    public
    static
    MingleTypeReference
    getType( MingleValue mv )
    {
        inputs.notNull( mv, "mv" );

        if ( mv instanceof MingleStructure )
        {
            return ( (MingleStructure) mv ).getType();
        }
        else if ( mv instanceof MingleEnum )
        {
            return ( (MingleEnum) mv ).getType();
        }
        else 
        {
            throw state.createFail( 
                "Type refs not yet supported for instances of", mv.getClass() );
        }
    }

    public
    static
    AtomicTypeReference.Name
    typeNameIn( MingleTypeReference ref )
    {
        inputs.notNull( ref, "ref" );

        if ( ref instanceof AtomicTypeReference )
        {
            return ( (AtomicTypeReference) ref ).getName();
        }
        else if ( ref instanceof ListTypeReference )
        {
            return 
                typeNameIn( 
                    ( (ListTypeReference) ref ).getElementTypeReference() );
        }
        else if ( ref instanceof NullableTypeReference )
        {
            return
                typeNameIn(
                    ( (NullableTypeReference) ref ).getTypeReference() );
        }
        else throw state.createFail( "Unexpected type ref:", ref );
    }

    // A simple map lookup won't suffice here since we're dealing with
    // interfaces, some of which are more specific than others (MingleStructure
    // < MingleException), so we do a multiway if/else testing for assignment
    // compatibility where order matters
    public
    static
    MingleTypeReference
    typeReferenceOf( Class< ? extends MingleValue > cls )
    {
        inputs.notNull( cls, "cls" );

        if ( MingleString.class.isAssignableFrom( cls ) )
        {
            return TYPE_REF_MINGLE_STRING;
        }
        else if ( MingleInt32.class.isAssignableFrom( cls ) )
        {
            return TYPE_REF_MINGLE_INT32;
        }
        else if ( MingleInt64.class.isAssignableFrom( cls ) )
        {
            return TYPE_REF_MINGLE_INT64;
        }
        else if ( MingleDouble.class.isAssignableFrom( cls ) )
        {
            return TYPE_REF_MINGLE_DOUBLE;
        }
        else if ( MingleFloat.class.isAssignableFrom( cls ) )
        {
            return TYPE_REF_MINGLE_FLOAT;
        }
        else if ( MingleBoolean.class.isAssignableFrom( cls ) )
        {
            return TYPE_REF_MINGLE_BOOLEAN;
        }
        else if ( MingleBuffer.class.isAssignableFrom( cls ) )
        {
            return TYPE_REF_MINGLE_BUFFER;
        }
        else if ( MingleTimestamp.class.isAssignableFrom( cls ) )
        {
            return TYPE_REF_MINGLE_TIMESTAMP;
        }
        else if ( MingleList.class.isAssignableFrom( cls ) )
        {
            return ListTypeReference.create( TYPE_REF_MINGLE_VALUE, true );
        }
        else if ( MingleSymbolMap.class.isAssignableFrom( cls ) )
        {
            return TYPE_REF_MINGLE_SYMBOL_MAP;
        }
        else if ( MingleStruct.class.isAssignableFrom( cls ) )
        {
            return TYPE_REF_MINGLE_STRUCT;
        }
        else if ( MingleException.class.isAssignableFrom( cls ) )
        {
            return TYPE_REF_MINGLE_EXCEPTION;
        }
        // This needs to come below the checks for MingleStruct and
        // MingleException
        else if ( MingleStructure.class.isAssignableFrom( cls ) )
        {
            return TYPE_REF_MINGLE_STRUCTURE;
        }
        else if ( MingleNull.class.isAssignableFrom( cls ) )
        {
            return TYPE_REF_MINGLE_NULL;
        }
        else if ( MingleEnum.class.isAssignableFrom( cls ) )
        {
            return TYPE_REF_MINGLE_ENUM;
        }
        else if ( MingleValue.class.isAssignableFrom( cls ) )
        {
            return TYPE_REF_MINGLE_VALUE;
        }
        else
        {
            throw inputs.createFail(
                "Don't know how to get mingle type name for", cls );
        }
    }

    public
    static
    MingleTypeReference
    typeReferenceOf( MingleValue mv )
    {
        inputs.notNull( mv, "mv" );

        if ( mv instanceof MingleEnum ) return ( (MingleEnum) mv ).getType();
        else if ( mv instanceof MingleStructure )
        {
            return ( (MingleStructure) mv ).getType();
        }
        else return typeReferenceOf( mv.getClass() );
    }

    public
    static
    MingleSymbolMap
    asStructureFields( MingleValue mv,
                       MingleTypeReference expct,
                       ObjectPath< MingleIdentifier > path )
    {
        inputs.notNull( mv, "mv" );
        inputs.notNull( expct, "expct" );
        inputs.notNull( path, "path" );

        MingleSymbolMap res = null;

        if ( mv instanceof MingleSymbolMap ) res = (MingleSymbolMap) mv;
        else if ( mv instanceof MingleStructure )
        {
            MingleStructure ms = (MingleStructure) mv;
            if ( expct.equals( ms.getType() ) ) res = ms.getFields();
        }

        if ( res == null )
        {
            throw 
                new MingleTypeCastException( 
                    expct, typeReferenceOf( mv ), path );
        }
        else return res;
    }

    public
    static
    MingleSymbolMapAccessor
    expectStruct( MingleValue mv,
                  ObjectPath< MingleIdentifier > path,
                  MingleTypeReference expct )
    {
        inputs.notNull( mv, "mv" );
        inputs.notNull( path, "path" );
        inputs.notNull( expct, "expct" );

        MingleStruct ms = 
            (MingleStruct) asMingleInstance( TYPE_REF_MINGLE_STRUCT, mv, path );

        MingleValidation.isTrue(
            ms.getType().equals( expct ),
            path,
            "Unexpected type:", ms.getType()
        );

        return MingleSymbolMapAccessor.create( ms, path );
    }

    // Only gets the raw value, supplying a raw default as appropriate. Does not
    // make attempts to coerce mingle value or to validate it against
    // restriction types; callers should do that on their own
    public
    static
    MingleValue
    get( MingleSymbolMap m,
         FieldDefinition fd )
    {
        inputs.notNull( m, "m" );
        inputs.notNull( fd, "fd" );

        MingleValue res = m.get( fd.getName() );
        
        if ( res == null || res instanceof MingleNull ) res = fd.getDefault();

        MingleTypeReference typ = fd.getType();

        if ( res == null || res instanceof MingleNull )
        {
            if ( typ instanceof ListTypeReference &&
                 ( (ListTypeReference) typ ).allowsEmpty() ) 
            {
                res = getEmptyList();
            }
        }
        
        return res;
    }

    // does input checking for public frontend methods
    private
    static
    MingleTypeReference
    accessTypeReference( MingleSymbolMapAccessor acc,
                         MingleIdentifier id,
                         boolean useExpect )
    {
        inputs.notNull( acc, "acc" );
        inputs.notNull( id, "id" );

        MingleString str =
            useExpect 
                ? acc.expectMingleString( id ) : acc.getMingleString( id );
        
        if ( str == null ) return null; // only if useExpect == false
        else
        {
            try { return MingleTypeReference.parse( str ); }
            catch ( SyntaxException se )
            {
                throw new MingleValidationException(
                    se.getMessage(), acc.getPath().descend( id ) );
            }
        }
    }

    public
    static
    MingleTypeReference
    getTypeReference( MingleSymbolMapAccessor acc,
                      MingleIdentifier id )
    {
        return accessTypeReference( acc, id, false );
    }

    public
    static
    MingleTypeReference
    expectTypeReference( MingleSymbolMapAccessor acc,
                         MingleIdentifier id )
    {
        return accessTypeReference( acc, id, true );
    }

    public
    static
    Class< ? extends MingleValue >
    javaClassFor( MingleTypeName nm )
    {
        inputs.notNull( nm, "nm" );

        Class< ? extends MingleValue > res = JAVA_TYPE_NAMES.get( nm );

        if ( res == null )
        {
            throw inputs.createFail( "Unrecognized mingle type:", nm );
        }
        else return res;
    }

    public
    static
    String
    asJavaEnumName( MingleIdentifier name )
    {
        inputs.notNull( name, "name" );
        
        return 
            Strings.join( "_", name.getPartsArray() ).toString().toUpperCase();
    }

    public
    static
    < E extends Enum< E > >
    E
    valueOf( Class< E > enCls,
             MingleIdentifier constant )
    {
        inputs.notNull( enCls, "enCls" );
        inputs.notNull( constant, "constant" );

        return Enum.valueOf( enCls, asJavaEnumName( constant ) );
    }

    public
    static
    StringBuilder
    appendFilePathFor( MingleIdentifiedName nm,
                       StringBuilder sb )
    {
        inputs.notNull( nm, "nm" );
        inputs.notNull( sb, "sb" );

        MingleNamespace ns = nm.getNamespace();

        Iterator< MingleIdentifier > it = ns.getParts().iterator();

        while ( it.hasNext() )
        {
            sb.append( it.next().getExternalForm() );
            if ( it.hasNext() ) sb.append( '/' );
        }

        sb.append( '.' ).append( ns.getVersion().getExternalForm() );

        for ( MingleIdentifier id : nm.getNames() )
        {
            sb.append( '/' ).append( id.getExternalForm() );
        }

        return sb;
    }

    public
    static
    CharSequence
    filePathFor( MingleIdentifiedName nm )
    {
        return appendFilePathFor( nm, new StringBuilder() );
    }

    public
    static
    boolean
    wasSerialized( Throwable th )
    {
        return
            ( th instanceof MingleValidationException ) &&
            ( (MingleValidationException) th ).wasSerialized();
    }

    private
    static
    abstract
    class ValidationExceptionExchanger< E >
    extends AbstractExceptionExchanger< E >
    {
        final static MingleIdentifier ID_LOCATION =
            MingleIdentifier.create( "location" );

        private 
        ValidationExceptionExchanger( AtomicTypeReference typ,
                                      Class< E > exCls ) 
        { 
            super( typ, exCls );
        }

        final
        ObjectPath< MingleIdentifier >
        expectLocation( MingleSymbolMapAccessor acc )
        {
            String locStr = acc.expectString( ID_LOCATION );

            try { return MingleParsers.parseObjectPath( locStr ); }
            catch ( SyntaxException se )
            { 
                throw createRethrow( se, acc, ID_LOCATION ); 
            }
        }
    }

    private
    final
    static
    class MingleValidationExceptionExchanger
    extends ValidationExceptionExchanger< MingleValidationException >
    {
        private
        MingleValidationExceptionExchanger()
        {
            super( 
                TYPE_REF_VALIDATION_EXCEPTION,
                MingleValidationException.class 
            );
        }

        protected
        MingleValidationException
        buildException( MingleSymbolMapAccessor acc )
        {
            return
                new MingleValidationException(
                    expectMessage( acc ),
                    expectLocation( acc ),
                    true,
                    null
                );
        }

        protected
        MingleValue
        implAsMingleValue( MingleValidationException me,
                           ObjectPath< String > path )
        {
            return
                exceptionBuilder().
                    f().setString( ID_MESSAGE, me.getDescription() ).
                    f().setString( ID_LOCATION, format( me.getLocation() ) ).
                    build();
        }
    }

    private
    final
    static
    class MingleTypeCastExceptionExchanger
    extends ValidationExceptionExchanger< MingleTypeCastException >
    {
        private final static MingleIdentifier ID_EXPCT =
            MingleIdentifier.create( "expected-type" );

        private final static MingleIdentifier ID_ACTUAL =
            MingleIdentifier.create( "actual-type" );

        private
        MingleTypeCastExceptionExchanger()
        {
            super( 
                TYPE_REF_TYPE_CAST_EXCEPTION, MingleTypeCastException.class );
        }

        protected
        MingleTypeCastException
        buildException( MingleSymbolMapAccessor acc )
        {
            return
                new MingleTypeCastException(
                    expectTypeRef( acc, ID_EXPCT ),
                    expectTypeRef( acc, ID_ACTUAL ),
                    expectLocation( acc ),
                    true
                );
        }

        protected
        MingleValue
        implAsMingleValue( MingleTypeCastException ex,
                           ObjectPath< String > path )
        {
            return
                exceptionBuilder().
                    f().setString( 
                        ID_EXPCT, ex.getExpectedType().getExternalForm() ).
                    f().setString(
                        ID_ACTUAL, ex.getActualType().getExternalForm() ).
                    f().setString( ID_LOCATION, format( ex.getLocation() ) ).
                    build();
        }
    }

    public
    static
    MingleValueExchanger
    exchangerFor( MingleTypeReference typ )
    {
        inputs.notNull( typ, "typ" );

        MingleValueExchanger res = EXCHANGERS.get( typ );

        inputs.isFalse( res == null, "No exchanger known for", typ );
        return res;
    }

    public
    static
    < E extends Enum< E > >
    MingleValueExchanger
    createExchanger( AtomicTypeReference typ,
                     Class< E > cls )
    {
        inputs.notNull( typ, "typ" );
        inputs.notNull( cls, "cls" );

        return new EnumExchanger< E >( typ, cls );
    }

    public
    static
    Iterable< MingleValueExchanger >
    getExchangers()
    {
        return EXCHANGERS.values();
    }

    static
    {
        INTEGRAL_JAVA_TYPES =
            Lang.unmodifiableSet(
                new HashSet< Class< ? > >(
                    Arrays.< Class< ? > >asList(
                        Byte.class,
                        Byte.TYPE,
                        Short.class,
                        Short.TYPE,
                        Character.class,
                        Character.TYPE,
                        Integer.class,
                        Integer.TYPE,
                        Long.class,
                        Long.TYPE ) ) );
        
        DECIMAL_JAVA_TYPES =
            Lang.unmodifiableSet(
                new HashSet< Class< ? > >(
                    Arrays.< Class< ? > >asList(
                        Float.class,
                        Float.TYPE,
                        Double.class,
                        Double.TYPE ) ) );
                        
    }

    private
    final
    static
    class PathFormatterImpl
    implements ObjectPathFormatter< MingleIdentifier >
    {
        public void formatPathStart( StringBuilder sb ) {}

        public void formatSeparator( StringBuilder sb ) { sb.append( '.' ); }

        public
        void
        formatDictionaryKey( StringBuilder sb,
                             MingleIdentifier key )
        {
            sb.append( key.getExternalForm() );
        }

        public
        void
        formatListIndex( StringBuilder sb,
                         int indx )
        {
            sb.append( "[ " ).append( indx ).append( " ]" );
        }
    }

    static
    {
        Map< MingleTypeName, Class< ? extends MingleValue > > m =
            Lang.newMap();

        m.put(
            MingleParsers.createTypeName( "Boolean" ),
            MingleBoolean.class );

        m.put(
            MingleParsers.createTypeName( "Buffer" ),
            MingleBuffer.class );

        m.put(
            MingleParsers.createTypeName( "Exception" ),
            MingleException.class );

        m.put(
            MingleParsers.createTypeName( "List" ),
            MingleList.class );

        m.put(
            MingleParsers.createTypeName( "Null" ),
            MingleNull.class );

        m.put(
            MingleParsers.createTypeName( "String" ),
            MingleString.class );

        m.put(
            MingleParsers.createTypeName( "Struct" ),
            MingleStruct.class );

        m.put(
            MingleParsers.createTypeName( "SymbolMap" ),
            MingleSymbolMap.class );

        m.put(
            MingleParsers.createTypeName( "Timestamp" ),
            MingleTimestamp.class );

        
        JAVA_TYPE_NAMES = Lang.unmodifiableMap( m );
    }

    private
    static
    Map< MingleTypeReference, MingleValueExchanger >
    makeExchangers( MingleValueExchanger ... arr )
    {
        Map< MingleTypeReference, MingleValueExchanger > res = Lang.newMap();

        for ( MingleValueExchanger exch : arr ) 
        {
            res.put( exch.getMingleType(), exch );
        }

        return Lang.unmodifiableMap( res );
    }

    static
    {
        EXCHANGERS =
            makeExchangers(

                new MingleValidationExceptionExchanger(),
                new MingleTypeCastExceptionExchanger(),
                
                new AbstractParsedStringExchanger< MingleIdentifier >(
                    TYPE_REF_MINGLE_IDENTIFIER, MingleIdentifier.class ) 
                {
                    protected MingleIdentifier parse( CharSequence cs )
                        throws SyntaxException
                    {   
                        return MingleIdentifier.parse( cs );
                    }
 
                    protected CharSequence asString( MingleIdentifier obj ) {
                        return obj.getExternalForm();
                    }
                },
                
                new AbstractParsedStringExchanger< MingleNamespace >(
                    TYPE_REF_MINGLE_NAMESPACE, MingleNamespace.class ) 
                {
                    protected MingleNamespace parse( CharSequence cs )
                        throws SyntaxException
                    {   
                        return MingleNamespace.parse( cs );
                    }
 
                    protected CharSequence asString( MingleNamespace obj ) {
                        return obj.getExternalForm();
                    }
                },
                
                new AbstractParsedStringExchanger< QualifiedTypeName >(
                    TYPE_REF_QUALIFIED_TYPE_NAME, QualifiedTypeName.class ) 
                {
                    protected QualifiedTypeName parse( CharSequence cs )
                        throws SyntaxException
                    {   
                        return QualifiedTypeName.parse( cs );
                    }
 
                    protected CharSequence asString( QualifiedTypeName obj ) {
                        return obj.getExternalForm();
                    }
                },
                
                new AbstractParsedStringExchanger< MingleTypeReference >(
                    TYPE_REF_MINGLE_TYPE_REFERENCE, MingleTypeReference.class ) 
                {
                    protected MingleTypeReference parse( CharSequence cs )
                        throws SyntaxException
                    {   
                        return MingleTypeReference.parse( cs );
                    }
 
                    protected CharSequence asString( MingleTypeReference obj ) {
                        return obj.getExternalForm();
                    }
                },
                
                new AbstractParsedStringExchanger< MingleIdentifiedName >(
                    TYPE_REF_MINGLE_IDENTIFIED_NAME, 
                    MingleIdentifiedName.class ) 
                {
                    protected MingleIdentifiedName parse( CharSequence cs )
                        throws SyntaxException
                    {   
                        return MingleIdentifiedName.parse( cs );
                    }
 
                    protected CharSequence asString( MingleIdentifiedName obj ) 
                    {
                        return obj.getExternalForm();
                    }
                }
            );
    }
}
