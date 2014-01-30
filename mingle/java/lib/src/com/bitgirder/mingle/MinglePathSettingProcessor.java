package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPaths;
import com.bitgirder.lang.path.MutableListPath;
import com.bitgirder.lang.path.ListPath;
import com.bitgirder.lang.path.DictionaryPath;

import com.bitgirder.pipeline.PipelineInitializationContext;
import com.bitgirder.pipeline.PipelineInitializer;

import java.util.Deque;

public
final
class MinglePathSettingProcessor
implements MingleValueReactorPipeline.Processor,
           PipelineInitializer< Object >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private ObjectPath< MingleIdentifier > path;

    // true between arrival of START_LIST and completion of the first list value
    private boolean awaitingList0;

    private final Deque< MingleValueReactorEvent.Type > endTypes = 
        Lang.newDeque();

    private 
    MinglePathSettingProcessor( ObjectPath< MingleIdentifier > start )
    {   
        this.path = start;
    }

    public ObjectPath< MingleIdentifier > path() { return path; }

    public
    void
    initialize( PipelineInitializationContext< Object > ctx )
    {
        MingleValueReactors.ensureStructuralCheck( ctx );
    }

    private 
    void 
    pathPop() 
    { 
        path = path.getParent(); 
        if ( path != null && path.isEmpty() ) path = null;
    }

    private
    void
    updateList()
    {
        if ( awaitingList0 ) {
            if ( path == null ) path = ObjectPath.getRoot();
            path = path.startMutableList();
            awaitingList0 = false;
        } else {
            if ( path instanceof MutableListPath ) {
                ( (MutableListPath< ? >) path ).increment();
            }
        }
    }

    private void prepareValue() { updateList(); }

    private
    void
    prepareStartList( MingleValueReactorEvent ev )
    {
        // The call to updateList() would be updating the containing list, not
        // the one being started.
        updateList();

        endTypes.push( ev.type() );

        // Set awaitingList0 for this list being started
        awaitingList0 = true;
    }

    private
    void
    prepareStartField( MingleValueReactorEvent ev )
    {
        endTypes.push( ev.type() );
    
        MingleIdentifier fld = ev.field();
        path = path == null ? ObjectPath.getRoot( fld ) : path.descend( fld );
    }

    private
    void
    prepareStartStruct( MingleValueReactorEvent ev )
    {
        endTypes.push( ev.type() );
        prepareValue();
    }

    private
    void
    prepareStartMap( MingleValueReactorEvent ev )
    {
        endTypes.push( ev.type() );
        prepareValue();
    }

    private
    void
    prepareEvent( MingleValueReactorEvent ev )
    {
        codef( "in prepareEvent, awaitingList0: %s, path type: %s",
            awaitingList0, path == null ? null : path.getClass().getSimpleName()
        );

        switch ( ev.type() ) {
        case VALUE: prepareValue(); break;
        case START_STRUCT: prepareStartStruct( ev ); break;
        case START_MAP: prepareStartMap( ev ); break;
        case START_LIST: prepareStartList( ev ); break;
        case START_FIELD: prepareStartField( ev ); break;
        case END: break;
        default: state.failf( "unhandled event: %s", ev.type() );
        }

        ev.setPath( path );
    }

    private
    void
    valueCompleted()
    {
        if ( endTypes.isEmpty() ) return;

        if ( endTypes.peek() == MingleValueReactorEvent.Type.START_FIELD ) {
            endTypes.pop();
            pathPop();
        }
//        if ( path instanceof DictionaryPath ) pathPop();
    }

    private
    void
    endCompleted()
    {
        MingleValueReactorEvent.Type evTyp = endTypes.pop();

        switch ( evTyp ) {
        case START_FIELD: pathPop(); break;
        case START_LIST: 
            pathPop();
            valueCompleted();
            break;
        case START_STRUCT: valueCompleted(); break;
        case START_MAP: valueCompleted(); break;
        default: state.failf( "unexpected end type: %s", evTyp );
        }
    }

    private
    void
    eventProcessed( MingleValueReactorEvent ev )
    {
        switch ( ev.type() ) {
        case VALUE: valueCompleted(); break;
        case END: endCompleted(); break;
        }
    }

    public
    void
    processPipelineEvent( MingleValueReactorEvent ev,
                          MingleValueReactor next )
        throws Exception
    {
        prepareEvent( ev );
        next.processEvent( ev );
        eventProcessed( ev );
    }

    public
    static
    MinglePathSettingProcessor
    create()
    {
        return new MinglePathSettingProcessor( null );
    }

    public
    static
    MinglePathSettingProcessor
    create( ObjectPath< MingleIdentifier > start )
    {
        inputs.notNull( start, "start" );
        return new MinglePathSettingProcessor( start );
    }
}
