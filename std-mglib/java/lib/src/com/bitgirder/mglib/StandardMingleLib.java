package com.bitgirder.mglib;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;

import com.bitgirder.io.DataUnit;
import com.bitgirder.io.DataSize;

import com.bitgirder.mglib.v1.TimeUnitImpl;

import com.bitgirder.concurrent.Duration;

import com.bitgirder.mingle.model.MingleEnum;
import com.bitgirder.mingle.model.MingleInt64;
import com.bitgirder.mingle.model.MingleInt32;
import com.bitgirder.mingle.model.MingleDouble;
import com.bitgirder.mingle.model.MingleFloat;
import com.bitgirder.mingle.model.MingleString;
import com.bitgirder.mingle.model.MingleStructure;
import com.bitgirder.mingle.model.MingleSymbolMapAccessor;
import com.bitgirder.mingle.model.MingleModels;
import com.bitgirder.mingle.model.MingleValue;
import com.bitgirder.mingle.model.MingleIdentifier;
import com.bitgirder.mingle.model.AtomicTypeReference;
import com.bitgirder.mingle.model.QualifiedTypeName;
import com.bitgirder.mingle.model.MingleValidationException;
import com.bitgirder.mingle.model.MingleTypeReference;
import com.bitgirder.mingle.model.MingleTypeCastException;
import com.bitgirder.mingle.model.TypeDefinitionLookup;

import com.bitgirder.mingle.bind.MingleBinder;
import com.bitgirder.mingle.bind.MingleBinding;
import com.bitgirder.mingle.bind.MingleBinders;

