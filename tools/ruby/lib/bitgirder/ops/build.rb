require 'bitgirder/core'
require 'bitgirder/io'
require 'bitgirder/ops/ruby'
require 'mingle'

module BitGirder
module Ops
module Build

include BitGirder::Core
include Mingle

include BitGirder::Ops::Ruby
include RubyEnvVarNames

# This will probably move somewhere more general eventually
class TreeLinker < BitGirderClass
    
    bg_attr :dest

    bg_attr :realpath, :processor => :boolean, :default => true

    private
    def impl_initialize
        @links = {}
    end

    private
    def select_targets( sel )
        
        case sel
        when String then Dir.glob( sel )
        else raise "Unhandled selector type: #{sel.class}"
        end
    end

    # :src may not be nil, but okay if directory does not exist
    public
    def update_from( opts )
        
        src_dir = has_key( opts, :src )
        sel = opts[ :selector ] || "**/*"
        
        return unless File.exists?( src_dir )

        Dir.chdir( src_dir ) do
 
            select_targets( sel ).each do |f|

                targ = "#{src_dir}/#{f}"

                if prev = @links[ f ]
                    raise "Linking #{targ} into #{dest} would conflict " \
                          "with previously linked #{prev}"
                else
                    @links[ f ] = targ
                end
            end
        end
    end

    # Returns @dest on success
    public
    def build
        
        ensure_wiped( @dest )

        # Go in sorted order to simplify debugging (make runs deterministic)
        @links.keys.sort.each do |targ|
            
            src = @links[ targ ]
            src = File.realpath( src ) if @realpath

            dest = ensure_parent( "#@dest/#{targ}" )
            fu().ln_s( src, dest )
        end

        @dest
    end
end

module Constants
    
    MOD_LIB = Mingle::MingleIdentifier.get( :lib )
    MOD_TEST = Mingle::MingleIdentifier.get( :test )
    MOD_BIN = Mingle::MingleIdentifier.get( :bin )
    MOD_INTEG = Mingle::MingleIdentifier.get( :integ )
end

module BuildVersions

    def get_version( opts = {} )
        
        if ( run_opts = opts[ :run_opts ] ) &&
           ( ver = run_opts.get_string( :build_version ) )
            ver
        else
            cmd = which( "hg" )
            
            case tag = `hg identify -t`.chomp
            when "tip", nil, /^\s*$/
                raise "No build version given or inferred"
            else tag
            end
        end
    end

    module_function :get_version
end

module MingleCodecMixin
    
    @@bgm = BitGirder::Core::BitGirderMethods
    @@mgc = nil # Mingle::Codec
    @@json_codec = nil

    def get_json_codec
        
        unless @@json_codec
            
            require 'mingle/json'
            @@json_codec = Mingle::Json::JsonMingleCodec.new
        end

        @@json_codec
    end

    def get_mg_codec_mod
        
        unless @@mgc
            
            require 'mingle/codec'
            @@mgc = Mingle::Codec
        end

        @@mgc
    end

    def get_mg_codecs
        get_mg_codec_mod::MingleCodecs
    end

    def mg_encode( *argv )
        get_mg_codecs.encode( *argv )
    end

    def mg_decode( *argv )
        get_mg_codecs.decode( *argv )
    end

    def load_mingle_struct_file( f )

        @@bgm.not_nil( f, :f )

        File.open( f ) { |io| mg_decode( get_json_codec, io ) }
    end
end

