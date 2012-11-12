package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.mingle.model.TypeDefinition;
import com.bitgirder.mingle.model.QualifiedTypeName;
import com.bitgirder.mingle.model.MingleNamespace;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleIdentifierFormat;

import com.bitgirder.mingle.codec.MingleCodecFactory;

import com.bitgirder.mingle.codegen.MingleCodeGeneratorContext;

import com.bitgirder.mingle.runtime.MingleRuntime;

import java.util.Map;
import java.util.List;
import java.util.Iterator;

final
class CodegenContext
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final CodegenControl cgCtl;
    final MingleCodeGeneratorContext cgCtx;
    
    CodegenContext( CodegenControl cgCtl,
                    MingleCodeGeneratorContext cgCtx ) 
    { 
        this.cgCtl = state.notNull( cgCtl, "cgCtl" );
        this.cgCtx = state.notNull( cgCtx, "cgCtx" );
    }

    // startsWith() and length() are using admittedly inefficient string-based
    // algs for now

    final void code( Object... args ) { cgCtx.log().code( args ); }
    final MingleRuntime runtime() { return cgCtx.runtime(); }
    final MingleCodecFactory codecFactory() { return cgCtx.codecFactory(); }
    final CodegenControl codegenControl() { return cgCtl; }

    private
    boolean
    startsWith( MingleNamespace targ,
                MingleNamespace arg )
    {
        String targPartStr = Strings.join( ":", targ.getParts() ).toString();
        String argPartStr = Strings.join( ":", arg.getParts() ).toString();

        return 
            targ.getVersion().equals( arg.getVersion() ) &&
            targPartStr.startsWith( argPartStr );
    }

    private
    int
    length( MingleNamespace ns )
    {
        return ns.getParts().size();
    }

    private
    int
    length( Map.Entry< MingleNamespace, ? > e )
    {
        return length( e.getKey() );
    }

    private
    Map.Entry< MingleNamespace, String >
    getLongestPrefixMatch( MingleNamespace ns )
    {
        Map.Entry< MingleNamespace, String > res = null;

        for ( Map.Entry< MingleNamespace, String > e : cgCtl.nsMap.entrySet() )
        {
            if ( startsWith( ns, e.getKey() ) )
            {
                if ( res == null || length( res ) < length( e ) ) res = e;
            }
        }

        return res;
    }

    private
    CharSequence
    fmtPkgTok( MingleIdentifier id )
    {
        return 
            MingleModels.format( id, MingleIdentifierFormat.LC_CAMEL_CAPPED );
    }

    private
    JvPackage
    jvPackageOf( MingleNamespace ns )
    {
        List< CharSequence > pkgToks = Lang.newList();

        Map.Entry< MingleNamespace, String > pref = getLongestPrefixMatch( ns );
        int indx = 0;
        if ( pref != null ) 
        {
            pkgToks.add( pref.getValue() );
            indx = length( pref.getKey() );
        }

        Iterator< MingleIdentifier > it = ns.getParts().iterator();
        while ( indx-- > 0 ) it.next();
        while ( it.hasNext() ) pkgToks.add( fmtPkgTok( it.next() ) );
        pkgToks.add( fmtPkgTok( ns.getVersion() ) );

        return new JvPackage( Strings.join( ".", pkgToks ) );
    }

    private
    JvPackage
    jvPackageOf( QualifiedTypeName qn )
    {
        return jvPackageOf( qn.getNamespace() );
    }

    JvQname
    jvQnameOf( QualifiedTypeName qn )
    {
        JvQname res = cgCtl.typeMap.get( qn );

        if ( res == null )
        {
            return 
                JvQname.create( jvPackageOf( qn ), JvTypeName.ofQname( qn ) );
        }
        else return res;
    }

    JvPackage
    jvPackageOf( TypeDefinition td )
    {
        return jvQnameOf( td.getName() ).pkg;
    }

    JvTypeName
    jvTypeNameOf( QualifiedTypeName qn )
    {
        return jvQnameOf( qn ).nm;
    }

    TypeDefinition
    getSuperDef( TypeDefinition td )
    {
        MingleTypeReference ref = td.getSuperType();

        if ( ref == null ) return null;
        else 
        {
            QualifiedTypeName qn = 
                (QualifiedTypeName) MingleModels.typeNameIn( ref );

            return runtime().getTypes().expectType( qn );
        }
    }
}
