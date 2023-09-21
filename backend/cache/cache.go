package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	gocachestore "github.com/eko/gocache/store/go_cache/v4"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"time"
)

type Params struct{}

type Client struct {
	client *cache.Cache[[]byte]
}

func New(params Params) (*Client, error) {
	client := gocache.New(10*time.Minute, 10*time.Minute)
	return &Client{client: cache.New[[]byte](gocachestore.NewGoCache(client))}, nil
}

func (c *Client) SetInstance(ctx context.Context, instance Instance) error {
	if err := c.SetCluster(ctx, instance.ClusterName); err != nil {
		return fmt.Errorf("could not SetCluster: %v", err)
	}

	cluster, err := c.getClusterInstances(ctx, instance.ClusterName)
	if err != nil {
		return fmt.Errorf("could not getClusterInstances: %v", err)
	}

	cluster[instance.Name] = instance

	marshal, err := json.Marshal(cluster)
	if err != nil {
		return fmt.Errorf("could not Marshal: %v", err)
	}
	return c.client.Set(ctx, c.getInstanceKey(instance.ClusterName), marshal)
}

func (c *Client) GetInstance(ctx context.Context, clusterName, name string) (Instance, error) {
	cluster, err := c.getClusterInstances(ctx, clusterName)
	if err != nil {
		return Instance{}, err
	}

	return cluster[name], nil
}

func (c *Client) SetCluster(ctx context.Context, clusterName string) error {
	clusters, err := c.GetClusters(ctx)
	if err != nil {
		return fmt.Errorf("could not GetClusters: %v", err)
	}

	clusters[clusterName] = clusterName
	marshal, err := json.Marshal(clusters)
	if err != nil {
		return fmt.Errorf("could not marshal clusters: %v", err)
	}

	return c.client.Set(ctx, "clusters", marshal)
}

func (c *Client) GetClusters(ctx context.Context) (map[string]string, error) {
	clusters := make(map[string]string)

	get, err := c.client.Get(ctx, "clusters")
	if errors.Is(err, store.NotFound{}) {
		return clusters, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not Get from cache: %v", err)
	}

	if err := json.Unmarshal(get, &clusters); err != nil {
		return nil, err
	}

	return clusters, nil
}

func (c *Client) GetClusterInstances(ctx context.Context, clusterName string) (map[string]Instance, error) {
	return c.getClusterInstances(ctx, clusterName)
}

func (c *Client) getClusterInstances(ctx context.Context, clusterName string) (map[string]Instance, error) {
	instancesMap := make(map[string]Instance)
	key := c.getInstanceKey(clusterName)
	get, err := c.client.Get(ctx, key)
	if errors.Is(err, store.NotFound{}) {
		return instancesMap, nil
	}
	if err != nil {
		return nil, fmt.Errorf("could not Get object from cache with key %v: %v", key, err)
	}

	if err := json.Unmarshal(get, &instancesMap); err != nil {
		return nil, fmt.Errorf("could not Unmarshal: %v", err)
	}
	return instancesMap, nil
}

func (c *Client) getInstanceKey(clusterName string) string {
	return fmt.Sprintf("cluster-%v", clusterName)
}

type Instance struct {
	ClusterName   string `json:"cluster_name"`
	Name          string `json:"instance_name"`
	Host          string `json:"host"`
	CollectorHost string `json:"collector_host"`
}
