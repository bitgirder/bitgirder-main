package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;

import java.util.Deque;
import java.util.Set;

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

    // LIST_ACC_OBJ is used in the stack to indicate that we are accumulating a
    // list
    private final static Object LIST_ACC_OBJ =
        MingleValueReactorEvent.Type.START_LIST;

    private final MingleValueReactorTopType topType;

    private final Deque< Object > stack = Lang.newDeque();

    private boolean done;

    private
    MingleValueStructuralCheck( MingleValueReactorTopType topType )
    {
        this.topType = topType;
    }

    private
    static
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
    static
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
    static
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
    static
    void
    failStructure( CharSequence msg )
        throws Exception
    {
        throw new MingleValueReactorException( msg.toString() );
    }

    private
    static
    void
    failStructuref( String tmpl,
                    Object... args )
        throws Exception
    {
        failStructure( String.format( tmpl, args ) );
    }

    private
    final
    static
    class MapContext
    {
        private final Set< MingleIdentifier > seen = Lang.newSet();

        private
        void
        startField( MingleIdentifier fld )
            throws Exception
        {
            if ( seen.add( fld ) ) return;

            failStructuref( "Multiple entries for field: %s", 
                fld.getExternalForm() );
        }
    }

    private
    void
    failUnexpectedMapEnd( Object sawDesc )
        throws Exception
    {
        failStructuref( "Expected field name or end of fields but got %s",
            sawDescFor( sawDesc ) );
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
        if ( topType.couldStartWithEvent( ev ) ) return;

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

        if ( top instanceof MapContext ) failUnexpectedMapEnd( ev );

        failStructuref( "Saw %s while expecting %s", 
            sawDescFor( ev ), expectDescFor( top ) );
    }

    private
    void
    completeValue()
    {
        // if this value completed a field value pop the field
        if ( stack.peek() instanceof MingleIdentifier ) stack.pop();

        if ( stack.isEmpty() ) done = true;
    }

    private
    void
    checkValue( MingleValueReactorEvent ev )
        throws Exception
    {
        execValueCheck( ev, null );
        completeValue();
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
    checkStartStructure( MingleValueReactorEvent ev )
        throws Exception
    {
        execValueCheck( ev, new MapContext() );
    }

    private
    void
    checkStartField( MingleValueReactorEvent ev )
        throws Exception
    {
        if ( stack.isEmpty() ) failTopType( ev );

        Object top = stack.peek();

        if ( top instanceof MapContext ) {
            MapContext mc = (MapContext) top;
            mc.startField( ev.field() );
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

        if ( top == LIST_ACC_OBJ || top instanceof MapContext ) {
            stack.pop();
            completeValue();
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
        case START_MAP: checkStartStructure( ev ); break;
        case START_LIST: checkStartList( ev ); break;
        case START_STRUCT: checkStartStructure( ev ); break;
        case START_FIELD: checkStartField( ev ); break;
        case END: checkEnd( ev ); break;
        default: state.failf( "unhandled event: %s", ev.type() );
        }
    }

    public
    static
    MingleValueStructuralCheck
    createWithTopType( MingleValueReactorTopType topType )
    {
        inputs.notNull( topType, "topType" );
        return new MingleValueStructuralCheck( topType );
    }

    public
    static
    MingleValueStructuralCheck
    create()
    {
        return createWithTopType( MingleValueReactorTopType.VALUE );
    }
}
