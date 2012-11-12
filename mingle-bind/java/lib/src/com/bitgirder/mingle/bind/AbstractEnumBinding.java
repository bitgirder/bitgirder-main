package com.bitgirder.mingle.bind;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleEnum;
import com.bitgirder.mingle.model.MingleString;
import com.bitgirder.mingle.model.MingleValidationException;
import com.bitgirder.mingle.model.AtomicTypeReference;

import com.bitgirder.parser.SyntaxException;

public
abstract
class AbstractEnumBinding< E extends Enum< E > >
implements MingleBinding
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final Class< E > enCls;

    protected
    AbstractEnumBinding( Class< E > enCls )
    {
        this.enCls = state.notNull( enCls, "enCls" );
    }

    protected
    abstract
    MingleEnum
    getMingleEnum( E enVal );

    public
    final
    MingleValue
    asMingleValue( Object jvObj,
                   MingleBinder mb,
                   ObjectPath< String > path )
    {
        E enVal = enCls.cast( jvObj );
        MingleEnum res = getMingleEnum( enVal );

        if ( res == null )
        {
            throw MingleBindingException.createOutbound(
                "impl return null from getValueFor(), enVal: " + enVal, path );
        }
        else return res;
    }

    private
    MingleIdentifier
    parseId( MingleString str,
             ObjectPath< MingleIdentifier > path )
    {
        try { return MingleIdentifier.parse( str ); }
        catch ( SyntaxException se )
        {
            throw 
                new MingleValidationException(
                    "Invalid enum value: " + str, path );
        }
    }

    private
    MingleIdentifier
    extractEnumId( MingleValue mv,
                   ObjectPath< MingleIdentifier > path )
    {
        if ( mv instanceof MingleString )
        {
            return parseId( (MingleString) mv, path );
        }
        else if ( mv instanceof MingleEnum )
        {
            return ( (MingleEnum) mv ).getValue();
        }
        else
        {
            throw 
                new MingleValidationException( 
                    "unrecognized enum value", path );
        }
    }

    protected
    abstract
    E
    getJavaEnum( MingleIdentifier id );

    public
    final
    Object
    asJavaValue( AtomicTypeReference typ,
                 MingleValue mv,
                 MingleBinder mb,
                 ObjectPath< MingleIdentifier > path )
    {
        MingleIdentifier id = extractEnumId( mv, path );
        E res = getJavaEnum( id );

        if ( res == null )
        {
            throw 
                new MingleValidationException( "no such enum constant", path );
        }
        else return res;
    }
}
