require 'bitgirder/core'
require 'bitgirder/io'
require 'bitgirder/ops/build'
require 'bitgirder/ops/build/distro'

require 'erb'
require 'set'

include BitGirder::Core
include BitGirder::Ops::Build

module BitGirder
module Ops
module Build
module Ruby

ALL_MODS = [ :lib, :test, :bin, :integ ]

ENV_PATH = "PATH"
ENV_BITGIRDER_DEBUG = "BITGIRDER_DEBUG"
ENV_BITGIRDER_TEST_RUBY = "BITGIRDER_TEST_RUBY"

TEST_DEBUG_DEFAULT = "true"

include RubyEnvVarNames

# Meant to be mixed in to something responding to ws_ctx, ruby_ctx. This is true
# of all TaskExecutors as well as of TestRunner
module RubyTaskMethods

    require 'set'

    extend BitGirderMethods

    def get_rb_incl_dirs( chain )
        
        res = Set.new

        BuildChains.tasks( chain, RubyModBuilder ).each do |t|
            res += t.get_incl_dirs
        end

        res.to_a
    end

    def self.get_test_dep_targets( dep, ws_ctx )

        dep_def = ws_ctx.proj_def( proj: dep )[ :test ] || {}
 
        ( dep_def[ :dep_targets ] || [] ).map { |str| TaskTarget.parse( str ) }
    end

    def get_test_dep_targets( dep )
        RubyTaskMethods.get_test_dep_targets( dep, ws_ctx )
    end
end

class TestRunner < BitGirderClass
    
    include RubyTaskMethods
    include BitGirder::Io

    require 'set'

    bg_attr :chain
    bg_attr :test_projs
    bg_attr :task
    bg_attr :run_opts
    bg_attr :run_ctx
    bg_attr :run_log, required: false
    bg_attr :test_mod
    bg_attr :ruby_ctx
    bg_attr :ws_ctx, required: false
    bg_attr :proc_env, default: {}

    private
    def impl_initialize

        super

        @ws_ctx ||= @task.ws_ctx
    end

    private
    def testing_proj
        ws_ctx.proj_dir( proj: :testing )
    end

    private
    def get_test_runner
        "#{testing_proj}/bin/unit-test-runner"
    end

    private
    def get_test_runner_incl_dirs
        [ "#{ws_ctx.mod_dir( proj: :core, mod: :lib )}",
          "#{testing_proj}/lib",
        ]
    end

    private
    def input_selectors( proj )
            
        proj_def = ws_ctx.proj_def( proj: proj )
        proj_def[ :"#{test_mod}_selectors" ] || [ "test_*" ]
    end

    private
    def get_test_inputs
 
        test_projs.inject( [] ) do |arr, proj|

            input_selectors( proj ).each do |glob|
 
                proj_dir = ws_ctx.proj_dir( proj: proj )

                Dir.glob( "#{proj_dir}/#{test_mod}/#{glob}" ).
                    reject { |f| f.end_with?( ".orig" ) }.
                    each { |f| arr << f }
            end

            arr
        end
    end

    private
    def get_filter_args
        
        if filt = run_opts.get_string( :filter_pattern )
            [ "--filter-pattern", filt ]
        else
            []
        end
    end

    private
    def get_run_argv( chain )
        
        res = []

        incls = get_rb_incl_dirs( chain ) + get_test_runner_incl_dirs
        res += Set.new( incls ).map { |d| [ "-I", d ] }.flatten

        res << get_test_runner

        res += get_filter_args
        res += ( run_ctx[ :argv_remain ] || [] )
        res += get_test_inputs

        res
    end

    private
    def get_proc_opts
        
        if @run_opts.get_boolean( :log_to_console )
            {}
        else
 
            unless run_log = @run_log
                build_dir = ws_ctx.mod_build_dir( mod: test_mod )
                run_log = ensure_parent( "#{build_dir}/log/run.log" )
            end

            console( "Sending test output to #{run_log}" )
            { [ :out, :err ] => [ run_log, "w" ] }
        end
    end

    private
    def get_run_path
        
        elts = []

        BuildChains.tasks( @chain ) do |tsk|
            
            if tsk.target.to_s =~ %r{go/build-bin/[^/]+/test} 
                elts.unshift( tsk.build_bin_dir( :test ) )
            end
        end

        elts.join( ":" )
    end

    private
    def update_run_env( opts )
 
        env = opts[ :env ]

        env.merge!( proc_env )

        unless env.key?( ENV_BITGIRDER_DEBUG )
            env[ ENV_BITGIRDER_DEBUG ] = TEST_DEBUG_DEFAULT
        end

        env[ ENV_PATH ] = "#{get_run_path}:#{env[ ENV_PATH ]}"
        env[ TestData::ENV_PATH ] = TestData.get_test_data_path( chain )
        env[ ENV_BITGIRDER_TEST_RUBY ] = opts[ :cmd ]
    end

    private
    def get_builder_opts

        res = ruby_ctx.proc_builder_opts( "ruby" )
        
        res[ :argv ].push( *( get_run_argv( chain ) ) )
        res[ :opts ].merge!( get_proc_opts )
        update_run_env( res )
        res[ :show_env_in_debug ] = true
        
        res
    end

    public
    def run

        if @run_opts.get_boolean( :test_dry_run )
            code( "Dry run; not running tests" )
        else
            UnixProcessBuilder.new( get_builder_opts ).system
        end
    end

    def self.run( *argv ); self.new( *argv ).run; end

    def self.get_integ_env( chain, path )
        
        rt_dir = BuildChains.expect_task( chain, path ).runtime_dir
        { "BITGIRDER_INTEG_RUNTIME" => rt_dir }
    end
