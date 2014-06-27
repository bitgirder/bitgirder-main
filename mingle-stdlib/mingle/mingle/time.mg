@version v1

namespace mingle:time

# Not adding units like minutes/hours/days, since the calculations and
# conversions can get a little involved. 
enum TimeUnit { 
    nanosecond,
    microsecond,
    millisecond,
    second,
}

# Use this rather than bare Uint64 when storing amounts of time, since just
# passing around Uint64 can lead to confusion (is it nanos? seconds? millis?)
struct Duration {
    
    unit TimeUnit
    size Uint64

    @constructor( String )
}
