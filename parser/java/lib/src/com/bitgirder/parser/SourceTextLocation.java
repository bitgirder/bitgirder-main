package com.bitgirder.parser;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class SourceTextLocation
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final CharSequence fileName;
    private final int line;
    private final int col;

    private
    SourceTextLocation( CharSequence fileName,
                        int line,
                        int col )
    {
        this.fileName = fileName;
        this.line = line;
        this.col = col;
    }

    public CharSequence getFileName() { return fileName; }
    public int getLine() { return line; }
    public int getColumn() { return col; }

    @Override
    public
    String
    toString()
    {
        return 
            new StringBuilder().
                append( "[ " ).
                append( fileName ).
                append( "; line " ).
                append( line ).
                append( ", col " ).
                append( col ).
                append( " ]" ).
                toString();
    }

    // line and col may be 0, indicating for the former that the location occurs
    // before the first line (aka, input is empty) and for the latter that the
    // location occurs before the first character of the line (aka, line is
    // empty)
    public
    static
    SourceTextLocation
    create( CharSequence fileName,
            int line,
            int col )
    {
        return
            new SourceTextLocation(
                inputs.notNull( fileName, "fileName" ),
                inputs.nonnegativeI( line, "line" ),
                inputs.nonnegativeI( col, "col" ) );
    }
}
