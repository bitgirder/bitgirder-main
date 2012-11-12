package com.bitgirder.mingle.http;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleIdentifierFormat;

import java.util.Map;

public
final
class MingleHttpProcessing
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static Map< MingleIdentifierFormat, CharSequence > 
        ID_STYLE_HDR_VALS;

    private MingleHttpProcessing() {}

    public
    static
    CharSequence
    asIdStyleHeaderValue( MingleIdentifierFormat fmt )
    {
        inputs.notNull( fmt, "fmt" );
        return state.get( ID_STYLE_HDR_VALS, fmt, "ID_STYLE_HDR_VALS" );
    }

    static
    {
        Map< MingleIdentifierFormat, CharSequence > m = Lang.newMap();

        for ( MingleIdentifierFormat fmt :
                MingleIdentifierFormat.class.getEnumConstants() )
        {
            m.put( 
                fmt, 
                MingleIdentifier.create( fmt.name().toLowerCase() ).
                    getExternalForm().
                    toString()
            );
        }

        ID_STYLE_HDR_VALS = Lang.unmodifiableMap( m );
    }
}
