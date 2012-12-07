package com.bitgirder.lang.path;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.TypedString;

import com.bitgirder.test.Test;

@Test
final
class PathTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private
    final
    static
    class Node
    extends TypedString< Node >
    {
        private Node( CharSequence str ) { super( str, "str" ); }
    }

    private
    final
    static
    class FormatterImpl
    implements ObjectPathFormatter< Node >
    {
        public void formatPathStart( StringBuilder sb ) { sb.append( '/' ); }
        public void formatSeparator( StringBuilder sb ) { sb.append( '/' ); }

        public
        void
        formatDictionaryKey( StringBuilder sb,
                             Node n )
        {
            sb.append( n );
        }

        public
        void
        formatListIndex( StringBuilder sb,
                         int indx )
        {
            sb.append( "[ " ).append( indx ).append( " ]" );
        }
    }

    private
    void
    assertFormat( CharSequence expct,
                  ObjectPath< Node > p )
    {
        state.equalString( 
            expct, ObjectPaths.format( p, new FormatterImpl() ) );
    }

    @Test
    private
    void
    testFormat0()
    {
        assertFormat( 
            "/node1/node2[ 2 ]/node3", 
            ObjectPath.< Node >getRoot().
                descend( new Node( "node1" ) ).
                descend( new Node( "node2" ) ).
                startImmutableList().
                next().
                next().
                descend( new Node( "node3" ) ) );
    }

    @Test
    private
    void
    testFormat1()
    {
        assertFormat( 
            "/node1[ 1 ][ 2 ]/node2/node3",
            ObjectPath.< Node >getRoot().
                descend( new Node( "node1" ) ).
                startImmutableList().
                next().
                startImmutableList().
                next().
                next().
                descend( new Node( "node2" ) ).
                descend( new Node( "node3" ) ) );
    }

    @Test
    private
    void
    testDefaultFormats()
    {
        ObjectPath< Node > path =
            ObjectPath.< Node >getRoot().
                descend( new Node( "node1" ) ).
                startImmutableList().
                next().
                startImmutableList().
                next().
                next().
                descend( new Node( "node2" ) ).
                descend( new Node( "node3" ) );

        state.equalString(
            "node1[ 1 ][ 2 ].node2.node3",
            ObjectPaths.format( path, ObjectPaths.DOT_FORMATTER ) );
        
        state.equalString(
            "node1[ 1 ][ 2 ]/node2/node3",
            ObjectPaths.format( path, ObjectPaths.SLASH_FORMATTER ) );
    }

    private
    < V >
    void
    assertRoot( ObjectPath< V > rootExpct,
                ObjectPath< V > path )
    {
        state.equal( rootExpct, ObjectPaths.rootOf( path ) );
    }

    @Test
    private
    void
    testPathRootIdentities()
    {
        ObjectPath< String > root = ObjectPath.newRoot();

        assertRoot( root, root );
        assertRoot( root, root.descend( "p1" ) );
        assertRoot( root, root.descend( "p1" ).descend( "p2" ) );

        assertRoot( root, 
            root.descend( "p1" ).startImmutableList().next().next() );

        // Seems kind of trivial but useful just for code coverage to assert
        // that newRoot() above really does in fact return a new root
        state.isFalse(
            ObjectPaths.rootOf( ObjectPath.< String >getRoot( "p1" ) ).
                equals( root ) 
        );
    }

    @Test
    private
    void
    testStartListWithIndex()
    {
        ImmutableListPath< Node > p = 
            ObjectPath.< Node >getRoot().
                descend( new Node( "a" ) ).
                descend( new Node( "b" ) ).
                startImmutableList( 5 );
        
        assertFormat( "/a/b[ 5 ]", p );
        assertFormat( "/a/b[ 6 ]", p.next() );
    }
}
