package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.pipeline.Pipeline;
import com.bitgirder.pipeline.Pipelines;

import java.util.ArrayList;
import java.util.List;

public
final
class MingleReactorPipeline
implements MingleReactor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Pipeline< Object > pipeline;
    private final ElementImpl head;

    private
    MingleReactorPipeline( Pipeline< Object > pipeline,
                           ElementImpl head )
    {
        this.pipeline = pipeline;
        this.head = head;
    }

    public Pipeline< Object > pipeline() { return pipeline; }

    private
    static
    abstract
    class ElementImpl
    implements MingleReactor
    {
        MingleReactor next;
    }

    private
    static
    class ReactorElement
    extends ElementImpl
    {
        final MingleReactor rct;

        ReactorElement( MingleReactor rct ) { this.rct = rct; }

        public
        void
        processEvent( MingleReactorEvent ev )
            throws Exception
        {
            rct.processEvent( ev );
            next.processEvent( ev );
        }
    }

    public
    static
    interface Processor
    {
        public
        void
        processPipelineEvent( MingleReactorEvent ev,
                              MingleReactor next )
            throws Exception;
    }

    private
    final
    static
    class ProcessorElement
    extends ElementImpl
    {
        private Processor proc;

        private ProcessorElement( Processor proc ) { this.proc = proc; }

        public
        void
        processEvent( MingleReactorEvent ev )
            throws Exception
        {
            proc.processPipelineEvent( ev, next );
        }
    }

    public
    void
    processEvent( MingleReactorEvent ev )
        throws Exception
    {
        inputs.notNull( ev, "ev" );
        head.processEvent( ev );
    }

    public
    final
    static
    class Builder
    {
        private final Pipelines.Builder< Object > b =
            new Pipelines.Builder< Object >();

        public
        Builder
        addReactor( MingleReactor rct )
        {
            b.addElement( inputs.notNull( rct, "rct" ) );
            return this;
        }

        public
        Builder
        addProcessor( Processor proc )
        {
            b.addElement( inputs.notNull( proc, "proc" ) );
            return this;
        }

        private
        ElementImpl
        elementForObject( Object obj )
        {
            if ( obj instanceof MingleReactor ) {
                return new ReactorElement( (MingleReactor) obj );
            } else if ( obj instanceof Processor ) {
                return new ProcessorElement( (Processor) obj );
            }

            throw state.failf( "unhandled object: %s", obj );
        }

        private
        ElementImpl
        buildChain( Pipeline< Object > pip )
        {
            ElementImpl head = null;

            for ( int i = pip.size() - 1; i >= 0; --i ) 
            {
                MingleReactor next = head;
                head = elementForObject( pip.get( i ) );

                head.next = next == null ? 
                    MingleReactors.discardReactor() : next;
            }

            return head;
        }

        public
        MingleReactorPipeline
        build()
        {
            Pipeline< Object > pipeline = b.build();

            inputs.isFalse( 
                Pipelines.isEmpty( pipeline ), "pipeline is empty" );

            ElementImpl head = buildChain( pipeline );

            return new MingleReactorPipeline( pipeline, head );
        }
    }
}
