require 'bitgirder/core'
require 'bitgirder/testing'
require 'bitgirder/ops/java'
require 'mingle'

module BitGirder
module Ops
module Java

class JavaOpsTests

    include TestClassMixin
    include Mingle

    def jv_home
        @jv_home ||= JavaEnvironments.get_java_home
    end

    def jv_env
        @jv_env ||= JavaEnvironment.new( jv_home )
    end

    # Mainly of value as basic code coverage of get_default()
    def test_jv_env_get_default
        assert_equal( jv_home, JavaEnvironment.get_default.java_home )
    end

    def create_jrunner1
            
        JavaRunner.new(
            java_env: jv_env,
            command: "java",
            classpath: %w{ cp-dir1 cp-dir2 },
            jvm_args: %w{ -Xmx512m -XmaxBlah=ddd },
            sys_props: { 
                "prop1" => "val1", 
                prop2: :val2,
                MingleIdentifier.get( :prop3 ) => MingleInt64.new( 3 )
            },
            argv: [ "arg1", :arg2, 77 ],
            proc_env: { "VAR1" => "val1" },
            proc_opts: { k: :v }
        )
    end

    def assert_jrunner1_argv( b )
        
        work = Array.new( b.argv )
        assert_equal( "-classpath", work.shift )
        assert_equal( "cp-dir1:cp-dir2", work.shift )
        assert_equal( "-Xmx512m", work.shift )
        assert_equal( "-XmaxBlah=ddd", work.shift )

        # We don't use shift since we have no guarantees as to the ordering that
        # the props were added; we only care to assert that they all occur
        # exactly once and grouped together here at the front of the array; in
        # the block we check for inclusion in the front but call delete on the
        # actual array (not a slice)
        %w{ -Dprop1=val1 -Dprop2=val2 -Dprop3=3 }.each_with_index do |val, i| 
            assert( work[ 0 .. ( 2 - i ) ].include?( val ) )
            work.delete( val ) 
        end

        assert_equal( %w{ arg1 arg2 77 }, work ) # quick check of rest
    end

    def test_jrunner1_proc_builder
 
        b = create_jrunner1.process_builder
 
        assert_equal( "#{jv_home}/bin/java", b.cmd )

        assert_jrunner1_argv( b )

        assert_equal( { "VAR1" => "val1" }, b.env )
        assert_equal( { k: :v }, b.opts )
    end

    def test_jv_env_as_classpath
        
        assert_equal( "a:b:c", jv_env.as_classpath( "a:b:c" ) )

        assert_equal( 
            "a:b:c", jv_env.as_classpath( MingleString.new( "a:b:c" ) ) )
        
        assert_equal( "a:b:c", jv_env.as_classpath( %w{ a b c } ) )

        assert_equal(
            "a:b:c", jv_env.as_classpath( MingleList.new( %w{ a b c } ) ) )
    end

    def test_create_app_runner
        
        b =
            JavaRunner.create_application_runner(
                java_env: jv_env,
                app_class: "Foo",
                classpath: "foo.jar",
                jvm_args: [ "-Xblah" ],
                sys_props: { "p1" => "val1" },
                argv: %w{ arg1 arg2 },
                proc_env: { "VAR1" => "val1" },
                proc_opts: { k: :v }
            ).
            process_builder
        
        assert_equal( "#{jv_home}/bin/java", b.cmd )

        assert_equal(
            %w{ -classpath foo.jar -Xblah -Dp1=val1
                com.bitgirder.application.ApplicationRunner Foo arg1 arg2 },
            b.argv
        )

        assert_equal( { "VAR1" => "val1" }, b.env )
        assert_equal( { k: :v }, b.opts )
    end

    # Not re-testing all of what is tested in test_create_app_runner, just the
    # app_runner specific part
    def test_create_app_runner_custom_runner_class
        
        b = 
            JavaRunner.create_application_runner(
                java_env: jv_env,
                app_class: "Foo",
                app_runner: "com.bitgirder.runner.Runner",
                classpath: "foo.jar"
            ).
            process_builder

        assert_equal(
            %w{ -classpath foo.jar com.bitgirder.runner.Runner Foo }, b.argv )
    end

    def test_create_mingle_app_runner
        
        b =
            JavaRunner.create_mingle_app_runner(
                java_env: jv_env,
                app_class: "Foo",
                classpath: "foo.jar"
            ).
            process_builder
        
        argv_expct =
            %w{ -classpath foo.jar 
                com.bitgirder.mingle.application.MingleApplicationRunner Foo }

        assert_equal( argv_expct, b.argv )
    end
end

end
end
end
