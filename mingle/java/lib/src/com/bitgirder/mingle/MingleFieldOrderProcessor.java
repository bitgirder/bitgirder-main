package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.pipeline.PipelineInitializer;
import com.bitgirder.pipeline.PipelineInitializerContext;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPaths;

import java.util.Deque;
import java.util.Queue;
import java.util.List;
import java.util.Map;
import java.util.Set;

public
final
class MingleFieldOrderProcessor
implements MingleValueReactorPipeline.Processor,
           PipelineInitializer< Object >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final OrderGetter og;

    private final Deque< MingleValueReactor > stack = Lang.newDeque();

    private MingleFieldOrderProcessor( OrderGetter og ) { this.og = og; }

    public
    void
    initialize( PipelineInitializerContext< Object > ctx )
    {
        MingleValueReactors.ensurePathSetter( ctx );
    }

    public
    static
    interface OrderGetter
    {
        // can return null to indicate that there is no specific order for type
        public
        MingleValueReactorFieldOrder
        fieldOrderFor( QualifiedTypeName type );
    }

    private
    static
    ObjectPath< MingleIdentifier >
    pathForEvent( MingleValueReactorEvent ev )
    {
        if ( ev.path() == null ) {
            return ObjectPath.getRoot();
        } else {
            return ObjectPaths.asImmutableCopy( ev.path() );
        }
    }

    private
    static
    void
    listAddEvent( List< MingleValueReactorEvent > l,
                  MingleValueReactorEvent ev )
    {
        l.add( ev.copy( false ) );
    }

    private
    final
    static
    class ListAcc
    implements MingleValueReactor
    {
        private final List< MingleValueReactorEvent > l;

        private ListAcc( List< MingleValueReactorEvent > l ) { this.l = l; }

        public
        void
        processEvent( MingleValueReactorEvent ev )
        {
            listAddEvent( l, ev );
        }
    }

    private
    final
    static
    class StructAcc
    implements MingleValueReactor
    {
        // never null, may be the empty path
        private ObjectPath< MingleIdentifier > startPath;

        // never null, may be the downstream reactor or another impl in this
        // class
        private MingleValueReactor next;

        // never null
        private MingleValueReactorFieldOrder order;

        // fieldQueue, saved, specs, and requiredRemaining are lazily
        // instantiated upon arrival of this instance's START_STRUCT event

        private Queue< MingleIdentifier > fieldQueue;

        private Set< MingleIdentifier > requiredRemaining;

        private Map< MingleIdentifier, List< MingleValueReactorEvent > > saved;

        private Map< MingleIdentifier, MingleValueReactorFieldSpecification >
            specs;

        // set in START_FIELD, cleared after field value
        private MingleIdentifier curFld;
        private List< MingleValueReactorEvent > curAcc;

        // lazily instantiated on the event that we send saved field values
        private MinglePathSettingProcessor pathSetter;

        private
        void
        processStartStruct( MingleValueReactorEvent ev )
            throws Exception
        {
            startPath = pathForEvent( ev );

            saved = Lang.newMap();
            fieldQueue = Lang.newQueue();
            specs = Lang.newMap();
            requiredRemaining = Lang.newSet();

            for ( MingleValueReactorFieldSpecification spec : order.fields() ) {
                fieldQueue.add( spec.field() );
                specs.put( spec.field(), spec );
                if ( spec.required() ) requiredRemaining.add( spec.field() );
            }

            next.processEvent( ev );
        }

        private
        boolean
        shouldAccumulate( MingleIdentifier fld )
        {
            if ( fieldQueue.isEmpty() ) return false;
            
            if ( fieldQueue.peek().equals( fld ) ) {
                fieldQueue.remove();
                return false;
            }

            return specs.containsKey( fld );
        }

        private
        void
        processStartField( MingleValueReactorEvent ev )
            throws Exception
        {
            state.isTruef( curFld == null, 
                "saw field '%s' while curFld is '%s'", ev.field(), curFld );
 
            curFld = ev.field();
            requiredRemaining.remove( curFld );

            if ( shouldAccumulate( curFld ) ) {
                curAcc = Lang.newList();
            } else {
                next.processEvent( ev );
            }
        }

        private
        MingleValueReactor
        fieldReactor()
        {
            return curAcc == null ? next : new ListAcc( curAcc );
        }

        public
        void
        processEvent( MingleValueReactorEvent ev )
            throws Exception
        {
            switch ( ev.type() ) {
            case START_STRUCT: processStartStruct( ev ); break;
            case START_FIELD: processStartField( ev ); break;
            default: state.failf( "unhandled struct event: %s", ev.type() );
            }
        }

        private
        boolean
        isOptional( MingleIdentifier fld )
        {
            return ! specs.get( fld ).required();
        }

        private
        void
        sendReadyField( MingleIdentifier fld,
                        List< MingleValueReactorEvent > evs )
            throws Exception
        {
            if ( pathSetter == null ) {
                pathSetter = MinglePathSettingProcessor.create( startPath );
            }

            MingleValueReactorEvent fsEv = new MingleValueReactorEvent();
            fsEv.setStartField( fld );
            pathSetter.processPipelineEvent( fsEv, next );

            for ( MingleValueReactorEvent ev : evs ) {
                pathSetter.processPipelineEvent( ev, next );
            }
        }

        private
        void
        sendReadyFields( boolean isFinal )
            throws Exception
        {
            while ( ! fieldQueue.isEmpty() )
            {
                MingleIdentifier fld = fieldQueue.peek();
                List< MingleValueReactorEvent > evs = saved.remove( fld );

                if ( evs == null ) {
                    if ( isFinal && isOptional( fld ) ) {
                        fieldQueue.remove();
                        continue;
                    } else {
                        return;
                    }
                }

                // evs != null if we get here
                sendReadyField( fieldQueue.remove(), evs );
            }
        }

        private
        void
        completeField()
            throws Exception
        {
            state.isFalse( curFld == null, 
                "completeField() called but curFld is null" );

            if ( curAcc != null ) saved.put( curFld, curAcc );
            curAcc = null;
            curFld = null;

            sendReadyFields( false );
        }

        private
        void
        valueEnded()
            throws Exception
        {
            if ( curFld != null ) completeField();
        }

        private
        void
        processValue( MingleValueReactorEvent ev )
            throws Exception
        {
            if ( curAcc == null ) {
                next.processEvent( ev );
            } else {
                listAddEvent( curAcc, ev );
            }
            
            completeField();
        }

        private
        void
        endStruct( MingleValueReactorEvent ev )
            throws Exception
        {
            sendReadyFields( true );

            if ( ! requiredRemaining.isEmpty() ) 
            {
                throw new MingleMissingFieldsException(
                    requiredRemaining, ev.path() );
            }

            next.processEvent( ev );              
        }
    }

    private
    void
    push( MingleValueReactor rct,
          MingleValueReactorEvent ev )
        throws Exception
    {
        stack.push( rct );
        rct.processEvent( ev );
    }

    private
    void
    processValue( MingleValueReactorEvent ev )
        throws Exception
    {
        MingleValueReactor rct = stack.peek();

        if ( rct instanceof StructAcc ) {
            ( (StructAcc) rct ).processValue( ev );
        } else {
            rct.processEvent( ev );
        }
    }

    // called for list/map or a struct with no order. impl: we just re-push the
    // current stack top to continue accumulating the container
    private
    void
    processStartContainer( MingleValueReactorEvent ev,
                           MingleValueReactor next )
        throws Exception
    {
        if ( stack.peek() instanceof StructAcc ) {
            push( ( (StructAcc) stack.peek() ).fieldReactor(), ev );
        } else {
            MingleValueReactor rct = stack.isEmpty() ? next : stack.peek();
            push( rct, ev );
        }
    }

    private
    void
    structAccSetNext( StructAcc acc,
                      MingleValueReactor next )
    {
        if ( stack.isEmpty() ) {
            acc.next = next;
        } else {
            if ( stack.peek() instanceof StructAcc ) {
                acc.next = ( (StructAcc) stack.peek() ).fieldReactor();
            } else {
                acc.next = stack.peek();
            }
        }
    }

    private
    void
    processStartStruct( MingleValueReactorEvent ev,
                        MingleValueReactor next )
        throws Exception
    {
        MingleValueReactorFieldOrder ord = og.fieldOrderFor( ev.structType() );

        if ( ord == null ) {
            processStartContainer( ev, next );
            return;
        }
 
        StructAcc acc = new StructAcc();
        acc.order = ord;
        structAccSetNext( acc, next );

        push( acc, ev );
    }

    private
    void
    processEnd( MingleValueReactorEvent ev )
        throws Exception
    {
        MingleValueReactor rct = stack.pop();

        if ( rct instanceof StructAcc ) {
            ( (StructAcc) rct ).endStruct( ev );
        } else {
            rct.processEvent( ev );
        }

        if ( stack.peek() instanceof StructAcc ) {
            ( (StructAcc) stack.peek() ).valueEnded();
        }
    }

    private
    void
    processEvent( MingleValueReactorEvent ev,
                  MingleValueReactor next )
        throws Exception
    {
        switch ( ev.type() ) {
        case START_LIST: processStartContainer( ev, next ); break;
        case START_MAP: processStartContainer( ev, next ); break;
        case START_STRUCT: processStartStruct( ev, next ); break;
        case VALUE: processValue( ev ); break;
        case START_FIELD: stack.peek().processEvent( ev ); break;
        case END: processEnd( ev ); break; 
        default: state.failf( "unhandled event: %s", ev.type() );
        }
    }

    public
    void
    processPipelineEvent( MingleValueReactorEvent ev,
                          MingleValueReactor next )
        throws Exception
    {
        codef( "processing pipeline event: %s", ev.inspect() );

        if ( stack.isEmpty() && 
                ev.type() != MingleValueReactorEvent.Type.START_STRUCT ) 
        {
            next.processEvent( ev );
            return;
        }

        processEvent( ev, next );
    }

    public
    static
    MingleFieldOrderProcessor
    create( OrderGetter og )
    {
        inputs.notNull( og, "og" );
        return new MingleFieldOrderProcessor( og );
    }
}
