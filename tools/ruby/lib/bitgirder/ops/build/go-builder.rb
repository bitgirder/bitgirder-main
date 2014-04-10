require 'bitgirder/io'
require 'bitgirder/core'
require 'bitgirder/ops/build'
require 'bitgirder/ops/build/distro'

require 'forwardable'
require 'set'

module BitGirder
module Ops
module Build
module Go

include BitGirder::Ops::Build

ENV_GOPATH = "GOPATH"

TYPE_MOD_BUILD_INFO = 
    QualifiedTypeName.get( :"bitgirder:ops:build:go@v1/ModBuildInfo" )

TYPE_BIN_BUILD_INFO =
    QualifiedTypeName.get( :"bitgirder:ops:build:go@v1/BinBuildInfo" )

TYPE_GO_COMMAND_RUN = 
    QualifiedTypeName.get( :"bitgirder:ops:build:go@v1/GoCommandRun" )

# Meant to be mixed into a StandardProjTask
module GoEnvMixin

    def add_src_links( mod, links )
        
        mod_src = ws_ctx.mod_dir( :mod => mod )

        return unless File.exists?( mod_src )

        Dir.chdir( mod_src ) do
 
            Dir.glob( "**/*.go" ).each do |f|

                    targ = "#{mod_src}/#{f}"

                    if prev = links[ f ]
                        raise "Linking #{targ} would conflict with " \
                              "previously linked #{prev}"
                    else
                        links[ f ] = targ
                    end
                end
        end
    end

    def link_src_dir( dest_mod, src_mods )
        
        dest = "#{ws_ctx.mod_build_dir( :mod => dest_mod )}/src"

        links = {}

        src_mods.each { |mod| add_src_links( mod, links ) }
 
        ensure_wiped( dest )

        links.keys.sort.each do |targ|
            fu().ln_s( links[ targ ], ensure_parent( "#{dest}/#{targ}" ) )
        end

        dest
    end

    def get_go_build_dependencies( mod )
        direct_deps.map { |proj| TaskTarget.create( :go, :build, proj, mod ) }
    end

    def build_bin_dir( mod )
        ws_ctx.mod_build_dir( mod: mod )
    end

    def go_cmd( nm )
        BitGirder::Io.which( nm ) or raise "No '#{nm}' in path"
    end 

    def make_go_env( chain, tsk_cls = self.class )
        
        paths = 
            BitGirder::Ops::Build::BuildChains.tasks( chain, tsk_cls ).
                inject( [] ) { |arr, tsk| arr << tsk.go_path_dir }
        
        paths << go_path_dir if self.class == tsk_cls

        # Now clean up any dupes -- removing all but the first occurrence of
        # each entry but otherwise preserving visit order
        s = Set.new
        paths = paths.delete_if { |v| s.include?( v ).tap { s << v } }

        { ENV_GOPATH => paths.join( ":" ) }
    end

    def go_packages( opts = {} )
 
        pd = opts[ :proj_def ] || ws_ctx.proj_def
        pkgs = ( pd.fields.get_mingle_list( :packages ) || [] ).to_a

        if opts[ :include_test ]
            pkgs += ( pd.fields.get_mingle_list( :test_packages ) || [] ).to_a
        end

        if pkgs.empty? && ( ! opts[ :allow_empty ] )
            raise "No packages specified in #{ws_ctx.proj_def_file}" 
        end

        pkgs
    end

    def collect_dep_sig_array( chain )
        
        BuildChains.elements( chain, GoModBuilder ).inject( [] ) do |arr, elt|
            
            if build_res = elt[ :result ]
                arr << MingleSymbolMap.create(
                    target: elt[ :task ].target.to_s,
                    api_sig: elt[ :result ][ :api_sig ],
                )
            end

            arr
        end
    end

    # Returns map keyed by ruby string, vals are mingle bufs; info is assumed to
    # have a field :dep_sigs built previously by collect_dep_sig_array()
    def prev_sigs_by_target( info )
        
        if info
            ( info[ :dep_sigs ] || [] ).inject( {} ) do |h, elt|

                h[ elt.fields.expect_string( :target ) ] = 
                    elt.fields.expect_mingle_buffer( :api_sig )
                
                h
            end
        else
            {}
        end
    end

    def upstream_api_changed?( info, chain )
 
        prev_sigs = prev_sigs_by_target( info )

        BuildChains.elements( chain, GoModBuilder ).find do |elt|
            
            if build_res = elt[ :result ]

                cur_sig = build_res.fields.expect_mingle_buffer( :api_sig )
                path = elt[ :task ].target.to_s
                prev_sigs[ path ] != cur_sig
            else
                false
            end
        end
    end

    def source_changed?( src_sig, m )
        
        if m 
            src_sig != m.expect_mingle_buffer( :src_sig )
        else
            true
        end
    end
