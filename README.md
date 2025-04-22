# WireGuard installer

![Lint](https://github.com/angristan/wireguard-install/workflows/Lint/badge.svg)
[![Say Thanks!](https://img.shields.io/badge/Say%20Thanks-!-1EAEDB.svg)](https://saythanks.io/to/angristan)

**This project is a bash script that aims to setup a [WireGuard](https://www.wireguard.com/) VPN on a Linux server, as easily as possible!**

WireGuard is a point-to-point VPN that can be used in different ways. Here, we mean a VPN as in: the client will forward all its traffic through an encrypted tunnel to the server.
The server will apply NAT to the client's traffic so it will appear as if the client is browsing the web with the server's IP.

The script supports both IPv4 and IPv6. Please check the [issues](https://github.com/angristan/wireguard-install/issues) for ongoing development, bugs and planned features! You might also want to check the [discussions](https://github.com/angristan/wireguard-install/discussions) for help.

WireGuard does not fit your environment? Check out [openvpn-install](https://github.com/angristan/openvpn-install).

## Requirements

Supported distributions:

- AlmaLinux >= 8
- Alpine Linux
- Arch Linux
- CentOS Stream >= 8
- Debian >= 10
- Fedora >= 32
- Oracle Linux
- Rocky Linux >= 8
- Ubuntu >= 18.04

## Usage

Download and execute the script. Answer the questions asked by the script and it will take care of the rest.

```bash
curl -O https://raw.githubusercontent.com/angristan/wireguard-install/master/wireguard-install.sh
chmod +x wireguard-install.sh
./wireguard-install.sh
```

It will install WireGuard (kernel module and tools) on the server, configure it, create a systemd service and a client configuration file.

Run the script again to add or remove clients!

## Providers

I recommend these cheap cloud providers for your VPN server:

- [Vultr](https://www.vultr.com/?ref=8948982-8H): Worldwide locations, IPv6 support, starting at \$5/month
- [Hetzner](https://hetzner.cloud/?ref=ywtlvZsjgeDq): Germany, Finland and USA. IPv6, 20 TB of traffic, starting at 4.5€/month
- [Digital Ocean](https://m.do.co/c/ed0ba143fe53): Worldwide locations, IPv6 support, starting at \$4/month

## Contributing

## Discuss changes

Please open an issue before submitting a PR if you want to discuss a change, especially if it's a big one.

### Code formatting

We use [shellcheck](https://github.com/koalaman/shellcheck) and [shfmt](https://github.com/mvdan/sh) to enforce bash styling guidelines and good practices. They are executed for each commit / PR with GitHub Actions, so you can check the configuration [here](https://github.com/angristan/wireguard-install/blob/master/.github/workflows/lint.yml).

## Say thanks

You can [say thanks](https://saythanks.io/to/angristan) if you want!

## Credits & Licence

This project is under the [MIT Licence](https://raw.githubusercontent.com/angristan/wireguard-install/master/LICENSE)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=angristan/wireguard-install&type=Date)](https://star-history.com/#angristan/wireguard-install&Date)

# WireGuard Manager API

A simple REST API for managing WireGuard clients using the wireguard-install.sh script.

## Prerequisites

- Go 1.21 or later
- WireGuard installed and configured using wireguard-install.sh
- Root access (required for WireGuard operations)
- Docker (optional, for containerized deployment)

## Installation

### Local Installation

1. Clone the repository:

```bash
git clone https://github.com/yourusername/wireguard-manager.git
cd wireguard-manager
```

2. Install dependencies:

```bash
go mod download
```

3. Build the application:

```bash
go build
```

4. Run the application:

```bash
sudo WG_AUTH_TOKEN=your-secure-token ./wireguard-manager
```

### Docker Installation

1. Build the Docker image:

```bash
docker build -t wireguard-manager .
```

2. Run the container:

```bash
docker run -d \
  --name wireguard-manager \
  -p 8080:8080 \
  -e WG_AUTH_TOKEN=your-secure-token \
  --cap-add=NET_ADMIN \
  wireguard-manager
```

## Usage

The API will be available at `http://localhost:8080`

### Authentication

All API endpoints require bearer token authentication. Include the token in the Authorization header:

```bash
curl -H "Authorization: Bearer your-secure-token" http://localhost:8080/clients
```

### API Endpoints

#### Create a new client

```bash
curl -X POST http://localhost:8080/clients \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secure-token" \
  -d '{"name": "client1"}'
```

#### List all clients

```bash
curl -H "Authorization: Bearer your-secure-token" http://localhost:8080/clients
```

#### Delete a client

```bash
curl -X DELETE \
  -H "Authorization: Bearer your-secure-token" \
  http://localhost:8080/clients/client1
```

## Security Note

This application must be run as root since it needs to execute WireGuard commands. Make sure to:

1. Only expose the API to trusted networks
2. Use HTTPS in production
3. Use a strong bearer token
4. Regularly update the application and its dependencies

## License

MIT License
