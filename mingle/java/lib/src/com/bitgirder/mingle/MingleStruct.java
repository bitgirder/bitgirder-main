package com.bitgirder.mingle;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class MingleStruct
extends TypedMingleValue
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final MingleSymbolMap flds;

    MingleStruct( AtomicTypeReference type,
                  MingleSymbolMap flds )
    {
        super( type );
        this.flds = flds;
    }

    private
    MingleStruct( Builder b )
    {
        this( inputs.notNull( b.type, "type" ), b.buildMap() );
    }

    public MingleSymbolMap getFields() { return flds; }

    public
    static
    MingleStruct
    create( AtomicTypeReference type,
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
        private AtomicTypeReference type;

        public
        Builder
        setType( AtomicTypeReference type )
        {
            this.type = inputs.notNull( type, "type" );
            return this;
        }

        public
        Builder
        setType( CharSequence type )
        {
            inputs.notNull( type, "type" );

            MingleTypeReference t = MingleTypeReference.create( type );

            inputs.isTrue( t instanceof AtomicTypeReference,
                "Not an atomic type reference:", type );

            return setType( (AtomicTypeReference) t );
        }

        public MingleStruct build() { return new MingleStruct( this ); }
    }
}