end

class GoCleanAll < StandardTask

    public 
    def get_direct_dependencies
        []
    end

    public
    def execute( chain )
        [ ws_ctx.build_info_root, ws_ctx.code_type_build_dir ].each do |dir|
            fu().rm_rf( dir )
        end
    end
end

TaskRegistry.instance.register_path( GoCleanAll, :go, :clean_all )

class GoModCleaner < StandardProjTask

    include GoEnvMixin
    
    public
    def get_direct_dependencies
        []
    end

    public
    def execute( chain )

        fu().rm_rf( ws_ctx.proj_build_dir )

        Dir.glob( "#{build_info_root}/go/*/#{proj}" ).
            each { |f| fu().rm_rf( f ) }
    end
end

TaskRegistry.instance.register_path( GoModCleaner, :go, :clean )

class GoModBuilder < StandardModTask
    
    include GoEnvMixin
    include FileSigMixin

    public
    def get_direct_dependencies
        get_go_build_dependencies( mod )
    end

    private
    def get_src_sig( root_dir )
        
        files = Dir.glob( "#{root_dir}/**/*.go" )
        get_file_sig( files: files )
    end

    public
    def go_path_dir
        ws_ctx.mod_build_dir
    end

    private
    def should_build?( chain, info, src_sig )
        
        return true if upstream_api_changed?( info, chain )
        return true unless info
        return true if source_changed?( src_sig, info.fields )
        false
    end

    # pwd will be build root
    private
    def exec_build( pkgs, chain )

        cmd = go_cmd( "go" )
        argv = [ "install" ] + pkgs

        UnixProcessBuilder.new(
            cmd: cmd, 
            argv: argv, 
            env: make_go_env( chain ),
            show_env_in_debug: true,
        ).
        system()
    end

    private
    def create_build_info( chain, src_sig )
 
        MingleStruct.new(
            type: TYPE_MOD_BUILD_INFO,
            fields: {
                src_sig: src_sig,
                api_sig: src_sig,
                dep_sigs: collect_dep_sig_array( chain ),
            },
        )
    end

    public
    def execute( chain )
 
        src_mods = [ :lib ]
        src_mods << :test if mod?( :test )

        Dir.chdir( link_src_dir( mod(), src_mods ) ) do |src_dir|
 
            unless ( pkgs = go_packages( allow_empty: true ) ).empty?
 
                src_sig, info = get_src_sig( src_dir ), load_build_info
 
                if should_build?( chain, info, src_sig )
                    exec_build( pkgs, chain )
                    info = create_build_info( chain, src_sig )
                    save_build_info( info )
                end

                info # Return it, possibly unchanged
            end
        end
    end
end

TaskRegistry.instance.register_path( GoModBuilder, :go, :build )

