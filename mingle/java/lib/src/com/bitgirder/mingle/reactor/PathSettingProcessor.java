package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPaths;
import com.bitgirder.lang.path.MutableListPath;

import com.bitgirder.pipeline.PipelineInitializerContext;
import com.bitgirder.pipeline.PipelineInitializer;

import com.bitgirder.mingle.MingleIdentifier;

import java.util.Deque;

public
final
class PathSettingProcessor
implements MingleReactorPipeline.Processor,
           PipelineInitializer< Object >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final ObjectPath< MingleIdentifier > startPath;

    private final Deque< Object > stack = Lang.newDeque();

    private 
    PathSettingProcessor( ObjectPath< MingleIdentifier > startPath )
    {   
        this.startPath = startPath;
    }

    private
    final
    static
    class ListContext
    {
        private ObjectPath< MingleIdentifier > basePath;
        private MutableListPath< MingleIdentifier > path;
    }

    private
    final
    static
    class MapContext
    {
        private ObjectPath< MingleIdentifier > path;
    }

    private
    final
    static
    class FieldContext
    {
        private ObjectPath< MingleIdentifier > path;
    }

    public 
    ObjectPath< MingleIdentifier > 
    path() 
    { 
        if ( stack.isEmpty() ) return startPath;

        Object top = stack.peek();

        if ( top instanceof FieldContext ) {
            return ( (FieldContext) top ).path;
        } else if ( top instanceof MapContext ) {
            return ( (MapContext) top ).path;
        } else if ( top instanceof ListContext ) {
            ListContext lc = (ListContext) top;
            return lc.path == null ? lc.basePath : lc.path;
        } else {
            throw state.failf( "unhandled stack element: %s", top );
        }
    }

    public
    void
    initialize( PipelineInitializerContext< Object > ctx )
    {
        MingleReactors.ensureStructuralCheck( ctx );
    }

    private
    void
    updateList()
    {
        Object top = stack.peek();
        if ( top == null || ( ! ( top instanceof ListContext ) ) ) return;

        ListContext lc = (ListContext) top;

        if ( lc.path == null ) {
            if ( lc.basePath == null ) {
                ObjectPath< MingleIdentifier > root = ObjectPath.getRoot();
                lc.path = root.startMutableList();
            } else {
                lc.path = lc.basePath.startMutableList();
            }
        } else lc.path.increment();
    }

    private void prepareValue() { updateList(); }

    private
    void
    prepareListStart()
    {
        prepareValue();

        ListContext lc = new ListContext();
        lc.basePath = path();
        stack.push( lc );
    }

    private
    void
    prepareStartField( MingleReactorEvent ev )
    {
        FieldContext fc = new FieldContext();
        fc.path = ObjectPaths.descend( path(), ev.field() );
        stack.push( fc );
    }

    private
    void
    prepareStructure()
    {
        prepareValue();

        MapContext mc = new MapContext();
        mc.path = path();

        stack.push( mc );
    }

    private
    void
    prepareEnd()
    {
        if ( stack.peek() instanceof ListContext ) stack.pop();
    }

    private
    void
    prepareEvent( MingleReactorEvent ev )
    {
        switch ( ev.type() ) {
        case VALUE: prepareValue(); break;
        case STRUCT_START: prepareStructure(); break;
        case MAP_START: prepareStructure(); break;
        case LIST_START: prepareListStart(); break;
        case FIELD_START: prepareStartField( ev ); break;
        case END: prepareEnd(); break;
        default: state.failf( "unhandled event: %s", ev.type() );
        }

        ev.setPath( path() );
    }
 
    private
    void
    processedValue()
    {
        if ( stack.peek() instanceof FieldContext ) stack.pop();
    }

    private
    void
    processedEnd()
    {
        Object top = stack.peek();

        if ( top == null ) return;

        if ( top instanceof MapContext ) stack.pop();

        processedValue();
    }

    private
    void
    processedEvent( MingleReactorEvent ev )
    {
        switch ( ev.type() ) {
        case VALUE: processedValue(); break;
        case END: processedEnd(); break;
        }
    }

    public
    void
    processPipelineEvent( MingleReactorEvent ev,
                          MingleReactor next )
        throws Exception
    {
        prepareEvent( ev );
        next.processEvent( ev );
        processedEvent( ev );
    }

    public
    static
    PathSettingProcessor
    create()
    {
        return new PathSettingProcessor( null );
    }

    public
    static
    PathSettingProcessor
    create( ObjectPath< MingleIdentifier > start )
    {
        inputs.notNull( start, "start" );

        start = ObjectPaths.asMutableCopy( start );
        codef( "start path instanceof: %s", start.getClass() );
        
        return new PathSettingProcessor( start );
    }
}
