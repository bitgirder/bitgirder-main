package com.bitgirder.lang.path;

public
final
class ImmutableListPath< E >
extends ListPath< E >
{
    private final int indx;

    private
    ImmutableListPath( int indx,
                       ObjectPath< E > parent,
                       String paramName )
    {
        super( parent, paramName );

        this.indx = indx;
    }

    public int getIndex() { return indx; }

    public
    ImmutableListPath< E >
    next()
    {
        return new ImmutableListPath< E >( indx + 1, getParent(), null );
    }

    public
    static
    < E >
    ImmutableListPath< E >
    start( ObjectPath< E > parent )
    {
        return new ImmutableListPath< E >( 0, parent, "parent" );
    }
}
