package com.bitgirder.jetty7;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import org.eclipse.jetty.server.Server;

public
final
class Jetty7ServerManager
extends Jetty7LifeCycleManager< Server >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private Jetty7ServerManager( Server server ) { super( server ); }

    public
    static
    Jetty7ServerManager
    manage( Server server )
    {
        inputs.notNull( server, "server" );
        return new Jetty7ServerManager( server );
    }
}
