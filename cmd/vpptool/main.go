/*
 * Copyright 2021 Filip Varga
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

const (
	tmp_container = "vpptool-container"
	vpp_name      = "vpp-run"
	vpp_image     = "vpptool-images"
	setup_tag     = "setup"
	build_tag     = "master"
	code_pswd     = "toor"
	git_mail      = "john.doe@example.com"
	git_name      = "John Doe"
	idu           = 1000
	idg           = 1000
)

var (
	context    = ""
	git_url    = ""
	cs_version = ""
	go_version = ""
)

type image struct {
	vpp_image string
	vpp_tag   string
}

type tool struct {
	setup        image
	build        image
	run          image
	startup_file string
	config_file  string
	context      string
	plugin       string
	start_vpp    bool
	get_commit   bool
	commit       string
	quiet        bool
	git_mail     string
	git_name     string
	idu          int
	idg          int
	cs_version   string
	go_version   string
	user_pswd    string
	code_pswd    string
}

func log(w io.Writer, format string, args ...interface{}) (i int, err error) {
	return fmt.Fprintf(w, "%s: %s\n", os.Args[0], fmt.Sprintf(format, args...))
}

func logInfo(format string, args ...interface{}) {
	log(os.Stdout, format, args...)
}

func logError(format string, args ...interface{}) {
	log(os.Stderr, format, args...)
}

func run_command(cmd exec.Cmd) bool {
	var status = false
	err := cmd.Run()
	if err == nil {
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		if ws.ExitStatus() == 0 {
			status = true
		}
	}
	return status
}

func run_command_v2(cmd exec.Cmd) (bool, string) {
	var status = false
	output, err := cmd.CombinedOutput()
	if err == nil {
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		if ws.ExitStatus() == 0 {
			status = true
		}
	}
	return status, string(output)
}

func run(quiet bool, command string, args ...string) bool {
	cmd := exec.Command(command, args...)
	if quiet == false {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return run_command(*cmd)
}

func get_current_commit_id() (bool, string) {
	cmd := exec.Command("git", "show", "--pretty=format:\"%H\"", "--no-patch")
	return run_command_v2(*cmd)
}

func get_commit_id() string {
	var (
		status bool
		output string
	)
	status, output = get_current_commit_id()

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
	return run(true, "docker", "rm", "-f", name)
}

func (t tool) check_image(img image) bool {
	return run(true, "docker", "image", "inspect", "--format='.'",
		fmt.Sprintf("%s:%s", img.vpp_image, img.vpp_tag))
}

func (t tool) build_base_image() bool {
	return run(t.quiet, "docker", "build",
		"--build-arg", fmt.Sprintf("USER_PSWD:%s", t.user_pswd),
		"--build-arg", fmt.Sprintf("GIT_MAIL:%s", t.git_mail),
		"--build-arg", fmt.Sprintf("GIT_NAME:%s", t.git_name),
		"--build-arg", fmt.Sprintf("GIT_URL:%s", git_url),
		"--build-arg", fmt.Sprintf("IDU:%d", t.idu),
		"--build-arg", fmt.Sprintf("IDG:%d", t.idg),
		"--target", "base",
		"-t", fmt.Sprintf("%s:%s", t.setup.vpp_image, t.setup.vpp_tag),
		t.context)
}

func (t tool) build_tool_image() bool {
	hash := sha256.Sum256([]byte(t.code_pswd))
	// hash=$(echo -n $plain | sha256sum | cut -c -64 -z)
	return run(t.quiet, "docker", "build",
		"--build-arg", fmt.Sprintf("CS_VERSION:%s", cs_version),
		"--build-arg", fmt.Sprintf("GO_VERSION:%s", go_version),
		"--build-arg", fmt.Sprintf("CODE_PSWD:%x", hash[:]),
		"--target", "tool",
		"-t", fmt.Sprintf("%s:%s", t.setup.vpp_image, t.setup.vpp_tag),
		t.context)
}

func (t tool) cache_base_image(name string, script string, src image, dst image) bool {
	var success bool

	del_container(name)
	defer del_container(name)

	success = run(t.quiet, "docker", "run", "--name", name,
		"-e", fmt.Sprintf("CID=%s", t.commit),
		fmt.Sprintf("%s:%s", src.vpp_image, src.vpp_tag), script)

	if !success {
		return false
	}

	success = run(t.quiet, "docker", "commit",
		"--change=cmd [\"/usr/local/bin/entrypoint\"]", name,
		fmt.Sprintf("%s:%s", dst.vpp_image, dst.vpp_tag))

	return success
}

func (t tool) deploy_base(name string) bool {
	var start_vpp int8
	if t.start_vpp {
		start_vpp = 1
	}

	del_container(name)

	if len(t.plugin) > 0 {
		return run(t.quiet, "docker", "run", "-it", "--cap-add=all", "--privileged",
			"-e", fmt.Sprintf("START_VPP=%d", start_vpp),
			"-d", "--network", "host", "--name", name, "-v",
			fmt.Sprintf("%s:/work/vpp/src/plugins/%s", t.plugin, filepath.Base(t.plugin)),
			fmt.Sprintf("%s:%s", t.build.vpp_image, t.build.vpp_tag))
	} else {
		return run(t.quiet, "docker", "run", "-it", "--cap-add=all", "--privileged",
			"-e", fmt.Sprintf("START_VPP=%d", start_vpp),
			"-d", "--network", "host", "--name", name,
			fmt.Sprintf("%s:%s", t.build.vpp_image, t.build.vpp_tag))
	}
}

func print_usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s: <build|deploy> <vpp|env>\n",
		os.Args[0])
	flag.PrintDefaults()
}

func notifySend(persistent bool, message string) {
	if persistent == false {
		run(true, "notify-send", "--icon", "applications-utilities",
			"--urgency=normal", "vpptool", message)
	} else {
		run(true, "notify-send", "--icon", "applications-utilities",
			"--urgency=critical", "vpptool", message)
	}
}

func paplay() {
	run(true, "paplay", "/usr/share/sounds/freedesktop/stereo/alarm-clock-elapsed.oga")
}

func exitFailure(message string) {
	notifySend(true, message)
	logError(message)
	paplay()
	os.Exit(1)
}

func (t tool) build_base(setup bool, cache bool) {

	var success bool
	var src, dst image

	src = t.build
	dst = t.build

	// if setup also rebuild cache
	if !t.check_image(t.setup) || setup {
		logInfo("building setup image...")

		success = t.build_base_image()
		if !success {
			exitFailure("error building setup image")
		}
		src = t.setup
	}

	if len(t.commit) <= 0 && t.get_commit {
		t.commit = get_commit_id()
		logInfo("using commit-id: %s", t.commit)
	}

	if !t.check_image(t.build) || cache {
		logInfo("rebuilding cache...")
		src = t.setup
	}

	success = t.cache_base_image(tmp_container,
		"/usr/local/bin/stage1", src, dst)
	if !success {
		exitFailure("error cache stage one")
	} else {
		notifySend(false, "cache stage one")
	}

	success = t.cache_base_image(tmp_container,
		"/usr/local/bin/stage2", dst, dst)
	if !success {
		exitFailure("error cache stage two")
	} else {
		notifySend(true, "cache stage two")
	}
}

func (t tool) build_tool() {

}

func (t tool) deploy_tool(name string) bool {

	return true
}

// TODO: populate arguments in different calls, flag populate or so
// based on the process and the requirements like base / tool store
// those arguments in two separate data structures

func main() {

	success := true

	t := tool{
		setup: image{
			vpp_image: vpp_image,
			vpp_tag:   setup_tag,
		},
	}

	flag.BoolVar(&t.quiet, "quiet", false, "run quietly")

	setup := flag.Bool("setup", false, "rebuild setup image using context url")
	cache := flag.Bool("cache", false, "rebuild cache image")

	name := flag.String("name", vpp_name, "container name")

	// required for setup phase
	flag.StringVar(&t.commit, "commit-id", "", "commit id")
	// will be unused when we switch to full workstation scenario
	flag.BoolVar(&t.get_commit, "commit-get", false,
		"use current dir commit id")

	// required for install phase (for building base docker image)
	flag.StringVar(&t.context, "context", context, "setup docker context url")

	flag.StringVar(&t.git_mail, "git-user-mail", git_mail, "git user mail")
	flag.StringVar(&t.git_name, "git-user-name", git_name, "git user name")

	flag.IntVar(&t.idu, "uid", idu, "user id")
	flag.IntVar(&t.idg, "gid", idg, "user id")

	flag.StringVar(&t.code_pswd, "code-pswd", code_pswd, "code-server password")
	flag.StringVar(&t.user_pswd, "user-pswd", "", "user password")

	// TODO: consider only the final image being tagged
	flag.StringVar(&t.build.vpp_image, "image", vpp_image, "build docker image")
	flag.StringVar(&t.build.vpp_tag, "tag", build_tag, "build docker tag")

	// mounting ./vpp/src does not work (cmake issues preventing building)
	// mounts in container ./vpp/src/plugins/<plugin>
	flag.StringVar(&t.plugin, "plugin", "", "custom plugin folder")

	// will be unused when we switch to full workstation scenario
	flag.BoolVar(&t.start_vpp, "running", false,
		"vpp is running in deployed container")

	flag.Parse()

	switch flag.Arg(0) {
	default:
		print_usage()
	case "deploy":
		switch flag.Arg(1) {
		default:
			print_usage()
		case "vpp":
			success = t.deploy_base(*name)
		case "env":
			success = t.deploy_tool(*name)
		}
	case "build":
		switch flag.Arg(1) {
		default:
			print_usage()
		case "vpp":
			t.build_base(*setup, *cache)
		case "env":
			t.build_tool()
		}
	}
	if !success {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

/* vim: set ts=2: */
