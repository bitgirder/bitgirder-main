package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPaths;

import com.bitgirder.parser.SyntaxException;

public
abstract
class AbstractValueExchanger< V >
implements MingleValueExchanger
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final AtomicTypeReference typ;
    private final Class< V > valCls;

    protected
    AbstractValueExchanger( AtomicTypeReference typ,
                            Class< V > valCls )
    {
        this.typ = inputs.notNull( typ, "typ" );
        this.valCls = inputs.notNull( valCls, "valCls" );
    }

    public final Class< V > getJavaClass() { return valCls; }
    public final AtomicTypeReference getMingleType() { return typ; }

    // cause can be null
    protected
    MingleValidationException
    createInbound( String msg,
                   ObjectPath< MingleIdentifier > path,
                   Throwable cause )
    {
        return new MingleValidationException( msg, path, cause );
    }

    protected
    MingleValidationException
    createInbound( String msg,
                   ObjectPath< MingleIdentifier > path )
    {
        return createInbound( msg, path, null );
    }

    protected
    final
    MingleValidationException
    createRethrow( SyntaxException se,
                   MingleSymbolMapAccessor acc,
                   MingleIdentifier key )
    {
        inputs.notNull( se, "se" );
        inputs.notNull( acc, "acc" );
        inputs.notNull( key, "key" );

        return createInbound( se.getMessage(), acc.getPath().descend( key ) );
    }

    protected
    final
    MingleIdentifier
    createIdentifier( CharSequence idStr,
                      ObjectPath< MingleIdentifier > path )
    {
        inputs.notNull( idStr, "idStr" );
        inputs.notNull( path, "path" );

        try { return MingleIdentifier.parse( idStr ); }
        catch ( SyntaxException se )
        {
            throw createInbound( se.getMessage(), path, se );
        }
    }

    // Convenience method to effectively inline null checks of acc and key and
    // then call expectString()
    private
    MingleString
    expectString( MingleSymbolMapAccessor acc,
                  MingleIdentifier key )
    {
        inputs.notNull( acc, "acc" );
        inputs.notNull( key, "key" );

        return acc.expectMingleString( key );
    }

    protected
    final
    MingleTypeReference
    expectTypeRef( MingleSymbolMapAccessor acc,
                   MingleIdentifier key )
    {
        try { return MingleTypeReference.parse( expectString( acc, key ) ); }
        catch ( SyntaxException se ) { throw createRethrow( se, acc, key ); }
    }

    protected
    final
    MingleNamespace
    expectNamespace( MingleSymbolMapAccessor acc,
                     MingleIdentifier key )
    {
        try { return MingleNamespace.parse( expectString( acc, key ) ); }
        catch ( SyntaxException se ) { throw createRethrow( se, acc, key ); }
    }

    protected
    final
    MingleIdentifier
    expectIdentifier( MingleSymbolMapAccessor acc,
                      MingleIdentifier key )
    {
        try { return MingleIdentifier.parse( expectString( acc, key ) ); }
        catch ( SyntaxException se ) { throw createRethrow( se, acc, key ); }
    }

    protected
    final
    RuntimeException
    createOutbound( String msg,
                    ObjectPath< String > path )
    {
        return 
            new RuntimeException( 
                ObjectPaths.format( path, ObjectPaths.DOT_FORMATTER ) + 
                " " +
                msg
            );
    }
    
    private
    QualifiedTypeName
    expectQname( AtomicTypeReference typ,
                 ObjectPath< MingleIdentifier > path )
    {
        AtomicTypeReference.Name nm = typ.getName();

        if ( nm instanceof QualifiedTypeName ) return (QualifiedTypeName) nm;
        else
        {
            throw 
                createInbound( "Got non-qualified atomic type: " +  typ, path );
        } 
    }

    private
    MingleSymbolMap
    assertInboundType( MingleStructure ms,
                       QualifiedTypeName qnExpct,
                       ObjectPath< MingleIdentifier > path )
    {
        AtomicTypeReference typ = ms.getType();
        QualifiedTypeName qn = expectQname( typ, path );

        if ( ! qn.equals( qnExpct ) )
        {
            AtomicTypeReference typExpct = 
                AtomicTypeReference.create( qnExpct );

            throw new MingleTypeCastException( typExpct, typ, path );
        }

        return ms.getFields();
    }

    protected
    final
    MingleSymbolMap
    expectSymbolMap( AtomicTypeReference typ,
                     MingleValue mv,
                     ObjectPath< MingleIdentifier > path )
    {
        inputs.notNull( typ, "typ" );
        inputs.notNull( mv, "mv" );
        inputs.notNull( path, "path" );

        // If this cast fails it is not something we report to the caller and is
        // just an issue with our code
        QualifiedTypeName qn = (QualifiedTypeName) typ.getName();

        if ( mv instanceof MingleStructure ) 
        {
            return assertInboundType( (MingleStructure) mv, qn, path );
        }
        else if ( mv instanceof MingleSymbolMap ) return (MingleSymbolMap) mv;
        else
        {
            throw 
                new MingleValidationException(
                    "Invalid structure type: " + typ, path );
        }
    }

    protected
    abstract
    MingleValue
    implAsMingleValue( V val,
                       ObjectPath< String > path );

    public
    final
    MingleValue
    asMingleValue( Object obj,
                   ObjectPath< String > path )
    {
        if ( valCls.isInstance( obj ) )
        {
            return implAsMingleValue( valCls.cast( obj ), path );
        }
        else
        {
            throw 
                createOutbound(
                    "Expected instance of " + valCls + " but got " +
                    ( obj == null ? null : obj.getClass().getName() ),
                    path
                );
        }
    }
}
