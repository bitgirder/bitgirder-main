package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ImmutableListPath;

import com.bitgirder.parser.SyntaxException;

import java.util.List;

// Note about test coverage: we currently don't have test coverage in this
// package, although the class is defined here. Test coverage does come though
// in other places, such as in compiler tests which exercise this class by
// roundtripping compiled output through its mingle version before passing it to
// assertion code
public
final
class TypeDefinitions
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static MingleIdentifier ID_NAME =
        MingleIdentifier.create( "name" );
    
    private final static MingleIdentifier ID_SUPER_TYPE_REF =
        MingleIdentifier.create( "super-type-reference" );

    private final static MingleIdentifier ID_CONSTRUCTORS =
        MingleIdentifier.create( "constructors" );

    private final static MingleIdentifier ID_SIGNATURE =
        MingleIdentifier.create( "signature" );
        
    private final static MingleIdentifier ID_FIELDS =
        MingleIdentifier.create( "fields" );

    private final static MingleIdentifier ID_TYPES =
        MingleIdentifier.create( "types" );

    private final static MingleIdentifier ID_RETURN_TYPE =
        MingleIdentifier.create( "return-type" );

    private final static MingleIdentifier ID_THROWN =
        MingleIdentifier.create( "thrown" );

    private final static MingleIdentifier ID_AUTH_INPUT_TYPE =
        MingleIdentifier.create( "auth-input-type" );
    
    private final static MingleIdentifier ID_AUTHENTICATION =
        MingleIdentifier.create( "authentication" );

    private final static MingleIdentifier ID_TYPE =
        MingleIdentifier.create( "type" );

    private final static MingleIdentifier ID_DEFAULT =
        MingleIdentifier.create( "default" );

    private final static MingleIdentifier ID_OPERATIONS = 
        MingleIdentifier.create( "operations" );
    
    private final static MingleIdentifier ID_SECURITY =
        MingleIdentifier.create( "security" );

    private final static MingleIdentifier ID_VALUES =
        MingleIdentifier.create( "values" );

    private final static MingleIdentifier ID_ALIASED_TYPE =
        MingleIdentifier.create( "aliased-type" );

    private final static MingleNamespace NS = 
        MingleNamespace.create( "mingle:model@v1" );
    
    private final static ObjectPath< String > JPATH_ROOT = ObjectPath.getRoot();

    private final static MingleValueExchanger SVC_DEF_EXCH =
        new ServiceDefinitionExchanger();

    private final static MingleValueExchanger PROTO_DEF_EXCH =
        new PrototypeDefinitionExchanger();

    private final static MingleValueExchanger OP_SIG_EXCH =
        new OperationSignatureExchanger();

    private final static MingleValueExchanger OP_DEF_EXCH =
        new OperationDefinitionExchanger();

    private final static MingleValueExchanger EXCPT_DEF_EXCH =
        new StructureDefinitionExchanger< ExceptionDefinition >(
            ExceptionDefinition.class )
        {
            ExceptionDefinition.Builder defBuilder() {
                return new ExceptionDefinition.Builder();
            }
        };

    private final static MingleValueExchanger STRUCT_DEF_EXCH =
        new StructureDefinitionExchanger< StructDefinition >(
            StructDefinition.class )
        {
            StructDefinition.Builder defBuilder() {
                return new StructDefinition.Builder();
            }
        };

    private final static MingleValueExchanger ENUM_DEF_EXCH =
        new EnumDefinitionExchanger();

    private final static MingleValueExchanger FLD_DEF_EXCH =
        new FieldDefinitionExchanger();

    private final static MingleValueExchanger ALIAS_TYP_DEF_EXCH =
        new AliasedTypeDefinitionExchanger();

    private final static MingleValueExchanger TYPE_DEF_COLL_EXCH =
        new TypeDefinitionCollectionExchanger();

    private TypeDefinitions() {}

    private
    static
    AtomicTypeReference
    createTypeReference( MingleTypeName nm )
    {
        state.notNull( nm, "nm" );

        return AtomicTypeReference.create( nm.resolveIn( NS ) );
    }

    private
    static
    AtomicTypeReference
    createTypeReference( CharSequence nm )
    {
        state.notNull( nm, "nm" );
        return createTypeReference( MingleTypeName.create( nm ) );
    }

    private
    static
    MingleValidationException
    createRethrow( SyntaxException se,
                   ObjectPath< MingleIdentifier > path )
    {
        return new MingleValidationException( se.getMessage(), path );
    }

    private
    static
    MingleValidationException
    createRethrow( SyntaxException se,  
                   MingleSymbolMapAccessor acc,
                   MingleIdentifier key )
    {
        return createRethrow( se, acc.getPath().descend( key ) );
    }

    private
    static
    QualifiedTypeName
    getQualifiedTypeName( MingleSymbolMapAccessor acc,
                          MingleIdentifier key )
    {
        MingleString s = acc.getMingleString( key );

        if ( s == null ) return null;
        else
        {
            try { return QualifiedTypeName.parse( s ); }
            catch ( SyntaxException se )
            {
                throw createRethrow( se, acc, key );
            }
        }
    }

    private
    static
    MingleIdentifier
    expectIdentifier( MingleSymbolMapAccessor acc,
                      MingleIdentifier key )
    {
        state.notNull( acc, "acc" );
        state.notNull( key, "key" );

        try { return MingleIdentifier.parse( acc.expectMingleString( key ) ); }
        catch ( SyntaxException se ) { throw createRethrow( se, acc, key ); }
    }

    private
    static
    List< MingleIdentifier >
    getIdentifiers( MingleSymbolMapAccessor acc,
                    MingleIdentifier key )
    {
        state.notNull( acc, "acc" );
        state.notNull( key, "key" );

        List< MingleIdentifier > res = Lang.newList();
        MingleListIterator li = acc.getMingleListIterator( key );

        while ( li != null && li.hasNext() )
        {
            try { res.add( MingleIdentifier.parse( li.nextMingleString() ) ); }
            catch ( SyntaxException se )
            {
                throw createRethrow( se, li.getPath() );
            }
        }

        return res;
    }

    private
    static
    List< MingleIdentifier >
    expectIdentifiers( MingleSymbolMapAccessor acc,
                       MingleIdentifier key )
    {
        List< MingleIdentifier > res = getIdentifiers( acc, key );

        MingleValidation.isFalse( 
            res.isEmpty(), acc.getPath().descend( key ), "list is empty" );
        
        return res;
    }

    private
    static
    MingleTypeReference
    asTypeRef( MingleString str,
               ObjectPath< MingleIdentifier > path )
    {
        try { return str == null ? null : MingleTypeReference.parse( str ); }
        catch ( SyntaxException se ) { throw createRethrow( se, path ); }
    }

    private
    static
    MingleTypeReference
    asTypeRef( MingleString str,
               MingleSymbolMapAccessor acc,
               MingleIdentifier key )
    {
        return asTypeRef( str, acc.getPath().descend( key ) );
    }

    private
    static
    MingleTypeReference
    expectTypeReference( MingleSymbolMapAccessor acc,
                         MingleIdentifier key )
    {
        state.notNull( acc, "acc" );
        state.notNull( key, "key" );

        return asTypeRef( acc.expectMingleString( key ), acc, key );
    }

    private
    static
    MingleTypeReference
    getTypeReference( MingleSymbolMapAccessor acc,
                      MingleIdentifier key )
    {
        state.notNull( acc, "acc" );
        state.notNull( key, "key" );

        return asTypeRef( acc.getMingleString( key ), acc, key );
    }

    private
    static
    List< MingleTypeReference >
    getTypeReferences( MingleSymbolMapAccessor acc,
                       MingleIdentifier key )
    {
        state.notNull( acc, "acc" );
        state.notNull( key, "key" );

        List< MingleTypeReference > res = Lang.newList();

        MingleListIterator li = acc.getMingleListIterator( key );
        while ( li != null && li.hasNext() )
        {
            MingleString str = li.nextMingleString();
            MingleTypeReference ref = asTypeRef( str, li.getPath() );
            res.add( ref );
        }

        return res;
    }

    private
    static
    FieldSet
    getFields( MingleSymbolMapAccessor acc,
               MingleIdentifier key )
    {
        state.notNull( acc, "acc" );
        state.notNull( key, "key" );

        List< FieldDefinition > flds = Lang.newList();

        MingleListIterator li = acc.getMingleListIterator( key );
        while ( li != null && li.hasNext() )
        {
            MingleStruct ms = li.nextMingleStruct();
            flds.add( asFieldDefinition( ms, li.getPath() ) );
        }

        return FieldSet.create( flds );
    }

    private
    static
    QualifiedTypeName
    expectQname( MingleSymbolMapAccessor acc,
                 MingleIdentifier key )
    {
        state.notNull( acc, "acc" );
        state.notNull( key, "key" );

        try { return QualifiedTypeName.parse( acc.expectMingleString( key ) ); }
        catch ( SyntaxException se ) { throw createRethrow( se, acc, key ); }
    }

    private
    static
    MingleList
    asMingleFieldSet( FieldSet fs,
                      ObjectPath< String > path )
    {
        state.notNull( fs, "fs" );

        MingleList.Builder b = new MingleList.Builder();

        ImmutableListPath< String > lp = path.startImmutableList();

        for ( FieldDefinition fd : fs.getFields() )
        {
            b.add( asMingleStruct( fd, lp.next() ) );
        }

        return b.build();
    }

    private
    static
    MingleList
    asMingleOpDefs( List< OperationDefinition > opDefs,
                    ObjectPath< String > path )
    {
        state.notNull( opDefs, "opDefs" );

        MingleList.Builder b = new MingleList.Builder();

        ImmutableListPath< String > lp = path.startImmutableList();

        for ( OperationDefinition opDef : opDefs )
        {
            b.add( OP_DEF_EXCH.asMingleValue( opDef, lp.next() ) );
        }

        return b.build();
    }

    private
    static
    MingleList
    asMingleTypeRefs( List< MingleTypeReference > types )
    {
        state.notNull( types, "types" );

        MingleList.Builder b = new MingleList.Builder();

        for ( MingleTypeReference type : types )
        {
            b.add( MingleModels.asMingleString( type.getExternalForm() ) );
        }

        return b.build();
    }

    private
    static
    MingleValueExchanger
    exchangerFor( TypeDefinition td )
    {
        if ( td instanceof AliasedTypeDefinition ) return ALIAS_TYP_DEF_EXCH;
        else if ( td instanceof ServiceDefinition ) return SVC_DEF_EXCH;
        else if ( td instanceof PrototypeDefinition ) return PROTO_DEF_EXCH;
        else if ( td instanceof ExceptionDefinition ) return EXCPT_DEF_EXCH;
        else if ( td instanceof StructDefinition ) return STRUCT_DEF_EXCH;
        else if ( td instanceof EnumDefinition ) return ENUM_DEF_EXCH;
        else throw state.createFail( "Unhandled def:", td );
    }

    private
    static
    MingleStruct
    asMingleStruct( TypeDefinition td,
                    ObjectPath< String > path )
    {
        MingleValueExchanger exch = exchangerFor( td );

        MingleValue res = exch.asMingleValue( td, path );

        return state.cast( MingleStruct.class, res ); // also asserts non-null
    }

    public
    static
    MingleList
    asMingleList( Iterable< ? extends TypeDefinition > types,
                  ObjectPath< String > path )
    {
        inputs.noneNull( types, "types" );
        inputs.notNull( path, "path" );

        MingleList.Builder b = new MingleList.Builder();

        ImmutableListPath< String > lp = path.startImmutableList();

        for ( TypeDefinition td : types ) 
        {
            b.add( asMingleStruct( td, lp.next() ) );
        }

        return b.build();
    }

    public
    static
    MingleList
    asMingleList( Iterable< ? extends TypeDefinition > types )
    {
        return asMingleList( types, JPATH_ROOT );
    }

    private
    static
    String
    accessTypeDefName( MingleStruct ms,
                       ObjectPath< MingleIdentifier > p )
    {
        try
        {
            QualifiedTypeName qn =
                (QualifiedTypeName)
                    ( (AtomicTypeReference) ms.getType() ).getName();
            
            List< MingleTypeName > nm = qn.getName();
            state.isTrue( NS.equals( qn.getNamespace() ), "Invalid ns:", qn );
            state.isTrue( nm.size() == 1, "Unhandled type:", qn );

            return nm.get( 0 ).getExternalForm().toString();
        }
        catch ( Throwable th )
        {
            throw new MingleValidationException( th.getMessage(), p, th );
        }
    }

    private
    static
    MingleValueExchanger
    exchangerFor( String typNm,
                  ObjectPath< MingleIdentifier > p )
    {
        if ( typNm.equals( "AliasedTypeDefinition" ) ) 
        {
            return ALIAS_TYP_DEF_EXCH;
        }
        else if ( typNm.equals( "ServiceDefinition" ) ) return SVC_DEF_EXCH;
        else if ( typNm.equals( "PrototypeDefinition" ) ) return PROTO_DEF_EXCH;
        else if ( typNm.equals( "ExceptionDefinition" ) ) return EXCPT_DEF_EXCH;
        else if ( typNm.equals( "StructDefinition" ) ) return STRUCT_DEF_EXCH;
        else if ( typNm.equals( "EnumDefinition" ) ) return ENUM_DEF_EXCH;
        else throw MingleValidation.createFail( p, "Unhandled type:", typNm );
    }

    private
    static
    TypeDefinition
    asTypeDefinition( MingleStruct ms,
                      ObjectPath< MingleIdentifier > p )
    {
        String typNm = accessTypeDefName( ms, p );
        MingleValueExchanger exch = exchangerFor( typNm, p );

        return (TypeDefinition) exch.asJavaValue( ms, p );
    }

    public
    static
    List< TypeDefinition >
    asTypeDefinitions( MingleList ml,
                       ObjectPath< MingleIdentifier > path )
    {
        inputs.notNull( ml, "ml" );
        inputs.notNull( path, "path" );

        List< TypeDefinition > res = Lang.newList();

        ImmutableListPath< MingleIdentifier > p = path.startImmutableList();

        for ( MingleValue mv : ml )
        {
            MingleStruct ms = (MingleStruct)
                MingleModels.   
                    asMingleInstance( 
                        MingleModels.TYPE_REF_MINGLE_STRUCT, mv, p );
            
//            res.add( TypeDefinition.asTypeDefinition( ms, p ) );
            res.add( asTypeDefinition( ms, p ) );
            p = p.next();
        }

        return res;
    }

    public
    static
    List< TypeDefinition >
    asTypeDefinitions( MingleList ml )
    {
        return 
            asTypeDefinitions( ml, ObjectPath.< MingleIdentifier >getRoot() );
    }

    public
    static
    MingleStruct
    asMingleStruct( FieldDefinition fd,
                    ObjectPath< String > path )
    {
        inputs.notNull( fd, "fd" );
        inputs.notNull( path, "path" );

        return (MingleStruct) FLD_DEF_EXCH.asMingleValue( fd, path );
    }

    public
    static
    MingleStruct
    asMingleStruct( FieldDefinition fd )
    {
        return asMingleStruct( fd, JPATH_ROOT );
    }

    public
    static
    MingleStruct
    asMingleStruct( TypeDefinitionCollection coll,
                    ObjectPath< String > path )
    {
        inputs.notNull( coll, "coll" );
        inputs.notNull( path, "path" );

        return (MingleStruct) TYPE_DEF_COLL_EXCH.asMingleValue( coll, path );
    }

    public
    static
    TypeDefinitionCollection
    asTypeDefinitionCollection( MingleValue mv,
                                ObjectPath< MingleIdentifier > path )
    {
        inputs.notNull( mv, "mv" );
        inputs.notNull( path, "path" );

        return (TypeDefinitionCollection) 
            TYPE_DEF_COLL_EXCH.asJavaValue( mv, path );
    }

    public
    static
    TypeDefinitionLookup
    asTypeDefinitionLookup( MingleValue mv,
                            ObjectPath< MingleIdentifier > path )
    {
        inputs.notNull( mv, "mv" );
        inputs.notNull( path, "path" );

        return 
            new TypeDefinitionLookup.Builder().
                addTypes( asTypeDefinitionCollection( mv, path ) ).
                build();
    }

    private
    static
    FieldDefinition
    asFieldDefinition( MingleValue mv,
                       ObjectPath< MingleIdentifier > path )
    {
        return (FieldDefinition) FLD_DEF_EXCH.asJavaValue( mv, path );
    }

    public
    static
    FieldDefinition
    asFieldDefinition( MingleStruct ms )
    {
        return 
            asFieldDefinition( ms, ObjectPath.< MingleIdentifier >getRoot() );
    }

    private
    static
    OperationSignature
    expectOpSig( MingleSymbolMapAccessor acc,
                 MingleIdentifier fld )
    {
        return (OperationSignature)
            OP_SIG_EXCH.asJavaValue(
                acc.expectMingleStruct( fld ),
                acc.getPath().descend( fld ) 
            );
    }

    private
    static
    abstract
    class ExchangerImpl< V >
    extends AbstractStructExchanger< V >
    {
        private
        ExchangerImpl( Class< V > cls )
        {
            super(
                AtomicTypeReference.create(
                    MingleTypeName.create( cls.getSimpleName() ).resolveIn( NS )
                ),
                cls
            );
        }
    }

    private
    final
    static
    class OperationSignatureExchanger
    extends ExchangerImpl< OperationSignature >
    {
        private
        OperationSignatureExchanger()
        {
            super( OperationSignature.class );
        }

        protected
        MingleValue
        implAsMingleValue( OperationSignature sig,
                           ObjectPath< String > path )
        {
            return 
                structBuilder().
                f().set( 
                    ID_FIELDS, 
                    asMingleFieldSet( 
                        sig.getFieldSet(), path.descend( "fields" ) )
                ).
                f().setString( 
                    ID_RETURN_TYPE, sig.getReturnType().getExternalForm() ).
                f().set( ID_THROWN, asMingleTypeRefs( sig.getThrown() ) ).
                build();
        }

        protected
        OperationSignature
        buildStruct( MingleSymbolMapAccessor acc )
        {
            return
                new OperationSignature.Builder().
                    setFields( getFields( acc, ID_FIELDS ) ).
                    setReturnType( 
                        MingleModels.expectTypeReference( acc, ID_RETURN_TYPE )
                    ).
                    setThrown( getTypeReferences( acc, ID_THROWN ) ).
                    build();
        }
    }

    private
    final
    static
    class OperationDefinitionExchanger
    extends ExchangerImpl< OperationDefinition >
    {
        private
        OperationDefinitionExchanger()
        {
            super( OperationDefinition.class );
        }

        protected
        MingleValue
        implAsMingleValue( OperationDefinition def,
                           ObjectPath< String > path )
        {
            return
                structBuilder().
                f().setString( ID_NAME, def.getName().getExternalForm() ).
                f().set( ID_SIGNATURE, 
                    OP_SIG_EXCH.asMingleValue( 
                        def.getSignature(), path.descend( "signature" )
                    )
                ).
                build();
        }

        protected
        OperationDefinition
        buildStruct( MingleSymbolMapAccessor acc )
        {
            return
                new OperationDefinition.Builder().
                    setName( expectIdentifier( acc, ID_NAME ) ).
                    setSignature( expectOpSig( acc, ID_SIGNATURE ) ).
                    build();
        }
    }

    private
    static
    abstract
    class TypeDefinitionExchanger< T extends TypeDefinition >
    extends ExchangerImpl< T >
    {
        private TypeDefinitionExchanger( Class< T > cls ) { super( cls ); }

        final
        MingleStructBuilder
        structBuilder( T def )
        {
            MingleStructBuilder res = structBuilder();

            res.f().setString( ID_NAME, def.getName().getExternalForm() );

            MingleTypeReference sprTyp = def.getSuperType();

            if ( sprTyp != null ) 
            {
                res.f().
                    setString( ID_SUPER_TYPE_REF, sprTyp.getExternalForm() );
            }

            return res;
        }

        final
        void
        init( TypeDefinition.Builder< ?, ? > b,
              MingleSymbolMapAccessor acc )
        {
            b.setName( expectQname( acc, ID_NAME ) );

            MingleTypeReference sprTyp =
                MingleModels.getTypeReference( acc, ID_SUPER_TYPE_REF );

            if ( sprTyp != null ) b.setSuperType( sprTyp );
        }
    }

    private
    final
    static
    class AliasedTypeDefinitionExchanger
    extends TypeDefinitionExchanger< AliasedTypeDefinition >
    {
        private
        AliasedTypeDefinitionExchanger()
        {
            super( AliasedTypeDefinition.class );
        }

        protected
        MingleStruct
        implAsMingleValue( AliasedTypeDefinition def,
                           ObjectPath< String > path )
        {
            return
                structBuilder( def ).
                f().setString( 
                    ID_ALIASED_TYPE, def.getAliasedType().getExternalForm() ).
                build();
        }

        protected
        AliasedTypeDefinition
        buildStruct( MingleSymbolMapAccessor acc )
        {
            AliasedTypeDefinition.Builder b = new
                AliasedTypeDefinition.Builder();

            init( b, acc );

            b.setAliasedType( expectTypeReference( acc, ID_ALIASED_TYPE ) );

            return b.build();
        }
    }
            
    private
    final
    static
    class ServiceDefinitionExchanger
    extends TypeDefinitionExchanger< ServiceDefinition >
    {
        private
        ServiceDefinitionExchanger()
        {
            super( ServiceDefinition.class );
        }

        protected
        MingleStruct
        implAsMingleValue( ServiceDefinition sd,
                           ObjectPath< String > path )
        {
            MingleStructBuilder b = structBuilder( sd );

            b.f().set( 
                ID_OPERATIONS, 
                asMingleOpDefs( 
                    sd.getOperations(), path.descend( "operations" ) )
            );
            
            QualifiedTypeName secRef = sd.getSecurity();

            if ( secRef != null )
            {
                b.f().setString( ID_SECURITY, secRef.getExternalForm() );
            }

            return b.build();
        }

        private
        List< OperationDefinition >
        getOperations( MingleSymbolMapAccessor acc )
        {
            List< OperationDefinition > res = Lang.newList();
            MingleListIterator mi = acc.getMingleListIterator( ID_OPERATIONS );
            
            while ( mi != null && mi.hasNext() )
            {
                MingleStruct ms = mi.nextMingleStruct();

                res.add( (OperationDefinition)
                    OP_DEF_EXCH.asJavaValue( ms, mi.getPath() ) );
            }

            return res;
        }

        private
        void
        setOptSecurity( ServiceDefinition.Builder b,
                        MingleSymbolMapAccessor acc )
        {
            QualifiedTypeName qn = getQualifiedTypeName( acc, ID_SECURITY );
            if ( qn != null ) b.setSecurity( qn );
        }

        protected
        ServiceDefinition
        buildStruct( MingleSymbolMapAccessor acc )
        {
            ServiceDefinition.Builder b = new ServiceDefinition.Builder();
            init( b, acc );

            b.setOperations( getOperations( acc ) );
            setOptSecurity( b, acc );

            return b.build();
        }
    }

    private
    final
    static
    class PrototypeDefinitionExchanger
    extends TypeDefinitionExchanger< PrototypeDefinition >
    {
        private
        PrototypeDefinitionExchanger()
        {
            super( PrototypeDefinition.class );
        }

        protected
        MingleStruct
        implAsMingleValue( PrototypeDefinition pd,
                           ObjectPath< String > path )
        {
            MingleValue sig =
                OP_SIG_EXCH.asMingleValue( 
                    pd.getSignature(), path.descend( "signature" )
                );

            return
                structBuilder( pd ).
                f().set( ID_SIGNATURE, sig ).
                build();
        }

        protected
        PrototypeDefinition
        buildStruct( MingleSymbolMapAccessor acc )
        {
            PrototypeDefinition.Builder b = new PrototypeDefinition.Builder();
            init( b, acc );

            b.setSignature( expectOpSig( acc, ID_SIGNATURE ) );
            return b.build();
        }
    }

    private
    static
    abstract
    class StructureDefinitionExchanger< S extends StructureDefinition >
    extends TypeDefinitionExchanger< S >
    {
        private
        StructureDefinitionExchanger( Class< S > cls )
        {
            super( cls );
        }

        protected
        final
        MingleValue
        implAsMingleValue( S def,
                           ObjectPath< String > path )
        {
            MingleStructBuilder b = structBuilder( def );

            b.f().set( 
                ID_FIELDS, 
                asMingleFieldSet( def.getFieldSet(), path.descend( "fields" ) )
            );

            b.f().set( 
                ID_CONSTRUCTORS, asMingleTypeRefs( def.getConstructors() ) );
            
            return b.build();
        }

        abstract
        StructureDefinition.Builder< S, ? >
        defBuilder();

        protected
        final
        S
        buildStruct( MingleSymbolMapAccessor acc )
        {
            StructureDefinition.Builder< S, ? > b = defBuilder();
            init( b, acc );

            b.setFields( getFields( acc, ID_FIELDS ) );
            b.setConstructors( getTypeReferences( acc, ID_CONSTRUCTORS ) );

            return b.build();
        }
    }

    private
    final
    static
    class EnumDefinitionExchanger
    extends TypeDefinitionExchanger< EnumDefinition >
    {
        private
        EnumDefinitionExchanger()
        {
            super( EnumDefinition.class );
        }

        protected
        MingleValue
        implAsMingleValue( EnumDefinition ed,
                           ObjectPath< String > path )
        {
            MingleStructBuilder b = structBuilder( ed );
    
            MingleList.Builder lb = new MingleList.Builder();
    
            for ( MingleIdentifier v : ed.getNames() )
            {
                lb.add( MingleModels.asMingleString( v.getExternalForm() ) );
            }
    
            b.f().set( ID_VALUES, lb.build() );
            
            return b.build();
        }

        protected
        EnumDefinition
        buildStruct( MingleSymbolMapAccessor acc )
        {
            EnumDefinition.Builder b = new EnumDefinition.Builder();
            init( b, acc );

            b.setValues( expectIdentifiers( acc, ID_VALUES ) );

            return b.build();
        }
    }

    private
    final
    static
    class FieldDefinitionExchanger
    extends ExchangerImpl< FieldDefinition >
    {
        private
        FieldDefinitionExchanger()
        {
            super( FieldDefinition.class );
        }

        protected
        MingleValue
        implAsMingleValue( FieldDefinition fd,
                           ObjectPath< String > path )
        {
            MingleStructBuilder b = structBuilder();
            
            b.f().setString( ID_NAME, fd.getName().getExternalForm() );
            b.f().setString( ID_TYPE, fd.getType().getExternalForm() );
    
            MingleValue defl = fd.getDefault();
            if ( defl != null ) b.f().set( ID_DEFAULT, defl );
            
            return b.build();
        }

        protected
        FieldDefinition
        buildStruct( MingleSymbolMapAccessor acc )
        {
            FieldDefinition.Builder b = new FieldDefinition.Builder();
    
            b.setName( expectIdentifier( acc, ID_NAME ) );
            b.setType( expectTypeReference( acc, ID_TYPE ) );
            
            MingleValue defl = acc.getMingleValue( ID_DEFAULT );
            if ( defl != null ) b.setDefault( defl );
    
            return b.build();
        }
    }

    private
    final
    static
    class TypeDefinitionCollectionExchanger
    extends ExchangerImpl< TypeDefinitionCollection >
    {
        private
        TypeDefinitionCollectionExchanger()
        {
            super( TypeDefinitionCollection.class );
        }

        protected
        TypeDefinitionCollection
        buildStruct( MingleSymbolMapAccessor acc )
        {
            MingleList ml = acc.expectMingleList( ID_TYPES ); 
            List< TypeDefinition > types = asTypeDefinitions( ml );

            return TypeDefinitionCollection.create( types );
        }

        protected
        MingleValue
        implAsMingleValue( TypeDefinitionCollection coll,
                           ObjectPath< String > path )
        {
            MingleList ml = asMingleList( coll.getTypes(), path ); 

            return
                structBuilder().
                f().set( ID_TYPES, ml ).
                build();
        }
    }

    public
    static
    MingleTypeReference
    expectAuthInputType( OperationSignature secSig )
    {
        inputs.notNull( secSig, "secSig" );

        FieldDefinition fd = secSig.getFieldSet().getField( ID_AUTHENTICATION );

        if ( fd == null ) 
        {
            throw state.createFail( "Signature has no auth field" );
        }
        else return fd.getType();
    }
}
