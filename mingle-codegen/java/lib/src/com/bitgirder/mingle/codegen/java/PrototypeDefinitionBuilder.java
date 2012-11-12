package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.model.PrototypeDefinition;
import com.bitgirder.mingle.model.OperationSignature;
import com.bitgirder.mingle.model.FieldDefinition;
import com.bitgirder.mingle.model.MingleIdentifier;

import java.util.List;
import java.util.Map;

final
class PrototypeDefinitionBuilder
extends TypeDefinitionBuilder< PrototypeDefinition >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static JvId JV_ID_CREATE = new JvId( "create" );

    private final static JvTypeName JV_TYPE_NAME_INPUT =
        new JvTypeName( "Input" );

    private final static JvTypeName JV_TYPE_NAME_BUILDER =
        new JvTypeName( "Builder" );

    private final static JvId JV_ID_BUILD = new JvId( "build" );

    private final JvClass cls = new JvClass();
    private final JvIdMapper idMap = new JvIdMapper( cls );

    private final List< JvField > jvFields = Lang.newList();
    private final Map< MingleIdentifier, JvId > fldDeflIds = Lang.newMap();

    private OperationSignature sig() { return typeDef().getSignature(); }

    private
    void
    init()
    {
        // make default fields non-final, non-static, private
        for ( FieldDefinition fd : sig().getFieldSet() )
        {
            JvField f = new JvField();

            f.mods.vis = JvVisibility.PRIVATE;
            f.type = jvTypeOf( fd.getType() );
            f.name = asJvId( fd.getName() );
            f.mgField = fd;

            jvFields.add( f );

            addJvConstField( fd, f.type, fldDeflIds, cls );
        }
    }

    private
    void
    addValidateStatement( JvMethod m,
                          JvField f )
    {
        m.body.add( new JvStatement( expValidateFieldValue( f, idMap ) ) );
    }

    private
    void
    addInputFields( JvClass inptCls )
    {
        for ( JvField f : jvFields )
        {
            f = f.copyOf();
            f.mods.vis = JvVisibility.PUBLIC;
            f.mods.isFinal = true;

            inptCls.fields.add( f );
        }
    }

    private
    void
    addInputConstructor( JvClass inptCls )
    {
        JvConstructor c = new JvConstructor();
        c.vis = JvVisibility.PRIVATE;

        for ( JvField f : jvFields )
        {
            c.params.add( JvParam.forField( f ) );

            c.body.add(
                new JvStatement(
                    new JvAssign( new JvAccess( JvId.THIS, f.name ), f.name ) 
                )
            );
        }

        inptCls.constructors.add( c );
    }

    private
    JvMethod
    createSetter( JvType retType,
                  JvField f )
    {
        JvMethod m = new JvMethod();
        m.mods.vis = JvVisibility.PUBLIC;
        m.retType = retType;
        m.name = setterNameFor( f );

        m.params.add( JvParam.forField( f ) );

        m.body.add( 
            new JvStatement(
                new JvAssign( new JvAccess( JvId.THIS, f.name ), f.name ) ) );

        m.body.add( new JvReturn( JvId.THIS ) );

        return m;
    }

    private
    JvMethod
    createInputBuilderBuild( JvClass inptCls )
    {
        JvMethod m = new JvMethod();
        m.mods.vis = JvVisibility.PUBLIC;
        m.retType = inptCls.name;
        m.name = JV_ID_BUILD;

        JvInstantiate inst = new JvInstantiate();
        inst.target = m.retType;

        for ( JvField f : jvFields ) 
        {
            inst.params.add( f.name );
            addValidateStatement( m, f );
        }

        m.body.add( new JvReturn( inst ) );

        return m;
    }

    private
    void
    addInputBuilder( JvClass inptCls )
    {
        JvClass bldrCls = new JvClass();
        bldrCls.mods.vis = JvVisibility.PUBLIC;
        bldrCls.mods.isStatic = true;
        bldrCls.name = JV_TYPE_NAME_BUILDER;

        for ( JvField f : jvFields )
        {
            f = f.copyOf();
            f.mods.isFinal = false;
            f.assign = fldDeflIds.get( f.mgField.getName() ); // maybe null
            bldrCls.fields.add( f );
            bldrCls.methods.add( createSetter( bldrCls.name, f ) );
        }

        bldrCls.methods.add( createInputBuilderBuild( inptCls ) );

        inptCls.nestedTypes.add( bldrCls );
    }

    private
    void
    addInputCreate( JvClass inptCls )
    {
        JvMethod m = new JvMethod();
        m.mods.vis = JvVisibility.PUBLIC;
        m.mods.isStatic = true;
        m.retType = inptCls.name;
        m.name = JV_ID_CREATE;
        
        JvInstantiate c = new JvInstantiate();
        c.target = inptCls.name;

        for ( JvField f : jvFields )
        {
            m.params.add( JvParam.forField( f ) );
            addValidateStatement( m, f );
            c.params.add( f.name );
        }

        m.body.add( new JvReturn( c ) );
        inptCls.methods.add( m );
    }

    private
    JvClass
    createInputClass()
    {
        JvClass inptCls = new JvClass();
        inptCls.mods.vis = JvVisibility.PUBLIC;
        inptCls.mods.isStatic = true;
        inptCls.name = JV_TYPE_NAME_INPUT;

        addInputFields( inptCls );
        addInputConstructor( inptCls );
        addInputBuilder( inptCls );
        addInputCreate( inptCls );

        return inptCls;
    }

    private
    void
    buildProtoClass()
    {
        cls.mods.vis = JvVisibility.PUBLIC;
        cls.name = jvTypeNameOf( typeDef() );

        cls.nestedTypes.add( createInputClass() );
    }

    BuildResult
    buildImpl()
    {
        init();
        buildProtoClass();

        BuildResult res = new BuildResult();

        JvCompilationUnit u = new JvCompilationUnit();
        u.pkg = jvPackageOf( typeDef() );
        u.decl = cls;
        
        res.units.add( u );

        return res;
    }
}