end

class RubyModTask < StandardModTask
    
    include RubyTaskMethods
end

class RubyModBuilder < RubyModTask

    public
    def get_incl_dirs
        File.exist?( f = ws_ctx.mod_dir ) ? [ f ] : []
    end

    public
    def get_doc_src_files
        Dir.glob( "#{ws_ctx.mod_dir}/**/*.rb" )
    end

    public
    def get_direct_dependencies
        
        res = 
            direct_deps().inject( [] ) do |arr, dep| 
                arr << TaskTarget.create( :ruby, :build, dep, mod() )
            end
 
        get_mods.each do |mod|
            if ws_ctx.has_mod_src?( mod: mod )
                res << TaskTarget.create( :ruby, :build, proj(), mod )
            end
        end

        res
    end

    public
    def execute( chain )
    end
end

TaskRegistry.instance.register_path( RubyModBuilder, :ruby, :build )

class RubyProjTask < StandardProjTask
    
    include RubyTaskMethods
end

class RubyTestRunner < RubyProjTask

    public
    def get_direct_dependencies

        res = [ TaskTarget.create( :ruby, :build, proj(), :test ) ] +
              get_test_dep_targets( proj() )

        direct_deps().each do |dep| 
            res << TaskTarget.create( :ruby, :build, dep, :test )
            res += get_test_dep_targets( dep )
        end

        res
    end

    public
    def execute( chain )
 
        TestRunner.new( 
            chain: chain, 
            test_projs: [ proj() ],
            task: self,
            run_opts: run_opts(),
            run_ctx: run_ctx(),
            ruby_ctx: ruby_ctx,
            test_mod: :test
        ).
        run
    end
end

TaskRegistry.instance.register_path( RubyTestRunner, :ruby, :test )

class RubyIntegTester < RubyProjTask

    public
    def get_direct_dependencies

        res = [] 
        res << TaskTarget.create( :ruby, :build, :testing_integ, :lib )
        res << TaskTarget.create( :integ, :write, proj ) 
    end

    public
    def execute( chain )
 
        TestRunner.new( 
            chain: chain, 
            test_projs: [ proj() ],
            task: self,
            run_opts: run_opts(),
            run_ctx: run_ctx(),
            ruby_ctx: ruby_ctx,
            proc_env: 
                TestRunner.get_integ_env( chain, [ :integ, :write, proj ] ),
            test_mod: :integ
        ).
        run
    end
