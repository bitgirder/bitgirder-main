package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.Mingle;
import com.bitgirder.mingle.MingleIdentifier;

import com.bitgirder.lang.Strings;

import com.bitgirder.lang.path.ObjectPath;

public
final
class EventExpectation
{
    public final MingleReactorEvent event;
    public final ObjectPath< MingleIdentifier > path;

    EventExpectation( MingleReactorEvent event,
                      ObjectPath< MingleIdentifier > path )
    {
        this.event = event;
        this.path = path;
    }

    @Override
    public
    String
    toString()
    {
        return Strings.inspect( this, true,
            "event", event.inspect(),
            "path", path == null ? null : Mingle.formatIdPath( path )
        ).
        toString();
    }
}
