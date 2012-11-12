package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Inspector;
import com.bitgirder.lang.Inspectable;

import com.bitgirder.log.CodeLoggers;

import java.util.Deque;
import java.util.ArrayList;
import java.util.List;
import java.util.Iterator;

public
final
class RecursiveDescentParser< N, T >
implements Inspectable
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();
    private static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final Grammar< N, T > grammar;
    private final N goal;
    private final boolean logChildMatchedPre;
    private final boolean logChildUnmatchedEnter;
    private final boolean logTerminalConsume;

    // Typed as ArrayList to emphasize the need for random access
    private ArrayList< T > buffer = new ArrayList< T >();
 
    // Set to true when no more terminals will arrive
    private boolean eof;

    private final Deque< ProductionProcessor< N, T > > stack = Lang.newDeque();

    private ProductionMatcherState ms;
    private DerivationMatch< N, T > match;
    private int finalPos;

    private
    RecursiveDescentParser( Builder< N, T > b )
    {
        this.grammar = inputs.notNull( b.grammar, "grammar" );
        this.goal = inputs.notNull( b.goal, "goal" );
        this.logChildMatchedPre = b.logChildMatchedPre;
        this.logChildUnmatchedEnter = b.logChildUnmatchedEnter;
        this.logTerminalConsume = b.logTerminalConsume;

        reset();
    }

    public ProductionMatcherState getMatcherState() { return ms; }

    public 
    boolean 
    isMatching() 
    { 
        return ms == ProductionMatcherState.MATCHING; 
    }

    public boolean isMatched() { return ms == ProductionMatcherState.MATCHED; }

    public
    boolean
    isUnmatched()
    {
        return ms == ProductionMatcherState.UNMATCHED; 
    }

    // Impl note: returns an unmodifiable view of buffer, but not a copy, since
    // we will not be making any other modifications to buffer after any time
    // when it is legal to call this method
    public
    List< T >
    getUnconsumedInput()
    {
        state.isFalse( 
            ms == ProductionMatcherState.MATCHING,
            "This parser is still matching" );
        
        List< T > readOnly = Lang.unmodifiableList( buffer );
        return readOnly.subList( finalPos, readOnly.size() );
    }

    private
    void
    checkState( String methName,
                ProductionMatcherState expct )
    {
        state.isTrue( 
            expct == ms, methName, "called with parser in state", ms );
    }

    public
    int
    getConsumedTerminals()
    {
        checkState( "getConsumedTerminals()", ProductionMatcherState.MATCHED );
        state.nonnegativeI( finalPos, "finalPos" );

        return finalPos;
    }

    public
    void
    reset()
    {
        buffer.clear();
        eof = false;
        stack.clear();
        ms = ProductionMatcherState.MATCHING;
        match = null;
        finalPos = -1;

        DerivationMatcher< N > dm = DerivationMatcher.forHead( goal );
        pushDerivationMatcher( dm, 0 );
    }

    public
    DerivationMatch< N, T >
    getMatch()
    {
        checkState( "getMatch()", ProductionMatcherState.MATCHED );
        return state.notNull( match );
    }

    private
    static
    abstract
    class ProductionProcessor< N, T >
    implements Inspectable
    {
        private final int startIndx;

        private
        ProductionProcessor( int startIndx )
        {
            this.startIndx = startIndx;
        }

        final int startIndx() { return startIndx; }

        public
        Inspector
        accept( Inspector i )
        {
            return i.add( "startIndx", startIndx );
        }
    }

    private
    final
    static
    class TerminalProcessor< N, T >
    extends ProductionProcessor< N, T >
    {
        private final TerminalMatcher< T > tm;

        private 
        TerminalProcessor( TerminalMatcher< T > tm,
                           int startIndx ) 
        { 
            super( startIndx ); 

            this.tm = tm;
        }

        @Override
        public 
        Inspector 
        accept( Inspector i ) 
        { 
            return super.accept( i ).add( "tm", tm ); 
        }
    }

    private
    final
    static
    class UnionProcessor< N, T >
    extends ProductionProcessor< N, T >
    {
        private final UnionMatcher um;
        private final Iterator< ProductionMatcher > it;

        private int alt = -1;

        // match is the match currently in the running to win this union;
        // matchEnd is the index after which the parser should continue should
        // match win. Initializing matchEnd to -1 ensures that the first match
        // encountered is set regardless of the value of um.getMode()
        private UnionMatch< T > match;
        private int matchEnd = -1; 

        private
        UnionProcessor( UnionMatcher um,
                        int startIndx )
        {
            super( startIndx );

            this.um = um;
            this.it = um.getMatchers().iterator();
        }

        private 
        ProductionMatcher
        next()
        {
            state.isTrue( it.hasNext() );

            ++alt;
            return it.next();
        }

        @Override
        public 
        Inspector 
        accept( Inspector i ) 
        { 
            return super.accept( i ).
                         add( "match", match ).
                         add( "matchEnd", matchEnd ).
                         add( "alt", alt );
        }
    }

    private
    final
    static
    class DerivationProcessor< N, T >
    extends ProductionProcessor< N, T >
    {
        private final N head;

        private
        DerivationProcessor( N head,
                             int startIndx )
        {
            super( startIndx );

            this.head = head;
        }

        @Override
        public 
        Inspector 
        accept( Inspector i ) 
        { 
            return super.accept( i ).add( "head", head ); 
        }
    }

    private
    final
    static
    class SequenceProcessor< N, T >
    extends ProductionProcessor< N, T >
    {
        private final Iterator< ProductionMatcher > it;

        private final List< ProductionMatch< T > > matches;

        private
        SequenceProcessor( SequenceMatcher sm,
                           int startIndx )
        {
            super( startIndx );

            List< ProductionMatcher > seq = sm.getSequence();

            this.it = seq.iterator();
            this.matches = Lang.newList( seq.size() );
        }

        private
        ProductionMatcher
        next()
        {
            state.isTrue( it.hasNext() );
            return it.next();
        }

        @Override
        public 
        Inspector 
        accept( Inspector i ) 
        { 
            return 
                super.accept( i ).
                      add( "matches", matches ).
                      add( "pos", matches.size() );
        }
    }

    private
    final
    static
    class QuantifierProcessor< N, T >
    extends ProductionProcessor< N, T >
    {
        private final QuantifierMatcher qm;

        private final List< ProductionMatch< T > > matches = Lang.newList();

        private int lastStartIndx;

        private 
        QuantifierProcessor( QuantifierMatcher qm,
                             int startIndx ) 
        { 
            super( startIndx );
            this.qm = qm; 
            this.lastStartIndx = startIndx;
        }

        @Override
        public
        Inspector
        accept( Inspector i )
        {
            return super.accept( i ).
                         add( "minIncl", qm.getMinInclusive() ).
                         add( "maxIncl", qm.getMaxInclusive() ).
                         add( "lastStartIndx", lastStartIndx ).
                         add( "matches", matches );
        }
    }

    private 
    T 
    terminal( int logicalIndx ) 
    { 
        return buffer.get( logicalIndx  ); 
    }

    private
    < Z >
    Z
    castUnchecked( ProductionProcessor p )
    {
        @SuppressWarnings( "unchecked" )
        Z res = (Z) p;

        return res;
    }

    private
    DerivationProcessor< N, T >
    asDerivation( ProductionProcessor pp )
    {
        return this.< DerivationProcessor< N, T > >castUnchecked( pp );
    }

    private
    UnionProcessor< N, T >
    asUnion( ProductionProcessor pp )
    {
        return this.< UnionProcessor< N, T > >castUnchecked( pp );
    }

    private
    SequenceProcessor< N, T >
    asSeq( ProductionProcessor pp )
    {
        return this.< SequenceProcessor< N, T > >castUnchecked( pp );
    }

    private
    QuantifierProcessor< N, T >
    asQp( ProductionProcessor pp )
    {
        return this.< QuantifierProcessor< N, T > >castUnchecked( pp );
    }

    private
    TerminalProcessor< N, T >
    asTp( ProductionProcessor pp )
    {
        return this.< TerminalProcessor< N, T > >castUnchecked( pp );
    }

    private
    void
    pushMatcher( ProductionMatcher pm,
                 int startIndx )
    {
        if ( pm instanceof SequenceMatcher )
        {
            pushSequenceMatcher( (SequenceMatcher) pm, startIndx );
        }
        else if ( pm instanceof QuantifierMatcher )
        {
            pushQuantifier( (QuantifierMatcher) pm, startIndx );
        }
        else if ( pm instanceof UnionMatcher )
        {
            pushUnionMatcher( (UnionMatcher) pm, startIndx );
        }
        else if ( pm instanceof DerivationMatcher )
        {
            pushDerivationMatcher( (DerivationMatcher) pm, startIndx );
        }
        else if ( pm instanceof TerminalMatcher )
        {
            pushTerminalMatcher( (TerminalMatcher) pm, startIndx );
        }
        else state.fail( "Unexpected production matcher:", pm );
    }

    private
    void
    pushDerivationMatcher( DerivationMatcher dm,
                           int startIndx )
    {
        N head = grammar.headType().cast( dm.getHead() );
 
        ProductionMatcher pm = grammar.getProduction( head );

        DerivationProcessor< N, T > dp = 
            new DerivationProcessor< N, T >( head, startIndx );
 
        stack.addFirst( dp );
        pushMatcher( pm, startIndx );
    }

    private
    void
    pushQuantifier( QuantifierMatcher qm,
                    int startIndx )
    {
        QuantifierProcessor< N, T > qp = 
            new QuantifierProcessor< N, T >( qm, startIndx );

        stack.addFirst( qp );
        pushNext( qp, qp.startIndx() );
    }

    private
    void
    pushTerminalMatcher( TerminalMatcher tm,
                         int startIndx )
    {
        @SuppressWarnings( "unchecked" )
        TerminalMatcher< T > castTm = (TerminalMatcher< T >) tm;

        stack.addFirst( new TerminalProcessor< N, T >( castTm, startIndx ) );
    }

    private
    void
    pushUnionMatcher( UnionMatcher um,
                      int startIndx )
    {
        UnionProcessor< N, T > up = new UnionProcessor< N, T >( um, startIndx );

        stack.addFirst( up );
        pushMatcher( up.next(), startIndx );
    }

    private
    void
    pushSequenceMatcher( SequenceMatcher sm,
                         int startIndx )
    {
        SequenceProcessor< N, T > sp = 
            new SequenceProcessor< N, T >( sm, startIndx );

        stack.addFirst( sp );
        pushNext( sp, startIndx );
    }

    private
    void
    pushNext( SequenceProcessor< N, T > sp,
              int startIndx )
    {
        pushMatcher( sp.next(), startIndx );
    }

    private
    void
    pushNext( QuantifierProcessor< N, T > qp,
              int startIndx )
    {
        qp.lastStartIndx = startIndx;
        pushMatcher( qp.qm.getMatcher(), startIndx );
    }

    // Take the first derivation encountered from head
    private
    void
    childMatched( DerivationProcessor< N, T > dp,
                  ProductionMatch< T > m,
                  int nextIndx )
    {
        DerivationMatch< N, T > dm = DerivationMatch.create( dp.head, m );

        stack.removeFirst();
        childMatched( dm, nextIndx );
    }

    private
    void
    childMatched( SequenceProcessor< N, T > sp,
                  ProductionMatch< T > m,
                  int nextIndx )
    {
        sp.matches.add( m );

        if ( sp.it.hasNext() ) pushNext( sp, nextIndx );
        else
        {
            SequenceMatch< T > sm = SequenceMatch.create( sp.matches );
            stack.removeFirst();
            childMatched( sm, nextIndx );
        }
    }

    private
    UnionMatch< T >
    createMatchAndUpdate( UnionProcessor< N, T > up,
                          ProductionMatch< T > m,
                          int nextIndx )
    {
        UnionMatch< T > match = UnionMatch.create( m, up.alt );

        if ( nextIndx > up.matchEnd ) 
        {
            up.match = match;
            up.matchEnd = nextIndx;
        }

        return match;
    }

    private
    boolean
    updateUnionMatch( UnionProcessor< N, T > up,
                      ProductionMatch< T > m,
                      int nextIndx )
    {
        boolean res;

        UnionMatch< T > match = createMatchAndUpdate( up, m, nextIndx );
 
        switch ( up.um.getMode() )
        {
            case FIRST_WINS: res = true; break;

            case LONGEST_WINS: res = ! up.it.hasNext(); break;
            
            default: 
                throw state.createFail( "Unexpected mode:", up.um.getMode() );
        }

        return res;
    }

    // First match wins, currently. Need to keep behavior here in sync with
    // childUnmatched when we change to allow longest-wins.
    private
    void
    childMatched( UnionProcessor< N, T > up,
                  ProductionMatch< T > m,
                  int nextIndx )
    {
        if ( updateUnionMatch( up, m, nextIndx ) )
        {
            stack.removeFirst();
            childMatched( up.match, up.matchEnd );
        }
        else pushNextUnion( up );
    }

    private
    void
    sendQuantifierMatch( QuantifierProcessor< N, T > qp,
                         int nextIndx )
    {
        QuantifierMatch< T > qm = QuantifierMatch.create( qp.matches );

        stack.removeFirst();
        childMatched( qm, nextIndx );
    }

    private
    void
    childMatchedEmpty( QuantifierProcessor< N, T > qp,
                       ProductionMatch< T > m,
                       int nextIndx )
    {
        if ( qp.matches.size() >= qp.qm.getMinInclusive() )
        {
            sendQuantifierMatch( qp, nextIndx ); 
        }
        else
        {
            stack.removeFirst();
            childUnmatched();
        }
    }

    private
    void
    childMatchedNonEmpty( QuantifierProcessor< N, T > qp,
                          ProductionMatch< T > m,
                          int nextIndx )
    {
        qp.matches.add( m );

        if ( qp.matches.size() == qp.qm.getMaxInclusive() )
        {
            sendQuantifierMatch( qp, nextIndx );
        }
        else if ( qp.matches.size() >= qp.qm.getMinInclusive() )
        {
            if ( eof && nextIndx == buffer.size() )
            {
                sendQuantifierMatch( qp, nextIndx );
            }
            else pushNext( qp, nextIndx );
        }
        else 
        {
            if ( eof )
            {
                stack.removeFirst();
                childUnmatched();
            }
            else pushNext( qp, nextIndx );
        }
    }

    private
    void
    childMatched( QuantifierProcessor< N, T > qp,
                  ProductionMatch< T > m,
                  int nextIndx )
    {
        if ( qp.lastStartIndx == nextIndx ) 
        {
            childMatchedEmpty( qp, m, nextIndx );
        }
        else childMatchedNonEmpty( qp, m, nextIndx );
    }

    private
    void
    setMatch( ProductionMatch< T > m,
              int nextIndx )
    {
        ms = ProductionMatcherState.MATCHED;

        @SuppressWarnings( "unchecked" )
        DerivationMatch< N, T > cast = (DerivationMatch< N, T >) m;
        match = cast;
 
        finalPos = nextIndx;
    }

    private
    void
    logChildMatchedPre( ProductionMatch< T > m,
                   int nextIndx )
    {
        if ( logChildMatchedPre )
        {
            code( 
                "Child matched", Strings.inspect( m ), 
                "; nextIndx:", nextIndx );
        }
    }

    private
    void
    childMatched( ProductionMatch< T > m,
                  int nextIndx )
    {
        if ( stack.isEmpty() ) setMatch( m, nextIndx );
        else
        {
            ProductionProcessor pp = (ProductionProcessor) stack.peekFirst();
            
            logChildMatchedPre( m, nextIndx );
 
            if ( pp instanceof DerivationProcessor )
            {
                childMatched( asDerivation( pp ), m, nextIndx );
            }
            else if ( pp instanceof UnionProcessor )
            {
                childMatched( asUnion( pp ), m, nextIndx );
            }
            else if ( pp instanceof SequenceProcessor )
            {
                childMatched( asSeq( pp ), m, nextIndx );
            }
            else if ( pp instanceof QuantifierProcessor )
            {
                childMatched( asQp( pp ), m, nextIndx );
            }
            else state.fail( "Unexpected processor:", pp );
        }
    }

    private
    void
    terminalMatched( T terminal,
                     int nextIndx )
    {
        TerminalMatch< T > m = TerminalMatch.forTerminal( terminal );

        ProductionProcessor pp = (ProductionProcessor) stack.peekFirst();
        childMatched( m, nextIndx );
    }

    private
    void
    pushNextUnion( UnionProcessor< N, T > up )
    {
        pushMatcher( up.next(), up.startIndx() );
    }

    private
    void
    childUnmatched( UnionProcessor< N, T > up )
    {
        if ( up.it.hasNext() ) pushNextUnion( up );
        else
        {
            stack.removeFirst();

            if ( up.match == null ) childUnmatched();
            else childMatched( up.match, up.matchEnd );
        }
    }

    private
    void
    childUnmatched( DerivationProcessor< N, T > dp )
    {
        stack.removeFirst();
        childUnmatched();
    }

    private
    void
    childUnmatched( QuantifierProcessor< N, T > qp )
    {
        // This quantifier processor is done at this point, whether it turns out
        // to have matched its target quantity or not, so we remove it first
        stack.removeFirst();

        if ( qp.matches.size() >= qp.qm.getMinInclusive() )
        {
            QuantifierMatch< T > qm = QuantifierMatch.create( qp.matches );
            childMatched( qm, qp.lastStartIndx );
        }
        else childUnmatched();
    }

    private
    void
    logChildUnmatchedEnter()
    {
        if ( logChildUnmatchedEnter )
        {
            code( "Entering childUnmatched, stack:", Strings.inspect( stack ) );
        }
    }

    private
    void
    childUnmatched()
    {
        logChildUnmatchedEnter();

        if ( stack.isEmpty() ) ms = ProductionMatcherState.UNMATCHED;
        else
        {
            ProductionProcessor pp = (ProductionProcessor) stack.peekFirst();
 
            if ( pp instanceof SequenceProcessor )
            {
                stack.removeFirst();
                childUnmatched();
            }
            else if ( pp instanceof DerivationProcessor ) 
            {
                childUnmatched( asDerivation( pp ) );
            }
            else if ( pp instanceof UnionProcessor )
            {
                childUnmatched( asUnion( pp ) );
            }
            else if ( pp instanceof QuantifierProcessor )
            {
                childUnmatched( asQp( pp ) );
            }
            else state.fail( "Unexpected processor:", pp );
        }
    }

    private
    void
    consumeOneTerminal()
    {
        @SuppressWarnings( "unchecked" )
        TerminalProcessor< N, T > tp = 
            (TerminalProcessor< N, T >) stack.removeFirst();

        int indx = tp.startIndx();

        T terminal = terminal( indx );
 
        if ( logTerminalConsume )
        {
            code( 
                "Feeding terminal", terminal, "(index " + indx + ");",
                "parser stack:", Strings.inspect( stack ) );
        }

        if ( tp.tm.isMatch( terminal ) ) terminalMatched( terminal, indx + 1 );
        else childUnmatched();
    }

    private
    void
    consumeTerminals()
    {
        while ( ( ! stack.isEmpty() ) &&
                stack.peekFirst().startIndx() < buffer.size() )
        {
            consumeOneTerminal();
        }
    }

    public
    void
    consumeTerminal( T terminal )
    {
        inputs.notNull( terminal, "terminal" );
        checkState( "consumeTerminal()", ProductionMatcherState.MATCHING );
        
        buffer.add( terminal );
        state.equalInt( buffer.size() - 1, stack.peekFirst().startIndx() );

        consumeTerminals();
    }

    public
    void
    complete()
    {
        checkState( "complete()", ProductionMatcherState.MATCHING );

        eof = true;

        if ( stack.isEmpty() ) ms = ProductionMatcherState.MATCHED;
        else
        {
            while ( ! stack.isEmpty() )
            {
                TerminalProcessor< N, T > tp = asTp( stack.peekFirst() );
    
                if ( tp.startIndx() == buffer.size() ) 
                {
                    stack.removeFirst();
                    childUnmatched();
                }
                else consumeTerminals();
            }
        }
    }

    public
    Inspector
    accept( Inspector i )
    {
        return i.add( "stack", stack ).
                 add( "ms", ms );
    }

    public
    final
    static
    class Builder< N, T >
    {
        private Grammar< N, T > grammar;
        private N goal;
        private boolean logChildMatchedPre;
        private boolean logChildUnmatchedEnter;
        private boolean logTerminalConsume;

        Builder() {}

        Builder< N, T >
        setGrammar( Grammar< N, T > grammar )
        {
            this.grammar = inputs.notNull( grammar, "grammar" );
            return this;
        }

        public
        Builder< N, T >
        setGoal( N goal )
        {
            this.goal = inputs.notNull( goal, "goal" );
            return this;
        }

        public
        Builder< N, T >
        setLogChildMatchedPre( boolean logChildMatchedPre )
        {
            this.logChildMatchedPre = logChildMatchedPre;
            return this;
        }

        public
        Builder< N, T >
        setLogChildUnmatchedEnter( boolean logChildUnmatchedEnter )
        {
            this.logChildUnmatchedEnter = logChildUnmatchedEnter;
            return this;
        }

        public
        Builder< N, T >
        setLogMatchEvents( boolean logMatchEvents )
        {
            setLogChildMatchedPre( logMatchEvents );
            setLogChildUnmatchedEnter( logMatchEvents );

            return this;
        }

        public
        Builder< N, T >
        setLogTerminalConsume( boolean logTerminalConsume )
        {
            this.logTerminalConsume = logTerminalConsume;
            return this;
        }
 
        public
        RecursiveDescentParser< N, T >
        build()
        {
            return new RecursiveDescentParser< N, T >( this );
        }
    }
}
