# Mail-in-a-Box DNS update for ACME lego via Traefik
This is a **very** simple application for automated DNS TXT record updates on a [Mail-in-a-Box](https://mailinabox.email) (MIAB) DNS server via the [lego ACME client](https://github.com/go-acme/lego) for the purpose of TLS certificate generation via DNS-01 challenge. lego already has a [mailinabox provider](https://go-acme.github.io/lego/dns/mailinabox/) built-in, however it is not supported when using lego via [Traefik](https://github.com/traefik/traefik) reverse proxy (see [list of supported providers](https://doc.traefik.io/traefik/https/acme/#providers)). 

To fill this gap, this application uses the [External Program provider](https://go-acme.github.io/lego/dns/exec/) in lego and communicates with the Mail-in-a-Box server via the [MIAB DNS API](https://mailinabox.email/api-docs.html#tag/DNS).

## Build
```
git clone https://github.com/jstnryan/lego-miab-update
cd lego-miab-update
go build
```
The application is name-agnostic, however as written it expects environment variables prefixed as `LEGO_MIAB_`, so it is suggested to keep with that convention.

## Configuration
Three environment variables are required:
* `LEGO_MIAB_HOST` - FQDN of the MIAB server
* `LEGO_MIAB_USER` - email address of an administrative user on the MAIB server
* `LEGO_MIAB_PASS` - the above user's password

### ENV File
Alternately, the environment variables can be loaded from a file. The application will look for the existence of `<program name>.env` within the same directory, and read the values from its contents if it exists. See [lego-miab-update.env](lego-miab-update.env) for an example.

## Usage
Three command line arguments are expected:
* `action` - either `present` to create a TXT record, or `cleanup` to remove the record
* `FQDN` - the domain for which to create a record
* `record` - the record to post

### Standalone
These arguments will be passed automatically by lego, however an example of executing manually (if desired) could look like this:
```shell
LEGO_MIAB_HOST="box.example.com" \
LEGO_MIAB_USER="admin@example.com" \
LEGO_MIAB_PASS="P@55w0rd!" \
./lego-miab-update "present" "_acme-challenge.my.example.com" "MsijOYZxqyjGnFGwhjrhfg-Xgbl5r68WPda0J9EgqqI"
```
### lego
Using the application via lego could look like this:
```shell
LEGO_MIAB_HOST="box.example.com" \
LEGO_MIAB_USER="admin@example.com" \
LEGO_MIAB_PASS="P@55w0rd!" \
EXEC_PATH="./lego-miab-update" \
./lego --email "admin@example.com" --dns exec --domains "my.example.com" run
```
Note that lego's `RAW` mode (`EXEC_MODE=RAW`) is not supported.

### Traefik
There are many ways to configure and launch Traefik (as well as the many ways to configure, for example, Docker or Kubernetes); two things are important when using Traefik with lego, configuring how Traefik interacts with lego, and ensuring the required environment variables are passed.

Using Docker Compose, an example could look like this:
```yaml
services:
  traefik:
    environment:
      - "LEGO_MIAB_HOST=box.example.com"
      - "LEGO_MIAB_USER=admin@example.com"
      - "LEGO_MIAB_PASS=P@55w0rd!"
      - "EXEC_PATH=./lego-miab-update"
    command:
      - --certificatesresolvers.myresolver.acme.email=admin@example.com
      - --certificatesresolvers.myresolver.acme.storage=acme.json
      - --certificatesresolvers.myresolver.acme.dnschallenge=true
      - --certificatesresolvers.myresolver.acme.dnschallenge.provider=exec
```
As with most settings in Traefik, these command line arguments could also be defined via file-based Static Configuration.

**Use of the `*.env` file, [Docker secrets](https://docs.docker.com/engine/swarm/secrets/#simple-example-get-started-with-secrets) (or both), or other creative combination of obfuscation is encouraged to protect sensitive credentials.**

## Logging and Troubleshooting
Unfortunately lego does not capture the External Program output from `stderr` when writing its own errors to the log, so to aid in troubleshooting this app uses a combination of exit codes and logging.

### Log File
A log file will automatically be created (as long as OS permissions allow) in the same directory, depending on the program name (by default this would be `lego-miab-update.log`). This file will contain any errors thrown by the application before it exited, as well as "request" and "success" messages for DNS record updates.

### Exit Codes
Exit codes will appear in lego error logging output (and subsequently in Traefik logs) containing something similar to `Program exit (5)`, where the exit code `5` can be manually correlated to the relevant `lego-miab-update.go` source (in this example, the line containing `os.Exit(5)`).

## Aknowledgements
This code uses the (MIT licensed) [GoDotEnv](https://github.com/joho/godotenv) library by John Barton to load environment variables from disk.