end

TaskRegistry.instance.register_path( RubyIntegTester, :ruby, :integ )

class SelfCheckRun < BitGirderClass
 
    include RubyTaskMethods

    bg_attr :ws_ctx
    bg_attr :ruby_ctx
    bg_attr :chain
    bg_attr :run_log, required: false

    def self.get_direct_dependencies
        
        [ :core, :testing ].map do |proj| 
            TaskTarget.create( :ruby, :build, proj, :test )
        end
    end

    private
    def get_mod_dir( mod )

        file_exists( 
            @ws_ctx.mod_dir( proj: :testing, code_type: :ruby, mod: mod ) )
    end

    private
    def set_run_log( opts )
        
        if @run_log
            
            code( "Self check run logging to #@run_log" )
            opts[ [ :out, :err ] ] = [ ensure_parent( @run_log ), "w" ]
        end
    end

    public
    def run
        
        opts = @ruby_ctx.proc_builder_opts( "ruby" )
 
        argv = opts[ :argv ]
        set_run_log( opts[ :opts ] )

        argv.push( RubyIncludes.as_include_argv( get_rb_incl_dirs( @chain ) ) )

        test_dir, bin_dir = get_mod_dir( :test ), get_mod_dir( :bin )

        argv << file_exists( "#{bin_dir}/unit-test-runner" )

        tr_tests = file_exists( "#{test_dir}/test_runner_tests.rb" )
        argv << tr_tests

        argv << "--reporter" << "BitGirder::Testing::TestRunnerAssertReporter"
        argv << "--reporter-require" <<  tr_tests
        argv << "-v"
        
        UnixProcessBuilder.new( opts ).system
    end

    def self.run( *argv ); self.new( *argv ).run; end
end

class RubyTestRunnerSelfCheck < TaskExecutor

    include RubyTaskMethods

    public
    def get_direct_dependencies
        SelfCheckRun.get_direct_dependencies
    end

    public
    def execute( chain )
        
        SelfCheckRun.new(
            ws_ctx: ws_ctx,
            ruby_ctx: ruby_ctx,
            run_log: run_log,
            chain: chain
        ).
        run
    end
end

TaskRegistry.instance.register_path( 
    RubyTestRunnerSelfCheck, :ruby, :test_runner, :self_check )

class RubyCommandRunner < RubyProjTask

    public
    def get_direct_dependencies
        
        mods = [ :lib ]
        mods << :test if @run_opts.get_boolean( :include_test )

        mods.map { |mod| TaskTarget.create( :ruby, :build, proj(), mod ) }
    end

    public
    def execute( chain )
        
        opts = ruby_ctx.proc_builder_opts( "ruby" )

        argv = get_rb_incl_dirs( chain ).map { |d| [ "-I", d ] }.flatten
        argv += argv_remain

        opts[ :argv ].push( *argv )

        UnixProcessBuilder.new( opts ).exec
    end
end

TaskRegistry.instance.register_path( RubyCommandRunner, :ruby, :run_command )

# Helper class for doc source gen: instances of this class are created in a
# child process which can freely require everything in @reqs in order to load
# all necessary subclasses of BitGirderClass, over which it reflects to generate
# the ruby source used for rdoc
class RubyDocGenerator < BitGirderClass
    
    bg_attr :out
    bg_attr :reqs
    bg_attr :mask

    private
    def templ( b, str )
        ERB.new( str ).result( b )
    end

    private
    def open_gen
        
        @out.print <<-END

# Autogenerated docs on #{Time.now}
#

# This code is only included for rdoc purposes and would not normally get
# required. Even so, in case of overzealous scripts which might auto-require
# this file, we enclose the entire body of this file in a guard that will
# prevent it from being interpreted in an actual run
#
if false

        END
    end

    private
    def attrs_by_name( cd )
        cd.attrs.sort.map { |pair| pair[ 1 ] }
    end

    private
    def gen_attr_desc_str( attr )
        
        ( attr.description || "" ).
            split( /\r?\n/ ).
            map { |s| s.strip }.
            join( " " )
    end

    private
    def gen_attr_desc( attr )
        "# #{gen_attr_desc_str( attr )}"
    end

    private
    def gen_attrs( cd )

        templ binding, <<-END

