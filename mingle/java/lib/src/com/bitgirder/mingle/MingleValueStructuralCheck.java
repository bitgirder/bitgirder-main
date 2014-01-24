package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import java.util.Deque;

public
final
class MingleValueStructuralCheck
implements MingleValueReactor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private static void code( Object... args ) { CodeLoggers.code( args ); }

    private 
    static 
    void 
    codef( String tmpl, 
           Object... args ) 
    { 
        CodeLoggers.codef( tmpl, args ); 
    }

    // LIST_ACC_OBJ and MAP_ACC_OBJ are opaque objects used in the stack to
    // indicate that we are accumulating a map or list

    private final static Object LIST_ACC_OBJ =
        MingleValueReactorEvent.Type.START_LIST;
    
    private final static Object MAP_ACC_OBJ =
        MingleValueReactorEvent.Type.START_MAP;

    private final static Object END_OBJ = MingleValueReactorEvent.Type.END;

    public
    static
    enum TopType
    {
        VALUE( MingleValueReactorEvent.Type.VALUE ),
        LIST( MingleValueReactorEvent.Type.START_LIST ),
        MAP( MingleValueReactorEvent.Type.START_MAP ),
        STRUCT( MingleValueReactorEvent.Type.START_STRUCT );
        
        private final MingleValueReactorEvent.Type expectType;

        private
        TopType( MingleValueReactorEvent.Type expectType )
        {
            this.expectType = expectType;
        }
    }

    private final TopType topType;

    private final Deque< Object > stack = Lang.newDeque();

    private boolean done;

    private
    MingleValueStructuralCheck( TopType topType )
    {
        this.topType = topType;
    }

    private
    String
    descForEvent( MingleValueReactorEvent ev )
    {
        switch ( ev.type() ) {
        case START_LIST: return "list start";
        case START_MAP: return "map start";
        case END: return "end";
        case VALUE: return "value";
        case START_FIELD: 
            return "start of field '" + ev.field().getExternalForm() + "'";
        case START_STRUCT:
            return "start of struct " + ev.structType().getExternalForm();
        default: throw state.failf( "unhandled type: %s", ev.type() );
        }
    }

    private
    String
    expectDescFor( Object obj )
    {
        if ( obj == null ) return "BEGIN";

        if ( obj instanceof MingleIdentifier ) {
            return "a value for field '" + 
                ( (MingleIdentifier) obj ).getExternalForm() + "'";
        }

        if ( obj == LIST_ACC_OBJ ) return "a list value";

        return obj.toString();
    }

    private
    String
    sawDescFor( Object obj )
    {
        if ( obj == null ) return "BEGIN";

        if ( obj instanceof MingleIdentifier ) {
            return "start of field '" + 
                ( (MingleIdentifier) obj ).getExternalForm() + "'";
        }

        if ( obj instanceof QualifiedTypeName ) {
            return "start of struct " +
                ( (QualifiedTypeName) obj ).getExternalForm();
        }

        if ( obj instanceof MingleValueReactorEvent ) {
            return descForEvent( (MingleValueReactorEvent) obj );
        }
        
        return obj.toString();
    }

    private
    void
    failStructure( CharSequence msg )
        throws Exception
    {
        throw new MingleValueReactorException( msg.toString() );
    }

    private
    void
    failStructuref( String tmpl,
                    Object... args )
        throws Exception
    {
        failStructure( String.format( tmpl, args ) );
    }

    private
    void
    failUnexpectedMapEnd( Object sawDesc )
        throws Exception
    {
        failStructuref( "Expected field name or end of fields but got %s",
            sawDescFor( sawDesc ) );
    }

    // indicates whether we're reading a map or struct body
    private
    boolean
    isAccumulatingMap( Object obj )
    {
        return obj == MAP_ACC_OBJ || ( obj instanceof QualifiedTypeName );
    }

    private
    void
    checkNotDone( MingleValueReactorEvent ev )
        throws Exception
    {
        if ( ! done ) return;

        failStructuref( "Saw %s after value was built", 
            sawDescFor( ev ) );
    }

    private
    void
    failTopType( MingleValueReactorEvent ev )
        throws Exception
    {
        failStructuref( "Expected %s but got %s", 
            topType.toString().toLowerCase(), descForEvent( ev ) );
    }

    // only to be called when ev arrives with an stack empty
    private
    void
    checkTopType( MingleValueReactorEvent ev )
        throws Exception
    {
        if ( topType == TopType.VALUE || topType.expectType == ev.type() ) {
            return;
        }

        failTopType( ev );
    }

    private
    void
    execValueCheck( MingleValueReactorEvent ev,
                    Object pushIfOk )
        throws Exception
    {
        Object top = stack.peek();

        if ( top == null || top == LIST_ACC_OBJ || 
                top instanceof MingleIdentifier ) 
        {
            if ( top == null ) checkTopType( ev );
            if ( pushIfOk != null ) stack.push( pushIfOk );
            return; 
        }

        if ( isAccumulatingMap( top ) ) failUnexpectedMapEnd( ev );

        failStructuref( "Saw %s while expecting %s", 
            sawDescFor( ev ), expectDescFor( top ) );
    }

    private
    void
    checkValue( MingleValueReactorEvent ev )
        throws Exception
    {
        execValueCheck( ev, null );

        // if this value completed a field value pop the field
        if ( stack.peek() instanceof MingleIdentifier ) stack.pop();
    }

    private
    void
    checkStartMap( MingleValueReactorEvent ev )
        throws Exception
    {
        execValueCheck( ev, MAP_ACC_OBJ );
    }

    private
    void
    checkStartList( MingleValueReactorEvent ev )
        throws Exception
    {
        execValueCheck( ev, LIST_ACC_OBJ );
    }

    private
    void
    checkStartStruct( MingleValueReactorEvent ev )
        throws Exception
    {
        execValueCheck( ev, ev.structType() );
    }

    private
    void
    checkStartField( MingleValueReactorEvent ev )
        throws Exception
    {
        if ( stack.isEmpty() ) failTopType( ev );

        Object top = stack.peek();

        if ( top == MAP_ACC_OBJ || ( top instanceof QualifiedTypeName ) ) {
            stack.push( ev.field() );
            return;
        }

        failStructuref( "Saw start of field '%s' while expecting %s",
            ev.field().getExternalForm(), expectDescFor( top ) );
    }

    private
    void
    checkEnd( MingleValueReactorEvent ev )
        throws Exception
    {
        if ( stack.isEmpty() ) failTopType( ev );

        Object top = stack.peek();

        if ( top == LIST_ACC_OBJ || top == MAP_ACC_OBJ ||
                ( top instanceof QualifiedTypeName ) ) 
        {
            stack.pop();
            if ( stack.isEmpty() ) done = true;
            return;
        }
        
        failStructuref( "Saw end while expecting %s", expectDescFor( top ) );
    }

    public
    void
    processEvent( MingleValueReactorEvent ev )
        throws Exception
    {
        checkNotDone( ev );

        switch ( ev.type() ) {
        case VALUE: checkValue( ev ); break;
        case START_MAP: checkStartMap( ev ); break;
        case START_LIST: checkStartList( ev ); break;
        case START_STRUCT: checkStartStruct( ev ); break;
        case START_FIELD: checkStartField( ev ); break;
        case END: checkEnd( ev ); break;
        default: state.failf( "unhandled event: %s", ev.type() );
        }
    }

    public
    static
    MingleValueStructuralCheck
    createWithTopType( TopType topType )
    {
        inputs.notNull( topType, "topType" );
        return new MingleValueStructuralCheck( topType );
    }

    public
    static
    MingleValueStructuralCheck
    create()
    {
        return createWithTopType( TopType.VALUE );
    }
}
