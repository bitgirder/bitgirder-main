package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.pipeline.PipelineInitializerContext;
import com.bitgirder.pipeline.Pipelines;

import java.util.Map;

public
final
class MingleReactors
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private MingleReactors() {}

    private final static MingleReactor DISCARD_REACTOR =
        new MingleReactor() {
            public void processEvent( MingleReactorEvent ev ) {}
        };

    public
    static
    MingleReactor
    discardReactor() 
    { 
        return DISCARD_REACTOR; 
    }

    private
    final
    static
    class DebugReactor
    implements MingleReactor
    {
        private final String prefix;

        private DebugReactor( String prefix ) { this.prefix = prefix; }

        public
        void
        processEvent( MingleReactorEvent ev )
        {
            String evStr = ev.inspect();
            if ( prefix != null ) evStr = prefix + " " + evStr;

            code( evStr );
        }
    }

    public
    static
    MingleReactor
    createDebugReactor()
    {
        return new DebugReactor( null );
    }

    public
    static
    MingleReactor
    createDebugReactor( String prefix )
    {
        inputs.notNull( prefix, "prefix" );
        return new DebugReactor( prefix );
    }

//    public
//    static
//    MingleReactorPipeline
//    createValueBuilderPipeline()
//        throws Exception
//    {
//        return new MingleReactorPipeline.Builder().
//            addReactor( StructuralCheck.create() ).
//            addReactor( MingleValueBuilder.create() ).
//            build();
//    }
//
//    private
//    static
//    void
//    visitEnd( MingleReactor rct,
//              MingleReactorEvent ev )
//        throws Exception
//    {
//        ev.setEnd();
//        rct.processEvent( ev );
//    }
//
//    private
//    static
//    void
//    visitList( MingleList ml,
//               MingleReactor rct,
//               MingleReactorEvent ev )
//        throws Exception
//    {
//        ev.setStartList( ml.type() );
//        rct.processEvent( ev );
//
//        for ( MingleValue mv : ml ) visitValue( mv, rct, ev );
//
//        visitEnd( rct, ev );
//    }
//
//    private
//    static
//    void
//    concludeVisitMap( MingleSymbolMap mp,
//                      MingleReactor rct,
//                      MingleReactorEvent ev )
//        throws Exception
//    {
//        for ( Map.Entry< MingleIdentifier, MingleValue > e : mp.entrySet() ) 
//        {
//            ev.setStartField( e.getKey() );
//            rct.processEvent( ev );
//
//            visitValue( e.getValue(), rct, ev );
//        }
//
//        visitEnd( rct, ev );
//    }
//
//    private
//    static
//    void
//    visitMap( MingleSymbolMap mp,
//              MingleReactor rct,
//              MingleReactorEvent ev )
//        throws Exception
//    {
//        ev.setStartMap();
//        rct.processEvent( ev );
//
//        concludeVisitMap( mp, rct, ev );
//    }
//
//    private
//    static
//    void
//    visitStruct( MingleStruct ms,
//                 MingleReactor rct,
//                 MingleReactorEvent ev )
//        throws Exception
//    {
//        ev.setStartStruct( ms.getType() );
//        rct.processEvent( ev );
//
//        concludeVisitMap( ms.getFields(), rct, ev );
//    }
//
//    private
//    static
//    void
//    visitScalar( MingleValue mv,
//                 MingleReactor rct,
//                 MingleReactorEvent ev )
//        throws Exception
//    {
//        ev.setValue( mv );
//        rct.processEvent( ev );
//    }
//
//    private
//    static
//    void
//    visitValue( MingleValue mv,
//                MingleReactor rct,
//                MingleReactorEvent ev )
//        throws Exception
//    {
//        if ( mv instanceof MingleList ) {
//            visitList( (MingleList) mv, rct, ev );
//        } else if ( mv instanceof MingleSymbolMap ) {
//            visitMap( (MingleSymbolMap) mv, rct, ev );
//        } else if ( mv instanceof MingleStruct ) {
//            visitStruct( (MingleStruct) mv, rct, ev );
//        } else {
//            visitScalar( mv, rct, ev );
//        }
//    }
//
//    public
//    static
//    void
//    visitValue( MingleValue mv,
//                MingleReactor rct )
//        throws Exception
//    {
//        inputs.notNull( mv, "mv" );
//        inputs.notNull( rct, "rct" );
//
//        visitValue( mv, rct, new MingleReactorEvent() );
//    }

    public
    static
    StructuralCheck
    ensureStructuralCheck( PipelineInitializerContext< Object > ctx )
    {
        inputs.notNull( ctx, "ctx" );
        
        StructuralCheck res = Pipelines.
            lastElementOfType( ctx.pipeline(), StructuralCheck.class );

        if ( res == null ) {
            res = StructuralCheck.create();
            ctx.addElement( res );
        }

        return res;
    }

//    public
//    static
//    MinglePathSettingProcessor
//    ensurePathSetter( PipelineInitializerContext< Object > ctx )
//    {
//        inputs.notNull( ctx, "ctx" );
//
//        ensureStructuralCheck( ctx );
//
//        MinglePathSettingProcessor res = Pipelines.lastElementOfType(
//            ctx.pipeline(), MinglePathSettingProcessor.class );
//
//        if ( res == null ) {
//            res = MinglePathSettingProcessor.create();
//            ctx.addElement( res );
//        }
//
//        return res;
//    }
//
//    static
//    abstract
//    class DepthTracker
//    {
//        private int depth;
//
//        protected
//        abstract
//        void
//        depthBecameOne();
//
//        public
//        final
//        void
//        update( MingleReactorEvent ev )
//        {
//            switch ( ev.type() ) {
//            case FIELD_START: return;
//            case MAP_START:
//            case STRUCT_START:
//            case LIST_START:
//                ++depth; break;
//            case END: --depth; break;
//            }
//
//            if ( depth == 1 ) depthBecameOne();
//        }
//    }
}
