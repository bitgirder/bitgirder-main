package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.pipeline.PipelineInitializer;
import com.bitgirder.pipeline.PipelineInitializerContext;

// impl follows that of the go libs
public
final
class MingleResponseReactor
implements MingleValueReactor,
           PipelineInitializer< Object >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final Delegate del;

    private MingleValueReactor evProc;
    private boolean hadProc;

    private final MingleValueReactors.DepthTracker depthTracker =
        new MingleValueReactors.DepthTracker() {
            public void depthBecameOne() {
                if ( evProc != null ) {
                    hadProc = true;
                    evProc = null;
                }
            }
        };

    private MingleResponseReactor( Delegate del ) { this.del = del; }

    public
    void
    initialize( PipelineInitializerContext< Object > ctx )
    {
        MingleValueReactors.ensureStructuralCheck( ctx );
        MingleValueReactors.ensurePathSetter( ctx );

        ctx.addElement( 
            new MingleValueCastReactor.Builder().
                setTargetType( Mingle.TYPE_RESPONSE ).
                setDelegate( CAST_DELEGATE ).
                build()
        );

        ctx.addElement(
            MingleValueReactors.createDebugReactor( "[post-cast]" ) );
    }

    private
    void
    sendEvProcEvent( MingleValueReactorEvent ev )
        throws Exception
    {
        boolean isNullVal = ev.type() == MingleValueReactorEvent.Type.VALUE &&
            ev.value() instanceof MingleNull;

        if ( hadProc && ( ! isNullVal ) )
        {
            throw new MingleValueCastException(
                "response has both a result and an error value", 
                ev.path().getParent() );
        }

        evProc.processEvent( ev );
    }

    private
    void
    startStruct( MingleValueReactorEvent ev )
        throws Exception
    {
        state.isTruef( ev.structType().equals( Mingle.QNAME_RESPONSE ),
            "unexpected response top struct type: %s", ev.structType() );
    }

    private
    void
    startField( MingleValueReactorEvent ev )
        throws Exception
    {
        MingleIdentifier fld = ev.field();
        ObjectPath< MingleIdentifier > p = ev.path();

        if ( fld.equals( Mingle.ID_RESULT ) ) {
            evProc = del.getResultReactor( p );
        } else if ( fld.equals( Mingle.ID_ERROR ) ) {
            evProc = del.getErrorReactor( p );
        } else {
            throw new MingleUnrecognizedFieldException( fld, p.getParent() );
        }
    }

    private
    void
    implProcessEvent( MingleValueReactorEvent ev )
        throws Exception
    {
        if ( evProc != null ) {
            sendEvProcEvent( ev );
            return;
        }

        switch ( ev.type() ) {
        case STRUCT_START: startStruct( ev ); break;
        case FIELD_START: startField( ev ); break;
        case END: break;
        default: state.failf( "saw %s while evProc is null", ev.inspect() );
        }
    }

    public
    void
    processEvent( MingleValueReactorEvent ev )
        throws Exception
    {
        implProcessEvent( ev );
        depthTracker.update( ev );
    }

    public
    static
    interface Delegate
    {
        public
        MingleValueReactor
        getResultReactor( ObjectPath< MingleIdentifier > path )
            throws Exception;
        
        public
        MingleValueReactor
        getErrorReactor( ObjectPath< MingleIdentifier > path )
            throws Exception;
    }

    public
    static
    MingleResponseReactor
    create( Delegate del )
    {
        return new MingleResponseReactor( inputs.notNull( del, "del" ) );
    }

    private final static MingleValueCastReactor.Delegate CAST_DELEGATE =
        new MingleValueCastReactor.Delegate()
        {
            public 
            MingleValueCastReactor.FieldTyper 
            fieldTyperFor( QualifiedTypeName qn,
                           ObjectPath< MingleIdentifier > path )
                throws MingleValueCastException
            {
                return MingleValueCastReactor.getDefaultFieldTyper();
            }
            
            public
            boolean
            inferStructFor( QualifiedTypeName qn )
            {
                return qn.equals( Mingle.QNAME_RESPONSE );
            }
    
            public
            MingleValue
            castAtomic( MingleValue mv,
                        AtomicTypeReference at,
                        ObjectPath< MingleIdentifier > path )
                throws MingleValueCastException
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
}