<% cd.attrs.each_pair do |nm, attr| %>
<%= gen_attr_desc( attr ) %>
<%= attr.mutable ? "attr_accessor" : "attr_reader" %> :<%= attr.identifier %>
<% end %>

        END
    end

    private
    def gen_initializer( cd )

        templ binding, <<-END

# Default constructor which takes a hash containing the following attributes:
# <% attrs_by_name( cd ).each do |attr| %>
# :+<%= attr.identifier %>+ ::
#   <%= gen_attr_desc_str( attr ) %>
# <% end %>
def initialize( opts = {} )
    # Autogenerated stub for docs
end

        END
    end

    private
    def write_class_def( cls, cd )

        @out.print( templ( binding, <<-END

class #{cls.to_s}

    <%= gen_attrs( cd ) %>

    <%= gen_initializer( cd ) %>
end

        END
        ))
    end

    private
    def write_class_defs
        
        BitGirderClassDefinition.get_class_defs.each_pair do |cls, cd|
            write_class_def( cls, cd ) if @mask.match( cls.to_s )
        end
    end

    private
    def close_gen

        @out.print <<-END

end # 'if false...' block

        END
    end

    public
    def run

        @reqs.each { |req| require req }

        open_gen
        write_class_defs
        close_gen
    end
end

class GenDocSrc < RubyModTask
    
    public
    def get_direct_dependencies
        
        res = [ TaskTarget.create( :ruby, :build, proj, mod ) ]

        direct_deps.each do |dep|
            res << TaskTarget.create( :ruby, :build, dep, mod() )
            res << TaskTarget.create( :ruby, :gen_doc_src, dep, mod() )
        end

        res
    end

    public
    def get_doc_gen_mask
        
        if lib_opts = ws_ctx.proj_def[ :lib ]
            if mask = lib_opts[ :doc_gen_mask ]
                Regexp.new( mask.to_s )
            end
        end
    end

    public
    def doc_gen_dir
        "#{ws_ctx.proj_build_dir}/doc-gen-#{mod}"
    end

    public
    def doc_gen_file
        "#{doc_gen_dir}/generated_rdoc.rb"
    end

    private
    def write_doc_synth( io, mask )
        
        reqs = Dir.glob( "#{ws_ctx.mod_dir( mod: :lib )}/**/*.rb" )
        RubyDocGenerator.new( out: io, reqs: reqs, mask: mask ).run
    end

    private
    def run_doc_synth_child
        
        stat = 0

        begin
            File.open( ensure_parent( doc_gen_file ), "w" ) do |io| 
                if mask = get_doc_gen_mask
                    write_doc_synth( io, mask )
                end # else: just leave empty file 
            end
        rescue Exception => e
            warn( e, "Doc gen failed" )
            stat = -1
        end

        stat
    end

    public
    def execute( chain )
        
        if pid = Kernel.fork
            wait_opts = { pid: pid, name: :"doc gen", check_status: true }
            _, stat = debug_wait2( wait_opts )
        else
            $:.unshift( *( get_rb_incl_dirs( chain ) ) )
            exit! run_doc_synth_child
        end
    end
end

TaskRegistry.instance.register_path( GenDocSrc, :ruby, :gen_doc_src )

