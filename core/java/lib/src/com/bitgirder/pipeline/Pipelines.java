package com.bitgirder.pipeline;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.ArrayList;

public
final
class Pipelines
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private Pipelines() {};

    private
    final
    static
    class DefaultPipeline< V >
    implements Pipeline< V >
    {
        private final ArrayList< V > l;

        private DefaultPipeline( ArrayList< V > l ) { this.l = l; }

        public V get( int idx ) { return l.get( idx ); }

        public int size() { return l.size(); }
    }

    private
    final
    static
    class InitializerImpl< V >
    implements PipelineInitializationContext< V >
    {
        private final DefaultPipeline< V > pip = 
            new DefaultPipeline< V >( new ArrayList< V >() );

        public Pipeline< V > pipeline() { return pip; }

        public 
        void 
        addElement( V elt ) 
            throws Exception
        { 
            inputs.notNull( elt, "elt" );

            if ( elt instanceof PipelineInitializer ) 
            {
                PipelineInitializer< V > pi = Lang.castUnchecked( elt );
                pi.initialize( this );
            }
            
            pip.l.add( elt );
        }
    }

    public
    final
    static
    class Builder< V >
    {
        private final InitializerImpl< V > init = new InitializerImpl< V >();

        // Behaves similarly to PipelineInitializationContext.addElement(); see
        // that method for a note about what happens when elt implemnts
        // PipelineInitializer
        public
        Builder< V >
        addElement( V elt )
            throws Exception
        {
            init.addElement( inputs.notNull( elt, "elt" ) );
            return this;
        }

        public
        Pipeline< V >
        build()
            throws Exception
        {
            return new DefaultPipeline< V >( new ArrayList< V >( init.pip.l ) );
        }
    }

    public
    static
    < V, T extends V >
    T
    lastElementOfType( Pipeline< V > pip,
                       Class< T > cls )
    {
        inputs.notNull( pip, "pip" );
        inputs.notNull( cls, "cls" );
        
        for ( int i = pip.size() - 1; i >= 0; --i ) {
            V elt = pip.get( i );
            if ( cls.isInstance( elt ) ) return cls.cast( elt );
        }

        return null;
    }

    public
    static
    boolean
    isEmpty( Pipeline< ? > p )
    {
        inputs.notNull( p, "p" );
        return p.size() == 0;
    }
}
