package com.bitgirder.json;

import static com.bitgirder.json.JsonGrammars.SxNt;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Strings;
import com.bitgirder.lang.Lang;

import com.bitgirder.parser.SyntaxBuilder;
import com.bitgirder.parser.ProductionMatch;
import com.bitgirder.parser.DerivationMatch;
import com.bitgirder.parser.TerminalMatch;
import com.bitgirder.parser.SequenceMatch;
import com.bitgirder.parser.UnionMatch;
import com.bitgirder.parser.Parsers;

import java.math.BigInteger;
import java.math.BigDecimal;

import java.util.List;

final
class JsonModelBuilder< T extends JsonText >
implements SyntaxBuilder< SxNt, JsonToken, T >
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final Class< T > cls;

    JsonModelBuilder( Class< T > cls )
    {
        this.cls = state.notNull( cls, "cls" );
    }

    private
    boolean
    buildDecimal( JsonTokenNumber numTok,
                  StringBuilder numStr )
    {
        boolean res = false; // integral until proven decimal

        if ( numTok.getFracPart() != null )
        {
            res = true;
            numStr.append( '.' );
            numStr.append( numTok.getFracPart() );
        }

        if ( numTok.getExponent() != null )
        {
            res = true;
            
            JsonTokenNumber.Exponent exp = numTok.getExponent();

            numStr.append( 'e' );
            if ( exp.isNegated() ) numStr.append( '-' );
            numStr.append( exp.getNumber() );
        }

        return res;
    }

    private
    JsonNumber
    buildJsonNumber( ProductionMatch< JsonToken > pm )
    {
        JsonTokenNumber numTok = 
            Parsers.extractTerminal( JsonTokenNumber.class, pm );

        StringBuilder numStr = new StringBuilder();

        if ( numTok.isNegated() ) numStr.append( '-' );
        numStr.append( numTok.getIntPart() );

        boolean isDecimal = buildDecimal( numTok, numStr );

        Number num = isDecimal 
            ? new BigDecimal( numStr.toString() )
            : new BigInteger( numStr.toString() );
        
        return JsonNumber.forNumber( num );
    }

    private
    JsonString 
    buildJsonString( ProductionMatch< JsonToken > pm )
    {
        TerminalMatch< JsonToken > tm = Parsers.castTerminalMatch( pm );
        JsonTokenString jts = (JsonTokenString) tm.getTerminal();

        return JsonString.create( jts.toJavaString() );
    }

    private
    JsonValue
    buildJsonValue( DerivationMatch< SxNt, JsonToken > dm )
    {
        state.equal( SxNt.JSON_VALUE, dm.getHead() );

        UnionMatch< JsonToken > um = Parsers.castUnionMatch( dm.getMatch() );

        int alt = um.getAlternative();
        ProductionMatch< JsonToken > pm2 = um.getMatch();

        switch ( alt )
        {
            case 0: return JsonNull.INSTANCE;
            case 1: return JsonBoolean.FALSE;
            case 2: return JsonBoolean.TRUE;
            case 3: return buildJsonObject( pm2 );
            case 4: return buildJsonArray( pm2 );
            case 5: return buildJsonNumber( pm2 );
            case 6: return buildJsonString( pm2 );

            default: throw state.createFail( "Unexpected alternative:", alt );
        }
    }
 
    private
    void
    addObjectMember( JsonObject.Builder b,
                     DerivationMatch< SxNt, JsonToken > dm )
    {
        SequenceMatch< JsonToken > sm = 
            Parsers.castSequenceMatch( dm.getMatch() );

        JsonString key = buildJsonString( sm.get( 0 ) );

        DerivationMatch< SxNt, JsonToken > dm2 =
            Parsers.castDerivationMatch( sm.get( 2 ) );
        JsonValue val = buildJsonValue( dm2 );

        b.addMember( key, val );
    }

    private
    JsonObject
    buildJsonObject( ProductionMatch< JsonToken > pm )
    {
        List< DerivationMatch< SxNt, JsonToken > > derivs =
            Parsers.extractDerivations( pm, SxNt.JSON_OBJECT_MEMBER );

        JsonObject.Builder b = new JsonObject.Builder();

        for ( DerivationMatch< SxNt, JsonToken > deriv : derivs )
        {
            addObjectMember( b, deriv );
        }

        return b.build();
    }

    private
    JsonArray
    buildJsonArray( ProductionMatch< JsonToken > pm )
    {
        List< DerivationMatch< SxNt, JsonToken > > derivs =
            Parsers.extractDerivations( pm, SxNt.JSON_VALUE );

        JsonArray.Builder b = new JsonArray.Builder();

        for ( DerivationMatch< SxNt, JsonToken > deriv : derivs )
        {
            b.add( buildJsonValue( deriv ) );
        }

        return b.build();
    }

    private
    DerivationMatch< SxNt, JsonToken >
    extractTextMatch( DerivationMatch< SxNt, JsonToken > dm )
    {
        state.equal( SxNt.JSON_TEXT, dm.getHead() );

        UnionMatch< JsonToken > um = Parsers.castUnionMatch( dm.getMatch() );

        DerivationMatch< SxNt, JsonToken > res =
            Parsers.castDerivationMatch( um.getMatch() );
 
        return res;
    }

    public
    T
    buildSyntax( DerivationMatch< SxNt, JsonToken > dm )
    {
        DerivationMatch< SxNt, JsonToken > textMatch = extractTextMatch( dm );

        ProductionMatch< JsonToken > pm = textMatch.getMatch();
        SxNt head = textMatch.getHead();

        JsonText res;

        switch ( head )
        {
            case JSON_OBJECT: res = buildJsonObject( pm ); break;
            case JSON_ARRAY: res = buildJsonArray( pm ); break;

            default: throw state.createFail( "Unrecognized head:", head );
        }

        return cls.cast( res );
    }
}
