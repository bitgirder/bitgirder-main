package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.mingle.Mingle;
import com.bitgirder.mingle.MingleValue;
import com.bitgirder.mingle.MingleIdentifier;
import com.bitgirder.mingle.MingleUnrecognizedFieldException;
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

    public final static Object UNHANDLED_VALUE_MARKER = new Object();

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

    private final static ExceptionFactory DEFAULT_EXCPT_FACT;

    private final Deque< Object > stack = Lang.newDeque();

    private boolean hasVal; // indicates whether val may be returned
    private Object val; // the built value

    private final ExceptionFactory excptFact;

    private 
    BuildReactor( Builder b )
    {
        stack.push( inputs.notNull( b.fact, "fact" ) );
        this.excptFact = b.excptFact;
    }

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
    Exception    
    failBuilderBadInput( MingleReactorEvent ev )
    {
        String msg = String.format( "unhandled value: %s", 
            MingleReactors.typeOfEvent( ev ).getExternalForm() ); 

        return excptFact.createException( ev.path(), msg );
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
            throw failBuilderBadInput( ev );
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

        if ( val == UNHANDLED_VALUE_MARKER ) throw failBuilderBadInput( ev );
        completeValue( val, ev );
    }

    private
    void
    processFieldStart( MingleReactorEvent ev )
        throws Exception
    {
        FieldSetBuilder fsb = state.cast( FieldSetBuilder.class, stack.peek() );

        MingleIdentifier fld = ev.field();

        Factory f = fsb.startField( fld, ev.path() );

        if ( f == null ) {
            ObjectPath< MingleIdentifier > p = ev.path().getParent();
            throw new MingleUnrecognizedFieldException( fld, p );
        }

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

        FieldSetBuilder fsb = f.startStruct( ev.structType(), ev.path() );
        if ( fsb == null ) throw failBuilderBadInput( ev );

        startFieldSet( fsb );
    }

    private
    void
    processListStart( MingleReactorEvent ev )
        throws Exception
    {
        Factory f = nextFactory( ev );

        ListBuilder lb = f.startList( ev.listType(), ev.path() );
        if ( lb == null ) throw failBuilderBadInput( ev );

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
        default: state.failf( "unhandled event: %s", ev.type() );
        }
    }

    public
    final
    static
    class Builder
    {
        private Factory fact;
        private ExceptionFactory excptFact = DEFAULT_EXCPT_FACT;

        public
        Builder
        setFactory( Factory fact )
        {
            this.fact = inputs.notNull( fact, "fact" );
            return this;
        }

        public
        Builder
        setExceptionFactory( ExceptionFactory excptFact )
        {
            this.excptFact = inputs.notNull( excptFact, "excptFact" );
            return this;
        }

        public BuildReactor build() { return new BuildReactor( this ); }
    }

    static
    {
        DEFAULT_EXCPT_FACT = new ExceptionFactory()
        {
            public
            Exception
            createException( ObjectPath< MingleIdentifier > path,
                             String msg )
            {
                return new Exception( Mingle.formatError( path, msg ) );
            }
        };
    }
}
