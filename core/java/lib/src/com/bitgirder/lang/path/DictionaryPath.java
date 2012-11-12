package com.bitgirder.lang.path;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class DictionaryPath< E >
extends ObjectPath< E >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final E key;

    private
    DictionaryPath( ObjectPath< E > parent,
                    E key )
    {
        super( parent, "parent" );
        this.key = inputs.notNull( key, "key" );
    }

    E getKey() { return key; }

    public
    static
    < E >
    DictionaryPath< E >
    create( ObjectPath< E > parent,
            E key )
    {
        return new DictionaryPath< E >( parent, key );
    }
}
