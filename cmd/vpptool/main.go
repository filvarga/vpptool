package main

/*
TODO:

	1) enable setup & build stage different tags


	default paths, for example ~/vpp/
	container build - placed on dockerhub already ?
	mounts - should be optional and configurable
*/

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"syscall"
)

const (
	context       = "git@git.server:/git-server/repos/vpptool.git#master:docker"
	tmp_container = "vpptool-container"
	vpp_image     = "vpptool-images"
	vpp_setup_tag = "setup"
	vpp_build_tag = "build"
)

type image struct {
	vpp_image string
	vpp_tag   string
}

type tool struct {
	setup          image
	build          image
	run            image
	startup_file   string
	config_file    string
	context        string
	src            string
	plugin         string
	current_commit bool
	second_commit  bool
	commit         string
	debug          bool
}

func log(w io.Writer, format string, args ...interface{}) (i int, err error) {
	return fmt.Fprintf(w, "%s: %s\n", os.Args[0], fmt.Sprintf(format, args...))
}

func logDebug(format string, args ...interface{}) (i int, err error) {
	return fmt.Fprintf(os.Stderr, "%s (debug):\n%s", os.Args[0], fmt.Sprintf(format, args...))
}

func logInfo(format string, args ...interface{}) {
	log(os.Stdout, format, args...)
}

func logError(format string, args ...interface{}) {
	log(os.Stderr, format, args...)
}

func run_command(command string, args ...string) (bool, string) {

	cmd := exec.Command(command, args...)

	dir, err := os.Getwd()
	if err == nil {
		cmd.Dir = dir
	}

	var status = true
	var output []byte
	var ws syscall.WaitStatus

	output, err = cmd.CombinedOutput()
	if err != nil {
		status = false
		goto done
	}

	ws = cmd.ProcessState.Sys().(syscall.WaitStatus)
	if ws.ExitStatus() != 0 {
		status = false
		goto done
	}

done:
	return status, string(output)
}

func run(debug bool, command string, args ...string) bool {
	status, out := run_command(command, args...)
	if debug {
		logDebug(out)
	}
	return status
}

func get_current_commit_id() (bool, string) {
	return run_command("git", "show", "--pretty=format:\"%H\"", "--no-patch")
}

func get_previous_commit_id() (bool, string) {
	return run_command("git", "show", "HEAD^", "--pretty=format:\"%H\"", "--no-patch")
}

func get_commit_id(current bool) string {
	var (
		status bool
		output string
	)
	if current {
		status, output = get_current_commit_id()
	} else {
		status, output = get_previous_commit_id()
	}
	if !status {
		fmt.Fprintf(os.Stderr, "%s: not in a git repository\n",
			os.Args[0])
	}
	if len(output) > 0 && output[0] == '"' {
		output = output[1:]
	}
	if len(output) > 0 && output[len(output)-1] == '"' {
		output = output[:len(output)-1]
	}
	return output
}

func del_container(name string) bool {
	return run(false, "docker", "rm", "-f", name)
}

func (t tool) install_image(online bool) bool {
	var success = false

	user, err := user.Current()
	if err != nil {
		logError(err.Error())
		goto done
	}

	// make pull optional
	// --pull
	if online {
		success = run(t.debug, "docker", "build", "--pull",
			"--build-arg", fmt.Sprintf("IDU=%s", user.Uid),
			"--build-arg", fmt.Sprintf("IDG=%s", user.Gid),
			"-t", fmt.Sprintf("%s:%s", t.setup.vpp_image, t.setup.vpp_tag),
			t.context)

	} else {
		success = run(t.debug, "docker", "build",
			"--build-arg", fmt.Sprintf("IDU=%s", user.Uid),
			"--build-arg", fmt.Sprintf("IDG=%s", user.Gid),
			"-t", fmt.Sprintf("%s:%s", t.setup.vpp_image, t.setup.vpp_tag),
			t.context)
	}

done:
	return success
}

func (t tool) setup_image(name string, src image, dst image) bool {

	del_container(name)
	defer del_container(name)

	success := run(t.debug, "docker", "run", "--name", name,
		"-e", fmt.Sprintf("CID=%s", t.commit),
		fmt.Sprintf("%s:%s", src.vpp_image, src.vpp_tag)) &&
		run(t.debug, "docker", "commit", name,
			fmt.Sprintf("%s:%s", dst.vpp_image, dst.vpp_tag))

	return success
}

