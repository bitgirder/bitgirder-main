package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MingleStruct
implements MingleValue,
           TypedMingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final QualifiedTypeName type;
    private final MingleSymbolMap flds;

    MingleStruct( QualifiedTypeName type,
                  MingleSymbolMap flds )
    {
        this.type = type;
        this.flds = flds;
    }

    private
    MingleStruct( Builder b )
    {
        this( inputs.notNull( b.type, "type" ), b.buildMap() );
    }

    public QualifiedTypeName getType() { return type; }
    public MingleSymbolMap getFields() { return flds; }

    public 
    MingleTypeReference
    getTypeReference()
    {
        return new AtomicTypeReference( type, null );
    }

    public int hashCode() { return type.hashCode() | flds.hashCode(); }

    public
    boolean
    equals( Object o )
    {
        if ( o == this ) return true;
        if ( ! ( o instanceof MingleStruct ) ) return false;

        MingleStruct ms = (MingleStruct) o;
        return type.equals( ms.type ) && flds.equals( ms.flds );
    }

    public
    static
    MingleStruct
    create( QualifiedTypeName type,
            MingleSymbolMap flds )
    {
        return new MingleStruct(
            inputs.notNull( type, "type" ),
            inputs.notNull( flds, "flds" )
        );
    }

    public
    final
    static
    class Builder
    extends MingleSymbolMap.BuilderImpl< Builder >
    {
        private QualifiedTypeName type;

        public
        Builder
        setType( QualifiedTypeName type )
        {
            this.type = inputs.notNull( type, "type" );
            return this;
        }

        public
        Builder
        setType( CharSequence type )
        {
            inputs.notNull( type, "type" );
            return setType( QualifiedTypeName.create( type ) );
        }

        public MingleStruct build() { return new MingleStruct( this ); }
    }
}
