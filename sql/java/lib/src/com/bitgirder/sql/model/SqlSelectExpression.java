package com.bitgirder.sql.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.util.List;

public
final
class SqlSelectExpression
implements SqlExpression
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final List< SqlExpression > cols; // nonempty
    private final List< SqlExpression > tblRefs; // maybe empty
    private final SqlExpression where; // maybe null
    private final List< SqlOrderBy > orderBy; // maybe empty

    private
    SqlSelectExpression( Builder b )
    {
        this.cols = Lang.unmodifiableCopy( b.cols );
        this.tblRefs = Lang.unmodifiableCopy( b.tblRefs );
        this.where = b.where;
        this.orderBy = Lang.unmodifiableCopy( b.orderBy );
    }

    public List< SqlExpression > columns() { return cols; }
    public List< SqlExpression > tableReferences() { return tblRefs; }
    public SqlExpression where() { return where; }
    public List< SqlOrderBy > orderBy() { return orderBy; }

    public
    final
    static
    class Builder
    extends AbstractSqlExpressionBuilder< Builder >
    {
        private final List< SqlExpression > cols = Lang.newList();
        private final List< SqlExpression > tblRefs = Lang.newList();
        private SqlExpression where;
        private final List< SqlOrderBy > orderBy = Lang.newList();

        public
        Builder
        addColumn( SqlExpression expr )
        {
            cols.add( inputs.notNull( expr, "expr" ) );
            return this;
        }

        public
        Builder
        addTableReference( SqlExpression expr )
        {
            tblRefs.add( inputs.notNull( expr, "expr" ) );
            return this;
        }

        public
        Builder
        setWhere( SqlExpression expr )
        {
            this.where = inputs.notNull( expr, "expr" );
            return this;
        }

        public
        Builder
        addOrderBy( SqlOrderBy ob )
        {
            orderBy.add( inputs.notNull( ob, "ob" ) );
            return this;
        }

        public
        Builder
        addOrderBy( SqlExpression expr,
                    SqlOrderBy.Order ordr )
        {
            inputs.notNull( expr, "expr" );
            inputs.notNull( ordr, "ordr" );

            return addOrderBy( SqlOrderBy.create( expr, ordr ) );
        }

        public
        Builder
        addOrderBy( SqlExpression expr,
                    CharSequence ordrStr )
        {
            inputs.notNull( ordrStr, "ordrStr" );
            
            SqlOrderBy.Order ordr = 
                SqlOrderBy.Order.valueOf( ordrStr.toString().toUpperCase() );

            return addOrderBy( expr, ordr );
        }

        public
        Builder
        addOrderBy( SqlExpression expr )
        {
            inputs.notNull( expr, "expr" );
            return addOrderBy( SqlOrderBy.create( expr ) );
        }

        public
        SqlSelectExpression
        build()
        {
            return new SqlSelectExpression( this );
        }
    }
}
