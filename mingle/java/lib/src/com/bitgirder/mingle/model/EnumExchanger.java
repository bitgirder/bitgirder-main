package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;

import java.util.Map;

final
class EnumExchanger< E extends Enum< E > >
extends AbstractValueExchanger< E >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleEnum[] mgEnums;
    private final Map< MingleIdentifier, E > jVals;

    protected
    EnumExchanger( AtomicTypeReference typ,
                   Class< E > cls )
    {
        super( inputs.notNull( typ, "typ" ), inputs.notNull( cls, "cls" ) );

        E[] enVals = cls.getEnumConstants();
        mgEnums = new MingleEnum[ enVals.length ];
        jVals = Lang.newMap();

        for ( int i = 0, e = enVals.length; i < e; ++i )
        {
            mgEnums[ i ] = makeMgEnum( enVals[ i ], getMingleType() );
            jVals.put( mgEnums[ i ].getValue(), enVals[ i ] );
        }
    }

    protected
    final
    MingleValue
    implAsMingleValue( E en,
                       ObjectPath< String > path )
    {
        return mgEnums[ en.ordinal() ];
    }

    private
    E
    forIdentifier( MingleIdentifier id,
                   MingleString orig,
                   ObjectPath< MingleIdentifier > path )
    {
        E res = jVals.get( id );

        if ( res == null )
        {
            CharSequence errVal = orig == null ? id.getExternalForm() : orig;

            throw 
                MingleValidation.
                    createFail( path, "Invalid enum value:", errVal );
        }
        else return res;
    }

    private
    E
    fromMgEnum( MingleEnum me,
                ObjectPath< MingleIdentifier > path )
    {
        AtomicTypeReference expct = getMingleType();
        AtomicTypeReference actual = me.getType();

        if ( me.getType().equals( expct ) )
        {
            MingleIdentifier id = me.getValue();
            return forIdentifier( id, null, path );
        }
        else throw new MingleTypeCastException( expct, actual, path );
    }

    private
    E
    fromMgString( MingleString ms,
                  ObjectPath< MingleIdentifier > path )
    {
        return forIdentifier( createIdentifier( ms, path ), ms, path );
    }

    public
    final
    Object
    asJavaValue( MingleValue mv,
                 ObjectPath< MingleIdentifier > path )
    {
        if ( mv instanceof MingleEnum ) 
        {
            return fromMgEnum( (MingleEnum) mv, path );
        }
        else
        {
            MingleString ms = 
                (MingleString) MingleModels.
                    asMingleInstance( 
                        MingleModels.TYPE_REF_MINGLE_STRING, mv, path );
            
            return fromMgString( ms, path );
        }
    }

    private
    static
    MingleEnum
    makeMgEnum( Enum< ? > en,
                AtomicTypeReference typ )
    {
        return
            new MingleEnum.Builder().
                setType( typ ).
                setValue( MingleIdentifier.create( en.name().toLowerCase() ) ).
                build();
    }
}
