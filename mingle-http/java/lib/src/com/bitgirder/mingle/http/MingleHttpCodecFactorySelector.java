package com.bitgirder.mingle.http;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.PatternHelper;
import com.bitgirder.lang.Lang;

import com.bitgirder.mingle.codec.MingleCodecException;

import com.bitgirder.http.HttpRequestMessage;

import java.util.List;

import java.util.regex.Pattern;

public
final
class MingleHttpCodecFactorySelector
implements MingleHttpCodecFactory
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    // This is less strict than the rfc2616 grammar, but we don't select based
    // on anything after the ctype at this point and don't make any attempts to
    // interact with it beyond acknowleging and ignoring its possible presence
    private final static String PARAM_TAIL = "(\\s*;.*)?";

    private final List< SelectorContext > selectors;

    private
    MingleHttpCodecFactorySelector( Builder b )
    {
        this.selectors = Lang.unmodifiableCopy( b.selectors );
    }

    private
    final
    static
    class SelectorContext
    {
        private final Pattern pat;
        private final MingleHttpCodecFactory fact;

        private
        SelectorContext( Pattern pat,
                         MingleHttpCodecFactory fact )
        {
            this.pat = pat;
            this.fact = fact;
        }
    }

    private
    RuntimeException
    createAmbiguousMatchException( CharSequence ctype,
                                   List< SelectorContext > matches )
    {
        List< Pattern > pats = Lang.newList( matches.size() );
        for ( SelectorContext match : matches ) pats.add( match.pat );

        return
            state.createFail(
                "Multiple selector patterns matched content type '" + ctype +
                "':", pats
            );
    }

    private
    MingleHttpCodecContext
    matchResultOf( HttpRequestMessage req,
                   CharSequence ctype,
                   List< SelectorContext > matches )
        throws MingleCodecException
    {
        if ( matches.size() == 0 )
        {
            String msg = "Unrecognized content type: " + ctype;
            throw new MingleCodecException( msg );
        }
        else if ( matches.size() == 1 ) 
        {
            return matches.get( 0 ).fact.codecContextFor( req );
        }
        else throw createAmbiguousMatchException( ctype, matches );
    }

    private
    MingleHttpCodecContext
    matchCtype( HttpRequestMessage req,
                CharSequence ctype )
        throws MingleCodecException
    {
        List< SelectorContext > matches = Lang.newList();

        for ( SelectorContext sc : selectors )
        {
            if ( sc.pat.matcher( ctype ).matches() ) matches.add( sc );
        }

        return matchResultOf( req, ctype, matches );
    }

    // Throws IllegalStateException if multiple selectors could match req, since
    // this is deemed a programming error; if req has no ctype or ctype is not
    // recognized, this is thrown as a codec exception suitable to report to the
    // client. Also throws any exception from the selected factory's call to
    // codecContextFor()
    public
    MingleHttpCodecContext
    codecContextFor( HttpRequestMessage req )
        throws MingleCodecException
    {
        inputs.notNull( req, "req" );

        CharSequence ctype = req.h().getContentTypeString();

        if ( ctype == null ) 
        {
            throw new MingleCodecException( "Missing content type" );
        }
        else return matchCtype( req, ctype );
    }

    public
    final
    static
    class Builder
    {
        private final List< SelectorContext > selectors = Lang.newList();

        public
        final
        Builder
        selectRegex( Pattern pat,
                     MingleHttpCodecFactory fact )
        {
            inputs.notNull( pat, "pat" );
            inputs.notNull( fact, "fact" );

            selectors.add( new SelectorContext( pat, fact ) );
            return this;
        }

        public
        final
        Builder
        selectRegex( CharSequence pat,
                     MingleHttpCodecFactory fact )
        {
            inputs.notNull( pat, "pat" );

            return selectRegex( PatternHelper.compile( pat ), fact );
        }

        public
        final
        Builder
        selectFullType( CharSequence type,
                        MingleHttpCodecFactory fact )
        {
            inputs.notNull( type, "type" );

            return selectRegex( type + PARAM_TAIL, fact );
        }

        public
        final
        Builder
        selectSubType( CharSequence subType,
                       MingleHttpCodecFactory fact )
        {
            inputs.notNull( subType, "subType" );

            return selectRegex( "^[^/]+/" + subType + PARAM_TAIL, fact );
        }

        public
        MingleHttpCodecFactorySelector
        build()
        {
            return new MingleHttpCodecFactorySelector( this );
        }
    }
}
