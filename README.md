<!-- gully-header:start -->
<p align="center">
  <a href="https://gullylabs.xyz"><img src="https://img.shields.io/badge/gullylabs.xyz-E8C060?style=for-the-badge&labelColor=0B0E13&logo=googlechrome&logoColor=E8C060" alt="Website"></a>
  <a href="https://x.com/Gully_Labs"><img src="https://img.shields.io/badge/@Gully__Labs-0B0E13?style=for-the-badge&labelColor=0B0E13&logo=x&logoColor=white" alt="X"></a>
  <a href="https://t.me/GullyLabsNews"><img src="https://img.shields.io/badge/Telegram-0B0E13?style=for-the-badge&labelColor=0B0E13&logo=telegram&logoColor=26A5E4" alt="Telegram"></a>
  <a href="https://docs.gullylabs.xyz"><img src="https://img.shields.io/badge/Docs-0B0E13?style=for-the-badge&labelColor=0B0E13&logo=readthedocs&logoColor=E8C060" alt="Docs"></a>
  <a href="https://gullylabs.xyz/chain-kit"><img src="https://img.shields.io/badge/Gully_Chains-0B0E13?style=for-the-badge&labelColor=0B0E13&logo=ethereum&logoColor=E8C060" alt="Gully Chains"></a>
  <a href="https://github.com/gully-labs"><img src="https://img.shields.io/badge/GitHub-0B0E13?style=for-the-badge&labelColor=0B0E13&logo=github&logoColor=white" alt="GitHub"></a>
</p>

<p align="center"><b>Gully Faucet: testnet funds, for the Gully L3 testnet (planned).</b></p>

<p align="center"><sub>Maintained by <a href="https://gullylabs.xyz">Gully Labs</a> as part of Gully Chains. Fork of <a href="https://github.com/chainflag/eth-faucet">chainflag/eth-faucet</a>, used under its MIT licence. Not affiliated with or endorsed by the upstream authors. For the Gully L3 testnet (planned); nothing is running yet.</sub></p>

---
<!-- gully-header:end -->

# eth-faucet

[![Build](https://img.shields.io/github/actions/workflow/status/chainflag/eth-faucet/build.yml?branch=main)](https://github.com/chainflag/eth-faucet/actions/workflows/build.yml)
[![Release](https://img.shields.io/github/v/release/chainflag/eth-faucet)](https://github.com/chainflag/eth-faucet/releases)
[![Report](https://goreportcard.com/badge/github.com/chainflag/eth-faucet)](https://goreportcard.com/report/github.com/chainflag/eth-faucet)
[![Go](https://img.shields.io/github/go-mod/go-version/chainflag/eth-faucet)](https://go.dev/)
[![License](https://img.shields.io/github/license/chainflag/eth-faucet)](https://github.com/chainflag/eth-faucet/blob/main/LICENSE)

The faucet is a web application with the goal of distributing small amounts of Ether in private and test networks.

## Features

* Configure the funding account using a private key or keystore
* Implement CAPTCHA verification to prevent abuse
* Rate-limit requests by ETH address and IP address to prevent spam
* Prevent X-Forwarded-For spoofing by specifying the number of reverse proxies

## Get started

### Prerequisites

* Go (version 1.17 or later)
* Node.js

### Installation

1. Clone the repository and navigate to the app’s directory
```bash
git clone https://github.com/chainflag/eth-faucet.git
cd eth-faucet
```

2. Bundle frontend with Vite
```bash
go generate
```

3. Build Go project 
```bash
go build -o eth-faucet
```

## Usage

**Use a private key**

```bash
./eth-faucet -httpport 8080 -wallet.provider http://localhost:8545 -wallet.privkey privkey
```

**Use a keystore**

```bash
./eth-faucet -httpport 8080 -wallet.provider http://localhost:8545 -wallet.keyjson keystore -wallet.keypass password.txt
```

### Configuration

You can configure the funding account by using environment variables instead of command-line flags:
```bash
export WEB3_PROVIDER=rpc_endpoint
export PRIVATE_KEY=hex_private_key
```

or

```bash
export WEB3_PROVIDER=rpc_endpoint
export KEYSTORE=keystore_path
echo "your_keystore_password" > `pwd`/password.txt
```

Then run the faucet application without the wallet command-line flags:
```bash
./eth-faucet -httpport 8080
```

**Optional Flags**

The following are the available command-line flags(excluding above wallet flags):

| Flag              | Description                                      | Default Value |
|-------------------|--------------------------------------------------|---------------|
| -httpport         | Listener port to serve HTTP connection           | 8080          |
| -proxycount       | Count of reverse proxies in front of the server  | 0             |
| -faucet.amount    | Number of Ethers to transfer per user request    | 1.0           |
| -faucet.minutes   | Number of minutes to wait between funding rounds | 1440          |
| -faucet.name      | Network name to display on the frontend          | testnet       |
| -faucet.symbol    | Token symbol to display on the frontend          | ETH           |
| -hcaptcha.sitekey | hCaptcha sitekey                                 |               |
| -hcaptcha.secret  | hCaptcha secret                                  |               |

### Docker deployment

```bash
docker run -d -p 8080:8080 -e WEB3_PROVIDER=rpc_endpoint -e PRIVATE_KEY=hex_private_key chainflag/eth-faucet:1.2.0
```

or

```bash
docker run -d -p 8080:8080 -e WEB3_PROVIDER=rpc_endpoint -e KEYSTORE=keystore_path -v `pwd`/keystore:/app/keystore -v `pwd`/password.txt:/app/password.txt chainflag/eth-faucet:1.2.0
```

## License

Distributed under the MIT License. See LICENSE for more information.