class GoModTestRunner < StandardModTask
    
    include GoEnvMixin

    public
    def get_direct_dependencies
        [ TaskTarget.create( :go, :build, proj(), :test ) ]
    end

    private
    def get_test_packages
        if s = @run_opts.get_string( :test_packages )
            s.split( /,/ )
        else
            go_packages( include_test: true )
        end
    end

    public
    def execute( chain )
        
        Dir.chdir( ws_ctx.mod_build_dir ) do 

            cmd = go_cmd( "go" )
    
            env = make_go_env( chain, GoModBuilder )
    
            argv = [ "test" ]
            argv += get_test_packages
            argv << "-v"

            if filt = @run_opts.get_string( :filter_pattern )
                argv << "-test.run" << filt
            end
        
            UnixProcessBuilder.new( 
                cmd: cmd, 
                env: env, 
                argv: argv,
                show_env_in_debug: true 
            ).system
        end
    end
end

TaskRegistry.instance.register_path( GoModTestRunner, :go, :test )

class GoBinBuilder < StandardProjTask

    include GoEnvMixin
    include BitGirder::Io
    include FileSigMixin

    private
    def bin_type

        case res = @target[ 3 ] 
        when nil then nil
        when MOD_TEST, MOD_BIN then res
        else raise "Unrecognized bin build target: #{res}"
        end
    end

    private
    def test_bin?
        bin_type == MOD_TEST
    end

    private
    def lib_mod
        test_bin? ? :test : :lib
    end

    private
    def bin_mod
        test_bin? ? :test : :bin
    end

    private
    def commands_key
        test_bin? ? :test_commands : :commands
    end

    public
    def bin_src_dir
        ws_ctx.mod_dir( mod: bin_mod )
    end

    public
    def bin_build_dir
        ensure_dir( build_bin_dir( bin_mod ) )
    end

    public
    def bin_path_dirs
        [ bin_build_dir ]
    end

    public
    def get_direct_dependencies
        [ TaskTarget.create( :go, :build, proj(), lib_mod() ) ]
    end

    private
    def get_commands_to_build

        key = commands_key

        cmds = ws_ctx.proj_def.fields.get_mingle_symbol_map( key ) || {} 

        if name = @run_opts.get_string( :command )
            spec = cmds[ name ] or raise "No spec for command: #{name}"
            cmds = MingleSymbolMap.create( name => spec )
        end

        cmds
    end

    # Placeholder now; later may look in spec for more source files
    private
    def get_cmd_src_files( name, spec )
        [ file_exists( "#{bin_src_dir}/#{name}.go" ) ]
    end

    private
    def get_src_sig( name, spec, src_files )
        get_file_sig( files: src_files )
    end

    private
    def should_build?( name, src_sig, info, chain )
        
        return true if upstream_api_changed?( info, chain )

        cmd_sig = ( ( info[ :command_sigs ] || {} )[ name ] or return true )
        return true if source_changed?( src_sig, cmd_sig )

        false
    end

    private
    def exec_build( name, spec, src_files, chain )

        cmd = go_cmd( "go" )
        env = make_go_env( chain, GoModBuilder )

        argv = %w{ build }

        bin_dir = bin_build_dir
        argv << "-o" << "#{bin_dir}/#{name}"

        argv += src_files
 
        b = UnixProcessBuilder.new( cmd: cmd, argv: argv, env: env )
        b.show_env_in_debug = true
        b.system
    end

    private
    def get_command_build_info( name, src_sig, chain )
        
        MingleSymbolMap.create( src_sig: src_sig )
    end

    private
    def build_command( name, spec, src_files, info, chain )

        src_sig = get_src_sig( name, spec, src_files )

        if should_build?( name, src_sig, info, chain )

            exec_build( name, spec, src_files, chain )
            get_command_build_info( name, src_sig, chain )
        end
    end

    private
    def create_build_info( cmd_sigs, info_prev, chain )
        
        if info_prev
            if cmd_sigs_prev = info_prev[ :command_sigs ]
                cmd_sigs = cmd_sigs_prev.to_hash.merge( cmd_sigs )
            end
        end

        flds = {
            dep_sigs: collect_dep_sig_array( chain ),
            command_sigs: cmd_sigs,
        }

        MingleStruct.new( type: TYPE_BIN_BUILD_INFO, fields: flds )
    end

    private
    def process_info( cmd_sigs, info, chain )

        unless cmd_sigs.empty?

            info = create_build_info( cmd_sigs, info, chain ) 
            save_build_info( info )
        end

        info
    end

    public
    def execute( chain )

        cmds = get_commands_to_build

        info = load_build_info || MingleStruct.new( type: TYPE_BIN_BUILD_INFO )
        cmd_sigs = {}

        cmds.each_pair do |name, spec| 

            src_files = get_cmd_src_files( name, spec )

            if cmd_sig = build_command( name, spec, src_files, info, chain ) 
                cmd_sigs[ name ] = cmd_sig
            end
        end

        process_info( cmd_sigs, info, chain )
    end
