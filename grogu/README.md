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

1. run bothan follow this instruction [setup_bothan](https://github.com/bandprotocol/bothan/blob/87e12df17d016548b9cf0c928f041615fd432a59/docs/setup_bothan.md)
2. Export bothan url with `export BOTHAN_URL=<Your Bothan URL>`
3. Go to chain directory
4. run `chmod +x ./scripts/start_grogu.sh` to change the access permission of start_grogu.script
5. run `./scripts/start_grogu.sh` to start Grogu

