package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

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

    private final static FieldTyper DEFAULT_FIELD_TYPER = new FieldTyper() {
        public MingleTypeReference fieldTypeFor( MingleIdentifier fld ) {
            return Mingle.TYPE_VALUE;
        }
    };

    private final Deque< Object > stack = Lang.newDeque();

    private MingleValueCastReactor() {}

    public
    void
    initialize( PipelineInitializationContext< Object > ctx )
    {
        MingleValueReactors.ensureStructuralCheck( ctx );
        MingleValueReactors.ensurePathSetter( ctx );
    }

    public
    static
    interface FieldTyper
    {
        public 
        MingleTypeReference
        fieldTypeFor( MingleIdentifier fld );
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
        private final MingleTypeReference callTyp;

        private boolean sawValues;

        private 
        ListCast( ListTypeReference type,
                  MingleTypeReference callTyp ) 
        { 
            this.type = type; 
            this.callTyp = callTyp;
        }
    }

    private void push( Object obj ) { stack.push( obj ); }

    private
    void
    processAtomicValue( MingleValueReactorEvent ev,
                        AtomicTypeReference typ,
                        MingleTypeReference callTyp,
                        MingleValueReactor next )
        throws Exception
    {
        MingleValue mv = 
            Mingle.castAtomic( ev.value(), typ, callTyp, ev.path() );

        ev.setValue( mv );

        next.processEvent( ev );
    }

    private
    void
    processNullableValue( MingleValueReactorEvent ev,
                          NullableTypeReference typ,
                          MingleTypeReference callTyp,
                          MingleValueReactor next )
        throws Exception
    {
        if ( ev.value() instanceof MingleNull ) {
            next.processEvent( ev );
            return;
        }

        processValueWithType( ev, typ.getValueType(), callTyp, next );
    }

    private
    void
    processValueWithType( MingleValueReactorEvent ev,
                          MingleTypeReference typ,
                          MingleTypeReference callTyp,
                          MingleValueReactor next )
        throws Exception
    {
        if ( typ instanceof AtomicTypeReference ) {
            processAtomicValue( ev, (AtomicTypeReference) typ, callTyp, next );
        } else if ( typ instanceof NullableTypeReference ) {
            NullableTypeReference nt = (NullableTypeReference) typ;
            processNullableValue( ev, nt, callTyp, next );
        } else if ( typ instanceof ListTypeReference ) {
            failCastType( ev, callTyp, Mingle.inferredTypeOf( ev.value() ) );
        } else {
            state.failf( "unhandled type: %s", typ );
        }
    }

    private
    void
    processValue( MingleValueReactorEvent ev,
                  MingleValueReactor next )
        throws Exception
    {
        Object obj = stack.peek();

        if ( obj instanceof MingleTypeReference ) {
            MingleTypeReference typ = (MingleTypeReference) stack.pop();
            processValueWithType( ev, typ, typ, next );
        } else if ( obj instanceof ListCast ) {
            ListCast lc = (ListCast) obj;
            MingleTypeReference typ = lc.type.getElementType();
            processValueWithType( ev, typ, typ, next );
            lc.sawValues = true;
        } else {
            failUnhandledStackValue( obj );
        }
    } 

    private
    void
    processStartListWithAtomicType( MingleValueReactorEvent ev,
                                    AtomicTypeReference at,
                                    MingleTypeReference callTyp,
                                    MingleValueReactor next )
        throws Exception
    {
        if ( at.equals( Mingle.TYPE_VALUE ) ) 
        {
            processStartListWithType( 
                ev, Mingle.TYPE_VALUE_LIST, callTyp, next );

            return;
        }

        failCastType( ev, callTyp, Mingle.TYPE_VALUE_LIST );
    }

    private
    void
    processStartListWithType( MingleValueReactorEvent ev,
                              MingleTypeReference typ,
                              MingleTypeReference callTyp,
                              MingleValueReactor next )
        throws Exception
    {
        if ( typ instanceof AtomicTypeReference ) {
            AtomicTypeReference at = (AtomicTypeReference) typ;
            processStartListWithAtomicType( ev, at, callTyp, next );
        } else if ( typ instanceof ListTypeReference ) {
            stack.push( new ListCast( (ListTypeReference) typ, callTyp ) );
            next.processEvent( ev );
        } else if ( typ instanceof NullableTypeReference ) {
            NullableTypeReference nt = (NullableTypeReference) typ;
            MingleTypeReference valTyp = nt.getValueType();
            processStartListWithType( ev, valTyp, callTyp, next );
        } else {
            failCastType( ev, callTyp, Mingle.TYPE_VALUE_LIST );
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
            MingleTypeReference typ = (MingleTypeReference) stack.pop();
            processStartListWithType( ev, typ, typ, next );
        } else if ( obj instanceof ListCast ) {
            ListCast lc = (ListCast) obj;
            lc.sawValues = true;
            MingleTypeReference eltTyp = lc.type.getElementType();
            processStartListWithType( ev, eltTyp, lc.type, next );
        } else {
            failUnhandledStackValue( obj );
        }
    }

    private
    void
    implStartMap( MingleValueReactorEvent ev,
                  FieldTyper ft,
                  MingleValueReactor next )
        throws Exception
    {
        stack.push( ft );
        next.processEvent( ev );
    }

    private
    void
    processStartMapWithAtomicType( MingleValueReactorEvent ev,
                                   AtomicTypeReference at,
                                   MingleTypeReference callTyp,
                                   MingleValueReactor next )
        throws Exception
    {
        if ( at.equals( Mingle.TYPE_SYMBOL_MAP ) || 
             at.equals( Mingle.TYPE_VALUE ) )
        {
            implStartMap( ev, DEFAULT_FIELD_TYPER, next );
            return;
        }

        failCastType( ev, callTyp, at );
    }

    private
    void
    processStartMapWithType( MingleValueReactorEvent ev,
                             MingleTypeReference typ,
                             MingleTypeReference callTyp,
                             MingleValueReactor next )
        throws Exception
    {
        if ( typ instanceof AtomicTypeReference ) {
            AtomicTypeReference at = (AtomicTypeReference) typ;
            processStartMapWithAtomicType( ev, at, callTyp, next );
        } else if ( typ instanceof NullableTypeReference ) {
            NullableTypeReference nt = (NullableTypeReference) typ;
            MingleTypeReference valTyp = nt.getValueType();
            processStartMapWithType( ev, valTyp, callTyp, next );
        } else {
            failCastType( ev, callTyp, typ );
        }
    }

    private
    void
    processStartMap( MingleValueReactorEvent ev,
                     MingleValueReactor next )
        throws Exception
    {
        Object obj = stack.peek();

        if ( obj instanceof MingleTypeReference ) {
            MingleTypeReference typ = (MingleTypeReference) stack.pop();
            processStartMapWithType( ev, typ, typ, next );
        } else if ( obj instanceof ListCast ) {
            ListCast lc = (ListCast) obj;
            lc.sawValues = true;
            MingleTypeReference typ = lc.type.getElementType();
            processStartMapWithType( ev, typ, typ, next );
        } else {
            failUnhandledStackValue( obj );
        }
    }

    private
    void
    processStartField( MingleValueReactorEvent ev,
                       MingleValueReactor next )
        throws Exception
    {
        FieldTyper ft = state.cast( FieldTyper.class, stack.peek() );

        MingleTypeReference typ = ft.fieldTypeFor( ev.field() );
        stack.push( typ );

        next.processEvent( ev );
    }

    private
    void
    processStartStructWithAtomicType( MingleValueReactorEvent ev,
                                      AtomicTypeReference at,
                                      MingleTypeReference callTyp,
                                      MingleValueReactor next )
        throws Exception
    {
        boolean ok = at.getName().equals( ev.structType() ) || 
            at.equals( Mingle.TYPE_VALUE );

        if ( ( ! ok ) && at.equals( Mingle.TYPE_SYMBOL_MAP ) ) {
            ok = true;
            ev.setStartMap();
        }

        if ( ok ) {
            implStartMap( ev, DEFAULT_FIELD_TYPER, next );
            return;
        }
 
        AtomicTypeReference failTyp = 
            new AtomicTypeReference( ev.structType(), null );

        failCastType( ev, callTyp, failTyp );
    }

    private
    void
    processStartStructWithType( MingleValueReactorEvent ev,
                                MingleTypeReference typ,
                                MingleTypeReference callTyp,
                                MingleValueReactor next )
        throws Exception
    {
        if ( typ instanceof AtomicTypeReference ) {
            AtomicTypeReference at = (AtomicTypeReference) typ;
            processStartStructWithAtomicType( ev, at, callTyp, next );
        } else if ( typ instanceof NullableTypeReference ) {
            NullableTypeReference nt = (NullableTypeReference) typ;
            MingleTypeReference valTyp = nt.getValueType();
            processStartStructWithType( ev, valTyp, callTyp, next );
        } else {
            failCastType( ev, callTyp, typ );
        }
    }

    private
    void
    processStartStruct( MingleValueReactorEvent ev,
                        MingleValueReactor next )
        throws Exception
    {
        Object obj = stack.peek();

        if ( obj instanceof MingleTypeReference ) {
            MingleTypeReference typ = (MingleTypeReference) stack.pop();
            processStartStructWithType( ev, typ, typ, next );
        } else if ( obj instanceof ListCast ) {
            ListCast lc = (ListCast) obj;
            lc.sawValues = true;
            MingleTypeReference eltTyp = lc.type.getElementType();
            processStartStructWithType( ev, eltTyp, eltTyp, next );
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
        if ( stack.peek() instanceof ListCast ) 
        {
            ListCast lc = (ListCast) stack.pop();
            if ( ! ( lc.sawValues || lc.type.allowsEmpty() ) ) {
                failCastf( ev, "List is empty" );
            }
        } 
        else if ( stack.peek() instanceof FieldTyper ) stack.pop();

        next.processEvent( ev );
    }

    public
    void
    processPipelineEvent( MingleValueReactorEvent ev,
                          MingleValueReactor next )
        throws Exception
    {
        switch ( ev.type() ) {
        case VALUE: processValue( ev, next ); return;
        case START_LIST: processStartList( ev, next ); return;
        case START_MAP: processStartMap( ev, next ); return;
        case START_STRUCT: processStartStruct( ev, next ); return;
        case START_FIELD: processStartField( ev, next ); return;
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