end

TaskRegistry.instance.register_path( GoBinBuilder, :go, :build_bin )

class GoRunQuick < StandardProjTask
    
    include GoEnvMixin

    public
    def get_direct_dependencies
        [ TaskTarget.create( :go, :build, proj() ) ]
    end

    public
    def execute( chain )

        cmd = go_cmd( "go" )
        env = make_go_env( chain )

        argv = [ "run", @run_opts.expect_string( :file ) ]
        argv += ( run_ctx[ :argv_remain ] || [] )

        UnixProcessBuilder.new( cmd: cmd, argv: argv, env: env ).exec
    end
end

TaskRegistry.instance.register_path( GoRunQuick, :go, :run_quick )

class GoCommandRunner < StandardProjTask 
    
    public
    def get_direct_dependencies
        [ TaskTarget.create( :go, :build_bin, proj() ) ]
    end

    public
    def execute( chain )

        cmd_base = @run_opts.expect_string( :command )
        cmd = file_exists( "#{ws_ctx.mod_build_dir( mod: :bin )}/#{cmd_base}" )

        argv = run_ctx[ :argv_remain ] || []

        UnixProcessBuilder.new( cmd: cmd, argv: argv ).exec
    end
end

TaskRegistry.instance.register_path( GoCommandRunner, :go, :run_command )

class GoTestDataGenerator < StandardProjTask
 
    include GoEnvMixin

    public
    def get_direct_dependencies
        [ TaskTarget.create( :go, :build_bin, proj(), :test ) ]
    end

    public
    def data_gen_dir
        ws_ctx.mod_build_dir( mod: :test_data )
    end

    private
    def replace_arg_vars( arg )
        
        repls = MingleSymbolMap.create(
            data_gen_dir: data_gen_dir,
        )

        # Replaces all ${...} sequences with the replacement from repls, failing
        # if not present; \${...} will be carried over untouched
        arg.to_s.gsub( /(^|[^\\])\$\{([^\}]*)\}/ ) { |m| 
            if repl = ( repls[ var = $2 ] )
                "#$1#{repl}"
            else
                raise "No value known for ${#{var}}"
            end
        }
    end

    private
    def run_go_command( gen_obj )
        
        cmd = file_exists( "#{build_bin_dir( :test )}/#{gen_obj[ :command ]}" )
        
        argv = ( gen_obj[ :argv ] || [] ).
               map { |arg| replace_arg_vars( arg ) }
        
        UnixProcessBuilder.new( cmd: cmd, argv: argv ).system
    end

    private
    def run_generator( gen_id, gen_obj, chain )
        
        code( "#{gen_id} type: #{gen_obj.class}" )

        case gen_obj.type
            when TYPE_GO_COMMAND_RUN then run_go_command( gen_obj )
            else raise "Unhandled value for #{gen_id}: #{gen_obj}"
        end
    end

    public
    def execute( chain )
        
        pd = ws_ctx.proj_def
        ( pd[ :test_data_generators ] || {} ).each_pair do |gen_id, gen_obj| 
            run_generator( gen_id, gen_obj, chain )
        end
    end
