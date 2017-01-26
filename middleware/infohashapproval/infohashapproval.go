// Package infohashapproval implements a Hook that fails an Announce based on a
// whitelist or blacklist of BitTorrent Infohashes.

package infohashapproval

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	ed "github.com/FactomProject/ed25519"
	"github.com/chihaya/chihaya/bittorrent"
	"github.com/chihaya/chihaya/middleware"
)

// Valid public keys
var (
	publicKeys = []string{
		"cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a",
	}
)

// ErrInfohashUnapproved is the error returned when a infohash is invalid.
var ErrInfohashUnapproved = bittorrent.ClientError("unapproved infohash")

var ErrInvalidSignature = bittorrent.ClientError("Invalid Signature")

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

	var b [20]byte
	copy(b[:], infohash[:])

	str, exists := req.Params.String("sig")
	_, whitlisted := h.approved[infohash]
	// If already whitelisted, we do not care
	if exists || !whitlisted {
		// We have a signed infohash
		signature, err := hex.DecodeString(str)
		if err != nil || len(signature) != ed.SignatureSize {
			return ctx, ErrInvalidSignature
		}

		var sigFixed [ed.SignatureSize]byte
		copy(sigFixed[:], signature[:])

		for _, k := range publicKeys {
			key, err := hex.DecodeString(k)
			if err != nil || len(key) != ed.PublicKeySize {
				continue
			}

			var pubKey [ed.PublicKeySize]byte
			copy(pubKey[:], key[:])

			valid := ed.VerifyCanonical(&pubKey, b[:], &sigFixed)
			if valid {
				fmt.Printf("Added a new infohash to the whitelist: %x\n", b[:])
				h.approved[infohash] = struct{}{}
			}
		}
	}
	fmt.Println(str)

	// In blacklist
	if len(h.unapproved) > 0 {
		if _, found := h.unapproved[infohash]; found {
			fmt.Printf("Found in blacklist, rejecting: %x\n", b[:])
			return ctx, ErrInfohashUnapproved
		}
	}

	// In whitlist
	if len(h.approved) > 0 {
		if _, found := h.approved[infohash]; found {
			fmt.Printf("Found in whitelist, accepting: %x\n", b[:])
			return ctx, nil
		}
	}

	fmt.Printf("Not found in whitelist, rejecting: %x\n", b[:])
	return ctx, ErrInfohashUnapproved
}

func (h *hook) HandleScrape(ctx context.Context, req *bittorrent.ScrapeRequest, resp *bittorrent.ScrapeResponse) (context.Context, error) {
	// Scrapes don't require any protection.
	return ctx, nil
}