class RubyDocBuilder < RubyProjTask
    
    public
    def get_direct_dependencies
        
        [ 
            TaskTarget.create( :ruby, :build, proj, :lib ),
            TaskTarget.create( :ruby, :gen_doc_src, proj, :lib ),
        ]
    end

    private
    def get_doc_src( chain )
        
        BuildChains.tasks( chain, RubyModBuilder, GenDocSrc ).
                    inject( [] ) do |arr, tsk|
            
            if tsk.mod == mg_id( :lib )

                case tsk

                    when RubyModBuilder 
                        arr.push( *( tsk.get_doc_src_files ) )

                    when GenDocSrc then arr << tsk.doc_gen_file
                end
            end

            arr
        end
    end 

    public
    def execute( chain )

        src_files = get_doc_src( chain )
        
        opts_base = ruby_ctx.proc_builder_opts( "rdoc" )
        
        fu().rm_rf( dest_dir = "#{ws_ctx.proj_build_dir}/doc" )
        opts_base[ :argv ].push( "--output", dest_dir )

        [ [], [ "--ri" ] ].each do |arr|
            
            opts = opts_base.clone
            opts[ :argv ].push( *( arr + src_files ) )
            
            UnixProcessBuilder.new( opts ).system
        end
    end
end

TaskRegistry.instance.register_path( RubyDocBuilder, :ruby, :build_doc )

class RubyDistTask < BitGirder::Ops::Build::Distro::AbstractDistroTask
end

class RubyBuildDist < RubyDistTask
    
    public
    def get_direct_dependencies
        
        direct_deps.inject( [] ) do |arr, dep|

            ALL_MODS.each do |mod|
                if ws_ctx.has_mod_src?( proj: dep, mod: mod )
                    arr << TaskTarget.create( :ruby, :build, dep, mod )
                end
            end

            arr
        end
    end

    public; def execute( chain ); end
end

TaskRegistry.instance.register_path( RubyBuildDist, :ruby, :dist, :build )

class DistTestRun < BitGirderClass
    
    bg_attr :task
    bg_attr :ruby_ctx
    bg_attr :chain

    # Include test-runner self-check as part of dist test to reduce risk of
    # false-positives in which tests seem to pass but test-runner itself is not
    # operating correctly. Of course it could be that :self_check itself also
    # passes erroneously, but we'd have to have that happen in combination with
    # the test framework also failing with false positives to not get a failure
    # here.
    def self.get_direct_dependencies( task )

        res = [ TaskTarget.create( :ruby, :dist, :build, task.dist ) ]
        res += SelfCheckRun.get_direct_dependencies

        task.direct_deps.each do |dep| 
            res += RubyTaskMethods.get_test_dep_targets( dep, task.ws_ctx )
        end

        res
    end

    private
    def opt_run_log( rel )

        unless @task.run_opts.get_boolean( :log_to_console )

            build_dir = task.ws_ctx.mod_build_dir( mod: :test )
            "#{build_dir}/#{@ruby_ctx.id}/log/#{rel}"
        end
    end

    public
    def run
 
        SelfCheckRun.run(
            ws_ctx: task.ws_ctx, 
            ruby_ctx: @ruby_ctx, 
            run_log: opt_run_log( "test-runner-check.log" ),
            chain: @chain
        )

        TestRunner.run(
            chain: @chain,
            test_projs: @task.direct_deps,
            task: @task,
            run_opts: task.run_opts(),
            run_ctx: task.run_ctx(),
            run_log: opt_run_log( "dist-test.log" ),
            ruby_ctx: @ruby_ctx,
            test_mod: :test,
        )
    end

    def self.run( *argv ); self.new( *argv ).run; end
end

class RubyDistTest < RubyDistTask
    
    include RubyTaskMethods

    public
    def get_direct_dependencies
        DistTestRun.get_direct_dependencies( self )
    end

    public
    def execute( chain )
        DistTestRun.run( task: self, ruby_ctx: ruby_ctx, chain: chain )
    end
end

TaskRegistry.instance.register_path( RubyDistTest, :ruby, :dist, :test )

class RubyDistValidate < RubyDistTask

    public
    def get_direct_dependencies
        DistTestRun.get_direct_dependencies( self )
    end

    public
    def execute( chain )
        
        build_env.ruby_env.rubies.each_pair do |id, ruby_ctx|

            code( "Validating dist #{dist} with ruby #{id}" )
            DistTestRun.run( task: self, chain: chain, ruby_ctx: ruby_ctx )
        end
    end
end

TaskRegistry.instance.register_path( RubyDistValidate, :ruby, :dist, :validate )

