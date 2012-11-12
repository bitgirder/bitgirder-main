package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.parser.MingleParsers;

import java.util.Map;

import java.nio.ByteBuffer;

public
class MingleSymbolMapBuilder< B >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    // non-null when this is returning something other than itself -- see
    // MingleStructureBuilder
    private B bldrRes;

    final Map< MingleIdentifier, MingleValue > fields = Lang.newMap();

    MingleSymbolMapBuilder() {}

    private
    MingleIdentifier
    makeIdent( CharSequence symStr,
               String paramName )
    {
        return 
            MingleParsers.createIdentifier(
                inputs.notNull( symStr, paramName ) );
    }

    private
    B
    retVal()
    {
        if ( bldrRes == null ) return Lang.castUnchecked( this );
        else return bldrRes;
    }

    public
    B
    set( MingleIdentifier fld,
         MingleValue data )
    {
        inputs.notNull( fld, "fld" );
        inputs.notNull( data, "data" );

        if ( ! ( data instanceof MingleNull ) ) fields.put( fld, data );

        return retVal();
    }

    public
    B
    set( CharSequence fld,
         MingleValue data )
    {
        return set( makeIdent( fld, "fld" ), data );
    }

    public
    B
    setString( MingleIdentifier fld,
               CharSequence str )
    {
        inputs.notNull( str, "str" );

        return set( 
            fld, (MingleValue) MingleModels.asMingleString( str ) );
    }

    public
    B
    setString( CharSequence fld,
               CharSequence str )
    {
        return setString( makeIdent( fld, "fld" ), str );
    }

    public
    B
    setBoolean( MingleIdentifier fld,
                boolean b )
    {
        return set( fld, MingleModels.asMingleBoolean( b ) );
    }

    public
    B
    setBoolean( CharSequence fld,
                boolean b )
    {
        return setBoolean( makeIdent( fld, "fld" ), b );
    }

    public
    B
    setInt64( MingleIdentifier fld,
              long num )
    {
        return set( fld, MingleModels.asMingleInt64( num ) );
    }

    public
    B
    setInt64( CharSequence fld,
              long num )
    {
        return setInt64( makeIdent( fld, "fld" ), num );
    }

    public
    B
    setInt32( MingleIdentifier fld,
              int num )
    {
        return set( fld, MingleModels.asMingleInt32( num ) );
    }

    public
    B
    setInt32( CharSequence fld,
              int num )
    {
        return setInt32( makeIdent( fld, "fld" ), num );
    }

    public
    B
    setDouble( MingleIdentifier fld,
               double num )
    {
        return set( fld, MingleModels.asMingleDouble( num ) );
    }

    public
    B
    setDouble( CharSequence fld,
               double num )
    {
        return setDouble( makeIdent( fld, "fld" ), num );
    }

    public
    B
    setFloat( MingleIdentifier fld,
              float num )
    {
        return set( fld, MingleModels.asMingleFloat( num ) );
    }

    public
    B
    setFloat( CharSequence fld,
              float num )
    {
        return setFloat( makeIdent( fld, "fld" ), num );
    }

    public
    B
    setBuffer( MingleIdentifier fld,
               ByteBuffer data )
    {
        MingleBuffer mb = MingleModels.asMingleBuffer( data );
        return set( fld, mb );
    }

    public
    B
    setBuffer( CharSequence fld,
               ByteBuffer data )
    {
        return setBuffer( makeIdent( fld, "fld" ), data );
    }

    public
    B
    setBuffer( MingleIdentifier fld,
               byte[] data )
    {
        return 
            setBuffer( fld, ByteBuffer.wrap( inputs.notNull( data, "data" ) ) );
    }

    public
    B
    setBuffer( CharSequence fld,
               byte[] data )
    {
        return 
            setBuffer( fld, ByteBuffer.wrap( inputs.notNull( data, "data" ) ) );
    }

    public
    B
    setAll( MingleSymbolMap other )
    {
        inputs.notNull( other, "other" );

        for ( MingleIdentifier fld : other.getFields() ) 
        {
            set( fld, other.get( fld ) );
        }

        return retVal();
    }

    public 
    MingleSymbolMap 
    build() 
    { 
        return new DefaultMingleSymbolMap( this ); 
    }

    static
    < B >
    MingleSymbolMapBuilder< B >
    create( B bldrRes )
    {
        inputs.notNull( bldrRes, "bldrRes" );

        MingleSymbolMapBuilder< B > res = new MingleSymbolMapBuilder< B >();
        res.bldrRes = bldrRes;

        return res;
    }
}
