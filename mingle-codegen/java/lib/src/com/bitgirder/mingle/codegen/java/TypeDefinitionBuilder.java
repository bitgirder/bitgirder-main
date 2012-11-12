package com.bitgirder.mingle.codegen.java;

import static com.bitgirder.mingle.codegen.java.CodegenConstants.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.io.IoUtils;
import com.bitgirder.io.Base64Encoder;

import com.bitgirder.mingle.model.TypeDefinition;
import com.bitgirder.mingle.model.TypeDefinitions;
import com.bitgirder.mingle.model.FieldDefinition;
import com.bitgirder.mingle.model.EnumDefinition;
import com.bitgirder.mingle.model.PrimitiveDefinition;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleEnum;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleString;
import com.bitgirder.mingle.model.MingleInt64;
import com.bitgirder.mingle.model.MingleInt32;
import com.bitgirder.mingle.model.MingleDouble;
import com.bitgirder.mingle.model.MingleFloat;
import com.bitgirder.mingle.model.MingleList;
import com.bitgirder.mingle.model.MingleBoolean;
import com.bitgirder.mingle.model.MingleTimestamp;
import com.bitgirder.mingle.model.MingleNamespace;
import com.bitgirder.mingle.model.MingleIdentifierFormat;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.ListTypeReference;
import com.bitgirder.mingle.model.NullableTypeReference;
import com.bitgirder.mingle.model.QualifiedTypeName;

import com.bitgirder.mingle.runtime.MingleRuntimes;

import com.bitgirder.mingle.codec.MingleCodecs;
import com.bitgirder.mingle.codec.MingleCodec;
import com.bitgirder.mingle.codec.MingleCodecFactory;

import java.util.List;
import java.util.Map;
import java.util.Iterator;

import java.nio.ByteBuffer;

