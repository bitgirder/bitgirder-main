package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MingleBoolean
implements MingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public final static MingleBoolean TRUE = new MingleBoolean( Boolean.TRUE );

    public final static MingleBoolean FALSE = 
        new MingleBoolean( Boolean.FALSE );

    private final Boolean b;

    private MingleBoolean( Boolean b ) { this.b = b; }

    public boolean booleanValue() { return b.booleanValue(); }

    public int hashCode() { return b.hashCode(); }

    public
    boolean
    equals( Object other )
    {
        return 
            other == this ||
            ( other instanceof MingleBoolean &&
              b.equals( ( (MingleBoolean) other ).b ) );
    }

    @Override public String toString() { return b.toString(); }

    public
    static
    MingleBoolean
    valueOf( boolean b )
    {
        return b ? TRUE : FALSE; 
    }

    public
    static
    MingleBoolean
    parse( CharSequence str )
        throws MingleSyntaxException
    {
        inputs.notNull( str, "str" );

        String s = str.toString();

        if ( s.equalsIgnoreCase( "true" ) ) return TRUE;
        else if ( s.equalsIgnoreCase( "false" ) ) return FALSE;
        else 
        {
            String msg = "Invalid boolean string: " + str;
            throw new MingleSyntaxException( msg, 1 );
        }
    }
}
