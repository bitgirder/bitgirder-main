package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;

import java.util.List;

public
final
class MissingFieldsException
extends MingleValueException
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    MissingFieldsException( ObjectPath< MingleIdentifier > loc,
                            MingleIdentifier... flds )
    {
        super( makeMessage( inputs.noneNull( flds, "flds" ) ), loc );
    }

    private
    static
    String
    makeMessage( MingleIdentifier[] flds )
    {
        String tmpl = "missing field(s): %s";

        List< CharSequence > fldNms = Lang.newList( flds.length );
        for ( MingleIdentifier fld : flds ) fldNms.add( fld.getExternalForm() );
        return String.format( tmpl, Strings.join( ", ", fldNms ) );
    }
}
