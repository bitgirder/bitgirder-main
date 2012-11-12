package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.Set;
import java.util.Iterator;

final
class Grammars
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private Grammars() {}

    private
    static
    < N >
    void
    addFollowNonTerminals( DerivationMatcher dm,
                           Set< N > follow,
                           Grammar< N, ? > g )
    {
        @SuppressWarnings( "unchecked" )
        DerivationMatcher< N > dm2 = (DerivationMatcher< N >) dm;

        N head = dm2.getHead();

        if ( ! follow.contains( head ) )
        {
            follow.add( head );
            addFollowNonTerminals( head, follow, g );
        }
    }

    private
    static
    < N >
    void
    addFollowNonTerminals( UnionMatcher um,
                           Set< N > follow,
                           Grammar< N, ? > g )
    {
        for ( ProductionMatcher pm : um.getMatchers() )
        {
            addFollowNonTerminals( pm, follow, g );
        }
    }

    private
    static
    < N >
    void
    addFollowNonTerminals( SequenceMatcher sm,
                           Set< N > follow,
                           Grammar< N, ? > g )
    {
        Iterator< ProductionMatcher > it = sm.getSequence().iterator();

        for ( boolean loop = true; it.hasNext() && loop; )
        {
            ProductionMatcher pm = it.next();
            addFollowNonTerminals( pm, follow, g );

            loop =
                pm instanceof QuantifierMatcher && 
                ( (QuantifierMatcher) pm ).getMinInclusive() == 0;
        }
    }

    private
    static
    < N >
    void
    addFollowNonTerminals( QuantifierMatcher qm,
                           Set< N > follow,
                           Grammar< N, ? > g )
    {
        addFollowNonTerminals( qm.getMatcher(), follow, g );
    }

    private
    static
    < N >
    void
    addFollowNonTerminals( ProductionMatcher pm,
                           Set< N > follow,
                           Grammar< N, ? > g )
    {
        if ( pm instanceof TerminalMatcher );
        else if ( pm instanceof DerivationMatcher )
        {
            addFollowNonTerminals( (DerivationMatcher) pm, follow, g );
        }
        else if ( pm instanceof UnionMatcher )
        {
            addFollowNonTerminals( (UnionMatcher) pm, follow, g );
        }
        else if ( pm instanceof SequenceMatcher )
        {
            addFollowNonTerminals( (SequenceMatcher) pm, follow, g );
        }
        else if ( pm instanceof QuantifierMatcher )
        {
            addFollowNonTerminals( (QuantifierMatcher) pm, follow, g );
        }
        else throw state.createFail( "Unrecognized matcher:", pm );
    }

    private
    static
    < N >
    void
    addFollowNonTerminals( N head,
                           Set< N > follow,
                           Grammar< N, ? > g )
    {
        addFollowNonTerminals( g.getProduction( head ), follow, g );
    }

    // This is a very simplistic, ineffecient, and
    // not-terribly-helpful-from-an-error-reporting-standpoint implementation of
    // left recursion detection: we do a walk from each head in turn and look
    // for it to show up in its own follow set. Eventually we'll likely want to
    // make this more meaningful (error messages) and efficient.
    static
    < N >
    void
    assertNoLeftRecursion( Grammar< N, ? > g )
    {
        for ( N head : g.getNonTerminals() ) 
        {
            Set< N > reachable = Lang.newSet();
            addFollowNonTerminals( head, reachable, g );
 
            if ( reachable.contains( head ) )
            {
                throw new LeftRecursiveGrammarException( head );
            }
        }
    }

////    static
////    < N >
////    void
////    assertNoEmptySentences( Grammar< N, ? > g )
////    {
////        for ( N head : g.getNonTerminals() )
////        {
////            assertNoEmptySentences
//    
//    private
//    static
//    < N >
//    boolean
//    canDeriveEmpty( UnionMatcher um,
//                    Grammar< N, ? > g,
//                    Set< N > visited )
//    {
//        boolean res = false;
//
//        Iterator< ProductionMatcher > it = um.getMatchers().iterator();
//
//        while ( it.hasNext() && ! res )
//        {
//            res = canDeriveEmpty( it.next(), g, visited );
//        }
//    
//    private
//    static
//    < N >
//    boolean
//    canDeriveEmpty( ProductionMatcher pm,
//                    Grammar< N, ? > g,
//                    Set< N > visited )
//    {
//        if ( pm instanceof UnionMatcher )
//        {
//            return canDeriveEmpty( (UnionMatcher) pm, g, visited );
//        }
//        else if ( pm instanceof SequenceMatcher )
//        {
//            return canDeriveEmpty( (SequenceMatcher) pm, g, visited );
//        }
//        else if ( pm instanceof QuantifierMatcher )
//        {
//            return canDeriveEmpty( (QuantifierMatcher) qm, g, visited );
//        }
//        else if ( pm instanceof DerivationMatcher )
//        {
//            return canDeriveEmpty( (DerivationMatcher) dm, g, visited );
//        }
//        else if ( pm instanceof TerminalMatcher );
//        else throw state.createFail( "Unrecognized matcher:", m );
//    }
//    
//    private
//    static
//    < N >
//    Map< N, EpsilonClosure< N > >
//    getEpsilonClosures( Grammar< N, ? > g )
//    {
//        Map< N, EpsilonClosure< N > > res = Lang.newMap();
//
//        for ( N head : g.getNonTerminals() )
//        {
//            res.put( head, getEpsilonClosure( 
//        initEpsilonClosure
//        for ( 
}
