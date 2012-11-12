package com.bitgirder.mingle.bind;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPathFormatter;
import com.bitgirder.lang.path.ObjectPaths;

import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleIdentifier;

// Extensible only within this package
public
class MingleBindingException
extends RuntimeException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final static ObjectPathFormatter< MingleIdentifier > INBOUND_FORMATTER =
        MingleModels.getIdentifierPathFormatter();

    final static ObjectPathFormatter< Object > OUTBOUND_FORMATTER = 
        ObjectPaths.DOT_FORMATTER;

    MingleBindingException( String msg ) { super( msg ); }

    MingleBindingException( String msg,
                            Throwable th )
    {
        super( msg, th );
    }

    private
    static
    < V >
    MingleBindingException
    createImpl( String msg,
                ObjectPath< ? extends V > path,
                ObjectPathFormatter< ? super V > fmt,
                Exception ex )
    {
        state.notNull( path, "path" );

        StringBuilder sb = new StringBuilder();

        // the string could still be empty if this is a root path, so we only
        // add the ': ' separator if the path formatted to something nontrivial
        ObjectPaths.appendFormat( path, fmt, sb );
        if ( sb.length() > 0 ) sb.append( ": " );

        sb.append( msg );

        return new MingleBindingException( sb.toString(), ex );
    }

    static
    < V >
    MingleBindingException
    create( String msg,
            ObjectPath< ? extends V > path,
            ObjectPathFormatter< ? super V > fmt )
    {
        return createImpl( msg, path, state.notNull( fmt, "fmt" ), null );
    }

    static
    < V >
    MingleBindingException
    create( String msg,
            ObjectPath< ? extends V > path,
            ObjectPathFormatter< ? super V > fmt,
            Exception ex )
    {
        return createImpl(
            msg, path, state.notNull( fmt, "fmt" ), state.notNull( ex, "ex" ) );
    }

    private
    static
    MingleBindingException
    createOutboundImpl( String msg,
                        ObjectPath< String > path,
                        Exception ex )
    {
        return createImpl( msg, path, OUTBOUND_FORMATTER, ex );
    }

    static
    MingleBindingException
    createOutbound( String msg,
                    ObjectPath< String > path )
    {
        return createOutboundImpl( msg, path, null );
    }

    static
    MingleBindingException
    createOutbound( String msg,
                    ObjectPath< String > path,
                    Exception ex )
    {
        return createOutboundImpl( msg, path, state.notNull( ex, "ex" ) );
    }

    private
    static
    MingleBindingException
    createInboundImpl( String msg,
                       ObjectPath< MingleIdentifier > path,
                       Exception ex )
    {
        return createImpl( msg, path, INBOUND_FORMATTER, ex );
    }

    static
    MingleBindingException
    createInbound( String msg,
                   ObjectPath< MingleIdentifier > path )
    {
        return createInboundImpl( msg, path, null );
    }

    static
    MingleBindingException
    createInbound( String msg,
                   ObjectPath< MingleIdentifier > path,
                   Exception ex )
    {
        return createInboundImpl( msg, path, inputs.notNull( ex, "ex" ) );
    }
}
