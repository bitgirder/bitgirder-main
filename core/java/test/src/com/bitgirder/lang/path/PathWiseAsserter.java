package com.bitgirder.lang.path;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
class PathWiseAsserter< V >
extends State
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final ObjectPath< V > path;
    private final ObjectPathFormatter< ? super V > fmt;

    public
    PathWiseAsserter( ObjectPath< V > path,
                      ObjectPathFormatter< ? super V > fmt ) 
    { 
        this.path = inputs.notNull( path, "path" );
        this.fmt = inputs.notNull( fmt, "fmt" );
    }

    public
    PathWiseAsserter( ObjectPathFormatter< ? super V > fmt )
    { 
        this( ObjectPath.< V >getRoot(), fmt );
    }

    public final ObjectPath< V > getPath() { return path; }
    public final ObjectPathFormatter< ? super V > getFormatter() { return fmt; }

    public CharSequence formatPath() { return ObjectPaths.format( path, fmt ); }

    public
    final
    PathWiseAsserter< V >
    descend( V elt )
    {
        inputs.notNull( elt, "elt" );

        return new PathWiseAsserter< V >( path.descend( elt ), fmt );
    }

    @Override
    public
    final
    IllegalStateException
    createException( CharSequence inputName,
                     CharSequence msg )
    {
        StringBuilder sb = new StringBuilder();

        if ( path != null ) 
        {
            ObjectPaths.appendFormat( path, fmt, sb );
            sb.append( ": " );
        }

        return new IllegalStateException( sb.append( msg ).toString() );
    }
}
