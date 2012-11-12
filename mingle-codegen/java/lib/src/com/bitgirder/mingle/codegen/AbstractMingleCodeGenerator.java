package com.bitgirder.mingle.codegen;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
abstract
class AbstractMingleCodeGenerator
implements MingleCodeGenerator
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private MingleCodeGeneratorContext ctx;

    protected final MingleCodeGeneratorContext context() { return ctx; }

    protected
    final
    void
    code( Throwable th,
          Object... msg )
    {
        ctx.log().code( th, msg );
    }

    protected final void code( Object... msg ) { ctx.log().code( msg ); }

    protected
    final
    void
    warn( Throwable th,
          Object... msg )
    {
        ctx.log().warn( th, msg );
    }

    protected final void warn( Object... msg ) { ctx.log().warn( msg ); }

    protected
    abstract
    void
    startGen()
        throws Exception;

    public
    final
    void
    start( MingleCodeGeneratorContext ctx )
        throws Exception
    {
        state.isTrue( this.ctx == null, "start() already called" );
        this.ctx = state.notNull( ctx, "ctx" );

        startGen();
    }
}
