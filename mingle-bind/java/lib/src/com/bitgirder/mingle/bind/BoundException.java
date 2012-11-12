package com.bitgirder.mingle.bind;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleIdentifier;

public
abstract
class BoundException
extends Exception
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private boolean wasSerialized;

    protected
    BoundException( String msg,
                    Throwable cause )
    {
        super( msg, cause );
    }

    protected BoundException( Throwable cause ) { this( null, cause ); }

    final boolean wasSerialized() { return wasSerialized; }

    public
    static
    abstract
    class AbstractBuilder< B extends AbstractBuilder< B > >
    extends BoundStructureBuilder< B >
    {
        protected Throwable _cause;

        protected AbstractBuilder() {}
        
        public
        final
        B
        setCause( Throwable cause )
        {
            this._cause = cause;
            return castThis();
        }
    }

    protected
    static
    abstract
    class AbstractBindImplementation
    extends com.bitgirder.mingle.bind.AbstractBindImplementation
    {
        protected AbstractBindImplementation() {};

        @Override
        final
        void
        completeJavaValue( Object jVal,
                           AtomicTypeReference typ,
                           MingleValue mv,
                           MingleBinder mb,
                           ObjectPath< MingleIdentifier > path )
        {
            ( (BoundException) jVal ).wasSerialized = true;
        }
    }
}
