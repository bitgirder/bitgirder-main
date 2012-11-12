package com.bitgirder.lang.reflect;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import java.lang.reflect.AnnotatedElement;
import java.lang.reflect.Constructor;
import java.lang.reflect.Method;
import java.lang.reflect.Field;
import java.lang.reflect.Modifier;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.InvocationHandler;
import java.lang.reflect.Type;
import java.lang.reflect.Member;
import java.lang.reflect.Array;
import java.lang.reflect.Proxy;

import java.lang.annotation.Annotation;

import java.util.List;
import java.util.Set;
import java.util.HashSet;
import java.util.Collection;
import java.util.LinkedList;
import java.util.Arrays;

import java.io.InputStream;
import java.io.IOException;

public
final
class ReflectUtils
{
    private static Inputs inputs = new Inputs();
    private static State state = new State();

    public final static Class< ? >[] EMPTY_CLASS_ARRAY = new Class< ? >[] {};
    public final static Object[] EMPTY_ARG_ARRAY = new Object[] {};

    private final static Annotation[][] EMPTY_ANNOTATIONS = 
        new Annotation[ 0 ][ 0 ];

    // Convenience method to call getResourceAsStream and assert that the result
    // is non-null
    public
    static
    InputStream
    getResourceAsStream( Class< ? > cls,
                         CharSequence rsrc )
    {
        inputs.notNull( cls, "cls" );
        inputs.notNull( rsrc, "rsrc" );

        InputStream res = cls.getResourceAsStream( rsrc.toString() );

        state.isFalse( 
            res == null, 
            "Couldn't find resource '" + rsrc + "' in", cls );

        return res;
    }

    public
    static
    void
    rethrow( InvocationTargetException ite )
        throws Exception
    {
        Throwable cause = state.notNull( ite.getCause() );
        
        // In the highly, highly unlikely event that Sun one day adds a new
        // subclass of Throwable, we'll at least know about it here.
        state.isTrue( cause instanceof Error || cause instanceof Exception );

        if ( cause instanceof Error ) throw (Error) cause;
        else throw (Exception) cause;
    }

    public
    static
    Object
    getStaticFieldValue( Class< ? > cls,
                         String fieldName )
        throws IllegalAccessException,
               NoSuchFieldException
    {
        inputs.notNull( cls, "cls" );
        inputs.notNull( fieldName, "fieldName" );

        Field f = cls.getDeclaredField( fieldName );
        f.setAccessible( true );

        return f.get( null );
    }

    // obj and args may be null, as specified in the underlying Method.invoke
    // docs
    public
    static
    Object
    invoke( Method m,
            Object obj,
            Object... args )
        throws Exception
    {
        inputs.notNull( m, "m" );
        
        Object res = null;

        try { res = m.invoke( obj, args ); }
        catch ( InvocationTargetException ite ) { rethrow( ite ); }

        return res;
    }

    public
    static
    < T >
    T
    invoke( Constructor< T > c,
            Object... args )
        throws Exception
    {
        inputs.notNull( c, "c" );

        T res = null;

        try { res = c.newInstance( args ); }
        catch ( InvocationTargetException ite ) { rethrow( ite ); }

        return res;
    }

    // paramTypes or params can be null
    public
    static
    < T >
    T
    newInstance( Class< T > cls,
                 Class< ? >[] paramTypes,
                 Object[] params )
        throws Exception
    {
        inputs.notNull( cls, "cls" );

        T res = null;

        Constructor< ? extends T > cons = 
            getDeclaredConstructor( cls, paramTypes );

        cons.setAccessible( true );

        try { res = cons.newInstance( params ); }
        catch ( InvocationTargetException ite ) { rethrow( ite ); }

        return res;
    }

    // Declared to throw Exception, since it could catch an
    // InvocationTargetException and throw the resulting target Exception as the
    // result of this method. Besides these, the normal reflection methods may
    // also be thrown as declared in the Class.getDeclaredConstructor and
    // Constructor.newInstance methods.
    //
    // The main difference between this method and cls.newInstance() is that
    // this will work for classes of all visibilities (private, package, private
    // static inners, etc)
    public
    static
    < T >
    T
    newInstance( Class< ? extends T > cls )
        throws Exception
    {
        inputs.notNull( cls, "cls" );
        return newInstance( cls, null, null );
    }

    // In addition to exceptions from {@link #newInstance newInstance}, may
    // throw {@link ClassNotFoundException}.
    public
    static
    Object
    newInstance( String clsName )
        throws Exception
    {
        inputs.notNull( clsName, "clsName" );
        return newInstance( Class.forName( clsName ) );
    }

