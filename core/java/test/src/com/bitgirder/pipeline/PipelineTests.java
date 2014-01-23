package com.bitgirder.pipeline;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.test.Test;

@Test
public
final
class PipelineTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    @Test
    public
    void
    testPipelineEmpty()
        throws Exception
    {
        Pipeline< Object > p = new Pipelines.Builder< Object >().build();
        state.equalInt( 0, p.size() );
        state.isTrue( Pipelines.lastElementOfType( p, String.class ) == null );
    }

    private
    static
    abstract
    class ElementType
    {
        final int val;

        private ElementType( int val ) { this.val = val; }

        void
        assertEquals( ElementType et )
        {
            state.equal( getClass(), et.getClass() );
            state.equalInt( val, et.val );
        }
    }

    private
    final
    static
    class ElementType1
    extends ElementType
    {
        private ElementType1( int val ) { super( val ); }
    }

    private
    final
    static
    class ElementType2
    extends ElementType
    {
        private ElementType2( int val ) { super( val ); }
    }

    private
    final
    static
    class ElementType3
    extends ElementType
    implements PipelineInitializer< ElementType >
    {
        private ElementType3( int val ) { super( val ); }

        public
        void
        initialize( PipelineInitializationContext< ElementType > ctx )
            throws Exception
        {
            ctx.addElement( new ElementType2( val ) );
        }
    }

    private
    final
    static
    class ElementType4
    extends ElementType
    implements PipelineInitializer< ElementType >
    {
        private ElementType4( int val ) { super( val ); }

        public
        void
        initialize( PipelineInitializationContext< ElementType > ctx )
            throws Exception
        {
            ctx.addElement( new ElementType3( val ) );
        }
    }

    private
    void
    assertEquals( Pipeline< ElementType > p,
                  int idx,
                  ElementType expct )
    {
        expct.assertEquals( p.get( idx ) );
    }

    @Test
    public
    void
    testPipelineBasic()
        throws Exception
    {
        Pipelines.Builder< ElementType > b = 
            new Pipelines.Builder< ElementType >();
 
        b.addElement( new ElementType1( 1 ) );
        b.addElement( new ElementType3( 1 ) );
        b.addElement( new ElementType4( 2 ) );

        Pipeline< ElementType > p = b.build();

        state.equalInt( 6, p.size() );
        assertEquals( p, 0, new ElementType1( 1 ) );
        assertEquals( p, 1, new ElementType2( 1 ) );
        assertEquals( p, 2, new ElementType3( 1 ) );
        assertEquals( p, 3, new ElementType2( 2 ) );
        assertEquals( p, 4, new ElementType3( 2 ) );
        assertEquals( p, 5, new ElementType4( 2 ) );
    }
}
