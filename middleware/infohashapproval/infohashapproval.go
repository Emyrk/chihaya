// Package infohashapproval implements a Hook that fails an Announce based on a
// whitelist or blacklist of BitTorrent Infohashes. To allow identities in the
// authority add to the whitelist, they must sign their torrents, by using the
// factomd-torrent

package infohashapproval

import (
	"context"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"os/user"

	ed "github.com/FactomProject/ed25519"
	"github.com/chihaya/chihaya/bittorrent"
	"github.com/chihaya/chihaya/middleware"

	"github.com/FactomProject/factomd/common/interfaces"
)

// Valid public keys
var (
	publicKeys = []string{
		"cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a",
	}

	// DBPaths
	ldbPath  string = "/.factom/m2/tracker-storage/infohash_ldb.db"
	boltPath string = "/.factom/m2/tracker-storage/infohash_ldb.db"
)

// ErrInfohashUnapproved is the error returned when a infohash is invalid.
var ErrInfohashUnapproved = bittorrent.ClientError("unapproved infohash")

var ErrInvalidSignature = bittorrent.ClientError("Invalid Signature")

// Config represents all the values required by this middleware to validate
// announce urls based on their BitTorrent Infohash.
type Config struct {
	Whitelist []string `yaml:"whitelist"`
	Blacklist []string `yaml:"blacklist"`
	Database  string   `yaml:database"`
}

type hook struct {
	approved   map[bittorrent.InfoHash]struct{}
	unapproved map[bittorrent.InfoHash]struct{}
}

var MiddleWareDatabase interfaces.IDatabase

// NewHook returns an instance of the infohash approval middleware.
func NewHook(cfg Config) (middleware.Hook, error) {
	h := &hook{
		approved:   make(map[bittorrent.InfoHash]struct{}),
		unapproved: make(map[bittorrent.InfoHash]struct{}),
	}

	// Load from Config. If loaded from config, it will not go into the database.
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

	switch cfg.Database {
	case "Map":
		log.Println("Infohash middleware is running without a database, and will not save")
		MiddleWareDatabase = nil
	case "Bolt":
		db, err := NewOrOpenBoltDBWallet(GetHomeDir() + boltPath)
		if err != nil {
			panic("Failed to create a bolt database, " + err.Error())
		}
		MiddleWareDatabase = db

	case "LDB":
		db, err := NewOrOpenLevelDBWallet(GetHomeDir() + boltPath)
		if err != nil {
			panic("Failed to create a bolt database, " + err.Error())
		}
		MiddleWareDatabase = db
	}

	// Load from database and update our map
	if MiddleWareDatabase != nil {
		whitelist, err := MiddleWareDatabase.ListAllKeys([]byte("whitelist"))
		if err != nil {
			panic("Could not read database: " + err.Error())
		}
		for _, key := range whitelist {
			var ih bittorrent.InfoHash
			copy(ih[:], key[:])
			h.approved[ih] = struct{}{}
		}

		blacklist, err := MiddleWareDatabase.ListAllKeys([]byte("blacklist"))
		if err != nil {
			panic("Could not read database: " + err.Error())
		}
		for _, key := range blacklist {
			var ih bittorrent.InfoHash
			copy(ih[:], key[:])
			h.unapproved[ih] = struct{}{}
		}
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
	if exists && !whitlisted {
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
				h.approved[infohash] = struct{}{}
				if MiddleWareDatabase != nil {
					var t interfaces.BinaryMarshallable
					err := MiddleWareDatabase.Put([]byte("whitelist"), b[:], t)
					if err != nil {
						log.Printf("Failed to write %x infohash to whitelist database: %s\n", b, err.Error())
					}
				}
				break
			}
		}
	}

	// In blacklist
	if len(h.unapproved) > 0 {
		if _, found := h.unapproved[infohash]; found {
			return ctx, ErrInfohashUnapproved
		}
	}

	// In whitlist
	if len(h.approved) > 0 {
		if _, found := h.approved[infohash]; found {
			return ctx, nil
		}
	}

	return ctx, ErrInfohashUnapproved
}

func (h *hook) HandleScrape(ctx context.Context, req *bittorrent.ScrapeRequest, resp *bittorrent.ScrapeResponse) (context.Context, error) {
	// Scrapes don't require any protection.
	return ctx, nil
}

func GetHomeDir() string {
	// Get the OS specific home directory via the Go standard lib.
	var homeDir string
	usr, err := user.Current()
	if err == nil {
		homeDir = usr.HomeDir
	}

	// Fall back to standard HOME environment variable that works
	// for most POSIX OSes if the directory from the Go standard
	// lib failed.
	if err != nil || homeDir == "" {
		homeDir = os.Getenv("HOME")
	}
	return homeDir
}
