package com.bitgirder.log;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.util.Map;

public
interface CodeEvent
{
    // not null
    public 
    CodeEventType 
    type();

    // not null; may be empty. Implementations may or may not make defensive
    // copies; event creators which for some reason may change the contents of a
    // message array should create events with a copy
    public 
    Object[] 
    message();

    // may be null
    public 
    Throwable 
    throwable();

    // Live map of arbitrary kv pairs associated with this event (network
    // connection info, classname, pid, etc); may be null so that impls can
    // lazily initialize, only adding if known needed
    public
    Map< Object, Object >
    attachments();

    // should be positive
    public 
    long 
    time();
}
