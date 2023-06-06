package main

import (
	"bufio"
	"context"
	"path/filepath"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

// Bind mounts require absolute paths for the source directory.
func makeAbsolute(rel_path string) string {
	abs_path, err := filepath.Abs(rel_path)
	if err != nil {
		panic(err)
	}
	return abs_path
}

// Create a container, start it, wait for it to stop running, and remove it.
func runContainer(cli *client.Client, ctx context.Context, args []string) {
	pullImage(cli, ctx)

	id := createContainer(cli, ctx, args)

	start_opts := types.ContainerStartOptions{
	}
	cli.ContainerStart(ctx, id, start_opts)

	watchContainer(cli, ctx, id)

	rm_opts := types.ContainerRemoveOptions{
		Force: true,
	}
	cli.ContainerRemove(ctx, id, rm_opts)
}

func createContainer(cli *client.Client, ctx context.Context, args []string) string {
	conf := container.Config{
		Image: "alpine:latest",
		Cmd: args,
	}

	con_conf := container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type: mount.TypeBind,
				Source: makeAbsolute("dir1"),
				Target: "/dir1",
				ReadOnly: true,
			},
			{
				Type: mount.TypeBind,
				Source: makeAbsolute("dir2"),
				Target: "/dir2",
				ReadOnly: false,
			},
		},
	}

	net_conf := network.NetworkingConfig{
	}

	plats := specs.Platform{
		Architecture: "amd64", //"arm64"
		OS: "linux",
	}

	con, err := cli.ContainerCreate(ctx, &conf, &con_conf, &net_conf, &plats, "")
	if err != nil {
		panic(err)
	}

	return con.ID
}

// Pull an image from DockerHub. We aren't particularly concerned about the
// output so it's thrown away.
func pullImage(cli *client.Client, ctx context.Context) {
	opts := types.ImagePullOptions{
		Platform: "amd64", // "arm64"
	}

	out, err := cli.ImagePull(ctx, "alpine:latest", opts)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	buf := make([]byte, 512)
	for {
		if _, err := out.Read(buf); err == io.EOF {
			break
		}
	}
}

// Watch for a container to stop running. Also handle user interrupts.
func watchContainer(cli *client.Client, ctx context.Context, id string) {
	statusC, errC := cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)

	sigC := make(chan os.Signal)
	signal.Notify(sigC, os.Interrupt)

	select {
	case _ = <-sigC:
		fmt.Println("(caught SIGINT)")

	case err := <-errC:
		if err != nil {
			fmt.Println("An error occured with the docker daemon")
		}

	case status := <-statusC:
		fmt.Printf("(exited with %d)\n", status.StatusCode)
	}

	dumpContainerLogs(cli, ctx, id)

	imageRemove(cli, ctx)
}

// Print logs from a container.
func dumpContainerLogs(cli *client.Client, ctx context.Context, id string) {
	conf := types.ContainerLogsOptions{
		ShowStdout: true,
	}

	out, err := cli.ContainerLogs(ctx, id, conf)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

// Remove an image.
func imageRemove(cli *client.Client, ctx context.Context) {
	opts := types.ImageRemoveOptions{
		Force: true,
	}

	id := identifyImage(cli, ctx)

	_, err := cli.ImageRemove(ctx, id, opts)
	if err != nil {
		panic(err)
	}
}

// Get the ID of an image.
func identifyImage(cli *client.Client, ctx context.Context) string {
	opts := types.ImageListOptions{}

	images, err := cli.ImageList(ctx, opts)
	if err != nil {
		panic(err)
	}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == "alpine:latest" {
				return image.ID
			}
		}
	}

	return ""
}

func main() {
	// Need a client to communicate with `dockerd(8)`
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	// Need a context for execution
	ctx := context.Background()

	// A command to run in an Alpine `sh(1)`
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"uname", "-a"}
	}
	fmt.Println(strings.Join(args, " "))

	// Run the container
	runContainer(cli, ctx, args)
}

