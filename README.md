# smbdriver
This driver mounts SMB shares. For more information about using this driver in your Cloud Foundry, visit the [smb-volume-release git repository](https://github.com/AbelHu/smb-volume-release).

# Install direnv
Please install [direnv](https://github.com/direnv/direnv).

# Run Tests

  ```
    git clone https://github.com/AbelHu/smbdriver.git
    cd smbdriver
    direnv allow .
    go get github.com/tools/godep
    godep get .
    go get -t .
    go test
    cd cmd/smbdriver
    go get -t .
    go test
  ```