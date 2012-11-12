package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.parser.SyntaxException;
import com.bitgirder.parser.SourceTextLocation;

public
abstract
class AbstractParsedStringExchanger< V >
extends AbstractValueExchanger< V >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    protected
    AbstractParsedStringExchanger( AtomicTypeReference typ,
                                   Class< V > valCls )
    {
        // let superclass do null-checks
        super( typ, valCls );
    }

    protected
    abstract
    V
    parse( CharSequence str )
        throws SyntaxException;
    
    private
    MingleValidationException
    createRethrow( SyntaxException se,
                   ObjectPath< MingleIdentifier > path )
    {
        SourceTextLocation loc = se.getLocation();

        String msg;

        if ( loc == null ) msg = se.getRawMessage();
        else
        {
            msg =
                new StringBuilder().
                    append( "[col " ).append( loc.getColumn() ).append( "]: " ).
                    append( se.getRawMessage() ).
                    toString();
        }

        return new MingleValidationException( msg, path );
    }

    public
    final
    Object
    asJavaValue( MingleValue mv,
                 ObjectPath< MingleIdentifier > path )
    {
        state.notNull( mv, "mv" );
        state.notNull( path, "path" );

        MingleString ms = 
            (MingleString) MingleModels.asMingleInstance(
                MingleModels.TYPE_REF_MINGLE_STRING,
                mv,
                path
            );
 
        try { return parse( ms ); }
        catch ( SyntaxException se ) { throw createRethrow( se, path ); }
    }

    protected
    abstract
    CharSequence
    asString( V obj );

    public
    final
    MingleValue
    implAsMingleValue( V obj,
                       ObjectPath< String > path )
    {
        inputs.notNull( obj, "obj" );

        CharSequence str = asString( obj );
        state.isFalse( str == null, "impl return null from asString()" );

        return MingleModels.asMingleString( str );
    }
}
