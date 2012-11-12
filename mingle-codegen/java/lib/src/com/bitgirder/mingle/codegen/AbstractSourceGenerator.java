package com.bitgirder.mingle.codegen;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Lang;

import java.util.Deque;

public
abstract
class AbstractSourceGenerator
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final StringBuilder sb = new StringBuilder();

    private final Deque< Integer > indents = Lang.newDeque();
    private int indent = 0;
    private int pos = -1;

    protected final CharSequence getString() { return sb; }

    protected
    final
    CharSequence
    joinSp( Iterable< ? > toks )
    {
        return Strings.join( " ", toks );
    }

    private boolean atLineStart() { return pos == -1; }

    private void clearLine() { pos = -1; }

    private
    void
    writeIndent()
    {
        state.equalInt( -1, pos );
        for ( pos = 0; pos < indent; ++pos ) sb.append( ' ' );
    }

    protected
    final
    void
    pushIndent( int len )
    {
        inputs.positiveI( len, "len" );

        indents.push( len );
        indent += len;
    }

    // Hardcoding default now; could make it paramterizable later
    protected final void pushIndent() { pushIndent( 4 ); }

    protected final void popIndent() { indent -= indents.pop(); }

    protected
    final
    void
    setIndent()
    {
        if ( pos > indent ) pushIndent( pos - indent );
    }

    protected
    void
    write( Object... parts )
    {
        inputs.noneNull( parts, "parts" );

        if ( atLineStart() ) writeIndent();

        int start = sb.length();

        for ( Object part : parts ) sb.append( part );
        pos += sb.length() - start;
    }

    protected
    final
    void
    writeLine( Object... parts )
    {
        write( parts );

        sb.append( '\n' );
        clearLine();
    }
}
