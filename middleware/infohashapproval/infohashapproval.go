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
	"sync"
	"time"

	ed "github.com/FactomProject/ed25519"
	"github.com/chihaya/chihaya/bittorrent"
	"github.com/chihaya/chihaya/middleware"
	"github.com/chihaya/chihaya/pkg/stopper"

	"github.com/FactomProject/factomd/common/interfaces"
)

// Valid public keys
var (
	// DBPaths
	ldbPath  = "/.factom/m2/tracker-storage/infohash_ldb.db"
	boltPath = "/.factom/m2/tracker-storage/infohash_ldb.db"
)

// ErrInfohashUnapproved is the error returned when a infohash is invalid.
var ErrInfohashUnapproved = bittorrent.ClientError("unapproved infohash")

// ErrInvalidSignature is the error is the signature is invalid
var ErrInvalidSignature = bittorrent.ClientError("Invalid Signature")

// Config represents all the values required by this middleware to validate
// announce urls based on their BitTorrent Infohash.
type Config struct {
	Whitelist []string `yaml:"whitelist"`
	Blacklist []string `yaml:"blacklist"`
	Database  string   `yaml:"database"`
	Signers   []string `yaml:"signers"`
}

type hook struct {
	approved   map[bittorrent.InfoHash]struct{}
	unapproved map[bittorrent.InfoHash]struct{}

	pendingWrites      chan bittorrent.InfoHash // Pending saves to database
	MiddleWareDatabase interfaces.IDatabase
	closing            chan struct{}

	Signers []string
	// We need 1 write opertation per infohash. The rest is reads,
	// for that one moment, we will need to lock the map
	sync.RWMutex
}

// NewHook returns an instance of the infohash approval middleware.
func NewHook(cfg Config) (middleware.Hook, error) {
	InitPrometheus()
	h := &hook{
		approved:      make(map[bittorrent.InfoHash]struct{}),
		unapproved:    make(map[bittorrent.InfoHash]struct{}),
		pendingWrites: make(chan bittorrent.InfoHash, 25000),
		closing:       make(chan struct{}),
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
		h.MiddleWareDatabase = nil
	case "Bolt":
		db, err := NewOrOpenBoltDB(GetHomeDir() + boltPath)
		if err != nil {
			panic("Failed to create a bolt database, " + err.Error())
		}
		h.MiddleWareDatabase = db

	case "LDB":
		db, err := NewOrOpenLevelDB(GetHomeDir() + boltPath)
		if err != nil {
			panic("Failed to create a bolt database, " + err.Error())
		}
		h.MiddleWareDatabase = db
	}

	go h.writeToDatabase()

	// Load from database and update our map
	if h.MiddleWareDatabase != nil {
		whitelist, err := h.MiddleWareDatabase.ListAllKeys([]byte("whitelist"))
		if err != nil {
			panic("Could not read database: " + err.Error())
		}
		for _, key := range whitelist {
			var ih bittorrent.InfoHash
			copy(ih[:], key[:])
			h.approved[ih] = struct{}{}
		}

		blacklist, err := h.MiddleWareDatabase.ListAllKeys([]byte("blacklist"))
		if err != nil {
			panic("Could not read database: " + err.Error())
		}
		for _, key := range blacklist {
			var ih bittorrent.InfoHash
			copy(ih[:], key[:])
			h.unapproved[ih] = struct{}{}
		}
	}

	h.Signers = cfg.Signers

	return h, nil
}

func (h *hook) Stop() <-chan error {
	log.Println("Attempting to shutdown InfohashApproval middleware")
	select {
	case <-h.closing:
		return stopper.AlreadyStopped
	default:
	}
	c := make(chan error)
	go func() {
		h.closing <- struct{}{}
		//close(h.closing)
		close(c)
	}()
	return c
}

func (h *hook) writeToDatabase() {
	for {
		select {
		case ih := <-h.pendingWrites:
			h.Lock()
			h.approved[ih] = struct{}{}
			h.Unlock()

			if h.MiddleWareDatabase != nil {
				var b [20]byte
				copy(b[:], ih[:])

				t := new(EmptyStruct)
				err := h.MiddleWareDatabase.Put([]byte("whitelist"), b[:], t)
				if err != nil {
					log.Printf("Failed to write %x infohash to whitelist database: %s\n", b, err.Error())
				}
			}
		case <-h.closing:
			log.Println("InfohashApproval Stopped")
			return
		}
	}
}

func (h *hook) HandleAnnounce(ctx context.Context, req *bittorrent.AnnounceRequest, resp *bittorrent.AnnounceResponse) (context.Context, error) {
	start := time.Now().UnixNano()
	defer chihayaAnnounceResponseTime.Observe(float64(time.Now().UnixNano()-start) / 1e9)
	infohash := req.InfoHash

	var b [20]byte
	copy(b[:], infohash[:])

	str, sigExists := req.Params.String("sig")
	h.RLock()
	_, whitlisted := h.approved[infohash]
	h.RUnlock()
	chihayaAnnounceCount.Add(1)
	// log.Infof("Announce recieved for infohash %x. Whitelisted: %t", b, whitlisted)
	// If already whitelisted, we do not care
	if sigExists && !whitlisted {
		// We have a signed infohash
		signature, err := hex.DecodeString(str)
		if err != nil || len(signature) != ed.SignatureSize {
			chihayaWhitelistFail.Add(1)
			return ctx, ErrInvalidSignature
		}

		var sigFixed [ed.SignatureSize]byte
		copy(sigFixed[:], signature[:])

		for _, k := range h.Signers {
			key, err := hex.DecodeString(k)
			if err != nil || len(key) != ed.PublicKeySize {
				continue
			}

			var pubKey [ed.PublicKeySize]byte
			copy(pubKey[:], key[:])

			valid := ed.VerifyCanonical(&pubKey, b[:], &sigFixed)
			if valid {
				/*h.Lock()
				h.approved[infohash] = struct{}{}
				h.Unlock()*/

				h.pendingWrites <- infohash

				break
			}
		}
	}

	h.RLock()
	defer h.RUnlock()
	// In blacklist
	if len(h.unapproved) > 0 {
		if _, found := h.unapproved[infohash]; found {
			chihayaAnnounceBlacklistCount.Add(1)
			return ctx, ErrInfohashUnapproved
		}
	}

	// In whitelist
	if len(h.approved) > 0 {
		if _, found := h.approved[infohash]; found {
			chihayaAnnounceWhitelistCount.Add(1)
			return ctx, nil
		}
	}

	chihayaAnnounceNolistCount.Add(1)
	return ctx, ErrInfohashUnapproved
}

func (h *hook) HandleScrape(ctx context.Context, req *bittorrent.ScrapeRequest, resp *bittorrent.ScrapeResponse) (context.Context, error) {
	// Scrapes don't require any protection.
	chihayaScrapCount.Add(1)
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

// EmptyStruct just an empty struct to work with the database.
type EmptyStruct struct {
}

func (e *EmptyStruct) MarshalBinary() ([]byte, error) {
	return []byte{}, nil
}

func (e *EmptyStruct) UnmarshalBinary(data []byte) error {
	return nil
}

func (e *EmptyStruct) UnmarshalBinaryData(data []byte) ([]byte, error) {
	return data, nil
}