    public
    static
    < T >
    T
    newInnerInstance( Object inst,
                      Class< T > innerCls )
        throws Exception
    {
        inputs.notNull( inst, "inst" );
        inputs.notNull( innerCls, "innerCls" );
        
        return newInstance( 
            innerCls,
            new Class[] { inst.getClass() }, 
            new Object[] { inst }
        );
    }

    private
    static
    boolean
    isModified( int mods,
                int expct )
    {
        return ( mods & expct ) > 0;
    }

    private
    static
    boolean
    isModified( Member m,
                int expct )
    {
        return isModified( inputs.notNull( m, "m" ).getModifiers(), expct );
    }

    private
    static
    boolean
    isModified( Class< ? > cls,
                int mods )
    {
        return isModified( inputs.notNull( cls, "cls" ).getModifiers(), mods );
    }

    // Convenience to allow callers to inline calls like 
    //
    //      ReflectUtils.isStatic( cls.getModifiers() );
    //
    // without having to import java.lang.reflect.Modifiers directly
    //
    public
    static
    boolean
    isStatic( int mods )
    {
        return isModified( mods, Modifier.STATIC );
    }

    public
    static
    boolean
    isStatic( Member m )
    {
        return isModified( m, Modifier.STATIC );
    }

    public
    static
    boolean
    isStatic( Class< ? > cls )
    {
        return isModified( cls, Modifier.STATIC );
    }

    public
    static
    boolean
    isAbstract( int mods )
    {
        return isModified( mods, Modifier.ABSTRACT );
    }

    public
    static
    boolean
    isAbstract( Member m )
    {
        return isModified( m, Modifier.ABSTRACT );
    }

    public
    static
    boolean 
    isAbstract( Class< ? > cls )
    {
        return isModified( cls, Modifier.ABSTRACT );
    }

    public
    static
    boolean
    hasPackageVisibility( int mods )
    {
        return ! ( Modifier.isPublic( mods ) ||
                   Modifier.isProtected( mods ) ||
                   Modifier.isPrivate( mods ) );
    }

    public
    static
    boolean
    hasPackageVisibility( Member m )
    {
        return hasPackageVisibility( inputs.notNull( m, "m" ).getModifiers() );
    }

    public
    static
    boolean
    hasPackageVisibility( Class< ? > cls )
    {
        return 
            hasPackageVisibility( inputs.notNull( cls, "cls" ).getModifiers() );
    }

    public
    static
    Method[]
    getStaticDeclaredMethods( Class< ? > cls )
    {
        inputs.notNull( cls, "cls" );

        List< Method > res = new LinkedList< Method >();

        Method[] methods = cls.getDeclaredMethods();
        for ( Method m : methods )
        {
            if ( isStatic( m ) )
            {
                m.setAccessible( true );
                res.add( m );
            }
        }

        return res.toArray( new Method[ res.size() ] );
    }

    // 'annCls' is null checked here to avoid repeating it, so we need to ensure
    // that public methods use the same param name 'annCls', or modify this
    // method to take the param name as an argument for inputs.notNull
    private
    static
    < A extends AnnotatedElement >
    Collection< A >
    filter( A[] elts,
            Class< ? extends Annotation > annCls )
    {
        inputs.notNull( annCls, "annCls" );

        Collection< A > res = new LinkedList< A >();

        for ( A elt : elts )
        {
            if ( elt.isAnnotationPresent( annCls ) ) res.add( elt );
        }

        return res;
    }

    public
    static
    Collection< Method >
    getMethods( Class< ? > cls,
                Class< ? extends Annotation > annCls )
    {
        return filter( inputs.notNull( cls, "cls" ).getMethods(), annCls );
    }

    public
    static
    Collection< Method >
    getDeclaredMethods( Class< ? > cls,
                        Class< ? extends Annotation > annCls )
    {
        return filter( 
            inputs.notNull( cls, "cls" ).getDeclaredMethods(), annCls );
    }  

    // returned collection is mutable for caller
    public
    static
    Collection< Method >
    getDeclaredAncestorMethods( Class< ? > cls,
                                Class< ? extends Annotation > annCls )
    {
        inputs.notNull( cls, "cls" );
        
        Collection< Method > res = new LinkedList< Method >();

        for ( Class< ? > ancestor : getAllAncestors( cls ) )
        {
            Collection< Method > mColl = getDeclaredMethods( ancestor, annCls );
            res.addAll( mColl );
        }

        return res;
    }

