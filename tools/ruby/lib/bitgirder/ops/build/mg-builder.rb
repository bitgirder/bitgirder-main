require 'bitgirder/core'
require 'mingle'
require 'bitgirder/ops/build'

# We need to alias this here, since the bulk of the code below is also in a
# module named Mingle
MG_LANG = Mingle

module BitGirder
module Ops
module Build
module Mingle

include BitGirder::Ops::Build

include MG_LANG

MG_NS = MingleNamespace.get( "bitgirder:ops:build:mingle@v1" )
TYPE_PROJ_DEF = QualifiedTypeName.get( "#{MG_NS}/MingleProject" )

class MingleCompile < StandardModTask

    include FileSigMixin
    include JavaEnvMixin

    public
    def get_output_file
        "#{ws_ctx.mod_build_dir}/out.mgo"
    end

    private
    def get_compiler_dependencies
        
        # Save this for later when we need to use this task to build our run
        # classpath
        @cp_target_expct = 
            TaskTarget.create( :java, :build, :mingle_compiler, :bin )
        
        [ @cp_target_expct ]
    end

    private
    def make_comp_task( proj, mod )
        TaskTarget.create( :mingle, :compile, proj, mod )
    end

    private
    def get_upstream_proj_dependencies

        direct_deps.
            map { |dep| MingleIdentifier.get( dep ) }.
            inject ( [] ) do |arr, dep|

                get_mods( true ).each do |mod|
                    if ws_ctx.has_mod_src?( proj: dep, mod: mod )
                        arr << make_comp_task( dep, mod )
                    end
                end

                arr
        end
    end

    private
    def get_mod_dependencies
 
        get_mods( false ).inject( [] ) do |res, mod|
            
            if ws_ctx.has_mod_src?( mod: mod )
                res << make_comp_task( proj, mod )
            end

            res
        end
    end

    public
    def get_direct_dependencies
 
        res = get_compiler_dependencies

        res += get_upstream_proj_dependencies
        res += get_mod_dependencies

        res
    end

    private
    def get_sources
        Dir.glob( "#{ws_ctx.mod_dir}/**/*.mg" )
    end

    private
    def get_src_sig( srcs )
        get_file_sig( files: srcs )
    end

    private
    def should_build?( info, srcs, chain )

        if info
 
            prev_sig = info.fields.expect_mingle_buffer( :src_sig )
            cur_sig = get_src_sig( srcs )

            prev_sig != cur_sig
        else
            true
        end
    end

    private
    def get_run_classpath_string( chain )
        
        BuildChains.expect_task( chain, @cp_target_expct ).
            get_run_classpath( chain ).
            join( ":" )
    end

    private
    def get_include_libs( chain )
        
        lib_files =
            BuildChains.tasks( chain, MingleCompile ).map do |comp|
                comp.get_output_file
            end

        Set.new( lib_files ).map { |f| [ "--include-lib", f ] }.flatten
    end

    private
    def run_compile( srcs, chain )
        
        argv = []

        argv << "-classpath" << get_run_classpath_string( chain )
        argv << "com.bitgirder.application.ApplicationRunner"
        argv << "com.bitgirder.mingle.compiler.MingleCompilerApp"

        argv += get_include_libs( chain )
        argv += srcs.map { |src| [ "--input", src ] }.flatten
        argv << "--output" << ensure_parent( get_output_file )
        
        UnixProcessBuilder.new( cmd: jcmd( "java" ), argv: argv ).system
    end

    private
    def create_build_info( srcs )
        
        MingleStruct.new(
            type: "#{MG_NS}/MingleCompilerBuildInfo",
            fields: {
                src_sig: get_src_sig( srcs )
            }
        )
    end

    public
    def execute( chain )

        srcs = get_sources

        if should_build?( info = load_build_info, srcs, chain )
            
            run_compile( srcs, chain )
            save_build_info( info = create_build_info( srcs ) )
        end

        info
    end
end

TaskRegistry.instance.register_path( MingleCompile, :mingle, :compile )

class CleanTask < StandardProjTask

    public
    def get_direct_dependencies
        []
    end

    private
    def get_mod_glob
        ( target.path[ 3 ] || "*" ).to_s
    end

    public
    def execute( chain )
        
        to_del = []

        mod_glob = get_mod_glob

        # Delete compiled code
        to_del += 
            Dir.glob( "#{workspace.build_dir}/mingle/#{proj}/#{mod_glob}" )

        # Delete build infos
        to_del += 
            Dir.glob( 
                "#{workspace.build_dir}/.build-info/mingle/*/#{proj}/" +
                mod_glob
            )
        
        fu().rm_rf( to_del ) unless to_del.empty?
    end
end

TaskRegistry.instance.register( CleanTask ) do |targ|
    targ.to_s.start_with?( "mingle/clean" )
end

end
end
end
end
