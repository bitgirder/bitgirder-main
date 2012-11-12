package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.Map;

public
final
class MingleServiceRequest
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleNamespace ns;
    private final MingleIdentifier svc;
    private final MingleIdentifier op;
    private final MingleSymbolMap params;
    private final MingleValue authentication;

    private
    MingleServiceRequest( Builder b )
    {
        this.ns = inputs.notNull( b.ns, "ns" );
        this.svc = inputs.notNull( b.svc, "svc" );
        this.op = inputs.notNull( b.op, "op" );
        this.params = b.paramsBld.build();
        this.authentication = b.authentication;
    }

    public MingleNamespace getNamespace() { return ns; }
    public MingleIdentifier getService() { return svc; }
    public MingleIdentifier getOperation() { return op; }
    public MingleSymbolMap getParameters() { return params; }
    public MingleValue getAuthentication() { return authentication; }

    public
    final
    static
    class Builder
    {
        private MingleNamespace ns;
        private MingleIdentifier svc;
        private MingleIdentifier op;
        private MingleValue authentication;

        private final MingleSymbolMapBuilder< Builder > paramsBld =
            MingleSymbolMapBuilder.create( this );

        public
        Builder
        setNamespace( MingleNamespace ns )
        {
            this.ns = inputs.notNull( ns, "ns" );
            return this;
        }

        public
        Builder
        setNamespace( CharSequence ns )
        {
            return
                setNamespace( 
                    MingleNamespace.create( inputs.notNull( ns, "ns" ) ) );
        }

        public
        Builder
        setService( MingleIdentifier svc )
        {
            this.svc = inputs.notNull( svc, "svc" );
            return this;
        }

        public
        Builder
        setService( CharSequence ns )
        {
            return
                setService( 
                    MingleIdentifier.create( inputs.notNull( ns, "ns" ) ) );
        }

        public
        Builder
        setOperation( MingleIdentifier op )
        {
            this.op = inputs.notNull( op, "op" );
            return this;
        }

        public
        Builder
        setOperation( CharSequence ns )
        {
            return
                setOperation( 
                    MingleIdentifier.create( inputs.notNull( ns, "ns" ) ) );
        }

        public MingleSymbolMapBuilder< Builder > params() { return paramsBld; }
        public MingleSymbolMapBuilder< Builder > p() { return params(); }

        public
        Builder
        setAuthentication( MingleValue authentication )
        {
            this.authentication = 
                inputs.notNull( authentication, "authentication" );

            return this;
        }

        public
        MingleServiceRequest
        build()
        {
            return new MingleServiceRequest( this );
        }
    }
}
