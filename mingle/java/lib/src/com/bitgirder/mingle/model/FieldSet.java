package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;
import java.util.Map;
import java.util.Iterator;

public
final
class FieldSet
implements Iterable< FieldDefinition >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static FieldSet EMPTY_FIELD_SET =
        create( Lang.< FieldDefinition >emptyList() );

    private final FieldDefinition[] flds;
    private final Map< MingleIdentifier, Integer > posById;

    private
    FieldSet( FieldDefinition[] flds,
              Map< MingleIdentifier, Integer > posById )
    {
        this.flds = flds;
        this.posById = Lang.unmodifiableMap( posById );
    }

    public
    List< FieldDefinition >
    getFields()
    {
        return Lang.unmodifiableList( Lang.asList( flds ) );
    }
    
    public 
    Iterator< FieldDefinition > 
    iterator()
    {
        return getFields().iterator();
    }

    public
    FieldDefinition
    getField( MingleIdentifier fld )
    {
        inputs.notNull( fld, "fld" );

        Integer pos = posById.get( fld );
        return pos == null ? null : flds[ pos.intValue() ];
    }

    public
    static
    FieldSet
    create( List< FieldDefinition > flds )
    {
        inputs.noneNull( flds, "flds" );

        FieldDefinition[] arr = new FieldDefinition[ flds.size() ];
        Map< MingleIdentifier, Integer > posById = Lang.newMap( arr.length );

        int i = 0;
        for ( Iterator< FieldDefinition > it = flds.iterator(); it.hasNext(); )
        {
            arr[ i ] = it.next();
            Lang.putUnique( posById, arr[ i ].getName(), i );

            ++i;
        }

        return new FieldSet( arr, posById );
    }

    public static FieldSet getEmptyFieldSet() { return EMPTY_FIELD_SET; }
}
