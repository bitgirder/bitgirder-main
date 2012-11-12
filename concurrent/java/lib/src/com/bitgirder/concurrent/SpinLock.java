package com.bitgirder.concurrent;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import java.util.concurrent.locks.ReentrantLock;

public
final
class SpinLock
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    private final ReentrantLock lock = new ReentrantLock();

    // Later we may want to add two public constructors which mirror the
    // constructors of ReentrantLock (default and one with fairness); for now we
    // use just the default
    public SpinLock() {}

    public void lock() { while ( ! lock.tryLock() ); }
    public void unlock() { lock.unlock(); }

    public 
    boolean 
    isHeldByCurrentThread()
    {
        return lock.isHeldByCurrentThread();
    }
}
