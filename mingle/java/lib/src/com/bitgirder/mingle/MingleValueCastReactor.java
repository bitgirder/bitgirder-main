package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.pipeline.PipelineInitializationContext;
import com.bitgirder.pipeline.PipelineInitializer;

import com.bitgirder.lang.Lang;

import java.util.Deque;

public
final
class MingleValueCastReactor
implements MingleValueReactorPipeline.Processor,
           PipelineInitializer< Object >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Deque< Object > stack = Lang.newDeque();

    private MingleValueCastReactor() {}

    public
    void
    initialize( PipelineInitializationContext< Object > ctx )
    {
        MingleValueReactors.ensureStructuralCheck( ctx );
        MingleValueReactors.ensurePathSetter( ctx );
    }

    private
    void
    push( Object obj )
    {
        if ( obj instanceof AtomicTypeReference ||
             obj instanceof NullableTypeReference ) 
        {
            stack.push( obj );
            return;
        }

        state.failf( "unhandled push value: %s", obj );
    }

    private
    void
    processAtomicValue( MingleValueReactorEvent ev,
                        AtomicTypeReference typ,
                        MingleValueReactor next )
    {
        throw new UnsupportedOperationException( "Unimplemented" );
    }

    private
    void
    processValue( MingleValueReactorEvent ev,
                  Object obj,
                  MingleValueReactor next )
    {
        if ( obj instanceof AtomicTypeReference ) {
            processAtomicValue( ev, (AtomicTypeReference) obj, next );
        } else {
            state.failf( "unhandled stack value: %s", obj );
        }
    } 

    public
    void
    processPipelineEvent( MingleValueReactorEvent ev,
                          MingleValueReactor next )
    {
        switch ( ev.type() ) {
        case VALUE: processValue( ev, stack.peek(), next ); return;
        }

        state.failf( "unhandled event: %s", ev.type() );
    }

    public
    static
    MingleValueCastReactor
    create( MingleTypeReference typ )
    {
        inputs.notNull( typ, "typ" );

        MingleValueCastReactor res = new MingleValueCastReactor();
        res.push( typ );

        return res;
    }
}
