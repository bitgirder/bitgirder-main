package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.MingleIdentifier;
import com.bitgirder.mingle.MingleValue;
import com.bitgirder.mingle.QualifiedTypeName;
import com.bitgirder.mingle.ListTypeReference;

public
final
class EventSend
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final MingleReactor rct;
    private final MingleReactorEvent ev = new MingleReactorEvent();

    private
    EventSend( MingleReactor rct )
    {
        this.rct = rct;
    }

    private void send() throws Exception { rct.processEvent( ev ); }

    public
    void
    value( MingleValue mv )
        throws Exception
    {
        ev.setValue( inputs.notNull( mv, "mv" ) );
        send();
    }

    public
    void
    startList( ListTypeReference lt )
        throws Exception
    {
        ev.setStartList( inputs.notNull( lt, "lt" ) );
        send();
    }

    public
    void
    startMap()
        throws Exception
    {
        ev.setStartMap();
        send();
    }

    public
    void
    startField( MingleIdentifier fld )
        throws Exception
    {
        ev.setStartField( inputs.notNull( fld, "fld" ) );
        send();
    }

    public
    void
    startStruct( QualifiedTypeName qn )
        throws Exception
    {
        ev.setStartStruct( qn );
        send();
    }

    public
    void
    end()
        throws Exception
    {
        ev.setEnd();
        send();
    }

    public
    static
    EventSend
    forReactor( MingleReactor rct )
    {
        return new EventSend( inputs.notNull( rct, "rct" ) );
    }
}
