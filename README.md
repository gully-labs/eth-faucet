<!-- gully-header:start -->
<p align="center">
  <a href="https://faucet.gullychain.xyz">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset=".github/assets/gullychain-wordmark-for-dark-bg.png">
      <img alt="Gully Chain" src=".github/assets/gullychain-wordmark-for-light-bg.png" width="360">
    </picture>
  </a>
</p>

<p align="center"><b>The social chain where your profile is your wallet. Arbitrum Orbit L3, testnet chain 485590.</b></p>

<p align="center">Gully Faucet: testnet ETH for Gully Chain Testnet (chain 485590).</p>

<p align="center">
  <a href="https://gullychain.xyz"><img src="https://img.shields.io/badge/gullychain.xyz-01D9FA?style=for-the-badge&labelColor=000000&logo=googlechrome&logoColor=000000" alt="Website"></a>
  <a href="https://faucet.gullychain.xyz"><img src="https://img.shields.io/badge/testnet--faucet.gullychain.xyz-0A0F1A?style=for-the-badge&labelColor=000000&logo=googlechrome&logoColor=01D9FA" alt="App"></a>
  <a href="https://docs.gullychain.xyz"><img src="https://img.shields.io/badge/Docs-0A0F1A?style=for-the-badge&labelColor=000000&logo=readthedocs&logoColor=01D9FA" alt="Docs"></a>
  <a href="https://t.me/GullyChain"><img src="https://img.shields.io/badge/Telegram-0A0F1A?style=for-the-badge&labelColor=000000&logo=telegram&logoColor=26A5E4" alt="Telegram"></a>
  <a href="https://discord.com/invite/tY2ErsNNBp"><img src="https://img.shields.io/badge/Discord-0A0F1A?style=for-the-badge&labelColor=000000&logo=discord&logoColor=5865F2" alt="Discord"></a>
  <a href="https://x.com/Gully_Labs"><img src="https://img.shields.io/badge/@Gully__Labs-0A0F1A?style=for-the-badge&labelColor=000000&logo=x&logoColor=white" alt="X"></a>
</p>

<p align="center"><sub>A Gully Chain fork of <a href="https://github.com/chainflag/eth-faucet">chainflag/eth-faucet</a>, maintained by <a href="https://gullylabs.xyz">Gully Labs</a>. It is used under its MIT licence; the upstream LICENSE, copyright and attribution notices are kept unchanged. The upstream README follows below as published. Not affiliated with or endorsed by chainflag.</sub></p>

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

`WEB3_PROVIDER` (and `-wallet.provider`) accept an ordered, comma-separated list of http(s) RPC URLs. Each call goes to
the first healthy URL and moves on to the next after a network error, an 8s timeout, HTTP 5xx/429 or a JSON-RPC
rate-limit error; a failed URL is skipped for 30s, then tried again. Attempts are sequential, so a signed transaction
is never broadcast to two URLs at once. Empty = the Gully Chain testnet public list:
`https://testnet-rpc.gullychain.xyz,https://rpc.gullyrpc.xyz/gullychain/evm/485590,https://rpc2.gullychain.xyz` (never
`rpc.gullychain.xyz`: it becomes the mainnet RPC on cutover day).

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

**Token drip (test stablecoin, Gully fork)**

With `-token.address` set, every successful ETH claim also sends a small ERC-20 amount (on Gully Chain: 1 glUSD,
`0xDBE84bE6496Fb3d07F601548555ec7Acc4CeBa8F`) from the same hot wallet. It runs behind the same gate as the ETH payout
(Gully ID sign-in, eligibility, denylist, real client IP, one claim per account/address/IP per window, ETH budget) and
has its own budget on top. It fails closed: when the token budget is used up, the wallet holds less than one payout,
the balance check fails or the send fails, the claim still pays ETH and the response says why no token was sent. Only
sent drips count against the budget; the lifetime total is kept in the state file.

| Flag / env | Description | Default |
|---|---|---|
| `-token.address` / `FAUCET_TOKEN_ADDRESS` | ERC-20 to drip (decimals are read from the token) | empty = ETH only |
| `-token.symbol` / `FAUCET_TOKEN_SYMBOL` | symbol shown on the page | glUSD |
| `-token.amount` / `FAUCET_TOKEN_AMOUNT` | whole tokens per claim | 1 |
| `-token.daycap` / `FAUCET_TOKEN_BUDGET_DAY` | max tokens per rolling 24 h (required, >= one payout) | 0 = refuse to start |
| `-token.totalcap` / `FAUCET_TOKEN_BUDGET_TOTAL` | max tokens over the faucet's lifetime | 0 = no lifetime cap |

