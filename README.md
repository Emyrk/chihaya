# Chihaya Tracker
This is the torrent tracker going to be used for factomd releated torrents. The main.go almost directly comes from [chihaya's repo](https://github.com/chihaya/chihaya/tree/master/cmd/chihaya). The difference is this uses a [custom middleware](https://github.com/chihaya/chihaya#architecture) that has a whitelist for certain infohashes.

To add a torrent to the whitelist, a signed infohash by a signer must be announced to the tracker. The tracker will add it to it's active list, and save the infohash to a database it can read from on launch.

Signers are read from the chihaya.yaml file, in `/etc/chihaya.yaml`. To add a signer, edit the config and send a SIGUSR1 signal to the chihaya process, E.G: `kill -10 PID`. That will tell chihaya to read from the config file. It will grab the signer list, and also reload it's whitelist from the database (asumming Map was not chosen)

You can manually add infohashes to the whitelist, but be advised these will NOT be saved in the database. Only infohashes that come through the announce url and are signed can be added to the database. If an infohash is in the config, and comes in signed, it will still not be saved. So if you wish for an infohash to be saved, it must not be in the config file.

A blacklist also exists, but is currently not used for anything. There is no codepath for an infohash to be saved to the database for blacklists, but the config's blacklist will be enforced.

factomd-torrent library has a CreateAndSignTorrent() function that this tracker will recognize.