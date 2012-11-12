package com.bitgirder.etl;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.model.MingleIdentifiedName;

public
final
class EtlProcessors
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static ShutdownRequest NORMAL_SHUTDOWN_REQ = 
        new ShutdownRequest( false );

    private final static ShutdownRequest URGENT_SHUTDOWN_REQ =
        new ShutdownRequest( true );

    private EtlProcessors() {}

    public 
    final 
    static 
    class ShutdownRequest 
    { 
        private final boolean isUrgent;

        private 
        ShutdownRequest( boolean isUrgent ) 
        { 
            this.isUrgent = isUrgent; 
        }

        public boolean isUrgent() { return isUrgent; }
    }

    public 
    static 
    ShutdownRequest 
    getShutdownRequest( boolean isUrgent ) 
    { 
        return isUrgent ? URGENT_SHUTDOWN_REQ : NORMAL_SHUTDOWN_REQ; 
    }

    private
    abstract
    static
    class ProcessorOpImpl
    {
        private final MingleIdentifiedName id;

        // does input check for public subclasses
        private
        ProcessorOpImpl( MingleIdentifiedName id )
        {
            this.id = inputs.notNull( id, "id" );
        }

        public final MingleIdentifiedName getId() { return id; }
    }

    private
    abstract
    static
    class GetProcessorObject
    extends ProcessorOpImpl
    {
        private GetProcessorObject( MingleIdentifiedName id ) { super( id ); }
    }

    private
    abstract
    static
    class SetProcessorObject
    extends ProcessorOpImpl
    {
        private final Object obj;

        private
        SetProcessorObject( MingleIdentifiedName id,
                            Object obj )
        {
            super( id );
            this.obj = obj;
        }

        public final Object getObject() { return obj; }
    }

    public
    final
    static
    class GetProcessorState
    extends GetProcessorObject
    {
        private GetProcessorState( MingleIdentifiedName id ) { super( id ); }
    }

    public
    static
    GetProcessorState
    createGetProcessorState( MingleIdentifiedName id )
    {
        inputs.notNull( id, "id" );
        return new GetProcessorState( id );
    }

    public
    final
    static
    class SetProcessorState
    extends SetProcessorObject
    {
        private
        SetProcessorState( MingleIdentifiedName id,
                           Object procState )
        {
            super( id, procState );
        }
    }

    public
    static
    SetProcessorState
    createSetProcessorState( MingleIdentifiedName id,
                             Object procState )
    {
        inputs.notNull( procState, "procState" );
        return new SetProcessorState( id, procState );
    }

    public
    final
    static
    class GetProcessorFeedPosition
    extends GetProcessorObject
    {
        private
        GetProcessorFeedPosition( MingleIdentifiedName id )
        {
            super( id );
        }
    }

    public
    static
    GetProcessorFeedPosition
    createGetProcessorFeedPosition( MingleIdentifiedName id )
    {
        inputs.notNull( id, "id" );
        return new GetProcessorFeedPosition( id );
    }

    public
    final
    static
    class SetProcessorFeedPosition
    extends SetProcessorObject
    {
        private
        SetProcessorFeedPosition( MingleIdentifiedName id,
                                  Object pos )
        {
            super( id, pos );
        }
    }

    public
    static
    SetProcessorFeedPosition
    createSetProcessorFeedPosition( MingleIdentifiedName id,
                                    Object pos )
    {
        inputs.notNull( id, "id" );
        inputs.notNull( pos, "pos" );

        return new SetProcessorFeedPosition( id, pos );
    }
}
