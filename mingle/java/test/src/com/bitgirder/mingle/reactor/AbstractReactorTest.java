package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.test.LabeledTestCall;

import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.MingleValue;
import com.bitgirder.mingle.Mingle;

import java.util.List;

public
abstract
class AbstractReactorTest
extends LabeledTestCall
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    protected AbstractReactorTest( CharSequence nm ) { super( nm ); }

    protected AbstractReactorTest() { super(); }

    protected
    final
    void
    setLabel( Object... pairs )
    {
        String lbl = getClass().getSimpleName() + ":" + 
            Strings.crossJoin( "=", ",", pairs );
        
        super.setLabel( lbl );
    }

    protected
    final
    String
    sourceToString( Object src )
    {
        if ( src == null ) {
            return "null";
        } else if ( src instanceof MingleValue ) {
            return Mingle.inspect( (MingleValue) src ).toString();
        } else {
            return src.toString();
        }
    }

    protected
    final 
    void
    feedReactorEvents( List< MingleReactorEvent > evs,
                       MingleReactor rct )
        throws Exception
    {
        for ( MingleReactorEvent ev : evs ) rct.processEvent( ev );
    }

    protected
    final
    void
    feedValue( MingleValue mv,
               MingleReactor rct )
        throws Exception
    {
        inputs.notNull( mv, "mv" );
        inputs.notNull( rct, "rct" );

        MingleReactors.visitValue( mv, rct );
    }

    protected
    final
    void
    feedSource( Object src,
                MingleReactor rct )
        throws Exception
    {
        inputs.notNull( src, "src" );
        inputs.notNull( rct, "rct" );

        if ( src instanceof MingleValue ) {
            feedValue( (MingleValue) src, rct );
        } 
        else if ( src instanceof List ) {
            List< MingleReactorEvent > evs = Lang.castUnchecked( src );
            feedReactorEvents( evs, rct );
        } else {
            state.failf( "unhandled source: %s", src );
        }
    }
}
