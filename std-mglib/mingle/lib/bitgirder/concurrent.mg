@version: v1;

namespace bitgirder:concurrent
{
    enum TimeUnit 
    { 
        nanosecond,
        millisecond,
        second,
        minute,
        hour,
        day,
        fortnight;
    }

    struct Duration
    {
        duration: Int64~[0,);
        unit: TimeUnit;

        @constructor( String );
        @constructor( Int64~[0,) );
    }
}
