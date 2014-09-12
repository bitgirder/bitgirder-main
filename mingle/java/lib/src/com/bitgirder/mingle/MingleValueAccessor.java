package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import static com.bitgirder.log.CodeLoggers.Statics.*;

import com.bitgirder.lang.path.ObjectPath;

abstract
class MingleValueAccessor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    MingleValueAccessor() {}

    private
    String
    errorTypeNameForClass( Class< ? extends MingleValue > cls )
    {
        if ( cls.equals( MingleStruct.class ) ) return "struct";

        return cls.getSimpleName();
    }

    private
    MingleValue
    accessValueByClass( MingleValue val,
                        ObjectPath< MingleIdentifier > path,
                        Class< ? extends MingleValue > valCls )
    {
        if ( valCls.isInstance( val ) ) return val;

        String expctTyp = errorTypeNameForClass( valCls );
        CharSequence valTyp = Mingle.typeOf( val ).getExternalForm();

        throw new MingleValueCastException(
            "expected " + expctTyp + " but got " + valTyp, path );
    }

    // we take advantage here of the fact that all accessor cast targets are
    // atomic types or ar TYPE_OPAQUE_LIST
    private
    MingleValue
    castValue( MingleValue val,
               MingleTypeReference typ,
               ObjectPath< MingleIdentifier > path )
    {
        if ( typ.equals( Mingle.TYPE_OPAQUE_LIST ) ) {
            if ( val instanceof MingleList ) return val;
            throw Mingle.failCastType( typ, val, path );
        }

        return Mingle.castAtomic( val, (AtomicTypeReference) typ, typ, path );
    }

    final
    MingleValue
    accessValue( MingleValue val,
                 ObjectPath< MingleIdentifier > path,
                 MingleTypeReference typ,
                 Class< ? extends MingleValue > valCls )
    {
        if ( val == null || ( val instanceof MingleNull ) ) {
            return MingleNull.getInstance();
        }

        if ( typ == null ) return accessValueByClass( val, path, valCls );

        return castValue( val, typ, path );
    }
}
