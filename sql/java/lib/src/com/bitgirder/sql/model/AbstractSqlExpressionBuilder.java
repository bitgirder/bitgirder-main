package com.bitgirder.sql.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

public
abstract
class AbstractSqlExpressionBuilder< B extends AbstractSqlExpressionBuilder >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    AbstractSqlExpressionBuilder() {}

    final B castThis() { return Lang.< B >castUnchecked( this ); }

    public
    final
    SqlId
    id( CharSequence id )
    {
        return new SqlId( inputs.notNull( id, "id" ) );
    }

    public
    final
    SqlBoundParamTarget
    paramTarget()
    {
        return SqlBoundParamTarget.INSTANCE;
    }

    public
    final
    SqlNumberLiteral
    num( Number num )
    {
        return new SqlNumberLiteral( inputs.notNull( num, "num" ) );
    }

    public
    final
    SqlInfixExpression
    infix( SqlExpression left,
           SqlOperator op,
           SqlExpression right )
    {
        inputs.notNull( left, "left" );
        inputs.notNull( op, "op" );
        inputs.notNull( right, "right" );
        
        return new SqlInfixExpression( left, op, right );
    }

    public
    final
    SqlInfixExpression
    eq( SqlExpression left,
        SqlExpression right )
    {
        return infix( left, SqlOperator.EQUALS, right );
    }

    public
    final
    SqlInfixExpression
    lt( SqlExpression left,
        SqlExpression right )
    {
        return infix( left, SqlOperator.LESS_THAN, right );
    }

    public
    final
    SqlInfixExpression
    gt( SqlExpression left,
        SqlExpression right )
    {
        return infix( left, SqlOperator.GREATER_THAN, right );
    }

    public
    final
    SqlInfixExpression
    plus( SqlExpression left,
          SqlExpression right )
    {
        return infix( left, SqlOperator.PLUS, right );
    }

    public
    final
    SqlInfixExpression
    and( SqlExpression left,
         SqlExpression right )
    {
        return infix( left, SqlOperator.AND, right );
    }

    public
    final
    SqlInfixExpression
    or( SqlExpression left,
        SqlExpression right )
    {
        return infix( left, SqlOperator.OR, right );
    }

    public
    final
    SqlParenthesizedExpression
    paren( SqlExpression e )
    {
        return new SqlParenthesizedExpression( inputs.notNull( e, "e" ) );
    }

    public
    final
    SqlAliasExpression
    alias( SqlExpression expr,
           SqlId id )
    {
        inputs.notNull( expr, "expr" );
        inputs.notNull( id, "id" );

        return new SqlAliasExpression( expr, id );
    }
}
