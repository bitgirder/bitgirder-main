require 'bitgirder/io'
require 'bitgirder/ops/build'
require 'bitgirder/ops/build/distro'
require 'bitgirder/ops/build/mg-builder'
require 'mingle'

require 'set'
require 'digest/md5'
require 'digest/sha1'
require 'erb'
require 'rexml/document'
require 'pathname'

module BitGirder
module Ops
module Build
module Java

ENV_JBUILDER_TEST_RESOURCES = "JBUILDER_TEST_RESOURCES"

NS_STR = "bitgirder:ops:build:java@v1"

include BitGirder::Core
include Mingle
include BitGirder::Ops::Build

MOD_TEST = MingleIdentifier.get( :test )
ALL_MODS = %w{ lib test bin }.map { |s| MingleIdentifier.get( s ) }

MG_OPS = BitGirder::Ops::Build::Mingle

# Assumes it is being mixed into something which has @workspace and @run_opts
# defined -- notably all tasks have these
module JavaTaskMixin
   
    include BitGirder::Core::BitGirderMethods

    def has_mg_src?( opts = {} )
        ws_ctx.has_mod_src?( opts.merge( code_type: :mingle ) )
    end

    def mod_src( opts = {} )
        "#{ws_ctx.mod_dir( opts )}/src"
    end

    def is_build_dry_run?
        @run_opts.get_mingle_boolean( :build_dry_run )
    end

    def gen_build_dir( opts )

        not_nil( opts, :opts )
        gen_key = opts[ :gen_key ] || @generator_key

        not_nil( gen_key, :gen_key )
        
        "#{ws_ctx.mod_build_dir( opts )}/#{gen_key}"
    end

    def gen_src_dir( opts = {} )
        "#{gen_build_dir( opts )}/gensrc"
    end

    def gen_classes_dir( opts = {} )
        "#{gen_build_dir( opts )}/classes"
    end

    def mod_resource
        "#{ws_ctx.mod_dir}/resource"
    end

    def mod_classes
        "#{ws_ctx.mod_build_dir}/classes"
    end

    def get_extra_projs

        if str = run_opts[ :with_proj ] 
            [ str ]
        else
            []
        end
    end

    def cp_includes_test?
        @run_opts.get_boolean( :cp_includes_test )
    end

    def get_build_classpath( chain )
        
        chain.inject( [] ) do |cp, elt|
            
            case task = elt[ :task ] 
                when JavaBuilder then cp << task.mod_classes
                when AbstractCodegenTask then cp << task.gen_classes_dir
                when MavenPullTask then cp += task.read_classpath
                else cp
            end
        end
    end

    def get_run_classpath( chain )
        
        arr = chain.inject( get_build_classpath( chain ) ) do |arr, elt|

            case t = elt[ :task ]
                
                when JavaBuilder 
                    ( File.exist?( f = t.mod_resource ) && arr << f ) || arr
 
                when BitGirder::Ops::Build::Mingle::MingleCompile
                    arr << ws_ctx.code_type_build_dir( code_type: :mingle )

                else arr
            end
        end

        Set.new( arr ).to_a
    end

    def make_mg_codegen_dep( proj, mod )
        TaskTarget.create( :java, :codegen, :mingle, proj, mod )
    end

    def proj_def_mtime
        File.exist?( f = ws_ctx.proj_def_file ) ? File.mtime( f ) : nil
    end
end

class AbstractJavaModuleTask < StandardModTask
    
    include JavaEnvMixin
    include FileSigMixin
    include JavaTaskMixin

    public
    def is_test?
        mod.to_sym == :test
    end

    private
    def get_resources
        Dir.glob( "#{mod_resource}/**/*" ).select { |f| File.file?( f ) }
    end

    public
    def get_declared_test_deps
        
        res = []

        pd = ws_ctx.proj_def

        if deps = pd.fields.get_mingle_list( :test_deps )
            deps.each { |dep| res << TaskTarget.parse( dep ) }
        end

        res
    end

    private
    def get_mods
        
        res = Set.new( [ mod ] )

        case mod.to_sym
            when :test, :bin, :demo then res << mg_id( :lib )
        end

        res.to_a
    end
end

