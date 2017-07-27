FROM golang:1.8.3

# Get git
RUN apt-get update \
    && apt-get -y install curl git \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Get glide
RUN go get github.com/Masterminds/glide

# Where chihaya  sources will live
WORKDIR $GOPATH/src/github.com/FactomProject/chihaya

# Get the dependencies
COPY glide.yaml glide.lock ./

# Install dependencies
RUN glide install -v

# Populate the rest of the source
COPY . .

# Build and install chihaya 
RUN go install