class RubyDistGem < RubyDistTask

    private
    def impl_initialize

        super
        @doc_gen_seq = 0
    end

    private
    def spec_config_file
        "#{ws_ctx.proj_dir( code_type: :ruby )}/spec-config.yaml"
    end

    public
    def get_direct_dependencies

        res = [ TaskTarget.create( :ruby, :dist, :build, dist ) ]

        direct_deps.each do |proj| 
            res << TaskTarget.create( :ruby, :gen_doc_src, proj, :lib )
        end

        res
    end

    private
    def link_tree( src_dir, dest_dir )

        Dir.chdir( src_dir ) do
            Dir.glob( "**/*" ).
                select { |f| File.file?( f ) }.
                each do |f|
                    dest_file = "#{dest_dir}/#{f}"
                    fu().ln_s( "#{src_dir}/#{f}", ensure_parent( dest_file ) )
                end
        end
    end

    private
    def link_ruby_mod_builder( tsk, gem_work )
        
        if tsk.ws_ctx.has_mod_src? && ( mod = tsk.mod ) == MOD_LIB
            link_tree( tsk.ws_ctx.mod_dir, "#{gem_work}/#{mod}" ) 
        end
    end

    private
    def link_gen_doc_src( tsk, gem_work )
        
        dest = "#{gem_work}/lib/doc-gen#@doc_gen_seq.rb"
        fu.cp( tsk.doc_gen_file, dest )

        @doc_gen_seq += 1
    end

    private
    def build_gem_lib( gem_work, chain )
        
        BuildChains.tasks( chain ).each do |tsk|

            case tsk
                when RubyModBuilder then link_ruby_mod_builder( tsk, gem_work )
                when GenDocSrc then link_gen_doc_src( tsk, gem_work )
            end
        end
    end

    private
    def collect_bin_dirs( chain )
        
        res = {}

        BuildChains.tasks( chain, RubyModBuilder ).each do |tsk|
            
            bin_dir = ws_ctx.mod_dir( proj: tsk.proj, mod: :bin )

            if File.exist?( bin_dir )
                res[ tsk.proj ] = bin_dir
            end
        end

        res
    end

    private
    def build_gem_bin( gem_work, chain )
        
        bin_dirs = collect_bin_dirs( chain )

        lnk = TreeLinker.new( dest: "#{gem_work}/bin" )

        bin_dirs.each_pair do |proj, dir|
            
            pd = ws_ctx.proj_def( proj: proj )

            if msk = pd.fields.get_string( :publish_bin )
                lnk.update_from( src: dir, selector: msk )
            end
        end

        lnk.build
    end

    private
    def add_bg_license( gem_work )
        
        File.open( "#{gem_work}/LICENSE.txt", "w" ) do |io|
            io.print APACHE2_LICENSE
        end
    end

    private
    def build_gem_files( gem_work, chain )

        build_gem_lib( gem_work, chain )
        build_gem_bin( gem_work, chain )
        add_bg_license( gem_work )
    end

    private
    def get_gem_executables
        Dir.chdir( "bin" ) { Dir.glob( "*" ) }
    end

    # Called in gem work dir
    private
    def create_gem_spec
 
        cfg = load_yaml( spec_config_file )
        spec = Gem::Specification.new

        [ :summary, :description ].each do |attr|
            spec.send( :"#{attr}=", has_key( cfg, attr ) )
        end

        spec.name = "bitgirder-#{dist}"
        spec.version = BuildVersions.get_version( run_opts: @run_opts )
        spec.authors = ["BitGirder Technologies, Inc"]
        spec.email = "dev-support@bitgirder.com"
        spec.homepage = "http://www.bitgirder.com"
        spec.licenses = ["BitGirder Master Software License"]
        spec.require_paths = ["lib"]
        spec.executables.push( *( get_gem_executables ) )

        spec.files = Dir.glob( "**/*" )

        spec
    end

    public
    def execute( chain )

        gem_work = ensure_wiped( "#{dist_build_dir}/gem-work" )
        build_gem_files( gem_work, chain )
        
        @built_gem = Dir.chdir( gem_work ) do 
            spec = create_gem_spec
            File.expand_path( Gem::Builder.new( spec ).build )
        end

        code( "Built: #@built_gem" )
    end
