require 'bitgirder/core'
require 'bitgirder/testing'
require 'bitgirder/ops/java'
require 'bitgirder/io'
require 'mingle'

module BitGirder
module Ops
module Java

class JavaOpsTests

    include TestClassMixin
    include Mingle
    include BitGirder::Io

    def jv_home
        @jv_home ||= JavaEnvironments.get_java_home
    end

    def jv_env
        @jv_env ||= JavaEnvironment.new( jv_home )
    end

    # Mainly of value as basic code coverage of get_default()
    def test_jv_env_get_default

        assert_equal( 
            jv_home.sub( %r{/$}, "" ), 
            JavaEnvironment.get_default.java_home.sub( %r{/$}, "" )
        )
    end

    def run_forked_jv_env_test
        
        begin
            ENV.delete( "JAVA_HOME" )
            which_dir = File.dirname( which( "which" ) ) 

            Dir.mktmpdir do |tmp|
                ENV[ "PATH" ] = "#{tmp}/bin:/more/blah:#{which_dir}"
                fu().touch( java = ensure_parent( "#{tmp}/bin/java" ) )
                fu().chmod( 0755, java )
                assert_equal( tmp, JavaEnvironments.get_java_home )
            end

            exit! 0
        rescue => e
            code( "Forked #{__method__} failed: #{e} (#{e.class})" )
            exit! -1
        end
    end

    # Because this runs concurrently with other tests and intentionally sets
    # PATH to a value that would cause other tests to fail, we run the actual
    # meat of the test in a forked process
    def test_jv_env_get_default_uses_path_inference
        
        return if RubyVersions.jruby?

        if pid = fork

            opts = { :name => "forked #{__method__}", :pid => pid }
            pid, stat = debug_wait2( opts )
            assert( stat.success? )
        else
            run_forked_jv_env_test
        end 
    end

    def create_jrunner1
            
        JavaRunner.new(
            :java_env => jv_env,
            :command => "java",
            :classpath => %w{ cp-dir1 cp-dir2 },
            :jvm_args => %w{ -Xmx512m -XmaxBlah=ddd },
            :sys_props => { 
                "prop1" => "val1", 
                :prop2 => :val2,
                MingleIdentifier.get( :prop3 ) => MingleInt64.new( 3 )
            },
            :argv => [ "arg1", :arg2, 77 ],
            :proc_env => { "VAR1" => "val1" },
            :proc_opts => { :k => :v }
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
        assert_equal( { :k => :v }, b.opts )
    end

    def test_jv_env_as_classpath
        
        assert_equal( "a:b:c", jv_env.as_classpath( "a:b:c" ) )

        assert_equal( 
            "a:b:c", jv_env.as_classpath( MingleString.new( "a:b:c" ) ) )
        
        assert_equal( "a:b:c", jv_env.as_classpath( %w{ a b c } ) )

        assert_equal(
            "a:b:c", jv_env.as_classpath( MingleList.new( %w{ a b c } ) ) )
    end

    # Not re-testing all of what is tested in test_create_app_runner, just the
    # app_runner specific part
    def test_create_app_runner_custom_runner_class
        
        b = 
            JavaRunner.create_application_runner(
                :java_env => jv_env,
                :main => "com.bitgirder.runner.Runner",
                :classpath => "foo.jar",
                :argv => [ "arg1" ]
            ).
            process_builder

        assert_equal(
            %w{ -classpath foo.jar com.bitgirder.runner.Runner arg1 }, b.argv )
    end

    def test_jrunner_split_argv
        
        [
            [ %w{ foo }, { :argv => %w{ foo } } ],
            [ %w{ -Dk=v }, { :sys_props => { "k" => "v" } } ],
            [ %w{ -Xmx=2m }, { :jvm_args => %w{ -Xmx=2m } } ],

            # Current behavior is to accept questionable or malformed args and
            # let the jvm handle them
            [ %w{ -Dk1= -Dk2 -X -Xm= },
              { sys_props: { "k1" => "", "k2" => "" }, jvm_args: %w{ -X -Xm= } }
            ],

            [ %w{ foo -Dk1=v -Dk2=v2 -Xmx blah -Xmx=23 blah2 },
              { :argv => %w{ foo blah blah2 },
                :sys_props => { "k1" => "v", "k2" => "v2" },
                :jvm_args => %w{ -Xmx -Xmx=23 } }
            ]

        ].each do |pair|
            
            argv, expct = *pair
            act = JavaRunner.split_argv( argv )

            expct = { :argv => [], :sys_props => {}, :jvm_args => [] }.
                    merge( expct )

            assert_equal( expct, act )
        end
    end

    def test_jrunner_split_argv_fail
        
        assert_raised( "Property without name", RuntimeError ) do
            JavaRunner.split_argv( %w{ blah -D foo } )
        end
    end
end

end
end
end
