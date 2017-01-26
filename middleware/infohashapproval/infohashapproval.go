// Package infohashapproval implements a Hook that fails an Announce based on a
// whitelist or blacklist of BitTorrent Infohashes.

package infohashapproval

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/chihaya/chihaya/bittorrent"
	"github.com/chihaya/chihaya/middleware"
)

// ErrInfohashUnapproved is the error returned when a infohash is invalid.
var ErrInfohashUnapproved = bittorrent.ClientError("unapproved infohash")

// Config represents all the values required by this middleware to validate
// announce urls based on their BitTorrent Infohash.
type Config struct {
	Whitelist []string `yaml:"whitelist"`
	Blacklist []string `yaml:"blacklist"`
}

type hook struct {
	approved   map[bittorrent.InfoHash]struct{}
	unapproved map[bittorrent.InfoHash]struct{}
}

// NewHook returns an instance of the infohash approval middleware.
func NewHook(cfg Config) (middleware.Hook, error) {
	h := &hook{
		approved:   make(map[bittorrent.InfoHash]struct{}),
		unapproved: make(map[bittorrent.InfoHash]struct{}),
	}

	for _, ihString := range cfg.Whitelist {
		ihBytes, err := hex.DecodeString(ihString)
		if err != nil {
			return nil, err
		}

		fmt.Printf("%x", ihBytes)
		if len(ihBytes) != 20 {
			return nil, errors.New("Infohash " + ihString + " must be 20 bytes")
		}
		var ih bittorrent.InfoHash
		copy(ih[:], ihBytes)
		h.approved[ih] = struct{}{}
	}

	for _, ihString := range cfg.Blacklist {
		ihBytes, err := hex.DecodeString(ihString)
		if err != nil {
			return nil, err
		}

		if len(ihBytes) != 20 {
			return nil, errors.New("Infohash " + ihString + " must be 20 bytes")
		}
		var ih bittorrent.InfoHash
		copy(ih[:], ihBytes)
		h.unapproved[ih] = struct{}{}
	}

	return h, nil
}

func (h *hook) HandleAnnounce(ctx context.Context, req *bittorrent.AnnounceRequest, resp *bittorrent.AnnounceResponse) (context.Context, error) {
	infohash := req.InfoHash

	// In blacklist
	if len(h.unapproved) > 0 {
		if _, found := h.unapproved[infohash]; found {
			return ctx, ErrInfohashUnapproved
		}
	}

	// In whitlist
	if len(h.approved) > 0 {
		if _, found := h.approved[infohash]; found {
			fmt.Printf("Found in whitelist, accepting: %s\n", req.InfoHash[:])
			return ctx, nil
		}
	}

	fmt.Printf("Not found in whitelist, rejecting: %s\n", req.InfoHash[:])
	return ctx, ErrInfohashUnapproved
}

func (h *hook) HandleScrape(ctx context.Context, req *bittorrent.ScrapeRequest, resp *bittorrent.ScrapeResponse) (context.Context, error) {
	// Scrapes don't require any protection.
	return ctx, nil
}
