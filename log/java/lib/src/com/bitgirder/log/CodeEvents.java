package com.bitgirder.log;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.Map;

public
final
class CodeEvents
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final static Object[] EMPTY_MSG = new Object[] {};

    private final static DefaultCodeEventFormatter DEFAULT_FORMATTER =
        new DefaultCodeEventFormatter();
    
    private CodeEvents() {}

    public
    static
    CodeEventType
    type( CodeEvent ev )
    {
        return inputs.notNull( inputs.notNull( ev, "ev" ).type(), "ev.type()" );
    }

    public
    static
    Object[]
    message( CodeEvent ev )
    {
        inputs.notNull( ev, "ev" );
        return inputs.notNull( ev.message(), "ev.message()" );
    }

    public
    static
    Map< Object, Object >
    attachments( CodeEvent ev )
    {
        inputs.notNull( ev, "ev" );
        return inputs.notNull( ev.attachments(), "ev.attachments()" );
    }

    public
    static
    long
    time( CodeEvent ev )
    {
        inputs.notNull( ev, "ev" );
        return inputs.positiveL( ev.time(), "ev.time()" );
    }

    private
    static
    CharSequence
    doFormat( CodeEvent ev,
              CodeEventFormatter fmtr )
    {
        inputs.notNull( ev, "ev" );

        StringBuilder res = new StringBuilder();
        fmtr.appendFormat( res, ev );

        return res;
    }

    public
    static
    CharSequence
    format( CodeEvent ev,
            CodeEventFormatter fmtr )
    {
        return doFormat( ev, inputs.notNull( fmtr, "fmtr" ) );
    }

    public
    static
    CharSequence
    format( CodeEvent ev )
    {
        return doFormat( ev, DEFAULT_FORMATTER );
    }

    // msg may be null; returned event will have a non-null but empty message;
    // similarly for attachements (null --> empty read-only map)
    public
    static
    CodeEvent
    create( final CodeEventType type,
            final Object[] msg,
            final Throwable th,
            final Map< Object, Object > attachments,
            final long time )
    {
        inputs.notNull( type, "type" );
        inputs.positiveL( time, "time" );

        final Map< Object, Object > atts = 
            attachments == null ? Lang.newMap() : attachments;
        
        return new CodeEvent() {
            public CodeEventType type() { return type; }
            public Object[] message() { return msg == null ? EMPTY_MSG : msg; }
            public Throwable throwable() { return th; }
            public Map< Object, Object > attachments() { return atts; }
            public long time() { return time; }
        };
    }
}
