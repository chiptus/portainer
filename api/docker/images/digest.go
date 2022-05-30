package images

import (
	"context"
	"time"

	"github.com/containers/image/v5/docker"
	imagetypes "github.com/containers/image/v5/types"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/opencontainers/go-digest"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	_imageLocalDigestCache = cache.New(5*time.Second, 5*time.Second)
)

// Options holds docker registry object options
type Options struct {
	Auth    imagetypes.DockerAuthConfig
	Timeout time.Duration
}

type DigestClient struct {
	client         *client.Client
	opts           Options
	sysCtx         *imagetypes.SystemContext
	registryClient *RegistryClient
}

func NewClientWithOpts(opts Options, client *client.Client) *DigestClient {
	return &DigestClient{
		client: client,
		opts:   opts,
		sysCtx: &imagetypes.SystemContext{
			DockerAuthConfig: &opts.Auth,
		},
	}
}

func NewClientWithRegistry(registryClient *RegistryClient, client *client.Client) *DigestClient {
	return &DigestClient{
		client:         client,
		registryClient: registryClient,
	}
}

func (c *DigestClient) LocalDigest(img Image) (digest.Digest, error) {
	localDigest, err := cachedLocalDigest(img.FullName())
	if err == nil {
		return localDigest, nil
	}

	c.cacheAllLocalDigest()
	// get from cache again after triggering caching
	localDigest, err = cachedLocalDigest(img.FullName())
	if err != nil {
		return "", err
	}

	return localDigest, nil
}

func (c *DigestClient) RemoteDigest(image Image) (digest.Digest, error) {
	ctx, cancel := c.timeoutContext()
	defer cancel()
	// Docker references with both a tag and digest are currently not supported
	if image.Tag != "" && image.Digest != "" {
		err := image.trimDigest()
		if err != nil {
			return "", err
		}
	}

	rmRef, err := ParseReference(image.String())
	if err != nil {
		return "", errors.Wrap(err, "Cannot parse reference")
	}

	sysCtx := c.sysCtx
	if c.registryClient != nil {
		username, password, err := c.registryClient.RegistryAuth(image)
		if err != nil {
			log.Warnf("Can not find registry auth for image %s", image)
		} else {
			sysCtx = &imagetypes.SystemContext{
				DockerAuthConfig: &imagetypes.DockerAuthConfig{
					Username: username,
					Password: password,
				},
			}
		}
	}

	// Retrieve remote digest through HEAD request
	rmDigest, err := docker.GetDigest(ctx, sysCtx, rmRef)
	if err != nil {
		// fallback to public registry for hub
		if image.HubLink != "" {
			rmDigest, err = docker.GetDigest(ctx, c.sysCtx, rmRef)
			if err == nil {
				return rmDigest, nil
			}
		}
		log.Debugf("get remote digest err: %v", err)
		return "", errors.Wrap(err, "Cannot get image digest from HEAD request")
	}

	return rmDigest, nil
}

func ParseLocalImage(inspect types.ImageInspect) (*Image, error) {
	if IsLocalImage(inspect) || IsDanglingImage(inspect) {
		return nil, errors.New("the image is not regular")
	}
	fromRepoDigests, err := ParseImage(ParseImageOptions{
		// including image name but no tag
		Name: inspect.RepoDigests[0],
	})
	if err != nil {
		return nil, err
	}

	if IsNoTagImage(inspect) {
		return &fromRepoDigests, nil
	}
	fromRepoTags, err := ParseImage(ParseImageOptions{
		Name: inspect.RepoTags[0],
	})
	if err != nil {
		return nil, err
	}

	fromRepoDigests.Tag = fromRepoTags.Tag
	return &fromRepoDigests, nil
}

func (c *DigestClient) cacheAllLocalDigest() {
	inspects, err := c.client.ImageList(context.TODO(), types.ImageListOptions{
		All: true,
	})
	if err != nil {
		log.Error("run docker images error:", err)
		return
	}

	for _, inspect := range inspects {
		_inspect := types.ImageInspect{
			ID:          inspect.ID,
			RepoTags:    inspect.RepoTags,
			RepoDigests: inspect.RepoDigests,
			Parent:      inspect.ParentID,
		}
		localImage, err := ParseLocalImage(_inspect)
		if err != nil {
			continue
		}

		_imageLocalDigestCache.Set(localImage.FullName(), localImage.Digest, 0)
	}
}

func (c *DigestClient) timeoutContext() (context.Context, context.CancelFunc) {
	ctx := context.Background()
	var cancel context.CancelFunc = func() {}
	if c.opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, c.opts.Timeout)
	}
	return ctx, cancel
}

func cachedLocalDigest(imageName string) (digest.Digest, error) {
	cacheDigest, ok := _imageLocalDigestCache.Get(imageName)
	if ok {
		cachedDigest, ok := cacheDigest.(digest.Digest)
		if ok {
			return cachedDigest, nil
		}
	}

	return "", errors.Errorf("no local digest found for image: %s", imageName)
}
