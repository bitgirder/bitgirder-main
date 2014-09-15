package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.mingle.MingleValue;
import com.bitgirder.mingle.MingleStruct;
import com.bitgirder.mingle.MingleSymbolMap;
import com.bitgirder.mingle.MingleList;
import com.bitgirder.mingle.MingleIdentifier;

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

    private
    static
    void
    visitList( MingleList ml,
               EventSend es )
        throws Exception
    {
        es.startList( ml.type() );
        for ( MingleValue mv : ml ) visitValue( mv, es );
        es.end();
    }

    private
    static
    void
    visitFields( MingleSymbolMap mp,
                 EventSend es )
        throws Exception
    {
        for ( Map.Entry< MingleIdentifier, MingleValue > e : mp.entrySet() ) {
            es.startField( e.getKey() );
            visitValue( e.getValue(), es );
        }

        es.end();
    }

    private
    static
    void
    visitValue( MingleValue mv,
                EventSend es )
        throws Exception
    {
        if ( mv instanceof MingleList ) {
            visitList( (MingleList) mv, es );
        } else if ( mv instanceof MingleSymbolMap ) {
            MingleSymbolMap m = (MingleSymbolMap) mv;
            es.startMap();
            visitFields( m, es );
        } else if ( mv instanceof MingleStruct ) {
            MingleStruct ms = (MingleStruct) mv;
            es.startStruct( ms.getType() );
            visitFields( ms.getFields(), es );
        } else {
            es.value( mv );
        }
    }

    public
    static
    void
    visitValue( MingleValue mv,
                MingleReactor rct )
        throws Exception
    {
        inputs.notNull( mv, "mv" );
        inputs.notNull( rct, "rct" );

        visitValue( mv, EventSend.forReactor( rct ) );
    }

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
