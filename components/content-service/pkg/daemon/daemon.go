// Copyright (c) 2021 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package daemon

import (
	"context"
	"io/ioutil"

	"github.com/gitpod-io/gitpod/content-service/api"
	"github.com/gitpod-io/gitpod/content-service/pkg/service"
	"github.com/gitpod-io/gitpod/content-service/pkg/storage"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
)

// NewDaemon produces a new daemon
func NewDaemon(strg storage.Config) (*Daemon, error) {
	srv, err := service.NewContentService()
	if err != nil {
		return nil, xerrors.Errorf("cannot create content service: %w", err)
	}
	return &Daemon{srv, strg}, nil
}

// Daemon provides the content service
type Daemon struct {
	service *service.ContentService
	strg    storage.Config
}

// Start runs all parts of the daemon until stop is called
func (d *Daemon) Start() error {
	return tmp(d.strg)
}

func tmp(strg storage.Config) error {
	ctx := context.Background()
	remoteStorage, err := storage.NewDirectAccess(&strg)
	if err != nil {
		return xerrors.Errorf("cannot use configured storage: %w", err)
	}

	err = remoteStorage.Init(ctx, "123", "abc")
	if err != nil {
		return xerrors.Errorf("cannot use configured storage: %w", err)
	}

	err = remoteStorage.EnsureExists(ctx)
	if err != nil {
		return xerrors.Errorf("cannot use configured storage: %w", err)
	}

	_, _, err = remoteStorage.Upload(ctx, "/config/config.json", "config.json")
	if err != nil {
		return xerrors.Errorf("cannot upload to configured storage: %w", err)
	}

	err = remoteStorage.BlobStore(ctx)
	if err != nil {
		return xerrors.Errorf("cannot do blobstore: %w", err)
	}

	data := make([]byte, 1e+9) // 1 GB
	err = ioutil.WriteFile("/testfile", data, 0644)
	if err != nil {
		return xerrors.Errorf("error writing big file: %w", err)
	}

	_, _, err = remoteStorage.Upload(ctx, "/testfile", "config.json")
	if err != nil {
		return xerrors.Errorf("cannot upload to configured storage (2): %w", err)
	}
	return nil
}

// Register registers all gRPC services provided by this daemon
func (d *Daemon) Register(srv *grpc.Server) {
	api.RegisterContentServiceServer(srv, d.service)
}

// Stop gracefully shuts down the daemon. Once stopped, it
// cannot be started again.
func (d *Daemon) Stop() error {
	return nil
}
