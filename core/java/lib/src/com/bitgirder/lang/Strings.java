package com.bitgirder.lang;

import com.bitgirder.validation.Inputs;

import java.util.List;
import java.util.Iterator;
import java.util.Arrays;
import java.util.ArrayList;
import java.util.Collection;

import java.text.DecimalFormat;

import java.nio.CharBuffer;

// Note: code should always use Strings.join varargs methods. At some point
// we're likely to find that hardcoding certain small common arg list lengths
// (2,3,4 strings) and unrolling the join loop will be a performance gain, but
// we'll add that as a top-level check to the varargs methods as needed. The
// alternative, adding methods like join2, join3, etc, would litter code and
// confuse things.
//
// Another likely optimization at some point may be to add variants of join
// which take lists of ? extends CharSequence and which return a CharSequence
// view on top of the underlying list. This will only be good in cases where
// the caller is sure that no other process will attempt to mutate an underlying
// joined string, but there are actually quite a lot of those cases around. 
public
class Strings
{
    private static Inputs inputs = new Inputs();

    public
    static
    CharSequence
    join( CharSequence delim,
          Iterable< ? > toks )
    {
        inputs.notNull( delim, "delim" );
        inputs.notNull( toks, "toks" );

        StringBuilder res = new StringBuilder();

        for ( Iterator< ? > it = toks.iterator(); it.hasNext(); )
        {
            res.append( String.valueOf( it.next() ) );
            if ( it.hasNext() ) { res.append( delim ); }
        }

        return res;
    }

    public
    static
    < T extends CharSequence >
    CharSequence
    join( CharSequence delim,
          T[] toks )
    {
        inputs.notNull( toks, "toks" );
        return join( delim, Arrays.asList( toks ) );
    }

    public
    static
    CharSequence
    join( CharSequence delim,
          Object... toks )
          
    {
        inputs.notNull( toks, "toks" );
        return join( delim, Arrays.asList( toks ) );
    }

    public
    static
    CharSequence
    crossJoin( CharSequence inner,
               CharSequence outer,
               List< ? > toks )
    {
        inputs.notNull( toks, "toks" );
        inputs.isTrue( toks.size() % 2 == 0, 
            "Need an even number of tokens (got " + toks.size() + ")" );

        inputs.notNull( inner, "inner" );
        inputs.notNull( outer, "outer" );

        ArrayList< StringBuilder > toks2 = 
            new ArrayList< StringBuilder >( toks.size() / 2 );

        for ( Iterator< ? > it = toks.iterator(); it.hasNext(); )
        {
            StringBuilder sb = new StringBuilder();
            sb.append( it.next() ).append( inner ).append( it.next() );
            toks2.add( sb );
        }

        return join( outer, toks2 );
    }

    public
    static
    CharSequence
    crossJoin( CharSequence inner,
               CharSequence outer,
               Object... toks )
    {
        return crossJoin( 
            inner, outer, Arrays.asList( inputs.notNull( toks, "toks" ) ) );
    }

    private
    static
    CharSequence
    getInspectionName( Class< ? > cls,
                       boolean shortName )
    {
        String name = cls.getName();

        if ( shortName )
        {
            // Will be either the name unchanged or everything after the last
            // '.', if there is a '.' in the name
            name = name.substring( name.lastIndexOf( '.' ) + 1, name.length() );
        }

        return name;
    }

    // Docs should mention performance considerations here, particularly that
    // this method is not ultra-optimized and is meant for non-performance
    // critical debugging and logging usage. The focus is on
    // speed of development, not speed of execution.
    //
    public
    static
    CharSequence
    inspect( Object o,
             boolean shortName,
             Object... elts )
    {
        inputs.notNull( o, "o" );
        inputs.notNull( elts, "elts" );

        StringBuilder res = 
            new StringBuilder( getInspectionName( o.getClass(), shortName ) );

        res.append( ":{ " );
        res.append( crossJoin( " = ", ", ", elts ) );
        res.append( " }" );

        return res;
    }

    public
    static
    CharSequence
    inspect( Object o,
             boolean shortName,
             List< ? > toks )
    {
        return inspect( o, shortName, toks.toArray() );
    }

    private
    static
    void
    appendIndent( StringBuilder sb,
                  CharSequence indentStr,
                  int level )
    {
        for ( int i = 0; i < level; ++i ) sb.append( indentStr );
    }

    private
    final
    static
    class StringInspector
    implements Inspector
    {
        private final StringBuilder sb;
        private final CharSequence indentStr;
        private final int indentLevel;

        private
        StringInspector( StringBuilder sb,
                         CharSequence indentStr,
                         int indentLevel )
        {
            this.sb = sb;
            this.indentStr = indentStr;
            this.indentLevel = indentLevel;
        }

        public
        Inspector
        add( CharSequence fieldName,
             Object fieldVal )
        {
            appendIndent( sb, indentStr, indentLevel );

            sb.append( fieldName ).
               append( ": " );
            
            inspect( fieldVal, sb, indentStr, indentLevel );

            sb.append( "\n" ); // This could be some other separator (';', ',')

            return this;
        }
    }

    private
    static
    void
    inspectInspectable( Inspectable i,
                        StringBuilder sb,
                        CharSequence indentStr,
                        int indentLevel )   
    {
        sb.append( getInspectionName( i.getClass(), true ) ).
           append( ":{\n" );

        StringInspector insp = 
            new StringInspector( sb, indentStr, indentLevel + 1 );
        i.accept( insp );

        appendIndent( sb, indentStr, indentLevel );
        sb.append( "}" );
    }

    private
    static
    void
    inspectCollection( Collection< ? > coll,
                       StringBuilder sb,
                       CharSequence indentStr,
                       int indentLevel )
    {
        int nextLevel = indentLevel + 1;

        switch ( coll.size() )
        {
            case 0: sb.append( "[]" ); break;

            case 1: 
                sb.append( "[ " );
                inspect( coll.iterator().next(), sb, indentStr, nextLevel );
                sb.append( " ]" );
                break;

            default:
                sb.append( "[\n" );

                for ( Object obj : coll )
                {
                    appendIndent( sb, indentStr, nextLevel );
                    inspect( obj, sb, indentStr, nextLevel );
                    sb.append( "\n" );
                }
        
                appendIndent( sb, indentStr, indentLevel );
                sb.append( "]" );
        }
    }

    private
    static
    void
    inspect( Object obj,
             StringBuilder sb,
             CharSequence indentStr,
             int indentLevel )
    {
        if ( obj instanceof Inspectable )
        {
            inspectInspectable( (Inspectable) obj, sb, indentStr, indentLevel );
        }
        else if ( obj instanceof Collection< ? > )
        {
            inspectCollection( 
                (Collection< ? >) obj, sb, indentStr, indentLevel );
        }
        else sb.append( String.valueOf( obj ) );
    }

    // this is a transition point from the old form of inspect, and hopefully
    // will supplant that version eventually; 
    public
    static
    CharSequence
    inspect( Object obj )
    {
        StringBuilder res = new StringBuilder();
        inspect( obj, res, "    ", 0 ); 

        return res;
    }

    public
    static
    String
    toString( double d,
              int fracDigits )
    {
        DecimalFormat f = new DecimalFormat();
        f.setMaximumFractionDigits( fracDigits );
        f.setMinimumFractionDigits( fracDigits );
        
        return f.format( d );
    }

    // Placeholder for now
    public
    static
    String
    valueOf( Object obj )
    {
        return String.valueOf( obj );
    }
}
