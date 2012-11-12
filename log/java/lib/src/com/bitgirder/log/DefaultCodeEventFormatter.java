package com.bitgirder.log;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.util.Map;

import java.text.SimpleDateFormat;

public
class DefaultCodeEventFormatter
implements CodeEventFormatter
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static String EXCPT_SEP = "\n-------\n";
 
    // This format will be approximately that of rfc3339, except that the
    // time zone will not have a ':' between hours and mins in the timezone
    // offset. We add that in manually in appendTime()
    private final static SimpleDateFormat RFC3339_APPROX_DATE_FMT =
        new SimpleDateFormat( "yyyy-MM-dd'T'HH:mm:ss.SSSZ" );

    public DefaultCodeEventFormatter() {}

    protected final void openBracket( StringBuilder sb ) { sb.append( "[ " ); }
    protected final void closeBracket( StringBuilder sb ) { sb.append( " ]" ); }

    protected
    void
    appendType( StringBuilder sb,
                CodeEvent ev )
    {
        openBracket( sb );
        sb.append( CodeEvents.type( ev ) );
        closeBracket( sb );
    }

    // Overridable
    protected
    void
    addKeyValSep( StringBuilder sb )
    {
        sb.append( " = " );
    }

    // Should return true if this method appended data; false otherwise. 'pref'
    // should always be added first if and only if the return value from this
    // method will be false.
    protected
    boolean
    appendAttachment( StringBuilder sb,
                      Object key,
                      Object val,
                      String pref )
    {
        sb.append( pref );
        sb.append( key );
        addKeyValSep( sb );
        sb.append( val );

        return true;
    }

    protected
    void
    appendAttachments( StringBuilder sb,
                       Map< Object, Object > atts )
    {
        if ( atts.isEmpty() ) sb.append( "[]" );
        else
        {
            boolean didAdd = false;
    
            for ( Map.Entry< Object, Object > e : atts.entrySet() )
            {
                String sep = didAdd ? ", " : "[ ";
                
                // order here matters -- we always want to eval
                // appendAttachment(), even if didAdd is already true
                didAdd = 
                    appendAttachment( sb, e.getKey(), e.getValue(), sep ) ||
                    didAdd;
            }

            if ( didAdd ) closeBracket( sb ); else sb.append( "[]" );
        }
    }

    protected
    void
    appendMessage( StringBuilder sb,
                   CodeEvent ev )
    {
        Object[] msg = CodeEvents.message( ev );

        int len = msg.length;

        if ( len > 0 ) 
        {
            sb.append( ": " );

            for ( int i = 0, l = len - 1; i < len; ++i )
            {
                sb.append( msg[ i ] );
                if ( i < l ) sb.append( ' ' );
            }
        }
    }

    private
    void
    appendStackTrace( StringBuilder sb,
                      StackTraceElement[] trace )
    {
        if ( trace == null ) sb.append( "<No Trace>" );
        else
        {
            for ( StackTraceElement elt : trace )
            {
                sb.append( '\n' ).append( elt );
            }
        }
    }

    protected
    void
    appendThrowable( StringBuilder sb,
                     Throwable th,
                     boolean isTop )
    {
        if ( th != null )
        {
            sb.append( EXCPT_SEP );
            sb.append( isTop ? "Throwable " : "Caused by " );

            sb.append( th.getClass().getName() ).
               append( ": " ).
               append( th.getLocalizedMessage() );
     
            appendStackTrace( sb, th.getStackTrace() );
    
            appendThrowable( sb, th.getCause(), false );
        }
    }

    protected
    void
    appendTime( StringBuilder sb,
                long time )
    {
        String ts = RFC3339_APPROX_DATE_FMT.format( time );
 
        int len = ts.length();

        openBracket( sb );
        sb.append( ts, 0, len - 2 );
        sb.append( ':' );
        sb.append( ts, len - 2, len );
        closeBracket( sb );
    }

    public
    final
    void
    appendFormat( StringBuilder sb,
                  CodeEvent ev )
    {
        appendTime( sb, CodeEvents.time( ev ) );
        appendAttachments( sb, CodeEvents.attachments( ev ) );
        appendType( sb, ev );
        appendMessage( sb, ev );
        appendThrowable( sb, ev.throwable(), true );
    }
}
