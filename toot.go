package fedilib

import (
	"errors"
	"io"
	"os"

	"github.com/mattn/go-mastodon"
)

type Toot interface {
	TootWithImageReader(toot mastodon.Toot, img io.Reader, alt string) error
	TootWithImage(toot mastodon.Toot, ipath string) error
}

// TootWithImage posts a new status with image (unless empty)
func (s *Fedi) TootWithImageReader(toot mastodon.Toot, img io.Reader, alt string) error {

	if img != nil {
		if a, err := s.Client().UploadMediaFromReader(s.Ctx(), img); err != nil {
			return errors.Join(errors.New("mastodon upload media failed"), err)
		} else {
			a.Description = alt
			s.Log().Println("media: ", a)
			toot.MediaIDs = append(toot.MediaIDs, a.ID)
		}
	}

	if st, err := s.Client().PostStatus(s.Ctx(), &toot); err != nil {
		return errors.Join(errors.New("mastodon post status failed"), err)
	} else {
		s.Log().Println("posted new status ", st.ID)
	}
	return nil
}

// TootWithImage posts a new status with image (unless empty)
func (s *Fedi) TootWithImage(toot mastodon.Toot, ipath string) (err error) {
	var ifile *os.File
	if ipath != "" {
		s.Log().Println("post status with image: ", toot.Status, ipath)
		ifile, err = os.Open(ipath)
		if err != nil {
			return err
		}
		defer ifile.Close()
		err = s.TootWithImageReader(toot, ifile, ipath)
	} else {
		err = s.TootWithImageReader(toot, nil, ipath)
	}
	return err
}