    public
    static
    Collection< Method >
    getStaticDeclaredMethods( Class< ? > cls,
                              Class< ? extends Annotation > annCls )
    {
        return filter( getStaticDeclaredMethods( cls ), annCls );
    }

    public
    static
    Collection< Field >
    getDeclaredFields( Class< ? > cls,
                       Class< ? extends Annotation > annCls )
    {
        return filter(
            inputs.notNull( cls, "cls" ).getDeclaredFields(), annCls );
    }

    public
    static
    Collection< Field >
    getDeclaredAncestorFields( Class< ? > cls,
                               Class< ? extends Annotation > annCls )
    {
        inputs.notNull( cls, "cls" );
        inputs.notNull( annCls, "annCls" );

        Collection< Field > res = Lang.newList();

        for ( Class< ? > anc : getAllAncestors( cls ) )
        {
            res.addAll( getDeclaredFields( anc, annCls ) );
        }

        return res;
    }

    public
    static
    Collection< Class< ? > >
    getDeclaredClasses( Class< ? > cls,
                        Class< ? extends Annotation > annCls )
    {
        return filter(
            inputs.notNull( cls, "cls" ).getDeclaredClasses(), annCls );
    }

    public
    static
    Collection< Class< ? > >
    getDeclaredAncestorClasses( Class< ? > cls,
                                Class< ? extends Annotation > annCls )
    {
        inputs.notNull( cls, "cls" );
        inputs.notNull( annCls, "annCls" );

        Collection< Class< ? > > res = Lang.newList();

        for ( Class< ? > anc : getAllAncestors( cls ) )
        {
            res.addAll( getDeclaredClasses( anc, annCls ) );
        }

        return res;
    }

    // Returns all ancestors of cls, including cls itself, in order of ancestor
    // --> descendant (so cls itself is the last element in the list)
    //
    // The returned list is owned by the caller and may be modified
    public
    static
    List< Class< ? > >
    getAllAncestors( Class< ? > cls )
    {
        inputs.notNull( cls, "cls" );

        LinkedList< Class< ? > > res = new LinkedList< Class< ? > >();

        for ( Class< ? > cur = cls; cur != null; cur = cur.getSuperclass() )
        {
            res.addFirst( cur );
        }

        return res;
    }

    // Returns all distinct generic interfaces implemented by cls, in no
    // particular order and using the default notion of Type.equals() (which is
    // to say, no particular notion at all) to determine duplicates
    public
    static
    Collection< Type >
    getAllGenericInterfaces( Class< ? > cls )
    {
        inputs.notNull( cls, "cls" );

        Set< Type > res = new HashSet< Type >();

        for ( Class< ? > c : getAllAncestors( cls ) )
        {
            for ( Type t : c.getGenericInterfaces() ) res.add( t );
        }

        return res;
    }

    // Note: includes cls itself if cls is an interface
    public
    static
    Collection< Class< ? > >
    getAllInterfaces( Class< ? > cls )
    {
        inputs.notNull( cls, "cls" );

        Collection< Class< ? > > res = new HashSet< Class< ? > >();

        for ( Class< ? > c : getAllAncestors( cls ) )
        {
            for ( Class< ? > iFace : c.getInterfaces() ) res.add( iFace );
        }

        if ( cls.isInterface() ) res.add( cls );

        return res;
    }

    public
    static
    Method
    getDeclaredMethod( Class< ? > cls,
                       String methodName,
                       Class< ? >... args ) // args can be null
        throws NoSuchMethodException,
               IllegalAccessException
    {
        inputs.notNull( cls, "cls" );
        inputs.notNull( methodName, "methodName" );

        Method res = cls.getDeclaredMethod( methodName, args );
        res.setAccessible( true );

        return res;
    }

    public
    static
    < T >
    Constructor< T >
    getDeclaredConstructor( Class< T > cls,
                            Class< ? >... paramTypes ) // can be null
        throws NoSuchMethodException,
               IllegalAccessException
    {
        inputs.notNull( cls, "cls" );

        Constructor< T > res = cls.getDeclaredConstructor( paramTypes );
        res.setAccessible( true );

        return res;
    }

    public
    static
    < T >
    Constructor< T >
    getDeclaredConstructor( Class< T > cls )
        throws NoSuchMethodException,
               IllegalAccessException
    {
        return getDeclaredConstructor( cls, EMPTY_CLASS_ARRAY );
    }

