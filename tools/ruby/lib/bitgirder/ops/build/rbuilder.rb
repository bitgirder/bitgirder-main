require 'bitgirder/core'
require 'bitgirder/io'
require 'bitgirder/ops/build'
require 'bitgirder/ops/build/distro'

require 'erb'

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

class RubyDistInteg < RubyDistTask
    
    public
    def get_direct_dependencies
        [ TaskTarget.create( :integ, :dist, :build, dist ) ]
    end

    public
    def execute( chain )

        env = TestRunner.get_integ_env( chain, [ :integ, :dist, :build, dist ] )

        TestRunner.new(
            chain: chain, 
            test_projs: direct_deps,
            task: self,
            run_opts: run_opts(),
            run_ctx: run_ctx(),
            ruby_ctx: ruby_ctx,
            proc_env: env,
            test_mod: :integ
        ).
        run
    end
end

TaskRegistry.instance.register_path( RubyDistInteg, :ruby, :dist, :integ )

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
    def add_bg_license( gem_work )
        
        File.open( "#{gem_work}/LICENSE.txt", "w" ) do |io|
            io.print MASTER_LICENSE
        end
    end

    private
    def build_gem_files( gem_work, chain )

        build_gem_lib( gem_work, chain )
        add_bg_license( gem_work )
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
        spec.version = "0.1.0"
        spec.authors = ["BitGirder Technologies, Inc"]
        spec.email = "dev-support@bitgirder.com"
        spec.homepage = "http://www.bitgirder.com"
        spec.licenses = ["BitGirder Master Software License"]
        spec.require_paths = ["lib"]

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

MASTER_LICENSE = <<END_LIC
THE FOLLOWING TERMS AND CONDITIONS CONTROL THE USE OF THE SOFTWARE PROVIDED BY
BITGIRDER AND CONTAIN SIGNIFICANT RESTRICTIONS AND LIMITATIONS ON RIGHTS AND
REMEDIES, AND CREATE OBLIGATIONS ON ANYONE WHO ACCEPTS THIS AGREEMENT.
THEREFORE, YOU SHOULD READ THIS AGREEMENT CAREFULLY.

BITGIRDER TECHNOLOGIES, INC. SOFTWARE LICENSE AGREEMENT:

BY USING THE BITGIRDER SOFTWARE (THE "SOFTWARE"), YOU AGREE TO THE FOLLOWING
TERMS AND CONDITIONS WHICH CONSTITUTE A LEGALLY ENFORCEABLE LICENSE AGREEMENT
(THE "AGREEMENT"). IF YOU ARE ENTERING INTO THIS AGREEMENT ON BEHALF OF A
COMPANY OR OTHER LEGAL ENTITY, YOU REPRESENT THAT YOU HAVE THE COMPLETE
AUTHORITY TO ENTER INTO THIS AGREEMENT ON BEHALF OF YOUR COMPANY.  IF YOU ARE
USING THE SOFTWARE AS AN INDIVIDUAL, YOU REPRESENT THAT YOU ARE OVER THE AGE OF
18. AS USED IN THIS LICENSE AGREEMENT, THE TERM "CUSTOMER" ENCOMPASSES THE
ENTITY AND/OR EACH USER OF THE SOFTWARE, INCLUDING, IF YOU ARE A CORPORATE
ENTITY, ALL EMPLOYEES OF YOUR COMPANY.  IF YOU DO NOT HAVE THE REQUISITE
AUTHORITY, OR IF YOU DO NOT AGREE WITH THESE TERMS AND CONDITIONS, YOU MAY NOT
USE THIS SOFTWARE.

Your registration for, or use of, the Software shall be deemed to be your
agreement to abide by this Agreement and any materials available on the
BitGirder website or provided to you by BitGirder which are incorporated by
reference herein (including but not limited to any privacy or security
policies).  For reference, a Definitions section is included at the end of this
Agreement.

1. Grant of Rights; Restrictions Pursuant to the terms and conditions of this
Agreement, BitGirder hereby grants Customer a limited, non-exclusive,
non-transferable, worldwide right to use the Software solely for Customer's own
internal business purposes (the "License"). 

Customer may copy, reverse engineer or modify the Software for its own internal
purposes, but shall not (unless indicated in the applicable Order Form) license,
grant, sell, resell, transfer, assign, or distribute to any third party the
Software or any derivative works thereof in any way.  

BitGirder and its licensors reserve all rights not expressly granted to
Customer.
  
2. The Software BitGirder is providing Customer with the Software in the edition
selected by Customer in the applicable Order Form.  

3. Support BitGirder has no obligation under this Agreement to provide any
support to Customer unless specifically contracted for by Customer.  Additional
service and support options are available, may be selected on the applicable
Order Form(s) and are governed by a Professional Service and Support Agreement
(the "Service Agreement") available from BitGirder.

