package com.bitgirder.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class SqlColumnDescriptor
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final String name;
    private final SqlType sqlType;
    private final int size;

    private
    SqlColumnDescriptor( Builder b )
    {
        this.name = inputs.notNull( b.name, "name" );
        this.sqlType = inputs.notNull( b.sqlType, "sqlType" );
        this.size = b.size;
    }

    public String getName() { return name; }
    public SqlType getSqlType() { return sqlType; }
    public int getSize() { return size; }

    public
    final
    static
    class Builder
    {
        private String name;
        private SqlType sqlType;
        private int size;

        public
        Builder
        setName( String name )
        {
            this.name = inputs.notNull( name, "name" );
            return this;
        }

        public
        Builder
        setSqlType( SqlType sqlType )
        {
            this.sqlType = inputs.notNull( sqlType, "sqlType" );
            return this;
        }

        public
        Builder
        setSize( Integer size )
        {
            this.size = inputs.notNull( size, "size" );
            return this;
        }

        public
        SqlColumnDescriptor
        build() 
        { 
            return new SqlColumnDescriptor( this ); 
        }
    }
}
