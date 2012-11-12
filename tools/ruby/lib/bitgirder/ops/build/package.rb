require 'bitgirder/core'
require 'bitgirder/ops/build'

include BitGirder::Core
include BitGirder::Ops::Build
include BitGirder::Ops::Build::Ruby

module BitGirder
module Ops
module Build
module Package

class PackageTask < TaskExecutor
    
    public
    def package 
        target.path[ 2 ] or raise "No package given"
    end
    
    private
    def ws_ctx( opts = {} )
        super( { proj: package, code_type: target.path[ 0 ] }.merge( opts ) )
    end

    private
    def pkg_dir
        ws_ctx.proj_build_dir
    end
end

class PackageBuild < PackageTask

    public
    def get_direct_dependencies
        
        [ 
            TaskTarget.create( :ruby, :dist, :build, package ),
            TaskTarget.create( :java, :dist, :jar, package )
        ]
    end

    private
    def link_uniq( src, dest )
        
        if File.exist?( dest )

            msg = "Link target #{dest} for #{src} exists"

            if File.directory?( dest )
                raise "#{msg} and is a directory"
            else
                raise msg
            end
        else
            fu().ln_s( src, ensure_parent( dest ) )
        end
    end

    private
    def link_uniq_r( dir, dest_dir )
        
        Dir.chdir( dir ) do
            Dir.glob( "*" ).each do |f|

                src = "#{dir}/#{f}"
                dest = "#{dest_dir}/#{f}"

                if File.directory?( f )
                    link_uniq_r( src, dest )
                else
                    link_uniq( src, dest )
                end
            end
        end
    end

    private
    def pkg_java( chain )
        
        t = BuildChains.expect_task( chain, [ :java, :dist, :jar, package ] )
        
        Dir.glob( "#{t.get_jar_dir}/*.jar" ).each do |jar|
            link_uniq( jar, "#{pkg_dir}/lib/#{File.basename( jar )}" )
        end
    end

    private
    def pkg_ruby( chain )
        
        BuildChains.tasks( chain, RubyBuilder ).each do |t|
            t.get_incl_dirs.each do |dir|
                link_uniq_r( dir, "#{pkg_dir}/#{File.basename( dir )}" )
            end
        end
    end

    public
    def execute( chain )
        
        ensure_wiped( pkg_dir )
        pkg_java( chain )
        pkg_ruby( chain )
    end
end

TaskRegistry.instance.register_path( PackageBuild, :package, :build )

class PackageTest < PackageTask
    
    public
    def get_direct_dependencies

        [ TaskTarget.create( :package, :build, package ) ] +

        [ :ruby, :java ].map do |dist|
            TaskTarget.create( dist, :dist, :test, package )
        end
    end

    public
    def execute( chain )
        code( "Testing package" )
    end
end

TaskRegistry.instance.register_path( PackageTest, :package, :test )

class PackageInteg < PackageTask
    
    public
    def get_direct_dependencies
        
        [ TaskTarget.create( :package, :build, package ) ] +

        [ :ruby, :java ].map do |dist|
            TaskTarget.create( dist, :dist, :integ, package )
        end
    end

    public
    def execute( chain )
    end
end

TaskRegistry.instance.register_path( PackageInteg, :package, :integ )

end
end
end
end
