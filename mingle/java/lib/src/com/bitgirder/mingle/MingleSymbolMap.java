package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.Map;
import java.util.Set;

public
final
class MingleSymbolMap
implements MingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static MingleSymbolMap EMPTY = new Builder().buildMap();

    private final Map< MingleIdentifier, MingleValue > fields;

    MingleSymbolMap( Map< MingleIdentifier, MingleValue > fields )
    {
        this.fields = fields;
    }

    MingleSymbolMap( BuilderImpl< ?, ? > b )
    {
        this( Lang.unmodifiableCopy( b.fields ) );
    }

    public Set< MingleIdentifier > getKeySet() { return fields.keySet(); }
    public Iterable< MingleIdentifier > getFields() { return getKeySet(); }

    public int hashCode() { return fields.hashCode(); }

    public
    boolean
    equals( Object other )
    {
        if ( this == other ) return true;
        if ( ! ( other instanceof MingleSymbolMap ) ) return false;

        return fields.equals( ( (MingleSymbolMap) other ).fields );
    }

    public 
    Set< Map.Entry< MingleIdentifier, MingleValue > >
    entrySet()
    {
        return fields.entrySet();
    }

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

    static
    abstract
    class BuilderImpl< V extends MingleValue, B extends BuilderImpl >
    {
        final Map< MingleIdentifier, MingleValue > fields = Lang.newMap();
    
        BuilderImpl() {}

        final MingleSymbolMap buildMap() { return new MingleSymbolMap( this ); }

        private
        MingleIdentifier
        makeIdent( CharSequence symStr,
                   String paramName )
        {
            return 
                MingleIdentifier.create( inputs.notNull( symStr, paramName ) );
        }
    
        private B castThis() { return Lang.castUnchecked( this ); }
 
        public
        final
        B
        set( MingleIdentifier fld,
             MingleValue data )
        {
            inputs.notNull( fld, "fld" );
            inputs.notNull( data, "data" );
    
            if ( ! ( data instanceof MingleNull ) ) fields.put( fld, data );
    
            return castThis();
        }
    
        public
        final
        B
        set( CharSequence fld,
             MingleValue data )
        {
            return set( makeIdent( fld, "fld" ), data );
        }
    
        public
        final
        B
        setString( MingleIdentifier fld,
                   CharSequence str )
        {
            inputs.notNull( str, "str" );
    
            return set( fld, (MingleValue) new MingleString( str ) );
        }
    
        public
        final
        B
        setString( CharSequence fld,
                   CharSequence str )
        {
            return setString( makeIdent( fld, "fld" ), str );
        }
    
        public
        final
        B
        setBoolean( MingleIdentifier fld,
                    boolean b )
        {
            return set( fld, MingleBoolean.valueOf( b ) );
        }
    
        public
        final
        B
        setBoolean( CharSequence fld,
                    boolean b )
        {
            return setBoolean( makeIdent( fld, "fld" ), b );
        }
    
        public
        final
        B
        setInt64( MingleIdentifier fld,
                  long num )
        {
            return set( fld, new MingleInt64( num ) );
        }
    
        public
        final
        B
        setInt64( CharSequence fld,
                  long num )
        {
            return setInt64( makeIdent( fld, "fld" ), num );
        }
    
        public
        final
        B
        setInt32( MingleIdentifier fld,
                  int num )
        {
            return set( fld, new MingleInt32( num ) );
        }
    
        public
        final
        B
        setInt32( CharSequence fld,
                  int num )
        {
            return setInt32( makeIdent( fld, "fld" ), num );
        }
    
        public
        final
        B
        setFloat64( MingleIdentifier fld,
                    double num )
        {
            return set( fld, new MingleFloat64( num ) );
        }
    
        public
        final
        B
        setFloat64( CharSequence fld,
                    double num )
        {
            return setFloat64( makeIdent( fld, "fld" ), num );
        }
    
        public
        final
        B
        setFloat32( MingleIdentifier fld,
                    float num )
        {
            return set( fld, new MingleFloat32( num ) );
        }
    
        public
        final
        B
        setFloat32( CharSequence fld,
                    float num )
        {
            return setFloat32( makeIdent( fld, "fld" ), num );
        }
    
        public
        final
        B
        setBuffer( MingleIdentifier fld,
                   byte[] data )
        {
            inputs.notNull( data, "data" );
            return set( fld, new MingleBuffer( data ) );
        }
    
        public
        final
        B
        setBuffer( CharSequence fld,
                   byte[] data )
        {
            inputs.notNull( data, "data" );
            return set( fld, new MingleBuffer( data ) );
        }
    
        public
        final
        B
        setAll( MingleSymbolMap other )
        {
            inputs.notNull( other, "other" );
    
            for ( MingleIdentifier fld : other.getFields() ) 
            {
                set( fld, other.get( fld ) );
            }
    
            return castThis();
        }

        public
        abstract
        V
        build();
    }

    public
    final
    static
    class Builder
    extends BuilderImpl< MingleSymbolMap, Builder >
    {
        public MingleSymbolMap build() { return buildMap(); }
    }

    public static MingleSymbolMap empty() { return EMPTY; }
}