4. Customer's Responsibilities Customer shall abide by all applicable local,
state, national and foreign laws, treaties and regulations in connection with
Customer's use of the Software, including those related to data privacy,
international communications and the transmission of technical or personal data.
Customer is solely responsible for protecting any of its password and
authentication information.  Customer shall report to BitGirder immediately and
use reasonable efforts to stop immediately, any unauthorized copying or
distribution of Software or Content that is known or suspected by Customer or
any User. Customer further (a) understands that the Software contains
information which is considered a trade secret of BitGirder, (b) will preserve
as confidential all trade secrets, confidential knowledge, data or other
proprietary information relating to the Software and (c) will take all
reasonable and necessary precautions to protect such information.

5. Intellectual Property Ownership BitGirder alone (and its licensors, where
applicable) shall own all right, title and interest, including all related
Intellectual Property Rights, in and to the BitGirder Technology, the Content,
and the Software and any suggestions, ideas, enhancement requests, feedback,
recommendations or other information provided by Customer or any other party
relating to the Software.  The BitGirder name, the BitGirder logo, and the
product names associated with the Software are trademarks of BitGirder or third
parties, and no right or license is granted to use them.   Customer acknowledges
that, except as specifically provided under this Agreement, no other right,
title, or interest in the Software, Content or BitGirder Technology is granted.

6. Representations & Warranties Each party represents and warrants that it has
the legal power and authority to enter into this Agreement.  Customer represents
and warrants that Customer has not falsely identified itself to gain access to
the Software and that, if applicable, Customer's billing information is correct. 

Disclaimer of Warranties UNLESS OTHERWISE SET FORTH ON THE APPLICABLE ORDER
FORM, BITGIRDER AND ITS LICENSORS MAKE NO REPRESENTATION, WARRANTY, OR GUARANTY
AS TO THE RELIABILITY, TIMELINESS, QUALITY, SUITABILITY, TRUTH, AVAILABILITY,
ACCURACY OR COMPLETENESS OF THE SOFTWARE OR ANY CONTENT.  BITGIRDER AND ITS
LICENSORS DO NOT REPRESENT OR WARRANT THAT (A) THE SOFTWARE WILL BE ERROR-FREE
OR OPERATE IN COMBINATION WITH ANY OTHER HARDWARE, SOFTWARE, SYSTEM OR DATA, (B)
THE SOFTWARE WILL MEET CUSTOMER'S REQUIREMENTS OR EXPECTATIONS, (C) ANY STORED
DATA WILL BE ACCURATE OR RELIABLE, (D) THE QUALITY OF ANY PRODUCTS, SOFTWARE,
INFORMATION, OR OTHER MATERIAL PURCHASED OR OBTAINED BY CUSTOMER THROUGH THE
SOFTWARE WILL MEET CUSTOMER'S REQUIREMENTS OR EXPECTATIONS, OR (E) ERRORS OR
DEFECTS WILL BE CORRECTED.  THE SOFTWARE AND ALL CONTENT IS PROVIDED TO CUSTOMER
STRICTLY ON AN "AS IS" BASIS.  ALL CONDITIONS, REPRESENTATIONS AND WARRANTIES,
WHETHER EXPRESS, IMPLIED, STATUTORY OR OTHERWISE, INCLUDING, WITHOUT LIMITATION,
ANY IMPLIED WARRANTY OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, OR
NON-INFRINGEMENT OF THIRD PARTY RIGHTS, ARE HEREBY DISCLAIMED TO THE MAXIMUM
EXTENT PERMITTED BY APPLICABLE LAW BY BITGIRDER AND ITS LICENSORS. CUSTOMER
ACKNOWLEDGES THAT SOFTWARE IS OF AN EXPERIMENTAL NATURE.

7. Limitation of Liability CUSTOMER ASSUMES ALL RISKS IN CONNECTION WITH ITS USE
OF THE SOFTWARE. IN NO EVENT SHALL BITGIRDER AND/OR ITS LICENSORS BE LIABLE TO
ANYONE FOR ANY INDIRECT, PUNITIVE, SPECIAL, EXEMPLARY, INCIDENTAL, CONSEQUENTIAL
OR OTHER DAMAGES OF ANY TYPE OR KIND (INCLUDING LOSS OF DATA, REVENUE, PROFITS,
USE OR OTHER ECONOMIC ADVANTAGE) ARISING OUT OF, OR IN ANY WAY CONNECTED WITH
THIS SOFTWARE, INCLUDING BUT NOT LIMITED TO THE USE OR INABILITY TO USE THE
SOFTWARE, OR FOR ANY CONTENT OBTAINED FROM OR THROUGH THE SOFTWARE, INACCURACY,
ERROR OR OMISSION, REGARDLESS OF CAUSE IN THE CONTENT, EVEN IF THE PARTY FROM
WHICH DAMAGES ARE BEING SOUGHT OR SUCH PARTY'S LICENSORS HAVE BEEN PREVIOUSLY
ADVISED OF THE POSSIBILITY OF SUCH DAMAGES. 

