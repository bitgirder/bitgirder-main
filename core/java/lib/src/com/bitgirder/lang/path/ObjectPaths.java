package com.bitgirder.lang.path;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

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
}
