package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.path.ObjectPath;
import com.bitgirder.lang.path.ImmutableListPath;

import java.util.Iterator;

public
final
class MingleListIterator
implements Iterator< MingleValue >
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final ObjectPath< MingleIdentifier > parent;
    private final Iterator< MingleValue > it;

    private ImmutableListPath< MingleIdentifier > path;

    private
    MingleListIterator( Iterator< MingleValue > it,
                        ObjectPath< MingleIdentifier > parent )
    {
        this.it = it;
        this.parent = parent;
    }

    public
    ObjectPath< MingleIdentifier >
    getPath()
    {
        state.isFalse( 
            path == null, "No path set (iteration has not started)" );
        
        return path;
    }

    public boolean hasNext() { return it.hasNext(); }
    public void remove() { throw new UnsupportedOperationException(); }

    // Invariant: path is always associated with the element returned from
    // the most recent call to this method (null if this method hasn't been
    // called yet)
    public
    MingleValue
    next()
    {
        // do the access first in case we get NoSuchElementException
        MingleValue res = it.next();

        path = path == null ? parent.startImmutableList() : path.next();

        return res;
    }

    private
    < V extends MingleValue >
    V
    nextTyped( MingleTypeReference typ,
               Class< V > cls )
    {
        MingleValue v = next();

        return 
            cls.cast( MingleModels.asMingleInstance( typ, v, path ) );
    }

    public
    MingleNull
    nextMingleNull()
    {
        return nextTyped( MingleModels.TYPE_REF_MINGLE_NULL, MingleNull.class );
    }

    public
    MingleString
    nextMingleString()
    {
        return 
            nextTyped( 
                MingleModels.TYPE_REF_MINGLE_STRING, MingleString.class );
    }

    public
    MingleInt64
    nextMingleInt64()
    {
        return 
            nextTyped( MingleModels.TYPE_REF_MINGLE_INT64, MingleInt64.class );
    }

    public
    MingleInt32
    nextMingleInt32()
    {
        return 
            nextTyped( MingleModels.TYPE_REF_MINGLE_INT32, MingleInt32.class );
    }

    public
    MingleDouble
    nextMingleDouble()
    {
        return 
            nextTyped( 
                MingleModels.TYPE_REF_MINGLE_DOUBLE, MingleDouble.class );
    }

    public
    MingleFloat
    nextMingleFloat()
    {
        return 
            nextTyped( MingleModels.TYPE_REF_MINGLE_FLOAT, MingleFloat.class );
    }

    public
    MingleBoolean
    nextMingleBoolean()
    {
        return 
            nextTyped( 
                MingleModels.TYPE_REF_MINGLE_BOOLEAN, MingleBoolean.class );
    }

    public
    MingleBuffer
    nextMingleBuffer()
    {
        return 
            nextTyped( 
                MingleModels.TYPE_REF_MINGLE_BUFFER, MingleBuffer.class );
    }

    public
    MingleTimestamp
    nextMingleTimestamp()
    {
        return 
            nextTyped( 
                MingleModels.TYPE_REF_MINGLE_TIMESTAMP, MingleTimestamp.class );
    }

    public
    MingleList
    nextMingleList()
    {
        return 
            nextTyped( 
                MingleModels.TYPE_REF_MINGLE_VALUE_LIST, MingleList.class );
    }

    public
    MingleEnum
    nextMingleEnum()
    {
        return nextTyped( MingleModels.TYPE_REF_MINGLE_ENUM, MingleEnum.class );
    }

    public
    MingleStruct
    nextMingleStruct()
    {
        return 
            nextTyped( 
                MingleModels.TYPE_REF_MINGLE_STRUCT, MingleStruct.class );
    }

    public
    MingleSymbolMap
    nextMingleSymbolMap()
    {
        return 
            nextTyped( 
                MingleModels.TYPE_REF_MINGLE_SYMBOL_MAP, 
                MingleSymbolMap.class );
    }

    public
    MingleSymbolMapAccessor
    nextMingleSymbolMapAccessor()
    {
        return
            MingleSymbolMapAccessor.create( nextMingleSymbolMap(), getPath() );
    }

    public
    static
    MingleListIterator
    forList( MingleList l,
             ObjectPath< MingleIdentifier > parent )
    {
        inputs.notNull( l, "l" );
        inputs.notNull( parent, "parent" );

        return new MingleListIterator( l.iterator(), parent );
    }
}