class TaskTarget < BitGirderClass
    
    include Mingle, Comparable

    bg_attr :path,
            is_list: true,
            list_validation: :not_empty,
            processor: lambda { |id| MingleIdentifier.get( id ) }

    public
    def []( idx )
        @path[ idx ]
    end

    public
    def ==( other )
        other && ( @path == other.path )
    end

    alias eql? ==

    public
    def hash
        @path.hash
    end

    public
    def to_s
        @path.join( "/" )
    end

    public
    def <=>( o )
        return o.is_a?( TaskTarget ) ? to_s <=> o.to_s : nil
    end

    def self.create( *path )
        self.new( path: path )
    end

    def self.parse( str )
        
        BitGirder::Core::BitGirderMethods.not_nil( str, :str )
        
        if ( toks = str.split( /\// ) ).empty?
            raise "Empty target"
        else
            self.create( *toks )
        end
    end
end

class Workspace < BitGirderClass
    
    bg_attr :root
    bg_attr :build_dir
end

class WorkspaceContext < BitGirderClass
    
    bg_attr :task
    bg_attr :environment_file, required: false
    bg_attr :defaults, default: proc { {} }
 
    include BitGirder::Io
    include MingleCodecMixin

    def initialize( *argv )
        
        super( *argv )
        @def_cache = {}
    end

    public
    def workspace
        @task.workspace
    end

    private
    def expect_opts( h, *keys )

        keys.map do |k| 
            h[ k ] or @defaults[ k ] or raise "No value for key #{k.inspect}"
        end
    end

    private
    def expect_opt( h, key )
        expect_opts( h, key )[ 0 ]
    end

    private
    def get_opt( h, key )
        h[ key ] || @defaults[ key ]
    end

    public
    def proj_root( opts = {} )

        proj = expect_opt( opts, :proj )
        "#{workspace.root}/#{proj}"
    end

    public
    def proj_dir( opts = {} )
        
        ct = expect_opt( opts, :code_type )
        "#{proj_root( opts )}/#{ct}"
    end

    public
    def mod_dir( opts = {} )
        
        mod = expect_opt( opts, :mod )
        "#{proj_dir( opts )}/#{mod}"
    end

    public
    def proj_def_file( opts = {} )
        "#{proj_dir( opts )}/project.json"
    end

    public
    def has_project?( opts = {} )
        File.exist?( proj_def_file( opts ) )
    end

    public
    def proj_def( opts = {} )
 
        f = file_exists( proj_def_file( opts ) )

        if res = @def_cache[ f ]
            res
        else
            @def_cache[ f ] = load_mingle_struct_file( f )
        end
    end

    public
    def validate_dep( opts = {} )

        code_type, proj = expect_opts( opts, :code_type, :proj )

        err_suffix = opts[ :err_suffix ]

        unless File.exist?( f = proj_def_file( opts ) )

            msg = "#{code_type} project '#{proj}' doesn't exist" 
            ( msg << " " << err_suffix ) if err_suffix
            
            raise msg
        end
    end

    public
    def validate_deps( opts )
        
        deps = expect_opt( opts, :deps )

        deps.each { |dep| validate_dep( opts.merge( proj: dep ) ) }
    end

    public
    def build_info_root( opts = {} )
        
        dirs = [ "#{task.workspace.build_dir}/.build-info" ]

        if ct = get_opt( opts, :code_type ) then dirs << ct end
        if proj = get_opt( opts, :proj ) then dirs << proj end

        dirs.join( "/" )
    end

    public
    def code_type_build_dir( opts = {} )

        ct = expect_opt( opts, :code_type )
        "#{workspace.build_dir}/code/#{ct}"        
    end

    public
    def proj_build_dir( opts = {} )
        
        proj = expect_opt( opts, :proj )
        "#{code_type_build_dir( opts )}/#{proj}"
    end

    public
    def mod_build_dir( opts = {} )
        
        mod = expect_opt( opts, :mod )
        "#{proj_build_dir( opts )}/#{mod}"
    end

    public
    def has_mod_src?( opts = {} )
        
        src_dir = mod_dir( opts )

        if ( File.exists?( src_dir ) )

            files = Dir.glob( "#{src_dir}/**/*" ).select { |f| File.file?( f ) }
            files.size > 0
        else
            false
        end
    end
end

class BuildEnv < BitGirderClass
    
    bg_attr :ruby_env

    def self.from_mingle_struct( s )
        
        self.new(
            ruby_env: RubyEnv.from_mingle_struct( s[ :ruby_env ] ),
        )
    end
end

class TaskExecutor < BitGirder::Core::BitGirderClass

    include Mingle
    include BitGirder::Io
    include Constants

    include MingleCodecMixin

    bg_attr :target
    bg_attr :workspace
    bg_attr :env_config, required: false
    bg_attr :run_opts
    bg_attr :run_ctx

    private
    def mg_id( id )
        MingleIdentifier.get( id )
    end

    private
    def build_info_root
        "#{@workspace.build_dir}/.build-info"
    end

    private
    def build_info_dir
        "#{build_info_root}/#@target"
    end

    public
    def info_file
        "#{build_info_dir}/build-info.json"
    end

    private
    def argv_remain
        run_ctx[ :argv_remain ] || []
    end

    # Saves and returns info (retval is nice for chaining as retval of other
    # methods)
    private
    def save_build_info( info )
        
        not_nil( info, :info )

        File.open( ensure_parent( info_file ), "w" ) do |io| 
            io.puts( mg_encode( get_json_codec, info ) )
        end

        info
    end

    public
    def load_build_info( expct_file = false )
        
        if File.exist?( f = info_file )
            File.open( f ) { |io| mg_decode( get_json_codec, io ) }
        else
            raise "Build info file #{f} not found" if expct_file
        end
    end

    public
    def ws_ctx( opts = {} )

        not_nil( opts, :opts )
        WorkspaceContext.new( task: self, defaults: opts ) 
    end

    public
    def env_config
        if @env_config
            load_mingle_struct_file( @env_config )
        else
            MingleStruct.new( type: TYPE_ENV_CFG )
        end
    end

    public
    def build_env
        @build_env ||= BuildEnv.from_mingle_struct( env_config )
    end

    public
    def ruby_ctx
        build_env.ruby_env.get_context( @run_opts.get_string( :ruby_context ) )
    end
end

module FileSigMixin

    require 'digest/md5'

    @@bgm = BitGirder::Core::BitGirderMethods

    private
    def get_file_sig( opts )

        @@bgm.not_nil( opts, :opts )

        digest = opts[ :digest ] || Digest::MD5.new
        
        @@bgm.has_key( opts, :files ).sort.each do |f|
            File.open( f ) { |io| digest.update( io.read ) }
        end

        Mingle::MingleBuffer.new( digest.digest )
    end
end

module JavaEnvMixin
    
    ENV_JAVA_HOME = "JAVA_HOME"

    @@bgm = BitGirder::Core::BitGirderMethods

    private
    def java_home
        
        unless res = @run_opts.get_string( :java_home )
            res = ENV[ ENV_JAVA_HOME ]
        end

        if res
            @java_home = file_exists( File.expand_path( res ) )
        else
            if java = which( "java" )
                @java_home = File.expand_path( "#{java}/../.." )
            else
                raise "java-home run opt not provided and " \
                      "#{ENV_JAVA_HOME} is not set"
            end
        end
    end

    private
    def jcmd( cmd )

        @@bgm.not_nil( cmd, :cmd )
        file_exists( "#{java_home}/bin/#{cmd}" )
    end
end

module RubyEnvMixin
    
    extend BitGirder::Core::BitGirderMethods

    ENV_RUBY_HOME = "RUBY_HOME"
    ENV_GEM_HOME = "GEM_HOME"

    def impl_get_rb_env_var( nm, on_miss = nil )

        if @run_opts && ( res = @run_opts.get_string( nm ) )
            return res
        elsif res = ENV[ nm.to_s.upcase ]
            return res
        else
            return on_miss
        end
    end

    def ruby_home
        
        unless @ruby_home
            if @ruby_home = impl_get_rb_env_var( :ruby_home )
                BitGirder::Io.file_exists( @ruby_home )
            else
                f = BitGirder::Io.which( "ruby" ) or 
                        raise "No ruby in path and ruby home not explicitly set"
                
                @ruby_home = Pathname.new( "#{f}/../.." ).cleanpath.to_s
            end
        end

        @ruby_home
    end

    def gem_home
        @gem_home ||= impl_get_rb_env_var( :gem_home, "" )
    end

    def rcmd( cmd, alts = true )
        
        not_nil( cmd, :cmd )

        cmds = [ cmd ]
        cmds << "j#{cmd}" if alts

        bin_dir = "#{ruby_home}/bin"

        cmds.map { |f| "#{bin_dir}/#{f}" }.find { |f| File.exists?( f ) } or
            raise "No #{cmd} in #{bin_dir}"
    end
end

class StandardTask < TaskExecutor

    # Could be overridden
    def code_type
        target.path[ 0 ]
    end

    def ws_ctx( opts = {} )

        not_nil( opts, :opts )
        super( opts.merge( code_type: code_type ) )
    end
end

class StandardProjTask < StandardTask
 
    def proj
        target.path[ 2 ] or raise "No project set"
    end

    def ws_ctx( opts = {} )

        not_nil( opts, :opts )
        super( opts.merge( proj: proj ) )
    end

    def direct_deps( validate_deps = true )
 
        res = ws_ctx.proj_def[ :direct_deps ] || []

        if validate_deps
            ws_ctx.validate_deps( 
                deps: res, err_suffix: "(in deps for '#{proj}')" )
        end

        res
    end
end

class StandardModTask < StandardProjTask
    
    def mod
        target.path[ 3 ] or raise "No module set"
    end

    def mod?( m )
        mg_id( m ) == mod()
    end

    def ws_ctx( opts = {} )
        super( not_nil( opts, :opts ).merge( mod: mod ) )
    end

    def get_mods( incl_self = false )

        res = Set.new
        res << mod() if incl_self

        case mod().to_sym
            when :integ, :test, :bin, :demo then res << mg_id( :lib )
        end

        res.to_a
    end
end

module BuildChains

    @@bgm = BitGirder::Core::BitGirderMethods

    def self.make_target( targ )
        
        case targ

            when TaskTarget then targ
            when Array then TaskTarget.new( path: targ )
            else raise "Unhandled target format: #{targ}"
        end
    end
    
    def expect_task( chain, targ )
        
        @@bgm.not_nil( chain, :chain )

        targ = self.make_target( targ )

        elt = chain.find { |elt| elt[ :task ].target == targ }

        ( elt && elt[ :task ] ) or raise "Couldn't find target #{targ} in chain"
    end

    module_function :expect_task

    # filters first by retaining only elements of some typ in the list *typs, if
    # any (otherwise rejecting none) and then selecting according to the
    # optional block, which behaves the same as Enumerable.select. Without a
    # block just returns the result array
    def elements( chain, *typs )
        
        @@bgm.not_nil( chain, :chain )

        res = Array.new( chain )
        
        unless typs.empty? 
            res = res.select do |elt| 
                typs.detect { |typ| elt[ :task ].is_a?( typ ) }
            end
        end

        block_given? ? res.select { |elt| yield( elt ) } : res
    end

    module_function :elements

    def tasks( chain, *typs )
        
        if block_given?
            elements( chain, *typs ) { |elt| yield( elt[ :task ] ) }
        else
            elements( chain, *typs ).map { |elt| elt[ :task ] }
        end
    end

    module_function :tasks

    # Returns hash of TaskTarget --> build result
    def as_result_map( chain )

        chain.inject( {} ) do |h, elt| 
            h[ elt[ :task ].target ] = elt[ :result ]
            h 
        end
    end
    
    module_function :as_result_map
end

module TestData

    ENV_TEST_DATA_PATH = "TEST_DATA_PATH"
    ENV_TEST_BIN_PATH = "TEST_BIN_PATH"

    extend BitGirderMethods

    def get_test_data_path( chain )
        
        not_nil( chain, :chain )

        dirs = BuildChains.tasks( chain ).inject( [] ) do |arr, t|
            if t.respond_to?( :data_gen_dir )
                arr << t.data_gen_dir
            else
                arr
            end
        end

        dirs.join( ":" )
    end

    module_function :get_test_data_path

    def get_bin_path_dirs( chain )
        
        not_nil( chain, :chain )

        BuildChains.tasks( chain ).inject( [] ) do |arr, t|
            arr += t.bin_path_dirs if t.respond_to?( :bin_path_dirs )
            arr
        end
    end

    module_function :get_bin_path_dirs

    def get_bin_path( chain )
        get_bin_path_dirs( chain ).join( ":" )
    end

    module_function :get_bin_path
end

class TaskRegistry < BitGirder::Core::BitGirderClass

    include Mingle

    private_class_method :new

    def self.instance
        @inst ||= new
    end

    def mg_id( id, param_name )
        MingleIdentifier.get( not_nil( id, param_name ) )
    end

    def initialize
        @selectors = []
    end

    def register( cls, &blk )

        @selectors << [ not_nil( cls, :cls ), not_nil( blk, :blk ) ]
        nil
    end

    def register_path( cls, *path )
        
        path = not_nil( path, :path ).map { |s| MingleIdentifier.get( s ) }

        register( cls ) { |targ| targ.path[ 0 .. path.size - 1 ] == path }
    end

    def create_task( targ, argh = {} )
        
        not_nil( targ, :targ )
        
        pair = 
            @selectors.find { |pair| pair[ 1 ].call( targ ) } or
                raise "Can't find executor class for #{targ}"

        pair[ 0 ].new( argh.merge( target: targ ) )
    end
end

class FullClean < TaskExecutor
    
    include BitGirder::Core

    public
    def get_direct_dependencies
        []
    end

    public
    def execute( chain )

        if File.exist?( d = @workspace.build_dir )
            BitGirder::Io.fu().rm_rf( d )
        end
    end
end

# Don't allow for any ambiguity with a partial match or anything and accept only
# the single-id path [ :clean ]
TaskRegistry.instance.register( FullClean ) do |targ| 
    targ.path.size == 1 && targ.path[ 0 ].to_sym == :clean
end

class BuilderBootstrapInstaller < TaskExecutor

    include BitGirder::Core
    include BitGirder::Io
 
    BOOTSTRAP_PROJS = 
        %w{ core ops-ruby ops-java tools io mingle mingle-codec mingle-json }
    
    require 'pathname'

    public
    def get_direct_dependencies
        []
    end

    private
    def build_dir
        "#{workspace().build_dir}/#{target()}"
    end

    private
    def rb_src_root
        "#{workspace().root}/tools/ruby"
    end

    private
    def install_static_lib
        
        res = ensure_wiped( "#{build_dir()}/lib" )

        BOOTSTRAP_PROJS.each do |proj|
            
            Dir.chdir( "#{workspace().root}/#{proj}/ruby/lib" ) do |dir|
                Dir.glob( "**/*.rb" ).each do |f|

                    if File.exist?( dest = "#{res}/#{f}" )
                        raise "#{dest} already exists"
                    else
                        fu().cp( "#{dir}/#{f}", ensure_parent( dest ) )
                    end
                end
            end
        end

        res
    end 

    private
    def install_static_bin
        
        res = "#{build_dir()}/.static-bin"
        fu().rm_rf( res ) if File.exist?( res )

        fu().cp_r( 
            file_exists( "#{workspace().root}/tools/ruby/bin" ), 
            ensure_parent( res )
        )

        res
    end

    private
    def get_gen_script( dir )
        file_exists( "#{dir}/generate-ruby-call-wrapper" )
    end

    private
    def install_script( opts )
        
        argv = []

        argv += opts[ :call_incls ].map { |i| [ "-I", i ] }.flatten
        argv << opts[ :gen_script ]

        f = opts[ :targ_script ]
        out_file = "#{opts[ :dest ]}/#{File.basename( f )}"

        argv << "--script" << f
        argv << "--out-file" << out_file
        argv << "--ruby-dir" << opts[ :ruby_dir ]
        argv += opts[ :script_incls ].map { |i| [ "--include-dir", i ] }.flatten

        UnixProcessBuilder.new( cmd: opts[ :ruby ], argv: argv ).system
    end

    private
    def build_scripts( argh )
        
        ruby = has_key( argh, :ruby )
        bin_dir = has_key( argh, :bin_dir )

        opts = {
            dest: has_key( argh, :dest ),
            ruby: ruby,
            ruby_dir: Pathname.new( "#{ruby}/../.." ).expand_path.to_s,
            gen_script: get_gen_script( bin_dir ),
            call_incls: has_key( argh, :call_incls ),
            script_incls: has_key( argh, :script_incls )
        }

        Dir.glob( "#{bin_dir}/**/*" ).
            select { |f| File.file?( f ) }.
            each { |f| install_script( opts.merge( targ_script: f ) ) }
    end

    private
    def build_live_linked_scripts( ruby )
        
        incls = BOOTSTRAP_PROJS.map do |proj|
            file_exists( "#{workspace().root}/#{proj}/ruby/lib" )
        end

        build_scripts(
            dest: ensure_wiped( "#{workspace().build_dir}/#@target/live-bin" ),
            call_incls: incls,
            script_incls: incls,
            ruby: ruby,
            bin_dir: file_exists( "#{workspace().root}/tools/ruby/bin" )
        )
    end

    public
    def execute( chain )

        # Can get this from @run_opts later if needed
        ruby = which( "ruby" ) or raise "No ruby in path"

        static_lib = install_static_lib
        static_bin = install_static_bin

        build_scripts(
            dest: ensure_wiped( "#{@workspace.build_dir}/#@target/bin" ),
            call_incls: [ static_lib ],
            script_incls: [ static_lib ],
            bin_dir: static_bin,
            ruby: ruby
        )

        if to_bool( @run_opts.get_string( :with_live_linked ) )
            build_live_linked_scripts( ruby )
        end
    end
end

TaskRegistry.instance.register_path( BuilderBootstrapInstaller, :bootstrap )

end
end
end
