package com.bitgirder.mingle.model;

import com.bitgirder.validation.Inputs;
import com.bitgirder.validation.State;

import com.bitgirder.lang.Lang;

import com.bitgirder.lang.path.ObjectPath;

import java.util.List;
import java.util.Set;

// This class is public, so an instance can be instantiated by token and syntax
// builder classes in other mingle packages, but has constructors which require
// a non-null instance of an object which can only be created by trusted parts
// of the implementation. This is done as a way to provide what the java
// language (currently) does not, which would be something like a friend method
// or a super-package, or in general any other way to allow code in the parser
// or compiler packages create instances of classes in this package without
// requiring that said classes in this package expose public initializers. We
// can give API developers that illusion by having the public classes in this
// package use only package-level initializers, and allowing code in other
// internal mingle packages, such as parser or compiler, to access them via
// delegation through an instance of this factory. Combined with leaving this
// class out of any public javadocs, we can keep the public API simple (no
// public initializers on public classes, etc).
//
// As a general rule, all input checking is done in the delegation target, and
// parameters are passed through by the methods in this class unchecked
public
final
class ImplFactory
{
    private final static Inputs inputs = new Inputs();
    private final static State state = new State();

    private final static Set< String > CLASS_NAMES =
        Lang.unmodifiableSet(
            Lang.newSet(
                "com.bitgirder.mingle.parser.MingleGrammars$FactoryAccessor",
                "com.bitgirder.mingle.parser.MingleParsers$FactoryAccessor",
                "com.bitgirder.mingle.service.MingleServices$FactoryAccessor",
                "com.bitgirder.mingle.model.bind.MingleBindingContext$" +
                    "FactoryAccessor"
            )
        );

    // Ensure that acc is not-null, so that a cast of null to FactoryAccessor
    // will not succeed, but otherwise ignore acc
    public 
    ImplFactory( Object acc )
    {
        inputs.notNull( acc, "acc" );

        inputs.isTrue(
            CLASS_NAMES.contains( acc.getClass().getName() ),
            "Unrecognized factory object:", acc
        );
    }

    public
    MingleIdentifier
    createMingleIdentifier( String[] parts )
    {
        return new MingleIdentifier( parts );
    }

    public
    MingleNamespace
    createMingleNamespace( MingleIdentifier[] ns,
                           MingleIdentifier ver )
    {
        return new MingleNamespace( ns, ver );
    }

    public
    MingleTypeName
    createMingleTypeName( String[] parts )
    {
        return new MingleTypeName( parts );
    }

    public
    MingleIdentifiedName
    createIdentifiedNameUnsafe( MingleNamespace ns,
                                MingleIdentifier[] ids )
    {
        return MingleIdentifiedName.createUnsafe( ns, ids );
    }

    public
    MingleRangeRestriction< ? >
    createRangeRestriction( boolean closedMin,
                            MingleValue min,
                            MingleValue max,
                            boolean closedMax,
                            Class< ? extends MingleValue > cls )
    {
        return 
            MingleRangeRestriction.
                create( closedMin, min, max, closedMax, cls );
    }
}
