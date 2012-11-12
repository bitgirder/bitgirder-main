package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.MingleString;
import com.bitgirder.mingle.model.MingleBoolean;
import com.bitgirder.mingle.model.MingleSymbolMapAccessor;
import com.bitgirder.mingle.model.MingleSymbolMap;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleNamespace;
import com.bitgirder.mingle.model.MingleListIterator;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.MingleTypeName;
import com.bitgirder.mingle.model.MingleValidation;
import com.bitgirder.mingle.model.MingleListIterator;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleValueExchanger;
import com.bitgirder.mingle.model.QualifiedTypeName;
import com.bitgirder.mingle.model.AtomicTypeReference;

import java.util.Map;
import java.util.List;

import java.util.regex.Pattern;
import java.util.regex.PatternSyntaxException;

final
class CodegenControl
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static MingleNamespace NS =
        MingleNamespace.create( "mingle:codegen:java@v1" );

    private final static MingleTypeReference TYPE = 
        MingleTypeReference.create( NS + "/CodegenControl" );

    private final static MingleIdentifier ID_GENERATOR_PARAMETERS =
        MingleIdentifier.create( "generator-parameters" );

    private final static MingleIdentifier ID_TYPE_MASK =
        MingleIdentifier.create( "type-mask" );

    private final static MingleIdentifier ID_SERVICE_BASE_CLASS =
        MingleIdentifier.create( "service-base-class" );

    private final static MingleTypeReference TYPE_SVC_GEN_PARAMS =
        MingleTypeReference.create( NS + "/ServiceGeneratorParameters" );

    private final static MingleTypeReference TYPE_STRUCT_GEN_PARAMS =
        MingleTypeReference.create( NS + "/StructGeneratorParameters" );

    private final static MingleIdentifier ID_FLD_PARAMS =
        MingleIdentifier.create( "fieldParams" );

    private final static MingleTypeReference TYPE_FLD_GEN_PARAMS =
        MingleTypeReference.create( NS + "/FieldGeneratorParameters" );

    private final static MingleTypeReference TYPE_OP_PARAMS =
        MingleTypeReference.create( NS + "/OperationGeneratorParameters" );

    private final static MingleIdentifier ID_VISIBILITY = 
        MingleIdentifier.create( "visibility" );

    private final static MingleValueExchanger EXCH_VIS =
        MingleModels.createExchanger(
            (AtomicTypeReference) 
                MingleTypeReference.create( NS + "/JvVisibility" ),
            JvVisibility.class
        );

    private final static MingleIdentifier ID_TYPE_MAPPINGS =
        MingleIdentifier.create( "type-mappings" );

    private final static MingleTypeReference TYPE_TYPE_MAPPING =
        MingleTypeReference.create( NS + "/TypeMapping" );

    private final static MingleIdentifier ID_MINGLE =
        MingleIdentifier.create( "mingle" );

    private final static MingleIdentifier ID_JAVA =
        MingleIdentifier.create( "java" );

    final Map< MingleNamespace, String > nsMap;
    final Map< QualifiedTypeName, JvQname > typeMap;
    final List< Pattern > excludes;
    final List< TypeMaskedGeneratorParameters > gpList;

    private
    CodegenControl( Builder b )
    {
        this.nsMap = Lang.unmodifiableCopy( b.nsMap );
        this.typeMap = Lang.unmodifiableCopy( b.typeMap );
        this.excludes = Lang.unmodifiableCopy( b.excludes );
        this.gpList = Lang.unmodifiableCopy( b.gpList );
    }

    boolean
    excludes( QualifiedTypeName qn )
    {
        inputs.notNull( qn, "qn" );

        CharSequence str = qn.getExternalForm();

        for ( Pattern p : excludes ) 
        {
            if ( p.matcher( str ).matches() ) return true;
        }

        return false;
    }

    // Finds the last in load order which matches typ and is an instanceof cls,
    // or null if none matched
    < P extends TypeMaskedGeneratorParameters >
    P
    getGeneratorParameters( MingleTypeReference typ,
                            Class< P > cls )
    {
        inputs.notNull( typ, "typ" );
        inputs.notNull( cls, "cls" );

        P res = null;

        for ( TypeMaskedGeneratorParameters gp : gpList )
        {
            if ( cls.isInstance( gp ) && gp.matches( typ ) ) 
            {
                res = cls.cast( gp );
            }
        }

        return res;
    }

    private
    final
    static
    class Builder
    {
        private final Map< MingleNamespace, String > nsMap = Lang.newMap();
        private final Map< QualifiedTypeName, JvQname > typeMap = Lang.newMap();
        private final List< Pattern > excludes = Lang.newList();

        private final List< TypeMaskedGeneratorParameters > gpList = 
            Lang.newList();
    }

    private
    static
    MingleSymbolMapAccessor
    getAcc( MingleStruct ms )
    {
        ObjectPath< MingleIdentifier > path = ObjectPath.getRoot();

        return MingleModels.expectStruct( ms, path, TYPE );
    }

    private
    static
    void
    buildNsMap( Builder b,
                MingleSymbolMapAccessor acc )
    {
        MingleListIterator li = 
            acc.getMingleListIterator( "namespace-mappings" );

        while ( li != null && li.hasNext() )
        {
            MingleSymbolMapAccessor m = li.nextMingleSymbolMapAccessor(); 

            b.nsMap.put( 
                MingleNamespace.create( m.expectMingleString( "mingle" ) ),
                m.expectString( "java" )
            );
        }
    }

    private
    static
    void
    addTypeMapping( Builder b,
                    MingleSymbolMapAccessor acc )
    {
        QualifiedTypeName qn = 
            QualifiedTypeName.create( acc.expectMingleString( ID_MINGLE ) );

        JvQname jvQn = JvQname.parse( acc.expectString( ID_JAVA ) );

        Lang.putUnique( b.typeMap, qn, jvQn );
    }

    private
    static
    void
    buildTypeMapping( Builder b,
                      MingleSymbolMapAccessor acc )
    {
        MingleListIterator li = acc.getMingleListIterator( ID_TYPE_MAPPINGS );

        while ( li != null && li.hasNext() )
        {
            MingleSymbolMapAccessor mapAcc =
                expectStructureFields( li, TYPE_TYPE_MAPPING );

            addTypeMapping( b, mapAcc );
        }
    }

    private
    static
    void
    addExcludes( Builder b,
                 MingleSymbolMapAccessor acc )
    {
        MingleListIterator li = acc.getMingleListIterator( "excludes" );

        while ( li != null && li.hasNext() )
        {
            try
            {
                b.excludes.add( 
                    Pattern.compile( li.nextMingleString().toString() ) );
            }
            catch ( PatternSyntaxException pse )
            {
                throw MingleValidation.createFail(
                    li.getPath(), "Invalid exclude pattern:", pse.getMessage()
                );
            }
        }
    }

    private
    static
    void
    setTypeMask( TypeMaskedGeneratorParameters.Builder< ? > b,
                 MingleSymbolMapAccessor acc )
    {
        try
        {
            b.setTypeMask( 
                Pattern.compile( acc.expectString( ID_TYPE_MASK ) ) );
        }
        catch ( PatternSyntaxException pse )
        {
            throw
                MingleValidation.createFail(
                    acc.getPath().descend( ID_TYPE_MASK ),
                    "Invalid mask regex:", pse.getMessage()
                );
        }
    }

    private
    static
    < B extends TypeMaskedGeneratorParameters.Builder< B > >
    B
    init( B b,
          MingleSymbolMapAccessor acc )
    {
        setTypeMask( b, acc );

        return b;
    }

    private
    static
    MingleSymbolMapAccessor
    expectStructureFields( MingleListIterator li,
                           MingleTypeReference typ )
    {
        MingleValue mv = li.next();

        MingleSymbolMap msm = 
            MingleModels.asStructureFields( mv, typ, li.getPath() );

        return MingleSymbolMapAccessor.create( msm, li.getPath() );
    }

    private
    static
    FieldGeneratorParameters
    asFieldGeneratorParameters( MingleSymbolMapAccessor acc )
    {
        FieldGeneratorParameters.Builder b =
            new FieldGeneratorParameters.Builder();

        b.setName( MingleIdentifier.create( acc.expectString( "name" ) ) );

        MingleBoolean useOpaque = acc.getMingleBoolean( "useOpaqueJavaType" );
        if ( useOpaque != null )
        {
            b.setUseOpaqueJavaType( useOpaque.booleanValue() );
        }

        return b.build();
    }

    private
    static
    List< FieldGeneratorParameters >
    getFieldParameters( MingleSymbolMapAccessor acc )
    {
        List< FieldGeneratorParameters > res = Lang.newList();

        MingleListIterator li = acc.getMingleListIterator( ID_FLD_PARAMS );

        while ( li != null && li.hasNext() )
        {
            res.add( 
                asFieldGeneratorParameters(
                    expectStructureFields( li, TYPE_FLD_GEN_PARAMS ) ) );
        }

        return res;
    }

    private
    static
    JvVisibility
    getVisibility( MingleSymbolMapAccessor acc )
    {
        MingleValue mv = acc.getMingleValue( ID_VISIBILITY );

        if ( mv == null ) return null;
        else
        {
            return 
                (JvVisibility) EXCH_VIS.asJavaValue( 
                    mv, acc.getPath().descend( ID_VISIBILITY ) 
                );
        }
    }

    private
    static
    OperationGeneratorParameters
    buildOpParams( MingleSymbolMapAccessor acc )
    {
        OperationGeneratorParameters.Builder b =
            new OperationGeneratorParameters.Builder();

        b.setName(
            MingleIdentifier.create( acc.expectMingleString( "name" ) ) );

        MingleBoolean ort = acc.getMingleBoolean( "useOpaqueJavaReturnType" );
        if ( ort != null ) b.setUseOpaqueJavaReturnType( ort.booleanValue() );

        b.setFieldParams( getFieldParameters( acc ) );

        return b.build();
    }

    private
    static
    void
    addOpParameters( ServiceGeneratorParameters.Builder b,
                     MingleSymbolMapAccessor acc )
    {
        MingleListIterator li = acc.getMingleListIterator( "opParams" );

        List< OperationGeneratorParameters > l = Lang.newList();

        while ( li != null && li.hasNext() )
        {
            l.add( 
                buildOpParams( expectStructureFields( li, TYPE_OP_PARAMS ) ) );
        }

        b.setOpParams( l );
    }

    private
    static
    ServiceGeneratorParameters
    asServiceGeneratorParameters( MingleSymbolMapAccessor acc )
    {
        ServiceGeneratorParameters.Builder b = 
            init( new ServiceGeneratorParameters.Builder(), acc );

        MingleString typ = acc.getMingleString( ID_SERVICE_BASE_CLASS );
        if ( typ != null ) b.setServiceBaseClass( JvQname.parse( typ ) );

        addOpParameters( b, acc );

        return b.build();
    }

    private
    static
    StructGeneratorParameters
    asStructGeneratorParameters( MingleSymbolMapAccessor acc )
    {
        StructGeneratorParameters.Builder b =
            init( new StructGeneratorParameters.Builder(), acc );
        
        b.setFieldParameters( getFieldParameters( acc ) );

        JvVisibility vis = getVisibility( acc );
        if ( vis != null ) b.setVisibility( vis );

        return b.build();
    }

    private
    static
    TypeMaskedGeneratorParameters
    asMaskedGeneratorParameters( MingleStruct ms,
                                 ObjectPath< MingleIdentifier > path )
    {
        MingleSymbolMapAccessor acc = 
            MingleSymbolMapAccessor.create( ms, path );
        
        if ( ms.getType().equals( TYPE_SVC_GEN_PARAMS ) )
        {
            return asServiceGeneratorParameters( acc );
        }
        else if ( ms.getType().equals( TYPE_STRUCT_GEN_PARAMS ) )
        {
            return asStructGeneratorParameters( acc );
        }
        else 
        {
            throw 
                MingleValidation.createFail( 
                    path, "Unrecognized generator parameters" );
        }
    }

    private
    static
    void
    addGeneratorParameters( Builder b,
                            MingleSymbolMapAccessor acc )
    {
        MingleListIterator ml = 
            acc.getMingleListIterator( ID_GENERATOR_PARAMETERS );

        while ( ml != null && ml.hasNext() )
        {
            TypeMaskedGeneratorParameters gp = 
                asMaskedGeneratorParameters( 
                    ml.nextMingleStruct(), ml.getPath() );
            
            b.gpList.add( gp );
        }
    }

    // merges a (possibly empty) list of CodegenControl structs into a single
    // instance
    static
    CodegenControl
    forStructs( List< MingleStruct > l )
    {
        state.noneNull( l, "l" );

        Builder b = new Builder();

        for ( MingleStruct ms : l )
        {
            MingleSymbolMapAccessor acc = getAcc( ms );
            buildNsMap( b, acc );
            buildTypeMapping( b, acc );
            addExcludes( b, acc );
            addGeneratorParameters( b, acc );
        }

        return new CodegenControl( b );
    }
}
