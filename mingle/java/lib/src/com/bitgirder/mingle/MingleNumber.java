package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.PatternHelper;

import java.util.regex.Pattern;

public
abstract
class MingleNumber
extends Number
implements MingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static Pattern INT_STR = Pattern.compile( "^-?\\d+" );

    static
    NumberFormatException
    asNumberFormatException( NumberFormatException nfe,
                             CharSequence input )
    {
        if ( INT_STR.matcher( input ).matches() )
        {
            String msg = "For input string: \"" + input + "\"";
            if ( nfe.getMessage().equals( msg ) )
            {
                String newMsg = "value out of range: " + input;
                nfe = new NumberFormatException( newMsg );
            }
        }

        return nfe;
    }
}
