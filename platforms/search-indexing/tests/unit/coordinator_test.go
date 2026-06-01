package unit

import (
	"context"
	"testing"

	"github.com/tikiclone/tiki/platforms/search-indexing/internal/coordinator"
)

func TestRegisterNode(t *testing.T) {
	repo := coordinator.NewInMemoryRepository()
	svc := coordinator.NewService(repo)
	ctx := context.Background()

	node, err := svc.RegisterNode(ctx, &coordinator.IndexNode{
		Address:         "node-1:9200",
		Region:          "us-east-1",
		AvailableShards: 10,
		LoadPercentage:  0,
	})
	if err != nil {
		t.Fatalf("RegisterNode failed: %v", err)
	}
	if node.ID == "" {
		t.Error("expected node ID to be set")
	}
	if node.IsActive != true {
		t.Error("expected node to be active")
	}
}

func TestListNode(t *testing.T) {
	repo := coordinator.NewInMemoryRepository()
	svc := coordinator.NewService(repo)
	ctx := context.Background()

	svc.RegisterNode(ctx, &coordinator.IndexNode{Address: "node-1:9200", Region: "us-east-1", AvailableShards: 10})
	svc.RegisterNode(ctx, &coordinator.IndexNode{Address: "node-2:9200", Region: "us-west-2", AvailableShards: 8})

	nodes, err := svc.ListNodes(ctx)
	if err != nil {
		t.Fatalf("ListNodes failed: %v", err)
	}
	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(nodes))
	}
}

func TestAssignShard(t *testing.T) {
	repo := coordinator.NewInMemoryRepository()
	svc := coordinator.NewService(repo)
	ctx := context.Background()

	svc.RegisterNode(ctx, &coordinator.IndexNode{Address: "node-1:9200", Region: "us-east-1", AvailableShards: 10, LoadPercentage: 20})
	svc.RegisterNode(ctx, &coordinator.IndexNode{Address: "node-2:9200", Region: "us-west-2", AvailableShards: 8, LoadPercentage: 50})

	shard, err := svc.AssignShard(ctx, &coordinator.IndexShard{
		IndexName:   "products",
		ShardNumber: 0,
		DocCount:    1000,
		SizeBytes:   1048576,
	})
	if err != nil {
		t.Fatalf("AssignShard failed: %v", err)
	}
	if shard.NodeID == "" {
		t.Error("expected shard to be assigned to a node")
	}
	if shard.Status != coordinator.ShardActive {
		t.Errorf("expected shard status active, got %s", shard.Status)
	}

	shard2, err := svc.AssignShard(ctx, &coordinator.IndexShard{
		IndexName:   "products",
		ShardNumber: 1,
		DocCount:    500,
		SizeBytes:   524288,
	})
	if err != nil {
		t.Fatalf("AssignShard second shard failed: %v", err)
	}
	if shard2.NodeID == shard.NodeID {
		t.Log("both shards assigned to same node (expected due to load balancing)")
	}
}

func TestAssignShardNoNodes(t *testing.T) {
	repo := coordinator.NewInMemoryRepository()
	svc := coordinator.NewService(repo)
	ctx := context.Background()

	_, err := svc.AssignShard(ctx, &coordinator.IndexShard{
		IndexName:   "products",
		ShardNumber: 0,
	})
	if err != coordinator.ErrNoAvailableNodes {
		t.Errorf("expected ErrNoAvailableNodes, got %v", err)
	}
}

func TestRebalanceShards(t *testing.T) {
	repo := coordinator.NewInMemoryRepository()
	svc := coordinator.NewService(repo)
	ctx := context.Background()

	n1, _ := svc.RegisterNode(ctx, &coordinator.IndexNode{Address: "node-1:9200", Region: "us-east-1", AvailableShards: 10, LoadPercentage: 10})
	n2, _ := svc.RegisterNode(ctx, &coordinator.IndexNode{Address: "node-2:9200", Region: "us-west-2", AvailableShards: 10, LoadPercentage: 10})

	svc.AssignShard(ctx, &coordinator.IndexShard{ID: "s1", IndexName: "products", ShardNumber: 0, NodeID: n1.ID})
	svc.AssignShard(ctx, &coordinator.IndexShard{ID: "s2", IndexName: "products", ShardNumber: 1, NodeID: n1.ID})
	svc.AssignShard(ctx, &coordinator.IndexShard{ID: "s3", IndexName: "products", ShardNumber: 2, NodeID: n2.ID})

	if err := svc.RebalanceShards(ctx); err != nil {
		t.Fatalf("RebalanceShards failed: %v", err)
	}

	dist, _ := svc.GetShardDistribution(ctx)
	for _, d := range dist {
		t.Logf("Node %s has %d shards", d.NodeID, d.Count)
	}
}

func TestGetShardDistribution(t *testing.T) {
	repo := coordinator.NewInMemoryRepository()
	svc := coordinator.NewService(repo)
	ctx := context.Background()

	svc.RegisterNode(ctx, &coordinator.IndexNode{Address: "node-1:9200", Region: "us-east-1", AvailableShards: 10})
	svc.RegisterNode(ctx, &coordinator.IndexNode{Address: "node-2:9200", Region: "us-west-2", AvailableShards: 10})

	svc.AssignShard(ctx, &coordinator.IndexShard{ID: "s1", IndexName: "products", ShardNumber: 0})
	svc.AssignShard(ctx, &coordinator.IndexShard{ID: "s2", IndexName: "orders", ShardNumber: 0})

	dist, err := svc.GetShardDistribution(ctx)
	if err != nil {
		t.Fatalf("GetShardDistribution failed: %v", err)
	}
	if len(dist) != 2 {
		t.Errorf("expected distribution for 2 nodes, got %d", len(dist))
	}
}

func TestHandleNodeFailure(t *testing.T) {
	repo := coordinator.NewInMemoryRepository()
	svc := coordinator.NewService(repo)
	ctx := context.Background()

	n1, _ := svc.RegisterNode(ctx, &coordinator.IndexNode{Address: "node-1:9200", Region: "us-east-1", AvailableShards: 10, LoadPercentage: 30})
	svc.RegisterNode(ctx, &coordinator.IndexNode{Address: "node-2:9200", Region: "us-west-2", AvailableShards: 10, LoadPercentage: 10})

	svc.AssignShard(ctx, &coordinator.IndexShard{ID: "s1", IndexName: "products", ShardNumber: 0, NodeID: n1.ID})
	svc.AssignShard(ctx, &coordinator.IndexShard{ID: "s2", IndexName: "products", ShardNumber: 1, NodeID: n1.ID})

	if err := svc.HandleNodeFailure(ctx, n1.ID); err != nil {
		t.Fatalf("HandleNodeFailure failed: %v", err)
	}

	node, _ := svc.GetNode(ctx, n1.ID)
	if node.IsActive {
		t.Error("expected node to be inactive after failure")
	}
}
