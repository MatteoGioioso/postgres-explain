package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eko/gocache/lib/v4/cache"
	gocachestore "github.com/eko/gocache/store/go_cache/v4"
	gocache "github.com/patrickmn/go-cache"
	"time"
)

type Params struct {
}

type Client struct {
	client *cache.Cache[[]byte]
}

func New(params Params) (*Client, error) {
	client := gocache.New(10*time.Minute, 10*time.Minute)
	store := gocachestore.NewGoCache(client)
	return &Client{client: cache.New[[]byte](store)}, nil
}

func (c *Client) SetInstance(ctx context.Context, instance Instance) error {
	if err := c.SetCluster(ctx, instance.ClusterName); err != nil {
		return fmt.Errorf("could not SetCluster: %v", err)
	}

	clusterBytes, err := c.client.Get(ctx, c.getInstanceKey(instance.ClusterName))
	if err != nil {
		return fmt.Errorf("could not Get: %v", err)
	}

	var cluster map[string]Instance
	if err := json.Unmarshal(clusterBytes, &cluster); err != nil {
		return fmt.Errorf("could not Unmarshal: %v", err)
	}

	cluster[instance.Name] = instance

	marshal, err := json.Marshal(cluster)
	if err != nil {
		return fmt.Errorf("could not Marshal: %v", err)
	}
	return c.client.Set(ctx, c.getInstanceKey(instance.ClusterName), marshal)
}

func (c *Client) GetInstance(ctx context.Context, clusterName, name string) (Instance, error) {
	var cluster map[string]Instance

	get, err := c.client.Get(ctx, c.getInstanceKey(clusterName))
	if err != nil {
		return Instance{}, fmt.Errorf("could not Get: %v", err)
	}

	if err := json.Unmarshal(get, &cluster); err != nil {
		return Instance{}, fmt.Errorf("could not Unmarshal: %v", err)
	}

	return cluster[name], nil
}

func (c *Client) SetCluster(ctx context.Context, clusterName string) error {
	var clusters map[string]string
	get, err := c.client.Get(ctx, "clusters")
	if err != nil {
		return err
	}

	if err := json.Unmarshal(get, &clusters); err != nil {
		return err
	}

	clusters[clusterName] = clusterName
	marshal, err := json.Marshal(clusters)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, "clusters", marshal)
}

func (c *Client) GetClusters(ctx context.Context) (map[string]string, error) {
	var clusters map[string]string

	get, err := c.client.Get(ctx, "clusters")
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(get, &clusters); err != nil {
		return nil, err
	}

	return clusters, nil
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