8. Additional Rights Certain states and/or jurisdictions do not allow the
disclaimer of warranties or limitation of liability, so the exclusions set forth
above may not apply to Customer. 

9. Mutual Indemnification Customer and every User under this Agreement shall
indemnify and hold BitGirder, its licensors and their parent organizations,
subsidiaries, affiliates, officers, directors, employees, attorneys and agents
harmless from and against any and all claims, causes of action, costs, damages,
losses, liabilities and expenses (including attorneys' fees and costs) arising
out of or in connection with: (i) violation by Customer of Customer's
representations and warranties or (ii) the breach by Customer or any User
pursuant to this Agreement, provided in any such case, that BitGirder (a) gives
written notice of the claim promptly to Customer; (b) gives Customer sole
control of the defense and settlement of the claim (except Customer may not
settle any claim, without BitGirder's consent, unless Customer unconditionally
releases BitGirder of all liability and such settlement does not affect
BitGirder's business or Software,); (c) provides to Customer all available
information and assistance; and (d) has not compromised or settled such claim.

BitGirder shall indemnify and hold Customer and Customer's authorized Users,
parent organizations, subsidiaries, affiliates, officers, directors, employees,
attorneys and agents harmless from and against any and all claims, causes of
action, costs, damages, losses, liabilities and expenses (including attorneys'
fees and costs) arising out of or in connection with: (i) an allegation that the
Software directly infringes a copyright, a U.S. patent issued as of the date
Customer accepts this Agreement, or a trademark of a third party; (ii) a
violation by BitGirder of its representations or warranties; or (iii) breach of
this Agreement by BitGirder; provided in any such case, that Customer (a)
promptly gives written notice of the claim to BitGirder; (b) gives BitGirder
sole control of the defense and settlement of the claim (except BitGirder may
not settle any claim, without Customer's consent, unless it unconditionally
releases Customer of all liability); (c) provides to BitGirder all available
information and assistance; and (d) has not compromised or settled such claim.  

BitGirder shall have no indemnification obligation, and Customer shall indemnify
BitGirder pursuant to this Agreement, for claims arising from any infringement
alleged to be caused by the combination of the Software with any of Customer's
products, software, and hardware or business process. 

10. Local Laws and Export Control The Software uses technology that may be
subject to United States export controls administered by the U.S. Department of
Commerce, the United States Department of Treasury Office of Foreign Assets
Control, and other U.S. agencies and the export control regulations of the
European Union.  Customer and each User acknowledges and agrees that the
Software shall not be used, and none of the Software, Content or BitGirder
Technology may be transferred or otherwise exported or re-exported to countries
as to which the United States and/or the European Union maintains an embargo
(collectively, "Embargoed Countries"), or to or by a national or resident
thereof, or any person or entity on the U.S. Department of Treasury's List of
Specially Designated Nationals or the U.S. Department of Commerce's Table of
Denial Orders (collectively, "Designated Nationals").  The lists of Embargoed
Countries and Designated Nationals are subject to change without notice.  By
using the Software, Customer represents and warrants that Customer is not
located in, under the control of, or a national or resident of an Embargoed
Country or Designated National.  Customer agrees to comply strictly with all
U.S. and European Union export laws and assume sole responsibility for obtaining
any necessary licenses to export or re-export.

The Software provided on the site may use encryption technology that is subject
to licensing requirements under the U.S. Export Administration Regulations, 15
C.F.R. Parts 730-774 and Council Regulation (EC) No. 1334/2000.

BitGirder and its licensors make no representation that the Software is
appropriate or available for use in other locations.  If Customer uses the
Software from outside the United States of America and/or the European Union,
Customer is solely responsible for compliance with all applicable laws,
including without limitation export and import regulations of other countries.
Any diversion of the Content contrary to United States or European Union
(including European Union Member States) law is prohibited. None of the Software
or Content, nor any information acquired through the use of the Software, is or
will be used for nuclear activities, chemical or biological weapons or missile
projects, unless specifically authorized by the United States government or
appropriate European body for such purposes. 

11. Notice BitGirder may give notice by means of a general notice on the
Website, electronic mail to Customer's e-mail address on record in BitGirder's
account information, or by written communication sent by first class mail or
pre-paid post to Customer's address on record in BitGirder's account
information.  Such notice shall be deemed to have been given upon the expiration
of 48 hours after mailing or posting (if sent by first class mail or pre-paid
post) or 12 hours after sending (if sent by email) or posted on the Website.
Customer may give notice to BitGirder (such notice shall be deemed given when
received by BitGirder) at any time by any of the following: letter sent by
confirmed electronic mail to BitGirder at info@BitGirder.com; letter delivered
by nationally recognized overnight delivery service or first class postage
prepaid mail to BitGirder at the following address: BitGirder Technologies,
Inc., 530 Brannan Street, #103, San Francisco, CA 94107, addressed to the
attention of: Chief Technology Officer. 

12. Modification to Terms BitGirder reserves the right to modify the terms and
conditions of this Agreement or its policies relating to the Software at any
time, effective upon posting of an updated version of this Agreement on the
Website.  Customer is responsible for regularly reviewing this Agreement.
Continued use of the Software after any such changes shall constitute Customer's
consent to such changes.  

13. Assignment This Agreement may not be assigned without the prior written
approval of the parties, but may be assigned without any prior consent by either
party to (i) a parent or subsidiary, (ii) an acquirer of substantially all of
such assigning party's assets, or (iii) a successor by merger.  Any purported
assignment in violation of this section shall be void.

14. General This Agreement shall be governed by California law and controlling
United States federal law, without regard to the choice or conflicts of law
provisions of any jurisdiction, and any disputes, actions, claims or causes of
action arising out of or in connection with this Agreement or the Software shall
be subject to the exclusive jurisdiction of the state and federal courts located
in San Francisco County, California.  If any provision of this Agreement is held
by a court of competent jurisdiction to be invalid or unenforceable, then such
provision(s) shall be construed, as nearly as possible, to reflect the
intentions of the invalid or unenforceable provision(s), with all other
provisions remaining in full force and effect.  No joint venture, partnership,
employment, or agency relationship exists between Customer or any User and
BitGirder as a result of this Agreement or use of the Software.  The failure of
BitGirder to enforce any right or provision in this Agreement shall not
constitute a waiver of such right or provision unless acknowledged and agreed to
by BitGirder in writing.  This Agreement, together with any applicable Order
Form(s) and the Service Agreement, comprises the entire agreement between
Customer and BitGirder and supersedes all prior or contemporaneous negotiations,
discussions or agreements, whether written or oral, between the parties
regarding the subject matter contained herein.

15. Definitions As used in this Agreement and in any Order Forms now or
hereafter associated herewith: 

"Agreement" means these terms of use, the original Order Form, any subsequent
Order Forms, submitted by Customer, and any materials specifically incorporated
by reference herein, as such materials, including the terms of this Agreement,
may be updated by BitGirder from time to time in its sole discretion; 

"BitGirder" means collectively BitGirder Technologies, Inc., a Delaware
corporation, having its principal place of business at: 387 Dolores Street, San
Francisco, CA 94110, and any of its direct or indirect subsidiaries; 

"BitGirder Technology" means all of BitGirder's proprietary technology
(including software, hardware, products, business concepts, and processes, logic
algorithms, graphical User interfaces (GUI), techniques, designs and other
tangible or intangible technical material or information) made available to
Customer by BitGirder in providing the Software; 

"Content" means the audio and visual information, documents, software, products
and Software contained or made available to Customer and the User(s) authorized
to use the Software under this License in the course of using the Software; 

"Intellectual Property Rights" means all rights, title and interest in and to
the BitGirder Technology, the Content, and all copyrights, patents, trade
secrets, trademarks, Software marks or other intellectual property or
proprietary rights and any corrections, bug fixes, enhancements, updates,
releases, or other modifications, including custom modifications made by
BitGirder relating thereto, and the media on which same are furnished; 

"Order Form(s)" means the form evidencing the initial designation of Software
and any subsequent Order Forms, specifying, among other things, the edition of
the Software selected and covered by the License, the number of Users, the
applicable fees, and any applicable billing information, as agreed to between
BitGirder and Customer, each such Order Form to be incorporated into and to
become a part of this Agreement (in the event of any conflict between the terms
of this Agreement and the terms of any such Order Form, the terms of the
applicable Order Form(s) shall prevail);  

"User(s)" means Customer's employees, representatives, consultants, contractors
or agents who are authorized under the License made by this Agreement to use the
Software.

"Website" means http://www.BitGirder.com 

Questions or Additional Information: If you have questions regarding this User
Agreement or wish to obtain additional information, please send an e-mail to
info@BitGirder.com.

Copyright 2007 BitGirder Technologies, Inc. All rights reserved.
END_LIC

end
end
end
end
