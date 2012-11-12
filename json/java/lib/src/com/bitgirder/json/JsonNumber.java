package com.bitgirder.json;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

public
final
class JsonNumber
implements JsonValue
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final Number num;

    private JsonNumber( Number num ) { this.num = num; }

    public Number getNumber() { return num; }
    public byte byteValue() { return num.byteValue(); }
    public short shortValue() { return num.shortValue(); }
    public int intValue() { return num.intValue(); }
    public long longValue() { return num.longValue(); }
    public float floatValue() { return num.floatValue(); }
    public double doubleValue() { return num.doubleValue(); }

    // Can add getNumber() also for access to the num object itself

    public
    static
    JsonNumber
    forNumber( Number num )
    {
        inputs.notNull( num, "num" );
        return new JsonNumber( num );
    }

    @Override public String toString() { return num.toString(); }
}
