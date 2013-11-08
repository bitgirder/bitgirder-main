package com.bitgirder.lang.path;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
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

    private
    < V >
    void
    assertPathsEq( ObjectPath< V > p1,
                   ObjectPath< V > p2 )
    {
        ObjectPaths.areEqual( p1, p2 );
        ObjectPaths.areEqual( p2, p1 );

        ObjectPaths.areEqual( p1, p1 ); 
        ObjectPaths.areEqual( p2, p2 ); 
    }

    private
    < V >
    void
    assertPathsNeq( ObjectPath< V > p1,
                    ObjectPath< V > p2 )
    {
        boolean res = ObjectPaths.areEqual( p1, p2 );
        state.isFalse( ObjectPaths.areEqual( p1, p2 ) );
        state.isFalse( ObjectPaths.areEqual( p2, p1 ) );
    }

    @Test
    private
    void
    testPathEquals()
    {
        ObjectPath< String > p1 = ObjectPath.getRoot();
        ObjectPath< String > p2 = ObjectPath.getRoot();
        assertPathsEq( p1, p2 );
        p2 = p2.descend( "n1" );
        assertPathsNeq( p1, p2 );
        p1 = p1.descend( "n1" );
        assertPathsEq( p1, p2 );
        p1 = p1.descend( "n2" ).startImmutableList().next().next();
        assertPathsNeq( p1, p2 );
        state.isFalse( p2.equals( p1 ) );
        p2 = p2.descend( "n2" ).startImmutableList().next();
        assertPathsNeq( p1, p2 );
        ImmutableListPath< String > p2Cast = Lang.castUnchecked( p2 );
        p2 = p2Cast.next();
        assertPathsEq( p1, p2 );
    }
}
