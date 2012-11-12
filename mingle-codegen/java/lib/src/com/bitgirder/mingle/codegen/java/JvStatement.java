package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class JvStatement
implements JvExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    final JvExpression e;

    JvStatement( JvExpression e ) { this.e = state.notNull( e, "e" ); }

    public void validate() {}

    // name and params are assumed to be set and finalized at this point
    static
    JvStatement
    superCallTo( JvMethod m )
    {
        JvFuncCall call = new JvFuncCall();
        call.target = new JvAccess( JvId.SUPER, m.name );

        for ( JvParam p : m.params ) call.params.add( p.id );

        return new JvStatement( call );
    }
}