    // null-checks annCls on behalf of public frontends
    private
    static
    < A extends Annotation >
    A[]
    getParameterAnnotations( Annotation[][] anns,
                             Class< A > annCls )
    {
        inputs.notNull( annCls, "annCls" );

        Object arr = Array.newInstance( annCls, anns.length );

        for ( int i = 0, e = anns.length; i < e; ++i )
        {
            for ( Annotation ann : anns[ i ] )
            {
                if ( annCls.isInstance( ann ) )
                {
                    Array.set( arr, i, ann );
                    break;
                }
            }
        }

        @SuppressWarnings( "unchecked" )
        A[] res = (A[]) arr;

        return res;
    }

    public
    static
    < A extends Annotation >
    A[]
    getParameterAnnotations( Method m,
                             Class< A > annCls )
    {
        return 
            getParameterAnnotations( 
                inputs.notNull( m, "m" ).getParameterAnnotations(), annCls );
    }

    // Workaround for http://bugs.sun.com/bugdatabase/view_bug.do?bug_id=6962494
    public
    static
    Annotation[][]
    getParameterAnnotations( Constructor c )
    {
        inputs.notNull( c, "c" );

        Class< ? > cls = c.getDeclaringClass();

        if ( c.getParameterTypes().length == 1 &&
             cls.getEnclosingClass() != null && 
             ( ! isStatic( cls ) ) )
        {
            return EMPTY_ANNOTATIONS;
        }
        else return c.getParameterAnnotations();
    }

    // Placeholder in case we end up finding some reason to supply our own
    // implementation (such as was done with getParameterAnnotations(
    // Constructor ) in this class). All BG code should use this method rather
    // than directly calling m.getParameterAnnotations()
    public
    static
    Annotation[][]
    getParameterAnnotations( Method m )
    {
        return inputs.notNull( m, "m" ).getParameterAnnotations();
    }

    public
    static
    < A extends Annotation >
    A[]
    getParameterAnnotations( Constructor c,
                             Class< A > annCls )
    {
        // null check on c implicit in this call
        Annotation[][] anns = getParameterAnnotations( c );

        return getParameterAnnotations( anns, annCls );
    }

    private
    static
    boolean
    canCall( Class< ? >[] sigTyps,
             Class< ? >[] argTyps )
    {
        if ( sigTyps.length == argTyps.length )
        {
            for ( int i = 0, e = sigTyps.length; i < e; ++i )
            {
                if ( ! sigTyps[ i ].isAssignableFrom( argTyps[ i ] ) )
                {
                    return false;
                }
            }

            return true;
        }
        else return false;
    }

    private
    static
    void
    assertCanCall( boolean passed,
                   Object elt,
                   Class< ? >[] argTyps )
    {
        if ( ! passed )
        {
            throw state.createFail(
                "Cannot call", elt, "with argument types:",
                Arrays.toString( argTyps )
            );
        }
    }

    public
    static
    boolean
    canCall( Method m,
             Class< ? >... argTyps )
    {
        inputs.notNull( m, "m" );
        inputs.noneNull( argTyps, "argTyps" );

        return canCall( m.getParameterTypes(), argTyps );
    }

    public
    static
    void
    assertCanCall( Method m,
                   Class< ? >... argTyps )
    {
        assertCanCall( canCall( m, argTyps ), m, argTyps );
    }

    public
    static
    boolean
    canCall( Constructor< ? > c,
             Class< ? >... argTyps )
    {
        inputs.notNull( c, "c" );
        inputs.noneNull( argTyps, "argTyps" );

        return canCall( c.getParameterTypes(), argTyps );
    }

    public
    static
    void
    assertCanCall( Constructor< ? > c,
                   Class< ? >... argTyps )
    {
        assertCanCall( canCall( c, argTyps ), c, argTyps );
    }

    private
    static
    final
    class CallLoopHandler
    implements InvocationHandler
    {
        private final List< ? > impls;

        private CallLoopHandler( List< ? > impls ) { this.impls = impls; }
                        
        public
        Object
        invoke( Object proxy,
                Method m,
                Object[] args )
            throws Throwable
        {
            Object res = null;

            for ( Object impl : impls ) 
            {
                res = ReflectUtils.invoke( m, impl, args );
            }

            return res;
        }
    }

    public
    static
    < V >
    V
    createCallLoopProxy( Class< V > cls,
                         List< ? extends V > impls )
    {
        inputs.notNull( cls, "cls" );
        inputs.isFalse( impls.isEmpty(), "impls is empty" );
        List< ? > copy = Lang.unmodifiableCopy( impls, "impls" );

        return
            cls.cast(
                Proxy.newProxyInstance(
                    cls.getClassLoader(),
                    new Class< ? >[] { cls },
                    new CallLoopHandler( copy )
                )
            );
    }
}
