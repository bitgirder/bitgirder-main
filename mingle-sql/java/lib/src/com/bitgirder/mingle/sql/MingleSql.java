package com.bitgirder.mingle.sql;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.MingleString;
import com.bitgirder.mingle.model.MingleBuffer;
import com.bitgirder.mingle.model.MingleInt64;
import com.bitgirder.mingle.model.MingleDouble;
import com.bitgirder.mingle.model.MingleTimestamp;
import com.bitgirder.mingle.model.MingleEnum;
import com.bitgirder.mingle.model.MingleNull;
import com.bitgirder.mingle.model.MingleBoolean;

import com.bitgirder.sql.SqlType;
import com.bitgirder.sql.Sql;

import java.sql.PreparedStatement;
import java.sql.SQLException;

public
final
class MingleSql
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static ObjectPath< MingleIdentifier > ROOT_PATH =
        ObjectPath.getRoot();

    private MingleSql() {}

    private
    static
    < V extends MingleValue >
    V
    cast( MingleValue mv,
          Class< V > cls )
    {
        return 
            cls.cast(
                MingleModels.asMingleInstance( 
                    MingleModels.typeReferenceOf( cls ),
                    mv,
                    ROOT_PATH 
                )
            );
    }

    private
    static
    MingleInt64
    mgInt( MingleValue mv )
    {
        return cast( mv, MingleInt64.class );
    }

    private
    static
    MingleDouble
    mgDec( MingleValue mv )
    {  
        return cast( mv, MingleDouble.class );
    }

    private
    static
    void
    setCalendar( MingleValue mv,
                 PreparedStatement ps,
                 int indx )
        throws SQLException
    {
        MingleTimestamp mgTm = cast( mv, MingleTimestamp.class );
        ps.setTimestamp( indx, mgTm.asSqlTimestamp(), mgTm.asJavaCalendar() );
    }

    private
    static
    void
    setBuffer( MingleValue mv,
               SqlType sqlTyp,
               PreparedStatement ps,
               int indx )
        throws SQLException
    {
        Sql.setValue(
            cast( mv, MingleBuffer.class ).getByteBuffer(), 
            sqlTyp, 
            ps, 
            indx 
        );
    }

    private
    static
    void
    setNonNullValue( MingleValue mv,
                     SqlType sqlTyp,
                     PreparedStatement ps,
                     int indx )
        throws SQLException
    {
        switch ( sqlTyp )
        {
            case INTEGER:
            case SMALLINT:
            case TINYINT:
            case BIGINT:
                ps.setLong( indx, mgInt( mv ).longValue() ); break;
            
            case NUMERIC:
            case DECIMAL:
            case REAL:
            case DOUBLE:
                ps.setDouble( indx, mgDec( mv ).doubleValue() ); break;
            
            case FLOAT:
                ps.setFloat( indx, mgDec( mv ).floatValue() ); break;

            case CHAR:
            case LONGVARCHAR:
            case VARCHAR:
                Sql.setValue( 
                    cast( mv, MingleString.class ), sqlTyp, ps, indx ); 
                break;

            case BIT:
            case BOOLEAN:
                ps.setBoolean( 
                    indx, cast( mv, MingleBoolean.class ).booleanValue() );
                break;

            case DATE:
            case TIMESTAMP:
            case TIME:
                setCalendar( mv, ps, indx ); break;

            case BINARY:
            case BLOB:
            case LONGVARBINARY:
            case VARBINARY:
                setBuffer( mv, sqlTyp, ps, indx ); break;

            default:
                state.fail(
                    "Don't know how to map", mv.getClass(), "to SqlType",
                    sqlTyp
                );
        }
    }

    public
    static
    void
    setValue( MingleValue mv,
              SqlType sqlTyp,
              PreparedStatement ps,
              int indx )
        throws SQLException
    {
        inputs.notNull( sqlTyp, "sqlTyp" );
        inputs.positiveI( indx, "indx" );
        inputs.notNull( ps, "ps" );

        if ( mv == null || mv instanceof MingleNull ) 
        {
            ps.setNull( indx, sqlTyp.jdbcType() );
        }
        else setNonNullValue( mv, sqlTyp, ps, indx );
    }
}
