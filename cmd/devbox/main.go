/*
 * Copyright 2021 Filip Varga
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

const (
	tmp_container = "devbox-container"
	box_name      = "devbox"
	box_image     = "devbox-images"
	setup_tag     = "setup"
	build_tag     = "build"
	version       = "3.9.2"
)

var (
	context = ""
)

type image struct {
	box_image string
	box_tag   string
}

type prog struct {
	setup   image
	build   image
	run     image
	context string
	plugin  string
	version string
	quiet   bool
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

func get_version_id_v2() (bool, string) {
	cmd := exec.Command("curl", "-fsSLI", "-o", "/dev/null", "-w", "'%{url_effective}'",
		"https://github.com/cdr/code-server/releases/latest")
	return run_command_v2(*cmd)
}

func get_version_id() string {
	var (
		status bool
		output string
	)
	// TODO: should be regex tested rather than this
	status, output = get_version_id_v2()

	if !status || len(output) < 49 {
		fmt.Fprintf(os.Stderr, "%s: unable to get latest code-server version\n",
			os.Args[0])
		output = version
	} else {
		output = output[49:]
	}
	return output
}

func del_container(name string) bool {
	return run(true, "docker", "rm", "-f", name)
}

func (p prog) check_image(img image) bool {
	return run(true, "docker", "image", "inspect", "--format='.'",
		fmt.Sprintf("%s:%s", img.box_image, img.box_tag))
}

func (p prog) build_setup_image() bool {
	// TODO: add IDU, IDG, GIT_MAIL, GIT_NAME, VERSION, SUDO_PASS, CODE_PASS
	p.version = get_version_id()

	// TODO: checkout arguments for the build process
	return run(p.quiet, "docker", "build",
		"-t", fmt.Sprintf("%s:%s", p.setup.box_image, p.setup.box_tag),
		p.context)
}

func (p prog) build_cache_image(name string, script string, src image, dst image) bool {
	var success bool

	del_container(name)
	defer del_container(name)

	success = run(p.quiet, "docker", "run", "--name", name,
		"-e", fmt.Sprintf("CID=%s", p.commit),
		fmt.Sprintf("%s:%s", src.box_image, src.box_tag), script)

	if !success {
		return false
	}

	// TODO: this is unnecessary
	success = run(p.quiet, "docker", "commit",
		"--change=cmd [\"/scripts/start\"]", name,
		fmt.Sprintf("%s:%s", dst.box_image, dst.box_tag))

	return success
}

func (p prog) deploy_box(name string) bool {

	del_container(name)

	// TODO: install golang binaries
	// TODO: fill environment variables

	if len(p.mount) > 0 {
		return run(p.quiet, "docker", "run", "-it", "--cap-add=all", "--privileged",
			"-e", fmt.Sprintf("START_VPP=%d"),
			"-d", "--network", "host", "--name", name, "-v",
			fmt.Sprintf("%s:/opt/user/host", p.mount),
			fmt.Sprintf("%s:%s", p.build.box_image, p.build.box_tag))
	} else {
		return run(p.quiet, "docker", "run", "-it", "--cap-add=all", "--privileged",
			"-e", fmt.Sprintf("START_VPP=%d"),
			"-d", "--network", "host", "--name", name,
			fmt.Sprintf("%s:%s", p.build.box_image, p.build.box_tag))
	}
}

func print_usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s: <build|<deploy [box-run]>>\n",
		os.Args[0])
	fmt.Fprintf(os.Stderr, "build & deploy\n")
	flag.PrintDefaults()
}

func main() {

	var success bool
	var src, dst image

	t := prog{
		setup: image{
			box_image: box_image,
			box_tag:   setup_tag,
		},
	}

	flag.BoolVar(&p.quiet, "quiet", false, "run quietly")

	// for updating cache / setup images
	setup := flag.Bool("setup", false, "rebuild setup image using context url")
	cache := flag.Bool("cache", false, "rebuild cache image")

	// required for install phase (for building base docker image)
	flag.StringVar(&p.context, "context", context, "setup docker context url")

	flag.StringVar(&p.build.box_image, "image", box_image, "build docker image")
	// only the final image should be tagged
	flag.StringVar(&p.build.box_tag, "tag", build_tag, "build docker tag")
	flag.StringVar(&p.mount, "mount", "", "bind mount")
	flag.Parse()

	switch flag.Arg(0) {
	default:
		print_usage()
	case "deploy":
		if flag.NArg() < 2 {
			success = p.deploy_box(box_name)
		} else {
			success = p.deploy_box(flag.Arg(1))
		}
		if !success {
			os.Exit(1)
		}
	case "build":
		src = p.build
		dst = p.build

		// if setup also rebuild cache
		if !p.check_image(p.setup) || *setup {
			logInfo("building setup image...")
			success = p.build_setup_image()
			if !success {
				logError("error building image")
				os.Exit(1)
			}
			src = p.setup
		}

		if !p.check_image(p.build) || *cache {
			logInfo("building cache image...")
			src = p.setup
		}

		// TODO: this should be reconsidered what we want and what we don't want
		// consider adding additional files as patches etc. in this stage
		success = p.build_cache_image(tmp_container, "/home/user/bin/project-init", src, dst)
		if !success {
			logError("error updating image")
			os.Exit(1)
		}
	}
	os.Exit(0)
}

/* vim: set ts=2: */
