package commandutils
//TODO: Move to pkg services?

import (
	"context"
	"fmt"

	"github.com/dboxed/dboxed/pkg/baseclient"
	"github.com/dboxed/dboxed/pkg/clients"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/google/uuid"
)

func GetWorkspace(ctx context.Context, c *baseclient.Client, workspace string) (*models.Workspace, error) {
	c2 := clients.WorkspacesClient{Client: c}
	if uuid.Validate(workspace) == nil {
		v, err := c2.GetWorkspaceById(ctx, workspace)
		if err != nil {
			return nil, err
		}
		return v, nil
	} else {
		l, err := c2.ListWorkspaces(ctx)
		if err != nil {
			return nil, err
		}
		for _, w := range l {
			if w.Name == workspace {
				return &w, nil
			}
		}
		return nil, fmt.Errorf("workspace not found")
	}
}

func GetBox(ctx context.Context, c *baseclient.Client, box string) (*models.Box, error) {
	c2 := clients.BoxClient{Client: c}
	if uuid.Validate(box) == nil {
		v, err := c2.GetBoxById(ctx, box)
		if err != nil {
			return nil, err
		}
		return v, nil
	} else {
		v, err := c2.GetBoxByName(ctx, box)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func GetVolumeProvider(ctx context.Context, c *baseclient.Client, volumeProvider string) (*models.VolumeProvider, error) {
	c2 := clients.VolumeProvidersClient{Client: c}
	if uuid.Validate(volumeProvider) == nil {
		v, err := c2.GetVolumeProviderById(ctx, volumeProvider)
		if err != nil {
			return nil, err
		}
		return v, nil
	} else {
		v, err := c2.GetVolumeProviderByName(ctx, volumeProvider)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func GetVolume(ctx context.Context, c *baseclient.Client, volume string) (*models.Volume, error) {
	c2 := clients.VolumesClient{Client: c}
	if uuid.Validate(volume) == nil {
		v, err := c2.GetVolumeById(ctx, volume)
		if err != nil {
			return nil, err
		}
		return v, nil
	} else {
		v, err := c2.GetVolumeByName(ctx, volume)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func GetNetwork(ctx context.Context, c *baseclient.Client, network string) (*models.Network, error) {
	c2 := clients.NetworkClient{Client: c}
	if uuid.Validate(network) == nil {
		n, err := c2.GetNetworkById(ctx, network)
		if err != nil {
			return nil, err
		}
		return n, nil
	} else {
		n, err := c2.GetNetworkByName(ctx, network)
		if err != nil {
			return nil, err
		}
		return n, nil
	}
}

func GetS3Bucket(ctx context.Context, c *baseclient.Client, s3Bucket string) (*models.S3Bucket, error) {
	c2 := clients.S3BucketsClient{Client: c}
	if uuid.Validate(s3Bucket) == nil {
		s, err := c2.GetS3BucketById(ctx, s3Bucket)
		if err != nil {
			return nil, err
		}
		return s, nil
	} else {
		s, err := c2.GetS3BucketByBucketName(ctx, s3Bucket)
		if err != nil {
			return nil, err
		}
		return s, nil
	}
}

func GetLoadBalancer(ctx context.Context, c *baseclient.Client, lb string) (*models.LoadBalancer, error) {
	c2 := clients.LoadBalancerClient{Client: c}
	if uuid.Validate(lb) == nil {
		p, err := c2.GetLoadBalancerById(ctx, lb)
		if err != nil {
			return nil, err
		}
		return p, nil
	} else {
		l, err := c2.ListLoadBalancers(ctx)
		if err != nil {
			return nil, err
		}
		for _, p := range l {
			if p.Name == lb {
				return &p, nil
			}
		}
		return nil, fmt.Errorf("load balancer not found: %s", lb)
	}
}

func GetMachine(ctx context.Context, c *baseclient.Client, machine string) (*models.Machine, error) {
	c2 := clients.MachineClient{Client: c}
	if uuid.Validate(machine) == nil {
		m, err := c2.GetMachineById(ctx, machine)
		if err != nil {
			return nil, err
		}
		return m, nil
	} else {
		l, err := c2.ListMachines(ctx)
		if err != nil {
			return nil, err
		}
		for _, m := range l {
			if m.Name == machine {
				return &m, nil
			}
		}
		return nil, fmt.Errorf("machine not found: %s", machine)
	}
}

func GetMachineProvider(ctx context.Context, c *baseclient.Client, machineProvider string) (*models.MachineProvider, error) {
	c2 := clients.MachineProviderClient{Client: c}
	if uuid.Validate(machineProvider) == nil {
		mp, err := c2.GetMachineProviderById(ctx, machineProvider)
		if err != nil {
			return nil, err
		}
		return mp, nil
	} else {
		l, err := c2.ListMachineProviders(ctx)
		if err != nil {
			return nil, err
		}
		for _, mp := range l {
			if mp.Name == machineProvider {
				return &mp, nil
			}
		}
		return nil, fmt.Errorf("machine provider not found: %s", machineProvider)
	}
}

func GetGitCredentials(ctx context.Context, c *baseclient.Client, gitCredentials string) (*models.GitCredentials, error) {
	c2 := clients.GitCredentialsClient{Client: c}
	if uuid.Validate(gitCredentials) == nil {
		gc, err := c2.GetGitCredentialsById(ctx, gitCredentials)
		if err != nil {
			return nil, err
		}
		return gc, nil
	} else {
		// GitCredentials doesn't have a name field, so we search by host
		l, err := c2.ListGitCredentials(ctx)
		if err != nil {
			return nil, err
		}
		for _, gc := range l {
			if gc.Host == gitCredentials {
				return &gc, nil
			}
		}
		return nil, fmt.Errorf("git credentials not found: %s", gitCredentials)
	}
}

func GetGitSpec(ctx context.Context, c *baseclient.Client, gitSpec string) (*models.GitSpec, error) {
	c2 := clients.GitSpecClient{Client: c}
	// GitSpec only supports ID lookup
	gs, err := c2.GetGitSpecById(ctx, gitSpec)
	if err != nil {
		return nil, err
	}
	return gs, nil
}
