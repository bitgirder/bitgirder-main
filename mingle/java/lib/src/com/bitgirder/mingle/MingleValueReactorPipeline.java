package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.pipeline.Pipeline;
import com.bitgirder.pipeline.Pipelines;

import java.util.ArrayList;
import java.util.List;

public
final
class MingleValueReactorPipeline
implements MingleValueReactor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Pipeline< Object > pipeline;
    private final ElementImpl head;

    private
    MingleValueReactorPipeline( Pipeline< Object > pipeline,
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
    implements MingleValueReactor
    {
        MingleValueReactor next;
    }

    private
    static
    class ReactorElement
    extends ElementImpl
    {
        final MingleValueReactor rct;

        ReactorElement( MingleValueReactor rct ) { this.rct = rct; }

        public
        void
        processEvent( MingleValueReactorEvent ev )
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
        processPipelineEvent( MingleValueReactorEvent ev,
                              MingleValueReactor next )
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
        processEvent( MingleValueReactorEvent ev )
            throws Exception
        {
            proc.processPipelineEvent( ev, next );
        }
    }

    public
    void
    processEvent( MingleValueReactorEvent ev )
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
        addReactor( MingleValueReactor rct )
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
            if ( obj instanceof MingleValueReactor ) {
                return new ReactorElement( (MingleValueReactor) obj );
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
                MingleValueReactor next = head;
                head = elementForObject( pip.get( i ) );

                head.next = next == null ? 
                    MingleValueReactors.discardReactor() : next;
            }

            return head;
        }

        public
        MingleValueReactorPipeline
        build()
        {
            Pipeline< Object > pipeline = b.build();

            inputs.isFalse( 
                Pipelines.isEmpty( pipeline ), "pipeline is empty" );

            ElementImpl head = buildChain( pipeline );

            return new MingleValueReactorPipeline( pipeline, head );
        }
    }
}