`/api/claim` adds `"token": {"symbol","amount","sent","tx","msg"}` to a successful response; `/api/info` adds
`"token": {"address","symbol","decimals","amount"}`.

**Developer tier (Gully fork)**

With `-devtier.amount` above 0, signed-in Gully ID users who are verified developers get a larger drip on a shorter
cooldown, so teams building on Gully Chain can deploy and test without asking for ETH by hand. hCaptcha and the Gully ID
sign-in stay mandatory for the tier (when `-gullyid.required` is off, a valid session only upgrades a claim; without one
the claim is a normal one). A user qualifies when **any** of these holds:

- their Gully ID subject (`sub`) is on the allowlist (`-devtier.subs` or the allowlist file),
- one of their Gully ID wallets, or the recipient address, is on the wallet allowlist (`-devtier.wallets` or the file),
- `-devtier.profile` is on and their Gully ID profile says `"developer": true` or lists `"developer"` in `"roles"`.

A developer claim: pays `-devtier.amount`; is limited per Gully ID account and per recipient address by
`-devtier.minutes` (its own keys, no per-IP window, since offices and CI share an IP); skips the fresh-address check
(developers deploy from new keys); counts toward the global hourly/daily ETH budget and the tier's own daily budget. The
denylist still applies. The startup log line `developer tier on` shows the counts, never the lists.

| Flag / env | Description | Default |
|---|---|---|
| `-devtier.amount` / `FAUCET_DEV_AMOUNT` | ETH per developer claim | 0 = tier off |
| `-devtier.minutes` / `FAUCET_DEV_MINUTES` | minutes between developer claims per account and per address (must not exceed `-faucet.minutes`) | 360 |
| `-devtier.daycap` / `FAUCET_DEV_BUDGET_DAY` | max ETH per rolling 24 h across all developer claims | 0 = only the global caps |
| `-devtier.subs` / `FAUCET_DEV_SUBS` | comma-separated Gully ID subjects | empty |
| `-devtier.wallets` / `FAUCET_DEV_WALLETS` | comma-separated 0x addresses | empty |
| `-devtier.file` / `FAUCET_DEV_ALLOWLIST_FILE` | one sub or 0x address per line, `#` comments (read at start) | empty |
| `-devtier.profile` / `FAUCET_DEV_TRUST_PROFILE` | trust the Gully ID profile's developer flag | false |

The tier needs `-gullyid.client` and at least one way to qualify, or the faucet refuses to start. A successful
`/api/claim` response adds `"tier": "standard"|"developer"` and `"amount"` (ETH sent); `/api/info` adds
`"dev_tier": {"payout","window_hours"}` when the tier is on, and the page then tells developers how to get it.

### Docker deployment

```bash
docker run -d -p 8080:8080 -e WEB3_PROVIDER=rpc_endpoint -e PRIVATE_KEY=hex_private_key chainflag/eth-faucet:1.2.0
```

or

```bash
docker run -d -p 8080:8080 -e WEB3_PROVIDER=rpc_endpoint -e KEYSTORE=keystore_path -v `pwd`/keystore:/app/keystore -v `pwd`/password.txt:/app/password.txt chainflag/eth-faucet:1.2.0
```

### Gully Chain hosts

The Gully Chain faucet is served at https://faucet.gullychain.xyz (canonical) and https://faucet.gullychain.xyz
(alias). The server has no Host check, so one container answers both names behind Caddy. The hCaptcha site must list
both hostnames (hCaptcha dashboard > Sites > the faucet site key > Domains), or the widget fails on the missing one.
A Gully ID client used with `-gullyid.client` needs both origins in its allowed origins / redirect URIs.

The Gully Chain app (https://kre8.fun, and https://app.gullychain.xyz during the move) calls this server same-origin
through `/faucet-api/*` on its own host. The server has no Origin/Referer check, so there is nothing to allow-list here;
if the app renders the hCaptcha widget with this site key, the hCaptcha site must also list `kre8.fun`
(keep `app.gullychain.xyz`).

## License

Distributed under the MIT License. See LICENSE for more information.


<!-- gully-footer:start -->
---

<p align="center">
  <a href="https://gullylabs.xyz">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset=".github/assets/powered-by-gully-labs-dark-2x.png">
      <img alt="Powered by Gully Labs" src=".github/assets/powered-by-gully-labs-light-2x.png" width="254" height="64">
    </picture>
  </a>
</p>
<!-- gully-footer:end -->
