package images

import (
	"context"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Status constants
const (
	Processing = Status("processing")
	Outdated   = Status("outdated")
	Updated    = Status("updated")
	Skipped    = Status("skipped")
	Error      = Status("error")
)

var (
	_statusCache = cache.New(5*time.Second, 5*time.Second)
)

// Status holds Docker image status analysis
type Status string

func (c *DigestClient) Status(ctx context.Context, image string) (Status, error) {
	log.Debugf("orginal incoming image name is %s", image)
	img, err := ParseImage(ParseImageOptions{
		Name: image,
	})
	if err != nil {
		log.Debugf("image parse failed: %s", image)
		return Error, err
	}
	imageString := img.String()
	s, err := cachedImageStatus(imageString)
	if err == nil {
		return s, nil
	}
	s, err = c.status(img)
	if err != nil {
		log.Debugf("error when fetch a certain image status: %v", err)
		_statusCache.Set(imageString, Error, 0)
		return Error, err
	}
	_statusCache.Set(imageString, s, 0)
	return s, err
}

func (c *DigestClient) status(img Image) (Status, error) {
	image := img.FullName()
	log.Debugf("start image: %s", image)
	dg := img.Digest
	if dg == "" {
		log.Debugf("incoming local digest is null, fetch via docker images")
		var err error
		dg, err = c.LocalDigest(img)
		if err != nil {
			log.Debugf("err when fetch local digest for image: %s, %v", image, err)
			return Skipped, err
		}
	}
	remoteDigest, err := c.RemoteDigest(img)
	if err != nil {
		log.Errorf("error when fetch remote digest for image: %s, %v", image, err)
		return Error, err
	}
	log.Debugf("digest from remote image %s is %s, local is: %s", image, remoteDigest, dg)
	var imageStatus Status
	if dg == remoteDigest {
		imageStatus = Updated
	} else {
		imageStatus = Outdated
	}
	return imageStatus, nil
}

func cachedImageStatus(imageName string) (Status, error) {
	s, ok := _statusCache.Get(imageName)
	if ok {
		s, ok := s.(Status)
		if ok {
			return s, nil
		}
	}
	return "", errors.New(fmt.Sprintf("no image status found in cache: %s", imageName))
}
