package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.io.Charsets;

import java.util.Map;
import java.util.Set;
import java.util.HashSet;
import java.util.List;
import java.util.Arrays;

import java.nio.ByteBuffer;

import java.nio.charset.CharacterCodingException;

public
final
class Grammar< N, T >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final Map< N, ProductionMatcher > productions;
    private final Class< ? extends N > headType;

    private 
    Grammar( AbstractBuilder< N, T, ? > b )
    {
        this.productions = Lang.unmodifiableCopy( b.productions );
        this.headType = inputs.notNull( b.headType, "headType" );
    }

    public
    ProductionMatcher
    getProduction( N head )
    {
        inputs.notNull( head, "head" );
        return state.get( productions, head, "productions" );
    }

    public Iterable< N > getNonTerminals() { return productions.keySet(); }

    Class< ? extends N > headType() { return headType; }

    public
    static
    < N, T >
    Builder< N, T >
    createBuilder( Class< ? extends N > headType )
    {
        return new Builder< N, T >().setHeadType( headType );
    }

    public
    static
    < T >
    LexBuilder< T >
    createLexBuilder( Class< ? extends T > headType )
    {
        return new LexBuilder< T >().setHeadType( headType );
    }

    public
    static
    < N >
    BinaryBuilder< N >
    createBinaryBuilder( Class< ? extends N > headType )
    {
        return new BinaryBuilder< N >().setHeadType( headType );
    }

    public
    abstract
    static
    class AbstractBuilder< N, T, B extends AbstractBuilder< N, T, ? > >
    {
        private final Map< N, ProductionMatcher > productions = Lang.newMap();

        private Class< ? extends N > headType;

        private final Set< N > referencedNonTerminals = Lang.newSet();

        private AbstractBuilder() {}

        B
        castThis()
        {
            @SuppressWarnings( "unchecked" )
            B res = (B) this;

            return res;
        }

        final
        B
        setHeadType( Class< ? extends N > headType )
        {
            this.headType = inputs.notNull( headType, "headType" );
            return castThis();
        }

        public
        final
        ProductionMatcher
        unary( ProductionMatcher body )
        {
            inputs.notNull( body, "body" );
            return QuantifierMatcher.unary( body );
        }

        public
        final
        ProductionMatcher
        objectInstance( T obj )
        {
            inputs.notNull( obj, "obj" );
            return new ObjectInstanceMatcher< T >( obj );
        }

        public
        final
        ProductionMatcher
        atLeastOne( ProductionMatcher body )
        {
            inputs.notNull( body, "body" );
            return QuantifierMatcher.atLeastOne( body );
        }

        public
        final
        ProductionMatcher
        kleene( ProductionMatcher body )
        {
            inputs.notNull( body, "body" );
            return QuantifierMatcher.kleene( body );
        }

        public
        final
        ProductionMatcher
        exactly( ProductionMatcher body,
                 int count )
        {
            inputs.notNull( body, "body" );
            inputs.positiveI( count, "count" );

            return QuantifierMatcher.exactly( body, count );
        }

        // A list of at least one match for eltMatcher, separated by exactly one
        // match of joinerMatcher. If needed, we could add another method to
        // allow adjacent joinerMatches for empty/missing list elements.
        public
        final
        ProductionMatcher
        nonEmptyList( ProductionMatcher eltMatcher,
                      ProductionMatcher joinerMatcher )
        {
            inputs.notNull( eltMatcher, "eltMatcher" );
            inputs.notNull( joinerMatcher, "joinerMatcher" );

            return
                sequence(
                    eltMatcher, 
                    kleene( sequence( joinerMatcher, eltMatcher ) ) );
        }

        // A possibly unmatched (empty) list of joined elements
        public
        final
        ProductionMatcher
        list( ProductionMatcher eltMatcher,
              ProductionMatcher joinerMatcher )
        {
            return unary( nonEmptyList( eltMatcher, joinerMatcher ) );
        }

        public
        final
        ProductionMatcher
        list( N eltHead,
              N joinerHead )
        {
            inputs.notNull( eltHead, "eltHead" );
            inputs.notNull( joinerHead, "joinerHead" );

            return list( derivation( eltHead ), derivation( joinerHead ) );
        }

        public
        final
        ProductionMatcher
        terminalInstance( Class< ? extends T > terminalCls )
        {
            inputs.notNull( terminalCls, "terminalCls" );
            return TerminalInstanceMatcher.forClass( terminalCls );
        }

        public
        final
        ProductionMatcher
        derivation( N head )
        {
            inputs.notNull( head, "head" );

            referencedNonTerminals.add( head );
            return DerivationMatcher.forHead( head );
        }

        public
        final
        ProductionMatcher
        sequence( List< ProductionMatcher > exprs )
        {
            exprs = Lang.unmodifiableCopy( exprs, "exprs" );
            return SequenceMatcher.forSequence( exprs );
        }

        public
        final
        ProductionMatcher
        sequence( ProductionMatcher... exprs )
        {
            return 
                sequence( Arrays.asList( inputs.notNull( exprs, "exprs" ) ) );
        }

        private
        final
        ProductionMatcher
        union( UnionMatcher.Mode mode,
               ProductionMatcher... alts )
        {
            List< ProductionMatcher > l = 
                Lang.unmodifiableCopy( 
                    Arrays.asList( inputs.notNull( alts, "alts" ) ),
                    "alts",
                    false );
 
            return UnionMatcher.forMatchers( l, mode );
        }

        public
        final
        ProductionMatcher
        union( ProductionMatcher... alts )
        {
            return union( UnionMatcher.Mode.FIRST_WINS, alts );
        }

        public
        final
        B
        addProduction( N head,
                       ProductionMatcher body )
        {
            inputs.notNull( head, "head" );
            inputs.notNull( body, "body" );

            Lang.putUnique( productions, head, body );
 
            return castThis();
        }

        public
        final 
        Grammar< N, T > 
        build() 
        { 
            Set< N > unreferenced = Lang.copyOf( referencedNonTerminals );
            unreferenced.removeAll( productions.keySet() );

            if ( ! unreferenced.isEmpty() )
            {
                throw new UnrecognizedNonTerminalException(
                    "One or more productions reference undefined " +
                    "non-terminal(s): " + Strings.join( ", ", unreferenced ) );
            }

            return new Grammar< N, T >( this ); 
        }
    }

    public
    final
    static
    class Builder< N, T >
    extends AbstractBuilder< N, T, Builder< N, T > >
    {}

    public
    final
    static
    class LexBuilder< N >
    extends AbstractBuilder< N, Character, LexBuilder< N > >
    {
        private LexBuilder() {}

        public
        final
        ProductionMatcher
        charRange( char begIncl,
                   char endIncl )
        {
            return CharRangeMatcher.forRange( begIncl, endIncl );
        }

        public
        final
        ProductionMatcher
        ch( char ch )
        {
            return CharRangeMatcher.forRange( ch, ch );
        }

        // could generalize this eventually to all grammars as something like
        // terminalSet

        private
        final
        ProductionMatcher
        charSet( char[] set,
                 boolean isComplement )
        {
            inputs.notNull( set, "set" );
            inputs.isFalse( set.length == 0, "Set is empty" );

            Set< Character > copy = new HashSet< Character >( set.length );
            for ( char ch : set ) copy.add( Character.valueOf( ch ) );

            return TerminalSetMatcher.forSet( copy, isComplement );
        }

        public
        final
        ProductionMatcher
        charSet( char... set )
        { 
            return charSet( set, false );
        }

        public
        final
        ProductionMatcher
        charSetComplement( char... set )
        {
            return charSet( set, true );
        }

        public
        final
        ProductionMatcher
        string( CharSequence str )
        {
            inputs.notNull( str, "str" );

            List< ProductionMatcher > seq = Lang.newList( str.length() );

            for ( int i = 0, e = str.length(); i < e; ++i )
            {
                seq.add( ch( str.charAt( i ) ) );
            }

            return sequence( seq );
        }

        // Could move this to the base class if needed, buf for now this type of
        // union really only makes sense in a lexical grammar
        public
        final
        ProductionMatcher
        unionLongest( ProductionMatcher... alts )
        {
            return super.union( UnionMatcher.Mode.LONGEST_WINS, alts );
        }
    }

    public
    final
    static
    class BinaryBuilder< N >
    extends AbstractBuilder< N, Byte, BinaryBuilder< N > >
    {
        private BinaryBuilder() {}

        public
        ProductionMatcher
        octet( int i )
        {
            return OctetRangeMatcher.forOctet( i );
        }

        public
        ProductionMatcher
        octetRange( int minIncl,
                    int maxIncl )
        {
            return OctetRangeMatcher.forRange( minIncl, maxIncl );
        }

        public
        ProductionMatcher
        asciiString( CharSequence str )
        {
            inputs.notNull( str, "str" );

            try
            {
                ByteBuffer bb = Charsets.US_ASCII.asByteBuffer( str );
    
                List< ProductionMatcher > matchers = 
                    Lang.newList( bb.remaining() );
    
                while ( bb.hasRemaining() )
                {
                    matchers.add( OctetRangeMatcher.forOctet( bb.get() ) );
                }
    
                return sequence( matchers );
            }
            catch ( CharacterCodingException cce )
            {
                throw new RuntimeException( "Invalid ascii string: " + str );
            }
        }
    }
}
