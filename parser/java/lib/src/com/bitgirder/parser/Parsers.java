package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;
import java.util.Set;

public
final
class Parsers
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private Parsers() {}

    private
    static
    < U >
    U
    castUnchecked( Object inst )
    {
        @SuppressWarnings( "unchecked" )
        U res = (U) inst;

        return res;
    }

    public
    static
    < T >
    SequenceMatch< T >
    castSequenceMatch( ProductionMatch< T > pm )
    {
        return Parsers.< SequenceMatch< T > >castUnchecked( pm );
    }

    public
    static
    < T >
    QuantifierMatch< T >
    castQuantifierMatch( ProductionMatch< T > pm )
    {
        return Parsers.< QuantifierMatch< T > >castUnchecked( pm );
    }

    public
    static
    < T >
    UnionMatch< T >
    castUnionMatch( ProductionMatch< T > pm )
    {
        return Parsers.< UnionMatch< T > >castUnchecked( pm );
    }

    public
    static
    < N, T >
    DerivationMatch< N, T >
    castDerivationMatch( ProductionMatch< T > pm )
    {
        return Parsers.< DerivationMatch< N, T > >castUnchecked( pm );
    }

    public
    static
    < T >
    TerminalMatch< T >
    castTerminalMatch( ProductionMatch< T > pm )
    {
        return Parsers.< TerminalMatch< T > >castUnchecked( pm );
    }

    public
    static
    < T >
    ProductionMatch< T >
    extractUnaryMatch( ProductionMatch< T > pm )
    {
        QuantifierMatch< T > qm = castQuantifierMatch( pm );
        
        int sz = qm.size();

        switch ( sz )
        {
            case 0: return null;
            case 1: return qm.get( 0 );
            
            default:
                throw state.createFail( 
                    "Expected a unary match, found", sz, "elements" );
        }
    }

    public
    static
    < T >
    SequenceMatch< T >
    extractUnarySequenceMatch( ProductionMatch< T > pm )
    {
        return castSequenceMatch( extractUnaryMatch( pm ) );
    }

    public
    static
    < N, T >
    DerivationMatch< N, T >
    extractUnaryDerivationMatch( ProductionMatch< T > pm )
    {
        return castDerivationMatch( extractUnaryMatch( pm ) );
    }

    public
    static
    < T, U extends T >
    U
    extractTerminal( Class< U > termCls,
                     ProductionMatch< T > termMatch )
    {
        TerminalMatch< T > tm = castTerminalMatch( termMatch );
        return termCls.cast( tm.getTerminal() );
    }

    // Visitors will likely become a standalone interface eventually, not
    // enclosed in here

    // non-terminal visit methods return true if the calling code should
    // continue to visit the match's sub-matches.
    private
    static
    interface Visitor< N, T >
    {
        public
        void
        visitTerminal( T terminal );

        public
        boolean
        visitSequence( SequenceMatch< T > sm );

        public
        boolean
        visitQuantifier( QuantifierMatch< T > qm );

        public
        boolean
        visitUnion( UnionMatch< T > um );

        public
        boolean
        visitDerivation( DerivationMatch< N, T > dm );
    }
    
    private
    static
    abstract
    class AbstractVisitor< N, T >
    implements Visitor< N, T >
    {
        public void visitTerminal( T terminal ) {}

        public boolean visitSequence( SequenceMatch< T > sm ) { return true; }

        public
        boolean
        visitQuantifier( QuantifierMatch< T > qm ) 
        { 
            return true; 
        }

        public boolean visitUnion( UnionMatch< T > um ) { return true; }

        public
        boolean
        visitDerivation( DerivationMatch< N, T > dm ) 
        { 
            return true; 
        }
    }

    private
    static
    < T >
    void
    visitTerminal( TerminalMatch< T > tm,
                   Visitor< ?, T > v )
    {
        v.visitTerminal( tm.getTerminal() );
    }

    private
    static
    < N, T >
    void
    visitList( Iterable< ProductionMatch< T > > list,
               Visitor< N, T > v )
    {
        for ( ProductionMatch< T > pm : list ) visit( pm, v );
    }

    private
    static
    < N, T >
    void
    visitQuantifier( QuantifierMatch< T > qm,
                     Visitor< N, T > v )
    {
        if ( v.visitQuantifier( qm ) ) visitList( qm, v );
    }

    private
    static
    < N, T >
    void
    visitSequence( SequenceMatch< T > sm,
                   Visitor< N, T > v )
    {
        if ( v.visitSequence( sm ) ) visitList( sm, v );
    }

    private
    static
    < N, T >
    void
    visitUnion( UnionMatch< T > um,
                Visitor< N, T > v )
    {
        if ( v.visitUnion( um ) ) visit( um.getMatch(), v );
    }

    private
    static
    < N, T >
    void
    visitDerivation( DerivationMatch< N, T > dm,
                     Visitor< N, T > v )
    {
        if ( v.visitDerivation( dm ) ) visit( dm.getMatch(), v );
    }

    private
    static
    < N, T >
    void
    visit( ProductionMatch< T > pm,
           Visitor< N, T > v )
    {
        if ( pm instanceof TerminalMatch )
        {
            visitTerminal( castTerminalMatch( pm ), v );
        }
        else if ( pm instanceof QuantifierMatch )
        {
            visitQuantifier( castQuantifierMatch( pm ), v );
        }
        else if ( pm instanceof SequenceMatch )
        {
            visitSequence( castSequenceMatch( pm ), v );
        }
        else if ( pm instanceof UnionMatch ) 
        {
            visitUnion( castUnionMatch( pm ), v );
        }
        else if ( pm instanceof DerivationMatch )
        {
            visitDerivation( Parsers.< N, T >castDerivationMatch( pm ), v );
        }
        else throw state.createFail( "Unrecognized match:", pm );
    }

    private
    final
    static
    class DerivationExtractor< N, T >
    extends AbstractVisitor< N, T >
    {
        private final Set< N > heads;

        private final List< DerivationMatch< N, T > > matches = Lang.newList();

        private DerivationExtractor( Set< N > heads ) { this.heads = heads; }

        @Override
        public
        boolean
        visitDerivation( DerivationMatch< N, T > dm )
        {
            boolean res;

            if ( heads.contains( dm.getHead() ) )
            {
                matches.add( dm );
                res = false;
            }
            else res = true;

            return res;
        }
    }

    public
    static
    < N, T >
    List< DerivationMatch< N, T > >
    extractDerivations( ProductionMatch< T > pm,
                        Set< N > heads )
    {
        inputs.notNull( pm, "pm" );
        inputs.noneNull( heads, "heads" );

        DerivationExtractor< N, T > de = 
            new DerivationExtractor< N, T >( heads );

        visit( pm, de );

        return Lang.unmodifiableCopy( de.matches );
    }

    @SafeVarargs
    public
    static
    < N, T >
    List< DerivationMatch< N, T > >
    extractDerivations( ProductionMatch< T > pm,
                        N... heads )
    {
        inputs.noneNull( heads, "heads" );

        Set< N > set = Lang.newSet();
        for ( N head : heads ) set.add( head );

        return extractDerivations( pm, set );
    }

    private
    final
    static
    class TerminalExtractor< N, T, U extends T >
    extends AbstractVisitor< N, T >
    {
        private final Class< U > cls;
        private final List< U > matches = Lang.newList();

        private TerminalExtractor( Class< U > cls ) { this.cls = cls; }

        @Override
        public
        void
        visitTerminal( T terminal )
        {
            if ( cls.isInstance( terminal ) ) 
            {
                matches.add( cls.cast( terminal ) );
            }
        }
    }

    public
    static
    < N, T, U extends T >
    List< U >
    extractTerminals( ProductionMatch< T > pm,
                      Class< U > cls )
    {
        inputs.notNull( pm, "pm" );
        inputs.notNull( cls, "cls" );

        TerminalExtractor< N, T, U > te =
            new TerminalExtractor< N, T, U >( cls );
        
        visit( pm, te );

        return Lang.unmodifiableCopy( te.matches );
    }

    // This uses Object as the non-terminal type, which leads to unsafe casts in
    // the visit, but these are okay because we never attempt to access any of
    // the actual non-terminals
    private
    final
    static
    class StringBuildVisitor
    extends AbstractVisitor< Object, Character >
    {
        private final StringBuilder sb = new StringBuilder();

        @Override
        public
        void
        visitTerminal( Character terminal )
        {
            sb.append( terminal.charValue() );
        }
    }

    public
    static
    CharSequence
    asString( ProductionMatch< Character > pm )
    {
        StringBuildVisitor sbv = new StringBuildVisitor();
        visit( pm, sbv );

        return sbv.sb;
    }

    private
    final
    static
    class ParseResult< N, T >
    {
        // after a parse, exactly one of the below will be non-null
        private String msg;
        private DerivationMatch< N, T > match;
    }

    private
    static
    void
    assertParseMatch( RecursiveDescentParser< ?, ? > p,
                      CharSequence str,
                      int strLen,
                      ParseResult< ?, ? > pr )
    {
        int consumed = p.getConsumedTerminals();

        if ( consumed < strLen )
        {
            pr.msg =
                "Trailing characters (match stopped at index " + consumed + 
                "): " + str;
        }
    }

    private
    static
    void
    assertParseString( RecursiveDescentParser< ?, ? > p,
                       CharSequence str,
                       ParseResult< ?, ? > pr )
    {
        switch ( p.getMatcherState() )
        {
            case MATCHED: assertParseMatch( p, str, str.length(), pr ); break;

            case UNMATCHED: pr.msg = "Invalid syntax: " + str; break;

            case MATCHING: 
                throw state.createFail( 
                    "p is in state MATCHING after complete()" );
        }
    }

    // does input checking on behalf of public method facades
    private
    static
    < N >
    ParseResult< N, Character >
    getStringParseResult( RecursiveDescentParserFactory< N, Character > pf,
                          N goal,
                          CharSequence str )
    {
        inputs.notNull( pf, "pf" );
        inputs.notNull( goal, "goal" );
        inputs.notNull( str, "str" );

        RecursiveDescentParser< N, Character > p = pf.createParser( goal );

        for ( int i = 0, e = str.length(); i < e && p.isMatching(); ++i )
        {
            p.consumeTerminal( Character.valueOf( str.charAt( i ) ) );
        }

        if ( p.isMatching() ) p.complete();

        ParseResult< N, Character > res = new ParseResult< N, Character >();

        assertParseString( p, str, res );
        if ( res.msg == null ) res.match = p.getMatch();

        return res;
    }

    public
    static
    < N >
    DerivationMatch< N, Character >
    parseStringMatch( RecursiveDescentParserFactory< N, Character > pf,
                      N goal,
                      CharSequence str )
        throws SyntaxException
    {
        ParseResult< N, Character > pr = getStringParseResult( pf, goal, str );

        if ( pr.match == null ) throw new SyntaxException( pr.msg );
        else return pr.match;
    }
 
    public
    static
    < N >
    DerivationMatch< N, Character >
    createStringMatch( RecursiveDescentParserFactory< N, Character > pf,
                       N goal,
                       CharSequence str )
    {
        ParseResult< N, Character > pr = getStringParseResult( pf, goal, str );

        if ( pr.match == null ) throw new IllegalArgumentException( pr.msg );
        else return pr.match;
    }
}
