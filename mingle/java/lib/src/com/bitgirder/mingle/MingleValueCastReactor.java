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
    failCast( MingleValueReactorEvent ev,
              String msg )
    {
        throw new MingleValueCastException( msg, ev.path() );
    }

    private
    void
    failCastf( MingleValueReactorEvent ev,
               String tmpl,
               Object... args )
    {
        failCast( ev, String.format( tmpl, args ) );
    }

    private
    void
    failCastType( MingleValueReactorEvent ev,
                  MingleTypeReference expct,
                  MingleTypeReference act )
    {
        throw Mingle.failCastType( expct, act, ev.path() );
    }

    private
    void
    failUnhandledStackValue( Object obj )
    {
        state.failf( "unhandled stack value: %s", obj );
    }

    private
    final
    static
    class ListCast
    {
        private final ListTypeReference type;
        private final MingleTypeReference callType;

        private boolean sawValues;

        private 
        ListCast( ListTypeReference type,
                  MingleTypeReference callType ) 
        { 
            this.type = type; 
            this.callType = callType;
        }
    }

    private void push( Object obj ) { stack.push( obj ); }

    private
    void
    processAtomicValue( MingleValueReactorEvent ev,
                        AtomicTypeReference typ,
                        MingleValueReactor next )
        throws Exception
    {
        MingleValue mv = Mingle.castAtomic( ev.value(), typ, ev.path() );
        ev.setValue( mv );
        next.processEvent( ev );
    }

    private
    void
    processNullableValue( MingleValueReactorEvent ev,
                          NullableTypeReference typ,
                          MingleValueReactor next )
        throws Exception
    {
        if ( ev.value() instanceof MingleNull ) {
            next.processEvent( ev );
            return;
        }

        processValueWithType( ev, typ.getTypeReference(), next );
    }

    private
    void
    processValueWithType( MingleValueReactorEvent ev,
                          MingleTypeReference typ,
                          MingleValueReactor next )
        throws Exception
    {
        if ( typ instanceof AtomicTypeReference ) {
            processAtomicValue( ev, (AtomicTypeReference) typ, next );
        } else if ( typ instanceof NullableTypeReference ) {
            processNullableValue( ev, (NullableTypeReference) typ, next );
        } else if ( typ instanceof ListTypeReference ) {
            failCastType( ev, typ, Mingle.inferredTypeOf( ev.value() ) );
        } else {
            state.failf( "unhandled type: %s", typ );
        }
    }

    private
    void
    processValue( MingleValueReactorEvent ev,
                  Object obj,
                  MingleValueReactor next )
        throws Exception
    {
        if ( obj instanceof MingleTypeReference ) {
            processValueWithType( ev, (MingleTypeReference) obj, next );
        } else if ( obj instanceof ListCast ) {
            ListCast lc = (ListCast) obj;
            processValueWithType( ev, lc.type.getElementTypeReference(), next );
            lc.sawValues = true;
        } else {
            failUnhandledStackValue( obj );
        }
    } 

    private
    void
    processStartListWithType( MingleValueReactorEvent ev,
                              MingleTypeReference typ,
                              MingleTypeReference callType,
                              MingleValueReactor next )
        throws Exception
    {
        if ( typ instanceof ListTypeReference ) {
            stack.push( new ListCast( (ListTypeReference) typ, callType ) );
            next.processEvent( ev );
        } else if ( typ instanceof NullableTypeReference ) {
            NullableTypeReference nt = (NullableTypeReference) typ;
            MingleTypeReference eltTyp = nt.getTypeReference();
            processStartListWithType( ev, eltTyp, callType, next );
        } else {
            failCastType( ev, callType, Mingle.TYPE_VALUE_LIST );
        }
    }

    private
    void
    processStartList( MingleValueReactorEvent ev,
                      MingleValueReactor next )
        throws Exception
    {
        Object obj = stack.peek();

        if ( obj instanceof MingleTypeReference ) {
            MingleTypeReference typ = (MingleTypeReference) obj;
            processStartListWithType( ev, typ, typ, next );
        } else {
            failUnhandledStackValue( obj );
        }
    }

    private
    void
    processStartStructWithType( MingleValueReactorEvent ev,
                                MingleTypeReference typ,
                                MingleValueReactor next )
        throws Exception
    {
        if ( typ instanceof AtomicTypeReference ) 
        {
            AtomicTypeReference at = (AtomicTypeReference) typ;

            if ( ! at.getName().equals( ev.structType() ) ) 
            {
                AtomicTypeReference failTyp = 
                    new AtomicTypeReference( ev.structType(), null );

                failCastType( ev, typ, failTyp );
            }

            next.processEvent( ev );
        }
        else throw new UnsupportedOperationException( "Unimplemented" );
    }

    private
    void
    processStartStruct( MingleValueReactorEvent ev,
                        MingleValueReactor next )
        throws Exception
    {
        Object obj = stack.peek();

        if ( obj instanceof MingleTypeReference ) {
            processStartStructWithType( ev, (MingleTypeReference) obj, next );
        } else if ( obj instanceof ListCast ) {
            ListCast lc = (ListCast) obj;
            lc.sawValues = true;
            MingleTypeReference eltTyp = lc.type.getElementTypeReference();
            processStartStructWithType( ev, eltTyp, next );
        } else {
            failUnhandledStackValue( obj );
        }
    }

    private
    void
    processEnd( MingleValueReactorEvent ev,
                MingleValueReactor next )
        throws Exception
    {
        if ( stack.peek() instanceof ListCast ) {
            ListCast lc = (ListCast) stack.pop();
            if ( ! ( lc.sawValues || lc.type.allowsEmpty() ) ) {
                failCastf( ev, "List is empty" );
            }
        }

        next.processEvent( ev );
    }

    public
    void
    processPipelineEvent( MingleValueReactorEvent ev,
                          MingleValueReactor next )
        throws Exception
    {
        switch ( ev.type() ) {
        case VALUE: processValue( ev, stack.peek(), next ); return;
        case START_LIST: processStartList( ev, next ); return;
        case START_MAP: next.processEvent( ev ); return;
        case START_STRUCT: processStartStruct( ev, next ); return;
        case START_FIELD: next.processEvent( ev ); return;
        case END: processEnd( ev, next ); return;
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
