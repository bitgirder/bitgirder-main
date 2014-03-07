package com.bitgirder.lang.path;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.Iterator;

public
final
class ObjectPaths
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public final static ObjectPathFormatter< Object > DOT_FORMATTER =
        new PathSepFormatter( "." );

    public final static ObjectPathFormatter< Object > SLASH_FORMATTER =
        new PathSepFormatter( "/" );

    private ObjectPaths() {}

    public
    static
    < V >
    ObjectPath< V >
    rootOf( ObjectPath< V > p )
    {
        inputs.notNull( p, "p" );

        while ( p.getParent() != null ) p = p.getParent();

        return p;
    }

    private
    static
    < V >
    void
    doAppendFormat( ObjectPath< ? extends V > path, 
                    ObjectPathFormatter< ? super V > fmt,
                    StringBuilder sb )
    {
        inputs.notNull( path, "path" );
        inputs.notNull( fmt, "fmt" );

        path.appendFormat( fmt, sb );
    }

    public
    static
    < V >
    void
    appendFormat( ObjectPath< ? extends V > path,
                  ObjectPathFormatter< ? super V > fmt,
                  StringBuilder sb )
    {
        doAppendFormat( path, fmt, inputs.notNull( sb, "sb" ) );
    }

    public
    static
    < V >
    CharSequence
    format( ObjectPath< ? extends V > path,
            ObjectPathFormatter< ? super V > fmt )
    {
        StringBuilder res = new StringBuilder();
        doAppendFormat( path, fmt, res );

        return res;
    }

    private
    final
    static
    class PathSepFormatter
    implements ObjectPathFormatter< Object >
    {
        private final String pathSep;

        private PathSepFormatter( String pathSep ) { this.pathSep = pathSep; }

        public
        void
        formatDictionaryKey( StringBuilder sb,
                             Object key )
        {
            sb.append( key.toString() );
        }

        public
        void
        formatListIndex( StringBuilder sb,
                         int indx )
        {
            sb.append( "[ " ).
               append( indx ).
               append( " ]" );
        }

        public void formatPathStart( StringBuilder sb ) {}

        public 
        void 
        formatSeparator( StringBuilder sb ) 
        { 
            sb.append( pathSep ); 
        }
    }

    private
    static
    boolean
    areEqualNodes( ObjectPath< ? > p1,
                   ObjectPath< ? > p2 )
    {
        if ( p1 instanceof DictionaryPath )
        {
            DictionaryPath< ? > dp1 = (DictionaryPath< ? >) p1;

            if ( ! ( p2 instanceof DictionaryPath ) ) return false;
            return dp1.getKey().equals( ( (DictionaryPath< ? >) p2 ).getKey() );
        }

        if ( p1 instanceof ListPath )
        {
            ListPath< ? > lp1 = (ListPath< ? >) p1;
            if ( ! ( p2 instanceof ListPath ) ) return false;
            return lp1.getIndex() == ( (ListPath< ? >) p2 ).getIndex();
        }

        state.isTrue( p1.getParent() == null ); // p1 must be a root
        return p2.getParent() == null;
    }

    public
    static
    < V >
    boolean
    areEqual( ObjectPath< V > p1,
              ObjectPath< V > p2 )
    {
        if ( p1 == null ) return p2 == null;
        if ( p2 == null ) return false;

        if ( ! areEqualNodes( p1, p2 ) ) return false;
        return areEqual( p1.getParent(), p2.getParent() );
    }

    private
    static
    < V >
    ObjectPath< V >
    applyPathCopy( ObjectPath< V > targ,
                   ObjectPath< V > elt,
                   boolean mutable )
    {
        if ( elt instanceof DictionaryPath ) {
            DictionaryPath< V > dp = Lang.castUnchecked( elt );
            return targ.descend( dp.getKey() );
        } else if ( elt instanceof ListPath ) {
            ListPath< V > lp = Lang.castUnchecked( elt );
            int i = lp.getIndex();
            if ( mutable ) return targ.startMutableList( i );
            return targ.startImmutableList( i );
        } else {
            throw state.createFailf( "unhandled path type: %s", elt );
        }
    }

    private
    static
    < V >
    ObjectPath< V >
    asCopy( ObjectPath< V > p,
            boolean mutable )
    {
        inputs.notNull( p, "p" );

        ObjectPath< V > res = ObjectPath.getRoot();

        for ( Iterator< ObjectPath< V > > it = p.getDescent(); it.hasNext(); ) {
            res = applyPathCopy( res, it.next(), mutable );
        }

        return res;
    }

    public
    static
    < V >
    ObjectPath< V >
    asMutableCopy( ObjectPath< V > p )
    {
        return asCopy( p, true );
    }
            
    // currently makes a copy; more efficient versions may opt to only make
    // copies as needed
    public
    static
    < V >
    ObjectPath< V >
    asImmutableCopy( ObjectPath< V > p )
    {
        return asCopy( p, false );
    }

    public
    static
    < V >
    ObjectPath< V >
    descend( ObjectPath< V > p,
             V elt )
    {
        inputs.notNull( elt, "elt" );

        return p == null ? ObjectPath.getRoot( elt ) : p.descend( elt );
    }
}
