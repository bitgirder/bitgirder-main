package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.PatternHelper;
import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;

import java.util.regex.Matcher;
import java.util.regex.Pattern;

public
final
class MingleRegexRestriction
extends MingleValueRestriction
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Pattern pat;

    private MingleRegexRestriction( Pattern pat ) { this.pat = pat; }

    private
    boolean
    checkPat( CharSequence str )
    {
        return pat.matcher( str ).matches();
    }

    boolean
    implValidate( MingleValue mv )
    {
        return checkPat( state.cast( MingleString.class, mv ) );
    }

    void
    appendExternalForm( StringBuilder sb )
    {
        Lang.appendRfc4627String( sb, pat.toString() );
    }

    public int hashCode() { return pat.toString().hashCode(); }

    public
    boolean
    equals( Object other )
    {
        if ( other == this ) return true;
        else if ( other instanceof MingleRegexRestriction )
        {
            MingleRegexRestriction o = (MingleRegexRestriction) other;
            return pat.toString().equals( o.pat.toString() );
        }
        else return false;
    }

    public 
    boolean 
    matches( CharSequence str )
    {
        return checkPat( inputs.notNull( str, "str" ) );
    }

    public
    static
    MingleRegexRestriction
    create( Pattern pat )
    {
        return new MingleRegexRestriction( inputs.notNull( pat, "pat" ) );
    }

    public
    static
    MingleRegexRestriction
    create( CharSequence pat )
    {
        inputs.notNull( pat, "pat" );
        return create( PatternHelper.compile( pat ) );
    }
}
