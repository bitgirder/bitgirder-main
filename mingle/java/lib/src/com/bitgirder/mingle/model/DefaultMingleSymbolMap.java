package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.Map;
import java.util.Set;

final
class DefaultMingleSymbolMap
implements MingleSymbolMap
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Map< MingleIdentifier, MingleValue > fields;

    DefaultMingleSymbolMap( MingleSymbolMapBuilder< ? > b )
    {
        this.fields = Lang.unmodifiableCopy( b.fields );
    }

    public Set< MingleIdentifier > getKeySet() { return fields.keySet(); }
    public Iterable< MingleIdentifier > getFields() { return getKeySet(); }

    public
    boolean
    hasField( MingleIdentifier fld )
    {
        return fields.containsKey( inputs.notNull( fld, "fld" ) );
    }

    public
    MingleValue
    get( MingleIdentifier fld )
    {
        return fields.get( inputs.notNull( fld, "fld" ) );
    }
}
