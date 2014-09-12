package com.bitgirder.mingle.reactor;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.MingleIdentifier;

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
}
