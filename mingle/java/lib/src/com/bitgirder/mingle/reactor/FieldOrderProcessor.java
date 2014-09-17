package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.mingle.MingleIdentifier;
import com.bitgirder.mingle.QualifiedTypeName;
import com.bitgirder.mingle.MingleMissingFieldsException;

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
class FieldOrderProcessor
implements MingleReactorPipeline.Processor,
           PipelineInitializer< Object >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final OrderGetter og;

    private final Deque< MingleReactor > stack = Lang.newDeque();

    private FieldOrderProcessor( OrderGetter og ) { this.og = og; }

    public
    void
    initialize( PipelineInitializerContext< Object > ctx )
    {
        MingleReactors.ensureStructuralCheck( ctx );
    }

    public
    static
    interface OrderGetter
    {
        // can return null to indicate that there is no specific order for type
        public
        FieldOrder
        fieldOrderFor( QualifiedTypeName type );
    }

    private
    static
    ObjectPath< MingleIdentifier >
    pathForEvent( MingleReactorEvent ev )
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
    listAddEvent( List< MingleReactorEvent > l,
                  MingleReactorEvent ev )
    {
        l.add( ev.copy( false ) );
    }

    private
    final
    static
    class ListAcc
    implements MingleReactor
    {
        private final List< MingleReactorEvent > l;

        private ListAcc( List< MingleReactorEvent > l ) { this.l = l; }

        public
        void
        processEvent( MingleReactorEvent ev )
        {
            listAddEvent( l, ev );
        }
    }

    private
    final
    static
    class StructAcc
    implements MingleReactor
    {
        // never null, may be the empty path
        private ObjectPath< MingleIdentifier > startPath;

        // never null, may be the downstream reactor or another impl in this
        // class
        private MingleReactor next;

        // never null
        private FieldOrder order;

        // fieldQueue, saved, specs, and requiredRemaining are lazily
        // instantiated upon arrival of this instance's STRUCT_START event

        private Queue< MingleIdentifier > fieldQueue;

        private Set< MingleIdentifier > requiredRemaining;

        private Map< MingleIdentifier, List< MingleReactorEvent > > saved;

        private Map< MingleIdentifier, FieldOrder.FieldSpecification >
            specs;

        // set in FIELD_START, cleared after field value
        private MingleIdentifier curFld;
        private List< MingleReactorEvent > curAcc;

        // lazily instantiated on the event that we send saved field values
        private PathSettingProcessor pathSetter;

        private
        void
        processStartStruct( MingleReactorEvent ev )
            throws Exception
        {
            startPath = pathForEvent( ev );

            saved = Lang.newMap();
            fieldQueue = Lang.newQueue();
            specs = Lang.newMap();
            requiredRemaining = Lang.newSet();

            for ( FieldOrder.FieldSpecification spec : order.fields() ) {
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
        processStartField( MingleReactorEvent ev )
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
        MingleReactor
        fieldReactor()
        {
            return curAcc == null ? next : new ListAcc( curAcc );
        }

        public
        void
        processEvent( MingleReactorEvent ev )
            throws Exception
        {
            switch ( ev.type() ) {
            case STRUCT_START: processStartStruct( ev ); break;
            case FIELD_START: processStartField( ev ); break;
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
                        List< MingleReactorEvent > evs )
            throws Exception
        {
            if ( pathSetter == null ) {
                pathSetter = PathSettingProcessor.create( startPath );
            }

            MingleReactorEvent fsEv = new MingleReactorEvent();
            fsEv.setStartField( fld );
            pathSetter.processPipelineEvent( fsEv, next );

            for ( MingleReactorEvent ev : evs ) {
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
                List< MingleReactorEvent > evs = saved.remove( fld );

                if ( evs == null ) {
                    if ( isFinal && isOptional( fld ) ) {
                        fieldQueue.remove();
                        continue;
                    } else return;
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
        processValue( MingleReactorEvent ev )
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
        endStruct( MingleReactorEvent ev )
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
    push( MingleReactor rct,
          MingleReactorEvent ev )
        throws Exception
    {
        stack.push( rct );
        rct.processEvent( ev );
    }

    private
    void
    processValue( MingleReactorEvent ev )
        throws Exception
    {
        MingleReactor rct = stack.peek();

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
    processStartContainer( MingleReactorEvent ev,
                           MingleReactor next )
        throws Exception
    {
        if ( stack.peek() instanceof StructAcc ) {
            push( ( (StructAcc) stack.peek() ).fieldReactor(), ev );
        } else {
            MingleReactor rct = stack.isEmpty() ? next : stack.peek();
            push( rct, ev );
        }
    }

    private
    MingleReactor
    getStructAccNext( MingleReactor next )
    {
        if ( stack.isEmpty() ) return next;

        if ( stack.peek() instanceof StructAcc ) {
            return ( (StructAcc) stack.peek() ).fieldReactor();
        } else {
            return stack.peek();
        }
    }

    private
    void
    processStartStruct( MingleReactorEvent ev,
                        MingleReactor next )
        throws Exception
    {
        FieldOrder ord = og.fieldOrderFor( ev.structType() );

        if ( ord == null ) {
            processStartContainer( ev, next );
            return;
        }
 
        StructAcc acc = new StructAcc();
        acc.order = ord;
        acc.next = getStructAccNext( next );

        push( acc, ev );
    }

    private
    void
    processEnd( MingleReactorEvent ev )
        throws Exception
    {
        MingleReactor rct = stack.pop();

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
    processEvent( MingleReactorEvent ev,
                  MingleReactor next )
        throws Exception
    {
        switch ( ev.type() ) {
        case LIST_START: processStartContainer( ev, next ); break;
        case MAP_START: processStartContainer( ev, next ); break;
        case STRUCT_START: processStartStruct( ev, next ); break;
        case VALUE: processValue( ev ); break;
        case FIELD_START: stack.peek().processEvent( ev ); break;
        case END: processEnd( ev ); break; 
        default: state.failf( "unhandled event: %s", ev.type() );
        }
    }

    public
    void
    processPipelineEvent( MingleReactorEvent ev,
                          MingleReactor next )
        throws Exception
    {
        if ( stack.isEmpty() && 
             ev.type() != MingleReactorEvent.Type.STRUCT_START ) 
        {
            next.processEvent( ev );
            return;
        }

        processEvent( ev, next );
    }

    public
    static
    FieldOrderProcessor
    create( OrderGetter og )
    {
        inputs.notNull( og, "og" );
        return new FieldOrderProcessor( og );
    }
}
