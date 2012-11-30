package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.Map;

public
final
class Mingle
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final static MingleNameResolver CORE_NAME_RESOLVER =
        new MingleNameResolver() {
            public QualifiedTypeName resolve( DeclaredTypeName nm ) {
                return Mingle.resolveInCore( nm );
            }
        };
    
    public final static MingleNamespace NS_CORE;
    public final static QualifiedTypeName QNAME_BOOLEAN;
    public final static MingleTypeReference TYPE_BOOLEAN;
    public final static QualifiedTypeName QNAME_INT32;
    public final static MingleTypeReference TYPE_INT32;
    public final static QualifiedTypeName QNAME_INT64;
    public final static MingleTypeReference TYPE_INT64;
    public final static QualifiedTypeName QNAME_UINT32;
    public final static MingleTypeReference TYPE_UINT32;
    public final static QualifiedTypeName QNAME_UINT64;
    public final static MingleTypeReference TYPE_UINT64;
    public final static QualifiedTypeName QNAME_FLOAT32;
    public final static MingleTypeReference TYPE_FLOAT32;
    public final static QualifiedTypeName QNAME_FLOAT64;
    public final static MingleTypeReference TYPE_FLOAT64;
    public final static QualifiedTypeName QNAME_STRING;
    public final static MingleTypeReference TYPE_STRING;
    public final static QualifiedTypeName QNAME_BUFFER;
    public final static MingleTypeReference TYPE_BUFFER;
    public final static QualifiedTypeName QNAME_TIMESTAMP;
    public final static MingleTypeReference TYPE_TIMESTAMP;
    public final static QualifiedTypeName QNAME_NULL;
    public final static MingleTypeReference TYPE_NULL;
    public final static QualifiedTypeName QNAME_VALUE;
    public final static MingleTypeReference TYPE_VALUE;

    private final static Map< DeclaredTypeName, QualifiedTypeName > 
        CORE_DECL_NAMES;

    private final static 
        Map< MingleTypeReference, Class< ? extends MingleValue > > 
            VALUE_CLASSES;

    private Mingle() {}
    
    public
    static
    AtomicTypeReference
    atomicTypeReferenceIn( MingleTypeReference ref )
    {
        inputs.notNull( ref, "ref" );

        if ( ref instanceof AtomicTypeReference )
        {
            return (AtomicTypeReference) ref;
        }
        else if ( ref instanceof ListTypeReference )
        {
            return 
                atomicTypeReferenceIn( 
                    ( (ListTypeReference) ref ).getElementTypeReference() );
        }
        else if ( ref instanceof NullableTypeReference )
        {
            return
                atomicTypeReferenceIn(
                    ( (NullableTypeReference) ref ).getTypeReference() );
        }
        else throw state.createFail( "Unexpected type ref:", ref );
    }

    static
    QualifiedTypeName
    resolveInCore( DeclaredTypeName nm )
    {
        inputs.notNull( nm, "nm" );
        return CORE_DECL_NAMES.get( nm );
    }

    public
    static
    boolean
    isIntegralType( MingleTypeReference t )
    {
        inputs.notNull( t, "t" );

        return t.equals( TYPE_INT32 ) ||
               t.equals( TYPE_UINT32 ) ||
               t.equals( TYPE_INT64 ) ||
               t.equals( TYPE_UINT64 );
    }

    public
    static
    boolean
    isDecimalType( MingleTypeReference t )
    {
        inputs.notNull( t, "t" );
        return t.equals( TYPE_FLOAT32 ) || t.equals( TYPE_FLOAT64 );
    }

    public
    static
    boolean
    isNumberType( MingleTypeReference t )
    {
        return isIntegralType( t ) || isDecimalType( t );
    }

    public
    static
    Class< ? extends MingleValue >
    valueClassFor( MingleTypeReference typ )
    {
        inputs.notNull( typ, "typ" );

        return VALUE_CLASSES.get( typ );
    }

    public
    static
    Class< ? extends MingleValue >
    expectValueClassFor( MingleTypeReference typ )
    {
        Class< ? extends MingleValue > res = valueClassFor( typ );

        if ( res == null )
        {
            throw inputs.createFail( "No value class known for", typ );
        }

        return res;
    }
    
    private
    static
    StringBuilder
    implInspect( StringBuilder sb,
                 MingleValue mv )
    {
        if ( mv instanceof MingleNull ) sb.append( "null" );
        else if ( mv instanceof MingleBoolean ||
                  mv instanceof MingleUint32 ||
                  mv instanceof MingleUint64 ||
                  mv instanceof MingleInt32 ||
                  mv instanceof MingleInt64 ||
                  mv instanceof MingleFloat32 ||
                  mv instanceof MingleFloat64 ) 
        {
            sb.append( mv.toString() );
        }
        else if ( mv instanceof MingleString )
        {
            Lang.appendRfc4627String( sb, (MingleString) mv );
        }
        else if ( mv instanceof MingleTimestamp )
        {
            sb.append( ( (MingleTimestamp) mv ).getRfc3339String() );
        }
        else if ( mv instanceof MingleBuffer )
        {
            sb.append( "buffer:[" );
            sb.append( ( (MingleBuffer) mv ).asHexString() );
            sb.append( "]" );
        }
        else state.fail( "Unhandled inspect type:", mv.getClass().getName() );
        
        return sb;
    }

    public
    static
    StringBuilder
    appendInspection( StringBuilder sb,
                      MingleValue mv )
    {
        inputs.notNull( sb, "sb" );
        inputs.notNull( mv, "mv" );

        return implInspect( sb, mv );
    }

    public
    static
    CharSequence
    inspect( MingleValue mv )
    {
        inputs.notNull( mv, "mv" );
        return implInspect( new StringBuilder(), mv );
    }

    // Hardcoding impl right now that mv is string since we're only trying to
    // use this to get past type ref parse tests. Once those are done this will
    // be replaced with a more broadly applicable version.
    public
    static
    MingleValue
    castValue( MingleValue mv,
               MingleTypeReference typ )
    {
        inputs.notNull( mv, "mv" );
        inputs.notNull( typ, "typ" );

        String s = ( (MingleString) mv ).toString();

        if ( typ.equals( TYPE_STRING ) ) return mv;
        else if ( typ.equals( TYPE_INT32 ) )
        {
            return new MingleInt32( Integer.parseInt( s ) );
        }
        else if ( typ.equals( TYPE_UINT32 ) )
        {
            return new MingleUint32( (int) Long.parseLong( s ) );
        }
        else if ( typ.equals( TYPE_INT64 ) )
        {
            return new MingleInt64( Long.parseLong( s ) );
        }
        else if ( typ.equals( TYPE_UINT64 ) )
        {
            return new MingleUint64( Lang.parseUint64( s ) );
        }
        else if ( typ.equals( TYPE_FLOAT32 ) )
        {
            return new MingleFloat32( Float.parseFloat( s ) );
        }
        else if ( typ.equals( TYPE_FLOAT64 ) )
        {
            return new MingleFloat64( Double.parseDouble( s ) );
        }
        else if ( typ.equals( TYPE_TIMESTAMP ) )
        {
            try { return MingleTimestamp.parse( s ); }
            catch ( MingleSyntaxException mse )
            {
                throw new MingleValidationException( mse.getError() );
            }
        }
        else throw state.createFail( "Unhandled type:", typ );
    }

    // Static class init follows

    private
    static
    MingleIdentifier
    initId( String... parts )
    {
        return new MingleIdentifier( parts );
    }

    private
    static
    MingleNamespace
    initNs( MingleIdentifier ver,
            MingleIdentifier... parts )
    {
        return new MingleNamespace( parts, ver );
    }

    private
    static
    QualifiedTypeName
    initCoreQname( String nm )
    {
        state.notNull( NS_CORE, "NS_CORE" );
        state.notNull( CORE_DECL_NAMES, "CORE_DECL_NAMES" );

        DeclaredTypeName dn = new DeclaredTypeName( nm );
        QualifiedTypeName res = new QualifiedTypeName( NS_CORE, dn );

        Lang.putUnique( CORE_DECL_NAMES, dn, res );

        return res;
    }

    private
    static
    AtomicTypeReference
    initCoreType( QualifiedTypeName qn )
    {
        state.notNull( VALUE_CLASSES, "VALUE_CLASSES" );
        AtomicTypeReference res = new AtomicTypeReference( qn, null );

        try
        {
            String nm = "com.bitgirder.mingle.Mingle" + qn.getName();
            Class< ? > c1 = Class.forName( nm );

            VALUE_CLASSES.put( res, c1.asSubclass( MingleValue.class ) );
        }
        catch ( Exception ex )
        {
            String msg = "Couldn't initialize core type class for " + qn;
            throw new RuntimeException( msg );
        }

        return res;
    }

    static
    {
        NS_CORE = 
            initNs( initId( "v1" ), initId( "mingle" ), initId( "core" ) );

        CORE_DECL_NAMES = Lang.newMap();
        VALUE_CLASSES = Lang.newMap();
        
        QNAME_BOOLEAN = initCoreQname( "Boolean" );
        TYPE_BOOLEAN = initCoreType( QNAME_BOOLEAN );
        QNAME_INT32 = initCoreQname( "Int32" );
        TYPE_INT32 = initCoreType( QNAME_INT32 );
        QNAME_INT64 = initCoreQname( "Int64" );
        TYPE_INT64 = initCoreType( QNAME_INT64 );
        QNAME_UINT32 = initCoreQname( "Uint32" );
        TYPE_UINT32 = initCoreType( QNAME_UINT32 );
        QNAME_UINT64 = initCoreQname( "Uint64" );
        TYPE_UINT64 = initCoreType( QNAME_UINT64 );
        QNAME_FLOAT32 = initCoreQname( "Float32" );
        TYPE_FLOAT32 = initCoreType( QNAME_FLOAT32 );
        QNAME_FLOAT64 = initCoreQname( "Float64" );
        TYPE_FLOAT64 = initCoreType( QNAME_FLOAT64 );
        QNAME_STRING = initCoreQname( "String" );
        TYPE_STRING = initCoreType( QNAME_STRING );
        QNAME_BUFFER = initCoreQname( "Buffer" );
        TYPE_BUFFER = initCoreType( QNAME_BUFFER );
        QNAME_TIMESTAMP = initCoreQname( "Timestamp" );
        TYPE_TIMESTAMP = initCoreType( QNAME_TIMESTAMP );
        QNAME_NULL = initCoreQname( "Null" );
        TYPE_NULL = initCoreType( QNAME_NULL );
        QNAME_VALUE = initCoreQname( "Value" );
        TYPE_VALUE = initCoreType( QNAME_VALUE );
    }
}