func (t tool) build_image(name string, src image, dst image) bool {
	var success bool

	del_container(name)
	defer del_container(name)

	if len(t.src) > 0 {
		success = run(t.debug, "docker", "run", "--name", name, "-v",
			fmt.Sprintf("%s:/opt/vpp/src", t.plugin),
			fmt.Sprintf("%s:%s", src.vpp_image, src.vpp_tag), "make", "build")
	} else if len(t.plugin) > 0 {
		success = run(t.debug, "docker", "run", "--name", name, "-v",
			fmt.Sprintf("%s:/opt/vpp/src/plugins/%s", t.plugin, filepath.Base(t.plugin)),
			fmt.Sprintf("%s:%s", src.vpp_image, src.vpp_tag), "make", "build")
	} else {
		success = run(t.debug, "docker", "run", "--name", name,
			fmt.Sprintf("%s:%s", src.vpp_image, src.vpp_tag), "make", "build")
	}

	if !success {
		return false
	}

	success = run(t.debug, "docker", "commit",
		"--change=cmd [\"/scripts/startup\"]", name,
		fmt.Sprintf("%s:%s", dst.vpp_image, dst.vpp_tag))

	return success
}

func (t tool) deploy_vpp(name string) bool {

	del_container(name)

	if len(t.plugin) > 0 {
		return run(t.debug, "docker", "run", "-it", "--cap-add=all", "--privileged",
			"-d", "--network", "host", "--name", name, "-v",
			fmt.Sprintf("%s:/opt/vpp/src/plugins/%s", t.plugin, filepath.Base(t.plugin)),
			fmt.Sprintf("%s:%s", t.build.vpp_image, t.build.vpp_tag))
	} else {
		return run(t.debug, "docker", "run", "-it", "--cap-add=all", "--privileged",
			"-d", "--network", "host", "--name", name,
			fmt.Sprintf("%s:%s", t.build.vpp_image, t.build.vpp_tag))
	}
}

func (t tool) copy_to_etc(name string, src string, dst string) bool {
	return run(t.debug, "docker", "cp", src,
		fmt.Sprintf("%s:/etc/%s", name, dst))
}

func (t tool) try_configure_vpp(name string) {
	if len(t.startup_file) > 0 {
		t.copy_to_etc(name, t.startup_file, "startup.conf")
	}
	if len(t.config_file) > 0 {
		t.copy_to_etc(name, t.config_file, "vpp.cnf")
	}
}

func print_usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s: <install|setup|build|<deploy|configure <name>>\n",
		os.Args[0])
	fmt.Fprintf(os.Stderr, "\t1) install (required once)"+
		"\n\t2) setup (requirements image)"+
		"\n\t3) build (debug vpp image)\n")
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {

	t := tool{
		setup: image{
			vpp_image: vpp_image,
			vpp_tag:   vpp_setup_tag,
		},
	}

	flag.BoolVar(&t.debug, "debug", false, "print debug output")
	online := flag.Bool("online", false, "force update base docker image")

	flag.StringVar(&t.startup_file, "vpp-startup", "", "vpp startup file")
	flag.StringVar(&t.config_file, "vpp-config", "", "vpp config file")

	flag.BoolVar(&t.current_commit, "commit-get", false, "use current dir commit id")
	flag.BoolVar(&t.second_commit, "commit-get-sec", false, "use current dir second commit id")

	flag.StringVar(&t.commit, "commit-id", "", "commit id")

	flag.StringVar(&t.context, "context", context, "setup docker context url")

	flag.StringVar(&t.build.vpp_image, "docker-image", vpp_image, "build docker image")
	flag.StringVar(&t.build.vpp_tag, "build-tag", vpp_build_tag, "build docker tag")

	// TODO: add support for multiple plugins/mounts
	flag.StringVar(&t.src, "src", "", "src folder")
	flag.StringVar(&t.plugin, "plugin", "", "plugin folder")

	flag.Parse()

	if flag.NArg() < 1 {
		print_usage()
	}

	var success = true

	switch flag.Arg(0) {
	case "install":
		success = t.install_image(*online)
	case "setup":
		if len(t.commit) <= 0 {
			if t.second_commit {
				t.commit = get_commit_id(false)
			} else if t.current_commit {
				t.commit = get_commit_id(true)
			}
		}
		if len(t.commit) > 0 {
			logInfo("using commit-id: %s", t.commit)
		}
		success = t.setup_image(tmp_container, t.setup, t.build)
		// TODO:
		// hack to try build second time
		// while ! success or count of times
		if !success {
			t.setup_image(tmp_container, t.setup, t.build)
		}
	case "build":
		success = t.build_image(tmp_container, t.build, t.build)
	case "deploy":
		if flag.NArg() < 2 {
			print_usage()
		}
		name := flag.Arg(1)
		success = t.deploy_vpp(name)
		t.try_configure_vpp(name)
	case "configure":
		if flag.NArg() < 2 {
			print_usage()
		}
		name := flag.Arg(1)
		t.try_configure_vpp(name)
	default:
		print_usage()
	}

	if success {
		os.Exit(0)
	}
	os.Exit(1)
}

/* vim: set ts=2: */
