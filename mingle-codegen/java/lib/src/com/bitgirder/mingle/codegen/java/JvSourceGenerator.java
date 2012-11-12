package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.codegen.AbstractSourceGenerator;

import java.util.List;
import java.util.Iterator;

final
class JvSourceGenerator
extends AbstractSourceGenerator
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private
    boolean
    isPackage( JvVisibility v )
    {
        return v == JvVisibility.PACKAGE;
    }

    private
    boolean
    isPackage( JvModifiers mods )
    {
        return isPackage( mods.vis );
    }

    private
    CharSequence
    asString( JvVisibility v )
    {
        return v.name().toLowerCase();
    }

    private
    void
    writeParams( List< JvParam > params )
    {
        if ( params.isEmpty() ) write( "()" );
        else
        {
            write( "( " );
            setIndent();

            for ( Iterator< JvParam > it = params.iterator(); it.hasNext(); )
            {
                JvParam p = it.next();
                writeExpression( p.type );
                write( " " );
                writeExpression( p.id );

                if ( it.hasNext() ) writeLine( "," );
                else
                {
                    write( " )" );
                    popIndent();
                }
            }
        }
    }

    private
    void
    writeAssign( JvAssign a )
    {
        writeExpression( a.left );
        write( " = " );
        writeExpression( a.right );
    }

    private
    void
    writeAccess( JvAccess a )
    {
        writeExpression( a.left );
        write( "." );
        writeExpression( a.right );
    }

    private
    void
    writeStatement( JvStatement s )
    {
        writeExpression( s.e );
        writeLine( ";" );
    }

    private
    void
    writeReturn( JvReturn r )
    {
        write( "return " );
        writeExpression( r.e );
        writeLine( ";" );
    }

    private
    void
    writeCallParamPart( List< JvExpression > params )
    {
        if ( params.isEmpty() ) write( "()" );
        else
        {
            write( "( " );

            for ( Iterator< JvExpression > it = params.iterator();
                    it.hasNext(); )
            {
                writeExpression( it.next() );
                write( it.hasNext() ? ", " : " )" );
            }
        }
    }
    
    private
    void
    writeInvocationPart( JvInvocation i )
    {
        writeExpression( i.target );
        writeCallParamPart( i.params );
    }

    private
    void
    writeFuncCall( JvFuncCall f )
    {
        writeInvocationPart( f );
    }

    private
    void
    writeInstantiate( JvInstantiate i )
    {
        write( "new " );
        writeInvocationPart( i );
    }

    private
    void
    writeOptTypeArgs( List< JvType > args )
    {
        if ( ! args.isEmpty() )
        {
            write( "< " );
            for ( Iterator< JvType > it = args.iterator(); it.hasNext(); )
            {
                writeType( it.next() );
                if ( it.hasNext() ) write( ", " ); else write( " >" );
            }
        }
    }

    private
    void
    writeTypeExpression( JvTypeExpression e )
    {
        writeType( e.type );

        writeOptTypeArgs( e.args );

        if ( e.next != null )
        {
            write( "." );
            writeType( e.next );
        }
    }

    private
    void
    writeTypeExtendParam( JvTypeExtendParameter p )
    {
        writeType( p.type );
        write( " extends " );
        writeType( p.sprTyp );
    }

    private
    void
    writeType( JvType t )
    {
        if ( ( t instanceof JvTypeName ) ||
             ( t instanceof JvQname ) ||
             ( t instanceof JvPrimitiveType ) )
        {
            write( t.toString() );
        }
        else if ( t instanceof JvTypeExpression )
        {
            writeTypeExpression( (JvTypeExpression) t );
        }
        else if ( t instanceof JvTypeExtendParameter )
        {
            writeTypeExtendParam( (JvTypeExtendParameter) t );
        }
        else throw state.createFail( "Unhandled type:", t );
    }

    private
    void
    writeCast( JvCast c )
    {
        write( "(" );
        writeType( c.target );
        write( ") " );
        writeExpression( c.e );
    }

    // Using JSON string serialization as ours as well
    private
    void
    writeString( JvString s )
    {
        write( Lang.getRfc4627String( s.str ) );
    }

    private
    void
    writeNumber( JvNumber n )
    {
        write( n.lit );
    }

    private
    void
    writeParExpression( JvParExpression e )
    {
        write( "( " );
        writeExpression( e.e );
        write( " )" );
    }

    private
    void
    writeLocalVar( JvLocalVar v )
    {
        writeType( v.type );
        write( " " );
        write( v.name );

        if ( v.assign != null )
        {
            write( " = " );
            writeExpression( v.assign );
        }

        writeLine( ";" );
    }

    private
    void
    writeAnnotation( JvAnnotation a )
    {
        write( "@" );
        writeType( a.type );
        writeLine();
    }

    private
    void
    writeBranch( JvBranch b )
    {
        write( "if ( " );
        writeExpression( b.test );
        write( " )" );
        writeBlock( b.onTrue );

        if ( ! b.onFalse.isEmpty() )
        {
            write( "else" );
            writeBlock( b.onFalse );
        }
    }

    private
    void
    writeCase( JvCase c )
    {
        write( "case " );
        writeExpression( c.label );
        writeLine( ":" );

        pushIndent();

        for ( JvExpression e : c.body ) 
        {
            writeExpression( e );
            writeLine();
        }

        popIndent();
    }

    private
    void
    writeSwitchDefault( List< JvExpression > defl )
    {
        writeLine( "default:" );

        pushIndent();

        for ( JvExpression e : defl )
        {
            writeExpression( e );
            writeLine();
        }

        popIndent();
    }

    private
    void
    writeSwitch( JvSwitch s )
    {
        write( "switch ( " );
        writeExpression( s.target );
        writeLine( " )" );
        writeLine( "{" );
        pushIndent();
        for ( JvCase c : s.cases ) writeCase( c );
        if ( ! s.defl.isEmpty() ) writeSwitchDefault( s.defl );
        popIndent();
        writeLine( "}" );
    }

    private
    void
    writeCatch( JvCatch c )
    {
        write( "catch ( " );
        writeType( c.type );
        write( " " );
        write( c.id );
        write( " )" );

        if ( c.body.isEmpty() ) writeLine( " {}" );
        else
        {
            writeLine(); // close the final ')' of the catch decl
            writeLine( "{" );
            pushIndent();
            for ( JvExpression e : c.body ) writeExpression( e );
            popIndent();
            writeLine( "}" );
        }
    }

    private
    void
    writeTry( JvTry t )
    {
        writeLine( "try" );
        writeLine( "{" );
        pushIndent();
        for ( JvExpression e : t.body ) writeExpression( e );
        popIndent();
        writeLine( "}" );

        for ( JvCatch c : t.catches ) writeCatch( c );
    }

    private
    void
    writeExpression( JvExpression e )
    {
        if ( e instanceof JvId ) write( e );
        else if ( e instanceof JvLiteral ) write( e );
        else if ( e instanceof JvType ) writeType( (JvType) e );
        else if ( e instanceof JvAssign ) writeAssign( (JvAssign) e );
        else if ( e instanceof JvAccess ) writeAccess( (JvAccess) e );
        else if ( e instanceof JvStatement ) writeStatement( (JvStatement) e );
        else if ( e instanceof JvReturn ) writeReturn( (JvReturn) e );
        else if ( e instanceof JvFuncCall ) writeFuncCall( (JvFuncCall) e );
        else if ( e instanceof JvCast ) writeCast( (JvCast) e );
        else if ( e instanceof JvString ) writeString( (JvString) e );
        else if ( e instanceof JvNumber ) writeNumber( (JvNumber) e );
        else if ( e instanceof JvLocalVar ) writeLocalVar( (JvLocalVar) e );
        else if ( e instanceof JvBranch ) writeBranch( (JvBranch) e );
        else if ( e instanceof JvSwitch ) writeSwitch( (JvSwitch) e );
        else if ( e instanceof JvInstantiate ) 
        {
            writeInstantiate( (JvInstantiate) e );
        }
        else if ( e instanceof JvParExpression ) 
        {
            writeParExpression( (JvParExpression) e );
        }
        else if ( e instanceof JvTry ) writeTry( (JvTry) e );
        else state.fail( "Unhandled expression:", e );
    }

    private
    void
    writeBlock( List< JvExpression > body )
    {
        if ( body.isEmpty() ) writeLine( " {}" );
        else
        {
            writeLine();
            writeLine( "{" );
            pushIndent();
            for ( JvExpression e : body ) writeExpression( e );
            popIndent();
            writeLine( "}" );
        }
    }

    private
    void
    writeOptVisibilityLine( JvVisibility vis )
    {
        if ( vis != JvVisibility.PACKAGE ) writeLine( asString( vis ) );
    }

    private
    void
    writeDecl( JvDeclaredType t,
               String kwd )
    {
        write( kwd, " " );
        writeType( t.name );
        writeOptTypeArgs( t.typeParams );
        writeLine();
    }

    private
    void
    writeField( JvField f )
    {
        if ( ! isPackage( f.mods ) ) write( asString( f.mods.vis ) );
        if ( f.mods.isFinal ) write( " final" );
        if ( f.mods.isStatic ) write( " static" );

        write( " " );
        writeType( f.type );

        write( " " );
        writeExpression( f.name );

        if ( f.assign != null ) 
        {
            write( " = " );
            writeExpression( f.assign );
        }

        writeLine( ";" );
    }

    private
    int
    writeFields( List< JvField > flds,
                 boolean wantStatic )
    {
        int res = 0;

        for ( JvField f : flds )
        {
            boolean isStatic = f.mods.isStatic;

            if ( ( isStatic && wantStatic ) || 
                 ( ! ( isStatic || wantStatic ) ) )
            {
                writeField( f );
                ++res;
            }
        }

        return res;
    }

    private
    void
    writeFields( List< JvField > flds )
    {
        if ( writeFields( flds, true ) > 0 ) writeLine();
        writeFields( flds, false );
    }

    private
    void
    writeImplemented( List< JvType > implemented )
    {
        write( "implements " );
        setIndent();

        for ( Iterator< JvType > it = implemented.iterator(); it.hasNext(); )
        {
            writeType( it.next() );

            if ( it.hasNext() ) writeLine( "," ); 
            else
            {
                writeLine();
                popIndent();
            }
        }
    }

    private
    void
    writeClassOpen( JvClass cls )
    {
        writeOptVisibilityLine( cls.mods.vis );
        if ( cls.mods.isFinal ) writeLine( "final" );
        if ( cls.mods.isAbstract ) writeLine( "abstract" );
        if ( cls.mods.isStatic ) writeLine( "static" );
        writeDecl( cls, "class" );

        if ( cls.sprTyp != null ) 
        {
            write( "extends " );
            writeType( cls.sprTyp );
            writeLine();
        }

        if ( ! cls.implemented.isEmpty() ) writeImplemented( cls.implemented );
        writeLine( "{" );
        pushIndent();
    }

    private
    void
    writeThrowsClause( List< JvType > thrown )
    {
        if ( ! thrown.isEmpty() )
        {
            writeLine();
            pushIndent(); // POP1
            write( "throws " ); // POP2
            setIndent();

            for ( Iterator< JvType > it = thrown.iterator(); it.hasNext(); )
            {
                writeType( it.next() );

                if ( it.hasNext() ) writeLine( ", " ); 
                else 
                {
                    popIndent(); // pops indent from POP2
                    popIndent(); // pops POP1
                }
            }
        }
    }

    private
    void
    writeConstructor( JvClass cls,
                      JvConstructor cons )
    {
        if ( ! isPackage( cons.vis ) ) writeLine( asString( cons.vis ) );
        write( cls.name );
        writeParams( cons.params );
        writeBlock( cons.body );
    }

    private
    void
    writeConstructors( JvClass cls )
    {
        for ( JvConstructor cons : cls.constructors ) 
        {
            writeLine();
            writeConstructor( cls, cons );
        }
    }

    private
    void
    writeMethod( JvMethod m )
    {
        for ( JvAnnotation a : m.anns ) writeAnnotation( a );
        if ( ! isPackage( m.mods.vis ) ) writeLine( asString( m.mods.vis ) );
        if ( m.mods.isFinal ) writeLine( "final" );
        if ( m.mods.isStatic ) writeLine( "static" );
        if ( m.mods.isAbstract ) writeLine( "abstract" );
        
        writeType( m.retType );
        writeLine();

        write( m.name );
        writeParams( m.params );
        writeThrowsClause( m.thrown );
        
        if ( m.mods.isAbstract ) writeLine( ";" );
        else writeBlock( m.body );
    }

    private
    void
    writeMethods( JvClass cls )
    {
        for ( JvMethod m : cls.methods )
        {
            writeLine();
            writeMethod( m );
        }
    }

    private
    void
    writeDeclaredTypes( List< JvDeclaredType > types )
    {
        for ( JvDeclaredType decl : types )
        {
            writeLine();
            writeDeclaredType( decl );
        }
    }

    private
    void
    writeClassBody( JvClass cls )
    {
        writeFields( cls.fields );
        writeConstructors( cls );
        writeMethods( cls );
        writeDeclaredTypes( cls.nestedTypes );
    }

    private
    void
    writeClassClose()
    {
        popIndent();
        writeLine( "}" );
    }

    private
    void
    writeClass( JvClass cls )
    {
        writeClassOpen( cls );
        writeClassBody( cls );
        writeClassClose();
    }

    private
    void
    writeEnumConstants( JvEnumDecl en )
    {
        for ( Iterator< JvExpression > it = en.constants.iterator();
                it.hasNext(); )
        {
            writeExpression( it.next() );
            writeLine( it.hasNext() ? "," : ";" );
        }
    }

    private
    void
    writeEnum( JvEnumDecl en )
    {
        writeOptVisibilityLine( en.vis );
        writeDecl( en, "enum" );
        writeLine( "{" );
        pushIndent();
        writeEnumConstants( en );
        writeFields( en.fields );
        writeDeclaredTypes( en.nestedTypes );
        popIndent();
        writeLine( "}" );
    }

    private
    void
    writeDeclaredType( JvDeclaredType decl )
    {
        if ( decl instanceof JvClass ) writeClass( (JvClass) decl );
        else if ( decl instanceof JvEnumDecl ) writeEnum( (JvEnumDecl) decl );
        else throw state.createFail( "Unhandled decl:", decl );
    }

    CharSequence
    asSource( JvCompilationUnit u )
    {
        state.notNull( u, "u" );
        u.validate();

        writeLine( "package ", u.pkg, ";" );
        writeLine();
        
        writeDeclaredType( u.decl );

        return getString();
    }
}
