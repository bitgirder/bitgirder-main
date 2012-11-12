package com.bitgirder.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.sql.DatabaseMetaData;
import java.sql.SQLException;

// This may ultimately be rebuilt as an spi interface or base class
public
final
class SqlStatementWriter
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    // will be null if DatabaseMetaData.getIdentifierQuoteString() is " "
    private final String identQuote; 

    private final String catalogSep;

    private
    SqlStatementWriter( String identQuote,
                        String catalogSep )
    {
        this.identQuote = identQuote;
        this.catalogSep = catalogSep;
    }

    public String getIdentifierQuoteString() { return identQuote; }
    public String getCatalogSeparator() { return catalogSep; }

    static
    SqlStatementWriter
    create( DatabaseMetaData md )
        throws SQLException
    {
        state.notNull( md, "md" );

        String identQuote = md.getIdentifierQuoteString();

        return
            new SqlStatementWriter(
                identQuote.equals( " " ) ? null : identQuote,
                md.getCatalogSeparator()
            );
    }
}