final
class StandardMingleLib
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static QualifiedTypeName QNAME_DATA_UNIT =
        QualifiedTypeName.create( "bitgirder:io@v1/DataUnit" );

    private final static AtomicTypeReference TYPE_DATA_UNIT =   
        AtomicTypeReference.create( QNAME_DATA_UNIT );
 
    private final static QualifiedTypeName QNAME_DATA_SIZE =
        QualifiedTypeName.create( "bitgirder:io@v1/DataSize" );

    private final static AtomicTypeReference TYPE_DATA_SIZE =   
        AtomicTypeReference.create( QNAME_DATA_SIZE );
    
    private final static QualifiedTypeName QNAME_DURATION =
        QualifiedTypeName.create( "bitgirder:concurrent@v1/Duration" );

    private final static AtomicTypeReference TYPE_DURATION =
        AtomicTypeReference.create( QNAME_DURATION );

    private final static AtomicTypeReference TYPE_TIME_UNIT =
        (AtomicTypeReference)
            MingleTypeReference.create( "bitgirder:concurrent@v1/TimeUnit" );

    private
    static
    abstract
    class AbstractMetricBinding< V, U >
    implements MingleBinding
    {
        private final AtomicTypeReference mgTyp;
        private final Class< V > cls;
        private final MingleIdentifier szId;
        private final AtomicTypeReference unitTyp;
        private final Class< U > unitCls;
        private final MingleIdentifier unitId;
        private final String jvUnitId;

        private
        AbstractMetricBinding( AtomicTypeReference mgTyp,
                               Class< V > cls,
                               MingleIdentifier szId,
                               AtomicTypeReference unitTyp,
                               Class< U > unitCls,
                               MingleIdentifier unitId,
                               String jvUnitId )
        {
            this.mgTyp = mgTyp;
            this.cls = cls;
            this.szId = szId;
            this.unitTyp = unitTyp;
            this.unitCls = unitCls;
            this.unitId = unitId;
            this.jvUnitId = jvUnitId;
        }

        abstract
        V
        createMetric( long sz,
                      U unit );

        private
        V
        fromMingleStructure( AtomicTypeReference typ,
                             MingleStructure ms,
                             MingleBinder mb,
                             ObjectPath< MingleIdentifier > path )
        {
            MingleSymbolMapAccessor acc =
                MingleModels.expectStruct( ms, path, typ );

            MingleValue unitVal = acc.expectMingleValue( unitId );
            ObjectPath< MingleIdentifier > unitPath = path.descend( unitId );

            U unit = 
                state.cast( 
                    unitCls, 
                    MingleBinders.asJavaValue( mb, unitTyp, unitVal, unitPath )
                );
            
            long sz = 
                getSize( 
                    acc.expectLong( szId ), 
                    acc.getPath().descend( szId ) );

            return createMetric( sz, unit );
        }

        abstract
        V
        implFromString( CharSequence str )
            throws Exception;

        private
        long
        getSize( long l,
                 ObjectPath< MingleIdentifier > path )
        {
            if ( l < 0 )
            {
                throw 
                    new MingleValidationException( 
                        "Value is negative: " + l, path );
            }
            else return l;
        }

        private
        V
        fromString( MingleString ms,
                    ObjectPath< MingleIdentifier > path )
        {
            // Right now we squash anything thrown; later we may look for
            // specific things (SyntaxException) and take the message from those
            // and report it to the mingle caller
            try { return implFromString( ms ); }
            catch ( Throwable th )
            {
                throw new MingleValidationException(
                    "Invalid string for instance of type " + mgTyp + ": " + ms,
                    path, 
                    th
                );
            }
        }

        abstract
        V
        implFromNum( long sz );

        private
        V
        fromNum( MingleValue mv,
                 ObjectPath< MingleIdentifier > path )
        {
            MingleInt64 num =
                (MingleInt64) MingleModels.asMingleInstance(
                    MingleModels.TYPE_REF_MINGLE_INT64, mv, path );
            
            try { return implFromNum( getSize( num.longValue(), path ) ); }
            catch ( Throwable th )
            {
                throw 
                    new MingleValidationException(
                        "Invalid number for instance of type " + mgTyp + ": " +
                            mv,
                        path,
                        th
                    );
            }
        }

        public
        final
        Object
        asJavaValue( AtomicTypeReference typ,
                     MingleValue mv,
                     MingleBinder mb,
                     ObjectPath< MingleIdentifier > path )
        {
            if ( mv instanceof MingleStructure )
            {
                return
                    fromMingleStructure( typ, (MingleStructure) mv, mb, path );
            }
            else if ( mv instanceof MingleString )
            {
                return fromString( (MingleString) mv, path );
            }
            else if ( mv instanceof MingleInt64 ||
                      mv instanceof MingleInt32 ||
                      mv instanceof MingleDouble ||
                      mv instanceof MingleFloat )
            {
                return fromNum( mv, path );
            }
            else 
            {
                throw
                    new MingleTypeCastException( 
                        mgTyp, MingleModels.typeReferenceOf( mv ), path );
            }
        }

        abstract
        U
        unitOf( V inst );

        abstract
        long
        sizeOf( V inst );

        private
        MingleValue
        mgUnitOf( V inst,
                  MingleBinder mb,
                  ObjectPath< String > path )
        {
            U unit = state.notNull( unitOf( inst ) );

            return
                MingleBinders.asMingleValue(
                    mb, unitTyp, unit, path.descend( jvUnitId ) );
        }

        public
        final
        MingleValue
        asMingleValue( Object obj,
                       MingleBinder mb,
                       ObjectPath< String > path )
        {
            V inst = state.cast( cls, obj );

            MingleValue mgUnit = mgUnitOf( inst, mb, path );

            long sz = sizeOf( inst );

            return
                MingleModels.structBuilder().
                    setType( mgTyp ).
                    f().setInt64( szId, sizeOf( inst ) ).
                    f().set( unitId, mgUnit ).
                    build();
        }
    }

    private
    final
    static
    class DataSizeBinding
    extends AbstractMetricBinding< DataSize, DataUnit >
    {
        private
        DataSizeBinding()
        {
            super( 
                TYPE_DATA_SIZE, 
                DataSize.class,
                MingleIdentifier.create( "size" ),
                TYPE_DATA_UNIT,
                DataUnit.class,
                MingleIdentifier.create( "unit" ),
                "unit"
            );
        }

        DataUnit unitOf( DataSize ds ) { return ds.unit(); }
        long sizeOf( DataSize ds ) { return ds.size(); }

        DataSize
        createMetric( long sz,
                      DataUnit unit )
        {
            return new DataSize( sz, unit );
        }

        DataSize
        implFromString( CharSequence cs )
        {
            return DataSize.fromString( cs );
        }

        DataSize implFromNum( long sz ) { return DataSize.ofBytes( sz ); }
    }

    private
    final
    static
    class DurationBinding
    extends AbstractMetricBinding< Duration, TimeUnitImpl >
    {
        private
        DurationBinding()
        {
            super(
                TYPE_DURATION,
                Duration.class,
                MingleIdentifier.create( "duration" ),
                TYPE_TIME_UNIT,
                TimeUnitImpl.class,
                MingleIdentifier.create( "unit" ),
                "timeUnit"
            );
        }

        TimeUnitImpl
        unitOf( Duration dur )
        {
            switch ( dur.getTimeUnit() )
            {
                case NANOSECONDS: return TimeUnitImpl.NANOSECOND;
                case MILLISECONDS: return TimeUnitImpl.MILLISECOND;
                case SECONDS: return TimeUnitImpl.SECOND;
                case MINUTES: return TimeUnitImpl.MINUTE;
                case HOURS: return TimeUnitImpl.HOUR;
                case DAYS: return TimeUnitImpl.DAY;

                default: 
                    throw state.createFail( 
                        "Unexpected java.util.concurrent.TimeUnit:", 
                        dur.getTimeUnit()
                    );
            }
        }

        long sizeOf( Duration dur ) { return dur.getDuration(); }

        Duration
        createMetric( long sz,
                      TimeUnitImpl unit )
        {
            switch ( unit )
            {
                case NANOSECOND: return Duration.fromNanos( sz );
                case MILLISECOND: return Duration.fromMillis( sz );
                case SECOND: return Duration.fromSeconds( sz );
                case MINUTE: return Duration.fromMinutes( sz );
                case HOUR: return Duration.fromHours( sz );
                case DAY: return Duration.fromDays( sz );
                case FORTNIGHT: return Duration.fromFortnights( sz );

                default: throw state.createFail( "Unexpected unit:", unit );
            }
        }

        Duration
        implFromString( CharSequence cs )
        {
            return Duration.fromString( cs );
        }

        Duration
        implFromNum( long sz ) 
        {
            return Duration.fromMillis( sz );
        }
    }

    private
    final
    static
    class BindingsLoader
    implements MingleBinders.Initializer
    {
        public
        void
        initialize( MingleBinder.Builder b,
                    TypeDefinitionLookup types )
        {
            b.addBinding(
                QNAME_DATA_UNIT, 
                MingleBinders.asMingleBinding(
                    MingleModels.createExchanger( 
                        TYPE_DATA_UNIT, DataUnit.class ) ),
                DataUnit.class 
            );

            b.addBinding( 
                QNAME_DATA_SIZE, new DataSizeBinding(), DataSize.class );
            
            b.addBinding(
                QNAME_DURATION, new DurationBinding(), Duration.class );
        }
    }
}
