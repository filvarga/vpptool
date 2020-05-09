/*
 * Copyright 2020 Filip Varga
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

// TODO:
//		 production/development builds
//		 building over buildkit
//		 remote building
// TODO:
//		 consider developing inside the container all the time ...
//		 - remote building colides wit this ?
//		 - or if we have remote docker containers with all vim data
//			 could this make it even better in the sense of being able
//			 to connect from anywhere and develop righ in docker container

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

const (
	context       = "https://github.com/filvarga/vpptool.git#master:docker"
	tmp_container = "vpptool-container"
	vpp_name      = "vpp-run"
	vpp_image     = "vpptool-images"
	setup_tag     = "setup"
	build_tag     = "build"
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
	src          string
	plugin       string
	get_commit   bool
	commit       string
	debug        bool
}

func log(w io.Writer, format string, args ...interface{}) (i int, err error) {
	return fmt.Fprintf(w, "%s: %s\n", os.Args[0], fmt.Sprintf(format, args...))
}

func logDebug(format string, args ...interface{}) (i int, err error) {
	return fmt.Fprintf(os.Stderr, "%s (debug):\n%s", os.Args[0],
		fmt.Sprintf(format, args...))
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
	return run(false, "docker", "rm", "-f", name)
}

func (t tool) check_image(img image) bool {
	return run(t.debug, "docker", "image", "inspect", "--format='.'",
		fmt.Sprintf("%s:%s", img.vpp_image, img.vpp_tag))
}

func (t tool) build_setup_image() bool {
	return run(t.debug, "docker", "build",
		"-t", fmt.Sprintf("%s:%s", t.setup.vpp_image, t.setup.vpp_tag),
		t.context)
}

func (t tool) build_cache_image(name string, script string, src image, dst image) bool {
	var success bool

	del_container(name)
	defer del_container(name)

	success = run(t.debug, "docker", "run", "--name", name,
		"-e", fmt.Sprintf("CID=%s", t.commit),
		fmt.Sprintf("%s:%s", src.vpp_image, src.vpp_tag), script)

	if !success {
		return false
	}

	success = run(t.debug, "docker", "commit", name,
		fmt.Sprintf("%s:%s", dst.vpp_image, dst.vpp_tag))

	return success
}

func (t tool) deploy_vpp(name string) bool {

	del_container(name)

	if len(t.src) > 0 {
		return run(t.debug, "docker", "run", "-it", "--cap-add=all", "--privileged",
			"-d", "--network", "host", "--name", name, "-v",
			fmt.Sprintf("%s:/opt/vpp/src", t.src),
			fmt.Sprintf("%s:%s", t.build.vpp_image, t.build.vpp_tag))
	} else if len(t.plugin) > 0 {
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

func print_usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s: <build|<deploy [vpp-run]>>\n",
		os.Args[0])
	fmt.Fprintf(os.Stderr, "build & deploy\n")
	flag.PrintDefaults()
}

func main() {

	var success bool
	var src, dst image

	t := tool{
		setup: image{
			vpp_image: vpp_image,
			vpp_tag:   setup_tag,
		},
	}

	// TODO: don't log output of build commands rather have better info logging
	flag.BoolVar(&t.debug, "debug", false, "print debug output")

	// for updating cache / setup images
	setup := flag.Bool("setup", false, "rebuild setup image")
	cache := flag.Bool("cache", false, "rebuild cache image")

	// required for setup phase
	flag.StringVar(&t.commit, "commit-id", "", "commit id")
	// will be unused when we switch to full workstation scenario
	flag.BoolVar(&t.get_commit, "commit-get", false,
		"use current dir commit id")

	// required for install phase (for building base docker image)
	flag.StringVar(&t.context, "context", context, "setup docker context url")

	flag.StringVar(&t.build.vpp_image, "image", vpp_image, "build docker image")
	// only the final image should be tagged
	flag.StringVar(&t.build.vpp_tag, "tag", build_tag, "build docker tag")

	// mounts over container src ./vpp/src
	flag.StringVar(&t.src, "src", "", "src folder")
	// mounts over container src ./vpp/src/plugins/<plugin>
	flag.StringVar(&t.plugin, "plugin", "", "custom plugin folder")

	flag.Parse()

	switch flag.Arg(0) {
	default:
		print_usage()
	case "deploy":
		if flag.NArg() < 2 {
			success = t.deploy_vpp(vpp_name)
		} else {
			success = t.deploy_vpp(flag.Arg(1))
		}
		if !success {
			os.Exit(1)
		}
	case "build":
		if !t.check_image(t.setup) || *setup {
			logInfo("building setup image...")
			success = t.build_setup_image()
			if !success {
				logError("error building image")
				os.Exit(1)
			}
		}

		if len(t.commit) <= 0 && t.get_commit {
			t.commit = get_commit_id()
			logInfo("using commit-id: %s", t.commit)
		}

		src = t.build
		dst = t.build

		if !t.check_image(t.build) || *cache {
			logInfo("building cache image...")
			src = t.setup
		}

		// consider adding additional files as patches etc. in this stage

		success = t.build_cache_image(tmp_container, "/scripts/cache", src, dst)
		if !success {
			logError("error caching dependencies")
			os.Exit(1)
		}

		success = t.build_cache_image(tmp_container, "/scripts/build", dst, dst)
		if !success {
			logError("error building")
			os.Exit(1)
		}
	}
	os.Exit(0)
}

/* vim: set ts=2: */
