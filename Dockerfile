FROM golang:1.8.3-alpine


# Get git
RUN apk add --no-cache curl git

# Get glide
RUN go get github.com/Masterminds/glide

# Where chihaya  sources will live
WORKDIR $GOPATH/src/github.com/FactomProject/chihaya

# Populate the source
COPY . .

# Install dependencies
RUN glide install -v

# Set the default
ARG GOOS=linux

# Build and install chihaya 
RUN go install
