package spinup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
)

type Container struct {
	ID        string
	Image     string
	Cmd       []string
	IPAddress string
	Timeout   time.Duration
	cli       *client.Client
}

func NewContainer(image string, timeout time.Duration, cmd ...string) Container {
	return Container{
		Image:   image,
		Cmd:     cmd,
		Timeout: timeout,
	}
}

func (c *Container) SpinUp(ctx context.Context) error {
	var err error
	c.cli, err = client.NewEnvClient()
	if err != nil {
		return err
	}

	pullResponse, err := c.cli.ImagePull(ctx, c.Image,
		types.ImagePullOptions{})
	if err != nil {
		return err
	}

	dec := json.NewDecoder(pullResponse)
	for {
		var p jsonmessage.JSONMessage
		if err := dec.Decode(&p); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		switch p.Status {
		case "Extracting":
			fmt.Printf("%s: %s\r", p.ID, p.Progress)
		case "Pull complete":
			fmt.Printf("%s: %s\r\n", p.ID, p.Progress)
		}
	}

	cfg := &container.Config{
		Image: c.Image,
	}
	if len(c.Cmd) > 0 {
		cfg.Cmd = c.Cmd
	}
	hostCfg := &container.HostConfig{
		AutoRemove: true,
	}
	resp, err := c.cli.ContainerCreate(ctx, cfg, hostCfg, nil, "")
	if err != nil {
		return err
	}
	c.ID = resp.ID

	err = c.cli.ContainerStart(ctx, c.ID,
		types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	info, err := c.cli.ContainerInspect(ctx, c.ID)
	if err != nil {
		return err
	}

	c.IPAddress = info.NetworkSettings.DefaultNetworkSettings.IPAddress
	return nil
}

func (c *Container) SpinDown(ctx context.Context) error {
	return c.cli.ContainerStop(ctx, c.ID, &c.Timeout)
}