class JavaBuilder < AbstractJavaModuleTask

    include Mingle
    include BitGirder::Io

    private
    def is_empty_proj?
        ! File.exist?( ws_ctx.proj_def_file )
    end

    private
    def make_java_dep( proj, mod )
        TaskTarget.create( :java, :build, proj, mod )
    end

    private
    def get_self_deps( mods )
        
        mods.inject( [] ) do |arr, mod| 

            if mod != mod() && 
               ( ws_ctx.has_mod_src?( mod: mod ) || has_mg_src?( mod: mod ) )

                arr << make_java_dep( proj(), mod )
            else
                arr 
            end
        end
    end

    private
    def get_base_direct_deps

        mods = get_mods
        
        mods.inject( get_self_deps( mods ) ) do |arr, mod|

            direct_deps.each do |dep|

                if ws_ctx.has_mod_src?( proj: dep, mod: mod ) || 
                   has_mg_src?( proj: dep, mod: mod )

                    arr << make_java_dep( dep, mod ) 
                end
            end

            arr
        end
    end

    private
    def get_test_bootstrap_deps
        
        if mod.to_sym == :test
            case proj.to_sym
                when :core then [ make_java_dep( :core, :bin ) ]
                when :testing then [ make_java_dep( :testing, :bin ) ]
                else []
            end
        else
            []
        end
    end

    private
    def get_codegen_deps
        
        if cdgn = ( ws_ctx.proj_def[ :codegen ] || {} )[ mod ]

            cdgn.inject( [] ) do |arr, pair|

                arr << 
                    TaskTarget.create( :java, :codegen, pair[ 0 ], proj, mod )
            end
        else
            []
        end
    end

    private
    def get_maven_pull_deps
        
        if ws_ctx.proj_def[ :maven_deps ]
            [ TaskTarget.create( :java, :maven, :pull, proj ) ]
        else
            []
        end
    end

    private
    def get_mingle_deps
        has_mg_src? ? [ make_mg_codegen_dep( proj, mod ) ] : []
    end

    public
    def get_direct_dependencies
 
        if is_empty_proj?
            []
        else
            get_base_direct_deps +
            get_test_bootstrap_deps + 
            get_codegen_deps + 
            get_maven_pull_deps +
            get_mingle_deps
        end
    end

    private
    def get_src_sig( srcs )
        get_file_sig( files: srcs + get_resources )
    end

    private
    def sigs_by_targ( info )
        
        ( info[ :built_with ] || [] ).inject( {} ) do |h, elt|
            
            h[ elt.fields.expect_string( :target ) ] =
                elt.fields.expect_mingle_buffer( :sig )
            
            h
        end
    end

    private
    def upstream_java_build_changed?( chain_elt, prev_sigs )

        t = chain_elt[ :task ]

        if prev_sig = prev_sigs[ t.target.to_s ]
 
            sig = chain_elt[ :result ].fields.get_mingle_buffer( :api_sig )
            prev_sig != sig
        else
            false
        end
    end

    private
    def upstream_codegen_changed?( elt, prev_sigs )
        
        t = elt[ :task ]

        if prev_sig = prev_sigs[ t.target.to_s ] 

            info = t.load_build_info( true )
            cur_sig = info.fields.expect_mingle_buffer( :gen_sig )
            prev_sig != cur_sig
        else
            false
        end
    end

    private
    def upstream_api_changed?( info, chain )
        
        prev_sigs = sigs_by_targ( info )

        changed = chain.find do |elt|
            
            case elt[ :task ]

                when JavaBuilder 
                    upstream_java_build_changed?( elt, prev_sigs )
                
                when AbstractCodegenTask
                    upstream_codegen_changed?( elt, prev_sigs )

                else false
            end
        end

        changed != nil
    end

    private
    def should_build?( info, srcs, chain )
        
        if info

            sig_prev = info.fields.expect_mingle_buffer( :src_sig )
            sig_cur = get_src_sig( srcs )

            if sig_prev == sig_cur
                upstream_api_changed?( info, chain )
            else
                true
            end
        else
            true
        end
    end

    private
    def get_java_src( chain )
        
        res = Dir.glob( "#{mod_src}/**/*.java" )
 
        BuildChains.tasks( chain ).inject( res ) do |arr, t|

            if t.is_a?( AbstractCodegenTask ) && 
               t.mod == mod() && t.proj == proj() &&
               ( ! t.compile_generated? )

                arr += Dir.glob( "#{t.gen_src_dir}/**/*.java" )
            end

            arr
        end
    end
    
    private
    def update_api_sig_classes( md5 )
        
        if File.exist?( mod_classes )
    
            classes = 
                Dir.chdir( mod_classes ) { |dir|
                    Dir.glob( "**/*.class" ).map { |cls_file|
                        cls_file.gsub( /\//, "." ).sub( /\.class$/, "" )
                    }
                }
        
            argv = []
    
            argv << "-protected"
            argv << "-classpath" << mod_classes
            argv += classes
        
            UnixProcessBuilder.new( cmd: jcmd( "javap" ), argv: argv ).
                popen( "r" ) { |io| io.each_line { |line| md5 << line } }
        end
    end
    
    private
    def update_api_sig_resources( md5 )

        get_resources.sort.each do |file| 
            File.open( file ) { |io| io.each_line { |line| md5 << line } }
        end
    end
 
    private
    def get_api_sig
 
        md5 = Digest::MD5.new
    
        update_api_sig_classes( md5 )
        update_api_sig_resources( md5 )
    
        MingleBuffer.new( md5.digest )
    end

    private
    def get_built_with( chain )
        
        chain.inject( [] ) do |arr, elt|
            
            case b = elt[ :task ]
                
                when JavaBuilder
                    arr << { 
                        target: b.target.to_s, 
                        sig: elt[ :result ][ :api_sig ] 
                    }
                
                when AbstractCodegenTask
                    arr << {
                        target: b.target.to_s,
                        sig: elt[ :result ][ :gen_sig ]
                    }

                else arr
            end
        end
    end

    private
    def create_build_info( srcs, chain )
        
        MingleStruct.new(
            type: "#{NS_STR}/JavaBuildInfo",
            fields: { 
                api_sig: get_api_sig,
                src_sig: get_src_sig( srcs ),
                built_with: get_built_with( chain )
            }
        )
    end

    private
    def do_dry_run_build
        code( "Doing dry run build of #{target}" )
    end

    private
    def do_real_build( srcs, chain )

        fu().rm_rf( mod_classes )

        unless srcs.empty?
        
            argv = []
            
            argv << "-Xlint:unchecked,deprecation"
            argv << "-classpath" << get_build_classpath( chain ).join( ":" )
            argv << "-d" << ensure_dir( mod_classes )
     
            argv += srcs
    
            UnixProcessBuilder.new( cmd: jcmd( "javac" ), argv: argv ).system
        end

        save_build_info( info = create_build_info( srcs, chain ) )
        
        info
    end

    public
    def execute( chain )
        
        unless is_empty_proj?

            srcs = get_java_src( chain )
    
            if should_build?( info = load_build_info, srcs, chain )
     
                if is_build_dry_run?
                    do_dry_run_build
                else
                    console( "Building #@target" )
                    info = do_real_build( srcs, chain )
                end
            end
    
            info
        end
    end
end

TaskRegistry.instance.register_path( JavaBuilder, :java, :build )

POM_TEMPL_HEADER = <<END
<project xmlns="http://maven.apache.org/POM/4.0.0" 
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/maven-v4_0_0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.bitgirder</groupId>
    <version>0.1</version>
    <artifactId>bitgirder-<%= proj() %></artifactId>
    <packaging><%= packaging %></packaging>
END

# bindings: deps
POM_TEMPL_DEPS = <<END
    <dependencies>
        <% deps.each do |dep| %>
            <dependency>
                <% 
                dep.each_pair do |k, v| 
                    mvn_key = k.format( :lc_camel_capped )
                    mvn_val = v.to_s
                %>
                <<%= mvn_key %>><%= mvn_val %></<%= mvn_key %>>
                <% end %>
            </dependency>
        <% end %>
    </dependencies>
END

# bindings: cp_file
POM_TEMPL_BUILD = <<END
    <build>
        <plugins>
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-dependency-plugin</artifactId>
                <executions>
                    <execution>
                        <id>build-classpath</id>
                        <phase>generate-sources</phase>
                        <goals>
                            <goal>build-classpath</goal>
                        </goals>
                        <configuration>
                            <outputFile>
                                <%= cp_file %>
                            </outputFile>
                            <outputAbsoluteArtifactFilename>
                                true
                            </outputAbsoluteArtifactFilename>
                        </configuration>
                    </execution>
                </executions>
            </plugin>
        </plugins>
    </build>
END

POM_TEMPL_FOOTER = <<END
</project>
END

class MavenPullTask < StandardProjTask
    
    include JavaEnvMixin
    include FileSigMixin
    include JavaTaskMixin

    #Override
    public
    def proj
        target.path[ 3 ]
    end

    public
    def mod
        target.path[ 4 ]
    end

    public
    def get_direct_dependencies
        []
    end

    private
    def build_dir
        "#{ws_ctx.proj_build_dir}/maven/pull"
    end

    private
    def generated_pom
        "#{build_dir}/pom.xml"
    end

    private
    def get_classpath_file
        "#{build_dir}/classpath.txt"
    end

    # Reads and parses the classpath file, returning elts as an array
    public
    def read_classpath
        File.open( get_classpath_file ) { |io| io.read.split( /:/ ) }
    end

    # This actually will return false if the generated classpath and the proj
    # def are changed in the same second, but we'll just be willing to live with
    # that for the moment, given that it's quite unlikely that generating a pom
    # and then running mvn would happen in < 1s. We can always make this alg
    # based on a diest of the proj def associated with the time the pom was
    # generated, but that seems extreme at the moment
    private
    def should_build?
        
        if File.exist?( cp_file = get_classpath_file )
            File.mtime( cp_file ) <= proj_def_mtime
        else
            true
        end
    end

    private
    def get_templ_binding

        packaging = "pom"
        deps = ws_ctx.proj_def[ :maven_deps ]
        cp_file = get_classpath_file

        binding
    end

    private
    def generate_pom
        
        bnd = get_templ_binding

        File.open( pom = generated_pom, "w" ) do |io|

            code( "generating #{pom}" )

            templ =
                POM_TEMPL_HEADER + POM_TEMPL_DEPS + POM_TEMPL_BUILD +
                POM_TEMPL_FOOTER

            doc_str = ERB.new( templ ).result( bnd )

            # parse it first as a sanity check
            REXML::Document.new( doc_str ).write( io )
        end
    end

    private
    def get_mvn_binary
        
        if res = run_opts[ :mvn ] 
            file_exists( res )
        else
            which( "mvn" ) or raise "No mvn in path"
        end
    end
 
    private
    def exec_mvn
        
        argv = []

        argv << "-f" << generated_pom
        argv << "package"

        UnixProcessBuilder.new( cmd: get_mvn_binary, argv: argv ).system
    end

    public
    def execute( chain )

        if should_build?
            
            ensure_wiped( build_dir )
            generate_pom
            exec_mvn
        end
    end
end

TaskRegistry.instance.register_path( MavenPullTask, :java, :maven, :pull )

class AbstractCodegenTask < AbstractJavaModuleTask
    
    bg_attr :generator_key, 
            required: false,
            processor: lambda { |s| MingleIdentifier.get( s ) }

    bg_attr :build_info_type_name

    #Override
    public
    def proj
        target.path[ 3 ]
    end

    public
    def mod
        target.path[ 4 ]
    end

    public
    def get_gen_sig

        files = 
            Dir.glob( "#{gen_src_dir}/**/*" ).
                select { |f| File.file?( f ) }.
                sort

        get_file_sig( files: files )
    end

    private
    def expect_codegen
        
        cdgn = ws_ctx.proj_def[ :codegen ] or raise "No codegen defs in #{proj}"
        mod_cdgn = cdgn[ mod ] or raise "No codegen for module #{proj}/#{mod}"

        mod_cdgn[ @generator_key ] or 
            raise "No codegen def for #@generator_key in #{proj}/#{mod}"
    end

    public
    def get_direct_dependencies
        []
    end
    
    private
    def get_src_sig( sources )
        get_file_sig( files: sources )
    end

    private
    def should_build?( info, sources )
        
        if info
           sig = get_src_sig( sources )
           prev_sig = info.fields.expect_mingle_buffer( :src_sig )

           sig != prev_sig
        else
            true
        end
    end

    # Can be overridden/made abstract as necessary
    private
    def get_generator_inputs( chain )

        Dir.chdir( ws_ctx.proj_root ) do |dir|
            expect_codegen.
                inject( Set.new ) { |s, glob| s + Dir.glob( glob.to_s ) }.
                map { |f| "#{dir}/#{f}" }.
                to_a
        end
    end

    bg_abstract :generate # ( srcs, chain )

    # Overridable -- returns an array, not a path string
    private
    def get_compile_classpath( chain )
        []
    end

    private
    def compile_generated( chain )
        
        argv = []

        argv << "-d" << ensure_wiped( gen_classes_dir )
        
        unless ( cp = get_compile_classpath( chain ) ).empty?
            argv << "-classpath" << cp.join( ":" )
        end

        argv += Dir.glob( "#{gen_src_dir}/**/*.java" )
        
        UnixProcessBuilder.new( cmd: jcmd( "javac" ), argv: argv ).system
    end

    # Overridable
    private 
    def install_generated_resources( chain ); end

    # Overridable
    public
    def compile_generated?
        true
    end

    private
    def generate_and_compile( srcs, chain )
        
        generate( srcs, chain )
        compile_generated( chain ) if compile_generated?
        install_generated_resources( chain )
    end

    private
    def create_build_info( srcs )
        
        MingleStruct.new(
            type: "#{NS_STR}/#@build_info_type_name",
            fields: { 
                src_sig: get_src_sig( srcs ),
                gen_sig: get_gen_sig
            }
        )
    end

    public
    def execute( chain )
        
        srcs = get_generator_inputs( chain )

        if should_build?( info = load_build_info, srcs )
    
            generate_and_compile( srcs, chain )
            save_build_info( info = create_build_info( srcs ) )
        end

        info
    end
end

class XjcExecutor < AbstractCodegenTask
    
    def initialize( opts )
        super( 
            opts.merge( 
                generator_key: :xjc,
                build_info_type_name: "XjcBuildInfo" 
            ) 
        )
    end

    private
    def generate( srcs, chain )

        fu().rm_rf( gen_src_dir )

        argv = []
        argv << "-d" << ensure_wiped( gen_src_dir )
        argv += srcs

        UnixProcessBuilder.new( cmd: jcmd( "xjc" ), argv: argv ).system
    end
    
    def install_generated_resources( chain )
        
        root = Pathname.new( ws_ctx.mod_dir( code_type: :xml ) )

        get_generator_inputs( chain ).map do |s| 
            
            rel = Pathname.new( s ).relative_path_from( root ).to_s
            dest = "#{gen_classes_dir}/xsd/#{rel}"

            fu().ln_s( s, ensure_parent( dest ) )
        end
    end
end

TaskRegistry.instance.register_path( XjcExecutor, :java, :codegen, :xjc )

class WsImportExecutor < AbstractCodegenTask

    def initialize( opts )

        super( 
            opts.merge( 
                generator_key: :ws_import,
                build_info_type_name: "WsImportBuildInfo"
            ) 
        )
    end

    private
    def generate( srcs, chain )

        argv = []

        argv += srcs
        argv << "-keep"
        argv << "-extension"
        argv << "-d" << ensure_wiped( gen_src_dir )
        argv << "-Xnocompile"

        UnixProcessBuilder.new( cmd: jcmd( "wsimport" ), argv: argv ).system
    end
end

TaskRegistry.
    instance.
    register_path( WsImportExecutor, :java, :codegen, :ws_import )

class MingleCodegen < AbstractCodegenTask
    
    GEN_KEY = MingleIdentifier.get( :java_mingle_codegen )

    def initialize( opts )
 
        super(
            opts.merge(
                generator_key: GEN_KEY,
                build_info_type_name: "MingleCodegenBuildInfo"
            )
        )
    end

    private
    def get_mingle_codegen_deps
        
        [ 
            TaskTarget.create( :mingle, :compile, proj, mod ),
        
            @compile_targ = 
                TaskTarget.create( :java, :build, :mingle_bind, :lib ),

            @codegen_targ = 
                TaskTarget.create( :java, :build, :mingle_codegen, :bin )
        ]
    end

    private
    def get_upstream_codegen_deps
        
        direct_deps.inject( [] ) do |arr, dep|
            
            if has_mg_src?( proj: dep )
                arr << TaskTarget.create( :java, :codegen, :mingle, dep, mod )
            else
                arr
            end
        end
    end

    # Override
    public
    def get_direct_dependencies
        
        res = get_mingle_codegen_deps
        res += get_upstream_codegen_deps

        get_mods().reject { |mod| mod == mod() }.each do |mod|
            
            has_src = ws_ctx.has_mod_src?( mod: mod, code_type: :mingle )
            res << make_mg_codegen_dep( proj(), mod ) if has_src
        end

        res
    end

    private
    def expect_comp_task( chain )
        BuildChains.expect_task( chain, [ :mingle, :compile, proj, mod ] )
    end

    # Override
    private
    def get_generator_inputs( chain )

        [ expect_comp_task( chain ).get_output_file ]
    end

    private
    def get_run_classpath_string( chain )
        
        BuildChains.expect_task( chain, @codegen_targ ).
            get_run_classpath( chain ).
            join( ":" )
    end

    private
    def add_ctl_obj_file_in( dir, targ_set )
        
        if File.exist?( f = "#{dir}/java-codegen-control.json" )
            targ_set << f
            code( "added #{f}" )
        else
            code( "no ctl obj in #{dir}" )
        end
    end

    private
    def get_ctl_obj_files( chain )
        
        ( BuildChains.tasks( chain ) << self ).
            inject( Set.new ) do |s, task|
                
                if task.is_a?( JavaBuilder ) || task.is_a?( MingleCodegen )
                    proj_dir = ws_ctx.proj_dir( proj: task.proj )
                    add_ctl_obj_file_in( proj_dir, s )
                    add_ctl_obj_file_in( "#{proj_dir}/#{task.mod}", s )
                end

                s
            end
    end

    private
    def generate( srcs, chain )
        
        argv = []

        argv << "-classpath" << get_run_classpath_string( chain )
        argv << "com.bitgirder.application.ApplicationRunner"
        argv << "com.bitgirder.mingle.codegen.MingleCodeGeneratorApp"

        argv << "--language" << "java"
        argv += srcs.map { |src| [ "--input", src ] }.flatten
        argv << "--out-dir" << ensure_wiped( gen_src_dir )

        argv += 
            get_ctl_obj_files( chain ).
            map { |f| [ "--control-object", f ] }.
            flatten

        UnixProcessBuilder.new( cmd: jcmd( "java" ), argv: argv ).system
    end

    public
    def compile_generated?
        false
    end

    private
    def install_autoload( chain )
        
        mgo = expect_comp_task( chain ).get_output_file

        # rsrc will be the resource name of mgo relative to the path leading up
        # to it when the runtime path is created later, either because the
        # runtime path will point directly at the build directory of mgo (ie,
        # everthing up to path_toks[ -3 ]) or because the jar includes a copy of
        # mgo at the path specified here
        path_toks = mgo.split( /\// )
        rsrc = path_toks[ -3 .. -1 ].join( "/" )

        ld_file = "#{gen_classes_dir}/mingle-runtime-autoload.txt"
        File.open( ensure_parent( ld_file ), "w" ) { |io| io.puts( rsrc ) }
    end

    # Override
    private
    def install_generated_resources( chain )
        
        if File.exist?( init = "#{gen_src_dir}/mingle-bind-init.txt" )
            fu().ln_s( init, ensure_dir( gen_classes_dir ), force: true )
        end

        install_autoload( chain )
    end
end

TaskRegistry.instance.register_path( MingleCodegen, :java, :codegen, :mingle )

module JavaTesting
    
    extend BitGirder::Core::BitGirderMethods

    def get_test_jvm_args( run_ctx = nil )
        [ "-Xmx512m" ]
    end

    module_function :get_test_jvm_args

    # jv_tsk needs to be an AbstractJavaModuleTask (JavaTestRunner, JavaBuilder,
    # etc) or something else which responds correctly to mod_classes
    def get_test_class_names( jv_tsk )
        
        not_nil( jv_tsk, :jv_tsk )

        all_names = Dir.chdir( jv_tsk.mod_classes ) do |dir|
            Dir.glob( File.join( "**", "*.class" ) ).
                map { |f| f.sub( /((\$.*)|(\.class$))/, "" ).gsub( /\//, "." ) }
        end

        Set.new( all_names ).to_a
    end

    module_function :get_test_class_names

    def get_testing_test_runner_args( names, run_ctx )
        
        not_nil( names, :names )
        not_nil( run_ctx, :run_ctx )

        res =
            [ "com.bitgirder.application.ApplicationRunner",
               "com.bitgirder.testing.UnitTestRunner" ]

        res += ( run_ctx[ :argv_remain ] || [] )

        names.inject( res ) { |arr, nm| arr << "--test-class" << nm }
    end

    module_function :get_testing_test_runner_args

    def get_test_cp_extra( run_opts )
        
        nm = ENV_JBUILDER_TEST_RESOURCES
    
        if f = ( run_opts[ :test_cp_extra ] || ENV[ nm ] )
            
            f = f.to_s

            if File.exist?( f ) then [ f ]
            else raise "Location specified in env #{nm} does not exist: #{f}"
            end
        else
            []
        end
    end

    module_function :get_test_cp_extra

    def add_filter_args( argv, run_opts )
        
        if filt = run_opts.get_string( :filter_pattern )
            argv << "--filter-pattern" << filt
        end
    end

    module_function :add_filter_args

    def add_test_run_paths( path, chain )

        opt_add = lambda { |nm, arr|
            path << "-Dbitgirder.#{nm}=#{arr}" unless arr.empty?
        }

        opt_add.call( "testDataPath", TestData.get_test_data_path( chain ) )
        opt_add.call( "testBinPath", TestData.get_bin_path( chain ) )
    end

    module_function :add_test_run_paths
end

class JavaTestRunner < AbstractJavaModuleTask

    private
    def get_test_classpath( chain )
        
        res = get_run_classpath( chain )
        res << mod_classes
        res += JavaTesting.get_test_cp_extra( @run_opts )
    end

    public
    def get_direct_dependencies

        res = [ TaskTarget.create( :java, :build, proj(), mod() ) ]

        res +=
            get_extra_projs.map do |proj|
                TaskTarget.create( :java, :build, proj, mod()  )
            end
        
        res += get_declared_test_deps

        res
    end

    private
    def get_test_class_names
        JavaTesting.get_test_class_names( self )
    end

    private
    def get_test_runner_args( chain, names )
        
        res = chain.inject( nil ) do |res, elt|
            
            case elt[ :task ].target.path[ -2 .. -1 ].join( "/" )

                when "core/bin" 
                    [ "com.bitgirder.core.UnitTestRunner" ] + names

                when "testing/bin" 
                    JavaTesting.get_test_jvm_args( @run_ctx ) +
                    JavaTesting.get_testing_test_runner_args( names, @run_ctx )
                
                else res
            end
        end

        ( res && res.flatten ) or raise "No invoke args determined"
    end

    public
    def execute( chain )
 
        argv = []
        argv << "-classpath" << get_test_classpath( chain ).join( ":" )

        JavaTesting.add_test_run_paths( argv, chain )

        argv += get_test_runner_args( chain, get_test_class_names )
        JavaTesting.add_filter_args( argv, @run_opts )

        UnixProcessBuilder.new( cmd: jcmd( "java" ), argv: argv ).system
    end
end

TaskRegistry.instance.register_path( JavaTestRunner, :java, :test )

class JavaAppRunner < StandardProjTask
    
    include JavaEnvMixin
    include JavaTaskMixin

    public
    def get_direct_dependencies
        
        res = [ 
            TaskTarget.create( :java, :build, proj(), :bin ),
            TaskTarget.create( :java, :build, :application, :bin )
        ]

        if cp_includes_test?
            res << TaskTarget.create( :java, :build, proj(), :test )
        end

        res
    end

    public
    def execute( chain )
        
        argv = []

        argv << "-classpath" << get_run_classpath( chain ).join( ":" )
        argv << "com.bitgirder.application.ApplicationRunner"
        argv += ( @run_ctx[ :argv_remain ] or [] )

        UnixProcessBuilder.new( cmd: jcmd( "java" ), argv: argv ).exec
    end
end

TaskRegistry.instance.register_path( JavaAppRunner, :java, :run_app )

class MingleAppRunner < StandardProjTask
    
    include JavaEnvMixin
    include JavaTaskMixin

    public
    def get_direct_dependencies

        res = [
            TaskTarget.create( :java, :build, :mingle_application, :bin ),
            TaskTarget.create( :java, :build, proj(), :bin )
        ]

        if cp_includes_test?
            res << TaskTarget.create( :java, :build, proj(), :test )
        end

        res
    end

    public
    def execute( chain )
        
        argv = []

        argv << "-classpath" << get_run_classpath( chain ).join( ":" )
        argv << "com.bitgirder.mingle.application.MingleApplicationRunner"
        argv += ( @run_ctx[ :argv_remain ] or raise "No app args given" )

        UnixProcessBuilder.new( cmd: jcmd( "java" ), argv: argv ).exec
    end
end

TaskRegistry.instance.register_path( MingleAppRunner, :java, :run_mingle_app )

class AbstractJavaDistTask < BitGirder::Ops::Build::Distro::AbstractDistroTask

    include JavaEnvMixin 
    include JavaTaskMixin

    public
    def get_jar_dir
        "#{dist_build_dir()}/jar"
    end

    public
    def dist_distrib
        dist_def().fields.expect_string( :distributor )
    end

    public
    def dist_name
        dist_def().fields.expect_string( :name )
    end

    public
    def get_jar_file( mod )

        not_nil( mod, :mod )

        distrib = dist_distrib
        name = dist_name

        "#{get_jar_dir()}/#{distrib}-#{name}-#{mod}.jar"
    end
end

class JavaDistBuilder < AbstractJavaDistTask

    public 
    def get_direct_dependencies

        direct_deps.inject( [] ) do |arr, proj|

            ALL_MODS.inject( arr ) do |arr, mod|

                if ws_ctx.has_mod_src?( proj: proj, mod: mod )
                    arr << TaskTarget.create( :java, :build, proj, mod )
                else
                    arr
                end
            end
        end
    end

    public
    def execute( chain )
        code( "dist built" )
    end
end

TaskRegistry.instance.register_path( JavaDistBuilder, :java, :dist, :build )

class JavaDistJavadoc < AbstractJavaDistTask

    public
    def get_direct_dependencies
        []
    end

    private
    def add_lib_src_args( deps, argv )
        
        deps.each do |proj|
            
            src_dir = mod_src( proj: proj, mod: :lib )

            # Be sure to use the mutable add for argv
            Dir.glob( "#{src_dir}/**/*.java" ).each { |f| argv << f }
        end
    end

    private
    def add_mg_generated( deps, argv )
        
        deps.each do |dep|
            
            dir = 
                gen_src_dir( 
                    proj: dep, mod: :lib, gen_key: MingleCodegen::GEN_KEY )

            Dir.glob( "#{dir}/**/*.java" ).each { |f| argv << f }
        end
    end

    private
    def add_src_args( argv )
        
        deps = Set.new( direct_deps )

        add_lib_src_args( deps, argv )
        add_mg_generated( deps, argv )
    end

    public
    def execute( chain )
        
        argv = []

        dest = ensure_wiped( "#{ws_ctx.proj_build_dir( proj: dist )}/doc/api" )
        argv << "-d" << dest
    
        argv << "-linksource"

        argv << "-linkoffline" <<
                "http://java.sun.com/javase/6/docs/api/" <<
                file_exists( "#{workspace.root}/tools/ruby/javadoc-support" )

        add_src_args( argv )

        unless is_build_dry_run?
            UnixProcessBuilder.new( cmd: jcmd( "javadoc" ), argv: argv ).system
        end
    end
end

TaskRegistry.instance.register_path( JavaDistJavadoc, :java, :dist, :javadoc )

class JavaDistJar < AbstractJavaDistTask
    
    require 'tempfile'
    require 'pathname'
    require 'set'

    AGGREGATED_PATHS = 
        Set.new()
#        Set.new( %w{ 
#            test-runtime.txt
#            mingle-bind-init.txt
#            mingle-runtime-autoload.txt
#            codec-factory-init.txt
#            mingle-http-client-codec-init.txt
#            mingle-http-codec-selector-init.txt
#            com/bitgirder/sql/sql-test-context.properties
#        })

    class JarContext < BitGirder::Core::BitGirderClass
        
        bg_attr :roots, default: proc { [] }
        bg_attr :aggregates, default: proc { [] }
        bg_attr :mod
        bg_attr :mg_out, default: proc { {} }
    end

    public
    def get_direct_dependencies
        [ TaskTarget.create( :java, :dist, :build, dist ) ]
    end

    private
    def aggregated_dir
        "#{dist_build_dir()}/aggregated"
    end

    private
    def wipe_dirs
        [ aggregated_dir(), get_jar_dir() ].each { |d| fu().rm_rf( d ) }
    end

    private
    def get_tasks( chain, jr_ctx, *argv )
        BuildChains.tasks( chain, *argv ).select { |t| t.mod == jr_ctx.mod }
    end

    private
    def init_ctx_map( chain )
        
        BuildChains.tasks( chain, JavaBuilder ).inject( {} ) do |h, t|
            h[ t.mod ] ||= JarContext.new( mod: t.mod ) if t.respond_to?( :mod )
            h
        end
    end

    private
    def add_task_roots( t, jr_ctx )
        
        case t

            when JavaBuilder 
                jr_ctx.roots << t.mod_classes
                if File.exist?( f = t.mod_resource ) then jr_ctx.roots << f end

            when AbstractCodegenTask
                jr_ctx.roots << t.gen_classes_dir if t.compile_generated? 

            else raise "Unexpected task #{t} (#{t.class})"
        end
    end

    private
    def add_jar_roots( chain, ctx_map )
        
        BuildChains.tasks( chain, JavaBuilder, AbstractCodegenTask ).each do |t|
            add_task_roots( t, ctx_map[ t.mod ] )
        end
    end

    private
    def add_mg_out_paths( chain, ctx_map )
        
        ctx_map.values.each do |jr_ctx|
            get_tasks( chain, jr_ctx, MG_OPS::MingleCompile ).each do |t|
                
                out_path = Pathname.new( file_exists( t.get_output_file ) )

                root_path = Pathname.new( "#{out_path}/../../.." )
                root = root_path.cleanpath.to_s

                rel_path = out_path.relative_path_from( root_path )

                ( jr_ctx.mg_out[ root ] ||= [] ) << rel_path.to_s
            end
        end
    end

    private
    def aggregated_path_dir( jr_ctx )
        "#{aggregated_dir()}/#{jr_ctx.mod}"
    end

    private
    def write_aggregated_path( jr_ctx, path )
        
        f = "#{aggregated_path_dir( jr_ctx )}/#{path}"

        File.open( ensure_parent( f ), "w" ) { |io| yield( io ) }

        jr_ctx.aggregates << f
    end

    private
    def get_path_matches( chain, jr_ctx, path )
        
        get_tasks( chain, jr_ctx, JavaBuilder, AbstractCodegenTask ).
            inject( [] ) do |arr, t|
    
                dir =
                    case t
                        when JavaBuilder then t.mod_resource
                        when AbstractCodegenTask then t.gen_classes_dir
                        else raise "Unhandled task type: #{t.class}"
                    end

                File.exist?( f = "#{dir}/#{path}" ) ? ( arr << f ) : arr
            end
    end

    private
    def merge_aggregated_text_path( chain, jr_ctx, path )
 
        inputs = get_path_matches( chain, jr_ctx, path )

        unless inputs.empty?
            write_aggregated_path( jr_ctx, path ) do |dest|
                inputs.each do |f| 
                    File.open( f ) { |src| IO.copy_stream( src, dest ) }
                end
            end
        end
    end

    private
    def build_aggregated_path( chain, jr_ctx, path )
        
#        case path
#
#            when "test-runtime.txt",
#                 "mingle-bind-init.txt",
#                 "mingle-runtime-autoload.txt",
#                 "codec-factory-init.txt",
#                 "mingle-http-client-codec-init.txt",
#                 "mingle-http-codec-selector-init.txt",
#                 "com/bitgirder/sql/sql-test-context.properties"
#                merge_aggregated_text_path( chain, jr_ctx, path )
#
#            else raise "Unhandled aggregated path: #{path}"
#        end

        raise "Unhandled aggregated path: #{path}"
    end

    private
    def build_aggregated_paths( chain, ctx_map )
        
        ctx_map.each_pair do |mod, jr_ctx|
            AGGREGATED_PATHS.each do |path|
                build_aggregated_path( chain, jr_ctx, path )
            end
        end
    end

    private
    def write_file_arg( rel, root, io )
        io.puts( "-C #{root} #{rel}" )
    end

    private
    def write_jar_args( jr_ctx, path_tree, arg_f )
 
        File.open( arg_f, "w" ) do |io|
 
            path_tree.each_pair { |rel, root| write_file_arg( rel, root, io ) }

            if File.exist?( agg_dir = aggregated_path_dir( jr_ctx ) )
                io.puts( "-C #{agg_dir} ." )
            end

            jr_ctx.mg_out.each_pair do |root, arr|
                arr.each { |rel| write_file_arg( rel, root, io ) }
            end
        end  
    end

    private
    def clean_dir_name( root )
        root.gsub( /\/+$/, "" )
    end

    private
    def get_path_tree( jr_ctx )
        
        jr_ctx.roots.inject( {} ) do |tree, root|

            Dir.chdir( root ) do
                Dir.glob( "**/*" ).
                    select { |f| File.file?( f ) }.
                    reject { |f| AGGREGATED_PATHS.include?( f ) }.
                    each do |rel_path|

                        if prev = tree[ rel_path ]
                            raise "File #{rel_path} is in #{root} and #{prev}"
                        else
                            tree[ rel_path ] = clean_dir_name( root )
                        end
                end
            end

            tree
        end
    end

    private
    def write_jar_file( jar_ctx, path_tree, arg_f )
 
        argv = [
            "cf", 
            ensure_parent( get_jar_file( jar_ctx.mod ) ), 
            "@#{arg_f}" 
        ]

        UnixProcessBuilder.new( cmd: jcmd( "jar" ), argv: argv ).system
    end

    private
    def create_jars( ctx_map )
 
        ctx_map.each_pair do |mod, jr_ctx|
        
            path_tree = get_path_tree( jr_ctx )

            Tempfile.open( [ "jar-#{mod}", ".txt" ] ) do |io|

                arg_f = io.path

                write_jar_args( jr_ctx, path_tree, arg_f )
                write_jar_file( jr_ctx, path_tree, arg_f )
            end
        end
    end

    public
    def execute( chain )
 
        wipe_dirs
        ctx_map = init_ctx_map( chain )
        add_jar_roots( chain, ctx_map )
        add_mg_out_paths( chain, ctx_map )
        build_aggregated_paths( chain, ctx_map )
        create_jars( ctx_map )
    end
end

TaskRegistry.instance.register_path( JavaDistJar, :java, :dist, :jar )

class MavenOpts < BitGirderClass
    
    bg_attr :group_id
    bg_attr :artifact_id
    bg_attr :version

    public
    def get_repo_path
        "#{ @group_id.gsub( '.', '/' ) }/#@artifact_id/#@version"
    end
end

class MavenRepoBuilder < AbstractJavaDistTask
    
    private
    def dist_target
        TaskTarget.create( :java, :dist, :jar, dist )
    end

    public
    def get_direct_dependencies
        [ dist_target ]
    end

    public
    def repo_dir
        "#{dist_build_dir}/maven-repo"
    end

    private
    def get_group_id
        
        jv_opts = dist_def.fields.expect_mingle_symbol_map( :java )
        jv_opts.expect_string( :maven_group_id )
    end

    public
    def get_maven_opts
        
        MavenOpts.new(
            :group_id => get_group_id,
            :artifact_id => "#{dist_distrib}-#{dist_name}",
            :version => BuildVersions.get_version( run_opts: @run_opts )
        )
    end

    private
    def get_name_prefix( mvn_opts )
        "#{mvn_opts.artifact_id}-#{mvn_opts.version}"
    end

    private
    def write_digest_file( src, dig_file, dig_cls )
        
        dig = dig_cls.new

        buf_sz = 2 ** 20

        File.open( src ) do |io|

            while str = io.read( buf_sz ) do
                dig.update( str )
            end
        end

        dig_hex = Digest.hexencode( dig.digest )

        File.open( dig_file, "w" ) { |io| io.print( dig_hex ) }
    end

    private
    def write_digests( file )
        
        write_digest_file( file, "#{file}.md5", Digest::MD5 )
        write_digest_file( file, "#{file}.sha1", Digest::SHA1 )
    end

    private
    def write_pom( pom_file, mvn_opts )
        
        pom = <<-POM
<project xmlns="http://maven.apache.org/POM/4.0.0" 
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/maven-v4_0_0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <groupId>#{mvn_opts.group_id}</groupId>
    <version>#{mvn_opts.version}</version>
    <artifactId>#{mvn_opts.artifact_id}</artifactId>
    <packaging>jar</packaging>
</project>
        POM

        File.open( pom_file, "w" ) { |io| io.print pom }

        pom_file
    end

    # Impl note: we copy lib_jar --> lib_jar_targ instead of symlinking it only
    # because MavenRepoBuilder is using (as of this writing) an external tool
    # (s3cmd) which doesn't seem to deal well with symlinks. Once we switch
    # tools, a symlink would be preferred here.
    public
    def execute( chain )
 
        lib_jar = get_jar_file( :lib )

        ensure_wiped( repo_dir )
        mvn_opts = get_maven_opts
        name_pref = get_name_prefix( mvn_opts )

        lib_jar_targ = "#{repo_dir}/#{name_pref}.jar"
        fu().cp( lib_jar, lib_jar_targ )
        write_digests( lib_jar_targ )

        pom = write_pom( "#{repo_dir}/#{name_pref}.pom", mvn_opts )
        write_digests( pom )
    end
end

TaskRegistry.instance.
    register_path( MavenRepoBuilder, :java, :dist, :maven_repo )

class MavenRepoUploader < AbstractJavaDistTask

    private
    def repo_task_target
        TaskTarget.create( :java, :dist, :maven_repo, dist )
    end

    public
    def get_direct_dependencies
        [ repo_task_target ]
    end

    private
    def get_local_files( repo_dir )
        Dir.glob( "#{repo_dir}/*" )
    end

    private
    def get_remote_root( tsk )
        
        res = @run_opts.expect_string( :remote_prefix )
        res << "/" << tsk.get_maven_opts.get_repo_path

        "s3://" + res.gsub( %r{/+}, "/" ) + "/"
    end

    private
    def exec_upload( tsk )
        
        cmd = file_exists( @run_opts.expect_string( :s3cmd ) )

        argv = []
        argv << "-c" << @run_opts.expect_string( :s3cmd_cfg )
        argv << "--acl-public"
        argv << "put" 
        argv += get_local_files( tsk.repo_dir )
        argv << get_remote_root( tsk )
        
        UnixProcessBuilder.new( cmd: cmd, argv: argv ).system
    end

    public
    def execute( chain )
        
        tsk = BuildChains.expect_task( chain, repo_task_target )
        exec_upload( tsk )
    end
end

TaskRegistry.instance.
    register_path( MavenRepoUploader, :java, :dist, :upload_maven_repo )

class JavaDistTestRunner < AbstractJavaDistTask

    public 
    def get_direct_dependencies
        [ TaskTarget.create( :java, :dist, :jar, dist ) ]
    end

    public
    def get_closure_dependencies( tasks )
        
        tasks.inject( [] ) do |arr, t| 
            
            if t.is_a?( JavaBuilder ) && t.mod?( :test )
                arr += t.get_declared_test_deps
            end

            arr
        end
    end

    private
    def get_test_class_names( chain )
        
        res = []

        argv = [ "tf", get_jar_file( :test ) ]
        cmd = jcmd( "jar" )

        UnixProcessBuilder.new( cmd: cmd, argv: argv ).popen( "r" ) do |io|

            io.each_line do |l|
                if /^(?<nm>[^\$]+)\.class$/ =~ l
                    res << nm.gsub( /\//, '.' )
                end
            end
        end

        res
    end

    private
    def get_test_classpath( chain )
 
        res = JavaTesting.get_test_cp_extra( @run_opts )

        BuildChains.tasks( chain, MavenPullTask ) do |t| 
            res << t.read_classpath
        end

        res += 
            ALL_MODS.map { |mod| get_jar_file( mod ) }.
                     select { |f| File.exist?( f ) }

        res
    end

    private
    def set_log_opts( opts )

        unless @run_opts.get_boolean( :log_to_console )
        
            run_log = ensure_parent( "#{dist_build_dir}/test/log/run.log" )
            console( "Sending test output to #{run_log}" )
            opts[ [ :out, :err ] ] = [ run_log, "w" ]
        end
    end

    public
    def execute( chain )

        argv = []

        argv += JavaTesting.get_test_jvm_args( @run_ctx )
        argv << "-classpath" << get_test_classpath( chain ).join( ":" )
        JavaTesting.add_test_run_paths( argv, chain )

        nms = get_test_class_names( chain )
        argv += JavaTesting.get_testing_test_runner_args( nms, @run_ctx )

        opts = {}
        set_log_opts( opts )

        UnixProcessBuilder.
            new( cmd: jcmd( "java" ), argv: argv, opts: opts ).
            system
    end
end

TaskRegistry.instance.register_path( JavaDistTestRunner, :java, :dist, :test )

class JavaDistIntegRunner < AbstractJavaDistTask
    
    public
    def get_direct_dependencies
        []
    end

    public
    def execute( chain )
    end
end

TaskRegistry.instance.register_path( JavaDistIntegRunner, :java, :dist, :integ )

class CleanTask < StandardProjTask
    
    public
    def get_direct_dependencies
        []
    end

    public
    def execute( chain )
        
        to_del = [ ws_ctx.proj_build_dir ]

        inf_pref = "#{workspace.build_dir}/.build-info/java"
        to_del << "#{inf_pref}/build/#{proj}"
        to_del += Dir.glob( "#{inf_pref}/codegen/*/#{proj}" )

        to_del.each { |dir| fu().rm_rf( dir ) if File.exist?( dir ) }
    end
end

TaskRegistry.instance.register_path( CleanTask, :java, :clean )

end
end
end
end
