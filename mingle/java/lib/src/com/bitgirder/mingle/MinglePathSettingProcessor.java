package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPaths;
import com.bitgirder.lang.path.MutableListPath;
import com.bitgirder.lang.path.ListPath;
import com.bitgirder.lang.path.DictionaryPath;

import com.bitgirder.pipeline.PipelineInitializationContext;
import com.bitgirder.pipeline.PipelineInitializer;

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
    prepareStartList()
    {
        // The call to updateList() would be updating the containing list, not
        // the one being started.
        updateList();

        // Set awaitingList0 for this list being started
        awaitingList0 = true;
    }

    private
    void
    prepareStartField( MingleIdentifier fld )
    {
        path = path == null ? ObjectPath.getRoot( fld ) : path.descend( fld );
    }

    private
    void
    prepareEnd()
    {
//        if ( path instanceof ListPath ) pathPop();
    }

    private
    void
    prepareEvent( MingleValueReactorEvent ev )
    {
        switch ( ev.type() ) {
        case VALUE: prepareValue(); break;
        case START_STRUCT: prepareValue(); break;
        case START_MAP: prepareValue(); break;
        case START_LIST: prepareStartList(); break;
        case START_FIELD: prepareStartField( ev.field() ); break;
        case END: prepareEnd(); break;
        default: state.failf( "unhandled event: %s", ev.type() );
        }

        ev.setPath( path );
    }

//    // if we were accumulating a field value we pop the field from the path;
//    // otherwise there is nothing to do (end of list will have popped list
//    // earlier in prepareEnd())
    private
    void
    valueCompleted()
    {
        if ( path instanceof DictionaryPath ) pathPop();
    }

    private
    void
    endCompleted()
    {
        if ( path instanceof DictionaryPath || path instanceof ListPath ) {
            pathPop();
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
