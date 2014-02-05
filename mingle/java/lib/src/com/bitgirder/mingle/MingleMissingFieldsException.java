package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;

import com.bitgirder.lang.path.ObjectPath;

import java.util.Collection;
import java.util.Arrays;

public
final
class MingleMissingFieldsException
extends MingleValueException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    public
    MingleMissingFieldsException( Collection< MingleIdentifier > fields,
                                  ObjectPath< MingleIdentifier > location )
    {
        super( makeMessage( checkFields( fields ) ), location );
    }

    private
    static
    Collection< MingleIdentifier >
    checkFields( Collection< MingleIdentifier > fields )
    {
        inputs.noneNull( fields, "fields" );
        inputs.isFalse( fields.isEmpty(), "empty fields" );

        return fields;
    }

    private
    static
    String
    makeMessage( Collection< MingleIdentifier > fields )
    {
        MingleIdentifier[] arr =
            fields.toArray( new MingleIdentifier[ fields.size() ] );

        Arrays.sort( arr );
        
        return String.format( "missing field(s): %s",
            Strings.join( ", ", (Object[]) arr ) );
    }
}
