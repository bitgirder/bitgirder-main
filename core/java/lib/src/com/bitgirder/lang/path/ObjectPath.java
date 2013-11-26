package com.bitgirder.lang.path;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.ObjectReceiver;

import java.util.Deque;
import java.util.Iterator;

public
class ObjectPath< E >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static ObjectPath< Object > EMPTY_PATH = 
        new ObjectPath< Object >( null );

    private final ObjectPath< E > parent; // null only when this is a root

    // parent == null if and only if this is a root
    private ObjectPath( ObjectPath< E > parent ) { this.parent = parent; }

    ObjectPath( ObjectPath< E > parent,
                String paramName )
    {
        this( 
            paramName == null ? 
                state.notNull( parent, "parent" ) :
                inputs.notNull( parent, paramName )
        );
    }

    public final ObjectPath< E > getParent() { return parent; }
    public boolean isEmpty() { return this == EMPTY_PATH; }

    public
    final
    DictionaryPath< E >
    descend( E key )
    {
        // we let DictionaryPath.create do the key null checking
        return DictionaryPath.< E >create( this, key );
    }

    Iterator< ObjectPath< E > >
    getDescent()
    {
        Deque< ObjectPath< E > > d = Lang.newDeque();

        for ( ObjectPath< E > p = this; p.parent != null; p = p.parent )
        {
            d.push( p );
        }

        return d.iterator();
    }

    public
    void
    visitDescent( ObjectReceiver< ObjectPath< E > > vis )
        throws Exception
    {
        inputs.notNull( vis, "vis" );

        for ( Iterator< ObjectPath< E > > it = getDescent(); it.hasNext(); ) {
            vis.receive( it.next() );
        }
    }

    private
    void
    formatPathElement( StringBuilder sb,
                       ObjectPathFormatter< ? super E > f,
                       ObjectPath< E > p )
    {
        if ( p instanceof DictionaryPath )
        {
            @SuppressWarnings( "unchecked" )
            DictionaryPath< E > dp = (DictionaryPath< E >) p;
            f.formatDictionaryKey( sb, dp.getKey() );
        }
        else if ( p instanceof ListPath )
        {
            @SuppressWarnings( "unchecked" )
            ListPath< E > lp = (ListPath< E >) p;
            f.formatListIndex( sb, lp.getIndex() );
        }
        else state.fail( "Unexpected path element:", p );
    }

    private
    void
    formatPathInterior( StringBuilder sb,
                        ObjectPathFormatter< ? super E > f,
                        Iterator< ObjectPath< E > > it )
    {
        ObjectPath< E > prev = null;

        while ( it.hasNext() )
        {
            ObjectPath< E > p = it.next();

            // if this is not the first element and we're not formatting a list
            // path then add a separator first
            if ( ! ( prev == null || p instanceof ListPath ) ) 
            {
                f.formatSeparator( sb );
            }

            formatPathElement( sb, f, p );
            prev = p;
        }
    }

    final
    StringBuilder
    appendFormat( ObjectPathFormatter< ? super E > f,
                  StringBuilder sb )
    {
        f.formatPathStart( sb );
        formatPathInterior( sb, f, getDescent() );

        return sb;
    }

    public
    final
    ImmutableListPath< E >
    startImmutableList( int idx )
    {
        inputs.nonnegativeI( idx, "idx" );
        return ImmutableListPath.< E >start( this, idx );
    }

    public
    final
    ImmutableListPath< E >
    startImmutableList()
    {
        return startImmutableList( 0 );
    }

    public
    final
    MutableListPath< E >
    startMutableList( int idx )
    {
        inputs.nonnegativeI( idx, "idx" );
        return new MutableListPath< E >( this, idx );
    }

    public
    final
    MutableListPath< E >
    startMutableList()
    {
        return startMutableList( 0 ); 
    }

    public
    static
    < E >
    ObjectPath< E >
    getRoot() 
    { 
        @SuppressWarnings( "unchecked" )
        ObjectPath< E > res = (ObjectPath< E >) EMPTY_PATH;

        return res;
    }

    public
    static
    < E >
    ObjectPath< E >
    getRoot( E root )
    {
        inputs.notNull( root, "root" );

        return ObjectPath.< E >getRoot().descend( root );
    }
}
