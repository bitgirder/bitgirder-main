require 'bitgirder/io'
require 'bitgirder/ops/build'
require 'bitgirder/ops/build/jbuilder'
require 'bitgirder/ops/build/rbuilder'
require 'bitgirder/ops/build/distro'
require 'bitgirder/ops/java'

require 'set'

module BitGirder
module Ops
module Build
module Integ

include BitGirder::Core
include BitGirder::Ops::Build
include Mingle

INTEG_CODE_TYPES = [ :ruby, :java ]

class IntegContextWriter < BitGirderClass

    include BitGirder::Io
    include MingleCodecMixin
    
    bg_attr :chain
    bg_attr :dest_dir

    private
    def get_run_classpath
        
        BuildChains.
            tasks( @chain, BitGirder::Ops::Build::Java::JavaBuilder ).
            inject( Set.new ) { |s, t| s + t.get_run_classpath( @chain ) }.to_a
    end 

    private
    def create_java_context

        { 
            classpath: get_run_classpath,
            java_home: BitGirder::Ops::Java::JavaEnvironments.get_java_home
        }
    end

    private
    def get_rb_incl_dirs

        BuildChains.
            tasks( @chain, BitGirder::Ops::Build::Ruby::RubyBuilder ).
            inject( Set.new ) { |s, t| s + t.get_rb_incl_dirs( @chain ) }.to_a
    end

    private
    def create_ruby_context
        { include_dirs: get_rb_incl_dirs }
    end

    private
    def get_go_proj_builds
        
        BuildChains.
            tasks( @chain, BitGirder::Ops::Build::Go::GoBinBuilder ).
            inject( Set.new ) { |s, t| s << t.ws_ctx.proj_build_dir }.to_a
    end

    private
    def create_go_context
        { go_proj_builds: get_go_proj_builds }
    end

    private
    def create_integ_context
        
        MingleStruct.new(
            type: :"bitgirder:testing:integ@v1/IntegrationContext",
            fields: {
                java: create_java_context,
                ruby: create_ruby_context,
                go: create_go_context,
            }
        )
    end

    public
    def write
        
        ctx = create_integ_context

        ctx_file = "#@dest_dir/integration-context.json"

        code( "Writing #{ctx_file}" )

        File.open( ensure_parent( ctx_file ), "w" ) do |io|
            io.print( mg_encode( get_json_codec, ctx ) )
        end
    end
end

class ProjIntegBuilder < StandardProjTask

    BLD = BitGirder::Ops::Build

    private
    def integ_deps_for_type( ct )

        if ( intg = ws_ctx.proj_def( code_type: ct )[ :integ ] ) &&
           ( targs = intg[ :dep_targets ] )

           targs.map { |s| TaskTarget.parse( s ) }
        else
           []
        end
    end

    private
    def deps_for_type( ct )
        
        if ws_ctx.has_project?( code_type: ct )

            res = [ TaskTarget.create( ct, :build, proj(), :test ) ]

            proj_def = ws_ctx.proj_def( code_type: ct )
            
            res += 
                ( proj_def[ :direct_deps ] || [] ).map do |proj|
                    TaskTarget.create( :integ, :build, proj )
                end
            
            res + integ_deps_for_type( ct )
        else
            []
        end
    end

    public
    def get_direct_dependencies
        INTEG_CODE_TYPES.map { |ct| deps_for_type( ct ) }.flatten
    end
        
    # Currently nothing to do in execute; so long as upstreams are built we're
    # done
    public
    def execute( chain )
    end
end

TaskRegistry.instance.register_path( ProjIntegBuilder, :integ, :build )

class ProjIntegWriter < StandardProjTask
    
    public
    def get_direct_dependencies
        [ TaskTarget.create( :integ, :build, proj ) ]
    end

    public
    def runtime_dir
        ws_ctx.proj_build_dir
    end

    public
    def execute( chain )

        IntegContextWriter.new( 
            chain: chain, 
            dest_dir: ensure_wiped( runtime_dir )
        ).
        write
    end 
end

TaskRegistry.instance.register_path( ProjIntegWriter, :integ, :write )

class IntegDistBuilder < BitGirder::Ops::Build::Distro::AbstractDistroTask

    require 'set'

    public
    def get_direct_dependencies
        
        s = 
            INTEG_CODE_TYPES.inject( Set.new ) do |s, ct|
                s + direct_deps( code_type: ct )
            end
            
        s.map { |d| TaskTarget.create( :integ, :build, d ) }
    end

    public
    def runtime_dir
        ws_ctx.proj_build_dir
    end

    public
    def execute( chain )

        IntegContextWriter.new(
            chain: chain,
            dest_dir: ensure_wiped( runtime_dir )
        ).
        write
    end
end

TaskRegistry.instance.register_path( IntegDistBuilder, :integ, :dist, :build )

end
end
end
end