end

TaskRegistry.instance.register_path( RubyDistGem, :ruby, :dist, :gem )

APACHE2_LICENSE = <<END_LIC

                                 Apache License
                           Version 2.0, January 2004
                        http://www.apache.org/licenses/

   TERMS AND CONDITIONS FOR USE, REPRODUCTION, AND DISTRIBUTION

   1. Definitions.

      "License" shall mean the terms and conditions for use, reproduction,
      and distribution as defined by Sections 1 through 9 of this document.

      "Licensor" shall mean the copyright owner or entity authorized by
      the copyright owner that is granting the License.

      "Legal Entity" shall mean the union of the acting entity and all
      other entities that control, are controlled by, or are under common
      control with that entity. For the purposes of this definition,
      "control" means (i) the power, direct or indirect, to cause the
      direction or management of such entity, whether by contract or
      otherwise, or (ii) ownership of fifty percent (50%) or more of the
      outstanding shares, or (iii) beneficial ownership of such entity.

      "You" (or "Your") shall mean an individual or Legal Entity
      exercising permissions granted by this License.

      "Source" form shall mean the preferred form for making modifications,
      including but not limited to software source code, documentation
      source, and configuration files.

      "Object" form shall mean any form resulting from mechanical
      transformation or translation of a Source form, including but
      not limited to compiled object code, generated documentation,
      and conversions to other media types.

      "Work" shall mean the work of authorship, whether in Source or
      Object form, made available under the License, as indicated by a
      copyright notice that is included in or attached to the work
      (an example is provided in the Appendix below).

      "Derivative Works" shall mean any work, whether in Source or Object
      form, that is based on (or derived from) the Work and for which the
      editorial revisions, annotations, elaborations, or other modifications
      represent, as a whole, an original work of authorship. For the purposes
      of this License, Derivative Works shall not include works that remain
      separable from, or merely link (or bind by name) to the interfaces of,
      the Work and Derivative Works thereof.

      "Contribution" shall mean any work of authorship, including
      the original version of the Work and any modifications or additions
      to that Work or Derivative Works thereof, that is intentionally
      submitted to Licensor for inclusion in the Work by the copyright owner
      or by an individual or Legal Entity authorized to submit on behalf of
      the copyright owner. For the purposes of this definition, "submitted"
      means any form of electronic, verbal, or written communication sent
      to the Licensor or its representatives, including but not limited to
      communication on electronic mailing lists, source code control systems,
      and issue tracking systems that are managed by, or on behalf of, the
      Licensor for the purpose of discussing and improving the Work, but
      excluding communication that is conspicuously marked or otherwise
      designated in writing by the copyright owner as "Not a Contribution."

      "Contributor" shall mean Licensor and any individual or Legal Entity
      on behalf of whom a Contribution has been received by Licensor and
      subsequently incorporated within the Work.

   2. Grant of Copyright License. Subject to the terms and conditions of
      this License, each Contributor hereby grants to You a perpetual,
      worldwide, non-exclusive, no-charge, royalty-free, irrevocable
      copyright license to reproduce, prepare Derivative Works of,
      publicly display, publicly perform, sublicense, and distribute the
      Work and such Derivative Works in Source or Object form.

   3. Grant of Patent License. Subject to the terms and conditions of
      this License, each Contributor hereby grants to You a perpetual,
      worldwide, non-exclusive, no-charge, royalty-free, irrevocable
      (except as stated in this section) patent license to make, have made,
      use, offer to sell, sell, import, and otherwise transfer the Work,
      where such license applies only to those patent claims licensable
      by such Contributor that are necessarily infringed by their
      Contribution(s) alone or by combination of their Contribution(s)
      with the Work to which such Contribution(s) was submitted. If You
      institute patent litigation against any entity (including a
      cross-claim or counterclaim in a lawsuit) alleging that the Work
      or a Contribution incorporated within the Work constitutes direct
      or contributory patent infringement, then any patent licenses
      granted to You under this License for that Work shall terminate
      as of the date such litigation is filed.

   4. Redistribution. You may reproduce and distribute copies of the
      Work or Derivative Works thereof in any medium, with or without
      modifications, and in Source or Object form, provided that You
      meet the following conditions:

      (a) You must give any other recipients of the Work or
          Derivative Works a copy of this License; and

      (b) You must cause any modified files to carry prominent notices
          stating that You changed the files; and

      (c) You must retain, in the Source form of any Derivative Works
          that You distribute, all copyright, patent, trademark, and
          attribution notices from the Source form of the Work,
          excluding those notices that do not pertain to any part of
          the Derivative Works; and

      (d) If the Work includes a "NOTICE" text file as part of its
          distribution, then any Derivative Works that You distribute must
          include a readable copy of the attribution notices contained
          within such NOTICE file, excluding those notices that do not
          pertain to any part of the Derivative Works, in at least one
          of the following places: within a NOTICE text file distributed
          as part of the Derivative Works; within the Source form or
          documentation, if provided along with the Derivative Works; or,
          within a display generated by the Derivative Works, if and
          wherever such third-party notices normally appear. The contents
          of the NOTICE file are for informational purposes only and
          do not modify the License. You may add Your own attribution
          notices within Derivative Works that You distribute, alongside
          or as an addendum to the NOTICE text from the Work, provided
          that such additional attribution notices cannot be construed
          as modifying the License.

      You may add Your own copyright statement to Your modifications and
      may provide additional or different license terms and conditions
      for use, reproduction, or distribution of Your modifications, or
      for any such Derivative Works as a whole, provided Your use,
      reproduction, and distribution of the Work otherwise complies with
      the conditions stated in this License.

   5. Submission of Contributions. Unless You explicitly state otherwise,
      any Contribution intentionally submitted for inclusion in the Work
      by You to the Licensor shall be under the terms and conditions of
      this License, without any additional terms or conditions.
      Notwithstanding the above, nothing herein shall supersede or modify
      the terms of any separate license agreement you may have executed
      with Licensor regarding such Contributions.

   6. Trademarks. This License does not grant permission to use the trade
      names, trademarks, service marks, or product names of the Licensor,
      except as required for reasonable and customary use in describing the
      origin of the Work and reproducing the content of the NOTICE file.

   7. Disclaimer of Warranty. Unless required by applicable law or
      agreed to in writing, Licensor provides the Work (and each
      Contributor provides its Contributions) on an "AS IS" BASIS,
      WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
      implied, including, without limitation, any warranties or conditions
      of TITLE, NON-INFRINGEMENT, MERCHANTABILITY, or FITNESS FOR A
      PARTICULAR PURPOSE. You are solely responsible for determining the
      appropriateness of using or redistributing the Work and assume any
      risks associated with Your exercise of permissions under this License.

   8. Limitation of Liability. In no event and under no legal theory,
      whether in tort (including negligence), contract, or otherwise,
      unless required by applicable law (such as deliberate and grossly
      negligent acts) or agreed to in writing, shall any Contributor be
      liable to You for damages, including any direct, indirect, special,
      incidental, or consequential damages of any character arising as a
      result of this License or out of the use or inability to use the
      Work (including but not limited to damages for loss of goodwill,
      work stoppage, computer failure or malfunction, or any and all
      other commercial damages or losses), even if such Contributor
      has been advised of the possibility of such damages.

   9. Accepting Warranty or Additional Liability. While redistributing
      the Work or Derivative Works thereof, You may choose to offer,
      and charge a fee for, acceptance of support, warranty, indemnity,
      or other liability obligations and/or rights consistent with this
      License. However, in accepting such obligations, You may act only
      on Your own behalf and on Your sole responsibility, not on behalf
      of any other Contributor, and only if You agree to indemnify,
      defend, and hold each Contributor harmless for any liability
      incurred by, or claims asserted against, such Contributor by reason
      of your accepting any such warranty or additional liability.

END_LIC

end
end
end
end
