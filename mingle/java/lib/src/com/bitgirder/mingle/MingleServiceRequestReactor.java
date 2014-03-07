package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.pipeline.PipelineInitializer;
import com.bitgirder.pipeline.PipelineInitializerContext;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ObjectPaths;

import java.util.List;

// this impl follows the go impl, so see there for more on its design
public
final
class MingleServiceRequestReactor
implements MingleValueReactor,
           PipelineInitializer< Object >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    static 
    enum TopFieldType
    {
        NONE,
        NAMESPACE,
        SERVICE,
        OPERATION,
        AUTHENTICATION,
        PARAMETERS;
    }
    
    private final Delegate del;

    private int depth;
    private TopFieldType topFld = TopFieldType.NONE;

    private MingleValueReactor evProc;

    private boolean hadParams;

    private
    MingleServiceRequestReactor( Delegate del )
    {
        this.del = del;
    }

    public
    void
    initialize( PipelineInitializerContext< Object > ctx )
    {
        MingleValueReactors.ensurePathSetter( ctx );
        
        ctx.addElement(
            new MingleValueCastReactor.Builder().
                setTargetType( Mingle.TYPE_REQUEST ).
                setDelegate( CAST_DELEGATE ).
                build()
        );
        
        ctx.addElement( MingleFieldOrderProcessor.create( ORDER_GETTER ) );
    }

    private
    void
    failInvalidValue( ObjectPath< MingleIdentifier > p,
                      CharSequence desc )
    {
        throw new MingleValueCastException( "invalid value: " + desc, p );
    }                     

    private
    void
    failInvalidValue( ObjectPath< MingleIdentifier > p,
                      MingleTypeReference typ )
    {
        failInvalidValue( p, typ.getExternalForm() );
    }

    private
    void
    updateDepth( MingleValueReactorEvent ev )
        throws Exception
    {
        switch ( ev.type() ) {
        case START_FIELD: return;
        case START_MAP:
        case START_STRUCT:
        case START_LIST:
            ++depth; break;
        case END: --depth; break;
        }

        if ( depth == 1 ) {
            evProc = null;
            topFld = TopFieldType.NONE;
        }
    }

    private
    MingleIdentifier
    checkFieldStart( MingleValueReactorEvent ev )
    {
        state.equalf( TopFieldType.NONE, topFld, "saw %s when top field is %s",
            ev.inspect(), topFld );

        return ev.field();
    }

    private
    void
    startField( MingleValueReactorEvent ev )
        throws Exception
    {
        MingleIdentifier fld = checkFieldStart( ev );

        if ( fld.equals( Mingle.ID_NAMESPACE ) ) {
            topFld = TopFieldType.NAMESPACE;
        } else if ( fld.equals( Mingle.ID_SERVICE ) ) {
            topFld = TopFieldType.SERVICE;
        } else if ( fld.equals( Mingle.ID_OPERATION ) ) {
            topFld = TopFieldType.OPERATION;
        } else if ( fld.equals( Mingle.ID_AUTHENTICATION ) ) {
            topFld = TopFieldType.AUTHENTICATION;
            evProc = del.getAuthenticationReactor( ev.path() );
        } else if ( fld.equals( Mingle.ID_PARAMETERS ) ) {
            topFld = TopFieldType.PARAMETERS;
            evProc = del.getParametersReactor( ev.path() );
            hadParams = true;
        } else {
            throw new MingleUnrecognizedFieldException( fld, ev.path() );
        }
    }

    private
    void
    startStruct( MingleValueReactorEvent ev )
        throws Exception
    {
        if ( topFld == TopFieldType.NONE ) 
        {
            state.equalf( Mingle.QNAME_REQUEST, ev.structType(),
                "unexpected service request type: %s", ev.structType() );

            return;
        }

        failInvalidValue( ev.path(), ev.structType().getExternalForm() );
    }

    private
    MingleIdentifier
    idVal( MingleValueReactorEvent ev )
        throws Exception
    {
        return Mingle.identifierForValue( ev.value(), ev.path() );
    }
    
    private
    void
    value( MingleValueReactorEvent ev )
        throws Exception
    {
        ObjectPath< MingleIdentifier > p = ev.path();

        switch ( topFld ) {
        case NAMESPACE: 
            MingleNamespace ns = Mingle.namespaceForValue( ev.value(), p );
            del.namespace( ns, p );
            break;
        case SERVICE: del.service( idVal( ev ), p ); break;
        case OPERATION: del.operation( idVal( ev ), p ); break;
        default: state.failf( "unhandled topFld: %s", topFld );
        }

        topFld = TopFieldType.NONE;
    }

    private
    void
    sendEmptyParams( ObjectPath< MingleIdentifier > path )
        throws Exception
    {
        MingleIdentifier fld = Mingle.ID_PARAMETERS;

        ObjectPath< MingleIdentifier > paramsPath = 
            ObjectPaths.descend( path, Mingle.ID_PARAMETERS );

        MingleValueReactor rct = new MingleValueReactorPipeline.Builder().
            addProcessor( MinglePathSettingProcessor.create( paramsPath ) ).
            addReactor( del.getParametersReactor( path ) ).
            build();

        MingleValueReactors.visitValue( MingleSymbolMap.empty(), rct );
    }

    private
    void
    end( MingleValueReactorEvent ev )
        throws Exception
    {
        if ( hadParams ) return;
        sendEmptyParams( ev.path() );
    }

    private
    void
    implProcessEvent( MingleValueReactorEvent ev )
        throws Exception
    {
        switch ( ev.type() ) {
        case START_FIELD: startField( ev ); return;
        case START_STRUCT: startStruct( ev ); return;
        case START_LIST: 
            failInvalidValue( ev.path(), Mingle.TYPE_VALUE_LIST ); return;
        case START_MAP: 
            failInvalidValue( ev.path(), Mingle.TYPE_SYMBOL_MAP ); return;
        case VALUE: value( ev ); return;
        case END: end( ev ); return;
        default: state.failf( "unhandled event: %s", ev.inspect() );
        }
    }

    public
    void
    processEvent( MingleValueReactorEvent ev )
        throws Exception
    {
        if ( evProc == null ) implProcessEvent( ev );
        else evProc.processEvent( ev );

        updateDepth( ev );
    }

    public
    static
    interface Delegate
    {
        public 
        void 
        namespace( MingleNamespace ns,
                   ObjectPath< MingleIdentifier > path ) 
            throws Exception;

        public
        void
        service( MingleIdentifier svc,
                 ObjectPath< MingleIdentifier > path )
            throws Exception;

        public
        void
        operation( MingleIdentifier op,
                   ObjectPath< MingleIdentifier > path )
            throws Exception;

        public
        MingleValueReactor
        getAuthenticationReactor( ObjectPath< MingleIdentifier > path )
            throws Exception;
 
        public
        MingleValueReactor
        getParametersReactor( ObjectPath< MingleIdentifier > path )
            throws Exception;
    }

    public
    static
    MingleServiceRequestReactor
    create( Delegate del )
    {
        inputs.notNull( del, "del" );
        return new MingleServiceRequestReactor( del );
    }
            
    private final static MingleValueReactorFieldOrder REQ_ORDER =
        new MingleValueReactorFieldOrder(
            Lang.< MingleValueReactorFieldSpecification >asList(
                new MingleValueReactorFieldSpecification(
                    Mingle.ID_NAMESPACE, 
                    true 
                ),
                new MingleValueReactorFieldSpecification(
                    Mingle.ID_SERVICE, 
                    true 
                ),
                new MingleValueReactorFieldSpecification(
                    Mingle.ID_OPERATION, 
                    true 
                ),
                new MingleValueReactorFieldSpecification(
                    Mingle.ID_AUTHENTICATION, 
                    false 
                ),
                new MingleValueReactorFieldSpecification(
                    Mingle.ID_PARAMETERS, 
                    false
                )
            )
        );

    private final static MingleFieldOrderProcessor.OrderGetter ORDER_GETTER =
        new MingleFieldOrderProcessor.OrderGetter()
        {
            public 
            MingleValueReactorFieldOrder
            fieldOrderFor( QualifiedTypeName qn )
            {
                if ( qn.equals( Mingle.QNAME_REQUEST ) ) return REQ_ORDER;
                return null;
            }
        };
    
    private final static MingleValueCastReactor.FieldTyper FIELD_TYPER =
        new MingleValueCastReactor.FieldTyper()
        {
            public 
            MingleTypeReference
            fieldTypeFor( MingleIdentifier fld,
                          ObjectPath< MingleIdentifier > path )
                throws MingleValueCastException
            {
                return fld.equals( Mingle.ID_PARAMETERS ) ?
                    Mingle.TYPE_SYMBOL_MAP : Mingle.TYPE_VALUE;
            }
        };

    private final static MingleValueCastReactor.Delegate CAST_DELEGATE =
        new MingleValueCastReactor.Delegate()
        {
            public 
            MingleValueCastReactor.FieldTyper 
            fieldTyperFor( QualifiedTypeName qn,
                           ObjectPath< MingleIdentifier > path )
            {
                return qn.equals( Mingle.QNAME_REQUEST ) ? FIELD_TYPER : null;
            }

            public
            boolean
            inferStructFor( QualifiedTypeName qn )
            {
                return qn.equals( Mingle.QNAME_REQUEST );
            }
    
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
}
