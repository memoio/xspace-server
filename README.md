# xspace-server
This is xspace server.

## Install from source code

Installation of xspace server includes the following steps:

1. Delete all the local dependency libraries include xspace-server, and download them again.

   ```bash
   cd $HOME
   ### delete all the local dependency libraries include xspace-server
   rm -rf xspace-server nft-solidity memov2-contractsv2 go-mefs-v2 relay memo-go-contracts-v2
   ### download dependency libraries
   git clone https://github.com/memoio/nft-solidity.git
   git clone http://132.232.87.203:8088/508dev/memov2-contractsv2.git
   git clone http://132.232.87.203:8088/508dev/go-mefs-v2.git
   git clone http://132.232.87.203:8088/508dev/relay.git
   git clone http://132.232.87.203:8088/508dev/memo-go-contracts-v2.git
   ### git checkout to the newest branch(go-mefs-v2)
   cd go-mefs-v2
   git checkout 2.7.4
   cd ..
   ### download meeda-node
   git clone https://github.com/memoio/xspace-server.git
   ### 
   cd xspace-server
   ### git checkout to the newest branch
   git checkout dev
   ```

2. Compile xspace-server

   ```bash
   make build
   ```

3. Install xspace-server

   ```bash
   make install
   ```

4. Check if the xspace-server has been installed successfully

   ```bash
   xspace version
   ```

   This command will print the version information.
