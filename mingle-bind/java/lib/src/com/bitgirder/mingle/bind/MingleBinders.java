package com.bitgirder.mingle.bind;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.ObjectReceiver;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPaths;
import com.bitgirder.lang.path.ImmutableListPath;

import com.bitgirder.lang.reflect.ReflectUtils;

import com.bitgirder.io.IoUtils;
import com.bitgirder.io.FileWrapper;
import com.bitgirder.io.IoProcessor;
import com.bitgirder.io.Base64Encoder;

import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.FieldDefinition;
import com.bitgirder.mingle.model.ListTypeReference;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleEnum;
import com.bitgirder.mingle.model.MingleList;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleNull;
import com.bitgirder.mingle.model.MingleRangeRestriction;
import com.bitgirder.mingle.model.MingleRegexRestriction;
import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleStructure;
import com.bitgirder.mingle.model.MingleSymbolMap;
import com.bitgirder.mingle.model.MingleSymbolMapBuilder;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.MingleValidationException;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleValueExchanger;
import com.bitgirder.mingle.model.MingleValueRestriction;
import com.bitgirder.mingle.model.NullableTypeReference;
import com.bitgirder.mingle.model.MingleTypeCastException;
import com.bitgirder.mingle.model.PrimitiveDefinition;
import com.bitgirder.mingle.model.QualifiedTypeName;
import com.bitgirder.mingle.model.TypeDefinitions;
import com.bitgirder.mingle.model.TypeDefinitionLookup;
import com.bitgirder.mingle.model.NoSuchTypeDefinitionException;

import com.bitgirder.mingle.service.MingleServices;

import com.bitgirder.mingle.codec.MingleCodec;
import com.bitgirder.mingle.codec.MingleCodecs;

import com.bitgirder.mingle.bincodec.MingleBinaryCodecs;

import com.bitgirder.process.ProcessActivity;

import java.util.List;
import java.util.Map;
import java.util.Iterator;

import java.net.URL;

import java.nio.ByteBuffer;

