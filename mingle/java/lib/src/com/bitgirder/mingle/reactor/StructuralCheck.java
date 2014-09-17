package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.mingle.Mingle;
import com.bitgirder.mingle.MingleTypeReference;
import com.bitgirder.mingle.ListTypeReference;
import com.bitgirder.mingle.AtomicTypeReference;
import com.bitgirder.mingle.PointerTypeReference;
import com.bitgirder.mingle.MingleIdentifier;
import com.bitgirder.mingle.QualifiedTypeName;

import com.bitgirder.lang.Lang;

import java.util.Deque;
import java.util.Set;

public
final
class StructuralCheck
implements MingleReactor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleReactorTopType topType;

    private final Deque< Object > stack = Lang.newDeque();

    private boolean done;

    private
    StructuralCheck( MingleReactorTopType topType )
    {
        this.topType = topType;
    }

    private
    final
    static
    class ListCheck
    {
        private final MingleTypeReference eltTyp;

        private 
        ListCheck( ListTypeReference lt ) 
        { 
            this.eltTyp = lt.getElementType();
        }
    }

    private
    static
    String
    descForEvent( MingleReactorEvent ev )
    {
        switch ( ev.type() ) {
        case LIST_START: return "start of " + ev.listType().getExternalForm();
        case MAP_START: return Mingle.TYPE_SYMBOL_MAP.getExternalForm();
        case END: return "end";
        case VALUE: return Mingle.typeOf( ev.value() ).getExternalForm();
        case FIELD_START: 
            return "start of field '" + ev.field().getExternalForm() + "'";
        case STRUCT_START:
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

        if ( obj instanceof ListCheck ) return "a list value";

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

        if ( obj instanceof MingleReactorEvent ) {
            return descForEvent( (MingleReactorEvent) obj );
        }
        
        return obj.toString();
    }

    private
    static
    void
    failStructure( CharSequence msg )
        throws Exception
    {
        throw new MingleReactorException( msg.toString() );
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
    class MapCheck
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
    checkNotDone( MingleReactorEvent ev )
        throws Exception
    {
        if ( ! done ) return;

        failStructuref( "Saw %s after value was built", 
            sawDescFor( ev ) );
    }

    private
    void
    failTopType( MingleReactorEvent ev )
        throws Exception
    {
        failStructuref( "Expected %s but got %s", 
            topType.toString().toLowerCase(), descForEvent( ev ) );
    }

    // only to be called when ev arrives with an stack empty
    private
    void
    checkTopType( MingleReactorEvent ev )
        throws Exception
    {
        if ( topType.couldStartWithEvent( ev ) ) return;

        failTopType( ev );
    }

    private
    void
    failListValueTypeError( MingleTypeReference expct, 
                            MingleReactorEvent ev )
        throws Exception
    {
        failStructuref( "expected list value of type %s but saw %s",
            expct, sawDescFor( ev ) );
    }

    private
    void
    checkValueTypeForList( MingleTypeReference calledType,
                           MingleTypeReference effectiveType,
                           MingleTypeReference valType,
                           MingleReactorEvent ev )
        throws Exception
    {
        if ( effectiveType instanceof PointerTypeReference ) {
            PointerTypeReference pt = (PointerTypeReference) effectiveType;
            checkValueTypeForList( calledType, pt.getType(), valType, ev );
            return;
        } else if ( effectiveType instanceof AtomicTypeReference ) {
            AtomicTypeReference at = (AtomicTypeReference) effectiveType;
            if ( at.getRestriction() != null ) {
                AtomicTypeReference t2 = at.asUnrestrictedType();
                checkValueTypeForList( calledType, t2, valType, ev );
                return;
            }
        } 
 
        if ( Mingle.canAssignType( valType, effectiveType ) ) return;
        failListValueTypeError( calledType, ev );
    }

    private
    void
    checkEventForList( MingleReactorEvent ev,
                       ListCheck lc )
        throws Exception
    {
        MingleTypeReference evTyp = null;

        switch ( ev.type() ) {
        case VALUE: evTyp = Mingle.typeOf( ev.value() ); break;
        case LIST_START: evTyp = ev.listType(); break;
        case MAP_START: evTyp = Mingle.TYPE_SYMBOL_MAP; break;
        case STRUCT_START: 
            evTyp = AtomicTypeReference.create( ev.structType() );
            break;
        default: return;
        }

        checkValueTypeForList( lc.eltTyp, lc.eltTyp, evTyp, ev );
    }

    private
    void
    execValueCheck( MingleReactorEvent ev,
                    Object pushIfOk )
        throws Exception
    {
        Object top = stack.peek();

        if ( top instanceof MapCheck ) failUnexpectedMapEnd( ev );

        if ( top == null || top instanceof ListCheck || 
             top instanceof MingleIdentifier ) 
        {
            if ( top == null ) checkTopType( ev );

            if ( top instanceof ListCheck ) {
                checkEventForList( ev, (ListCheck) top );
            }

            if ( pushIfOk != null ) stack.push( pushIfOk );
            return; 
        }

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
    checkValue( MingleReactorEvent ev )
        throws Exception
    {
        execValueCheck( ev, null );
        completeValue();
    }

    private
    void
    checkStartList( MingleReactorEvent ev )
        throws Exception
    {
        execValueCheck( ev, new ListCheck( ev.listType() ) );
    }

    private
    void
    checkStartStructure( MingleReactorEvent ev )
        throws Exception
    {
        execValueCheck( ev, new MapCheck() );
    }

    private
    void
    checkStartField( MingleReactorEvent ev )
        throws Exception
    {
        if ( stack.isEmpty() ) failTopType( ev );

        Object top = stack.peek();

        if ( top instanceof MapCheck ) {
            MapCheck mc = (MapCheck) top;
            mc.startField( ev.field() );
            stack.push( ev.field() );
            return;
        }

        failStructuref( "Saw start of field '%s' while expecting %s",
            ev.field().getExternalForm(), expectDescFor( top ) );
    }

    private
    void
    checkEnd( MingleReactorEvent ev )
        throws Exception
    {
        if ( stack.isEmpty() ) failTopType( ev );

        Object top = stack.peek();

        if ( top instanceof ListCheck || top instanceof MapCheck ) {
            stack.pop();
            completeValue();
            return;
        }
        
        failStructuref( "Saw end while expecting %s", expectDescFor( top ) );
    }

    public
    void
    processEvent( MingleReactorEvent ev )
        throws Exception
    {
        checkNotDone( ev );

        switch ( ev.type() ) {
        case VALUE: checkValue( ev ); break;
        case MAP_START: checkStartStructure( ev ); break;
        case LIST_START: checkStartList( ev ); break;
        case STRUCT_START: checkStartStructure( ev ); break;
        case FIELD_START: checkStartField( ev ); break;
        case END: checkEnd( ev ); break;
        default: state.failf( "unhandled event: %s", ev.type() );
        }
    }

    public
    static
    StructuralCheck
    forTopType( MingleReactorTopType topType )
    {
        inputs.notNull( topType, "topType" );
        return new StructuralCheck( topType );
    }

    public
    static
    StructuralCheck
    create()
    {
        return forTopType( MingleReactorTopType.VALUE );
    }
}
