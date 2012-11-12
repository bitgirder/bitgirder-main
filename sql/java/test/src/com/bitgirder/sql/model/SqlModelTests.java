package com.bitgirder.sql.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.test.Test;
import com.bitgirder.test.LabeledTestCall;
import com.bitgirder.test.InvocationFactory;

import java.util.List;

@Test
final
class SqlModelTests
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final static SqlSelectExpression SELECT_STMT1;

    private
    class ExpressionWriteTest
    extends LabeledTestCall
    {
        SqlExpression expr;
        CharSequence expct;
        
        private ExpressionWriteTest( CharSequence lbl ) { super( lbl ); }

        private
        void
        validate()
        {
            state.notNull( expr, "expr" );
            state.notNull( expct, "expct" );
        }

        private
        CharSequence
        writeExpression()
        {
            return SqlStatementWriters.writeExpression( expr );
        }

        public
        void
        call()
        {
            validate();

            CharSequence sql = writeExpression();
            state.equalString( expct, sql );
            code( "checked:", sql, "; expct:", expct );
        }
    }

    @InvocationFactory
    private
    List< ExpressionWriteTest >
    testWriteExpression()
    {
        return Lang.< ExpressionWriteTest >asList(
            
            new ExpressionWriteTest( "select-stmt1" ) 
            {{
                expr = SELECT_STMT1;

                expct = 
                    "select `col1`, `col1` + `col2` as `sum` from `t1` " +
                    "where ( `col1` < `sum` and 1 + `col2` > 4 ) " +
                    "or `col1` = ? order by `col1`, `col2` desc";
            }},

            new ExpressionWriteTest( "basic-string" )
            {{
                expr = new SqlStringLiteral( "hello" );
                expct = "'hello'";
            }},

            new ExpressionWriteTest( "empty-string" )
            {{
                expr = new SqlStringLiteral( "" );
                expct = "''";
            }},

            new ExpressionWriteTest( "conservative-quoting" )
            {{
                expr = 
                    new SqlStringLiteral( 
                        "abc" + // normal text
                        "\u0000" + // \0
                        "'" + // '
                        "\"" + // "
                        "\b\n\r\t\u0026" + // \b\n\r\t\Z
                        "\\" + // \
                        "%\\%_\\_" // check that %, \%, _, and \_ are unchanged
                    );

                expct = "'abc\\0\\'\"\\b\\n\\r\\t\\Z\\\\%\\\\%_\\\\_'";
            }}
        );
    }

    static
    {
        SqlSelectExpression.Builder b = new SqlSelectExpression.Builder();

        b.addColumn( b.id( "col1" ) ).
          addColumn( 
            b.alias( b.plus( b.id( "col1" ), b.id( "col2" ) ), b.id( "sum" ) ) 
          ).
          addTableReference( b.id( "t1" ) ).
          setWhere(
            b.or(
                b.paren(
                    b.and(
                        b.lt( b.id( "col1" ), b.id( "sum" ) ),
                        b.gt( b.plus( b.num( 1 ), b.id( "col2" ) ), b.num( 4 ) )
                    )
                ),
                b.eq( b.id( "col1" ), b.paramTarget() )
            )
          ).
          addOrderBy( b.id( "col1" ) ).
          addOrderBy( b.id( "col2" ), "desc" );
        
        SELECT_STMT1 = b.build();
    }
}
