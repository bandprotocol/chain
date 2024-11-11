### Grogu

## Prepare environment

1. Install Golang
2. Install Docker
3. `make install` in chain directory

### How to run BandChain in development mode

1. Go to chain directory
2. run `chmod +x scripts/generate_genesis.sh` to generate genesis.json file
3. run `bandd start` to start BandChain

### How to run Grogu in development mode

1. run `chmod +x ./scripts/bothan/start_bothan.sh` to change the access permission of start_bothan script
2. run bothan with `./scripts/bothan/start_bothan.sh`
3. Export bothan url with `export BOTHAN_URL=<Your Bothan URL>`
4. Go to chain directory
5. run `chmod +x ./scripts/start_grogu.sh` to change the access permission of start_grogu script
6. run `./scripts/start_grogu.sh` to start Grogu

