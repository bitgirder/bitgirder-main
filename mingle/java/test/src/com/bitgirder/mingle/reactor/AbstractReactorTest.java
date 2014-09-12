package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.test.LabeledTestCall;

import com.bitgirder.lang.Strings;

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
        setLabel( getClass().getSimpleName() + ":" + 
            Strings.crossJoin( "=", ",", pairs ) );
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

//        final
//        void
//        feedSource( Object src,
//                    MingleReactor rct )
//            throws Exception
//        {
//            if ( src instanceof MingleValue ) {
//                MingleReactors.visitValue( (MingleValue) src, rct );
//            } 
//            else if ( src instanceof List ) 
//            {
//                List< MingleReactorEvent > evs = 
//                    Lang.castUnchecked( src );
//
//                feedReactorEvents( evs, rct );
//            } 
//            else {
//                state.failf( "unhandled source: %s", src );
//            }
//        }
}
