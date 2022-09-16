package images

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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
	log.Debug().Str("image", image).Msg("original incoming image")

	img, err := ParseImage(ParseImageOptions{
		Name: image,
	})
	if err != nil {
		log.Debug().Str("image", image).Msg("parse failed")

		return Error, err
	}

	imageString := img.String()
	s, err := cachedImageStatus(imageString)
	if err == nil {
		return s, nil
	}

	s, err = c.status(img)
	if err != nil {
		log.Debug().Err(err).Msg("fetching a certain image status")
		_statusCache.Set(imageString, Error, 0)

		return Error, err
	}

	_statusCache.Set(imageString, s, 0)
	return s, err
}

func (c *DigestClient) status(img Image) (Status, error) {
	image := img.FullName()

	log.Debug().Str("image", image).Msg("start image")

	dg := img.Digest
	if dg == "" {
		log.Debug().Msg("incoming local digest is null, fetching via docker images")

		var err error
		dg, err = c.LocalDigest(img)
		if err != nil {
			log.Debug().Str("image", image).Err(err).Msg("fetching local digest for image")

			return Skipped, err
		}
	}
	remoteDigest, err := c.RemoteDigest(img)
	if err != nil {
		log.Error().Str("image", image).Err(err).Msg("fetching remote digest for image")

		return Error, err
	}

	log.Debug().
		Str("image", image).
		Stringer("remote_digest", remoteDigest).
		Stringer("local_digest", dg).
		Msg("")

	imageStatus := Outdated
	if dg == remoteDigest {
		imageStatus = Updated
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

	return "", errors.Errorf("no image status found in cache: %s", imageName)
}