abstract
class TypeDefinitionBuilder< T extends TypeDefinition >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static Base64Encoder enc = new Base64Encoder();

    final static JvType TYPE_BOUND_STRUCT =
        JvQname.create( "com.bitgirder.mingle.bind", "BoundStruct" );

    private final static JvId JV_ID_DECODE_BASE64_FIELD_DEF =
        new JvId( "decodeBase64FieldDef" );

    private int idSeq;

    private T def;
    private CodegenContext ctx;

    final T typeDef() { return def; }
    final CodegenContext context() { return ctx; }

    final void code( Object... args ) { context().code( args ); }

    final
    MingleTypeReference
    typeRefOf( TypeDefinition td )
    {
        inputs.notNull( td, "td" );
        return AtomicTypeReference.create( td.getName() );
    }
    
    final
    Map< MingleIdentifier, FieldGeneratorParameters >
    fieldParamsByName( List< FieldGeneratorParameters > l )
    {
        inputs.noneNull( l, "l" );

        Map< MingleIdentifier, FieldGeneratorParameters > res = Lang.newMap();

        for ( FieldGeneratorParameters p : l ) res.put( p.name(), p );

        return res;
    }

    final
    < P extends TypeMaskedGeneratorParameters >
    P
    getGeneratorParameters( TypeDefinition td,
                            Class< P > cls )
    {
        inputs.notNull( td, "td" );
        inputs.notNull( cls, "cls" );

        MingleTypeReference typ = typeRefOf( td );

        return ctx.codegenControl().getGeneratorParameters( typ, cls );
    }

    final
    < P extends TypeMaskedGeneratorParameters >
    P
    getGeneratorParameters( Class< P > cls )
    {
        return getGeneratorParameters( typeDef(), cls );
    }

    final
    JvId
    nextJvId()
    {
        return new JvId( "_id" + idSeq++ );
    }

    final
    JvId
    asJvId( MingleIdentifier id )
    {
        return
            new JvId(
                MingleModels.
                    format( id, MingleIdentifierFormat.LC_CAMEL_CAPPED )
            );
    }

    private
    static
    interface JvFieldFactory
    {
        public
        JvField
        generateField();
    }

    final
    CharSequence
    upcaseIdent( MingleIdentifier id )
    {
        return
            MingleModels.format( id, MingleIdentifierFormat.LC_UNDERSCORE ).
                toString().
                toUpperCase();
    }

    // A lazy map for common statically instantiated fields in generated
    // classes
    final
    class JvIdMapper
    {
        private final List< JvField > flds;

        private final Map< MingleIdentifier, JvId > identIds = Lang.newMap();

        private final Map< MingleTypeReference, JvId > typeRefIds =
            Lang.newMap();

        private final Map< QualifiedTypeName, JvId > qnameIds = Lang.newMap();

        private final Map< MingleNamespace, JvId > nsIds = Lang.newMap();

        // Ends up using just FieldDefinition's reference equality, which is
        // fine for us
        private final Map< FieldDefinition, JvId > fldDefIds = Lang.newMap();

        private final Map< JvId, JvId > jvFldPathIds = Lang.newMap();

        private JvId inputsId;

        JvIdMapper( List< JvField > flds ) { this.flds = flds; }
        JvIdMapper( JvClass cls ) { this( cls.fields ); }

        private
        < K >
        JvId
        lazyGet( K key,
                 Map< K, JvId > m,
                 JvFieldFactory ff )
        {
            JvId res = m.get( key );

            if ( res == null )
            {
                JvField f = ff.generateField();
                if ( f.name == null ) f.name = nextJvId();
                res = f.name;
                
                flds.add( f );
                m.put( key, f.name );
            }

            return state.notNull( res );
        }

        private
        < K >
        JvId
        lazyGet( K key,
                 Map< K, JvId > m,
                 final JvType fType,
                 final CharSequence extForm )
        {
            return lazyGet( key, m,
                new JvFieldFactory() {
                    public JvField generateField()
                    {
                        JvField f = JvField.createConstField();
                        f.type = fType;
        
                        f.assign =
                            JvFuncCall.create(
                                new JvAccess( f.type, JV_ID_CREATE ),
                                new JvString( extForm )
                            );
                        
                        return f;
                    }
                }
            );
        }

        final
        JvId
        idFor( final MingleIdentifier id )
        {
            return
                lazyGet(
                    id, identIds, JV_QNAME_MG_IDENTIFIER, id.getExternalForm() 
                );
        }

        final
        JvId
        idFor( MingleTypeReference ref )
        {
            return 
                lazyGet(
                    ref, typeRefIds, JV_QNAME_MG_TYPE_REF, ref.getExternalForm()
                );
        }

        final
        JvId
        idFor( QualifiedTypeName qn )
        {
            return
                lazyGet(
                    qn, qnameIds, JV_QNAME_MG_QNAME, qn.getExternalForm() );
        }

        final
        JvId
        idFor( MingleNamespace ns )
        {
            return 
                lazyGet( 
                    ns, nsIds, JV_QNAME_MG_NAMESPACE, ns.getExternalForm() );
        }

        final
        JvId
        idFor( final FieldDefinition fd )
        {
            return lazyGet( fd, fldDefIds,
                new JvFieldFactory() {
                    public JvField generateField() 
                    {
                        JvField f = JvField.createConstField();
                        f.type = JV_QNAME_FIELD_DEF;
                
                        f.assign = 
                            JvFuncCall.create(
                                new JvAccess( 
                                    JV_QNAME_MINGLE_BINDERS, 
                                    JV_ID_DECODE_BASE64_FIELD_DEF
                                ),
                                new JvString( toBase64( fd ) )
                            );
                        
                        return f;
                    }
                }
            );
        }

        final
        JvId
        jvFldPathFor( final JvId fldId )
        {
            return lazyGet( fldId, jvFldPathIds,
                new JvFieldFactory() {
                    public JvField generateField()
                    {
                        JvField f = JvField.createConstField();
                        f.type = JV_TYPE_OBJ_PATH_STRING;

                        f.assign =
                            JvFuncCall.create(
                                new JvAccess( 
                                    JV_QNAME_OBJ_PATH, JV_ID_GET_ROOT ),
                                new JvString( fldId )
                            );

                        return f;
                    }
                }
            );
        }

        private
        void
        initInputs()
        {
            JvField f = JvField.createConstField();
            f.name = nextJvId();
            f.type = JV_QNAME_INPUTS;

            f.assign = JvInstantiate.create( f.type );

            f.validate();
            flds.add( f );

            inputsId = f.name;
        }

        final
        JvExpression
        jvNotNull( JvId paramId )
        {
            inputs.notNull( paramId, "paramId" );

            if ( inputsId == null ) initInputs();
            
            return
                JvFuncCall.create(
                    new JvAccess( inputsId, JV_ID_NOT_NULL ),
                    paramId,
                    new JvString( paramId )
                );
        }
    }

    final
    JvExpression
    expBoolLiteral( boolean b )
    {
        return new JvId( b ? "true" : "false" );
    }

    final
    JvExpression
    expValidateFieldValue( JvId valId,
                           JvId errId,
                           FieldDefinition fd,
                           JvIdMapper idMap )
    {
        inputs.notNull( valId, "valId" );
        inputs.notNull( fd, "fd" );
        inputs.notNull( idMap, "idMap" );

        return
            JvFuncCall.create(
                new JvAccess( 
                    JV_QNAME_MINGLE_BINDERS, JV_ID_VALIDATE_FIELD_VALUE ),
                valId,
                idMap.idFor( fd.getType() ),
                idMap.jvFldPathFor( errId )
            );
    }

    final
    JvExpression
    expCastUnchecked( JvExpression e )
    {
        inputs.notNull( e, "e" );

        return
            JvFuncCall.create(
                new JvAccess( JV_QNAME_BG_LANG, JV_ID_CAST_UNCHECKED ), e );
    }

    final
    JvExpression
    expValidateFieldValue( JvField f,
                           JvIdMapper idMap )
    {
        return expValidateFieldValue( f.name, f.name, f.mgField, idMap );
    }

    final
    static
    class AsJavaValCtx
    {
        final JvField f;
        final JvId smId; // the symbol map
        final JvId mbId; // the mingle binder
        final JvId pathId; // the object path

        AsJavaValCtx( JvField f,
                      JvId smId,
                      JvId mbId,
                      JvId pathId )
        {
            this.f = inputs.notNull( f, "f" );
            this.smId = inputs.notNull( smId, "smId" );
            this.mbId = inputs.notNull( mbId, "mbId" );
            this.pathId = inputs.notNull( pathId, "pathId" );
        }
    }

    final
    JvExpression
    expAsJavaValue( AsJavaValCtx ctx,
                    JvIdMapper idMap )
    {
        inputs.notNull( ctx, "ctx" );
        inputs.notNull( idMap, "idMap" );

        return
            expCastUnchecked(
                JvFuncCall.create(
                    new JvAccess(
                        JV_QNAME_MINGLE_BINDERS, JV_ID_AS_JAVA_VALUE ),
                    idMap.idFor( ctx.f.mgField ),
                    ctx.smId,
                    ctx.mbId,
                    ctx.pathId,
                    expBoolLiteral( useOpaqueJavaType( ctx.f ) )
                )
            );
    }

    final
    JvPackage
    jvPackageOf( TypeDefinition td )
    {
        return context().jvPackageOf( td );
    }

    final
    JvTypeName
    jvTypeNameOf( QualifiedTypeName qn )
    {
        return context().jvTypeNameOf( qn );
    }

    final
    JvTypeName
    jvTypeNameOf( TypeDefinition td )
    {
        return jvTypeNameOf( td.getName() );
    }

    private
    JvType
    jvPrimTypeOf( QualifiedTypeName qn )
    {
        PrimitiveDefinition pd = PrimitiveDefinition.forName( qn );

        if ( pd == null ) return null;
        else
        {
            String nm = pd.getName().getName().get( 0 ).toString();

            if ( nm.equals( "Value" ) ) return JV_QNAME_MINGLE_VALUE;
            else if ( nm.equals( "Null" ) ) return JV_QNAME_VOID;
            else if ( nm.equals( "String" ) ) return JV_QNAME_JSTRING;
            else if ( nm.equals( "Int64" ) ) return JvPrimitiveType.LONG;
            else if ( nm.equals( "Int32" ) ) return JvPrimitiveType.INT;
            else if ( nm.equals( "Double" ) ) return JvPrimitiveType.DOUBLE;
            else if ( nm.equals( "Float" ) ) return JvPrimitiveType.FLOAT;
            else if ( nm.equals( "Boolean" ) ) return JvPrimitiveType.BOOLEAN;
            else if ( nm.equals( "Buffer" ) ) return JV_QNAME_BYTE_BUFFER;
            else if ( nm.equals( "Timestamp" ) ) return JV_QNAME_MG_TIMESTAMP;
            else return null;
        }
    }

    private
    JvType
    jvTypeOf( QualifiedTypeName qn,
              boolean isOuterMost )
    {
        JvType res = jvPrimTypeOf( qn );

        if ( res == null ) return context().jvQnameOf( qn );
        else
        {
            if ( res instanceof JvPrimitiveType && ( ! isOuterMost ) )
            {
                res = ( (JvPrimitiveType) res ).boxed;
            }

            return res;
        }
    }

    final
    JvType
    jvTypeOf( QualifiedTypeName qn )
    {
        return jvTypeOf( qn, false ); 
    }

    private
    JvType
    jvTypeOf( AtomicTypeReference typ,
              boolean isOuterMost )
    {
        return jvTypeOf( (QualifiedTypeName) typ.getName(), isOuterMost );
    }

    private
    JvType
    jvListTypeOf( ListTypeReference ref )
    {
        return
            JvTypeExpression.withParams( 
                JV_QNAME_JLIST, 
                jvTypeOf( ref.getElementTypeReference(), false ) 
            );
    }

    final
    JvType
    jvTypeOf( MingleTypeReference ref,
              boolean isOuterMost )
    {
        if ( ref instanceof AtomicTypeReference )
        {
            QualifiedTypeName qn =
                (QualifiedTypeName) ( (AtomicTypeReference) ref ).getName();

            return jvTypeOf( qn, isOuterMost );
        }
        else if ( ref instanceof NullableTypeReference )
        {
            return 
                jvTypeOf( 
                    ( (NullableTypeReference) ref ).getTypeReference(),
                    false
                );
        }
        else if ( ref instanceof ListTypeReference )
        {
            return jvListTypeOf( (ListTypeReference) ref );
        }
        else throw state.createFail( "Unhandled typ ref:", ref );
    }
    
    final
    JvType
    jvTypeOf( MingleTypeReference ref )
    {
        return jvTypeOf( ref, true );
    }

    final
    JvId
    createMethodName( CharSequence pref,
                      CharSequence suff )
    {
        StringBuilder sb = new StringBuilder();
        sb.append( pref );
        sb.append( Character.toUpperCase( suff.charAt( 0 ) ) );

        if ( suff.length() > 1 ) 
        {
            sb.append( suff.subSequence( 1, suff.length() ) );
        }

        return new JvId( sb );
    }

    final
    JvId
    setterNameFor( JvField f )
    {
        return createMethodName( "set", f.name );
    }

    final
    boolean
    useOpaqueJavaType( JvField f )
    {
        inputs.notNull( f, "f" );
        return f.fgParams != null && f.fgParams.useOpaqueJavaType();
    }

    final JvType jvOpaqueType() { return JV_QNAME_OBJECT; }

    private
    JvType
    jvFieldTypeOf( MingleTypeReference typ,
                   FieldGeneratorParameters fgParams )
    {
        if ( fgParams != null && fgParams.useOpaqueJavaType() )
        {
            return jvOpaqueType();
        }
        else
        {
            JvType res = jvTypeOf( typ );
    
            if ( res instanceof JvPrimitiveType )
            {
                res = ( (JvPrimitiveType) res ).boxed;
            }
    
            return res;
        }
    }

    // fd not null; fgParams maybe null
    final
    JvField
    asJvField( FieldDefinition fd,
               FieldGeneratorParameters fgParams )
    {
        state.notNull( fd, "fd" );
        JvField jf = new JvField();

        jf.type = jvFieldTypeOf( fd.getType(), fgParams );
        jf.name = asJvId( fd.getName() );
        jf.mgField = fd;
        jf.fgParams = fgParams;

        return jf;
    }

    final
    JvField
    asJvField( FieldDefinition fd )
    {
        return asJvField( fd, null ); 
    }

    private
    JvExpression
    jvPrimitiveLiteralFor( JvPrimitiveType jvTyp,
                           MingleValue mv )
    {
        if ( jvTyp == JvPrimitiveType.LONG )
        {
            return new JvNumber( ( (MingleInt64) mv ).toString() + "L" );
        }
        else if ( jvTyp == JvPrimitiveType.INT )
        {
            return new JvNumber( ( (MingleInt32) mv ).toString() );
        }
        else if ( jvTyp == JvPrimitiveType.DOUBLE )
        {
            return new JvNumber( ( (MingleDouble) mv ).toString() + "d" );
        }
        else if ( jvTyp == JvPrimitiveType.FLOAT )
        {
            return new JvNumber( ( (MingleFloat) mv ).toString() + "f" );
        }
        else if ( jvTyp == JvPrimitiveType.BOOLEAN )
        {
            return 
                new JvId( 
                    Boolean.toString( ( (MingleBoolean) mv ).booleanValue() ) );
        }
        else throw state.createFail( "Unhandled prim:", jvTyp );
    }

    private
    JvExpression
    jvPrimitiveExpressionOf( JvPrimitiveType jvTyp,
                             MingleValue mv,
                             boolean isOuterMost )
    {
        JvExpression e = jvPrimitiveLiteralFor( jvTyp, mv );

        if ( isOuterMost ) return e;
        else
        {
            return
                JvFuncCall.create( 
                    new JvAccess( jvTyp.boxed, JV_ID_VALUE_OF ), 
                    e 
                );
        }
    }

    private
    JvExpression
    jvTimestampExpressionOf( MingleTimestamp ts )
    {
        return
            JvFuncCall.create(
                new JvAccess( JV_QNAME_MG_TIMESTAMP, JV_ID_CREATE ),
                new JvString( ts.getRfc3339String() )
            );
    }

    private
    TypeDefinition
    expectTypeDefFor( MingleTypeReference typ )
    {
        QualifiedTypeName qn = 
            (QualifiedTypeName) MingleModels.typeNameIn( typ );

        return ctx.runtime().getTypes().expectType( qn );
    }

    private
    boolean
    isEnumType( AtomicTypeReference typ )
    {
        return expectTypeDefFor( typ ) instanceof EnumDefinition;
    }

    // This just re-applies the algorithm used in EnumDefinitionBuilder; if at
    // some point we allow custom overrides or name mappings, we'll need to make
    // sure that this method uses the same algs or consults the same mappings as
    // EnumDefinitionBuilder
    private
    JvExpression
    jvEnumExpressionOf( AtomicTypeReference typ,
                        MingleValue mv )
    {
        MingleEnum me = (MingleEnum) mv;

        return
            new JvAccess(
                jvTypeNameOf( (QualifiedTypeName) me.getType().getName() ),
                new JvId( MingleModels.asJavaEnumName( me.getValue() ) )
            );
    }

    private
    JvExpression
    jvConstExpressionOf( AtomicTypeReference typ,
                         MingleValue mv,
                         boolean isOuterMost )
    {
        JvType jvTyp = jvTypeOf( typ, isOuterMost );

        if ( jvTyp.equals( JV_QNAME_JSTRING ) )
        {
            return new JvString( ( (MingleString) mv ).toString() );
        }
        else if ( jvTyp.equals( JV_QNAME_MG_TIMESTAMP ) )
        {
            return jvTimestampExpressionOf( (MingleTimestamp) mv );
        }
        else if ( jvTyp instanceof JvPrimitiveType )
        {
            return 
                jvPrimitiveExpressionOf( 
                    (JvPrimitiveType) jvTyp, mv, isOuterMost );
        }
        else if ( jvTyp.equals( JV_QNAME_LONG ) )
        {
            return jvPrimitiveExpressionOf( JvPrimitiveType.LONG, mv, false );
        }
        else if ( jvTyp.equals( JV_QNAME_INT ) )
        {
            return jvPrimitiveExpressionOf( JvPrimitiveType.INT, mv, false );
        }
        else if ( jvTyp.equals( JV_QNAME_DOUBLE ) )
        {
            return jvPrimitiveExpressionOf( JvPrimitiveType.DOUBLE, mv, false );
        }
        else if ( jvTyp.equals( JV_QNAME_FLOAT ) )
        {
            return jvPrimitiveExpressionOf( JvPrimitiveType.FLOAT, mv, false );
        }
        else if ( isEnumType( typ ) ) return jvEnumExpressionOf( typ, mv );
        else throw state.createFail( "Unhandled const, jvTyp:", jvTyp );
    }

    private
    JvExpression
    jvListConstExpressionOf( ListTypeReference mgTyp,
                             MingleList ml )
    {
        MingleTypeReference typ = mgTyp.getElementTypeReference();
        Iterator< MingleValue > it = ml.iterator();

        JvExpression res =
            JvInstantiate.create(
                JvTypeExpression.withParams(
                    JV_TYPE_IMMUTABLE_LIST_BUILDER, jvTypeOf( typ, false ) ) );
        
        while ( it.hasNext() )
        {
            res = 
                JvFuncCall.create(
                    new JvAccess( res, JV_ID_ADD ),
                    jvConstExpressionOf( typ, it.next(), false )
                );
        }

        return JvFuncCall.create( new JvAccess( res, JV_ID_BUILD ) );
    }

    // mv known non-null at this point
    private
    JvExpression
    jvConstExpressionOf( MingleTypeReference mgTyp,
                         MingleValue mv,
                         boolean isOuterMost )
    {
        if ( mgTyp instanceof ListTypeReference )
        {
            return 
                jvListConstExpressionOf( 
                    (ListTypeReference) mgTyp, (MingleList) mv );
        }
        else if ( mgTyp instanceof NullableTypeReference )
        {
            MingleTypeReference t =
                    ( (NullableTypeReference) mgTyp ).getTypeReference();

            return jvConstExpressionOf( t, mv, false );
        }
        else if ( mgTyp instanceof AtomicTypeReference )
        {
            return 
                jvConstExpressionOf( 
                    (AtomicTypeReference) mgTyp, mv, isOuterMost );
        }
        else throw state.createFail( "Unhandled type:", mgTyp );
    }

    private
    MingleValue
    getDefault( FieldDefinition fd )
    {
        MingleValue res = fd.getDefault();

        if ( res == null )
        {
            MingleTypeReference t = fd.getType();

            if ( t instanceof ListTypeReference && 
                 ( (ListTypeReference) t ).allowsEmpty() )
            {
                res = MingleModels.getEmptyList();
            }
        }

        return res;
    }

    // Creates and adds a default value constant
    // field to the specified class based on the default value present (or not)
    // in fd.
    final
    void
    addJvConstField( FieldDefinition fd,
                     JvType jvTyp,
                     Map< MingleIdentifier, JvId > jvFldDeflIds,
                     JvClass cls )
    {
        MingleValue defl = getDefault( fd );

        if ( defl != null )
        {
            JvField res = JvField.createConstField();
            res.name = nextJvId();
            res.type = jvTyp;
            res.assign = jvConstExpressionOf( fd.getType(), defl, true );
 
            jvFldDeflIds.put( fd.getName(), res.name );
            cls.fields.add( res );
        }
    }

    final
    JvExpression
    optUnbox( JvType targTyp,
              JvExpression e )
    {
        if ( targTyp instanceof JvPrimitiveType )
        {
            return 
                new JvParExpression(
                    new JvCast( ( (JvPrimitiveType) targTyp ).boxed, e ) );
        }
        else return e;
    }

    final
    JvId
    addParam( JvConstructor cons,
              JvType typ )
    {
        JvId res = nextJvId();
        cons.params.add( JvParam.create( res, typ ) );

        return res;
    }

    // We store the serialized field def as a string literal. Another
    // option would be to store it in generated code as a ByteBuffer literal
    // (aka, ByteBuffer.wrap( new byte[] { ... } )), but it seems advantageous
    // to have the buffer as a string in the generated code to aid debugging,
    // particularly to simplify copy pasting into other tools to inspect the
    // data.
    private
    CharSequence
    toBase64( FieldDefinition fd )
    {
        MingleStruct mgFd = TypeDefinitions.asMingleStruct( fd );

        try
        {
            MingleCodec codec = 
                context().codecFactory().expectCodec( "binary" );

            ByteBuffer bb = MingleCodecs.toByteBuffer( codec, mgFd );

            return enc.encode( bb );
        }
        catch ( Exception ex ) { throw new RuntimeException( ex ); }
    }

    final
    static
    class BuildResult
    {
        final List< JvCompilationUnit > units = Lang.newList();

        final List< CharSequence > initializerEntries = Lang.newList();

        void
        addInitializerEntry( JvPackage pkg,
                             JvTypeName... names )
        {
            initializerEntries.add(
                new StringBuilder().
                    append( pkg ).
                    append( "." ).
                    append( Strings.join( "$", names ) )
            );
        }

        void
        addInitializerEntry( JvCompilationUnit u,
                             JvTypeName nested )
        {
            addInitializerEntry( u.pkg, u.decl.name, nested );
        }
    }

    abstract
    BuildResult
    buildImpl();

    final
    BuildResult
    build( T def,
           CodegenContext ctx )
    {
        this.def = def;
        this.ctx = ctx;

        return buildImpl();
    }
}
