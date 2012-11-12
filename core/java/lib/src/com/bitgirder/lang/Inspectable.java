package com.bitgirder.lang;

public
interface Inspectable
{
    // Return the inspector for chaining purposes, especially useful for
    // chaining invocations in subclasses:
    //
    //      public
    //      Inspector
    //      accept( Inspector i )
    //      {
    //          return super.accept( i ).add( "foo", foo ).add( "bar", bar );
    //      }
    //
    public
    Inspector
    accept( Inspector i );
}
