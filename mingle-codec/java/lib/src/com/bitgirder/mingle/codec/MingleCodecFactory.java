package com.bitgirder.mingle.codec;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.log.CodeLoggers;

import com.bitgirder.lang.Lang;
import com.bitgirder.lang.Strings;

import com.bitgirder.mingle.model.MingleIdentifier;

import java.util.Map;

import java.nio.ByteBuffer;

public
final
class MingleCodecFactory
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static void code( Object... msg ) { CodeLoggers.code( msg ); }

    private final Map< MingleIdentifier, MingleCodec > codecs;
    private final Map< MingleIdentifier, MingleCodecDetectorFactory > detFacts;

    private
    MingleCodecFactory( Builder b )
    {
        this.codecs = Lang.unmodifiableCopy( b.codecs );
        this.detFacts = Lang.unmodifiableCopy( b.detFacts );
    }
    
    public
    MingleCodec
    expectCodec( MingleIdentifier id )
        throws NoSuchMingleCodecException
    {
        inputs.notNull( id, "id" );

        MingleCodec res = codecs.get( id );

        if ( res == null ) throw new NoSuchMingleCodecException( id );
        else return res;
    }

    public
    MingleCodec
    expectCodec( CharSequence id )
        throws NoSuchMingleCodecException
    {
        inputs.notNull( id, "id" );

        return expectCodec( MingleIdentifier.create( id ) );
    }

    private
    final
    class DetectionImpl
    implements MingleCodecDetection
    {
        // The active detections
        private final Map< MingleIdentifier, MingleCodecDetector > dets =
            Lang.newMap();

        // On success will have exactly one entry
        private final Map< MingleIdentifier, MingleCodec > matched =
            Lang.newMap();

        private
        DetectionImpl()
        {
            for ( Map.Entry< MingleIdentifier, MingleCodecDetectorFactory > e :
                    detFacts.entrySet() )
            {
                dets.put( e.getKey(), e.getValue().createCodecDetector() );
            }
        }

        public
        boolean
        update( ByteBuffer bb )
            throws Exception
        {
            inputs.notNull( bb, "bb" );
            inputs.isTrue( bb.hasRemaining(), "bb is empty" );

            // Iterate over a copy since we may be changing dets during the
            // iteration
            for ( MingleIdentifier id : Lang.newList( dets.keySet() ) )
            {
                MingleCodecDetector det = dets.get( id );
                Boolean res = det.update( bb.duplicate() );

                if ( res != null ) 
                {
                    dets.remove( id );
                    if ( res ) matched.put( id, codecs.get( id ) );
                }
            }

            bb.position( bb.limit() );
            return dets.isEmpty();
        }

        private
        void
        checkDone()
            throws MingleCodecException
        {
            if ( ! dets.isEmpty() )
            {
                throw MingleCodecs.createDetectionNotCompletedException();
            }
        }

        public
        MingleCodec
        getResult()
            throws MingleCodecException
        {
            checkDone();

            switch ( matched.size() )
            {
                case 0: throw new NoSuchMingleCodecException();
                case 1: return matched.values().iterator().next();

                default:
                    throw 
                        new MingleCodecDetectionException(
                            "Multiple matching codecs detected: " +
                            Strings.join( ", ", matched.keySet() )
                        );
            }
        }
    }

    public
    MingleCodecDetection
    createCodecDetection()
    {
        return new DetectionImpl();
    }

    public
    final
    static
    class Builder
    {
        private final Map< MingleIdentifier, MingleCodec > codecs = 
            Lang.newMap();

        private final Map< MingleIdentifier, MingleCodecDetectorFactory > 
            detFacts = Lang.newMap();
        
        private
        Builder
        doAddCodec( MingleIdentifier id,
                    MingleCodec codec,
                    MingleCodecDetectorFactory detFact )
        {
            inputs.notNull( id, "id" );
            inputs.notNull( codec, "codec" );

            Lang.putUnique( codecs, id, codec );
            
            if ( detFact != null ) Lang.putUnique( detFacts, id, detFact );

            return this;
        }

        public
        Builder
        addCodec( MingleIdentifier id,
                  MingleCodec codec,
                  MingleCodecDetectorFactory detFact )
        {
            inputs.notNull( detFact, "detFact" );
            return doAddCodec( id, codec, detFact );
        }

        public
        Builder
        addCodec( CharSequence id,
                  MingleCodec codec,
                  MingleCodecDetectorFactory detFact )
        {
            return
                addCodec(
                    MingleIdentifier.create( inputs.notNull( id, "id" ) ),
                    codec,
                    detFact
                );
        }

        public
        Builder
        addCodec( MingleIdentifier id,
                  MingleCodec codec )
        {
            return doAddCodec( id, codec, null );
        }

        public
        Builder
        addCodec( CharSequence id,
                  MingleCodec codec )
        {
            return
                addCodec(
                    MingleIdentifier.create( inputs.notNull( id, "id" ) ),
                    codec
                );
        }

        public
        MingleCodecFactory
        build()
        {
            return new MingleCodecFactory( this );
        }
    }
}
