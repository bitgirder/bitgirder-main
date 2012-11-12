package com.bitgirder.mingle.bind;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.mingle.service.MingleServiceEndpoint;

public
final
class BoundServices
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private BoundServices() {}

    public
    static
    void
    addRoute( BoundService svc,
              MingleServiceEndpoint.Builder b )
    {
        inputs.notNull( svc, "svc" );
        inputs.notNull( b, "b" );

        b.addRoute( svc.getNamespace(), svc.getServiceId(), svc );
    }
}
