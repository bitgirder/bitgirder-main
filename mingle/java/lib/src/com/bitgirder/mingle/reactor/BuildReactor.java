package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.mingle.MingleValue;
import com.bitgirder.mingle.MingleIdentifier;
import com.bitgirder.mingle.QualifiedTypeName;
import com.bitgirder.mingle.ListTypeReference;

import com.bitgirder.pipeline.PipelineInitializer;
import com.bitgirder.pipeline.PipelineInitializerContext;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import java.util.Deque;

public
final
class BuildReactor
implements MingleReactor,
           PipelineInitializer< Object >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    public
    static
    interface ValueProducer
    {
        public
        Object
        produceValue( ObjectPath< MingleIdentifier > path )
            throws Exception;
    }

    public
    static
    interface FieldSetBuilder
    extends ValueProducer
    {
        public
        Factory
        startField( MingleIdentifier fld,
                    ObjectPath< MingleIdentifier > path )
            throws Exception;
        
        public
        void
        setValue( MingleIdentifier fld,
                  Object val,
                  ObjectPath< MingleIdentifier > path )
            throws Exception;
    }

    public
    static
    interface ListBuilder
    extends ValueProducer
    {
        public
        void
        addValue( Object val,
                  ObjectPath< MingleIdentifier > path )
            throws Exception;

        public
        Factory
        nextFactory();
    }

    public
    static
    interface Factory
    {
        public
        Object
        buildValue( MingleValue mv,
                    ObjectPath< MingleIdentifier > path )
            throws Exception;
        
        public
        FieldSetBuilder
        startMap( ObjectPath< MingleIdentifier > path )
            throws Exception;

        public
        FieldSetBuilder
        startStruct( QualifiedTypeName typ,
                     ObjectPath< MingleIdentifier > path )
            throws Exception;
        
        public
        ListBuilder
        startList( ListTypeReference lt,
                   ObjectPath< MingleIdentifier > path )
            throws Exception;
    }

    private final Deque< Object > stack = Lang.newDeque();

    private boolean hasVal; // indicates whether val may be returned
    private Object val; // the built value

    private BuildReactor( Factory fact ) { stack.push( fact ); }

    public
    void
    initialize( PipelineInitializerContext< Object > ctx )
    {
        MingleReactors.ensureStructuralCheck( ctx );
        MingleReactors.ensurePathSetter( ctx );
    }
    
    public
    Object
    value()
    {
        state.isTrue( hasVal, "no value is built" );
        return val;
    }

    private
    void
    completeFieldValue( Object val,
                        MingleReactorEvent ev )
        throws Exception
    {
        MingleIdentifier fld = 
            state.cast( MingleIdentifier.class, stack.pop() );

        FieldSetBuilder fsb = state.cast( FieldSetBuilder.class, stack.peek() );

        fsb.setValue( fld, val, ev.path() );
    }

    private
    void
    completeValue( Object val,
                   MingleReactorEvent ev )
        throws Exception
    {
        if ( stack.isEmpty() ) {
            this.val = val;
            this.hasVal = true;
            return;
        }

        Object top = stack.peek();

        if ( top instanceof MingleIdentifier ) {
            completeFieldValue( val, ev );
        } else if ( top instanceof ListBuilder ) {
            ( (ListBuilder) top ).addValue( val, ev.path() );
        } else {
            state.failf( "unhandled value recipient: %s", top );
        }
    }

    private
    Factory
    nextFactory( MingleReactorEvent ev )
        throws Exception
    {
        Object top = stack.peek();

        if ( top instanceof Factory ) {
            stack.pop();
            return Lang.castUnchecked( top );
        } else if ( top instanceof ListBuilder ) {
            Factory f = ( (ListBuilder) top ).nextFactory();
            if ( f != null ) return f;
            throw new UnsupportedOperationException( "Unimplemented" );
//        return nil, failBuilderBadInput( ev, br.ErrorFactory )
        } else {
            throw state.failf( "unhandled stack elt: %s", top );
        }
    }

    private
    void
    processValue( MingleReactorEvent ev )
        throws Exception
    {
        Factory f = nextFactory( ev );
        Object val = f.buildValue( ev.value(), ev.path() );
        completeValue( val, ev );
    }

    private
    void
    processFieldStart( MingleReactorEvent ev )
        throws Exception
    {
        FieldSetBuilder fsb = state.cast( FieldSetBuilder.class, stack.peek() );
        Factory f = fsb.startField( ev.field(), ev.path() );

        stack.push( ev.field() );
        stack.push( f );
    }

    private void startFieldSet( FieldSetBuilder fsb ) { stack.push( fsb ); }

    private
    void
    processMapStart( MingleReactorEvent ev )
        throws Exception
    {
        Factory f = nextFactory( ev );
        startFieldSet( f.startMap( ev.path() ) );
    }

    private
    void
    processStructStart( MingleReactorEvent ev )
        throws Exception
    {
        Factory f = nextFactory( ev );
        startFieldSet( f.startStruct( ev.structType(), ev.path() ) );
    }

    private
    void
    processListStart( MingleReactorEvent ev )
        throws Exception
    {
        Factory f = nextFactory( ev );

        ListBuilder lb = f.startList( ev.listType(), ev.path() );
        stack.push( lb );
    }

    private
    void
    processEnd( MingleReactorEvent ev )
        throws Exception
    {
        ValueProducer vp = state.cast( ValueProducer.class, stack.pop() );
        completeValue( vp.produceValue( ev.path() ), ev );
    }

    public
    void
    processEvent( MingleReactorEvent ev )
        throws Exception
    {
        state.isFalse( hasVal, "value already built" );

        switch ( ev.type() ) {
        case VALUE: processValue( ev ); break;
        case MAP_START: processMapStart( ev ); break;
        case STRUCT_START: processStructStart( ev ); break;
        case FIELD_START: processFieldStart( ev ); break;
        case LIST_START: processListStart( ev ); break;
        case END: processEnd( ev ); break;
        default: throw new UnsupportedOperationException( "Unimplemented" );
        }
    }

    public
    static
    BuildReactor
    forFactory( Factory fact )
    {
        inputs.notNull( fact, "fact" );
        return new BuildReactor( fact );
    }
}
