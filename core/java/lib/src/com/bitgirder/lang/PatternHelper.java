package com.bitgirder.lang;

import com.bitgirder.validation.Inputs;

import java.util.regex.Pattern;
import java.util.regex.PatternSyntaxException;

public
final
class PatternHelper
{
    private static Inputs inputs = new Inputs();

    public 
    static 
    Pattern 
    compile( CharSequence pat ) 
    { 
        inputs.notNull( pat, "pat" );

        try { return Pattern.compile( pat.toString() ); }
        catch ( PatternSyntaxException pse ) 
        {
            throw new RuntimeException( pse );
        }
    }

    public
    static
    Pattern
    compile( CharSequence pat,
             int flags )
    {
        inputs.notNull( pat, "pat" );
 
        try { return Pattern.compile( pat.toString(), flags ); }
        catch ( PatternSyntaxException pse ) 
        {
            throw new RuntimeException( pse );
        }
    }

    public
    static
    Pattern
    compileCaseInsensitive( String pat )
    {
        return compile( pat, Pattern.CASE_INSENSITIVE );
    }

    public
    static
    CharSequence
    getSingleLineMessage( PatternSyntaxException pse )
    {
        inputs.notNull( pse, "pse" );

        StringBuilder msg = new StringBuilder();
        msg.append( pse.getDescription() );

        int indx = pse.getIndex();

        if ( indx >= 0 )
        {
            msg.append( " (near index " ).append( indx ).append( ")" );
        }

        return msg;
    }
}
