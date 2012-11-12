package com.bitgirder.sql.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.util.Iterator;
import java.util.List;

public
final
class SqlStatementWriter
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private
    void
    writeId( SqlId id,
             StringBuilder sb )
    {
        sb.append( '`' );
        sb.append( id );
        sb.append( '`' );
    }

    private
    void
    writeNumber( SqlNumberLiteral e,
                 StringBuilder sb )
    {
        sb.append( e.number() );
    }

    // Assumes that enclosing quotes are single quotes. Note that we explicitly
    // handle the case of ch being '"', even though it really is just the
    // default behavior, in order to call out that a decision is nevertheless
    // being made here to treat it as a normal char and to help ensure that we
    // handle it should our larger assumptions change (namely that ch is part of
    // a single-quoted literal)
    private
    void
    writeCharLiteral( char ch,
                      StringBuilder sb )
    {
        switch ( ch )
        {
            case '\u0000': sb.append( "\\0" ); break;
            case '\'': sb.append( "\\'" ); break;
            case '"': sb.append( '"' ); break;
            case '\b': sb.append( "\\b" ); break;
            case '\n': sb.append( "\\n" ); break;
            case '\r': sb.append( "\\r" ); break;
            case '\t': sb.append( "\\t" ); break;
            case '\u0026': sb.append( "\\Z" ); break;
            case '\\': sb.append( "\\\\" ); break;
            
            default: sb.append( ch );
        }
    }

    private
    void
    writeString( SqlStringLiteral s,
                 StringBuilder sb )
    {
        sb.append( '\'' );

        for ( int i = 0, e = s.length(); i < e; ++i ) 
        {
            writeCharLiteral( s.charAt( i ), sb );
        }

        sb.append( '\'' );
    }

    private
    void
    writeBoundParamTarget( SqlBoundParamTarget e,
                           StringBuilder sb )
    {
        sb.append( '?' );
    }

    private
    void
    writeParen( SqlParenthesizedExpression e,
                StringBuilder sb )
    {
        sb.append( "( " );
        doWriteExpression( e.expression(), sb );
        sb.append( " )" );
    }

    private
    void
    writeColumnSelectors( List< SqlExpression > cols,
                          StringBuilder sb )
    {
        for ( Iterator< SqlExpression > it = cols.iterator(); it.hasNext(); )
        {
            doWriteExpression( it.next(), sb );
            if ( it.hasNext() ) sb.append( ", " );
        }
    }

    private
    void
    writeSelectTableReferences( SqlSelectExpression e,
                                StringBuilder sb )
    {
        Iterator< SqlExpression > it = e.tableReferences().iterator();

        if ( it.hasNext() )
        {
            sb.append( " from " );

            while ( it.hasNext() )
            {
                doWriteExpression( it.next(), sb );
                if ( it.hasNext() ) sb.append( ", " );
            }
        }
    }

    private
    void
    writeOptWhereClause( SqlExpression where,
                         StringBuilder sb )
    {
        if ( where != null )
        {
            sb.append( " where " );
            doWriteExpression( where, sb );
        }
    }

    private
    void
    writeOrderBy( SqlOrderBy o,
                  StringBuilder sb )
    {
        doWriteExpression( o.expression(), sb );

        SqlOrderBy.Order ordr = o.order();

        if ( ordr != null )
        {
            sb.append( ' ' );
            sb.append( ordr.name().toLowerCase() );
        }
    }

    private
    void
    writeOrderByList( List< SqlOrderBy > l,
                      StringBuilder sb )
    {
        if ( ! l.isEmpty() )
        {
            sb.append( " order by " );

            for ( Iterator< SqlOrderBy > it = l.iterator(); it.hasNext(); )
            {
                writeOrderBy( it.next(), sb );
                if ( it.hasNext() ) sb.append( ", " );
            }
        }
    }

    private
    void
    writeSelect( SqlSelectExpression e,
                 StringBuilder sb )
    {
        sb.append( "select " );
        writeColumnSelectors( e.columns(), sb );
        writeSelectTableReferences( e, sb );
        writeOptWhereClause( e.where(), sb );
        writeOrderByList( e.orderBy(), sb );
    }

    private
    void
    writeAlias( SqlAliasExpression e,
                StringBuilder sb )
    {
        doWriteExpression( e.expression(), sb );
        sb.append( " as " );
        writeId( e.id(), sb );
    }

    private
    void
    writeInfix( SqlInfixExpression e,
                StringBuilder sb )
    {
        doWriteExpression( e.left(), sb );
        sb.append( ' ' );
        sb.append( e.operator() );
        sb.append( ' ' );
        doWriteExpression( e.right(), sb );
    }

    // main entry point for public writeExpression method (which checks its
    // inputs) and internal calls (which need not do so)
    private
    void
    doWriteExpression( SqlExpression e,
                       StringBuilder sb )
    {
        if ( e instanceof SqlSelectExpression )
        {
            writeSelect( (SqlSelectExpression) e, sb );
        }
        else if ( e instanceof SqlId ) writeId( (SqlId) e, sb );
        else if ( e instanceof SqlAliasExpression )
        {
            writeAlias( (SqlAliasExpression) e, sb );
        }
        else if ( e instanceof SqlInfixExpression )
        {
            writeInfix( (SqlInfixExpression) e, sb );
        }
        else if ( e instanceof SqlParenthesizedExpression )
        {
            writeParen( (SqlParenthesizedExpression) e, sb );
        }
        else if ( e instanceof SqlNumberLiteral )
        {
            writeNumber( (SqlNumberLiteral) e, sb );
        }
        else if ( e instanceof SqlStringLiteral )
        {
            writeString( (SqlStringLiteral) e, sb );
        }
        else if ( e instanceof SqlBoundParamTarget )
        {
            writeBoundParamTarget( (SqlBoundParamTarget) e, sb );
        }
        else 
        {
            throw state.createFail( 
                "Unhandled expression type:", e.getClass().getName() );
        }
    }

    public
    void
    writeExpression( SqlExpression e,
                     StringBuilder sb )
    {
        inputs.notNull( e, "e" );
        inputs.notNull( sb, "sb" );

        doWriteExpression( e, sb );
    }
}
