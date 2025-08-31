package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-getter/v2"
	"github.com/sprokhorov/kubeinit/pkg/config"
	"go.uber.org/zap"
)

const FILE_DESTINATION = "./helmfile.yaml"

// configureCluster configures the Kubernetes client based on the cloud provider and cluster name.
func configureCluster(cfg *config.Config, log *zap.SugaredLogger) error {
	switch cfg.CloudProvider {
	case config.AWS:
		log.Infow("Configuring AWS cluster", "cluster", cfg.ClusterName)
		if err := exec.Command("aws", "eks", "update-kubeconfig", "--name", cfg.ClusterName).Run(); err != nil {
			return err
		}
	case config.Azure:
		log.Infow("Configuring Azure cluster", "cluster", cfg.ClusterName)
		log.Warn("GCP support is not implemented yet")
	case config.GCP:
		log.Infow("Configuring GCP cluster", "cluster", cfg.ClusterName)
		log.Warn("GCP support is not implemented yet")
	default:
		return errors.New("unknown cloud provider")
	}
	return nil
}

// getHelmfile downloads the Helmfile from the specified source to the destination path.
// It supports URL-based sources with the following schemas: http://, https://, git://, s3://.
func getHelmfile(ctx context.Context, source string, destination string) error {
	client := &getter.Client{}
	req := &getter.Request{
		Src:     source,
		Dst:     destination,
		GetMode: getter.ModeFile,
	}

	if _, err := client.Get(ctx, req); err != nil {
		return err
	}
	stat, err := os.Stat(destination)
	if err != nil {
		return fmt.Errorf("failed to stat downloaded file: %w", err)
	}
	if stat.Size() == 0 {
		return fmt.Errorf("downloaded file is empty")
	}
	return nil
}

func main() {
	// Configure logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	log := logger.Sugar()
	cfg, err := config.New()
	if err != nil {
		log.Fatalw("Failed to load config", "error", err)
	}
	log.Infow("kubeinit started", "config", cfg)

	// Download Helmfile file
	ctx := context.Background()
	if err := getHelmfile(ctx, cfg.HelmfileFile, FILE_DESTINATION); err != nil {
		log.Fatalw("Failed to download Helmfile", "error", err)
	}

	log.Infow("Helmfile file downloaded successfully")

	// Configure kubernetes cluster
	if err := configureCluster(cfg, log); err != nil {
		log.Fatal(err)
	}
	log.Infow("Kubernetes cluster configured successfully", "cluster", cfg.ClusterName)

	// Run helmfile
	command := []string{"helmfile", "-f", FILE_DESTINATION, "apply"}
	cmd := exec.Command(command[0], command[1:]...)

	// Capture combined stdout + stderr
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("command failed", zap.Error(err), zap.ByteString("output", out))
		return
	}
	logger.Info("command output", zap.ByteString("output", out))
}
