# chihaya Docker Helper

The chihaya Docker Helper is a simple tool to help build chihaya in a container

## Prerequisites

You must have at least Docker v17 installed on your system.

Having this repo cloned helps too 😇

## Build
From wherever you have cloned this repo, run

`docker build -t chihaya_container .`

(yes, you can replace **chihaya_container** with whatever you want to call the container.  e.g. **chihaya**, **foo**, etc.)

#### Cross-Compile
To cross-compile for a different target, you can pass in a `build-arg` as so

`docker build -t chihaya_container --build-arg GOOS=darwin .`

## Copy
So yeah, you want to get your binary _out_ of the container. To do so, you basically mount your target into the container, and copy the binary over, like so


`docker run --rm -v <FULLY_QUALIFIED_PATH_TO_TARGET_DIRECTORY>:/destination chihaya_container /bin/cp /go/bin/chihaya /destination`

e.g.

`docker run --rm -v /tmp:/destination chihaya_container /bin/cp /go/bin/chihaya /destination`

which will copy the binary to `/tmp/chihaya`

**Note** : You should replace **chihaya_container** with whatever you called it in the **build** section above  e.g. **chihaya**, **foo**, etc.

#### Cross-Compile
If you cross-compiled to a different target, your binary will be in `/go/bin/<target>/chihaya`.  e.g. If you built with `--build-arg GOOS=darwin`, then you can copy out the binary with

`docker run --rm -v <FULLY_QUALIFIED_PATH_TO_TARGET_DIRECTORY>:/destination chihaya_container /bin/cp /go/bin/darwin_amd64/chihaya /destination`

e.g.

`docker run --rm -v /tmp:/destination chihaya_container /bin/cp /go/bin/darwin_amd64/chihaya /destination` 

which will copy the darwin_amd64 version of the binary to `/tmp/chihaya`

**Note** : You should replace **chihaya_container** with whatever you called it in the **build** section above  e.g. **chihaya**, **foo**, etc.