public
final
class MingleBinders
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static MingleCodec codec = MingleBinaryCodecs.getCodec();

    private final static Base64Encoder enc = new Base64Encoder();

    final static Iterable< MingleTypeReference > STANDARD_BINDINGS;
    private final static List< MingleValueExchanger > STANDARD_EXCHANGERS;

    final static Iterable< MingleTypeReference > SVC_EXCEPTION_BINDINGS;

    final static Iterable< MingleTypeReference > AUTH_EXCEPTION_BINDINGS;
    private final static List< MingleValueExchanger > AUTH_EXCEPTION_EXCHANGERS;

    private final static MingleTypeReference MG_TYP_VAL =
        AtomicTypeReference.create( PrimitiveDefinition.QNAME_VALUE );

    public final static String RSRC_NAME =
        "mingle-bind-init.txt";

    private MingleBinders() {}

    public
    static
    interface Initializer
    {
        public
        void
        initialize( MingleBinder.Builder b,
                    TypeDefinitionLookup types )
            throws Exception;
    }

    private
    static
    RuntimeException
    createInitFail( int line,
                    URL rsrc,
                    Throwable cause,
                    String msg )
    {
        msg = 
            "[Related to line " + line + " in " + rsrc + "]: " +
            msg +
            " (See attached cause)";

        return new RuntimeException( msg, cause );
    }

    private
    static
    Class< ? extends Initializer >
    asInitializerClass( String clsNm,
                        int line,
                        URL rsrc )
    {
        Class< ? > cls;

        try { cls = Class.forName( clsNm ); }
        catch ( Throwable th )
        {
            throw 
                createInitFail( 
                    line, rsrc, th, "Couldn't get class: " + clsNm );
        }

        try { return cls.asSubclass( Initializer.class ); }
        catch ( Throwable th )
        {
            throw
                createInitFail( 
                    line, rsrc, th, 
                    "Can't interpret " + cls + " as an Initializer" );
        }
    }

    private
    static
    void
    loadBindings( URL rsrc,
                  MingleBinder.Builder b,
                  TypeDefinitionLookup types )
        throws Exception
    {
        int line = 0;
        for ( String clsNm : IoUtils.readLines( rsrc.openStream(), "utf-8" ) )
        {
            Class< ? extends Initializer > cls =
                asInitializerClass( clsNm, ++line, rsrc );

            Initializer iz = ReflectUtils.newInstance( cls );
            iz.initialize( b, types );
        }
    }

    public
    static
    MingleBinder
    loadDefault( TypeDefinitionLookup types )
        throws Exception
    {
        inputs.notNull( types, "types" );

        MingleBinder.Builder b = new MingleBinder.Builder();
        b.setTypes( types );

        for ( URL rsrc : IoUtils.getResources( RSRC_NAME ) )
        {
            loadBindings( rsrc, b, types );
        }

        return b.build();
    }

    // could eventually expand this to other bound objects or make public as
    // needed
    static
    boolean
    wasSerialized( Throwable th )
    {
        return
            ( th instanceof BoundException ) &&
            ( (BoundException) th ).wasSerialized();
    }

    private
    final
    static
    class ExchangerAdapter
    implements MingleBinding
    {
        private final MingleValueExchanger exch;

        private
        ExchangerAdapter( MingleValueExchanger exch )
        {
            this.exch = exch; 
        }

        public
        Object
        asJavaValue( AtomicTypeReference typ,
                     MingleValue mv,
                     MingleBinder mb,
                     ObjectPath< MingleIdentifier > path )
        {
            return exch.asJavaValue( mv, path );
        }

        public
        MingleValue
        asMingleValue( Object obj,
                       MingleBinder mb,
                       ObjectPath< String > path )
        {
            return exch.asMingleValue( obj, path );
        }
    }

    public
    static
    MingleBinding
    asMingleBinding( MingleValueExchanger exch )
    {
        return new ExchangerAdapter( inputs.notNull( exch, "exch" ) );
    }

    private
    static
    void
    addExchangerBindings( MingleBinder.Builder b,
                          List< MingleValueExchanger > l )
    {
        for ( MingleValueExchanger exch : l )
        {
            b.addBinding( 
                expectQname( (AtomicTypeReference) exch.getMingleType() ), 
                new ExchangerAdapter( exch ),
                exch.getJavaClass()
            );
        }
    }

    public
    static
    void
    addStandardBindings( MingleBinder.Builder b,
                         TypeDefinitionLookup types )
    {
        inputs.notNull( b, "b" );
        inputs.notNull( types, "types" );

        StandardException.addStandardBinding( b, types );
        addExchangerBindings( b, STANDARD_EXCHANGERS );
        addExchangerBindings( b, AUTH_EXCEPTION_EXCHANGERS );
    }

    private
    final
    static
    class StandardBindingsLoader
    implements Initializer
    {
        public
        void
        initialize( MingleBinder.Builder b,
                    TypeDefinitionLookup types )
        {
            addStandardBindings( b, types );
        }
    }

    private
    static
    QualifiedTypeName
    expectQname( AtomicTypeReference typ )
    {
        AtomicTypeReference.Name nm = typ.getName();

        if ( nm instanceof QualifiedTypeName ) return (QualifiedTypeName) nm;
        else throw state.createFail( "Type is not qualified:", typ );
    }

    private
    static
    boolean
    isMgType( AtomicTypeReference typ,
              QualifiedTypeName test )
    {
        return expectQname( typ ).equals( test );
    }

    private
    static
    boolean
    isMgValue( AtomicTypeReference typ )
    {
        return isMgType( typ, PrimitiveDefinition.QNAME_VALUE );
    }

    private
    static
    boolean
    isMgNull( AtomicTypeReference typ )
    {
        return isMgType( typ, PrimitiveDefinition.QNAME_NULL );
    } 

    private
    static
    Object
    asOpaqueJavaValue( MingleBinder mb,
                       MingleValue mv,
                       ObjectPath< MingleIdentifier > path )
    {
        QualifiedTypeName qn = null;

        if ( mv instanceof MingleStructure )
        {
            qn = expectQname( ( (MingleStructure) mv ).getType() );
        }
        else if ( mv instanceof MingleEnum )
        {
            qn = expectQname( ( (MingleEnum) mv ).getType() );
        }

        if ( qn != null && mb.hasBindingFor( qn ) )
        {
            return mb.asJavaValue( AtomicTypeReference.create( qn ), mv, path );
        }
        else return mv;
    }

    private
    static
    boolean
    isAssignable( AtomicTypeReference expct,
                  AtomicTypeReference act,
                  MingleBinder mb,
                  ObjectPath< MingleIdentifier > path )
    {
        try { return MingleModels.isAssignable( expct, act, mb.getTypes() ); }
        catch ( NoSuchTypeDefinitionException ex )
        {
            throw new MingleTypeCastException( expct, act, path );
        }
    }

    private
    static
    AtomicTypeReference
    effectiveInboundTypeOf( MingleBinder mb,
                            AtomicTypeReference typ,
                            MingleValue mv,
                            ObjectPath< MingleIdentifier > path )
    {
        // branch order is important, since we want the effective type of
        // MingleValue to be MingleValue, even if mv is itself a structure which
        // we know more about
        if ( typ.equals( MingleModels.TYPE_REF_MINGLE_VALUE ) ) return typ;
        else if ( mv instanceof MingleStructure )
        {
            AtomicTypeReference act = ( (MingleStructure) mv ).getType();

            if ( isAssignable( typ, act, mb, path ) ) return act;
            else throw new MingleTypeCastException( typ, act, path );
        }
        else return typ;
    }

    private
    static
    Object
    asJavaValue( MingleBinder mb,
                 AtomicTypeReference typ,
                 MingleValue mv,
                 ObjectPath< MingleIdentifier > path,
                 boolean useOpaque )
    {
        if ( isMgValue( typ ) && useOpaque )
        {
            return asOpaqueJavaValue( mb, mv, path );
        }
        else 
        {
            typ = effectiveInboundTypeOf( mb, typ, mv, path );
            return mb.asJavaValue( typ, mv, path );
        }
    }

    private
    static
    Object
    asAtomicJavaValue( MingleBinder mb,
                       AtomicTypeReference typ,
                       MingleValue mv,
                       ObjectPath< MingleIdentifier > path,
                       boolean useOpaque )
    {
        if ( mv == null || mv instanceof MingleNull )
        {
            if ( isMgNull( typ ) ) return null;
            else throw new MingleValidationException( "Value is null", path );
        }
        else return asJavaValue( mb, typ, mv, path, useOpaque );
    }

    private
    static
    Object
    asNullableJavaValue( MingleBinder mb,
                         NullableTypeReference typ,
                         MingleValue mv,
                         ObjectPath< MingleIdentifier > path,
                         boolean useOpaque )
    {
        if ( mv == null || mv instanceof MingleNull ) return null;
        else 
        {
            return 
                asJavaValueImpl( 
                    mb, typ.getTypeReference(), mv, path, useOpaque );
        }
    }

    private
    static
    void
    checkEmpty( List< ? > l,
                ListTypeReference typ,
                ObjectPath< MingleIdentifier > path )
    {
        if ( l.isEmpty() && ( ! typ.allowsEmpty() ) )
        {
            throw new MingleValidationException( "list is empty", path );
        }
    }

    private
    static
    Object
    asJavaListValue( MingleBinder mb,
                     ListTypeReference typ,
                     MingleValue mv,
                     ObjectPath< MingleIdentifier > path,
                     boolean useOpaque )
    {
        MingleList ml = 
            MingleModels.asMingleListInstance( typ, mv, true, path );

        List< Object > res = Lang.newList();

        ImmutableListPath< MingleIdentifier > lp = path.startImmutableList();
        MingleTypeReference eltTyp = typ.getElementType();

        for ( MingleValue listVal : ml )
        {
            res.add( asJavaValueImpl( mb, eltTyp, listVal, lp, useOpaque ) );
            lp = lp.next();
        }

        checkEmpty( res, typ, path );
        return Lang.unmodifiableList( res );
    }

    private
    static
    Object
    asJavaValueImpl( MingleBinder mb,
                     MingleTypeReference typ,
                     MingleValue mv,
                     ObjectPath< MingleIdentifier > path,
                     boolean useOpaque )
    {
        if ( typ instanceof AtomicTypeReference )
        {
            return 
                asAtomicJavaValue( 
                    mb, (AtomicTypeReference) typ, mv, path, useOpaque );
        }
        else if ( typ instanceof NullableTypeReference )
        {
            return 
                asNullableJavaValue( 
                    mb, (NullableTypeReference) typ, mv, path, useOpaque ); 
        }
        else if ( typ instanceof ListTypeReference )
        {
            return 
                asJavaListValue( 
                    mb, (ListTypeReference) typ, mv, path, useOpaque );
        }
        else throw state.createFail( "Unhandled type:", typ );
    }
 
    static
    Object
    asJavaValue( MingleBinder mb,
                 MingleTypeReference typ,
                 MingleValue mv,
                 ObjectPath< MingleIdentifier > path,
                 boolean useOpaque )
    {            
        inputs.notNull( mb, "mb" );
        inputs.notNull( typ, "typ" );
        inputs.notNull( path, "path" );

        return asJavaValueImpl( mb, typ, mv, path, useOpaque );
    }

    public
    static
    Object
    asJavaValue( MingleBinder mb,
                 MingleTypeReference typ,
                 MingleValue mv,
                 ObjectPath< MingleIdentifier > path )
    {
        return asJavaValue( mb, typ, mv, path, false );
    }

    public
    static
    Object
    asJavaValue( MingleBinder mb,
                 MingleTypeReference typ,
                 MingleValue mv )
    {
        return 
            asJavaValue( 
                mb, typ, mv, ObjectPath.< MingleIdentifier >getRoot() );
    }

    private
    static
    MingleBindingException
    createOutboundFail( String msg,
                        ObjectPath< String > path )
    {
        return MingleBindingException.createOutbound( msg, path );
    }

    private
    static
    MingleValue
    asOpaqueMingleValue( MingleBinder mb,
                         Object jvObj,
                         ObjectPath< String > path )
    {
        if ( jvObj instanceof Iterable< ? > )
        {
            MingleList.Builder b = new MingleList.Builder();
            ImmutableListPath< String > lp = path.startImmutableList();

            for ( Object o : (Iterable< ? >) jvObj )
            {
                MingleValue mv = o == null
                    ? MingleNull.getInstance()
                    : asMingleValueImpl( mb, MG_TYP_VAL, o, lp, true ); 
                
                b.add( mv );
            }

            return b.build();
        }
        else return MingleModels.asMingleValue( jvObj, path );
    }

    private
    static
    AtomicTypeReference
    effectiveOutboundTypeOf( MingleBinder mb,
                             AtomicTypeReference typ,
                             Object jvObj )
    {
        QualifiedTypeName qn = mb.bindingNameForClass( jvObj.getClass() );

        return qn == null ? typ : AtomicTypeReference.create( qn );
    }

    // jvObj known not-null
    private
    static
    MingleValue
    asMingleValue( MingleBinder mb,
                   AtomicTypeReference typ,
                   Object jvObj,
                   ObjectPath< String > path,
                   boolean useOpaque )
    {
        if ( isMgValue( typ ) && useOpaque )
        {
            QualifiedTypeName qn = mb.bindingNameForClass( jvObj.getClass() );

            if ( qn == null ) return asOpaqueMingleValue( mb, jvObj, path );
            else 
            {
                AtomicTypeReference callTyp = AtomicTypeReference.create( qn );
                return mb.asMingleValue( callTyp, jvObj, path );
            }
        }
        else 
        {
            typ = effectiveOutboundTypeOf( mb, typ, jvObj );
            return mb.asMingleValue( typ, jvObj, path );
        }
    }

    private
    static
    MingleValue
    asAtomicMingleValue( MingleBinder mb,
                         AtomicTypeReference typ,
                         Object jvObj,
                         ObjectPath< String > path,
                         boolean useOpaque )
    {
        if ( jvObj == null ) 
        {
            if ( isMgNull( typ ) ) return MingleNull.getInstance();
            else throw createOutboundFail( "Value is null", path );
        }
        else return asMingleValue( mb, typ, jvObj, path, useOpaque );
    }

    private
    static
    MingleValue
    asNullableMingleValue( MingleBinder mb,
                           NullableTypeReference typ,
                           Object jvObj,
                           ObjectPath< String > path,
                           boolean useOpaque )
    {
        if ( jvObj == null ) return MingleNull.getInstance();
        else 
        {
            return 
                asMingleValueImpl( 
                    mb, typ.getTypeReference(), jvObj, path, useOpaque );
        }
    }

    private
    static
    List< ? >
    expectJavaList( Object obj,
                    ObjectPath< String > path )
    {
        if ( obj instanceof List ) return (List< ? >) obj;
        else
        {
            String typ = obj == null ? "null" : obj.getClass().getName();
            throw createOutboundFail( "Expected List but got " + typ, path );
        }
    }

    private
    static
    MingleList
    asMingleListValue( MingleBinder mb,
                       ListTypeReference typ,
                       Object jvObj,
                       ObjectPath< String > path,
                       boolean useOpaque )
    {
        List< ? > jList = expectJavaList( jvObj, path );

        if ( jList.isEmpty() && ( ! typ.allowsEmpty() ) )
        {
            throw createOutboundFail( "List is empty", path );
        }
        else
        {
            ImmutableListPath< String > lp = path.startImmutableList();
            MingleTypeReference eltTyp = typ.getElementType();
            MingleList.Builder b = new MingleList.Builder();
    
            for ( Object listVal : jList )
            {
                b.add( 
                    asMingleValueImpl( mb, eltTyp, listVal, lp, useOpaque ) );
                lp = lp.next();
            }
    
            return b.build();
        }
    }

    private
    static
    MingleValue
    asMingleValueImpl( MingleBinder mb,
                       MingleTypeReference typ,
                       Object jvObj,
                       ObjectPath< String > path,
                       boolean useOpaque )
    {
        if ( typ instanceof AtomicTypeReference )
        {
            return 
                asAtomicMingleValue( 
                    mb, (AtomicTypeReference) typ, jvObj, path, useOpaque );
        }
        else if ( typ instanceof NullableTypeReference )
        {
            return 
                asNullableMingleValue(
                    mb, (NullableTypeReference) typ, jvObj, path, useOpaque );
        }
        else if ( typ instanceof ListTypeReference )
        {
            return 
                asMingleListValue( 
                    mb, (ListTypeReference) typ, jvObj, path, useOpaque );
        }
        else throw state.createFail( "Unhandled type:", typ );
    }

    static
    MingleValue
    asMingleValue( MingleBinder mb,
                   MingleTypeReference typ,
                   Object jvObj,
                   ObjectPath< String > path,
                   boolean useOpaque )
    {
        inputs.notNull( mb, "mb" );
        inputs.notNull( typ, "typ" );
        inputs.notNull( path, "path" );

        return asMingleValueImpl( mb, typ, jvObj, path, useOpaque );
    }

    public
    static
    MingleValue
    asMingleValue( MingleBinder mb,
                   MingleTypeReference typ,
                   Object jvObj,
                   ObjectPath< String > path )
    {
        return asMingleValue( mb, typ, jvObj, path, false );
    }

    public
    static
    MingleValue
    asMingleValue( MingleBinder mb,
                   MingleTypeReference typ,
                   Object jvObj )
    {
        return asMingleValue( mb, typ, jvObj, ObjectPath.< String >getRoot() );
    }

    public
    static
    void
    toFile( MingleBinder mb,
            Object jvObj,
            MingleCodec codec,
            FileWrapper dest,
            ProcessActivity.Context ctx,
            IoProcessor ioProc,
            Runnable onComp )
    {
        inputs.notNull( mb, "mb" );
        inputs.notNull( jvObj, "jvObj" );
        inputs.notNull( codec, "codec" );
        inputs.notNull( dest, "dest" );
        inputs.notNull( ctx, "ctx" );
        inputs.notNull( ioProc, "ioProc" );
        inputs.notNull( onComp, "onComp" );
        
        MingleStruct ms = (MingleStruct) asMingleValue( mb, jvObj );

        MingleCodecs.toFile( codec, ms, dest, ctx, ioProc, onComp );
    }

    public
    static
    Object
    asJavaValue( FieldDefinition fd,
                 MingleSymbolMap flds,
                 MingleBinder mb,
                 ObjectPath< MingleIdentifier > basePath,
                 boolean useOpaque )
    {
        inputs.notNull( fd, "fd" );
        inputs.notNull( flds, "flds" );
        inputs.notNull( mb, "mb" );
        inputs.notNull( basePath, "basePath" );

        MingleIdentifier fldId = fd.getName();

        MingleValue mv = MingleModels.get( flds, fd );
        ObjectPath< MingleIdentifier > path = basePath.descend( fldId );

        MingleTypeReference typ = fd.getType();

        return asJavaValueImpl( mb, typ, mv, path, useOpaque );
    }

    public
    static
    Object
    asJavaValue( FieldDefinition fd,
                 MingleSymbolMap flds,
                 MingleBinder mb,
                 ObjectPath< MingleIdentifier > basePath )
    {
        return asJavaValue( fd, flds, mb, basePath, false );
    }

    private
    static
    void
    checkSetFieldNullValue( FieldDefinition fd,
                            String jvFldId,
                            ObjectPath< String > path )
    {
        MingleTypeReference typ = fd.getType();

        boolean isNullable = typ instanceof NullableTypeReference;

        boolean allowsEmpty = 
            typ instanceof ListTypeReference &&
            ( (ListTypeReference) typ ).allowsEmpty();
        
        boolean hasDefl = fd.getDefault() != null;

        if ( ! ( isNullable || allowsEmpty || hasDefl ) )
        {
            throw
                MingleBindingException.
                    createOutbound( "Value is null", path.descend( jvFldId ) );
        }
    }

    static
    void
    setField( FieldDefinition fd,
              String jvFldId,
              Object val,
              MingleSymbolMapBuilder b,
              MingleBinder mb,
              ObjectPath< String > path,
              boolean useOpaque )
    {
        state.notNull( fd, "fd" );
        state.notNull( jvFldId, "jvFldId" );
        state.notNull( b, "b" );
        state.notNull( mb, "mb" );
        state.notNull( path, "path" );

        if ( val == null ) checkSetFieldNullValue( fd, jvFldId, path );
        else
        {
            MingleValue mgVal = 
                asMingleValueImpl( 
                    mb, fd.getType(), val, path.descend( jvFldId ), useOpaque );
    
            if ( mgVal != null && ! ( mgVal instanceof MingleNull ) )
            {
                b.set( fd.getName(), mgVal );
            }
        }
    }

    // Used primarily by generated code (and our manual testing to emulate it)
    // to initialize static fields with FieldDefinitions encoded in generated
    // source as base64 literals. We rethrow any exceptions as runtimes so that
    // generated code doesn't have to wrap this in try/catch/rethrow
    public
    static
    FieldDefinition
    decodeBase64FieldDef( CharSequence base64 )
    {
        inputs.notNull( base64, "base64" );

        try
        {
            ByteBuffer bb = enc.asByteBuffer( base64 );

            MingleStruct ms = 
                MingleCodecs.fromByteBuffer( codec, bb, MingleStruct.class );
            
            return TypeDefinitions.asFieldDefinition( ms );
        }
        catch ( Exception ex ) { throw new RuntimeException( ex ); }
    }

    private
    static
    RuntimeException
    createJavaValidationException( String msg,
                                   ObjectPath< String > path )
    {
        throw MingleBindingException.createOutbound( msg, path );
    }

    private
    static
    < V >
    V
    validateNotNull( V obj,
                     ObjectPath< String > path )
    {
        if ( obj == null )
        {
            throw createJavaValidationException( "Value is null", path );
        }
        else return obj;
    }

    private
    static
    void
    validateString( CharSequence str,
                    MingleRegexRestriction r,
                    ObjectPath< String > path )
    {
        if ( ! r.matches( str ) )
        {
            throw 
                createJavaValidationException(
                    "Value does not match " + r.getExternalForm() + 
                    ": " + Lang.getRfc4627String( str ),
                    path 
                );
        }
    }

    private
    static
    void
    validateRange( Object obj,
                   MingleRangeRestriction r,
                   ObjectPath< String > path )
    {
        if ( ! r.acceptsJavaValue( obj ) )
        {
            throw
                createJavaValidationException(
                    "Value is not in range " + r.getExternalForm() + 
                    ": " + obj, 
                    path 
                );
        }
    }

    private
    static
    void
    validateRestriction( Object obj,
                         MingleValueRestriction r,
                         ObjectPath< String > path )
    {
        if ( r instanceof MingleRegexRestriction )
        {
            validateString( 
                (CharSequence) obj, (MingleRegexRestriction) r, path );
        }
        else if ( r instanceof MingleRangeRestriction )
        {
            validateRange( obj, (MingleRangeRestriction) r, path );
        }
        else state.fail( "Unhandled restriction:", r );
    }

    private
    static
    < V >
    V
    validateAtomicFieldValue( V obj,
                              AtomicTypeReference typ,
                              ObjectPath< String > path )
    {
        validateNotNull( obj, path );

        MingleValueRestriction r = typ.getRestriction();
        if ( r != null ) validateRestriction( obj, r, path );

        return obj;
    }

    private
    static
    < V >
    V
    validateNullableFieldValue( V obj,
                                NullableTypeReference typ,
                                ObjectPath< String > path )
    {
        if ( obj == null ) return null;
        else return validateFieldValue( obj, typ.getTypeReference(), path );
    }

    private
    static
    < V >
    V
    validateListFieldValue( V obj,
                            ListTypeReference typ,
                            ObjectPath< String > path )
    {
        // make sure obj itself is not null and cast it as the expected list
        List< ? > l = (List< ? >) validateNotNull( obj, path ); 
        
        if ( l.isEmpty() )
        {
            if ( typ.allowsEmpty() ) return obj;
            else throw createJavaValidationException( "list is empty", path );
        }
        else
        {
            ImmutableListPath< String > lp = path.startImmutableList();
            MingleTypeReference eltTyp = typ.getElementType();

            for ( Object elt : l )
            {
                validateFieldValue( elt, eltTyp, lp );
                lp = lp.next();
            }

            return obj;
        }
    } 

    // Validates obj as appropriate to typ in terms of nullity, list
    // constraints, and other constraints as provided. One thing that is not
    // checked is that the type of V is compatible with the parameter 'typ', as
    // the assumption is that the binding code calling this method has already
    // verified that part on its own. Note that this also means this method
    // assumes that if typ is a ListTypeReference then obj may be cast as
    // java.util.List< ? >
    public
    static
    < V >
    V
    validateFieldValue( V obj,
                        MingleTypeReference typ,
                        ObjectPath< String > path )
    {
        if ( typ instanceof AtomicTypeReference )
        {
            return 
                validateAtomicFieldValue( 
                    obj, (AtomicTypeReference) typ, path );
        }
        else if ( typ instanceof NullableTypeReference )
        {
            return 
                validateNullableFieldValue( 
                    obj, (NullableTypeReference) typ, path );
        }
        else if ( typ instanceof ListTypeReference )
        {
            return validateListFieldValue( obj, (ListTypeReference) typ, path );
        }
        else throw new UnsupportedOperationException( "Unimplemented" );
    }

    // Putting this here as a public proxy from the method as currently housed
    // in MingleBinder
    public
    static
    QualifiedTypeName
    bindingNameForClass( Class< ? > cls,
                         MingleBinder mb )
    {
        inputs.notNull( cls, "cls" );
        inputs.notNull( mb, "mb" );

        return mb.bindingNameForClass( cls );
    }

    private
    static
    AtomicTypeReference
    expectKeyedType( Class< ? > cls,
                     MingleBinder mb )
    {
        QualifiedTypeName qn = bindingNameForClass( cls, mb );
        state.isFalse( qn == null, "No binding keyed for", cls );

        return AtomicTypeReference.create( qn );
    }

    public
    static
    < V >
    V
    asJavaValue( MingleBinder mb,
                 Class< V > cls,
                 MingleValue mv )
    {
        inputs.notNull( mb, "mb" );
        inputs.notNull( cls, "cls" );
        inputs.notNull( mv, "mv" );

        AtomicTypeReference typ = expectKeyedType( cls, mb );
        return cls.cast( asJavaValue( mb, typ, mv ) );
    }

    public
    static
    < V >
    void
    fromFile( final MingleBinder mb,
              final Class< V > cls,
              MingleCodec codec,
              FileWrapper src,
              ProcessActivity.Context ctx,
              IoProcessor ioProc,
              final ObjectReceiver< ? super V > recv )
    {
        inputs.notNull( mb, "mb" );
        inputs.notNull( cls, "cls" );
        inputs.notNull( codec, "codec" );
        inputs.notNull( src, "src" );
        inputs.notNull( ctx, "ctx" );
        inputs.notNull( ioProc, "ioProc" );
        inputs.notNull( recv, "recv" );

        MingleCodecs.fromFile(
            codec, 
            src,
            MingleStruct.class,
            ctx,
            ioProc,
            new ObjectReceiver< MingleStruct >() {
                public void receive( MingleStruct ms ) throws Exception {
                    recv.receive( asJavaValue( mb, cls, ms ) );
                }
            }
        );
    }

    public
    static
    MingleValue
    asMingleValue( MingleBinder mb,
                   Object jvObj )
    {
        inputs.notNull( mb, "mb" );
        inputs.notNull( jvObj, "jvObj" );

        AtomicTypeReference typ = expectKeyedType( jvObj.getClass(), mb );
        return asMingleValue( mb, typ, jvObj );
    }

    private
    static
    void
    addExchangers( List< MingleTypeReference > typs,
                   List< MingleValueExchanger > l,
                   Iterable< MingleValueExchanger > iter )
    {
        for ( MingleValueExchanger exch : iter )
        {
            l.add( exch );
            typs.add( exch.getMingleType() );
        }
    }

    private
    static
    List< MingleTypeReference >
    createServiceExceptionBindings( List< MingleTypeReference > typs )
    {
        List< MingleTypeReference > res = Lang.newList( typs );
        res.remove( StandardException.MINGLE_TYPE );

        return Lang.unmodifiableList( res );
    }

    static
    {
        List< MingleValueExchanger > l = Lang.newList();
        List< MingleTypeReference > typs = Lang.newList();

        typs.add( StandardException.MINGLE_TYPE );

        addExchangers( typs, l, MingleModels.getExchangers() );
        addExchangers( typs, l, MingleServices.getExchangers() );

        STANDARD_BINDINGS = Lang.unmodifiableList( typs );
        STANDARD_EXCHANGERS = Lang.unmodifiableList( l );

        SVC_EXCEPTION_BINDINGS = createServiceExceptionBindings( typs );
    }

    static
    {
        List< MingleTypeReference > typs = Lang.newList();
        List< MingleValueExchanger > l = Lang.newList();

        addExchangers( typs, l, MingleServices.getAuthExceptionExchangers() );

        AUTH_EXCEPTION_BINDINGS = Lang.unmodifiableList( typs );
        AUTH_EXCEPTION_EXCHANGERS = Lang.unmodifiableList( l );
    }
}
