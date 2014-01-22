package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.util.Map;

public
final
class MingleValueReactors
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private MingleValueReactors() {}

    private final static MingleValueReactor DISCARD_REACTOR =
        new MingleValueReactor() {
            public void processEvent( MingleValueReactorEvent ev ) {}
        };

    public
    static
    MingleValueReactor
    discardReactor() 
    { 
        return DISCARD_REACTOR; 
    }

    public
    static
    MingleValueReactorPipeline
    createValueBuilderPipeline()
    {
        return new MingleValueReactorPipeline.Builder().
            addReactor( MingleValueBuilder.create() ).
            build();
    }

    private
    static
    void
    visitEnd( MingleValueReactor rct,
              MingleValueReactorEvent ev )
        throws Exception
    {
        ev.setEnd();
        rct.processEvent( ev );
    }

    private
    static
    void
    visitList( MingleList ml,
               MingleValueReactor rct,
               MingleValueReactorEvent ev )
        throws Exception
    {
        ev.setStartList();
        rct.processEvent( ev );

        for ( MingleValue mv : ml ) visitValue( mv, rct, ev );

        visitEnd( rct, ev );
    }

    private
    static
    void
    concludeVisitMap( MingleSymbolMap mp,
                      MingleValueReactor rct,
                      MingleValueReactorEvent ev )
        throws Exception
    {
        for ( Map.Entry< MingleIdentifier, MingleValue > e : mp.entrySet() ) 
        {
            ev.setStartField( e.getKey() );
            rct.processEvent( ev );

            visitValue( e.getValue(), rct, ev );
        }

        visitEnd( rct, ev );
    }

    private
    static
    void
    visitMap( MingleSymbolMap mp,
              MingleValueReactor rct,
              MingleValueReactorEvent ev )
        throws Exception
    {
        ev.setStartMap();
        rct.processEvent( ev );

        concludeVisitMap( mp, rct, ev );
    }

    private
    static
    void
    visitStruct( MingleStruct ms,
                 MingleValueReactor rct,
                 MingleValueReactorEvent ev )
        throws Exception
    {
        ev.setStartStruct( ms.getType() );
        rct.processEvent( ev );

        concludeVisitMap( ms.getFields(), rct, ev );
    }

    private
    static
    void
    visitScalar( MingleValue mv,
                 MingleValueReactor rct,
                 MingleValueReactorEvent ev )
        throws Exception
    {
        ev.setValue( mv );
        rct.processEvent( ev );
    }

    private
    static
    void
    visitValue( MingleValue mv,
                MingleValueReactor rct,
                MingleValueReactorEvent ev )
        throws Exception
    {
        if ( mv instanceof MingleList ) {
            visitList( (MingleList) mv, rct, ev );
        } else if ( mv instanceof MingleSymbolMap ) {
            visitMap( (MingleSymbolMap) mv, rct, ev );
        } else if ( mv instanceof MingleStruct ) {
            visitStruct( (MingleStruct) mv, rct, ev );
        } else {
            visitScalar( mv, rct, ev );
        }
    }

    public
    static
    void
    visitValue( MingleValue mv,
                MingleValueReactor rct )
        throws Exception
    {
        inputs.notNull( mv, "mv" );
        inputs.notNull( rct, "rct" );

        visitValue( mv, rct, new MingleValueReactorEvent() );
    }
}