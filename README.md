# README

## Generate RPC stubs

`protoc --twirp_out=./ --go_out=./ --go_opt=paths=source_relative --twirp_opt=paths=source_relative ./rpc/chassis/rpc.proto`
