package com.bitgirder.mingle.codegen.java;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

final
class JvParam
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    JvId id;
    JvType type;

    void
    validate()
    {
        state.notNull( id, "id" );
        state.notNull( type, "type" );
    }

    static
    JvParam
    create( JvId id,
            JvType type )
    {
        JvParam res = new JvParam();
        res.id = id;
        res.type = type;

        return res;
    }

    static JvParam forField( JvField f ) { return create( f.name, f.type ); }
}
