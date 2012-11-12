package com.bitgirder.mingle.codegen.java;

import static com.bitgirder.mingle.codegen.java.CodegenConstants.*;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.mingle.model.TypeDefinition;
import com.bitgirder.mingle.model.AliasedTypeDefinition;
import com.bitgirder.mingle.model.EnumDefinition;
import com.bitgirder.mingle.model.StructureDefinition;
import com.bitgirder.mingle.model.ServiceDefinition;
import com.bitgirder.mingle.model.PrototypeDefinition;
import com.bitgirder.mingle.model.MingleStruct;
import com.bitgirder.mingle.model.QualifiedTypeName;

import com.bitgirder.mingle.bind.MingleBinders;

import com.bitgirder.mingle.codegen.AbstractMingleCodeGenerator;

import java.util.List;

final
class JvCodeGenerator
extends AbstractMingleCodeGenerator
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private
    CodegenControl
    createCodegenControl()
    {
        return CodegenControl.forStructs( context().controlObjects() );
    }

    private JvSourceGenerator sourceGen() { return new JvSourceGenerator(); }

    private
    CharSequence
    sourceFileFor( JvCompilationUnit u )
    {
        return
            u.pkg.toString().replace( '.', '/' ) +
            "/" +
            u.decl.name +
            ".java";
    }

    private
    void
    writeAsync( JvCompilationUnit u )
        throws Exception
    {
        CharSequence src = sourceGen().asSource( u );
        context().writeAsync( sourceFileFor( u ), src );
    }

    private
    TypeDefinitionBuilder.BuildResult
    buildType( TypeDefinition td,
               CodegenContext cgCtx )
    {
        if ( td instanceof StructureDefinition )
        {
            return
                new StructureDefinitionBuilder().
                    build( (StructureDefinition) td, cgCtx );
        }
        else if ( td instanceof EnumDefinition )
        {
            return
                new EnumDefinitionBuilder().
                    build( (EnumDefinition) td, cgCtx );
        }
        else if ( td instanceof ServiceDefinition )
        {
            return
                new ServiceDefinitionBuilder().
                    build( (ServiceDefinition) td, cgCtx );
        }
        else if ( td instanceof PrototypeDefinition ) 
        {
            return
                new PrototypeDefinitionBuilder().
                    build( (PrototypeDefinition) td, cgCtx );
        }
        else if ( td instanceof AliasedTypeDefinition ) return null;
        else throw state.createFail( "Unexpected def:", td );
    }

    // returns possibly empty list of class names that should have entries in
    // the generated binder initializer
    private
    List< CharSequence >
    generateType( TypeDefinition td,
                  CodegenContext cgCtx )
        throws Exception
    {
        TypeDefinitionBuilder.BuildResult br = buildType( td, cgCtx );
 
        if ( br == null ) return Lang.emptyList();
        else
        {
            for ( JvCompilationUnit u : br.units ) writeAsync( u );
            return br.initializerEntries;
        }
    }

    private
    void
    writeBinderInitializer( List< CharSequence > nms )
        throws Exception
    {
        if ( ! nms.isEmpty() )
        {
            context().writeAsync(
                MingleBinders.RSRC_NAME,
                Strings.join( "\n", nms ) + "\n"
            );
        }
    }

    protected
    void
    startGen()
        throws Exception
    {
        CodegenContext cgCtx = 
            new CodegenContext( createCodegenControl(), context() );

        List< CharSequence > built = Lang.newList();

        for ( QualifiedTypeName qn : context().getTargets() )
        {
            TypeDefinition td = context().runtime().getTypes().expectType( qn );
            
            if ( ! cgCtx.codegenControl().excludes( qn ) )
            {
                built.addAll( generateType( td, cgCtx ) );
            }
        }

        writeBinderInitializer( built );

        context().complete();
    }
}
