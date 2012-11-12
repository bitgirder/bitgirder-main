package com.bitgirder.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

public
final
class SqlTableDescriptor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final String name;
    private final String catalog;
    private final List< SqlColumnDescriptor > cols;

    private
    SqlTableDescriptor( Builder b )
    {
        this.name = inputs.notNull( b.name, "name" );
        this.catalog = inputs.notNull( b.catalog, "catalog" );
        
        this.cols = Lang.unmodifiableCopy( b.cols, "cols" );
        inputs.isFalse( cols.isEmpty(), "Table has no columns" );
    }

    public String getName() { return name; }
    public String getCatalog() { return catalog; }
    public List< SqlColumnDescriptor > getColumns() { return cols; }

    public
    final
    static
    class Builder
    {
        private String name;
        private String catalog;
        private List< SqlColumnDescriptor > cols;

        public
        Builder
        setName( String name )
        {
            this.name = inputs.notNull( name, "name" );
            return this;
        }

        public
        Builder
        setCatalog( String catalog )
        {
            this.catalog = inputs.notNull( catalog, "catalog" );
            return this;
        }

        public
        Builder
        setColumns( List< SqlColumnDescriptor > cols )
        {
            this.cols = inputs.noneNull( cols, "cols" );
            return this;
        }

        public
        SqlTableDescriptor
        build()
        {
            return new SqlTableDescriptor( this );
        }
    }
}
