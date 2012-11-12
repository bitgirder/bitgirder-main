package com.bitgirder.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

// straight mapping from java.sql.Types
public
enum SqlType
{
    ARRAY( 2003 ),
    BIGINT( -5 ),
    BINARY( -2 ),
    BIT( -7 ),
    BLOB( 2004 ),
    BOOLEAN( 16 ),
    CHAR( 1 ),
    CLOB( 2005 ),
    DATALINK( 70 ),
    DATE( 91 ),
    DECIMAL( 3 ),
    DISTINCT( 2001 ),
    DOUBLE( 8 ),
    FLOAT( 6 ),
    INTEGER( 4 ),
    JAVA_OBJECT( 2000 ),
    LONGNVARCHAR( -16 ),
    LONGVARBINARY( -4 ),
    LONGVARCHAR( -1 ),
    NCHAR( -15 ),
    NCLOB( 2011 ),
    NULL( 0 ),
    NUMERIC( 2 ),
    NVARCHAR( -9 ),
    OTHER( 1111 ),
    REAL( 7 ),
    REF( 2006 ),
    ROWID( -8 ),
    SMALLINT( 5 ),
    SQLXML( 2009 ),
    STRUCT( 2002 ),
    TIME( 92 ),
    TIMESTAMP( 93 ),
    TINYINT( -6 ),
    VARBINARY( -3 ),
    VARCHAR( 12 );

    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final int type; // the value from java.sql.Types

    private SqlType( int type ) { this.type = type; }

    public int jdbcType() { return type; }

    public
    static
    SqlType
    fromJdbcType( int c )
    {
        switch( c )
        {
            case 2003: return ARRAY;
            case -5: return BIGINT;
            case -2: return BINARY;
            case -7: return BIT;
            case 2004: return BLOB;
            case 16: return BOOLEAN;
            case 1: return CHAR;
            case 2005: return CLOB;
            case 70: return DATALINK;
            case 91: return DATE;
            case 3: return DECIMAL;
            case 2001: return DISTINCT;
            case 8: return DOUBLE;
            case 6: return FLOAT;
            case 4: return INTEGER;
            case 2000: return JAVA_OBJECT;
            case -16: return LONGNVARCHAR;
            case -4: return LONGVARBINARY;
            case -1: return LONGVARCHAR;
            case -15: return NCHAR;
            case 2011: return NCLOB;
            case 0: return NULL;
            case 2: return NUMERIC;
            case -9: return NVARCHAR;
            case 1111: return OTHER;
            case 7: return REAL;
            case 2006: return REF;
            case -8: return ROWID;
            case 5: return SMALLINT;
            case 2009: return SQLXML;
            case 2002: return STRUCT;
            case 92: return TIME;
            case 93: return TIMESTAMP;
            case -6: return TINYINT;
            case -3: return VARBINARY;
            case 12: return VARCHAR;
    
            default: throw inputs.createFail( "Unrecognized type code:", c );
        }
    }
}