end

TaskRegistry.instance.register_path( GoTestDataGenerator, :go, :gen_test_data )

# Should be mixed in to an AbstractDistroTask or something which responds
# similarly
module GoDistMixin
    
    include GoEnvMixin
    include BitGirder::Io

    def dist_work_dir( mod )
        "#{dist_build_dir}/work/#{mod}"
    end 

    def get_dist_packages( include_test = false )
        
        res = {}

        direct_deps.each do |dep|
            go_packages( 
                proj_def: ws_ctx.proj_def( proj: dep ),
                include_test: include_test 
            ).each do |pkg|
                if prev = res[ pkg ]
                    raise "#{pkg} defined in both #{prev} and #{dep}"
                else
                    res[ pkg ] = dep
                end
            end
        end

        res.keys.map { |k| k.to_s }
    end
end

class GoDistTask < BitGirder::Ops::Build::Distro::AbstractDistroTask
    
    include GoDistMixin
    include GoEnvMixin
    include BitGirder::Io

    private
    def get_dist_build_deps( mod )

        direct_deps.inject( [] ) do |arr, proj|
            arr << TaskTarget.create( :go, :build, proj, mod )
        end
    end

    private
    def link_dist_src( opts )

        dest = has_key( opts, :dest )
        mod = has_key( opts, :mod )
        chain = has_key( opts, :chain )
 
        linker = TreeLinker.new( :dest => dest )

        BuildChains.tasks( chain, GoModBuilder ).each do |tsk|

            if tsk.mod?( mod )

                linker.update_from( 
                    :src => "#{tsk.go_path_dir}/src", 
                    :selector => "**/*.go" 
                )
            end
        end

        linker.build
    end
end

class GoDistTester < GoDistTask
    
    public
    def get_direct_dependencies
        get_dist_build_deps( :test )
    end

    public
    def execute( chain )

        work_dir = dist_work_dir( :test )

        src_dir = link_dist_src(
            :chain => chain, :dest => "#{work_dir}/src", :mod => :test )

        cmd = go_cmd( "go" )

        argv = [ "test", "-v" ]
        argv += get_dist_packages( true )
        
        env = { ENV_GOPATH => work_dir }

        UnixProcessBuilder.new( 
            cmd: cmd, 
            argv: argv, 
            env: env,
            show_env_in_debug: true
        ).system
    end
end

TaskRegistry.instance.register_path( GoDistTester, :go, :dist, :test )

class ErrorfBuilder < StandardModTask
    
    include GoEnvMixin

    public
    def get_direct_dependencies
        []
    end

    private
    def errorf_file( go_pkg )
        code( "ws_ctx: #{ws_ctx}" )
        "#{ws_ctx.mod_dir}/#{go_pkg}/errorf.go"
    end

    private
    def gen_errorf( go_pkg )

        pkg = go_pkg.split( "/" )[ -1 ] or 
            raise "Empty package in go module #{go_pkg}"

        <<-END_SRC
package #{pkg}

import (
    "fmt"
    "errors"
)

func libError( msg string ) error {
    return errors.New( "#{go_pkg}: " + msg )
}

func libErrorf( tmpl string, argv ...interface{} ) error {
    return fmt.Errorf( "#{go_pkg}: " + tmpl, argv... )
}
        END_SRC
    end

    public
    def execute( chain )

        opt = mg_id( :go_package )

        if go_pkg = @run_opts.get_string( opt )

            File.open( ensure_parent( errorf_file( go_pkg ) ), "w" ) do |io|
                io.print( gen_errorf( go_pkg ) )
            end
        else
            raise "No run option given for '#{opt}'"
        end
    end
end

TaskRegistry.instance.register_path( ErrorfBuilder, :go, :errorf_gen )

end
end
end
end
