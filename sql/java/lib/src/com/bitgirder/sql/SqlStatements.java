package com.bitgirder.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.util.Iterator;

public
final
class SqlStatements
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private
    static
    StringBuilder
    appendIdent( StringBuilder sb,
                 CharSequence ident,
                 SqlStatementWriter w )
    {
        String qt = w.getIdentifierQuoteString();

        if ( qt != null ) sb.append( qt );
        sb.append( ident );
        if ( qt != null ) sb.append( qt );

        return sb;
    }

    private
    static
    StringBuilder
    appendTableName( StringBuilder sb,
                     SqlTableDescriptor td,
                     SqlStatementWriter w )
    {
        appendIdent( sb, td.getCatalog(), w );
        sb.append( w.getCatalogSeparator() );
        appendIdent( sb, td.getName(), w );

        return sb;
    }

    private
    static
    StringBuilder
    appendColumnList( StringBuilder sb,
                      SqlTableDescriptor td,
                      SqlStatementWriter w )
    {
        sb.append( " ( " );

        for ( Iterator< SqlColumnDescriptor > it = td.getColumns().iterator();
                it.hasNext(); )
        {
            appendIdent( sb, it.next().getName(), w );
            sb.append( it.hasNext() ? ", " : " ) " );
        }

        return sb;
    } 

    private
    static
    StringBuilder
    appendValuesClause( StringBuilder sb,
                        SqlTableDescriptor td,
                        SqlStatementWriter w )
    {
        sb.append( " values ( " );

        for ( int i = 0, e = td.getColumns().size(); i < e; ++i )
        {
            sb.append( "?" );
            sb.append( i == e - 1 ? " )" : ", " );
        }

        return sb;
    }

    private
    static
    StringBuilder
    beginInsert( SqlTableDescriptor td,
                 boolean useIgnore,
                 SqlStatementWriter w )
    {
        StringBuilder res = new StringBuilder();

        res.append( "insert" );
        if ( useIgnore ) res.append( " ignore " );
        res.append( " into " );
        appendTableName( res, td, w );
        appendColumnList( res, td, w );
        appendValuesClause( res, td, w );

        return res;
    }

    public
    static
    String
    getInsert( SqlTableDescriptor td,
               boolean useIgnore,
               SqlStatementWriter w )
    {
        inputs.notNull( td, "td" );
        inputs.notNull( w, "w" );

        return beginInsert( td, useIgnore, w ).toString();
    }

    public
    static
    String
    getUpsert( SqlTableDescriptor td,
               SqlStatementWriter w )
    {
        StringBuilder sb = beginInsert( td, false, w );

        sb.append( " on duplicate key update " );

        for ( Iterator< SqlColumnDescriptor > it = td.getColumns().iterator();
                it.hasNext(); )
        {
            SqlColumnDescriptor col = it.next();

            appendIdent( sb, col.getName(), w );
            sb.append( " = values( " );
            appendIdent( sb, col.getName(), w );
            sb.append( " ) " );

            if ( it.hasNext() ) sb.append( " , " );
        }

        return sb.toString();
    }
}
