package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPaths;

import com.bitgirder.pipeline.PipelineInitializerContext;
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

    public
    static
    interface FieldTyper
    {
        public 
        MingleTypeReference
        fieldTypeFor( MingleIdentifier fld,
                      ObjectPath< MingleIdentifier > path )
            throws MingleValueCastException;
    }

    private final static FieldTyper DEFAULT_FIELD_TYPER = new FieldTyper() 
    {
        public 
        MingleTypeReference 
        fieldTypeFor( MingleIdentifier fld,
                      ObjectPath< MingleIdentifier > path ) 
        {
            return Mingle.TYPE_NULLABLE_VALUE;
        }
    };

    public 
    static 
    FieldTyper 
    getDefaultFieldTyper() 
    { 
        return DEFAULT_FIELD_TYPER;
    }

    public
    static
    interface Delegate
    {
        public 
        FieldTyper 
        fieldTyperFor( QualifiedTypeName qn,
                       ObjectPath< MingleIdentifier > path )
            throws MingleValueCastException;
        
        public
        boolean
        inferStructFor( QualifiedTypeName qn );

        // returns a non-null MingleValue (which may be an instanceof of
        // MingleNull) or throws a cast exception if this delegate can
        // definitively perform the atomic cast. Return java null to indicate
        // that the reactor should take its default action and execute the cast.
        public
        MingleValue
        castAtomic( MingleValue mv,
                    AtomicTypeReference at,
                    ObjectPath< MingleIdentifier > path )
            throws MingleValueCastException;
        
        public
        boolean
        allowAssign( QualifiedTypeName targ,
                     QualifiedTypeName act );
    }

    private final static Delegate DEFAULT_DELEGATE = new Delegate() 
    {
        public 
        FieldTyper 
        fieldTyperFor( QualifiedTypeName qn,
                       ObjectPath< MingleIdentifier > path ) 
        {
            return DEFAULT_FIELD_TYPER;
        }

        public boolean inferStructFor( QualifiedTypeName qn ) { return false; }

        public
        MingleValue
        castAtomic( MingleValue mv,
                    AtomicTypeReference at,
                    ObjectPath< MingleIdentifier > path )
        {
            return null;
        }

        public
        boolean
        allowAssign( QualifiedTypeName targ,
                     QualifiedTypeName act )
        {
            return false;
        }
    };

    private final Delegate del;

    private final Deque< Object > stack = Lang.newDeque();

    private MingleValueCastReactor( Builder b ) { this.del = b.del; }

    public
    void
    initialize( PipelineInitializerContext< Object > ctx )
    {
        MingleValueReactors.ensureStructuralCheck( ctx );
        MingleValueReactors.ensurePathSetter( ctx );
    }

    private
    void
    failCast( ObjectPath< MingleIdentifier > path,
              String msg )
    {
        throw new MingleValueCastException( msg, path );
    }

    private
    void
    failCast( MingleValueReactorEvent ev,
              String msg )
    {
        failCast( ev.path(), msg );
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

        private ObjectPath< MingleIdentifier > startPath;

        private 
        ListCast( ListTypeReference type,
                  MingleTypeReference callTyp,
                  ObjectPath< MingleIdentifier > startPath ) 
        { 
            this.type = type; 
            this.callTyp = callTyp;

            this.startPath = startPath == null ? 
                null : ObjectPaths.asImmutableCopy( startPath );
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
        MingleValue in = ev.value();
        ObjectPath< MingleIdentifier > path = ev.path();

        MingleValue mv = del.castAtomic( in, typ, path );
        if ( mv == null ) mv = Mingle.castAtomic( in, typ, callTyp, path );

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
            ListTypeReference lt = (ListTypeReference) typ;
            stack.push( new ListCast( lt, callTyp, ev.path() ) );
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
    boolean
    inferredStructForMap( MingleValueReactorEvent ev,
                          AtomicTypeReference at,
                          MingleValueReactor next )
        throws Exception
    {
        TypeName nm = at.getName();
        if ( ! ( nm instanceof QualifiedTypeName ) ) return false;

        QualifiedTypeName qn = (QualifiedTypeName) nm;
        if ( ! del.inferStructFor( qn ) ) return false;

        ev.setStartStruct( qn );
        completeStartStruct( ev, next );
        return true;
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

        if ( inferredStructForMap( ev, at, next ) ) return;
        
        failCastType( ev, callTyp, Mingle.TYPE_SYMBOL_MAP );
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

        // ev.path() will be something like 'foo.f1', but we pass just 'foo' to
        // the field typer
        ObjectPath< MingleIdentifier > loc = ev.path().getParent();
        MingleTypeReference typ = ft.fieldTypeFor( ev.field(), loc );

        state.isFalsef( typ == null, "field typer %s returned null for %s",
            ft, ev.field() );

        stack.push( typ );

        next.processEvent( ev );
    }

    private
    void
    completeStartStruct( MingleValueReactorEvent ev,
                         MingleValueReactor next )
        throws Exception
    {
        FieldTyper ft = del.fieldTyperFor( ev.structType(), ev.path() );
        if ( ft == null ) ft = DEFAULT_FIELD_TYPER;

        implStartMap( ev, ft, next );
    }

    private
    void
    processStartStructWithAtomicType( MingleValueReactorEvent ev,
                                      AtomicTypeReference at,
                                      MingleTypeReference callTyp,
                                      MingleValueReactor next )
        throws Exception
    {
        if ( at.equals( Mingle.TYPE_SYMBOL_MAP ) ) {
            ev.setStartMap();
            processStartMapWithAtomicType( ev, at, callTyp, next );
            return;
        } 

        QualifiedTypeName act = ev.structType();
        QualifiedTypeName targ = (QualifiedTypeName) at.getName();

        if ( at.getName().equals( act ) || at.equals( Mingle.TYPE_VALUE ) ||
             del.allowAssign( targ, act ) )
        {
            completeStartStruct( ev, next );
            return;
        }
 
        AtomicTypeReference failTyp = new AtomicTypeReference( act, null );
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
                failCast( lc.startPath, "List is empty" );
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
        case LIST_START: processStartList( ev, next ); return;
        case MAP_START: processStartMap( ev, next ); return;
        case STRUCT_START: processStartStruct( ev, next ); return;
        case FIELD_START: processStartField( ev, next ); return;
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
        return new Builder().setTargetType( typ ).build();
    }

    public
    final
    static
    class Builder
    {
        private MingleTypeReference typ;
        private Delegate del = DEFAULT_DELEGATE;

        public
        Builder
        setTargetType( MingleTypeReference typ )
        {
            this.typ = inputs.notNull( typ, "typ" );
            return this;
        }

        public
        Builder
        setDelegate( Delegate del )
        {
            this.del = inputs.notNull( del, "del" );
            return this;
        }

        public
        MingleValueCastReactor
        build()
        {
            MingleValueCastReactor res = new MingleValueCastReactor( this );
            res.push( typ );

            return res;
        }
    }
}
