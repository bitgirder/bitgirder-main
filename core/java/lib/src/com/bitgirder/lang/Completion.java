package com.bitgirder.lang;

public
interface Completion< V >
{
    public
    boolean
    isOk();

    // getResult() should throw IllegalStateThrowable if isOk() would return
    // false; getThrowable() likewise if isOk() would return true

    // the result itself can be null (example: an instance of Completion< Void >
    // would return null when isOk() is true
    public
    V
    getResult();

    // if isOk() is false must return a non-null exception object
    public
    Throwable
    getThrowable();

    // should return getResult() if isOk() would return true; should throw
    // the result of getThrowable() if isOk() would return false
    //
    // For consistency's sake with most every other java method ever written
    // this is typed to throw Exception instead of Throwable. Implementations
    // should correctly cast Throwables which are Errors to that type before
    // throwing them.
    public
    V
    get()
        throws Exception;
}
