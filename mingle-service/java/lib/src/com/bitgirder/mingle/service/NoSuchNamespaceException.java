package com.bitgirder.mingle.service;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.model.MingleNamespace;

public
final
class NoSuchNamespaceException
extends RuntimeException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleNamespace ns;

    public
    NoSuchNamespaceException( MingleNamespace ns )
    {
        super( inputs.notNull( ns, "ns" ).getExternalForm().toString() );
        this.ns = ns;
    }

    public MingleNamespace getNamespace() { return ns; }
}
